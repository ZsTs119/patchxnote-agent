package cli

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/config"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
	"github.com/ZsTs119/patchxnote-agent/internal/webhook"
)

func TestWebhookSetListShowRemoveSecretFree(t *testing.T) {
	store := keychain.NewMemoryStore()
	deps := webhookTestDeps(t, store, nil, nil)
	rawURL := "https://example.test/hook/abcdefghijklmnopqrstuvwxyz123456"

	stdout, stderr, err := executeForTestWithDeps(t, deps, "webhook", "set", "产品群 飞书", "--type", "feishu", "--url", rawURL, "--secret", "signing-secret-material")
	if err != nil {
		t.Fatalf("webhook set: %v", err)
	}
	assertNoWebhookSecretLeak(t, stdout+stderr, rawURL, "signing-secret-material")
	if !strings.Contains(stdout, "产品群 飞书") || !strings.Contains(stdout, "feishu") {
		t.Fatalf("unexpected set output:\n%s", stdout)
	}

	stdout, stderr, err = executeForTestWithDeps(t, deps, "webhook", "list")
	if err != nil {
		t.Fatalf("webhook list: %v", err)
	}
	assertNoWebhookSecretLeak(t, stdout+stderr, rawURL, "signing-secret-material")
	if !strings.Contains(stdout, "secret=ok") {
		t.Fatalf("expected secret ok in list, got:\n%s", stdout)
	}

	stdout, stderr, err = executeForTestWithDeps(t, deps, "webhook", "show", "产品群 飞书")
	if err != nil {
		t.Fatalf("webhook show: %v", err)
	}
	assertNoWebhookSecretLeak(t, stdout+stderr, rawURL, "signing-secret-material")
	if !strings.Contains(stdout, "url https://example.test") {
		t.Fatalf("unexpected show output:\n%s", stdout)
	}

	stdout, _, err = executeForTestWithDeps(t, deps, "webhook", "remove", "产品群 飞书")
	if err != nil {
		t.Fatalf("webhook remove: %v", err)
	}
	if !strings.Contains(stdout, "removed") {
		t.Fatalf("unexpected remove output:\n%s", stdout)
	}
	secrets := webhook.NewSecretStore(store, "default")
	if _, err := secrets.URL(context.Background(), "产品群 飞书"); !errors.Is(err, webhook.ErrWebhookSecretMissing) {
		t.Fatalf("expected URL secret removed, got %v", err)
	}
}

func TestWebhookSetStdinAndClearSecret(t *testing.T) {
	store := keychain.NewMemoryStore()
	rawURL := "https://example.test/hook/abcdefghijklmnopqrstuvwxyz123456"
	deps := webhookTestDeps(t, store, strings.NewReader(rawURL+"\nsigning-secret-material\n"), nil)
	if _, _, err := executeForTestWithDeps(t, deps, "webhook", "set", "运营群 钉钉", "--type", "dingtalk", "--url-stdin", "--secret-stdin"); err != nil {
		t.Fatalf("webhook set stdin: %v", err)
	}
	secrets := webhook.NewSecretStore(store, "default")
	if got, ok, err := secrets.SigningSecret(context.Background(), "运营群 钉钉"); err != nil || !ok || got == "" {
		t.Fatalf("expected signing secret, ok=%v err=%v", ok, err)
	}

	if _, _, err := executeForTestWithDeps(t, webhookTestDeps(t, store, nil, nil), "webhook", "set", "运营群 钉钉", "--clear-secret"); err != nil {
		t.Fatalf("clear secret: %v", err)
	}
	if _, ok, err := secrets.SigningSecret(context.Background(), "运营群 钉钉"); err != nil || ok {
		t.Fatalf("expected signing secret cleared, ok=%v err=%v", ok, err)
	}
	if got, err := secrets.URL(context.Background(), "运营群 钉钉"); err != nil || got != rawURL {
		t.Fatalf("expected URL preserved, got %q err=%v", got, err)
	}
}

