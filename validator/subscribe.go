package validator

import (
	"errors"
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/catalog"
	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateSubscribe validates the subscribe action.
func validateSubscribe(ctx *ValidationContext, integration openapi.Integration, basePath string) {
	if integration.Subscribe == nil {
		return
	}

	// Check that read is also defined (subscribe requires read)
	if integration.Read == nil {
		ctx.AddErrorWithSuggestion(
			types.ErrSubscribeRequiresRead,
			fmt.Sprintf("%s.subscribe", basePath),
			types.RuleSubscribeRequiresRead,
			"Add a read section to this integration",
		)
	}

	// Check that objects list exists and is not empty
	if integration.Subscribe.Objects == nil || len(*integration.Subscribe.Objects) == 0 {
		ctx.AddErrorWithSuggestion(
			types.ErrMissingSubscribeObjects,
			fmt.Sprintf("%s.subscribe.objects", basePath),
			types.RuleSubscribeObjects,
			"Add at least one object to the subscribe.objects list",
		)
		return
	}

	// Validate each object
	for i, obj := range *integration.Subscribe.Objects {
		validateSubscribeObject(ctx, integration, obj, basePath, i)
	}
}

// validateSubscribeObject validates a single subscribe object.
func validateSubscribeObject(ctx *ValidationContext, integration openapi.Integration, obj openapi.IntegrationSubscribeObject, basePath string, index int) {
	objectPath := fmt.Sprintf("%s.subscribe.objects[%d]", basePath, index)

	// Check required fields
	if obj.ObjectName == "" {
		ctx.AddErrorWithSuggestion(
			"Object name is required",
			fmt.Sprintf("%s.objectName", objectPath),
			types.RuleRequiredField,
			"Add an objectName for this subscribe object",
		)
	} else {
		// Validate object name against provider schema
		validateObjectNameForSubscribe(ctx, integration.Provider, integration.Module, obj.ObjectName, fmt.Sprintf("%s.objectName", objectPath))
	}

	if obj.Destination == "" {
		ctx.AddErrorWithSuggestion(
			"Destination is required",
			fmt.Sprintf("%s.destination", objectPath),
			types.RuleRequiredField,
			"Add a destination for this subscribe object",
		)
	}

	// Check that inheritFieldsAndMapping is true (v1 requirement)
	if !obj.InheritFieldsAndMapping {
		ctx.AddErrorWithSuggestion(
			types.ErrSubscribeInheritFieldsAndMapping,
			fmt.Sprintf("%s.inheritFieldsAndMapping", objectPath),
			types.RuleSubscribeInheritFields,
			"Set inheritFieldsAndMapping to true",
		)
	}

	// Validate update event if present
	if obj.UpdateEvent != nil {
		validateUpdateEvent(ctx, obj.UpdateEvent, fmt.Sprintf("%s.updateEvent", objectPath))
	}
}

// validateUpdateEvent validates the update event configuration.
func validateUpdateEvent(ctx *ValidationContext, event *openapi.UpdateEvent, path string) {
	// Check that enabled is 'always' if specified
	if event.Enabled != nil && *event.Enabled != openapi.Always {
		ctx.AddErrorWithSuggestion(
			types.ErrInvalidInputEnabled,
			fmt.Sprintf("%s.enabled", path),
			types.RuleUpdateEventEnabled,
			"Set enabled to 'always' or remove the field",
		)
	}

	// Check watch fields configuration
	hasRequiredWatchFields := event.RequiredWatchFields != nil && len(*event.RequiredWatchFields) > 0
	hasWatchFieldsAuto := event.WatchFieldsAuto != nil &&
		(*event.WatchFieldsAuto == openapi.UpdateEventWatchFieldsAutoAll || *event.WatchFieldsAuto == openapi.UpdateEventWatchFieldsAutoSelected)

	// Must have either RequiredWatchFields OR WatchFieldsAuto, but not both
	if !hasRequiredWatchFields && !hasWatchFieldsAuto {
		ctx.AddErrorWithSuggestion(
			types.ErrWatchFieldsRequired,
			path,
			types.RuleUpdateEventWatchFields,
			"Add either requiredWatchFields (with at least 1 field) or set watchFieldsAuto to 'all' or 'selected'",
		)
	}

	if hasRequiredWatchFields && hasWatchFieldsAuto {
		ctx.AddErrorWithSuggestion(
			types.ErrWatchFieldsAndRequiredWatchFields,
			path,
			types.RuleUpdateEventWatchFields,
			"Use either requiredWatchFields or watchFieldsAuto, not both",
		)
	}
}

// validateObjectNameForSubscribe validates that an object name exists in the provider's schema.
func validateObjectNameForSubscribe(ctx *ValidationContext, provider string, module string, objectName string, path string) {
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
