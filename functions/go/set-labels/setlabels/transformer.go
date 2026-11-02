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

package setlabels

import (
	"sort"

	"github.com/kptdev/kpt/api/kubeobject"
	schema "github.com/kptdev/kpt/api/schema/v1"
	"github.com/kptdev/krm-functions-sdk/go/fn"
)

type FieldPath []string

var (
	metaLabelsPath  = FieldPath{"metadata", "labels"}
	selectorPath    = FieldPath{"selector", "matchLabels"}
	podSelectorPath = FieldPath{"podSelector", "matchLabels"}
	labelSelector   = FieldPath{"labelSelector", "matchLabels"}
)

var _ fn.Runner = &SetLabels{}

// SetLabels supports the set-labels workflow, it uses Config to parse functionConfig, Transform to change the labels
type SetLabels struct {
	// labels is the desired labels
	Labels map[string]string `json:"labels,omitempty"`
	count  int
}

// EmptyfnConfig is a workaround since kpt creates a FunctionConfig placeholder if users don't provide the functionConfig.
// `kpt fn eval` uses placeholder ConfigMap with name "function-input"
// `kpt fn render` uses placeholder "{}"
func EmptyfnConfig(o *kubeobject.KubeObject) bool {
	if o.GetKind() == "ConfigMap" && o.GetName() == "function-input" {
		data, _, _ := o.NestedStringMap("data")
		return len(data) == 0
	}
	if o.GetKind() == "" && o.GetName() == "" {
		return true
	}
	return false
}

// Transform updates the labels in the right path using GVK filter and other configurable fields
func (p *SetLabels) Run(_ *fn.Context, fnConfig *kubeobject.KubeObject, items kubeobject.KubeObjects, results *kubeobject.Results) bool {
	objects := items.WhereNot(kubeobject.IsLocalConfig)
	if objects.Len() == 0 || objects[0] == nil {
		results.Infof("no input resources")
		return true
	}
	if EmptyfnConfig(fnConfig) {
		return false
	}
	if len(p.Labels) == 0 {
		results.Warningf("no `labels` arguments are given in FunctionConfig")
		return true
	}
	p.count = 0
	for _, o := range objects {
		oErr := func() error {
			if err := p.setLabelsInMeta(&o.SubObject); err != nil {
				return err
			}
			if err := p.setLabelsInSelector(o); err != nil {
				return err
			}
			spec := o.GetMap("spec")
			if spec == nil {
				return nil
			}
			if containPodSpec(o) {
				if err := p.setLabelsInPod(spec.GetMap("template")); err != nil {
					return nil
				}
			}
			if containJobSpec(o) {
				if err := p.setLabelsInJob(spec.GetMap("jobTemplate")); err != nil {
					return err
				}
			}
			if containVolume(o) {
				volumes, _, err := spec.NestedSlice("volumeClaimTemplates")
				if err != nil {
					return err
				}
				if err = p.setLabelsInVolumes(volumes); err != nil {
					return err
				}
			}
			if containNetworkPolicy(o) {
				if err := p.setNetworkPolicyRule(o); err != nil {
					return err
				}
			}
			return nil
		}()
		if oErr != nil {
			results.ErrorE(oErr)
		}
	}
	results.Infof("set %v labels in total", p.count)
	return results.ExitCode() != 1
}

// setLabelsInMeta set ObjectMeta labels for all resources
func (p *SetLabels) setLabelsInMeta(o *kubeobject.SubObject) error {
	return p.updateLabels(o, metaLabelsPath, p.Labels, true)
}

// containJobSpec check if the object is CronJob, this kind might have JobTemplateSpec, which contains label fieldPath
func containJobSpec(o *kubeobject.KubeObject) bool {
	return o.IsGroupKind(schema.GroupKind{Group: "batch", Kind: "CronJob"})
}

