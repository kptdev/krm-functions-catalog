// Copyright 2022 Google LLC
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

package internal_test

import (
	"bytes"
	"testing"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn/internal"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	yaml2 "sigs.k8s.io/kustomize/kyaml/yaml"
)

const deploymentYaml = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
    env: prod
  finalizers:
    - foo
    - bar
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
        - name: nginx
          image: nginx:1.14.2
          ports:
            - containerPort: 80
`

func parseRaw(in []byte) ([]*internal.MapVariant, error) {
	d := yaml2.NewDecoder(bytes.NewBuffer(in))
	node := &yaml2.Node{}
	err := d.Decode(node)
	if err != nil {
		return nil, err
	}
	return internal.ExtractObjects(node)
}

func TestHelpers(t *testing.T) {
	rawmvs, _ := parseRaw([]byte(deploymentYaml))
	assert.Len(t, rawmvs, 1, "expect 1 object after parsing")
	mv := rawmvs[0]

	name, found, err := mv.GetNestedString("metadata", "name")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "nginx-deployment", name)

	ns, found, err := mv.GetNestedString("metadata", "namespace")
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Equal(t, "", ns)
	err = mv.SetNestedString("test-ns", "metadata", "namespace")
	assert.NoError(t, err)
	ns, found, err = mv.GetNestedString("metadata", "namespace")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "test-ns", ns)

	replicas, found, err := mv.GetNestedInt("spec", "replicas")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, 3, replicas)
	err = mv.SetNestedInt(10, "spec", "replicas")
	assert.NoError(t, err)
	replicas, found, err = mv.GetNestedInt("spec", "replicas")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, 10, replicas)

	notExistInt, found, err := mv.GetNestedInt("spec", "foo")
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Equal(t, 0, notExistInt)

	labels, found, err := mv.GetNestedStringMap("metadata", "labels")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Len(t, labels, 2)
	assert.Equal(t, map[string]string{"app": "nginx", "env": "prod"}, labels)

	notExistStringMap, found, err := mv.GetNestedStringMap("metadata", "something")
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Len(t, notExistStringMap, 0)
	var emptyMap map[string]string
	assert.Equal(t, emptyMap, notExistStringMap)

	annotations, found, err := mv.GetNestedStringMap("metadata", "annotations")
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Len(t, annotations, 0)
	assert.Equal(t, emptyMap, annotations)
	err = mv.SetNestedStringMap(map[string]string{"hello": "world"}, "metadata", "annotations")
	assert.NoError(t, err)
	annotation, found, err := mv.GetNestedString("metadata", "annotations", "hello")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "world", annotation)

	metadata, found, err := mv.GetNestedMap("metadata")
	assert.NoError(t, err)
	assert.True(t, found)
	err = mv.SetNestedMap(metadata, "spec", "template", "metadata")
	assert.NoError(t, err)
	label, found, err := mv.GetNestedString("spec", "template", "metadata", "labels", "env")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "prod", label)

	container := map[string]string{
		"name":  "logger",
		"image": "my-logger",
	}
	containers, found, err := mv.GetNestedSlice("spec", "template", "spec", "containers")
	assert.NoError(t, err)
	assert.True(t, found)
	mvs, err := containers.Elements()
	assert.NoError(t, err)
	assert.Len(t, mvs, 1)
	containers.Add(internal.NewStringMapVariant(container))
	containers, found, err = mv.GetNestedSlice("spec", "template", "spec", "containers")
	assert.NoError(t, err)
	assert.True(t, found)
	mvs, err = containers.Elements()
	assert.NoError(t, err)
	assert.Len(t, mvs, 2)

	containers, found, err = mv.GetNestedSlice("spec", "template", "spec", "containers")
	assert.NoError(t, err)
	assert.True(t, found)
	conts, err := containers.Elements()
	assert.NoError(t, err)
	for i, cont := range conts {
		name, found, err = cont.GetNestedString("name")
		if err == nil && found && name == "nginx" {
			err = conts[i].SetNestedString("nginx:1.21.3", "image")
			assert.NoError(t, err)
		}
	}
	containers, found, err = mv.GetNestedSlice("spec", "template", "spec", "containers")
	assert.NoError(t, err)
	assert.True(t, found)
	mvs, err = containers.Elements()
	assert.NoError(t, err)
	assert.Len(t, mvs, 2)
	img, found, err := mvs[0].GetNestedString("image")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "nginx:1.21.3", img)

	trueVar := true
	falseVar := false
	sc := corev1.SecurityContext{
		RunAsNonRoot:             &trueVar,
		AllowPrivilegeEscalation: &falseVar,
	}
	scmv, err := internal.TypedObjectToMapVariant(sc)
	assert.NoError(t, err)
	err = mv.SetNestedMap(scmv, "spec", "template", "spec", "securityContext")
	assert.NoError(t, err)

	runAsNonRoot, found, err := mv.GetNestedBool("spec", "template", "spec", "securityContext", "runAsNonRoot")
	assert.NoError(t, err)
	assert.True(t, found)
	assert.True(t, runAsNonRoot)
}
