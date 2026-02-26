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
package searchreplace

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
	"sigs.k8s.io/kustomize/kyaml/sets"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	ByValue       = "by-value"
	ByValueRegex  = "by-value-regex"
	ByPath        = "by-path"
	ByFilePath    = "by-file-path"
	PutValue      = "put-value"
	PutComment    = "put-comment"
	PathDelimiter = "."
)

// matchers returns the list of supported matchers
func matchers() []string {
	return []string{ByValue, ByFilePath, ByValueRegex, ByPath, PutValue, PutComment}
}

// SearchReplace struct holds the input parameters and results for
// Search and Replace operations on resource configs
type SearchReplace struct {
	// ByValue is the value of the field to be matched
	ByValue string

	// ByValueRegex is the value regex of the field to be matched
	ByValueRegex string

	// ByPath is the path of the field to be matched
	ByPath string

	// ByFilePath is the filepath of the resource to be matched
	ByFilePath string

	// Count is the number of matches
	Count int

	// PutValue is the value to be put at to field
	// filtered by path and/or value
	PutValue string

	// PutComment is the comment to be added at to field
	PutComment string

	// Results stores the results of executing the command
	Results []SearchResult

	// regex compiled regular expression for input by-value-regex
	regex *regexp.Regexp

	// filePath file path of resource
	filePath string
}

// SearchResult holds result of search and replace operation
type SearchResult struct {
	// FilePath is the file path of the matching field
	FilePath string

	// FieldPath is field path of the matching field
	FieldPath string

	// Value of the matching field
	Value string
}

// Filter performs the search and replace operation on all input nodes
func (sr *SearchReplace) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	if err := sr.validateMatchers(); err != nil {
		return nodes, err
	}

	// compile regex once so that it can be used everywhere
	if sr.ByValueRegex != "" {
		re, err := regexp.Compile(sr.ByValueRegex)
		if err != nil {
			return nodes, errors.Wrap(err)
		}
		sr.regex = re
	}

	// perform search/replace on all nodes
	for _, object := range nodes {
		_, err := sr.Perform(object)
		if err != nil {
			return nodes, err
		}
	}
	return nodes, nil
}

// Perform parses input node and performs search and replace operation on the node
func (sr *SearchReplace) Perform(object *yaml.RNode) (*yaml.RNode, error) {
	// get the filepath from the annotations to pass it to child methods
	filePath, _, err := kioutil.GetFileAnnotations(object)
	if err != nil {
		return object, err
	}

	if sr.ByFilePath != "" {
		match, err := doublestar.Match(sr.ByFilePath, filePath)
		if err != nil {
			return object, err
		}
		if !match {
			return object, nil
		}
	}

	sr.filePath = filePath
	if err != nil {
		return object, err
	}

	// check if value should be put by path and process it directly without needing
	// to traverse all elements of the node
	if sr.shouldPutValueByPath() {
		return object, sr.putValueByPath(object)
	}

	// traverse the node to perform search/put operation
	err = accept(sr, object)
	return object, err
}

/*
visitMapping parses mapping node and adds input comment to the mapping node key

e.g. for input of Mapping node

environments:
- dev
- stage

For input by-path = **.environments put-comment = 'kpt-set: ${env}', the node
is transformed to

environments: # kpt-set: ${env}
- stage
- prod
*/

func (sr *SearchReplace) visitMapping(object *yaml.RNode, path string) error {
	return object.VisitFields(func(node *yaml.MapNode) error {
		// the aim of this method is to put-comment to sequence node matched --by-path
		if sr.PutComment == "" {
			return nil
		}

		if node.IsNilOrEmpty() {
			return nil
		}

		if node.Value.YNode().Kind != yaml.SequenceNode {
			// return if it is not a sequence node
			return nil
		}

		key, err := node.Key.String()
		if err != nil {
			return err
		}

		// pathToKey refers to the path address of the key node ex: metadata.annotations
		// path is the path till parent node, pathToKey is obtained by appending child key
		pathToKey := fmt.Sprintf("%s.%s", path, strings.TrimSpace(key))
		if sr.pathMatch(strings.TrimPrefix(pathToKey, ".")) {
			node.Key.YNode().LineComment = sr.PutComment
			// change the style just to print the values to stdout e.g. [foo, bar]
			node.Value.YNode().Style = yaml.FlowStyle
			val, err := yaml.String(node.Value.YNode())
			if err != nil {
				return err
			}

			// change to folded style as it looks clean with comment in key node
			node.Value.YNode().Style = yaml.FoldedStyle
			res := SearchResult{
				FilePath:  sr.filePath,
				FieldPath: sr.ByPath + fmt.Sprintf(" # %s", sr.PutComment),
				Value:     strings.TrimSpace(val),
			}
			sr.Results = append(sr.Results, res)
			sr.Count++
		}
		return nil
	})
}

