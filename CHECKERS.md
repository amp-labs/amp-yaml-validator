# Validator Checkers

This document describes the checker interfaces available for extending the validator with runtime-specific validation logic.

## Overview

The validator supports three optional checker interfaces that allow client-side and server-side implementations to inject custom validation logic for things that cannot be determined statically from the YAML manifest alone:

1. **DestinationChecker** - Validates destination existence
2. **ProviderAppChecker** - Validates provider app/credential configuration
3. **RateLimitChecker** - Provides provider-specific rate limit information

All checkers are **optional**. When a checker is not provided, the validator gracefully degrades (typically issuing warnings instead of errors).

## Checker Interfaces

### DestinationChecker

Validates that destinations referenced in the manifest exist and are accessible.

```go
type DestinationChecker interface {
    // CheckDestination verifies if a destination exists and is accessible.
    // Returns nil if the destination exists and is valid.
    // Returns ErrDestinationNotFound if the destination doesn't exist.
    CheckDestination(destinationName string) error
}
```

**Usage:**
- **Without checker**: Issues warnings reminding users to verify destinations exist
- **With checker**: Issues errors if destinations don't exist (caught at validation time)

**Example:**
```go
type MyDestinationChecker struct {
    db *sql.DB
}

func (c *MyDestinationChecker) CheckDestination(destinationName string) error {
    var exists bool
    err := c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM destinations WHERE name = $1)", destinationName).Scan(&exists)
    if err != nil {
        return err
    }
    if !exists {
        return checker.ErrDestinationNotFound
    }
    return nil
}

// Use it
validator := validator.NewValidator(
    validator.WithDestinationChecker(&MyDestinationChecker{db: db}),
)
```

### ProviderAppChecker

Validates that provider apps (OAuth apps, API credentials) are configured for a given provider.

```go
type ProviderAppChecker interface {
    // CheckProviderApp verifies if a provider has valid credentials/OAuth apps configured.
    // Returns nil if the provider app exists and is properly configured.
    // Returns ErrProviderAppNotFound if the provider app doesn't exist.
    CheckProviderApp(providerName string) error
}
```

**Usage:**
- **Without checker**: No validation of provider app configuration
- **With checker**: Can issue warnings or errors if provider apps are not configured

**Example:**
```go
type MyProviderAppChecker struct {
    projectID string
    apiClient *APIClient
}

func (c *MyProviderAppChecker) CheckProviderApp(providerName string) error {
    // Check if OAuth app or credentials are configured for this provider
    configured, err := c.apiClient.IsProviderConfigured(c.projectID, providerName)
    if err != nil {
        return fmt.Errorf("failed to check provider config: %w", err)
    }
    if !configured {
        return checker.ErrProviderAppNotFound
    }
    return nil
}

// Use it
validator := validator.NewValidator(
    validator.WithProviderAppChecker(&MyProviderAppChecker{
        projectID: "proj-123",
        apiClient: client,
    }),
)
```

### RateLimitChecker

Provides provider-specific or account-specific rate limit recommendations for schedule validation.

```go
type RateLimitInfo struct {
    MinScheduleInterval time.Duration // Minimum recommended interval
    RecommendedInterval time.Duration // Recommended "safe" interval
    Description         string        // Context about the limit
}

type RateLimitChecker interface {
    // GetRateLimitInfo returns rate limit recommendations for a provider.
    GetRateLimitInfo(providerName string) (*RateLimitInfo, error)
}
```

**Usage:**
- **Without checker**: Uses universal default thresholds (10 min minimum, 15 min warning)
- **With checker**: Can provide provider-specific or account-tier-specific limits

**Example:**
```go
type MyRateLimitChecker struct {
    accountTier string
}

func (c *MyRateLimitChecker) GetRateLimitInfo(providerName string) (*checker.RateLimitInfo, error) {
    // Return account-tier-specific limits
    if providerName == "salesforce" {
        if c.accountTier == "enterprise" {
            return &checker.RateLimitInfo{
                MinScheduleInterval: 5 * time.Minute,
                RecommendedInterval: 15 * time.Minute,
                Description:         "Salesforce Enterprise: 1M API calls/24h",
            }, nil
        }
        return &checker.RateLimitInfo{
            MinScheduleInterval: 15 * time.Minute,
            RecommendedInterval: 30 * time.Minute,
            Description:         "Salesforce Standard: 15K API calls/24h",
        }, nil
    }

    // Return default for unknown providers
    return &checker.RateLimitInfo{
        MinScheduleInterval: 10 * time.Minute,
        RecommendedInterval: 15 * time.Minute,
        Description:         "Default rate limits",
    }, nil
}

// Use it
validator := validator.NewValidator(
    validator.WithRateLimitChecker(&MyRateLimitChecker{
        accountTier: "enterprise",
    }),
)
```

