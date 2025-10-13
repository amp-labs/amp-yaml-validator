package parser

import (
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"gopkg.in/yaml.v3"
	sigsyaml "sigs.k8s.io/yaml"
)

// ParseYAML parses YAML bytes into a Manifest and builds a position map.
// It uses a two-pass approach:
// 1. First pass: Unmarshal into yaml.v3.Node and walk the tree to build position map (preserves line/column info)
// 2. Second pass: Unmarshal into openapi.Manifest using sigs.k8s.io/yaml (handles JSON tags, same as server)
func ParseYAML(yamlBytes []byte) (*openapi.Manifest, PositionMap, error) {
	// First pass: Build position map from yaml.Node
	var node yaml.Node
	if err := yaml.Unmarshal(yamlBytes, &node); err != nil {
		return nil, nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	posMap := buildPositionMap(&node)

	// Second pass: Unmarshal into typed Manifest
	// Use sigs.k8s.io/yaml which handles JSON tags (same as server)
	var manifest openapi.Manifest
	if err := sigsyaml.Unmarshal(yamlBytes, &manifest); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal into Manifest: %w", err)
	}

	return &manifest, posMap, nil
}

// buildPositionMap recursively walks the yaml.Node tree to build a position map.
func buildPositionMap(node *yaml.Node) PositionMap {
	posMap := NewPositionMap()
	path := NewPathBuilder()
	walkNode(node, path, posMap)
	return posMap
}

// walkNode recursively walks a yaml.Node and records positions in the position map.
func walkNode(node *yaml.Node, path *PathBuilder, posMap PositionMap) {
	if node == nil {
		return
	}

	// Record position for the current node
	posMap.Set(path.String(), NewPosition(node.Line, node.Column))

	switch node.Kind {
	case yaml.DocumentNode:
		// Document nodes have a single child (the root node)
		if len(node.Content) > 0 {
			walkNode(node.Content[0], path, posMap)
		}

	case yaml.MappingNode:
		// Mapping nodes have key-value pairs
		for i := 0; i < len(node.Content); i += 2 {
			if i+1 >= len(node.Content) {
				break
			}
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			fieldName := keyNode.Value
			newPath := path.PushObject(fieldName)

			// Record position for the value node
			posMap.Set(newPath.String(), NewPosition(valueNode.Line, valueNode.Column))

			// Recursively walk the value node
			walkNode(valueNode, newPath, posMap)
		}

	case yaml.SequenceNode:
		// Sequence nodes have array elements
		for idx, childNode := range node.Content {
			newPath := path.PushArray(idx)

			// Record position for the array element
			posMap.Set(newPath.String(), NewPosition(childNode.Line, childNode.Column))

			// Recursively walk the child node
			walkNode(childNode, newPath, posMap)
		}

	case yaml.ScalarNode:
		// Scalar nodes are leaf values - already recorded position above
		return

	case yaml.AliasNode:
		// Alias nodes reference other nodes - follow the alias
		if node.Alias != nil {
			walkNode(node.Alias, path, posMap)
		}
	}
}
