package cli

import (
	"encoding/json"
	"strings"
	"testing"
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
