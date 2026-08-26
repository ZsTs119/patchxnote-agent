#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TMP_DIR="$ROOT/.tmp/e2e"
BIN_DIR="$TMP_DIR/bin"
BIN="$BIN_DIR/patchxnote"
EVIDENCE="$TMP_DIR/evidence.json"
NPM_DRY_RUN="$TMP_DIR/npm-dry-run.txt"
NPM_PACK_JSON="$TMP_DIR/npm-pack.json"
NPM_PACK_LIST="$TMP_DIR/npm-pack-list.txt"
NPM_MCP_CONFIG="$TMP_DIR/npm-mcp-config.json"
NPM_MCP_STDOUT="$TMP_DIR/npm-mcp-stdout.txt"
NPM_MCP_STDERR="$TMP_DIR/npm-mcp-stderr.txt"
NPM_PACKED_MCP_CONFIG="$TMP_DIR/npm-packed-mcp-config.json"
NPM_PACKED_MCP_STDOUT="$TMP_DIR/npm-packed-mcp-stdout.txt"
NPM_PACKED_MCP_STDERR="$TMP_DIR/npm-packed-mcp-stderr.txt"
NPM_WRAPPER="$ROOT/packages/npm/bin/patchxnote-agent.js"
PACKAGE_VERSION="$(sed -n 's/.*"version": "\([^"]*\)".*/\1/p' "$ROOT/packages/npm/package.json" | head -1)"

mkdir -p "$BIN_DIR"

cd "$ROOT"

go build -o "$BIN" ./cmd/patchxnote

NODE_BIN="${PATCHXNOTE_E2E_NODE:-}"
if [ -z "$NODE_BIN" ] && command -v node >/dev/null 2>&1; then
  NODE_BIN="$(command -v node)"
fi
if [ -z "$NODE_BIN" ] && command -v node.exe >/dev/null 2>&1; then
  NODE_BIN="$(command -v node.exe)"
fi

NPM_BIN="${PATCHXNOTE_E2E_NPM:-}"
if [ -z "$NPM_BIN" ] && command -v npm >/dev/null 2>&1; then
  NPM_BIN="$(command -v npm)"
fi
if [ -z "$NPM_BIN" ] && command -v npm.exe >/dev/null 2>&1; then
  NPM_BIN="$(command -v npm.exe)"
fi

node_is_windows() {
  case "$(basename "$NODE_BIN" | tr '[:upper:]' '[:lower:]')" in
    node.exe) return 0 ;;
    *) return 1 ;;
  esac
}

node_path() {
  if node_is_windows; then
    wslpath -w "$1"
  else
    printf "%s" "$1"
  fi
}

pack_npm_wrapper() {
  local pack_root
  pack_root="$TMP_DIR/npm-pack"
  rm -rf "$pack_root"
  mkdir -p "$pack_root"

  if node_is_windows; then
    local source_win
    local work_root_win
    local work_root_unix
    local pack_script
    local package_tgz_win
    source_win="$(wslpath -w "$ROOT/packages/npm")"
    work_root_win="$(powershell.exe -NoProfile -Command '[Console]::OutputEncoding=[Text.Encoding]::UTF8; [IO.Path]::Combine([IO.Path]::GetTempPath(), "patchxnote-agent-pack-" + [guid]::NewGuid().ToString())' | tr -d '\r')"
    work_root_unix="$(wslpath -u "$work_root_win")"
    mkdir -p "$work_root_unix"
    pack_script="$work_root_unix/pack-npm.ps1"
    cat > "$pack_script" <<'PS1'
param(
  [string]$SourceDir,
  [string]$WorkRoot
)

$ErrorActionPreference = "Stop"
New-Item -ItemType Directory -Force -Path $WorkRoot | Out-Null
$pkgSrc = Join-Path $WorkRoot "package-src"
New-Item -ItemType Directory -Force -Path $pkgSrc | Out-Null
Copy-Item -LiteralPath (Join-Path $SourceDir "package.json") -Destination $pkgSrc
Copy-Item -LiteralPath (Join-Path $SourceDir "README.md") -Destination $pkgSrc
New-Item -ItemType Directory -Force -Path (Join-Path $pkgSrc "bin") | Out-Null
Copy-Item -LiteralPath (Join-Path $SourceDir "bin\patchxnote-agent.js") -Destination (Join-Path $pkgSrc "bin")
Set-Location $pkgSrc
$json = npm pack --json
if ($LASTEXITCODE -ne 0) {
  exit $LASTEXITCODE
}
$json | Set-Content -LiteralPath (Join-Path $WorkRoot "npm-pack.json") -Encoding UTF8
$items = $json | ConvertFrom-Json
[Console]::Out.WriteLine((Join-Path $pkgSrc @($items)[0].filename))
PS1
    package_tgz_win="$(powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$(wslpath -w "$pack_script")" "$source_win" "$work_root_win" | tr -d '\r' | tail -1)"
    cp "$work_root_unix/npm-pack.json" "$NPM_PACK_JSON"
    wslpath -u "$package_tgz_win"
  else
    (
      cd "$ROOT/packages/npm"
      "$NPM_BIN" pack --json --pack-destination "$pack_root"
    ) > "$NPM_PACK_JSON"
    ls "$pack_root"/patchxnote-agent-*.tgz | head -1
  fi
}

