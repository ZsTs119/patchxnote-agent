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
		case "/v1/agent/memories/mem_fixture_1/delivery-document":
			requireBearer(t, r, credentialMaterial)
			if got := r.URL.Query().Get("platform"); got != "mobile" {
				t.Fatalf("unexpected delivery document platform: %s", r.URL.RawQuery)
			}
			writeJSON(t, w, http.StatusOK, deliveryFixture())
		case "/v1/agent/memories/mem_fixture_1/model-io":
			requireBearer(t, r, credentialMaterial)
			if got := r.URL.Query().Get("platform"); got != "mobile" {
				t.Fatalf("unexpected model IO platform: %s", r.URL.RawQuery)
			}
			writeJSON(t, w, http.StatusOK, modelIOFixture())
		case "/v1/agent/model-io-traces":
			requireBearer(t, r, credentialMaterial)
			if got := r.URL.Query().Get("platform"); got != "mobile" {
				t.Fatalf("unexpected model IO trace list platform: %s", r.URL.RawQuery)
			}
			writeJSON(t, w, http.StatusOK, modelIOTraceListFixture())
		case "/v1/agent/model-runs/req_fixture_1/io-trace":
			requireBearer(t, r, credentialMaterial)
			if got := r.URL.Query().Get("platform"); got != "mobile" {
				t.Fatalf("unexpected model IO trace platform: %s", r.URL.RawQuery)
			}
			writeJSON(t, w, http.StatusOK, modelIOFixture())
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

	mcpDraftDir := filepath.Join(home, "mcp-draft")
	mcpModelIOFile := filepath.Join(home, "mcp-model-io.json")
	mcpParsedResultFile := filepath.Join(home, "mcp-parsed-result.json")
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
		mcpToolCallLine(t, 10, "patchxnote_configure_webhook_target", map[string]any{"alias": "MCP 网关", "type": "generic", "webhook_url": server.URL + "/webhook/generic", "template": "daily-review"}),
		mcpToolCallLine(t, 11, "patchxnote_list_webhook_targets", map[string]any{"include_disabled": false}),
		mcpToolCallLine(t, 12, "patchxnote_list_webhook_templates", map[string]any{}),
		mcpToolCallLine(t, 13, "patchxnote_render_webhook_message", map[string]any{"memory_id": "mem_fixture_1", "platform": "mobile", "template": "daily-review", "save_draft": true, "out": mcpDraftDir, "include_model_io": true, "force": true}),
		mcpToolCallLine(t, 14, "patchxnote_export_model_io", map[string]any{"memory_id": "mem_fixture_1", "platform": "mobile", "out": mcpModelIOFile, "force": true}),
		mcpToolCallLine(t, 15, "patchxnote_send_webhook", map[string]any{"target_aliases": []string{"MCP 网关"}, "title": "MCP Webhook Smoke", "markdown": "## MCP Webhook Smoke\n\n发送内容\n"}),
		mcpToolCallLine(t, 16, "patchxnote_remove_webhook_target", map[string]any{"alias": "MCP 网关"}),
		mcpToolCallLine(t, 17, "patchxnote_list_webhook_targets", map[string]any{}),
		mcpToolCallLine(t, 18, "patchxnote_list_model_io_traces", map[string]any{"platform": "mobile", "limit": 20}),
		mcpToolCallLine(t, 19, "patchxnote_get_model_io_source_text", map[string]any{"memory_id": "mem_fixture_1", "platform": "mobile"}),
		mcpToolCallLine(t, 20, "patchxnote_get_model_io_provider_response", map[string]any{"memory_id": "mem_fixture_1", "platform": "mobile"}),
		mcpToolCallLine(t, 21, "patchxnote_get_model_io_parsed_result", map[string]any{"memory_id": "mem_fixture_1", "platform": "mobile", "out": mcpParsedResultFile}),
		mcpToolCallLine(t, 22, "patchxnote_get_model_io_packaged_result", map[string]any{"request_id": "req_fixture_1", "platform": "mobile"}),
	}, "\n") + "\n"
	mcpResult := runCLIWithInput(t, binary, env, mcpInput, "mcp", "serve")
	mcpResponses := assertMCPResponses(t, mcpResult.stdout, 22)
	assertMCPToolCount(t, mcpResponses, 19)
	if strings.Contains(mcpResult.stdout, server.URL+"/webhook/generic") {
		t.Fatalf("MCP output leaked raw webhook URL:\n%s", mcpResult.stdout)
	}
	traceListText := mcpResponseText(t, mcpResponses, 18)
	sourceFieldText := mcpResponseText(t, mcpResponses, 19)
	providerFieldText := mcpResponseText(t, mcpResponses, 20)
	packagedFieldText := mcpResponseText(t, mcpResponses, 22)
	if !strings.Contains(traceListText, "req_fixture_1") || !strings.Contains(traceListText, `"source_text_availability": "available"`) {
		t.Fatalf("MCP model IO trace list did not return expected metadata:\n%s", traceListText)
	}
	if !strings.Contains(sourceFieldText, "安全投影文本") || !strings.Contains(providerFieldText, `"content": "ok"`) || !strings.Contains(packagedFieldText, "合成记录摘要") {
		t.Fatalf("MCP model IO field tools did not return expected inline fields:\nsource=%s\nprovider=%s\npackaged=%s", sourceFieldText, providerFieldText, packagedFieldText)
	}
	if strings.Contains(providerFieldText, `"messages"`) {
		t.Fatalf("MCP provider-response leaked provider request content:\n%s", providerFieldText)
	}
	for _, name := range []string{"source.json", "message.md", "metadata.json", "model-io.json"} {
		if _, err := os.Stat(filepath.Join(mcpDraftDir, name)); err != nil {
			t.Fatalf("expected MCP draft file %s: %v", name, err)
		}
	}
	if _, err := os.Stat(mcpModelIOFile); err != nil {
		t.Fatalf("expected MCP model IO export file: %v", err)
	}
	if body, err := os.ReadFile(mcpParsedResultFile); err != nil || !json.Valid(body) || !strings.Contains(string(body), `"summary": "ok"`) {
		t.Fatalf("expected MCP parsed result file, err=%v body=%s", err, body)
	}

	sourceText := runCLI(t, binary, env, "model-io", "source-text", "--memory-id", "mem_fixture_1", "--platform", "mobile")
	if sourceText.stdout != "安全投影文本\n" {
		t.Fatalf("unexpected model-io source-text output:\n%s", sourceText.stdout)
	}
	providerOut := filepath.Join(home, "provider-response.json")
	runCLI(t, binary, env, "model-io", "provider-response", "--memory-id", "mem_fixture_1", "--platform", "mobile", "--out", providerOut)
	if body, err := os.ReadFile(providerOut); err != nil || !json.Valid(body) || !strings.Contains(string(body), `"content": "ok"`) {
		t.Fatalf("expected CLI provider response file, err=%v body=%s", err, body)
	}
	fullModelIOOut := filepath.Join(home, "cli-model-io.json")
	runCLI(t, binary, env, "model-io", "export", "--memory-id", "mem_fixture_1", "--platform", "mobile", "--out", fullModelIOOut)
	if _, err := os.Stat(fullModelIOOut); err != nil {
		t.Fatalf("expected CLI model IO export file: %v", err)
	}
	traceList := runCLI(t, binary, env, "model-io", "list", "--platform", "mobile")
	if !strings.Contains(traceList.stdout, "req_fixture_1") || !strings.Contains(traceList.stdout, "available") {
		t.Fatalf("unexpected model-io list output:\n%s", traceList.stdout)
	}

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
	if webhookCalls != 4 {
		t.Fatalf("expected webhook MCP/test/file/draft calls, got %d", webhookCalls)
	}

	runCLI(t, binary, env, "logout")
	status = runCLI(t, binary, env, "auth", "status")
	if !strings.Contains(status.stdout, "unauthenticated") {
		t.Fatalf("expected unauthenticated after logout, got:\n%s", status.stdout)
	}

	writeEvidence(t, map[string]any{
		"status":      "pass",
		"commands":    []string{"login", "auth status", "mcp serve read tools", "mcp serve webhook tools", "mcp serve model-io tools", "model-io list/source-text/provider-response/export", "webhook set/list/test/send", "logout", "auth status"},
		"tool_count":  19,
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

func mcpToolCallLine(t *testing.T, id int, name string, arguments map[string]any) string {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      name,
			"arguments": arguments,
		},
	})
	if err != nil {
		t.Fatalf("encode MCP call %s: %v", name, err)
	}
	return string(body)
}

