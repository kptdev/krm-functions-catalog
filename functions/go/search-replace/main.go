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

	"github.com/kptdev/krm-functions-catalog/functions/go/search-replace/generated"
	"github.com/kptdev/krm-functions-catalog/functions/go/search-replace/searchreplace"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// nolint
func main() {
	srp := SearchReplaceProcessor{}
	cmd := command.Build(&srp, command.StandaloneEnabled, false)

	cmd.Short = generated.SearchReplaceShort
	cmd.Long = generated.SearchReplaceLong
	cmd.Example = generated.SearchReplaceExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type SearchReplaceProcessor struct{}

func (srp *SearchReplaceProcessor) Process(resourceList *framework.ResourceList) error {
	results, err := run(resourceList)
	if err != nil {
		resourceList.Results = getErrorResult(err.Error())
		return err
	}
	resourceList.Results = results
	return nil
}

// run resolves the function params from input ResourceList and runs the function on resources
func run(resourceList *framework.ResourceList) (framework.Results, error) {
	sr, err := getSearchReplaceParams(resourceList.FunctionConfig)
	if err != nil {
		return nil, err
	}

	_, err = sr.Filter(resourceList.Items)
	if err != nil {
		return nil, err
	}

	return searchResultsToItems(sr), nil
}

// getSearchReplaceParams retrieve the search parameters from input config
func getSearchReplaceParams(fc *kyaml.RNode) (searchreplace.SearchReplace, error) {
	var fcd searchreplace.SearchReplace
	if err := searchreplace.Decode(fc, &fcd); err != nil {
		return fcd, err
	}
	return fcd, nil
}

// searchResultsToItems converts the Search and Replace results to
// equivalent items(framework.Results)
func searchResultsToItems(sr searchreplace.SearchReplace) framework.Results {
	var results framework.Results
	if len(sr.Results) == 0 {
		results = append(results, &framework.Result{
			Message: "no matches",
		})
		return results
	}
	for _, res := range sr.Results {
		var message string
		if sr.PutComment != "" || sr.PutValue != "" {
			message = fmt.Sprintf("Mutated field value to %q", res.Value)
		} else {
			message = fmt.Sprintf("Matched field value %q", res.Value)
		}

		results = append(results, &framework.Result{
			Message: message,
			Field:   &framework.Field{Path: res.FieldPath},
			File:    &framework.File{Path: res.FilePath},
		})
	}
	return results
}

// getErrorResult returns the result for input error message
func getErrorResult(errMsg string) framework.Results {
	return framework.Results{
		&framework.Result{
			Message:  fmt.Sprintf("failed to perform search-replace operation: %q", errMsg),
			Severity: framework.Error,
		},
	}
}
