package budget

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ResourceMonitor monitors system resource usage
type ResourceMonitor struct {
	budgetManager *Manager
	
	// Current metrics
	currentMetrics *ResourceMetrics
	metricsHistory []ResourceMetrics
	
	// Configuration
	monitorInterval   time.Duration
	historySize       int
	cpuThreshold      float64 // 0-1
	memoryThreshold   float64 // 0-1
	
	// State
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor(budgetManager *Manager) *ResourceMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ResourceMonitor{
		budgetManager:   budgetManager,
		metricsHistory:  make([]ResourceMetrics, 0),
		monitorInterval: 10 * time.Second,
		historySize:     100,
		cpuThreshold:    0.8,    // 80%
		memoryThreshold: 0.85,   // 85%
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Start starts the resource monitor
func (rm *ResourceMonitor) Start() error {
	rm.wg.Add(1)
	go rm.monitorLoop()
	return nil
}

// Stop stops the resource monitor
func (rm *ResourceMonitor) Stop() error {
	rm.cancel()
	rm.wg.Wait()
	return nil
}

// GetCurrentMetrics returns current resource metrics
func (rm *ResourceMonitor) GetCurrentMetrics() *ResourceMetrics {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	if rm.currentMetrics == nil {
		return &ResourceMetrics{
			Timestamp: time.Now(),
		}
	}
	
	// Return a copy
	metrics := *rm.currentMetrics
	return &metrics
}

// GetMetricsHistory returns historical resource metrics
func (rm *ResourceMonitor) GetMetricsHistory() []ResourceMetrics {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	
	// Return a copy
	history := make([]ResourceMetrics, len(rm.metricsHistory))
	copy(history, rm.metricsHistory)
	return history
}

// CheckResourceLimits checks if current resource usage violates limits
func (rm *ResourceMonitor) CheckResourceLimits(req *ResourceCheckRequest) (*BudgetCheckResponse, error) {
	metrics := rm.GetCurrentMetrics()
	
	response := &BudgetCheckResponse{
		Allowed:        true,
		Violations:     make([]BudgetViolation, 0),
		RemainingQuota: make(map[string]interface{}),
	}
	
	// Check CPU limits
	if req.RequiredCPU > 0 {
		cpuResponse, err := rm.checkCPULimits(req, metrics)
		if err != nil {
			return nil, err
		}
		
		if !cpuResponse.Allowed {
			response.Allowed = false
		}
		response.Violations = append(response.Violations, cpuResponse.Violations...)
	}
	
	// Check memory limits
	if req.RequiredMemory > 0 {
		memResponse, err := rm.checkMemoryLimits(req, metrics)
		if err != nil {
			return nil, err
		}
		
		if !memResponse.Allowed {
			response.Allowed = false
		}
		response.Violations = append(response.Violations, memResponse.Violations...)
	}
	
	// Add remaining quota information
	rm.addResourceQuota(response, metrics)
	
	return response, nil
}

// checkCPULimits checks CPU usage limits
func (rm *ResourceMonitor) checkCPULimits(req *ResourceCheckRequest, metrics *ResourceMetrics) (*BudgetCheckResponse, error) {
	budgetReq := &BudgetCheckRequest{
		Type:           BudgetTypeCPU,
		Scope:          req.Scope,
		ScopeID:        req.ScopeID,
		CPU:            req.RequiredCPU,
		WorkflowID:     req.WorkflowID,
		SessionID:      req.SessionID,
		UserID:         req.UserID,
		OrganizationID: req.OrganizationID,
	}
	
	return rm.budgetManager.CheckBudget(budgetReq)
}

// checkMemoryLimits checks memory usage limits
func (rm *ResourceMonitor) checkMemoryLimits(req *ResourceCheckRequest, metrics *ResourceMetrics) (*BudgetCheckResponse, error) {
	budgetReq := &BudgetCheckRequest{
		Type:           BudgetTypeMemory,
		Scope:          req.Scope,
		ScopeID:        req.ScopeID,
		Memory:         req.RequiredMemory,
		WorkflowID:     req.WorkflowID,
		SessionID:      req.SessionID,
		UserID:         req.UserID,
		OrganizationID: req.OrganizationID,
	}
	
	return rm.budgetManager.CheckBudget(budgetReq)
}

// addResourceQuota adds remaining resource quota to response
func (rm *ResourceMonitor) addResourceQuota(response *BudgetCheckResponse, metrics *ResourceMetrics) {
	// Add available CPU
	if metrics.MemoryTotal > 0 {
		availableMemory := metrics.MemoryTotal - metrics.MemoryUsage
		response.RemainingQuota["available_memory"] = availableMemory
		response.RemainingQuota["memory_percentage"] = float64(metrics.MemoryUsage) / float64(metrics.MemoryTotal) * 100
	}
	
	response.RemainingQuota["cpu_usage"] = metrics.CPUUsage * 100
	response.RemainingQuota["active_workflows"] = metrics.ActiveWorkflows
	response.RemainingQuota["active_sessions"] = metrics.ActiveSessions
}

// monitorLoop runs the resource monitoring loop
func (rm *ResourceMonitor) monitorLoop() {
	defer rm.wg.Done()
	
	ticker := time.NewTicker(rm.monitorInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.collectMetrics()
		}
	}
}

// collectMetrics collects current resource metrics
func (rm *ResourceMonitor) collectMetrics() {
	metrics := rm.gatherSystemMetrics()
	
	rm.mu.Lock()
	defer rm.mu.Unlock()
	
	rm.currentMetrics = &metrics
	
	// Add to history
	rm.metricsHistory = append(rm.metricsHistory, metrics)
	
	// Trim history if too large
	if len(rm.metricsHistory) > rm.historySize {
		rm.metricsHistory = rm.metricsHistory[1:]
	}
	
	// Check for threshold violations
	rm.checkThresholds(metrics)
	
	// Send metrics to budget manager if configured
	select {
	case rm.budgetManager.metricsChan <- metrics:
	default:
		// Channel full, skip this update
	}
}

// gatherSystemMetrics gathers current system metrics
func (rm *ResourceMonitor) gatherSystemMetrics() ResourceMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	metrics := ResourceMetrics{
		Timestamp:   time.Now(),
		MemoryUsage: int64(m.Alloc),
		MemoryTotal: int64(m.Sys),
		// CPUUsage would need platform-specific implementation
		// For now, we'll use a placeholder
		CPUUsage: rm.estimateCPUUsage(),
		
		// These would come from the executor state
		ActiveWorkflows: rm.getActiveWorkflowCount(),
		ActiveSessions:  rm.getActiveSessionCount(),
		QueuedJobs:      rm.getQueuedJobCount(),
	}
	
	return metrics
}

