package main

import (
	"context"
	"fmt"
	"time"

	"github.com/amp-labs/amp-yaml-validator/checker"
	"github.com/amp-labs/amp-yaml-validator/validator"
)

// Example implementation of ProviderAppChecker that checks a database or API
type DatabaseProviderAppChecker struct {
	// In a real implementation, this would have a database connection or API client
}

func (c *DatabaseProviderAppChecker) CheckProviderApp(ctx context.Context, providerName string) error {
	// In a real implementation, this would query the database or API
	// to check if the provider has valid credentials configured
	// For this example, we'll just simulate checking
	configuredProviders := map[string]bool{
		"salesforce": true,
		"hubspot":    true,
		// "zendesk" is NOT configured
	}

	if !configuredProviders[providerName] {
		return checker.ErrProviderAppNotFound
	}

	return nil
}

// Example implementation of RateLimitChecker with provider-specific limits
type CustomRateLimitChecker struct {
	// In a real implementation, this might query account tier information
}

func (c *CustomRateLimitChecker) GetRateLimitInfo(ctx context.Context, providerName string) (*checker.RateLimitInfo, error) {
	// Provide provider-specific rate limit recommendations
	limits := map[string]checker.RateLimitInfo{
		"salesforce": {
			MinScheduleInterval: 15 * time.Minute,
			RecommendedInterval: 30 * time.Minute,
			Description:         "Salesforce API limits: 15,000 calls/24h",
		},
		"hubspot": {
			MinScheduleInterval: 10 * time.Minute,
			RecommendedInterval: 20 * time.Minute,
			Description:         "HubSpot API limits: 10,000 calls/day (free tier)",
		},
	}

	info, ok := limits[providerName]
	if !ok {
		// Return default limits if provider-specific limits not found
		return &checker.RateLimitInfo{
			MinScheduleInterval: 10 * time.Minute,
			RecommendedInterval: 15 * time.Minute,
			Description:         "Default API rate limits",
		}, nil
	}

	return &info, nil
}

func main() {
	yaml := []byte(`
specVersion: 1.0.0
integrations:
  - name: myIntegration
    provider: zendesk
    read:
      objects:
        - objectName: tickets
          destination: webhook
          schedule: "*/5 * * * *"
`)

	// Create validator with custom checkers
	v := validator.NewValidator(
		validator.WithProviderAppChecker(&DatabaseProviderAppChecker{}),
		validator.WithRateLimitChecker(&CustomRateLimitChecker{}),
	)

	// Validate
	result, err := v.ValidateBytes(context.Background(), yaml)
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	// Print results
	fmt.Println("Validation Result:")
	fmt.Printf("Valid: %v\n", result.Valid)
	fmt.Printf("Errors: %d\n", len(result.Errors))
	fmt.Printf("Warnings: %d\n", len(result.Warnings))

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")

		for i, issue := range result.Errors {
			fmt.Printf("  %d. [%s] %s\n", i+1, issue.Rule, issue.Message)

			if issue.Path != "" {
				fmt.Printf("     Path: %s\n", issue.Path)
			}
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println("\nWarnings:")

		for i, issue := range result.Warnings {
			fmt.Printf("  %d. [%s] %s\n", i+1, issue.Rule, issue.Message)

			if issue.Path != "" {
				fmt.Printf("     Path: %s\n", issue.Path)
			}
		}
	}
}
