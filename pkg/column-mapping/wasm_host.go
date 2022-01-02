package column

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/pingcap/tidb-tools/pkg/wasm/common"
	v1 "github.com/pingcap/tidb-tools/pkg/wasm/v1"
)

type ColumnMappingImportsHandler struct {
	v1.DefaultImportsHandler
	ValMap common.CommonHeader

	currentVals []interface{}
}

// 用于将map[idx]val转换回[]string, 并保留顺序
type keyValues [][2]string

func NewColumnMappingImportsHandler() *ColumnMappingImportsHandler {
	valMap := make(common.CommonHeader)
	return &ColumnMappingImportsHandler{ValMap: valMap}
}

func (h *ColumnMappingImportsHandler) Log(level v1.LogLevel, msg string) v1.WasmResult {
	fmt.Printf("[wasm cm] on Log: %s\n", msg)
	return v1.WasmResultOk
}

func (h *ColumnMappingImportsHandler) GetHttpRequestHeader() common.HeaderMap {
	return h.ValMap
}

func (h *ColumnMappingImportsHandler) ClearAndSet(vals []interface{}) {
	//valMap := make(common.CommonHeader)
	valMap := h.ValMap
	for idx, val := range vals {
		idxStr := strconv.Itoa(idx)
		valStr := fmt.Sprintf("%v", val)
		valMap[idxStr] = valStr
	}
	h.currentVals = vals
}

// TODO: 需要做类型转换, 使用h.currentVals
func (h *ColumnMappingImportsHandler) GetVals() []interface{} {
	kv := buildKeyValues(h.ValMap)
	var rets []interface{}
	for _, v := range kv.ToValues() {
		rets = append(rets, v)
	}
	return rets
}

func buildKeyValues(valMap map[string]string) keyValues {
	var rets [][2]string
	for k, v := range valMap {
		rets = append(rets, [2]string{k, v})
	}
	return rets
}

func (kv keyValues) ToValues() []string {
	sort.Slice(kv, func(i int, j int) bool {
		idxI, _ := strconv.Atoi(kv[i][0])
		idxJ, _ := strconv.Atoi(kv[j][0])
		return idxI < idxJ
	})
	var rets []string
	for _, v := range kv {
		rets = append(rets, v[1])
	}
	return rets
}
