"""Tests for MOVA Python SDK client."""

from unittest.mock import patch

import pytest
import responses
from mova.client import MOVAClient
from mova.exceptions import (
    MOVAAPIError,
    MOVAConnectionError,
    MOVATimeoutError,
    MOVAValidationError,
)
from mova.models import Action, Intent, MOVAEnvelope


@pytest.fixture
def client():
    """Create test client."""
    return MOVAClient(base_url="http://localhost:8080")


@pytest.fixture
def sample_envelope():
    """Create sample MOVA envelope."""
    return MOVAEnvelope(
        mova_version="3.1",
        intent=Intent(
            name="test-workflow",
            version="1.0.0",
            description="Test workflow",
        ),
        payload={"test": "data"},
        actions=[
            Action(
                type="set",
                name="test-action",
                config={
                    "variable": "test_var",
                    "value": "test_value",
                },
            ),
        ],
    )


class TestMOVAClient:
    """Test MOVAClient class."""

    def test_init_default(self):
        """Test client initialization with defaults."""
        client = MOVAClient()
        assert client.base_url == "http://localhost:8080"
        assert client.timeout == 30.0
        assert "Content-Type" in client.session.headers
        assert client.session.headers["Content-Type"] == "application/json"

    def test_init_custom(self):
        """Test client initialization with custom values."""
        client = MOVAClient(
            base_url="https://api.example.com",
            timeout=60.0,
            headers={"X-Custom": "test"},
        )
        assert client.base_url == "https://api.example.com"
        assert client.timeout == 60.0
        assert client.session.headers["X-Custom"] == "test"

    @responses.activate
    def test_execute_sync(self, client, sample_envelope):
        """Test synchronous workflow execution."""
        mock_result = {
            "run_id": "test-run-123",
            "workflow_id": "test-workflow",
            "status": "completed",
            "start_time": "2024-01-01T00:00:00Z",
            "end_time": "2024-01-01T00:01:00Z",
            "variables": {"test_var": "test_value"},
            "results": {},
            "logs": [],
        }

        responses.add(
            responses.POST,
            "http://localhost:8080/v1/execute",
            json=mock_result,
            status=200,
        )

        result = client.execute(sample_envelope, wait=True)

        assert len(responses.calls) == 1
        assert (
            responses.calls[0].request.url
            == "http://localhost:8080/v1/execute?wait=true"
        )
        assert result.run_id == "test-run-123"
        assert result.status == "completed"

    @responses.activate
    def test_execute_async(self, client, sample_envelope):
        """Test asynchronous workflow execution."""
        mock_result = {
            "run_id": "test-run-123",
            "status": "accepted",
            "message": "Execution started asynchronously",
        }

        responses.add(
            responses.POST,
            "http://localhost:8080/v1/execute",
            json=mock_result,
            status=202,
        )

        result = client.execute(sample_envelope, wait=False)

        assert len(responses.calls) == 1
        assert responses.calls[0].request.url == "http://localhost:8080/v1/execute"
        assert result.run_id == "test-run-123"
        assert result.status == "accepted"

    @responses.activate
    def test_execute_dict_envelope(self, client):
        """Test execution with dict envelope."""
        envelope_dict = {
            "mova_version": "3.1",
            "intent": {"name": "test", "version": "1.0.0"},
            "payload": {"test": "data"},
            "actions": [{"type": "set", "name": "test", "config": {}}],
        }

        mock_result = {
            "run_id": "test-run-123",
            "status": "accepted",
            "message": "Execution started",
        }

        responses.add(
            responses.POST,
            "http://localhost:8080/v1/execute",
            json=mock_result,
            status=202,
        )

        result = client.execute(envelope_dict)
        assert result.run_id == "test-run-123"

    @responses.activate
    def test_execute_error(self, client, sample_envelope):
        """Test execution error handling."""
        responses.add(
            responses.POST,
            "http://localhost:8080/v1/execute",
            json={"error": "Validation failed", "details": "Invalid envelope"},
            status=400,
        )

        with pytest.raises(
            MOVAValidationError, match="Validation failed: Invalid envelope"
        ):
            client.execute(sample_envelope)

    @responses.activate
    def test_validate_success(self, client, sample_envelope):
        """Test successful envelope validation."""
        mock_result = {
            "valid": True,
            "message": "Envelope is valid",
        }

        responses.add(
            responses.POST,
            "http://localhost:8080/v1/validate",
            json=mock_result,
            status=200,
        )

        result = client.validate(sample_envelope)

        assert len(responses.calls) == 1
        assert result.valid is True
        assert result.message == "Envelope is valid"

    @responses.activate
    def test_validate_failure(self, client, sample_envelope):
        """Test envelope validation failure."""
        mock_result = {
            "valid": False,
            "message": "Validation failed",
            "errors": ["Missing required field: intent"],
        }

        responses.add(
            responses.POST,
            "http://localhost:8080/v1/validate",
            json=mock_result,
            status=200,
        )

        result = client.validate(sample_envelope)

        assert result.valid is False
        assert result.message == "Validation failed"
        assert result.errors == ["Missing required field: intent"]

    @responses.activate
    def test_get_run(self, client):
        """Test getting run status."""
        run_id = "test-run-123"
        mock_result = {
            "run_id": run_id,
            "workflow_id": "test-workflow",
            "status": "completed",
            "start_time": "2024-01-01T00:00:00Z",
            "end_time": "2024-01-01T00:01:00Z",
            "variables": {},
            "results": {},
            "logs": [],
        }

        responses.add(
            responses.GET,
            f"http://localhost:8080/v1/runs/{run_id}",
            json=mock_result,
            status=200,
        )

        result = client.get_run(run_id)

        assert len(responses.calls) == 1
        assert result.run_id == run_id
        assert result.status == "completed"

    @responses.activate
    def test_get_run_not_found(self, client):
        """Test getting non-existent run."""
        run_id = "non-existent-run"

        responses.add(
            responses.GET,
            f"http://localhost:8080/v1/runs/{run_id}",
            json={"error": "Run not found"},
            status=404,
        )

        with pytest.raises(MOVAAPIError, match="Run not found"):
            client.get_run(run_id)

    @responses.activate
    def test_get_logs(self, client):
        """Test getting execution logs."""
        run_id = "test-run-123"
        mock_logs = [
            '{"timestamp":"2024-01-01T00:00:00Z","message":"Test log 1"}',
            '{"timestamp":"2024-01-01T00:00:01Z","message":"Test log 2"}',
        ]

        responses.add(
            responses.GET,
            f"http://localhost:8080/v1/runs/{run_id}/logs",
            body="\n".join(mock_logs),
            status=200,
            content_type="application/jsonl",
        )

        result = client.get_logs(run_id)

        assert len(responses.calls) == 1
        assert result == mock_logs

    @responses.activate
    def test_get_logs_empty(self, client):
        """Test getting empty logs."""
        run_id = "test-run-123"

        responses.add(
            responses.GET,
            f"http://localhost:8080/v1/runs/{run_id}/logs",
            body="",
            status=200,
            content_type="application/jsonl",
        )

        result = client.get_logs(run_id)
        assert result == []

    @responses.activate
    def test_get_schemas(self, client):
        """Test getting available schemas."""
        mock_schemas = {
            "schemas": [
                {
                    "name": "envelope",
                    "version": "3.1",
                    "description": "MOVA v3.1 envelope schema",
                    "url": "/v1/schemas/envelope",
                },
            ],
        }

        responses.add(
            responses.GET,
            "http://localhost:8080/v1/schemas",
            json=mock_schemas,
            status=200,
        )

        result = client.get_schemas()

        assert len(responses.calls) == 1
        assert len(result.schemas) == 1
        assert result.schemas[0].name == "envelope"

    @responses.activate
    def test_get_schema(self, client):
        """Test getting specific schema."""
        schema_name = "envelope"
        mock_schema = {
            "$schema": "http://json-schema.org/draft-07/schema#",
            "type": "object",
            "properties": {
                "mova_version": {"type": "string"},
                "intent": {"type": "object"},
                "payload": {"type": "object"},
                "actions": {"type": "array"},
            },
            "required": ["mova_version", "intent", "payload", "actions"],
        }

        responses.add(
            responses.GET,
            f"http://localhost:8080/v1/schemas/{schema_name}",
            json=mock_schema,
            status=200,
        )

        result = client.get_schema(schema_name)

        assert len(responses.calls) == 1
        assert result["type"] == "object"
        assert "mova_version" in result["properties"]

    @responses.activate
    def test_introspect(self, client):
        """Test API introspection."""
        mock_info = {
            "name": "MOVA Automation Engine API",
            "version": "1.0.0",
            "description": "REST API for MOVA workflow execution",
            "mova_version": "3.1",
            "endpoints": [],
            "supported_actions": ["set", "http_fetch", "parse_json", "sleep"],
        }

        responses.add(
            responses.GET,
            "http://localhost:8080/v1/introspect",
            json=mock_info,
            status=200,
        )

        result = client.introspect()

        assert len(responses.calls) == 1
        assert result.name == "MOVA Automation Engine API"
        assert result.mova_version == "3.1"
        assert "set" in result.supported_actions

    @responses.activate
    def test_health(self, client):
        """Test health check."""
        mock_health = {
            "status": "healthy",
            "timestamp": "2024-01-01T00:00:00Z",
            "version": "1.0.0",
        }

        responses.add(
            responses.GET,
            "http://localhost:8080/health",
            json=mock_health,
            status=200,
        )

        result = client.health()

        assert len(responses.calls) == 1
        assert result["status"] == "healthy"

    def test_timeout_error(self, client, sample_envelope):
        """Test timeout error handling."""
        with patch.object(client.session, "request") as mock_request:
            mock_request.side_effect = MOVATimeoutError("Timeout")

            with pytest.raises(MOVATimeoutError):
                client.execute(sample_envelope)

    def test_connection_error(self, client, sample_envelope):
        """Test connection error handling."""
        with patch.object(client.session, "request") as mock_request:
            mock_request.side_effect = MOVAConnectionError("Connection failed")

            with pytest.raises(MOVAConnectionError):
                client.execute(sample_envelope)

    def test_context_manager(self):
        """Test client as context manager."""
        with MOVAClient() as client:
            assert isinstance(client, MOVAClient)
            # Session should be available during context
            assert client.session is not None

    def test_build_url(self, client):
        """Test URL building."""
        assert client._build_url("/test") == "http://localhost:8080/test"
        assert client._build_url("test") == "http://localhost:8080/test"

    @responses.activate
    def test_invalid_json_response(self, client, sample_envelope):
        """Test handling of invalid JSON response."""
        responses.add(
            responses.POST,
            "http://localhost:8080/v1/execute",
            body="invalid json",
            status=200,
        )

        with pytest.raises(MOVAAPIError, match="Invalid JSON response"):
            client.execute(sample_envelope)
