package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/cybertortuga/aitriage/internal/scanner/external"
)

func TestWriteWebScanManifestIsPrivateAndComplete(t *testing.T) {
	reports := filepath.Join(t.TempDir(), "aitriage-reports")
	manifest := webScanManifest{
		SchemaVersion:   1,
		ScanID:          "scan-test",
		RequestedPath:   ".",
		ScannerCoverage: "full",
		Scanners: []external.ScannerExecution{
			{Scanner: "semgrep", Status: external.StatusCompleted, Findings: 2},
		},
		CreatedAt: time.Unix(1, 0).UTC(),
	}
	rel, err := writeWebScanManifest(reports, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if rel != "aitriage-reports/web-scans/scan-test/manifest.json" {
		t.Fatalf("manifest path = %q", rel)
	}
	path := filepath.Join(reports, "web-scans", "scan-test", "manifest.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("manifest mode = %o, want 600", info.Mode().Perm())
	}
	var round webScanManifest
	b, _ := os.ReadFile(path)
	if err := json.Unmarshal(b, &round); err != nil {
		t.Fatal(err)
	}
	if round.ScannerCoverage != "full" || len(round.Scanners) != 1 {
		t.Fatalf("manifest lost scanner evidence: %+v", round)
	}
}

func TestRunwayUsesCanonicalRootReportsDirectory(t *testing.T) {
	reports := filepath.Join(t.TempDir(), "aitriage-reports")
	t.Setenv("AITRIAGE_REPORTS_DIR", reports)
	if got := runwayReportsRoot("/workspace/nested/app"); got != reports {
		t.Fatalf("runway report root = %q, want canonical root %q", got, reports)
	}
}
