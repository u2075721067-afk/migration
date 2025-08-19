# MOVA Engine Security

This document describes the security features and hardening measures implemented in MOVA Engine to ensure production-ready operation.

## Overview

MOVA Engine implements multiple layers of security controls to protect against:

- **Secret exposure** in logs and traces
- **Unlimited execution** through timeout controls
- **Dangerous network access** via allow/deny lists
- **Resource exhaustion** through size limits
- **Malicious payloads** via input validation

## Secret Redaction

### Automatic Secret Masking

All sensitive information is automatically redacted from logs and execution traces:

**Sensitive Key Patterns:**
- `authorization`, `x-api-key`, `password`, `secret`, `token`
- `key`, `credential`, `auth`, `bearer`, `jwt`
- `api_key`, `apikey`, `client_secret`, `access_token`, `refresh_token`

**Example:**
```json
// Original
{
  "headers": {
    "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "X-API-KEY": "sk-1234567890abcdef",
    "Content-Type": "application/json"
  }
}

// Redacted in logs
{
  "headers": {
    "Authorization": "Bearer *****",
    "X-API-KEY": "sk*****ef",
    "Content-Type": "application/json"
  }
}
```

### Masking Rules

- **Bearer tokens**: `Bearer *****`
- **Basic auth**: `Basic *****`
- **Short values** (≤4 chars): `*****`
- **Medium values** (5-8 chars): `a*****z`
- **Long values** (>8 chars): `ab*****yz`

### String Pattern Redaction

Secrets in URLs, JSON strings, and log messages are also redacted:

```
# URL parameters
https://api.com/data?password=secret123 → https://api.com/data?password=*****

# JSON fields
{"password": "secret123"} → {"password": "*****"}

# Authorization headers
Authorization: Bearer token123 → Authorization: Bearer *****
```

## HTTP Security Controls

### Allow/Deny Lists

**Default Allowed Hosts:**
- `api.github.com`
- `httpbin.org`
- `jsonplaceholder.typicode.com`
- `*.googleapis.com`
- `*.amazonaws.com`

**Default Denied Hosts:**
- `localhost`, `127.0.0.1`, `0.0.0.0`
- `*.internal`, `*.local`
- `metadata.google.internal`
- `169.254.169.254` (AWS metadata service)

**Default Allowed Ports:**
- `80` (HTTP), `443` (HTTPS)
- `8080`, `8443` (Alternative HTTP/HTTPS)

**Default Denied Ports:**
- `22` (SSH), `23` (Telnet), `25` (SMTP)
- `53` (DNS), `135` (RPC), `139`/`445` (SMB)
- `1433` (MSSQL), `1521` (Oracle), `3306` (MySQL)
- `3389` (RDP), `5432` (PostgreSQL), `6379` (Redis)

**Denied Networks:**
- Private networks: `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`
- Loopback: `127.0.0.0/8`
- Link-local: `169.254.0.0/16`
- IPv6 loopback: `::1/128`
- IPv6 private: `fc00::/7`, `fe80::/10`

### Size Limits

- **Max Request Size**: 10MB
- **Max Response Size**: 50MB
- **Max Log Entry Size**: 1MB

### Other Controls

- **User-Agent**: `MOVA-Engine/1.0`
- **Redirects**: Disabled by default
- **Schemes**: Only `http` and `https` allowed

## Timeout Controls

### Default Timeouts

- **HTTP Timeout**: 30 seconds
- **Action Timeout**: 5 minutes
- **Workflow Timeout**: 30 minutes

### Timeout Behavior

**Action Timeout:**
- Applied to each individual action
- Context cancellation propagated to HTTP requests
- Failed actions logged with timeout reason

**Workflow Timeout:**
- Applied to entire workflow execution
- Checked before each action execution
- Workflow fails immediately on timeout

**HTTP Timeout:**
- Applied to individual HTTP requests
- Cannot be overridden by action config
- Enforced at HTTP client level

## Configuration

### Default Configuration

```go
config := config.DefaultSecurityConfig()
executor := executor.NewExecutorWithConfig(config)
```

### Custom Configuration

```go
config := &config.SecurityConfig{
    HTTP: config.HTTPSecurityConfig{
        AllowedHosts: []string{"api.mycompany.com"},
        DeniedHosts:  []string{"internal.mycompany.com"},
        AllowedPorts: []int{80, 443},
        MaxRequestSize: 5 * 1024 * 1024, // 5MB
    },
    Timeouts: config.TimeoutSecurityConfig{
        HTTPTimeout:     15 * time.Second,
        ActionTimeout:   2 * time.Minute,
        WorkflowTimeout: 10 * time.Minute,
    },
}
```

