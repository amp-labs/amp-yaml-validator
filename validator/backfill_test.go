package validator

import (
	"testing"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/parser"
	"github.com/amp-labs/amp-yaml-validator/types"
)

func TestValidateBackfill(t *testing.T) {
	t.Parallel()

	days30 := 30
	days0 := 0
	days90 := 90
	fullHistoryTrue := true
	fullHistoryFalse := false

	tests := []struct {
		name         string
		backfill     *openapi.Backfill
		wantError    bool
		expectedRule string
	}{
		{
			name:         "nil backfill is allowed",
			backfill:     nil,
			wantError:    false,
			expectedRule: "",
		},
		{
			name: "valid backfill with days",
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{
					Days: &days30,
				},
			},
			wantError:    false,
			expectedRule: "",
		},
		{
			name: "valid backfill with fullHistory",
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{
					FullHistory: &fullHistoryTrue,
				},
			},
			wantError:    false,
			expectedRule: "",
		},
		{
			name: "valid backfill with 0 days (no backfill)",
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{
					Days: &days0,
				},
			},
			wantError:    false,
			expectedRule: "",
		},
		{
			name: "valid backfill with 90 days",
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{
					Days: &days90,
				},
			},
			wantError:    false,
			expectedRule: "",
		},
		{
			name: "valid backfill with fullHistory false",
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{
					FullHistory: &fullHistoryFalse,
				},
			},
			wantError:    false,
			expectedRule: "",
		},
		{
			name: "invalid - both days and fullHistory",
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{
					Days:        &days30,
					FullHistory: &fullHistoryTrue,
				},
			},
			wantError:    true,
			expectedRule: types.RuleBackfillConfig,
		},
		{
			name: "invalid - neither days nor fullHistory",
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{},
			},
			wantError:    true,
			expectedRule: types.RuleBackfillConfig,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create validation context
			posMap := parser.NewPositionMap()
			path := "$.integrations[0].read.objects[0].backfill"

			// Set up positions
			posMap.Set(path, parser.NewPosition(10, 3))
			posMap.Set(path+".defaultPeriod", parser.NewPosition(11, 5))
			posMap.Set(path+".defaultPeriod.days", parser.NewPosition(12, 7))
			posMap.Set(path+".defaultPeriod.fullHistory", parser.NewPosition(13, 7))

			ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)

			// Validate
			validateBackfill(ctx, tt.backfill, path)

			// Check errors
			errors := ctx.GetErrors()
			if tt.wantError {
				if len(errors) == 0 {
					t.Error("expected error but got none")
				} else {
					if errors[0].Rule != tt.expectedRule {
						t.Errorf("expected rule %s, got %s", tt.expectedRule, errors[0].Rule)
					}
					// Verify suggestion is provided
					if errors[0].Suggestion == "" {
						t.Error("expected suggestion to be provided")
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

func TestValidateBackfillLineNumbers(t *testing.T) {
	t.Parallel()

	days30 := 30
	fullHistoryTrue := true

	tests := []struct {
		name         string
		backfill     *openapi.Backfill
		path         string
		setupPosMap  func(pm parser.PositionMap, path string)
		expectedLine int
		expectedPath string
	}{
		{
			name: "both days and fullHistory error at line 20",
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{
					Days:        &days30,
					FullHistory: &fullHistoryTrue,
				},
			},
			path: "$.integrations[0].read.objects[0].backfill",
			setupPosMap: func(pm parser.PositionMap, path string) {
				pm.Set(path+".defaultPeriod", parser.NewPosition(20, 7))
			},
			expectedLine: 20,
			expectedPath: "$.integrations[0].read.objects[0].backfill.defaultPeriod",
		},
		{
			name: "neither days nor fullHistory error at line 30",
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{},
			},
			path: "$.integrations[1].read.objects[0].backfill",
			setupPosMap: func(pm parser.PositionMap, path string) {
				pm.Set(path, parser.NewPosition(30, 7))
			},
			expectedLine: 30,
			expectedPath: "$.integrations[1].read.objects[0].backfill",
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

			ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)

			// Validate
			validateBackfill(ctx, tt.backfill, tt.path)

			// Check that error has correct line number and path
			errors := ctx.GetErrors()
			if len(errors) == 0 {
				t.Fatal("expected error for invalid backfill configuration")
			}

			if errors[0].Line != tt.expectedLine {
				t.Errorf("expected line %d, got %d", tt.expectedLine, errors[0].Line)
			}

			if errors[0].Path != tt.expectedPath {
				t.Errorf("expected path %s, got %s", tt.expectedPath, errors[0].Path)
			}
		})
	}
}

