package configmanager

import (
	"time"
)

// ConfigFormat represents supported configuration formats
type ConfigFormat string

const (
	FormatJSON ConfigFormat = "json"
	FormatYAML ConfigFormat = "yaml"
	FormatHCL  ConfigFormat = "hcl"
)

// ImportMode represents how to handle existing configurations
type ImportMode string

const (
	ModeOverwrite ImportMode = "overwrite"
	ModeMerge     ImportMode = "merge"
	ModeValidate  ImportMode = "validate"
)

// ConfigMetadata represents metadata for configuration files
type ConfigMetadata struct {
	Version     string    `json:"version" yaml:"version"`
	Format      string    `json:"format" yaml:"format"`
	GeneratedAt time.Time `json:"generatedAt" yaml:"generatedAt"`
	Source      string    `json:"source,omitempty" yaml:"source,omitempty"`
	Checksum    string    `json:"checksum,omitempty" yaml:"checksum,omitempty"`
}

// ConfigBundle represents a complete configuration export
type ConfigBundle struct {
	Metadata      ConfigMetadata       `json:"metadata" yaml:"metadata"`
	Policies      []PolicyConfig       `json:"policies,omitempty" yaml:"policies,omitempty"`
	Budgets       []BudgetConfig       `json:"budgets,omitempty" yaml:"budgets,omitempty"`
	DLQEntries    []DLQEntryConfig     `json:"dlqEntries,omitempty" yaml:"dlqEntries,omitempty"`
	Workflows     []WorkflowConfig     `json:"workflows,omitempty" yaml:"workflows,omitempty"`
	RetryProfiles []RetryProfileConfig `json:"retryProfiles,omitempty" yaml:"retryProfiles,omitempty"`
	Rules         []RuleConfig         `json:"rules,omitempty" yaml:"rules,omitempty"`
	RuleSets      []RuleSetConfig      `json:"ruleSets,omitempty" yaml:"ruleSets,omitempty"`
}

// PolicyConfig represents a policy configuration for export/import
type PolicyConfig struct {
	ID                string                 `json:"id" yaml:"id"`
	Name              string                 `json:"name" yaml:"name"`
	Description       string                 `json:"description" yaml:"description"`
	RetryProfile      string                 `json:"retryProfile" yaml:"retryProfile"`
	Conditions        []ConditionConfig      `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	BudgetConstraints BudgetConstraintConfig `json:"budgetConstraints,omitempty" yaml:"budgetConstraints,omitempty"`
	Enabled           bool                   `json:"enabled" yaml:"enabled"`
	CreatedAt         time.Time              `json:"createdAt" yaml:"createdAt"`
	UpdatedAt         time.Time              `json:"updatedAt" yaml:"updatedAt"`
}

// ConditionConfig represents a condition configuration
type ConditionConfig struct {
	ErrorType           string `json:"errorType,omitempty" yaml:"errorType,omitempty"`
	HTTPStatus          int    `json:"httpStatus,omitempty" yaml:"httpStatus,omitempty"`
	ErrorMessagePattern string `json:"errorMessagePattern,omitempty" yaml:"errorMessagePattern,omitempty"`
	ActionType          string `json:"actionType,omitempty" yaml:"actionType,omitempty"`
}

// BudgetConstraintConfig represents budget constraint configuration
type BudgetConstraintConfig struct {
	MaxRetriesPerWorkflow int           `json:"maxRetriesPerWorkflow,omitempty" yaml:"maxRetriesPerWorkflow,omitempty"`
	MaxRetriesPerSession  int           `json:"maxRetriesPerSession,omitempty" yaml:"maxRetriesPerSession,omitempty"`
	MaxTotalRetryTime     time.Duration `json:"maxTotalRetryTime,omitempty" yaml:"maxTotalRetryTime,omitempty"`
}

// BudgetConfig represents a budget configuration for export/import
type BudgetConfig struct {
	ID          string        `json:"id" yaml:"id"`
	Name        string        `json:"name" yaml:"name"`
	Description string        `json:"description" yaml:"description"`
	Type        string        `json:"type" yaml:"type"`
	Scope       string        `json:"scope" yaml:"scope"`
	MaxCount    int           `json:"maxCount,omitempty" yaml:"maxCount,omitempty"`
	MaxMemory   int64         `json:"maxMemory,omitempty" yaml:"maxMemory,omitempty"`
	MaxCPU      float64       `json:"maxCPU,omitempty" yaml:"maxCPU,omitempty"`
	MaxDuration time.Duration `json:"maxDuration,omitempty" yaml:"maxDuration,omitempty"`
	TimeWindow  time.Duration `json:"timeWindow" yaml:"timeWindow"`
	Enabled     bool          `json:"enabled" yaml:"enabled"`
	CreatedAt   time.Time     `json:"createdAt" yaml:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt" yaml:"updatedAt"`
}

