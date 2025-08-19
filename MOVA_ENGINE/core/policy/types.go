package policy

import (
	"time"
)

// RetryProfile represents a predefined retry configuration
type RetryProfile struct {
	Name              string        `json:"name" yaml:"name"`
	Description       string        `json:"description" yaml:"description"`
	MaxRetries        int           `json:"maxRetries" yaml:"maxRetries"`
	InitialDelay      time.Duration `json:"initialDelay" yaml:"initialDelay"`
	MaxDelay          time.Duration `json:"maxDelay" yaml:"maxDelay"`
	BackoffMultiplier float64       `json:"backoffMultiplier" yaml:"backoffMultiplier"`
	Jitter            float64       `json:"jitter" yaml:"jitter"`
	Timeout           time.Duration `json:"timeout" yaml:"timeout"`
}

// Condition represents a condition for applying a retry policy
type Condition struct {
	ErrorType           string `json:"errorType,omitempty" yaml:"errorType,omitempty"`
	HTTPStatus          int    `json:"httpStatus,omitempty" yaml:"httpStatus,omitempty"`
	ErrorMessagePattern string `json:"errorMessagePattern,omitempty" yaml:"errorMessagePattern,omitempty"`
	ActionType          string `json:"actionType,omitempty" yaml:"actionType,omitempty"`
}

// BudgetConstraint represents budget limitations for retries
type BudgetConstraint struct {
	MaxRetriesPerWorkflow int           `json:"maxRetriesPerWorkflow,omitempty" yaml:"maxRetriesPerWorkflow,omitempty"`
	MaxRetriesPerSession  int           `json:"maxRetriesPerSession,omitempty" yaml:"maxRetriesPerSession,omitempty"`
	MaxTotalRetryTime     time.Duration `json:"maxTotalRetryTime,omitempty" yaml:"maxTotalRetryTime,omitempty"`
}

// Policy represents a complete retry policy
type Policy struct {
	ID                string           `json:"id" yaml:"id"`
	Name              string           `json:"name" yaml:"name"`
	Description       string           `json:"description" yaml:"description"`
	RetryProfile      string           `json:"retryProfile" yaml:"retryProfile"`
	Conditions        []Condition      `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	BudgetConstraints BudgetConstraint `json:"budgetConstraints,omitempty" yaml:"budgetConstraints,omitempty"`
	Enabled           bool             `json:"enabled" yaml:"enabled"`
	CreatedAt         time.Time        `json:"createdAt" yaml:"createdAt"`
	UpdatedAt         time.Time        `json:"updatedAt" yaml:"updatedAt"`
}

// PolicyMatch represents a policy match result
type PolicyMatch struct {
	Policy       *Policy
	RetryProfile *RetryProfile
	Score        int
}

