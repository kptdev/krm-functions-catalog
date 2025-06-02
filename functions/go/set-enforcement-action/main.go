package main

import (
	"errors"
	"fmt"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	policyAPIVersion     = "constraints.gatekeeper.sh/v1beta1"
	enforcementActionKey = "enforcementAction"
	denyActionValue      = "deny"
	warnActionValue      = "warn"
	dryRunActionValue    = "dryrun"
)

// nolint
func main() {
	fp := SetEnforcementActionProcessor{}
	cmd := command.Build(&fp, command.StandaloneEnabled, false)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
func (fp *SetEnforcementActionProcessor) Process(resourceList *framework.ResourceList) error {
	resourceList.Result = &framework.Result{
		Name: "set-enforcement-action",
	}

	// get the enforcementAction value from functionConfig
	var acn string
	err := getEnforcementAction(resourceList.FunctionConfig, &acn)
	if err != nil {
		resourceList.Result.Items = getErrorItem(err.Error())
		return err
	}

	// process policies in the package and display results
	items, err := processPolicies(resourceList.Items, acn)
	if err != nil {
		resourceList.Result.Items = getErrorItem(err.Error())
		return err
	}
	resourceList.Result.Items = items
	return nil
}

func processPolicies(resourceList []*yaml.RNode, acn string) ([]framework.ResultItem, error) {
	var resultItems []framework.ResultItem
	for _, node := range resourceList {
		if node.IsNilOrEmpty() {
			continue
		}

		//check for resources of type policy
		metadata, err := node.GetMeta()
		if err != nil {
			return nil, err
		}

		if metadata.Name == "" || metadata.Kind == "" || metadata.APIVersion != policyAPIVersion {
			continue
		}

		err = node.PipeE(
			yaml.LookupCreate(yaml.ScalarNode, "spec"),
			yaml.SetField("enforcementAction", yaml.NewScalarRNode(acn)))
		if err != nil {
			return nil, err
		}

		itemFilePath := node.GetAnnotations()["internal.config.kubernetes.io/path"]
		if itemFilePath == "" {
			itemFilePath = node.GetAnnotations()["config.kubernetes.io/path"]
		}

		resultItems = append(resultItems, framework.ResultItem{
			Message: fmt.Sprintf("Policy name: [%s]", node.GetName()),
			File: framework.File{
				Path: itemFilePath,
			},
			Severity: framework.Info,
		})
	}

	if len(resultItems) > 0 {
		infoResultSlice := []framework.ResultItem{}
		infoResultSlice = append(infoResultSlice, framework.ResultItem{
			Severity: framework.Info,
			Message:  fmt.Sprintf("Number of policies set to [%s]: %d", acn, len(resultItems)),
		})

		resultItems = append(infoResultSlice, resultItems...)
	} else if len(resultItems) == 0 {
		resultItems = append(resultItems, framework.ResultItem{
			Message:  fmt.Sprintf("Found no policy to set to [%s]", acn),
			Severity: framework.Warning,
		})
	}

	return resultItems, nil
}

// getEnforcementAction gets the value to set for enforcementAction from the functionConfig
func getEnforcementAction(fc *yaml.RNode, acn *string) error {
	if len(fc.GetDataMap()) != 1 {
		return errors.New("expecting exactly 1 enforcementAction as part of the ConfigMap")
	}

	*acn = fc.GetDataMap()[enforcementActionKey]
	if *acn != denyActionValue && *acn != warnActionValue && *acn != dryRunActionValue {
		return fmt.Errorf("expected values for enforcementAction are [%s] or [%s] or [%s]", denyActionValue, warnActionValue, dryRunActionValue)
	}

	return nil
}

// getErrorItem returns the item for an error message
func getErrorItem(errMsg string) []framework.ResultItem {
	return []framework.ResultItem{
		{
			Message:  fmt.Sprintf("failed to process policies: %s", errMsg),
			Severity: framework.Error,
		},
	}
}

type SetEnforcementActionProcessor struct{}
