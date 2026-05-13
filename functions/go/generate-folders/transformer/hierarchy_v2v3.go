// Copyright 2026 The kpt Authors
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

package generate_folders

import (
	"errors"
	"fmt"

	"github.com/kptdev/krm-functions-sdk/go/fn"
	"sigs.k8s.io/yaml"
)

// processV2V3Hierarchy handles v2 and v3 ResourceHierarchy.
func processV2V3Hierarchy(obj *fn.KubeObject, rl *fn.ResourceList, isV3 bool) fn.Results {
	var results fn.Results

	parentExternal, found, err := obj.NestedString("spec", "parentRef", "external")
	if err != nil || !found || parentExternal == "" {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s has an undefined parentRef", obj.GetName()), obj))
		return results
	}

	parentKind, _, _ := obj.NestedString("spec", "parentRef", "kind")
	if parentKind != "" && parentKind != organizationKind && parentKind != folderKind {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s has an unsupported parentRef kind", obj.GetName()), obj))
		return results
	}
	if parentKind == "" {
		parentKind = organizationKind
	}

	namespace := obj.GetNamespace()
	annotations := filterNonInheritableAnnotations(obj.GetAnnotations())
	placement := newSourcePlacement(obj)

	root := &hierarchyNode{
		name: parentExternal,
		kind: parentKind,
	}

	var rawObj map[string]any
	if err := yaml.Unmarshal([]byte(obj.String()), &rawObj); err != nil {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s: failed to parse spec: %v", obj.GetName(), err), obj))
		return results
	}

	spec, ok := rawObj["spec"].(map[string]any)
	if !ok {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s has an invalid spec", obj.GetName()), obj))
		return results
	}

	subtrees := map[string]*hierarchyNode{}
	if subtreeRaw, ok := spec["subtrees"]; ok {
		if subtreeMap, ok := subtreeRaw.(map[string]any); ok {
			for name := range subtreeMap {
				subtrees[name] = &hierarchyNode{name: name, kind: subtreeKind}
			}
			for name, val := range subtreeMap {
				subtreeNode := subtrees[name]
				if children, ok := val.([]any); ok {
					if err := generateTree(subtreeNode, children, subtrees); err != nil {
						results = append(results, fn.ErrorConfigObjectResult(
							fmt.Errorf("ResourceHierarchy %s: error processing subtree %q: %v", obj.GetName(), name, err), obj))
						return results
					}
				}
				subtrees[name] = subtreeNode
			}
		}
	}

	configRaw, ok := spec["config"]
	if !ok {
		results = append(results, fn.GeneralResult(fmt.Sprintf(
			"ResourceHierarchy %s has no spec.config defined; no folders will be generated",
			obj.GetName()), fn.Info))
		return results
	}
	configSlice, ok := configRaw.([]any)
	if !ok || len(configSlice) == 0 {
		results = append(results, fn.GeneralResult(fmt.Sprintf(
			"ResourceHierarchy %s has an empty spec.config; no folders will be generated",
			obj.GetName()), fn.Info))
		return results
	}

	buildErr := buildTreeFromRawConfig(root, configSlice, subtrees)
	if buildErr != nil {
		var missingErr *missingSubtreeError
		if errors.As(buildErr, &missingErr) {
			results = append(results, fn.ErrorConfigObjectResult(
				fmt.Errorf("ResourceHierarchy %s references non-existent subtree %q", obj.GetName(), missingErr.name), obj))
			return results
		}

		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s: invalid spec.config: %v", obj.GetName(), buildErr), obj))
		return results
	}

	generateConfigs(root, []string{}, annotations, namespace, isV3, placement, rl)
	return results
}

// buildTreeFromRawConfig builds the hierarchy from raw config.
func buildTreeFromRawConfig(parent *hierarchyNode, configSlice []any, subtrees map[string]*hierarchyNode) error {
	for _, item := range configSlice {
		if item == nil {
			continue
		}

		switch v := item.(type) {
		case string:
			parent.children = append(parent.children, &hierarchyNode{name: v})
		case map[string]any:
			for name, val := range v {
				node := &hierarchyNode{name: name}

				switch innerVal := val.(type) {
				case []any:
					if err := generateTree(node, innerVal, subtrees); err != nil {
						return err
					}
				case map[string]any:
					if subtreeName, ok := innerVal["$subtree"]; ok {
						stName, ok := subtreeName.(string)
						if !ok {
							return fmt.Errorf("$subtree value is not a string")
						}
						subtreeNode, exists := subtrees[stName]
						if !exists {
							return &missingSubtreeError{name: stName}
						}
						node.children = subtreeNode.children
					}
				case nil:
				}

				parent.children = append(parent.children, node)
			}
		}
	}
	return nil
}

// generateTree builds a hierarchy from children.
func generateTree(root *hierarchyNode, children []any, subtrees map[string]*hierarchyNode) error {
	for _, child := range children {
		if child == nil {
			continue
		}

		switch v := child.(type) {
		case string:
			root.children = append(root.children, &hierarchyNode{name: v})
		case map[string]any:
			if subtreeName, ok := v["$subtree"]; ok && len(v) == 1 {
				stName, ok := subtreeName.(string)
				if !ok {
					return fmt.Errorf("$subtree value is not a string")
				}
				subtreeNode, exists := subtrees[stName]
				if !exists {
					return &missingSubtreeError{name: stName}
				}
				root.children = append(root.children, subtreeNode.children...)
				continue
			}

			for name, val := range v {
				node := &hierarchyNode{name: name}
				switch innerVal := val.(type) {
				case []any:
					if err := generateTree(node, innerVal, subtrees); err != nil {
						return err
					}
				case map[string]any:
					if subtreeName, ok := innerVal["$subtree"]; ok {
						stName, ok := subtreeName.(string)
						if !ok {
							return fmt.Errorf("$subtree value is not a string")
						}
						subtreeNode, exists := subtrees[stName]
						if !exists {
							return &missingSubtreeError{name: stName}
						}
						node.children = subtreeNode.children
					}
				}
				root.children = append(root.children, node)
			}
		}
	}
	return nil
}
