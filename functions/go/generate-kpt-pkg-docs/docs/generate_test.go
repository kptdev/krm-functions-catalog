package docs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateReadme(t *testing.T) {
	tests := []struct {
		name      string
		r         string
		pkgName   string
		wantTitle string
		wantDoc   string
		err       string
	}{
		{
			name: "simple",
			r: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: project
  annotations:
    blueprints.cloud.google.com/title: Project blueprint
    internal.config.kubernetes.io/path: Kptfile
info:
  description: |
    A project and a project namespace in which to manage project resources with
    Config Connector.
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.1
      configPath: setters.yaml
---
apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
kind: Project
metadata:
  name: project-id # kpt-set: ${project-id}
  namespace: projects # kpt-set: ${projects-namespace}
  annotations:
    cnrm.cloud.google.com/auto-create-network: "false"
    cnrm.cloud.google.com/blueprint: cnrm/landing-zone:project/v0.4.1
    internal.config.kubernetes.io/path: project.yaml
spec:
  name: project-id # kpt-set: ${project-id}
  billingAccountRef:
    external: "AAAAAA-BBBBBB-CCCCCC" # kpt-set: ${billing-account-id}
  folderRef:
    name: name.of.folder # kpt-set: ${folder-name}
    namespace: hierarchy # kpt-set: ${folder-namespace}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
  annotations:
    internal.config.kubernetes.io/path: setters.yaml
data:
  folder-name: name.of.folder
  project-id: project-id
  # These defaults can be kept
  folder-namespace: hierarchy
`,
			wantTitle: `# Project blueprint

`,
			wantDoc: `A project and a project namespace in which to manage project resources with
Config Connector.

## Setters

|        Name        |        Value         | Type | Count |
|--------------------|----------------------|------|-------|
| billing-account-id | AAAAAA-BBBBBB-CCCCCC | str  |     1 |
| folder-name        | name.of.folder       | str  |     1 |
| folder-namespace   | hierarchy            | str  |     1 |
| project-id         | project-id           | str  |     2 |
| projects-namespace | projects             | str  |     1 |

## Sub-packages

This package has no sub-packages.

## Resources

|     File     |                  APIVersion                   |  Kind   |    Name    | Namespace |
|--------------|-----------------------------------------------|---------|------------|-----------|
| project.yaml | resourcemanager.cnrm.cloud.google.com/v1beta1 | Project | project-id | projects  |

## Resource References

- [Project](https://cloud.google.com/config-connector/docs/reference/resource-docs/resourcemanager/project)

## Usage

1.  Clone the package:
    ¬¬¬shell
    kpt pkg get https://github.com/GoogleCloudPlatform/blueprints.git/catalog/project@${VERSION}
    ¬¬¬
    Replace ¬${VERSION}¬ with the desired repo branch or tag
    (for example, ¬main¬).

1.  Move into the local package:
    ¬¬¬shell
    cd "./project/"
    ¬¬¬

1.  Edit the function config file(s):
    - setters.yaml

1.  Execute the function pipeline
    ¬¬¬shell
    kpt fn render
    ¬¬¬

1.  Initialize the resource inventory
    ¬¬¬shell
    kpt live init --namespace ${NAMESPACE}
    ¬¬¬
    Replace ¬${NAMESPACE}¬ with the namespace in which to manage
    the inventory ResourceGroup (for example, ¬config-control¬).

1.  Apply the package resources to your cluster
    ¬¬¬shell
    kpt live apply
    ¬¬¬

1.  Wait for the resources to be ready
    ¬¬¬shell
    kpt live status --output table --poll-until current
    ¬¬¬
`,
		},
		{
			name: "simple with custom pkg name",
			r: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: project
  annotations:
    blueprints.cloud.google.com/title: Project blueprint
    internal.config.kubernetes.io/path: Kptfile
info:
  description: |
    A project and a project namespace in which to manage project resources with
    Config Connector.
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.1
      configPath: setters.yaml
---
apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
kind: Project
metadata:
  name: project-id # kpt-set: ${project-id}
  namespace: projects # kpt-set: ${projects-namespace}
  annotations:
    cnrm.cloud.google.com/auto-create-network: "false"
    cnrm.cloud.google.com/blueprint: cnrm/landing-zone:project/v0.4.1
    internal.config.kubernetes.io/path: project.yaml
spec:
  name: project-id # kpt-set: ${project-id}
  billingAccountRef:
    external: "AAAAAA-BBBBBB-CCCCCC" # kpt-set: ${billing-account-id}
  folderRef:
    name: name.of.folder # kpt-set: ${folder-name}
    namespace: hierarchy # kpt-set: ${folder-namespace}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
  annotations:
    internal.config.kubernetes.io/path: setters.yaml
data:
  folder-name: name.of.folder
  project-id: project-id
  # These defaults can be kept
  folder-namespace: hierarchy
`,
			pkgName: "foopkg",
			wantTitle: `# Project blueprint

`,
			wantDoc: `A project and a project namespace in which to manage project resources with
Config Connector.

## Setters

|        Name        |        Value         | Type | Count |
|--------------------|----------------------|------|-------|
| billing-account-id | AAAAAA-BBBBBB-CCCCCC | str  |     1 |
| folder-name        | name.of.folder       | str  |     1 |
| folder-namespace   | hierarchy            | str  |     1 |
| project-id         | project-id           | str  |     2 |
| projects-namespace | projects             | str  |     1 |

## Sub-packages

This package has no sub-packages.

## Resources

|     File     |                  APIVersion                   |  Kind   |    Name    | Namespace |
|--------------|-----------------------------------------------|---------|------------|-----------|
| project.yaml | resourcemanager.cnrm.cloud.google.com/v1beta1 | Project | project-id | projects  |

## Resource References

- [Project](https://cloud.google.com/config-connector/docs/reference/resource-docs/resourcemanager/project)

## Usage

1.  Clone the package:
    ¬¬¬shell
    kpt pkg get https://github.com/GoogleCloudPlatform/blueprints.git/catalog/foopkg@${VERSION}
    ¬¬¬
    Replace ¬${VERSION}¬ with the desired repo branch or tag
    (for example, ¬main¬).

1.  Move into the local package:
    ¬¬¬shell
    cd "./foopkg/"
    ¬¬¬

1.  Edit the function config file(s):
    - setters.yaml

1.  Execute the function pipeline
    ¬¬¬shell
    kpt fn render
    ¬¬¬

1.  Initialize the resource inventory
    ¬¬¬shell
    kpt live init --namespace ${NAMESPACE}
    ¬¬¬
    Replace ¬${NAMESPACE}¬ with the namespace in which to manage
    the inventory ResourceGroup (for example, ¬config-control¬).

1.  Apply the package resources to your cluster
    ¬¬¬shell
    kpt live apply
    ¬¬¬

1.  Wait for the resources to be ready
    ¬¬¬shell
    kpt live status --output table --poll-until current
    ¬¬¬
`,
		},
		{
			name: "missing root kf",
			r: `
apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
kind: Project
metadata:
  name: project-id # kpt-set: ${project-id}
  namespace: projects # kpt-set: ${projects-namespace}
  annotations:
    cnrm.cloud.google.com/auto-create-network: "false"
    cnrm.cloud.google.com/blueprint: cnrm/landing-zone:project/v0.4.1
    internal.config.kubernetes.io/path: project.yaml
spec:
  name: project-id # kpt-set: ${project-id}
  billingAccountRef:
    external: "AAAAAA-BBBBBB-CCCCCC" # kpt-set: ${billing-account-id}
  folderRef:
    name: name.of.folder # kpt-set: ${folder-name}
    namespace: hierarchy # kpt-set: ${folder-namespace}
`,
			err: "unable to find root Kptfile, please include --include-meta-resources flag if a Kptfile is present",
		},
		{
			name: "invalid kf",
			r: `
apiVersion: kpt.dev/v1
kind: Kptfile
annotations:
    blueprints.cloud.google.com/title: Project blueprint
    internal.config.kubernetes.io/path: Kptfile
info:
  description: |
    A project and a project namespace in which to manage project resources with
    Config Connector.
---
apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
kind: Project
metadata:
  name: project-id # kpt-set: ${project-id}
  namespace: projects # kpt-set: ${projects-namespace}
  annotations:
    cnrm.cloud.google.com/auto-create-network: "false"
    cnrm.cloud.google.com/blueprint: cnrm/landing-zone:project/v0.4.1
    internal.config.kubernetes.io/path: project.yaml
spec:
  name: project-id # kpt-set: ${project-id}
  billingAccountRef:
    external: "AAAAAA-BBBBBB-CCCCCC" # kpt-set: ${billing-account-id}
  folderRef:
    name: name.of.folder # kpt-set: ${folder-name}
    namespace: hierarchy # kpt-set: ${folder-namespace}
`,
			err: "failed to decode Kptfile: yaml: unmarshal errors:\n  line 3: field annotations not found in type v1.KptFile",
		},
		{
			name: "missing path",
			r: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: project
  annotations:
    blueprints.cloud.google.com/title: Project blueprint
info:
  description: |
    A project and a project namespace in which to manage project resources with
    Config Connector.
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.1
      configPath: setters.yaml
---
apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
kind: Project
metadata:
  name: project-id # kpt-set: ${project-id}
  namespace: projects # kpt-set: ${projects-namespace}
  annotations:
    cnrm.cloud.google.com/auto-create-network: "false"
    cnrm.cloud.google.com/blueprint: cnrm/landing-zone:project/v0.4.1
    internal.config.kubernetes.io/path: project.yaml
spec:
  name: project-id # kpt-set: ${project-id}
  billingAccountRef:
    external: "AAAAAA-BBBBBB-CCCCCC" # kpt-set: ${billing-account-id}
  folderRef:
    name: name.of.folder # kpt-set: ${folder-name}
    namespace: hierarchy # kpt-set: ${folder-namespace}
`,
			err: "unable find resource path for Kptfile.v1.kpt.dev",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := getRNodesFromStr(t, tt.r)
			gotTitle, gotDoc, err := GenerateBlueprintReadme(res, "https://github.com/GoogleCloudPlatform/blueprints.git/catalog/", tt.pkgName)
			require := require.New(t)
			if tt.err != "" {
				require.EqualError(err, tt.err)
			} else {
				require.NoError(err)
				require.Equal(tt.wantTitle, gotTitle)
				require.Equal(strings.NewReplacer("¬", "`").Replace(tt.wantDoc), gotDoc)
			}

		})
	}
}

