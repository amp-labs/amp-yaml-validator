// Package ampyamlvalidator provides validation for Ampersand amp.yaml manifest files.
package ampyamlvalidator

import (
	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
	"github.com/amp-labs/amp-yaml-validator/validator"
)

// ValidationResult contains the outcome of validating an amp.yaml file.
type ValidationResult = types.ValidationResult

// ValidationIssue represents a single validation error or warning with location information.
type ValidationIssue = types.ValidationIssue

// Validator orchestrates validation of amp.yaml files.
type Validator = validator.Validator

// Option is a functional option for configuring the Validator.
type Option = validator.Option

// NewValidator creates a new validator with the given options.
func NewValidator(opts ...Option) *Validator {
	return validator.NewValidator(opts...)
}

// ValidateFile reads a YAML file and validates it.
// This is a convenience function that creates a validator and validates the file.
func ValidateFile(yamlPath string, opts ...Option) (*ValidationResult, error) {
	v := NewValidator(opts...)
	return v.ValidateFile(yamlPath)
}

// ValidateBytes validates YAML bytes.
// This is a convenience function that creates a validator and validates the bytes.
func ValidateBytes(yamlBytes []byte, opts ...Option) (*ValidationResult, error) {
	v := NewValidator(opts...)
	return v.ValidateBytes(yamlBytes)
}

// ValidateManifest validates an already-parsed manifest.
// This is a convenience function that creates a validator and validates the manifest.
func ValidateManifest(manifest *openapi.Manifest, opts ...Option) (*ValidationResult, error) {
	v := NewValidator(opts...)
	return v.ValidateManifest(manifest)
}

// WithStrictMode treats warnings as errors.
func WithStrictMode(strict bool) Option {
	return validator.WithStrictMode(strict)
}

// WithSkipProviderValidation skips provider-specific validation (for Phase 3).
func WithSkipProviderValidation() Option {
	return validator.WithSkipProviderValidation()
}

// WithSkipAsyncValidation skips async error prevention validation (for Phase 3).
func WithSkipAsyncValidation() Option {
	return validator.WithSkipAsyncValidation()
}
