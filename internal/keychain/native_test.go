package keychain

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/zalando/go-keyring"
)

type fakeNativeBackend struct {
	values map[string]string
	err    error
}

func newFakeNativeBackend() *fakeNativeBackend {
	return &fakeNativeBackend{values: make(map[string]string)}
}

func (b *fakeNativeBackend) Get(service, user string) (string, error) {
	if b.err != nil {
		return "", b.err
	}
	value, ok := b.values[service+"\x00"+user]
	if !ok {
		return "", keyring.ErrNotFound
	}
	return value, nil
}

func (b *fakeNativeBackend) Set(service, user, secret string) error {
	if b.err != nil {
		return b.err
	}
	b.values[service+"\x00"+user] = secret
	return nil
}

func (b *fakeNativeBackend) Delete(service, user string) error {
	if b.err != nil {
		return b.err
	}
	key := service + "\x00" + user
	if _, ok := b.values[key]; !ok {
		return keyring.ErrNotFound
	}
	delete(b.values, key)
	return nil
}

func TestNativeStorePutGetDelete(t *testing.T) {
	ctx := context.Background()
	backend := newFakeNativeBackend()
	store := newNativeStore(backend, "patchxnote-agent-test")
	expiresAt := time.Date(2026, 8, 7, 12, 30, 0, 123, time.UTC)
	refreshExpiresAt := expiresAt.Add(30 * 24 * time.Hour)
	credential := Credential{
		AccountID:             "acct_test",
		AccessToken:           strings.Repeat("a", 32),
		RefreshToken:          strings.Repeat("r", 43),
		AccessTokenExpiresAt:  expiresAt,
		RefreshTokenExpiresAt: refreshExpiresAt,
		Scopes:                []string{"agent:account.read", "agent:quota.read"},
	}

	if err := store.Put(ctx, "default", credential); err != nil {
		t.Fatalf("put native credential: %v", err)
	}

	got, err := store.Get(ctx, "default")
	if err != nil {
		t.Fatalf("get native credential: %v", err)
	}
	if got.AccountID != credential.AccountID || got.AccessToken != credential.AccessToken || got.RefreshToken != credential.RefreshToken {
		t.Fatalf("unexpected credential account=%q has_access=%v has_refresh=%v",
			got.AccountID, got.AccessToken != "", got.RefreshToken != "")
	}
	if !got.AccessTokenExpiresAt.Equal(expiresAt) {
		t.Fatalf("expected expires_at %s, got %s", expiresAt, got.AccessTokenExpiresAt)
	}
	if !got.RefreshTokenExpiresAt.Equal(refreshExpiresAt) {
		t.Fatalf("expected refresh_expires_at %s, got %s", refreshExpiresAt, got.RefreshTokenExpiresAt)
	}
	got.Scopes[0] = "changed"
	again, err := store.Get(ctx, "default")
	if err != nil {
		t.Fatalf("get native credential again: %v", err)
	}
	if again.Scopes[0] != "agent:account.read" {
		t.Fatal("expected native store to protect stored scope slice from caller mutation")
	}

	if err := store.Delete(ctx, "default"); err != nil {
		t.Fatalf("delete native credential: %v", err)
	}
	if _, err := store.Get(ctx, "default"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing native credential after delete, got %v", err)
	}
}

func TestNativeStoreSplitsTokenMaterialFromMetadata(t *testing.T) {
	ctx := context.Background()
	backend := newFakeNativeBackend()
	store := newNativeStore(backend, "patchxnote-agent-test")
	credential := Credential{
		AccountID:    "acct_test",
		AccessToken:  strings.Repeat("a", 32),
		RefreshToken: strings.Repeat("r", 43),
		Scopes:       []string{"agent:account.read"},
	}

	if err := store.Put(ctx, "dev", credential); err != nil {
		t.Fatalf("put native credential: %v", err)
	}

	metadata := backend.values["patchxnote-agent-test\x00profile:dev:metadata"]
	if strings.Contains(metadata, credential.AccessToken) || strings.Contains(metadata, credential.RefreshToken) {
		t.Fatalf("metadata contains token material: %s", metadata)
	}
	if backend.values["patchxnote-agent-test\x00profile:dev:access_token"] != credential.AccessToken {
		t.Fatal("expected access token to be stored in its own keychain item")
	}
	if backend.values["patchxnote-agent-test\x00profile:dev:refresh_token"] != credential.RefreshToken {
		t.Fatal("expected refresh token to be stored in its own keychain item")
	}
}

