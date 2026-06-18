# AITriage V2.1: The "Adult" Roadmap

This document outlines the results of the Deep Audit of the AITriage architecture and defines the immediate roadmap to transform the project from a "Poc/Teenager" state to a robust, production-ready "Adult" tool.

## 1. Executive Summary of the Audit

AITriage has successfully evolved past its TypeScript MVP. The transition to Go, the introduction of the context-aware `Workspace`/`ProjectContext`, and the `DashDecode` philosophy represent a massive leap. The foundational engine is fast, parallelized, and data-driven (YAML configurations). 

However, the rapid iteration has introduced "identity crises" and architectural gaps that must be addressed before this tool can be trusted in CI/CD pipelines.

### The Identity Crisis: Presence-Checker vs. Vulnerability-Scanner
Originally, AITriage was a **Presence Checker** ("Is this practice missing?"). The new YAML rules have transformed it into a **Vulnerability Scanner** ("Is this specific hacky Entropy/AST pattern present?"). 

**The Bug:** The engine currently tries to serve both paradigms with the same output schema. `engine.go` generates a `core.CheckResult` with `Status: Present, Suggestion: "Good Job", Severity: CRITICAL` when a vulnerability is *not* found. Because the Dashboard just loops over all `Results`, the user sees terrifying red "CRITICAL" cards that actually tell them "Good Job".

## 2. Deep Technical Audit Findings

### 🔴 Critical Gaps (Toddler-Level Issues)
1.  **Rule Result Noise:** The Engine returns "Good Job" findings for passed vulnerability checks, polluting the final report. The report must contain ONLY `Violations` / `Alerts`, or cleanly separate `Passed` from `Failed` checks.
2.  **Ignored Errors:** `content, _ := f.GetContent()`. Ignoring file/read errors in a security scanner is a fundamental flaw. If a file cannot be read, the user must be warned, otherwise it's a silent false negative.
3.  **Taint Analysis Illusion:** `runTaintAnalysis` uses highly brittle regexes (`req.body` -> variable -> `eval`). It is easily bypassed by destructuring or simple function calls.

### 🟡 High-Priority Refactoring (Teenager-Level Issues)
1.  **Stripper Weaknesses:** `internal/entropy/stripper.go` is a custom state machine. It breaks on Javascript regex literals (`const regex = /foo/`) thinking they are comments. Given we *already* use `tree-sitter` in `engine.go`, we should migrate the Stripper to use AST for 100% precision instead of regex.
2.  **Linear Lookups:** `ProjectContext.GetFile()` implements an $O(N)$ linear search. With 10,000 files, this introduces unnecessary lag.
3.  **YAML Schema Ambiguity:** The `target` fields (`code`, `docs`, `ast`, `lines_percentage`, `filename`) lack strict enforcement. 

### 🟢 Solid Foundations (Adult-Level Architecture)
1.  **Parallel Execution:** The `Engine.Run` with `sync.WaitGroup` prevents IO-blocking and is horizontally scalable.
2.  **Reporting Aesthetics:** The HTML/CSS in `dashboard.go` perfectly matches the premium specification defined in `DESIGN.md`.
3.  **Caching:** `FileInfo` caches `rawCache`, `strippedCache`, and `docsCache` safely with mutexes.

---

## 3. The Roadmap (Immediate Action Plan)

We will execute this plan sequentially to stabilize and elevate the core engine.

### Phase 1: Engine Sanity & Noise Reduction (Expected: Immediate)
*   [x] **Fix the Result Logic:** Refactor `engine.go:Run`. Differentiate between `PresenceRules` (e.g., missing middleware) and `VulnerabilityRules` (e.g., hardcoded JWTs). Only append `CheckResult` to the report if there is a true VIOLATION. Remove the "Good job" spam.
*   [x] **Dashboard Data Contract:** Ensure `dashboard.go` only renders actual alerts and calculates the `SecurityGrade` correctly based on the volume and severity of alerts.
*   [x] **Error Handling:** Replace `content, _ := ...` with proper error handling and logging using `slog`.

### Phase 2: AST Precision Upgrade (Expected: Next)
*   [x] **Retire State-Machine Stripper:** Rewrite `internal/entropy/stripper.go` to leverage `tree-sitter`. We use queries like `(comment) @c` and `(string) @s` to blank out docs and strings with absolute deterministic accuracy. 
*   [x] **Refine Taint Analysis:** Either remove the fragile regex taint-analysis for now (to avoid selling false security) or rewrite it using `tree-sitter` variable tracking. 

### Phase 3: Monorepo & Performance Polish (Expected: Later)
*   [x] **$O(1)$ File Lookups:** Upgrade `core.Workspace` and `ProjectContext` to hold a `map[string]*FileInfo` for instant fetching via `GetFile()`.
*   [x] **Security Score Formula:** Calibrate the metrics formula so that `SecurityGrade` is dynamically calculating A-F based on the rule weights.

### Phase 4: Big-O Architecture Inversion (Expected: Immediate)
*   [x] **O(N) Single-Pass Scanner:** `engine.go:Run` currently spawns Goroutines per RULE and loops over N files inside each rule (`O(N*M)` execution). This causes massive GC spam and redundant IO checks. Refactor to loop over Files ONCE (`O(N)`) and apply all applicable rules per file.

### Phase 5: Production Maturity & Intelligence
*   [x] **Smart GitIgnore:** Stop hardcoding `node_modules` and `venv`. The scanner must natively parse `.gitignore` and skip excluded directories to prevent crashing on large build artifacts.
*   [x] **Real Agentic Provider:** ~Remove `MockProvider` and implement a genuine minimal REST Client to OpenAI/Gemini to enable true LLM-based False Positive suppression.~ *(Deferred to `docs/TODO.md`)*

## 4. Next Steps
Executing Phase 4 immediately. Inverting the `O(N*M)` rule-file loop to a deterministic `O(N)` pipeline.
