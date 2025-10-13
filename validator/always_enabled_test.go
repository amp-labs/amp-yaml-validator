package validator

import (
	"testing"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/parser"
	"github.com/amp-labs/amp-yaml-validator/types"
)

func TestValidateAlwaysEnabledObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		object        openapi.IntegrationObject
		wantErrors    int
		expectedRules []string
	}{
		{
			name: "valid always-enabled object with required fields and schedule",
			object: openapi.IntegrationObject{
				ObjectName: "Account",
				RequiredFields: &[]openapi.IntegrationField{
					mustCreateField("id"),
					mustCreateField("name"),
				},
				Schedule: "*/15 * * * *",
			},
			wantErrors:    0,
			expectedRules: []string{},
		},
		{
			name: "missing required fields",
			object: openapi.IntegrationObject{
				ObjectName: "Account",
				Schedule:   "*/15 * * * *",
			},
			wantErrors:    1,
			expectedRules: []string{types.RuleAlwaysEnabledFields},
		},
		{
			name: "empty required fields",
			object: openapi.IntegrationObject{
				ObjectName:     "Account",
				RequiredFields: &[]openapi.IntegrationField{},
				Schedule:       "*/15 * * * *",
			},
			wantErrors:    1,
			expectedRules: []string{types.RuleAlwaysEnabledFields},
		},
		{
			name: "missing schedule",
			object: openapi.IntegrationObject{
				ObjectName: "Account",
				RequiredFields: &[]openapi.IntegrationField{
					mustCreateField("id"),
				},
				Schedule: "",
			},
			wantErrors:    1,
			expectedRules: []string{types.RuleAlwaysEnabledFields},
		},
		{
			name: "missing both required fields and schedule",
			object: openapi.IntegrationObject{
				ObjectName: "Account",
			},
			wantErrors:    2,
			expectedRules: []string{types.RuleAlwaysEnabledFields, types.RuleAlwaysEnabledFields},
		},
		{
			name: "prohibited mapToName in required fields",
			object: openapi.IntegrationObject{
				ObjectName: "Account",
				RequiredFields: &[]openapi.IntegrationField{
					mustCreateFieldWithMapToName("id", "account_id"),
				},
				Schedule: "*/15 * * * *",
			},
			wantErrors:    1,
			expectedRules: []string{types.RuleAlwaysEnabledFields},
		},
		{
			name: "multiple fields with one having prohibited mapToName",
			object: openapi.IntegrationObject{
				ObjectName: "Account",
				RequiredFields: &[]openapi.IntegrationField{
					mustCreateField("id"),
					mustCreateFieldWithMapToName("name", "account_name"),
					mustCreateField("email"),
				},
				Schedule: "*/15 * * * *",
			},
			wantErrors:    1,
			expectedRules: []string{types.RuleAlwaysEnabledFields},
		},
		{
			name: "multiple prohibited mapToName",
			object: openapi.IntegrationObject{
				ObjectName: "Account",
				RequiredFields: &[]openapi.IntegrationField{
					mustCreateFieldWithMapToName("id", "account_id"),
					mustCreateFieldWithMapToName("name", "account_name"),
				},
				Schedule: "*/15 * * * *",
			},
			wantErrors:    2,
			expectedRules: []string{types.RuleAlwaysEnabledFields, types.RuleAlwaysEnabledFields},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create validation context
			posMap := parser.NewPositionMap()
			path := "$.integrations[0].read.objects[0]"

			// Set up positions
			posMap.Set(path, parser.NewPosition(10, 1))
			posMap.Set(path+".requiredFields", parser.NewPosition(11, 3))
			posMap.Set(path+".schedule", parser.NewPosition(15, 3))

			if tt.object.RequiredFields != nil {
				for i := range *tt.object.RequiredFields {
					posMap.Set(path+".requiredFields["+string(rune(i+48))+"].mapToName", parser.NewPosition(12+i, 5))
				}
			}

			ctx := NewValidationContext(nil, posMap, nil)

			// Validate
			validateAlwaysEnabledObject(ctx, tt.object, path)

			// Check errors
			errors := ctx.GetErrors()
			if len(errors) != tt.wantErrors {
				t.Errorf("expected %d errors, got %d", tt.wantErrors, len(errors))
				for _, err := range errors {
					t.Logf("  Error: %s (rule: %s, path: %s)", err.Message, err.Rule, err.Path)
				}
			}

			// Check rules
			foundRules := make(map[string]int)
			for _, err := range errors {
				foundRules[err.Rule]++
			}

			for _, expectedRule := range tt.expectedRules {
				if foundRules[expectedRule] == 0 {
					t.Errorf("expected rule %s not found", expectedRule)
				}
			}
		})
	}
}

