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

	"github.com/kptdev/krm-functions-catalog/functions/go/create-setters/createsetters"
	"github.com/kptdev/krm-functions-catalog/functions/go/create-setters/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// nolint
func main() {
	csp := CreateSettersProcessor{}
	cmd := command.Build(&csp, command.StandaloneEnabled, false)

	cmd.Short = generated.CreateSettersShort
	cmd.Long = generated.CreateSettersLong
	cmd.Example = generated.CreateSettersExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type CreateSettersProcessor struct{}

func (csp *CreateSettersProcessor) Process(resourceList *framework.ResourceList) error {
	resourceList.Result = &framework.Result{
		Name: "create-setters",
	}
	items, err := run(resourceList)
	if err != nil {
		resourceList.Result.Items = getErrorItem(err.Error())
		return err
	}
	resourceList.Result.Items = items
	return nil
}

func run(resourceList *framework.ResourceList) ([]framework.ResultItem, error) {
	s, err := getSetters(resourceList.FunctionConfig)
	if err != nil {
		return nil, err
	}
	_, err = s.Filter(resourceList.Items)
	if err != nil {
		return nil, err
	}
	resultItems, err := resultsToItems(s)
	if err != nil {
		return nil, err
	}
	return resultItems, nil
}

// getSetters retrieve the setters from input config
func getSetters(fc *kyaml.RNode) (createsetters.CreateSetters, error) {
	var fcd createsetters.CreateSetters
	err := createsetters.Decode(fc, &fcd)
	return fcd, err
}

// resultsToItems converts the create-setters results to
// equivalent items([]framework.Item)
func resultsToItems(sr createsetters.CreateSetters) ([]framework.ResultItem, error) {
	var items []framework.ResultItem
	if len(sr.Results) == 0 {
		return nil, fmt.Errorf("no matches for the input list of setters")
	}
	for _, res := range sr.Results {
		items = append(items, framework.ResultItem{
			Message: fmt.Sprintf("Added line comment %q for field with value %q", res.Comment, res.Value),
			Field:   framework.Field{Path: res.FieldPath},
			File:    framework.File{Path: res.FilePath},
		})
	}
	return items, nil
}

// getErrorItem returns the item for input error message
func getErrorItem(errMsg string) []framework.ResultItem {
	return []framework.ResultItem{
		{
			Message:  fmt.Sprintf("failed to create setters: %s", errMsg),
			Severity: framework.Error,
		},
	}
}
