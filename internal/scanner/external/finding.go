package external

// UnifiedFinding — общая структура для результатов от всех сканеров
type UnifiedFinding struct {
	Source             string `json:"source"` // "aitriage" | "semgrep" | "gitleaks" | "trivy" | "bandit" | "securecoder"
	RuleID             string `json:"rule_id"`
	Severity           string `json:"severity"` // "CRITICAL" | "HIGH" | "MEDIUM" | "LOW" | "INFO"
	Message            string `json:"message"`
	File               string `json:"file"`
	Line               int    `json:"line"`
	Suggestion         string `json:"suggestion,omitempty"`
	OWASP              string `json:"owasp,omitempty"`
	CWE                string `json:"cwe,omitempty"`
	VulnerabilityClass string `json:"vulnerability_class,omitempty"`
	EndLine            int    `json:"end_line,omitempty"`
	StartColumn        int    `json:"start_column,omitempty"`
	EndColumn          int    `json:"end_column,omitempty"`
}
