package checker

import (
	"errors"
	"time"
)

// ErrDestinationNotFound indicates that a destination does not exist or is not accessible
var ErrDestinationNotFound = errors.New("destination not found")

// ErrProviderAppNotFound indicates that a provider app/credentials are not configured
var ErrProviderAppNotFound = errors.New("provider app not found")

// DestinationChecker defines an interface for checking if destinations exist.
// This abstraction allows different implementations for client-side vs server-side validation.
//
// Client-side implementations might:
//   - Return ErrNotSupported (no access to project data)
//   - Make API calls to check destination existence
//
// Server-side implementations might:
//   - Query the database directly
//   - Check project configuration
//   - Verify access permissions
type DestinationChecker interface {
	// CheckDestination verifies if a destination exists and is accessible.
	// Returns nil if the destination exists and is valid.
	// Returns ErrDestinationNotFound if the destination doesn't exist.
	// Returns other errors for unexpected failures (network issues, auth failures, etc.)
	CheckDestination(destinationName string) error
}

// MockDestinationChecker is a mock implementation of DestinationChecker for testing.
type MockDestinationChecker struct {
	destinations map[string]bool // destination name -> exists
}

// NewMockDestinationChecker creates a new mock destination checker with the given destinations.
func NewMockDestinationChecker(destinations map[string]bool) DestinationChecker {
	return &MockDestinationChecker{
		destinations: destinations,
	}
}

// CheckDestination checks if a destination exists in the mock.
func (m *MockDestinationChecker) CheckDestination(destinationName string) error {
	if m.destinations == nil {
		return errors.New("mock not configured")
	}

	exists, ok := m.destinations[destinationName]
	if !ok || !exists {
		return ErrDestinationNotFound
	}

	return nil
}

// ProviderAppChecker defines an interface for checking if provider apps/credentials are configured.
// This abstraction allows different implementations for client-side vs server-side validation.
//
// Client-side implementations might:
//   - Return ErrNotSupported (no access to project data)
//   - Make API calls to check provider app configuration
//
// Server-side implementations might:
//   - Query the database for OAuth app configurations
//   - Check project provider credentials
//   - Verify authentication setup
type ProviderAppChecker interface {
	// CheckProviderApp verifies if a provider has valid credentials/OAuth apps configured.
	// Returns nil if the provider app exists and is properly configured.
	// Returns ErrProviderAppNotFound if the provider app doesn't exist.
	// Returns other errors for unexpected failures (network issues, auth failures, etc.)
	CheckProviderApp(providerName string) error
}

// MockProviderAppChecker is a mock implementation of ProviderAppChecker for testing.
type MockProviderAppChecker struct {
	providerApps map[string]bool // provider name -> configured
}

// NewMockProviderAppChecker creates a new mock provider app checker with the given providers.
func NewMockProviderAppChecker(providerApps map[string]bool) ProviderAppChecker {
	return &MockProviderAppChecker{
		providerApps: providerApps,
	}
}

// CheckProviderApp checks if a provider app is configured in the mock.
func (m *MockProviderAppChecker) CheckProviderApp(providerName string) error {
	if m.providerApps == nil {
		return errors.New("mock not configured")
	}

	configured, ok := m.providerApps[providerName]
	if !ok || !configured {
		return ErrProviderAppNotFound
	}

	return nil
}

// RateLimitInfo contains rate limit recommendations for a provider.
type RateLimitInfo struct {
	// MinScheduleInterval is the minimum recommended schedule interval for this provider
	MinScheduleInterval time.Duration
	// RecommendedInterval is the recommended "safe" schedule interval
	RecommendedInterval time.Duration
	// Description provides context about the rate limit (e.g., "Salesforce API limits: 15,000 calls/24h")
	Description string
}

// RateLimitChecker defines an interface for checking provider-specific rate limit recommendations.
// This abstraction allows injecting provider-specific or account-specific rate limit information.
//
// Client-side implementations might:
//   - Return common-sense defaults based on provider documentation
//   - Return account-specific limits if available via API
//
// Server-side implementations might:
//   - Query account tier information
//   - Return dynamic limits based on usage patterns
//   - Provide custom limits per customer agreement
type RateLimitChecker interface {
	// GetRateLimitInfo returns rate limit recommendations for a provider.
	// Returns RateLimitInfo with recommended intervals.
	// Returns error if provider is not recognized or info cannot be retrieved.
	GetRateLimitInfo(providerName string) (*RateLimitInfo, error)
}

// MockRateLimitChecker is a mock implementation of RateLimitChecker for testing.
type MockRateLimitChecker struct {
	rateLimits map[string]RateLimitInfo // provider name -> rate limit info
}

// NewMockRateLimitChecker creates a new mock rate limit checker with the given rate limits.
func NewMockRateLimitChecker(rateLimits map[string]RateLimitInfo) RateLimitChecker {
	return &MockRateLimitChecker{
		rateLimits: rateLimits,
	}
}

// GetRateLimitInfo returns rate limit info from the mock.
func (m *MockRateLimitChecker) GetRateLimitInfo(providerName string) (*RateLimitInfo, error) {
	if m.rateLimits == nil {
		return nil, errors.New("mock not configured")
	}

	info, ok := m.rateLimits[providerName]
	if !ok {
		return nil, errors.New("provider rate limit info not found")
	}

	return &info, nil
}
