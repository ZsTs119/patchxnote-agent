package keychain

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileStorePutGetDelete(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "credentials.dev.json")
	store := NewFileStore(path)
	credential := Credential{
		AccountID:    "acct_test",
		AccessToken:  strings.Repeat("a", 32),
		RefreshToken: strings.Repeat("b", 43),
		Scopes:       []string{"agent:account.read"},
	}

	if err := store.Put(ctx, "default", credential); err != nil {
		t.Fatalf("put credential: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat credential file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected credential file mode 0600, got %v", info.Mode().Perm())
	}

	got, err := store.Get(ctx, "default")
	if err != nil {
		t.Fatalf("get credential: %v", err)
	}
	if got.AccountID != credential.AccountID || got.AccessToken != credential.AccessToken || got.RefreshToken != credential.RefreshToken {
		t.Fatalf("unexpected credential account=%q has_access=%v has_refresh=%v",
			got.AccountID, got.AccessToken != "", got.RefreshToken != "")
	}

	if err := store.Delete(ctx, "default"); err != nil {
		t.Fatalf("delete credential: %v", err)
	}
	if _, err := store.Get(ctx, "default"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing credential, got %v", err)
	}
}
