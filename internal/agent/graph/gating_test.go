package graph

import "testing"

func TestGatingDisabledByDefault(t *testing.T) {
	t.Setenv("AITRIAGE_GATING", "")
	g := defaultGatingConfig()
	for _, sev := range []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "INFO", ""} {
		if !g.shouldTriageWithLLM(EnrichedFinding{Severity: sev}) {
			t.Fatalf("gating disabled: severity %q should still go to LLM", sev)
		}
	}
}

func TestGatingEnabledOnlyHighAndCritical(t *testing.T) {
	t.Setenv("AITRIAGE_GATING", "on")
	g := defaultGatingConfig()

	want := map[string]bool{"CRITICAL": true, "HIGH": true, "MEDIUM": false, "LOW": false, "INFO": false, "": false}
	for sev, expected := range want {
		if got := g.shouldTriageWithLLM(EnrichedFinding{Severity: sev}); got != expected {
			t.Errorf("shouldTriageWithLLM(%q) = %v, want %v", sev, got, expected)
		}
	}
}

func TestDeterministicDispositionNeverFalsePositive(t *testing.T) {
	d := deterministicDisposition(EnrichedFinding{Severity: "LOW"})
	if d.Disposition != "Needs Manual Review" {
		t.Fatalf("deterministic disposition must be NR, got %q", d.Disposition)
	}
	if d.DispositionSource != dispositionSourceDeterministic {
		t.Fatalf("source = %q, want deterministic", d.DispositionSource)
	}
}
