package fnsdk

import (
	"strconv"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	tagString = "!!str"
	tagBool   = "!!bool"
	tagInt    = "!!int"
	tagFloat  = "!!float"
	tagNull   = "!!null"
)

type scalarVariant struct {
	node *yaml.Node
}

func (v *scalarVariant) Kind() variantKind {
	return variantKindScalar
}

func newStringScalarVariant(s string) *scalarVariant {
	return &scalarVariant{
		node: buildStringNode(s),
	}
}

func newBoolScalarVariant(b bool) *scalarVariant {
	return &scalarVariant{
		node: buildBoolNode(b),
	}
}

func newIntScalarVariant(i int) *scalarVariant {
	return &scalarVariant{
		node: buildIntNode(i),
	}
}

func newFloatScalarVariant(f float64) *scalarVariant {
	return &scalarVariant{
		node: buildFloatNode(f),
	}
}

func (v *scalarVariant) IsNull() bool {
	return v.node.Tag == tagNull
}

func (v *scalarVariant) StringValue() (string, bool) {
	switch v.node.Tag {
	case tagString:
		return v.node.Value, true
	default:
		return "", false
	}
}

func (v *scalarVariant) BoolValue() (bool, bool) {
	switch v.node.Tag {
	case tagBool:
		b, err := strconv.ParseBool(v.node.Value)
		if err != nil {
			return b, false
		}
		return b, true
	default:
		return false, false
	}
}

func (v *scalarVariant) IntValue() (int, bool) {
	switch v.node.Tag {
	case tagInt:
		i, err := strconv.Atoi(v.node.Value)
		if err != nil {
			return i, false
		}
		return i, true
	default:
		return 0, false
	}
}

func (v *scalarVariant) FloatValue() (float64, bool) {
	switch v.node.Tag {
	case tagFloat:
		f, err := strconv.ParseFloat(v.node.Value, 64)
		if err != nil {
			return f, false
		}
		return f, true
	default:
		return 0, false
	}
}

func (v *scalarVariant) Node() *yaml.Node {
	return v.node
}
