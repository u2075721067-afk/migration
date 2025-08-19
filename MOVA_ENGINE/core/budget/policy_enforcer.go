package budget

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PolicyEnforcer provides centralized policy enforcement
type PolicyEnforcer struct {
	budgetManager     *Manager
	retryController   *RetryController
	resourceMonitor   *ResourceMonitor
	
	// Policy configurations
	policies          map[string]*EnforcementPolicy
	policyRules       map[string][]PolicyRule
	
	// State
	mu                sync.RWMutex
	enforcementStats  map[string]*EnforcementStats
}

// NewPolicyEnforcer creates a new policy enforcer
func NewPolicyEnforcer(budgetManager *Manager, retryController *RetryController, resourceMonitor *ResourceMonitor) *PolicyEnforcer {
	return &PolicyEnforcer{
		budgetManager:    budgetManager,
		retryController:  retryController,
		resourceMonitor:  resourceMonitor,
		policies:         make(map[string]*EnforcementPolicy),
		policyRules:      make(map[string][]PolicyRule),
		enforcementStats: make(map[string]*EnforcementStats),
	}
}

// EnforcementPolicy represents a policy enforcement configuration
type EnforcementPolicy struct {
	ID          string    `json:"id" yaml:"id"`
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description" yaml:"description"`
	
	// Enforcement settings
	Enabled     bool      `json:"enabled" yaml:"enabled"`
	Priority    int       `json:"priority" yaml:"priority"` // Higher priority = enforced first
	
	// Scope
	Scope       BudgetScope `json:"scope" yaml:"scope"`
	ScopeID     string      `json:"scopeId" yaml:"scopeId"`
	
	// Budget types to enforce
	BudgetTypes []BudgetType `json:"budgetTypes" yaml:"budgetTypes"`
	
	// Actions on violation
	Actions     []EnforcementAction `json:"actions" yaml:"actions"`
	
	// Conditions
	Conditions  []PolicyCondition   `json:"conditions" yaml:"conditions"`
	
	CreatedAt   time.Time `json:"createdAt" yaml:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" yaml:"updatedAt"`
}

// PolicyRule represents a rule within a policy
type PolicyRule struct {
	ID          string      `json:"id"`
	PolicyID    string      `json:"policyId"`
	Type        string      `json:"type"` // "threshold", "quota", "rate_limit"
	
	// Rule parameters
	Threshold   float64     `json:"threshold,omitempty"`   // Percentage threshold (0-1)
	Value       interface{} `json:"value,omitempty"`       // Rule value
	Operator    string      `json:"operator"`              // "gt", "lt", "eq", "gte", "lte"
	
	// Actions
	Actions     []EnforcementAction `json:"actions"`
	
	Enabled     bool        `json:"enabled"`
}

// PolicyCondition represents conditions for policy application
type PolicyCondition struct {
	Type        string      `json:"type"`        // "time", "user", "organization", "workflow"
	Operator    string      `json:"operator"`    // "eq", "in", "matches"
	Value       interface{} `json:"value"`       // Condition value
	
	// Time-based conditions
	TimeRange   *TimeRange  `json:"timeRange,omitempty"`
	DaysOfWeek  []string    `json:"daysOfWeek,omitempty"`
}

// TimeRange represents a time range condition
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// EnforcementAction represents an action to take on policy violation
type EnforcementAction struct {
	Type        string                 `json:"type"`        // "block", "throttle", "alert", "log"
	Parameters  map[string]interface{} `json:"parameters"`  // Action-specific parameters
	Severity    string                 `json:"severity"`    // "low", "medium", "high", "critical"
}

// EnforcementStats tracks policy enforcement statistics
type EnforcementStats struct {
	PolicyID        string    `json:"policyId"`
	TotalChecks     int64     `json:"totalChecks"`
	TotalViolations int64     `json:"totalViolations"`
	BlockedRequests int64     `json:"blockedRequests"`
	ThrottledRequests int64   `json:"throttledRequests"`
	AlertsSent      int64     `json:"alertsSent"`
	LastViolation   time.Time `json:"lastViolation"`
	LastUpdated     time.Time `json:"lastUpdated"`
}

