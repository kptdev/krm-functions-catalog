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

	"github.com/kptdev/krm-functions-catalog/functions/go/apply-setters/applysetters"
	"github.com/kptdev/krm-functions-catalog/functions/go/apply-setters/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// nolint
func main() {
	asp := ApplySettersProcessor{}
	cmd := command.Build(&asp, command.StandaloneEnabled, false)

	cmd.Short = generated.ApplySettersShort
	cmd.Long = generated.ApplySettersLong
	cmd.Example = generated.ApplySettersExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type ApplySettersProcessor struct{}

func (asp *ApplySettersProcessor) Process(resourceList *framework.ResourceList) error {
	items, err := run(resourceList)
	if err != nil {
		resourceList.Results = getErrorItem(err.Error())
		return err
	}
	resourceList.Results = items
	return nil
}

func run(resourceList *framework.ResourceList) ([]*framework.Result, error) {
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
func getSetters(fc *kyaml.RNode) (applysetters.ApplySetters, error) {
	var fcd applysetters.ApplySetters
	applysetters.Decode(fc, &fcd)
	return fcd, nil
}

// resultsToItems converts the Search and Replace results to
// equivalent items([]*framework.Result)
func resultsToItems(sr applysetters.ApplySetters) ([]*framework.Result, error) {
	var items []*framework.Result
	if len(sr.Results) == 0 {
		return []*framework.Result{{Message: "no matches for input setter(s)"}}, nil
	}
	for _, res := range sr.Results {
		items = append(items, &framework.Result{
			Message:  res.Message,
			Severity: framework.Severity(res.Severity),
			File:     &framework.File{Path: res.File.Path},
			ResourceRef: &kyaml.ResourceIdentifier{
				TypeMeta: kyaml.TypeMeta{
					APIVersion: res.ResourceRef.APIVersion,
					Kind:       res.ResourceRef.Kind,
				},
				NameMeta: kyaml.NameMeta{
					Name:      res.ResourceRef.Name,
					Namespace: res.ResourceRef.Namespace,
				},
			},
			Field: &framework.Field{
				Path:          res.Field.Path,
				CurrentValue:  res.Field.CurrentValue,
				ProposedValue: res.Field.ProposedValue,
			},
		})
	}
	return items, nil
}

// getErrorItem returns the item for input error message
func getErrorItem(errMsg string) []*framework.Result {
	return []*framework.Result{
		{
			Message:  fmt.Sprintf("failed to apply setters: %s", errMsg),
			Severity: framework.Error,
		},
	}
}
