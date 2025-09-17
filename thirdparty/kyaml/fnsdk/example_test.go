package fnsdk_test

import (
	"github.com/kptdev/krm-functions-catalog/thirdparty/kyaml/fnsdk"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var (
	deployment fnsdk.KubeObject
	configMap  fnsdk.KubeObject
)

func ExampleKubeObject_mutatePrimitiveField() {
	replicas, found, err := deployment.GetInt("spec", "replicas")
	if err != nil { /* do something */
	}
	if !found { /* do something */
	}

	// mutate the replicas variable

	err = deployment.Set(&replicas, "spec", "replicas")
	if err != nil { /* do something */
	}
}

func ExampleKubeObject_mutatePrimitiveSlice() {
	var finalizers []string
	found, err := deployment.Get(&finalizers, "metadata", "finalizers")
	if err != nil { /* do something */
	}
	if !found { /* do something */
	}

	// mutate the finalizers slice

	err = deployment.Set(finalizers, "metadata", "finalizers")
	if err != nil { /* do something */
	}
}

func ExampleKubeObject_mutatePrimitiveMap() {
	var data map[string]string
	found, err := configMap.Get(&data, "data")
	if err != nil { /* do something */
	}
	if !found { /* do something */
	}

	// mutate the data map

	err = deployment.Set(data, "data")
	if err != nil { /* do something */
	}
}

func ExampleKubeObject_mutateStrongTypedField() {
	var podTemplate corev1.PodTemplate
	found, err := configMap.Get(&podTemplate, "spec", "template")
	if err != nil { /* do something */
	}
	if !found { /* do something */
	}

	// mutate the podTemplate object

	err = deployment.Set(podTemplate, "spec", "template")
	if err != nil { /* do something */
	}
}

func ExampleKubeObject_mutateStrongTypedSlice() {
	var containers []corev1.Container
	found, err := deployment.Get(&containers, "spec", "template", "spec", "containers")
	if err != nil { /* do something */
	}
	if !found { /* do something */
	}

	// mutate the podTemplate object

	err = deployment.Set(containers, "spec", "template", "spec", "containers")
	if err != nil { /* do something */
	}
}

func ExampleKubeObject_mutateRNode() {
	var rnode yaml.RNode
	// Get a field as RNode. This may be useful if you want to deal with low-level
	// yaml manipulation (e.g. dealing with comments).
	found, err := deployment.Get(&rnode, "metadata", "namespace")
	if err != nil { /* do something */
	}
	if !found { /* do something */
	}

	// Any modification done on the rnode will be reflected on the original object.
	// No need to invoke Set method when using RNode
	ynode := rnode.YNode()
	ynode.HeadComment = ynode.LineComment
	ynode.LineComment = ""
}
