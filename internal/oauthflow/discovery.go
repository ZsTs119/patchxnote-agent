package oauthflow

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
)

const DefaultClientID = "patchxnote-local-dev"

type AuthorizeURLInput struct {
	ClientID      string
	RedirectURI   string
	State         string
	CodeChallenge string
	Scope         string
}

func NormalizeBaseURL(value string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(value), "/")
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("server base URL must be an absolute URL without query or fragment")
	}
	if parsed.Scheme == "https" {
		return trimmed, nil
	}
	if parsed.Scheme == "http" && isLoopbackHost(parsed.Hostname()) {
		return trimmed, nil
	}
	return "", fmt.Errorf("server base URL must use HTTPS unless it is local loopback")
}

func ValidateAuthorizationServerMetadata(serverBaseURL string, metadata api.OAuthAuthorizationServerMetadata) error {
	normalizedBase, err := NormalizeBaseURL(serverBaseURL)
	if err != nil {
		return err
	}
	if strings.TrimSpace(metadata.Issuer) != normalizedBase {
		return fmt.Errorf("oauth issuer does not match server base URL")
	}
	if metadata.AuthorizationEndpoint == "" || metadata.TokenEndpoint == "" || metadata.RevocationEndpoint == "" {
		return fmt.Errorf("oauth metadata is missing required endpoints")
	}
	if !stringSliceContains(metadata.ResponseTypesSupported, "code") {
		return fmt.Errorf("oauth server does not support authorization code response type")
	}
	if !stringSliceContains(metadata.GrantTypesSupported, "authorization_code") ||
		!stringSliceContains(metadata.GrantTypesSupported, "refresh_token") {
		return fmt.Errorf("oauth server does not support required grant types")
	}
	if !stringSliceContains(metadata.CodeChallengeMethodsSupported, "S256") {
		return fmt.Errorf("oauth server does not support PKCE S256")
	}
	for _, endpoint := range []string{metadata.AuthorizationEndpoint, metadata.TokenEndpoint, metadata.RevocationEndpoint} {
		if err := validateEndpointUnderBase(normalizedBase, endpoint); err != nil {
			return err
		}
	}
	return nil
}

func BuildAuthorizeURL(metadata api.OAuthAuthorizationServerMetadata, input AuthorizeURLInput) (string, error) {
	if strings.TrimSpace(input.ClientID) == "" {
		return "", fmt.Errorf("oauth client id is required")
	}
	if strings.TrimSpace(input.RedirectURI) == "" {
		return "", fmt.Errorf("oauth redirect URI is required")
	}
	if strings.TrimSpace(input.State) == "" {
		return "", fmt.Errorf("oauth state is required")
	}
	if strings.TrimSpace(input.CodeChallenge) == "" {
		return "", fmt.Errorf("oauth code challenge is required")
	}
	authorizeURL, err := url.Parse(metadata.AuthorizationEndpoint)
	if err != nil || authorizeURL.Scheme == "" || authorizeURL.Host == "" {
		return "", fmt.Errorf("oauth authorization endpoint is invalid")
	}
	query := authorizeURL.Query()
	query.Set("response_type", "code")
	query.Set("client_id", input.ClientID)
	query.Set("redirect_uri", input.RedirectURI)
	query.Set("state", input.State)
	query.Set("code_challenge", input.CodeChallenge)
	query.Set("code_challenge_method", "S256")
	if strings.TrimSpace(input.Scope) != "" {
		query.Set("scope", strings.Join(strings.Fields(input.Scope), " "))
	}
	authorizeURL.RawQuery = query.Encode()
	return authorizeURL.String(), nil
}

func validateEndpointUnderBase(serverBaseURL string, endpoint string) error {
	base, err := url.Parse(serverBaseURL)
	if err != nil {
		return err
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("oauth endpoint is invalid")
	}
	if parsed.Scheme != base.Scheme || parsed.Host != base.Host {
		return fmt.Errorf("oauth endpoint origin does not match server base URL")
	}
	basePath := strings.TrimRight(base.EscapedPath(), "/")
	if basePath != "" && parsed.EscapedPath() != basePath && !strings.HasPrefix(parsed.EscapedPath(), basePath+"/") {
		return fmt.Errorf("oauth endpoint path is outside server base URL")
	}
	if base.Scheme != "https" && !(base.Scheme == "http" && isLoopbackHost(base.Hostname())) {
		return fmt.Errorf("oauth endpoint must use HTTPS unless local loopback")
	}
	return nil
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func isLoopbackHost(host string) bool {
	return host == "127.0.0.1" || host == "localhost" || host == "::1"
}
