# amp-yaml-validator Architecture

This document describes the design and architecture of the `amp-yaml-validator` library.

## Overview

The `amp-yaml-validator` library validates Ampersand `amp.yaml` manifest files against spec version 1.0.0. It performs comprehensive validation including schema validation, business logic rules, provider-specific constraints, and async error prevention. The library reports errors and warnings with precise line numbers for easy debugging.

### Design Principles

1. **Modularity**: Validators are organized by concern (universal, provider-specific, async)
2. **Accuracy**: Every error/warning includes exact line and column numbers
3. **Graceful degradation**: Continues validation when optional dependencies (e.g., provider catalog) are unavailable
4. **Extensibility**: Easy to add new validation rules
5. **Performance**: Efficient parsing and validation with parallel test execution
6. **Clarity**: Clear error messages with actionable suggestions

---

## Library Structure

### Package Organization

```
amp-yaml-validator/
├── validator/                    # Core validation logic
│   ├── parser.go                # YAML parsing with line number tracking
│   ├── position.go              # Position tracking types and utilities
│   ├── universal.go             # Universal validation rules
│   ├── provider.go              # Provider-specific validation rules
│   ├── async.go                 # Async error prevention rules
│   ├── validator.go             # Main validator orchestration
│   └── options.go               # Functional options
│
├── types/                        # Type definitions
│   ├── result.go                # ValidationResult, ValidationIssue types
│   ├── manifest.go              # Extended manifest types with positions
│   └── errors.go                # Error constants and helpers
│
├── catalog/                      # Provider catalog integration
│   ├── provider.go              # Provider capability lookup
│   └── cache.go                 # Catalog caching (optional)
│
├── testdata/                     # Test fixtures
│   ├── valid/                   # Valid amp.yaml samples
│   │   ├── minimal.yaml
│   │   ├── full-featured.yaml
│   │   ├── salesforce-read.yaml
│   │   └── ...
│   ├── invalid/                 # Invalid amp.yaml samples for testing
│   │   ├── missing-spec-version.yaml
│   │   ├── invalid-schedule.yaml
│   │   ├── subscribe-without-read.yaml
│   │   ├── duplicate-mappings.yaml
│   │   └── ...
│   └── samples/                 # Copies from /samples/ directory
│
├── validator_test.go            # Integration tests
├── README.md                    # User documentation
├── VALIDATION_RULES.md          # Complete rule specification
└── ARCHITECTURE.md              # This file
```

---

## Component Design

### 1. YAML Parser with Line Number Tracking

**File**: `validator/parser.go`

**Responsibility**: Parse YAML files while maintaining mapping of YAML paths to line/column positions.

#### Types

```go
// Position represents a location in the YAML file
type Position struct {
    Line   int  // 1-based line number
    Column int  // 1-based column number
}

// PositionMap maps YAML paths to their positions in the file
type PositionMap map[string]Position

// ManifestWithPositions wraps the parsed manifest with position metadata
type ManifestWithPositions struct {
    Manifest  *openapi.Manifest
    Positions PositionMap
}
```

#### Implementation Approach

**Step 1: Parse into yaml.Node**
```go
// Use gopkg.in/yaml.v3 for position tracking
import "gopkg.in/yaml.v3"

func ParseWithPositions(yamlBytes []byte) (*ManifestWithPositions, error) {
    // First pass: Parse into yaml.Node to extract positions
    var rootNode yaml.Node
    if err := yaml.Unmarshal(yamlBytes, &rootNode); err != nil {
        return nil, fmt.Errorf("failed to parse YAML: %w", err)
    }

    // Build position map by walking the node tree
    positions := make(PositionMap)
    buildPositionMap(&rootNode, "$", positions)

    // Second pass: Parse into typed structs
    var manifest openapi.Manifest
    if err := yaml.Unmarshal(yamlBytes, &manifest); err != nil {
        return nil, fmt.Errorf("failed to parse manifest: %w", err)
    }

    return &ManifestWithPositions{
        Manifest:  &manifest,
        Positions: positions,
    }, nil
}
```

**Step 2: Build Position Map**
```go
// buildPositionMap recursively walks the yaml.Node tree to extract positions
func buildPositionMap(node *yaml.Node, path string, positions PositionMap) {
    if node == nil {
        return
    }

    // Store position for this node
    positions[path] = Position{
        Line:   node.Line,
        Column: node.Column,
    }

    switch node.Kind {
    case yaml.DocumentNode:
        // Document node wraps the root content
        if len(node.Content) > 0 {
            buildPositionMap(node.Content[0], path, positions)
        }

    case yaml.MappingNode:
        // Mapping node contains key-value pairs
        for i := 0; i < len(node.Content); i += 2 {
            keyNode := node.Content[i]
            valueNode := node.Content[i+1]

            key := keyNode.Value
            childPath := fmt.Sprintf("%s.%s", path, key)

            // Store key position
            positions[childPath] = Position{
                Line:   keyNode.Line,
                Column: keyNode.Column,
            }

            // Recurse into value
            buildPositionMap(valueNode, childPath, positions)
        }

    case yaml.SequenceNode:
        // Sequence node contains array elements
        for i, itemNode := range node.Content {
            childPath := fmt.Sprintf("%s[%d]", path, i)
            buildPositionMap(itemNode, childPath, positions)
        }

    case yaml.ScalarNode:
        // Leaf node - position already stored
        // No recursion needed
    }
}
```

**Step 3: Lookup Positions During Validation**
```go
// GetPosition looks up the position for a given YAML path
func (m *ManifestWithPositions) GetPosition(path string) Position {
    if pos, ok := m.Positions[path]; ok {
        return pos
    }
    // Return zero position if not found
    return Position{Line: 0, Column: 0}
}

// Helper to build paths during validation
func buildReadObjectPath(integrationIdx, objectIdx int) string {
    return fmt.Sprintf("$.integrations[%d].read.objects[%d]", integrationIdx, objectIdx)
}

func buildSchedulePath(integrationIdx, objectIdx int) string {
    return fmt.Sprintf("$.integrations[%d].read.objects[%d].schedule", integrationIdx, objectIdx)
}
```

