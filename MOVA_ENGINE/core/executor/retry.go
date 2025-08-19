package executor

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// BackoffType defines the backoff strategy
type BackoffType string

const (
	BackoffLinear      BackoffType = "linear"
	BackoffExponential BackoffType = "exponential"
	BackoffFixed       BackoffType = "fixed"
)

// ExtendedRetryPolicy adds additional fields to the base RetryPolicy
type ExtendedRetryPolicy struct {
	*RetryPolicy
	MaxDelay time.Duration `json:"max_delay,omitempty"`
	Jitter   bool          `json:"jitter,omitempty"`
}

// RetryManager handles retry logic for failed actions
type RetryManager struct {
	dlq    *DeadLetterQueue
	logger *logrus.Logger
}

// NewRetryManager creates a new retry manager
func NewRetryManager(dlq *DeadLetterQueue) *RetryManager {
	return &RetryManager{
		dlq:    dlq,
		logger: logrus.New(),
	}
}

// ExecuteWithRetry executes an action with retry logic
func (rm *RetryManager) ExecuteWithRetry(ctx context.Context, action Action, executor ExecutorInterface, envelope MOVAEnvelope, execCtx *ExecutionContext) error {
	retryPolicy := rm.getExtendedRetryPolicy(action, envelope)

	var lastErr error
	var errorHistory []string

	for attempt := 1; attempt <= retryPolicy.MaxAttempts; attempt++ {
		rm.logger.WithFields(logrus.Fields{
			"action_name":  action.Name,
			"attempt":      attempt,
			"max_attempts": retryPolicy.MaxAttempts,
		}).Debug("Executing action with retry")

		// Execute the action
		err := rm.executeAction(ctx, action, executor, envelope, execCtx)
		if err == nil {
			// Success - update metrics and return
			if attempt > 1 {
				rm.logger.WithFields(logrus.Fields{
					"action_name": action.Name,
					"attempts":    attempt,
				}).Info("Action succeeded after retry")
			}
			return nil
		}

		// Record the error
		lastErr = err
		errorHistory = append(errorHistory, fmt.Sprintf("Attempt %d: %s", attempt, err.Error()))

		rm.logger.WithFields(logrus.Fields{
			"action_name": action.Name,
			"attempt":     attempt,
			"error":       err.Error(),
		}).Warn("Action attempt failed")

		// If this was the last attempt, move to DLQ
		if attempt == retryPolicy.MaxAttempts {
			rm.moveToDLQ(envelope, *execCtx, &action, lastErr, errorHistory, retryPolicy)
			return fmt.Errorf("action failed after %d attempts: %w", attempt, lastErr)
		}

		// Calculate delay and wait
		delay := rm.calculateDelay(retryPolicy, attempt)
		rm.logger.WithFields(logrus.Fields{
			"action_name": action.Name,
			"delay":       delay,
		}).Debug("Waiting before retry")

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return lastErr
}

// getExtendedRetryPolicy extracts retry policy from action or envelope and extends it
func (rm *RetryManager) getExtendedRetryPolicy(action Action, envelope MOVAEnvelope) *ExtendedRetryPolicy {
	var basePolicy *RetryPolicy

	// Action-level retry policy takes precedence
	if action.Retry != nil {
		basePolicy = action.Retry
	} else if envelope.Intent.Retry != nil {
		// Fall back to envelope-level retry policy
		basePolicy = envelope.Intent.Retry
	} else {
		// Default retry policy
		basePolicy = &RetryPolicy{
			MaxAttempts: 3,
			Backoff:     "exponential",
			Delay:       2, // seconds
		}
	}

	// Create extended policy with defaults
	return &ExtendedRetryPolicy{
		RetryPolicy: basePolicy,
		MaxDelay:    time.Minute * 5,
		Jitter:      true,
	}
}

// calculateDelay calculates the delay for the next retry attempt
func (rm *RetryManager) calculateDelay(policy *ExtendedRetryPolicy, attempt int) time.Duration {
	baseDelay := time.Duration(policy.Delay) * time.Second
	var delay time.Duration

	switch policy.Backoff {
	case "fixed":
		delay = baseDelay
	case "linear":
		delay = time.Duration(attempt) * baseDelay
	case "exponential":
		delay = time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
	default:
		delay = baseDelay
	}

	// Apply maximum delay limit
	if policy.MaxDelay > 0 && delay > policy.MaxDelay {
		delay = policy.MaxDelay
	}

	// Apply jitter if enabled
	if policy.Jitter {
		jitterAmount := time.Duration(rand.Float64() * float64(delay) * 0.1)
		delay += jitterAmount
	}

	return delay
}

// executeAction executes a single action
func (rm *RetryManager) executeAction(ctx context.Context, action Action, executor ExecutorInterface, envelope MOVAEnvelope, execCtx *ExecutionContext) error {
	// This would delegate to the actual action execution logic
	// For now, we'll simulate the execution

	startTime := time.Now()

	// Update execution context
	result := ActionResult{
		ActionName: action.Name,
		Status:     ActionStatusRunning,
		StartTime:  startTime,
		Attempts:   1,
	}

	if execCtx.Results == nil {
		execCtx.Results = make(map[string]ActionResult)
	}
	execCtx.Results[action.Name] = result

	// Simulate action execution based on type
	err := rm.simulateActionExecution(ctx, action)

	// Update result
	endTime := time.Now()
	result.EndTime = &endTime

	if err != nil {
		result.Status = ActionStatusFailed
		result.Error = err.Error()
	} else {
		result.Status = ActionStatusCompleted
		result.Output = map[string]interface{}{
			"success":  true,
			"duration": endTime.Sub(startTime).Milliseconds(),
		}
	}

	execCtx.Results[action.Name] = result

	return err
}

// simulateActionExecution simulates action execution for different action types
func (rm *RetryManager) simulateActionExecution(ctx context.Context, action Action) error {
	// Simulate different failure scenarios based on action type
	switch action.Type {
	case "http_fetch":
		// Simulate network failures
		if rand.Float64() < 0.3 { // 30% failure rate
			return fmt.Errorf("network timeout: failed to connect to remote server")
		}
	case "parse_json":
		// Simulate parsing failures
		if rand.Float64() < 0.1 { // 10% failure rate
			return fmt.Errorf("invalid JSON format: unexpected character at position 42")
		}
	case "set":
		// Set actions rarely fail
		if rand.Float64() < 0.01 { // 1% failure rate
			return fmt.Errorf("variable assignment failed: invalid expression")
		}
	}

	// Simulate processing time
	processingTime := time.Duration(rand.Intn(100)) * time.Millisecond
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(processingTime):
		return nil
	}
}

