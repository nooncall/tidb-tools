/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1

import (
	"fmt"
	"time"

	"github.com/pingcap/tidb-tools/pkg/wasm/common"
)

type DefaultImportsHandler struct{}

func (d *DefaultImportsHandler) Wait() Action { return ActionContinue }

func (d *DefaultImportsHandler) GetRootContextID() int32 { return 0 }

func (d *DefaultImportsHandler) GetVmConfig() common.IoBuffer { return nil }

func (d *DefaultImportsHandler) GetPluginConfig() common.IoBuffer { return nil }

func (d *DefaultImportsHandler) Log(level LogLevel, msg string) WasmResult {
	fmt.Println(msg)
	return WasmResultOk
}

func (d *DefaultImportsHandler) SetEffectiveContextID(contextID int32) WasmResult {
	return WasmResultUnimplemented
}

func (d *DefaultImportsHandler) SetTickPeriodMilliseconds(tickPeriodMilliseconds int32) WasmResult {
	return WasmResultUnimplemented
}

func (d *DefaultImportsHandler) GetCurrentTimeNanoseconds() (int32, WasmResult) {
	nano := time.Now().Nanosecond()
	return int32(nano), WasmResultOk
}

func (d *DefaultImportsHandler) Done() WasmResult { return WasmResultUnimplemented }

func (d *DefaultImportsHandler) GetHttpRequestHeader() common.HeaderMap { return nil }

func (d *DefaultImportsHandler) GetHttpRequestTrailer() common.HeaderMap { return nil }

func (d *DefaultImportsHandler) GetCustomBuffer(bufferType BufferType) common.IoBuffer { return nil }

func (d *DefaultImportsHandler) GetCustomHeader(mapType MapType) common.HeaderMap { return nil }
