package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ZsTs119/patchxnote-agent/internal/auth"
)

type staticAuthenticator struct {
	status auth.Status
	err    error
}

func (a staticAuthenticator) Status(ctx context.Context) (auth.Status, error) {
	return a.status, a.err
}

func TestInitializeAndToolsList(t *testing.T) {
	responses := serveForTest(t, strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
	}, "\n"))

	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}
	initialize := responses[0]
	if initialize["error"] != nil {
		t.Fatalf("unexpected initialize error: %+v", initialize["error"])
	}
	result := initialize["result"].(map[string]any)
	if result["protocolVersion"] != protocolVersion {
		t.Fatalf("unexpected protocol version: %+v", result["protocolVersion"])
	}

	toolsResult := responses[1]["result"].(map[string]any)
	tools := toolsResult["tools"].([]any)
	if len(tools) != 7 {
		t.Fatalf("expected seven tools, got %d", len(tools))
	}
	names := make(map[string]bool)
	for _, item := range tools {
		tool := item.(map[string]any)
		names[tool["name"].(string)] = true
		if _, ok := tool["inputSchema"].(map[string]any); !ok {
			t.Fatalf("tool missing input schema: %+v", tool)
		}
	}
	for _, want := range []string{
		"patchxnote_get_current_user",
		"patchxnote_list_recorder_cards",
		"patchxnote_get_quota_summary",
		"patchxnote_get_model_usage_summary",
		"patchxnote_list_memories",
		"patchxnote_search_memories",
		"patchxnote_get_memory",
	} {
		if !names[want] {
			t.Fatalf("tool %s missing from tools/list", want)
		}
	}
}

func TestUnauthenticatedToolCallReturnsAuthRequired(t *testing.T) {
	responses := serveForTest(t, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"patchxnote_get_current_user","arguments":{}}}`)
	if len(responses) != 1 {
		t.Fatalf("expected one response, got %d", len(responses))
	}

	assertRPCError(t, responses[0], codeAuthRequired, "auth_required")
}

func TestUnknownMethodAndToolAreStableErrors(t *testing.T) {
	responses := serveForTest(t, strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"unknown/method","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"patchxnote_missing","arguments":{}}}`,
	}, "\n"))
	if len(responses) != 2 {
		t.Fatalf("expected two responses, got %d", len(responses))
	}

	assertRPCError(t, responses[0], codeMethodNotFound, "method_not_found")
	assertRPCError(t, responses[1], codeInvalidParams, "tool_not_found")
}

func TestToolArgumentValidation(t *testing.T) {
	responses := serveForTest(t, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"patchxnote_list_memories","arguments":{"platform":"mobile","limit":100}}}`)
	if len(responses) != 1 {
		t.Fatalf("expected one response, got %d", len(responses))
	}

	assertRPCError(t, responses[0], codeInvalidParams, "invalid_params")
}

func TestNotificationDoesNotRespond(t *testing.T) {
	responses := serveForTest(t, `{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}`)
	if len(responses) != 0 {
		t.Fatalf("expected no response for notification, got %+v", responses)
	}
}

func serveForTest(t *testing.T, input string) []map[string]any {
	t.Helper()
	server := NewServer(Options{
		Authenticator: staticAuthenticator{status: auth.Status{Authenticated: false, Profile: "default"}},
	})
	var stdout strings.Builder
	if err := server.Serve(context.Background(), strings.NewReader(input+"\n"), &stdout); err != nil {
		t.Fatalf("serve: %v", err)
	}
	if strings.TrimSpace(stdout.String()) == "" {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	responses := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		var response map[string]any
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			t.Fatalf("stdout line is not JSON: %v\n%s", err, line)
		}
		if response["jsonrpc"] != "2.0" {
			t.Fatalf("unexpected jsonrpc marker: %+v", response)
		}
		responses = append(responses, response)
	}
	return responses
}

func assertRPCError(t *testing.T, response map[string]any, rpcCode int, stableCode string) {
	t.Helper()
	errorBody, ok := response["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected rpc error, got %+v", response)
	}
	if int(errorBody["code"].(float64)) != rpcCode {
		t.Fatalf("expected rpc code %d, got %+v", rpcCode, errorBody["code"])
	}
	data := errorBody["data"].(map[string]any)
	if data["code"] != stableCode {
		t.Fatalf("expected stable code %s, got %+v", stableCode, data)
	}
}
