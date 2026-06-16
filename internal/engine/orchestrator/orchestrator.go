package orchestrator

import (
	"context"
	"log/slog"
	"sync"

	"github.com/cybertortuga/aitriage/internal/agent/architect"
	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/scanner"
	"github.com/cybertortuga/aitriage/internal/scanner/deployaudit"
	"github.com/cybertortuga/aitriage/internal/scanner/entropy"
	"github.com/cybertortuga/aitriage/internal/scanner/external"
	"github.com/cybertortuga/aitriage/internal/scanner/network"
	"github.com/cybertortuga/aitriage/internal/scanner/nfr"
)

// Options configuration for the scan engine.
type Options struct {
	ProjectPath  string
	ProbeHost    string
	ForceStack   string
	RunExternal  bool
	FullPortScan bool // scan all 65535 ports instead of common ones
}

// RunAllScanners executes all SAST, NFR, Deploy, Git, Network and architecture diagram generators concurrently.
func RunAllScanners(ctx context.Context, opts Options) llm.RichScanResult {
	var wg sync.WaitGroup
	var mu sync.Mutex
	result := llm.RichScanResult{ProjectPath: opts.ProjectPath}

	// 1: Core SAST
	wg.Add(1)
	go func() {
		defer wg.Done()
		r, err := scanner.Scan(ctx, opts.ProjectPath, scanner.ScanOptions{
			ForceStack: opts.ForceStack,
		})
		if err == nil {
			mu.Lock()
			result.Report = r
			mu.Unlock()
		}
	}()

	// 2: External Scanners
	if opts.RunExternal {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var scanners [][]external.UnifiedFinding
			var swg sync.WaitGroup

			if external.IsInstalled("semgrep") {
				swg.Add(1)
				go func() {
					defer swg.Done()
					findings, err := external.RunSemgrep(ctx, opts.ProjectPath, "auto")
					if err == nil {
						mu.Lock()
						scanners = append(scanners, findings)
						mu.Unlock()
					}
				}()
			} else {
				slog.Debug("Semgrep not installed, skipping")
			}
			if external.IsInstalled("gitleaks") {
				swg.Add(1)
				go func() {
					defer swg.Done()
					findings, err := external.RunGitleaks(ctx, opts.ProjectPath)
					if err == nil {
						mu.Lock()
						scanners = append(scanners, findings)
						mu.Unlock()
					}
				}()
			} else {
				slog.Debug("Gitleaks not installed, skipping")
			}
			if external.IsInstalled("bandit") {
				swg.Add(1)
				go func() {
					defer swg.Done()
					findings, err := external.RunBandit(ctx, opts.ProjectPath)
					if err == nil {
						mu.Lock()
						scanners = append(scanners, findings)
						mu.Unlock()
					}
				}()
			} else {
				slog.Debug("Bandit not installed, skipping")
			}
			if external.IsInstalled("trivy") {
				for _, scanType := range []string{"fs", "config"} {
					st := scanType
					swg.Add(1)
					go func() {
						defer swg.Done()
						findings, err := external.RunTrivy(ctx, opts.ProjectPath, st)
						if err == nil {
							mu.Lock()
							scanners = append(scanners, findings)
							mu.Unlock()
						}
					}()
				}
			} else {
				slog.Debug("Trivy not installed, skipping")
			}

			swg.Wait()
			mu.Lock()
			for _, f := range scanners {
				result.External = append(result.External, f...)
			}
			mu.Unlock()
		}()
	}

	// 3: NFR Checks (now using embedded filesystem)
	wg.Add(1)
	go func() {
		defer wg.Done()
		nfrFindings, err := nfr.CheckNFR(opts.ProjectPath)
		if err == nil {
			mu.Lock()
			result.NFR = nfrFindings
			mu.Unlock()
		}
	}()

	// 4: DeployAudit (IaC)
	wg.Add(1)
	go func() {
		defer wg.Done()
		findings, err := deployaudit.AuditDeployFiles(opts.ProjectPath)
		if err == nil {
			mu.Lock()
			result.Deploy = findings
			mu.Unlock()
		}
	}()

	// 5: Git Deep Analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		critFiles := entropy.FindCriticalFiles(opts.ProjectPath)
		historyLeaks := entropy.ScanGitHistory(opts.ProjectPath)
		if len(critFiles) > 0 || len(historyLeaks) > 0 {
			mu.Lock()
			result.CriticalFiles = critFiles
			result.HistoryLeaks = historyLeaks
			mu.Unlock()
		}
	}()

	// 6: Architecture Diagram
	wg.Add(1)
	go func() {
		defer wg.Done()
		diag, err := architect.GenerateMermaidDiagram(opts.ProjectPath)
		if err == nil {
			mu.Lock()
			result.Diagram = diag
			mu.Unlock()
		}
	}()

	// 7: Network Probe
	wg.Add(1)
	go func() {
		defer wg.Done()
		var netFindings []network.NetworkFinding

		// Probe Docker Compose hosts if present
		if composeFindings := network.ProbeDockerCompose(opts.ProjectPath, opts.FullPortScan); len(composeFindings) > 0 {
			netFindings = append(netFindings, composeFindings...)
		}

		// Probe specific target if provided
		if opts.ProbeHost != "" {
			if targetFindings := network.ProbeHost(opts.ProbeHost, opts.FullPortScan); len(targetFindings) > 0 {
				netFindings = append(netFindings, targetFindings...)
			}
		}

		if len(netFindings) > 0 {
			mu.Lock()
			// Deduplicate if needed, though ProbeDockerCompose and ProbeHost might have different targets
			result.Network = netFindings
			mu.Unlock()
		}
	}()

	wg.Wait()
	return result
}
