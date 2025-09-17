package fnsdk

import (
	"testing"

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
		mv := &mapVariant{node: rn.YNode()}
		err := mapVariantToTypedObject(mv, tc.dst)
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, tc.dst)
	}
}

func TestTypedObjectToMapVariant(t *testing.T) {
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
	}

	for _, tc := range testcases {
		mv, err := typedObjectToMapVariant(tc.input)
		assert.NoError(t, err)
		s, err := yaml.NewRNode(mv.node).String()
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, s)
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
