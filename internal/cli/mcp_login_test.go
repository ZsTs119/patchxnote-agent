package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/config"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
	"github.com/ZsTs119/patchxnote-agent/internal/oauthflow"
)

func TestRootCommandIncludesMCPSubcommands(t *testing.T) {
	stdout, stderr, err := executeForTest(t, "mcp", "--help")
	if err != nil {
		t.Fatalf("mcp help: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
	for _, want := range []string{"config", "login", "logout", "serve", "status"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected mcp help to contain %q, got:\n%s", want, stdout)
		}
	}
}

func TestMCPConfigPrintsSecretFreeJSON(t *testing.T) {
	stdout, stderr, err := executeForTest(t, "mcp", "config")
	if err != nil {
		t.Fatalf("mcp config: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(stdout), &decoded); err != nil {
		t.Fatalf("expected config JSON: %v\n%s", err, stdout)
	}
	combined := stdout + stderr
	for _, disallowed := range []string{"access_token", "refresh_token", "authorization_code"} {
		if strings.Contains(combined, disallowed) {
			t.Fatalf("mcp config leaked %q:\n%s", disallowed, combined)
		}
	}
}

func TestMCPLoginBrowserOAuthStoresCredentialAndDoesNotLeak(t *testing.T) {
	store := keychain.NewMemoryStore()
	serverBaseURL := "http://127.0.0.1:18080/patchnote-test-api"
	fakeAPI := &fakeAgentAPI{oauthMeta: oauthMetadataForBase(serverBaseURL)}
	codeValue := "code_should_not_be_printed"
	accessToken := "access_should_not_be_printed"
	refreshToken := "refresh_should_not_be_printed"
	fakeAPI.oauthToken = api.OAuthTokenResponse{
		AccessToken:           accessToken,
		TokenType:             "Bearer",
		ExpiresIn:             3600,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresIn: 2592000,
		Scope:                 "agent:account.read agent:content.read:mobile",
		ConnectorSessionID:    "connector_fixture",
	}

	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: store,
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return fakeAPI, nil
		},
		BrowserOpen: func(rawURL string) error {
			parsed, err := url.Parse(rawURL)
			if err != nil {
				return err
			}
			query := parsed.Query()
			if query.Get("code_verifier") != "" || query.Get("access_token") != "" || query.Get("refresh_token") != "" {
				return fmt.Errorf("authorize URL contains secret material")
			}
			if query.Get("response_type") != "code" || query.Get("code_challenge_method") != "S256" {
				return fmt.Errorf("authorize URL missing OAuth parameters")
			}
			callbackURL := query.Get("redirect_uri")
			state := query.Get("state")
			go func() {
				_, _ = http.Get(callbackURL + "?code=" + url.QueryEscape(codeValue) + "&state=" + url.QueryEscape(state))
			}()
			return nil
		},
	}, "--output", "json", "--server-base-url", serverBaseURL, "mcp", "login", "--callback-timeout", "2s", "--skip-smoke")
	if err != nil {
		t.Fatalf("mcp login: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("expected login JSON: %v\n%s", err, stdout)
	}
	if result["logged_in"] != true {
		t.Fatalf("expected logged_in true, got %#v", result["logged_in"])
	}
	if _, ok := result["connector_session_id"]; ok {
		t.Fatal("connector_session_id should not be printed by default")
	}
	if fakeAPI.oauthRequest.Code != codeValue || fakeAPI.oauthRequest.CodeVerifier == "" {
		t.Fatalf("expected token exchange with code and verifier, got %+v", fakeAPI.oauthRequest)
	}
	credential, ok, err := oauthflow.NewStore(store, "default").Load(context.Background())
	if err != nil || !ok {
		t.Fatalf("load stored mcp credential ok=%v err=%v", ok, err)
	}
	if credential.AccessToken != accessToken || credential.RefreshToken != refreshToken {
		t.Fatal("stored MCP credential does not match token response")
	}
	if _, err := store.Get(context.Background(), "default"); !keychain.IsNotFound(err) {
		t.Fatalf("legacy Agent credential should be untouched, got %v", err)
	}
	assertNoMCPSecretLeak(t, stdout+stderr, codeValue, accessToken, refreshToken, fakeAPI.oauthRequest.CodeVerifier)
}

