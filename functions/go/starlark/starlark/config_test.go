package starlark

import (
	"testing"

	"github.com/kptdev/krm-functions-sdk/go/fn"
	"github.com/stretchr/testify/assert"
)

func TestStarlarkConfig(t *testing.T) {
	testcases := []struct {
		name         string
		config       string
		expectErrMsg string
	}{
		{
			name: "valid StarlarkRun",
			config: `apiVersion: fn.kpt.dev/v1alpha1
kind: StarlarkRun
metadata:
  name: my-star-fn
  namespace: foo
source: |
  def run(r, ns_value):
    for resource in r:
      resource["metadata"]["namespace"] = ns_value
  run(ctx.resource_list["items"], "baz")
`,
		},
		{
			name: "StarlarkRun missing Source",
			config: `apiVersion: fn.kpt.dev/v1alpha1
kind: StarlarkRun
metadata:
  name: my-star-fn
`,
			expectErrMsg: "`source` must not be empty",
		},
		{
			name: "valid ConfigMap",
			config: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-star-fn
data:
  source: |
    def run(r, ns_value):
      for resource in r:
        resource["metadata"]["namespace"] = ns_value
    run(ctx.resource_list["items"], "baz")
`,
		},
		{
			name: "ConfigMap missing source",
			config: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-star-fn
`,
			expectErrMsg: "`source` must not be empty",
		},
		{
			name: "ConfigMap with parameter but missing source",
			config: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-star-fn
data:
  param1: foo
`,
			expectErrMsg: "`source` must not be empty",
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			sr := &StarlarkRun{}
			ko, err := fn.ParseKubeObject([]byte(tc.config))
			assert.NoError(t, err)
			err = sr.Config(ko)
			if tc.expectErrMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectErrMsg)
			}
		})
	}
}
