#!/usr/bin/env python3
"""
Navigator Agent Runner Service

FastAPI-based microservice that executes allow-listed terminal commands
for the MOVA Engine demonstration system.

Security: Only executes commands from allow-list with strict parameter validation.
"""

import json
import os
import subprocess
import time
from contextlib import asynccontextmanager
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, List, Optional

import httpx
import uvicorn
import yaml
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field, validator

# Configuration
PROJECT_ROOT = Path(os.getenv("PROJECT_ROOT", ".")).resolve()
RUNNER_PORT = int(os.getenv("RUNNER_PORT", "9090"))
RUNNER_BIND = os.getenv("RUNNER_BIND", "127.0.0.1")
ALLOWLIST_FILE = PROJECT_ROOT / "runner.allowlist.yaml"

# MOVA Engine API Configuration
MOVA_API_BASE = os.getenv("MOVA_API_BASE", "http://localhost:8080")
MOVA_API_TIMEOUT = int(os.getenv("MOVA_API_TIMEOUT", "30"))

# Rate limiting (simplified in-memory for demo)
rate_limit_store: Dict[str, List[float]] = {}

# HTTP client for MOVA Engine API
http_client: Optional[httpx.AsyncClient] = None


class CommandRequest(BaseModel):
    """Request model for command execution."""

    cmd_id: str = Field(..., description="Command ID from allow-list")
    args: Dict[str, Any] = Field(default_factory=dict, description="Command arguments")
    dry_run: bool = Field(
        False, description="If true, return planned argv without execution"
    )
    timeout_sec: int = Field(30, description="Execution timeout in seconds")

    @validator("cmd_id")
    def validate_cmd_id(cls, v):
        if not v.replace("_", "").replace("-", "").isalnum():
            raise ValueError(
                "cmd_id must contain only alphanumeric characters, "
                "hyphens, and underscores"
            )
        return v

    @validator("args")
    def validate_args(cls, v):
        for key, value in v.items():
            if isinstance(value, str) and ("\n" in value or "\r" in value):
                raise ValueError(f"Argument {key} contains newlines - not allowed")
        return v


class CommandResponse(BaseModel):
    """Response model for command execution."""

    ok: bool
    argv: List[str] = None
    returncode: int = None
    stdout_tail: str = None
    stderr_tail: str = None
    duration_ms: int = None
    error: str = None


def load_allowlist() -> Dict[str, Any]:
    """Load and parse the command allow-list."""
    if not ALLOWLIST_FILE.exists():
        raise FileNotFoundError(f"Allow-list file not found: {ALLOWLIST_FILE}")

    with open(ALLOWLIST_FILE, "r") as f:
        data = yaml.safe_load(f)
        # Handle nested structure with 'commands' key
        if isinstance(data, dict) and "commands" in data:
            return data["commands"]
        return data


def check_rate_limit() -> None:
    """Simple rate limiting check."""
    now = time.time()
    window_start = now - 60  # 1 minute window

    # Clean old entries
    rate_limit_store["requests"] = [
        t for t in rate_limit_store.get("requests", []) if t > window_start
    ]

    # Check limit (15 requests per minute for testing)
    if len(rate_limit_store["requests"]) >= 15:
        raise HTTPException(
            status_code=429, detail="Rate limit exceeded: 15 requests per minute"
        )

    rate_limit_store["requests"].append(now)


def sanitize_path(path: str, base_path: Path = PROJECT_ROOT) -> Path:
    """Sanitize file path to stay within allowed directory."""
    try:
        # Don't resolve the path, just check if it would be inside base_path
        full_path = base_path / path

        # Check for path traversal
        if ".." in path or path.startswith("/"):
            raise ValueError(f"Path outside allowed directory: {path}")

        # Check if the resolved path starts with base_path
        resolved = full_path.resolve()
        if not str(resolved).startswith(str(base_path)):
            raise ValueError(f"Path outside allowed directory: {path}")

        return full_path
    except Exception as e:
        raise ValueError(f"Invalid path: {path}") from e


