package pkg

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cli-utils/pkg/object/mutation"
)

// ApplyTimeMutation is a Kubernetes resource that allows specifying mutations
// using a seperate KRM object, instead of an annotation string on the target
// object.
type ApplyTimeMutation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ApplyTimeMutationSpec `json:"spec,omitempty"`
}

// ApplyTimeMutationSpec specifies a one or more substitutions to perform on a
// target object at apply-time.
type ApplyTimeMutationSpec struct {
	TargetRef     mutation.ResourceReference `json:"targetRef,omitempty"`
	Substitutions mutation.ApplyTimeMutation `json:"substitutions,omitempty"`
}
