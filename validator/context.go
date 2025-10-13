package validator

import (
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/catalog"
	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/parser"
	"github.com/amp-labs/amp-yaml-validator/types"
	"github.com/amp-labs/connectors/providers"
)

// ValidationContext holds the state during validation.
type ValidationContext struct {
	Manifest        *openapi.Manifest       // The parsed manifest to validate
	PositionMap     parser.PositionMap      // Map of YAML paths to line/column positions
	Issues          []types.ValidationIssue // Accumulated validation issues
	CatalogProvider catalog.CatalogProvider // Provider catalog access
}

// NewValidationContext creates a new validation context.
func NewValidationContext(manifest *openapi.Manifest, posMap parser.PositionMap, catalogProvider catalog.CatalogProvider) *ValidationContext {
	// Use default catalog provider if none provided
	if catalogProvider == nil {
		catalogProvider = catalog.NewDefaultCatalogProvider()
	}

	return &ValidationContext{
		Manifest:        manifest,
		PositionMap:     posMap,
		Issues:          []types.ValidationIssue{},
		CatalogProvider: catalogProvider,
	}
}

// AddError adds an error-level issue to the context.
func (vc *ValidationContext) AddError(message, path, rule string) {
	pos := vc.GetPosition(path)
	vc.Issues = append(vc.Issues, types.NewError(message, path, rule, pos.Line, pos.Column))
}

// AddErrorWithSuggestion adds an error-level issue with a suggestion.
func (vc *ValidationContext) AddErrorWithSuggestion(message, path, rule, suggestion string) {
	pos := vc.GetPosition(path)
	issue := types.NewError(message, path, rule, pos.Line, pos.Column)
	issue.Suggestion = suggestion
	vc.Issues = append(vc.Issues, issue)
}

// AddWarning adds a warning-level issue to the context.
func (vc *ValidationContext) AddWarning(message, path, rule string) {
	pos := vc.GetPosition(path)
	vc.Issues = append(vc.Issues, types.NewWarning(message, path, rule, pos.Line, pos.Column))
}

// AddWarningWithSuggestion adds a warning-level issue with a suggestion.
func (vc *ValidationContext) AddWarningWithSuggestion(message, path, rule, suggestion string) {
	pos := vc.GetPosition(path)
	issue := types.NewWarning(message, path, rule, pos.Line, pos.Column)
	issue.Suggestion = suggestion
	vc.Issues = append(vc.Issues, issue)
}

// GetPosition looks up the position for a given path.
func (vc *ValidationContext) GetPosition(path string) parser.Position {
	return vc.PositionMap.GetOrDefault(path)
}

// GetErrors returns all error-level issues.
func (vc *ValidationContext) GetErrors() []types.ValidationIssue {
	var errors []types.ValidationIssue

	for _, issue := range vc.Issues {
		if issue.Severity == "error" {
			errors = append(errors, issue)
		}
	}

	return errors
}

// GetWarnings returns all warning-level issues.
func (vc *ValidationContext) GetWarnings() []types.ValidationIssue {
	var warnings []types.ValidationIssue

	for _, issue := range vc.Issues {
		if issue.Severity == "warning" {
			warnings = append(warnings, issue)
		}
	}

	return warnings
}

// GetProviderInfo retrieves provider information from the catalog.
func (vc *ValidationContext) GetProviderInfo(providerName string) (*providers.ProviderInfo, error) {
	if vc.CatalogProvider == nil {
		return nil, fmt.Errorf("catalog provider not initialized")
	}

	return vc.CatalogProvider.GetProviderInfo(providerName)
}

// HasCatalogAccess checks if the catalog is accessible.
func (vc *ValidationContext) HasCatalogAccess() bool {
	if vc.CatalogProvider == nil {
		return false
	}

	// Use Ping() to check if catalog is accessible
	return vc.CatalogProvider.Ping() == nil
}
