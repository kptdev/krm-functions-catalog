package fnsdk_test

import (
	"os"

	"github.com/kptdev/krm-functions-catalog/archived/functions/go/export-terraform/thirdparty/kyaml/fnsdk"
)

// This example implements a function that updates the replicas field for all deployments.

func Example_filterGVK() {
	if err := fnsdk.AsMain(fnsdk.ResourceListProcessorFunc(updateReplicas)); err != nil {
		os.Exit(1)
	}
}

// updateReplicas sets a field in resources selecting by GVK.
func updateReplicas(rl *fnsdk.ResourceList) error {
	if rl.FunctionConfig == nil {
		return fnsdk.ErrMissingFnConfig{}
	}
	var replicas int
	rl.FunctionConfig.GetOrDie(&replicas, "replicas")
	for i, obj := range rl.Items {
		if obj.APIVersion() == "apps/v1" && obj.Kind() == "Deployment" {
			rl.Items[i].SetOrDie(replicas, "spec", "replicas")
		}
	}
	return nil
}
