# MOVA Engine Deployment Guide

This guide covers deploying MOVA Engine using Docker and Docker Compose.

## Prerequisites

- Docker 20.10+ and Docker Compose 2.0+
- Git (for cloning the repository)
- 2GB+ RAM, 10GB+ disk space

## Quick Start with Docker Compose

### 1. Clone Repository

```bash
git clone https://github.com/your-org/mova-engine.git
cd mova-engine
```

### 2. Deploy with Docker Compose

```bash
# Navigate to docker directory
cd infra/docker

# Start all services
docker-compose up -d

# Check status
docker-compose ps
```

### 3. Verify Deployment

```bash
# Check API health
curl http://localhost:8080/health

# Access Web Console
open http://localhost:3000
```

## Services

### MOVA API (Port 8080)

- **Health Check**: `GET /health`
- **API Documentation**: `GET /v1/schemas`
- **Execute Workflow**: `POST /v1/execute`

### MOVA Console (Port 3000)

- Web interface for running and monitoring workflows
- Real-time execution logs
- Schema validation

## Configuration

### Environment Variables

#### MOVA API
```bash
MOVA_PORT=8080              # API server port
MOVA_HOST=0.0.0.0          # Bind address
MOVA_LOG_LEVEL=info        # Log level (debug, info, warn, error)
MOVA_STATE_DIR=/app/state  # State persistence directory
```

#### MOVA Console
```bash
NODE_ENV=production                    # Runtime environment
NEXT_PUBLIC_API_URL=http://mova-api:8080  # API endpoint
```

### Volumes

- `mova_state`: Persistent workflow state and context
- `mova_logs`: Application logs and execution history

## Production Deployment

### 1. Using Pre-built Images

```bash
# Pull latest images
docker pull ghcr.io/your-org/mova-engine/mova-engine:latest

# Update docker-compose.yml to use pre-built image
services:
  mova-api:
    image: ghcr.io/your-org/mova-engine/mova-engine:latest
    # Remove build section
```

### 2. Resource Limits

Add resource constraints to `docker-compose.yml`:

```yaml
services:
  mova-api:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

### 3. Security Hardening

```yaml
services:
  mova-api:
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp:rw,noexec,nosuid,size=100m
```

### 4. Reverse Proxy (Nginx)

```nginx
upstream mova-api {
    server localhost:8080;
}

upstream mova-console {
    server localhost:3000;
}

server {
    listen 80;
    server_name your-domain.com;
    
    location /api/ {
        proxy_pass http://mova-api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    location / {
        proxy_pass http://mova-console/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Monitoring

### Health Checks

```bash
# API health
curl -f http://localhost:8080/health || exit 1

# Container health
docker-compose ps --filter health=healthy
```

### Logs

```bash
# View all logs
docker-compose logs -f

# API logs only
docker-compose logs -f mova-api

# Console logs only
docker-compose logs -f mova-console
```

### Metrics

MOVA Engine exposes Prometheus metrics at `/metrics`:

```bash
curl http://localhost:8080/metrics
```

## Troubleshooting

### Common Issues

#### API not responding
```bash
# Check container status
docker-compose ps

# Check logs
docker-compose logs mova-api

# Restart service
docker-compose restart mova-api
```

#### Console can't connect to API
```bash
# Verify network connectivity
docker-compose exec mova-console curl http://mova-api:8080/health

# Check environment variables
docker-compose exec mova-console env | grep API_URL
```

#### Permission issues with volumes
```bash
# Fix volume permissions
sudo chown -R 1001:1001 /var/lib/docker/volumes/docker_mova_state
```

### Debug Mode

Enable debug logging:

```bash
# Update docker-compose.yml
environment:
  - MOVA_LOG_LEVEL=debug

# Restart services
docker-compose restart
```

## Backup and Recovery

### Backup State

```bash
# Create backup
docker run --rm -v docker_mova_state:/data -v $(pwd):/backup alpine \
  tar czf /backup/mova-state-$(date +%Y%m%d).tar.gz -C /data .
```

### Restore State

```bash
# Restore from backup
docker run --rm -v docker_mova_state:/data -v $(pwd):/backup alpine \
  tar xzf /backup/mova-state-20240819.tar.gz -C /data
```

## Scaling

### Horizontal Scaling

```yaml
services:
  mova-api:
    deploy:
      replicas: 3
    # Add load balancer configuration
```

### Database Backend

For production, consider using external state storage:

```yaml
services:
  mova-api:
    environment:
      - MOVA_STATE_BACKEND=postgres
      - MOVA_DB_URL=postgresql://user:pass@db:5432/mova
```

## Support

- **Documentation**: `/docs`
- **Issues**: GitHub Issues
- **Logs**: `docker-compose logs`
- **Health**: `http://localhost:8080/health`

