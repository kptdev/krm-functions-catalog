// Copyright 2019 The kpt Authors
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
	"sort"
	"strconv"

	opatypes "github.com/open-policy-agent/frameworks/constraint/pkg/types"
	// The gatekeeper/pkg/gator/test package is the underlying libraries for
	// the `gator test` subcommand, not a library for testing golang code.
	gatortest "github.com/open-policy-agent/gatekeeper/pkg/gator/test"
	opautil "github.com/open-policy-agent/gatekeeper/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Validate makes sure the configs passed to it comply with any Constraints and
// Constraint Templates present in the list of configs
func Validate(objects []*unstructured.Unstructured) (*framework.Result, error) {
	resps, err := gatortest.Test(objects)
	if err != nil {
		return nil, err
	}

	results := resps.Results()
	if len(results) > 0 {
		return parseResults(results)
	}
	return nil, nil
}

func parseResults(results []*opatypes.Result) (*framework.Result, error) {
	var items []framework.ResultItem

	for _, r := range results {
		u, ok := r.Resource.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("could not cast to unstructured: %+v", r.Resource)
		}

		item := framework.ResultItem{
			Message: fmt.Sprintf("%s\nviolatedConstraint: %s", r.Msg, r.Constraint.GetName()),
			ResourceRef: yaml.ResourceIdentifier{
				TypeMeta: yaml.TypeMeta{
					APIVersion: u.GetAPIVersion(),
					Kind:       u.GetKind(),
				},
				NameMeta: yaml.NameMeta{
					Name:      u.GetName(),
					Namespace: u.GetNamespace(),
				},
			},
		}

		switch r.EnforcementAction {
		case string(opautil.Dryrun):
			item.Severity = framework.Info
		case string(opautil.Warn):
			item.Severity = framework.Warning
		default:
			item.Severity = framework.Error
		}

		path, foundPath := u.GetAnnotations()[kioutil.PathAnnotation]
		index, foundIndex := u.GetAnnotations()[kioutil.IndexAnnotation]
		if foundPath {
			item.File = framework.File{
				Path: path,
			}
			if foundIndex {
				idx, err := strconv.Atoi(index)
				if err != nil {
					return nil, err
				}
				item.File.Index = idx
			}
		}

		items = append(items, item)
	}
	sortResultItems(items)

	return &framework.Result{
		Items: items,
	}, nil
}

// TODO(mengqiy): upstream this to the SDK
func sortResultItems(items []framework.ResultItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if fileLess(items, i, j) != 0 {
			return fileLess(items, i, j) < 0
		}
		if severityLess(items, i, j) != 0 {
			return severityLess(items, i, j) < 0
		}
		return resultItemToString(items[i]) < resultItemToString(items[j])
	})
}

func severityLess(items []framework.ResultItem, i, j int) int {
	severityToNumber := map[framework.Severity]int{
		framework.Error:   0,
		framework.Warning: 1,
		framework.Info:    2,
	}

	severityLevelI, found := severityToNumber[items[i].Severity]
	if !found {
		severityLevelI = 3
	}
	severityLevelJ, found := severityToNumber[items[j].Severity]
	if !found {
		severityLevelJ = 3
	}
	return severityLevelI - severityLevelJ
}

func fileLess(items []framework.ResultItem, i, j int) int {
	if items[i].File.Path != items[j].File.Path {
		if items[i].File.Path < items[j].File.Path {
			return -1
		} else {
			return 1
		}
	}
	return items[i].File.Index - items[j].File.Index
}

func resultItemToString(item framework.ResultItem) string {
	return fmt.Sprintf("resource-ref:%s,field:%s,message:%s",
		item.ResourceRef, item.Field, item.Message)
}
