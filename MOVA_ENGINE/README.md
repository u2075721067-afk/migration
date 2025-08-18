# MOVA Automation Engine

## Overview
MOVA (Model-View-Automation) Engine is a JSON-based declarative automation platform for business processes and DevOps pipelines. This repository contains the MVP implementation of the MOVA v3.1 specification.

## Architecture
- **Core Engine**: Go-based executor for interpreting MOVA JSON actions
- **REST API**: HTTP endpoints for execution, validation, and management
- **CLI**: Command-line interface for running and validating workflows
- **SDK**: TypeScript and Python client libraries
- **Web Console**: Next.js-based web interface
- **Infrastructure**: Docker, CI/CD, and monitoring setup

## Quick Start
```bash
# Run a MOVA workflow
mova run workflow.json

# Validate a workflow
mova validate workflow.json

# View execution logs
mova logs <run_id>
```

## Project Structure
```
MOVA_ENGINE/
├── core/           # Core engine implementation
├── api/            # REST API server
├── cli/            # Command-line interface
├── sdk/            # Client libraries
├── web/            # Web console
├── infra/          # Infrastructure & deployment
├── docs/           # Documentation
└── tests/          # Test suite
```

## Development
- **Language**: Go (core), TypeScript (web), Python (SDK)
- **Database**: PostgreSQL + Redis
- **Message Queue**: Apache Kafka
- **Monitoring**: Prometheus + Grafana

## License
MIT License
