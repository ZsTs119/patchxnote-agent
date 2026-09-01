package oauthflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
)

func TestGeneratePKCE(t *testing.T) {
	pair, err := GeneratePKCEWithReader(bytes.NewReader(bytes.Repeat([]byte{7}, 32)))
	if err != nil {
		t.Fatalf("generate pkce: %v", err)
	}
	if len(pair.Verifier) != 43 || !verifierPattern.MatchString(pair.Verifier) {
		t.Fatalf("unexpected verifier: %q", pair.Verifier)
	}
	if pair.Challenge != ChallengeS256(pair.Verifier) || strings.Contains(pair.Challenge, "=") {
		t.Fatalf("unexpected challenge: %q", pair.Challenge)
	}
}

func TestAuthorizeURLValidationAndBuilder(t *testing.T) {
	metadata := api.OAuthAuthorizationServerMetadata{
		Issuer:                        "https://ws-lab.patch-x.cn/patchnote-test-api",
		AuthorizationEndpoint:         "https://ws-lab.patch-x.cn/patchnote-test-api/v1/agent/oauth/authorize",
		TokenEndpoint:                 "https://ws-lab.patch-x.cn/patchnote-test-api/v1/agent/oauth/token",
		RevocationEndpoint:            "https://ws-lab.patch-x.cn/patchnote-test-api/v1/agent/oauth/revoke",
		ResponseTypesSupported:        []string{"code"},
		GrantTypesSupported:           []string{"authorization_code", "refresh_token"},
		CodeChallengeMethodsSupported: []string{"S256"},
	}
	if err := ValidateAuthorizationServerMetadata("https://ws-lab.patch-x.cn/patchnote-test-api", metadata); err != nil {
		t.Fatalf("validate metadata: %v", err)
	}
	authorizeURL, err := BuildAuthorizeURL(metadata, AuthorizeURLInput{
		ClientID:      DefaultClientID,
		RedirectURI:   "http://127.0.0.1:49152/callback",
		State:         "state_fixture",
		CodeChallenge: strings.Repeat("c", 43),
		Scope:         "agent:account.read agent:model_io.read",
	})
	if err != nil {
		t.Fatalf("build authorize url: %v", err)
	}
	parsed, err := url.Parse(authorizeURL)
	if err != nil {
		t.Fatalf("parse authorize url: %v", err)
	}
	if parsed.Path != "/patchnote-test-api/v1/agent/oauth/authorize" {
		t.Fatalf("expected path prefix to be preserved, got %s", parsed.Path)
	}
	query := parsed.Query()
	if query.Get("response_type") != "code" || query.Get("client_id") != DefaultClientID ||
		query.Get("redirect_uri") != "http://127.0.0.1:49152/callback" ||
		query.Get("state") != "state_fixture" ||
		query.Get("code_challenge_method") != "S256" {
		t.Fatalf("unexpected authorize query: %s", parsed.RawQuery)
	}
	if strings.Contains(authorizeURL, "verifier") || strings.Contains(authorizeURL, "token") {
		t.Fatalf("authorize URL leaked forbidden material: %s", authorizeURL)
	}
}

