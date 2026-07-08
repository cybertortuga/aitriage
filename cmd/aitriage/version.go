package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of AITriage",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("AITriage %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