func TestWebhookTestAndSendFile(t *testing.T) {
	var received []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode webhook request: %v", err)
		}
		received = append(received, payload)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	store := keychain.NewMemoryStore()
	deps := webhookTestDeps(t, store, nil, nil)
	if _, _, err := executeForTestWithDeps(t, deps, "webhook", "set", "内部网关", "--type", "generic", "--url", server.URL); err != nil {
		t.Fatalf("set generic: %v", err)
	}
	if stdout, _, err := executeForTestWithDeps(t, deps, "webhook", "test", "内部网关"); err != nil || !strings.Contains(stdout, "sent 内部网关") {
		t.Fatalf("webhook test stdout=%q err=%v", stdout, err)
	}
	file := filepath.Join(t.TempDir(), "本周复盘.md")
	if err := os.WriteFile(file, []byte("# 本周复盘\n\n## 摘要\n\n合成内容\n"), 0o600); err != nil {
		t.Fatalf("write markdown: %v", err)
	}
	if stdout, _, err := executeForTestWithDeps(t, deps, "webhook", "send", "--target", "内部网关", "--file", file); err != nil || !strings.Contains(stdout, "sent 内部网关") {
		t.Fatalf("webhook send file stdout=%q err=%v", stdout, err)
	}
	if len(received) != 2 {
		t.Fatalf("expected test and send requests, got %d", len(received))
	}
	if received[1]["title"] != "本周复盘" || !strings.Contains(received[1]["markdown"].(string), "合成内容") {
		t.Fatalf("unexpected send payload: %+v", received[1])
	}
}

func TestWebhookSendDraftAndDuplicateTarget(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if strings.Contains(stringFromAny(payload["markdown"]), "model_io") {
			t.Fatal("draft send should not include model IO content")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	store := keychain.NewMemoryStore()
	deps := webhookTestDeps(t, store, nil, nil)
	if _, _, err := executeForTestWithDeps(t, deps, "webhook", "set", "内部网关", "--type", "generic", "--url", server.URL); err != nil {
		t.Fatalf("set generic: %v", err)
	}
	draft := t.TempDir()
	if err := os.WriteFile(filepath.Join(draft, "message.md"), []byte("# 草稿标题\n\n草稿内容\n"), 0o600); err != nil {
		t.Fatalf("write draft message: %v", err)
	}
	if err := os.WriteFile(filepath.Join(draft, "metadata.json"), []byte(`{"source":"patchxnote","memory_id":"mem_fixture","platform":"desktop"}`), 0o600); err != nil {
		t.Fatalf("write draft metadata: %v", err)
	}
	if stdout, _, err := executeForTestWithDeps(t, deps, "webhook", "send", "--target", "内部网关", "--draft", draft); err != nil || !strings.Contains(stdout, "sent 内部网关") {
		t.Fatalf("send draft stdout=%q err=%v", stdout, err)
	}
	if calls != 1 {
		t.Fatalf("expected one draft send call, got %d", calls)
	}
	if _, _, err := executeForTestWithDeps(t, deps, "webhook", "send", "--target", "内部网关", "--target", "内部网关", "--draft", draft); err == nil {
		t.Fatal("expected duplicate target to fail")
	}
	if calls != 1 {
		t.Fatalf("duplicate target should fail before request, got %d calls", calls)
	}
}

func TestWebhookDraftAndExportModelIO(t *testing.T) {
	store := keychain.NewMemoryStore()
	seedCredential(t, store)
	deps := webhookTestDeps(t, store, nil, &fakeAgentAPI{
		delivery: api.AgentDeliveryDocument{
			Source:  "patchxnote",
			Version: "1",
			Title:   "合成会议纪要",
			Summary: "合成摘要",
			Memory:  &api.AgentDeliveryMemory{ID: "mem_fixture", Platform: "desktop"},
			Trace:   api.AgentDeliveryTrace{RequestID: "mrun_fixture", Platform: "desktop", State: "completed"},
		},
		modelIO: api.AgentModelIOExport{
			Source:           "patchxnote",
			Version:          "1",
			Memory:           &api.AgentDeliveryMemory{ID: "mem_fixture", Platform: "desktop"},
			Trace:            api.AgentDeliveryTrace{RequestID: "mrun_fixture", Platform: "desktop", State: "completed"},
			FieldStatus:      api.AgentModelIOFieldStatus{ClientRequestJSON: "available"},
			SourceText:       &api.AgentSourceText{Availability: "available", Text: "合成文本"},
			ParsedResultJSON: []byte(`{"title":"合成会议纪要"}`),
		},
	})
	out := filepath.Join(t.TempDir(), "draft")
	if stdout, _, err := executeForTestWithDeps(t, deps, "webhook", "draft", "--memory-id", "mem_fixture", "--platform", "desktop", "--out", out, "--include-model-io"); err != nil || !strings.Contains(stdout, "draft written") {
		t.Fatalf("draft stdout=%q err=%v", stdout, err)
	}
	for _, name := range []string{"source.json", "message.md", "metadata.json", "model-io.json"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatalf("expected draft file %s: %v", name, err)
		}
	}
	exportPath := filepath.Join(t.TempDir(), "model-io.json")
	stdout, stderr, err := executeForTestWithDeps(t, deps, "webhook", "export-model-io", "--memory-id", "mem_fixture", "--platform", "desktop", "--out", exportPath)
	if err != nil {
		t.Fatalf("export model io: %v", err)
	}
	if strings.Contains(stdout+stderr, "合成文本") {
		t.Fatalf("export command output leaked raw model IO: stdout=%q stderr=%q", stdout, stderr)
	}
	if _, err := os.Stat(exportPath); err != nil {
		t.Fatalf("expected export file: %v", err)
	}
}

