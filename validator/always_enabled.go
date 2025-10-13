package validator

import (
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateAlwaysEnabledObject validates constraints for always-enabled objects.
func validateAlwaysEnabledObject(ctx *ValidationContext, obj openapi.IntegrationObject, path string) {
	// Check that requiredFields is not empty
	if obj.RequiredFields == nil || len(*obj.RequiredFields) == 0 {
		ctx.AddErrorWithSuggestion(
			"Required fields are required for always enabled objects",
			fmt.Sprintf("%s.requiredFields", path),
			types.RuleAlwaysEnabledFields,
			"Add at least one required field for this always-enabled object",
		)
	} else {
		// Iterate through required fields
		for i, field := range *obj.RequiredFields {
			// Check if it's an existent field type
			fieldExistent, err := field.AsIntegrationFieldExistent()
			if err == nil && fieldExistent.MapToName != "" {
				ctx.AddErrorWithSuggestion(
					"mapToName is not allowed for always enabled objects",
					fmt.Sprintf("%s.requiredFields[%d].mapToName", path, i),
					types.RuleAlwaysEnabledFields,
					"Remove mapToName from required fields for always-enabled objects. Use fieldName only.",
				)
			}
		}
	}

	// Check that schedule is not empty
	if obj.Schedule == "" {
		ctx.AddErrorWithSuggestion(
			"Schedule is required for always enabled objects",
			fmt.Sprintf("%s.schedule", path),
			types.RuleAlwaysEnabledFields,
			"Add a schedule for this always-enabled object",
		)
	}
}
