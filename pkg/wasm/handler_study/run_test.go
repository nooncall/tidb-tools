package handler_study

import (
	"fmt"
	"os"
	path2 "path"
	"sync"
	"testing"

	"github.com/pingcap/tidb-tools/pkg/wasm/common"
	v1 "github.com/pingcap/tidb-tools/pkg/wasm/v1"
	"github.com/pingcap/tidb-tools/pkg/wasm/wasmer"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	dir, _ := os.Getwd()
	path := path2.Join(dir, "module/column_mapping.wasm")
	instance := wasmer.NewWasmerInstanceFromFile(path)
	v1.RegisterImports(instance)
	err := instance.Start()
	require.NoError(t, err)

	ctx := &v1.ABIContext{
		Imports:  NewImportsHandlerImpl(),
		Instance: instance,
	}

	var rootCtxID int32 = 1

	once := sync.Once{}
	once.Do(func() {
		err = ctx.GetExports().ProxyOnContextCreate(rootCtxID, 0)
		require.NoError(t, err)
	})

	instance.Lock(ctx)
	defer instance.Unlock()

	err = ctx.GetExports().ProxyOnContextCreate(2, rootCtxID)
	require.NoError(t, err)
	//err = ctx.GetExports().ProxyOnContextCreate(3, 2)
	//require.NoError(t, err)

	//var ctxID int32 = 2
	//request start
	//err = ctx.GetExports().ProxyOnContextCreate(ctxID, rootCtxID)
	require.NoError(t, err)

	// handle logic

}

func TestHeaderMap(t *testing.T) {
	commonHeader := common.CommonHeader{
		"hello": "world",
	}
	importHandler := &importHandler{
		reqHeader: commonHeader,
	}

	dir, _ := os.Getwd()
	path := path2.Join(dir, "module/headers_set.wasm")
	instance := wasmer.NewWasmerInstanceFromFile(path)
	v1.RegisterImports(instance)
	err := instance.Start()
	require.NoError(t, err)

	ctx := &v1.ABIContext{
		Imports:  importHandler,
		Instance: instance,
	}

	var rootContextID int32 = 1
	once := sync.Once{}
	once.Do(func() {
		err = ctx.GetExports().ProxyOnContextCreate(rootContextID, 0)
		require.NoError(t, err)
	})

	instance.Lock(ctx)
	defer instance.Unlock()

	var contextID int32 = 2
	err = ctx.GetExports().ProxyOnContextCreate(contextID, rootContextID)
	require.NoError(t, err)

	// call wasm-side on_request_header
	action, err := ctx.GetExports().ProxyOnRequestHeaders(contextID, int32(len(commonHeader)), 1)
	fmt.Printf("[server] ProxyOnRequestHeaders contextID: %d, action: %v, err: %v\n", contextID, action, err)

	// 如果key也变了, 原来的key不会被删除
	fmt.Printf("[server]new header: %+v", commonHeader)
}

func TestColumnMapping(t *testing.T) {
	commonHeader := common.CommonHeader{
		"test_schema": "test_table",
	}
	importHandler := &importHandler{
		reqHeader: commonHeader,
	}

	dir, _ := os.Getwd()
	path := path2.Join(dir, "module/headers_set.wasm")
	instance := wasmer.NewWasmerInstanceFromFile(path)
	v1.RegisterImports(instance)
	err := instance.Start()
	require.NoError(t, err)

	ctx := &v1.ABIContext{
		Imports:  importHandler,
		Instance: instance,
	}

	var rootContextID int32 = 1
	once := sync.Once{}
	once.Do(func() {
		err = ctx.GetExports().ProxyOnContextCreate(rootContextID, 0)
		require.NoError(t, err)
	})

	instance.Lock(ctx)
	defer instance.Unlock()

	var contextID int32 = 2
	err = ctx.GetExports().ProxyOnContextCreate(contextID, rootContextID)
	require.NoError(t, err)

	// call wasm-side on_request_header
	action, err := ctx.GetExports().ProxyOnRequestHeaders(contextID, int32(len(commonHeader)), 1)
	fmt.Printf("[server] ProxyOnRequestHeaders contextID: %d, action: %v, err: %v\n", contextID, action, err)

	// 如果key也变了, 原来的key不会被删除
	fmt.Printf("[server]new header: %+v", commonHeader)

	importHandler.reqHeader = make(common.CommonHeader)
	importHandler.reqHeader.Set("next_key", "next_value")

	contextID = 3
	err = ctx.GetExports().ProxyOnContextCreate(contextID, rootContextID)
	require.NoError(t, err)

	// call wasm-side on_request_header
	action, err = ctx.GetExports().ProxyOnRequestHeaders(contextID, int32(len(commonHeader)), 1)
	fmt.Printf("[server] ProxyOnRequestHeaders contextID: %d, action: %v, err: %v\n", contextID, action, err)

	// 如果key也变了, 原来的key不会被删除
	fmt.Printf("[server]new header: %+v", commonHeader)

}
