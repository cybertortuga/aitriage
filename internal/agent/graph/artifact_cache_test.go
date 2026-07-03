package graph

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
)

func TestArtifactCacheDisabledByDefault(t *testing.T) {
	t.Setenv("AITRIAGE_ARTIFACT_CACHE_DIR", "")
	t.Setenv("AITRIAGE_CACHE_DIR", "")

	cache := newArtifactBundleCache()
	if cache.enabled || cache.Stats().Enabled {
		t.Fatalf("artifact cache should be disabled by default: %+v", cache.Stats())
	}
}

func TestArtifactCacheStoreAndExactHit(t *testing.T) {
	t.Setenv("AITRIAGE_ARTIFACT_CACHE_DIR", t.TempDir())
	state := artifactCacheTestState("True Positive")
	state.ReportMarkdown = "# report"
	state.AIFixSpec = "# fixspec"
	state.PoCResults = []PoCResult{{VulnerabilityType: "x", Severity: "HIGH", Conclusion: "Needs Manual Review"}}

	key, miss := buildArtifactCacheKey(state)
	if miss != "" {
		t.Fatalf("buildArtifactCacheKey miss = %q", miss)
	}
	cache := newArtifactBundleCache()
	cache.Store(state, key)
	cache.Save()
	if stats := cache.Stats(); stats.Stores != 1 || !stats.Saved {
		t.Fatalf("store stats = %+v, want one saved store", stats)
	}

	reloaded := newArtifactBundleCache()
	restored := &AgentState{}
	if !reloaded.Restore(restored, key) {
		t.Fatalf("expected artifact cache hit, stats=%+v", reloaded.Stats())
	}
	if restored.ReportMarkdown != "# report" || restored.AIFixSpec != "# fixspec" || len(restored.PoCResults) != 1 {
		t.Fatalf("restored state = %+v, want full bundle", restored)
	}
	if stats := reloaded.Stats(); !stats.ExactHit || !stats.RestoredPoC || !stats.RestoredReport || !stats.RestoredFixSpec {
		t.Fatalf("restore stats = %+v, want exact restored bundle", stats)
	}
}

func TestArtifactCacheKeyInvalidatesOnDispositionAndPolicyChange(t *testing.T) {
	t.Setenv("AITRIAGE_LLM_PROVIDER", "provider-a")
	t.Setenv("AITRIAGE_MODEL", "model-a")
	base := artifactCacheTestState("True Positive")
	key1, miss := buildArtifactCacheKey(base)
	if miss != "" {
		t.Fatalf("base key miss = %q", miss)
	}

	changedDisposition := artifactCacheTestState("Needs Manual Review")
	key2, miss := buildArtifactCacheKey(changedDisposition)
	if miss != "" {
		t.Fatalf("changed disposition key miss = %q", miss)
	}
	if key1 == key2 {
		t.Fatal("artifact cache key should change when disposition changes")
	}

	changedPolicy := artifactCacheTestState("True Positive")
	changedPolicy.Policy = healthcheck.Policy{Profile: "strict", FailOn: healthcheck.FailOnAny}
	key3, miss := buildArtifactCacheKey(changedPolicy)
	if miss != "" {
		t.Fatalf("changed policy key miss = %q", miss)
	}
	if key1 == key3 {
		t.Fatal("artifact cache key should change when policy changes")
	}

	t.Setenv("AITRIAGE_MODEL", "model-b")
	changedModel := artifactCacheTestState("True Positive")
	key4, miss := buildArtifactCacheKey(changedModel)
	if miss != "" {
		t.Fatalf("changed model key miss = %q", miss)
	}
	if key1 == key4 {
		t.Fatal("artifact cache key should change when model changes")
	}

	t.Setenv("AITRIAGE_LLM_PROVIDER", "provider-b")
	changedProvider := artifactCacheTestState("True Positive")
	key5, miss := buildArtifactCacheKey(changedProvider)
	if miss != "" {
		t.Fatalf("changed provider key miss = %q", miss)
	}
	if key4 == key5 {
		t.Fatal("artifact cache key should change when provider changes")
	}
}

func TestArtifactCacheKeyIgnoresGeneratedVulnIDs(t *testing.T) {
	base := artifactCacheTestState("True Positive")
	key1, miss := buildArtifactCacheKey(base)
	if miss != "" {
		t.Fatalf("base key miss = %q", miss)
	}

	renamed := artifactCacheTestState("True Positive")
	renamed.EnrichedFindings[0].VulnID = "CS-AUTH-042"
	renamed.FindingDispositions[0].FindingID = "CS-AUTH-042"
	key2, miss := buildArtifactCacheKey(renamed)
	if miss != "" {
		t.Fatalf("renamed key miss = %q", miss)
	}
	if key1 != key2 {
		t.Fatal("artifact cache key must not depend on generated CS-* IDs")
	}
}

