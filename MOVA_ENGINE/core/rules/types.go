package rules

import (
	"time"
)

// Rule represents a single rule with conditions and actions
type Rule struct {
	ID          string      `json:"id" yaml:"id"`
	Name        string      `json:"name" yaml:"name"`
	Description string      `json:"description" yaml:"description"`
	Priority    int         `json:"priority" yaml:"priority"`
	Enabled     bool        `json:"enabled" yaml:"enabled"`
	Conditions  []Condition `json:"conditions" yaml:"conditions"`
	Actions     []Action    `json:"actions" yaml:"actions"`
	CreatedAt   time.Time   `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" yaml:"updated_at"`
}

// Condition represents a rule condition
type Condition struct {
	Field    string      `json:"field" yaml:"field"`
	Operator string      `json:"operator" yaml:"operator"`
	Value    interface{} `json:"value" yaml:"value"`
	Negate   bool        `json:"negate,omitempty" yaml:"negate,omitempty"`
}

// Action represents an action to execute when rule matches
type Action struct {
	Type   string                 `json:"type" yaml:"type"`
	Params map[string]interface{} `json:"params" yaml:"params"`
}

// RuleSet represents a collection of rules with versioning
type RuleSet struct {
	Version     string            `json:"version" yaml:"version"`
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Rules       []Rule            `json:"rules" yaml:"rules"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
}

// Context represents the execution context for rule evaluation
type Context struct {
	Variables map[string]interface{} `json:"variables"`
	Request   map[string]interface{} `json:"request"`
	Response  map[string]interface{} `json:"response"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// Result represents the result of rule evaluation
type Result struct {
	RuleID       string                 `json:"rule_id"`
	Matched      bool                   `json:"matched"`
	Actions      []Action               `json:"actions"`
	Variables    map[string]interface{} `json:"variables"`
	ExecutedAt   time.Time              `json:"executed_at"`
	ExecutionLog []string               `json:"execution_log"`
	Error        string                 `json:"error,omitempty"`
}

// ExecutionResult represents the result of action execution
type ExecutionResult struct {
	ActionType string                 `json:"action_type"`
	Success    bool                   `json:"success"`
	Output     map[string]interface{} `json:"output"`
	Error      string                 `json:"error,omitempty"`
	Duration   time.Duration          `json:"duration"`
}

// Supported operators
const (
	OpEquals      = "=="
	OpNotEquals   = "!="
	OpGreater     = ">"
	OpGreaterEq   = ">="
	OpLess        = "<"
	OpLessEq      = "<="
	OpContains    = "contains"
	OpNotContains = "not_contains"
	OpRegex       = "regex"
	OpIn          = "in"
	OpNotIn       = "not_in"
	OpExists      = "exists"
	OpNotExists   = "not_exists"
)

// Supported action types
const (
	ActionSetVar    = "set_var"
	ActionRetry     = "retry"
	ActionHTTPCall  = "http_call"
	ActionSkip      = "skip"
	ActionLog       = "log"
	ActionRoute     = "route"
	ActionStop      = "stop"
	ActionTransform = "transform"
)

// RuleEngine interface defines the core rule engine operations
type RuleEngine interface {
	// Evaluate evaluates a single rule against the context
	Evaluate(rule Rule, ctx Context) (Result, error)

	// Execute executes a list of actions in the context
	Execute(actions []Action, ctx Context) ([]ExecutionResult, error)

	// Run runs a complete ruleset against the context
	Run(ruleset RuleSet, ctx Context) ([]Result, error)

	// ValidateRule validates rule syntax and structure
	ValidateRule(rule Rule) error

	// ValidateRuleSet validates ruleset syntax and structure
	ValidateRuleSet(ruleset RuleSet) error
}

// RuleRepository interface for rule persistence
type RuleRepository interface {
	// GetRule retrieves a rule by ID
	GetRule(id string) (Rule, error)

	// ListRules lists all rules with optional filtering
	ListRules(filter map[string]interface{}) ([]Rule, error)

	// CreateRule creates a new rule
	CreateRule(rule Rule) error

	// UpdateRule updates an existing rule
	UpdateRule(rule Rule) error

	// DeleteRule deletes a rule by ID
	DeleteRule(id string) error

	// GetRuleSet retrieves a ruleset by name
	GetRuleSet(name string) (RuleSet, error)

	// SaveRuleSet saves a ruleset
	SaveRuleSet(ruleset RuleSet) error

	// ListRuleSets lists all rulesets
	ListRuleSets() ([]RuleSet, error)
}
