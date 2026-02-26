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
package transformer

import (
	"fmt"

	"github.com/kptdev/krm-functions-catalog/functions/go/set-image/custom"
	"github.com/kptdev/krm-functions-sdk/go/fn"
	"sigs.k8s.io/kustomize/api/filters/imagetag"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
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

// TODO: is there a nicer way to do this?
var containersFsSlice = func() types.FsSlice {
	out := types.FsSlice{
		{
			Gvk: resid.Gvk{
				Group:   "",
				Kind:    "Pod",
				Version: "v1",
			},
			Path: "spec/containers[]/image",
		},
		{
			Gvk: resid.Gvk{
				Group:   "",
				Kind:    "Pod",
				Version: "v1",
			},
			Path: "spec/initContainers[]/image",
		},
	}

	// TODO: is PodTemplate a kind?
	templateKinds := []string{"Deployment", "StatefulSet", "ReplicaSet", "DaemonSet"}

	for _, kind := range templateKinds {
		out, _ = out.MergeAll(types.FsSlice{
			{
				Gvk: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    kind,
				},
				Path: "spec/template/spec/containers[]/image",
			},
			{
				Gvk: resid.Gvk{
					Group:   "apps",
					Version: "v1",
					Kind:    kind,
				},
				Path: "spec/template/spec/initContainers[]/image",
			},
		})
	}

	return out
}()

// SetImage supports the set-image workflow, it uses Config to parse functionConfig, Transform to change the image
type SetImage struct {
	// Image is the desired image
	Image types.Image `json:"image,omitempty" yaml:"image,omitempty"`
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
		if err = t.updateContainerImages(o); err != nil {
			res.Errorf(err.Error(), o)
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
func (t *SetImage) updateContainerImages(obj *fn.KubeObject) error {
	filter := imagetag.Filter{
		ImageTag: t.Image,
		FsSlice:  containersFsSlice,
	}
	filter.WithMutationTracker(custom.LogResultCallback(&t.resultCount))

	objRN, err := yaml.Parse(obj.String())
	if err != nil {
		return err
	}

	if err := filtersutil.ApplyToJSON(filter, objRN); err != nil {
		return err
	}

	newObj, err := fn.ParseKubeObject([]byte(objRN.MustString()))
	if err != nil {
		return err
	}

	*obj = *newObj

	return nil
}
