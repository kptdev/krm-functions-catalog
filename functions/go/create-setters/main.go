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
	"github.com/kptdev/krm-functions-catalog/functions/go/create-setters/createsetters"
	"github.com/kptdev/krm-functions-sdk/go/fn"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
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
	items, err := run(resourceList)
	if err != nil {
		resourceList.Results = getErrorItem(err.Error())
		return false, err
	}
	resourceList.Results = items
	return true, nil
}

func run(resourceList *kubeobject.ResourceList) ([]*kubeobject.Result, error) {
	s, err := getSetters(resourceList.FunctionConfig.CopyToResourceNode())
	if err != nil {
		return nil, err
	}
	filtered, err := s.Filter(resourceList.Items.CopyToResourceNodes())
	if err != nil {
		return nil, err
	}
	resourceList.Items = kubeobject.MoveToKubeObjects(filtered)
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
// equivalent items([]*kubeobject.Result)
func resultsToItems(sr createsetters.CreateSetters) ([]*kubeobject.Result, error) {
	var items []*kubeobject.Result
	if len(sr.Results) == 0 {
		return nil, fmt.Errorf("no matches for the input list of setters")
	}
	for _, res := range sr.Results {
		items = append(items, &kubeobject.Result{
			Message: fmt.Sprintf("Added line comment %q for field with value %q", res.Comment, res.Value),
			Field:   &kubeobject.Field{Path: res.FieldPath},
			File:    &kubeobject.File{Path: res.FilePath},
		})
	}
	return items, nil
}

// getErrorItem returns the item for input error message
func getErrorItem(errMsg string) []*kubeobject.Result {
	return []*kubeobject.Result{
		{
			Message:  fmt.Sprintf("failed to create setters: %s", errMsg),
			Severity: kubeobject.Error,
		},
	}
}
