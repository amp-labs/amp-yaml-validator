package types

// Error constants matching server/shared/common/validate.go
const (
	ErrInvalidSpecVersion                                            = "invalid spec version"
	ErrMissingReadObjects                                            = "read must contain an 'objects' list"
	ErrMissingWriteObjects                                           = "write must contain an 'objects' list"
	ErrMissingSubscribeObjects                                       = "subscribe must contain an 'objects' list"
	ErrInvalidCronSchedule                                           = "invalid cron schedule"
	ErrScheduleTooFrequent                                           = "schedule cannot be more frequent than every 10 minutes"
	ErrInvalidSchedule                                               = "invalid schedule"
	ErrInvalidDeliveryConfig                                         = "invalid delivery config"
	ErrInvalidPageSizeConfig                                         = "invalid page size configuration"
	ErrInvalidBackfill                                               = "invalid backfill"
	ErrDuplicateFieldMapping                                         = "duplicate field mappings present"
	ErrMissingMinimumRequiredFields                                  = "missing minimum required fields"
	ErrSubscribeRequiresRead                                         = "subscribe requires read to be defined at the same time"
	ErrSubscribeInheritFieldsAndMapping                              = "subscribe must have inheritFieldsAndMapping set to true"
	ErrWatchFieldsRequired                                           = "requiredWatchFields should have minimum 1 field or watchFieldsAuto should be set"
	ErrInvalidInputEnabled                                           = "invalid input enabled value"
	ErrWatchFieldsAndRequiredWatchFields                             = "watchFieldsAuto and requiredWatchFields cannot be used together"
	ErrAlwaysEnabledObjectHasMoreRequiredMappingsThanBuilderMappings = "always enabled object has more mappings required than builder mappings"
	ErrIntegrationStructure                                          = "integration structure validation failed"
	ErrMissingRequiredField                                          = "missing required field"
	ErrTooManySubscribeObjects                                       = "you cannot define an integration with more than 5 objects for subscribe actions, due to Salesforce Change Data Capture limits"
	ErrProviderNotSupported                                          = "provider not supported"
	ErrProviderCapabilityNotSupported                                = "provider does not support this capability"
	ErrModuleNotSupported                                            = "provider does not support this module"
	ErrCatalogAccessFailed                                           = "failed to access provider catalog"
)

// Rule identifiers for ValidationIssue.Rule field
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
	RuleProviderCapabilityBulkWrite = "provider-capability-bulk-write"
	RuleSalesforceSubscribeLimit    = "salesforce-subscribe-limit"
	RuleProviderModule              = "provider-module"
	RuleProviderNotSupported        = "provider-not-supported"
	RuleCatalogAccess               = "catalog-access"
	RuleObjectExists                = "object-exists"
	RuleProviderAppNotConfigured    = "provider-app-not-configured"
	RuleRateLimitExceeded           = "rate-limit-exceeded"
)

// Constants for validation rules
const (
	CurrentSpecVersion            = "1.0.0" // The only supported spec version
	MinScheduleIntervalMinutes    = 10      // Minimum schedule frequency in minutes
	MinOnRequestPageSize          = 50      // Minimum page size for onRequest delivery
	MaxOnRequestPageSize          = 500     // Maximum page size for onRequest delivery
	MaxSalesforceSubscribeObjects = 5       // Maximum subscribe objects for Salesforce
)
