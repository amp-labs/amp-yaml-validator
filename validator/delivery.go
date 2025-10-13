package validator

import (
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateDeliveryMode validates delivery mode configuration.
func validateDeliveryMode(ctx *ValidationContext, delivery *openapi.Delivery, path string) {
	if delivery == nil {
		// No delivery config means auto mode, which is valid
		return
	}

	// Check delivery mode
	if delivery.Mode == nil {
		// No mode specified means auto, which is valid
		return
	}

	switch *delivery.Mode {
	case openapi.Auto:
		// Auto mode should not have page size
		if delivery.PageSize != nil {
			ctx.AddErrorWithSuggestion(
				"Page size is not valid when delivery mode is auto",
				fmt.Sprintf("%s.pageSize", path),
				types.RuleDeliveryMode,
				"Remove pageSize or change delivery mode to 'onRequest'",
			)
		}

	case openapi.OnRequest:
		// OnRequest mode requires page size
		if delivery.PageSize == nil {
			ctx.AddErrorWithSuggestion(
				"Page size is required for on-request delivery mode",
				fmt.Sprintf("%s.pageSize", path),
				types.RuleDeliveryMode,
				fmt.Sprintf("Add pageSize between %d and %d", types.MinOnRequestPageSize, types.MaxOnRequestPageSize),
			)
		} else {
			// Check page size is within range
			if *delivery.PageSize < types.MinOnRequestPageSize || *delivery.PageSize > types.MaxOnRequestPageSize {
				ctx.AddErrorWithSuggestion(
					fmt.Sprintf("Page size must be between %d and %d. Found: %d",
						types.MinOnRequestPageSize, types.MaxOnRequestPageSize, *delivery.PageSize),
					fmt.Sprintf("%s.pageSize", path),
					types.RuleDeliveryMode,
					fmt.Sprintf("Set pageSize to a value between %d and %d", types.MinOnRequestPageSize, types.MaxOnRequestPageSize),
				)
			}
		}

	default:
		ctx.AddErrorWithSuggestion(
			"Invalid delivery mode",
			fmt.Sprintf("%s.mode", path),
			types.RuleDeliveryMode,
			"Use 'auto' or 'onRequest' as the delivery mode",
		)
	}
}
