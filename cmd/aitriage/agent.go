package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/cybertortuga/aitriage/internal/agent/graph"
	"github.com/cybertortuga/aitriage/internal/agent/llm"
	"github.com/cybertortuga/aitriage/internal/config"
	"github.com/cybertortuga/aitriage/internal/engine/core"
	"github.com/cybertortuga/aitriage/internal/engine/orchestrator"
	"github.com/cybertortuga/aitriage/internal/healthpolicy"
	"github.com/cybertortuga/aitriage/internal/report/healthcheck"
	"github.com/spf13/cobra"
)

var (
	agentProvider      string
	agentModel         string
	agentAPIKey        string
	agentNoChat        bool
	agentOutput        string
	agentProbe         string
	agentFullScan      bool
	agentRuleID        string
	agentTargetFile    string
	agentTargetLine    int
	agentReportOut     string
	agentFixSpecOut    string
	agentSummaryOut    string
	agentFailOn        string
	agentFailScore     int
	agentHealthProfile string
)

var agentCmd = &cobra.Command{
	Use:   "agent [path]",
	Short: "Run AI-powered security audit with LLM analysis and interactive Q&A",
	Long: `Run a full security scan, then use an LLM to triage findings, generate
a prioritized report, and produce an AI fix specification.

The LLM provider is auto-detected from environment variables:
  GEMINI_API_KEY    → Google Gemini (default model: gemini-2.0-flash)
  ANTHROPIC_API_KEY → Anthropic Claude
  OPENAI_API_KEY    → OpenAI GPT

You can also configure it in .aitriage.yaml under the "llm:" section.`,
	Example: `  aitriage agent                        # Audit current directory
  aitriage agent ./my-project           # Audit specific path
  aitriage agent --no-chat              # Skip interactive Q&A (CI/CD)
  aitriage agent --provider gemini      # Force specific provider
  aitriage agent --model gemini-1.5-pro # Use specific model`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAgent,
}

func init() {
	rootCmd.AddCommand(agentCmd)
	agentCmd.Flags().StringVar(&agentProvider, "provider", "", "LLM provider: gemini, anthropic, openai, ollama, groq (auto-detected from env)")
	agentCmd.Flags().StringVar(&agentModel, "model", "", "LLM model name")
	agentCmd.Flags().StringVar(&agentAPIKey, "api-key", "", "LLM API key (or set via env)")
	agentCmd.Flags().BoolVar(&agentNoChat, "no-chat", false, "Skip interactive Q&A (for CI/CD)")
	agentCmd.Flags().StringVar(&agentOutput, "output", "text", "Output format: text | json | md")
	agentCmd.Flags().StringVar(&agentProbe, "probe", "", "Target host to probe for open DBs/Services (e.g. localhost, example.com)")
	agentCmd.Flags().BoolVar(&agentFullScan, "full-scan", false, "Probe all 65535 ports (slow, ~30-60s)")
	agentCmd.Flags().StringVar(&agentRuleID, "rule-id", "", "Target a specific rule ID to fix")
	agentCmd.Flags().StringVar(&agentTargetFile, "file", "", "Target a specific file to fix (used with --rule-id)")
	agentCmd.Flags().IntVar(&agentTargetLine, "line", 0, "Target a specific line to fix (used with --rule-id)")
	agentCmd.Flags().StringVar(&agentReportOut, "report-out", "", "Write the final Markdown triage report to this file (for CI/CD)")
	agentCmd.Flags().StringVar(&agentFixSpecOut, "fixspec-out", "", "Write the AI fix specification to this file (for CI/CD)")
	agentCmd.Flags().StringVar(&agentSummaryOut, "summary-out", "", "Write the actionable summary (TP/NR only, no FP) to this file")
	agentCmd.Flags().StringVar(&agentFailOn, "fail-on", "never", "CI gate: exit 1 when 'critical' (active CRITICAL/HIGH after AI triage), 'any' finding, or 'never'")
	agentCmd.Flags().IntVar(&agentFailScore, "fail-score", 0, "CI gate: exit 1 if the post-AI Health Check score is below this threshold (0 = disabled)")
	agentCmd.Flags().StringVar(&agentHealthProfile, "health-profile", "", "Health Check policy profile: baseline, standard, strict")
}

