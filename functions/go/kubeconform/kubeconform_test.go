// Copyright (C) 2025 OpenInfra Foundation Europe
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
	"github.com/kptdev/krm-functions-sdk/go/fn"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExtractConfig(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		expected KubeconformConfig
	}{
		{
			name: "all fields as strings",
			yamlData: `apiVersion: v1
kind: ConfigMap
metadata:
  name: fn-config
data:
  schema_location: "/schemas"
  additional_schema_locations: "one,two"
  ignore_missing_schemas: "true"
  skip_kinds: MyCustom,AnotherKind
  strict: "true"
`,
			expected: KubeconformConfig{
				SchemaLocation:            "/schemas",
				AdditionalSchemaLocations: []string{"one", "two"},
				IgnoreMissingSchemas:      true,
				SkipKinds:                 []string{"MyCustom", "AnotherKind"},
				Strict:                    true,
			},
		},
		{
			name: "bools and empty additional schemas",
			yamlData: `
ignore_missing_schemas: true
strict: false
additional_schema_locations: ""
`,
			expected: KubeconformConfig{
				SchemaLocation:            "",
				AdditionalSchemaLocations: nil,
				IgnoreMissingSchemas:      true,
				SkipKinds:                 nil,
				Strict:                    false,
			},
		},
		{
			name: "fallback to nested data",
			yamlData: `
data:
  skip_kinds: NestedKind
  strict: "true"
`,
			expected: KubeconformConfig{
				SchemaLocation:            "",
				AdditionalSchemaLocations: nil,
				IgnoreMissingSchemas:      false,
				SkipKinds:                 []string{"NestedKind"},
				Strict:                    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := fn.ParseKubeObject([]byte(tt.yamlData))
			if err != nil {
				t.Fatalf("failed to parse YAML: %v", err)
			}

			cfg := extractConfig(obj)

			if cfg.SchemaLocation != tt.expected.SchemaLocation {
				t.Errorf("SchemaLocation: got %q, want %q", cfg.SchemaLocation, tt.expected.SchemaLocation)
			}
			if !reflect.DeepEqual(cfg.AdditionalSchemaLocations, tt.expected.AdditionalSchemaLocations) {
				t.Errorf("AdditionalSchemaLocations: got %v, want %v", cfg.AdditionalSchemaLocations, tt.expected.AdditionalSchemaLocations)
			}
			if cfg.IgnoreMissingSchemas != tt.expected.IgnoreMissingSchemas {
				t.Errorf("IgnoreMissingSchemas: got %v, want %v", cfg.IgnoreMissingSchemas, tt.expected.IgnoreMissingSchemas)
			}
			if !reflect.DeepEqual(cfg.SkipKinds, tt.expected.SkipKinds) {
				t.Errorf("SkipKinds: got %v, want %v", cfg.SkipKinds, tt.expected.SkipKinds)
			}
			if cfg.Strict != tt.expected.Strict {
				t.Errorf("Strict: got %v, want %v", cfg.Strict, tt.expected.Strict)
			}
		})
	}
}

func TestBuildKubeconformArgs(t *testing.T) {

	tests := []struct {
		name              string
		schemaLocation    string
		additionalSchemas []string
		ignoreMissing     bool
		skipKinds         []string
		strict            bool
		wantContains      []string
	}{
		{
			name:           "default schema",
			schemaLocation: "",
			wantContains:   []string{SchemaLocationStr, "file://" + DefaultSchemaLocation},
		},
		{
			name:           "with schema location",
			schemaLocation: "/my/schema",
			wantContains:   []string{SchemaLocationStr, "/my/schema"},
		},
		{
			name:              "with additional schema",
			additionalSchemas: []string{"/my/schema1", "/their/schema2"},
			wantContains:      []string{SchemaLocationStr, "/my/schema1", SchemaLocationStr, "/their/schema2"},
		},
		{
			name:          "ignore missing + strict",
			ignoreMissing: true,
			strict:        true,
			wantContains:  []string{"-ignore-missing-schemas", "-strict"},
		},
		{
			name:         "skip kinds",
			skipKinds:    []string{"Pod", "Deployment"},
			wantContains: []string{"-skip", "Pod,Deployment"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildKubeconformArgs(tt.schemaLocation, tt.additionalSchemas, tt.ignoreMissing, tt.skipKinds, tt.strict)
			for _, want := range tt.wantContains {
				assert.Contains(t, got, want)
			}
		})
	}

}

