// Copyright 2022 Google LLC
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

package main

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"strings"
)

func reverse(s []string) []string {
	var out []string
	for i := len(s) - 1; i >= 0; i-- {
		out = append(out, s[i])
	}
	return out
}

func GenerateProjectID(name string, parentPath string) (string, error) {
	parentPathTokens := strings.Split(parentPath, "/")
	fullProjectID := name

	parentPathTokens = reverse(parentPathTokens)

	for _, token := range parentPathTokens {
		fullProjectID += "-"

		// TODO: more normalization?
		fullProjectID += strings.ReplaceAll(token, ".", "-")
	}

	// 	Project ID requirements:

	// Must be 6 to 30 characters in length.
	// Can only contain lowercase letters, numbers, and hyphens.
	// Must start with a letter.
	// Cannot end with a hyphen.
	projectMaxLength := 30
	projectIDHash := sha512.Sum512([]byte(fullProjectID))

	// TODO: Do we want the suffix to be deterministic?  Might be better to make it random, from a confidentiality point of view
	// Also, multiple users are likely to have the same input values!
	suffixMaxLength := 4
	suffixLength := 0
	var suffix strings.Builder
	for _, r := range strings.ToLower(base64.StdEncoding.EncodeToString(projectIDHash[:])) {
		if ('0' <= r && r <= '9') || ('a' <= r && r <= 'z') {
			suffix.WriteRune(r)
			suffixLength++
			if suffixLength >= suffixMaxLength {
				break
			}
		}
	}

	// TODO: Don't change project id once set
	// TODO: Also handle collisions?
	if suffixLength != suffixMaxLength {
		// This should be vanishingly unlikely with sha512
		return "", fmt.Errorf("unable to generate suffix")
	}

	var projectID string
	if len(fullProjectID)+1+suffixLength > projectMaxLength {
		projectID = fullProjectID[:projectMaxLength-1-suffixLength] + "-" + suffix.String()
	} else {
		projectID = fullProjectID + "-" + suffix.String()
	}

	// TODO: Check that name is short enough that project starts with prefix?
	// (have to be careful because we normalize name a bit, not a simple StartsWith)

	return projectID, nil
}
