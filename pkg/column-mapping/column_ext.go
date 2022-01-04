package column

import (
	"fmt"
	"os"
	"path"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/pingcap/errors"
	v1 "github.com/pingcap/tidb-tools/pkg/wasm/v1"
	"github.com/pingcap/tidb-tools/pkg/wasm/wasmer"
)

func (r *Rule) initWasm() error {
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

func (r *Rule) startWatchFileChange() {
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

func (r *Rule) InitWasmInstance(p string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.closeCurrentWasm()

	instance := wasmer.NewWasmerInstanceFromFile(p)
	v1.RegisterImports(instance)
	if err := instance.Start(); err != nil {
		return err
	}

	r.wasmInstance = instance
	handler := NewColumnMappingImportsHandler()
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

func (r *Rule) closeCurrentWasm() {
	fmt.Println("close current wasm!!!")
	if r.wasmInstance != nil {
		r.wasmInstance.Stop()
		r.wasmInstance = nil
	}
	r.wasmCtx = nil
	r.wasmHandler = nil
	r.wasmCtxID = 0
}

func (r *Rule) WasmHandle(info *mappingInfo, vals []interface{}) ([]interface{}, error) {
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
		return nil, errors.Annotatef(err, "ProxyOnContextCreate error, vals: %+v", vals)
	}

	// 先把map中的上一次的值清掉, 并写入当前值, 并传入position
	r.wasmHandler.ClearAndSet(vals, []int{info.targetPosition})

	// 通知wasm module处理数据
	action, err := ctx.GetExports().ProxyOnRequestHeaders(ctxID, int32(len(vals)), 1)
	fmt.Printf("[rule] ProxyOnContextCreate, id: %d, action: %v, err: %v\n", ctxID, action, err)
	if err != nil {
		return nil, errors.Annotatef(err, "ProxyOnRequestHeaders error, vals: %+v", vals)
	}

	// 从wasm中获取处理后的值
	return r.wasmHandler.GetVals()
}

func (r *Rule) nextWasmCtxID() int32 {
	return atomic.AddInt32(&r.wasmCtxID, 1)
}
