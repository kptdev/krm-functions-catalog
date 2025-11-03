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

package meta

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// IsClusterScoped returns true if the specified GVK is cluster scoped, or an error if this cannot be determined.
func IsClusterScoped(gvk schema.GroupVersionKind) (bool, error) {
	ki := findKindInfo(gvk)
	if ki == nil {
		return false, fmt.Errorf("kind %v not known", gvk)
	}
	return ki.ClusterScoped, nil
}

type kindInfo struct {
	GVK           schema.GroupVersionKind
	ClusterScoped bool
}

// kindInfos holds a hard-coded list of type information.
// This should be replaced with OpenAPI-derived information in the future.
var kindInfos = []kindInfo{
	{
		GVK:           schema.GroupVersionKind{Group: "container.cnrm.cloud.google.com", Version: "v1beta1", Kind: "ContainerCluster"},
		ClusterScoped: false,
	},
	{
		GVK:           schema.GroupVersionKind{Group: "container.cnrm.cloud.google.com", Version: "v1beta1", Kind: "ContainerNodePool"},
		ClusterScoped: false,
	},
	{
		GVK:           schema.GroupVersionKind{Group: "iam.cnrm.cloud.google.com", Version: "v1beta1", Kind: "IAMPolicyMember"},
		ClusterScoped: false,
	},
	{
		GVK:           schema.GroupVersionKind{Group: "iam.cnrm.cloud.google.com", Version: "v1beta1", Kind: "IAMServiceAccount"},
		ClusterScoped: false,
	},
	{
		GVK:           schema.GroupVersionKind{Group: "resourcemanager.cnrm.cloud.google.com", Version: "v1beta1", Kind: "Project"},
		ClusterScoped: false,
	},
	{
		GVK:           schema.GroupVersionKind{Group: "resourcemanager.cnrm.cloud.google.com", Version: "v1beta1", Kind: "Folder"},
		ClusterScoped: false,
	},
	{
		GVK:           schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
		ClusterScoped: true,
	},
	{
		GVK:           schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role"},
		ClusterScoped: false,
	},
	{
		GVK:           schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"},
		ClusterScoped: false,
	},
	{
		GVK:           schema.GroupVersionKind{Group: "config.porch.kpt.dev", Version: "v1alpha1", Kind: "WorkloadIdentityBinding"},
		ClusterScoped: false,
	},
}

func findKindInfo(gvk schema.GroupVersionKind) *kindInfo {
	for i := range kindInfos {
		ki := &kindInfos[i]
		if ki.GVK == gvk {
			return ki
		}
	}
	return nil
}
