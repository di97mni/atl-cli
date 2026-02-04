package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

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

// Flags for jira issue create
var (
	createProject     string
	createType        string
	createSummary     string
	createDescription string
	createParent      string
	createLabels      string
	createTemplate    string
	createVars        []string
)

var jiraIssueCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Jira issue",
	Long:  "Creates a new Jira issue (story, subtask, task, or bug)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load and validate config
		cfg, err := config.LoadFromEnv()
		if err != nil {
			return outputError(httpclient.NewConfigError(err.Error()))
		}
		if err := cfg.Validate(); err != nil {
			return outputError(httpclient.NewConfigError(err.Error()))
		}

		var project, issueType, summary, description string
		var labels []string

		// If template is provided, load and process it
		if createTemplate != "" {
			tmpl, err := jira.LoadTemplate(createTemplate)
			if err != nil {
				return outputError(httpclient.NewValidationError(err.Error()))
			}

			// Parse variable flags
			vars, err := jira.ParseVarFlags(createVars)
			if err != nil {
				return outputError(httpclient.NewValidationError(err.Error()))
			}

			// Apply template
			parsed, err := tmpl.Apply(vars)
			if err != nil {
				return outputError(httpclient.NewValidationError(err.Error()))
			}

			project = parsed.Project
			issueType = parsed.IssueType
			summary = parsed.Summary
			description = parsed.Description
			labels = parsed.Labels
		}

		// Command-line flags override template values
		if createProject != "" {
			project = createProject
		}
		if createType != "" {
			issueType = createType
		}
		if createSummary != "" {
			summary = createSummary
		}
		if createDescription != "" {
			description = createDescription
		}
		if createLabels != "" {
			labels = strings.Split(createLabels, ",")
			for i := range labels {
				labels[i] = strings.TrimSpace(labels[i])
			}
		}

		// Validate required fields
		if project == "" {
			return outputError(httpclient.NewValidationError("--project is required"))
		}
		if issueType == "" {
			return outputError(httpclient.NewValidationError("--type is required"))
		}
		if summary == "" {
			return outputError(httpclient.NewValidationError("--summary is required"))
		}

		// Validate project key format
		if err := jira.ValidateProjectKey(project); err != nil {
			return outputError(httpclient.NewValidationError(err.Error()))
		}

		// Normalize issue type
		normalizedType := jira.NormalizeIssueType(issueType)
		if normalizedType == "" {
			return outputError(httpclient.NewValidationError(
				fmt.Sprintf("invalid issue type %q (valid types: %s)",
					issueType, strings.Join(jira.ValidIssueTypes(), ", "))))
		}

		// Validate parent for subtask
		if strings.ToLower(issueType) == "subtask" {
			if createParent == "" {
				return outputError(httpclient.NewValidationError("--parent is required for subtask"))
			}
			if err := jira.ValidateIssueKey(createParent); err != nil {
				return outputError(httpclient.NewValidationError(err.Error()))
			}
		}

		// Build request
		req := &jira.CreateIssueRequest{
			Fields: jira.CreateIssueFields{
				Project:   jira.ProjectRef{Key: project},
				IssueType: jira.IssueType{Name: normalizedType},
				Summary:   summary,
			},
		}

		if description != "" {
			req.Fields.Description = jira.TextToADF(description)
		}

		if createParent != "" {
			req.Fields.Parent = &jira.ParentRef{Key: createParent}
		}

		if len(labels) > 0 {
			req.Fields.Labels = labels
		}

		// Create client and issue
		client := jira.NewClient(cfg, debug)
		created, err := client.CreateIssue(context.Background(), req)
		if err != nil {
			return outputError(&httpclient.ErrorResponse{
				Error:   httpclient.ErrTypeUnknown,
				Message: err.Error(),
			})
		}

		// Output created issue
		return created.Write(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(jiraCmd)
	jiraCmd.AddCommand(jiraIssueCmd)
	jiraIssueCmd.AddCommand(jiraIssueGetCmd)
	jiraIssueCmd.AddCommand(jiraIssueCreateCmd)

	// Register create command flags
	jiraIssueCreateCmd.Flags().StringVar(&createProject, "project", "", "Project key (e.g., CST)")
	jiraIssueCreateCmd.Flags().StringVar(&createType, "type", "", "Issue type: story, subtask, task, bug")
	jiraIssueCreateCmd.Flags().StringVar(&createSummary, "summary", "", "Issue summary")
	jiraIssueCreateCmd.Flags().StringVar(&createDescription, "description", "", "Issue description")
	jiraIssueCreateCmd.Flags().StringVar(&createParent, "parent", "", "Parent issue key (required for subtask)")
	jiraIssueCreateCmd.Flags().StringVar(&createLabels, "labels", "", "Comma-separated labels")
	jiraIssueCreateCmd.Flags().StringVar(&createTemplate, "template", "", "Path to template file")
	jiraIssueCreateCmd.Flags().StringArrayVar(&createVars, "var", nil, "Template variable (key=value), repeatable")
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
