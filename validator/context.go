package validator

import (
	"context"
	"errors"

	"github.com/amp-labs/amp-yaml-validator/catalog"
	"github.com/amp-labs/amp-yaml-validator/checker"
	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/parser"
	"github.com/amp-labs/amp-yaml-validator/types"
	"github.com/amp-labs/connectors/providers"
)

// ErrCatalogProviderNotInitialized indicates that the catalog provider was not initialized.
var ErrCatalogProviderNotInitialized = errors.New("catalog provider not initialized")

// ValidationContext holds the state during validation.
type ValidationContext struct {
	Manifest           *openapi.Manifest          // The parsed manifest to validate
	PositionMap        parser.PositionMap         // Map of YAML paths to line/column positions
	DirectiveMap       parser.DirectiveMap        // Map of amp:ignore directives for suppressing warnings
	UnknownKeys        []parser.UnknownKey        // Keys present in the YAML but not part of the schema
	Warnings           []types.ValidationIssue    // Accumulated validation issues
	Errors             []types.ValidationIssue    // Accumulated validation issues
	CatalogProvider    catalog.CatalogProvider    // Provider catalog access
	DestinationChecker checker.DestinationChecker // Optional destination checker for async validation
	ProviderAppChecker checker.ProviderAppChecker // Optional provider app/credentials checker
	RateLimitChecker   checker.RateLimitChecker   // Optional rate limit checker
}

// NewValidationContext creates a new validation context.
func NewValidationContext(
	manifest *openapi.Manifest,
	posMap parser.PositionMap,
	dirMap parser.DirectiveMap,
	catalogProvider catalog.CatalogProvider,
	destinationChecker checker.DestinationChecker,
	providerAppChecker checker.ProviderAppChecker,
	rateLimitChecker checker.RateLimitChecker,
) *ValidationContext {
	// Use default catalog provider if none provided
	if catalogProvider == nil {
		catalogProvider = catalog.NewDefaultCatalogProvider()
	}

	return &ValidationContext{
		Manifest:           manifest,
		PositionMap:        posMap,
		DirectiveMap:       dirMap,
		Errors:             []types.ValidationIssue{},
		Warnings:           []types.ValidationIssue{},
		CatalogProvider:    catalogProvider,
		DestinationChecker: destinationChecker, // Can be nil
		ProviderAppChecker: providerAppChecker, // Can be nil
		RateLimitChecker:   rateLimitChecker,   // Can be nil
	}
}

// AddError adds an error-level issue to the context.
func (vc *ValidationContext) AddError(message, path, rule string) {
	pos := vc.GetPosition(path)
	vc.Errors = append(vc.Errors, types.NewValidationIssue(message, path, rule, pos.Line, pos.Column))
}

// AddErrorWithSuggestion adds an error-level issue with a suggestion.
func (vc *ValidationContext) AddErrorWithSuggestion(message, path, rule, suggestion string) {
	pos := vc.GetPosition(path)
	issue := types.NewValidationIssue(message, path, rule, pos.Line, pos.Column)
	issue.Suggestion = suggestion
	vc.Errors = append(vc.Errors, issue)
}

// AddWarning adds a warning-level issue to the context.
// Warnings can be suppressed using amp:ignore directives in the YAML.
func (vc *ValidationContext) AddWarning(message, path, rule string) {
	// Check if this warning should be suppressed by an amp:ignore directive
	if vc.DirectiveMap.ShouldIgnore(path, rule) {
		return
	}

	pos := vc.GetPosition(path)
	vc.Warnings = append(vc.Warnings, types.NewValidationIssue(message, path, rule, pos.Line, pos.Column))
}

// AddWarningWithSuggestion adds a warning-level issue with a suggestion.
// Warnings can be suppressed using amp:ignore directives in the YAML.
func (vc *ValidationContext) AddWarningWithSuggestion(message, path, rule, suggestion string) {
	// Check if this warning should be suppressed by an amp:ignore directive
	if vc.DirectiveMap.ShouldIgnore(path, rule) {
		return
	}

	pos := vc.GetPosition(path)
	issue := types.NewValidationIssue(message, path, rule, pos.Line, pos.Column)
	issue.Suggestion = suggestion
	vc.Warnings = append(vc.Warnings, issue)
}

// GetPosition looks up the position for a given path.
func (vc *ValidationContext) GetPosition(path string) parser.Position {
	return vc.PositionMap.GetOrDefault(path)
}

// GetErrors returns all error-level issues.
func (vc *ValidationContext) GetErrors() []types.ValidationIssue {
	return vc.Errors
}

// GetWarnings returns all warning-level issues.
func (vc *ValidationContext) GetWarnings() []types.ValidationIssue {
	return vc.Warnings
}

// GetProviderInfo retrieves provider information from the catalog.
func (vc *ValidationContext) GetProviderInfo(
	ctx context.Context,
	providerName string,
) (*providers.ProviderInfo, error) {
	if vc.CatalogProvider == nil {
		return nil, ErrCatalogProviderNotInitialized
	}

	return vc.CatalogProvider.GetProviderInfo(ctx, providerName)
}

// HasCatalogAccess checks if the catalog is accessible.
func (vc *ValidationContext) HasCatalogAccess(ctx context.Context) bool {
	if vc.CatalogProvider == nil {
		return false
	}

	// Use Ping() to check if catalog is accessible
	return vc.CatalogProvider.Ping(ctx) == nil
}
