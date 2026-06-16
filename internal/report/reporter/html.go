package reporter

import (
	"fmt"
	"html/template"
	"os"

	"github.com/cybertortuga/aitriage/internal/scanner"
)

const reportTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>AITriage Security Audit</title>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        body {
            font-family: 'Inter', sans-serif;
            background: #0a0a0a;
            color: #f1f5f9;
            padding: 60px 40px;
            font-size: 15px;
            line-height: 1.6;
            margin: 0;
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            border-bottom: 1px solid #1e2433;
            margin-bottom: 40px;
            padding-bottom: 20px;
        }
        h1 {
            font-weight: 700;
            font-size: 24px;
            letter-spacing: -0.02em;
            margin: 0 0 10px 0;
        }
        .meta {
            color: #64748b;
            font-size: 13px;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
        .summary {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
            margin-bottom: 50px;
        }
        .stat {
            background: #11141d;
            border: 1px solid #1e2433;
            border-radius: 8px;
            padding: 24px;
            display: flex;
            flex-direction: column;
            justify-content: center;
        }
        .stat-label {
            font-size: 12px;
            color: #64748b;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-bottom: 8px;
            font-weight: 600;
        }
        .grade { 
            font-size: 42px; 
            font-weight: 800; 
            line-height: 1;
        }
        .grade-a { color: #10b981; }
        .grade-b { color: #3b82f6; }
        .grade-c { color: #f59e0b; }
        .grade-d { color: #f97316; }
        .grade-f { color: #ef4444; }
        
        .table {
            width: 100%;
            border-collapse: collapse;
        }
        th { 
            text-align: left; 
            border-bottom: 1px solid #1e2433; 
            padding: 16px; 
            color: #64748b;
            font-size: 12px;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            font-weight: 600;
        }
        td { 
            padding: 16px; 
            border-bottom: 1px solid #1e2433; 
            vertical-align: top;
        }
        .status-ABSENT { color: #ef4444; font-weight: 600; }
        .status-PRESENT { color: #10b981; font-weight: 600; }
        .status-UNKNOWN { color: #64748b; }
        .entropy-tag { 
            background: #1e2433; 
            color: #f1f5f9; 
            padding: 2px 6px; 
            font-size: 10px; 
            border-radius: 4px;
            margin-left: 8px; 
            text-transform: uppercase;
            letter-spacing: 0.05em;
            font-weight: 600;
        }
        .text-subtle { color: #94a3b8; font-size: 14px; margin-top: 4px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>SECURITY AUDIT REPORT</h1>
        <div class="meta">STACKS: {{.Stacks}}</div>
    </div>

    <div class="summary">
        <div class="stat">
            <div class="stat-label">SECURITY GRADE</div>
            <div class="grade">{{.SecurityGrade}}</div>
        </div>
        <div class="stat">
            <div class="stat-label">SECURITY SCORE</div>
            <div class="grade">{{.SecurityScore}}/100</div>
        </div>
    </div>

    <table class="table">
        <thead>
            <tr>
                <th>Identifier</th>
                <th>Check Details</th>
                <th>Status & Resolution</th>
            </tr>
        </thead>
        <tbody>
            {{range .Results}}
            <tr>
                <td style="font-family: monospace; color: #64748b;">{{.ID}}</td>
                <td>
                    <strong>[{{.Framework}}] {{.Name}}</strong>
                    {{if eq (slice .ID 0 4) "ENTR"}}<span class="entropy-tag">ENTROPY_RISK</span>{{end}}
                </td>
                <td>
                    {{if eq .Status "PRESENT"}}
                    <div class="status-PRESENT">PASSED</div>
                    <div class="text-subtle">{{.Evidence}}</div>
                    {{else if eq .Status "UNKNOWN"}}
                    <div class="status-UNKNOWN">MANUAL REVIEW REQUIRED</div>
                    <div class="text-subtle">{{.Suggestion}}</div>
                    {{else}}
                    <div class="status-ABSENT">FAILED</div>
                    <div class="text-subtle">{{.Suggestion}}</div>
                    {{end}}
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
</body>
</html>
`

func PrintHTML(report scanner.ScanReport) {
	tmpl, err := template.New("report").Parse(reportTemplate)
	if err != nil {
		fmt.Printf("ERR_TEMPLATE_PARSE: %v\n", err)
		return
	}

	reportFile := "audit_report.html"
	f, err := os.Create(reportFile)
	if err != nil {
		fmt.Printf("ERR_FILE_CREATE: %v\n", err)
		return
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			fmt.Printf("warning: failed to close html report file: %v\n", closeErr)
		}
	}()

	if err := tmpl.Execute(f, report); err != nil {
		fmt.Printf("ERR_EXECUTE: %v\n", err)
		return
	}

	fmt.Printf("\nAudit report generated: %s\n", reportFile)
}