#### Reference Implementation

The path tracking approach is similar to `cli/files/path_builder.go` which uses a `pathTracker` struct:
```go
// From path_builder.go (reference pattern)
type pathTracker struct {
    path []string
}

func (p *pathTracker) push(segment string) {
    p.path = append(p.path, segment)
}

func (p *pathTracker) pop() {
    if len(p.path) > 0 {
        p.path = p.path[:len(p.path)-1]
    }
}

func (p *pathTracker) current() string {
    return strings.Join(p.path, ".")
}
```

---

### 2. Universal Rule Validator

**File**: `validator/universal.go`

**Responsibility**: Validate rules that apply to all integrations regardless of provider.

#### Function Signature

```go
// ValidateUniversal performs universal validation rules
func ValidateUniversal(mwp *ManifestWithPositions) []ValidationIssue {
    var issues []ValidationIssue

    issues = append(issues, validateSpecVersion(mwp)...)
    issues = append(issues, validateIntegrationStructure(mwp)...)
    issues = append(issues, validateReadObjects(mwp)...)
    issues = append(issues, validateWriteObjects(mwp)...)
    issues = append(issues, validateSubscribeObjects(mwp)...)
    issues = append(issues, validateSchedules(mwp)...)
    issues = append(issues, validateDeliveryModes(mwp)...)
    issues = append(issues, validateBackfill(mwp)...)
    issues = append(issues, validateFieldMappings(mwp)...)
    issues = append(issues, validateAlwaysEnabledObjects(mwp)...)
    issues = append(issues, validateMinimumFields(mwp)...)

    return issues
}
```

#### Example Validator: Schedule Validation

```go
// validateSchedules checks all schedule configurations
func validateSchedules(mwp *ManifestWithPositions) []ValidationIssue {
    var issues []ValidationIssue
    manifest := mwp.Manifest

    for i, integration := range manifest.Integrations {
        if integration.Read != nil {
            for j, obj := range integration.Read.Objects {
                if obj.Schedule == nil || *obj.Schedule == "" {
                    path := buildSchedulePath(i, j)
                    pos := mwp.GetPosition(path)
                    issues = append(issues, ValidationIssue{
                        Severity: "error",
                        Message:  "Schedule is required for read objects",
                        Line:     pos.Line,
                        Column:   pos.Column,
                        Path:     path,
                        Rule:     "read-object-schedule-required",
                    })
                    continue
                }

                schedule := *obj.Schedule
                path := buildSchedulePath(i, j)
                pos := mwp.GetPosition(path)

                // Validate cron syntax
                if !isValidCron(schedule) {
                    issues = append(issues, ValidationIssue{
                        Severity:   "error",
                        Message:    fmt.Sprintf("Invalid cron schedule: %s", schedule),
                        Line:       pos.Line,
                        Column:     pos.Column,
                        Path:       path,
                        Rule:       "schedule-valid-cron",
                        Suggestion: "Use valid cron syntax with 5 fields: minute hour day-of-month month day-of-week",
                    })
                    continue
                }

                // Check for asterisk in minute field
                if hasAsteriskInMinute(schedule) {
                    issues = append(issues, ValidationIssue{
                        Severity:   "error",
                        Message:    "Schedule cannot use '*' in minute field (runs every minute)",
                        Line:       pos.Line,
                        Column:     pos.Column,
                        Path:       path,
                        Rule:       "schedule-no-asterisk-minute",
                        Suggestion: "Use interval syntax like '*/10 * * * *' for every 10 minutes",
                    })
                    continue
                }

                // Check minimum frequency (10 minutes in production)
                freq, err := getScheduleFrequency(schedule)
                if err != nil {
                    issues = append(issues, ValidationIssue{
                        Severity: "error",
                        Message:  fmt.Sprintf("Unable to calculate schedule frequency: %v", err),
                        Line:     pos.Line,
                        Column:   pos.Column,
                        Path:     path,
                        Rule:     "schedule-tick-calculation",
                    })
                    continue
                }

                minFrequency := 10 * time.Minute  // Production minimum
                if freq < minFrequency {
                    issues = append(issues, ValidationIssue{
                        Severity: "error",
                        Message: fmt.Sprintf(
                            "Schedule interval must be at least 10 minutes. Found: %s (%s)",
                            schedule,
                            freq.String(),
                        ),
                        Line:   pos.Line,
                        Column: pos.Column,
                        Path:   path,
                        Rule:   "schedule-minimum-interval",
                        Suggestion: "Change schedule to */10 or greater (e.g., '*/10 * * * *' for every 10 minutes)",
                    })
                }
            }
        }
    }

    return issues
}

// Helper functions
func isValidCron(schedule string) bool {
    // Use github.com/adhocore/gronx for validation
    gron := gronx.New()
    return gron.IsValid(schedule)
}

func hasAsteriskInMinute(schedule string) bool {
    parts := strings.Fields(schedule)
    if len(parts) < 5 {
        return false
    }
    return parts[0] == "*"
}

func getScheduleFrequency(schedule string) (time.Duration, error) {
    // Use gronx to calculate next tick - prev tick
    gron := gronx.New()

    now := time.Now()
    prevTick, err := gron.PrevTick(schedule, now)
    if err != nil {
        return 0, err
    }

    nextTick, err := gron.NextTick(schedule, now)
    if err != nil {
        return 0, err
    }

    return nextTick.Sub(prevTick), nil
}
```

#### Example Validator: Field Mapping Uniqueness

