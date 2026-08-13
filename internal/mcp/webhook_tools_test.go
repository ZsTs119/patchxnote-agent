package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/auth"
	"github.com/ZsTs119/patchxnote-agent/internal/config"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
	"github.com/ZsTs119/patchxnote-agent/internal/webhook"
)

func TestWebhookToolsConfigureListSendAndRemove(t *testing.T) {
	var calls int
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode provider payload: %v", err)
		}
		if payload["markdown"] == "" {
			t.Fatalf("provider received empty markdown: %+v", payload)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer provider.Close()

	rawURL := provider.URL + "/webhook/token-fixture-secret"
	server := newWebhookTestServer(t, false)

	configure := callWebhookTool(t, server, "patchxnote_configure_webhook_target", map[string]any{
		"alias":       "产品群 飞书",
		"type":        "generic",
		"webhook_url": rawURL,
		"template":    "default",
	})
	if configure["error"] != nil {
		t.Fatalf("configure error: %+v", configure["error"])
	}
	text := toolTextResult(t, configure)
	if strings.Contains(text, rawURL) || !strings.Contains(text, `"secret_status": "ok"`) || !strings.Contains(text, "toke...cret") {
		t.Fatalf("configure response leaked raw URL or missed masked metadata:\n%s", text)
	}

	list := callWebhookTool(t, server, "patchxnote_list_webhook_targets", map[string]any{})
	if list["error"] != nil {
		t.Fatalf("list error: %+v", list["error"])
	}
	listText := toolTextResult(t, list)
	if strings.Contains(listText, rawURL) || !strings.Contains(listText, `"alias": "产品群 飞书"`) {
		t.Fatalf("list response invalid:\n%s", listText)
	}

	updateTemplate := callWebhookTool(t, server, "patchxnote_configure_webhook_target", map[string]any{
		"alias":    "产品群 飞书",
		"template": "daily-review",
	})
	if updateTemplate["error"] != nil {
		t.Fatalf("template-only update error: %+v", updateTemplate["error"])
	}
	if text := toolTextResult(t, updateTemplate); strings.Contains(text, rawURL) || !strings.Contains(text, `"template": "daily-review"`) || !strings.Contains(text, `"secret_status": "ok"`) {
		t.Fatalf("template-only update did not preserve masked metadata:\n%s", text)
	}

	templates := callWebhookTool(t, server, "patchxnote_list_webhook_templates", map[string]any{})
	if templates["error"] != nil {
		t.Fatalf("templates error: %+v", templates["error"])
	}
	if !strings.Contains(toolTextResult(t, templates), `"name": "meeting-summary"`) {
		t.Fatalf("expected built-in templates, got:\n%s", toolTextResult(t, templates))
	}

	send := callWebhookTool(t, server, "patchxnote_send_webhook", map[string]any{
		"target_aliases": []string{"产品群 飞书"},
		"title":          "Webhook MCP Test",
		"markdown":       "# Webhook MCP Test\n\nhello\n",
	})
	if send["error"] != nil {
		t.Fatalf("send error: %+v", send["error"])
	}
	if calls != 1 {
		t.Fatalf("expected one provider call, got %d", calls)
	}
	if !strings.Contains(toolTextResult(t, send), `"success": true`) {
		t.Fatalf("unexpected send response:\n%s", toolTextResult(t, send))
	}

	duplicate := callWebhookTool(t, server, "patchxnote_send_webhook", map[string]any{
		"target_aliases": []string{"产品群 飞书", "产品群 飞书"},
		"markdown":       "hello",
	})
	assertRPCError(t, duplicate, codeInvalidParams, "invalid_params")
	if calls != 1 {
		t.Fatalf("duplicate target should fail before provider call, got %d calls", calls)
	}

	removed := callWebhookTool(t, server, "patchxnote_remove_webhook_target", map[string]any{"alias": "产品群 飞书"})
	if removed["error"] != nil {
		t.Fatalf("remove error: %+v", removed["error"])
	}
	empty := callWebhookTool(t, server, "patchxnote_list_webhook_targets", map[string]any{})
	if strings.TrimSpace(toolTextResult(t, empty)) != "[]" {
		t.Fatalf("expected empty target list, got:\n%s", toolTextResult(t, empty))
	}
}

func TestWebhookToolsDisabledAndMissingSecretErrorsAreUserFacing(t *testing.T) {
	server := newWebhookTestServer(t, false)
	configure := callWebhookTool(t, server, "patchxnote_configure_webhook_target", map[string]any{
		"alias":       "运营群 钉钉",
		"type":        "generic",
		"webhook_url": "https://example.invalid/webhook/token-fixture-secret",
		"enabled":     false,
	})
	if configure["error"] != nil {
		t.Fatalf("configure error: %+v", configure["error"])
	}
	send := callWebhookTool(t, server, "patchxnote_send_webhook", map[string]any{
		"target_aliases": []string{"运营群 钉钉"},
		"markdown":       "hello",
	})
	assertRPCError(t, send, codeToolError, "tool_error")

	enabled := callWebhookTool(t, server, "patchxnote_configure_webhook_target", map[string]any{
		"alias":       "产品群 飞书",
		"type":        "generic",
		"webhook_url": "https://example.invalid/webhook/token-fixture-secret",
	})
	if enabled["error"] != nil {
		t.Fatalf("configure enabled error: %+v", enabled["error"])
	}
	secrets := webhook.NewSecretStore(server.secrets, server.config.Profile)
	if err := secrets.DeleteTarget(context.Background(), "产品群 飞书"); err != nil {
		t.Fatalf("delete stored secret: %v", err)
	}
	missingSecret := callWebhookTool(t, server, "patchxnote_send_webhook", map[string]any{
		"target_aliases": []string{"产品群 飞书"},
		"markdown":       "hello",
	})
	assertRPCError(t, missingSecret, codeToolError, "webhook_secret_missing")
}

