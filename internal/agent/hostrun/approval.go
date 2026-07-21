package hostrun

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cybertortuga/aitriage/internal/agent/pipeline"
	"github.com/cybertortuga/aitriage/internal/agent/runstore"
	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
)

// Approval is the recorded, auditable user decision that gates source changes.
// AITriage never mutates the project; the active coding agent may only touch
// files after this record exists and only for the selected canonical IDs.
type Approval struct {
	RunID             string    `json:"run_id"`
	SelectedIDs       []string  `json:"selected_ids"`
	Timestamp         time.Time `json:"timestamp"`
	TriageFingerprint string    `json:"triage_fingerprint"`
	BeforeTree        string    `json:"before_tree_fingerprint"`
}

// Verification links a before/after state around a fix and records the new gate.
type Verification struct {
	ParentRunID       string   `json:"parent_run_id"`
	VerificationRunID string   `json:"verification_run_id"`
	SelectedIDs       []string `json:"selected_ids"`
	BeforeTree        string   `json:"before_tree_fingerprint"`
	AfterTree         string   `json:"after_tree_fingerprint"`
	TreeChanged       bool     `json:"tree_changed"`
	// TPResults maps each approved True Positive ID to its post-fix state:
	// "resolved" (no longer an active TP) or "still_present". Tree changes are
	// never treated as proof of a fix — only the re-run disposition is.
	TPResults           map[string]string `json:"tp_results"`
	AllApprovedResolved bool              `json:"all_approved_resolved"`
	// Gate is the overall project gate after the fix — reported separately from
	// the per-TP result, since unrelated findings can still fail the gate.
	Gate      healthcheck.Verdict `json:"gate"`
	Health    healthcheck.Result  `json:"health"`
	Timestamp time.Time           `json:"timestamp"`
}

