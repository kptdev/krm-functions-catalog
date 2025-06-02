// Copyright 2021 Google LLC
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
//
// Usage: generate_catalog SRC_REPO_DIR/ DEST_MD_DIR/
//
// The command will create a README.md file under DEST_MD_DIR/ containing a table
// of collected functions. Source files for the catalog will also appear in
// DEST_MD_DIR/
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"
)

var keywordsPool = []string{
	"setters",
	"replacement",
	"setters",
	"name",
	"gcp",
	"terraform",
	"gatekeeper",
	"generate",
	"kubeval",
	"helm",
	"annotation",
	"image",
	"label",
	"namespace",
	"project-id",
	"starlark",
}

var (
	branchesToSkip = []string{
		"set-name-prefix/v0.0",
		"sops/v0.2",
	}
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: generate-catalog SRC_REPO_DIR/ DEST_MD_DIR/\n")
		os.Exit(1)
	}
	source, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	dest, err := filepath.Abs(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	branches, err := getBranches()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	functions := getFunctions(branches, source, dest)
	curatedFns, contribFns, err := writeFunctionIndex(functions, source, dest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = writeExampleIndex(curatedFns, source, dest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = writeExampleIndexV2(curatedFns, source, dest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	err = writeExampleIndex(contribFns, source, filepath.Join(dest, "contrib"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = writeExampleIndexV2(contribFns, source, filepath.Join(dest, "contrib"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

type function struct {
	FunctionName      string
	ImagePath         string
	VersionToExamples map[string]map[string]example
	LatestVersion     string
	Path              string
	Description       string
	Tags              string
	Gcp               bool
	Types             []string
	Keywords          []string
}

type example struct {
	LocalExamplePath  string
	RemoteExamplePath string
	RemoteSourcePath  string
}

type metadata struct {
	Image              string
	Description        string
	Tags               []string
	SourceUrl          string   `yaml:"sourceURL"`
	ExamplePackageUrls []string `yaml:"examplePackageURLs"`
	Hidden             bool
	Types              []string `yaml:"types"`
	Keywords           []string `yaml:"keywords"`
}

type catalogEntry struct {
	// LatestPatchVersion is the latest Patch semver for a MajorMinor semver
	LatestPatchVersion string
	// Examples maps exampleName->example
	Examples map[string]example
	// Types are the function types, only `validator` and `mutator` are accepted.
	Types    []string
	Keywords []string
}

var (
	// Match start of a version such as v1.9.1
	branchSemverPrefix = regexp.MustCompile(`[-\w]*\/(v\d*\.\d*)`)
	functionDirPrefix  = regexp.MustCompile(`.+/functions/`)
	patchTagPattern    = regexp.MustCompile(`.*(go|ts)\/[-\w]*\/(v\d*\.\d*\.\d*)`)
)

func getBranches() ([]string, error) {
	verBranches := make([]string, 0)

	var buf bytes.Buffer
	cmd := exec.Command("git", "branch", "-a")
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return verBranches, err
	}

	for _, branch := range strings.Split(buf.String(), "\n") {
		skipCurrentBranch := false
		for _, branchToSkip := range branchesToSkip {
			if strings.Contains(branch, branchToSkip) {
				skipCurrentBranch = true
				break
			}
		}
		if skipCurrentBranch {
			continue
		}

		if branchSemverPrefix.MatchString(branch) {
			verBranches = append(verBranches, strings.TrimSpace(branch))
		}
	}
	return verBranches, err
}

type functionRelease struct {
	FunctionName string
	MajorMinor   string
}

type patchVersionReader struct {
	latestPatchVersions map[functionRelease]string
}

// NewPatchVersionReader constructs a new patchVersionReader
func NewPatchVersionReader() (*patchVersionReader, error) {
	pvr := &patchVersionReader{}
	err := pvr.Init()
	return pvr, err
}

// Init initializes patchVersionReader
func (pvr *patchVersionReader) Init() error {
	pvr.latestPatchVersions = make(map[functionRelease]string)

	var buf bytes.Buffer
	cmd := exec.Command("git", "tag")
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return err
	}
	for _, tag := range strings.Split(buf.String(), "\n") {
		segments := strings.Split(tag, "/")
		// e.g. functions/go/some-fn/v1.2.3, go/some-fn/v1.2.3
		if len(segments) < 3 || !patchTagPattern.MatchString(tag) {
			continue
		}

		patchVersion := segments[len(segments)-1]
		fr := functionRelease{
			FunctionName: segments[len(segments)-2],
			MajorMinor:   semver.MajorMinor(patchVersion),
		}
		ver, ok := pvr.latestPatchVersions[fr]
		if !ok {
			pvr.latestPatchVersions[fr] = patchVersion
		} else if semver.Compare(patchVersion, ver) == 1 {
			pvr.latestPatchVersions[fr] = patchVersion
		}
	}
	return nil
}

// LatestPatchVersion for a given major/minor version of a function
func (pvr *patchVersionReader) LatestPatchVersion(funcName, majorMinor string) (string, error) {
	val, ok := pvr.latestPatchVersions[functionRelease{
		FunctionName: funcName,
		MajorMinor:   majorMinor,
	}]
	if !ok {
		return "", fmt.Errorf("could not find patch version for %s %s\n",
			funcName, majorMinor)
	}
	return val, nil
}

func getFunctions(branches []string, source string, dest string) []function {
	functions := make(map[string]function)
	for _, b := range branches {
		segments := strings.Split(b, "/")
		funcName := segments[len(segments)-2]
		minorVersion := segments[len(segments)-1]
		relativeFuncPath, err := getRelativeFunctionPath(source, funcName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}
		versionDest := filepath.Join(dest, funcName, minorVersion)
		if strings.Contains(relativeFuncPath, "contrib") {
			versionDest = filepath.Join(dest, "contrib", funcName, minorVersion)
		}

		// Functions with the hidden field enabled should not be processed.
		metadataPath := strings.TrimSpace(fmt.Sprintf("%v:%v", b, filepath.Join(relativeFuncPath, "metadata.yaml")))
		md, err := getMetadata(metadataPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting metadata for %q from %q: %v\n", funcName, b, err)
			os.Exit(1)
		}
		if md.Hidden {
			continue
		}
		err = copyExamples(b, md.ExamplePackageUrls, versionDest, minorVersion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting examples for %q from %q: %v\n", funcName, b, err)
			os.Exit(1)
		}

		err = copyReadme(b, funcName, relativeFuncPath, versionDest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting README for %q from %q: %v\n", funcName, b, err)
			os.Exit(1)
		}

		f := functions[funcName]
		f.FunctionName = funcName
		if f.VersionToExamples == nil {
			f.VersionToExamples = make(map[string]map[string]example)
		}

		functions[funcName] = parseMetadata(f, md, minorVersion, versionDest)
	}

	flattenedFunctions := make([]function, 0)
	for _, f := range functions {
		flattenedFunctions = append(flattenedFunctions, f)
	}
	sort.Slice(flattenedFunctions, func(i int, j int) bool {
		return flattenedFunctions[i].FunctionName < flattenedFunctions[j].FunctionName
	})
	return flattenedFunctions
}

func copyExamples(b string, exampleSources []string, versionDest, minorVersion string) error {
	// Copy examples for the function's version to a temporary directory.
	tempDir, err := ioutil.TempDir("", "examples")
	if err != nil {
		return err
	}

	// Prepare destination for versioned examples.
	err = os.MkdirAll(versionDest, 0744)
	if err != nil {
		return err
	}

	for _, exampleSource := range exampleSources {
		// We will split something like "https://github.com/kptdev/krm-functions-catalog/tree/set-namespace/v0.3.2/examples/set-namespace-simple" by something like "v0.3/" or "v0.3.2/"
		splitedPaths := regexp.MustCompile(fmt.Sprintf("%v(.[0-9])?/", minorVersion)).Split(exampleSource, 2)
		if len(splitedPaths) != 2 {
			return fmt.Errorf("expect 2 substring after spliting %q by %q", exampleSource, minorVersion)
		}
		relativePath := splitedPaths[1]
		// Fetch example into temporary directory.
		cmd := exec.Command("git", fmt.Sprintf("--work-tree=%v", tempDir), "checkout", b, "--", relativePath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Error running %v: %v\nCombined output: %v", cmd, err, string(out))
		}

		exampleName := filepath.Base(relativePath)
		src := filepath.Join(tempDir, relativePath)
		dest := filepath.Join(versionDest, exampleName)

		// Move example content to the site's example directory.
		if err = os.RemoveAll(dest); err != nil {
			return err
		}
		err = os.Rename(src, dest)
		if err != nil {
			return err
		}

	}

	return nil
}

func copyReadme(b string, funcName string, relativeFuncPath string, versionDest string) error {
	// Copy README for the function's version to the function's directory.
	tempDir, err := ioutil.TempDir("", "functions")
	if err != nil {
		return err
	}
	cmd := exec.Command("git", fmt.Sprintf("--work-tree=%v", tempDir), "checkout", b, "--", filepath.Join(relativeFuncPath, "README.md"))
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Error running %v: %v", cmd, err)
	}

	// Find the README in the function's directory.
	m, err := filepath.Glob(filepath.Join(tempDir, "functions", "*", funcName, "README.md"))
	if err != nil {
		return err
	}

	if m == nil {
		m, err = filepath.Glob(filepath.Join(tempDir, "contrib", "functions", "*", funcName, "README.md"))
		if err != nil {
			return err
		}
	}

	// Move the README to the destination directory.
	err = os.Rename(m[0], filepath.Join(versionDest, "README.md"))
	if err != nil {
		return err
	}

	return nil
}

func getMetadata(metadataPath string) (metadata, error) {
	var buf bytes.Buffer
	var md metadata
	// Get the content of metadata.yaml from the appropriate release branch.
	cmd := exec.Command("git", "cat-file", "blob", metadataPath)
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return md, err
	}

	yaml.Unmarshal(buf.Bytes(), &md)
	return md, nil
}

func parseMetadata(f function, md metadata, version string, versionDest string) function {

	// Add examples to version map.
	if f.VersionToExamples[version] == nil {
		f.VersionToExamples[version] = make(map[string]example)
	}
	e := f.VersionToExamples[version]
	for _, exUrl := range md.ExamplePackageUrls {
		exUrlSegments := strings.Split(exUrl, "/")
		exName := exUrlSegments[len(exUrlSegments)-1]
		ex := example{
			RemoteSourcePath:  md.SourceUrl,
			RemoteExamplePath: exUrl,
			LocalExamplePath:  filepath.Join(versionDest, exName),
		}
		e[exName] = ex
	}
	f.VersionToExamples[version] = e
	if semver.Compare(f.LatestVersion, version) == 1 {
		return f
	}

	// If this is the latest version,
	// update latest version, default path, description and rags.
	f.LatestVersion = version
	f.Path = versionDest
	f.Description = md.Description
	functionTags := make([]string, 0)
	if len(md.Types) != 0 {
		addType(&f, md.Types...)
	}
	for _, tag := range md.Tags {
		normalizedTag := strings.ToLower(tag)
		if normalizedTag == "gcp" {
			f.Gcp = true
		} else {
			functionTags = append(functionTags, normalizedTag)
		}
		// If function type is not given, read the "types" from "tags". The types have to be `mutator` or/and `validator`.
		// While "tags" can be custom field like `config sync`, `gcp`
		addType(&f, tag)
	}
	sort.Strings(functionTags)
	f.Tags = strings.Join(functionTags, ", ")
	f.ImagePath = md.Image
	addKeywordsFromTag(&f, md.Tags, md.Keywords)
	return f
}

// Ideally, we should read the keywords from the function metadata. However, each function has its own git release branch,
// so adding a single meta field requires code changing and releasing for each function branch.
// As a workaround before we take advantage of the "keywords" meta, we get keywords from two sources
// 1. For horizontal functions, we keep a KeywordPool and find keywords if it's a substring of the function.
// 2. For product specific functions, we use its `tags` excluding `mutator` and `validator` as keywords. These keywords can be
// `gcp`, `config-sync`, `terraform` etc.
func addKeywordsFromTag(f *function, tags []string, customKeywords []string) {
	var newKeywords []string
	keywordMap := map[string]bool{}
	keywords := append(tags, customKeywords...)
	for _, kw := range keywords {
		kw = strings.ToLower(kw)
		kw = strings.ReplaceAll(kw, " ", "-")
		if kw == "mutator" {
			continue
		}
		if kw == "validator" {
			continue
		}
		if ok := keywordMap[kw]; !ok {
			newKeywords = append(newKeywords, kw)
			keywordMap[kw] = true
		}
	}
	for _, kw := range keywordsPool {
		if strings.Contains(f.FunctionName, kw) {
			if ok := keywordMap[kw]; !ok {
				newKeywords = append(newKeywords, kw)
				keywordMap[kw] = true
			}
		}
	}
	f.Keywords = newKeywords
}

func addType(f *function, fnTypes ...string) {
	uniqueMap := map[string]bool{}
	allTypes := append(fnTypes, f.Types...)
	var newTypes []string
	for _, fnType := range allTypes {
		var k string
		switch strings.ToLower(fnType) {
		case "validator":
			k = "validator"
		case "mutator":
			k = "mutator"
		default:
			continue
		}
		if ok := uniqueMap[k]; !ok {
			uniqueMap[k] = true
			newTypes = append(newTypes, k)
		}
	}
	f.Types = newTypes
}

func getRelativeFunctionPath(source string, funcName string) (string, error) {
	// Find the directory for the function's source.
	sourcePattern := filepath.Join(source, "functions", "*", funcName)
	m, err := filepath.Glob(sourcePattern)
	if err != nil {
		return "", err
	}
	if m != nil {
		return functionDirPrefix.ReplaceAllString(m[0], "functions/"), nil
	}

	contribPattern := filepath.Join(source, "contrib", "functions", "*", funcName)
	m, err = filepath.Glob(contribPattern)
	if err != nil {
		return "", err
	}
	if m == nil {
		return "", fmt.Errorf("Could not find a function with the following name: %v", funcName)
	}
	return functionDirPrefix.ReplaceAllString(m[0], "contrib/functions/"), nil
}

func writeFunctionIndex(functions []function, source string, dest string) ([]function, []function, error) {
	out := []string{"# Curated Functions Catalog", ""}
	var contribOut []string

	genericFunctions := make([]function, 0)
	gcp := make([]function, 0)
	contribFunctions := make([]function, 0)
	for _, f := range functions {
		if strings.Contains(f.ImagePath, "contrib") {
			contribFunctions = append(contribFunctions, f)
		} else {
			if f.Gcp {
				gcp = append(gcp, f)
			} else {
				genericFunctions = append(genericFunctions, f)
			}
		}
	}

	out = append(out, getFunctionTable(genericFunctions, source)...)

	if len(gcp) > 0 {
		out = append(out, "", "## GCP Functions", "")
		out = append(out, getFunctionTable(gcp, source)...)
	}

	if len(contribFunctions) > 0 {
		contribOut = append(contribOut, "# Contrib Functions Catalog", "")
		contribOut = append(contribOut, getFunctionTable(contribFunctions, source)...)
	}

	o := strings.Join(out, "\n")
	if err := ioutil.WriteFile(filepath.Join(dest, "README.md"), []byte(o), 0744); err != nil {
		return nil, nil, err
	}
	co := strings.Join(contribOut, "\n")
	if err := ioutil.WriteFile(filepath.Join(dest, "contrib", "README.md"), []byte(co), 0744); err != nil {
		return nil, nil, err
	}
	return append(genericFunctions, gcp...), contribFunctions, nil
}

func getFunctionTable(functions []function, source string) []string {
	out := []string{"| Name | Description | Tags |", "| ---- | ----------- | ---- |"}
	for _, f := range functions {
		functionEntry := fmt.Sprintf("| [%v](%v/) | %v | %v |", f.FunctionName, strings.Replace(f.Path, filepath.Join(source, "site"), "", 1), f.Description, f.Tags)
		out = append(out, functionEntry)
	}
	return out
}

func writeExampleIndex(functions []function, source string, dest string) error {
	// Key a function's version's examples by the function's name -> version
	functionVersionMap := make(map[string]map[string]map[string]example)
	for _, f := range functions {
		vToE := make(map[string]map[string]example)
		for v, examples := range f.VersionToExamples {
			exampleToPaths := make(map[string]example)
			for exName, ex := range examples {
				e := ex
				e.LocalExamplePath = strings.Replace(ex.LocalExamplePath, filepath.Join(source, "site"), "", 1)
				exampleToPaths[exName] = e
			}
			vToE[v] = exampleToPaths
		}
		functionVersionMap[f.FunctionName] = vToE
	}

	funcJson, err := json.Marshal(functionVersionMap)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(dest, "catalog.json"), funcJson, 0600)
	return err
}

// writeExampleIndexV2 forms and writes the output for catalog-v2.json
func writeExampleIndexV2(functions []function, source string, dest string) error {
	// Map functionName->majorMinor->catalogEntry
	functionVersionMap := make(map[string]map[string]catalogEntry)
	pvr, err := NewPatchVersionReader()
	if err != nil {
		return err
	}
	for _, f := range functions {
		// majorMinor->CatalogEntry
		catalogEntryMap := make(map[string]catalogEntry)
		for majorMinor, examples := range f.VersionToExamples {
			// exampleName->example
			exampleMap := make(map[string]example)
			for exName, ex := range examples {
				e := ex
				e.LocalExamplePath = strings.Replace(ex.LocalExamplePath, filepath.Join(source, "site"), "", 1)
				exampleMap[exName] = e
			}
			patch, err := pvr.LatestPatchVersion(f.FunctionName, majorMinor)
			if err != nil {
				return err
			}
			catalogEntryMap[majorMinor] = catalogEntry{
				LatestPatchVersion: patch,
				Examples:           exampleMap,
				Types:              f.Types,
				Keywords:           f.Keywords,
			}
		}
		functionVersionMap[f.FunctionName] = catalogEntryMap
	}

	funcJson, err := json.Marshal(functionVersionMap)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(dest, "catalog-v2.json"), funcJson, 0600)
	return err
}
