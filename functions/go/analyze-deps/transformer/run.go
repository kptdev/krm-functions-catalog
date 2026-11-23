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
package transformer

import (
	"fmt"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Run reports missing resources (or dependencies) for given set of resources.
func Run(rl *fn.ResourceList) (bool, error) {
	tc := AnalyzeDeps{}
	err := tc.Config(rl.FunctionConfig)
	if err != nil {
		rl.Results = append(rl.Results, fn.ErrorConfigObjectResult(err, rl.FunctionConfig))
		return true, nil
	}
	// Analyze
	results := tc.Analyze(rl.Items)
	rl.Results = append(rl.Results, results...)
	return true, nil
}

// TODO: Define TypeMeta and ObjectMeta in go/fn module. This types could be merged with fn.ResourceIdentifier.
type TypeMeta struct {
	// APIVersion is the apiVersion field of a Resource
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	// Kind is the kind field of a Resource
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

type ObjectMeta struct {
	// Name is the metadata.name field of a Resource
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Namespace is the metadata.namespace field of a Resource
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	// Labels is the metadata.labels field of a Resource
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	// Annotations is the metadata.annotations field of a Resource.
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

// SetNamespace defines structs to parse KRM resource "SetNamespace" (the custom function config) and "ConfigMap" data.
// it provides the method "Config" to read the function configs from ResourceList.FunctionConfig
// it provides the method "Transform" to change the "namespace" and update the "config.kubernetes.io/depends-on" annotation.
type AnalyzeDeps struct {
	TypeMeta   `json:",inline" yaml:",inline"`
	ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Config parses the given kubeObject into functionConfig for the analyzer.
func (p *AnalyzeDeps) Config(o *fn.KubeObject) error {
	switch {
	case o.IsEmpty():
	case o.IsGVK("", "v1", "ConfigMap"):
		var cm corev1.ConfigMap
		o.AsOrDie(&cm)
	case o.IsGVK(fnConfigGroup, fnConfigVersion, fnConfigKind):
		o.AsOrDie(&p)
	default:
		return fmt.Errorf("unknown functionConfig Kind=%v ApiVersion=%v, expect `ConfigMap.v1` or `%s.%s.%s`",
			o.GetKind(), o.GetAPIVersion(), fnConfigKind, fnConfigVersion, fnConfigGroup)
	}
	return nil
}

// Analyze the dependencies.
func (p *AnalyzeDeps) Analyze(objects fn.KubeObjects) fn.Results {
	var results fn.Results

	resourcesById := map[fn.ResourceIdentifier]*fn.KubeObject{}
	// refs points to references needed by a given resource.
	// There could be multiple resources referring to same resource, but for now,
	// we just keep the first one here.
	refs := map[fn.ResourceIdentifier]*fn.KubeObject{}

	// Skip local configs
	objects = objects.WhereNot(func(o *fn.KubeObject) bool { return o.IsLocalConfig() })

	for _, o := range objects {
		id := o.GetId()
		resourcesById[*id] = o
		if deps := analyzeDeps(o); len(deps) > 0 {
			for _, d := range deps {
				refs[d] = o
			}
		}
	}
	// identify missing dependencies
	for ref, referredBy := range refs {
		if _, exists := resourcesById[ref]; !exists {
			results = append(results,
				// Throwing a warning instead of error
				// because the use case is for informational purpose.
				&fn.Result{
					Message:  fmt.Sprintf("resource not found in the package, but referred by %q.", idToStr(referredBy.GetId())),
					Severity: fn.Warning,
					ResourceRef: &yaml.ResourceIdentifier{
						TypeMeta: yaml.TypeMeta{
							APIVersion: ref.Group + "/" + ref.Version,
							Kind:       ref.Kind,
						},
						NameMeta: yaml.NameMeta{
							Name:      ref.Name,
							Namespace: ref.Namespace,
						},
					},
					File: &fn.File{
						Path:  referredBy.PathAnnotation(),
						Index: referredBy.IndexAnnotation(),
					},
				})
		}
	}

	return results
}

func analyzeDeps(o *fn.KubeObject) (deps []fn.ResourceIdentifier) {
	myID := o.GetId()
	if o.IsNamespaceScoped() {
		// if this is namespace scoped, it requires namespace to exist
		deps = append(deps, fn.ResourceIdentifier{
			Version: "v1",
			Kind:    "Namespace",
			Name:    myID.Namespace,
			// Namespace: fn.UnknownNamespace,
		})
	}

	switch {
	case o.IsGVK("apps", "v1", "Deployment"), o.IsGVK("apps", "v1", "StatefulSet"):

		var containers []corev1.Container

		err := o.GetMap("spec").GetMap("template").GetMap("spec").GetMap("containers").As(&containers)
		if err != nil {
			fn.Logf("containers missing in the deployment: %s \n", o.GetId())
		}

		for _, c := range containers {
			for _, env := range c.Env {
				if env.ValueFrom == nil {
					continue
				}
				// the deployment refers to a secret object
				if env.ValueFrom.SecretKeyRef != nil {
					secretName := env.ValueFrom.SecretKeyRef.Name
					secretID := fn.ResourceIdentifier{
						Version:   "v1",
						Kind:      "Secret",
						Name:      secretName,
						Namespace: myID.Namespace,
					}
					// fn.Logf("detected dependency on secret: %s \n", secretID)
					deps = append(deps, secretID)
				}
				// the deployment refers to a secret object
				if env.ValueFrom.ConfigMapKeyRef != nil {
					cmName := env.ValueFrom.ConfigMapKeyRef.Name
					cmID := fn.ResourceIdentifier{
						Version:   "v1",
						Kind:      "ConfigMap",
						Name:      cmName,
						Namespace: myID.Namespace,
					}
					// fn.Logf("detected dependency on ConfigMap: %s \n", cmID)
					deps = append(deps, cmID)
				}
			}
			for _, envfrom := range c.EnvFrom {
				// the deployment refers to a secret object
				if envfrom.SecretRef != nil {
					secretName := envfrom.SecretRef.Name
					secretID := fn.ResourceIdentifier{
						Version:   "v1",
						Kind:      "Secret",
						Name:      secretName,
						Namespace: myID.Namespace,
					}
					// fn.Logf("detected dependency on secret: %s \n", secretID)
					deps = append(deps, secretID)
				}
				// the deployment refers to a secret object
				if envfrom.ConfigMapRef != nil {
					cmName := envfrom.ConfigMapRef.Name
					cmID := fn.ResourceIdentifier{
						Version:   "v1",
						Kind:      "ConfigMap",
						Name:      cmName,
						Namespace: myID.Namespace,
					}
					// fn.Logf("detected dependency on configmap: %s \n", cmID)
					deps = append(deps, cmID)
				}
			}
			// Analyze volumes for configmap/secret projections
		}
	}
	return
}

func idToStr(identifier *fn.ResourceIdentifier) string {
	var idStringList []string
	if identifier != nil {
		if identifier.Group != "" {
			idStringList = append(idStringList, identifier.Group)
		}
		if identifier.Version != "" {
			idStringList = append(idStringList, identifier.Version)
		}
		if identifier.Kind != "" {
			idStringList = append(idStringList, identifier.Kind)
		}
		if identifier.Namespace != "" {
			idStringList = append(idStringList, identifier.Namespace)
		}
		if identifier.Name != "" {
			idStringList = append(idStringList, identifier.Name)
		}
	}
	return strings.Join(idStringList, "/")
}
