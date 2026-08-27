package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/oauthflow"
	"github.com/ZsTs119/patchxnote-agent/internal/remotemcp"
	clientsetup "github.com/ZsTs119/patchxnote-agent/internal/setup"
	"github.com/ZsTs119/patchxnote-agent/internal/version"

	"github.com/spf13/cobra"
)

const defaultMCPLoginTimeout = oauthflow.DefaultCallbackTimeout

type mcpOAuthStatusResult struct {
	Authenticated          bool       `json:"authenticated"`
	Profile                string     `json:"profile"`
	ServerBaseURL          string     `json:"server_base_url"`
	ClientID               string     `json:"client_id"`
	Reason                 string     `json:"reason,omitempty"`
	Scopes                 []string   `json:"scopes,omitempty"`
	AccessTokenExpiresAt   *time.Time `json:"access_expires_at,omitempty"`
	RefreshTokenExpiresAt  *time.Time `json:"refresh_expires_at,omitempty"`
	Verified               bool       `json:"verified,omitempty"`
	PatchXNoteSchemaNotice string     `json:"patchxnote_schema_notice,omitempty"`
}

type mcpLoginResult struct {
	LoggedIn               bool      `json:"logged_in"`
	AlreadyLoggedIn        bool      `json:"already_logged_in,omitempty"`
	Profile                string    `json:"profile"`
	ServerBaseURL          string    `json:"server_base_url"`
	ClientID               string    `json:"client_id"`
	Scopes                 []string  `json:"scopes,omitempty"`
	AccessTokenExpiresAt   time.Time `json:"access_expires_at"`
	RefreshTokenExpiresAt  time.Time `json:"refresh_expires_at"`
	SmokeVerified          bool      `json:"smoke_verified,omitempty"`
	PatchXNoteSchemaNotice string    `json:"patchxnote_schema_notice,omitempty"`
}

func newMCPConfigCommand(state *rootState) *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Print generic secret-free MCP configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), clientsetup.MCPJSONSnippet()); err != nil {
				return err
			}
			return nil
		},
	}
}

func newMCPLoginCommand(state *rootState) *cobra.Command {
	var noBrowser bool
	var force bool
	var skipSmoke bool
	var callbackTimeout time.Duration
	var clientID string
	var scope string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in to PatchXNote MCP with browser OAuth",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runtimeState, err := loadRuntime(state)
			if err != nil {
				return err
			}
			result, err := runMCPLoginFlow(cmd, state, runtimeState, mcpLoginOptions{
				NoBrowser:       noBrowser,
				Force:           force,
				SkipSmoke:       skipSmoke,
				CallbackTimeout: callbackTimeout,
				ClientID:        clientID,
				Scope:           scope,
			})
			if err != nil {
				return err
			}
			return printMCPLoginResult(cmd, state, result)
		},
	}
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Do not open a browser; print the OAuth URL to stderr")
	cmd.Flags().BoolVar(&force, "force", false, "Replace existing MCP OAuth credentials after a successful login")
	cmd.Flags().BoolVar(&skipSmoke, "skip-smoke", false, "Skip the post-login remote MCP current-user smoke")
	cmd.Flags().DurationVar(&callbackTimeout, "callback-timeout", defaultMCPLoginTimeout, "Maximum time to wait for browser OAuth callback")
	cmd.Flags().StringVar(&clientID, "client-id", oauthflow.DefaultClientID, "OAuth client id")
	cmd.Flags().StringVar(&scope, "scope", "", "Optional OAuth scope string; defaults to server/client policy")
	return cmd
}

type mcpLoginOptions struct {
	NoBrowser       bool
	Force           bool
	SkipSmoke       bool
	CallbackTimeout time.Duration
	ClientID        string
	Scope           string
}