```go
// validateFieldMappings checks for duplicate mapToName values
func validateFieldMappings(mwp *ManifestWithPositions) []ValidationIssue {
    var issues []ValidationIssue
    manifest := mwp.Manifest

    for i, integration := range manifest.Integrations {
        if integration.Read != nil {
            for j, obj := range integration.Read.Objects {
                if obj.SelectedFieldMappings != nil {
                    seen := make(map[string]int)  // mapToName -> first occurrence index

                    for k, mapping := range obj.SelectedFieldMappings {
                        if mapping.MapToName == nil {
                            continue
                        }

                        mapToName := *mapping.MapToName
                        if firstIdx, exists := seen[mapToName]; exists {
                            // Found duplicate
                            path := fmt.Sprintf(
                                "$.integrations[%d].read.objects[%d].selectedFieldMappings[%d].mapToName",
                                i, j, k,
                            )
                            pos := mwp.GetPosition(path)

                            issues = append(issues, ValidationIssue{
                                Severity: "error",
                                Message: fmt.Sprintf(
                                    "Duplicate field mapping: mapToName %q is used more than once",
                                    mapToName,
                                ),
                                Line:   pos.Line,
                                Column: pos.Column,
                                Path:   path,
                                Rule:   "field-mapping-unique",
                                Suggestion: fmt.Sprintf(
                                    "Each mapToName must be unique. This duplicates mapping at index %d",
                                    firstIdx,
                                ),
                            })
                        } else {
                            seen[mapToName] = k
                        }
                    }
                }
            }
        }
    }

    return issues
}
```

#### Rules Implemented in Universal Validator

See VALIDATION_RULES.md section 2 for complete list:
- Spec version validation (`validateSpecVersion`)
- Integration structure validation (`validateIntegrationStructure`)
- Read/write/subscribe object validation (`validateReadObjects`, `validateWriteObjects`, `validateSubscribeObjects`)
- Schedule validation (`validateSchedules`)
- Delivery mode validation (`validateDeliveryModes`)
- Backfill validation (`validateBackfill`)
- Field mapping validation (`validateFieldMappings`)
- Always-enabled object validation (`validateAlwaysEnabledObjects`)
- Minimum field requirements (`validateMinimumFields`)

#### Dependencies

- `github.com/adhocore/gronx`: Cron expression validation
- `github.com/go-playground/validator/v10`: Struct tag validation (for backfill config)

---

### 3. Provider-Specific Rule Validator

**File**: `validator/provider.go`

**Responsibility**: Validate provider-specific constraints and capabilities using the provider catalog.

#### Function Signature

```go
// ValidateProviderSpecific performs provider-specific validation
func ValidateProviderSpecific(mwp *ManifestWithPositions, catalog providers.Catalog) []ValidationIssue {
    var issues []ValidationIssue

    for i, integration := range mwp.Manifest.Integrations {
        providerName := integration.Provider

        // Check provider exists in catalog
        providerInfo, exists := catalog[providerName]
        if !exists {
            path := fmt.Sprintf("$.integrations[%d].provider", i)
            pos := mwp.GetPosition(path)
            issues = append(issues, ValidationIssue{
                Severity:   "error",
                Message:    fmt.Sprintf("Unknown provider: %s", providerName),
                Line:       pos.Line,
                Column:     pos.Column,
                Path:       path,
                Rule:       "provider-exists",
                Suggestion: "Check provider name spelling or update to a supported provider",
            })
            continue
        }

        // Validate provider capabilities
        issues = append(issues, validateProviderCapabilities(mwp, i, &integration, providerInfo)...)

        // Validate provider-specific rules
        issues = append(issues, validateProviderSpecificRules(mwp, i, &integration, providerInfo)...)

        // Validate module support
        if integration.Module != nil && *integration.Module != "" {
            issues = append(issues, validateModule(mwp, i, &integration, providerInfo)...)
        }
    }

    return issues
}
```

#### Provider Capability Validation

```go
// validateProviderCapabilities checks if provider supports the configured actions
func validateProviderCapabilities(
    mwp *ManifestWithPositions,
    integrationIdx int,
    integration *openapi.Integration,
    providerInfo *providers.ProviderInfo,
) []ValidationIssue {
    var issues []ValidationIssue

    // Check read support
    if integration.Read != nil && !providerInfo.Support.Read {
        path := fmt.Sprintf("$.integrations[%d].read", integrationIdx)
        pos := mwp.GetPosition(path)
        issues = append(issues, ValidationIssue{
            Severity:   "error",
            Message:    fmt.Sprintf("Provider %s does not support read action", integration.Provider),
            Line:       pos.Line,
            Column:     pos.Column,
            Path:       path,
            Rule:       "provider-read-support",
            Suggestion: "Remove read section or choose a provider that supports read operations",
        })
    }

    // Check write support
    if integration.Write != nil && !providerInfo.Support.Write {
        path := fmt.Sprintf("$.integrations[%d].write", integrationIdx)
        pos := mwp.GetPosition(path)
        issues = append(issues, ValidationIssue{
            Severity:   "error",
            Message:    fmt.Sprintf("Provider %s does not support write action", integration.Provider),
            Line:       pos.Line,
            Column:     pos.Column,
            Path:       path,
            Rule:       "provider-write-support",
            Suggestion: "Remove write section or choose a provider that supports write operations",
        })
    }

    // Check subscribe support
    if integration.Subscribe != nil && !providerInfo.Support.Subscribe {
        path := fmt.Sprintf("$.integrations[%d].subscribe", integrationIdx)
        pos := mwp.GetPosition(path)
        issues = append(issues, ValidationIssue{
            Severity:   "error",
            Message:    fmt.Sprintf("Provider %s does not support subscribe action", integration.Provider),
            Line:       pos.Line,
            Column:     pos.Column,
            Path:       path,
            Rule:       "provider-subscribe-support",
            Suggestion: fmt.Sprintf(
                "Remove subscribe section. Only %s support subscribe.",
                "Salesforce, HubSpot, Zoho, and Intercom",
            ),
        })
    }

    // Check proxy support
    if integration.Proxy != nil && !providerInfo.Support.Proxy {
        path := fmt.Sprintf("$.integrations[%d].proxy", integrationIdx)
        pos := mwp.GetPosition(path)
        issues = append(issues, ValidationIssue{
            Severity:   "error",
            Message:    fmt.Sprintf("Provider %s does not support proxy action", integration.Provider),
            Line:       pos.Line,
            Column:     pos.Column,
            Path:       path,
            Rule:       "provider-proxy-support",
            Suggestion: "Remove proxy section or choose a provider that supports proxy",
        })
    }

    return issues
}
```

