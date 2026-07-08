package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/scanner"
)

// ScanRecord is a timestamped scan result stored on disk.
type ScanRecord struct {
	Timestamp   time.Time          `json:"timestamp"`
	ProjectPath string             `json:"project_path"`
	Report      scanner.ScanReport `json:"report"`
}

// DiffEntry describes a change between two scans.
type DiffEntry struct {
	RuleID   string `json:"rule_id"`
	Name     string `json:"name"`
	Severity string `json:"severity"`
	Change   string `json:"change"` // "added" | "fixed"
}

// historyDir returns (and creates if needed) the .aitriage/history directory inside the project.
func historyDir(projectPath string) (string, error) {
	dir := filepath.Join(projectPath, ".aitriage", "history")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("cannot create history dir: %w", err)
	}
	return dir, nil
}

// Save persists the scan report to disk as a JSON snapshot.
// Returns the path to the saved file.
func Save(projectPath string, report scanner.ScanReport) (string, error) {
	dir, err := historyDir(projectPath)
	if err != nil {
		return "", err
	}

	rec := ScanRecord{
		Timestamp:   time.Now().UTC(),
		ProjectPath: projectPath,
		Report:      report,
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return "", fmt.Errorf("cannot marshal scan record: %w", err)
	}

	filename := rec.Timestamp.Format("2006-01-02T15-04-05") + ".json"
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("cannot write scan history: %w", err)
	}
	return path, nil
}

// LoadLast returns the most recent saved ScanRecord for the project (excluding current run).
// Returns nil, nil if no history exists yet.
func LoadLast(projectPath string) (*ScanRecord, error) {
	dir, err := historyDir(projectPath)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil // no history
	}

	// Sort descending by name (timestamps are ISO, so lexicographic = chronological)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() > entries[j].Name()
	})

	// Skip the very first entry (most recent = just saved) and return second
	// Actually LoadLast is called BEFORE Save, so we just take the last file.
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var rec ScanRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			continue
		}
		return &rec, nil
	}
	return nil, nil
}

// Diff compares two scan reports and returns what was added (new findings) and fixed (resolved).
func Diff(prev, curr scanner.ScanReport) []DiffEntry {
	prevSet := make(map[string]core.CheckResult)
	for _, r := range prev.Results {
		prevSet[r.ID] = r
	}
	currSet := make(map[string]core.CheckResult)
	for _, r := range curr.Results {
		currSet[r.ID] = r
	}

	var diffs []DiffEntry

	// New findings (in curr but not prev)
	for id, r := range currSet {
		if _, existed := prevSet[id]; !existed {
			diffs = append(diffs, DiffEntry{
				RuleID:   id,
				Name:     r.Name,
				Severity: r.Severity,
				Change:   "added",
			})
		}
	}

	// Fixed findings (in prev but not curr)
	for id, r := range prevSet {
		if _, stillPresent := currSet[id]; !stillPresent {
			diffs = append(diffs, DiffEntry{
				RuleID:   id,
				Name:     r.Name,
				Severity: r.Severity,
				Change:   "fixed",
			})
		}
	}

	// Sort: fixed first (good news), then added (bad news)
	sort.Slice(diffs, func(i, j int) bool {
		if diffs[i].Change != diffs[j].Change {
			return diffs[i].Change == "fixed"
		}
		return diffs[i].RuleID < diffs[j].RuleID
	})

	return diffs
}

// FormatDiff renders a human-readable diff summary.
func FormatDiff(diffs []DiffEntry, prevScore, currScore int) string {
	if len(diffs) == 0 {
		return "✔ No changes since last scan."
	}

	var fixed, added []DiffEntry
	for _, d := range diffs {
		if d.Change == "fixed" {
			fixed = append(fixed, d)
		} else {
			added = append(added, d)
		}
	}

	out := fmt.Sprintf("Δ SecurityScore: %d → %d (%+d)\n\n", prevScore, currScore, currScore-prevScore)

	if len(fixed) > 0 {
		out += fmt.Sprintf("✔ Fixed (%d):\n", len(fixed))
		for _, d := range fixed {
			out += fmt.Sprintf("  ✔ [%s] %s (%s)\n", d.Severity, d.Name, d.RuleID)
		}
		out += "\n"
	}
	if len(added) > 0 {
		out += fmt.Sprintf("✘ New findings (%d):\n", len(added))
		for _, d := range added {
			out += fmt.Sprintf("  ✘ [%s] %s (%s)\n", d.Severity, d.Name, d.RuleID)
		}
	}
	return out
}