func runMCPLoginFlow(cmd *cobra.Command, state *rootState, runtimeState runtimeState, options mcpLoginOptions) (mcpLoginResult, error) {
	if options.CallbackTimeout <= 0 {
		return mcpLoginResult{}, fmt.Errorf("--callback-timeout must be positive")
	}
	if runtimeState.API == nil {
		return mcpLoginResult{}, fmt.Errorf("server base URL is required; set --server-base-url or PATCHXNOTE_SERVER_BASE_URL")
	}
	clientID := normalizedMCPClientID(options.ClientID)
	serverBaseURL, err := oauthflow.NormalizeBaseURL(runtimeState.Config.Server.BaseURL)
	if err != nil {
		return mcpLoginResult{}, err
	}
	store := newMCPOAuthStore(runtimeState)
	now := time.Now().UTC()
	if !options.Force {
		if existing, ok, err := store.Load(cmd.Context()); err != nil {
			return mcpLoginResult{}, err
		} else if ok && existing.Matches(serverBaseURL, clientID) && existing.RefreshValid(now) {
			return mcpLoginResult{
				LoggedIn:               true,
				AlreadyLoggedIn:        true,
				Profile:                runtimeState.Config.Profile,
				ServerBaseURL:          serverBaseURL,
				ClientID:               clientID,
				Scopes:                 existing.Scopes(),
				AccessTokenExpiresAt:   existing.Metadata.AccessTokenExpiresAt,
				RefreshTokenExpiresAt:  existing.Metadata.RefreshTokenExpiresAt,
				PatchXNoteSchemaNotice: existing.Metadata.PatchXNoteSchemaNotice,
			}, nil
		}
	}

	metadata, err := runtimeState.API.GetOAuthAuthorizationServer(cmd.Context())
	if err != nil {
		return mcpLoginResult{}, err
	}
	if err := oauthflow.ValidateAuthorizationServerMetadata(serverBaseURL, metadata); err != nil {
		return mcpLoginResult{}, err
	}
	pkce, err := oauthflow.GeneratePKCE()
	if err != nil {
		return mcpLoginResult{}, err
	}
	oauthState, err := newOpaqueID("state")
	if err != nil {
		return mcpLoginResult{}, err
	}
	callbackContext, cancelCallback := context.WithCancel(cmd.Context())
	defer cancelCallback()
	callback, err := oauthflow.StartCallbackServer(callbackContext, oauthState)
	if err != nil {
		return mcpLoginResult{}, err
	}
	defer callback.Close(context.Background())

	authorizeURL, err := oauthflow.BuildAuthorizeURL(metadata, oauthflow.AuthorizeURLInput{
		ClientID:      clientID,
		RedirectURI:   callback.RedirectURI(),
		State:         oauthState,
		CodeChallenge: pkce.Challenge,
		Scope:         options.Scope,
	})
	if err != nil {
		return mcpLoginResult{}, err
	}
	if options.NoBrowser {
		fmt.Fprintf(cmd.ErrOrStderr(), "Open PatchXNote MCP login URL in your browser:\n%s\n", authorizeURL)
	} else {
		fmt.Fprintln(cmd.ErrOrStderr(), "Opening PatchXNote MCP login in your browser...")
		opener := state.browserOpen
		if opener == nil {
			opener = oauthflow.OpenBrowser
		}
		if err := opener(authorizeURL); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Could not open browser automatically: %v\nOpen PatchXNote MCP login URL in your browser:\n%s\n", err, authorizeURL)
		}
	}

	waitContext, cancelWait := context.WithTimeout(cmd.Context(), options.CallbackTimeout)
	defer cancelWait()
	callbackResult, err := callback.Wait(waitContext)
	if err != nil {
		return mcpLoginResult{}, err
	}
	token, err := runtimeState.API.ExchangeOAuthCode(cmd.Context(), apiOAuthTokenRequest(callbackResult.Code, callback.RedirectURI(), clientID, pkce.Verifier))
	if err != nil {
		return mcpLoginResult{}, err
	}
	credential, err := oauthflow.CredentialFromTokenResponse(token, serverBaseURL, clientID, now)
	if err != nil {
		return mcpLoginResult{}, err
	}
	if err := store.Save(cmd.Context(), credential); err != nil {
		return mcpLoginResult{}, err
	}
	smokeVerified := false
	if !options.SkipSmoke {
		if err := runRemoteCurrentUserSmoke(cmd.Context(), serverBaseURL, credential.AccessToken); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: MCP login saved, but remote smoke failed: %v\n", err)
		} else {
			smokeVerified = true
		}
	}
	return mcpLoginResult{
		LoggedIn:               true,
		Profile:                runtimeState.Config.Profile,
		ServerBaseURL:          serverBaseURL,
		ClientID:               clientID,
		Scopes:                 credential.Scopes(),
		AccessTokenExpiresAt:   credential.Metadata.AccessTokenExpiresAt,
		RefreshTokenExpiresAt:  credential.Metadata.RefreshTokenExpiresAt,
		SmokeVerified:          smokeVerified,
		PatchXNoteSchemaNotice: credential.Metadata.PatchXNoteSchemaNotice,
	}, nil
}

