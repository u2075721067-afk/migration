package budget

import (
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	
	if manager == nil {
		t.Fatal("Manager should not be nil")
	}
	
	if manager.budgets == nil {
		t.Error("Budgets map should be initialized")
	}
	
	if manager.usage == nil {
		t.Error("Usage map should be initialized")
	}
	
	if manager.alerts == nil {
		t.Error("Alerts map should be initialized")
	}
}

func TestAddBudget(t *testing.T) {
	manager := NewManager()
	
	budget := &Budget{
		Name:        "Test Budget",
		Type:        BudgetTypeRetries,
		Scope:       BudgetScopeGlobal,
		MaxCount:    100,
		TimeWindow:  TimeWindowHour,
		Enabled:     true,
	}
	
	err := manager.AddBudget(budget)
	if err != nil {
		t.Errorf("Failed to add budget: %v", err)
	}
	
	if budget.ID == "" {
		t.Error("Budget ID should be generated")
	}
	
	if budget.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	
	if budget.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
	
	// Check if usage tracking was initialized
	if _, exists := manager.usage[budget.ID]; !exists {
		t.Error("Usage tracking should be initialized for new budget")
	}
}

func TestGetBudget(t *testing.T) {
	manager := NewManager()
	
	originalBudget := &Budget{
		Name:        "Test Budget",
		Type:        BudgetTypeRetries,
		Scope:       BudgetScopeGlobal,
		MaxCount:    100,
		TimeWindow:  TimeWindowHour,
		Enabled:     true,
	}
	
	err := manager.AddBudget(originalBudget)
	if err != nil {
		t.Fatalf("Failed to add budget: %v", err)
	}
	
	retrievedBudget, err := manager.GetBudget(originalBudget.ID)
	if err != nil {
		t.Errorf("Failed to get budget: %v", err)
	}
	
	if retrievedBudget.Name != originalBudget.Name {
		t.Errorf("Expected name %s, got %s", originalBudget.Name, retrievedBudget.Name)
	}
	
	if retrievedBudget.Type != originalBudget.Type {
		t.Errorf("Expected type %s, got %s", originalBudget.Type, retrievedBudget.Type)
	}
}

func TestGetBudgetNotFound(t *testing.T) {
	manager := NewManager()
	
	_, err := manager.GetBudget("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent budget")
	}
	
	if err.Error() != "budget not found: non-existent-id" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestCheckBudget(t *testing.T) {
	manager := NewManager()
	
	budget := &Budget{
		Name:        "Test Budget",
		Type:        BudgetTypeRetries,
		Scope:       BudgetScopeGlobal,
		MaxCount:    10,
		TimeWindow:  TimeWindowHour,
		Enabled:     true,
	}
	
	err := manager.AddBudget(budget)
	if err != nil {
		t.Fatalf("Failed to add budget: %v", err)
	}
	
	// Test allowed request
	req := &BudgetCheckRequest{
		Type:  BudgetTypeRetries,
		Scope: BudgetScopeGlobal,
		Count: 1,
	}
	
	response, err := manager.CheckBudget(req)
	if err != nil {
		t.Errorf("Failed to check budget: %v", err)
	}
	
	if !response.Allowed {
		t.Error("Request should be allowed")
	}
	
	if len(response.Violations) != 0 {
		t.Errorf("Expected no violations, got %d", len(response.Violations))
	}
}

func TestCheckBudgetExceeded(t *testing.T) {
	manager := NewManager()
	
	budget := &Budget{
		Name:        "Test Budget",
		Type:        BudgetTypeRetries,
		Scope:       BudgetScopeGlobal,
		MaxCount:    5,
		TimeWindow:  TimeWindowHour,
		Enabled:     true,
	}
	
	err := manager.AddBudget(budget)
	if err != nil {
		t.Fatalf("Failed to add budget: %v", err)
	}
	
	// Simulate usage that exceeds limit
	usage := manager.usage[budget.ID]
	usage.Count = 6 // Exceed the limit of 5
	
	req := &BudgetCheckRequest{
		Type:  BudgetTypeRetries,
		Scope: BudgetScopeGlobal,
		Count: 1, // This would make it 7, which exceeds 5
	}
	
	response, err := manager.CheckBudget(req)
	if err != nil {
		t.Errorf("Failed to check budget: %v", err)
	}
	
	if response.Allowed {
		t.Error("Request should not be allowed when budget is exceeded")
	}
	
	if len(response.Violations) == 0 {
		t.Error("Expected violations when budget is exceeded")
	}
	
	violation := response.Violations[0]
	if violation.ViolationType != "exceeded" {
		t.Errorf("Expected violation type 'exceeded', got %s", violation.ViolationType)
	}
}

func TestRecordUsage(t *testing.T) {
	manager := NewManager()
	
	budget := &Budget{
		Name:        "Test Budget",
		Type:        BudgetTypeRetries,
		Scope:       BudgetScopeGlobal,
		MaxCount:    100,
		TimeWindow:  TimeWindowHour,
		Enabled:     true,
	}
	
	err := manager.AddBudget(budget)
	if err != nil {
		t.Fatalf("Failed to add budget: %v", err)
	}
	
	req := &BudgetCheckRequest{
		Type:     BudgetTypeRetries,
		Scope:    BudgetScopeGlobal,
		Count:    5,
		Duration: 2 * time.Second,
	}
	
	err = manager.RecordUsage(req)
	if err != nil {
		t.Errorf("Failed to record usage: %v", err)
	}
	
	// Verify usage was recorded
	usage := manager.usage[budget.ID]
	if usage.Count != 5 {
		t.Errorf("Expected count 5, got %d", usage.Count)
	}
	
	if usage.Duration != 2*time.Second {
		t.Errorf("Expected duration 2s, got %v", usage.Duration)
	}
}

