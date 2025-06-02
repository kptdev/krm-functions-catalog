package main

import (
	"fmt"
	"os"

	"github.com/kptdev/krm-functions-catalog/functions/go/list-setters/generated"
	"github.com/kptdev/krm-functions-catalog/functions/go/list-setters/listsetters"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

// nolint
func main() {
	lsp := ListSettersProcessor{}
	cmd := command.Build(&lsp, command.StandaloneEnabled, false)

	cmd.Short = generated.ListSettersShort
	cmd.Long = generated.ListSettersLong
	cmd.Example = generated.ListSettersExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type ListSettersProcessor struct{}

func (lsp *ListSettersProcessor) Process(resourceList *framework.ResourceList) error {
	resourceList.Result = &framework.Result{
		Name: "list-setters",
	}
	items, err := run(resourceList)
	if err != nil {
		resourceList.Result.Items = getErrorItem(fmt.Sprintf("failed to list setters: %s", err.Error()), framework.Error)
		return err
	}
	resourceList.Result.Items = items
	return nil
}

func run(resourceList *framework.ResourceList) ([]framework.ResultItem, error) {
	ls := listsetters.New()
	_, err := ls.Filter(resourceList.Items)
	if err != nil {
		return nil, err
	}
	resultItems, err := resultsToItems(ls)
	if err != nil {
		return nil, err
	}
	return resultItems, nil
}

// resultsToItems converts the listsetters results to
// equivalent items([]framework.Item)
func resultsToItems(sr listsetters.ListSetters) ([]framework.ResultItem, error) {
	var items []framework.ResultItem
	rs := sr.GetResults()
	if len(rs) == 0 {
		return getErrorItem("no setters found", framework.Warning), nil
	}
	for _, r := range rs {
		items = append(items, framework.ResultItem{
			Message: r.String(),
		})
	}
	return items, nil
}

// getErrorItem returns the item for an error message
func getErrorItem(errMsg string, severity framework.Severity) []framework.ResultItem {
	return []framework.ResultItem{
		{
			Message:  errMsg,
			Severity: severity,
		},
	}
}
