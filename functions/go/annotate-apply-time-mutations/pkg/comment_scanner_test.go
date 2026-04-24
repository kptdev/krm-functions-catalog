// Copyright 2021-2025 The kpt Authors
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

package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/cli-utils/pkg/object/mutation"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestApplySettersReferenceParse(t *testing.T) {
	testCases := map[string]struct {
		s          string
		wantStruct mutation.ResourceReference
		wantPath   string
	}{
		"group": {
			s: "${foo.bar.com/namespaces/example-namespace/aKind/example-name:$.my.field}",
			wantStruct: mutation.ResourceReference{
				Group:     "foo.bar.com",
				Kind:      "aKind",
				Name:      "example-name",
				Namespace: "example-namespace",
			},
			wantPath: "$.my.field",
		},
		"apiVersion": {
			s: "${foo.bar.com/v1alpha1/namespaces/example-namespace/aKind/example-name:$.my.field}",
			wantStruct: mutation.ResourceReference{
				APIVersion: "foo.bar.com/v1alpha1",
				Kind:       "aKind",
				Name:       "example-name",
				Namespace:  "example-namespace",
			},
			wantPath: "$.my.field",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			doesHaveRef := hasRef(tc.s)
			if !doesHaveRef {
				t.Fatalf("String %q doesn't have a valid ref", tc.s)
			}

			gotRef, gotPath := commentToReference(tc.s)

			if gotRef != tc.wantStruct {
				t.Errorf("CommentToReference returned struct %v wanted %v", gotRef, tc.wantStruct)
			}

			if gotPath != tc.wantPath {
				t.Errorf("CommentToReference returned path %q wanted %q", gotPath, tc.wantPath)
			}
		})
	}
}

func TestCommentToTokenField(t *testing.T) {
	testCases := map[string]struct {
		s          string
		givenIndex int
		wantValue  string
		wantToken  string
	}{
		"generated token": {
			s:          "prefix-${foo.bar.com/namespaces/example-namespace/aKind/example-name:$.my.field}-suffix",
			givenIndex: 5,
			wantValue:  "prefix-${ref5}-suffix",
			wantToken:  "${ref5}",
		},
		"full field replacement": {
			s:          "${foo.bar.com/v1alpha1/namespaces/example-namespace/aKind/example-name:$.my.field}",
			givenIndex: 2,
			wantValue:  "",
			wantToken:  "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			gotReplace, gotToken := commentToTokenField(tc.s, tc.givenIndex)
			if gotReplace != tc.wantValue {
				t.Errorf("CommentToTokenField returned replacement %q want %q", gotReplace, tc.wantValue)
			}
			if gotToken != tc.wantToken {
				t.Errorf("CommentToTokenField returned token %q want %q", gotToken, tc.wantToken)
			}
		})
	}
}

func TestCommentScan(t *testing.T) {
	testCases := map[string]struct {
		config        string
		expectResults []ScanResult
	}{
		"two fields with comments": {
			config: `apiVersion: bar.foo/v1beta1
kind: MyTestKind
metadata:
  name: my-test-resource
  namespace: test-namespace
  annotations:
    unmodified-key: foobarbaz
spec:
  a: 0 # apply-time-mutation: ${foo.bar/v0/namespaces/example-namespace/OtherKind/example-name2:$.status.count}
  b: "" # apply-time-mutation: prefix${foo.bar/v0/namespaces/example-namespace/OtherKind/example-name2:$.status.count}suffix
`,
			expectResults: []ScanResult{
				{
					Path:    "$.spec.a",
					Value:   0,
					Comment: "# apply-time-mutation: ${foo.bar/v0/namespaces/example-namespace/OtherKind/example-name2:$.status.count}",
					Substitution: mutation.FieldSubstitution{
						SourceRef: mutation.ResourceReference{
							APIVersion: "foo.bar/v0",
							Kind:       "OtherKind",
							Name:       "example-name2",
							Namespace:  "example-namespace",
						},
						SourcePath: "$.status.count",
						TargetPath: "$.spec.a",
					},
				},
				{
					Path:    "$.spec.b",
					Value:   "",
					Comment: "# apply-time-mutation: prefix${foo.bar/v0/namespaces/example-namespace/OtherKind/example-name2:$.status.count}suffix",
					Substitution: mutation.FieldSubstitution{
						SourceRef: mutation.ResourceReference{
							APIVersion: "foo.bar/v0",
							Kind:       "OtherKind",
							Name:       "example-name2",
							Namespace:  "example-namespace",
						},
						SourcePath: "$.status.count",
						TargetPath: "$.spec.b",
						Token:      "${ref1}",
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			node, err := kyaml.Parse(tc.config)
			assert.NoError(t, err)
			meta, err := node.GetMeta()
			assert.NoError(t, err)
			scanner := CommentScanner{
				ObjMeta: meta.GetIdentifier(),
				ObjFile: framework.File{
					Path:  "test.yaml",
					Index: 0,
				},
			}
			results, err := scanner.Scan(node)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectResults, results)
		})
	}
}
