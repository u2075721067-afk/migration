# MOVA Console

Web interface for MOVA Automation Engine - Execute envelopes and monitor workflows.

## Overview

The MOVA Console is a Next.js web application that provides a user-friendly interface for:

- Executing MOVA envelopes with a visual editor
- Monitoring execution runs and their status
- Viewing detailed logs and results
- Managing and debugging workflows

## Features

### 🚀 Envelope Execution
- Interactive JSON editor with syntax highlighting
- Real-time validation and error handling
- One-click execution with progress indicators

### 📊 Run Monitoring
- List all execution runs with status indicators
- Filter and search through execution history
- Real-time status updates for running processes

### 🔍 Detailed Inspection
- View complete envelope definitions
- Examine execution results and context
- Browse JSONL logs with structured formatting
- Performance metrics and timing information

### 🎨 Modern UI/UX
- Built with Next.js 14 and TypeScript
- Styled with TailwindCSS and shadcn/ui components
- Responsive design for desktop and mobile
- Syntax highlighting with Prism.js

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   MOVA Console  │    │   MOVA Engine   │    │   PostgreSQL    │
│   (Next.js)     │◄──►│   REST API      │◄──►│   Database      │
│   Port: 3000    │    │   Port: 8080    │    │   Port: 5432    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Installation & Setup

### Development Mode

1. **Install dependencies:**
   ```bash
   cd mova-console
   npm install
   ```

2. **Configure environment:**
   ```bash
   # Create .env.local
   NEXT_PUBLIC_MOVA_API_URL=http://localhost:8080
   ```

3. **Start development server:**
   ```bash
   npm run dev
   ```

4. **Open browser:**
   ```
   http://localhost:3000
   ```

### Docker Deployment

The console is integrated into the main docker-compose setup:

1. **Start full stack:**
   ```bash
   cd MOVA_ENGINE/infra/docker
   docker-compose up -d
   ```

2. **Services available:**
   - **Console:** http://localhost:3000
   - **API:** http://localhost:8080
   - **Grafana:** http://localhost:3001
   - **Prometheus:** http://localhost:9090

### Production Build

```bash
# Build for production
npm run build

# Start production server
npm start
```

## Usage Guide

### 1. Executing Envelopes

1. Navigate to the home page
2. Edit the JSON envelope in the left panel
3. Click "Run Envelope" to execute
4. View results in the right panel

**Example envelope:**
```json
{
  "intent": "demo-workflow",
  "payload": {
    "message": "Hello MOVA"
  },
  "actions": [
    {
      "type": "set",
      "key": "greeting", 
      "value": "Hello from MOVA Engine!"
    },
    {
      "type": "sleep",
      "duration": 1000
    }
  ]
}
```

### 2. Monitoring Runs

1. Click "View Runs" or navigate to `/runs`
2. Browse execution history with status indicators:
   - 🟢 **Completed:** Successful execution
   - 🔴 **Failed:** Execution error
   - 🔵 **Running:** In progress
3. Click "View" to see detailed information

### 3. Analyzing Results

1. Select a run from the runs list
2. View the original envelope and execution result
3. Browse detailed JSONL logs with timestamps
4. Check performance metrics and timing data

## API Integration

The console uses the MOVA TypeScript SDK for API communication:

```typescript
import { movaClient } from '@/lib/api';

// Execute envelope
const result = await movaClient.execute(envelope);

// Validate envelope
const validation = await movaClient.validate(envelope);
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NEXT_PUBLIC_MOVA_API_URL` | MOVA API base URL | `http://localhost:8080` |
| `NODE_ENV` | Runtime environment | `development` |

### Docker Configuration

The console service in docker-compose.yml:

```yaml
mova-console:
  build:
    context: ../../../mova-console
    dockerfile: Dockerfile
  ports:
    - "3000:3000"
  environment:
    - NEXT_PUBLIC_MOVA_API_URL=http://mova-api:8080
  depends_on:
    - mova-api
  networks:
    - mova-network
```

## Development

### Project Structure

```
mova-console/
├── src/
│   ├── app/                 # Next.js app router pages
│   │   ├── page.tsx         # Home page (envelope editor)
│   │   ├── runs/
│   │   │   ├── page.tsx     # Runs list
│   │   │   └── [id]/        # Run details
│   │   └── layout.tsx       # Root layout
│   ├── components/
│   │   └── ui/              # shadcn/ui components
│   └── lib/
│       ├── api.ts           # API client using MOVA SDK
│       └── utils.ts         # Utility functions
├── public/                  # Static assets
├── Dockerfile              # Production container
└── package.json            # Dependencies and scripts
```

### Available Scripts

```bash
# Development
npm run dev          # Start dev server
npm run build        # Production build
npm run start        # Start production server
npm run lint         # Run ESLint
npm run type-check   # TypeScript checking
```

### Adding New Features

1. **New Pages:** Add to `src/app/` directory
2. **Components:** Create in `src/components/`
3. **API Calls:** Extend `src/lib/api.ts`
4. **Styling:** Use TailwindCSS classes and shadcn/ui components

## Troubleshooting

### Common Issues

**Console not connecting to API:**
- Verify `NEXT_PUBLIC_MOVA_API_URL` environment variable
- Check that MOVA Engine API is running on the specified port
- Ensure network connectivity between containers

**Build failures:**
- Clear Next.js cache: `rm -rf .next`
- Reinstall dependencies: `rm -rf node_modules && npm install`
- Check TypeScript errors: `npm run type-check`

**Docker issues:**
- Rebuild containers: `docker-compose build --no-cache`
- Check container logs: `docker-compose logs mova-console`
- Verify port conflicts

### Logs and Debugging

**Development logs:**
```bash
# Console logs
npm run dev

# Docker logs
docker-compose logs -f mova-console
```

**Production debugging:**
- Enable Next.js debug mode: `DEBUG=* npm start`
- Check browser console for client-side errors
- Monitor API requests in Network tab

## Contributing

1. Follow the existing code style and patterns
2. Add TypeScript types for new features
3. Update documentation for API changes
4. Test in both development and Docker environments
5. Ensure responsive design works on mobile devices

## License

Same as MOVA Engine project.
