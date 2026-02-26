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
package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var tests = []struct {
	name           string
	config         string
	input          string
	expectedResult []string
	errMsg         string
}{
	{
		name: "set policy as dryrun",
		input: `apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sRestrictRoleBindings
metadata: # kpt-merge: /restrict-clusteradmin-rolebindings
  name: restrict-clusteradmin-rolebindings
  annotations:
  # This constraint is not certified by CIS.
  description: "Restricts use of the cluster-admin role."
  bundles.validator.forsetisecurity.org/cis-k8s-v1.5.1: 5.1.1
spec:
  enforcementAction: deny # kpt-set: ${enforcementAction}
  parameters:
    restrictedRole:
      apiGroup: "rbac.authorization.k8s.io"
      kind: "ClusterRole"
      name: "cluster-admin"
    allowedSubjects:
    - apiGroup: "rbac.authorization.k8s.io"
      kind: "Group"
      name: "system:masters"
`,
		config: `data:
  enforcementAction: dryrun
`,
		expectedResult: []string{
			"Number of policies set to [dryrun]: 1",
			"Policy name: [restrict-clusteradmin-rolebindings]",
		},
	},
	{
		name: "set policy as deny",
		input: `apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sRestrictRoleBindings
metadata: # kpt-merge: /restrict-clusteradmin-rolebindings
  name: restrict-clusteradmin-rolebindings
  annotations:
  # This constraint is not certified by CIS.
  description: "Restricts use of the cluster-admin role."
  bundles.validator.forsetisecurity.org/cis-k8s-v1.5.1: 5.1.1
spec:
  enforcementAction: dryrun # kpt-set: ${enforcementAction}
  parameters:
    restrictedRole:
      apiGroup: "rbac.authorization.k8s.io"
      kind: "ClusterRole"
      name: "cluster-admin"
    allowedSubjects:
    - apiGroup: "rbac.authorization.k8s.io"
      kind: "Group"
      name: "system:masters"
`,
		config: `data:
  enforcementAction: deny
`,
		expectedResult: []string{
			"Number of policies set to [deny]: 1",
			"Policy name: [restrict-clusteradmin-rolebindings]",
		},
	},
	{
		name: "no policy found",
		input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: fc-config
data:
  foo: bar
`,
		config: `data:
  enforcementAction: deny
`,
		expectedResult: []string{
			"Found no policy to set to [deny]",
		},
	},
	{
		name: "incorrect enforementAction",
		input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: fc-config
data:
  foo: bar
`,
		config: `data:
  enforcementAction: dry-run
`,
		expectedResult: []string{
			"expected values for enforcementAction are [deny] or [warn] or [dryrun]",
		},
	},
	{
		name: "too many enforementAction configs",
		input: `apiVersion: v1
kind: ConfigMap
metadata:
  name: fc-config
data:
  foo: bar
`,
		config: `data:
  enforcementAction: dryrun
  foo: bar
`,
		expectedResult: []string{
			"expecting exactly 1 enforcementAction as part of the ConfigMap",
		},
	},
}

func TestPolicyResources(t *testing.T) {
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {

			var acn string
			fcNode, err := yaml.Parse(test.config)
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			err = getEnforcementAction(fcNode, &acn)
			if err != nil {
				require.Equal(t, test.expectedResult[0], err.Error())
				return
			}

			policy, err := yaml.Parse(test.input)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			policies := append([]*yaml.RNode{}, policy)
			items, err := processPolicies(policies, acn)
			if err != nil {
				t.Errorf("Error when calling processPolicies %s", err.Error())
			}

			for j := range items {
				require.Equal(t, test.expectedResult[j], items[j].Message)
			}
		})
	}
}
