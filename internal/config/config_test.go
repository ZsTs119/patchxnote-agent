package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
)

func TestLoadUsesDefaultsWhenConfigIsMissing(t *testing.T) {
	cfg, err := Load(NewViper(), LoadOptions{
		GOOS:    "linux",
		PathEnv: PathEnv{HomeDir: t.TempDir()},
	})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Profile != "default" {
		t.Fatalf("expected default profile, got %q", cfg.Profile)
	}
	if cfg.Output != "plain" {
		t.Fatalf("expected default output, got %q", cfg.Output)
	}
	if cfg.Server.BaseURL != "" {
		t.Fatalf("expected empty server base URL, got %q", cfg.Server.BaseURL)
	}
}

func TestLoadPrecedenceFlagEnvFileDefault(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("profile: file-profile\noutput: file-output\nserver:\n  base_url: https://file.example\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("PATCHNOTE_PROFILE", "env-profile")
	t.Setenv("PATCHNOTE_SERVER_BASE_URL", "https://env.example")

	v := NewViper()
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.String("output", "", "")
	if err := flags.Set("output", "flag-output"); err != nil {
		t.Fatalf("set output flag: %v", err)
	}
	if err := v.BindPFlag("output", flags.Lookup("output")); err != nil {
		t.Fatalf("bind output flag: %v", err)
	}

	cfg, err := Load(v, LoadOptions{
		ConfigFile: configFile,
		GOOS:       "linux",
		PathEnv:    PathEnv{HomeDir: dir},
	})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Profile != "env-profile" {
		t.Fatalf("expected env profile, got %q", cfg.Profile)
	}
	if cfg.Output != "flag-output" {
		t.Fatalf("expected flag output, got %q", cfg.Output)
	}
	if cfg.Server.BaseURL != "https://env.example" {
		t.Fatalf("expected env server base URL, got %q", cfg.Server.BaseURL)
	}
}

func TestLoadReturnsErrorForExplicitMissingConfig(t *testing.T) {
	_, err := Load(NewViper(), LoadOptions{
		ConfigFile: filepath.Join(t.TempDir(), "missing.yaml"),
		GOOS:       "linux",
		PathEnv:    PathEnv{HomeDir: t.TempDir()},
	})
	if err == nil {
		t.Fatal("expected explicit missing config to fail")
	}
}
