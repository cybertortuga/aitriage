package runstore

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cybertortuga/aitriage/internal/agent/hostagent"
	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func testVersions() Versions {
	return Versions{AITriage: "test", SecureCoder: "securecoder-v1", TriageSchema: 1}
}

// fp returns a valid 64-hex request fingerprint derived from a seed.
func fp(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(sum[:])
}

func mkReq(id, content string) hostagent.Request {
	return hostagent.Request{
		RunID:     "x",
		RequestID: id,
		Messages:  []llm.Message{{Role: "user", Content: content}},
	}
}

func TestCreateOpenRoundTrip(t *testing.T) {
	s := newTestStore(t)
	run, err := s.Create("codex", testVersions(), healthcheck.DefaultPolicy(), GitInfo{Commit: "abc"}, "audit")
	if err != nil {
		t.Fatal(err)
	}
	if run.Status() != StatusScanning {
		t.Fatalf("new run status = %q", run.Status())
	}
	reopened, err := s.Open(run.ID())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if reopened.Manifest().HostClient != "codex" {
		t.Fatalf("host client not persisted: %+v", reopened.Manifest())
	}
	if reopened.Manifest().Git.Commit != "abc" {
		t.Fatal("git info not persisted in manifest")
	}
	if reopened.Manifest().Versions.RunStoreSchema != SchemaVersion {
		t.Fatal("run store schema not recorded")
	}
}

func TestManifestPermissions(t *testing.T) {
	s := newTestStore(t)
	run, _ := s.Create("", testVersions(), healthcheck.DefaultPolicy(), GitInfo{}, "audit")
	info, err := os.Stat(filepath.Join(run.Dir(), "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("manifest perm = %o, want 0600", perm)
	}
	dirInfo, _ := os.Stat(run.Dir())
	if perm := dirInfo.Mode().Perm(); perm != 0o700 {
		t.Fatalf("run dir perm = %o, want 0700", perm)
	}
}

func TestRunIDValidationRejectsTraversal(t *testing.T) {
	s := newTestStore(t)
	for _, bad := range []string{"../evil", "run-x/../y", "..", "/etc/passwd", "run-bad"} {
		if _, err := s.Open(bad); err == nil {
			t.Errorf("Open(%q) should fail", bad)
		}
	}
}

func TestOpenRejectsForeignProject(t *testing.T) {
	s := newTestStore(t)
	run, _ := s.Create("", testVersions(), healthcheck.DefaultPolicy(), GitInfo{}, "audit")
	data, _ := os.ReadFile(filepath.Join(run.Dir(), "manifest.json"))
	tampered := strings.Replace(string(data), s.projectRootHash, "deadbeef", 1)
	_ = os.WriteFile(filepath.Join(run.Dir(), "manifest.json"), []byte(tampered), 0o600)
	if _, err := s.Open(run.ID()); err == nil {
		t.Fatal("Open must reject a run bound to a different project")
	}
}

func TestCorruptedManifestRejected(t *testing.T) {
	s := newTestStore(t)
	run, _ := s.Create("", testVersions(), healthcheck.DefaultPolicy(), GitInfo{}, "audit")
	_ = os.WriteFile(filepath.Join(run.Dir(), "manifest.json"), []byte("{not json"), 0o600)
	if _, err := s.Open(run.ID()); err == nil {
		t.Fatal("Open must reject a corrupted manifest fail-closed")
	}
}

func TestRequestResponseFlowAndReplay(t *testing.T) {
	s := newTestStore(t)
	run, _ := s.Create("", testVersions(), healthcheck.DefaultPolicy(), GitInfo{}, "audit")

	id1 := fp("req-1")
	req := mkReq(id1, "analyze this code")
	if err := run.SaveRequest(req); err != nil {
		t.Fatal(err)
	}
	if err := run.SaveRequest(req); err != nil { // idempotent
		t.Fatalf("idempotent SaveRequest failed: %v", err)
	}

	// A non-hex fingerprint (traversal attempt) is rejected.
	if err := run.SaveRequest(mkReq("../escape", "x")); err == nil {
		t.Fatal("SaveRequest must reject a non-fingerprint id")
	}

	pending, ok, err := run.PendingRequest()
	if err != nil || !ok || pending.RequestID != id1 {
		t.Fatalf("PendingRequest = %+v ok=%v err=%v", pending, ok, err)
	}
	// Request payload is stored VERBATIM (the exact prompt the host must answer).
	if pending.Messages[0].Content != "analyze this code" {
		t.Fatalf("request payload not verbatim: %+v", pending.Messages)
	}

	if err := run.SaveResponse(hostagent.Response{RequestID: fp("nope"), Content: "y"}); err == nil {
		t.Fatal("SaveResponse must reject a response with no matching request")
	}
	resp := hostagent.Response{RequestID: id1, Content: "verdict", UsageReported: true, Usage: llm.Usage{TotalTokens: 5}}
	if err := run.SaveResponse(resp); err != nil {
		t.Fatal(err)
	}
	if err := run.SaveResponse(resp); err != nil { // idempotent duplicate
		t.Fatalf("idempotent SaveResponse failed: %v", err)
	}
	if err := run.SaveResponse(hostagent.Response{RequestID: id1, Content: "different"}); err == nil {
		t.Fatal("SaveResponse must reject conflicting duplicate")
	}

	got, ok, err := run.Response(id1)
	if err != nil || !ok || got.Content != "verdict" {
		t.Fatalf("Response = %+v ok=%v err=%v", got, ok, err)
	}
	if _, ok, _ := run.PendingRequest(); ok {
		t.Fatal("no request should be pending after response")
	}
}

// TestPayloadStoredVerbatim proves the blocker-2 contract: the request the host
// receives is byte-identical to what was recorded (not redacted), so it matches
// the fingerprint and is the exact SecureCoder prompt.
func TestPayloadStoredVerbatim(t *testing.T) {
	s := newTestStore(t)
	run, _ := s.Create("", testVersions(), healthcheck.DefaultPolicy(), GitInfo{}, "audit")
	secretish := `password = "supersecretvalue123"`
	_ = run.SaveRequest(mkReq(fp("a"), secretish))
	data, _ := os.ReadFile(run.requestPath(fp("a")))
	if !strings.Contains(string(data), "supersecretvalue123") {
		t.Fatal("request payload must be stored verbatim (host needs the exact prompt)")
	}
	// Response content is also verbatim (it is replayed into graph.Run).
	_ = run.SaveResponse(hostagent.Response{RequestID: fp("a"), Content: `{"verdict":"tp","note":"token = abc"}`})
	rdata, _ := os.ReadFile(run.responsePath(fp("a")))
	if !strings.Contains(string(rdata), "token = abc") {
		t.Fatal("response content must be stored verbatim for replay")
	}
}

// TestScanSnapshotVerbatimAndHashed proves the snapshot is stored VERBATIM (it
// is the triage input and must match CI/direct exactly) and integrity-checked
// on load so tampering fails closed.
func TestScanSnapshotVerbatimAndHashed(t *testing.T) {
	s := newTestStore(t)
	run, _ := s.Create("", testVersions(), healthcheck.DefaultPolicy(), GitInfo{}, "audit")
	snap := `{"evidence":"AKIAIOSFODNN7EXAMPLE in code"}`
	if err := run.SaveScan([]byte(snap)); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(run.Dir(), "scan.json"))
	if string(raw) != snap {
		t.Fatal("scan snapshot must be stored verbatim (triage-input parity)")
	}
	if run.Manifest().ScanHash == "" {
		t.Fatal("scan hash not recorded in manifest")
	}
	// Tamper the snapshot; LoadScan must fail closed.
	_ = os.WriteFile(filepath.Join(run.Dir(), "scan.json"), []byte(`{"evidence":"tampered"}`), 0o600)
	if _, _, err := run.LoadScan(); err == nil {
		t.Fatal("LoadScan must reject a tampered snapshot")
	}
}

