package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseDirective(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		comment     string
		wantRules   []string
		wantMatched bool
	}{
		{
			name:        "ignore all",
			comment:     "amp:ignore",
			wantRules:   []string{},
			wantMatched: true,
		},
		{
			name:        "ignore single rule",
			comment:     "amp:ignore[destination-exists]",
			wantRules:   []string{"destination-exists"},
			wantMatched: true,
		},
		{
			name:        "ignore multiple rules",
			comment:     "amp:ignore[rule1,rule2,rule3]",
			wantRules:   []string{"rule1", "rule2", "rule3"},
			wantMatched: true,
		},
		{
			name:        "ignore with spaces",
			comment:     "amp:ignore[rule1, rule2, rule3]",
			wantRules:   []string{"rule1", "rule2", "rule3"},
			wantMatched: true,
		},
		{
			name:        "ignore in longer comment",
			comment:     "This is ignored amp:ignore[test-rule] for testing",
			wantRules:   []string{"test-rule"},
			wantMatched: true,
		},
		{
			name:        "no directive",
			comment:     "Just a regular comment",
			wantRules:   nil,
			wantMatched: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rules, matched := parseDirective(tt.comment)
			require.Equal(t, tt.wantMatched, matched)
			require.Equal(t, tt.wantRules, rules)
		})
	}
}

func TestExtractDirectives(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		yaml         string
		wantDirCount int
		checkPath    string
		checkRule    string
		shouldIgnore bool
	}{
		{
			name: "inline comment",
			yaml: `
specVersion: 1.0.0  # amp:ignore[test-rule]
`,
			wantDirCount: 1,
			checkPath:    "$.specVersion",
			checkRule:    "test-rule",
			shouldIgnore: true,
		},
		{
			name: "head comment",
			yaml: `
# amp:ignore[another-rule]
specVersion: 1.0.0
`,
			wantDirCount: 1,
			checkPath:    "$.specVersion",
			checkRule:    "another-rule",
			shouldIgnore: true,
		},
		{
			name: "ignore all inline",
			yaml: `
specVersion: 1.0.0  # amp:ignore
`,
			wantDirCount: 1,
			checkPath:    "$.specVersion",
			checkRule:    "any-rule",
			shouldIgnore: true,
		},
		{
			name: "nested path",
			yaml: `
integrations:
  - name: test
    provider: salesforce  # amp:ignore[provider-rule]
`,
			wantDirCount: 1,
			checkPath:    "$.integrations[0].provider",
			checkRule:    "provider-rule",
			shouldIgnore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var node yaml.Node
			err := yaml.Unmarshal([]byte(tt.yaml), &node)
			require.NoError(t, err)

			dirMap := extractDirectives(&node)
			require.Len(t, dirMap, tt.wantDirCount, "unexpected number of directives")

			if tt.checkPath != "" {
				ignored := dirMap.ShouldIgnore(tt.checkPath, tt.checkRule)
				require.Equal(t, tt.shouldIgnore, ignored,
					"ShouldIgnore(%s, %s) = %v, want %v",
					tt.checkPath, tt.checkRule, ignored, tt.shouldIgnore)
			}
		})
	}
}

func TestDirectiveMapShouldIgnore(t *testing.T) {
	t.Parallel()

	dirMap := NewDirectiveMap()
	dirMap["$.path1"] = Directive{
		Path:  "$.path1",
		Rules: []string{"rule1", "rule2"},
	}
	dirMap["$.path2"] = Directive{
		Path:  "$.path2",
		Rules: []string{}, // Ignore all
	}

	tests := []struct {
		name         string
		path         string
		rule         string
		shouldIgnore bool
	}{
		{
			name:         "specific rule matched",
			path:         "$.path1",
			rule:         "rule1",
			shouldIgnore: true,
		},
		{
			name:         "specific rule not matched",
			path:         "$.path1",
			rule:         "rule3",
			shouldIgnore: false,
		},
		{
			name:         "ignore all - any rule",
			path:         "$.path2",
			rule:         "any-rule",
			shouldIgnore: true,
		},
		{
			name:         "path not in map",
			path:         "$.path3",
			rule:         "rule1",
			shouldIgnore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := dirMap.ShouldIgnore(tt.path, tt.rule)
			require.Equal(t, tt.shouldIgnore, result)
		})
	}
}
