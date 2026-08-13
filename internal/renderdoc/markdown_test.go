package renderdoc

import "testing"

func TestFirstH1(t *testing.T) {
	if got := FirstH1("intro\n# 中文标题\n\n正文"); got != "中文标题" {
		t.Fatalf("expected unicode H1, got %q", got)
	}
	if got := FirstH1("## 二级标题\n正文"); got != "" {
		t.Fatalf("expected no H1, got %q", got)
	}
}

func TestInferTitle(t *testing.T) {
	tests := []struct {
		name       string
		markdown   string
		sourcePath string
		override   string
		want       string
	}{
		{name: "override wins", markdown: "# H1", sourcePath: "file.md", override: "自定义", want: "自定义"},
		{name: "h1 wins", markdown: "# H1 标题", sourcePath: "file.md", want: "H1 标题"},
		{name: "filename fallback", markdown: "正文", sourcePath: "/tmp/本周复盘.md", want: "本周复盘"},
		{name: "default fallback", markdown: "", sourcePath: "", want: DefaultTitle},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InferTitle(tt.markdown, tt.sourcePath, tt.override); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
