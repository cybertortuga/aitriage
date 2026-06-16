"""
AITriage Agent — CLI entry point.

Usage:
    # Run full remediation workflow
    python main.py /path/to/project

    # Resume a previous interrupted run (via thread-id)
    python main.py /path/to/project --thread-id my-run-1

    # Enable Human-in-the-Loop (pauses before applying each fix)
    AGENT_HITL=1 python main.py /path/to/project

    # Use a different LLM model
    AGENT_MODEL=anthropic:claude-sonnet-4-5 python main.py /path/to/project

    # Use local AITriage binary (stdio transport instead of SSE)
    AITRIAGE_MCP_TRANSPORT=stdio AITRIAGE_MCP_BINARY=./aitriage python main.py /path/to/project

Environment variables:
    AITRIAGE_MCP_URL          URL of AITriage MCP SSE server (default: http://localhost:9090/sse)
    AITRIAGE_MCP_TRANSPORT    "sse" (default) or "stdio"
    AITRIAGE_MCP_BINARY       Path to aitriage binary (for stdio transport)
    AGENT_MODEL               LangChain model string (default: google-genai:gemini-2.5-flash)
    AGENT_HITL                "1" to enable Human-in-the-Loop pauses before fixes
    LANGCHAIN_TRACING_V2      "true" to enable LangSmith tracing
    LANGCHAIN_API_KEY         LangSmith API key
    LANGCHAIN_PROJECT         LangSmith project name (default: aitriage)
    GEMINI_API_KEY            Google Gemini API key
    ANTHROPIC_API_KEY         Anthropic Claude API key
    OPENAI_API_KEY            OpenAI API key
"""
from __future__ import annotations

import asyncio
import os
import sys
import uuid
from argparse import ArgumentParser

from dotenv import load_dotenv

# Load .env file if present
load_dotenv()

# Set LangSmith project default
os.environ.setdefault("LANGCHAIN_PROJECT", "aitriage")


async def run(project_path: str, thread_id: str) -> None:
    """Execute the security remediation graph for a project."""
    from graph import build_graph

    graph = await build_graph()
    config = {"configurable": {"thread_id": thread_id}}

    initial_state = {
        "project_path": os.path.abspath(project_path),
        "scan_report": None,
        "findings": [],
        "current_finding": None,
        "generated_fix": None,
        "fix_verified": False,
        "attempts": 0,
        "messages": [],
        "final_report": None,
    }

    hitl_enabled = os.getenv("AGENT_HITL", "0") == "1"

    print(f"\n{'='*60}")
    print(f"  AITriage Security Agent")
    print(f"  Project: {initial_state['project_path']}")
    print(f"  Thread:  {thread_id}")
    print(f"  Model:   {os.getenv('AGENT_MODEL', 'google-genai:gemini-2.5-flash')}")
    print(f"  HITL:    {'enabled' if hitl_enabled else 'disabled'}")
    print(f"{'='*60}\n")

    # Stream graph execution — prints node outputs as they complete
    async for event in graph.astream(initial_state, config=config, stream_mode="values"):
        # HITL: pause before 'fix' node if enabled
        state = await graph.aget_state(config)
        if state.next and "fix" in state.next and hitl_enabled:
            finding = event.get("current_finding", {})
            print(f"\n⏸  Human-in-the-Loop pause")
            print(f"   Ready to generate fix for:")
            print(f"   [{finding.get('severity','?')}] {finding.get('file','?')}:{finding.get('line','?')}")
            print(f"   {finding.get('message','')}\n")
            answer = input("Proceed with fix? [Y/n]: ").strip().lower()
            if answer in ("n", "no"):
                print("❌ Fix skipped by user.")
                break
            # Resume the graph
            await graph.aupdate_state(config, {}, as_node="fix")

    # Retrieve final state
    final_state = await graph.aget_state(config)
    values = final_state.values

    # Print final report
    report = values.get("final_report", "")
    if report:
        print(f"\n{'='*60}")
        print(report)
        print(f"{'='*60}\n")

    # Write report to file
    report_path = f"aitriage_agent_report_{thread_id[:8]}.md"
    with open(report_path, "w") as f:
        f.write(report or "No report generated.")
    print(f"📄 Report saved to: {report_path}")


def main() -> None:
    parser = ArgumentParser(
        description="AITriage Security Remediation Agent (LangGraph)",
        formatter_class=lambda prog: __import__("argparse").RawDescriptionHelpFormatter(
            prog, max_help_position=40
        ),
    )
    parser.add_argument(
        "path",
        nargs="?",
        default=".",
        help="Path to the project to analyze (default: current directory)",
    )
    parser.add_argument(
        "--thread-id",
        default=None,
        help="Resume a previous run by thread ID (default: new random UUID)",
    )
    parser.add_argument(
        "--checkpoint-db",
        default="agent_state.db",
        help="SQLite checkpoint file path (default: agent_state.db)",
    )
    args = parser.parse_args()

    thread_id = args.thread_id or str(uuid.uuid4())
    asyncio.run(run(args.path, thread_id))


if __name__ == "__main__":
    main()
