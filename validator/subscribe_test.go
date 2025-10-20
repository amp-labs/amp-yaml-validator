package validator

import (
	"testing"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/parser"
	"github.com/amp-labs/amp-yaml-validator/types"
)

func TestValidateSubscribe(t *testing.T) {
	t.Parallel()

	enabled := openapi.Always
	watchFieldsAutoAll := openapi.UpdateEventWatchFieldsAutoAll
	watchFieldsAutoSelected := openapi.UpdateEventWatchFieldsAutoSelected
	watchFields := []string{"field1", "field2"}

	tests := []struct {
		name         string
		integration  openapi.Integration
		wantErrors   int
		expectedRule string
	}{
		{
			name: "valid subscribe with read",
			integration: openapi.Integration{
				Provider: "salesforce",
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "Account"},
					},
				},
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: &[]openapi.IntegrationSubscribeObject{
						{
							ObjectName:              "Account",
							Destination:             "webhook",
							InheritFieldsAndMapping: true,
						},
					},
				},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "subscribe without read",
			integration: openapi.Integration{
				Provider: "salesforce",
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: &[]openapi.IntegrationSubscribeObject{
						{
							ObjectName:              "Account",
							Destination:             "webhook",
							InheritFieldsAndMapping: true,
						},
					},
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleSubscribeRequiresRead,
		},
		{
			name: "subscribe with empty objects list",
			integration: openapi.Integration{
				Provider: "salesforce",
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "Account"},
					},
				},
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: &[]openapi.IntegrationSubscribeObject{},
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleSubscribeObjects,
		},
		{
			name: "subscribe with nil objects",
			integration: openapi.Integration{
				Provider: "salesforce",
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "Account"},
					},
				},
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: nil,
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleSubscribeObjects,
		},
		{
			name: "subscribe with inheritFieldsAndMapping false",
			integration: openapi.Integration{
				Provider: "salesforce",
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "Account"},
					},
				},
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: &[]openapi.IntegrationSubscribeObject{
						{
							ObjectName:              "Account",
							Destination:             "webhook",
							InheritFieldsAndMapping: false,
						},
					},
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleSubscribeInheritFields,
		},
		{
			name: "subscribe with valid updateEvent - watchFieldsAuto all",
			integration: openapi.Integration{
				Provider: "salesforce",
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "Account"},
					},
				},
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: &[]openapi.IntegrationSubscribeObject{
						{
							ObjectName:              "Account",
							Destination:             "webhook",
							InheritFieldsAndMapping: true,
							UpdateEvent: &openapi.UpdateEvent{
								Enabled:         &enabled,
								WatchFieldsAuto: &watchFieldsAutoAll,
							},
						},
					},
				},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "subscribe with valid updateEvent - watchFieldsAuto selected",
			integration: openapi.Integration{
				Provider: "salesforce",
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "Account"},
					},
				},
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: &[]openapi.IntegrationSubscribeObject{
						{
							ObjectName:              "Account",
							Destination:             "webhook",
							InheritFieldsAndMapping: true,
							UpdateEvent: &openapi.UpdateEvent{
								Enabled:         &enabled,
								WatchFieldsAuto: &watchFieldsAutoSelected,
							},
						},
					},
				},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "subscribe with valid updateEvent - requiredWatchFields",
			integration: openapi.Integration{
				Provider: "salesforce",
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "Account"},
					},
				},
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: &[]openapi.IntegrationSubscribeObject{
						{
							ObjectName:              "Account",
							Destination:             "webhook",
							InheritFieldsAndMapping: true,
							UpdateEvent: &openapi.UpdateEvent{
								Enabled:             &enabled,
								RequiredWatchFields: &watchFields,
							},
						},
					},
				},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "subscribe with updateEvent - no watch fields configuration",
			integration: openapi.Integration{
				Provider: "salesforce",
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "Account"},
					},
				},
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: &[]openapi.IntegrationSubscribeObject{
						{
							ObjectName:              "Account",
							Destination:             "webhook",
							InheritFieldsAndMapping: true,
							UpdateEvent: &openapi.UpdateEvent{
								Enabled: &enabled,
							},
						},
					},
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleUpdateEventWatchFields,
		},
		{
			name: "subscribe with updateEvent - both watch fields configurations",
			integration: openapi.Integration{
				Provider: "salesforce",
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "Account"},
					},
				},
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: &[]openapi.IntegrationSubscribeObject{
						{
							ObjectName:              "Account",
							Destination:             "webhook",
							InheritFieldsAndMapping: true,
							UpdateEvent: &openapi.UpdateEvent{
								Enabled:             &enabled,
								RequiredWatchFields: &watchFields,
								WatchFieldsAuto:     &watchFieldsAutoAll,
							},
						},
					},
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleUpdateEventWatchFields,
		},
		{
			name: "subscribe with updateEvent - empty requiredWatchFields",
			integration: openapi.Integration{
				Provider: "salesforce",
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "Account"},
					},
				},
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: &[]openapi.IntegrationSubscribeObject{
						{
							ObjectName:              "Account",
							Destination:             "webhook",
							InheritFieldsAndMapping: true,
							UpdateEvent: &openapi.UpdateEvent{
								Enabled:             &enabled,
								RequiredWatchFields: &[]string{},
							},
						},
					},
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleUpdateEventWatchFields,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create validation context
			posMap := parser.NewPositionMap()
			path := "$.integrations[0]"

			// Set up positions
			posMap.Set(path+".subscribe", parser.NewPosition(10, 3))
			posMap.Set(path+".subscribe.objects", parser.NewPosition(11, 5))
			posMap.Set(path+".subscribe.objects[0]", parser.NewPosition(12, 7))
			posMap.Set(path+".subscribe.objects[0].inheritFieldsAndMapping", parser.NewPosition(13, 9))
			posMap.Set(path+".subscribe.objects[0].updateEvent", parser.NewPosition(14, 9))

			ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)

			// Validate
			validateSubscribe(t.Context(), ctx, tt.integration, path)

			// Check errors
			errors := ctx.GetErrors()
			if tt.wantErrors > 0 {
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

func TestValidateSubscribeLineNumbers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		integration  openapi.Integration
		path         string
		setupPosMap  func(pm parser.PositionMap, path string)
		expectedLine int
		expectedPath string
	}{
		{
			name: "subscribe without read error at line 10",
			integration: openapi.Integration{
				Provider: "salesforce",
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: &[]openapi.IntegrationSubscribeObject{
						{
							ObjectName:              "Account",
							Destination:             "webhook",
							InheritFieldsAndMapping: true,
						},
					},
				},
			},
			path: "$.integrations[0]",
			setupPosMap: func(pm parser.PositionMap, path string) {
				pm.Set(path+".subscribe", parser.NewPosition(10, 3))
			},
			expectedLine: 10,
			expectedPath: "$.integrations[0].subscribe",
		},
		{
			name: "inheritFieldsAndMapping false error at line 25",
			integration: openapi.Integration{
				Provider: "salesforce",
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "Account"},
					},
				},
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: &[]openapi.IntegrationSubscribeObject{
						{
							ObjectName:              "Account",
							Destination:             "webhook",
							InheritFieldsAndMapping: false,
						},
					},
				},
			},
			path: "$.integrations[1]",
			setupPosMap: func(pm parser.PositionMap, path string) {
				pm.Set(path+".subscribe.objects[0].inheritFieldsAndMapping", parser.NewPosition(25, 7))
			},
			expectedLine: 25,
			expectedPath: "$.integrations[1].subscribe.objects[0].inheritFieldsAndMapping",
		},
		{
			name: "watch fields error at line 30",
			integration: openapi.Integration{
				Provider: "salesforce",
				Read: &openapi.IntegrationRead{
					Objects: &[]openapi.IntegrationObject{
						{ObjectName: "Account"},
					},
				},
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: &[]openapi.IntegrationSubscribeObject{
						{
							ObjectName:              "Account",
							Destination:             "webhook",
							InheritFieldsAndMapping: true,
							UpdateEvent: &openapi.UpdateEvent{
								Enabled: func() *openapi.UpdateEventEnabled { e := openapi.Always; return &e }(),
							},
						},
					},
				},
			},
			path: "$.integrations[2]",
			setupPosMap: func(pm parser.PositionMap, path string) {
				pm.Set(path+".subscribe.objects[0].updateEvent", parser.NewPosition(30, 9))
			},
			expectedLine: 30,
			expectedPath: "$.integrations[2].subscribe.objects[0].updateEvent",
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
			validateSubscribe(t.Context(), ctx, tt.integration, tt.path)

			// Check that error has correct line number and path
			errors := ctx.GetErrors()
			if len(errors) == 0 {
				t.Fatal("expected error for invalid subscribe configuration")
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

func TestValidateSubscribeMultipleObjects(t *testing.T) {
	t.Parallel()

	// Test multiple subscribe objects with different configurations
	objects := []openapi.IntegrationSubscribeObject{
		{
			ObjectName:              "Account",
			Destination:             "webhook",
			InheritFieldsAndMapping: true, // valid
		},
		{
			ObjectName:              "Contact",
			Destination:             "webhook",
			InheritFieldsAndMapping: false, // invalid - must be true
		},
		{
			ObjectName:              "Lead",
			Destination:             "webhook",
			InheritFieldsAndMapping: true,
			UpdateEvent: &openapi.UpdateEvent{
				Enabled:         func() *openapi.UpdateEventEnabled { e := openapi.Always; return &e }(),
				WatchFieldsAuto: func() *openapi.UpdateEventWatchFieldsAuto { w := openapi.UpdateEventWatchFieldsAutoAll; return &w }(),
			}, // valid
		},
		{
			ObjectName:              "Opportunity",
			Destination:             "webhook",
			InheritFieldsAndMapping: true,
			UpdateEvent: &openapi.UpdateEvent{
				Enabled: func() *openapi.UpdateEventEnabled { e := openapi.Always; return &e }(),
				// missing watch fields configuration - invalid
			},
		},
	}

	integration := openapi.Integration{
		Provider: "salesforce",
		Read: &openapi.IntegrationRead{
			Objects: &[]openapi.IntegrationObject{
				{ObjectName: "Account"},
			},
		},
		Subscribe: &openapi.IntegrationSubscribe{
			Objects: &objects,
		},
	}

	posMap := parser.NewPositionMap()
	path := "$.integrations[0]"

	// Set positions for each object
	for i := range objects {
		posMap.Set(path+".subscribe.objects["+string(rune('0'+i))+"]", parser.NewPosition(10+i*10, 5))
		posMap.Set(path+".subscribe.objects["+string(rune('0'+i))+"].inheritFieldsAndMapping", parser.NewPosition(10+i*10+1, 7))
		posMap.Set(path+".subscribe.objects["+string(rune('0'+i))+"].updateEvent", parser.NewPosition(10+i*10+2, 7))
	}

	ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)

	// Validate
	validateSubscribe(t.Context(), ctx, integration, path)

	// Check errors - should have 2 errors (indices 1 and 3)
	errors := ctx.GetErrors()
	expectedErrors := 2

	if len(errors) != expectedErrors {
		t.Errorf("expected %d errors, got %d", expectedErrors, len(errors))

		for i, err := range errors {
			t.Logf("Error %d: %s (line %d, rule: %s)", i, err.Message, err.Line, err.Rule)
		}
	}

	// Verify errors contain expected rules
	foundRules := make(map[string]int)
	for _, err := range errors {
		foundRules[err.Rule]++
	}

	if foundRules[types.RuleSubscribeInheritFields] != 1 {
		t.Errorf("expected 1 RuleSubscribeInheritFields error, got %d", foundRules[types.RuleSubscribeInheritFields])
	}

	if foundRules[types.RuleUpdateEventWatchFields] != 1 {
		t.Errorf("expected 1 RuleUpdateEventWatchFields error, got %d", foundRules[types.RuleUpdateEventWatchFields])
	}
}

