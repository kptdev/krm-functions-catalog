package fnsdk_test

import (
	"os"

	"github.com/kptdev/krm-functions-catalog/thirdparty/kyaml/fnsdk"
)

// In this example, we read a field from the input object and print it to the log.

func Example_aReadField() {
	if err := fnsdk.AsMain(fnsdk.ResourceListProcessorFunc(readField)); err != nil {
		os.Exit(1)
	}
}

func readField(rl *fnsdk.ResourceList) error {
	for _, obj := range rl.Items {
		if obj.APIVersion() == "apps/v1" && obj.Kind() == "Deployment" {
			var replicas int
			obj.GetOrDie(&replicas, "spec", "replicas")
			fnsdk.Logf("replicas is %v\n", replicas)
		}
	}
	return nil
}
