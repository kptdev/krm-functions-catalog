package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/kptdev/krm-functions-catalog/archived/functions/go/project-services/gcpservices"
	"github.com/kptdev/krm-functions-catalog/archived/functions/go/project-services/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

// nolint
func main() {
	psf := ProjectServiceSetFunction{}
	cmd := command.Build(&psf, command.StandaloneEnabled, false)

	cmd.Short = generated.EnableGcpServicesShort
	cmd.Long = generated.EnableGcpServicesLong
	cmd.Example = generated.EnableGcpServicesExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type ProjectServiceSetFunction struct{}

func (psf *ProjectServiceSetFunction) Process(resourceList *framework.ResourceList) error {
	var pslr gcpservices.ProjectServiceSetRunner
	var err error
	resourceList.Items, err = pslr.Filter(resourceList.Items)
	if err != nil {
		resourceList.Results = getErrorItem(err.Error())
		return err
	}
	resourceList.Results = pslr.GetResults()
	sortResultItems(resourceList.Results)
	return nil
}

// getErrorItem returns the item for an error message
func getErrorItem(errMsg string) []*framework.Result {
	return []*framework.Result{
		{
			Message:  fmt.Sprintf("failed to add services: %s", errMsg),
			Severity: framework.Error,
		},
	}
}

// from https://github.com/kptdev/kpt/issues/2195
// refactor once upstreamed
func sortResultItems(items []*framework.Result) {
	sort.SliceStable(items, func(i, j int) bool {
		if fileLess(items, i, j) != 0 {
			return fileLess(items, i, j) < 0
		}
		return resultItemToString(items[i]) < resultItemToString(items[j])
	})
}

func fileLess(items []*framework.Result, i, j int) int {
	if items[i].File.Path != items[j].File.Path {
		if items[i].File.Path < items[j].File.Path {
			return -1
		} else {
			return 1
		}
	}
	return items[i].File.Index - items[j].File.Index
}

func resultItemToString(item *framework.Result) string {
	return fmt.Sprintf("resource-ref:%s,field:%s,message:%s",
		item.ResourceRef, item.Field, item.Message)
}
