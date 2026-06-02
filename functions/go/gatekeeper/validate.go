// Copyright 2019, 2026 The kpt Authors
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
	"strconv"

	// The gatekeeper/v3/pkg/gator/test package is the underlying libraries for
	// the `gator test` subcommand, not a library for testing golang code.
	gatortest "github.com/open-policy-agent/gatekeeper/v3/pkg/gator/test"
	opautil "github.com/open-policy-agent/gatekeeper/v3/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Validate makes sure the configs passed to it comply with any Constraints and
// Constraint Templates present in the list of configs
func Validate(objects []*unstructured.Unstructured) (framework.Results, error) {
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

func parseResults(results []*gatortest.GatorResult) (framework.Results, error) {
	var items framework.Results

	for _, r := range results {
		u := r.ViolatingObject
		if u == nil {
			continue
		}

		item := &framework.Result{
			Message: fmt.Sprintf("%s\nviolatedConstraint: %s", r.Msg, r.Constraint.GetName()),
			ResourceRef: &yaml.ResourceIdentifier{
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
			item.File = &framework.File{
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
	items.Sort()

	return items, nil
}
