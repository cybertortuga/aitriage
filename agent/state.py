"""
State definition for the AITriage security remediation agent.

Enhanced with SecureCoder-style workflow fields:
  - threat_model: structured threat analysis
  - finding_dispositions: True Positive / False Positive / Needs Review
  - security_plan: implementation plan with security verification section
  - poc_results: PoC verification results per finding
  - securecoder_findings: findings from SecureCoder scanner (when available)
  - findings_count_before: baseline count for fix_completed reporting
  - audit_report: CS-XXX-NNN formatted security audit report
"""
from __future__ import annotations

from typing import Annotated, Dict, List, Optional, TypedDict

from langgraph.graph.message import add_messages


class FindingDisposition(TypedDict, total=False):
    """Classification of a finding against the threat model."""

    finding: dict
    disposition: str  # "True Positive" | "False Positive" | "Needs Manual Review"
    rationale: str
    suppressed: bool  # True if auto-suppressed via SecureCoder API


class PoCResult(TypedDict, total=False):
    """PoC verification result for a single finding."""

    finding: dict
    vulnerability_type: str
    severity: str
    fix_summary: str
    reasoning_steps: List[dict]  # [{step, description, result}]
    conclusion: str  # "Fix verified" | "Fix incomplete"
    exploit_blocked: bool


class SecurityState(TypedDict):
    """
    TypedDict state container for the security remediation workflow.

    Lifecycle (SecureCoder-enhanced):
        scan → threat_model → analyze → plan → fix → verify → poc → report
    """

    project_path: str
    """Absolute path to the project being audited."""

    # ── Core scan data ────────────────────────────────────────────────
    scan_report: Optional[dict]
    """Raw JSON report from aitriage_scan MCP tool."""

    findings: List[dict]
    """Filtered list of security findings from AITriage scan."""

    securecoder_findings: List[dict]
    """Findings from SecureCoder scanner (when IDE is running)."""

    # ── Threat model (from SecureCoder determine_threat_model) ────────
    threat_model: Optional[dict]
    """Structured threat model: entry points, trust boundaries, etc."""

    finding_dispositions: List[Dict]
    """Per-finding classification: True Positive / False Positive / Needs Review."""

    # ── Planning & fixing ─────────────────────────────────────────────
    security_plan: Optional[str]
    """Implementation plan with security verification section."""

    current_finding: Optional[dict]
    """The specific finding currently being remediated."""

    generated_fix: Optional[str]
    """LLM-generated code patch for current_finding."""

    fix_verified: bool
    """True when the re-scan confirms the fix resolved the finding."""

    attempts: int
    """Number of fix attempts for current_finding (max 3 before give_up)."""

    # ── Verification & reporting ──────────────────────────────────────
    poc_results: List[Dict]
    """PoC verification results per finding (reasoning, not execution)."""

    findings_count_before: int
    """Baseline finding count (for SecureCoder /fix_completed reporting)."""

    audit_report: Optional[str]
    """CS-XXX-NNN formatted security audit report."""

    # ── LLM state ─────────────────────────────────────────────────────
    messages: Annotated[list, add_messages]
    """Chat history with the LLM (used for multi-turn reasoning)."""

    final_report: Optional[str]
    """Human-readable markdown summary of all actions taken."""
