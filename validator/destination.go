package validator

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/checker"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateDestinationsExist validates that all referenced destinations exist and are accessible.
// This validation only runs if a DestinationChecker is provided via dependency injection.
func validateDestinationsExist(ctx context.Context, valCtx *ValidationContext) {
	// Skip if no destination checker available
	if valCtx.DestinationChecker == nil {
		return
	}

	if valCtx.Manifest.Integrations == nil {
		return
	}

	// Collect all unique destination names with their paths
	destinationRefs := make(map[string][]string)

	for i, integration := range valCtx.Manifest.Integrations {
		basePath := fmt.Sprintf("$.integrations[%d]", i)

		// Check read objects
		if integration.Read != nil && integration.Read.Objects != nil {
			for j, obj := range *integration.Read.Objects {
				if obj.Destination != "" {
					path := fmt.Sprintf("%s.read.objects[%d].destination", basePath, j)
					destinationRefs[obj.Destination] = append(destinationRefs[obj.Destination], path)
				}
			}
		}

		// Note: Write objects don't have destination fields - they inherit from read objects

		// Check subscribe objects
		if integration.Subscribe != nil && integration.Subscribe.Objects != nil {
			for j, obj := range *integration.Subscribe.Objects {
				if obj.Destination != "" {
					path := fmt.Sprintf("%s.subscribe.objects[%d].destination", basePath, j)
					destinationRefs[obj.Destination] = append(destinationRefs[obj.Destination], path)
				}
			}
		}
	}

	// Validate each unique destination
	for destName, paths := range destinationRefs {
		validateSingleDestination(ctx, valCtx, destName, paths)
	}
}

// validateSingleDestination validates a single destination exists and is accessible.
func validateSingleDestination(ctx context.Context, valCtx *ValidationContext, destName string, paths []string) {
	err := valCtx.DestinationChecker.CheckDestination(ctx, destName)
	if err != nil {
		// Checker doesn't support this check - skip validation silently
		if errors.Is(err, checker.ErrNotSupported) {
			return
		}

		// Check for specific error types
		if errors.Is(err, checker.ErrDestinationNotFound) {
			for _, path := range paths {
				valCtx.AddErrorWithSuggestion(
					fmt.Sprintf("Destination '%s' not found. Create the destination before deploying this integration.",
						destName),
					path,
					types.RuleDestinationNotFound,
					fmt.Sprintf("Create a destination named '%s' using the amp_create_destination tool or Ampersand dashboard",
						destName),
				)
			}
		} else {
			// Other errors (network issues, auth failures, etc.)
			for _, path := range paths {
				valCtx.AddWarning(
					fmt.Sprintf("Unable to verify destination '%s': %v. Ensure the destination exists before deploying.",
						destName, err),
					path,
					types.RuleDestinationCheckFailed,
				)
			}
		}
	}
}
