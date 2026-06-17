package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	agentcontext "github.com/cybertortuga/aitriage/internal/agent/context"
	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/agent/prompts"
	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
)

// Run Orchestrates the full SecureCoder-enhanced pipeline:
//
//	enrichFindings → buildThreatModel → runWorkers (Map-Reduce) →
//	runPoCVerification → generateReport → generateAIFixSpec
func Run(ctx context.Context, state *AgentState, llmClient llm.Client) error {
	// Step 0: Gather repository context (reads files from disk, no LLM)
	fmt.Fprintf(os.Stderr, "📂 Gathering Repository Context...\n")
	gatherRepoContext(state)

	fmt.Fprintf(os.Stderr, "🤖 Context Enrichment...\n")
	enrichFindings(state)

	// SecureCoder Step 1: Threat Model
	fmt.Fprintf(os.Stderr, "🏗️ Building Threat Model (SecureCoder)...\n")
	if err := buildThreatModel(ctx, state, llmClient); err != nil {
		// Non-fatal: continue without threat model
		fmt.Fprintf(os.Stderr, "⚠️ Threat model step failed (continuing): %v\n", err)
	}

	// Health Check: recompute the authoritative IB posture score across ALL
	// sources now that AI dispositions (TP/FP) are available. False Positives
	// no longer penalise the repository.
	fmt.Fprintf(os.Stderr, "🩺 Computing Security Health Check (all sources, FP-aware)...\n")
	computeHealthCheck(state)
	fmt.Fprintf(os.Stderr, "   ✅ Health Check: %d/100 (%s) — %d active, %d ignored (FP), %d deduped\n",
		state.HealthCheck.Score, state.HealthCheck.Grade,
		state.HealthCheck.Breakdown.ActiveFindings,
		state.HealthCheck.Breakdown.IgnoredFindings,
		state.HealthCheck.Breakdown.DedupedFindings)

	fmt.Fprintf(os.Stderr, "🤖 Map-Reduce Triaging (%d batches)...\n", len(state.Batches))
	if err := runWorkers(ctx, state, llmClient); err != nil {
		return err
	}

	// SecureCoder Step 2: PoC Verification
	fmt.Fprintf(os.Stderr, "🧪 PoC Verification (SecureCoder)...\n")
	if err := runPoCVerification(ctx, state, llmClient); err != nil {
		// Non-fatal: continue without PoC
		fmt.Fprintf(os.Stderr, "⚠️ PoC verification step failed (continuing): %v\n", err)
	}

	fmt.Fprintf(os.Stderr, "🤖 Generating Security Report (CS-XXX-NNN format)...\n")
	if err := generateReport(ctx, state, llmClient); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "🤖 Generating AI Fix Specification...\n")
	if err := generateAIFixSpec(ctx, state, llmClient); err != nil {
		return err
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
			Severity: f.Severity,
			Message:  fmt.Sprintf("%s: %s", f.Name, f.Message),
		})
	}
	for _, f := range state.DeployFindings {
		enriched = append(enriched, EnrichedFinding{
			ID:       f.Issue,
			Type:     "deploy",
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

// computeHealthCheck recomputes the authoritative IB Health Check across ALL
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

func buildThreatModel(ctx context.Context, state *AgentState, llmClient llm.Client) error {
	if len(state.EnrichedFindings) == 0 {
		fmt.Fprintf(os.Stderr, "   ℹ️ No findings — skipping threat model\n")
		return nil
	}

	// Serialize findings for the prompt (cap at 20 to stay within token limits)
	findingsToSend := state.EnrichedFindings
	if len(findingsToSend) > 20 {
		findingsToSend = findingsToSend[:20]
	}
	findingsJSON, _ := json.MarshalIndent(findingsToSend, "", "  ")

	// Build repo context summary for the prompt.
	repoContextText := ""
	if state.RepoContext != nil {
		repoContextText = state.RepoContext.FormatForLLM(5000) // ~5K tokens for threat model
	}

	userPrompt := fmt.Sprintf(prompts.ThreatModelUserPromptTemplate,
		repoContextText,
		state.ProjectPath,
		len(state.EnrichedFindings),
		string(findingsJSON),
	)

	messages := []llm.Message{
		{Role: "system", Content: prompts.ThreatModelSystemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, _, err := llmClient.Chat(ctx, messages)
	if err != nil {
		return fmt.Errorf("threat model LLM call failed: %w", err)
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
		// Fallback: mark all as True Positive if we can't parse
		fmt.Fprintf(os.Stderr, "   ⚠️ Could not parse threat model JSON: %v (defaulting all to True Positive)\n", err)
		for i, f := range state.EnrichedFindings {
			state.FindingDispositions = append(state.FindingDispositions, FindingDisposition{
				FindingIndex: i,
				FindingID:    f.VulnID,
				Disposition:  "True Positive",
				Rationale:    "Could not build threat model; defaulting to True Positive.",
			})
		}
		return nil
	}

	state.ThreatModel = &ThreatModel{
		ComponentOverview:  rawResult.ComponentOverview,
		EntryPoints:        rawResult.EntryPoints,
		TrustBoundaries:    rawResult.TrustBoundaries,
		SensitiveDataPaths: rawResult.SensitiveDataPaths,
		PrivilegedActions:  rawResult.PrivilegedActions,
		PriorityAreas:      rawResult.PriorityAreas,
	}

	for _, d := range rawResult.FindingDispositions {
		findingID := ""
		if d.FindingIndex < len(state.EnrichedFindings) {
			findingID = state.EnrichedFindings[d.FindingIndex].VulnID
		}
		state.FindingDispositions = append(state.FindingDispositions, FindingDisposition{
			FindingIndex: d.FindingIndex,
			FindingID:    findingID,
			Disposition:  d.Disposition,
			Rationale:    d.Rationale,
		})
	}

	tp := 0
	fp := 0
	nr := 0
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
	fmt.Fprintf(os.Stderr, "   ✅ Threat model: %d True Positives, %d False Positives, %d Needs Review\n", tp, fp, nr)

	return nil
}

func runWorkers(ctx context.Context, state *AgentState, llmClient llm.Client) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(state.Batches))

	// Build threat model context summary (if available)
	threatContext := ""
	if state.ThreatModel != nil {
		tmSummary, _ := json.MarshalIndent(struct {
			ComponentOverview string   `json:"component_overview"`
			PriorityAreas     []string `json:"priority_areas"`
		}{
			ComponentOverview: state.ThreatModel.ComponentOverview,
			PriorityAreas:     state.ThreatModel.PriorityAreas,
		}, "", "  ")
		threatContext = string(tmSummary)
	}

	// Build disposition lookup for enriching batch context
	dispositionMap := make(map[string]string) // VulnID → Disposition
	for _, d := range state.FindingDispositions {
		if d.FindingID != "" {
			dispositionMap[d.FindingID] = d.Disposition
		}
	}

	for i, batch := range state.Batches {
		wg.Add(1)
		go func(idx int, b []EnrichedFinding) {
			defer wg.Done()

			batchJSON, _ := json.MarshalIndent(b, "", "  ")

			var userPrompt string
			if threatContext != "" {
				userPrompt = fmt.Sprintf(prompts.TriageUserPromptWithThreatModelTemplate,
					threatContext, string(batchJSON))
			} else {
				userPrompt = fmt.Sprintf(prompts.TriageUserPromptTemplate, string(batchJSON))
			}

			messages := []llm.Message{
				{Role: "system", Content: prompts.TriageSystemPrompt},
				{Role: "user", Content: userPrompt},
			}

			response, _, err := llmClient.Chat(ctx, messages)
			if err != nil {
				errChan <- fmt.Errorf("batch %d failed: %w", idx, err)
				return
			}

			mu.Lock()
			state.TriagedResults = append(state.TriagedResults, fmt.Sprintf("--- BATCH %d ---\n%s", idx, response))
			mu.Unlock()
		}(i, batch)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for e := range errChan {
		errs = append(errs, e)
	}
	if len(errs) > 0 {
		return fmt.Errorf("encountered %d errors during triaging, first: %v", len(errs), errs[0])
	}

	return nil
}

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

	// If no dispositions yet (threat model failed), use all HIGH/CRITICAL
	if len(tpFindings) == 0 && len(state.FindingDispositions) == 0 {
		for _, f := range state.EnrichedFindings {
			sev := strings.ToUpper(f.Severity)
			if sev == "CRITICAL" || sev == "HIGH" {
				tpFindings = append(tpFindings, f)
			}
		}
	}

	if len(tpFindings) == 0 {
		fmt.Fprintf(os.Stderr, "   ℹ️ No True Positives — skipping PoC verification\n")
		return nil
	}

	// Cap at 15 findings for PoC to stay within token limits
	if len(tpFindings) > 15 {
		tpFindings = tpFindings[:15]
	}

	findingsJSON, _ := json.MarshalIndent(tpFindings, "", "  ")
	userPrompt := fmt.Sprintf(prompts.PoCUserPromptTemplate, len(tpFindings), string(findingsJSON))

	messages := []llm.Message{
		{Role: "system", Content: prompts.PoCSystemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, _, err := llmClient.Chat(ctx, messages)
	if err != nil {
		return fmt.Errorf("PoC verification LLM call failed: %w", err)
	}

	// Parse JSON response
	jsonText := extractJSON(response)
	var pocResults []PoCResult
	if err := json.Unmarshal([]byte(jsonText), &pocResults); err != nil {
		// If we can't parse as array, try single object
		var single PoCResult
		if err2 := json.Unmarshal([]byte(jsonText), &single); err2 == nil {
			pocResults = []PoCResult{single}
		} else {
			fmt.Fprintf(os.Stderr, "   ⚠️ Could not parse PoC JSON: %v\n", err)
			return nil // Non-fatal
		}
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
	fmt.Fprintf(os.Stderr, "   ✅ PoC: %d exploitable, %d blocked/unknown\n", verified, incomplete)

	return nil
}

func generateReport(ctx context.Context, state *AgentState, llmClient llm.Client) error {
	combinedTriage := strings.Join(state.TriagedResults, "\n\n")

	if len(combinedTriage) > 30000 {
		combinedTriage = combinedTriage[:30000] + "\n...[TRUNCATED — too many findings]"
	}

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
					pocBlock += fmt.Sprintf("| %d | %s | %s |\n", step.Step, step.Description, step.Result)
				}
			}
		}
	}

	hc := state.HealthCheck
	healthBlock := fmt.Sprintf("- **Health Check**: %d/100 (%s) — the authoritative IB posture score\n- **IB Gate Verdict**: %s under `%s` policy (`fail_on=%s`)\n- **Health Check Breakdown**: %d active findings, %d ignored (False Positives), %d deduplicated; penalty %d, bonus %d\n",
		hc.Score, hc.Grade,
		strings.ToUpper(hc.Verdict.Status), hc.Policy.Profile, hc.Policy.FailOn,
		hc.Breakdown.ActiveFindings, hc.Breakdown.IgnoredFindings, hc.Breakdown.DedupedFindings,
		hc.Breakdown.Penalty, hc.Breakdown.Bonus)
	if len(hc.Verdict.BlockingReasons) > 0 {
		var reasonLines []string
		for _, reason := range hc.Verdict.BlockingReasons {
			reasonLines = append(reasonLines, fmt.Sprintf("  - `%s`: %s", reason.Code, reason.Message))
		}
		healthBlock += "- **IB Gate Blocking Reasons**:\n" + strings.Join(reasonLines, "\n") + "\n"
	}

	metadataBlock := fmt.Sprintf("## AITriage + SecureCoder Engine Summary\n- **Date**: %s\n%s- **Total raw findings**: %d\n%s%s\n### Original Findings Reference Table (CRITICAL: Use these Vulnerability ID/File/Line mappings for your output):\n%s\n%s\n",
		time.Now().Format("January 2, 2006"), healthBlock, len(state.EnrichedFindings),
		threatModelBlock, dispositionBlock, lookupTable, pocBlock)

	userPrompt := fmt.Sprintf(prompts.ReportUserPromptTemplate, metadataBlock+combinedTriage)

	messages := []llm.Message{
		{Role: "system", Content: prompts.ReportSystemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, _, err := llmClient.Chat(ctx, messages)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	state.ReportMarkdown = response
	return nil
}

func generateAIFixSpec(ctx context.Context, state *AgentState, llmClient llm.Client) error {
	userPrompt := fmt.Sprintf(prompts.FixSpecUserPromptTemplate, state.ReportMarkdown)

	messages := []llm.Message{
		{Role: "system", Content: prompts.FixSpecSystemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, _, err := llmClient.Chat(ctx, messages)
	if err != nil {
		return fmt.Errorf("failed to generate fix spec: %w", err)
	}

	state.AIFixSpec = response
	return nil
}

// ── Helpers ──────────────────────────────────────────────────────────────────

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
