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
package main

import (
	"fmt"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/api/filters/prefixsuffix"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

const (
	fnConfigGroup       = "fn.kpt.dev"
	fnConfigVersion     = "v1alpha1"
	fnConfigAPIVersion  = fnConfigGroup + "/" + fnConfigVersion
	fnConfigKind        = "EnsureNameSubstring"
	dependsOnAnnotation = "config.kubernetes.io/depends-on"
)

var (
	// Assumes alphanumeric characters, '-', or '.'
	// <group>/namespaces/<namespace>/<kind>/<name>
	namespacedResourcePattern = regexp.MustCompile(`\A([-.\w]*)/namespaces/([-.\w]*)/([-.\w]*)/([-.\w]*)\z`)
	// <group>/<kind>/<name>
	clusterScopedResourcePattern = regexp.MustCompile(`\A([-.\w]*)/([-.\w]*)/([-.\w]*)\z`)
)

type EnsureNameSubstring struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	// Substring is the desired name substring.
	Substring string `json:"substring" yaml:"substring"`
	// EditMode controls the desired action when the desired substring is not found in the name.
	// If not specified, prepend will be the default.
	EditMode EditMode `json:"editMode,omitempty" yaml:"editMode,omitempty"`
	// FieldSpecs is deprecated, please use AdditionalNameFields instead.
	FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
	// AdditionalNameFields is used to specify additional fields to modify name.
	AdditionalNameFields []types.FieldSpec `json:"additionalNameFields,omitempty" yaml:"additionalNameFields,omitempty"`
	// inputResourceLookup is used internally to track input resources
	inputResourceLookup resourceLookup
}

type EditMode string

const (
	Prepend EditMode = "prepend"
	Append  EditMode = "append"
)

var _ framework.Defaulter = &EnsureNameSubstring{}

func (ens *EnsureNameSubstring) Default() error {
	if ens.EditMode == "" {
		ens.EditMode = Prepend
	}
	if ens.AdditionalNameFields == nil && ens.FieldSpecs != nil {
		ens.AdditionalNameFields = ens.FieldSpecs
		ens.FieldSpecs = nil
	}
	return nil
}

var _ framework.Validator = &EnsureNameSubstring{}

func (ens *EnsureNameSubstring) Validate() error {
	if len(ens.Substring) == 0 {
		return fmt.Errorf("substring must not be empty")
	}

	if ens.AdditionalNameFields != nil && ens.FieldSpecs != nil {
		return fmt.Errorf("`fieldSpecs` has been deprecated, please rename it to `additionalNameFields`")
	}
	return nil
}

func (ens *EnsureNameSubstring) Transform(m resmap.ResMap) error {
	ens.inputResourceLookup.FromResMap(m)
	for _, r := range m.Resources() {
		if shouldSkip(r.OrgId()) {
			continue
		}
		id := r.OrgId()
		// current default configuration contains
		// only one entry: "metadata/name" with no GVK
		for _, fs := range ens.AdditionalNameFields {
			if !id.IsSelected(&fs.Gvk) {
				continue
			}

			// Idempotency check: if the substring is already part of the name, we
			// don't need to do anything.
			hasSubstring, err := resourceContainsSubstring(r, ens.Substring, fs)
			if err != nil {
				return err
			}
			if hasSubstring {
				continue
			}

			if isNameChange(&fs) {
				// If we are changing "metadata/name", we tracks the original
				// name and the prefix or suffix being added.
				r.StorePreviousId()
				switch ens.EditMode {
				case Prepend:
					r.AddNamePrefix(ens.Substring)
				case Append:
					r.AddNameSuffix(ens.Substring)
				}
			}

			fltr := prefixsuffix.Filter{
				FieldSpec: fs,
			}
			switch ens.EditMode {
			case Prepend:
				fltr.Prefix = ens.Substring
			case Append:
				fltr.Suffix = ens.Substring
			}
			err = r.ApplyFilter(fltr)
			if err != nil {
				return err
			}
			if err = ens.updateDependsOnAnnotation(r); err != nil {
				return err
			}
		}
	}
	return nil
}

var _ yaml.Unmarshaler = &EnsureNameSubstring{}

func (ens *EnsureNameSubstring) UnmarshalYAML(value *yaml.Node) error {
	rn := yaml.NewRNode(value)
	meta, err := rn.GetValidatedMetadata()
	if err != nil {
		return err
	}
	s, err := rn.String()
	if err != nil {
		return err
	}

	switch {
	case meta.APIVersion == "v1" && meta.Kind == "ConfigMap":
		cm := corev1.ConfigMap{}
		if err = k8syaml.Unmarshal([]byte(s), &cm); err != nil {
			return err
		}
		if err = configMapToEnsureNameSubstring(&cm, ens); err != nil {
			return err
		}
	case meta.APIVersion == fnConfigAPIVersion && meta.Kind == fnConfigKind:
		if err = k8syaml.Unmarshal([]byte(s), ens); err != nil {
			return err
		}
	default:
		return fmt.Errorf("function config must be either %v or %v, but we got: %v",
			schema.FromAPIVersionAndKind("v1", "ConfigMap").String(),
			schema.FromAPIVersionAndKind(fnConfigAPIVersion, fnConfigKind).String(),
			schema.FromAPIVersionAndKind(meta.APIVersion, meta.Kind).String())
	}
	return nil
}

