// Copyright 2021 Google LLC
// Modifications Copyright (C) 2025 OpenInfra Foundation Europe.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kptdev/krm-functions-catalog/scripts/patch_reader/pkg/latestpatch"
	"gopkg.in/yaml.v2"
)

var (
	// pattern of release branches, e.g. apply-setters/v1.0
	releaseBranchPattern = regexp.MustCompile(`[-\w]*/(v\d*\.\d*)`)
	// pattern for version tags, e.g. latest, v0.1.1, v0.1
	versionGroup = `latest|v\d*\.\d*\.\d*|v\d*\.\d*`
)

func dirExists(path string) bool {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true
	}
	return false
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

type functionExample struct {
	ExamplePath string
	ExampleName string
}

type functionExamples []functionExample

// exampleNames returns a list of the functionExample names
func (fe functionExamples) exampleNames() []string {
	var exampleNames []string
	for _, example := range fe {
		exampleNames = append(exampleNames, example.ExampleName)
	}
	return exampleNames
}

type functionRelease struct {
	FunctionName       string
	MinorVersion       string
	Language           string
	LatestPatchVersion string
	FunctionPath       string
	Examples           functionExamples
	IsContrib          bool
}

// newFunctionRelease allocates and initializes a functionRelease
func newFunctionRelease(branch string) (*functionRelease, error) {
	fr := &functionRelease{}
	if !releaseBranchPattern.MatchString(branch) {
		return nil, fmt.Errorf("invalid branch format")
	}
	segments := strings.Split(branch, "/")
	// assume branch format: */<func_name>/<minor_version>
	fr.MinorVersion = segments[len(segments)-1]
	fr.FunctionName = segments[len(segments)-2]
	if err := fr.readLatestPatchVersion(); err != nil {
		return nil, err
	}
	if err := fr.readDocPaths(); err != nil {
		return nil, err
	}
	return fr, nil
}

// readLatestPatchVersion of the release from git tags
func (fr *functionRelease) readLatestPatchVersion() error {
	if fr.FunctionName == "" || fr.MinorVersion == "" {
		return fmt.Errorf("missing function name and/or minor version")
	}
	patch, err := latestpatch.GetLatestPatch(fr.FunctionName, fr.MinorVersion)
	if err != nil {
		return err
	}
	fr.Language = patch.Lang
	fr.LatestPatchVersion = patch.LatestPatch
	return nil
}

// readDocPaths and set documentation paths
func (fr *functionRelease) readDocPaths() error {
	executablePath, err := os.Executable()
	if err != nil {
		return err
	}
	repoBase := filepath.Dir(filepath.Dir(filepath.Dir(executablePath)))
	pathsToTry := []struct {
		functionPath string
		examplesPath string
		isContrib    bool
	}{
		{
			functionPath: filepath.Join(repoBase, "functions", fr.Language, fr.FunctionName),
			examplesPath: filepath.Join(repoBase, "examples"),
			isContrib:    false,
		},
		{
			functionPath: filepath.Join(repoBase, "contrib", "functions", fr.Language, fr.FunctionName),
			examplesPath: filepath.Join(repoBase, "contrib", "examples"),
			isContrib:    true,
		},
	}
	var examplesPath string
	for _, pathToTry := range pathsToTry {
		if dirExists(pathToTry.functionPath) {
			fr.FunctionPath = pathToTry.functionPath
			fr.IsContrib = pathToTry.isContrib
			examplesPath = pathToTry.examplesPath
			break
		}
	}
	if fr.FunctionPath == "" {
		return fmt.Errorf("function doc paths not found from %+v", pathsToTry)
	}
	if err = fr.parseMetadata(examplesPath); err != nil {
		return err
	}
	return nil
}

// parseMetadata from metadata.yaml and set example paths
func (fr *functionRelease) parseMetadata(examplesPath string) error {
	type metadata struct {
		ExamplePackageUrls []string `yaml:"examplePackageURLs"`
	}
	if fr.FunctionPath == "" {
		return fmt.Errorf("expected FunctionPath in parseMetadata")
	}

	metadataPath := filepath.Join(fr.FunctionPath, "metadata.yaml")
	var md metadata
	yamlFile, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, &md)
	if err != nil {
		return err
	}
	for _, exampleURL := range md.ExamplePackageUrls {
		segments := strings.Split(exampleURL, "/")
		exampleName := segments[len(segments)-1]
		examplePath := filepath.Join(examplesPath, exampleName)
		if !dirExists(examplePath) {
			return fmt.Errorf("example dir does not exist: %s", examplePath)
		}
		fr.Examples = append(fr.Examples, functionExample{
			ExamplePath: examplePath,
			ExampleName: exampleName,
		})
	}
	return nil
}

// updateDocs updates all the docs for the functionRelease on the filesystem
func (fr *functionRelease) updateDocs() error {
	if err := fr.updateFunctionDoc(); err != nil {
		return err
	}
	if err := fr.updateExampleDocs(); err != nil {
		return err
	}
	return nil
}

