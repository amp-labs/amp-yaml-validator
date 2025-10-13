# amp-yaml-validator

A comprehensive validation library for Ampersand `amp.yaml` manifest files.

## Overview

`amp-yaml-validator` is a Go library that validates Ampersand integration manifest files (`amp.yaml`) against the spec version 1.0.0 schema and business rules. It provides detailed error and warning messages with precise line numbers to help developers quickly identify and fix configuration issues.

The library performs:
- **Schema validation**: Ensures structural correctness and required fields
- **Business logic validation**: Enforces semantic rules (schedule frequency, backfill constraints, field mappings)
- **Provider-specific validation**: Checks provider capabilities and limits (Salesforce subscribe limits, provider action support)
- **Async error prevention**: Catches configuration issues that would fail at runtime in Temporal workflows or messenger services

## Features

### Comprehensive Rule Coverage
- **60+ validation rules** covering all aspects of amp.yaml configuration
- **Universal rules**: Apply to all integrations regardless of provider
- **Provider-specific rules**: Validate against provider capabilities from the connectors catalog
- **Async error prevention**: Warn about issues that would fail at runtime

### Precise Error Reporting
- **Line and column numbers**: Every error/warning includes the exact location in the YAML file
- **YAML paths**: JSONPath-style references (e.g., `$.integrations[0].read.objects[1].schedule`)
- **Clear messages**: Human-readable explanations with actionable suggestions
- **Structured output**: JSON and text formats available

### Error Severity Levels
- **ERROR**: Structural or semantic issues that block deployment (invalid schedule, missing required fields)
- **WARNING**: Potential runtime issues or best practice violations (large backfills, missing destinations)

### Easy Integration
- Simple API: `ValidateFile(yamlPath)` or `ValidateBytes(yamlBytes)`
- Functional options pattern for customization
- Graceful degradation when provider catalog is unavailable

## Installation

```bash
go get github.com/amp-labs/amp-yaml-validator
```

## Usage

### Basic Validation

```go
package main

import (
    "fmt"
    "log"

    validator "github.com/amp-labs/amp-yaml-validator"
)

func main() {
    // Validate a file
    result, err := validator.ValidateFile("path/to/amp.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // Check if valid
    if !result.Valid {
        fmt.Println("Validation failed!")

        // Print errors
        for _, issue := range result.Errors {
            fmt.Printf("ERROR [%s] Line %d: %s\n", issue.Rule, issue.Line, issue.Message)
            if issue.Suggestion != "" {
                fmt.Printf("  Suggestion: %s\n", issue.Suggestion)
            }
        }
    }

    // Print warnings (even if valid)
    for _, issue := range result.Warnings {
        fmt.Printf("WARNING [%s] Line %d: %s\n", issue.Rule, issue.Line, issue.Message)
    }
}
```

### Validate Bytes

```go
yamlBytes := []byte(`
specVersion: "1.0.0"
integrations:
  - name: my-integration
    provider: salesforce
    read:
      objects:
        - objectName: Account
          destination: my-db
          schedule: "*/15 * * * *"
`)

result, err := validator.ValidateBytes(yamlBytes)
if err != nil {
    log.Fatal(err)
}
```

### Advanced Usage with Options

```go
// Strict mode - treat warnings as errors
validator := validator.NewValidator(
    validator.WithStrictMode(true),
)
result, err := validator.ValidateFile("amp.yaml")

// Skip provider validation (for testing)
validator := validator.NewValidator(
    validator.WithSkipProviderValidation(),
)

// Custom catalog provider (for testing with mock data)
mockCatalog := catalog.NewMockCatalogProvider(customProviders)
validator := validator.NewValidator(
    validator.WithCatalogProvider(mockCatalog),
)
```

### Structured Output

```go
result, err := validator.ValidateFile("amp.yaml")
if err != nil {
    log.Fatal(err)
}

// Convert to JSON
jsonBytes, _ := json.MarshalIndent(result, "", "  ")
fmt.Println(string(jsonBytes))
```

Output:
```json
{
  "valid": false,
  "errors": [
    {
      "severity": "error",
      "message": "Schedule interval must be at least 10 minutes. Found: */5 * * * * (5 minutes)",
      "line": 12,
      "column": 18,
      "path": "$.integrations[0].read.objects[0].schedule",
      "rule": "schedule-minimum-interval",
      "suggestion": "Change schedule to */10 or greater (e.g., '*/10 * * * *' for every 10 minutes)"
    }
  ],
  "warnings": []
}
```

## Validation Rules

The library validates over 60 rules across multiple categories:

