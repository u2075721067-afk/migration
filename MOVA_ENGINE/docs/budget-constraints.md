# Budget Constraints & Quotas

## Overview

The MOVA Engine budget system provides comprehensive resource management and quota enforcement across different scopes and resource types. It enables fine-grained control over retries, API requests, CPU/memory usage, and execution time with policy-based enforcement and alerting.

## Core Concepts

### Budget Types

- **Retries** (`BudgetTypeRetries`): Limits the number of retry operations
- **Workflows** (`BudgetTypeWorkflows`): Limits the number of workflow executions
- **CPU** (`BudgetTypeCPU`): Limits CPU usage (0-1 scale)
- **Memory** (`BudgetTypeMemory`): Limits memory usage in bytes
- **API Requests** (`BudgetTypeAPIRequests`): Limits API request frequency
- **Execution Time** (`BudgetTypeExecution`): Limits total execution time

### Budget Scopes

- **Global** (`BudgetScopeGlobal`): System-wide limits
- **Organization** (`BudgetScopeOrganization`): Per-organization limits
- **User** (`BudgetScopeUser`): Per-user limits
- **Workflow** (`BudgetScopeWorkflow`): Per-workflow limits
- **Session** (`BudgetScopeSession`): Per-session limits

### Time Windows

- **Minute** (`TimeWindowMinute`): Resets every minute
- **Hour** (`TimeWindowHour`): Resets every hour
- **Day** (`TimeWindowDay`): Resets every day
- **Month** (`TimeWindowMonth`): Resets every month

## Configuration

### Basic Budget Configuration

```yaml
id: "user-retry-budget"
name: "User Retry Budget"
description: "Limit retry operations per user per hour"
type: "retries"
scope: "user"
maxCount: 100
timeWindow: "hour"
enabled: true
```

### Advanced Budget with Multiple Constraints

```yaml
id: "workflow-resource-budget"
name: "Workflow Resource Budget"
description: "Resource limits per workflow execution"
type: "memory"
scope: "workflow"
maxMemory: 104857600  # 100MB in bytes
maxDuration: "10m"
timeWindow: "hour"
enabled: true
```

### Policy-Based Enforcement

```yaml
id: "high-priority-policy"
name: "High Priority Enforcement Policy"
description: "Strict enforcement for critical resources"
enabled: true
priority: 100
scope: "organization"
scopeId: "org-123"
budgetTypes:
  - "retries"
  - "cpu"
  - "memory"
actions:
  - type: "block"
    severity: "high"
  - type: "alert"
    severity: "critical"
conditions:
  - type: "time"
    timeRange:
      start: "2024-01-01T00:00:00Z"
      end: "2024-12-31T23:59:59Z"
```

## Usage Examples

### 1. Retry Budget Control

```go
// Create retry controller
retryController := budget.NewRetryController(budgetManager)

// Check if retry is allowed
retryCtx := &budget.RetryContext{
    WorkflowID:     "workflow-123",
    UserID:         "user-456",
    OrganizationID: "org-789",
    ActionType:     "http_fetch",
    ErrorType:      "timeout",
    AttemptNumber:  3,
    ExecutionTime:  2 * time.Second,
}

response, err := retryController.CheckRetryAllowed(ctx, retryCtx)
if err != nil {
    return fmt.Errorf("retry budget check failed: %w", err)
}

if !response.Allowed {
    return fmt.Errorf("retry blocked: %s", response.Message)
}

// Record retry usage
err = retryController.RecordRetryUsage(ctx, retryCtx)
if err != nil {
    log.Printf("Failed to record retry usage: %v", err)
}
```

### 2. Resource Monitoring

```go
// Create resource monitor
resourceMonitor := budget.NewResourceMonitor(budgetManager)
resourceMonitor.Start()

// Check resource availability
resourceReq := &budget.ResourceCheckRequest{
    Scope:          budget.BudgetScopeWorkflow,
    ScopeID:        "workflow-123",
    RequiredCPU:    0.5,  // 50% CPU
    RequiredMemory: 1024 * 1024 * 100, // 100MB
    WorkflowID:     "workflow-123",
    UserID:         "user-456",
}

response, err := resourceMonitor.CheckResourceLimits(resourceReq)
if err != nil {
    return fmt.Errorf("resource check failed: %w", err)
}

if !response.Allowed {
    return fmt.Errorf("insufficient resources: %s", response.Message)
}
```

### 3. API Rate Limiting