func TestWebhookMemoryReadRefreshesOnUnauthorized(t *testing.T) {
	store := keychain.NewMemoryStore()
	now := time.Now()
	if err := store.Put(context.Background(), "default", keychain.Credential{
		AccountID:             "acct_cached",
		AccessToken:           "stale-access",
		RefreshToken:          strings.Repeat("r", 43),
		AccessTokenExpiresAt:  now.Add(time.Hour),
		RefreshTokenExpiresAt: now.Add(30 * 24 * time.Hour),
	}); err != nil {
		t.Fatalf("seed credential: %v", err)
	}
	api := &unauthorizedThenDeliveryAPI{}
	deps := webhookTestDeps(t, store, nil, api)
	out := filepath.Join(t.TempDir(), "draft")
	if _, _, err := executeForTestWithDeps(t, deps, "webhook", "draft", "--memory-id", "mem_fixture", "--out", out); err != nil {
		t.Fatalf("draft after refresh: %v", err)
	}
	if api.refreshCalls != 1 || api.deliveryCalls != 2 {
		t.Fatalf("expected one refresh and two delivery calls, got refresh=%d delivery=%d", api.refreshCalls, api.deliveryCalls)
	}
}

func webhookTestDeps(t *testing.T, store keychain.Store, stdin *strings.Reader, agent agentAPI) Deps {
	t.Helper()
	home := t.TempDir()
	if testHome, ok := webhookTestHomeFromStore.LoadOrStore(store, home); ok {
		home = testHome.(string)
	}
	if agent == nil {
		agent = &fakeAgentAPI{}
	}
	return Deps{
		TargetOS:        "linux",
		PathEnv:         config.PathEnv{HomeDir: home},
		CredentialStore: store,
		Stdin:           stdin,
		APIFactory: func(cfg config.Config) (agentAPI, error) {
			return agent, nil
		},
	}
}

var webhookTestHomeFromStore syncMap

type syncMap struct {
	items sync.Map
}

func (m *syncMap) LoadOrStore(key any, value any) (actual any, loaded bool) {
	return m.items.LoadOrStore(key, value)
}

func seedCredential(t *testing.T, store *keychain.MemoryStore) {
	t.Helper()
	if err := store.Put(context.Background(), "default", keychain.Credential{
		AccountID:             "acct_fixture",
		AccessToken:           strings.Repeat("a", 32),
		RefreshToken:          strings.Repeat("r", 43),
		AccessTokenExpiresAt:  time.Now().Add(time.Hour),
		RefreshTokenExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}); err != nil {
		t.Fatalf("seed credential: %v", err)
	}
}

func assertNoWebhookSecretLeak(t *testing.T, text string, values ...string) {
	t.Helper()
	for _, value := range values {
		if value != "" && strings.Contains(text, value) {
			t.Fatalf("output leaked secret value %q in:\n%s", value, text)
		}
	}
}

func stringFromAny(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

type unauthorizedThenDeliveryAPI struct {
	fakeAgentAPI
	deliveryCalls int
	refreshCalls  int
}

func (f *unauthorizedThenDeliveryAPI) GetMemoryDeliveryDocument(ctx context.Context, accessToken string, platform string, memoryID string) (api.AgentDeliveryDocument, error) {
	f.deliveryCalls++
	if f.deliveryCalls == 1 {
		return api.AgentDeliveryDocument{}, &api.Error{StatusCode: http.StatusUnauthorized, Code: "auth_required"}
	}
	return api.AgentDeliveryDocument{
		Source:   "patchxnote",
		Version:  "1",
		Title:    "刷新后文档",
		Markdown: "# 刷新后文档\n\n正文",
		Memory:   &api.AgentDeliveryMemory{ID: memoryID, Platform: platform},
		Trace:    api.AgentDeliveryTrace{RequestID: "mrun_after_refresh", Platform: platform, State: "completed"},
	}, nil
}

func (f *unauthorizedThenDeliveryAPI) RefreshAgentSession(ctx context.Context, request api.AgentRefreshRequest, idempotencyKey string) (api.AgentSessionResponse, error) {
	f.refreshCalls++
	return f.fakeAgentAPI.RefreshAgentSession(ctx, request, idempotencyKey)
}
