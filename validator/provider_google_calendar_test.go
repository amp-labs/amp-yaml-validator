package validator

import (
	"testing"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/parser"
	"github.com/amp-labs/amp-yaml-validator/types"
	"github.com/stretchr/testify/require"
)

func TestValidateGoogleCalendarRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		manifest     *openapi.Manifest
		wantErrors   int
		expectedRule string
	}{
		{
			name: "valid - no backfill",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "googlecalendar",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{
									ObjectName:  "events",
									Destination: "webhook",
									Schedule:    "0 */12 * * *",
								},
							},
						},
					},
				},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "valid - backfill within limit (28 days)",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "googlecalendar",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{
									ObjectName:  "events",
									Destination: "webhook",
									Schedule:    "0 */12 * * *",
									Backfill: &openapi.Backfill{
										DefaultPeriod: openapi.DefaultPeriod{
											Days: intPtr(28),
										},
									},
								},
							},
						},
					},
				},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "valid - backfill less than limit (7 days)",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "googlecalendar",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{
									ObjectName:  "events",
									Destination: "webhook",
									Schedule:    "0 */12 * * *",
									Backfill: &openapi.Backfill{
										DefaultPeriod: openapi.DefaultPeriod{
											Days: intPtr(7),
										},
									},
								},
							},
						},
					},
				},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "invalid - fullHistory not allowed",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "googlecalendar",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{
									ObjectName:  "events",
									Destination: "webhook",
									Schedule:    "0 */12 * * *",
									Backfill: &openapi.Backfill{
										DefaultPeriod: openapi.DefaultPeriod{
											FullHistory: boolPtr(true),
										},
									},
								},
							},
						},
					},
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleGoogleCalendarBackfill,
		},
		{
			name: "invalid - backfill exceeds 28 days",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "googlecalendar",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{
									ObjectName:  "events",
									Destination: "webhook",
									Schedule:    "0 */12 * * *",
									Backfill: &openapi.Backfill{
										DefaultPeriod: openapi.DefaultPeriod{
											Days: intPtr(30),
										},
									},
								},
							},
						},
					},
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleGoogleCalendarBackfill,
		},
		{
			name: "invalid - backfill far exceeds limit (180 days)",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "googlecalendar",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{
									ObjectName:  "events",
									Destination: "webhook",
									Schedule:    "0 */12 * * *",
									Backfill: &openapi.Backfill{
										DefaultPeriod: openapi.DefaultPeriod{
											Days: intPtr(180),
										},
									},
								},
							},
						},
					},
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleGoogleCalendarBackfill,
		},
		{
			name: "valid - other Google Calendar objects not affected",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "googlecalendar",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{
									ObjectName:  "contacts",
									Destination: "webhook",
									Schedule:    "0 */12 * * *",
									Backfill: &openapi.Backfill{
										DefaultPeriod: openapi.DefaultPeriod{
											Days: intPtr(90), // Other objects can have longer backfill
										},
									},
								},
							},
						},
					},
				},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "valid - other providers not affected",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "salesforce",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{
									ObjectName:  "events",
									Destination: "webhook",
									Schedule:    "0 */12 * * *",
									Backfill: &openapi.Backfill{
										DefaultPeriod: openapi.DefaultPeriod{
											FullHistory: boolPtr(true), // Other providers can use fullHistory
										},
									},
								},
							},
						},
					},
				},
			},
			wantErrors:   0,
			expectedRule: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			posMap := parser.NewPositionMap()
			dirMap := parser.NewDirectiveMap()
			valCtx := NewValidationContext(tt.manifest, posMap, dirMap, nil, nil, nil, nil)

			validateGoogleCalendarRules(valCtx)

			errors := valCtx.GetErrors()
			require.Len(t, errors, tt.wantErrors, "unexpected number of errors")

			if tt.expectedRule != "" && len(errors) > 0 {
				require.Equal(t, tt.expectedRule, errors[0].Rule, "unexpected rule type")
			}
		})
	}
}

func TestValidateGoogleCalendarEventsBackfill(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		obj          openapi.IntegrationObject
		wantErrors   int
		expectedRule string
	}{
		{
			name: "valid - 28 days",
			obj: openapi.IntegrationObject{
				ObjectName: "events",
				Backfill: &openapi.Backfill{
					DefaultPeriod: openapi.DefaultPeriod{
						Days: intPtr(28),
					},
				},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "valid - no backfill",
			obj: openapi.IntegrationObject{
				ObjectName: "events",
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "invalid - fullHistory",
			obj: openapi.IntegrationObject{
				ObjectName: "events",
				Backfill: &openapi.Backfill{
					DefaultPeriod: openapi.DefaultPeriod{
						FullHistory: boolPtr(true),
					},
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleGoogleCalendarBackfill,
		},
		{
			name: "invalid - 29 days (exceeds limit)",
			obj: openapi.IntegrationObject{
				ObjectName: "events",
				Backfill: &openapi.Backfill{
					DefaultPeriod: openapi.DefaultPeriod{
						Days: intPtr(29),
					},
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleGoogleCalendarBackfill,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			posMap := parser.NewPositionMap()
			dirMap := parser.NewDirectiveMap()
			manifest := &openapi.Manifest{SpecVersion: types.CurrentSpecVersion}
			valCtx := NewValidationContext(manifest, posMap, dirMap, nil, nil, nil, nil)

			validateGoogleCalendarEventsBackfill(valCtx, tt.obj, "$.integrations[0]", 0)

			errors := valCtx.GetErrors()
			require.Len(t, errors, tt.wantErrors, "unexpected number of errors")

			if tt.expectedRule != "" && len(errors) > 0 {
				require.Equal(t, tt.expectedRule, errors[0].Rule, "unexpected rule type")
			}
		})
	}
}

// Helper functions for creating pointers.
func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
