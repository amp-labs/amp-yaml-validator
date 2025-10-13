package validator

import (
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateSpecVersion checks that the spec version is valid.
func validateSpecVersion(ctx *ValidationContext) {
	path := "$.specVersion"

	if ctx.Manifest.SpecVersion == "" {
		ctx.AddErrorWithSuggestion(
			"Missing spec version",
			path,
			types.RuleSpecVersion,
			fmt.Sprintf("Add specVersion: \"%s\" to your amp.yaml", types.CurrentSpecVersion),
		)
		return
	}

	if ctx.Manifest.SpecVersion != types.CurrentSpecVersion {
		ctx.AddErrorWithSuggestion(
			fmt.Sprintf("Invalid spec version: %s, the current spec version is %s",
				ctx.Manifest.SpecVersion, types.CurrentSpecVersion),
			path,
			types.RuleSpecVersion,
			fmt.Sprintf("Update specVersion to \"%s\"", types.CurrentSpecVersion),
		)
	}
}
