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

package generate_folders

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kptdev/krm-functions-sdk/go/fn"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple lowercase", "test", "test"},
		{"spaces to dashes", "test spaced", "test-spaced"},
		{"colon with space removed", "test: spaced colon", "test-spaced-colon"},
		{"colon without space removed", "test:colon", "testcolon"},
		{"mixed case with dots", "Environ Set.Environ.Team", "environ-set.environ.team"},
		{"quotes removed", `Team "One"`, "team-one"},
		{"underscores to dashes", "Team_2", "team-2"},
		{"single quotes removed", "Team 'One'", "team-one"},
		{"leading dashes trimmed", "_test_", "test"},
		{"consecutive dashes collapsed", "a--b", "a-b"},
		{"empty result fallback", "___", "unnamed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalize(tt.input)
			if result != tt.expected {
				t.Errorf("normalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		{"single part", []string{"Dev"}, "dev"},
		{"two parts", []string{"Dev", "Team_2"}, "dev.team-2"},
		{"three parts with quotes", []string{"Dev", `Team "One"`, "sub"}, "dev.team-one.sub"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.parts)
			if result != tt.expected {
				t.Errorf("Normalize(%v) = %q, want %q", tt.parts, result, tt.expected)
			}
		})
	}
}

func TestFilterNonInheritableAnnotations(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected map[string]string
	}{
		{"nil annotations", nil, map[string]string{}},
		{"empty annotations", map[string]string{}, map[string]string{}},
		{
			"filter local-config",
			map[string]string{
				"config.kubernetes.io/local-config":     "true",
				"cnrm.cloud.google.com/deletion-policy": "abandon",
			},
			map[string]string{
				"cnrm.cloud.google.com/deletion-policy": "abandon",
			},
		},
		{
			"filter all internal annotations",
			map[string]string{
				"config.kubernetes.io/local-config":     "true",
				"config.k8s.io/function":                "something",
				"internal.config.kubernetes.io/id":      "some-id",
				"config.kubernetes.io/id":               "some-id",
				"internal.config.kubernetes.io/path":    "path",
				"config.kubernetes.io/path":             "path",
				"internal.config.kubernetes.io/index":   "0",
				"config.kubernetes.io/index":            "0",
				"cnrm.cloud.google.com/deletion-policy": "abandon",
				"another-annotation":                    "value",
			},
			map[string]string{
				"cnrm.cloud.google.com/deletion-policy": "abandon",
				"another-annotation":                    "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterNonInheritableAnnotations(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("filterNonInheritableAnnotations() got %d entries, want %d", len(result), len(tt.expected))
				return
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("filterNonInheritableAnnotations()[%q] = %q, want %q", k, result[k], v)
				}
			}
		})
	}
}

