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

// Package generate_folders transforms ResourceHierarchy resources into Folders.
package generate_folders

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/kptdev/krm-functions-sdk/go/fn"
	"sigs.k8s.io/yaml"
)

const (
	// ResourceHierarchy API versions
	v3APIVersion = "blueprints.cloud.google.com/v1alpha3"
	v2APIVersion = "dev.cft.v1alpha2"
	v1APIVersion = "dev.cft.v1alpha1"

	// ResourceHierarchy kind
	resourceHierarchyKind = "ResourceHierarchy"

	// Folder constants
	folderAPIVersion = "resourcemanager.cnrm.cloud.google.com/v1beta1"
	folderKind       = "Folder"
	folderGroup      = "resourcemanager.cnrm.cloud.google.com"

	// Annotation constants
	dependsOnAnnotation = "config.kubernetes.io/depends-on"
)

// Annotations that should not be inherited by generated Folders.
var nonInheritableAnnotations = map[string]bool{
	"config.kubernetes.io/local-config":   true,
	"config.k8s.io/function":              true,
	"internal.config.kubernetes.io/id":    true,
	"config.kubernetes.io/id":             true,
	"internal.config.kubernetes.io/path":  true,
	"config.kubernetes.io/path":           true,
	"internal.config.kubernetes.io/index": true,
	"config.kubernetes.io/index":          true,
}

var normalizeRegex = regexp.MustCompile(`[^a-z0-9.\- ]`)
var quoteRegex = regexp.MustCompile(`['"]`)
var underscoreSpaceRegex = regexp.MustCompile(`[_ ]`)

// hierarchyNode represents a node in the folder hierarchy tree.
type hierarchyNode struct {
	name     string
	kind     string
	children []*hierarchyNode
}

type missingSubtreeError struct {
	name string
}

func (e *missingSubtreeError) Error() string {
	return e.name
}

// Run processes ResourceHierarchy resources and generates Folders.
func Run(rl *fn.ResourceList) (bool, error) {
	for _, obj := range rl.Items {
		apiVersion := obj.GetAPIVersion()
		kind := obj.GetKind()

		if kind != resourceHierarchyKind {
			continue
		}

		switch apiVersion {
		case v1APIVersion:
			results := processV1Hierarchy(obj, rl)
			rl.Results = append(rl.Results, results...)
		case v2APIVersion:
			rl.Results = append(rl.Results, oldHierarchyWarning(obj))
			results := processV2V3Hierarchy(obj, rl, false)
			rl.Results = append(rl.Results, results...)
		case v3APIVersion:
			results := processV2V3Hierarchy(obj, rl, true)
			rl.Results = append(rl.Results, results...)
		default:
			rl.Results = append(rl.Results, fn.GeneralResult(fmt.Sprintf(
				"ResourceHierarchy %s has unrecognized apiVersion %q; supported versions: %s, %s, %s",
				obj.GetName(), apiVersion, v3APIVersion, v2APIVersion, v1APIVersion), fn.Warning))
		}
	}

	return true, nil
}

// processV1Hierarchy handles v1alpha1 ResourceHierarchy.
func processV1Hierarchy(obj *fn.KubeObject, rl *fn.ResourceList) fn.Results {
	var results fn.Results
	results = append(results, oldHierarchyWarning(obj))

	// Get spec.organization
	orgID, found, err := obj.NestedString("spec", "organization")
	if err != nil || !found || orgID == "" {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s has an undefined organization", obj.GetName()), obj))
		return results
	}

	// Get spec.layers as string slice
	layers, found, err := obj.NestedStringSlice("spec", "layers")
	if err != nil || !found || len(layers) == 0 {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s has no layers defined", obj.GetName()), obj))
		return results
	}

	var rawObj map[string]interface{}
	if err := yaml.Unmarshal([]byte(obj.String()), &rawObj); err != nil {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s: failed to parse spec: %v", obj.GetName(), err), obj))
		return results
	}
	spec, ok := rawObj["spec"].(map[string]interface{})
	if !ok {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s has an invalid spec", obj.GetName()), obj))
		return results
	}
	configMap, found := spec["config"].(map[string]interface{})
	if !found {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s has no config defined", obj.GetName()), obj))
		return results
	}

	root := &hierarchyNode{
		name: orgID,
		kind: "Organization",
	}

	namespace := obj.GetNamespace()
	annotations := map[string]string{}

	// Start with empty path for v1 hierarchy
	errResult := generateV1HierarchyTree(root, layers, 0, configMap, namespace, annotations, []string{}, rl)
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
	config map[string]interface{},
	namespace string,
	annotations map[string]string,
	path []string,
	rl *fn.ResourceList,
) *fn.Result {
	if layerIndex >= len(layers) {
		return nil
	}

	layer := layers[layerIndex]

	// Check if this layer key is present in config
	foldersRaw, found := config[layer]
	if !found {
		return fn.GeneralResult(
			fmt.Sprintf("Layer %q has no corresponding config entry. Either add to spec.config.%s or remove it from spec.layers", layer, layer),
			fn.Error,
		)
	}

	foldersSlice, ok := foldersRaw.([]interface{})
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
			kind: "Folder",
		}

		folder := generateManifest(folderName, path, node, annotations, namespace, false)
		rl.Items = append(rl.Items, folder)

		currentPath := append(clonePath(path), folderName)
		errResult := generateV1HierarchyTree(child, layers, layerIndex+1, config, namespace, annotations, currentPath, rl)
		if errResult != nil {
			return errResult
		}

		node.children = append(node.children, child)
	}

	return nil
}

