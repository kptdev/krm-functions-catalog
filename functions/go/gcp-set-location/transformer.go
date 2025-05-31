package main

import (
	"fmt"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gcp-set-location/consts"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/gcp-set-location/filedspec"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	ZoneKey   = "zone"
	RegionKey = "region"
)

type Transformer struct {
	RegionFsSlice []types.FieldSpec `json:"regions,omitempty" yaml:"regions,omitempty"`
	ZoneFsSlice   []types.FieldSpec `json:"zones,omitempty" yaml:"zones,omitempty"`
	custom        LocationFilter
}

type NoOp struct {
}

func (e *NoOp) Error() string {
	return fmt.Sprintf("neither `region` nor `zone` is given")
}

func (t *Transformer) Config(fnConfigNode *yaml.RNode) error {
	data := fnConfigNode.GetDataMap()
	if data == nil {
		return fmt.Errorf("missing `data` field in `ConfigMap` FunctionConfig")
	}
	region, ok := data[RegionKey]
	if ok && region != "" {
		if deactive, ok := consts.CloudRegions[region]; !ok || !deactive {
			return fmt.Errorf("unknown region %v", region)
		}
	}
	t.custom.Region = region

	zone, ok := data[ZoneKey]
	if ok && zone != "" {
		if deactive, ok := consts.CloudZones[zone]; !ok || !deactive {
			return fmt.Errorf("unknown zone %v", zone)
		}
	}
	if zone == "" && region == "" {
		return &NoOp{}
	}
	t.custom.Zone = zone

	if err := yaml.Unmarshal([]byte(filedspec.LocationFieldSpecs), &t); err != nil {
		return err
	}
	t.custom.RegionFsSlice = t.RegionFsSlice
	t.custom.ZoneFsSlice = t.ZoneFsSlice
	return nil
}

func (p *Transformer) Transform(m resmap.ResMap) error {
	return m.ApplyFilter(p.custom)
}
