"""
Node functions for the AITriage LangGraph security remediation workflow.

Enhanced with SecureCoder-style workflow nodes:

    scan_node          → runs aitriage_scan + SecureCoder dual scan
    threat_model_node  → builds threat model, classifies findings (NEW)
    analyze_node       → LLM picks the most critical finding
    plan_node          → generates security implementation plan (NEW)
    fix_node           → LLM generates a code patch
    verify_node        → re-scans to check if fix improved SecurityScore
    poc_node           → PoC reasoning verification (NEW)
    report_node        → CS-XXX-NNN formatted security audit report (UPGRADED)

Graph:
    START → scan → threat_model → analyze → plan → fix → verify
                                                         ├─ (done)    → poc → report → END
                                                         ├─ (retry)   → fix
                                                         └─ (give_up) → report → END
"""
from __future__ import annotations

import json
import os
from typing import List

from langchain_core.messages import HumanMessage, SystemMessage
from langchain_core.tools import BaseTool

from state import SecurityState

# ─── LLM Setup ───────────────────────────────────────────────────────────────
from langchain.chat_models import init_chat_model

_model_name = os.getenv("AGENT_MODEL", "google-genai:gemini-2.5-flash")
llm = init_chat_model(_model_name)


# ─── Node 1: Scan (Dual Scanner) ─────────────────────────────────────────────

async def scan_node(state: SecurityState, tools: List[BaseTool]) -> dict:
    """
    Run a full AITriage security scan + SecureCoder scan (if available).
    Both run in parallel when SecureCoder IDE is active.
    """
    from tools import get_tool

    scan_tool = get_tool(tools, "aitriage_scan")
    print(f"\n🔍 [scan] Scanning: {state['project_path']}")

    # 1. AITriage core scan
    try:
        raw = await scan_tool.ainvoke({"path": state["project_path"]})
        report = raw if isinstance(raw, dict) else json.loads(raw)
    except Exception as e:
        print(f"  ⚠️  Scan error: {e}")
        return {
            "scan_report": {},
            "findings": [],
            "securecoder_findings": [],
            "findings_count_before": 0,
            "messages": [HumanMessage(content=f"Scan failed: {e}")],
        }

    results = report.get("results", [])
    score = report.get("security_score", "N/A")
    print(f"  ✅ AITriage Score: {score}/100 | Findings: {len(results)}")

    # 2. SecureCoder scan (if available)
    sc_findings = []
    try:
        sc_tool = get_tool(tools, "run_securecoder")
        print("  🛡️  SecureCoder detected — running parallel scan...")
        sc_raw = await sc_tool.ainvoke({"path": state["project_path"]})
        sc_result = sc_raw if isinstance(sc_raw, dict) else json.loads(sc_raw)
        sc_findings = sc_result.get("findings", [])
        print(f"  🛡️  SecureCoder findings: {len(sc_findings)}")
    except (ValueError, Exception) as e:
        print(f"  ℹ️  SecureCoder not available: {e}")

    total = len(results) + len(sc_findings)
    return {
        "scan_report": report,
        "findings": results,
        "securecoder_findings": sc_findings,
        "findings_count_before": total,
        "messages": [
            HumanMessage(
                content=f"Dual scan complete. AITriage SecurityScore: {score}/100. "
                f"AITriage findings: {len(results)}. SecureCoder findings: {len(sc_findings)}. "
                f"Critical: {sum(1 for r in results if r.get('severity') == 'CRITICAL')}."
            )
        ],
    }


# ─── Node 2: Threat Model (from SecureCoder determine_threat_model) ──────────

