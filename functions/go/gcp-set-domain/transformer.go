package main

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gcp-set-domain/consts"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gcp-set-domain/plugins"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	DomainKey = "domain"
)

type Transformer struct {
	Domain   string
	FsFields []types.FieldSpec `json:"domains,omitempty" yaml:"domains,omitempty"`
}

func (p *Transformer) Config(fnConfigNode *yaml.RNode) error {
	if err := yaml.Unmarshal([]byte(consts.DomainFieldSpecs), &p); err != nil {
		return err
	}
	data := fnConfigNode.GetDataMap()
	if data == nil {
		return fmt.Errorf("missing `data` field in `ConfigMap` FunctionConfig")
	}
	domain, ok := data[DomainKey]
	if !ok {
		return fmt.Errorf("missing `.data.%s` field in `ConfigMap` FunctionConfig", DomainKey)
	}
	p.Domain = domain
	return nil
}

func (p *Transformer) Transform(m resmap.ResMap) error {
	return m.ApplyFilter(plugins.Filter{
		Domain:   p.Domain,
		FsFields: p.FsFields,
	})

}