// moveToDLQ moves a failed workflow to the dead letter queue
func (rm *RetryManager) moveToDLQ(envelope MOVAEnvelope, execCtx ExecutionContext, failedAction *Action, lastErr error, errorHistory []string, retryPolicy *ExtendedRetryPolicy) {
	errorDetails := ErrorDetails{
		LastError:     lastErr.Error(),
		ErrorHistory:  errorHistory,
		FailureReason: "max_retries_exceeded",
		Attempts:      retryPolicy.MaxAttempts,
		RetryPolicy:   retryPolicy.RetryPolicy,
		Environment: map[string]string{
			"go_version":       "1.23.0",
			"executor_version": "1.0.0",
		},
	}

	entry, err := rm.dlq.Add(envelope, execCtx, failedAction, errorDetails)
	if err != nil {
		rm.logger.WithFields(logrus.Fields{
			"run_id": execCtx.RunID,
			"error":  err.Error(),
		}).Error("Failed to add workflow to DLQ")
		return
	}

	rm.logger.WithFields(logrus.Fields{
		"dlq_id":      entry.ID,
		"run_id":      execCtx.RunID,
		"action_name": failedAction.Name,
		"attempts":    retryPolicy.MaxAttempts,
		"last_error":  lastErr.Error(),
	}).Warn("Workflow moved to dead letter queue after max retries exceeded")
}

// RetryFromDLQ retries a workflow from the DLQ in sandbox mode
func (rm *RetryManager) RetryFromDLQ(ctx context.Context, dlqID string, executor ExecutorInterface, sandboxMode bool) (*ExecutionContext, error) {
	// Get the DLQ entry
	entry, err := rm.dlq.Get(dlqID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DLQ entry: %w", err)
	}

	// Update DLQ status
	if err := rm.dlq.UpdateStatus(dlqID, DLQStatusRetrying); err != nil {
		rm.logger.Warnf("Failed to update DLQ status: %v", err)
	}

	// Create new execution context for retry
	retryCtx := ExecutionContext{
		RunID:      fmt.Sprintf("retry_%s_%d", entry.RunID, time.Now().Unix()),
		WorkflowID: entry.Context.WorkflowID,
		StartTime:  time.Now(),
		Status:     StatusRunning,
		Variables:  make(map[string]interface{}),
		Results:    make(map[string]ActionResult),
		Logs:       []ExecutionLog{},
	}

	// Copy original variables
	for k, v := range entry.Context.Variables {
		retryCtx.Variables[k] = v
	}

	// Add retry metadata
	retryCtx.Variables["__retry_metadata"] = map[string]interface{}{
		"original_run_id": entry.RunID,
		"dlq_id":          dlqID,
		"sandbox_mode":    sandboxMode,
		"retry_count":     entry.Metadata.RetryCount,
	}

	rm.logger.WithFields(logrus.Fields{
		"dlq_id":       dlqID,
		"original_run": entry.RunID,
		"retry_run":    retryCtx.RunID,
		"sandbox":      sandboxMode,
	}).Info("Starting DLQ retry execution")

	// Execute the workflow
	_, err = executor.Execute(ctx, entry.Envelope)
	if err != nil {
		retryCtx.Status = StatusFailed
		rm.logger.WithFields(logrus.Fields{
			"dlq_id": dlqID,
			"error":  err.Error(),
		}).Error("DLQ retry execution failed")
		return &retryCtx, err
	}

	// Update DLQ status based on result
	if err == nil {
		if err := rm.dlq.UpdateStatus(dlqID, DLQStatusResolved); err != nil {
			rm.logger.Warnf("Failed to update DLQ status to resolved: %v", err)
		}
		retryCtx.Status = StatusCompleted
	}

	now := time.Now()
	retryCtx.EndTime = &now

	rm.logger.WithFields(logrus.Fields{
		"dlq_id":    dlqID,
		"retry_run": retryCtx.RunID,
		"status":    retryCtx.Status,
	}).Info("DLQ retry execution completed")

	return &retryCtx, nil
}