func TestGenerateResourcesSection(t *testing.T) {
	tests := []struct {
		name      string
		resources string
		want      string
	}{
		{name: "simple",
			resources: `apiVersion: compute.cnrm.cloud.google.com/v1beta1
kind: ComputeNetwork
metadata:
  name: network-name # kpt-set: ${network-name}
  namespace: networking # kpt-set: ${namespace}
  annotations:
    cnrm.cloud.google.com/blueprint: cnrm/landing-zone:networking/v0.4.0
    cnrm.cloud.google.com/project-id: project-id # kpt-set: ${project-id}
    internal.config.kubernetes.io/path: network.yaml
spec:
  autoCreateSubnetworks: false
  deleteDefaultRoutesOnCreate: false
  routingMode: GLOBAL
`,
			want: `## Resource References

- [ComputeNetwork](https://cloud.google.com/config-connector/docs/reference/resource-docs/compute/computenetwork)
`,
		},
		{name: "multiple",
			resources: `apiVersion: compute.cnrm.cloud.google.com/v1beta1
kind: ComputeNetwork
metadata:
  name: network-name # kpt-set: ${network-name}
  namespace: networking # kpt-set: ${namespace}
  annotations:
    cnrm.cloud.google.com/blueprint: cnrm/landing-zone:networking/v0.4.0
    cnrm.cloud.google.com/project-id: project-id # kpt-set: ${project-id}
    internal.config.kubernetes.io/path: network.yaml
spec:
  autoCreateSubnetworks: false
  deleteDefaultRoutesOnCreate: false
  routingMode: GLOBAL
---
apiVersion: v1
kind: Namespace
metadata:
  name: project-id # kpt-set: ${project-id}
  annotations:
    cnrm.cloud.google.com/project-id: project-id # kpt-set: ${project-id}
    internal.config.kubernetes.io/path: ns.yaml
---
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMServiceAccount
metadata:
  name: kcc-project-id # kpt-set: kcc-${project-id}
  namespace: config-control # kpt-set: ${management-namespace}
  annotations:
    cnrm.cloud.google.com/blueprint: cnrm/kcc-namespace/v0.4.1
    cnrm.cloud.google.com/project-id: management-project-id # kpt-set: ${management-project-id}
    internal.config.kubernetes.io/path: ns.yaml
spec:
  displayName: kcc-project-id # kpt-set: kcc-${project-id}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: cnrm-network-viewer-project-id # kpt-set: cnrm-network-viewer-${project-id}
  namespace: networking # kpt-set: ${networking-namespace}
roleRef:
  name: cnrm-viewer
  kind: ClusterRole
  apiGroup: rbac.authorization.k8s.io
subjects:
  - name: cnrm-controller-manager-project-id # kpt-set: cnrm-controller-manager-${project-id}
    namespace: cnrm-system
    kind: ServiceAccount
`,
			want: `## Resource References

- [ComputeNetwork](https://cloud.google.com/config-connector/docs/reference/resource-docs/compute/computenetwork)
- [IAMServiceAccount](https://cloud.google.com/config-connector/docs/reference/resource-docs/iam/iamserviceaccount)
- [Namespace](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#namespace-v1-core)
- [RoleBinding](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#rolebinding-v1-rbac-authorization-k8s-io)
`,
		},
		{name: "empty",
			resources: "",
			want: `## Resource References

This package has no top-level resources. See sub-packages.
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			res := getRNodesFromStr(t, tt.resources)
			r := blueprintReadme{content: &strings.Builder{}, filteredNodes: res}
			err := generateResourceRefsSection(&r)
			require.NoError(err)
			require.Equal(tt.want, r.string())
		})
	}
}

func TestGenerateSubPkgSection(t *testing.T) {
	tests := []struct {
		name string
		pkgs string
		want string
	}{
		{
			name: "simple",
			pkgs: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: rootpkg
  annotations:
    internal.config.kubernetes.io/path: Kptfile
---
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: foosubpkg
  annotations:
    internal.config.kubernetes.io/path: foosubpkg/Kptfile
---
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: barsubpkg
  annotations:
    internal.config.kubernetes.io/path: barsubpkg/Kptfile
`,
			want: `## Sub-packages

- [barsubpkg](barsubpkg)
- [foosubpkg](foosubpkg)
`,
		},
		{
			name: "multiple nested",
			pkgs: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: rootpkg
  annotations:
    internal.config.kubernetes.io/path: Kptfile
---
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: foosubpkg
  annotations:
    internal.config.kubernetes.io/path: foosubpkg/Kptfile
---
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: nested
  annotations:
    internal.config.kubernetes.io/path: foo/bar/Kptfile
---
apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: deepnested
  annotations:
    internal.config.kubernetes.io/path: foo/bar/baz/Kptfile
`,
			want: `## Sub-packages

- [deepnested](foo/bar/baz)
- [foosubpkg](foosubpkg)
- [nested](foo/bar)
`,
		},
		{
			name: "no subpkg",
			pkgs: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: rootpkg
  annotations:
    internal.config.kubernetes.io/path: Kptfile
`,
			want: `## Sub-packages

This package has no sub-packages.
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			nodes := getRNodesFromStr(t, tt.pkgs)
			kfs, err := findPkgs(nodes)
			require.NoError(err)
			r := blueprintReadme{content: &strings.Builder{}, bp: blueprint{kfs: kfs}}
			err = generateSubPkgSection(&r)
			require.NoError(err)
			require.Equal(tt.want, r.string())

		})
	}
}

func TestGenerateResourceTableSection(t *testing.T) {
	tests := []struct {
		name string
		r    string
		want string
		err  string
	}{
		{
			name: "simple",
			r: `apiVersion: v1
kind: Namespace
metadata:
  name: project-id # kpt-set: ${project-id}
  annotations:
    cnrm.cloud.google.com/project-id: project-id # kpt-set: ${project-id}
    internal.config.kubernetes.io/path: ns.yaml
---
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMServiceAccount
metadata:
  name: kcc-project-id # kpt-set: kcc-${project-id}
  namespace: config-control # kpt-set: ${management-namespace}
  annotations:
    cnrm.cloud.google.com/blueprint: cnrm/kcc-namespace/v0.4.1
    cnrm.cloud.google.com/project-id: management-project-id # kpt-set: ${management-project-id}
    internal.config.kubernetes.io/path: iam.yaml
spec:
  displayName: kcc-project-id # kpt-set: kcc-${project-id}
`,
			want: `## Resources

|   File   |            APIVersion             |       Kind        |      Name      |   Namespace    |
|----------|-----------------------------------|-------------------|----------------|----------------|
| ns.yaml  | v1                                | Namespace         | project-id     |                |
| iam.yaml | iam.cnrm.cloud.google.com/v1beta1 | IAMServiceAccount | kcc-project-id | config-control |
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := getRNodesFromStr(t, tt.r)
			r := blueprintReadme{content: &strings.Builder{}, filteredNodes: res}
			err := generateResourceTableSection(&r)
			require := require.New(t)
			if tt.err != "" {
				require.EqualError(err, tt.err)
			} else {
				require.NoError(err)
				require.Equal(tt.want, r.string())
			}

		})
	}
}

