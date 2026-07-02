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
	"fmt"
	"strings"
)

const (
	maxLoggedValueLen   = 128
	redactedPlaceholder = "<redacted>"
)

// sensitiveKeywords triggers full redaction when matched in a field path or setter name.
var sensitiveKeywords = []string{
	"password",
	"passwd",
	"secret",
	"token",
	"apikey",
	"api_key",
	"api-key",
	"credential",
	"privatekey",
	"private_key",
	"cert",
	"certificate",
	"bearer",
	"auth",
}

// Sanitize returns a value safe to embed in ResultItem fields.
func Sanitize(value, fieldPath, setterPattern string) string {
	if looksLikeSecret(value, fieldPath, setterPattern) {
		return redactedPlaceholder
	}
	return truncateValue(value)
}

func looksLikeSecret(value, fieldPath, setterPattern string) bool {
	if hasSensitiveContext(fieldPath, setterPattern) {
		return true
	}
	return globalScanner.matches(value)
}

func hasSensitiveContext(fieldPath, setterPattern string) bool {
	if containsSensitiveKeyword(fieldPath) {
		return true
	}
	for _, setter := range unresolvedSetters(setterPattern) {
		if containsSensitiveKeyword(clean(setter)) {
			return true
		}
	}
	return false
}

func containsSensitiveKeyword(s string) bool {
	lower := strings.ToLower(s)
	for _, keyword := range sensitiveKeywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

func truncateValue(value string) string {
	runes := []rune(value)
	if len(runes) <= maxLoggedValueLen {
		return value
	}
	truncated := runes[:maxLoggedValueLen]
	return string(truncated) + fmt.Sprintf("... (truncated %d chars)", len(runes)-maxLoggedValueLen)
}
