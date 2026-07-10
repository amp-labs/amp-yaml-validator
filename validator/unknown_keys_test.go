package validator

import (
	"strings"
	"testing"

	"github.com/amp-labs/amp-yaml-validator/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unknownKeyWarnings returns only the unknown-key warnings from a result.
func unknownKeyWarnings(warnings []types.ValidationIssue) []types.ValidationIssue {
	var out []types.ValidationIssue

	for _, w := range warnings {
		if w.Rule == types.RuleUnknownKey {
			out = append(out, w)
		}
	}

	return out
}

func TestValidateUnknownKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		yaml string
		// wantKeys maps each expected orphan key name to the line it should be
		// reported on. The set must match exactly (no more, no fewer).
		wantKeys map[string]int
	}{
		{
			name: "no unknown keys in a valid manifest",
			yaml: `specVersion: 1.0.0
integrations:
  - name: t
    provider: salesforce
    read:
      objects:
        - objectName: account
          destination: webhook
          schedule: "*/10 * * * *"
`,
			wantKeys: map[string]int{},
		},
		{
			name: "top-level unknown key",
			yaml: `specVersion: 1.0.0
descrption: typo of description
integrations: []
`,
			wantKeys: map[string]int{"descrption": 2},
		},
		{
			name: "unknown key nested in a read object",
			yaml: `specVersion: 1.0.0
integrations:
  - name: t
    provider: salesforce
    read:
      objects:
        - objectName: account
          destination: webhook
          scheduel: "*/10 * * * *"
`,
			wantKeys: map[string]int{"scheduel": 9},
		},
		{
			name: "unknown key inside a union-typed field entry",
			yaml: `specVersion: 1.0.0
integrations:
  - name: t
    provider: salesforce
    read:
      objects:
        - objectName: account
          destination: webhook
          schedule: "*/10 * * * *"
          requiredFields:
            - feildName: name
`,
			wantKeys: map[string]int{"feildName": 11},
		},
		{
			name: "unknown parent key is reported once and does not cascade to its children",
			yaml: `specVersion: 1.0.0
integrations:
  - name: t
    provider: salesforce
    reed:
      objects:
        - objectName: account
          destination: webhook
          schedule: "*/10 * * * *"
`,
			wantKeys: map[string]int{"reed": 5},
		},
		{
			name: "case-insensitive keys are not flagged",
			yaml: `specVersion: 1.0.0
integrations:
  - name: t
    provider: salesforce
    read:
      objects:
        - objectName: account
          destination: webhook
          Schedule: "*/10 * * * *"
`,
			wantKeys: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			validator := NewValidator(WithSkipProviderValidation(), WithSkipAsyncValidation())
			result, err := validator.ValidateBytes(t.Context(), []byte(tt.yaml))
			require.NoError(t, err)

			got := unknownKeyWarnings(result.Warnings)
			require.Len(t, got, len(tt.wantKeys), "unexpected unknown-key warnings: %+v", got)

			for wantKey, wantLine := range tt.wantKeys {
				found := false

				for _, w := range got {
					if mentionsKey(w.Message, wantKey) {
						found = true

						assert.Equal(t, wantLine, w.Line, "wrong line for key %q", wantKey)
						assert.Equal(t, types.RuleUnknownKey, w.Rule)

						break
					}
				}

				assert.True(t, found, "expected an unknown-key warning for %q, got %+v", wantKey, got)
			}
		})
	}
}

// mentionsKey reports whether a warning message refers to the given key.
// Messages quote the key as %q, so we look for the quoted form.
func mentionsKey(message, key string) bool {
	return strings.Contains(message, `"`+key+`"`)
}

func TestValidateUnknownKeys_AmpIgnoreSuppression(t *testing.T) {
	t.Parallel()

	yaml := `specVersion: 1.0.0
integrations:
  - name: t
    provider: salesforce
    bogus: value # amp:ignore[unknown-key]
    stillBad: value
    read:
      objects:
        - objectName: account
          destination: webhook
          schedule: "*/10 * * * *"
`

	validator := NewValidator(WithSkipProviderValidation(), WithSkipAsyncValidation())
	result, err := validator.ValidateBytes(t.Context(), []byte(yaml))
	require.NoError(t, err)

	got := unknownKeyWarnings(result.Warnings)
	require.Len(t, got, 1, "expected only the un-ignored key to be reported: %+v", got)
	assert.True(t, mentionsKey(got[0].Message, "stillBad"))
}
