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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	longValue := strings.Repeat("a", 500)
	pemValue := "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----"
	jwtValue := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	tests := []struct {
		name          string
		value         string
		fieldPath     string
		setterPattern string
		expected      string
	}{
		{
			name:     "short benign value",
			value:    "my-app",
			expected: "my-app",
		},
		{
			name:     "long yaml blob",
			value:    longValue,
			expected: longValue[:maxLoggedValueLen] + "... (truncated 372 chars)",
		},
		{
			name:     "pem cert in value",
			value:    pemValue,
			expected: redactedPlaceholder,
		},
		{
			name:     "jwt in value",
			value:    jwtValue,
			expected: redactedPlaceholder,
		},
		{
			name:      "path heuristic",
			value:     "short",
			fieldPath: "spec.tls.privateKey",
			expected:  redactedPlaceholder,
		},
		{
			name:          "setter name heuristic",
			value:         "short",
			setterPattern: "${db-password}",
			expected:      redactedPlaceholder,
		},
		{
			name:     "aws key prefix",
			value:    "AKIAUM2VG3ANQNP3IXQG",
			expected: redactedPlaceholder,
		},
		{
			name:     "asia temporary aws key",
			value:    "ASIAUM2VG3ANQNP3IXQG",
			expected: redactedPlaceholder,
		},
		{
			name:     "github token",
			value:    "ghp_abcdefghijklmnopqrstuvwxyz1234567890ab",
			expected: redactedPlaceholder,
		},
		{
			name:     "github fine grained pat",
			value:    "github_pat_" + strings.Repeat("a", 82),
			expected: redactedPlaceholder,
		},
		{
			name:     "mongodb connection string",
			value:    "mongodb+srv://user:pass@cluster.example.net/db",
			expected: redactedPlaceholder,
		},
		{
			name:     "postgres connection string",
			value:    "postgresql://user:secret@host:5432/db",
			expected: redactedPlaceholder,
		},
		{
			name:     "benign image ref",
			value:    "nginx:1.7.9",
			expected: "nginx:1.7.9",
		},
		{
			name:     "setter template only",
			value:    "${image}:${tag}",
			expected: "${image}:${tag}",
		},
		{
			name:     "empty value",
			value:    "",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Sanitize(test.value, test.fieldPath, test.setterPattern)
			assert.Equal(t, test.expected, got)
		})
	}
}
