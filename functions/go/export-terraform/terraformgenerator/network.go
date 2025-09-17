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
	sdk "github.com/kptdev/krm-functions-catalog/thirdparty/kyaml/fnsdk"
)

type firewallAllow struct {
	Ports    []string `yaml:"ports,omitempty"`
	Protocol string   `yaml:"protocol"`
}

func (resource *terraformResource) GetFirewallAllowPortsProtocol() []firewallAllow {
	var firewallAllows []firewallAllow
	found, err := resource.Item.Get(&firewallAllows, "spec", "allow")
	if !found || err != nil {
		sdk.Logf("unable to find allowed FW ports/protocol in %s (found = %t): %s\n", resource.Name, found, err)
	}
	return firewallAllows
}

// IsSVPCHost checks if the resource is a SVPC Host project.
// The resource is a SVPC Host project if and only if it is of kind Project
// and has a corresponding ComputeSharedVPCHostProject child resource.
func (resource *terraformResource) IsSVPCHost() bool {
	if resource.Kind != "Project" {
		return false
	}
	projectID, found, err := resource.Item.GetString("metadata", "name")
	if !found || err != nil {
		return false
	}
	for _, child := range resource.Children {
		if child.Kind != "ComputeSharedVPCHostProject" {
			continue
		}
		// ComputeSharedVPCHostProject has no spec and relies on anno
		// https://cloud.google.com/config-connector/docs/reference/resource-docs/compute/computesharedvpchostproject#annotations
		svpcHostProjectID, found, err := child.Item.GetString("metadata", "annotations", "cnrm.cloud.google.com/project-id")
		if !found || err != nil {
			continue
		}
		if projectID == svpcHostProjectID {
			return true
		}
	}
	return false
}