func (p *SetLabels) setLabelsInJob(job *kubeobject.SubObject) error {
	if job == nil {
		return nil
	}
	if err := p.setLabelsInMeta(job); err != nil {
		return err
	}
	spec := job.GetMap("spec")
	if err := p.updateLabels(spec, selectorPath, p.Labels, false); err != nil {
		return err
	}
	pod := spec.GetMap("template")
	return p.setLabelsInPod(pod)
}

// containNetworkPolicy checking if this kind is NetworkPolicy, it would contain struct NetworkPolicySpec
func containNetworkPolicy(o *kubeobject.KubeObject) bool {
	return o.IsGroupKind(schema.GroupKind{Group: "networking.k8s.io", Kind: "NetworkPolicy"})
}

// setNetworkPolicyRule set ingress and egress rules for kind NetworkPolicy
func (p *SetLabels) setNetworkPolicyRule(o *kubeobject.KubeObject) error {
	spec := o.GetMap("spec")
	if spec == nil {
		return nil
	}
	for _, vecObj := range spec.GetSlice("ingress") {
		for _, nextVecObj := range vecObj.GetSlice("from") {
			err := p.updateLabels(nextVecObj, podSelectorPath, p.Labels, false)
			if err != nil {
				return err
			}
		}
	}
	for _, vecObj := range spec.GetSlice("egress") {
		for _, nextVecObj := range vecObj.GetSlice("to") {
			err := p.updateLabels(nextVecObj, podSelectorPath, p.Labels, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// hasSpecSelector check if the resource contains struct SpecSelector, kind Service and ReplicationController has it
func hasSpecSelector(o *kubeobject.KubeObject) bool {
	return o.IsGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"}) ||
		o.IsGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ReplicationController"})
}

// setLabelsInSelector set labels for all selectors, including spec selector map, spec selector LabelSelector, LabelSelector in JobTemplate, and podSelector in NetworkPolicy, and
func (p *SetLabels) setLabelsInSelector(o *kubeobject.KubeObject) error {
	if hasSpecSelector(o) {
		fieldPath := FieldPath{"spec", "selector"}
		if err := p.updateLabels(&o.SubObject, fieldPath, p.Labels, true); err != nil {
			return err
		}
	}
	if found, create := hasLabelSelector(o); found {
		fieldPath := FieldPath{"spec", "selector", "matchLabels"}
		if err := p.updateLabels(&o.SubObject, fieldPath, p.Labels, create); err != nil {
			return err
		}
	}
	if containNetworkPolicy(o) {
		fieldPath := FieldPath{"spec", "podSelector", "matchLabels"}
		if err := p.updateLabels(&o.SubObject, fieldPath, p.Labels, false); err != nil {
			return err
		}
	}
	return nil
}

// hasLabelSelector check if the resource contains struct LabelSelector and whether labels should be created if not exist.
func hasLabelSelector(o *kubeobject.KubeObject) (bool, bool) {
	if o.IsGroupKind(schema.GroupKind{Group: "apps", Kind: "Deployment"}) ||
		o.IsGroupKind(schema.GroupKind{Group: "apps", Kind: "ReplicaSet"}) ||
		o.IsGroupKind(schema.GroupKind{Group: "apps", Kind: "DaemonSet"}) ||
		o.IsGroupKind(schema.GroupKind{Group: "apps", Kind: "StatefulSet"}) {
		return true, true
	}
	if o.IsGroupKind(schema.GroupKind{Group: "batch", Kind: "Job"}) ||
		o.IsGroupKind(schema.GroupKind{Group: "policy", Kind: "PodDisruptionBudget"}) {
		return true, false
	}
	return false, false
}

// hasPod check if the resource contains struct PodTemplateSpec, ReplicationController, Deployment, ReplicaSet, DaemonSet, StatefulSet, Job kind has it
func containPodSpec(o *kubeobject.KubeObject) bool {
	switch {
	case o.IsGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ReplicationController"}):
		return true
	case o.IsGroupKind(schema.GroupKind{Group: "apps", Kind: "Deployment"}):
		return true
	case o.IsGroupKind(schema.GroupKind{Group: "apps", Kind: "ReplicaSet"}):
		return true
	case o.IsGroupKind(schema.GroupKind{Group: "apps", Kind: "DaemonSet"}):
		return true
	case o.IsGroupKind(schema.GroupKind{Group: "apps", Kind: "StatefulSet"}):
		return true
	case o.IsGroupKind(schema.GroupKind{Group: "batch", Kind: "Job"}):
		return true
	default:
		return false
	}
}

func (p *SetLabels) setLabelsInPod(pod *kubeobject.SubObject) error {
	if pod == nil {
		return nil
	}
	err := p.setLabelsInMeta(pod)
	if err != nil {
		return err
	}
	spec := pod.GetMap("spec")
	if spec == nil {
		return nil
	}
	return p.setLabelsInPodSpec(spec)
}

// containVolume check if the resource contains struct VolumeClaimTemplates, kind StatefulSet has it
func containVolume(o *kubeobject.KubeObject) bool {
	return o.IsGroupKind(schema.GroupKind{Group: "apps", Kind: "StatefulSet"})
}

// setLabelsInVolume set VolumeClaimTemplates label path
func (p *SetLabels) setLabelsInVolumes(volumes kubeobject.SliceSubObjects) error {
	if len(volumes) == 0 {
		return nil
	}
	for _, volume := range volumes {
		if err := p.setLabelsInVolume(volume); err != nil {
			return err
		}
	}
	return nil
}

func (p *SetLabels) setLabelsInVolume(volume *kubeobject.SubObject) error {
	if volume == nil {
		return nil
	}
	if err := p.setLabelsInMeta(volume); err != nil {
		return err
	}
	return p.updateLabels(volume, selectorPath, p.Labels, false)
}

// setLabelsInPodSpec set label path in PodSpec, that include path under topologySpreadConstraints and affinity
func (p *SetLabels) setLabelsInPodSpec(podSpec *kubeobject.SubObject) error {
	_, exist, _ := podSpec.NestedSlice("topologySpreadConstraints")
	if exist {
		for _, obj := range podSpec.GetSlice("topologySpreadConstraints") {
			err := p.updateLabels(obj, labelSelector, p.Labels, false)
			if err != nil {
				return err
			}
		}
	}

	subObj := podSpec.GetMap("affinity")
	if subObj != nil {
		for _, aff := range []string{"podAffinity", "podAntiAffinity"} {
			podAff := subObj.GetMap(aff)
			if podAff != nil {
				for _, obj := range subObj.GetSlice("preferredDuringSchedulingIgnoredDuringExecution") {
					nxtObj := obj.GetMap("podAffinityTerm")
					if nxtObj != nil {
						err := p.updateLabels(nxtObj, labelSelector, p.Labels, false)
						if err != nil {
							return err
						}
					}

				}
				for _, obj := range subObj.GetSlice("requiredDuringSchedulingIgnoredDuringExecution") {
					err := p.updateLabels(obj, labelSelector, p.Labels, false)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// updateLabels the update process for each label, sort the keys to preserve sequence, return if the update was performed and potential error
func (p *SetLabels) updateLabels(o *kubeobject.SubObject, labelPath FieldPath, labels map[string]string, create bool) error {
	if o == nil {
		return nil
	}
	keys := make([]string, 0)
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i := 0; i < len(keys); i++ {
		key := keys[i]
		val := labels[key]
		newPath := append(labelPath, key)
		oldValue, exist, err := o.NestedString(newPath...)
		if err != nil {
			return err
		}
		if (exist && oldValue != val) || (!exist && create) {
			if err = o.SetNestedString(val, newPath...); err != nil {
				return err
			}
		}
		p.count += 1
	}
	return nil
}
