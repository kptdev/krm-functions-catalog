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
package searchreplace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

type test struct {
	name              string
	config            string
	input             string
	expectedResources string
	out               string
	errMsg            string
}

func TestSearchCommand(t *testing.T) {
	for _, tests := range [][]test{searchReplaceCases, putPatternCases} {
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

				s := &SearchReplace{}
				node, err := kyaml.Parse(test.config)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				inout := &kio.LocalPackageReadWriter{
					PackagePath:     baseDir,
					NoDeleteFiles:   true,
					PackageFileName: "Kptfile",
				}
				err = Decode(node, s)
				if !assert.NoError(t, err) {
					t.FailNow()
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
					return
				}

				if test.errMsg == "" && !assert.NoError(t, err) {
					t.FailNow()
				}

				actualResources, err := os.ReadFile(r.Name())
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				if !assert.Equal(t,
					strings.TrimSpace(test.expectedResources),
					strings.TrimSpace(string(actualResources))) {
					t.FailNow()
				}

				expectedOut := strings.ReplaceAll(test.out, "${filePath}", filepath.Base(r.Name()))
				if !assert.Equal(t, expectedOut, s.resultsString()) {
					t.FailNow()
				}
			})
		}
	}
}

func TestDecode(t *testing.T) {
	rn, err := kyaml.Parse(`data:
  by-value: foo
  put-values: bar`)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	err = Decode(rn, &SearchReplace{})
	if !assert.Error(t, err) {
		t.FailNow()
	}
	expected := `invalid matcher "put-values", must be one of ["by-value" "by-file-path" "by-value-regex" "by-path" "put-value" "put-comment"]`
	if !assert.Equal(t, expected, err.Error()) {
		t.FailNow()
	}

}
