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
package docs

import (
	"fmt"
	"strings"
)

const (
	titleMarkerStart = "<!-- BEGINNING OF PRE-COMMIT-BLUEPRINT DOCS HOOK:TITLE -->"
	titleMarkerEnd   = "<!-- END OF PRE-COMMIT-BLUEPRINT DOCS HOOK:TITLE -->"
	bodyMarkerStart  = "<!-- BEGINNING OF PRE-COMMIT-BLUEPRINT DOCS HOOK:BODY -->"
	bodyMarkerEnd    = "<!-- END OF PRE-COMMIT-BLUEPRINT DOCS HOOK:BODY -->"
)

// InsertIntoReadme inserts the generated content and title into current readme
func InsertIntoReadme(title string, current string, generated string) (string, error) {
	lines := strings.Split(current, "\n")
	titleStart, titleStop, err := findInsertPoint(lines, titleMarkerStart, titleMarkerEnd)
	// a blueprint may have a custom title, only insert title if markers are found
	if err == nil {
		lines, err = insertBetweenIdx(lines, titleStart, titleStop, []string{title})
		if err != nil {
			return "", err
		}
	}
	bodyStart, bodyStop, err := findInsertPoint(lines, bodyMarkerStart, bodyMarkerEnd)
	// a blueprint may have custom body, only insert title if markers are found
	if err == nil {
		lines, err = insertBetweenIdx(lines, bodyStart, bodyStop, strings.Split(generated, "\n"))
		if err != nil {
			return "", err
		}
	}
	// return processed test with an extra newline
	return strings.Join(lines, "\n"), nil
}

// insertBetweenIdx replaces contents of original slice with toInsert between start and stop
func insertBetweenIdx(original []string, start int, stop int, toInsert []string) ([]string, error) {
	if start > len(original)-1 || stop > len(original)-1 || start < 0 || stop < 0 || start > stop {
		return nil, fmt.Errorf("unable insert, invalid start: %d or stop: %d", start, stop)
	}
	return append(original[:start], append(toInsert, original[stop+1:]...)...), nil
}

// findInsertPoint identifies positions of start and stop markers in a given slice of strings
func findInsertPoint(doc []string, docMarkerStart string, docMarkerEnd string) (int, int, error) {
	start, stop := -1, -1
	for i, line := range doc {
		if line == docMarkerStart {
			start = i
		}
		if line == docMarkerEnd {
			stop = i
		}
		if start != -1 && stop != -1 {
			return start + 1, stop - 1, nil
		}
	}
	if start == -1 {
		return start, stop, fmt.Errorf("unable to find start marker: %s", docMarkerStart)
	}
	return start, stop, fmt.Errorf("unable to find end marker: %s", docMarkerEnd)
}