func TestGenerateTree(t *testing.T) {
	t.Run("simple flat list", func(t *testing.T) {
		root := &hierarchyNode{name: "root", kind: "Organization"}
		children := []interface{}{"Dev", "Prod"}
		err := generateTree(root, children, nil)
		if err != nil {
			t.Fatalf("generateTree() returned error: %v", err)
		}

		if len(root.children) != 2 {
			t.Fatalf("expected 2 children, got %d", len(root.children))
		}
		if root.children[0].name != "Dev" {
			t.Errorf("expected child 0 to be Dev, got %s", root.children[0].name)
		}
		if root.children[1].name != "Prod" {
			t.Errorf("expected child 1 to be Prod, got %s", root.children[1].name)
		}
	})

	t.Run("nested structure", func(t *testing.T) {
		root := &hierarchyNode{name: "root", kind: "Organization"}
		children := []interface{}{
			map[string]interface{}{
				"Dev": []interface{}{"Team1", "Team2"},
			},
		}
		err := generateTree(root, children, nil)
		if err != nil {
			t.Fatalf("generateTree() returned error: %v", err)
		}

		if len(root.children) != 1 {
			t.Fatalf("expected 1 child, got %d", len(root.children))
		}
		dev := root.children[0]
		if dev.name != "Dev" {
			t.Errorf("expected child to be Dev, got %s", dev.name)
		}
		if len(dev.children) != 2 {
			t.Fatalf("expected 2 grandchildren, got %d", len(dev.children))
		}
		if dev.children[0].name != "Team1" {
			t.Errorf("expected grandchild 0 to be Team1, got %s", dev.children[0].name)
		}
	})

	t.Run("subtree expansion", func(t *testing.T) {
		subtrees := map[string]*hierarchyNode{
			"teams": {
				name: "teams",
				kind: "Subtree",
				children: []*hierarchyNode{
					{name: "Team1"},
					{name: "Team2"},
				},
			},
		}

		root := &hierarchyNode{name: "root", kind: "Organization"}
		children := []interface{}{
			map[string]interface{}{
				"Dev": map[string]interface{}{
					"$subtree": "teams",
				},
			},
		}
		err := generateTree(root, children, subtrees)
		if err != nil {
			t.Fatalf("generateTree() returned error: %v", err)
		}

		if len(root.children) != 1 {
			t.Fatalf("expected 1 child, got %d", len(root.children))
		}
		dev := root.children[0]
		if len(dev.children) != 2 {
			t.Fatalf("expected 2 grandchildren (from subtree), got %d", len(dev.children))
		}
		if dev.children[0].name != "Team1" {
			t.Errorf("expected Team1, got %s", dev.children[0].name)
		}
	})

	t.Run("subtree inline expansion in list", func(t *testing.T) {
		subtrees := map[string]*hierarchyNode{
			"teams": {
				name: "teams",
				kind: "Subtree",
				children: []*hierarchyNode{
					{name: "Team1"},
					{name: "Team2"},
				},
			},
		}

		// This mirrors YAML: Dev:\n  - $subtree: teams\n  - QA
		root := &hierarchyNode{name: "root", kind: "Organization"}
		children := []interface{}{
			map[string]interface{}{
				"Dev": []interface{}{
					map[string]interface{}{"$subtree": "teams"},
					"QA",
				},
			},
		}
		err := generateTree(root, children, subtrees)
		if err != nil {
			t.Fatalf("generateTree() returned error: %v", err)
		}

		if len(root.children) != 1 {
			t.Fatalf("expected 1 child (Dev), got %d", len(root.children))
		}
		dev := root.children[0]
		// Should have Team1, Team2 (from subtree) + QA = 3 children
		if len(dev.children) != 3 {
			t.Fatalf("expected 3 children (Team1,Team2,QA), got %d", len(dev.children))
		}
		if dev.children[0].name != "Team1" {
			t.Errorf("expected Team1, got %s", dev.children[0].name)
		}
		if dev.children[1].name != "Team2" {
			t.Errorf("expected Team2, got %s", dev.children[1].name)
		}
		if dev.children[2].name != "QA" {
			t.Errorf("expected QA, got %s", dev.children[2].name)
		}
	})

	t.Run("missing subtree reference returns error", func(t *testing.T) {
		root := &hierarchyNode{name: "root", kind: "Organization"}
		children := []interface{}{
			map[string]interface{}{"$subtree": "nonexistent"},
		}
		err := generateTree(root, children, map[string]*hierarchyNode{})
		if err == nil {
			t.Fatal("expected error for missing subtree, got nil")
		}
	})
}

func loadTestData(t *testing.T, filename string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("failed to read %s: %v", filename, err)
	}
	return data
}

// Integration tests using full YAML ResourceList inputs

func TestRunSimpleV3(t *testing.T) {
	input := loadTestData(t, "simple_v3.yaml")

	rl, err := fn.ParseResourceList(input)
	if err != nil {
		t.Fatalf("failed to parse resource list: %v", err)
	}

	ok, err := Run(rl)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	if !ok {
		t.Fatal("Run() returned false")
	}

	// Count generated folders (items minus the original ResourceHierarchy)
	folderCount := 0
	for _, item := range rl.Items {
		if item.GetKind() == "Folder" {
			folderCount++
		}
	}

	if folderCount < 4 {
		t.Errorf("expected at least 4 folders, got %d", folderCount)
	}

	// Verify first folder is Dev
	var devFolder *fn.KubeObject
	for _, item := range rl.Items {
		if item.GetKind() == "Folder" && item.GetName() == "dev" {
			devFolder = item
			break
		}
	}
	if devFolder == nil {
		t.Fatal("expected to find Folder with name 'dev'")
	}

	// Verify v3 uses native ref
	orgRef, found, _ := devFolder.NestedString("spec", "organizationRef", "external")
	if !found {
		t.Error("expected organizationRef.external to be set for root-level folder")
	}
	if orgRef != "test-organization" {
		t.Errorf("expected organizationRef.external = 'test-organization', got %q", orgRef)
	}

	// Verify displayName
	displayName, found, _ := devFolder.NestedString("spec", "displayName")
	if !found || displayName != "Dev" {
		t.Errorf("expected displayName = 'Dev', got %q", displayName)
	}
}

