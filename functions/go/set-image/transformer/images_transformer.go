package transformer

import (
	"fmt"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-image/custom"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-image/third_party/sigs.k8s.io/kustomize/api/image"
	"github.com/GoogleContainerTools/kpt-functions-catalog/functions/go/set-image/third_party/sigs.k8s.io/kustomize/api/types"
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
)

type Image struct {
	// DEPRECATED
	//Name is a tag-less image name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// ContainerName is the Pod's container name. It is used to choose the matching containers to update the images.
	ContainerName string `json:"containerName,omitempty" yaml:"containerName,omitempty"`

	// ImageName is the image name. It is used to choose images for update.
	ImageName string `json:"imageName,omitempty" yaml:"imageName,omitempty"`

	// NewName is the new image name
	NewName string `json:"newName,omitempty" yaml:"newName,omitempty"`

	// NewTag is the value used to replace the original tag.
	NewTag string `json:"newTag,omitempty" yaml:"newTag,omitempty"`

	// Digest is the value used to replace the original image tag.
	// If digest is present NewTag value is ignored.
	Digest string `json:"digest,omitempty" yaml:"digest,omitempty"`
}

// SetImage supports the set-image workflow, it uses Config to parse functionConfig, Transform to change the image
type SetImage struct {
	// Image is the desired image
	Image Image `json:"image,omitempty" yaml:"image,omitempty"`
	// ConfigMap keeps the data field that holds image information
	DataFromDefaultConfig map[string]string `json:"data,omitempty" yaml:"data,omitempty"`
	// ONLY for kustomize, AdditionalImageFields is the user supplied fieldspec
	AdditionalImageFields types.FsSlice `json:"additionalImageFields,omitempty" yaml:"additionalImageFields,omitempty"`
	// resultCount logs the total count image change
	resultCount int
}

// Run implements the Runner interface that transforms the resource and log the results
func (t SetImage) Run(ctx *fn.Context, functionConfig *fn.KubeObject, items fn.KubeObjects) {
	err := t.configDefaultData()
	if err != nil {
		ctx.ResultErrAndDie(err.Error(), nil)
	}

	items = items.WhereNot(fn.IsLocalConfig)

	res := t.validateInput(items)
	if res != nil {
		if res.Severity == fn.Error {
			ctx.ResultErrAndDie(res.Message, nil)
		}
		ctx.Result(res.Message, res.Severity)
		return
	}

	for _, o := range items {
		containers := getContainers(o)
		warnings, err := t.updateContainerImages(containers)

		for _, w := range warnings {
			ctx.ResultWarn(w.Message, o)
		}

		if err != nil {
			ctx.ResultErr(err.Error(), o)
		}
	}

	if t.AdditionalImageFields != nil {
		custom.SetAdditionalFieldSpec(functionConfig.GetMap("image"), items, functionConfig.GetSlice("additionalImageFields"), ctx, &t.resultCount)
	}

	summary := fmt.Sprintf("summary: updated a total of %v image(s)", t.resultCount)
	ctx.ResultInfo(summary, nil)
}

// configDefaultData transforms the data from ConfigMap to SetImage struct
func (t *SetImage) configDefaultData() error {
	for key, val := range t.DataFromDefaultConfig {
		switch key {
		case "name":
			t.Image.ImageName = val
		case "containerName":
			t.Image.ContainerName = val
		case "imageName":
			t.Image.ImageName = val
		case "newName":
			t.Image.NewName = val
		case "newTag":
			t.Image.NewTag = val
		case "digest":
			t.Image.Digest = val
		default:
			return fmt.Errorf("ConfigMap has wrong field name %v", key)
		}
	}
	return nil
}

// validateInput validates the inputs passed into via the functionConfig
func (t *SetImage) validateInput(items fn.KubeObjects) *fn.Result {
	// if user does not input image name or container name, there should be only one image name to select from
	if t.Image.Name == "" && t.Image.ImageName == "" && t.Image.ContainerName == "" {
		if !t.isImageNameUnique(items) {
			return fn.GeneralResult("must specify `imageName`, resources contain non-unique image names", fn.Error)
		}
	}
	if t.Image.Name != "" && t.Image.ImageName != "" && t.Image.Name != t.Image.ImageName {
		return fn.GeneralResult("must not fill `imageName` and `name` at same time, their values should be equal", fn.Error)
	}
	if t.Image.NewName == "" && t.Image.NewTag == "" && t.Image.Digest == "" {
		return fn.GeneralResult("must specify one of `newName`, `newTag`, or `digest`", fn.Error)
	}
	if len(items) == 0 {
		return fn.GeneralResult("no input resources", fn.Info)
	}
	return nil
}

