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
	"regexp"
	"strings"
	"unicode"
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

var (
	pemPattern    = regexp.MustCompile(`-----BEGIN [A-Z ]+-----`)
	jwtPattern    = regexp.MustCompile(`eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`)
	awsKeyPattern = regexp.MustCompile(`AKIA[0-9A-Z]{16}`)
	githubPattern = regexp.MustCompile(`ghp_[A-Za-z0-9]{36,}`)
	skPattern     = regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`)
	slackPattern  = regexp.MustCompile(`xox[baprs]-[A-Za-z0-9-]+`)
)

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
	if value == "" {
		return false
	}
	if pemPattern.MatchString(value) ||
		jwtPattern.MatchString(value) ||
		awsKeyPattern.MatchString(value) ||
		githubPattern.MatchString(value) ||
		skPattern.MatchString(value) ||
		slackPattern.MatchString(value) {
		return true
	}
	return isLongBase64Like(value)
}

func hasSensitiveContext(fieldPath, setterPattern string) bool {
	lowerPath := strings.ToLower(fieldPath)
	for _, keyword := range sensitiveKeywords {
		if strings.Contains(lowerPath, keyword) {
			return true
		}
	}
	for _, setter := range unresolvedSetters(setterPattern) {
		lowerName := strings.ToLower(clean(setter))
		for _, keyword := range sensitiveKeywords {
			if strings.Contains(lowerName, keyword) {
				return true
			}
		}
	}
	return false
}

// TODO: this looks a bit like magic
func isLongBase64Like(value string) bool {
	if len(value) < 40 {
		return false
	}
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSymbol := false
	base64Chars := 0
	for _, r := range value {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
			base64Chars++
		case unicode.IsLower(r):
			hasLower = true
			base64Chars++
		case unicode.IsDigit(r):
			hasDigit = true
			base64Chars++
		case r == '+' || r == '/' || r == '=':
			hasSymbol = true
			base64Chars++
		default:
			return false
		}
	}
	if float64(base64Chars)/float64(len(value)) < 0.9 {
		return false
	}
	diversity := 0
	if hasUpper {
		diversity++
	}
	if hasLower {
		diversity++
	}
	if hasDigit {
		diversity++
	}
	if hasSymbol {
		diversity++
	}
	return diversity >= 2
}

func truncateValue(value string) string {
	if len(value) <= maxLoggedValueLen {
		return value
	}
	return value[:maxLoggedValueLen] + fmt.Sprintf("... (truncated %d chars)", len(value)-maxLoggedValueLen)
}
