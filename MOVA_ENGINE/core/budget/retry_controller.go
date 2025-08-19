package budget

import (
	"context"
	"fmt"
	"time"
)

// RetryController manages retry budget enforcement
type RetryController struct {
	budgetManager *Manager
}

// NewRetryController creates a new retry controller
func NewRetryController(budgetManager *Manager) *RetryController {
	return &RetryController{
		budgetManager: budgetManager,
	}
}

// RetryContext contains context information for retry operations
type RetryContext struct {
	WorkflowID     string
	SessionID      string
	UserID         string
	OrganizationID string
	ActionType     string
	ErrorType      string
	AttemptNumber  int
	ExecutionTime  time.Duration
}

// CheckRetryAllowed checks if a retry operation is allowed within budget constraints
func (rc *RetryController) CheckRetryAllowed(ctx context.Context, retryCtx *RetryContext) (*BudgetCheckResponse, error) {
	// Create budget check request for retry operation
	req := &BudgetCheckRequest{
		Type:           BudgetTypeRetries,
		Count:          1, // One retry attempt
		Duration:       retryCtx.ExecutionTime,
		WorkflowID:     retryCtx.WorkflowID,
		SessionID:      retryCtx.SessionID,
		UserID:         retryCtx.UserID,
		OrganizationID: retryCtx.OrganizationID,
	}
	
	// Check global retry budget
	req.Scope = BudgetScopeGlobal
	globalResponse, err := rc.budgetManager.CheckBudget(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check global retry budget: %w", err)
	}
	
	if !globalResponse.Allowed {
		return globalResponse, nil
	}
	
	// Check organization-level retry budget
	if retryCtx.OrganizationID != "" {
		req.Scope = BudgetScopeOrganization
		req.ScopeID = retryCtx.OrganizationID
		orgResponse, err := rc.budgetManager.CheckBudget(req)
		if err != nil {
			return nil, fmt.Errorf("failed to check organization retry budget: %w", err)
		}
		
		if !orgResponse.Allowed {
			return orgResponse, nil
		}
		
		// Merge violations
		globalResponse.Violations = append(globalResponse.Violations, orgResponse.Violations...)
	}
	
	// Check user-level retry budget
	if retryCtx.UserID != "" {
		req.Scope = BudgetScopeUser
		req.ScopeID = retryCtx.UserID
		userResponse, err := rc.budgetManager.CheckBudget(req)
		if err != nil {
			return nil, fmt.Errorf("failed to check user retry budget: %w", err)
		}
		
		if !userResponse.Allowed {
			return userResponse, nil
		}
		
		// Merge violations
		globalResponse.Violations = append(globalResponse.Violations, userResponse.Violations...)
	}
	
	// Check workflow-level retry budget
	if retryCtx.WorkflowID != "" {
		req.Scope = BudgetScopeWorkflow
		req.ScopeID = retryCtx.WorkflowID
		workflowResponse, err := rc.budgetManager.CheckBudget(req)
		if err != nil {
			return nil, fmt.Errorf("failed to check workflow retry budget: %w", err)
		}
		
		if !workflowResponse.Allowed {
			return workflowResponse, nil
		}
		
		// Merge violations
		globalResponse.Violations = append(globalResponse.Violations, workflowResponse.Violations...)
	}
	
	// Check session-level retry budget
	if retryCtx.SessionID != "" {
		req.Scope = BudgetScopeSession
		req.ScopeID = retryCtx.SessionID
		sessionResponse, err := rc.budgetManager.CheckBudget(req)
		if err != nil {
			return nil, fmt.Errorf("failed to check session retry budget: %w", err)
		}
		
		if !sessionResponse.Allowed {
			return sessionResponse, nil
		}
		
		// Merge violations
		globalResponse.Violations = append(globalResponse.Violations, sessionResponse.Violations...)
	}
	
	return globalResponse, nil
}

// RecordRetryUsage records retry usage for budget tracking
func (rc *RetryController) RecordRetryUsage(ctx context.Context, retryCtx *RetryContext) error {
	req := &BudgetCheckRequest{
		Type:           BudgetTypeRetries,
		Count:          1,
		Duration:       retryCtx.ExecutionTime,
		WorkflowID:     retryCtx.WorkflowID,
		SessionID:      retryCtx.SessionID,
		UserID:         retryCtx.UserID,
		OrganizationID: retryCtx.OrganizationID,
	}
	
	// Record usage for all applicable scopes
	scopes := []struct {
		scope   BudgetScope
		scopeID string
	}{
		{BudgetScopeGlobal, ""},
		{BudgetScopeOrganization, retryCtx.OrganizationID},
		{BudgetScopeUser, retryCtx.UserID},
		{BudgetScopeWorkflow, retryCtx.WorkflowID},
		{BudgetScopeSession, retryCtx.SessionID},
	}
	
	for _, scope := range scopes {
		if scope.scope == BudgetScopeGlobal || scope.scopeID != "" {
			req.Scope = scope.scope
			req.ScopeID = scope.scopeID
			
			if err := rc.budgetManager.RecordUsage(req); err != nil {
				return fmt.Errorf("failed to record retry usage for scope %s:%s: %w", scope.scope, scope.scopeID, err)
			}
		}
	}
	
	return nil
}

