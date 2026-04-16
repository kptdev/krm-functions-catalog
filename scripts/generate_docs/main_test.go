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
)

func TestParseSemver(t *testing.T) {
	tests := []struct {
		version              string
		major, minor, patch  int
		wantErr              bool
	}{
		{"v0.1.0", 0, 1, 0, false},
		{"v1.2.3", 1, 2, 3, false},
		{"v0.10.5", 0, 10, 5, false},
		{"invalid", 0, 0, 0, true},
		{"0.1.0", 0, 0, 0, true},
		{"", 0, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			major, minor, patch, err := parseSemver(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSemver(%q) error = %v, wantErr %v", tt.version, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if major != tt.major || minor != tt.minor || patch != tt.patch {
					t.Errorf("parseSemver(%q) = %d.%d.%d, want %d.%d.%d",
						tt.version, major, minor, patch, tt.major, tt.minor, tt.patch)
				}
			}
		})
	}
}

func TestCompareSemver(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"v0.1.0", "v0.2.0", -1},
		{"v0.2.0", "v0.1.0", 1},
		{"v0.1.0", "v0.1.0", 0},
		{"v0.1.1", "v0.1.2", -1},
		{"v1.0.0", "v0.9.9", 1},
		{"v0.10.0", "v0.9.0", 1},
		{"invalid", "v0.1.0", -1}, // falls back to string compare
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := compareSemver(tt.a, tt.b)
			if (tt.want < 0 && got >= 0) || (tt.want > 0 && got <= 0) || (tt.want == 0 && got != 0) {
				t.Errorf("compareSemver(%q, %q) = %d, want sign %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestGetLatestMinor(t *testing.T) {
	tags := []string{
		"functions/go/set-namespace/v0.2.0",
		"functions/go/set-namespace/v0.3.0",
		"functions/go/set-namespace/v0.3.1",
		"functions/go/set-namespace/v0.4.0",
		"functions/go/set-namespace/v0.4.5",
		"functions/go/other-fn/v0.1.0",
	}

	tests := []struct {
		fnName string
		want   string
	}{
		{"set-namespace", "v0.4"},
		{"other-fn", "v0.1"},
		{"nonexistent", ""},
	}
	for _, tt := range tests {
		t.Run(tt.fnName, func(t *testing.T) {
			got := getLatestMinor(tt.fnName, tags)
			if got != tt.want {
				t.Errorf("getLatestMinor(%q) = %q, want %q", tt.fnName, got, tt.want)
			}
		})
	}
}

func TestReadMetadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "metadata.yaml")

	t.Run("valid", func(t *testing.T) {
		os.WriteFile(path, []byte(`
image: ghcr.io/test
description: test
tags:
  - mutator
hidden: true
`), 0644)
		m, err := readMetadata(path)
		if err != nil {
			t.Fatal(err)
		}
		if m.Description != "test" {
			t.Errorf("description = %q, want %q", m.Description, "test")
		}
		if !m.Hidden {
			t.Error("expected hidden = true")
		}
		if len(m.Tags) != 1 || m.Tags[0] != "mutator" {
			t.Errorf("tags = %v, want [mutator]", m.Tags)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := readMetadata(filepath.Join(dir, "nonexistent.yaml"))
		if err == nil {
			t.Error("expected error for missing file")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		os.WriteFile(path, []byte(`{invalid`), 0644)
		_, err := readMetadata(path)
		if err == nil {
			t.Error("expected error for invalid YAML")
		}
	})
}

func TestGenerateIndex(t *testing.T) {
	data := docData{
		Name:        "my-fn",
		Tags:        "mutator, validator",
		Description: "A test function",
		ReadmeBody:  "\n## Overview\n\nSome content\n",
	}
	content, err := generateIndex(data)
	if err != nil {
		t.Fatal(err)
	}

	checks := []string{
		`title: "my-fn"`,
		`tags: "mutator, validator"`,
		`description: |`,
		`  A test function`,
		`{{< listversions >}}`,
		`{{< listexamples >}}`,
		`## Overview`,
		`Some content`,
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("generated content missing %q", check)
		}
	}
}

