package setup

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type TOMLWriteOptions struct {
	Force bool
	Now   func() time.Time
}

func WriteTOMLMCPServer(configPath string, expectedRoot string, name string, server MCPServerConfig, options TOMLWriteOptions) (WriteResult, error) {
	if err := validateWriteTarget(configPath, expectedRoot); err != nil {
		return WriteResult{}, err
	}

	document := map[string]any{}
	original, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return WriteResult{}, fmt.Errorf("read %s: %w", configPath, err)
		}
	} else if len(bytes.TrimSpace(original)) > 0 {
		if err := toml.Unmarshal(original, &document); err != nil {
			return WriteResult{}, fmt.Errorf("decode %s: %w", configPath, err)
		}
	}

	servers := tomlMCPServers(document)
	desired, err := serverToMap(server)
	if err != nil {
		return WriteResult{}, err
	}
	existing, exists := servers[name]
	if exists && !reflect.DeepEqual(existing, desired) && !options.Force {
		return WriteResult{}, fmt.Errorf("%s already has a different %q MCP server; rerun with --force to replace only that entry", configPath, name)
	}
	if exists && reflect.DeepEqual(existing, desired) {
		return WriteResult{}, nil
	}

	servers[name] = desired
	document["mcp_servers"] = servers
	next, err := toml.Marshal(document)
	if err != nil {
		return WriteResult{}, fmt.Errorf("encode %s: %w", configPath, err)
	}
	return atomicWriteWithBackup(configPath, original, next, options.Now)
}

func tomlMCPServers(document map[string]any) map[string]any {
	raw, ok := document["mcp_servers"]
	if !ok || raw == nil {
		return map[string]any{}
	}
	servers, ok := raw.(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	if !ok {
		return map[string]any{}
	}
	return servers
}