## Complete Example

```go
package main

import (
    "fmt"
    "time"

    validator "github.com/amp-labs/amp-yaml-validator"
    "github.com/amp-labs/amp-yaml-validator/checker"
)

func main() {
    // Create custom checkers
    destChecker := &MyDestinationChecker{db: db}
    providerChecker := &MyProviderAppChecker{projectID: "proj-123"}
    rateLimitChecker := &MyRateLimitChecker{accountTier: "enterprise"}

    // Create validator with all checkers
    v := validator.NewValidator(
        validator.WithDestinationChecker(destChecker),
        validator.WithProviderAppChecker(providerChecker),
        validator.WithRateLimitChecker(rateLimitChecker),
    )

    // Validate a manifest
    result, err := v.ValidateFile("amp.yaml")
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
        return
    }

    // Check results
    if !result.Valid {
        fmt.Printf("Validation failed with %d errors\n", len(result.Errors))
        for _, issue := range result.Errors {
            fmt.Printf("  - %s: %s\n", issue.Rule, issue.Message)
        }
    }
}
```

## Mock Implementations

The `checker` package provides mock implementations for testing:

```go
// Mock destination checker
destChecker := checker.NewMockDestinationChecker(map[string]bool{
    "webhook":  true,
    "postgres": true,
})

// Mock provider app checker
providerChecker := checker.NewMockProviderAppChecker(map[string]bool{
    "salesforce": true,
    "hubspot":    true,
})

// Mock rate limit checker
rateLimitChecker := checker.NewMockRateLimitChecker(map[string]checker.RateLimitInfo{
    "salesforce": {
        MinScheduleInterval: 15 * time.Minute,
        RecommendedInterval: 30 * time.Minute,
        Description:         "Salesforce test limits",
    },
})

validator := validator.NewValidator(
    validator.WithDestinationChecker(destChecker),
    validator.WithProviderAppChecker(providerChecker),
    validator.WithRateLimitChecker(rateLimitChecker),
)
```

## Design Philosophy

### Why Checkers are Optional

The manifest represents a **formula or recipe**, not a specific deployed integration. Many validation concerns depend on runtime context:

- **Destinations**: Only exist after project setup
- **Provider apps**: Only configured when project connects to provider
- **Rate limits**: Vary by account tier, usage patterns, and agreements

By making checkers optional, the validator can:
1. Work in **build-time scenarios** (e.g., CI/CD, manifest linting)
2. Work in **runtime scenarios** (e.g., server-side validation before deployment)
3. Gracefully degrade with warnings when context is unavailable

### Client-Side vs Server-Side

**Client-Side (CLI, CI/CD)**
- Checkers typically not provided
- Gets warnings to remind about runtime requirements
- Fast, no network dependencies

**Server-Side (API, Deployment Pipeline)**
- Checkers provided with database/API access
- Gets errors for invalid configurations
- Prevents deployment of bad manifests

## Implementation Notes

### Error Handling

Checkers should return:
- `nil` - Resource is valid
- Specific error types (e.g., `ErrDestinationNotFound`, `ErrProviderAppNotFound`) - Known validation failure
- Other errors - Unexpected failures (network issues, auth failures)

The validator handles errors appropriately based on type.

### Performance

Checkers may be called multiple times during validation (once per resource reference). Implementations should:
- Cache results when possible
- Handle concurrent calls if using shared state
- Fail fast on errors

### Testing

Use mock implementations for testing:
```go
func TestMyValidation(t *testing.T) {
    destChecker := checker.NewMockDestinationChecker(map[string]bool{
        "webhook": true,
    })

    validator := validator.NewValidator(
        validator.WithDestinationChecker(destChecker),
    )

    // Test validation logic
}
```

## See Also

- [README.md](README.md) - Main documentation
- [ARCHITECTURE.md](ARCHITECTURE.md) - Design decisions
- [examples/custom_checkers_example.go](examples/custom_checkers_example.go) - Complete example
