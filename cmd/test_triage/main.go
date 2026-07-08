package main

import (
	"fmt"

	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/scanner"
	"github.com/cybertortuga/aitriage/internal/ui/tui"
)

func main() {
	m := tui.InitialModel(llm.RichScanResult{
		Report: scanner.ScanReport{
			Results: []core.CheckResult{
				{ID: "TEST1", File: "main.go", Line: 10, Severity: "CRITICAL", Name: "Test Vulnerability"},
			},
		},
	}, "dev")

	m.Ready = true
	m.Width = 100
	m.Height = 40
	m.ActiveView = tui.ViewTriage
	fmt.Print(m.View())
}
