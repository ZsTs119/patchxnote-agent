#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TMP_DIR="$ROOT/.tmp/e2e"
BIN_DIR="$TMP_DIR/bin"
BIN="$BIN_DIR/patchxnote"
EVIDENCE="$TMP_DIR/evidence.json"
NPM_DRY_RUN="$TMP_DIR/npm-dry-run.txt"

mkdir -p "$BIN_DIR"

cd "$ROOT"

go build -o "$BIN" ./cmd/patchxnote

NODE_BIN="${PATCHXNOTE_E2E_NODE:-}"
if [ -z "$NODE_BIN" ] && command -v node >/dev/null 2>&1; then
  NODE_BIN="$(command -v node)"
fi

if [ -n "$NODE_BIN" ]; then
  "$NODE_BIN" "$ROOT/packages/npm/bin/patchxnote-agent.js" install \
    --dry-run \
    --print-config \
    --platform linux \
    --arch x64 \
    --install-dir "$TMP_DIR/install" > "$NPM_DRY_RUN"
else
  printf "node unavailable; npm wrapper dry-run skipped\n" > "$NPM_DRY_RUN"
fi

PATCHXNOTE_E2E_BINARY="$BIN" \
PATCHXNOTE_E2E_ARTIFACT="$EVIDENCE" \
  go test -count=1 ./test/e2e -run TestMVP

if grep -n -E "000000|access_token|refresh_token|protocol_mac|sk_|raw_audio|transcript|prompt|response_payload" "$NPM_DRY_RUN" "$EVIDENCE" >/tmp/patchxnote-agent-e2e-scan.txt 2>/dev/null; then
  cat /tmp/patchxnote-agent-e2e-scan.txt
  exit 1
fi

printf "MVP smoke PASS\nEvidence: %s\n" "$EVIDENCE"
