package krmfn

import (
	"github.com/kptdev/krm-functions-sdk/go/fn"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const ModuleName = "krmfn.star"

var Module = &starlarkstruct.Module{
	Name: "krmfn",
	Members: starlark.StringDict{
		"match_gvk":       starlark.NewBuiltin("match_gvk", matchGVK),
		"match_name":      starlark.NewBuiltin("match_name", matchName),
		"match_namespace": starlark.NewBuiltin("match_namespace", matchNamespace),
	},
}

func matchGVK(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var resourceList starlark.Value
	var apiVersion, kind string
	if err := starlark.UnpackPositionalArgs("match_gvk", args, kwargs, 3,
		&resourceList, &apiVersion, &kind); err != nil {
		return nil, err
	}
	obj, err := fn.ParseKubeObject([]byte(resourceList.String()))
	if err != nil {
		return nil, err
	}
	return starlark.Bool(obj.IsGroupVersionKind(schema.GroupVersionKind{Group: "", Version: apiVersion, Kind: kind})), nil
}

func matchName(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var resourceList starlark.Value
	var name string
	if err := starlark.UnpackPositionalArgs("match_name", args, kwargs, 2, &resourceList, &name); err != nil {
		return nil, err
	}
	obj, err := fn.ParseKubeObject([]byte(resourceList.String()))
	if err != nil {
		return nil, err
	}
	return starlark.Bool(obj.GetName() == name), nil
}

func matchNamespace(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var resourceList starlark.Value
	var namespace string
	if err := starlark.UnpackPositionalArgs("match_namespace", args, kwargs, 2, &resourceList, &namespace); err != nil {
		return nil, err
	}
	obj, err := fn.ParseKubeObject([]byte(resourceList.String()))
	if err != nil {
		return nil, err
	}
	return starlark.Bool(obj.GetNamespace() == namespace), nil
}
