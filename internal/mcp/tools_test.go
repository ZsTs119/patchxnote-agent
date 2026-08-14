package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/auth"
	"github.com/ZsTs119/patchxnote-agent/internal/cache"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
)

type fakeToolAPI struct {
	err             error
	listParams      api.ListMemoriesParams
	traceListParams api.ListModelIOTracesParams
	largeCards      bool
}

func (f *fakeToolAPI) CurrentUser(ctx context.Context, accessToken string) (api.CurrentAccount, error) {
	if f.err != nil {
		return api.CurrentAccount{}, f.err
	}
	return toolFixtureAccount(), nil
}

func (f *fakeToolAPI) ListRecorderCards(ctx context.Context, accessToken string) (api.AgentRecorderCardPage, error) {
	if f.err != nil {
		return api.AgentRecorderCardPage{}, f.err
	}
	if f.largeCards {
		items := make([]api.AgentRecorderCard, 300)
		for index := range items {
			items[index] = api.AgentRecorderCard{
				ID:             "hw_fixture_" + strings.Repeat("x", 80),
				BindingEpochID: "binding_epoch_fixture",
				IdentityMasked: "MR20-****-0000",
				BindingStatus:  "active",
				CreatedAt:      fixedToolTime(),
			}
		}
		return api.AgentRecorderCardPage{Items: items}, nil
	}
	return api.AgentRecorderCardPage{Items: []api.AgentRecorderCard{
		{
			ID:                "hw_fixture",
			BindingEpochID:    "binding_epoch_fixture",
			IdentityMasked:    "MR20-****-0000",
			BindingStatus:     "active",
			CredentialVersion: 1,
			CreatedAt:         fixedToolTime(),
		},
	}}, nil
}

func (f *fakeToolAPI) GetQuotaSummary(ctx context.Context, accessToken string) (api.AgentQuotaSummary, error) {
	if f.err != nil {
		return api.AgentQuotaSummary{}, f.err
	}
	return api.AgentQuotaSummary{
		AvailableTokens:           120000,
		GiftAvailableTokens:       20000,
		PaidAvailableTokens:       90000,
		AdjustmentAvailableTokens: 10000,
		CalculatedAt:              fixedToolTime(),
	}, nil
}

func (f *fakeToolAPI) GetModelUsageSummary(ctx context.Context, accessToken string) (api.AgentModelUsageSummary, error) {
	if f.err != nil {
		return api.AgentModelUsageSummary{}, f.err
	}
	return api.AgentModelUsageSummary{
		Period:              "current_month",
		PeriodStartsAt:      fixedToolTime(),
		PeriodEndsAt:        fixedToolTime().Add(24 * time.Hour),
		ProviderTotalTokens: 30000,
		ChargedQuotaTokens:  15000,
		RunCount:            12,
		AttemptCount:        14,
		GiftChargedTokens:   5000,
		PaidChargedTokens:   10000,
		ValueBasis:          "versioned_reference_price",
		CalculatedAt:        fixedToolTime(),
	}, nil
}

func (f *fakeToolAPI) ListMemories(ctx context.Context, accessToken string, params api.ListMemoriesParams) (api.AgentMemoryPage, error) {
	if f.err != nil {
		return api.AgentMemoryPage{}, f.err
	}
	f.listParams = params
	return api.AgentMemoryPage{
		Items:      []api.AgentMemory{toolFixtureMemory(), toolFixtureSyntheticMemory()},
		NextCursor: "cursor_page_2",
	}, nil
}

func (f *fakeToolAPI) GetMemory(ctx context.Context, accessToken string, platform string, memoryID string) (api.AgentMemory, error) {
	if f.err != nil {
		return api.AgentMemory{}, f.err
	}
	return toolFixtureMemory(), nil
}