func runAgent(cmd *cobra.Command, args []string) error {
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}
	ctx := context.Background()

	// Load config
	cfg := config.LoadConfig(projectPath)
	policy := agentPolicyFromFlags(cmd, cfg)

	// CLI flags override config file values
	llmCfg := cfg.LLM
	if agentProvider != "" {
		llmCfg.Provider = agentProvider
	}
	if agentModel != "" {
		llmCfg.Model = agentModel
	}
	if agentAPIKey != "" {
		llmCfg.APIKey = agentAPIKey
	}

	// Create LLM client
	client, err := llm.NewClient(llm.Config{
		Provider: llmCfg.Provider,
		Model:    llmCfg.Model,
		APIKey:   llmCfg.APIKey,
		BaseURL:  llmCfg.BaseURL,
		Timeout:  llmCfg.Timeout,
	})
	if err != nil {
		cmd.SilenceUsage = true
		return err
	}

	fmt.Fprintf(os.Stderr, "🔍 AITriage Agent starting...\n\n")

	// STEP 1: PARALLEL SCANNING
	fmt.Fprintf(os.Stderr, "📡 Step 1/3: Scanning (parallel)...\n")

	richResult := orchestrator.RunAllScanners(ctx, orchestrator.Options{
		ProjectPath:  projectPath,
		ProbeHost:    agentProbe,
		RunExternal:  true,
		FullPortScan: agentFullScan,
	})

	fmt.Fprintf(os.Stderr, "   ✓ AITriage: %d findings\n", len(richResult.Report.Results))
	fmt.Fprintf(os.Stderr, "   ✓ External: %d findings\n", len(richResult.External))
	fmt.Fprintf(os.Stderr, "   ✓ NFR: %d issues\n", len(richResult.NFR))
	fmt.Fprintf(os.Stderr, "   ✓ Deploy: %d issues\n", len(richResult.Deploy))

	if agentProbe != "" {
		fmt.Fprintf(os.Stderr, "   ✓ Network: %d ports open\n", len(richResult.Network))
	}
	fmt.Fprintf(os.Stderr, "   Health Check (pre-AI, core-only): %s (%d/100)\n\n", richResult.Report.SecurityGrade, richResult.Report.SecurityScore)

	// Filter results if target flags are provided
	if agentRuleID != "" {
		var filteredResults []core.CheckResult
		for _, r := range richResult.Report.Results {
			if r.ID == agentRuleID {
				if agentTargetFile != "" && r.File != agentTargetFile {
					continue
				}
				if agentTargetLine > 0 && r.Line != agentTargetLine {
					continue
				}
				filteredResults = append(filteredResults, r)
			}
		}
		richResult.Report.Results = filteredResults

		// Clear other findings as we are focusing on one specific issue
		richResult.External = nil
		richResult.NFR = nil
		richResult.Deploy = nil
		richResult.Network = nil

		fmt.Fprintf(os.Stderr, "🎯 Targeted Mode: Focusing on finding %s in %s:%d\n\n", agentRuleID, agentTargetFile, agentTargetLine)
	}

	// STEP 2: LLM ANALYSIS (Map-Reduce Graph)
	fmt.Fprintf(os.Stderr, "🤖 Step 2/3: LLM Analysis (Map-Reduce)...\n")

	state := &graph.AgentState{
		ProjectPath:      projectPath,
		DeepScan:         true,
		CoreFindings:     richResult.Report.Results,
		ExternalFindings: richResult.External,
		NFRFindings:      richResult.NFR,
		DeployFindings:   richResult.Deploy,
		NetworkFindings:  richResult.Network,
		SecurityScore:    richResult.Report.SecurityScore,
		SecurityGrade:    richResult.Report.SecurityGrade,
		Policy:           policy,
		Diagram:          richResult.Diagram,
		CriticalFiles:    richResult.CriticalFiles,
		HistoryLeaks:     richResult.HistoryLeaks,
	}

	if err := graph.Run(ctx, state, client); err != nil {
		return fmt.Errorf("LLM analysis failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\n📝 FINAL REPORT:\n\n%s\n\n", state.ReportMarkdown)
	fmt.Fprintf(os.Stderr, "🛠 AI FIX SPECIFICATION:\n\n%s\n\n", state.AIFixSpec)

	// Persist artifacts to files when requested (CI/CD consumes these).
	if agentReportOut != "" {
		if err := os.WriteFile(agentReportOut, []byte(state.ReportMarkdown), 0644); err != nil {
			return fmt.Errorf("failed to write report to %s: %w", agentReportOut, err)
		}
		fmt.Fprintf(os.Stderr, "   ✓ Report written to %s\n", agentReportOut)
	}
	if agentFixSpecOut != "" {
		if err := os.WriteFile(agentFixSpecOut, []byte(state.AIFixSpec), 0644); err != nil {
			return fmt.Errorf("failed to write fix spec to %s: %w", agentFixSpecOut, err)
		}
		fmt.Fprintf(os.Stderr, "   ✓ Fix spec written to %s\n", agentFixSpecOut)
	}
	if agentSummaryOut != "" {
		if err := os.WriteFile(agentSummaryOut, []byte(state.SummaryMarkdown), 0644); err != nil {
			return fmt.Errorf("failed to write summary to %s: %w", agentSummaryOut, err)
		}
		fmt.Fprintf(os.Stderr, "   ✓ Summary written to %s\n", agentSummaryOut)
	}

	// Auto-write actionable summary to GitHub Actions Step Summary.
	// The agent writes a clean, FP-free summary — workflows no longer need
	// to `cat report.md >> $GITHUB_STEP_SUMMARY`.
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		writeAgentGHASummary(state)
	}

	// CI/CD GATE: decide the exit code from the post-AI Health Check verdict.
	// False Positives are already excluded from state.HealthCheck, so the gate
	// only trips on findings the AI triage considered real.
	shouldFail := !state.HealthCheck.Verdict.Passed

	// STEP 3: INTERACTIVE CONSULTATION
	if !agentNoChat {
		fmt.Fprintf(os.Stderr, "💬 Step 3/3: Consultation mode (type 'exit' to quit)\n")

		messages := []llm.Message{
			{Role: "system", Content: "You are an expert security consultant. You have generated a report and fix spec. Answer the user's questions based on this context."},
			{Role: "user", Content: "Context: \nReport:\n" + state.ReportMarkdown + "\nFix Spec:\n" + state.AIFixSpec},
			{Role: "assistant", Content: "I have loaded the security context. How can I help you?"},
		}

		runConsultation(ctx, client, messages)
	}

	if shouldFail {
		printPolicyFailure(os.Stderr, state.HealthCheck.Verdict)
		cmd.SilenceErrors = true
		return ErrPolicyViolation
	}
	return nil
}

