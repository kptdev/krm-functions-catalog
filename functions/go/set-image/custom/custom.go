// Copyright 2026 The kpt Authors
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
package custom

import (
	"github.com/kptdev/krm-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/api/filters/imagetag"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// SetAdditionalFieldSpec updates the image in user given fieldPaths. To be deprecated in around a year, to avoid possible invalid fieldPaths.
func SetAdditionalFieldSpec(img *fn.SubObject, objects fn.KubeObjects, addImgFields fn.SliceSubObjects, res *fn.Results, count *int) {
	image := NewImageAdaptor(img)
	additionalImageFields := NewFieldSpecSliceAdaptor(addImgFields)

	for i, obj := range objects {
		objRN, err := yaml.Parse(obj.String())
		if err != nil {
			res.Errorf(err.Error(), obj)
		}
		filter := imagetag.Filter{
			ImageTag: image,
			FsSlice:  additionalImageFields,
		}
		filter.WithMutationTracker(LogResultCallback(count))
		err = filtersutil.ApplyToJSON(filter, objRN)
		if err != nil {
			res.Errorf(err.Error(), obj)
		}
		newObj, err := fn.ParseKubeObject([]byte(objRN.MustString()))
		if err != nil {
			res.Errorf(err.Error(), obj)
		}
		objects[i] = newObj
	}
}

func LogResultCallback(count *int) func(key, value, tag string, node *yaml.RNode) {
	return func(key, value, tag string, node *yaml.RNode) {
		*count += 1
	}
}

// NewImageAdaptor transforms the image struct inside transformer to the struct inside kustomize
func NewImageAdaptor(imgObj *fn.SubObject) types.Image {
	imgPtr := &types.Image{}
	// nolint
	imgObj.As(imgPtr)
	return *imgPtr
}

// NewFieldSpecSliceAdaptor transforms the additionalImageFields struct inside transformer to the struct inside kustomize
func NewFieldSpecSliceAdaptor(addImgFields fn.SliceSubObjects) types.FsSlice {
	additionalImageFields := types.FsSlice{}
	for _, v := range addImgFields {
		fieldPtr := &types.FieldSpec{}
		// nolint
		v.As(fieldPtr)
		additionalImageFields = append(additionalImageFields, *fieldPtr)
	}
	return additionalImageFields
}
