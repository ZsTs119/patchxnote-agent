package auth

import (
	"context"
	"testing"

	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
)

func TestStatusUnauthenticatedWhenCredentialMissing(t *testing.T) {
	manager := NewManager(keychain.NewMemoryStore(), "default")

	status, err := manager.Status(context.Background())
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.Authenticated {
		t.Fatal("expected unauthenticated status")
	}
	if status.Profile != "default" {
		t.Fatalf("expected default profile, got %q", status.Profile)
	}
}

func TestSaveStatusAndLogoutUseKeychainStore(t *testing.T) {
	ctx := context.Background()
	store := keychain.NewMemoryStore()
	manager := NewManager(store, "dev")

	if err := manager.Save(ctx, keychain.Credential{
		AccountID:    "acct_test",
		AccessToken:  "access-material",
		RefreshToken: "refresh-material",
		Scopes:       []string{"agent:account.read"},
	}); err != nil {
		t.Fatalf("save: %v", err)
	}

	status, err := manager.Status(ctx)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !status.Authenticated {
		t.Fatal("expected authenticated status")
	}
	if status.AccountID != "acct_test" {
		t.Fatalf("expected account id, got %q", status.AccountID)
	}

	if err := manager.Logout(ctx); err != nil {
		t.Fatalf("logout: %v", err)
	}
	status, err = manager.Status(ctx)
	if err != nil {
		t.Fatalf("status after logout: %v", err)
	}
	if status.Authenticated {
		t.Fatal("expected unauthenticated after logout")
	}
}

func TestUnavailableStoreReadsAsUnauthenticatedAndLogoutSucceeds(t *testing.T) {
	manager := NewManager(keychain.UnavailableStore{}, "default")

	status, err := manager.Status(context.Background())
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.Authenticated {
		t.Fatal("expected unauthenticated status")
	}
	if err := manager.Logout(context.Background()); err != nil {
		t.Fatalf("logout should ignore unavailable store: %v", err)
	}
	if err := manager.Save(context.Background(), keychain.Credential{AccountID: "acct_test"}); err == nil {
		t.Fatal("expected save to fail when store is unavailable")
	}
}
