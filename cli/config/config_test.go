package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLifecycle(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "vaultify-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	SetConfigPathOverride(filepath.Join(tempDir, ".vaultify", "config"))

	// Initially not logged in
	_, err = LoadConfig()
	if err == nil {
		t.Fatalf("expected error loading missing config")
	}

	// Save config
	cfg := &Config{Host: "http://localhost:8080", Token: "vt_test123"}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify permissions
	path, _ := GetConfigPath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}

	// Windows doesn't map POSIX strictly, but checking mode on non-Windows is good.
	if info.Mode().Perm() != 0600 && os.PathSeparator == '/' {
		t.Errorf("expected 0600 permissions, got %o", info.Mode().Perm())
	}

	// Load config
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if loaded.Host != cfg.Host || loaded.Token != cfg.Token {
		t.Fatalf("config mismatch: got %v, want %v", loaded, cfg)
	}

	// Delete config
	if err := DeleteConfig(); err != nil {
		t.Fatalf("failed to delete config: %v", err)
	}

	// Load should fail again
	_, err = LoadConfig()
	if err == nil {
		t.Fatalf("expected error loading deleted config")
	}
}
