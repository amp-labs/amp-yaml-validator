package validator

import (
	"testing"

	"github.com/amp-labs/amp-yaml-validator/catalog"
	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/parser"
	"github.com/amp-labs/amp-yaml-validator/types"
	"github.com/amp-labs/connectors/providers"
)

func TestValidateObjectNameRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		provider            string
		module              string
		objectName          string
		supportedObjects    []string
		catalogSupported    bool
		wantErrorCount      int
		wantWarningCount    int
		expectedErrorRule   string
		expectedWarningRule string
	}{
		{
			name:             "valid object name",
			provider:         "salesforce",
			module:           "",
			objectName:       "Account",
			supportedObjects: []string{"Account", "Contact", "Lead"},
			catalogSupported: true,
			wantErrorCount:   0,
			wantWarningCount: 0,
		},
		{
			name:              "invalid object name",
			provider:          "salesforce",
			module:            "",
			objectName:        "InvalidObject",
			supportedObjects:  []string{"Account", "Contact", "Lead"},
			catalogSupported:  true,
			wantErrorCount:    1,
			wantWarningCount:  0,
			expectedErrorRule: types.RuleObjectExists,
		},
		{
			name:                "catalog not supported - warning issued",
			provider:            "salesforce",
			module:              "",
			objectName:          "Account",
			supportedObjects:    nil,
			catalogSupported:    false,
			wantErrorCount:      0,
			wantWarningCount:    1,
			expectedWarningRule: types.RuleCatalogAccess,
		},
		{
			name:             "valid object name with module",
			provider:         "salesforce",
			module:           "sales",
			objectName:       "Opportunity",
			supportedObjects: []string{"Opportunity", "Account"},
			catalogSupported: true,
			wantErrorCount:   0,
			wantWarningCount: 0,
		},
		{
			name:              "invalid object name with module",
			provider:          "salesforce",
			module:            "sales",
			objectName:        "InvalidObject",
			supportedObjects:  []string{"Opportunity", "Account"},
			catalogSupported:  true,
			wantErrorCount:    1,
			wantWarningCount:  0,
			expectedErrorRule: types.RuleObjectExists,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock catalog with objects
			objectsMap := make(map[string][]string)
			if tt.catalogSupported {
				key := tt.provider
				if tt.module != "" {
					key = tt.provider + ":" + tt.module
				}
				objectsMap[key] = tt.supportedObjects
			}

			mockCatalog := catalog.NewMockCatalogProviderWithObjects(
				map[string]providers.ProviderInfo{
					tt.provider: {
						Name: tt.provider,
						Support: providers.Support{
							Read: true,
						},
					},
				},
				objectsMap,
			)

			// Create integration with read object
			objects := []openapi.IntegrationObject{
				{
					ObjectName:  tt.objectName,
					Destination: "dest",
					Schedule:    "*/10 * * * *",
				},
			}

			integration := openapi.Integration{
				Provider: tt.provider,
				Module:   tt.module,
				Read: &openapi.IntegrationRead{
					Objects: &objects,
				},
			}

			// Create validation context
			ctx := NewValidationContext(&openapi.Manifest{}, parser.NewPositionMap(), parser.NewDirectiveMap(), mockCatalog, nil, nil, nil)

			// Validate
			validateRead(ctx, integration, integration.Read, "$.integrations[0]")

			// Check errors
			errors := ctx.GetErrors()
			if len(errors) != tt.wantErrorCount {
				t.Errorf("expected %d errors, got %d", tt.wantErrorCount, len(errors))
				for _, err := range errors {
					t.Logf("  Error: %s (rule: %s)", err.Message, err.Rule)
				}
			}

			// Check warnings
			warnings := ctx.GetWarnings()
			if len(warnings) != tt.wantWarningCount {
				t.Errorf("expected %d warnings, got %d", tt.wantWarningCount, len(warnings))
				for _, warn := range warnings {
					t.Logf("  Warning: %s (rule: %s)", warn.Message, warn.Rule)
				}
			}

			// Check specific error rule if expected
			if tt.expectedErrorRule != "" && len(errors) > 0 {
				found := false
				for _, err := range errors {
					if err.Rule == tt.expectedErrorRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error rule %s not found", tt.expectedErrorRule)
				}
			}

			// Check specific warning rule if expected
			if tt.expectedWarningRule != "" && len(warnings) > 0 {
				found := false
				for _, warn := range warnings {
					if warn.Rule == tt.expectedWarningRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected warning rule %s not found", tt.expectedWarningRule)
				}
			}
		})
	}
}

