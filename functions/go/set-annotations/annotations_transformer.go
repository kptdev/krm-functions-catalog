// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate pluginator
package main

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/api/filters/annotations"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"
)

// Add the given annotations to the given field specifications.
type plugin struct {
	// Desired annotations
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	// FieldSpecs is deprecated, please use AdditionalAnnotationFields instead.
	FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	// AdditionalAnnotationFields is used to specify additional fields to add annotations.
	AdditionalAnnotationFields []types.FieldSpec `json:"additionalAnnotationFields,omitempty" yaml:"additionalAnnotationFields,omitempty"`
	// Results are the results of applying annotations
	Results AnnotationResults
}

// AnnotationResults maps annotation paths to key/value pairs
type AnnotationResults map[AnnotationResultKey]AnnotationValues

// AnnotationResultKey is a unique representation for an annotations field path
type AnnotationResultKey struct {
	// FilePath is the file path of the resource
	FilePath string
	// FileIndex is the file index of the resource
	FileIndex string
	// FieldPath is field path of the annotations
	FieldPath string
}

// AnnotationValues represents annotation key/value pairs
type AnnotationValues map[string]string

// noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	p.Annotations = nil
	p.FieldSpecs = nil
	p.AdditionalAnnotationFields = nil
	if err = yaml.Unmarshal(c, p); err != nil {
		return fmt.Errorf("failed to unmarshal config %#v: %w", string(c), err)
	}
	if p.AdditionalAnnotationFields != nil && p.FieldSpecs != nil {
		return fmt.Errorf("`fieldSpecs` has been deprecated, please rename it to `additionalAnnotationFields`")
	}
	if p.AdditionalAnnotationFields == nil && p.FieldSpecs != nil {
		p.AdditionalAnnotationFields = p.FieldSpecs
	}
	return nil
}

func (p *plugin) Transform(m resmap.ResMap) error {
	if len(p.Annotations) == 0 {
		return nil
	}
	if p.Results == nil {
		p.Results = make(AnnotationResults)
	}
	for _, r := range m.Resources() {
		filePath, fileIndex, err := kioutil.GetFileAnnotations(&r.RNode)
		if err != nil {
			return err
		}
		err = r.ApplyFilter(annotations.Filter{
			FsSlice:     p.AdditionalAnnotationFields,
			Annotations: p.Annotations,
			SetEntryCallback: func(key, value, tag string, node *kyaml.RNode) {
				resultKey := AnnotationResultKey{
					FieldPath: strings.Join(node.FieldPath(), "."),
					FilePath:  filePath,
					FileIndex: fileIndex,
				}
				result, ok := p.Results[resultKey]
				if ok {
					result[key] = value
				} else {
					p.Results[resultKey] = AnnotationValues{key: value}
				}
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
