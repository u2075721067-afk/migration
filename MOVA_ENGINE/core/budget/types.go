package budget

import (
	"time"
)

// BudgetType represents different types of budgets
type BudgetType string

const (
	BudgetTypeRetries    BudgetType = "retries"
	BudgetTypeWorkflows  BudgetType = "workflows"
	BudgetTypeCPU        BudgetType = "cpu"
	BudgetTypeMemory     BudgetType = "memory"
	BudgetTypeAPIRequests BudgetType = "api_requests"
	BudgetTypeExecution  BudgetType = "execution_time"
)

// TimeWindow represents the time window for budget calculations
type TimeWindow string

const (
	TimeWindowMinute TimeWindow = "minute"
	TimeWindowHour   TimeWindow = "hour"
	TimeWindowDay    TimeWindow = "day"
	TimeWindowMonth  TimeWindow = "month"
)

// BudgetScope represents the scope of budget application
type BudgetScope string

const (
	BudgetScopeGlobal       BudgetScope = "global"
	BudgetScopeOrganization BudgetScope = "organization"
	BudgetScopeUser         BudgetScope = "user"
	BudgetScopeWorkflow     BudgetScope = "workflow"
	BudgetScopeSession      BudgetScope = "session"
)

// Budget represents a budget constraint
type Budget struct {
	ID          string      `json:"id" yaml:"id"`
	Name        string      `json:"name" yaml:"name"`
	Description string      `json:"description" yaml:"description"`
	Type        BudgetType  `json:"type" yaml:"type"`
	Scope       BudgetScope `json:"scope" yaml:"scope"`
	ScopeID     string      `json:"scopeId" yaml:"scopeId"` // Organization ID, User ID, etc.
	
	// Limits
	MaxCount    int64         `json:"maxCount,omitempty" yaml:"maxCount,omitempty"`       // Max number of operations
	MaxDuration time.Duration `json:"maxDuration,omitempty" yaml:"maxDuration,omitempty"` // Max execution time
	MaxMemory   int64         `json:"maxMemory,omitempty" yaml:"maxMemory,omitempty"`     // Max memory in bytes
	MaxCPU      float64       `json:"maxCPU,omitempty" yaml:"maxCPU,omitempty"`           // Max CPU usage (0-1)
	
	// Time constraints
	TimeWindow TimeWindow `json:"timeWindow" yaml:"timeWindow"`
	ResetTime  time.Time  `json:"resetTime,omitempty" yaml:"resetTime,omitempty"`
	
	// Status
	Enabled   bool      `json:"enabled" yaml:"enabled"`
	CreatedAt time.Time `json:"createdAt" yaml:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" yaml:"updatedAt"`
	
	// Current usage (runtime state)
	CurrentCount    int64         `json:"currentCount,omitempty" yaml:"-"`
	CurrentDuration time.Duration `json:"currentDuration,omitempty" yaml:"-"`
	CurrentMemory   int64         `json:"currentMemory,omitempty" yaml:"-"`
	CurrentCPU      float64       `json:"currentCPU,omitempty" yaml:"-"`
	LastUsedAt      time.Time     `json:"lastUsedAt,omitempty" yaml:"-"`
}

// BudgetUsage represents current usage of a budget
type BudgetUsage struct {
	BudgetID        string        `json:"budgetId"`
	Count           int64         `json:"count"`
	Duration        time.Duration `json:"duration"`
	Memory          int64         `json:"memory"`
	CPU             float64       `json:"cpu"`
	WindowStartTime time.Time     `json:"windowStartTime"`
	LastUpdated     time.Time     `json:"lastUpdated"`
}

// BudgetViolation represents a budget constraint violation
type BudgetViolation struct {
	ID              string      `json:"id"`
	BudgetID        string      `json:"budgetId"`
	BudgetName      string      `json:"budgetName"`
	Type            BudgetType  `json:"type"`
	Scope           BudgetScope `json:"scope"`
	ScopeID         string      `json:"scopeId"`
	ViolationType   string      `json:"violationType"` // "exceeded", "approaching"
	
	// Violation details
	Limit           interface{} `json:"limit"`
	CurrentValue    interface{} `json:"currentValue"`
	PercentageUsed  float64     `json:"percentageUsed"`
	
	// Context
	WorkflowID      string    `json:"workflowId,omitempty"`
	SessionID       string    `json:"sessionId,omitempty"`
	UserID          string    `json:"userId,omitempty"`
	OrganizationID  string    `json:"organizationId,omitempty"`
	
	Timestamp       time.Time `json:"timestamp"`
	Message         string    `json:"message"`
	Severity        string    `json:"severity"` // "warning", "error", "critical"
}

// BudgetCheckRequest represents a request to check budget constraints
type BudgetCheckRequest struct {
	Type            BudgetType  `json:"type"`
	Scope           BudgetScope `json:"scope"`
	ScopeID         string      `json:"scopeId"`
	
	// Resource usage to check
	Count           int64         `json:"count,omitempty"`
	Duration        time.Duration `json:"duration,omitempty"`
	Memory          int64         `json:"memory,omitempty"`
	CPU             float64       `json:"cpu,omitempty"`
	
	// Context
	WorkflowID      string `json:"workflowId,omitempty"`
	SessionID       string `json:"sessionId,omitempty"`
	UserID          string `json:"userId,omitempty"`
	OrganizationID  string `json:"organizationId,omitempty"`
}

// BudgetCheckResponse represents the response from a budget check
type BudgetCheckResponse struct {
	Allowed         bool               `json:"allowed"`
	Violations      []BudgetViolation  `json:"violations,omitempty"`
	RemainingQuota  map[string]interface{} `json:"remainingQuota,omitempty"`
	ResetTime       time.Time          `json:"resetTime,omitempty"`
	Message         string             `json:"message,omitempty"`
}

// BudgetAlert represents an alert configuration
type BudgetAlert struct {
	ID          string      `json:"id"`
	BudgetID    string      `json:"budgetId"`
	Threshold   float64     `json:"threshold"` // Percentage threshold (0-1)
	Channels    []string    `json:"channels"`  // email, slack, webhook
	Enabled     bool        `json:"enabled"`
	LastSent    time.Time   `json:"lastSent"`
	Cooldown    time.Duration `json:"cooldown"` // Minimum time between alerts
}

// ResourceMetrics represents current resource usage metrics
type ResourceMetrics struct {
	Timestamp       time.Time `json:"timestamp"`
	CPUUsage        float64   `json:"cpuUsage"`        // 0-1
	MemoryUsage     int64     `json:"memoryUsage"`     // bytes
	MemoryTotal     int64     `json:"memoryTotal"`     // bytes
	ActiveWorkflows int       `json:"activeWorkflows"`
	ActiveSessions  int       `json:"activeSessions"`
	QueuedJobs      int       `json:"queuedJobs"`
}