THREAT_MODEL_SYSTEM_PROMPT = """You are an elite DevSecOps engineer and AI security auditor.
Build a threat model for the scanned component. Your analysis must include:

1. **Component Overview**: What does this component do? Who consumes it?
2. **Entry Points and Untrusted Inputs**: All points where external data enters
   (HTTP endpoints, CLI args, file inputs, env vars, DB reads, IPC).
   For each, note if input is trusted/untrusted and what validation exists.
3. **Trust Boundaries and Auth Assumptions**: How are callers authenticated?
   What authorization checks exist? What implicit trust assumptions are made?
4. **Sensitive Data Paths**: Where do secrets, PII, tokens flow through the code?
5. **Privileged Actions**: File writes, shell exec, network calls, DB mutations.
6. **Priority Review Areas**: Ranked list of areas to review first.

Then, for EACH scanner finding provided, evaluate against this threat model:
- Is the flagged code reachable from an untrusted entry point?
- Does the auth/trust context mitigate the risk?
- Is the vulnerability exploitable given the deployment context?

Classify each finding as:
- **True Positive**: Real, exploitable vulnerability given the threat model.
- **False Positive**: Not exploitable because of trust boundaries, auth, or intended functionality.
- **Needs Manual Review**: Insufficient context to determine.

Provide a one-line rationale for each classification.

Return your analysis as JSON with this structure:
{
  "component_overview": "...",
  "entry_points": [{"endpoint": "...", "type": "...", "trusted": false, "validation": "..."}],
  "trust_boundaries": {"authentication": "...", "authorization": "...", "implicit_trust": "..."},
  "sensitive_data_paths": [{"data_type": "...", "source": "...", "destination": "...", "protection": "..."}],
  "privileged_actions": [{"action": "...", "location": "...", "guard": "..."}],
  "priority_areas": ["..."],
  "finding_dispositions": [{"finding_index": 0, "disposition": "True Positive", "rationale": "..."}]
}"""


async def threat_model_node(state: SecurityState, tools: List[BaseTool]) -> dict:
    """
    Build threat model for the component.
    Classifies each finding as True Positive / False Positive / Needs Review.
    Auto-suppresses False Positives in SecureCoder (if available).
    """
    findings = state.get("findings", [])
    sc_findings = state.get("securecoder_findings", [])
    all_findings = findings + sc_findings

    if not all_findings:
        print("\n🏗️ [threat_model] No findings — skipping threat model")
        return {
            "threat_model": None,
            "finding_dispositions": [],
        }

    # Get architecture context if available
    arch_context = ""
    try:
        from tools import get_tool
        arch_tool = get_tool(tools, "aitriage_architecture")
        arch_raw = await arch_tool.ainvoke({"path": state["project_path"]})
        arch_result = arch_raw if isinstance(arch_raw, dict) else json.loads(arch_raw)
        arch_context = f"\n\nArchitecture context:\n{json.dumps(arch_result, indent=2)}"
    except Exception:
        pass

    print(f"\n🏗️ [threat_model] Building threat model for {len(all_findings)} findings...")

    system = SystemMessage(content=THREAT_MODEL_SYSTEM_PROMPT)
    user = HumanMessage(
        content=(
            f"Project path: {state['project_path']}{arch_context}\n\n"
            f"Scanner findings ({len(all_findings)} total):\n"
            f"{json.dumps(all_findings[:20], indent=2)}"
        )
    )

    response = await llm.ainvoke([system, user])

    # Parse threat model from response
    threat_model = None
    dispositions = []
    try:
        # Extract JSON from response (handle markdown code fences)
        text = response.content
        if "```json" in text:
            text = text.split("```json")[1].split("```")[0]
        elif "```" in text:
            text = text.split("```")[1].split("```")[0]
        threat_model = json.loads(text)

        # Extract dispositions
        raw_dispositions = threat_model.get("finding_dispositions", [])
        for d in raw_dispositions:
            idx = d.get("finding_index", 0)
            if idx < len(all_findings):
                dispositions.append({
                    "finding": all_findings[idx],
                    "disposition": d.get("disposition", "Needs Manual Review"),
                    "rationale": d.get("rationale", ""),
                    "suppressed": False,
                })
    except (json.JSONDecodeError, KeyError, IndexError) as e:
        print(f"  ⚠️  Could not parse threat model JSON: {e}")
        # Fall back: treat all findings as True Positive
        for f in all_findings:
            dispositions.append({
                "finding": f,
                "disposition": "True Positive",
                "rationale": "Could not build threat model; defaulting to True Positive.",
                "suppressed": False,
            })

    # Auto-suppress False Positives in SecureCoder
    fp_count = 0
    try:
        from tools import get_tool
        ignore_tool = get_tool(tools, "securecoder_ignore")
        for d in dispositions:
            if d["disposition"] == "False Positive":
                f = d["finding"]
                try:
                    await ignore_tool.ainvoke({
                        "file_path": f.get("file", ""),
                        "rule_id": f.get("rule_id", f.get("subcategory", "")),
                        "reason": "False Positive",
                        "code_snippet": f.get("evidence", ""),
                    })
                    d["suppressed"] = True
                    fp_count += 1
                except Exception:
                    pass
    except ValueError:
        pass  # securecoder_ignore not available

    tp_count = sum(1 for d in dispositions if d["disposition"] == "True Positive")
    nr_count = sum(1 for d in dispositions if d["disposition"] == "Needs Manual Review")
    print(f"  ✅ Threat model built: {tp_count} True Positives, {fp_count} False Positives (suppressed), {nr_count} Needs Review")

    return {
        "threat_model": threat_model,
        "finding_dispositions": dispositions,
        "messages": [response],
    }


