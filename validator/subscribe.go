package validator

import (
	"context"
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateSubscribe validates the subscribe action.
func validateSubscribe(
	ctx context.Context,
	valCtx *ValidationContext,
	integration openapi.Integration,
	basePath string,
) {
	if integration.Subscribe == nil {
		return
	}

	// Check that read is also defined (subscribe requires read)
	if integration.Read == nil {
		valCtx.AddErrorWithSuggestion(
			types.ErrSubscribeRequiresRead,
			basePath+".subscribe",
			types.RuleSubscribeRequiresRead,
			"Add a read section to this integration",
		)
	}

	// Check that objects list exists and is not empty
	if integration.Subscribe.Objects == nil || len(*integration.Subscribe.Objects) == 0 {
		valCtx.AddErrorWithSuggestion(
			types.ErrMissingSubscribeObjects,
			basePath+".subscribe.objects",
			types.RuleSubscribeObjects,
			"Add at least one object to the subscribe.objects list",
		)

		return
	}

	// Validate each object
	for i, obj := range *integration.Subscribe.Objects {
		validateSubscribeObject(ctx, valCtx, integration, obj, basePath, i)
	}
}

// validateSubscribeObject validates a single subscribe object.
func validateSubscribeObject(
	ctx context.Context,
	valCtx *ValidationContext,
	integration openapi.Integration,
	obj openapi.IntegrationSubscribeObject,
	basePath string,
	index int,
) {
	objectPath := fmt.Sprintf("%s.subscribe.objects[%d]", basePath, index)

	// Check required fields
	if obj.ObjectName == "" {
		valCtx.AddErrorWithSuggestion(
			"Object name is required",
			objectPath+".objectName",
			types.RuleRequiredField,
			"Add an objectName for this subscribe object",
		)
	} else {
		// Validate object name against provider schema
		validateObjectNameForSubscribe(
			ctx, valCtx, integration.Provider, integration.Module,
			obj.ObjectName, objectPath+".objectName",
		)
	}

	if obj.Destination == "" {
		valCtx.AddErrorWithSuggestion(
			"Destination is required",
			objectPath+".destination",
			types.RuleRequiredField,
			"Add a destination for this subscribe object",
		)
	}

	// Check that inheritFieldsAndMapping is true (v1 requirement)
	if !obj.InheritFieldsAndMapping {
		valCtx.AddErrorWithSuggestion(
			types.ErrSubscribeInheritFieldsAndMapping,
			objectPath+".inheritFieldsAndMapping",
			types.RuleSubscribeInheritFields,
			"Set inheritFieldsAndMapping to true",
		)
	}

	// NEW: Validate that at least one event type is enabled
	validateSubscribeEventTypes(valCtx, obj, objectPath)

	// Validate update event if present
	if obj.UpdateEvent != nil {
		validateUpdateEvent(valCtx, obj.UpdateEvent, objectPath+".updateEvent")
	}
}

// validateUpdateEvent validates the update event configuration.
func validateUpdateEvent(ctx *ValidationContext, event *openapi.UpdateEvent, path string) {
	// Check that enabled is 'always' if specified
	if event.Enabled != nil && *event.Enabled != openapi.Always {
		ctx.AddErrorWithSuggestion(
			types.ErrInvalidInputEnabled,
			path+".enabled",
			types.RuleUpdateEventEnabled,
			"Set enabled to 'always' or remove the field",
		)
	}

	// Check watch fields configuration
	hasRequiredWatchFields := event.RequiredWatchFields != nil && len(*event.RequiredWatchFields) > 0
	hasWatchFieldsAuto := event.WatchFieldsAuto != nil &&
		(*event.WatchFieldsAuto == openapi.UpdateEventWatchFieldsAutoAll ||
			*event.WatchFieldsAuto == openapi.UpdateEventWatchFieldsAutoSelected)

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
func validateObjectNameForSubscribe(
	ctx context.Context,
	valCtx *ValidationContext,
	provider string,
	module string,
	objectName string,
	path string,
) {
	validateObjectNameCommon(ctx, valCtx, provider, module, objectName, path)
}