func TestArtifactCacheRestoreDistinguishesEmptyCacheFromKeyMiss(t *testing.T) {
	t.Setenv("AITRIAGE_ARTIFACT_CACHE_DIR", t.TempDir())

	empty := newArtifactBundleCache()
	if empty.Restore(&AgentState{}, "v2:deadbeef") {
		t.Fatal("empty cache should miss")
	}
	if reason := empty.Stats().MissReason; reason != "no_entries_loaded" {
		t.Fatalf("empty cache miss reason = %q, want no_entries_loaded", reason)
	}

	state := artifactCacheTestState("True Positive")
	state.ReportMarkdown = "# report"
	state.AIFixSpec = "# fixspec"
	key, miss := buildArtifactCacheKey(state)
	if miss != "" {
		t.Fatalf("key miss = %q", miss)
	}
	warm := newArtifactBundleCache()
	warm.Store(state, key)
	warm.Save()

	populated := newArtifactBundleCache()
	if populated.Restore(&AgentState{}, "v2:other-key") {
		t.Fatal("unknown key should miss")
	}
	if reason := populated.Stats().MissReason; reason != "key_miss" {
		t.Fatalf("populated cache miss reason = %q, want key_miss", reason)
	}
	if populated.Stats().Key != "" {
		t.Fatalf("stats key is set by the orchestrator, not Restore: %q", populated.Stats().Key)
	}
}

func TestArtifactCacheCorruptCacheIgnored(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AITRIAGE_ARTIFACT_CACHE_DIR", dir)
	if err := os.WriteFile(filepath.Join(dir, "artifact_bundle_cache.json"), []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write corrupt cache: %v", err)
	}

	cache := newArtifactBundleCache()
	if !cache.Stats().CorruptCacheIgnored {
		t.Fatalf("corrupt cache should be ignored, stats=%+v", cache.Stats())
	}
}

func TestArtifactCacheSkipsSensitiveBundle(t *testing.T) {
	t.Setenv("AITRIAGE_ARTIFACT_CACHE_DIR", t.TempDir())
	state := artifactCacheTestState("True Positive")
	state.ReportMarkdown = "token sk-aaaaaaaaaaaaaaaaaaaaaaaa should not be cached"
	state.AIFixSpec = "# fixspec"
	key, miss := buildArtifactCacheKey(state)
	if miss != "" {
		t.Fatalf("key miss = %q", miss)
	}

	cache := newArtifactBundleCache()
	cache.Store(state, key)
	if stats := cache.Stats(); stats.Stores != 0 || stats.SkippedSensitive != 1 {
		t.Fatalf("sensitive store stats = %+v, want skipped", stats)
	}
}

func TestArtifactCacheStoresSafePackageNameBundle(t *testing.T) {
	t.Setenv("AITRIAGE_ARTIFACT_CACHE_DIR", t.TempDir())
	state := artifactCacheTestState("True Positive")
	state.ReportMarkdown = "Use flask-limiter or slowapi for rate limiting."
	state.AIFixSpec = "# fixspec"
	key, miss := buildArtifactCacheKey(state)
	if miss != "" {
		t.Fatalf("key miss = %q", miss)
	}

	cache := newArtifactBundleCache()
	cache.Store(state, key)
	cache.Save()
	if stats := cache.Stats(); stats.Stores != 1 || stats.SkippedSensitive != 0 || !stats.Saved {
		t.Fatalf("safe store stats = %+v, want saved store", stats)
	}
}