// GetRetryBudgetStatus returns current retry budget status
func (rc *RetryController) GetRetryBudgetStatus(scope BudgetScope, scopeID string) (*RetryBudgetStatus, error) {
	budgets := rc.budgetManager.ListBudgets()
	
	status := &RetryBudgetStatus{
		Scope:   scope,
		ScopeID: scopeID,
		Budgets: make([]RetryBudgetInfo, 0),
	}
	
	for _, budget := range budgets {
		if budget.Type != BudgetTypeRetries {
			continue
		}
		
		if budget.Scope != scope {
			continue
		}
		
		if scope != BudgetScopeGlobal && budget.ScopeID != scopeID {
			continue
		}
		
		info := RetryBudgetInfo{
			BudgetID:        budget.ID,
			Name:            budget.Name,
			MaxCount:        budget.MaxCount,
			CurrentCount:    budget.CurrentCount,
			TimeWindow:      budget.TimeWindow,
			WindowStartTime: time.Now(), // This should come from usage data
			ResetTime:       time.Now().Add(time.Hour), // Calculate based on window
		}
		
		if budget.MaxCount > 0 {
			info.PercentageUsed = float64(budget.CurrentCount) / float64(budget.MaxCount) * 100
			info.RemainingCount = budget.MaxCount - budget.CurrentCount
			if info.RemainingCount < 0 {
				info.RemainingCount = 0
			}
		}
		
		status.Budgets = append(status.Budgets, info)
	}
	
	return status, nil
}

// RetryBudgetStatus represents the current status of retry budgets
type RetryBudgetStatus struct {
	Scope   BudgetScope       `json:"scope"`
	ScopeID string            `json:"scopeId"`
	Budgets []RetryBudgetInfo `json:"budgets"`
}

// RetryBudgetInfo contains information about a specific retry budget
type RetryBudgetInfo struct {
	BudgetID        string      `json:"budgetId"`
	Name            string      `json:"name"`
	MaxCount        int64       `json:"maxCount"`
	CurrentCount    int64       `json:"currentCount"`
	RemainingCount  int64       `json:"remainingCount"`
	PercentageUsed  float64     `json:"percentageUsed"`
	TimeWindow      TimeWindow  `json:"timeWindow"`
	WindowStartTime time.Time   `json:"windowStartTime"`
	ResetTime       time.Time   `json:"resetTime"`
}

// CreateDefaultRetryBudgets creates default retry budgets for common scenarios
func (rc *RetryController) CreateDefaultRetryBudgets() error {
	defaultBudgets := []*Budget{
		{
			ID:          "global-retry-budget",
			Name:        "Global Retry Budget",
			Description: "Global limit on retry operations per hour",
			Type:        BudgetTypeRetries,
			Scope:       BudgetScopeGlobal,
			MaxCount:    10000, // 10k retries per hour globally
			TimeWindow:  TimeWindowHour,
			Enabled:     true,
		},
		{
			ID:          "user-retry-budget",
			Name:        "Per-User Retry Budget",
			Description: "Limit retry operations per user per hour",
			Type:        BudgetTypeRetries,
			Scope:       BudgetScopeUser,
			MaxCount:    100, // 100 retries per user per hour
			TimeWindow:  TimeWindowHour,
			Enabled:     true,
		},
		{
			ID:          "workflow-retry-budget",
			Name:        "Per-Workflow Retry Budget",
			Description: "Limit retry operations per workflow per execution",
			Type:        BudgetTypeRetries,
			Scope:       BudgetScopeWorkflow,
			MaxCount:    50, // 50 retries per workflow execution
			TimeWindow:  TimeWindowHour,
			Enabled:     true,
		},
		{
			ID:          "session-retry-budget",
			Name:        "Per-Session Retry Budget",
			Description: "Limit retry operations per session",
			Type:        BudgetTypeRetries,
			Scope:       BudgetScopeSession,
			MaxCount:    20, // 20 retries per session
			TimeWindow:  TimeWindowHour,
			Enabled:     true,
		},
	}
	
	for _, budget := range defaultBudgets {
		if err := rc.budgetManager.AddBudget(budget); err != nil {
			return fmt.Errorf("failed to create default retry budget %s: %w", budget.ID, err)
		}
	}
	
	return nil
}

// RetryBudgetMiddleware provides middleware for retry budget enforcement
type RetryBudgetMiddleware struct {
	controller *RetryController
}

// NewRetryBudgetMiddleware creates new retry budget middleware
func NewRetryBudgetMiddleware(controller *RetryController) *RetryBudgetMiddleware {
	return &RetryBudgetMiddleware{
		controller: controller,
	}
}

// EnforceRetryBudget enforces retry budget constraints
func (m *RetryBudgetMiddleware) EnforceRetryBudget(ctx context.Context, retryCtx *RetryContext) error {
	response, err := m.controller.CheckRetryAllowed(ctx, retryCtx)
	if err != nil {
		return fmt.Errorf("retry budget check failed: %w", err)
	}
	
	if !response.Allowed {
		// Create detailed error message
		msg := "Retry operation blocked due to budget constraints"
		if len(response.Violations) > 0 {
			violation := response.Violations[0] // Use first violation
			msg = fmt.Sprintf("Retry budget exceeded: %s (%s)", violation.Message, violation.BudgetName)
		}
		
		return &RetryBudgetError{
			Message:    msg,
			Violations: response.Violations,
			ResetTime:  response.ResetTime,
		}
	}
	
	// Record the retry usage
	if err := m.controller.RecordRetryUsage(ctx, retryCtx); err != nil {
		// Log error but don't fail the operation
		// In production, this would use proper logging
		fmt.Printf("Warning: failed to record retry usage: %v\n", err)
	}
	
	return nil
}

// RetryBudgetError represents a retry budget violation error
type RetryBudgetError struct {
	Message    string            `json:"message"`
	Violations []BudgetViolation `json:"violations"`
	ResetTime  time.Time         `json:"resetTime"`
}

func (e *RetryBudgetError) Error() string {
	return e.Message
}

// IsRetryBudgetError checks if an error is a retry budget error
func IsRetryBudgetError(err error) bool {
	_, ok := err.(*RetryBudgetError)
	return ok
}

