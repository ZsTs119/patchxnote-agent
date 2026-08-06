package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"codeup.aliyun.com/689c25f21da8ac0447bef869/patchnote-agent/internal/config"
	"codeup.aliyun.com/689c25f21da8ac0447bef869/patchnote-agent/internal/version"
)

func TestRootCommandUsesSilentErrorBehavior(t *testing.T) {
	cmd := NewRootCommand()

	if !cmd.SilenceUsage {
		t.Fatal("expected root command to silence usage on errors")
	}
	if !cmd.SilenceErrors {
		t.Fatal("expected root command to silence cobra error printing")
	}
}

func TestRootHelpIncludesCommandAndGlobalFlags(t *testing.T) {
	stdout, stderr, err := executeForTest(t, "--help")
	if err != nil {
		t.Fatalf("expected help to succeed: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}

	for _, want := range []string{"auth", "login", "logout", "version", "--config", "--profile", "--output", "--server-base-url"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected help output to contain %q, got:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout, "completion") {
		t.Fatalf("expected help output to omit cobra default completion command, got:\n%s", stdout)
	}
}

func TestRootUsesDefaultServerBaseURL(t *testing.T) {
	var got string
	_, _, err := executeForTestWithDeps(t, Deps{
		CredentialStore: nil,
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			got = cfg.Server.BaseURL
			return &fakeAgentAPI{}, nil
		},
	}, "auth", "status")
	if err != nil {
		t.Fatalf("auth status: %v", err)
	}
	if got != config.DefaultServerBaseURL {
		t.Fatalf("expected default server base URL %q, got %q", config.DefaultServerBaseURL, got)
	}
}

func TestVersionPlainOutput(t *testing.T) {
	stdout, stderr, err := executeForTest(t, "--output", "plain", "version")
	if err != nil {
		t.Fatalf("expected version to succeed: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}

	for _, want := range []string{"patchnote " + version.Version, "commit " + version.Commit, "date " + version.Date, "platform "} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected plain version output to contain %q, got:\n%s", want, stdout)
		}
	}
}

func TestVersionJSONOutput(t *testing.T) {
	stdout, stderr, err := executeForTest(t, "--output", "json", "version")
	if err != nil {
		t.Fatalf("expected json version to succeed: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}

	var got version.BuildInfo
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("expected valid JSON version output: %v\n%s", err, stdout)
	}
	if got.Version != version.Version {
		t.Fatalf("expected version %q, got %q", version.Version, got.Version)
	}
	if got.Commit != version.Commit {
		t.Fatalf("expected commit %q, got %q", version.Commit, got.Commit)
	}
}

func TestVersionRejectsUnsupportedOutputWithoutUsageNoise(t *testing.T) {
	stdout, stderr, err := executeForTest(t, "--output", "xml", "version")
	if err == nil {
		t.Fatal("expected unsupported output to fail")
	}
	if stdout != "" {
		t.Fatalf("expected no stdout, got %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("expected no cobra stderr noise, got %q", stderr)
	}
	if strings.Contains(err.Error(), "Usage:") {
		t.Fatalf("expected error without usage text, got %v", err)
	}
}

func executeForTest(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	return executeForTestWithDeps(t, Deps{}, args...)
}

func executeForTestWithDeps(t *testing.T, deps Deps, args ...string) (string, string, error) {
	t.Helper()
	if deps.TargetOS == "" {
		deps.TargetOS = "linux"
	}
	if deps.PathEnv == (config.PathEnv{}) {
		deps.PathEnv = config.PathEnv{HomeDir: t.TempDir()}
	}

	cmd := NewRootCommandWithDeps(deps)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}
