package graph

import (
	"os"
	"strings"
	"testing"
)

func TestVerdictCacheDisabledWithoutEnv(t *testing.T) {
	t.Setenv("AITRIAGE_CACHE_DIR", "")
	t.Setenv("AITRIAGE_VERDICT_CACHE_DIR", "")
	c := newVerdictCache("model-a")
	c.Set("fp1", FindingDisposition{Disposition: "True Positive"})
	if _, ok := c.Get("fp1"); ok {
		t.Fatal("disabled cache must not return entries")
	}
}

func TestVerdictCacheUsesNewVerdictCacheDirEnv(t *testing.T) {
	legacyDir := t.TempDir()
	verdictDir := t.TempDir()
	t.Setenv("AITRIAGE_CACHE_DIR", legacyDir)
	t.Setenv("AITRIAGE_VERDICT_CACHE_DIR", verdictDir)

	c := newVerdictCache("model-a")
	c.Set("fp1", FindingDisposition{Disposition: "True Positive"})
	c.Save()

	if _, err := os.Stat(verdictDir + "/triage_cache.json"); err != nil {
		t.Fatalf("new verdict cache dir was not used: %v", err)
	}
	if _, err := os.Stat(legacyDir + "/triage_cache.json"); err == nil {
		t.Fatal("legacy cache dir must not be used when AITRIAGE_VERDICT_CACHE_DIR is set")
	}
}

func TestVerdictCacheRoundTripAndPersistence(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AITRIAGE_CACHE_DIR", dir)

	c := newVerdictCache("model-a")
	c.Set("fp1", FindingDisposition{Disposition: "False Positive", Rationale: "mitigated", Confidence: "high"})
	if got, ok := c.Get("fp1"); !ok || got.Disposition != "False Positive" {
		t.Fatalf("roundtrip failed: %+v ok=%v", got, ok)
	}
	c.Save()
	if stats := c.Stats(); stats.Stores != 1 || stats.Hits != 1 || !stats.Saved {
		t.Fatalf("cache stats after save = %+v, want one store, one hit, saved", stats)
	}

	// A fresh cache with the same model must load the persisted verdict.
	c2 := newVerdictCache("model-a")
	if got, ok := c2.Get("fp1"); !ok || got.Disposition != "False Positive" {
		t.Fatalf("persisted verdict not reloaded: %+v ok=%v", got, ok)
	}
	if stats := c2.Stats(); stats.LoadedEntries != 1 || stats.Hits != 1 {
		t.Fatalf("reloaded cache stats = %+v, want loaded entry and hit", stats)
	}
}

func TestVerdictCacheInvalidatesOnModelChange(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AITRIAGE_CACHE_DIR", dir)

	c := newVerdictCache("model-a")
	c.Set("fp1", FindingDisposition{Disposition: "True Positive"})
	c.Save()

	// Different model => different key namespace => cache miss (safe invalidation).
	c2 := newVerdictCache("model-b")
	if _, ok := c2.Get("fp1"); ok {
		t.Fatal("verdict must not be reused across different models")
	}
}

func TestVerdictCacheInvalidatesOnProviderBaseURLAndPolicyChange(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AITRIAGE_CACHE_DIR", dir)

	ctx := defaultVerdictCacheKeyContext("model-a")
	ctx.Provider = "openai"
	ctx.BaseURLHash = hashCacheField("https://api.z.ai/api/coding/paas/v4")
	ctx.PolicyProfile = "standard"
	ctx.PolicyFailOn = "any"

	c := newVerdictCacheWithContext(ctx)
	c.Set("fp1", FindingDisposition{Disposition: "True Positive"})
	c.Save()

	providerCtx := ctx
	providerCtx.Provider = "anthropic"
	if _, ok := newVerdictCacheWithContext(providerCtx).Get("fp1"); ok {
		t.Fatal("verdict must not be reused across providers")
	}

	baseURLCtx := ctx
	baseURLCtx.BaseURLHash = hashCacheField("https://token-plan-sgp.xiaomimimo.com/v1")
	if _, ok := newVerdictCacheWithContext(baseURLCtx).Get("fp1"); ok {
		t.Fatal("verdict must not be reused across base URLs")
	}

	policyCtx := ctx
	policyCtx.PolicyProfile = "strict"
	if _, ok := newVerdictCacheWithContext(policyCtx).Get("fp1"); ok {
		t.Fatal("verdict must not be reused across policy profiles")
	}
}

