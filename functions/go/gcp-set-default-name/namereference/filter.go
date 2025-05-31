package namereference

import (
	"fmt"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type UnlimitedNameRefFilter struct {
	Referrer *resource.Resource
	NameFieldToUpdate types.FieldSpec
	ReferralTarget resid.Gvk
	ReferralCandidates resmap.ResMap

}

func (f UnlimitedNameRefFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	return kio.FilterAll(yaml.FilterFunc(f.run)).Filter(nodes)
}

func (f UnlimitedNameRefFilter) run(node *yaml.RNode) (*yaml.RNode, error) {
	if node.GetNamespace() != f.Referrer.GetNamespace() {
		return nil, fmt.Errorf("not in the same namespace")
	}
	if err := node.PipeE(fieldspec.Filter{
		FieldSpec: f.NameFieldToUpdate,
		SetValue:  f.updateName,
	}); err != nil {
		return nil, errors.Wrapf(
			err, "updating name reference in '%s' field of '%s'",
			f.NameFieldToUpdate.Path, f.Referrer.CurId().String())
	}
	return node, nil
}

func (f UnlimitedNameRefFilter) updateName(node *yaml.RNode) error {
	if yaml.IsMissingOrNull(node) {
		return nil
	}
	candidates := f.ReferralCandidates.Resources()
	candidates = doSieve(candidates, previousIdSelectedByGvk(&f.ReferralTarget))
	candidates = doSieve(candidates, f.sameCurrentNamespaceAsReferrer())
	if len(candidates) == 0 {
		return nil
	}
	referral := candidates[0]
	return node.PipeE(yaml.FieldSetter{StringValue: referral.GetName()})
}

type sieveFunc func(*resource.Resource) bool

func doSieve(list []*resource.Resource, fn sieveFunc) (s []*resource.Resource) {
	for _, r := range list {
		if fn(r) {
			s = append(s, r)
		}
	}
	return
}

func previousIdSelectedByGvk(gvk *resid.Gvk) sieveFunc {
	return func(r *resource.Resource) bool {
		if r.OrgId().IsSelected(gvk) {
			return true
		}
		return false
	}
}

func (f UnlimitedNameRefFilter) sameCurrentNamespaceAsReferrer() sieveFunc {
	referrerCurId := f.Referrer.CurId()
	if referrerCurId.IsClusterScoped() {
		// If the referrer is cluster-scoped, let anything through.
		return func(_ *resource.Resource) bool {return true}
	}
	return func(r *resource.Resource) bool {
		if r.CurId().IsClusterScoped() {
			// Allow cluster-scoped through.
			return true
		}
		if r.GetKind() == "ServiceAccount" {
			// Allow service accounts through, even though they
			// are in a namespace.  A RoleBinding in another namespace
			// can reference them.
			return true
		}
		return referrerCurId.IsNsEquals(r.CurId())
	}
}