// Copyright 2021-2025 The kpt Authors
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
	metav1.ObjectMeta `json:"metadata"`
	Spec              ApplyTimeMutationSpec `json:"spec"`
}

// ApplyTimeMutationSpec specifies a one or more substitutions to perform on a
// target object at apply-time.
type ApplyTimeMutationSpec struct {
	TargetRef     mutation.ResourceReference `json:"targetRef"`
	Substitutions mutation.ApplyTimeMutation `json:"substitutions,omitempty"`
}