func (f *fakeToolAPI) GetMemoryDeliveryDocument(ctx context.Context, accessToken string, platform string, memoryID string) (api.AgentDeliveryDocument, error) {
	if f.err != nil {
		return api.AgentDeliveryDocument{}, f.err
	}
	return toolFixtureDeliveryDocument(), nil
}

func (f *fakeToolAPI) GetMemoryModelIO(ctx context.Context, accessToken string, platform string, memoryID string) (api.AgentModelIOExport, error) {
	if f.err != nil {
		return api.AgentModelIOExport{}, f.err
	}
	return toolFixtureModelIOExport(), nil
}

func (f *fakeToolAPI) GetModelRunIOTrace(ctx context.Context, accessToken string, platform string, requestID string) (api.AgentModelIOExport, error) {
	if f.err != nil {
		return api.AgentModelIOExport{}, f.err
	}
	return toolFixtureModelIOExport(), nil
}

func (f *fakeToolAPI) ListModelIOTraces(ctx context.Context, accessToken string, params api.ListModelIOTracesParams) (api.AgentModelIOTracePage, error) {
	if f.err != nil {
		return api.AgentModelIOTracePage{}, f.err
	}
	f.traceListParams = params
	return api.AgentModelIOTracePage{
		Items: []api.AgentModelIOTraceSummary{
			{
				RequestID:              "mrun_fixture_trace_1",
				Platform:               params.Platform,
				APIContractVersion:     "1.0.0",
				TaskType:               "daily_digest",
				State:                  "completed",
				CreatedAt:              fixedToolTime(),
				UpdatedAt:              fixedToolTime(),
				SourceTextAvailability: "available",
				FieldStatus: api.AgentModelIOFieldStatus{
					ProviderResponseJSON: "available",
					ParsedResultJSON:     "available",
					PackagedResultJSON:   "missing",
				},
				FieldBytes: api.AgentModelIOFieldBytes{ProviderResponseJSON: 256, ParsedResultJSON: 96},
			},
		},
		NextCursor: "cursor_trace_page_2",
	}, nil
}

func TestV1ToolSuccessCallsAPIAndSearchesCache(t *testing.T) {
	fakeAPI := &fakeToolAPI{}
	server := NewServer(Options{
		Authenticator: authenticatedManager(t),
		API:           fakeAPI,
		MemoryCache:   cache.NewMemoryIndex(),
	})

	calls := []string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"patchxnote_get_current_user","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"patchxnote_list_recorder_cards","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"patchxnote_get_quota_summary","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"patchxnote_get_model_usage_summary","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"patchxnote_list_memories","arguments":{"platform":"mobile","limit":20}}}`,
		`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"patchxnote_search_memories","arguments":{"platform":"mobile","query":"模型整理","limit":10}}}`,
		`{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"patchxnote_get_memory","arguments":{"platform":"mobile","memory_id":"mem_fixture_1"}}}`,
	}
	responses := serveCustomForTest(t, server, strings.Join(calls, "\n"))
	if len(responses) != len(calls) {
		t.Fatalf("expected %d responses, got %d", len(calls), len(responses))
	}
	for _, response := range responses {
		if response["error"] != nil {
			t.Fatalf("unexpected tool error: %+v", response["error"])
		}
		text := toolTextResult(t, response)
		if strings.Contains(text, strings.Repeat("z", 32)) {
			t.Fatal("tool result leaked credential material")
		}
	}
	if fakeAPI.listParams.Platform != "mobile" || fakeAPI.listParams.Limit != 20 {
		t.Fatalf("unexpected list params: %+v", fakeAPI.listParams)
	}

	searchText := toolTextResult(t, responses[5])
	if !strings.Contains(searchText, `"id": "mrun_fixture_synthetic_1"`) ||
		!strings.Contains(searchText, `"source": "model_io_trace"`) ||
		!strings.Contains(searchText, `"title": "合成记录总结"`) {
		t.Fatalf("expected cached memory search result, got:\n%s", searchText)
	}
}

func TestEveryV1ToolRequiresAuth(t *testing.T) {
	for _, call := range validToolCalls() {
		input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":` + call + `}`
		responses := serveForTest(t, input)
		if len(responses) != 1 {
			t.Fatalf("expected one response, got %d", len(responses))
		}
		assertRPCError(t, responses[0], codeAuthRequired, "auth_required")
	}
}

