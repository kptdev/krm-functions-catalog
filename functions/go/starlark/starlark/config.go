package starlark

import (
	"fmt"

	"github.com/kptdev/krm-functions-catalog/functions/go/starlark/third_party/sigs.k8s.io/kustomize/kyaml/fn/runtime/starlark"
	"github.com/kptdev/krm-functions-sdk/go/fn"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	starlarkRunGroup      = "fn.kpt.dev"
	starlarkRunVersion    = "v1alpha1"
	starlarkRunAPIVersion = starlarkRunGroup + "/" + starlarkRunVersion
	starlarkRunKind       = "StarlarkRun"

	configMapApiVersion = "v1"
	configMapKind       = "ConfigMap"

	sourceKey = "source"

	defaultProgramName = "starlark-function-run"
)

type StarlarkRun struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	// Source is a required field for providing a starlark script inline.
	Source string `json:"source" yaml:"source"`
	// Params are the parameters in key-value pairs format.
	Params map[string]interface{} `json:"params,omitempty" yaml:"params,omitempty"`
}

func (sr *StarlarkRun) Config(fnCfg *fn.KubeObject) error {
	switch {
	case fnCfg.IsEmpty():
		return fmt.Errorf("FunctionConfig is missing. Expect `ConfigMap` or `StarlarkRun`")
	case fnCfg.IsGroupKind(schema.GroupKind{Group: "", Kind: configMapKind}):
		cm := &corev1.ConfigMap{}
		if err := fnCfg.As(cm); err != nil {
			return err
		}
		// Convert ConfigMap to StarlarkRun
		sr.Name = cm.Name
		sr.Namespace = cm.Namespace
		sr.Params = map[string]interface{}{}
		for k, v := range cm.Data {
			if k == sourceKey {
				sr.Source = v
			}
			sr.Params[k] = v
		}
	case fnCfg.IsGroupKind(schema.GroupKind{Group: starlarkRunGroup, Kind: starlarkRunKind}):
		if err := fnCfg.As(sr); err != nil {
			return err
		}
	default:
		return fmt.Errorf("`functionConfig` must be either %v or %v, but we got: %v",
			schema.FromAPIVersionAndKind(configMapApiVersion, configMapKind).String(),
			schema.FromAPIVersionAndKind(starlarkRunAPIVersion, starlarkRunKind).String(),
			schema.FromAPIVersionAndKind(fnCfg.GetAPIVersion(), fnCfg.GetKind()).String())
	}

	// Defaulting
	if sr.Name == "" {
		sr.Name = defaultProgramName
	}
	// Validation
	if sr.Source == "" {
		return fmt.Errorf("`source` must not be empty")
	}
	return nil
}

func (sr *StarlarkRun) Transform(rl *fn.ResourceList) error {
	var transformedObjects []*fn.KubeObject
	var nodes []*yaml.RNode

	fcRN, err := yaml.Parse(rl.FunctionConfig.String())
	if err != nil {
		return err
	}
	for _, obj := range rl.Items {
		objRN, err := yaml.Parse(obj.String())
		if err != nil {
			return err
		}
		nodes = append(nodes, objRN)
	}

	starFltr := &starlark.SimpleFilter{
		Name:           sr.Name,
		Program:        sr.Source,
		FunctionConfig: fcRN,
	}
	transformedNodes, err := starFltr.Filter(nodes)
	if err != nil {
		return err
	}

	for _, n := range transformedNodes {
		obj, err := fn.ParseKubeObject([]byte(n.MustString()))
		if err != nil {
			return err
		}
		transformedObjects = append(transformedObjects, obj)
	}
	rl.Items = transformedObjects
	return nil
}
