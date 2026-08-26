# Universal MCP Onboarding Implementation Plan

**Goal:** Make PatchXNote Agent connectable from any local-command-capable MCP host through one generic npm command, with terminal-based login and no platform-specific plugin required for V1.

**Architecture:** Keep the existing `patchxnote` Go binary as the only product runtime and MCP server. Extend the npm wrapper into a thin launcher that can install or verify the pinned binary, then exec `patchxnote mcp serve` or `patchxnote login`; the wrapper must never print non-JSON text to stdout while serving MCP. Client-specific marketplaces and VS Code extensions remain later discovery layers, not V1 dependencies.

**Tech Stack:** Go, Cobra, local stdio MCP, npm installer wrapper, Node.js `child_process`, GitHub Release assets, OS-native keychain storage.

**Execution Rule:** Work sequentially in the primary agent. Do not use sub-agents or parallel task execution. Keep the implementation small, testable, and easy to revert.

---

## Confirmed Decisions

- [x] V1 supports the simple command shape: `mcp serve`, `login`, and `mcp config`.
- [x] V1 documentation can start with generic MCP setup only; no client-specific plugin is required.
- [x] Public copy-paste MCP config defaults to `patchxnote-agent@latest`.
- [x] `mcp config` prints pure JSON only, so users can paste it directly into editor MCP settings.
- [x] No `baseURL` is needed in the first public flow.
- [x] Do not put phone number, OTP, access token, refresh token, webhook secret, or server credentials in MCP config.
- [x] Login remains a terminal CLI flow inside the user's editor terminal.
- [x] Do not add new server endpoints or new MCP tools for this feature.
- [x] Do not auto-edit WorkBuddy, Trae, Qoder, VS Code, Cursor, Claude Desktop, or Codex config files in V1.

## V1 User Flow

```text
User opens any MCP-capable desktop agent/editor
 -> adds the generic PatchXNote MCP JSON
 -> MCP host starts `npx -y patchxnote-agent@latest mcp serve`
 -> npm wrapper ensures the local `patchxnote` binary exists
 -> wrapper execs `patchxnote mcp serve`
 -> if not logged in, tools return auth_required
 -> user runs `npx -y patchxnote-agent@latest login` in the editor terminal
 -> login stores credentials in OS-native keychain
 -> user restarts or refreshes the MCP server in the editor
 -> PatchXNote MCP tools work
```

Generic MCP config:

```json
{
  "mcpServers": {
    "patchxnote": {
      "command": "npx",
      "args": ["-y", "patchxnote-agent@latest", "mcp", "serve"]
    }
  }
}
```

Terminal login command:

```sh
npx -y patchxnote-agent@latest login
```

Config generation command:

```sh
npx -y patchxnote-agent@latest mcp config
```

## Non-Goals

- [ ] Do not publish a VS Code extension in this implementation.
- [ ] Do not publish to Open VSX in this implementation.
- [ ] Do not submit to WorkBuddy, Trae, Qoder, Cursor, Claude, or Codex marketplaces in this implementation.
- [ ] Do not build a remote HTTP/SSE MCP service in this implementation; V1 is local stdio only.
- [ ] Do not add OAuth/device-code browser login.
- [ ] Do not add phone number or OTP flags to MCP config.
- [ ] Do not add background login, background sync, or background webhook sending.
- [ ] Do not change GoServer Agent auth/read contracts.
- [ ] Do not change the current 19-tool MCP surface.

## Expected Files

- Modify: `packages/npm/bin/patchxnote-agent.js`
- Modify: `packages/npm/test/install.test.js`
- Modify as needed: `scripts/e2e/mvp-smoke.sh`
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `packages/npm/README.md`
- Modify as needed: `docs/release-and-maintenance-runbook.zh-CN.md`
- Modify before Agent work if still failing: `../patchxNoteGoServer/tests/smoke/agent-model-io-read/agent_model_io_read_test.go`
- Create or update release notes only if the implementation ships in a new public version.

