// Copyright 2022 Google LLC
// Modifications Copyright (C) 2025 OpenInfra Foundation Europe.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/kptdev/krm-functions-catalog/functions/go/set-labels/setlabels"
	"github.com/kptdev/krm-functions-sdk/go/fn"
)

func run() error {
	resourceList := []byte("")

	js.Global().Set("processResourceList", resourceListProcessorWrapper(&resourceList))
	// Provide a second function that serves purely to also return the resourceList,
	// in case of the above function failing.
	js.Global().Set("processResourceListErrors", resourceListProcessorErrors(&resourceList))
	// We need to ensure that the Go program is running when JavaScript calls it.
	// Otherwise, it will complain the Go program has already exited.
	<-make(chan bool)
	return nil
}

func transformLabels(input []byte) ([]byte, error) {
	runner := fn.WithContext(fn.Context{Context: nil}, &setlabels.SetLabels{})
	return fn.Run(runner, []byte(input))
}

// This funcion will return ALL Results with Severity error,
// meaning unrelated errors may also be included.
func resourceListProcessorWrapper(resourceList *[]byte) js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 {
			return "Invalid number of arguments passed"
		}
		input := args[0].String()
		applied, err := transformLabels([]byte(input))
		if err != nil {
			*resourceList = applied
			return "unable to process resource list: " + err.Error()
		}
		*resourceList = applied
		return string(applied)
	})

	return jsonFunc
}

func resourceListProcessorErrors(resourceList *[]byte) js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		rl, err := fn.ParseResourceList(*resourceList)
		if err != nil {
			return ""
		}
		if len(rl.Results) == 0 {
			return ""
		}
		errorMessages := ""
		for _, r := range rl.Results {
			if r.Severity == "error" {
				errorMessages += r.Message
			}
		}
		return errorMessages
	})
	return jsonFunc
}
