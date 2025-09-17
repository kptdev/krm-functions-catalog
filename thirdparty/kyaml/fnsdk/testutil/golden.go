package testutil

import (
	"fmt"
	"io/ioutil"

	"github.com/kptdev/krm-functions-catalog/thirdparty/kyaml/fnsdk"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

// ResourceListFromFile reads a yaml file and converts it to a fnsdk.ResourceList.
func ResourceListFromFile(path string) (*fnsdk.ResourceList, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return fnsdk.ParseResourceList(content)
}

// ResourceListFromDirectory reads yaml files from dir and functionConfig file,
// and then assemble them as a fnsdk.ResourceList.
func ResourceListFromDirectory(dir string, fnConfigFile string) (*fnsdk.ResourceList, error) {
	reader := &kio.LocalPackageReader{
		PackagePath:        dir,
		IncludeSubpackages: true,
	}
	items, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("unable to read resources from %v: %w", dir, err)
	}

	rl := &fnsdk.ResourceList{
		Items: fnsdk.NewFromRNodes(items),
	}

	if fnConfigFile != "" {
		content, err := ioutil.ReadFile(fnConfigFile)
		if err != nil {
			return nil, err
		}
		fnConfig, err := fnsdk.ParseKubeObject(content)
		if err != nil {
			return nil, fmt.Errorf("unable to parse the functionConfig object: %w", err)
		}
		rl.FunctionConfig = fnConfig
	}
	return rl, nil
}

// ResourceListToDirectory write ResourceList.items to yaml files according to
// the path annotation (https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md#internalconfigkubernetesiopath).
func ResourceListToDirectory(rl *fnsdk.ResourceList, dir string) error {
	writer := &kio.LocalPackageWriter{
		PackagePath: dir,
	}
	err := writer.Write(fnsdk.ToRNodes(rl.Items))
	if err != nil {
		return fmt.Errorf("unable to write resources to %v: %w", dir, err)
	}
	return nil
}