// estimateCPUUsage estimates CPU usage (placeholder implementation)
func (rm *ResourceMonitor) estimateCPUUsage() float64 {
	// This is a simplified implementation
	// In production, this would use platform-specific APIs
	return float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()) / 100.0
}

// getActiveWorkflowCount returns the number of active workflows
func (rm *ResourceMonitor) getActiveWorkflowCount() int {
	// This would integrate with the executor to get actual counts
	return 0
}

// getActiveSessionCount returns the number of active sessions
func (rm *ResourceMonitor) getActiveSessionCount() int {
	// This would integrate with the session manager
	return 0
}

// getQueuedJobCount returns the number of queued jobs
func (rm *ResourceMonitor) getQueuedJobCount() int {
	// This would integrate with the job queue
	return 0
}

// checkThresholds checks if metrics exceed configured thresholds
func (rm *ResourceMonitor) checkThresholds(metrics ResourceMetrics) {
	now := time.Now()
	
	// Check CPU threshold
	if metrics.CPUUsage > rm.cpuThreshold {
		violation := BudgetViolation{
			ID:             fmt.Sprintf("cpu-threshold-%d", now.Unix()),
			BudgetID:       "system-cpu-threshold",
			BudgetName:     "System CPU Threshold",
			Type:           BudgetTypeCPU,
			Scope:          BudgetScopeGlobal,
			ViolationType:  "threshold_exceeded",
			Limit:          rm.cpuThreshold,
			CurrentValue:   metrics.CPUUsage,
			PercentageUsed: metrics.CPUUsage * 100,
			Timestamp:      now,
			Message:        fmt.Sprintf("CPU usage %.1f%% exceeds threshold %.1f%%", metrics.CPUUsage*100, rm.cpuThreshold*100),
			Severity:       "warning",
		}
		
		select {
		case rm.budgetManager.violationChan <- violation:
		default:
			// Channel full, skip violation
		}
	}
	
	// Check memory threshold
	if metrics.MemoryTotal > 0 {
		memoryPercentage := float64(metrics.MemoryUsage) / float64(metrics.MemoryTotal)
		if memoryPercentage > rm.memoryThreshold {
			violation := BudgetViolation{
				ID:             fmt.Sprintf("memory-threshold-%d", now.Unix()),
				BudgetID:       "system-memory-threshold",
				BudgetName:     "System Memory Threshold",
				Type:           BudgetTypeMemory,
				Scope:          BudgetScopeGlobal,
				ViolationType:  "threshold_exceeded",
				Limit:          rm.memoryThreshold,
				CurrentValue:   memoryPercentage,
				PercentageUsed: memoryPercentage * 100,
				Timestamp:      now,
				Message:        fmt.Sprintf("Memory usage %.1f%% exceeds threshold %.1f%%", memoryPercentage*100, rm.memoryThreshold*100),
				Severity:       "warning",
			}
			
			select {
			case rm.budgetManager.violationChan <- violation:
			default:
				// Channel full, skip violation
			}
		}
	}
}

