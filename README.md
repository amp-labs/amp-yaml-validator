# amp-yaml-validator

A Go library and CLI tool for validating Ampersand integration manifest files (`amp.yaml`).

## Overview

The `amp-yaml-validator` validates `amp.yaml` files against spec version 1.0.0, catching configuration errors before deployment. It provides detailed error messages with precise line numbers to help Ampersand developers quickly identify and fix issues.

### What it validates

- **Schema validation**: Structural correctness, required fields, data types
- **Orphan/unknown key detection**: Warns about keys that aren't part of the schema (typically typos) which would otherwise be silently ignored
- **Business rules**: Schedule frequency, backfill constraints, field mappings, subscribe constraints
- **Provider capabilities**: Verifies provider supports requested actions (read/write/subscribe/proxy)
- **Provider-specific limits**: Salesforce CDC limits, Google Calendar backfill constraints, Snowflake requirements
- **Async error prevention**: Large backfills, destination references, rate limit risks

## Key Features

- **70 validation rules** covering all aspects of amp.yaml configuration
- **Precise error reporting** with line numbers and YAML paths (e.g., `$.integrations[0].read.objects[1].schedule`)
- **Provider catalog integration** validates against actual provider capabilities from connectors library
- **Flexible destination checking** with injectable interface for client-side or server-side validation
- **Async error prevention** warns about configuration issues that fail at runtime
- **Severity levels**: Errors block deployment, warnings indicate best practice violations
- **Simple API**: `ValidateFile()` or `ValidateBytes()` with functional options
- **CLI tool**: Validate files from command line with `--strict`, `--skip-provider`, `--skip-async` flags

## Installation

### CLI Tool

```bash
# Clone the repository
git clone https://github.com/amp-labs/amp-yaml-validator.git
cd amp-yaml-validator

# Build
make build

# Run
./amp-yaml-validator path/to/amp.yaml
```

### Go Library

```bash
go get github.com/amp-labs/amp-yaml-validator
```

## Usage

### CLI

```bash
# Validate a file
./amp-yaml-validator amp.yaml

# Strict mode (warnings treated as errors)
./amp-yaml-validator --strict amp.yaml

# Skip provider validation
./amp-yaml-validator --skip-provider amp.yaml

# Skip async validation
./amp-yaml-validator --skip-async amp.yaml
```

### Go Library - Basic Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/amp-labs/amp-yaml-validator/validator"
)

