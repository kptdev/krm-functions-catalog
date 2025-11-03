package main

import (
	"fmt"
	"os"

	"github.com/kptdev/krm-functions-catalog/archived/functions/go/format/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
)

// nolint
func main() {
	asp := FormatProcessor{}
	cmd := command.Build(&asp, command.StandaloneEnabled, false)

	cmd.Short = generated.FormatShort
	cmd.Long = generated.FormatLong
	cmd.Example = generated.FormatExamples

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (fp *FormatProcessor) Process(resourceList *framework.ResourceList) error {
	f := filters.FormatFilter{
		UseSchema: true,
	}
	_, err := f.Filter(resourceList.Items)
	if err != nil {
		return fmt.Errorf("failed to format resources: %w", err)
	}
	return nil
}

type FormatProcessor struct{}
