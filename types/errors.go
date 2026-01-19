// Package types defines types and constants for validation results and error messages.
package types

// Error constants matching server/shared/common/validate.go.
const (
	ErrMissingReadObjects                = "read must contain an 'objects' list"
	ErrMissingWriteObjects               = "write must contain an 'objects' list"
	ErrMissingSubscribeObjects           = "subscribe must contain an 'objects' list"
	ErrInvalidCronSchedule               = "invalid cron schedule"
	ErrScheduleTooFrequent               = "schedule cannot be more frequent than every 10 minutes"
	ErrInvalidSchedule                   = "invalid schedule"
	ErrInvalidBackfill                   = "invalid backfill"
	ErrSubscribeRequiresRead             = "subscribe requires read to be defined at the same time"
	ErrSubscribeInheritFieldsAndMapping  = "subscribe must have inheritFieldsAndMapping set to true"
	ErrWatchFieldsRequired               = "requiredWatchFields needs min 1 field or watchFieldsAuto set"
	ErrInvalidInputEnabled               = "invalid input enabled value"
	ErrWatchFieldsAndRequiredWatchFields = "watchFieldsAuto and requiredWatchFields cannot be used together"
)

// ErrDuplicateReadObject is returned when duplicate object found in read.objects.
const ErrDuplicateReadObject = "duplicate object found in read.objects"

// ErrDuplicateWriteObject is returned when duplicate object found in write.objects.
const ErrDuplicateWriteObject = "duplicate object found in write.objects"

// ErrDuplicateSubscribeObject is returned when duplicate object found in subscribe.objects.
const ErrDuplicateSubscribeObject = "duplicate object found in subscribe.objects"

// ErrNoSubscribeEvents is returned when subscribe object must have at least one event type enabled.
const ErrNoSubscribeEvents = "subscribe object must have at least one event type enabled"

// ErrNestedWatchField is returned when requiredWatchFields cannot contain nested paths.
const ErrNestedWatchField = "requiredWatchFields cannot contain nested paths"

// ErrInvalidBracketNotation is returned when invalid bracket notation in field mapping.
const ErrInvalidBracketNotation = "invalid bracket notation in field mapping"

// ErrInvalidJSONPath is returned when invalid JSONPath expression.
const ErrInvalidJSONPath = "invalid JSONPath expression"

// ErrGoogleCalendarFullHistory is returned when Google Calendar events cannot use fullHistory backfill.
const ErrGoogleCalendarFullHistory = "Google Calendar events cannot use fullHistory backfill"

// ErrGoogleCalendarMaxBackfill is returned when Google Calendar events backfill cannot exceed 28 days.
const ErrGoogleCalendarMaxBackfill = "Google Calendar events backfill cannot exceed 28 days"

// ErrSnowflakeBackfillDays is returned when Snowflake only supports fullHistory backfill.
const ErrSnowflakeBackfillDays = "Snowflake only supports fullHistory backfill"

// ErrAlwaysEnabledMappingCount is returned when always-enabled object has more required fields than builder mappings.
const ErrAlwaysEnabledMappingCount = "always-enabled object has more required fields than builder mappings"

// WarnAlwaysEnabledMinFields is a warning when always-enabled objects should have at least 3 required fields.
const WarnAlwaysEnabledMinFields = "always-enabled objects should have at least 3 required fields"

// Rule identifiers for ValidationIssue.Rule field.
const (
	RuleSpecVersion                 = "spec-version"
	RuleIntegrationStructure        = "integration-structure"
	RuleReadObjects                 = "read-objects"
	RuleWriteObjects                = "write-objects"
	RuleSubscribeObjects            = "subscribe-objects"
	RuleScheduleSyntax              = "schedule-syntax"
	RuleScheduleMinimumInterval     = "schedule-minimum-interval"
	RuleDeliveryMode                = "delivery-mode"
	RuleBackfillConfig              = "backfill-config"
	RuleFieldMappingUnique          = "field-mapping-unique"
	RuleSubscribeRequiresRead       = "subscribe-requires-read"
	RuleSubscribeInheritFields      = "subscribe-inherit-fields"
	RuleAlwaysEnabledFields         = "always-enabled-fields"
	RuleUpdateEventEnabled          = "update-event-enabled"
	RuleUpdateEventWatchFields      = "update-event-watch-fields"
	RuleRequiredField               = "required-field"
	RuleProviderCapabilityRead      = "provider-capability-read"
	RuleProviderCapabilityWrite     = "provider-capability-write"
	RuleProviderCapabilitySubscribe = "provider-capability-subscribe"
	RuleProviderCapabilityProxy     = "provider-capability-proxy"
	RuleSalesforceSubscribeLimit    = "salesforce-subscribe-limit"
	RuleProviderModule              = "provider-module"
	RuleProviderNotSupported        = "provider-not-supported"
	RuleCatalogAccess               = "catalog-access"
	RuleObjectExists                = "object-exists"
	RuleProviderAppNotConfigured    = "provider-app-not-configured"
	RuleProviderAppCheckFailed      = "provider-app-check-failed"
	RuleDestinationNotFound         = "destination-not-found"
	RuleDestinationCheckFailed      = "destination-check-failed"
)

// RuleDuplicateReadObject is the rule for duplicate read object detection.
const RuleDuplicateReadObject = "duplicate-read-object"

// RuleDuplicateWriteObject is the rule for duplicate write object detection.
const RuleDuplicateWriteObject = "duplicate-write-object"

// RuleDuplicateSubscribeObject is the rule for duplicate subscribe object detection.
const RuleDuplicateSubscribeObject = "duplicate-subscribe-object"

// RuleSubscribeMinimumEvents is the rule for subscribe minimum events.
const RuleSubscribeMinimumEvents = "subscribe-minimum-events"

// RuleNestedWatchFields is the rule for watch fields no nesting.
const RuleNestedWatchFields = "watch-fields-no-nesting"

// RuleFieldMappingJSONPath is the rule for field mapping JSONPath.
const RuleFieldMappingJSONPath = "field-mapping-jsonpath"

// RuleGoogleCalendarBackfill is the rule for Google Calendar backfill.
const RuleGoogleCalendarBackfill = "google-calendar-backfill"

// RuleSnowflakeBackfill is the rule for Snowflake backfill.
const RuleSnowflakeBackfill = "snowflake-backfill"

// RuleAlwaysEnabledMappingCount is the rule for always-enabled mapping count.
const RuleAlwaysEnabledMappingCount = "always-enabled-mapping-count"

// RuleAlwaysEnabledMinFields is the rule for always-enabled minimum fields.
const RuleAlwaysEnabledMinFields = "always-enabled-minimum-fields"

// Constants for validation rules.
const (
	CurrentSpecVersion            = "1.0.0" // The only supported spec version
	MinScheduleIntervalMinutes    = 10      // Minimum schedule frequency in minutes
	MinOnRequestPageSize          = 50      // Minimum page size for onRequest delivery
	MaxOnRequestPageSize          = 500     // Maximum page size for onRequest delivery
	MaxSalesforceSubscribeObjects = 5       // Maximum subscribe objects for Salesforce
	MaxGoogleCalendarBackfillDays = 28      // Maximum backfill days for Google Calendar events
	EventEnabledAlways            = "always" // Event enabled value for spec 1.0.0
	MinAlwaysEnabledFields        = 3       // Minimum recommended required fields for always-enabled objects
)