### Universal Rules
- **Spec version**: Must be "1.0.0"
- **Integration structure**: Must have name, provider, and at least one action
- **Read/Write/Subscribe**: Object lists, required fields, destination references
- **Schedules**: Valid cron syntax, minimum 10-minute frequency
- **Delivery modes**: Auto vs onRequest configuration
- **Backfill**: Period constraints (days XOR fullHistory)
- **Always-enabled objects**: Special constraints on field configuration
- **Subscribe actions**: Requires read, inheritFieldsAndMapping, watch fields

**Note**: Field mapping validation (duplicate destination checks) is not currently supported for Manifest types. This validation is only applicable to ConfigContent types, which are outside the scope of amp.yaml manifest validation.

### Provider-Specific Rules (Phase 3)
- **Provider capability checks**: Validates that the provider supports requested actions (read/write/subscribe/proxy)
- **Salesforce-specific rules**:
  - Maximum 5 subscribe objects (Salesforce Change Data Capture API limits)
  - Automatically enforced based on provider catalog
- **Module validation**: Verifies module exists for provider and supports requested actions
- **Graceful degradation**: If provider catalog is unavailable, warnings are issued instead of blocking validation

The library integrates with the `connectors` package to dynamically check provider capabilities:
```go
// Provider capabilities are automatically checked
// Example: Salesforce supports read, write, subscribe, proxy
// Example: Some providers only support read/write but not subscribe
```

### Async Error Prevention (Warnings)
- **Destination references**: Warn if destination names may not exist
- **Object existence**: Warn if object names may be invalid
- **Large backfills**: Risk of workflow timeouts
- **Frequent schedules**: Increased throttling risk

For a complete reference of all validation rules, see [VALIDATION_RULES.md](./VALIDATION_RULES.md).

## Configuration Options

### WithStrictMode(strict bool)
Treat warnings as errors. When enabled, any warnings will cause `Valid` to be false.

```go
result, _ := validator.ValidateFile("amp.yaml", validator.WithStrictMode(true))
```

### WithSkipProviderValidation()
Skip provider-specific validation rules. Useful when provider catalog is unavailable.

```go
result, _ := validator.ValidateFile("amp.yaml", validator.WithSkipProviderValidation())
```

### WithSkipAsyncValidation()
Skip async error prevention rules (warnings about runtime issues).

```go
result, _ := validator.ValidateFile("amp.yaml", validator.WithSkipAsyncValidation())
```

### WithCatalogProvider(provider)
Inject a custom catalog provider for testing or offline validation.

```go
// Use default provider (reads from connectors package)
validator := validator.NewValidator()

// Or use a custom/mock provider for testing
mockCatalog := catalog.NewMockCatalogProvider(map[string]providers.ProviderInfo{
    "test_provider": {
        Name: "test_provider",
        Support: providers.Support{
            Read:      true,
            Write:     false,
            Subscribe: false,
        },
    },
})
validator := validator.NewValidator(validator.WithCatalogProvider(mockCatalog))
```

## Architecture

The library is organized into focused components:

```
amp-yaml-validator/
├── validator/          # Core validation logic
│   ├── parser.go      # YAML parsing with line number tracking
│   ├── universal.go   # Universal validation rules
│   ├── provider.go    # Provider-specific validation rules
│   ├── async.go       # Async error prevention rules
│   └── validator.go   # Main validator orchestration
├── types/             # Type definitions
│   ├── result.go      # ValidationResult, ValidationIssue
│   └── errors.go      # Error constants
├── catalog/           # Provider catalog integration
│   └── provider.go    # Provider capability lookup
└── testdata/          # Test fixtures
    ├── valid/         # Valid amp.yaml samples
    └── invalid/       # Invalid samples for testing
```

Key design features:
- **yaml.Node-based parsing**: Preserves line/column information for accurate error reporting
- **Modular validators**: Universal, provider-specific, and async validators are independent
- **Graceful degradation**: Continues validation if provider catalog is unavailable
- **Extensive testing**: 100% rule coverage with table-driven tests

For detailed architecture documentation, see [ARCHITECTURE.md](./ARCHITECTURE.md).

## Testing

The library has comprehensive test coverage:

### Test Categories
- **Schema validation tests**: Required fields, types, enums, structure
- **Business logic tests**: Schedules, backfill, mappings, subscribe constraints
- **Provider-specific tests**: Salesforce limits, provider capabilities
- **Edge case tests**: Boundary values, empty lists, nil pointers
- **Line number accuracy tests**: Verify reported positions match YAML
- **Valid sample tests**: All samples in `/samples/` must pass
- **Invalid sample tests**: Intentionally broken configs for error detection

