package configmanager

import (
	"fmt"
)

// Validator implements the ConfigValidator interface
type Validator struct{}

// NewValidator creates a new configuration validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates a ConfigBundle
func (v *Validator) Validate(bundle *ConfigBundle) ([]ImportError, error) {
	var errors []ImportError

	// Validate metadata
	if bundle.Metadata.Version == "" {
		errors = append(errors, ImportError{
			Type:    "metadata",
			Message: "version is required",
		})
	}

	// Validate policies
	for i, policy := range bundle.Policies {
		policyErrors := v.ValidatePolicy(&policy)
		for _, err := range policyErrors {
			err.Line = i + 1
			errors = append(errors, err)
		}
	}

	// Validate budgets
	for i, budget := range bundle.Budgets {
		budgetErrors := v.ValidateBudget(&budget)
		for _, err := range budgetErrors {
			err.Line = i + 1
			errors = append(errors, err)
		}
	}

	// Validate retry profiles
	for i, profile := range bundle.RetryProfiles {
		profileErrors := v.ValidateRetryProfile(&profile)
		for _, err := range profileErrors {
			err.Line = i + 1
			errors = append(errors, err)
		}
	}

	// Validate DLQ entries
	for i, entry := range bundle.DLQEntries {
		entryErrors := v.ValidateDLQEntry(&entry)
		for _, err := range entryErrors {
			err.Line = i + 1
			errors = append(errors, err)
		}
	}

	// Validate workflows
	for i, workflow := range bundle.Workflows {
		workflowErrors := v.ValidateWorkflow(&workflow)
		for _, err := range workflowErrors {
			err.Line = i + 1
			errors = append(errors, err)
		}
	}

	return errors, nil
}

// ValidatePolicy validates a single policy configuration
func (v *Validator) ValidatePolicy(policy *PolicyConfig) []ImportError {
	var errors []ImportError

	// Required fields
	if policy.ID == "" {
		errors = append(errors, ImportError{
			Type:    "policy",
			Message: "ID is required",
			Context: "policy",
		})
	}

	if policy.Name == "" {
		errors = append(errors, ImportError{
			Type:    "policy",
			Message: "name is required",
			Context: "policy",
		})
	}

	if policy.RetryProfile == "" {
		errors = append(errors, ImportError{
			Type:    "policy",
			Message: "retry profile is required",
			Context: "policy",
		})
	}

	// Validate conditions
	for i, condition := range policy.Conditions {
		conditionErrors := v.ValidateCondition(&condition)
		for _, err := range conditionErrors {
			err.Context = fmt.Sprintf("policy.%s.conditions[%d]", policy.ID, i)
			errors = append(errors, err)
		}
	}

	// Validate budget constraints
	if policy.BudgetConstraints.MaxRetriesPerWorkflow < 0 {
		errors = append(errors, ImportError{
			Type:    "policy",
			Message: "max retries per workflow cannot be negative",
			Context: fmt.Sprintf("policy.%s.budget_constraints", policy.ID),
		})
	}

	if policy.BudgetConstraints.MaxRetriesPerSession < 0 {
		errors = append(errors, ImportError{
			Type:    "policy",
			Message: "max retries per session cannot be negative",
			Context: fmt.Sprintf("policy.%s.budget_constraints", policy.ID),
		})
	}

	if policy.BudgetConstraints.MaxTotalRetryTime < 0 {
		errors = append(errors, ImportError{
			Type:    "policy",
			Message: "max total retry time cannot be negative",
			Context: fmt.Sprintf("policy.%s.budget_constraints", policy.ID),
		})
	}

	return errors
}

// ValidateCondition validates a single condition configuration
func (v *Validator) ValidateCondition(condition *ConditionConfig) []ImportError {
	var errors []ImportError

	// At least one condition field must be set
	if condition.ErrorType == "" && condition.HTTPStatus == 0 &&
		condition.ErrorMessagePattern == "" && condition.ActionType == "" {
		errors = append(errors, ImportError{
			Type:    "condition",
			Message: "at least one condition field must be set",
			Context: "condition",
		})
	}

	// Validate HTTP status if set
	if condition.HTTPStatus != 0 && (condition.HTTPStatus < 100 || condition.HTTPStatus > 599) {
		errors = append(errors, ImportError{
			Type:    "condition",
			Message: "HTTP status must be between 100 and 599",
			Context: "condition",
		})
	}

	return errors
}