#### Salesforce-Specific Validation

```go
// validateProviderSpecificRules applies provider-specific constraints
func validateProviderSpecificRules(
    mwp *ManifestWithPositions,
    integrationIdx int,
    integration *openapi.Integration,
    providerInfo *providers.ProviderInfo,
) []ValidationIssue {
    var issues []ValidationIssue

    // Salesforce: Maximum 5 subscribe objects
    if integration.Provider == "salesforce" && integration.Subscribe != nil {
        const maxSalesforceSubscribeObjects = 5

        objectCount := len(integration.Subscribe.Objects)
        if objectCount > maxSalesforceSubscribeObjects {
            path := fmt.Sprintf("$.integrations[%d].subscribe.objects", integrationIdx)
            pos := mwp.GetPosition(path)
            issues = append(issues, ValidationIssue{
                Severity: "error",
                Message: fmt.Sprintf(
                    "Salesforce integrations are limited to %d subscribe objects. Found: %d",
                    maxSalesforceSubscribeObjects,
                    objectCount,
                ),
                Line:   pos.Line,
                Column: pos.Column,
                Path:   path,
                Rule:   "salesforce-subscribe-limit",
                Suggestion: fmt.Sprintf(
                    "Reduce to %d or fewer subscribe objects due to Salesforce Change Data Capture limits",
                    maxSalesforceSubscribeObjects,
                ),
            })
        }
    }

    return issues
}
```

#### Module Validation

```go
// validateModule checks if the provider supports the specified module
func validateModule(
    mwp *ManifestWithPositions,
    integrationIdx int,
    integration *openapi.Integration,
    providerInfo *providers.ProviderInfo,
) []ValidationIssue {
    var issues []ValidationIssue

    moduleName := *integration.Module

    moduleInfo, exists := providerInfo.Modules[moduleName]
    if !exists {
        path := fmt.Sprintf("$.integrations[%d].module", integrationIdx)
        pos := mwp.GetPosition(path)

        // List available modules for helpful error message
        availableModules := make([]string, 0, len(providerInfo.Modules))
        for name := range providerInfo.Modules {
            availableModules = append(availableModules, name)
        }
        sort.Strings(availableModules)

        issues = append(issues, ValidationIssue{
            Severity: "error",
            Message: fmt.Sprintf(
                "Provider %s does not support module: %s",
                integration.Provider,
                moduleName,
            ),
            Line:   pos.Line,
            Column: pos.Column,
            Path:   path,
            Rule:   "provider-module-support",
            Suggestion: fmt.Sprintf(
                "Available modules: %s",
                strings.Join(availableModules, ", "),
            ),
        })
    }

    // Note: If module exists, moduleInfo.Support can override providerInfo.Support
    // This would require re-checking capabilities with module-specific support flags

    return issues
}
```

#### Catalog Integration

```go
// File: catalog/provider.go

import "github.com/amp-labs/connectors/providers"

// GetCatalog retrieves the provider catalog with caching
func GetCatalog() (providers.Catalog, error) {
    // Option 1: Direct read
    return providers.ReadCatalog()

    // Option 2: With caching (if implemented)
    // return getCachedCatalog()
}
```

#### Graceful Degradation

```go
// In validator.go main orchestration
func Validate(mwp *ManifestWithPositions, opts ...Option) *ValidationResult {
    var issues []ValidationIssue

    // Always run universal validation
    issues = append(issues, ValidateUniversal(mwp)...)

    // Try provider-specific validation
    if !options.skipProviderValidation {
        catalog, err := catalog.GetCatalog()
        if err != nil {
            // Warn but continue
            issues = append(issues, ValidationIssue{
                Severity:   "warning",
                Message:    "Unable to access provider catalog, skipping provider capability checks",
                Rule:       "catalog-unavailable",
                Suggestion: "Ensure connectors package is available and up to date",
            })
        } else {
            issues = append(issues, ValidateProviderSpecific(mwp, catalog)...)
        }
    }

    return buildResult(issues)
}
```

---

### 4. Async Error Prevention Validator

**File**: `validator/async.go`

**Responsibility**: Detect configuration issues that would cause runtime failures in asynchronous services (Temporal workflows, messenger services). These are typically warnings since they can't always be validated statically.

#### Function Signature

```go
// ValidateAsync performs async error prevention checks (warnings)
func ValidateAsync(mwp *ManifestWithPositions) []ValidationIssue {
    var issues []ValidationIssue

    issues = append(issues, validateDestinationReferences(mwp)...)
    issues = append(issues, validateBackfillRisks(mwp)...)
    issues = append(issues, validateScheduleFrequencyRisks(mwp)...)

    return issues
}
```

#### Destination Reference Validation

```go
// validateDestinationReferences warns about potential destination issues
func validateDestinationReferences(mwp *ManifestWithPositions) []ValidationIssue {
    var issues []ValidationIssue
    manifest := mwp.Manifest

    // Collect all destination references
    destinations := make(map[string][]string)  // destination name -> paths that reference it

    for i, integration := range manifest.Integrations {
        // Check read destinations
        if integration.Read != nil {
            for j, obj := range integration.Read.Objects {
                if obj.Destination != "" {
                    path := fmt.Sprintf("$.integrations[%d].read.objects[%d].destination", i, j)
                    destinations[obj.Destination] = append(destinations[obj.Destination], path)
                }
            }
        }

        // Check write destinations
        if integration.Write != nil {
            for j, obj := range integration.Write.Objects {
                if obj.Destination != nil && *obj.Destination != "" {
                    path := fmt.Sprintf("$.integrations[%d].write.objects[%d].destination", i, j)
                    destinations[*obj.Destination] = append(destinations[*obj.Destination], path)
                }
            }
        }

        // Check subscribe destinations
        if integration.Subscribe != nil {
            for j, obj := range integration.Subscribe.Objects {
                if obj.Destination != "" {
                    path := fmt.Sprintf("$.integrations[%d].subscribe.objects[%d].destination", i, j)
                    destinations[obj.Destination] = append(destinations[obj.Destination], path)
                }
            }
        }
    }

    // Warn about each unique destination
    for destName, paths := range destinations {
        // Can't validate existence statically, so just warn
        path := paths[0]  // Use first occurrence for position
        pos := mwp.GetPosition(path)

        issues = append(issues, ValidationIssue{
            Severity: "warning",
            Message: fmt.Sprintf(
                "Destination %q is referenced but cannot be validated statically",
                destName,
            ),
            Line:       pos.Line,
            Column:     pos.Column,
            Path:       path,
            Rule:       "destination-exists",
            Suggestion: fmt.Sprintf(
                "Ensure destination %q is configured in your project before deploying",
                destName,
            ),
        })
    }

    return issues
}
```

