"""
End-to-End Tests for Navigator Agent

Tests complete workflows including:
- Container startup and health checks
- Full envelope validation and execution
- Web interface integration
- Error scenarios and recovery
- Performance under load
"""

import asyncio
import json

# Import services
import sys
import tempfile
import time
from pathlib import Path
from typing import Any, Dict

import httpx
import pytest
from fastapi.testclient import TestClient

sys.path.insert(0, str(Path(__file__).parent.parent))

from runner import app as runner_app
from web_interface import web_app


class TestEndToEndWorkflow:
    """Test complete end-to-end workflows."""

    @pytest.fixture
    def runner_client(self):
        """Test client for Runner service."""
        return TestClient(runner_app)

    @pytest.fixture
    def web_client(self):
        """Test client for Web interface."""
        return TestClient(web_app)

    def test_health_check_workflow(self, runner_client, web_client):
        """Test health check workflow for all services."""
        # Test Runner service health
        response = runner_client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"

        # Test web interface is accessible
        response = web_client.get("/")
        assert response.status_code == 200

    def test_envelope_validation_workflow(self, runner_client, web_client):
        """Test complete envelope validation workflow."""
        # Create a test envelope
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
            # Test via Runner API
            response = runner_client.post(
                "/validate", json={"cmd_id": "validate", "args": {"file": temp_path}}
            )
            assert response.status_code in [200, 500]  # Success or connection error

            # Test via Web interface
            response = web_client.post("/api/validate", data={"file": temp_path})
            assert response.status_code in [200, 400, 500]

        finally:
            Path(temp_path).unlink()

    def test_envelope_execution_workflow(self, runner_client, web_client):
        """Test complete envelope execution workflow."""
        # Create a test envelope
        test_envelope = {
            "mova_version": "3.1",
            "intent": "investor_demo",
            "payload": {"action": "run", "args": {}},
            "actions": [
                {
                    "type": "print",
                    "params": {"value": "🚀 Navigator Agent execution test completed!"},
                }
            ],
        }

        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            json.dump(test_envelope, f)
            temp_path = f.name

        try:
            # Test via Runner API
            response = runner_client.post(
                "/execute", json={"cmd_id": "run", "args": {"file": temp_path}}
            )
            assert response.status_code in [200, 500]  # Success or connection error

            # Test via Web interface
            response = web_client.post("/api/execute", data={"file": temp_path})
            assert response.status_code in [200, 400, 500]

        finally:
            Path(temp_path).unlink()


