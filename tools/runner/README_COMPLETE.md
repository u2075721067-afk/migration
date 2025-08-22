# Navigator Agent Demo System

Complete demonstration system for showcasing MOVA Engine capabilities to investors and stakeholders.

## Overview

The Navigator Agent is a comprehensive demonstration system that provides:

- **🔒 Secure Command Execution**: Allow-list based command runner
- **🌐 Interactive Web Interface**: Modern demo interface for testing
- **🔗 MOVA Engine Integration**: Direct API integration with MOVA Engine
- **📊 Real-time Monitoring**: Live execution logs and results
- **✅ Validation & Testing**: Complete envelope validation workflow

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Interface │────│  Runner Service │────│  MOVA Engine   │
│   (Port 9091)   │    │   (Port 9090)   │    │   (Port 8080)   │
│                 │    │                 │    │                 │
│ • Demo UI       │    │ • Allow-list    │    │ • Validation    │
│ • Envelope Mgmt │    │ • Rate Limiting │    │ • Execution    │
│ • Real-time Logs│    │ • Path Security │    │ • Introspection│
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Quick Start

### Prerequisites
- Python 3.8+
- Running MOVA Engine instance at http://localhost:8080
- Demo envelopes in `envelopes/` directory

### Installation
```bash
cd tools/runner
pip install -r requirements.txt
```

### Start Services

1. **Start Runner Service** (Port 9090):
```bash
python runner.py
```
Service will be available at: http://localhost:9090

2. **Start Web Interface** (Port 9091):
```bash
python web_interface.py
```
Web interface will be available at: http://localhost:9091

### 3. Ensure MOVA Engine is Running
Make sure MOVA Engine API is running at http://localhost:8080

## Features

### 🔒 Security Features
- Command allow-list with strict validation
- Path sanitization and sandboxing
- Rate limiting (5 requests/minute)
- Input validation and sanitization
- Timeout protection for all operations

### 🌐 Web Interface
- **Introspect Engine**: Discover MOVA capabilities
- **Validate Envelope**: Check envelope syntax
- **Execute Envelope**: Run demonstrations
- **View Logs**: Real-time execution monitoring
- **Envelope Management**: Browse available demos

### 🔗 MOVA Integration
- Direct API calls to MOVA Engine
- Envelope validation via `/v1/validate`
- Envelope execution via `/v1/execute`
- Log retrieval via `/v1/runs/{id}/logs`
- Introspection via `/v1/introspect`

## Configuration

### Environment Variables
```bash
export PROJECT_ROOT="."           # Root for file operations
export RUNNER_PORT="9090"         # Runner service port
export RUNNER_BIND="127.0.0.1"    # Bind address
export MOVA_API_BASE="http://localhost:8080"  # MOVA Engine URL
export MOVA_API_TIMEOUT="30"      # API timeout in seconds
```

### Allow-list Configuration
Commands are defined in `runner.allowlist.yaml`:

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

## API Endpoints

### Runner Service (Port 9090)

- `GET /`: Health check
- `GET /health`: Detailed health status
- `POST /run`: Execute allow-listed commands (legacy)
- `POST /validate`: Validate MOVA envelope
- `POST /execute`: Execute MOVA envelope
- `GET /logs/{run_id}`: Get execution logs
- `GET /introspect`: Get MOVA Engine capabilities

### Web Interface (Port 9091)

- `GET /`: Interactive demo interface
- `GET /api/introspect`: Get engine capabilities
- `POST /api/validate`: Validate envelope via web form
- `POST /api/execute`: Execute envelope via web form
- `GET /api/logs/{run_id}`: Get logs via web form
- `GET /api/envelopes`: List available demo envelopes

## Usage Examples

### Using cURL

```bash
# Validate an envelope
curl -X POST http://localhost:9090/validate \
  -H "Content-Type: application/json" \
  -d '{"cmd_id": "validate", "args": {"file": "envelopes/demo_agent.json"}}'

# Execute an envelope
curl -X POST http://localhost:9090/execute \
  -H "Content-Type: application/json" \
  -d '{"cmd_id": "run", "args": {"file": "envelopes/demo_agent.json"}}'

# Get introspection
curl http://localhost:9090/introspect
```

### Using Web Interface

1. Open http://localhost:9091 in your browser
2. Click "Get Capabilities" to see MOVA Engine features
3. Select an envelope from the list
4. Click "Validate Selected" to validate it
5. Click "Execute Selected" to run it
6. View results and logs in real-time

## Security Notes

- All file operations are sandboxed to PROJECT_ROOT
- Commands are validated against strict allow-list
- Rate limiting prevents abuse (5 requests/minute)
- All inputs are sanitized and validated
- Execution timeouts prevent hanging processes
- Comprehensive logging for audit trails

## Development

### Adding New Demo Envelopes
1. Create envelope in `envelopes/` directory
2. Follow MOVA v3.1 specification
3. Test validation and execution
4. Update documentation

### Extending Allow-list
1. Edit `runner.allowlist.yaml`
2. Add new command templates
3. Test security implications
4. Update documentation

### Custom Web Interface
1. Modify `web_interface.py`
2. Update templates in `templates/`
3. Add static files to `static/`
4. Test all functionality

## Troubleshooting

### Connection Issues
- Ensure MOVA Engine is running on port 8080
- Check network connectivity
- Verify API endpoints are accessible

### Permission Issues
- Ensure PROJECT_ROOT has proper permissions
- Check file ownership for envelope files
- Verify command execution permissions

### Performance Issues
- Monitor system resources during execution
- Check timeout configurations
- Review rate limiting settings
