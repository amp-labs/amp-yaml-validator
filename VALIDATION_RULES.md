# Ampersand amp.yaml Validation Rules Specification

## 1. Introduction

### Purpose
This document serves as the authoritative specification for all validation rules applied to Ampersand `amp.yaml` manifest files. It catalogs rules discovered through comprehensive analysis of the Ampersand codebase, including validation logic, async error scenarios, and provider-specific constraints.

**Scope:** This specification focuses on validation rules for the `amp.yaml` manifest file structure and semantics. API-layer validation rules (e.g., source validation, installation validation) are documented in the appendix for reference but are considered out-of-scope for the `amp-yaml-validator` library implementation.

### How to Use This Document
This specification is intended for:
- **Implementers** building the `amp-yaml-validator` library
- **Maintainers** updating validation logic as requirements evolve
- **Users** understanding why their amp.yaml files fail validation
- **Documentation writers** creating user-facing validation guides

Each rule includes:
- **Rule description**: What is being validated
- **Severity**: ERROR (blocks deployment) or WARNING (potential runtime issue)
- **Source reference**: File path and line numbers in the codebase
- **Error constant**: Go constant name for the error (if applicable)
- **Rationale**: Why this rule exists (for provider-specific or non-obvious rules)

### Severity Levels

#### ERROR
- **Definition**: Structural or semantic issues that violate the amp.yaml schema or business logic
- **Impact**: Must be fixed before the manifest can be deployed
- **Examples**: Invalid spec version, missing required fields, invalid cron syntax, constraint violations

#### WARNING
- **Definition**: Potential runtime issues or best practice violations that don't violate schema
- **Impact**: Manifest can be deployed, but may fail at runtime or have suboptimal behavior
- **Examples**: Destination references that can't be validated statically, large backfills that may timeout, non-standard naming conventions

### Line Number Requirements
All validation errors and warnings must include:
- **Line number**: The line in the YAML file where the issue occurs (1-based)
- **Column number**: Optional but recommended for precise error location (1-based)
- **YAML path**: JSONPath-style reference (e.g., `$.integrations[0].read.objects[1].schedule`)
- **Context**: The field name or value that caused the issue

Implementation approach uses `gopkg.in/yaml.v3` to parse into `yaml.Node` structures, which preserve position information. See ARCHITECTURE.md for detailed implementation guidance.

---

## 2. Universal Validation Rules (Errors)

These rules apply to all amp.yaml manifests regardless of provider or configuration.

### 2.1 Spec Version Rules

#### Rule: specVersion must be "1.0.0"
- **Severity**: ERROR
- **Source**: `ValidateRevision` function in `server/shared/common/validate.go` (lines 57-60 as reference hint)
- **Error constant**: `errInvalidSpecVersion`
- **Description**: The manifest must specify `specVersion: "1.0.0"` as the first field
- **Example violation**:
```yaml
specVersion: "2.0.0"  # Invalid version
```
- **Example valid**:
```yaml
specVersion: "1.0.0"  # Correct
```

---

### 2.2 Integration Structure Rules

#### Rule: Integration must have at least one action type
- **Severity**: ERROR
- **Source**: `cli/files/manifest.go:211-213`
- **Description**: Each integration must define at least one of: `read`, `write`, `subscribe`, or `proxy`
- **Example violation**:
```yaml
integrations:
  - name: my-integration
    provider: salesforce
    # No actions defined - invalid
```
- **Example valid**:
```yaml
integrations:
  - name: my-integration
    provider: salesforce
    read:
      objects:
        - objectName: Account
```

#### Rule: Integration name required
- **Severity**: ERROR
- **Source**: OpenAPI schema `openapi/manifest/manifest.yaml`
- **Description**: Each integration must have a non-empty `name` field
- **Example violation**:
```yaml
integrations:
  - provider: salesforce  # Missing name
```

#### Rule: Integration provider required
- **Severity**: ERROR
- **Source**: OpenAPI schema `openapi/manifest/manifest.yaml`
- **Description**: Each integration must specify a `provider` field
- **Example violation**:
```yaml
integrations:
  - name: my-integration  # Missing provider
```

---

### 2.3 Read Action Rules

#### Rule: read.objects must be a non-empty list
- **Severity**: ERROR
- **Source**: `validateReadContent` function in `server/shared/common/validate.go` (lines 84-86 as reference hint)
- **Error constant**: `errMissingReadObjects`
- **Description**: If `read` is defined, it must contain at least one object in the `objects` array
- **Example violation**:
```yaml
read:
  objects: []  # Empty list not allowed
```
- **Example valid**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
```

#### Rule: Each read object must have objectName
- **Severity**: ERROR
- **Source**: OpenAPI schema validation
- **Description**: Every object in `read.objects` must specify `objectName`
- **Example violation**:
```yaml
read:
  objects:
    - destination: my-db  # Missing objectName
```

#### Rule: Each read object must have destination
- **Severity**: ERROR
- **Source**: OpenAPI schema validation
- **Description**: Every object in `read.objects` must specify a `destination`
- **Example violation**:
```yaml
read:
  objects:
    - objectName: Account  # Missing destination
```

#### Rule: Each read object must have schedule
- **Severity**: ERROR
- **Source**: OpenAPI schema (required field)
- **Description**: Every object in `read.objects` must specify a `schedule` in cron format
- **Example violation**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      # Missing schedule
```

---

### 2.4 Write Action Rules

#### Rule: write.objects must be a non-empty list
- **Severity**: ERROR
- **Source**: `ValidateRevision` function in `server/shared/common/validate.go` (lines 66-70 as reference hint)
- **Error constant**: `errMissingWriteObjects`
- **Description**: If `write` is defined, it must contain at least one object in the `objects` array
- **Example violation**:
```yaml
write:
  objects: []  # Empty list not allowed
```
- **Example valid**:
```yaml
write:
  objects:
    - objectName: Contact
```

#### Rule: Each write object must have objectName
- **Severity**: ERROR
- **Source**: OpenAPI schema validation
- **Description**: Every object in `write.objects` must specify `objectName`
- **Example violation**:
```yaml
write:
  objects:
    - {}  # Missing objectName
```

---

### 2.5 Subscribe Action Rules

#### Rule: Subscribe requires read to be defined
- **Severity**: ERROR
- **Source**: `validateSubscribeContent` function in `server/shared/common/validate.go` (lines 126-128 as reference hint)
- **Error constant**: `ErrSubscribeRequiresRead`
- **Description**: If `subscribe` is present, the integration must also define `read`. Subscribe and read work together to provide real-time updates.
- **Rationale**: Subscribe creates event listeners, but read must be configured to process the events
- **Example violation**:
```yaml
integrations:
  - name: my-integration
    provider: salesforce
    subscribe:
      objects:
        - objectName: Account
          destination: my-db
    # Missing read - invalid
```
- **Example valid**:
```yaml
integrations:
  - name: my-integration
    provider: salesforce
    read:
      objects:
        - objectName: Account
          destination: my-db
          schedule: "*/15 * * * *"
    subscribe:
      objects:
        - objectName: Account
          destination: my-db
```

#### Rule: subscribe.objects must be a non-empty list
- **Severity**: ERROR
- **Source**: `validateSubscribeContent` function in `server/shared/common/validate.go` (lines 130-132 as reference hint)
- **Error constant**: `errMissingSubscribeObjects`
- **Description**: If `subscribe` is defined, it must contain at least one object
- **Example violation**:
```yaml
subscribe:
  objects: []  # Empty list not allowed
```

#### Rule: inheritFieldsAndMapping must be true for subscribe objects with events
- **Severity**: ERROR
- **Source**: `validateSubscribeContent` function in `server/shared/common/validate.go` (lines 143-145 as reference hint)
- **Error constant**: `ErrSubscribeInheritFieldsAndMapping`
- **Description**: In spec version 1.0.0, subscribe objects that define `createEvent`, `updateEvent`, `deleteEvent`, or `associationChangeEvent` must have `inheritFieldsAndMapping: true`. Independent field configuration for subscribe is not supported in v1. Objects with no events (base definitions) and pure `otherEvents` (passThrough) subscriptions have nothing to map and are exempt, matching the server's revision validation.
- **Rationale**: v1 simplifies configuration by requiring subscribe to use the same field configuration as the corresponding read object
- **Example violation**:
```yaml
subscribe:
  objects:
    - objectName: Account
      destination: my-db
      inheritFieldsAndMapping: false  # Invalid in v1
```
- **Example valid**:
```yaml
subscribe:
  objects:
    - objectName: Account
      destination: my-db
      inheritFieldsAndMapping: true
```

#### Rule: Each subscribe object must have objectName
- **Severity**: ERROR
- **Source**: OpenAPI schema validation
- **Description**: Every object in `subscribe.objects` must specify `objectName`

#### Rule: Each subscribe object must have destination
- **Severity**: ERROR
- **Source**: OpenAPI schema validation
- **Description**: Every object in `subscribe.objects` must specify a `destination`

---

### 2.6 Subscribe Update Event Rules

#### Rule: updateEvent.enabled must be "always" if set
- **Severity**: ERROR
- **Source**: `validateSubscribeContent` function in `server/shared/common/validate.go` (lines 153-155 as reference hint)
- **Error constant**: `ErrInvalidInputEnabled`
- **Description**: If `updateEvent.enabled` is specified, its value must be `"always"`. This is the only supported mode in v1.
- **Example violation**:
```yaml
subscribe:
  objects:
    - objectName: Account
      destination: my-db
      updateEvent:
        enabled: "sometimes"  # Invalid value
```
- **Example valid**:
```yaml
subscribe:
  objects:
    - objectName: Account
      destination: my-db
      updateEvent:
        enabled: "always"
```

#### Rule: UpdateEvent must specify watch fields
- **Severity**: ERROR
- **Source**: `validateSubscribeContent` function in `server/shared/common/validate.go` (lines 157-160 as reference hint)
- **Error constant**: `ErrWatchFieldsRequired`
- **Description**: When `updateEvent` is configured, you must specify EITHER:
  - `requiredWatchFields`: A non-empty list of field names to watch
  - `watchFieldsAuto`: Set to `"all"` or `"selected"`
- **Rationale**: Update events need to know which fields to monitor for changes
- **Example violation**:
```yaml
subscribe:
  objects:
    - objectName: Account
      destination: my-db
      updateEvent:
        enabled: "always"
        # Neither requiredWatchFields nor watchFieldsAuto specified
```
- **Example valid (explicit fields)**:
```yaml
subscribe:
  objects:
    - objectName: Account
      destination: my-db
      updateEvent:
        enabled: "always"
        requiredWatchFields:
          - Name
          - Email
```
- **Example valid (auto mode)**:
```yaml
subscribe:
  objects:
    - objectName: Account
      destination: my-db
      updateEvent:
        enabled: "always"
        watchFieldsAuto: "all"
```

#### Rule: Cannot use both watchFieldsAuto and requiredWatchFields
- **Severity**: ERROR
- **Source**: `validateSubscribeContent` function in `server/shared/common/validate.go` (lines 162-165 as reference hint)
- **Error constant**: `ErrWatchFieldsAndRequiredWatchFields`
- **Description**: You must choose ONE watch field configuration method, not both
- **Example violation**:
```yaml
subscribe:
  objects:
    - objectName: Account
      destination: my-db
      updateEvent:
        enabled: "always"
        watchFieldsAuto: "all"
        requiredWatchFields:  # Cannot specify both
          - Name
```

---

### 2.7 Schedule Validation Rules

