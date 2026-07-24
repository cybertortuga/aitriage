package runstore

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dodobrands/aitriage/internal/agent/hostagent"
	"github.com/dodobrands/aitriage/internal/report/healthcheck"
	"github.com/dodobrands/aitriage/internal/scanner/external"
)

// SchemaVersion is bumped only for incompatible run-bundle changes. Open()
// refuses a bundle whose schema it does not understand rather than guessing.
const SchemaVersion = 1

// runsDirName is the single project-local directory that holds ALL AITriage
// scan/run artifacts — host-agent run bundles live directly under it as
// <root>/aitriage-reports/<run-id>/, and raw scan history under
// aitriage-reports/history/. It is git-ignored and excluded from scans.
const runsDirName = "aitriage-reports"

// Default fail-closed limits. They bound a single response, the whole bundle,
// and the number of deferred turns so a malicious or runaway host cannot fill
// the disk or loop forever.
const (
	defaultMaxResponseBytes = 8 << 20   // 8 MiB per submitted response
	defaultMaxBundleBytes   = 256 << 20 // 256 MiB per run bundle
	defaultMaxTurns         = 500
)

// runIDPattern allows only safe, self-generated identifiers. It blocks absolute
// paths, "..", and separator injection by construction.
var runIDPattern = regexp.MustCompile(`^run-[0-9]{8}T[0-9]{6}-[0-9a-f]{8}$`)

// fingerprintPattern is the sha256 hex shape used for request/response IDs. It
// blocks any path separator or traversal in a bundle filename by construction.
var fingerprintPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// Limits bounds resource usage for a run.
type Limits struct {
	MaxResponseBytes int64
	MaxBundleBytes   int64
	MaxTurns         int
}

func (l Limits) withDefaults() Limits {
	if l.MaxResponseBytes <= 0 {
		l.MaxResponseBytes = defaultMaxResponseBytes
	}
	if l.MaxBundleBytes <= 0 {
		l.MaxBundleBytes = defaultMaxBundleBytes
	}
	if l.MaxTurns <= 0 {
		l.MaxTurns = defaultMaxTurns
	}
	return l
}

// Versions records the code identity that produced a run, so a bundle is only
// replayed under compatible logic.
type Versions struct {
	AITriage       string `json:"aitriage"`
	ScannerOrch    string `json:"scanner_orchestration"`
	SecureCoder    string `json:"securecoder_prompt"`
	Graph          string `json:"graph"`
	TriageSchema   int    `json:"triage_schema"`
	RunStoreSchema int    `json:"run_store_schema"`
}

// GitInfo is optional VCS context. Absent fields are simply empty.
type GitInfo struct {
	Commit string `json:"commit,omitempty"`
	Dirty  bool   `json:"dirty,omitempty"`
	// TreeFingerprint is a safe digest of the working tree, never file contents.
	TreeFingerprint string `json:"tree_fingerprint,omitempty"`
}

