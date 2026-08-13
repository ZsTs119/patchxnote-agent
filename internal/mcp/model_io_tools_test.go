package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
)

func TestModelIOFieldToolsReturnOnlyRequestedFields(t *testing.T) {
	server := newWebhookTestServer(t, true)
	tests := []struct {
		name      string
		args      map[string]any
		want      string
		forbidden []string
	}{
		{
			name:      "patchxnote_get_model_io_source_text",
			args:      map[string]any{"memory_id": "mem_fixture_1", "platform": "mobile"},
			want:      "转写安全文本",
			forbidden: []string{"模型原始响应", "解析结果", "封装结构"},
		},
		{
			name:      "patchxnote_get_model_io_provider_response",
			args:      map[string]any{"memory_id": "mem_fixture_1", "platform": "mobile"},
			want:      "模型原始响应",
			forbidden: []string{"客户端请求", "模型请求", "解析结果", "封装结构", "供应商尝试"},
		},
		{
			name:      "patchxnote_get_model_io_parsed_result",
			args:      map[string]any{"memory_id": "mem_fixture_1", "platform": "mobile"},
			want:      "解析结果",
			forbidden: []string{"客户端请求", "模型请求", "模型原始响应", "封装结构", "供应商尝试"},
		},
		{
			name:      "patchxnote_get_model_io_packaged_result",
			args:      map[string]any{"request_id": "mrun_fixture_1", "platform": "mobile"},
			want:      "封装结构",
			forbidden: []string{"客户端请求", "模型请求", "模型原始响应", "解析结果", "供应商尝试"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := callWebhookTool(t, server, tt.name, tt.args)
			if response["error"] != nil {
				t.Fatalf("tool error: %+v", response["error"])
			}
			text := toolTextResult(t, response)
			if !strings.Contains(text, tt.want) {
				t.Fatalf("expected %q in response:\n%s", tt.want, text)
			}
			for _, forbidden := range tt.forbidden {
				if strings.Contains(text, forbidden) {
					t.Fatalf("response leaked unrelated field %q:\n%s", forbidden, text)
				}
			}
			if strings.Contains(tt.name, "provider_response") && !strings.Contains(text, `"json": {`) {
				t.Fatalf("expected JSON field as object, got:\n%s", text)
			}
		})
	}
}

func TestListModelIOTracesTool(t *testing.T) {
	fakeAPI := &fakeToolAPI{}
	server := NewServer(Options{
		Authenticator: authenticatedManager(t),
		API:           fakeAPI,
	})
	response := callWebhookTool(t, server, "patchxnote_list_model_io_traces", map[string]any{
		"platform":    "mobile",
		"task_type":   "daily_digest",
		"state":       "completed",
		"business_id": "business-fixture",
		"date_from":   "2026-08-13T00:00:00Z",
		"limit":       10,
		"cursor":      "cursor_trace_page_1",
	})
	if response["error"] != nil {
		t.Fatalf("tool error: %+v", response["error"])
	}
	text := toolTextResult(t, response)
	if !strings.Contains(text, "mrun_fixture_trace_1") || !strings.Contains(text, "cursor_trace_page_2") {
		t.Fatalf("unexpected trace list response:\n%s", text)
	}
	for _, forbidden := range []string{"模型原始响应", "转写安全文本", "客户端请求"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("trace list leaked forbidden marker %q:\n%s", forbidden, text)
		}
	}
	if fakeAPI.traceListParams.Platform != "mobile" || fakeAPI.traceListParams.TaskType != "daily_digest" ||
		fakeAPI.traceListParams.BusinessID != "business-fixture" || fakeAPI.traceListParams.Limit != 10 ||
		fakeAPI.traceListParams.Cursor != "cursor_trace_page_1" {
		t.Fatalf("unexpected trace list params: %+v", fakeAPI.traceListParams)
	}
}

