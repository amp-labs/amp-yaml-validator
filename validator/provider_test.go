package validator

import (
	"testing"

	"github.com/amp-labs/amp-yaml-validator/catalog"
	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/parser"
	"github.com/amp-labs/amp-yaml-validator/types"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

func TestValidateProviderCapabilities(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		provider         string
		hasRead          bool
		hasWrite         bool
		hasSubscribe     bool
		hasProxy         bool
		supportRead      bool
		supportWrite     bool
		supportSubscribe bool
		supportProxy     bool
		wantErrors       int
		expectedRules    []string
	}{
		{
			name:             "all capabilities supported",
			provider:         "salesforce",
			hasRead:          true,
			hasWrite:         true,
			hasSubscribe:     true,
			hasProxy:         true,
			supportRead:      true,
			supportWrite:     true,
			supportSubscribe: true,
			supportProxy:     true,
			wantErrors:       0,
			expectedRules:    []string{},
		},
		{
			name:             "read not supported",
			provider:         "test_provider",
			hasRead:          true,
			hasWrite:         false,
			hasSubscribe:     false,
			hasProxy:         false,
			supportRead:      false,
			supportWrite:     false,
			supportSubscribe: false,
			supportProxy:     false,
			wantErrors:       1,
			expectedRules:    []string{types.RuleProviderCapabilityRead},
		},
		{
			name:             "write not supported",
			provider:         "test_provider",
			hasRead:          false,
			hasWrite:         true,
			hasSubscribe:     false,
			hasProxy:         false,
			supportRead:      true,
			supportWrite:     false,
			supportSubscribe: false,
			supportProxy:     false,
			wantErrors:       1,
			expectedRules:    []string{types.RuleProviderCapabilityWrite},
		},
		{
			name:             "subscribe not supported",
			provider:         "test_provider",
			hasRead:          false,
			hasWrite:         false,
			hasSubscribe:     true,
			hasProxy:         false,
			supportRead:      true,
			supportWrite:     true,
			supportSubscribe: false,
			supportProxy:     false,
			wantErrors:       1,
			expectedRules:    []string{types.RuleProviderCapabilitySubscribe},
		},
		{
			name:             "proxy not supported",
			provider:         "test_provider",
			hasRead:          false,
			hasWrite:         false,
			hasSubscribe:     false,
			hasProxy:         true,
			supportRead:      true,
			supportWrite:     true,
			supportSubscribe: true,
			supportProxy:     false,
			wantErrors:       1,
			expectedRules:    []string{types.RuleProviderCapabilityProxy},
		},
		{
			name:             "multiple capabilities not supported",
			provider:         "test_provider",
			hasRead:          true,
			hasWrite:         true,
			hasSubscribe:     true,
			hasProxy:         false,
			supportRead:      false,
			supportWrite:     false,
			supportSubscribe: true,
			supportProxy:     false,
			wantErrors:       2,
			expectedRules:    []string{types.RuleProviderCapabilityRead, types.RuleProviderCapabilityWrite},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock catalog
			mockCatalog := catalog.NewMockCatalogProvider(map[string]providers.ProviderInfo{
				tt.provider: {
					Name: tt.provider,
					Support: providers.Support{
						Read:      tt.supportRead,
						Write:     tt.supportWrite,
						Subscribe: tt.supportSubscribe,
						Proxy:     tt.supportProxy,
						BulkWrite: providers.BulkWriteSupport{},
					},
				},
			})

			// Create integration
			integration := openapi.Integration{
				Provider: tt.provider,
			}

			// Add read if requested
			if tt.hasRead {
				objects := []openapi.IntegrationObject{{ObjectName: "Account"}}
				integration.Read = &openapi.IntegrationRead{
					Objects: &objects,
				}
			}

			// Add write if requested
			if tt.hasWrite {
				objects := []openapi.IntegrationWriteObject{{ObjectName: "Account"}}
				integration.Write = &openapi.IntegrationWrite{
					Objects: &objects,
				}
			}

			// Add subscribe if requested
			if tt.hasSubscribe {
				objects := []openapi.IntegrationSubscribeObject{{ObjectName: "Account"}}
				integration.Subscribe = &openapi.IntegrationSubscribe{
					Objects: &objects,
				}
			}

			// Add proxy if requested
			if tt.hasProxy {
				enabled := true
				integration.Proxy = &openapi.IntegrationProxy{
					Enabled: &enabled,
				}
			}

			// Create validation context
			ctx := NewValidationContext(&openapi.Manifest{}, parser.NewPositionMap(), parser.NewDirectiveMap(), mockCatalog, nil, nil, nil)

			// Validate
			validateProviderCapabilities(ctx, integration, "$.integrations[0]")

			// Check errors
			errors := ctx.GetErrors()
			if len(errors) != tt.wantErrors {
				t.Errorf("expected %d errors, got %d", tt.wantErrors, len(errors))
				for _, err := range errors {
					t.Logf("  Error: %s (rule: %s)", err.Message, err.Rule)
				}
			}

			// Check rules
			foundRules := make(map[string]bool)
			for _, err := range errors {
				foundRules[err.Rule] = true
			}

			for _, expectedRule := range tt.expectedRules {
				if !foundRules[expectedRule] {
					t.Errorf("expected rule %s not found", expectedRule)
				}
			}
		})
	}
}

func TestValidateSalesforceSubscribeLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		provider     string
		objectCount  int
		wantError    bool
		expectedRule string
	}{
		{
			name:         "salesforce - within limit",
			provider:     "salesforce",
			objectCount:  5,
			wantError:    false,
			expectedRule: "",
		},
		{
			name:         "salesforce - exceeds limit",
			provider:     "salesforce",
			objectCount:  6,
			wantError:    true,
			expectedRule: types.RuleSalesforceSubscribeLimit,
		},
		{
			name:         "non-salesforce - no limit",
			provider:     "hubspot",
			objectCount:  10,
			wantError:    false,
			expectedRule: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create integration with subscribe objects
			objects := make([]openapi.IntegrationSubscribeObject, tt.objectCount)
			for i := 0; i < tt.objectCount; i++ {
				objects[i] = openapi.IntegrationSubscribeObject{
					ObjectName: "Account",
				}
			}

			integration := openapi.Integration{
				Provider: tt.provider,
				Subscribe: &openapi.IntegrationSubscribe{
					Objects: &objects,
				},
			}

			// Create validation context
			ctx := NewValidationContext(&openapi.Manifest{}, parser.NewPositionMap(), parser.NewDirectiveMap(), nil, nil, nil, nil) // No catalog needed for this test

			// Validate
			validateSalesforceRules(ctx, integration, "$.integrations[0]")

			// Check errors
			errors := ctx.GetErrors()
			if tt.wantError {
				if len(errors) == 0 {
					t.Error("expected error but got none")
				} else if errors[0].Rule != tt.expectedRule {
					t.Errorf("expected rule %s, got %s", tt.expectedRule, errors[0].Rule)
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("expected no errors, got %d: %v", len(errors), errors)
				}
			}
		})
	}
}

