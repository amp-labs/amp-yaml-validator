# Warning Suppression Directives

The validator supports selective warning suppression using `amp:ignore` comment directives in your `amp.yaml` file, similar to linter directives in programming languages (like `//nolint` in Go).

## Syntax

### Ignore Specific Rules

Suppress specific validation rules by listing them in brackets:

```yaml
# amp:ignore[rule-name]
```

Multiple rules can be suppressed:

```yaml
# amp:ignore[rule1,rule2,rule3]
```

### Ignore All Warnings

Suppress all warnings with no brackets:

```yaml
# amp:ignore
```

## Placement

Directives can be placed in two ways:

### 1. Next-Line (Head Comment)

Place the directive on the line immediately before the field:

```yaml
# amp:ignore[destination-exists]
destination: temp-webhook
```

```yaml
# amp:ignore[frequent-schedule-risk]
- objectName: account
  schedule: "*/5 * * * *"
```

### 2. Inline (Line Comment)

Place the directive on the same line as the field:

```yaml
destination: temp-webhook  # amp:ignore[destination-exists]
schedule: "*/5 * * * *"  # amp:ignore[frequent-schedule-risk]
```

Inline comments also work on complex structure keys (objects, arrays):

```yaml
objects:  # amp:ignore[catalog-access]
  - objectName: account
  - objectName: contact

read:  # amp:ignore[destination-exists]
  objects:
    - objectName: account
      destination: webhook
```

## Scope

Directives apply to the field they're attached to **and all nested children**. For example:

```yaml
# amp:ignore[destination-exists]
read:
  objects:
    - objectName: account
      destination: webhook1  # ← suppressed
      schedule: "*/10 * * * *"
    - objectName: contact
      destination: webhook2  # ← suppressed
      schedule: "*/10 * * * *"
```

The directive on `read:` suppresses `destination-exists` warnings for all destinations under the `read` section.

## Examples

### Suppress Destination Warnings for Temporary Webhooks

```yaml
specVersion: 1.0.0
integrations:
  - name: myIntegration
    provider: salesforce
    read:
      objects:
        - objectName: account
          destination: temp-webhook  # amp:ignore[destination-exists]
          schedule: "*/10 * * * *"
```

### Suppress Frequent Schedule Warnings

```yaml
specVersion: 1.0.0
integrations:
  - name: myIntegration
    provider: salesforce
    read:
      objects:
        # amp:ignore[frequent-schedule-risk]
        - objectName: account
          destination: webhook
          schedule: "*/5 * * * *"
```

### Suppress Multiple Warnings

```yaml
specVersion: 1.0.0
integrations:
  - name: myIntegration
    provider: salesforce
    read:
      objects:
        # amp:ignore[destination-exists,frequent-schedule-risk]
        - objectName: account
          destination: temp-webhook
          schedule: "*/5 * * * *"
```

### Suppress All Warnings for an Object

```yaml
specVersion: 1.0.0
integrations:
  - name: myIntegration
    provider: salesforce
    read:
      objects:
        # amp:ignore
        - objectName: account
          destination: temp-webhook
          schedule: "*/5 * * * *"
```

## Available Rule Names

Common warning rules that can be suppressed:

| Rule Name | Description |
|-----------|-------------|
| `destination-exists` | Destination reference cannot be validated statically |
| `frequent-schedule-risk` | Schedule frequency may hit API rate limits |
| `catalog-access` | Object name validation skipped (catalog unavailable) |
| `object-exists` | Object not found in provider catalog |
| `large-backfill-risk` | Large backfill may cause timeouts |

To find the rule name for a warning, check the `[rule-name]` in the validation output:

```
⚠ Warnings (1):

  1. [destination-exists] Destination "webhook" is referenced but cannot be validated statically
     Path: $.integrations[0].read.objects[0].destination
     Location: line 9, column 24
```

The rule name is `destination-exists`.

## Important Notes

### Errors Cannot Be Suppressed

Directives **only suppress warnings**, not errors. This is intentional - errors indicate critical validation failures that would prevent your manifest from working correctly.

```yaml
# amp:ignore  ← This will NOT suppress the error below
schedule: "*/5 * * * *"  # ERROR: schedule too frequent (< 10 min minimum)
```

### Use Sparingly

Directives should be used sparingly and with good reason. Suppressing warnings means you're acknowledging the risk and taking responsibility for ensuring the configuration is correct.

Common valid use cases:
- Temporary configurations during development
- Known-good configurations that trigger false positives
- Provider-specific configurations where you have verified compatibility

### Documentation

Consider adding inline comments explaining **why** you're suppressing a warning:

```yaml
# Temporary webhook for testing - will be replaced before production
# amp:ignore[destination-exists]
destination: test-webhook-local
```

## Known Limitations

### Comment Position for Array Items

When placing a comment before an array item's first property, YAML attaches the comment to that property, not the item:

```yaml
# ❌ This doesn't work as expected
objects:
  - # amp:ignore[catalog-access]
    objectName: account    # ← Comment attaches HERE
    destination: webhook   # ← Not suppressed

# ✅ Workaround: Put directive before the dash
objects:
  # amp:ignore[catalog-access]
  - objectName: account    # ← Suppressed
    destination: webhook   # ← Also suppressed
```

This is a YAML parser limitation, not specific to our implementation.

### Hierarchical Suppression is Greedy

Directives suppress warnings for **all child paths**. Be mindful of broad suppressions:

```yaml
# This suppresses ALL warnings under integrations (very broad!)
integrations:  # amp:ignore
  - name: test
    # ... everything suppressed ...
```

Prefer more targeted suppressions when possible.

## CLI Behavior

When using strict mode (`--strict`), suppressed warnings are still treated as warnings (not errors), so suppression works as expected:

```bash
# Without directive: fails in strict mode
amp-yaml-validator --strict amp.yaml

# With directive: passes even in strict mode
# (assuming the warning is suppressed)
amp-yaml-validator --strict amp.yaml
```

## See Also

- [README.md](README.md) - Main documentation
- [VALIDATION_RULES.md](VALIDATION_RULES.md) - Complete list of validation rules
