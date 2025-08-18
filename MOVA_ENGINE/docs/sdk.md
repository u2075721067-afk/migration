# MOVA SDK Documentation

This document provides comprehensive information about the MOVA SDK packages for TypeScript/JavaScript and Python.

## Overview

The MOVA SDK provides easy-to-use client libraries for interacting with the MOVA Automation Engine API. Both TypeScript and Python SDKs offer the same core functionality:

- Execute MOVA workflows synchronously or asynchronously
- Validate MOVA envelopes against JSON schemas
- Retrieve execution status and logs
- Access API schemas and introspection information

## Installation

### TypeScript/JavaScript

```bash
npm install @mova-engine/sdk-typescript
```

### Python

```bash
pip install mova-engine-sdk
```

## Quick Start Examples

### TypeScript

```typescript
import { MOVAClient, MOVAEnvelope } from '@mova-engine/sdk-typescript';

const mova = new MOVAClient({ baseURL: 'http://localhost:8080' });

const envelope: MOVAEnvelope = {
  mova_version: '3.1',
  intent: {
    name: 'hello-world',
    version: '1.0.0',
  },
  payload: { name: 'World' },
  actions: [
    {
      type: 'set',
      name: 'greeting',
      config: {
        variable: 'message',
        value: 'Hello, {{payload.name}}!'
      }
    }
  ]
};

// Synchronous execution
const result = await mova.execute(envelope, true);
console.log('Result:', result.variables.message);

// Asynchronous execution
const asyncResult = await mova.execute(envelope, false);
const status = await mova.getRun(asyncResult.run_id);
```

### Python

```python
from mova import MOVAClient, MOVAEnvelope, Intent, Action

client = MOVAClient(base_url="http://localhost:8080")

envelope = MOVAEnvelope(
    mova_version="3.1",
    intent=Intent(name="hello-world", version="1.0.0"),
    payload={"name": "World"},
    actions=[
        Action(
            type="set",
            name="greeting",
            config={
                "variable": "message",
                "value": "Hello, {{payload.name}}!"
            }
        )
    ]
)

# Synchronous execution
result = client.execute(envelope, wait=True)
print(f"Result: {result.variables['message']}")

# Asynchronous execution
async_result = client.execute(envelope, wait=False)
status = client.get_run(async_result.run_id)
```

## API Methods Comparison

| Method | TypeScript | Python | Description |
|--------|------------|--------|-------------|
| Execute (sync) | `await mova.execute(envelope, true)` | `client.execute(envelope, wait=True)` | Execute workflow and wait for completion |
| Execute (async) | `await mova.execute(envelope, false)` | `client.execute(envelope, wait=False)` | Start workflow execution asynchronously |
| Validate | `await mova.validate(envelope)` | `client.validate(envelope)` | Validate envelope against schema |
| Get Run Status | `await mova.getRun(runId)` | `client.get_run(run_id)` | Get execution status and results |
| Get Logs | `await mova.getLogs(runId)` | `client.get_logs(run_id)` | Get execution logs (JSONL format) |
| Get Schemas | `await mova.getSchemas()` | `client.get_schemas()` | List available schemas |
| Get Schema | `await mova.getSchema(name)` | `client.get_schema(name)` | Get specific schema definition |
| Introspect | `await mova.introspect()` | `client.introspect()` | Get API information |

## Common Workflow Examples

### 1. Basic Variable Setting

**TypeScript:**
```typescript
const envelope = {
  mova_version: '3.1',
  intent: { name: 'set-variables', version: '1.0.0' },
  payload: { user: 'Alice', role: 'admin' },
  actions: [
    {
      type: 'set',
      name: 'user-info',
      config: {
        variable: 'user_message',
        value: 'User {{payload.user}} has role {{payload.role}}'
      }
    },
    {
      type: 'print',
      name: 'output',
      config: { value: '{{user_message}}' }
    }
  ]
};

const result = await mova.execute(envelope, true);
```

