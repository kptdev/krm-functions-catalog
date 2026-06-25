// Copyright 2026 The kpt Authors
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

package applysetters

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// accept invokes the appropriate function on v for each field in object
func accept(v visitor, object *yaml.RNode) error {
	// get the OpenAPI for the type if it exists
	return acceptImpl(v, object, "")
}

// acceptImpl implements accept using recursion
func acceptImpl(v visitor, object *yaml.RNode, p string) error {
	switch object.YNode().Kind {
	case yaml.DocumentNode:
		// Traverse the child of the document
		return accept(v, yaml.NewRNode(object.YNode()))
	case yaml.MappingNode:
		if err := v.visitMapping(object, p); err != nil {
			return err
		}
		return object.VisitFields(func(node *yaml.MapNode) error {
			// Traverse each field value
			return acceptImpl(v, node.Value, p+"."+node.Key.YNode().Value)
		})
	case yaml.SequenceNode:
		return VisitElements(object, func(node *yaml.RNode, i int) error {
			// Traverse each list element
			return acceptImpl(v, node, p+fmt.Sprintf("[%d]", i))
		})
	case yaml.ScalarNode:
		// Visit the scalar field
		return v.visitScalar(object, p)
	}
	return nil
}