func newMCPStatusCommand(state *rootState) *cobra.Command {
	var verify bool
	var clientID string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print PatchXNote MCP OAuth login status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runtimeState, err := loadRuntime(state)
			if err != nil {
				return err
			}
			clientID = normalizedMCPClientID(clientID)
			result := mcpOAuthStatusResult{
				Authenticated: false,
				Profile:       runtimeState.Config.Profile,
				ServerBaseURL: strings.TrimRight(strings.TrimSpace(runtimeState.Config.Server.BaseURL), "/"),
				ClientID:      clientID,
			}
			store := newMCPOAuthStore(runtimeState)
			credential, ok, err := store.Load(cmd.Context())
			if err != nil {
				if errors.Is(err, oauthflow.ErrInvalidCredential) {
					result.Reason = "invalid_credential"
				} else {
					result.Reason = "credential_unavailable"
				}
				return printMCPStatusResult(cmd, state, result)
			}
			if !ok {
				result.Reason = "no_credential"
				return printMCPStatusResult(cmd, state, result)
			}
			if !credential.Matches(runtimeState.Config.Server.BaseURL, clientID) {
				result.Reason = "server_or_client_mismatch"
				return printMCPStatusResult(cmd, state, result)
			}
			if !credential.RefreshValid(time.Now().UTC()) {
				result.Reason = "refresh_expired"
				return printMCPStatusResult(cmd, state, result)
			}
			result.Authenticated = true
			result.Reason = "ok"
			result.ServerBaseURL = credential.Metadata.ServerBaseURL
			result.ClientID = credential.Metadata.ClientID
			result.Scopes = credential.Scopes()
			accessExpiresAt := credential.Metadata.AccessTokenExpiresAt
			refreshExpiresAt := credential.Metadata.RefreshTokenExpiresAt
			result.AccessTokenExpiresAt = &accessExpiresAt
			result.RefreshTokenExpiresAt = &refreshExpiresAt
			result.PatchXNoteSchemaNotice = credential.Metadata.PatchXNoteSchemaNotice
			if verify {
				provider := newMCPOAuthRefreshProvider(runtimeState, store, clientID)
				refreshed, ok, err := provider.Credential(cmd.Context())
				if err != nil {
					return err
				}
				if !ok {
					result.Authenticated = false
					result.Reason = "refresh_expired"
					return printMCPStatusResult(cmd, state, result)
				}
				if err := runRemoteCurrentUserSmoke(cmd.Context(), refreshed.Metadata.ServerBaseURL, refreshed.AccessToken); err != nil {
					return err
				}
				result.Verified = true
				accessExpiresAt = refreshed.Metadata.AccessTokenExpiresAt
				refreshExpiresAt = refreshed.Metadata.RefreshTokenExpiresAt
				result.AccessTokenExpiresAt = &accessExpiresAt
				result.RefreshTokenExpiresAt = &refreshExpiresAt
			}
			return printMCPStatusResult(cmd, state, result)
		},
	}
	cmd.Flags().BoolVar(&verify, "verify", false, "Verify the stored credential with a safe remote MCP current-user call")
	cmd.Flags().StringVar(&clientID, "client-id", oauthflow.DefaultClientID, "OAuth client id")
	return cmd
}

