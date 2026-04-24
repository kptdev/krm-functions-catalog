package pkg

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/cli-utils/pkg/object/mutation"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	tokenPattern              = "${ref%d}"
	sourceRegex               = `\${(?P<group>[^/]+)/((?P<version>[^/]+)/)?namespaces/(?P<namespace>[^/]+)/(?P<kind>[^/]+)/(?P<name>[^:]+):(?P<path>[^}]+)}`
	onlySourceRegex           = "^" + sourceRegex + "$"
	mutationCommentIdentifier = "# apply-time-mutation: "
	ResourceGroup             = "fn.kpt.dev"
	ResourceVersion           = "v1alpha1"
)

var (
	resourceReferencePattern = regexp.MustCompile(sourceRegex)
	onlyReferencePattern     = regexp.MustCompile(onlySourceRegex)
)

type CommentScanner struct {
	// ObjMeta is a reference to the object using metadata
	ObjMeta yaml.ResourceIdentifier
	// ObjFile is a reference to the object using file and object index
	ObjFile framework.File
}

type ScanResult struct {
	Path         string
	Value        any
	Comment      string
	Substitution mutation.FieldSubstitution
}

// Scan searches for mutation markup comments and parses them as substitutions.
func (cs *CommentScanner) Scan(obj *yaml.RNode) ([]ScanResult, error) {
	return cs.scanAny(obj, &PathBuilder{}, &indexer{})
}

func (cs *CommentScanner) scanAny(obj *yaml.RNode, fieldPath *PathBuilder, ix *indexer) ([]ScanResult, error) {
	var results []ScanResult
	switch obj.YNode().Kind {
	case yaml.MappingNode:
		// iterate over map (key->value)
		err := obj.VisitFields(func(node *yaml.MapNode) error {
			key, err := nodeValue(node.Key)
			if err != nil {
				return fmt.Errorf("invalid map key (path: %q, key: %q): %w", fieldPath, node.Key.YNode().Value, err)
			}
			fieldPath.Push(key)
			subResults, err := cs.scanAny(node.Value, fieldPath, ix)
			if err != nil {
				return err
			}
			results = append(results, subResults...)
			_ = fieldPath.Pop()
			return nil
		})
		return results, err
	case yaml.SequenceNode:
		// iterate over sequence (index->value)
		// Can't use VisitElements, because it doesn't provide the index...
		els, err := obj.Elements()
		if err != nil {
			return results, fmt.Errorf("invalid sequence (path: %q): %w", fieldPath, err)
		}
		for i, field := range els {
			fieldPath.Push(i)
			subResults, err := cs.scanAny(field, fieldPath, ix)
			if err != nil {
				return results, err
			}
			results = append(results, subResults...)
			_ = fieldPath.Pop()
		}
		return results, nil
	case yaml.ScalarNode:
		// scan the scalar node
		return cs.scanScalar(obj, fieldPath, ix)
	}
	return results, nil
}

func (cs *CommentScanner) scanScalar(node *yaml.RNode, fieldPath *PathBuilder, ix *indexer) ([]ScanResult, error) {
	var results []ScanResult

	comment := node.YNode().LineComment
	if comment == "" {
		// empty or missing comment
		return results, nil
	}

	fmt.Fprintf(os.Stderr, "Parsing comment (path: %q): %s\n", fieldPath, comment)
	// Check if comment is a mutation annotation.
	mutationPattern := extractMutationPattern(comment)
	if mutationPattern == "" {
		// The comment is not a mutation annotation.
		return results, nil
	}
	if !hasRef(mutationPattern) {
		// Mutation comment is tagged but no valid reference found.
		return results, fmt.Errorf("apply mutation comment found with no valid reference to source path")
	}

	resourceRef, refPath := commentToReference(mutationPattern)

	currentValue, err := nodeValue(node)
	if err != nil {
		return results, fmt.Errorf("invalid node value (path: %q): %w", fieldPath, err)
	}

	// replace the setter names in comment pattern with provided values
	replacementValue, replacementToken := commentToTokenField(mutationPattern, ix.Next())
	if replacementValue != "" {
		node.YNode().Value = replacementValue
	}

	fieldPathStr := fieldPath.String()

	results = append(results, ScanResult{
		Path:    fieldPathStr,
		Value:   currentValue,
		Comment: comment,
		Substitution: mutation.FieldSubstitution{
			SourceRef:  resourceRef,
			SourcePath: refPath,
			TargetPath: fieldPathStr,
			Token:      replacementToken,
		},
	})
	return results, nil
}

// hasRef returns whether or not the comment has a source reference embeded.
func hasRef(comment string) bool {
	return resourceReferencePattern.MatchString(comment)
}

// commentToReference parses a comment source reference to return the structured annotation fields.
func commentToReference(comment string) (mutation.ResourceReference, string) {
	outs := resourceReferencePattern.FindStringSubmatch(comment)
	group := outs[1]
	version := outs[3]
	retVal := mutation.ResourceReference{
		Namespace: outs[4],
		Kind:      outs[5],
		Name:      outs[6],
	}
	if version != "" {
		retVal.APIVersion = fmt.Sprintf("%s/%s", group, version)
	} else {
		retVal.Group = group
	}
	return retVal, outs[7]
}

// commentToTokenField replaces source reference strings with a replacement token.
// Returns replaced tokenized field value, and the replacement token to reference in the annotation.
func commentToTokenField(comment string, index int) (string, string) {
	// If the mutation comment is *only* a source reference, do not tokenize.
	if onlyReferencePattern.MatchString(comment) {
		return "", ""
	}
	token := fmt.Sprintf(tokenPattern, index)
	return resourceReferencePattern.ReplaceAllLiteralString(comment, token), token
}

// extractMutationPattern extracts the mutation pattern from the line comment.
// If the the line comment doesn't contain MutationCommentIdentifier
// prefix, then it returns an empty string.
func extractMutationPattern(lineComment string) string {
	if !strings.HasPrefix(lineComment, mutationCommentIdentifier) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(lineComment, mutationCommentIdentifier))
}

// nodeValue decodes a YAML node value of any type.
// Returns an error if the value is unparsable/invalid.
func nodeValue(node *yaml.RNode) (any, error) {
	var value any
	err := node.YNode().Decode(&value)
	if err != nil {
		return value, fmt.Errorf("failed to decode field value: %w", err)
	}
	return value, nil
}

type PathBuilder struct {
	Fields []any
}

// String returns a JSONPath expression to the specified field. This conforms
// to the path syntax used by the apply-time-mutation annotation.
func (pb *PathBuilder) String() string {
	return "$" + object.FieldPath(pb.Fields)
}

// Push appends a field to the path.
func (pb *PathBuilder) Push(field any) {
	pb.Fields = append(pb.Fields, field)
}

// Pop removes and returns the last field in the path. Returns nil if empty.
func (pb *PathBuilder) Pop() any {
	if len(pb.Fields) == 0 {
		return nil
	}
	last := pb.Fields[len(pb.Fields)-1]
	pb.Fields = pb.Fields[:len(pb.Fields)-1]
	return last
}

type indexer struct {
	index int
}

func (i *indexer) Next() int {
	i.index++
	return i.index - 1
}
