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

package applysetters

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretRulesCompile(t *testing.T) {
	require.NotEmpty(t, globalScanner.rules, "expected at least one compiled secret rule")
	for _, rule := range secretRules {
		_, err := regexp.Compile(rule.pattern)
		assert.NoErrorf(t, err, "rule %q must compile under Go regexp", rule.id)
	}
}

func TestSecretScannerMatches(t *testing.T) {
	scanner := newSecretScanner()

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{name: "github pat", value: "ghp_abcdefghijklmnopqrstuvwxyz1234567890ab", expected: true},
		{name: "github fine grained pat", value: "github_pat_" + stringsRepeat("a", 82), expected: true},
		{name: "asia temp aws key invalid", value: "ASIATESTKEY0000001", expected: false},
		{name: "asia temp aws key valid", value: "ASIAUM2VG3ANQNP3IXQG", expected: true},
		{name: "mongodb credentialed uri", value: "mongodb+srv://user:pass@cluster.example.net/db", expected: true},
		{name: "postgres credentialed uri", value: "postgresql://user:secret@host:5432/db", expected: true},
		{name: "benign image ref", value: "nginx:1.7.9", expected: false},
		{name: "setter template only", value: "${image}:${tag}", expected: false},
		{name: "pem private key", value: "-----BEGIN RSA PRIVATE KEY-----\nMIIE\n-----END RSA PRIVATE KEY-----", expected: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, scanner.matches(test.value))
		})
	}
}

func stringsRepeat(s string, count int) string {
	out := make([]byte, 0, len(s)*count)
	for i := 0; i < count; i++ {
		out = append(out, s...)
	}
	return string(out)
}