func TestCallbackServerSuccessAndFailurePagesDoNotLeakQuery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	callback, err := StartCallbackServer(ctx, "expected_state")
	if err != nil {
		t.Fatalf("start callback: %v", err)
	}
	defer callback.Close(context.Background())

	response, err := http.Get(callback.RedirectURI() + "?code=code_fixture&state=expected_state")
	if err != nil {
		t.Fatalf("call callback: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected callback status: %d", response.StatusCode)
	}
	for _, header := range []string{"Cache-Control", "Content-Security-Policy", "Referrer-Policy", "X-Content-Type-Options"} {
		if response.Header.Get(header) == "" {
			t.Fatalf("expected callback header %s", header)
		}
	}
	var successBody bytes.Buffer
	if _, err := successBody.ReadFrom(response.Body); err != nil {
		t.Fatalf("read success body: %v", err)
	}
	for _, expected := range []string{
		"登录已完成",
		"可以回到编辑器继续使用。",
		"此页面可以关闭。",
		"PatchXNote",
	} {
		if !strings.Contains(successBody.String(), expected) {
			t.Fatalf("success page missing %q", expected)
		}
	}
	for _, forbidden := range []string{
		"MCP",
		"OAuth",
		"code_fixture",
		"expected_state",
	} {
		if strings.Contains(successBody.String(), forbidden) {
			t.Fatalf("success page leaked or exposed technical content %q", forbidden)
		}
	}
	result, err := callback.Wait(context.Background())
	if err != nil {
		t.Fatalf("wait callback: %v", err)
	}
	if result.Code != "code_fixture" || result.State != "expected_state" {
		t.Fatalf("unexpected callback result: %+v", result)
	}

	callback2, err := StartCallbackServer(context.Background(), "expected_state")
	if err != nil {
		t.Fatalf("start callback2: %v", err)
	}
	defer callback2.Close(context.Background())
	failed, err := http.Get(callback2.RedirectURI() + "?code=code_fixture&state=wrong_state")
	if err != nil {
		t.Fatalf("call failed callback: %v", err)
	}
	defer failed.Body.Close()
	var body bytes.Buffer
	if _, err := body.ReadFrom(failed.Body); err != nil {
		t.Fatalf("read failure body: %v", err)
	}
	if failed.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected failed callback status: %d", failed.StatusCode)
	}
	for _, expected := range []string{
		"登录未完成",
		"请回到应用重新打开登录。",
		"没有保存新的登录信息。",
		"PatchXNote",
	} {
		if !strings.Contains(body.String(), expected) {
			t.Fatalf("failure page missing %q", expected)
		}
	}
	for _, forbidden := range []string{
		"MCP",
		"OAuth",
		"code_fixture",
		"wrong_state",
		"expected_state",
	} {
		if strings.Contains(body.String(), forbidden) {
			t.Fatalf("failure page leaked or exposed technical content %q: %s", forbidden, body.String())
		}
	}
	_, err = callback2.Wait(context.Background())
	if !errors.Is(err, ErrCallbackDenied) {
		t.Fatalf("expected callback denied error, got %v", err)
	}
}

func TestStoreSeparatesMCPAuthFromLegacyCredential(t *testing.T) {
	memory := keychain.NewMemoryStore()
	if err := memory.Put(context.Background(), "default", keychain.Credential{AccountID: "legacy", AccessToken: "legacy_access"}); err != nil {
		t.Fatalf("seed legacy credential: %v", err)
	}
	store := NewStore(memory, "default")
	now := fixedOAuthTime()
	credential, err := CredentialFromTokenResponse(api.OAuthTokenResponse{
		AccessToken:           strings.Repeat("a", 43),
		TokenType:             "Bearer",
		ExpiresIn:             900,
		RefreshToken:          strings.Repeat("r", 43),
		RefreshTokenExpiresIn: 3600,
		Scope:                 "agent:account.read agent:model_io.read",
		ConnectorSessionID:    "mcpconn_fixture",
	}, "https://ws-lab.patch-x.cn/patchnote-test-api", DefaultClientID, now)
	if err != nil {
		t.Fatalf("build credential: %v", err)
	}
	if err := store.Save(context.Background(), credential); err != nil {
		t.Fatalf("save credential: %v", err)
	}
	loaded, ok, err := store.Load(context.Background())
	if err != nil || !ok {
		t.Fatalf("load credential ok=%t err=%v", ok, err)
	}
	if loaded.AccessToken != strings.Repeat("a", 43) || loaded.Metadata.Scope != "agent:account.read agent:model_io.read" {
		t.Fatalf("unexpected loaded credential: %+v", loaded)
	}
	legacy, err := memory.Get(context.Background(), "default")
	if err != nil {
		t.Fatalf("legacy credential missing: %v", err)
	}
	if legacy.AccountID != "legacy" || legacy.AccessToken != "legacy_access" {
		t.Fatalf("legacy credential was modified: %+v", legacy)
	}
	metadataRaw, err := memory.GetSecret(context.Background(), "default", SecretMetadata)
	if err != nil {
		t.Fatalf("metadata missing: %v", err)
	}
	if strings.Contains(metadataRaw, loaded.AccessToken) || strings.Contains(metadataRaw, loaded.RefreshToken) ||
		strings.Contains(metadataRaw, `"access_token":"`) || strings.Contains(metadataRaw, `"refresh_token":"`) {
		t.Fatalf("metadata leaked token fields: %s", metadataRaw)
	}
}

