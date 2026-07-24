package rules

import (
	"io/fs"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/dodobrands/aitriage/internal/models"
	"gopkg.in/yaml.v3"
)

// catalogPerDir returns the number of top-level rules per top-level category
// directory (e.g. "fastapi") and the grand total, counting every documented
// rule in the catalog (regex, AST, taint and entropy-analysis alike).
func catalogPerDir(t *testing.T) (perDir map[string]int, total int) {
	t.Helper()
	perDir = map[string]int{}
	err := fs.WalkDir(FS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || (!strings.HasSuffix(p, ".yaml") && !strings.HasSuffix(p, ".yml")) {
			return nil
		}
		data, err := fs.ReadFile(FS, p)
		if err != nil {
			return err
		}
		var rs models.Ruleset
		if err := yaml.Unmarshal(data, &rs); err != nil {
			return err
		}
		dir := path.Dir(p) // e.g. "fastapi", "universal"
		perDir[dir] += len(rs.Rules)
		total += len(rs.Rules)
		return nil
	})
	if err != nil {
		t.Fatalf("walk catalog: %v", err)
	}
	return perDir, total
}

var (
	reHeaderTotal = regexp.MustCompile(`\*\*(\d+) security rules\*\*`)
	reTableLink   = regexp.MustCompile(`\]\(\./([a-z0-9]+)/\)`)
	reCell        = regexp.MustCompile(`^\d+$`)
)

// TestREADMECountsMatchCatalog is the anti-drift guard: the per-category numbers
// and the header total in rules/README.md must equal the real embedded catalog
// counts. If a rule is added or removed and the README is not updated, this test
// fails — the numbers can never silently desynchronize again.
func TestREADMECountsMatchCatalog(t *testing.T) {
	perDir, total := catalogPerDir(t)

	data, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	readme := string(data)

	// 1. Header total.
	m := reHeaderTotal.FindStringSubmatch(readme)
	if m == nil {
		t.Fatal("could not find '**N security rules**' header in README.md")
	}
	headerTotal, _ := strconv.Atoi(m[1])
	if headerTotal != total {
		t.Errorf("README header says %d rules, catalog has %d", headerTotal, total)
	}

	// 2. Per-directory rows: each category row links to ./<dir>/ and carries a
	// count cell. Every directory in the catalog must have a matching, correct row.
	documented := map[string]bool{}
	for _, line := range strings.Split(readme, "\n") {
		lm := reTableLink.FindStringSubmatch(line)
		if lm == nil {
			continue
		}
		dir := lm[1]
		var count int
		found := false
		for _, cell := range strings.Split(line, "|") {
			c := strings.TrimSpace(cell)
			if reCell.MatchString(c) {
				count, _ = strconv.Atoi(c)
				found = true
				break
			}
		}
		if !found {
			t.Errorf("README row for %q has no numeric count cell: %s", dir, strings.TrimSpace(line))
			continue
		}
		documented[dir] = true
		if want, ok := perDir[dir]; !ok {
			t.Errorf("README documents unknown category %q", dir)
		} else if count != want {
			t.Errorf("README says %q has %d rules, catalog has %d", dir, count, want)
		}
	}

	// 3. Every catalog directory must be documented.
	for dir := range perDir {
		if !documented[dir] {
			t.Errorf("catalog directory %q is not documented in README.md", dir)
		}
	}
}
