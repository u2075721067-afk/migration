# MOVA Rule Engine & Low-Code Workflows

## Overview

The MOVA Rule Engine provides a powerful low-code solution for defining business rules and automation workflows. It allows users to create conditions and actions that are executed when certain criteria are met during workflow execution.

## Architecture

### Core Components

1. **Rule Engine** (`core/rules/`): Core logic for rule evaluation and execution
2. **API Handlers** (`api/handlers/rules.go`): REST API endpoints for rule management
3. **CLI Commands** (`cli/commands/rules.go`): Command-line interface for rules
4. **Web Console** (`mova-console/components/RuleBuilder.tsx`): Visual rule editor
5. **ConfigManager Integration**: Export/import rules with configuration management

## Rule Structure

### Rule Definition

```yaml
id: "example-rule-1"
name: "High Priority Alert"
description: "Send alert for high priority errors"
priority: 100
enabled: true
conditions:
  - field: "error.severity"
    operator: "=="
    value: "critical"
  - field: "retry_count"
    operator: ">"
    value: 3
actions:
  - type: "log"
    params:
      message: "Critical error detected with high retry count"
      level: "error"
  - type: "http_call"
    params:
      url: "https://alerts.example.com/webhook"
      method: "POST"
      headers:
        Content-Type: "application/json"
      body:
        alert_type: "critical_error"
        workflow_id: "{{workflow_id}}"
        error_message: "{{error.message}}"
```

### RuleSet Definition

```yaml
version: "1.0.0"
name: "Production Monitoring Rules"
description: "Rules for production workflow monitoring"
metadata:
  environment: "production"
  team: "platform"
rules:
  - id: "rule-1"
    name: "Database Connection Error"
    # ... rule definition
  - id: "rule-2"
    name: "API Rate Limit Exceeded"
    # ... rule definition
```

## Supported Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equals | `field: "status", operator: "==", value: "error"` |
| `!=` | Not equals | `field: "status", operator: "!=", value: "success"` |
| `>` | Greater than | `field: "retry_count", operator: ">", value: 3` |
| `>=` | Greater than or equal | `field: "duration", operator: ">=", value: 5000` |
| `<` | Less than | `field: "response_time", operator: "<", value: 1000` |
| `<=` | Less than or equal | `field: "memory_usage", operator: "<=", value: 80` |
| `contains` | Contains substring | `field: "error_message", operator: "contains", value: "timeout"` |
| `not_contains` | Does not contain | `field: "user_agent", operator: "not_contains", value: "bot"` |
| `regex` | Regular expression | `field: "email", operator: "regex", value: "^[^@]+@[^@]+\\.[^@]+"` |
| `in` | In list | `field: "status_code", operator: "in", value: [400, 401, 403]` |
| `not_in` | Not in list | `field: "method", operator: "not_in", value: ["GET", "HEAD"]` |
| `exists` | Field exists | `field: "user_id", operator: "exists"` |
| `not_exists` | Field does not exist | `field: "optional_field", operator: "not_exists"` |

## Supported Actions

### Set Variable
Sets a variable in the execution context.

```yaml
type: "set_var"
params:
  variable: "alert_sent"
  value: true
```

### Retry
Triggers a retry with specified profile.

```yaml
type: "retry"
params:
  profile: "aggressive"
  max_attempts: 5
  delay: 2000
```

### HTTP Call
Makes an HTTP request.

```yaml
type: "http_call"
params:
  url: "https://api.example.com/webhook"
  method: "POST"
  timeout: 30000
  headers:
    Authorization: "Bearer {{token}}"
    Content-Type: "application/json"
  body:
    event: "rule_triggered"
    rule_id: "{{rule_id}}"
    timestamp: "{{timestamp}}"
```

### Log
Logs a message.

```yaml
type: "log"
params:
  message: "Rule {{rule_id}} matched for workflow {{workflow_id}}"
  level: "info"
```

### Skip
Skips the current execution.

```yaml
type: "skip"
params:
  reason: "Maintenance window active"
```

### Route
Routes to a different workflow.

```yaml
type: "route"
params:
  workflow: "error_handling_workflow"
  reason: "Critical error detected"
```

### Stop
Stops the current execution.

