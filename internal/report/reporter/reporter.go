package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/scanner"
	"github.com/fatih/color"
)

// PrintTerminal formats the ScanReport nicely for the terminal.
func PrintTerminal(report scanner.ScanReport) {
	fmt.Println()

	banner := `
   ___  _____________      _                
  / _ |/  _/_  __/ _ \____(_)__ ____ ____   
 / __ |/ /  / / / , _/ __/ / _ \/ _ \/ -_)  
/_/ |_/___/ /_/ /_/|_\__/_/\_,_/\_, /\__/   
                               /___/        
`
	if _, err := color.New(color.FgHiCyan, color.Bold).Println(banner); err != nil {
		fmt.Println(banner)
	}
	if _, err := color.New(color.BgBlack, color.FgHiBlue, color.Bold).Println("  AITRIAGE ENTERPRISE SAST ENGINE V1.0  "); err != nil {
		fmt.Println("  AITRIAGE ENTERPRISE SAST ENGINE V1.0  ")
	}
	fmt.Println("-------------------------------------------")
	fmt.Printf(" [STX] Stacks: %v\n", report.Stacks)

	scoreColor := color.FgGreen
	if report.SecurityScore < 80 {
		scoreColor = color.FgYellow
	}
	if report.SecurityScore < 50 {
		scoreColor = color.FgRed
	}

	barWidth := 20
	filled := (report.SecurityScore * barWidth) / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	fmt.Printf(" [SEC] Security Score: ")
	if _, err := color.New(scoreColor, color.Bold).Printf("%d/100 [%s]\n", report.SecurityScore, report.SecurityGrade); err != nil {
		fmt.Printf("%d/100 [%s]\n", report.SecurityScore, report.SecurityGrade)
	}
	if _, err := color.New(scoreColor).Printf("       [%s]\n", bar); err != nil {
		fmt.Printf("       [%s]\n", bar)
	}
	fmt.Println("----------------------------------------")

	for _, r := range report.Results {
		switch r.Status {
		case core.Present:
			if _, err := color.New(color.FgHiGreen).Printf(" ✔ [%s] %s\n", r.ID, r.Name); err != nil {
				fmt.Printf(" ✔ [%s] %s\n", r.ID, r.Name)
			}
		case core.Unknown:
			if _, err := color.New(color.FgHiYellow).Printf(" ? [%s] %s\n", r.ID, r.Name); err != nil {
				fmt.Printf(" ? [%s] %s\n", r.ID, r.Name)
			}
			fmt.Printf("   » Suggestion: %s\n", r.Suggestion)
		default:
			if _, err := color.New(color.FgHiRed).Printf(" ✘ [%s] %s\n", r.ID, r.Name); err != nil {
				fmt.Printf(" ✘ [%s] %s\n", r.ID, r.Name)
			}
			fmt.Printf("   » Issue: %s\n", r.Suggestion)
			if r.File != "" {
				if r.Line > 0 {
					fmt.Printf("   » Location: %s:%d\n", r.File, r.Line)
				} else {
					fmt.Printf("   » Location: %s\n", r.File)
				}
			} else if r.Line > 0 {
				fmt.Printf("   » At line: %d\n", r.Line)
			}
		}
	}

	fmt.Println("----------------------------------------")

	if report.HasCriticalFailures {
		if _, err := color.New(color.BgRed, color.FgWhite, color.Bold).Println("  STATUS: AUDIT FAILED (CRITICAL ISSUES)  "); err != nil {
			fmt.Println("  STATUS: AUDIT FAILED (CRITICAL ISSUES)  ")
		}
	} else {
		if _, err := color.New(color.BgGreen, color.FgBlack, color.Bold).Println("  STATUS: AUDIT PASSED (CLEAN)      "); err != nil {
			fmt.Println("  STATUS: AUDIT PASSED (CLEAN)      ")
		}
	}
	fmt.Println()
}

// PrintJSON outputs the ScanReport as JSON.
func PrintJSON(report scanner.ScanReport, w io.Writer) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
	}
}

// SarifReport represents the simplified SARIF structure
type SarifReport struct {
	Version string `json:"version"`
	Schema  string `json:"$schema"`
	Runs    []Run  `json:"runs"`
}

type Run struct {
	Tool    Tool     `json:"tool"`
	Results []Result `json:"results"`
}

type Tool struct {
	Driver Driver `json:"driver"`
}

type Driver struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Result struct {
	RuleID    string     `json:"ruleId"`
	Message   Message    `json:"message"`
	Level     string     `json:"level,omitempty"`
	Locations []Location `json:"locations,omitempty"`
}

type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           Region           `json:"region"`
}

type ArtifactLocation struct {
	URI string `json:"uri"`
}

type Region struct {
	StartLine int `json:"startLine"`
}

type Message struct {
	Text string `json:"text"`
}

// PrintSARIF outputs the ScanReport in SARIF format.
func PrintSARIF(report scanner.ScanReport, w io.Writer) {
	sarif := SarifReport{
		Version: "2.1.0",
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						Name:    "AITriage",
						Version: "1.5.0",
					},
				},
				Results: []Result{},
			},
		},
	}

	for _, r := range report.Results {
		if r.Status == core.Absent {
			level := "warning"
			switch strings.ToUpper(r.Severity) {
			case "CRITICAL", "HIGH":
				level = "error"
			case "MEDIUM":
				level = "warning"
			case "LOW":
				level = "note"
			}

			result := Result{
				RuleID: r.ID,
				Message: Message{
					Text: fmt.Sprintf("%s. %s", r.Name, r.Suggestion),
				},
				Level: level,
			}
			if r.File != "" {
				uri := r.File
				// Ensure relative path for SARIF
				result.Locations = []Location{
					{
						PhysicalLocation: PhysicalLocation{
							ArtifactLocation: ArtifactLocation{
								URI: uri,
							},
							Region: Region{
								StartLine: r.Line,
							},
						},
					},
				}
			}
			sarif.Runs[0].Results = append(sarif.Runs[0].Results, result)
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(sarif); err != nil {
		fmt.Printf("failed to encode SARIF JSON: %v\n", err)
	}
}
