package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	agentcontext "github.com/cybertortuga/aitriage/internal/agent/context"
	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/agent/prompts"
	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
)

// Run Orchestrates the full SecureCoder-enhanced pipeline:
//
//	enrichFindings → buildThreatModel → runPoCVerification →
//	computeHealthCheck → generateReport → generateSummary → generateAIFixSpec
func Run(ctx context.Context, state *AgentState, llmClient llm.Client) error {
	// Step 0: Gather repository context (reads files from disk, no LLM)
	fmt.Fprintf(os.Stderr, "📂 Gathering Repository Context...\n")
	gatherRepoContext(state)

	fmt.Fprintf(os.Stderr, "🤖 Context Enrichment...\n")
	enrichFindings(state)

	// SecureCoder Step 1: Threat Model (classifies each finding as TP/FP/NR)
	fmt.Fprintf(os.Stderr, "🏗️ Building Threat Model (SecureCoder)...\n")
	if err := buildThreatModel(ctx, state, llmClient); err != nil {
		return fmt.Errorf("threat-model analysis failed: %w", err)
	}

	// SecureCoder Step 2: PoC Verification (proves exploitability of True Positives)
	fmt.Fprintf(os.Stderr, "🧪 PoC Verification (SecureCoder)...\n")
	if err := runPoCVerification(ctx, state, llmClient); err != nil {
		return fmt.Errorf("PoC verification failed: %w", err)
	}

	// Health Check: compute AFTER all triage is complete (Threat Model + PoC).
	// This ensures the CI gate verdict uses the final, authoritative dispositions.
	fmt.Fprintf(os.Stderr, "🩺 Computing Security Health Check (all sources, FP-aware)...\n")
	computeHealthCheck(state)
	fmt.Fprintf(os.Stderr, "   ✅ Health Check: %d/100 (%s) — %d active, %d ignored (FP), %d deduped\n",
		state.HealthCheck.Score, state.HealthCheck.Grade,
		state.HealthCheck.Breakdown.ActiveFindings,
		state.HealthCheck.Breakdown.IgnoredFindings,
		state.HealthCheck.Breakdown.DedupedFindings)

	fmt.Fprintf(os.Stderr, "🤖 Generating Security Report (CS-XXX-NNN format)...\n")
	if err := generateReport(ctx, state, llmClient); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "🤖 Generating AI Fix Specification...\n")
	if err := generateAIFixSpec(ctx, state, llmClient); err != nil {
		return err
	}

	// Generate the final summary only after every required AI stage has finished,
	// so its usage and findings cannot be partial or pre-triage.
	fmt.Fprintf(os.Stderr, "📋 Generating Actionable Summary (TP/NR only)...\n")
	generateSummary(state)

	// Print LLM usage summary
	u := state.TotalUsage
	if u.TotalTokens > 0 {
		fmt.Fprintf(os.Stderr, "\nLLM usage (provider reported): %s. Cost is not estimated because it depends on provider, model, caching, and billing tier.\n",
			formatLLMUsage(u))
	}

	return nil
}

// gatherRepoContext reads the repository from disk and builds structured context.
func gatherRepoContext(state *AgentState) {
	state.RepoContext = agentcontext.BuildRepoContext(state.ProjectPath)

	keyCount := 0
	if state.RepoContext != nil {
		keyCount = len(state.RepoContext.KeyFiles)
	}
	fmt.Fprintf(os.Stderr, "   ✅ Tree built, %d key files read, stack: %s\n",
		keyCount, state.RepoContext.Stack)
}

func enrichFindings(state *AgentState) {
	var enriched []EnrichedFinding

	for _, f := range state.CoreFindings {
		enriched = append(enriched, EnrichedFinding{
			ID:       f.ID,
			Type:     "core",
			Source:   "aitriage",
			Severity: f.Severity,
			File:     f.File,
			Line:     f.Line,
			Message:  fmt.Sprintf("%s: %s", f.Name, f.Evidence),
			Snippet:  extractFullContext(state.ProjectPath, f.File, f.Line),
		})
	}
	for _, f := range state.ExternalFindings {
		enriched = append(enriched, EnrichedFinding{
			ID:       f.RuleID,
			Type:     "external",
			Source:   f.Source,
			Severity: f.Severity,
			File:     f.File,
			Line:     f.Line,
			Message:  fmt.Sprintf("[%s] %s", f.Source, f.Message),
			Snippet:  extractFullContext(state.ProjectPath, f.File, f.Line),
		})
	}
	// Add other findings without snippets or with basic info
	for _, f := range state.NFRFindings {
		enriched = append(enriched, EnrichedFinding{
			ID:       f.RuleID,
			Type:     "nfr",
			Source:   "aitriage",
			Severity: f.Severity,
			Message:  fmt.Sprintf("%s: %s", f.Name, f.Message),
		})
	}
	for _, f := range state.DeployFindings {
		enriched = append(enriched, EnrichedFinding{
			ID:       f.Issue,
			Type:     "deploy",
			Source:   "aitriage",
			Severity: f.Severity,
			File:     f.File,
			Line:     f.Line,
			Message:  fmt.Sprintf("%s. Advice: %s", f.Issue, f.Advice),
			Snippet:  extractFullContext(state.ProjectPath, f.File, f.Line),
		})
	}
	for _, f := range state.NetworkFindings {
		enriched = append(enriched, EnrichedFinding{
			ID:       fmt.Sprintf("port-%d", f.Port),
			Type:     "network",
			Source:   "aitriage",
			Severity: f.Severity,
			Message:  fmt.Sprintf("Port %d (%s): %s", f.Port, f.Service, f.Message),
		})
	}

	// Assign CS-XXX-NNN vulnerability IDs
	assignVulnIDs(enriched)

	state.EnrichedFindings = enriched

	// Map into batches of 5
	var targetFindings []EnrichedFinding
	if len(enriched) > 50 {
		for _, e := range enriched {
			if strings.ToUpper(e.Severity) == "HIGH" || strings.ToUpper(e.Severity) == "CRITICAL" {
				targetFindings = append(targetFindings, e)
			}
		}
	} else {
		targetFindings = enriched
	}

	chunkSize := 5
	for i := 0; i < len(targetFindings); i += chunkSize {
		end := i + chunkSize
		if end > len(targetFindings) {
			end = len(targetFindings)
		}
		state.Batches = append(state.Batches, targetFindings[i:end])
	}
}

