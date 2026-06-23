package graph

import "testing"

func TestVerdictCacheDisabledWithoutEnv(t *testing.T) {
	t.Setenv("AITRIAGE_CACHE_DIR", "")
	c := newVerdictCache("model-a")
	c.Set("fp1", FindingDisposition{Disposition: "True Positive"})
	if _, ok := c.Get("fp1"); ok {
		t.Fatal("disabled cache must not return entries")
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

	// A fresh cache with the same model must load the persisted verdict.
	c2 := newVerdictCache("model-a")
	if got, ok := c2.Get("fp1"); !ok || got.Disposition != "False Positive" {
		t.Fatalf("persisted verdict not reloaded: %+v ok=%v", got, ok)
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
