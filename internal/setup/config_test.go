package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteJSONMCPServerCreatesAndVerifiesConfig(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Code User")
	configPath := filepath.Join(root, "mcp.json")
	result, err := WriteJSONMCPServer(configPath, root, ServerName, DefaultServerConfig(), JSONWriteOptions{})
	if err != nil {
		t.Fatalf("write json config: %v", err)
	}
	if !result.Changed || result.BackupPath != "" {
		t.Fatalf("unexpected write result: %+v", result)
	}

	plan := ClientPlan{
		Client:     Client{ID: "vscode", ConfigFormat: "json"},
		Server:     DefaultServerConfig(),
		ConfigPath: configPath,
	}
	if err := VerifyPlanConfig(plan); err != nil {
		t.Fatalf("verify json config: %v", err)
	}
}

func TestWriteJSONMCPServerPreservesUnrelatedServersAndCreatesBackup(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "mcp.json")
	original := `{
  "theme": "dark",
  "mcpServers": {
    "github": {
      "type": "sse",
      "url": "https://example.invalid/sse"
    }
  }
}
`
	if err := os.WriteFile(configPath, []byte(original), 0600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	now := func() time.Time { return time.Date(2026, 8, 27, 8, 0, 0, 0, time.UTC) }
	result, err := WriteJSONMCPServer(configPath, root, ServerName, DefaultServerConfig(), JSONWriteOptions{Now: now})
	if err != nil {
		t.Fatalf("write json config: %v", err)
	}
	if !strings.HasSuffix(result.BackupPath, ".bak-20260827T080000Z") {
		t.Fatalf("unexpected backup path: %s", result.BackupPath)
	}
	if backup, err := os.ReadFile(result.BackupPath); err != nil || string(backup) != original {
		t.Fatalf("backup mismatch err=%v", err)
	}

	var decoded map[string]any
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode config: %v\n%s", err, raw)
	}
	servers := decoded["mcpServers"].(map[string]any)
	if _, ok := servers["github"].(map[string]any)["url"]; !ok {
		t.Fatalf("unrelated server was not preserved: %+v", servers["github"])
	}
	if _, ok := servers[ServerName]; !ok {
		t.Fatalf("patchxnote server missing: %+v", servers)
	}
}

func TestWriteJSONMCPServerRejectsConflictUnlessForce(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "mcp.json")
	if err := os.WriteFile(configPath, []byte(`{"mcpServers":{"patchxnote":{"command":"other"}}}`), 0600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	_, err := WriteJSONMCPServer(configPath, root, ServerName, DefaultServerConfig(), JSONWriteOptions{})
	if err == nil {
		t.Fatal("expected conflict")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := WriteJSONMCPServer(configPath, root, ServerName, DefaultServerConfig(), JSONWriteOptions{Force: true}); err != nil {
		t.Fatalf("force write: %v", err)
	}
}

func TestWriteJSONMCPServerRejectsJSONC(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "mcp.json")
	if err := os.WriteFile(configPath, []byte("{\n  // keep this comment\n  \"mcpServers\": {}\n}\n"), 0600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	_, err := WriteJSONMCPServer(configPath, root, ServerName, DefaultServerConfig(), JSONWriteOptions{})
	if err == nil || !strings.Contains(err.Error(), "JSONC") {
		t.Fatalf("expected JSONC manual-mode error, got %v", err)
	}
}

func TestWriteJSONMCPServerRejectsSymlinkTarget(t *testing.T) {
	root := t.TempDir()
	realPath := filepath.Join(root, "real.json")
	linkPath := filepath.Join(root, "mcp.json")
	if err := os.WriteFile(realPath, []byte(`{}`), 0600); err != nil {
		t.Fatalf("write real file: %v", err)
	}
	if err := os.Symlink(realPath, linkPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	_, err := WriteJSONMCPServer(linkPath, root, ServerName, DefaultServerConfig(), JSONWriteOptions{})
	if err == nil || !strings.Contains(err.Error(), "symlinked") {
		t.Fatalf("expected symlink error, got %v", err)
	}
}

func TestWriteTOMLMCPServerPreservesExistingTables(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.toml")
	original := "model = \"gpt-5\"\n\n[mcp_servers.github]\ncommand = \"gh\"\nargs = [\"mcp\"]\n"
	if err := os.WriteFile(configPath, []byte(original), 0600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	result, err := WriteTOMLMCPServer(configPath, root, ServerName, DefaultServerConfig(), TOMLWriteOptions{
		Now: func() time.Time { return time.Date(2026, 8, 27, 8, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("write toml config: %v", err)
	}
	if result.BackupPath == "" {
		t.Fatal("expected backup")
	}
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	text := string(raw)
	for _, want := range []string{"model = 'gpt-5'", "[mcp_servers.github]", "[mcp_servers.patchxnote]", "patchxnote-agent@latest"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected TOML to contain %q:\n%s", want, text)
		}
	}
}

func TestApplyPlanDryRunDoesNotWrite(t *testing.T) {
	root := t.TempDir()
	plan := ClientPlan{
		Client:       Client{ID: "vscode", ConfigFormat: "json"},
		Server:       DefaultServerConfig(),
		ConfigPath:   filepath.Join(root, "mcp.json"),
		ExpectedRoot: root,
	}
	result, err := ApplyPlan(plan, InstallOptions{DryRun: true})
	if err != nil {
		t.Fatalf("apply dry-run: %v", err)
	}
	if result.Status != "dry_run" || result.Changed {
		t.Fatalf("unexpected dry-run result: %+v", result)
	}
	if _, err := os.Stat(plan.ConfigPath); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create config, stat err=%v", err)
	}
}

func TestRollbackConfigRestoresBackup(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "mcp.json")
	backupPath := configPath + ".bak"
	if err := os.WriteFile(configPath, []byte(`{"changed":true}`), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(backupPath, []byte(`{"original":true}`), 0600); err != nil {
		t.Fatalf("write backup: %v", err)
	}
	if err := RollbackConfig(configPath, backupPath); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read restored config: %v", err)
	}
	if string(raw) != `{"original":true}` {
		t.Fatalf("unexpected restored config: %s", raw)
	}
}