#### Rule: Schedule must be valid cron syntax (5 parts)
- **Severity**: ERROR
- **Source**: `validateSchedule` function in `server/shared/common/validate.go` (lines 196-204 as reference hint)
- **Error constant**: `errInvalidCronSchedule`, `ErrInvalidSchedule`
- **Description**: Schedule must be a valid cron expression with exactly 5 parts: minute, hour, day-of-month, month, day-of-week
- **Implementation**: Uses `github.com/adhocore/gronx` for validation
- **Example violation**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "invalid cron"  # Not valid cron syntax
```
- **Example valid**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"  # Every 15 minutes
```

#### Rule: Schedule cannot have '*' in minute field
- **Severity**: ERROR
- **Source**: `validateSchedule` function in `server/shared/common/validate.go` (lines 206-208 as reference hint)
- **Error constant**: `ErrScheduleTooFrequent`
- **Description**: The minute field cannot be `*` (run every minute). Must use interval syntax like `*/n`
- **Rationale**: Running every minute is too frequent and could overload systems
- **Example violation**:
```yaml
schedule: "* * * * *"  # Every minute - too frequent
```
- **Example valid**:
```yaml
schedule: "*/10 * * * *"  # Every 10 minutes
```

#### Rule: Schedule minimum interval must be 10 minutes
- **Severity**: ERROR
- **Source**: `validateSchedule` function in `server/shared/common/validate.go` (lines 210-226 as reference hint)
- **Error constant**: `ErrScheduleTooFrequent`
- **Description**: In production environments, schedules must have a minimum interval of 10 minutes between runs. The `*/n` syntax in the minute field must have `n >= 10`.
- **Rationale**: Prevents system overload from too-frequent syncs
- **Note**: In local development, minimum is 30 seconds (configured in `schedule_const_local.go`)
- **Example violation**:
```yaml
schedule: "*/5 * * * *"  # Every 5 minutes - too frequent for production
```
- **Example valid**:
```yaml
schedule: "*/10 * * * *"  # Every 10 minutes - minimum allowed
schedule: "*/30 * * * *"  # Every 30 minutes
schedule: "0 */2 * * *"   # Every 2 hours
schedule: "0 9 * * *"     # Daily at 9am
```

#### Rule: Schedule must have valid prev/next tick calculation
- **Severity**: ERROR
- **Source**: `validateSchedule` function in `server/shared/common/validate.go` (lines 228-243 as reference hint)
- **Description**: The cron expression must be valid enough to calculate previous and next execution times
- **Implementation**: Uses gronx `PrevTick()` and `NextTick()` validation
- **Note**: Part of minFrequency validation logic

#### Rule: Schedule interval between ticks must meet minimum frequency
- **Severity**: ERROR
- **Source**: `validateSchedule` function in `server/shared/common/validate.go` (lines 228-243 as reference hint)
- **Error constant**: `ErrScheduleTooFrequent`
- **Description**: The calculated time between consecutive schedule executions must be >= minFrequency (10 minutes in production, 30 seconds local)
- **Implementation**: Validates by computing `nextTick - prevTick >= minFrequency`
- **Note**: `minFrequency` defined in `schedule_const.go` (prod) and `schedule_const_local.go` (local)
- **Rationale**: Ensures schedules don't create excessive load regardless of cron syntax used

---

### 2.8 Delivery Mode Rules

#### Rule: If delivery.mode is "auto", pageSize must not be set
- **Severity**: ERROR
- **Source**: `validateDeliveryMode` function in `server/shared/common/validate.go` (lines 254-262 as reference hint)
- **Error constant**: `ErrInvalidPageSizeConfig`
- **Description**: When using automatic delivery mode, pageSize is managed by the system and cannot be manually specified
- **Example violation**:
```yaml
read:
  delivery:
    mode: auto
    pageSize: 100  # Cannot set pageSize in auto mode
```
- **Example valid**:
```yaml
read:
  delivery:
    mode: auto
    # pageSize not specified - system manages it
```

#### Rule: If delivery.mode is "onRequest", pageSize is required
- **Severity**: ERROR
- **Source**: `validateDeliveryMode` function in `server/shared/common/validate.go` (lines 264-266 as reference hint)
- **Error constant**: `ErrInvalidPageSizeConfig`
- **Description**: When using on-request delivery, you must specify the page size
- **Example violation**:
```yaml
read:
  delivery:
    mode: onRequest
    # Missing pageSize
```
- **Example valid**:
```yaml
read:
  delivery:
    mode: onRequest
    pageSize: 100
```

#### Rule: If delivery.mode is "onRequest", pageSize must be between 50 and 500
- **Severity**: ERROR
- **Source**: `validateDeliveryMode` function in `server/shared/common/validate.go` (lines 268-274 as reference hint)
- **Constants**: `minOnRequestPageSize = 50`, `maxOnRequestPageSize = 500`
- **Description**: On-request page size must be within the allowed range
- **Rationale**: Too small is inefficient, too large may cause timeouts or memory issues
- **Example violation**:
```yaml
read:
  delivery:
    mode: onRequest
    pageSize: 25  # Too small (< 50)
```
```yaml
read:
  delivery:
    mode: onRequest
    pageSize: 1000  # Too large (> 500)
```
- **Example valid**:
```yaml
read:
  delivery:
    mode: onRequest
    pageSize: 100  # Within range [50, 500]
```

#### Rule: delivery.mode must be "auto" or "onRequest"
- **Severity**: ERROR
- **Source**: `validateDeliveryMode` function in `server/shared/common/validate.go` (lines 277-279 as reference hint)
- **Error constant**: `ErrInvalidDeliveryConfig`
- **Description**: Only two delivery modes are supported
- **Example violation**:
```yaml
read:
  delivery:
    mode: streaming  # Invalid mode
```
- **Example valid**:
```yaml
read:
  delivery:
    mode: auto  # or "onRequest"
```

---

### 2.9 Backfill Validation Rules

#### Rule: Backfill config must pass struct validation
- **Severity**: ERROR
- **Source**: `ValidateBackfillFromConfig` function in `server/shared/common/validate.go` (lines 289-291 as reference hint)
- **Error constant**: `errInvalidBackfill`
- **Description**: Backfill configuration must satisfy go-validator constraints defined in the struct tags
- **Implementation**: Uses `github.com/go-playground/validator` for struct validation

#### Rule: Must provide only one of defaultPeriod.days OR defaultPeriod.fullHistory
- **Severity**: ERROR
- **Source**: `ValidateBackfillFromConfig` function in `server/shared/common/validate.go` (lines 293-299 as reference hint)
- **Error constant**: `errInvalidBackfill`
- **Description**: Backfill default period configuration is mutually exclusive: specify either `days` (number of days to backfill) OR `fullHistory` (boolean to backfill all historical data), but not both
- **Note**: Schema validation uses `required_without` tags to enforce this
- **Example violation**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      backfill:
        defaultPeriod:
          days: 30
          fullHistory: true  # Cannot specify both
```
- **Example valid (days)**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      backfill:
        defaultPeriod:
          days: 30
```
- **Example valid (fullHistory)**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      backfill:
        defaultPeriod:
          fullHistory: true
```

---

### 2.10 Field Mapping Rules

#### Rule: No duplicate field mappings
- **Severity**: ERROR
- **Source**: `ValidateNoDuplicateFieldMappings` function in `server/shared/common/validate.go` (lines 334-338 as reference hint)
- **Error constant**: `errDuplicateFieldMapping`
- **Description**: Two source fields cannot map to the same destination field name. Each `mapToName` must be unique within an object's field configuration.
- **Rationale**: Duplicate mappings would cause data conflicts and ambiguous field resolution
- **Example violation**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      selectedFieldMappings:
        - fieldName: FirstName
          mapToName: name
        - fieldName: LastName
          mapToName: name  # Duplicate mapToName
```
- **Example valid**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      selectedFieldMappings:
        - fieldName: FirstName
          mapToName: first_name
        - fieldName: LastName
          mapToName: last_name
```

#### Rule: Required mappings must be present in selectedFieldMappings
- **Severity**: ERROR
- **Source**: `ValidateFields` function in `server/shared/common/validate.go` (lines 475-501 as reference hint)
- **Error constant**: `ErrMissingMinimumRequiredFields`
- **Description**: If an object specifies `requiredFieldMappings`, all of those mappings must also appear in `selectedFieldMappings`. Required mappings define the minimum set of fields that must be included.
- **Rationale**: Ensures that required fields are actually selected for syncing
- **Example violation**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      requiredFieldMappings:
        - fieldName: Id
          mapToName: id
        - fieldName: Name
          mapToName: name
      selectedFieldMappings:
        - fieldName: Id
          mapToName: id
        # Missing Name mapping - invalid
```
- **Example valid**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      requiredFieldMappings:
        - fieldName: Id
          mapToName: id
        - fieldName: Name
          mapToName: name
      selectedFieldMappings:
        - fieldName: Id
          mapToName: id
        - fieldName: Name
          mapToName: name
        - fieldName: Email
          mapToName: email
```

---

### 2.11 Always-Enabled Object Rules

Always-enabled objects are objects that are automatically enabled for all installations without user configuration. They have stricter validation requirements.

#### Rule: Always-enabled read objects must have requiredFields
- **Severity**: ERROR
- **Source**: `validateReadContent` function in `server/shared/common/validate.go` (lines 90-92 as reference hint)
- **Error constant**: `ErrMissingMinimumRequiredFields`
- **Description**: Objects with `alwaysEnabled: true` must specify a non-empty `requiredFields` list
- **Rationale**: Always-enabled objects must define minimum field requirements since users don't configure them
- **Example violation**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      alwaysEnabled: true
      # Missing requiredFields
```
- **Example valid**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      alwaysEnabled: true
      requiredFields:
        - fieldName: Id
        - fieldName: Name
```

#### Rule: Always-enabled objects cannot have mapToName in requiredFields
- **Severity**: ERROR
- **Source**: `validateReadContent` function in `server/shared/common/validate.go` (lines 100-102 as reference hint)
- **Error constant**: `ErrMissingMinimumRequiredFields`
- **Description**: For always-enabled objects, `requiredFields` must use `IntegrationFieldExistent` (fieldName only), not `IntegrationFieldMapping` (with mapToName)
- **Rationale**: Field mappings should be defined separately in `requiredFieldMappings`
- **Example violation**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      alwaysEnabled: true
      requiredFields:
        - fieldName: Id
          mapToName: id  # Cannot use mapToName here
```
- **Example valid**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      alwaysEnabled: true
      requiredFields:
        - fieldName: Id
        - fieldName: Name
```

#### Rule: Always-enabled objects must have schedule
- **Severity**: ERROR
- **Source**: `validateReadContent` function in `server/shared/common/validate.go` (lines 105-107 as reference hint)
- **Error constant**: `ErrInvalidSchedule`
- **Description**: Objects with `alwaysEnabled: true` must define a sync schedule
- **Example violation**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      alwaysEnabled: true
      requiredFields:
        - fieldName: Id
      # Missing schedule
```

#### Rule: Always-enabled objects cannot have more required mappings than builder mappings
- **Severity**: ERROR
- **Source**: `ValidateFields` function in `server/shared/common/validate.go` (lines 468-471 as reference hint)
- **Error constant**: `ErrAlwaysEnabledObjectHasMoreRequiredMappingsThanBuilderMappings`
- **Description**: For always-enabled objects, the number of `requiredFieldMappings` cannot exceed the number of `builderFieldMappings`. Builder mappings define the UI-configurable fields.
- **Rationale**: All required mappings must be configurable through the builder interface
- **Example violation**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      alwaysEnabled: true
      requiredFieldMappings:
        - fieldName: Id
          mapToName: id
        - fieldName: Name
          mapToName: name
        - fieldName: Email
          mapToName: email
      builderFieldMappings:
        - fieldName: Id
          mapToName: id
        - fieldName: Name
          mapToName: name
        # Only 2 builder mappings, but 3 required mappings - invalid
