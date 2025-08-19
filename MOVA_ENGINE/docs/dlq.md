# Dead Letter Queue & Retry Sandbox

This document covers the Dead Letter Queue (DLQ) and retry sandbox functionality in MOVA Engine.

## Overview

The Dead Letter Queue system provides robust error handling and recovery mechanisms for failed workflow executions:

- **Automatic Retry**: Actions are retried with configurable backoff strategies
- **Dead Letter Queue**: Failed workflows are stored for manual inspection and retry
- **Sandbox Mode**: Isolated retry execution for testing fixes
- **Management Interface**: API, CLI, and Web UI for DLQ operations

## Retry Policy Configuration

### Schema

Retry policies can be defined at the workflow level (in `intent.retry`) or action level (`action.retry`):

```json
{
  "retry": {
    "max_attempts": 3,
    "backoff": "exponential",
    "delay": 2,
    "max_delay": 300,
    "jitter": true
  }
}
```

### Backoff Strategies

#### Fixed Backoff
```json
{
  "backoff": "fixed",
  "delay": 5
}
```
- Constant delay between retries (5 seconds)

#### Linear Backoff
```json
{
  "backoff": "linear",
  "delay": 2
}
```
- Delay increases linearly: 2s, 4s, 6s, 8s...

#### Exponential Backoff
```json
{
  "backoff": "exponential",
  "delay": 1,
  "max_delay": 60,
  "jitter": true
}
```
- Delay doubles each retry: 1s, 2s, 4s, 8s, 16s...
- `max_delay` caps the maximum delay
- `jitter` adds randomness to prevent thundering herd

### Example Workflow with Retry Policy

```json
{
  "intent": {
    "name": "data-processing",
    "version": "1.0",
    "description": "Process external data with retry",
    "retry": {
      "max_attempts": 5,
      "backoff": "exponential",
      "delay": 2,
      "max_delay": 120,
      "jitter": true
    }
  },
  "actions": [
    {
      "id": "fetch_data",
      "type": "http_fetch",
      "params": {
        "url": "https://api.example.com/data"
      },
      "retry": {
        "max_attempts": 3,
        "backoff": "linear",
        "delay": 5
      }
    }
  ]
}
```

## Dead Letter Queue

### Storage Format

DLQ entries are stored as JSON files in `state/deadletter/`:

```json
{
  "id": "dlq_12345678-1234-1234-1234-123456789012",
  "run_id": "run_1642588245123456789",
  "created_at": "2024-01-19T10:30:45.123456789Z",
  "status": "active",
  "envelope": {
    "intent": { ... },
    "actions": [ ... ]
  },
  "context": {
    "run_id": "run_1642588245123456789",
    "status": "failed",
    "variables": { ... },
    "results": { ... }
  },
  "failed_action": {
    "id": "fetch_data",
    "type": "http_fetch",
    "params": { ... }
  },
  "error_details": {
    "last_error": "network timeout: failed to connect to remote server",
    "error_history": [
      "Attempt 1: connection refused",
      "Attempt 2: timeout after 30s",
      "Attempt 3: network timeout: failed to connect to remote server"
    ],
    "failure_reason": "max_retries_exceeded",
    "attempts": 3,
    "retry_policy": { ... }
  },
  "metadata": {
    "source": "executor",
    "priority": 1,
    "workflow_type": "data-processing",
    "retry_count": 0,
    "tags": ["data", "external-api"]
  }
}
```

### DLQ Entry Statuses

- **active**: Ready for manual retry or investigation
- **retrying**: Currently being retried
- **resolved**: Successfully retried and completed
- **archived**: Archived for historical purposes

## API Endpoints

### List DLQ Entries
```http
GET /v1/dlq?status=active&workflow_type=data-processing&limit=50
```

**Query Parameters:**
- `status`: Filter by status (active, retrying, resolved, archived)
- `workflow_type`: Filter by workflow type
- `user_id`: Filter by user ID
- `since`: Filter entries created after timestamp (RFC3339)
- `until`: Filter entries created before timestamp (RFC3339)
- `limit`: Limit number of results

