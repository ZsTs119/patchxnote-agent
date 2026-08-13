package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/keychain"
)

func TestModelIOFieldCommands(t *testing.T) {
	store := keychain.NewMemoryStore()
	seedCredential(t, store)
	deps := webhookTestDeps(t, store, nil, &fakeAgentAPI{modelIO: cliModelIOFixture()})

	stdout, stderr, err := executeForTestWithDeps(t, deps, "model-io", "source-text", "--memory-id", "mem_fixture", "--platform", "desktop")
	if err != nil {
		t.Fatalf("source-text: %v", err)
	}
	if stdout != "转写安全文本\n" || stderr != "" {
		t.Fatalf("unexpected source-text stdout=%q stderr=%q", stdout, stderr)
	}

	stdout, stderr, err = executeForTestWithDeps(t, deps, "--output", "json", "model-io", "provider-response", "--memory-id", "mem_fixture", "--platform", "desktop")
	if err != nil {
		t.Fatalf("provider-response json: %v", err)
	}
	if strings.Contains(stdout+stderr, "客户端请求") || !strings.Contains(stdout, `"json": {`) || !strings.Contains(stdout, "模型原始响应") {
		t.Fatalf("provider response output leaked unrelated fields or double-encoded JSON:\nstdout=%s\nstderr=%s", stdout, stderr)
	}

	out := filepath.Join(t.TempDir(), "解析 结果.json")
	stdout, stderr, err = executeForTestWithDeps(t, deps, "model-io", "parsed-result", "--memory-id", "mem_fixture", "--platform", "desktop", "--out", out)
	if err != nil {
		t.Fatalf("parsed-result out: %v", err)
	}
	if strings.Contains(stdout+stderr, "解析结果") {
		t.Fatalf("out summary leaked field content: stdout=%q stderr=%q", stdout, stderr)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read parsed result: %v", err)
	}
	if !json.Valid(body) || !strings.Contains(string(body), "解析结果") {
		t.Fatalf("unexpected parsed result file:\n%s", body)
	}

	stdout, _, err = executeForTestWithDeps(t, deps, "model-io", "packaged-result", "--request-id", "mrun_fixture", "--platform", "desktop")
	if err != nil {
		t.Fatalf("packaged-result by request id: %v", err)
	}
	if !strings.Contains(stdout, "封装结构") {
		t.Fatalf("unexpected packaged-result output:\n%s", stdout)
	}
}

func TestModelIOExportCommand(t *testing.T) {
	store := keychain.NewMemoryStore()
	seedCredential(t, store)
	deps := webhookTestDeps(t, store, nil, &fakeAgentAPI{modelIO: cliModelIOFixture()})
	out := filepath.Join(t.TempDir(), "model-io.json")
	stdout, stderr, err := executeForTestWithDeps(t, deps, "model-io", "export", "--memory-id", "mem_fixture", "--platform", "desktop", "--out", out)
	if err != nil {
		t.Fatalf("model-io export: %v", err)
	}
	if strings.Contains(stdout+stderr, "转写安全文本") {
		t.Fatalf("export summary leaked model IO content: stdout=%q stderr=%q", stdout, stderr)
	}
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read export: %v", err)
	}
	if !strings.Contains(string(body), "模型原始响应") {
		t.Fatalf("unexpected export body:\n%s", body)
	}
}

