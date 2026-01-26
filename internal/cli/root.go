package cli

import (
	"github.com/spf13/cobra"
)

var (
	debug bool
)

var rootCmd = &cobra.Command{
	Use:   "atl-cli",
	Short: "Minimal Atlassian Cloud CLI for Jira and Confluence",
	Long: `atl-cli is a lightweight, agent-friendly CLI for Atlassian Cloud
that supports querying Jira issues and Confluence pages.

All output is JSON to stdout, errors to stderr.`,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug output (redacts auth)")
}

func Execute() error {
	return rootCmd.Execute()
}
