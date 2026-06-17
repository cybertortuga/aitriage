# Health Check / CI-CD Migration Plan

Date: 2026-06-17

## Goal

Make Health Check the authoritative repository security posture score without
breaking existing CI/CD, API, SARIF, TUI, history, telemetry, server, or web
consumers.

The migration must be conservative:

- Keep `security_score` and `security_grade` JSON fields for compatibility.
- Add/propagate `health_check` as the richer explanation object.
- Gate CI on active findings and Health Check thresholds, not on suppressed
  or false-positive findings.
- Avoid mass renames unless every consumer is updated in the same phase.

## IB Policy Gate Update

The CI/CD decision is now represented by `health_check.verdict.passed`.
`security_score`, `security_grade`, and `has_critical_failures` remain
compatibility aliases for existing consumers.

Runtime policy sources, in priority order:

- explicit CLI/action flags: `--health-profile`, `--fail-on`, `--fail-score`
- `.aitriage.yaml` `health_check:` block
- legacy `.aitriage.yaml` `strict_mode` / `fail_score`
- built-in defaults

`scan` defaults to a blocking baseline gate (`fail_on=critical`). `agent`
defaults to advisory/non-blocking when no policy is configured, but uses the same
verdict engine once a config or explicit gate flag is present.

## External Research Notes

- GitHub code scanning expects stable SARIF rule IDs and uses fingerprints to
  match alerts across runs. Result `ruleId`, location, and `partialFingerprints`
  matter for duplicate prevention.
  Source: https://docs.github.com/en/enterprise-cloud@latest/code-security/reference/code-scanning/sarif-files/sarif-support
- GitHub SARIF upload can calculate fingerprints when source code is present,
  but producers get better consistency when they emit stable rule/location data.
  Source: https://docs.github.com/en/code-security/how-tos/find-and-fix-code-vulnerabilities/integrate-with-existing-tools/upload-sarif-file
- GitHub job summaries are Markdown appended to `$GITHUB_STEP_SUMMARY`, are
  isolated per step, and have a 1 MiB per-step limit. Summary output should stay
  compact and avoid dumping huge raw finding sets.
  Source: https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands
- OWASP risk rating separates likelihood and impact and warns that business
  context can change severity. Health Check should therefore expose a breakdown
  and not pretend a single number is the full risk model.
  Source: https://owasp.org/www-community/OWASP_Risk_Rating_Methodology

## Current Working Tree State

Already present in the working tree:

- New package: `internal/report/healthcheck`.
- Legacy `internal/report/scorer` removed, OWASP mapping moved under
  `healthcheck`.
- `scanner.Scan` now computes core-only Health Check and serializes
  `health_check`.
- Agent graph computes multi-source Health Check after threat-model
  dispositions.
- CI gate labels and terminal output now say Health Check in the main scan path.
- Additional hardening already added:
  - Critical/high findings suppress positive-practice bonus.
  - Dedup prefers active findings over ignored duplicates.
  - `scan --fail-on any` and `strict_mode` use active finding count.
  - GitHub annotations and summary skip suppressed core findings.
  - Deploy/network findings get stable IDs for FP matching.

## Impact Map

### CLI / GitHub Action

Files:

- `cmd/aitriage/scan.go`
- `cmd/aitriage/agent.go`
- `cmd/aitriage/watch.go`
- `cmd/aitriage/init.go`
- `action.yml`
- `entrypoint.sh`
- `scripts/entrypoint.sh`

Risks:

- `fail_score` semantics changed because Health Check uses saturation and
  positive bonuses.
- `fail-on any` must mean active findings, not suppressed findings.
- The Docker Action currently exposes `fail-on` but not first-class
  `fail-score`; users can still pass it through `args`.
- `agent --fail-on` defaults to `never`, while `scan` defaults to `critical`.
  This is acceptable only if documented as "AI advisor is non-blocking by
  default".

Required checks:

- `aitriage scan --format json --no-summary`
- `aitriage scan --format sarif --out /tmp/aitriage.sarif --no-summary`
- fail-on behavior with active vs ignored findings where feasible in tests.

### Scanner / Report Contract

Files:

- `internal/scanner/scanner.go`
- `internal/engine/core/context.go`
- `internal/report/healthcheck/*`
- `internal/report/reporter/reporter.go`
- `internal/report/reporter/reporter_test.go`

Risks:

- Existing consumers expect `security_score` and `security_grade`.
- New `health_check` must remain additive.
- `HasCriticalFailures` must remain based on active CRITICAL/HIGH findings.
- Unknown severities should not silently become critical; they currently get a
  small default weight.

Required checks:

- Health Check unit tests for clean repos, FP/ignored, dedup, bonus suppression,
  source weighting, grade thresholds.
- JSON snapshot sanity check includes both legacy score fields and
  `health_check.breakdown`.

### SARIF

Files:

- `internal/report/reporter/reporter.go` used by `aitriage scan --format sarif`.
- `internal/scanner/sarif.go` used by TUI export.
- `internal/ui/tui/update.go` calls `ScanReport.ToSARIF()`.

Risks:

- There are two SARIF implementations with different behavior.
- `internal/scanner/sarif.go` appears to have status filtering that does not
  match the CLI exporter and needs a focused audit before unification.
- GitHub code scanning benefits from stable rule IDs, relative URIs, valid line
  numbers, and optionally `partialFingerprints`.
- Suppressed/ignored findings should not be uploaded as active alerts unless
  there is an explicit "include suppressed" mode.

Required checks:

- Compare CLI SARIF and TUI SARIF for the same mock report.
- Ensure ignored/triage core findings are omitted from both exporters.
- Consider adding `partialFingerprints` only after verifying compatibility with
  GitHub's supported SARIF subset.

