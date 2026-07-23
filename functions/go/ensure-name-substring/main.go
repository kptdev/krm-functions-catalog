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
	"slices"

	"github.com/kptdev/krm-functions-catalog/functions/go/ensure-name-substring/generated"
	nameref "github.com/kptdev/krm-functions-catalog/functions/go/ensure-name-substring/third_party/sigs.k8s.io/kustomize/api/accumulator"
	consts "github.com/kptdev/krm-functions-catalog/functions/go/ensure-name-substring/third_party/sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/hasher"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
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

	// Remove kustomize tracking annotations (previousNames, prefixes, etc.)
	// without dropping kpt path/index annotations. RemoveBuildAnnotations()
	// also clears those, which would rewrite files such as Kptfile to
	// kind_name.yaml when the package is written back.
	if err = removeKustomizeTrackingAnnotations(resMap); err != nil {
		return fmt.Errorf("failed to remove kustomize tracking annotations: %w", err)
	}
	resourceList.Items = resMap.ToRNodeSlice()
	return nil
}

var kustomizeTrackingAnnos = func() []string {
	// kptPackageAnnotations must be preserved so kpt can write resources back to
	// their original package paths (especially the literal "Kptfile" filename).
	kptPackageAnnotations := []string{
		kioutil.PathAnnotation,
		kioutil.IndexAnnotation,
		kioutil.SeqIndentAnnotation,
		kioutil.IdAnnotation,
		kioutil.InternalAnnotationsMigrationResourceIDAnnotation,
		kioutil.LegacyPathAnnotation,  //nolint:staticcheck
		kioutil.LegacyIndexAnnotation, //nolint:staticcheck
		kioutil.LegacyIdAnnotation,    //nolint:staticcheck
	}

	return slices.DeleteFunc(slices.Clone(resource.BuildAnnotations), func(s string) bool {
		return slices.Contains(kptPackageAnnotations, s)
	})
}()

func removeKustomizeTrackingAnnotations(m resmap.ResMap) error {
	for _, r := range m.Resources() {
		annotations := r.GetAnnotations()
		if len(annotations) == 0 {
			continue
		}
		for _, a := range kustomizeTrackingAnnos {
			delete(annotations, a)
		}
		if err := r.SetAnnotations(annotations); err != nil {
			return err
		}
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
	defaultConfigString := consts.NamePrefixFieldSpecs
	var tc transformerConfig
	err := yaml.Unmarshal([]byte(defaultConfigString), &tc)
	return tc, err
}