**Response:**
```json
{
  "entries": [
    {
      "id": "dlq_12345678-1234-1234-1234-123456789012",
      "run_id": "run_1642588245123456789",
      "status": "active",
      "created_at": "2024-01-19T10:30:45.123456789Z",
      "envelope": { ... },
      "error_details": { ... },
      "metadata": { ... }
    }
  ],
  "count": 1,
  "filter": { ... }
}
```

### Get DLQ Entry Details
```http
GET /v1/dlq/{dlq-id}
```

### Retry DLQ Entry
```http
POST /v1/dlq/{dlq-id}/retry
Content-Type: application/json

{
  "sandbox_mode": true,
  "overrides": {
    "variables": {
      "api_url": "https://staging-api.example.com"
    }
  }
}
```

**Response:**
```json
{
  "message": "DLQ entry retry initiated",
  "dlq_id": "dlq_12345678-1234-1234-1234-123456789012",
  "retry_run_id": "retry_run_1642588245123456789_1642588300",
  "sandbox_mode": true,
  "execution_result": {
    "run_id": "retry_run_1642588245123456789_1642588300",
    "status": "completed",
    "start_time": "2024-01-19T10:35:00.123456789Z",
    "end_time": "2024-01-19T10:35:15.987654321Z"
  }
}
```

### Update DLQ Status
```http
PUT /v1/dlq/{dlq-id}/status
Content-Type: application/json

{
  "status": "archived"
}
```

### Archive DLQ Entry
```http
POST /v1/dlq/{dlq-id}/archive
```

### Delete DLQ Entry
```http
DELETE /v1/dlq/{dlq-id}
```

### DLQ Statistics
```http
GET /v1/dlq/stats
```

**Response:**
```json
{
  "total_entries": 25,
  "by_status": {
    "active": 10,
    "retrying": 2,
    "resolved": 8,
    "archived": 5
  },
  "by_workflow_type": {
    "data-processing": 15,
    "notification": 6,
    "reporting": 4
  },
  "oldest_entry": "2024-01-15T08:20:00.000Z",
  "newest_entry": "2024-01-19T10:30:45.123Z"
}
```

## CLI Commands

### List DLQ Entries
```bash
# List all active entries
mova dlq list --status active

# List with filtering
mova dlq list --workflow-type data-processing --limit 10

# Output as JSON
mova dlq list --format json
```

### Show DLQ Entry Details
```bash
# Show detailed information
mova dlq show dlq_12345678

# Output as JSON
mova dlq show dlq_12345678 --format json
```

### Retry DLQ Entry
```bash
# Retry in sandbox mode (default)
mova dlq retry dlq_12345678

# Retry in production mode
mova dlq retry dlq_12345678 --sandbox=false

# Wait for completion
mova dlq retry dlq_12345678 --wait
```

### Archive DLQ Entry
```bash
mova dlq archive dlq_12345678
```

### Delete DLQ Entry
```bash
# With confirmation prompt
mova dlq delete dlq_12345678

# Force delete without prompt
mova dlq delete dlq_12345678 --force
```

### DLQ Statistics
```bash
mova dlq stats
```

## Sandbox Mode

Sandbox mode provides isolated execution for testing workflow fixes:

### Features

- **Isolated Execution**: No impact on production systems
- **Variable Overrides**: Test with different parameters
- **Safe Testing**: Verify fixes before production retry
- **Detailed Logging**: Enhanced logging for debugging

### Usage

#### API
```json
{
  "sandbox_mode": true,
  "overrides": {
    "variables": {
      "api_url": "https://test-api.example.com",
      "timeout": 60
    },
    "secrets": {
      "api_key": "test_key_123"
    }
  }
}
```

#### CLI
```bash
mova dlq retry dlq_12345678 --sandbox
```

#### Web Console
- Select "Retry in Sandbox" button
- Configure variable overrides in modal
- Monitor execution in real-time

### Sandbox Execution Context

Sandbox runs include metadata to identify them:

```json
{
  "variables": {
    "__retry_metadata": {
      "original_run_id": "run_1642588245123456789",
      "dlq_id": "dlq_12345678-1234-1234-1234-123456789012",
      "sandbox_mode": true,
      "retry_count": 1
    }
  }
}
```

## Web Console Integration

### DLQ Dashboard

Access the DLQ dashboard at `/dlq` in the web console:

