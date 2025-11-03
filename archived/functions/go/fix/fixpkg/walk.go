package fixpkg

import (
	"encoding/json"
	"strings"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kustomize/kyaml/fieldmeta"
	"sigs.k8s.io/kustomize/kyaml/openapi"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// visitor is implemented by structs which need to walk the configuration.
// visitor is provided to accept to walk configuration
type visitor interface {
	// visitScalar is called for each scalar field value on a resource
	// node is the scalar field value
	// path is the path to the field; path elements are separated by '.'
	// setterSchema is the OpenAPI schema of the setter from Kptfile
	visitScalar(node *yaml.RNode, setterSchema *openapi.ResourceSchema) error

	// visitMapping is called for each Mapping field value on a resource
	// node is the mapping field value
	visitMapping(node *yaml.RNode) error
}

// accept invokes the appropriate function on v for each field in object
// settersSchema is the schema equivalent of openAPI section in Kptfile
func accept(v visitor, object *yaml.RNode, settersSchema *spec.Schema) error {
	oa := getSchema(object, nil, "", settersSchema)
	return acceptImpl(v, object, "", oa, settersSchema)
}

// acceptImpl implements accept using recursion
func acceptImpl(v visitor, object *yaml.RNode, p string, oa *openapi.ResourceSchema, settersSchema *spec.Schema) error {
	switch object.YNode().Kind {
	case yaml.DocumentNode:
		// Traverse the child of the document
		return accept(v, yaml.NewRNode(object.YNode()), settersSchema)
	case yaml.MappingNode:
		if err := v.visitMapping(object); err != nil {
			return err
		}
		return object.VisitFields(func(node *yaml.MapNode) error {
			// get the schema for the field and propagate it
			fieldSchema := getSchema(node.Key, oa, node.Key.YNode().Value, settersSchema)
			// Traverse each field value
			return acceptImpl(v, node.Value, p+"."+node.Key.YNode().Value, fieldSchema, settersSchema)
		})
	case yaml.SequenceNode:
		// get the schema for the elements
		schema := getSchema(object, oa, "", settersSchema)
		return object.VisitElements(func(node *yaml.RNode) error {
			// Traverse each list element
			return acceptImpl(v, node, p, schema, settersSchema)
		})
	case yaml.ScalarNode:
		// Visit the scalar field
		setterSchema := getSchema(object, oa, "", settersSchema)
		return v.visitScalar(object, setterSchema)
	}
	return nil
}

// getSchema returns setter OpenAPI schema from Kptfile for a field.
// r is the Node to get the Schema for
// s is the provided schema for the field if known
// field is the name of the field
// settersSchema is the schema equivalent of openAPI section in Kptfile
func getSchema(r *yaml.RNode, s *openapi.ResourceSchema, field string, settersSchema *spec.Schema) *openapi.ResourceSchema {
	// get the override schema if it exists on the field
	fm := &fieldmeta.FieldMeta{SettersSchema: settersSchema}
	if err := Read(r, fm); err == nil && !fm.IsEmpty() {
		if fm.Schema.Ref.String() != "" {
			// resolve the reference
			s, err := openapi.Resolve(&fm.Schema.Ref, settersSchema)
			if err == nil && s != nil {
				fm.Schema = *s
			}
		}
		return &openapi.ResourceSchema{Schema: &fm.Schema}
	}

	// get the schema for a field of the node if the field is provided
	if s != nil && field != "" {
		return s.Field(field)
	}

	// get the schema for the elements if this is a list
	if s != nil && r.YNode().Kind == yaml.SequenceNode {
		return s.Elements()
	}

	// use the provided schema if present
	if s != nil {
		return s
	}

	if yaml.IsMissingOrNull(r) {
		return nil
	}

	// lookup the schema for the type
	m, _ := r.GetMeta()
	if m.Kind == "" || m.APIVersion == "" {
		return nil
	}
	return openapi.SchemaForResourceType(yaml.TypeMeta{Kind: m.Kind, APIVersion: m.APIVersion})
}

// Read reads the FieldMeta from a node
func Read(n *yaml.RNode, fm *fieldmeta.FieldMeta) error {
	// check for metadata on head and line comments
	comments := []string{n.YNode().LineComment, n.YNode().HeadComment}
	for _, c := range comments {
		if c == "" {
			continue
		}
		c := strings.TrimLeft(c, "#")

		// check for new short hand notation or fall back to openAPI ref format
		if !processShortHand(c, fm) {
			// if it doesn't Unmarshal that is fine, it means there is no metadata
			// other comments are valid, they just don't parse
			// TODO: consider more sophisticated parsing techniques similar to what is used
			// for go struct tags.
			if err := fm.Schema.UnmarshalJSON([]byte(c)); err != nil {
				// note: don't return an error if the comment isn't a fieldmeta struct
				return nil
			}
		}
		fe := fm.Schema.Extensions["x-kustomize"]
		if fe == nil {
			return nil
		}
		b, err := json.Marshal(fe)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, &fm.Extensions)
	}
	return nil
}

// processShortHand parses the comment for short hand ref, loads schema to fm
// and returns true if successful, returns false for any other cases and not throw
// error, as the comment might not be a setter ref
func processShortHand(comment string, fm *fieldmeta.FieldMeta) bool {
	input := map[string]string{}
	err := json.Unmarshal([]byte(comment), &input)
	if err != nil {
		return false
	}
	name := input["$kpt-set"]
	if name == "" {
		return false
	}

	// check if setter with the name exists, else check for a substitution
	// setter and substitution can't have same name in shorthand

	setterRef, err := spec.NewRef(fieldmeta.DefinitionsPrefix + fieldmeta.SetterDefinitionPrefix + name)
	if err != nil {
		return false
	}

	setterRefBytes, err := setterRef.MarshalJSON()
	if err != nil {
		return false
	}

	if _, err := openapi.Resolve(&setterRef, fm.SettersSchema); err == nil {
		setterErr := fm.Schema.UnmarshalJSON(setterRefBytes)
		return setterErr == nil
	}

	substRef, err := spec.NewRef(fieldmeta.DefinitionsPrefix + fieldmeta.SubstitutionDefinitionPrefix + name)
	if err != nil {
		return false
	}

	substRefBytes, err := substRef.MarshalJSON()
	if err != nil {
		return false
	}

	if _, err := openapi.Resolve(&substRef, fm.SettersSchema); err == nil {
		substErr := fm.Schema.UnmarshalJSON(substRefBytes)
		return substErr == nil
	}
	return false
}
