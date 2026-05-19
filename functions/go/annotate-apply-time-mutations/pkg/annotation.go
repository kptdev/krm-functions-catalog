// Copyright 2021-2025 The kpt Authors
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

package pkg

import (
	"errors"
	"fmt"

	"sigs.k8s.io/cli-utils/pkg/object/mutation"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

// WriteAnnotation updates the supplied RNode object to add the
// apply-time-mutation annotation with a multi-line yaml value.
func WriteAnnotation(obj *yaml.RNode, atm mutation.ApplyTimeMutation) error {
	if obj == nil {
		return errors.New("object is nil")
	}
	if atm.Equal(mutation.ApplyTimeMutation{}) {
		return errors.New("mutation is empty")
	}
	// Use sigs.k8s.io/yaml because the ApplyTimeMutation struct uses json field tags,
	// and sigs.k8s.io/kustomize/kyaml/yaml requires yaml field tags...
	yamlBytes, err := k8syaml.Marshal(atm)
	if err != nil {
		return fmt.Errorf("failed to format apply-time-mutation annotation: %v", err)
	}
	a := obj.GetAnnotations()
	if a == nil {
		a = map[string]string{}
	}
	a[mutation.Annotation] = string(yamlBytes)
	err = obj.SetAnnotations(a)
	if err != nil {
		return fmt.Errorf("failed to update annotations: %v", err)
	}
	return nil
}
