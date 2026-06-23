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

// ClassifyFindings is the enterprise-scale entry point that builds a SecureCoder
// threat model and classifies EVERY finding as True Positive / False Positive /
// Needs Manual Review. It composes five layers on top of the batching safety net:
//
//	Layer 1  dedup identical findings by fingerprint (no silent drop)
//	Layer 2  reuse cached verdicts by fingerprint (cheap re-runs in CI)
//	Layer 3  severity/category gating: deterministic disposition for the rest
//	Layer 4  SecureCoder threat-model-once + structured per-finding classification
//	Layer 5  bounded concurrency + budget; anything left defaults to NR
//
// Invariants: every finding gets exactly one disposition; unclassified work
// defaults to Needs Manual Review (never False Positive); transport/provider
// failures fail the pipeline; the threat model is built ONCE and reused.
//
// The returned slice always has exactly len(findings) entries, ordered by index.
func ClassifyFindings(ctx context.Context, repoContextText, projectPath string, findings []EnrichedFinding, llmClient llm.Client, usage *llm.Usage) (*ThreatModel, []FindingDisposition, error) {
	if len(findings) == 0 {
		return nil, nil, nil
	}

	// Layer 1: collapse byte-for-byte identical findings.
	unique, groups := dedupFindings(findings)

	cache := newVerdictCache(strings.TrimSpace(os.Getenv("AITRIAGE_MODEL")))
	gating := defaultGatingConfig()

	tm, uniqueDisps, err := classifyUnique(ctx, repoContextText, projectPath, unique, llmClient, usage, cache, gating)
	if err != nil {
		return nil, nil, err
	}

	// Layer 1 (reverse): project the unique verdicts back onto every original.
	dispositions := projectDispositions(uniqueDisps, groups, findings)

	logTriageMetrics(findings, unique, uniqueDisps)
	return tm, dispositions, nil
}

// classifyUnique classifies the deduplicated findings, returning one disposition
// per unique finding (indexed by unique position).
func classifyUnique(ctx context.Context, repoContextText, projectPath string, unique []EnrichedFinding, llmClient llm.Client, usage *llm.Usage, cache *verdictCache, gating gatingConfig) (*ThreatModel, []FindingDisposition, error) {
	n := len(unique)
	result := make([]*FindingDisposition, n)

	// Layer 4a: build the SecureCoder threat model ONCE from a representative
	// sample. This call is authoritative; a transport OR parse failure here is
	// fatal (we must not mask a broken provider before any classification).
	sample := unique
	if len(sample) > threatModelBatchSize {
		sample = sample[:threatModelBatchSize]
	}
	tm, _, err := threatModelLLMCall(ctx, repoContextText, projectPath, sample, llmClient, usage)
	if err != nil {
		return nil, nil, err
	}
	tmSummary := threatModelSummary(tm)

	// Layers 2 & 3: resolve as many findings as possible without an LLM call.
	var toLLM []int
	for i, f := range unique {
		fp := Fingerprint(f)
		if cached, ok := cache.Get(fp); ok {
			cached.FindingIndex = i
			cached.DispositionSource = dispositionSourceCache
			cached.Fingerprint = fp
			result[i] = &cached
			continue
		}
		if !gating.shouldTriageWithLLM(f) {
			d := deterministicDisposition(f)
			d.FindingIndex = i
			d.Fingerprint = fp
			result[i] = &d
			continue
		}
		toLLM = append(toLLM, i)
	}

	// Layer 5: enforce an optional hard budget on LLM-bound findings. Anything
	// over budget defaults to Needs Manual Review (safe, never auto-suppressed).
	if budget := llmBudget(); budget >= 0 && len(toLLM) > budget {
		for _, i := range toLLM[budget:] {
			fp := Fingerprint(unique[i])
			result[i] = &FindingDisposition{
				FindingIndex:      i,
				Disposition:       "Needs Manual Review",
				Rationale:         budgetRationale,
				Confidence:        "low",
				DispositionSource: dispositionSourceNRFallback,
				Fingerprint:       fp,
			}
		}
		toLLM = toLLM[:budget]
	}

	// Layer 4b + 5: classify the remaining findings against the threat model,
	// in bounded-concurrency batches with per-batch retry of omitted findings.
	classified, err := classifyWithLLM(ctx, tmSummary, unique, toLLM, llmClient, usage)
	if err != nil {
		return nil, nil, err
	}
	for gi, d := range classified {
		fp := Fingerprint(unique[gi])
		d.Fingerprint = fp
		d.DispositionSource = dispositionSourceLLM
		stored := d
		result[gi] = &stored
		cache.Set(fp, stored)
	}

	// NR fallback for any LLM-bound finding the model never classified.
	for _, i := range toLLM {
		if result[i] != nil {
			continue
		}
		fp := Fingerprint(unique[i])
		result[i] = &FindingDisposition{
			FindingIndex:      i,
			Disposition:       "Needs Manual Review",
			Rationale:         nrFallbackRationale,
			Confidence:        "low",
			DispositionSource: dispositionSourceNRFallback,
			Fingerprint:       fp,
		}
	}

	cache.Save()

	disps := make([]FindingDisposition, n)
	for i := range result {
		result[i].FindingIndex = i
		result[i].FindingID = unique[i].VulnID
		disps[i] = *result[i]
	}
	return tm, disps, nil
}