// Approve records the user's fix decision. It requires the run to be at the
// approval gate and every selected ID to be a known, fixable finding. It never
// changes source; it only authorises the agent to do so afterwards.
func (m *Manager) Approve(runID string, selectedIDs []string) (*Progress, error) {
	run, err := m.store.Open(runID)
	if err != nil {
		return nil, err
	}
	if run.Status() != runstore.StatusAwaitingUserApproval {
		return nil, fmt.Errorf("run %s is %q, not awaiting user approval", runID, run.Status())
	}

	triageBytes, ok, err := run.ReadArtifact("triage-findings.json")
	if err != nil || !ok {
		return nil, fmt.Errorf("run %s has no finalized triage artifact", runID)
	}
	allowed := fixableIDsFromTriage(triageBytes)
	for _, id := range selectedIDs {
		if !allowed[id] {
			return nil, fmt.Errorf("finding %q is not a known fixable finding for run %s", id, runID)
		}
	}

	projectPath, err := m.projectPathForRun(run)
	if err != nil {
		return nil, err
	}
	before, err := runstore.TreeFingerprint(projectPath)
	if err != nil {
		return nil, err
	}
	approval := Approval{
		RunID:             runID,
		SelectedIDs:       selectedIDs,
		Timestamp:         time.Now().UTC(),
		TriageFingerprint: sha256Hex(triageBytes),
		BeforeTree:        before,
	}
	data, err := json.MarshalIndent(approval, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := run.WriteArtifact("approval.json", data); err != nil {
		return nil, err
	}
	if err := run.SetStatus(runstore.StatusFixing); err != nil {
		return nil, err
	}
	_ = run.AppendAudit(runstore.AuditEvent{Event: "approved", Status: run.Status(), Note: fmt.Sprintf("%d selected", len(selectedIDs))})
	return &Progress{
		RunID:  runID,
		Status: run.Status(),
		Turns:  run.Manifest().Turns,
		Fix:    fixContext(runID, selectedIDs),
		Note:   "Approval recorded. Open summary.md and follow its AI Remediation Prompt / Operating Contract. Implement fixes ONLY for the approved True Positive IDs, run tests, then call aitriage_run_verify.",
	}, nil
}

// Decline finishes a run without any source change.
func (m *Manager) Decline(runID string) (*Progress, error) {
	run, err := m.store.Open(runID)
	if err != nil {
		return nil, err
	}
	if run.Status() != runstore.StatusAwaitingUserApproval {
		return nil, fmt.Errorf("run %s is %q, not awaiting user approval", runID, run.Status())
	}
	if err := run.SetStatus(runstore.StatusCompleted); err != nil {
		return nil, err
	}
	return m.statusOf(run)
}

// Verify starts a linked verification run on the current tree. It is only
// allowed once an approval record exists. The returned Progress belongs to the
// verification run; callers submit host answers against that run ID until it
// finalizes, at which point the parent's verification.json is written.
func (m *Manager) Verify(ctx context.Context, parentRunID string, opts StartOptions) (*Progress, error) {
	parent, err := m.store.Open(parentRunID)
	if err != nil {
		return nil, err
	}
	approvalBytes, ok, err := parent.ReadArtifact("approval.json")
	if err != nil || !ok {
		return nil, fmt.Errorf("run %s cannot be verified without an approval record", parentRunID)
	}
	var approval Approval
	if err := json.Unmarshal(approvalBytes, &approval); err != nil {
		return nil, err
	}
	if parent.Status() != runstore.StatusFixing && parent.Status() != runstore.StatusVerifying {
		return nil, fmt.Errorf("run %s is %q; verification requires the fixing state", parentRunID, parent.Status())
	}
	if err := parent.SetStatus(runstore.StatusVerifying); err != nil {
		return nil, err
	}
	data, ok, err := parent.LoadScan()
	if err != nil || !ok {
		return nil, fmt.Errorf("run %s has no scan snapshot", parentRunID)
	}
	var parentSnapshot snapshot
	if err := json.Unmarshal(data, &parentSnapshot); err != nil {
		return nil, fmt.Errorf("decode parent scan snapshot: %w", err)
	}
	// Verification must run against the exact same selected subproject and
	// runtime identity as the parent audit.
	opts.ProjectPath = parentSnapshot.ProjectPath
	opts.Scan = parentSnapshot.Scan
	opts.Target = parentSnapshot.Target
	opts.LLM = parentSnapshot.LLM
	opts.HostClient = parent.Manifest().HostClient
	opts.Intent = "audit"
	opts.Policy = parent.Manifest().Policy

	child, err := m.startLinked(ctx, opts, parentRunID)
	if err != nil {
		return nil, err
	}
	return child, nil
}

// fixableIDsFromTriage extracts the canonical IDs of active (non-FP) findings
// from a serialized triage artifact.
func fixableIDsFromTriage(triageBytes []byte) map[string]bool {
	out := map[string]bool{}
	var artifact struct {
		Findings []struct {
			Finding struct {
				VulnID string `json:"vuln_id"`
			} `json:"finding"`
			Disposition struct {
				FindingID   string `json:"finding_id"`
				Disposition string `json:"disposition"`
			} `json:"disposition"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(triageBytes, &artifact); err != nil {
		return out
	}
	for _, f := range artifact.Findings {
		// Only confirmed True Positives are fixable — never FP or Needs Manual Review.
		if f.Disposition.Disposition != "True Positive" {
			continue
		}
		id := f.Disposition.FindingID
		if id == "" {
			id = f.Finding.VulnID
		}
		if id != "" {
			out[id] = true
		}
	}
	return out
}

// finalizeVerification records the after-fix gate into the parent run and
// completes both the verification run and its parent.
func (m *Manager) finalizeVerification(run *runstore.Run, res *pipeline.Result) (*Progress, error) {
	parentID := run.Manifest().ParentRunID
	parent, err := m.store.Open(parentID)
	if err != nil {
		return nil, err
	}
	approvalBytes, ok, err := parent.ReadArtifact("approval.json")
	if err != nil || !ok {
		return nil, fmt.Errorf("verification parent %s lost its approval record", parentID)
	}
	var approval Approval
	if err := json.Unmarshal(approvalBytes, &approval); err != nil {
		return nil, err
	}

	projectPath, err := m.projectPathForRun(run)
	if err != nil {
		return nil, err
	}
	after, err := runstore.TreeFingerprint(projectPath)
	if err != nil {
		return nil, err
	}

	// Per-approved-TP result: an approved TP is "resolved" only if it is no longer
	// an active True Positive in the re-run's authoritative dispositions.
	afterTP := map[string]bool{}
	for _, id := range fixableIDs(res.State) {
		afterTP[id] = true
	}
	tpResults := map[string]string{}
	allResolved := true
	for _, id := range approval.SelectedIDs {
		if afterTP[id] {
			tpResults[id] = "still_present"
			allResolved = false
		} else {
			tpResults[id] = "resolved"
		}
	}

	verification := Verification{
		ParentRunID:         parentID,
		VerificationRunID:   run.ID(),
		SelectedIDs:         approval.SelectedIDs,
		BeforeTree:          approval.BeforeTree,
		AfterTree:           after,
		TreeChanged:         approval.BeforeTree != after,
		TPResults:           tpResults,
		AllApprovedResolved: allResolved,
		Gate:                res.Gate,
		Health:              res.Health,
		Timestamp:           time.Now().UTC(),
	}
	data, err := json.MarshalIndent(verification, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := parent.WriteArtifact("verification.json", data); err != nil {
		return nil, err
	}
	if err := parent.SetStatus(runstore.StatusCompleted); err != nil {
		return nil, err
	}
	if err := run.SetStatus(runstore.StatusCompleted); err != nil {
		return nil, err
	}

	resolvedCount := 0
	for _, s := range tpResults {
		if s == "resolved" {
			resolvedCount++
		}
	}
	return &Progress{
		RunID:  run.ID(),
		Status: run.Status(),
		Turns:  run.Manifest().Turns,
		Result: &FinalResult{
			Gate:                res.Gate,
			Health:              res.Health,
			TriageStatus:        "verified",
			VerifiedTP:          tpResults,
			AllApprovedResolved: allResolved,
			ArtifactPaths: map[string]string{
				"verification": artifactRelPath(parentID, "verification.json"),
			},
		},
		Note: fmt.Sprintf("Verification complete for parent run %s: %d/%d approved TP resolved; overall gate passed=%v. Remaining/unrelated findings affect the overall gate, not the per-TP result.",
			parentID, resolvedCount, len(approval.SelectedIDs), res.Gate.Passed),
	}, nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
