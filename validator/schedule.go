package validator

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/adhocore/gronx"
	"github.com/amp-labs/amp-yaml-validator/types"
)

const (
	cronFieldCount            = 5 // Standard cron format has 5 fields
	intervalPartsExpectedSize = 2 // Interval syntax like */10 has 2 parts
)

// validateSchedule validates cron schedule syntax and frequency.
func validateSchedule(ctx *ValidationContext, schedule string, path string) {
	gron := gronx.New()

	// Check cron syntax
	if !gron.IsValid(schedule) {
		ctx.AddErrorWithSuggestion(
			fmt.Sprintf("%s: %s", types.ErrInvalidCronSchedule, schedule),
			path,
			types.RuleScheduleSyntax,
			"Use valid cron syntax (e.g., '*/10 * * * *' for every 10 minutes)",
		)

		return
	}

	// Split schedule into parts (should be 5 parts for standard cron)
	parts := strings.Fields(schedule)
	if len(parts) != cronFieldCount {
		ctx.AddErrorWithSuggestion(
			types.ErrInvalidSchedule,
			path,
			types.RuleScheduleSyntax,
			"Use 5-part cron syntax: minute hour day month weekday",
		)

		return
	}

	// Check minute field (first part)
	minuteField := parts[0]

	// If minute field is "*", it means every minute, which is too frequent
	if minuteField == "*" {
		ctx.AddErrorWithSuggestion(
			"Schedule cannot have '*' in minute field. Try */n with minimum of n=10",
			path,
			types.RuleScheduleMinimumInterval,
			"Change schedule to */10 or greater (e.g., '*/10 * * * *' for every 10 minutes)",
		)

		return
	}

	// If minute field contains "/" (e.g., "*/15"), check the interval
	//nolint:nestif // Nested validation for interval parsing is justified
	if strings.Contains(minuteField, "/") {
		intervalParts := strings.Split(minuteField, "/")
		if len(intervalParts) == intervalPartsExpectedSize {
			if intervalVal, err := strconv.Atoi(intervalParts[1]); err == nil {
				if intervalVal < types.MinScheduleIntervalMinutes {
					ctx.AddErrorWithSuggestion(
						fmt.Sprintf("%s. Found: */%d * * * * (%d minutes)",
							types.ErrScheduleTooFrequent, intervalVal, intervalVal),
						path,
						types.RuleScheduleMinimumInterval,
						"Change schedule to */10 or greater (e.g., '*/10 * * * *' for every 10 minutes)",
					)

					return
				}
			}
		}
	}

	// Validate actual schedule frequency using gronx
	// Use PrevTickBefore and NextTickAfter to check the interval, mirroring server logic
	now := time.Now()

	past, err := gronx.PrevTickBefore(schedule, now, true)
	if err != nil {
		ctx.AddErrorWithSuggestion(
			fmt.Sprintf("Invalid schedule %s: %v", schedule, err),
			path,
			types.RuleScheduleSyntax,
			"Use valid cron syntax (e.g., '*/10 * * * *' for every 10 minutes)",
		)

		return
	}

	next, err := gronx.NextTickAfter(schedule, now, true)
	if err != nil {
		ctx.AddErrorWithSuggestion(
			fmt.Sprintf("Invalid schedule %s: %v", schedule, err),
			path,
			types.RuleScheduleSyntax,
			"Use valid cron syntax (e.g., '*/10 * * * *' for every 10 minutes)",
		)

		return
	}

	duration := next.Sub(past)
	minDuration := time.Duration(types.MinScheduleIntervalMinutes) * time.Minute

	if duration < minDuration {
		ctx.AddErrorWithSuggestion(
			fmt.Sprintf("Schedule interval must be at least %d minutes. Found: %s (%.0f minutes)",
				types.MinScheduleIntervalMinutes, schedule, duration.Minutes()),
			path,
			types.RuleScheduleMinimumInterval,
			"Change schedule to run at most every 10 minutes (e.g., '*/10 * * * *')",
		)
	}
}