// classifyWithLLM classifies the given unique-index targets in bounded-concurrency
// batches. It returns a map keyed by unique index. Transport/provider errors are
// fatal (returned); malformed responses are tolerated (left for the NR fallback).
func classifyWithLLM(ctx context.Context, tmSummary string, unique []EnrichedFinding, targets []int, llmClient llm.Client, usage *llm.Usage) (map[int]FindingDisposition, error) {
	out := make(map[int]FindingDisposition)
	if len(targets) == 0 {
		return out, nil
	}

	var batches [][]int
	for off := 0; off < len(targets); off += threatModelBatchSize {
		batches = append(batches, targets[off:min(off+threatModelBatchSize, len(targets))])
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		mu       sync.Mutex // guards out
		usageMu  sync.Mutex // guards usage
		errMu    sync.Mutex // guards firstErr
		firstErr error
		wg       sync.WaitGroup
	)
	sem := make(chan struct{}, triageConcurrency())

	addUsageSafe := func(u llm.Usage) {
		usageMu.Lock()
		addUsage(usage, u)
		usageMu.Unlock()
	}
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
		go func(batch []int) {
			defer wg.Done()
			defer func() { <-sem }()

			subset := make([]EnrichedFinding, len(batch))
			for li, gi := range batch {
				subset[li] = unique[gi]
			}
			local := classifyBatchWithRetry(ctx, tmSummary, subset, llmClient, addUsageSafe, setErr)

			mu.Lock()
			for li, d := range local {
				out[batch[li]] = d
			}
			mu.Unlock()
		}(batch)
	}
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	return out, nil
}

// classifyBatchWithRetry classifies a single batch, retrying omitted findings up
// to threatModelMaxRetries times. Returns a map keyed by LOCAL batch index.
func classifyBatchWithRetry(ctx context.Context, tmSummary string, subset []EnrichedFinding, llmClient llm.Client, addUsageSafe func(llm.Usage), setErr func(error)) map[int]FindingDisposition {
	res := make(map[int]FindingDisposition)

	pass := func(items []EnrichedFinding, localMap []int) {
		disps, u, err := classifyBatchLLM(ctx, tmSummary, items, llmClient)
		addUsageSafe(u)
		if err != nil {
			// Malformed responses are tolerated (NR fallback covers them); only
			// transport/provider failures abort the whole pipeline.
			if !errors.Is(err, errThreatModelParse) {
				setErr(err)
			}
			return
		}
		for _, d := range disps {
			if d.FindingIndex < 0 || d.FindingIndex >= len(items) {
				continue
			}
			if !isSupportedDisposition(d.Disposition) {
				continue
			}
			gi := localMap[d.FindingIndex]
			if _, done := res[gi]; done {
				continue
			}
			res[gi] = FindingDisposition{
				Disposition: d.Disposition,
				Rationale:   d.Rationale,
				Confidence:  normalizeConfidence(d.Confidence),
			}
		}
	}

	full := make([]int, len(subset))
	for i := range full {
		full[i] = i
	}
	pass(subset, full)

	for attempt := 0; attempt < threatModelMaxRetries; attempt++ {
		var missing []int
		for i := range subset {
			if _, done := res[i]; !done {
				missing = append(missing, i)
			}
		}
		if len(missing) == 0 {
			break
		}
		items := make([]EnrichedFinding, len(missing))
		for j, li := range missing {
			items[j] = subset[li]
		}
		pass(items, missing)
	}
	return res
}

