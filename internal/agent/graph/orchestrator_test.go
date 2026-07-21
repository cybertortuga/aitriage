package graph

import (
	"reflect"
	"testing"

	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/scanner/external"
	"github.com/cybertortuga/aitriage/internal/scanner/nfr"
)

func TestEnrichFindingsCanonicalizesOrderBeforeVulnIDs(t *testing.T) {
	findings := []core.CheckResult{
		{ID: "FAST-RATELIMIT", Name: "Missing Rate Limiting", Evidence: "Required pattern not found", Severity: "HIGH"},
		{ID: "ENTR-INSECURE-RANDOM", Name: "Insecure Randomness", Evidence: "Math.random", Severity: "HIGH", File: "frontend/src/Confetti.tsx", Line: 21},
		{ID: "FAST-CORS-WILDCARD", Name: "CORS Allow All Origins", Evidence: "allow_origins=['*']", Severity: "HIGH", File: "backend/main.py", Line: 24},
		{ID: "FAST-LOGGING", Name: "Missing Security Logging", Evidence: "Required pattern not found", Severity: "MEDIUM"},
	}

	stateA := &AgentState{CoreFindings: findings}
	stateB := &AgentState{CoreFindings: []core.CheckResult{findings[1], findings[3], findings[2], findings[0]}}
	enrichFindings(stateA)
	enrichFindings(stateB)

	gotA := enrichedOrderSignature(stateA.EnrichedFindings)
	gotB := enrichedOrderSignature(stateB.EnrichedFindings)
	if !reflect.DeepEqual(gotA, gotB) {
		t.Fatalf("canonical enriched order differs:\nA=%v\nB=%v", gotA, gotB)
	}
}

func TestEnrichFindingsMergesSemanticAliasesAndKeepsOrigins(t *testing.T) {
	state := &AgentState{
		CoreFindings: []core.CheckResult{
			{ID: "FAST-LAZY-EXC", Name: "Lazy Exception Handling", Evidence: "except pass", Severity: "HIGH", File: "app.py", Line: 10},
			{ID: "FAST-AUTH", Name: "Missing Authentication", Evidence: "required pattern", Severity: "HIGH"},
			{ID: "FAST-RATELIMIT", Name: "Missing Rate Limiting", Evidence: "required pattern", Severity: "HIGH"},
		},
		ExternalFindings: []external.UnifiedFinding{
			{RuleID: "B110", Source: "bandit", Severity: "LOW", File: "app.py", Line: 10, Message: "try_except_pass"},
		},
		NFRFindings: []nfr.NFRFinding{
			{RuleID: "NFR-API-003", Name: "Authentication Middleware Missing", Severity: "HIGH", Message: "No authentication middleware detected"},
			{RuleID: "NFR-API-001", Name: "Rate Limiting Missing", Severity: "HIGH", Message: "No rate limiting detected"},
		},
	}

	enrichFindings(state)
	if len(state.EnrichedFindings) != 3 {
		t.Fatalf("semantic findings = %d, want 3 instead of six aliases: %+v", len(state.EnrichedFindings), state.EnrichedFindings)
	}
	classes := map[string]EnrichedFinding{}
	for _, finding := range state.EnrichedFindings {
		classes[semanticFindingClass(finding)] = finding
	}
	for _, class := range []string{"exception_swallow", "missing_authentication", "missing_rate_limiting"} {
		finding, ok := classes[class]
		if !ok {
			t.Fatalf("missing semantic class %q: %+v", class, state.EnrichedFindings)
		}
		if len(finding.Origins) != 2 {
			t.Fatalf("%s origins = %+v, want both scanners", class, finding.Origins)
		}
	}

	computeHealthCheck(state)
	if state.HealthCheck.Breakdown.ActiveFindings != 3 {
		t.Fatalf("health active findings = %d, want semantic count 3", state.HealthCheck.Breakdown.ActiveFindings)
	}
}

func enrichedOrderSignature(findings []EnrichedFinding) []string {
	out := make([]string, 0, len(findings))
	for _, f := range findings {
		out = append(out, f.VulnID+"|"+f.ID+"|"+Fingerprint(f))
	}
	return out
}

