package graph

import (
	"encoding/json"

	agentcontext "github.com/cybertortuga/aitriage/internal/agent/context"
	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
	"github.com/cybertortuga/aitriage/internal/scanner/deployaudit"
	"github.com/cybertortuga/aitriage/internal/scanner/entropy"
	"github.com/cybertortuga/aitriage/internal/scanner/external"
	"github.com/cybertortuga/aitriage/internal/scanner/network"
	"github.com/cybertortuga/aitriage/internal/scanner/nfr"
)

// AgentState represents the shared state flowing through the orchestrator.
type AgentState struct {
	ProjectPath string
	DeepScan    bool

	// Deterministic Go findings
	CoreFindings     []core.CheckResult
	ExternalFindings []external.UnifiedFinding
	NFRFindings      []nfr.NFRFinding
	DeployFindings   []deployaudit.DeployFinding
	NetworkFindings  []network.NetworkFinding

	SecurityScore int
	SecurityGrade string

	// HealthCheck holds the unified, multi-source Security Health Check result.
	// It is the authoritative security posture score, recomputed after AI
	// triage so that False Positives no longer penalise the repository.
	HealthCheck healthcheck.Result
	Policy      healthcheck.Policy

	// Repository context (gathered by gatherRepoContext)
	RepoContext *agentcontext.RepoContext

	// Data from scanners that was previously lost
	CriticalFiles []entropy.CriticalFile
	HistoryLeaks  []entropy.HistoryLeak
	Diagram       string

	// Map-Reduce state
	EnrichedFindings []EnrichedFinding
	Batches          [][]EnrichedFinding // Deprecated: runWorkers removed (June 2026)
	TriagedResults   []string            // Deprecated: runWorkers removed (June 2026)

	// SecureCoder-enhanced fields
	ThreatModel         *ThreatModel         // Structured threat model analysis
	FindingDispositions []FindingDisposition // TP/FP/NR classification per finding
	PoCResults          []PoCResult          // PoC verification results

	// LLM usage tracking (accumulated across all Chat calls)
	TotalUsage llm.Usage

	// Outputs
	ReportMarkdown  string // Full report (includes FP rationale) → artifact
	SummaryMarkdown string // Actionable summary (TP + NR only) → GHA Step Summary
	AIFixSpec       string
}

// EnrichedFinding is a unified representation of any finding with its source snippet attached.
type EnrichedFinding struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // "core", "external", "nfr", "deploy", "network"
	Source    string `json:"source,omitempty"`
	Severity  string `json:"severity"`
	File      string `json:"file,omitempty"`
	Line      int    `json:"line,omitempty"`
	Message   string `json:"message"`
	Snippet   string `json:"code_snippet,omitempty"`
	ExtraData string `json:"extra_data,omitempty"`
	VulnID    string `json:"vuln_id,omitempty"` // CS-XXX-NNN format
}

// ── SecureCoder Threat Model Types ───────────────────────────────────────────

// ThreatModel holds the structured output from the LLM threat model step.
type ThreatModel struct {
	ComponentOverview  string       `json:"component_overview"`
	EntryPoints        []EntryPoint `json:"entry_points"`
	TrustBoundaries    TrustBounds  `json:"trust_boundaries"`
	SensitiveDataPaths []DataPath   `json:"sensitive_data_paths"`
	PrivilegedActions  []PrivAction `json:"privileged_actions"`
	PriorityAreas      []string     `json:"priority_areas"`
}

// EntryPoint describes a point where external data enters the system.
type EntryPoint struct {
	Endpoint   string `json:"endpoint"`
	Type       string `json:"type"`
	Trusted    bool   `json:"trusted"`
	Validation string `json:"validation"`
}

// TrustBounds describes authentication/authorization assumptions.
type TrustBounds struct {
	Authentication string `json:"authentication"`
	Authorization  string `json:"authorization"`
	ImplicitTrust  string `json:"implicit_trust"`
}

// DataPath describes a flow of sensitive data through the system.
type DataPath struct {
	DataType    string `json:"data_type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Protection  string `json:"protection"`
}

// PrivAction describes a privileged operation.
type PrivAction struct {
	Action   string `json:"action"`
	Location string `json:"location"`
	Guard    string `json:"guard"`
}

// FindingDisposition records the TP/FP/NR classification for a single finding.
type FindingDisposition struct {
	FindingIndex int    `json:"finding_index"`
	FindingID    string `json:"finding_id"`
	Disposition  string `json:"disposition"` // "True Positive", "False Positive", "Needs Manual Review"
	Rationale    string `json:"rationale"`
	// Confidence is the model's self-reported certainty: "high" | "medium" | "low".
	Confidence string `json:"confidence,omitempty"`
	// DispositionSource records how the disposition was produced for the audit
	// trail: "llm" | "cache" | "deterministic" | "nr-fallback".
	DispositionSource string `json:"disposition_source,omitempty"`
	// Fingerprint is the stable content hash used for dedup/caching.
	Fingerprint string `json:"fingerprint,omitempty"`
}

// PoCResult holds reasoning-based PoC verification for a finding.
type PoCResult struct {
	VulnerabilityType string    `json:"vulnerability_type"`
	Severity          string    `json:"severity"`
	AffectedFile      string    `json:"affected_file"`
	ReasoningSteps    []PoCStep `json:"reasoning_steps"`
	Conclusion        string    `json:"conclusion"`      // "Fix verified", "Fix incomplete", "Needs Manual Review"
	ExploitBlocked    *bool     `json:"exploit_blocked"` // nil = unknown
}

// PoCStep is one reasoning step in the PoC verification chain.
type PoCStep struct {
	// JSON numbers preserve labels such as 2.1 without weakening validation of
	// the surrounding PoC result.
	Step        json.Number `json:"step"`
	Description string      `json:"description"`
	Result      string      `json:"result"`
}