// matchImage takes the resources image name and container name, return if there is a match and potential warning
func matchImage(oldImageName string, oldContainer string, newImage *Image) (bool, *fn.Result) {
	// name would be deprecated, means image name
	if newImage.Name != "" {
		newImage.ImageName = newImage.Name
	}

	// match image name, no container name
	if newImage.ImageName != "" && newImage.ContainerName == "" {
		if !image.IsImageMatched(oldImageName, newImage.ImageName) {
			return false, nil
		}
	}
	// match container name, no image name
	if newImage.ImageName == "" && newImage.ContainerName != "" {
		if oldContainer != newImage.ContainerName {
			return false, nil
		}
	}
	// match both
	if newImage.ImageName != "" && newImage.ContainerName != "" {
		if !image.IsImageMatched(oldImageName, newImage.ImageName) {
			return false, nil
		}
		//if only image name match, container name does not match, provide a warning
		if oldContainer != newImage.ContainerName {
			msg := fmt.Sprintf("container name `%v` does not match `%v`, only image name matches", newImage.ContainerName, oldContainer)
			warning := &fn.Result{
				Message:  msg,
				Severity: fn.Warning,
			}
			return true, warning
		}
	}
	return true, nil
}

// isImageNameUnique checks if there is only one image name in all resources, return true if name is unique
func (t *SetImage) isImageNameUnique(items fn.KubeObjects) bool {
	curImageName := ""
	for _, o := range items {
		containers := getContainers(o)
		for _, c := range containers {
			oldValue := c.NestedStringOrDie("image")
			imageName, _, _ := image.Split(oldValue)
			if curImageName == "" {
				curImageName = imageName
			} else if curImageName != imageName {
				return false
			}
		}
	}
	return true
}

// updateContainerImages updates the images inside containers, return warnings and error
func (t *SetImage) updateContainerImages(containers fn.SliceSubObjects) (fn.Results, error) {
	var warningList fn.Results
	for _, o := range containers {
		oldValue := o.NestedStringOrDie("image")
		oldContainer := o.NestedStringOrDie("name")

		matched, warning := matchImage(oldValue, oldContainer, &t.Image)
		if !matched {
			continue
		}

		newName := getNewImageName(oldValue, t.Image)
		if oldValue == newName {
			continue
		}

		if warning != nil {
			warningList = append(warningList, warning)
		}

		if err := o.SetNestedString(newName, "image"); err != nil {
			return warningList, err
		}
		t.resultCount += 1
	}
	return warningList, nil
}

// getContainers gets the containers inside kubeObject
func getContainers(o *fn.KubeObject) fn.SliceSubObjects {
	switch o.GetKind() {
	case "Pod":
		return getPodContainers(o)
	case "Deployment", "StatefulSet", "ReplicaSet", "DaemonSet", "PodTemplate":
		return getPodSpecContainers(o)
	}
	return nil
}

// getPodContainers gets the containers from pod
func getPodContainers(o *fn.KubeObject) fn.SliceSubObjects {
	spec := o.GetMap("spec")
	if spec == nil {
		return nil
	}
	return append(spec.GetSlice("iniContainers"), spec.GetSlice("containers")...)
}

// getPodSpecContainers gets the containers from podSpec
func getPodSpecContainers(o *fn.KubeObject) fn.SliceSubObjects {
	spec := o.GetMap("spec")
	if spec == nil {
		return nil
	}
	template := spec.GetMap("template")
	if template == nil {
		return nil
	}
	podSpec := template.GetMap("spec")
	if podSpec == nil {
		return nil
	}
	return append(podSpec.GetSlice("iniContainers"), podSpec.GetSlice("containers")...)
}

// getNewImageName return the new name for image field
func getNewImageName(oldValue string, newImage Image) string {
	name, tag, digest := image.Split(oldValue)
	if newImage.NewName != "" {
		name = newImage.NewName
	}
	if newImage.NewTag != "" {
		tag = ":" + newImage.NewTag
	}
	if newImage.Digest != "" {
		tag = "@" + newImage.Digest
	}
	var newName string
	if tag == "" {
		newName = name + digest
	} else {
		newName = name + tag
	}

	return newName
}