func TestWebhookToolsRenderAndExportModelIO(t *testing.T) {
	server := newWebhookTestServer(t, true)
	outDir := filepath.Join(t.TempDir(), "draft")
	rendered := callWebhookTool(t, server, "patchxnote_render_webhook_message", map[string]any{
		"memory_id":        "mem_fixture_1",
		"platform":         "mobile",
		"template":         "default",
		"save_draft":       true,
		"out":              outDir,
		"include_model_io": true,
	})
	if rendered["error"] != nil {
		t.Fatalf("render error: %+v", rendered["error"])
	}
	text := toolTextResult(t, rendered)
	if !strings.Contains(text, `"draft_path"`) || !strings.Contains(text, `"memory"`) {
		t.Fatalf("unexpected render response:\n%s", text)
	}
	for _, name := range []string{"source.json", "message.md", "metadata.json", "model-io.json"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("expected draft file %s: %v", name, err)
		}
	}

	outFile := filepath.Join(t.TempDir(), "model-io.json")
	exported := callWebhookTool(t, server, "patchxnote_export_model_io", map[string]any{
		"memory_id": "mem_fixture_1",
		"platform":  "mobile",
		"out":       outFile,
	})
	if exported["error"] != nil {
		t.Fatalf("export error: %+v", exported["error"])
	}
	if !strings.Contains(toolTextResult(t, exported), `"exported": true`) {
		t.Fatalf("unexpected export response:\n%s", toolTextResult(t, exported))
	}
	if _, err := os.Stat(outFile); err != nil {
		t.Fatalf("expected model IO export file: %v", err)
	}
}

func TestWebhookRenderLargeOutputUsesDraftPathInsteadOfInlineMarkdown(t *testing.T) {
	server := newWebhookTestServerWithAPI(t, true, &largeDeliveryAPI{})
	noDraft := callWebhookTool(t, server, "patchxnote_render_webhook_message", map[string]any{
		"memory_id": "mem_fixture_1",
		"platform":  "mobile",
		"template":  "raw-markdown",
	})
	assertRPCError(t, noDraft, codeToolError, "output_too_large")

	outDir := filepath.Join(t.TempDir(), "large-draft")
	withDraft := callWebhookTool(t, server, "patchxnote_render_webhook_message", map[string]any{
		"memory_id":  "mem_fixture_1",
		"platform":   "mobile",
		"template":   "raw-markdown",
		"save_draft": true,
		"out":        outDir,
	})
	if withDraft["error"] != nil {
		t.Fatalf("render with draft error: %+v", withDraft["error"])
	}
	text := toolTextResult(t, withDraft)
	if !strings.Contains(text, `"draft_path"`) {
		t.Fatalf("expected draft path for large render:\n%s", text)
	}
	if strings.Contains(text, strings.Repeat("x", 2048)) {
		t.Fatalf("large Markdown should not be inlined in MCP response")
	}
	if _, err := os.Stat(filepath.Join(outDir, "message.md")); err != nil {
		t.Fatalf("expected large draft message file: %v", err)
	}
}

func TestWebhookMemoryToolsRequireAuth(t *testing.T) {
	server := newWebhookTestServer(t, false)
	rendered := callWebhookTool(t, server, "patchxnote_render_webhook_message", map[string]any{
		"memory_id": "mem_fixture_1",
		"platform":  "mobile",
	})
	assertRPCError(t, rendered, codeAuthRequired, "auth_required")
}

func newWebhookTestServer(t *testing.T, authenticated bool) *Server {
	return newWebhookTestServerWithAPI(t, authenticated, &fakeToolAPI{})
}

func newWebhookTestServerWithAPI(t *testing.T, authenticated bool, agentAPI AgentAPI) *Server {
	t.Helper()
	tempDir := t.TempDir()
	store := keychain.NewMemoryStore()
	var authenticator Authenticator = staticAuthenticator{}
	var credentials CredentialProvider
	if authenticated {
		if err := store.Put(context.Background(), "default", keychain.Credential{
			AccountID:    "acct_fixture",
			AccessToken:  strings.Repeat("z", 32),
			RefreshToken: strings.Repeat("y", 43),
			Scopes: []string{
				"agent:content.read:mobile",
				"agent:content.read:desktop",
			},
		}); err != nil {
			t.Fatalf("seed credential: %v", err)
		}
		manager := auth.NewManager(store, "default")
		authenticator = manager
		credentials = manager
	}
	return NewServer(Options{
		Authenticator: authenticator,
		Credentials:   credentials,
		API:           agentAPI,
		Config: config.Config{
			Profile: "default",
			Paths: config.Paths{
				ConfigFile: filepath.Join(tempDir, "config.yaml"),
			},
		},
		Secrets: store,
	})
}

type largeDeliveryAPI struct {
	fakeToolAPI
}

func (f *largeDeliveryAPI) GetMemoryDeliveryDocument(ctx context.Context, accessToken string, platform string, memoryID string) (api.AgentDeliveryDocument, error) {
	delivery := toolFixtureDeliveryDocument()
	delivery.Markdown = strings.Repeat("x", maxToolOutputBytes+2048)
	return delivery, nil
}

func callWebhookTool(t *testing.T, server *Server, name string, arguments map[string]any) map[string]any {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      name,
			"arguments": arguments,
		},
	})
	if err != nil {
		t.Fatalf("marshal call: %v", err)
	}
	responses := serveCustomForTest(t, server, string(body))
	if len(responses) != 1 {
		t.Fatalf("expected one response, got %d", len(responses))
	}
	return responses[0]
}