func TestEveryV1ToolMapsForbiddenAndServerError(t *testing.T) {
	for _, tt := range []struct {
		name     string
		apiErr   *api.Error
		wantCode string
		wantRPC  int
	}{
		{
			name:     "forbidden",
			apiErr:   &api.Error{StatusCode: 403, Code: "permission_denied", RequestID: "req_fixture_forbidden"},
			wantCode: "permission_denied",
			wantRPC:  codeToolError,
		},
		{
			name:     "unavailable",
			apiErr:   &api.Error{StatusCode: 503, Code: "dependency_unavailable", RequestID: "req_fixture_unavailable", Retryable: true},
			wantCode: "dependency_unavailable",
			wantRPC:  codeToolError,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for _, call := range apiBackedToolCalls() {
				server := NewServer(Options{
					Authenticator: authenticatedManager(t),
					API:           &fakeToolAPI{err: tt.apiErr},
					MemoryCache:   cache.NewMemoryIndex(),
				})
				input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":` + call + `}`
				responses := serveCustomForTest(t, server, input)
				assertRPCError(t, responses[0], tt.wantRPC, tt.wantCode)
			}
		})
	}
}

func TestEveryV1ToolChecksRequiredScope(t *testing.T) {
	for _, call := range validToolCalls() {
		store := keychain.NewMemoryStore()
		if err := store.Put(context.Background(), "default", keychain.Credential{
			AccountID:    "acct_fixture",
			AccessToken:  strings.Repeat("z", 32),
			RefreshToken: strings.Repeat("y", 43),
			Scopes:       []string{"agent:account.read"},
		}); err != nil {
			t.Fatalf("seed credential: %v", err)
		}
		server := NewServer(Options{
			Authenticator: auth.NewManager(store, "default"),
			API:           &fakeToolAPI{},
			MemoryCache:   cache.NewMemoryIndex(),
		})
		input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":` + call + `}`
		responses := serveCustomForTest(t, server, input)
		if strings.Contains(call, "patchxnote_get_current_user") {
			if responses[0]["error"] != nil {
				t.Fatalf("account scope should allow current user, got %+v", responses[0]["error"])
			}
			continue
		}
		assertRPCError(t, responses[0], codeToolError, "permission_denied")
	}
}

func TestEveryV1ToolHasValidationCoverage(t *testing.T) {
	tests := []string{
		`{"name":"patchxnote_get_current_user","arguments":{"unexpected":true}}`,
		`{"name":"patchxnote_list_recorder_cards","arguments":{"unexpected":true}}`,
		`{"name":"patchxnote_get_quota_summary","arguments":{"unexpected":true}}`,
		`{"name":"patchxnote_get_model_usage_summary","arguments":{"unexpected":true}}`,
		`{"name":"patchxnote_list_memories","arguments":{"platform":"agent"}}`,
		`{"name":"patchxnote_search_memories","arguments":{"platform":"mobile"}}`,
		`{"name":"patchxnote_get_memory","arguments":{"platform":"mobile"}}`,
	}
	for _, call := range tests {
		server := NewServer(Options{
			Authenticator: authenticatedManager(t),
			API:           &fakeToolAPI{},
			MemoryCache:   cache.NewMemoryIndex(),
		})
		input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":` + call + `}`
		responses := serveCustomForTest(t, server, input)
		assertRPCError(t, responses[0], codeInvalidParams, "invalid_params")
	}
}

