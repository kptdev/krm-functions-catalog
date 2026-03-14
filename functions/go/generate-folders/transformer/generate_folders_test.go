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
	"errors"
	"testing"

	"github.com/kptdev/krm-functions-sdk/go/fn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			assert.Equal(t, tt.expected, normalize(tt.input))
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
			assert.Equal(t, tt.expected, Normalize(tt.parts))
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
			assert.Equal(t, tt.expected, filterNonInheritableAnnotations(tt.input))
		})
	}
}

func TestGenerateTree(t *testing.T) {
	t.Run("simple flat list", func(t *testing.T) {
		root := &hierarchyNode{name: "root", kind: "Organization"}
		err := generateTree(root, []interface{}{"Dev", "Prod"}, nil)

		require.NoError(t, err)
		require.Len(t, root.children, 2)
		assert.Equal(t, "Dev", root.children[0].name)
		assert.Equal(t, "Prod", root.children[1].name)
	})

	t.Run("nested structure", func(t *testing.T) {
		root := &hierarchyNode{name: "root", kind: "Organization"}
		children := []interface{}{
			map[string]interface{}{
				"Dev": []interface{}{"Team1", "Team2"},
			},
		}

		err := generateTree(root, children, nil)

		require.NoError(t, err)
		require.Len(t, root.children, 1)
		require.Len(t, root.children[0].children, 2)
		assert.Equal(t, "Dev", root.children[0].name)
		assert.Equal(t, "Team1", root.children[0].children[0].name)
		assert.Equal(t, "Team2", root.children[0].children[1].name)
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

		require.NoError(t, err)
		require.Len(t, root.children, 1)
		require.Len(t, root.children[0].children, 2)
		assert.Equal(t, "Team1", root.children[0].children[0].name)
		assert.Equal(t, "Team2", root.children[0].children[1].name)
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

		require.NoError(t, err)
		require.Len(t, root.children, 1)
		require.Len(t, root.children[0].children, 3)
		assert.Equal(t, "Team1", root.children[0].children[0].name)
		assert.Equal(t, "Team2", root.children[0].children[1].name)
		assert.Equal(t, "QA", root.children[0].children[2].name)
	})

	t.Run("missing subtree reference returns typed error", func(t *testing.T) {
		root := &hierarchyNode{name: "root", kind: "Organization"}
		err := generateTree(root, []interface{}{map[string]interface{}{"$subtree": "nonexistent"}}, map[string]*hierarchyNode{})

		require.Error(t, err)
		var missingErr *missingSubtreeError
		require.True(t, errors.As(err, &missingErr))
		assert.Equal(t, "nonexistent", missingErr.name)
	})

	t.Run("invalid subtree value returns parse error", func(t *testing.T) {
		root := &hierarchyNode{name: "root", kind: "Organization"}
		err := generateTree(root, []interface{}{map[string]interface{}{"$subtree": 123}}, map[string]*hierarchyNode{})

		require.EqualError(t, err, "$subtree value is not a string")
	})
}

func TestBuildTreeFromRawConfigAllowsForwardSubtreeReferences(t *testing.T) {
	root := &hierarchyNode{name: "root", kind: "Organization"}
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
	config := []interface{}{
		map[string]interface{}{
			"Dev": map[string]interface{}{"$subtree": "teams"},
		},
	}

	err := buildTreeFromRawConfig(root, config, subtrees)

	require.NoError(t, err)
	require.Len(t, root.children, 1)
	require.Len(t, root.children[0].children, 2)
	assert.Equal(t, "dev.team1", generatedFolderName(root.children[0].children[0].name, []string{"Dev"}))
}

func TestBuildTreeFromRawConfigPreservesParseErrors(t *testing.T) {
	root := &hierarchyNode{name: "root", kind: "Organization"}
	config := []interface{}{
		map[string]interface{}{
			"Dev": map[string]interface{}{"$subtree": 123},
		},
	}

	err := buildTreeFromRawConfig(root, config, map[string]*hierarchyNode{})

	require.EqualError(t, err, "$subtree value is not a string")
}

