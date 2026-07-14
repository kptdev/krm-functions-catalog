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
	"math"
	"regexp"
	"strings"
	"unicode"
)

// False-positive guards adapted from privacy-filter secrets.go (MIT).
// https://github.com/packyme/privacy-filter/blob/main/filter/secrets.go

var (
	reTemplateVar = regexp.MustCompile(`^(?:\{\{[^{}]+\}\}|\$\{[^{}]+\}|%\{[^{}]+\}|<[^<>]+>)$`)
	reUUID        = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	reHexOnly     = regexp.MustCompile(`^[0-9a-fA-F]+$`)
	reHostPort    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]*\.[A-Za-z0-9-]+:`)
)

var commonPlaceholders = []string{
	"REPLACE_ME", "REPLACE_THIS", "REPLACE_WITH",
	"YOUR_KEY", "YOUR_TOKEN", "YOUR_SECRET", "YOUR_API_KEY", "YOUR_PASSWORD",
	"INSERT_HERE", "INSERT_KEY", "INSERT_TOKEN",
	"PLACEHOLDER", "EXAMPLE_KEY", "EXAMPLE_TOKEN",
	"TODO", "FIXME", "XXXX",
}

type secretScanner struct {
	rules []compiledSecretRule
}

var globalScanner = newSecretScanner()

func newSecretScanner() secretScanner {
	compiled := make([]compiledSecretRule, 0, len(secretRules))
	for _, rule := range secretRules {
		keywords := make([]string, len(rule.keywords))
		for i, kw := range rule.keywords {
			keywords[i] = strings.ToLower(kw)
		}
		compiled = append(compiled, compiledSecretRule{
			id:             rule.id,
			re:             regexp.MustCompile(rule.pattern),
			keywords:       keywords,
			entropyMin:     rule.entropyMin,
			secretGroup:    rule.secretGroup,
			allowURI:       rule.allowURI,
			skipAWSExample: rule.skipAWSExample,
		})
	}
	return secretScanner{rules: compiled}
}

func (s *secretScanner) matches(value string) bool {
	if value == "" || isSetterTemplateOnly(value) {
		return false
	}
	lower := strings.ToLower(value)
	for i := range s.rules {
		rule := &s.rules[i]
		if !ruleKeywordsMatch(rule.keywords, lower) {
			continue
		}
		for _, loc := range rule.re.FindAllStringSubmatchIndex(value, -1) {
			secret := extractSecret(value, loc, rule.secretGroup)
			if secret == "" {
				continue
			}
			if rule.entropyMin > 0 && shannonEntropy(secret) < rule.entropyMin {
				continue
			}
			if isBenignSecretCandidate(rule, secret, value, loc) {
				continue
			}
			return true
		}
	}
	return isLongBase64Like(value)
}

func ruleKeywordsMatch(keywords []string, lowerValue string) bool {
	if len(keywords) == 0 {
		return true
	}
	for _, kw := range keywords {
		if strings.Contains(lowerValue, kw) {
			return true
		}
	}
	return false
}

func extractSecret(value string, loc []int, group int) string {
	if len(loc) < 2 {
		return ""
	}
	if group == 0 {
		return value[loc[0]:loc[1]]
	}
	start, end := 2*group, 2*group+1
	if end >= len(loc) || loc[start] < 0 {
		return value[loc[0]:loc[1]]
	}
	return value[loc[start]:loc[end]]
}

func isBenignSecretCandidate(rule *compiledSecretRule, secret, fullValue string, loc []int) bool {
	if isTemplateVar(secret) {
		return true
	}
	if isLikelyPlaceholder(secret) {
		return true
	}
	if rule.skipAWSExample && strings.HasSuffix(strings.ToUpper(secret), "EXAMPLE") {
		return true
	}
	if isUUID(secret) || isHexHash(secret) {
		return true
	}
	if !rule.allowURI && looksLikeURLMatch(secret) {
		return true
	}
	if !rule.allowURI && loc[0] > 0 && isOnPathBoundary(fullValue, loc[0], loc[1]) {
		return true
	}
	return false
}

func isSetterTemplateOnly(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	if isTemplateVar(trimmed) {
		return true
	}
	setters := unresolvedSetters(trimmed)
	if len(setters) == 0 {
		return false
	}
	remainder := trimmed
	for _, setter := range setters {
		remainder = strings.ReplaceAll(remainder, setter, "")
	}
	for _, r := range strings.TrimSpace(remainder) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func isTemplateVar(s string) bool {
	return reTemplateVar.MatchString(strings.TrimSpace(s))
}

func isLikelyPlaceholder(s string) bool {
	upper := strings.ToUpper(s)
	for _, placeholder := range commonPlaceholders {
		if strings.Contains(upper, placeholder) {
			return true
		}
	}
	return false
}

func isUUID(s string) bool {
	return reUUID.MatchString(s)
}

func isHexHash(s string) bool {
	n := len(s)
	return (n == 32 || n == 40 || n == 64) && reHexOnly.MatchString(s)
}

func looksLikeURLMatch(s string) bool {
	if strings.Contains(s, "://") {
		return true
	}
	return reHostPort.MatchString(s)
}

func isOnPathBoundary(value string, start, end int) bool {
	const boundary = `/\:.@?=`
	if strings.ContainsAny(value[start:end], `/\\:`) {
		return true
	}
	if start > 0 && strings.ContainsRune(boundary, rune(value[start-1])) {
		return true
	}
	if end < len(value) && strings.ContainsRune(boundary, rune(value[end])) {
		return true
	}
	return false
}

func shannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	var freq [256]float64
	for i := 0; i < len(s); i++ {
		freq[s[i]]++
	}
	n := float64(len(s))
	ent := 0.0
	for _, count := range freq {
		if count > 0 {
			p := count / n
			ent -= p * math.Log2(p)
		}
	}
	return ent
}

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
