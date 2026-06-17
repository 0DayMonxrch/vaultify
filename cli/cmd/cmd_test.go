package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0DayMonxrch/vaultify/cli/config"
)

func TestCLILoginLogout(t *testing.T) {
	// Setup temporary directory for config
	tempDir, err := os.MkdirTemp("", "vaultify-cmd-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	config.SetConfigPathOverride(filepath.Join(tempDir, ".vaultify", "config"))

	// Test Login
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)

	// Reset global flag variables
	host = ""
	token = ""

	RootCmd.SetArgs([]string{"login", "--host", "http://localhost:8080", "--token", "vt_1234567890"})
	err = RootCmd.Execute()
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if !strings.Contains(buf.String(), "Successfully logged in.") {
		t.Errorf("expected success message, got: %q", buf.String())
	}

	// Verify config file was written and has correct content
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}
	if cfg.Host != "http://localhost:8080" || cfg.Token != "vt_1234567890" {
		t.Errorf("incorrect saved config: %+v", cfg)
	}

	// Test Logout
	buf.Reset()
	RootCmd.SetArgs([]string{"logout"})
	err = RootCmd.Execute()
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	if !strings.Contains(buf.String(), "Successfully logged out.") {
		t.Errorf("expected success message, got: %q", buf.String())
	}

	// Verify config is gone
	_, err = config.LoadConfig()
	if err == nil {
		t.Error("expected config to be deleted")
	}
}

func TestCLILoginValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "vaultify-cmd-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	config.SetConfigPathOverride(filepath.Join(tempDir, ".vaultify", "config"))

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)

	// Reset global flag variables
	host = ""
	token = ""

	RootCmd.SetArgs([]string{"login"}) // missing host and token
	err = RootCmd.Execute()
	if err == nil {
		t.Fatal("expected login validation error")
	}

	if !strings.Contains(err.Error(), "--token and --host are required") {
		t.Errorf("expected validation message, got: %v", err)
	}
}

func TestCLIReadCommands(t *testing.T) {
	// Start local mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		if r.Header.Get("X-Vaultify-Token") != "vt_mock_token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.URL.Path == "/projects" || r.URL.Path == "/api/v1/projects" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[
				{
					"ID": "11111111-1111-1111-1111-111111111111",
					"Name": "Project One",
					"Slug": "project-one",
					"CreatedAt": "2026-06-16T12:00:00Z"
				}
			]`))
			return
		}

		if r.URL.Path == "/projects/11111111-1111-1111-1111-111111111111/secrets" || r.URL.Path == "/api/v1/projects/11111111-1111-1111-1111-111111111111/secrets" {
			env := r.URL.Query().Get("env")
			if env != "" && env != "dev" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[]`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[
				{
					"ID": "22222222-2222-2222-2222-222222222222",
					"KeyName": "DATABASE_URL",
					"Environment": "dev",
					"CreatedAt": "2026-06-16T12:05:00Z",
					"UpdatedAt": "2026-06-16T12:10:00Z"
				}
			]`))
			return
		}

		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer server.Close()

	// Setup temporary directory for config
	tempDir, err := os.MkdirTemp("", "vaultify-cmd-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	config.SetConfigPathOverride(filepath.Join(tempDir, ".vaultify", "config"))

	// Save config pointing to mock server
	cfg := &config.Config{
		Host:  server.URL,
		Token: "vt_mock_token",
	}
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// 1. Test "projects list"
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)

	RootCmd.SetArgs([]string{"projects", "list"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("projects list failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Project One") || !strings.Contains(out, "project-one") {
		t.Errorf("expected project details in output, got: %q", out)
	}

	// 2. Test "secrets list --project project-one --env dev"
	buf.Reset()
	projectSlug = ""
	secretEnv = ""
	RootCmd.SetArgs([]string{"secrets", "list", "--project", "project-one", "--env", "dev"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("secrets list failed: %v", err)
	}

	out = buf.String()
	if !strings.Contains(out, "DATABASE_URL") || !strings.Contains(out, "dev") {
		t.Errorf("expected secret details in output, got: %q", out)
	}

	// 3. Test "secrets list --project project-one --env prod" (no secrets found)
	buf.Reset()
	projectSlug = ""
	secretEnv = ""
	RootCmd.SetArgs([]string{"secrets", "list", "--project", "project-one", "--env", "prod"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("secrets list failed: %v", err)
	}

	out = buf.String()
	if !strings.Contains(out, "No secrets found.") {
		t.Errorf("expected 'No secrets found.' message, got: %q", out)
	}

	// 4. Test "secrets list --project unknown-project" (not found error)
	buf.Reset()
	projectSlug = ""
	secretEnv = ""
	RootCmd.SetArgs([]string{"secrets", "list", "--project", "unknown-project"})
	err = RootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown project slug")
	}
	if !strings.Contains(err.Error(), `project with slug "unknown-project" not found`) {
		t.Errorf("expected slug not found error, got: %v", err)
	}
}
