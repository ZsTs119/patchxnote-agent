package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMVP(t *testing.T) {
	binary := os.Getenv("PATCHXNOTE_E2E_BINARY")
	if binary == "" {
		t.Skip("PATCHXNOTE_E2E_BINARY is required")
	}

	credentialMaterial := strings.Repeat("a", 32)
	refreshMaterial := strings.Repeat("r", 32)
	var webhookCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/agent/auth/otp/requests":
			if r.Header.Get("Authorization") != "" {
				t.Fatal("otp request must not send authorization")
			}
			writeJSON(t, w, http.StatusAccepted, map[string]any{
				"request_id":       "otp_request_fixture",
				"status":           "accepted",
				"cooldown_seconds": 1,
			})
		case "/v1/agent/auth/otp/verifications":
			writeJSON(t, w, http.StatusOK, map[string]any{
				"access_token":               credentialMaterial,
				"access_expires_in_seconds":  3600,
				"refresh_token":              refreshMaterial,
				"refresh_expires_in_seconds": 2592000,
				"account":                    accountFixture(),
				"scopes": []string{
					"agent:account.read",
					"agent:hardware.read",
					"agent:quota.read",
					"agent:model_usage.read",
					"agent:content.read:mobile",
					"agent:content.read:desktop",
				},
			})
		case "/v1/agent/me":
			requireBearer(t, r, credentialMaterial)
			writeJSON(t, w, http.StatusOK, accountFixture())
		case "/v1/agent/recorder-cards":
			requireBearer(t, r, credentialMaterial)
			writeJSON(t, w, http.StatusOK, map[string]any{"items": []map[string]any{{
				"id":                 "hw_fixture",
				"binding_epoch_id":   "binding_epoch_fixture",
				"identity_masked":    "MR20-****-0000",
				"binding_status":     "active",
				"credential_version": 1,
				"created_at":         "2026-08-06T10:00:00Z",
			}}})
		case "/v1/agent/quota/summary":
			requireBearer(t, r, credentialMaterial)
			writeJSON(t, w, http.StatusOK, map[string]any{
				"available_tokens":            120000,
				"gift_available_tokens":       20000,
				"paid_available_tokens":       90000,
				"adjustment_available_tokens": 10000,
				"calculated_at":               "2026-08-06T10:00:00Z",
			})
		case "/v1/agent/model-usage/summary":
			requireBearer(t, r, credentialMaterial)
			writeJSON(t, w, http.StatusOK, map[string]any{
				"period":                                "current_month",
				"period_starts_at":                      "2026-08-01T00:00:00Z",
				"period_ends_at":                        "2026-09-01T00:00:00Z",
				"provider_total_tokens":                 30000,
				"charged_quota_tokens":                  15000,
				"run_count":                             12,
				"attempt_count":                         14,
				"gift_charged_tokens":                   5000,
				"paid_charged_tokens":                   10000,
				"estimated_platform_savings_microunits": 2300,
				"value_basis":                           "versioned_reference_price",
				"calculated_at":                         "2026-08-06T10:00:00Z",
			})
		case "/v1/agent/memories":
			requireBearer(t, r, credentialMaterial)
			if r.URL.Query().Get("platform") != "mobile" {
				t.Fatalf("unexpected memories platform: %s", r.URL.RawQuery)
			}
			writeJSON(t, w, http.StatusOK, map[string]any{
				"items":       []map[string]any{memoryFixture()},
				"next_cursor": "cursor_page_2",
			})
		case "/v1/agent/memories/mem_fixture_1":
			requireBearer(t, r, credentialMaterial)
			writeJSON(t, w, http.StatusOK, memoryFixture())
		case "/v1/agent/auth/logout":
			requireBearer(t, r, credentialMaterial)
			w.WriteHeader(http.StatusNoContent)
		case "/webhook/generic":
			webhookCalls++
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode webhook payload: %v", err)
			}
			if payload["title"] == "" || payload["markdown"] == "" {
				t.Fatalf("unexpected webhook payload: %+v", payload)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	home := t.TempDir()
	env := append(os.Environ(),
		"HOME="+home,
		"XDG_CONFIG_HOME="+filepath.Join(home, "config"),
		"XDG_CACHE_HOME="+filepath.Join(home, "cache"),
		"PATCHXNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true",
		"PATCHXNOTE_SERVER_BASE_URL="+server.URL,
	)

	runCLI(t, binary, env, "login", "--phone", "+86*******0000", "--code", "000000")
	status := runCLI(t, binary, env, "auth", "status", "--output", "json")
	if !strings.Contains(status.stdout, `"authenticated": true`) || !strings.Contains(status.stdout, `"account_id": "acct_fixture"`) {
		t.Fatalf("unexpected auth status output:\n%s", status.stdout)
	}

	mcpInput := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"patchxnote_get_current_user","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"patchxnote_list_recorder_cards","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"patchxnote_get_quota_summary","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"patchxnote_get_model_usage_summary","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"patchxnote_list_memories","arguments":{"platform":"mobile","limit":20}}}`,
		`{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"patchxnote_search_memories","arguments":{"platform":"mobile","query":"event","limit":20}}}`,
		`{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"patchxnote_get_memory","arguments":{"platform":"mobile","memory_id":"mem_fixture_1"}}}`,
	}, "\n") + "\n"
	mcpResult := runCLIWithInput(t, binary, env, mcpInput, "mcp", "serve")
	assertMCPResponses(t, mcpResult.stdout, 9)

	webhookURL := server.URL + "/webhook/generic"
	runCLI(t, binary, env, "webhook", "set", "内部网关", "--type", "generic", "--url", webhookURL)
	runCLI(t, binary, env, "webhook", "list")
	runCLI(t, binary, env, "webhook", "test", "内部网关")
	messageFile := filepath.Join(home, "message.md")
	if err := os.WriteFile(messageFile, []byte("# 合成周报\n\n合成内容\n"), 0o600); err != nil {
		t.Fatalf("write webhook message file: %v", err)
	}
	runCLI(t, binary, env, "webhook", "send", "--target", "内部网关", "--file", messageFile)
	draftDir := filepath.Join(home, "draft")
	if err := os.MkdirAll(draftDir, 0o700); err != nil {
		t.Fatalf("create draft dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(draftDir, "message.md"), []byte("# 草稿周报\n\n草稿内容\n"), 0o600); err != nil {
		t.Fatalf("write draft message: %v", err)
	}
	if err := os.WriteFile(filepath.Join(draftDir, "metadata.json"), []byte(`{"source":"patchxnote","memory_id":"mem_fixture_1","platform":"mobile"}`), 0o600); err != nil {
		t.Fatalf("write draft metadata: %v", err)
	}
	runCLI(t, binary, env, "webhook", "send", "--target", "内部网关", "--draft", draftDir)
	runCLIExpectError(t, binary, env, "webhook", "send", "--target", "内部网关", "--target", "内部网关", "--file", messageFile)
	if webhookCalls != 3 {
		t.Fatalf("expected webhook test/file/draft calls, got %d", webhookCalls)
	}

	runCLI(t, binary, env, "logout")
	status = runCLI(t, binary, env, "auth", "status")
	if !strings.Contains(status.stdout, "unauthenticated") {
		t.Fatalf("expected unauthenticated after logout, got:\n%s", status.stdout)
	}

	writeEvidence(t, map[string]any{
		"status":      "pass",
		"commands":    []string{"login", "auth status", "mcp serve", "webhook set/list/test/send", "logout", "auth status"},
		"tool_count":  7,
		"server_kind": "httptest-agent-v1",
	})
}

type cliResult struct {
	stdout string
	stderr string
}

func runCLI(t *testing.T, binary string, env []string, args ...string) cliResult {
	t.Helper()
	return runCLIWithInput(t, binary, env, "", args...)
}

func runCLIWithInput(t *testing.T, binary string, env []string, input string, args ...string) cliResult {
	t.Helper()
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = env
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("command %v failed: %v\nstdout:\n%s\nstderr:\n%s", args, err, stdout.String(), stderr.String())
	}
	combined := stdout.String() + stderr.String()
	for _, disallowed := range []string{strings.Repeat("a", 32), strings.Repeat("r", 32), "000000"} {
		if strings.Contains(combined, disallowed) {
			t.Fatalf("command %v leaked sensitive value %q", args, disallowed)
		}
	}
	return cliResult{stdout: stdout.String(), stderr: stderr.String()}
}

func runCLIExpectError(t *testing.T, binary string, env []string, args ...string) cliResult {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), binary, args...)
	cmd.Env = env
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err == nil {
		t.Fatalf("command %v unexpectedly succeeded\nstdout:\n%s\nstderr:\n%s", args, stdout.String(), stderr.String())
	}
	return cliResult{stdout: stdout.String(), stderr: stderr.String()}
}

func assertMCPResponses(t *testing.T, stdout string, want int) {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != want {
		t.Fatalf("expected %d MCP responses, got %d\n%s", want, len(lines), stdout)
	}
	for _, line := range lines {
		var response map[string]any
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			t.Fatalf("MCP stdout line is not JSON: %v\n%s", err, line)
		}
		if response["error"] != nil {
			t.Fatalf("unexpected MCP error response: %+v", response["error"])
		}
	}
}

func accountFixture() map[string]any {
	return map[string]any{
		"id":                    "acct_fixture",
		"status":                "active",
		"registration_platform": "mobile",
		"phone_masked":          "+86*******0000",
		"state_version":         2,
	}
}

func memoryFixture() map[string]any {
	return map[string]any{
		"id":                      "mem_fixture_1",
		"platform":                "mobile",
		"object_type":             "event",
		"client_object_id":        "local_event_fixture",
		"revision_id":             "rev_fixture_1",
		"revision":                3,
		"schema_id":               "patchnote.event.v1",
		"schema_version":          1,
		"source_availability":     "metadata_only",
		"payload_plaintext_bytes": 256,
		"created_at":              "2026-08-06T09:00:00Z",
		"updated_at":              "2026-08-06T09:30:00Z",
	}
}

func requireBearer(t *testing.T, r *http.Request, credentialMaterial string) {
	t.Helper()
	if r.Header.Get("Authorization") != "Bearer "+credentialMaterial {
		t.Fatalf("unexpected authorization header for %s", r.URL.Path)
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, status int, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("write json: %v", err)
	}
}

func writeEvidence(t *testing.T, evidence map[string]any) {
	t.Helper()
	path := os.Getenv("PATCHXNOTE_E2E_ARTIFACT")
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create evidence dir: %v", err)
	}
	body, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		t.Fatalf("encode evidence: %v", err)
	}
	if err := os.WriteFile(path, append(body, '\n'), 0o600); err != nil {
		t.Fatalf("write evidence: %v", err)
	}
}

func Example_commandShape() {
	fmt.Println("patchxnote mcp serve")
	// Output: patchxnote mcp serve
}
