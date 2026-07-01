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

import "regexp"

// Secret detection rules derived from Betterleaks v1.5.0 default config (MIT).
// https://github.com/betterleaks/betterleaks/blob/v1.5.0/config/betterleaks.toml
//
// Maintenance: when bumping rules, diff upstream betterleaks.toml for these IDs
// and copy regex/keywords/entropy hints. Verify with go test ./...
//
// Rule IDs sourced from Betterleaks: pem-block, jwt, aws-access-token,
// github-pat/github-oauth/github-app-token (merged), github-fine-grained-pat,
// gitlab-pat, gcp-api-key, openai-api-key, mongodb-connection-string,
// slack-bot-token, slack-app-token, stripe-access-token, azure-ad-client-secret.
// Custom rules: credentialed-uri, bearer-token, basic-auth.
// Long opaque base64 blobs are handled by isLongBase64Like in secret_scan.go.

type secretRule struct {
	id             string
	pattern        string
	keywords       []string
	entropyMin     float64
	secretGroup    int
	allowURI       bool
	skipAWSExample bool
}

var secretRules = []secretRule{
	{
		id:       "pem-block",
		pattern:  `-----BEGIN [A-Z ]+-----`,
		keywords: []string{"-----begin"},
	},
	{
		id:          "jwt",
		pattern:     `\b(ey[a-zA-Z0-9]{17,}\.ey[a-zA-Z0-9/_-]{17,}\.(?:[a-zA-Z0-9/_-]{10,}={0,2})?)(?:['"\x60]|[\s;]|$)`,
		keywords:    []string{"eyj"},
		entropyMin:  3.0,
		secretGroup: 1,
	},
	{
		id:             "aws-access-key-id",
		pattern:        `\b((?:A3T[A-Z0-9]|AKIA|ASIA|ABIA|ACCA)[A-Z2-7]{16})\b`,
		keywords:       []string{"a3t", "akia", "asia", "abia", "acca"},
		entropyMin:     3.0,
		secretGroup:    1,
		skipAWSExample: true,
	},
	{
		id:      "github-token",
		pattern: `(?:ghp|gho|ghu|ghs)_[0-9a-zA-Z]{36}`,
		keywords: []string{"ghp_", "gho_", "ghu_", "ghs_"},
	},
	{
		id:       "github-fine-grained-pat",
		pattern:  `github_pat_\w{82}`,
		keywords: []string{"github_pat_"},
	},
	{
		id:         "gitlab-pat",
		pattern:    `glpat-[\w-]{20}`,
		keywords:   []string{"glpat-"},
		entropyMin: 3.0,
	},
	{
		id:          "gcp-api-key",
		pattern:     `\b(AIza[\w-]{35})(?:['"\x60]|[\s;]|$)`,
		keywords:    []string{"aiza"},
		entropyMin:  4.0,
		secretGroup: 1,
	},
	{
		id:      "openai-api-key",
		pattern: `\b(sk-(?:proj|svcacct|admin)-(?:[A-Za-z0-9_-]{74}|[A-Za-z0-9_-]{58}|[A-Za-z0-9_-]{20})T3BlbkFJ(?:[A-Za-z0-9_-]{74}|[A-Za-z0-9_-]{58}|[A-Za-z0-9_-]{20})|sk-[a-zA-Z0-9]{20}T3BlbkFJ[a-zA-Z0-9]{20})(?:['"\x60]|[\s;]|$)`,
		keywords: []string{"t3blbkfj", "sk-proj", "sk-svcacct", "sk-admin"},
	},
	{
		id:          "mongodb-connection-string",
		pattern:     `\b(mongodb(?:\+srv)?://(?P<username>[!-9;-~]{3,50}):(?P<password>[!-?A-~]{3,88})@(?P<host>(?:[a-zA-Z0-9][\w.-]+|\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(?::\d{1,5})?(?:,(?:[a-zA-Z0-9][\w.-]+|\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(?::\d{1,5})?)*)/?(?:(?P<authdb>[\w-]+)?(?P<options>\?\w+=[\w@/.$-]+(?:&(?:amp;)?\w+=[\w@/.$-]+)*)?)?)(?:['"\s;\x60]|$)`,
		keywords:    []string{"mongodb://", "mongodb+srv://"},
		entropyMin:  4.0,
		secretGroup: 1,
		allowURI:    true,
	},
	{
		id:      "credentialed-uri",
		pattern: `(?i)\b(?:postgres(?:ql)?|mysql|redis|amqp)://[^\s/:@]{1,64}:[^\s@]{1,256}@`,
		keywords: []string{
			"postgres://", "postgresql://",
			"mysql://", "redis://", "amqp://",
		},
		allowURI: true,
	},
	{
		id:         "slack-bot-token",
		pattern:    `xoxb-[0-9]{10,13}-[0-9]{10,13}[a-zA-Z0-9-]*`,
		keywords:   []string{"xoxb"},
		entropyMin: 3.0,
	},
	{
		id:         "slack-app-token",
		pattern:    `(?i)xapp-\d-[A-Z0-9]+-\d+-[a-z0-9]+`,
		keywords:   []string{"xapp"},
		entropyMin: 2.0,
	},
	{
		id:          "stripe-secret-key",
		pattern:     `\b((?:sk|rk)_(?:test|live|prod)_[a-zA-Z0-9]{10,99})(?:['"\x60]|[\s;]|$)`,
		keywords:    []string{"sk_test", "sk_live", "sk_prod", "rk_test", "rk_live", "rk_prod"},
		entropyMin:  2.0,
		secretGroup: 1,
	},
	{
		id:          "azure-ad-client-secret",
		pattern:     `(?:^|[\\'"\x60\s>=:(,)])([a-zA-Z0-9_~.]{3}\dQ~[a-zA-Z0-9_~.-]{31,34})(?:$|[\\'"\x60\s<),])`,
		keywords:    []string{"q~"},
		secretGroup: 1,
	},
	{
		id:       "bearer-token",
		pattern:  `(?i)(?:\bBearer\s+[A-Za-z0-9._~+/=-]{10,}|Authorization:\s*Bearer\s+[A-Za-z0-9._~+/=-]{10,})`,
		keywords: []string{"bearer"},
	},
	{
		id:       "basic-auth",
		pattern:  `(?i)\bBasic\s+[A-Za-z0-9+/=]{8,}`,
		keywords: []string{"basic"},
	},
}

type compiledSecretRule struct {
	id             string
	re             *regexp.Regexp
	keywords       []string
	entropyMin     float64
	secretGroup    int
	allowURI       bool
	skipAWSExample bool
}
