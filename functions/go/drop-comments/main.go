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
	"os"

	"github.com/kptdev/krm-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

//go:embed README.md
var readme []byte

//go:embed metadata.yaml
var metadata []byte

func main() {
	if err := fn.AsMain(fn.ResourceListProcessorFunc(processDropComments), fn.WithDocs(readme, metadata)); err != nil {
		os.Exit(1)
	}
}

func processDropComments(rl *fn.ResourceList) (bool, error) {
	for i, item := range rl.Items {
		jsonBytes, err := item.CopyToResourceNode().MarshalJSON()
		if err != nil {
			return false, err
		}
		node, err := yaml.ConvertJSONToYamlNode(string(jsonBytes))
		if err != nil {
			return false, err
		}
		rl.Items[i] = fn.MoveToKubeObject(node)
	}
	return true, nil
}
