package validator

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/amp-labs/amp-yaml-validator/catalog"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateObjectNameCommon validates that an object name exists in the provider's schema.
// This is a shared helper used by read, write, and subscribe validators.
func validateObjectNameCommon(
	ctx context.Context,
	valCtx *ValidationContext,
	provider string,
	module string,
	objectName string,
	path string,
) {
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
			"Failed to retrieve object list from catalog: "+err.Error(),
			path,
			types.RuleCatalogAccess,
			"Consider manually verifying that this object is supported by the provider",
		)

		return
	}

	// Check if object is in the list
	found := slices.Contains(objects, objectName)

	if !found {
		providerDesc := provider
		if module != "" {
			providerDesc = fmt.Sprintf("%s (module: %s)", provider, module)
		}

		valCtx.AddErrorWithSuggestion(
			fmt.Sprintf("Object '%s' is not supported by provider %s", objectName, providerDesc),
			path,
			types.RuleObjectExists,
			"Use one of the supported objects for this provider",
		)
	}
}
