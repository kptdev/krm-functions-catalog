package main

import (
	"fmt"
	"os"

	"github.com/kptdev/krm-functions-catalog/functions/go/apply-setters/applysetters"
	"github.com/kptdev/krm-functions-catalog/functions/go/apply-setters/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

// nolint
func main() {
	asp := ApplySettersProcessor{}
	cmd := command.Build(&asp, command.StandaloneEnabled, false)

	cmd.Short = generated.ApplySettersShort
	cmd.Long = generated.ApplySettersLong
	cmd.Example = generated.ApplySettersExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type ApplySettersProcessor struct{}

func (asp *ApplySettersProcessor) Process(resourceList *framework.ResourceList) error {
	resourceList.Result = &framework.Result{
		Name: "apply-setters",
	}
	items, err := run(resourceList)
	if err != nil {
		resourceList.Result.Items = getErrorItem(err.Error())
		return err
	}
	resourceList.Result.Items = items
	return nil
}

func run(resourceList *framework.ResourceList) ([]framework.ResultItem, error) {
	s, err := getSetters(resourceList.FunctionConfig)
	if err != nil {
		return nil, err
	}
	_, err = s.Filter(resourceList.Items)
	if err != nil {
		return nil, err
	}
	resultItems, err := resultsToItems(s)
	if err != nil {
		return nil, err
	}
	return resultItems, nil
}

// getSetters retrieve the setters from input config
func getSetters(fc *kyaml.RNode) (applysetters.ApplySetters, error) {
	var fcd applysetters.ApplySetters
	applysetters.Decode(fc, &fcd)
	return fcd, nil
}

// resultsToItems converts the Search and Replace results to
// equivalent items([]framework.Item)
func resultsToItems(sr applysetters.ApplySetters) ([]framework.ResultItem, error) {
	var items []framework.ResultItem
	if len(sr.Results) == 0 {
		items = append(items, framework.ResultItem{
			Message: "no matches for input setter(s)",
		})
		return items, nil
	}
	for _, res := range sr.Results {
		items = append(items, framework.ResultItem{
			Message: fmt.Sprintf("set field value to %q", res.Value),
			Field:   framework.Field{Path: res.FieldPath},
			File:    framework.File{Path: res.FilePath},
		})
	}
	return items, nil
}

// getErrorItem returns the item for input error message
func getErrorItem(errMsg string) []framework.ResultItem {
	return []framework.ResultItem{
		{
			Message:  fmt.Sprintf("failed to apply setters: %s", errMsg),
			Severity: framework.Error,
		},
	}
}
