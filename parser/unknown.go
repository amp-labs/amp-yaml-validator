package parser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"gopkg.in/yaml.v3"
)

// UnknownKey describes a mapping key found in the YAML that has no corresponding
// field in the amp.yaml schema (openapi types). These are "orphan" keys: they are
// silently dropped during unmarshaling, so surfacing them helps catch typos and
// misplaced configuration.
type UnknownKey struct {
	Path   string // JSONPath-style path to the key (e.g. "$.integrations[0].read.scheduel")
	Key    string // The raw key name as written in the YAML
	Line   int    // Line of the key in the YAML file (1-based)
	Column int    // Column of the key in the YAML file (1-based)
}

// unionMembers maps oneOf/union struct types (whose only field is an unexported
// json.RawMessage) to the concrete member structs that make up the union. A key
// is considered valid for a union type if it is valid in ANY member.
var unionMembers = map[reflect.Type][]reflect.Type{
	reflect.TypeFor[openapi.IntegrationField](): {
		reflect.TypeFor[openapi.IntegrationFieldExistent](),
		reflect.TypeFor[openapi.IntegrationFieldMapping](),
	},
	reflect.TypeFor[openapi.HydratedIntegrationField](): {
		reflect.TypeFor[openapi.HydratedIntegrationFieldExistent](),
		reflect.TypeFor[openapi.IntegrationFieldMapping](),
	},
}

// DetectUnknownKeys parses the YAML into a node tree and reports every mapping
// key that does not correspond to a field in the amp.yaml schema. It is a
// non-fatal, additive check: keys not recognized by the schema are returned so
// callers can surface them as warnings.
func DetectUnknownKeys(yamlBytes []byte) ([]UnknownKey, error) {
	var node yaml.Node

	err := yaml.Unmarshal(yamlBytes, &node)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	var out []UnknownKey
	walkForUnknown(&node, reflect.TypeFor[openapi.Manifest](), NewPathBuilder(), &out)

	return out, nil
}

// walkForUnknown recursively walks a YAML node alongside the schema type it is
// expected to conform to, accumulating any keys that are not part of the schema.
func walkForUnknown(node *yaml.Node, expected reflect.Type, path *PathBuilder, out *[]UnknownKey) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) > 0 {
			walkForUnknown(node.Content[0], expected, path, out)
		}

	case yaml.AliasNode:
		walkForUnknown(node.Alias, expected, path, out)

	case yaml.MappingNode:
		walkMappingForUnknown(node, expected, path, out)

	case yaml.SequenceNode:
		elemType := deref(expected)
		if elemType.Kind() == reflect.Slice || elemType.Kind() == reflect.Array {
			item := elemType.Elem()
			for idx, child := range node.Content {
				walkForUnknown(child, item, path.PushArray(idx), out)
			}
		}
		// If the schema does not expect a sequence here, it is a type mismatch,
		// not an unknown key; leave that to the schema validators.

	case yaml.ScalarNode:
		return
	}
}

// walkMappingForUnknown checks each key of a mapping node against the expected
// schema type and recurses into values whose keys are recognized.
func walkMappingForUnknown(node *yaml.Node, expected reflect.Type, path *PathBuilder, out *[]UnknownKey) {
	expected = deref(expected)

	// Union (oneOf) types: a key is valid if it appears in any member. Members are
	// flat scalar shapes, so there is nothing to recurse into.
	if members, ok := unionMembers[expected]; ok {
		allowed := unionAllowedKeys(members)

		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			if isMergeKey(keyNode.Value) {
				continue
			}

			if _, known := allowed[strings.ToLower(keyNode.Value)]; !known {
				appendUnknown(out, path, keyNode)
			}
		}

		return
	}

	switch {
	case expected.Kind() == reflect.Struct:
		fields := jsonFieldMap(expected)

		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if isMergeKey(keyNode.Value) {
				continue
			}

			fieldType, known := fields[strings.ToLower(keyNode.Value)]
			if !known {
				appendUnknown(out, path, keyNode)

				continue // Do not descend into an unknown subtree.
			}

			walkForUnknown(valueNode, fieldType, path.PushObject(keyNode.Value), out)
		}

	case expected.Kind() == reflect.Map:
		// Maps allow arbitrary keys (additionalProperties); only the values are typed.
		valueType := expected.Elem()

		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]
			walkForUnknown(valueNode, valueType, path.PushObject(keyNode.Value), out)
		}

	default:
		// The schema expects a scalar or an opaque type (e.g. interface{}) here;
		// unknown-key detection does not apply.
		return
	}
}

// appendUnknown records an orphan key at the given key node's position.
func appendUnknown(out *[]UnknownKey, path *PathBuilder, keyNode *yaml.Node) {
	keyPath := path.PushObject(keyNode.Value)
	*out = append(*out, UnknownKey{
		Path:   keyPath.String(),
		Key:    keyNode.Value,
		Line:   keyNode.Line,
		Column: keyNode.Column,
	})
}

// jsonFieldMap returns a map of lowercased JSON field names to their Go types for
// a struct. Matching is case-insensitive to mirror encoding/json (which
// sigs.k8s.io/yaml uses), so case variants are not falsely reported as unknown.
func jsonFieldMap(structType reflect.Type) map[string]reflect.Type {
	fields := make(map[string]reflect.Type)

	for field := range structType.Fields() {
		if field.PkgPath != "" {
			continue // Unexported field.
		}

		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		switch name {
		case "-":
			continue
		case "":
			name = field.Name
		}

		fields[strings.ToLower(name)] = field.Type
	}

	return fields
}

// unionAllowedKeys returns the set of lowercased keys valid across all members of
// a union type.
func unionAllowedKeys(members []reflect.Type) map[string]struct{} {
	allowed := make(map[string]struct{})

	for _, member := range members {
		for key := range jsonFieldMap(member) {
			allowed[key] = struct{}{}
		}
	}

	return allowed
}

// deref unwraps pointer types to their element type.
func deref(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t
}

// isMergeKey reports whether a key is a YAML merge key ("<<"), which is a YAML
// mechanism rather than a schema field and must not be flagged.
func isMergeKey(key string) bool {
	return key == "<<"
}