func TestValidateObjectNameWrite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		provider            string
		module              string
		objectName          string
		supportedObjects    []string
		catalogSupported    bool
		wantErrorCount      int
		wantWarningCount    int
		expectedErrorRule   string
		expectedWarningRule string
	}{
		{
			name:             "valid object name",
			provider:         "salesforce",
			module:           "",
			objectName:       "Account",
			supportedObjects: []string{"Account", "Contact", "Lead"},
			catalogSupported: true,
			wantErrorCount:   0,
			wantWarningCount: 0,
		},
		{
			name:              "invalid object name",
			provider:          "salesforce",
			module:            "",
			objectName:        "InvalidObject",
			supportedObjects:  []string{"Account", "Contact", "Lead"},
			catalogSupported:  true,
			wantErrorCount:    1,
			wantWarningCount:  0,
			expectedErrorRule: types.RuleObjectExists,
		},
		{
			name:                "catalog not supported - warning issued",
			provider:            "salesforce",
			module:              "",
			objectName:          "Account",
			supportedObjects:    nil,
			catalogSupported:    false,
			wantErrorCount:      0,
			wantWarningCount:    1,
			expectedWarningRule: types.RuleCatalogAccess,
		},
		{
			name:             "valid object name with module",
			provider:         "hubspot",
			module:           "crm",
			objectName:       "Contact",
			supportedObjects: []string{"Contact", "Company"},
			catalogSupported: true,
			wantErrorCount:   0,
			wantWarningCount: 0,
		},
		{
			name:              "invalid object name with module",
			provider:          "hubspot",
			module:            "crm",
			objectName:        "InvalidObject",
			supportedObjects:  []string{"Contact", "Company"},
			catalogSupported:  true,
			wantErrorCount:    1,
			wantWarningCount:  0,
			expectedErrorRule: types.RuleObjectExists,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock catalog with objects
			objectsMap := make(map[string][]string)
			if tt.catalogSupported {
				key := tt.provider
				if tt.module != "" {
					key = tt.provider + ":" + tt.module
				}
				objectsMap[key] = tt.supportedObjects
			}

			mockCatalog := catalog.NewMockCatalogProviderWithObjects(
				map[string]providers.ProviderInfo{
					tt.provider: {
						Name: tt.provider,
						Support: providers.Support{
							Write: true,
						},
					},
				},
				objectsMap,
			)

			// Create integration with write object
			objects := []openapi.IntegrationWriteObject{
				{
					ObjectName: tt.objectName,
				},
			}

			integration := openapi.Integration{
				Provider: tt.provider,
				Module:   tt.module,
				Write: &openapi.IntegrationWrite{
					Objects: &objects,
				},
			}

			// Create validation context
			ctx := NewValidationContext(&openapi.Manifest{}, parser.NewPositionMap(), parser.NewDirectiveMap(), mockCatalog, nil, nil, nil)

			// Validate
			validateWrite(ctx, integration, integration.Write, "$.integrations[0]")

			// Check errors
			errors := ctx.GetErrors()
			if len(errors) != tt.wantErrorCount {
				t.Errorf("expected %d errors, got %d", tt.wantErrorCount, len(errors))
				for _, err := range errors {
					t.Logf("  Error: %s (rule: %s)", err.Message, err.Rule)
				}
			}

			// Check warnings
			warnings := ctx.GetWarnings()
			if len(warnings) != tt.wantWarningCount {
				t.Errorf("expected %d warnings, got %d", tt.wantWarningCount, len(warnings))
				for _, warn := range warnings {
					t.Logf("  Warning: %s (rule: %s)", warn.Message, warn.Rule)
				}
			}

			// Check specific error rule if expected
			if tt.expectedErrorRule != "" && len(errors) > 0 {
				found := false
				for _, err := range errors {
					if err.Rule == tt.expectedErrorRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error rule %s not found", tt.expectedErrorRule)
				}
			}

			// Check specific warning rule if expected
			if tt.expectedWarningRule != "" && len(warnings) > 0 {
				found := false
				for _, warn := range warnings {
					if warn.Rule == tt.expectedWarningRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected warning rule %s not found", tt.expectedWarningRule)
				}
			}
		})
	}
}