func TestVerdictCacheInvalidatesOnPromptAndRulesVersionChange(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AITRIAGE_CACHE_DIR", dir)

	ctx := defaultVerdictCacheKeyContext("model-a")
	ctx.PromptVersion = "prompt-v1"
	ctx.RulesDigest = "rules-v1"

	c := newVerdictCacheWithContext(ctx)
	c.Set("fp1", FindingDisposition{Disposition: "True Positive"})
	c.Save()

	promptCtx := ctx
	promptCtx.PromptVersion = "prompt-v2"
	if _, ok := newVerdictCacheWithContext(promptCtx).Get("fp1"); ok {
		t.Fatal("verdict must not be reused across prompt versions")
	}

	rulesCtx := ctx
	rulesCtx.RulesDigest = "rules-v2"
	if _, ok := newVerdictCacheWithContext(rulesCtx).Get("fp1"); ok {
		t.Fatal("verdict must not be reused across rules digests")
	}
}

func TestVerdictCacheUsesResolvedLLMModelEnvForNamespace(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AITRIAGE_CACHE_DIR", dir)
	t.Setenv("AITRIAGE_LLM_MODEL", "mimo-v2.5-pro")

	c := newVerdictCache("")
	c.Set("fp1", FindingDisposition{Disposition: "True Positive"})
	c.Save()

	if _, ok := newVerdictCache("").Get("fp1"); !ok {
		t.Fatal("verdict must be reused when resolved LLM model env is unchanged")
	}

	t.Setenv("AITRIAGE_LLM_MODEL", "glm-5.2")
	if _, ok := newVerdictCache("").Get("fp1"); ok {
		t.Fatal("verdict must not be reused across resolved LLM models")
	}
}

func TestVerdictCacheDoesNotPersistRawBaseURL(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AITRIAGE_CACHE_DIR", dir)

	rawBaseURL := "https://example.invalid/tenant/token-in-url"
	ctx := defaultVerdictCacheKeyContext("model-a")
	ctx.Provider = "openai"
	ctx.BaseURLHash = hashCacheField(rawBaseURL)

	c := newVerdictCacheWithContext(ctx)
	c.Set("fp1", FindingDisposition{Disposition: "True Positive"})
	c.Save()

	data, err := os.ReadFile(c.path)
	if err != nil {
		t.Fatalf("read cache file: %v", err)
	}
	if strings.Contains(string(data), rawBaseURL) || strings.Contains(string(data), "token-in-url") {
		t.Fatalf("cache file leaked raw base URL: %s", data)
	}
}

func TestVerdictCacheIgnoresCorruptCache(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AITRIAGE_CACHE_DIR", dir)
	if err := os.WriteFile(dir+"/triage_cache.json", []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write corrupt cache: %v", err)
	}

	c := newVerdictCache("model-a")
	if _, ok := c.Get("fp1"); ok {
		t.Fatal("corrupt cache must not produce a hit")
	}
	if stats := c.Stats(); !stats.CorruptCacheIgnored || stats.Misses != 1 {
		t.Fatalf("corrupt cache stats = %+v, want corrupt ignored and one miss", stats)
	}
	c.Set("fp1", FindingDisposition{Disposition: "True Positive"})
	if got, ok := c.Get("fp1"); !ok || got.Disposition != "True Positive" {
		t.Fatalf("cache should recover after corrupt load, got %+v ok=%v", got, ok)
	}
}

func TestVerdictCacheDoesNotPersistSensitiveEvidence(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AITRIAGE_CACHE_DIR", dir)

	c := newVerdictCache("model-a")
	c.Set("fp-secret", FindingDisposition{
		Disposition: "False Positive",
		Rationale:   "Mitigated, but sample token sk-live-secret-value must not persist",
		Confidence:  "high",
		Evidence:    &DispositionEvidence{Basis: "code_mitigation", File: "config.go", Line: 1, Observed: "token := \"sk-live-secret-value\""},
	})
	c.Save()

	if _, ok := c.Get("fp-secret"); ok {
		t.Fatal("sensitive verdict evidence must not be cached")
	}
	if stats := c.Stats(); stats.SkippedSensitive != 1 {
		t.Fatalf("sensitive cache stats = %+v, want skipped sensitive", stats)
	}
	if _, err := os.Stat(c.path); err == nil {
		data, readErr := os.ReadFile(c.path)
		if readErr != nil {
			t.Fatalf("read cache file: %v", readErr)
		}
		if strings.Contains(string(data), "sk-live-secret-value") {
			t.Fatalf("cache file persisted sensitive evidence: %s", data)
		}
	}
}
