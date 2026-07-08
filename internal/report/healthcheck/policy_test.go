package healthcheck

import "testing"

func TestEvaluate_DefaultPolicyBlocksCriticalOrHigh(t *testing.T) {
	res := Evaluate(Input{Findings: []Finding{
		{Source: "core", Class: "AUTH", Severity: "HIGH", File: "a.go", Line: 1},
	}})
	if res.Verdict.Passed {
		t.Fatalf("default verdict passed; want failed for active HIGH")
	}
	if res.Policy.Profile != PolicyBaseline {
		t.Fatalf("profile = %q; want baseline", res.Policy.Profile)
	}
	if len(res.Verdict.BlockingReasons) == 0 {
		t.Fatalf("expected blocking reasons")
	}
	if res.Verdict.BlockingReasons[0].Code != "critical_or_high_findings" {
		t.Fatalf("reason code = %q; want critical_or_high_findings", res.Verdict.BlockingReasons[0].Code)
	}
}

func TestEvaluate_DefaultPolicyIgnoresFalsePositive(t *testing.T) {
	res := Evaluate(Input{Findings: []Finding{
		{Source: "core", Class: "AUTH", Severity: "CRITICAL", File: "a.go", Line: 1, Ignored: true},
	}})
	if !res.Verdict.Passed {
		t.Fatalf("verdict failed for ignored finding: %+v", res.Verdict.BlockingReasons)
	}
}

func TestApplyPolicy_MinimumScoreReason(t *testing.T) {
	res := Evaluate(Input{Findings: []Finding{
		{Source: "core", Class: "M1", Severity: "MEDIUM", File: "a.go", Line: 1},
	}})
	res = ApplyPolicy(res, Policy{
		Profile:      PolicyBaseline,
		FailOn:       FailOnNever,
		MinimumScore: 100,
		MaxCritical:  unlimitedThreshold,
		MaxHigh:      unlimitedThreshold,
		MaxMedium:    unlimitedThreshold,
	})
	if !res.Verdict.Passed {
		t.Fatalf("fail_on=never should disable gate, got reasons: %+v", res.Verdict.BlockingReasons)
	}

	res = ApplyPolicy(res, Policy{
		Profile:      PolicyBaseline,
		FailOn:       FailOnCritical,
		MinimumScore: 100,
		MaxCritical:  unlimitedThreshold,
		MaxHigh:      unlimitedThreshold,
		MaxMedium:    unlimitedThreshold,
	})
	if res.Verdict.Passed {
		t.Fatalf("minimum score policy passed; want failed")
	}
	if got := res.Verdict.BlockingReasons[0].Code; got != "minimum_score" {
		t.Fatalf("reason code = %q; want minimum_score", got)
	}
}

func TestApplyPolicy_BlockSourceAndClass(t *testing.T) {
	res := Evaluate(Input{Findings: []Finding{
		{Source: "gitleaks", Class: "SECRET", Severity: "CRITICAL", File: "a.go", Line: 1},
	}})
	res = ApplyPolicy(res, Policy{
		Profile:      PolicyBaseline,
		FailOn:       FailOnNever,
		MaxCritical:  unlimitedThreshold,
		MaxHigh:      unlimitedThreshold,
		MaxMedium:    unlimitedThreshold,
		BlockSources: []string{"gitleaks"},
		BlockClasses: []string{"secret"},
	})
	if !res.Verdict.Passed {
		t.Fatalf("fail_on=never should disable source/class gate")
	}

	res = ApplyPolicy(res, Policy{
		Profile:      PolicyBaseline,
		FailOn:       FailOnCritical,
		MaxCritical:  unlimitedThreshold,
		MaxHigh:      unlimitedThreshold,
		MaxMedium:    unlimitedThreshold,
		BlockSources: []string{"gitleaks"},
		BlockClasses: []string{"secret"},
	})
	foundSource := false
	foundClass := false
	for _, reason := range res.Verdict.BlockingReasons {
		if reason.Code == "blocked_source" && reason.Source == "gitleaks" {
			foundSource = true
		}
		if reason.Code == "blocked_class" && reason.Class == "secret" {
			foundClass = true
		}
	}
	if !foundSource || !foundClass {
		t.Fatalf("missing source/class reasons: %+v", res.Verdict.BlockingReasons)
	}
}

func TestPolicyForProfile(t *testing.T) {
	standard := PolicyForProfile(PolicyStandard)
	if standard.MinimumScore != 70 || standard.MaxHigh != 0 {
		t.Fatalf("unexpected standard policy: %+v", standard)
	}
	strict := PolicyForProfile(PolicyStrict)
	if strict.FailOn != FailOnAny || strict.MinimumScore != 90 {
		t.Fatalf("unexpected strict policy: %+v", strict)
	}
}
