package validator

import (
	"fmt"

	"github.com/amp-labs/amp-yaml-validator/openapi"
	"github.com/amp-labs/amp-yaml-validator/types"
)

// validateSubscribeEventTypes validates the event configuration of a subscribe object.
// An object with no event types is allowed as a base definition: it only supplies defaults
// (such as destination) and subscribes to nothing until an installation config enables an
// event. It is surfaced as a warning so an accidentally event-less object is still visible.
func validateSubscribeEventTypes(
	valCtx *ValidationContext,
	obj openapi.IntegrationSubscribeObject,
	objPath string,
) {
	// Warn if no event type is enabled
	if !hasAnySubscribeEventEnabled(obj) {
		valCtx.AddWarningWithSuggestion(
			types.ErrNoSubscribeEvents,
			objPath,
			types.RuleSubscribeMinimumEvents,
			"No events are enabled by default; this object only provides defaults such as destination, "+
				"and installations must enable events via their config. To subscribe by default, add "+
				"createEvent, updateEvent, deleteEvent, or associationChangeEvent with enabled: always",
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
