// Copyright 2021 Google LLC
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

type variable struct {
	Name        string
	Description string
	Default     string
}

type resourceVariable struct {
	*variable
	kind       string
	underlying *variable
}

// Find common values and make them into variables
func (rs *terraformResources) makeVariables() {
	rs.Variables = make(map[string]*variable)

	resources := rs.getGrouped()
	vars := []resourceVariable{
		{
			kind: "Organization",
			underlying: &variable{
				Name:        "org_id",
				Description: "The organization id for the associated resources",
			},
		},
		{
			kind: "BillingAccount",
			underlying: &variable{
				Name:        "billing_account",
				Description: "The ID of the billing account to associate projects with",
			},
		},
	}

	for _, candidate := range vars {
		if len(resources[candidate.kind]) != 1 {
			continue
		}
		resource := resources[candidate.kind][0]
		candidate.underlying.Default = resource.Name
		resource.variable = candidate.underlying
		rs.Variables[candidate.underlying.Name] = candidate.underlying
	}
}
