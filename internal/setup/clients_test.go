package setup

import (
	"strings"
	"testing"

	"github.com/ZsTs119/patchxnote-agent/internal/config"
)

func TestRegistryContainsPlannedClientSet(t *testing.T) {
	required := []string{
		"vscode",
		"cursor",
		"codex",
		"claude-code",
		"claude-desktop",
		"windsurf",
		"trae",
		"qoder",
		"workbuddy",
		"feishu-aily",
		"tencent-agent-platform",
		"jetbrains",
		"zed",
		"gemini-cli",
		"qwen-code",
		"kimi-code",
		"opencode",
		"vscode-derived-agents",
	}
	for _, id := range required {
		if _, ok := FindClient(id); !ok {
			t.Fatalf("missing client %s", id)
		}
	}
}

func TestBuildPlanForWritableAndManualClients(t *testing.T) {
	home := t.TempDir()
	plan, err := BuildPlan("cursor", PlanOptions{
		TargetOS: "linux",
		PathEnv:  config.PathEnv{HomeDir: home},
	})
	if err != nil {
		t.Fatalf("build cursor plan: %v", err)
	}
	if plan.ManualRequired {
		t.Fatal("cursor should be auto-writable")
	}
	if !strings.HasSuffix(plan.ConfigPath, "/.cursor/mcp.json") {
		t.Fatalf("unexpected cursor config path: %s", plan.ConfigPath)
	}
	if !strings.Contains(plan.ConfigSnippet, "patchxnote-agent@latest") {
		t.Fatalf("missing config snippet: %s", plan.ConfigSnippet)
	}
	if !strings.HasPrefix(plan.Deeplink, "cursor://") {
		t.Fatalf("missing cursor deeplink: %s", plan.Deeplink)
	}

	manual, err := BuildPlan("workbuddy", PlanOptions{TargetOS: "linux", PathEnv: config.PathEnv{HomeDir: home}})
	if err != nil {
		t.Fatalf("build workbuddy plan: %v", err)
	}
	if !manual.ManualRequired || manual.ManualReason == "" {
		t.Fatalf("expected manual workbuddy plan: %+v", manual)
	}
}

func TestBuildPlanRejectsUnknownClient(t *testing.T) {
	_, err := BuildPlan("missing-client", PlanOptions{TargetOS: "linux", PathEnv: config.PathEnv{HomeDir: t.TempDir()}})
	if err == nil {
		t.Fatal("expected unknown client error")
	}
	if !strings.Contains(err.Error(), "supported ids") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveClientConfigTargetsAcrossOS(t *testing.T) {
	windows, err := ResolveClientConfigTarget("vscode", "windows", config.PathEnv{
		HomeDir: "C:\\Users\\Alice Smith",
		AppData: "C:\\Users\\Alice Smith\\AppData\\Roaming",
	})
	if err != nil {
		t.Fatalf("windows target: %v", err)
	}
	if windows.Path != "C:\\Users\\Alice Smith\\AppData\\Roaming\\Code\\User\\mcp.json" {
		t.Fatalf("unexpected windows path: %s", windows.Path)
	}

	darwin, err := ResolveClientConfigTarget("claude-desktop", "darwin", config.PathEnv{HomeDir: "/Users/alice"})
	if err != nil {
		t.Fatalf("darwin target: %v", err)
	}
	if !strings.HasSuffix(darwin.Path, "/Library/Application Support/Claude/claude_desktop_config.json") {
		t.Fatalf("unexpected darwin path: %s", darwin.Path)
	}

	linux, err := ResolveClientConfigTarget("windsurf", "linux", config.PathEnv{HomeDir: "/home/alice"})
	if err != nil {
		t.Fatalf("linux target: %v", err)
	}
	if linux.Path != "/home/alice/.codeium/windsurf/mcp_config.json" {
		t.Fatalf("unexpected linux path: %s", linux.Path)
	}
}
