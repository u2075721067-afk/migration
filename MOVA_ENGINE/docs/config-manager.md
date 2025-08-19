# MOVA Config Manager

## Overview

The MOVA Config Manager provides comprehensive configuration export/import capabilities for the MOVA Automation Engine. It supports multiple formats (JSON, YAML, HCL) and provides validation, merge, and overwrite operations.

## Features

- **Multi-format Support**: Export/import in JSON, YAML, and HCL formats
- **Comprehensive Validation**: Detailed validation with error reporting
- **Flexible Import Modes**: Merge, overwrite, and validate-only modes
- **Dry-run Support**: Preview changes without applying them
- **Version Control**: Built-in versioning and checksum validation
- **Terraform Integration**: HCL export for infrastructure as code

## Supported Configuration Objects

### 1. Policies
- Retry policies with conditions
- Budget constraints
- Error handling rules

### 2. Budgets & Quotas
- Resource limits (CPU, memory, API calls)
- Time-based constraints
- Scope-based enforcement

### 3. Retry Profiles
- Exponential backoff configurations
- Jitter and timeout settings
- Custom retry strategies

### 4. DLQ Entries (Sandbox Mode)
- Dead letter queue configurations
- Retry count tracking
- Error payload preservation

### 5. Workflow Intents
- Envelope configurations
- Action definitions
- Execution parameters

## API Endpoints

### Export Configuration
```http
GET /v1/config/export?format=yaml&include_dlq=true&include_workflows=true
```

**Query Parameters:**
- `format`: Export format (json, yaml, hcl) - defaults to yaml
- `include_dlq`: Include DLQ entries (true/false)
- `include_workflows`: Include workflows (true/false)
- `compress`: Enable compression (true/false)

**Response:** File download with appropriate Content-Type

### Import Configuration
```http
POST /v1/config/import?format=yaml&mode=merge&validate_only=false&dry_run=false
```

**Query Parameters:**
- `format`: Import format (json, yaml, hcl) - auto-detected from Content-Type if not specified
- `mode`: Import mode (merge, overwrite, validate)
- `validate_only`: Only validate, don't import (true/false)
- `dry_run`: Show what would be imported without applying (true/false)
- `overwrite`: Overwrite existing configurations (true/false)

**Request Body:** Configuration file content

**Response:**
```json
{
  "success": true,
  "imported": 5,
  "skipped": 2,
  "errors": [],
  "warnings": [],
  "summary": {
    "policies": 2,
    "budgets": 1,
    "retryProfiles": 1,
    "dlqEntries": 0,
    "workflows": 1
  }
}
```

### Validate Configuration
```http
POST /v1/config/validate?format=yaml
```

**Query Parameters:**
- `format`: File format (json, yaml, hcl) - auto-detected from Content-Type if not specified

**Request Body:** Configuration file content

**Response:**
```json
{
  "valid": false,
  "errors": [
    {
      "type": "policy",
      "message": "ID is required",
      "line": 15,
      "column": 5,
      "context": "policy.test-policy"
    }
  ],
  "count": 1
}
```

### Get Supported Formats
```http
GET /v1/config/formats
```

**Response:**
```json
{
  "formats": ["json", "yaml", "hcl"],
  "default": "yaml"
}
```

### Get Configuration Info
```http
GET /v1/config/info
```

**Response:**
```json
{
  "version": "1.0.0",
  "supported_formats": ["json", "yaml", "hcl"],
  "default_format": "yaml",
  "features": [
    "export_policies",
    "export_budgets",
    "export_retry_profiles",
    "export_dlq_entries",
    "export_workflows",
    "import_merge",
    "import_overwrite",
    "validation",
    "dry_run"
  ]
}
```

## CLI Commands

### Export Configuration
```bash
# Export to YAML (default)
mova config export --out=policies.yaml

# Export to JSON with DLQ entries
mova config export --format=json --out=config.json --include-dlq

# Export to HCL with workflows
mova config export --format=hcl --out=terraform.hcl --include-workflows

# Export with compression
mova config export --format=yaml --out=config.yaml --compress
```

### Import Configuration
```bash
# Import with merge mode (default)
mova config import policies.yaml

# Import with overwrite mode
mova config import config.json --mode=overwrite

# Validate only, don't import
mova config import policies.yaml --validate-only

# Dry run to preview changes
mova config import policies.yaml --dry-run

# Force overwrite
mova config import policies.yaml --overwrite
```

### Validate Configuration
```bash
# Validate YAML file
mova config validate policies.yaml

# Validate JSON file
mova config validate config.json --format=json

# Validate HCL file
mova config validate terraform.hcl --format=hcl
```

### Show Information
```bash
# Show configuration system info
mova config info
```

## Configuration Formats

### JSON Format
```json
{
  "metadata": {
    "version": "1.0.0",
    "format": "json",
    "generatedAt": "2024-01-15T10:30:00Z",
    "source": "mova-engine",
    "checksum": "sha256:abc123..."
  },
  "policies": [
    {
      "id": "retry-policy-1",
      "name": "Retry Policy 1",
      "description": "Policy for HTTP retries",
      "retryProfile": "balanced",
      "enabled": true,
      "conditions": [
        {
          "errorType": "http_error",
          "httpStatus": 500
        }
      ],
      "budgetConstraints": {
        "maxRetriesPerWorkflow": 10,
        "maxRetriesPerSession": 5
      }
    }
  ],
  "budgets": [
    {
      "id": "api-budget",
      "name": "API Request Budget",
      "description": "Global API request limit",
      "type": "api_requests",
      "scope": "global",
      "maxCount": 10000,
      "timeWindow": "1h",
      "enabled": true
    }
  ],
  "retryProfiles": [
    {
      "name": "balanced",
      "description": "Balanced retry strategy",
      "maxRetries": 5,
      "initialDelay": "500ms",
      "maxDelay": "10s",
      "backoffMultiplier": 2.0,
      "jitter": 0.2,
      "timeout": "30s"
    }
  ]
}
```

