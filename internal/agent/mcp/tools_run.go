package mcp

import (
	"context"
	"fmt"

	"github.com/cybertortuga/aitriage/internal/agent/hostagent"
	"github.com/cybertortuga/aitriage/internal/agent/hostrun"
	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/agent/pipeline"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// The run tools expose the full deferred AI-triage pipeline (scanners +
// graph.Run + canonical artifact + post-AI gate) driven by the host agent's own
// model session. They are the real "run a security check" entry point; the raw
// aitriage_scan tool is only a deterministic pre-scan and never a final verdict.
//
// AITriage never edits source. aitriage_run_approve records an auditable user
// decision; the active coding agent applies the fix afterward and calls
// aitriage_run_verify to prove the after-fix state.
//
// These tools are registered only when a scan root is configured (the safe
// profile), so every run is confined to one project directory.

func registerRunTools(srv *mcp.Server, guard *PathGuard, version string) {
	if guard == nil || guard.Root() == "" {
		return // run workflow requires a confined project root
	}
	root := guard.Root()

	newManager := func() (*hostrun.Manager, error) {
		return hostrun.NewManager(root, version)
	}

	// ── aitriage_run_start ────────────────────────────────────────────────────
	mcp.AddTool(srv, &mcp.Tool{
		Name: "aitriage_run_start",
		Description: "Start a FULL AITriage security check: runs all deterministic scanners and the AI triage graph, then defers each model request to YOU (the host agent). This is the correct entry point for a security review — not aitriage_scan. " +
			"When the user names a project/subproject, pass its path; it may be relative to the configured repository root and is safely confined inside that root. " +
			"Returns a run_id and either the next exact request to answer (submit via aitriage_run_submit) or the final result. Answer requests with your own session; do not substitute your own analysis for the prompt.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input runStartInput) (*mcp.CallToolResult, hostrun.Progress, error) {
		projectPath, err := guard.Resolve(input.Path)
		if err != nil {
			return nil, hostrun.Progress{}, err
		}
		mgr, err := newManager()
		if err != nil {
			return nil, hostrun.Progress{}, err
		}
		opts := hostrun.StartOptions{
			ProjectPath: projectPath,
			HostClient:  sanitizeHostClient(input.HostClient),
			Intent:      input.Intent, // sanitised by the run store; default "audit"
			Policy:      hostrun.AuditPolicy(projectPath),
			Scan:        pipeline.ScanOptions{RunExternal: true},
			LLM:         pipeline.LLMIdentity{Provider: input.Provider, Model: input.Model, BatchSize: input.BatchSize},
		}
		prog, err := mgr.Start(ctx, opts)
		if err != nil {
			return nil, hostrun.Progress{}, err
		}
		return nil, *prog, nil
	})

	// ── aitriage_run_submit ───────────────────────────────────────────────────
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_run_submit",
		Description: "Submit your model's answer to a pending AITriage request. Use the SAME request_id you received. Include usage only if your session actually reports token counts. Advances the run to the next request or the final result.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input runSubmitInput) (*mcp.CallToolResult, hostrun.Progress, error) {
		if input.RunID == "" || input.RequestID == "" {
			return nil, hostrun.Progress{}, fmt.Errorf("run_id and request_id are required")
		}
		mgr, err := newManager()
		if err != nil {
			return nil, hostrun.Progress{}, err
		}
		resp := hostagent.Response{
			RequestID:     input.RequestID,
			Content:       input.Content,
			UsageReported: input.UsageReported,
		}
		if input.UsageReported {
			resp.Usage = llm.Usage{
				PromptTokens:     input.PromptTokens,
				CompletionTokens: input.CompletionTokens,
				TotalTokens:      input.TotalTokens,
			}
		}
		prog, err := mgr.Submit(ctx, input.RunID, resp)
		if err != nil {
			return nil, hostrun.Progress{}, err
		}
		return nil, *prog, nil
	})

	// ── aitriage_run_status ───────────────────────────────────────────────────
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_run_status",
		Description: "Report a run's current state, progress, any pending request, artifact paths and the gate verdict (when finalized). Does not execute the pipeline.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input runIDInput) (*mcp.CallToolResult, hostrun.Progress, error) {
		mgr, err := newManager()
		if err != nil {
			return nil, hostrun.Progress{}, err
		}
		prog, err := mgr.Status(input.RunID)
		if err != nil {
			return nil, hostrun.Progress{}, err
		}
		return nil, *prog, nil
	})

	// ── aitriage_run_continue ─────────────────────────────────────────────────
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_run_continue",
		Description: "Resume a run after an interruption without submitting a new answer. Replays confirmed answers and stops at the next missing request or the final result.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input runIDInput) (*mcp.CallToolResult, hostrun.Progress, error) {
		mgr, err := newManager()
		if err != nil {
			return nil, hostrun.Progress{}, err
		}
		prog, err := mgr.Continue(ctx, input.RunID)
		if err != nil {
			return nil, hostrun.Progress{}, err
		}
		return nil, *prog, nil
	})

	// ── aitriage_run_approve ──────────────────────────────────────────────────
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_run_approve",
		Description: "Record the USER's explicit decision to fix specific findings. Call ONLY after the user has seen the gate and canonical finding IDs and chosen. selected_ids must be canonical fixable IDs from the result. This authorises YOU to apply the fix spec for those IDs; AITriage never edits source itself. Then call aitriage_run_verify.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input runApproveInput) (*mcp.CallToolResult, hostrun.Progress, error) {
		mgr, err := newManager()
		if err != nil {
			return nil, hostrun.Progress{}, err
		}
		prog, err := mgr.Approve(input.RunID, input.SelectedIDs)
		if err != nil {
			return nil, hostrun.Progress{}, err
		}
		return nil, *prog, nil
	})

	// ── aitriage_run_decline ──────────────────────────────────────────────────
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_run_decline",
		Description: "Record the user's decision to make NO changes and complete the run without any source modification.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input runIDInput) (*mcp.CallToolResult, hostrun.Progress, error) {
		mgr, err := newManager()
		if err != nil {
			return nil, hostrun.Progress{}, err
		}
		prog, err := mgr.Decline(input.RunID)
		if err != nil {
			return nil, hostrun.Progress{}, err
		}
		return nil, *prog, nil
	})

	// ── aitriage_run_verify ───────────────────────────────────────────────────
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "aitriage_run_verify",
		Description: "After you have applied the approved fixes, start a linked verification run on the current tree. Allowed only once an approval record exists. Answer its requests like a normal run; when it finalizes, the before/after gate is recorded.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input runIDInput) (*mcp.CallToolResult, hostrun.Progress, error) {
		mgr, err := newManager()
		if err != nil {
			return nil, hostrun.Progress{}, err
		}
		opts := hostrun.StartOptions{
			Policy: hostrun.AuditPolicy(root),
			Scan:   pipeline.ScanOptions{RunExternal: true},
		}
		prog, err := mgr.Verify(ctx, input.RunID, opts)
		if err != nil {
			return nil, hostrun.Progress{}, err
		}
		return nil, *prog, nil
	})
}