```go
// In your Gin router setup
rateLimiter := middleware.NewRateLimiterMiddleware(budgetManager)

// Create default API budgets
err := rateLimiter.CreateDefaultAPIBudgets()
if err != nil {
    log.Fatalf("Failed to create API budgets: %v", err)
}

// Apply middleware
router.Use(rateLimiter.RateLimitMiddleware())
```

### 4. Policy Enforcement

```go
// Create policy enforcer
policyEnforcer := budget.NewPolicyEnforcer(
    budgetManager,
    retryController,
    resourceMonitor,
)

// Enforce policies
enforcementCtx := &budget.EnforcementContext{
    RequestType:    "workflow",
    UserID:         "user-123",
    OrganizationID: "org-456",
    WorkflowID:     "workflow-789",
    RequiredCPU:    0.3,
    RequiredMemory: 1024 * 1024 * 50, // 50MB
    Timestamp:      time.Now(),
}

result, err := policyEnforcer.EnforcePolicy(ctx, enforcementCtx)
if err != nil {
    return fmt.Errorf("policy enforcement failed: %w", err)
}

if !result.Allowed {
    return fmt.Errorf("request blocked by policies: %v", result.BlockedBy)
}
```

## Alerting Configuration

### Alert Rules

```yaml
id: "cpu-threshold-alert"
name: "CPU Threshold Alert"
description: "Alert when CPU usage exceeds 80%"
budgetTypes:
  - "cpu"
scopes:
  - "global"
  - "organization"
thresholds:
  - 0.8  # 80%
channels:
  - "email"
  - "slack"
template: "cpu-alert-template"
severity: "high"
cooldown: "5m"
enabled: true
```

### Alert Templates

```yaml
id: "cpu-alert-template"
name: "CPU Usage Alert"
subject: "High CPU Usage Alert - {{.BudgetName}}"
body: |
  CPU usage has exceeded the threshold.
  
  Details:
  - Budget: {{.BudgetName}}
  - Scope: {{.Scope}}:{{.ScopeID}}
  - Current Usage: {{.PercentageUsed}}
  - Threshold: 80%
  - Timestamp: {{.Timestamp}}
  
  Please investigate immediately.
format: "text"
```

### Alert Channels

```go
// Add log channel
logChannel := budget.NewLogChannel("log", "System Log")
alertManager.AddChannel(logChannel)

// Add email channel (custom implementation)
emailChannel := NewEmailChannel("email", "Email Notifications", 
    "smtp.example.com", "alerts@example.com")
alertManager.AddChannel(emailChannel)

// Add Slack channel (custom implementation)
slackChannel := NewSlackChannel("slack", "Slack Notifications", 
    "https://hooks.slack.com/services/...")
alertManager.AddChannel(slackChannel)
```

## Default Budgets

The system provides sensible defaults for common scenarios:

### Retry Budgets

- **Global**: 10,000 retries/hour
- **Per-User**: 100 retries/hour
- **Per-Workflow**: 50 retries/execution
- **Per-Session**: 20 retries/hour

### API Rate Limits

- **Global**: 100,000 requests/minute
- **Per-Organization**: 10,000 requests/minute
- **Per-User**: 1,000 requests/minute
- **Execute Endpoint**: 100 requests/minute per user

### Resource Limits

- **Global CPU**: 80% usage limit
- **Global Memory**: 2GB limit
- **Per-Workflow Memory**: 100MB limit
- **Per-User CPU**: 20% usage limit

## Monitoring and Observability

### Metrics

The budget system exposes comprehensive metrics:

```go
// Get budget status
budget, err := budgetManager.GetBudget("user-retry-budget")
if err != nil {
    return err
}

fmt.Printf("Budget: %s\n", budget.Name)
fmt.Printf("Current Usage: %d/%d (%.1f%%)\n", 
    budget.CurrentCount, 
    budget.MaxCount,
    float64(budget.CurrentCount)/float64(budget.MaxCount)*100)
fmt.Printf("Reset Time: %v\n", budget.ResetTime)
```

### Enforcement Statistics

```go
// Get enforcement statistics
stats := policyEnforcer.GetEnforcementStats()

for policyID, stat := range stats {
    fmt.Printf("Policy %s:\n", policyID)
    fmt.Printf("  Total Checks: %d\n", stat.TotalChecks)
    fmt.Printf("  Violations: %d\n", stat.TotalViolations)
    fmt.Printf("  Blocked Requests: %d\n", stat.BlockedRequests)
    fmt.Printf("  Last Violation: %v\n", stat.LastViolation)
}
```

### Alert History

```go
// Get alert history
history := alertManager.GetAlertHistory()

for _, event := range history {
    fmt.Printf("Alert %s: %s at %v\n", 
        event.AlertID, 
        event.Action, 
        event.Timestamp)
}
```