func TestValidateUpdateEvent(t *testing.T) {
	t.Parallel()

	enabled := openapi.Always
	watchFieldsAutoAll := openapi.UpdateEventWatchFieldsAutoAll
	watchFieldsAutoSelected := openapi.UpdateEventWatchFieldsAutoSelected
	watchFields := []string{"field1"}
	emptyWatchFields := []string{}

	tests := []struct {
		name         string
		updateEvent  *openapi.UpdateEvent
		wantError    bool
		expectedRule string
	}{
		{
			name: "valid - watchFieldsAuto all",
			updateEvent: &openapi.UpdateEvent{
				Enabled:         &enabled,
				WatchFieldsAuto: &watchFieldsAutoAll,
			},
			wantError:    false,
			expectedRule: "",
		},
		{
			name: "valid - watchFieldsAuto selected",
			updateEvent: &openapi.UpdateEvent{
				Enabled:         &enabled,
				WatchFieldsAuto: &watchFieldsAutoSelected,
			},
			wantError:    false,
			expectedRule: "",
		},
		{
			name: "valid - requiredWatchFields with fields",
			updateEvent: &openapi.UpdateEvent{
				Enabled:             &enabled,
				RequiredWatchFields: &watchFields,
			},
			wantError:    false,
			expectedRule: "",
		},
		{
			name: "invalid - no watch fields configuration",
			updateEvent: &openapi.UpdateEvent{
				Enabled: &enabled,
			},
			wantError:    true,
			expectedRule: types.RuleUpdateEventWatchFields,
		},
		{
			name: "invalid - empty requiredWatchFields",
			updateEvent: &openapi.UpdateEvent{
				Enabled:             &enabled,
				RequiredWatchFields: &emptyWatchFields,
			},
			wantError:    true,
			expectedRule: types.RuleUpdateEventWatchFields,
		},
		{
			name: "invalid - both configurations",
			updateEvent: &openapi.UpdateEvent{
				Enabled:             &enabled,
				RequiredWatchFields: &watchFields,
				WatchFieldsAuto:     &watchFieldsAutoAll,
			},
			wantError:    true,
			expectedRule: types.RuleUpdateEventWatchFields,
		},
		{
			name: "valid - enabled not set, but watchFieldsAuto set",
			updateEvent: &openapi.UpdateEvent{
				WatchFieldsAuto: &watchFieldsAutoAll,
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
			path := "$.integrations[0].subscribe.objects[0].updateEvent"
			posMap.Set(path, parser.NewPosition(10, 5))
			posMap.Set(path+".enabled", parser.NewPosition(11, 7))

			ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)

			// Validate
			validateUpdateEvent(ctx, tt.updateEvent, path)

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

func TestValidateSubscribeRequiredFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		object       openapi.IntegrationSubscribeObject
		wantErrors   int
		expectedRule string
	}{
		{
			name: "valid object with all required fields",
			object: openapi.IntegrationSubscribeObject{
				ObjectName:              "Account",
				Destination:             "webhook",
				InheritFieldsAndMapping: true,
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "missing objectName",
			object: openapi.IntegrationSubscribeObject{
				ObjectName:              "",
				Destination:             "webhook",
				InheritFieldsAndMapping: true,
			},
			wantErrors:   1,
			expectedRule: types.RuleRequiredField,
		},
		{
			name: "missing destination",
			object: openapi.IntegrationSubscribeObject{
				ObjectName:              "Account",
				Destination:             "",
				InheritFieldsAndMapping: true,
			},
			wantErrors:   1,
			expectedRule: types.RuleRequiredField,
		},
		{
			name: "missing both objectName and destination",
			object: openapi.IntegrationSubscribeObject{
				ObjectName:              "",
				Destination:             "",
				InheritFieldsAndMapping: true,
			},
			wantErrors:   2,
			expectedRule: types.RuleRequiredField,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create validation context
			posMap := parser.NewPositionMap()
			path := "$.integrations[0]"
			posMap.Set(path+".subscribe.objects[0]", parser.NewPosition(10, 5))
			posMap.Set(path+".subscribe.objects[0].objectName", parser.NewPosition(11, 7))
			posMap.Set(path+".subscribe.objects[0].destination", parser.NewPosition(12, 7))

			ctx := NewValidationContext(nil, posMap, parser.NewDirectiveMap(), nil, nil, nil, nil)

			// Create a minimal integration for the test
			integration := openapi.Integration{
				Provider: "salesforce",
			}

			// Validate the object directly
			validateSubscribeObject(t.Context(), ctx, integration, tt.object, path, 0)

			// Check errors
			errors := ctx.GetErrors()
			if tt.wantErrors > 0 {
				if len(errors) == 0 {
					t.Error("expected error but got none")
				} else if len(errors) != tt.wantErrors {
					t.Errorf("expected %d errors, got %d", tt.wantErrors, len(errors))
				}
				// Check that all errors match expected rule
				for _, err := range errors {
					if err.Rule != tt.expectedRule {
						t.Errorf("expected rule %s, got %s", tt.expectedRule, err.Rule)
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
