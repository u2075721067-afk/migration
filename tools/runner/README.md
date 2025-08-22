# Navigator Agent Runner Service

FastAPI-based microservice that provides safe command execution for MOVA Engine demonstrations.

## Overview

The Runner Service is a critical security component of the Navigator Agent system. It executes only pre-approved terminal commands through a strict allow-list mechanism, ensuring safe and controlled demonstrations of MOVA Engine capabilities.

## Key Features

- **🔒 Security First**: Only executes commands from allow-list
- **📋 Command Templates**: Parameterized command execution with validation
- **⏱️ Timeout Protection**: Configurable execution timeouts
- **📊 Rate Limiting**: Built-in rate limiting (5 req/min)
- **🚫 Path Sandboxing**: File paths restricted to PROJECT_ROOT
- **📝 JSON Logging**: Structured execution logs

## Installation

```bash
cd tools/runner
pip install -r requirements.txt
```

## Configuration

### Environment Variables

- `PROJECT_ROOT`: Base directory for path sanitization (default: `.`)
- `RUNNER_PORT`: Service port (default: `9090`)
- `RUNNER_BIND`: Bind address (default: `127.0.0.1`)

### Allow-list Configuration

Commands are defined in `runner.allowlist.yaml`. Each command has:

```yaml
commands:
  build:
    - "make"
    - "build"
  validate:
    - "mova"
    - "validate"
    - {"file": {"type": "file", "required": true}}
```

## Usage

### Start Service

```bash
python runner.py
# or
uvicorn runner:app --host 127.0.0.1 --port 9090
```

### API Endpoints

#### Health Check
```bash
curl http://localhost:9090/health
```

#### Execute Command
```bash
curl -X POST http://localhost:9090/run \
  -H "Content-Type: application/json" \
  -d '{
    "cmd_id": "build",
    "args": {},
    "dry_run": false,
    "timeout_sec": 30
  }'
```

### Dry Run (Testing)

Set `"dry_run": true` to see what command would be executed without actually running it:

```json
{
  "cmd_id": "validate",
  "args": {"file": "examples/test.json"},
  "dry_run": true
}
```

## Security

### Path Validation
- All file paths are resolved relative to `PROJECT_ROOT`
- `../` patterns are blocked
- Absolute paths outside `PROJECT_ROOT` are rejected

### Command Validation
- Only commands in allow-list can be executed
- Parameters are strictly validated
- No shell interpretation (direct argv execution)

### Rate Limiting
- 5 requests per minute per client
- HTTP 429 status for violations

## Development

### Project Structure

```
tools/runner/
├── runner.py           # Main FastAPI application
├── runner.allowlist.yaml # Command allow-list
├── requirements.txt    # Python dependencies
└── README.md          # This file
```

### Testing

```bash
# Install test dependencies
pip install pytest httpx

# Run tests
pytest tests/
```

### Logging

The service outputs structured JSON logs:

```json
{"ts":"2025-08-21T15:01:23.123Z","action":"build","argv":["make","build"],"rc":0,"dur_ms":1234}
```

## Integration

The Runner Service integrates with MOVA Engine through:

1. **HTTP API**: Standard REST endpoints
2. **JSON Schema**: Structured request/response validation
3. **Error Handling**: Consistent error responses
4. **Monitoring**: Health checks and metrics

## Troubleshooting

### Common Issues

1. **Allow-list not found**: Ensure `runner.allowlist.yaml` exists
2. **Path validation errors**: Check file paths are within `PROJECT_ROOT`
3. **Rate limiting**: Wait 60 seconds between requests
4. **Timeout errors**: Increase `timeout_sec` or check command execution

### Debug Mode

Set environment variable for detailed logging:

```bash
DEBUG=1 python runner.py
```
