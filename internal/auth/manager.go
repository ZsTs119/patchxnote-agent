package auth

import (
	"context"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
)

type Manager struct {
	store   keychain.Store
	profile string
}

type Status struct {
	Authenticated         bool      `json:"authenticated"`
	Profile               string    `json:"profile"`
	AccountID             string    `json:"account_id,omitempty"`
	AccountStatus         string    `json:"account_status,omitempty"`
	RegistrationPlatform  string    `json:"registration_platform,omitempty"`
	PhoneMasked           string    `json:"phone_masked,omitempty"`
	StateVersion          int64     `json:"state_version,omitempty"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at,omitempty"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at,omitempty"`
	Scopes                []string  `json:"scopes,omitempty"`
}

func NewManager(store keychain.Store, profile string) *Manager {
	if profile == "" {
		profile = "default"
	}
	if store == nil {
		store = keychain.UnavailableStore{}
	}
	return &Manager{
		store:   store,
		profile: profile,
	}
}

func (m *Manager) Save(ctx context.Context, credential keychain.Credential) error {
	return m.store.Put(ctx, m.profile, credential)
}

func (m *Manager) Credential(ctx context.Context) (keychain.Credential, bool, error) {
	credential, err := m.store.Get(ctx, m.profile)
	if err != nil {
		if keychain.IsNotFound(err) || keychain.IsUnavailable(err) {
			return keychain.Credential{}, false, nil
		}
		return keychain.Credential{}, false, err
	}
	if credential.AccessToken == "" && credential.RefreshToken == "" {
		return keychain.Credential{}, false, nil
	}
	return credential, true, nil
}

func (m *Manager) Status(ctx context.Context) (Status, error) {
	status := Status{
		Authenticated: false,
		Profile:       m.profile,
	}

	credential, ok, err := m.Credential(ctx)
	if err != nil {
		return Status{}, err
	}
	if !ok {
		return status, nil
	}

	status.Authenticated = true
	status.AccountID = credential.AccountID
	status.AccessTokenExpiresAt = credential.AccessTokenExpiresAt
	status.RefreshTokenExpiresAt = credential.RefreshTokenExpiresAt
	status.Scopes = append([]string(nil), credential.Scopes...)
	return status, nil
}

func (m *Manager) Logout(ctx context.Context) error {
	if err := m.store.Delete(ctx, m.profile); err != nil {
		if keychain.IsNotFound(err) || keychain.IsUnavailable(err) {
			return nil
		}
		return err
	}
	return nil
}