// computeHealthCheck recomputes the authoritative Security Health Check across ALL
// scanner sources (core, external, NFR, deploy, network) and applies AI triage
// dispositions: findings classified as False Positive are excluded from the
// penalty. The result becomes the canonical SecurityScore/SecurityGrade.
func computeHealthCheck(state *AgentState) {
	// Build the set of False-Positive locations from AI dispositions.
	fp := make(map[string]bool)
	for _, d := range state.FindingDispositions {
		if d.Disposition != "False Positive" {
			continue
		}
		if d.FindingIndex >= 0 && d.FindingIndex < len(state.EnrichedFindings) {
			ef := state.EnrichedFindings[d.FindingIndex]
			fp[hcKey(ef.ID, ef.File, ef.Line)] = true
		}
	}

	in := healthcheck.Input{}

	for _, r := range state.CoreFindings {
		switch r.Status {
		case core.Present:
			in.Positives = append(in.Positives, healthcheck.Positive{ID: r.ID})
		case core.Absent:
			ignored := r.AuditStatus == core.AuditStatusIgnored ||
				r.AuditStatus == core.AuditStatusTriage ||
				fp[hcKey(r.ID, r.File, r.Line)]
			in.Findings = append(in.Findings, healthcheck.Finding{
				Source:   "core",
				Class:    r.ID,
				Severity: r.Severity,
				File:     r.File,
				Line:     r.Line,
				Ignored:  ignored,
			})
		}
	}
	for _, f := range state.ExternalFindings {
		in.Findings = append(in.Findings, healthcheck.Finding{
			Source:   f.Source,
			Class:    f.RuleID,
			Severity: f.Severity,
			File:     f.File,
			Line:     f.Line,
			Ignored:  fp[hcKey(f.RuleID, f.File, f.Line)],
		})
	}
	for _, f := range state.NFRFindings {
		in.Findings = append(in.Findings, healthcheck.Finding{
			Source:   "nfr",
			Class:    f.RuleID,
			Severity: f.Severity,
			Ignored:  fp[hcKey(f.RuleID, "", 0)],
		})
	}
	for _, f := range state.DeployFindings {
		in.Findings = append(in.Findings, healthcheck.Finding{
			Source:   "deploy",
			Class:    f.Issue,
			Severity: f.Severity,
			File:     f.File,
			Line:     f.Line,
			Ignored:  fp[hcKey(f.Issue, f.File, f.Line)],
		})
	}
	for _, f := range state.NetworkFindings {
		class := fmt.Sprintf("port-%d", f.Port)
		in.Findings = append(in.Findings, healthcheck.Finding{
			Source:   "network",
			Class:    class,
			Severity: f.Severity,
			Ignored:  fp[hcKey(class, "", 0)],
		})
	}

	res := healthcheck.ApplyPolicy(healthcheck.Evaluate(in), state.Policy)
	state.HealthCheck = res
	state.SecurityScore = res.Score
	state.SecurityGrade = res.Grade
}

// hcKey builds a location key used to match AI dispositions to findings.
func hcKey(id, file string, line int) string {
	return fmt.Sprintf("%s|%s|%d", strings.ToLower(id), strings.ToLower(file), line)
}

// assignVulnIDs generates CS-XXX-NNN identifiers for each finding.
func assignVulnIDs(findings []EnrichedFinding) {
	counters := make(map[string]int)
	for i := range findings {
		code := classifyVulnCode(findings[i].Message)
		counters[code]++
		findings[i].VulnID = fmt.Sprintf("CS-%s-%03d", code, counters[code])
	}
}

// AssignVulnIDsPublic is the exported version of assignVulnIDs for use by the web pipeline.
func AssignVulnIDsPublic(findings []EnrichedFinding) {
	assignVulnIDs(findings)
}

// classifyVulnCode maps a finding message to a short vulnerability class code.
func classifyVulnCode(message string) string {
	lower := strings.ToLower(message)
	for key, code := range prompts.VulnClassCodes {
		if strings.Contains(lower, key) {
			return code
		}
	}
	return "MISC"
}

// ── Threat Model Step ────────────────────────────────────────────────────────

// threatModelBatchSize caps how many findings are sent to the LLM in a single
// threat-model request to stay within token limits. ALL findings are processed
// by iterating over batches — findings beyond the first batch are never dropped.
const threatModelBatchSize = 150

// threatModelMaxRetries bounds how many extra LLM passes are made to classify
// findings the model omitted in earlier passes before they default to NR.
const threatModelMaxRetries = 2

// nrFallbackRationale is recorded for findings the LLM never classified, even
// after bounded retries. They default to Needs Manual Review (never False
// Positive) so they keep penalising the Health Check score.
const nrFallbackRationale = "LLM did not classify this finding after retries; defaulting to Needs Manual Review for safety."

