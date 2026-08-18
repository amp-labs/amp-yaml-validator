package validator

import (
	"testing"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/parser"
	"github.com/amp-labs/amp-yaml-validator/types"
	"github.com/stretchr/testify/require"
)

func TestValidateSubscribeEventTypes(t *testing.T) {
	t.Parallel()

	createEnabled := openapi.CreateEventEnabledAlways
	updateEnabled := openapi.Always
	deleteEnabled := openapi.DeleteEventEnabledAlways
	assocEnabled := openapi.AssociationChangeEventEnabledAlways

	tests := []struct {
		name         string
		obj          openapi.IntegrationSubscribeObject
		wantErrors   int
		wantWarnings int
		expectedRule string
	}{
		{
			name: "valid - createEvent enabled",
			obj: openapi.IntegrationSubscribeObject{
				ObjectName: "Account",
				CreateEvent: &openapi.CreateEvent{
					Enabled: &createEnabled,
				},
				InheritFieldsAndMapping: true,
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "valid - updateEvent enabled",
			obj: openapi.IntegrationSubscribeObject{
				ObjectName: "Account",
				UpdateEvent: &openapi.UpdateEvent{
					Enabled: &updateEnabled,
				},
				InheritFieldsAndMapping: true,
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "valid - deleteEvent enabled",
			obj: openapi.IntegrationSubscribeObject{
				ObjectName: "Account",
				DeleteEvent: &openapi.DeleteEvent{
					Enabled: &deleteEnabled,
				},
				InheritFieldsAndMapping: true,
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "valid - associationChangeEvent enabled",
			obj: openapi.IntegrationSubscribeObject{
				ObjectName: "Account",
				AssociationChangeEvent: &openapi.AssociationChangeEvent{
					Enabled: &assocEnabled,
				},
				InheritFieldsAndMapping: true,
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "valid - multiple events enabled",
			obj: openapi.IntegrationSubscribeObject{
				ObjectName: "Account",
				CreateEvent: &openapi.CreateEvent{
					Enabled: &createEnabled,
				},
				UpdateEvent: &openapi.UpdateEvent{
					Enabled: &updateEnabled,
				},
				DeleteEvent: &openapi.DeleteEvent{
					Enabled: &deleteEnabled,
				},
				InheritFieldsAndMapping: true,
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "warning - base definition with no events enabled",
			obj: openapi.IntegrationSubscribeObject{
				ObjectName:              "Account",
				InheritFieldsAndMapping: true,
				// No events specified: valid base definition, warns
			},
			wantErrors:   0,
			wantWarnings: 1,
			expectedRule: types.RuleSubscribeMinimumEvents,
		},
		{
			name: "warning - events exist but no enabled field set",
			obj: openapi.IntegrationSubscribeObject{
				ObjectName:  "Account",
				CreateEvent: &openapi.CreateEvent{
					// No Enabled field
				},
				UpdateEvent: &openapi.UpdateEvent{
					// No Enabled field
				},
				InheritFieldsAndMapping: true,
			},
			wantErrors:   0,
			wantWarnings: 1,
			expectedRule: types.RuleSubscribeMinimumEvents,
		},
		{
			name: "valid - otherEvents specified",
			obj: openapi.IntegrationSubscribeObject{
				ObjectName: "Account",
				OtherEvents: &openapi.OtherEvents{
					"customEvent1",
					"customEvent2",
				},
				InheritFieldsAndMapping: true,
			},
			wantErrors:   0,
			expectedRule: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			posMap := parser.NewPositionMap()
			dirMap := parser.NewDirectiveMap()
			manifest := &openapi.Manifest{SpecVersion: types.CurrentSpecVersion}
			valCtx := NewValidationContext(manifest, posMap, dirMap, nil, nil, nil, nil)

			validateSubscribeEventTypes(valCtx, testCase.obj, "$.integrations[0].subscribe.objects[0]")

			errors := valCtx.GetErrors()
			require.Len(t, errors, testCase.wantErrors, "unexpected number of errors")

			warnings := valCtx.GetWarnings()
			require.Len(t, warnings, testCase.wantWarnings, "unexpected number of warnings")

			if testCase.expectedRule != "" && len(errors) > 0 {
				require.Equal(t, testCase.expectedRule, errors[0].Rule, "unexpected rule type")
			}

			if testCase.expectedRule != "" && len(warnings) > 0 {
				require.Equal(t, testCase.expectedRule, warnings[0].Rule, "unexpected rule type")
			}
		})
	}
}

func TestHasAnySubscribeEventEnabled(t *testing.T) {
	t.Parallel()

	createEnabled := openapi.CreateEventEnabledAlways
	updateEnabled := openapi.Always
	deleteEnabled := openapi.DeleteEventEnabledAlways
	assocEnabled := openapi.AssociationChangeEventEnabledAlways

	tests := []struct {
		name     string
		obj      openapi.IntegrationSubscribeObject
		expected bool
	}{
		{
			name: "createEvent enabled",
			obj: openapi.IntegrationSubscribeObject{
				CreateEvent: &openapi.CreateEvent{
					Enabled: &createEnabled,
				},
			},
			expected: true,
		},
		{
			name: "updateEvent enabled",
			obj: openapi.IntegrationSubscribeObject{
				UpdateEvent: &openapi.UpdateEvent{
					Enabled: &updateEnabled,
				},
			},
			expected: true,
		},
		{
			name: "deleteEvent enabled",
			obj: openapi.IntegrationSubscribeObject{
				DeleteEvent: &openapi.DeleteEvent{
					Enabled: &deleteEnabled,
				},
			},
			expected: true,
		},
		{
			name: "associationChangeEvent enabled",
			obj: openapi.IntegrationSubscribeObject{
				AssociationChangeEvent: &openapi.AssociationChangeEvent{
					Enabled: &assocEnabled,
				},
			},
			expected: true,
		},
		{
			name: "otherEvents specified",
			obj: openapi.IntegrationSubscribeObject{
				OtherEvents: &openapi.OtherEvents{
					"customEvent1",
				},
			},
			expected: true,
		},
		{
			name: "no events enabled",
			obj:  openapi.IntegrationSubscribeObject{
				// No events
			},
			expected: false,
		},
		{
			name: "events exist but no enabled field",
			obj: openapi.IntegrationSubscribeObject{
				CreateEvent: &openapi.CreateEvent{
					// No Enabled field
				},
				UpdateEvent: &openapi.UpdateEvent{
					// No Enabled field
				},
			},
			expected: false,
		},
		{
			name: "multiple events enabled",
			obj: openapi.IntegrationSubscribeObject{
				CreateEvent: &openapi.CreateEvent{
					Enabled: &createEnabled,
				},
				UpdateEvent: &openapi.UpdateEvent{
					Enabled: &updateEnabled,
				},
				DeleteEvent: &openapi.DeleteEvent{
					Enabled: &deleteEnabled,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := hasAnySubscribeEventEnabled(tt.obj)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateSubscribeEventEnabledFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		obj          openapi.IntegrationSubscribeObject
		wantErrors   int
		expectedRule string
	}{
		{
			name: "valid - always enabled",
			obj: openapi.IntegrationSubscribeObject{
				ObjectName: "Account",
				CreateEvent: &openapi.CreateEvent{
					Enabled: func() *openapi.CreateEventEnabled {
						val := openapi.CreateEventEnabledAlways

						return &val
					}(),
				},
				InheritFieldsAndMapping: true,
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "invalid - wrong enabled value for createEvent",
			obj: openapi.IntegrationSubscribeObject{
				ObjectName: "Account",
				CreateEvent: &openapi.CreateEvent{
					Enabled: func() *openapi.CreateEventEnabled {
						val := openapi.CreateEventEnabled("never")

						return &val
					}(),
				},
				InheritFieldsAndMapping: true,
			},
			wantErrors:   1,
			expectedRule: types.RuleUpdateEventEnabled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			posMap := parser.NewPositionMap()
			dirMap := parser.NewDirectiveMap()
			manifest := &openapi.Manifest{SpecVersion: types.CurrentSpecVersion}
			valCtx := NewValidationContext(manifest, posMap, dirMap, nil, nil, nil, nil)

			validateSubscribeEventTypes(valCtx, tt.obj, "$.integrations[0].subscribe.objects[0]")

			errors := valCtx.GetErrors()
			require.Len(t, errors, tt.wantErrors, "unexpected number of errors")

			if tt.expectedRule != "" && len(errors) > 0 {
				// Check if any error has the expected rule
				found := false

				for _, err := range errors {
					if err.Rule == tt.expectedRule {
						found = true

						break
					}
				}

				require.True(t, found, "expected rule %s not found", tt.expectedRule)
			}
		})
	}
}
