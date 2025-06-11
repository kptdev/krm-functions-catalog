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

//go:build !(js && wasm)

package main

import (
	"context"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/kptdev/krm-functions-catalog/functions/go/set-labels/setlabels"
)

func NewTransformer() fn.ResourceListProcessor {
	// Create a new SetLabels struct
	labels := make(map[string]string)
	options := map[string]bool{
		"setSelectorLabels": true,
	}
	setLabels := &setlabels.SetLabels{
		Labels:  labels,
		Options: options,
	}

	// Return the SetLabels struct as a fn.ResourceListProcessor
	return fn.WithContext(fn.Context{Context: context.Background()}, setLabels)
}

func run() error {
	return fn.AsMain(NewTransformer())
}
