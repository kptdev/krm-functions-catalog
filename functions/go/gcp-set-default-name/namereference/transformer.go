package namereference

import (
	"log"

	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

type nameReferenceTransformer struct {
	backRefs []NameBackReferences
}

type NameBackReferences struct {
	resid.Gvk `json:",inline,omitempty" yaml:",inline,omitempty"`
	Referrers types.FsSlice `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

type NbrSlice []NameBackReferences

var _ resmap.Transformer = &nameReferenceTransformer{}

type filterMap map[*resource.Resource][]UnlimitedNameRefFilter

func NewNameReferenceTransformer(
	br []NameBackReferences) resmap.Transformer {
	if br == nil {
		log.Fatal("backrefs not expected to be nil")
	}
	return &nameReferenceTransformer{backRefs: br}
}

func (t *nameReferenceTransformer) Transform(m resmap.ResMap) error {
	fMap := t.determineFilters(m.Resources())
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

func (t *nameReferenceTransformer) determineFilters(
	resources []*resource.Resource) (fMap filterMap) {
	// We cache the resource OrgId values because they don't change and otherwise are very visible in a memory pprof
	fMap = make(filterMap)
	for _, backReference := range t.backRefs {
		for _, referrerSpec := range backReference.Referrers {
			for _, res := range resources {
				if res.OrgId().IsSelected(&referrerSpec.Gvk) {
						fMap[res] = append(fMap[res], UnlimitedNameRefFilter{
							NameFieldToUpdate: referrerSpec,
							ReferralTarget: backReference.Gvk,
						})
				}
			}
		}
	}
	return fMap
}