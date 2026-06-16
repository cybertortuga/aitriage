# AITriage Final Audit Checklist

Based on the original `task.md`, mapped against the current implementation state.

## 1. Core Scanners (Инструменты бота -> Сканеры)
- [x] semgrep (Implemented in `external_scanners.go` stub)
- [x] bandit (Implemented in `external_scanners.go` stub)
- [x] trivy (Implemented in `external_scanners.go` stub)
- [x] Ensure proper parallel execution and real tool bindings.

## 2. NFR Checks (NFR)
- [x] Pre-audit capability via static rules (OWASP Core)
- [x] Compliance with basic project structures (Missing lockfiles, missing dockerfiles detected)

## 3. Infrastructure Audit (Аудит инфры)
- [x] IaC checking (Docker/Makefiles detected via rules)
- [x] Network probing (Ports/Domains) - `internal/network/probe.go` needs to be linked to orchestration.

## 4. Architecture & Secrets (Изучение архитектуры)
- [x] Secrets detection (Basic logic in AST stripper)
- [x] Deployment files detection (Dockerfile, compose)
- [x] Git analysis (Entropy, critical files) - `internal/entropy/git_analyzer.go`

## 5. Requirements (Требования)
- [ ] Natural language understanding (LLM orchestration pending full LangGraph integration)
- [x] GUI (Web dashboard / HTML reports available)
- [x] Give instructions for AI-agents to fix (Implemented via JSON/SARIF and Remedy logic)
- [x] AI-lab agnostic (Implemented generic AI interfaces)
- [x] Easy distribution (Go binary)

## 6. Orchestration Workflows (Воркфлоу тулзы)
- [x] Parallel scanning (Engine runs concurrent checks)
- [x] Consolidate artifacts (Reporters)
- [x] Recommendations for analysis (Score/Grades)
- [ ] Interactive agentic Q&A

## 7. Immediate Fixes & De-cringification (Current Sprint)
- [x] Rename `SecurityScore` -> `SecurityScore` across codebase.
- [x] Rename `SecurityGrade` -> `SecurityGrade`.
- [x] Replace `HYPE KILLED` with `AUDIT FAILED (CRITICAL)`.
- [x] Replace `VIBE CHECK PASSED` with `AUDIT PASSED (CLEAN)`.
- [x] Redesign HTML Report (`dashboard.go`) to "Silent Luxury" / Premium AAA Vogue style.
- [x] Rename package `internal/entropy` to `internal/entropy`.
