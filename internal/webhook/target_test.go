package webhook

import (
	"strings"
	"testing"
)

func TestValidateAlias(t *testing.T) {
	tests := []struct {
		name    string
		alias   string
		want    string
		wantErr bool
	}{
		{name: "chinese alias", alias: " 产品群 飞书 ", want: "产品群 飞书"},
		{name: "spaces and punctuation", alias: "研发, 周报 - A", want: "研发, 周报 - A"},
		{name: "empty", alias: "  ", wantErr: true},
		{name: "newline", alias: "产品\n群", wantErr: true},
		{name: "tab", alias: "产品\t群", wantErr: true},
		{name: "path separator", alias: "产品/飞书", wantErr: true},
		{name: "too long", alias: strings.Repeat("你", 65), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateAlias(tt.alias)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("validate alias: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestValidateType(t *testing.T) {
	for _, value := range []string{"feishu", "dingtalk", "generic", " FEISHU "} {
		if _, err := ValidateType(value); err != nil {
			t.Fatalf("expected type %q to pass: %v", value, err)
		}
	}
	if _, err := ValidateType("slack"); err == nil {
		t.Fatal("expected unknown target type to fail")
	}
}

func TestValidateWebhookURL(t *testing.T) {
	valid := []string{
		"http://127.0.0.1:8080/hook",
		"https://open.feishu.cn/open-apis/bot/v2/hook/token_fixture",
	}
	for _, value := range valid {
		if _, err := ValidateWebhookURL(value); err != nil {
			t.Fatalf("expected URL %q to pass: %v", value, err)
		}
	}

	invalid := []string{
		"open.feishu.cn/hook",
		"ftp://example.test/hook",
		"https:///hook",
		"https://example.test/hook#fragment",
		"https://example.test/hook\n",
	}
	for _, value := range invalid {
		if _, err := ValidateWebhookURL(value); err == nil {
			t.Fatalf("expected URL %q to fail", value)
		}
	}
}

func TestMaskWebhookURL(t *testing.T) {
	masked := MaskWebhookURL("https://open.feishu.cn/open-apis/bot/v2/hook/token-fixture-abcdefghijklmnopqrstuvwxyz?secret=abcdef")
	if strings.Contains(masked, "abcdefghijklmnopqrstuvwxyz") || strings.Contains(masked, "abcdef") {
		t.Fatalf("masked URL leaked token-like material: %s", masked)
	}
	if !strings.Contains(masked, "open.feishu.cn") || !strings.Contains(masked, "secret=%2A%2A%2A") {
		t.Fatalf("unexpected masked URL: %s", masked)
	}
}
