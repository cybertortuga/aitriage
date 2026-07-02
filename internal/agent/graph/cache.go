package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/cybertortuga/aitriage/internal/agent/prompts"
	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
	"github.com/cybertortuga/aitriage/rules"
)

// ── Layer 2: Verdict Cache (keyed by fingerprint) ────────────────────────────
//
// In CI, the same findings recur on every push/PR. Caching the LLM verdict by
// fingerprint means a re-run costs (almost) zero LLM calls. This is the single
// biggest cost lever at scale.
//
// Correctness: the cache key embeds provider/model/prompt/rules/policy namespace
// data and a schema version, so material changes transparently invalidate stale
// verdicts.
//
// The cache is OFF unless AITRIAGE_CACHE_DIR is set, so default behaviour is
// unchanged and tests are hermetic.

const verdictCacheSchemaVersion = 3
const verdictCacheRulesDigestVersion = "rules-v1"

type verdictCacheKeyContext struct {
	SchemaVersion   int      `json:"schema_version"`
	Provider        string   `json:"provider"`
	Model           string   `json:"model"`
	BaseURLHash     string   `json:"base_url_hash"`
	AITriageVersion string   `json:"aitriage_version"`
	PromptVersion   string   `json:"prompt_version"`
	RulesDigest     string   `json:"rules_digest"`
	PolicyProfile   string   `json:"policy_profile"`
	PolicyFailOn    string   `json:"policy_fail_on"`
	MinimumScore    int      `json:"minimum_score"`
	MaxCritical     int      `json:"max_critical"`
	MaxHigh         int      `json:"max_high"`
	MaxMedium       int      `json:"max_medium"`
	BlockSources    []string `json:"block_sources,omitempty"`
	BlockClasses    []string `json:"block_classes,omitempty"`
}

type verdictCacheOption func(*verdictCacheKeyContext)

func withVerdictCachePolicy(policy healthcheck.Policy) verdictCacheOption {
	return func(ctx *verdictCacheKeyContext) {
		ctx.PolicyProfile = strings.TrimSpace(policy.Profile)
		ctx.PolicyFailOn = strings.TrimSpace(policy.FailOn)
		ctx.MinimumScore = policy.MinimumScore
		ctx.MaxCritical = policy.MaxCritical
		ctx.MaxHigh = policy.MaxHigh
		ctx.MaxMedium = policy.MaxMedium
		ctx.BlockSources = sortedCopy(policy.BlockSources)
		ctx.BlockClasses = sortedCopy(policy.BlockClasses)
	}
}

type cachedVerdict struct {
	Disposition string               `json:"disposition"`
	Rationale   string               `json:"rationale"`
	Confidence  string               `json:"confidence"`
	Evidence    *DispositionEvidence `json:"evidence,omitempty"`
}

type verdictCache struct {
	enabled   bool
	path      string
	namespace string
	entries   map[string]cachedVerdict
	dirty     bool
	stats     VerdictCacheStats
}

type VerdictCacheStats struct {
	Enabled                   bool `json:"enabled"`
	LoadedEntries             int  `json:"loaded_entries"`
	Hits                      int  `json:"hits"`
	Misses                    int  `json:"misses"`
	Stores                    int  `json:"stores"`
	SkippedSensitive          int  `json:"skipped_sensitive"`
	InvalidatedFalsePositives int  `json:"invalidated_false_positives"`
	CorruptCacheIgnored       bool `json:"corrupt_cache_ignored"`
	Saved                     bool `json:"saved"`
}

func newVerdictCache(model string, opts ...verdictCacheOption) *verdictCache {
	ctx := defaultVerdictCacheKeyContext(model)
	for _, opt := range opts {
		opt(&ctx)
	}
	return newVerdictCacheWithContext(ctx)
}

func newVerdictCacheWithContext(ctx verdictCacheKeyContext) *verdictCache {
	c := &verdictCache{namespace: ctx.namespace(), entries: make(map[string]cachedVerdict)}
	dir := verdictCacheDir()
	if dir == "" {
		return c // disabled
	}
	c.enabled = true
	c.stats.Enabled = true
	c.path = filepath.Join(dir, "triage_cache.json")
	c.load()
	return c
}

func (c *verdictCache) key(fingerprint string) string {
	return fmt.Sprintf("%s|%s", c.namespace, fingerprint)
}

func (c *verdictCache) load() {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return // missing cache is fine
	}
	var entries map[string]cachedVerdict
	if err := json.Unmarshal(data, &entries); err != nil {
		c.stats.CorruptCacheIgnored = true
		return // corrupt cache is ignored, not fatal
	}
	c.entries = entries
	c.stats.LoadedEntries = len(entries)
}

// Get returns a previously cached disposition for a fingerprint, if present.
func (c *verdictCache) Get(fingerprint string) (FindingDisposition, bool) {
	if !c.enabled {
		return FindingDisposition{}, false
	}
	v, ok := c.entries[c.key(fingerprint)]
	if !ok {
		c.stats.Misses++
		return FindingDisposition{}, false
	}
	c.stats.Hits++
	return FindingDisposition{
		Disposition: v.Disposition,
		Rationale:   v.Rationale,
		Confidence:  v.Confidence,
		Evidence:    v.Evidence,
	}, true
}

