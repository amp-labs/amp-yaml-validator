package validator

import (
	"fmt"
	"testing"

	"github.com/amp-labs/amp-yaml-validator/parser"
	"github.com/amp-labs/amp-yaml-validator/types"
)

func TestValidateSchedule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		schedule     string
		wantError    bool
		expectedRule string
	}{
		{
			name:         "valid 15 minute schedule",
			schedule:     "*/15 * * * *",
			wantError:    false,
			expectedRule: "",
		},
		{
			name:         "valid 10 minute schedule (minimum)",
			schedule:     "*/10 * * * *",
			wantError:    false,
			expectedRule: "",
		},
		{
			name:         "valid hourly schedule",
			schedule:     "0 * * * *",
			wantError:    false,
			expectedRule: "",
		},
		{
			name:         "valid daily schedule at midnight",
			schedule:     "0 0 * * *",
			wantError:    false,
			expectedRule: "",
		},
		{
			name:         "valid daily schedule at noon",
			schedule:     "0 12 * * *",
			wantError:    false,
			expectedRule: "",
		},
		{
			name:         "valid 30 minute schedule",
			schedule:     "*/30 * * * *",
			wantError:    false,
			expectedRule: "",
		},
		{
			name:         "valid specific times (0,30 minutes)",
			schedule:     "0,30 * * * *",
			wantError:    false,
			expectedRule: "",
		},
		{
			name:         "invalid cron syntax - missing fields",
			schedule:     "*/15 * *",
			wantError:    true,
			expectedRule: types.RuleScheduleSyntax,
		},
		{
			name:         "invalid cron syntax - too many fields",
			schedule:     "*/15 * * * * *",
			wantError:    true,
			expectedRule: types.RuleScheduleSyntax,
		},
		{
			name:         "invalid cron syntax - empty string",
			schedule:     "",
			wantError:    true,
			expectedRule: types.RuleScheduleSyntax,
		},
		{
			name:         "invalid cron syntax - random text",
			schedule:     "every 5 minutes",
			wantError:    true,
			expectedRule: types.RuleScheduleSyntax,
		},
		{
			name:         "too frequent - 5 minute interval",
			schedule:     "*/5 * * * *",
			wantError:    true,
			expectedRule: types.RuleScheduleMinimumInterval,
		},
		{
			name:         "too frequent - 1 minute interval",
			schedule:     "*/1 * * * *",
			wantError:    true,
			expectedRule: types.RuleScheduleMinimumInterval,
		},
		{
			name:         "too frequent - every minute",
			schedule:     "* * * * *",
			wantError:    true,
			expectedRule: types.RuleScheduleMinimumInterval,
		},
		{
			name:         "too frequent - 9 minute interval",
			schedule:     "*/9 * * * *",
			wantError:    true,
			expectedRule: types.RuleScheduleMinimumInterval,
		},
		{
			name:         "invalid minute value",
			schedule:     "99 * * * *",
			wantError:    true,
			expectedRule: types.RuleScheduleSyntax,
		},
		{
			name:         "invalid hour value",
			schedule:     "0 25 * * *",
			wantError:    true,
			expectedRule: types.RuleScheduleSyntax,
		},
		{
			name:         "invalid day value",
			schedule:     "0 0 32 * *",
			wantError:    true,
			expectedRule: types.RuleScheduleSyntax,
		},
		{
			name:         "invalid month value",
			schedule:     "0 0 1 13 *",
			wantError:    true,
			expectedRule: types.RuleScheduleSyntax,
		},
		{
			name:         "invalid weekday value",
			schedule:     "0 0 * * 8",
			wantError:    true,
			expectedRule: types.RuleScheduleSyntax,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create validation context
			posMap := parser.NewPositionMap()
			path := "$.integrations[0].read.objects[0].schedule"
			posMap.Set(path, parser.NewPosition(10, 5))

			ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)

			// Validate schedule
			validateSchedule(ctx, tt.schedule, path)

			// Check errors
			errors := ctx.GetErrors()
			if tt.wantError {
				if len(errors) == 0 {
					t.Error("expected error but got none")
				} else {
					if errors[0].Rule != tt.expectedRule {
						t.Errorf("expected rule %s, got %s", tt.expectedRule, errors[0].Rule)
					}
					// Verify path is set correctly
					if errors[0].Path != path {
						t.Errorf("expected path %s, got %s", path, errors[0].Path)
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

func TestValidateScheduleLineNumbers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		schedule string
		path     string
		line     int
		column   int
	}{
		{
			name:     "invalid schedule at line 5",
			schedule: "*/5 * * * *",
			path:     "$.integrations[0].read.objects[0].schedule",
			line:     5,
			column:   3,
		},
		{
			name:     "invalid schedule at line 20",
			schedule: "* * * * *",
			path:     "$.integrations[0].read.objects[1].schedule",
			line:     20,
			column:   7,
		},
		{
			name:     "invalid syntax at line 15",
			schedule: "invalid",
			path:     "$.integrations[1].read.objects[0].schedule",
			line:     15,
			column:   10,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create position map with specific line/column
			posMap := parser.NewPositionMap()
			posMap.Set(tt.path, parser.NewPosition(tt.line, tt.column))

			// Create validation context
			ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)

			// Validate
			validateSchedule(ctx, tt.schedule, tt.path)

			// Check that error has correct line number
			errors := ctx.GetErrors()
			if len(errors) == 0 {
				t.Fatal("expected error for invalid schedule")
			}

			if errors[0].Line != tt.line {
				t.Errorf("expected line %d, got %d", tt.line, errors[0].Line)
			}

			if errors[0].Column != tt.column {
				t.Errorf("expected column %d, got %d", tt.column, errors[0].Column)
			}

			if errors[0].Path != tt.path {
				t.Errorf("expected path %s, got %s", tt.path, errors[0].Path)
			}
		})
	}
}

