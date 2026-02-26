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
package docs

import (
	"strings"
	"testing"

	kptfilev1 "github.com/kptdev/kpt/pkg/api/kptfile/v1"
	"github.com/kptdev/kpt/pkg/kptfile/kptfileutil"
	"github.com/stretchr/testify/require"
)

func TestGetFnCfgPaths(t *testing.T) {
	tests := []struct {
		name string
		kf   string
		want []string
	}{
		{
			name: "simple",
			kf: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: kcc-namespace
  annotations:
    blueprints.cloud.google.com/title: Project Namespace Package
    internal.config.kubernetes.io/path: Kptfile
info:
  description: |
    Kubernetes namespace configured for use with Config Connector to manage GCP
    resources in a specific project.
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.1
      configPath: setters.yaml
  validators:
    - image: ghcr.io/kptdev/krm-functions-catalog/starlark:v0.3
      configPath: validation.yaml`,
			want: []string{"setters.yaml", "validation.yaml"},
		},
		{
			name: "no pipeline",
			kf: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: kcc-namespace
  annotations:
    blueprints.cloud.google.com/title: Project Namespace Package
    internal.config.kubernetes.io/path: Kptfile
info:
  description: |
    Kubernetes namespace configured for use with Config Connector to manage GCP
    resources in a specific project.
`,
			want: []string{},
		},
		{
			name: "with cm",
			kf: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: kcc-namespace
  annotations:
    blueprints.cloud.google.com/title: Project Namespace Package
    internal.config.kubernetes.io/path: Kptfile
info:
  description: |
    Kubernetes namespace configured for use with Config Connector to manage GCP
    resources in a specific project.
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.1
      configMap:
        foo: bar
  validators:
    - image: ghcr.io/kptdev/krm-functions-catalog/starlark:v0.3
      configPath: validation.yaml`,
			want: []string{"validation.yaml"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			kf := getKfFromStr(t, tt.kf)
			got := getFnCfgPaths(kf)
			require.ElementsMatch(tt.want, got)
		})
	}
}

func getKfFromStr(t *testing.T, k string) *kptfilev1.KptFile {
	t.Helper()
	require := require.New(t)
	kf, err := kptfileutil.DecodeKptfile(strings.NewReader(k))
	require.NoError(err)
	return kf
}