func assertMCPResponses(t *testing.T, stdout string, want int) []map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != want {
		t.Fatalf("expected %d MCP responses, got %d\n%s", want, len(lines), stdout)
	}
	responses := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		var response map[string]any
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			t.Fatalf("MCP stdout line is not JSON: %v\n%s", err, line)
		}
		if response["error"] != nil {
			t.Fatalf("unexpected MCP error response: %+v", response["error"])
		}
		responses = append(responses, response)
	}
	return responses
}

func assertMCPToolCount(t *testing.T, responses []map[string]any, want int) {
	t.Helper()
	for _, response := range responses {
		if response["id"] != float64(2) {
			continue
		}
		result, ok := response["result"].(map[string]any)
		if !ok {
			t.Fatalf("tools/list result has unexpected shape: %+v", response["result"])
		}
		tools, ok := result["tools"].([]any)
		if !ok {
			t.Fatalf("tools/list tools has unexpected shape: %+v", result["tools"])
		}
		if len(tools) != want {
			t.Fatalf("expected %d MCP tools, got %d", want, len(tools))
		}
		return
	}
	t.Fatalf("tools/list response was not found")
}

func mcpResponseText(t *testing.T, responses []map[string]any, id int) string {
	t.Helper()
	for _, response := range responses {
		if response["id"] != float64(id) {
			continue
		}
		return toolTextFromMCPResult(t, response)
	}
	t.Fatalf("MCP response id %d was not found", id)
	return ""
}

