package validator

import (
	"context"
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateRead validates the read action.
func validateRead(
	ctx context.Context,
	valCtx *ValidationContext,
	integration openapi.Integration,
	read *openapi.IntegrationRead,
	basePath string,
) {
	if read == nil {
		return
	}

	// Check that objects list exists and is not empty
	if read.Objects == nil || len(*read.Objects) == 0 {
		valCtx.AddErrorWithSuggestion(
			types.ErrMissingReadObjects,
			basePath+".read.objects",
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
func validateReadObject(
	ctx context.Context,
	valCtx *ValidationContext,
	integration openapi.Integration,
	obj openapi.IntegrationObject,
	basePath string,
	index int,
) {
	objectPath := fmt.Sprintf("%s.read.objects[%d]", basePath, index)

	// Check required fields
	if obj.ObjectName == "" {
		valCtx.AddErrorWithSuggestion(
			"Object name is required",
			objectPath+".objectName",
			types.RuleRequiredField,
			"Add an objectName for this read object",
		)
	} else {
		// Validate object name against provider schema
		validateObjectName(ctx, valCtx, integration.Provider, integration.Module, obj.ObjectName, objectPath+".objectName")
	}

	if obj.Destination == "" {
		valCtx.AddErrorWithSuggestion(
			"Destination is required",
			objectPath+".destination",
			types.RuleRequiredField,
			"Add a destination for this read object",
		)
	}

	if obj.Schedule == "" {
		valCtx.AddErrorWithSuggestion(
			"Schedule is required",
			objectPath+".schedule",
			types.RuleRequiredField,
			"Add a schedule for this read object (e.g., '*/10 * * * *' for every 10 minutes)",
		)
	} else {
		// Validate schedule
		validateSchedule(valCtx, obj.Schedule, objectPath+".schedule")
	}

	// Validate delivery mode
	if obj.Delivery != nil {
		validateDeliveryMode(valCtx, obj.Delivery, objectPath+".delivery")
	}

	// Validate backfill
	if obj.Backfill != nil {
		validateBackfill(valCtx, obj.Backfill, objectPath+".backfill")
	}

	// If enabled is always, validate always-enabled constraints
	if obj.Enabled == openapi.IntegrationObjectEnabledAlways {
		validateAlwaysEnabledObject(valCtx, obj, objectPath)
	}
}

// validateObjectName validates that an object name exists in the provider's schema.
func validateObjectName(
	ctx context.Context,
	valCtx *ValidationContext,
	provider string,
	module string,
	objectName string,
	path string,
) {
	validateObjectNameCommon(ctx, valCtx, provider, module, objectName, path)
}
