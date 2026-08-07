package cli

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/auth"
	"github.com/ZsTs119/patchxnote-agent/internal/config"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
)

func TestAuthStatusPlainUnauthenticated(t *testing.T) {
	stdout, stderr, err := executeForTest(t, "auth", "status")
	if err != nil {
		t.Fatalf("auth status: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "unauthenticated") {
		t.Fatalf("expected unauthenticated output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "profile default") {
		t.Fatalf("expected profile output, got:\n%s", stdout)
	}
}

func TestAuthStatusJSONUnauthenticated(t *testing.T) {
	stdout, stderr, err := executeForTest(t, "--output", "json", "auth", "status")
	if err != nil {
		t.Fatalf("auth status json: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}

	var got auth.Status
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("expected auth status JSON: %v\n%s", err, stdout)
	}
	if got.Authenticated {
		t.Fatal("expected unauthenticated JSON status")
	}
	if got.Profile != "default" {
		t.Fatalf("expected default profile, got %q", got.Profile)
	}
}

func TestAuthStatusDoesNotPrintCredentialMaterial(t *testing.T) {
	store := keychain.NewMemoryStore()
	if err := store.Put(context.Background(), "default", keychain.Credential{
		AccountID:    "acct_test",
		AccessToken:  "access-material",
		RefreshToken: "refresh-material",
		Scopes:       []string{"agent:account.read"},
	}); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: store,
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return nil, nil
		},
	}, "auth", "status")
	if err != nil {
		t.Fatalf("auth status: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "authenticated") || !strings.Contains(stdout, "acct_test") {
		t.Fatalf("expected authenticated account output, got:\n%s", stdout)
	}
	for _, disallowed := range []string{"access-material", "refresh-material"} {
		if strings.Contains(stdout, disallowed) {
			t.Fatalf("credential material leaked in auth status output: %q", disallowed)
		}
	}
}

func TestAuthStatusJSONRefreshesExpiredCredentialMetadata(t *testing.T) {
	before := time.Now()
	store := keychain.NewMemoryStore()
	if err := store.Put(context.Background(), "default", keychain.Credential{
		AccountID:             "acct_cached",
		AccessToken:           strings.Repeat("x", 32),
		RefreshToken:          strings.Repeat("y", 43),
		AccessTokenExpiresAt:  before.Add(-time.Minute),
		RefreshTokenExpiresAt: before.Add(24 * time.Hour),
		Scopes:                []string{"agent:account.read"},
	}); err != nil {
		t.Fatalf("seed credential: %v", err)
	}
	fakeAPI := &fakeAgentAPI{}

	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: store,
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return fakeAPI, nil
		},
	}, "--output", "json", "auth", "status")
	if err != nil {
		t.Fatalf("auth status json: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}

	var got auth.Status
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("expected auth status JSON: %v\n%s", err, stdout)
	}
	if fakeAPI.refreshCalls != 1 {
		t.Fatalf("expected one refresh call, got %d", fakeAPI.refreshCalls)
	}
	if !got.Authenticated || got.AccountID != "acct_fixture" {
		t.Fatalf("unexpected auth status: %+v", got)
	}
	if !got.AccessTokenExpiresAt.After(before) {
		t.Fatalf("expected refreshed access expiry after %s, got %s", before, got.AccessTokenExpiresAt)
	}
	if !got.RefreshTokenExpiresAt.After(before.Add(29 * 24 * time.Hour)) {
		t.Fatalf("expected refreshed refresh expiry near 30 days, got %s", got.RefreshTokenExpiresAt)
	}
}

func TestLogoutDeletesLocalCredential(t *testing.T) {
	store := keychain.NewMemoryStore()
	if err := store.Put(context.Background(), "default", keychain.Credential{
		AccountID:    "acct_test",
		AccessToken:  "access-material",
		RefreshToken: "refresh-material",
	}); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	stdout, stderr, err := executeForTestWithDeps(t, Deps{
		CredentialStore: store,
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return nil, nil
		},
	}, "logout")
	if err != nil {
		t.Fatalf("logout: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}
	if !strings.Contains(stdout, "logged out") {
		t.Fatalf("expected logout output, got:\n%s", stdout)
	}
	if _, err := store.Get(context.Background(), "default"); !errors.Is(err, keychain.ErrNotFound) {
		t.Fatalf("expected credential to be deleted, got %v", err)
	}
}
