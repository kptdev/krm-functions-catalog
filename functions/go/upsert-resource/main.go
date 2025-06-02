package main

import (
	"fmt"
	"os"

	"github.com/kptdev/krm-functions-catalog/functions/go/upsert-resource/generated"
	"github.com/kptdev/krm-functions-catalog/functions/go/upsert-resource/upsertresource"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

// nolint
func main() {
	asp := UpsertResourceProcessor{}
	cmd := command.Build(&asp, command.StandaloneEnabled, false)

	cmd.Short = generated.UpsertResourceShort
	cmd.Long = generated.UpsertResourceLong
	cmd.Example = generated.UpsertResourceExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type UpsertResourceProcessor struct{}

func (urp *UpsertResourceProcessor) Process(resourceList *framework.ResourceList) error {
	ur := &upsertresource.UpsertResource{
		List: resourceList.FunctionConfig,
	}
	var err error
	resourceList.Items, err = ur.Filter(resourceList.Items)
	if err != nil {
		return fmt.Errorf("failed to upsert resource: %w", err)
	}
	return nil
}
