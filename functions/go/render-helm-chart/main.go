// Copyright 2021 Google LLC
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

package main

import (
	"os"

	"github.com/kptdev/krm-functions-catalog/functions/go/render-helm-chart/helmfn"
	"github.com/kptdev/krm-functions-sdk/go/fn"
)

func main() {
	if err := fn.AsMain(fn.ResourceListProcessorFunc(helmfn.RenderHelmChart)); err != nil {
		os.Exit(1)
	}
}
