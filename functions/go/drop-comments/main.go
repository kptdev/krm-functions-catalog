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
	"fmt"
	"os"

	"github.com/kptdev/krm-functions-catalog/functions/go/drop-comments/generated"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func main() {
	cmd := command.Build(&DropCommentsProcessor{}, command.StandaloneEnabled, false)
	cmd.Short = generated.DropCommentsShort
	cmd.Long = generated.DropCommentsLong
	cmd.Example = generated.DropCommentsExamples

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type DropCommentsProcessor struct{}

func (dcp *DropCommentsProcessor) Process(resourceList *framework.ResourceList) error {
	for i := range resourceList.Items {
		jsonItem, err := resourceList.Items[i].MarshalJSON()
		if err != nil {
			return err
		}
		node, err := yaml.ConvertJSONToYamlNode(string(jsonItem))
		if err != nil {
			return err
		}
		resourceList.Items[i] = node
	}
	return nil
}