// DLQEntryConfig represents a DLQ entry configuration (sandbox mode only)
type DLQEntryConfig struct {
	ID          string                 `json:"id" yaml:"id"`
	WorkflowID  string                 `json:"workflowId" yaml:"workflowId"`
	ActionID    string                 `json:"actionId" yaml:"actionId"`
	Error       string                 `json:"error" yaml:"error"`
	Payload     map[string]interface{} `json:"payload" yaml:"payload"`
	RetryCount  int                    `json:"retryCount" yaml:"retryCount"`
	CreatedAt   time.Time              `json:"createdAt" yaml:"createdAt"`
	LastRetryAt *time.Time             `json:"lastRetryAt,omitempty" yaml:"lastRetryAt,omitempty"`
}

// WorkflowConfig represents a workflow configuration for export/import
type WorkflowConfig struct {
	ID          string                 `json:"id" yaml:"id"`
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description" yaml:"description"`
	Intent      string                 `json:"intent" yaml:"intent"`
	Payload     map[string]interface{} `json:"payload" yaml:"payload"`
	Actions     []ActionConfig         `json:"actions" yaml:"actions"`
	CreatedAt   time.Time              `json:"createdAt" yaml:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt" yaml:"updatedAt"`
}

// ActionConfig represents an action configuration
type ActionConfig struct {
	ID      string                 `json:"id" yaml:"id"`
	Type    string                 `json:"type" yaml:"type"`
	Payload map[string]interface{} `json:"payload" yaml:"payload"`
	Timeout time.Duration          `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Retry   *RetryConfig           `json:"retry,omitempty" yaml:"retry,omitempty"`
}

// RetryConfig represents retry configuration for actions
type RetryConfig struct {
	MaxRetries        int           `json:"maxRetries" yaml:"maxRetries"`
	InitialDelay      time.Duration `json:"initialDelay" yaml:"initialDelay"`
	MaxDelay          time.Duration `json:"maxDelay" yaml:"maxDelay"`
	BackoffMultiplier float64       `json:"backoffMultiplier" yaml:"backoffMultiplier"`
	Jitter            float64       `json:"jitter" yaml:"jitter"`
}

// RetryProfileConfig represents a retry profile configuration
type RetryProfileConfig struct {
	Name              string        `json:"name" yaml:"name"`
	Description       string        `json:"description" yaml:"description"`
	MaxRetries        int           `json:"maxRetries" yaml:"maxRetries"`
	InitialDelay      time.Duration `json:"initialDelay" yaml:"initialDelay"`
	MaxDelay          time.Duration `json:"maxDelay" yaml:"maxDelay"`
	BackoffMultiplier float64       `json:"backoffMultiplier" yaml:"backoffMultiplier"`
	Jitter            float64       `json:"jitter" yaml:"jitter"`
	Timeout           time.Duration `json:"timeout" yaml:"timeout"`
	CreatedAt         time.Time     `json:"createdAt" yaml:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt" yaml:"updatedAt"`
}

// ExportOptions represents options for configuration export
type ExportOptions struct {
	Format           ConfigFormat `json:"format" yaml:"format"`
	IncludeDLQ       bool         `json:"includeDLQ" yaml:"includeDLQ"`
	IncludeWorkflows bool         `json:"includeWorkflows" yaml:"includeWorkflows"`
	Compress         bool         `json:"compress" yaml:"compress"`
}

