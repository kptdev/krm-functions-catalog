// Copyright 2021 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fnsdk_test

import (
	"reflect"
	"testing"

	"github.com/kptdev/krm-functions-catalog/archived/functions/go/export-terraform/thirdparty/kyaml/fnsdk"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestResults_Sort(t *testing.T) {
	testcases := []struct {
		name   string
		input  fnsdk.Results
		output fnsdk.Results
	}{
		{
			name: "sort based on severity",
			input: fnsdk.Results{
				{
					Message:  "Error message 1",
					Severity: fnsdk.Info,
				},
				{
					Message:  "Error message 2",
					Severity: fnsdk.Error,
				},
			},
			output: fnsdk.Results{
				{
					Message:  "Error message 2",
					Severity: fnsdk.Error,
				},
				{
					Message:  "Error message 1",
					Severity: fnsdk.Info,
				},
			},
		},
		{
			name: "sort based on file",
			input: fnsdk.Results{
				{
					Message:  "Error message",
					Severity: fnsdk.Error,
					File: &fnsdk.File{
						Path:  "resource.yaml",
						Index: 1,
					},
				},
				{
					Message:  "Error message",
					Severity: fnsdk.Info,
					File: &fnsdk.File{
						Path:  "resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: fnsdk.Info,
					File: &fnsdk.File{
						Path:  "other-resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: fnsdk.Warning,
					File: &fnsdk.File{
						Path:  "resource.yaml",
						Index: 2,
					},
				},
				{
					Message:  "Error message",
					Severity: fnsdk.Warning,
				},
			},
			output: fnsdk.Results{
				{
					Message:  "Error message",
					Severity: fnsdk.Warning,
				},
				{
					Message:  "Error message",
					Severity: fnsdk.Info,
					File: &fnsdk.File{
						Path:  "other-resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: fnsdk.Info,
					File: &fnsdk.File{
						Path:  "resource.yaml",
						Index: 0,
					},
				},
				{
					Message:  "Error message",
					Severity: fnsdk.Error,
					File: &fnsdk.File{
						Path:  "resource.yaml",
						Index: 1,
					},
				},
				{
					Message:  "Error message",
					Severity: fnsdk.Warning,
					File: &fnsdk.File{
						Path:  "resource.yaml",
						Index: 2,
					},
				},
			},
		},

		{
			name: "sort based on other fields",
			input: fnsdk.Results{
				{
					Message:  "Error message",
					Severity: fnsdk.Error,
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
					Field: &fnsdk.Field{
						Path: "spec",
					},
				},
				{
					Message:  "Error message",
					Severity: fnsdk.Error,
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
					Field: &fnsdk.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Another error message",
					Severity: fnsdk.Error,
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
					Field: &fnsdk.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Another error message",
					Severity: fnsdk.Error,
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
					Field: &fnsdk.Field{
						Path: "metadata.name",
					},
				},
			},
			output: fnsdk.Results{
				{
					Message:  "Another error message",
					Severity: fnsdk.Error,
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
					Field: &fnsdk.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Another error message",
					Severity: fnsdk.Error,
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
					Field: &fnsdk.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Error message",
					Severity: fnsdk.Error,
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
					Field: &fnsdk.Field{
						Path: "metadata.name",
					},
				},
				{
					Message:  "Error message",
					Severity: fnsdk.Error,
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
					Field: &fnsdk.Field{
						Path: "spec",
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		tc.input.Sort()
		if !reflect.DeepEqual(tc.input, tc.output) {
			t.Errorf("in testcase %q, expect: %#v, but got: %#v", tc.name, tc.output, tc.input)
		}
	}
}