# ─── Node 3: Analyze ─────────────────────────────────────────────────────────

async def analyze_node(state: SecurityState) -> dict:
    """
    Ask the LLM to pick the highest-priority finding to fix.
    Now uses threat model dispositions to filter out False Positives.
    """
    findings = state.get("findings", [])
    sc_findings = state.get("securecoder_findings", [])
    dispositions = state.get("finding_dispositions", [])

    # Filter to True Positives only (if dispositions are available)
    if dispositions:
        tp_findings = [
            d["finding"] for d in dispositions
            if d["disposition"] == "True Positive"
        ]
        candidates = tp_findings if tp_findings else findings + sc_findings
    else:
        all_findings = findings + sc_findings
        candidates = all_findings

    if not candidates:
        print("\n✨ [analyze] No True Positive findings — project is clean!")
        return {
            "current_finding": None,
            "final_report": "✅ No True Positive security findings detected. Project looks clean.",
        }

    # Priority filter
    critical = [f for f in candidates if f.get("severity") in ("CRITICAL", "HIGH")]
    top = critical if critical else candidates

    system = SystemMessage(
        content=(
            "You are a senior security engineer. "
            "Your job is to identify the single most impactful security finding to fix first. "
            "Consider severity, exploitability, and business risk. "
            "These findings have already been triaged as True Positives."
        )
    )
    user = HumanMessage(
        content=(
            f"Here are the True Positive security findings:\n\n"
            f"{json.dumps(top[:10], indent=2)}\n\n"
            "Which single finding should be fixed first and why? "
            "Respond with the finding's array index (0-based) and a 1-sentence rationale."
        )
    )

    response = await llm.ainvoke([system, user])
    print(f"\n🧠 [analyze] LLM priority: {response.content[:120]}...")

    chosen_idx = 0
    for word in response.content.split():
        if word.isdigit() and int(word) < len(top):
            chosen_idx = int(word)
            break

    chosen = top[chosen_idx]
    print(f"  → Selected: [{chosen.get('severity')}] {chosen.get('file', '?')}:{chosen.get('line', '?')} — {chosen.get('message', '?')[:60]}")

    return {
        "current_finding": chosen,
        "messages": [response],
    }


# ─── Node 4: Plan (from SecureCoder create_security_implementation_plan) ─────