build_windows_fake_patchxnote() {
  local win_temp
  local fake_dir
  local go_exe
  win_temp="$(powershell.exe -NoProfile -Command '[Console]::OutputEncoding=[Text.Encoding]::UTF8; [IO.Path]::GetTempPath()' | tr -d '\r')"
  fake_dir="$(wslpath -u "$win_temp")/patchxnote-agent-e2e-$$"
  mkdir -p "$fake_dir"
  cat > "$fake_dir/fake-patchxnote.go" <<GO
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) >= 2 && args[0] == "version" && args[1] == "--output" {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"version": "$PACKAGE_VERSION"})
		return
	}
	if len(args) >= 2 && args[0] == "mcp" && args[1] == "serve" {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			var request map[string]any
			if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
				continue
			}
			_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      request["id"],
				"result": map[string]any{
					"protocolVersion": "2025-06-18",
					"serverInfo":      map[string]any{"name": "fake-patchxnote", "version": "$PACKAGE_VERSION"},
					"capabilities":    map[string]any{"tools": map[string]any{}},
				},
			})
		}
		return
	}
	fmt.Fprintf(os.Stderr, "unexpected args: %v\n", args)
	os.Exit(2)
}
GO
  go_exe="$(command -v go.exe)"
  if [ -z "$go_exe" ]; then
    return 1
  fi
  (
    cd "$fake_dir"
    GO111MODULE=off "$go_exe" build -o patchxnote.exe fake-patchxnote.go
  )
  printf "%s" "$fake_dir/patchxnote.exe"
}

