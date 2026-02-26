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
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func runEnsureNameSubstringTransformerE(config, input string) (string, error) {
	resmapFactory := newResMapFactory()
	resMap, err := resmapFactory.NewResMapFromBytes([]byte(input))
	if err != nil {
		return "", err
	}
	configRN, err := yaml.Parse(config)
	if err != nil {
		return "", err
	}
	ens := &EnsureNameSubstring{}
	if err = framework.LoadFunctionConfig(configRN, ens); err != nil {
		return "", err
	}
	if defaultConfig, err := getDefaultConfig(); err != nil {
		return "", err
	} else {
		ens.AdditionalNameFields = append(ens.AdditionalNameFields, defaultConfig.FieldSpecs...)
	}
	if err = ens.Transform(resMap); err != nil {
		return "", err
	}
	resMap.RemoveBuildAnnotations()
	y, err := resMap.AsYaml()
	if err != nil {
		return "", err
	}
	return string(y), nil
}

func runEnsureNameSubstringTransformer(t *testing.T, config, input string) string {
	s, err := runEnsureNameSubstringTransformerE(config, input)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestEnsureNameSubstringDependsOn(t *testing.T) {
	testCases := []struct {
		TestName string
		Config   string
		Input    string
		Expected string
	}{
		{
			TestName: "support prepend for the depends-on annotation",
			Config: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: fn-config
data:
  prepend: dev-
`,
			Input: `apiVersion: v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/depends-on: /namespaces/default/StatefulSet/wordpress-mysql
  labels:
    app: wordpress
  name: wordpress
  namespace: default
---
apiVersion: v1
kind: StatefulSet
metadata:
  labels:
    app: wordpress
  name: wordpress-mysql
  namespace: default
`,
			Expected: `apiVersion: v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/depends-on: /namespaces/default/StatefulSet/dev-wordpress-mysql
  labels:
    app: wordpress
  name: dev-wordpress
  namespace: default
---
apiVersion: v1
kind: StatefulSet
metadata:
  labels:
    app: wordpress
  name: dev-wordpress-mysql
  namespace: default
`,
		},
		{
			TestName: "support append for the depends-on annotation",
			Config: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: fn-config
data:
  append: -dev
`,
			Input: `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  annotations:
    config.kubernetes.io/depends-on: rbac.authorization.k8s.io/ClusterRole/secret-reader
  name: read-secrets-global
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: secret-reader
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: Group
  name: admin
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: secret-reader
`,
			Expected: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  annotations:
    config.kubernetes.io/depends-on: rbac.authorization.k8s.io/ClusterRole/secret-reader-dev
  name: read-secrets-global-dev
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: secret-reader
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: Group
  name: admin
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: secret-reader-dev
`,
		},
		{
			TestName: "not update the depends-on annotation if the referenced resource is not included",
			Config: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: fn-config
data:
  append: -dev
`,
			Input: `apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/depends-on: apps/namespaces/default/StatefulSet/wordpress-mysql
  labels:
    app: wordpress
  name: wordpress
  namespace: default
`,
			Expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    config.kubernetes.io/depends-on: apps/namespaces/default/StatefulSet/wordpress-mysql
  labels:
    app: wordpress
  name: wordpress-dev
  namespace: default
`,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("ensure-name-substring should %s", tc.TestName), func(t *testing.T) {
			actual := runEnsureNameSubstringTransformer(t, tc.Config, tc.Input)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}

func TestSetDependsOnNameSubstring(t *testing.T) {
	testCases := []struct {
		TestName            string
		EditMode            EditMode
		InputResourceLookup resourceLookup
		Input               string
		Expected            string
		IsSet               bool
	}{
		{
			TestName: "prepend the name substring for a namespaced resource",
			EditMode: Prepend,
			Input:    "group/namespaces/ns/kind/name",
			InputResourceLookup: resourceLookup{
				resourceMap: map[resourceKey]bool{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					}: true,
				},
			},
			Expected: "group/namespaces/ns/kind/substrname",
			IsSet:    true,
		},
		{
			TestName: "append the name substring for a namespaced resource",
			EditMode: Append,
			Input:    "group/namespaces/ns/kind/name",
			InputResourceLookup: resourceLookup{
				resourceMap: map[resourceKey]bool{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					}: true,
				},
			},
			Expected: "group/namespaces/ns/kind/namesubstr",
			IsSet:    true,
		},
		{
			TestName: "append the name substring for a core namespaced resource",
			EditMode: Append,
			Input:    "/namespaces/ns/kind/name",
			InputResourceLookup: resourceLookup{
				resourceMap: map[resourceKey]bool{
					{
						Group: "",
						Kind:  "kind",
						Name:  "name",
					}: true,
				},
			},
			Expected: "/namespaces/ns/kind/namesubstr",
			IsSet:    true,
		},
		{
			TestName: "prepend the name substring for a cluster scoped resource",
			EditMode: Prepend,
			Input:    "group/kind/name",
			InputResourceLookup: resourceLookup{
				resourceMap: map[resourceKey]bool{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					}: true,
				},
			},
			Expected: "group/kind/substrname",
			IsSet:    true,
		},
		{
			TestName: "append the name substring for a cluster scoped resource",
			EditMode: Append,
			Input:    "group/kind/name",
			InputResourceLookup: resourceLookup{
				resourceMap: map[resourceKey]bool{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					}: true,
				},
			},
			Expected: "group/kind/namesubstr",
			IsSet:    true,
		},
		{
			TestName: "append the name substring for a core cluster scoped resource",
			EditMode: Append,
			Input:    "/kind/name",
			InputResourceLookup: resourceLookup{
				resourceMap: map[resourceKey]bool{
					{
						Group: "",
						Kind:  "kind",
						Name:  "name",
					}: true,
				},
			},
			Expected: "/kind/namesubstr",
			IsSet:    true,
		},
		{
			TestName: "not change name for namespaced resource with unexpected prefix",
			EditMode: Append,
			Input:    "some-prefix/group/namespaces/ns/kind/name",
			Expected: "some-prefix/group/namespaces/ns/kind/name",
			IsSet:    false,
		},
		{
			TestName: "not change name for namespaced resource with unexpected suffix",
			EditMode: Append,
			Input:    "group/namespaces/ns/kind/name/some-suffix",
			Expected: "group/namespaces/ns/kind/name/some-suffix",
			IsSet:    false,
		},
		{
			TestName: "not change name for malformed namespaced resource",
			EditMode: Append,
			Input:    "group/oops/ns/kind/name/some-suffix",
			Expected: "group/oops/ns/kind/name/some-suffix",
			IsSet:    false,
		},
		{
			TestName: "not change name for cluster scoped resource with unexpected prefix",
			EditMode: Append,
			Input:    "some-prefix/group/kind/name",
			Expected: "some-prefix/group/kind/name",
			IsSet:    false,
		},
		{
			TestName: "not change name for cluster scoped resource with unexpected suffix",
			EditMode: Append,
			Input:    "group/kind/name/some-suffix",
			Expected: "group/kind/name/some-suffix",
			IsSet:    false,
		},
		{
			TestName: "not change name if EditMode is unset",
			Input:    "group/kind/name",
			Expected: "group/kind/name",
			IsSet:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("setDependsOnNameSubstring should %s", tc.TestName), func(t *testing.T) {
			ens := EnsureNameSubstring{
				Substring:           "substr",
				EditMode:            tc.EditMode,
				inputResourceLookup: tc.InputResourceLookup,
			}
			actual, set := ens.setDependsOnNameSubstring(tc.Input)
			assert.Equal(t, tc.IsSet, set)
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
