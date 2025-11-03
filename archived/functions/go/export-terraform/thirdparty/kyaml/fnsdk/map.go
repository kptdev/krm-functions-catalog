package fnsdk

import (
	"fmt"
	"log"
	"sort"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func newMap() *mapVariant {
	node := &yaml.Node{
		Kind: yaml.MappingNode,
	}
	return &mapVariant{node: node}
}

func newStringMapVariant(m map[string]string) *mapVariant {
	node := &yaml.Node{
		Kind: yaml.MappingNode,
	}
	for k, v := range m {
		node.Content = append(node.Content, buildStringNode(k), buildStringNode(v))
	}
	return &mapVariant{node: node}
}

type mapVariant struct {
	node *yaml.Node
}

func (o *mapVariant) Kind() variantKind {
	return variantKindMap
}

func (o *mapVariant) Node() *yaml.Node {
	return o.node
}

func (o *mapVariant) Entries() (map[string]variant, error) {
	entries := make(map[string]variant)

	ynode := o.node
	children := ynode.Content
	if len(children)%2 != 0 {
		return nil, fmt.Errorf("unexpected number of children for map %d", len(children))
	}

	for i := 0; i < len(children); i += 2 {
		keyNode := children[i]
		valueNode := children[i+1]

		keyVariant := toVariant(keyNode)
		valueVariant := toVariant(valueNode)

		switch keyVariant := keyVariant.(type) {
		case *scalarVariant:
			sv, isString := keyVariant.StringValue()
			if isString {
				entries[sv] = valueVariant
			} else {
				return nil, fmt.Errorf("key was not a string %v", keyVariant)
			}
		default:
			return nil, fmt.Errorf("unexpected variant kind %T", keyVariant)
		}
	}
	return entries, nil
}

func asString(node *yaml.Node) (string, bool) {
	if node.Kind == yaml.ScalarNode && (node.Tag == "!!str" || node.Tag == "") {
		return node.Value, true
	}
	return "", false
}

func (o *mapVariant) getVariant(key string) (variant, bool) {
	valueNode, found := getValueNode(o.node, key)
	if !found {
		return nil, found
	}

	v := toVariant(valueNode)
	return v, true
}

func getValueNode(m *yaml.Node, key string) (*yaml.Node, bool) {
	children := m.Content
	if len(children)%2 != 0 {
		log.Fatalf("unexpected number of children for map %d", len(children))
	}

	for i := 0; i < len(children); i += 2 {
		keyNode := children[i]

		k, ok := asString(keyNode)
		if ok && k == key {
			valueNode := children[i+1]
			return valueNode, true
		}
	}
	return nil, false
}

func (o *mapVariant) set(key string, val variant) {
	o.setYAMLNode(key, val.Node())
}

func (o *mapVariant) setYAMLNode(key string, node *yaml.Node) {
	children := o.node.Content
	if len(children)%2 != 0 {
		log.Fatalf("unexpected number of children for map %d", len(children))
	}

	for i := 0; i < len(children); i += 2 {
		keyNode := children[i]

		k, ok := asString(keyNode)
		if ok && k == key {
			// TODO: Copy comments?
			oldNode := children[i+1]
			children[i+1] = node
			children[i+1].FootComment = oldNode.FootComment
			children[i+1].HeadComment = oldNode.HeadComment
			children[i+1].LineComment = oldNode.LineComment
			return
		}
	}

	o.node.Content = append(o.node.Content, buildStringNode(key), node)

	return
}

func (o *mapVariant) remove(key string) (bool, error) {
	removed := false

	children := o.node.Content
	if len(children)%2 != 0 {
		return false, fmt.Errorf("unexpected number of children for map %d", len(children))
	}

	var keep []*yaml.Node
	for i := 0; i < len(children); i += 2 {
		keyNode := children[i]

		k, ok := asString(keyNode)
		if ok && k == key {
			removed = true
			continue
		}

		keep = append(keep, children[i], children[i+1])
	}

	o.node.Content = keep

	return removed, nil
}

// remove field metadata.creationTimestamp when it's null.
func (o *mapVariant) cleanupCreationTimestamp() {
	if o.node.Kind != yaml.MappingNode {
		return
	}
	scalar, found, err := o.GetNestedScalar("metadata", "creationTimestamp")
	if err != nil || !found {
		return
	}
	if scalar.IsNull() {
		_, _ = o.RemoveNestedField("metadata", "creationTimestamp")
	}
}

// sortFields tried to sort fields that it understands. e.g. data should come
// after apiVersion, kind and metadata in corev1.ConfigMap.
func (o *mapVariant) sortFields() error {
	return sortFields(o.node)
}

func sortFields(ynode *yaml.Node) error {
	pairs, err := ynodeToYamlKeyValuePairs(ynode)
	if err != nil {
		return fmt.Errorf("unable to sort fields in yaml: %w", err)
	}
	for _, pair := range pairs {
		if err = sortFields(pair.value); err != nil {
			return err
		}
	}
	sort.Sort(pairs)
	ynode.Content = yamlKeyValuePairsToYnode(pairs)
	return nil
}

func ynodeToYamlKeyValuePairs(ynode *yaml.Node) (yamlKeyValuePairs, error) {
	if len(ynode.Content)%2 != 0 {
		return nil, fmt.Errorf("invalid number of nodes: %d", len(ynode.Content))
	}

	var pairs yamlKeyValuePairs
	for i := 0; i < len(ynode.Content); i += 2 {
		pairs = append(pairs, &yamlKeyValuePair{name: ynode.Content[i], value: ynode.Content[i+1]})
	}
	return pairs, nil
}

func yamlKeyValuePairsToYnode(pairs yamlKeyValuePairs) []*yaml.Node {
	var nodes []*yaml.Node
	for _, pair := range pairs {
		nodes = append(nodes, pair.name, pair.value)
	}
	return nodes
}

type yamlKeyValuePair struct {
	name  *yaml.Node
	value *yaml.Node
}

type yamlKeyValuePairs []*yamlKeyValuePair

func (nodes yamlKeyValuePairs) Len() int { return len(nodes) }

func (nodes yamlKeyValuePairs) Less(i, j int) bool {
	iIndex, iFound := yaml.FieldOrder[nodes[i].name.Value]
	jIndex, jFound := yaml.FieldOrder[nodes[j].name.Value]
	if iFound && jFound {
		return iIndex < jIndex
	}
	if iFound {
		return true
	}
	if jFound {
		return false
	}

	if nodes[i].name != nodes[j].name {
		return nodes[i].name.Value < nodes[j].name.Value
	}
	return false
}

func (nodes yamlKeyValuePairs) Swap(i, j int) { nodes[i], nodes[j] = nodes[j], nodes[i] }
