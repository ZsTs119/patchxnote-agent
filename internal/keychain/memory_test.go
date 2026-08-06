package keychain

import (
	"context"
	"errors"
	"testing"
)

func TestMemoryStorePutGetDelete(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	credential := Credential{
		AccountID:    "acct_test",
		AccessToken:  "access-material",
		RefreshToken: "refresh-material",
		Scopes:       []string{"agent:account.read"},
	}

	if _, err := store.Get(ctx, "default"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing credential, got %v", err)
	}
	if err := store.Put(ctx, "default", credential); err != nil {
		t.Fatalf("put credential: %v", err)
	}

	got, err := store.Get(ctx, "default")
	if err != nil {
		t.Fatalf("get credential: %v", err)
	}
	if got.AccountID != credential.AccountID {
		t.Fatalf("expected account %q, got %q", credential.AccountID, got.AccountID)
	}

	got.Scopes[0] = "changed"
	again, err := store.Get(ctx, "default")
	if err != nil {
		t.Fatalf("get credential again: %v", err)
	}
	if again.Scopes[0] != "agent:account.read" {
		t.Fatal("expected memory store to protect stored scope slice from caller mutation")
	}

	if err := store.Delete(ctx, "default"); err != nil {
		t.Fatalf("delete credential: %v", err)
	}
	if _, err := store.Get(ctx, "default"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing credential after delete, got %v", err)
	}
}