## Key Product Rules

- [ ] `npx -y patchxnote-agent@latest mcp serve` is the main cross-client entrypoint.
- [ ] `mcp serve` may install or repair the local binary before serving.
- [ ] `mcp serve` must write installer/repair diagnostics only to stderr.
- [ ] Once MCP serving starts, stdout is reserved for JSON-RPC only.
- [ ] `login` may use normal terminal stdin/stdout because it is not an MCP protocol process.
- [ ] `mcp config` prints pure copy-paste JSON to stdout, with no label or explanatory prefix.
- [ ] The generated JSON contains no `env` block by default.
- [ ] The generated JSON contains no base URL by default.
- [ ] The generated JSON contains no secret or personal data.
- [ ] Existing `install --print-config` remains supported as the stable absolute-path fallback.
- [ ] Existing `install`, `update`, and `uninstall` behavior remains backward-compatible.
- [ ] The npm package version and downloaded Go binary version are treated as a pair.
- [ ] Startup version probes must never leak their stdout into the MCP stdout stream.
- [ ] Process spawning uses argument arrays, not shell-concatenated command strings, so Windows paths with spaces and POSIX paths both work.
- [ ] Install or repair never leaves a partial binary at the final path; keep the previous working binary until a replacement is downloaded and checksum-verified.
- [ ] If the matching binary is already installed, `mcp serve` works without network access.

## Implementation Checklist

### Task 0: Repair GoServer Agent Model IO Smoke Fixture

Files:

- Modify if still failing: `../patchxNoteGoServer/tests/smoke/agent-model-io-read/agent_model_io_read_test.go`
- Read if needed: `../patchxNoteGoServer/internal/modelgateway/repository.go`
- Read if needed: `../patchxNoteGoServer/openapi/openapi.yaml`

Why this is in this plan:

- The universal MCP flow depends on the existing Agent-only read APIs staying green.
- The current Agent client code still matches the online `/v1/agent/**` contract, but the GoServer `agent-model-io-read` smoke fixture can fail if its handcrafted `model_request` rows omit newly required model gateway fields.
- This is a service-side regression gate repair, not a product scope expansion.

Checklist:

- [ ] Re-run `MODULE=agent-model-io-read make smoke-module` in `../patchxNoteGoServer` to confirm whether the fixture is still failing.
- [ ] If it still fails on `model_request.final_output_schema_version`, update each handcrafted `INSERT INTO model_request` in `tests/smoke/agent-model-io-read/agent_model_io_read_test.go` to include `final_output_schema_version`.
- [ ] Use the same final schema version value that the seeded task/run expects from the current model gateway contract.
- [ ] Do not change Agent API paths, response envelopes, auth scopes, or MCP tool schemas as part of this fixture repair.
- [ ] Do not store phone numbers, OTPs, access tokens, refresh tokens, full source text, provider payloads, full MAC, SK, or webhook URLs in tracked evidence.

Validation:

```sh
cd ../patchxNoteGoServer
MODULE=agent-model-io-read make test-module
MODULE=agent-model-io-read make smoke-module
```

Expected:

- `module_tests=PASS module=agent-model-io-read`
- `module_smoke=PASS module=agent-model-io-read`
- Agent-only interfaces remain the same 14 `/v1/agent/**` paths in OpenAPI.

### Task 1: Confirm Current Local State

Files:

- Read: `packages/npm/bin/patchxnote-agent.js`
- Read: `packages/npm/test/install.test.js`
- Read: `internal/cli/mcp.go`
- Read: `internal/cli/login.go`
- Read: `scripts/e2e/mvp-smoke.sh`

Checklist:

- [ ] Run `git status --short` and record unrelated local changes before editing.
- [ ] Confirm whether `packages/npm/bin/patchxnote-agent.js` already has user changes and preserve them.
- [ ] Confirm current npm wrapper exports helper functions used by tests.
- [ ] Confirm `patchxnote mcp serve` already uses stdout only for MCP JSON-RPC.
- [ ] Confirm `patchxnote login` already works as the terminal login path.

Validation:

```sh
git status --short
```

Expected:

- Existing unrelated changes, if any, are recorded and left untouched.

### Task 2: Add npm Command Parsing For Launcher Commands

Files:

- Modify: `packages/npm/bin/patchxnote-agent.js`
- Modify: `packages/npm/test/install.test.js`

Checklist:

- [ ] Add accepted top-level commands: `install`, `update`, `uninstall`, `login`, and `mcp`.
- [ ] Add `mcp` subcommands: `serve` and `config`.
- [ ] Keep unknown commands failing with non-zero exit and a clear usage message.
- [ ] Preserve existing install/update/uninstall options.
- [ ] Allow `--install-dir`, `--platform`, `--arch`, and `--from-local` where tests need them.
- [ ] Do not allow `--print-config` to change MCP protocol serving behavior.

Suggested usage text:

```text
usage: patchxnote-agent <install|update|uninstall|login|mcp> [options]
usage: patchxnote-agent mcp <serve|config> [options]
```

Validation:

```sh
node packages/npm/test/install.test.js
```

Expected:

- Existing installer tests still pass after parser changes.
- New parser tests cover `login`, `mcp serve`, `mcp config`, and unknown subcommands.

### Task 3: Extract Quiet Binary Ensure Helper

Files:

- Modify: `packages/npm/bin/patchxnote-agent.js`
- Modify: `packages/npm/test/install.test.js`

Checklist:

- [ ] Extract current install planning into a reusable helper, for example `createInstallPlan(options)`.
- [ ] Extract current download/checksum/write path into a reusable helper, for example `installBinary(plan, options)`.
- [ ] Add a quiet mode so `mcp serve` can install without writing to stdout.
- [ ] In quiet mode, send any install diagnostics to stderr.
- [ ] Keep normal `install` command stdout unchanged unless tests are intentionally updated.
- [ ] Preserve checksum verification for every downloaded binary.
- [ ] Preserve user-writable install location.
- [ ] Preserve executable permission checks on macOS/Linux.
- [ ] Make install/repair robust against interrupted or concurrent starts:
  - [ ] download to a temporary path before touching the final binary
  - [ ] verify checksum before moving the replacement into place
  - [ ] avoid deleting a working existing binary until the replacement is ready
  - [ ] use a small install lock if the helper can do so simply
  - [ ] clean stale temp/lock files without blocking future startup

Validation:

```sh
node packages/npm/test/install.test.js
```

Expected:

- Normal dry-run output still includes install plan and absolute-path MCP config when requested.
- Quiet helper tests prove no install text goes to stdout.
- Sensitive-value scan still passes.
- Fixture tests cover replacing a stale binary without corrupting an existing good binary.

### Task 4: Implement `patchxnote-agent mcp serve`

Files:

- Modify: `packages/npm/bin/patchxnote-agent.js`
- Modify: `packages/npm/test/install.test.js`
- Modify as needed: `scripts/e2e/mvp-smoke.sh`

Checklist:

- [ ] Resolve the current package platform, arch, version, install dir, and binary path.
- [ ] If the binary is missing, install the package-pinned binary quietly.
- [ ] If the binary exists but cannot run `version --output json`, reinstall quietly.
- [ ] If the binary version differs from the npm package version, reinstall quietly.
- [ ] Spawn the installed binary with args `["mcp", "serve"]`.
- [ ] Connect wrapper stdin/stdout/stderr directly to the child process after preflight.
- [ ] Capture any preflight command stdout/stderr before MCP starts and forward only safe diagnostics to stderr.
- [ ] Forward the child process exit code.
- [ ] Forward termination signals where practical.
- [ ] Exit cleanly when the MCP client closes stdin, without leaving an orphan child process.
- [ ] Use `spawn` with argument arrays; do not depend on shell quoting.
- [ ] Do not print success banners, config JSON, install paths, or PATH hints to stdout in this command.

