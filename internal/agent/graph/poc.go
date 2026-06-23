package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/agent/prompts"
)

// ── Phase 5b: PoC Verification scaling (SecureCoder step 5) ───────────────────
//
// PoC verification is also SecureCoder and previously capped True Positives at
// 75 with a SILENT drop. We apply the same scaling layers: dedup identical TPs
// by fingerprint, process ALL of them in bounded-concurrency batches, and for
// anything left over by the budget emit a "Needs Manual Review" PoC result
// instead of dropping it.

// pocBatchSize is smaller than the classification batch because PoC reasoning is
// far heavier per finding (full step-by-step exploit trace).
const pocBatchSize = 25

// errPoCParse marks a malformed PoC response, tolerated per batch (the finding
// simply gets no PoC trace) rather than failing the whole pipeline.
var errPoCParse = errors.New("parse PoC JSON")

// verifyPoCs runs PoC verification over ALL true-positive findings (deduped),
// returning the collected results. Transport/provider errors are fatal.
func verifyPoCs(ctx context.Context, tpFindings []EnrichedFinding, llmClient llm.Client, usage *llm.Usage) ([]PoCResult, error) {
	if len(tpFindings) == 0 {
		return nil, nil
	}

	// Dedup identical TPs so the same vulnerability is traced only once.
	unique, _ := dedupFindings(tpFindings)

	// Optional budget: anything beyond it becomes a Needs-Manual-Review result.
	var overflow []EnrichedFinding
	if budget := pocBudget(); budget >= 0 && len(unique) > budget {
		overflow = unique[budget:]
		unique = unique[:budget]
	}

	var batches [][]EnrichedFinding
	for off := 0; off < len(unique); off += pocBatchSize {
		batches = append(batches, unique[off:min(off+pocBatchSize, len(unique))])
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		mu       sync.Mutex
		usageMu  sync.Mutex
		errMu    sync.Mutex
		firstErr error
		results  []PoCResult
		wg       sync.WaitGroup
	)
	sem := make(chan struct{}, triageConcurrency())
	setErr := func(e error) {
		errMu.Lock()
		if firstErr == nil {
			firstErr = e
			cancel()
		}
		errMu.Unlock()
	}

	for _, batch := range batches {
		wg.Add(1)
		sem <- struct{}{}
		go func(batch []EnrichedFinding) {
			defer wg.Done()
			defer func() { <-sem }()

			res, u, err := pocBatchLLM(ctx, batch, llmClient)
			usageMu.Lock()
			addUsage(usage, u)
			usageMu.Unlock()
			if err != nil {
				// Malformed PoC responses are tolerated; only transport errors abort.
				if !errors.Is(err, errPoCParse) {
					setErr(err)
				}
				return
			}
			mu.Lock()
			results = append(results, res...)
			mu.Unlock()
		}(batch)
	}
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	// Budget overflow: never silently dropped — surfaced as Needs Manual Review.
	for _, f := range overflow {
		results = append(results, PoCResult{
			VulnerabilityType: pocTypeOf(f),
			Severity:          f.Severity,
			AffectedFile:      f.File,
			Conclusion:        "Needs Manual Review",
		})
	}

	return results, nil
}

// pocBatchLLM verifies a single batch of findings. Transport errors are wrapped
// plainly; malformed JSON is wrapped with errPoCParse.
func pocBatchLLM(ctx context.Context, batch []EnrichedFinding, llmClient llm.Client) ([]PoCResult, llm.Usage, error) {
	findingsJSON, _ := json.MarshalIndent(batch, "", "  ")
	userPrompt := fmt.Sprintf(prompts.PoCUserPromptTemplate, len(batch), string(findingsJSON))

	messages := []llm.Message{
		{Role: "system", Content: prompts.PoCSystemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, u, err := llmClient.Chat(ctx, messages)
	if err != nil {
		return nil, u, fmt.Errorf("PoC verification LLM call failed: %w", err)
	}

	jsonText := extractJSON(response)
	var results []PoCResult
	if err := json.Unmarshal([]byte(jsonText), &results); err != nil {
		var single PoCResult
		if err2 := json.Unmarshal([]byte(jsonText), &single); err2 == nil {
			return []PoCResult{single}, u, nil
		}
		return nil, u, fmt.Errorf("%w: %v", errPoCParse, err)
	}
	return results, u, nil
}

func pocTypeOf(f EnrichedFinding) string {
	if f.ID != "" {
		return f.ID
	}
	return strings.TrimSpace(f.Message)
}

// pocBudget caps how many unique TP findings get a (heavy) PoC trace.
// Negative (default) means unlimited.
func pocBudget() int {
	if v := strings.TrimSpace(os.Getenv("AITRIAGE_POC_BUDGET")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return -1
}