// ── tool input types ──────────────────────────────────────────────────────────

type runStartInput struct {
	// Path selects the repository root itself (empty/".") or any project below
	// it. PathGuard rejects traversal, symlink escapes, and external paths.
	Path string `json:"path,omitempty"`
	// Intent must come only from the user's actual command: "audit" (default,
	// read-only) or "audit_and_fix".
	Intent     string `json:"intent,omitempty"`
	HostClient string `json:"host_client,omitempty"` // "codex" | "claude-code"
	Provider   string `json:"provider,omitempty"`
	Model      string `json:"model,omitempty"`
	BatchSize  int    `json:"batch_size,omitempty"`
}

type runSubmitInput struct {
	RunID            string `json:"run_id"`
	RequestID        string `json:"request_id"`
	Content          string `json:"content"`
	UsageReported    bool   `json:"usage_reported,omitempty"`
	PromptTokens     int    `json:"prompt_tokens,omitempty"`
	CompletionTokens int    `json:"completion_tokens,omitempty"`
	TotalTokens      int    `json:"total_tokens,omitempty"`
}

type runIDInput struct {
	RunID string `json:"run_id"`
}

type runApproveInput struct {
	RunID       string   `json:"run_id"`
	SelectedIDs []string `json:"selected_ids"`
}

func sanitizeHostClient(c string) string {
	switch c {
	case "codex", "claude-code":
		return c
	default:
		return ""
	}
}
