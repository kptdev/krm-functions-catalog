// Copyright 2022 Google LLC
// Modifications Copyright (C) 2025 OpenInfra Foundation Europe.
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
package rename

import (
	"fmt"

	"github.com/kptdev/krm-functions-catalog/functions/go/bind/pkg/meta"
	"github.com/kptdev/krm-functions-sdk/go/fn"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Rename will rename/renamespace an object, and will update all references to it found in the specified objects.
// This is a convenience wrapper around RenameSpec.
func Rename(object *fn.KubeObject, newName string, newNamespace string, objects fn.KubeObjects) error {
	spec := RenameSpec{
		OldName:      object.GetName(),
		Name:         newName,
		OldNamespace: object.GetNamespace(),
		Namespace:    newNamespace,
		APIVersion:   object.GetAPIVersion(),
		Kind:         object.GetKind(),
	}

	if err := spec.Validate(); err != nil {
		return err
	}

	return spec.Transform(objects)
}

// RenameSpec describes an object rename where references to that object are also updated.
type RenameSpec struct {
	OldName      string `json:"oldName"`
	Name         string `json:"name"`
	OldNamespace string `json:"oldNamespace,omitempty"`
	Namespace    string `json:"namespace,omitempty"`
	// TODO: Accept group instead of group+version?
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`

	IgnoreObjectNotFound bool `json:"ignoreObjectNotFound"`
}

// Validate verifies that required fields are set.
func (f *RenameSpec) Validate() error {
	if f.OldName == "" {
		return fmt.Errorf("`oldName` is required")
	}
	if f.Name == "" {
		return fmt.Errorf("`name` is required")
	}
	if f.APIVersion == "" {
		return fmt.Errorf("`apiVersion` is required")
	}
	if f.Kind == "" {
		return fmt.Errorf("`kind` is required")
	}
	return nil
}

// Run is an entrypoint for RenameSpec, supporting invocation as a KRM function.
func Run(rl *fn.ResourceList) (bool, error) {
	f := RenameSpec{}

	err := f.LoadConfig(rl.FunctionConfig)
	if err != nil {
		rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(fmt.Errorf("functionConfig error: %w", err), rl.FunctionConfig))
		return true, nil
	}

	if err := f.Validate(); err != nil {
		return false, nil
	}

	if err := f.Transform(rl.Items); err != nil {
		rl.Results = append(rl.Results, fn.ErrorResult(err))
	}
	return true, nil
}

// LoadConfig parses the configuration from a specified KubeObject.
func (f *RenameSpec) LoadConfig(fnConfig *fn.KubeObject) error {
	if fnConfig != nil {
		// fmt.Printf("fnConfig: %+v\n", fnConfig) // TODO: Gives error
		switch { //TODO: o.GroupVersionKind()
		case fnConfig.IsGVK("", "v1", "ConfigMap"):
			data := fnConfig.UpsertMap("data") // TODO: Why does GetMap fail?
			f.Name = data.GetString("name")
			f.OldName = data.GetString("oldName")
			f.OldNamespace = data.GetString("oldNamespace")
			f.Namespace = data.GetString("namespace")
			f.APIVersion = data.GetString("apiVersion")
			f.Kind = data.GetString("kind")
			// TODO: IgnoreObjectNotFound

		default:
			gvk := schema.GroupVersionKind{}                         // TODO: o.GroupVersionKind()
			return fmt.Errorf("unknown functionConfig Kind %v", gvk) //o.GroupVersionKind())
		}
	}

	return nil
}

// Transform runs the Rename operation.
func (f *RenameSpec) Transform(objects fn.KubeObjects) error {
	// TODO: Are there helpers we can use?
	gv, err := schema.ParseGroupVersion(f.APIVersion)
	if err != nil {
		return fmt.Errorf("error parsing apiVersion %q: %w", f.APIVersion, err)
	}
	gvk := gv.WithKind(f.Kind)
	var matches []*fn.KubeObject
	for _, object := range objects {
		if object.GetName() != f.OldName {
			continue
		}
		if object.GetNamespace() != f.OldNamespace {
			continue
		}
		if !object.IsGVK(gvk.Group, gvk.Version, gvk.Kind) {
			continue
		}
		matches = append(matches, object)
	}
	if len(matches) == 0 {
		if !f.IgnoreObjectNotFound {
			return fmt.Errorf("no object found matching %s/%s:%s/%s", gvk.GroupVersion(), gvk.Kind, f.OldNamespace, f.OldName)
		}
	}
	if len(matches) > 1 {
		return fmt.Errorf("multiple objects found matching %s/%s:%s/%s", gvk.GroupVersion(), gvk.Kind, f.Namespace, f.Name)
	}

	for _, match := range matches {
		match.SetName(f.Name)
		match.SetNamespace(f.Namespace)
	}

	gk := gvk.GroupKind()
	if err := meta.VisitRefs(objects, func(ref meta.Ref) error {
		if ref.GroupKind() != gk {
			return nil
		}
		if ref.GetNamespace() != f.OldNamespace {
			return nil
		}
		if ref.GetName() == f.OldName {
			ref.SetName(f.Name)
			if f.Namespace != f.OldNamespace {
				if err := ref.SetNamespace(f.Namespace); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
