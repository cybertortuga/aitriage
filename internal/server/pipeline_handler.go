package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/cybertortuga/aitriage/internal/agent/graph"
	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/agent/prompts"
	"github.com/cybertortuga/aitriage/internal/models"
)

// handlePipeline runs the full SecureCoder pipeline (ThreatModel → PoC → Report → FixSpec)
// for all findings belonging to a product. It streams progress via SSE.
//
// GET /api/pipeline?product_id=12
func (s *Server) handlePipeline(w http.ResponseWriter, r *http.Request) {
	if s.llmClient == nil {
		jsonError(w, "AI Pipeline is offline. Please provide a GEMINI_API_KEY.", http.StatusServiceUnavailable)
		return
	}

	productIDStr := r.URL.Query().Get("product_id")
	if productIDStr == "" {
		jsonError(w, "product_id is required", http.StatusBadRequest)
		return
	}
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		jsonError(w, "invalid product_id", http.StatusBadRequest)
		return
	}

	// Set up SSE headers
	flusher, ok := w.(http.Flusher)
	if !ok {
		jsonError(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	sendSSE := func(data any) {
		jsonBytes, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", jsonBytes)
		flusher.Flush()
	}

	ctx := r.Context()

	// ── Step 0: Load findings from DB ────────────────────────────────
	sendSSE(map[string]any{"step": 0, "total": 4, "label": "Loading findings from database...", "progress": 5})

	dbFindings, err := s.findingRepo.ListByProductID(ctx, productID)
	if err != nil {
		sendSSE(map[string]any{"error": fmt.Sprintf("Failed to load findings: %v", err)})
		return
	}
	if len(dbFindings) == 0 {
		sendSSE(map[string]any{"error": "No findings found for this product. Run a scan first."})
		return
	}

	slog.Info("Pipeline started", "product_id", productID, "findings_count", len(dbFindings))

	// Convert DB findings → EnrichedFindings (same format as CI/CD)
	enriched := dbFindingsToEnriched(dbFindings)
	graph.AssignVulnIDsPublic(enriched)

	// Build agent state (reuse CI/CD pipeline structures)
	state := &graph.AgentState{
		EnrichedFindings: enriched,
	}

	// ── Step 1: Threat Model ─────────────────────────────────────────
	sendSSE(map[string]any{"step": 1, "total": 4, "label": "Building Threat Model (classifying TP/FP/NR)...", "progress": 15})

	if err := runWebThreatModel(ctx, state, s.llmClient); err != nil {
		slog.Error("Pipeline: threat model failed", "error", err)
		sendSSE(map[string]any{"step": 1, "warning": fmt.Sprintf("Threat model failed (continuing): %v", err)})
	}

	// Count dispositions
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
	sendSSE(map[string]any{
		"step": 1, "total": 4, "label": fmt.Sprintf("Threat Model: %d TP, %d FP, %d NR", tp, fp, nr),
		"progress": 30, "stats": map[string]int{"tp": tp, "fp": fp, "nr": nr},
	})

	// ── Step 2: PoC Verification ─────────────────────────────────────
	sendSSE(map[string]any{"step": 2, "total": 4, "label": "PoC Verification (proving exploitability)...", "progress": 35})

	if err := runWebPoC(ctx, state, s.llmClient); err != nil {
		slog.Error("Pipeline: PoC verification failed", "error", err)
		sendSSE(map[string]any{"step": 2, "warning": fmt.Sprintf("PoC failed (continuing): %v", err)})
	}
	sendSSE(map[string]any{
		"step": 2, "total": 4, "label": fmt.Sprintf("PoC: %d results", len(state.PoCResults)),
		"progress": 55,
	})

	// ── Step 3: Report ───────────────────────────────────────────────
	sendSSE(map[string]any{"step": 3, "total": 4, "label": "Generating Security Report (CS-XXX-NNN format)...", "progress": 60})

	if err := runWebReport(ctx, state, s.llmClient); err != nil {
		slog.Error("Pipeline: report generation failed", "error", err)
		sendSSE(map[string]any{"step": 3, "warning": fmt.Sprintf("Report failed: %v", err)})
	}
	sendSSE(map[string]any{
		"step": 3, "total": 4, "label": "Report generated",
		"progress": 80,
	})

	// ── Step 4: Fix Spec ─────────────────────────────────────────────
	sendSSE(map[string]any{"step": 4, "total": 4, "label": "Generating AI Fix Specification...", "progress": 85})

	if err := runWebFixSpec(ctx, state, s.llmClient); err != nil {
		slog.Error("Pipeline: fix spec generation failed", "error", err)
		sendSSE(map[string]any{"step": 4, "warning": fmt.Sprintf("FixSpec failed: %v", err)})
	}
	sendSSE(map[string]any{"step": 4, "total": 4, "label": "Fix Spec generated", "progress": 95})

	// ── Step 5: Update DB ────────────────────────────────────────────
	updateDBFromPipeline(ctx, s, dbFindings, state)

	slog.Info("Pipeline completed", "product_id", productID,
		"tp", tp, "fp", fp, "nr", nr,
		"poc_count", len(state.PoCResults),
		"report_len", len(state.ReportMarkdown),
		"fixspec_len", len(state.AIFixSpec))

	// ── Done ─────────────────────────────────────────────────────────
	sendSSE(map[string]any{
		"done":     true,
		"progress": 100,
		"stats":    map[string]int{"tp": tp, "fp": fp, "nr": nr, "poc": len(state.PoCResults), "total": len(enriched)},
		"report":   state.ReportMarkdown,
		"fix_spec": state.AIFixSpec,
		"usage": map[string]int{
			"prompt_tokens":     state.TotalUsage.PromptTokens,
			"completion_tokens": state.TotalUsage.CompletionTokens,
			"total_tokens":      state.TotalUsage.TotalTokens,
		},
	})
}

// ── DB → EnrichedFinding conversion ─────────────────────────────────────────

func dbFindingsToEnriched(findings []models.Finding) []graph.EnrichedFinding {
	var enriched []graph.EnrichedFinding
	for _, f := range findings {
		ef := graph.EnrichedFinding{
			ID:       f.RuleID,
			Type:     categorizeStack(f.Stack),
			Source:   f.Stack,
			Severity: f.Severity,
			Message:  f.Title,
		}
		if f.FilePath != nil {
			ef.File = *f.FilePath
		}
		if f.LineNumber != nil {
			ef.Line = *f.LineNumber
		}
		if f.Description != nil {
			ef.ExtraData = *f.Description
		}
		if f.CodeSnippet != nil {
			ef.Snippet = *f.CodeSnippet
		}
		enriched = append(enriched, ef)
	}
	return enriched
}

func categorizeStack(stack string) string {
	switch strings.ToLower(stack) {
	case "semgrep", "bandit", "trivy", "gitleaks":
		return "external"
	case "nfr":
		return "nfr"
	case "deploy":
		return "deploy"
	case "network":
		return "network"
	case "git-history":
		return "external"
	default:
		return "core"
	}
}

// ── Pipeline steps (wrappers around graph functions using the same prompts) ──

func runWebThreatModel(ctx context.Context, state *graph.AgentState, llmClient llm.Client) error {
	if len(state.EnrichedFindings) == 0 {
		return nil
	}

	// Reuse the same robust classifier as the CLI/CI pipeline: it batches over
	// ALL findings (no silent drop), retries omitted findings, and defaults any
	// still-unclassified finding to Needs Manual Review (never False Positive).
	// Transport/provider failures are returned so the caller can surface them.
	tm, dispositions, err := graph.ClassifyFindings(ctx, "", "web-scan", state.EnrichedFindings, llmClient, &state.TotalUsage)
	if err != nil {
		return err
	}

	state.ThreatModel = tm
	state.FindingDispositions = dispositions
	return nil
}

func runWebPoC(ctx context.Context, state *graph.AgentState, llmClient llm.Client) error {
	// Collect True Positive findings
	var tpFindings []graph.EnrichedFinding
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

	if len(tpFindings) == 0 && len(state.FindingDispositions) == 0 {
		for _, f := range state.EnrichedFindings {
			sev := strings.ToUpper(f.Severity)
			if sev == "CRITICAL" || sev == "HIGH" {
				tpFindings = append(tpFindings, f)
			}
		}
	}

	if len(tpFindings) == 0 {
		return nil
	}
	if len(tpFindings) > 75 {
		tpFindings = tpFindings[:75]
	}

	findingsJSON, _ := json.MarshalIndent(tpFindings, "", "  ")
	userPrompt := fmt.Sprintf(prompts.PoCUserPromptTemplate, len(tpFindings), string(findingsJSON))

	messages := []llm.Message{
		{Role: "system", Content: prompts.PoCSystemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, usage, err := llmClient.Chat(ctx, messages)
	addPipelineUsage(&state.TotalUsage, usage)
	if err != nil {
		return fmt.Errorf("PoC verification LLM call failed: %w", err)
	}

	jsonText := extractPipelineJSON(response)
	var pocResults []graph.PoCResult
	if err := json.Unmarshal([]byte(jsonText), &pocResults); err != nil {
		var single graph.PoCResult
		if err2 := json.Unmarshal([]byte(jsonText), &single); err2 == nil {
			pocResults = []graph.PoCResult{single}
		} else {
			return nil // Non-fatal
		}
	}

	state.PoCResults = pocResults
	return nil
}

func runWebReport(ctx context.Context, state *graph.AgentState, llmClient llm.Client) error {
	// Build lookup table
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

	// Build threat model block
	threatModelBlock := ""
	if state.ThreatModel != nil {
		threatModelBlock = fmt.Sprintf("\n## Threat Model Summary\n- **Component**: %s\n- **Priority Areas**: %s\n",
			state.ThreatModel.ComponentOverview,
			strings.Join(state.ThreatModel.PriorityAreas, ", "))
	}

	// Build disposition block
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

	// Build PoC block
	pocBlock := ""
	if len(state.PoCResults) > 0 {
		pocBlock = "\n## PoC Verification Results\n"
		for _, poc := range state.PoCResults {
			pocBlock += fmt.Sprintf("\n### %s (%s)\n- **File**: %s\n- **Conclusion**: %s\n",
				poc.VulnerabilityType, poc.Severity, poc.AffectedFile, poc.Conclusion)
		}
	}

	metadataBlock := fmt.Sprintf("## AITriage + SecureCoder Engine Summary\n- **Total raw findings**: %d\n%s%s\n### Findings Reference Table:\n%s\n%s\n",
		len(state.EnrichedFindings), threatModelBlock, dispositionBlock, lookupTable, pocBlock)

	userPrompt := fmt.Sprintf(prompts.ReportUserPromptTemplate, metadataBlock)

	messages := []llm.Message{
		{Role: "system", Content: prompts.ReportSystemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, usage, err := llmClient.Chat(ctx, messages)
	addPipelineUsage(&state.TotalUsage, usage)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	state.ReportMarkdown = response
	return nil
}

func runWebFixSpec(ctx context.Context, state *graph.AgentState, llmClient llm.Client) error {
	if state.ReportMarkdown == "" {
		return nil
	}

	userPrompt := fmt.Sprintf(prompts.FixSpecUserPromptTemplate, "web-product", "unknown", "", state.ReportMarkdown)

	messages := []llm.Message{
		{Role: "system", Content: prompts.FixSpecSystemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, usage, err := llmClient.Chat(ctx, messages)
	addPipelineUsage(&state.TotalUsage, usage)
	if err != nil {
		return fmt.Errorf("failed to generate fix spec: %w", err)
	}

	state.AIFixSpec = response
	return nil
}

// ── DB update ───────────────────────────────────────────────────────────────

func updateDBFromPipeline(ctx context.Context, s *Server, dbFindings []models.Finding, state *graph.AgentState) {
	// Map dispositions by finding index
	dispositionMap := make(map[int]graph.FindingDisposition)
	for _, d := range state.FindingDispositions {
		dispositionMap[d.FindingIndex] = d
	}

	for i, dbF := range dbFindings {
		d, ok := dispositionMap[i]
		if !ok {
			continue
		}

		var status string
		switch d.Disposition {
		case "True Positive":
			status = "true_positive"
		case "False Positive":
			status = "false_positive"
		default:
			status = "needs_review"
		}

		_ = s.findingRepo.UpdateAITriage(ctx, dbF.ID, status, d.Rationale)

		if d.Disposition == "False Positive" {
			_ = s.findingRepo.UpdateStatus(ctx, dbF.ID, "false_positive")
			_, _ = s.db.ExecContext(ctx, "UPDATE findings SET is_false_positive = 1, fp_reason = ? WHERE id = ?", d.Rationale, dbF.ID)
		} else if d.Disposition == "True Positive" {
			_ = s.findingRepo.UpdateStatus(ctx, dbF.ID, "triage")
		}
	}
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func addPipelineUsage(total *llm.Usage, u llm.Usage) {
	total.PromptTokens += u.PromptTokens
	total.CompletionTokens += u.CompletionTokens
	total.TotalTokens += u.TotalTokens
}

func extractPipelineJSON(text string) string {
	if idx := strings.Index(text, "```json"); idx >= 0 {
		rest := text[idx+7:]
		if endIdx := strings.Index(rest, "```"); endIdx >= 0 {
			return strings.TrimSpace(rest[:endIdx])
		}
	}
	if idx := strings.Index(text, "```"); idx >= 0 {
		rest := text[idx+3:]
		if endIdx := strings.Index(rest, "```"); endIdx >= 0 {
			return strings.TrimSpace(rest[:endIdx])
		}
	}
	for i, ch := range text {
		if ch == '{' || ch == '[' {
			return strings.TrimSpace(text[i:])
		}
	}
	return text
}
