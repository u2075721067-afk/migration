package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"../../core/budget"
)

// RateLimiterMiddleware provides API rate limiting functionality
type RateLimiterMiddleware struct {
	budgetManager *budget.Manager
	
	// Configuration
	defaultRateLimit int64         // requests per minute
	burstLimit       int64         // burst allowance
	headerPrefix     string        // header prefix for rate limit info
	skipPaths        []string      // paths to skip rate limiting
	
	// Rate limit configurations by endpoint
	endpointLimits map[string]int64
}

// NewRateLimiterMiddleware creates a new rate limiter middleware
func NewRateLimiterMiddleware(budgetManager *budget.Manager) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		budgetManager:    budgetManager,
		defaultRateLimit: 1000, // 1000 requests per minute by default
		burstLimit:       100,  // Allow burst of 100 requests
		headerPrefix:     "X-RateLimit",
		skipPaths: []string{
			"/health",
			"/metrics",
		},
		endpointLimits: map[string]int64{
			"POST /v1/execute":   500,  // 500 executions per minute
			"POST /v1/validate":  2000, // 2000 validations per minute
			"GET /v1/runs":       1000, // 1000 run queries per minute
			"POST /v1/policies":  100,  // 100 policy operations per minute
		},
	}
}

// RateLimitMiddleware returns the Gin middleware function
func (rlm *RateLimiterMiddleware) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for certain paths
		if rlm.shouldSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}
		
		// Extract client identification
		clientID := rlm.extractClientID(c)
		organizationID := rlm.extractOrganizationID(c)
		userID := rlm.extractUserID(c)
		
		// Get rate limit for this endpoint
		endpoint := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		rateLimit := rlm.getRateLimitForEndpoint(endpoint)
		
		// Check rate limits at different scopes
		if err := rlm.checkRateLimit(c, clientID, organizationID, userID, rateLimit); err != nil {
			rlm.handleRateLimitExceeded(c, err)
			return
		}
		
		// Record API usage
		if err := rlm.recordAPIUsage(c, clientID, organizationID, userID); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: failed to record API usage: %v\n", err)
		}
		
		c.Next()
	}
}

