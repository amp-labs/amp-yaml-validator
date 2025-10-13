package types

// ValidationIssue represents a single validation error or warning with location information.
type ValidationIssue struct {
	Severity   string `json:"severity"`   // "error" or "warning"
	Message    string `json:"message"`    // Human-readable error message
	Line       int    `json:"line"`       // Line number in YAML file (1-based)
	Column     int    `json:"column"`     // Column number (1-based)
	Path       string `json:"path"`       // JSONPath-style path (e.g., "$.integrations[0].read.objects[1].schedule")
	Rule       string `json:"rule"`       // Rule identifier (e.g., "schedule-minimum-interval")
	Suggestion string `json:"suggestion"` // Optional suggestion for fixing the issue
}

// ValidationResult contains the outcome of validating an amp.yaml file.
type ValidationResult struct {
	Valid    bool              `json:"valid"`    // True if no errors (warnings don't affect validity unless strict mode)
	Errors   []ValidationIssue `json:"errors"`   // List of error-level issues
	Warnings []ValidationIssue `json:"warnings"` // List of warning-level issues
}

// NewError creates a new error-level validation issue.
func NewError(message, path, rule string, line, column int) ValidationIssue {
	return ValidationIssue{
		Severity: "error",
		Message:  message,
		Line:     line,
		Column:   column,
		Path:     path,
		Rule:     rule,
	}
}

// NewWarning creates a new warning-level validation issue.
func NewWarning(message, path, rule string, line, column int) ValidationIssue {
	return ValidationIssue{
		Severity: "warning",
		Message:  message,
		Line:     line,
		Column:   column,
		Path:     path,
		Rule:     rule,
	}
}

// WithSuggestion adds a suggestion to an existing validation issue.
func WithSuggestion(issue ValidationIssue, suggestion string) ValidationIssue {
	issue.Suggestion = suggestion
	return issue
}
