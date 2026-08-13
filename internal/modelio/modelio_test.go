package modelio

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
)

func TestSelectFieldReturnsOnlyRequestedContent(t *testing.T) {
	export := fixtureExport()
	tests := []struct {
		field     Field
		wantKey   string
		wantValue string
	}{
		{FieldSourceText, "text", "转写安全文本"},
		{FieldProviderResponse, "json", "模型原始响应"},
		{FieldParsedResult, "json", "解析结果"},
		{FieldPackagedResult, "json", "封装结构"},
	}
	for _, tt := range tests {
		t.Run(string(tt.field), func(t *testing.T) {
			result, err := SelectField(export, Lookup{MemoryID: "mem_fixture_1", Platform: "mobile"}, tt.field)
			if err != nil {
				t.Fatalf("select field: %v", err)
			}
			if !result.Available || result.Status != "available" {
				t.Fatalf("expected available field, got %+v", result)
			}
			summary, err := result.Summary(true)
			if err != nil {
				t.Fatalf("summary: %v", err)
			}
			if _, ok := summary[tt.wantKey]; !ok {
				t.Fatalf("expected key %s in summary %+v", tt.wantKey, summary)
			}
			body, err := json.Marshal(summary)
			if err != nil {
				t.Fatalf("marshal summary: %v", err)
			}
			text := string(body)
			if !strings.Contains(text, tt.wantValue) {
				t.Fatalf("summary missed requested content %q:\n%s", tt.wantValue, text)
			}
			for _, forbidden := range []string{"客户端请求", "模型请求", "供应商尝试"} {
				if strings.Contains(text, forbidden) {
					t.Fatalf("summary leaked unrelated model IO field %q:\n%s", forbidden, text)
				}
			}
		})
	}
}

func TestSelectFieldUnavailableAndEmptyStates(t *testing.T) {
	export := fixtureExport()
	export.SourceText = nil
	source, err := SelectField(export, Lookup{MemoryID: "mem_fixture_1"}, FieldSourceText)
	if err != nil {
		t.Fatalf("select source: %v", err)
	}
	if source.Available || source.Status != "missing" {
		t.Fatalf("expected missing source text, got %+v", source)
	}

	export = fixtureExport()
	export.FieldStatus.ProviderResponseJSON = "available"
	export.ProviderResponseJSON = nil
	provider, err := SelectField(export, Lookup{MemoryID: "mem_fixture_1"}, FieldProviderResponse)
	if err != nil {
		t.Fatalf("select provider: %v", err)
	}
	if provider.Available || provider.Status != "empty" {
		t.Fatalf("expected empty provider response, got %+v", provider)
	}

	export = fixtureExport()
	export.FieldStatus.ParsedResultJSON = "unavailable"
	parsed, err := SelectField(export, Lookup{MemoryID: "mem_fixture_1"}, FieldParsedResult)
	if err != nil {
		t.Fatalf("select parsed: %v", err)
	}
	if parsed.Available || parsed.Status != "unavailable" || len(parsed.RawJSON) != 0 {
		t.Fatalf("expected unavailable parsed result without content, got %+v", parsed)
	}
}

func TestSelectFieldJSONEdgeShapes(t *testing.T) {
	for _, raw := range []json.RawMessage{
		json.RawMessage(`"scalar"`),
		json.RawMessage(`[1,2,3]`),
		json.RawMessage(`{"ok":true}`),
	} {
		export := fixtureExport()
		export.ProviderResponseJSON = raw
		result, err := SelectField(export, Lookup{MemoryID: "mem_fixture_1"}, FieldProviderResponse)
		if err != nil {
			t.Fatalf("select field: %v", err)
		}
		if !result.Available {
			t.Fatalf("expected raw JSON shape %s to be available", raw)
		}
		if _, err := result.InlineValue(); err != nil {
			t.Fatalf("inline value for %s: %v", raw, err)
		}
	}

	export := fixtureExport()
	export.ProviderResponseJSON = json.RawMessage(`null`)
	result, err := SelectField(export, Lookup{MemoryID: "mem_fixture_1"}, FieldProviderResponse)
	if err != nil {
		t.Fatalf("select null field: %v", err)
	}
	if result.Available || result.Status != "null" {
		t.Fatalf("expected null provider response, got %+v", result)
	}

	export = fixtureExport()
	export.ProviderResponseJSON = json.RawMessage(`{"bad"`)
	result, err = SelectField(export, Lookup{MemoryID: "mem_fixture_1"}, FieldProviderResponse)
	if err != nil {
		t.Fatalf("select invalid JSON field: %v", err)
	}
	if result.Available || result.Status != "invalid_json" {
		t.Fatalf("expected invalid JSON status, got %+v", result)
	}
}

func TestValidateLookupAndField(t *testing.T) {
	if _, err := ValidateField("unknown"); err == nil {
		t.Fatal("expected invalid field to fail")
	}
	for _, lookup := range []Lookup{
		{},
		{MemoryID: "mem", RequestID: "req"},
		{MemoryID: "mem", Platform: "web"},
	} {
		if err := ValidateLookup(lookup); err == nil {
			t.Fatalf("expected invalid lookup to fail: %+v", lookup)
		}
	}
	if err := ValidateLookup(Lookup{RequestID: "req", Platform: "desktop"}); err != nil {
		t.Fatalf("expected request lookup to pass: %v", err)
	}
}

func TestWriteFieldFileAndExportFile(t *testing.T) {
	result, err := SelectField(fixtureExport(), Lookup{MemoryID: "mem_fixture_1", Platform: "mobile"}, FieldProviderResponse)
	if err != nil {
		t.Fatalf("select field: %v", err)
	}
	out := filepath.Join(t.TempDir(), "模型 输出.json")
	written, err := WriteFieldFile(out, result, false)
	if err != nil {
		t.Fatalf("write field: %v", err)
	}
	if !written.Written || !filepath.IsAbs(written.OutputPath) {
		t.Fatalf("expected absolute written path, got %+v", written)
	}
	body, err := os.ReadFile(written.OutputPath)
	if err != nil {
		t.Fatalf("read written field: %v", err)
	}
	if !strings.Contains(string(body), "模型原始响应") || !strings.HasSuffix(string(body), "\n") {
		t.Fatalf("unexpected field file body:\n%s", body)
	}
	if _, err := WriteFieldFile(out, result, false); err == nil {
		t.Fatal("expected existing output without force to fail")
	}
	if _, err := WriteFieldFile(t.TempDir(), result, true); err == nil {
		t.Fatal("expected directory output path to fail")
	}

	exportOut := filepath.Join(t.TempDir(), "full.json")
	exportPath, err := WriteExportFile(exportOut, fixtureExport(), false)
	if err != nil {
		t.Fatalf("write export: %v", err)
	}
	if !filepath.IsAbs(exportPath) {
		t.Fatalf("expected absolute export path, got %s", exportPath)
	}
}

func fixtureExport() api.AgentModelIOExport {
	return api.AgentModelIOExport{
		Source:  "patchxnote-agent",
		Version: "v1",
		Memory: &api.AgentDeliveryMemory{
			ID:       "mem_fixture_1",
			Platform: "mobile",
		},
		Trace: api.AgentDeliveryTrace{
			RequestID: "req_fixture_1",
			Platform:  "mobile",
			State:     "succeeded",
		},
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
