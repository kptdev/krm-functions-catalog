// Copyright 2022 Google LLC
// Modifications Copyright (C) 2025 OpenInfra Foundation Europe.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package gcloudconfig

import (
	"errors"

	"github.com/kptdev/krm-functions-catalog/archived/functions/go/source-gcloud-generator/exec"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/resid"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const GcloudMetaName = "gcloud-config.kpt.dev"

var LocalConfigAnnos = map[string]string{filters.LocalConfigAnnotation: "true"}

// GcloudConfigGenerator is the generator function to generate a GcloudConfig resource.
type GcloudConfigGenerator struct{}

// Generate executes `gcloud` commands to create a GcloudConfig RNode.
func (g *GcloudConfigGenerator) Generate(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	gcloudData, err := exec.GetGcloudContextFn()
	var gcloudErr *exec.GcloudErr
	if err != nil {
		if !errors.As(err, &gcloudErr) {
			return nil, err
		}
	}
	gcloudNode, err := NewGcloudConfigNode(gcloudData)
	if err != nil {
		return nil, err
	}
	gcloudGvk := resid.GvkFromNode(gcloudNode)
	var newNodes []*yaml.RNode
	exist := false
	for _, curNode := range nodes {
		curNodeGvk := resid.GvkFromNode(curNode)
		if curNode.GetName() == gcloudNode.GetName() && curNodeGvk.Equals(gcloudGvk) {
			curNode.SetDataMap(gcloudNode.GetDataMap())
			if path, ok := curNode.GetAnnotations()[kioutil.PathAnnotation]; ok && path == ResultFile {
				exist = true
			}
		}
		newNodes = append(newNodes, curNode)
	}
	if !exist {
		newNodes = append(newNodes, gcloudNode)
	}
	if gcloudErr == nil {
		return newNodes, nil
	}
	return newNodes, gcloudErr
}

// NewGcloudConfigNode creates a `GcloudConfig` RNode resource.
func NewGcloudConfigNode(data map[string]string) (*yaml.RNode, error) {
	cm := yaml.MustParse(`
apiVersion: v1
kind: ConfigMap
metadata:
  name:
data: {}
`)
	// !! The ConfigMap should always be assigned to this value to make it "convention over configuration".
	if err := cm.SetName(GcloudMetaName); err != nil {
		return nil, err
	}
	// This resource is pseudo resource and not expected to be deployed to a cluster.
	if err := cm.SetAnnotations(LocalConfigAnnos); err != nil {
		return nil, err
	}
	cm.SetDataMap(data)
	return cm, nil
}
