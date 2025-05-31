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
	"strings"
	"testing"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn/internal"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestMapVariantToTypedObject(t *testing.T) {
	testcases := []struct {
		name     string
		src      string
		dst      interface{}
		expected interface{}
	}{
		{
			name: "k8s built-in types",
			src: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-cm
  namespace: my-ns
data:
  foo: bar
`,
			dst: &corev1.ConfigMap{},
			expected: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-cm",
					Namespace: "my-ns",
				},
				Data: map[string]string{
					"foo": "bar",
				},
			},
		},
		{
			name: "crd type with metav1.ObjectMeta",
			src: `apiVersion: example.co/v1
kind: Foo
metadata:
  name: my-foo
desiredReplicas: 1
`,
			dst: &Foo{},
			expected: &Foo{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "example.co/v1",
					Kind:       "Foo",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-foo",
				},
				DesiredReplicas: 1,
			},
		},
		{
			name: "crd type with yaml.ResourceIdentifier",
			src: `apiVersion: example.co/v1
kind: Bar
metadata:
  name: my-bar
desiredReplicas: 1
`,
			dst: &Bar{},
			expected: &Bar{
				ResourceMeta: yaml.ResourceMeta{
					TypeMeta: yaml.TypeMeta{
						APIVersion: "example.co/v1",
						Kind:       "Bar",
					},
					ObjectMeta: yaml.ObjectMeta{
						NameMeta: yaml.NameMeta{
							Name: "my-bar",
						},
					},
				},
				DesiredReplicas: 1,
			},
		},
	}

	for _, tc := range testcases {
		rn := yaml.MustParse(tc.src)
		mv := internal.NewMap(rn.YNode())
		err := internal.MapVariantToTypedObject(mv, tc.dst)
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, tc.dst)
	}
}

func TestNewFromTypedObject(t *testing.T) {
	testcases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name: "k8s built-in types",
			input: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-cm",
					Namespace: "my-ns",
				},
				Data: map[string]string{
					"foo": "bar",
				},
			},
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-cm
  namespace: my-ns
data:
  foo: bar
`,
		},
		{
			name: "crd type with metav1.ObjectMeta",
			input: &Foo{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "example.co/v1",
					Kind:       "Foo",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-foo",
				},
				DesiredReplicas: 1,
			},
			expected: `apiVersion: example.co/v1
kind: Foo
metadata:
  name: my-foo
desiredReplicas: 1
`,
		},
		{
			name: "crd type with yaml.ResourceIdentifier",
			input: &Bar{
				ResourceMeta: yaml.ResourceMeta{
					TypeMeta: yaml.TypeMeta{
						APIVersion: "example.co/v1",
						Kind:       "Bar",
					},
					ObjectMeta: yaml.ObjectMeta{
						NameMeta: yaml.NameMeta{
							Name: "my-bar",
						},
					},
				},
				DesiredReplicas: 1,
			},
			expected: `apiVersion: example.co/v1
kind: Bar
metadata:
  name: my-bar
desiredReplicas: 1
`,
		},
		{
			name: "k8s built in pod type",
			input: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-pod",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "container",
						Image: "test",
					}},
				},
			},
			expected: `apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - name: container
    image: test
    resources: {}
status: {}
`,
		},
	}

	for _, tc := range testcases {
		mv, err := fn.NewFromTypedObject(tc.input)
		assert.NoError(t, err)
		s := mv.String()
		assert.Equal(t, tc.expected, s)
	}
}

func TestBadNewFromTypedObject(t *testing.T) {
	input := []corev1.ConfigMap{
		{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-cm",
				Namespace: "my-ns",
			},
			Data: map[string]string{
				"foo": "bar",
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-cm2",
				Namespace: "my-ns2",
			},
			Data: map[string]string{
				"foo2": "bar2",
			},
		},
	}
	_, err := fn.NewFromTypedObject(input)
	if err == nil {
		t.Errorf("expect error, got nil")
	}
	if !strings.Contains(err.Error(), "got reflect.Slice") {
		t.Errorf("got unexpected error %v", err)
	}
}

type Foo struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	DesiredReplicas   int `json:"desiredReplicas,omitempty" yaml:"desiredReplicas,omitempty"`
}

type Bar struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	DesiredReplicas   int `json:"desiredReplicas,omitempty" yaml:"desiredReplicas,omitempty"`
}
