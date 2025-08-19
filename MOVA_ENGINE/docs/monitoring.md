# MOVA Engine Monitoring & Observability

This document covers the monitoring and observability features of MOVA Engine.

## Overview

MOVA Engine provides comprehensive observability through:

- **Prometheus Metrics** - Performance and business metrics
- **Grafana Dashboards** - Visual monitoring and alerting
- **Structured Logging** - JSON logs compatible with ELK, CloudWatch, Loki
- **OpenTelemetry Tracing** - Distributed tracing for workflow execution

## Metrics

### Prometheus Endpoint

Metrics are exposed at `http://localhost:8080/metrics` in Prometheus format.

### Available Metrics

#### HTTP Metrics
- `http_requests_total` - Total HTTP requests (labels: endpoint, method, status)
- `http_request_duration_seconds` - HTTP request duration histogram (labels: endpoint, method)

#### Workflow Metrics
- `workflow_runs_total` - Total workflow executions (labels: status)
- `workflow_duration_seconds` - Workflow execution duration histogram (labels: status)

#### Executor Metrics
- `executor_action_total` - Total executor actions (labels: type, status)
- `executor_action_duration_seconds` - Action execution duration histogram (labels: type)

#### System Metrics
- `active_goroutines` - Number of active goroutines
- `memory_usage_bytes` - Memory usage in bytes

### Custom Metrics Integration

```go
import "github.com/mova-engine/mova-engine/api"

// Record workflow metrics
RecordWorkflowMetrics("success", duration)
RecordWorkflowMetrics("error", duration)

// Record action metrics
RecordExecutorMetrics("http_fetch", "success", duration)
RecordExecutorMetrics("parse_json", "error", duration)
```

## Grafana Dashboards

### Setup

1. Access Grafana at `http://localhost:3001`
2. Login with `admin/admin`
3. Dashboards are automatically provisioned

### Available Dashboards

#### 1. API Health Dashboard
- Request rate and error rate
- Response time percentiles (P50, P95, P99)
- HTTP status code distribution
- Memory and goroutine usage

#### 2. Executor Dashboard
- Workflow execution statistics
- Action execution breakdown by type
- Duration distributions
- Success/failure rates

#### 3. System Dashboard
- CPU and memory usage
- Goroutine count over time
- Garbage collection metrics
- File descriptor usage

### Dashboard Customization

Dashboards are stored as JSON in `infra/monitoring/grafana/dashboards/`:

```bash
# Edit dashboard
vim infra/monitoring/grafana/dashboards/api-health.json

# Restart Grafana to reload
docker-compose restart grafana
```

## Structured Logging

### Log Format

All logs are output in JSON format for easy parsing:

```json
{
  "timestamp": "2024-01-19T10:30:45.123456789Z",
  "level": "info",
  "message": "HTTP request completed",
  "method": "POST",
  "path": "/v1/execute",
  "status": 200,
  "latency_ms": 150,
  "client_ip": "192.168.1.100",
  "user_agent": "curl/7.68.0"
}
```

### Log Levels

Set via `MOVA_LOG_LEVEL` environment variable:

- `debug` - Detailed execution information
- `info` - General application flow (default)
- `warn` - Warning conditions
- `error` - Error conditions

### Event Types

#### System Events
```json
{
  "event": "system_startup",
  "version": "1.0.0",
  "go_version": "go1.23.0"
}
```

#### Workflow Events
```json
{
  "event": "workflow_start",
  "run_id": "run_1642588245123456789",
  "intent": "process_data",
  "action_count": 5
}
```

#### Security Events
```json
{
  "event": "security_rate_limit_exceeded",
  "client_ip": "192.168.1.100",
  "endpoint": "/v1/execute"
}
```

### ELK Stack Integration

#### Logstash Configuration

```ruby
input {
  docker {
    type => "mova-engine"
  }
}

filter {
  if [type] == "mova-engine" {
    json {
      source => "message"
    }
    
    date {
      match => [ "timestamp", "ISO8601" ]
    }
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "mova-engine-%{+YYYY.MM.dd}"
  }
}
```

#### Elasticsearch Index Template

