package parser

import (
	"fmt"
	"strings"
)

// pathComponent represents a single component of a YAML path.
type pathComponent struct {
	componentType string // "root", "object", or "array"
	name          string // Field name for objects, index for arrays
}

// PathBuilder builds JSONPath-style YAML paths.
type PathBuilder struct {
	components []pathComponent
}

// NewPathBuilder creates a new path builder starting with the root path "$".
func NewPathBuilder() *PathBuilder {
	return &PathBuilder{
		components: []pathComponent{{componentType: "root", name: "$"}},
	}
}

// PushObject adds an object field to the path (e.g., ".read").
func (pb *PathBuilder) PushObject(fieldName string) *PathBuilder {
	newBuilder := pb.Copy()
	newBuilder.components = append(newBuilder.components, pathComponent{
		componentType: "object",
		name:          fieldName,
	})

	return newBuilder
}

// PushArray adds an array index to the path (e.g., "[0]").
func (pb *PathBuilder) PushArray(index int) *PathBuilder {
	newBuilder := pb.Copy()
	newBuilder.components = append(newBuilder.components, pathComponent{
		componentType: "array",
		name:          fmt.Sprintf("%d", index),
	})

	return newBuilder
}

// String converts the path to a JSONPath string (e.g., "$.integrations[0].read").
func (pb *PathBuilder) String() string {
	if len(pb.components) == 0 {
		return "$"
	}

	var builder strings.Builder

	for _, comp := range pb.components {
		switch comp.componentType {
		case "root":
			builder.WriteString(comp.name)
		case "object":
			builder.WriteString(".")
			builder.WriteString(comp.name)
		case "array":
			builder.WriteString("[")
			builder.WriteString(comp.name)
			builder.WriteString("]")
		}
	}

	return builder.String()
}

// Copy creates an independent copy of the path builder for branching.
func (pb *PathBuilder) Copy() *PathBuilder {
	newComponents := make([]pathComponent, len(pb.components))
	copy(newComponents, pb.components)

	return &PathBuilder{
		components: newComponents,
	}
}
