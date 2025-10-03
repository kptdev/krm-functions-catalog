package listsetters

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/kio"
)

func TestListSettersFilter(t *testing.T) {
	var tests = []struct {
		name           string
		resourceMap    map[string]string
		expectedResult []*Result
		errMsg         string
		warnings       []*WarnSetterDiscovery
	}{
		{
			name: "No setters",
			resourceMap: map[string]string{"test.yaml": `apiVersion: v1
kind: Service
metadata:
  name: my-app
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app
  name: mungebot`},
			expectedResult: []*Result{},
			warnings:       []*WarnSetterDiscovery{{"unable to find Kptfile, please include --include-meta-resources flag if a Kptfile is present"}},
		},
		{
			name: "Scalar Simple",
			resourceMap: map[string]string{"Kptfile": `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.2
      configMap:
        app: my-app
`, "test.yaml": `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
  name: mungebot
`},
			expectedResult: []*Result{{Name: "app", Value: "my-app", Count: 2, Type: "str"}},
		},
		{
			name: "Scalar Simple invalid kf",
			resourceMap: map[string]string{"Kptfile": `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
foo: bar
`, "test.yaml": `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
  name: mungebot
`},
			errMsg: "unable to read Kptfile: yaml: unmarshal errors:\n  line 11: field foo not found in type v1.KptFile",
		},
		{
			name: "Scalar Simple missing apply-setters",
			resourceMap: map[string]string{"Kptfile": `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/foo:v0.1
      configMap:
        app: my-app
`, "test.yaml": `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
  name: mungebot`},
			expectedResult: []*Result{{Name: "app", Value: "my-app", Count: 2, Type: "str"}},
			warnings:       []*WarnSetterDiscovery{{"unable to find apply-setters fn in Kptfile Pipeline.Mutators"}},
		},
		{
			name: "Scalar Simple missing kf pipeline",
			resourceMap: map[string]string{"Kptfile": `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
`, "test.yaml": `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
  name: mungebot`},
			expectedResult: []*Result{{Name: "app", Value: "my-app", Count: 2, Type: "str"}},
			warnings:       []*WarnSetterDiscovery{{"unable to find Pipeline declaration in Kptfile"}},
		},
		{
			name: "Scalar Simple no apply-setter fnConfig",
			resourceMap: map[string]string{"Kptfile": `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.2
`, "test.yaml": `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
  name: mungebot`},
			expectedResult: []*Result{{Name: "app", Value: "my-app", Count: 2, Type: "str"}},
			warnings:       []*WarnSetterDiscovery{{"unable to find ConfigMap or ConfigPath fnConfig for apply-setters"}},
		},
		{
			name: "Scalar Simple missing apply-setter configPath file",
			resourceMap: map[string]string{"Kptfile": `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.2
      configPath: setters.yaml
`, "test.yaml": `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
  name: mungebot`},
			expectedResult: []*Result{{Name: "app", Value: "my-app", Count: 2, Type: "str"}},
			warnings:       []*WarnSetterDiscovery{{"file setters.yaml doesn't exist, please ensure the file specified in \"configPath\" exists and retry"}},
		},
		{
			name: "Scalar with zero count setter",
			resourceMap: map[string]string{"Kptfile": `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.2
      configMap:
        app: my-app
        foo: bar
`, "test.yaml": `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
  name: mungebot
`},
			expectedResult: []*Result{{Name: "app", Value: "my-app", Count: 2, Type: "str"}, {Name: "foo", Value: "bar", Count: 0, Type: "str"}},
		},
		{
			name: "Scalar with two apply-setter configMap declarations",
			resourceMap: map[string]string{"Kptfile": `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.2
      configMap:
        app: my-app-old
        foo: bar
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.2
      configMap:
        app: my-app
        baz: qux
`, "test.yaml": `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
  name: mungebot
`},
			expectedResult: []*Result{{Name: "app", Value: "my-app", Count: 2, Type: "str"}, {Name: "foo", Value: "bar", Count: 0, Type: "str"}, {Name: "baz", Value: "qux", Count: 0, Type: "str"}},
		},
		{
			name: "Mapping Simple",
			resourceMap: map[string]string{"test.yaml": `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: # kpt-set: ${images}
    - ubuntu
    - hbase
 `},
			expectedResult: []*Result{{Name: "images", Value: "[hbase, ubuntu]", Count: 1, Type: "array"}},
			warnings:       []*WarnSetterDiscovery{{"unable to find Kptfile, please include --include-meta-resources flag if a Kptfile is present"}},
		},
		{
			name: "Mapping with kptfile and setterYml",
			resourceMap: map[string]string{"Kptfile": `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.2
      configPath: setters.yaml
`, "setters.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-setters-fn-config
data:
  images: |
      - ubuntu
      - hbase
`, "test.yaml": `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: # kpt-set: ${images}
    - ubuntu
    - hbase
 `},
			expectedResult: []*Result{{Name: "images", Value: "[ubuntu, hbase]", Count: 1, Type: "array"}},
		},
		{
			name: "Mapping with ConfigMap and ConfigPath apply-setter declarations",
			resourceMap: map[string]string{"Kptfile": `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: test
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.2
      configPath: setters.yaml
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.2
      configMap:
        baz: qux
`, "setters.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-setters-fn-config
data:
  images: |
      - ubuntu
      - hbase
`, "test.yaml": `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: # kpt-set: ${images}
    - ubuntu
    - hbase
 `},
			expectedResult: []*Result{{Name: "images", Value: "[ubuntu, hbase]", Count: 1, Type: "array"}, {Name: "baz", Value: "qux", Count: 0, Type: "str"}},
		},
		{
			name: "Scalar and Mapping",
			resourceMap: map[string]string{"test.yaml": `apiVersion: dns.cnrm.cloud.google.com/v1beta1
kind: DNSRecordSet
metadata:
  name: dnsrecordset-sample-mx # kpt-set: ${record-set-name}
  labels:
    typettl: "MX-300" # kpt-set: ${type}-${ttl}
spec:
  name: "mail.example.com." # kpt-set: ${domain}
  type: "MX" # kpt-set: ${type}
  ttl: 300 # kpt-set: ${ttl}
  managedZoneRef:
    name: dnsrecordset-dep-mx # kpt-set: ${managed-zone-name}
  rrdatas: # kpt-set: ${records}
    - "5 gmr-stmp-in.l.google.com."
    - "10 alt1.gmr-stmp-in.l.google.com."
    - "10 alt2.gmr-stmp-in.l.google.com."
    - "10 alt3.gmr-stmp-in.l.google.com."
    - "10 alt4.gmr-stmp-in.l.google.com."
`},
			expectedResult: []*Result{
				{Name: "record-set-name", Value: "dnsrecordset-sample-mx", Count: 1, Type: "str"},
				{Name: "type", Value: "MX", Count: 2, Type: "str"},
				{Name: "domain", Value: "mail.example.com.", Count: 1, Type: "str"},
				{Name: "managed-zone-name", Value: "dnsrecordset-dep-mx", Count: 1, Type: "str"},
				{Name: "ttl", Value: "300", Count: 2, Type: "int"},
				{Name: "records", Value: "[10 alt1.gmr-stmp-in.l.google.com., 10 alt2.gmr-stmp-in.l.google.com., 10 alt3.gmr-stmp-in.l.google.com., 10 alt4.gmr-stmp-in.l.google.com., 5 gmr-stmp-in.l.google.com.]", Count: 1, Type: "array"},
			},
			warnings: []*WarnSetterDiscovery{{"unable to find Kptfile, please include --include-meta-resources flag if a Kptfile is present"}},
		},
		{
			name: "with subpackages",
			resourceMap: map[string]string{"Kptfile": `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: project-package
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.2
      configPath: setters.yaml
`, "setters.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: apply-setters-fn-config
data:
  folder-name: name.of.folder
  folder-namespace: hierarchy
  networking-namespace: networking
  project-id: project-id
`, "test.yaml": `apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
kind: Project
metadata:
  name: project-id # kpt-set: ${project-id}
  namespace: projects # kpt-set: ${projects-namespace}
  annotations:
    cnrm.cloud.google.com/auto-create-network: "false"
spec:
  name: project-id # kpt-set: ${project-id}
  billingAccountRef:
    external: "AAAAAA-BBBBBB-CCCCCC" # kpt-set: ${billing-account-id}
  folderRef:
    name: name.of.folder # kpt-set: ${folder-name}
    namespace: hierarchy # kpt-set: ${folder-namespace}
`, "subpkg/vpc.yaml": `apiVersion: compute.cnrm.cloud.google.com/v1beta1
kind: ComputeNetwork
metadata:
  name: network-name # kpt-set: ${network-name}
  namespace: networking # kpt-set: ${networking-namespace}
  annotations:
    cnrm.cloud.google.com/project-id: project-id # kpt-set: ${project-id}
spec:
  autoCreateSubnetworks: false
  deleteDefaultRoutesOnCreate: false
  routingMode: GLOBAL
`, "subpkg/setters.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
data:
  foo: bar
  network-name: network-name
  networking-namespace: networking
  project-id: project-id
`, "subpkg/Kptfile": `apiVersion: kpt.dev/v1
kind: Kptfile
metadata:
  name: vpc-package
pipeline:
  mutators:
    - image: ghcr.io/kptdev/krm-functions-catalog/apply-setters:v0.2
      configPath: setters.yaml
`},
			expectedResult: []*Result{
				{Name: "billing-account-id", Value: "AAAAAA-BBBBBB-CCCCCC", Count: 1, Type: "str"},
				{Name: "folder-name", Value: "name.of.folder", Count: 1, Type: "str"},
				{Name: "folder-namespace", Value: "hierarchy", Count: 1, Type: "str"},
				{Name: "network-name", Value: "network-name", Count: 1, Type: "str"},
				{Name: "networking-namespace", Value: "networking", Count: 1, Type: "str"},
				{Name: "project-id", Value: "project-id", Count: 3, Type: "str"},
				{Name: "projects-namespace", Value: "projects", Count: 1, Type: "str"},
			},
		},
		{
			name: "multi type setters",
			resourceMap: map[string]string{"test.yaml": `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app # kpt-set: ${app}
    pi: 3.14 # kpt-set: ${pi}
  name: mungebot
spec:
  replicas: 3 # kpt-set: ${replicas}
  paused: true # kpt-set: ${paused}
`},
			expectedResult: []*Result{
				{Name: "app", Value: "my-app", Count: 2, Type: "str"},
				{Name: "paused", Value: "true", Count: 1, Type: "bool"},
				{Name: "pi", Value: "3.14", Count: 1, Type: "float"},
				{Name: "replicas", Value: "3", Count: 1, Type: "int"}},
			warnings: []*WarnSetterDiscovery{{"unable to find Kptfile, please include --include-meta-resources flag if a Kptfile is present"}},
		},
		{
			name: "multiple interpolated type setters",
			resourceMap: map[string]string{"test.yaml": `apiVersion: v1
kind: Service
metadata:
  name: my-app # kpt-set: ${app}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app-3 # kpt-set: ${app}-${replicas}
  name: mungebot
spec:
  replicas: 3 # kpt-set: ${replicas}
  paused: true # kpt-set: ${paused}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: my-app-3 # kpt-set: ${app}-${replicas}
  name: mungebot2
`},
			expectedResult: []*Result{
				{Name: "app", Value: "my-app", Count: 3, Type: "str"},
				{Name: "paused", Value: "true", Count: 1, Type: "bool"},
				{Name: "replicas", Value: "3", Count: 3, Type: "int"}},
			warnings: []*WarnSetterDiscovery{{"unable to find Kptfile, please include --include-meta-resources flag if a Kptfile is present"}},
		},
		{
			name: "ambiguous setter value picks first value",
			resourceMap: map[string]string{"test.yaml": `apiVersion: container.cnrm.cloud.google.com/v1beta1
kind: ContainerCluster
metadata:
  name: example-us-east4 # kpt-set: ${cluster-name}
  annotations:
    cnrm.cloud.google.com/project-id: platform-project-id # kpt-set: ${platform-project-id}
spec:
  subnetworkRef:
    name: platform-project-id-example-us-east4 # kpt-set: ${platform-project-id}-${cluster-name}
`},
			expectedResult: []*Result{
				{Name: "cluster-name", Value: "example-us-east4", Count: 2, Type: "str"},
				{Name: "platform-project-id", Value: "platform-project-id", Count: 2, Type: "str"}},
			warnings: []*WarnSetterDiscovery{{"unable to find Kptfile, please include --include-meta-resources flag if a Kptfile is present"}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require := require.New(t)
			pkgDir := setupInputs(t, test.resourceMap)
			//nolint:errcheck
			defer os.RemoveAll(pkgDir)

			ls := New()
			inout := &kio.LocalPackageReadWriter{
				PackagePath:        pkgDir,
				NoDeleteFiles:      true,
				PackageFileName:    "Kptfile",
				MatchFilesGlob:     append(kio.DefaultMatch, "Kptfile"),
				IncludeSubpackages: true,
			}
			err := kio.Pipeline{
				Inputs:  []kio.Reader{inout},
				Filters: []kio.Filter{&ls},
				Outputs: []kio.Writer{inout},
			}.Execute()
			if test.errMsg != "" {
				require.NotNil(err)
				require.Contains(err.Error(), test.errMsg)
			} else {
				require.NoError(err)
				require.ElementsMatch(ls.Warnings, test.warnings)
				actualResources := ls.GetResults()
				require.ElementsMatch(actualResources, test.expectedResult)
			}

		})
	}
}