// classifyBatchLLM sends one batch to the LLM using the SecureCoder classification
// prompt (which references the prebuilt threat model and the MUST/MUST NOT
// ruleset) and returns the raw per-finding dispositions.
func classifyBatchLLM(ctx context.Context, tmSummary string, batch []EnrichedFinding, llmClient llm.Client) ([]rawDisposition, llm.Usage, error) {
	findingsJSON, _ := json.MarshalIndent(batch, "", "  ")
	userPrompt := fmt.Sprintf(prompts.ClassificationUserPromptTemplate, tmSummary, len(batch), string(findingsJSON))

	messages := []llm.Message{
		{Role: "system", Content: prompts.ClassificationSystemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, u, err := llmClient.Chat(ctx, messages)
	if err != nil {
		return nil, u, fmt.Errorf("classification LLM call failed: %w", err)
	}

	jsonText := extractJSON(response)
	var raw struct {
		FindingDispositions []struct {
			FindingIndex int    `json:"finding_index"`
			Disposition  string `json:"disposition"`
			Confidence   string `json:"confidence"`
			Rationale    string `json:"rationale"`
		} `json:"finding_dispositions"`
	}
	if err := json.Unmarshal([]byte(jsonText), &raw); err != nil {
		return nil, u, fmt.Errorf("%w: %v", errThreatModelParse, err)
	}

	disps := make([]rawDisposition, 0, len(raw.FindingDispositions))
	for _, d := range raw.FindingDispositions {
		disps = append(disps, rawDisposition{
			FindingIndex: d.FindingIndex,
			Disposition:  d.Disposition,
			Confidence:   d.Confidence,
			Rationale:    d.Rationale,
		})
	}
	return disps, u, nil
}

// threatModelSummary renders a compact text view of the threat model for use as
// classification context (keeps SecureCoder methodology without resending the
// full repository to every batch).
func threatModelSummary(tm *ThreatModel) string {
	if tm == nil {
		return "(no threat model available)"
	}
	var sb strings.Builder
	if tm.ComponentOverview != "" {
		sb.WriteString("Overview: " + tm.ComponentOverview + "\n")
	}
	if tm.TrustBoundaries.Authentication != "" || tm.TrustBoundaries.Authorization != "" {
		sb.WriteString(fmt.Sprintf("Trust boundaries: auth=%s; authz=%s; implicit=%s\n",
			tm.TrustBoundaries.Authentication, tm.TrustBoundaries.Authorization, tm.TrustBoundaries.ImplicitTrust))
	}
	if len(tm.EntryPoints) > 0 {
		sb.WriteString("Entry points:\n")
		for _, e := range tm.EntryPoints {
			sb.WriteString(fmt.Sprintf("  - %s (%s), trusted=%v, validation=%s\n", e.Endpoint, e.Type, e.Trusted, e.Validation))
		}
	}
	if len(tm.PriorityAreas) > 0 {
		sb.WriteString("Priority areas: " + strings.Join(tm.PriorityAreas, "; ") + "\n")
	}
	return strings.TrimSpace(sb.String())
}

func normalizeConfidence(c string) string {
	switch strings.ToLower(strings.TrimSpace(c)) {
	case "high":
		return "high"
	case "medium", "med":
		return "medium"
	case "low":
		return "low"
	default:
		return ""
	}
}

// triageConcurrency returns the number of parallel classification workers.
func triageConcurrency() int {
	if v := strings.TrimSpace(os.Getenv("AITRIAGE_CONCURRENCY")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			return n
		}
	}
	return 4
}

// llmBudget returns the max number of unique findings to send to the LLM.
// A negative value (the default) means unlimited.
func llmBudget() int {
	if v := strings.TrimSpace(os.Getenv("AITRIAGE_LLM_BUDGET")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return -1
}

const budgetRationale = "Exceeded the LLM triage budget; defaulting to Needs Manual Review for safety (never auto-suppressed)."

// logTriageMetrics prints a concise breakdown of how dispositions were produced.
func logTriageMetrics(findings, unique []EnrichedFinding, uniqueDisps []FindingDisposition) {
	bySource := map[string]int{}
	for _, d := range uniqueDisps {
		bySource[d.DispositionSource]++
	}
	fmt.Fprintf(os.Stderr, "   🔎 Triage scale: %d findings → %d unique (%d deduped) | sources: %d llm, %d cache, %d deterministic, %d nr-fallback\n",
		len(findings), len(unique), len(findings)-len(unique),
		bySource[dispositionSourceLLM], bySource[dispositionSourceCache],
		bySource[dispositionSourceDeterministic], bySource[dispositionSourceNRFallback])
}
