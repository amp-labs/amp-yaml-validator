package validator

import (
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateFieldMappings validates field mappings for uniqueness.
// NOTE: This validation is currently disabled because SelectedFieldMappings is not present
// on the Manifest IntegrationObject type. Field mapping validation is only applicable to
// ConfigContent types, not Manifest types. If needed in the future, reintroduce using
// ConfigContent types rather than Manifest types.
func validateFieldMappings(ctx *ValidationContext, read *openapi.IntegrationRead, basePath string) {
	// Disabled - SelectedFieldMappings not available on Manifest types
	_ = ctx
	_ = read
	_ = basePath

	return
	//	if read == nil || read.Objects == nil {
	//		return
	//	}
	//
	// // Validate each object's field mappings
	//
	//	for i, obj := range *read.Objects {
	//		objectPath := fmt.Sprintf("%s.read.objects[%d]", basePath, i)
	//		if obj.SelectedFieldMappings == nil || len(*obj.SelectedFieldMappings) == 0 {
	//			continue
	//		}
	//		checkDuplicateMappings(ctx, *obj.SelectedFieldMappings, objectPath)
	//	}
}

// checkDuplicateMappings checks for duplicate destination field names in field mappings.
func checkDuplicateMappings(ctx *ValidationContext, mappings map[string]string, basePath string) {
	// Track used destination field names
	usedDestinations := make(map[string]string) // destination -> source

	for sourceField, destField := range mappings {
		if existingSource, exists := usedDestinations[destField]; exists {
			// Found a duplicate destination
			ctx.AddErrorWithSuggestion(
				fmt.Sprintf("Duplicate field mapping: field '%s' and '%s' both map to '%s'",
					existingSource, sourceField, destField),
				fmt.Sprintf("%s.selectedFieldMappings", basePath),
				types.RuleFieldMappingUnique,
				fmt.Sprintf("Each destination field name must be unique. Consider using a different name for one of the mappings."),
			)
		} else {
			usedDestinations[destField] = sourceField
		}
	}
}