#### Backfill Risk Validation

```go
// validateBackfillRisks warns about large backfills that may timeout
func validateBackfillRisks(mwp *ManifestWithPositions) []ValidationIssue {
    var issues []ValidationIssue
    manifest := mwp.Manifest

    const largeBackfillThreshold = 180  // days

    for i, integration := range manifest.Integrations {
        if integration.Read != nil {
            for j, obj := range integration.Read.Objects {
                if obj.Backfill != nil && obj.Backfill.DefaultPeriod != nil {
                    period := obj.Backfill.DefaultPeriod

                    // Check for full history backfill
                    if period.FullHistory != nil && *period.FullHistory {
                        path := fmt.Sprintf(
                            "$.integrations[%d].read.objects[%d].backfill.defaultPeriod.fullHistory",
                            i, j,
                        )
                        pos := mwp.GetPosition(path)

                        issues = append(issues, ValidationIssue{
                            Severity: "warning",
                            Message: fmt.Sprintf(
                                "Object %s is configured for full history backfill, which may cause timeouts with large datasets",
                                obj.ObjectName,
                            ),
                            Line:       pos.Line,
                            Column:     pos.Column,
                            Path:       path,
                            Rule:       "large-backfill-risk",
                            Suggestion: "Consider using a limited days backfill first (e.g., 90 days) to test performance",
                        })
                    }

                    // Check for large days backfill
                    if period.Days != nil && *period.Days > largeBackfillThreshold {
                        path := fmt.Sprintf(
                            "$.integrations[%d].read.objects[%d].backfill.defaultPeriod.days",
                            i, j,
                        )
                        pos := mwp.GetPosition(path)

                        issues = append(issues, ValidationIssue{
                            Severity: "warning",
                            Message: fmt.Sprintf(
                                "Object %s has a large backfill period (%d days), which may cause timeouts",
                                obj.ObjectName,
                                *period.Days,
                            ),
                            Line:       pos.Line,
                            Column:     pos.Column,
                            Path:       path,
                            Rule:       "large-backfill-risk",
                            Suggestion: fmt.Sprintf(
                                "Consider reducing backfill to %d days or less for initial sync",
                                largeBackfillThreshold,
                            ),
                        })
                    }
                }
            }
        }
    }

    return issues
}
```

#### Schedule Frequency Risk Validation

```go
// validateScheduleFrequencyRisks warns about very frequent schedules
func validateScheduleFrequencyRisks(mwp *ManifestWithPositions) []ValidationIssue {
    var issues []ValidationIssue
    manifest := mwp.Manifest

    const frequentScheduleThreshold = 15 * time.Minute

    for i, integration := range manifest.Integrations {
        if integration.Read != nil {
            for j, obj := range integration.Read.Objects {
                if obj.Schedule == nil {
                    continue
                }

                schedule := *obj.Schedule
                freq, err := getScheduleFrequency(schedule)
                if err != nil {
                    continue  // Skip if can't calculate
                }

                if freq <= frequentScheduleThreshold {
                    path := fmt.Sprintf("$.integrations[%d].read.objects[%d].schedule", i, j)
                    pos := mwp.GetPosition(path)

                    issues = append(issues, ValidationIssue{
                        Severity: "warning",
                        Message: fmt.Sprintf(
                            "Object %s has a very frequent schedule (%s), which may hit rate limits",
                            obj.ObjectName,
                            freq.String(),
                        ),
                        Line:       pos.Line,
                        Column:     pos.Column,
                        Path:       path,
                        Rule:       "frequent-schedule-risk",
                        Suggestion: "Monitor API rate limits when using frequent sync schedules",
                    })
                }
            }
        }
    }

    return issues
}
```

#### Rules Implemented in Async Validator

See VALIDATION_RULES.md section 4:
- Destination reference warnings (`validateDestinationReferences`)
- Large backfill warnings (`validateBackfillRisks`)
- Frequent schedule warnings (`validateScheduleFrequencyRisks`)

---

### 5. Error Reporter

**File**: `types/result.go`

**Responsibility**: Collect validation issues and format output.

#### Types

```go
// ValidationIssue represents a single validation error or warning
type ValidationIssue struct {
    Severity   string `json:"severity"`    // "error" or "warning"
    Message    string `json:"message"`     // Human-readable error message
    Line       int    `json:"line"`        // Line number (1-based)
    Column     int    `json:"column"`      // Column number (1-based, 0 if unavailable)
    Path       string `json:"path"`        // YAML path (e.g., "$.integrations[0].read.objects[1].schedule")
    Rule       string `json:"rule"`        // Rule identifier (e.g., "schedule-minimum-interval")
    Suggestion string `json:"suggestion"`  // Optional suggestion for fixing
}

// ValidationResult contains all validation issues
type ValidationResult struct {
    Valid    bool               `json:"valid"`     // false if any errors present
    Errors   []ValidationIssue  `json:"errors"`    // Blocking errors
    Warnings []ValidationIssue  `json:"warnings"`  // Non-blocking warnings
}
```

#### Builder Pattern

