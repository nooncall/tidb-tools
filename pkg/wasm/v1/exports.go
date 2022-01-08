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

func (a *ABIContext) CallWasmFunction(funcName string, args ...interface{}) (interface{}, Action, error) {
	ff, err := a.Instance.GetExportsFunc(funcName)
	if err != nil {
		return nil, ActionContinue, err
	}

	res, err := ff.Call(args...)
	if err != nil {
		a.Instance.HandleError(err)
		return nil, ActionContinue, err
	}

	// if we have sync call, e.g. HttpCall, then unlock the wasm instance and wait until it resp
	action := a.Imports.Wait()

	return res, action, nil
}

func (a *ABIContext) ProxyOnContextCreate(contextID int32, parentContextID int32) error {
	_, _, err := a.CallWasmFunction("proxy_on_context_create", contextID, parentContextID)
	if err != nil {
		return err
	}
	return nil
}

func (a *ABIContext) ProxyOnDone(contextID int32) (int32, error) {
	res, _, err := a.CallWasmFunction("proxy_on_done", contextID)
	if err != nil {
		return 0, err
	}
	return res.(int32), nil
}

func (a *ABIContext) ProxyOnLog(contextID int32) error {
	_, _, err := a.CallWasmFunction("proxy_on_log", contextID)
	if err != nil {
		return err
	}
	return nil
}

func (a *ABIContext) ProxyOnVmStart(rootContextID int32, vmConfigurationSize int32) (int32, error) {
	res, _, err := a.CallWasmFunction("proxy_on_vm_start", rootContextID, vmConfigurationSize)
	if err != nil {
		return 0, err
	}
	return res.(int32), nil
}

func (a *ABIContext) ProxyOnDelete(contextID int32) error {
	_, _, err := a.CallWasmFunction("proxy_on_delete", contextID)
	if err != nil {
		return err
	}
	return nil
}

func (a *ABIContext) ProxyOnConfigure(rootContextID int32, configurationSize int32) (int32, error) {
	res, _, err := a.CallWasmFunction("proxy_on_configure", rootContextID, configurationSize)
	if err != nil {
		return 0, err
	}
	return res.(int32), nil
}

func (a *ABIContext) ProxyOnTick(rootContextID int32) error {
	_, _, err := a.CallWasmFunction("proxy_on_tick", rootContextID)
	if err != nil {
		return err
	}
	return nil
}

func (a *ABIContext) ProxyOnRequestHeaders(contextID int32, numHeaders int32, endOfStream int32) (Action, error) {
	_, action, err := a.CallWasmFunction("proxy_on_request_headers", contextID, numHeaders, endOfStream)
	if err != nil {
		return ActionPause, err
	}
	return action, nil
}

func (a *ABIContext) ProxyOnRequestTrailers(contextID int32, trailers int32) (Action, error) {
	_, action, err := a.CallWasmFunction("proxy_on_request_trailers", contextID, trailers)
	if err != nil {
		return ActionPause, err
	}
	return action, nil
}

func (a *ABIContext) ProxyOnMemoryAllocate(size int32) (int32, error) {
	res, _, err := a.CallWasmFunction("proxy_on_memory_allocate", size)
	if err != nil {
		return 0, err
	}
	return res.(int32), nil
}
