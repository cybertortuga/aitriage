package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cybertortuga/aitriage/internal/agent/remedy"
	"github.com/cybertortuga/aitriage/internal/engine/orchestrator"
	"github.com/spf13/cobra"
)

var specCmd = &cobra.Command{
	Use:   "generate-spec [path]",
	Short: "Generate AI-Agent specification file (CLAUDE.md) based on scan results",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath := args[0]
		ctx := context.Background()

		fmt.Fprintf(os.Stderr, "📡 Scanning project to generate AI Spec...\n")

		probeArg, _ := cmd.Flags().GetString("probe")
		richResult := orchestrator.RunAllScanners(ctx, orchestrator.Options{
			ProjectPath: projectPath,
			ProbeHost:   probeArg,
			RunExternal: true,
		})

		specMarkdown := remedy.GenerateClaudeSpec(richResult)

		outPath := filepath.Join(projectPath, "CLAUDE.md")
		err := os.WriteFile(outPath, []byte(specMarkdown), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠ Failed to write CLAUDE.md: %v\n\n", err)
			fmt.Println(specMarkdown)
		} else {
			fmt.Fprintf(os.Stderr, "✅ Successfully generated: %s\n", outPath)
		}

		return nil
	},
}

func init() {
	specCmd.Flags().String("probe", "", "Target host to probe for open DBs/Services")
	rootCmd.AddCommand(specCmd)
}
