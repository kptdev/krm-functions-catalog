// Copyright 2026 The kpt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generate_folders

import "strings"

// Normalize normalizes a path into a K8s DNS subdomain compatible name.
func Normalize(parts []string) string {
	return normalize(strings.Join(parts, "."))
}

// normalize applies K8s DNS subdomain naming normalization to a string:
// 1. Convert to lowercase
// 2. Remove quotes
// 3. Replace underscores and spaces with dashes
// 4. Remove any remaining invalid characters
// 5. Collapse consecutive dashes/dots
// 6. Trim non-alphanumeric from start/end
// 7. Enforce 253-character max length
func normalize(name string) string {
	name = strings.ToLower(name)
	name = quoteRegex.ReplaceAllString(name, "")
	name = underscoreSpaceRegex.ReplaceAllString(name, "-")
	name = normalizeRegex.ReplaceAllString(name, "")
	name = consecutivePunctRegex.ReplaceAllString(name, "-")
	name = trimPunctRegex.ReplaceAllString(name, "")
	if len(name) > 253 {
		name = name[:253]
		name = trimPunctRegex.ReplaceAllString(name, "")
	}
	if name == "" {
		name = "unnamed"
	}
	return name
}

// filterNonInheritableAnnotations removes non-inheritable annotations.
func filterNonInheritableAnnotations(annotations map[string]string) map[string]string {
	if len(annotations) == 0 {
		return map[string]string{}
	}
	filtered := make(map[string]string)
	for k, v := range annotations {
		if !nonInheritableAnnotations[k] {
			filtered[k] = v
		}
	}
	return filtered
}

func clonePath(path []string) []string {
	cloned := make([]string, len(path))
	copy(cloned, path)
	return cloned
}
