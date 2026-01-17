# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`amp-yaml-validator` is a Go library that validates Ampersand integration manifest files (`amp.yaml`) against spec version 1.0.0 schema and business rules. The library provides detailed error and warning messages with precise line numbers to help developers identify and fix configuration issues.

The validator performs:
- **Schema validation**: Structural correctness and required fields
- **Business logic validation**: Semantic rules (schedule frequency, backfill constraints, field mappings)
- **Provider-specific validation**: Provider capabilities and limits (Salesforce subscribe limits, provider action support)
- **Async error prevention**: Configuration issues that would fail at runtime in Temporal workflows or messenger services

## Build and Development Commands

### Building
```bash
# Build the CLI tool
make build

# Output: ./amp-yaml-validator

# Build for all platforms
make build-all
# Outputs: amp-yaml-validator-linux-amd64, amp-yaml-validator-darwin-amd64,
#          amp-yaml-validator-darwin-arm64, amp-yaml-validator-windows-amd64.exe

# Run the validator on a file
./amp-yaml-validator path/to/amp.yaml
```

### Testing
```bash
# Run all tests with verbose output
make test

# Run all tests (go test directly)
go test -v ./...

# Run a specific test
go test -v -run TestValidateSchedule ./validator

# Run tests in parallel
go test -parallel 8 ./...

# Run tests for a specific package
go test -v ./validator
go test -v ./parser
```

### Linting and Formatting
```bash
# Run linters
make lint

# Auto-fix formatting issues
make fix

# Fix markdown files
make fix-markdown
```

## Code Architecture

The library is organized into focused components:

### Core Packages

#### `parser/` - YAML parsing with position tracking
- **`parser.go`**: Two-pass parsing using `gopkg.in/yaml.v3` (position tracking) and `sigs.k8s.io/yaml` (struct unmarshaling with JSON tags)
- **`position.go`**: Position tracking types (`Position`, `PositionMap`)
- **`path.go`**: YAML path construction utilities for error reporting (e.g., `$.integrations[0].read.objects[1].schedule`)

#### `validator/` - Core validation logic
- **`validator.go`**: Main orchestrator that runs all validation rules, supports functional options for configuration
- **`context.go`**: `ValidationContext` tracks manifest, positions, catalog provider, destination checker, and accumulates issues
- **`spec_version.go`**: Validates specVersion must be "1.0.0"
- **`schedule.go`**: Schedule validation (cron syntax, minimum 10-minute frequency)
- **`delivery.go`**: Delivery mode validation (auto vs onRequest, pageSize constraints)
- **`backfill.go`**: Backfill validation (days XOR fullHistory constraint)
- **`field_mapping.go`**: Field mapping uniqueness validation
- **`always_enabled.go`**: Always-enabled object constraints
- **`read.go`**: Read action validation (objects, fields, schedules)
- **`write.go`**: Write action validation
- **`subscribe.go`**: Subscribe action validation (requires read, inheritFieldsAndMapping, watch fields)
- **`subscribe_events.go`**: Subscribe event type validation (at least one event enabled, nested watch fields check)
- **`duplicate.go`**: Duplicate object detection within same action (read/write/subscribe)
- **`jsonpath.go`**: JSONPath validation and nested field detection (placeholder for amp-common/jsonpath integration)
- **`provider_google_calendar.go`**: Google Calendar-specific validation (events backfill constraints)
- **`provider_snowflake.go`**: Snowflake-specific validation (fullHistory requirement)
- **`integration.go`**: Integration structure validation
- **`provider.go`**: Provider-specific validation (capability checks, Salesforce limits, module validation)
- **`async.go`**: Async error prevention validation (destination references, backfill risks, schedule frequency risks)

#### `catalog/` - Provider catalog integration
- **`catalog.go`**: Interface to `github.com/amp-labs/connectors/providers` for provider capability checks
- Provides `CatalogProvider` interface with default and mock implementations
- Used by provider-specific validators to check if provider supports read/write/subscribe/proxy

#### `checker/` - Validation checkers
- **`checker.go`**: Defines interfaces for external validation checks
- `DestinationChecker` interface: Allows client-side or server-side destination validation
- Mock implementation for testing
- Used by async validators to check if destinations exist (when provided)