// EnforcementContext contains context for policy enforcement
type EnforcementContext struct {
	// Request context
	RequestType    string `json:"requestType"`    // "api", "workflow", "retry"
	
	// Identity
	UserID         string `json:"userId"`
	OrganizationID string `json:"organizationId"`
	ClientID       string `json:"clientId"`
	
	// Resource context
	WorkflowID     string `json:"workflowId"`
	SessionID      string `json:"sessionId"`
	ActionType     string `json:"actionType"`
	
	// Resource requirements
	RequiredCPU    float64 `json:"requiredCPU"`
	RequiredMemory int64   `json:"requiredMemory"`
	
	// Timing
	Timestamp      time.Time `json:"timestamp"`
}

// EnforcementResult represents the result of policy enforcement
type EnforcementResult struct {
	Allowed         bool                  `json:"allowed"`
	BlockedBy       []string              `json:"blockedBy,omitempty"`       // Policy IDs that blocked the request
	ThrottledBy     []string              `json:"throttledBy,omitempty"`     // Policy IDs that throttled the request
	Violations      []BudgetViolation     `json:"violations,omitempty"`
	Actions         []EnforcementAction   `json:"actions,omitempty"`         // Actions that were triggered
	Message         string                `json:"message,omitempty"`
	RetryAfter      *time.Time            `json:"retryAfter,omitempty"`      // When the request can be retried
}

// EnforcePolicy enforces all applicable policies for a given context
func (pe *PolicyEnforcer) EnforcePolicy(ctx context.Context, enforcementCtx *EnforcementContext) (*EnforcementResult, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	
	result := &EnforcementResult{
		Allowed:     true,
		BlockedBy:   make([]string, 0),
		ThrottledBy: make([]string, 0),
		Violations:  make([]BudgetViolation, 0),
		Actions:     make([]EnforcementAction, 0),
	}
	
	// Find applicable policies
	applicablePolicies := pe.findApplicablePolicies(enforcementCtx)
	
	// Sort policies by priority
	pe.sortPoliciesByPriority(applicablePolicies)
	
	// Enforce each policy
	for _, policy := range applicablePolicies {
		if !policy.Enabled {
			continue
		}
		
		// Update stats
		pe.updateEnforcementStats(policy.ID, "check")
		
		// Check policy conditions
		if !pe.evaluatePolicyConditions(policy, enforcementCtx) {
			continue
		}
		
		// Enforce budget constraints
		policyResult, err := pe.enforcePolicyBudgets(policy, enforcementCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to enforce policy %s: %w", policy.ID, err)
		}
		
		// Merge results
		result.Violations = append(result.Violations, policyResult.Violations...)
		result.Actions = append(result.Actions, policyResult.Actions...)
		
		// Handle blocking actions
		if pe.containsBlockingAction(policyResult.Actions) {
			result.Allowed = false
			result.BlockedBy = append(result.BlockedBy, policy.ID)
			pe.updateEnforcementStats(policy.ID, "blocked")
		}
		
		// Handle throttling actions
		if pe.containsThrottlingAction(policyResult.Actions) {
			result.ThrottledBy = append(result.ThrottledBy, policy.ID)
			pe.updateEnforcementStats(policy.ID, "throttled")
		}
		
		// Record violations
		if len(policyResult.Violations) > 0 {
			pe.updateEnforcementStats(policy.ID, "violation")
		}
	}
	
	// Generate result message
	result.Message = pe.generateResultMessage(result)
	
	return result, nil
}

// findApplicablePolicies finds policies that apply to the enforcement context
func (pe *PolicyEnforcer) findApplicablePolicies(enforcementCtx *EnforcementContext) []*EnforcementPolicy {
	applicable := make([]*EnforcementPolicy, 0)
	
	for _, policy := range pe.policies {
		if pe.isPolicyApplicable(policy, enforcementCtx) {
			applicable = append(applicable, policy)
		}
	}
	
	return applicable
}

