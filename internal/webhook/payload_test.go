package webhook

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestRenderFeishuPayload(t *testing.T) {
	target := Target{Alias: "产品群", Type: TargetTypeFeishu}
	rendered, err := RenderRequest(target, "https://example.test/hook", "", fixtureMessage(), time.Unix(1700000000, 0))
	if err != nil {
		t.Fatalf("render feishu: %v", err)
	}
	if rendered.ContentType != "application/json" {
		t.Fatalf("unexpected content type: %s", rendered.ContentType)
	}
	var payload map[string]any
	if err := json.Unmarshal(rendered.Body, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload["msg_type"] != "interactive" {
		t.Fatalf("expected interactive msg_type, got %+v", payload)
	}
	card := payload["card"].(map[string]any)
	elements := card["elements"].([]any)
	text := elements[0].(map[string]any)["text"].(map[string]any)
	if text["tag"] != "lark_md" || !strings.Contains(text["content"].(string), "合成内容") {
		t.Fatalf("unexpected lark_md text: %+v", text)
	}
	content := text["content"].(string)
	if strings.Contains(content, "# 合成标题") || strings.Contains(content, "## 摘要") {
		t.Fatalf("feishu lark_md content should not keep normal Markdown headings: %q", content)
	}
	if !strings.Contains(content, "**摘要**") {
		t.Fatalf("feishu lark_md content should convert headings to bold text: %q", content)
	}
	if _, ok := payload["sign"]; ok {
		t.Fatal("unsigned feishu payload must not include sign")
	}
}

func TestFeishuCardMarkdownKeepsCodeFenceHash(t *testing.T) {
	got := feishuCardMarkdown("# 标题\n\n## 摘要\n\n```text\n# 代码里的井号\n```\n", "标题")
	if strings.Contains(got, "# 标题") || strings.Contains(got, "## 摘要") {
		t.Fatalf("expected headings converted/removed, got:\n%s", got)
	}
	if !strings.Contains(got, "**摘要**") {
		t.Fatalf("expected heading converted to bold, got:\n%s", got)
	}
	if !strings.Contains(got, "# 代码里的井号") {
		t.Fatalf("expected code fence content preserved, got:\n%s", got)
	}
}

func TestRenderFeishuSignedPayload(t *testing.T) {
	rendered, err := RenderRequest(Target{Alias: "产品群", Type: TargetTypeFeishu}, "https://example.test/hook", "secret_fixture", fixtureMessage(), time.Unix(1700000000, 0))
	if err != nil {
		t.Fatalf("render signed feishu: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(rendered.Body, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload["timestamp"] != "1700000000" || payload["sign"] == "" {
		t.Fatalf("expected timestamp/sign fields, got %+v", payload)
	}
	sign, err := FeishuSign(1700000000, "secret_fixture")
	if err != nil {
		t.Fatalf("feishu sign: %v", err)
	}
	if payload["sign"] != sign {
		t.Fatalf("unexpected feishu sign")
	}
}

func TestRenderDingTalkPayloadAndSigning(t *testing.T) {
	rendered, err := RenderRequest(Target{Alias: "钉钉群", Type: TargetTypeDingTalk}, "https://example.test/robot/send?access_token=fixture", "secret_fixture", fixtureMessage(), time.UnixMilli(1700000000123))
	if err != nil {
		t.Fatalf("render dingtalk: %v", err)
	}
	parsed, err := url.Parse(rendered.URL)
	if err != nil {
		t.Fatalf("parse signed url: %v", err)
	}
	if parsed.Query().Get("timestamp") != "1700000000123" || parsed.Query().Get("sign") == "" {
		t.Fatalf("expected dingtalk timestamp/sign query, got %s", parsed.RawQuery)
	}
	var payload map[string]any
	if err := json.Unmarshal(rendered.Body, &payload); err != nil {
		t.Fatalf("decode dingtalk payload: %v", err)
	}
	if payload["msgtype"] != "markdown" {
		t.Fatalf("expected markdown msgtype, got %+v", payload)
	}
	markdown := payload["markdown"].(map[string]any)
	if markdown["title"] != "合成标题" || !strings.Contains(markdown["text"].(string), "## 摘要") {
		t.Fatalf("unexpected markdown payload: %+v", markdown)
	}
}

func TestRenderGenericPayload(t *testing.T) {
	rendered, err := RenderRequest(Target{Alias: "网关", Type: TargetTypeGeneric}, "https://example.test/hook", "", fixtureMessage(), time.Now())
	if err != nil {
		t.Fatalf("render generic: %v", err)
	}
	var payload struct {
		Source   string            `json:"source"`
		Version  string            `json:"version"`
		Title    string            `json:"title"`
		Markdown string            `json:"markdown"`
		Memory   *MessageMemory    `json:"memory"`
		Metadata map[string]string `json:"metadata"`
	}
	if err := json.Unmarshal(rendered.Body, &payload); err != nil {
		t.Fatalf("decode generic payload: %v", err)
	}
	if payload.Source != "patchxnote" || payload.Version != "1" || payload.Memory.ID != "mem_fixture" {
		t.Fatalf("unexpected generic payload: %+v", payload)
	}
	if payload.Metadata["source"] != "patchxnote" || payload.Metadata["template"] != "default" {
		t.Fatalf("unexpected metadata: %+v", payload.Metadata)
	}
}

func TestRenderRequestRejectsEmptyMarkdown(t *testing.T) {
	_, err := RenderRequest(Target{Alias: "产品群", Type: TargetTypeFeishu}, "https://example.test/hook", "", Message{Title: "empty"}, time.Now())
	if err == nil {
		t.Fatal("expected empty markdown to fail")
	}
}

func fixtureMessage() Message {
	return Message{
		Title:    "合成标题",
		Markdown: "# 合成标题\n\n## 摘要\n\n合成内容",
		Memory:   &MessageMemory{ID: "mem_fixture", Platform: "desktop"},
		Metadata: map[string]string{"source": "patchxnote", "template": "default"},
	}
}