Validation:

```sh
node packages/npm/test/install.test.js
```

Expected:

- Unit tests cover missing binary, existing matching binary, stale binary, failing binary preflight, stdin close, and paths with spaces using fixtures.

Protocol smoke:

```sh
printf '%s\n' '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}' | npx -y ./packages/npm mcp serve
```

Expected:

- stdout contains exactly one JSON-RPC response line.
- stderr may contain bounded diagnostics only if installation or repair happened.
- stdout does not contain installer text.

### Task 5: Implement `patchxnote-agent login`

Files:

- Modify: `packages/npm/bin/patchxnote-agent.js`
- Modify: `packages/npm/test/install.test.js`

Checklist:

- [ ] Resolve and ensure the local binary using the same helper as `mcp serve`.
- [ ] Spawn the installed binary with args `["login"]` plus any explicit user flags passed after `login`.
- [ ] Use inherited stdio so the user can enter phone number and OTP in the editor terminal.
- [ ] Keep phone number and OTP out of generated config and logs.
- [ ] Forward the child process exit code.

Validation:

```sh
node packages/npm/test/install.test.js
```

Expected:

- Fixture tests prove `login` delegates to the installed binary with expected args.
- Test output does not contain OTP-like or token-like values.

Manual smoke after implementation:

```sh
npx -y ./packages/npm login
```

Expected:

- The terminal shows the existing PatchXNote login prompts from the Go CLI.

### Task 6: Implement `patchxnote-agent mcp config`

Files:

- Modify: `packages/npm/bin/patchxnote-agent.js`
- Modify: `packages/npm/test/install.test.js`

Checklist:

- [ ] Print the generic `npx`-based MCP JSON to stdout as pure JSON.
- [ ] Use `patchxnote-agent@latest` by default in the generated args.
- [ ] Do not include `baseUrl`, `server-base-url`, `env`, profile, phone, OTP, or token fields.
- [ ] Keep existing `install --print-config` output as the absolute-path fallback.
- [ ] If a future pinned output flag is desired, leave it out unless needed for tests or release docs.
- [ ] Add a test that `JSON.parse(stdout)` succeeds without trimming labels or explanatory text.

Generated output:

```json
{
  "mcpServers": {
    "patchxnote": {
      "command": "npx",
      "args": ["-y", "patchxnote-agent@latest", "mcp", "serve"]
    }
  }
}
```

Validation:

```sh
node packages/npm/test/install.test.js
```

Expected:

- `mcp config` stdout parses as JSON directly.
- Output contains `patchxnote-agent@latest`.
- Output contains no secrets or personal data.

### Task 7: Update E2E Smoke For Universal Entry

Files:

- Modify: `scripts/e2e/mvp-smoke.sh`

Checklist:

- [ ] Keep the existing installed-binary MCP smoke.
- [ ] Add one smoke for npm wrapper `mcp config`.
- [ ] Add one smoke for npm wrapper `mcp serve` using a local package path and local binary fixture/build output.
- [ ] Add a packaged npm smoke so behavior is validated from the package boundary, not only from source files:
  - [ ] `npm pack --dry-run` includes the launcher and required installer files
  - [ ] a local packed package can print `mcp config`
  - [ ] a local packed package can delegate `mcp serve` to a test binary without polluting stdout
- [ ] Force the smoke to run sequentially.
- [ ] Assert the wrapper does not write installer text to MCP stdout.
- [ ] Keep evidence sanitized.
- [ ] Do not record phone numbers, OTP, access tokens, refresh tokens, raw source text, provider payloads, or webhook URLs.

Validation:

```sh
scripts/e2e/mvp-smoke.sh
```

