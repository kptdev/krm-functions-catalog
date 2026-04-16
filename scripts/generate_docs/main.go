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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func getGitTags() ([]string, error) {
	out, err := exec.Command("git", "tag").Output()
	if err != nil {
		return nil, err
	}
	var tags []string
	for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			tags = append(tags, line)
		}
	}
	return tags, nil
}

func resolveScriptDir() string {
	ex, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(ex)
		if _, err := os.Stat(filepath.Join(dir, "metadata-schema.json")); err == nil {
			return dir
		}
	}
	cwd, _ := os.Getwd()
	return cwd
}

func cmdGenerate(args []string, scriptDir string) error {
	dryRun := false
	targetFn := ""
	for _, arg := range args {
		switch arg {
		case "--dry-run":
			dryRun = true
		default:
			targetFn = arg
		}
	}

	repoRoot := filepath.Dir(filepath.Dir(scriptDir))
	functionsDir := filepath.Join(repoRoot, "functions", "go")
	docsDir := filepath.Join(repoRoot, "documentation", "content", "en")

	fmt.Println("Validating metadata...")
	schemaPath := filepath.Join(scriptDir, "metadata-schema.json")
	if err := runValidation(schemaPath, functionsDir); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	fmt.Println("\nFetching tags...")
	_ = exec.Command("git", "fetch", "--tags", "--quiet").Run()

	tags, err := getGitTags()
	if err != nil {
		return fmt.Errorf("cannot get git tags: %w", err)
	}

	if targetFn != "" {
		if _, err := os.Stat(filepath.Join(functionsDir, targetFn)); os.IsNotExist(err) {
			return fmt.Errorf("function '%s' not found in %s", targetFn, functionsDir)
		}
		processFunction(targetFn, functionsDir, docsDir, tags, dryRun)
	} else {
		entries, err := os.ReadDir(functionsDir)
		if err != nil {
			return fmt.Errorf("cannot read %s: %w", functionsDir, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() || entry.Name() == "_template" {
				continue
			}
			processFunction(entry.Name(), functionsDir, docsDir, tags, dryRun)
		}
	}
	return nil
}

func cmdValidate(args []string, scriptDir string) error {
	functionsDir := filepath.Join(filepath.Dir(filepath.Dir(scriptDir)), "functions", "go")
	if len(args) > 0 {
		functionsDir = args[0]
	}
	schemaPath := filepath.Join(scriptDir, "metadata-schema.json")
	return runValidation(schemaPath, functionsDir)
}

func usage() {
	fmt.Println("Usage: generate_docs <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  generate [--dry-run] [function-name]   Generate/sync Hugo doc pages")
	fmt.Println("  validate [functions-dir]               Validate metadata.yaml files")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  generate_docs generate                        # sync all functions")
	fmt.Println("  generate_docs generate set-namespace          # sync one function")
	fmt.Println("  generate_docs generate --dry-run              # preview all")
	fmt.Println("  generate_docs validate                        # validate all metadata")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	scriptDir := resolveScriptDir()
	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "generate":
		err = cmdGenerate(args, scriptDir)
	case "validate":
		err = cmdValidate(args, scriptDir)
	case "--help", "-h", "help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}