#### `types/` - Type definitions
- **`result.go`**: `ValidationResult` and `ValidationIssue` structures
- **`errors.go`**: Error message constants

#### `openapi/` - Generated OpenAPI types
- **`manifest.gen.go`**: Generated types from OpenAPI spec for amp.yaml structure
- **`config.gen.go`**: Generated config types

#### `cmd/amp-yaml-validator/` - CLI tool
- **`main.go`**: Command-line interface using cobra, supports `--strict`, `--skip-provider`, `--skip-async` flags

### Async Error Prevention (validator/async.go)

Validates configuration issues that would cause runtime failures in asynchronous services:

- **`validateAsyncRisks()`**: Main orchestrator for async validation
- **`validateDestinationReferences()`**: Checks destination existence using optional `DestinationChecker`
  - If checker provided: produces **errors** for non-existent destinations
  - If no checker: produces **warnings** reminding user to verify destinations exist
- **`validateObjectExistence()`**: Warns about objects not found in provider catalog (when catalog has object schemas)
- **`validateBackfillRisks()`**: Warns about large backfills that may timeout (>180 days or fullHistory)
- **`validateScheduleFrequencyRisks()`**: Warns about very frequent schedules (≤15 minutes) that may hit rate limits

Most async validations produce warnings by default. Destination checking can produce errors when a `DestinationChecker` is provided.

### Key Design Patterns

#### Two-Pass YAML Parsing
The parser uses a hybrid approach to preserve position information while correctly parsing OpenAPI structs:
1. **First pass**: Parse into `yaml.v3.Node` using `gopkg.in/yaml.v3` to extract line/column positions → builds `PositionMap`
2. **Second pass**: Parse into typed structs (`openapi.Manifest`) using `sigs.k8s.io/yaml` which handles JSON struct tags

This hybrid approach is necessary because:
- `yaml.v3` provides position tracking via `Node.Line` and `Node.Column`
- `sigs.k8s.io/yaml` correctly unmarshals into OpenAPI structs with JSON tags (matches server behavior)

This enables accurate error reporting with line numbers while correctly parsing the manifest structure.

#### ValidationContext Pattern
All validators receive a `ValidationContext` that provides:
- `Manifest`: Access to parsed manifest
- `PositionMap`: Lookup line/column for YAML paths via `GetPosition(path string)`
- `CatalogProvider`: Access provider catalog for capability checks via `GetProviderInfo(providerName)`
- `DestinationChecker`: Optional checker for validating destination existence (nil by default)
- `AddError()` / `AddWarning()`: Accumulate validation issues with suggestions
- `GetErrors()` / `GetWarnings()`: Retrieve accumulated issues by severity

#### Validation Rules Organization
- **Universal rules**: Apply to all integrations (schedule frequency, required fields, etc.)
- **Provider-specific rules**: Check provider capabilities via catalog (read/write/subscribe support, Salesforce limits)
- **Async rules**: Warn about runtime issues (large backfills, destination references, schedule frequency)

#### Functional Options Pattern
The validator uses functional options for configuration:
```go
validator := NewValidator(
    WithStrictMode(true),              // Treat warnings as errors
    WithSkipProviderValidation(),      // Skip provider-specific checks
    WithSkipAsyncValidation(),         // Skip async error prevention checks
    WithCatalogProvider(mockCatalog),  // Inject custom catalog (for testing)
    WithDestinationChecker(checker),   // Inject destination checker (client/server)
)
```

This allows flexible configuration for different use cases (client-side vs server-side, testing, etc.).

## Important Implementation Details

### Schedule Validation
- Uses `github.com/adhocore/gronx` for cron validation
- Production minimum: 10 minutes between runs
- Rejects `*` in minute field (would run every minute)
- Validates using tick calculation (nextTick - prevTick >= minFrequency)

### Provider Validation
- Accesses provider catalog via `catalog.GetProvider(providerName)`
- Checks `ProviderInfo.Support.Read/Write/Subscribe/Proxy` flags
- Module overrides: If integration specifies a module, use `ModuleInfo.Support` instead
- **Graceful degradation**: If catalog unavailable, issues warning and continues with universal validation

### Salesforce-Specific Rules
- Maximum 5 subscribe objects (Salesforce CDC platform limit)
- Enforced in `validateProviderSpecific()` when provider is "salesforce"