- **Entry List**: View all DLQ entries with filtering
- **Statistics**: Real-time DLQ statistics
- **Entry Details**: Detailed view with full context
- **Retry Actions**: One-click retry with sandbox/production options
- **Bulk Operations**: Archive or delete multiple entries

### Features

- **Real-time Updates**: Auto-refresh DLQ data
- **Advanced Filtering**: Filter by status, type, date range
- **Error Analysis**: Detailed error information and history
- **Retry Management**: Sandbox and production retry options
- **Export Capabilities**: Export DLQ data for analysis

## Best Practices

### Retry Policy Design

1. **Start Conservative**: Begin with 3 attempts and exponential backoff
2. **Set Reasonable Delays**: Avoid overwhelming external systems
3. **Use Jitter**: Prevent thundering herd problems
4. **Action-Specific Policies**: Different actions may need different strategies

### DLQ Management

1. **Regular Review**: Monitor DLQ entries regularly
2. **Root Cause Analysis**: Investigate patterns in failures
3. **Sandbox Testing**: Always test fixes in sandbox first
4. **Archive Old Entries**: Keep DLQ size manageable
5. **Alert on Growth**: Monitor DLQ size and alert on unusual growth

### Error Handling

1. **Meaningful Errors**: Provide clear error messages
2. **Context Preservation**: Include relevant context in DLQ entries
3. **Categorize Failures**: Use failure reasons for analysis
4. **Log Correlation**: Ensure logs can be correlated with DLQ entries

## Monitoring and Alerting

### Metrics

Monitor these key metrics:

- `dlq_entries_total{status="active"}` - Active DLQ entries
- `dlq_retry_attempts_total` - Total retry attempts
- `dlq_resolution_rate` - Successful retry rate
- `workflow_retry_duration_seconds` - Retry execution time

### Alerts

Set up alerts for:

- DLQ growth rate exceeding threshold
- High retry failure rate
- Long-running sandbox executions
- DLQ entries older than threshold

### Dashboards

Create Grafana dashboards showing:

- DLQ entry trends over time
- Retry success/failure rates
- Top failing workflow types
- Average time to resolution

## Troubleshooting

### Common Issues

#### High DLQ Growth
```bash
# Check DLQ statistics
mova dlq stats

# Identify problematic workflows
mova dlq list --format json | jq '.[] | .metadata.workflow_type' | sort | uniq -c
```

#### Retry Failures
```bash
# Check specific entry details
mova dlq show dlq_12345678

# Review error history
mova dlq show dlq_12345678 --format json | jq '.error_details.error_history'
```

#### Sandbox Issues
- Verify sandbox environment configuration
- Check variable overrides are correct
- Review sandbox execution logs
- Ensure test data availability

### Recovery Procedures

#### Mass Retry
```bash
# List active entries
mova dlq list --status active --format json > active_entries.json

# Retry each entry (script)
for dlq_id in $(jq -r '.[] | .id' active_entries.json); do
  mova dlq retry "$dlq_id" --sandbox
done
```

#### Emergency Cleanup
```bash
# Archive old resolved entries
mova dlq list --status resolved --format json | \
  jq -r '.[] | select(.created_at < "2024-01-01") | .id' | \
  xargs -I {} mova dlq archive {}
```

## Configuration

### Environment Variables

- `MOVA_DLQ_PATH`: DLQ storage directory (default: `./state/deadletter`)
- `MOVA_RETRY_DEFAULT_MAX_ATTEMPTS`: Default max retry attempts (default: `3`)
- `MOVA_RETRY_DEFAULT_DELAY`: Default retry delay in seconds (default: `2`)
- `MOVA_DLQ_RETENTION_DAYS`: DLQ entry retention period (default: `30`)

### Storage Configuration

```yaml
# docker-compose.yml
services:
  mova-api:
    volumes:
      - dlq_data:/app/state/deadletter
    environment:
      - MOVA_DLQ_PATH=/app/state/deadletter

volumes:
  dlq_data:
    driver: local
```

## Security Considerations

- **Sensitive Data**: DLQ entries may contain sensitive workflow data
- **Access Control**: Implement proper RBAC for DLQ operations
- **Audit Logging**: Log all DLQ management operations
- **Data Retention**: Implement appropriate data retention policies
- **Sandbox Isolation**: Ensure sandbox mode doesn't access production resources