func main() {
    // Create validator
    v := validator.NewValidator()

    // Validate a file
    result, err := v.ValidateFile("amp.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // Check results
    if !result.Valid {
        fmt.Println("Validation failed:")
        for _, issue := range result.Errors {
            fmt.Printf("  [%s] Line %d: %s\n", issue.Rule, issue.Line, issue.Message)
            if issue.Suggestion != "" {
                fmt.Printf("    Suggestion: %s\n", issue.Suggestion)
            }
        }
    }

    // Show warnings
    if len(result.Warnings) > 0 {
        fmt.Println("\nWarnings:")
        for _, issue := range result.Warnings {
            fmt.Printf("  [%s] Line %d: %s\n", issue.Rule, issue.Line, issue.Message)
        }
    }
}
```

### Advanced Usage - Destination Checking

The validator supports optional destination checking for both client-side and server-side use cases:

```go
import (
    "github.com/amp-labs/amp-yaml-validator/validator"
    "github.com/amp-labs/amp-yaml-validator/checker"
)

// Server-side: Check destinations in database
type ServerDestinationChecker struct {
    db *sql.DB
}

func (c *ServerDestinationChecker) CheckDestination(name string) error {
    var exists bool
    err := c.db.QueryRow(
        "SELECT EXISTS(SELECT 1 FROM destinations WHERE name = $1 AND active = true)",
        name,
    ).Scan(&exists)
    if err != nil {
        return fmt.Errorf("database error: %w", err)
    }
    if !exists {
        return checker.ErrDestinationNotFound
    }
    return nil
}

// Client-side: Check destinations via API
type ClientDestinationChecker struct {
    apiClient *ampersand.Client
    projectID string
}

func (c *ClientDestinationChecker) CheckDestination(name string) error {
    destinations, err := c.apiClient.ListDestinations(c.projectID)
    if err != nil {
        return fmt.Errorf("API error: %w", err)
    }

    for _, dest := range destinations {
        if dest.Name == name && dest.Status == "active" {
            return nil
        }
    }

    return checker.ErrDestinationNotFound
}

// Use with validator
v := validator.NewValidator(
    validator.WithDestinationChecker(&ServerDestinationChecker{db: db}),
)
```

When a destination checker is provided, destination references produce **errors** if destinations don't exist. Without a checker, they produce **warnings**.

### Advanced Usage - Other Options

```go
// Strict mode - treat warnings as errors
v := validator.NewValidator(
    validator.WithStrictMode(true),
)

// Skip provider validation (useful for testing without connectors catalog)
v := validator.NewValidator(
    validator.WithSkipProviderValidation(),
)

// Skip async validation
v := validator.NewValidator(
    validator.WithSkipAsyncValidation(),
)

// Custom catalog provider (for testing with mock data)
mockCatalog := catalog.NewMockCatalogProvider(customProviders)
v := validator.NewValidator(
    validator.WithCatalogProvider(mockCatalog),
)

// Combine multiple options
v := validator.NewValidator(
    validator.WithStrictMode(true),
    validator.WithDestinationChecker(destChecker),
)
```

## Example Output

```
Validation failed with 2 errors and 1 warning:

Errors:
  [schedule-minimum-interval] Line 15, Column 7: Schedule runs every 5 minutes, which is below the minimum frequency of 10 minutes
    Path: $.integrations[0].read.objects[0].schedule
    Suggestion: Use a schedule with at least 10 minutes between runs, e.g., "*/10 * * * *"

  [provider-capability-subscribe] Line 3, Column 3: Provider "hubspot" does not support subscribe capability
    Path: $.integrations[0].provider
    Suggestion: Remove the subscribe section or choose a provider that supports webhooks/CDC

Warnings:
  [large-backfill-risk] Line 18, Column 9: Object "Contact" has a large backfill period (365 days), which may cause timeouts with large datasets
    Path: $.integrations[0].read.objects[0].backfill.defaultPeriod.days
    Suggestion: Consider reducing backfill to 180 days or less for initial sync
```

## Validation Rules

### Universal Rules (Apply to all integrations)

- **Spec version**: Must be "1.0.0"
- **Schedule syntax**: Valid 5-field cron expression
- **Schedule frequency**: Minimum 10 minutes between runs
- **Delivery mode**:
  - `auto`: Cannot specify `pageSize`
  - `onRequest`: Must specify `pageSize` (50-500)
- **Backfill**: Must specify either `days` or `fullHistory`, not both
- **Field mappings**: No duplicate `mapToName` within an object
- **Always-enabled objects**: Must have `requiredFields` and `schedule`, cannot use `mapToName` in required fields
- **Subscribe actions**:
  - Must have corresponding `read` action
  - `inheritFieldsAndMapping` must be `true`
  - Update events require either `requiredWatchFields` or `watchFieldsAuto` (not both)
  - At least one event type must be enabled (`createEvent`, `updateEvent`, `deleteEvent`, or `associationChangeEvent`)
  - `requiredWatchFields` cannot contain nested paths (dots or brackets)
- **Duplicate object detection**: Same `objectName` cannot appear twice within the same action (read/write/subscribe)

### Provider-Specific Rules

The validator integrates with the connectors catalog to check provider capabilities:

- **Provider exists**: Provider must be in the catalog
- **Capability support**: Provider must support requested actions (read/write/subscribe/proxy)
- **Module support**: If module specified, provider must support that module and its capabilities
- **Salesforce limits**: Maximum 5 subscribe objects (CDC platform limit)
- **Google Calendar constraints**:
  - `events` object cannot use `fullHistory` backfill
  - `events` object backfill limited to maximum 28 days
- **Snowflake constraints**: Must use `fullHistory` backfill (days-based not supported)
- **Graceful degradation**: If catalog unavailable, issues warnings instead of errors

### Async Error Prevention (Warnings)

These rules warn about potential runtime issues:

- **Destination references**:
  - With checker: Produces **errors** for non-existent destinations
  - Without checker: Produces **warnings** reminding users to verify destinations exist
- **Object existence**: Warns about objects not found in provider catalog (when catalog has schemas)
- **Large backfills**: Warns about backfills >180 days or `fullHistory` (timeout risk)
- **Frequent schedules**: Warns about schedules ≤15 minutes (rate limit risk)

For complete rule documentation with examples, see [VALIDATION_RULES.md](./VALIDATION_RULES.md).

## Development

### Project Structure

```
amp-yaml-validator/
├── catalog/           # Provider catalog integration
├── checker/           # Validation checker interfaces (e.g., DestinationChecker)
├── cmd/              # CLI tool
├── openapi/          # Generated OpenAPI types
├── parser/           # YAML parsing with position tracking
├── types/            # Shared type definitions
├── validator/        # Core validation logic
│   ├── validator.go      # Main orchestrator
│   ├── context.go        # ValidationContext
│   ├── async.go          # Async error prevention
│   ├── provider.go       # Provider-specific rules
│   ├── schedule.go       # Schedule validation
│   ├── backfill.go       # Backfill validation
│   ├── subscribe.go      # Subscribe action validation
│   └── ...
└── testdata/         # Test fixtures
```

### Building

```bash
# Build CLI
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run linters
make lint

# Auto-fix issues
make fix
```

### Key Design Features

- **Hybrid YAML parsing**: Uses `yaml.v3` for position tracking and `sigs.k8s.io/yaml` for struct unmarshaling (handles JSON tags, matches server)
- **ValidationContext pattern**: Passes manifest, positions, catalog, and destination checker to all validators
- **Functional options**: Flexible configuration via `WithStrictMode()`, `WithDestinationChecker()`, etc.
- **Modular validators**: Universal, provider-specific, and async validators are independent
- **Graceful degradation**: Continues validation if provider catalog is unavailable

For detailed architecture documentation, see [ARCHITECTURE.md](./ARCHITECTURE.md) and [CLAUDE.md](./CLAUDE.md).

## Testing

```bash
# Run all tests
go test -v ./...

# Run specific package tests
go test -v ./validator

# Run specific test
go test -v -run TestValidateSchedule ./validator

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

The library has comprehensive test coverage with table-driven, parallel tests for all validation rules.

## Contributing

For Ampersand developers adding new validation rules:

1. **Implement the validator function** in the appropriate file (e.g., `validator/schedule.go`)
2. **Add to orchestrator** in `validator/validator.go`
3. **Add tests** in `validator/*_test.go`
4. **Update docs**: `VALIDATION_RULES.md`, `CLAUDE.md`, and `README.md`

See [CLAUDE.md](./CLAUDE.md) for detailed development patterns and examples.

## Documentation

- **[VALIDATION_RULES.md](./VALIDATION_RULES.md)**: Complete specification of all 70 validation rules with examples
- **[ARCHITECTURE.md](./ARCHITECTURE.md)**: Detailed design decisions and implementation phases
- **[CLAUDE.md](./CLAUDE.md)**: Developer guide for working with the codebase (for AI assistants and developers)

## Project Status

**Current Phase**: Phase 5 Complete (Semantic Validation Enhancement)

- ✅ Phase 1: Documentation and architecture design
- ✅ Phase 2: Universal validation rules implementation
- ✅ Phase 3: Provider-specific validation with catalog integration
- ✅ Phase 4: Async error prevention validation with destination checking
- ✅ Phase 5: Semantic validation enhancement
  - ✅ Duplicate object detection (read/write/subscribe)
  - ✅ Subscribe event type validation (minimum one event required)
  - ✅ Nested watch fields validation (no dots or brackets)
  - ✅ Google Calendar backfill constraints (no fullHistory, max 28 days for events)
  - ✅ Snowflake backfill requirements (fullHistory only)
  - ✅ JSONPath validation utilities (nested field path detection)

## License

Copyright © 2026 Ampersand Technologies, Inc.