```

#### Rule: Always-enabled objects must have at least one field configured
- **Severity**: ERROR
- **Source**: `ValidateFields` function in `server/shared/common/validate.go` (lines 463-466 as reference hint)
- **Description**: Always-enabled objects must specify at least one field via `requiredFields` or builder mappings
- **Rationale**: Objects need at least one field to sync meaningful data
- **Example violation**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      alwaysEnabled: true
      # No fields specified
```

---

### 2.12 Minimum Field Requirements

#### Rule: Each read object must have at least one field selected
- **Severity**: ERROR
- **Source**: `ValidateFields` function in `server/shared/common/validate.go` (lines 521-524 as reference hint)
- **Error constant**: `ErrMissingMinimumRequiredFields`
- **Description**: Every read object must specify at least one field to sync via one of:
  - `selectedFields` (list of field names)
  - `selectedFieldMappings` (list of field mappings)
  - `selectedFieldsAuto: true` (automatically select all available fields)
- **Rationale**: Objects need at least one field to provide useful data
- **Example violation**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      # No field selection specified
```
- **Example valid (selectedFields)**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      selectedFields:
        - fieldName: Id
        - fieldName: Name
```
- **Example valid (selectedFieldMappings)**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      selectedFieldMappings:
        - fieldName: Id
          mapToName: id
```
- **Example valid (selectedFieldsAuto)**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      selectedFieldsAuto: true
```

---

### 2.13 Duplicate Object Detection Rules

#### Rule: No duplicate objects within same action
- **Severity**: ERROR
- **Rule ID**: `duplicate-read-object`, `duplicate-write-object`, `duplicate-subscribe-object`
- **Source**: `validator/duplicate.go`
- **Error constants**: `ErrDuplicateReadObject`, `ErrDuplicateWriteObject`, `ErrDuplicateSubscribeObject`
- **Description**: The same `objectName` cannot appear twice within the same action (read.objects, write.objects, or subscribe.objects)
- **Rationale**: Duplicate objects create ambiguous configuration and undefined behavior at runtime. The system cannot determine which configuration to use for the same object.
- **Example violation**:
```yaml
integrations:
  - provider: salesforce
    read:
      objects:
        - objectName: Account
          destination: webhook
          schedule: "0 */12 * * *"
          selectedFields:
            Name: true
        - objectName: Account  # Duplicate - ERROR
          destination: webhook
          schedule: "0 */6 * * *"
          selectedFields:
            Email: true
```
- **Example valid (same object in different actions is allowed)**:
```yaml
integrations:
  - provider: salesforce
    read:
      objects:
        - objectName: Account
          destination: webhook
          schedule: "0 */12 * * *"
    write:
      objects:
        - objectName: Account  # Same object in write is OK
          selectedFieldSettings:
            Name:
              writeOnCreate: always
```

---

### 2.14 Subscribe Event Type Rules

#### Rule: Subscribe object with no event type enabled warns
- **Severity**: WARNING
- **Rule ID**: `subscribe-minimum-events`
- **Source**: `validator/subscribe_events.go`
- **Error constant**: `ErrNoSubscribeEvents`
- **Description**: A subscribe object with no event type enabled (`createEvent`, `updateEvent`, `deleteEvent`, `associationChangeEvent`, or `otherEvents`) produces a warning
- **Rationale**: Such an object is a valid base definition: it only supplies defaults (such as `destination`) and subscribes to nothing until an installation config enables an event. The warning keeps an accidentally event-less object visible; suppress it with an `amp:ignore` directive when the base definition is intentional
- **Example warning**:
```yaml
subscribe:
  objects:
    - objectName: Account
      destination: webhook
      # No events enabled - WARNING (base definition; installations opt in via config)
```
- **Example valid**:
```yaml
subscribe:
  objects:
    - objectName: Account
      destination: webhook
      inheritFieldsAndMapping: true
      updateEvent:
        enabled: always
        watchFieldsAuto: all
```

---

### 2.15 Field Mapping Naming Rules

#### Rule: Field mappings must use simple field names (no bracket notation)
- **Severity**: NOTE
- **Rule ID**: `field-mapping-simple-names`
- **Source**: `validator/jsonpath.go`
- **Description**: Field mappings in `mapToName` should use simple field names without bracket notation or complex JSONPath expressions
- **Rationale**: The system is designed to work with simple field name mappings. Bracket notation and nested path expressions are intentionally not supported
- **Current Implementation**: The `isNestedFieldPath()` function detects bracket notation but does not enforce restrictions (no error/warning generated for Manifest types)
- **Note**: This is an intentional design decision, not a limitation
- **Example (simple field names - standard pattern)**:
```yaml
read:
  objects:
    - objectName: Account
      destination: webhook
      selectedFieldMappings:
        Name: name
        Email: contact_email
        Phone: phone_number
```

---

### 2.16 RequiredWatchFields Nested Path Rules

#### Rule: RequiredWatchFields cannot contain nested paths
- **Severity**: ERROR
- **Rule ID**: `watch-fields-no-nesting`
- **Source**: `validator/subscribe_events.go`, `validator/jsonpath.go`
- **Error constant**: `ErrNestedWatchField`
- **Description**: The `requiredWatchFields` array in `updateEvent` cannot contain nested paths (fields with dots or brackets)
- **Rationale**: Provider CDC/webhook implementations don't support watching nested fields. Only top-level fields can trigger update events
- **Example violation**:
```yaml
subscribe:
  objects:
    - objectName: Account
      destination: webhook
      inheritFieldsAndMapping: true
      updateEvent:
        enabled: always
        requiredWatchFields:
          - Name
          - Address.Street  # Nested path with dot - ERROR
          - Contact[0].Email  # Bracket notation - ERROR
```
- **Example valid**:
```yaml
subscribe:
  objects:
    - objectName: Account
      destination: webhook
      inheritFieldsAndMapping: true
      updateEvent:
        enabled: always
        requiredWatchFields:
          - Name
          - Email  # Top-level field only - valid
```

---

### 2.17 Unknown Key Detection Rules

#### Rule: Keys must be part of the amp.yaml schema

- **Severity**: WARNING
- **Rule ID**: `unknown-key`
- **Source**: `validator/unknown_keys.go`, `parser/unknown.go` (`DetectUnknownKeys`)
- **Description**: Any mapping key that does not correspond to a field in the amp.yaml v1.0.0 schema is reported. Such "orphan" keys are silently dropped during unmarshaling, so they are almost always typos or misplaced configuration (e.g. `scheduel` instead of `schedule`, or a valid key nested under the wrong parent).
- **Rationale**: The manifest is unmarshaled with `sigs.k8s.io/yaml`, which ignores unrecognized keys without error. Without this check, a misspelled field name simply has no effect, producing confusing "why is my config being ignored?" behavior at runtime. The warning surfaces the exact key and line so it can be fixed.
- **Behavior notes**:
  - Reports the key's own line/column (not the value's).
  - An unknown parent key is reported once; detection does not descend into its (unvalidatable) subtree.
  - Matching is case-insensitive, mirroring `encoding/json`, so case variants of a valid key are not flagged.
  - Keys inside union-typed field entries (`requiredFields`/`optionalFields`) are checked against the union of both allowed shapes (existent field and field mapping).
  - Arbitrary-key maps (`additionalProperties`) are not flagged.
  - Suppressible per-key or per-subtree with an `amp:ignore` / `amp:ignore[unknown-key]` directive.
- **Example violation**:
```yaml
specVersion: 1.0.0
descrption: typo of "description"  # WARNING: unknown-key
integrations:
  - name: readSalesforce
    provider: salesforce
    read:
      objects:
        - objectName: account
          destination: webhook
          scheduel: "*/10 * * * *"  # WARNING: unknown-key (typo of "schedule")
          requiredFields:
            - feildName: name  # WARNING: unknown-key (typo of "fieldName")
```
- **Example valid**:
```yaml
specVersion: 1.0.0
integrations:
  - name: readSalesforce
    provider: salesforce
    read:
      objects:
        - objectName: account
          destination: webhook
          schedule: "*/10 * * * *"
          requiredFields:
            - fieldName: name
```

---

## 3. Provider-Specific Validation Rules

These rules apply only to specific providers or depend on provider capabilities.

### 3.1 Salesforce Rules

#### Rule: Maximum 5 subscribe objects for Salesforce
- **Severity**: ERROR
- **Source**: `validateSubscribeContent` function in `server/shared/common/validate.go` (lines 135-138 as reference hint)
- **Constant**: `maxSalesforceSubscribeObjects = 5`
- **Error constant**: `ErrTooManySubscribeObjects`
- **Description**: Salesforce integrations are limited to 5 subscribe objects maximum
- **Rationale**: Salesforce Change Data Capture has platform limits on the number of CDC channels per organization
- **Example violation**:
```yaml
integrations:
  - name: my-salesforce-integration
    provider: salesforce
    subscribe:
      objects:
        - objectName: Account
          destination: my-db
        - objectName: Contact
          destination: my-db
        - objectName: Opportunity
          destination: my-db
        - objectName: Lead
          destination: my-db
        - objectName: Case
          destination: my-db
        - objectName: Campaign
          destination: my-db  # 6th object - exceeds limit
```
- **Example valid**:
```yaml
integrations:
  - name: my-salesforce-integration
    provider: salesforce
    subscribe:
      objects:
        - objectName: Account
          destination: my-db
        - objectName: Contact
          destination: my-db
        - objectName: Opportunity
          destination: my-db
        - objectName: Lead
          destination: my-db
        - objectName: Case
          destination: my-db
        # Maximum 5 objects
```

---

### 3.2 HubSpot Rules

#### Rule: HubSpot provider apps require scopes
- **Severity**: ERROR
- **Source**: `server/api/routes/api/providerApp.go:58-62`
- **Description**: When creating a HubSpot provider app, the `scopes` field must be specified
- **Note**: This rule applies to provider app configuration, not directly to amp.yaml manifests
- **Example violation**:
```yaml
# In provider app configuration
provider: hubspot
# Missing scopes field
```
- **Example valid**:
```yaml
# In provider app configuration
provider: hubspot
scopes:
  - crm.objects.contacts.read
  - crm.objects.companies.read
```

#### Note: HubSpot Subscribe Limitations
- **Limitation**: Subscribe actions for HubSpot cannot notify of custom field changes
- **Source**: Documentation
- **Severity**: Not a validation error, but a documented limitation
- **Impact**: Custom fields in HubSpot objects won't trigger subscribe update events

---

### 3.3 Provider Capability Rules

These rules validate that the provider supports the actions being used.

#### Rule: Provider must support read action
- **Severity**: ERROR
- **Description**: If `read` is specified, the provider must have `Support.Read = true` in the catalog
- **Source**: Provider catalog `ProviderInfo.Support.Read` field
- **Access via**: `providers.ReadCatalog()` or `catalog.ReadLatestInfo()`
- **Example**: If a provider doesn't support read operations, configuring a read action will fail validation

#### Rule: Provider must support write action
- **Severity**: ERROR
- **Description**: If `write` is specified, the provider must have `Support.Write = true` in the catalog
- **Source**: Provider catalog `ProviderInfo.Support.Write` field

#### Rule: Provider must support subscribe action
- **Severity**: ERROR
- **Description**: If `subscribe` is specified, the provider must have `Support.Subscribe = true` in the catalog
- **Source**: Provider catalog `ProviderInfo.Support.Subscribe` field
- **Implementation**: Check provider capabilities via `providers.ReadCatalog()` or `catalog.ReadLatestInfo()`. Module overrides may alter base provider support (check `ModuleInfo.Support` if module is specified).
- **Example violation**:
```yaml
integrations:
  - name: my-integration
    provider: some-provider  # Provider without subscribe support
    subscribe:  # Will fail validation
      objects:
        - objectName: Customer
          destination: my-db
```