func TestListModelIOTracesToolValidation(t *testing.T) {
	server := newWebhookTestServer(t, true)
	missingPlatform := callWebhookTool(t, server, "patchxnote_list_model_io_traces", map[string]any{"limit": 10})
	assertRPCError(t, missingPlatform, codeInvalidParams, "invalid_params")

	badTask := callWebhookTool(t, server, "patchxnote_list_model_io_traces", map[string]any{"platform": "mobile", "task_type": "unknown"})
	assertRPCError(t, badTask, codeInvalidParams, "invalid_params")

	badDate := callWebhookTool(t, server, "patchxnote_list_model_io_traces", map[string]any{
		"platform":  "mobile",
		"date_from": "2026-08-14T00:00:00Z",
		"date_to":   "2026-08-13T00:00:00Z",
	})
	assertRPCError(t, badDate, codeInvalidParams, "invalid_params")
}

func TestModelIOToolWritesOutWithoutInlinePayload(t *testing.T) {
	server := newWebhookTestServer(t, true)
	out := filepath.Join(t.TempDir(), "解析 结果.json")
	response := callWebhookTool(t, server, "patchxnote_get_model_io_parsed_result", map[string]any{
		"memory_id": "mem_fixture_1",
		"platform":  "mobile",
		"out":       out,
	})
	if response["error"] != nil {
		t.Fatalf("tool error: %+v", response["error"])
	}
	text := toolTextResult(t, response)
	if strings.Contains(text, "解析结果") {
		t.Fatalf("file-write summary should not inline field content:\n%s", text)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !json.Valid(body) || !strings.Contains(string(body), "解析结果") {
		t.Fatalf("unexpected output file:\n%s", body)
	}
}

func TestModelIOToolAuthValidationAndUnavailable(t *testing.T) {
	unauthenticated := newWebhookTestServer(t, false)
	response := callWebhookTool(t, unauthenticated, "patchxnote_get_model_io_source_text", map[string]any{"memory_id": "mem_fixture_1"})
	assertRPCError(t, response, codeAuthRequired, "auth_required")

	authenticated := newWebhookTestServer(t, true)
	invalid := callWebhookTool(t, authenticated, "patchxnote_get_model_io_source_text", map[string]any{
		"memory_id":  "mem_fixture_1",
		"request_id": "mrun_fixture_1",
	})
	assertRPCError(t, invalid, codeInvalidParams, "invalid_params")

	unavailable := newWebhookTestServerWithAPI(t, true, &unavailableModelIOAPI{})
	got := callWebhookTool(t, unavailable, "patchxnote_get_model_io_provider_response", map[string]any{"memory_id": "mem_fixture_1"})
	if got["error"] != nil {
		t.Fatalf("unavailable field should be a structured result, got error: %+v", got["error"])
	}
	text := toolTextResult(t, got)
	if !strings.Contains(text, `"available": false`) || !strings.Contains(text, `"status": "unavailable"`) || strings.Contains(text, "模型原始响应") {
		t.Fatalf("unexpected unavailable response:\n%s", text)
	}
}

func TestModelIOToolLargeInlineRequiresOut(t *testing.T) {
	server := newWebhookTestServerWithAPI(t, true, &largeModelIOAPI{})
	response := callWebhookTool(t, server, "patchxnote_get_model_io_provider_response", map[string]any{"memory_id": "mem_fixture_1"})
	assertRPCError(t, response, codeToolError, "output_too_large")

	out := filepath.Join(t.TempDir(), "large.json")
	withOut := callWebhookTool(t, server, "patchxnote_get_model_io_provider_response", map[string]any{
		"memory_id": "mem_fixture_1",
		"out":       out,
	})
	if withOut["error"] != nil {
		t.Fatalf("large field with out failed: %+v", withOut["error"])
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("expected large output file: %v", err)
	}
}

type unavailableModelIOAPI struct {
	fakeToolAPI
}

func (f *unavailableModelIOAPI) GetMemoryModelIO(ctx context.Context, accessToken string, platform string, memoryID string) (api.AgentModelIOExport, error) {
	export := toolFixtureModelIOExport()
	export.FieldStatus.ProviderResponseJSON = "unavailable"
	return export, nil
}

type largeModelIOAPI struct {
	fakeToolAPI
}

func (f *largeModelIOAPI) GetMemoryModelIO(ctx context.Context, accessToken string, platform string, memoryID string) (api.AgentModelIOExport, error) {
	export := toolFixtureModelIOExport()
	export.ProviderResponseJSON = json.RawMessage(`{"content":"` + strings.Repeat("x", maxToolOutputBytes+2048) + `"}`)
	return export, nil
}
