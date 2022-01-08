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
	"github.com/pingcap/tidb-tools/pkg/wasm/common"
)

func RegisterImports(instance common.WasmInstance) {
	_ = instance.RegisterFunc("env", "proxy_log", ProxyLog)

	_ = instance.RegisterFunc("env", "proxy_set_effective_context", ProxySetEffectiveContext)

	_ = instance.RegisterFunc("env", "proxy_get_header_map_pairs", ProxyGetHeaderMapPairs)
	_ = instance.RegisterFunc("env", "proxy_set_header_map_pairs", ProxySetHeaderMapPairs)

	_ = instance.RegisterFunc("env", "proxy_get_header_map_value", ProxyGetHeaderMapValue)
	_ = instance.RegisterFunc("env", "proxy_replace_header_map_value", ProxyReplaceHeaderMapValue)
	_ = instance.RegisterFunc("env", "proxy_add_header_map_value", ProxyAddHeaderMapValue)
	_ = instance.RegisterFunc("env", "proxy_remove_header_map_value", ProxyRemoveHeaderMapValue)
}

func ProxyLog(instance common.WasmInstance, level int32, logDataPtr int32, logDataSize int32) int32 {
	logContent, err := instance.GetMemory(uint64(logDataPtr), uint64(logDataSize))
	if err != nil {
		return WasmResultInvalidMemoryAccess.Int32()
	}

	callback := getImportHandler(instance)

	return callback.Log(LogLevel(level), string(logContent)).Int32()
}

func ProxySetEffectiveContext(instance common.WasmInstance, contextID int32) int32 {
	ctx := getImportHandler(instance)
	return ctx.SetEffectiveContextID(contextID).Int32()
}