/*
visitScalar parses input scalar node and performs search and replace operation

e.g.for input of scalar node nginx:1.7.1 in the yaml node

apiVersion: v1
...
image: nginx:1.7.1

and for input by-value-regex = nginx-(.*) and put-value ubuntu-${1}

The yaml node is transformed to
apiVersion: v1
...
image: ubuntu:1.7.1
*/
func (sr *SearchReplace) visitScalar(object *yaml.RNode, path string) error {
	return sr.matchAndReplace(object.Document(), path)
}

// matchAndReplace matches the input scalar value against the input criteria,
// performs replace operation(if any) and appends the matched result
func (sr *SearchReplace) matchAndReplace(node *yaml.Node, path string) error {
	if node.Kind != yaml.ScalarNode {
		return nil
	}

	// check if the node matches search criteria
	if !sr.searchCriteriaMatch(node, path) {
		return nil
	}

	// increment the matched count
	sr.Count++

	// put comment if put-comment is provided as input
	if sr.PutComment != "" {
		var err error
		node.LineComment, err = resolvePattern(node.Value, sr.ByValueRegex, sr.PutComment)
		if err != nil {
			return err
		}
	}

	// put value if put-value is provided as input
	if sr.PutValue != "" {
		// TODO: pmarupaka Check if the new value honors the openAPI schema and/or
		// current field type, throw error if it doesn't
		var err error
		node.Value, err = resolvePattern(node.Value, sr.ByValueRegex, sr.PutValue)
		if err != nil {
			return err
		}
		// When encoding, if this tag is unset the value type will be
		// implied from the node properties
		node.Tag = yaml.NodeTagEmpty
	}

	// append the results of the search and replace operation
	if sr.filePath != "" {
		nodeVal, err := yaml.String(node)
		if err != nil {
			return err
		}
		res := SearchResult{
			FilePath:  sr.filePath,
			FieldPath: strings.TrimPrefix(path, PathDelimiter),
			Value:     strings.TrimSpace(nodeVal),
		}
		sr.Results = append(sr.Results, res)
	}

	return nil
}

// regexMatch checks if ValueRegex in SearchReplace struct matches with the input
// value, returns error if any
func (sr *SearchReplace) regexMatch(value string) bool {
	if sr.ByValueRegex == "" {
		return false
	}
	return sr.regex.Match([]byte(value))
}

// searchCriteriaMatch checks if the traversed node matches the input search criteria
func (sr *SearchReplace) searchCriteriaMatch(node *yaml.Node, path string) bool {
	// check if traversed path of node matches the input --by-path
	pathMatch := sr.pathMatch(path)

	// check if the node value matches with the input by-value-regex or the by-value
	// empty node values are not matched
	valueMatch := (sr.ByValue != "" && sr.ByValue == node.Value) || sr.regexMatch(node.Value)

	return (valueMatch && pathMatch) || // both value and path matched
		(valueMatch && sr.ByPath == "") || // match by value only
		(pathMatch && sr.ByValue == "" && sr.ByValueRegex == "") // match by path only
}

