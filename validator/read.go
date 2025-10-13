package validator

import (
	"errors"
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/catalog"
	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateRead validates the read action.
func validateRead(ctx *ValidationContext, integration openapi.Integration, read *openapi.IntegrationRead, basePath string) {
	if read == nil {
		return
	}

	// Check that objects list exists and is not empty
	if read.Objects == nil || len(*read.Objects) == 0 {
		ctx.AddErrorWithSuggestion(
			types.ErrMissingReadObjects,
			fmt.Sprintf("%s.read.objects", basePath),
			types.RuleReadObjects,
			"Add at least one object to the read.objects list",
		)
		return
	}

	// Validate each object
	for i, obj := range *read.Objects {
		validateReadObject(ctx, integration, obj, basePath, i)
	}
}

// validateReadObject validates a single read object.
func validateReadObject(ctx *ValidationContext, integration openapi.Integration, obj openapi.IntegrationObject, basePath string, index int) {
	objectPath := fmt.Sprintf("%s.read.objects[%d]", basePath, index)

	// Check required fields
	if obj.ObjectName == "" {
		ctx.AddErrorWithSuggestion(
			"Object name is required",
			fmt.Sprintf("%s.objectName", objectPath),
			types.RuleRequiredField,
			"Add an objectName for this read object",
		)
	} else {
		// Validate object name against provider schema
		validateObjectName(ctx, integration.Provider, integration.Module, obj.ObjectName, fmt.Sprintf("%s.objectName", objectPath))
	}

	if obj.Destination == "" {
		ctx.AddErrorWithSuggestion(
			"Destination is required",
			fmt.Sprintf("%s.destination", objectPath),
			types.RuleRequiredField,
			"Add a destination for this read object",
		)
	}

	if obj.Schedule == "" {
		ctx.AddErrorWithSuggestion(
			"Schedule is required",
			fmt.Sprintf("%s.schedule", objectPath),
			types.RuleRequiredField,
			"Add a schedule for this read object (e.g., '*/10 * * * *' for every 10 minutes)",
		)
	} else {
		// Validate schedule
		validateSchedule(ctx, obj.Schedule, fmt.Sprintf("%s.schedule", objectPath))
	}

	// Validate delivery mode
	if obj.Delivery != nil {
		validateDeliveryMode(ctx, obj.Delivery, fmt.Sprintf("%s.delivery", objectPath))
	}

	// Validate backfill
	if obj.Backfill != nil {
		validateBackfill(ctx, obj.Backfill, fmt.Sprintf("%s.backfill", objectPath))
	}

	// If enabled is always, validate always-enabled constraints
	if obj.Enabled == openapi.IntegrationObjectEnabledAlways {
		validateAlwaysEnabledObject(ctx, obj, objectPath)
	}
}

// validateObjectName validates that an object name exists in the provider's schema.
func validateObjectName(ctx *ValidationContext, provider string, module string, objectName string, path string) {
	// Skip validation if provider is not set
	if provider == "" {
		return
	}

	// Try to get object list from catalog
	objects, err := ctx.CatalogProvider.ListObjects(provider, module)

	// If catalog doesn't support object enumeration, add a warning
	if err != nil && errors.Is(err, catalog.ErrNotSupported) {
		ctx.AddWarningWithSuggestion(
			"Object name validation skipped (catalog does not support object enumeration)",
			path,
			types.RuleCatalogAccess,
			"Consider manually verifying that this object is supported by the provider",
		)
		return
	}

	// If we got an error other than ErrNotSupported, add a warning
	if err != nil {
		ctx.AddWarningWithSuggestion(
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
		ctx.AddErrorWithSuggestion(
			fmt.Sprintf("Object '%s' is not supported by provider %s", objectName, providerDesc),
			path,
			types.RuleObjectExists,
			fmt.Sprintf("Use one of the supported objects for this provider"),
		)
	}
}
