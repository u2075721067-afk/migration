# MOVA Python SDK

Python SDK for the MOVA Automation Engine.

## Installation

```bash
pip install mova-engine-sdk
```

## Quick Start

```python
from mova import MOVAClient, MOVAEnvelope, Intent, Action

# Create client
client = MOVAClient(
    base_url="http://localhost:8080",  # Optional, defaults to localhost:8080
    timeout=30.0,  # Optional, defaults to 30s
)

# Define a workflow envelope
envelope = MOVAEnvelope(
    mova_version="3.1",
    intent=Intent(
        name="my-workflow",
        version="1.0.0",
        description="Example workflow",
    ),
    payload={
        "message": "Hello, MOVA!",
    },
    actions=[
        Action(
            type="set",
            name="greeting",
            config={
                "variable": "greeting",
                "value": "{{payload.message}}",
            },
        ),
        Action(
            type="print",
            name="print-greeting",
            config={
                "value": "{{greeting}}",
            },
        ),
    ],
)

# Execute workflow
try:
    # Synchronous execution (wait for completion)
    result = client.execute(envelope, wait=True)
    print(f"Workflow completed: {result.status}")
    print(f"Variables: {result.variables}")

    # Asynchronous execution
    async_result = client.execute(envelope, wait=False)
    print(f"Workflow started: {async_result.run_id}")
    
    # Check status later
    status = client.get_run(async_result.run_id)
    print(f"Current status: {status.status}")
    
    # Get logs
    logs = client.get_logs(async_result.run_id)
    for log in logs:
        print(log)
        
except Exception as error:
    print(f"Workflow execution failed: {error}")
```

## API Reference

### MOVAClient

#### Constructor

```python
MOVAClient(
    base_url: str = "http://localhost:8080",
    timeout: float = 30.0,
    headers: Optional[Dict[str, str]] = None,
    retry_config: Optional[Dict[str, int]] = None,
)
```

**Parameters:**
- `base_url: str` - Base URL of the MOVA API server
- `timeout: float` - Request timeout in seconds
- `headers: Dict[str, str]` - Additional headers to send with requests
- `retry_config: Dict[str, int]` - Retry configuration (total, backoff_factor, status_forcelist)

#### Methods

##### execute(envelope, wait=False)

Execute a MOVA workflow envelope.

```python
def execute(
    envelope: Union[MOVAEnvelope, Dict], 
    wait: bool = False
) -> Union[ExecutionResult, AsyncExecutionResult]
```

**Parameters:**
- `envelope: Union[MOVAEnvelope, Dict]` - The workflow envelope to execute
- `wait: bool` - Whether to wait for execution to complete

**Returns:** Execution result (sync) or async execution info

##### validate(envelope)

Validate a MOVA envelope against the schema.

```python
def validate(envelope: Union[MOVAEnvelope, Dict]) -> ValidationResult
```

**Parameters:**
- `envelope: Union[MOVAEnvelope, Dict]` - The envelope to validate

**Returns:** Validation result

##### get_run(run_id)

Get the status and result of a workflow execution.

```python
def get_run(run_id: str) -> ExecutionResult
```

**Parameters:**
- `run_id: str` - The run ID to retrieve

**Returns:** Execution result

##### get_logs(run_id)

Get the logs for a workflow execution.

```python
def get_logs(run_id: str) -> List[str]
```

**Parameters:**
- `run_id: str` - The run ID to retrieve logs for

**Returns:** List of JSONL log entries

##### get_schemas()

Get available schemas.

```python
def get_schemas() -> SchemasResponse
```

**Returns:** Schemas information

##### get_schema(name)

Get a specific schema by name.

```python
def get_schema(name: str) -> Dict
```

**Parameters:**
- `name: str` - Schema name (e.g., 'envelope', 'action')

**Returns:** Schema definition

##### introspect()

Get API introspection information.

```python
def introspect() -> IntrospectionResult
```

**Returns:** API information

##### health()

Check API health status.

```python
def health() -> Dict
```

**Returns:** Health status information

## Models

### MOVAEnvelope

```python
class MOVAEnvelope(BaseModel):
    mova_version: str
    intent: Intent
    payload: Dict[str, Any]
    actions: List[Action]
    variables: Optional[Dict[str, Any]] = None
    secrets: Optional[Dict[str, str]] = None
```

