package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/kptdev/krm-functions-catalog/functions/go/generate-kpt-pkg-docs/docs"
	"github.com/kptdev/krm-functions-catalog/functions/go/generate-kpt-pkg-docs/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const defaultReadmePath = "/tmp/README.md"
const defaultRepoPath = "https://github.com/GoogleCloudPlatform/blueprints.git/catalog/"

// nolint
func main() {
	rp := ReadmeProcessor{}
	cmd := command.Build(&rp, command.StandaloneEnabled, false)

	cmd.Short = generated.GenerateKptPkgDocsShort
	cmd.Long = generated.GenerateKptPkgDocsLong
	cmd.Example = generated.GenerateKptPkgDocsExamples
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type ReadmeProcessor struct{}

func (rp *ReadmeProcessor) Process(resourceList *framework.ResourceList) error {
	readmePath, repoPath, pkgName := parseFnCfg(resourceList.FunctionConfig)

	currentDoc, err := os.ReadFile(readmePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			resourceList.Results = getResults(fmt.Sprintf("Skipping readme generation: %s", err), framework.Warning)
			return nil
		} else {
			resourceList.Results = getResults(err.Error(), framework.Error)
		}
	}
	err = generateReadme(repoPath, readmePath, pkgName, string(currentDoc), resourceList)
	if err != nil {
		resourceList.Results = getResults(err.Error(), framework.Error)
		return err
	}
	return nil
}

func generateReadme(repoPath, readmePath, pkgName, currentDoc string, resourceList *framework.ResourceList) error {
	title, generatedDoc, err := docs.GenerateBlueprintReadme(resourceList.Items, repoPath, pkgName)
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
func getResults(msg string, severity framework.Severity) []*framework.Result {
	return []*framework.Result{
		{
			Message:  fmt.Sprintf("failed to generate doc: %s", msg),
			Severity: severity,
		},
	}
}