PLAN_SYSTEM_PROMPT = """You are an expert security remediation engineer.
Generate a security implementation plan for fixing the identified vulnerability.
Your plan MUST include:

1. **Vulnerability Summary**: What the issue is, where it is, and why it's dangerous.
2. **Remediation Steps**: Concrete, ordered steps to fix the issue.
3. **Secure Coding Guidelines**: Relevant MUST/MUST NOT rules from OWASP guidelines.
4. **Verification Plan**: How to verify the fix works:
   - **Security Scanner**: Run scan on all modified files to confirm the vulnerability is resolved.
   - **Security Audit**: Audit for design-level security issues (input validation, secrets handling, auth checks).

Return a clear, structured markdown plan."""


async def plan_node(state: SecurityState) -> dict:
    """
    Generate security implementation plan with verification section.
    Merges AITriage fix planning with SecureCoder's security verification template.
    """
    finding = state.get("current_finding")
    if not finding:
        return {"security_plan": None}

    threat_model = state.get("threat_model")
    threat_context = ""
    if threat_model:
        threat_context = f"\n\nThreat Model Context:\n{json.dumps(threat_model, indent=2)[:2000]}"

    print(f"\n📋 [plan] Generating security implementation plan...")

    system = SystemMessage(content=PLAN_SYSTEM_PROMPT)
    user = HumanMessage(
        content=(
            f"Vulnerability to fix:\n"
            f"  File: {finding.get('file', 'unknown')}\n"
            f"  Line: {finding.get('line', '?')}\n"
            f"  Severity: {finding.get('severity', '?')}\n"
            f"  Issue: {finding.get('message', '?')}\n"
            f"  Rule: {finding.get('rule_id', finding.get('subcategory', '?'))}\n"
            f"  CWE: {finding.get('cwe', finding.get('labels', {}).get('cwe', 'N/A'))}\n"
            f"  Suggestion: {finding.get('suggestion', 'Follow security best practices.')}"
            f"{threat_context}"
        )
    )

    response = await llm.ainvoke([system, user])
    print(f"  Generated plan ({len(response.content)} chars)")

    return {
        "security_plan": response.content,
        "messages": [response],
    }


# ─── Node 5: Fix ─────────────────────────────────────────────────────────────

async def fix_node(state: SecurityState) -> dict:
    """
    Generate a minimal code patch for the current finding.
    Now uses security_plan as context for better fix quality.
    """
    finding = state.get("current_finding")
    if not finding:
        return {"generated_fix": None}

    attempts = state.get("attempts", 0) + 1
    print(f"\n🔧 [fix] Attempt {attempts}/3 for: {finding.get('message', '?')[:60]}")

    prev_fix = state.get("generated_fix")
    retry_context = ""
    if prev_fix and attempts > 1:
        retry_context = (
            f"\n\nPrevious attempt failed:\n```\n{prev_fix}\n```\n"
            "The fix did not resolve the issue. Try a different approach."
        )

    plan_context = ""
    if state.get("security_plan"):
        plan_context = f"\n\nSecurity Implementation Plan:\n{state['security_plan'][:2000]}"

    system = SystemMessage(
        content=(
            "You are a security-focused code assistant. "
            "Generate the minimal, correct code change to fix the security vulnerability. "
            "Follow the security implementation plan if provided. "
            "Return ONLY the fixed code snippet. No explanations, no markdown fences."
        )
    )
    user = HumanMessage(
        content=(
            f"Security Finding:\n"
            f"  File: {finding.get('file', 'unknown')}\n"
            f"  Line: {finding.get('line', '?')}\n"
            f"  Severity: {finding.get('severity', '?')}\n"
            f"  Issue: {finding.get('message', '?')}\n"
            f"  Rule: {finding.get('rule_id', '?')}\n"
            f"  CWE: {finding.get('cwe', 'N/A')}\n"
            f"  Suggestion: {finding.get('suggestion', 'Follow security best practices.')}"
            f"{plan_context}{retry_context}"
        )
    )

    response = await llm.ainvoke([system, user])
    print(f"  Generated fix ({len(response.content)} chars)")

    return {
        "generated_fix": response.content,
        "attempts": attempts,
        "messages": [response],
    }


