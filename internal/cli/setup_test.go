package cli

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/config"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
	"github.com/ZsTs119/patchxnote-agent/internal/oauthflow"
)

func TestSetupDryRunPrintsPlanWithoutAuthOrWrites(t *testing.T) {
	home := t.TempDir()
	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: keychain.NewMemoryStore(),
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return &fakeAgentAPI{}, nil
		},
		PathEnv:  config.PathEnv{HomeDir: home},
		TargetOS: "linux",
	}, "setup", "--client", "cursor", "--dry-run", "--print-config")
	if err != nil {
		t.Fatalf("setup dry-run: %v", err)
	}
	for _, want := range []string{"Dry run: Cursor", ".cursor/mcp.json", "patchxnote-agent@latest", "No files changed"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected stdout to contain %q:\n%s", want, stdout)
		}
	}
	if strings.Contains(stdout+stderr, "access_token") || strings.Contains(stdout+stderr, "refresh_token") {
		t.Fatalf("setup output leaked secret-like fields:\nstdout=%s\nstderr=%s", stdout, stderr)
	}
}

func TestSetupJSONDryRunWritesOnlyJSONToStdout(t *testing.T) {
	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: keychain.NewMemoryStore(),
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return &fakeAgentAPI{}, nil
		},
	}, "--output", "json", "setup", "--client", "cursor", "--dry-run")
	if err != nil {
		t.Fatalf("setup json dry-run: %v", err)
	}
	if stderr != "" && strings.Contains(stderr, "{") {
		t.Fatalf("expected no JSON-looking stderr noise: %s", stderr)
	}
	var decoded setupCommandResult
	if err := json.Unmarshal([]byte(stdout), &decoded); err != nil {
		t.Fatalf("decode setup json: %v\n%s", err, stdout)
	}
	if decoded.AuthStatus != "not_checked" || len(decoded.Clients) != 1 || decoded.Clients[0].Status != "dry_run" {
		t.Fatalf("unexpected setup result: %+v", decoded)
	}
}

func TestSetupAllLocalSupportedDryRun(t *testing.T) {
	stdout, _, err := executeForTestWithDeps(t, Deps{
		CredentialStore: keychain.NewMemoryStore(),
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return &fakeAgentAPI{}, nil
		},
	}, "setup", "--all-local-supported", "--dry-run")
	if err != nil {
		t.Fatalf("setup all local dry-run: %v", err)
	}
	for _, want := range []string{"codex", "cursor", "claude-desktop", "vscode", "windsurf"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected all-local dry-run to include %s:\n%s", want, stdout)
		}
	}
}

func TestSetupManualClientUsesMCPOAuthAndDoesNotLeak(t *testing.T) {
	store := keychain.NewMemoryStore()
	serverBaseURL := "http://127.0.0.1:18084/patchnote-test-api"
	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: store,
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return &fakeAgentAPI{oauthMeta: oauthMetadataForBase(serverBaseURL)}, nil
		},
		BrowserOpen: browserCallbackOpener(t, "setup_code"),
	}, "--server-base-url", serverBaseURL, "setup", "--client", "workbuddy", "--skip-mcp-smoke")
	if err != nil {
		t.Fatalf("setup manual mcp oauth: %v", err)
	}
	if !strings.Contains(stdout, "Manual setup required") || !strings.Contains(stdout, "workbuddy") {
		t.Fatalf("unexpected stdout:\n%s", stdout)
	}
	for _, disallowed := range []string{"setup_code", strings.Repeat("a", 32), strings.Repeat("b", 43)} {
		if strings.Contains(stdout+stderr, disallowed) {
			t.Fatalf("setup output leaked %q\nstdout=%s\nstderr=%s", disallowed, stdout, stderr)
		}
	}
	if _, err := store.Get(context.Background(), "default"); !keychain.IsNotFound(err) {
		t.Fatalf("legacy Agent credential should be untouched, got %v", err)
	}
	if _, ok, err := oauthflow.NewStore(store, "default").Load(context.Background()); err != nil || !ok {
		t.Fatalf("expected stored MCP OAuth credential, ok=%v err=%v", ok, err)
	}
}