func newMCPLogoutCommand(state *rootState) *cobra.Command {
	var localOnly bool
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Revoke and remove PatchXNote MCP OAuth credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			runtimeState, err := loadRuntime(state)
			if err != nil {
				return err
			}
			store := newMCPOAuthStore(runtimeState)
			credential, ok, loadErr := store.Load(cmd.Context())
			if loadErr != nil && !errors.Is(loadErr, oauthflow.ErrInvalidCredential) {
				return loadErr
			}
			if ok && !localOnly && runtimeState.API != nil {
				if credential.RefreshToken != "" {
					if err := runtimeState.API.RevokeOAuthToken(cmd.Context(), credential.RefreshToken); err != nil {
						fmt.Fprintln(cmd.ErrOrStderr(), "warning: remote MCP refresh token revoke failed; local credentials will still be removed")
					}
				}
				if credential.AccessToken != "" {
					if err := runtimeState.API.RevokeOAuthToken(cmd.Context(), credential.AccessToken); err != nil {
						fmt.Fprintln(cmd.ErrOrStderr(), "warning: remote MCP access token revoke failed; local credentials will still be removed")
					}
				}
			}
			if err := store.Delete(cmd.Context()); err != nil {
				return err
			}
			result := struct {
				LoggedOut bool   `json:"logged_out"`
				Profile   string `json:"profile"`
			}{
				LoggedOut: true,
				Profile:   runtimeState.Config.Profile,
			}
			switch format := normalizedOutputFormat(state); format {
			case "", "plain":
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "mcp logged out\nprofile %s\n", result.Profile)
				return err
			case "json":
				return writeJSON(cmd.OutOrStdout(), result)
			default:
				return unsupportedOutputFormatError(format)
			}
		},
	}
	cmd.Flags().BoolVar(&localOnly, "local-only", false, "Remove local MCP OAuth credentials without contacting the server")
	return cmd
}

func printMCPLoginResult(cmd *cobra.Command, state *rootState, result mcpLoginResult) error {
	switch format := normalizedOutputFormat(state); format {
	case "", "plain":
		status := "mcp login succeeded"
		if result.AlreadyLoggedIn {
			status = "mcp already logged in"
		}
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\nprofile %s\nserver %s\nscopes %s\naccess expires %s\nrefresh expires %s\n",
			status,
			result.Profile,
			result.ServerBaseURL,
			strings.Join(result.Scopes, " "),
			result.AccessTokenExpiresAt.UTC().Format(time.RFC3339),
			result.RefreshTokenExpiresAt.UTC().Format(time.RFC3339),
		)
		return err
	case "json":
		return writeJSON(cmd.OutOrStdout(), result)
	default:
		return unsupportedOutputFormatError(format)
	}
}

func printMCPStatusResult(cmd *cobra.Command, state *rootState, result mcpOAuthStatusResult) error {
	switch format := normalizedOutputFormat(state); format {
	case "", "plain":
		if result.Authenticated {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "mcp authenticated\nprofile %s\nserver %s\nscopes %s\naccess expires %s\nrefresh expires %s\n",
				result.Profile,
				result.ServerBaseURL,
				strings.Join(result.Scopes, " "),
				formatOptionalTime(result.AccessTokenExpiresAt),
				formatOptionalTime(result.RefreshTokenExpiresAt),
			)
			return err
		}
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "mcp unauthenticated\nprofile %s\nreason %s\n", result.Profile, result.Reason)
		return err
	case "json":
		return writeJSON(cmd.OutOrStdout(), result)
	default:
		return unsupportedOutputFormatError(format)
	}
}