func TestToolAPIErrorMappingAndOutputLimit(t *testing.T) {
	forbiddenServer := NewServer(Options{
		Authenticator: authenticatedManager(t),
		API: &fakeToolAPI{err: &api.Error{
			StatusCode: 403,
			Code:       "permission_denied",
			RequestID:  "req_fixture_forbidden",
		}},
		MemoryCache: cache.NewMemoryIndex(),
	})
	responses := serveCustomForTest(t, forbiddenServer, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"patchxnote_get_quota_summary","arguments":{}}}`)
	assertRPCError(t, responses[0], codeToolError, "permission_denied")

	largeServer := NewServer(Options{
		Authenticator: authenticatedManager(t),
		API:           &fakeToolAPI{largeCards: true},
		MemoryCache:   cache.NewMemoryIndex(),
	})
	responses = serveCustomForTest(t, largeServer, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"patchxnote_list_recorder_cards","arguments":{}}}`)
	assertRPCError(t, responses[0], codeToolError, "output_too_large")
}

func validToolCalls() []string {
	return []string{
		`{"name":"patchxnote_get_current_user","arguments":{}}`,
		`{"name":"patchxnote_list_recorder_cards","arguments":{}}`,
		`{"name":"patchxnote_get_quota_summary","arguments":{}}`,
		`{"name":"patchxnote_get_model_usage_summary","arguments":{}}`,
		`{"name":"patchxnote_list_memories","arguments":{"platform":"mobile"}}`,
		`{"name":"patchxnote_search_memories","arguments":{"platform":"mobile","query":"event"}}`,
		`{"name":"patchxnote_get_memory","arguments":{"platform":"mobile","memory_id":"mem_fixture_1"}}`,
	}
}

func apiBackedToolCalls() []string {
	return []string{
		`{"name":"patchxnote_get_current_user","arguments":{}}`,
		`{"name":"patchxnote_list_recorder_cards","arguments":{}}`,
		`{"name":"patchxnote_get_quota_summary","arguments":{}}`,
		`{"name":"patchxnote_get_model_usage_summary","arguments":{}}`,
		`{"name":"patchxnote_list_memories","arguments":{"platform":"mobile"}}`,
		`{"name":"patchxnote_get_memory","arguments":{"platform":"mobile","memory_id":"mem_fixture_1"}}`,
	}
}

func authenticatedManager(t *testing.T) *auth.Manager {
	t.Helper()
	store := keychain.NewMemoryStore()
	if err := store.Put(context.Background(), "default", keychain.Credential{
		AccountID:    "acct_fixture",
		AccessToken:  strings.Repeat("z", 32),
		RefreshToken: strings.Repeat("y", 43),
		Scopes: []string{
			"agent:account.read",
			"agent:hardware.read",
			"agent:quota.read",
			"agent:model_usage.read",
			"agent:content.read:mobile",
			"agent:content.read:desktop",
		},
	}); err != nil {
		t.Fatalf("seed credential: %v", err)
	}
	return auth.NewManager(store, "default")
}

func serveCustomForTest(t *testing.T, server *Server, input string) []map[string]any {
	t.Helper()
	var stdout strings.Builder
	if err := server.Serve(context.Background(), strings.NewReader(input+"\n"), &stdout); err != nil {
		t.Fatalf("serve: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	responses := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		var response map[string]any
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			t.Fatalf("stdout line is not JSON: %v\n%s", err, line)
		}
		responses = append(responses, response)
	}
	return responses
}

func toolTextResult(t *testing.T, response map[string]any) string {
	t.Helper()
	result := response["result"].(map[string]any)
	content := result["content"].([]any)
	first := content[0].(map[string]any)
	return first["text"].(string)
}

func toolFixtureAccount() api.CurrentAccount {
	return api.CurrentAccount{
		ID:                   "acct_fixture",
		Status:               "active",
		RegistrationPlatform: "mobile",
		PhoneMasked:          "+86*******0000",
		StateVersion:         2,
	}
}