```yaml
type: "stop"
params:
  reason: "Maximum retry limit reached"
```

### Transform
Transforms data in the context.

```yaml
type: "transform"
params:
  type: "uppercase"
  source: "user_input"
  target: "normalized_input"
```

Available transform types:
- `uppercase`: Convert to uppercase
- `lowercase`: Convert to lowercase
- `json_parse`: Parse JSON string to object
- `json_stringify`: Convert object to JSON string

## API Endpoints

### Rules Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/rules` | List all rules |
| GET | `/v1/rules/:id` | Get specific rule |
| POST | `/v1/rules` | Create new rule |
| PUT | `/v1/rules/:id` | Update existing rule |
| DELETE | `/v1/rules/:id` | Delete rule |
| POST | `/v1/rules/evaluate` | Dry-run rule evaluation |

### RuleSets Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/rulesets` | List all rulesets |
| GET | `/v1/rulesets/:name` | Get specific ruleset |
| POST | `/v1/rulesets` | Create/update ruleset |
| POST | `/v1/rulesets/validate` | Validate ruleset |

### Example API Usage

```bash
# List all rules
curl -X GET http://localhost:8080/v1/rules

# Create a new rule
curl -X POST http://localhost:8080/v1/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Rule",
    "description": "A test rule",
    "priority": 100,
    "enabled": true,
    "conditions": [
      {
        "field": "status",
        "operator": "==",
        "value": "error"
      }
    ],
    "actions": [
      {
        "type": "log",
        "params": {
          "message": "Error detected",
          "level": "error"
        }
      }
    ]
  }'

# Evaluate rules against context
curl -X POST http://localhost:8080/v1/rules/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "rules": [
      {
        "id": "test-rule",
        "name": "Test Rule",
        "conditions": [{"field": "status", "operator": "==", "value": "error"}],
        "actions": [{"type": "log", "params": {"message": "Test", "level": "info"}}]
      }
    ],
    "context": {
      "variables": {"status": "error"},
      "request": {},
      "response": {},
      "metadata": {}
    }
  }'
```

## CLI Commands

### Basic Operations

```bash
# List all rules
mova rules list

# List rules with filtering
mova rules list --enabled --priority 100 --format json

# Apply rules from file
mova rules apply rules.yaml

# Validate rules file
mova rules validate rules.yaml --verbose

# Evaluate rules against context
mova rules eval rules.yaml --context context.json --format summary
```

### Example Files

**rules.yaml:**
```yaml
version: "1.0.0"
name: "Example Rules"
description: "Example ruleset for demonstration"
rules:
  - id: "error-alert"
    name: "Error Alert Rule"
    description: "Send alert on errors"
    priority: 100
    enabled: true
    conditions:
      - field: "status"
        operator: "=="
        value: "error"
    actions:
      - type: "log"
        params:
          message: "Error detected: {{error.message}}"
          level: "error"
```

**context.json:**
```json
{
  "variables": {
    "status": "error",
    "retry_count": 2
  },
  "request": {
    "user_id": "12345",
    "action": "process_payment"
  },
  "response": {
    "status_code": 500
  },
  "metadata": {
    "workflow_id": "wf-123",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## Web Console

The Web Console provides a visual interface for managing rules:

### Features

1. **Rule List**: View all rules with status, priority, and actions
2. **Rule Editor**: Visual form-based editor for conditions and actions
3. **YAML View**: Direct YAML editing for advanced users
4. **Live Preview**: Real-time YAML generation from form inputs
5. **Rule Evaluation**: Test rules against sample contexts
6. **Import/Export**: Integration with ConfigManager

### Usage

1. Navigate to the Rules section in the Web Console
2. Click "Add Rule" to create a new rule
3. Fill in rule details, conditions, and actions
4. Use "Evaluate" to test the rule with sample data
5. Save the rule to apply it to your workflows

## Integration with ConfigManager

Rules can be exported and imported using the ConfigManager:

```bash
# Export rules to different formats
mova config export --include-rules --format yaml > config.yaml
mova config export --include-rules --format json > config.json
mova config export --include-rules --format hcl > config.hcl