// processV2V3Hierarchy handles v2 and v3 ResourceHierarchy.
func processV2V3Hierarchy(obj *fn.KubeObject, rl *fn.ResourceList, isV3 bool) fn.Results {
	var results fn.Results

	// Validate parentRef
	parentExternal, found, err := obj.NestedString("spec", "parentRef", "external")
	if err != nil || !found || parentExternal == "" {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s has an undefined parentRef", obj.GetName()), obj))
		return results
	}

	parentKind, _, _ := obj.NestedString("spec", "parentRef", "kind")
	if parentKind != "" && parentKind != "Organization" && parentKind != "Folder" {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s has an unsupported parentRef kind", obj.GetName()), obj))
		return results
	}

	if parentKind == "" {
		parentKind = "Organization"
	}

	namespace := obj.GetNamespace()
	annotations := filterNonInheritableAnnotations(obj.GetAnnotations())

	root := &hierarchyNode{
		name: parentExternal,
		kind: parentKind,
	}

	// Use raw YAML unmarshalling to parse spec.config and spec.subtrees
	// because the fn SDK's NestedSlice cannot handle mixed-type YAML lists
	// (scalars + maps) which are common in ResourceHierarchy configs.
	var rawObj map[string]interface{}
	if err := yaml.Unmarshal([]byte(obj.String()), &rawObj); err != nil {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s: failed to parse spec: %v", obj.GetName(), err), obj))
		return results
	}

	spec, ok := rawObj["spec"].(map[string]interface{})
	if !ok {
		results = append(results, fn.ErrorConfigObjectResult(
			fmt.Errorf("ResourceHierarchy %s has an invalid spec", obj.GetName()), obj))
		return results
	}

	// Process subtrees
	subtrees := map[string]*hierarchyNode{}
	if subtreeRaw, ok := spec["subtrees"]; ok {
		if subtreeMap, ok := subtreeRaw.(map[string]interface{}); ok {
			for name := range subtreeMap {
				subtrees[name] = &hierarchyNode{
					name: name,
					kind: "Subtree",
				}
			}

			for name, val := range subtreeMap {
				subtreeNode := subtrees[name]
				if children, ok := val.([]interface{}); ok {
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

	// Process config
	configRaw, ok := spec["config"]
	if !ok {
		results = append(results, fn.GeneralResult(fmt.Sprintf(
			"ResourceHierarchy %s has no spec.config defined; no folders will be generated",
			obj.GetName()), fn.Info))
		return results
	}
	configSlice, ok := configRaw.([]interface{})
	if !ok || len(configSlice) == 0 {
		results = append(results, fn.GeneralResult(fmt.Sprintf(
			"ResourceHierarchy %s has an empty spec.config; no folders will be generated",
			obj.GetName()), fn.Info))
		return results
	}

	// Build tree from config
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

	// Generate Folder manifests from the tree
	generateConfigs(root, []string{}, annotations, namespace, isV3, rl)

	return results
}

// buildTreeFromRawConfig builds the hierarchy from raw config.
func buildTreeFromRawConfig(parent *hierarchyNode, configSlice []interface{}, subtrees map[string]*hierarchyNode) error {
	for _, item := range configSlice {
		if item == nil {
			continue
		}

		switch v := item.(type) {
		case string:
			// Simple string leaf folder
			parent.children = append(parent.children, &hierarchyNode{
				name: v,
			})
		case map[string]interface{}:
			for name, val := range v {
				node := &hierarchyNode{
					name: name,
				}

				switch innerVal := val.(type) {
				case []interface{}:
					if err := generateTree(node, innerVal, subtrees); err != nil {
						return err
					}
				case map[string]interface{}:
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
					// Map key with nil value = leaf folder
				}

				parent.children = append(parent.children, node)
			}
		}
	}
	return nil
}

// generateTree builds a hierarchy from children.
// Returns an error if a referenced subtree doesn't exist.
func generateTree(root *hierarchyNode, children []interface{}, subtrees map[string]*hierarchyNode) error {
	for _, child := range children {
		if child == nil {
			continue
		}

		switch v := child.(type) {
		case string:
			root.children = append(root.children, &hierarchyNode{
				name: v,
			})
		case map[string]interface{}:
			// Check if this map is a $subtree reference as a list item
			// e.g. config: [- $subtree: teams] — this should expand inline
			if subtreeName, ok := v["$subtree"]; ok && len(v) == 1 {
				stName, ok := subtreeName.(string)
				if !ok {
					return fmt.Errorf("$subtree value is not a string")
				}
				subtreeNode, exists := subtrees[stName]
				if !exists {
					return &missingSubtreeError{name: stName}
				}
				// Expand subtree children directly into parent
				root.children = append(root.children, subtreeNode.children...)
				continue
			}

			for name, val := range v {
				node := &hierarchyNode{
					name: name,
				}

				switch innerVal := val.(type) {
				case []interface{}:
					if err := generateTree(node, innerVal, subtrees); err != nil {
						return err
					}
				case map[string]interface{}:
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

// generateConfigs recursively generates Folders for each node.
func generateConfigs(node *hierarchyNode, path []string, annotations map[string]string, namespace string, isV3 bool, rl *fn.ResourceList) {
	for _, child := range node.children {
		folder := generateManifest(child.name, path, node, annotations, namespace, isV3)
		rl.Items = append(rl.Items, folder)
		newPath := append(clonePath(path), child.name)
		generateConfigs(child, newPath, annotations, namespace, isV3, rl)
	}
}

// generateManifest creates a Folder resource.
func generateManifest(name string, path []string, parent *hierarchyNode, annotations map[string]string, namespace string, nativeRef bool) *fn.KubeObject {
	// Build the folder object
	folderObj := fn.NewEmptyKubeObject()
	_ = folderObj.SetAPIVersion(folderAPIVersion)
	_ = folderObj.SetKind(folderKind)

	// Set metadata.name
	fullPath := make([]string, len(path))
	copy(fullPath, path)
	fullPath = append(fullPath, name)
	normalizedName := Normalize(fullPath)
	_ = folderObj.SetName(normalizedName)

	// Set namespace if provided
	if namespace != "" {
		_ = folderObj.SetNamespace(namespace)
	}

	// Build annotations
	combinedAnnotations := make(map[string]string)
	for k, v := range annotations {
		combinedAnnotations[k] = v
	}

	isRoot := len(path) == 0
	parentRef := parent.name
	if !isRoot {
		parentRef = strings.Join(path, ".")
	}

	if nativeRef {
		if isRoot {
			// Root node uses external refs
			if parent.kind == "Organization" {
				_ = folderObj.SetNestedField(parentRef, "spec", "organizationRef", "external")
			} else {
				_ = folderObj.SetNestedField(parentRef, "spec", "folderRef", "external")
			}
		} else {
			// Non-root uses name refs
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
		// Annotation-based refs (v2 and v1)
		if isRoot && parent.kind == "Organization" {
			combinedAnnotations["cnrm.cloud.google.com/organization-id"] = parentRef
		} else {
			combinedAnnotations["cnrm.cloud.google.com/folder-ref"] = normalize(parentRef)
		}
	}

	// Set annotations
	if len(combinedAnnotations) > 0 {
		for k, v := range combinedAnnotations {
			_ = folderObj.SetAnnotation(k, v)
		}
	}

	// Set spec.displayName
	_ = folderObj.SetNestedField(name, "spec", "displayName")

	return folderObj
}

// Normalize normalizes a path into a K8s DNS subdomain compatible name.
func Normalize(parts []string) string {
	return normalize(strings.Join(parts, "."))
}

// consecutivePunctRegex collapses consecutive dots and dashes.
var consecutivePunctRegex = regexp.MustCompile(`[-\.]{2,}`)

// trimPunctRegex matches leading/trailing non-alphanumeric characters.
var trimPunctRegex = regexp.MustCompile(`^[^a-z0-9]+|[^a-z0-9]+$`)

// normalize applies K8s DNS subdomain naming normalization to a string:
// 1. Convert to lowercase
// 2. Remove quotes
// 3. Replace underscores and spaces with dashes
// 4. Remove any remaining invalid characters
// 5. Collapse consecutive dashes/dots
// 6. Trim non-alphanumeric from start/end
// 7. Enforce 253-character max length
func normalize(name string) string {
	name = strings.ToLower(name)
	name = quoteRegex.ReplaceAllString(name, "")
	name = underscoreSpaceRegex.ReplaceAllString(name, "-")
	name = normalizeRegex.ReplaceAllString(name, "")
	name = consecutivePunctRegex.ReplaceAllString(name, "-")
	name = trimPunctRegex.ReplaceAllString(name, "")
	if len(name) > 253 {
		name = name[:253]
		name = trimPunctRegex.ReplaceAllString(name, "")
	}
	if name == "" {
		name = "unnamed"
	}
	return name
}

// filterNonInheritableAnnotations removes non-inheritable annotations.
func filterNonInheritableAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		return map[string]string{}
	}
	filtered := make(map[string]string)
	for k, v := range annotations {
		if !nonInheritableAnnotations[k] {
			filtered[k] = v
		}
	}
	return filtered
}

func clonePath(path []string) []string {
	cloned := make([]string, len(path))
	copy(cloned, path)
	return cloned
}

// oldHierarchyWarning generates a deprecation warning for v1/v2.
func oldHierarchyWarning(obj *fn.KubeObject) *fn.Result {
	return fn.GeneralResult(
		fmt.Sprintf("ResourceHierarchy %s references an older Resource Hierarchy GroupVersion. Latest GroupVersion is blueprints.cloud.google.com/v1alpha3.", obj.GetName()),
		fn.Warning,
	)
}
