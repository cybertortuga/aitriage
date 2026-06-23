package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// classificationPromptFinding binds a model-visible batch position to a stable
// finding identity. A model cannot safely reuse an answer from another finding
// without echoing this identity back.
type classificationPromptFinding struct {
	FindingIndex int             `json:"finding_index"`
	FindingID    string          `json:"finding_id"`
	Fingerprint  string          `json:"fingerprint"`
	Finding      EnrichedFinding `json:"finding"`
}

type classificationAuditCollector struct {
	mu      sync.Mutex
	entries []ClassificationAuditEntry
}

func (c *classificationAuditCollector) record(entry ClassificationAuditEntry) {
	c.mu.Lock()
	c.entries = append(c.entries, entry)
	c.mu.Unlock()
}

func (c *classificationAuditCollector) snapshot() []ClassificationAuditEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]ClassificationAuditEntry, len(c.entries))
	copy(out, c.entries)
	return out
}

func newClassificationAuditEntry(attempt int, globalIndices []int, findings []EnrichedFinding, rawResponse string) ClassificationAuditEntry {
	entry := ClassificationAuditEntry{
		Attempt:              attempt,
		UniqueFindingIndices: append([]int(nil), globalIndices...),
		RawResponse:          rawResponse,
	}
	for _, finding := range findings {
		entry.FindingIDs = append(entry.FindingIDs, findingIdentity(finding))
		entry.Fingerprints = append(entry.Fingerprints, Fingerprint(finding))
	}
	return entry
}

func findingIdentity(f EnrichedFinding) string {
	if strings.TrimSpace(f.VulnID) != "" {
		return f.VulnID
	}
	return f.ID
}

func validateLLMDisposition(projectPath string, finding EnrichedFinding, raw rawDisposition) (FindingDisposition, string) {
	if raw.FindingID != findingIdentity(finding) {
		return FindingDisposition{}, "finding_id does not match requested finding"
	}
	if raw.Fingerprint != Fingerprint(finding) {
		return FindingDisposition{}, "fingerprint does not match requested finding"
	}
	if raw.Disposition == "False Positive" {
		if err := validateFalsePositiveEvidence(projectPath, finding, raw.Evidence); err != nil {
			return FindingDisposition{}, err.Error()
		}
	}
	return FindingDisposition{
		Disposition: raw.Disposition,
		Rationale:   raw.Rationale,
		Confidence:  normalizeConfidence(raw.Confidence),
		Evidence:    raw.Evidence,
	}, ""
}

func validateFalsePositiveEvidence(projectPath string, finding EnrichedFinding, evidence *DispositionEvidence) error {
	if evidence == nil {
		return fmt.Errorf("False Positive has no evidence")
	}
	switch evidence.Basis {
	case "test_only":
		if !isTestPath(finding.File) {
			return fmt.Errorf("test_only evidence does not reference a test finding")
		}
		if evidence.File == "" || normalizePath(evidence.File) != normalizePath(finding.File) {
			return fmt.Errorf("test_only evidence file does not match finding file")
		}
		if evidence.Line > 0 && finding.Line > 0 && evidence.Line != finding.Line {
			return fmt.Errorf("test_only evidence line does not match finding line")
		}
		return nil
	case "code_mitigation":
		if evidence.File == "" || evidence.Line < 1 || len(strings.TrimSpace(evidence.Observed)) < 4 {
			return fmt.Errorf("code_mitigation evidence is incomplete")
		}
		if finding.File == "" || normalizePath(evidence.File) != normalizePath(finding.File) {
			return fmt.Errorf("code_mitigation evidence file does not match finding file")
		}
		line, err := readProjectLine(projectPath, evidence.File, evidence.Line)
		if err != nil {
			return fmt.Errorf("code_mitigation evidence is unreadable: %w", err)
		}
		if !strings.Contains(line, evidence.Observed) {
			return fmt.Errorf("code_mitigation observed text is absent from source")
		}
		return nil
	default:
		return fmt.Errorf("False Positive evidence basis %q is not deterministically verifiable", evidence.Basis)
	}
}

func isTestPath(path string) bool {
	p := strings.ToLower("/" + normalizePath(path) + "/")
	base := strings.TrimSuffix(filepath.Base(p), "/")
	return strings.Contains(p, "/test/") || strings.Contains(p, "/tests/") ||
		strings.HasSuffix(base, "_test.go") || strings.HasSuffix(base, "_test.py") ||
		strings.Contains(base, ".test.") || strings.Contains(base, ".spec.")
}

func readProjectLine(projectPath, evidencePath string, lineNumber int) (string, error) {
	if filepath.IsAbs(evidencePath) {
		return "", fmt.Errorf("absolute evidence path is not allowed")
	}
	root, err := filepath.Abs(projectPath)
	if err != nil {
		return "", err
	}
	path := filepath.Join(root, filepath.Clean(evidencePath))
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("evidence path escapes project root")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	if lineNumber > len(lines) {
		return "", fmt.Errorf("line %d is outside source file", lineNumber)
	}
	return lines[lineNumber-1], nil
}
