package models

import "time"

const (
	RunwayArtifactSummaryMarkdown           = "summary_markdown"
	RunwayArtifactRemediationPromptMarkdown = "remediation_prompt_markdown"
	RunwayArtifactAgentDataJSON             = "agent_data_json"
	RunwayArtifactReportMarkdown            = "report_markdown"
	RunwayArtifactFixSpecMarkdown           = "fixspec_markdown"
	RunwayArtifactTriageFindingsJSON        = "triage_findings_json"
)

var validRunwayArtifactKinds = map[string]struct{}{
	RunwayArtifactSummaryMarkdown:           {},
	RunwayArtifactRemediationPromptMarkdown: {},
	RunwayArtifactAgentDataJSON:             {},
	RunwayArtifactReportMarkdown:            {},
	RunwayArtifactFixSpecMarkdown:           {},
	RunwayArtifactTriageFindingsJSON:        {},
}

func IsValidRunwayArtifactKind(kind string) bool {
	_, ok := validRunwayArtifactKinds[kind]
	return ok
}

type RunwayArtifact struct {
	ID            int64     `json:"id"`
	SessionID     int64     `json:"session_id"`
	Kind          string    `json:"kind"`
	MediaType     string    `json:"media_type"`
	SchemaVersion int       `json:"schema_version"`
	Content       string    `json:"-"`
	SHA256        string    `json:"sha256"`
	SizeBytes     int       `json:"size_bytes"`
	CreatedAt     time.Time `json:"created_at"`
}
