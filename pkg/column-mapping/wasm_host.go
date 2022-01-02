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

func (h *ColumnMappingImportsHandler) GetVals() ([]interface{}, error) {
	kv := buildKeyValues(h.ValMap)
	var rets []interface{}
	for idx, v := range kv.ToValues() {
		vi, err := buildNewInterfaceValue(h.currentVals[idx], v)
		if err != nil {
			// 这里不返回错误了, 直接返回原值, 先保证能兼容用
			rets = append(rets, h.currentVals[idx])
			continue
		}
		rets = append(rets, vi)
	}
	return rets, nil
}

func buildNewInterfaceValue(originValue interface{}, newValueStr string) (interface{}, error) {
	switch originValue.(type) {
	case int:
		ret, err := strconv.ParseInt(newValueStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return int(ret), nil
	case uint:
		ret, err := strconv.ParseInt(newValueStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return uint(ret), nil
	case int32:
		ret, err := strconv.ParseInt(newValueStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return int32(ret), nil
	case uint32:
		ret, err := strconv.ParseInt(newValueStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return uint32(ret), nil
	case int64:
		ret, err := strconv.ParseInt(newValueStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return int64(ret), nil
	case uint64:
		ret, err := strconv.ParseInt(newValueStr, 10, 64)
		if err != nil {
			return nil, err
		}
		return uint64(ret), nil
	case string:
		return newValueStr, nil
	default:
		return nil, fmt.Errorf("type not support: %T", originValue)
	}
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
