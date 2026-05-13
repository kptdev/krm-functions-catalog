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
	"fmt"

	"github.com/kptdev/krm-functions-sdk/go/fn"
	"sigs.k8s.io/yaml"
)

// processV1Hierarchy handles v1alpha1 ResourceHierarchy.
func processV1Hierarchy(obj *fn.KubeObject, rl *fn.ResourceList) fn.Results {
	var results fn.Results
	results = append(results, oldHierarchyWarning(obj))

	orgID, found, err := obj.NestedString("spec", "organization")
	if err != nil || !found || orgID == "" {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s has an undefined organization", obj.GetName()), obj))
		return results
	}

	layers, found, err := obj.NestedStringSlice("spec", "layers")
	if err != nil || !found || len(layers) == 0 {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s has no layers defined", obj.GetName()), obj))
		return results
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
	configMap, found := spec["config"].(map[string]any)
	if !found {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s has no config defined", obj.GetName()), obj))
		return results
	}

	root := &hierarchyNode{
		name: orgID,
		kind: organizationKind,
	}

	namespace := obj.GetNamespace()
	annotations := map[string]string{}
	placement := newSourcePlacement(obj)

	errResult := generateV1HierarchyTree(root, layers, 0, configMap, namespace, annotations, []string{}, placement, rl)
	if errResult != nil {
		results = append(results, errResult)
	}

	return results
}

// generateV1HierarchyTree recursively builds the folder tree for v1.
func generateV1HierarchyTree(
	node *hierarchyNode,
	layers []string,
	layerIndex int,
	config map[string]any,
	namespace string,
	annotations map[string]string,
	path []string,
	placement *sourcePlacement,
	rl *fn.ResourceList,
) *fn.Result {
	if layerIndex >= len(layers) {
		return nil
	}

	layer := layers[layerIndex]
	foldersRaw, found := config[layer]
	if !found {
		return fn.GeneralResult(
			fmt.Sprintf("Layer %q has no corresponding config entry. Either add to spec.config.%s or remove it from spec.layers", layer, layer),
			fn.Error,
		)
	}

	foldersSlice, ok := foldersRaw.([]any)
	if !ok {
		return fn.GeneralResult(
			fmt.Sprintf("Layer %q has an invalid config entry; expected a list of folder names", layer),
			fn.Error,
		)
	}

	for _, folderItem := range foldersSlice {
		folderName, ok := folderItem.(string)
		if !ok {
			return fn.GeneralResult(
				fmt.Sprintf("Layer %q contains a non-string folder entry", layer),
				fn.Error,
			)
		}
		if folderName == "" {
			continue
		}

		child := &hierarchyNode{
			name: folderName,
			kind: folderKind,
		}

		folder := generateManifest(folderName, path, node, annotations, namespace, false, placement)
		_ = rl.UpsertObjectToItems(folder, nil, true)

		currentPath := append(clonePath(path), folderName)
		errResult := generateV1HierarchyTree(child, layers, layerIndex+1, config, namespace, annotations, currentPath, placement, rl)
		if errResult != nil {
			return errResult
		}

		node.children = append(node.children, child)
	}

	return nil
}
