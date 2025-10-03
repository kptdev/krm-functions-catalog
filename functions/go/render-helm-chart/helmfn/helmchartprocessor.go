// Copyright 2021 Google LLC
// Modifications Copyright (C) 2025 OpenInfra Foundation Europe.
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

package helmfn

import (
	"fmt"

	"github.com/kptdev/krm-functions-catalog/functions/go/render-helm-chart/third_party/sigs.k8s.io/kustomize/api/builtins"
	"github.com/kptdev/krm-functions-sdk/go/fn"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	fnConfigKind = "RenderHelmChart"
	configMap    = "ConfigMap"
)

type HelmChartProcessor struct {
	plugins []builtins.HelmChartInflationGeneratorPlugin
}

func RenderHelmChart(rl *fn.ResourceList) (bool, error) {
	r := HelmChartProcessor{}
	return r.Process(rl)
}

func (hcp *HelmChartProcessor) Process(rl *fn.ResourceList) (bool, error) {
	err := hcp.config(rl.FunctionConfig)
	if err != nil {
		return false, fmt.Errorf("failed to configure function: %w", err)
	}
	rl.Items, err = hcp.run(rl.Items)
	if err != nil {
		return false, fmt.Errorf("failed to run function: %w", err)
	}
	return true, nil
}

func (f *HelmChartProcessor) config(o *fn.KubeObject) error {
	var err error
	switch o.GetKind() {
	case fnConfigKind:
		err = f.renderHelmChartArgs(o)
	case configMap:
		err = f.configMapArgs(o)
	default:
		err = fmt.Errorf("`functionConfig` must be `%s` or `%s`", configMap, fnConfigKind)
	}
	return err
}

func (hcp *HelmChartProcessor) run(objs []*fn.KubeObject) ([]*fn.KubeObject, error) {
	for _, p := range hcp.plugins {
		err := p.ConfigureAuth(objs)
		if err != nil {
			return nil, err
		}
		generated, err := p.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to run generator: %w", err)
		}

		for _, gen := range generated {
			duplicate := false
			for _, o := range objs {
				// check for duplicates for idempotency
				if gen.IsGroupKind(o.GroupKind()) &&
					gen.GetName() == o.GetName() &&
					gen.GetNamespace() == o.GetNamespace() {
					duplicate = true
				}
			}
			if !duplicate {
				objs = append(objs, gen)
			}
		}
	}
	return objs, nil
}

func (hcp *HelmChartProcessor) renderHelmChartArgs(o *fn.KubeObject) (err error) {
	y := o.String()
	c := []byte(y)
	args := &builtins.HelmArgs{}
	if err = kyaml.Unmarshal(c, args); err != nil {
		return
	}
	for _, helmChart := range args.HelmCharts {
		p := builtins.HelmChartInflationGeneratorPlugin{
			HelmGlobals: args.HelmGlobals,
			HelmChart:   helmChart,
		}
		if err := p.ValidateArgs(); err != nil {
			return err
		}
		hcp.plugins = append(hcp.plugins, p)
	}

	return nil
}

func (hcp *HelmChartProcessor) configMapArgs(
	m *fn.KubeObject) (err error) {
	var p builtins.HelmChartInflationGeneratorPlugin
	if val, found, _ := m.NestedString("data", "chartHome"); found {
		p.ChartHome = val
	}
	if val, found, _ := m.NestedString("data", "configHome"); found {
		p.ConfigHome = val
	}
	if val, found, _ := m.NestedString("data", "name"); found {
		p.Name = val
	}
	if val, found, _ := m.NestedString("data", "version"); found {
		p.Version = val
	}
	if val, found, _ := m.NestedString("data", "repo"); found {
		p.Repo = val
	}
	if val, found, _ := m.NestedString("data", "releaseName"); found {
		p.ReleaseName = val
	}
	if val, found, _ := m.NestedString("data", "namespace"); found {
		p.Namespace = val
	}
	if val, found, _ := m.NestedString("data", "nameTemplate"); found {
		p.NameTemplate = val
	}
	if val, found, _ := m.NestedString("data", "includeCRDs"); found {
		if val == "true" {
			p.IncludeCRDs = true
		}
	}
	if val, found, _ := m.NestedString("data", "skipTests"); found {
		if val == "true" {
			p.SkipTests = true
		}
	}
	if val, found, _ := m.NestedString("data", "valuesFile"); found {
		p.ValuesFiles = []string{val}
	}
	if err := p.ValidateArgs(); err != nil {
		return err
	}
	hcp.plugins = append(hcp.plugins, p)
	return nil
}
