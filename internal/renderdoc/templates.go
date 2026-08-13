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
