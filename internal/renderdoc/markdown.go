package renderdoc

import (
	"path/filepath"
	"strings"
)

const DefaultTitle = "PatchXNote 记录"

func FirstH1(markdown string) string {
	lines := strings.Split(markdown, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "# ") {
			continue
		}
		title := strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
		if title != "" {
			return title
		}
	}
	return ""
}

func InferTitle(markdown string, sourcePath string, override string) string {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override)
	}
	if title := FirstH1(markdown); title != "" {
		return title
	}
	if strings.TrimSpace(sourcePath) != "" {
		name := filepath.Base(sourcePath)
		if ext := filepath.Ext(name); ext != "" {
			name = strings.TrimSuffix(name, ext)
		}
		if strings.TrimSpace(name) != "" && name != "." && name != string(filepath.Separator) {
			return name
		}
	}
	return DefaultTitle
}
