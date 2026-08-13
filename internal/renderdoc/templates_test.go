package renderdoc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuiltInTemplatesRender(t *testing.T) {
	doc := fixtureDocument()
	for _, name := range []string{"default", "meeting-summary", "daily-review", "key-items", "raw-markdown"} {
		t.Run(name, func(t *testing.T) {
			got, err := RenderTemplate(name, doc)
			if err != nil {
				t.Fatalf("render template %s: %v", name, err)
			}
			if !strings.Contains(got, "合成会议纪要") && name != "raw-markdown" {
				t.Fatalf("rendered template missing title:\n%s", got)
			}
			if name != "raw-markdown" && !strings.Contains(got, "##") {
				t.Fatalf("rendered template missing heading structure:\n%s", got)
			}
		})
	}
}

func TestLocalTemplatePathRenders(t *testing.T) {
	path := filepath.Join(t.TempDir(), "custom.tmpl")
	if err := os.WriteFile(path, []byte("# {{.Title}}\n\n{{.Summary}}\n"), 0o600); err != nil {
		t.Fatalf("write template: %v", err)
	}
	got, err := RenderTemplate(path, fixtureDocument())
	if err != nil {
		t.Fatalf("render local template: %v", err)
	}
	if !strings.Contains(got, "合成会议纪要") || !strings.Contains(got, "合成摘要") {
		t.Fatalf("unexpected local render:\n%s", got)
	}
}

func TestMissingTemplateFails(t *testing.T) {
	if _, err := RenderTemplate("missing-template", fixtureDocument()); err == nil {
		t.Fatal("expected missing template error")
	}
}

func TestRenderedMarkdownSafetyCap(t *testing.T) {
	doc := fixtureDocument()
	doc.Summary = strings.Repeat("x", MaxRenderedMarkdownBytes+1)
	_, err := RenderTemplate("default", doc)
	if err == nil {
		t.Fatal("expected safety cap error")
	}
	if !strings.Contains(err.Error(), "safety cap") {
		t.Fatalf("unexpected safety cap error: %v", err)
	}
}

func fixtureDocument() Document {
	return Document{
		Title:    "合成会议纪要",
		Summary:  "合成摘要",
		Markdown: "# 合成会议纪要\n\n正文",
		Sections: []Section{
			{Title: "关键结论", Markdown: "- 保持 Agent 只读。"},
		},
		KeyItems: []KeyItem{
			{Title: "完成 CLI 联调", Status: "open"},
		},
	}
}