### Intent

```python
class Intent(BaseModel):
    name: str
    version: str
    description: Optional[str] = None
    author: Optional[str] = None
    tags: Optional[List[str]] = None
    timeout: Optional[int] = None
    retry: Optional[RetryPolicy] = None
    budget: Optional[BudgetConstraints] = None
```

### Action

```python
class Action(BaseModel):
    type: str
    name: str
    description: Optional[str] = None
    enabled: Optional[bool] = True
    timeout: Optional[int] = None
    retry: Optional[RetryPolicy] = None
    config: Optional[Dict[str, Any]] = None
```

### ExecutionResult

```python
class ExecutionResult(BaseModel):
    run_id: str
    workflow_id: str
    start_time: str
    end_time: Optional[str] = None
    status: Literal["pending", "running", "completed", "failed", "cancelled"]
    variables: Dict[str, Any]
    results: Dict[str, Any]
    logs: List[ExecutionLog]
```

### ValidationResult

```python
class ValidationResult(BaseModel):
    valid: bool
    message: str
    errors: Optional[List[str]] = None
```

## Examples

### Basic Workflow

```python
from mova import mova, MOVAEnvelope, Intent, Action

envelope = MOVAEnvelope(
    mova_version="3.1",
    intent=Intent(name="hello-world", version="1.0.0"),
    payload={"name": "World"},
    actions=[
        Action(
            type="set",
            name="greeting",
            config={"variable": "message", "value": "Hello, {{payload.name}}!"}
        ),
        Action(
            type="print",
            name="output",
            config={"value": "{{message}}"}
        )
    ]
)

result = mova.execute(envelope, wait=True)
print(result.variables["message"])  # "Hello, World!"
```

### HTTP Request Workflow

```python
http_workflow = MOVAEnvelope(
    mova_version="3.1",
    intent=Intent(name="fetch-data", version="1.0.0"),
    payload={"url": "https://api.example.com/data"},
    actions=[
        Action(
            type="http_fetch",
            name="fetch",
            config={
                "url": "{{payload.url}}",
                "method": "GET",
                "headers": {"Accept": "application/json"}
            }
        ),
        Action(
            type="parse_json",
            name="extract",
            config={"path": "data.items[0].name"}
        )
    ]
)

result = mova.execute(http_workflow, wait=True)
```

### Using Dictionary Instead of Models

```python
from mova import mova

# You can also use plain dictionaries
envelope_dict = {
    "mova_version": "3.1",
    "intent": {"name": "test", "version": "1.0.0"},
    "payload": {"message": "Hello!"},
    "actions": [
        {
            "type": "set",
            "name": "greeting",
            "config": {"variable": "msg", "value": "{{payload.message}}"}
        }
    ]
}

result = mova.execute(envelope_dict, wait=True)
```

### Error Handling

```python
from mova import mova, MOVAAPIError, MOVAValidationError

try:
    result = mova.execute(envelope, wait=True)
    print(f"Success: {result.status}")
except MOVAValidationError as e:
    print(f"Validation error: {e}")
except MOVAAPIError as e:
    print(f"API error: {e}")
except Exception as e:
    print(f"Unexpected error: {e}")
```

### Validation

```python
validation = mova.validate(envelope)
if validation.valid:
    print("Envelope is valid")
else:
    print(f"Validation errors: {validation.errors}")
```

### Context Manager

```python
from mova import MOVAClient

with MOVAClient() as client:
    result = client.execute(envelope, wait=True)
    print(f"Result: {result.status}")
# Client session is automatically closed
```

### Custom Configuration

```python
from mova import MOVAClient

# Custom client with retry configuration
client = MOVAClient(
    base_url="https://api.mova.example.com",
    timeout=60.0,
    headers={"Authorization": "Bearer your-token"},
    retry_config={
        "total": 5,
        "backoff_factor": 1.0,
        "status_forcelist": [500, 502, 503, 504],
    }
)

result = client.execute(envelope, wait=True)
```

## Development

```bash
# Install in development mode
pip install -e ".[dev]"

# Run tests
pytest

# Run tests with coverage
pytest --cov=src/mova --cov-report=html

# Format code
black src tests

# Sort imports
isort src tests

# Lint code
flake8 src tests

# Type checking
mypy src
```

## License

MIT
