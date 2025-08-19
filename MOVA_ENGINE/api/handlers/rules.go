package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mova-engine/mova-engine/core/rules"
)

// RuleHandler handles rule-related HTTP requests
type RuleHandler struct {
	engine     rules.RuleEngine
	repository rules.RuleRepository
}

// NewRuleHandler creates a new rule handler
func NewRuleHandler(engine rules.RuleEngine, repository rules.RuleRepository) *RuleHandler {
	return &RuleHandler{
		engine:     engine,
		repository: repository,
	}
}

// ListRules returns all rules with optional filtering
// GET /v1/rules
func (h *RuleHandler) ListRules(c *gin.Context) {
	// Parse query parameters for filtering
	filter := make(map[string]interface{})

	if enabled := c.Query("enabled"); enabled != "" {
		if enabledBool, err := strconv.ParseBool(enabled); err == nil {
			filter["enabled"] = enabledBool
		}
	}

	if priority := c.Query("priority"); priority != "" {
		if priorityInt, err := strconv.Atoi(priority); err == nil {
			filter["priority"] = priorityInt
		}
	}

	if name := c.Query("name"); name != "" {
		filter["name"] = name
	}

	rules, err := h.repository.ListRules(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list rules",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"count": len(rules),
	})
}

// GetRule returns a specific rule by ID
// GET /v1/rules/:id
func (h *RuleHandler) GetRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Rule ID is required",
		})
		return
	}

	rule, err := h.repository.GetRule(ruleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Rule not found",
			"rule_id": ruleID,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rule": rule,
	})
}

// CreateRule creates a new rule
// POST /v1/rules
func (h *RuleHandler) CreateRule(c *gin.Context) {
	var rule rules.Rule

	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid rule format",
			"details": err.Error(),
		})
		return
	}

	// Generate ID if not provided
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	// Validate rule
	if err := h.engine.ValidateRule(rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Rule validation failed",
			"details": err.Error(),
		})
		return
	}

	// Save rule
	if err := h.repository.CreateRule(rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create rule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"rule":    rule,
		"message": "Rule created successfully",
	})
}

// UpdateRule updates an existing rule
// PUT /v1/rules/:id
func (h *RuleHandler) UpdateRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Rule ID is required",
		})
		return
	}

	var rule rules.Rule

	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid rule format",
			"details": err.Error(),
		})
		return
	}

	// Ensure ID matches
	rule.ID = ruleID
	rule.UpdatedAt = time.Now()

	// Validate rule
	if err := h.engine.ValidateRule(rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Rule validation failed",
			"details": err.Error(),
		})
		return
	}

	// Update rule
	if err := h.repository.UpdateRule(rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update rule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rule":    rule,
		"message": "Rule updated successfully",
	})
}

// DeleteRule deletes a rule by ID
// DELETE /v1/rules/:id
func (h *RuleHandler) DeleteRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Rule ID is required",
		})
		return
	}

	if err := h.repository.DeleteRule(ruleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete rule",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rule deleted successfully",
		"rule_id": ruleID,
	})
}

// EvaluateRule performs a dry-run evaluation of rules against provided context
// POST /v1/rules/evaluate
func (h *RuleHandler) EvaluateRule(c *gin.Context) {
	var request struct {
		Rules   []rules.Rule   `json:"rules"`
		RuleSet *rules.RuleSet `json:"ruleset,omitempty"`
		Context rules.Context  `json:"context"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	var results []rules.Result
	var err error

	if request.RuleSet != nil {
		// Evaluate entire ruleset
		results, err = h.engine.Run(*request.RuleSet, request.Context)
	} else if len(request.Rules) > 0 {
		// Evaluate individual rules
		for _, rule := range request.Rules {
			result, evalErr := h.engine.Evaluate(rule, request.Context)
			if evalErr != nil {
				result.Error = evalErr.Error()
			}
			results = append(results, result)
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Either 'rules' or 'ruleset' must be provided",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Evaluation failed",
			"details": err.Error(),
		})
		return
	}

	// Count matches
	matchedCount := 0
	for _, result := range results {
		if result.Matched {
			matchedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"results":         results,
		"total_rules":     len(results),
		"matched_rules":   matchedCount,
		"evaluation_time": time.Now(),
	})
}

// ListRuleSets returns all rulesets
// GET /v1/rulesets
func (h *RuleHandler) ListRuleSets(c *gin.Context) {
	rulesets, err := h.repository.ListRuleSets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list rulesets",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rulesets": rulesets,
		"count":    len(rulesets),
	})
}

// GetRuleSet returns a specific ruleset by name
// GET /v1/rulesets/:name
func (h *RuleHandler) GetRuleSet(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "RuleSet name is required",
		})
		return
	}

	ruleset, err := h.repository.GetRuleSet(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "RuleSet not found",
			"name":    name,
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ruleset": ruleset,
	})
}

// SaveRuleSet creates or updates a ruleset
// POST /v1/rulesets
func (h *RuleHandler) SaveRuleSet(c *gin.Context) {
	var ruleset rules.RuleSet

	if err := c.ShouldBindJSON(&ruleset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ruleset format",
			"details": err.Error(),
		})
		return
	}

	// Set timestamps
	now := time.Now()
	if ruleset.CreatedAt.IsZero() {
		ruleset.CreatedAt = now
	}
	ruleset.UpdatedAt = now

	// Validate ruleset
	if err := h.engine.ValidateRuleSet(ruleset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "RuleSet validation failed",
			"details": err.Error(),
		})
		return
	}

	// Save ruleset
	if err := h.repository.SaveRuleSet(ruleset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to save ruleset",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ruleset": ruleset,
		"message": "RuleSet saved successfully",
	})
}

// ValidateRuleSet validates a ruleset without saving
// POST /v1/rulesets/validate
func (h *RuleHandler) ValidateRuleSet(c *gin.Context) {
	var ruleset rules.RuleSet

	if err := c.ShouldBindJSON(&ruleset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid ruleset format",
			"details": err.Error(),
		})
		return
	}

	// Validate ruleset
	if err := h.engine.ValidateRuleSet(ruleset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid":   false,
			"error":   "RuleSet validation failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":       true,
		"message":     "RuleSet is valid",
		"rules_count": len(ruleset.Rules),
	})
}

// RegisterRuleRoutes registers all rule-related routes
func RegisterRuleRoutes(router *gin.RouterGroup, handler *RuleHandler) {
	// Rule routes
	rules := router.Group("/rules")
	{
		rules.GET("", handler.ListRules)
		rules.GET("/:id", handler.GetRule)
		rules.POST("", handler.CreateRule)
		rules.PUT("/:id", handler.UpdateRule)
		rules.DELETE("/:id", handler.DeleteRule)
		rules.POST("/evaluate", handler.EvaluateRule)
	}

	// RuleSet routes
	rulesets := router.Group("/rulesets")
	{
		rulesets.GET("", handler.ListRuleSets)
		rulesets.GET("/:name", handler.GetRuleSet)
		rulesets.POST("", handler.SaveRuleSet)
		rulesets.POST("/validate", handler.ValidateRuleSet)
	}
}