func setupInputs(t *testing.T, resourceMap map[string]string) string {
	t.Helper()
	require := require.New(t)
	baseDir, err := os.MkdirTemp("", "")
	require.NoError(err)

	for rpath, data := range resourceMap {
		filePath := path.Join(baseDir, rpath)
		err = os.MkdirAll(path.Dir(filePath), os.ModePerm)
		require.NoError(err)
		err = os.WriteFile(path.Join(baseDir, rpath), []byte(data), 0644)
		require.NoError(err)
	}
	return baseDir
}

func TestCurrentSetterValues(t *testing.T) {
	var tests = []struct {
		name     string
		value    string
		pattern  string
		expected map[string]string
	}{
		{
			name:    "setter values from pattern 1",
			value:   "foo-dev-bar-us-east-1-baz",
			pattern: `foo-${environment}-bar-${region}-baz`,
			expected: map[string]string{
				"environment": "dev",
				"region":      "us-east-1",
			},
		},
		{
			name:    "setter values from pattern 2",
			value:   "foo-dev-bar-us-east-1-baz",
			pattern: `foo-${environment}-bar-${region}-baz`,
			expected: map[string]string{
				"environment": "dev",
				"region":      "us-east-1",
			},
		},
		{
			name:    "setter values from pattern 3",
			value:   "ghcr.io/my-app/my-app-backend:1.0.0",
			pattern: `${registry}/${app~!@#$%^&*()<>?:"|}/${app-image-name}:${app-image-tag}`,
			expected: map[string]string{
				"registry":             "ghcr.io",
				`app~!@#$%^&*()<>?:"|`: "my-app",
				"app-image-name":       "my-app-backend",
				"app-image-tag":        "1.0.0",
			},
		},
		{
			name:     "setter values from pattern unresolved",
			value:    "foo-dev-bar-us-east-1-baz",
			pattern:  `${image}:${tag}`,
			expected: map[string]string{},
		},
		{
			name:     "setter values from pattern unresolved 2",
			value:    "nginx:1.2",
			pattern:  `${image}${tag}`,
			expected: map[string]string{},
		},
		{
			name:     "setter values from pattern unresolved 3",
			value:    "my-project/nginx:1.2",
			pattern:  `${project-id}/${image}${tag}`,
			expected: map[string]string{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := currentSetterValues(test.pattern, test.value)
			require.Equal(t, test.expected, res)
		})
	}
}
