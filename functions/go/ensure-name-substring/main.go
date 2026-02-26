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

	"github.com/kptdev/krm-functions-catalog/functions/go/ensure-name-substring/generated"
	"github.com/kptdev/krm-functions-catalog/functions/go/ensure-name-substring/nameref"
	"sigs.k8s.io/kustomize/api/hasher"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/yaml"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	tc, err := getDefaultConfig()
	if err != nil {
		return err
	}

	ensp := EnsureNameSubstringProcessor{
		tc: &tc,
	}
	cmd := command.Build(&ensp, command.StandaloneEnabled, false)

	cmd.Short = generated.EnsureNameSubstringShort
	cmd.Long = generated.EnsureNameSubstringLong
	return cmd.Execute()
}

type EnsureNameSubstringProcessor struct {
	tc *transformerConfig
}

func (ensp *EnsureNameSubstringProcessor) Process(resourceList *framework.ResourceList) error {
	var ens EnsureNameSubstring
	if err := framework.LoadFunctionConfig(resourceList.FunctionConfig, &ens); err != nil {
		return fmt.Errorf("failed to load the `functionConfig`: %w", err)
	}

	if ensp.tc == nil {
		return fmt.Errorf("failed to load the default configuration")
	}

	ens.AdditionalNameFields = append(ensp.tc.FieldSpecs, ens.AdditionalNameFields...)

	resmapFactory := newResMapFactory()

	resMap, err := resmapFactory.NewResMapFromRNodeSlice(resourceList.Items)
	if err != nil {
		return fmt.Errorf("failed to convert items to resource map: %w", err)
	}

	if err = ens.Transform(resMap); err != nil {
		return fmt.Errorf("failed to transform name substring: %w", err)
	}
	// update name back reference
	err = nameref.FixNameBackReference(resMap)
	if err != nil {
		return fmt.Errorf("failed to fix name back reference: %w", err)
	}

	// remove kustomize build annotations
	resMap.RemoveBuildAnnotations()
	resourceList.Items = resMap.ToRNodeSlice()
	if err != nil {
		return fmt.Errorf("failed to convert resource map to items: %w", err)
	}
	return nil
}

func newResMapFactory() *resmap.Factory {
	resourceFactory := resource.NewFactory(&hasher.Hasher{})
	resourceFactory.IncludeLocalConfigs = true
	return resmap.NewFactory(resourceFactory)
}

type transformerConfig struct {
	FieldSpecs []types.FieldSpec `json:"namePrefix,omitempty" yaml:"namePrefix,omitempty"`
}

func getDefaultConfig() (transformerConfig, error) {
	defaultConfigString := builtinpluginconsts.GetDefaultFieldSpecsAsMap()["nameprefix"]
	var tc transformerConfig
	err := yaml.Unmarshal([]byte(defaultConfigString), &tc)
	return tc, err
}
