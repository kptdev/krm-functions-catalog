package setlabels

import (
	"sort"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
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
	Labels  map[string]string `json:"labels,omitempty"`

	// options is the configurable and optional options
	Options map[string]bool   `json:"options,omitempty"`

	count   int
}

// EmptyfnConfig is a workaround since kpt creates a FunctionConfig placeholder if users don't provide the functionConfig.
// `kpt fn eval` uses placeholder ConfigMap with name "function-input"
// `kpt fn render` uses placeholder "{}"
func EmptyfnConfig(o *fn.KubeObject) bool {
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
func (p *SetLabels) Run(_ *fn.Context, fnConfig *fn.KubeObject, items fn.KubeObjects, results *fn.Results) bool {
	objects := items.WhereNot(fn.IsLocalConfig)
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
func (p *SetLabels) setLabelsInMeta(o *fn.SubObject) error {
	return p.updateLabels(o, metaLabelsPath, p.Labels, true)
}

// containJobSpec check if the object is CronJob, this kind might have JobTemplateSpec, which contains label fieldPath
func containJobSpec(o *fn.KubeObject) bool {
	return o.IsGVK("batch", "", "CronJob")
}

func (p *SetLabels) setLabelsInJob(job *fn.SubObject) error {
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
func containNetworkPolicy(o *fn.KubeObject) bool {
	return o.IsGVK("networking.k8s.io", "", "NetworkPolicy")
}

// setNetworkPolicyRule set ingress and egress rules for kind NetworkPolicy
func (p *SetLabels) setNetworkPolicyRule(o *fn.KubeObject) error {
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
func hasSpecSelector(o *fn.KubeObject) bool {
	return o.IsGVK("", "v1", "Service") || o.IsGVK("", "v1", "ReplicationController")
}

// setLabelsInSelector set labels for all selectors, including spec selector map, spec selector LabelSelector, LabelSelector in JobTemplate, and podSelector in NetworkPolicy, and
func (p *SetLabels) setLabelsInSelector(o *fn.KubeObject) error {
	// setSelectorLabels is true by default
	if !p.Options["setSelectorLabels"] {
		return nil
	}

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
func hasLabelSelector(o *fn.KubeObject) (bool, bool) {
	if o.IsGVK("", "", "Deployment") || o.IsGVK("", "", "ReplicaSet") || o.IsGVK("", "", "DaemonSet") || o.IsGVK("apps", "", "StatefulSet") {
		return true, true
	}
	if o.IsGVK("batch", "", "Job") || o.IsGVK("policy", "", "PodDisruptionBudget") {
		return true, false
	}
	return false, false
}

// hasPod check if the resource contains struct PodTemplateSpec, ReplicationController, Deployment, ReplicaSet, DaemonSet, StatefulSet, Job kind has it
func containPodSpec(o *fn.KubeObject) bool {
	switch {
	case o.IsGVK("", "v1", "ReplicationController"):
		return true
	case o.IsGVK("", "", "Deployment"):
		return true
	case o.IsGVK("", "", "ReplicaSet"):
		return true
	case o.IsGVK("", "", "DaemonSet"):
		return true
	case o.IsGVK("apps", "", "StatefulSet"):
		return true
	case o.IsGVK("batch", "", "Job"):
		return true
	default:
		return false
	}
}

func (p *SetLabels) setLabelsInPod(pod *fn.SubObject) error {
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
func containVolume(o *fn.KubeObject) bool {
	return o.IsGVK("apps", "", "StatefulSet")
}

// setLabelsInVolume set VolumeClaimTemplates label path
func (p *SetLabels) setLabelsInVolumes(volumes fn.SliceSubObjects) error {
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

func (p *SetLabels) setLabelsInVolume(volume *fn.SubObject) error {
	if volume == nil {
		return nil
	}
	if err := p.setLabelsInMeta(volume); err != nil {
		return err
	}
	return p.updateLabels(volume, selectorPath, p.Labels, false)
}

// setLabelsInPodSpec set label path in PodSpec, that include path under topologySpreadConstraints and affinity
func (p *SetLabels) setLabelsInPodSpec(podSpec *fn.SubObject) error {
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
func (p *SetLabels) updateLabels(o *fn.SubObject, labelPath FieldPath, labels map[string]string, create bool) error {
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