func TestRunV1HierarchyUsesAncestorPathInFolderNames(t *testing.T) {
	rl, err := fn.ParseResourceList([]byte(`
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
  - apiVersion: dev.cft.v1alpha1
    kind: ResourceHierarchy
    metadata:
      name: v1-hierarchy
    spec:
      organization: "123456789"
      layers:
        - environments
        - teams
      config:
        environments:
          - Dev
        teams:
          - Team1
`))
	require.NoError(t, err)

	ok, err := Run(rl)

	require.NoError(t, err)
	require.True(t, ok)

	folders := map[string]*fn.KubeObject{}
	for _, item := range rl.Items {
		if item.GetKind() == folderKind {
			folders[item.GetName()] = item
		}
	}

	require.Contains(t, folders, "dev")
	require.Contains(t, folders, "dev.team1")
	assert.Equal(t, "123456789", folders["dev"].GetAnnotation("cnrm.cloud.google.com/organization-id"))
	assert.Equal(t, "dev", folders["dev.team1"].GetAnnotation("cnrm.cloud.google.com/folder-ref"))
}

func TestRunV3HierarchyPreservesSourceFilePlacement(t *testing.T) {
	rl, err := fn.ParseResourceList([]byte(`
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
  - apiVersion: blueprints.cloud.google.com/v1alpha3
    kind: ResourceHierarchy
    metadata:
      name: test
      annotations:
        config.kubernetes.io/path: resources.yaml
        internal.config.kubernetes.io/path: resources.yaml
        config.kubernetes.io/index: "0"
        internal.config.kubernetes.io/index: "0"
    spec:
      parentRef:
        kind: Organization
        external: "123456789"
      config:
        - Dev:
            - Team1
`))
	require.NoError(t, err)

	ok, err := Run(rl)

	require.NoError(t, err)
	require.True(t, ok)

	var folders []*fn.KubeObject
	for _, item := range rl.Items {
		if item.GetKind() == folderKind {
			folders = append(folders, item)
		}
	}

	require.Len(t, folders, 2)
	assert.Equal(t, "resources.yaml", folders[0].GetAnnotation(configPathAnnotation))
	assert.Equal(t, "resources.yaml", folders[0].GetAnnotation(internalPathAnnotation))
	assert.Equal(t, "1", folders[0].GetAnnotation(configIndexAnnotation))
	assert.Equal(t, "1", folders[0].GetAnnotation(internalIndexAnnotation))
	assert.Equal(t, "resources.yaml", folders[1].GetAnnotation(configPathAnnotation))
	assert.Equal(t, "resources.yaml", folders[1].GetAnnotation(internalPathAnnotation))
	assert.Equal(t, "2", folders[1].GetAnnotation(configIndexAnnotation))
	assert.Equal(t, "2", folders[1].GetAnnotation(internalIndexAnnotation))
}

func TestRunV3HierarchyIsIdempotentWithSourceFilePlacement(t *testing.T) {
	rl, err := fn.ParseResourceList([]byte(`
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
  - apiVersion: blueprints.cloud.google.com/v1alpha3
    kind: ResourceHierarchy
    metadata:
      name: test
      annotations:
        config.kubernetes.io/path: resources.yaml
        internal.config.kubernetes.io/path: resources.yaml
        config.kubernetes.io/index: "0"
        internal.config.kubernetes.io/index: "0"
    spec:
      parentRef:
        kind: Organization
        external: "123456789"
      config:
        - Dev:
            - Team1
`))
	require.NoError(t, err)

	ok, err := Run(rl)
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = Run(rl)
	require.NoError(t, err)
	require.True(t, ok)

	var folders []*fn.KubeObject
	for _, item := range rl.Items {
		if item.GetKind() == folderKind {
			folders = append(folders, item)
		}
	}

	require.Len(t, folders, 2)
	assert.Equal(t, "dev", folders[0].GetName())
	assert.Equal(t, "dev.team1", folders[1].GetName())
	assert.Equal(t, "1", folders[0].GetAnnotation(configIndexAnnotation))
	assert.Equal(t, "2", folders[1].GetAnnotation(configIndexAnnotation))
}

func generatedFolderName(name string, path []string) string {
	return generateManifest(name, path, &hierarchyNode{name: "parent", kind: "Folder"}, nil, "", false, nil).GetName()
}