func apiOAuthTokenRequest(code string, redirectURI string, clientID string, verifier string) api.OAuthTokenRequest {
	return api.OAuthTokenRequest{
		GrantType:    "authorization_code",
		Code:         code,
		RedirectURI:  redirectURI,
		ClientID:     clientID,
		CodeVerifier: verifier,
	}
}

func newMCPOAuthStore(runtimeState runtimeState) *oauthflow.Store {
	return oauthflow.NewStore(runtimeState.Secrets, runtimeState.Config.Profile)
}

func newMCPOAuthRefreshProvider(runtimeState runtimeState, store *oauthflow.Store, clientID string) *oauthflow.RefreshProvider {
	return &oauthflow.RefreshProvider{
		Store:         store,
		Client:        runtimeState.API,
		ServerBaseURL: runtimeState.Config.Server.BaseURL,
		ClientID:      normalizedMCPClientID(clientID),
		LockPath:      filepath.Join(runtimeState.Config.Paths.ConfigDir, "mcp-oauth-refresh.lock"),
	}
}

func runRemoteCurrentUserSmoke(ctx context.Context, serverBaseURL string, accessToken string) error {
	client, err := remotemcp.New(remotemcp.Options{
		ServerBaseURL: serverBaseURL,
		UserAgent:     fmt.Sprintf("patchxnote-agent/%s", version.Version),
	})
	if err != nil {
		return err
	}
	request := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"patchxnote_get_current_user","arguments":{}}}`)
	response, err := client.Do(ctx, request, accessToken)
	if err != nil {
		return err
	}
	if response.NoResponse {
		return fmt.Errorf("remote MCP smoke returned no response")
	}
	var decoded struct {
		Error *struct {
			Message string         `json:"message"`
			Data    map[string]any `json:"data"`
		} `json:"error,omitempty"`
		Result json.RawMessage `json:"result,omitempty"`
	}
	if err := json.Unmarshal(response.Body, &decoded); err != nil {
		return fmt.Errorf("decode remote MCP smoke response: %w", err)
	}
	if decoded.Error != nil {
		code, _ := decoded.Error.Data["code"].(string)
		if code == "" {
			code = "tool_error"
		}
		return fmt.Errorf("remote MCP smoke failed with %s", code)
	}
	if len(decoded.Result) == 0 {
		return fmt.Errorf("remote MCP smoke response has no result")
	}
	return nil
}

func shouldUseRemoteMCP(ctx context.Context, runtimeState runtimeState, store *oauthflow.Store, clientID string) (bool, error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("PATCHXNOTE_MCP_MODE")))
	switch mode {
	case "", "auto":
		credential, ok, err := store.Load(ctx)
		if err != nil {
			return false, nil
		}
		if !ok {
			return false, nil
		}
		return credential.Matches(runtimeState.Config.Server.BaseURL, normalizedMCPClientID(clientID)) && credential.RefreshValid(time.Now().UTC()), nil
	case "local":
		return false, nil
	case "remote":
		return true, nil
	default:
		return false, fmt.Errorf("unsupported PATCHXNOTE_MCP_MODE %q", mode)
	}
}

func normalizedMCPClientID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return oauthflow.DefaultClientID
	}
	return value
}

func formatOptionalTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return "unknown"
	}
	return value.UTC().Format(time.RFC3339)
}

type mcpOAuthRemoteTokenProvider struct {
	provider *oauthflow.RefreshProvider
}

func (p mcpOAuthRemoteTokenProvider) AccessToken(ctx context.Context) (string, bool, error) {
	if p.provider == nil {
		return "", false, nil
	}
	return p.provider.AccessToken(ctx)
}

func (p mcpOAuthRemoteTokenProvider) RefreshNow(ctx context.Context) (string, bool, error) {
	if p.provider == nil {
		return "", false, nil
	}
	credential, ok, err := p.provider.RefreshNow(ctx)
	if err != nil || !ok {
		return "", ok, err
	}
	return credential.AccessToken, true, nil
}