func toolTextFromMCPResult(t *testing.T, response map[string]any) string {
	t.Helper()
	result, ok := response["result"].(map[string]any)
	if !ok {
		t.Fatalf("MCP response result has unexpected shape: %+v", response["result"])
	}
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("MCP result content has unexpected shape: %+v", result["content"])
	}
	item, ok := content[0].(map[string]any)
	if !ok {
		t.Fatalf("MCP content item has unexpected shape: %+v", content[0])
	}
	text, ok := item["text"].(string)
	if !ok {
		t.Fatalf("MCP content text has unexpected shape: %+v", item["text"])
	}
	return text
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

func deliveryFixture() map[string]any {
	return map[string]any{
		"source":   "patchxnote-agent",
		"version":  "v1",
		"title":    "合成记录摘要",
		"summary":  "这是一条用于 Agent webhook MCP 测试的安全摘要。",
		"markdown": "## 记录摘要\n\n这是一条用于 Agent webhook MCP 测试的安全摘要。",
		"sections": []map[string]any{
			{"title": "背景", "markdown": "用户完成了一条记录同步。"},
			{"title": "结论", "markdown": "Agent 可以渲染并发送 webhook。"},
		},
		"key_items": []map[string]any{
			{"title": "继续验收", "status": "open", "owner": "Agent", "due_at": "2026-08-14", "markdown": "完成本地安装链路验证。"},
		},
		"memory": map[string]any{
			"id":                  "mem_fixture_1",
			"platform":            "mobile",
			"object_type":         "event",
			"client_object_id":    "local_event_fixture",
			"revision_id":         "rev_fixture_1",
			"revision":            3,
			"schema_id":           "patchnote.event.v1",
			"schema_version":      1,
			"source_availability": "safe_plaintext_projection",
		},
		"trace": map[string]any{
			"trace_id":   "trace_fixture_1",
			"request_id": "req_fixture_1",
			"platform":   "mobile",
			"task_type":  "event_summary",
			"state":      "succeeded",
		},
		"generated_at": "2026-08-13T12:00:00Z",
	}
}

func modelIOFixture() map[string]any {
	return map[string]any{
		"source":  "patchxnote-agent",
		"version": "v1",
		"memory": map[string]any{
			"id":                  "mem_fixture_1",
			"platform":            "mobile",
			"object_type":         "event",
			"client_object_id":    "local_event_fixture",
			"revision_id":         "rev_fixture_1",
			"revision":            3,
			"schema_id":           "patchnote.event.v1",
			"schema_version":      1,
			"source_availability": "safe_plaintext_projection",
		},
		"trace": map[string]any{
			"trace_id":   "trace_fixture_1",
			"request_id": "req_fixture_1",
			"platform":   "mobile",
			"task_type":  "event_summary",
			"state":      "succeeded",
		},
		"source_text": map[string]any{"availability": "available", "text": "安全投影文本"},
		"field_status": map[string]any{
			"client_request_json":    "available",
			"provider_request_json":  "available",
			"provider_response_json": "available",
			"parsed_result_json":     "available",
			"packaged_result_json":   "available",
			"provider_attempts_json": "available",
		},
		"client_request_json":    map[string]any{"memory_id": "mem_fixture_1"},
		"provider_request_json":  map[string]any{"messages": []string{"safe"}},
		"provider_response_json": map[string]any{"content": "ok"},
		"parsed_result_json":     map[string]any{"summary": "ok"},
		"packaged_result_json":   map[string]any{"title": "合成记录摘要"},
		"provider_attempts_json": []map[string]any{{"status": "succeeded"}},
	}
}

func modelIOTraceListFixture() map[string]any {
	return map[string]any{
		"items": []map[string]any{{
			"request_id":               "req_fixture_1",
			"platform":                 "mobile",
			"api_contract_version":     "v1",
			"task_type":                "event_summary",
			"state":                    "succeeded",
			"safe_error_code":          nil,
			"recording_id":             "rec_fixture_1",
			"event_id":                 "event_fixture_1",
			"business_id":              "biz_fixture_1",
			"created_at":               "2026-08-13T12:00:00Z",
			"updated_at":               "2026-08-13T12:00:05Z",
			"completed_at":             "2026-08-13T12:00:05Z",
			"source_text_availability": "available",
			"field_status": map[string]any{
				"client_request_json":    "available",
				"provider_request_json":  "available",
				"provider_response_json": "available",
				"parsed_result_json":     "available",
				"packaged_result_json":   "available",
				"provider_attempts_json": "available",
			},
			"field_bytes": map[string]any{
				"client_request_json":    29,
				"provider_request_json":  21,
				"provider_response_json": 17,
				"parsed_result_json":     17,
				"packaged_result_json":   28,
				"provider_attempts_json": 26,
			},
			"memory": map[string]any{
				"id":                  "mem_fixture_1",
				"platform":            "mobile",
				"object_type":         "event",
				"client_object_id":    "local_event_fixture",
				"revision_id":         "rev_fixture_1",
				"revision":            3,
				"schema_id":           "patchnote.event.v1",
				"schema_version":      1,
				"source_availability": "safe_plaintext_projection",
			},
		}},
		"next_cursor": "",
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
