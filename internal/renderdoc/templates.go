package renderdoc

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const MaxRenderedMarkdownBytes = 256 * 1024

//go:embed templates/*.tmpl
var builtInTemplates embed.FS

type TemplateInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func BuiltInTemplates() []TemplateInfo {
	return []TemplateInfo{
		{Name: "default", Description: "通用记录摘要，适合大多数 PatchXNote 记录。"},
		{Name: "meeting-summary", Description: "会议纪要格式，突出结论、事项和跟进。"},
		{Name: "daily-review", Description: "日复盘格式，适合当天记录整理。"},
		{Name: "key-items", Description: "只突出关键事项和待办。"},
		{Name: "raw-markdown", Description: "尽量保留服务端交付文档的 Markdown。"},
	}
}

func RenderTemplate(nameOrPath string, doc Document) (string, error) {
	nameOrPath = strings.TrimSpace(nameOrPath)
	if nameOrPath == "" {
		nameOrPath = "default"
	}

	tmpl, err := loadTemplate(nameOrPath)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, doc); err != nil {
		return "", fmt.Errorf("render template: %w", err)
	}
	if buf.Len() > MaxRenderedMarkdownBytes {
		return "", fmt.Errorf("rendered Markdown exceeds local safety cap")
	}
	return strings.TrimSpace(buf.String()) + "\n", nil
}

func loadTemplate(nameOrPath string) (*template.Template, error) {
	funcs := template.FuncMap{
		"trim": strings.TrimSpace,
	}
	if info, err := os.Stat(nameOrPath); err == nil && !info.IsDir() {
		body, err := os.ReadFile(nameOrPath)
		if err != nil {
			return nil, fmt.Errorf("read template: %w", err)
		}
		return template.New(filepath.Base(nameOrPath)).Funcs(funcs).Parse(string(body))
	}

	builtInName := strings.TrimSuffix(nameOrPath, ".tmpl") + ".tmpl"
	body, err := builtInTemplates.ReadFile("templates/" + builtInName)
	if err != nil {
		return nil, fmt.Errorf("template %q not found", nameOrPath)
	}
	return template.New(builtInName).Funcs(funcs).Parse(string(body))
}
