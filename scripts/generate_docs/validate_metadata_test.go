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

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func loadTestSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile("metadata-schema.json")
	if err != nil {
		t.Fatalf("failed to compile schema: %v", err)
	}
	return schema
}

func writeMetadata(t *testing.T, dir string, content string) string {
	t.Helper()
	fnDir := filepath.Join(dir, "test-fn")
	if err := os.MkdirAll(fnDir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(fnDir, "metadata.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestValidMetadata(t *testing.T) {
	schema := loadTestSchema(t)
	dir := t.TempDir()
	path := writeMetadata(t, dir, `
image: ghcr.io/kptdev/krm-functions-catalog/my-fn
description: A test function
tags:
  - mutator
sourceURL: https://github.com/kptdev/krm-functions-catalog/tree/main/functions/go/my-fn
license: Apache-2.0
`)
	errs := validateMetadata(path, schema, dir)
	if len(errs) > 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestMissingRequiredFields(t *testing.T) {
	schema := loadTestSchema(t)
	dir := t.TempDir()

	tests := []struct {
		name    string
		content string
		expect  string
	}{
		{
			name: "missing image",
			content: `
description: A test function
tags:
  - mutator
`,
			expect: "image",
		},
		{
			name: "missing description",
			content: `
image: ghcr.io/kptdev/krm-functions-catalog/my-fn
tags:
  - mutator
`,
			expect: "description",
		},
		{
			name: "missing tags",
			content: `
image: ghcr.io/kptdev/krm-functions-catalog/my-fn
description: A test function
`,
			expect: "tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeMetadata(t, dir, tt.content)
			errs := validateMetadata(path, schema, dir)
			if len(errs) == 0 {
				t.Error("expected validation errors, got none")
				return
			}
			found := false
			for _, e := range errs {
				if strings.Contains(e, tt.expect) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected error containing %q, got: %v", tt.expect, errs)
			}
		})
	}
}

func TestInvalidImagePrefix(t *testing.T) {
	schema := loadTestSchema(t)
	dir := t.TempDir()
	path := writeMetadata(t, dir, `
image: docker.io/my-fn
description: A test function
tags:
  - mutator
`)
	errs := validateMetadata(path, schema, dir)
	if len(errs) == 0 {
		t.Error("expected validation error for image prefix, got none")
	}
}

func TestInvalidTag(t *testing.T) {
	schema := loadTestSchema(t)
	dir := t.TempDir()
	path := writeMetadata(t, dir, `
image: ghcr.io/kptdev/krm-functions-catalog/my-fn
description: A test function
tags:
  - not-a-real-tag
`)
	errs := validateMetadata(path, schema, dir)
	if len(errs) == 0 {
		t.Error("expected validation error for invalid tag, got none")
	}
}

func TestEmptyTags(t *testing.T) {
	schema := loadTestSchema(t)
	dir := t.TempDir()
	path := writeMetadata(t, dir, `
image: ghcr.io/kptdev/krm-functions-catalog/my-fn
description: A test function
tags: []
`)
	errs := validateMetadata(path, schema, dir)
	if len(errs) == 0 {
		t.Error("expected validation error for empty tags, got none")
	}
}

func TestMultipleTags(t *testing.T) {
	schema := loadTestSchema(t)
	dir := t.TempDir()
	path := writeMetadata(t, dir, `
image: ghcr.io/kptdev/krm-functions-catalog/my-fn
description: A test function
tags:
  - mutator
  - validator
`)
	errs := validateMetadata(path, schema, dir)
	if len(errs) > 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestHiddenField(t *testing.T) {
	schema := loadTestSchema(t)
	dir := t.TempDir()
	path := writeMetadata(t, dir, `
image: ghcr.io/kptdev/krm-functions-catalog/my-fn
description: A test function
hidden: true
tags:
  - mutator
`)
	errs := validateMetadata(path, schema, dir)
	if len(errs) > 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestAdditionalPropertiesRejected(t *testing.T) {
	schema := loadTestSchema(t)
	dir := t.TempDir()
	path := writeMetadata(t, dir, `
image: ghcr.io/kptdev/krm-functions-catalog/my-fn
description: A test function
tags:
  - mutator
bogusField: should not be here
`)
	errs := validateMetadata(path, schema, dir)
	if len(errs) == 0 {
		t.Error("expected validation error for additional property, got none")
	}
}

func TestMissingExampleDirectory(t *testing.T) {
	schema := loadTestSchema(t)
	dir := t.TempDir()
	// Create a fake functions/go structure so repoRoot resolves
	functionsDir := filepath.Join(dir, "functions", "go")
	os.MkdirAll(functionsDir, 0755)

	fnDir := filepath.Join(functionsDir, "test-fn")
	os.MkdirAll(fnDir, 0755)
	path := filepath.Join(fnDir, "metadata.yaml")
	content := `
image: ghcr.io/kptdev/krm-functions-catalog/my-fn
description: A test function
tags:
  - mutator
examplePackageURLs:
  - https://github.com/kptdev/krm-functions-catalog/tree/main/examples/nonexistent-example
`
	os.WriteFile(path, []byte(content), 0644)

	errs := validateMetadata(path, schema, functionsDir)
	found := false
	for _, e := range errs {
		if strings.Contains(e, "example directory does not exist") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about missing example directory, got: %v", errs)
	}
}

func TestExistingExampleDirectory(t *testing.T) {
	schema := loadTestSchema(t)
	dir := t.TempDir()
	// Create repo structure with example dir
	functionsDir := filepath.Join(dir, "functions", "go")
	os.MkdirAll(functionsDir, 0755)
	os.MkdirAll(filepath.Join(dir, "examples", "my-fn-simple"), 0755)

	fnDir := filepath.Join(functionsDir, "test-fn")
	os.MkdirAll(fnDir, 0755)
	path := filepath.Join(fnDir, "metadata.yaml")
	content := `
image: ghcr.io/kptdev/krm-functions-catalog/my-fn
description: A test function
tags:
  - mutator
examplePackageURLs:
  - https://github.com/kptdev/krm-functions-catalog/tree/main/examples/my-fn-simple
`
	os.WriteFile(path, []byte(content), 0644)

	errs := validateMetadata(path, schema, functionsDir)
	if len(errs) > 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}