// isPolicyApplicable checks if a policy applies to the given context
func (pe *PolicyEnforcer) isPolicyApplicable(policy *EnforcementPolicy, enforcementCtx *EnforcementContext) bool {
	// Check scope matching
	switch policy.Scope {
	case BudgetScopeGlobal:
		return true
	case BudgetScopeOrganization:
		return policy.ScopeID == enforcementCtx.OrganizationID
	case BudgetScopeUser:
		return policy.ScopeID == enforcementCtx.UserID
	case BudgetScopeWorkflow:
		return policy.ScopeID == enforcementCtx.WorkflowID
	case BudgetScopeSession:
		return policy.ScopeID == enforcementCtx.SessionID
	default:
		return false
	}
}

// evaluatePolicyConditions evaluates policy conditions
func (pe *PolicyEnforcer) evaluatePolicyConditions(policy *EnforcementPolicy, enforcementCtx *EnforcementContext) bool {
	for _, condition := range policy.Conditions {
		if !pe.evaluateCondition(condition, enforcementCtx) {
			return false
		}
	}
	return true
}

// evaluateCondition evaluates a single condition
func (pe *PolicyEnforcer) evaluateCondition(condition PolicyCondition, enforcementCtx *EnforcementContext) bool {
	switch condition.Type {
	case "time":
		return pe.evaluateTimeCondition(condition, enforcementCtx.Timestamp)
	case "user":
		return pe.evaluateStringCondition(condition, enforcementCtx.UserID)
	case "organization":
		return pe.evaluateStringCondition(condition, enforcementCtx.OrganizationID)
	case "workflow":
		return pe.evaluateStringCondition(condition, enforcementCtx.WorkflowID)
	case "request_type":
		return pe.evaluateStringCondition(condition, enforcementCtx.RequestType)
	default:
		return true // Unknown conditions default to true
	}
}