// Set records a freshly produced verdict for a fingerprint.
func (c *verdictCache) Set(fingerprint string, d FindingDisposition) {
	if !c.enabled || fingerprint == "" {
		return
	}
	v, ok := cacheableVerdict(d)
	if !ok {
		c.stats.SkippedSensitive++
		return
	}
	c.entries[c.key(fingerprint)] = v
	c.dirty = true
	c.stats.Stores++
}

func (c *verdictCache) InvalidateFalsePositive() {
	if !c.enabled {
		return
	}
	c.stats.InvalidatedFalsePositives++
}

func cacheableVerdict(d FindingDisposition) (cachedVerdict, bool) {
	if containsSensitiveCacheValue(d.Rationale) {
		return cachedVerdict{}, false
	}
	if d.Evidence != nil && containsSensitiveCacheValue(d.Evidence.Observed) {
		return cachedVerdict{}, false
	}
	return cachedVerdict{
		Disposition: d.Disposition,
		Rationale:   d.Rationale,
		Confidence:  d.Confidence,
		Evidence:    d.Evidence,
	}, true
}

func containsSensitiveCacheValue(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	lower := strings.ToLower(value)
	for _, marker := range []string{
		"-----begin ",
		"sk-",
		"ghp_",
		"github_pat_",
		"xoxb-",
		"xoxp-",
		"eyj",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return strings.Contains(value, "AKIA")
}

// Save persists the cache to disk (best-effort; failures are non-fatal).
func (c *verdictCache) Save() {
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

func (c *verdictCache) Stats() VerdictCacheStats {
	if c == nil {
		return VerdictCacheStats{}
	}
	return c.stats
}

func verdictCacheDir() string {
	if dir := strings.TrimSpace(os.Getenv("AITRIAGE_VERDICT_CACHE_DIR")); dir != "" {
		return dir
	}
	return strings.TrimSpace(os.Getenv("AITRIAGE_CACHE_DIR"))
}

func defaultVerdictCacheKeyContext(model string) verdictCacheKeyContext {
	return verdictCacheKeyContext{
		SchemaVersion:   verdictCacheSchemaVersion,
		Provider:        cacheField(os.Getenv("AITRIAGE_LLM_PROVIDER")),
		Model:           cacheField(firstNonEmpty(model, os.Getenv("AITRIAGE_LLM_MODEL"), os.Getenv("AITRIAGE_MODEL"))),
		BaseURLHash:     hashCacheField(os.Getenv("AITRIAGE_LLM_BASE_URL")),
		AITriageVersion: cacheField(os.Getenv("AITRIAGE_VERSION")),
		PromptVersion:   prompts.SecureCoderPromptVersion,
		RulesDigest:     embeddedRulesDigest(),
		PolicyProfile:   cacheField(os.Getenv("AITRIAGE_HEALTH_PROFILE")),
		PolicyFailOn:    cacheField(os.Getenv("AITRIAGE_FAIL_ON")),
	}
}

func (ctx verdictCacheKeyContext) namespace() string {
	normalized := ctx
	normalized.Provider = cacheField(normalized.Provider)
	normalized.Model = cacheField(normalized.Model)
	normalized.BaseURLHash = cacheField(normalized.BaseURLHash)
	normalized.AITriageVersion = cacheField(normalized.AITriageVersion)
	normalized.PromptVersion = cacheField(normalized.PromptVersion)
	normalized.RulesDigest = cacheField(normalized.RulesDigest)
	normalized.PolicyProfile = cacheField(normalized.PolicyProfile)
	normalized.PolicyFailOn = cacheField(normalized.PolicyFailOn)
	normalized.BlockSources = sortedCopy(normalized.BlockSources)
	normalized.BlockClasses = sortedCopy(normalized.BlockClasses)

	data, _ := json.Marshal(normalized)
	sum := sha256.Sum256(data)
	return fmt.Sprintf("v%d:%s", normalized.SchemaVersion, hex.EncodeToString(sum[:]))
}

func cacheField(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "default"
	}
	return value
}

func hashCacheField(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "default"
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func sortedCopy(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

var (
	embeddedRulesDigestOnce  sync.Once
	embeddedRulesDigestValue string
)

func embeddedRulesDigest() string {
	embeddedRulesDigestOnce.Do(func() {
		h := sha256.New()
		h.Write([]byte(verdictCacheRulesDigestVersion))

		var paths []string
		_ = fs.WalkDir(rules.FS, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
				paths = append(paths, path)
			}
			return nil
		})
		sort.Strings(paths)
		for _, path := range paths {
			data, err := rules.FS.ReadFile(path)
			if err != nil {
				continue
			}
			h.Write([]byte(path))
			h.Write([]byte{0})
			h.Write(data)
			h.Write([]byte{0})
		}
		embeddedRulesDigestValue = hex.EncodeToString(h.Sum(nil))
	})
	return embeddedRulesDigestValue
}
