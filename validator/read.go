package validator

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/catalog"
	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateRead validates the read action.
func validateRead(ctx context.Context, valCtx *ValidationContext, integration openapi.Integration, read *openapi.IntegrationRead, basePath string) {
	if read == nil {
		return
	}

	// Check that objects list exists and is not empty
	if read.Objects == nil || len(*read.Objects) == 0 {
		valCtx.AddErrorWithSuggestion(
			types.ErrMissingReadObjects,
			fmt.Sprintf("%s.read.objects", basePath),
			types.RuleReadObjects,
			"Add at least one object to the read.objects list",
		)

		return
	}

	// Validate each object
	for i, obj := range *read.Objects {
		validateReadObject(ctx, valCtx, integration, obj, basePath, i)
	}
}

// validateReadObject validates a single read object.
func validateReadObject(ctx context.Context, valCtx *ValidationContext, integration openapi.Integration, obj openapi.IntegrationObject, basePath string, index int) {
	objectPath := fmt.Sprintf("%s.read.objects[%d]", basePath, index)

	// Check required fields
	if obj.ObjectName == "" {
		valCtx.AddErrorWithSuggestion(
			"Object name is required",
			fmt.Sprintf("%s.objectName", objectPath),
			types.RuleRequiredField,
			"Add an objectName for this read object",
		)
	} else {
		// Validate object name against provider schema
		validateObjectName(ctx, valCtx, integration.Provider, integration.Module, obj.ObjectName, fmt.Sprintf("%s.objectName", objectPath))
	}

	if obj.Destination == "" {
		valCtx.AddErrorWithSuggestion(
			"Destination is required",
			fmt.Sprintf("%s.destination", objectPath),
			types.RuleRequiredField,
			"Add a destination for this read object",
		)
	}

	if obj.Schedule == "" {
		valCtx.AddErrorWithSuggestion(
			"Schedule is required",
			fmt.Sprintf("%s.schedule", objectPath),
			types.RuleRequiredField,
			"Add a schedule for this read object (e.g., '*/10 * * * *' for every 10 minutes)",
		)
	} else {
		// Validate schedule
		validateSchedule(valCtx, obj.Schedule, fmt.Sprintf("%s.schedule", objectPath))
	}

	// Validate delivery mode
	if obj.Delivery != nil {
		validateDeliveryMode(valCtx, obj.Delivery, fmt.Sprintf("%s.delivery", objectPath))
	}

	// Validate backfill
	if obj.Backfill != nil {
		validateBackfill(valCtx, obj.Backfill, fmt.Sprintf("%s.backfill", objectPath))
	}

	// If enabled is always, validate always-enabled constraints
	if obj.Enabled == openapi.IntegrationObjectEnabledAlways {
		validateAlwaysEnabledObject(valCtx, obj, objectPath)
	}
}

// validateObjectName validates that an object name exists in the provider's schema.
func validateObjectName(ctx context.Context, valCtx *ValidationContext, provider string, module string, objectName string, path string) {
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
