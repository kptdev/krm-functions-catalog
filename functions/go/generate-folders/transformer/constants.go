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

import "regexp"

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
	organizationKind = "Organization"
	subtreeKind      = "Subtree"

	// Annotation constants
	dependsOnAnnotation      = "config.kubernetes.io/depends-on"
	organizationIDAnnotation = "cnrm.cloud.google.com/organization-id"
	folderRefAnnotation      = "cnrm.cloud.google.com/folder-ref"
	localConfigAnnotation    = "config.kubernetes.io/local-config"
	functionAnnotation       = "config.k8s.io/function"
	internalIDAnnotation     = "internal.config.kubernetes.io/id"
	configIDAnnotation       = "config.kubernetes.io/id"
	internalPathAnnotation   = "internal.config.kubernetes.io/path"
	configPathAnnotation     = "config.kubernetes.io/path"
	internalIndexAnnotation  = "internal.config.kubernetes.io/index"
	configIndexAnnotation    = "config.kubernetes.io/index"
)

// Annotations that should not be inherited by generated Folders.
var nonInheritableAnnotations = map[string]bool{
	localConfigAnnotation:   true,
	functionAnnotation:      true,
	internalIDAnnotation:    true,
	configIDAnnotation:      true,
	internalPathAnnotation:  true,
	configPathAnnotation:    true,
	internalIndexAnnotation: true,
	configIndexAnnotation:   true,
}

var normalizeRegex = regexp.MustCompile(`[^a-z0-9.\- ]`)
var quoteRegex = regexp.MustCompile(`['"]`)
var underscoreSpaceRegex = regexp.MustCompile(`[_ ]`)
var consecutivePunctRegex = regexp.MustCompile(`[-\.]{2,}`)
var trimPunctRegex = regexp.MustCompile(`^[^a-z0-9]+|[^a-z0-9]+$`)

// hierarchyNode represents a node in the folder hierarchy tree.
type hierarchyNode struct {
	name     string
	kind     string
	children []*hierarchyNode
}

type missingSubtreeError struct {
	name string
}

type sourcePlacement struct {
	path      string
	nextIndex int
}

func (e *missingSubtreeError) Error() string {
	return e.name
}
