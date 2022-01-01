package handler_study

import (
	"fmt"

	"github.com/pingcap/tidb-tools/pkg/wasm/common"
	proxywasm "github.com/pingcap/tidb-tools/pkg/wasm/v1"
)

// implement proxywasm.ImportsHandler.
type importHandler struct {
	reqHeader common.HeaderMap
	proxywasm.DefaultImportsHandler
}

// override.
func (im *importHandler) GetHttpRequestHeader() common.HeaderMap {
	return im.reqHeader
}

// override.
func (im *importHandler) Log(level proxywasm.LogLevel, msg string) proxywasm.WasmResult {
	fmt.Println(msg)
	return proxywasm.WasmResultOk
}
