package setup

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ZsTs119/patchxnote-agent/internal/config"
)

const ServerName = "patchxnote"

type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

type Client struct {
	ID               string
	Name             string
	Priority         string
	SupportStatus    string
	ConfigFormat     string
	AutoWrite        bool
	RequiresRestart  string
	PrimaryStrategy  string
	CLICommand       string
	DeeplinkTemplate string
}

type PlanOptions struct {
	TargetOS string
	PathEnv  config.PathEnv
}

type ClientPlan struct {
	Client               Client          `json:"client"`
	Server               MCPServerConfig `json:"server"`
	ConfigPath           string          `json:"config_path,omitempty"`
	ExpectedRoot         string          `json:"expected_root,omitempty"`
	ManualRequired       bool            `json:"manual_required"`
	ManualReason         string          `json:"manual_reason,omitempty"`
	ConfigSnippet        string          `json:"config_snippet"`
	CodexCommand         string          `json:"codex_command,omitempty"`
	ClaudeCommand        string          `json:"claude_command,omitempty"`
	Deeplink             string          `json:"deeplink,omitempty"`
	Warnings             []string        `json:"warnings,omitempty"`
	RuntimeCredentialTip string          `json:"runtime_credential_tip"`
}

type InstallOptions struct {
	DryRun bool
	Force  bool
}

type InstallResult struct {
	Status         string   `json:"status"`
	ClientID       string   `json:"client_id"`
	ConfigPath     string   `json:"config_path,omitempty"`
	BackupPath     string   `json:"backup_path,omitempty"`
	Changed        bool     `json:"changed"`
	ManualRequired bool     `json:"manual_required"`
	ManualReason   string   `json:"manual_reason,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
	Verified       bool     `json:"verified"`
}

type WriteResult struct {
	Changed    bool
	BackupPath string
}

func DefaultServerConfig() MCPServerConfig {
	return MCPServerConfig{
		Command: "npx",
		Args:    []string{"-y", "patchxnote-agent@latest", "mcp", "serve"},
	}
}

func MCPJSONSnippet() string {
	payload := map[string]map[string]MCPServerConfig{
		"mcpServers": {
			ServerName: DefaultServerConfig(),
		},
	}
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return ""
	}
	return string(encoded)
}

func CodexTOMLSnippet() string {
	server := DefaultServerConfig()
	return fmt.Sprintf("[mcp_servers.%s]\ncommand = %q\nargs = [%s]\n",
		ServerName,
		server.Command,
		quoteTOMLArray(server.Args),
	)
}

func quoteTOMLArray(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("%q", value))
	}
	return strings.Join(quoted, ", ")
}