func TestValidateModuleSupport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		provider      string
		module        string
		hasRead       bool
		hasWrite      bool
		moduleExists  bool
		supportRead   bool
		supportWrite  bool
		wantErrors    int
		expectedRules []string
	}{
		{
			name:          "module exists and supports capabilities",
			provider:      "salesforce",
			module:        "sales",
			hasRead:       true,
			hasWrite:      true,
			moduleExists:  true,
			supportRead:   true,
			supportWrite:  true,
			wantErrors:    0,
			expectedRules: []string{},
		},
		{
			name:          "module does not exist",
			provider:      "salesforce",
			module:        "nonexistent",
			hasRead:       true,
			hasWrite:      false,
			moduleExists:  false,
			supportRead:   false,
			supportWrite:  false,
			wantErrors:    1,
			expectedRules: []string{types.RuleProviderModule},
		},
		{
			name:          "module exists but doesn't support read",
			provider:      "salesforce",
			module:        "sales",
			hasRead:       true,
			hasWrite:      false,
			moduleExists:  true,
			supportRead:   false,
			supportWrite:  true,
			wantErrors:    1,
			expectedRules: []string{types.RuleProviderCapabilityRead},
		},
		{
			name:          "module exists but doesn't support write",
			provider:      "salesforce",
			module:        "sales",
			hasRead:       false,
			hasWrite:      true,
			moduleExists:  true,
			supportRead:   true,
			supportWrite:  false,
			wantErrors:    1,
			expectedRules: []string{types.RuleProviderCapabilityWrite},
		},
		{
			name:          "provider has modules but requested module doesn't exist",
			provider:      "salesforce",
			module:        "marketing",
			hasRead:       true,
			hasWrite:      false,
			moduleExists:  false,
			supportRead:   false,
			supportWrite:  false,
			wantErrors:    1,
			expectedRules: []string{types.RuleProviderModule},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock catalog
			providerInfo := providers.ProviderInfo{
				Name: tt.provider,
				Support: providers.Support{
					Read:      true,
					Write:     true,
					Subscribe: false,
					Proxy:     false,
					BulkWrite: providers.BulkWriteSupport{},
				},
			}

			// Add module if it exists
			if tt.moduleExists {
				providerInfo.Modules = &providers.Modules{
					common.ModuleID(tt.module): providers.ModuleInfo{
						DisplayName: tt.module,
						BaseURL:     "https://test.com",
						Support: providers.Support{
							Read:      tt.supportRead,
							Write:     tt.supportWrite,
							Subscribe: false,
							Proxy:     false,
							BulkWrite: providers.BulkWriteSupport{},
						},
					},
				}
			} else if tt.module == "marketing" {
				// For the test where provider has modules but requested module doesn't exist,
				// create a provider with a different module
				providerInfo.Modules = &providers.Modules{
					common.ModuleID("sales"): providers.ModuleInfo{
						DisplayName: "sales",
						BaseURL:     "https://test.com",
						Support: providers.Support{
							Read:      true,
							Write:     true,
							Subscribe: false,
							Proxy:     false,
							BulkWrite: providers.BulkWriteSupport{},
						},
					},
				}
			}

			mockCatalog := catalog.NewMockCatalogProvider(map[string]providers.ProviderInfo{
				tt.provider: providerInfo,
			})

			// Create integration
			integration := openapi.Integration{
				Provider: tt.provider,
				Module:   tt.module,
			}

			// Add read if requested
			if tt.hasRead {
				objects := []openapi.IntegrationObject{{ObjectName: "Account"}}
				integration.Read = &openapi.IntegrationRead{
					Objects: &objects,
				}
			}

			// Add write if requested
			if tt.hasWrite {
				objects := []openapi.IntegrationWriteObject{{ObjectName: "Account"}}
				integration.Write = &openapi.IntegrationWrite{
					Objects: &objects,
				}
			}

			// Create validation context
			ctx := NewValidationContext(&openapi.Manifest{}, parser.NewPositionMap(), parser.NewDirectiveMap(), mockCatalog, nil, nil, nil)

			// Validate
			validateModuleSupport(ctx, integration, "$.integrations[0]")

			// Check errors
			errors := ctx.GetErrors()
			if len(errors) != tt.wantErrors {
				t.Errorf("expected %d errors, got %d", tt.wantErrors, len(errors))
				for _, err := range errors {
					t.Logf("  Error: %s (rule: %s)", err.Message, err.Rule)
				}
			}

			// Check rules
			foundRules := make(map[string]bool)
			for _, err := range errors {
				foundRules[err.Rule] = true
			}

			for _, expectedRule := range tt.expectedRules {
				if !foundRules[expectedRule] {
					t.Errorf("expected rule %s not found", expectedRule)
				}
			}
		})
	}
}

func TestValidateProviderNotFound(t *testing.T) {
	t.Parallel()

	// Create mock catalog without the provider
	mockCatalog := catalog.NewMockCatalogProvider(map[string]providers.ProviderInfo{
		"salesforce": {
			Name: "salesforce",
			Support: providers.Support{
				Read:  true,
				Write: true,
			},
		},
	})

	integration := openapi.Integration{
		Provider: "nonexistent_provider",
		Read: &openapi.IntegrationRead{
			Objects: &[]openapi.IntegrationObject{{ObjectName: "Account"}},
		},
	}

	ctx := NewValidationContext(&openapi.Manifest{}, parser.NewPositionMap(), parser.NewDirectiveMap(), mockCatalog, nil, nil, nil)

	validateProviderCapabilities(ctx, integration, "$.integrations[0]")

	errors := ctx.GetErrors()
	if len(errors) == 0 {
		t.Fatal("expected error for nonexistent provider")
	}

	if errors[0].Rule != types.RuleProviderNotSupported {
		t.Errorf("expected rule %s, got %s", types.RuleProviderNotSupported, errors[0].Rule)
	}
}
