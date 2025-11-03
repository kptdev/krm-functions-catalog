package fnsdk_test

import (
	"os"

	"github.com/kptdev/krm-functions-catalog/archived/functions/go/export-terraform/thirdparty/kyaml/fnsdk"
)

// This example implements a function that validate resources to ensure
// spec.template.spec.securityContext.runAsNonRoot is set in workload APIs.

func Example_validator() {
	if err := fnsdk.AsMain(fnsdk.ResourceListProcessorFunc(validator)); err != nil {
		os.Exit(1)
	}
}

func validator(rl *fnsdk.ResourceList) error {
	var results fnsdk.Results
	for _, obj := range rl.Items {
		if obj.APIVersion() == "apps/v1" && (obj.Kind() == "Deployment" || obj.Kind() == "StatefulSet" || obj.Kind() == "DaemonSet" || obj.Kind() == "ReplicaSet") {
			var runAsNonRoot bool
			obj.GetOrDie(&runAsNonRoot, "spec", "template", "spec", "securityContext", "runAsNonRoot")
			if !runAsNonRoot {
				results = append(results, fnsdk.ConfigObjectResult("`spec.template.spec.securityContext.runAsNonRoot` must be set to true", obj, fnsdk.Error))
			}
		}
	}
	return results
}
