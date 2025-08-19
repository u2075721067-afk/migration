package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// DeadLetterEntry represents an entry in the dead letter queue
type DeadLetterEntry struct {
	ID           string           `json:"id"`
	RunID        string           `json:"run_id"`
	CreatedAt    time.Time        `json:"created_at"`
	Envelope     MOVAEnvelope     `json:"envelope"`
	Context      ExecutionContext `json:"context"`
	FailedAction *Action          `json:"failed_action,omitempty"`
	ErrorDetails ErrorDetails     `json:"error_details"`
	Metadata     DLQMetadata      `json:"metadata"`
	Status       DLQStatus        `json:"status"`
}

// ErrorDetails contains detailed error information
type ErrorDetails struct {
	LastError     string            `json:"last_error"`
	ErrorHistory  []string          `json:"error_history"`
	FailureReason string            `json:"failure_reason"`
	Attempts      int               `json:"attempts"`
	RetryPolicy   *RetryPolicy      `json:"retry_policy,omitempty"`
	StackTrace    string            `json:"stack_trace,omitempty"`
	Environment   map[string]string `json:"environment,omitempty"`
}

// DLQMetadata contains metadata about the DLQ entry
type DLQMetadata struct {
	Source       string                 `json:"source"`
	Priority     int                    `json:"priority"`
	Tags         []string               `json:"tags"`
	UserID       string                 `json:"user_id,omitempty"`
	WorkflowType string                 `json:"workflow_type"`
	RetryCount   int                    `json:"retry_count"`
	LastRetryAt  *time.Time             `json:"last_retry_at,omitempty"`
	CustomData   map[string]interface{} `json:"custom_data,omitempty"`
}

// DLQStatus represents the status of a DLQ entry
type DLQStatus string

const (
	DLQStatusActive   DLQStatus = "active"
	DLQStatusRetrying DLQStatus = "retrying"
	DLQStatusResolved DLQStatus = "resolved"
	DLQStatusArchived DLQStatus = "archived"
)

// DeadLetterQueue manages failed workflow entries
type DeadLetterQueue struct {
	StoragePath string
	logger      *logrus.Logger
}

// NewDeadLetterQueue creates a new DLQ instance
func NewDeadLetterQueue(storagePath string) *DeadLetterQueue {
	dlq := &DeadLetterQueue{
		StoragePath: storagePath,
		logger:      logrus.New(),
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		dlq.logger.Errorf("Failed to create DLQ storage directory: %v", err)
	}

	return dlq
}

// Add adds a failed workflow to the DLQ
func (dlq *DeadLetterQueue) Add(envelope MOVAEnvelope, context ExecutionContext, failedAction *Action, errorDetails ErrorDetails) (*DeadLetterEntry, error) {
	entry := &DeadLetterEntry{
		ID:           uuid.New().String(),
		RunID:        context.RunID,
		CreatedAt:    time.Now(),
		Envelope:     envelope,
		Context:      context,
		FailedAction: failedAction,
		ErrorDetails: errorDetails,
		Metadata: DLQMetadata{
			Source:       "executor",
			Priority:     1,
			WorkflowType: envelope.Intent.Name,
			RetryCount:   0,
			Tags:         envelope.Intent.Tags,
		},
		Status: DLQStatusActive,
	}

	if err := dlq.store(entry); err != nil {
		return nil, fmt.Errorf("failed to store DLQ entry: %w", err)
	}

	dlq.logger.WithFields(logrus.Fields{
		"dlq_id": entry.ID,
		"run_id": entry.RunID,
		"error":  errorDetails.LastError,
	}).Warn("Workflow moved to dead letter queue")

	return entry, nil
}

// Get retrieves a DLQ entry by ID
func (dlq *DeadLetterQueue) Get(id string) (*DeadLetterEntry, error) {
	filePath := filepath.Join(dlq.StoragePath, fmt.Sprintf("%s.json", id))

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("DLQ entry not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read DLQ entry: %w", err)
	}

	var entry DeadLetterEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DLQ entry: %w", err)
	}

	return &entry, nil
}

// List returns all DLQ entries with optional filtering
func (dlq *DeadLetterQueue) List(filter DLQFilter) ([]*DeadLetterEntry, error) {
	files, err := filepath.Glob(filepath.Join(dlq.StoragePath, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list DLQ files: %w", err)
	}

	var entries []*DeadLetterEntry
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			dlq.logger.Warnf("Failed to read DLQ file %s: %v", file, err)
			continue
		}

		var entry DeadLetterEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			dlq.logger.Warnf("Failed to unmarshal DLQ file %s: %v", file, err)
			continue
		}

		if filter.Matches(&entry) {
			entries = append(entries, &entry)
		}
	}

	return entries, nil
}

// UpdateStatus updates the status of a DLQ entry
func (dlq *DeadLetterQueue) UpdateStatus(id string, status DLQStatus) error {
	entry, err := dlq.Get(id)
	if err != nil {
		return err
	}

	entry.Status = status
	if status == DLQStatusRetrying {
		now := time.Now()
		entry.Metadata.LastRetryAt = &now
		entry.Metadata.RetryCount++
	}

	return dlq.store(entry)
}

// Archive moves a DLQ entry to archived status
func (dlq *DeadLetterQueue) Archive(id string) error {
	return dlq.UpdateStatus(id, DLQStatusArchived)
}

// Delete removes a DLQ entry
func (dlq *DeadLetterQueue) Delete(id string) error {
	filePath := filepath.Join(dlq.StoragePath, fmt.Sprintf("%s.json", id))

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("DLQ entry not found: %s", id)
		}
		return fmt.Errorf("failed to delete DLQ entry: %w", err)
	}

	dlq.logger.WithField("dlq_id", id).Info("DLQ entry deleted")
	return nil
}

// store persists a DLQ entry to disk
func (dlq *DeadLetterQueue) store(entry *DeadLetterEntry) error {
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal DLQ entry: %w", err)
	}

	filePath := filepath.Join(dlq.StoragePath, fmt.Sprintf("%s.json", entry.ID))
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write DLQ entry: %w", err)
	}

	return nil
}

// DLQFilter provides filtering options for DLQ queries
type DLQFilter struct {
	Status       *DLQStatus `json:"status,omitempty"`
	WorkflowType string     `json:"workflow_type,omitempty"`
	UserID       string     `json:"user_id,omitempty"`
	Since        *time.Time `json:"since,omitempty"`
	Until        *time.Time `json:"until,omitempty"`
	Limit        int        `json:"limit,omitempty"`
}

// Matches checks if an entry matches the filter criteria
func (f *DLQFilter) Matches(entry *DeadLetterEntry) bool {
	if f.Status != nil && entry.Status != *f.Status {
		return false
	}

	if f.WorkflowType != "" && entry.Metadata.WorkflowType != f.WorkflowType {
		return false
	}

	if f.UserID != "" && entry.Metadata.UserID != f.UserID {
		return false
	}

	if f.Since != nil && entry.CreatedAt.Before(*f.Since) {
		return false
	}

	if f.Until != nil && entry.CreatedAt.After(*f.Until) {
		return false
	}

	return true
}

