package validator

import (
	"github.com/amp-labs/amp-yaml-validator/openapi"
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
} //nolint:wsl // Commented code block
