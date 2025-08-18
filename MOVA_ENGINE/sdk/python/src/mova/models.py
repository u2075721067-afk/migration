"""MOVA SDK data models."""

from typing import Any, Dict, List, Literal, Optional

from pydantic import BaseModel, Field


class RetryPolicy(BaseModel):
    """Retry policy configuration."""

    count: int = Field(ge=0, description="Number of retries")
    backoff_ms: int = Field(ge=0, description="Backoff time in milliseconds")


class BudgetConstraints(BaseModel):
    """Budget constraints configuration."""

    tokens: Optional[int] = Field(None, ge=0, description="Maximum tokens")
    cost_usd: Optional[float] = Field(None, ge=0, description="Maximum cost in USD")


class Intent(BaseModel):
    """Workflow intent definition."""

    name: str = Field(description="Intent name")
    version: str = Field(description="Intent version")
    description: Optional[str] = Field(None, description="Intent description")
    author: Optional[str] = Field(None, description="Intent author")
    tags: Optional[List[str]] = Field(None, description="Intent tags")
    timeout: Optional[int] = Field(None, ge=0, description="Timeout in seconds")
    retry: Optional[RetryPolicy] = Field(None, description="Retry policy")
    budget: Optional[BudgetConstraints] = Field(None, description="Budget constraints")


class Action(BaseModel):
    """Workflow action definition."""

    type: str = Field(description="Action type")
    name: str = Field(description="Action name")
    description: Optional[str] = Field(None, description="Action description")
    enabled: Optional[bool] = Field(True, description="Whether action is enabled")
    timeout: Optional[int] = Field(None, ge=0, description="Action timeout in seconds")
    retry: Optional[RetryPolicy] = Field(None, description="Action retry policy")
    config: Optional[Dict[str, Any]] = Field(None, description="Action configuration")


class MOVAEnvelope(BaseModel):
    """MOVA workflow envelope."""

    mova_version: str = Field(description="MOVA version")
    intent: Intent = Field(description="Workflow intent")
    payload: Dict[str, Any] = Field(description="Workflow payload")
    actions: List[Action] = Field(description="Workflow actions")
    variables: Optional[Dict[str, Any]] = Field(None, description="Workflow variables")
    secrets: Optional[Dict[str, str]] = Field(None, description="Workflow secrets")


class ExecutionLog(BaseModel):
    """Execution log entry."""

    timestamp: str = Field(description="Log timestamp")
    level: str = Field(description="Log level")
    step: str = Field(description="Execution step")
    type: str = Field(description="Log type")
    action: Optional[str] = Field(None, description="Action name")
    message: str = Field(description="Log message")
    params_redacted: Optional[Dict[str, Any]] = Field(
        None, description="Redacted parameters"
    )
    status: str = Field(description="Status")
    data: Optional[Dict[str, Any]] = Field(None, description="Additional data")


class ExecutionResult(BaseModel):
    """Workflow execution result."""

    run_id: str = Field(description="Execution run ID")
    workflow_id: str = Field(description="Workflow ID")
    start_time: str = Field(description="Execution start time")
    end_time: Optional[str] = Field(None, description="Execution end time")
    status: Literal["pending", "running", "completed", "failed", "cancelled"] = Field(
        description="Execution status"
    )
    variables: Dict[str, Any] = Field(description="Execution variables")
    results: Dict[str, Any] = Field(description="Execution results")
    logs: List[ExecutionLog] = Field(description="Execution logs")


class ValidationResult(BaseModel):
    """Envelope validation result."""

    valid: bool = Field(description="Whether envelope is valid")
    message: str = Field(description="Validation message")
    errors: Optional[List[str]] = Field(None, description="Validation errors")


class AsyncExecutionResult(BaseModel):
    """Asynchronous execution result."""

    run_id: str = Field(description="Execution run ID")
    status: str = Field(description="Execution status")
    message: str = Field(description="Status message")


class SchemaInfo(BaseModel):
    """Schema information."""

    name: str = Field(description="Schema name")
    version: str = Field(description="Schema version")
    description: str = Field(description="Schema description")
    url: str = Field(description="Schema URL")


class SchemasResponse(BaseModel):
    """Schemas list response."""

    schemas: List[SchemaInfo] = Field(description="Available schemas")


class EndpointInfo(BaseModel):
    """API endpoint information."""

    method: str = Field(description="HTTP method")
    path: str = Field(description="Endpoint path")
    description: str = Field(description="Endpoint description")
    query_params: Optional[List[str]] = Field(None, description="Query parameters")


class IntrospectionResult(BaseModel):
    """API introspection result."""

    name: str = Field(description="API name")
    version: str = Field(description="API version")
    description: str = Field(description="API description")
    mova_version: str = Field(description="MOVA version")
    endpoints: List[EndpointInfo] = Field(description="Available endpoints")
    supported_actions: List[str] = Field(description="Supported actions")
