package scanner

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cybertortuga/aitriage/internal/engine/core"
)

// SARIF Format structs
type SarifLog struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []SarifRun `json:"runs"`
}

type SarifRun struct {
	Tool    SarifTool     `json:"tool"`
	Results []SarifResult `json:"results"`
}

type SarifTool struct {
	Driver SarifDriver `json:"driver"`
}

type SarifDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version"`
	InformationUri string      `json:"informationUri"`
	Rules          []SarifRule `json:"rules"`
}

type SarifRule struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	ShortDescription SarifMessage     `json:"shortDescription"`
	Help             SarifMessage     `json:"help"`
	Properties       *SarifProperties `json:"properties,omitempty"`
}

type SarifProperties struct {
	SecuritySeverity string `json:"security-severity,omitempty"`
}

type SarifResult struct {
	RuleID    string          `json:"ruleId"`
	Message   SarifMessage    `json:"message"`
	Locations []SarifLocation `json:"locations"`
	Level     string          `json:"level"`
}

type SarifMessage struct {
	Text     string `json:"text"`
	Markdown string `json:"markdown,omitempty"`
}

type SarifLocation struct {
	PhysicalLocation SarifPhysicalLocation `json:"physicalLocation"`
}

type SarifPhysicalLocation struct {
	ArtifactLocation SarifArtifactLocation `json:"artifactLocation"`
	Region           SarifRegion           `json:"region,omitempty"`
}

type SarifArtifactLocation struct {
	Uri   string `json:"uri"`
	Index int    `json:"index,omitempty"`
}

type SarifRegion struct {
	StartLine   int `json:"startLine"`
	EndLine     int `json:"endLine,omitempty"`
	StartColumn int `json:"startColumn,omitempty"`
	EndColumn   int `json:"endColumn,omitempty"`
}

// ToSARIF converts the ScanReport to the SARIF format required by GitHub Advanced Security
func (r ScanReport) ToSARIF() ([]byte, error) {
	run := SarifRun{
		Tool: SarifTool{
			Driver: SarifDriver{
				Name:           "AITriage",
				Version:        "1.0.0", // Update with actual version later if needed
				InformationUri: "https://github.com/cybertortuga/aitriage",
				Rules:          []SarifRule{},
			},
		},
		Results: []SarifResult{},
	}

	seenRules := make(map[string]bool)

	for _, res := range r.Results {
		if res.Status != core.Absent && res.Status != core.Unknown {
			// In AITriage, a "Present" status for a vulnerability rule implies a finding.
			// However, sometimes it's missing security controls (Absent).
			// We only want to export actual findings to SARIF.
			if res.AuditStatus == core.AuditStatusIgnored || res.AuditStatus == core.AuditStatusTriage {
				continue // Skip ignored or triaged rules
			}

			// Add Rule to Driver if not seen
			if !seenRules[res.ID] {
				rule := SarifRule{
					ID:   res.ID,
					Name: res.Name,
					ShortDescription: SarifMessage{
						Text: res.Name,
					},
					Help: SarifMessage{
						Text:     fmt.Sprintf("%s\n\nSuggestion: %s", res.Evidence, res.Suggestion),
						Markdown: fmt.Sprintf("%s\n\n**Suggestion:** %s", res.Evidence, res.Suggestion),
					},
					Properties: &SarifProperties{
						SecuritySeverity: convertSeverity(res.Severity),
					},
				}
				run.Tool.Driver.Rules = append(run.Tool.Driver.Rules, rule)
				seenRules[res.ID] = true
			}

			// Format level for SARIF
			level := "warning"
			switch strings.ToUpper(res.Severity) {
			case "CRITICAL", "HIGH":
				level = "error"
			case "LOW":
				level = "note"
			}

			// Relative path mapping
			relPath := res.File
			if filepath.IsAbs(res.File) {
				if r, err := filepath.Rel(r.ProjectPath, res.File); err == nil {
					relPath = r
				}
			}

			// Convert backslashes for SARIF (URIs use forward slashes)
			relPath = filepath.ToSlash(relPath)

			// Determine line number
			line := res.Line
			if line <= 0 {
				line = 1 // SARIF requires line > 0
			}

			result := SarifResult{
				RuleID: res.ID,
				Message: SarifMessage{
					Text: fmt.Sprintf("[%s] %s: %s", res.Severity, res.Name, res.Suggestion),
				},
				Level: level,
				Locations: []SarifLocation{
					{
						PhysicalLocation: SarifPhysicalLocation{
							ArtifactLocation: SarifArtifactLocation{
								Uri: relPath,
							},
							Region: SarifRegion{
								StartLine: line,
							},
						},
					},
				},
			}
			run.Results = append(run.Results, result)
		}
	}

	log := SarifLog{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs:    []SarifRun{run},
	}

	return json.MarshalIndent(log, "", "  ")
}

// Map AITriage Severity to CVSS Score Scale (approximated for SARIF security-severity)
func convertSeverity(sev string) string {
	switch strings.ToUpper(sev) {
	case "CRITICAL":
		return "9.0"
	case "HIGH":
		return "7.0"
	case "MEDIUM":
		return "5.0"
	case "LOW":
		return "2.0"
	default:
		return "5.0"
	}
}
