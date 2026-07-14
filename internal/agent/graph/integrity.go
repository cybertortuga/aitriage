package graph

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// ── Artifact integrity validation ────────────────────────────────────────────
//
// The security gate is computed from state.FindingDispositions, never from the
// generated Markdown, so a drifting report cannot change the verdict. But the
// report/fixspec are consumed by humans and downstream AI agents, and the
// benchmark caught models inventing CS-*/CVE identifiers and contradicting the
// canonical dispositions. We validate the generated artifacts against the
// canonical state and, on failure, fall back to a deterministic-only artifact
// instead of publishing (or caching) misleading prose.

// validateArtifactIntegrity returns a list of human-readable violations found
// in the generated report/fixspec relative to canonical state. Empty means the
// artifacts are canonical-safe.
func validateArtifactIntegrity(state *AgentState) []string {
	validVulnIDs := canonicalDispositionByVulnID(state)
	validCVEs := canonicalCVEs(state)

	var violations []string
	seenBadVuln := map[string]bool{}
	seenBadCVE := map[string]bool{}

	for _, doc := range []struct{ name, body string }{
		{"report.md", state.ReportMarkdown},
		{"fixspec.md", state.AIFixSpec},
	} {
		for _, id := range vulnIDPattern.FindAllString(doc.body, -1) {
			if _, ok := validVulnIDs[id]; !ok && !seenBadVuln[doc.name+id] {
				seenBadVuln[doc.name+id] = true
				violations = append(violations, fmt.Sprintf("%s references unknown vulnerability ID %s (not in canonical findings)", doc.name, id))
			}
		}
		for _, cve := range cvePattern.FindAllString(doc.body, -1) {
			up := strings.ToUpper(cve)
			if !validCVEs[up] && !seenBadCVE[doc.name+up] {
				seenBadCVE[doc.name+up] = true
				violations = append(violations, fmt.Sprintf("%s references unknown %s (not in any canonical finding)", doc.name, up))
			}
		}
	}

	sort.Strings(violations)
	return violations
}

// enforceArtifactIntegrity validates the generated artifacts and, if they drift
// from canonical state, replaces them with deterministic-only safe versions.
// It returns true when a fallback was applied (the caller records the stat so
// it survives the final ArtifactCacheStats assignment).
func enforceArtifactIntegrity(state *AgentState) bool {
	violations := validateArtifactIntegrity(state)
	if len(violations) == 0 {
		return false
	}
	// Detailed violations go to the operator log; the published artifact carries
	// only a count so the rebuilt document itself stays canonical-safe (echoing
	// the offending identifiers would re-introduce non-canonical references).
	for _, v := range violations {
		fmt.Fprintf(os.Stderr, "      · integrity: %s\n", v)
	}
	state.ReportMarkdown = buildDeterministicReport(state, len(violations))
	state.AIFixSpec = buildDeterministicFixSpec(state, len(violations))
	return true
}

func integrityNotice(count int) string {
	return fmt.Sprintf("> **Integrity notice:** the AI-generated narrative was withheld because it did not match the canonical triage results (%d issue(s) detected; see run logs). This document was rebuilt deterministically from the triage engine.\n\n", count)
}

func buildDeterministicReport(state *AgentState, violationCount int) string {
	var sb strings.Builder
	sb.WriteString("# Security Report (deterministic)\n\n")
	sb.WriteString(integrityNotice(violationCount))
	sb.WriteString(buildCanonicalFindingsSection(state))
	return sb.String()
}

func buildDeterministicFixSpec(state *AgentState, violationCount int) string {
	var sb strings.Builder
	sb.WriteString("# AI Fix Specification (deterministic)\n\n")
	sb.WriteString(integrityNotice(violationCount))
	sb.WriteString(canonicalActiveFindingsBrief(state))
	return sb.String()
}