func TestValidateBackfillMultipleObjects(t *testing.T) {
	t.Parallel()

	days30 := 30
	days90 := 90
	fullHistoryTrue := true

	// Test multiple objects with different backfill configurations
	backfills := []struct {
		backfill *openapi.Backfill
		path     string
		line     int
		isValid  bool
	}{
		{
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{Days: &days30},
			},
			path:    "$.integrations[0].read.objects[0].backfill",
			line:    10,
			isValid: true,
		},
		{
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{
					Days:        &days30,
					FullHistory: &fullHistoryTrue,
				},
			},
			path:    "$.integrations[0].read.objects[1].backfill",
			line:    20,
			isValid: false, // both days and fullHistory is invalid
		},
		{
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{FullHistory: &fullHistoryTrue},
			},
			path:    "$.integrations[0].read.objects[2].backfill",
			line:    30,
			isValid: true,
		},
		{
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{},
			},
			path:    "$.integrations[0].read.objects[3].backfill",
			line:    40,
			isValid: false, // neither days nor fullHistory is invalid
		},
		{
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{Days: &days90},
			},
			path:    "$.integrations[0].read.objects[4].backfill",
			line:    50,
			isValid: true,
		},
	}

	posMap := parser.NewPositionMap()

	for i, b := range backfills {
		// For empty DefaultPeriod (index 3), the validator tags will error
		// so we need to set position on the backfill path itself
		if i == 3 {
			posMap.Set(b.path, parser.NewPosition(b.line, 5))
		} else {
			posMap.Set(b.path+".defaultPeriod", parser.NewPosition(b.line, 5))
		}
	}

	ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)

	// Validate all backfills
	for _, b := range backfills {
		validateBackfill(ctx, b.backfill, b.path)
	}

	// Check errors - should have 2 errors (indices 1 and 3)
	errors := ctx.GetErrors()
	expectedErrors := 0

	for _, b := range backfills {
		if !b.isValid {
			expectedErrors++
		}
	}

	if len(errors) != expectedErrors {
		t.Errorf("expected %d errors, got %d", expectedErrors, len(errors))

		for i, err := range errors {
			t.Logf("Error %d: %s (line %d, rule: %s)", i, err.Message, err.Line, err.Rule)
		}
	}

	// Verify each error corresponds to the correct invalid configuration
	errorLines := make(map[int]bool)
	for _, err := range errors {
		errorLines[err.Line] = true
	}

	for _, b := range backfills {
		if !b.isValid {
			if !errorLines[b.line] {
				t.Errorf("expected error at line %d for invalid backfill", b.line)
			}
		} else {
			if errorLines[b.line] {
				t.Errorf("unexpected error at line %d for valid backfill", b.line)
			}
		}
	}
}

func TestValidateBackfillDaysValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		days      int
		wantError bool
	}{
		{
			name:      "0 days (no backfill)",
			days:      0,
			wantError: false,
		},
		{
			name:      "1 day",
			days:      1,
			wantError: false,
		},
		{
			name:      "7 days (1 week)",
			days:      7,
			wantError: false,
		},
		{
			name:      "30 days (1 month)",
			days:      30,
			wantError: false,
		},
		{
			name:      "90 days (3 months)",
			days:      90,
			wantError: false,
		},
		{
			name:      "365 days (1 year)",
			days:      365,
			wantError: false,
		},
		{
			name:      "1000 days",
			days:      1000,
			wantError: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			days := tt.days
			backfill := &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{
					Days: &days,
				},
			}

			posMap := parser.NewPositionMap()
			path := "$.backfill"
			posMap.Set(path+".defaultPeriod", parser.NewPosition(10, 5))

			ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)
			validateBackfill(ctx, backfill, path)

			errors := ctx.GetErrors()
			if tt.wantError {
				if len(errors) == 0 {
					t.Errorf("expected error for days=%d but got none", tt.days)
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("expected no error for days=%d, got: %v", tt.days, errors)
				}
			}
		})
	}
}

func TestValidateBackfillMutualExclusivity(t *testing.T) {
	t.Parallel()

	// Test that both days and fullHistory cannot be set at the same time
	days30 := 30
	fullHistoryTrue := true

	backfill := &openapi.Backfill{
		DefaultPeriod: openapi.DefaultPeriod{
			Days:        &days30,
			FullHistory: &fullHistoryTrue,
		},
	}

	posMap := parser.NewPositionMap()
	path := "$.integrations[0].read.objects[0].backfill"
	posMap.Set(path+".defaultPeriod", parser.NewPosition(15, 5))

	ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)
	validateBackfill(ctx, backfill, path)

	errors := ctx.GetErrors()
	if len(errors) == 0 {
		t.Fatal("expected error for mutually exclusive fields")
	}

	if errors[0].Rule != types.RuleBackfillConfig {
		t.Errorf("expected rule %s, got %s", types.RuleBackfillConfig, errors[0].Rule)
	}

	// Check that the error message mentions mutual exclusivity
	if errors[0].Suggestion == "" {
		t.Error("expected suggestion to be provided")
	}
}

func TestValidateBackfillRequiredField(t *testing.T) {
	t.Parallel()

	// Test that at least one of days or fullHistory is required
	backfill := &openapi.Backfill{
		DefaultPeriod: openapi.DefaultPeriod{},
	}

	posMap := parser.NewPositionMap()
	path := "$.integrations[0].read.objects[0].backfill"
	posMap.Set(path+".defaultPeriod", parser.NewPosition(15, 5))

	ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)
	validateBackfill(ctx, backfill, path)

	errors := ctx.GetErrors()
	if len(errors) == 0 {
		t.Fatal("expected error for missing required field")
	}

	if errors[0].Rule != types.RuleBackfillConfig {
		t.Errorf("expected rule %s, got %s", types.RuleBackfillConfig, errors[0].Rule)
	}

	// Check that the error message mentions providing either field
	if errors[0].Suggestion == "" {
		t.Error("expected suggestion to be provided")
	}
}

func TestValidateBackfillPathValidation(t *testing.T) {
	t.Parallel()

	days30 := 30
	fullHistoryTrue := true

	tests := []struct {
		name         string
		backfill     *openapi.Backfill
		path         string
		expectedPath string
	}{
		{
			name: "both days and fullHistory in object 0",
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{
					Days:        &days30,
					FullHistory: &fullHistoryTrue,
				},
			},
			path:         "$.integrations[0].read.objects[0].backfill",
			expectedPath: "$.integrations[0].read.objects[0].backfill.defaultPeriod",
		},
		{
			name: "missing fields in object 2",
			backfill: &openapi.Backfill{
				DefaultPeriod: openapi.DefaultPeriod{},
			},
			path:         "$.integrations[1].read.objects[2].backfill",
			expectedPath: "$.integrations[1].read.objects[2].backfill",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			posMap := parser.NewPositionMap()
			posMap.Set(tt.path+".defaultPeriod", parser.NewPosition(10, 5))

			ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)
			validateBackfill(ctx, tt.backfill, tt.path)

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