```go
// IssueBuilder helps construct validation issues with fluent API
type IssueBuilder struct {
    issue ValidationIssue
}

func NewError(rule, message string) *IssueBuilder {
    return &IssueBuilder{
        issue: ValidationIssue{
            Severity: "error",
            Rule:     rule,
            Message:  message,
        },
    }
}

func NewWarning(rule, message string) *IssueBuilder {
    return &IssueBuilder{
        issue: ValidationIssue{
            Severity: "warning",
            Rule:     rule,
            Message:  message,
        },
    }
}

func (b *IssueBuilder) At(line, column int) *IssueBuilder {
    b.issue.Line = line
    b.issue.Column = column
    return b
}

func (b *IssueBuilder) WithPath(path string) *IssueBuilder {
    b.issue.Path = path
    return b
}

func (b *IssueBuilder) WithSuggestion(suggestion string) *IssueBuilder {
    b.issue.Suggestion = suggestion
    return b
}

func (b *IssueBuilder) Build() ValidationIssue {
    return b.issue
}

// Usage example:
issue := NewError("schedule-minimum-interval", "Schedule interval must be at least 10 minutes").
    At(12, 18).
    WithPath("$.integrations[0].read.objects[0].schedule").
    WithSuggestion("Change to */10 or greater").
    Build()
```

#### Result Builder

```go
// buildResult constructs ValidationResult from issues
func buildResult(issues []ValidationIssue) *ValidationResult {
    result := &ValidationResult{
        Valid:    true,
        Errors:   make([]ValidationIssue, 0),
        Warnings: make([]ValidationIssue, 0),
    }

    // Separate errors and warnings
    for _, issue := range issues {
        if issue.Severity == "error" {
            result.Errors = append(result.Errors, issue)
            result.Valid = false
        } else if issue.Severity == "warning" {
            result.Warnings = append(result.Warnings, issue)
        }
    }

    // Sort by line number for easier debugging
    sort.Slice(result.Errors, func(i, j int) bool {
        return result.Errors[i].Line < result.Errors[j].Line
    })
    sort.Slice(result.Warnings, func(i, j int) bool {
        return result.Warnings[i].Line < result.Warnings[j].Line
    })

    return result
}
```

#### Text Output Formatter

```go
// FormatText formats validation result as human-readable text
func (r *ValidationResult) FormatText() string {
    var buf strings.Builder

    if r.Valid {
        buf.WriteString("✓ Validation passed")
        if len(r.Warnings) > 0 {
            buf.WriteString(fmt.Sprintf(" with %d warning(s)\n", len(r.Warnings)))
        } else {
            buf.WriteString("\n")
        }
    } else {
        buf.WriteString(fmt.Sprintf("✗ Validation failed with %d error(s)", len(r.Errors)))
        if len(r.Warnings) > 0 {
            buf.WriteString(fmt.Sprintf(" and %d warning(s)", len(r.Warnings)))
        }
        buf.WriteString("\n")
    }

    // Print errors
    if len(r.Errors) > 0 {
        buf.WriteString("\nERRORS:\n")
        for _, issue := range r.Errors {
            buf.WriteString(formatIssue(issue))
        }
    }

    // Print warnings
    if len(r.Warnings) > 0 {
        buf.WriteString("\nWARNINGS:\n")
        for _, issue := range r.Warnings {
            buf.WriteString(formatIssue(issue))
        }
    }

    return buf.String()
}

func formatIssue(issue ValidationIssue) string {
    var buf strings.Builder

    // Line and column
    if issue.Column > 0 {
        buf.WriteString(fmt.Sprintf("  Line %d, Column %d", issue.Line, issue.Column))
    } else {
        buf.WriteString(fmt.Sprintf("  Line %d", issue.Line))
    }

    // Path
    if issue.Path != "" {
        buf.WriteString(fmt.Sprintf(" (%s)", issue.Path))
    }
    buf.WriteString("\n")

    // Message
    buf.WriteString(fmt.Sprintf("    %s\n", issue.Message))

    // Suggestion
    if issue.Suggestion != "" {
        buf.WriteString(fmt.Sprintf("    Suggestion: %s\n", issue.Suggestion))
    }

    buf.WriteString("\n")
    return buf.String()
}
```

---

### 6. Main Validator Orchestrator

**File**: `validator/validator.go`

**Responsibility**: Coordinate all validators and produce final result.

#### Public API

```go
// ValidateFile validates an amp.yaml file at the given path
func ValidateFile(yamlPath string, opts ...Option) (*ValidationResult, error) {
    yamlBytes, err := os.ReadFile(yamlPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    return ValidateBytes(yamlBytes, opts...)
}

// ValidateBytes validates amp.yaml content from bytes
func ValidateBytes(yamlBytes []byte, opts ...Option) (*ValidationResult, error) {
    // Parse with position tracking
    mwp, err := ParseWithPositions(yamlBytes)
    if err != nil {
        return nil, fmt.Errorf("failed to parse YAML: %w", err)
    }

    return ValidateManifest(mwp, opts...)
}

// ValidateManifest validates a parsed manifest
func ValidateManifest(mwp *ManifestWithPositions, opts ...Option) (*ValidationResult, error) {
    // Apply options
    options := defaultOptions()
    for _, opt := range opts {
        opt(&options)
    }

    var issues []ValidationIssue

    // 1. Universal validation (always run)
    issues = append(issues, ValidateUniversal(mwp)...)

    // 2. Provider-specific validation (if enabled and catalog available)
    if !options.skipProviderValidation {
        catalog, err := getCatalogWithOptions(options)
        if err != nil {
            // Warn but continue
            issues = append(issues, ValidationIssue{
                Severity:   "warning",
                Message:    "Unable to access provider catalog, skipping provider capability checks",
                Rule:       "catalog-unavailable",
                Suggestion: "Ensure connectors package is available and up to date",
            })
        } else {
            issues = append(issues, ValidateProviderSpecific(mwp, catalog)...)
        }
    }

    // 3. Async error prevention (if enabled)
    if !options.skipAsyncValidation {
        issues = append(issues, ValidateAsync(mwp)...)
    }

    // 4. Build result
    result := buildResult(issues)

    // 5. Apply strict mode if enabled
    if options.strictMode && len(result.Warnings) > 0 {
        // Convert warnings to errors
        result.Errors = append(result.Errors, result.Warnings...)
        result.Warnings = nil
        result.Valid = false
    }

    return result, nil
}

// getCatalogWithOptions retrieves catalog based on options
func getCatalogWithOptions(options validatorOptions) (providers.Catalog, error) {
    if options.providerCatalog != nil {
        return options.providerCatalog, nil
    }
    return catalog.GetCatalog()
}
```