func TestProcessFunction(t *testing.T) {
	dir := t.TempDir()
	functionsDir := filepath.Join(dir, "functions", "go")
	docsDir := filepath.Join(dir, "documentation", "content", "en")

	setupFn := func(name, metadataContent, readmeContent string) {
		fnDir := filepath.Join(functionsDir, name)
		os.MkdirAll(fnDir, 0755)
		if metadataContent != "" {
			os.WriteFile(filepath.Join(fnDir, "metadata.yaml"), []byte(metadataContent), 0644)
		}
		if readmeContent != "" {
			os.WriteFile(filepath.Join(fnDir, "README.md"), []byte(readmeContent), 0644)
		}
	}

	tags := []string{"functions/go/test-fn/v0.1.0"}

	t.Run("generates doc", func(t *testing.T) {
		setupFn("test-fn",
			"image: ghcr.io/test\ndescription: Test\ntags:\n  - mutator\n",
			"# test-fn\n\n## Overview\n\nContent here\n")

		processFunction("test-fn", functionsDir, docsDir, tags, false)

		docFile := filepath.Join(docsDir, "test-fn", "v0.1", "_index.md")
		content, err := os.ReadFile(docFile)
		if err != nil {
			t.Fatalf("expected doc file to be created: %v", err)
		}
		if !strings.Contains(string(content), `title: "test-fn"`) {
			t.Error("doc missing title")
		}
		if !strings.Contains(string(content), "## Overview") {
			t.Error("doc missing README body")
		}
	})

	t.Run("dry run", func(t *testing.T) {
		dryDocsDir := filepath.Join(dir, "dry-docs")
		processFunction("test-fn", functionsDir, dryDocsDir, tags, true)

		docFile := filepath.Join(dryDocsDir, "test-fn", "v0.1", "_index.md")
		if _, err := os.Stat(docFile); !os.IsNotExist(err) {
			t.Error("dry run should not create files")
		}
	})

	t.Run("skip hidden", func(t *testing.T) {
		setupFn("hidden-fn",
			"image: ghcr.io/test\ndescription: Test\ntags:\n  - mutator\nhidden: true\n",
			"# hidden-fn\n\nContent\n")

		hiddenDocsDir := filepath.Join(dir, "hidden-docs")
		hiddenTags := []string{"functions/go/hidden-fn/v0.1.0"}
		processFunction("hidden-fn", functionsDir, hiddenDocsDir, hiddenTags, false)

		docFile := filepath.Join(hiddenDocsDir, "hidden-fn", "v0.1", "_index.md")
		if _, err := os.Stat(docFile); !os.IsNotExist(err) {
			t.Error("hidden function should not generate docs")
		}
	})

	t.Run("skip no tags", func(t *testing.T) {
		setupFn("no-tags-fn",
			"image: ghcr.io/test\ndescription: Test\ntags:\n  - mutator\n",
			"# no-tags-fn\n\nContent\n")

		noTagDocsDir := filepath.Join(dir, "notag-docs")
		processFunction("no-tags-fn", functionsDir, noTagDocsDir, []string{}, false)

		docFile := filepath.Join(noTagDocsDir, "no-tags-fn", "v0.1", "_index.md")
		if _, err := os.Stat(docFile); !os.IsNotExist(err) {
			t.Error("function with no tags should not generate docs")
		}
	})

	t.Run("skip no metadata", func(t *testing.T) {
		fnDir := filepath.Join(functionsDir, "no-meta-fn")
		os.MkdirAll(fnDir, 0755)
		os.WriteFile(filepath.Join(fnDir, "README.md"), []byte("# no-meta-fn\n"), 0644)

		noMetaDocsDir := filepath.Join(dir, "nometa-docs")
		processFunction("no-meta-fn", functionsDir, noMetaDocsDir, tags, false)

		if _, err := os.Stat(filepath.Join(noMetaDocsDir, "no-meta-fn")); !os.IsNotExist(err) {
			t.Error("function with no metadata should not generate docs")
		}
	})

	t.Run("skip no readme", func(t *testing.T) {
		fnDir := filepath.Join(functionsDir, "no-readme-fn")
		os.MkdirAll(fnDir, 0755)
		os.WriteFile(filepath.Join(fnDir, "metadata.yaml"), []byte("image: ghcr.io/test\ndescription: Test\ntags:\n  - mutator\n"), 0644)

		noReadmeDocsDir := filepath.Join(dir, "noreadme-docs")
		processFunction("no-readme-fn", functionsDir, noReadmeDocsDir, tags, false)

		if _, err := os.Stat(filepath.Join(noReadmeDocsDir, "no-readme-fn")); !os.IsNotExist(err) {
			t.Error("function with no README should not generate docs")
		}
	})

	t.Run("warn bad readme format", func(t *testing.T) {
		setupFn("bad-readme-fn",
			"image: ghcr.io/test\ndescription: Test\ntags:\n  - mutator\n",
			"not a heading\n\nContent\n")

		badDocsDir := filepath.Join(dir, "bad-docs")
		badTags := []string{"functions/go/bad-readme-fn/v0.1.0"}
		processFunction("bad-readme-fn", functionsDir, badDocsDir, badTags, false)

		if _, err := os.Stat(filepath.Join(badDocsDir, "bad-readme-fn")); !os.IsNotExist(err) {
			t.Error("function with bad README format should not generate docs")
		}
	})
}
