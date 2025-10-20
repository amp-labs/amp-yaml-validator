package parser

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Directive represents an amp:ignore comment directive.
type Directive struct {
	Path  string   // YAML path where the directive applies
	Rules []string // List of rules to ignore (empty = ignore all)
}

// DirectiveMap maps YAML paths to directives for suppressing warnings.
type DirectiveMap map[string]Directive

// NewDirectiveMap creates a new empty directive map.
func NewDirectiveMap() DirectiveMap {
	return make(DirectiveMap)
}

// ShouldIgnore checks if a warning should be ignored for a given path and rule.
// It checks both exact path matches and parent path matches (prefix matching).
// For example, a directive on "$.integrations[0].read.objects[0]" will suppress
// warnings for "$.integrations[0].read.objects[0].schedule".
func (dm DirectiveMap) ShouldIgnore(path, rule string) bool {
	// First check exact path match
	if dm.matchesDirective(path, rule) {
		return true
	}

	// Check parent paths (progressively remove path segments)
	currentPath := path

	for {
		// Find last dot or bracket
		lastDot := strings.LastIndex(currentPath, ".")
		lastBracket := strings.LastIndex(currentPath, "[")

		cutPos := lastDot
		if lastBracket > lastDot {
			cutPos = lastBracket
		}

		if cutPos == -1 {
			break
		}

		currentPath = currentPath[:cutPos]
		if dm.matchesDirective(currentPath, rule) {
			return true
		}
	}

	return false
}

// matchesDirective checks if a specific path has a directive that matches the rule.
func (dm DirectiveMap) matchesDirective(path, rule string) bool {
	directive, exists := dm[path]
	if !exists {
		return false
	}

	// If no specific rules, ignore all warnings
	if len(directive.Rules) == 0 {
		return true
	}

	// Check if this specific rule should be ignored
	for _, ignoredRule := range directive.Rules {
		if ignoredRule == rule {
			return true
		}
	}

	return false
}

// Directive syntax patterns
var (
	// Matches: amp:ignore or amp:ignore[rule1,rule2]
	directiveRegex = regexp.MustCompile(`amp:ignore(?:\[([^\]]+)\])?`)
)

// parseDirective extracts directive information from a comment string.
// Returns the list of rules to ignore (empty slice = ignore all).
func parseDirective(comment string) ([]string, bool) {
	matches := directiveRegex.FindStringSubmatch(comment)
	if matches == nil {
		return nil, false
	}

	// If no brackets, ignore all rules
	if matches[1] == "" {
		return []string{}, true
	}

	// Parse comma-separated rules
	rulesStr := strings.TrimSpace(matches[1])

	rules := strings.Split(rulesStr, ",")
	for i, rule := range rules {
		rules[i] = strings.TrimSpace(rule)
	}

	return rules, true
}

// extractDirectives walks the YAML node tree and extracts amp:ignore directives.
// It returns a DirectiveMap that maps YAML paths to ignore directives.
func extractDirectives(node *yaml.Node) DirectiveMap {
	dirMap := NewDirectiveMap()
	path := NewPathBuilder()
	walkNodeForDirectives(node, path, dirMap)

	return dirMap
}

// walkNodeForDirectives recursively walks the YAML tree extracting directives.
func walkNodeForDirectives(node *yaml.Node, path *PathBuilder, dirMap DirectiveMap) {
	if node == nil {
		return
	}

	currentPath := path.String()

	// Check for inline comment on this node
	if node.LineComment != "" {
		if rules, ok := parseDirective(node.LineComment); ok {
			dirMap[currentPath] = Directive{
				Path:  currentPath,
				Rules: rules,
			}
		}
	}

	// Check for head comment (appears before this node - next-line directive)
	if node.HeadComment != "" {
		if rules, ok := parseDirective(node.HeadComment); ok {
			dirMap[currentPath] = Directive{
				Path:  currentPath,
				Rules: rules,
			}
		}
	}

	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) > 0 {
			walkNodeForDirectives(node.Content[0], path, dirMap)
		}

	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			if i+1 >= len(node.Content) {
				break
			}

			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			fieldName := keyNode.Value
			newPath := path.PushObject(fieldName)

			// Check key node for head comment (next-line directive)
			if keyNode.HeadComment != "" {
				if rules, ok := parseDirective(keyNode.HeadComment); ok {
					dirMap[newPath.String()] = Directive{
						Path:  newPath.String(),
						Rules: rules,
					}
				}
			}

			// Check key node for inline comment (e.g., "objects:  # amp:ignore")
			if keyNode.LineComment != "" {
				if rules, ok := parseDirective(keyNode.LineComment); ok {
					dirMap[newPath.String()] = Directive{
						Path:  newPath.String(),
						Rules: rules,
					}
				}
			}

			// Check value node for inline comment
			walkNodeForDirectives(valueNode, newPath, dirMap)
		}

	case yaml.SequenceNode:
		for idx, childNode := range node.Content {
			newPath := path.PushArray(idx)

			// Check child node for head comment (next-line directive on array item)
			if childNode.HeadComment != "" {
				if rules, ok := parseDirective(childNode.HeadComment); ok {
					dirMap[newPath.String()] = Directive{
						Path:  newPath.String(),
						Rules: rules,
					}
				}
			}

			walkNodeForDirectives(childNode, newPath, dirMap)
		}

	case yaml.ScalarNode:
		// Leaf node - comments already processed above
		return

	case yaml.AliasNode:
		if node.Alias != nil {
			walkNodeForDirectives(node.Alias, path, dirMap)
		}
	}
}
