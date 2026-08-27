package setup

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type MCPProtocolSmokeResult struct {
	Initialized bool `json:"initialized"`
	ToolsListed bool `json:"tools_listed"`
	ToolCount   int  `json:"tool_count"`
}

func VerifyPlanConfig(plan ClientPlan) error {
	if plan.ManualRequired {
		return nil
	}
	switch plan.Client.ConfigFormat {
	case "json":
		return verifyJSONConfig(plan.ConfigPath, ServerName, plan.Server)
	case "toml":
		return verifyTOMLConfig(plan.ConfigPath, ServerName, plan.Server)
	default:
		return fmt.Errorf("unsupported config format %q", plan.Client.ConfigFormat)
	}
}

func SmokeMCPCommand(ctx context.Context, command string, args []string, timeout time.Duration) (MCPProtocolSmokeResult, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return MCPProtocolSmokeResult{}, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return MCPProtocolSmokeResult{}, err
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return MCPProtocolSmokeResult{}, err
	}
	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"patchxnote-setup","version":"0.0.0"}}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
	}
	for _, request := range requests {
		if _, err := fmt.Fprintln(stdin, request); err != nil {
			_ = cmd.Process.Kill()
			return MCPProtocolSmokeResult{}, err
		}
	}
	_ = stdin.Close()

	var result MCPProtocolSmokeResult
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		var response struct {
			ID     int             `json:"id"`
			Result json.RawMessage `json:"result"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &response); err != nil {
			continue
		}
		switch response.ID {
		case 1:
			result.Initialized = true
		case 2:
			var list struct {
				Tools []json.RawMessage `json:"tools"`
			}
			if err := json.Unmarshal(response.Result, &list); err == nil {
				result.ToolsListed = true
				result.ToolCount = len(list.Tools)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		_ = cmd.Process.Kill()
		return result, err
	}
	if err := cmd.Wait(); err != nil {
		return result, err
	}
	if !result.Initialized || !result.ToolsListed {
		return result, fmt.Errorf("mcp smoke did not complete initialize/tools-list")
	}
	return result, nil
}

func verifyJSONConfig(configPath string, name string, server MCPServerConfig) error {
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	document := map[string]any{}
	if err := json.Unmarshal(raw, &document); err != nil {
		return err
	}
	servers, err := jsonMCPServers(document)
	if err != nil {
		return err
	}
	existing, ok := servers[name]
	if !ok {
		return fmt.Errorf("missing %q MCP server", name)
	}
	desired, err := serverToMap(server)
	if err != nil {
		return err
	}
	if !jsonEquivalent(existing, desired) {
		return fmt.Errorf("%q MCP server does not match expected command", name)
	}
	return nil
}

func verifyTOMLConfig(configPath string, name string, server MCPServerConfig) error {
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	document := map[string]any{}
	if err := toml.Unmarshal(raw, &document); err != nil {
		return err
	}
	servers := tomlMCPServers(document)
	existing, ok := servers[name]
	if !ok {
		return fmt.Errorf("missing %q MCP server", name)
	}
	desired, err := serverToMap(server)
	if err != nil {
		return err
	}
	if !jsonEquivalent(existing, desired) {
		return fmt.Errorf("%q MCP server does not match expected command", name)
	}
	return nil
}

func jsonEquivalent(left any, right any) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftJSON) == string(rightJSON)
}
