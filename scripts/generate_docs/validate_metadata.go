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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

func validateMetadata(metadataPath string, schema *jsonschema.Schema, functionsDir string) []string {
	var errs []string

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return []string{fmt.Sprintf("cannot read file: %v", err)}
	}

	var yamlData any
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return []string{fmt.Sprintf("invalid YAML: %v", err)}
	}
	jsonBytes, err := json.Marshal(yamlData)
	if err != nil {
		return []string{fmt.Sprintf("cannot convert to JSON: %v", err)}
	}
	var jsonData any
	if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
		return []string{fmt.Sprintf("cannot parse JSON: %v", err)}
	}

	if err := schema.Validate(jsonData); err != nil {
		errs = append(errs, err.Error())
	}

	// Check example directories exist on disk (not expressible in JSON schema)
	repoRoot := filepath.Dir(filepath.Dir(functionsDir))
	var m struct {
		ExamplePackageURLs []string `yaml:"examplePackageURLs"`
	}
	if err := yaml.Unmarshal(data, &m); err != nil {
		errs = append(errs, fmt.Sprintf("cannot parse examplePackageURLs: %v", err))
		return errs
	}
	examplePrefix := "https://github.com/kptdev/krm-functions-catalog/tree/main/examples/"
	for _, u := range m.ExamplePackageURLs {
		if after, ok := strings.CutPrefix(u, examplePrefix); ok {
			exampleDir := filepath.Join(repoRoot, "examples", after)
			if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
				errs = append(errs, fmt.Sprintf("example directory does not exist: examples/%s", after))
			}
		}
	}

	return errs
}

func runValidation(schemaPath string, functionsDir string) error {
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(schemaPath)
	if err != nil {
		return fmt.Errorf("cannot compile schema: %w", err)
	}

	entries, err := os.ReadDir(functionsDir)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", functionsDir, err)
	}

	failed := false
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "_template" {
			continue
		}
		path := filepath.Join(functionsDir, entry.Name(), "metadata.yaml")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}
		errs := validateMetadata(path, schema, functionsDir)
		if len(errs) > 0 {
			fmt.Printf("FAIL %s:\n", entry.Name())
			for _, e := range errs {
				fmt.Printf("  - %s\n", e)
			}
			failed = true
		} else {
			fmt.Printf("OK %s\n", entry.Name())
		}
	}
	if failed {
		return fmt.Errorf("one or more metadata files failed validation")
	}
	fmt.Println("\nAll metadata files valid")
	return nil
}