func TestMCPLoginAlreadyLoggedInAndForceReplacement(t *testing.T) {
	store := keychain.NewMemoryStore()
	serverBaseURL := "http://127.0.0.1:18081/patchnote-test-api"
	initial := mcpCredentialForTest(serverBaseURL, "old_access", "old_refresh", time.Now().UTC().Add(time.Hour))
	if err := oauthflow.NewStore(store, "default").Save(context.Background(), initial); err != nil {
		t.Fatalf("seed credential: %v", err)
	}

	stdout, stderr, err := executeForTestWithDeps(t, Deps{CredentialStore: store}, "--output", "json", "--server-base-url", serverBaseURL, "mcp", "login", "--skip-smoke")
	if err != nil {
		t.Fatalf("already logged in: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no browser stderr for existing credential, got %q", stderr)
	}
	var already map[string]any
	if err := json.Unmarshal([]byte(stdout), &already); err != nil {
		t.Fatalf("expected already logged in JSON: %v\n%s", err, stdout)
	}
	if already["already_logged_in"] != true {
		t.Fatalf("expected already_logged_in true, got:\n%s", stdout)
	}

	fakeAPI := &fakeAgentAPI{oauthMeta: oauthMetadataForBase(serverBaseURL)}
	fakeAPI.oauthToken = api.OAuthTokenResponse{
		AccessToken:           "new_access",
		TokenType:             "Bearer",
		ExpiresIn:             3600,
		RefreshToken:          "new_refresh",
		RefreshTokenExpiresIn: 2592000,
		Scope:                 "agent:account.read",
		ConnectorSessionID:    "connector_fixture",
	}
	_, _, err = executeForTestWithDeps(t, Deps{
		CredentialStore: store,
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return fakeAPI, nil
		},
		BrowserOpen: browserCallbackOpener(t, "new_code"),
	}, "--server-base-url", serverBaseURL, "mcp", "login", "--force", "--callback-timeout", "2s", "--skip-smoke")
	if err != nil {
		t.Fatalf("force login: %v", err)
	}
	credential, ok, err := oauthflow.NewStore(store, "default").Load(context.Background())
	if err != nil || !ok {
		t.Fatalf("load replacement credential ok=%v err=%v", ok, err)
	}
	if credential.AccessToken != "new_access" || credential.RefreshToken != "new_refresh" {
		t.Fatalf("expected replacement credential, got access=%q refresh=%q", credential.AccessToken, credential.RefreshToken)
	}
}

func TestMCPLoginTimeoutAndStateMismatchDoNotStoreCredential(t *testing.T) {
	for _, tc := range []struct {
		name        string
		browserOpen oauthflow.BrowserOpener
		args        []string
	}{
		{name: "timeout", args: []string{"--no-browser", "--callback-timeout", "1ms"}},
		{name: "state-mismatch", browserOpen: browserCallbackOpenerWithState(t, "code", "wrong_state"), args: []string{"--callback-timeout", "2s"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := keychain.NewMemoryStore()
			serverBaseURL := "http://127.0.0.1:18082/patchnote-test-api"
			fakeAPI := &fakeAgentAPI{oauthMeta: oauthMetadataForBase(serverBaseURL)}
			args := append([]string{"--server-base-url", serverBaseURL, "mcp", "login", "--skip-smoke"}, tc.args...)
			_, _, err := executeForTestWithDeps(t, Deps{
				CredentialStore: store,
				APIFactory: func(cfg config.Config) (agentAPI, error) {
					return fakeAPI, nil
				},
				BrowserOpen: tc.browserOpen,
			}, args...)
			if err == nil {
				t.Fatal("expected login failure")
			}
			if _, ok, loadErr := oauthflow.NewStore(store, "default").Load(context.Background()); loadErr != nil || ok {
				t.Fatalf("expected no stored credential, ok=%v err=%v", ok, loadErr)
			}
		})
	}
}

func TestMCPStatusNoCredentialJSON(t *testing.T) {
	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: keychain.NewMemoryStore(),
	}, "--output", "json", "mcp", "status")
	if err != nil {
		t.Fatalf("mcp status: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("expected status JSON: %v\n%s", err, stdout)
	}
	if result["authenticated"] != false || result["reason"] != "no_credential" {
		t.Fatalf("unexpected status result: %+v", result)
	}
}

func TestMCPLogoutRevokesAndDeletesCredential(t *testing.T) {
	store := keychain.NewMemoryStore()
	serverBaseURL := "http://127.0.0.1:18083/patchnote-test-api"
	credential := mcpCredentialForTest(serverBaseURL, "logout_access", "logout_refresh", time.Now().UTC().Add(time.Hour))
	if err := oauthflow.NewStore(store, "default").Save(context.Background(), credential); err != nil {
		t.Fatalf("seed credential: %v", err)
	}
	fakeAPI := &fakeAgentAPI{}
	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: store,
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return fakeAPI, nil
		},
	}, "--output", "json", "--server-base-url", serverBaseURL, "mcp", "logout")
	if err != nil {
		t.Fatalf("mcp logout: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
	if len(fakeAPI.oauthRevoked) != 2 || fakeAPI.oauthRevoked[0] != "logout_refresh" || fakeAPI.oauthRevoked[1] != "logout_access" {
		t.Fatalf("expected refresh and access token revokes, got %#v", fakeAPI.oauthRevoked)
	}
	if _, ok, loadErr := oauthflow.NewStore(store, "default").Load(context.Background()); loadErr != nil || ok {
		t.Fatalf("expected credential deletion, ok=%v err=%v", ok, loadErr)
	}
	assertNoMCPSecretLeak(t, stdout+stderr, "logout_access", "logout_refresh")
}

func TestMCPLogoutLocalOnlyIsIdempotent(t *testing.T) {
	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: keychain.NewMemoryStore(),
	}, "--output", "json", "mcp", "logout", "--local-only")
	if err != nil {
		t.Fatalf("mcp logout --local-only: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("expected logout JSON: %v\n%s", err, stdout)
	}
	if result["logged_out"] != true {
		t.Fatalf("unexpected logout result: %+v", result)
	}
}