**Python:**
```python
envelope = MOVAEnvelope(
    mova_version="3.1",
    intent=Intent(name="set-variables", version="1.0.0"),
    payload={"user": "Alice", "role": "admin"},
    actions=[
        Action(
            type="set",
            name="user-info",
            config={
                "variable": "user_message",
                "value": "User {{payload.user}} has role {{payload.role}}"
            }
        ),
        Action(
            type="print",
            name="output",
            config={"value": "{{user_message}}"}
        )
    ]
)

result = client.execute(envelope, wait=True)
```

### 2. HTTP Request with JSON Parsing

**TypeScript:**
```typescript
const httpWorkflow = {
  mova_version: '3.1',
  intent: { name: 'fetch-api-data', version: '1.0.0' },
  payload: { api_url: 'https://jsonplaceholder.typicode.com/users/1' },
  actions: [
    {
      type: 'http_fetch',
      name: 'get-user',
      config: {
        url: '{{payload.api_url}}',
        method: 'GET',
        headers: { 'Accept': 'application/json' }
      }
    },
    {
      type: 'parse_json',
      name: 'extract-name',
      config: { path: 'name' }
    },
    {
      type: 'set',
      name: 'save-result',
      config: {
        variable: 'user_name',
        value: '{{last_result}}'
      }
    }
  ]
};

const result = await mova.execute(httpWorkflow, true);
console.log('User name:', result.variables.user_name);
```

**Python:**
```python
http_workflow = MOVAEnvelope(
    mova_version="3.1",
    intent=Intent(name="fetch-api-data", version="1.0.0"),
    payload={"api_url": "https://jsonplaceholder.typicode.com/users/1"},
    actions=[
        Action(
            type="http_fetch",
            name="get-user",
            config={
                "url": "{{payload.api_url}}",
                "method": "GET",
                "headers": {"Accept": "application/json"}
            }
        ),
        Action(
            type="parse_json",
            name="extract-name",
            config={"path": "name"}
        ),
        Action(
            type="set",
            name="save-result",
            config={
                "variable": "user_name",
                "value": "{{last_result}}"
            }
        )
    ]
)

result = client.execute(http_workflow, wait=True)
print(f"User name: {result.variables['user_name']}")
```

### 3. Conditional Logic

**TypeScript:**
```typescript
const conditionalWorkflow = {
  mova_version: '3.1',
  intent: { name: 'conditional-processing', version: '1.0.0' },
  payload: { user_type: 'premium', credits: 100 },
  actions: [
    {
      type: 'if',
      name: 'check-user-type',
      config: {
        condition: {
          op: 'eq',
          left: '{{payload.user_type}}',
          right: 'premium'
        }
      },
      then: [
        {
          type: 'set',
          name: 'premium-bonus',
          config: {
            variable: 'bonus_credits',
            value: 50
          }
        }
      ],
      else: [
        {
          type: 'set',
          name: 'standard-bonus',
          config: {
            variable: 'bonus_credits',
            value: 10
          }
        }
      ]
    },
    {
      type: 'set',
      name: 'total-credits',
      config: {
        variable: 'total',
        value: '{{payload.credits + bonus_credits}}'
      }
    }
  ]
};

const result = await mova.execute(conditionalWorkflow, true);
```

**Python:**
```python
# Note: Python SDK uses dict for complex action configurations
conditional_workflow = {
    "mova_version": "3.1",
    "intent": {"name": "conditional-processing", "version": "1.0.0"},
    "payload": {"user_type": "premium", "credits": 100},
    "actions": [
        {
            "type": "if",
            "name": "check-user-type",
            "config": {
                "condition": {
                    "op": "eq",
                    "left": "{{payload.user_type}}",
                    "right": "premium"
                }
            },
            "then": [
                {
                    "type": "set",
                    "name": "premium-bonus",
                    "config": {
                        "variable": "bonus_credits",
                        "value": 50
                    }
                }
            ],
            "else": [
                {
                    "type": "set",
                    "name": "standard-bonus",
                    "config": {
                        "variable": "bonus_credits",
                        "value": 10
                    }
                }
            ]
        },
        {
            "type": "set",
            "name": "total-credits",
            "config": {
                "variable": "total",
                "value": "{{payload.credits + bonus_credits}}"
            }
        }
    ]
}

result = client.execute(conditional_workflow, wait=True)
```