### YAML Format
```yaml
metadata:
  version: "1.0.0"
  format: "yaml"
  generatedAt: "2024-01-15T10:30:00Z"
  source: "mova-engine"
  checksum: "sha256:abc123..."

policies:
  - id: "retry-policy-1"
    name: "Retry Policy 1"
    description: "Policy for HTTP retries"
    retryProfile: "balanced"
    enabled: true
    conditions:
      - errorType: "http_error"
        httpStatus: 500
    budgetConstraints:
      maxRetriesPerWorkflow: 10
      maxRetriesPerSession: 5

budgets:
  - id: "api-budget"
    name: "API Request Budget"
    description: "Global API request limit"
    type: "api_requests"
    scope: "global"
    maxCount: 10000
    timeWindow: "1h"
    enabled: true

retryProfiles:
  - name: "balanced"
    description: "Balanced retry strategy"
    maxRetries: 5
    initialDelay: "500ms"
    maxDelay: "10s"
    backoffMultiplier: 2.0
    jitter: 0.2
    timeout: "30s"
```

### HCL Format (Terraform-compatible)
```hcl
# MOVA Configuration Export
# Generated at: 2024-01-15T10:30:00Z
# Version: 1.0.0
# Source: mova-engine

# Retry Profiles
retry_profile "balanced" {
  description = "Balanced retry strategy"
  max_retries = 5
  initial_delay = "500ms"
  max_delay = "10s"
  backoff_multiplier = 2.0
  jitter = 0.2
  timeout = "30s"
}

# Budget Constraints
budget "api-budget" {
  name = "API Request Budget"
  description = "Global API request limit"
  type = "api_requests"
  scope = "global"
  max_count = 10000
  time_window = "1h"
  enabled = true
}

# Policies
policy "retry-policy-1" {
  name = "Retry Policy 1"
  description = "Policy for HTTP retries"
  retry_profile = "balanced"
  enabled = true
  
  conditions {
    error_type = "http_error"
    http_status = 500
  }
  
  budget_constraints {
    max_retries_per_workflow = 10
    max_retries_per_session = 5
  }
}
```

## Import Modes

### 1. Merge Mode (Default)
- New configurations are added
- Existing configurations are updated if newer
- No data is lost
- Safe for production use

### 2. Overwrite Mode
- All existing configurations are replaced
- Use with caution
- Good for environment setup

### 3. Validate Mode
- Only validates configuration
- No changes are made
- Useful for CI/CD pipelines

## Validation Rules

### Required Fields
- All configurations must have required fields
- Version information is mandatory
- IDs must be unique within their scope

### Data Types
- Time durations must be valid Go duration strings
- Numeric values must be within valid ranges
- Boolean flags must be true/false

### Logical Constraints
- Max delay must be >= initial delay
- Jitter must be between 0 and 1.0
- HTTP status codes must be valid (100-599)

### Business Rules
- Retry profiles must exist before policies reference them
- Budget constraints must be positive
- Time windows must be positive durations

## Error Handling

### Validation Errors
- Detailed error messages with context
- Line and column information when available
- Error categorization by type

### Import Errors
- Partial import support
- Detailed error reporting
- Rollback capabilities

### Format Errors
- Clear format detection
- Helpful error messages
- Format conversion suggestions

## Best Practices

### 1. Version Control
- Always include version information
- Use semantic versioning
- Document breaking changes

### 2. Validation
- Validate before importing
- Use dry-run mode in production
- Test configurations in staging first

### 3. Backup
- Export current configuration before import
- Keep configuration backups
- Document configuration changes

### 4. Testing
- Test configurations in sandbox
- Validate all formats
- Test import/export cycles

## Integration Examples

### CI/CD Pipeline
```yaml
# .github/workflows/config-validation.yml
name: Validate Configuration
on: [push, pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Validate Config
        run: |
          mova config validate config/policies.yaml
          mova config validate config/budgets.yaml
```

### Terraform Integration
```hcl
# main.tf
resource "mova_config" "policies" {
  source = file("${path.module}/config/policies.hcl")
  format = "hcl"
  mode   = "merge"
}
```

### Kubernetes ConfigMap
```yaml
# k8s-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mova-config
data:
  policies.yaml: |
    # MOVA policies configuration
    policies:
      - id: "default-retry"
        name: "Default Retry Policy"
        retryProfile: "balanced"
        enabled: true
```

## Troubleshooting

### Common Issues

1. **Import Fails with Validation Errors**
   - Check required fields
   - Validate data types
   - Review business rules

2. **Format Detection Issues**
   - Specify format explicitly
   - Check file extensions
   - Verify Content-Type headers

3. **Permission Errors**
   - Check file permissions
   - Verify storage access
   - Review security policies

### Debug Mode
```bash
# Enable debug logging
export MOVA_LOG_LEVEL=debug
mova config import --debug config.yaml
```

### Support
- Check validation errors first
- Review configuration format
- Test with minimal configuration
- Check system logs for details

## Future Enhancements

- **Compression Support**: Gzip and other compression formats
- **Encryption**: Secure configuration storage
- **Templates**: Configuration templates and variables
- **Diff View**: Visual configuration differences
- **Rollback**: Automatic rollback on import failure
- **Audit Trail**: Configuration change history
- **Multi-environment**: Environment-specific configurations
- **Scheduled Imports**: Automated configuration updates