func agentPolicyFromFlags(cmd *cobra.Command, cfg *config.Config) healthcheck.Policy {
	policy := healthpolicy.FromConfig(cfg)
	if !healthpolicy.HasConfiguredGate(cfg) {
		policy.FailOn = healthcheck.FailOnNever
		policy.MinimumScore = 0
	}

	failOnSet := cmd.Flags().Changed("fail-on")
	failScoreSet := cmd.Flags().Changed("fail-score")
	policy = healthpolicy.ApplyOverrides(policy, healthpolicy.Overrides{
		Profile:         agentHealthProfile,
		ProfileSet:      cmd.Flags().Changed("health-profile"),
		FailOn:          agentFailOn,
		FailOnSet:       failOnSet,
		MinimumScore:    agentFailScore,
		MinimumScoreSet: failScoreSet,
	})
	if failScoreSet && !failOnSet && policy.FailOn == healthcheck.FailOnNever {
		policy.FailOn = healthcheck.FailOnCritical
	}
	return healthcheck.NormalizePolicy(policy)
}

func runConsultation(ctx context.Context, client llm.Client, history []llm.Message) {
	scan := bufio.NewScanner(os.Stdin)
	fmt.Print("\n> ")
	for scan.Scan() {
		question := scan.Text()
		if question == "exit" || question == "quit" {
			break
		}
		if question == "" {
			fmt.Print("> ")
			continue
		}
		history = append(history, llm.Message{
			Role:    "user",
			Content: question,
		})
		answer, _, err := client.Chat(ctx, history)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		} else {
			history = append(history, llm.Message{Role: "assistant", Content: answer})
			fmt.Println(answer)
		}
		fmt.Print("\n> ")
	}
}

// writeAgentGHASummary writes the actionable summary (TP/NR only) to the
// GitHub Actions Step Summary. This runs automatically when GITHUB_ACTIONS=true,
// so workflows no longer need to `cat report.md >> $GITHUB_STEP_SUMMARY`.
func writeAgentGHASummary(state *graph.AgentState) {
	summaryFile := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryFile == "" {
		return
	}
	f, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "   ⚠ Failed to write GHA Step Summary: %v\n", err)
		return
	}
	defer f.Close()
	_, _ = f.WriteString(state.SummaryMarkdown)
	fmt.Fprintf(os.Stderr, "   ✓ GHA Step Summary written (actionable findings only)\n")
}
