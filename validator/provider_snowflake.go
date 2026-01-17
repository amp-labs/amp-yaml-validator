package validator

import (
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateSnowflakeRules validates Snowflake-specific constraints.
// Snowflake only supports fullHistory backfill due to its data warehouse nature.
// Time-based backfill (days) is not supported.
func validateSnowflakeRules(valCtx *ValidationContext) {
	manifest := valCtx.Manifest

	for i, integration := range manifest.Integrations {
		// Only apply to Snowflake provider
		if integration.Provider != "snowflake" {
			continue
		}

		basePath := fmt.Sprintf("$.integrations[%d]", i)

		// Validate read objects
		if integration.Read != nil && integration.Read.Objects != nil {
			for j, obj := range *integration.Read.Objects {
				validateSnowflakeBackfill(valCtx, obj, basePath, j)
			}
		}
	}
}

// validateSnowflakeBackfill validates that Snowflake backfill uses fullHistory, not days.
func validateSnowflakeBackfill(
	valCtx *ValidationContext,
	obj openapi.IntegrationObject,
	basePath string,
	index int,
) {
	objPath := fmt.Sprintf("%s.read.objects[%d]", basePath, index)

	// Check if backfill is configured
	if obj.Backfill == nil {
		return
	}

	backfillPath := objPath + ".backfill.defaultPeriod"

	// Snowflake only supports fullHistory, not days-based backfill
	if obj.Backfill.DefaultPeriod.Days != nil {
		valCtx.AddErrorWithSuggestion(
			types.ErrSnowflakeBackfillDays,
			backfillPath+".days",
			types.RuleSnowflakeBackfill,
			"Remove days and use fullHistory: true instead",
		)
	}
}
