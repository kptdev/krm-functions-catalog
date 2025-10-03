// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

type ownerRef struct {
	ref *fn.SubObject

	gvk       schema.GroupVersionKind
	name      string
	namespace string
}

// buildOwnerReferences returns a Ref for any ownerReferences.
func buildOwnerReferences(parentObject *fn.KubeObject) ([]Ref, error) {
	var refs []Ref
	ownerReferences := parentObject.GetMap("metadata").GetSlice("ownerReferences")
	for _, ownerReference := range ownerReferences {
		namespace := ownerReference.GetString("namespace")
		if namespace == "" {
			namespace = parentObject.GetNamespace()
		}
		name := ownerReference.GetString("name")
		if name == "" {
			return nil, fmt.Errorf("expected name to be set")
		}
		apiVersion := ownerReference.GetString("apiVersion")
		if apiVersion == "" {
			return nil, fmt.Errorf("expected apiVersion to be set")
		}
		kind := ownerReference.GetString("kind")
		if kind == "" {
			return nil, fmt.Errorf("expected kind to be set")
		}

		gk, err := schema.ParseGroupVersion(apiVersion)
		if err != nil {
			return nil, fmt.Errorf("error parsing apiVersion %q: %w", apiVersion, err)
		}
		gvk := gk.WithKind(kind)
		refs = append(refs, &ownerRef{
			name:      name,
			namespace: namespace,
			ref:       ownerReference,
			gvk:       gvk,
		})

		if namespace != "" {
			// TODO: Should we allow this?  It's not part of the "real" ownerReference
			r, err := buildRefNamespaceReference(ownerReference, "namespace")
			if err != nil {
				return nil, err
			}
			refs = append(refs, r)
		}
	}
	return refs, nil
}

var _ Ref = &ownerRef{}

func (r *ownerRef) GetName() string {
	return r.name
}

func (r *ownerRef) SetName(name string) {
	r.ref.SetNestedString(name, "name")
	r.name = name
}

func (r *ownerRef) GetNamespace() string {
	return r.namespace
}

func (r *ownerRef) SetNamespace(namespace string) error {
	if namespace == r.namespace {
		return nil
	}

	return fmt.Errorf("cannot change namespace on ownerReference")
}

func (r *ownerRef) GroupVersionKind() schema.GroupVersionKind {
	return r.gvk
}

func (r *ownerRef) GroupKind() schema.GroupKind {
	return r.gvk.GroupKind()
}
