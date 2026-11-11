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
	"github.com/kptdev/krm-functions-catalog/functions/go/apply-setters/applysetters"
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
	// ApplySetters.Filter mutates the RNodes in place, so capture the filtered
	// nodes and move them back into the ResourceList to persist the applied
	// setter values (the Items here are copies and are not mutated otherwise).
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
func getSetters(fc *kyaml.RNode) (applysetters.ApplySetters, error) {
	var fcd applysetters.ApplySetters
	applysetters.Decode(fc, &fcd)
	return fcd, nil
}

// resultsToItems converts the Search and Replace results to
// equivalent items([]*kubeobject.Result)
func resultsToItems(sr applysetters.ApplySetters) ([]*kubeobject.Result, error) {
	var items []*kubeobject.Result
	if len(sr.Results) == 0 {
		return []*kubeobject.Result{{Message: "no matches for input setter(s)"}}, nil
	}
	for _, res := range sr.Results {
		items = append(items, &kubeobject.Result{
			Message:     res.Message,
			Severity:    res.Severity,
			File:        res.File,
			ResourceRef: res.ResourceRef,
			Field: &kubeobject.Field{
				Path:          res.Field.Path,
				CurrentValue:  res.Field.CurrentValue,
				ProposedValue: res.Field.ProposedValue,
			},
		})
	}
	return items, nil
}

// getErrorItem returns the item for input error message
func getErrorItem(errMsg string) []*kubeobject.Result {
	return []*kubeobject.Result{
		{
			Message:  fmt.Sprintf("failed to apply setters: %s", errMsg),
			Severity: kubeobject.Error,
		},
	}
}
