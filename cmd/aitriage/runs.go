package main

import (
	"fmt"

	"github.com/cybertortuga/aitriage/internal/agent/runstore"
	"github.com/spf13/cobra"
)

var (
	runsDryRun bool
	runsForce  bool
)

var runsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Inspect and clean up local AITriage host-agent run bundles (aitriage-reports/)",
}

var runsListCmd = &cobra.Command{
	Use:   "list [project-path]",
	Short: "List host-agent run bundles for a project and their status",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := runsStore(args)
		if err != nil {
			return err
		}
		infos, err := store.Runs()
		if err != nil {
			return err
		}
		if len(infos) == 0 {
			fmt.Println("No runs found.")
			return nil
		}
		for _, in := range infos {
			fmt.Printf("%s  %-22s %s\n", in.RunID, in.Status, terminalMark(in.Terminal))
		}
		return nil
	},
}

var runsCleanCmd = &cobra.Command{
	Use:   "clean [project-path]",
	Short: "Remove finished run bundles (use --force to also remove unfinished runs)",
	Long: `Remove host-agent run bundles under <project>/aitriage-reports.

By default only finished runs (completed/failed) are removed; an interrupted run
is kept unless --force is given. --dry-run prints what would be removed without
deleting anything. The scope is strictly this project's aitriage-reports directory
(raw scan history under aitriage-reports/history is left untouched).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := runsStore(args)
		if err != nil {
			return err
		}
		infos, err := store.Runs()
		if err != nil {
			return err
		}
		var removed, kept int
		for _, in := range infos {
			if !in.Terminal && !runsForce {
				kept++
				continue
			}
			if runsDryRun {
				fmt.Printf("would remove %s (%s)\n", in.RunID, in.Status)
				removed++
				continue
			}
			if err := store.RemoveRun(in.RunID, runsForce); err != nil {
				fmt.Printf("skip %s: %v\n", in.RunID, err)
				kept++
				continue
			}
			fmt.Printf("removed %s (%s)\n", in.RunID, in.Status)
			removed++
		}
		verb := "removed"
		if runsDryRun {
			verb = "would remove"
		}
		fmt.Printf("\n%s %d run(s); kept %d.\n", verb, removed, kept)
		return nil
	},
}

func runsStore(args []string) (*runstore.Store, error) {
	root, err := resolveProjectRoot(args)
	if err != nil {
		return nil, err
	}
	return runstore.NewStore(root)
}

func terminalMark(terminal bool) string {
	if terminal {
		return "(finished)"
	}
	return "(in progress — kept unless --force)"
}

func init() {
	rootCmd.AddCommand(runsCmd)
	runsCmd.AddCommand(runsListCmd)
	runsCmd.AddCommand(runsCleanCmd)
	runsCleanCmd.Flags().BoolVar(&runsDryRun, "dry-run", false, "Show what would be removed without deleting")
	runsCleanCmd.Flags().BoolVar(&runsForce, "force", false, "Also remove unfinished runs")
}
