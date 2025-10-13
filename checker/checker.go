package checker

import "errors"

// ErrDestinationNotFound indicates that a destination does not exist or is not accessible
var ErrDestinationNotFound = errors.New("destination not found")

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
