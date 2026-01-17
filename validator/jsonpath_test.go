package validator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIsNestedFieldPath tests the nested field path detection function.
func TestIsNestedFieldPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fieldName  string
		wantNested bool
	}{
		// Simple field names (not nested)
		{
			name:       "simple field name",
			fieldName:  "Name",
			wantNested: false,
		},
		{
			name:       "simple lowercase field",
			fieldName:  "email",
			wantNested: false,
		},
		{
			name:       "field with underscore",
			fieldName:  "first_name",
			wantNested: false,
		},
		{
			name:       "field with number",
			fieldName:  "field123",
			wantNested: false,
		},
		{
			name:       "UPPERCASE_FIELD",
			fieldName:  "CUSTOMER_ID",
			wantNested: false,
		},

		// Dot notation (nested)
		{
			name:       "dot notation - two levels",
			fieldName:  "Address.Street",
			wantNested: true,
		},
		{
			name:       "dot notation - three levels",
			fieldName:  "Contact.Address.City",
			wantNested: true,
		},
		{
			name:       "dot notation - lowercase",
			fieldName:  "user.email",
			wantNested: true,
		},

		// Bracket notation (nested)
		{
			name:       "bracket notation - array index",
			fieldName:  "items[0]",
			wantNested: true,
		},
		{
			name:       "bracket notation - property access",
			fieldName:  "data[id]",
			wantNested: true,
		},
		{
			name:       "bracket notation - complex",
			fieldName:  "records[0].name",
			wantNested: true,
		},
		{
			name:       "bracket notation - opening bracket only",
			fieldName:  "field[",
			wantNested: true,
		},
		{
			name:       "bracket notation - closing bracket only",
			fieldName:  "field]",
			wantNested: true,
		},

		// Edge cases
		{
			name:       "empty string",
			fieldName:  "",
			wantNested: false,
		},
		{
			name:       "just a dot",
			fieldName:  ".",
			wantNested: true,
		},
		{
			name:       "dot at start",
			fieldName:  ".field",
			wantNested: true,
		},
		{
			name:       "dot at end",
			fieldName:  "field.",
			wantNested: true,
		},
		{
			name:       "multiple dots",
			fieldName:  "a.b.c.d",
			wantNested: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isNestedFieldPath(tt.fieldName)
			assert.Equal(t, tt.wantNested, got,
				"isNestedFieldPath(%q) = %v, want %v", tt.fieldName, got, tt.wantNested)
		})
	}
}

// TestValidateJSONPathRules tests the validateJSONPathRules function integration.
func TestValidateJSONPathRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		yaml       string
		wantErrors int
	}{
		{
			name: "no field mappings - valid",
			yaml: `
specVersion: 1.0.0
integrations:
  - name: test-integration
    provider: salesforce
    read:
      objects:
        - objectName: Account
          destination: webhook
          schedule: "0 */12 * * *"
          selectedFields:
            Name: true
            Email: true
`,
			wantErrors: 0,
		},
		{
			name: "simple field mappings - valid",
			yaml: `
specVersion: 1.0.0
integrations:
  - name: test-integration
    provider: salesforce
    read:
      objects:
        - objectName: Account
          destination: webhook
          schedule: "0 */12 * * *"
          selectedFields:
            Name: true
            Email: true
          selectedFieldMappings:
            Name: account_name
            Email: contact_email
`,
			wantErrors: 0,
		},
		{
			name: "field mappings with underscores - valid",
			yaml: `
specVersion: 1.0.0
integrations:
  - name: test-integration
    provider: salesforce
    read:
      objects:
        - objectName: Account
          destination: webhook
          schedule: "0 */12 * * *"
          selectedFields:
            FirstName: true
            LastName: true
          selectedFieldMappings:
            FirstName: first_name
            LastName: last_name
`,
			wantErrors: 0,
		},
		{
			name: "no read action - valid (nothing to validate)",
			yaml: `
specVersion: 1.0.0
integrations:
  - name: test-integration
    provider: salesforce
    write:
      objects:
        - objectName: Account
          selectedFieldSettings:
            Name:
              writeOnCreate: always
              writeOnUpdate: always
`,
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := NewValidator()
			result, err := v.ValidateBytes(context.Background(), []byte(tt.yaml))
			assert.NoError(t, err)

			if tt.wantErrors == 0 {
				assert.Empty(t, result.Errors, "expected no errors, got: %v", result.Errors)
			} else {
				assert.Len(t, result.Errors, tt.wantErrors,
					"expected %d errors, got %d: %v", tt.wantErrors, len(result.Errors), result.Errors)
			}
		})
	}
}

// TestValidateJSONPathRules_Note tests that validateJSONPathRules is called but doesn't
// produce errors for bracket notation (since it's intentionally not supported/enforced).
func TestValidateJSONPathRules_NoEnforcement(t *testing.T) {
	t.Parallel()

	// This test verifies that bracket notation in field mappings does NOT produce errors
	// because bracket notation support is intentionally not implemented (design decision).
	yaml := `
specVersion: 1.0.0
integrations:
  - name: test-integration
    provider: salesforce
    read:
      objects:
        - objectName: Account
          destination: webhook
          schedule: "0 */12 * * *"
          selectedFields:
            Name: true
            Email: true
          selectedFieldMappings:
            Name: name
            Email: contact[0]  # Bracket notation - not enforced
`

	v := NewValidator(
		WithSkipAsyncValidation(), // Skip async warnings (destination exists, etc.)
	)
	result, err := v.ValidateBytes(context.Background(), []byte(yaml))
	assert.NoError(t, err)

	// Should have no errors because bracket notation is not enforced
	assert.Empty(t, result.Errors, "bracket notation should not produce errors (intentionally not enforced)")

	// Should have no warnings about JSONPath or field mappings
	for _, w := range result.Warnings {
		assert.NotContains(t, w.Rule, "jsonpath", "should not have JSONPath warnings")
		assert.NotContains(t, w.Rule, "field-mapping", "should not have field mapping warnings")
	}
}
