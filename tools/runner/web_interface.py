#!/usr/bin/env python3
"""
Navigator Agent Web Interface

Simple web interface for demonstrating Navigator Agent capabilities
to investors and stakeholders.
"""

from pathlib import Path

import uvicorn
from fastapi import FastAPI, Form, HTTPException, Request
from fastapi.responses import HTMLResponse, JSONResponse
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates
from runner import (
    PROJECT_ROOT,
    execute_envelope,
    get_introspection,
    get_run_logs,
    sanitize_path,
    validate_envelope,
)

# Web interface app
web_app = FastAPI(
    title="Navigator Agent Demo",
    description="Web interface for MOVA Engine demonstrations",
    version="1.0.0",
)

# Templates
templates = Jinja2Templates(directory=Path(__file__).parent / "templates")

# Static files
static_path = Path(__file__).parent / "static"
static_path.mkdir(exist_ok=True)
web_app.mount("/static", StaticFiles(directory=static_path), name="static")


@web_app.get("/", response_class=HTMLResponse)
async def home(request: Request):
    """Home page with demo interface."""
    return templates.TemplateResponse("index.html", {"request": request})


@web_app.get("/api/introspect")
async def api_introspect():
    """Get MOVA Engine introspection."""
    try:
        result = await get_introspection()
        return JSONResponse(content=result)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@web_app.post("/api/validate")
async def api_validate(file: str = Form(...)):
    """Validate a MOVA envelope."""
    try:
        file_path = sanitize_path(file)
        result = await validate_envelope(file_path)
        return JSONResponse(content=result)
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@web_app.post("/api/execute")
async def api_execute(file: str = Form(...)):
    """Execute a MOVA envelope."""
    try:
        file_path = sanitize_path(file)
        result = await execute_envelope(file_path)
        return JSONResponse(content=result)
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@web_app.get("/api/logs/{run_id}")
async def api_logs(run_id: str):
    """Get logs for a specific run."""
    try:
        result = await get_run_logs(run_id)
        return JSONResponse(content=result)
    except Exception as e:
        raise HTTPException(status_code=404, detail=str(e))


@web_app.get("/api/envelopes")
async def api_envelopes():
    """List available demo envelopes."""
    envelopes_dir = PROJECT_ROOT / "envelopes"
    if not envelopes_dir.exists():
        return JSONResponse(content=[])

    envelopes = []
    for file in envelopes_dir.glob("*.json"):
        envelopes.append(
            {
                "name": file.name,
                "path": str(file.relative_to(PROJECT_ROOT)),
                "size": file.stat().st_size,
            }
        )

    return JSONResponse(content=envelopes)


if __name__ == "__main__":
    import os

    host = os.getenv("WEB_BIND", "127.0.0.1")
    port = int(os.getenv("WEB_PORT", "9091"))
    uvicorn.run("web_interface:web_app", host=host, port=port, reload=True)
