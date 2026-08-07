package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ZsTs119/patchxnote-agent/internal/version"
)

func TestMCPServeWritesJSONRPCToStdoutOnly(t *testing.T) {
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
	}, "\n") + "\n"

	stdout, stderr, err := executeForTestWithDeps(t, Deps{Stdin: strings.NewReader(input)}, "mcp", "serve")
	if err != nil {
		t.Fatalf("mcp serve: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr diagnostics, got %q", stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected two JSON-RPC responses, got %d\n%s", len(lines), stdout)
	}
	for _, line := range lines {
		var response map[string]any
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			t.Fatalf("stdout line is not JSON: %v\n%s", err, line)
		}
		if response["jsonrpc"] != "2.0" {
			t.Fatalf("unexpected JSON-RPC marker: %+v", response)
		}
	}
}

func TestMCPServeInitializeReportsBuildVersion(t *testing.T) {
	originalVersion := version.Version
	version.Version = "9.8.7-cli-test"
	defer func() {
		version.Version = originalVersion
	}()

	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}` + "\n"

	stdout, stderr, err := executeForTestWithDeps(t, Deps{Stdin: strings.NewReader(input)}, "mcp", "serve")
	if err != nil {
		t.Fatalf("mcp serve: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr diagnostics, got %q", stderr)
	}

	var response map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &response); err != nil {
		t.Fatalf("stdout line is not JSON: %v\n%s", err, stdout)
	}
	result := response["result"].(map[string]any)
	serverInfo := result["serverInfo"].(map[string]any)
	if serverInfo["version"] != "9.8.7-cli-test" {
		t.Fatalf("expected MCP server info version to match build version, got %+v", serverInfo["version"])
	}
}
