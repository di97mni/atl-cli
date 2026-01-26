package cli

import (
	"context"
	"os"

	"github.com/martin/atl-cli/internal/config"
	"github.com/martin/atl-cli/internal/confluence"
	"github.com/martin/atl-cli/internal/httpclient"
	"github.com/spf13/cobra"
)

var confluenceCmd = &cobra.Command{
	Use:   "confluence",
	Short: "Confluence commands",
	Long:  "Commands for interacting with Confluence",
}

var confluencePageCmd = &cobra.Command{
	Use:   "page",
	Short: "Confluence page commands",
	Long:  "Commands for working with Confluence pages",
}

var confluencePageGetCmd = &cobra.Command{
	Use:   "get <page-id>",
	Short: "Get a Confluence page by ID",
	Long:  "Retrieves content of a Confluence page and outputs as JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pageID := args[0]

		// Load and validate config
		cfg, err := config.LoadFromEnv()
		if err != nil {
			return outputError(httpclient.NewConfigError(err.Error()))
		}
		if err := cfg.Validate(); err != nil {
			return outputError(httpclient.NewConfigError(err.Error()))
		}

		// Create client with debug flag from root command
		client := confluence.NewClient(cfg, debug)

		// Get page
		page, err := client.GetPage(context.Background(), pageID)
		if err != nil {
			// Validation errors
			if verr := confluence.ValidatePageID(pageID); verr != nil {
				return outputError(httpclient.NewValidationError(verr.Error()))
			}
			// Other errors (API errors are already formatted)
			return outputError(&httpclient.ErrorResponse{
				Error:   httpclient.ErrTypeUnknown,
				Message: err.Error(),
			})
		}

		// Output page as JSON
		return page.Write(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(confluenceCmd)
	confluenceCmd.AddCommand(confluencePageCmd)
	confluencePageCmd.AddCommand(confluencePageGetCmd)
}
