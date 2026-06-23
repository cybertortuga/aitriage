package graph

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ── Layer 2: Verdict Cache (keyed by fingerprint) ────────────────────────────
//
// In CI, the same findings recur on every push/PR. Caching the LLM verdict by
// fingerprint means a re-run costs (almost) zero LLM calls. This is the single
// biggest cost lever at scale.
//
// Correctness: the cache key embeds the model label and a schema version, so a
// model upgrade or schema change transparently invalidates stale verdicts.
//
// The cache is OFF unless AITRIAGE_CACHE_DIR is set, so default behaviour is
// unchanged and tests are hermetic.

const verdictCacheSchemaVersion = 2

type cachedVerdict struct {
	Disposition string               `json:"disposition"`
	Rationale   string               `json:"rationale"`
	Confidence  string               `json:"confidence"`
	Evidence    *DispositionEvidence `json:"evidence,omitempty"`
}

type verdictCache struct {
	enabled bool
	path    string
	model   string
	entries map[string]cachedVerdict
	dirty   bool
}

func newVerdictCache(model string) *verdictCache {
	c := &verdictCache{model: model, entries: make(map[string]cachedVerdict)}
	dir := strings.TrimSpace(os.Getenv("AITRIAGE_CACHE_DIR"))
	if dir == "" {
		return c // disabled
	}
	c.enabled = true
	c.path = filepath.Join(dir, "triage_cache.json")
	c.load()
	return c
}

func (c *verdictCache) key(fingerprint string) string {
	return fmt.Sprintf("%s|v%d|%s", c.model, verdictCacheSchemaVersion, fingerprint)
}

func (c *verdictCache) load() {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return // missing cache is fine
	}
	var entries map[string]cachedVerdict
	if err := json.Unmarshal(data, &entries); err != nil {
		return // corrupt cache is ignored, not fatal
	}
	c.entries = entries
}

// Get returns a previously cached disposition for a fingerprint, if present.
func (c *verdictCache) Get(fingerprint string) (FindingDisposition, bool) {
	if !c.enabled {
		return FindingDisposition{}, false
	}
	v, ok := c.entries[c.key(fingerprint)]
	if !ok {
		return FindingDisposition{}, false
	}
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
	c.entries[c.key(fingerprint)] = cachedVerdict{
		Disposition: d.Disposition,
		Rationale:   d.Rationale,
		Confidence:  d.Confidence,
		Evidence:    d.Evidence,
	}
	c.dirty = true
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
	_ = os.WriteFile(c.path, data, 0o644)
}
