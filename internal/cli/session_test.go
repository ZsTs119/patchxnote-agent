package cli

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/auth"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
)

func TestSessionCredentialProviderRefreshesExpiredAccess(t *testing.T) {
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	store := keychain.NewMemoryStore()
	manager := auth.NewManager(store, "default")
	if err := manager.Save(context.Background(), keychain.Credential{
		AccountID:             "acct_cached",
		AccessToken:           strings.Repeat("o", 32),
		RefreshToken:          strings.Repeat("r", 43),
		AccessTokenExpiresAt:  now.Add(-time.Minute),
		RefreshTokenExpiresAt: now.Add(30 * 24 * time.Hour),
		Scopes:                []string{"agent:account.read"},
	}); err != nil {
		t.Fatalf("seed credential: %v", err)
	}
	fakeAPI := &fakeAgentAPI{}
	provider := newTestSessionCredentialProvider(manager, fakeAPI, now, t.TempDir())

	credential, ok, err := provider.Credential(context.Background())
	if err != nil {
		t.Fatalf("credential: %v", err)
	}
	if !ok || credential.AccessToken != strings.Repeat("n", 32) || credential.RefreshToken != strings.Repeat("m", 43) {
		t.Fatalf("unexpected refreshed credential ok=%v has_new_access=%v has_new_refresh=%v",
			ok, credential.AccessToken == strings.Repeat("n", 32), credential.RefreshToken == strings.Repeat("m", 43))
	}
	if fakeAPI.refreshCalls != 1 {
		t.Fatalf("refresh calls=%d", fakeAPI.refreshCalls)
	}
	stored, err := store.Get(context.Background(), "default")
	if err != nil {
		t.Fatalf("stored credential: %v", err)
	}
	if stored.AccessToken != credential.AccessToken || stored.RefreshToken != credential.RefreshToken || stored.RefreshTokenExpiresAt.IsZero() {
		t.Fatal("expected rotated credentials to be stored")
	}
}

func TestSessionCredentialProviderSkipsFreshAccess(t *testing.T) {
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	store := keychain.NewMemoryStore()
	manager := auth.NewManager(store, "default")
	if err := manager.Save(context.Background(), keychain.Credential{
		AccountID:            "acct_cached",
		AccessToken:          strings.Repeat("o", 32),
		RefreshToken:         strings.Repeat("r", 43),
		AccessTokenExpiresAt: now.Add(time.Hour),
		Scopes:               []string{"agent:account.read"},
	}); err != nil {
		t.Fatalf("seed credential: %v", err)
	}
	fakeAPI := &fakeAgentAPI{}
	provider := newTestSessionCredentialProvider(manager, fakeAPI, now, t.TempDir())

	credential, ok, err := provider.Credential(context.Background())
	if err != nil {
		t.Fatalf("credential: %v", err)
	}
	if !ok || credential.AccessToken != strings.Repeat("o", 32) || fakeAPI.refreshCalls != 0 {
		t.Fatalf("unexpected credential ok=%v refresh_calls=%d", ok, fakeAPI.refreshCalls)
	}
}

func TestSessionCredentialProviderRequiresLoginWithoutRefresh(t *testing.T) {
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	store := keychain.NewMemoryStore()
	manager := auth.NewManager(store, "default")
	if err := manager.Save(context.Background(), keychain.Credential{
		AccountID:            "acct_cached",
		AccessToken:          strings.Repeat("o", 32),
		AccessTokenExpiresAt: now.Add(-time.Minute),
		Scopes:               []string{"agent:account.read"},
	}); err != nil {
		t.Fatalf("seed credential: %v", err)
	}
	provider := newTestSessionCredentialProvider(manager, &fakeAgentAPI{}, now, t.TempDir())

	_, ok, err := provider.Credential(context.Background())
	if err != nil {
		t.Fatalf("credential: %v", err)
	}
	if ok {
		t.Fatal("expected expired access without refresh token to require login")
	}
}