func TestRunArtifactCacheWarmThenExactHitSkipsSecondaryLLMStages(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AITRIAGE_VERDICT_CACHE_DIR", dir)
	t.Setenv("AITRIAGE_ARTIFACT_CACHE_DIR", dir)
	pinConcurrency(t)

	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "app.go"), []byte("package app\n"), 0o644); err != nil {
		t.Fatalf("write project file: %v", err)
	}

	warmLLM := newArtifactRunLLM(t)
	warmState := artifactRunState(project)
	if err := Run(context.Background(), warmState, warmLLM); err != nil {
		t.Fatalf("warm Run() error = %v", err)
	}
	for _, stage := range []string{"poc", "report", "fixspec"} {
		if warmLLM.calls[stage] == 0 {
			t.Fatalf("warm run did not call %s stage: calls=%v", stage, warmLLM.calls)
		}
	}
	if warmState.ArtifactCacheStats.Stores != 1 || !warmState.ArtifactCacheStats.Saved {
		t.Fatalf("warm artifact stats = %+v, want saved store", warmState.ArtifactCacheStats)
	}

	hitLLM := newArtifactRunLLM(t)
	hitState := artifactRunState(project)
	if err := Run(context.Background(), hitState, hitLLM); err != nil {
		t.Fatalf("hit Run() error = %v", err)
	}
	for _, stage := range []string{"poc", "report", "fixspec"} {
		if hitLLM.calls[stage] != 0 {
			t.Fatalf("artifact hit should skip %s LLM stage, calls=%v", stage, hitLLM.calls)
		}
	}
	if !hitState.ArtifactCacheStats.ExactHit || !hitState.ArtifactCacheStats.RestoredReport || !hitState.ArtifactCacheStats.RestoredFixSpec {
		t.Fatalf("hit artifact stats = %+v, want exact restored hit", hitState.ArtifactCacheStats)
	}
	if hitState.ReportMarkdown != "# report\n" || hitState.AIFixSpec != "# fixspec\n" {
		t.Fatalf("hit state did not restore report/fixspec: report=%q fixspec=%q", hitState.ReportMarkdown, hitState.AIFixSpec)
	}
}

func artifactCacheTestState(disposition string) *AgentState {
	finding := EnrichedFinding{ID: "R1", VulnID: "CS-MISC-001", Type: "core", Severity: "HIGH", File: "app.go", Line: 1, Message: "Issue"}
	fp := Fingerprint(finding)
	return &AgentState{
		EnrichedFindings: []EnrichedFinding{finding},
		FindingDispositions: []FindingDisposition{{
			FindingIndex: 0,
			FindingID:    "CS-MISC-001",
			Disposition:  disposition,
			Rationale:    "r",
			Confidence:   "high",
			Fingerprint:  fp,
		}},
		Policy: healthcheck.Policy{FailOn: healthcheck.FailOnNever},
	}
}

func artifactRunState(project string) *AgentState {
	return &AgentState{
		ProjectPath: project,
		CoreFindings: []core.CheckResult{{
			ID:       "R1",
			Name:     "Issue",
			Status:   core.Absent,
			Evidence: "unsafe behavior",
			Severity: "HIGH",
			File:     "app.go",
			Line:     1,
		}},
		Policy: healthcheck.Policy{FailOn: healthcheck.FailOnNever},
	}
}

type artifactRunLLM struct {
	t     *testing.T
	calls map[string]int
}

func newArtifactRunLLM(t *testing.T) *artifactRunLLM {
	return &artifactRunLLM{t: t, calls: make(map[string]int)}
}

func (m *artifactRunLLM) Chat(ctx context.Context, messages []llm.Message) (string, llm.Usage, error) {
	if len(messages) == 0 {
		m.t.Fatal("empty messages")
	}
	system := messages[0].Content
	switch {
	case strings.Contains(system, "Current Task: Threat Model & Finding Classification"):
		m.calls["threat_model"]++
		return `{"component_overview":"test","entry_points":[],"trust_boundaries":{"authentication":"n/a","authorization":"n/a","implicit_trust":"n/a"},"sensitive_data_paths":[],"privileged_actions":[],"priority_areas":[],"finding_dispositions":[{"finding_index":0,"disposition":"True Positive","rationale":"r"}]}`, llm.Usage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2}, nil
	case strings.Contains(system, "Current Task: Finding Classification"):
		m.calls["classification"]++
		batch := parseClassificationPromptFindings(messages[1].Content)
		return classifyAll(batch, "True Positive"), llm.Usage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2}, nil
	case strings.Contains(system, "Current Task: PoC Verification"):
		m.calls["poc"]++
		return pocResultsJSON(1), llm.Usage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2}, nil
	case strings.Contains(system, "Current Task: Compile Security Report"):
		m.calls["report"]++
		return "# report\n", llm.Usage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2}, nil
	case strings.Contains(system, "Generate AI IDE Fix Plan"):
		m.calls["fixspec"]++
		return "# fixspec\n", llm.Usage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2}, nil
	default:
		m.t.Fatalf("unexpected LLM prompt: %s", system)
		return "", llm.Usage{}, nil
	}
}

func parseClassificationPromptFindings(user string) []classificationPromptFinding {
	i := strings.Index(user, "[")
	if i < 0 {
		return nil
	}
	dec := json.NewDecoder(strings.NewReader(user[i:]))
	var findings []classificationPromptFinding
	_ = dec.Decode(&findings)
	return findings
}