Expected:

- Existing 19-tool MCP smoke still passes.
- New universal npm launcher smoke passes.

### Task 8: Update Public Documentation

Files:

- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `packages/npm/README.md`
- Modify as needed: `docs/release-and-maintenance-runbook.zh-CN.md`

Checklist:

- [ ] Add the generic MCP config as the primary manual client setup snippet.
- [ ] Keep `npx -y patchxnote-agent@latest install --print-config` documented as the stable fallback.
- [ ] Add `npx -y patchxnote-agent@latest login` as the editor-terminal login command.
- [ ] Explain that MCP config contains no credentials.
- [ ] Explain that auth failures mean the user should run login in the editor terminal and restart/refresh the MCP server.
- [ ] Explain cold-start behavior: the first MCP start may download the platform binary.
- [ ] If an editor times out on cold start, document the workaround:
  - [ ] run `npx -y patchxnote-agent@latest install --print-config` once
  - [ ] paste the generated absolute-path fallback config
- [ ] State that V1 does not require VS Code Marketplace, WorkBuddy Marketplace, Trae Marketplace, Qoder Marketplace, or Open VSX.
- [ ] State that the client must support local stdio MCP commands.
- [ ] Note that some clients may require wrapper-specific fields such as `type: "stdio"` or a different top-level key, while the `command` and `args` stay the same.
- [ ] Keep public examples free of phone numbers, OTPs, tokens, full MAC, SK, real webhook URLs, source text, and provider payloads.

Validation:

```sh
git diff --check
```

Expected:

- No trailing whitespace or Markdown formatting errors from the diff check.

### Task 9: Run Local Validation Gates

Files:

- No source files expected unless a failing test exposes a necessary fix.

Checklist:

- [ ] Run Go tests because the e2e path still depends on the Go binary and MCP protocol.
- [ ] Run npm wrapper tests because this feature changes the npm wrapper.
- [ ] Run npm packed-package smoke from a temporary directory, especially on Windows where install paths can contain spaces.
- [ ] Run e2e smoke because this feature changes the public MCP startup path.
- [ ] Run `git diff --check`.
- [ ] Run a sensitive-value scan over changed files.
- [ ] Run a sanitized real-account read smoke when credentials are available:
  - [ ] `auth status` reports authenticated without printing raw phone/token
  - [ ] MCP `tools/list` works
  - [ ] memory list works for the authorized platform
  - [ ] model-io list works and reports only counts/statuses
  - [ ] source/provider/packaged export probes write to temp files only and are deleted immediately

Commands:

```sh
go test ./...
node packages/npm/test/install.test.js
(cd packages/npm && npm pack --dry-run)
scripts/e2e/mvp-smoke.sh
git diff --check
grep -RInE "access_token|refresh_token|Bearer |otp|sk_|protocol_mac|NPM_TOKEN|NODE_AUTH_TOKEN" README.md README.zh-CN.md packages/npm docs scripts internal --exclude-dir=.git --exclude-dir=.tmp --exclude-dir=dist
```

Expected:

- All runtime tests pass.
- Packaged npm boundary contains the launcher files needed for `npx`.
- Sensitive-value scan has no real secrets or user data; harmless field-name matches are reviewed.

### Task 10: Release Preparation

Files:

- Modify only if releasing now: `packages/npm/package.json`
- Modify only if release docs need a version pin update: `README.md`
- Modify only if release docs need a version pin update: `README.zh-CN.md`
- Modify only if release docs need a version pin update: `packages/npm/README.md`
- Modify only if recording release state: `docs/plans/2026-08-06-agent-v1-mvp.md`

Checklist:

- [ ] Decide target version, likely the next patch after `0.2.6`.
- [ ] Do not reuse `0.2.6` for this implementation: npm registry already has `patchxnote-agent@0.2.6` as `latest` as of 2026-08-26.
- [ ] Before public release, bump `packages/npm/package.json` and the Git tag/GitHub Release binary version together to the same unpublished version, likely `0.2.7`.
- [ ] Confirm release assets will be built from the exact tagged commit.
- [ ] Confirm npm Trusted Publishing remains OIDC-based and does not use `NPM_TOKEN` or `NODE_AUTH_TOKEN`.
- [ ] Confirm GitHub Release produces all six OS/arch binaries and `checksums.txt`.
- [ ] Confirm npm wrapper installs the package-pinned binary version.
- [ ] Run Windows, Linux, and macOS install/MCP smokes before public promotion.
- [ ] After publish, verify the published package itself:
  - [ ] `npx -y patchxnote-agent@<version> mcp config`
  - [ ] `npx -y patchxnote-agent@<version> install --print-config`
  - [ ] `npx -y patchxnote-agent@<version> mcp serve` protocol smoke from a clean install dir

Validation:

```sh
npm view patchxnote-agent version dist-tags.latest repository.url --registry https://registry.npmjs.org
```

Expected:

- npm metadata points to `ZsTs119/patchxnote-agent`.
- Published version matches the intended release after publish.
- Post-publish `@latest` smoke remains pending until a new version is tagged and published; do not treat the currently published `0.2.6` package as evidence for this local implementation.

## Acceptance Criteria

- [ ] A user can paste the generic MCP JSON into a local stdio MCP-capable editor or agent.
- [ ] The MCP host can start PatchXNote through `npx -y patchxnote-agent@latest mcp serve`.
- [ ] First-run binary install or repair does not corrupt MCP stdout.
- [ ] `mcp config` stdout is parseable JSON with no explanatory text.
- [ ] A cold start with no binary installed either completes install and starts MCP, or fails before protocol startup with stderr-only diagnostics.
- [ ] A warm start with a matching installed binary works offline.
- [ ] Closing the MCP client's stdin terminates the wrapper and child process cleanly.
- [ ] Concurrent or interrupted startup does not corrupt a previously working installed binary.
- [ ] `tools/list` still returns the expected 19 PatchXNote MCP tools.
- [ ] If unauthenticated, MCP tools return the existing `auth_required` behavior.
- [ ] The user can run `npx -y patchxnote-agent@latest login` in the editor terminal.
- [ ] Login stores credentials in OS-native keychain as before.
- [ ] After MCP restart/refresh, authenticated tools can call PatchXNote as before.
- [ ] Existing absolute-path install flow remains available.
- [ ] No MCP config or docs example contains secrets or personal data.

## Known Risks And Later Follow-Ups

- First MCP startup through `npx` may be slower than an absolute binary path because npm and GitHub Release downloads can be involved; docs should include the one-time install fallback.
- Some enterprise editors may block custom local commands even if they support MCP generally.
- Some MCP clients may use a different JSON shape from `mcpServers`; V1 documentation will start generic, then add client-specific snippets after real acceptance.
- If `npx`, a client wrapper, a binary version probe, or installer code writes unexpected text to stdout, recommend the existing absolute-path `install --print-config` fallback.
- Concurrent editor launches can race on first install or repair; installer logic should keep the previous working binary until replacement is verified.
- Users without Node.js still need a separate installer path later.
- Exposing `auth status` through the npm wrapper would improve terminal UX, but keep it as a later follow-up unless V1 scope is explicitly expanded.
- Marketplaces are a discovery problem, not a V1 connectivity blocker; evaluate VS Code Marketplace, Open VSX, WorkBuddy, Trae, and Qoder after this generic flow is stable.

## Open Questions Before Coding

- Is the npm wrapper allowed to auto-repair a stale installed binary during `mcp serve`, or should it only install when missing and let `install/update` handle upgrades?
- Should `login` pass through extra flags such as `--profile`, or should V1 keep `npx ... login` flagless in public docs while still forwarding advanced args internally?