func TestResponseSizeLimit(t *testing.T) {
	s := newTestStore(t)
	s.limits.MaxResponseBytes = 16 // persisted into the manifest by Create
	run, _ := s.Create("", testVersions(), healthcheck.DefaultPolicy(), GitInfo{}, "audit")
	_ = run.SaveRequest(mkReq(fp("a"), "q"))
	if err := run.SaveResponse(hostagent.Response{RequestID: fp("a"), Content: strings.Repeat("A", 100)}); err == nil {
		t.Fatal("oversize response must be rejected")
	}
}

func TestTurnLimit(t *testing.T) {
	s := newTestStore(t)
	s.limits.MaxTurns = 1 // persisted into the manifest by Create
	run, _ := s.Create("", testVersions(), healthcheck.DefaultPolicy(), GitInfo{}, "audit")
	if err := run.SaveRequest(mkReq(fp("a"), "x")); err != nil {
		t.Fatal(err)
	}
	if err := run.SaveRequest(mkReq(fp("b"), "y")); err == nil {
		t.Fatal("request beyond turn limit must be rejected")
	}
}

func TestStatusTransitions(t *testing.T) {
	s := newTestStore(t)
	run, _ := s.Create("", testVersions(), healthcheck.DefaultPolicy(), GitInfo{}, "audit")
	if err := run.SetStatus(StatusFixing); err == nil {
		t.Fatal("scanning -> fixing must be rejected")
	}
	for _, st := range []Status{StatusTriaging, StatusFinalized, StatusAwaitingUserApproval} {
		if err := run.SetStatus(st); err != nil {
			t.Fatalf("SetStatus(%q): %v", st, err)
		}
	}
	if err := run.SetStatus(StatusFixing); err != nil {
		t.Fatalf("awaiting_user_approval -> fixing should be legal: %v", err)
	}
}

