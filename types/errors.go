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
)

// Constants for validation rules.
const (
	CurrentSpecVersion            = "1.0.0" // The only supported spec version
	MinScheduleIntervalMinutes    = 10      // Minimum schedule frequency in minutes
	MinOnRequestPageSize          = 50      // Minimum page size for onRequest delivery
	MaxOnRequestPageSize          = 500     // Maximum page size for onRequest delivery
	MaxSalesforceSubscribeObjects = 5       // Maximum subscribe objects for Salesforce
)
