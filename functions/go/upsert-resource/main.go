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
	"fmt"
	"os"

	"github.com/kptdev/kpt/api/kubeobject"
	"github.com/kptdev/krm-functions-catalog/functions/go/upsert-resource/upsertresource"
	"github.com/kptdev/krm-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

//go:embed README.md
var readme []byte

//go:embed metadata.yaml
var metadata []byte

// nolint
func main() {
	if err := fn.AsMain(kubeobject.ResourceListProcessorFunc(Process), fn.WithDocs(readme, metadata)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Process bridges the kubeobject ResourceList to the kyaml-based upsertresource
// filter, which still operates on *yaml.RNode. Items and the functionConfig are
// converted to RNodes, filtered, then moved back into KubeObjects.
func Process(resourceList *kubeobject.ResourceList) (bool, error) {
	var fnConfig *yaml.RNode
	if resourceList.FunctionConfig != nil {
		fnConfig = resourceList.FunctionConfig.CopyToResourceNode()
	}

	ur := &upsertresource.UpsertResource{
		List: fnConfig,
	}

	items := resourceList.Items.CopyToResourceNodes()
	filtered, err := ur.Filter(items)
	if err != nil {
		return false, fmt.Errorf("failed to upsert resource: %w", err)
	}

	resourceList.Items = kubeobject.MoveToKubeObjects(filtered)
	return true, nil
}
