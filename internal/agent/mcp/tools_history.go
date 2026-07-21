package mcp

import (
	"context"
	"fmt"

	"github.com/cybertortuga/aitriage/internal/engine/history"
	"github.com/cybertortuga/aitriage/internal/scanner"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type historyInput struct {
	Path string `json:"path"`
}

type diffInput struct {
	Path string `json:"path"`
}

type historyResult struct {
	HasHistory        bool   `json:"has_history"`
	LastScanTime      string `json:"last_scan_time,omitempty"`
	LastSecurityScore int    `json:"last_security_score,omitempty"`
	Message           string `json:"message"`
}

type diffResult struct {
	PreviousScore int                 `json:"previous_score"`
	CurrentScore  int                 `json:"current_score"`
	Delta         int                 `json:"delta"`
	Diffs         []history.DiffEntry `json:"diffs"`
	Summary       string              `json:"summary"`
}

func registerHistoryTool(srv *mcp.Server, guard *PathGuard) {
	// Show last scan
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_last_scan",
		Description: "Show the most recent saved scan result for a project. Returns scan timestamp and SecurityScore.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input historyInput) (*mcp.CallToolResult, historyResult, error) {
		path, err := guard.Resolve(input.Path)
		if err != nil {
			return nil, historyResult{}, err
		}
		rec, err := history.LoadLast(path)
		if err != nil {
			return nil, historyResult{}, fmt.Errorf("history error: %w", err)
		}
		if rec == nil {
			return nil, historyResult{
				HasHistory: false,
				Message:    "No previous scan found. Run `aitriage scan` first to start tracking history.",
			}, nil
		}
		return nil, historyResult{
			HasHistory:        true,
			LastScanTime:      rec.Timestamp.Format("2006-01-02 15:04:05 UTC"),
			LastSecurityScore: rec.Report.SecurityScore,
			Message:           fmt.Sprintf("Last scan: %s, SecurityScore: %d/100 (%s)", rec.Timestamp.Format("2006-01-02 15:04"), rec.Report.SecurityScore, rec.Report.SecurityGrade),
		}, nil
	})

	// Diff: scan now vs last history
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_diff",
		Description: "Run a fresh scan and diff it against the previous saved scan. Shows new findings and what was fixed since last run.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input diffInput) (*mcp.CallToolResult, diffResult, error) {
		path, err := guard.Resolve(input.Path)
		if err != nil {
			return nil, diffResult{}, err
		}
		// Load previous
		prev, err := history.LoadLast(path)
		if err != nil {
			return nil, diffResult{}, fmt.Errorf("history load error: %w", err)
		}
		// Run current scan
		currReport, err := scanner.Scan(ctx, path, scanner.ScanOptions{})
		if err != nil {
			return nil, diffResult{}, fmt.Errorf("scan error: %w", err)
		}
		// Save current
		history.Save(path, currReport) //nolint:errcheck

		if prev == nil {
			history.Save(path, currReport) //nolint:errcheck
			return nil, diffResult{
				CurrentScore: currReport.SecurityScore,
				Summary:      "No previous scan to compare against. This scan has been saved as the baseline.",
			}, nil
		}

		diffs := history.Diff(prev.Report, currReport)
		summary := history.FormatDiff(diffs, prev.Report.SecurityScore, currReport.SecurityScore)

		return nil, diffResult{
			PreviousScore: prev.Report.SecurityScore,
			CurrentScore:  currReport.SecurityScore,
			Delta:         currReport.SecurityScore - prev.Report.SecurityScore,
			Diffs:         diffs,
			Summary:       summary,
		}, nil
	})
}
