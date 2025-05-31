package generator

import (
	"fmt"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/native-config-adaptor/augeasclient"
	generator "github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/native-config-adaptor/generator"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	corev1 "k8s.io/api/core/v1"
)

type BuiltinGeneratorKind string

const NativeConfigAdaptor BuiltinGeneratorKind = "NativeConfigAdaptor"

var BuiltinGenerators = map[BuiltinGeneratorKind]fn.Runner{
	NativeConfigAdaptor: nil,
	// ... others
}

var _ fn.Generator = &ConfigMapGenerator{}

type ConfigMapGenerator struct {
	Spec ConfigMapGeneratorSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type ConfigMapGeneratorSpec struct {
	Source []*SourceObjectReference `json:"source,omitempty" yaml:"source,omitempty"`
}

type SourceObjectReference struct {
	corev1.TypedLocalObjectReference `json:",inline,omitempty" yaml:",inline,omitempty"`
	augeasclient.AugeasConfigSource  `json:",inline,omitempty" yaml:",inline,omitempty"`
}

func (r *ConfigMapGenerator) Generate(ctx *fn.Context, functionConfig *fn.KubeObject, items fn.KubeObjects) fn.KubeObjects {
	var newObjects []*fn.KubeObject
	data := map[string]string{}
	rawConfigMapName := ""
	var nativeFnConfig *fn.KubeObject
	for _, source := range r.Spec.Source {
		if source.LocalFile == "" && source.LocalFileRef == "" {
			ctx.ResultErrAndDie("required either `spec.source.localFilePath` or `spec.source.localFileRef`", functionConfig)
		}
		// configmap generator always expects a ConfigMap object from NativeConfigAdaptor, not custom typed object.
		source.AugeasConfigSource.AsConfigMap = true
		if source.Kind != "" {
			if source.Kind != string(NativeConfigAdaptor) {
				ctx.ResultWarn(fmt.Sprintf("unknown generator type %s", r.Spec.Source), functionConfig)
				return nil
			}
			nativeFnConfig = SelectNativeFnConfig(ctx, source, items)
			if nativeFnConfig == nil {
				continue
			}
		} else {
			var err error
			nativeFnConfig, err = generator.NewFromSource(&source.AugeasConfigSource)
			if err != nil {
				ctx.ResultErrAndDie(fmt.Sprintf("bad source: %v", err), functionConfig)
			}
		}
		nativeFnConfig.SetName(functionConfig.GetName())
		nativeConfigAdaptor := generator.NewNativeConfigAdaptor(nativeFnConfig)
		nativeConfigObjects := nativeConfigAdaptor.Generate(ctx, nativeFnConfig, items)
		if source.Name != "" {
			nativeFnConfig.SetName(source.Name)
		}
		if nativeConfigObjects == nil || len(nativeConfigObjects) == 0 {
			ctx.ResultErr("no native configuration generated", nativeFnConfig)
			continue
		}
		granularObj := new(fn.KubeObject)
		rawObject := new(fn.KubeObject)
		for _, newObj := range nativeConfigObjects {
			if strings.HasSuffix(newObj.GetName(), "-internal") {
				granularObj = newObj
			} else {
				rawObject = newObj
			}
		}
		// Update ConfigMap data from Raw granular object.
		{
			newData, found, err := granularObj.NestedStringMap("data")
			if !found || err != nil {
				ctx.ResultErrAndDie(err.Error(), granularObj)
				return nil
			}
			for k, v := range newData {
				if _, ok := data[k]; ok {
					ctx.ResultErrAndDie("duplicate key value %v", nil)
				}
				data[k] = v
			}
		}
		// RawObject is exposed to user as immutable ConfigMap.
		{
			SetRawConfigMapObject(functionConfig, rawObject)
			newObjects = append(newObjects, rawObject)
			rawConfigMapName = rawObject.GetName()
		}
	}
	cmObject, err := r.CreateConfigMap(functionConfig, data)
	if err != nil {
		ctx.ResultErrAndDie(err.Error(), nil)
	}
	newObjects = append(newObjects, cmObject)
	UpdateConfigMapReference(items, functionConfig.GetName(), rawConfigMapName)
	return newObjects
}

func SetRawConfigMapObject(fnConfig *fn.KubeObject, object *fn.KubeObject) {
	hashsuffix, _ := Hash(object)
	hashedName := fmt.Sprintf("%s-%s", fnConfig.GetName(), hashsuffix)
	object.SetName(hashedName)
	object.SetAnnotation(fn.GeneratorIdentifier, fnConfig.GetOriginId().String())
	object.SetNestedBool(true, "immutable")
}

func SelectNativeFnConfig(ctx *fn.Context, source *SourceObjectReference, items fn.KubeObjects) *fn.KubeObject {
	nativeFnConfigs := items.Where(func(o *fn.KubeObject) bool {
		switch true {
		case o.GetKind() != source.Kind:
			return false
		case o.GetName() != source.Name:
			return false
		default:
			return true
		}
	})
	if len(nativeFnConfigs) > 1 {
		ctx.ResultWarn(fmt.Sprintf("found more than one %v", source.Kind), nativeFnConfigs[0])
		return nil
	}
	if len(nativeFnConfigs) == 0 {
		ctx.ResultWarn(fmt.Sprintf("unable to find source %v", source.Kind), nil)
		return nil
	}
	return nativeFnConfigs[0]
}

func UpdateConfigMapReference(items fn.KubeObjects, oldName, newName string) {
	for _, object := range items {
		if object.IsGVK("apps", "v1", "StatefulSet") {
			volumes := object.GetMap("spec").GetMap("template").GetMap("spec").GetSlice("volumes")
			for _, volume := range volumes {
				var v corev1.Volume
				volume.As(&v)
				if v.ConfigMap != nil && v.ConfigMap.Name == oldName {
					volume.GetMap("configMap").SetNestedString(newName, "name")
				}
			}
		}
	}
}

func (r *ConfigMapGenerator) CreateConfigMap(functionConfig *fn.KubeObject, data map[string]string) (*fn.KubeObject, error) {
	name := functionConfig.GetName() + "-internal"
	namespace := functionConfig.GetNamespace()

	object := fn.NewEmptyKubeObject()
	object.SetKind("ConfigMap")
	object.SetAPIVersion("v1")
	object.SetName(name)
	if namespace != "" {
		object.SetNamespace(namespace)
	}
	object.SetNestedStringMap(data, "data")
	hashsuffix, err := Hash(object)
	if err != nil {
		return nil, err
	}
	hashedName := fmt.Sprintf("%s-%s", name, hashsuffix)
	object.SetName(hashedName)
	object.SetAnnotation(fn.IndexAnnotation, "0")
	object.SetAnnotation(fn.PathAnnotation, object.GetKind()+"_"+object.GetName()+".yaml")
	object.SetAnnotation(fn.GeneratorBuiltinIdentifier, functionConfig.GetOriginId().String())
	return object, nil
}
