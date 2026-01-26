package cli

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/martin/atl-cli/internal/config"
)

func TestDoctorResult_JSON(t *testing.T) {
	msg := "Missing value"
	result := &DoctorResult{
		Status: "error",
		Checks: []Check{
			{Name: "ATL_CLI_SITE", Status: "ok", Message: nil},
			{Name: "ATL_CLI_EMAIL", Status: "missing", Message: &msg},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal result: %v", err)
	}

	var parsed DoctorResult
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if parsed.Status != result.Status {
		t.Errorf("expected status %s, got %s", result.Status, parsed.Status)
	}
	if len(parsed.Checks) != len(result.Checks) {
		t.Errorf("expected %d checks, got %d", len(result.Checks), len(parsed.Checks))
	}
}

func TestCheckEnvVars_AllPresent(t *testing.T) {
	// Save original env
	origSite := os.Getenv(config.EnvSite)
	origEmail := os.Getenv(config.EnvEmail)
	origToken := os.Getenv(config.EnvToken)
	defer func() {
		os.Setenv(config.EnvSite, origSite)
		os.Setenv(config.EnvEmail, origEmail)
		os.Setenv(config.EnvToken, origToken)
	}()

	// Set all env vars
	os.Setenv(config.EnvSite, "test.atlassian.net")
	os.Setenv(config.EnvEmail, "test@example.com")
	os.Setenv(config.EnvToken, "test-token")

	checks := checkEnvVars()
	if len(checks) != 3 {
		t.Errorf("expected 3 checks, got %d", len(checks))
	}

	for _, check := range checks {
		if check.Status != "ok" {
			t.Errorf("expected check %s to be ok, got %s", check.Name, check.Status)
		}
	}
}

func TestCheckEnvVars_SomeMissing(t *testing.T) {
	// Save original env
	origSite := os.Getenv(config.EnvSite)
	origEmail := os.Getenv(config.EnvEmail)
	origToken := os.Getenv(config.EnvToken)
	defer func() {
		os.Setenv(config.EnvSite, origSite)
		os.Setenv(config.EnvEmail, origEmail)
		os.Setenv(config.EnvToken, origToken)
	}()

	// Set only one env var
	os.Setenv(config.EnvSite, "test.atlassian.net")
	os.Unsetenv(config.EnvEmail)
	os.Unsetenv(config.EnvToken)

	checks := checkEnvVars()

	okCount := 0
	missingCount := 0
	for _, check := range checks {
		if check.Status == "ok" {
			okCount++
		} else if check.Status == "missing" {
			missingCount++
		}
	}

	if okCount != 1 {
		t.Errorf("expected 1 ok check, got %d", okCount)
	}
	if missingCount != 2 {
		t.Errorf("expected 2 missing checks, got %d", missingCount)
	}
}

func TestCheckEnvVars_AllMissing(t *testing.T) {
	// Save original env
	origSite := os.Getenv(config.EnvSite)
	origEmail := os.Getenv(config.EnvEmail)
	origToken := os.Getenv(config.EnvToken)
	defer func() {
		os.Setenv(config.EnvSite, origSite)
		os.Setenv(config.EnvEmail, origEmail)
		os.Setenv(config.EnvToken, origToken)
	}()

	// Clear all env vars
	os.Unsetenv(config.EnvSite)
	os.Unsetenv(config.EnvEmail)
	os.Unsetenv(config.EnvToken)

	checks := checkEnvVars()

	for _, check := range checks {
		if check.Status != "missing" {
			t.Errorf("expected check %s to be missing, got %s", check.Name, check.Status)
		}
		if check.Message == nil {
			t.Errorf("expected check %s to have message", check.Name)
		}
	}
}

func TestBuildResult_AllOk(t *testing.T) {
	checks := []Check{
		{Name: "check1", Status: "ok"},
		{Name: "check2", Status: "ok"},
	}

	result := buildResult(checks)
	if result.Status != "ok" {
		t.Errorf("expected status ok, got %s", result.Status)
	}
}

func TestBuildResult_HasErrors(t *testing.T) {
	msg := "Error message"
	checks := []Check{
		{Name: "check1", Status: "ok"},
		{Name: "check2", Status: "error", Message: &msg},
	}

	result := buildResult(checks)
	if result.Status != "error" {
		t.Errorf("expected status error, got %s", result.Status)
	}
}

func TestBuildResult_HasMissing(t *testing.T) {
	msg := "Missing"
	checks := []Check{
		{Name: "check1", Status: "ok"},
		{Name: "check2", Status: "missing", Message: &msg},
	}

	result := buildResult(checks)
	if result.Status != "error" {
		t.Errorf("expected status error, got %s", result.Status)
	}
}
