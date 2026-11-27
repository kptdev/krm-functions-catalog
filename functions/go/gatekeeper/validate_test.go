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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kptdev/kpt/api/kubeobject"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestSortResults(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name   string
		input  kubeobject.Results
		output kubeobject.Results
	}{
		{
			name: "sort based on severity",
			input: kubeobject.Results{
				{
					Message:  "Error message 1",
					Severity: kubeobject.Info,
				},
				{
					Message:  "Error message 2",
					Severity: kubeobject.Error,
				},
			},
			output: kubeobject.Results{
				{
					Message:  "Error message 2",
					Severity: kubeobject.Error,
				},
				{
					Message:  "Error message 1",
					Severity: kubeobject.Info,
				},
			},
		},
		{
			name: "sort based on file",
			input: kubeobject.Results{
				{
					Message:  "Error message",
					Severity: kubeobject.Error,
					File: &kubeobject.File{
						Path:  "resource.yaml",
						Index: 1,
					},
				},
				{
					Message:  "Error message",
					Severity: kubeobject.Info,
					File: &kubeobject.File{
						Path:  "resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: kubeobject.Info,
					File: &kubeobject.File{
						Path:  "other-resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: kubeobject.Warning,
					File: &kubeobject.File{
						Path:  "resource.yaml",
						Index: 2,
					},
				},
			},
			output: kubeobject.Results{
				{
					Message:  "Error message",
					Severity: kubeobject.Info,
					File: &kubeobject.File{
						Path:  "other-resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: kubeobject.Info,
					File: &kubeobject.File{
						Path:  "resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: kubeobject.Error,
					File: &kubeobject.File{
						Path:  "resource.yaml",
						Index: 1,
					},
				},
				{
					Message:  "Error message",
					Severity: kubeobject.Warning,
					File: &kubeobject.File{
						Path:  "resource.yaml",
						Index: 2,
					},
				},
			},
		},

		{
			name: "sort based on other fields",
			input: kubeobject.Results{
				{
					Message:  "Error message",
					Severity: kubeobject.Error,
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &kubeobject.Field{
						Path: "spec",
					},
				},
				{
					Message:  "Error message",
					Severity: kubeobject.Error,
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &kubeobject.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Another error message",
					Severity: kubeobject.Error,
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &kubeobject.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Another error message",
					Severity: kubeobject.Error,
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "ConfigMap",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &kubeobject.Field{
						Path: "metadata.name",
					},
				},
			},
			output: kubeobject.Results{
				{
					Message:  "Another error message",
					Severity: kubeobject.Error,
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "ConfigMap",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &kubeobject.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Another error message",
					Severity: kubeobject.Error,
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &kubeobject.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Error message",
					Severity: kubeobject.Error,
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &kubeobject.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Error message",
					Severity: kubeobject.Error,
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: "v1",
							Kind:       "Pod",
						},
						NameMeta: yaml.NameMeta{
							Namespace: "foo-ns",
							Name:      "bar",
						},
					},
					Field: &kubeobject.Field{
						Path: "spec",
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		tc.input.Sort()
		if diff := cmp.Diff(tc.output, tc.input); diff != "" {
			t.Errorf("testcase %q mismatch (-want +got):\n%s", tc.name, diff)
		}
	}
}