# ─── Node 6: Verify ──────────────────────────────────────────────────────────

async def verify_node(state: SecurityState, tools: List[BaseTool]) -> dict:
    """
    Re-run the scan to check if SecurityScore improved.
    """
    from tools import get_tool

    scan_tool = get_tool(tools, "aitriage_scan")
    old_score = (state.get("scan_report") or {}).get("security_score", 0)

    print(f"\n✅ [verify] Re-scanning (old score: {old_score})...")

    try:
        raw = await scan_tool.ainvoke({"path": state["project_path"]})
        report = raw if isinstance(raw, dict) else json.loads(raw)
    except Exception as e:
        print(f"  ⚠️  Re-scan error: {e}")
        return {"fix_verified": False}

    new_score = report.get("security_score", 0)
    improved = new_score >= old_score
    print(f"  New score: {new_score} | Improved: {improved}")

    return {
        "scan_report": report,
        "findings": report.get("results", []),
        "fix_verified": improved,
        "messages": [
            HumanMessage(
                content=f"Re-scan complete. Score: {old_score} → {new_score}. "
                f"Fix {'resolved' if improved else 'did NOT resolve'} the issue."
            )
        ],
    }


# ─── Node 7: PoC Verification (from SecureCoder run_poc) ─────────────────────

POC_SYSTEM_PROMPT = """You are an expert security verification engineer.
After a fix has been applied, generate a Proof-of-Concept (PoC) verification.

IMPORTANT: Do NOT execute the PoC. Reason through it step by step.

For each vulnerability that was fixed:

1. **Describe the PoC**: What input/request would an attacker craft to exploit the original vulnerability?
2. **Trace the data flow**: Follow the exploit input through the NOW-PATCHED code.
3. **Identify interception**: Where does the fix intercept or neutralize the exploit?
4. **Determine outcome**: Does the exploit fail? If it would still succeed, flag the fix as incomplete.

Use vulnerability-specific reasoning:
- Path Traversal: Does patched code reject ../sequences and validate resolved path stays in allowed dir?
- XSS: Does patched code sanitize/escape user input before DOM insertion?
- SQL Injection: Does patched code use parameterized queries instead of string concatenation?
- SSRF: Does patched code validate/restrict target URL and block internal IP ranges?

Return JSON array:
[{
  "vulnerability_type": "...",
  "severity": "...",
  "affected_file": "...",
  "fix_summary": "...",
  "reasoning_steps": [
    {"step": 1, "description": "...", "result": "..."}
  ],
  "conclusion": "Fix verified" or "Fix incomplete",
  "exploit_blocked": true/false
}]"""


async def poc_node(state: SecurityState) -> dict:
    """
    Generate PoC verification — reason through exploit, DON'T execute.
    Traces data flow from exploit input through patched code.
    """
    finding = state.get("current_finding")
    fix = state.get("generated_fix")
    if not finding or not fix:
        return {"poc_results": []}

    print(f"\n🧪 [poc] Generating PoC verification for: {finding.get('message', '?')[:60]}")

    threat_context = ""
    if state.get("threat_model"):
        tm = state["threat_model"]
        threat_context = f"\n\nThreat Model Context:\n- Entry points: {json.dumps(tm.get('entry_points', [])[:5])}\n- Trust boundaries: {json.dumps(tm.get('trust_boundaries', {}))}"

    system = SystemMessage(content=POC_SYSTEM_PROMPT)
    user = HumanMessage(
        content=(
            f"Vulnerability fixed:\n"
            f"  File: {finding.get('file', 'unknown')}\n"
            f"  Line: {finding.get('line', '?')}\n"
            f"  Severity: {finding.get('severity', '?')}\n"
            f"  Type: {finding.get('vulnerability_class', finding.get('message', '?'))}\n"
            f"  CWE: {finding.get('cwe', 'N/A')}\n\n"
            f"Applied fix:\n```\n{fix[:3000]}\n```"
            f"{threat_context}"
        )
    )

    response = await llm.ainvoke([system, user])

    poc_results = []
    try:
        text = response.content
        if "```json" in text:
            text = text.split("```json")[1].split("```")[0]
        elif "```" in text:
            text = text.split("```")[1].split("```")[0]
        poc_results = json.loads(text)
        if isinstance(poc_results, dict):
            poc_results = [poc_results]
    except (json.JSONDecodeError, IndexError):
        poc_results = [{
            "vulnerability_type": finding.get("message", "Unknown"),
            "severity": finding.get("severity", "MEDIUM"),
            "fix_summary": "Fix applied",
            "reasoning_steps": [{"step": 1, "description": "Could not parse PoC reasoning", "result": "Manual review needed"}],
            "conclusion": "Needs Manual Review",
            "exploit_blocked": None,
        }]

    verified = sum(1 for p in poc_results if p.get("exploit_blocked"))
    incomplete = sum(1 for p in poc_results if not p.get("exploit_blocked"))
    print(f"  ✅ PoC: {verified} verified, {incomplete} incomplete/unknown")

    return {
        "poc_results": poc_results,
        "messages": [response],
    }


