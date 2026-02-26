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
package processor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestSleepProcessor(t *testing.T) {
	const delta = float64(time.Second) / 2

	testCases := map[string]struct {
		yaml             string
		expectedDuration time.Duration
		expectedErr      string
	}{
		"defaultDuration": {
			expectedDuration: DefaultDuration,
		},
		"customDuration": {
			yaml: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  duration: "2s"
`,
			expectedDuration: 2 * time.Second,
		},
		"backwardCompat": {
			yaml: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  sleepSeconds: "2"
`,
			expectedDuration: 2 * time.Second,
		},
		"configParseError": {
			yaml: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  duration: "2kk"
`,
			expectedErr: "couldn't parse",
		},
		"invalidDataError": {
			yaml: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data: "duration:2kk"
`,
			expectedErr: "data field contains neither",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var node *yaml.RNode
			var err error

			if tc.yaml != "" {
				node, err = yaml.Parse(tc.yaml)
				require.NoError(t, err, "failed to parse yaml")
			}

			p := &SleepProcessor{}
			rl := &framework.ResourceList{
				FunctionConfig: node,
			}

			start := time.Now()
			err = p.Process(rl)

			if tc.expectedErr == "" {
				require.NoError(t, err)
				duration := time.Since(start)
				assert.InDelta(t, tc.expectedDuration, duration, delta)
			} else {
				assert.ErrorContains(t, err, tc.expectedErr)
			}
		})
	}
}
