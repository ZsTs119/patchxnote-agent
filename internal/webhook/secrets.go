package webhook

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
)

var ErrWebhookSecretMissing = errors.New("webhook secret missing")

type SecretStore struct {
	backend keychain.SecretStore
	profile string
}

func NewSecretStore(backend keychain.SecretStore, profile string) *SecretStore {
	if profile == "" {
		profile = "default"
	}
	if backend == nil {
		backend = keychain.UnavailableStore{}
	}
	return &SecretStore{backend: backend, profile: profile}
}

func (s *SecretStore) URL(ctx context.Context, alias string) (string, error) {
	normalized, err := ValidateAlias(alias)
	if err != nil {
		return "", err
	}
	value, err := s.backend.GetSecret(ctx, s.profile, secretName("url", normalized))
	if err != nil {
		if keychain.IsNotFound(err) {
			return "", fmt.Errorf("%w: webhook URL missing; run webhook set again", ErrWebhookSecretMissing)
		}
		return "", err
	}
	return value, nil
}

func (s *SecretStore) PutURL(ctx context.Context, alias string, value string) error {
	normalized, err := ValidateAlias(alias)
	if err != nil {
		return err
	}
	validated, err := ValidateWebhookURL(value)
	if err != nil {
		return err
	}
	return s.backend.PutSecret(ctx, s.profile, secretName("url", normalized), validated)
}

func (s *SecretStore) SigningSecret(ctx context.Context, alias string) (string, bool, error) {
	normalized, err := ValidateAlias(alias)
	if err != nil {
		return "", false, err
	}
	value, err := s.backend.GetSecret(ctx, s.profile, secretName("signing", normalized))
	if err != nil {
		if keychain.IsNotFound(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return value, true, nil
}

func (s *SecretStore) PutSigningSecret(ctx context.Context, alias string, value string) error {
	normalized, err := ValidateAlias(alias)
	if err != nil {
		return err
	}
	return s.backend.PutSecret(ctx, s.profile, secretName("signing", normalized), value)
}

func (s *SecretStore) DeleteSigningSecret(ctx context.Context, alias string) error {
	normalized, err := ValidateAlias(alias)
	if err != nil {
		return err
	}
	return s.backend.DeleteSecret(ctx, s.profile, secretName("signing", normalized))
}

func (s *SecretStore) DeleteTarget(ctx context.Context, alias string) error {
	normalized, err := ValidateAlias(alias)
	if err != nil {
		return err
	}
	if err := s.backend.DeleteSecret(ctx, s.profile, secretName("url", normalized)); err != nil {
		return err
	}
	return s.backend.DeleteSecret(ctx, s.profile, secretName("signing", normalized))
}

func secretName(kind string, alias string) string {
	return "webhook:" + kind + ":" + base64.RawURLEncoding.EncodeToString([]byte(alias))
}