#### Rule: Provider must support proxy action
- **Severity**: ERROR
- **Description**: If `proxy` is specified, the provider must have `Support.Proxy = true` in the catalog
- **Source**: Provider catalog `ProviderInfo.Support.Proxy` field

#### Rule: Provider must support bulk write operations
- **Severity**: ERROR
- **Description**: If bulk write operations (insert/update/delete/upsert) are used, the provider must support them
- **Source**: Provider catalog `ProviderInfo.Support.BulkWrite` structure
- **Fields to check**:
  - `Support.BulkWrite.Insert` (bool)
  - `Support.BulkWrite.Update` (bool)
  - `Support.BulkWrite.Delete` (bool)
  - `Support.BulkWrite.Upsert` (bool)

---

### 3.4 Module-Specific Rules

#### Rule: Provider must support specified module
- **Severity**: ERROR
- **Description**: If an integration specifies a `module`, the provider must list that module in its catalog
- **Source**: Provider catalog `ProviderInfo.Modules` map
- **Note**: Module support can override base provider support via `ModuleInfo.Support`
- **Example violation**:
```yaml
integrations:
  - name: my-integration
    provider: salesforce
    module: marketing-cloud  # If Salesforce doesn't support this module
```
- **Example valid**:
```yaml
integrations:
  - name: my-integration
    provider: hubspot
    module: crm  # HubSpot supports CRM module
```

---

### 3.5 Google Calendar Rules

#### Rule: Google Calendar events object cannot use fullHistory backfill
- **Severity**: ERROR
- **Rule ID**: `google-calendar-no-full-history`
- **Provider**: `googlecalendar`
- **Object**: `events`
- **Source**: `validator/provider_google_calendar.go`
- **Error constant**: `ErrGoogleCalendarFullHistory`
- **Description**: The Google Calendar `events` object cannot use `fullHistory: true` in backfill configuration
- **Rationale**: Google Calendar API does not support fetching full event history; it only allows date-range queries
- **Example violation**:
```yaml
integrations:
  - provider: googlecalendar
    read:
      objects:
        - objectName: events
          destination: webhook
          schedule: "0 */12 * * *"
          backfill:
            defaultPeriod:
              fullHistory: true  # ERROR for Google Calendar events
```
- **Example valid**:
```yaml
integrations:
  - provider: googlecalendar
    read:
      objects:
        - objectName: events
          destination: webhook
          schedule: "0 */12 * * *"
          backfill:
            defaultPeriod:
              days: 28  # Use days-based backfill instead
```

#### Rule: Google Calendar events backfill maximum 28 days
- **Severity**: ERROR
- **Rule ID**: `google-calendar-max-backfill`
- **Provider**: `googlecalendar`
- **Object**: `events`
- **Source**: `validator/provider_google_calendar.go`
- **Error constant**: `ErrGoogleCalendarMaxBackfill`
- **Constant**: `MaxGoogleCalendarBackfillDays = 28`
- **Description**: The Google Calendar `events` object backfill period cannot exceed 28 days
- **Rationale**: Google Calendar API performance and rate limits make longer backfills impractical; 28 days is the recommended maximum
- **Example violation**:
```yaml
integrations:
  - provider: googlecalendar
    read:
      objects:
        - objectName: events
          destination: webhook
          schedule: "0 */12 * * *"
          backfill:
            defaultPeriod:
              days: 30  # Exceeds 28-day limit - ERROR
```
- **Example valid**:
```yaml
integrations:
  - provider: googlecalendar
    read:
      objects:
        - objectName: events
          destination: webhook
          schedule: "0 */12 * * *"
          backfill:
            defaultPeriod:
              days: 28  # Maximum allowed
```

**Note**: These restrictions only apply to the `events` object. Other Google Calendar objects (if any) are not subject to these limits.

---

### 3.6 Snowflake Rules

#### Rule: Snowflake only supports fullHistory backfill
- **Severity**: ERROR
- **Rule ID**: `snowflake-only-full-history`
- **Provider**: `snowflake`
- **Source**: `validator/provider_snowflake.go`
- **Error constant**: `ErrSnowflakeBackfillDays`
- **Description**: Snowflake integrations must use `fullHistory: true` for backfill; days-based backfill is not supported
- **Rationale**: Snowflake's data architecture and querying model require full table scans; partial date-range backfills are not efficient or supported
- **Example violation**:
```yaml
integrations:
  - provider: snowflake
    read:
      objects:
        - objectName: CUSTOMERS
          destination: webhook
          schedule: "0 */12 * * *"
          backfill:
            defaultPeriod:
              days: 30  # ERROR - Snowflake requires fullHistory
```
- **Example valid**:
```yaml
integrations:
  - provider: snowflake
    read:
      objects:
        - objectName: CUSTOMERS
          destination: webhook
          schedule: "0 */12 * * *"
          backfill:
            defaultPeriod:
              fullHistory: true  # Required for Snowflake
```

---

## 4. Async Error Prevention Rules (Warnings)

These rules catch scenarios that would fail at runtime in asynchronous services (Temporal workflows, messenger services). They are warnings because they often can't be fully validated statically.

### 4.1 Destination Resolution Rules

#### Rule: Destination names should exist in project configuration
- **Severity**: WARNING
- **Source**: `server/shared/dbservice/destination.go:27`
- **Runtime error**: `errDestinationNotFound`
- **Description**: Destination names referenced in `read`, `write`, and `subscribe` objects should correspond to actual destinations configured in the project
- **Note**: Static validation can only check that a destination name is specified, not that it exists in the database. Full validation requires runtime context.
- **Suggestion**: "Ensure destination '{name}' is configured in your project before deploying this integration"
- **Example**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-warehouse  # Warning: verify this destination exists
```

---

### 4.2 Object Existence Rules

#### Rule: Objects should exist in provider's object catalog
- **Severity**: WARNING (when catalog unavailable) / ERROR (when object not found in catalog)
- **Source**: `amp-yaml-validator/validator/read.go`, `write.go`, `subscribe.go` (`validateObjectName` function)
- **Rule ID**: `RuleObjectExists`
- **Runtime error**: `ErrMissingObjectInRevision`
- **Description**: Object names referenced in amp.yaml should exist in the provider's object catalog
- **Implementation**: The validator calls `catalog.ListObjects(provider, module)` to retrieve the list of valid objects. If the catalog returns `ErrNotSupported`, a **WARNING** is issued with `RuleCatalogAccess`. If the catalog returns a list and the object is not found, an **ERROR** is issued with `RuleObjectExists`.
- **Graceful degradation**: Since the connectors package currently does not expose object schemas, the `ListObjects` method returns `ErrNotSupported`, causing validation to issue a warning and continue rather than blocking deployment.
- **Future enhancement**: When object schemas become available in the catalog, this validation will enforce object existence as an error.
- **Suggestion**: "Verify that object '{objectName}' exists in {provider}'s API documentation" (warning) or "Object '{objectName}' is not supported by provider '{provider}' module '{module}'" (error)
- **Example**:
```yaml
read:
  objects:
    - objectName: CustomObject__c
      destination: my-db
      # Warning: Object validation skipped (catalog does not provide object list)
      # Future: Error if object not found in catalog
```

---

### 4.3 Inactive Destination Rules

#### Rule: Destinations should be active at runtime
- **Severity**: WARNING
- **Source**: `server/shared/workflow/read/workflow.go:102-117`
- **Runtime error**: `errInactiveDestination`
- **Description**: If a destination is marked inactive at runtime, syncs will be skipped
- **Note**: Destination active status can only be verified at runtime against the database
- **Suggestion**: "Ensure all destinations are active before enabling this integration"

---

### 4.4 Workflow Timeout Risks

#### Rule: Large backfills may cause workflow timeouts
- **Severity**: WARNING
- **Description**: Backfilling many days or using `fullHistory: true` with large datasets can cause workflow timeouts
- **Source**: Workflow timeout patterns in `server/shared/workflow/read/workflow.go`
- **Suggestion**: "Consider reducing backfill period or testing with a smaller time range first"
- **Example**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/15 * * * *"
      backfill:
        defaultPeriod:
          days: 365  # Warning: 1 year backfill may timeout with large datasets
```

#### Rule: Very frequent schedules increase throttling risk
- **Severity**: WARNING
- **Description**: Schedules near the minimum frequency (*/10) may hit rate limits with high-volume objects
- **Suggestion**: "Monitor API rate limits when using frequent sync schedules"
- **Example**:
```yaml
read:
  objects:
    - objectName: Account
      destination: my-db
      schedule: "*/10 * * * *"  # Warning: very frequent sync
```

---

### 4.5 Subscribe Webhook Processing Risks

These scenarios can cause webhook processing failures at runtime in the messenger service:

#### Risk: Webhook signature verification failures
- **Severity**: WARNING
- **Source**: `verifyWebhook` function in `server/messenger/subscribeEventProcessor.go` (lines 1069-1139 as reference hint)
- **Runtime error**: Webhook verification returns `false`, event is dropped
- **Description**: Provider webhook signature verification may fail if provider app credentials are misconfigured, expired, or if the webhook payload is malformed
- **Mitigation**: Validate that provider app configuration is complete before enabling subscribe
- **Example scenarios**:
  - HubSpot: Missing or incorrect app client secret for signature validation
  - Salesforce: AWS EventBridge configuration issues (though Salesforce currently bypasses verification)
  - Zoho: Invalid client secret for webhook signature verification

#### Risk: Unsupported provider for webhook verification
- **Severity**: WARNING
- **Source**: `getWebhookVerifiableConnector` function in `server/messenger/subscribeEventProcessor.go` (lines 1154-1166 as reference hint)
- **Runtime error**: `errWebhookVerificationNotSupported`
- **Description**: Not all providers support webhook verification. Currently supported: HubSpot, Salesforce, Zoho
- **Mitigation**: Ensure provider is one of the supported webhook providers
- **Suggestion**: "Verify that provider supports webhook subscriptions and verification"

#### Risk: JSON shape mismatches (array vs object)
- **Severity**: WARNING
- **Source**: `parseWebhookMessages` function in `server/messenger/subscribeEventProcessor.go` (lines 1583-1637 as reference hint)
- **Runtime error**: `determineJSONType` fails or unmarshal errors
- **Description**: Webhook payloads may arrive as either JSON objects or arrays depending on the provider. Parser expects specific shapes:
  - **Object type**: Salesforce (requires unwrapping from EventBridge envelope), Zoho (direct collapsed event)
  - **Array type**: HubSpot (array of subscription events)
- **Mitigation**: Ensure provider app webhooks are configured correctly
- **Example failures**:
  - Salesforce event missing `detail.payload` structure (lines 1509-1531)
  - HubSpot sending malformed event array
  - Zoho event with unexpected structure

#### Risk: Provider-specific event unwrapping failures
- **Severity**: WARNING
- **Source**: `unwrapSalesforceEvent` function in `server/messenger/subscribeEventProcessor.go` (lines 1504-1531 as reference hint)
- **Runtime error**: `errSalesforceUnwrap`
- **Description**: Salesforce events arrive through AWS EventBridge and require unwrapping from the EventBridge envelope structure. Failures occur if:
  - `detail` field is missing or not a map
  - `payload` field is missing within detail or not a map
- **Mitigation**: Verify AWS EventBridge integration is properly configured for Salesforce CDC
- **Example error messages**:
  - "failed to unwrap salesforce event: detail field not found"
  - "failed to unwrap salesforce event: detail field is not a map"
  - "failed to unwrap salesforce event: payload field not found"

#### Risk: Event filtering and field matching failures
- **Severity**: WARNING
- **Source**: `filterPassThrough` function in `server/messenger/subscribeEventProcessor.go` (lines 1250-1502 as reference hint)
- **Description**: Subscribe update events may be dropped if:
  - Updated fields don't match any watch fields specified in `updateEvent.requiredWatchFields`
  - Updated fields don't match any fields selected in the read configuration when using `watchFieldsAuto: "selected"`
  - Object name from webhook doesn't match any object configured in the installation
