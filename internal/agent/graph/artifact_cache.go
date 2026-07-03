package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cybertortuga/aitriage/internal/agent/prompts"
)

const artifactCacheSchemaVersion = 1

type artifactCacheKeyContext struct {
	SchemaVersion       int      `json:"schema_version"`
	VerdictNamespace    string   `json:"verdict_namespace"`
	PoCPromptVersion    string   `json:"poc_prompt_version"`
	ReportPromptVersion string   `json:"report_prompt_version"`
	FixSpecVersion      string   `json:"fixspec_prompt_version"`
	FindingFingerprints []string `json:"finding_fingerprints"`
	DispositionHashes   []string `json:"disposition_hashes"`
}

type artifactBundleCache struct {
	enabled bool
	path    string
	entries map[string]cachedArtifactBundle
	dirty   bool
	stats   ArtifactCacheStats
}

type ArtifactCacheStats struct {
	Enabled             bool   `json:"enabled"`
	LoadedEntries       int    `json:"loaded_entries,omitempty"`
	ExactHit            bool   `json:"exact_hit"`
	MissReason          string `json:"miss_reason,omitempty"`
	RestoredPoC         bool   `json:"restored_poc"`
	RestoredReport      bool   `json:"restored_report"`
	RestoredFixSpec     bool   `json:"restored_fixspec"`
	Stores              int    `json:"stores"`
	SkippedSensitive    int    `json:"skipped_sensitive"`
	CorruptCacheIgnored bool   `json:"corrupt_cache_ignored"`
	Saved               bool   `json:"saved"`
}

type cachedArtifactBundle struct {
	SchemaVersion   int                  `json:"schema_version"`
	Key             string               `json:"key"`
	CreatedAt       string               `json:"created_at"`
	PoCResults      []PoCResult          `json:"poc_results"`
	ReportMarkdown  string               `json:"report_markdown"`
	FixSpecMarkdown string               `json:"fixspec_markdown"`
	Counts          artifactBundleCounts `json:"counts"`
}

type artifactBundleCounts struct {
	Findings     int `json:"findings"`
	Dispositions int `json:"dispositions"`
	PoCResults   int `json:"poc_results"`
}

func newArtifactBundleCache() *artifactBundleCache {
	c := &artifactBundleCache{entries: make(map[string]cachedArtifactBundle)}
	dir := artifactCacheDir()
	if dir == "" {
		return c
	}
	c.enabled = true
	c.stats.Enabled = true
	c.path = filepath.Join(dir, "artifact_bundle_cache.json")
	c.load()
	return c
}

func artifactCacheDir() string {
	if dir := strings.TrimSpace(os.Getenv("AITRIAGE_ARTIFACT_CACHE_DIR")); dir != "" {
		return dir
	}
	return strings.TrimSpace(os.Getenv("AITRIAGE_CACHE_DIR"))
}

func (c *artifactBundleCache) load() {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return
	}
	var entries map[string]cachedArtifactBundle
	if err := json.Unmarshal(data, &entries); err != nil {
		c.stats.CorruptCacheIgnored = true
		return
	}
	c.entries = entries
	c.stats.LoadedEntries = len(entries)
}

func (c *artifactBundleCache) Restore(state *AgentState, key string) bool {
	if !c.enabled {
		return false
	}
	if key == "" {
		c.stats.MissReason = "empty_key"
		return false
	}
	entry, ok := c.entries[key]
	if !ok {
		c.stats.MissReason = "key_miss"
		return false
	}
	if entry.SchemaVersion != artifactCacheSchemaVersion || entry.Key != key {
		c.stats.MissReason = "stale_or_mismatched_entry"
		return false
	}
	if strings.TrimSpace(entry.ReportMarkdown) == "" || strings.TrimSpace(entry.FixSpecMarkdown) == "" {
		c.stats.MissReason = "incomplete_bundle"
		return false
	}
	state.PoCResults = append([]PoCResult(nil), entry.PoCResults...)
	state.ReportMarkdown = entry.ReportMarkdown
	state.AIFixSpec = entry.FixSpecMarkdown
	c.stats.ExactHit = true
	c.stats.RestoredPoC = true
	c.stats.RestoredReport = true
	c.stats.RestoredFixSpec = true
	return true
}