func TestValidateScheduleMultipleObjects(t *testing.T) {
	t.Parallel()

	// Test multiple objects with different schedules
	schedules := []struct {
		schedule string
		path     string
		line     int
		isValid  bool
	}{
		{schedule: "*/15 * * * *", path: "$.integrations[0].read.objects[0].schedule", line: 10, isValid: true},
		{schedule: "*/5 * * * *", path: "$.integrations[0].read.objects[1].schedule", line: 20, isValid: false},
		{schedule: "0 * * * *", path: "$.integrations[0].read.objects[2].schedule", line: 30, isValid: true},
		{schedule: "invalid", path: "$.integrations[0].read.objects[3].schedule", line: 40, isValid: false},
	}

	posMap := parser.NewPositionMap()
	for _, s := range schedules {
		posMap.Set(s.path, parser.NewPosition(s.line, 5))
	}

	ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)

	// Validate all schedules
	for _, s := range schedules {
		validateSchedule(ctx, s.schedule, s.path)
	}

	// Check errors - should have 2 errors (indices 1 and 3)
	errors := ctx.GetErrors()
	expectedErrors := 0

	for _, s := range schedules {
		if !s.isValid {
			expectedErrors++
		}
	}

	if len(errors) != expectedErrors {
		t.Errorf("expected %d errors, got %d", expectedErrors, len(errors))

		for i, err := range errors {
			t.Logf("Error %d: %s (line %d, rule: %s)", i, err.Message, err.Line, err.Rule)
		}
	}

	// Verify each error corresponds to the correct invalid schedule
	errorLines := make(map[int]bool)
	for _, err := range errors {
		errorLines[err.Line] = true
	}

	for _, s := range schedules {
		if !s.isValid {
			if !errorLines[s.line] {
				t.Errorf("expected error at line %d for schedule %s", s.line, s.schedule)
			}
		} else {
			if errorLines[s.line] {
				t.Errorf("unexpected error at line %d for valid schedule %s", s.line, s.schedule)
			}
		}
	}
}

func TestValidateScheduleRuleTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		schedule     string
		expectedRule string
	}{
		{
			name:         "syntax error produces RuleScheduleSyntax",
			schedule:     "invalid cron",
			expectedRule: types.RuleScheduleSyntax,
		},
		{
			name:         "too frequent produces RuleScheduleMinimumInterval",
			schedule:     "*/5 * * * *",
			expectedRule: types.RuleScheduleMinimumInterval,
		},
		{
			name:         "every minute produces RuleScheduleMinimumInterval",
			schedule:     "* * * * *",
			expectedRule: types.RuleScheduleMinimumInterval,
		},
		{
			name:         "missing fields produces RuleScheduleSyntax",
			schedule:     "*/15 * *",
			expectedRule: types.RuleScheduleSyntax,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			posMap := parser.NewPositionMap()
			path := "$.schedule"
			posMap.Set(path, parser.NewPosition(1, 1))

			ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)
			validateSchedule(ctx, tt.schedule, path)

			errors := ctx.GetErrors()
			if len(errors) == 0 {
				t.Fatal("expected error but got none")
			}

			if errors[0].Rule != tt.expectedRule {
				t.Errorf("expected rule %s, got %s", tt.expectedRule, errors[0].Rule)
			}

			// Verify that a suggestion is provided
			if errors[0].Suggestion == "" {
				t.Error("expected suggestion to be provided")
			}
		})
	}
}

func TestValidateScheduleEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		schedule     string
		wantError    bool
		expectedRule string
		description  string
	}{
		{
			name:         "exactly 10 minutes is valid",
			schedule:     "*/10 * * * *",
			wantError:    false,
			expectedRule: "",
			description:  "boundary case: minimum valid interval",
		},
		{
			name:         "exactly 9 minutes is invalid",
			schedule:     "*/9 * * * *",
			wantError:    true,
			expectedRule: types.RuleScheduleMinimumInterval,
			description:  "boundary case: just below minimum",
		},
		{
			name:         "weekly schedule is valid",
			schedule:     "0 0 * * 0",
			wantError:    false,
			expectedRule: "",
			description:  "weekly schedule (every Sunday)",
		},
		{
			name:         "monthly schedule is valid",
			schedule:     "0 0 1 * *",
			wantError:    false,
			expectedRule: "",
			description:  "monthly schedule (1st of month)",
		},
		{
			name:         "business hours only",
			schedule:     "0 9-17 * * 1-5",
			wantError:    false,
			expectedRule: "",
			description:  "hourly during business hours on weekdays",
		},
		{
			name:         "multiple specific times",
			schedule:     fmt.Sprintf("%d,%d,%d * * * *", 0, 15, 30),
			wantError:    false,
			expectedRule: "",
			description:  "specific minutes: 0, 15, 30",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			posMap := parser.NewPositionMap()
			path := "$.schedule"
			posMap.Set(path, parser.NewPosition(1, 1))

			ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)
			validateSchedule(ctx, tt.schedule, path)

			errors := ctx.GetErrors()
			if tt.wantError {
				if len(errors) == 0 {
					t.Errorf("%s: expected error but got none", tt.description)
				} else if errors[0].Rule != tt.expectedRule {
					t.Errorf("%s: expected rule %s, got %s", tt.description, tt.expectedRule, errors[0].Rule)
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("%s: expected no errors, got %d: %v", tt.description, len(errors), errors)
				}
			}
		})
	}
}
