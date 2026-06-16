package telemetry

import "time"

// ScanMetric records anonymized metrics from a single scan.
type ScanMetric struct {
	Timestamp     time.Time `json:"timestamp"`
	ProjectHash   string    `json:"project_hash"` // SHA256[:8] of project path for privacy
	Duration      int64     `json:"duration_ms"`  // scan duration in milliseconds
	FilesScanned  int       `json:"files_scanned"`
	FindingsTotal int       `json:"findings_total"`
	FindingsCrit  int       `json:"findings_crit"`
	SecurityScore int       `json:"security_score"`
	Stacks        []string  `json:"stacks"`
	ScannersUsed  []string  `json:"scanners_used"`
	TokensUsed    int       `json:"tokens_used"`
	LLMProvider   string    `json:"llm_provider,omitempty"`
}

// TelemetryStore holds aggregated telemetry data.
type TelemetryStore struct {
	TotalScans    int          `json:"total_scans"`
	TotalFiles    int          `json:"total_files"`
	TotalFindings int          `json:"total_findings"`
	AvgScore      float64      `json:"avg_score"`
	Scans         []ScanMetric `json:"scans"` // ring buffer, max 100
}
