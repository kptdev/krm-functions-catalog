package transformer

import (
	"fmt"

	"github.com/kptdev/krm-functions-catalog/functions/go/set-image/custom"
	"github.com/kptdev/krm-functions-catalog/functions/go/set-image/third_party/sigs.k8s.io/kustomize/api/image"
	"github.com/kptdev/krm-functions-catalog/functions/go/set-image/third_party/sigs.k8s.io/kustomize/api/types"
	"github.com/kptdev/krm-functions-sdk/go/fn"
)

// Image contains an image name, a new name, a new tag or digest, which will replace the original name and tag.
type Image struct {
	// Name is a tag-less image name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// NewName is the value used to replace the original name.
	NewName string `json:"newName,omitempty" yaml:"newName,omitempty"`

	// NewTag is the value used to replace the original tag.
	NewTag string `json:"newTag,omitempty" yaml:"newTag,omitempty"`

	// Digest is the value used to replace the original image tag.
	// If digest is present NewTag value is ignored.
	Digest string `json:"digest,omitempty" yaml:"digest,omitempty"`
}

var _ fn.Runner = &SetImage{}

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
func (t *SetImage) Run(_ *fn.Context, fnConfig *fn.KubeObject, items fn.KubeObjects, res *fn.Results) bool {
	err := t.configDefaultData()
	if err != nil {
		res.Errorf(err.Error(), nil)
	}
	err = t.validateInput()
	if err != nil {
		res.Errorf("invalid FunctionConfig: %v", err)
	}

	for _, o := range items {
		switch o.GetKind() {
		case "Pod":
			if err = t.setPodContainers(o); err != nil {
				res.Errorf(err.Error(), o)
			}
		case "Deployment", "StatefulSet", "ReplicaSet", "DaemonSet", "PodTemplate":
			if err = t.setPodSpecContainers(o); err != nil {
				res.Errorf(err.Error(), o)
			}
		}
	}

	if t.AdditionalImageFields != nil {
		custom.SetAdditionalFieldSpec(fnConfig.GetMap("image"), items, fnConfig.GetSlice("additionalImageFields"), res, &t.resultCount)
	}

	summary := fmt.Sprintf("summary: updated a total of %v image(s)", t.resultCount)
	res.Infof("%s", summary)
	return res.ExitCode() != 1
}

// configDefaultData transforms the data from ConfigMap to SetImage struct
func (t *SetImage) configDefaultData() error {
	for key, val := range t.DataFromDefaultConfig {
		switch key {
		case "name":
			t.Image.Name = val
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
func (t *SetImage) validateInput() error {
	// TODO: support container name and only one argument input in the next PR
	if t.Image.Name == "" {
		return fmt.Errorf("must specify `name`")
	}
	if t.Image.NewName == "" && t.Image.NewTag == "" && t.Image.Digest == "" {
		return fmt.Errorf("must specify one of `newName`, `newTag`, or `digest`")
	}
	return nil
}

// updateContainerImages updates the images inside containers, return potential error
func (t *SetImage) updateContainerImages(pod *fn.SubObject) error {
	var containers fn.SliceSubObjects
	containers = append(containers, pod.GetSlice("iniContainers")...)
	containers = append(containers, pod.GetSlice("containers")...)

	for _, o := range containers {
		oldValue, _, err := o.NestedString("image")
		if err != nil {
			return err
		}
		if !image.IsImageMatched(oldValue, t.Image.Name) {
			continue
		}
		newName := getNewImageName(oldValue, t.Image)
		if oldValue == newName {
			continue
		}

		if err := o.SetNestedString(newName, "image"); err != nil {
			return err
		}
		t.resultCount += 1
	}
	return nil
}

func (t *SetImage) setPodSpecContainers(o *fn.KubeObject) error {
	spec := o.GetMap("spec")
	if spec == nil {
		return nil
	}
	template := spec.GetMap("template")
	if template == nil {
		return nil
	}
	podSpec := template.GetMap("spec")
	err := t.updateContainerImages(podSpec)
	if err != nil {
		return err
	}
	return nil
}

func (t *SetImage) setPodContainers(o *fn.KubeObject) error {
	spec := o.GetMap("spec")
	if spec == nil {
		return nil
	}
	err := t.updateContainerImages(spec)
	if err != nil {
		return err
	}
	return nil
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
