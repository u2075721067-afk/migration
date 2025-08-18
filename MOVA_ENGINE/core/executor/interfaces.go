package executor

import "context"

// ExecutorInterface defines the interface for workflow execution
type ExecutorInterface interface {
	Execute(ctx context.Context, envelope MOVAEnvelope) (*ExecutionContext, error)
	GetExecutionStatus(runID string) (*ExecutionContext, error)
	GetExecutionLogs(runID string) ([]string, error)
	CancelExecution(runID string) error
}

// ValidatorInterface defines the interface for envelope validation
type ValidatorInterface interface {
	ValidateEnvelope(file string) (bool, []error)
}
