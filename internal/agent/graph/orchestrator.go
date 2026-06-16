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

	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/agent/prompts"
)

// Run Orchestrates the map-reduce pipeline.
func Run(ctx context.Context, state *AgentState, llmClient llm.Client) error {
	fmt.Fprintf(os.Stderr, "🤖 Context Enrichment...\n")
	enrichFindings(state)

	fmt.Fprintf(os.Stderr, "🤖 Map-Reduce Triaging (%d batches)...\n", len(state.Batches))
	if err := runWorkers(ctx, state, llmClient); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "🤖 Generating Security Report...\n")
	if err := generateReport(ctx, state, llmClient); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "🤖 Generating AI Fix Specification...\n")
	if err := generateAIFixSpec(ctx, state, llmClient); err != nil {
		return err
	}

	return nil
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
			Snippet:  readSnippet(state.ProjectPath, f.File, f.Line),
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
			Snippet:  readSnippet(state.ProjectPath, f.File, f.Line),
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
			Type:     "deploy",
			Severity: f.Severity,
			File:     f.File,
			Line:     f.Line,
			Message:  fmt.Sprintf("%s. Advice: %s", f.Issue, f.Advice),
			Snippet:  readSnippet(state.ProjectPath, f.File, f.Line),
		})
	}
	for _, f := range state.NetworkFindings {
		enriched = append(enriched, EnrichedFinding{
			Type:     "network",
			Severity: f.Severity,
			Message:  fmt.Sprintf("Port %d (%s): %s", f.Port, f.Service, f.Message),
		})
	}

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

func runWorkers(ctx context.Context, state *AgentState, llmClient llm.Client) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(state.Batches))

	for i, batch := range state.Batches {
		wg.Add(1)
		go func(idx int, b []EnrichedFinding) {
			defer wg.Done()

			batchJSON, _ := json.MarshalIndent(b, "", "  ")
			userPrompt := fmt.Sprintf(prompts.TriageUserPromptTemplate, string(batchJSON))

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

func generateReport(ctx context.Context, state *AgentState, llmClient llm.Client) error {
	combinedTriage := strings.Join(state.TriagedResults, "\n\n")

	if len(combinedTriage) > 30000 {
		combinedTriage = combinedTriage[:30000] + "\n...[TRUNCATED — too many findings]"
	}

	metadataBlock := fmt.Sprintf("## AITriage Core Engine Summary\n- **Date**: %s\n- **Security Score**: %d/100 (%s)\n- **Total raw findings**: %d\n\n",
		time.Now().Format("January 2, 2006"), state.SecurityScore, state.SecurityGrade, len(state.EnrichedFindings))

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

func readSnippet(projectPath, file string, line int) string {
	if file == "" || line <= 0 {
		return "Snippet not available."
	}
	cleanPath := strings.TrimPrefix(file, "/src/")
	fullPath := cleanPath
	if !filepath.IsAbs(cleanPath) {
		fullPath = filepath.Join(projectPath, cleanPath)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "Snippet not available."
	}

	lines := strings.Split(string(content), "\n")
	idx := line - 1
	start := idx - 3
	if start < 0 {
		start = 0
	}
	end := idx + 8
	if end > len(lines) {
		end = len(lines)
	}

	return strings.Join(lines[start:end], "\n")
}
