package validator

import (
	"fmt"
	"os"

	"github.com/amp-labs/amp-yaml-validator/catalog"
	"github.com/amp-labs/amp-yaml-validator/checker"
	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/parser"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// Validator orchestrates validation of amp.yaml files.
type Validator struct {
	strictMode             bool
	skipProviderValidation bool
	skipAsyncValidation    bool
	catalogProvider        catalog.CatalogProvider
	destinationChecker     checker.DestinationChecker
	providerAppChecker     checker.ProviderAppChecker
	rateLimitChecker       checker.RateLimitChecker
}

// Option is a functional option for configuring the Validator.
type Option func(*Validator)

// WithStrictMode treats warnings as errors.
func WithStrictMode(strict bool) Option {
	return func(v *Validator) {
		v.strictMode = strict
	}
}

// WithSkipProviderValidation skips provider-specific validation (for Phase 3).
func WithSkipProviderValidation() Option {
	return func(v *Validator) {
		v.skipProviderValidation = true
	}
}

// WithSkipAsyncValidation skips async error prevention validation (for Phase 3).
func WithSkipAsyncValidation() Option {
	return func(v *Validator) {
		v.skipAsyncValidation = true
	}
}

// WithCatalogProvider injects a custom catalog provider (useful for testing).
func WithCatalogProvider(provider catalog.CatalogProvider) Option {
	return func(v *Validator) {
		v.catalogProvider = provider
	}
}

// WithDestinationChecker injects a custom destination checker for validating destinations.
// This allows client-side and server-side implementations to provide their own logic
// for checking if destinations exist and are accessible.
func WithDestinationChecker(checker checker.DestinationChecker) Option {
	return func(v *Validator) {
		v.destinationChecker = checker
	}
}

// WithProviderAppChecker injects a custom provider app checker for validating provider credentials.
// This allows client-side and server-side implementations to provide their own logic
// for checking if provider apps/OAuth credentials are configured.
func WithProviderAppChecker(checker checker.ProviderAppChecker) Option {
	return func(v *Validator) {
		v.providerAppChecker = checker
	}
}

// WithRateLimitChecker injects a custom rate limit checker for provider-specific rate limit info.
// This allows implementations to provide provider-specific or account-specific rate limit
// recommendations for schedule validation.
func WithRateLimitChecker(checker checker.RateLimitChecker) Option {
	return func(v *Validator) {
		v.rateLimitChecker = checker
	}
}

// NewValidator creates a new validator with the given options.
func NewValidator(opts ...Option) *Validator {
	v := &Validator{
		strictMode:             false,
		skipProviderValidation: false,
		skipAsyncValidation:    false,
		catalogProvider:        nil, // Will use default if not provided
		destinationChecker:     nil, // Optional, nil by default
		providerAppChecker:     nil, // Optional, nil by default
		rateLimitChecker:       nil, // Optional, nil by default
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// ValidateFile reads a YAML file and validates it.
func (v *Validator) ValidateFile(yamlPath string) (*types.ValidationResult, error) {
	yamlBytes, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return v.ValidateBytes(yamlBytes)
}

// ValidateBytes validates YAML bytes.
func (v *Validator) ValidateBytes(yamlBytes []byte) (*types.ValidationResult, error) {
	// Parse YAML
	manifest, posMap, dirMap, err := parser.ParseYAML(yamlBytes)
	if err != nil {
		return nil, err
	}

	// Create validation context
	ctx := NewValidationContext(manifest, posMap, dirMap, v.catalogProvider, v.destinationChecker, v.providerAppChecker, v.rateLimitChecker)

	// Run all validators
	v.runValidators(ctx)

	// Build result
	errors := ctx.GetErrors()
	warnings := ctx.GetWarnings()

	valid := len(errors) == 0
	if v.strictMode {
		valid = valid && len(warnings) == 0
	}

	return &types.ValidationResult{
		Valid:    valid,
		Errors:   errors,
		Warnings: warnings,
	}, nil
}

// ValidateManifest validates an already-parsed manifest.
func (v *Validator) ValidateManifest(manifest *openapi.Manifest) (*types.ValidationResult, error) {
	// Create empty position map and directive map (line numbers will be 0, no directives)
	posMap := parser.NewPositionMap()
	dirMap := parser.NewDirectiveMap()
	ctx := NewValidationContext(manifest, posMap, dirMap, v.catalogProvider, v.destinationChecker, v.providerAppChecker, v.rateLimitChecker)

	// Run all validators
	v.runValidators(ctx)

	// Build result
	errors := ctx.GetErrors()
	warnings := ctx.GetWarnings()

	valid := len(errors) == 0
	if v.strictMode {
		valid = valid && len(warnings) == 0
	}

	return &types.ValidationResult{
		Valid:    valid,
		Errors:   errors,
		Warnings: warnings,
	}, nil
}

// runValidators runs all validation rules in sequence.
func (v *Validator) runValidators(ctx *ValidationContext) {
	// Universal validation
	validateSpecVersion(ctx)
	validateIntegrations(ctx)

	// Provider-specific validation
	if !v.skipProviderValidation {
		validateProviderSpecific(ctx)
	}

	// Async error prevention validation
	if !v.skipAsyncValidation {
		validateAsyncRisks(ctx)
	}
}
