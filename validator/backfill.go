package validator

import (
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
	"github.com/go-playground/validator/v10"
)

// validateBackfill validates backfill configuration.
func validateBackfill(ctx *ValidationContext, backfill *openapi.Backfill, path string) {
	if backfill == nil {
		// No backfill is valid
		return
	}

	// Validate struct using validator tags
	validate := validator.New()
	if err := validate.Struct(backfill); err != nil {
		ctx.AddErrorWithSuggestion(
			fmt.Sprintf("%s: %v", types.ErrInvalidBackfill, err),
			path,
			types.RuleBackfillConfig,
			"Check backfill configuration structure",
		)

		return
	}

	// Check mutual exclusivity of Days and FullHistory
	hasDays := backfill.DefaultPeriod.Days != nil
	hasFullHistory := backfill.DefaultPeriod.FullHistory != nil

	if hasDays && hasFullHistory {
		ctx.AddErrorWithSuggestion(
			"You must provide only one of 'days' or 'fullHistory'",
			path+".defaultPeriod",
			types.RuleBackfillConfig,
			"Remove either 'days' or 'fullHistory' from the backfill configuration",
		)
	}

	if !hasDays && !hasFullHistory {
		ctx.AddErrorWithSuggestion(
			"You must provide either 'days' or 'fullHistory'",
			path+".defaultPeriod",
			types.RuleBackfillConfig,
			"Add either 'days' (e.g., 30) or 'fullHistory' (true) to the backfill configuration",
		)
	}
}
