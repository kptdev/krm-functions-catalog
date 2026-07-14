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

package applysetters

import (
	fnresult "github.com/kptdev/kpt/api/fnresult/v1"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

const SetterCommentIdentifier = "# kpt-set: "

var _ kio.Filter = &ApplySetters{}

// ApplySetters applies the setter values to the resource fields which are tagged
// by the setter reference comments
type ApplySetters struct {
	// Setters holds the user provided values for all the setters
	Setters []Setter

	// Results are the results of applying setter values
	Results []*fnresult.ResultItem

	// filePath file path of current resource
	filePath string

	// metadata of the current resource
	metadata *kyaml.ResourceIdentifier
}

type Setter struct {
	// Name is the name of the setter
	Name string

	// Value is the input value for setter
	Value string
}

// Filter implements Set as a yaml.Filter
func (as *ApplySetters) Filter(nodes []*kyaml.RNode) ([]*kyaml.RNode, error) {
	for _, node := range nodes {
		filePath, _, err := kioutil.GetFileAnnotations(node)
		if err != nil {
			return nodes, err
		}
		as.filePath = filePath
		as.metadata = &kyaml.ResourceIdentifier{
			TypeMeta: kyaml.TypeMeta{
				APIVersion: node.GetApiVersion(),
				Kind:       node.GetKind(),
			},
			NameMeta: kyaml.NameMeta{
				Name:      node.GetName(),
				Namespace: node.GetNamespace(),
			},
		}
		err = accept(as, node)
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}
	return nodes, nil
}

// Decode decodes the input yaml node into Set struct
func Decode(rn *kyaml.RNode, fcd *ApplySetters) {
	for k, v := range rn.GetDataMap() {
		fcd.Setters = append(fcd.Setters, Setter{Name: k, Value: v})
	}
}