// putValueByPath puts the value in the user specified sr.ByPath
func (sr *SearchReplace) putValueByPath(object *yaml.RNode) error {
	path := strings.Split(sr.ByPath, PathDelimiter)
	// lookup(or create) node for n-1 path elements
	node, err := object.Pipe(yaml.LookupCreate(yaml.MappingNode, path[:len(path)-1]...))
	if err != nil {
		return errors.Wrap(err)
	}
	// set the last path element key with the input value
	sn := yaml.NewScalarRNode(sr.PutValue)
	// When encoding, if this tag is unset the value type will be
	// implied from the node properties
	sn.YNode().Tag = yaml.NodeTagEmpty
	err = node.PipeE(yaml.SetField(path[len(path)-1], sn))
	if err != nil {
		return errors.Wrap(err)
	}
	res := SearchResult{
		FilePath:  sr.filePath,
		FieldPath: sr.ByPath,
		Value:     sr.PutValue,
	}
	sr.Results = append(sr.Results, res)
	sr.Count++
	return nil
}

// shouldPutValueByPath returns true if only absolute path and literal are provided,
// so that the value can be directly put without needing to traverse the entire node,
// handles the case of adding non-existent field-value to node
func (sr *SearchReplace) shouldPutValueByPath() bool {
	return isAbsPath(sr.ByPath) &&
		!strings.Contains(sr.ByPath, "[") && // TODO: pmarupaka Support appending value for arrays
		sr.ByValue == "" &&
		sr.ByValueRegex == "" &&
		sr.PutValue != ""
}

// resolvePattern takes the field value of a node, valueRegex provided by
// user from by-value-regex, patternRegex provided by user from put-value/put-comment,
// and makes best effort to derive the corresponding capture groups and resolve the pattern
// refer to tests for expected behavior
func resolvePattern(fieldValue, valueRegex, patternRegex string) (string, error) {
	if valueRegex == "" {
		return patternRegex, nil
	}
	r, err := regexp.Compile(valueRegex)
	if err != nil {
		return "", errors.Errorf("failed to compile input pattern %q: %s", valueRegex, err.Error())
	}
	captureGroup := r.FindStringSubmatch(fieldValue)
	res := patternRegex
	for i, val := range captureGroup {
		if i == 0 {
			continue
		}
		res = strings.ReplaceAll(res, fmt.Sprintf("${%d}", i), val)
	}

	// make sure that all capture groups are resolved and throw error if they are not
	re := regexp.MustCompile(`\$\{([0-9]+)\}`)
	if re.Match([]byte(res)) {
		return "", errors.Errorf("unable to resolve capture groups")
	}

	return res, nil
}

// resultsString return the serialized string results
func (sr *SearchReplace) resultsString() string {
	var action string
	if sr.PutComment != "" || sr.PutValue != "" {
		action = "Mutated"
	} else {
		action = "Matched"
	}
	var out string
	for _, res := range sr.Results {
		out += fmt.Sprintf("%s\nfieldPath: %s\nvalue: %s\n\n", res.FilePath, res.FieldPath, res.Value)
	}
	out += fmt.Sprintf("%s %d field(s)\n", action, sr.Count)
	return out
}

// Decode decodes the input yaml RNode into SearchReplace struct
// returns error if input yaml RNode contains invalid matcher name inputs
func Decode(rn *yaml.RNode, fcd *SearchReplace) error {
	dm := rn.GetDataMap()
	if err := validateMatcherNames(dm); err != nil {
		return err
	}
	fcd.ByPath = dm[ByPath]
	fcd.ByValue = dm[ByValue]
	fcd.ByValueRegex = dm[ByValueRegex]
	fcd.PutValue = dm[PutValue]
	fcd.PutComment = dm[PutComment]
	fcd.ByFilePath = dm[ByFilePath]
	return nil
}

// validateMatcherNames validates the input matcher names
func validateMatcherNames(m map[string]string) error {
	matcherSet := sets.String{}
	matcherSet.Insert(matchers()...)
	for key := range m {
		if !matcherSet.Has(key) {
			return errors.Errorf("invalid matcher %q, must be one of %q", key, matchers())
		}
	}
	return nil
}

// validateMatchers validates the input matchers in SearchReplace struct
func (sr *SearchReplace) validateMatchers() error {
	if sr.ByValue != "" && sr.ByValueRegex != "" {
		return errors.Errorf(`only one of [%q, %q] can be provided`, ByValue, ByValueRegex)
	}
	return nil
}
