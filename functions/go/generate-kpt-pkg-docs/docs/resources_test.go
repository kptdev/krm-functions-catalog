package docs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestGetResourceDocsLink(t *testing.T) {
	tests := []struct {
		name string
		r    resid.Gvk
		want string
	}{
		{
			name: "generates doc link for ConfigConnectorContext",
			r:    resid.Gvk{Kind: "ConfigConnectorContext", Group: "core.cnrm.cloud.google.com"},
			want: "[ConfigConnectorContext](https://cloud.google.com/config-connector/docs/how-to/advanced-install#addon-configuring)",
		},
		{
			name: "generates doc link for ConfigManagement",
			r:    resid.Gvk{Kind: "ConfigManagement", Group: "configmanagement.gke.io"},
			want: "[ConfigManagement](https://cloud.google.com/anthos-config-management/docs/configmanagement-fields)",
		},
		{
			name: "generates doc link for a KCC resource",
			r:    resid.Gvk{Kind: "IAMServiceAccount", Group: "iam.cnrm.cloud.google.com"},
			want: "[IAMServiceAccount](https://cloud.google.com/config-connector/docs/reference/resource-docs/iam/iamserviceaccount)",
		},
		{
			name: "empty for unknown resource",
			r:    resid.Gvk{Kind: "Foo", Group: "bar.baz", Version: "v1"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getResourceDocsLink(tt.r)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestShouldSkipResource(t *testing.T) {
	tests := []struct {
		name      string
		r         string
		skipFiles map[string]bool
		want      bool
	}{
		{
			name: "should skip no path",
			r: `apiVersion: v1
kind: Namespace
metadata:
  name: project-id # kpt-set: ${project-id}
  annotations:
    cnrm.cloud.google.com/project-id: project-id # kpt-set: ${project-id}`,
			want: true,
		},
		{
			name: "should skip subpkg",
			r: `apiVersion: v1
kind: Namespace
metadata:
  name: project-id # kpt-set: ${project-id}
  annotations:
    cnrm.cloud.google.com/project-id: project-id # kpt-set: ${project-id}
    internal.config.kubernetes.io/path: foo/bar.yaml`,
			want:      true,
			skipFiles: map[string]bool{"foo": true},
		},
		{
			name: "should skip file",
			r: `apiVersion: v1
kind: Namespace
metadata:
  name: project-id # kpt-set: ${project-id}
  annotations:
    cnrm.cloud.google.com/project-id: project-id # kpt-set: ${project-id}
    internal.config.kubernetes.io/path: foo/bar.yaml`,
			want:      true,
			skipFiles: map[string]bool{"foo/bar.yaml": true},
		},
		{
			name: "should skip setter cfg",
			r: `apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
  annotations:
    internal.config.kubernetes.io/path: setters.yaml
data:
  namespace: networking
  network-name: network-name
  project-id: project-id`,
			want:      true,
			skipFiles: map[string]bool{"setters.yaml": true},
		},
		{
			name: "should skip setter kptfile",
			r: `apiVersion: kpt.dev/v1
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
			want:      true,
			skipFiles: map[string]bool{"Kptfile": true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := getRNodesFromStr(t, tt.r)
			got := shouldSkipResource(res[0], tt.skipFiles)
			require.Equal(t, tt.want, got)
		})
	}
}

func getRNodesFromStr(t *testing.T, res string) []*yaml.RNode {
	t.Helper()
	require := require.New(t)
	nodes, err := (&kio.ByteReader{
		Reader: strings.NewReader(res),
	}).Read()
	require.NoError(err)
	return nodes
}

func TestFindResourcePath(t *testing.T) {
	tests := []struct {
		name string
		r    string
		want string
		err  string
	}{
		{
			name: "simple",
			r: `apiVersion: compute.cnrm.cloud.google.com/v1beta1
kind: ComputeNetwork
metadata:
  name: network-name # kpt-set: ${network-name}
  namespace: networking # kpt-set: ${namespace}
  annotations:
    cnrm.cloud.google.com/blueprint: cnrm/landing-zone:networking/v0.4.0
    cnrm.cloud.google.com/project-id: project-id # kpt-set: ${project-id}
    internal.config.kubernetes.io/path: vpc.yaml
spec:
  autoCreateSubnetworks: false
  deleteDefaultRoutesOnCreate: false
  routingMode: GLOBAL
`,
			want: "vpc.yaml",
		},
		{
			name: "legacy",
			r: `apiVersion: compute.cnrm.cloud.google.com/v1beta1
kind: ComputeNetwork
metadata:
  name: network-name # kpt-set: ${network-name}
  namespace: networking # kpt-set: ${namespace}
  annotations:
    cnrm.cloud.google.com/blueprint: cnrm/landing-zone:networking/v0.4.0
    cnrm.cloud.google.com/project-id: project-id # kpt-set: ${project-id}
    config.kubernetes.io/path: vpc.yaml
spec:
  autoCreateSubnetworks: false
  deleteDefaultRoutesOnCreate: false
  routingMode: GLOBAL
`,
			want: "vpc.yaml",
		},
		{
			name: "missing",
			r: `apiVersion: compute.cnrm.cloud.google.com/v1beta1
kind: ComputeNetwork
metadata:
  name: network-name # kpt-set: ${network-name}
  namespace: networking # kpt-set: ${namespace}
  annotations:
    cnrm.cloud.google.com/blueprint: cnrm/landing-zone:networking/v0.4.0
    cnrm.cloud.google.com/project-id: project-id # kpt-set: ${project-id}
spec:
  autoCreateSubnetworks: false
  deleteDefaultRoutesOnCreate: false
  routingMode: GLOBAL
  project-id: project-id
`,
			err: "unable find resource path for ComputeNetwork.v1beta1.compute.cnrm.cloud.google.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := getRNodesFromStr(t, tt.r)
			got, err := findResourcePath(res[0])
			require := require.New(t)
			if tt.err != "" {
				require.EqualError(err, tt.err)
			} else {
				require.NoError(err)
				require.Equal(tt.want, got)
			}

		})
	}
}
