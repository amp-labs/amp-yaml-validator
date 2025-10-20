package validator

import (
	"context"
	"fmt"
	"strings"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
	"github.com/amp-labs/connectors/common"
)

// validateProviderSpecific performs provider-specific validation.
func validateProviderSpecific(ctx context.Context, valCtx *ValidationContext) {
	if valCtx.Manifest.Integrations == nil {
		return
	}

	for i, integration := range valCtx.Manifest.Integrations {
		basePath := fmt.Sprintf("$.integrations[%d]", i)
		validateProviderForIntegration(ctx, valCtx, integration, basePath)
	}
}

// validateProviderForIntegration validates provider-specific rules for a single integration.
func validateProviderForIntegration(
	ctx context.Context,
	valCtx *ValidationContext,
	integration openapi.Integration,
	basePath string,
) {
	// Skip if catalog not accessible
	if !valCtx.HasCatalogAccess(ctx) {
		return
	}

	// Validate provider capabilities
	validateProviderCapabilities(ctx, valCtx, integration, basePath)

	// Validate provider-specific rules
	validateSalesforceRules(valCtx, integration, basePath)

	// Validate module support if module specified
	if integration.Module != "" {
		validateModuleSupport(ctx, valCtx, integration, basePath)
	}
}

// validateProviderCapabilities validates that the provider supports the requested actions.
func validateProviderCapabilities(
	ctx context.Context,
	valCtx *ValidationContext,
	integration openapi.Integration,
	basePath string,
) {
	providerInfo, err := valCtx.GetProviderInfo(ctx, integration.Provider)
	if err != nil {
		if strings.Contains(err.Error(), "failed to load catalog") {
			valCtx.AddWarningWithSuggestion(
				fmt.Sprintf("Catalog unavailable: %v", err),
				basePath+".provider",
				types.RuleCatalogAccess,
				"Ensure provider catalog is accessible",
			)
		} else {
			valCtx.AddError(
				fmt.Sprintf("Provider '%s' not found", integration.Provider),
				basePath+".provider",
				types.RuleProviderNotSupported,
			)
		}

		return
	}

	// Get module-specific support if module is specified
	support := providerInfo.Support

	if integration.Module != "" {
		moduleInfo := providerInfo.ReadModuleInfo(common.ModuleID(integration.Module))
		if moduleInfo != nil {
			support = moduleInfo.Support
		}
	}

	// Check read capability
	if integration.Read != nil && !support.Read {
		valCtx.AddError(
			fmt.Sprintf("Provider '%s' does not support read actions", integration.Provider),
			basePath+".read",
			types.RuleProviderCapabilityRead,
		)
	}

	// Check write capability
	if integration.Write != nil && !support.Write {
		valCtx.AddError(
			fmt.Sprintf("Provider '%s' does not support write actions", integration.Provider),
			basePath+".write",
			types.RuleProviderCapabilityWrite,
		)
	}

	// Check subscribe capability
	if integration.Subscribe != nil && !support.Subscribe {
		valCtx.AddError(
			fmt.Sprintf("Provider '%s' does not support subscribe actions", integration.Provider),
			basePath+".subscribe",
			types.RuleProviderCapabilitySubscribe,
		)
	}

	// Check proxy capability
	if integration.Proxy != nil && integration.Proxy.Enabled != nil && *integration.Proxy.Enabled && !support.Proxy {
		valCtx.AddError(
			fmt.Sprintf("Provider '%s' does not support proxy", integration.Provider),
			basePath+".proxy",
			types.RuleProviderCapabilityProxy,
		)
	}
	// REMOVED: Bulk write check until manifest signals bulk usage
	// If manifest has an explicit bulk flag in the future, check that flag against
	// the relevant support.BulkWrite.Insert/Update/Upsert/Delete fields
} //nolint:wsl // Commented code block

// validateSalesforceRules validates Salesforce-specific constraints.
func validateSalesforceRules(ctx *ValidationContext, integration openapi.Integration, basePath string) {
	if integration.Provider != "salesforce" {
		return
	}

	// Check subscribe object limit
	if integration.Subscribe != nil && integration.Subscribe.Objects != nil {
		objectCount := len(*integration.Subscribe.Objects)
		if objectCount > types.MaxSalesforceSubscribeObjects {
			ctx.AddErrorWithSuggestion(
				fmt.Sprintf("Salesforce integrations cannot have more than %d subscribe objects (found %d)",
					types.MaxSalesforceSubscribeObjects, objectCount),
				basePath+".subscribe.objects",
				types.RuleSalesforceSubscribeLimit,
				fmt.Sprintf("Reduce the number of subscribe objects to %d or fewer", types.MaxSalesforceSubscribeObjects),
			)
		}
	}
}

// validateModuleSupport validates that the specified module exists and supports the requested actions.
func validateModuleSupport(
	ctx context.Context,
	valCtx *ValidationContext,
	integration openapi.Integration,
	basePath string,
) {
	if integration.Module == "" {
		return
	}

	moduleInfo, err := valCtx.CatalogProvider.GetModuleInfo(ctx, integration.Provider, integration.Module)
	if err != nil {
		valCtx.AddErrorWithSuggestion(
			fmt.Sprintf("Provider '%s' does not support module '%s'", integration.Provider, integration.Module),
			basePath+".module",
			types.RuleProviderModule,
			"Check the provider catalog for available modules",
		)

		return
	}

	// Validate module-specific capabilities
	support := moduleInfo.Support

	// Check read capability for module
	if integration.Read != nil && !support.Read {
		valCtx.AddError(
			fmt.Sprintf("Module '%s' for provider '%s' does not support read actions", integration.Module, integration.Provider),
			basePath+".read",
			types.RuleProviderCapabilityRead,
		)
	}

	// Check write capability for module
	if integration.Write != nil && !support.Write {
		valCtx.AddError(
			fmt.Sprintf(
				"Module '%s' for provider '%s' does not support write actions",
				integration.Module, integration.Provider,
			),
			basePath+".write",
			types.RuleProviderCapabilityWrite,
		)
	}

	// Check subscribe capability for module
	if integration.Subscribe != nil && !support.Subscribe {
		valCtx.AddError(
			fmt.Sprintf(
				"Module '%s' for provider '%s' does not support subscribe actions",
				integration.Module, integration.Provider,
			),
			basePath+".subscribe",
			types.RuleProviderCapabilitySubscribe,
		)
	}

	// Check proxy capability for module
	if integration.Proxy != nil && integration.Proxy.Enabled != nil && *integration.Proxy.Enabled && !support.Proxy {
		valCtx.AddError(
			fmt.Sprintf("Module '%s' for provider '%s' does not support proxy", integration.Module, integration.Provider),
			basePath+".proxy",
			types.RuleProviderCapabilityProxy,
		)
	}
}
