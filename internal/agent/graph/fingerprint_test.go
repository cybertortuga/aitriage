package graph

import "testing"

func TestFingerprintDeterministicAndLocationSensitive(t *testing.T) {
	a := EnrichedFinding{ID: "rule1", Type: "core", File: "/src/app/x.go", Line: 10, Message: "boom"}
	b := EnrichedFinding{ID: "rule1", Type: "core", File: "app/x.go", Line: 10, Message: "boom"} // same after normalization
	c := EnrichedFinding{ID: "rule1", Type: "core", File: "app/y.go", Line: 10, Message: "boom"} // different location

	first := Fingerprint(a)
	if first != Fingerprint(a) {
		t.Fatal("fingerprint must be deterministic")
	}
	if Fingerprint(a) != Fingerprint(b) {
		t.Fatal("path normalization should make a and b identical")
	}
	if Fingerprint(a) == Fingerprint(c) {
		t.Fatal("same rule at different location must differ (location-sensitive)")
	}
}

func TestNormalizePath(t *testing.T) {
	cases := map[string]string{
		"/src/app/x.go": "app/x.go",
		"./app/x.go":    "app/x.go",
		"/app/x.go":     "app/x.go",
		"app\\x.go":     "app/x.go",
		"App/X.GO":      "app/x.go",
		"":              "",
	}
	for in, want := range cases {
		if got := normalizePath(in); got != want {
			t.Errorf("normalizePath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDedupFindingsNoDropAndGrouping(t *testing.T) {
	dup := EnrichedFinding{ID: "r", File: "a.go", Line: 1, Message: "m"}
	other := EnrichedFinding{ID: "r2", File: "b.go", Line: 2, Message: "m2"}
	findings := []EnrichedFinding{dup, other, dup, dup}

	unique, groups := dedupFindings(findings)
	if len(unique) != 2 {
		t.Fatalf("want 2 unique, got %d", len(unique))
	}
	// No finding dropped: group sizes sum to the original count.
	total := 0
	for _, g := range groups {
		total += len(g)
	}
	if total != len(findings) {
		t.Fatalf("group membership lost findings: sum=%d want=%d", total, len(findings))
	}
	// First occurrence order preserved.
	if unique[0].ID != "r" || unique[1].ID != "r2" {
		t.Fatalf("first-occurrence order not preserved: %+v", unique)
	}
	if len(groups[0]) != 3 {
		t.Fatalf("dup group should have 3 members, got %d", len(groups[0]))
	}
}

func TestProjectDispositionsMapsToAllMembers(t *testing.T) {
	findings := []EnrichedFinding{
		{ID: "r", VulnID: "CS-1", File: "a.go", Line: 1, Message: "m"},
		{ID: "r", VulnID: "CS-2", File: "a.go", Line: 1, Message: "m"}, // identical -> same group
	}
	unique, groups := dedupFindings(findings)
	if len(unique) != 1 {
		t.Fatalf("want 1 unique, got %d", len(unique))
	}
	uniqueDisps := []FindingDisposition{{Disposition: "True Positive", Rationale: "x", Confidence: "high", DispositionSource: dispositionSourceLLM}}

	out := projectDispositions(uniqueDisps, groups, findings)
	if len(out) != 2 {
		t.Fatalf("want 2 projected, got %d", len(out))
	}
	for i, d := range out {
		if d.FindingIndex != i {
			t.Errorf("projected[%d] index = %d", i, d.FindingIndex)
		}
		if d.FindingID != findings[i].VulnID {
			t.Errorf("projected[%d] vulnID = %q, want %q", i, d.FindingID, findings[i].VulnID)
		}
		if d.Disposition != "True Positive" {
			t.Errorf("projected[%d] disposition = %q", i, d.Disposition)
		}
	}
}
