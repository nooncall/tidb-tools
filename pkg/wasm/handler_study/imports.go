package handler_study

import (
	"github.com/pingcap/tidb-tools/pkg/wasm/common"
	v1 "github.com/pingcap/tidb-tools/pkg/wasm/v1"
)

type ImportsHandlerImpl struct {
	v1.DefaultImportsHandler
}

func NewImportsHandlerImpl() *ImportsHandlerImpl {
	return &ImportsHandlerImpl{}
}

func (h *ImportsHandlerImpl) Log(level v1.LogLevel, msg string) v1.WasmResult {
	//TODO implement me
	panic("implement me")
}

func (h *ImportsHandlerImpl) GetRootContextID() int32 {
	//TODO implement me
	panic("implement me")
}

func (h *ImportsHandlerImpl) SetEffectiveContextID(contextID int32) v1.WasmResult {
	//TODO implement me
	panic("implement me")
}

func (h *ImportsHandlerImpl) SetTickPeriodMilliseconds(tickPeriodMilliseconds int32) v1.WasmResult {
	//TODO implement me
	panic("implement me")
}

func (h *ImportsHandlerImpl) GetCurrentTimeNanoseconds() (int32, v1.WasmResult) {
	//TODO implement me
	panic("implement me")
}

func (h *ImportsHandlerImpl) Done() v1.WasmResult {
	//TODO implement me
	panic("implement me")
}

func (h *ImportsHandlerImpl) GetProperty(key string) (string, v1.WasmResult) {
	//TODO implement me
	panic("implement me")
}

func (h *ImportsHandlerImpl) SetProperty(key string, value string) v1.WasmResult {
	//TODO implement me
	panic("implement me")
}

func (h *ImportsHandlerImpl) CallForeignFunction(funcName string, param []byte) ([]byte, v1.WasmResult) {
	//TODO implement me
	panic("implement me")
}

func (h *ImportsHandlerImpl) GetFuncCallData() common.IoBuffer {
	//TODO implement me
	panic("implement me")
}
