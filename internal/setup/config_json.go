package setup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

type JSONWriteOptions struct {
	Force bool
	Now   func() time.Time
}

type jsonMCPDocument struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

func WriteJSONMCPServer(configPath string, expectedRoot string, name string, server MCPServerConfig, options JSONWriteOptions) (WriteResult, error) {
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
		if looksLikeJSONC(original) {
			return WriteResult{}, fmt.Errorf("%s appears to be JSONC; use manual setup so comments are preserved", configPath)
		}
		if err := json.Unmarshal(original, &document); err != nil {
			return WriteResult{}, fmt.Errorf("decode %s: %w", configPath, err)
		}
	}

	servers, err := jsonMCPServers(document)
	if err != nil {
		return WriteResult{}, err
	}
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
	document["mcpServers"] = servers
	next, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return WriteResult{}, fmt.Errorf("encode %s: %w", configPath, err)
	}
	next = append(next, '\n')

	return atomicWriteWithBackup(configPath, original, next, options.Now)
}

func jsonMCPServers(document map[string]any) (map[string]any, error) {
	raw, ok := document["mcpServers"]
	if !ok || raw == nil {
		return map[string]any{}, nil
	}
	servers, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("mcpServers must be an object")
	}
	return servers, nil
}

func serverToMap(server MCPServerConfig) (map[string]any, error) {
	encoded, err := json.Marshal(server)
	if err != nil {
		return nil, err
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

func atomicWriteWithBackup(configPath string, original []byte, next []byte, now func() time.Time) (WriteResult, error) {
	if bytes.Equal(original, next) {
		return WriteResult{}, nil
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return WriteResult{}, fmt.Errorf("create config directory: %w", err)
	}

	var backupPath string
	if len(original) > 0 {
		if now == nil {
			now = time.Now
		}
		backupPath = fmt.Sprintf("%s.bak-%s", configPath, now().UTC().Format("20060102T150405Z"))
		if err := os.WriteFile(backupPath, original, 0600); err != nil {
			return WriteResult{}, fmt.Errorf("write backup: %w", err)
		}
	}

	tempPath := fmt.Sprintf("%s.tmp-%d", configPath, os.Getpid())
	if err := os.WriteFile(tempPath, next, 0600); err != nil {
		return WriteResult{}, fmt.Errorf("write temp config: %w", err)
	}
	if err := os.Rename(tempPath, configPath); err != nil {
		_ = os.Remove(tempPath)
		if backupPath != "" {
			_ = os.WriteFile(configPath, original, 0600)
		}
		return WriteResult{}, fmt.Errorf("replace config: %w", err)
	}
	return WriteResult{Changed: true, BackupPath: backupPath}, nil
}

func RollbackConfig(configPath string, backupPath string) error {
	if backupPath == "" {
		if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	backup, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, backup, 0600)
}

func validateWriteTarget(configPath string, expectedRoot string) error {
	if configPath == "" {
		return fmt.Errorf("config path is required")
	}
	if expectedRoot == "" {
		return fmt.Errorf("expected config root is required")
	}
	if info, err := os.Lstat(configPath); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to write symlinked config path: %s", configPath)
	}
	targetOS := runningTargetOS()
	if !isPathUnder(targetOS, configPath, expectedRoot) {
		return fmt.Errorf("refusing to write outside expected config root: %s", configPath)
	}
	return nil
}
