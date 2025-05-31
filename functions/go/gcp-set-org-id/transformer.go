package main

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gcp-set-org-id/consts"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gcp-set-org-id/plugins"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	OrgIDKey = "orgID"
)

type Transformer struct {
	FieldSpecs []types.FieldSpec `json:"organizationsIDs,omitempty" yaml:"organizationsIDs,omitempty"`
	OrgID      string
}

func (p *Transformer) Config(fnConfigNode *yaml.RNode) error {
	if err := yaml.Unmarshal([]byte(consts.OrgFieldSpec), &p); err != nil {
		return err
	}
	data := fnConfigNode.GetDataMap()
	if data == nil {
		return fmt.Errorf("missing `data` field in `ConfigMap` FunctionConfig")
	}
	orgID, ok := data[OrgIDKey]
	if !ok {
		return fmt.Errorf("missing `.data.%s` field in `ConfigMap` FunctionConfig", OrgIDKey)
	}
	p.OrgID = orgID
	return nil
}

func (p *Transformer) Transform(m resmap.ResMap) error {
	return m.ApplyFilter(plugins.Filter{
		OrgID:   p.OrgID,
		FsSlice: p.FieldSpecs,
	})
}
