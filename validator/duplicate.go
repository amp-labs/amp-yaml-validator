package validator

import (
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateDuplicateObjects checks for duplicate objects within the same action.
// This prevents ambiguous configurations and runtime failures.
func validateDuplicateObjects(valCtx *ValidationContext) {
	manifest := valCtx.Manifest

	for i, integration := range manifest.Integrations {
		basePath := fmt.Sprintf("$.integrations[%d]", i)

		// Check read objects
		if integration.Read != nil && integration.Read.Objects != nil {
			checkReadDuplicates(valCtx, *integration.Read.Objects, basePath+".read.objects")
		}

		// Check write objects
		if integration.Write != nil && integration.Write.Objects != nil {
			checkWriteDuplicates(valCtx, *integration.Write.Objects, basePath+".write.objects")
		}

		// Check subscribe objects
		if integration.Subscribe != nil && integration.Subscribe.Objects != nil {
			checkSubscribeDuplicates(valCtx, *integration.Subscribe.Objects, basePath+".subscribe.objects")
		}
	}
}

// checkReadDuplicates finds and reports duplicate read objects.
func checkReadDuplicates(valCtx *ValidationContext, objects []openapi.IntegrationObject, basePath string) {
	seen := make(map[string]int) // objectName -> first occurrence index

	for i, obj := range objects {
		objectName := obj.ObjectName

		if firstIdx, exists := seen[objectName]; exists {
			// Found a duplicate - report error at current location
			path := fmt.Sprintf("%s[%d].objectName", basePath, i)
			firstPath := fmt.Sprintf("%s[%d].objectName", basePath, firstIdx)

			valCtx.AddErrorWithSuggestion(
				fmt.Sprintf("%s: '%s' (first occurrence at %s)", types.ErrDuplicateReadObject, objectName, firstPath),
				path,
				types.RuleDuplicateReadObject,
				fmt.Sprintf("Remove duplicate object '%s' or use different object names", objectName),
			)
		} else {
			seen[objectName] = i
		}
	}
}

// checkWriteDuplicates finds and reports duplicate write objects.
func checkWriteDuplicates(valCtx *ValidationContext, objects []openapi.IntegrationWriteObject, basePath string) {
	seen := make(map[string]int) // objectName -> first occurrence index

	for i, obj := range objects {
		objectName := obj.ObjectName

		if firstIdx, exists := seen[objectName]; exists {
			// Found a duplicate - report error at current location
			path := fmt.Sprintf("%s[%d].objectName", basePath, i)
			firstPath := fmt.Sprintf("%s[%d].objectName", basePath, firstIdx)

			valCtx.AddErrorWithSuggestion(
				fmt.Sprintf("%s: '%s' (first occurrence at %s)", types.ErrDuplicateWriteObject, objectName, firstPath),
				path,
				types.RuleDuplicateWriteObject,
				fmt.Sprintf("Remove duplicate object '%s' or use different object names", objectName),
			)
		} else {
			seen[objectName] = i
		}
	}
}

// checkSubscribeDuplicates finds and reports duplicate subscribe objects.
func checkSubscribeDuplicates(valCtx *ValidationContext, objects []openapi.IntegrationSubscribeObject, basePath string) {
	seen := make(map[string]int) // objectName -> first occurrence index

	for i, obj := range objects {
		objectName := obj.ObjectName

		if firstIdx, exists := seen[objectName]; exists {
			// Found a duplicate - report error at current location
			path := fmt.Sprintf("%s[%d].objectName", basePath, i)
			firstPath := fmt.Sprintf("%s[%d].objectName", basePath, firstIdx)

			valCtx.AddErrorWithSuggestion(
				fmt.Sprintf("%s: '%s' (first occurrence at %s)", types.ErrDuplicateSubscribeObject, objectName, firstPath),
				path,
				types.RuleDuplicateSubscribeObject,
				fmt.Sprintf("Remove duplicate object '%s' or use different object names", objectName),
			)
		} else {
			seen[objectName] = i
		}
	}
}
