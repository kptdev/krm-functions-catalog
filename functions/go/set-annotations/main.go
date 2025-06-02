// This file will be processed and embedded to pluginator.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/kptdev/krm-functions-catalog/functions/go/set-annotations/generated"
	"sigs.k8s.io/kustomize/api/hasher"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
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

type SetAnnotationsProcessor struct{}

func (sap *SetAnnotationsProcessor) Process(resourceList *framework.ResourceList) error {
	results, err := run(resourceList)
	if err != nil {
		resourceList.Results = framework.Results{
			&framework.Result{
				Message:  err.Error(),
				Severity: framework.Error,
			},
		}
		return resourceList.Results
	}
	resourceList.Results = results
	return nil
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
		f.plugin.Annotations = rn.GetDataMap()
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

	if len(f.plugin.Annotations) == 0 {
		return fmt.Errorf("input annotation list cannot be empty")
	}
	tc, err := getDefaultConfig()
	if err != nil {
		return err
	}
	// append default field specs
	f.plugin.AdditionalAnnotationFields = append(f.plugin.AdditionalAnnotationFields, tc.FieldSpecs...)
	return nil
}

func (f *setAnnotationFunction) Run(items []*kyaml.RNode) ([]*kyaml.RNode, error) {
	resmapFactory := newResMapFactory()
	resMap, err := resmapFactory.NewResMapFromRNodeSlice(items)
	if err != nil {
		return nil, err
	}
	err = f.plugin.Transform(resMap)
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
// equivalent framework.Results
func (f *setAnnotationFunction) resultsToItems() (framework.Results, error) {
	var results framework.Results
	if len(f.plugin.Results) == 0 {
		results = append(results, &framework.Result{
			Message: "no annotations applied",
		})
		return results, nil
	}
	for resKey, annoVals := range f.plugin.Results {
		fileIndex, _ := strconv.Atoi(resKey.FileIndex)
		annotationJson, err := json.Marshal(annoVals)
		if err != nil {
			return nil, err
		}
		results = append(results, &framework.Result{
			Message: fmt.Sprintf("set annotations: %s", annotationJson),
			Field:   &framework.Field{Path: resKey.FieldPath},
			File:    &framework.File{Path: resKey.FilePath, Index: fileIndex},
		})
	}
	results.Sort()
	return results, nil
}

func getDefaultConfig() (transformerConfig, error) {
	defaultConfigString := builtinpluginconsts.GetDefaultFieldSpecsAsMap()["commonannotations"]
	var tc transformerConfig
	err := yaml.Unmarshal([]byte(defaultConfigString), &tc)
	return tc, err
}

func newResMapFactory() *resmap.Factory {
	resourceFactory := resource.NewFactory(&hasher.Hasher{})
	resourceFactory.IncludeLocalConfigs = true
	return resmap.NewFactory(resourceFactory)
}

// nolint
func main() {
	asp := SetAnnotationsProcessor{}
	cmd := command.Build(&asp, command.StandaloneEnabled, false)

	cmd.Short = generated.SetAnnotationsShort
	cmd.Long = generated.SetAnnotationsLong
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(resourceList *framework.ResourceList) (framework.Results, error) {
	var fn setAnnotationFunction
	err := fn.Config(resourceList.FunctionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to configure function: %w", err)
	}

	resourceList.Items, err = fn.Run(resourceList.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to run function: %w", err)
	}
	return fn.resultsToItems()
}