// Manifest is the portable, secret-free description of a run. It stores only a
// hash of the canonical project root, never the absolute path, so bundles are
// not tied to one machine's layout while still being bound to one project.
type Manifest struct {
	SchemaVersion   int                `json:"schema_version"`
	RunID           string             `json:"run_id"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	Status          Status             `json:"status"`
	ProjectRootHash string             `json:"project_root_hash"`
	Git             GitInfo            `json:"git"`
	Versions        Versions           `json:"versions"`
	Policy          healthcheck.Policy `json:"policy"`
	HostClient      string             `json:"host_client,omitempty"`
	// Intent is the user's declared intent for this run: "audit" (default) or
	// "audit_and_fix". It is set once at creation and never auto-escalated.
	Intent string `json:"intent,omitempty"`
	// CacheDir is the effective verdict/artifact cache directory for this run
	// (under aitriage-reports/cache unless the environment overrides it). Recorded
	// for observability; never contains credentials.
	CacheDir string `json:"cache_dir,omitempty"`
	Turns    int    `json:"turns"`
	Limits   Limits `json:"limits"`
	// Kind is "triage" (default) or "verification" for a linked re-check run.
	Kind string `json:"kind,omitempty"`
	// ParentRunID links a verification run back to the triage run it verifies.
	ParentRunID string `json:"parent_run_id,omitempty"`
	// ScanHash is the sha256 of the persisted (redacted) scan snapshot, checked on
	// resume to detect tampering with the deterministic state input.
	ScanHash string `json:"scan_hash,omitempty"`
	// Scanners is the per-scanner execution manifest — proof of what actually ran.
	// A full audit records completed/missing/failed/not_applicable for every
	// mandatory scanner; it is never a silent skip.
	Scanners []external.ScannerExecution `json:"scanners,omitempty"`
	// ScannerCoverage is "full" only when every required external scanner
	// completed (or was not_applicable); otherwise "partial".
	ScannerCoverage string `json:"scanner_coverage,omitempty"`
	// LastError is a redacted, safe description of the last failure, if any.
	LastError string `json:"last_error,omitempty"`
}

// Store is rooted at a single project's <root>/aitriage-reports directory.
type Store struct {
	baseDir         string
	projectRoot     string
	projectRootHash string
	limits          Limits
}

// NewStore resolves the canonical project root and prepares the runs directory.
func NewStore(projectRoot string) (*Store, error) {
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, fmt.Errorf("project root %q inaccessible: %w", projectRoot, err)
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("project root %q is not a directory", projectRoot)
	}
	base := filepath.Join(resolved, runsDirName)
	if err := os.MkdirAll(base, 0o700); err != nil {
		return nil, fmt.Errorf("create runs dir: %w", err)
	}
	if err := os.Chmod(base, 0o700); err != nil {
		return nil, fmt.Errorf("secure runs dir: %w", err)
	}
	return &Store{
		baseDir:         base,
		projectRoot:     resolved,
		projectRootHash: hashPath(resolved),
		limits:          Limits{}.withDefaults(),
	}, nil
}

// ProjectRoot returns the canonical project root bound to this store.
func (s *Store) ProjectRoot() string { return s.projectRoot }

func hashPath(p string) string {
	sum := sha256.Sum256([]byte(p))
	return hex.EncodeToString(sum[:])
}

// newRunID generates a collision-resistant, path-safe identifier.
func newRunID(now time.Time) (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("run-%s-%s", now.UTC().Format("20060102T150405"), hex.EncodeToString(b[:])), nil
}

// validateRunID enforces the safe identifier shape and rejects traversal.
func validateRunID(id string) error {
	if !runIDPattern.MatchString(id) {
		return fmt.Errorf("invalid run id %q", id)
	}
	if strings.ContainsAny(id, `/\`) || strings.Contains(id, "..") {
		return fmt.Errorf("invalid run id %q", id)
	}
	return nil
}

// Create starts a new run bundle bound to this store's project.
func (s *Store) Create(hostClient string, versions Versions, policy healthcheck.Policy, git GitInfo, intent string) (*Run, error) {
	now := time.Now().UTC()
	id, err := newRunID(now)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(s.baseDir, id)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create run dir: %w", err)
	}
	for _, sub := range []string{"requests", "responses"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o700); err != nil {
			return nil, err
		}
	}
	versions.RunStoreSchema = SchemaVersion
	m := &Manifest{
		SchemaVersion:   SchemaVersion,
		RunID:           id,
		CreatedAt:       now,
		UpdatedAt:       now,
		Status:          StatusScanning,
		ProjectRootHash: s.projectRootHash,
		Git:             git,
		Versions:        versions,
		Policy:          policy,
		HostClient:      sanitizeClient(hostClient),
		Intent:          sanitizeIntent(intent),
		Limits:          s.limits,
	}
	r := &Run{store: s, id: id, dir: dir, manifest: m}
	if err := r.writeManifest(); err != nil {
		return nil, err
	}
	return r, nil
}

// Open loads an existing run and enforces project binding and schema.
func (s *Store) Open(runID string) (*Run, error) {
	if err := validateRunID(runID); err != nil {
		return nil, err
	}
	dir := filepath.Join(s.baseDir, runID)
	// Reject a run dir that is a symlink escaping the runs base.
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, fmt.Errorf("run %q not found", runID)
	}
	if !withinDir(s.baseDir, resolved) {
		return nil, fmt.Errorf("run %q escapes the runs directory", runID)
	}
	var m Manifest
	if err := readJSON(filepath.Join(dir, "manifest.json"), &m); err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	if m.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("run %q uses schema %d, this build supports %d", runID, m.SchemaVersion, SchemaVersion)
	}
	if m.ProjectRootHash != s.projectRootHash {
		return nil, fmt.Errorf("run %q belongs to a different project and cannot be opened here", runID)
	}
	if m.RunID != runID {
		return nil, fmt.Errorf("run %q manifest id mismatch", runID)
	}
	return &Run{store: s, id: runID, dir: dir, manifest: &m}, nil
}

// List returns the sorted run IDs known to this store's project.
func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() && validateRunID(e.Name()) == nil {
			ids = append(ids, e.Name())
		}
	}
	sort.Strings(ids)
	return ids, nil
}

// RemoveRun deletes a run bundle. Unfinished runs (not completed/failed) are
// refused unless force is true, so an interrupted triage is never silently
// discarded. The run ID is validated and the resolved directory must stay inside
// the runs base, so this can never delete outside the project's run store.
func (s *Store) RemoveRun(runID string, force bool) error {
	if err := validateRunID(runID); err != nil {
		return err
	}
	run, err := s.Open(runID)
	if err != nil {
		return err
	}
	if !force && !run.Status().isTerminal() {
		return fmt.Errorf("run %s is %q (not finished); use force to remove", runID, run.Status())
	}
	dir := filepath.Join(s.baseDir, runID)
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return err
	}
	if !withinDir(s.baseDir, resolved) || resolved == s.baseDir {
		return fmt.Errorf("refusing to remove %q: outside runs base", runID)
	}
	return os.RemoveAll(resolved)
}

// RunInfo is a compact summary for listing/cleanup.
type RunInfo struct {
	RunID    string
	Status   Status
	Terminal bool
}

// Runs returns compact info for every run bundle in the store.
func (s *Store) Runs() ([]RunInfo, error) {
	ids, err := s.List()
	if err != nil {
		return nil, err
	}
	var out []RunInfo
	for _, id := range ids {
		run, err := s.Open(id)
		if err != nil {
			continue // skip unreadable/foreign bundles
		}
		out = append(out, RunInfo{RunID: id, Status: run.Status(), Terminal: run.Status().isTerminal()})
	}
	return out, nil
}

func sanitizeClient(c string) string {
	c = strings.ToLower(strings.TrimSpace(c))
	switch c {
	case "codex", "claude-code":
		return c
	default:
		return ""
	}
}

// sanitizeIntent normalises the run intent; anything unrecognised falls back to
// the safe default "audit" so a run never silently escalates to fixing.
func sanitizeIntent(i string) string {
	switch strings.ToLower(strings.TrimSpace(i)) {
	case "audit_and_fix":
		return "audit_and_fix"
	default:
		return "audit"
	}
}

// Intent returns the recorded run intent ("audit" or "audit_and_fix").
func (r *Run) Intent() string {
	if r.manifest.Intent == "" {
		return "audit"
	}
	return r.manifest.Intent
}

// CacheDir returns the recorded effective cache directory, if any.
func (r *Run) CacheDir() string { return r.manifest.CacheDir }

// SetCacheDir records the effective cache directory in the manifest.
func (r *Run) SetCacheDir(dir string) error {
	return r.mutate(func() error {
		r.manifest.CacheDir = dir
		return r.writeManifest()
	})
}

// SetScanners records the scanner execution manifest and coverage.
func (r *Run) SetScanners(execs []external.ScannerExecution, coverage string) error {
	return r.mutate(func() error {
		r.manifest.Scanners = execs
		r.manifest.ScannerCoverage = coverage
		return r.writeManifest()
	})
}

// Run is a single run bundle. Its methods are safe for sequential use by one
// MCP handler at a time; concurrent submits to the same run are serialised by
// the process-wide runLock keyed on the bundle directory.
type Run struct {
	store    *Store
	id       string
	dir      string
	manifest *Manifest
	mu       sync.Mutex
}

// ID returns the run identifier.
func (r *Run) ID() string { return r.id }

// Dir returns the absolute bundle directory (in-process use only).
func (r *Run) Dir() string { return r.dir }

// Manifest returns a copy of the current manifest.
func (r *Run) Manifest() Manifest { return *r.manifest }

// Status returns the current run status.
func (r *Run) Status() Status { return r.manifest.Status }

// mutate runs fn under both the in-process mutex and the cross-process file
// lock, re-reading the manifest from disk first so a second Store.Open handle
// (any process) can never lose a concurrent read-modify-write. fn may mutate
// r.manifest and must call r.writeManifest() (or a file write) itself.
func (r *Run) mutate(fn func() error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	unlock, err := lockRun(r.dir)
	if err != nil {
		return err
	}
	defer unlock()
	if err := r.reloadManifest(); err != nil {
		return err
	}
	return fn()
}

// reloadManifest refreshes r.manifest from disk (called while the lock is held).
func (r *Run) reloadManifest() error {
	var m Manifest
	if err := readJSON(filepath.Join(r.dir, "manifest.json"), &m); err != nil {
		if os.IsNotExist(err) {
			return nil // fresh run, in-memory manifest is authoritative
		}
		return err
	}
	r.manifest = &m
	return nil
}

// SetStatus validates and persists a lifecycle transition.
func (r *Run) SetStatus(to Status) error {
	return r.mutate(func() error {
		if err := canTransition(r.manifest.Status, to); err != nil {
			return err
		}
		r.manifest.Status = to
		return r.writeManifest()
	})
}

// Fail records a redacted error and moves the run to the failed state.
func (r *Run) Fail(cause string) error {
	return r.mutate(func() error {
		r.manifest.LastError = truncate(Redact(cause), 2000)
		r.manifest.Status = StatusFailed
		return r.writeManifest()
	})
}

// SetLink marks this run as a verification of parentRunID.
func (r *Run) SetLink(kind, parentRunID string) error {
	return r.mutate(func() error {
		r.manifest.Kind = kind
		r.manifest.ParentRunID = parentRunID
		return r.writeManifest()
	})
}

// writeManifest persists the in-memory manifest. It does not lock; callers hold
// the run lock via mutate (or are in Create before the run is observable).
func (r *Run) writeManifest() error {
	r.manifest.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(r.manifest, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(filepath.Join(r.dir, "manifest.json"), data)
}

// ── hostagent.Store implementation ─────────────────────────────────────────

// SaveRequest persists req idempotently, keyed by its content fingerprint
// (RequestID). Concurrent saves from parallel graph.Run goroutines are safe:
// each request has its own file and the manifest turn counter is updated under
// the run lock. The request payload (messages) is stored VERBATIM — it is the
// exact prompt the host must answer, and the fingerprint is computed from these
// same bytes, so persisted representation and fingerprint stay consistent.
func (r *Run) SaveRequest(req hostagent.Request) error {
	if req.RequestID == "" {
		return fmt.Errorf("request has empty fingerprint")
	}
	if err := validateFingerprint(req.RequestID); err != nil {
		return err
	}
	return r.mutate(func() error {
		path := r.requestPath(req.RequestID)
		if _, ok, err := r.loadRequest(req.RequestID); err != nil {
			return err
		} else if ok {
			return nil // idempotent: identical fingerprint already recorded
		}
		if r.manifest.Turns >= r.manifest.Limits.MaxTurns {
			return fmt.Errorf("turn limit exceeded (%d)", r.manifest.Limits.MaxTurns)
		}
		r.manifest.Turns++
		req.Ordinal = r.manifest.Turns
		data, err := json.MarshalIndent(req, "", "  ")
		if err != nil {
			return err
		}
		if err := r.enforceBundleLimit(int64(len(data))); err != nil {
			return err
		}
		if err := atomicWrite(path, data); err != nil {
			return err
		}
		return r.writeManifest()
	})
}

// Response implements hostagent.Store: returns the stored response for a request
// fingerprint. Reads are lock-free — atomicWrite guarantees whole-file reads.
func (r *Run) Response(requestID string) (hostagent.Response, bool, error) {
	return r.loadResponse(requestID)
}

// SaveResponse validates a submitted response against the recorded request and
// stores it idempotently and VERBATIM (the content is replayed into graph.Run
// and must parse exactly as the model produced it). Missing request, fingerprint
// mismatch, or oversize are all fail-closed errors.
func (r *Run) SaveResponse(resp hostagent.Response) error {
	if err := validateFingerprint(resp.RequestID); err != nil {
		return err
	}
	return r.mutate(func() error {
		req, ok, err := r.loadRequest(resp.RequestID)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("no pending request for fingerprint %s", shortFP(resp.RequestID))
		}
		if resp.RequestID != req.RequestID {
			return fmt.Errorf("response fingerprint does not match request")
		}
		if int64(len(resp.Content)) > r.manifest.Limits.MaxResponseBytes {
			return fmt.Errorf("response exceeds size limit (%d bytes)", r.manifest.Limits.MaxResponseBytes)
		}
		if existing, ok, err := r.loadResponse(resp.RequestID); err != nil {
			return err
		} else if ok {
			if existing.Content != resp.Content {
				return fmt.Errorf("response for fingerprint %s already submitted with different content", shortFP(resp.RequestID))
			}
			return nil // idempotent duplicate submit
		}
		data, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return err
		}
		if err := r.enforceBundleLimit(int64(len(data))); err != nil {
			return err
		}
		return atomicWrite(r.responsePath(resp.RequestID), data)
	})
}

// PendingRequest returns a deterministic pending request: among all recorded
// requests that have no response yet, the one with the lexicographically
// smallest fingerprint. Determinism matters because parallel graph.Run passes
// can leave several requests pending at once; a stable choice keeps the workflow
// reproducible.
func (r *Run) PendingRequest() (hostagent.Request, bool, error) {
	pending, err := r.pendingRequests()
	if err != nil || len(pending) == 0 {
		return hostagent.Request{}, false, err
	}
	return pending[0], true, nil
}

// PendingCount returns the number of requests still awaiting a response.
func (r *Run) PendingCount() (int, error) {
	pending, err := r.pendingRequests()
	return len(pending), err
}

func (r *Run) pendingRequests() ([]hostagent.Request, error) {
	entries, err := os.ReadDir(filepath.Join(r.dir, "requests"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		ids = append(ids, strings.TrimSuffix(e.Name(), ".json"))
	}
	sort.Strings(ids)
	var pending []hostagent.Request
	for _, id := range ids {
		if _, hasResp, err := r.loadResponse(id); err != nil {
			return nil, err
		} else if hasResp {
			continue
		}
		req, ok, err := r.loadRequest(id)
		if err != nil {
			return nil, err
		}
		if ok {
			pending = append(pending, req)
		}
	}
	return pending, nil
}

func (r *Run) requestPath(requestID string) string {
	return filepath.Join(r.dir, "requests", requestID+".json")
}
func (r *Run) responsePath(requestID string) string {
	return filepath.Join(r.dir, "responses", requestID+".json")
}

func (r *Run) loadRequest(requestID string) (hostagent.Request, bool, error) {
	var req hostagent.Request
	err := readJSON(r.requestPath(requestID), &req)
	if os.IsNotExist(err) {
		return hostagent.Request{}, false, nil
	}
	if err != nil {
		return hostagent.Request{}, false, err
	}
	return req, true, nil
}

func (r *Run) loadResponse(requestID string) (hostagent.Response, bool, error) {
	var resp hostagent.Response
	err := readJSON(r.responsePath(requestID), &resp)
	if os.IsNotExist(err) {
		return hostagent.Response{}, false, nil
	}
	if err != nil {
		return hostagent.Response{}, false, err
	}
	return resp, true, nil
}

// validateFingerprint enforces the 64-hex-char shape so a request/response ID
// can never inject a path separator or traversal into a bundle filename.
func validateFingerprint(id string) error {
	if !fingerprintPattern.MatchString(id) {
		return fmt.Errorf("invalid request fingerprint")
	}
	return nil
}

func shortFP(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// AuditEvent is one structured, secret-free audit record. It deliberately omits
// prompt/response bodies — only identifiers, status, and a short redacted note.
type AuditEvent struct {
	Time      time.Time `json:"time"`
	RunID     string    `json:"run_id"`
	Event     string    `json:"event"`
	Status    Status    `json:"status"`
	RequestID string    `json:"request_id,omitempty"`
	Note      string    `json:"note,omitempty"`
}

// AppendAudit appends an audit event to the run's append-only audit.log (JSONL).
// It never records prompt or response content, only identifiers and status.
func (r *Run) AppendAudit(ev AuditEvent) error {
	ev.Time = time.Now().UTC()
	ev.RunID = r.id
	ev.Note = truncate(Redact(ev.Note), 500)
	line, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	return r.mutate(func() error {
		path := filepath.Join(r.dir, "audit.log")
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			return err
		}
		if _, err = f.Write(append(line, '\n')); err != nil {
			_ = f.Close()
			return err
		}
		return f.Close()
	})
}

// ── Generic bundle files ────────────────────────────────────────────────────

// SaveScan persists the deterministic scan snapshot used to rebuild state on
// resume. The snapshot is the authoritative triage INPUT: it must be stored
// VERBATIM (not redacted) so the deferred host-agent path triages exactly the
// same findings as the CI/CD direct path — parity would break if the model saw
// redacted evidence. It carries the same trust level as the verbatim request
// payloads (0600 bundle, git-ignored, excluded from scans). Its sha256 is
// recorded in the manifest so tampering is detected on resume.
func (r *Run) SaveScan(data []byte) error {
	sum := sha256.Sum256(data)
	return r.mutate(func() error {
		if err := r.enforceBundleLimit(int64(len(data))); err != nil {
			return err
		}
		if err := atomicWrite(filepath.Join(r.dir, "scan.json"), data); err != nil {
			return err
		}
		r.manifest.ScanHash = hex.EncodeToString(sum[:])
		return r.writeManifest()
	})
}

// LoadScan reads the scan snapshot and verifies its integrity hash against the
// manifest, failing closed on tampering. ok is false if none was written yet.
func (r *Run) LoadScan() ([]byte, bool, error) {
	data, err := os.ReadFile(filepath.Join(r.dir, "scan.json"))
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if r.manifest.ScanHash != "" {
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != r.manifest.ScanHash {
			return nil, false, fmt.Errorf("scan snapshot integrity check failed (tampered bundle)")
		}
	}
	return data, true, nil
}

// WriteArtifact writes a named bundle artifact (e.g. triage-findings.json).
// The name is validated to a small safe set to prevent path injection.
func (r *Run) WriteArtifact(name string, data []byte) error {
	if !allowedArtifact[name] {
		return fmt.Errorf("unknown artifact %q", name)
	}
	if err := r.enforceBundleLimit(int64(len(data))); err != nil {
		return err
	}
	return atomicWrite(filepath.Join(r.dir, name), data)
}

// ReadArtifact reads a named bundle artifact.
func (r *Run) ReadArtifact(name string) ([]byte, bool, error) {
	if !allowedArtifact[name] {
		return nil, false, fmt.Errorf("unknown artifact %q", name)
	}
	data, err := os.ReadFile(filepath.Join(r.dir, name))
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

// allowedArtifact is the exact set of CI-parity artifact filenames. The four AI
// artifacts use the same names as `aitriage agent` in CI/CD — no aliases.
var allowedArtifact = map[string]bool{
	"triage-findings.json": true,
	"report.md":            true,
	"summary.md":           true,
	"fixspec.md":           true,
	"aitriage.sarif":       true,
	"verification.json":    true,
	"approval.json":        true,
}

func (r *Run) enforceBundleLimit(add int64) error {
	size, err := dirSize(r.dir)
	if err != nil {
		return err
	}
	if size+add > r.manifest.Limits.MaxBundleBytes {
		return fmt.Errorf("bundle size limit exceeded (%d bytes)", r.manifest.Limits.MaxBundleBytes)
	}
	return nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