#### Options Pattern

**File**: `validator/options.go`

```go
// Option is a functional option for validator configuration
type Option func(*validatorOptions)

type validatorOptions struct {
    strictMode            bool
    skipProviderValidation bool
    skipAsyncValidation   bool
    providerCatalog       providers.Catalog
}

func defaultOptions() validatorOptions {
    return validatorOptions{
        strictMode:            false,
        skipProviderValidation: false,
        skipAsyncValidation:   false,
        providerCatalog:       nil,
    }
}

// WithStrictMode treats warnings as errors
func WithStrictMode(strict bool) Option {
    return func(opts *validatorOptions) {
        opts.strictMode = strict
    }
}

// WithSkipProviderValidation skips provider-specific validation
func WithSkipProviderValidation() Option {
    return func(opts *validatorOptions) {
        opts.skipProviderValidation = true
    }
}

// WithSkipAsyncValidation skips async error prevention validation
func WithSkipAsyncValidation() Option {
    return func(opts *validatorOptions) {
        opts.skipAsyncValidation = true
    }
}

// WithProviderCatalog injects a custom provider catalog
func WithProviderCatalog(catalog providers.Catalog) Option {
    return func(opts *validatorOptions) {
        opts.providerCatalog = catalog
    }
}
```

---

## Dependencies

### Required Packages

```go
// go.mod
module github.com/amp-labs/amp-yaml-validator

go 1.21

require (
    github.com/adhocore/gronx v1.6.3                    // Cron validation
    github.com/amp-labs/connectors v0.x.x               // Provider catalog
    github.com/go-playground/validator/v10 v10.15.5     // Struct validation
    gopkg.in/yaml.v3 v3.0.1                             // YAML parsing with positions
)

require (
    github.com/stretchr/testify v1.8.4                  // Testing
)
```

### Import Paths

```go
import (
    "gopkg.in/yaml.v3"
    "github.com/adhocore/gronx"
    "github.com/amp-labs/connectors/providers"
    "github.com/go-playground/validator/v10"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

---

## Testing Strategy

### Test Organization

```
amp-yaml-validator/
├── validator/
│   ├── parser_test.go           # Parser and position tracking tests
│   ├── universal_test.go        # Universal rule tests
│   ├── provider_test.go         # Provider-specific rule tests
│   ├── async_test.go            # Async-risk rule tests
│   └── validator_test.go        # Integration tests
├── testdata/
│   ├── valid/
│   │   ├── minimal.yaml
│   │   ├── full-featured.yaml
│   │   └── ...
│   ├── invalid/
│   │   ├── missing-spec-version.yaml
│   │   ├── invalid-schedule.yaml
│   │   └── ...
│   └── samples/
└── integration_test.go          # End-to-end validation tests
```

### Test Patterns

#### Table-Driven Tests

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
        {
            name:         "asterisk in minute field",
            schedule:     "* * * * *",
            wantErr:      true,
            expectedRule: "schedule-no-asterisk-minute",
        },
        {
            name:         "invalid cron syntax",
            schedule:     "not a cron",
            wantErr:      true,
            expectedRule: "schedule-valid-cron",
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            // Create minimal manifest with schedule
            yaml := fmt.Sprintf(`
specVersion: "1.0.0"
integrations:
  - name: test
    provider: salesforce
    read:
      objects:
        - objectName: Account
          destination: db
          schedule: "%s"
`, tt.schedule)

            result, err := ValidateBytes([]byte(yaml))
            require.NoError(t, err)

            if tt.wantErr {
                assert.False(t, result.Valid)
                assert.NotEmpty(t, result.Errors)

                // Check for expected rule
                found := false
                for _, issue := range result.Errors {
                    if issue.Rule == tt.expectedRule {
                        found = true
                        break
                    }
                }
                assert.True(t, found, "Expected rule %s not found", tt.expectedRule)
            } else {
                assert.True(t, result.Valid)
                assert.Empty(t, result.Errors)
            }
        })
    }
}
```

#### Line Number Accuracy Tests

```go
func TestLineNumberAccuracy(t *testing.T) {
    t.Parallel()

    yaml := `specVersion: "1.0.0"
integrations:
  - name: test
    provider: salesforce
    read:
      objects:
        - objectName: Account
          destination: db
          schedule: "*/5 * * * *"`

    result, err := ValidateBytes([]byte(yaml))
    require.NoError(t, err)

    assert.False(t, result.Valid)
    require.Len(t, result.Errors, 1)

    issue := result.Errors[0]
    assert.Equal(t, "schedule-minimum-interval", issue.Rule)
    assert.Equal(t, 9, issue.Line, "Schedule error should be on line 9")
}
```

#### Valid Sample Tests

```go
func TestValidSamples(t *testing.T) {
    t.Parallel()

    samples, err := filepath.Glob("testdata/samples/*/amp.yaml")
    require.NoError(t, err)

    for _, samplePath := range samples {
        samplePath := samplePath
        t.Run(filepath.Base(filepath.Dir(samplePath)), func(t *testing.T) {
            t.Parallel()

            result, err := ValidateFile(samplePath)
            require.NoError(t, err)
            assert.True(t, result.Valid, "Sample %s should be valid", samplePath)
            assert.Empty(t, result.Errors, "Sample %s should have no errors", samplePath)
        })
    }
}
```

---

## Implementation Phases

### Phase 1: Documentation ✅ Complete
- ✅ VALIDATION_RULES.md: Complete rule specification
- ✅ README.md: User documentation
- ✅ ARCHITECTURE.md: This document

