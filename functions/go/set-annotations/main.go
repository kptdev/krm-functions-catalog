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
// This file will be processed and embedded to pluginator.

package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/kptdev/kpt/api/kubeobject"
	consts "github.com/kptdev/krm-functions-catalog/functions/go/set-annotations/third_party/sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"github.com/kptdev/krm-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/api/hasher"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"
)

const (
	fnConfigGroup      = "fn.kpt.dev"
	fnConfigVersion    = "v1alpha1"
	fnConfigAPIVersion = fnConfigGroup + "/" + fnConfigVersion
	legacyFnConfigKind = "SetAnnotationConfig"
	fnConfigKind       = "SetAnnotations"
)

//go:embed README.md
var readme []byte

//go:embed metadata.yaml
var metadata []byte

func Process(resourceList *kubeobject.ResourceList) (bool, error) {
	results, err := run(resourceList)
	if err != nil {
		resourceList.Results = kubeobject.Results{
			&kubeobject.Result{
				Message:  err.Error(),
				Severity: kubeobject.Error,
			},
		}
		return false, resourceList.Results
	}
	resourceList.Results = results
	return true, nil
}

type transformerConfig struct {
	FieldSpecs types.FsSlice `json:"commonAnnotations,omitempty" yaml:"commonAnnotations,omitempty"`
}

type setAnnotationFunction struct {
	kyaml.ResourceMeta `json:",inline" yaml:",inline"`
	plugin             `json:",inline" yaml:",inline"`
}

func (f *setAnnotationFunction) Config(rn *kyaml.RNode) error {
	switch {
	case f.validGVK(rn, "v1", "ConfigMap"):
		f.Annotations = rn.GetDataMap()
	case f.validGVK(rn, fnConfigAPIVersion, legacyFnConfigKind):
		fallthrough
	case f.validGVK(rn, fnConfigAPIVersion, fnConfigKind):
		// input config is a CRD
		y, err := rn.String()
		if err != nil {
			return fmt.Errorf("cannot get YAML from RNode: %w", err)
		}
		err = f.plugin.Config(nil, []byte(y))
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("`functionConfig` must be a `ConfigMap` or `%s`", fnConfigKind)
	}

	if len(f.Annotations) == 0 {
		return fmt.Errorf("input annotation list cannot be empty")
	}
	tc, err := getDefaultConfig()
	if err != nil {
		return err
	}
	// append default field specs
	f.AdditionalAnnotationFields = append(f.AdditionalAnnotationFields, tc.FieldSpecs...)
	return nil
}

func (f *setAnnotationFunction) Run(items []*kyaml.RNode) ([]*kyaml.RNode, error) {
	resmapFactory := newResMapFactory()
	resMap, err := resmapFactory.NewResMapFromRNodeSlice(items)
	if err != nil {
		return nil, err
	}
	err = f.Transform(resMap)
	if err != nil {
		return nil, fmt.Errorf("failed to run transformer: %w", err)
	}
	return resMap.ToRNodeSlice(), nil
}

func (f *setAnnotationFunction) validGVK(rn *kyaml.RNode, apiVersion, kind string) bool {
	meta, err := rn.GetMeta()
	if err != nil {
		return false
	}
	if meta.APIVersion != apiVersion || meta.Kind != kind {
		return false
	}
	return true
}

// resultsToItems converts the set annotation results to
// equivalent kubeobject.Results
func (f *setAnnotationFunction) resultsToItems() (kubeobject.Results, error) {
	var results kubeobject.Results
	if len(f.Results) == 0 {
		results = append(results, &kubeobject.Result{
			Message: "no annotations applied",
		})
		return results, nil
	}
	for resKey, annoVals := range f.Results {
		fileIndex, _ := strconv.Atoi(resKey.FileIndex)
		annotationJson, err := json.Marshal(annoVals)
		if err != nil {
			return nil, err
		}
		results = append(results, &kubeobject.Result{
			Message: fmt.Sprintf("set annotations: %s", annotationJson),
			Field:   &kubeobject.Field{Path: resKey.FieldPath},
			File:    &kubeobject.File{Path: resKey.FilePath, Index: fileIndex},
		})
	}
	results.Sort()
	return results, nil
}

func getDefaultConfig() (transformerConfig, error) {
	var tc transformerConfig
	err := yaml.Unmarshal([]byte(consts.CommonAnnotationFieldSpecs), &tc)
	return tc, err
}

func newResMapFactory() *resmap.Factory {
	resourceFactory := resource.NewFactory(&hasher.Hasher{})
	resourceFactory.IncludeLocalConfigs = true
	return resmap.NewFactory(resourceFactory)
}

// nolint
func main() {
	if err := fn.AsMain(kubeobject.ResourceListProcessorFunc(Process), fn.WithDocs(readme, metadata)); err != nil {
		os.Exit(1)
	}
}

func run(resourceList *kubeobject.ResourceList) (kubeobject.Results, error) {
	var fn setAnnotationFunction
	err := fn.Config(resourceList.FunctionConfig.CopyToResourceNode())
	if err != nil {
		return nil, fmt.Errorf("failed to configure function: %w", err)
	}

	items, err := fn.Run(resourceList.Items.CopyToResourceNodes())
	if err != nil {
		return nil, fmt.Errorf("failed to run function: %w", err)
	}
	resourceList.Items = kubeobject.MoveToKubeObjects(items)
	return fn.resultsToItems()
}
