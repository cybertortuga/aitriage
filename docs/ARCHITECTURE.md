# Architecture: DashDecode Paradigm

AITriage 2.0 is built on the **DashDecode** philosophy: High-cadence security "decoding" delivered through a "Premium Dashboard" experience. This document outlines the technical layers that transform raw code into actionable security intelligence.

## 1. The Decoder (Core Engine)
The heart of AITriage is a deterministic, high-performance engine written in Go.

- **AST-Native**: Unlike legacy linters that rely on regex, AITriage uses **Tree-sitter** to parse code into a concrete syntax tree. This allows for precise matching of logical structures (e.g., function calls, variable assignments) rather than just character sequences.
- **Spectral Analysis (Entropy-Checks)**:
    - **Entropy Detection**: Identifying high-entropy strings that suggest embedded secrets or tokens.
    - **AI Intensity**: Analyzing git history and comments to identify the "Entropy" of the commit (e.g., detecting if a file was generated as an "AI monolithic dump").
- **Taint-Lite**: Tracking the flow of data from sources (parameters, request bodies) to sinks (eval, db.query) without the overhead of full data-flow analysis.

## 2. The Agentic Layer (The Brain)
To avoid "Triage Slop" (meaningless AI alerts), AITriage employs an agentic layer for validation.

- **Contextual Reasoning**: When a rule is triggered, the Agentic Orchestrator pulls the surrounding code block and provides a "Confidence Score" and a "Reasoning Sentence."
- **False Positive Elimination**: The agent compares the finding against known "safe entropys" (e.g., test mocks or authenticated internal utilities).

## 3. The Dashboard (The Face)
The results are delivered via a high-end HTML/JS interface designed for stakeholders.

- **Aesthetic First**: A dark-mode, glassmorphic UI that provides instant visibility into project health.
- **Security Grade**: A high-level metric (A-F) that combines hard security findings with soft "entropy" metrics (coding chaos, technical debt).

## 4. The Remedy Hand
AITriage doesn't just complain; it fixes.
- **Shadow Patching**: The `remedy` engine suggests or automatically applies patches to address identified failures, maintaining the "entropy" of the original developer while hardening the code.

---
*AITriage: Fast enough for Entropy-Coding. Safe enough for Production.*
