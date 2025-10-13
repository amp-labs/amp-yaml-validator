package validator

import (
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateIntegrations validates that integrations list is present and non-empty,
// then validates each integration.
func validateIntegrations(ctx *ValidationContext) {
	// Check that integrations list is present and non-empty
	if len(ctx.Manifest.Integrations) == 0 {
		ctx.AddError(
			"Integrations list is required and must contain at least one integration",
			"$.integrations",
			types.RuleIntegrationStructure,
		)
		return
	}

	// Validate each integration
	for i, integration := range ctx.Manifest.Integrations {
		basePath := fmt.Sprintf("$.integrations[%d]", i)
		validateIntegration(ctx, integration, i)

		// Validate read action
		if integration.Read != nil {
			validateRead(ctx, integration, integration.Read, basePath)
		}

		// Validate write action
		if integration.Write != nil {
			validateWrite(ctx, integration, integration.Write, basePath)
		}

		// Validate subscribe action
		if integration.Subscribe != nil {
			validateSubscribe(ctx, integration, basePath)
		}

		// Validate field mappings
		if integration.Read != nil {
			validateFieldMappings(ctx, integration.Read, basePath)
		}
	}
}

// validateIntegration validates the structure of a single integration.
func validateIntegration(ctx *ValidationContext, integration openapi.Integration, index int) {
	basePath := fmt.Sprintf("$.integrations[%d]", index)

	// Check required fields
	if integration.Name == "" {
		ctx.AddErrorWithSuggestion(
			"Integration name is required",
			fmt.Sprintf("%s.name", basePath),
			types.RuleRequiredField,
			"Add a name for this integration",
		)
	}

	if integration.Provider == "" {
		ctx.AddErrorWithSuggestion(
			"Integration provider is required",
			fmt.Sprintf("%s.provider", basePath),
			types.RuleRequiredField,
			"Add a provider for this integration (e.g., 'salesforce', 'hubspot')",
		)
	}

	// Check that at least one action is defined
	hasAction := integration.Read != nil ||
		integration.Write != nil ||
		integration.Subscribe != nil ||
		integration.Proxy != nil

	if !hasAction {
		ctx.AddErrorWithSuggestion(
			"Integration must have at least one of read, write, subscribe, or proxy",
			basePath,
			types.RuleIntegrationStructure,
			"Add a read, write, subscribe, or proxy section to this integration",
		)
	}
}