// updateFunctionDoc updates the function docs for the functionRelease
func (fr *functionRelease) updateFunctionDoc() error {
	functionReadme := filepath.Join(fr.FunctionPath, "README.md")
	if err := fr.updateDoc(functionReadme); err != nil {
		return err
	}
	functionMetadata := filepath.Join(fr.FunctionPath, "metadata.yaml")
	if err := fr.updateDoc(functionMetadata); err != nil {
		return err
	}
	return nil
}

// updateExampleDocs updates the example docs for the functionRelease
func (fr *functionRelease) updateExampleDocs() error {
	for _, example := range fr.Examples {
		exampleReadme := filepath.Join(example.ExamplePath, "README.md")
		if err := fr.updateDoc(exampleReadme); err != nil {
			return err
		}
		exampleKptfile := filepath.Join(example.ExamplePath, "Kptfile")
		if fileExists(exampleKptfile) {
			if err := fr.updateDoc(exampleKptfile); err != nil {
				return err
			}
		}
	}
	return nil
}

// Perform in place search/replace operations on a documentation file
func (fr *functionRelease) updateDoc(filePath string) error {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	contents = fr.replaceTags(contents)
	contents = fr.replaceURLs(contents)
	contents = fr.replaceKptPackages(contents)
	contents = fr.replaceGithubURLs(contents)
	if err = os.WriteFile(filePath, contents, 0644); err != nil {
		return err
	}
	return nil
}

// replace tags with patch e.g. apply-setters:v1.0.1, apply-setters/v1.0.1
func (fr *functionRelease) replaceTags(contents []byte) []byte {
	tagPattern := regexp.MustCompile(
		fmt.Sprintf(`(%s)(:|/)(%s)`, fr.FunctionName, versionGroup))
	contents = tagPattern.ReplaceAll(contents,
		[]byte(fmt.Sprintf(`${1}${2}%s`, fr.LatestPatchVersion)))
	return contents
}

// replace url with minor e.g. https://catalog.kpt.dev/apply-setters/v1.0
func (fr *functionRelease) replaceURLs(contents []byte) []byte {
	urlPattern := regexp.MustCompile(
		fmt.Sprintf(`(https://catalog\.kpt\.dev/%s/)(%s)`, fr.FunctionName, versionGroup))
	contents = urlPattern.ReplaceAll(contents,
		[]byte(fmt.Sprintf(`${1}%s`, fr.MinorVersion)))
	return contents
}

// get sub-path to examples e.g. examples, contrib/examples
func (fr *functionRelease) exampleSubPath() string {
	exampleSubPath := "examples"
	if fr.IsContrib {
		exampleSubPath = "contrib/examples"
	}
	return exampleSubPath
}

// replace kpt package names for all examples, e.g.
// https://github.com/kptdev/krm-functions-catalog.git/examples/apply-setters-simple ->
// https://github.com/kptdev/krm-functions-catalog.git/examples/apply-setters-simple@apply-setters/v1.0.1
func (fr *functionRelease) replaceKptPackages(contents []byte) []byte {
	exampleGroup := strings.Join(fr.Examples.exampleNames(), "|")
	exampleSubPath := fr.exampleSubPath()
	kptPkgPattern := regexp.MustCompile(
		fmt.Sprintf(`(https://github\.com/kptdev/krm-functions-catalog\.git/%s/)(%s)(\s+)`,
			exampleSubPath, exampleGroup))
	contents = kptPkgPattern.ReplaceAll(contents,
		[]byte(fmt.Sprintf(`${1}${2}@%s/%s${3}`, fr.FunctionName, fr.LatestPatchVersion)))
	return contents
}

// replace branch name with release branch for all GitHub URLs, e.g.
// https://github.com/kptdev/krm-functions-catalog/tree/main/examples/set-namespace-simple ->
// https://github.com/kptdev/krm-functions-catalog/tree/set-namespace/v0.2/examples/set-namespace-simple
func (fr *functionRelease) replaceGithubURLs(contents []byte) []byte {
	exampleSubPath := fr.exampleSubPath()
	sourceUrlSuffix := fmt.Sprintf(`/functions/%s/%s`, fr.Language, fr.FunctionName)
	if fr.IsContrib {
		sourceUrlSuffix = fmt.Sprintf(`/contrib/functions/%s/%s`, fr.Language, fr.FunctionName)
	}
	suffixes := []string{
		sourceUrlSuffix,
	}
	for _, ex := range fr.Examples.exampleNames() {
		suffixes = append(suffixes, fmt.Sprintf(`/%s/%s`, exampleSubPath, ex))
	}
	suffixGroup := strings.Join(suffixes, "|")
	refGroup := fmt.Sprintf(`main|%[1]s/v\d*\.\d*\.\d*|%[1]s/v\d*\.\d*`, fr.FunctionName)
	githubURLPattern := regexp.MustCompile(
		fmt.Sprintf(`(https://github\.com/kptdev/krm-functions-catalog/tree/)(%s)(%s)`,
			refGroup, suffixGroup))
	contents = githubURLPattern.ReplaceAll(contents,
		[]byte(fmt.Sprintf(`${1}%s/%s${3}`, fr.FunctionName, fr.MinorVersion)))
	return contents
}