if [ -n "$NODE_BIN" ]; then
  "$NODE_BIN" "$(node_path "$NPM_WRAPPER")" install \
    --dry-run \
    --print-config \
    --platform linux \
    --arch x64 \
    --install-dir "$TMP_DIR/install" > "$NPM_DRY_RUN"
  "$NODE_BIN" "$(node_path "$NPM_WRAPPER")" mcp config > "$NPM_MCP_CONFIG"
  "$NODE_BIN" -e 'JSON.parse(require("fs").readFileSync(process.argv[1], "utf8"))' "$(node_path "$NPM_MCP_CONFIG")"
  if grep -n -E "MCP config:|access_token|refresh_token|otp|sk_|protocol_mac" "$NPM_MCP_CONFIG"; then
    exit 1
  fi

  if [ -z "$NPM_BIN" ]; then
    printf "npm unavailable; packed npm smoke skipped\n" > "$NPM_PACK_JSON"
    printf "npm unavailable; packed npm smoke skipped\n" > "$NPM_PACK_LIST"
    printf "npm unavailable; packed npm smoke skipped\n" > "$NPM_PACKED_MCP_CONFIG"
    printf "npm unavailable; packed npm smoke skipped\n" > "$NPM_PACKED_MCP_STDOUT"
    printf "" > "$NPM_PACKED_MCP_STDERR"
  else
    PACKAGE_TGZ="$(pack_npm_wrapper)"
    tar -tzf "$PACKAGE_TGZ" > "$NPM_PACK_LIST"
    grep -Fx "package/package.json" "$NPM_PACK_LIST" >/dev/null
    grep -Fx "package/README.md" "$NPM_PACK_LIST" >/dev/null
    grep -Fx "package/bin/patchxnote-agent.js" "$NPM_PACK_LIST" >/dev/null

    PACKED_EXTRACT="$TMP_DIR/npm-packed-extract"
    rm -rf "$PACKED_EXTRACT"
    mkdir -p "$PACKED_EXTRACT"
    tar -xzf "$PACKAGE_TGZ" -C "$PACKED_EXTRACT"
    PACKED_WRAPPER="$PACKED_EXTRACT/package/bin/patchxnote-agent.js"
    "$NODE_BIN" "$(node_path "$PACKED_WRAPPER")" mcp config > "$NPM_PACKED_MCP_CONFIG"
    "$NODE_BIN" -e 'JSON.parse(require("fs").readFileSync(process.argv[1], "utf8"))' "$(node_path "$NPM_PACKED_MCP_CONFIG")"
    if grep -n -E "MCP config:|access_token|refresh_token|otp|sk_|protocol_mac" "$NPM_PACKED_MCP_CONFIG"; then
      exit 1
    fi
  fi

  MCP_INIT='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}'
  if node_is_windows; then
    NPM_FAKE_BINARY="$(build_windows_fake_patchxnote)"
    NPM_FAKE_INSTALL="$(dirname "$NPM_FAKE_BINARY")/install dir"
    printf '%s\n' "$MCP_INIT" | "$NODE_BIN" "$(node_path "$NPM_WRAPPER")" mcp serve \
      --from-local "$(node_path "$NPM_FAKE_BINARY")" \
      --install-dir "$(node_path "$NPM_FAKE_INSTALL")" > "$NPM_MCP_STDOUT" 2> "$NPM_MCP_STDERR"
  else
    printf '%s\n' "$MCP_INIT" | "$NODE_BIN" "$(node_path "$NPM_WRAPPER")" mcp serve \
      --from-local "$BIN" \
      --install-dir "$TMP_DIR/npm-install" > "$NPM_MCP_STDOUT" 2> "$NPM_MCP_STDERR"
  fi
  "$NODE_BIN" -e 'const text=require("fs").readFileSync(process.argv[1],"utf8").trim(); const lines=text ? text.split(/\r?\n/) : []; if (lines.length !== 1) throw new Error(text); JSON.parse(lines[0]);' "$(node_path "$NPM_MCP_STDOUT")"
  if grep -n -E "Installed|PatchXNote Agent|binary missing|reinstalling" "$NPM_MCP_STDOUT"; then
    exit 1
  fi

  if [ -n "${PACKED_WRAPPER:-}" ]; then
    if node_is_windows; then
      printf '%s\n' "$MCP_INIT" | "$NODE_BIN" "$(node_path "$PACKED_WRAPPER")" mcp serve \
        --from-local "$(node_path "$NPM_FAKE_BINARY")" \
        --install-dir "$(node_path "$(dirname "$NPM_FAKE_BINARY")/packed install dir")" > "$NPM_PACKED_MCP_STDOUT" 2> "$NPM_PACKED_MCP_STDERR"
    else
      printf '%s\n' "$MCP_INIT" | "$NODE_BIN" "$(node_path "$PACKED_WRAPPER")" mcp serve \
        --from-local "$BIN" \
        --install-dir "$TMP_DIR/npm-packed-install" > "$NPM_PACKED_MCP_STDOUT" 2> "$NPM_PACKED_MCP_STDERR"
    fi
    "$NODE_BIN" -e 'const text=require("fs").readFileSync(process.argv[1],"utf8").trim(); const lines=text ? text.split(/\r?\n/) : []; if (lines.length !== 1) throw new Error(text); JSON.parse(lines[0]);' "$(node_path "$NPM_PACKED_MCP_STDOUT")"
    if grep -n -E "Installed|PatchXNote Agent|binary missing|reinstalling" "$NPM_PACKED_MCP_STDOUT"; then
      exit 1
    fi
  fi
else
  printf "node unavailable; npm wrapper dry-run skipped\n" > "$NPM_DRY_RUN"
  printf "node unavailable; npm pack skipped\n" > "$NPM_PACK_JSON"
  printf "node unavailable; npm pack skipped\n" > "$NPM_PACK_LIST"
  printf "node unavailable; npm wrapper mcp config skipped\n" > "$NPM_MCP_CONFIG"
  printf "node unavailable; npm wrapper mcp serve skipped\n" > "$NPM_MCP_STDOUT"
  printf "" > "$NPM_MCP_STDERR"
  printf "node unavailable; packed npm wrapper mcp config skipped\n" > "$NPM_PACKED_MCP_CONFIG"
  printf "node unavailable; packed npm wrapper mcp serve skipped\n" > "$NPM_PACKED_MCP_STDOUT"
  printf "" > "$NPM_PACKED_MCP_STDERR"
fi

PATCHXNOTE_E2E_BINARY="$BIN" \
PATCHXNOTE_E2E_ARTIFACT="$EVIDENCE" \
  go test -count=1 ./test/e2e -run TestMVP

if grep -n -E "000000|access_token|refresh_token|protocol_mac|sk_|raw_audio|transcript|prompt|response_payload" "$NPM_DRY_RUN" "$NPM_PACK_JSON" "$NPM_PACK_LIST" "$NPM_MCP_CONFIG" "$NPM_MCP_STDOUT" "$NPM_MCP_STDERR" "$NPM_PACKED_MCP_CONFIG" "$NPM_PACKED_MCP_STDOUT" "$NPM_PACKED_MCP_STDERR" "$EVIDENCE" >/tmp/patchxnote-agent-e2e-scan.txt 2>/dev/null; then
  cat /tmp/patchxnote-agent-e2e-scan.txt
  exit 1
fi

printf "MVP smoke PASS\nEvidence: %s\n" "$EVIDENCE"
