"""Tests for MOVA SDK models."""

import pytest
from mova.models import Action, ExecutionResult, Intent, MOVAEnvelope, ValidationResult
from pydantic import ValidationError


class TestIntent:
    """Test Intent model."""

    def test_minimal_intent(self):
        """Test intent with minimal required fields."""
        intent = Intent(name="test", version="1.0.0")
        assert intent.name == "test"
        assert intent.version == "1.0.0"
        assert intent.description is None

    def test_full_intent(self):
        """Test intent with all fields."""
        intent = Intent(
            name="test",
            version="1.0.0",
            description="Test intent",
            author="Test Author",
            tags=["test", "example"],
            timeout=300,
        )
        assert intent.name == "test"
        assert intent.description == "Test intent"
        assert intent.tags == ["test", "example"]
        assert intent.timeout == 300

    def test_invalid_timeout(self):
        """Test intent with invalid timeout."""
        with pytest.raises(ValidationError):
            Intent(name="test", version="1.0.0", timeout=-1)


class TestAction:
    """Test Action model."""

    def test_minimal_action(self):
        """Test action with minimal required fields."""
        action = Action(type="set", name="test-action")
        assert action.type == "set"
        assert action.name == "test-action"
        assert action.enabled is True

    def test_full_action(self):
        """Test action with all fields."""
        action = Action(
            type="set",
            name="test-action",
            description="Test action",
            enabled=False,
            timeout=60,
            config={"variable": "test", "value": "test"},
        )
        assert action.type == "set"
        assert action.description == "Test action"
        assert action.enabled is False
        assert action.timeout == 60
        assert action.config == {"variable": "test", "value": "test"}


class TestMOVAEnvelope:
    """Test MOVAEnvelope model."""

    def test_minimal_envelope(self):
        """Test envelope with minimal required fields."""
        envelope = MOVAEnvelope(
            mova_version="3.1",
            intent=Intent(name="test", version="1.0.0"),
            payload={"test": "data"},
            actions=[Action(type="set", name="test")],
        )
        assert envelope.mova_version == "3.1"
        assert envelope.intent.name == "test"
        assert envelope.payload == {"test": "data"}
        assert len(envelope.actions) == 1

    def test_full_envelope(self):
        """Test envelope with all fields."""
        envelope = MOVAEnvelope(
            mova_version="3.1",
            intent=Intent(name="test", version="1.0.0"),
            payload={"test": "data"},
            actions=[Action(type="set", name="test")],
            variables={"var1": "value1"},
            secrets={"secret1": "value1"},
        )
        assert envelope.variables == {"var1": "value1"}
        assert envelope.secrets == {"secret1": "value1"}

    def test_envelope_serialization(self):
        """Test envelope serialization."""
        envelope = MOVAEnvelope(
            mova_version="3.1",
            intent=Intent(name="test", version="1.0.0"),
            payload={"test": "data"},
            actions=[Action(type="set", name="test")],
        )

        data = envelope.model_dump()
        assert data["mova_version"] == "3.1"
        assert data["intent"]["name"] == "test"
        assert data["payload"]["test"] == "data"

    def test_envelope_from_dict(self):
        """Test creating envelope from dictionary."""
        data = {
            "mova_version": "3.1",
            "intent": {"name": "test", "version": "1.0.0"},
            "payload": {"test": "data"},
            "actions": [{"type": "set", "name": "test"}],
        }

        envelope = MOVAEnvelope(**data)
        assert envelope.mova_version == "3.1"
        assert envelope.intent.name == "test"


class TestExecutionResult:
    """Test ExecutionResult model."""

    def test_minimal_execution_result(self):
        """Test execution result with minimal fields."""
        result = ExecutionResult(
            run_id="test-123",
            workflow_id="test-workflow",
            start_time="2024-01-01T00:00:00Z",
            status="completed",
            variables={},
            results={},
            logs=[],
        )
        assert result.run_id == "test-123"
        assert result.status == "completed"
        assert result.end_time is None

    def test_full_execution_result(self):
        """Test execution result with all fields."""
        result = ExecutionResult(
            run_id="test-123",
            workflow_id="test-workflow",
            start_time="2024-01-01T00:00:00Z",
            end_time="2024-01-01T00:01:00Z",
            status="completed",
            variables={"var1": "value1"},
            results={"action1": {"status": "completed"}},
            logs=[
                {
                    "timestamp": "2024-01-01T00:00:00Z",
                    "level": "info",
                    "step": "action",
                    "type": "set",
                    "message": "Variable set",
                    "status": "success",
                }
            ],
        )
        assert result.end_time == "2024-01-01T00:01:00Z"
        assert result.variables == {"var1": "value1"}
        assert len(result.logs) == 1

    def test_invalid_status(self):
        """Test execution result with invalid status."""
        with pytest.raises(ValidationError):
            ExecutionResult(
                run_id="test-123",
                workflow_id="test-workflow",
                start_time="2024-01-01T00:00:00Z",
                status="invalid_status",
                variables={},
                results={},
                logs=[],
            )


class TestValidationResult:
    """Test ValidationResult model."""

    def test_valid_result(self):
        """Test valid validation result."""
        result = ValidationResult(valid=True, message="Envelope is valid")
        assert result.valid is True
        assert result.message == "Envelope is valid"
        assert result.errors is None

    def test_invalid_result(self):
        """Test invalid validation result."""
        result = ValidationResult(
            valid=False,
            message="Validation failed",
            errors=["Missing field: intent", "Invalid action type"],
        )
        assert result.valid is False
        assert result.message == "Validation failed"
        assert len(result.errors) == 2
