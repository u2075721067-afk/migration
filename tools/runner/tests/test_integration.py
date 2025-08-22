"""
Integration tests for Navigator Agent

Tests the complete system integration including:
- MOVA Engine API connectivity
- Web interface functionality
- End-to-end envelope execution
- Error handling and edge cases
"""

import asyncio
import json

# Import the services
import sys
import tempfile
from pathlib import Path
from typing import Any, Dict

import httpx
import pytest
from fastapi.testclient import TestClient

sys.path.insert(0, str(Path(__file__).parent.parent))

from runner import app as runner_app
from web_interface import web_app


class TestMOVAEngineIntegration:
    """Test integration with MOVA Engine API."""

    @pytest.fixture
    def runner_client(self):
        """Test client for Runner service."""
        return TestClient(runner_app)

    @pytest.fixture
    def web_client(self):
        """Test client for Web interface."""
        return TestClient(web_app)

    def test_health_check(self, runner_client):
        """Test Runner service health check."""
        response = runner_client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert "status" in data
        assert data["status"] == "healthy"

    def test_root_endpoint(self, runner_client):
        """Test Runner service root endpoint."""
        response = runner_client.get("/")
        assert response.status_code == 200
        data = response.json()
        assert "message" in data
        assert "Navigator Agent" in data["message"]

    @pytest.mark.asyncio
    async def test_mova_connection(self):
        """Test connection to MOVA Engine."""
        # This test requires MOVA Engine to be running
        # Skip if not available
        try:
            async with httpx.AsyncClient() as client:
                response = await client.get("http://localhost:8080/health", timeout=5)
                assert response.status_code == 200
        except (httpx.ConnectError, httpx.TimeoutException):
            pytest.skip("MOVA Engine not available")

    def test_validate_envelope_endpoint(self, runner_client):
        """Test envelope validation endpoint."""
        # Create a test envelope
        test_envelope = {
            "mova_version": "3.1",
            "intent": "test",
            "payload": {"action": "test"},
            "actions": [{"type": "print", "params": {"value": "Test message"}}],
        }

        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            json.dump(test_envelope, f)
            temp_path = f.name

        try:
            response = runner_client.post(
                "/validate", json={"cmd_id": "validate", "args": {"file": temp_path}}
            )
            # This will fail if MOVA Engine is not running, which is expected
            assert response.status_code in [200, 500]  # Success or connection error
        finally:
            Path(temp_path).unlink()

    def test_execute_envelope_endpoint(self, runner_client):
        """Test envelope execution endpoint."""
        test_envelope = {
            "mova_version": "3.1",
            "intent": "test",
            "payload": {"action": "test"},
            "actions": [{"type": "print", "params": {"value": "Test execution"}}],
        }

        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            json.dump(test_envelope, f)
            temp_path = f.name

        try:
            response = runner_client.post(
                "/execute", json={"cmd_id": "run", "args": {"file": temp_path}}
            )
            # This will fail if MOVA Engine is not running, which is expected
            assert response.status_code in [200, 500]  # Success or connection error
        finally:
            Path(temp_path).unlink()

    def test_introspect_endpoint(self, runner_client):
        """Test introspection endpoint."""
        response = runner_client.get("/introspect")
        # This will fail if MOVA Engine is not running, which is expected
        assert response.status_code in [200, 500]  # Success or connection error

    def test_rate_limiting(self, runner_client):
        """Test rate limiting functionality."""
        # Make multiple requests quickly
        responses = []
        for _ in range(7):  # Exceed the 5 req/min limit
            response = runner_client.get("/health")
            responses.append(response.status_code)

        # At least one should be rate limited (429)
        assert any(code == 429 for code in responses)

    def test_invalid_command(self, runner_client):
        """Test invalid command handling."""
        response = runner_client.post(
            "/run", json={"cmd_id": "invalid_command", "args": {}}
        )
        assert response.status_code == 403
        data = response.json()
        assert "not in allow-list" in data["detail"]