### Field Mapping Validation
- Ensures no duplicate `mapToName` values within an object's field configuration
- Required mappings must be present in selected mappings
- Always-enabled objects cannot use `mapToName` in `requiredFields`

### Subscribe Action Constraints
- Must have corresponding `read` action
- `inheritFieldsAndMapping` must be `true` (v1.0.0 constraint)
- `updateEvent` requires either `requiredWatchFields` or `watchFieldsAuto` (not both)
- `updateEvent.enabled` must be `"always"` if specified

### Backfill Validation
- Must specify EITHER `defaultPeriod.days` OR `defaultPeriod.fullHistory`, not both
- Uses `github.com/go-playground/validator/v10` for struct tag validation

### Destination Validation
The validator supports optional destination checking through the `DestinationChecker` interface:

```go
// Implement DestinationChecker for your environment
type MyDestinationChecker struct {
    // client-side: might make API calls
    // server-side: might query database directly
}

func (c *MyDestinationChecker) CheckDestination(destinationName string) error {
    // Check if destination exists
    if !exists {
        return checker.ErrDestinationNotFound
    }
    return nil
}

// Use it with the validator
validator := NewValidator(
    WithDestinationChecker(&MyDestinationChecker{}),
)
```

When a destination checker is provided:
- Destination references produce **errors** if destinations don't exist
- This catches runtime failures at validation time

When no destination checker is provided (default):
- Destination references produce **warnings** reminding users to verify destinations exist
- Allows validation to continue without runtime dependencies

## Testing Strategy

### Test Organization
```
validator/
  schedule_test.go                     - Schedule validation tests
  delivery_test.go                     - Delivery mode tests
  backfill_test.go                     - Backfill tests
  subscribe_test.go                    - Subscribe action tests
  subscribe_events_test.go             - Subscribe event type tests
  duplicate_test.go                    - Duplicate object detection tests
  provider_test.go                     - Provider capability tests
  provider_google_calendar_test.go     - Google Calendar-specific tests
  provider_snowflake_test.go           - Snowflake-specific tests
  spec_version_test.go                 - Spec version tests
  always_enabled_test.go               - Always-enabled object tests
  integration_test.go                  - End-to-end integration tests
  object_test.go                       - Object validation tests

testdata/
  valid/                    - Valid amp.yaml samples for testing
    minimal.yaml
    full-featured.yaml
    google-calendar-valid-backfill.yaml
    snowflake-full-history.yaml
    subscribe-all-event-types.yaml
    field-mapping-bracket-notation.yaml
  invalid/                  - Invalid samples with specific errors
    schedule-too-frequent.yaml
    subscribe-without-read.yaml
    subscribe-no-events.yaml
    subscribe-nested-watch-fields.yaml
    salesforce-too-many-subscribe.yaml
    backfill-both-days-and-fullhistory.yaml
    duplicate-read-objects.yaml
    duplicate-write-objects.yaml
    duplicate-subscribe-objects.yaml
    google-calendar-full-history.yaml
    google-calendar-backfill-too-long.yaml
    snowflake-backfill-days.yaml
    field-mapping-invalid-jsonpath.yaml
    always-enabled-too-many-required.yaml
```

### Test Patterns
- **Table-driven tests**: Most validators use `tests := []struct{...}` pattern
- **Parallel execution**: All tests use `t.Parallel()` for performance
- **Line number accuracy**: Tests verify reported line numbers match actual YAML line
- **testify/require**: For assertions that should stop test execution
- **testify/assert**: For assertions that should continue

### Running Specific Tests
```bash
# Test schedule validation
go test -v -run TestSchedule ./validator

# Test provider-specific validation
go test -v -run TestProvider ./validator

# Test with verbose output and show line numbers
go test -v ./validator -count=1

# Run integration tests
go test -v -run TestIntegration ./validator
```

## Common Development Patterns

### Adding a New Validation Rule

1. **Implement the validator function** in appropriate file (e.g., `validator/schedule.go`)
```go
func validateNewRule(ctx *ValidationContext) {
    manifest := ctx.GetManifest()

    for i, integration := range manifest.Integrations {
        // Validation logic
        if violation {
            path := fmt.Sprintf("$.integrations[%d].fieldName", i)
            pos := ctx.GetPosition(path)
            ctx.AddError(pos, path, "rule-id", "Error message", "Suggestion")
        }
    }
}
```

