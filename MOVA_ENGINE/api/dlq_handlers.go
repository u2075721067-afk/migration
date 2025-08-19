package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mova-engine/mova-engine/core/executor"
	"github.com/sirupsen/logrus"
)

// DLQService provides DLQ management functionality
type DLQService struct {
	dlq          *executor.DeadLetterQueue
	retryManager *executor.RetryManager
	executor     executor.ExecutorInterface
}

// NewDLQService creates a new DLQ service
func NewDLQService(dlq *executor.DeadLetterQueue, retryManager *executor.RetryManager, exec executor.ExecutorInterface) *DLQService {
	return &DLQService{
		dlq:          dlq,
		retryManager: retryManager,
		executor:     exec,
	}
}

// handleListDLQEntries returns all DLQ entries with optional filtering
func (s *DLQService) handleListDLQEntries(c *gin.Context) {
	// Parse query parameters for filtering
	filter := executor.DLQFilter{}

	if status := c.Query("status"); status != "" {
		dlqStatus := executor.DLQStatus(status)
		filter.Status = &dlqStatus
	}

	if workflowType := c.Query("workflow_type"); workflowType != "" {
		filter.WorkflowType = workflowType
	}

	if userID := c.Query("user_id"); userID != "" {
		filter.UserID = userID
	}

	if since := c.Query("since"); since != "" {
		if sinceTime, err := time.Parse(time.RFC3339, since); err == nil {
			filter.Since = &sinceTime
		}
	}

	if until := c.Query("until"); until != "" {
		if untilTime, err := time.Parse(time.RFC3339, until); err == nil {
			filter.Until = &untilTime
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if limitInt, err := strconv.Atoi(limit); err == nil {
			filter.Limit = limitInt
		}
	}

	// Get filtered entries
	entries, err := s.dlq.List(filter)
	if err != nil {
		LogError("dlq_api", "list_entries", err, logrus.Fields{
			"filter": filter,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list DLQ entries",
			"details": err.Error(),
		})
		return
	}

	// Apply limit if specified
	if filter.Limit > 0 && len(entries) > filter.Limit {
		entries = entries[:filter.Limit]
	}

	LogSystemEvent("dlq_list", logrus.Fields{
		"count":  len(entries),
		"filter": filter,
	})

	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
		"count":   len(entries),
		"filter":  filter,
	})
}

// handleGetDLQEntry returns a specific DLQ entry
func (s *DLQService) handleGetDLQEntry(c *gin.Context) {
	dlqID := c.Param("id")
	if dlqID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DLQ ID is required",
		})
		return
	}

	entry, err := s.dlq.Get(dlqID)
	if err != nil {
		LogError("dlq_api", "get_entry", err, logrus.Fields{
			"dlq_id": dlqID,
		})
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "DLQ entry not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entry)
}

// handleRetryDLQEntry retries a workflow from DLQ
func (s *DLQService) handleRetryDLQEntry(c *gin.Context) {
	dlqID := c.Param("id")
	if dlqID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DLQ ID is required",
		})
		return
	}

	// Parse request body for retry options
	var retryRequest struct {
		SandboxMode bool                   `json:"sandbox_mode"`
		Overrides   map[string]interface{} `json:"overrides,omitempty"`
	}

	if err := c.ShouldBindJSON(&retryRequest); err != nil {
		// Default to sandbox mode if no body provided
		retryRequest.SandboxMode = true
	}

	LogSystemEvent("dlq_retry_start", logrus.Fields{
		"dlq_id":       dlqID,
		"sandbox_mode": retryRequest.SandboxMode,
	})

	// Retry the workflow
	ctx := c.Request.Context()
	result, err := s.retryManager.RetryFromDLQ(ctx, dlqID, s.executor, retryRequest.SandboxMode)
	if err != nil {
		LogError("dlq_api", "retry_entry", err, logrus.Fields{
			"dlq_id": dlqID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retry DLQ entry",
			"details": err.Error(),
		})
		return
	}

	LogSystemEvent("dlq_retry_complete", logrus.Fields{
		"dlq_id":    dlqID,
		"retry_run": result.RunID,
		"status":    result.Status,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":          "DLQ entry retry initiated",
		"dlq_id":           dlqID,
		"retry_run_id":     result.RunID,
		"sandbox_mode":     retryRequest.SandboxMode,
		"execution_result": result,
	})
}

