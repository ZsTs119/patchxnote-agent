package keychain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound    = errors.New("credential not found")
	ErrUnavailable = errors.New("credential store unavailable")
)

type Credential struct {
	AccountID             string    `json:"account_id"`
	AccessToken           string    `json:"access_token,omitempty"`
	RefreshToken          string    `json:"refresh_token,omitempty"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at,omitempty"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at,omitempty"`
	Scopes                []string  `json:"scopes,omitempty"`
}

type Store interface {
	Get(ctx context.Context, profile string) (Credential, error)
	Put(ctx context.Context, profile string, credential Credential) error
	Delete(ctx context.Context, profile string) error
}

type SecretStore interface {
	GetSecret(ctx context.Context, profile string, name string) (string, error)
	PutSecret(ctx context.Context, profile string, name string, value string) error
	DeleteSecret(ctx context.Context, profile string, name string) error
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsUnavailable(err error) bool {
	return errors.Is(err, ErrUnavailable)
}

type UnavailableStore struct {
	Reason string
}

func (s UnavailableStore) Get(ctx context.Context, profile string) (Credential, error) {
	return Credential{}, s.err()
}

func (s UnavailableStore) Put(ctx context.Context, profile string, credential Credential) error {
	return s.err()
}

func (s UnavailableStore) Delete(ctx context.Context, profile string) error {
	return s.err()
}

func (s UnavailableStore) GetSecret(ctx context.Context, profile string, name string) (string, error) {
	return "", s.err()
}

func (s UnavailableStore) PutSecret(ctx context.Context, profile string, name string, value string) error {
	return s.err()
}

func (s UnavailableStore) DeleteSecret(ctx context.Context, profile string, name string) error {
	return s.err()
}

func (s UnavailableStore) err() error {
	if s.Reason == "" {
		return ErrUnavailable
	}
	return fmt.Errorf("%w: %s", ErrUnavailable, s.Reason)
}
