# Retry Policies Documentation

## Overview

MOVA Engine supports configurable retry policies that allow you to define how workflows should handle failures and retries. Policies can be configured through YAML files and applied dynamically to workflow executions.

## Retry Profiles

### Aggressive Profile
Fast retry with minimal backoff for time-sensitive operations.

```yaml
name: aggressive
description: "Fast retry with minimal backoff for time-sensitive operations"
maxRetries: 3
initialDelay: "100ms"
maxDelay: "1s"
backoffMultiplier: 1.5
jitter: 0.1
timeout: "5s"
```

**Use cases:**
- User-facing operations
- Real-time data processing
- Interactive workflows

### Balanced Profile
Balanced retry with exponential backoff for general use cases.

```yaml
name: balanced
description: "Balanced retry with exponential backoff for general use cases"
maxRetries: 5
initialDelay: "500ms"
maxDelay: "10s"
backoffMultiplier: 2.0
jitter: 0.2
timeout: "30s"
```

**Use cases:**
- API integrations
- Database operations
- File processing

### Conservative Profile
Conservative retry with long intervals for resource-intensive operations.

```yaml
name: conservative
description: "Conservative retry with long intervals for resource-intensive operations"
maxRetries: 10
initialDelay: "2s"
maxDelay: "60s"
backoffMultiplier: 2.5
jitter: 0.3
timeout: "300s"
```

**Use cases:**
- Heavy computations
- External service calls
- Resource-intensive operations

## Policy Configuration

### Basic Policy Structure

```yaml
id: "policy-001"
name: "Timeout Handling Policy"
description: "Aggressive retry for timeout errors"
retryProfile: "aggressive"
enabled: true
conditions:
  - errorType: "timeout"
    httpStatus: 408
    errorMessagePattern: ".*timeout.*"
    actionType: "http_fetch"
budgetConstraints:
  maxRetriesPerWorkflow: 10
  maxRetriesPerSession: 50
  maxTotalRetryTime: "5m"
```

### Condition Types

#### Error Type Conditions
```yaml
conditions:
  - errorType: "timeout"
  - errorType: "network"
  - errorType: "rate_limit"
```

#### HTTP Status Conditions
```yaml
conditions:
  - httpStatus: 408  # Request Timeout
  - httpStatus: 429  # Too Many Requests
  - httpStatus: 500  # Internal Server Error
  - httpStatus: 502  # Bad Gateway
  - httpStatus: 503  # Service Unavailable
```

#### Error Message Pattern Conditions
```yaml
conditions:
  - errorMessagePattern: ".*timeout.*"
  - errorMessagePattern: ".*rate limit.*"
  - errorMessagePattern: ".*connection refused.*"
```

#### Action Type Conditions
```yaml
conditions:
  - actionType: "http_fetch"
  - actionType: "parse_json"
  - actionType: "set"
```

### Budget Constraints

```yaml
budgetConstraints:
  maxRetriesPerWorkflow: 10    # Max retries per single workflow
  maxRetriesPerSession: 50     # Max retries per user session
  maxTotalRetryTime: "5m"      # Max total time spent retrying
```

## Best Practices

### 1. Timeout Handling
Use aggressive profiles for timeout errors to maintain responsiveness.

```yaml
name: "Timeout Policy"
retryProfile: "aggressive"
conditions:
  - errorType: "timeout"
    httpStatus: 408
```

### 2. Rate Limiting
Use conservative profiles for rate limit errors to avoid overwhelming services.

```yaml
name: "Rate Limit Policy"
retryProfile: "conservative"
conditions:
  - httpStatus: 429
  - errorMessagePattern: ".*rate limit.*"
```

### 3. Network Issues
Use balanced profiles for network-related errors.

```yaml
name: "Network Policy"
retryProfile: "balanced"
conditions:
  - errorType: "network"
  - httpStatus: 502
  - httpStatus: 503
```

### 4. Service-Specific Policies
Create policies for specific external services.

```yaml
name: "Payment Service Policy"
retryProfile: "conservative"
conditions:
  - actionType: "http_fetch"
    errorMessagePattern: ".*payment.*service.*"
```

## Policy Application

### Via API
```bash
# Create policy
curl -X POST /api/v1/policies \
  -H "Content-Type: application/json" \
  -d @policy.yaml

# Apply policy to workflow
curl -X POST /api/v1/policies/apply \
  -H "Content-Type: application/json" \
  -d '{"policyId": "policy-001", "workflowId": "workflow-123"}'
```

### Via CLI
```bash
# List policies
mova policies list

# Apply policy from file
mova policies apply policy.yaml

# Export policy
mova policies export policy-001 output.yaml

# Delete policy
mova policies delete policy-001

# Show available profiles
mova policies profiles
```

### Via Console
Use the Retry Policy Editor in the MOVA Console to:
1. Select a retry profile
2. Define conditions
3. Set budget constraints
4. Preview YAML configuration
5. Apply the policy

## Policy Matching

The engine uses a scoring system to determine the best matching policy:

- **Error Type Match**: +10 points
- **HTTP Status Match**: +8 points
- **Action Type Match**: +6 points
- **Error Message Pattern Match**: +5 points

The policy with the highest score is applied. If no policies match, the default balanced profile is used.

## Monitoring and Observability

### Policy Metrics
- Policy application count
- Policy match scores
- Retry success/failure rates
- Policy execution time

### Logging
```json
{
  "level": "info",
  "message": "Policy applied to workflow",
  "workflowId": "workflow-123",
  "policyId": "policy-001",
  "profile": "aggressive",
  "score": 15,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Troubleshooting

### Common Issues

1. **Policy Not Applied**
   - Check if policy is enabled
   - Verify conditions match the error context
   - Check policy validation errors

2. **Unexpected Retry Behavior**
   - Verify retry profile settings
   - Check budget constraints
   - Review condition logic

3. **Performance Issues**
   - Monitor retry frequency
   - Check timeout settings
   - Review backoff multipliers

### Debug Commands
```bash
# Check policy engine status
mova policies status

# Validate policy file
mova policies validate policy.yaml

# Test policy matching
mova policies test --error-type timeout --http-status 408
```

## Examples

### Complete Policy Examples

#### E-commerce Order Processing
```yaml
id: "order-processing-policy"
name: "Order Processing Retry Policy"
description: "Handles retries for order processing workflows"
retryProfile: "balanced"
enabled: true
conditions:
  - actionType: "http_fetch"
    errorMessagePattern: ".*order.*"
  - httpStatus: 500
    errorMessagePattern: ".*database.*"
budgetConstraints:
  maxRetriesPerWorkflow: 5
  maxRetriesPerSession: 20
  maxTotalRetryTime: "10m"
```

#### Data Synchronization
```yaml
id: "data-sync-policy"
name: "Data Sync Retry Policy"
description: "Conservative retry for data synchronization"
retryProfile: "conservative"
enabled: true
conditions:
  - errorType: "network"
  - httpStatus: 503
  - errorMessagePattern: ".*sync.*"
budgetConstraints:
  maxRetriesPerWorkflow: 15
  maxRetriesPerSession: 100
  maxTotalRetryTime: "1h"
```

## Migration Guide

### From Static Retry Configuration
1. Export existing retry settings
2. Create corresponding policies
3. Update workflow definitions
4. Test policy application
5. Remove static configurations

### Version Compatibility
- MOVA v3.1+ supports retry policies
- Legacy retry configurations are automatically converted
- Backward compatibility maintained for existing workflows

