package validator

import (
	"testing"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/parser"
	"github.com/amp-labs/amp-yaml-validator/types"
	"github.com/stretchr/testify/require"
)

func TestValidateSnowflakeRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		manifest     *openapi.Manifest
		wantErrors   int
		expectedRule string
	}{
		{
			name: "valid - fullHistory backfill",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "snowflake",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{
									ObjectName:  "table1",
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
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "valid - no backfill",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "snowflake",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{
									ObjectName:  "table1",
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
			name: "invalid - days backfill (7 days)",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "snowflake",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{
									ObjectName:  "table1",
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
			wantErrors:   1,
			expectedRule: types.RuleSnowflakeBackfill,
		},
		{
			name: "invalid - days backfill (30 days)",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "snowflake",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{
									ObjectName:  "table1",
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
			expectedRule: types.RuleSnowflakeBackfill,
		},
		{
			name: "valid - multiple objects with fullHistory",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "snowflake",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{
									ObjectName:  "table1",
									Destination: "webhook",
									Schedule:    "0 */12 * * *",
									Backfill: &openapi.Backfill{
										DefaultPeriod: openapi.DefaultPeriod{
											FullHistory: boolPtr(true),
										},
									},
								},
								{
									ObjectName:  "table2",
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
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "invalid - multiple objects with days backfill",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "snowflake",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{
									ObjectName:  "table1",
									Destination: "webhook",
									Schedule:    "0 */12 * * *",
									Backfill: &openapi.Backfill{
										DefaultPeriod: openapi.DefaultPeriod{
											Days: intPtr(7),
										},
									},
								},
								{
									ObjectName:  "table2",
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
			wantErrors:   2, // Two errors, one per object
			expectedRule: types.RuleSnowflakeBackfill,
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
									ObjectName:  "Account",
									Destination: "webhook",
									Schedule:    "0 */12 * * *",
									Backfill: &openapi.Backfill{
										DefaultPeriod: openapi.DefaultPeriod{
											Days: intPtr(30), // Other providers can use days
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

			validateSnowflakeRules(valCtx)

			errors := valCtx.GetErrors()
			require.Len(t, errors, tt.wantErrors, "unexpected number of errors")

			if tt.expectedRule != "" && len(errors) > 0 {
				for _, err := range errors {
					require.Equal(t, tt.expectedRule, err.Rule, "unexpected rule type")
				}
			}
		})
	}
}

func TestValidateSnowflakeBackfill(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		obj          openapi.IntegrationObject
		wantErrors   int
		expectedRule string
	}{
		{
			name: "valid - fullHistory",
			obj: openapi.IntegrationObject{
				ObjectName: "table1",
				Backfill: &openapi.Backfill{
					DefaultPeriod: openapi.DefaultPeriod{
						FullHistory: boolPtr(true),
					},
				},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "valid - no backfill",
			obj: openapi.IntegrationObject{
				ObjectName: "table1",
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "invalid - days backfill",
			obj: openapi.IntegrationObject{
				ObjectName: "table1",
				Backfill: &openapi.Backfill{
					DefaultPeriod: openapi.DefaultPeriod{
						Days: intPtr(7),
					},
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleSnowflakeBackfill,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			posMap := parser.NewPositionMap()
			dirMap := parser.NewDirectiveMap()
			manifest := &openapi.Manifest{SpecVersion: types.CurrentSpecVersion}
			valCtx := NewValidationContext(manifest, posMap, dirMap, nil, nil, nil, nil)

			validateSnowflakeBackfill(valCtx, tt.obj, "$.integrations[0]", 0)

			errors := valCtx.GetErrors()
			require.Len(t, errors, tt.wantErrors, "unexpected number of errors")

			if tt.expectedRule != "" && len(errors) > 0 {
				require.Equal(t, tt.expectedRule, errors[0].Rule, "unexpected rule type")
			}
		})
	}
}
