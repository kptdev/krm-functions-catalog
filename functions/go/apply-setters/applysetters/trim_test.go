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
			value:    "AKIAIOSFODNN7EXAMPLE",
			expected: redactedPlaceholder,
		},
		{
			name:     "github token",
			value:    "ghp_123456789012345678901234567890123456",
			expected: redactedPlaceholder,
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