func TestScanSnapshotResume(t *testing.T) {
	s := newTestStore(t)
	run, _ := s.Create("", testVersions(), healthcheck.DefaultPolicy(), GitInfo{}, "audit")
	if err := run.SaveScan([]byte(`{"findings":1}`)); err != nil {
		t.Fatal(err)
	}
	run2, err := s.Open(run.ID())
	if err != nil {
		t.Fatal(err)
	}
	data, ok, err := run2.LoadScan()
	if err != nil || !ok || string(data) != `{"findings":1}` {
		t.Fatalf("resume LoadScan = %q ok=%v err=%v", data, ok, err)
	}
}

func TestSymlinkEscapeRejected(t *testing.T) {
	s := newTestStore(t)
	outside := t.TempDir()
	link := filepath.Join(s.baseDir, "run-20200101T000000-deadbeef")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if _, err := s.Open("run-20200101T000000-deadbeef"); err == nil {
		t.Fatal("Open must reject a run dir symlink escaping the runs base")
	}
}

// TestConcurrentSubmitTwoHandles is the blocker-5 regression: two independent
// Store.Open handles for the same run submit distinct responses concurrently.
// The cross-process file lock + reload-under-lock must serialise the manifest
// read-modify-write so both responses persist and no update is lost.
func TestConcurrentSubmitTwoHandles(t *testing.T) {
	s := newTestStore(t)
	run, _ := s.Create("", testVersions(), healthcheck.DefaultPolicy(), GitInfo{}, "audit")
	idA, idB := fp("A"), fp("B")
	_ = run.SaveRequest(mkReq(idA, "qa"))
	_ = run.SaveRequest(mkReq(idB, "qb"))

	hA, _ := s.Open(run.ID())
	hB, _ := s.Open(run.ID())

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); _ = hA.SaveResponse(hostagent.Response{RequestID: idA, Content: "ra"}) }()
	go func() { defer wg.Done(); _ = hB.SaveResponse(hostagent.Response{RequestID: idB, Content: "rb"}) }()
	wg.Wait()

	fresh, _ := s.Open(run.ID())
	if _, ok, _ := fresh.Response(idA); !ok {
		t.Error("response A lost under concurrent submit")
	}
	if _, ok, _ := fresh.Response(idB); !ok {
		t.Error("response B lost under concurrent submit")
	}
}

func TestTreeFingerprintUsesContentNotMtime(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "app.go")
	fixedTime := time.Unix(1_700_000_000, 0)
	if err := os.WriteFile(path, []byte("AAAA"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, fixedTime, fixedTime); err != nil {
		t.Fatal(err)
	}
	before, err := TreeFingerprint(root)
	if err != nil {
		t.Fatal(err)
	}

	// Preserve both size and mtime while changing the bytes. A metadata-only
	// fingerprint would miss this pre-approval source edit.
	if err := os.WriteFile(path, []byte("BBBB"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, fixedTime, fixedTime); err != nil {
		t.Fatal(err)
	}
	after, err := TreeFingerprint(root)
	if err != nil {
		t.Fatal(err)
	}
	if before == after {
		t.Fatal("tree fingerprint did not detect same-size, same-mtime content change")
	}
}

func TestTreeFingerprintIsContentDeterministicAndSkipsReports(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	for _, root := range []string{left, right} {
		if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "src", "app.py"), []byte("print('ok')\n"), 0o640); err != nil {
			t.Fatal(err)
		}
	}
	// Different mtimes must not affect an identical source tree.
	if err := os.Chtimes(filepath.Join(right, "src", "app.py"), time.Unix(1_600_000_000, 0), time.Unix(1_600_000_000, 0)); err != nil {
		t.Fatal(err)
	}
	leftFP, _ := TreeFingerprint(left)
	rightFP, _ := TreeFingerprint(right)
	if leftFP != rightFP {
		t.Fatalf("identical content fingerprints differ: %s != %s", leftFP, rightFP)
	}

	if err := os.MkdirAll(filepath.Join(left, "aitriage-reports"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(left, "aitriage-reports", "run.json"), []byte("generated"), 0o600); err != nil {
		t.Fatal(err)
	}
	afterReports, _ := TreeFingerprint(left)
	if leftFP != afterReports {
		t.Fatal("generated aitriage-reports changed the source-tree fingerprint")
	}
}

func TestTreeFingerprintDoesNotFollowSymlinks(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	outsideFile := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("first"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsideFile, filepath.Join(root, "external-link")); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	before, err := TreeFingerprint(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outsideFile, []byte("second"), 0o600); err != nil {
		t.Fatal(err)
	}
	after, err := TreeFingerprint(root)
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("tree fingerprint followed a symlink outside the project")
	}
}