- **Mitigation**: Ensure `updateEvent` watch field configuration aligns with actual provider field names
- **Suggestion**: "Verify that watch fields in subscribe configuration match actual provider field names"

---

## 5. Field-Level Validation Rules

These rules validate individual field types and structures within the manifest.

### 5.1 Required Field Rules

#### Rule: IntegrationFieldExistent must have fieldName
- **Severity**: ERROR
- **Source**: OpenAPI schema `openapi/manifest/manifest.yaml`
- **Description**: When using `IntegrationFieldExistent` (field reference without mapping), `fieldName` is required
- **Example violation**:
```yaml
selectedFields:
  - {}  # Missing fieldName
```
- **Example valid**:
```yaml
selectedFields:
  - fieldName: Id
  - fieldName: Name
```

#### Rule: IntegrationFieldMapping must have mapToName
- **Severity**: ERROR
- **Source**: OpenAPI schema `openapi/manifest/manifest.yaml`
- **Description**: When using `IntegrationFieldMapping` (field with custom mapping), both `fieldName` and `mapToName` are required
- **Example violation**:
```yaml
selectedFieldMappings:
  - fieldName: FirstName  # Missing mapToName
```
- **Example valid**:
```yaml
selectedFieldMappings:
  - fieldName: FirstName
    mapToName: first_name
```

---

### 5.2 Field Type Rules

#### Rule: IntegrationField is oneOf (Existent or Mapping)
- **Severity**: ERROR
- **Source**: OpenAPI schema
- **Description**: Each field specification must be either `IntegrationFieldExistent` (fieldName only) OR `IntegrationFieldMapping` (fieldName + mapToName), not a mix
- **Example violation**:
```yaml
# This is handled at schema level - invalid YAML structure would fail parsing
```

#### Rule: Cannot mix fieldName with mapToName in requiredFields for always-enabled objects
- **Severity**: ERROR
- **Source**: `validateReadContent` function in `server/shared/common/validate.go` (lines 100-102 as reference hint)
- **Description**: For always-enabled objects, `requiredFields` must use fieldName only (no mapToName)
- **See**: Section 2.11 - Always-Enabled Object Rules

---

### 5.3 Watch Field Rules

#### Rule: UpdateEvent watch fields must be properly specified
- **Severity**: ERROR
- **Description**: UpdateEvent watch fields must be either:
  - `requiredWatchFields`: A list of specific field names
  - `watchFieldsAuto`: Either `"all"` or `"selected"`
- **See**: Section 2.6 - Subscribe Update Event Rules

#### Rule: watchFieldsAuto must be "all" or "selected"
- **Severity**: ERROR
- **Source**: OpenAPI schema enum constraint
- **Description**: If `watchFieldsAuto` is specified, it must be one of the allowed enum values
- **Example violation**:
```yaml
updateEvent:
  enabled: "always"
  watchFieldsAuto: "some"  # Invalid enum value
```
- **Example valid**:
```yaml
updateEvent:
  enabled: "always"
  watchFieldsAuto: "all"
```

---

## 6. Special Cases and Edge Cases

### 6.1 Project-Specific Exceptions

#### Note: Schedule frequency exceptions for specific projects
- **Source**: `server/api/routes/api/installation.go:1415-1422`
- **Description**: Some projects (e.g., "11x Alice") have `ErrScheduleTooFrequent` suppressed in production
- **Recommendation**: Document this exception but don't implement in the general-purpose validator
- **Rationale**: Business exceptions are handled at the API/runtime layer, not in static validation

---

### 6.2 Permission Error Handling

#### Note: Subscribe permission errors are non-retryable
- **Source**: `server/shared/temporal/subscribeinstallation/subscribeInstallation.go`
- **Description**: If a provider returns `INSUFFICIENT_ACCESS` for subscribe operations, this is treated as a non-retryable error
- **Impact**: Installations will fail immediately rather than retrying
- **Recommendation**: Document that users should verify permissions before enabling subscribe

---

### 6.3 Provider Subscribe Support

#### Note: Provider support for subscribe varies
- **Validation approach**: Check `ProviderInfo.Support.Subscribe` flag via the connectors catalog
- **Source**: Provider catalog accessed through `providers.ReadCatalog()` or `catalog.ReadLatestInfo()`
- **Module overrides**: Module-specific support may differ from base provider support (check `ModuleInfo.Support.Subscribe` if module is specified)
- **Recommendation**: Always validate against the catalog rather than maintaining static lists of supported providers (see section 3.3)

---

## 7. Validation Rule Categories for Implementation

This section organizes rules by implementation category to guide library development.

### Category A: Schema Validation (Structural)
These can be validated through JSON Schema or struct tag validation:
- Required fields present (specVersion, name, provider, objectName, destination, schedule)
- Field types correct (strings, booleans, arrays, objects)
- Enum values valid (delivery.mode, watchFieldsAuto, updateEvent.enabled)
- Array/object structure correct (non-empty lists, proper nesting)

**Implementation approach**: Use OpenAPI schema validation + go-playground/validator

### Category B: Business Logic Validation (Semantic)
These require custom validation logic:
- Schedule frequency constraints (minimum 10 minutes, no `*` in minute field)
- Backfill mutual exclusivity (days XOR fullHistory)
- Field mapping uniqueness (no duplicate mapToName)
- Always-enabled object constraints (requiredFields, schedule, field count)
- Subscribe/read co-requirements (subscribe requires read)
- UpdateEvent watch field requirements

**Implementation approach**: Custom validator functions per rule

### Category C: Provider-Specific Validation
These require provider catalog access:
- Provider capability checks (read/write/subscribe/proxy support)
- Provider-specific limits (Salesforce max 5 subscribe objects)
- Module validation (module exists in provider catalog)
- HubSpot scopes requirement

**Implementation approach**: Catalog-based validators with graceful fallback

### Category D: Cross-Reference Validation (Warnings)
These can only be partially validated statically:
- Destination name references (warn if name format suspicious)
- Object existence in provider (warn if object name looks invalid)
- Field name validity (requires provider metadata)

**Implementation approach**: Heuristic-based warnings with clear suggestions

---

## 8. Line Number Tracking Requirements

### Requirements for Error Reporting

For each validation error or warning, the validator must provide:

1. **Line number**: The line in the YAML file where the issue occurs (1-based indexing)
2. **Column number**: Optional but recommended for precise error location (1-based indexing)
3. **YAML path**: JSONPath-style reference to the field (e.g., `$.integrations[0].read.objects[1].schedule`)
4. **Field name or value**: The specific field or value that caused the issue

### Implementation Approach

**Step 1: Parse with position tracking**
```go
// Use gopkg.in/yaml.v3 to parse into yaml.Node
var rootNode yaml.Node
yaml.Unmarshal(yamlBytes, &rootNode)

// Walk the node tree to build position map
type Position struct {
    Line   int
    Column int
}
positionMap := make(map[string]Position)  // YAML path -> position
buildPositionMap(&rootNode, "$", positionMap)
```

**Step 2: Parse into typed structs**
```go
// Parse again into openapi.Manifest for validation
var manifest openapi.Manifest
yaml.Unmarshal(yamlBytes, &manifest)
```

**Step 3: Validate and attach positions**
```go
// During validation, use position map to attach line numbers
for i, obj := range manifest.Integrations[0].Read.Objects {
    path := fmt.Sprintf("$.integrations[0].read.objects[%d].schedule", i)
    if err := validateSchedule(obj.Schedule); err != nil {
        pos := positionMap[path]
        issues = append(issues, ValidationIssue{
            Severity: "error",
            Message:  err.Error(),
            Line:     pos.Line,
            Column:   pos.Column,
            Path:     path,
        })
    }
}
```

### Reference Implementation

**Pattern reference**: `cli/files/path_builder.go` demonstrates path tracking approach:
- Uses `pathTracker` struct to maintain current YAML path during tree traversal
- Tracks path components as it descends into nested structures
- Similar approach can be used with `yaml.Node` to build position map

---

## 9. Error and Warning Output Format

### ValidationIssue Structure

```go
type ValidationIssue struct {
    Severity   string  `json:"severity"`    // "error" or "warning"
    Message    string  `json:"message"`     // Human-readable error message
    Line       int     `json:"line"`        // Line number in YAML file (1-based)
    Column     int     `json:"column"`      // Column number (1-based, 0 if unavailable)
    Path       string  `json:"path"`        // YAML path (e.g., "$.integrations[0].read.objects[1].schedule")
    Rule       string  `json:"rule"`        // Rule identifier (e.g., "schedule-too-frequent")
    Suggestion string  `json:"suggestion"`  // Optional suggestion for fixing (especially for warnings)
}
```

### ValidationResult Structure

```go
type ValidationResult struct {
    Valid    bool               `json:"valid"`     // false if any errors present
    Errors   []ValidationIssue  `json:"errors"`    // Blocking errors
    Warnings []ValidationIssue  `json:"warnings"`  // Non-blocking warnings
}
```

### Example Output

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
      "rule": "schedule-too-frequent",
      "suggestion": "Change schedule to */10 or greater (e.g., '*/10 * * * *' for every 10 minutes)"
    },
    {
      "severity": "error",
      "message": "Subscribe action requires read action to be defined",
      "line": 28,
      "column": 5,
      "path": "$.integrations[0].subscribe",
      "rule": "subscribe-requires-read",
      "suggestion": "Add a read section to this integration"
    }
  ],
  "warnings": [
    {
      "severity": "warning",
      "message": "Large backfill period may cause workflow timeouts",
      "line": 15,
      "column": 14,
      "path": "$.integrations[0].read.objects[0].backfill.defaultPeriod.days",
      "rule": "large-backfill-risk",
      "suggestion": "Consider reducing backfill period to 90 days or less for initial sync"
    }
  ]
}
```

### Human-Readable Text Format

```
Validation failed with 2 errors and 1 warning.

ERRORS:
  Line 12, Column 18 ($.integrations[0].read.objects[0].schedule)
    Schedule interval must be at least 10 minutes. Found: */5 * * * * (5 minutes)
    Suggestion: Change schedule to */10 or greater (e.g., '*/10 * * * *' for every 10 minutes)

  Line 28, Column 5 ($.integrations[0].subscribe)
    Subscribe action requires read action to be defined
    Suggestion: Add a read section to this integration

WARNINGS:
  Line 15, Column 14 ($.integrations[0].read.objects[0].backfill.defaultPeriod.days)
    Large backfill period may cause workflow timeouts
    Suggestion: Consider reducing backfill period to 90 days or less for initial sync
