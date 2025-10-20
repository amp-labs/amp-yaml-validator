package validator

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/catalog"
	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateWrite validates the write action.
func validateWrite(ctx context.Context, valCtx *ValidationContext, integration openapi.Integration, write *openapi.IntegrationWrite, basePath string) {
	if write == nil {
		return
	}

	// Check that objects list exists and is not empty
	if write.Objects == nil || len(*write.Objects) == 0 {
		valCtx.AddErrorWithSuggestion(
			types.ErrMissingWriteObjects,
			fmt.Sprintf("%s.write.objects", basePath),
			types.RuleWriteObjects,
			"Add at least one object to the write.objects list",
		)

		return
	}

	// Validate each object
	for i, obj := range *write.Objects {
		objectPath := fmt.Sprintf("%s.write.objects[%d]", basePath, i)

		// Check required fields
		if obj.ObjectName == "" {
			valCtx.AddErrorWithSuggestion(
				"Object name is required",
				fmt.Sprintf("%s.objectName", objectPath),
				types.RuleRequiredField,
				"Add an objectName for this write object",
			)
		} else {
			// Validate object name against provider schema
			validateObjectNameForWrite(ctx, valCtx, integration.Provider, integration.Module, obj.ObjectName, fmt.Sprintf("%s.objectName", objectPath))
		}
	}
}

// validateObjectNameForWrite validates that an object name exists in the provider's schema.
func validateObjectNameForWrite(ctx context.Context, valCtx *ValidationContext, provider string, module string, objectName string, path string) {
	// Skip validation if provider is not set
	if provider == "" {
		return
	}

	// Try to get object list from catalog
	objects, err := valCtx.CatalogProvider.ListObjects(ctx, provider, module)

	// If catalog doesn't support object enumeration, add a warning
	if err != nil && errors.Is(err, catalog.ErrNotSupported) {
		valCtx.AddWarningWithSuggestion(
			"Object name validation skipped (catalog does not support object enumeration)",
			path,
			types.RuleCatalogAccess,
			"Consider manually verifying that this object is supported by the provider",
		)

		return
	}

	// If we got an error other than ErrNotSupported, add a warning
	if err != nil {
		valCtx.AddWarningWithSuggestion(
			fmt.Sprintf("Failed to retrieve object list from catalog: %s", err.Error()),
			path,
			types.RuleCatalogAccess,
			"Consider manually verifying that this object is supported by the provider",
		)

		return
	}

	// Check if object is in the list
	found := false

	for _, obj := range objects {
		if obj == objectName {
			found = true
			break
		}
	}

	if !found {
		providerDesc := provider
		if module != "" {
			providerDesc = fmt.Sprintf("%s (module: %s)", provider, module)
		}

		valCtx.AddErrorWithSuggestion(
			fmt.Sprintf("Object '%s' is not supported by provider %s", objectName, providerDesc),
			path,
			types.RuleObjectExists,
			fmt.Sprintf("Use one of the supported objects for this provider"),
		)
	}
}
