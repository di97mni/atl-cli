package cli

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		info := map[string]string{
			"version":    Version,
			"git_commit": GitCommit,
			"build_date": BuildDate,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(info)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
