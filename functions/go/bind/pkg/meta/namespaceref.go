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

package meta

import (
	"fmt"

	"github.com/kptdev/krm-functions-sdk/go/fn"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type namespaceRef struct {
	parentObject *fn.SubObject

	name      string
	fieldPath []string
}

func buildMetadataNamespaceReference(parentObject *fn.KubeObject) (Ref, error) {
	namespace := parentObject.GetMap("metadata").GetString("namespace")
	if namespace == "" {
		return nil, fmt.Errorf("expected namespace to be set")
	}
	return &namespaceRef{
		name:         namespace,
		parentObject: &parentObject.SubObject,
		fieldPath:    []string{"metadata", "namespace"},
	}, nil
}

// buildRefNamespaceReference returns a Ref represented the namespace in a reference
func buildRefNamespaceReference(ref *fn.SubObject, fieldPath ...string) (Ref, error) {
	namespace, _, _ := ref.NestedString(fieldPath...)
	if namespace == "" {
		return nil, fmt.Errorf("expected namespace to be set")
	}
	return &namespaceRef{
		name:         namespace,
		parentObject: ref,
		fieldPath:    fieldPath,
	}, nil
}

var _ Ref = &namespaceRef{}

func (r *namespaceRef) GetName() string {
	return r.name
}

func (r *namespaceRef) SetName(name string) {
	r.parentObject.SetNestedString(name, r.fieldPath...)
	// r.parentObject.SetNestedString(name, "metadata", "namespace")
	r.name = name
}

func (r *namespaceRef) GetNamespace() string {
	return ""
}

func (r *namespaceRef) SetNamespace(namespace string) error {
	return fmt.Errorf("cannot set namespace on Namespace ref")
}

func (r *namespaceRef) GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"}
}

func (r *namespaceRef) GroupKind() schema.GroupKind {
	return schema.GroupKind{Group: "", Kind: "Namespace"}
}
