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
	"strings"

	"github.com/kptdev/krm-functions-sdk/go/fn"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type fullyQualifiedRef struct {
	parentObject *fn.KubeObject
	ref          *fn.SubObject

	namespaceFieldPath []string

	name      string
	namespace string
	gvk       schema.GroupVersionKind
}

// buildFullyQualifiedRef returns a Ref for the specified "full" reference.
// A full reference is an object, usually with an apiVersion, kind, name, namespace etc.
func (i *refInfo) buildFullyQualifiedRef(parentObject *fn.KubeObject, ref *fn.SubObject) (Ref, error) {
	name := ref.GetString("name")
	namespace := ref.GetString("namespace")
	apiVersion := ref.GetString("apiVersion")
	apiGroup := ref.GetString("apiGroup")
	kind := ref.GetString("kind")

	targetGVKs := i.TargetGVKs
	if kind == "" || apiVersion == "" {
		if len(targetGVKs) == 0 {
			if kind == "" {
				return nil, fmt.Errorf("expected kind to be specified in ref %v:%v", i.GVK, i.FieldPath)
			} else if apiVersion == "" {
				return nil, fmt.Errorf("expected apiVersion to be specified in ref %v:%v", i.GVK, i.FieldPath)
			}
		}

		if kind != "" {
			var matches []schema.GroupVersionKind
			for _, t := range targetGVKs {
				if t.Kind != kind {
					continue
				}
				if apiGroup != "" && t.Group != apiGroup {
					continue
				}
				matches = append(matches, t)
			}

			if len(matches) == 0 {
				return nil, fmt.Errorf("unexpected target kind=%q, group=%q", kind, apiGroup)
			}
			targetGVKs = matches
		}

		if len(targetGVKs) != 1 {
			return nil, fmt.Errorf("found multiple target kinds")
		}

		targetGVK := targetGVKs[0]
		if kind == "" {
			kind = targetGVK.Kind
		}
		if apiVersion == "" {
			apiVersion, _ = targetGVK.ToAPIVersionAndKind()
		}
	}

	targetGV, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, fmt.Errorf("error parsing apiVersion %q: %w", apiVersion, err)
	}
	targetGVK := targetGV.WithKind(kind)

	if len(i.TargetGVKs) == 1 {
		if i.TargetGVKs[0] != targetGVK {
			return nil, fmt.Errorf("unexpected GVK for path %q: got %v want %v", i.FieldPath, targetGVK, i.TargetGVKs[0])
		}
	}

	targetClusterScoped, err := IsClusterScoped(targetGVK)
	if err != nil {
		return nil, err
	}
	if namespace == "" {
		if !targetClusterScoped {
			namespace = parentObject.GetNamespace()
		}
	}

	if targetClusterScoped {
		if namespace != "" {
			return nil, fmt.Errorf("namespace not expected on cluster-scoped kind %v", targetGVK)
		}
	}

	if name == "" {
		return nil, fmt.Errorf("name not set on reference")
	}

	var namespaceFieldPath []string
	if i.NamespaceFieldPath != "" {
		namespaceFieldPath = strings.Split(i.NamespaceFieldPath, ".")
	}

	return &fullyQualifiedRef{
		name:               name,
		namespace:          namespace,
		parentObject:       parentObject,
		ref:                ref,
		gvk:                targetGVK,
		namespaceFieldPath: namespaceFieldPath,
	}, nil
}

var _ Ref = &fullyQualifiedRef{}

func (r *fullyQualifiedRef) GetName() string {
	return r.name
}

func (r *fullyQualifiedRef) SetName(name string) {
	r.ref.SetNestedString(name, "name")
	r.name = name
}

func (r *fullyQualifiedRef) SetNamespace(namespace string) error {
	if len(r.namespaceFieldPath) == 0 {
		return fmt.Errorf("ref does not support changing namespace: %v", r.gvk)
	}
	if err := r.ref.SetNestedString(namespace, r.namespaceFieldPath...); err != nil {
		return err
	}
	r.namespace = namespace
	return nil
}

func (r *fullyQualifiedRef) GetNamespace() string {
	return r.namespace
}

func (r *fullyQualifiedRef) GroupVersionKind() schema.GroupVersionKind {
	return r.gvk
}

func (r *fullyQualifiedRef) GroupKind() schema.GroupKind {
	return r.gvk.GroupKind()
}