### 4. Loop Processing

**TypeScript:**
```typescript
const loopWorkflow = {
  mova_version: '3.1',
  intent: { name: 'process-items', version: '1.0.0' },
  payload: { items: ['apple', 'banana', 'cherry'] },
  actions: [
    {
      type: 'repeat',
      name: 'process-each-item',
      config: {
        for_each: '{{payload.items}}',
        as: 'item',
        index_as: 'i'
      },
      do: [
        {
          type: 'set',
          name: 'process-item',
          config: {
            variable: 'processed_{{i}}',
            value: 'Item {{i}}: {{item}}'
          }
        },
        {
          type: 'print',
          name: 'log-item',
          config: {
            value: 'Processing {{item}} at index {{i}}'
          }
        }
      ]
    }
  ]
};

const result = await mova.execute(loopWorkflow, true);
```

**Python:**
```python
loop_workflow = {
    "mova_version": "3.1",
    "intent": {"name": "process-items", "version": "1.0.0"},
    "payload": {"items": ["apple", "banana", "cherry"]},
    "actions": [
        {
            "type": "repeat",
            "name": "process-each-item",
            "config": {
                "for_each": "{{payload.items}}",
                "as": "item",
                "index_as": "i"
            },
            "do": [
                {
                    "type": "set",
                    "name": "process-item",
                    "config": {
                        "variable": "processed_{{i}}",
                        "value": "Item {{i}}: {{item}}"
                    }
                },
                {
                    "type": "print",
                    "name": "log-item",
                    "config": {
                        "value": "Processing {{item}} at index {{i}}"
                    }
                }
            ]
        }
    ]
}

result = client.execute(loop_workflow, wait=True)
```

## Error Handling

### TypeScript

```typescript
import { MOVAClient } from '@mova-engine/sdk-typescript';

const mova = new MOVAClient();

try {
  const result = await mova.execute(envelope, true);
  console.log('Success:', result);
} catch (error) {
  if (error instanceof Error) {
    console.error('Error:', error.message);
  }
}

// Validation
try {
  const validation = await mova.validate(envelope);
  if (!validation.valid) {
    console.error('Validation errors:', validation.errors);
  }
} catch (error) {
  console.error('Validation request failed:', error);
}
```

### Python

```python
from mova import MOVAClient, MOVAAPIError, MOVAValidationError

client = MOVAClient()

try:
    result = client.execute(envelope, wait=True)
    print(f"Success: {result.status}")
except MOVAValidationError as e:
    print(f"Validation error: {e}")
except MOVAAPIError as e:
    print(f"API error: {e}")
except Exception as e:
    print(f"Unexpected error: {e}")

# Validation
try:
    validation = client.validate(envelope)
    if not validation.valid:
        print(f"Validation errors: {validation.errors}")
except Exception as e:
    print(f"Validation request failed: {e}")
```

## Configuration

### TypeScript

```typescript
import { MOVAClient } from '@mova-engine/sdk-typescript';

const client = new MOVAClient({
  baseURL: 'https://api.mova.example.com',
  timeout: 60000, // 60 seconds
  headers: {
    'Authorization': 'Bearer your-token',
    'X-Custom-Header': 'value'
  }
});
```

### Python

```python
from mova import MOVAClient

client = MOVAClient(
    base_url="https://api.mova.example.com",
    timeout=60.0,  # 60 seconds
    headers={
        "Authorization": "Bearer your-token",
        "X-Custom-Header": "value"
    },
    retry_config={
        "total": 5,
        "backoff_factor": 1.0,
        "status_forcelist": [500, 502, 503, 504]
    }
)
```

## Best Practices

### 1. Use Type Safety (TypeScript)