## Best Practices

### 1. Hierarchical Budget Design

Design budgets in a hierarchical manner:
- Start with global limits
- Add organization-level limits (typically 10x individual limits)
- Set user-level limits
- Define workflow/session-specific limits

### 2. Time Window Selection

Choose appropriate time windows:
- **Minute**: For high-frequency operations (API requests)
- **Hour**: For moderate-frequency operations (retries, workflows)
- **Day**: For resource-intensive operations
- **Month**: For billing and quota management

### 3. Alert Configuration

Configure alerts proactively:
- Set warning thresholds at 80% usage
- Set critical thresholds at 95% usage
- Use appropriate cooldown periods to avoid alert spam
- Include actionable information in alert messages

### 4. Policy Priorities

Assign policy priorities thoughtfully:
- Critical system protection: Priority 1000+
- Organization policies: Priority 500-999
- User policies: Priority 100-499
- Default policies: Priority 1-99

### 5. Monitoring and Tuning

Regularly monitor and tune budgets:
- Review violation patterns
- Adjust limits based on actual usage
- Monitor alert frequency and effectiveness
- Analyze enforcement statistics

## Troubleshooting

### Common Issues

1. **Budget Not Applied**
   - Check if budget is enabled
   - Verify scope and scope ID matching
   - Check time window and reset times

2. **False Violations**
   - Review budget calculation logic
   - Check for clock synchronization issues
   - Verify usage recording accuracy

3. **Performance Issues**
   - Monitor budget check frequency
   - Review policy complexity
   - Check resource monitor overhead

### Debug Commands

```go
// Enable debug logging
manager.SetDebugMode(true)

// Check budget calculation
response, err := manager.CheckBudget(req)
log.Printf("Budget check: allowed=%v, violations=%d", 
    response.Allowed, len(response.Violations))

// Verify usage tracking
usage := manager.GetUsage(budgetID)
log.Printf("Current usage: count=%d, window_start=%v", 
    usage.Count, usage.WindowStartTime)
```

## Integration Examples

### With Workflow Executor

```go
// Before workflow execution
enforcementCtx := &budget.EnforcementContext{
    RequestType:    "workflow",
    WorkflowID:     workflow.ID,
    UserID:         workflow.UserID,
    OrganizationID: workflow.OrganizationID,
    RequiredCPU:    workflow.EstimatedCPU,
    RequiredMemory: workflow.EstimatedMemory,
    Timestamp:      time.Now(),
}

result, err := policyEnforcer.EnforcePolicy(ctx, enforcementCtx)
if err != nil {
    return fmt.Errorf("policy check failed: %w", err)
}

if !result.Allowed {
    return &WorkflowBlockedError{
        Reason:    "Budget constraints",
        Policies:  result.BlockedBy,
        RetryAfter: result.RetryAfter,
    }
}

// Execute workflow...

// Record actual usage
actualUsage := &budget.BudgetCheckRequest{
    Type:           budget.BudgetTypeWorkflows,
    Count:          1,
    Duration:       executionTime,
    Memory:         peakMemoryUsage,
    CPU:            averageCPUUsage,
    WorkflowID:     workflow.ID,
    UserID:         workflow.UserID,
    OrganizationID: workflow.OrganizationID,
}

err = budgetManager.RecordUsage(actualUsage)
if err != nil {
    log.Printf("Failed to record workflow usage: %v", err)
}
```

### With API Gateway

```go
// In API middleware
func BudgetMiddleware(budgetManager *budget.Manager) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract context
        userID := c.GetHeader("X-User-ID")
        orgID := c.GetHeader("X-Organization-ID")
        
        // Check API budget
        req := &budget.BudgetCheckRequest{
            Type:           budget.BudgetTypeAPIRequests,
            Scope:          budget.BudgetScopeUser,
            ScopeID:        userID,
            Count:          1,
            UserID:         userID,
            OrganizationID: orgID,
        }
        
        response, err := budgetManager.CheckBudget(req)
        if err != nil {
            c.JSON(500, gin.H{"error": "Budget check failed"})
            c.Abort()
            return
        }
        
        if !response.Allowed {
            c.Header("X-RateLimit-Reset", 
                response.ResetTime.Format(time.RFC3339))
            c.JSON(429, gin.H{
                "error": "Rate limit exceeded",
                "reset_time": response.ResetTime,
            })
            c.Abort()
            return
        }
        
        // Record usage
        budgetManager.RecordUsage(req)
        
        c.Next()
    }
}

