package validator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirectiveSuppressionIntegration(t *testing.T) {
	t.Parallel()

	t.Run("inline directive suppresses specific warning", func(t *testing.T) {
		t.Parallel()

		yaml := `
specVersion: 1.0.0
integrations:
  - name: test
    provider: salesforce
    read:
      objects:
        - objectName: account
          destination: webhook  # amp:ignore[destination-exists]
          schedule: "*/10 * * * *"
`
		validator := NewValidator()
		result, err := validator.ValidateBytes([]byte(yaml))
		require.NoError(t, err)

		// Should have no destination-exists warning
		for _, warning := range result.Warnings {
			require.NotEqual(t, "destination-exists", warning.Rule,
				"destination-exists warning should be suppressed")
		}
	})

	t.Run("next-line directive on array item suppresses warnings", func(t *testing.T) {
		t.Parallel()

		yaml := `
specVersion: 1.0.0
integrations:
  - name: test
    provider: salesforce
    read:
      objects:
        # amp:ignore[frequent-schedule-risk]
        - objectName: account
          destination: webhook
          schedule: "*/5 * * * *"
`
		validator := NewValidator()
		result, err := validator.ValidateBytes([]byte(yaml))
		require.NoError(t, err)

		// Should have no frequent-schedule-risk warning
		for _, warning := range result.Warnings {
			require.NotEqual(t, "frequent-schedule-risk", warning.Rule,
				"frequent-schedule-risk warning should be suppressed")
		}
	})

	t.Run("wildcard amp:ignore suppresses all warnings", func(t *testing.T) {
		t.Parallel()

		yaml := `
specVersion: 1.0.0
integrations:
  - name: test
    provider: salesforce
    read:
      objects:
        # amp:ignore
        - objectName: account
          destination: temp-webhook
          schedule: "*/5 * * * *"
`
		validator := NewValidator()
		result, err := validator.ValidateBytes([]byte(yaml))
		require.NoError(t, err)

		// Should have no warnings about this object
		for _, warning := range result.Warnings {
			require.NotContains(t, warning.Path, "$.integrations[0].read.objects[0]",
				"all warnings for this object should be suppressed")
		}
	})

	t.Run("multiple rules can be suppressed", func(t *testing.T) {
		t.Parallel()

		yaml := `
specVersion: 1.0.0
integrations:
  - name: test
    provider: salesforce
    read:
      objects:
        # amp:ignore[destination-exists,frequent-schedule-risk]
        - objectName: account
          destination: temp-webhook
          schedule: "*/5 * * * *"
`
		validator := NewValidator()
		result, err := validator.ValidateBytes([]byte(yaml))
		require.NoError(t, err)

		// Should have neither destination-exists nor frequent-schedule-risk warnings
		for _, warning := range result.Warnings {
			require.NotEqual(t, "destination-exists", warning.Rule)
			require.NotEqual(t, "frequent-schedule-risk", warning.Rule)
		}
	})

	t.Run("errors cannot be suppressed", func(t *testing.T) {
		t.Parallel()

		yaml := `
specVersion: 1.0.0  # amp:ignore
integrations:
  - name: test
    provider: salesforce
    # amp:ignore[schedule-minimum-interval]
    read:
      objects:
        - objectName: account
          destination: webhook
          schedule: "*/5 * * * *"  # amp:ignore
`
		validator := NewValidator()
		result, err := validator.ValidateBytes([]byte(yaml))
		require.NoError(t, err)

		// Should still have the schedule error (errors can't be suppressed)
		hasScheduleError := false
		for _, error := range result.Errors {
			if error.Rule == "schedule-minimum-interval" {
				hasScheduleError = true
				break
			}
		}
		require.True(t, hasScheduleError, "errors should not be suppressible")
	})

	t.Run("directive applies to child paths", func(t *testing.T) {
		t.Parallel()

		yaml := `
specVersion: 1.0.0
integrations:
  - name: test
    provider: salesforce
    # amp:ignore[destination-exists]
    read:
      objects:
        - objectName: account
          destination: webhook1
          schedule: "*/10 * * * *"
        - objectName: contact
          destination: webhook2
          schedule: "*/10 * * * *"
`
		validator := NewValidator()
		result, err := validator.ValidateBytes([]byte(yaml))
		require.NoError(t, err)

		// Should have no destination-exists warnings for any object under read
		for _, warning := range result.Warnings {
			require.NotEqual(t, "destination-exists", warning.Rule,
				"directive on parent should suppress warnings on all children")
		}
	})

	t.Run("inline directive on complex structure key", func(t *testing.T) {
		t.Parallel()

		yaml := `
specVersion: 1.0.0
integrations:
  - name: test
    provider: salesforce
    read:
      objects:  # amp:ignore[catalog-access]
        - objectName: account
          destination: webhook
          schedule: "*/10 * * * *"
        - objectName: contact
          destination: webhook
          schedule: "*/10 * * * *"
`
		validator := NewValidator()
		result, err := validator.ValidateBytes([]byte(yaml))
		require.NoError(t, err)

		// Should have no catalog-access warnings for any object under objects
		for _, warning := range result.Warnings {
			require.NotEqual(t, "catalog-access", warning.Rule,
				"inline directive on objects: key should suppress catalog warnings")
		}
	})

	t.Run("inline directive on section suppresses all child warnings", func(t *testing.T) {
		t.Parallel()

		yaml := `
specVersion: 1.0.0
integrations:
  - name: test
    provider: salesforce
    read:  # amp:ignore
      objects:
        - objectName: account
          destination: webhook
          schedule: "*/10 * * * *"
`
		validator := NewValidator()
		result, err := validator.ValidateBytes([]byte(yaml))
		require.NoError(t, err)

		// Should have no warnings under read section
		for _, warning := range result.Warnings {
			require.NotContains(t, warning.Path, "$.integrations[0].read",
				"inline wildcard directive on read: should suppress all warnings")
		}
	})
}
