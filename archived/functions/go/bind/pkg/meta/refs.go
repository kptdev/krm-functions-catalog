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

import "k8s.io/apimachinery/pkg/runtime/schema"

type refInfo struct {
	GVK                schema.GroupVersionKind
	FieldPath          string
	TargetGVKs         []schema.GroupVersionKind
	CrossNamespace     bool
	NamespaceFieldPath string
}

// refInfos is a hard-coded list of reference fields in various types.
// This should be replaced by OpenAPI derived information in the future.
var refInfos = []refInfo{
	{
		GVK:       schema.GroupVersionKind{Group: "container.cnrm.cloud.google.com", Version: "v1beta1", Kind: "ContainerNodePool"},
		FieldPath: "spec.clusterRef",
		TargetGVKs: []schema.GroupVersionKind{
			{Group: "container.cnrm.cloud.google.com", Version: "v1beta1", Kind: "ContainerCluster"},
		},
		// CrossNamespace: ?
	},
	{
		GVK:       schema.GroupVersionKind{Group: "container.cnrm.cloud.google.com", Version: "v1beta1", Kind: "ContainerNodePool"},
		FieldPath: "spec.nodeConfig.serviceAccountRef",
		TargetGVKs: []schema.GroupVersionKind{
			{Group: "iam.cnrm.cloud.google.com", Version: "v1beta1", Kind: "IAMServiceAccount"},
		},
		// CrossNamespace: ?
	},
	{
		GVK:       schema.GroupVersionKind{Group: "iam.cnrm.cloud.google.com", Version: "v1beta1", Kind: "IAMPolicyMember"},
		FieldPath: "spec.memberFrom.serviceAccountRef",
		TargetGVKs: []schema.GroupVersionKind{
			{Group: "iam.cnrm.cloud.google.com", Version: "v1beta1", Kind: "IAMServiceAccount"},
		},
		// CrossNamespace: ?
	},
	{
		GVK:       schema.GroupVersionKind{Group: "resourcemanager.cnrm.cloud.google.com", Version: "v1beta1", Kind: "Folder"},
		FieldPath: "spec.folderRef",
		TargetGVKs: []schema.GroupVersionKind{
			{Group: "resourcemanager.cnrm.cloud.google.com", Version: "v1beta1", Kind: "Folder"},
		},
		// CrossNamespace: ?
		NamespaceFieldPath: "namespace",
	},
	{
		GVK:       schema.GroupVersionKind{Group: "resourcemanager.cnrm.cloud.google.com", Version: "v1beta1", Kind: "Project"},
		FieldPath: "spec.folderRef",
		TargetGVKs: []schema.GroupVersionKind{
			{Group: "resourcemanager.cnrm.cloud.google.com", Version: "v1beta1", Kind: "Folder"},
		},
		// CrossNamespace: ?
		NamespaceFieldPath: "namespace",
	},

	{
		GVK:        schema.GroupVersionKind{Group: "iam.cnrm.cloud.google.com", Version: "v1beta1", Kind: "IAMPolicyMember"},
		FieldPath:  "spec.resourceRef",
		TargetGVKs: nil, // Any (?)
		// CrossNamespace: ?
		NamespaceFieldPath: "namespace",
	},

	{
		GVK:        schema.GroupVersionKind{Group: "iam.cnrm.cloud.google.com", Version: "v1beta1", Kind: "IAMPartialPolicy"},
		FieldPath:  "spec.resourceRef",
		TargetGVKs: []schema.GroupVersionKind{},
		// CrossNamespace: ?
	},
	{
		GVK:       schema.GroupVersionKind{Group: "iam.cnrm.cloud.google.com", Version: "v1beta1", Kind: "IAMPartialPolicy"},
		FieldPath: "spec.bindings[].members[].memberFrom.serviceAccountRef",
		TargetGVKs: []schema.GroupVersionKind{
			{Group: "iam.cnrm.cloud.google.com", Version: "v1beta1", Kind: "IAMServiceAccount"},
		},
		// CrossNamespace: ?
	},

	{
		GVK:       schema.GroupVersionKind{Group: "serviceusage.cnrm.cloud.google.com", Version: "v1beta1", Kind: "Service"},
		FieldPath: "spec.projectRef",
		TargetGVKs: []schema.GroupVersionKind{
			{Group: "resourcemanager.cnrm.cloud.google.com", Version: "v1beta1", Kind: "Project"},
		},
		NamespaceFieldPath: "namespace",
		// CrossNamespace: ?
	},

	{
		GVK:       schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding"},
		FieldPath: "roleRef",
		TargetGVKs: []schema.GroupVersionKind{
			{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role"},
			{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
		},
		// CrossNamespace: ?
	},
	{
		GVK:       schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding"},
		FieldPath: "subjects[]",
		TargetGVKs: []schema.GroupVersionKind{
			{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "User"},
			{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Group"},
			{Group: "", Version: "v1", Kind: "ServiceAccount"},
		},
		// CrossNamespace: ?
	},
	{
		GVK:       schema.GroupVersionKind{Group: "config.porch.kpt.dev", Version: "v1alpha1", Kind: "WorkloadIdentityBinding"},
		FieldPath: "spec.resourceRef",
		TargetGVKs: []schema.GroupVersionKind{
			{Group: "iam.cnrm.cloud.google.com", Version: "v1beta1", Kind: "IAMServiceAccount"},
		},
		// CrossNamespace: ?
	},
	{
		GVK:       schema.GroupVersionKind{Group: "config.porch.kpt.dev", Version: "v1alpha1", Kind: "WorkloadIdentityBinding"},
		FieldPath: "spec.serviceAccountRef",
		TargetGVKs: []schema.GroupVersionKind{
			{Group: "", Version: "v1", Kind: "ServiceAccount"},
		},
		// CrossNamespace: ?
		NamespaceFieldPath: "namespace",
	},
}
