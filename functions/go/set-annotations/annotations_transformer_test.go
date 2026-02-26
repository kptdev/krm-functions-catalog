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
// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"reflect"
	"testing"
)

func runAnnotationTransformerE(config, input string) (string, error) {
	resmapFactory := newResMapFactory()
	resMap, err := resmapFactory.NewResMapFromBytes([]byte(input))
	if err != nil {
		return "", err
	}

	var plugin = &KustomizePlugin
	plugin.Results = nil
	err = plugin.Config(nil, []byte(config))
	if err != nil {
		return "", err
	}
	tc, err := getDefaultConfig()
	if err != nil {
		return "", err
	}
	plugin.AdditionalAnnotationFields = append(plugin.AdditionalAnnotationFields, tc.FieldSpecs...)
	err = plugin.Transform(resMap)
	if err != nil {
		return "", err
	}
	y, err := resMap.AsYaml()
	if err != nil {
		return "", err
	}
	return string(y), nil
}

func runAnnotationTransformer(t *testing.T, config, input string) string {
	s, err := runAnnotationTransformerE(config, input)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestAnnotationsTransformer(t *testing.T) {
	config := `
annotations:
  app: myApp
`
	input := `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mungebot
  labels:
    app: mungebot
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: mungebot
    spec:
      containers:
      - name: nginx
        image: nginx
`
	expected := `apiVersion: v1
kind: Service
metadata:
  annotations:
    app: myApp
  name: myService
spec:
  ports:
  - port: 7002
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    app: myApp
  labels:
    app: mungebot
  name: mungebot
spec:
  replicas: 1
  template:
    metadata:
      annotations:
        app: myApp
      labels:
        app: mungebot
    spec:
      containers:
      - image: nginx
        name: nginx
`
	output := runAnnotationTransformer(t, config, input)
	if output != expected {
		fmt.Println("Actual:")
		fmt.Println(output)
		fmt.Println("===")
		fmt.Println("Expected:")
		fmt.Println(expected)
		t.Fatalf("Actual doesn't equal to expected")
	}
}

func TestAnnotationsTransformerIdempotence(t *testing.T) {
	config := `
annotations:
  app: myApp
`
	input := `
apiVersion: v1
kind: Service
metadata:
  name: myService
spec:
  ports:
  - port: 7002
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mungebot
  labels:
    app: mungebot
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: mungebot
    spec:
      containers:
      - name: nginx
        image: nginx
`
	expected := `apiVersion: v1
kind: Service
metadata:
  annotations:
    app: myApp
  name: myService
spec:
  ports:
  - port: 7002
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    app: myApp
  labels:
    app: mungebot
  name: mungebot
spec:
  replicas: 1
  template:
    metadata:
      annotations:
        app: myApp
      labels:
        app: mungebot
    spec:
      containers:
      - image: nginx
        name: nginx
`
	// do the transformation twice
	output := runAnnotationTransformer(t, config, input)
	output = runAnnotationTransformer(t, config, output)
	if output != expected {
		fmt.Println("Actual:")
		fmt.Println(output)
		fmt.Println("===")
		fmt.Println("Expected:")
		fmt.Println(expected)
		t.Fatalf("Actual doesn't equal to expected")
	}
}

func TestAnnotationsTransformerResults(t *testing.T) {
	config := `
annotations:
  app: myApp
`
	input := `
apiVersion: v1
kind: Service
metadata:
  annotations:
    internal.config.kubernetes.io/path: foo.yaml
    internal.config.kubernetes.io/index: 0
  name: myService
spec:
  ports:
  - port: 7002
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    internal.config.kubernetes.io/path: bar.yaml
    internal.config.kubernetes.io/index: 1
  name: mungebot
  labels:
    app: mungebot
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: mungebot
    spec:
      containers:
      - name: nginx
        image: nginx
`
	expectedResults := AnnotationResults{
		{
			FilePath:  "foo.yaml",
			FileIndex: "0",
			FieldPath: "metadata.annotations",
		}: {"app": "myApp"},
		{
			FilePath:  "bar.yaml",
			FileIndex: "1",
			FieldPath: "metadata.annotations",
		}: {"app": "myApp"},
		{
			FilePath:  "bar.yaml",
			FileIndex: "1",
			FieldPath: "spec.template.metadata.annotations",
		}: {"app": "myApp"},
	}
	runAnnotationTransformer(t, config, input)
	if !reflect.DeepEqual(KustomizePlugin.Results, expectedResults) {
		fmt.Println("Actual:")
		fmt.Println(KustomizePlugin.Results)
		fmt.Println("===")
		fmt.Println("Expected:")
		fmt.Println(expectedResults)
		t.Fatalf("Actual doesn't equal to expected")
	}
}