func (c *artifactBundleCache) Store(state *AgentState, key string) {
	if !c.enabled {
		return
	}
	if key == "" {
		c.stats.MissReason = "empty_key"
		return
	}
	if strings.TrimSpace(state.ReportMarkdown) == "" || strings.TrimSpace(state.AIFixSpec) == "" {
		c.stats.MissReason = "incomplete_bundle"
		return
	}
	if artifactBundleContainsSensitiveValue(state) {
		c.stats.SkippedSensitive++
		c.stats.MissReason = "sensitive_bundle_skipped"
		return
	}
	c.entries[key] = cachedArtifactBundle{
		SchemaVersion:   artifactCacheSchemaVersion,
		Key:             key,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
		PoCResults:      append([]PoCResult(nil), state.PoCResults...),
		ReportMarkdown:  state.ReportMarkdown,
		FixSpecMarkdown: state.AIFixSpec,
		Counts: artifactBundleCounts{
			Findings:     len(state.EnrichedFindings),
			Dispositions: len(state.FindingDispositions),
			PoCResults:   len(state.PoCResults),
		},
	}
	c.dirty = true
	c.stats.Stores++
}

func (c *artifactBundleCache) Save() {
	if !c.enabled || !c.dirty {
		return
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return
	}
	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return
	}
	if err := os.WriteFile(c.path, data, 0o644); err != nil {
		return
	}
	c.stats.Saved = true
}

func (c *artifactBundleCache) Stats() ArtifactCacheStats {
	if c == nil {
		return ArtifactCacheStats{}
	}
	return c.stats
}

func buildArtifactCacheKey(state *AgentState) (string, string) {
	if state == nil {
		return "", "nil_state"
	}
	if err := validateFindingDispositions(state.FindingDispositions, len(state.EnrichedFindings)); err != nil {
		return "", "incomplete_dispositions"
	}
	verdictCtx := defaultVerdictCacheKeyContext(strings.TrimSpace(os.Getenv("AITRIAGE_MODEL")))
	withVerdictCachePolicy(state.Policy)(&verdictCtx)

	ctx := artifactCacheKeyContext{
		SchemaVersion:       artifactCacheSchemaVersion,
		VerdictNamespace:    verdictCtx.namespace(),
		PoCPromptVersion:    prompts.PoCPromptVersion,
		ReportPromptVersion: prompts.ReportPromptVersion,
		FixSpecVersion:      prompts.FixSpecPromptVersion,
		FindingFingerprints: make([]string, 0, len(state.EnrichedFindings)),
		DispositionHashes:   make([]string, 0, len(state.FindingDispositions)),
	}
	for _, finding := range state.EnrichedFindings {
		ctx.FindingFingerprints = append(ctx.FindingFingerprints, Fingerprint(finding))
	}
	for _, disposition := range state.FindingDispositions {
		ctx.DispositionHashes = append(ctx.DispositionHashes, hashArtifactDisposition(disposition))
	}
	data, _ := json.Marshal(ctx)
	sum := sha256.Sum256(data)
	return fmt.Sprintf("v%d:%s", artifactCacheSchemaVersion, hex.EncodeToString(sum[:])), ""
}

func hashArtifactDisposition(d FindingDisposition) string {
	type artifactDispositionKey struct {
		FindingID   string `json:"finding_id"`
		Fingerprint string `json:"fingerprint"`
		Disposition string `json:"disposition"`
		Confidence  string `json:"confidence,omitempty"`
		Rationale   string `json:"rationale"`
		Evidence    string `json:"evidence,omitempty"`
	}
	evidence := ""
	if d.Evidence != nil {
		data, _ := json.Marshal(d.Evidence)
		evidenceSum := sha256.Sum256(data)
		evidence = hex.EncodeToString(evidenceSum[:])
	}
	payload := artifactDispositionKey{
		FindingID:   strings.TrimSpace(d.FindingID),
		Fingerprint: strings.TrimSpace(d.Fingerprint),
		Disposition: strings.TrimSpace(d.Disposition),
		Confidence:  strings.TrimSpace(d.Confidence),
		Rationale:   strings.TrimSpace(d.Rationale),
		Evidence:    evidence,
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func artifactBundleContainsSensitiveValue(state *AgentState) bool {
	if containsSensitiveCacheValue(state.ReportMarkdown) || containsSensitiveCacheValue(state.AIFixSpec) {
		return true
	}
	data, err := json.Marshal(state.PoCResults)
	if err != nil {
		return true
	}
	return containsSensitiveCacheValue(string(data))
}
