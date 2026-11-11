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
	"github.com/kptdev/krm-functions-catalog/functions/go/search-replace/searchreplace"
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
	results, err := run(resourceList)
	if err != nil {
		resourceList.Results = getErrorResult(err.Error())
		return false, err
	}
	resourceList.Results = results
	return true, nil
}

// run resolves the function params from input ResourceList and runs the function on resources
func run(resourceList *kubeobject.ResourceList) ([]*kubeobject.Result, error) {
	sr, err := getSearchReplaceParams(resourceList.FunctionConfig.CopyToResourceNode())
	if err != nil {
		return nil, err
	}

	// SearchReplace.Filter mutates the RNodes in place when a put-* parameter is
	// set (replace mode), so capture the filtered nodes and move them back into
	// the ResourceList to persist the replacements (the Items here are copies
	// and are not mutated otherwise). In search-only mode the nodes are
	// unchanged, so moving them back is a harmless round-trip.
	filtered, err := sr.Filter(resourceList.Items.CopyToResourceNodes())
	if err != nil {
		return nil, err
	}
	resourceList.Items = kubeobject.MoveToKubeObjects(filtered)

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
// equivalent items([]*kubeobject.Result)
func searchResultsToItems(sr searchreplace.SearchReplace) []*kubeobject.Result {
	var results []*kubeobject.Result
	if len(sr.Results) == 0 {
		results = append(results, &kubeobject.Result{
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

		results = append(results, &kubeobject.Result{
			Message: message,
			Field:   &kubeobject.Field{Path: res.FieldPath},
			File:    &kubeobject.File{Path: res.FilePath},
		})
	}
	return results
}

// getErrorResult returns the result for input error message
func getErrorResult(errMsg string) []*kubeobject.Result {
	return []*kubeobject.Result{
		{
			Message:  fmt.Sprintf("failed to perform search-replace operation: %q", errMsg),
			Severity: kubeobject.Error,
		},
	}
}