// errThreatModelParse marks a malformed (unparseable) threat-model response.
// Transport errors are returned unwrapped so they always fail the pipeline,
// whereas a malformed retry response is tolerated and handled by the NR fallback.
var errThreatModelParse = errors.New("parse threat-model JSON")

// rawDisposition is the LLM's unvalidated classification for a single finding,
// indexed relative to the batch that was sent.
type rawDisposition struct {
	FindingIndex int
	Disposition  string
	Confidence   string
	Rationale    string
}

func buildThreatModel(ctx context.Context, state *AgentState, llmClient llm.Client) error {
	if len(state.EnrichedFindings) == 0 {
		fmt.Fprintf(os.Stderr, "   ℹ️ No findings — skipping threat model\n")
		return nil
	}

	repoContextText := ""
	if state.RepoContext != nil {
		repoContextText = state.RepoContext.FormatForLLM(5000) // ~5K tokens for threat model
	}

	tm, dispositions, err := ClassifyFindings(ctx, repoContextText, state.ProjectPath, state.EnrichedFindings, llmClient, &state.TotalUsage)
	if err != nil {
		return err
	}

	state.ThreatModel = tm
	state.FindingDispositions = dispositions

	// Final invariant: every finding has exactly one valid disposition.
	if err := validateFindingDispositions(state.FindingDispositions, len(state.EnrichedFindings)); err != nil {
		return err
	}

	tp, fp, nr := countDispositions(state.FindingDispositions)
	fmt.Fprintf(os.Stderr, "   ✅ Threat model: %d True Positives, %d False Positives, %d Needs Review\n", tp, fp, nr)

	return nil
}