func toolFixtureMemory() api.AgentMemory {
	return api.AgentMemory{
		ID:                    "mem_fixture_1",
		Platform:              "mobile",
		Source:                "structured_result",
		ObjectType:            "event",
		ClientObjectID:        "local_event_fixture",
		RevisionID:            "rev_fixture_1",
		Revision:              3,
		SchemaID:              "patchnote.event.v1",
		SchemaVersion:         1,
		SourceAvailability:    "metadata_only",
		PayloadPlaintextBytes: 256,
		CreatedAt:             fixedToolTime(),
		UpdatedAt:             fixedToolTime(),
	}
}

func toolFixtureSyntheticMemory() api.AgentMemory {
	return api.AgentMemory{
		ID:                    "mrun_fixture_synthetic_1",
		Platform:              "mobile",
		Source:                "model_io_trace",
		RequestID:             "mrun_fixture_synthetic_1",
		TaskType:              "event_summary",
		Title:                 "合成记录总结",
		Summary:               "这是一条用于测试 Agent 列表的模型整理结果。",
		ObjectType:            "event",
		ClientObjectID:        "rec_fixture_synthetic_1",
		RevisionID:            "mrun_fixture_synthetic_1",
		Revision:              1,
		SchemaID:              "patchnote.agent.model-io-trace",
		SchemaVersion:         1,
		SourceAvailability:    "text_only",
		PayloadPlaintextBytes: 512,
		CreatedAt:             fixedToolTime(),
		UpdatedAt:             fixedToolTime(),
	}
}

func toolFixtureDeliveryDocument() api.AgentDeliveryDocument {
	return api.AgentDeliveryDocument{
		Source:   "patchxnote",
		Version:  "1",
		Title:    "Webhook 测试记录",
		Summary:  "用于 MCP webhook 测试。",
		Markdown: "# Webhook 测试记录\n\n用于 MCP webhook 测试。\n",
		Memory: &api.AgentDeliveryMemory{
			ID:                 "mem_fixture_1",
			Platform:           "mobile",
			Source:             "structured_result",
			ObjectType:         "event",
			ClientObjectID:     "local_event_fixture",
			RevisionID:         "rev_fixture_1",
			Revision:           3,
			SchemaID:           "patchnote.event.v1",
			SchemaVersion:      1,
			SourceAvailability: "text_only",
		},
		Trace: api.AgentDeliveryTrace{
			RequestID: "mrun_fixture_1",
			Platform:  "mobile",
			TaskType:  "event",
			State:     "completed",
		},
		GeneratedAt: fixedToolTime(),
	}
}

func toolFixtureModelIOExport() api.AgentModelIOExport {
	return api.AgentModelIOExport{
		Source:  "patchxnote",
		Version: "1",
		Memory:  toolFixtureDeliveryDocument().Memory,
		Trace:   toolFixtureDeliveryDocument().Trace,
		SourceText: &api.AgentSourceText{
			Availability: "available",
			Text:         "转写安全文本",
		},
		FieldStatus: api.AgentModelIOFieldStatus{
			ClientRequestJSON:    "available",
			ProviderRequestJSON:  "available",
			ProviderResponseJSON: "available",
			ParsedResultJSON:     "available",
			PackagedResultJSON:   "available",
			ProviderAttemptsJSON: "available",
		},
		ClientRequestJSON:    json.RawMessage(`{"content":"客户端请求"}`),
		ProviderRequestJSON:  json.RawMessage(`{"content":"模型请求"}`),
		ProviderResponseJSON: json.RawMessage(`{"content":"模型原始响应"}`),
		ParsedResultJSON:     json.RawMessage(`{"content":"解析结果"}`),
		PackagedResultJSON:   json.RawMessage(`{"content":"封装结构"}`),
		ProviderAttemptsJSON: json.RawMessage(`[{"content":"供应商尝试"}]`),
	}
}

func fixedToolTime() time.Time {
	return time.Date(2026, 8, 6, 10, 0, 0, 0, time.UTC)
}
