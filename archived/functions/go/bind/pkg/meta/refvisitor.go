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
	"strings"

	"github.com/kptdev/krm-functions-sdk/go/fn"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Ref is an interface representing a reference to another object.
// A Ref might be a full ObjectReference-style reference,
// or a shorter reference like metadata.namespace or serviceAccountName in a PodTemplate.
// We expose methods like SetNamespace; these will return an error if the reference type
// does not support changing the namespace.
type Ref interface {
	GroupKind() schema.GroupKind
	GroupVersionKind() schema.GroupVersionKind
	GetName() string
	GetNamespace() string
	SetName(name string)
	SetNamespace(namespace string) error
}

// VisitRefs will invoke the callback function for every reference discovered in the specified objects.
func VisitRefs(objects fn.KubeObjects, visitor func(ref Ref) error) error {
	for _, object := range objects {
		gvk := object.GroupVersionKind()

		namespace := object.GetNamespace()
		if namespace != "" {
			ref, err := buildMetadataNamespaceReference(object)
			if err != nil {
				return err
			}
			if err := visitor(ref); err != nil {
				return err
			}
		}

		refs, err := buildOwnerReferences(object)
		if err != nil {
			return err
		}
		for _, ref := range refs {
			if err := visitor(ref); err != nil {
				return err
			}
		}

		for i := range refInfos {
			refInfo := &refInfos[i]
			// TODO: Check GK?
			if refInfo.GVK != gvk {
				continue
			}
			fields := strings.Split(refInfo.FieldPath, ".")
			if err := visitFields(object, &object.SubObject, fields, refInfo, visitor); err != nil {
				return err
			}
		}
	}
	return nil
}

func visitFields(object *fn.KubeObject, subObject *fn.SubObject, fields []string, refInfo *refInfo, visitor func(ref Ref) error) error {
	if len(fields) == 0 {
		if subObject != nil {
			// Ignore external refs, these don't get renamed
			// TODO: Check if supports external ??
			external := subObject.GetString("external")
			if external != "" {
				return nil
			}

			ref, err := refInfo.buildFullyQualifiedRef(object, subObject)
			if err != nil {
				return err
			}
			if err := visitor(ref); err != nil {
				return err
			}

			if refInfo.NamespaceFieldPath != "" {
				namespaceFieldPath := strings.Split(refInfo.NamespaceFieldPath, ".")
				namespace, _, _ := subObject.NestedString(namespaceFieldPath...)
				if namespace != "" {
					ref, err := buildRefNamespaceReference(subObject, namespaceFieldPath...)
					if err != nil {
						return err
					}
					if err := visitor(ref); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	field := fields[0]
	if strings.HasSuffix(field, "[]") {
		field = strings.TrimSuffix(field, "[]")
		subObjects := subObject.GetSlice(field)
		for _, child := range subObjects {
			if err := visitFields(object, child, fields[1:], refInfo, visitor); err != nil {
				return err
			}
		}
		return nil
	}
	child := subObject.GetMap(field)
	return visitFields(object, child, fields[1:], refInfo, visitor)
}
