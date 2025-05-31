package custom

import (
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/api/filters/imagetag"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// SetAdditionalFieldSpec updates the image in user given fieldPaths. To be deprecated in around a year, to avoid possible invalid fieldPaths.
func SetAdditionalFieldSpec(img *fn.SubObject, objects fn.KubeObjects, addImgFields fn.SliceSubObjects, ctx *fn.Context, count *int) {
	image := NewImageAdaptor(img)
	additionalImageFields := NewFieldSpecSliceAdaptor(addImgFields)

	for i, obj := range objects {
		objRN, err := yaml.Parse(obj.String())
		if err != nil {
			ctx.ResultErr(err.Error(), obj)
		}
		filter := imagetag.Filter{
			ImageTag: image,
			FsSlice:  additionalImageFields,
		}
		filter.WithMutationTracker(logResultCallback(count))
		err = filtersutil.ApplyToJSON(filter, objRN)
		if err != nil {
			ctx.ResultErr(err.Error(), obj)
		}
		newObj, err := fn.ParseKubeObject([]byte(objRN.MustString()))
		if err != nil {
			ctx.ResultErr(err.Error(), obj)
		}
		*objects[i] = *newObj
	}
}

func logResultCallback(count *int) func(key, value, tag string, node *yaml.RNode) {
	return func(key, value, tag string, node *yaml.RNode) {
		*count += 1
	}
}

// NewImageAdaptor transforms the image struct inside transformer to the struct inside kustomize
func NewImageAdaptor(imgObj *fn.SubObject) types.Image {
	imgPtr := &types.Image{}
	imgObj.AsOrDie(imgPtr)
	return *imgPtr
}

// NewFieldSpecSliceAdaptor transforms the additionalImageFields struct inside transformer to the struct inside kustomize
func NewFieldSpecSliceAdaptor(addImgFields fn.SliceSubObjects) types.FsSlice {
	additionalImageFields := types.FsSlice{}
	for _, v := range addImgFields {
		fieldPtr := &types.FieldSpec{}
		v.AsOrDie(fieldPtr)
		additionalImageFields = append(additionalImageFields, *fieldPtr)
	}
	return additionalImageFields
}