# Import rules from configuration
mova config import config.yaml --mode merge
mova config import config.json --validate-only
mova config import config.hcl --dry-run
```

## Best Practices

### Rule Design

1. **Keep conditions simple**: Use clear, specific conditions
2. **Use meaningful names**: Make rule names descriptive
3. **Set appropriate priorities**: Higher priority rules execute first
4. **Test thoroughly**: Use the evaluation feature to test rules
5. **Document rules**: Add clear descriptions for maintainability

### Performance Considerations

1. **Limit rule count**: Too many rules can impact performance
2. **Optimize conditions**: Place most selective conditions first
3. **Use appropriate operators**: Choose the most efficient operator
4. **Monitor execution time**: Track rule evaluation performance

### Security Considerations

1. **Validate inputs**: Ensure rule conditions validate untrusted data
2. **Limit HTTP calls**: Be cautious with external HTTP actions
3. **Sanitize variables**: Clean data before setting variables
4. **Use timeouts**: Set appropriate timeouts for HTTP actions

## Troubleshooting

### Common Issues

1. **Rule not matching**: Check condition field names and values
2. **Action not executing**: Verify action parameters and types
3. **Performance issues**: Review rule count and complexity
4. **Import failures**: Validate YAML/JSON syntax

### Debug Mode

Enable debug logging to troubleshoot rule execution:

```bash
export MOVA_LOG_LEVEL=debug
mova rules eval rules.yaml --context context.json
```

### Validation Errors

Common validation errors and solutions:

- **"Unknown operator"**: Check operator spelling and supported operators
- **"Unknown action type"**: Verify action type is supported
- **"Field is required"**: Ensure all required fields are provided
- **"Duplicate rule ID"**: Use unique IDs for all rules in a ruleset

## Examples

### E-commerce Order Processing

```yaml
version: "1.0.0"
name: "E-commerce Rules"
description: "Rules for order processing workflows"
rules:
  - id: "high-value-order"
    name: "High Value Order Alert"
    description: "Alert for orders over $1000"
    priority: 200
    enabled: true
    conditions:
      - field: "order.total"
        operator: ">"
        value: 1000
    actions:
      - type: "log"
        params:
          message: "High value order: ${{order.total}}"
          level: "info"
      - type: "http_call"
        params:
          url: "https://alerts.company.com/high-value-order"
          method: "POST"
          body:
            order_id: "{{order.id}}"
            total: "{{order.total}}"
            customer: "{{customer.email}}"

  - id: "payment-failure"
    name: "Payment Failure Handler"
    description: "Handle payment failures with retry logic"
    priority: 150
    enabled: true
    conditions:
      - field: "payment.status"
        operator: "=="
        value: "failed"
      - field: "retry_count"
        operator: "<"
        value: 3
    actions:
      - type: "retry"
        params:
          profile: "payment_retry"
          max_attempts: 3
          delay: 5000
      - type: "set_var"
        params:
          variable: "payment_retry_attempted"
          value: true
```

### System Monitoring

```yaml
version: "1.0.0"
name: "System Monitoring Rules"
description: "Rules for system health monitoring"
rules:
  - id: "high-cpu-usage"
    name: "High CPU Usage Alert"
    description: "Alert when CPU usage exceeds 80%"
    priority: 100
    enabled: true
    conditions:
      - field: "system.cpu_usage"
        operator: ">"
        value: 80
    actions:
      - type: "log"
        params:
          message: "High CPU usage detected: {{system.cpu_usage}}%"
          level: "warning"
      - type: "http_call"
        params:
          url: "https://monitoring.company.com/alerts"
          method: "POST"
          headers:
            Authorization: "Bearer {{monitoring_token}}"
          body:
            alert_type: "high_cpu"
            value: "{{system.cpu_usage}}"
            threshold: 80

  - id: "disk-space-low"
    name: "Low Disk Space Alert"
    description: "Alert when disk space is below 10%"
    priority: 200
    enabled: true
    conditions:
      - field: "system.disk_free_percent"
        operator: "<"
        value: 10
    actions:
      - type: "log"
        params:
          message: "Low disk space: {{system.disk_free_percent}}% remaining"
          level: "error"
      - type: "stop"
        params:
          reason: "Insufficient disk space for continued operation"
```

This documentation provides comprehensive coverage of the MOVA Rule Engine, from basic concepts to advanced usage patterns and real-world examples.
