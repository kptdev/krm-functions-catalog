package generator

import (
	"fmt"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/configmap-generator/fn"
	generator "github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/configmap-generator/native-config-adaptor/generator"
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
	generator.NativeConfigSource     `json:",inline,omitempty" yaml:",inline,omitempty"`
}

func (r *ConfigMapGenerator) Generate(ctx *fn.Context, functionConfig *fn.KubeObject, items fn.KubeObjects) fn.KubeObjects {
	noConfigMapItemList := items.WhereNot(func(o *fn.KubeObject) bool {
		return o.GetKind() == "ConfigMap" && (o.GetAnnotation(fn.GeneratorIdentifier) == functionConfig.GetId().String() || o.GetAnnotation(fn.GeneratorBuiltinIdentifier) == functionConfig.GetId().String())
	})
	previousGeneratedConfigMapAndNonKRM := items.Where(func(o *fn.KubeObject) bool {
		if o.GetKind() == "ConfigMap" {
			return o.GetAnnotation(fn.GeneratorIdentifier) == functionConfig.GetId().String() || o.GetAnnotation(fn.GeneratorBuiltinIdentifier) == functionConfig.GetId().String()
		}
		if o.GetKind() == fn.NonKrmKind {
			return true
		}
		return false
	})
	rawConfigMapName := ""
	var newlyGeneratedConfigMaps, newItemsList fn.KubeObjects
	data := map[string]string{}
	var nativeFnConfig *fn.KubeObject
	for _, source := range r.Spec.Source {
		if source.LocalFile == "" && source.LocalFileRef == "" {
			ctx.ResultErrAndDie("required either `spec.source.localFilePath` or `spec.source.localFileRef`", functionConfig)
		}
		// configmap generator always expects a ConfigMap object from NativeConfigAdaptor, not custom typed object.
		source.AsConfigMap = true
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
			nativeFnConfig, err = generator.NewFromSource(&source.NativeConfigSource)
			if err != nil {
				ctx.ResultErrAndDie(fmt.Sprintf("bad source: %v", err), functionConfig)
			}
		}
		nativeFnConfig.SetName(functionConfig.GetName())
		nativeConfigAdaptor := generator.NewNativeConfigAdaptor(nativeFnConfig)
		newlyGeneratedConfigMaps = nativeConfigAdaptor.Generate(ctx, nativeFnConfig, previousGeneratedConfigMapAndNonKRM)
		if source.Name != "" {
			nativeFnConfig.SetName(source.Name)
		}
		if newlyGeneratedConfigMaps == nil || len(newlyGeneratedConfigMaps) == 0 {
			ctx.ResultErr("no ConfigMap generated from native-config-adaptor", nativeFnConfig)
			continue
		}
		var rawConfigMap, canonicalConfigMap *fn.KubeObject
		for _, newConfigMap := range newlyGeneratedConfigMaps {
			if strings.HasSuffix(newConfigMap.GetName(), "-internal") {
				canonicalConfigMap = newConfigMap
			} else {
				rawConfigMap = newConfigMap
			}
		}
		// Update ConfigMap data from Raw granular object.
		if canonicalConfigMap != nil {
			newData, found, err := canonicalConfigMap.NestedStringMap("data")
			if !found || err != nil {
				ctx.ResultErrAndDie(err.Error(), canonicalConfigMap)
				return nil
			}
			for k, v := range newData {
				if _, ok := data[k]; ok {
					ctx.ResultErrAndDie("duplicate key value %v", nil)
				}
				data[k] = v
			}
			cmObject, err := r.CreateCanonicalConfigMap(functionConfig, canonicalConfigMap, data)
			if err != nil {
				ctx.ResultErrAndDie(err.Error(), nil)
			}
			newItemsList = append(newItemsList, cmObject)
		}
		// RawObject is exposed to user as immutable ConfigMap.
		{
			SetRawConfigMapObject(functionConfig, rawConfigMap)
			newItemsList = append(newItemsList, rawConfigMap)
			rawConfigMapName = rawConfigMap.GetName()
		}
	}
	UpdateConfigMapReference(noConfigMapItemList, functionConfig.GetName(), rawConfigMapName)
	ctx.ResultInfo("ConfigMap references are updated to "+rawConfigMapName, nil)
	newItemsList = append(newItemsList, noConfigMapItemList...)
	return newItemsList
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
				// StatefulSet should hold the configmap origin.
				// if v.ConfigMap != nil && v.ConfigMap.Name == oldName {
				if v.ConfigMap != nil {
					volume.GetMap("configMap").SetNestedString(newName, "name")
				}
			}
		}
	}
}

func (r *ConfigMapGenerator) CreateCanonicalConfigMap(functionConfig, object *fn.KubeObject, data map[string]string) (*fn.KubeObject, error) {
	name := functionConfig.GetName() + "-internal"
	namespace := functionConfig.GetNamespace()
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