func TestModelIOListCommand(t *testing.T) {
	store := keychain.NewMemoryStore()
	seedCredential(t, store)
	fakeAPI := &fakeAgentAPI{tracePage: api.AgentModelIOTracePage{
		Items: []api.AgentModelIOTraceSummary{
			{
				RequestID:              "mrun_fixture_trace_1",
				Platform:               "mobile",
				TaskType:               "daily_digest",
				State:                  "completed",
				CreatedAt:              fixedCLITime(),
				UpdatedAt:              fixedCLITime(),
				SourceTextAvailability: "available",
				FieldStatus: api.AgentModelIOFieldStatus{
					ProviderResponseJSON: "available",
					ParsedResultJSON:     "available",
					PackagedResultJSON:   "missing",
				},
				FieldBytes: api.AgentModelIOFieldBytes{ProviderResponseJSON: 256},
				Memory:     &api.AgentDeliveryMemory{ID: "mem_fixture_1", Platform: "mobile"},
			},
		},
		NextCursor: "cursor_trace_page_2",
	}}
	deps := webhookTestDeps(t, store, nil, fakeAPI)
	stdout, stderr, err := executeForTestWithDeps(t, deps, "model-io", "list", "--platform", "mobile", "--task-type", "daily_digest", "--business-id", "business-fixture", "--date-from", "2026-08-13T00:00:00Z", "--limit", "10", "--cursor", "cursor_trace_page_1")
	if err != nil {
		t.Fatalf("model-io list: %v", err)
	}
	if stderr != "" || !strings.Contains(stdout, "mrun_fixture_trace_1") || !strings.Contains(stdout, "next_cursor cursor_trace_page_2") ||
		strings.Contains(stdout, "模型原始响应") || strings.Contains(stdout, "转写安全文本") {
		t.Fatalf("unexpected list output stdout=%q stderr=%q", stdout, stderr)
	}
	if fakeAPI.traceParams.Platform != "mobile" || fakeAPI.traceParams.TaskType != "daily_digest" ||
		fakeAPI.traceParams.BusinessID != "business-fixture" || fakeAPI.traceParams.Limit != 10 ||
		fakeAPI.traceParams.Cursor != "cursor_trace_page_1" {
		t.Fatalf("unexpected list params: %+v", fakeAPI.traceParams)
	}

	stdout, stderr, err = executeForTestWithDeps(t, deps, "--output", "json", "model-io", "list", "--platform", "mobile")
	if err != nil {
		t.Fatalf("model-io list json: %v", err)
	}
	if stderr != "" || !strings.Contains(stdout, `"request_id": "mrun_fixture_trace_1"`) ||
		!strings.Contains(stdout, `"next_cursor": "cursor_trace_page_2"`) {
		t.Fatalf("unexpected json list output stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestModelIOCommandValidationAndUnavailable(t *testing.T) {
	store := keychain.NewMemoryStore()
	seedCredential(t, store)
	deps := webhookTestDeps(t, store, nil, &fakeAgentAPI{modelIO: cliModelIOFixture()})
	if _, _, err := executeForTestWithDeps(t, deps, "model-io", "source-text"); err == nil {
		t.Fatal("expected missing lookup to fail")
	}
	if _, _, err := executeForTestWithDeps(t, deps, "model-io", "source-text", "--memory-id", "mem_fixture", "--request-id", "mrun_fixture"); err == nil {
		t.Fatal("expected duplicate lookup to fail")
	}

	unavailable := cliModelIOFixture()
	unavailable.SourceText = nil
	stdout, _, err := executeForTestWithDeps(t, webhookTestDeps(t, store, nil, &fakeAgentAPI{modelIO: unavailable}), "model-io", "source-text", "--memory-id", "mem_fixture")
	if err != nil {
		t.Fatalf("unavailable source text should be a clean result: %v", err)
	}
	if !strings.Contains(stdout, "model io field unavailable") || !strings.Contains(stdout, "missing") {
		t.Fatalf("unexpected unavailable output:\n%s", stdout)
	}
}

func TestModelIOInlineLargeContentRequiresOut(t *testing.T) {
	store := keychain.NewMemoryStore()
	seedCredential(t, store)
	large := cliModelIOFixture()
	large.ProviderResponseJSON = []byte(`{"content":"` + strings.Repeat("x", 20*1024) + `"}`)
	deps := webhookTestDeps(t, store, nil, &fakeAgentAPI{modelIO: large})
	if _, _, err := executeForTestWithDeps(t, deps, "model-io", "provider-response", "--memory-id", "mem_fixture"); err == nil {
		t.Fatal("expected large inline provider response to fail")
	}
	out := filepath.Join(t.TempDir(), "large.json")
	if _, _, err := executeForTestWithDeps(t, deps, "model-io", "provider-response", "--memory-id", "mem_fixture", "--out", out); err != nil {
		t.Fatalf("large provider response with out: %v", err)
	}
}

func cliModelIOFixture() api.AgentModelIOExport {
	return api.AgentModelIOExport{
		Source:  "patchxnote",
		Version: "1",
		Memory:  &api.AgentDeliveryMemory{ID: "mem_fixture", Platform: "desktop"},
		Trace:   api.AgentDeliveryTrace{RequestID: "mrun_fixture", Platform: "desktop", State: "completed"},
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
		ClientRequestJSON:    []byte(`{"content":"客户端请求"}`),
		ProviderRequestJSON:  []byte(`{"content":"模型请求"}`),
		ProviderResponseJSON: []byte(`{"content":"模型原始响应"}`),
		ParsedResultJSON:     []byte(`{"content":"解析结果"}`),
		PackagedResultJSON:   []byte(`{"content":"封装结构"}`),
		ProviderAttemptsJSON: []byte(`[{"content":"供应商尝试"}]`),
	}
}
