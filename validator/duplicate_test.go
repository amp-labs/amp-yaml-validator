package validator

import (
	"testing"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/parser"
	"github.com/amp-labs/amp-yaml-validator/types"
	"github.com/stretchr/testify/require"
)

func TestValidateDuplicateObjects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		manifest     *openapi.Manifest
		wantErrors   int
		expectedRule string
	}{
		{
			name: "no duplicates - all valid",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "salesforce",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{ObjectName: "Account"},
								{ObjectName: "Contact"},
								{ObjectName: "Opportunity"},
							},
						},
					},
				},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "duplicate in read objects",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "salesforce",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{ObjectName: "Account"},
								{ObjectName: "Contact"},
								{ObjectName: "Account"}, // Duplicate
							},
						},
					},
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleDuplicateReadObject,
		},
		{
			name: "duplicate in write objects",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "salesforce",
						Write: &openapi.IntegrationWrite{
							Objects: &[]openapi.IntegrationWriteObject{
								{ObjectName: "Account"},
								{ObjectName: "Contact"},
								{ObjectName: "Account"}, // Duplicate
							},
						},
					},
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleDuplicateWriteObject,
		},
		{
			name: "duplicate in subscribe objects",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "salesforce",
						Subscribe: &openapi.IntegrationSubscribe{
							Objects: &[]openapi.IntegrationSubscribeObject{
								{ObjectName: "Account", InheritFieldsAndMapping: true},
								{ObjectName: "Contact", InheritFieldsAndMapping: true},
								{ObjectName: "Account", InheritFieldsAndMapping: true}, // Duplicate
							},
						},
					},
				},
			},
			wantErrors:   1,
			expectedRule: types.RuleDuplicateSubscribeObject,
		},
		{
			name: "multiple duplicates in read objects",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "salesforce",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{ObjectName: "Account"},
								{ObjectName: "Contact"},
								{ObjectName: "Account"}, // Duplicate #1
								{ObjectName: "Contact"}, // Duplicate #2
								{ObjectName: "Opportunity"},
								{ObjectName: "Account"}, // Duplicate #3
							},
						},
					},
				},
			},
			wantErrors:   3, // Three duplicate errors (indices 2, 3, 5)
			expectedRule: types.RuleDuplicateReadObject,
		},
		{
			name: "same object name across different actions is allowed",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "salesforce",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{ObjectName: "Account"},
							},
						},
						Write: &openapi.IntegrationWrite{
							Objects: &[]openapi.IntegrationWriteObject{
								{ObjectName: "Account"}, // Same name but different action - OK
							},
						},
						Subscribe: &openapi.IntegrationSubscribe{
							Objects: &[]openapi.IntegrationSubscribeObject{
								{ObjectName: "Account", InheritFieldsAndMapping: true}, // Same name but different action - OK
							},
						},
					},
				},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "duplicates in multiple integrations",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "salesforce",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{ObjectName: "Account"},
								{ObjectName: "Account"}, // Duplicate in integration 0
							},
						},
					},
					{
						Provider: "hubspot",
						Write: &openapi.IntegrationWrite{
							Objects: &[]openapi.IntegrationWriteObject{
								{ObjectName: "Contact"},
								{ObjectName: "Contact"}, // Duplicate in integration 1
							},
						},
					},
				},
			},
			wantErrors:   2,
			expectedRule: "", // Different rules for different integrations
		},
		{
			name: "duplicates across all action types in single integration",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "salesforce",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{
								{ObjectName: "Account"},
								{ObjectName: "Account"}, // Duplicate
							},
						},
						Write: &openapi.IntegrationWrite{
							Objects: &[]openapi.IntegrationWriteObject{
								{ObjectName: "Contact"},
								{ObjectName: "Contact"}, // Duplicate
							},
						},
						Subscribe: &openapi.IntegrationSubscribe{
							Objects: &[]openapi.IntegrationSubscribeObject{
								{ObjectName: "Opportunity", InheritFieldsAndMapping: true},
								{ObjectName: "Opportunity", InheritFieldsAndMapping: true}, // Duplicate
							},
						},
					},
				},
			},
			wantErrors:   3,  // One error per action type
			expectedRule: "", // Multiple different rules
		},
		{
			name: "no read objects",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "salesforce",
						Read:     nil,
					},
				},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "empty read objects list",
			manifest: &openapi.Manifest{
				SpecVersion: types.CurrentSpecVersion,
				Integrations: []openapi.Integration{
					{
						Provider: "salesforce",
						Read: &openapi.IntegrationRead{
							Objects: &[]openapi.IntegrationObject{},
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

			validateDuplicateObjects(valCtx)

			errors := valCtx.GetErrors()
			require.Len(t, errors, tt.wantErrors, "unexpected number of errors")

			if tt.expectedRule != "" && len(errors) > 0 {
				require.Equal(t, tt.expectedRule, errors[0].Rule, "unexpected rule type")
			}
		})
	}
}

func TestCheckReadDuplicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		objects      []openapi.IntegrationObject
		wantErrors   int
		expectedRule string
	}{
		{
			name: "no duplicates",
			objects: []openapi.IntegrationObject{
				{ObjectName: "Account"},
				{ObjectName: "Contact"},
				{ObjectName: "Opportunity"},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "one duplicate",
			objects: []openapi.IntegrationObject{
				{ObjectName: "Account"},
				{ObjectName: "Contact"},
				{ObjectName: "Account"}, // Duplicate at index 2
			},
			wantErrors:   1,
			expectedRule: types.RuleDuplicateReadObject,
		},
		{
			name: "multiple duplicates of same object",
			objects: []openapi.IntegrationObject{
				{ObjectName: "Account"},
				{ObjectName: "Account"}, // Duplicate at index 1
				{ObjectName: "Account"}, // Duplicate at index 2
			},
			wantErrors:   2, // Two errors (indices 1 and 2)
			expectedRule: types.RuleDuplicateReadObject,
		},
		{
			name: "duplicate of different objects",
			objects: []openapi.IntegrationObject{
				{ObjectName: "Account"},
				{ObjectName: "Contact"},
				{ObjectName: "Account"}, // Duplicate
				{ObjectName: "Contact"}, // Duplicate
			},
			wantErrors:   2,
			expectedRule: types.RuleDuplicateReadObject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			posMap := parser.NewPositionMap()
			dirMap := parser.NewDirectiveMap()
			manifest := &openapi.Manifest{SpecVersion: types.CurrentSpecVersion}
			valCtx := NewValidationContext(manifest, posMap, dirMap, nil, nil, nil, nil)

			checkReadDuplicates(valCtx, tt.objects, "$.integrations[0].read.objects")

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

func TestCheckWriteDuplicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		objects      []openapi.IntegrationWriteObject
		wantErrors   int
		expectedRule string
	}{
		{
			name: "no duplicates",
			objects: []openapi.IntegrationWriteObject{
				{ObjectName: "Account"},
				{ObjectName: "Contact"},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "one duplicate",
			objects: []openapi.IntegrationWriteObject{
				{ObjectName: "Account"},
				{ObjectName: "Account"}, // Duplicate
			},
			wantErrors:   1,
			expectedRule: types.RuleDuplicateWriteObject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			posMap := parser.NewPositionMap()
			dirMap := parser.NewDirectiveMap()
			manifest := &openapi.Manifest{SpecVersion: types.CurrentSpecVersion}
			valCtx := NewValidationContext(manifest, posMap, dirMap, nil, nil, nil, nil)

			checkWriteDuplicates(valCtx, tt.objects, "$.integrations[0].write.objects")

			errors := valCtx.GetErrors()
			require.Len(t, errors, tt.wantErrors, "unexpected number of errors")

			if tt.expectedRule != "" && len(errors) > 0 {
				require.Equal(t, tt.expectedRule, errors[0].Rule, "unexpected rule type")
			}
		})
	}
}

func TestCheckSubscribeDuplicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		objects      []openapi.IntegrationSubscribeObject
		wantErrors   int
		expectedRule string
	}{
		{
			name: "no duplicates",
			objects: []openapi.IntegrationSubscribeObject{
				{ObjectName: "Account", InheritFieldsAndMapping: true},
				{ObjectName: "Contact", InheritFieldsAndMapping: true},
			},
			wantErrors:   0,
			expectedRule: "",
		},
		{
			name: "one duplicate",
			objects: []openapi.IntegrationSubscribeObject{
				{ObjectName: "Account", InheritFieldsAndMapping: true},
				{ObjectName: "Account", InheritFieldsAndMapping: true}, // Duplicate
			},
			wantErrors:   1,
			expectedRule: types.RuleDuplicateSubscribeObject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			posMap := parser.NewPositionMap()
			dirMap := parser.NewDirectiveMap()
			manifest := &openapi.Manifest{SpecVersion: types.CurrentSpecVersion}
			valCtx := NewValidationContext(manifest, posMap, dirMap, nil, nil, nil, nil)

			checkSubscribeDuplicates(valCtx, tt.objects, "$.integrations[0].subscribe.objects")

			errors := valCtx.GetErrors()
			require.Len(t, errors, tt.wantErrors, "unexpected number of errors")

			if tt.expectedRule != "" && len(errors) > 0 {
				require.Equal(t, tt.expectedRule, errors[0].Rule, "unexpected rule type")
			}
		})
	}
}