func TestValidateObjectNameSubscribe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		provider            string
		module              string
		objectName          string
		supportedObjects    []string
		catalogSupported    bool
		wantErrorCount      int
		wantWarningCount    int
		expectedErrorRule   string
		expectedWarningRule string
	}{
		{
			name:             "valid object name",
			provider:         "salesforce",
			module:           "",
			objectName:       "Account",
			supportedObjects: []string{"Account", "Contact", "Lead"},
			catalogSupported: true,
			wantErrorCount:   0,
			wantWarningCount: 0,
		},
		{
			name:              "invalid object name",
			provider:          "salesforce",
			module:            "",
			objectName:        "InvalidObject",
			supportedObjects:  []string{"Account", "Contact", "Lead"},
			catalogSupported:  true,
			wantErrorCount:    1, // Only for object validation
			wantWarningCount:  0,
			expectedErrorRule: types.RuleObjectExists,
		},
		{
			name:                "catalog not supported - warning issued",
			provider:            "salesforce",
			module:              "",
			objectName:          "Account",
			supportedObjects:    nil,
			catalogSupported:    false,
			wantErrorCount:      0,
			wantWarningCount:    1,
			expectedWarningRule: types.RuleCatalogAccess,
		},
		{
			name:             "valid object name with module",
			provider:         "salesforce",
			module:           "marketing",
			objectName:       "Campaign",
			supportedObjects: []string{"Campaign", "Lead"},
			catalogSupported: true,
			wantErrorCount:   0,
			wantWarningCount: 0,
		},
		{
			name:              "invalid object name with module",
			provider:          "salesforce",
			module:            "marketing",
			objectName:        "InvalidObject",
			supportedObjects:  []string{"Campaign", "Lead"},
			catalogSupported:  true,
			wantErrorCount:    1, // Only for object validation
			wantWarningCount:  0,
			expectedErrorRule: types.RuleObjectExists,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock catalog with objects
			objectsMap := make(map[string][]string)
			if tt.catalogSupported {
				key := tt.provider
				if tt.module != "" {
					key = tt.provider + ":" + tt.module
				}
				objectsMap[key] = tt.supportedObjects
			}

			mockCatalog := catalog.NewMockCatalogProviderWithObjects(
				map[string]providers.ProviderInfo{
					tt.provider: {
						Name: tt.provider,
						Support: providers.Support{
							Subscribe: true,
							Read:      true,
						},
					},
				},
				objectsMap,
			)

			// Create integration with subscribe object and read (required)
			readObjects := []openapi.IntegrationObject{
				{
					ObjectName:  tt.objectName,
					Destination: "dest",
					Schedule:    "*/10 * * * *",
				},
			}

			subscribeObjects := []openapi.IntegrationSubscribeObject{
				{
					ObjectName:              tt.objectName,
					Destination:             "dest",
					InheritFieldsAndMapping: true, // Must be true
				},
			}

			integration := openapi.Integration{
				Provider: tt.provider,
				Module:   tt.module,
				Read: &openapi.IntegrationRead{
					Objects: &readObjects,
				},
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: &subscribeObjects,
				},
			}

			// Create validation context
			ctx := NewValidationContext(&openapi.Manifest{}, parser.NewPositionMap(), parser.NewDirectiveMap(), mockCatalog, nil, nil, nil)

			// Validate
			validateSubscribe(ctx, integration, "$.integrations[0]")

			// Check errors
			errors := ctx.GetErrors()
			if len(errors) != tt.wantErrorCount {
				t.Errorf("expected %d errors, got %d", tt.wantErrorCount, len(errors))
				for _, err := range errors {
					t.Logf("  Error: %s (rule: %s)", err.Message, err.Rule)
				}
			}

			// Check warnings
			warnings := ctx.GetWarnings()
			if len(warnings) != tt.wantWarningCount {
				t.Errorf("expected %d warnings, got %d", tt.wantWarningCount, len(warnings))
				for _, warn := range warnings {
					t.Logf("  Warning: %s (rule: %s)", warn.Message, warn.Rule)
				}
			}

			// Check specific error rule if expected
			if tt.expectedErrorRule != "" && len(errors) > 0 {
				found := false
				for _, err := range errors {
					if err.Rule == tt.expectedErrorRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error rule %s not found", tt.expectedErrorRule)
				}
			}

			// Check specific warning rule if expected
			if tt.expectedWarningRule != "" && len(warnings) > 0 {
				found := false
				for _, warn := range warnings {
					if warn.Rule == tt.expectedWarningRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected warning rule %s not found", tt.expectedWarningRule)
				}
			}
		})
	}
}