```

---

## 10. Validation Rules Index

This machine-readable index maps all validation rules to their source implementations and provides examples.

### Table Columns
- **Rule ID**: Unique identifier for the rule
- **Description**: What the rule validates
- **Severity**: ERROR (blocks deployment) or WARNING (potential runtime issue)
- **Category**: Universal, Provider, Async, or Field
- **Source Symbol/Constant**: Function name or constant in source code
- **YAML Path**: JSONPath to the validated field (examples)
- **Error Constant**: Go error constant name
- **Example Invalid**: Brief example of violation
- **Example Valid**: Brief example of correct usage

### Rules Index Table

| Rule ID | Description | Severity | Category | Source Symbol/Constant | YAML Path | Error Constant | Example Invalid | Example Valid |
|---------|-------------|----------|----------|------------------------|-----------|----------------|-----------------|---------------|
| spec-version | specVersion must be "1.0.0" | ERROR | Universal | `ValidateRevision` | `$.specVersion` | `errInvalidSpecVersion` | `specVersion: "2.0.0"` | `specVersion: "1.0.0"` |
| integration-action-required | Integration must have at least one action | ERROR | Universal | manifest.go check | `$.integrations[*]` | - | integration with no actions | integration with `read:` |
| integration-name-required | Integration must have name | ERROR | Universal | OpenAPI schema | `$.integrations[*].name` | - | missing `name` field | `name: "my-integration"` |
| integration-provider-required | Integration must have provider | ERROR | Universal | OpenAPI schema | `$.integrations[*].provider` | - | missing `provider` field | `provider: "salesforce"` |
| read-objects-required | read.objects must be non-empty list | ERROR | Universal | `validateReadContent` | `$.integrations[*].read.objects` | `errMissingReadObjects` | `objects: []` | `objects: [{objectName: "Account"}]` |
| read-object-name-required | Each read object must have objectName | ERROR | Universal | OpenAPI schema | `$.integrations[*].read.objects[*].objectName` | - | missing objectName | `objectName: "Account"` |
| read-object-destination-required | Each read object must have destination | ERROR | Universal | OpenAPI schema | `$.integrations[*].read.objects[*].destination` | - | missing destination | `destination: "postgres-db"` |
| read-object-schedule-required | Each read object must have schedule | ERROR | Universal | OpenAPI schema | `$.integrations[*].read.objects[*].schedule` | - | missing schedule | `schedule: "*/15 * * * *"` |
| write-objects-required | write.objects must be non-empty list | ERROR | Universal | `ValidateRevision` | `$.integrations[*].write.objects` | `errMissingWriteObjects` | `objects: []` | `objects: [{objectName: "Contact"}]` |
| write-object-name-required | Each write object must have objectName | ERROR | Universal | OpenAPI schema | `$.integrations[*].write.objects[*].objectName` | - | missing objectName | `objectName: "Contact"` |
| subscribe-requires-read | Subscribe requires read to be defined | ERROR | Universal | `validateSubscribeContent` | `$.integrations[*].subscribe` | `ErrSubscribeRequiresRead` | subscribe without read | subscribe with read defined |
| subscribe-objects-required | subscribe.objects must be non-empty list | ERROR | Universal | `validateSubscribeContent` | `$.integrations[*].subscribe.objects` | `errMissingSubscribeObjects` | `objects: []` | `objects: [{objectName: "Account"}]` |
| subscribe-inherit-fields-required | inheritFieldsAndMapping must be true (v1) | ERROR | Universal | `validateSubscribeContent` | `$.integrations[*].subscribe.objects[*].inheritFieldsAndMapping` | `ErrSubscribeInheritFieldsAndMapping` | `inheritFieldsAndMapping: false` | `inheritFieldsAndMapping: true` |
| subscribe-object-name-required | Each subscribe object must have objectName | ERROR | Universal | OpenAPI schema | `$.integrations[*].subscribe.objects[*].objectName` | - | missing objectName | `objectName: "Account"` |
| subscribe-object-destination-required | Each subscribe object must have destination | ERROR | Universal | OpenAPI schema | `$.integrations[*].subscribe.objects[*].destination` | - | missing destination | `destination: "postgres-db"` |
| update-event-enabled-value | updateEvent.enabled must be "always" if set | ERROR | Universal | `validateSubscribeContent` | `$.integrations[*].subscribe.objects[*].updateEvent.enabled` | `ErrInvalidInputEnabled` | `enabled: "sometimes"` | `enabled: "always"` |
| update-event-watch-fields-required | UpdateEvent must specify watch fields | ERROR | Universal | `validateSubscribeContent` | `$.integrations[*].subscribe.objects[*].updateEvent` | `ErrWatchFieldsRequired` | neither watchFieldsAuto nor requiredWatchFields | `watchFieldsAuto: "all"` |
| update-event-watch-fields-exclusive | Cannot use both watchFieldsAuto and requiredWatchFields | ERROR | Universal | `validateSubscribeContent` | `$.integrations[*].subscribe.objects[*].updateEvent` | `ErrWatchFieldsAndRequiredWatchFields` | both watchFieldsAuto and requiredWatchFields set | only one watch field config |
| schedule-valid-cron | Schedule must be valid cron (5 parts) | ERROR | Universal | `validateSchedule` | `$.integrations[*].read.objects[*].schedule` | `errInvalidCronSchedule` | `schedule: "invalid"` | `schedule: "*/15 * * * *"` |
| schedule-no-asterisk-minute | Schedule cannot have '*' in minute field | ERROR | Universal | `validateSchedule` | `$.integrations[*].read.objects[*].schedule` | `ErrScheduleTooFrequent` | `schedule: "* * * * *"` | `schedule: "*/10 * * * *"` |
| schedule-minimum-interval | Schedule minimum interval 10 minutes (prod) | ERROR | Universal | `validateSchedule` | `$.integrations[*].read.objects[*].schedule` | `ErrScheduleTooFrequent` | `schedule: "*/5 * * * *"` | `schedule: "*/10 * * * *"` |
| schedule-tick-calculation | Schedule must have valid tick calculation | ERROR | Universal | `validateSchedule` | `$.integrations[*].read.objects[*].schedule` | `ErrScheduleTooFrequent` | invalid tick calc | valid cron schedule |
| delivery-auto-no-pagesize | delivery.mode=auto cannot set pageSize | ERROR | Universal | `validateDeliveryMode` | `$.integrations[*].read.delivery` | `ErrInvalidPageSizeConfig` | auto mode with pageSize | auto mode without pageSize |
| delivery-onrequest-pagesize-required | delivery.mode=onRequest requires pageSize | ERROR | Universal | `validateDeliveryMode` | `$.integrations[*].read.delivery.pageSize` | `ErrInvalidPageSizeConfig` | onRequest without pageSize | onRequest with pageSize |
| delivery-onrequest-pagesize-range | pageSize must be 50-500 for onRequest | ERROR | Universal | `validateDeliveryMode` | `$.integrations[*].read.delivery.pageSize` | `ErrInvalidPageSizeConfig` | `pageSize: 25` | `pageSize: 100` |
| delivery-mode-enum | delivery.mode must be "auto" or "onRequest" | ERROR | Universal | `validateDeliveryMode` | `$.integrations[*].read.delivery.mode` | `ErrInvalidDeliveryConfig` | `mode: streaming` | `mode: auto` |
| backfill-struct-valid | Backfill must pass struct validation | ERROR | Universal | `ValidateBackfillFromConfig` | `$.integrations[*].read.objects[*].backfill` | `errInvalidBackfill` | invalid backfill config | valid backfill config |
| backfill-period-exclusive | Must provide only days OR fullHistory, not both | ERROR | Universal | `ValidateBackfillFromConfig` | `$.integrations[*].read.objects[*].backfill.defaultPeriod` | `errInvalidBackfill` | both days and fullHistory set | only one set |
| field-mapping-unique | No duplicate field mappings (mapToName) | ERROR | Universal | `ValidateNoDuplicateFieldMappings` | `$.integrations[*].read.objects[*].selectedFieldMappings` | `errDuplicateFieldMapping` | two fields map to same mapToName | all mapToName values unique |
| required-mappings-selected | Required mappings must be in selectedFieldMappings | ERROR | Universal | `ValidateFields` | `$.integrations[*].read.objects[*].selectedFieldMappings` | `ErrMissingMinimumRequiredFields` | required mapping missing | all required mappings present |
| always-enabled-required-fields | Always-enabled objects need requiredFields | ERROR | Universal | `validateReadContent` | `$.integrations[*].read.objects[*].requiredFields` | `ErrMissingMinimumRequiredFields` | alwaysEnabled without requiredFields | alwaysEnabled with requiredFields |
| always-enabled-no-maptoname | Always-enabled requiredFields cannot use mapToName | ERROR | Universal | `validateReadContent` | `$.integrations[*].read.objects[*].requiredFields` | `ErrMissingMinimumRequiredFields` | requiredField with mapToName | requiredField without mapToName |
| always-enabled-schedule-required | Always-enabled objects must have schedule | ERROR | Universal | `validateReadContent` | `$.integrations[*].read.objects[*].schedule` | `ErrInvalidSchedule` | alwaysEnabled without schedule | alwaysEnabled with schedule |
| always-enabled-mapping-limit | Required mappings ≤ builder mappings | ERROR | Universal | `ValidateFields` | `$.integrations[*].read.objects[*]` | `ErrAlwaysEnabledObjectHasMoreRequiredMappingsThanBuilderMappings` | more required than builder | required ≤ builder |
| minimum-fields-required | Each read object needs ≥1 field selected | ERROR | Universal | `ValidateFields` | `$.integrations[*].read.objects[*]` | `ErrMissingMinimumRequiredFields` | no fields selected | at least one field selected |
| unknown-key | Key is not part of the amp.yaml schema (orphan/typo) | WARNING | Universal | `validateUnknownKeys` / `parser.DetectUnknownKeys` | `$.integrations[*].read.objects[*].<key>` | `RuleUnknownKey` | `scheduel: "*/10 * * * *"` | `schedule: "*/10 * * * *"` |
| salesforce-subscribe-limit | Salesforce max 5 subscribe objects | ERROR | Provider | `validateSubscribeContent` (maxSalesforceSubscribeObjects=5) | `$.integrations[*].subscribe.objects` | `ErrTooManySubscribeObjects` | 6+ subscribe objects for Salesforce | ≤5 subscribe objects |
| hubspot-scopes-required | HubSpot provider apps need scopes | ERROR | Provider | providerApp.go:58-62 | - |
| provider-read-support | Provider must support read action | ERROR | Provider | Catalog Support.Read | - |
| provider-write-support | Provider must support write action | ERROR | Provider | Catalog Support.Write | - |
| provider-subscribe-support | Provider must support subscribe action | ERROR | Provider | Catalog Support.Subscribe | - |
| provider-proxy-support | Provider must support proxy action | ERROR | Provider | Catalog Support.Proxy | - |
| provider-bulk-write-support | Provider must support bulk write operations | ERROR | Provider | Catalog Support.BulkWrite | - |
| provider-module-support | Provider must support specified module | ERROR | Provider | Catalog Modules | - |
| destination-exists | Destination should exist in project | WARNING | Async | destination.go:27 | errDestinationNotFound |
| object-exists | Object should exist in provider catalog | WARNING | Async | integration.go:12 | ErrMissingObjectInRevision |
| destination-active | Destination should be active | WARNING | Async | workflow.go:102-117 | errInactiveDestination |
| large-backfill-risk | Large backfills may timeout | WARNING | Async | workflow patterns | - |
| frequent-schedule-risk | Very frequent schedules increase throttling risk | WARNING | Async | workflow patterns | - |
| field-existent-name-required | IntegrationFieldExistent needs fieldName | ERROR | Field | OpenAPI schema | - |
| field-mapping-names-required | IntegrationFieldMapping needs fieldName and mapToName | ERROR | Field | OpenAPI schema | - |
| field-type-oneof | Field must be Existent OR Mapping | ERROR | Field | OpenAPI schema | - |
| watch-fields-auto-enum | watchFieldsAuto must be "all" or "selected" | ERROR | Field | OpenAPI schema | - |

### Example YAML Snippets

#### Valid minimal integration
```yaml
specVersion: "1.0.0"
integrations:
  - name: salesforce-sync
    provider: salesforce
    read:
      objects:
        - objectName: Account
          destination: postgres-db
          schedule: "*/15 * * * *"
          selectedFields:
            - fieldName: Id
            - fieldName: Name
```

#### Invalid: Subscribe without read
```yaml
specVersion: "1.0.0"
integrations:
  - name: salesforce-sync
    provider: salesforce
    subscribe:  # ERROR: subscribe-requires-read
      objects:
        - objectName: Account
          destination: postgres-db
          inheritFieldsAndMapping: true
