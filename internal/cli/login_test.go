package cli

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/auth"
	"github.com/ZsTs119/patchxnote-agent/internal/config"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
)

type fakeAgentAPI struct {
	requestPhone string
	verifyCode   string
	refreshCalls int
	currentUser  api.CurrentAccount
	logoutErr    error
	logoutCalled bool
}

func (f *fakeAgentAPI) RequestAgentOTP(ctx context.Context, request api.AgentOTPRequest, idempotencyKey string) (api.OTPRequestAccepted, error) {
	f.requestPhone = request.Phone
	if !strings.HasPrefix(request.ClientInstance, "agent_cli_") {
		return api.OTPRequestAccepted{}, errors.New("unexpected client instance")
	}
	if !strings.HasPrefix(idempotencyKey, "idem_") {
		return api.OTPRequestAccepted{}, errors.New("unexpected idempotency key")
	}
	return api.OTPRequestAccepted{
		RequestID:       "otp_request_fixture",
		Status:          "accepted",
		CooldownSeconds: 60,
	}, nil
}

func (f *fakeAgentAPI) VerifyAgentOTP(ctx context.Context, request api.AgentOTPVerificationRequest, idempotencyKey string) (api.AgentSessionResponse, error) {
	f.verifyCode = request.Code
	if request.RequestID != "otp_request_fixture" {
		return api.AgentSessionResponse{}, errors.New("unexpected request id")
	}
	if !strings.HasPrefix(idempotencyKey, "idem_") {
		return api.AgentSessionResponse{}, errors.New("unexpected verify idempotency key")
	}
	return api.AgentSessionResponse{
		AccessToken:             strings.Repeat("f", 32),
		AccessExpiresInSeconds:  3600,
		RefreshToken:            strings.Repeat("r", 43),
		RefreshExpiresInSeconds: 2592000,
		Account:                 currentFixtureAccount(),
		Scopes:                  []string{"agent:account.read", "agent:hardware.read"},
	}, nil
}

func (f *fakeAgentAPI) RefreshAgentSession(ctx context.Context, request api.AgentRefreshRequest, idempotencyKey string) (api.AgentSessionResponse, error) {
	f.refreshCalls++
	if request.RefreshToken == "" {
		return api.AgentSessionResponse{}, errors.New("missing refresh token")
	}
	if !strings.HasPrefix(idempotencyKey, "idem_") {
		return api.AgentSessionResponse{}, errors.New("unexpected refresh idempotency key")
	}
	return api.AgentSessionResponse{
		AccessToken:             strings.Repeat("n", 32),
		AccessExpiresInSeconds:  900,
		RefreshToken:            strings.Repeat("m", 43),
		RefreshExpiresInSeconds: 2592000,
		Account:                 currentFixtureAccount(),
		Scopes:                  []string{"agent:account.read", "agent:hardware.read"},
	}, nil
}

func (f *fakeAgentAPI) CurrentUser(ctx context.Context, accessToken string) (api.CurrentAccount, error) {
	if f.currentUser.ID != "" {
		return f.currentUser, nil
	}
	return currentFixtureAccount(), nil
}

func (f *fakeAgentAPI) ListRecorderCards(ctx context.Context, accessToken string) (api.AgentRecorderCardPage, error) {
	return api.AgentRecorderCardPage{}, nil
}

func (f *fakeAgentAPI) GetQuotaSummary(ctx context.Context, accessToken string) (api.AgentQuotaSummary, error) {
	return api.AgentQuotaSummary{}, nil
}

func (f *fakeAgentAPI) GetModelUsageSummary(ctx context.Context, accessToken string) (api.AgentModelUsageSummary, error) {
	return api.AgentModelUsageSummary{}, nil
}

func (f *fakeAgentAPI) ListMemories(ctx context.Context, accessToken string, params api.ListMemoriesParams) (api.AgentMemoryPage, error) {
	return api.AgentMemoryPage{}, nil
}

func (f *fakeAgentAPI) GetMemory(ctx context.Context, accessToken string, platform string, memoryID string) (api.AgentMemory, error) {
	return api.AgentMemory{}, nil
}

func (f *fakeAgentAPI) Logout(ctx context.Context, accessToken string) error {
	f.logoutCalled = true
	return f.logoutErr
}