// shouldSkipPath checks if rate limiting should be skipped for a path
func (rlm *RateLimiterMiddleware) shouldSkipPath(path string) bool {
	for _, skipPath := range rlm.skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// extractClientID extracts client ID from request
func (rlm *RateLimiterMiddleware) extractClientID(c *gin.Context) string {
	// Try API key first
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return fmt.Sprintf("api_key:%s", apiKey)
	}
	
	// Try Authorization header
	if auth := c.GetHeader("Authorization"); auth != "" {
		return fmt.Sprintf("auth:%s", auth)
	}
	
	// Fall back to IP address
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// extractOrganizationID extracts organization ID from request
func (rlm *RateLimiterMiddleware) extractOrganizationID(c *gin.Context) string {
	// Check header
	if orgID := c.GetHeader("X-Organization-ID"); orgID != "" {
		return orgID
	}
	
	// Check query parameter
	if orgID := c.Query("org_id"); orgID != "" {
		return orgID
	}
	
	return ""
}

// extractUserID extracts user ID from request
func (rlm *RateLimiterMiddleware) extractUserID(c *gin.Context) string {
	// Check header
	if userID := c.GetHeader("X-User-ID"); userID != "" {
		return userID
	}
	
	// Check query parameter
	if userID := c.Query("user_id"); userID != "" {
		return userID
	}
	
	return ""
}

// getRateLimitForEndpoint gets the rate limit for a specific endpoint
func (rlm *RateLimiterMiddleware) getRateLimitForEndpoint(endpoint string) int64 {
	// Check exact match first
	if limit, exists := rlm.endpointLimits[endpoint]; exists {
		return limit
	}
	
	// Check pattern matches
	for pattern, limit := range rlm.endpointLimits {
		if rlm.matchEndpoint(pattern, endpoint) {
			return limit
		}
	}
	
	return rlm.defaultRateLimit
}

// matchEndpoint checks if an endpoint matches a pattern
func (rlm *RateLimiterMiddleware) matchEndpoint(pattern, endpoint string) bool {
	// Simple pattern matching - could be enhanced with regex
	return strings.Contains(endpoint, strings.TrimPrefix(pattern, "* "))
}

// checkRateLimit checks rate limits at different scopes
func (rlm *RateLimiterMiddleware) checkRateLimit(c *gin.Context, clientID, organizationID, userID string, rateLimit int64) error {
	// Check global rate limit
	if err := rlm.checkScopedRateLimit(budget.BudgetScopeGlobal, "", rateLimit, c); err != nil {
		return err
	}
	
	// Check organization rate limit
	if organizationID != "" {
		orgLimit := rateLimit * 10 // Organizations get 10x the individual limit
		if err := rlm.checkScopedRateLimit(budget.BudgetScopeOrganization, organizationID, orgLimit, c); err != nil {
			return err
		}
	}
	
	// Check user rate limit
	if userID != "" {
		if err := rlm.checkScopedRateLimit(budget.BudgetScopeUser, userID, rateLimit, c); err != nil {
			return err
		}
	}
	
	// Check client rate limit (IP/API key based)
	if err := rlm.checkScopedRateLimit(budget.BudgetScopeUser, clientID, rateLimit, c); err != nil {
		return err
	}
	
	return nil
}

// checkScopedRateLimit checks rate limit for a specific scope
func (rlm *RateLimiterMiddleware) checkScopedRateLimit(scope budget.BudgetScope, scopeID string, rateLimit int64, c *gin.Context) error {
	req := &budget.BudgetCheckRequest{
		Type:    budget.BudgetTypeAPIRequests,
		Scope:   scope,
		ScopeID: scopeID,
		Count:   1, // One API request
	}
	
	response, err := rlm.budgetManager.CheckBudget(req)
	if err != nil {
		return fmt.Errorf("rate limit check failed: %w", err)
	}
	
	if !response.Allowed {
		return &RateLimitError{
			Scope:      scope,
			ScopeID:    scopeID,
			Limit:      rateLimit,
			ResetTime:  response.ResetTime,
			Violations: response.Violations,
		}
	}
	
	// Add rate limit headers
	rlm.addRateLimitHeaders(c, scope, rateLimit, response)
	
	return nil
}

// recordAPIUsage records API usage for budget tracking
func (rlm *RateLimiterMiddleware) recordAPIUsage(c *gin.Context, clientID, organizationID, userID string) error {
	req := &budget.BudgetCheckRequest{
		Type:           budget.BudgetTypeAPIRequests,
		Count:          1,
		OrganizationID: organizationID,
		UserID:         userID,
	}
	
	// Record usage for all applicable scopes
	scopes := []struct {
		scope   budget.BudgetScope
		scopeID string
	}{
		{budget.BudgetScopeGlobal, ""},
		{budget.BudgetScopeOrganization, organizationID},
		{budget.BudgetScopeUser, userID},
		{budget.BudgetScopeUser, clientID}, // Client-based tracking
	}
	
	for _, scope := range scopes {
		if scope.scope == budget.BudgetScopeGlobal || scope.scopeID != "" {
			req.Scope = scope.scope
			req.ScopeID = scope.scopeID
			
			if err := rlm.budgetManager.RecordUsage(req); err != nil {
				return fmt.Errorf("failed to record API usage for scope %s:%s: %w", scope.scope, scope.scopeID, err)
			}
		}
	}
	
	return nil
}

// addRateLimitHeaders adds rate limit information to response headers
func (rlm *RateLimiterMiddleware) addRateLimitHeaders(c *gin.Context, scope budget.BudgetScope, limit int64, response *budget.BudgetCheckResponse) {
	// Add standard rate limit headers
	c.Header(fmt.Sprintf("%s-Limit", rlm.headerPrefix), strconv.FormatInt(limit, 10))
	
	if remaining, exists := response.RemainingQuota["api_requests_count"]; exists {
		if remainingInt, ok := remaining.(int64); ok {
			c.Header(fmt.Sprintf("%s-Remaining", rlm.headerPrefix), strconv.FormatInt(remainingInt, 10))
		}
	}
	
	if !response.ResetTime.IsZero() {
		c.Header(fmt.Sprintf("%s-Reset", rlm.headerPrefix), strconv.FormatInt(response.ResetTime.Unix(), 10))
		c.Header(fmt.Sprintf("%s-Reset-After", rlm.headerPrefix), strconv.FormatInt(int64(time.Until(response.ResetTime).Seconds()), 10))
	}
	
	// Add scope information for debugging
	c.Header(fmt.Sprintf("%s-Scope", rlm.headerPrefix), fmt.Sprintf("%s:%s", scope, c.GetString("scope_id")))
}

// handleRateLimitExceeded handles rate limit exceeded scenarios
func (rlm *RateLimiterMiddleware) handleRateLimitExceeded(c *gin.Context, err error) {
	if rateLimitErr, ok := err.(*RateLimitError); ok {
		// Add rate limit headers
		c.Header(fmt.Sprintf("%s-Limit", rlm.headerPrefix), strconv.FormatInt(rateLimitErr.Limit, 10))
		c.Header(fmt.Sprintf("%s-Remaining", rlm.headerPrefix), "0")
		
		if !rateLimitErr.ResetTime.IsZero() {
			c.Header(fmt.Sprintf("%s-Reset", rlm.headerPrefix), strconv.FormatInt(rateLimitErr.ResetTime.Unix(), 10))
			c.Header(fmt.Sprintf("%s-Reset-After", rlm.headerPrefix), strconv.FormatInt(int64(time.Until(rateLimitErr.ResetTime).Seconds()), 10))
		}
		
		// Return 429 Too Many Requests
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "rate_limit_exceeded",
			"message": fmt.Sprintf("Rate limit exceeded for %s:%s", rateLimitErr.Scope, rateLimitErr.ScopeID),
			"details": gin.H{
				"scope":     rateLimitErr.Scope,
				"scope_id":  rateLimitErr.ScopeID,
				"limit":     rateLimitErr.Limit,
				"reset_at":  rateLimitErr.ResetTime,
			},
		})
	} else {
		// Generic error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "rate_limit_error",
			"message": "Failed to check rate limit",
		})
	}
	
	c.Abort()
}

