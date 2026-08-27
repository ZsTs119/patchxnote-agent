package setup

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

func Registry() []Client {
	return []Client{
		{ID: "vscode", Name: "VS Code / GitHub Copilot", Priority: "P0", SupportStatus: "supported", ConfigFormat: "json", AutoWrite: true, RequiresRestart: "Refresh MCP servers or reload the VS Code window.", PrimaryStrategy: "config-merge"},
		{ID: "cursor", Name: "Cursor", Priority: "P0", SupportStatus: "supported", ConfigFormat: "json", AutoWrite: true, RequiresRestart: "Refresh MCP servers in Cursor.", PrimaryStrategy: "config-merge", DeeplinkTemplate: "cursor://anysphere.cursor-deeplink/mcp/install?name=%s&config=%s"},
		{ID: "codex", Name: "Codex / ChatGPT Desktop / Codex IDE", Priority: "P0", SupportStatus: "supported", ConfigFormat: "toml", AutoWrite: true, RequiresRestart: "Start a new Codex session.", PrimaryStrategy: "cli-command", CLICommand: "codex mcp add patchxnote -- npx -y patchxnote-agent@latest mcp serve"},
		{ID: "claude-code", Name: "Claude Code", Priority: "P0", SupportStatus: "manual", ConfigFormat: "client-cli", AutoWrite: false, RequiresRestart: "Start a new Claude Code session.", PrimaryStrategy: "cli-command", CLICommand: "claude mcp add patchxnote -- npx -y patchxnote-agent@latest mcp serve"},
		{ID: "claude-desktop", Name: "Claude Desktop", Priority: "P0", SupportStatus: "supported", ConfigFormat: "json", AutoWrite: true, RequiresRestart: "Restart Claude Desktop.", PrimaryStrategy: "config-merge"},
		{ID: "windsurf", Name: "Windsurf", Priority: "P0", SupportStatus: "supported", ConfigFormat: "json", AutoWrite: true, RequiresRestart: "Refresh MCP servers in Windsurf.", PrimaryStrategy: "config-merge"},
		{ID: "trae", Name: "Trae / Trae CN / TraeWork Code", Priority: "P0", SupportStatus: "manual", ConfigFormat: "ui-json", AutoWrite: false, RequiresRestart: "Refresh Trae MCP servers.", PrimaryStrategy: "manual-ui"},
		{ID: "qoder", Name: "Qoder", Priority: "P0", SupportStatus: "manual", ConfigFormat: "ui-json", AutoWrite: false, RequiresRestart: "Refresh Qoder MCP servers.", PrimaryStrategy: "deeplink", DeeplinkTemplate: "qoder://aicoding.aicoding-deeplink/mcp/add?name=%s&config=%s"},
		{ID: "workbuddy", Name: "WorkBuddy / Tencent CodeBuddy WorkBuddy", Priority: "P0", SupportStatus: "manual", ConfigFormat: "ui-json", AutoWrite: false, RequiresRestart: "Refresh WorkBuddy MCP connector.", PrimaryStrategy: "manual-ui"},
		{ID: "feishu-aily", Name: "Feishu Aily / Doubao Work Partner", Priority: "P0.5", SupportStatus: "planned", ConfigFormat: "platform-form", AutoWrite: false, RequiresRestart: "Refresh the platform connector.", PrimaryStrategy: "remote-url"},
		{ID: "tencent-agent-platform", Name: "Tencent Agent Development Platform / Enterprise WorkBuddy", Priority: "P0.5", SupportStatus: "planned", ConfigFormat: "platform-form", AutoWrite: false, RequiresRestart: "Refresh the platform connector.", PrimaryStrategy: "remote-url"},
		{ID: "jetbrains", Name: "JetBrains AI Assistant", Priority: "P1", SupportStatus: "planned", ConfigFormat: "ui-json", AutoWrite: false, RequiresRestart: "Refresh the IDE MCP server list.", PrimaryStrategy: "copy"},
		{ID: "zed", Name: "Zed", Priority: "P1", SupportStatus: "planned", ConfigFormat: "json", AutoWrite: false, RequiresRestart: "Reload Zed.", PrimaryStrategy: "copy"},
		{ID: "gemini-cli", Name: "Gemini CLI", Priority: "P1", SupportStatus: "planned", ConfigFormat: "json", AutoWrite: false, RequiresRestart: "Start a new Gemini CLI session.", PrimaryStrategy: "copy"},
		{ID: "qwen-code", Name: "Qwen Code", Priority: "P1", SupportStatus: "planned", ConfigFormat: "json", AutoWrite: false, RequiresRestart: "Start a new Qwen Code session.", PrimaryStrategy: "copy"},
		{ID: "kimi-code", Name: "Kimi Code / Kimi CLI", Priority: "P1", SupportStatus: "planned", ConfigFormat: "json", AutoWrite: false, RequiresRestart: "Start a new Kimi Code session.", PrimaryStrategy: "copy"},
		{ID: "opencode", Name: "OpenCode", Priority: "P1", SupportStatus: "planned", ConfigFormat: "json", AutoWrite: false, RequiresRestart: "Start a new OpenCode session.", PrimaryStrategy: "copy"},
		{ID: "vscode-derived-agents", Name: "Cline / Continue / Roo-derived VS Code agents", Priority: "P1", SupportStatus: "planned", ConfigFormat: "json", AutoWrite: false, RequiresRestart: "Refresh the VS Code extension MCP server list.", PrimaryStrategy: "copy"},
	}
}