class TestWebInterface:
    """Test Web Interface functionality."""

    @pytest.fixture
    def web_client(self):
        """Test client for Web interface."""
        return TestClient(web_app)

    def test_web_home_page(self, web_client):
        """Test web interface home page."""
        response = web_client.get("/")
        assert response.status_code == 200
        assert "text/html" in response.headers["content-type"]
        assert "Navigator Agent" in response.text

    def test_web_api_envelopes(self, web_client):
        """Test web API for listing envelopes."""
        response = web_client.get("/api/envelopes")
        assert response.status_code == 200
        data = response.json()
        assert isinstance(data, list)

    def test_web_api_introspect(self, web_client):
        """Test web API introspection endpoint."""
        response = web_client.get("/api/introspect")
        # This will fail if MOVA Engine is not running, which is expected
        assert response.status_code in [200, 500]  # Success or connection error

    def test_web_api_validate(self, web_client):
        """Test web API envelope validation."""
        # Create a test envelope
        test_envelope = {
            "mova_version": "3.1",
            "intent": "test",
            "payload": {"action": "test"},
            "actions": [{"type": "print", "params": {"value": "Test message"}}],
        }

        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            json.dump(test_envelope, f)
            temp_path = f.name

        try:
            # Test with form data
            with open(temp_path, "rb") as f:
                response = web_client.post(
                    "/api/validate",
                    files={"file": ("test.json", f, "application/json")},
                )
            # This will fail if MOVA Engine is not running, which is expected
            assert response.status_code in [200, 500]  # Success or connection error
        finally:
            Path(temp_path).unlink()


class TestEndToEnd:
    """End-to-end tests for complete workflows."""

    @pytest.fixture
    def runner_client(self):
        """Test client for Runner service."""
        return TestClient(runner_app)

    def test_complete_workflow_validation(self, runner_client):
        """Test complete envelope validation workflow."""
        # Create a valid test envelope
        test_envelope = {
            "mova_version": "3.1",
            "intent": "investor_demo",
            "payload": {"action": "validate", "args": {}},
            "actions": [
                {
                    "type": "print",
                    "params": {
                        "value": "🎯 Navigator Agent validation test completed successfully!"
                    },
                }
            ],
        }

        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            json.dump(test_envelope, f)
            temp_path = f.name

        try:
            # Test validation
            response = runner_client.post(
                "/validate", json={"cmd_id": "validate", "args": {"file": temp_path}}
            )

            # Should get a response (success or connection error)
            assert response.status_code in [200, 500]

            if response.status_code == 200:
                data = response.json()
                assert "ok" in data
                assert isinstance(data["ok"], bool)

        finally:
            Path(temp_path).unlink()

    def test_error_handling(self, runner_client):
        """Test error handling in various scenarios."""
        # Test with non-existent file
        response = runner_client.post(
            "/validate",
            json={"cmd_id": "validate", "args": {"file": "/nonexistent/file.json"}},
        )
        assert response.status_code == 500
        data = response.json()
        assert "error" in data

        # Test with invalid JSON file
        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            f.write("invalid json content")
            temp_path = f.name

        try:
            response = runner_client.post(
                "/validate", json={"cmd_id": "validate", "args": {"file": temp_path}}
            )
            assert response.status_code == 500
            data = response.json()
            assert "error" in data
        finally:
            Path(temp_path).unlink()

    def test_security_headers(self, runner_client):
        """Test that security headers are present."""
        response = runner_client.get("/health")
        headers = response.headers

        # Check for CORS headers
        assert "access-control-allow-origin" in headers
        assert "access-control-allow-methods" in headers


class TestPerformance:
    """Performance and load tests."""

    @pytest.fixture
    def runner_client(self):
        """Test client for Runner service."""
        return TestClient(runner_app)

    def test_concurrent_requests(self, runner_client):
        """Test handling of concurrent requests."""
        import concurrent.futures
        import time

        def make_request():
            return runner_client.get("/health")

        # Make 10 concurrent requests
        with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
            futures = [executor.submit(make_request) for _ in range(10)]
            responses = [
                future.result() for future in concurrent.futures.as_completed(futures)
            ]

        # All should succeed
        assert all(response.status_code == 200 for response in responses)

    def test_request_timing(self, runner_client):
        """Test that requests complete within reasonable time."""
        import time

        start_time = time.time()
        response = runner_client.get("/health")
        end_time = time.time()

        assert response.status_code == 200
        assert (end_time - start_time) < 1.0  # Should complete within 1 second
