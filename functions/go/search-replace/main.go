package main

import (
	"fmt"
	"os"

	"github.com/kptdev/krm-functions-catalog/functions/go/search-replace/generated"
	"github.com/kptdev/krm-functions-catalog/functions/go/search-replace/searchreplace"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// nolint
func main() {
	srp := SearchReplaceProcessor{}
	cmd := command.Build(&srp, command.StandaloneEnabled, false)

	cmd.Short = generated.SearchReplaceShort
	cmd.Long = generated.SearchReplaceLong
	cmd.Example = generated.SearchReplaceExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type SearchReplaceProcessor struct{}

func (srp *SearchReplaceProcessor) Process(resourceList *framework.ResourceList) error {
	resourceList.Result = &framework.Result{
		Name: "search-replace",
	}
	items, err := run(resourceList)
	if err != nil {
		resourceList.Result.Items = getErrorItem(err.Error())
		return err
	}
	resourceList.Result.Items = items
	return nil
}

// run resolves the function params from input ResourceList and runs the function on resources
func run(resourceList *framework.ResourceList) ([]framework.ResultItem, error) {
	sr, err := getSearchReplaceParams(resourceList.FunctionConfig)
	if err != nil {
		return nil, err
	}

	_, err = sr.Filter(resourceList.Items)
	if err != nil {
		return nil, err
	}

	return searchResultsToItems(sr), nil
}

// getSearchReplaceParams retrieve the search parameters from input config
func getSearchReplaceParams(fc *kyaml.RNode) (searchreplace.SearchReplace, error) {
	var fcd searchreplace.SearchReplace
	if err := searchreplace.Decode(fc, &fcd); err != nil {
		return fcd, err
	}
	return fcd, nil
}

// searchResultsToItems converts the Search and Replace results to
// equivalent items([]framework.Item)
func searchResultsToItems(sr searchreplace.SearchReplace) []framework.ResultItem {
	var items []framework.ResultItem
	if len(sr.Results) == 0 {
		items = append(items, framework.ResultItem{
			Message: "no matches",
		})
		return items
	}
	for _, res := range sr.Results {
		var message string
		if sr.PutComment != "" || sr.PutValue != "" {
			message = fmt.Sprintf("Mutated field value to %q", res.Value)
		} else {
			message = fmt.Sprintf("Matched field value %q", res.Value)
		}

		items = append(items, framework.ResultItem{
			Message: message,
			Field:   framework.Field{Path: res.FieldPath},
			File:    framework.File{Path: res.FilePath},
		})
	}
	return items
}

// getErrorItem returns the item for input error message
func getErrorItem(errMsg string) []framework.ResultItem {
	return []framework.ResultItem{
		{
			Message:  fmt.Sprintf("failed to perform search-replace operation: %q", errMsg),
			Severity: framework.Error,
		},
	}
}
