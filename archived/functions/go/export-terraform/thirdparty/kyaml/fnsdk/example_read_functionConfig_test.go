package fnsdk_test

import (
	"os"

	"github.com/kptdev/krm-functions-catalog/archived/functions/go/export-terraform/thirdparty/kyaml/fnsdk"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// In this example, we convert the functionConfig as strong typed object and then
// read a field from the functionConfig object.

func Example_bReadFunctionConfig() {
	if err := fnsdk.AsMain(fnsdk.ResourceListProcessorFunc(readFunctionConfig)); err != nil {
		os.Exit(1)
	}
}

func readFunctionConfig(rl *fnsdk.ResourceList) error {
	var sr SetReplicas
	rl.FunctionConfig.AsOrDie(&sr)
	fnsdk.Logf("desired replicas is %v\n", sr.DesiredReplicas)
	return nil
}

// SetReplicas is the type definition of the functionConfig
type SetReplicas struct {
	yaml.ResourceIdentifier `json:",inline" yaml:",inline"`
	DesiredReplicas         int `json:"desiredReplicas,omitempty" yaml:"desiredReplicas,omitempty"`
}
