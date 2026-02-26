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
	"reflect"
	"testing"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestSortResultItems(t *testing.T) {
	testcases := []struct {
		name   string
		input  []framework.ResultItem
		output []framework.ResultItem
	}{
		{
			name: "sort based on severity",
			input: []framework.ResultItem{
				{
					Message:  "Error message 1",
					Severity: framework.Info,
				},
				{
					Message:  "Error message 2",
					Severity: framework.Error,
				},
			},
			output: []framework.ResultItem{
				{
					Message:  "Error message 2",
					Severity: framework.Error,
				},
				{
					Message:  "Error message 1",
					Severity: framework.Info,
				},
			},
		},
		{
			name: "sort based on file",
			input: []framework.ResultItem{
				{
					Message:  "Error message",
					Severity: framework.Error,
					File: framework.File{
						Path:  "resource.yaml",
						Index: 1,
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Info,
					File: framework.File{
						Path:  "resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Info,
					File: framework.File{
						Path:  "other-resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Warning,
					File: framework.File{
						Path:  "resource.yaml",
						Index: 2,
					},
				},
			},
			output: []framework.ResultItem{
				{
					Message:  "Error message",
					Severity: framework.Info,
					File: framework.File{
						Path:  "other-resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Info,
					File: framework.File{
						Path:  "resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Error,
					File: framework.File{
						Path:  "resource.yaml",
						Index: 1,
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Warning,
					File: framework.File{
						Path:  "resource.yaml",
						Index: 2,
					},
				},
			},
		},

		{
			name: "sort based on other fields",
			input: []framework.ResultItem{
				{
					Message:  "Error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: framework.Field{
						Path: "spec",
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: framework.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Another error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: framework.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Another error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "ConfigMap",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: framework.Field{
						Path: "metadata.name",
					},
				},
			},
			output: []framework.ResultItem{
				{
					Message:  "Another error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "ConfigMap",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: framework.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Another error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: framework.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: framework.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Error message",
					Severity: framework.Error,
					ResourceRef: yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: framework.Field{
						Path: "spec",
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		sortResultItems(tc.input)
		if !reflect.DeepEqual(tc.input, tc.output) {
			t.Errorf("in testcase %q, expect: %#v, but got: %#v", tc.name, tc.output, tc.input)
		}
	}
}