// handleUpdateDLQStatus updates the status of a DLQ entry
func (s *DLQService) handleUpdateDLQStatus(c *gin.Context) {
	dlqID := c.Param("id")
	if dlqID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DLQ ID is required",
		})
		return
	}

	var statusRequest struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&statusRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate status
	status := executor.DLQStatus(statusRequest.Status)
	validStatuses := []executor.DLQStatus{
		executor.DLQStatusActive,
		executor.DLQStatusRetrying,
		executor.DLQStatusResolved,
		executor.DLQStatusArchived,
	}

	valid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			valid = true
			break
		}
	}

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "Invalid status",
			"valid_statuses": validStatuses,
		})
		return
	}

	// Update status
	err := s.dlq.UpdateStatus(dlqID, status)
	if err != nil {
		LogError("dlq_api", "update_status", err, logrus.Fields{
			"dlq_id": dlqID,
			"status": status,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update DLQ status",
			"details": err.Error(),
		})
		return
	}

	LogSystemEvent("dlq_status_updated", logrus.Fields{
		"dlq_id": dlqID,
		"status": status,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "DLQ status updated successfully",
		"dlq_id":  dlqID,
		"status":  status,
	})
}

// handleArchiveDLQEntry archives a DLQ entry
func (s *DLQService) handleArchiveDLQEntry(c *gin.Context) {
	dlqID := c.Param("id")
	if dlqID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DLQ ID is required",
		})
		return
	}

	err := s.dlq.Archive(dlqID)
	if err != nil {
		LogError("dlq_api", "archive_entry", err, logrus.Fields{
			"dlq_id": dlqID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to archive DLQ entry",
			"details": err.Error(),
		})
		return
	}

	LogSystemEvent("dlq_archived", logrus.Fields{
		"dlq_id": dlqID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "DLQ entry archived successfully",
		"dlq_id":  dlqID,
	})
}

// handleDeleteDLQEntry deletes a DLQ entry
func (s *DLQService) handleDeleteDLQEntry(c *gin.Context) {
	dlqID := c.Param("id")
	if dlqID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "DLQ ID is required",
		})
		return
	}

	err := s.dlq.Delete(dlqID)
	if err != nil {
		LogError("dlq_api", "delete_entry", err, logrus.Fields{
			"dlq_id": dlqID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete DLQ entry",
			"details": err.Error(),
		})
		return
	}

	LogSystemEvent("dlq_deleted", logrus.Fields{
		"dlq_id": dlqID,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "DLQ entry deleted successfully",
		"dlq_id":  dlqID,
	})
}

// handleDLQStats returns statistics about the DLQ
func (s *DLQService) handleDLQStats(c *gin.Context) {
	// Get all entries
	allEntries, err := s.dlq.List(executor.DLQFilter{})
	if err != nil {
		LogError("dlq_api", "get_stats", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get DLQ statistics",
			"details": err.Error(),
		})
		return
	}

	// Calculate statistics
	stats := map[string]interface{}{
		"total_entries": len(allEntries),
		"by_status": map[executor.DLQStatus]int{
			executor.DLQStatusActive:   0,
			executor.DLQStatusRetrying: 0,
			executor.DLQStatusResolved: 0,
			executor.DLQStatusArchived: 0,
		},
		"by_workflow_type": make(map[string]int),
		"oldest_entry":     nil,
		"newest_entry":     nil,
	}

	statusCounts := stats["by_status"].(map[executor.DLQStatus]int)
	workflowTypeCounts := stats["by_workflow_type"].(map[string]int)

	for _, entry := range allEntries {
		// Count by status
		statusCounts[entry.Status]++

		// Count by workflow type
		workflowTypeCounts[entry.Metadata.WorkflowType]++

		// Track oldest and newest
		if stats["oldest_entry"] == nil {
			stats["oldest_entry"] = &entry.CreatedAt
		} else if entry.CreatedAt.Before(*stats["oldest_entry"].(*time.Time)) {
			stats["oldest_entry"] = &entry.CreatedAt
		}

		if stats["newest_entry"] == nil {
			stats["newest_entry"] = &entry.CreatedAt
		} else if entry.CreatedAt.After(*stats["newest_entry"].(*time.Time)) {
			stats["newest_entry"] = &entry.CreatedAt
		}
	}

	c.JSON(http.StatusOK, stats)
}
