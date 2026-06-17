package healthcheck

import "testing"

func TestEvaluate_CleanRepoIsPerfect(t *testing.T) {
	res := Evaluate(Input{
		Positives: []Positive{{ID: "auth"}, {ID: "cors"}},
	})
	if res.Score != 100 {
		t.Fatalf("clean repo score = %d; want 100", res.Score)
	}
	if res.Grade != "A+" {
		t.Fatalf("clean repo grade = %q; want A+", res.Grade)
	}
	if res.HasCriticalFailures {
		t.Fatalf("clean repo should not have critical failures")
	}
}

func TestEvaluate_FalsePositivesDoNotPenalise(t *testing.T) {
	withFP := Evaluate(Input{Findings: []Finding{
		{Source: "core", Class: "X1", Severity: "HIGH", File: "a.go", Line: 1, Ignored: true},
		{Source: "core", Class: "X2", Severity: "HIGH", File: "b.go", Line: 2, Ignored: true},
	}})
	if withFP.Score != 100 {
		t.Fatalf("all-FP score = %d; want 100", withFP.Score)
	}
	if withFP.Breakdown.IgnoredFindings != 2 {
		t.Fatalf("ignored count = %d; want 2", withFP.Breakdown.IgnoredFindings)
	}
}

func TestEvaluate_Deduplication(t *testing.T) {
	res := Evaluate(Input{Findings: []Finding{
		{Source: "core", Class: "DUP", Severity: "HIGH", File: "a.go", Line: 5},
		{Source: "core", Class: "DUP", Severity: "HIGH", File: "a.go", Line: 5},
	}})
	if res.Breakdown.ActiveFindings != 1 {
		t.Fatalf("active findings = %d; want 1 (deduped)", res.Breakdown.ActiveFindings)
	}
	if res.Breakdown.DedupedFindings != 1 {
		t.Fatalf("deduped findings = %d; want 1", res.Breakdown.DedupedFindings)
	}
}

func TestEvaluate_DeduplicationKeepsActiveFinding(t *testing.T) {
	res := Evaluate(Input{Findings: []Finding{
		{Source: "core", Class: "DUP", Severity: "HIGH", File: "a.go", Line: 5, Ignored: true},
		{Source: "core", Class: "DUP", Severity: "HIGH", File: "a.go", Line: 5},
	}})
	if res.Breakdown.ActiveFindings != 1 {
		t.Fatalf("active findings = %d; want 1 active duplicate to win", res.Breakdown.ActiveFindings)
	}
	if res.Breakdown.IgnoredFindings != 0 {
		t.Fatalf("ignored findings = %d; want 0 because the deduped representative is active", res.Breakdown.IgnoredFindings)
	}
	if !res.HasCriticalFailures {
		t.Fatalf("active HIGH duplicate should still fail critical gate")
	}
}

func TestEvaluate_DiminishingReturnsNeverInstantZero(t *testing.T) {
	var findings []Finding
	for i := 0; i < 7; i++ {
		findings = append(findings, Finding{
			Source:   "core",
			Class:    "H",
			Severity: "HIGH",
			File:     "f.go",
			Line:     i + 1,
		})
	}
	res := Evaluate(Input{Findings: findings})
	// Legacy scorer clamped this exact case to 0. The Health Check must degrade
	// smoothly and leave a non-zero, meaningful score.
	if res.Score <= 0 {
		t.Fatalf("7x HIGH score = %d; want > 0 (diminishing returns)", res.Score)
	}
	if res.Score >= 60 {
		t.Fatalf("7x HIGH score = %d; want a clearly failing score", res.Score)
	}
	if !res.HasCriticalFailures {
		t.Fatalf("HIGH findings should flag critical failures")
	}
}

func TestEvaluate_BonusSuppressedWhenCritical(t *testing.T) {
	positives := make([]Positive, 30)
	for i := range positives {
		positives[i] = Positive{ID: "good"}
	}
	res := Evaluate(Input{
		Positives: positives,
		Findings: []Finding{
			{Source: "core", Class: "C1", Severity: "CRITICAL", File: "a.go", Line: 1},
		},
	})
	// CI-safety: a real CRITICAL must not be masked by good-practice bonuses.
	if res.Breakdown.Bonus != 0 {
		t.Fatalf("bonus = %d; want 0 while a CRITICAL is active", res.Breakdown.Bonus)
	}
	if !res.HasCriticalFailures {
		t.Fatalf("expected HasCriticalFailures = true")
	}
}

func TestEvaluate_BonusAppliedWhenClean(t *testing.T) {
	positives := make([]Positive, 30)
	for i := range positives {
		positives[i] = Positive{ID: "good"}
	}
	res := Evaluate(Input{
		Positives: positives,
		Findings: []Finding{
			{Source: "core", Class: "M1", Severity: "MEDIUM", File: "a.go", Line: 1},
		},
	})
	if res.Breakdown.Bonus <= 0 {
		t.Fatalf("bonus = %d; want > 0 when only MEDIUM/LOW issues exist", res.Breakdown.Bonus)
	}
}

func TestEvaluate_SourceWeighting(t *testing.T) {
	core := Evaluate(Input{Findings: []Finding{
		{Source: "core", Class: "A", Severity: "HIGH", File: "a", Line: 1},
	}})
	nfr := Evaluate(Input{Findings: []Finding{
		{Source: "nfr", Class: "A", Severity: "HIGH"},
	}})
	if nfr.Score <= core.Score {
		t.Fatalf("nfr score (%d) should be higher than core score (%d) for same severity", nfr.Score, core.Score)
	}
}

func TestGrade(t *testing.T) {
	cases := map[int]string{100: "A+", 95: "A", 85: "B", 70: "C", 55: "D", 10: "F"}
	for score, want := range cases {
		if got := Grade(score); got != want {
			t.Errorf("Grade(%d) = %q; want %q", score, got, want)
		}
	}
}
