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
	"maps"
	"strings"

	"github.com/kptdev/krm-functions-sdk/go/fn"
)

// generateConfigs recursively generates Folders for each node.
func generateConfigs(node *hierarchyNode, path []string, annotations map[string]string, namespace string, isV3 bool, placement *sourcePlacement, rl *fn.ResourceList) {
	for _, child := range node.children {
		folder := generateManifest(child.name, path, node, annotations, namespace, isV3, placement)
		_ = rl.UpsertObjectToItems(folder, nil, true)
		newPath := append(clonePath(path), child.name)
		generateConfigs(child, newPath, annotations, namespace, isV3, placement, rl)
	}
}

// generateManifest creates a Folder resource.
func generateManifest(name string, path []string, parent *hierarchyNode, annotations map[string]string, namespace string, nativeRef bool, placement *sourcePlacement) *fn.KubeObject {
	folderObj := fn.NewEmptyKubeObject()
	_ = folderObj.SetAPIVersion(folderAPIVersion)
	_ = folderObj.SetKind(folderKind)

	fullPath := make([]string, len(path))
	copy(fullPath, path)
	fullPath = append(fullPath, name)
	normalizedName := Normalize(fullPath)
	_ = folderObj.SetName(normalizedName)

	if namespace != "" {
		_ = folderObj.SetNamespace(namespace)
	}

	combinedAnnotations := make(map[string]string)
	maps.Copy(combinedAnnotations, annotations)

	isRoot := len(path) == 0
	parentRef := parent.name
	if !isRoot {
		parentRef = strings.Join(path, ".")
	}

	if nativeRef {
		if isRoot {
			if parent.kind == organizationKind {
				_ = folderObj.SetNestedField(parentRef, "spec", "organizationRef", "external")
			} else {
				_ = folderObj.SetNestedField(parentRef, "spec", "folderRef", "external")
			}
		} else {
			normalizedParent := normalize(parentRef)
			_ = folderObj.SetNestedField(normalizedParent, "spec", "folderRef", "name")
			if namespace != "" {
				combinedAnnotations[dependsOnAnnotation] = fmt.Sprintf(
					"%s/namespaces/%s/%s/%s",
					folderGroup, namespace, folderKind, normalizedParent,
				)
			}
		}
	} else {
		if isRoot && parent.kind == organizationKind {
			combinedAnnotations[organizationIDAnnotation] = parentRef
		} else {
			combinedAnnotations[folderRefAnnotation] = normalize(parentRef)
		}
	}

	if len(combinedAnnotations) > 0 {
		for k, v := range combinedAnnotations {
			_ = folderObj.SetAnnotation(k, v)
		}
	}

	_ = folderObj.SetNestedField(name, "spec", "displayName")
	applySourcePlacement(folderObj, placement)
	return folderObj
}