// ResourceCheckRequest represents a request to check resource availability
type ResourceCheckRequest struct {
	Scope          BudgetScope `json:"scope"`
	ScopeID        string      `json:"scopeId"`
	RequiredCPU    float64     `json:"requiredCPU,omitempty"`
	RequiredMemory int64       `json:"requiredMemory,omitempty"`
	
	// Context
	WorkflowID     string `json:"workflowId,omitempty"`
	SessionID      string `json:"sessionId,omitempty"`
	UserID         string `json:"userId,omitempty"`
	OrganizationID string `json:"organizationId,omitempty"`
}

// ResourceLimiter provides resource limiting functionality
type ResourceLimiter struct {
	monitor *ResourceMonitor
}

// NewResourceLimiter creates a new resource limiter
func NewResourceLimiter(monitor *ResourceMonitor) *ResourceLimiter {
	return &ResourceLimiter{
		monitor: monitor,
	}
}

// CheckResourceAvailability checks if resources are available for a request
func (rl *ResourceLimiter) CheckResourceAvailability(req *ResourceCheckRequest) (*BudgetCheckResponse, error) {
	return rl.monitor.CheckResourceLimits(req)
}

// CreateDefaultResourceBudgets creates default resource budgets
func (rl *ResourceLimiter) CreateDefaultResourceBudgets(budgetManager *Manager) error {
	defaultBudgets := []*Budget{
		{
			ID:          "global-cpu-budget",
			Name:        "Global CPU Budget",
			Description: "Global CPU usage limit",
			Type:        BudgetTypeCPU,
			Scope:       BudgetScopeGlobal,
			MaxCPU:      0.8, // 80% CPU usage limit
			TimeWindow:  TimeWindowMinute,
			Enabled:     true,
		},
		{
			ID:          "global-memory-budget",
			Name:        "Global Memory Budget",
			Description: "Global memory usage limit",
			Type:        BudgetTypeMemory,
			Scope:       BudgetScopeGlobal,
			MaxMemory:   1024 * 1024 * 1024 * 2, // 2GB limit
			TimeWindow:  TimeWindowMinute,
			Enabled:     true,
		},
		{
			ID:          "workflow-memory-budget",
			Name:        "Per-Workflow Memory Budget",
			Description: "Memory limit per workflow execution",
			Type:        BudgetTypeMemory,
			Scope:       BudgetScopeWorkflow,
			MaxMemory:   1024 * 1024 * 100, // 100MB per workflow
			TimeWindow:  TimeWindowHour,
			Enabled:     true,
		},
		{
			ID:          "user-cpu-budget",
			Name:        "Per-User CPU Budget",
			Description: "CPU usage limit per user",
			Type:        BudgetTypeCPU,
			Scope:       BudgetScopeUser,
			MaxCPU:      0.2, // 20% CPU per user
			TimeWindow:  TimeWindowMinute,
			Enabled:     true,
		},
	}
	
	for _, budget := range defaultBudgets {
		if err := budgetManager.AddBudget(budget); err != nil {
			return fmt.Errorf("failed to create default resource budget %s: %w", budget.ID, err)
		}
	}
	
	return nil
}