// evaluateTimeCondition evaluates time-based conditions
func (pe *PolicyEnforcer) evaluateTimeCondition(condition PolicyCondition, timestamp time.Time) bool {
	if condition.TimeRange != nil {
		if timestamp.Before(condition.TimeRange.Start) || timestamp.After(condition.TimeRange.End) {
			return false
		}
	}
	
	if len(condition.DaysOfWeek) > 0 {
		weekday := timestamp.Weekday().String()
		found := false
		for _, day := range condition.DaysOfWeek {
			if day == weekday {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	return true
}

// evaluateStringCondition evaluates string-based conditions
func (pe *PolicyEnforcer) evaluateStringCondition(condition PolicyCondition, value string) bool {
	conditionValue, ok := condition.Value.(string)
	if !ok {
		return true // Invalid condition value, default to true
	}
	
	switch condition.Operator {
	case "eq":
		return value == conditionValue
	case "ne":
		return value != conditionValue
	// Add more operators as needed
	default:
		return true
	}
}

// enforcePolicyBudgets enforces budget constraints for a policy
func (pe *PolicyEnforcer) enforcePolicyBudgets(policy *EnforcementPolicy, enforcementCtx *EnforcementContext) (*EnforcementResult, error) {
	result := &EnforcementResult{
		Allowed:    true,
		Violations: make([]BudgetViolation, 0),
		Actions:    make([]EnforcementAction, 0),
	}
	
	// Check each budget type specified in the policy
	for _, budgetType := range policy.BudgetTypes {
		budgetResult, err := pe.checkBudgetType(budgetType, policy, enforcementCtx)
		if err != nil {
			return nil, err
		}
		
		result.Violations = append(result.Violations, budgetResult.Violations...)
		
		// If budget check failed, apply policy actions
		if !budgetResult.Allowed {
			result.Actions = append(result.Actions, policy.Actions...)
		}
	}
	
	return result, nil
}

// checkBudgetType checks a specific budget type
func (pe *PolicyEnforcer) checkBudgetType(budgetType BudgetType, policy *EnforcementPolicy, enforcementCtx *EnforcementContext) (*BudgetCheckResponse, error) {
	req := &BudgetCheckRequest{
		Type:           budgetType,
		Scope:          policy.Scope,
		ScopeID:        policy.ScopeID,
		Count:          1, // Default to 1 operation
		WorkflowID:     enforcementCtx.WorkflowID,
		SessionID:      enforcementCtx.SessionID,
		UserID:         enforcementCtx.UserID,
		OrganizationID: enforcementCtx.OrganizationID,
	}
	
	// Set resource-specific parameters
	switch budgetType {
	case BudgetTypeCPU:
		req.CPU = enforcementCtx.RequiredCPU
	case BudgetTypeMemory:
		req.Memory = enforcementCtx.RequiredMemory
	}
	
	return pe.budgetManager.CheckBudget(req)
}

// Helper methods
func (pe *PolicyEnforcer) sortPoliciesByPriority(policies []*EnforcementPolicy) {
	// Simple bubble sort by priority (descending)
	n := len(policies)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if policies[j].Priority < policies[j+1].Priority {
				policies[j], policies[j+1] = policies[j+1], policies[j]
			}
		}
	}
}

func (pe *PolicyEnforcer) containsBlockingAction(actions []EnforcementAction) bool {
	for _, action := range actions {
		if action.Type == "block" {
			return true
		}
	}
	return false
}

func (pe *PolicyEnforcer) containsThrottlingAction(actions []EnforcementAction) bool {
	for _, action := range actions {
		if action.Type == "throttle" {
			return true
		}
	}
	return false
}

func (pe *PolicyEnforcer) generateResultMessage(result *EnforcementResult) string {
	if result.Allowed {
		if len(result.ThrottledBy) > 0 {
			return fmt.Sprintf("Request allowed but throttled by policies: %v", result.ThrottledBy)
		}
		return "Request allowed"
	} else {
		return fmt.Sprintf("Request blocked by policies: %v", result.BlockedBy)
	}
}

func (pe *PolicyEnforcer) updateEnforcementStats(policyID, eventType string) {
	if stats, exists := pe.enforcementStats[policyID]; exists {
		stats.LastUpdated = time.Now()
		
		switch eventType {
		case "check":
			stats.TotalChecks++
		case "violation":
			stats.TotalViolations++
			stats.LastViolation = time.Now()
		case "blocked":
			stats.BlockedRequests++
		case "throttled":
			stats.ThrottledRequests++
		case "alert":
			stats.AlertsSent++
		}
	} else {
		// Create new stats
		pe.enforcementStats[policyID] = &EnforcementStats{
			PolicyID:    policyID,
			LastUpdated: time.Now(),
		}
	}
}

// AddPolicy adds a new enforcement policy
func (pe *PolicyEnforcer) AddPolicy(policy *EnforcementPolicy) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	
	if policy.ID == "" {
		return fmt.Errorf("policy ID is required")
	}
	
	now := time.Now()
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = now
	}
	policy.UpdatedAt = now
	
	pe.policies[policy.ID] = policy
	
	// Initialize stats
	pe.enforcementStats[policy.ID] = &EnforcementStats{
		PolicyID:    policy.ID,
		LastUpdated: now,
	}
	
	return nil
}

// RemovePolicy removes an enforcement policy
func (pe *PolicyEnforcer) RemovePolicy(policyID string) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	
	delete(pe.policies, policyID)
	delete(pe.enforcementStats, policyID)
	delete(pe.policyRules, policyID)
	
	return nil
}

// GetEnforcementStats returns enforcement statistics for all policies
func (pe *PolicyEnforcer) GetEnforcementStats() map[string]*EnforcementStats {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	
	// Return a copy
	stats := make(map[string]*EnforcementStats)
	for id, stat := range pe.enforcementStats {
		statCopy := *stat
		stats[id] = &statCopy
	}
	
	return stats
}

