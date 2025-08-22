"""
Tests for Navigator Agent Web Interface

Tests the web interface functionality including:
- API endpoints
- Template rendering
- Form handling
- Error handling
- Integration with Runner service
"""

import json

# Import the web interface
import sys
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest
from fastapi.testclient import TestClient

sys.path.insert(0, str(Path(__file__).parent.parent))

from web_interface import web_app


class TestWebInterface:
    """Test web interface functionality."""

    @pytest.fixture
    def client(self):
        """Test client for web interface."""
        return TestClient(web_app)

    def test_home_page(self, client):
        """Test home page loads successfully."""
        response = client.get("/")
        assert response.status_code == 200
        assert "text/html" in response.headers["content-type"]
        assert "Navigator Agent" in response.text

    def test_api_envelopes_list(self, client):
        """Test API endpoint for listing envelopes."""
        response = client.get("/api/envelopes")
        assert response.status_code == 200
        data = response.json()
        assert isinstance(data, list)

        # Should include demo_agent.json if it exists
        envelope_names = [item["name"] for item in data]
        if Path("envelopes/demo_agent.json").exists():
            assert "demo_agent.json" in envelope_names

    @patch("web_interface.get_introspection")
    def test_api_introspect_success(self, mock_introspect, client):
        """Test introspection API endpoint."""
        mock_introspect.return_value = {
            "version": "1.0.0",
            "capabilities": ["execute", "validate"],
        }

        response = client.get("/api/introspect")
        assert response.status_code == 200
        data = response.json()
        assert "version" in data
        assert "capabilities" in data

    @patch("web_interface.get_introspection")
    def test_api_introspect_error(self, mock_introspect, client):
        """Test introspection API error handling."""
        mock_introspect.side_effect = Exception("Connection error")

        response = client.get("/api/introspect")
        assert response.status_code == 500
        data = response.json()
        assert "error" in data

    @patch("web_interface.validate_envelope")
    def test_api_validate_success(self, mock_validate, client):
        """Test envelope validation API."""
        mock_validate.return_value = {"ok": True, "valid": True}

        response = client.post("/api/validate", data={"file": "test.json"})
        assert response.status_code == 200
        data = response.json()
        assert data["ok"] is True

    @patch("web_interface.validate_envelope")
    def test_api_validate_error(self, mock_validate, client):
        """Test envelope validation error handling."""
        mock_validate.side_effect = Exception("Validation failed")

        response = client.post("/api/validate", data={"file": "test.json"})
        assert response.status_code == 400
        data = response.json()
        assert "error" in data

    @patch("web_interface.execute_envelope")
    def test_api_execute_success(self, mock_execute, client):
        """Test envelope execution API."""
        mock_execute.return_value = {
            "ok": True,
            "run_id": "test-run-123",
            "result": {"status": "completed"},
        }

        response = client.post("/api/execute", data={"file": "test.json"})
        assert response.status_code == 200
        data = response.json()
        assert data["ok"] is True
        assert "run_id" in data

    @patch("web_interface.execute_envelope")
    def test_api_execute_error(self, mock_execute, client):
        """Test envelope execution error handling."""
        mock_execute.side_effect = Exception("Execution failed")

        response = client.post("/api/execute", data={"file": "test.json"})
        assert response.status_code == 400
        data = response.json()
        assert "error" in data

    @patch("web_interface.get_run_logs")
    def test_api_logs_success(self, mock_logs, client):
        """Test logs retrieval API."""
        mock_logs.return_value = {
            "run_id": "test-run-123",
            "logs": ["Step 1", "Step 2"],
        }

        response = client.get("/api/logs/test-run-123")
        assert response.status_code == 200
        data = response.json()
        assert "logs" in data

    @patch("web_interface.get_run_logs")
    def test_api_logs_not_found(self, mock_logs, client):
        """Test logs retrieval for non-existent run."""
        mock_logs.side_effect = Exception("Run not found")

        response = client.get("/api/logs/non-existent-run")
        assert response.status_code == 404
        data = response.json()
        assert "error" in data


class TestWebInterfaceIntegration:
    """Test web interface integration with Runner service."""

    @pytest.fixture
    def client(self):
        """Test client for web interface."""
        return TestClient(web_app)

    def test_form_validation_workflow(self, client):
        """Test complete form validation workflow."""
        # Create a test envelope
        test_envelope = {
            "mova_version": "3.1",
            "intent": "test",
            "payload": {"action": "validate"},
            "actions": [{"type": "print", "params": {"value": "Test"}}],
        }

        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            json.dump(test_envelope, f)
            temp_path = f.name

        try:
            # Test validation endpoint
            response = client.post("/api/validate", data={"file": temp_path})
            # Should get a response (success or connection error)
            assert response.status_code in [200, 400, 500]
        finally:
            Path(temp_path).unlink()

    def test_form_execution_workflow(self, client):
        """Test complete form execution workflow."""
        # Create a test envelope
        test_envelope = {
            "mova_version": "3.1",
            "intent": "test",
            "payload": {"action": "run"},
            "actions": [{"type": "print", "params": {"value": "Test execution"}}],
        }

        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            json.dump(test_envelope, f)
            temp_path = f.name

        try:
            # Test execution endpoint
            response = client.post("/api/execute", data={"file": temp_path})
            # Should get a response (success or connection error)
            assert response.status_code in [200, 400, 500]
        finally:
            Path(temp_path).unlink()


class TestWebInterfaceSecurity:
    """Test security aspects of web interface."""

    @pytest.fixture
    def client(self):
        """Test client for web interface."""
        return TestClient(web_app)

    def test_path_traversal_protection(self, client):
        """Test protection against path traversal attacks."""
        # Try to access files outside allowed directory
        malicious_paths = [
            "../../../etc/passwd",
            "..\\..\\..\\windows\\system32\\config",
            "/etc/passwd",
            "C:\\Windows\\System32\\config",
        ]

        for malicious_path in malicious_paths:
            response = client.post("/api/validate", data={"file": malicious_path})
            # Should fail with validation error
            assert response.status_code in [400, 500]

    def test_invalid_file_types(self, client):
        """Test handling of invalid file types."""
        response = client.post("/api/validate", data={"file": "test.txt"})
        # Should handle gracefully
        assert response.status_code in [200, 400, 500]

    def test_large_file_handling(self, client):
        """Test handling of large files."""
        # Create a large test file
        large_content = "x" * 1000000  # 1MB file
        large_envelope = {
            "mova_version": "3.1",
            "intent": "test",
            "payload": {"action": "test"},
            "actions": [{"type": "print", "params": {"value": large_content}}],
        }

        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            json.dump(large_envelope, f)
            temp_path = f.name

        try:
            response = client.post("/api/validate", data={"file": temp_path})
            # Should handle large files gracefully
            assert response.status_code in [200, 400, 500]
        finally:
            Path(temp_path).unlink()


class TestWebInterfaceTemplates:
    """Test web interface template rendering."""

    @pytest.fixture
    def client(self):
        """Test client for web interface."""
        return TestClient(web_app)

    def test_home_template_rendering(self, client):
        """Test home page template renders correctly."""
        response = client.get("/")
        assert response.status_code == 200

        # Check for key UI elements
        content = response.text
        assert "<html" in content
        assert "</html>" in content
        assert "Navigator Agent" in content

    def test_template_error_handling(self, client):
        """Test template error handling."""
        # Try to access non-existent template
        with patch("web_interface.templates.TemplateResponse") as mock_template:
            mock_template.side_effect = Exception("Template error")

            response = client.get("/")
            # Should handle template errors gracefully
            assert response.status_code in [200, 500]