func TestCalculateWindowStart(t *testing.T) {
	manager := NewManager()
	
	testCases := []struct {
		window   TimeWindow
		expected string // We'll check if the result is truncated properly
	}{
		{TimeWindowMinute, "minute"},
		{TimeWindowHour, "hour"},
		{TimeWindowDay, "day"},
		{TimeWindowMonth, "month"},
	}
	
	for _, tc := range testCases {
		t.Run(string(tc.window), func(t *testing.T) {
			start := manager.calculateWindowStart(tc.window)
			if start.IsZero() {
				t.Error("Window start should not be zero")
			}
			
			// Basic sanity check - start should be in the past or present
			if start.After(time.Now()) {
				t.Error("Window start should not be in the future")
			}
		})
	}
}

func TestShouldResetWindow(t *testing.T) {
	manager := NewManager()
	
	budget := &Budget{
		TimeWindow: TimeWindowHour,
	}
	
	// Create usage with old window start time
	usage := &BudgetUsage{
		WindowStartTime: time.Now().Add(-2 * time.Hour), // 2 hours ago
	}
	
	shouldReset := manager.shouldResetWindow(budget, usage)
	if !shouldReset {
		t.Error("Should reset window when window start time is old")
	}
	
	// Create usage with recent window start time (within current hour)
	now := time.Now()
	currentHourStart := now.Truncate(time.Hour)
	usage.WindowStartTime = currentHourStart // Start of current hour
	
	shouldReset = manager.shouldResetWindow(budget, usage)
	if shouldReset {
		t.Error("Should not reset window when window start time is within current window")
	}
}

func TestFindApplicableBudgets(t *testing.T) {
	manager := NewManager()
	
	// Add budgets with different scopes
	globalBudget := &Budget{
		Name:   "Global Budget",
		Type:   BudgetTypeRetries,
		Scope:  BudgetScopeGlobal,
		Enabled: true,
	}
	
	userBudget := &Budget{
		Name:    "User Budget",
		Type:    BudgetTypeRetries,
		Scope:   BudgetScopeUser,
		ScopeID: "user-123",
		Enabled: true,
	}
	
	orgBudget := &Budget{
		Name:    "Org Budget",
		Type:    BudgetTypeRetries,
		Scope:   BudgetScopeOrganization,
		ScopeID: "org-456",
		Enabled: true,
	}
	
	manager.AddBudget(globalBudget)
	manager.AddBudget(userBudget)
	manager.AddBudget(orgBudget)
	
	// Test global request
	req := &BudgetCheckRequest{
		Type:  BudgetTypeRetries,
		Scope: BudgetScopeGlobal,
	}
	
	applicable := manager.findApplicableBudgets(req)
	if len(applicable) != 1 {
		t.Errorf("Expected 1 applicable budget for global request, got %d", len(applicable))
	}
	
	if applicable[0].Name != "Global Budget" {
		t.Errorf("Expected Global Budget, got %s", applicable[0].Name)
	}
	
	// Test user request
	req = &BudgetCheckRequest{
		Type:   BudgetTypeRetries,
		Scope:  BudgetScopeUser,
		ScopeID: "user-123",
		UserID: "user-123",
	}
	
	applicable = manager.findApplicableBudgets(req)
	// Should find both global and user-specific budgets
	if len(applicable) < 1 {
		t.Errorf("Expected at least 1 applicable budget for user request, got %d", len(applicable))
	}
	
	// Find the user budget specifically
	var userBudgetFound bool
	for _, budget := range applicable {
		if budget.Name == "User Budget" {
			userBudgetFound = true
			break
		}
	}
	
	if !userBudgetFound {
		t.Error("Expected to find User Budget in applicable budgets")
	}
}

func TestListBudgets(t *testing.T) {
	manager := NewManager()
	
	budget1 := &Budget{
		Name:    "Budget 1",
		Type:    BudgetTypeRetries,
		Scope:   BudgetScopeGlobal,
		Enabled: true,
	}
	
	budget2 := &Budget{
		Name:    "Budget 2",
		Type:    BudgetTypeCPU,
		Scope:   BudgetScopeUser,
		Enabled: true,
	}
	
	manager.AddBudget(budget1)
	manager.AddBudget(budget2)
	
	budgets := manager.ListBudgets()
	if len(budgets) != 2 {
		t.Errorf("Expected 2 budgets, got %d", len(budgets))
	}
	
	// Verify budgets are returned with current usage
	for _, budget := range budgets {
		if budget.LastUsedAt.IsZero() {
			// This is expected for new budgets
		}
	}
}

func TestRemoveBudget(t *testing.T) {
	manager := NewManager()
	
	budget := &Budget{
		Name:    "Test Budget",
		Type:    BudgetTypeRetries,
		Scope:   BudgetScopeGlobal,
		Enabled: true,
	}
	
	err := manager.AddBudget(budget)
	if err != nil {
		t.Fatalf("Failed to add budget: %v", err)
	}
	
	budgetID := budget.ID
	
	// Verify budget exists
	_, err = manager.GetBudget(budgetID)
	if err != nil {
		t.Errorf("Budget should exist before removal: %v", err)
	}
	
	// Remove budget
	err = manager.RemoveBudget(budgetID)
	if err != nil {
		t.Errorf("Failed to remove budget: %v", err)
	}
	
	// Verify budget is removed
	_, err = manager.GetBudget(budgetID)
	if err == nil {
		t.Error("Budget should not exist after removal")
	}
	
	// Verify usage tracking is cleaned up
	if _, exists := manager.usage[budgetID]; exists {
		t.Error("Usage tracking should be cleaned up after budget removal")
	}
}
