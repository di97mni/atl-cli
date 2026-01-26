package cli

import (
	"context"
	"os"

	"github.com/martin/atl-cli/internal/config"
	"github.com/martin/atl-cli/internal/httpclient"
	"github.com/martin/atl-cli/internal/jira"
	"github.com/spf13/cobra"
)

var jiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Jira commands",
	Long:  "Commands for interacting with Jira",
}

var jiraIssueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Jira issue commands",
	Long:  "Commands for working with Jira issues",
}

var jiraIssueGetCmd = &cobra.Command{
	Use:   "get <issue-key>",
	Short: "Get a Jira issue by key",
	Long:  "Retrieves details of a Jira issue and outputs as JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueKey := args[0]

		// Load and validate config
		cfg, err := config.LoadFromEnv()
		if err != nil {
			return outputError(httpclient.NewConfigError(err.Error()))
		}
		if err := cfg.Validate(); err != nil {
			return outputError(httpclient.NewConfigError(err.Error()))
		}

		// Create client with debug flag from root command
		client := jira.NewClient(cfg, debug)

		// Get issue
		issue, err := client.GetIssue(context.Background(), issueKey)
		if err != nil {
			// Validation errors
			if verr := jira.ValidateIssueKey(issueKey); verr != nil {
				return outputError(httpclient.NewValidationError(verr.Error()))
			}
			// Other errors (API errors are already formatted)
			return outputError(&httpclient.ErrorResponse{
				Error:   httpclient.ErrTypeUnknown,
				Message: err.Error(),
			})
		}

		// Output issue as JSON
		return issue.Write(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(jiraCmd)
	jiraCmd.AddCommand(jiraIssueCmd)
	jiraIssueCmd.AddCommand(jiraIssueGetCmd)
}

// outputError writes an error response to stderr and returns an error to signal non-zero exit
func outputError(errResp *httpclient.ErrorResponse) error {
	errResp.Write(os.Stderr)
	return &exitError{code: 1}
}

// exitError is used to signal a non-zero exit code
type exitError struct {
	code int
}

func (e *exitError) Error() string {
	return ""
}
