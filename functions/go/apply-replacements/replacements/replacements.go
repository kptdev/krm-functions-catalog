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
	"sigs.k8s.io/kustomize/kyaml/resid"
	kyaml_utils "sigs.k8s.io/kustomize/kyaml/utils"
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

	// Fix: kustomize's setFieldValue only copies .Value but not .Tag/.Style,
	// causing string-typed sources like "18" to lose quoting at the target.
	// Post-process to restore string typing from source to target fields.
	preserveStringTyping(transformedNodes, nodes, r.Replacements)

	for _, n := range transformedNodes {
		obj, err := fn.ParseKubeObject([]byte(n.MustString()))
		if err != nil {
			return nil, err
		}
		transformedItems = append(transformedItems, obj)
	}
	return transformedItems, nil
}

// preserveStringTyping finds replacement targets whose source is a string scalar
// and ensures the target field retains the string tag and quoting style.
func preserveStringTyping(nodes, originalNodes []*yaml.RNode, replacements []types.Replacement) {
	for i := range replacements {
		r := &replacements[i]
		if r.Source == nil || r.Targets == nil {
			continue
		}

		srcNode := findSourceNode(originalNodes, r.Source)
		if srcNode == nil || srcNode.YNode().Kind != yaml.ScalarNode {
			continue
		}
		if srcNode.YNode().Tag != yaml.NodeTagString &&
			srcNode.YNode().Style != yaml.DoubleQuotedStyle &&
			srcNode.YNode().Style != yaml.SingleQuotedStyle {
			continue
		}

		for _, target := range r.Targets {
			if target.Select == nil {
				continue
			}
			for _, node := range nodes {
				if !nodeMatchesSelector(node, target.Select) {
					continue
				}
				for _, fp := range target.FieldPaths {
					path := kyaml_utils.SmarterPathSplitter(fp, ".")
					field, err := node.Pipe(yaml.Lookup(path...))
					if err != nil || field == nil {
						continue
					}
					if field.YNode().Kind == yaml.ScalarNode {
						field.YNode().Tag = srcNode.YNode().Tag
						field.YNode().Style = srcNode.YNode().Style
					}
				}
			}
		}
	}
}

// findSourceNode locates the source resource and field for a replacement.
func findSourceNode(nodes []*yaml.RNode, source *types.SourceSelector) *yaml.RNode {
	for _, n := range nodes {
		if !nodeMatchesResId(n, source.ResId) {
			continue
		}
		fieldPath := source.FieldPath
		if fieldPath == "" {
			fieldPath = types.DefaultReplacementFieldPath
		}
		path := kyaml_utils.SmarterPathSplitter(fieldPath, ".")
		rn, err := n.Pipe(yaml.Lookup(path...))
		if err != nil || rn.IsNilOrEmpty() {
			return nil
		}
		return rn
	}
	return nil
}

// nodeMatchesResId checks if a node matches a ResId by GVK, name, and namespace.
func nodeMatchesResId(n *yaml.RNode, id resid.ResId) bool {
	group, version := resid.ParseGroupVersion(n.GetApiVersion())
	nodeId := resid.NewResIdWithNamespace(
		resid.Gvk{Group: group, Version: version, Kind: n.GetKind()},
		n.GetName(), n.GetNamespace(),
	)
	return nodeId.IsSelectedBy(id)
}

// nodeMatchesSelector checks if a node matches a target Selector.
func nodeMatchesSelector(n *yaml.RNode, sel *types.Selector) bool {
	return nodeMatchesResId(n, sel.ResId)
}