func TestSessionCredentialProviderClearsCredentialOnUnauthorizedRefresh(t *testing.T) {
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	store := keychain.NewMemoryStore()
	manager := auth.NewManager(store, "default")
	if err := manager.Save(context.Background(), keychain.Credential{
		AccountID:            "acct_cached",
		AccessToken:          strings.Repeat("o", 32),
		RefreshToken:         strings.Repeat("r", 43),
		AccessTokenExpiresAt: now.Add(-time.Minute),
		Scopes:               []string{"agent:account.read"},
	}); err != nil {
		t.Fatalf("seed credential: %v", err)
	}
	provider := newTestSessionCredentialProvider(manager, &refreshErrorAPI{
		err: &api.Error{StatusCode: http.StatusUnauthorized, Code: "auth_required"},
	}, now, t.TempDir())

	_, ok, err := provider.Credential(context.Background())
	if err == nil || ok {
		t.Fatalf("expected unauthorized refresh error, ok=%v err=%v", ok, err)
	}
	if _, err := store.Get(context.Background(), "default"); !errors.Is(err, keychain.ErrNotFound) {
		t.Fatalf("expected credential to be cleared, got %v", err)
	}
}

func newTestSessionCredentialProvider(manager *auth.Manager, agentClient agentAPI, now time.Time, dir string) *sessionCredentialProvider {
	return &sessionCredentialProvider{
		auth:          manager,
		api:           agentClient,
		lockPath:      filepath.Join(dir, "agent-refresh.lock"),
		refreshWindow: defaultRefreshWindow,
		lockWait:      time.Second,
		lockStale:     time.Second,
		now:           func() time.Time { return now },
		sleep: func(context.Context, time.Duration) error {
			return nil
		},
	}
}

type refreshErrorAPI struct {
	err error
}

func (f *refreshErrorAPI) RequestAgentOTP(context.Context, api.AgentOTPRequest, string) (api.OTPRequestAccepted, error) {
	return api.OTPRequestAccepted{}, nil
}

func (f *refreshErrorAPI) VerifyAgentOTP(context.Context, api.AgentOTPVerificationRequest, string) (api.AgentSessionResponse, error) {
	return api.AgentSessionResponse{}, nil
}

func (f *refreshErrorAPI) RefreshAgentSession(context.Context, api.AgentRefreshRequest, string) (api.AgentSessionResponse, error) {
	return api.AgentSessionResponse{}, f.err
}

func (f *refreshErrorAPI) CurrentUser(context.Context, string) (api.CurrentAccount, error) {
	return api.CurrentAccount{}, nil
}

func (f *refreshErrorAPI) ListRecorderCards(context.Context, string) (api.AgentRecorderCardPage, error) {
	return api.AgentRecorderCardPage{}, nil
}

func (f *refreshErrorAPI) GetQuotaSummary(context.Context, string) (api.AgentQuotaSummary, error) {
	return api.AgentQuotaSummary{}, nil
}

func (f *refreshErrorAPI) GetModelUsageSummary(context.Context, string) (api.AgentModelUsageSummary, error) {
	return api.AgentModelUsageSummary{}, nil
}

func (f *refreshErrorAPI) ListMemories(context.Context, string, api.ListMemoriesParams) (api.AgentMemoryPage, error) {
	return api.AgentMemoryPage{}, nil
}

func (f *refreshErrorAPI) GetMemory(context.Context, string, string, string) (api.AgentMemory, error) {
	return api.AgentMemory{}, nil
}

func (f *refreshErrorAPI) GetMemoryDeliveryDocument(context.Context, string, string, string) (api.AgentDeliveryDocument, error) {
	return api.AgentDeliveryDocument{}, nil
}

func (f *refreshErrorAPI) GetMemoryModelIO(context.Context, string, string, string) (api.AgentModelIOExport, error) {
	return api.AgentModelIOExport{}, nil
}

func (f *refreshErrorAPI) GetModelRunIOTrace(context.Context, string, string, string) (api.AgentModelIOExport, error) {
	return api.AgentModelIOExport{}, nil
}

func (f *refreshErrorAPI) ListModelIOTraces(context.Context, string, api.ListModelIOTracesParams) (api.AgentModelIOTracePage, error) {
	return api.AgentModelIOTracePage{}, nil
}

func (f *refreshErrorAPI) GetOAuthAuthorizationServer(context.Context) (api.OAuthAuthorizationServerMetadata, error) {
	return api.OAuthAuthorizationServerMetadata{}, nil
}

func (f *refreshErrorAPI) ExchangeOAuthCode(context.Context, api.OAuthTokenRequest) (api.OAuthTokenResponse, error) {
	return api.OAuthTokenResponse{}, nil
}

func (f *refreshErrorAPI) RefreshOAuthToken(context.Context, api.OAuthTokenRequest) (api.OAuthTokenResponse, error) {
	return api.OAuthTokenResponse{}, f.err
}

func (f *refreshErrorAPI) RevokeOAuthToken(context.Context, string) error {
	return nil
}

func (f *refreshErrorAPI) Logout(context.Context, string) error {
	return nil
}