func TestRunV3WithNamespace(t *testing.T) {
	input := loadTestData(t, "v3_with_namespace.yaml")

	rl, err := fn.ParseResourceList(input)
	if err != nil {
		t.Fatalf("failed to parse resource list: %v", err)
	}

	ok, err := Run(rl)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	if !ok {
		t.Fatal("Run() returned false")
	}

	// Find the child folder (Team1)
	for _, item := range rl.Items {
		if item.GetKind() == "Folder" && item.GetName() == "dev.team1" {
			// Verify namespace is set
			ns := item.GetNamespace()
			if ns != "test-ns" {
				t.Errorf("expected namespace 'test-ns', got %q", ns)
			}

			// Verify depends-on annotation for non-root folder
			dependsOn := item.GetAnnotation(dependsOnAnnotation)
			expected := "resourcemanager.cnrm.cloud.google.com/namespaces/test-ns/Folder/dev"
			if dependsOn != expected {
				t.Errorf("expected depends-on annotation %q, got %q", expected, dependsOn)
			}
			return
		}
	}
	t.Error("expected to find Folder with name 'dev.team1'")
}

func TestRunV3WithFolderParent(t *testing.T) {
	input := loadTestData(t, "v3_folder_parent.yaml")

	rl, err := fn.ParseResourceList(input)
	if err != nil {
		t.Fatalf("failed to parse resource list: %v", err)
	}

	ok, err := Run(rl)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	if !ok {
		t.Fatal("Run() returned false")
	}

	// The root folder should use folderRef.external
	for _, item := range rl.Items {
		if item.GetKind() == "Folder" && item.GetName() == "dev" {
			ref, found, _ := item.NestedString("spec", "folderRef", "external")
			if !found || ref != "parent-folder-id" {
				t.Errorf("expected folderRef.external = 'parent-folder-id', got %q (found=%v)", ref, found)
			}
			return
		}
	}
	t.Error("expected to find Folder with name 'dev'")
}

func TestRunV2Deprecation(t *testing.T) {
	input := loadTestData(t, "v2_deprecation.yaml")

	rl, err := fn.ParseResourceList(input)
	if err != nil {
		t.Fatalf("failed to parse resource list: %v", err)
	}

	ok, err := Run(rl)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	if !ok {
		t.Fatal("Run() returned false")
	}

	// Should have at least one warning for deprecated version
	hasWarning := false
	for _, r := range rl.Results {
		if r.Severity == fn.Warning {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("expected deprecation warning for v2 hierarchy")
	}

	// Should use annotation-based refs for v2
	for _, item := range rl.Items {
		if item.GetKind() == "Folder" && item.GetName() == "dev" {
			orgAnnotation := item.GetAnnotation("cnrm.cloud.google.com/organization-id")
			if orgAnnotation != "test-organization" {
				t.Errorf("expected cnrm.cloud.google.com/organization-id = 'test-organization', got %q", orgAnnotation)
			}
			return
		}
	}
	t.Error("expected to find Folder with name 'dev'")
}

func TestRunMissingParentRef(t *testing.T) {
	input := loadTestData(t, "missing_parent_ref.yaml")

	rl, err := fn.ParseResourceList(input)
	if err != nil {
		t.Fatalf("failed to parse resource list: %v", err)
	}

	ok, err := Run(rl)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	if !ok {
		t.Fatal("Run() returned false")
	}

	// Should have an error result for missing parentRef
	hasError := false
	for _, r := range rl.Results {
		if r.Severity == fn.Error {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("expected error for missing parentRef")
	}
}

func TestRunAnnotationInheritance(t *testing.T) {
	input := loadTestData(t, "annotation_inheritance.yaml")

	rl, err := fn.ParseResourceList(input)
	if err != nil {
		t.Fatalf("failed to parse resource list: %v", err)
	}

	ok, err := Run(rl)
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	if !ok {
		t.Fatal("Run() returned false")
	}

	for _, item := range rl.Items {
		if item.GetKind() == "Folder" && item.GetName() == "dev" {
			// Should inherit deletion-policy
			deletionPolicy := item.GetAnnotation("cnrm.cloud.google.com/deletion-policy")
			if deletionPolicy != "abandon" {
				t.Errorf("expected inherited deletion-policy annotation, got %q", deletionPolicy)
			}

			// Should NOT inherit local-config
			localConfig := item.GetAnnotation("config.kubernetes.io/local-config")
			if localConfig != "" {
				t.Error("local-config annotation should not be inherited")
			}
			return
		}
	}
	t.Error("expected to find Folder with name 'dev'")
}