func TestMCPServeUsesRemoteProxyWhenCredentialExists(t *testing.T) {
	var sawAuth bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mcp" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "Bearer remote_access" {
			sawAuth = true
		}
		var request struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case "initialize":
			fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":{"protocolVersion":"2025-06-18","serverInfo":{"name":"patchxnote","version":"test"},"capabilities":{}}}`, request.ID)
		case "tools/list":
			fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":{"tools":[]}}`, request.ID)
		case "tools/call":
			fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":{"content":[{"type":"text","text":"ok"}]}}`, request.ID)
		default:
			fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"error":{"code":-32601,"message":"method not found"}}`, request.ID)
		}
	}))
	defer server.Close()

	store := keychain.NewMemoryStore()
	if err := oauthflow.NewStore(store, "default").Save(context.Background(), mcpCredentialForTest(server.URL, "remote_access", "remote_refresh", time.Now().UTC().Add(time.Hour))); err != nil {
		t.Fatalf("seed mcp credential: %v", err)
	}
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"patchxnote_get_current_user","arguments":{}}}`,
	}, "\n") + "\n"
	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: store,
		Stdin:           strings.NewReader(input),
	}, "--server-base-url", server.URL, "mcp", "serve")
	if err != nil {
		t.Fatalf("mcp serve remote: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
	if !sawAuth {
		t.Fatal("expected remote proxy to send bearer authorization")
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected three remote JSON-RPC responses, got %d\n%s", len(lines), stdout)
	}
	for index, line := range lines {
		var response struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      json.RawMessage `json:"id"`
			Result  json.RawMessage `json:"result"`
		}
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			t.Fatalf("line %d not JSON: %v\n%s", index, err, line)
		}
		if response.JSONRPC != "2.0" || string(response.ID) != fmt.Sprintf("%d", index+1) {
			t.Fatalf("unexpected response line %d: %s", index, line)
		}
	}
	assertNoMCPSecretLeak(t, stdout+stderr, "remote_access", "remote_refresh")
}

func oauthMetadataForBase(base string) api.OAuthAuthorizationServerMetadata {
	base = strings.TrimRight(base, "/")
	return api.OAuthAuthorizationServerMetadata{
		Issuer:                        base,
		AuthorizationEndpoint:         base + "/v1/agent/oauth/authorize",
		TokenEndpoint:                 base + "/v1/agent/oauth/token",
		RevocationEndpoint:            base + "/v1/agent/oauth/revoke",
		ResponseTypesSupported:        []string{"code"},
		GrantTypesSupported:           []string{"authorization_code", "refresh_token"},
		CodeChallengeMethodsSupported: []string{"S256"},
		ScopesSupported:               []string{"agent:account.read", "agent:content.read:mobile"},
	}
}

func browserCallbackOpener(t *testing.T, code string) oauthflow.BrowserOpener {
	t.Helper()
	return browserCallbackOpenerWithState(t, code, "")
}

func browserCallbackOpenerWithState(t *testing.T, code string, stateOverride string) oauthflow.BrowserOpener {
	t.Helper()
	return func(rawURL string) error {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return err
		}
		query := parsed.Query()
		callbackURL := query.Get("redirect_uri")
		if callbackURL == "" {
			return fmt.Errorf("missing redirect_uri")
		}
		state := query.Get("state")
		if stateOverride != "" {
			state = stateOverride
		}
		go func() {
			_, _ = http.Get(callbackURL + "?code=" + url.QueryEscape(code) + "&state=" + url.QueryEscape(state))
		}()
		return nil
	}
}

func mcpCredentialForTest(serverBaseURL string, accessToken string, refreshToken string, accessExpiry time.Time) oauthflow.Credential {
	now := time.Now().UTC()
	return oauthflow.Credential{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Metadata: oauthflow.Metadata{
			SchemaVersion:         "1",
			ServerBaseURL:         strings.TrimRight(serverBaseURL, "/"),
			ClientID:              oauthflow.DefaultClientID,
			TokenType:             "Bearer",
			AccessTokenExpiresAt:  accessExpiry.UTC(),
			RefreshTokenExpiresAt: now.Add(30 * 24 * time.Hour),
			Scope:                 "agent:account.read agent:content.read:mobile",
		},
	}
}

func assertNoMCPSecretLeak(t *testing.T, output string, values ...string) {
	t.Helper()
	for _, value := range values {
		if value != "" && strings.Contains(output, value) {
			t.Fatalf("output leaked sensitive value %q:\n%s", value, output)
		}
	}
}
