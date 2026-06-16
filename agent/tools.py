"""
MCP client configuration for connecting LangGraph to the AITriage Go server.

The AITriage binary exposes all its scanners as MCP tools via SSE transport:
    aitriage serve --transport sse --port 9090

This module creates a configured MultiServerMCPClient and provides helpers
for loading the full tool set from the running AITriage MCP server.
"""
from __future__ import annotations

import os
from contextlib import asynccontextmanager
from typing import AsyncIterator, List

from langchain_core.tools import BaseTool
from langchain_mcp_adapters.client import MultiServerMCPClient


def _build_client_config() -> dict:
    """Build MultiServerMCPClient config from environment variables."""
    url = os.getenv("AITRIAGE_MCP_URL", "http://localhost:9090/sse")
    transport = "sse"

    # Support local stdio transport for development:
    #   AITRIAGE_MCP_TRANSPORT=stdio
    #   AITRIAGE_MCP_BINARY=/path/to/aitriage
    if os.getenv("AITRIAGE_MCP_TRANSPORT") == "stdio":
        binary = os.getenv("AITRIAGE_MCP_BINARY", "aitriage")
        return {
            "aitriage": {
                "transport": "stdio",
                "command": binary,
                "args": ["serve"],
            }
        }

    config: dict = {
        "aitriage": {
            "transport": transport,
            "url": url,
        }
    }

    # Optional auth header
    token = os.getenv("AITRIAGE_MCP_TOKEN")
    if token:
        config["aitriage"]["headers"] = {"Authorization": f"Bearer {token}"}

    return config


@asynccontextmanager
async def aitriage_client() -> AsyncIterator[MultiServerMCPClient]:
    """
    Async context manager that yields a connected MultiServerMCPClient.

    Usage:
        async with aitriage_client() as client:
            tools = await client.get_tools()
    """
    config = _build_client_config()
    async with MultiServerMCPClient(config) as client:
        yield client


async def load_tools() -> List[BaseTool]:
    """
    Load all AITriage tools from the MCP server.

    Returns a list of LangChain-compatible tool objects that can be bound
    to a LangGraph agent or ReAct graph.

    Known tools returned by AITriage MCP server:
        - aitriage_scan          Full security scan (returns ScanReport JSON)
        - aitriage_secrets       Entropy-based secret detection
        - aitriage_entropy_check Shannon entropy check on a file
        - aitriage_architecture  Detect tech stack + threat model
        - aitriage_fix_plan      Generate remediation plan
        - aitriage_scanners_list List available external scanners
        - aitriage_semgrep       Run Semgrep SAST scan
        - aitriage_gitleaks      Run gitleaks secret scanner
        - aitriage_bandit        Run bandit Python scanner
        - aitriage_trivy         Run trivy container scanner
        - aitriage_deploy_audit  IaC / Dockerfile audit
        - aitriage_nfr           Non-functional requirements check
        - aitriage_diagram       Generate Mermaid architecture diagram
        - aitriage_history       Get scan history
    """
    config = _build_client_config()
    client = MultiServerMCPClient(config)
    return await client.get_tools()


def get_tool(tools: List[BaseTool], name: str) -> BaseTool:
    """Find a tool by name, raises ValueError if not found."""
    for t in tools:
        if t.name == name:
            return t
    available = [t.name for t in tools]
    raise ValueError(
        f"Tool '{name}' not found in AITriage MCP server. "
        f"Available tools: {available}"
    )