### Agent Graph

Files:

- `cmd/aitriage/agent.go`
- `internal/agent/graph/orchestrator.go`
- `internal/agent/graph/state.go`
- `internal/agent/prompts/templates.go`

Risks:

- Current Health Check is computed after threat-model dispositions, before
  map-reduce narrative triage. This is acceptable if threat-model
  `finding_dispositions` is the authoritative TP/FP source.
- Threat-model prompt currently caps findings sent to the LLM at 20. Findings
  beyond that cap get no disposition and remain active. This is safer for CI,
  but should be documented as conservative behavior.
- AI output is not deterministic enough for hard CI by default; `agent` should
  remain non-blocking unless users explicitly pass `--fail-on` or `--fail-score`.

Required checks:

- Unit test deploy/network FP matching.
- Unit test Health Check remains conservative when no dispositions are present.
- Verify `agent --no-chat` exits after writing artifacts and only then applies
  explicit CI gate.

### Server / Enterprise UI / Web

Files:

- `internal/server/server.go`
- `internal/server/repositories/metrics_repo.go`
- `internal/server/ui/*`
- `web/src/types.ts`
- `web/src/services/securityService.ts`
- `web/src/hooks/useMetrics.ts`
- Dashboard pages under `web/src/pages`.

Risks:

- Server scan response currently returns only `security_score` and
  `security_grade`, not `health_check`.
- Enterprise metrics repository computes a separate score directly from DB
  severity counts. That is a different domain-level product score, not the same
  scan-level Health Check. Do not silently swap formulas without a separate
  migration.
- Frontend types do not know about `health_check`.

Required checks:

- Add `health_check` to scan response only as additive API data.
- Keep existing score displays working.
- If UI copy is updated, label display as "Health Check" while keeping field
  names backwards-compatible.

### History / Telemetry / MCP

Files:

- `internal/engine/history/history.go`
- `internal/telemetry/*`
- `internal/agent/mcp/tools_history.go`
- `internal/agent/mcp/tools_entropy.go`

Risks:

- These components still use `SecurityScore` naming. This is acceptable as
  internal/backward-compatible naming, but user-facing strings should gradually
  move to Health Check.
- Historical diffs before and after formula migration are not apples-to-apples.

Required checks:

- Do not rewrite historical records.
- Add copy that score trend may reflect the Health Check formula after this
  release if release notes are updated.

### Documentation

Files:

- `README.md`
- `docs/aitriage.yaml.example`
- `docs/STRUCTURE.md`
- `docs/root_architecture.md`
- `agent/README.md`
- `docs/INTEGRATION*.md`

Risks:

- Docs still reference `scorer` and Security Score.
- Some older integration plans are historical and should not be mass-edited.

Required checks:

- Update active user-facing docs: README, `docs/aitriage.yaml.example`, and
  generated config comments.
- Leave historical planning docs alone unless they actively confuse current
  setup instructions.

## Work Phases

### Phase 0 - Safety Baseline

- Confirm current tree and tests before more edits.
- Keep all changes unstaged until review.
- Do not rename public JSON fields.

Exit criteria:

- `go test -p 1 ./...`
- `go vet ./...`

### Phase 1 - Health Check Engine Hardening

- Keep Health Check package focused and source-agnostic.
- Ensure dedup, ignored handling, bonus suppression, grade thresholds, and
  critical/high flags are tested.
- Keep legacy adapter for core-only scan path.

Exit criteria:

- `go test ./internal/report/healthcheck`

### Phase 2 - CI/CD Gate Alignment

- Gate on active findings.
- Keep `scan` default blocking behavior: `--fail-on critical`.
- Keep `agent` default non-blocking behavior: `--fail-on never`.
- GitHub annotations and summary must show active findings only.

Exit criteria:

- `go test ./cmd/aitriage`
- Manual CLI smoke for JSON and SARIF output.

### Phase 3 - SARIF Consistency

- Audit both SARIF exporters.
- Either unify TUI export with CLI exporter or make filtering behavior identical.
- Add tests for ignored/triage omission.

Exit criteria:

- Reporter SARIF tests pass.
- Scanner/TUI SARIF path tested or covered through shared exporter.

### Phase 4 - API/UI Additive Propagation

- Add `health_check` to server scan response as optional/additive.
- Update TypeScript types to accept it.
- Update visible labels only where safe and small.

Exit criteria:

- Go tests pass.
- Web type/build check if package scripts are available.

### Phase 5 - Documentation and Release Notes

- Update active docs only.
- Explain `fail_score` threshold behavior changed and may need recalibration.
- Explain `security_score` JSON is retained as compatibility alias for Health
  Check score.

Exit criteria:

- README and config example no longer contradict CLI behavior.

## Open Decisions

1. Should `action.yml` expose `fail-score` as a first-class input, or keep it in
   `args` for this release?
2. Should Health Check include external/NFR/deploy/network in `scan`, or only in
   `agent`? Current implementation keeps `scan` core-only and `agent`
   multi-source.
3. Should SARIF include `partialFingerprints` now, or defer until both exporters
   are unified?
4. Should enterprise DB metrics adopt Health Check formula, or remain a separate
   portfolio/product risk score?

## Final Verification Checklist

- `go test ./internal/report/healthcheck ./internal/agent/graph ./internal/scanner ./cmd/aitriage`
- `go test -p 1 ./...`
- `go vet ./...`
- `aitriage scan <fixture> --format json --no-summary`
- `aitriage scan <fixture> --format sarif --out /tmp/aitriage.sarif --no-summary`
- Validate generated SARIF JSON parses and contains only active findings.
- Review `git diff --stat` and `git diff --name-status` before final response.