func TestValidateAlwaysEnabledObjectLineNumbers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		object        openapi.IntegrationObject
		path          string
		setupPosMap   func(pm parser.PositionMap, path string)
		expectedLines []int
	}{
		{
			name: "missing required fields at line 20",
			object: openapi.IntegrationObject{
				ObjectName: "Account",
				Schedule:   "*/15 * * * *",
			},
			path: "$.integrations[0].read.objects[0]",
			setupPosMap: func(pm parser.PositionMap, path string) {
				pm.Set(path+".requiredFields", parser.NewPosition(20, 5))
			},
			expectedLines: []int{20},
		},
		{
			name: "missing schedule at line 30",
			object: openapi.IntegrationObject{
				ObjectName: "Account",
				RequiredFields: &[]openapi.IntegrationField{
					mustCreateField("id"),
				},
			},
			path: "$.integrations[0].read.objects[1]",
			setupPosMap: func(pm parser.PositionMap, path string) {
				pm.Set(path+".schedule", parser.NewPosition(30, 5))
			},
			expectedLines: []int{30},
		},
		{
			name: "prohibited mapToName at line 25",
			object: openapi.IntegrationObject{
				ObjectName: "Account",
				RequiredFields: &[]openapi.IntegrationField{
					mustCreateFieldWithMapToName("id", "account_id"),
				},
				Schedule: "*/15 * * * *",
			},
			path: "$.integrations[0].read.objects[2]",
			setupPosMap: func(pm parser.PositionMap, path string) {
				pm.Set(path+".requiredFields[0].mapToName", parser.NewPosition(25, 7))
			},
			expectedLines: []int{25},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create position map
			posMap := parser.NewPositionMap()
			if tt.setupPosMap != nil {
				tt.setupPosMap(posMap, tt.path)
			}

			ctx := NewValidationContext(nil, posMap, nil)

			// Validate
			validateAlwaysEnabledObject(ctx, tt.object, tt.path)

			// Check that errors have correct line numbers
			errors := ctx.GetErrors()
			if len(errors) != len(tt.expectedLines) {
				t.Errorf("expected %d errors, got %d", len(tt.expectedLines), len(errors))
			}

			for i, expectedLine := range tt.expectedLines {
				if i >= len(errors) {
					break
				}
				if errors[i].Line != expectedLine {
					t.Errorf("error %d: expected line %d, got %d", i, expectedLine, errors[i].Line)
				}
			}
		})
	}
}

func TestValidateAlwaysEnabledObjectMultipleFields(t *testing.T) {
	t.Parallel()

	// Test object with multiple required fields, some valid and some with mapToName
	fields := []openapi.IntegrationField{
		mustCreateField("id"),
		mustCreateField("name"),
		mustCreateFieldWithMapToName("email", "contact_email"),
		mustCreateField("phone"),
		mustCreateFieldWithMapToName("address", "contact_address"),
	}

	object := openapi.IntegrationObject{
		ObjectName:     "Contact",
		RequiredFields: &fields,
		Schedule:       "*/10 * * * *",
	}

	// Create position map
	posMap := parser.NewPositionMap()
	path := "$.integrations[0].read.objects[0]"

	for i := range fields {
		posMap.Set(path+".requiredFields["+string(rune(i+48))+"].mapToName", parser.NewPosition(20+i, 7))
	}

	ctx := NewValidationContext(nil, posMap, nil)

	// Validate
	validateAlwaysEnabledObject(ctx, object, path)

	// Should have 2 errors (for fields with mapToName)
	errors := ctx.GetErrors()
	if len(errors) != 2 {
		t.Errorf("expected 2 errors for fields with mapToName, got %d", len(errors))
		for _, err := range errors {
			t.Logf("  Error: %s (rule: %s, line: %d)", err.Message, err.Rule, err.Line)
		}
	}

	// Verify all errors are for the correct rule
	for _, err := range errors {
		if err.Rule != types.RuleAlwaysEnabledFields {
			t.Errorf("expected rule %s, got %s", types.RuleAlwaysEnabledFields, err.Rule)
		}
	}
}

func TestValidateAlwaysEnabledObjectPathValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		object       openapi.IntegrationObject
		path         string
		expectedPath string
	}{
		{
			name: "missing required fields",
			object: openapi.IntegrationObject{
				ObjectName: "Account",
				Schedule:   "*/15 * * * *",
			},
			path:         "$.integrations[0].read.objects[0]",
			expectedPath: "$.integrations[0].read.objects[0].requiredFields",
		},
		{
			name: "missing schedule",
			object: openapi.IntegrationObject{
				ObjectName: "Account",
				RequiredFields: &[]openapi.IntegrationField{
					mustCreateField("id"),
				},
			},
			path:         "$.integrations[1].read.objects[2]",
			expectedPath: "$.integrations[1].read.objects[2].schedule",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			posMap := parser.NewPositionMap()
			ctx := NewValidationContext(nil, posMap, nil)

			validateAlwaysEnabledObject(ctx, tt.object, tt.path)

			errors := ctx.GetErrors()
			if len(errors) == 0 {
				t.Fatal("expected at least one error")
			}

			if errors[0].Path != tt.expectedPath {
				t.Errorf("expected path %s, got %s", tt.expectedPath, errors[0].Path)
			}
		})
	}
}

// Helper function to create a field without mapToName
func mustCreateField(fieldName string) openapi.IntegrationField {
	field := openapi.IntegrationFieldExistent{
		FieldName: fieldName,
	}

	result := openapi.IntegrationField{}
	if err := result.FromIntegrationFieldExistent(field); err != nil {
		panic(err)
	}

	return result
}

// Helper function to create a field with mapToName
func mustCreateFieldWithMapToName(fieldName, mapToName string) openapi.IntegrationField {
	field := openapi.IntegrationFieldExistent{
		FieldName: fieldName,
		MapToName: mapToName,
	}

	result := openapi.IntegrationField{}
	if err := result.FromIntegrationFieldExistent(field); err != nil {
		panic(err)
	}

	return result
}