// ValidateBudget validates a single budget configuration
func (v *Validator) ValidateBudget(budget *BudgetConfig) []ImportError {
	var errors []ImportError

	// Required fields
	if budget.ID == "" {
		errors = append(errors, ImportError{
			Type:    "budget",
			Message: "ID is required",
			Context: "budget",
		})
	}

	if budget.Name == "" {
		errors = append(errors, ImportError{
			Type:    "budget",
			Message: "name is required",
			Context: "budget",
		})
	}

	if budget.Type == "" {
		errors = append(errors, ImportError{
			Type:    "budget",
			Message: "type is required",
			Context: "budget",
		})
	}

	if budget.Scope == "" {
		errors = append(errors, ImportError{
			Type:    "budget",
			Message: "scope is required",
			Context: "budget",
		})
	}

	// Validate type-specific constraints
	switch budget.Type {
	case "retries", "api_requests":
		if budget.MaxCount <= 0 {
			errors = append(errors, ImportError{
				Type:    "budget",
				Message: "max count must be positive for retries and API requests",
				Context: fmt.Sprintf("budget.%s", budget.ID),
			})
		}
	case "memory":
		if budget.MaxMemory <= 0 {
			errors = append(errors, ImportError{
				Type:    "budget",
				Message: "max memory must be positive for memory budgets",
				Context: fmt.Sprintf("budget.%s", budget.ID),
			})
		}
	case "cpu":
		if budget.MaxCPU <= 0 || budget.MaxCPU > 1.0 {
			errors = append(errors, ImportError{
				Type:    "budget",
				Message: "max CPU must be between 0 and 1.0 for CPU budgets",
				Context: fmt.Sprintf("budget.%s", budget.ID),
			})
		}
	}

	// Validate time window
	if budget.TimeWindow <= 0 {
		errors = append(errors, ImportError{
			Type:    "budget",
			Message: "time window must be positive",
			Context: fmt.Sprintf("budget.%s", budget.ID),
		})
	}

	// Validate duration if set
	if budget.MaxDuration < 0 {
		errors = append(errors, ImportError{
			Type:    "budget",
			Message: "max duration cannot be negative",
			Context: fmt.Sprintf("budget.%s", budget.ID),
		})
	}

	return errors
}

// ValidateRetryProfile validates a single retry profile configuration
func (v *Validator) ValidateRetryProfile(profile *RetryProfileConfig) []ImportError {
	var errors []ImportError

	// Required fields
	if profile.Name == "" {
		errors = append(errors, ImportError{
			Type:    "retry_profile",
			Message: "name is required",
			Context: "retry_profile",
		})
	}

	// Validate numeric fields
	if profile.MaxRetries < 0 {
		errors = append(errors, ImportError{
			Type:    "retry_profile",
			Message: "max retries cannot be negative",
			Context: fmt.Sprintf("retry_profile.%s", profile.Name),
		})
	}

	if profile.InitialDelay < 0 {
		errors = append(errors, ImportError{
			Type:    "retry_profile",
			Message: "initial delay cannot be negative",
			Context: fmt.Sprintf("retry_profile.%s", profile.Name),
		})
	}

	if profile.MaxDelay < 0 {
		errors = append(errors, ImportError{
			Type:    "retry_profile",
			Message: "max delay cannot be negative",
			Context: fmt.Sprintf("retry_profile.%s", profile.Name),
		})
	}

	if profile.BackoffMultiplier <= 0 {
		errors = append(errors, ImportError{
			Type:    "retry_profile",
			Message: "backoff multiplier must be positive",
			Context: fmt.Sprintf("retry_profile.%s", profile.Name),
		})
	}

	if profile.Jitter < 0 || profile.Jitter > 1.0 {
		errors = append(errors, ImportError{
			Type:    "retry_profile",
			Message: "jitter must be between 0 and 1.0",
			Context: fmt.Sprintf("retry_profile.%s", profile.Name),
		})
	}

	if profile.Timeout < 0 {
		errors = append(errors, ImportError{
			Type:    "retry_profile",
			Message: "timeout cannot be negative",
			Context: fmt.Sprintf("retry_profile.%s", profile.Name),
		})
	}

	// Validate logical constraints
	if profile.MaxDelay < profile.InitialDelay {
		errors = append(errors, ImportError{
			Type:    "retry_profile",
			Message: "max delay must be greater than or equal to initial delay",
			Context: fmt.Sprintf("retry_profile.%s", profile.Name),
		})
	}

	return errors
}

