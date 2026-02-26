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
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// nolint
func main() {
	fp := RemoveLocalConfigResourcesConfigProcessor{}
	cmd := command.Build(&fp, command.StandaloneEnabled, false)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func (fp *RemoveLocalConfigResourcesConfigProcessor) Process(resourceList *framework.ResourceList) error {
	resourceList.Result = &framework.Result{
		Name: "remove-local-config-resources",
	}

	items, err := processResources(resourceList)
	if err != nil {
		resourceList.Result.Items = getErrorItem(err.Error())
		return err
	}
	resourceList.Result.Items = items
	return nil
}

func processResources(resourceList *framework.ResourceList) ([]framework.ResultItem, error) {
	var resultItems []framework.ResultItem
	var res []*yaml.RNode
	for _, node := range resourceList.Items {
		if node.IsNilOrEmpty() {
			continue
		}
		// only add the resources which are not local configs
		if strings.ToLower(node.GetAnnotations()[filters.LocalConfigAnnotation]) != "true" {
			res = append(res, node)
		} else {
			itemFilePath := node.GetAnnotations()["internal.config.kubernetes.io/path"]
			if itemFilePath == "" {
				itemFilePath = node.GetAnnotations()["config.kubernetes.io/path"]
			}

			resultItems = append(resultItems, framework.ResultItem{
				Message: fmt.Sprintf("Resource name: [%s]", node.GetName()),
				File: framework.File{
					Path: itemFilePath,
				},
				Severity: framework.Info,
			})
		}
	}

	resourceList.Items = res

	if len(resultItems) > 0 {
		infoResultSlice := []framework.ResultItem{}
		infoResultSlice = append(infoResultSlice, framework.ResultItem{
			Severity: framework.Info,
			Message:  fmt.Sprintf("Number of resources pruned: %d", len(resultItems)),
		})

		resultItems = append(infoResultSlice, resultItems...)
	} else if len(resultItems) == 0 {
		resultItems = append(resultItems, framework.ResultItem{
			Message:  "Found no resources to prune with the local config annotation",
			Severity: framework.Warning,
		})
	}

	return resultItems, nil
}

// getErrorItem returns the item for an error message
func getErrorItem(errMsg string) []framework.ResultItem {
	return []framework.ResultItem{
		{
			Message:  fmt.Sprintf("failed to remove local configs: %s", errMsg),
			Severity: framework.Error,
		},
	}
}

type RemoveLocalConfigResourcesConfigProcessor struct{}