func TestCredentialFromTokenResponseRejectsIncompleteResponses(t *testing.T) {
	_, err := CredentialFromTokenResponse(api.OAuthTokenResponse{
		AccessToken:           strings.Repeat("a", 43),
		TokenType:             "bearer",
		ExpiresIn:             900,
		RefreshToken:          "",
		RefreshTokenExpiresIn: 3600,
		Scope:                 "agent:account.read",
		ConnectorSessionID:    "mcpconn_fixture",
	}, "https://ws-lab.patch-x.cn/patchnote-test-api", DefaultClientID, fixedOAuthTime())
	if err == nil {
		t.Fatal("expected incomplete token response error")
	}
}

func TestRefreshProviderRotatesAndStoresTokens(t *testing.T) {
	memory := keychain.NewMemoryStore()
	store := NewStore(memory, "default")
	now := fixedOAuthTime()
	oldCredential, err := CredentialFromTokenResponse(api.OAuthTokenResponse{
		AccessToken:           strings.Repeat("a", 43),
		TokenType:             "Bearer",
		ExpiresIn:             1,
		RefreshToken:          strings.Repeat("r", 43),
		RefreshTokenExpiresIn: 3600,
		Scope:                 "agent:account.read",
		ConnectorSessionID:    "mcpconn_fixture",
	}, "https://ws-lab.patch-x.cn/patchnote-test-api", DefaultClientID, now)
	if err != nil {
		t.Fatalf("old credential: %v", err)
	}
	if err := store.Save(context.Background(), oldCredential); err != nil {
		t.Fatalf("save old credential: %v", err)
	}
	client := &fakeOAuthTokenClient{}
	provider := &RefreshProvider{
		Store:         store,
		Client:        client,
		ServerBaseURL: "https://ws-lab.patch-x.cn/patchnote-test-api",
		ClientID:      DefaultClientID,
		LockPath:      t.TempDir() + "/mcp-oauth-refresh.lock",
		Now: func() time.Time {
			return now.Add(2 * time.Second)
		},
	}
	updated, ok, err := provider.Credential(context.Background())
	if err != nil || !ok {
		t.Fatalf("refresh credential ok=%t err=%v", ok, err)
	}
	if updated.AccessToken != strings.Repeat("b", 43) || updated.RefreshToken != strings.Repeat("n", 43) {
		t.Fatalf("expected rotated tokens, got %+v", updated)
	}
	if client.refreshToken != strings.Repeat("r", 43) {
		t.Fatal("refresh client did not receive old refresh token")
	}
}

type fakeOAuthTokenClient struct {
	refreshToken string
}

func (f *fakeOAuthTokenClient) RefreshOAuthToken(ctx context.Context, request api.OAuthTokenRequest) (api.OAuthTokenResponse, error) {
	f.refreshToken = request.RefreshToken
	return api.OAuthTokenResponse{
		AccessToken:           strings.Repeat("b", 43),
		TokenType:             "Bearer",
		ExpiresIn:             900,
		RefreshToken:          strings.Repeat("n", 43),
		RefreshTokenExpiresIn: 3600,
		Scope:                 "agent:account.read",
		ConnectorSessionID:    "mcpconn_fixture",
	}, nil
}

func fixedOAuthTime() time.Time {
	return time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)
}

func TestMetadataJSONRejectsTokens(t *testing.T) {
	credential, err := CredentialFromTokenResponse(api.OAuthTokenResponse{
		AccessToken:           strings.Repeat("a", 43),
		TokenType:             "Bearer",
		ExpiresIn:             900,
		RefreshToken:          strings.Repeat("r", 43),
		RefreshTokenExpiresIn: 3600,
		Scope:                 "agent:account.read",
		ConnectorSessionID:    "mcpconn_fixture",
	}, "https://ws-lab.patch-x.cn/patchnote-test-api", DefaultClientID, fixedOAuthTime())
	if err != nil {
		t.Fatalf("credential: %v", err)
	}
	body, err := json.Marshal(credential.Metadata)
	if err != nil {
		t.Fatalf("marshal metadata: %v", err)
	}
	for _, disallowed := range []string{credential.AccessToken, credential.RefreshToken, `"access_token":"`, `"refresh_token":"`} {
		if strings.Contains(string(body), disallowed) {
			t.Fatalf("metadata leaked %q: %s", disallowed, string(body))
		}
	}
}
