# AITriage Web: Simple Mode, SecureCoder and Runway Reports

## Goal

Make Simple mode understandable without security or CI/CD expertise: the primary remediation path must be obvious, completed Runway runs must expose the same useful output as GitHub Actions, and the interface must stay inside the existing AITriage visual language.

## Problems found

- SecureCoder competed visually with secondary navigation and project scanning.
- The reports entry point was easy to miss and the saved AI summary had weak hierarchy.
- Completed Runway output was fragmented across transient UI state instead of durable artifacts.
- Web did not expose the canonical `summary.md`, AI remediation prompt, or structured agent data used by CI/CD.
- Findings rows had small click targets, weak expanded-state hierarchy, and motion that affected too much of the page.
- The SecureCoder drawer used dense rows, inconsistent emphasis, and insufficient state feedback.
- Simple mode surfaces used fixed near-black values, so changing the background palette did not visibly affect the page.

## Implementation plan and completion checklist

### 1. Canonical Runway artifacts

- [x] Add a persistent Runway artifact model and repository.
- [x] Persist `summary.md`, remediation prompt, structured agent data, report, fix specification, and triage findings.
- [x] Store media type, schema version, SHA-256, byte size, and creation time.
- [x] Reject unsupported artifact kinds and oversized payloads.
- [x] Generate one deterministic AI handoff for both Web and GitHub Actions.
- [x] Exclude false positives from actionable agent data while retaining the complete audit trail.

### 2. Web API and Runway lifecycle

- [x] Add artifact manifest and download endpoints.
- [x] Add a verified handoff endpoint for the AI remediation prompt and structured agent data.
- [x] Persist artifacts when a Runway finishes.
- [x] Keep Runway progress and failure state durable.
- [x] Preserve the existing export path for compatibility.

### 3. Reports as a primary product path

- [x] Move **Reports** to the second position in Simple navigation and give it a consistent icon.
- [x] Add a clear completed-run result with Web preview and `summary.md` download.
- [x] Present Threat Model, Security Plan, Remediation, Verification, and Audit Report in one report flow.
- [x] Add expandable **AI Remediation Prompt** and **AI Agent Data** sections.
- [x] Add copy and direct artifact download actions with loading, empty, and error states.
- [x] Add Russian and English product copy explaining where the final CI/CD-equivalent report lives.

### 4. Simple mode redesign

- [x] Rebuild the overview hierarchy around repository risk, saved AI summary, vulnerability queue, and the primary SecureCoder action.
- [x] Make findings rows readable and fully expandable with explicit button semantics.
- [x] Keep severity and status colors semantic while routing product accents through theme tokens.
- [x] Replace broad layout-changing animations with short opacity/position transitions scoped to the opened content.
- [x] Respect reduced-motion preferences.
- [x] Increase interactive target clarity, focus treatment, text contrast, truncation behavior, and long-path handling.
- [x] Add loading, empty, error, selected, expanded, disabled, and completed states.

### 5. SecureCoder menu

- [x] Make SecureCoder the visually primary remediation instrument.
- [x] Rebuild the drawer as a stable side panel with clear integration status and grouped actions.
- [x] Keep project scanning secondary to SecureCoder remediation.
- [x] Add controlled accordion states for target scan, dependencies, ignored findings, and configuration.
- [x] Prevent the trigger from animating or reflowing the entire screen.
- [x] Restore focus to the trigger on close and support Escape/outside interaction behavior.

### 6. Theme and palette integration

- [x] Remove the fixed Simple-mode black background layer.
- [x] Derive the Simple page base and surface hierarchy from global background and surface tokens.
- [x] Keep selected accent and background palettes independent and visible.
- [x] Preserve fixed red/orange/yellow/green meanings for risk and lifecycle status.

## User flow

1. Open **Overview** and select or scan a repository.
2. Use **SecureCoder** as the primary remediation entry point.
3. Complete the Runway workflow.
4. Open **Reports**, now the second navigation item.
5. Open the full Web result or download the canonical `runway-{id}-summary.md`.
6. Expand and copy **AI Remediation Prompt** or **AI Agent Data** into Cursor, Claude, Antigravity, or another AI IDE.

## Validation

- [x] `go test ./...`
- [x] `npm run build` in `web/`
- [x] `git diff --check`
- [x] Docker Web container rebuilt and reported healthy during local verification.
- [x] `/api/health` returned `ok: true` with all scanner tools available.
- [x] Deployed CSS bundle contained the theme-driven Simple background token.

The Vite build still reports the existing large-chunk advisory; it does not fail the production build and is outside this feature's functional scope.

## Release

The completed implementation, tests, and this work log are committed together and pushed directly to `origin/main` as the final delivery step.