### Phase 2: Core Implementation ✅ Complete
- ✅ Parser with position tracking (`parser/parser.go`)
- ✅ Universal validators (`validator/*.go`)
- ✅ Error reporter (`types/result.go`, `types/errors.go`)
- ✅ Validation context (`validator/context.go`)
- ✅ Basic integration tests
- **Deliverable**: Library that validates universal rules with line numbers

### Phase 3: Provider Integration ✅ Complete
- ✅ Provider-specific validators (`validator/provider.go`)
- ✅ Catalog integration (`catalog/catalog.go`)
- ✅ Provider capability validation (read/write/subscribe/proxy checks)
- ✅ Salesforce-specific rules (max 5 subscribe objects)
- ✅ Module validation support
- ✅ Graceful degradation when catalog unavailable
- ✅ Mock catalog provider for testing
- ✅ Test fixtures (testdata/valid and testdata/invalid)
- **Deliverable**: Full provider-specific validation

### Phase 4: Comprehensive Testing (In Progress)
- 🔄 Async error prevention validators (`validator/async.go`) - Planned
- 🔄 Extensive test suite (100+ test cases) - In Progress
  - validator/provider_test.go
  - validator/spec_version_test.go
  - validator/schedule_test.go
  - validator/subscribe_test.go
  - validator/field_mapping_test.go
  - validator/delivery_test.go
  - validator/backfill_test.go
  - validator/integration_test.go
  - parser/parser_test.go
  - catalog/catalog_test.go
- ✅ Sample validation tests (testdata/valid/*.yaml, testdata/invalid/*.yaml)
- ✅ Make targets for testing (test, test-unit, test-provider, test-coverage-html)
- 🔄 Performance optimization - Planned
- **Deliverable**: Production-ready library with >90% code coverage

---

## Design Decisions

### Why separate universal and provider-specific validators?

**Rationale**:
- Universal rules can be validated without external dependencies (catalog)
- Provider-specific rules require catalog access, which may fail
- Allows graceful degradation: continue with universal validation if catalog unavailable
- Clear separation of concerns makes code easier to maintain and test

### Why use yaml.Node instead of direct struct unmarshaling?

**Rationale**:
- `yaml.Node` preserves line and column information from source file
- Allows building position map before validation
- Enables accurate error reporting with line numbers
- Direct unmarshaling loses position metadata

**Trade-off**: Requires two-pass parsing (once for positions, once for structs), but performance impact is minimal and accuracy benefit is significant.

### Why warnings vs errors?

**Rationale**:
- **Errors**: Structural/semantic issues that violate schema or business logic (must fix before deploy)
- **Warnings**: Potential runtime issues or best practices that can't be fully validated statically (user can choose to ignore or investigate)
- Gives users flexibility: strict mode can treat warnings as errors if desired

**Examples**:
- ERROR: Invalid schedule syntax (can be validated statically)
- WARNING: Large backfill may timeout (depends on runtime data volume)

### Why extensive testing?

**Rationale**:
- Validation library is critical infrastructure
- **False positives** frustrate users (correct configs rejected)
- **False negatives** allow bad configs to deploy (worse - runtime failures)
- **Line number accuracy** is essential for user experience
- Comprehensive tests ensure reliability and catch regressions

**Coverage goals**:
- 100% rule coverage
- 100% line coverage
- Edge cases and boundary conditions
- Real sample files from production

### Why functional options pattern?

**Rationale**:
- Flexible API: easy to add new options without breaking compatibility
- Clear intent: `WithStrictMode(true)` is more readable than bool parameter
- Optional parameters: users only specify what they need
- Testability: easy to inject mocks (e.g., `WithProviderCatalog`)

**Alternative considered**: Config struct, but less ergonomic for simple cases.

---

## Performance Considerations

### Parsing Optimization
- **Two-pass parsing**: Acceptable overhead (< 10ms for typical files)
- **Position map**: Hash map lookup is O(1), minimal impact
- **Lazy validation**: Rules can be short-circuited on first error if desired

### Catalog Caching
- **Optional caching**: Catalog can be cached to avoid repeated reads
- **Invalidation**: Cache invalidation strategy TBD (TTL vs version-based)
- **Memory**: Catalog is ~1-2MB, acceptable for long-running processes

### Parallel Testing
- All tests use `t.Parallel()` for faster CI runs
- No shared state between tests
- Table-driven tests run subtests in parallel

---

## Future Enhancements

### Phase 5+: Advanced Features

1. **Field-level validation**
   - Validate field names against provider metadata
   - Check required fields for specific objects
   - Requires provider object schema integration

2. **Cross-integration validation**
   - Detect duplicate integrations
   - Validate destination consistency across integrations

3. **Custom rule plugins**
   - Allow users to define custom validation rules
   - Plugin architecture for extensibility

4. **IDE integration**
   - Language server protocol (LSP) support
   - Real-time validation in VS Code, IntelliJ
   - Hover tooltips with rule explanations

5. **Auto-fix suggestions**
   - Programmatic fixes for common issues
   - Generate corrected YAML output

6. **Validation metrics**
   - Track most common errors
   - Identify problematic patterns
   - Guide documentation improvements

---

## Conclusion

The `amp-yaml-validator` architecture is designed for:
- **Accuracy**: Precise line number reporting through yaml.Node position tracking
- **Completeness**: 60+ validation rules covering all aspects of amp.yaml
- **Reliability**: Extensive testing with 100% rule and line coverage
- **Usability**: Clear error messages with actionable suggestions
- **Maintainability**: Modular design with clear separation of concerns
- **Extensibility**: Easy to add new rules and validation categories

The library will evolve through well-defined phases, with Phase 1 (documentation) complete and Phase 2 (core implementation) ready to begin.

---

**Document Version**: 1.3
**Last Updated**: 2025-10-12
**Status**: Phase 3 Complete (Provider-Specific Validation & Catalog Integration)
**Next Phase**: Phase 4 - Comprehensive Testing & Async Error Prevention
