package validator

import (
	"context"
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateWrite validates the write action.
func validateWrite(
	ctx context.Context,
	valCtx *ValidationContext,
	integration openapi.Integration,
	write *openapi.IntegrationWrite,
	basePath string,
) {
	if write == nil {
		return
	}

	// Check that objects list exists and is not empty
	if write.Objects == nil || len(*write.Objects) == 0 {
		valCtx.AddErrorWithSuggestion(
			types.ErrMissingWriteObjects,
			basePath+".write.objects",
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
				objectPath+".objectName",
				types.RuleRequiredField,
				"Add an objectName for this write object",
			)
		} else {
			// Validate object name against provider schema
			validateObjectNameForWrite(
				ctx, valCtx, integration.Provider, integration.Module,
				obj.ObjectName, objectPath+".objectName",
			)
		}
	}
}

// validateObjectNameForWrite validates that an object name exists in the provider's schema.
func validateObjectNameForWrite(
	ctx context.Context,
	valCtx *ValidationContext,
	provider string,
	module string,
	objectName string,
	path string,
) {
	validateObjectNameCommon(ctx, valCtx, provider, module, objectName, path)
}