func TestCountUncachedVerdicts(t *testing.T) {
	state := &AgentState{
		FindingDispositions: []FindingDisposition{
			{Fingerprint: "fp-a", DispositionSource: dispositionSourceNRFallback},
			{Fingerprint: "fp-a", DispositionSource: dispositionSourceNRFallback}, // duplicate projection
			{Fingerprint: "fp-b", DispositionSource: dispositionSourceNRFallback},
			{Fingerprint: "fp-c", DispositionSource: dispositionSourceLLM},
			{Fingerprint: "fp-d", DispositionSource: dispositionSourceCache},
		},
	}
	state.VerdictCacheStats.SkippedSensitive = 1

	if got := countUncachedVerdicts(state); got != 3 {
		t.Fatalf("countUncachedVerdicts = %d, want 2 unique NR-fallback + 1 sensitive-skipped", got)
	}
}

func TestClassifyMaxRetriesEnvOverride(t *testing.T) {
	t.Setenv("AITRIAGE_CLASSIFY_RETRIES", "")
	if got := classifyMaxRetries(); got != threatModelMaxRetries {
		t.Fatalf("default retries = %d, want %d", got, threatModelMaxRetries)
	}
	t.Setenv("AITRIAGE_CLASSIFY_RETRIES", "4")
	if got := classifyMaxRetries(); got != 4 {
		t.Fatalf("env retries = %d, want 4", got)
	}
	t.Setenv("AITRIAGE_CLASSIFY_RETRIES", "not-a-number")
	if got := classifyMaxRetries(); got != threatModelMaxRetries {
		t.Fatalf("invalid env retries = %d, want fallback %d", got, threatModelMaxRetries)
	}
}

func TestClassifyVulnCodeIsDeterministicForAmbiguousMessages(t *testing.T) {
	// Real remote-run messages that match several VulnClassCodes keys. Map
	// iteration used to pick a random winner, which changed CS-* IDs — and every
	// cache key derived from them — between otherwise identical runs.
	cases := []struct {
		message string
		want    string
	}{
		// "authentication" (14) beats "jwt" (3) and "secret" (6).
		{"[trivy] python-pyjwt: PyJWT: Authentication bypass due to forged JSON Web Tokens with a weak secret", "AUTH"},
		// "information disclosure" (22) beats "path traversal" (14).
		{"[trivy] vite: Vite: Information disclosure via path traversal in dev server's .map request handling", "INFO"},
		// "session" (7) beats "cookie" (6).
		{"Session cookie missing Secure flag", "SESSION"},
		// "hardcoded secret" (16) beats "secret" (6); same code, exercises specificity.
		{"Hardcoded secret in configuration file", "SECRETS"},
	}
	for _, tc := range cases {
		for i := 0; i < 500; i++ {
			if got := classifyVulnCode(tc.message); got != tc.want {
				t.Fatalf("classifyVulnCode(%q) = %q on call %d, want stable %q", tc.message, got, i, tc.want)
			}
		}
	}
}

func TestAssignVulnIDsStableAcrossRuns(t *testing.T) {
	build := func() []EnrichedFinding {
		findings := []EnrichedFinding{
			{ID: "T1", Type: "external", Source: "trivy", Severity: "HIGH", File: "req.txt", Line: 1, Message: "[trivy] PyJWT: Authentication bypass due to forged JSON Web Tokens with a weak secret"},
			{ID: "T2", Type: "external", Source: "trivy", Severity: "HIGH", File: "req.txt", Line: 2, Message: "[trivy] PyJWT: Denial of Service via crafted JWS tokens"},
			{ID: "T3", Type: "external", Source: "trivy", Severity: "MEDIUM", File: "pkg.json", Line: 3, Message: "[trivy] Vite: Information disclosure via path traversal in dev server"},
		}
		assignVulnIDs(findings)
		return findings
	}
	base := enrichedOrderSignature(build())
	for i := 0; i < 100; i++ {
		if got := enrichedOrderSignature(build()); !reflect.DeepEqual(base, got) {
			t.Fatalf("vuln IDs unstable across runs:\nbase=%v\ngot=%v", base, got)
		}
	}
}
