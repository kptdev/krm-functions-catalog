// Copyright 2026 The kpt Authors
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
	_ "embed"
	"fmt"
	"os"

	"github.com/kptdev/kpt/api/kubeobject"
	"github.com/kptdev/krm-functions-sdk/go/fn"
)

//go:embed README.md
var readme []byte

//go:embed metadata.yaml
var metadata []byte

// nolint
func main() {
	if err := fn.AsMain(kubeobject.ResourceListProcessorFunc(Process), fn.WithDocs(readme, metadata)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func Process(resourceList *kubeobject.ResourceList) (bool, error) {
	results, err := processResources(resourceList)
	if err != nil {
		resourceList.Results = getErrorResult(err.Error())
		return false, err
	}
	resourceList.Results = results
	return true, nil
}

func processResources(resourceList *kubeobject.ResourceList) (kubeobject.Results, error) {
	var resultItems kubeobject.Results
	var res []*kubeobject.KubeObject
	for _, node := range resourceList.Items {
		if node == nil || node.IsEmpty() {
			continue
		}
		if !node.IsLocalConfig() {
			res = append(res, node)
		} else {
			itemFilePath := node.GetAnnotations()["internal.config.kubernetes.io/path"]
			if itemFilePath == "" {
				itemFilePath = node.GetAnnotations()["config.kubernetes.io/path"]
			}

			resultItems = append(resultItems, &kubeobject.Result{
				Message: fmt.Sprintf("Resource name: [%s]", node.GetName()),
				File: &kubeobject.File{
					Path: itemFilePath,
				},
				Severity: kubeobject.Info,
			})
		}
	}

	resourceList.Items = res

	if len(resultItems) > 0 {
		infoResultSlice := kubeobject.Results{
			&kubeobject.Result{
				Severity: kubeobject.Info,
				Message:  fmt.Sprintf("Number of resources pruned: %d", len(resultItems)),
			},
		}

		resultItems = append(infoResultSlice, resultItems...)
	} else if len(resultItems) == 0 {
		resultItems = append(resultItems, &kubeobject.Result{
			Message:  "Found no resources to prune with the local config annotation",
			Severity: kubeobject.Warning,
		})
	}

	return resultItems, nil
}

// getErrorResult returns the result for an error message
func getErrorResult(errMsg string) kubeobject.Results {
	return kubeobject.Results{
		&kubeobject.Result{
			Message:  fmt.Sprintf("failed to remove local configs: %s", errMsg),
			Severity: kubeobject.Error,
		},
	}
}