func TestValidateObjectNameMultipleObjects(t *testing.T) {
	t.Parallel()

	// Test with multiple objects - some valid, some invalid
	mockCatalog := catalog.NewMockCatalogProviderWithObjects(
		map[string]providers.ProviderInfo{
			"salesforce": {
				Name: "salesforce",
				Support: providers.Support{
					Read: true,
				},
			},
		},
		map[string][]string{
			"salesforce": {"Account", "Contact", "Lead"},
		},
	)

	// Create integration with multiple read objects
	objects := []openapi.IntegrationObject{
		{
			ObjectName:  "Account",
			Destination: "dest1",
			Schedule:    "*/10 * * * *",
		},
		{
			ObjectName:  "InvalidObject1",
			Destination: "dest2",
			Schedule:    "*/10 * * * *",
		},
		{
			ObjectName:  "Contact",
			Destination: "dest3",
			Schedule:    "*/10 * * * *",
		},
		{
			ObjectName:  "InvalidObject2",
			Destination: "dest4",
			Schedule:    "*/10 * * * *",
		},
	}

	integration := openapi.Integration{
		Provider: "salesforce",
		Read: &openapi.IntegrationRead{
			Objects: &objects,
		},
	}

	// Create validation context
	ctx := NewValidationContext(&openapi.Manifest{}, parser.NewPositionMap(), parser.NewDirectiveMap(), mockCatalog, nil, nil, nil)

	// Validate
	validateRead(ctx, integration, integration.Read, "$.integrations[0]")

	// Check errors - should have 2 errors (for the 2 invalid objects)
	errors := ctx.GetErrors()
	if len(errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errors))
		for _, err := range errors {
			t.Logf("  Error: %s (rule: %s)", err.Message, err.Rule)
		}
	}

	// Both errors should be RuleObjectExists
	for _, err := range errors {
		if err.Rule != types.RuleObjectExists {
			t.Errorf("expected rule %s, got %s", types.RuleObjectExists, err.Rule)
		}
	}
}

func TestValidateObjectNameEmptyProvider(t *testing.T) {
	t.Parallel()

	// Test that validation is skipped when provider is empty
	mockCatalog := catalog.NewMockCatalogProviderWithObjects(
		map[string]providers.ProviderInfo{},
		map[string][]string{},
	)

	// Create integration with empty provider
	objects := []openapi.IntegrationObject{
		{
			ObjectName:  "Account",
			Destination: "dest",
			Schedule:    "*/10 * * * *",
		},
	}

	integration := openapi.Integration{
		Provider: "", // Empty provider
		Read: &openapi.IntegrationRead{
			Objects: &objects,
		},
	}

	// Create validation context
	ctx := NewValidationContext(&openapi.Manifest{}, parser.NewPositionMap(), parser.NewDirectiveMap(), mockCatalog, nil, nil, nil)

	// Validate
	validateRead(ctx, integration, integration.Read, "$.integrations[0]")

	// Check that there are no object validation errors (only required field errors)
	errors := ctx.GetErrors()
	for _, err := range errors {
		if err.Rule == types.RuleObjectExists {
			t.Errorf("should not validate object name when provider is empty, but got object validation error")
		}
	}
}