// ValidateDLQEntry validates a single DLQ entry configuration
func (v *Validator) ValidateDLQEntry(entry *DLQEntryConfig) []ImportError {
	var errors []ImportError

	// Required fields
	if entry.ID == "" {
		errors = append(errors, ImportError{
			Type:    "dlq_entry",
			Message: "ID is required",
			Context: "dlq_entry",
		})
	}

	if entry.WorkflowID == "" {
		errors = append(errors, ImportError{
			Type:    "dlq_entry",
			Message: "workflow ID is required",
			Context: "dlq_entry",
		})
	}

	if entry.ActionID == "" {
		errors = append(errors, ImportError{
			Type:    "dlq_entry",
			Message: "action ID is required",
			Context: "dlq_entry",
		})
	}

	if entry.Error == "" {
		errors = append(errors, ImportError{
			Type:    "dlq_entry",
			Message: "error message is required",
			Context: "dlq_entry",
		})
	}

	// Validate numeric fields
	if entry.RetryCount < 0 {
		errors = append(errors, ImportError{
			Type:    "dlq_entry",
			Message: "retry count cannot be negative",
			Context: fmt.Sprintf("dlq_entry.%s", entry.ID),
		})
	}

	return errors
}

// ValidateWorkflow validates a single workflow configuration
func (v *Validator) ValidateWorkflow(workflow *WorkflowConfig) []ImportError {
	var errors []ImportError

	// Required fields
	if workflow.ID == "" {
		errors = append(errors, ImportError{
			Type:    "workflow",
			Message: "ID is required",
			Context: "workflow",
		})
	}

	if workflow.Name == "" {
		errors = append(errors, ImportError{
			Type:    "workflow",
			Message: "name is required",
			Context: "workflow",
		})
	}

	if workflow.Intent == "" {
		errors = append(errors, ImportError{
			Type:    "workflow",
			Message: "intent is required",
			Context: "workflow",
		})
	}

	// Validate actions
	if len(workflow.Actions) == 0 {
		errors = append(errors, ImportError{
			Type:    "workflow",
			Message: "at least one action is required",
			Context: fmt.Sprintf("workflow.%s", workflow.ID),
		})
	}

	for i, action := range workflow.Actions {
		actionErrors := v.ValidateAction(&action)
		for _, err := range actionErrors {
			err.Context = fmt.Sprintf("workflow.%s.actions[%d]", workflow.ID, i)
			errors = append(errors, err)
		}
	}

	return errors
}

// ValidateAction validates a single action configuration
func (v *Validator) ValidateAction(action *ActionConfig) []ImportError {
	var errors []ImportError

	// Required fields
	if action.ID == "" {
		errors = append(errors, ImportError{
			Type:    "action",
			Message: "ID is required",
			Context: "action",
		})
	}

	if action.Type == "" {
		errors = append(errors, ImportError{
			Type:    "action",
			Message: "type is required",
			Context: "action",
		})
	}

	// Validate timeout if set
	if action.Timeout < 0 {
		errors = append(errors, ImportError{
			Type:    "action",
			Message: "timeout cannot be negative",
			Context: fmt.Sprintf("action.%s", action.ID),
		})
	}

	// Validate retry configuration if set
	if action.Retry != nil {
		retryErrors := v.ValidateRetryConfig(action.Retry)
		for _, err := range retryErrors {
			err.Context = fmt.Sprintf("action.%s.retry", action.ID)
			errors = append(errors, err)
		}
	}

	return errors
}

// ValidateRetryConfig validates a single retry configuration
func (v *Validator) ValidateRetryConfig(retry *RetryConfig) []ImportError {
	var errors []ImportError

	// Validate numeric fields
	if retry.MaxRetries < 0 {
		errors = append(errors, ImportError{
			Type:    "retry_config",
			Message: "max retries cannot be negative",
			Context: "retry_config",
		})
	}

	if retry.InitialDelay < 0 {
		errors = append(errors, ImportError{
			Type:    "retry_config",
			Message: "initial delay cannot be negative",
			Context: "retry_config",
		})
	}

	if retry.MaxDelay < 0 {
		errors = append(errors, ImportError{
			Type:    "retry_config",
			Message: "max delay cannot be negative",
			Context: "retry_config",
		})
	}

	if retry.BackoffMultiplier <= 0 {
		errors = append(errors, ImportError{
			Type:    "retry_config",
			Message: "backoff multiplier must be positive",
			Context: "retry_config",
		})
	}

	if retry.Jitter < 0 || retry.Jitter > 1.0 {
		errors = append(errors, ImportError{
			Type:    "retry_config",
			Message: "jitter must be between 0 and 1.0",
			Context: "retry_config",
		})
	}

	// Validate logical constraints
	if retry.MaxDelay < retry.InitialDelay {
		errors = append(errors, ImportError{
			Type:    "retry_config",
			Message: "max delay must be greater than or equal to initial delay",
			Context: "retry_config",
		})
	}

	return errors
}
