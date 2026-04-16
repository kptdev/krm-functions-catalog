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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

const indexTemplate = `{{"<!--"}} DO NOT EDIT: generated from functions/go/{{.Name}}/README.md and metadata.yaml {{"-->"}}
---
title: "{{.Name}}"
linkTitle: "{{.Name}}"
tags: "{{.Tags}}"
weight: 4
description: |
  {{.Description}}
menu:
  main:
    parent: "Function Catalog"
---
{{"{{"}}< listversions >{{"}}"}}

{{"{{"}}< listexamples >{{"}}"}}
{{.ReadmeBody}}`

type metadata struct {
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
	Hidden      bool     `yaml:"hidden"`
}

type docData struct {
	Name        string
	Tags        string
	Description string
	ReadmeBody  string
}

var semverPattern = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

// parseSemver extracts major, minor, patch from a version string like "v0.4.5".
// Returns an error if the string doesn't match the expected pattern.
func parseSemver(version string) (major, minor, patch int, err error) {
	matches := semverPattern.FindStringSubmatch(version)
    if matches == nil {
        return 0, 0, 0, fmt.Errorf("invalid semver: %s", version)
    }
    if _, err := fmt.Sscanf(matches[1], "%d", &major); err != nil {
        return 0, 0, 0, fmt.Errorf("cannot parse major version: %w", err)
    }
    if _, err := fmt.Sscanf(matches[2], "%d", &minor); err != nil {
        return 0, 0, 0, fmt.Errorf("cannot parse minor version: %w", err)
    }
    if _, err := fmt.Sscanf(matches[3], "%d", &patch); err != nil {
        return 0, 0, 0, fmt.Errorf("cannot parse patch version: %w", err)
    }
    return major, minor, patch, nil
}

// compareSemver compares two semver strings.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
// Falls back to lexicographic comparison if either string is not valid semver.
func compareSemver(a, b string) int {
	aMajor, aMinor, aPatch, aErr := parseSemver(a)
	bMajor, bMinor, bPatch, bErr := parseSemver(b)
	if aErr != nil || bErr != nil {
		return strings.Compare(a, b)
	}
	switch {
	case aMajor != bMajor:
		return aMajor - bMajor
	case aMinor != bMinor:
		return aMinor - bMinor
	case aPatch != bPatch:
		return aPatch - bPatch
	default:
		return 0
	}
}

func getLatestMinor(fnName string, tags []string) string {
	prefix := fmt.Sprintf("functions/go/%s/", fnName)
	var versions []string
	for _, tag := range tags {
		if after, ok := strings.CutPrefix(tag, prefix); ok  {
			versions = append(versions, after)
		}
	}
	if len(versions) == 0 {
		return ""
	}
	sort.Slice(versions, func(i, j int) bool {
		return compareSemver(versions[i], versions[j]) < 0
	})
	latest := versions[len(versions)-1]
	parts := strings.Split(latest, ".")
	if len(parts) >= 3 {
		return parts[0] + "." + parts[1]
	}
	return latest
}

func readMetadata(path string) (*metadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m metadata
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func generateIndex(data docData) (string, error) {
	tmpl, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func processFunction(fnName, functionsDir, docsDir string, tags []string, dryRun bool) {
	fnDir := filepath.Join(functionsDir, fnName)
	metadataPath := filepath.Join(fnDir, "metadata.yaml")
	readmePath := filepath.Join(fnDir, "README.md")

	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		fmt.Printf("SKIP %s: no metadata.yaml\n", fnName)
		return
	}
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		fmt.Printf("SKIP %s: no README.md\n", fnName)
		return
	}

	readmeBytes, err := os.ReadFile(readmePath)
	if err != nil {
		fmt.Printf("SKIP %s: cannot read README.md: %v\n", fnName, err)
		return
	}
	firstLine := strings.SplitN(string(readmeBytes), "\n", 2)[0]
	if !strings.HasPrefix(firstLine, "# ") {
		fmt.Printf("WARN %s: README.md first line is not '# function-name', got: %s\n", fnName, firstLine)
		return
	}

	m, err := readMetadata(metadataPath)
	if err != nil {
		fmt.Printf("SKIP %s: cannot read metadata: %v\n", fnName, err)
		return
	}
	if m.Hidden {
		fmt.Printf("SKIP %s: hidden\n", fnName)
		return
	}

	latestMinor := getLatestMinor(fnName, tags)
	if latestMinor == "" {
		fmt.Printf("SKIP %s: no release tags\n", fnName)
		return
	}

	body := ""
	if idx := strings.Index(string(readmeBytes), "\n"); idx >= 0 {
		body = string(readmeBytes[idx+1:])
	}

	data := docData{
		Name:        fnName,
		Tags:        strings.Join(m.Tags, ", "),
		Description: m.Description,
		ReadmeBody:  body,
	}

	docDir := filepath.Join(docsDir, fnName, latestMinor)
	docFile := filepath.Join(docDir, "_index.md")

	if dryRun {
		fmt.Printf("WOULD generate %s\n", docFile)
		return
	}

	if err := os.MkdirAll(docDir, 0755); err != nil {
		fmt.Printf("ERROR %s: cannot create dir: %v\n", fnName, err)
		return
	}
	content, err := generateIndex(data)
	if err != nil {
		fmt.Printf("ERROR %s: cannot generate index: %v\n", fnName, err)
		return
	}
	if err := os.WriteFile(docFile, []byte(content), 0644); err != nil {
		fmt.Printf("ERROR %s: cannot write file: %v\n", fnName, err)
		return
	}
	fmt.Printf("GENERATED %s\n", docFile)
}
