# MOVA TypeScript SDK

TypeScript/JavaScript SDK for the MOVA Automation Engine.

## Installation

```bash
npm install @mova-engine/sdk-typescript
```

## Quick Start

```typescript
import { MOVAClient, MOVAEnvelope } from '@mova-engine/sdk-typescript';

// Create client
const mova = new MOVAClient({
  baseURL: 'http://localhost:8080', // Optional, defaults to localhost:8080
  timeout: 30000, // Optional, defaults to 30s
});

// Define a workflow envelope
const envelope: MOVAEnvelope = {
  mova_version: '3.1',
  intent: {
    name: 'my-workflow',
    version: '1.0.0',
    description: 'Example workflow',
  },
  payload: {
    message: 'Hello, MOVA!',
  },
  actions: [
    {
      type: 'set',
      name: 'greeting',
      config: {
        variable: 'greeting',
        value: '{{payload.message}}',
      },
    },
    {
      type: 'print',
      name: 'print-greeting',
      config: {
        value: '{{greeting}}',
      },
    },
  ],
};

// Execute workflow
try {
  // Synchronous execution (wait for completion)
  const result = await mova.execute(envelope, true);
  console.log('Workflow completed:', result);

  // Asynchronous execution
  const asyncResult = await mova.execute(envelope, false);
  console.log('Workflow started:', asyncResult.run_id);
  
  // Check status later
  const status = await mova.getRun(asyncResult.run_id);
  console.log('Current status:', status.status);
  
  // Get logs
  const logs = await mova.getLogs(asyncResult.run_id);
  logs.forEach(log => console.log(log));
} catch (error) {
  console.error('Workflow execution failed:', error);
}
```

## API Reference

### MOVAClient

#### Constructor

```typescript
new MOVAClient(options?: MOVAClientOptions)
```

**Options:**
- `baseURL?: string` - Base URL of the MOVA API server (default: 'http://localhost:8080')
- `timeout?: number` - Request timeout in milliseconds (default: 30000)
- `headers?: Record<string, string>` - Additional headers to send with requests

#### Methods

##### execute(envelope, wait?)

Execute a MOVA workflow envelope.

```typescript
async execute(envelope: MOVAEnvelope, wait: boolean = false): Promise<ExecutionResult>
```

**Parameters:**
- `envelope: MOVAEnvelope` - The workflow envelope to execute
- `wait: boolean` - Whether to wait for execution to complete (default: false)

**Returns:** Promise resolving to execution result

##### validate(envelope)

Validate a MOVA envelope against the schema.

```typescript
async validate(envelope: MOVAEnvelope): Promise<ValidationResult>
```

**Parameters:**
- `envelope: MOVAEnvelope` - The envelope to validate

**Returns:** Promise resolving to validation result

##### getRun(runId)

Get the status and result of a workflow execution.

```typescript
async getRun(runId: string): Promise<ExecutionResult>
```

**Parameters:**
- `runId: string` - The run ID to retrieve

**Returns:** Promise resolving to execution result

##### getLogs(runId)

Get the logs for a workflow execution.

```typescript
async getLogs(runId: string): Promise<string[]>
```

**Parameters:**
- `runId: string` - The run ID to retrieve logs for

**Returns:** Promise resolving to array of JSONL log entries

##### getSchemas()

Get available schemas.

```typescript
async getSchemas(): Promise<any>
```

**Returns:** Promise resolving to schemas information

##### getSchema(name)

Get a specific schema by name.

```typescript
async getSchema(name: string): Promise<any>
```

**Parameters:**
- `name: string` - Schema name (e.g., 'envelope', 'action')

**Returns:** Promise resolving to schema definition

##### introspect()

Get API introspection information.

```typescript
async introspect(): Promise<any>
```

**Returns:** Promise resolving to API information

## Types

### MOVAEnvelope

```typescript
interface MOVAEnvelope {
  mova_version: string;
  intent: {
    name: string;
    version: string;
    description?: string;
    author?: string;
    tags?: string[];
    timeout?: number;
    retry?: {
      count: number;
      backoff_ms: number;
    };
    budget?: {
      tokens?: number;
      cost_usd?: number;
    };
  };
  payload: Record<string, any>;
  actions: Array<{
    type: string;
    name: string;
    description?: string;
    enabled?: boolean;
    timeout?: number;
    retry?: {
      count: number;
      backoff_ms: number;
    };
    config?: Record<string, any>;
  }>;
  variables?: Record<string, any>;
  secrets?: Record<string, string>;
}
```

### ExecutionResult

```typescript
interface ExecutionResult {
  run_id: string;
  workflow_id: string;
  start_time: string;
  end_time?: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  variables: Record<string, any>;
  results: Record<string, any>;
  logs: Array<{
    timestamp: string;
    level: string;
    step: string;
    type: string;
    action?: string;
    message: string;
    params_redacted?: Record<string, any>;
    status: string;
    data?: Record<string, any>;
  }>;
}
```

### ValidationResult

```typescript
interface ValidationResult {
  valid: boolean;
  message: string;
  errors?: string[];
}
```

## Examples

### Basic Workflow

```typescript
import { mova } from '@mova-engine/sdk-typescript';

const envelope = {
  mova_version: '3.1',
  intent: { name: 'hello-world', version: '1.0.0' },
  payload: { name: 'World' },
  actions: [
    {
      type: 'set',
      name: 'greeting',
      config: { variable: 'message', value: 'Hello, {{payload.name}}!' }
    },
    {
      type: 'print',
      name: 'output',
      config: { value: '{{message}}' }
    }
  ]
};

const result = await mova.execute(envelope, true);
console.log(result.variables.message); // "Hello, World!"
```

### HTTP Request Workflow

```typescript
const httpWorkflow = {
  mova_version: '3.1',
  intent: { name: 'fetch-data', version: '1.0.0' },
  payload: { url: 'https://api.example.com/data' },
  actions: [
    {
      type: 'http_fetch',
      name: 'fetch',
      config: {
        url: '{{payload.url}}',
        method: 'GET',
        headers: { 'Accept': 'application/json' }
      }
    },
    {
      type: 'parse_json',
      name: 'extract',
      config: { path: 'data.items[0].name' }
    }
  ]
};

const result = await mova.execute(httpWorkflow, true);
```

### Error Handling

```typescript
try {
  const result = await mova.execute(envelope, true);
  console.log('Success:', result);
} catch (error) {
  if (error instanceof Error) {
    console.error('Error:', error.message);
  }
}
```

### Validation

```typescript
const validation = await mova.validate(envelope);
if (validation.valid) {
  console.log('Envelope is valid');
} else {
  console.error('Validation errors:', validation.errors);
}
```

## Development

```bash
# Install dependencies
npm install

# Build
npm run build

# Run tests
npm test

# Run tests in watch mode
npm run test:watch

# Lint
npm run lint

# Format code
npm run format
```

## License

MIT
