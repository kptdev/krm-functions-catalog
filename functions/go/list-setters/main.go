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
	"github.com/kptdev/krm-functions-catalog/functions/go/list-setters/listsetters"
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
	results, err := run(resourceList)
	if err != nil {
		resourceList.Results = getErrorItem(fmt.Sprintf("failed to list setters: %s", err.Error()), kubeobject.Error)
		return false, err
	}
	resourceList.Results = results
	return true, nil
}

func run(resourceList *kubeobject.ResourceList) ([]*kubeobject.Result, error) {
	ls := listsetters.New()
	// list-setters only discovers and reports setters; Filter does not mutate
	// the nodes, so the filtered result is intentionally discarded.
	_, err := ls.Filter(resourceList.Items.CopyToResourceNodes())
	if err != nil {
		return nil, err
	}
	resultItems, err := resultsToItems(ls)
	if err != nil {
		return nil, err
	}
	return resultItems, nil
}

// resultsToItems converts the listsetters results to
// equivalent items([]*kubeobject.Result)
func resultsToItems(sr listsetters.ListSetters) ([]*kubeobject.Result, error) {
	var results []*kubeobject.Result
	rs := sr.GetResults()
	if len(rs) == 0 {
		return getErrorItem("no setters found", kubeobject.Warning), nil
	}
	for _, r := range rs {
		results = append(results, &kubeobject.Result{
			Message: r.String(),
		})
	}
	return results, nil
}

// getErrorItem returns the item for an error message
func getErrorItem(errMsg string, severity kubeobject.Severity) []*kubeobject.Result {
	return []*kubeobject.Result{
		{
			Message:  errMsg,
			Severity: severity,
		},
	}
}
