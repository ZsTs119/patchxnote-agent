package keychain

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNativeStoreIntegration(t *testing.T) {
	if os.Getenv("PATCHNOTE_KEYCHAIN_INTEGRATION") != "1" {
		t.Skip("set PATCHNOTE_KEYCHAIN_INTEGRATION=1 to exercise the OS-native keychain")
	}

	ctx := context.Background()
	profile := "ci-" + time.Now().UTC().Format("20060102150405.000000000")
	store := NewNativeStore()
	credential := Credential{
		AccountID:            "acct_keychain_ci",
		AccessToken:          strings.Repeat("i", 64),
		RefreshToken:         strings.Repeat("j", 64),
		AccessTokenExpiresAt: time.Now().UTC().Add(time.Hour).Truncate(time.Second),
		Scopes:               []string{"agent:account.read", "agent:quota.read"},
	}
	defer func() {
		if err := store.Delete(ctx, profile); err != nil {
			t.Logf("cleanup native keychain profile %q: %v", profile, err)
		}
	}()

	if err := store.Put(ctx, profile, credential); err != nil {
		t.Fatalf("put native keychain credential: %v", err)
	}
	got, err := store.Get(ctx, profile)
	if err != nil {
		t.Fatalf("get native keychain credential: %v", err)
	}
	if got.AccountID != credential.AccountID || got.AccessToken != credential.AccessToken || got.RefreshToken != credential.RefreshToken {
		t.Fatalf("unexpected native keychain credential metadata: %+v", got)
	}
	if len(got.Scopes) != len(credential.Scopes) || got.Scopes[0] != credential.Scopes[0] {
		t.Fatalf("unexpected native keychain scopes: %+v", got.Scopes)
	}

	if err := store.Delete(ctx, profile); err != nil {
		t.Fatalf("delete native keychain credential: %v", err)
	}
	if _, err := store.Get(ctx, profile); !IsNotFound(err) {
		t.Fatalf("expected deleted native keychain credential to be missing, got %v", err)
	}
}
