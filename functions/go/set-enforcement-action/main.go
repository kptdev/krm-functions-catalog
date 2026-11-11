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
	"github.com/kptdev/krm-functions-sdk/go/fn"
)

//go:embed README.md
var readme []byte

//go:embed metadata.yaml
var metadata []byte

const (
	policyAPIVersion     = "constraints.gatekeeper.sh/v1beta1"
	enforcementActionKey = "enforcementAction"
	denyActionValue      = "deny"
	warnActionValue      = "warn"
	dryRunActionValue    = "dryrun"
)

// nolint
func main() {
	if err := fn.AsMain(kubeobject.ResourceListProcessorFunc(Process), fn.WithDocs(readme, metadata)); err != nil {
		os.Exit(1)
	}
}

func Process(resourceList *kubeobject.ResourceList) (bool, error) {
	// get the enforcementAction value from functionConfig
	var acn string
	err := getEnforcementAction(resourceList.FunctionConfig, &acn)
	if err != nil {
		resourceList.Results = getErrorItem(err.Error())
		return false, err
	}

	// process policies in the package and display results
	items, err := processPolicies(resourceList.Items, acn)
	if err != nil {
		resourceList.Results = getErrorItem(err.Error())
		return false, err
	}
	resourceList.Results = items
	return true, nil
}

func processPolicies(resourceList []*kubeobject.KubeObject, acn string) (kubeobject.Results, error) {
	var resultItems kubeobject.Results
	for _, node := range resourceList {
		if node == nil || node.IsEmpty() {
			continue
		}

		//check for resources of type policy
		if node.GetName() == "" || node.GetKind() == "" || node.GetAPIVersion() != policyAPIVersion {
			continue
		}

		if err := node.SetNestedField(acn, "spec", "enforcementAction"); err != nil {
			return nil, err
		}

		itemFilePath := node.GetAnnotations()["internal.config.kubernetes.io/path"]
		if itemFilePath == "" {
			itemFilePath = node.GetAnnotations()["config.kubernetes.io/path"]
		}

		resultItems = append(resultItems, &kubeobject.Result{
			Message: fmt.Sprintf("Policy name: [%s]", node.GetName()),
			File: &kubeobject.File{
				Path: itemFilePath,
			},
			Severity: kubeobject.Info,
		})
	}

	if len(resultItems) > 0 {
		infoResultSlice := kubeobject.Results{
			{
				Severity: kubeobject.Info,
				Message:  fmt.Sprintf("Number of policies set to [%s]: %d", acn, len(resultItems)),
			},
		}

		resultItems = append(infoResultSlice, resultItems...)
	} else if len(resultItems) == 0 {
		resultItems = append(resultItems, &kubeobject.Result{
			Message:  fmt.Sprintf("Found no policy to set to [%s]", acn),
			Severity: kubeobject.Warning,
		})
	}

	return resultItems, nil
}

// getEnforcementAction gets the value to set for enforcementAction from the functionConfig
func getEnforcementAction(fc *kubeobject.KubeObject, acn *string) error {
	data, _, _ := fc.NestedStringMap("data")
	if len(data) != 1 {
		return errors.New("expecting exactly 1 enforcementAction as part of the ConfigMap")
	}

	*acn = data[enforcementActionKey]
	if *acn != denyActionValue && *acn != warnActionValue && *acn != dryRunActionValue {
		return fmt.Errorf("expected values for enforcementAction are [%s] or [%s] or [%s]", denyActionValue, warnActionValue, dryRunActionValue)
	}

	return nil
}

// getErrorItem returns the item for an error message
func getErrorItem(errMsg string) kubeobject.Results {
	return kubeobject.Results{
		{
			Message:  fmt.Sprintf("failed to process policies: %s", errMsg),
			Severity: kubeobject.Error,
		},
	}
}
