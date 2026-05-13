// Copyright 2026 The kpt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package generate_folders transforms ResourceHierarchy resources into Folders.
package generate_folders

import (
	"fmt"

	"github.com/kptdev/krm-functions-sdk/go/fn"
)

// Run processes ResourceHierarchy resources and generates Folders.
func Run(rl *fn.ResourceList) (bool, error) {
	for _, obj := range rl.Items {
		apiVersion := obj.GetAPIVersion()
		kind := obj.GetKind()

		if kind != resourceHierarchyKind {
			continue
		}

		switch apiVersion {
		case v1APIVersion:
			results := processV1Hierarchy(obj, rl)
			rl.Results = append(rl.Results, results...)
		case v2APIVersion:
			rl.Results = append(rl.Results, oldHierarchyWarning(obj))
			fallthrough
		case v3APIVersion:
			results := processV2V3Hierarchy(obj, rl, apiVersion == v3APIVersion)
			rl.Results = append(rl.Results, results...)
		default:
			rl.Results = append(rl.Results, fn.GeneralResult(fmt.Sprintf(
				"ResourceHierarchy %s has unrecognized apiVersion %q; supported versions: %s, %s, %s",
				obj.GetName(), apiVersion, v3APIVersion, v2APIVersion, v1APIVersion), fn.Warning))
		}
	}

	return true, nil
}

// oldHierarchyWarning generates a deprecation warning for v1/v2.
func oldHierarchyWarning(obj *fn.KubeObject) *fn.Result {
	return fn.GeneralResult(
		fmt.Sprintf("ResourceHierarchy %s references an older Resource Hierarchy GroupVersion. Latest GroupVersion is blueprints.cloud.google.com/v1alpha3.", obj.GetName()),
		fn.Warning,
	)
}
