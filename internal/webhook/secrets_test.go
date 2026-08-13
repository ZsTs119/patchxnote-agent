package webhook

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
)

func TestSecretStoreURLAndSigningSecret(t *testing.T) {
	ctx := context.Background()
	store := keychain.NewMemoryStore()
	secrets := NewSecretStore(store, "default")

	rawURL := "https://open.feishu.cn/open-apis/bot/v2/hook/token_fixture_123456"
	if err := secrets.PutURL(ctx, "产品群 飞书", rawURL); err != nil {
		t.Fatalf("put url: %v", err)
	}
	gotURL, err := secrets.URL(ctx, "产品群 飞书")
	if err != nil {
		t.Fatalf("get url: %v", err)
	}
	if gotURL != rawURL {
		t.Fatalf("expected URL round trip")
	}

	if err := secrets.PutSigningSecret(ctx, "产品群 飞书", "signing-secret-material"); err != nil {
		t.Fatalf("put signing secret: %v", err)
	}
	gotSecret, ok, err := secrets.SigningSecret(ctx, "产品群 飞书")
	if err != nil {
		t.Fatalf("get signing secret: %v", err)
	}
	if !ok || gotSecret != "signing-secret-material" {
		t.Fatalf("unexpected signing secret ok=%v value=%q", ok, gotSecret)
	}

	if err := secrets.DeleteSigningSecret(ctx, "产品群 飞书"); err != nil {
		t.Fatalf("clear signing secret: %v", err)
	}
	if _, ok, err := secrets.SigningSecret(ctx, "产品群 飞书"); err != nil || ok {
		t.Fatalf("expected signing secret cleared ok=%v err=%v", ok, err)
	}
	gotURL, err = secrets.URL(ctx, "产品群 飞书")
	if err != nil || gotURL == "" {
		t.Fatalf("URL should survive signing-secret clear, got url=%q err=%v", gotURL, err)
	}
}

func TestSecretStoreProfileIsolationAndAliasEncoding(t *testing.T) {
	ctx := context.Background()
	store := keychain.NewMemoryStore()
	a := NewSecretStore(store, "profile-a")
	b := NewSecretStore(store, "profile-b")

	if err := a.PutURL(ctx, "产品 群, A", "https://example.test/a"); err != nil {
		t.Fatalf("put profile a: %v", err)
	}
	if err := b.PutURL(ctx, "产品 群, A", "https://example.test/b"); err != nil {
		t.Fatalf("put profile b: %v", err)
	}
	gotA, err := a.URL(ctx, "产品 群, A")
	if err != nil {
		t.Fatalf("get profile a: %v", err)
	}
	gotB, err := b.URL(ctx, "产品 群, A")
	if err != nil {
		t.Fatalf("get profile b: %v", err)
	}
	if gotA == gotB {
		t.Fatalf("expected profile-specific secrets, got %q and %q", gotA, gotB)
	}
}

func TestSecretStoreMissingURLIsSafe(t *testing.T) {
	_, err := NewSecretStore(keychain.NewMemoryStore(), "default").URL(context.Background(), "产品群")
	if err == nil {
		t.Fatal("expected missing URL error")
	}
	if !errors.Is(err, ErrWebhookSecretMissing) {
		t.Fatalf("expected ErrWebhookSecretMissing, got %v", err)
	}
	if strings.Contains(err.Error(), "https://") {
		t.Fatalf("missing secret error leaked URL-like value: %v", err)
	}
}

func TestSecretStoreUnavailableFailsClosed(t *testing.T) {
	err := NewSecretStore(keychain.UnavailableStore{Reason: "fixture"}, "default").PutURL(context.Background(), "产品群", "https://example.test/hook")
	if err == nil {
		t.Fatal("expected unavailable secret store error")
	}
	if !keychain.IsUnavailable(err) {
		t.Fatalf("expected keychain unavailable, got %v", err)
	}
}
