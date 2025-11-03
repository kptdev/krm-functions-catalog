package fixpkg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestFixV1alpha1ToV1(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.RemoveAll(dir)
	err = copyutil.CopyDir("../testdata/fix/nginx-v1alpha1", dir)
	assert.NoError(t, err)
	inout := &kio.LocalPackageReadWriter{
		PackagePath:    dir,
		MatchFilesGlob: append(kio.DefaultMatch, "Kptfile"),
	}
	f := &Fix{}
	err = kio.Pipeline{
		Inputs:  []kio.Reader{inout},
		Filters: []kio.Filter{f},
		Outputs: []kio.Writer{inout},
	}.Execute()
	assert.NoError(t, err)
	diff, err := copyutil.Diff(dir, "../testdata/fix/nginx-v1")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(diff.List()))
	results, err := yaml.Marshal(f.Results)
	assert.NoError(t, err)
	assert.Equal(t, `- filepath: Kptfile
  message: Transformed "packageMetadata" to "info"
- filepath: Kptfile
  message: Transformed "upstream" to "upstream" and "upstreamLock"
- filepath: Kptfile
  message: Added "ghcr.io/kptdev/krm-functions-catalog/set-annotations:v0.1" to mutators list, please move it to validators section if it is a validator function
- filepath: Kptfile
  message: Added "ghcr.io/kptdev/krm-functions-catalog/set-labels:v0.1" to mutators list, please move it to validators section if it is a validator function
- filepath: Kptfile
  message: Transformed "openAPI" definitions to "apply-setters" function
- filepath: hello-world/Kptfile
  message: Transformed "packageMetadata" to "info"
- filepath: hello-world/Kptfile
  message: Transformed "upstream" to "upstream" and "upstreamLock"
- filepath: hello-world/Kptfile
  message: Added "ghcr.io/kptdev/krm-functions-catalog/set-annotations:v0.1" to mutators list, please move it to validators section if it is a validator function
- filepath: hello-world/Kptfile
  message: Added "ghcr.io/kptdev/krm-functions-catalog/set-namespace:v0.1" to mutators list, please move it to validators section if it is a validator function
- filepath: hello-world/Kptfile
  message: Transformed "openAPI" definitions to "apply-setters" function
`, string(results))
}

func TestFixV1alpha2ToV1(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	assert.NoError(t, err)
	//nolint:errcheck
	defer os.RemoveAll(dir)
	err = copyutil.CopyDir("../testdata/fix/nginx-v1alpha2", dir)
	assert.NoError(t, err)
	inout := &kio.LocalPackageReadWriter{
		PackagePath:    dir,
		MatchFilesGlob: append(kio.DefaultMatch, "Kptfile"),
	}
	f := &Fix{}
	err = kio.Pipeline{
		Inputs:  []kio.Reader{inout},
		Filters: []kio.Filter{f},
		Outputs: []kio.Writer{inout},
	}.Execute()
	assert.NoError(t, err)
	diff, err := copyutil.Diff(dir, "../testdata/fix/nginx-v1")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(diff.List()))
	results, err := yaml.Marshal(f.Results)
	assert.NoError(t, err)
	assert.Equal(t, `- filepath: Kptfile
  message: Updated apiVersion to kpt.dev/v1
- filepath: setters.yaml
  message: Moved setters from configMap to configPath
- filepath: hello-world/Kptfile
  message: Updated apiVersion to kpt.dev/v1
- filepath: hello-world/setters.yaml
  message: Moved setters from configMap to configPath
`, string(results))
}

type settersNodeTest struct {
	name     string
	setters  map[string]string
	path     string
	expected string
	errMsg   string
}

var settersNodeFromSettersCases = []settersNodeTest{
	{
		name: "Create setters file with all types",
		path: `foo/bar`,
		setters: map[string]string{
			"environment":         "",
			"integer":             "10",
			"number":              "1.1",
			"boolean":             "true",
			"string":              "foo",
			"region":              "us-east-1",
			"flow-style-setter":   "[dev, prod]",
			"folded-style-setter": "- hi\n- hello",
		},
		expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: setters
  annotations:
    config.kubernetes.io/index: "0"
    config.kubernetes.io/path: foo/bar
data:
  boolean: "true"
  environment: ""
  flow-style-setter: |
    - dev
    - prod
  folded-style-setter: |-
    - hi
    - hello
  integer: "10"
  number: "1.1"
  region: us-east-1
  string: foo
`,
	},
	{
		name: "invalid flow-style sequence node",
		path: `foo/bar`,
		setters: map[string]string{
			"setter": "[dev, prod,",
		},
		errMsg: `failed to parse the array node value "[dev, prod," with error "yaml: line 1: did not find expected node content"`,
	},
}

func TestSettersNodeFromSetters(t *testing.T) {
	for i := range settersNodeFromSettersCases {
		test := settersNodeFromSettersCases[i]
		t.Run(test.name, func(t *testing.T) {
			res, err := SettersNodeFromSetters(test.setters, test.path)
			if test.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, test.errMsg, err.Error())
				return
			}
			actual, err := res.String()
			assert.NoError(t, err)
			if !assert.Equal(t, test.expected, actual) {
				t.FailNow()
			}
		})
	}
}
