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
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

type ObjectScanner struct{}

// Scan searches for mutation markup comments and parses them as substitutions.
func (os *ObjectScanner) Scan(obj *yaml.RNode) (*ApplyTimeMutation, error) {
	if obj.GetKind() != "ApplyTimeMutation" {
		// no match
		return nil, nil
	}
	if obj.GetApiVersion() != "fn.kpt.dev/v1alpha1" {
		// no match
		return nil, nil
	}

	config, err := obj.String()
	if err != nil {
		return nil, fmt.Errorf("failed to format object as yaml: %w", err)
	}

	var atm ApplyTimeMutation
	err = k8syaml.Unmarshal([]byte(config), &atm)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ApplyTimeMutation object: %w", err)
	}
	// TODO: validate field values (ex: non-empty)
	return &atm, nil
}
