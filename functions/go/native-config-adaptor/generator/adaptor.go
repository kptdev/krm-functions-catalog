// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package generator

import (
	"path/filepath"
	"reflect"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/native-config-adaptor/fn"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/native-config-adaptor/parsers"
)

const NativeConfigOpBackToNative = "writeBackToNativeConfig"

var SupportedFromNative = map[string]func(object *fn.KubeObject, content string, ShouldFlatten bool) error{
	"IniFile": parsers.FromIniFile,
}

var SupportedToNative = map[string]func(object *fn.KubeObject, fileRef, filename string, IsFlatten bool) (*fn.KubeObject, error){
	"IniFile": parsers.ToIniFile,
}

var _ fn.Generator = &NativeConfigAdaptor{}

func NewNativeConfigAdaptor(config *fn.KubeObject) *NativeConfigAdaptor {
	var cfg NativeConfigAdaptor
	config.As(&cfg)
	return &cfg
}

func NewFromSource(source *NativeConfigSource) (*fn.KubeObject, error) {
	nca := &NativeConfigAdaptor{

		Spec: &NativeConfigSpec{
			Source: []*NativeConfigSource{
				source,
			},
		},
	}
	object, err := fn.NewFromTypedObject(nca)
	if err != nil {
		return nil, err
	}
	object.SetKind(reflect.TypeOf(NativeConfigAdaptor{}).Name())
	object.SetAPIVersion(fn.KptFunctionApiVersion)
	return object, nil
}

func (t *NativeConfigAdaptor) Generate(ctx *fn.Context, fnConfig *fn.KubeObject, items fn.KubeObjects) fn.KubeObjects {
	var generatedObjects []*fn.KubeObject
	for _, source := range t.Spec.Source {
		if source.LocalFileRef != "" {
			localfileObjects := items.Where(func(o *fn.KubeObject) bool {
				switch true {
				case o.GetKind() != fn.NonKrmKind:
					return false
				case o.GetAPIVersion() != fn.KptFunctionApiVersion:
					return false
				case o.GetName() != source.LocalFileRef:
					return false
				default:
					return true
				}
			})
			sotObjects := items.Where(func(o *fn.KubeObject) bool {
				return o.GetAnnotation(fn.GeneratorBuiltinIdentifier) != ""
			})

			switch source.Operation {
			case NativeConfigOpBackToNative:
				for _, canonicalObject := range sotObjects {
					nonKrmObject := t.WriteFromCanonicalToNative(ctx, canonicalObject, source)
					existingNonKrmObjects := items.Where(func(o *fn.KubeObject) bool {
						return o.GetId() == nonKrmObject.GetId()
					})
					if len(existingNonKrmObjects) > 0 {
						content := nonKrmObject.GetMap("spec").NestedStringMapOrDie("content")
						existingNonKrmObjects[0].GetMap("spec").SetNestedStringMap(content, "content")
					} else {
						generatedObjects = append(generatedObjects, nonKrmObject)
					}
					rawObject := StoreRawDataInConfigMap(fnConfig.GetName(),
						filepath.Base(source.LocalFile),
						nonKrmObject.NestedStringOrDie("spec", "content"))
					existingrawObjects := items.Where(func(o *fn.KubeObject) bool {
						return o.GetId() == rawObject.GetId()
					})
					if len(existingrawObjects) > 0 {
						existingrawObjects[0].SetNestedStringMapOrDie(rawObject.NestedStringMapOrDie("data"), "data")
					} else {
						generatedObjects = append(generatedObjects, rawObject)
					}
				}
			default:
				for _, object := range localfileObjects {
					content := object.NestedStringOrDie("spec", "content")
					localFileName := object.NestedStringOrDie("spec", "filename")
					newObject := t.ParseFromNativeToCanonical(ctx, object.GetName(), content, source)
					rawObject := StoreRawDataInConfigMap(object.GetName(), localFileName, content)
					generatedObjects = append(generatedObjects, newObject, rawObject)
				}
			}
		}
	}
	return generatedObjects
}

func StoreRawDataInConfigMap(name, filename, content string) *fn.KubeObject {
	object := fn.NewEmptyKubeObject()
	object.SetKind("ConfigMap")
	object.SetAPIVersion("v1")
	object.SetName(name)
	data := map[string]string{
		filename: content,
	}
	object.SetNestedStringMap(data, "data")
	return object
}

func (t *NativeConfigAdaptor) WriteFromCanonicalToNative(ctx *fn.Context, canonicalObject *fn.KubeObject, source *NativeConfigSource) *fn.KubeObject {
	toFn, ok := SupportedToNative[source.Format]
	if !ok {
		ctx.ResultErrAndDie("unknown parsing type "+source.Format, nil)
		return nil
	}
	nonKrmObject, err := toFn(canonicalObject, source.LocalFileRef, source.LocalFile, source.AsConfigMap)
	if err != nil {
		ctx.ResultErrAndDie(err.Error(), canonicalObject)
	}
	return nonKrmObject
}

func (t *NativeConfigAdaptor) ParseFromNativeToCanonical(ctx *fn.Context, name, content string, source *NativeConfigSource) *fn.KubeObject {
	fromFn, ok := SupportedFromNative[source.Format]
	if !ok {
		ctx.ResultErrAndDie("unknown parsing type "+source.Format, nil)
		return nil
	}
	object := NewCanonicalObject(name, source)
	if err := fromFn(object, content, source.AsConfigMap); err != nil {
		ctx.ResultErrAndDie(err.Error(), nil)
	}
	return object
}

func NewCanonicalObject(name string, source *NativeConfigSource) *fn.KubeObject {
	lenseNameSlitted := strings.Split(source.Format, "_")
	camelcaseLense := ""
	for _, segment := range lenseNameSlitted {
		camelcaseLense += strings.ToUpper(string(segment[0])) + segment[1:]
	}

	object := fn.NewEmptyKubeObject()

	if source.AsConfigMap {
		object.SetKind("ConfigMap")
		object.SetAPIVersion("v1")
		object.SetName(name + "-internal")
		object.SetAnnotation(fn.KptLocalConfig, "true")
	} else {
		object.SetKind(camelcaseLense)
		object.SetAPIVersion("config.kpt.dev/v1alpha1")
		object.SetName(name + "-internal")
		object.SetAnnotation(fn.KptLocalConfig, "true")
	}
	return object
}

type NativeConfigAdaptor struct {
	Spec *NativeConfigSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type NativeConfigSpec struct {
	Source []*NativeConfigSource `json:"source,omitempty" yaml:"source,omitempty"`
}

type NativeConfigSource struct {
	LocalFileRef string `json:"localFileRef,omitempty" yaml:"localFileRef,omitempty"`
	LocalFile    string `json:"localFile,omitempty" yaml:"localFile,omitempty"`
	Format       string `json:"format,omitempty" yaml:"format,omitempty"`
	Operation    string `json:"operation,omitempty" yaml:"operation,omitempty"`
	AsConfigMap  bool   `json:"asConfigMap,omitempty" yaml:"asConfigMap,omitempty"`
}
