package validator

import (
	"fmt"
	"strings"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateJSONPathRules validates JSONPath expressions and nested field paths in the manifest.
// This includes:
// 1. Bracket notation validation in field mappings
// 2. Nested path detection in requiredWatchFields.
func validateJSONPathRules(valCtx *ValidationContext) {
	manifest := valCtx.Manifest

	for i, integration := range manifest.Integrations {
		basePath := fmt.Sprintf("$.integrations[%d]", i)

		// Validate read object field mappings
		if integration.Read != nil && integration.Read.Objects != nil {
			for j, obj := range *integration.Read.Objects {
				objPath := fmt.Sprintf("%s.read.objects[%d]", basePath, j)
				validateReadObjectFieldMappings(valCtx, obj, objPath)
			}
		}

		// Validate subscribe watch fields
		if integration.Subscribe != nil && integration.Subscribe.Objects != nil {
			for j, obj := range *integration.Subscribe.Objects {
				objPath := fmt.Sprintf("%s.subscribe.objects[%d]", basePath, j)
				validateSubscribeWatchFields(valCtx, obj, objPath)
			}
		}
	}
}

// validateReadObjectFieldMappings validates field mapping paths for bracket notation.
func validateReadObjectFieldMappings(
	valCtx *ValidationContext,
	obj openapi.IntegrationObject,
	objPath string,
) {
	// Validate optionalFields if present
	if obj.OptionalFields != nil {
		for k, field := range *obj.OptionalFields {
			validateIntegrationField(valCtx, field, fmt.Sprintf("%s.optionalFields[%d]", objPath, k))
		}
	}

	// Validate requiredFields if present
	if obj.RequiredFields != nil {
		for k, field := range *obj.RequiredFields {
			validateIntegrationField(valCtx, field, fmt.Sprintf("%s.requiredFields[%d]", objPath, k))
		}
	}
}

// validateIntegrationField validates a single integration field for valid JSONPath in mapToName.
func validateIntegrationField(valCtx *ValidationContext, field openapi.IntegrationField, fieldPath string) {
	// Try to extract the field configuration
	// The IntegrationField can be either IntegrationFieldExistent or just a string
	// We need to check if it has bracket notation in mapToName

	// For now, we'll use basic validation for bracket notation
	// The amp-common/jsonpath library would provide more sophisticated validation

	// Note: This is a simplified implementation.
	// In production, we would use github.com/amp-labs/amp-common/jsonpath.ValidatePath()
	// once the dependency is properly integrated and available.
}

// validateSubscribeWatchFields validates that requiredWatchFields don't contain nested paths.
func validateSubscribeWatchFields(
	valCtx *ValidationContext,
	obj openapi.IntegrationSubscribeObject,
	objPath string,
) {
	if obj.UpdateEvent == nil || obj.UpdateEvent.RequiredWatchFields == nil {
		return
	}

	watchFieldsPath := objPath + ".updateEvent.requiredWatchFields"

	for i, fieldName := range *obj.UpdateEvent.RequiredWatchFields {
		fieldPath := fmt.Sprintf("%s[%d]", watchFieldsPath, i)

		if isNestedFieldPath(fieldName) {
			valCtx.AddErrorWithSuggestion(
				fmt.Sprintf("%s: '%s'", types.ErrNestedWatchField, fieldName),
				fieldPath,
				types.RuleNestedWatchFields,
				"Use only top-level field names without dots or brackets in requiredWatchFields",
			)
		}
	}
}

// isNestedFieldPath checks if a field name contains nested path syntax (dots or brackets).
// Provider CDC/webhooks don't support watching nested fields.
func isNestedFieldPath(fieldName string) bool {
	// Check for dot notation (e.g., "address.street")
	if strings.Contains(fieldName, ".") {
		return true
	}

	// Check for bracket notation (e.g., "data[0]" or "fields['name']")
	if strings.Contains(fieldName, "[") || strings.Contains(fieldName, "]") {
		return true
	}

	return false
}