func ClientIDs() []string {
	clients := Registry()
	ids := make([]string, 0, len(clients))
	for _, client := range clients {
		ids = append(ids, client.ID)
	}
	sort.Strings(ids)
	return ids
}

func FindClient(id string) (Client, bool) {
	id = strings.TrimSpace(strings.ToLower(id))
	for _, client := range Registry() {
		if client.ID == id {
			return client, true
		}
	}
	return Client{}, false
}

func LocalSupportedClientIDs() []string {
	var ids []string
	for _, client := range Registry() {
		if client.AutoWrite && (client.Priority == "P0" || client.Priority == "P0.5") {
			ids = append(ids, client.ID)
		}
	}
	sort.Strings(ids)
	return ids
}

func BuildPlan(clientID string, options PlanOptions) (ClientPlan, error) {
	client, ok := FindClient(clientID)
	if !ok {
		return ClientPlan{}, fmt.Errorf("unknown client %q; supported ids: %s", clientID, strings.Join(ClientIDs(), ", "))
	}
	server := DefaultServerConfig()
	plan := ClientPlan{
		Client:               client,
		Server:               server,
		ConfigSnippet:        MCPJSONSnippet(),
		CodexCommand:         commandForClient(client, "codex"),
		ClaudeCommand:        commandForClient(client, "claude-code"),
		RuntimeCredentialTip: runtimeCredentialTip(options.TargetOS),
	}
	if client.ID == "codex" {
		plan.ConfigSnippet = CodexTOMLSnippet()
	}
	if client.DeeplinkTemplate != "" {
		plan.Deeplink = buildDeeplink(client.DeeplinkTemplate, server)
	}
	if !client.AutoWrite {
		plan.ManualRequired = true
		plan.ManualReason = manualReason(client)
		return plan, nil
	}

	target, err := ResolveClientConfigTarget(client.ID, options.TargetOS, options.PathEnv)
	if err != nil {
		return ClientPlan{}, err
	}
	plan.ConfigPath = target.Path
	plan.ExpectedRoot = target.ExpectedRoot
	if target.Warning != "" {
		plan.Warnings = append(plan.Warnings, target.Warning)
	}
	if tip := runtimeCredentialTip(options.TargetOS); tip != "" {
		plan.Warnings = append(plan.Warnings, tip)
	}
	return plan, nil
}

func ApplyPlan(plan ClientPlan, options InstallOptions) (InstallResult, error) {
	result := InstallResult{
		Status:         "manual_required",
		ClientID:       plan.Client.ID,
		ConfigPath:     plan.ConfigPath,
		ManualRequired: plan.ManualRequired,
		ManualReason:   plan.ManualReason,
		Warnings:       append([]string(nil), plan.Warnings...),
	}
	if plan.ManualRequired {
		return result, nil
	}
	if options.DryRun {
		result.Status = "dry_run"
		return result, nil
	}

	var write WriteResult
	var err error
	switch plan.Client.ConfigFormat {
	case "json":
		write, err = WriteJSONMCPServer(plan.ConfigPath, plan.ExpectedRoot, ServerName, plan.Server, JSONWriteOptions{Force: options.Force})
	case "toml":
		write, err = WriteTOMLMCPServer(plan.ConfigPath, plan.ExpectedRoot, ServerName, plan.Server, TOMLWriteOptions{Force: options.Force})
	default:
		result.ManualRequired = true
		result.ManualReason = manualReason(plan.Client)
		return result, nil
	}
	if err != nil {
		return result, err
	}
	result.Status = "installed"
	result.BackupPath = write.BackupPath
	result.Changed = write.Changed
	if err := VerifyPlanConfig(plan); err != nil {
		_ = RollbackConfig(plan.ConfigPath, write.BackupPath)
		return result, fmt.Errorf("verify written config: %w", err)
	}
	result.Verified = true
	return result, nil
}

func commandForClient(client Client, id string) string {
	if client.ID == id {
		return client.CLICommand
	}
	return ""
}

func manualReason(client Client) string {
	if client.SupportStatus == "planned" {
		return "This client is tracked for the website and future setup, but automatic install is not enabled in V1."
	}
	if client.PrimaryStrategy == "remote-url" {
		return "This platform cannot run a local npx server; use the remote MCP gateway after platform authorization is available."
	}
	return "This client needs its official UI or CLI flow in V1, so setup prints a safe copyable config instead of editing local files."
}

func buildDeeplink(template string, server MCPServerConfig) string {
	payload, err := json.Marshal(server)
	if err != nil {
		return ""
	}
	encoded := base64.StdEncoding.EncodeToString(payload)
	return fmt.Sprintf(template, url.QueryEscape(ServerName), url.QueryEscape(encoded))
}

func runtimeCredentialTip(targetOS string) string {
	if strings.EqualFold(targetOS, "linux") && likelyWSL() {
		return "This looks like WSL. Windows desktop clients will not share Linux keychain credentials; run setup from Windows for Windows apps."
	}
	return ""
}
