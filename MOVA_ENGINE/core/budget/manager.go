package budget

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager manages budget constraints and quotas
type Manager struct {
	budgets     map[string]*Budget
	usage       map[string]*BudgetUsage
	alerts      map[string]*BudgetAlert
	violations  []BudgetViolation
	metrics     *ResourceMetrics
	
	mu          sync.RWMutex
	
	// Configuration
	checkInterval   time.Duration
	cleanupInterval time.Duration
	maxViolations   int
	
	// Channels
	violationChan chan BudgetViolation
	metricsChan   chan ResourceMetrics
	
	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewManager creates a new budget manager
func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Manager{
		budgets:         make(map[string]*Budget),
		usage:           make(map[string]*BudgetUsage),
		alerts:          make(map[string]*BudgetAlert),
		violations:      make([]BudgetViolation, 0),
		checkInterval:   time.Minute,
		cleanupInterval: time.Hour,
		maxViolations:   1000,
		violationChan:   make(chan BudgetViolation, 100),
		metricsChan:     make(chan ResourceMetrics, 10),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Start starts the budget manager background processes
func (m *Manager) Start() error {
	m.wg.Add(3)
	
	// Start budget checker
	go m.budgetChecker()
	
	// Start cleanup routine
	go m.cleanupRoutine()
	
	// Start violation processor
	go m.violationProcessor()
	
	return nil
}

// Stop stops the budget manager
func (m *Manager) Stop() error {
	m.cancel()
	m.wg.Wait()
	
	close(m.violationChan)
	close(m.metricsChan)
	
	return nil
}

// AddBudget adds a new budget constraint
func (m *Manager) AddBudget(budget *Budget) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if budget.ID == "" {
		budget.ID = uuid.New().String()
	}
	
	now := time.Now()
	if budget.CreatedAt.IsZero() {
		budget.CreatedAt = now
	}
	budget.UpdatedAt = now
	
	// Initialize usage tracking
	m.usage[budget.ID] = &BudgetUsage{
		BudgetID:        budget.ID,
		WindowStartTime: m.calculateWindowStart(budget.TimeWindow),
		LastUpdated:     now,
	}
	
	m.budgets[budget.ID] = budget
	return nil
}

// RemoveBudget removes a budget constraint
func (m *Manager) RemoveBudget(budgetID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	delete(m.budgets, budgetID)
	delete(m.usage, budgetID)
	
	// Remove related alerts
	for id, alert := range m.alerts {
		if alert.BudgetID == budgetID {
			delete(m.alerts, id)
		}
	}
	
	return nil
}

// UpdateBudget updates an existing budget
func (m *Manager) UpdateBudget(budget *Budget) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.budgets[budget.ID]; !exists {
		return fmt.Errorf("budget not found: %s", budget.ID)
	}
	
	budget.UpdatedAt = time.Now()
	m.budgets[budget.ID] = budget
	
	return nil
}

// GetBudget returns a budget by ID
func (m *Manager) GetBudget(budgetID string) (*Budget, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	budget, exists := m.budgets[budgetID]
	if !exists {
		return nil, fmt.Errorf("budget not found: %s", budgetID)
	}
	
	// Update current usage
	if usage, exists := m.usage[budgetID]; exists {
		budget.CurrentCount = usage.Count
		budget.CurrentDuration = usage.Duration
		budget.CurrentMemory = usage.Memory
		budget.CurrentCPU = usage.CPU
		budget.LastUsedAt = usage.LastUpdated
	}
	
	return budget, nil
}

// ListBudgets returns all budgets
func (m *Manager) ListBudgets() []*Budget {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	budgets := make([]*Budget, 0, len(m.budgets))
	for _, budget := range m.budgets {
		// Update current usage
		if usage, exists := m.usage[budget.ID]; exists {
			budget.CurrentCount = usage.Count
			budget.CurrentDuration = usage.Duration
			budget.CurrentMemory = usage.Memory
			budget.CurrentCPU = usage.CPU
			budget.LastUsedAt = usage.LastUpdated
		}
		budgets = append(budgets, budget)
	}
	
	return budgets
}

