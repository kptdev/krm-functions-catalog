package main

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gcp-set-default-name/fieldspec"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gcp-set-default-name/namereference"
	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
 	nameKey = "name"
	builtinNameRef = "namereference"
)

var (
	nameRefFieldSpecs = builtinpluginconsts.GetDefaultFieldSpecsAsMap()[builtinNameRef]
)

type CustomNameTransformer struct {
	nameFieldSpecs types.FsSlice
	nameRefTransformer resmap.Transformer
	metaName string
}

func (p *CustomNameTransformer) Config(fnConfigNode *yaml.RNode) error {
	data := fnConfigNode.GetDataMap()
	if data == nil {
		return fmt.Errorf("missing `data` field in `ConfigMap` FunctionConfig")
	}
	metaName, ok := data[nameKey]
	if !ok {
		return fmt.Errorf("missing `.data.%s` field in `ConfigMap` FunctionConfig", metaName)
	}
	p.metaName = metaName

	// Enumerate field specs
	var cfs CustomFieldSpec
	if err := yaml.Unmarshal([]byte(fieldspec.CustomNameFieldSpecs), &cfs); err != nil {
		return err
	}
	p.nameFieldSpecs = cfs.CustomMetaName

	var builtinNameRefFs NameBackRefs
	if err := yaml.Unmarshal([]byte(nameRefFieldSpecs), &builtinNameRefFs); err != nil {
		return err
	}
	var customNameRefFs NameBackRefs
	if err := yaml.Unmarshal([]byte(fieldspec.NameReferenceFieldSpecs), &customNameRefFs); err != nil {
		return err
	}
	nameRefs := append(builtinNameRefFs.NameReference, customNameRefFs.NameReference...)
	p.nameRefTransformer = namereference.NewNameReferenceTransformer(nameRefs)
	return nil
}

func (p *CustomNameTransformer) Transform(m resmap.ResMap) error {
	cf := CustomFieldSpecFilter{
		MetaName: p.metaName,
		FsSlice:  p.nameFieldSpecs,
	}
	if err := m.ApplyFilter(cf); err != nil {
		return err
	}
	if err := p.nameRefTransformer.Transform(m); err != nil {
		return err
	}
	return nil
}

type NameBackRefs struct {
	NameReference     namereference.NbrSlice      `json:"nameReference,omitempty" yaml:"nameReference,omitempty"`
}

var _ kio.Filter = CustomFieldSpecFilter{}

func (f CustomFieldSpecFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	_, err := kio.FilterAll(yaml.FilterFunc(f.filter)).Filter(nodes)
	return nodes, err
}

type CustomFieldSpec struct {
	CustomMetaName []types.FieldSpec `json:"customMetaName,omitempty" yaml:"customMetaName,omitempty"`
}

type CustomFieldSpecFilter struct {
	MetaName string
	FsSlice types.FsSlice `json:"customMetaName,omitempty" yaml:"customMetaName,omitempty"`
}

func (f CustomFieldSpecFilter) filter(node *yaml.RNode) (*yaml.RNode, error) {
	if err := node.PipeE(fsslice.Filter{
		FsSlice:  f.FsSlice,
		SetValue: f.updateMetaNameFn,
	}); err != nil {
		return nil, err
	}
	return node, nil
}

func (f CustomFieldSpecFilter) updateMetaNameFn(node *yaml.RNode) error {
	return node.PipeE(updater{metaName: f.MetaName})
}

type updater struct {
	metaName string
}

func (u updater) Filter(rn *yaml.RNode) (*yaml.RNode, error) {
	return rn.Pipe(yaml.FieldSetter{StringValue: u.metaName})
}