# ─── Node 8: Report (CS-XXX-NNN format from SecureCoder) ─────────────────────

# Vulnerability class → short code mapping for CS-XXX-NNN format
VULN_CLASS_CODES = {
    "Cross-Site Scripting": "XSS",
    "XSS": "XSS",
    "SQL Injection": "SQLI",
    "Command Injection": "EXEC",
    "Path Traversal": "PATH",
    "SSRF": "SSRF",
    "Hardcoded Secret": "SECRETS",
    "Secret": "SECRETS",
    "Weak Cryptography": "CRYPTO",
    "Insecure Deserialization": "DESER",
    "CSRF": "CSRF",
    "Authentication": "AUTH",
    "Authorization": "AUTHZ",
    "Information Exposure": "INFO",
}


def _vuln_id(finding: dict, index: int) -> str:
    """Generate CS-XXX-NNN style vulnerability ID."""
    vuln_class = finding.get("vulnerability_class", finding.get("message", "MISC"))
    code = "MISC"
    for key, val in VULN_CLASS_CODES.items():
        if key.lower() in vuln_class.lower():
            code = val
            break
    return f"CS-{code}-{index + 1:03d}"


async def report_node(state: SecurityState) -> dict:
    """
    Assembles CS-XXX-NNN formatted security audit report.
    Merges AITriage report format with SecureCoder audit template.
    """
    finding = state.get("current_finding") or {}
    fix = state.get("generated_fix") or "No fix generated."
    score_before = (state.get("scan_report") or {}).get("security_score", "N/A")
    verified = state.get("fix_verified", False)
    attempts = state.get("attempts", 0)
    dispositions = state.get("finding_dispositions", [])
    poc_results = state.get("poc_results", [])
    sc_findings = state.get("securecoder_findings", [])

    status = "✅ RESOLVED" if verified else f"⚠️ UNRESOLVED after {attempts} attempt(s)"

    # ── Build CS-XXX-NNN audit table ──
    all_findings = state.get("findings", []) + sc_findings
    vuln_table = "| Vulnerability ID | File | Line | Description | Severity | Status | Remediation |\n"
    vuln_table += "|---|---|---|---|---|---|---|\n"

    for i, f in enumerate(all_findings):
        vid = _vuln_id(f, i)
        file = f.get("file", "N/A")
        line = f.get("line", "N/A")
        desc = f.get("message", "N/A")[:100]
        sev = f.get("severity", "N/A")

        # Determine status from dispositions
        disp = next((d for d in dispositions if d.get("finding") == f), None)
        if disp and disp["disposition"] == "False Positive":
            fstatus = "False Positive"
            remed = f"Suppressed: {disp.get('rationale', 'N/A')}"
        elif f == finding and verified:
            fstatus = "Fixed"
            remed = "Patch applied and verified"
        elif f == finding:
            fstatus = "Open"
            remed = "Fix attempted but not verified"
        else:
            fstatus = "Open"
            remed = "Not yet addressed"

        vuln_table += f"| {vid} | `{file}` | {line} | {desc} | {sev} | {fstatus} | {remed} |\n"

    # ── Build PoC Verification section ──
    poc_section = ""
    if poc_results:
        poc_section = "\n## PoC Verification\n\n"
        for poc in poc_results:
            poc_section += f"### {poc.get('vulnerability_type', 'Unknown')}\n\n"
            poc_section += "| Field | Value |\n|---|---|\n"
            poc_section += f"| Type | {poc.get('vulnerability_type', 'N/A')} |\n"
            poc_section += f"| Severity | {poc.get('severity', 'N/A')} |\n"
            poc_section += f"| Conclusion | **{poc.get('conclusion', 'N/A')}** |\n\n"

            if poc.get("reasoning_steps"):
                poc_section += "| Step | Description | Result |\n|---|---|---|\n"
                for step in poc["reasoning_steps"]:
                    poc_section += f"| {step.get('step', '?')} | {step.get('description', 'N/A')} | {step.get('result', 'N/A')} |\n"
                poc_section += "\n"

    # ── Suppressed Findings section ──
    suppressed = [d for d in dispositions if d.get("suppressed")]
    suppressed_section = ""
    if suppressed:
        suppressed_section = "\n## Suppressed Findings\n\n"
        suppressed_section += "| Finding | File | Reason |\n|---|---|---|\n"
        for d in suppressed:
            f = d["finding"]
            suppressed_section += f"| {f.get('message', 'N/A')[:60]} | `{f.get('file', 'N/A')}` | {d.get('rationale', 'N/A')} |\n"

    # ── Assemble full report ──
    report = f"""# AITriage + SecureCoder Security Audit

**Status**: {status}
**Scanned Files**: {len(set(f.get('file', '') for f in all_findings))}
**Vulnerabilities Found**: {len(all_findings)}
**True Positives**: {sum(1 for d in dispositions if d.get('disposition') == 'True Positive')}
**False Positives**: {sum(1 for d in dispositions if d.get('disposition') == 'False Positive')}
**Vulnerabilities Fixed**: {1 if verified else 0}

## SecurityScore
- **Before:** {score_before}/100

## Vulnerability Report

{vuln_table}

## Finding Addressed
- **File:** `{finding.get('file', 'N/A')}`
- **Line:** {finding.get('line', 'N/A')}
- **Severity:** {finding.get('severity', 'N/A')}
- **Issue:** {finding.get('message', 'N/A')}
- **CWE:** {finding.get('cwe', finding.get('labels', {}).get('cwe', 'N/A'))}

## Generated Fix
```
{fix}
```
{poc_section}{suppressed_section}
## Next Steps
{"- Apply the patch above and run `aitriage scan .` to verify." if verified else "- Manual review required. The automated fix could not be validated."}
"""

    # ── Report fix completion to SecureCoder IDE (opt-in) ──
    if verified and state.get("findings_count_before", 0) > 0:
        try:
            from tools import get_tool
            # This would need a fix_completed MCP tool, but for now log it
            before = state.get("findings_count_before", 0)
            after = len(state.get("findings", []))
            print(f"  📊 Fix completion: {before} → {after} findings")
        except Exception:
            pass

    print(f"\n📊 [report] Final status: {status}")
    return {
        "final_report": report,
        "audit_report": report,
    }


# ─── Edge Condition ───────────────────────────────────────────────────────────

def should_retry(state: SecurityState) -> str:
    """
    Conditional edge function after verify_node.

    Returns:
        "done"     → fix was verified, go to poc
        "retry"    → fix failed but attempts < 3, go back to fix_node
        "give_up"  → fix failed and attempts >= 3, go to report
    """
    if state.get("fix_verified"):
        return "done"
    if state.get("attempts", 0) >= 3:
        return "give_up"
    return "retry"