// CheckBudget checks if a budget request is allowed
func (m *Manager) CheckBudget(req *BudgetCheckRequest) (*BudgetCheckResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	response := &BudgetCheckResponse{
		Allowed:        true,
		Violations:     make([]BudgetViolation, 0),
		RemainingQuota: make(map[string]interface{}),
	}
	
	// Find applicable budgets
	applicableBudgets := m.findApplicableBudgets(req)
	
	for _, budget := range applicableBudgets {
		if !budget.Enabled {
			continue
		}
		
		usage := m.usage[budget.ID]
		if usage == nil {
			continue
		}
		
		// Check if window needs reset
		if m.shouldResetWindow(budget, usage) {
			m.resetBudgetUsage(budget.ID)
			usage = m.usage[budget.ID]
		}
		
		// Check constraints
		violation := m.checkBudgetConstraints(budget, usage, req)
		if violation != nil {
			response.Violations = append(response.Violations, *violation)
			if violation.Severity == "error" || violation.Severity == "critical" {
				response.Allowed = false
			}
		}
		
		// Calculate remaining quota
		m.addRemainingQuota(response, budget, usage)
	}
	
	if len(response.Violations) > 0 {
		response.Message = fmt.Sprintf("Found %d budget violations", len(response.Violations))
	}
	
	return response, nil
}

// RecordUsage records resource usage for budget tracking
func (m *Manager) RecordUsage(req *BudgetCheckRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Find applicable budgets
	applicableBudgets := m.findApplicableBudgets(req)
	
	for _, budget := range applicableBudgets {
		if !budget.Enabled {
			continue
		}
		
		usage := m.usage[budget.ID]
		if usage == nil {
			continue
		}
		
		// Check if window needs reset
		if m.shouldResetWindow(budget, usage) {
			m.resetBudgetUsage(budget.ID)
			usage = m.usage[budget.ID]
		}
		
		// Update usage
		usage.Count += req.Count
		usage.Duration += req.Duration
		if req.Memory > usage.Memory {
			usage.Memory = req.Memory // Peak memory usage
		}
		if req.CPU > usage.CPU {
			usage.CPU = req.CPU // Peak CPU usage
		}
		usage.LastUpdated = time.Now()
	}
	
	return nil
}

// findApplicableBudgets finds budgets that apply to the request
func (m *Manager) findApplicableBudgets(req *BudgetCheckRequest) []*Budget {
	applicable := make([]*Budget, 0)
	
	for _, budget := range m.budgets {
		if budget.Type != req.Type {
			continue
		}
		
		// Check scope matching
		switch budget.Scope {
		case BudgetScopeGlobal:
			applicable = append(applicable, budget)
		case BudgetScopeOrganization:
			if budget.ScopeID == req.OrganizationID {
				applicable = append(applicable, budget)
			}
		case BudgetScopeUser:
			if budget.ScopeID == req.UserID {
				applicable = append(applicable, budget)
			}
		case BudgetScopeWorkflow:
			if budget.ScopeID == req.WorkflowID {
				applicable = append(applicable, budget)
			}
		case BudgetScopeSession:
			if budget.ScopeID == req.SessionID {
				applicable = append(applicable, budget)
			}
		}
	}
	
	return applicable
}

// checkBudgetConstraints checks if usage violates budget constraints
func (m *Manager) checkBudgetConstraints(budget *Budget, usage *BudgetUsage, req *BudgetCheckRequest) *BudgetViolation {
	now := time.Now()
	
	// Check count limit
	if budget.MaxCount > 0 {
		projectedCount := usage.Count + req.Count
		if projectedCount > budget.MaxCount {
			return &BudgetViolation{
				ID:             uuid.New().String(),
				BudgetID:       budget.ID,
				BudgetName:     budget.Name,
				Type:           budget.Type,
				Scope:          budget.Scope,
				ScopeID:        budget.ScopeID,
				ViolationType:  "exceeded",
				Limit:          budget.MaxCount,
				CurrentValue:   projectedCount,
				PercentageUsed: float64(projectedCount) / float64(budget.MaxCount) * 100,
				WorkflowID:     req.WorkflowID,
				SessionID:      req.SessionID,
				UserID:         req.UserID,
				OrganizationID: req.OrganizationID,
				Timestamp:      now,
				Message:        fmt.Sprintf("Count limit exceeded: %d/%d", projectedCount, budget.MaxCount),
				Severity:       "error",
			}
		}
		
		// Warning at 80%
		if float64(projectedCount)/float64(budget.MaxCount) >= 0.8 {
			return &BudgetViolation{
				ID:             uuid.New().String(),
				BudgetID:       budget.ID,
				BudgetName:     budget.Name,
				Type:           budget.Type,
				Scope:          budget.Scope,
				ScopeID:        budget.ScopeID,
				ViolationType:  "approaching",
				Limit:          budget.MaxCount,
				CurrentValue:   projectedCount,
				PercentageUsed: float64(projectedCount) / float64(budget.MaxCount) * 100,
				WorkflowID:     req.WorkflowID,
				SessionID:      req.SessionID,
				UserID:         req.UserID,
				OrganizationID: req.OrganizationID,
				Timestamp:      now,
				Message:        fmt.Sprintf("Count approaching limit: %d/%d (%.1f%%)", projectedCount, budget.MaxCount, float64(projectedCount)/float64(budget.MaxCount)*100),
				Severity:       "warning",
			}
		}
	}
	
	// Similar checks for duration, memory, CPU...
	// (Implementation details omitted for brevity)
	
	return nil
}

