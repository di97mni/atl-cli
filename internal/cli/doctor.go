package cli

import (
	"context"
	"encoding/json"
	"os"

	"github.com/martin/atl-cli/internal/config"
	"github.com/martin/atl-cli/internal/confluence"
	"github.com/martin/atl-cli/internal/jira"
	"github.com/spf13/cobra"
)

// DoctorResult represents the health check results.
type DoctorResult struct {
	Status string  `json:"status"` // "ok" or "error"
	Checks []Check `json:"checks"`
}

// Check represents a single health check result.
type Check struct {
	Name    string  `json:"name"`
	Status  string  `json:"status"` // "ok", "missing", or "error"
	Message *string `json:"message,omitempty"`
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate configuration and connectivity",
	Long:  "Checks environment variables and tests connectivity to Jira and Confluence APIs",
	RunE: func(cmd *cobra.Command, args []string) error {
		var checks []Check

		// Check environment variables
		envChecks := checkEnvVars()
		checks = append(checks, envChecks...)

		// Only run connectivity checks if all env vars are present
		allEnvPresent := true
		for _, check := range envChecks {
			if check.Status != "ok" {
				allEnvPresent = false
				break
			}
		}

		if allEnvPresent {
			cfg, _ := config.LoadFromEnv()

			// Check Jira connectivity
			jiraCheck := checkJiraConnectivity(cfg)
			checks = append(checks, jiraCheck)

			// Check Confluence connectivity
			confCheck := checkConfluenceConnectivity(cfg)
			checks = append(checks, confCheck)
		}

		// Build and output result
		result := buildResult(checks)
		return outputDoctorResult(result)
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func checkEnvVars() []Check {
	var checks []Check

	envVars := []string{config.EnvSite, config.EnvEmail, config.EnvToken}
	for _, envVar := range envVars {
		value := os.Getenv(envVar)
		if value == "" {
			msg := "Environment variable not set"
			checks = append(checks, Check{
				Name:    envVar,
				Status:  "missing",
				Message: &msg,
			})
		} else {
			checks = append(checks, Check{
				Name:   envVar,
				Status: "ok",
			})
		}
	}

	return checks
}

func checkJiraConnectivity(cfg *config.Config) Check {
	client := jira.NewClient(cfg, debug)
	err := client.CheckConnectivity(context.Background())

	if err != nil {
		msg := err.Error()
		return Check{
			Name:    "jira_connectivity",
			Status:  "error",
			Message: &msg,
		}
	}

	return Check{
		Name:   "jira_connectivity",
		Status: "ok",
	}
}

func checkConfluenceConnectivity(cfg *config.Config) Check {
	client := confluence.NewClient(cfg, debug)
	err := client.CheckConnectivity(context.Background())

	if err != nil {
		msg := err.Error()
		return Check{
			Name:    "confluence_connectivity",
			Status:  "error",
			Message: &msg,
		}
	}

	return Check{
		Name:   "confluence_connectivity",
		Status: "ok",
	}
}

func buildResult(checks []Check) *DoctorResult {
	status := "ok"
	for _, check := range checks {
		if check.Status != "ok" {
			status = "error"
			break
		}
	}

	return &DoctorResult{
		Status: status,
		Checks: checks,
	}
}

func outputDoctorResult(result *DoctorResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return err
	}

	// Return exit error if there were failures
	if result.Status != "ok" {
		return &exitError{code: 1}
	}

	return nil
}
