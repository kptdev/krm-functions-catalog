package plugins

import (
	"sigs.k8s.io/kustomize/api/filters/fieldspec"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const OrgKind = "Organization"

var _ kio.Filter = Filter{}

type Filter struct {
	OrgID   string
	FsSlice []types.FieldSpec
}

// isOrg determins if the rnode in current fieldspec has "kind: Organization". Expecting the orgId to be
// spec:
//   resourceRef:
//      apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
//      kind: Organization  <-- kind has to be Organization
//      external: 433637338589 <-- OrgID
func isOrg(node *yaml.RNode) bool {
	for i := 0; i < len(node.YNode().Content); i += 2 {
		if node.YNode().Content[i].Value == "kind" {
			return node.YNode().Content[i+1].Value == OrgKind
		}
	}
	return false
}

// setOrg updates the "external" rnode's value to OrgID.
func setOrg(node *yaml.RNode, orgID string) {
	for i := 0; i < len(node.YNode().Content); i += 2 {
		if node.YNode().Content[i].Value == "external" {
			node.YNode().Content[i+1].SetString(orgID)
		}
	}
}

func (f Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	_, err := kio.FilterAll(yaml.FilterFunc(
		func(node *yaml.RNode) (*yaml.RNode, error) {
			var fns []yaml.Filter
			for _, fs := range f.FsSlice {
				fn := fieldspec.Filter{
					FieldSpec: fs,
					SetValue: func(node *yaml.RNode) error {
						if isOrg(node) {
							setOrg(node, f.OrgID)
						}
						return nil
					},
				}
				fns = append(fns, fn)
			}
			return node.Pipe(fns...)
		})).Filter(nodes)
	return nodes, err
}