// Helper methods
func (m *Manager) calculateWindowStart(window TimeWindow) time.Time {
	now := time.Now()
	switch window {
	case TimeWindowMinute:
		return now.Truncate(time.Minute)
	case TimeWindowHour:
		return now.Truncate(time.Hour)
	case TimeWindowDay:
		return now.Truncate(24 * time.Hour)
	case TimeWindowMonth:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	default:
		return now.Truncate(time.Hour)
	}
}

func (m *Manager) shouldResetWindow(budget *Budget, usage *BudgetUsage) bool {
	windowStart := m.calculateWindowStart(budget.TimeWindow)
	return usage.WindowStartTime.Before(windowStart)
}

func (m *Manager) resetBudgetUsage(budgetID string) {
	budget := m.budgets[budgetID]
	m.usage[budgetID] = &BudgetUsage{
		BudgetID:        budgetID,
		WindowStartTime: m.calculateWindowStart(budget.TimeWindow),
		LastUpdated:     time.Now(),
	}
}

func (m *Manager) addRemainingQuota(response *BudgetCheckResponse, budget *Budget, usage *BudgetUsage) {
	if budget.MaxCount > 0 {
		remaining := budget.MaxCount - usage.Count
		if remaining < 0 {
			remaining = 0
		}
		response.RemainingQuota[fmt.Sprintf("%s_count", budget.Type)] = remaining
	}
	
	// Calculate reset time
	if response.ResetTime.IsZero() || usage.WindowStartTime.Add(m.getWindowDuration(budget.TimeWindow)).Before(response.ResetTime) {
		response.ResetTime = usage.WindowStartTime.Add(m.getWindowDuration(budget.TimeWindow))
	}
}

func (m *Manager) getWindowDuration(window TimeWindow) time.Duration {
	switch window {
	case TimeWindowMinute:
		return time.Minute
	case TimeWindowHour:
		return time.Hour
	case TimeWindowDay:
		return 24 * time.Hour
	case TimeWindowMonth:
		return 30 * 24 * time.Hour // Approximate
	default:
		return time.Hour
	}
}

// Background processes
func (m *Manager) budgetChecker() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.performBudgetCheck()
		}
	}
}

func (m *Manager) cleanupRoutine() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.performCleanup()
		}
	}
}

func (m *Manager) violationProcessor() {
	defer m.wg.Done()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case violation := <-m.violationChan:
			m.processViolation(violation)
		}
	}
}

func (m *Manager) performBudgetCheck() {
	// Implementation for periodic budget checks
}

func (m *Manager) performCleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Clean up old violations
	if len(m.violations) > m.maxViolations {
		m.violations = m.violations[len(m.violations)-m.maxViolations:]
	}
	
	// Clean up expired usage records
	now := time.Now()
	for budgetID, usage := range m.usage {
		budget := m.budgets[budgetID]
		if budget != nil && m.shouldResetWindow(budget, usage) {
			if now.Sub(usage.LastUpdated) > m.getWindowDuration(budget.TimeWindow)*2 {
				m.resetBudgetUsage(budgetID)
			}
		}
	}
}

func (m *Manager) processViolation(violation BudgetViolation) {
	m.mu.Lock()
	m.violations = append(m.violations, violation)
	m.mu.Unlock()
	
	// Send alerts if configured
	m.sendAlerts(violation)
}

func (m *Manager) sendAlerts(violation BudgetViolation) {
	// Implementation for sending alerts
	// This would integrate with notification systems
}

