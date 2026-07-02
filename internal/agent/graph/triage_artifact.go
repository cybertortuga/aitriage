package graph

import (
	"fmt"

	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
)

// TriageArtifactSchemaVersion is incremented only for incompatible changes to
// triage-findings.json. Consumers must reject versions they do not understand.
const TriageArtifactSchemaVersion = 1

// TriageFindingsArtifact is the canonical, machine-readable record of a
// completed AI triage. Reports are derived views; this artifact preserves every
// original finding, including findings classified as false positives.
type TriageFindingsArtifact struct {
	SchemaVersion int                `json:"schema_version"`
	TriageStatus  string             `json:"triage_status"`
	TotalFindings int                `json:"total_findings"`
	HealthCheck   healthcheck.Result `json:"health_check"`
	LLMUsage      LLMUsageArtifact   `json:"llm_usage"`
	VerdictCache  VerdictCacheStats  `json:"verdict_cache"`
	// ThreatModelSource is "llm", "cache_skipped", or "skipped_empty".
	ThreatModelSource string           `json:"threat_model_source,omitempty"`
	Findings          []TriagedFinding `json:"findings"`
	// ClassificationAudit records raw structured model responses and their
	// validation outcome. It is sensitive triage evidence and follows the same
	// artifact retention policy as the finding inventory.
	ClassificationAudit []ClassificationAuditEntry `json:"classification_audit"`
}

// LLMUsageArtifact records provider-reported token usage. Cost is deliberately
// not estimated here because pricing depends on provider, model, region,
// prompt-cache tier, and contract.
type LLMUsageArtifact struct {
	Total                llm.Usage            `json:"total"`
	Stages               map[string]llm.Usage `json:"stages,omitempty"`
	CacheTelemetryStatus string               `json:"cache_telemetry_status"`
}

// TriagedFinding pairs one original scanner finding with its one validated
// disposition. FindingIndex preserves original scanner order and makes the
// 1:1 mapping independently auditable.
type TriagedFinding struct {
	FindingIndex int                `json:"finding_index"`
	Finding      EnrichedFinding    `json:"finding"`
	Disposition  FindingDisposition `json:"disposition"`
}

// BuildTriageFindingsArtifact builds an export only from a complete, unambiguous
// triage result. It deliberately keeps duplicate findings: deduplication is an
// LLM optimization, while this artifact is the audit inventory of all inputs.
func BuildTriageFindingsArtifact(state *AgentState) (TriageFindingsArtifact, error) {
	if state == nil {
		return TriageFindingsArtifact{}, fmt.Errorf("build triage artifact: nil agent state")
	}
	if err := validateFindingDispositions(state.FindingDispositions, len(state.EnrichedFindings)); err != nil {
		return TriageFindingsArtifact{}, fmt.Errorf("build triage artifact: incomplete dispositions: %w", err)
	}

	findings := make([]TriagedFinding, len(state.EnrichedFindings))
	for _, disposition := range state.FindingDispositions {
		findings[disposition.FindingIndex] = TriagedFinding{
			FindingIndex: disposition.FindingIndex,
			Finding:      state.EnrichedFindings[disposition.FindingIndex],
			Disposition:  disposition,
		}
	}

	return TriageFindingsArtifact{
		SchemaVersion:       TriageArtifactSchemaVersion,
		TriageStatus:        "complete",
		TotalFindings:       len(findings),
		HealthCheck:         state.HealthCheck,
		LLMUsage:            buildLLMUsageArtifact(state),
		VerdictCache:        state.VerdictCacheStats,
		ThreatModelSource:   state.ThreatModelSource,
		Findings:            findings,
		ClassificationAudit: state.ClassificationAudit,
	}, nil
}

func buildLLMUsageArtifact(state *AgentState) LLMUsageArtifact {
	stages := make(map[string]llm.Usage, len(state.StageUsage))
	for name, usage := range state.StageUsage {
		if isZeroUsage(usage) {
			continue
		}
		stages[name] = usage
	}
	if len(stages) == 0 {
		stages = nil
	}

	status := "provider_did_not_report"
	if state.TotalUsage.CacheTelemetryReported {
		status = "reported"
	}

	return LLMUsageArtifact{
		Total:                state.TotalUsage,
		Stages:               stages,
		CacheTelemetryStatus: status,
	}
}
