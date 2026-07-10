package validator

import (
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateUnknownKeys emits a warning for every key found in the YAML that is not
// part of the amp.yaml schema. Such "orphan" keys are silently dropped during
// unmarshaling, so they are commonly typos or misplaced configuration.
//
// Warnings report the key's own line/column (not the value's) and can be
// suppressed per-key or per-subtree with an amp:ignore directive.
func validateUnknownKeys(ctx *ValidationContext) {
	for _, uk := range ctx.UnknownKeys {
		if ctx.DirectiveMap.ShouldIgnore(uk.Path, types.RuleUnknownKey) {
			continue
		}

		message := fmt.Sprintf(
			"Unknown key %q is not part of the amp.yaml v1.0.0 schema and will be ignored",
			uk.Key,
		)

		ctx.Warnings = append(ctx.Warnings, types.NewValidationIssue(
			message, uk.Path, types.RuleUnknownKey, uk.Line, uk.Column,
		))
	}
}
