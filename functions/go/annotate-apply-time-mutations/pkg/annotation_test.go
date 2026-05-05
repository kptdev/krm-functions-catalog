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
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestWriteAnnotation(t *testing.T) {
	testCases := map[string]struct {
		config         string
		subs           mutation.ApplyTimeMutation
		expectedResult string
	}{
		"one substitution, no token": {
			config: `apiVersion: bar.foo/v1beta1
kind: MyTestKind
metadata:
  name: my-test-resource
  namespace: test-namespace
spec: {}
`,
			subs: mutation.ApplyTimeMutation{
				mutation.FieldSubstitution{
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
			expectedResult: `apiVersion: bar.foo/v1beta1
kind: MyTestKind
metadata:
  name: my-test-resource
  namespace: test-namespace
  annotations:
    config.kubernetes.io/apply-time-mutation: |
      - sourcePath: $.status.count
        sourceRef:
          apiVersion: foo.bar/v0
          kind: OtherKind
          name: example-name2
          namespace: example-namespace
        targetPath: $.spec.a
spec: {}
`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			node, err := yaml.Parse(tc.config)
			assert.NoError(t, err)

			err = WriteAnnotation(node, tc.subs)
			assert.NoError(t, err)

			result, err := node.String()
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
