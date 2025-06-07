// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0
//
// Copied and modified from
// https://github.com/kubernetes-sigs/kustomize/blob/3265f64cd5ea76a8b64877b193576e2d120001db/api/internal/accumulator/namereferencetransformer.go

package nameref

import (
	"sigs.k8s.io/kustomize/api/filters/nameref"
	"sigs.k8s.io/kustomize/api/konfig/builtinpluginconsts"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/yaml"
)

type filterMap map[*resource.Resource][]nameref.Filter

type nameBackReferences struct {
	resid.Gvk `json:",inline,omitempty" yaml:",inline,omitempty"`
	Referrers types.FsSlice `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

type nameReferenceConfig struct {
	NameReference []nameBackReferences `json:"nameReference,omitempty" yaml:"nameReference,omitempty"`
}

// FixNameBackReference updates name references in resource A that
// refer to resource B, given that B's name may have
// changed.
//
// For example, a HorizontalPodAutoscaler (HPA)
// necessarily refers to a Deployment, the thing that
// an HPA scales. In this case:
//
//   - the HPA instance is the Referrer,
//   - the Deployment instance is the ReferralTarget.
//
// If the Deployment's name changes, e.g. a prefix is added,
// then the HPA's reference to the Deployment must be fixed.
func FixNameBackReference(m resmap.ResMap) error {
	c, err := getDefaultConfig()
	if err != nil {
		return err
	}
	fMap := determineFilters(m.Resources(), c.NameReference)
	for r, fList := range fMap {
		c := m.SubsetThatCouldBeReferencedByResource(r)
		for _, f := range fList {
			f.Referrer = r
			f.ReferralCandidates = c
			if err := f.Referrer.ApplyFilter(f); err != nil {
				return err
			}
		}
	}
	return nil
}

func getDefaultConfig() (nameReferenceConfig, error) {
	defaultConfigString := builtinpluginconsts.GetDefaultFieldSpecsAsMap()["namereference"]
	var tc nameReferenceConfig
	err := yaml.Unmarshal([]byte(defaultConfigString), &tc)
	return tc, err
}

func determineFilters(resources []*resource.Resource, backRefs []nameBackReferences) (fMap filterMap) {
	fMap = make(filterMap)
	for _, backReference := range backRefs {
		for _, referrerSpec := range backReference.Referrers {
			for _, res := range resources {
				if res.OrgId().IsSelected(&referrerSpec.Gvk) {
					// Optimization - the referrer has the field
					// that might need updating.
					fMap[res] = append(fMap[res], nameref.Filter{
						// Name field to write in the Referrer.
						// If the path specified here isn't found in
						// the Referrer, nothing happens (no error,
						// no field creation).
						NameFieldToUpdate: referrerSpec,
						// Specification of object class to read from.
						// Always read from metadata/name field.
						ReferralTarget: backReference.Gvk,
					})
				}
			}
		}
	}
	return fMap
}