### Running Tests

```bash
# Run all tests with race detection
make test

# Run with coverage report
make test-coverage

# Generate HTML coverage report and open in browser
make test-coverage-html

# Run only unit tests
make test-unit

# Run only provider-specific validation tests
make test-provider

# Run tests against sample files
make test-samples

# Run benchmark tests
make benchmark

# Run specific test
go test -run TestValidateSchedule

# Run tests in parallel
go test -parallel 8 ./...
```

### Test Pattern Example

```go
func TestValidateSchedule(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name          string
        schedule      string
        wantErr       bool
        expectedRule  string
    }{
        {
            name:     "valid 15 minute schedule",
            schedule: "*/15 * * * *",
            wantErr:  false,
        },
        {
            name:         "too frequent - 5 minutes",
            schedule:     "*/5 * * * *",
            wantErr:      true,
            expectedRule: "schedule-minimum-interval",
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // Test implementation...
        })
    }
}
```

## Examples

### Example: Invalid Schedule

```yaml
specVersion: "1.0.0"
integrations:
  - name: salesforce-sync
    provider: salesforce
    read:
      objects:
        - objectName: Account
          destination: my-db
          schedule: "*/5 * * * *"  # ERROR: Too frequent (< 10 minutes)
```

Validation output:
```
ERROR [schedule-minimum-interval] Line 8: Schedule interval must be at least 10 minutes. Found: */5 * * * * (5 minutes)
  Path: $.integrations[0].read.objects[0].schedule
  Suggestion: Change schedule to */10 or greater (e.g., '*/10 * * * *' for every 10 minutes)
```

### Example: Subscribe Without Read

```yaml
specVersion: "1.0.0"
integrations:
  - name: salesforce-sync
    provider: salesforce
    subscribe:  # ERROR: Subscribe requires read
      objects:
        - objectName: Account
          destination: my-db
          inheritFieldsAndMapping: true
```

Validation output:
```
ERROR [subscribe-requires-read] Line 5: Subscribe action requires read action to be defined
  Path: $.integrations[0].subscribe
  Suggestion: Add a read section to this integration
```

## Contributing

Contributions are welcome! To add new validation rules or extend the library:

1. **Document the rule**: Add to `VALIDATION_RULES.md` with severity, source reference, and examples
2. **Implement the validator**: Add validation logic to appropriate file (`universal.go`, `provider.go`, etc.)
3. **Add tests**: Create table-driven tests with valid and invalid cases
4. **Update docs**: Add examples to README and update ARCHITECTURE.md if needed

### Adding a New Rule

1. Define the rule in `VALIDATION_RULES.md`:
```markdown
#### Rule: My new rule
- **Severity**: ERROR
- **Source**: `path/to/source.go:123`
- **Description**: What this rule validates
- **Example violation**: ...
- **Example valid**: ...
```

2. Implement validation logic:
```go
func validateMyRule(manifest *Manifest) []ValidationIssue {
    var issues []ValidationIssue
    // Validation logic...
    return issues
}
```

3. Add tests:
```go
func TestValidateMyRule(t *testing.T) {
    tests := []struct{
        name string
        input string
        wantErr bool
    }{
        // Test cases...
    }
    // Test implementation...
}
```

4. Integrate into main validator in `validator.go`

## License

[License information to be added]

## References

- **Ampersand Documentation**: [docs.ampersand.ai](https://docs.ampersand.ai)
- **amp.yaml Schema**: See `openapi/manifest/manifest.yaml` in Ampersand codebase
- **Sample amp.yaml files**: See `/samples/` directory in Ampersand repository
- **Validation rules specification**: See [VALIDATION_RULES.md](./VALIDATION_RULES.md)
- **Architecture documentation**: See [ARCHITECTURE.md](./ARCHITECTURE.md)

## Support

For issues, questions, or contributions:
- **GitHub Issues**: [amp-labs/amp-yaml-validator/issues](https://github.com/amp-labs/amp-yaml-validator/issues)
- **Ampersand Support**: [support@ampersand.ai](mailto:support@ampersand.ai)
- **Documentation**: [docs.ampersand.ai](https://docs.ampersand.ai)

---

**Status**: Phase 3 Complete (Provider-Specific Validation & Testing)
- ✅ Phase 1: Documentation and architecture design
- ✅ Phase 2: Universal validation rules implementation
- ✅ Phase 3: Provider-specific validation with catalog integration
- 🔄 Phase 4: Comprehensive test suite (in progress)
**Version**: 1.0.0-beta
