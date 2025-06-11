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
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/native-config-adaptor/augeasclient"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

var _ fn.Generator = &NativeConfigAdaptor{}

func NewNativeConfigAdaptor(config *fn.KubeObject) *NativeConfigAdaptor {
	var cfg NativeConfigAdaptor
	config.As(&cfg)
	return &cfg
}

func NewFromSource(source *augeasclient.AugeasConfigSource) (*fn.KubeObject, error) {
	nca := &NativeConfigAdaptor{

		Spec: &augeasclient.AugeasConfigSpec{
			Source: []*augeasclient.AugeasConfigSource{
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
	/* TODO: Augeas only reads /etc and some other root access dir. maybe need to write a Augeas "lense" config file.
	tmpDir := "tmp"
	err := os.Mkdir(tmpDir, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	*/
	var nativeConfigObjects []*fn.KubeObject
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

			for _, object := range localfileObjects {
				// HACK: since I cannot let the function executable to write to /etc
				/*
					tmpfn := filepath.Join(tmpDir, object.GetName())
					defer os.Remove(tmpfn)

					content := object.NestedStringOrDie("spec", "content")
					if err := ioutil.WriteFile(tmpfn, []byte(content), 0777); err != nil {
						ctx.ResultErrAndDie(fmt.Sprintf("unable to write content %v: %v", object.GetName(), err), object)
					}
					source.LocalFile = filepath.Join(tmpDir, tmpfn)
				*/
				source.LocalFile = filepath.Join("/etc", object.NestedStringOrDie("spec", "filename"))
				newObjects, err := augeasclient.Generate(source.LocalFile, t.Spec)
				if err != nil {
					ctx.ResultErrAndDie(fmt.Sprintf("unable to generate typed object from Augeas: %v", err), object)
				}
				nativeConfigObjects = append(nativeConfigObjects, newObjects...)
			}
		}
	}
	return nativeConfigObjects
}

type NativeConfigAdaptor struct {
	Spec *augeasclient.AugeasConfigSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}