```json
{
  "template": "mova-engine-*",
  "mappings": {
    "properties": {
      "timestamp": {"type": "date"},
      "level": {"type": "keyword"},
      "message": {"type": "text"},
      "event": {"type": "keyword"},
      "run_id": {"type": "keyword"},
      "status": {"type": "keyword"},
      "latency_ms": {"type": "long"}
    }
  }
}
```

## OpenTelemetry Tracing

### Configuration

Tracing is configured via environment variables:

```bash
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
ENVIRONMENT=production
```

### Trace Structure

#### Workflow Execution
```
workflow.execute (run_id: run_123, intent: process_data)
├── action.execute (id: action_1, type: http_fetch)
├── action.execute (id: action_2, type: parse_json)
└── action.execute (id: action_3, type: set)
```

#### Custom Spans

```go
ctx, span := StartWorkflowSpan(ctx, runID, intent)
defer span.End()

// Add events
AddSpanEvent(ctx, "validation_complete", 
    attribute.Int("errors", 0))

// Set attributes
SetSpanAttributes(ctx,
    attribute.String("workflow.type", "data_processing"),
    attribute.Int("workflow.actions", actionCount))

// Handle errors
if err != nil {
    SetSpanError(ctx, err)
}
```

### Jaeger Integration

1. Access Jaeger UI at `http://localhost:16686`
2. Search traces by service: `mova-engine`
3. Filter by operation: `workflow.execute`, `action.execute`

## Alerting

### Prometheus Alert Rules

Create `infra/monitoring/prometheus/rules/alerts.yml`:

```yaml
groups:
- name: mova-engine
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High error rate detected"
      
  - alert: WorkflowFailures
    expr: rate(workflow_runs_total{status="error"}[5m]) > 0.05
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "High workflow failure rate"
```

### Grafana Alerts

1. Go to Alerting > Alert Rules
2. Create new rule with PromQL query
3. Set notification channels (Slack, email, etc.)

## Performance Monitoring

### Key Metrics to Monitor

1. **Request Rate**: `rate(http_requests_total[5m])`
2. **Error Rate**: `rate(http_requests_total{status=~"5.."}[5m])`
3. **Response Time P95**: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
4. **Memory Usage**: `memory_usage_bytes`
5. **Workflow Success Rate**: `rate(workflow_runs_total{status="success"}[5m])`

### SLIs/SLOs

#### Availability SLO: 99.9%
- **SLI**: `rate(http_requests_total{status!~"5.."}[5m]) / rate(http_requests_total[5m])`
- **Error Budget**: 0.1% (43.2 minutes/month)

#### Latency SLO: 95% < 500ms
- **SLI**: `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
- **Target**: < 0.5 seconds

#### Workflow Success SLO: 99%
- **SLI**: `rate(workflow_runs_total{status="success"}[5m]) / rate(workflow_runs_total[5m])`
- **Error Budget**: 1% failed workflows

## Troubleshooting

### High Memory Usage

```bash
# Check memory metrics
curl http://localhost:8080/metrics | grep memory_usage_bytes

# Check goroutine count
curl http://localhost:8080/metrics | grep active_goroutines
```

### Slow Requests

1. Check Grafana "Response Time Distribution" panel
2. Look at trace details in Jaeger
3. Review structured logs for specific requests

### Workflow Failures

```bash
# Query logs for failed workflows
curl -s http://localhost:8080/metrics | grep 'workflow_runs_total{status="error"}'

# Check Grafana "Executor Dashboard" for failure patterns
```

## Best Practices

1. **Set appropriate log levels** for different environments
2. **Use trace sampling** to reduce overhead (10% default)
3. **Monitor error budgets** and alert when approaching limits
4. **Regularly review dashboards** and adjust thresholds
5. **Correlate metrics, logs, and traces** for effective debugging
6. **Set up automated alerts** for critical metrics
7. **Use structured logging** for better searchability

## Integration Examples

### CloudWatch Logs

```yaml
# docker-compose.yml
services:
  mova-api:
    logging:
      driver: awslogs
      options:
        awslogs-group: mova-engine
        awslogs-region: us-east-1
        awslogs-stream-prefix: api
```

### Loki Integration

```yaml
# promtail config
clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
- job_name: mova-engine
  docker_sd_configs:
  - host: unix:///var/run/docker.sock
  relabel_configs:
  - source_labels: ['__meta_docker_container_name']
    regex: 'mova-api'
    target_label: 'job'
    replacement: 'mova-engine'
```

