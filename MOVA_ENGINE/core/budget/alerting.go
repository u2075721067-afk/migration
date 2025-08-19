package budget

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// AlertManager manages budget alerts and notifications
type AlertManager struct {
	// Configuration
	channels      map[string]AlertChannel
	templates     map[string]AlertTemplate
	rules         map[string]AlertRule
	
	// State
	mu            sync.RWMutex
	activeAlerts  map[string]*ActiveAlert
	alertHistory  []AlertEvent
	
	// Settings
	maxHistorySize int
	cooldownPeriod time.Duration
	
	// Context
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &AlertManager{
		channels:       make(map[string]AlertChannel),
		templates:      make(map[string]AlertTemplate),
		rules:          make(map[string]AlertRule),
		activeAlerts:   make(map[string]*ActiveAlert),
		alertHistory:   make([]AlertEvent, 0),
		maxHistorySize: 1000,
		cooldownPeriod: 5 * time.Minute, // Default 5 minute cooldown
		ctx:            ctx,
		cancel:         cancel,
	}
}

// AlertChannel represents a notification channel
type AlertChannel interface {
	Send(ctx context.Context, alert *Alert) error
	GetType() string
	GetID() string
}

// AlertTemplate represents an alert message template
type AlertTemplate struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Subject  string            `json:"subject"`
	Body     string            `json:"body"`
	Format   string            `json:"format"` // "text", "html", "json"
	Variables map[string]string `json:"variables"`
}

// AlertRule represents conditions for triggering alerts
type AlertRule struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	
	// Trigger conditions
	BudgetTypes  []BudgetType  `json:"budgetTypes"`
	Scopes       []BudgetScope `json:"scopes"`
	Thresholds   []float64     `json:"thresholds"` // Percentage thresholds (0-1)
	
	// Alert configuration
	Channels     []string      `json:"channels"`    // Channel IDs
	Template     string        `json:"template"`    // Template ID
	Severity     string        `json:"severity"`    // "low", "medium", "high", "critical"
	Cooldown     time.Duration `json:"cooldown"`    // Cooldown period between alerts
	
	// Conditions
	Conditions   []AlertCondition `json:"conditions"`
	
	Enabled      bool          `json:"enabled"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}

// AlertCondition represents a condition for alert triggering
type AlertCondition struct {
	Type     string      `json:"type"`     // "time", "count", "percentage"
	Operator string      `json:"operator"` // "gt", "lt", "eq", "gte", "lte"
	Value    interface{} `json:"value"`
	
	// Time-based conditions
	TimeWindow time.Duration `json:"timeWindow,omitempty"`
	DaysOfWeek []string      `json:"daysOfWeek,omitempty"`
	Hours      []int         `json:"hours,omitempty"`
}

// Alert represents an alert to be sent
type Alert struct {
	ID             string                 `json:"id"`
	RuleID         string                 `json:"ruleId"`
	RuleName       string                 `json:"ruleName"`
	Severity       string                 `json:"severity"`
	
	// Alert content
	Subject        string                 `json:"subject"`
	Message        string                 `json:"message"`
	
	// Context
	BudgetID       string                 `json:"budgetId"`
	BudgetName     string                 `json:"budgetName"`
	BudgetType     BudgetType             `json:"budgetType"`
	Scope          BudgetScope            `json:"scope"`
	ScopeID        string                 `json:"scopeId"`
	
	// Violation details
	Violation      *BudgetViolation       `json:"violation,omitempty"`
	CurrentValue   interface{}            `json:"currentValue"`
	Threshold      interface{}            `json:"threshold"`
	PercentageUsed float64                `json:"percentageUsed"`
	
	// Metadata
	Timestamp      time.Time              `json:"timestamp"`
	Tags           map[string]string      `json:"tags"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// ActiveAlert represents an active alert state
type ActiveAlert struct {
	Alert     *Alert    `json:"alert"`
	LastSent  time.Time `json:"lastSent"`
	Count     int       `json:"count"`
	Channels  []string  `json:"channels"`
}

