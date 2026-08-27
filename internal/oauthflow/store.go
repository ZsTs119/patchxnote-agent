package oauthflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
)

const (
	SecretAccessToken  = "mcp_oauth:access_token"
	SecretRefreshToken = "mcp_oauth:refresh_token"
	SecretMetadata     = "mcp_oauth:metadata"
)

var ErrInvalidCredential = errors.New("invalid mcp oauth credential")

type Metadata struct {
	SchemaVersion          string    `json:"schema_version"`
	ServerBaseURL          string    `json:"server_base_url"`
	ClientID               string    `json:"client_id"`
	TokenType              string    `json:"token_type"`
	AccessTokenExpiresAt   time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt  time.Time `json:"refresh_token_expires_at"`
	ConnectorSessionID     string    `json:"connector_session_id,omitempty"`
	Scope                  string    `json:"scope"`
	PatchXNoteSchemaNotice string    `json:"patchxnote_schema_notice,omitempty"`
}

type Credential struct {
	AccessToken  string
	RefreshToken string
	Metadata     Metadata
}

type Store struct {
	secrets keychain.SecretStore
	profile string
}

func NewStore(secrets keychain.SecretStore, profile string) *Store {
	if secrets == nil {
		secrets = keychain.UnavailableStore{}
	}
	if strings.TrimSpace(profile) == "" {
		profile = "default"
	}
	return &Store{secrets: secrets, profile: profile}
}

func (s *Store) Load(ctx context.Context) (Credential, bool, error) {
	metadataRaw, err := s.secrets.GetSecret(ctx, s.profile, SecretMetadata)
	if err != nil {
		if keychain.IsNotFound(err) {
			return Credential{}, false, nil
		}
		return Credential{}, false, err
	}
	var metadata Metadata
	if err := json.Unmarshal([]byte(metadataRaw), &metadata); err != nil {
		return Credential{}, false, fmt.Errorf("%w: decode metadata", ErrInvalidCredential)
	}
	if err := validateMetadata(metadata); err != nil {
		return Credential{}, false, err
	}
	accessToken, err := s.secrets.GetSecret(ctx, s.profile, SecretAccessToken)
	if err != nil {
		return Credential{}, false, fmt.Errorf("%w: access token missing", ErrInvalidCredential)
	}
	refreshToken, err := s.secrets.GetSecret(ctx, s.profile, SecretRefreshToken)
	if err != nil {
		return Credential{}, false, fmt.Errorf("%w: refresh token missing", ErrInvalidCredential)
	}
	if strings.TrimSpace(accessToken) == "" || strings.TrimSpace(refreshToken) == "" {
		return Credential{}, false, fmt.Errorf("%w: token values are incomplete", ErrInvalidCredential)
	}
	return Credential{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Metadata:     metadata,
	}, true, nil
}

func (s *Store) Save(ctx context.Context, credential Credential) error {
	if err := validateCredential(credential); err != nil {
		return err
	}
	if err := s.Delete(ctx); err != nil {
		return err
	}
	if err := s.secrets.PutSecret(ctx, s.profile, SecretAccessToken, credential.AccessToken); err != nil {
		return err
	}
	if err := s.secrets.PutSecret(ctx, s.profile, SecretRefreshToken, credential.RefreshToken); err != nil {
		return err
	}
	metadata, err := json.Marshal(credential.Metadata)
	if err != nil {
		return fmt.Errorf("encode mcp oauth metadata: %w", err)
	}
	return s.secrets.PutSecret(ctx, s.profile, SecretMetadata, string(metadata))
}

func (s *Store) Delete(ctx context.Context) error {
	var firstErr error
	for _, name := range []string{SecretAccessToken, SecretRefreshToken, SecretMetadata} {
		if err := s.secrets.DeleteSecret(ctx, s.profile, name); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func CredentialFromTokenResponse(response api.OAuthTokenResponse, serverBaseURL string, clientID string, now time.Time) (Credential, error) {
	serverBaseURL, err := NormalizeBaseURL(serverBaseURL)
	if err != nil {
		return Credential{}, err
	}
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return Credential{}, fmt.Errorf("oauth client id is required")
	}
	response.TokenType = canonicalTokenType(response.TokenType)
	response.Scope = strings.Join(strings.Fields(response.Scope), " ")
	if strings.TrimSpace(response.AccessToken) == "" ||
		response.TokenType != "Bearer" ||
		response.ExpiresIn <= 0 ||
		strings.TrimSpace(response.RefreshToken) == "" ||
		response.RefreshTokenExpiresIn <= 0 ||
		strings.TrimSpace(response.Scope) == "" ||
		strings.TrimSpace(response.ConnectorSessionID) == "" {
		return Credential{}, fmt.Errorf("oauth token response is incomplete")
	}
	if now.IsZero() {
		now = time.Now()
	}
	return Credential{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		Metadata: Metadata{
			SchemaVersion:          "1",
			ServerBaseURL:          serverBaseURL,
			ClientID:               clientID,
			TokenType:              response.TokenType,
			AccessTokenExpiresAt:   now.UTC().Add(time.Duration(response.ExpiresIn) * time.Second),
			RefreshTokenExpiresAt:  now.UTC().Add(time.Duration(response.RefreshTokenExpiresIn) * time.Second),
			ConnectorSessionID:     response.ConnectorSessionID,
			Scope:                  response.Scope,
			PatchXNoteSchemaNotice: response.PatchXNoteSchemaNotice,
		},
	}, nil
}

func (credential Credential) Scopes() []string {
	return strings.Fields(credential.Metadata.Scope)
}

func (credential Credential) Matches(serverBaseURL string, clientID string) bool {
	serverBaseURL, err := NormalizeBaseURL(serverBaseURL)
	if err != nil {
		return false
	}
	return credential.Metadata.ServerBaseURL == serverBaseURL &&
		credential.Metadata.ClientID == strings.TrimSpace(clientID) &&
		credential.Metadata.TokenType == "Bearer"
}

func (credential Credential) RefreshValid(now time.Time) bool {
	if now.IsZero() {
		now = time.Now()
	}
	return credential.RefreshToken != "" && credential.Metadata.RefreshTokenExpiresAt.After(now)
}

func validateCredential(credential Credential) error {
	if strings.TrimSpace(credential.AccessToken) == "" || strings.TrimSpace(credential.RefreshToken) == "" {
		return fmt.Errorf("%w: token values are incomplete", ErrInvalidCredential)
	}
	return validateMetadata(credential.Metadata)
}

func validateMetadata(metadata Metadata) error {
	if metadata.SchemaVersion != "1" ||
		strings.TrimSpace(metadata.ServerBaseURL) == "" ||
		strings.TrimSpace(metadata.ClientID) == "" ||
		metadata.TokenType != "Bearer" ||
		metadata.AccessTokenExpiresAt.IsZero() ||
		metadata.RefreshTokenExpiresAt.IsZero() ||
		strings.TrimSpace(metadata.Scope) == "" {
		return fmt.Errorf("%w: metadata is incomplete", ErrInvalidCredential)
	}
	if _, err := NormalizeBaseURL(metadata.ServerBaseURL); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCredential, err)
	}
	return nil
}

func canonicalTokenType(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "Bearer") {
		return "Bearer"
	}
	return strings.TrimSpace(value)
}
