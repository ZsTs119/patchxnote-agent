package keychain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zalando/go-keyring"
)

const (
	nativeServiceName = "patchnote-agent"
	nativeMetadataKey = "metadata"
	nativeAccessKey   = "access_token"
	nativeRefreshKey  = "refresh_token"
)

type nativeBackend interface {
	Get(service, user string) (string, error)
	Set(service, user, secret string) error
	Delete(service, user string) error
}

type keyringBackend struct{}

func (keyringBackend) Get(service, user string) (string, error) {
	return keyring.Get(service, user)
}

func (keyringBackend) Set(service, user, secret string) error {
	return keyring.Set(service, user, secret)
}

func (keyringBackend) Delete(service, user string) error {
	return keyring.Delete(service, user)
}

type NativeStore struct {
	backend nativeBackend
	service string
}

type nativeMetadata struct {
	AccountID            string   `json:"account_id,omitempty"`
	AccessTokenExpiresAt string   `json:"access_token_expires_at,omitempty"`
	Scopes               []string `json:"scopes,omitempty"`
	HasAccessToken       bool     `json:"has_access_token,omitempty"`
	HasRefreshToken      bool     `json:"has_refresh_token,omitempty"`
}

func NewNativeStore() *NativeStore {
	return newNativeStore(keyringBackend{}, nativeServiceName)
}

func newNativeStore(backend nativeBackend, service string) *NativeStore {
	if service == "" {
		service = nativeServiceName
	}
	return &NativeStore{
		backend: backend,
		service: service,
	}
}

func (s *NativeStore) Get(ctx context.Context, profile string) (Credential, error) {
	if err := ctx.Err(); err != nil {
		return Credential{}, err
	}
	profile = normalizeProfile(profile)

	metadataSecret, err := s.backend.Get(s.service, nativeAccount(profile, nativeMetadataKey))
	if err != nil {
		return Credential{}, mapNativeGetError(err)
	}

	var metadata nativeMetadata
	if err := json.Unmarshal([]byte(metadataSecret), &metadata); err != nil {
		return Credential{}, fmt.Errorf("%w: decode native keychain metadata", ErrUnavailable)
	}

	credential := Credential{
		AccountID: metadata.AccountID,
		Scopes:    append([]string(nil), metadata.Scopes...),
	}
	if metadata.AccessTokenExpiresAt != "" {
		expiresAt, err := parseNativeTime(metadata.AccessTokenExpiresAt)
		if err != nil {
			return Credential{}, err
		}
		credential.AccessTokenExpiresAt = expiresAt
	}

	if metadata.HasAccessToken {
		accessToken, err := s.backend.Get(s.service, nativeAccount(profile, nativeAccessKey))
		if err != nil {
			return Credential{}, mapNativeSecretError(err, nativeAccessKey)
		}
		credential.AccessToken = accessToken
	}
	if metadata.HasRefreshToken {
		refreshToken, err := s.backend.Get(s.service, nativeAccount(profile, nativeRefreshKey))
		if err != nil {
			return Credential{}, mapNativeSecretError(err, nativeRefreshKey)
		}
		credential.RefreshToken = refreshToken
	}

	return credential, nil
}

func (s *NativeStore) Put(ctx context.Context, profile string, credential Credential) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	profile = normalizeProfile(profile)
	credential = cloneCredential(credential)

	if err := s.deleteOptional(profile, nativeMetadataKey); err != nil {
		return err
	}
	if err := s.putOptional(profile, nativeAccessKey, credential.AccessToken); err != nil {
		return err
	}
	if err := s.putOptional(profile, nativeRefreshKey, credential.RefreshToken); err != nil {
		return err
	}

	metadata := nativeMetadata{
		AccountID:            credential.AccountID,
		Scopes:               append([]string(nil), credential.Scopes...),
		HasAccessToken:       credential.AccessToken != "",
		HasRefreshToken:      credential.RefreshToken != "",
		AccessTokenExpiresAt: formatNativeTime(credential),
	}
	body, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("%w: encode native keychain metadata", ErrUnavailable)
	}
	if err := s.backend.Set(s.service, nativeAccount(profile, nativeMetadataKey), string(body)); err != nil {
		return mapNativePutError(err)
	}
	return nil
}

func (s *NativeStore) Delete(ctx context.Context, profile string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	profile = normalizeProfile(profile)

	var firstErr error
	for _, field := range []string{nativeAccessKey, nativeRefreshKey, nativeMetadataKey} {
		if err := s.backend.Delete(s.service, nativeAccount(profile, field)); err != nil && !errors.Is(err, keyring.ErrNotFound) {
			if firstErr == nil {
				firstErr = mapNativeDeleteError(err)
			}
		}
	}
	return firstErr
}

func (s *NativeStore) putOptional(profile string, field string, value string) error {
	account := nativeAccount(profile, field)
	if value == "" {
		return s.deleteOptional(profile, field)
	}
	if err := s.backend.Set(s.service, account, value); err != nil {
		return mapNativePutError(err)
	}
	return nil
}

func (s *NativeStore) deleteOptional(profile string, field string) error {
	err := s.backend.Delete(s.service, nativeAccount(profile, field))
	if err == nil || errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return mapNativeDeleteError(err)
}

func normalizeProfile(profile string) string {
	if strings.TrimSpace(profile) == "" {
		return "default"
	}
	return profile
}

func nativeAccount(profile string, field string) string {
	return "profile:" + profile + ":" + field
}

func formatNativeTime(credential Credential) string {
	if credential.AccessTokenExpiresAt.IsZero() {
		return ""
	}
	return credential.AccessTokenExpiresAt.UTC().Format(time.RFC3339Nano)
}

func parseNativeTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: decode native keychain metadata", ErrUnavailable)
	}
	return parsed, nil
}

func mapNativeGetError(err error) error {
	if errors.Is(err, keyring.ErrNotFound) {
		return ErrNotFound
	}
	return fmt.Errorf("%w: native keychain unavailable", ErrUnavailable)
}

func mapNativeSecretError(err error, field string) error {
	if errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("%w: native keychain missing %s", ErrUnavailable, field)
	}
	return fmt.Errorf("%w: native keychain unavailable", ErrUnavailable)
}

func mapNativePutError(err error) error {
	if errors.Is(err, keyring.ErrSetDataTooBig) {
		return fmt.Errorf("%w: native keychain item is too large", ErrUnavailable)
	}
	return fmt.Errorf("%w: native keychain write failed", ErrUnavailable)
}

func mapNativeDeleteError(err error) error {
	return fmt.Errorf("%w: native keychain delete failed", ErrUnavailable)
}
