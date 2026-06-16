package graph

import (
	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/scanner/deployaudit"
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
	Diagram       string

	// Map-Reduce state
	EnrichedFindings []EnrichedFinding
	Batches          [][]EnrichedFinding
	TriagedResults   []string

	// Outputs
	ReportMarkdown string
	AIFixSpec      string
}

// EnrichedFinding is a unified representation of any finding with its source snippet attached.
type EnrichedFinding struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // "core", "external", "nfr", "deploy", "network"
	Severity  string `json:"severity"`
	File      string `json:"file,omitempty"`
	Line      int    `json:"line,omitempty"`
	Message   string `json:"message"`
	Snippet   string `json:"code_snippet,omitempty"`
	ExtraData string `json:"extra_data,omitempty"`
}