```typescript
import { MOVAEnvelope, Intent, Action } from '@mova-engine/sdk-typescript';

// Use typed interfaces for better IDE support and error catching
const envelope: MOVAEnvelope = {
  mova_version: '3.1',
  intent: {
    name: 'typed-workflow',
    version: '1.0.0'
  },
  payload: { message: 'Hello' },
  actions: [
    {
      type: 'set',
      name: 'greeting',
      config: { variable: 'msg', value: '{{payload.message}}' }
    }
  ]
};
```

### 2. Use Models (Python)

```python
from mova import MOVAEnvelope, Intent, Action

# Use Pydantic models for validation and IDE support
envelope = MOVAEnvelope(
    mova_version="3.1",
    intent=Intent(name="typed-workflow", version="1.0.0"),
    payload={"message": "Hello"},
    actions=[
        Action(
            type="set",
            name="greeting",
            config={"variable": "msg", "value": "{{payload.message}}"}
        )
    ]
)
```

### 3. Handle Async Operations

```typescript
// TypeScript - polling for completion
async function waitForCompletion(mova: MOVAClient, runId: string): Promise<ExecutionResult> {
  while (true) {
    const status = await mova.getRun(runId);
    if (status.status === 'completed' || status.status === 'failed') {
      return status;
    }
    await new Promise(resolve => setTimeout(resolve, 1000)); // Wait 1 second
  }
}

const asyncResult = await mova.execute(envelope, false);
const finalResult = await waitForCompletion(mova, asyncResult.run_id);
```

```python
# Python - polling for completion
import time

def wait_for_completion(client, run_id):
    while True:
        status = client.get_run(run_id)
        if status.status in ['completed', 'failed']:
            return status
        time.sleep(1)  # Wait 1 second

async_result = client.execute(envelope, wait=False)
final_result = wait_for_completion(client, async_result.run_id)
```

### 4. Use Context Managers (Python)

```python
from mova import MOVAClient

# Automatically close session when done
with MOVAClient() as client:
    result = client.execute(envelope, wait=True)
    logs = client.get_logs(result.run_id)
    print(f"Execution completed with {len(logs)} log entries")
```

## Testing

### TypeScript

```typescript
// Mock for testing
import { MOVAClient } from '@mova-engine/sdk-typescript';

// Mock fetch for testing
global.fetch = jest.fn();

const mockFetch = global.fetch as jest.MockedFunction<typeof fetch>;

beforeEach(() => {
  mockFetch.mockClear();
});

test('should execute workflow', async () => {
  mockFetch.mockResolvedValueOnce({
    ok: true,
    json: async () => ({ run_id: 'test-123', status: 'completed' })
  } as Response);

  const client = new MOVAClient();
  const result = await client.execute(envelope, true);
  
  expect(result.run_id).toBe('test-123');
});
```

### Python

```python
# Using responses library for testing
import responses
from mova import MOVAClient

@responses.activate
def test_execute_workflow():
    responses.add(
        responses.POST,
        "http://localhost:8080/v1/execute",
        json={"run_id": "test-123", "status": "completed"},
        status=200
    )
    
    client = MOVAClient()
    result = client.execute(envelope, wait=True)
    
    assert result.run_id == "test-123"
    assert result.status == "completed"
```

## Troubleshooting

### Common Issues

1. **Connection Refused**: Ensure MOVA API server is running on the specified URL
2. **Timeout Errors**: Increase timeout for long-running workflows
3. **Validation Errors**: Check envelope structure against MOVA v3.1 schema
4. **Authentication Issues**: Verify API tokens and headers

### Debug Mode

**TypeScript:**
```typescript
const client = new MOVAClient({
  baseURL: 'http://localhost:8080',
  headers: { 'X-Debug': 'true' }
});
```

**Python:**
```python
import logging

# Enable debug logging
logging.basicConfig(level=logging.DEBUG)

client = MOVAClient(
    base_url="http://localhost:8080",
    headers={"X-Debug": "true"}
)
```

## Support

- Documentation: [MOVA Engine Docs](https://github.com/mova-engine/mova-engine/tree/main/docs)
- Issues: [GitHub Issues](https://github.com/mova-engine/mova-engine/issues)
- Examples: [SDK Examples](https://github.com/mova-engine/mova-engine/tree/main/examples)