func TestGenerateSetterTableSection(t *testing.T) {
	tests := []struct {
		name string
		r    string
		want string
		err  string
	}{
		{
			name: "simple",
			r: `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: project
  annotations:
    blueprints.cloud.google.com/title: Project blueprint
    internal.config.kubernetes.io/path: Kptfile
info:
  description: |
    A project and a project namespace in which to manage project resources with
    Config Connector.
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.1
      configPath: setters.yaml
---
apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
kind: Project
metadata:
  name: project-id # kpt-set: ${project-id}
  namespace: projects # kpt-set: ${projects-namespace}
  annotations:
    cnrm.cloud.google.com/auto-create-network: "false"
    cnrm.cloud.google.com/blueprint: cnrm/landing-zone:project/v0.4.1
    internal.config.kubernetes.io/path: project.yaml
spec:
  name: project-id # kpt-set: ${project-id}
  billingAccountRef:
    external: "AAAAAA-BBBBBB-CCCCCC" # kpt-set: ${billing-account-id}
  folderRef:
    name: name.of.folder # kpt-set: ${folder-name}
    namespace: hierarchy # kpt-set: ${folder-namespace}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
  annotations:
    internal.config.kubernetes.io/path: setters.yaml
data:
  folder-name: name.of.folder
  project-id: project-id
  # These defaults can be kept
  folder-namespace: hierarchy
`,
			want: `## Setters

|        Name        |        Value         | Type | Count |
|--------------------|----------------------|------|-------|
| billing-account-id | AAAAAA-BBBBBB-CCCCCC | str  |     1 |
| folder-name        | name.of.folder       | str  |     1 |
| folder-namespace   | hierarchy            | str  |     1 |
| project-id         | project-id           | str  |     2 |
| projects-namespace | projects             | str  |     1 |
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes := getRNodesFromStr(t, tt.r)
			r := blueprintReadme{content: &strings.Builder{}, bp: blueprint{nodes: nodes}}
			err := generateSetterTableSection(&r)
			require := require.New(t)
			if tt.err != "" {
				require.EqualError(err, tt.err)
			} else {
				require.NoError(err)
				require.Equal(tt.want, r.string())
			}
		})
	}
}
