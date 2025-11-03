package main

import (
	"fmt"
	"os"

	"github.com/kptdev/krm-functions-catalog/archived/functions/go/fix/fixpkg"
	"github.com/kptdev/krm-functions-catalog/archived/functions/go/fix/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

// nolint
func main() {
	fp := FixProcessor{}
	cmd := command.Build(&fp, command.StandaloneEnabled, false)

	cmd.Short = generated.FixShort
	cmd.Long = generated.FixLong
	cmd.Example = generated.FixExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type FixProcessor struct{}

func (fp *FixProcessor) Process(resourceList *framework.ResourceList) error {
	s := &fixpkg.Fix{}
	var err error
	resourceList.Result = &framework.Result{
		Name: "fix",
	}
	resourceList.Items, err = s.Filter(resourceList.Items)
	if err != nil {
		resourceList.Result.Items = getErrorItem(err.Error())
		return err
	}
	resourceList.Result.Items = resultsToItems(s)
	return nil
}

// resultsToItems converts the Search and Replace results to
// equivalent items([]framework.Item)
func resultsToItems(sr *fixpkg.Fix) []framework.ResultItem {
	var items []framework.ResultItem
	for _, res := range sr.Results {
		items = append(items, framework.ResultItem{
			Message: res.Message,
			File:    framework.File{Path: res.FilePath},
		})
	}
	return items
}

// getErrorItem returns the item for input error message
func getErrorItem(errMsg string) []framework.ResultItem {
	return []framework.ResultItem{
		{
			Message:  fmt.Sprintf("failed to fix package: %s", errMsg),
			Severity: framework.Error,
		},
	}
}
