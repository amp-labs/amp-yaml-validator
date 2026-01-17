package validator

import (
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateSubscribeEventTypes validates that subscribe objects have at least one event type enabled.
// This ensures subscribe configurations are meaningful and will actually receive events.
func validateSubscribeEventTypes(
	valCtx *ValidationContext,
	obj openapi.IntegrationSubscribeObject,
	objPath string,
) {
	// Check if at least one event type is enabled
	if !hasAnySubscribeEventEnabled(obj) {
		valCtx.AddErrorWithSuggestion(
			types.ErrNoSubscribeEvents,
			objPath,
			types.RuleSubscribeMinimumEvents,
			"Enable at least one event type: createEvent, updateEvent, deleteEvent, or associationChangeEvent",
		)
	}

	// Validate individual event configurations
	validateSubscribeCreateEvent(valCtx, obj.CreateEvent, objPath+".createEvent")
	validateSubscribeUpdateEventEnabled(valCtx, obj.UpdateEvent, objPath+".updateEvent")
	validateSubscribeDeleteEvent(valCtx, obj.DeleteEvent, objPath+".deleteEvent")
	validateSubscribeAssociationChangeEvent(valCtx, obj.AssociationChangeEvent, objPath+".associationChangeEvent")
}

// hasAnySubscribeEventEnabled checks if at least one event type is enabled.
func hasAnySubscribeEventEnabled(obj openapi.IntegrationSubscribeObject) bool {
	// CreateEvent is enabled if it exists and has enabled field set
	if obj.CreateEvent != nil && obj.CreateEvent.Enabled != nil {
		return true
	}

	// UpdateEvent is enabled if it exists and has enabled field set
	if obj.UpdateEvent != nil && obj.UpdateEvent.Enabled != nil {
		return true
	}

	// DeleteEvent is enabled if it exists and has enabled field set
	if obj.DeleteEvent != nil && obj.DeleteEvent.Enabled != nil {
		return true
	}

	// AssociationChangeEvent is enabled if it exists and has enabled field set
	if obj.AssociationChangeEvent != nil && obj.AssociationChangeEvent.Enabled != nil {
		return true
	}

	// OtherEvents is enabled if it exists
	if obj.OtherEvents != nil {
		return true
	}

	return false
}

// validateSubscribeCreateEvent validates create event configuration.
func validateSubscribeCreateEvent(valCtx *ValidationContext, event *openapi.CreateEvent, path string) {
	if event == nil {
		return
	}

	// For spec 1.0.0, enabled should be "always" if set
	if event.Enabled != nil {
		enabledStr := string(*event.Enabled)
		if enabledStr != types.EventEnabledAlways {
			valCtx.AddErrorWithSuggestion(
				fmt.Sprintf("invalid createEvent.enabled value: %s (must be 'always')", enabledStr),
				path+".enabled",
				types.RuleUpdateEventEnabled,
				"Set enabled to 'always' or remove the createEvent section",
			)
		}
	}
}

// validateSubscribeUpdateEventEnabled validates update event enabled field.
// Note: Other update event validation (watch fields) is handled by validateUpdateEvent in subscribe.go.
func validateSubscribeUpdateEventEnabled(valCtx *ValidationContext, event *openapi.UpdateEvent, path string) {
	if event == nil {
		return
	}

	// For spec 1.0.0, enabled should be "always" if set
	if event.Enabled != nil {
		enabledStr := string(*event.Enabled)
		if enabledStr != types.EventEnabledAlways {
			valCtx.AddErrorWithSuggestion(
				fmt.Sprintf("invalid updateEvent.enabled value: %s (must be 'always')", enabledStr),
				path+".enabled",
				types.RuleUpdateEventEnabled,
				"Set enabled to 'always' or remove the updateEvent section",
			)
		}
	}
}

// validateSubscribeDeleteEvent validates delete event configuration.
func validateSubscribeDeleteEvent(valCtx *ValidationContext, event *openapi.DeleteEvent, path string) {
	if event == nil {
		return
	}

	// For spec 1.0.0, enabled should be "always" if set
	if event.Enabled != nil {
		enabledStr := string(*event.Enabled)
		if enabledStr != types.EventEnabledAlways {
			valCtx.AddErrorWithSuggestion(
				fmt.Sprintf("invalid deleteEvent.enabled value: %s (must be 'always')", enabledStr),
				path+".enabled",
				types.RuleUpdateEventEnabled,
				"Set enabled to 'always' or remove the deleteEvent section",
			)
		}
	}
}

// validateSubscribeAssociationChangeEvent validates association change event configuration.
func validateSubscribeAssociationChangeEvent(valCtx *ValidationContext, event *openapi.AssociationChangeEvent, path string) {
	if event == nil {
		return
	}

	// For spec 1.0.0, enabled should be "always" if set
	if event.Enabled != nil {
		enabledStr := string(*event.Enabled)
		if enabledStr != types.EventEnabledAlways {
			valCtx.AddErrorWithSuggestion(
				fmt.Sprintf("invalid associationChangeEvent.enabled value: %s (must be 'always')", enabledStr),
				path+".enabled",
				types.RuleUpdateEventEnabled,
				"Set enabled to 'always' or remove the associationChangeEvent section",
			)
		}
	}
}
