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

package terraformgenerator

import (
	sdk "github.com/kptdev/krm-functions-catalog/archived/functions/go/export-terraform/thirdparty/kyaml/fnsdk"
)

type iamMember struct {
	Member string
}

type iamBinding struct {
	Members []*iamMember
	Role    string
}

type iamPolicy struct {
	bindings map[string]*iamBinding
}

type iamAuditLogConfig struct {
	ExemptedMembers []string `yaml:"exemptedMembers,omitempty"`
	LogType         string   `yaml:"logType"`
}

func (policy *iamPolicy) addBinding(role string, member string) {
	if policy.bindings == nil {
		policy.bindings = make(map[string]*iamBinding)
	}
	_, found := policy.bindings[role]
	if !found {
		policy.bindings[role] = &iamBinding{Role: role}
	}
	policy.bindings[role].Members = append(policy.bindings[role].Members, &iamMember{Member: member})
}

func (resource *terraformResource) HasIAMBindings() bool {
	bindings := resource.GetIAMBindings()
	return len(bindings) > 0
}

// Retrieve IAM bindings for a given resource
func (resource *terraformResource) GetIAMBindings() map[string]*iamBinding {
	policy := &iamPolicy{}

	for _, child := range resource.Children {
		switch child.Kind {
		case "IAMPolicyMember":
			role := child.GetStringFromObject("spec", "role")
			member := child.GetStringFromObject("spec", "member")
			policy.addBinding(role, member)
		case "IAMPartialPolicy", "IAMPolicy":
			var bindings []iamBinding
			found, err := child.Item.Get(&bindings, "spec", "bindings")
			if !found || err != nil {
				sdk.Logf("Failure to find bindings in %s (found = %t): %s\n", resource.Name, found, err)
				continue
			}
			for _, binding := range bindings {
				for _, member := range binding.Members {
					policy.addBinding(binding.Role, member.Member)
				}
			}
		}
	}

	return policy.bindings
}

func (resource *terraformResource) GetIAMAuditLogConfigs() []iamAuditLogConfig {
	var auditLogConfigs []iamAuditLogConfig
	found, err := resource.Item.Get(&auditLogConfigs, "spec", "auditLogConfigs")
	if !found || err != nil {
		sdk.Logf("unable to find auditLogConfigs in %s (found = %t): %s\n", resource.Name, found, err)
	}
	return auditLogConfigs
}