// set name substring if input matches one of the following:
// - namespaced resource:     <group>/namespaces/<namespace>/<kind>/<name>
// - cluster-scoped resource: <group>/<kind>/<name>
// returns (string, bool) where the bool indicates if namespace was set
func (ens *EnsureNameSubstring) setDependsOnNameSubstring(dependsOn string) (string, bool) {
	var groupIdx, kindIdx, nameIdx int
	if namespacedResourcePattern.MatchString(dependsOn) {
		groupIdx = 0
		kindIdx = 3
		nameIdx = 4
	} else if clusterScopedResourcePattern.MatchString(dependsOn) {
		groupIdx = 0
		kindIdx = 1
		nameIdx = 2
	} else {
		return dependsOn, false
	}
	segments := strings.Split(dependsOn, "/")
	rk := resourceKey{
		Group: segments[groupIdx],
		Kind:  segments[kindIdx],
		Name:  segments[nameIdx],
	}
	if !ens.inputResourceLookup.HasResource(rk) {
		return dependsOn, false
	}
	switch ens.EditMode {
	case Prepend:
		segments[nameIdx] = ens.Substring + segments[nameIdx]
	case Append:
		segments[nameIdx] = segments[nameIdx] + ens.Substring
	default:
		return dependsOn, false
	}
	dependsOn = strings.Join(segments, "/")
	return dependsOn, true
}

// updateDependsOnAnnotation updates the name for the depends-on annotation.
// The expected syntax is one of the following:
// - namespaced resource:     <group>/namespaces/<namespace>/<kind>/<name>
// - cluster-scoped resource: <group>/<kind>/<name>
func (ens *EnsureNameSubstring) updateDependsOnAnnotation(r *resource.Resource) error {
	annotations := r.GetAnnotations()
	dependsOn, ok := annotations[dependsOnAnnotation]
	if !ok {
		return nil
	}
	dependsOn, ok = ens.setDependsOnNameSubstring(dependsOn)
	if !ok {
		return nil
	}
	annotations[dependsOnAnnotation] = dependsOn
	if err := r.SetAnnotations(annotations); err != nil {
		return err
	}
	return nil
}

func configMapToEnsureNameSubstring(cm *corev1.ConfigMap, ens *EnsureNameSubstring) error {
	if len(cm.Data) != 1 {
		return fmt.Errorf("only 1 entry is allowed in the ConfigMap, but got: %d", len(cm.Data))
	}
	for k, v := range cm.Data {
		switch k {
		case string(Prepend):
			ens.EditMode = Prepend
		case string(Append):
			ens.EditMode = Append
		default:
			return fmt.Errorf("unknown editMode: %v, only %v and %v are allowed", k, Prepend, Append)
		}
		ens.Substring = v
	}
	return nil
}

// A Gvk skip list for prefix/suffix modification.
// hard coded for now - eventually should be part of config.
var prefixSuffixFieldSpecsToSkip = types.FsSlice{
	{Gvk: resid.Gvk{Kind: "CustomResourceDefinition"}},
	{Gvk: resid.Gvk{Group: "apiregistration.k8s.io", Kind: "APIService"}},
	{Gvk: resid.Gvk{Kind: "Namespace"}},
	{Gvk: resid.Gvk{Group: "kpt.dev", Kind: "Kptfile"}},
}

func shouldSkip(id resid.ResId) bool {
	for _, path := range prefixSuffixFieldSpecsToSkip {
		if id.IsSelected(&path.Gvk) {
			return true
		}
	}
	return false
}

func isNameChange(fs *types.FieldSpec) bool {
	return fs.Path == "metadata/name"
}

func resourceContainsSubstring(r *resource.Resource, substring string, fs types.FieldSpec) (bool, error) {
	if isNameChange(&fs) {
		// Idempotency check: if the substring is already part of the name, we
		// don't need to do anything.
		return strings.Contains(r.GetName(), substring), nil
	}

	m, err := r.Map()
	if err != nil {
		return false, fmt.Errorf("unable to convert resource for %v: %w", r.OrgId().String(), err)
	}
	rn, err := yaml.FromMap(m)
	if err != nil {
		return false, fmt.Errorf("unable to check if the substring exsits in %v: %w", r.OrgId().String(), err)
	}
	pathElements := strings.Split(fs.Path, "/")
	val, err := rn.Pipe(yaml.Lookup(pathElements...))
	if err != nil {
		return false, fmt.Errorf("unable to lookup path %v in %v: %w", fs.Path, r.OrgId().String(), err)
	}
	valStr, err := val.String()
	if err != nil {
		return false, fmt.Errorf("unable to check if the substring exsits in %v: %w", r.OrgId().String(), err)
	}
	return strings.Contains(valStr, substring), nil
}

// resourceKey provides a key for looking up resources in resourceLookup
type resourceKey struct {
	Group string
	Kind  string
	Name  string
}

// resourceLookup provides an API for tracking resources
type resourceLookup struct {
	resourceMap map[resourceKey]bool
}

// FromResMap reads through a ResMap to populate ResourceLookup
func (rl *resourceLookup) FromResMap(m resmap.ResMap) {
	rl.resourceMap = make(map[resourceKey]bool)
	for _, r := range m.Resources() {
		gvk := r.GetGvk()
		rl.resourceMap[resourceKey{
			Group: gvk.Group,
			Kind:  gvk.Kind,
			Name:  r.GetName(),
		}] = true
	}
}

// HasResource returns whether a Resource with the provided resourceKey is present
func (rl *resourceLookup) HasResource(rk resourceKey) bool {
	return rl.resourceMap[rk]
}