```

#### Invalid: Schedule too frequent
```yaml
specVersion: "1.0.0"
integrations:
  - name: salesforce-sync
    provider: salesforce
    read:
      objects:
        - objectName: Account
          destination: postgres-db
          schedule: "*/5 * * * *"  # ERROR: schedule-minimum-interval (< 10 min)
```

#### Invalid: Duplicate field mappings
```yaml
specVersion: "1.0.0"
integrations:
  - name: salesforce-sync
    provider: salesforce
    read:
      objects:
        - objectName: Contact
          destination: postgres-db
          schedule: "*/15 * * * *"
          selectedFieldMappings:
            - fieldName: FirstName
              mapToName: name
            - fieldName: LastName
              mapToName: name  # ERROR: field-mapping-unique
```

#### Invalid: Salesforce too many subscribe objects
```yaml
specVersion: "1.0.0"
integrations:
  - name: salesforce-sync
    provider: salesforce
    read:
      objects:
        - objectName: Account
          destination: db
          schedule: "*/15 * * * *"
    subscribe:
      objects:
        - objectName: Account
          destination: db
        - objectName: Contact
          destination: db
        - objectName: Opportunity
          destination: db
        - objectName: Lead
          destination: db
        - objectName: Case
          destination: db
        - objectName: Campaign
          destination: db  # ERROR: salesforce-subscribe-limit (> 5)
```

---

## 11. Provider Catalog Integration

### Catalog Access

The validator needs to access provider capabilities from the catalog system to validate provider-specific rules.

#### Import Package
```go
import "github.com/amp-labs/connectors/providers"
```

#### Read Catalog
```go
// Option 1: Read full catalog
catalog, err := providers.ReadCatalog()
if err != nil {
    // Handle error - may want to skip provider-specific validation
    return warnings.Add("Unable to access provider catalog, skipping provider capability checks")
}

// Option 2: Read specific provider (via catalog package)
import "github.com/amp-labs/server/shared/catalog"
providerInfo, err := catalog.ReadLatestInfo(ctx, providerName)
```

#### Check Provider Capabilities
```go
providerInfo := catalog[providerName]
if providerInfo == nil {
    return ValidationIssue{
        Severity: "error",
        Message:  fmt.Sprintf("Unknown provider: %s", providerName),
        Rule:     "provider-exists",
    }
}

// Check action support
if manifest.Read != nil && !providerInfo.Support.Read {
    issues = append(issues, ValidationIssue{
        Severity: "error",
        Message:  fmt.Sprintf("Provider %s does not support read action", providerName),
        Rule:     "provider-read-support",
    })
}

if manifest.Write != nil && !providerInfo.Support.Write {
    issues = append(issues, ValidationIssue{
        Severity: "error",
        Message:  fmt.Sprintf("Provider %s does not support write action", providerName),
        Rule:     "provider-write-support",
    })
}

if manifest.Subscribe != nil && !providerInfo.Support.Subscribe {
    issues = append(issues, ValidationIssue{
        Severity: "error",
        Message:  fmt.Sprintf("Provider %s does not support subscribe action", providerName),
        Rule:     "provider-subscribe-support",
    })
}

if manifest.Proxy != nil && !providerInfo.Support.Proxy {
    issues = append(issues, ValidationIssue{
        Severity: "error",
        Message:  fmt.Sprintf("Provider %s does not support proxy action", providerName),
        Rule:     "provider-proxy-support",
    })
}
```

#### Check Module Support
```go
if manifest.Module != nil && *manifest.Module != "" {
    moduleInfo, moduleExists := providerInfo.Modules[*manifest.Module]
    if !moduleExists {
        issues = append(issues, ValidationIssue{
            Severity: "error",
            Message:  fmt.Sprintf("Provider %s does not support module: %s", providerName, *manifest.Module),
            Rule:     "provider-module-support",
        })
    }

    // Module can override base provider support
    if moduleExists && moduleInfo.Support != nil {
        // Use moduleInfo.Support instead of providerInfo.Support for capability checks
    }
}
```

#### Check Bulk Write Support
```go
if manifest.Write != nil {
    bulkSupport := providerInfo.Support.BulkWrite

    // Check specific bulk operations if configured
    if needsInsert && !bulkSupport.Insert {
        issues = append(issues, ValidationIssue{
            Severity: "error",
            Message:  fmt.Sprintf("Provider %s does not support bulk insert", providerName),
            Rule:     "provider-bulk-write-support",
        })
    }

    // Similar checks for Update, Delete, Upsert
}
```

### Graceful Degradation

If catalog access fails, the validator should continue with non-provider-specific validation:

```go
func validateWithCatalog(manifest *Manifest) []ValidationIssue {
    var issues []ValidationIssue

    // Always run universal validation
    issues = append(issues, validateUniversal(manifest)...)

    // Try to run provider-specific validation
    catalog, err := providers.ReadCatalog()
    if err != nil {
        issues = append(issues, ValidationIssue{
            Severity:   "warning",
            Message:    "Unable to access provider catalog, skipping provider capability checks",
            Rule:       "catalog-unavailable",
            Suggestion: "Ensure connectors package is available and up to date",
        })
        return issues
    }

    // Run provider-specific validation
    issues = append(issues, validateProviderSpecific(manifest, catalog)...)

    return issues
}
```

---

## 12. Test Coverage Requirements

### Test Categories

#### 1. Schema Validation Tests
Test each structural/type constraint:
- Required fields present
- Field types correct (string, int, bool, array, object)
- Enum values valid
- Array non-empty constraints
- Nested structure validation

**Example test cases**:
- Missing specVersion
- Invalid specVersion value
- Missing integration name
- Missing integration provider
- Empty read.objects array
- Invalid delivery.mode enum value

#### 2. Business Logic Tests
Test each semantic rule:
- Schedule frequency validation
- Backfill mutual exclusivity
- Field mapping uniqueness
- Always-enabled object constraints
- Subscribe/read co-requirements
- UpdateEvent watch fields configuration

**Example test cases**:
- Schedule `*/5 * * * *` (too frequent)
- Schedule `* * * * *` (asterisk in minute field)
- Backfill with both days AND fullHistory
- Duplicate mapToName in field mappings
- Subscribe without read
- UpdateEvent with neither watchFieldsAuto nor requiredWatchFields

#### 3. Provider-Specific Tests
Test provider capability validation:
- Salesforce subscribe object limit
- Provider action support (read/write/subscribe/proxy)
- Module validation
- HubSpot scopes requirement

**Example test cases**:
- Salesforce with 6 subscribe objects (exceeds limit)
- Provider that doesn't support subscribe, but subscribe configured
- Invalid module name for provider
- Provider that doesn't exist in catalog

#### 4. Edge Case Tests
Test boundary conditions and corner cases:
- Exactly 5 subscribe objects for Salesforce (valid)
- Exactly 10-minute schedule interval (minimum valid)
- Empty field lists
- Nil pointer handling
- Very long field names
- Special characters in names

#### 5. Line Number Accuracy Tests
Verify reported line numbers match actual YAML:
- Error on line 1 (specVersion)
- Error on deeply nested field
- Multiple errors with correct line numbers
- Column number accuracy
- YAML path accuracy

**Example test**:
```yaml
# Line 1
specVersion: "1.0.0"
# Line 3
integrations:
  - name: test
    provider: salesforce
    # Line 7
    read:
      objects:
        - objectName: Account
          destination: db
          # Line 12 - this should be reported
          schedule: "*/5 * * * *"  # Too frequent
