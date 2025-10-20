package validator

import (
	"testing"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/parser"
	"github.com/amp-labs/amp-yaml-validator/types"
)

func TestValidateSpecVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		specVersion  string
		wantError    bool
		expectedRule string
		setupPosMap  func(pm parser.PositionMap)
	}{
		{
			name:         "valid spec version 1.0.0",
			specVersion:  "1.0.0",
			wantError:    false,
			expectedRule: "",
			setupPosMap: func(pm parser.PositionMap) {
				pm.Set("$.specVersion", parser.NewPosition(1, 1))
			},
		},
		{
			name:         "missing spec version",
			specVersion:  "",
			wantError:    true,
			expectedRule: types.RuleSpecVersion,
			setupPosMap: func(pm parser.PositionMap) {
				pm.Set("$.specVersion", parser.NewPosition(1, 1))
			},
		},
		{
			name:         "invalid spec version 0.9.0",
			specVersion:  "0.9.0",
			wantError:    true,
			expectedRule: types.RuleSpecVersion,
			setupPosMap: func(pm parser.PositionMap) {
				pm.Set("$.specVersion", parser.NewPosition(2, 1))
			},
		},
		{
			name:         "invalid spec version 2.0.0",
			specVersion:  "2.0.0",
			wantError:    true,
			expectedRule: types.RuleSpecVersion,
			setupPosMap: func(pm parser.PositionMap) {
				pm.Set("$.specVersion", parser.NewPosition(3, 1))
			},
		},
		{
			name:         "invalid spec version random string",
			specVersion:  "invalid",
			wantError:    true,
			expectedRule: types.RuleSpecVersion,
			setupPosMap: func(pm parser.PositionMap) {
				pm.Set("$.specVersion", parser.NewPosition(4, 1))
			},
		},
	}

	//nolint:varnamelen // tt is idiomatic in table-driven tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create manifest with spec version
			manifest := &openapi.Manifest{
				SpecVersion: tt.specVersion,
			}

			// Create position map
			posMap := parser.NewPositionMap()
			if tt.setupPosMap != nil {
				tt.setupPosMap(posMap)
			}

			// Create validation context
			ctx := NewValidationContext(manifest, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)

			// Validate spec version
			validateSpecVersion(ctx)

			// Check errors
			errors := ctx.GetErrors()
			//nolint:nestif // Test assertion complexity is acceptable
			if tt.wantError {
				if len(errors) == 0 {
					t.Error("expected error but got none")
				} else {
					if errors[0].Rule != tt.expectedRule {
						t.Errorf("expected rule %s, got %s", tt.expectedRule, errors[0].Rule)
					}
					// Verify that line number is set correctly
					pos := posMap.GetOrDefault("$.specVersion")
					if errors[0].Line != pos.Line {
						t.Errorf("expected line %d, got %d", pos.Line, errors[0].Line)
					}
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("expected no errors, got %d: %v", len(errors), errors)
				}
			}
		})
	}
}

func TestValidateSpecVersionLineNumbers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		specVersion string
		line        int
		column      int
	}{
		{
			name:        "line 5 column 3",
			specVersion: "0.5.0",
			line:        5,
			column:      3,
		},
		{
			name:        "line 10 column 1",
			specVersion: "",
			line:        10,
			column:      1,
		},
		{
			name:        "line 1 column 15",
			specVersion: "invalid",
			line:        1,
			column:      15,
		},
	}

	//nolint:varnamelen // tt is idiomatic in table-driven tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create manifest
			manifest := &openapi.Manifest{
				SpecVersion: tt.specVersion,
			}

			// Create position map with specific line/column
			posMap := parser.NewPositionMap()
			posMap.Set("$.specVersion", parser.NewPosition(tt.line, tt.column))

			// Create validation context
			ctx := NewValidationContext(manifest, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)

			// Validate
			validateSpecVersion(ctx)

			// Check that error has correct line number
			errors := ctx.GetErrors()
			if len(errors) == 0 {
				t.Fatal("expected error for invalid spec version")
			}

			if errors[0].Line != tt.line {
				t.Errorf("expected line %d, got %d", tt.line, errors[0].Line)
			}

			if errors[0].Column != tt.column {
				t.Errorf("expected column %d, got %d", tt.column, errors[0].Column)
			}

			if errors[0].Path != "$.specVersion" {
				t.Errorf("expected path $.specVersion, got %s", errors[0].Path)
			}
		})
	}
}
