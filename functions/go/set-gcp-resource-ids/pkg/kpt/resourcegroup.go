package kpt

import (
	"github.com/kptdev/krm-functions-sdk/go/fn"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func IsResourceGroup(object *fn.KubeObject) bool {
	return object.IsGroupVersionKind(schema.GroupVersionKind{Group: "kpt.dev", Version: "v1alpha1", Kind: "ResourceGroup"})
}