// threatModelLLMCall sends a single batch of findings to the LLM and returns the
// parsed threat model plus the raw (unvalidated) dispositions. Transport errors
// are wrapped plainly; malformed JSON is wrapped with errThreatModelParse.
func threatModelLLMCall(ctx context.Context, repoContextText, projectPath string, batch []EnrichedFinding, llmClient llm.Client, usage *llm.Usage) (*ThreatModel, []rawDisposition, error) {
	findingsJSON, _ := json.MarshalIndent(batch, "", "  ")
	userPrompt := fmt.Sprintf(prompts.ThreatModelUserPromptTemplate,
		repoContextText,
		projectPath,
		len(batch),
		string(findingsJSON),
	)

	messages := []llm.Message{
		{Role: "system", Content: prompts.ThreatModelSystemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, u, err := llmClient.Chat(ctx, messages)
	addUsage(usage, u)
	if err != nil {
		return nil, nil, fmt.Errorf("threat model LLM call failed: %w", err)
	}

	// Parse JSON from response (handle markdown code fences)
	jsonText := extractJSON(response)

	var rawResult struct {
		ComponentOverview   string       `json:"component_overview"`
		EntryPoints         []EntryPoint `json:"entry_points"`
		TrustBoundaries     TrustBounds  `json:"trust_boundaries"`
		SensitiveDataPaths  []DataPath   `json:"sensitive_data_paths"`
		PrivilegedActions   []PrivAction `json:"privileged_actions"`
		PriorityAreas       []string     `json:"priority_areas"`
		FindingDispositions []struct {
			FindingIndex int    `json:"finding_index"`
			Disposition  string `json:"disposition"`
			Rationale    string `json:"rationale"`
		} `json:"finding_dispositions"`
	}

	if err := json.Unmarshal([]byte(jsonText), &rawResult); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", errThreatModelParse, err)
	}

	tm := &ThreatModel{
		ComponentOverview:  rawResult.ComponentOverview,
		EntryPoints:        rawResult.EntryPoints,
		TrustBoundaries:    rawResult.TrustBoundaries,
		SensitiveDataPaths: rawResult.SensitiveDataPaths,
		PrivilegedActions:  rawResult.PrivilegedActions,
		PriorityAreas:      rawResult.PriorityAreas,
	}

	disps := make([]rawDisposition, 0, len(rawResult.FindingDispositions))
	for _, d := range rawResult.FindingDispositions {
		disps = append(disps, rawDisposition{FindingIndex: d.FindingIndex, Disposition: d.Disposition, Rationale: d.Rationale})
	}
	return tm, disps, nil
}

func isSupportedDisposition(d string) bool {
	switch d {
	case "True Positive", "False Positive", "Needs Manual Review":
		return true
	default:
		return false
	}
}

func countDispositions(dispositions []FindingDisposition) (tp, fp, nr int) {
	for _, d := range dispositions {
		switch d.Disposition {
		case "True Positive":
			tp++
		case "False Positive":
			fp++
		default:
			nr++
		}
	}
	return tp, fp, nr
}

func validateFindingDispositions(dispositions []FindingDisposition, findingCount int) error {
	if findingCount == 0 {
		return nil
	}
	if len(dispositions) != findingCount {
		return fmt.Errorf("threat-model response classified %d of %d findings", len(dispositions), findingCount)
	}

	seen := make(map[int]struct{}, findingCount)
	for _, disposition := range dispositions {
		if disposition.FindingIndex < 0 || disposition.FindingIndex >= findingCount {
			return fmt.Errorf("threat-model response has out-of-range finding_index %d", disposition.FindingIndex)
		}
		if _, duplicate := seen[disposition.FindingIndex]; duplicate {
			return fmt.Errorf("threat-model response classifies finding_index %d more than once", disposition.FindingIndex)
		}
		seen[disposition.FindingIndex] = struct{}{}

		switch disposition.Disposition {
		case "True Positive", "False Positive", "Needs Manual Review":
		default:
			return fmt.Errorf("threat-model response has unsupported disposition %q", disposition.Disposition)
		}
	}

	return nil
}

// runWorkers was removed in the pipeline simplification (June 2026).
// The Threat Model step (buildThreatModel) now serves as the single authoritative
// source of TP/FP/NR dispositions. The old Map-Reduce workers duplicated this
// classification with 10+ extra LLM calls and the output (raw markdown) was never
// parsed back into structured FindingDispositions.

// ── PoC Verification Step ────────────────────────────────────────────────────

func runPoCVerification(ctx context.Context, state *AgentState, llmClient llm.Client) error {
	// Collect True Positive findings for PoC
	var tpFindings []EnrichedFinding
	tpSet := make(map[string]bool)
	for _, d := range state.FindingDispositions {
		if d.Disposition == "True Positive" {
			tpSet[d.FindingID] = true
		}
	}

	for _, f := range state.EnrichedFindings {
		if tpSet[f.VulnID] {
			tpFindings = append(tpFindings, f)
		}
	}

	if len(tpFindings) == 0 {
		fmt.Fprintf(os.Stderr, "   ℹ️ No True Positives — skipping PoC verification\n")
		return nil
	}

	// Phase 5b: verify ALL true positives (deduped, batched, bounded concurrency,
	// budget-capped) instead of silently dropping everything past the 75th.
	pocResults, err := verifyPoCs(ctx, tpFindings, llmClient, &state.TotalUsage)
	if err != nil {
		return fmt.Errorf("PoC verification LLM call failed: %w", err)
	}

	state.PoCResults = pocResults

	verified := 0
	incomplete := 0
	for _, p := range pocResults {
		if p.ExploitBlocked != nil && !*p.ExploitBlocked {
			verified++
		} else {
			incomplete++
		}
	}
	fmt.Fprintf(os.Stderr, "   ✅ PoC: %d unique TPs → %d exploitable, %d blocked/unknown\n", len(pocResults), verified, incomplete)

	return nil
}

func generateReport(ctx context.Context, state *AgentState, llmClient llm.Client) error {
	// Generate lookup table for original findings (now with CS-XXX-NNN IDs)
	var lookupLines []string
	lookupLines = append(lookupLines, "| Vulnerability ID | Rule ID | Severity | File | Line |")
	lookupLines = append(lookupLines, "|---|---|---|---|---|")
	for _, f := range state.EnrichedFindings {
		file := f.File
		if file == "" {
			file = "N/A"
		}
		lookupLines = append(lookupLines, fmt.Sprintf("| %s | %s | %s | %s | %d |", f.VulnID, f.ID, f.Severity, file, f.Line))
	}
	lookupTable := strings.Join(lookupLines, "\n")

	// Build threat model summary block
	threatModelBlock := ""
	if state.ThreatModel != nil {
		threatModelBlock = fmt.Sprintf("\n## Threat Model Summary\n- **Component**: %s\n- **Priority Areas**: %s\n",
			state.ThreatModel.ComponentOverview,
			strings.Join(state.ThreatModel.PriorityAreas, ", "))

		if len(state.ThreatModel.EntryPoints) > 0 {
			threatModelBlock += "\n### Entry Points\n"
			for _, ep := range state.ThreatModel.EntryPoints {
				trusted := "untrusted"
				if ep.Trusted {
					trusted = "trusted"
				}
				threatModelBlock += fmt.Sprintf("- **%s** (%s, %s) — validation: %s\n", ep.Endpoint, ep.Type, trusted, ep.Validation)
			}
		}
	}

	// Build disposition summary block
	dispositionBlock := ""
	if len(state.FindingDispositions) > 0 {
		tp, fp, nr := 0, 0, 0
		for _, d := range state.FindingDispositions {
			switch d.Disposition {
			case "True Positive":
				tp++
			case "False Positive":
				fp++
			default:
				nr++
			}
		}
		dispositionBlock = fmt.Sprintf("\n## Finding Dispositions (Threat Model)\n- True Positives: %d\n- False Positives: %d\n- Needs Manual Review: %d\n", tp, fp, nr)

		// Audit trail: how each disposition was produced (scale transparency).
		srcCounts := map[string]int{}
		for _, d := range state.FindingDispositions {
			if d.DispositionSource != "" {
				srcCounts[d.DispositionSource]++
			}
		}
		if len(srcCounts) > 0 {
			dispositionBlock += fmt.Sprintf("- Disposition sources: %d LLM, %d cached, %d deterministic, %d NR-fallback\n",
				srcCounts[dispositionSourceLLM], srcCounts[dispositionSourceCache],
				srcCounts[dispositionSourceDeterministic], srcCounts[dispositionSourceNRFallback])
		}

		// Include False Positive rationales
		var fpLines []string
		for _, d := range state.FindingDispositions {
			if d.Disposition == "False Positive" {
				fpLines = append(fpLines, fmt.Sprintf("- **%s**: %s", d.FindingID, d.Rationale))
			}
		}
		if len(fpLines) > 0 {
			dispositionBlock += "\n### False Positive Rationales\n" + strings.Join(fpLines, "\n") + "\n"
		}
	}

	// Build PoC summary block
	pocBlock := ""
	if len(state.PoCResults) > 0 {
		pocBlock = "\n## PoC Verification Results\n"
		for _, poc := range state.PoCResults {
			pocBlock += fmt.Sprintf("\n### %s (%s)\n- **File**: %s\n- **Conclusion**: %s\n",
				poc.VulnerabilityType, poc.Severity, poc.AffectedFile, poc.Conclusion)
			if len(poc.ReasoningSteps) > 0 {
				pocBlock += "\n| Step | Description | Result |\n|---|---|---|\n"
				for _, step := range poc.ReasoningSteps {
					pocBlock += fmt.Sprintf("| %s | %s | %s |\n", step.Step, step.Description, step.Result)
				}
			}
		}
	}

	hc := state.HealthCheck
	healthBlock := fmt.Sprintf("- **Health Check**: %d/100 (%s) — the authoritative security posture score\n- **Security Gate Verdict**: %s under `%s` policy (`fail_on=%s`)\n- **Health Check Breakdown**: %d active findings, %d ignored (False Positives), %d deduplicated; penalty %d, bonus %d\n",
		hc.Score, hc.Grade,
		strings.ToUpper(hc.Verdict.Status), hc.Policy.Profile, hc.Policy.FailOn,
		hc.Breakdown.ActiveFindings, hc.Breakdown.IgnoredFindings, hc.Breakdown.DedupedFindings,
		hc.Breakdown.Penalty, hc.Breakdown.Bonus)
	if len(hc.Verdict.BlockingReasons) > 0 {
		var reasonLines []string
		for _, reason := range hc.Verdict.BlockingReasons {
			reasonLines = append(reasonLines, fmt.Sprintf("  - `%s`: %s", reason.Code, reason.Message))
		}
		healthBlock += "- **Security Gate Blocking Reasons**:\n" + strings.Join(reasonLines, "\n") + "\n"
	}

	metadataBlock := fmt.Sprintf("## AITriage + SecureCoder Engine Summary\n- **Date**: %s\n%s- **Total raw findings**: %d\n%s%s\n### Original Findings Reference Table (CRITICAL: Use these Vulnerability ID/File/Line mappings for your output):\n%s\n%s\n",
		time.Now().Format("January 2, 2006"), healthBlock, len(state.EnrichedFindings),
		threatModelBlock, dispositionBlock, lookupTable, pocBlock)

	userPrompt := fmt.Sprintf(prompts.ReportUserPromptTemplate, metadataBlock)

	messages := []llm.Message{
		{Role: "system", Content: prompts.ReportSystemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, usage, err := llmClient.Chat(ctx, messages)
	addUsage(&state.TotalUsage, usage)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	state.ReportMarkdown = response
	return nil
}

// generateSummary builds a deterministic (no LLM), actionable Markdown summary
// for the GitHub Actions Step Summary. The output is split into three blocks:
//
//  1. Human Summary — compact health card for quick glance by humans
//  2. AI Remediation Prompt — copy-paste SecureCoder implementation brief
//  3. AI Agent Data — structured JSON in a collapsed &lt;details&gt; block
//
// False Positives are excluded from actionable sections but mentioned in stats.
func generateSummary(state *AgentState) {
	var sb strings.Builder

	// ── Precompute dispositions ───────────────────────────────────────────
	tp, fpCount, nr := 0, 0, 0
	dispositionMap := make(map[int]string) // findingIndex → disposition
	for _, d := range state.FindingDispositions {
		switch d.Disposition {
		case "True Positive":
			tp++
		case "False Positive":
			fpCount++
		default:
			nr++
		}
		dispositionMap[d.FindingIndex] = d.Disposition
	}

	// ── Collect actionable findings (TP + NR) ─────────────────────────────
	var actionable []actionableFinding
	for i, ef := range state.EnrichedFindings {
		disp, ok := dispositionMap[i]
		if !ok {
			disp = "Needs Manual Review"
		}
		if disp == "False Positive" {
			continue
		}
		msg := ef.Message
		if len(msg) > 120 {
			msg = msg[:117] + "..."
		}
		msg = strings.ReplaceAll(msg, "|", "\\|")
		msg = strings.ReplaceAll(msg, "\n", " ")

		actionable = append(actionable, actionableFinding{
			vulnID:      ef.VulnID,
			source:      ef.Source,
			severity:    ef.Severity,
			file:        ef.File,
			line:        ef.Line,
			message:     msg,
			disposition: disp,
		})
	}

	// ── Block 1: Human Summary ────────────────────────────────────────────
	writeHumanSummary(&sb, state, actionable, tp, fpCount, nr)

	// ── Block 2: AI Remediation Prompt ────────────────────────────────────
	writeAIRemediationPrompt(&sb, state, actionable)

	// ── Block 3: AI Agent Data ────────────────────────────────────────────
	writeAIAgentData(&sb, state, actionable, tp, fpCount, nr)

	// ── Footer ────────────────────────────────────────────────────────────
	sb.WriteString("\n---\n")
	if fpCount > 0 {
		sb.WriteString(fmt.Sprintf("\n_%d false positive(s) suppressed. Download `report.md` artifact for the full audit trail with FP rationale._\n",
			fpCount))
	}
	if state.TotalUsage.TotalTokens > 0 {
		sb.WriteString(fmt.Sprintf("\n_LLM usage (provider reported): %s. Cost is not estimated because it depends on provider, model, caching, and billing tier._\n",
			formatLLMUsage(state.TotalUsage)))
	}

	state.SummaryMarkdown = sb.String()
	fmt.Fprintf(os.Stderr, "   Summary: %d actionable, %d suppressed FP\n", len(actionable), fpCount)
}

// ── Block 1: Human-Readable Summary ─────────────────────────────────────────

type actionableFinding struct {
	vulnID      string
	source      string
	severity    string
	file        string
	line        int
	message     string
	disposition string
}

func writeHumanSummary(sb *strings.Builder, state *AgentState, actionable []actionableFinding, tp, fp, nr int) {
	hc := state.HealthCheck

	sb.WriteString("## 🛡 Security Assessment\n\n")

	verdict := "✅ PASSED"
	if !hc.Verdict.Passed {
		verdict = "❌ FAILED"
	}
	sb.WriteString(fmt.Sprintf("**Score**: %d/100 (%s) | **Gate**: %s | **Policy**: `%s` (`fail_on=%s`)\n\n",
		hc.Score, hc.Grade, verdict, hc.Policy.Profile, hc.Policy.FailOn))

	// ── Blocking Reasons ──────────────────────────────────────────────
	if len(hc.Verdict.BlockingReasons) > 0 {
		sb.WriteString("#### Blocking Reasons\n\n")
		for _, reason := range hc.Verdict.BlockingReasons {
			sb.WriteString(fmt.Sprintf("- `%s`: %s", reason.Code, reason.Message))
			if reason.Count != 0 || reason.Threshold != 0 {
				sb.WriteString(fmt.Sprintf(" (count %d, threshold %d)", reason.Count, reason.Threshold))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// ── Severity Matrix ───────────────────────────────────────────────
	sevByDisp := map[string]map[string]int{
		"True Positive":       {"CRITICAL": 0, "HIGH": 0, "MEDIUM": 0, "LOW": 0},
		"Needs Manual Review": {"CRITICAL": 0, "HIGH": 0, "MEDIUM": 0, "LOW": 0},
	}
	for _, f := range actionable {
		sev := strings.ToUpper(f.severity)
		if _, ok := sevByDisp[f.disposition]; !ok {
			continue
		}
		if _, ok := sevByDisp[f.disposition][sev]; ok {
			sevByDisp[f.disposition][sev]++
		}
	}

	sb.WriteString("### Overview\n\n")
	sb.WriteString("| | Critical | High | Medium | Low |\n")
	sb.WriteString("|---|---|---|---|---|\n")
	sb.WriteString(fmt.Sprintf("| True Positives | %d | %d | %d | %d |\n",
		sevByDisp["True Positive"]["CRITICAL"], sevByDisp["True Positive"]["HIGH"],
		sevByDisp["True Positive"]["MEDIUM"], sevByDisp["True Positive"]["LOW"]))
	sb.WriteString(fmt.Sprintf("| Needs Review | %d | %d | %d | %d |\n\n",
		sevByDisp["Needs Manual Review"]["CRITICAL"], sevByDisp["Needs Manual Review"]["HIGH"],
		sevByDisp["Needs Manual Review"]["MEDIUM"], sevByDisp["Needs Manual Review"]["LOW"]))

	sb.WriteString(fmt.Sprintf("> **%d** findings analyzed · **%d** true positives · **%d** needs review · **%d** false positives suppressed\n\n",
		len(state.EnrichedFindings), tp, nr, fp))

	// ── Top Critical Issues (max 5, CRITICAL first then HIGH) ─────────
	if len(actionable) > 0 {
		sb.WriteString("### ⚠️ Top Critical Issues\n\n")

		type ranked struct {
			severity string
			vulnID   string
			file     string
			line     int
			message  string
		}
		var top []ranked
		// First pass: CRITICAL
		for _, f := range actionable {
			if strings.ToUpper(f.severity) == "CRITICAL" {
				top = append(top, ranked{severity: f.severity, vulnID: f.vulnID, file: f.file, line: f.line, message: f.message})
			}
		}
		// Second pass: HIGH
		for _, f := range actionable {
			if strings.ToUpper(f.severity) == "HIGH" {
				top = append(top, ranked{severity: f.severity, vulnID: f.vulnID, file: f.file, line: f.line, message: f.message})
			}
		}

		limit := 5
		if len(top) < limit {
			limit = len(top)
		}
		for i := 0; i < limit; i++ {
			file := top[i].file
			if file == "" {
				file = "N/A"
			}
			// Clean message for display: take first part before ":"
			displayMsg := top[i].message
			if idx := strings.Index(displayMsg, ":"); idx > 0 && idx < 60 {
				displayMsg = displayMsg[:idx]
			}
			displayMsg = strings.ReplaceAll(displayMsg, "\\|", "|")
			if top[i].line > 0 {
				sb.WriteString(fmt.Sprintf("%d. **[%s]** %s — `%s:%d`\n", i+1, top[i].severity, displayMsg, file, top[i].line))
			} else {
				sb.WriteString(fmt.Sprintf("%d. **[%s]** %s — `%s`\n", i+1, top[i].severity, displayMsg, file))
			}
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("### Actionable Findings\n\nNo actionable security findings.\n\n")
	}

	// ── PoC Summary (compact) ─────────────────────────────────────────
	if len(state.PoCResults) > 0 {
		sb.WriteString("### PoC Verification\n\n")
		sb.WriteString("| Vulnerability | Severity | Conclusion |\n")
		sb.WriteString("|---|---|---|\n")
		for _, poc := range state.PoCResults {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
				poc.VulnerabilityType, poc.Severity, poc.Conclusion))
		}
		sb.WriteString("\n")
	}
}

// ── Block 2: AI Agent Structured Data ───────────────────────────────────────

func writeAIAgentData(sb *strings.Builder, state *AgentState, actionable []actionableFinding, tp, fp, nr int) {
	// Build JSON structure
	type findingJSON struct {
		ID             string `json:"id"`
		Severity       string `json:"severity"`
		File           string `json:"file,omitempty"`
		Line           int    `json:"line,omitempty"`
		Title          string `json:"title"`
		Disposition    string `json:"disposition"`
		Recommendation string `json:"recommendation,omitempty"`
	}

	type summaryJSON struct {
		ScanDate   string `json:"scan_date"`
		Score      int    `json:"score"`
		Grade      string `json:"grade"`
		GateStatus string `json:"gate_status"`
		Policy     struct {
			Profile string `json:"profile"`
			FailOn  string `json:"fail_on"`
		} `json:"policy"`
		Stats struct {
			TruePositives  int `json:"true_positives"`
			NeedsReview    int `json:"needs_review"`
			FalsePositives int `json:"false_positives"`
			Total          int `json:"total"`
		} `json:"stats"`
		Findings []findingJSON `json:"findings"`
	}

	hc := state.HealthCheck
	data := summaryJSON{
		ScanDate: time.Now().Format("2006-01-02"),
		Score:    hc.Score,
		Grade:    hc.Grade,
	}
	if hc.Verdict.Passed {
		data.GateStatus = "PASSED"
	} else {
		data.GateStatus = "FAILED"
	}
	data.Policy.Profile = string(hc.Policy.Profile)
	data.Policy.FailOn = string(hc.Policy.FailOn)
	data.Stats.TruePositives = tp
	data.Stats.NeedsReview = nr
	data.Stats.FalsePositives = fp
	data.Stats.Total = len(state.EnrichedFindings)

	// Build finding recommendations from dispositions
	recommendationMap := make(map[string]string)
	for _, d := range state.FindingDispositions {
		if d.Disposition != "False Positive" && d.Rationale != "" {
			recommendationMap[d.FindingID] = d.Rationale
		}
	}

	for _, f := range actionable {
		fj := findingJSON{
			ID:          f.vulnID,
			Severity:    f.severity,
			File:        f.file,
			Line:        f.line,
			Title:       strings.ReplaceAll(f.message, "\\|", "|"),
			Disposition: f.disposition,
		}
		if rec, ok := recommendationMap[f.vulnID]; ok {
			fj.Recommendation = rec
		}
		data.Findings = append(data.Findings, fj)
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return
	}

	sb.WriteString("\n<details>\n")
	sb.WriteString("<summary>🤖 AI Agent Data (structured findings for Cursor / Claude / Antigravity)</summary>\n\n")
	sb.WriteString("```json\n")
	sb.WriteString(string(jsonBytes))
	sb.WriteString("\n```\n\n")
	sb.WriteString("</details>\n")
}

// ── Block 2: AI Remediation Prompt ──────────────────────────────────────────

func writeAIRemediationPrompt(sb *strings.Builder, state *AgentState, actionable []actionableFinding) {
	if len(actionable) == 0 {
		return
	}

	hc := state.HealthCheck

	sb.WriteString("\n### 📋 AI Remediation Prompt\n\n")
	sb.WriteString("> Copy this implementation brief into your AI IDE. It must audit and plan first, then implement verified fixes.\n\n")
	sb.WriteString("<details>\n")
	sb.WriteString("<summary>Click to expand prompt</summary>\n\n")
	sb.WriteString("```markdown\n")

	sb.WriteString("You are a SecureCoder security engineer working in this repository. Below is triage evidence from an AITriage security scan.\n")
	sb.WriteString("Your goal is a secure, verified remediation — not merely a checklist. Follow every phase in order.\n\n")

	sb.WriteString("## SCAN METADATA\n")
	sb.WriteString(fmt.Sprintf("- Score: %d/100 (%s)\n", hc.Score, hc.Grade))
	sb.WriteString(fmt.Sprintf("- Date: %s\n", time.Now().Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("- Gate: %s\n\n", strings.ToUpper(hc.Verdict.Status)))

	sb.WriteString("## VULNERABILITIES FOUND\n\n")
	for i, f := range actionable {
		file := f.file
		if file == "" {
			file = "N/A"
		}
		title := strings.ReplaceAll(f.message, "\\|", "|")
		if f.line > 0 {
			sb.WriteString(fmt.Sprintf("%d. [%s] %s | %s | %s:%d\n",
				i+1, strings.ToUpper(f.severity), f.vulnID, title, file, f.line))
		} else {
			sb.WriteString(fmt.Sprintf("%d. [%s] %s | %s | %s\n",
				i+1, strings.ToUpper(f.severity), f.vulnID, title, file))
		}
		sb.WriteString(fmt.Sprintf("   Status: %s\n", f.disposition))
	}

	sb.WriteString("\n## OPERATING CONTRACT\n\n")
	sb.WriteString("### Phase 0 — Audit before code\n")
	sb.WriteString("- Do not modify code until you have completed a scoped read-only audit of the affected files, dependencies, entry points, configuration, and side effects.\n")
	sb.WriteString("- Inspect available tools first. Use filesystem, search, git, browser, and MCP capabilities when available; never invent unavailable APIs, libraries, or tool results.\n")
	sb.WriteString("- For every non-trivial API, framework, dependency, or version change, verify the current official documentation before implementation. Record material sources and compatibility constraints in the plan.\n\n")

	sb.WriteString("### Phase 1 — Create the remediation plan\n")
	sb.WriteString("- Before editing, create a short lowercase-kebab-case `*.md` plan file in the repository root that names this remediation. Do not overwrite an existing plan.\n")
	sb.WriteString("- Group correlated findings by root cause and component; do not create duplicate fixes for the same defect. Prioritize CRITICAL, then HIGH, MEDIUM, LOW.\n")
	sb.WriteString("- For every remediation unit record affected files, the exact intended code/configuration change, security invariant, compatibility or migration risk, dependencies, verification, and acceptance criteria.\n")
	sb.WriteString("- Track each task and subtask with checkboxes. Do not begin implementation until this plan is complete.\n\n")

	sb.WriteString("### Phase 2 — Implement verified fixes\n")
	sb.WriteString("- Do not stop after the plan. Implement only findings marked `True Positive`, one remediation unit at a time, and update the plan after every completed unit.\n")
	sb.WriteString("- For `Needs Manual Review`, do not make speculative changes. Record the evidence, decision required, safe options, and the verification needed from a human owner.\n")
	sb.WriteString("- Preserve least privilege and secure-by-default behaviour: enforce authentication and authorization server-side, deny by default, validate inputs with allowlists, parameterize data access, encode untrusted output, keep secrets out of code and logs, use narrow CORS/CSP/cookie settings, and remove insecure debug/default behaviour.\n")
	sb.WriteString("- Use official advisories and documentation for dependency remediation. Preserve lockfiles and compatibility; never suppress a scanner, weaken a policy, disable a test, or hide a finding merely to obtain a green result.\n")
	sb.WriteString("- Keep changes minimal and scoped. Do not perform unrelated mass refactors. If a safe implementation depends on missing authority, a product decision, or uncertain facts, stop that unit and record the blocker.\n\n")

	sb.WriteString("### Phase 3 — Verify and report\n")
	sb.WriteString("- After every remediation unit, run the narrowest relevant test, linter, or security check. At the end, run the complete applicable verification suite.\n")
	sb.WriteString("- Confirm each fixed vulnerability is no longer reproducible and that authorization, negative-path, regression, and compatibility tests cover the intended security invariant.\n")
	sb.WriteString("- Finish only when every plan item is checked or explicitly marked blocked with an owner decision. Report changed files, finding IDs addressed, commands run, results, and any residual risk.\n")

	sb.WriteString("```\n\n")
	sb.WriteString("</details>\n")
}

func generateAIFixSpec(ctx context.Context, state *AgentState, llmClient llm.Client) error {
	// Extract repository name from project path (last path component)
	repoName := filepath.Base(state.ProjectPath)
	if repoName == "." || repoName == "/" {
		repoName = "unknown"
	}

	// Get stack and project tree from RepoContext
	stack := "Not detected"
	projectTree := ""
	if state.RepoContext != nil {
		if state.RepoContext.Stack != "" {
			stack = state.RepoContext.Stack
		}
		if state.RepoContext.ProjectTree != "" {
			projectTree = state.RepoContext.ProjectTree
			if len(projectTree) > 3000 {
				projectTree = projectTree[:3000] + "\n... (truncated)"
			}
		}
	}

	userPrompt := fmt.Sprintf(prompts.FixSpecUserPromptTemplate, repoName, stack, projectTree, state.ReportMarkdown)

	messages := []llm.Message{
		{Role: "system", Content: prompts.FixSpecSystemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, usage, err := llmClient.Chat(ctx, messages)
	addUsage(&state.TotalUsage, usage)
	if err != nil {
		return fmt.Errorf("failed to generate fix spec: %w", err)
	}

	state.AIFixSpec = response
	return nil
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// addUsage accumulates LLM token usage from a single call into the total.
func addUsage(total *llm.Usage, u llm.Usage) {
	total.PromptTokens += u.PromptTokens
	total.CompletionTokens += u.CompletionTokens
	total.TotalTokens += u.TotalTokens
}

// formatLLMUsage preserves the provider's total instead of inventing a price.
// Gemini thinking models can report tokens beyond prompt and visible completion.
func formatLLMUsage(u llm.Usage) string {
	parts := []string{
		fmt.Sprintf("%d total", u.TotalTokens),
		fmt.Sprintf("%d prompt", u.PromptTokens),
		fmt.Sprintf("%d completion", u.CompletionTokens),
	}
	if additional := u.TotalTokens - u.PromptTokens - u.CompletionTokens; additional > 0 {
		parts = append(parts, fmt.Sprintf("%d reasoning/other", additional))
	}
	return strings.Join(parts, " · ")
}

// extractFullContext extracts the full function body + imports for a finding.
// Uses tree-sitter AST when possible, falls back to ±30 lines.
func extractFullContext(projectPath, file string, line int) string {
	if file == "" || line <= 0 {
		return "Context not available."
	}
	cleanPath := strings.TrimPrefix(file, "/src/")
	fullPath := cleanPath
	if !filepath.IsAbs(cleanPath) {
		fullPath = filepath.Join(projectPath, cleanPath)
	}

	fc, err := agentcontext.ExtractFunction(fullPath, line)
	if err != nil {
		return "Context not available."
	}

	var sb strings.Builder
	if fc.Imports != "" {
		sb.WriteString("// Imports:\n")
		sb.WriteString(fc.Imports)
		sb.WriteString("\n\n")
	}
	sb.WriteString(fmt.Sprintf("// Function: %s (lines %d-%d)\n", fc.Name, fc.StartLine, fc.EndLine))
	sb.WriteString(fc.Body)
	return sb.String()
}

// extractJSON extracts a JSON block from an LLM response that may contain
// markdown code fences or other text around the JSON.
func extractJSON(text string) string {
	// Try ```json ... ``` first
	if idx := strings.Index(text, "```json"); idx >= 0 {
		rest := text[idx+7:]
		if endIdx := strings.Index(rest, "```"); endIdx >= 0 {
			return strings.TrimSpace(rest[:endIdx])
		}
	}
	// Try ``` ... ```
	if idx := strings.Index(text, "```"); idx >= 0 {
		rest := text[idx+3:]
		if endIdx := strings.Index(rest, "```"); endIdx >= 0 {
			return strings.TrimSpace(rest[:endIdx])
		}
	}
	// Try to find raw JSON (starts with { or [)
	for i, ch := range text {
		if ch == '{' || ch == '[' {
			return strings.TrimSpace(text[i:])
		}
	}
	return text
}
