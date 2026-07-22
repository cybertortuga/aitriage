package orchestrator

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cybertortuga/aitriage/internal/agent/architect"
	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/scanner"
	"github.com/cybertortuga/aitriage/internal/scanner/deployaudit"
	"github.com/cybertortuga/aitriage/internal/scanner/entropy"
	"github.com/cybertortuga/aitriage/internal/scanner/external"
	"github.com/cybertortuga/aitriage/internal/scanner/network"
	"github.com/cybertortuga/aitriage/internal/scanner/nfr"
	"github.com/cybertortuga/aitriage/rules"
)

// Options configuration for the scan engine.
type Options struct {
	ProjectPath  string
	ProbeHost    string
	ForceStack   string
	RunExternal  bool
	FullPortScan bool // scan all 65535 ports instead of common ones
}

// redactScannerError returns a short, safe scanner error string for the
// execution manifest: it strips absolute paths and caps length so no secret or
// full environment dump reaches persisted evidence.
func redactScannerError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		msg = msg[:i]
	}
	msg = absPathRegex.ReplaceAllString(msg, "<path>")
	if len(msg) > 200 {
		msg = msg[:200] + "…"
	}
	return msg
}

var absPathRegex = regexp.MustCompile(`(/[^\s:]+){2,}`)

const externalScannerTimeout = 10 * time.Minute

// RunAllScanners executes all SAST, NFR, Deploy, Git, Network and architecture diagram generators concurrently.
func RunAllScanners(ctx context.Context, opts Options) llm.RichScanResult {
	var wg sync.WaitGroup
	var mu sync.Mutex
	// Keep collection fields non-nil even when a scanner legitimately returns no
	// findings. Callers use nil to distinguish "scanner did not initialize" from
	// "scanner completed with zero findings".
	result := llm.RichScanResult{
		ProjectPath: opts.ProjectPath,
		NFR:         []nfr.NFRFinding{},
	}

	// 1: Core SAST
	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		r, err := scanner.Scan(ctx, opts.ProjectPath, scanner.ScanOptions{
			ForceStack: opts.ForceStack,
		})
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			result.ScannerExecutions = append(result.ScannerExecutions, external.ScannerExecution{
				Scanner: "aitriage", Status: external.StatusFailed,
				DurationMs: time.Since(start).Milliseconds(), Error: redactScannerError(err),
			})
		} else {
			result.Report = r
			result.ScannerExecutions = append(result.ScannerExecutions, external.ScannerExecution{
				Scanner: "aitriage", Status: external.StatusCompleted,
				Findings: len(r.Results), DurationMs: time.Since(start).Milliseconds(),
			})
		}
	}()

	// 2: External Scanners — every scanner records a typed execution status so a
	// full audit can never silently skip a mandatory scanner. A missing binary is
	// recorded as "missing" (not omitted); an error as "failed".
	// The trusted taint config is generated once, from the compiled-in rule
	// catalog, into an owner-only temp file. It is built up-front so that a
	// failure to load the mandatory taint rules fails the full audit closed (via
	// the semgrep execution status) rather than silently downgrading coverage.
	var taintCfgPath string
	var taintRuleIDs []string
	var taintErr error
	if opts.RunExternal {
		if dir, err := os.MkdirTemp("", "aitriage-taint-"); err != nil {
			taintErr = fmt.Errorf("create trusted rules dir: %w", err)
		} else {
			defer func() { _ = os.RemoveAll(dir) }()
			taintCfgPath, taintRuleIDs, taintErr = rules.WriteTaintConfig(dir)
		}
	}

	if opts.RunExternal {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var swg sync.WaitGroup

			record := func(ex external.ScannerExecution) {
				mu.Lock()
				result.ScannerExecutions = append(result.ScannerExecutions, ex)
				mu.Unlock()
			}
			addFindings := func(f []external.UnifiedFinding) {
				mu.Lock()
				result.External = append(result.External, f...)
				mu.Unlock()
			}
			// run executes one scanner with a hard upper bound, records its status,
			// and appends findings. Request cancellation still propagates through the
			// parent context; an abandoned Web/MCP run cannot leak child processes.
			run := func(name, label string, fn func(context.Context) ([]external.UnifiedFinding, error)) {
				defer swg.Done()
				if !external.IsInstalled(name) {
					record(external.ScannerExecution{Scanner: label, Status: external.StatusMissing})
					fmt.Fprintf(os.Stderr, "   ▶ %s — MISSING (not installed)\n", label)
					return
				}
				version := external.ToolVersion(ctx, name)
				start := time.Now()
				scanCtx, cancel := context.WithTimeout(ctx, externalScannerTimeout)
				defer cancel()
				findings, err := fn(scanCtx)
				dur := time.Since(start).Milliseconds()
				if err != nil {
					status := external.StatusFailed
					if scanCtx.Err() == context.DeadlineExceeded {
						status = external.StatusTimedOut
					}
					record(external.ScannerExecution{Scanner: label, Status: status, Version: version, DurationMs: dur, Error: redactScannerError(err)})
					fmt.Fprintf(os.Stderr, "   ▶ %s ✗ FAILED: %v\n", label, err)
					return
				}
				findings = external.FilterTestLikeFindings(findings)
				addFindings(findings)
				record(external.ScannerExecution{Scanner: label, Status: external.StatusCompleted, Version: version, Findings: len(findings), DurationMs: dur})
				fmt.Fprintf(os.Stderr, "   ▶ %s ✓ %d findings (%dms)\n", label, len(findings), dur)
			}

			swg.Add(1)
			go run("semgrep", "semgrep", func(scanCtx context.Context) ([]external.UnifiedFinding, error) {
				// Mandatory taint rules must be loadable; otherwise the full audit
				// fails closed instead of running Semgrep without them.
				if taintErr != nil {
					return nil, fmt.Errorf("mandatory taint rules unavailable: %w", taintErr)
				}
				// Registry "auto" rules and the trusted AITriage taint rules run
				// simultaneously in one Semgrep pass.
				return external.RunSemgrepConfigs(scanCtx, opts.ProjectPath, taintRuleIDs, "auto", taintCfgPath)
			})
			swg.Add(1)
			go run("gitleaks", "gitleaks", func(scanCtx context.Context) ([]external.UnifiedFinding, error) {
				return external.RunGitleaks(scanCtx, opts.ProjectPath)
			})
			swg.Add(1)
			go run("bandit", "bandit", func(scanCtx context.Context) ([]external.UnifiedFinding, error) {
				return external.RunBandit(scanCtx, opts.ProjectPath)
			})
			for _, scanType := range []string{"fs", "config"} {
				st := scanType
				swg.Add(1)
				go run("trivy", "trivy_"+st, func(scanCtx context.Context) ([]external.UnifiedFinding, error) {
					return external.RunTrivy(scanCtx, opts.ProjectPath, st)
				})
			}

			swg.Wait()
		}()
	}

	// 3: NFR Checks (now using embedded filesystem)
	wg.Add(1)
	go func() {
		defer wg.Done()
		nfrFindings, err := nfr.CheckNFR(opts.ProjectPath)
		if err == nil {
			if nfrFindings == nil {
				nfrFindings = []nfr.NFRFinding{}
			}
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
	sort.Slice(result.ScannerExecutions, func(i, j int) bool {
		return result.ScannerExecutions[i].Scanner < result.ScannerExecutions[j].Scanner
	})
	return result
}
