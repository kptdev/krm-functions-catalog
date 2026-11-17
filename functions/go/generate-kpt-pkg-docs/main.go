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
	_ "embed"
	"errors"
	"fmt"
	"os"

	"github.com/kptdev/kpt/api/kubeobject"
	"github.com/kptdev/krm-functions-catalog/functions/go/generate-kpt-pkg-docs/docs"
	"github.com/kptdev/krm-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const defaultReadmePath = "/tmp/README.md"
const defaultRepoPath = "https://github.com/kptdev/kpt-samples.git/"

//go:embed README.md
var readme []byte

//go:embed metadata.yaml
var metadata []byte

// nolint
func main() {
	if err := fn.AsMain(kubeobject.ResourceListProcessorFunc(Process), fn.WithDocs(readme, metadata)); err != nil {
		os.Exit(1)
	}
}

func Process(resourceList *kubeobject.ResourceList) (bool, error) {
	readmePath, repoPath, pkgName := parseFnCfg(resourceList.FunctionConfig.CopyToResourceNode())

	currentDoc, err := os.ReadFile(readmePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			resourceList.Results = getResults(fmt.Sprintf("Skipping readme generation: %s", err), kubeobject.Warning)
			return true, nil
		} else {
			resourceList.Results = getResults(err.Error(), kubeobject.Error)
		}
	}
	err = generateReadme(repoPath, readmePath, pkgName, string(currentDoc), resourceList)
	if err != nil {
		resourceList.Results = getResults(err.Error(), kubeobject.Error)
		return false, err
	}
	return true, nil
}

func generateReadme(repoPath, readmePath, pkgName, currentDoc string, resourceList *kubeobject.ResourceList) error {
	title, generatedDoc, err := docs.GenerateBlueprintReadme(resourceList.Items.CopyToResourceNodes(), repoPath, pkgName)
	if err != nil {
		return err
	}
	readme, err := docs.InsertIntoReadme(title, currentDoc, generatedDoc)
	if err != nil {
		return err
	}
	err = os.WriteFile(readmePath, []byte(readme), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func parseFnCfg(r *yaml.RNode) (string, string, string) {
	cm := r.GetDataMap()
	readme, exists := cm["readme-path"]
	if !exists {
		readme = defaultReadmePath
	}
	repoPath, exists := cm["repo-path"]
	if !exists {
		repoPath = defaultRepoPath
	}
	pkgName := cm["pkg-name"]
	return readme, repoPath, pkgName

}

// getResults returns the item for input error message
func getResults(msg string, severity kubeobject.Severity) []*kubeobject.Result {
	return []*kubeobject.Result{
		{
			Message:  fmt.Sprintf("failed to generate doc: %s", msg),
			Severity: severity,
		},
	}
}
