"""
LangGraph StateGraph for AITriage security remediation.

Enhanced graph topology (SecureCoder workflow merge):
    START → scan → threat_model → analyze → plan → fix → verify
                                                          ├─(done)     → poc → report → END
                                                          ├─(retry)    → fix
                                                          └─(give_up)  → report → END

Features:
    - SQLite checkpointing (resume interrupted runs)
    - Human-in-the-Loop support (enable via AGENT_HITL=1)
    - LangSmith tracing (via LANGCHAIN_TRACING_V2=true)
    - Dual scanner: AITriage core + SecureCoder (when IDE is running)
    - Threat model-based finding triage (True Positive / False Positive)
    - PoC verification (reasoning, not execution)
    - CS-XXX-NNN formatted security audit reports
"""
from __future__ import annotations

import functools
import os

from langgraph.graph import END, START, StateGraph

from nodes import (
    analyze_node,
    fix_node,
    plan_node,
    poc_node,
    report_node,
    scan_node,
    should_retry,
    threat_model_node,
    verify_node,
)
from state import SecurityState
from tools import load_tools


async def build_graph(checkpoint_db: str = "agent_state.db"):
    """
    Build and compile the security remediation StateGraph.

    Args:
        checkpoint_db: Path to SQLite file for state persistence.
                       Pass ":memory:" for in-memory (no persistence).

    Returns:
        A compiled LangGraph CompiledGraph ready to be invoked.
    """
    # Load all AITriage tools from the MCP server once
    print("🔌 Loading AITriage tools from MCP server...")
    tools = await load_tools()
    tool_names = [t.name for t in tools]
    print(f"  Loaded {len(tools)} tools: {', '.join(tool_names)}")

    # Bind tools to node functions that need them
    _scan = functools.partial(scan_node, tools=tools)
    _threat_model = functools.partial(threat_model_node, tools=tools)
    _verify = functools.partial(verify_node, tools=tools)

    # Build the graph
    builder = StateGraph(SecurityState)

    # ── Nodes ──
    builder.add_node("scan", _scan)
    builder.add_node("threat_model", _threat_model)  # NEW: from SecureCoder
    builder.add_node("analyze", analyze_node)
    builder.add_node("plan", plan_node)               # NEW: from SecureCoder
    builder.add_node("fix", fix_node)
    builder.add_node("verify", _verify)
    builder.add_node("poc", poc_node)                  # NEW: from SecureCoder
    builder.add_node("report", report_node)

    # ── Edges ──
    builder.add_edge(START, "scan")
    builder.add_edge("scan", "threat_model")           # NEW: threat model before analysis
    builder.add_edge("threat_model", "analyze")
    builder.add_edge("analyze", "plan")                # NEW: security plan before fix
    builder.add_edge("plan", "fix")
    builder.add_edge("fix", "verify")

    # Conditional: success → poc | retry loop | give up
    builder.add_conditional_edges(
        "verify",
        should_retry,
        {
            "done": "poc",          # NEW: PoC verification before report
            "retry": "fix",
            "give_up": "report",
        },
    )
    builder.add_edge("poc", "report")
    builder.add_edge("report", END)

    # Checkpointing for persistence and HITL
    from langgraph.checkpoint.sqlite.aio import AsyncSqliteSaver

    checkpointer = AsyncSqliteSaver.from_conn_string(checkpoint_db)

    # Human-in-the-Loop: interrupt before fix if AGENT_HITL=1
    interrupt_nodes = []
    if os.getenv("AGENT_HITL", "0") == "1":
        interrupt_nodes = ["fix"]
        print("  ⏸  Human-in-the-Loop enabled (will pause before applying fixes)")

    return builder.compile(
        checkpointer=checkpointer,
        interrupt_before=interrupt_nodes if interrupt_nodes else None,
    )
