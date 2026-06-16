package remedy

import (
	"fmt"
	"sort"
)

// Edit represents a single text replacement.
type Edit struct {
	StartByte int
	EndByte   int
	NewText   string
}

// ApplyEdits applies a list of non-overlapping edits to a source string.
// Edits must be sorted in reverse order of StartByte to maintain offset validity.
func ApplyEdits(original string, edits []Edit) (string, error) {
	// Sort edits: descending by StartByte
	sort.Slice(edits, func(i, j int) bool {
		return edits[i].StartByte > edits[j].StartByte
	})

	result := []byte(original)
	for _, e := range edits {
		if e.StartByte < 0 || e.EndByte > len(result) || e.StartByte > e.EndByte {
			return "", fmt.Errorf("invalid edit range: [%d, %d]", e.StartByte, e.EndByte)
		}

		// This is a naive implementation that doesn't handle overlapping very well yet.
		// For AAA Premium, we should detect overlaps.

		prefix := result[:e.StartByte]
		suffix := result[e.EndByte:]

		updated := append([]byte{}, prefix...)
		updated = append(updated, []byte(e.NewText)...)
		updated = append(updated, suffix...)
		result = updated
	}

	return string(result), nil
}

// GenerateFix generates a suggested fix for a given Rule ID.
// In a real AAA tool, this would be powered by LLM or specialized Codemods.
func GenerateFix(ruleID string, originalText string) (string, bool) {
	switch ruleID {
	case "ENTR-02": // Missing lockfile -> No simple code fix
		return "", false
	case "ENTR-04": // Over-apologetic comments
		return "// [Cleaned by aitriage]", true
	case "ENTR-03": // jwt.decode -> jwt.verify
		// This needs sophisticated AST parsing to know WHERE the secret is.
		// For now, we'll mark as 'manual review needed' or provide a template.
		return "jwt.verify(token, process.env.JWT_SECRET)", true
	default:
		return "", false
	}
}