def build_argv(
    cmd_id: str, args: Dict[str, Any], allowlist: Dict[str, Any]
) -> List[str]:
    """Build argv from allow-list template and arguments."""
    if cmd_id not in allowlist:
        raise HTTPException(
            status_code=403, detail=f"Command '{cmd_id}' not in allow-list"
        )

    template = allowlist[cmd_id]
    argv = []

    for item in template:
        if isinstance(item, str):
            argv.append(item)
        elif isinstance(item, dict):
            # Placeholder with validation
            for placeholder, config in item.items():
                if placeholder not in args:
                    if config.get("required", True):
                        raise HTTPException(
                            status_code=400,
                            detail=f"Missing required argument: {placeholder}",
                        )
                    continue

                value = args[placeholder]
                if isinstance(value, str):
                    # Path validation for file arguments
                    if config.get("type") == "file":
                        sanitized = sanitize_path(value)
                        argv.append(str(sanitized))
                    elif config.get("type") == "run_id":
                        # run_id validation - no whitespace, alphanumeric +
                        # underscore/hyphen
                        if not value.replace("_", "").replace("-", "").isalnum():
                            raise HTTPException(
                                status_code=400,
                                detail=f"Invalid run_id format: {value}",
                            )
                        argv.append(value)
                    else:
                        # String validation
                        if len(value) > 1000:  # Reasonable limit
                            raise HTTPException(
                                status_code=400,
                                detail=f"Argument too long: {placeholder}",
                            )
                        argv.append(value)
                else:
                    argv.append(str(value))

    return argv