2. **Add to validator orchestrator** in `validator/validator.go`
```go
func (v *Validator) runValidators(ctx *ValidationContext) {
    // Existing validators...
    validateNewRule(ctx)
}
```

3. **Add tests** in `validator/newrule_test.go`
```go
func TestValidateNewRule(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name     string
        yaml     string
        wantErr  bool
        errRule  string
    }{
        // Test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            result, err := ValidateBytes([]byte(tt.yaml))
            require.NoError(t, err)
            // Assertions
        })
    }
}
```

4. **Add invalid test sample** in `testdata/invalid/newrule-violation.yaml`

### Path Construction for Error Reporting
```go
// Integration level
path := fmt.Sprintf("$.integrations[%d].provider", i)

// Object level
path := fmt.Sprintf("$.integrations[%d].read.objects[%d].schedule", i, j)

// Field mapping level
path := fmt.Sprintf("$.integrations[%d].read.objects[%d].selectedFieldMappings[%d].mapToName", i, j, k)
```

### Provider Capability Checks
```go
providerInfo, err := ctx.GetProvider(integration.Provider)
if err != nil {
    // Provider not found in catalog
    return
}

if integration.Subscribe != nil && !providerInfo.Support.Subscribe {
    ctx.AddError(pos, path, "provider-subscribe-support",
        fmt.Sprintf("Provider %s does not support subscribe", integration.Provider),
        "Remove subscribe section or choose a provider that supports it")
}

// Check module-specific support if module specified
if integration.Module != nil {
    moduleInfo, exists := providerInfo.Modules[*integration.Module]
    if !exists {
        ctx.AddError(pos, path, "provider-module-support",
            fmt.Sprintf("Provider does not support module: %s", *integration.Module),
            "")
    }
}
```

## Important Constants and Limits

### Schedule Constraints
- **Minimum frequency**: 10 minutes (production) - defined in schedule validation
- **Cron format**: 5 fields (minute hour day-of-month month day-of-week)
- **Forbidden**: `*` in minute field (would run every minute)

### Delivery Mode Constraints
- **Auto mode**: Cannot specify pageSize (system-managed)
- **OnRequest mode**: Must specify pageSize between 50-500

### Provider-Specific Limits
- **Salesforce subscribe**: Maximum 5 objects (CDC platform limit)

### Backfill Constraints
- **Mutual exclusivity**: Either `days` or `fullHistory`, not both

## Documentation References

- **README.md**: User-facing documentation, API examples, feature overview
- **ARCHITECTURE.md**: Detailed design decisions, implementation phases, parser approach
- **VALIDATION_RULES.md**: Complete specification of all 69 validation rules with examples

## Development Status

**Current Phase**: Phase 5 Complete (Semantic Validation Enhancement)
- ✅ Phase 1: Documentation and architecture design
- ✅ Phase 2: Universal validation rules implementation
- ✅ Phase 3: Provider-specific validation with catalog integration
- ✅ Phase 4: Async error prevention validation
- ✅ Phase 5: Semantic validation enhancement
  - ✅ Duplicate object detection (read/write/subscribe)
  - ✅ Subscribe event type validation (minimum one event required)
  - ✅ Nested watch fields validation (no dots or brackets)
  - ✅ Google Calendar backfill constraints (no fullHistory, max 28 days for events)
  - ✅ Snowflake backfill requirements (fullHistory only)
  - ✅ JSONPath validation utilities (nested field path detection)

## Key Dependencies

- `github.com/adhocore/gronx`: Cron expression validation and tick calculation
- `github.com/amp-labs/connectors/providers`: Provider catalog for capability checks
- `github.com/go-playground/validator/v10`: Struct tag validation for backfill config
- `gopkg.in/yaml.v3`: YAML parsing with line number preservation (position tracking)
- `sigs.k8s.io/yaml`: YAML unmarshaling into structs with JSON tags (same as server)
- `github.com/spf13/cobra`: CLI framework for command-line tool
- `github.com/oapi-codegen/runtime`: OpenAPI code generation runtime