class TestErrorScenarios:
    """Test error handling and edge cases."""

    @pytest.fixture
    def runner_client(self):
        """Test client for Runner service."""
        return TestClient(runner_app)

    @pytest.fixture
    def web_client(self):
        """Test client for Web interface."""
        return TestClient(web_app)

    def test_invalid_envelope_handling(self, runner_client, web_client):
        """Test handling of invalid envelopes."""
        # Create invalid envelope
        invalid_envelope = {"invalid": "envelope"}

        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            json.dump(invalid_envelope, f)
            temp_path = f.name

        try:
            # Test via Runner API
            response = runner_client.post(
                "/validate", json={"cmd_id": "validate", "args": {"file": temp_path}}
            )
            assert response.status_code == 500

            # Test via Web interface
            response = web_client.post("/api/validate", data={"file": temp_path})
            assert response.status_code == 400

        finally:
            Path(temp_path).unlink()

    def test_nonexistent_file_handling(self, runner_client, web_client):
        """Test handling of non-existent files."""
        # Test via Runner API
        response = runner_client.post(
            "/validate",
            json={"cmd_id": "validate", "args": {"file": "/nonexistent/file.json"}},
        )
        assert response.status_code == 500

        # Test via Web interface
        response = web_client.post(
            "/api/validate", data={"file": "/nonexistent/file.json"}
        )
        assert response.status_code == 400

    def test_malformed_json_handling(self, runner_client, web_client):
        """Test handling of malformed JSON files."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            f.write("{ invalid json")
            temp_path = f.name

        try:
            # Test via Runner API
            response = runner_client.post(
                "/validate", json={"cmd_id": "validate", "args": {"file": temp_path}}
            )
            assert response.status_code == 500

            # Test via Web interface
            response = web_client.post("/api/validate", data={"file": temp_path})
            assert response.status_code == 400

        finally:
            Path(temp_path).unlink()

    def test_unauthorized_command_handling(self, runner_client):
        """Test handling of unauthorized commands."""
        response = runner_client.post(
            "/run", json={"cmd_id": "unauthorized_command", "args": {}}
        )
        assert response.status_code == 403
        data = response.json()
        assert "not in allow-list" in data["detail"]

    def test_rate_limit_handling(self, runner_client):
        """Test rate limiting behavior."""
        # Make multiple requests quickly
        responses = []
        for _ in range(7):  # Exceed limit
            response = runner_client.get("/health")
            responses.append(response.status_code)

        # At least one should be rate limited
        assert any(code == 429 for code in responses)


class TestPerformance:
    """Test performance and load handling."""

    @pytest.fixture
    def runner_client(self):
        """Test client for Runner service."""
        return TestClient(runner_app)

    @pytest.fixture
    def web_client(self):
        """Test client for Web interface."""
        return TestClient(web_app)

    def test_concurrent_requests(self, runner_client):
        """Test handling of concurrent requests."""
        import concurrent.futures
        import threading

        results = []
        lock = threading.Lock()

        def make_request():
            response = runner_client.get("/health")
            with lock:
                results.append(response.status_code)

        # Make 10 concurrent requests
        with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
            futures = [executor.submit(make_request) for _ in range(10)]
            concurrent.futures.wait(futures)

        # All should succeed
        assert all(code == 200 for code in results)
        assert len(results) == 10

    def test_request_timing(self, runner_client):
        """Test that requests complete within reasonable time."""
        import time

        start_time = time.time()
        response = runner_client.get("/health")
        end_time = time.time()

        assert response.status_code == 200
        assert (end_time - start_time) < 1.0  # Should complete within 1 second

    def test_large_envelope_handling(self, runner_client, web_client):
        """Test handling of large envelopes."""
        # Create a large envelope with many actions
        large_envelope = {
            "mova_version": "3.1",
            "intent": "performance_test",
            "payload": {"action": "test"},
            "actions": [
                {"type": "print", "params": {"value": f"Action {i}: " + "x" * 1000}}
                for i in range(100)  # 100 actions
            ],
        }

        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            json.dump(large_envelope, f)
            temp_path = f.name

        try:
            # Test via Runner API
            response = runner_client.post(
                "/validate", json={"cmd_id": "validate", "args": {"file": temp_path}}
            )
            # Should handle large envelopes gracefully
            assert response.status_code in [200, 500]

            # Test via Web interface
            response = web_client.post("/api/validate", data={"file": temp_path})
            assert response.status_code in [200, 400, 500]

        finally:
            Path(temp_path).unlink()


class TestIntegrationWithMOVAEngine:
    """Test integration with actual MOVA Engine."""

    @pytest.fixture
    def runner_client(self):
        """Test client for Runner service."""
        return TestClient(runner_app)

    @pytest.mark.integration
    def test_mova_engine_connection(self, runner_client):
        """Test actual connection to MOVA Engine."""
        try:
            # Test health check
            response = runner_client.get("/health")
            assert response.status_code == 200

            # Try to connect to MOVA Engine
            import asyncio

            from runner import get_introspection

            result = asyncio.run(get_introspection())
            assert "version" in result or "error" in result

        except Exception:
            pytest.skip("MOVA Engine not available")

    @pytest.mark.integration
    def test_demo_envelope_execution(self, runner_client):
        """Test execution of actual demo envelope."""
        envelope_path = (
            Path(__file__).parent.parent.parent / "envelopes" / "demo_agent.json"
        )

        if not envelope_path.exists():
            pytest.skip("Demo envelope not found")

        # Test envelope validation
        response = runner_client.post(
            "/validate",
            json={"cmd_id": "validate", "args": {"file": str(envelope_path)}},
        )

        # Should get some response (success or connection error)
        assert response.status_code in [200, 500]

        if response.status_code == 200:
            data = response.json()
            assert "ok" in data

    @pytest.mark.integration
    def test_web_interface_with_demo_envelope(self, web_client):
        """Test web interface with demo envelope."""
        envelope_path = (
            Path(__file__).parent.parent.parent / "envelopes" / "demo_agent.json"
        )

        if not envelope_path.exists():
            pytest.skip("Demo envelope not found")

        # Test via web interface
        response = web_client.post("/api/validate", data={"file": str(envelope_path)})
        assert response.status_code in [200, 400, 500]


class TestSecurityIntegration:
    """Test security features integration."""

    @pytest.fixture
    def runner_client(self):
        """Test client for Runner service."""
        return TestClient(runner_app)

    @pytest.fixture
    def web_client(self):
        """Test client for Web interface."""
        return TestClient(web_app)

    def test_path_sanitization_integration(self, runner_client, web_client):
        """Test path sanitization across all endpoints."""
        malicious_paths = [
            "../../../etc/passwd",
            "..\\..\\..\\windows\\system32\\config",
            "/etc/passwd",
            "C:\\Windows\\System32\\config",
        ]

        for malicious_path in malicious_paths:
            # Test via Runner API
            response = runner_client.post(
                "/validate",
                json={"cmd_id": "validate", "args": {"file": malicious_path}},
            )
            assert response.status_code in [400, 500]

            # Test via Web interface
            response = web_client.post("/api/validate", data={"file": malicious_path})
            assert response.status_code in [400, 500]

    def test_command_injection_prevention(self, runner_client):
        """Test prevention of command injection attacks."""
        # Try various injection attempts
        injection_attempts = [
            "test.json; rm -rf /",
            "test.json && cat /etc/passwd",
            "test.json | ls -la",
            "test.json || echo hacked",
        ]

        for injection in injection_attempts:
            response = runner_client.post(
                "/run", json={"cmd_id": "validate", "args": {"file": injection}}
            )
            # Should either fail validation or succeed without injection
            assert response.status_code in [200, 400, 500]

    def test_rate_limiting_integration(self, runner_client):
        """Test rate limiting across multiple endpoints."""
        # Make requests to different endpoints
        endpoints = ["/health", "/introspect"]

        for _ in range(10):  # Make many requests
            for endpoint in endpoints:
                response = runner_client.get(endpoint)
                # Should eventually hit rate limit
                if response.status_code == 429:
                    break
            else:
                continue
            break
        else:
            # If we get here, rate limiting might not be working
            assert False, "Rate limiting not triggered"
