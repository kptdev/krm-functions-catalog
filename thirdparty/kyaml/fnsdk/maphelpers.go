package fnsdk

import (
	"fmt"
)

func (o *mapVariant) GetNestedValue(fields ...string) (variant, bool, error) {
	current := o
	n := len(fields)
	for i := 0; i < n; i++ {
		entry, found := current.getVariant(fields[i])
		if !found {
			return nil, found, nil
		}

		if i == n-1 {
			return entry, true, nil
		} else {
			entryM, ok := entry.(*mapVariant)
			if !ok {
				return nil, found, fmt.Errorf("wrong type, got: %T", entry)
			}
			current = entryM
		}
	}
	return nil, false, fmt.Errorf("unexpected code reached")
}

func (o *mapVariant) SetNestedValue(val variant, fields ...string) error {
	current := o
	n := len(fields)
	var err error
	for i := 0; i < n; i++ {
		if i == n-1 {
			current.set(fields[i], val)
		} else {
			current, _, err = current.getMap(fields[i], true)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *mapVariant) GetNestedMap(fields ...string) (*mapVariant, bool, error) {
	v, found, err := o.GetNestedValue(fields...)
	if err != nil || !found {
		return nil, found, err
	}
	mv, ok := v.(*mapVariant)
	if !ok {
		return nil, found, fmt.Errorf("wrong type, got: %T", v)
	}
	return mv, found, err
}

func (o *mapVariant) SetNestedMap(m *mapVariant, fields ...string) error {
	return o.SetNestedValue(m, fields...)
}

func (o *mapVariant) GetNestedStringMap(fields ...string) (map[string]string, bool, error) {
	v, found, err := o.GetNestedValue(fields...)
	if err != nil || !found {
		return nil, found, err
	}
	children := v.Node().Content
	if len(children)%2 != 0 {
		return nil, found, fmt.Errorf("invalid yaml map node")
	}
	m := make(map[string]string, len(children)/2)
	for i := 0; i < len(children); i = i + 2 {
		m[children[i].Value] = children[i+1].Value
	}
	return m, found, nil
}

func (o *mapVariant) SetNestedStringMap(m map[string]string, fields ...string) error {
	return o.SetNestedMap(newStringMapVariant(m), fields...)
}

func (o *mapVariant) GetNestedScalar(fields ...string) (*scalarVariant, bool, error) {
	node, found, err := o.GetNestedValue(fields...)
	if err != nil || !found {
		return nil, found, err
	}
	nodeS, ok := node.(*scalarVariant)
	if !ok {
		return nil, found, fmt.Errorf("incorrect type, was %T", node)
	}
	return nodeS, found, nil
}

func (o *mapVariant) GetNestedString(fields ...string) (string, bool, error) {
	scalar, found, err := o.GetNestedScalar(fields...)
	if err != nil || !found {
		return "", found, err
	}
	sv, isString := scalar.StringValue()
	if isString {
		return sv, found, nil
	}
	return "", found, fmt.Errorf("node was not a string, was %v", scalar.node.Tag)
}

func (o *mapVariant) SetNestedString(s string, fields ...string) error {
	return o.SetNestedValue(newStringScalarVariant(s), fields...)
}

func (o *mapVariant) GetNestedBool(fields ...string) (bool, bool, error) {
	scalar, found, err := o.GetNestedScalar(fields...)
	if err != nil || !found {
		return false, found, err
	}
	bv, isBool := scalar.BoolValue()
	if isBool {
		return bv, found, nil
	}
	return false, found, fmt.Errorf("node was not a bool, was %v", scalar.Node().Tag)
}

func (o *mapVariant) SetNestedBool(b bool, fields ...string) error {
	return o.SetNestedValue(newBoolScalarVariant(b), fields...)
}

func (o *mapVariant) GetNestedInt(fields ...string) (int, bool, error) {
	scalar, found, err := o.GetNestedScalar(fields...)
	if err != nil || !found {
		return 0, found, err
	}
	iv, isInt := scalar.IntValue()
	if isInt {
		return iv, found, nil
	}
	return 0, found, fmt.Errorf("node was not a int, was %v", scalar.node.Tag)
}

func (o *mapVariant) SetNestedInt(i int, fields ...string) error {
	return o.SetNestedValue(newIntScalarVariant(i), fields...)
}

func (o *mapVariant) GetNestedFloat(fields ...string) (float64, bool, error) {
	scalar, found, err := o.GetNestedScalar(fields...)
	if err != nil || !found {
		return 0, found, err
	}
	fv, isFloat := scalar.FloatValue()
	if isFloat {
		return fv, found, nil
	}
	return 0, found, fmt.Errorf("node was not a float, was %v", scalar.node.Tag)
}

func (o *mapVariant) SetNestedFloat(f float64, fields ...string) error {
	return o.SetNestedValue(newFloatScalarVariant(f), fields...)
}

func (o *mapVariant) GetNestedSlice(fields ...string) (*sliceVariant, bool, error) {
	node, found, err := o.GetNestedValue(fields...)
	if err != nil || !found {
		return nil, found, err
	}
	nodeS, ok := node.(*sliceVariant)
	if !ok {
		return nil, found, fmt.Errorf("incorrect type, was %T", node)
	}
	return nodeS, found, err
}

func (o *mapVariant) SetNestedSlice(s *sliceVariant, fields ...string) error {
	return o.SetNestedValue(s, fields...)
}

func (o *mapVariant) RemoveNestedField(fields ...string) (bool, error) {
	current := o
	n := len(fields)
	for i := 0; i < n; i++ {
		entry, found := current.getVariant(fields[i])
		if !found {
			return false, nil
		}

		if i == n-1 {
			return current.remove(fields[i])
		} else {
			switch entry := entry.(type) {
			case *mapVariant:
				current = entry
			default:
				return false, fmt.Errorf("value is of unexpected type %T", entry)
			}
		}
	}
	return false, fmt.Errorf("unexpected code reached")
}

func (o *mapVariant) getMap(field string, create bool) (*mapVariant, bool, error) {
	node, found := o.getVariant(field)

	if !found {
		if !create {
			return nil, found, nil
		}
		keyNode := buildStringNode(field)
		valueNode := buildMappingNode()
		o.node.Content = append(o.node.Content, keyNode, valueNode)
		valueVariant := &mapVariant{node: valueNode}
		return valueVariant, found, nil
	}

	if node, ok := node.(*mapVariant); ok {
		return node, found, nil
	} else {
		return nil, found, fmt.Errorf("incorrect type, was %T", node)
	}
}

func (o *mapVariant) getSlice(field string, create bool) (*sliceVariant, bool, error) {
	node, found := o.getVariant(field)

	if !found {
		if !create {
			return nil, found, nil
		}
		keyNode := buildStringNode(field)
		valueNode := buildSequenceNode()
		o.node.Content = append(o.node.Content, keyNode, valueNode)
		valueVariant := &sliceVariant{node: valueNode}
		return valueVariant, found, nil
	}

	if node, ok := node.(*sliceVariant); ok {
		return node, found, nil
	} else {
		return nil, found, fmt.Errorf("incorrect type, was %T", node)
	}
}
