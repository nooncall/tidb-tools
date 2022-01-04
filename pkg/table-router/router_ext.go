package router

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/pingcap/errors"
	"github.com/pingcap/tidb-tools/pkg/wasm/common"
	v1 "github.com/pingcap/tidb-tools/pkg/wasm/v1"
	"github.com/pingcap/tidb-tools/pkg/wasm/wasmer"
)

const (
	WasmRootContextID = 1
)

type WasmExtractor struct {
	TargetColumn string `json:"target-column" toml:"target-column" yaml:"target-column"`
	Table        string `json:"table" toml:"table" yaml:"table"`
	WasmModule   string `json:"wasm-module" toml:"wasm-module" yaml:"wasm-module"`

	wasmPath string

	wasmInstance common.WasmInstance
	wasmCtx      *v1.ABIContext
	wasmHandler  *TableRouterImportsHandler
	wasmCtxID    int32

	once sync.Once

	mu        sync.Mutex
	fsWatcher *fsnotify.Watcher
}

type TableRouterImportsHandler struct {
	v1.DefaultImportsHandler
	ValMap common.CommonHeader
	//ColIndexes common.CommonHeader

	currentVals []interface{}
	//currentColIndexes []int
}

func NewTableRouterImportsHandler() *TableRouterImportsHandler {
	return &TableRouterImportsHandler{
		ValMap: make(common.CommonHeader),
	}
}

func (r *WasmExtractor) initWasm() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.WithMessage(err, "fsnotify NewWatcher error")
	}
	r.fsWatcher = watcher

	dir := os.Getenv("WASM_MODULE_DIR")
	if dir == "" {
		dir, _ = os.Getwd()
	}
	p := path.Join(dir, r.WasmModule)
	r.wasmPath = p

	if err := r.fsWatcher.Add(dir); err != nil {
		return errors.WithMessage(err, "fsWatcher.Add error")
	}
	r.startWatchFileChange()
	return r.InitWasmInstance(p)
}

func (r *WasmExtractor) startWatchFileChange() {
	watcher := r.fsWatcher
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				/*
					fswatcher event:  "/Users/eastfisher/workspace/pingcap/tidb-tools/pkg/column-mapping/add_prefix.wasm": REMOVE
					fswatcher event:  "/Users/eastfisher/workspace/pingcap/tidb-tools/pkg/column-mapping/add_prefix.wasm": CREATE
					fswatcher modified file:  /Users/eastfisher/workspace/pingcap/tidb-tools/pkg/column-mapping/add_prefix.wasm
					fswatcher event:  "/Users/eastfisher/workspace/pingcap/tidb-tools/pkg/column-mapping/add_prefix.wasm": WRITE
					fswatcher event:  "/Users/eastfisher/workspace/pingcap/tidb-tools/pkg/column-mapping/add_prefix.wasm": WRITE|CHMOD
				*/
				fmt.Println("fswatcher event: ", event)

				if event.Name == r.wasmPath && (event.Op&fsnotify.Chmod) == fsnotify.Chmod {
					fmt.Println("fswatcher modified file: ", event.Name)
					if err := r.InitWasmInstance(r.wasmPath); err != nil {
						fmt.Printf("fswatcher re init wasm error: %v", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("fswatcher error: ", err)
			}
		}
	}()
}

func (r *WasmExtractor) InitWasmInstance(p string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.closeCurrentWasm()

	instance := wasmer.NewWasmerInstanceFromFile(p)
	v1.RegisterImports(instance)
	if err := instance.Start(); err != nil {
		return err
	}

	r.wasmInstance = instance
	handler := NewTableRouterImportsHandler()
	ctx := &v1.ABIContext{
		Imports:  handler,
		Instance: r.wasmInstance,
	}
	r.wasmCtx = ctx
	r.wasmHandler = handler

	var once sync.Once

	var err error
	once.Do(func() {
		err = ctx.GetExports().ProxyOnContextCreate(WasmRootContextID, 0)
	})
	if err != nil {
		return errors.Annotate(err, "init rootContextID error")
	}
	r.wasmCtxID = WasmRootContextID

	fmt.Println("init wasm success!!!")
	return nil
}

func (r *WasmExtractor) closeCurrentWasm() {
	fmt.Println("close current wasm!!!")
	if r.wasmInstance != nil {
		r.wasmInstance.Stop()
		r.wasmInstance = nil
	}
	r.wasmCtx = nil
	r.wasmHandler = nil
	r.wasmCtxID = 0
}

func (r *WasmExtractor) WasmHandle(vals []interface{}) (string, error) {
	r.once.Do(func() {
		if err := r.initWasm(); err != nil {
			panic(err)
		}
	})

	r.mu.Lock()
	defer r.mu.Unlock()

	ctx := r.wasmCtx
	r.wasmInstance.Lock(ctx)
	defer r.wasmInstance.Unlock()

	ctxID := r.nextWasmCtxID()
	// create wasm-side context id for current http req
	err := ctx.GetExports().ProxyOnContextCreate(ctxID, WasmRootContextID)
	fmt.Printf("[rule] ProxyOnContextCreate, id: %d, err: %v\n", ctxID, err)
	if err != nil {
		return "", errors.Annotatef(err, "ProxyOnContextCreate error, vals: %+v", vals)
	}

	// 先把map中的上一次的值清掉, 并写入当前值, 并传入position
	r.wasmHandler.ClearAndSet(vals)

	// 通知wasm module处理数据
	action, err := ctx.GetExports().ProxyOnRequestHeaders(ctxID, int32(len(vals)), 1)
	fmt.Printf("[rule] ProxyOnContextCreate, id: %d, action: %v, err: %v\n", ctxID, action, err)
	if err != nil {
		return "", errors.Annotatef(err, "ProxyOnRequestHeaders error, vals: %+v", vals)
	}

	// 从wasm中获取处理后的值
	return r.wasmHandler.GetValString(), nil
}

func (r *WasmExtractor) nextWasmCtxID() int32 {
	return atomic.AddInt32(&r.wasmCtxID, 1)
}

func (h *TableRouterImportsHandler) Log(level v1.LogLevel, msg string) v1.WasmResult {
	fmt.Printf("[wasm cm] on Log: %s\n", msg)
	return v1.WasmResultOk
}

// 获取row change的所有列值
func (h *TableRouterImportsHandler) GetHttpRequestHeader() common.HeaderMap {
	return h.ValMap
}

func (h *TableRouterImportsHandler) ClearAndSet(vals []interface{}) {
	// 设置值
	valMap := h.ValMap
	for idx, val := range vals {
		idxStr := strconv.Itoa(idx)
		valStr := fmt.Sprintf("%v", val)
		valMap[idxStr] = valStr
	}
	h.currentVals = vals
}

func (h *TableRouterImportsHandler) GetValString() string {
	val, _ := h.ValMap.Get("#@result")
	return val
}