func TestSetupAutoWritesCursorConfigWithExistingMCPOAuth(t *testing.T) {
	home := t.TempDir()
	store := keychain.NewMemoryStore()
	if err := oauthflow.NewStore(store, "default").Save(context.Background(), mcpCredentialForTest(config.DefaultServerBaseURL, "setup_access", "setup_refresh", time.Now().UTC().Add(time.Hour))); err != nil {
		t.Fatalf("seed credential: %v", err)
	}

	stdout, _, err := executeForTestWithDeps(t, Deps{
		CredentialStore: store,
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return &fakeAgentAPI{}, nil
		},
		PathEnv:  config.PathEnv{HomeDir: home},
		TargetOS: "linux",
	}, "setup", "--client", "cursor", "--yes", "--skip-mcp-smoke")
	if err != nil {
		t.Fatalf("setup cursor: %v", err)
	}
	if !strings.Contains(stdout, "Installed PatchXNote MCP config for cursor") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
	stdout, _, err = executeForTestWithDeps(t, Deps{
		CredentialStore: store,
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return &fakeAgentAPI{}, nil
		},
		PathEnv:  config.PathEnv{HomeDir: home},
		TargetOS: "linux",
	}, "setup", "--client", "cursor", "--yes", "--skip-mcp-smoke")
	if err != nil {
		t.Fatalf("setup cursor second run: %v", err)
	}
	if !strings.Contains(stdout, "already exists") {
		t.Fatalf("expected idempotent second run: %s", stdout)
	}
}

func TestSetupPlatformOnlyClientSkipsLocalOAuth(t *testing.T) {
	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: keychain.NewMemoryStore(),
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return &fakeAgentAPI{}, nil
		},
		BrowserOpen: func(rawURL string) error {
			t.Fatalf("browser opener should not be called for platform-only setup")
			return nil
		},
	}, "--output", "json", "setup", "--client", "feishu-aily", "--skip-mcp-smoke")
	if err != nil {
		t.Fatalf("setup platform-only: %v", err)
	}
	if !strings.Contains(stderr, "platform-side") {
		t.Fatalf("expected platform-side guidance, got stderr=%q", stderr)
	}
	var decoded setupCommandResult
	if err := json.Unmarshal([]byte(stdout), &decoded); err != nil {
		t.Fatalf("decode setup json: %v\n%s", err, stdout)
	}
	if decoded.AuthStatus != "not_required" || decoded.AuthMethod != "remote_platform_pending" || len(decoded.Clients) != 1 || !decoded.Clients[0].ManualRequired {
		t.Fatalf("unexpected setup result: %+v", decoded)
	}
}

func TestPollSetupSessionHandlesApprovedAndDenied(t *testing.T) {
	approvedClient := &fakeSetupSessionClient{statuses: []api.AgentSetupSessionStatus{
		{
			SessionID: "setup_fixture",
			Status:    "approved",
			Session: &api.AgentSessionResponse{
				AccessToken:             strings.Repeat("a", 32),
				AccessExpiresInSeconds:  3600,
				RefreshToken:            strings.Repeat("r", 43),
				RefreshExpiresInSeconds: 2592000,
				Account:                 currentFixtureAccount(),
				Scopes:                  []string{"agent:account.read"},
			},
		},
	}}
	session, err := pollSetupSession(context.Background(), approvedClient, api.AgentSetupSessionCreated{SessionID: "setup_fixture", ExpiresInSeconds: 5, PollIntervalSeconds: 1})
	if err != nil {
		t.Fatalf("poll approved: %v", err)
	}
	if session == nil || session.Account.ID != "acct_fixture" {
		t.Fatalf("unexpected session: %+v", session)
	}

	deniedClient := &fakeSetupSessionClient{statuses: []api.AgentSetupSessionStatus{{SessionID: "setup_fixture", Status: "denied"}}}
	if _, err := pollSetupSession(context.Background(), deniedClient, api.AgentSetupSessionCreated{SessionID: "setup_fixture", ExpiresInSeconds: 5, PollIntervalSeconds: 1}); err == nil {
		t.Fatal("expected denied setup session error")
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	pendingClient := &fakeSetupSessionClient{statuses: []api.AgentSetupSessionStatus{{SessionID: "setup_fixture", Status: "pending"}}}
	if _, err := pollSetupSession(timeoutCtx, pendingClient, api.AgentSetupSessionCreated{SessionID: "setup_fixture", ExpiresInSeconds: 5, PollIntervalSeconds: 1}); err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout, got %v", err)
	}
}

type fakeSetupSessionClient struct {
	statuses []api.AgentSetupSessionStatus
}

func (f *fakeSetupSessionClient) CreateAgentSetupSession(ctx context.Context, request api.AgentSetupSessionCreateRequest, idempotencyKey string) (api.AgentSetupSessionCreated, error) {
	return api.AgentSetupSessionCreated{}, nil
}

func (f *fakeSetupSessionClient) GetAgentSetupSession(ctx context.Context, sessionID string) (api.AgentSetupSessionStatus, error) {
	if len(f.statuses) == 0 {
		return api.AgentSetupSessionStatus{SessionID: sessionID, Status: "pending"}, nil
	}
	status := f.statuses[0]
	f.statuses = f.statuses[1:]
	return status, nil
}
