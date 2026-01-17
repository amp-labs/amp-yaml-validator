package validator

import (
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateGoogleCalendarRules validates Google Calendar-specific constraints.
// Google Calendar has specific limitations for the events object:
// 1. Cannot use fullHistory backfill (API limitation)
// 2. Maximum 28 days backfill (API limitation).
func validateGoogleCalendarRules(valCtx *ValidationContext) {
	manifest := valCtx.Manifest

	for i, integration := range manifest.Integrations {
		// Only apply to Google Calendar provider
		if integration.Provider != "googlecalendar" {
			continue
		}

		basePath := fmt.Sprintf("$.integrations[%d]", i)

		// Validate read objects
		if integration.Read != nil && integration.Read.Objects != nil {
			for j, obj := range *integration.Read.Objects {
				if obj.ObjectName == "events" {
					validateGoogleCalendarEventsBackfill(valCtx, obj, basePath, j)
				}
			}
		}
	}
}

// validateGoogleCalendarEventsBackfill validates backfill configuration for Google Calendar events object.
func validateGoogleCalendarEventsBackfill(
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

	// Rule 1: Google Calendar events cannot use fullHistory
	if obj.Backfill.DefaultPeriod.FullHistory != nil && *obj.Backfill.DefaultPeriod.FullHistory {
		valCtx.AddErrorWithSuggestion(
			types.ErrGoogleCalendarFullHistory,
			backfillPath+".fullHistory",
			types.RuleGoogleCalendarBackfill,
			"Remove fullHistory or use a specific number of days (maximum 28)",
		)
	}

	// Rule 2: Maximum 28 days backfill for Google Calendar events
	if obj.Backfill.DefaultPeriod.Days != nil && *obj.Backfill.DefaultPeriod.Days > types.MaxGoogleCalendarBackfillDays {
		valCtx.AddErrorWithSuggestion(
			fmt.Sprintf("%s: %d days (maximum is %d)", types.ErrGoogleCalendarMaxBackfill,
				*obj.Backfill.DefaultPeriod.Days, types.MaxGoogleCalendarBackfillDays),
			backfillPath+".days",
			types.RuleGoogleCalendarBackfill,
			fmt.Sprintf("Set days to %d or less", types.MaxGoogleCalendarBackfillDays),
		)
	}
}
