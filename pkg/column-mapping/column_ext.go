package column

import (
	"fmt"
	"os"
	"path"
	"sync"
	"sync/atomic"

	"github.com/pingcap/errors"
	"github.com/pingcap/tidb-tools/pkg/wasm/v1"
	"github.com/pingcap/tidb-tools/pkg/wasm/wasmer"
)

func (r *Rule) initWasm() error {
	dir := os.Getenv("WASM_MODULE_DIR")
	if dir == "" {
		dir, _ = os.Getwd()
	}
	p := path.Join(dir, r.WasmModule)
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

	return nil
}

func (r *Rule) wasmHandle(info *mappingInfo, vals []interface{}) ([]interface{}, error) {
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

	// 先把map中的上一次的值清掉, 并写入当前值
	r.wasmHandler.ClearAndSet(vals)

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