```
Expected error line: 12

#### 6. Valid Sample Tests
All samples in `/Users/chris/src/samples/` must pass validation:
- `samples/salesforce/amp.yaml`
- `samples/hubspot/amp.yaml`
- `samples/stripe/amp.yaml`
- And all others in the samples directory

**Test implementation**:
```go
func TestValidSamples(t *testing.T) {
    t.Parallel()

    samples, err := filepath.Glob("/Users/chris/src/samples/*/amp.yaml")
    require.NoError(t, err)

    for _, samplePath := range samples {
        samplePath := samplePath
        t.Run(filepath.Base(filepath.Dir(samplePath)), func(t *testing.T) {
            t.Parallel()

            result, err := validator.ValidateFile(samplePath)
            require.NoError(t, err)
            assert.True(t, result.Valid, "Sample %s should be valid", samplePath)
            assert.Empty(t, result.Errors, "Sample %s should have no errors", samplePath)
        })
    }
}
```

#### 7. Invalid Sample Tests
Create intentionally broken YAML files:
- `testdata/invalid/missing-spec-version.yaml`
- `testdata/invalid/invalid-schedule-too-frequent.yaml`
- `testdata/invalid/subscribe-without-read.yaml`
- `testdata/invalid/duplicate-field-mappings.yaml`
- `testdata/invalid/salesforce-too-many-subscribe.yaml`

**Test implementation**:
```go
func TestInvalidSamples(t *testing.T) {
    t.Parallel()

    testCases := []struct {
        file         string
        expectedRule string
        expectedLine int
    }{
        {
            file:         "testdata/invalid/missing-spec-version.yaml",
            expectedRule: "spec-version",
            expectedLine: 1,
        },
        {
            file:         "testdata/invalid/invalid-schedule-too-frequent.yaml",
            expectedRule: "schedule-minimum-interval",
            expectedLine: 12,
        },
        // ... more cases
    }

    for _, tc := range testCases {
        tc := tc
        t.Run(tc.file, func(t *testing.T) {
            t.Parallel()

            result, err := validator.ValidateFile(tc.file)
            require.NoError(t, err)
            assert.False(t, result.Valid)
            require.NotEmpty(t, result.Errors)

            // Check that expected rule is present
            foundRule := false
            foundLine := false
            for _, issue := range result.Errors {
                if issue.Rule == tc.expectedRule {
                    foundRule = true
                    if issue.Line == tc.expectedLine {
                        foundLine = true
                    }
                }
            }
            assert.True(t, foundRule, "Expected rule %s not found", tc.expectedRule)
            assert.True(t, foundLine, "Expected line %d not found", tc.expectedLine)
        })
    }
}
```

### Test Patterns

#### Table-Driven Tests
Reference pattern from `server/shared/common/validate_test.go`:
```go
func TestValidateSchedule(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name          string
        schedule      string
        wantErr       bool
        expectedError error
    }{
        {
            name:          "valid 15 minute schedule",
            schedule:      "*/15 * * * *",
            wantErr:       false,
        },
        {
            name:          "too frequent - 5 minutes",
            schedule:      "*/5 * * * *",
            wantErr:       true,
            expectedError: ErrScheduleTooFrequent,
        },
        {
            name:          "invalid - asterisk in minute",
            schedule:      "* * * * *",
            wantErr:       true,
            expectedError: ErrScheduleTooFrequent,
        },
        {
            name:          "invalid cron syntax",
            schedule:      "not a cron",
            wantErr:       true,
            expectedError: errInvalidCronSchedule,
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            err := validateSchedule(tt.schedule)
            if tt.wantErr {
                require.Error(t, err)
                if tt.expectedError != nil {
                    assert.ErrorIs(t, err, tt.expectedError)
                }
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

#### Parallel Test Execution
```go
func TestSomething(t *testing.T) {
    t.Parallel()  // Mark as parallel at top level

    t.Run("subtest", func(t *testing.T) {
        t.Parallel()  // Also mark subtests as parallel
        // ... test code
    })
}
```

#### Assertion Libraries
Use `testify/require` for assertions that should stop test execution:
```go
require.NoError(t, err)
require.NotNil(t, result)
require.Equal(t, expected, actual)
```

Use `testify/assert` for assertions that should continue:
```go
assert.True(t, result.Valid)
assert.Len(t, result.Errors, 2)
assert.Contains(t, errorMessage, "schedule")
```

### Coverage Goals

- **100% rule coverage**: Every validation rule must have at least one test
- **100% line coverage**: All code paths in validators should be tested
- **Edge case coverage**: Boundary values, empty/nil, special characters
- **Error message testing**: Verify error messages are clear and actionable
- **Line number testing**: Verify reported positions are accurate
- **Sample validation**: All real samples must pass validation

---

## 13. References and Source Code Locations

### Core Validation Logic
- **`server/shared/common/validate.go`**: Main validation rules (schedules, backfill, field mappings, always-enabled objects, subscribe constraints)
- **`server/shared/common/validate_test.go`**: Test patterns and edge cases
- **`server/shared/common/schedule_const.go`**: Production schedule frequency constants
- **`server/shared/common/schedule_const_local.go`**: Local development schedule constants

### Async Error Scenarios
- **`server/shared/workflow/read/workflow.go`**: Read workflow error handling (inactive destinations, missing objects)
- **`server/messenger/subscribeEventProcessor.go`**: Subscribe event processing errors
- **`server/shared/dbservice/destination.go`**: Destination resolution errors
- **`server/shared/dbservice/integration.go`**: Object existence validation
- **`server/shared/temporal/subscribeinstallation/subscribeInstallation.go`**: Subscribe permission errors

### API Layer Validation
- **`server/api/routes/api/installation.go`**: Installation validation (provider matching, schedule exceptions)
- **`server/api/routes/api/integration.go`**: Integration source validation
- **`server/api/routes/api/providerApp.go`**: Provider app validation (HubSpot scopes)
- **`server/api/routes/proxy/proxy.go`**: Proxy request validation

### CLI and Manifest Handling
- **`cli/files/manifest.go`**: CLI manifest parsing and basic validation
- **`cli/files/path_builder.go`**: YAML path tracking pattern (reference for line number implementation)

### Schema and Types
- **`openapi/manifest/manifest.yaml`**: OpenAPI schema definition for amp.yaml structure
- **`connectors/providers/types.gen.go`**: Provider catalog type definitions
- **`connectors/providers/catalog.go`**: Provider catalog access functions
- **`server/shared/catalog/catalog.go`**: Server-side catalog integration

### Documentation
- **`docs/src/*.mdx`**: User-facing documentation for amp.yaml configuration
- **`docs/src/examples/*.mdx`**: Example configurations and use cases

### Sample Files
All files in `/Users/chris/src/samples/`:
- **`samples/salesforce/amp.yaml`**: Salesforce integration example
- **`samples/hubspot/amp.yaml`**: HubSpot integration example
- **`samples/stripe/amp.yaml`**: Stripe integration example
- And other provider-specific samples

### Related Tools
- **`mcpanda/tools/ampconfig.go`**: Example of yaml.Node usage for parsing with position tracking

---

## Appendix A: Non-manifest (API) Validations - Reference Only

**Note:** The rules in this appendix are enforced at the API layer (e.g., during integration creation, installation, or proxy requests) and are **out-of-scope for the `amp-yaml-validator` library implementation**. They are documented here for completeness and to provide context for developers working across the full Ampersand platform.

### A.1 Integration Source Rules

#### Rule: Exactly one of sourceZipURL or sourceYAML required
- **Severity**: ERROR
- **Source**: `server/api/routes/api/integration.go` (ValidateSource function)
- **Error constants**: `errOneOfSourceZipURLOrSourceYAMLRequired`, `errOnlyOneOfSourceZipURLOrSourceYAMLAllowed`
- **Description**: When creating an integration via API, you must provide either `sourceZipURL` (URL to zipped amp.yaml) OR `sourceYAML` (inline YAML string), but not both
- **Note**: This is an API-level rule, not directly in amp.yaml content
- **Example violation (API request)**:
```json
{
  "name": "my-integration",
  "provider": "salesforce"
  // Missing both sourceZipURL and sourceYAML
}
```
```json
{
  "name": "my-integration",
  "provider": "salesforce",
  "sourceZipURL": "https://example.com/amp.zip",
  "sourceYAML": "specVersion: '1.0.0'..."  // Both specified - invalid
}
```

---

### A.2 Installation Validation Rules

#### Rule: Connection provider must match integration provider
- **Severity**: ERROR
- **Source**: `server/api/routes/api/installation.go` (validateInstallation function)
- **Description**: When installing an integration, the connection's provider must match the integration's provider
- **Note**: This is validated at installation time, not in amp.yaml
- **Example**: Cannot install a Salesforce integration with a HubSpot connection

#### Rule: Installation, connection, and integration must all exist
- **Severity**: ERROR
- **Source**: `server/api/routes/api/installation.go` (validateInstallation function)
- **Description**: All required entities must be present and non-nil during installation validation

---

### A.3 Proxy Validation Rules

#### Rule: Proxy requires installationId OR (groupRef + integrationName)
- **Severity**: ERROR
- **Source**: `server/api/routes/proxy/proxy.go` (validate function)
- **Error constant**: `ErrNoInstallationOrGroupRefAndIntegrationNameCombination`
- **Description**: Proxy requests must identify the target installation using either:
  - `installationId` header, OR
  - Both `groupRef` and `integrationName` headers together
- **Note**: This validates proxy API requests, not amp.yaml content

#### Rule: JWT credentials not allowed for proxy
- **Severity**: ERROR
- **Source**: `server/api/routes/proxy/proxy.go`
- **Description**: Proxy requests cannot use JWT-based authentication

#### Rule: ProjectId and Version required in proxy headers
- **Severity**: ERROR
- **Source**: `server/api/routes/proxy/proxy.go`
- **Description**: Proxy requests must include `projectId` and API version headers

---

## Appendix B: Error Constants Reference

This section lists all error constants from the codebase for cross-reference:

### Universal Validation Errors
```go
errInvalidSpecVersion                                          // validate.go:57
errMissingWriteObjects                                         // validate.go:67
errMissingReadObjects                                          // validate.go:84
ErrMissingMinimumRequiredFields                               // validate.go:90, 100, 475, 521
ErrInvalidSchedule                                            // validate.go:105, 196
ErrSubscribeRequiresRead                                      // validate.go:126
errMissingSubscribeObjects                                    // validate.go:130
ErrTooManySubscribeObjects                                    // validate.go:135
ErrSubscribeInheritFieldsAndMapping                          // validate.go:143
ErrInvalidInputEnabled                                        // validate.go:153
ErrWatchFieldsRequired                                        // validate.go:157
ErrWatchFieldsAndRequiredWatchFields                         // validate.go:162
errInvalidCronSchedule                                        // validate.go:196
ErrScheduleTooFrequent                                        // validate.go:206, 210, 228
ErrInvalidPageSizeConfig                                      // validate.go:254, 264, 268
ErrInvalidDeliveryConfig                                      // validate.go:277
errInvalidBackfill                                            // validate.go:289, 293
errDuplicateFieldMapping                                      // validate.go:334
ErrAlwaysEnabledObjectHasMoreRequiredMappingsThanBuilderMappings  // validate.go:468
```

### Runtime/Async Errors
```go
errDestinationNotFound                                        // destination.go:27
ErrMissingObjectInRevision                                    // integration.go:12
errInactiveDestination                                        // workflow.go:102
```

### API Layer Errors
```go
errOneOfSourceZipURLOrSourceYAMLRequired                      // integration.go
errOnlyOneOfSourceZipURLOrSourceYAMLAllowed                  // integration.go
ErrNoInstallationOrGroupRefAndIntegrationNameCombination     // proxy.go
```

---

### Rule Implementation Status

This table tracks which validation rules have been implemented in the `amp-yaml-validator` library and references to the corresponding server-side validation code.

| Rule ID | Rule Name | Status | Validator File | Server Reference | Notes |
|---------|-----------|--------|----------------|------------------|-------|
| **2.13 Duplicate Object Detection** |
| `duplicate-read-object` | No duplicate objects in read.objects | ✅ Implemented | `validator/duplicate.go` | `server/shared/common/validate.go` | Detects duplicates and reports both locations |
| `duplicate-write-object` | No duplicate objects in write.objects | ✅ Implemented | `validator/duplicate.go` | `server/shared/common/validate.go` | Same as read |
| `duplicate-subscribe-object` | No duplicate objects in subscribe.objects | ✅ Implemented | `validator/duplicate.go` | `server/shared/common/validate.go` | Same as read |
| **2.14 Subscribe Event Type Rules** |
| `subscribe-minimum-events` | Subscribe object with no event enabled warns (valid as a base definition) | ✅ Implemented (WARNING) | `validator/subscribe_events.go` | `server/shared/common/config.go:77-151` | Checks createEvent, updateEvent, deleteEvent, associationChangeEvent |
| **2.15 Field Mapping Naming Rules** |
| `field-mapping-simple-names` | Field mappings use simple names (no bracket notation) | ℹ️ Informational | `validator/jsonpath.go` | N/A | Bracket notation intentionally not supported (design decision) |
| **2.16 RequiredWatchFields Rules** |
| `watch-fields-no-nesting` | RequiredWatchFields cannot contain nested paths | ✅ Implemented | `validator/subscribe_events.go`, `validator/jsonpath.go` | `server/shared/common/config.go` | Detects dots and brackets in field names |
| **3.5 Google Calendar Rules** |
| `google-calendar-no-full-history` | Google Calendar events cannot use fullHistory | ✅ Implemented | `validator/provider_google_calendar.go` | `server/shared/common/validate.go:74-77` | Only applies to events object |
| `google-calendar-max-backfill` | Google Calendar events backfill max 28 days | ✅ Implemented | `validator/provider_google_calendar.go` | `server/shared/common/validate.go:74-77` | MaxGoogleCalendarBackfillDays = 28 |
| **3.6 Snowflake Rules** |
| `snowflake-only-full-history` | Snowflake only supports fullHistory backfill | ✅ Implemented | `validator/provider_snowflake.go` | `server/shared/common/validate.go` | Days-based backfill not allowed |

**Legend**:
- ✅ **Implemented**: Fully implemented and tested
- ⚠️ **Partial**: Basic implementation exists, advanced features pending
- ℹ️ **Informational**: Documented design decision or informational note (not enforced)
- ❌ **Not Implemented**: Planned but not yet implemented
- 🔄 **In Progress**: Currently being implemented

**Implementation Notes**:
- All implemented rules have corresponding test files in `validator/` with comprehensive test coverage
- Test data files exist in `testdata/invalid/` and `testdata/valid/` directories
- Server references point to the original validation logic in the Ampersand server codebase for maintainability

---

## Document Version

- **Version**: 1.1
- **Last Updated**: 2026-01-17
- **Based on**: Ampersand codebase analysis (70+ rules documented)
- **Changes in v1.1**:
  - Added section 2.13: Duplicate Object Detection Rules
  - Added section 2.14: Subscribe Event Type Rules
  - Added section 2.15: JSONPath Validation Rules
  - Added section 2.16: RequiredWatchFields Nested Path Rules
  - Added section 3.5: Google Calendar Rules
  - Added section 3.6: Snowflake Rules
  - Added Rule Implementation Status table in Appendix B
- **Implementation Status**: Core semantic validation rules implemented in amp-yaml-validator library

---

## Summary

This specification documents **60+ validation rules** discovered through comprehensive analysis of the Ampersand codebase:

- **Universal rules** (27 rules): Spec version, integration structure, read/write/subscribe validation, schedules, delivery modes, backfill, field mappings, always-enabled objects
- **Provider-specific rules** (7 rules): Salesforce limits, HubSpot requirements, provider capability checks
- **Async error prevention** (4 rules): Destination resolution, object existence, inactive destinations, timeout risks
- **API layer rules** (6 rules): Integration source, installation, proxy validation
- **Field-level rules** (5 rules): Required fields, type constraints, watch fields

Each rule includes severity (ERROR/WARNING), source code reference, error constants, examples, and implementation guidance. The specification provides a complete foundation for building the `amp-yaml-validator` library in subsequent phases.