def execute_command(argv: List[str], timeout_sec: int) -> Dict[str, Any]:
    """Execute command with timeout and return results."""
    start_time = time.time()
    duration_ms = 0

    try:
        env = {
            "PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
            "PROJECT_ROOT": str(PROJECT_ROOT),
        }

        result = subprocess.run(
            argv,
            capture_output=True,
            text=True,
            timeout=timeout_sec,
            env=env,
            cwd=PROJECT_ROOT,
        )

        duration_ms = int((time.time() - start_time) * 1000)

        # Limit output size (4000 chars max per stream)
        stdout_tail = result.stdout[-4000:] if result.stdout else ""
        stderr_tail = result.stderr[-4000:] if result.stderr else ""

        return {
            "returncode": result.returncode,
            "stdout_tail": stdout_tail,
            "stderr_tail": stderr_tail,
            "duration_ms": duration_ms,
        }

    except subprocess.TimeoutExpired:
        duration_ms = int((time.time() - start_time) * 1000)
        return {
            "returncode": -1,
            "stdout_tail": "",
            "stderr_tail": f"Command timed out after {timeout_sec} seconds",
            "duration_ms": duration_ms,
        }
    except Exception as e:
        duration_ms = int((time.time() - start_time) * 1000)
        return {
            "returncode": -1,
            "stdout_tail": "",
            "stderr_tail": f"Execution error: {str(e)}",
            "duration_ms": duration_ms,
        }


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan events."""
    global http_client

    # Startup
    print("🚀 Navigator Agent Runner starting...")
    print(f"📁 Project root: {PROJECT_ROOT}")
    print(f"📋 Allow-list: {ALLOWLIST_FILE}")
    print(f"🔗 MOVA API: {MOVA_API_BASE}")

    # Initialize HTTP client for MOVA Engine
    http_client = httpx.AsyncClient(
        base_url=MOVA_API_BASE,
        timeout=MOVA_API_TIMEOUT,
        headers={"Content-Type": "application/json"},
    )

    try:
        allowlist = load_allowlist()
        app.state.allowlist = allowlist
        print(f"✅ Allow-list loaded with {len(allowlist)} commands")

        # Test MOVA Engine connection
        await test_mova_connection()
        print("✅ MOVA Engine connection established")
    except Exception as e:
        print(f"❌ Failed to load allow-list: {e}")
        raise

    yield

    # Cleanup
    print("🛑 Navigator Agent Runner shutting down...")
    if http_client:
        await http_client.aclose()

    # Shutdown
    print("👋 Navigator Agent Runner shutting down...")


async def test_mova_connection() -> None:
    """Test connection to MOVA Engine API."""
    if not http_client:
        raise RuntimeError("HTTP client not initialized")

    try:
        # Test health endpoint
        response = await http_client.get("/health")
        if response.status_code != 200:
            raise RuntimeError(f"Health check failed: {response.status_code}")
    except Exception as e:
        raise RuntimeError(f"Failed to connect to MOVA Engine: {e}")


async def validate_envelope(envelope_path: Path) -> Dict[str, Any]:
    """Validate envelope using MOVA Engine API."""
    if not http_client:
        raise RuntimeError("HTTP client not initialized")

    if not envelope_path.exists():
        raise FileNotFoundError(f"Envelope not found: {envelope_path}")

    with open(envelope_path, "r") as f:
        envelope = json.load(f)

    response = await http_client.post("/v1/validate", json=envelope)

    if response.status_code != 200:
        raise HTTPException(
            status_code=400, detail=f"Envelope validation failed: {response.text}"
        )

    return response.json()


async def execute_envelope(envelope_path: Path) -> Dict[str, Any]:
    """Execute envelope using MOVA Engine API."""
    if not http_client:
        raise RuntimeError("HTTP client not initialized")

    if not envelope_path.exists():
        raise FileNotFoundError(f"Envelope not found: {envelope_path}")

    with open(envelope_path, "r") as f:
        envelope = json.load(f)

    response = await http_client.post("/v1/execute", json=envelope)

    if response.status_code != 200:
        raise HTTPException(
            status_code=400, detail=f"Envelope execution failed: {response.text}"
        )

    return response.json()


async def get_run_logs(run_id: str) -> Dict[str, Any]:
    """Get logs for a specific run using MOVA Engine API."""
    if not http_client:
        raise RuntimeError("HTTP client not initialized")

    response = await http_client.get(f"/v1/runs/{run_id}/logs")

    if response.status_code != 200:
        raise HTTPException(
            status_code=404,
            detail=f"Run not found or logs unavailable: {response.text}",
        )

    return response.json()


async def get_introspection() -> Dict[str, Any]:
    """Get MOVA Engine introspection/capabilities."""
    if not http_client:
        raise RuntimeError("HTTP client not initialized")

    response = await http_client.get("/v1/introspect")

    if response.status_code != 200:
        raise HTTPException(
            status_code=500, detail=f"Introspection failed: {response.text}"
        )

    return response.json()


# FastAPI app
app = FastAPI(
    title="Navigator Agent Runner",
    description="Safe command execution service for MOVA Engine demonstrations",
    version="1.0.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000", "http://127.0.0.1:3000"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/")
async def root():
    """Service health check."""
    return {"status": "running", "service": "navigator-runner"}


@app.get("/health")
async def health():
    """Health check endpoint."""
    return {"status": "healthy", "timestamp": datetime.now(timezone.utc).isoformat()}


@app.post("/run", response_model=CommandResponse)
async def run_command(request: CommandRequest):
    """Execute an allow-listed command."""

    # Rate limiting check
    check_rate_limit()

    allowlist = app.state.allowlist

    try:
        # Build argv from template
        argv = build_argv(request.cmd_id, request.args, allowlist)

        # Dry run - return argv without execution
        if request.dry_run:
            return CommandResponse(ok=True, argv=argv)

        # Execute command
        result = execute_command(argv, request.timeout_sec)

        # Log execution
        log_entry = {
            "ts": datetime.now(timezone.utc).isoformat(),
            "action": request.cmd_id,
            "argv": argv,
            "rc": result["returncode"],
            "dur_ms": result["duration_ms"],
        }
        print(json.dumps(log_entry), flush=True)

        return CommandResponse(
            ok=result["returncode"] == 0,
            argv=argv,
            returncode=result["returncode"],
            stdout_tail=result["stdout_tail"],
            stderr_tail=result["stderr_tail"],
            duration_ms=result["duration_ms"],
        )

    except HTTPException:
        raise
    except Exception as e:
        return CommandResponse(ok=False, error=f"Internal error: {str(e)}")


# MOVA Engine API Integration Endpoints


@app.post("/validate")
async def validate_endpoint(request: CommandRequest) -> CommandResponse:
    """Validate a MOVA envelope using MOVA Engine API."""
    try:
        check_rate_limit()

        # Validate command
        if request.cmd_id != "validate":
            raise HTTPException(
                status_code=400, detail="Only 'validate' command supported"
            )

        if "file" not in request.args:
            raise HTTPException(
                status_code=400, detail="Missing required 'file' argument"
            )

        # Sanitize file path
        file_path = sanitize_path(request.args["file"])

        # Call MOVA Engine validation
        result = await validate_envelope(file_path)

        return CommandResponse(
            ok=True, result=result, duration_ms=0  # TODO: measure actual duration
        )

    except HTTPException:
        raise
    except Exception as e:
        return CommandResponse(ok=False, error=f"Validation error: {str(e)}")


@app.post("/execute")
async def execute_endpoint(request: CommandRequest) -> CommandResponse:
    """Execute a MOVA envelope using MOVA Engine API."""
    try:
        check_rate_limit()

        # Validate command
        if request.cmd_id != "run":
            raise HTTPException(status_code=400, detail="Only 'run' command supported")

        if "file" not in request.args:
            raise HTTPException(
                status_code=400, detail="Missing required 'file' argument"
            )

        # Sanitize file path
        file_path = sanitize_path(request.args["file"])

        # Call MOVA Engine execution
        result = await execute_envelope(file_path)

        return CommandResponse(
            ok=True, result=result, duration_ms=0  # TODO: measure actual duration
        )

    except HTTPException:
        raise
    except Exception as e:
        return CommandResponse(ok=False, error=f"Execution error: {str(e)}")


@app.get("/logs/{run_id}")
async def logs_endpoint(run_id: str) -> CommandResponse:
    """Get logs for a specific run using MOVA Engine API."""
    try:
        check_rate_limit()

        # Validate run_id format
        if not run_id.replace("_", "").replace("-", "").isalnum():
            raise HTTPException(status_code=400, detail="Invalid run_id format")

        # Call MOVA Engine logs
        result = await get_run_logs(run_id)

        return CommandResponse(ok=True, result=result, duration_ms=0)

    except HTTPException:
        raise
    except Exception as e:
        return CommandResponse(ok=False, error=f"Logs error: {str(e)}")


@app.get("/introspect")
async def introspect_endpoint() -> CommandResponse:
    """Get MOVA Engine introspection/capabilities."""
    try:
        check_rate_limit()

        # Call MOVA Engine introspection
        result = await get_introspection()

        return CommandResponse(ok=True, result=result, duration_ms=0)

    except HTTPException:
        raise
    except Exception as e:
        return CommandResponse(ok=False, error=f"Introspection error: {str(e)}")


def validate_envelope_structure(envelope: Dict[str, Any]) -> bool:
    """Validate envelope structure and required fields."""
    try:
        # Check required fields
        required_fields = ["mova_version", "intent", "payload", "actions"]
        for field in required_fields:
            if field not in envelope:
                return False

        # Check version format
        if not isinstance(envelope["mova_version"], str):
            return False

        # Check intent is string
        if not isinstance(envelope["intent"], str):
            return False

        # Check payload structure
        payload = envelope["payload"]
        if not isinstance(payload, dict):
            return False
        if "action" not in payload:
            return False

        # Check actions is a list
        actions = envelope["actions"]
        if not isinstance(actions, list) or len(actions) == 0:
            return False

        # Validate each action has required fields
        for action in actions:
            if not isinstance(action, dict):
                return False
            if "type" not in action:
                return False

        return True

    except Exception:
        return False


if __name__ == "__main__":
    print("🔧 Starting Navigator Agent Runner...")
    print(f"🌐 http://{RUNNER_BIND}:{RUNNER_PORT}")
    print(f"📁 Project root: {PROJECT_ROOT}")
    print(f"📋 Allow-list: {ALLOWLIST_FILE}")

    uvicorn.run("runner:app", host=RUNNER_BIND, port=RUNNER_PORT, reload=False)