// AlertEvent represents a historical alert event
type AlertEvent struct {
	AlertID   string    `json:"alertId"`
	RuleID    string    `json:"ruleId"`
	Action    string    `json:"action"` // "triggered", "sent", "resolved"
	Channel   string    `json:"channel,omitempty"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Start starts the alert manager
func (am *AlertManager) Start() error {
	am.wg.Add(1)
	go am.alertProcessor()
	return nil
}

// Stop stops the alert manager
func (am *AlertManager) Stop() error {
	am.cancel()
	am.wg.Wait()
	return nil
}

// ProcessViolation processes a budget violation and triggers alerts if necessary
func (am *AlertManager) ProcessViolation(violation BudgetViolation) error {
	am.mu.RLock()
	applicableRules := am.findApplicableRules(violation)
	am.mu.RUnlock()
	
	for _, rule := range applicableRules {
		if !rule.Enabled {
			continue
		}
		
		// Check if alert should be triggered
		if am.shouldTriggerAlert(rule, violation) {
			alert := am.createAlert(rule, violation)
			if err := am.triggerAlert(alert); err != nil {
				log.Printf("Failed to trigger alert %s: %v", alert.ID, err)
			}
		}
	}
	
	return nil
}

// findApplicableRules finds alert rules that apply to a violation
func (am *AlertManager) findApplicableRules(violation BudgetViolation) []AlertRule {
	applicable := make([]AlertRule, 0)
	
	for _, rule := range am.rules {
		// Check budget type
		if len(rule.BudgetTypes) > 0 {
			found := false
			for _, budgetType := range rule.BudgetTypes {
				if budgetType == violation.Type {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		// Check scope
		if len(rule.Scopes) > 0 {
			found := false
			for _, scope := range rule.Scopes {
				if scope == violation.Scope {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		// Check thresholds
		if len(rule.Thresholds) > 0 {
			found := false
			for _, threshold := range rule.Thresholds {
				if violation.PercentageUsed >= threshold*100 {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		// Check conditions
		if am.evaluateAlertConditions(rule, violation) {
			applicable = append(applicable, rule)
		}
	}
	
	return applicable
}

// shouldTriggerAlert checks if an alert should be triggered based on cooldown and other factors
func (am *AlertManager) shouldTriggerAlert(rule AlertRule, violation BudgetViolation) bool {
	alertKey := fmt.Sprintf("%s:%s:%s", rule.ID, violation.BudgetID, violation.ScopeID)
	
	am.mu.RLock()
	activeAlert, exists := am.activeAlerts[alertKey]
	am.mu.RUnlock()
	
	if exists {
		// Check cooldown period
		cooldown := rule.Cooldown
		if cooldown == 0 {
			cooldown = am.cooldownPeriod
		}
		
		if time.Since(activeAlert.LastSent) < cooldown {
			return false
		}
	}
	
	return true
}

// createAlert creates an alert from a rule and violation
func (am *AlertManager) createAlert(rule AlertRule, violation BudgetViolation) *Alert {
	alert := &Alert{
		ID:             fmt.Sprintf("alert-%d", time.Now().UnixNano()),
		RuleID:         rule.ID,
		RuleName:       rule.Name,
		Severity:       rule.Severity,
		BudgetID:       violation.BudgetID,
		BudgetName:     violation.BudgetName,
		BudgetType:     violation.Type,
		Scope:          violation.Scope,
		ScopeID:        violation.ScopeID,
		Violation:      &violation,
		CurrentValue:   violation.CurrentValue,
		PercentageUsed: violation.PercentageUsed,
		Timestamp:      time.Now(),
		Tags:           make(map[string]string),
		Metadata:       make(map[string]interface{}),
	}
	
	// Apply template
	if template, exists := am.templates[rule.Template]; exists {
		alert.Subject = am.applyTemplate(template.Subject, alert)
		alert.Message = am.applyTemplate(template.Body, alert)
	} else {
		// Default message
		alert.Subject = fmt.Sprintf("Budget Alert: %s", violation.BudgetName)
		alert.Message = violation.Message
	}
	
	// Add metadata
	alert.Tags["severity"] = rule.Severity
	alert.Tags["budget_type"] = string(violation.Type)
	alert.Tags["scope"] = string(violation.Scope)
	alert.Metadata["rule_id"] = rule.ID
	alert.Metadata["violation_type"] = violation.ViolationType
	
	return alert
}

// triggerAlert triggers an alert by sending it through configured channels
func (am *AlertManager) triggerAlert(alert *Alert) error {
	rule := am.rules[alert.RuleID]
	
	// Record alert event
	am.recordAlertEvent(AlertEvent{
		AlertID:   alert.ID,
		RuleID:    alert.RuleID,
		Action:    "triggered",
		Timestamp: time.Now(),
		Success:   true,
	})
	
	// Send through each channel
	for _, channelID := range rule.Channels {
		if channel, exists := am.channels[channelID]; exists {
			if err := channel.Send(am.ctx, alert); err != nil {
				am.recordAlertEvent(AlertEvent{
					AlertID:   alert.ID,
					RuleID:    alert.RuleID,
					Action:    "sent",
					Channel:   channelID,
					Success:   false,
					Error:     err.Error(),
					Timestamp: time.Now(),
				})
				log.Printf("Failed to send alert %s via channel %s: %v", alert.ID, channelID, err)
			} else {
				am.recordAlertEvent(AlertEvent{
					AlertID:   alert.ID,
					RuleID:    alert.RuleID,
					Action:    "sent",
					Channel:   channelID,
					Success:   true,
					Timestamp: time.Now(),
				})
			}
		}
	}
	
	// Update active alert state
	am.updateActiveAlert(alert, rule.Channels)
	
	return nil
}

// updateActiveAlert updates the active alert state
func (am *AlertManager) updateActiveAlert(alert *Alert, channels []string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	alertKey := fmt.Sprintf("%s:%s:%s", alert.RuleID, alert.BudgetID, alert.ScopeID)
	
	if activeAlert, exists := am.activeAlerts[alertKey]; exists {
		activeAlert.LastSent = time.Now()
		activeAlert.Count++
	} else {
		am.activeAlerts[alertKey] = &ActiveAlert{
			Alert:    alert,
			LastSent: time.Now(),
			Count:    1,
			Channels: channels,
		}
	}
}

// recordAlertEvent records an alert event in history
func (am *AlertManager) recordAlertEvent(event AlertEvent) {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	am.alertHistory = append(am.alertHistory, event)
	
	// Trim history if too large
	if len(am.alertHistory) > am.maxHistorySize {
		am.alertHistory = am.alertHistory[1:]
	}
}

// applyTemplate applies a template to generate alert content
func (am *AlertManager) applyTemplate(template string, alert *Alert) string {
	// Simple template substitution
	// In production, this would use a proper template engine
	result := template
	
	// Replace common variables
	replacements := map[string]string{
		"{{.BudgetName}}":     alert.BudgetName,
		"{{.BudgetType}}":     string(alert.BudgetType),
		"{{.Scope}}":          string(alert.Scope),
		"{{.ScopeID}}":        alert.ScopeID,
		"{{.PercentageUsed}}": fmt.Sprintf("%.1f%%", alert.PercentageUsed),
		"{{.Severity}}":       alert.Severity,
		"{{.Timestamp}}":      alert.Timestamp.Format(time.RFC3339),
	}
	
	for _, _ = range replacements {
		// Placeholder for actual template replacement
		// In production, this would use a proper template engine like text/template
	}
	
	return result
}

// evaluateAlertConditions evaluates alert rule conditions
func (am *AlertManager) evaluateAlertConditions(rule AlertRule, violation BudgetViolation) bool {
	for _, condition := range rule.Conditions {
		if !am.evaluateAlertCondition(condition, violation) {
			return false
		}
	}
	return true
}

// evaluateAlertCondition evaluates a single alert condition
func (am *AlertManager) evaluateAlertCondition(condition AlertCondition, violation BudgetViolation) bool {
	switch condition.Type {
	case "percentage":
		threshold, ok := condition.Value.(float64)
		if !ok {
			return true
		}
		
		switch condition.Operator {
		case "gt":
			return violation.PercentageUsed > threshold
		case "gte":
			return violation.PercentageUsed >= threshold
		case "lt":
			return violation.PercentageUsed < threshold
		case "lte":
			return violation.PercentageUsed <= threshold
		case "eq":
			return violation.PercentageUsed == threshold
		}
		
	case "time":
		// Time-based conditions would be evaluated here
		return true
	}
	
	return true
}

// alertProcessor processes alerts in the background
func (am *AlertManager) alertProcessor() {
	defer am.wg.Done()
	
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-am.ctx.Done():
			return
		case <-ticker.C:
			am.cleanupActiveAlerts()
		}
	}
}

// cleanupActiveAlerts removes old active alerts
func (am *AlertManager) cleanupActiveAlerts() {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	cutoff := time.Now().Add(-24 * time.Hour) // Remove alerts older than 24 hours
	
	for key, activeAlert := range am.activeAlerts {
		if activeAlert.LastSent.Before(cutoff) {
			delete(am.activeAlerts, key)
		}
	}
}

// LogChannel implements AlertChannel for logging
type LogChannel struct {
	ID   string
	Name string
}

func NewLogChannel(id, name string) *LogChannel {
	return &LogChannel{
		ID:   id,
		Name: name,
	}
}

func (lc *LogChannel) Send(ctx context.Context, alert *Alert) error {
	alertJSON, _ := json.MarshalIndent(alert, "", "  ")
	log.Printf("BUDGET_ALERT: %s", string(alertJSON))
	return nil
}

func (lc *LogChannel) GetType() string {
	return "log"
}

func (lc *LogChannel) GetID() string {
	return lc.ID
}

// Management methods
func (am *AlertManager) AddChannel(channel AlertChannel) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.channels[channel.GetID()] = channel
}

func (am *AlertManager) AddTemplate(template AlertTemplate) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.templates[template.ID] = template
}

func (am *AlertManager) AddRule(rule AlertRule) {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	now := time.Now()
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	rule.UpdatedAt = now
	
	am.rules[rule.ID] = rule
}

func (am *AlertManager) GetAlertHistory() []AlertEvent {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	// Return a copy
	history := make([]AlertEvent, len(am.alertHistory))
	copy(history, am.alertHistory)
	return history
}

func (am *AlertManager) GetActiveAlerts() map[string]*ActiveAlert {
	am.mu.RLock()
	defer am.mu.RUnlock()
	
	// Return a copy
	active := make(map[string]*ActiveAlert)
	for key, alert := range am.activeAlerts {
		alertCopy := *alert
		active[key] = &alertCopy
	}
	return active
}
