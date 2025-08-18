"""MOVA Python SDK Client."""

import json
from typing import Dict, List, Optional, Union
from urllib.parse import urljoin

import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

from .exceptions import (
    MOVAAPIError,
    MOVAConnectionError,
    MOVATimeoutError,
    MOVAValidationError,
)
from .models import (
    AsyncExecutionResult,
    ExecutionResult,
    IntrospectionResult,
    MOVAEnvelope,
    SchemasResponse,
    ValidationResult,
)


class MOVAClient:
    """MOVA Automation Engine Python SDK Client."""

    def __init__(
        self,
        base_url: str = "http://localhost:8080",
        timeout: float = 30.0,
        headers: Optional[Dict[str, str]] = None,
        retry_config: Optional[Dict[str, int]] = None,
    ):
        """Initialize MOVA client.

        Args:
            base_url: Base URL of the MOVA API server
            timeout: Request timeout in seconds
            headers: Additional headers to send with requests
            retry_config: Retry configuration (total, backoff_factor, status_forcelist)
        """
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.session = requests.Session()

        # Set default headers
        default_headers = {
            "Content-Type": "application/json",
            "User-Agent": "mova-python-sdk/1.0.0",
        }
        if headers:
            default_headers.update(headers)
        self.session.headers.update(default_headers)

        # Configure retry strategy
        retry_defaults = {
            "total": 3,
            "backoff_factor": 0.3,
            "status_forcelist": [500, 502, 503, 504],
        }
        if retry_config:
            retry_defaults.update(retry_config)

        retry_strategy = Retry(**retry_defaults)
        adapter = HTTPAdapter(max_retries=retry_strategy)
        self.session.mount("http://", adapter)
        self.session.mount("https://", adapter)

    def execute(
        self, envelope: Union[MOVAEnvelope, Dict], wait: bool = False
    ) -> Union[ExecutionResult, AsyncExecutionResult]:
        """Execute a MOVA workflow envelope.

        Args:
            envelope: The MOVA envelope to execute
            wait: Whether to wait for execution to complete

        Returns:
            Execution result (sync) or async execution info

        Raises:
            MOVAAPIError: If the API returns an error
            MOVAConnectionError: If connection fails
            MOVATimeoutError: If request times out
        """
        if isinstance(envelope, MOVAEnvelope):
            envelope_dict = envelope.model_dump()
        else:
            envelope_dict = envelope

        url = self._build_url("/v1/execute")
        params = {"wait": "true"} if wait else {}

        try:
            response = self._make_request(
                "POST", url, json=envelope_dict, params=params
            )

            if wait:
                return ExecutionResult(**response)
            else:
                return AsyncExecutionResult(**response)

        except requests.exceptions.Timeout as e:
            raise MOVATimeoutError(f"Request timeout after {self.timeout}s") from e
        except requests.exceptions.ConnectionError as e:
            raise MOVAConnectionError(f"Connection failed: {e}") from e

    def validate(self, envelope: Union[MOVAEnvelope, Dict]) -> ValidationResult:
        """Validate a MOVA envelope against the schema.

        Args:
            envelope: The MOVA envelope to validate

        Returns:
            Validation result

        Raises:
            MOVAAPIError: If the API returns an error
            MOVAConnectionError: If connection fails
            MOVATimeoutError: If request times out
        """
        if isinstance(envelope, MOVAEnvelope):
            envelope_dict = envelope.model_dump()
        else:
            envelope_dict = envelope

        url = self._build_url("/v1/validate")

        try:
            response = self._make_request("POST", url, json=envelope_dict)
            return ValidationResult(**response)
        except requests.exceptions.Timeout as e:
            raise MOVATimeoutError(f"Request timeout after {self.timeout}s") from e
        except requests.exceptions.ConnectionError as e:
            raise MOVAConnectionError(f"Connection failed: {e}") from e

    def get_run(self, run_id: str) -> ExecutionResult:
        """Get the status and result of a workflow execution.

        Args:
            run_id: The run ID to retrieve

        Returns:
            Execution result

        Raises:
            MOVAAPIError: If the API returns an error
            MOVAConnectionError: If connection fails
            MOVATimeoutError: If request times out
        """
        url = self._build_url(f"/v1/runs/{run_id}")

        try:
            response = self._make_request("GET", url)
            return ExecutionResult(**response)
        except requests.exceptions.Timeout as e:
            raise MOVATimeoutError(f"Request timeout after {self.timeout}s") from e
        except requests.exceptions.ConnectionError as e:
            raise MOVAConnectionError(f"Connection failed: {e}") from e

    def get_logs(self, run_id: str) -> List[str]:
        """Get the logs for a workflow execution.

        Args:
            run_id: The run ID to retrieve logs for

        Returns:
            List of JSONL log entries

        Raises:
            MOVAAPIError: If the API returns an error
            MOVAConnectionError: If connection fails
            MOVATimeoutError: If request times out
        """
        url = self._build_url(f"/v1/runs/{run_id}/logs")

        try:
            response = self.session.get(url, timeout=self.timeout)

            if not response.ok:
                self._handle_error_response(response)

            text = response.text.strip()
            if not text:
                return []

            return [line for line in text.split("\n") if line.strip()]

        except requests.exceptions.Timeout as e:
            raise MOVATimeoutError(f"Request timeout after {self.timeout}s") from e
        except requests.exceptions.ConnectionError as e:
            raise MOVAConnectionError(f"Connection failed: {e}") from e

    def get_schemas(self) -> SchemasResponse:
        """Get available schemas.

        Returns:
            Schemas information

        Raises:
            MOVAAPIError: If the API returns an error
            MOVAConnectionError: If connection fails
            MOVATimeoutError: If request times out
        """
        url = self._build_url("/v1/schemas")

        try:
            response = self._make_request("GET", url)
            return SchemasResponse(**response)
        except requests.exceptions.Timeout as e:
            raise MOVATimeoutError(f"Request timeout after {self.timeout}s") from e
        except requests.exceptions.ConnectionError as e:
            raise MOVAConnectionError(f"Connection failed: {e}") from e

    def get_schema(self, name: str) -> Dict:
        """Get a specific schema by name.

        Args:
            name: Schema name (e.g., 'envelope', 'action')

        Returns:
            Schema definition

        Raises:
            MOVAAPIError: If the API returns an error
            MOVAConnectionError: If connection fails
            MOVATimeoutError: If request times out
        """
        url = self._build_url(f"/v1/schemas/{name}")

        try:
            response = self._make_request("GET", url)
            return response
        except requests.exceptions.Timeout as e:
            raise MOVATimeoutError(f"Request timeout after {self.timeout}s") from e
        except requests.exceptions.ConnectionError as e:
            raise MOVAConnectionError(f"Connection failed: {e}") from e

    def introspect(self) -> IntrospectionResult:
        """Get API introspection information.

        Returns:
            API information

        Raises:
            MOVAAPIError: If the API returns an error
            MOVAConnectionError: If connection fails
            MOVATimeoutError: If request times out
        """
        url = self._build_url("/v1/introspect")

        try:
            response = self._make_request("GET", url)
            return IntrospectionResult(**response)
        except requests.exceptions.Timeout as e:
            raise MOVATimeoutError(f"Request timeout after {self.timeout}s") from e
        except requests.exceptions.ConnectionError as e:
            raise MOVAConnectionError(f"Connection failed: {e}") from e

    def health(self) -> Dict:
        """Check API health status.

        Returns:
            Health status information

        Raises:
            MOVAAPIError: If the API returns an error
            MOVAConnectionError: If connection fails
            MOVATimeoutError: If request times out
        """
        url = self._build_url("/health")

        try:
            response = self._make_request("GET", url)
            return response
        except requests.exceptions.Timeout as e:
            raise MOVATimeoutError(f"Request timeout after {self.timeout}s") from e
        except requests.exceptions.ConnectionError as e:
            raise MOVAConnectionError(f"Connection failed: {e}") from e

    def _build_url(self, path: str) -> str:
        """Build full URL from path."""
        return urljoin(self.base_url, path)

    def _make_request(self, method: str, url: str, **kwargs) -> Dict:
        """Make HTTP request and handle response."""
        response = self.session.request(method, url, timeout=self.timeout, **kwargs)

        if not response.ok:
            self._handle_error_response(response)

        try:
            return response.json()
        except json.JSONDecodeError as e:
            raise MOVAAPIError(f"Invalid JSON response: {e}") from e

    def _handle_error_response(self, response: requests.Response) -> None:
        """Handle error response from API."""
        try:
            error_data = response.json()
            error_message = error_data.get("error", response.reason)
            details = error_data.get("details", "")
            full_message = f"{error_message}"
            if details:
                full_message += f": {details}"
        except json.JSONDecodeError:
            full_message = f"HTTP {response.status_code}: {response.reason}"

        if response.status_code == 400:
            raise MOVAValidationError(full_message)
        else:
            raise MOVAAPIError(full_message)

    def __enter__(self):
        """Context manager entry."""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.session.close()