func TestMarshalKubeObject(t *testing.T) {
	obj := fn.NewEmptyKubeObject()
	err := obj.SetAPIVersion("v1")
	assert.NoError(t, err)
	err = obj.SetKind("Pod")
	assert.NoError(t, err)
	err = obj.SetName("mypod")
	assert.NoError(t, err)

	data, err := marshalKubeObject(obj)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"kind":"Pod"`)
	assert.Contains(t, string(data), `"apiVersion":"v1"`)

}

func TestRunKubeconformForObject(t *testing.T) {
	tmpDir := t.TempDir()
	mockPath := filepath.Join(tmpDir, "kubeconform")
	mockOutput := `{
  "resources": [
    {
      "filename": "kubeval-simple/resources.yaml",
      "kind": "ReplicationController",
      "name": "bob",
      "version": "v1",
      "status": "statusInvalid",
      "msg": "problem validating schema. Check JSON formatting: jsonschema validation failed with 'https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/master-standalone-strict/replicationcontroller-v1.json#' - at '/spec/replicas': got string, want null or integer - at '/spec': additional properties 'templates' not allowed",
      "validationErrors": [
        {
          "path": "/spec/replicas",
          "msg": "got string, want null or integer"
        },
        {
          "path": "/spec",
          "msg": "additional properties 'templates' not allowed"
        }
      ]
    }
  ]
}`

	_ = os.WriteFile(mockPath, fmt.Appendf(nil, "#!/bin/sh\necho '%s'\n", mockOutput), 0755)
	oldPath := os.Getenv("PATH")
	// nolint
	os.Setenv("PATH", tmpDir+":"+oldPath)
	// nolint
	defer os.Setenv("PATH", oldPath)

	yamlStr := `
apiVersion: v1
kind: ReplicationController
metadata:
  name: bob
  annotations:
    internal.config.kubernetes.io/path: "resources.yaml"
spec:
  replicas: asdf
  selector:
    app: nginx
  templates:
    metadata:
      name: nginx
      labels:
        app: nginx
    spec:
      containers:
        - name: nginx
          image: nginx
          ports:
            - containerPort: 80
`
	obj, err := fn.ParseKubeObject([]byte(yamlStr))
	if err != nil {
		panic(fmt.Errorf("failed to parse object: %w", err))
	}

	results, _ := runKubeconformForObject(obj, []string{})
	assert.Len(t, results, 2)
	assert.Contains(t, results[0].String(), "spec.replicas: got string, want null or integer")
	assert.Contains(t, results[1].String(), "spec: additional properties templates not allowed")
}

func TestRunWithFakeFunctionConfig(t *testing.T) {
	tmpDir := t.TempDir()
	mockPath := filepath.Join(tmpDir, "kubeconform")
	mockOutput := `{"resources": []}`
	_ = os.WriteFile(mockPath, fmt.Appendf(nil, "#!/bin/sh\necho '%s'\n", mockOutput), 0755)

	oldPath := os.Getenv("PATH")
	// nolint
	os.Setenv("PATH", tmpDir+":"+oldPath)
	// nolint
	defer os.Setenv("PATH", oldPath)

	pod := fn.NewEmptyKubeObject()
	_ = pod.SetAPIVersion("v1")
	_ = pod.SetKind("Pod")
	_ = pod.SetName("mypod")

	fc := fn.NewEmptyKubeObject()
	_ = fc.SetAnnotation(SchemaLocationKey, "/jsonschema")
	_ = fc.SetAnnotation(StrictKey, "true")

	rl := &fn.ResourceList{
		FunctionConfig: fc,
		Items:          []*fn.KubeObject{pod},
	}

	isValid, err := Run(rl)
	assert.NoError(t, err)
	assert.True(t, isValid)
	assert.Empty(t, rl.Results)
}