// CreateDefaultAPIBudgets creates default API rate limit budgets
func (rlm *RateLimiterMiddleware) CreateDefaultAPIBudgets() error {
	defaultBudgets := []*budget.Budget{
		{
			ID:          "global-api-budget",
			Name:        "Global API Budget",
			Description: "Global API request limit per minute",
			Type:        budget.BudgetTypeAPIRequests,
			Scope:       budget.BudgetScopeGlobal,
			MaxCount:    100000, // 100k requests per minute globally
			TimeWindow:  budget.TimeWindowMinute,
			Enabled:     true,
		},
		{
			ID:          "org-api-budget",
			Name:        "Organization API Budget",
			Description: "API request limit per organization per minute",
			Type:        budget.BudgetTypeAPIRequests,
			Scope:       budget.BudgetScopeOrganization,
			MaxCount:    10000, // 10k requests per org per minute
			TimeWindow:  budget.TimeWindowMinute,
			Enabled:     true,
		},
		{
			ID:          "user-api-budget",
			Name:        "User API Budget",
			Description: "API request limit per user per minute",
			Type:        budget.BudgetTypeAPIRequests,
			Scope:       budget.BudgetScopeUser,
			MaxCount:    1000, // 1k requests per user per minute
			TimeWindow:  budget.TimeWindowMinute,
			Enabled:     true,
		},
		{
			ID:          "execute-api-budget",
			Name:        "Execute API Budget",
			Description: "Execution API request limit per user per minute",
			Type:        budget.BudgetTypeAPIRequests,
			Scope:       budget.BudgetScopeUser,
			MaxCount:    100, // 100 executions per user per minute
			TimeWindow:  budget.TimeWindowMinute,
			Enabled:     true,
		},
	}
	
	for _, budget := range defaultBudgets {
		if err := rlm.budgetManager.AddBudget(budget); err != nil {
			return fmt.Errorf("failed to create default API budget %s: %w", budget.ID, err)
		}
	}
	
	return nil
}

// RateLimitError represents a rate limit violation error
type RateLimitError struct {
	Scope      budget.BudgetScope      `json:"scope"`
	ScopeID    string                  `json:"scopeId"`
	Limit      int64                   `json:"limit"`
	ResetTime  time.Time               `json:"resetTime"`
	Violations []budget.BudgetViolation `json:"violations"`
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded for %s:%s (limit: %d)", e.Scope, e.ScopeID, e.Limit)
}

// IsRateLimitError checks if an error is a rate limit error
func IsRateLimitError(err error) bool {
	_, ok := err.(*RateLimitError)
	return ok
}