func TestLoginWithFlagsStoresCredentialAndDoesNotLeakSensitiveInput(t *testing.T) {
	store := keychain.NewMemoryStore()
	fakeAPI := &fakeAgentAPI{}

	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: store,
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return fakeAPI, nil
		},
	}, "login", "--phone", "+86*******0000", "--code", "000000")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if !strings.Contains(stdout, "login succeeded") || !strings.Contains(stdout, "acct_fixture") {
		t.Fatalf("unexpected login stdout:\n%s", stdout)
	}
	combined := stdout + stderr
	for _, disallowed := range []string{
		"+86*******0000",
		"000000",
		strings.Repeat("f", 32),
		strings.Repeat("r", 43),
	} {
		if strings.Contains(combined, disallowed) {
			t.Fatalf("login output leaked sensitive value %q\nstdout:\n%s\nstderr:\n%s", disallowed, stdout, stderr)
		}
	}

	credential, err := store.Get(context.Background(), "default")
	if err != nil {
		t.Fatalf("stored credential: %v", err)
	}
	if credential.AccountID != "acct_fixture" || credential.AccessToken == "" || credential.RefreshToken == "" {
		t.Fatalf("unexpected stored credential account=%q has_access=%v has_refresh=%v",
			credential.AccountID, credential.AccessToken != "", credential.RefreshToken != "")
	}
	if credential.RefreshTokenExpiresAt.IsZero() {
		t.Fatal("expected refresh token expiry metadata to be stored")
	}
	if fakeAPI.requestPhone != "+86*******0000" || fakeAPI.verifyCode != "000000" {
		t.Fatalf("fake api did not receive expected inputs")
	}
}

func TestLoginInteractivePromptsOnStderr(t *testing.T) {
	store := keychain.NewMemoryStore()
	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: store,
		Stdin:           strings.NewReader("+86*******0000\n000000\n"),
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return &fakeAgentAPI{}, nil
		},
	}, "login")
	if err != nil {
		t.Fatalf("interactive login: %v", err)
	}
	if !strings.Contains(stdout, "login succeeded") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
	if !strings.Contains(stderr, "Phone:") || !strings.Contains(stderr, "Verification code:") {
		t.Fatalf("expected prompts on stderr, got:\n%s", stderr)
	}
	if strings.Contains(stdout+stderr, "000000") {
		t.Fatal("interactive login output leaked verification code")
	}
}

func TestAuthStatusJSONUsesCurrentUserProjection(t *testing.T) {
	store := keychain.NewMemoryStore()
	if err := store.Put(context.Background(), "default", keychain.Credential{
		AccountID:    "acct_cached",
		AccessToken:  strings.Repeat("s", 32),
		RefreshToken: strings.Repeat("t", 43),
	}); err != nil {
		t.Fatalf("seed credential: %v", err)
	}

	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: store,
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return &fakeAgentAPI{currentUser: currentFixtureAccount()}, nil
		},
	}, "--output", "json", "auth", "status")
	if err != nil {
		t.Fatalf("auth status: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}

	var got auth.Status
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode auth status json: %v\n%s", err, stdout)
	}
	if !got.Authenticated || got.AccountID != "acct_fixture" || got.PhoneMasked != "+86*******0000" {
		t.Fatalf("unexpected auth status: %+v", got)
	}
}

func TestLogoutRemovesLocalCredentialWhenServerLogoutFails(t *testing.T) {
	store := keychain.NewMemoryStore()
	if err := store.Put(context.Background(), "default", keychain.Credential{
		AccountID:    "acct_fixture",
		AccessToken:  strings.Repeat("u", 32),
		RefreshToken: strings.Repeat("v", 43),
	}); err != nil {
		t.Fatalf("seed credential: %v", err)
	}
	fakeAPI := &fakeAgentAPI{logoutErr: errors.New("temporary revoke failure")}

	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: store,
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return fakeAPI, nil
		},
	}, "logout")
	if err != nil {
		t.Fatalf("logout: %v", err)
	}
	if !fakeAPI.logoutCalled {
		t.Fatal("expected server logout to be attempted")
	}
	if !strings.Contains(stderr, "warning: server logout failed") {
		t.Fatalf("expected warning on stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "logged out") {
		t.Fatalf("expected local logout stdout, got %s", stdout)
	}
	if _, err := store.Get(context.Background(), "default"); !errors.Is(err, keychain.ErrNotFound) {
		t.Fatalf("expected local credential deletion, got %v", err)
	}
}

func currentFixtureAccount() api.CurrentAccount {
	return api.CurrentAccount{
		ID:                   "acct_fixture",
		Status:               "active",
		RegistrationPlatform: "mobile",
		PhoneMasked:          "+86*******0000",
		StateVersion:         2,
	}
}
