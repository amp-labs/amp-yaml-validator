package validator

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/amp-labs/amp-yaml-validator/checker"
	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/parser"
)

var errCheckerBackendDown = errors.New("checker backend unavailable")

// stubDestinationChecker returns a fixed error for every destination check.
type stubDestinationChecker struct {
	err error
}

func (s *stubDestinationChecker) CheckDestination(_ context.Context, _ string) error {
	return s.err
}

// stubProviderAppChecker returns a fixed error for every provider app check.
type stubProviderAppChecker struct {
	err error
}

func (s *stubProviderAppChecker) CheckProviderApp(_ context.Context, _ string) error {
	return s.err
}

func manifestWithDestination() *openapi.Manifest {
	objects := []openapi.IntegrationObject{
		{
			ObjectName:  "account",
			Destination: "myDestination",
			Schedule:    "*/10 * * * *",
		},
	}

	return &openapi.Manifest{
		SpecVersion: "1.0.0",
		Integrations: []openapi.Integration{
			{
				Provider: "salesforce",
				Read: &openapi.IntegrationRead{
					Objects: &objects,
				},
			},
		},
	}
}

func TestDestinationCheckerNotSupportedIgnored(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantIssues bool
	}{
		{
			name:       "ErrNotSupported is silently ignored",
			err:        checker.ErrNotSupported,
			wantIssues: false,
		},
		{
			name:       "wrapped ErrNotSupported is silently ignored",
			err:        fmt.Errorf("client-side checker: %w", checker.ErrNotSupported),
			wantIssues: false,
		},
		{
			name:       "other errors still produce issues",
			err:        errCheckerBackendDown,
			wantIssues: true,
		},
	}

	//nolint:varnamelen // tt is idiomatic in table-driven tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			valCtx := NewValidationContext(
				manifestWithDestination(), parser.NewPositionMap(), parser.NewDirectiveMap(),
				nil, &stubDestinationChecker{err: tt.err}, nil, nil,
			)

			// Both destination validation paths consult the checker
			validateDestinationsExist(t.Context(), valCtx)
			validateDestinationReferences(t.Context(), valCtx)

			issueCount := len(valCtx.GetErrors()) + len(valCtx.GetWarnings())

			if tt.wantIssues && issueCount == 0 {
				t.Error("expected issues from checker error, got none")
			}

			if !tt.wantIssues && issueCount > 0 {
				t.Errorf("expected no issues, got %d", issueCount)

				for _, issue := range valCtx.GetErrors() {
					t.Logf("  Error: %s (rule: %s)", issue.Message, issue.Rule)
				}

				for _, issue := range valCtx.GetWarnings() {
					t.Logf("  Warning: %s (rule: %s)", issue.Message, issue.Rule)
				}
			}
		})
	}
}

func TestProviderAppCheckerNotSupportedIgnored(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantIssues bool
	}{
		{
			name:       "ErrNotSupported is silently ignored",
			err:        checker.ErrNotSupported,
			wantIssues: false,
		},
		{
			name:       "wrapped ErrNotSupported is silently ignored",
			err:        fmt.Errorf("client-side checker: %w", checker.ErrNotSupported),
			wantIssues: false,
		},
		{
			name:       "provider app not found still produces warning",
			err:        checker.ErrProviderAppNotFound,
			wantIssues: true,
		},
		{
			name:       "other errors still produce issues",
			err:        errCheckerBackendDown,
			wantIssues: true,
		},
	}

	//nolint:varnamelen // tt is idiomatic in table-driven tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			valCtx := NewValidationContext(
				&openapi.Manifest{}, parser.NewPositionMap(), parser.NewDirectiveMap(),
				nil, nil, &stubProviderAppChecker{err: tt.err}, nil,
			)

			integration := openapi.Integration{Provider: "salesforce"}

			validateProviderAppConfiguration(t.Context(), valCtx, integration, "$.integrations[0]")

			issueCount := len(valCtx.GetErrors()) + len(valCtx.GetWarnings())

			if tt.wantIssues && issueCount == 0 {
				t.Error("expected issues from checker error, got none")
			}

			if !tt.wantIssues && issueCount > 0 {
				t.Errorf("expected no issues, got %d", issueCount)

				for _, issue := range valCtx.GetErrors() {
					t.Logf("  Error: %s (rule: %s)", issue.Message, issue.Rule)
				}

				for _, issue := range valCtx.GetWarnings() {
					t.Logf("  Warning: %s (rule: %s)", issue.Message, issue.Rule)
				}
			}
		})
	}
}
