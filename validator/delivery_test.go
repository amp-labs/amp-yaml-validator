package validator

import (
	"testing"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/parser"
	"github.com/amp-labs/amp-yaml-validator/types"
)

func TestValidateDeliveryMode(t *testing.T) {
	t.Parallel()

	auto := openapi.Auto
	onRequest := openapi.OnRequest
	pageSize50 := 50
	pageSize100 := 100
	pageSize500 := 500
	pageSize25 := 25
	pageSize600 := 600

	tests := []struct {
		name         string
		delivery     *openapi.Delivery
		wantError    bool
		expectedRule string
	}{
		{
			name:         "nil delivery (defaults to auto)",
			delivery:     nil,
			wantError:    false,
			expectedRule: "",
		},
		{
			name: "auto mode without page size",
			delivery: &openapi.Delivery{
				Mode: &auto,
			},
			wantError:    false,
			expectedRule: "",
		},
		{
			name: "auto mode with page size (invalid)",
			delivery: &openapi.Delivery{
				Mode:     &auto,
				PageSize: &pageSize100,
			},
			wantError:    true,
			expectedRule: types.RuleDeliveryMode,
		},
		{
			name: "onRequest mode with valid page size (50)",
			delivery: &openapi.Delivery{
				Mode:     &onRequest,
				PageSize: &pageSize50,
			},
			wantError:    false,
			expectedRule: "",
		},
		{
			name: "onRequest mode with valid page size (100)",
			delivery: &openapi.Delivery{
				Mode:     &onRequest,
				PageSize: &pageSize100,
			},
			wantError:    false,
			expectedRule: "",
		},
		{
			name: "onRequest mode with valid page size (500)",
			delivery: &openapi.Delivery{
				Mode:     &onRequest,
				PageSize: &pageSize500,
			},
			wantError:    false,
			expectedRule: "",
		},
		{
			name: "onRequest mode without page size (invalid)",
			delivery: &openapi.Delivery{
				Mode: &onRequest,
			},
			wantError:    true,
			expectedRule: types.RuleDeliveryMode,
		},
		{
			name: "onRequest mode with page size too small (25)",
			delivery: &openapi.Delivery{
				Mode:     &onRequest,
				PageSize: &pageSize25,
			},
			wantError:    true,
			expectedRule: types.RuleDeliveryMode,
		},
		{
			name: "onRequest mode with page size too large (600)",
			delivery: &openapi.Delivery{
				Mode:     &onRequest,
				PageSize: &pageSize600,
			},
			wantError:    true,
			expectedRule: types.RuleDeliveryMode,
		},
		{
			name:         "no mode specified (defaults to auto)",
			delivery:     &openapi.Delivery{},
			wantError:    false,
			expectedRule: "",
		},
		{
			name: "no mode with page size (allowed - defaults to auto)",
			delivery: &openapi.Delivery{
				PageSize: &pageSize100,
			},
			wantError:    false,
			expectedRule: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create validation context
			posMap := parser.NewPositionMap()
			path := "$.integrations[0].read.objects[0].delivery"

			// Set up positions
			posMap.Set(path, parser.NewPosition(10, 3))
			posMap.Set(path+".mode", parser.NewPosition(11, 5))
			posMap.Set(path+".pageSize", parser.NewPosition(12, 5))

			ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)

			// Validate
			validateDeliveryMode(ctx, tt.delivery, path)

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

func TestValidateDeliveryModeLineNumbers(t *testing.T) {
	t.Parallel()

	auto := openapi.Auto
	onRequest := openapi.OnRequest
	pageSize100 := 100
	pageSize25 := 25

	tests := []struct {
		name         string
		delivery     *openapi.Delivery
		path         string
		setupPosMap  func(pm parser.PositionMap, path string)
		expectedLine int
		expectedPath string
	}{
		{
			name: "auto mode with page size error at line 15",
			delivery: &openapi.Delivery{
				Mode:     &auto,
				PageSize: &pageSize100,
			},
			path: "$.integrations[0].read.objects[0].delivery",
			setupPosMap: func(pm parser.PositionMap, path string) {
				pm.Set(path+".pageSize", parser.NewPosition(15, 7))
			},
			expectedLine: 15,
			expectedPath: "$.integrations[0].read.objects[0].delivery.pageSize",
		},
		{
			name: "onRequest without page size error at line 25",
			delivery: &openapi.Delivery{
				Mode: &onRequest,
			},
			path: "$.integrations[1].read.objects[0].delivery",
			setupPosMap: func(pm parser.PositionMap, path string) {
				pm.Set(path+".pageSize", parser.NewPosition(25, 7))
			},
			expectedLine: 25,
			expectedPath: "$.integrations[1].read.objects[0].delivery.pageSize",
		},
		{
			name: "onRequest with invalid page size at line 35",
			delivery: &openapi.Delivery{
				Mode:     &onRequest,
				PageSize: &pageSize25,
			},
			path: "$.integrations[2].read.objects[0].delivery",
			setupPosMap: func(pm parser.PositionMap, path string) {
				pm.Set(path+".pageSize", parser.NewPosition(35, 7))
			},
			expectedLine: 35,
			expectedPath: "$.integrations[2].read.objects[0].delivery.pageSize",
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
			validateDeliveryMode(ctx, tt.delivery, tt.path)

			// Check that error has correct line number and path
			errors := ctx.GetErrors()
			if len(errors) == 0 {
				t.Fatal("expected error for invalid delivery configuration")
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

func TestValidateDeliveryModePageSizeBoundaries(t *testing.T) {
	t.Parallel()

	onRequest := openapi.OnRequest

	tests := []struct {
		name      string
		pageSize  int
		wantError bool
	}{
		{
			name:      "page size 49 (below minimum)",
			pageSize:  49,
			wantError: true,
		},
		{
			name:      "page size 50 (minimum valid)",
			pageSize:  50,
			wantError: false,
		},
		{
			name:      "page size 51 (above minimum)",
			pageSize:  51,
			wantError: false,
		},
		{
			name:      "page size 250 (middle range)",
			pageSize:  250,
			wantError: false,
		},
		{
			name:      "page size 499 (below maximum)",
			pageSize:  499,
			wantError: false,
		},
		{
			name:      "page size 500 (maximum valid)",
			pageSize:  500,
			wantError: false,
		},
		{
			name:      "page size 501 (above maximum)",
			pageSize:  501,
			wantError: true,
		},
		{
			name:      "page size 1000 (way above maximum)",
			pageSize:  1000,
			wantError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pageSize := tt.pageSize
			delivery := &openapi.Delivery{
				Mode:     &onRequest,
				PageSize: &pageSize,
			}

			posMap := parser.NewPositionMap()
			path := "$.delivery"
			posMap.Set(path+".pageSize", parser.NewPosition(10, 5))

			ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)
			validateDeliveryMode(ctx, delivery, path)

			errors := ctx.GetErrors()
			if tt.wantError {
				if len(errors) == 0 {
					t.Errorf("expected error for page size %d but got none", tt.pageSize)
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("expected no error for page size %d, got: %v", tt.pageSize, errors)
				}
			}
		})
	}
}

func TestValidateDeliveryModeMultipleObjects(t *testing.T) {
	t.Parallel()

	auto := openapi.Auto
	onRequest := openapi.OnRequest
	pageSize100 := 100
	pageSize50 := 50

	// Test multiple objects with different delivery configurations
	deliveries := []struct {
		delivery *openapi.Delivery
		path     string
		line     int
		isValid  bool
	}{
		{
			delivery: &openapi.Delivery{Mode: &auto},
			path:     "$.integrations[0].read.objects[0].delivery",
			line:     10,
			isValid:  true,
		},
		{
			delivery: &openapi.Delivery{Mode: &auto, PageSize: &pageSize100},
			path:     "$.integrations[0].read.objects[1].delivery",
			line:     20,
			isValid:  false, // auto with page size is invalid
		},
		{
			delivery: &openapi.Delivery{Mode: &onRequest, PageSize: &pageSize50},
			path:     "$.integrations[0].read.objects[2].delivery",
			line:     30,
			isValid:  true,
		},
		{
			delivery: &openapi.Delivery{Mode: &onRequest},
			path:     "$.integrations[0].read.objects[3].delivery",
			line:     40,
			isValid:  false, // onRequest without page size is invalid
		},
	}

	posMap := parser.NewPositionMap()
	for _, d := range deliveries {
		posMap.Set(d.path+".pageSize", parser.NewPosition(d.line, 5))
	}

	ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)

	// Validate all deliveries
	for _, d := range deliveries {
		validateDeliveryMode(ctx, d.delivery, d.path)
	}

	// Check errors - should have 2 errors (indices 1 and 3)
	errors := ctx.GetErrors()
	expectedErrors := 0

	for _, d := range deliveries {
		if !d.isValid {
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

	for _, d := range deliveries {
		if !d.isValid {
			if !errorLines[d.line] {
				t.Errorf("expected error at line %d for invalid delivery", d.line)
			}
		} else {
			if errorLines[d.line] {
				t.Errorf("unexpected error at line %d for valid delivery", d.line)
			}
		}
	}
}

func TestValidateDeliveryModeSuggestions(t *testing.T) {
	t.Parallel()

	auto := openapi.Auto
	onRequest := openapi.OnRequest
	pageSize100 := 100
	pageSize25 := 25

	tests := []struct {
		name               string
		delivery           *openapi.Delivery
		wantSuggestion     bool
		suggestionContains string
	}{
		{
			name: "auto with page size suggests removal",
			delivery: &openapi.Delivery{
				Mode:     &auto,
				PageSize: &pageSize100,
			},
			wantSuggestion:     true,
			suggestionContains: "Remove pageSize",
		},
		{
			name: "onRequest without page size suggests adding",
			delivery: &openapi.Delivery{
				Mode: &onRequest,
			},
			wantSuggestion:     true,
			suggestionContains: "Add pageSize",
		},
		{
			name: "onRequest with invalid page size suggests range",
			delivery: &openapi.Delivery{
				Mode:     &onRequest,
				PageSize: &pageSize25,
			},
			wantSuggestion:     true,
			suggestionContains: "between",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			posMap := parser.NewPositionMap()
			path := "$.delivery"
			posMap.Set(path+".pageSize", parser.NewPosition(10, 5))

			ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)
			validateDeliveryMode(ctx, tt.delivery, path)

			errors := ctx.GetErrors()
			if len(errors) == 0 {
				t.Fatal("expected error but got none")
			}

			if tt.wantSuggestion {
				if errors[0].Suggestion == "" {
					t.Error("expected suggestion but got empty string")
				}
			}
		})
	}
}
