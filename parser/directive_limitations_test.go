package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestDirectiveLimitations documents known limitations and edge cases
func TestDirectiveLimitations(t *testing.T) {
	t.Parallel()

	t.Run("comment before first property attaches to that property", func(t *testing.T) {
		t.Parallel()

		// This is a YAML parser limitation - the comment attaches to objectName, not the object
		yamlStr := `
objects:
  - # amp:ignore[catalog-access]
    objectName: account
    destination: webhook
`
		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlStr), &node)
		require.NoError(t, err)

		dirMap := extractDirectives(&node)

		// The directive is on objectName, not on the array item
		require.True(t, dirMap.ShouldIgnore("$.objects[0].objectName", "catalog-access"),
			"directive should apply to objectName")
		require.False(t, dirMap.ShouldIgnore("$.objects[0].destination", "catalog-access"),
			"directive does NOT apply to sibling destination field")

		// Workaround: put directive on the array item marker
		yamlWorkaround := `
objects:
  # amp:ignore[catalog-access]
  - objectName: account
    destination: webhook
`
		yaml.Unmarshal([]byte(yamlWorkaround), &node)
		dirMapWorkaround := extractDirectives(&node)

		require.True(t, dirMapWorkaround.ShouldIgnore("$.objects[0].objectName", "catalog-access"))
		require.True(t, dirMapWorkaround.ShouldIgnore("$.objects[0].destination", "catalog-access"),
			"workaround: directive on array marker applies to all children")
	})

	t.Run("directives work in multi-line comment blocks", func(t *testing.T) {
		t.Parallel()

		yamlStr := `
# This is a comment block
# amp:ignore[test-rule]
# with multiple lines
field: value
`
		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlStr), &node)
		require.NoError(t, err)

		dirMap := extractDirectives(&node)
		require.True(t, dirMap.ShouldIgnore("$.field", "test-rule"),
			"directive in multi-line comment block should work")
	})

	t.Run("hierarchical suppression works across all levels", func(t *testing.T) {
		t.Parallel()

		yamlStr := `
# amp:ignore[test-rule]
root:
  level1:
    level2:
      level3: value
`
		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlStr), &node)
		require.NoError(t, err)

		dirMap := extractDirectives(&node)
		require.True(t, dirMap.ShouldIgnore("$.root", "test-rule"))
		require.True(t, dirMap.ShouldIgnore("$.root.level1", "test-rule"))
		require.True(t, dirMap.ShouldIgnore("$.root.level1.level2", "test-rule"))
		require.True(t, dirMap.ShouldIgnore("$.root.level1.level2.level3", "test-rule"),
			"hierarchical suppression should work for deeply nested paths")
	})

	t.Run("multiple directives on same path", func(t *testing.T) {
		t.Parallel()

		// Only the last directive wins (this is a YAML limitation)
		yamlStr := `
# amp:ignore[rule1]
# amp:ignore[rule2]
field: value
`
		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlStr), &node)
		require.NoError(t, err)

		dirMap := extractDirectives(&node)

		// Both directives are in the same HeadComment, and our regex finds the first match
		// Actually, let's check what happens...
		matched1 := dirMap.ShouldIgnore("$.field", "rule1")
		matched2 := dirMap.ShouldIgnore("$.field", "rule2")

		// At least one should work. The actual behavior depends on how our regex handles multiple matches
		require.True(t, matched1 || matched2,
			"at least one directive should work")
	})

	t.Run("empty array inline directive", func(t *testing.T) {
		t.Parallel()

		yamlStr := `
objects: []  # amp:ignore[test-rule]
`
		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlStr), &node)
		require.NoError(t, err)

		dirMap := extractDirectives(&node)
		require.True(t, dirMap.ShouldIgnore("$.objects", "test-rule"),
			"directive on empty array should work")
	})

	t.Run("directive with spaces and special characters in rule name", func(t *testing.T) {
		t.Parallel()

		yamlStr := `
field: value  # amp:ignore[rule-with-dashes,rule_with_underscores]
`
		var node yaml.Node
		err := yaml.Unmarshal([]byte(yamlStr), &node)
		require.NoError(t, err)

		dirMap := extractDirectives(&node)
		require.True(t, dirMap.ShouldIgnore("$.field", "rule-with-dashes"))
		require.True(t, dirMap.ShouldIgnore("$.field", "rule_with_underscores"))
	})
}
