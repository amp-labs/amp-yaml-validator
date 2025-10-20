package validator

import (
	"context"
	"fmt"
	"time"

	"github.com/adhocore/gronx"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateAsyncRisks performs async error prevention checks (warnings).
// These are configuration issues that would cause runtime failures in
// asynchronous services (Temporal workflows, messenger services).
func validateAsyncRisks(ctx context.Context, valCtx *ValidationContext) {
	validateDestinationReferences(ctx, valCtx)
	validateObjectExistence(ctx, valCtx)
	validateBackfillRisks(valCtx)
	validateScheduleFrequencyRisks(valCtx)
}

// validateDestinationReferences warns about destination references that
// cannot be validated statically. At runtime, these destinations must exist
// and be active in the project.
func validateDestinationReferences(ctx context.Context, valCtx *ValidationContext) {
	manifest := valCtx.Manifest

	// Track unique destinations to avoid duplicate warnings
	destinations := make(map[string]string) // destination name -> first path that references it

	//nolint:varnamelen // i is idiomatic for index in loops
	for i, integration := range manifest.Integrations {
		// Check read destinations
		if integration.Read != nil && integration.Read.Objects != nil {
			for j, obj := range *integration.Read.Objects {
				if obj.Destination != "" {
					path := fmt.Sprintf("$.integrations[%d].read.objects[%d].destination", i, j)
					if _, seen := destinations[obj.Destination]; !seen {
						destinations[obj.Destination] = path
					}
				}
			}
		}

		// Note: Write objects don't have destinations in the schema

		// Check subscribe destinations
		if integration.Subscribe != nil && integration.Subscribe.Objects != nil {
			for j, obj := range *integration.Subscribe.Objects {
				if obj.Destination != "" {
					path := fmt.Sprintf("$.integrations[%d].subscribe.objects[%d].destination", i, j)
					if _, seen := destinations[obj.Destination]; !seen {
						destinations[obj.Destination] = path
					}
				}
			}
		}
	}

	// Check or warn about each unique destination
	for destName, path := range destinations {
		if valCtx.DestinationChecker != nil {
			// If a destination checker is provided, use it to validate
			if err := valCtx.DestinationChecker.CheckDestination(ctx, destName); err != nil {
				valCtx.AddErrorWithSuggestion(
					fmt.Sprintf("Destination %q does not exist or is not accessible: %v", destName, err),
					path,
					"destination-exists",
					fmt.Sprintf("Create destination %q in your project or check access permissions", destName),
				)
			}
		} else {
			// No checker provided, issue a warning reminder
			valCtx.AddWarningWithSuggestion(
				fmt.Sprintf("Destination %q is referenced but cannot be validated statically", destName),
				path,
				"destination-exists",
				fmt.Sprintf("Ensure destination %q is configured and active in your project before deploying", destName),
			)
		}
	}
}

// validateObjectExistence warns about objects that cannot be validated against
// the provider catalog. In the future, when object schemas are available,
// this can be upgraded to an error.
func validateObjectExistence(ctx context.Context, valCtx *ValidationContext) {
	manifest := valCtx.Manifest

	//nolint:varnamelen // i is idiomatic for index in loops
	for i, integration := range manifest.Integrations {
		providerInfo, err := valCtx.GetProviderInfo(ctx, integration.Provider)
		if err != nil {
			// Provider catalog unavailable, skip object validation
			continue
		}

		moduleName := integration.Module

		// Try to get object list from catalog
		objects, err := valCtx.CatalogProvider.ListObjects(ctx, integration.Provider, moduleName)
		if err != nil {
			// Object list not available (expected for now), issue warning
			// In the future, when object schemas are available, this will be more strict
			continue
		}

		// If we got an object list, validate that referenced objects exist
		objectSet := make(map[string]bool)
		for _, obj := range objects {
			objectSet[obj] = true
		}

		// Check read objects
		if integration.Read != nil && integration.Read.Objects != nil {
			for j, obj := range *integration.Read.Objects {
				if !objectSet[obj.ObjectName] {
					path := fmt.Sprintf("$.integrations[%d].read.objects[%d].objectName", i, j)
					valCtx.AddWarningWithSuggestion(
						fmt.Sprintf("Object %q is not found in provider %q catalog", obj.ObjectName, integration.Provider),
						path,
						types.RuleObjectExists,
						fmt.Sprintf("Verify that object %q exists in %s's API documentation", obj.ObjectName, providerInfo.DisplayName),
					)
				}
			}
		}

		// Check write objects
		if integration.Write != nil && integration.Write.Objects != nil {
			for j, obj := range *integration.Write.Objects {
				if !objectSet[obj.ObjectName] {
					path := fmt.Sprintf("$.integrations[%d].write.objects[%d].objectName", i, j)
					valCtx.AddWarningWithSuggestion(
						fmt.Sprintf("Object %q is not found in provider %q catalog", obj.ObjectName, integration.Provider),
						path,
						types.RuleObjectExists,
						fmt.Sprintf("Verify that object %q exists in %s's API documentation", obj.ObjectName, providerInfo.DisplayName),
					)
				}
			}
		}

		// Check subscribe objects
		if integration.Subscribe != nil && integration.Subscribe.Objects != nil {
			for j, obj := range *integration.Subscribe.Objects {
				if !objectSet[obj.ObjectName] {
					path := fmt.Sprintf("$.integrations[%d].subscribe.objects[%d].objectName", i, j)
					valCtx.AddWarningWithSuggestion(
						fmt.Sprintf("Object %q is not found in provider %q catalog", obj.ObjectName, integration.Provider),
						path,
						types.RuleObjectExists,
						fmt.Sprintf("Verify that object %q exists in %s's API documentation", obj.ObjectName, providerInfo.DisplayName),
					)
				}
			}
		}
	}
}

// validateBackfillRisks warns about large backfills that may cause workflow
// timeouts or performance issues at runtime.
func validateBackfillRisks(ctx *ValidationContext) {
	manifest := ctx.Manifest

	const largeBackfillThreshold = 180 // days

	//nolint:varnamelen // i is idiomatic for index in loops
	for i, integration := range manifest.Integrations {
		//nolint:nestif // Nested validation logic for hierarchical config is justified
		if integration.Read != nil && integration.Read.Objects != nil {
			//nolint:varnamelen // j is idiomatic for nested loop index
			for j, obj := range *integration.Read.Objects {
				if obj.Backfill != nil {
					period := &obj.Backfill.DefaultPeriod

					// Check for full history backfill
					if period.FullHistory != nil && *period.FullHistory {
						path := fmt.Sprintf(
							"$.integrations[%d].read.objects[%d].backfill.defaultPeriod.fullHistory",
							i, j,
						)

						ctx.AddWarningWithSuggestion(
							fmt.Sprintf(
								"Object %q is configured for full history backfill, which may cause timeouts with large datasets",
								obj.ObjectName,
							),
							path,
							"large-backfill-risk",
							"Consider using a limited days backfill first (e.g., 90 days) to test performance",
						)
					}

					// Check for large days backfill
					if period.Days != nil && *period.Days > largeBackfillThreshold {
						path := fmt.Sprintf(
							"$.integrations[%d].read.objects[%d].backfill.defaultPeriod.days",
							i, j,
						)

						ctx.AddWarningWithSuggestion(
							fmt.Sprintf(
								"Object %q has a large backfill period (%d days), which may cause timeouts with large datasets",
								obj.ObjectName,
								*period.Days,
							),
							path,
							"large-backfill-risk",
							fmt.Sprintf(
								"Consider reducing backfill to %d days or less for initial sync",
								largeBackfillThreshold,
							),
						)
					}
				}
			}
		}
	}
}

// validateScheduleFrequencyRisks warns about very frequent schedules that may
// hit API rate limits or cause throttling issues at runtime.
func validateScheduleFrequencyRisks(ctx *ValidationContext) {
	manifest := ctx.Manifest

	const frequentScheduleThreshold = 5 // minutes

	//nolint:varnamelen // i is idiomatic for index in loops
	for i, integration := range manifest.Integrations {
		if integration.Read != nil && integration.Read.Objects != nil {
			for j, obj := range *integration.Read.Objects {
				if obj.Schedule == "" {
					continue
				}

				schedule := obj.Schedule

				freq, err := getScheduleFrequency(schedule)
				if err != nil {
					continue // Skip if can't calculate frequency
				}

				frequencyMinutes := int(freq.Minutes())
				if frequencyMinutes <= frequentScheduleThreshold {
					path := fmt.Sprintf("$.integrations[%d].read.objects[%d].schedule", i, j)

					ctx.AddWarningWithSuggestion(
						fmt.Sprintf(
							"Object %q has a very frequent schedule (%d minutes), which may hit API rate limits with high-volume objects",
							obj.ObjectName,
							frequencyMinutes,
						),
						path,
						"frequent-schedule-risk",
						"Monitor API rate limits when using frequent sync schedules, especially for objects with many records",
					)
				}
			}
		}
	}
}

// getScheduleFrequency calculates the interval between schedule executions.
// This is a helper function used by schedule frequency risk validation.
func getScheduleFrequency(schedule string) (time.Duration, error) {
	now := time.Now()

	prevTick, err := gronx.PrevTickBefore(schedule, now, true)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate prev tick: %w", err)
	}

	nextTick, err := gronx.NextTickAfter(schedule, now, true)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate next tick: %w", err)
	}

	return nextTick.Sub(prevTick), nil
}