func TestNativeStoreReadsAndMigratesLegacyService(t *testing.T) {
	ctx := context.Background()
	backend := newFakeNativeBackend()
	legacyStore := newNativeStore(backend, "patchnote-agent-test")
	credential := Credential{
		AccountID:    "acct_test",
		AccessToken:  strings.Repeat("a", 32),
		RefreshToken: strings.Repeat("r", 43),
		Scopes:       []string{"agent:account.read"},
	}
	if err := legacyStore.Put(ctx, "default", credential); err != nil {
		t.Fatalf("put legacy native credential: %v", err)
	}

	store := newNativeStoreWithLegacy(backend, "patchxnote-agent-test", "patchnote-agent-test")
	got, err := store.Get(ctx, "default")
	if err != nil {
		t.Fatalf("get migrated native credential: %v", err)
	}
	if got.AccountID != credential.AccountID || got.AccessToken != credential.AccessToken || got.RefreshToken != credential.RefreshToken {
		t.Fatalf("unexpected migrated credential account=%q has_access=%v has_refresh=%v",
			got.AccountID, got.AccessToken != "", got.RefreshToken != "")
	}
	if backend.values["patchxnote-agent-test\x00profile:default:metadata"] == "" {
		t.Fatal("expected legacy credential to be copied into the PatchXNote keychain service")
	}
}

func TestNativeStoreDeleteRemovesLegacyService(t *testing.T) {
	ctx := context.Background()
	backend := newFakeNativeBackend()
	store := newNativeStoreWithLegacy(backend, "patchxnote-agent-test", "patchnote-agent-test")
	credential := Credential{
		AccountID:    "acct_test",
		AccessToken:  strings.Repeat("a", 32),
		RefreshToken: strings.Repeat("r", 43),
	}
	if err := newNativeStore(backend, "patchxnote-agent-test").Put(ctx, "default", credential); err != nil {
		t.Fatalf("put current native credential: %v", err)
	}
	if err := newNativeStore(backend, "patchnote-agent-test").Put(ctx, "default", credential); err != nil {
		t.Fatalf("put legacy native credential: %v", err)
	}

	if err := store.Delete(ctx, "default"); err != nil {
		t.Fatalf("delete current and legacy native credentials: %v", err)
	}
	if _, ok := backend.values["patchxnote-agent-test\x00profile:default:metadata"]; ok {
		t.Fatal("expected current native credential metadata to be deleted")
	}
	if _, ok := backend.values["patchnote-agent-test\x00profile:default:metadata"]; ok {
		t.Fatal("expected legacy native credential metadata to be deleted")
	}
}

func TestNativeStoreMapsBackendErrorsToUnavailable(t *testing.T) {
	backend := newFakeNativeBackend()
	backend.err = errors.New("backend unavailable")
	store := newNativeStore(backend, "patchxnote-agent-test")

	if _, err := store.Get(context.Background(), "default"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected unavailable get, got %v", err)
	}
	if err := store.Put(context.Background(), "default", Credential{AccountID: "acct_test"}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected unavailable put, got %v", err)
	}
	if err := store.Delete(context.Background(), "default"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected unavailable delete, got %v", err)
	}
}

func TestNativeStoreDetectsIncompleteCredential(t *testing.T) {
	ctx := context.Background()
	backend := newFakeNativeBackend()
	store := newNativeStore(backend, "patchxnote-agent-test")

	if err := store.Put(ctx, "default", Credential{
		AccountID:    "acct_test",
		AccessToken:  strings.Repeat("a", 32),
		RefreshToken: strings.Repeat("r", 43),
	}); err != nil {
		t.Fatalf("put native credential: %v", err)
	}
	delete(backend.values, "patchxnote-agent-test\x00profile:default:access_token")

	if _, err := store.Get(ctx, "default"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected unavailable incomplete credential, got %v", err)
	}
}