// ImportOptions represents options for configuration import
type ImportOptions struct {
	Format       ConfigFormat `json:"format" yaml:"format"`
	Mode         ImportMode   `json:"mode" yaml:"mode"`
	ValidateOnly bool         `json:"validateOnly" yaml:"validateOnly"`
	DryRun       bool         `json:"dryRun" yaml:"dryRun"`
	Overwrite    bool         `json:"overwrite" yaml:"overwrite"`
}

// ImportResult represents the result of configuration import
type ImportResult struct {
	Success  bool            `json:"success" yaml:"success"`
	Imported int             `json:"imported" yaml:"imported"`
	Skipped  int             `json:"skipped" yaml:"skipped"`
	Errors   []ImportError   `json:"errors,omitempty" yaml:"errors,omitempty"`
	Warnings []ImportWarning `json:"warnings,omitempty" yaml:"warnings,omitempty"`
	Summary  ImportSummary   `json:"summary" yaml:"summary"`
}

// ImportError represents an error during import
type ImportError struct {
	Type    string `json:"type" yaml:"type"`
	Message string `json:"message" yaml:"message"`
	Line    int    `json:"line,omitempty" yaml:"line,omitempty"`
	Column  int    `json:"column,omitempty" yaml:"column,omitempty"`
	Context string `json:"context,omitempty" yaml:"context,omitempty"`
}

// ImportWarning represents a warning during import
type ImportWarning struct {
	Type    string `json:"type" yaml:"type"`
	Message string `json:"message" yaml:"message"`
	Line    int    `json:"line,omitempty" yaml:"line,omitempty"`
	Column  int    `json:"column,omitempty" yaml:"column,omitempty"`
	Context string `json:"context,omitempty" yaml:"context,omitempty"`
}

// ImportSummary represents a summary of import results
type ImportSummary struct {
	Policies      int `json:"policies" yaml:"policies"`
	Budgets       int `json:"budgets" yaml:"budgets"`
	DLQEntries    int `json:"dlqEntries" yaml:"dlqEntries"`
	Workflows     int `json:"workflows" yaml:"workflows"`
	RetryProfiles int `json:"retryProfiles" yaml:"retryProfiles"`
	Rules         int `json:"rules" yaml:"rules"`
	RuleSets      int `json:"ruleSets" yaml:"ruleSets"`
}

// RuleConfig represents a rule configuration for export/import
type RuleConfig struct {
	ID          string                `json:"id" yaml:"id"`
	Name        string                `json:"name" yaml:"name"`
	Description string                `json:"description" yaml:"description"`
	Priority    int                   `json:"priority" yaml:"priority"`
	Enabled     bool                  `json:"enabled" yaml:"enabled"`
	Conditions  []RuleConditionConfig `json:"conditions" yaml:"conditions"`
	Actions     []RuleActionConfig    `json:"actions" yaml:"actions"`
	CreatedAt   time.Time             `json:"createdAt" yaml:"createdAt"`
	UpdatedAt   time.Time             `json:"updatedAt" yaml:"updatedAt"`
}

// RuleConditionConfig represents a rule condition configuration
type RuleConditionConfig struct {
	Field    string      `json:"field" yaml:"field"`
	Operator string      `json:"operator" yaml:"operator"`
	Value    interface{} `json:"value" yaml:"value"`
	Negate   bool        `json:"negate,omitempty" yaml:"negate,omitempty"`
}

// RuleActionConfig represents a rule action configuration
type RuleActionConfig struct {
	Type   string                 `json:"type" yaml:"type"`
	Params map[string]interface{} `json:"params" yaml:"params"`
}

// RuleSetConfig represents a ruleset configuration for export/import
type RuleSetConfig struct {
	Version     string            `json:"version" yaml:"version"`
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Rules       []RuleConfig      `json:"rules" yaml:"rules"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"createdAt" yaml:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt" yaml:"updatedAt"`
}