### Environment-Based Configuration

```yaml
# security.yaml
http:
  allowed_hosts:
    - "api.production.com"
    - "*.trusted-partner.com"
  denied_hosts:
    - "*.internal"
    - "localhost"
  allowed_ports: [80, 443, 8080]
  denied_ports: [22, 3306, 6379]
  max_request_size: 10485760  # 10MB
  max_response_size: 52428800 # 50MB
  user_agent: "MyApp-MOVA/1.0"

logging:
  redact_secrets: true
  sensitive_keys: ["custom_token", "private_key"]
  max_log_entry_size: 1048576 # 1MB

timeouts:
  http_timeout: "30s"
  action_timeout: "5m"
  workflow_timeout: "30m"
```

## Security Examples

### Safe HTTP Request

```json
{
  "type": "http_fetch",
  "name": "get_user_data",
  "config": {
    "url": "https://api.github.com/user",
    "method": "GET",
    "headers": {
      "Authorization": "Bearer ghp_xxxxxxxxxxxx",
      "User-Agent": "MyApp/1.0"
    },
    "timeout_ms": 15000
  }
}
```

**Security Controls Applied:**
- URL validated against allow/deny lists
- Timeout limited to max HTTP timeout (30s)
- Authorization header redacted in logs
- Response size limited to 50MB
- User-Agent overridden to `MOVA-Engine/1.0`

### Blocked Request

```json
{
  "type": "http_fetch",
  "name": "blocked_request",
  "config": {
    "url": "http://localhost:3306/admin",
    "method": "POST"
  }
}
```

**Result:** Action fails with error:
```
security validation failed: host localhost is explicitly denied
```

## Best Practices

### 1. Principle of Least Privilege

- Use restrictive allow lists rather than broad deny lists
- Only allow necessary hosts and ports
- Set minimal timeout values for your use case

### 2. Secret Management

- Never hardcode secrets in workflow definitions
- Use environment variables or secure secret stores
- Regularly rotate API keys and tokens

### 3. Network Security

- Prefer HTTPS over HTTP
- Avoid internal network access
- Use public APIs with proper authentication

### 4. Monitoring

- Monitor failed security validations
- Set up alerts for timeout patterns
- Review logs for suspicious activity

### 5. Configuration Management

- Use version-controlled security configurations
- Test security rules in staging environments
- Document any security exceptions

## Security Testing

### Unit Tests

```bash
cd MOVA_ENGINE
go test ./core/executor -v -run TestSecurity
go test ./config -v -run TestSecurity
```

### Integration Tests

```bash
# Test with blocked URLs
curl -X POST http://localhost:8080/v1/execute \
  -H "Content-Type: application/json" \
  -d '{
    "intent": {"name": "security_test"},
    "actions": [{
      "type": "http_fetch",
      "name": "blocked_request",
      "config": {
        "url": "http://localhost:22/"
      }
    }]
  }'
```

Expected response: `security validation failed`

### Security Validation

1. **Secret Redaction Test**: Verify no secrets appear in logs
2. **Network Access Test**: Confirm blocked URLs are rejected
3. **Timeout Test**: Verify long-running actions are cancelled
4. **Size Limit Test**: Check large responses are truncated

## Troubleshooting

### Common Issues

**"security validation failed: host X is explicitly denied"**
- Host is in denied list or not in allowed list
- Check security configuration
- Consider if access is necessary

**"action timeout after 5m0s"**
- Action exceeded timeout limit
- Increase timeout in security config if needed
- Optimize action implementation

**"HTTP request failed: context deadline exceeded"**
- HTTP request timed out
- Check network connectivity
- Increase HTTP timeout if needed

**"failed to read response body: unexpected EOF"**
- Response exceeded size limit
- Increase MaxResponseSize if needed
- Consider streaming for large responses

### Debug Mode

Enable debug logging to see security decisions:

```go
executor := executor.NewExecutor()
executor.SetLogLevel(logrus.DebugLevel)
```

This will log:
- URL validation results
- Timeout applications
- Secret redaction operations
- Security policy decisions

## Security Updates

Keep MOVA Engine updated for latest security fixes:

```bash
go get -u github.com/your-org/mova-engine
```

Review security advisories and update configurations as needed.

## Compliance

MOVA Engine security features help meet compliance requirements:

- **Data Protection**: Secret redaction prevents credential leaks
- **Network Security**: Access controls prevent lateral movement
- **Audit Trail**: All security events are logged
- **Resource Limits**: Prevent denial-of-service attacks

For specific compliance requirements, consult your security team and adjust configurations accordingly.
