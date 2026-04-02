package createsetters

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestCreateSettersFilter(t *testing.T) {
	var tests = []struct {
		name              string
		config            string
		input             string
		expectedResources string
		errMsg            string
	}{
		{
			name: "set comment for array setter of flow style",
			config: `
data:
  env: |
    [foo, bar]
  name: nginx
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  env: [foo, bar]
 `,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${name}-deployment
  env: # kpt-set: ${env}
    - foo
    - bar
`,
		},
		{
			name: "set comment for scalar nodes",
			config: `
data:
  name: nginx
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  env: [foo, bar]
 `,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${name}-deployment
  env: [foo, bar]
`,
		},
		{
			name: "set comment for scalar and sequence nodes",
			input: `apiVersion: v1
kind: Service
metadata:
  name: myService
  namespace: foo 
image: nginx:1.7.1 
env: [foo, bar] 
`,
			config: `
data:
  app: myService
  ns: foo
  image: nginx
  tag: 1.7.1
  env: "[foo, bar]"
`,
			expectedResources: `apiVersion: v1
kind: Service
metadata:
  name: myService # kpt-set: ${app}
  namespace: foo # kpt-set: ${ns}
image: nginx:1.7.1 # kpt-set: ${image}:${tag}
env: # kpt-set: ${env}
  - foo # kpt-set: ${ns}
  - bar
`,
		},
		{
			name: "all scalar cases",
			input: `apiVersion: v1
kind: Deployment
metadata:
  name: ubuntu-deployment-1
spec:
  image: ubuntu
  app: "nginx:1.1.2"
  os:
    - ubuntu
    - mac
`,
			config: `
data:
  deploy: ubuntu-deployment
  env: ubuntu
  image: ngnix
  tag: 1.1.2
`,
			expectedResources: `apiVersion: v1
kind: Deployment
metadata:
  name: ubuntu-deployment-1 # kpt-set: ${deploy}-1
spec:
  image: ubuntu # kpt-set: ${env}
  app: "nginx:1.1.2" # kpt-set: nginx:${tag}
  os:
    - ubuntu # kpt-set: ${env}
    - mac
`,
		},

		{
			name: "scalar setter donot match",
			config: `
data:
  name: ubuntu
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  env: [foo, bar]
 `,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  env: [foo, bar]
`,
		},
		{
			name: "array setter with flow style donot match",
			config: `
data:
  env: |
    [foo, bar, pro]
  name: nginx
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  env: [foo, bar]
 `,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${name}-deployment
  env: [foo, bar]
`,
		},
		{
			name: "setter with no matching values",
			config: `
data:
  app: myService
  ns: foo
  images: |
    - ubuntu
    - linux
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images:
    - nginx
    - ubuntu
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images:
    - nginx
    - ubuntu
`,
		},
		{
			name: "FoldedStyle with matching values",
			config: `
data:
  images: |
    - nginx
    - ubuntu
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images:
    - nginx
    - ubuntu
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: # kpt-set: ${images}
    - nginx
    - ubuntu
`,
		},
		{
			name: "Multiple lines FoldedStyle",
			config: `
data:
  images: |
    - nginx
    - ubuntu
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: |
    - nginx
    - ubuntu
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: |
    - nginx
    - ubuntu
`,
		},
		{
			name: "Multiple lines ScalarNode",
			config: `
data:
  image: ubuntu
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: |
    nginx
    ubuntu
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: |
    nginx
    ubuntu
`,
		},
		{
			name: "containing overlap values",
			config: `
data:
  image: nginx
  name: image
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${image}-deployment
`,
		},
		{
			name: "Empty array values",
			config: `
data:
  image: "[]"
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: []
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: [] # kpt-set: ${image}
`,
		},
		{
			name: "Empty data map",
			config: `
data: {}
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: []
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: []
`,
			errMsg: "config map cannot be empty",
		},

		{
			name: "some values in sequence node match",
			config: `
data:
  app: myService
  image: nginx
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: [nginx, ubuntu]
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${image}-deployment
spec:
  images:
    - nginx # kpt-set: ${image}
    - ubuntu
`,
		},
		{
			name: "substrings",
			config: `
data:
  app: nginx
  image: nginx-abc
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-abc-deployment
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-abc-deployment # kpt-set: ${image}-deployment
`,
		},

		{
			name: "scalar setter donot match",
			config: `
data:
  name: ubuntu
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  env: [foo, bar]
 `,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  env: [foo, bar]
`,
		},
		{
			name: "setters donot match",
			config: `
data:
  app: myService
  ns: foo
  images: |
    - ubuntu
    - linux
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images:
    - nginx
    - ubuntu
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images:
    - nginx
    - ubuntu
`,
		},
		{
			name: "containing overlap values",
			config: `
data:
  image: nginx
  name: image
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${image}-deployment
`,
		},
		{
			name: "FlowStyle to FoldedStyle",
			config: `
data:
  image: "[nginx, ubuntu]"
  os: ubuntu
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: [nginx, ubuntu]
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: # kpt-set: ${image}
    - nginx
    - ubuntu # kpt-set: ${os}
`,
		},
		{
			name: "array with partial match",
			config: `
data:
  app: myService
  image: nginx
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  images: [nginx, ubuntu]
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment # kpt-set: ${image}-deployment
spec:
  images:
    - nginx # kpt-set: ${image}
    - ubuntu
`,
		},
		{
			name: "FlowStyle to FoldedStyle",
			config: `
data:
  image: "[nginx, ubuntu]"
  os: ubuntu
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: [nginx, ubuntu]
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: # kpt-set: ${image}
    - nginx
    - ubuntu # kpt-set: ${os}
`,
		},
		{
			name: "longest length match",
			config: `
data:
  app: development
  role: dev
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-development
spec:
  image: dev
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-development # kpt-set: nginx-${app}
spec:
  image: dev # kpt-set: ${role}
`,
		},
		{
			name: "skip internal kpt annotations",
			config: `
data:
  replicas: "4"
`,
			input: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  annotations:
    internal.config.kubernetes.io/package-path: /tmp/some-path-with-4-in-it
    config.kubernetes.io/local-config: "true"
spec:
  replicas: 4
`,
			expectedResources: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  annotations:
    internal.config.kubernetes.io/package-path: /tmp/some-path-with-4-in-it
    config.kubernetes.io/local-config: "true"
spec:
  replicas: 4 # kpt-set: ${replicas}
`,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			baseDir, err := os.MkdirTemp("", "")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			//nolint:errcheck
			defer os.RemoveAll(baseDir)

			r, err := os.CreateTemp(baseDir, "k8s-cli-*.yaml")
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			//nolint:errcheck
			defer os.Remove(r.Name())
			err = os.WriteFile(r.Name(), []byte(test.input), 0600)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			s := &CreateSetters{}
			node, err := kyaml.Parse(test.config)
			if err != nil {
				err = fmt.Errorf("parsing error in Config Map")
				if test.errMsg != "" && !assert.Equal(t, err.Error(), test.errMsg) {
					t.FailNow()
				} else if test.errMsg == "" {
					t.FailNow()
				}
				return
			}
			err = Decode(node, s)
			if err != nil {
				if test.errMsg != "" && !assert.Equal(t, err.Error(), test.errMsg) {
					t.FailNow()
				} else if test.errMsg == "" {
					t.FailNow()
				}
				return
			}
			inout := &kio.LocalPackageReadWriter{
				PackagePath:     baseDir,
				NoDeleteFiles:   true,
				PackageFileName: "Kptfile",
			}
			err = kio.Pipeline{
				Inputs:  []kio.Reader{inout},
				Filters: []kio.Filter{s},
				Outputs: []kio.Writer{inout},
			}.Execute()
			if test.errMsg != "" {
				if !assert.NotNil(t, err) {
					t.FailNow()
				}
				if !assert.Contains(t, err.Error(), test.errMsg) {
					t.FailNow()
				}
			}

			if test.errMsg == "" && !assert.NoError(t, err) {
				t.FailNow()
			}

			actualResources, err := os.ReadFile(r.Name())
			if !assert.NoError(t, err) {
				t.FailNow()
			}
			if !assert.Equal(t,
				test.expectedResources,
				string(actualResources)) {
				t.FailNow()
			}
		})
	}
}

type lineCommentTest struct {
	name    string
	value   string
	comment string
}

var resolveLineCommentCases = []lineCommentTest{
	{
		name:    "value matches multiple setters",
		value:   "foo-dev-bar-us-east-baz",
		comment: `foo-${role}-bar-${region}-baz`,
	},
	{
		name:    "setter matches part of a string",
		value:   "nginx:1.2.1",
		comment: `${app}:1.2.1`,
	},
	{
		name:    "value matches multiple setters",
		value:   "nginx:1.1.2",
		comment: `${app}:${tag}`,
	},
	{
		name:    "simple match setter",
		value:   "ubuntu",
		comment: `${env}`,
	},
	{
		name:    "no match",
		value:   "linux",
		comment: ``,
	},
	{
		name:    "longest length match",
		value:   "nginx-abc",
		comment: `${image}`,
	},
	{
		name:    "setters matches part of a string",
		value:   "nginx-base",
		comment: `${app}-base`,
	},
	{
		name:    "overlap case of setters",
		value:   "dev",
		comment: `${role}`,
	},
	{
		name:    "longest length match",
		value:   "development",
		comment: `${stage}`,
	},
}

var inputSetters = []ScalarSetter{
	{
		Name:  "tag",
		Value: "1.1.2",
	},
	{
		Name:  "role",
		Value: "dev",
	},
	{
		Name:  "stage",
		Value: "development",
	},
	{
		Name:  "app",
		Value: "nginx",
	},
	{
		Name:  "image",
		Value: "nginx-abc",
	},
	{
		Name:  "ns",
		Value: "role",
	},
	{
		Name:  "region",
		Value: "us-east",
	},
	{
		Name:  "env",
		Value: "ubuntu",
	},
}

func TestCurrentSetterValues(t *testing.T) {
	for _, tests := range [][]lineCommentTest{resolveLineCommentCases} {
		for i := range tests {
			test := tests[i]
			t.Run(test.name, func(t *testing.T) {
				sort.Sort(CompareSetters(inputSetters))
				replacerArgs := []string{}
				for _, setter := range inputSetters {
					replacerArgs = append(replacerArgs, setter.Value)
					replacerArgs = append(replacerArgs, fmt.Sprintf("${%s}", setter.Name))
				}
				Replacer := strings.NewReplacer(replacerArgs...)
				res, match := getLineComment(test.value, Replacer)
				if match {
					if !assert.Equal(t, test.comment, res) {
						t.FailNow()
					}
				} else {
					if !assert.Equal(t, test.comment, "") {
						t.FailNow()
					}
				}

			})
		}
	}
}
