package reporter

import (
	"fmt"
	"html/template"
	"os"

	"github.com/cybertortuga/aitriage/internal/engine/core"
)

const dashTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>AITriage Security Report</title>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
<style>
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
body{font-family:'Inter',sans-serif;background:#0a0a0a;color:#e2e8f0;font-size:14px;line-height:1.6}
.mono{font-family:'JetBrains Mono',monospace}
.wrap{max-width:1060px;margin:0 auto;padding:40px 24px}

/* Header */
.rh{display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:32px;padding-bottom:24px;border-bottom:1px solid #1e2433}
.rh-title{font-size:20px;font-weight:700;color:#fff}
.rh-meta{font-size:12px;color:#64748b;margin-top:5px;letter-spacing:0.05em;text-transform:uppercase}
.rh-meta span{margin-right:16px}
.grade{font-size:52px;font-weight:800;line-height:1}
.gA{color:#10b981}.gB{color:#3b82f6}.gC{color:#f59e0b}.gD{color:#f97316}.gF{color:#ef4444}

/* Summary */
.stats{display:grid;grid-template-columns:repeat(4,1fr);gap:12px;margin-bottom:28px}
@media(max-width:640px){.stats{grid-template-columns:repeat(2,1fr)}}
.sc{background:#11141d;border:1px solid #1e2433;border-radius:8px;padding:20px 24px;transition:border-color 0.2s ease}
.sc:hover{border-color:#3b82f640}
.sc-label{font-size:11px;font-weight:600;text-transform:uppercase;letter-spacing:.08em;color:#64748b;margin-bottom:8px}
.sc-val{font-size:28px;font-weight:700;color:#fff}
.sc-val.red{color:#ef4444}

/* Filters */
.filters{display:flex;gap:8px;margin-bottom:24px;flex-wrap:wrap}
.fb{padding:6px 16px;border-radius:6px;border:1px solid #1e2433;background:#11141d;color:#94a3b8;font-size:12px;font-weight:600;cursor:pointer;transition:all .2s ease}
.fb:hover,.fb.on{border-color:#3b82f6;color:#3b82f6;background:#3b82f610}

/* Finding card */
.fc{background:#11141d;border:1px solid #1e2433;border-left:3px solid #475569;border-radius:8px;padding:20px 24px;margin-bottom:12px;transition:transform 0.2s ease, box-shadow 0.2s ease}
.fc:hover{transform:translateY(-1px);box-shadow:0 4px 12px rgba(0,0,0,0.2);border-color:#3b82f630}
.fc.CRITICAL{border-left-color:#ef4444}
.fc.HIGH{border-left-color:#f97316}
.fc.MEDIUM{border-left-color:#f59e0b}
.fc-top{display:flex;align-items:flex-start;gap:12px}
.pill{font-size:10px;font-weight:700;text-transform:uppercase;letter-spacing:.08em;padding:4px 8px;border-radius:4px;white-space:nowrap;flex-shrink:0;margin-top:2px}
.pill.CRITICAL{background:#ef444415;color:#ef4444;border:1px solid #ef444430}
.pill.HIGH{background:#f9731615;color:#f97316;border:1px solid #f9731630}
.pill.MEDIUM{background:#f59e0b15;color:#f59e0b;border:1px solid #f59e0b30}
.pill.LOW{background:#47556915;color:#94a3b8;border:1px solid #47556930}
.fc-name{font-size:15px;font-weight:600;color:#f1f5f9;letter-spacing:-0.01em}
.fc-id{font-size:11px;color:#64748b;margin-top:4px;font-family:'JetBrains Mono',monospace}
.owasp{display:inline-block;margin-left:8px;font-size:11px;color:#6366f1;background:#6366f115;border:1px solid #6366f130;border-radius:4px;padding:2px 6px}
.loc{display:inline-block;margin-top:10px;font-size:12px;font-family:'JetBrains Mono',monospace;color:#94a3b8;background:#0a0a0a;border:1px solid #1e2433;border-radius:4px;padding:4px 10px}
.loc-ln{color:#3b82f6;margin-left:4px}

/* Two-column body */
.fc-body{display:grid;grid-template-columns:1fr 1fr;gap:20px;margin-top:16px}
@media(max-width:640px){.fc-body{grid-template-columns:1fr}}
.sec-label{font-size:10px;font-weight:700;text-transform:uppercase;letter-spacing:.08em;color:#64748b;margin-bottom:8px}
.fc-desc{font-size:13px;color:#cbd5e1;line-height:1.6;background:#0a0a0a;border:1px solid #1e2433;border-radius:6px;padding:12px 16px}
.fc-fix{font-size:13px;color:#cbd5e1;line-height:1.6;background:#0a0a0a;border:1px solid #1e2433;border-radius:6px;padding:12px 16px}

/* Footer */
.footer{margin-top:60px;padding-top:20px;border-top:1px solid #1e2433;display:flex;justify-content:space-between;font-size:12px;color:#64748b}
</style>
</head>
<body>
<div class="wrap">

<div class="rh">
  <div>
    <div class="rh-title">AITriage Security Report <svg style="display:inline;vertical-align:-2px;margin-left:6px" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#64748b" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg></div>
    <div class="rh-meta">
      <span>Stacks: <strong style="color:#e2e8f0">{{range .Stacks}}{{.}} {{end}}</strong></span>
      <span>Findings: <strong style="color:#e2e8f0">{{len .Results}}</strong></span>
    </div>
  </div>
  <div style="text-align:right">
    <div style="font-size:11px;color:#64748b;margin-bottom:4px;text-transform:uppercase;letter-spacing:.08em;font-weight:600">Grade</div>
    <div class="grade g{{.SecurityGrade}}">{{.SecurityGrade}}</div>
  </div>
</div>

<div class="stats">
  <div class="sc"><div class="sc-label">Total Findings</div><div class="sc-val">{{len .Results}}</div></div>
  <div class="sc"><div class="sc-label">Critical Risks</div><div class="sc-val red">{{.CriticalCount}}</div></div>
  <div class="sc"><div class="sc-label">Security Grade</div><div class="sc-val g{{.SecurityGrade}}">{{.SecurityGrade}}</div></div>
  <div class="sc"><div class="sc-label">Target Stacks</div><div class="sc-val" style="font-size:18px;padding-top:6px">{{len .Stacks}}</div></div>
</div>

<div class="filters">
  <button class="fb on" onclick="fAll(this)">All ({{len .Results}})</button>
  <button class="fb" onclick="fSev('CRITICAL',this)">Critical <svg width="10" height="10" viewBox="0 0 24 24" fill="#ef4444" style="margin-left:4px"><polygon points="12,2 22,22 2,22"/></svg></button>
  <button class="fb" onclick="fSev('HIGH',this)">High <svg width="10" height="10" viewBox="0 0 24 24" fill="#f97316" style="margin-left:4px"><circle cx="12" cy="12" r="10"/></svg></button>
  <button class="fb" onclick="fSev('MEDIUM',this)">Medium <svg width="10" height="10" viewBox="0 0 24 24" fill="#f59e0b" style="margin-left:4px"><circle cx="12" cy="12" r="10"/></svg></button>
  <button class="fb" onclick="fSev('LOW',this)">Low <svg width="10" height="10" viewBox="0 0 24 24" fill="#475569" style="margin-left:4px"><circle cx="12" cy="12" r="10"/></svg></button>
  <button class="fb" onclick="fEntropy(this)">Entropy <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-left:4px"><rect x="4" y="4" width="16" height="16" rx="2"/><path d="M9 9h.01M15 9h.01M9 15h6"/></svg></button>
</div>

<main id="cards">
{{range .Results}}
<div class="fc {{.Severity}}" data-sev="{{.Severity}}" data-id="{{.ID}}">
  <div class="fc-top">
    <span class="pill {{.Severity}}">{{.Severity}}</span>
    <div>
      <div class="fc-name">{{.Name}}</div>
      <div class="fc-id">{{.ID}}{{if .OWASPMapping}}<span class="owasp">{{.OWASPMapping}}</span>{{end}}</div>
    </div>
  </div>
  {{if .File}}<div class="loc">{{.File}}{{if gt .Line 0}}<span class="loc-ln">:{{.Line}}</span>{{end}}</div>{{end}}
  <div class="fc-body">
    <div>
      <div class="sec-label">Audit Evidence</div>
      <div class="fc-desc">{{.Evidence}}</div>
    </div>
    <div>
      <div class="sec-label">Remediation Action</div>
      <div class="fc-fix">{{.Suggestion}}</div>
    </div>
  </div>
</div>
{{end}}
</main>

<div class="footer">
  <span>AITriage · Enterprise SAST Engine</span>
  <span id="ts"></span>
</div>

</div>
<script>
document.getElementById('ts').textContent=new Date().toISOString().replace('T',' ').slice(0,19)+' UTC';
function clr(){document.querySelectorAll('.fb').forEach(b=>b.classList.remove('on'))}
function fAll(b){clr();b.classList.add('on');document.querySelectorAll('.fc').forEach(c=>c.style.display='')}
function fSev(s,b){clr();b.classList.add('on');document.querySelectorAll('.fc').forEach(c=>c.style.display=c.dataset.sev===s?'':'none')}
function fEntropy(b){clr();b.classList.add('on');document.querySelectorAll('.fc').forEach(c=>c.style.display=c.dataset.id.startsWith('ENTR-')||c.dataset.id.startsWith('ENTR-')?'':'none')}
</script>
</body>
</html>`

type ReportData struct {
	SecurityGrade string
	CriticalCount int
	Stacks        []string
	Results       []core.CheckResult
}

func GenerateHTMLReport(outputPath string, data ReportData) error {
	t, err := template.New("dashboard").Parse(dashTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			fmt.Printf("warning: failed to close report file: %v\n", closeErr)
		}
	}()

	return t.Execute(f, data)
}
