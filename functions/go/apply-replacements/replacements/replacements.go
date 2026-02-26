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

package replacements

import (
	"fmt"

	"github.com/kptdev/krm-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/api/filters/replacement"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const fnConfigKind = "ApplyReplacements"
const fnConfigApiVersion = "fn.kpt.dev/v1alpha1"

func ApplyReplacements(rl *fn.ResourceList) (bool, error) {
	r := Replacements{}
	return r.Process(rl)
}

type Replacements struct {
	Replacements []types.Replacement `json:"replacements,omitempty" yaml:"replacements,omitempty"`
}

// Config initializes Replacements from a functionConfig fn.KubeObject
func (r *Replacements) Config(functionConfig *fn.KubeObject) error {
	if functionConfig.IsEmpty() {
		return fmt.Errorf("FunctionConfig is missing. Expect `ApplyReplacements`")
	}
	if functionConfig.GetKind() != fnConfigKind || functionConfig.GetAPIVersion() != fnConfigApiVersion {
		return fmt.Errorf("received functionConfig of kind %s and apiVersion %s, "+
			"only functionConfig of kind %s and apiVersion %s is supported",
			functionConfig.GetKind(), functionConfig.GetAPIVersion(), fnConfigKind, fnConfigApiVersion)
	}
	r.Replacements = []types.Replacement{}
	if err := functionConfig.As(r); err != nil {
		return fmt.Errorf("unable to convert functionConfig to %s:\n%w",
			"replacements", err)
	}
	return nil
}

// Process configures the replacements and transformers them.
func (r *Replacements) Process(rl *fn.ResourceList) (bool, error) {
	if err := r.Config(rl.FunctionConfig); err != nil {
		rl.LogResult(err)
		return false, nil
	}
	transformedItems, err := r.Transform(rl.Items)
	if err != nil {
		rl.LogResult(err)
		return false, nil
	}
	rl.Items = transformedItems
	return true, nil
}

// Transform runs the replacement filter in order to apply the replacements - this
// does the actual work.
func (r *Replacements) Transform(items []*fn.KubeObject) ([]*fn.KubeObject, error) {
	var transformedItems []*fn.KubeObject
	var nodes []*yaml.RNode

	for _, obj := range items {
		objRN, err := yaml.Parse(obj.String())
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, objRN)
	}
	transformedNodes, err := replacement.Filter{
		Replacements: r.Replacements,
	}.Filter(nodes)
	if err != nil {
		return nil, err
	}
	for _, n := range transformedNodes {
		obj, err := fn.ParseKubeObject([]byte(n.MustString()))
		if err != nil {
			return nil, err
		}
		transformedItems = append(transformedItems, obj)
	}
	return transformedItems, nil
}
