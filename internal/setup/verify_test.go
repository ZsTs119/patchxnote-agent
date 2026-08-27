package setup

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSmokeMCPCommand(t *testing.T) {
	result, err := SmokeMCPCommand(context.Background(), os.Args[0], []string{"-test.run=TestMCPHelperProcess", "--", "helper-mcp"}, 5*time.Second)
	if err != nil {
		t.Fatalf("smoke command: %v", err)
	}
	if !result.Initialized || !result.ToolsListed || result.ToolCount != 1 {
		t.Fatalf("unexpected smoke result: %+v", result)
	}
}

func TestMCPHelperProcess(t *testing.T) {
	if len(os.Args) == 0 || !strings.Contains(os.Args[0], ".test") || os.Args[len(os.Args)-1] != "helper-mcp" {
		return
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var request map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			continue
		}
		switch request["method"] {
		case "initialize":
			_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      request["id"],
				"result": map[string]any{
					"protocolVersion": "2025-06-18",
					"serverInfo":      map[string]any{"name": "helper", "version": "0.0.0"},
					"capabilities":    map[string]any{"tools": map[string]any{}},
				},
			})
		case "tools/list":
			_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      request["id"],
				"result":  map[string]any{"tools": []any{map[string]any{"name": "patchxnote_get_current_user"}}},
			})
		}
	}
	fmt.Fprintln(os.Stderr, "helper done")
	os.Exit(0)
}
