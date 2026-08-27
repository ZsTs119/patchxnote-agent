# Context7-Style Setup And Client Platform MCP Implementation Plan

**Goal:** Build a product-style PatchXNote MCP onboarding experience that closes both local desktop-agent setup and platform-agent access for the common domestic and global AI clients as of 2026-08-27.

**Architecture:** Keep the existing npm universal stdio entrypoint as the local runtime base. Add a Context7-style setup wizard that reuses `patchxnote mcp login` browser OAuth, stores MCP connector credentials in OS-native secure storage, installs PatchXNote into selected local MCP clients, and verifies the connection. In parallel, design and implement a remote MCP gateway for platform agents that cannot run local `npx` commands.

**Tech Stack:** Go, Cobra, npm launcher, OS-native keychain storage, local stdio MCP, hosted `/mcp` JSON-RPC for platform mode, PatchXNote GoServer Agent APIs, OAuth authorization code + PKCE, JSON/TOML config merge helpers, website client registry data.

**Execution Rule:** Work sequentially in the primary agent. Do not use sub-agents or parallel task execution. Keep changes small, testable, and reversible.

## Implementation Status 2026-08-27

- [x] Task 0 baseline rechecked: npm latest is `0.2.7`, local installed binary reports `0.2.7`, `mcp config` prints pure JSON, and `mcp serve` stdout remains JSON-RPC only.
- [x] Task 1 client registry added under `docs/mcp-clients/clients.json` with schema validation script and Chinese registry notes.
- [x] Task 2 setup command contract implemented: `patchxnote setup`, npm wrapper delegation, `--client`, `--all-local-supported`, `--dry-run`, `--no-browser`, `--print-config`, `--force`, `--yes`, and `--output json`.
- [x] Task 3 has been superseded by MCP browser OAuth: `patchxnote mcp login/status/logout` now owns the browser login path, and `setup --client <id>` reuses it instead of promoting setup-session fallback.
- [x] Task 4 config merge adapters implemented for strict JSON and TOML with backup, conflict detection, rollback, and JSONC manual-mode protection. Path helpers are consolidated in `internal/setup/paths_by_os.go` because Go treats `*_windows.go`, `*_darwin.go`, and `*_linux.go` as build-constrained files.
- [x] Task 5 P0 local client setup wired: VS Code, Cursor, Codex, Claude Desktop, and Windsurf can auto-write after confirmation; Claude Code, Trae, Qoder, and WorkBuddy return manual instructions/deeplinks/config snippets in V1.
- [x] Task 6 local setup verification hooks added: config verification and direct MCP initialize/tools-list smoke are implemented; setup also verifies existing auth when credentials are present.
- [x] Task 7 interim website specs and client detail copy added under `docs/mcp-clients/`.
- [x] Task 8 remote MCP platform gateway design added under `docs/plans/2026-08-27-remote-mcp-platform-gateway-design.md`.
- [x] Task 9 platform PoC status captured under `docs/evidence/2026-08-27-platform-client-poc-status.zh-CN.md`; real Feishu/Tencent/enterprise WorkBuddy acceptance remains blocked by remote gateway deployment and platform test tenant credentials.
- [x] Task 10 README, Chinese README, npm README, runbook, and package version source state updated for the setup release line.
- [x] Task 11 final validation gates passed for this implementation pass: `go test ./...`, npm wrapper tests, client registry validation, MVP smoke, setup dry-runs on WSL and Windows, local MCP initialize/tools-list smoke, `git diff --check`, and sensitive-value scan review.

---

## Current Baseline

- [x] `patchxnote-agent@0.2.7` already provides the universal local MCP entrypoint:
  - `npx -y patchxnote-agent@latest mcp config`
  - `npx -y patchxnote-agent@latest mcp serve`
  - `npx -y patchxnote-agent@latest login`
- [x] Current local MCP transport is stdio.
- [x] MCP stdout is reserved for JSON-RPC; install diagnostics and startup repair logs must remain on stderr.
- [x] MCP config must not contain phone numbers, OTP codes, access tokens, refresh tokens, webhook secrets, full MAC, SK, raw audio, full transcripts, prompts, or provider payloads.
- [x] Existing local Agent data access is server-authorized and currently exposes 19 MCP tools.
- [x] Existing local login stores credentials through the secure-storage boundary.
- [ ] New setup work must preserve the existing universal `mcp config` and `mcp serve` behavior.
- [ ] New platform work must not weaken the local Agent security boundary.

## Audit Additions From Plan Review

- [ ] Treat local setup and platform remote MCP as separate release slices even if they share the same website.
- [ ] Do not claim one-click support for a client until the actual install link, config write, or marketplace flow is verified on that client.
- [ ] Add explicit cross-OS handling for Windows, WSL, VS Code Remote, macOS, Linux desktop, and headless terminals.
- [ ] Ensure credentials are stored in the same OS/runtime environment that will later run `patchxnote mcp serve`.
- [ ] Add a user confirmation step before setup modifies any editor or agent config file.
- [ ] Add a rollback command or documented recovery path for every auto-written local client config.
- [x] Re-scope remote MCP to 19-tool functional parity with local stdio; control risk through OAuth consent, explicit scopes, response caps, revocation, and safe logging.
- [ ] Add a marketplace/discovery lane separate from connectivity: website cards, VS Code/Cursor/Qoder deeplinks, Codex plugin, WorkBuddy/Feishu/Tencent platform submission, and generic MCP registries.
- [ ] Add logo/trademark usage checks before publishing the website client grid.
- [ ] Add client-support evidence states: `researched`, `implemented`, `locally_smoked`, `published_smoked`, `platform_accepted`.

## Release Slices

Avoid turning the first implementation into one large all-or-nothing release.

- [ ] Slice A: client registry and website card/detail pages.
- [ ] Slice B: local setup wizard for P0 local clients, using existing npm stdio MCP.
- [ ] Slice C: browser OAuth login and secure-storage handoff.
- [ ] Slice D: remote MCP gateway PoC for one domestic platform client.
- [ ] Slice E: broader platform acceptance and marketplace/discovery submissions.

Suggested public release order:

```text
0.2.8: website/client registry and manual client detail pages
0.2.9: local setup wizard for VS Code, Cursor, Codex, Claude Code, and WorkBuddy manual path
0.3.0: browser OAuth login and config merge hardening
0.3.1+: remote MCP platform PoC for Feishu Aily / Doubao Work Partner / Tencent platform
```

The exact versions can change, but each release should have its own acceptance evidence.

## Product Scope

This plan has two first-version closed loops.

### Local Closed Loop

User flow:

```text
User opens PatchXNote MCP website
 -> selects VS Code / Cursor / Codex / Claude Code / WorkBuddy / another local client
 -> copies or launches one setup command
 -> `npx -y patchxnote-agent@latest setup --client <client>` runs
 -> setup opens PatchXNote MCP browser OAuth login
 -> user logs in and approves Agent access
 -> setup stores MCP connector credentials in OS-native secure storage
 -> setup installs or updates the selected client's MCP config
 -> setup verifies PatchXNote MCP by listing tools and reading a safe account/record projection
 -> user can ask the selected client to read PatchXNote summaries
```

Local setup command examples:

```sh
npx -y patchxnote-agent@latest setup --client vscode
npx -y patchxnote-agent@latest setup --client cursor
npx -y patchxnote-agent@latest setup --client codex
npx -y patchxnote-agent@latest setup --client claude-code
npx -y patchxnote-agent@latest setup --client workbuddy
```

Local MCP config remains secret-free:

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

### Platform Closed Loop

User flow:

```text
User opens PatchXNote MCP website
 -> selects Feishu Aily / Doubao Work Partner / Tencent Agent Development Platform / WorkBuddy enterprise mode
 -> website shows PatchXNote remote MCP option
 -> user creates or authorizes a platform connector session
 -> platform connects to the GoServer-hosted `/mcp` endpoint
 -> platform can initialize, list tools, and call the same 19 PatchXNote MCP tools as local MCP
 -> platform verifies by reading recent PatchXNote summary records for the authorized user
```

Platform remote MCP target:

```text
https://ws-lab.patch-x.cn/patchnote-test-api/mcp
```

Remote platform V1 was re-scoped on 2026-08-27 to target functional parity with all current 19 local MCP tools. See `docs/plans/2026-08-27-remote-mcp-goserver-parity-checklist.md`.

Remote mode still needs environment-aware adaptations: webhook targets move from local config/keychain into GoServer account-scoped storage, and render/export tools return bounded content or a download handle instead of writing to a user's local filesystem.

## Client Priority Matrix

This is not a global MAU ranking. It is the first-version target list for common MCP-capable AI editors, terminal agents, desktop agents, and domestic office-agent platforms as of 2026-08-27.

| Priority | Client | Region/Type | First-Version Install Path | Notes |
| --- | --- | --- | --- | --- |
| P0 | VS Code / GitHub Copilot | Global editor | `setup --client vscode`; also support VS Code MCP install link or `code --add-mcp` where available | Must support user-level config and workspace config without overwriting other servers. |
| P0 | Cursor | Global AI editor | `setup --client cursor`; also support Cursor deeplink | Use `npx ... mcp serve` by default; absolute binary fallback if startup is slow. |
| P0 | Codex / ChatGPT Desktop / Codex IDE | Global CLI/desktop/IDE | `setup --client codex`; also print `codex mcp add patchxnote -- npx -y patchxnote-agent@latest mcp serve` | Web one-click should wait until a Codex plugin/marketplace path is real. |
| P0 | Claude Code | Global CLI agent | `setup --client claude-code`; also print `claude mcp add ...` command | Keep command-based install first. |
| P0 | Claude Desktop | Global desktop agent | `setup --client claude-desktop`; config-file merge first | `.mcpb` Desktop Extension is a later polish path. |
| P0 | Windsurf | Global AI editor | `setup --client windsurf`; config-file merge | Verify current config path and schema before coding. |
| P0 | Trae / Trae CN / TraeWork Code | Domestic/global AI editor | `setup --client trae`; config-file merge or documented paste flow | Treat Trae variants separately if their config paths diverge. |
| P0 | Qoder | Global/domestic AI coding agent | `setup --client qoder`; also support Qoder deeplink if stable | Deeplink should be optional; config merge remains fallback. |
| P0 | WorkBuddy / Tencent CodeBuddy WorkBuddy | Domestic office desktop agent | `setup --client workbuddy`; MCP + CLI local connector path | Must be shown as its own website card, not hidden under domestic platforms. |
| P0.5 | Feishu Aily / Doubao Work Partner | Domestic office-agent platform | Remote MCP gateway | Local `npx` is not enough for cloud/platform mode; verify Aily's current custom MCP flow. |
| P0.5 | Tencent Agent Development Platform / Enterprise WorkBuddy | Domestic agent platform | Remote MCP gateway | Platform docs require online MCP URL and initialize/tools/list/tools/call support. |
| P1 | JetBrains AI Assistant | Global IDE family | Config-template first; auto-write after validation | Support IntelliJ/WebStorm/PyCharm path differences later. |
| P1 | Zed | Global editor | `context_servers` config template | Extension or deeper install support can follow. |
| P1 | Gemini CLI | Global terminal agent | Config-template or `setup --client gemini-cli` | Confirm exact config path before implementing auto-write. |
| P1 | Qwen Code | Domestic/global terminal agent | Config-template or `setup --client qwen-code` | Verify official MCP schema before auto-write. |
| P1 | Kimi Code / Kimi CLI | Domestic/global terminal agent | Config-template or `setup --client kimi-code` | Supports stdio/HTTP/SSE; choose stdio for local mode. |
| P1 | OpenCode | Global terminal/desktop agent | Config-template or `setup --client opencode` | Useful for developer audience; not required for first website launch. |
| P1 | Cline / Continue / Roo-derived VS Code agents | Global VS Code extensions | VS Code-compatible template cards | Avoid overfitting each extension until VS Code base path is solid. |
| P2 watchlist | Aider, Augment Code, Sourcegraph Cody, Tabnine, Amazon Q Developer, Replit Agent, Bolt/Lovable-style builders | Global developer tools | Website watchlist only until MCP support and install paths are verified | Useful discovery targets, but do not spend first-version engineering on clients without a clear MCP path. |
| P2 watchlist | Baidu Comate, Alibaba Lingma, MarsCode, and other domestic AI IDE/office assistants | Domestic developer/office tools | Website watchlist only until public MCP/custom-tool path is verified | Track because of domestic distribution value; promote only after docs or manual acceptance confirm MCP support. |

## Website Scope

Website should feel like a product onboarding surface, not a documentation index.

- [ ] Create a visual client grid with cards for all P0/P0.5 clients and optional P1 cards.
- [ ] Each card should show client logo, client category, region/global signal, support status, and primary install action.
- [ ] Card click opens a client detail page.
- [ ] Detail page should include:
  - [ ] primary install command or deeplink
  - [ ] copyable MCP config
  - [ ] login/setup explanation
  - [ ] verification prompt
  - [ ] fallback path
  - [ ] known client-specific caveats
- [ ] Keep examples free of personal data and secrets.
- [ ] Put all client metadata into one registry file so website, docs, and CLI setup stay aligned.

Suggested client registry shape:

```json
{
  "id": "cursor",
  "name": "Cursor",
  "priority": "P0",
  "category": "ai-editor",
  "regions": ["global"],
  "local_stdio": true,
  "remote_mcp": true,
  "setup_command": "npx -y patchxnote-agent@latest setup --client cursor",
  "mcp_command": "npx",
  "mcp_args": ["-y", "patchxnote-agent@latest", "mcp", "serve"],
  "one_click": {
    "supported": true,
    "type": "deeplink"
  },
  "status": "planned",
  "evidence": {
    "researched_at": "2026-08-27",
    "config_smoke": false,
    "published_smoke": false,
    "platform_acceptance": false
  }
}
```

## Implementation Checklist

### Task 0: Confirm Baseline And Preserve User Changes

Files:

- Read: `README.md`
- Read: `docs/engineering-rules.md`
- Read: `docs/release-and-maintenance-runbook.zh-CN.md`
- Read: `docs/plans/2026-08-06-agent-v1-mvp.md`
- Read: `docs/plans/2026-08-26-universal-mcp-onboarding-checklist.md`
- Read: `../patchxNoteGoServer/docs/integrations/apifox/shared/integration-guide.zh-CN.md`

Checklist:

- [ ] Run `git status --short` before any implementation.
- [ ] Preserve unrelated existing changes in `packages/npm/bin/patchxnote-agent.js` and `scripts/e2e/mvp-smoke.sh`.
- [ ] Confirm current npm latest and local installed binary version before changing setup behavior.
- [ ] Confirm `mcp config` still prints pure JSON.
- [ ] Confirm `mcp serve` still keeps stdout JSON-RPC only.
- [ ] Confirm current local real-account smoke can list recent memories without printing raw content.
- [ ] Confirm whether implementation is running on Windows, WSL, macOS, Linux desktop, or Linux headless.
- [ ] Confirm which runtime will launch each selected MCP client.
- [ ] Confirm whether Node.js/npm, `npx`, GitHub Release downloads, and OS keychain are available in that same runtime.
- [ ] Confirm no implementation relies on unverified client config paths or stale web docs.

Validation:

```sh
git status --short
npx -y patchxnote-agent@latest mcp config
```

Expected:

- Existing user changes are recorded and left intact.
- `mcp config` output is directly parseable JSON.

### Task 1: Add Client Registry Contract

Files:

- Create: `docs/mcp-clients/clients.json`
- Create: `docs/mcp-clients/README.zh-CN.md`
- Modify later if website lives in this repo: website data import path
- Modify later if CLI consumes the same file: generated Go fixture or embedded JSON

Checklist:

- [ ] Define client IDs for all P0/P0.5/P1 clients.
- [ ] Record each client's supported transports: stdio, Streamable HTTP, SSE.
- [ ] Record install strategies: config merge, CLI command, deeplink, marketplace, remote URL.
- [ ] Record config format: JSON, JSONC, TOML, UI-only, platform form.
- [ ] Record default support status: `supported`, `manual`, `planned`, or `research`.
- [ ] Record whether first-version setup should auto-write config.
- [ ] Record official reference URL for each client.
- [ ] Record primary website card group: `global-local`, `domestic-local`, `domestic-platform`, or `watchlist`.
- [ ] Record whether the client runs MCP servers on the host OS, inside WSL/remote containers, or inside a cloud platform.
- [ ] Record whether config can be auto-written safely, must use a CLI command, must use a deeplink, or must remain manual.
- [ ] Record whether the client requires app restart, server refresh, or workspace reload after config changes.
- [ ] Record whether the client supports environment variables, headers, OAuth, API keys, or no auth in MCP config.
- [ ] Record whether website buttons are `copy`, `deeplink`, `open-settings`, `marketplace`, or `remote-url`.
- [ ] Add schema validation for required fields.
- [ ] Keep registry free of tokens, real account data, and user-specific paths.

Validation:

```sh
node -e "JSON.parse(require('fs').readFileSync('docs/mcp-clients/clients.json','utf8')); console.log('clients json ok')"
git diff --check
```

Expected:

- Registry parses as JSON.
- No trailing whitespace.

### Task 2: Define Setup Command Contract

Files:

- Modify: `internal/cli/root.go`
- Create: `internal/cli/setup.go`
- Create: `internal/setup/`
- Modify: `packages/npm/bin/patchxnote-agent.js`
- Modify: `packages/npm/test/install.test.js`

Checklist:

- [ ] Add `patchxnote setup`.
- [ ] Add npm wrapper command `patchxnote-agent setup` that ensures the binary and delegates to `patchxnote setup`.
- [ ] Add `--client <id>`.
- [ ] Add `--all-local-supported` only if it can safely install multiple clients sequentially.
- [ ] Add `--dry-run`.
- [ ] Add `--no-browser`.
- [ ] Add `--print-config`.
- [ ] Add `--force` only for explicit overwrite/repair behavior.
- [ ] Add `--output json` for scripted diagnostics.
- [ ] Add `--profile <name>` passthrough only if it maps to existing non-secret config behavior.
- [ ] Add `--server-base-url` only for development/test users; do not show it in normal public website commands.
- [ ] Ask for confirmation before modifying any client config, unless `--yes` is explicitly passed.
- [ ] Print exact manual fallback instructions when auto-write is not supported or not safe.
- [ ] Detect missing Node.js/npm and recommend the existing native binary fallback instead of failing opaquely.
- [ ] Detect enterprise proxy/download failures and recommend one-time install or pinned binary fallback.
- [ ] Never print phone, OTP, access token, refresh token, or raw content.
- [ ] Keep setup human diagnostics on stderr when stdout is structured.

Validation:

```sh
go test ./internal/cli ./internal/setup
node packages/npm/test/install.test.js
```

Expected:

- Unknown client IDs fail with a stable error.
- Dry-run prints planned actions without modifying any client config.
- npm wrapper delegates setup without changing existing install/mcp/login behavior.

### Task 3: Add Browser OAuth Login

Files:

- Agent repo:
  - Modify: `internal/api/`
  - Add: `internal/oauthflow/`
  - Modify: `internal/cli/setup.go`
  - Add tests under `internal/oauthflow`, `internal/api`, `internal/cli`
- GoServer repo:
  - OAuth authorize/token/revoke and hosted `/mcp` endpoints are now the active contract.

Checklist:

- [ ] CLI discovers `/.well-known/oauth-authorization-server`.
- [ ] CLI creates PKCE verifier/challenge and state in process memory only.
- [ ] CLI opens PatchXNote web authorization URL.
- [ ] Browser login redirects to a one-time `127.0.0.1:<port>/callback`.
- [ ] CLI exchanges the authorization code with `/v1/agent/oauth/token`.
- [ ] CLI stores received MCP connector credentials only in OS-native secure storage.
- [ ] CLI stores credentials under the selected profile and server environment, so test/prod sessions cannot be mixed silently.
- [ ] CLI clears callback listener, state, verifier, and authorization code after success or timeout.
- [ ] `--no-browser` prints the authorization URL for remote/headless terminals and waits for callback.
- [ ] OAuth authorization codes are short-lived and single-use.
- [ ] Authorization page displays the client name, requested scopes, server environment, and expiry time before approval.
- [ ] OAuth state mismatch fails closed.
- [ ] OAuth flow does not create or replace App/PC `mobile` or `desktop` installations.
- [ ] Existing phone OTP login remains available only as legacy local Agent fallback, not setup's default path.

Validation:

```sh
go test ./internal/auth ./internal/api ./internal/cli
```

Expected:

- Browser login success stores credentials.
- Timeout and denied authorization are handled cleanly.
- Local config still contains no secrets.

### Task 4: Implement Config Merge Adapters For Local Clients

Files:

- Create: `internal/setup/clients.go`
- Create: `internal/setup/config_json.go`
- Create: `internal/setup/config_jsonc.go`
- Create: `internal/setup/config_toml.go`
- Create: `internal/setup/paths_windows.go`
- Create: `internal/setup/paths_darwin.go`
- Create: `internal/setup/paths_linux.go`
- Add tests: `internal/setup/*_test.go`

Checklist:

- [ ] Implement a common adapter interface: detect, read, plan, write, verify.
- [ ] Always create a timestamped backup before modifying an existing client config.
- [ ] Merge only the `patchxnote` MCP server entry.
- [ ] Preserve all other user MCP servers and client settings.
- [ ] Preserve formatting where practical; otherwise document that the specific file is normalized.
- [ ] Preserve JSONC comments where practical; if not possible, keep manual mode for that client in V1.
- [ ] Support dry-run diff summary.
- [ ] Support rollback when write verification fails.
- [ ] Use absolute path fallback when a client rejects `npx` or starts from a WSL UNC cwd.
- [ ] Handle missing config directories by creating only the minimum needed path.
- [ ] Refuse to write if the target path resolves outside the expected user config area unless the user passes an explicit path.
- [ ] Detect symlinks, permission errors, locked files, readonly files, and client-owned config rewrites.
- [ ] If the client is currently running and may overwrite config on exit, warn the user or use the client's official CLI/deeplink path instead.
- [ ] Avoid storing remote MCP headers or API keys in local client config unless a separate remote-auth design allows it.
- [ ] Keep config writes atomic with temp-file and rename where the platform supports it.

Validation:

```sh
go test ./internal/setup
```

Expected:

- Tests cover missing file, existing file, invalid JSON/TOML, existing `patchxnote` entry, multiple unrelated servers, Windows paths with spaces, and WSL UNC fallback.

### Task 5: Implement P0 Local Client Installers

Files:

- Modify: `internal/setup/clients.go`
- Add per-client tests under `internal/setup/`
- Modify docs generated from `docs/mcp-clients/clients.json`

Checklist:

- [ ] VS Code adapter:
  - [ ] support copyable config
  - [ ] support install link or command path where available
  - [ ] support user-level config first
- [ ] Cursor adapter:
  - [ ] support config merge
  - [ ] support deeplink as optional website action
- [ ] Codex adapter:
  - [ ] support `codex mcp add` command generation
  - [ ] support TOML config merge only after current schema verification
- [ ] Claude Code adapter:
  - [ ] support `claude mcp add` command generation
  - [ ] avoid assuming Claude Code is installed
- [ ] Claude Desktop adapter:
  - [ ] support JSON config merge
  - [ ] require app restart guidance after config changes
- [ ] Windsurf adapter:
  - [ ] verify current config path
  - [ ] support JSON config merge or manual snippet
- [ ] Trae adapter:
  - [ ] verify global/CN config path
  - [ ] support JSON config merge or manual snippet
- [ ] Qoder adapter:
  - [ ] support deeplink if current official scheme is stable
  - [ ] support config fallback
- [ ] WorkBuddy adapter:
  - [ ] support MCP + CLI local connector instructions
  - [ ] verify whether desktop config can be written safely or should be manual in V1
- [ ] Do not include P1/P2 clients in automatic first-version setup unless their current official MCP path is verified.
- [ ] For VS Code-derived extensions, prefer the VS Code base install path first, then add extension-specific docs only after real acceptance.

Validation:

```sh
go test ./internal/setup
scripts/e2e/mvp-smoke.sh
```

Expected:

- Each P0 adapter has a dry-run test and a write/merge fixture test where config files are known.
- Manual-only clients clearly return `manual_required` with copyable instructions.

### Task 6: Add Local Setup Verification

Files:

- Create: `internal/setup/verify.go`
- Modify: `internal/cli/setup.go`
- Add tests under `internal/setup`

Checklist:

- [ ] After config install, verify the selected client config points to PatchXNote MCP.
- [ ] Verify the local Agent is authenticated.
- [ ] Start a short MCP protocol smoke directly against `patchxnote mcp serve` or the npm command without launching the editor.
- [ ] Call `initialize`.
- [ ] Call `tools/list`.
- [ ] Optionally call `patchxnote_get_current_user`.
- [ ] Optionally call `patchxnote_list_memories` with a user-selected platform.
- [ ] Treat empty memory lists as a valid authenticated state, not setup failure.
- [ ] Verify both `mobile` and `desktop` platform options when the account has data and the user chooses to test both.
- [ ] Keep verification output bounded and sanitized.
- [ ] Do not export source text or provider payload during setup verification.

Validation:

```sh
go test ./internal/setup ./internal/mcp
```

Expected:

- Verification distinguishes unauthenticated, client-config-written, MCP-start-failed, and server-data-empty states.

### Task 7: Website Client Cards And Detail Pages

Files:

- Create or modify in the website repo once confirmed.
- If website source is not yet chosen, create interim specs:
  - `docs/mcp-clients/website-page-spec.zh-CN.md`
  - `docs/mcp-clients/client-detail-copy.zh-CN.md`

Checklist:

- [ ] Build a dark product-style grid page with common AI client cards.
- [ ] Keep first screen focused on choosing a client, not marketing copy.
- [ ] Include P0 and P0.5 cards by default.
- [ ] Put P1 cards in an expandable or secondary section.
- [ ] Detail pages show one primary action per client.
- [ ] Support copy buttons for setup commands and config snippets.
- [ ] Support one-click/deeplink buttons only where the client officially supports them.
- [ ] Include fallback commands on every detail page.
- [ ] Include a verification prompt for every client.
- [ ] Do not put access tokens or user-specific values in website snippets.
- [ ] Add Chinese and English copy for P0 client pages.
- [ ] Add a visible status label: `One-click`, `Setup command`, `Manual`, `Remote MCP`, or `Coming soon`.
- [ ] Add logo/trademark attribution and confirm usage rights before public deployment.
- [ ] Add privacy-safe analytics only if product needs funnel data; do not record commands containing user paths or account identifiers.
- [ ] Add noindex or staging protection before the site is ready for public promotion.

Validation:

```sh
git diff --check
```

Expected:

- Website spec and copy have no secrets or user data.

### Task 8: Design Remote MCP Gateway For Platform Clients

Files:

- Create: `docs/plans/2026-08-27-remote-mcp-platform-gateway-design.md` if this grows beyond this checklist.
- GoServer repo:
  - Create design doc under `../patchxNoteGoServer/docs/engineering/` if implementing in GoServer.
  - Modify OpenAPI/server docs only after contract is accepted.
- Agent repo:
  - Update docs and website registry after the endpoint shape is finalized.

Checklist:

- [x] Decide deployment owner: GoServer integrated route first.
- [ ] Expose HTTPS endpoint under the GoServer base URL, currently `https://ws-lab.patch-x.cn/patchnote-test-api/mcp` for beta.
- [ ] Support `initialize`.
- [ ] Support `tools/list`.
- [ ] Support `tools/call`.
- [ ] Support Streamable HTTP first.
- [ ] Support SSE only if a P0 platform requires it.
- [x] Keep remote platform V1 functionally equivalent to the local 19-tool MCP surface.
- [ ] Require per-user OAuth authorization and a revocable connector token.
- [ ] Provide an admin/user revocation path.
- [x] Decide whether remote auth is OAuth, device-code connector token, signed platform token, or a combination: official OAuth flow that issues connector tokens.
- [ ] Store remote connector credentials server-side only; never ask the user to paste PatchXNote access tokens into a platform.
- [ ] Add connector/session listing so users can see and revoke which platforms are connected.
- [ ] Bind every remote call to account, platform client, scopes, and audit request ID.
- [ ] Add CORS/origin policy only where browser-based clients require it; do not use wildcard credentials.
- [ ] Add rate limits per account, connector, IP/platform, and tool.
- [ ] Add prompt-injection and data-exfiltration guardrails to tool descriptions and outputs.
- [ ] Add bounded pagination and output caps.
- [ ] Add stable MCP error mapping: unauthenticated, forbidden, not_found, rate_limited, upstream_unavailable.
- [ ] Expose webhook send/configure through GoServer account-scoped storage, not local config files.
- [ ] Expose model IO field tools with explicit scopes, bounded response size, and no logging/evidence of raw payload bodies.
- [ ] Expose only the existing Agent source-text/safe transcript projection; do not add raw audio download or broader transcript access.
- [ ] Do not cache full record content in the remote gateway unless a separate retention design is accepted.
- [ ] Log only safe diagnostics: version, request ID, status, stable error code, platform client name.

Validation:

```sh
curl -fsS https://ws-lab.patch-x.cn/patchnote-test-api/mcp/health
```

Expected:

- Health endpoint returns non-sensitive service status after deployment.
- MCP inspector or equivalent can initialize, list tools, and call safe read-only tools.

### Task 9: Platform Client PoC Acceptance

Files:

- Create: `docs/evidence/` sanitized notes only if evidence needs to be checked in.
- Update: `docs/mcp-clients/clients.json`
- Update: website detail pages/spec.

Checklist:

- [ ] Feishu Aily / Doubao Work Partner:
  - [ ] verify current custom MCP creation flow
  - [ ] add PatchXNote remote MCP URL
  - [ ] configure auth according to accepted method
  - [ ] run initialize/tools/list/tools/call
  - [ ] verify recent summaries can be read
- [ ] Tencent Agent Development Platform:
  - [ ] add custom MCP tool with online URL
  - [ ] verify Streamable HTTP or SSE mode
  - [ ] verify tools appear in intelligent workbench/workflow/Multi-Agent where supported
- [ ] WorkBuddy desktop:
  - [ ] verify local MCP + CLI path
  - [ ] verify whether custom connector can run local `npx`
  - [ ] record whether website should show local setup or remote setup as primary
- [ ] Record unsupported states explicitly:
  - [ ] platform can add MCP but cannot authenticate
  - [ ] platform can list tools but cannot call tools
  - [ ] platform can call tools but strips required arguments
  - [ ] platform requires remote URL only
  - [ ] platform requires enterprise approval before public use

Validation:

Expected:

- Each platform has a short sanitized acceptance note:
  - client name
  - connection method
  - transport
  - tool count
  - one successful safe read call
  - no raw phone/token/content recorded

### Task 10: Documentation And Release Update

Files:

- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `packages/npm/README.md`
- Modify: `docs/release-and-maintenance-runbook.zh-CN.md`
- Modify: `docs/plans/2026-08-06-agent-v1-mvp.md` only for release-record updates
- Modify: `docs/plans/2026-08-26-universal-mcp-onboarding-checklist.md` only if superseding status needs a note

Checklist:

- [ ] Document `setup --client`.
- [ ] Document browser-login setup flow.
- [ ] Document that MCP config remains secret-free.
- [ ] Document local vs platform mode.
- [ ] Document P0 supported clients.
- [ ] Document manual fallback for every client.
- [ ] Document WSL UNC caveat and absolute-path fallback.
- [ ] Document cross-OS credential caveat: run setup in the same environment that runs the MCP server.
- [ ] Document that both local MCP and remote platform MCP target the same 19-tool functional surface, with remote file/config behavior adapted to server-side state or bounded returned content.
- [ ] Document how to undo a client install or restore the backup config.
- [ ] Update npm README with setup commands.
- [ ] Update runbook release gates for setup wizard and remote MCP.
- [ ] Keep docs free of real phone numbers, OTPs, tokens, raw source text, provider payloads, full MAC, SK, and real webhook URLs.

Validation:

```sh
git diff --check
```

Expected:

- Markdown diff has no whitespace errors.

### Task 11: Validation Gates

Files:

- No source files expected unless a failing validation requires a fix.

Checklist:

- [ ] Run Go tests after code changes.
- [ ] Run npm wrapper tests after npm setup changes.
- [ ] Run existing MVP smoke.
- [ ] Add setup wizard smoke.
- [ ] Add client config merge fixture tests.
- [ ] Add remote MCP protocol smoke if platform gateway is implemented.
- [ ] Add WSL-vs-Windows setup detection tests.
- [ ] Add config backup and rollback tests.
- [ ] Add JSONC preservation or manual-mode tests.
- [ ] Add browser OAuth timeout/deny/state-mismatch/retry tests.
- [ ] Add platform remote auth failure and revocation tests if platform gateway is implemented.
- [ ] Run docs diff check.
- [ ] Run sensitive-value scan.
- [ ] After release, verify from npm registry, not only local source.
- [ ] Validate at least one Windows local client install path.
- [ ] Validate at least one macOS install path in CI or a real host.
- [ ] Validate at least one Linux install path.

Commands:

```sh
go test ./...
node packages/npm/test/install.test.js
scripts/e2e/mvp-smoke.sh
git diff --check
grep -RInE "access_token|refresh_token|Bearer |otp|sk_|protocol_mac|NPM_TOKEN|NODE_AUTH_TOKEN" README.md README.zh-CN.md packages/npm docs scripts internal --exclude-dir=.git --exclude-dir=.tmp --exclude-dir=dist
```

Expected:

- Existing 19-tool local MCP smoke remains green.
- Setup wizard tests pass.
- Remote MCP smoke passes if platform gateway is in this release slice.
- Sensitive-value scan has no real secrets or user data; harmless field-name matches are reviewed.

## Acceptance Criteria

### Local Acceptance

- [ ] Website shows P0 client cards for VS Code, Cursor, Codex, Claude Code, Claude Desktop, Windsurf, Trae, Qoder, and WorkBuddy.
- [ ] `npx -y patchxnote-agent@latest setup --client <p0-client>` either installs automatically or returns clear manual instructions.
- [ ] Browser login stores MCP connector credentials in OS-native secure storage.
- [ ] MCP config contains no credentials.
- [ ] Existing local `mcp config`, `mcp serve`, and `login` commands still work.
- [ ] Setup verifies `initialize` and `tools/list`.
- [ ] Setup can verify a safe read for an authenticated account.
- [ ] Existing `install --print-config` fallback remains available.
- [ ] Auto-write creates a backup and rollback path.
- [ ] Auto-write never deletes or rewrites unrelated MCP servers.
- [ ] Setup warns when the chosen client/runtime does not match the OS where credentials were stored.

### Platform Acceptance

- [ ] Remote MCP endpoint is reachable over HTTPS.
- [ ] Remote MCP supports initialize/tools/list/tools/call.
- [ ] Feishu Aily or Doubao Work Partner can add PatchXNote MCP and call at least one safe read tool.
- [ ] Tencent platform or WorkBuddy enterprise path can add PatchXNote MCP and call at least one safe read tool.
- [ ] Platform V1 uses the same curated 19-tool surface where platform permissions allow; model/source/provider field tools require explicit scopes, bounded output, and safe logs.
- [ ] Platform V1 does not expose raw audio or broader transcript/audio download surfaces.
- [ ] Platform auth can be revoked.
- [ ] Platform connector list/revoke is visible to the user or operator.
- [ ] Logs and evidence contain no tokens, raw phone numbers, OTPs, raw audio, full transcripts, source text dumps, prompts, provider payloads, full MAC, SK, or webhook secrets.

## Known Risks And Edge Cases

- Some clients support MCP but do not expose a stable one-click install link.
- Some clients require different config keys, such as `mcpServers`, `servers`, `context_servers`, or `mcp`.
- Some clients store JSONC rather than strict JSON.
- Some clients require restart after config changes.
- Some clients reject `npx` during MCP startup; absolute binary fallback must remain supported.
- Windows editors launched from WSL UNC paths can trigger command-shell warnings; setup should prefer normal Windows paths or absolute binary fallback.
- WSL can store credentials in the Linux keychain while a Windows editor later launches a Windows binary with Windows Credential Manager. Setup must detect this and guide the user to run setup in the same runtime.
- VS Code Remote/Dev Container/Codespaces-like sessions may run MCP servers remotely rather than on the desktop host. Local setup must not assume the desktop OS is the runtime.
- Browser login must work from remote terminals and WSL, so `--no-browser` and a copyable authorization URL are required.
- Browser login can fail because default-browser launch is blocked; setup must still work with a copyable authorization URL and callback timeout guidance.
- Corporate proxies, npm registry blocks, GitHub Release download blocks, antivirus quarantine, and PowerShell execution policy can break first-run install; setup needs actionable fallback messages.
- Remote platform clients may not be able to run local commands, so local setup does not satisfy Feishu Aily/Doubao/Tencent platform mode.
- Platform OAuth support differs by product; first platform PoC may require a revocable connector token or custom header.
- Platform clients may cache tool lists and descriptions; remote rollout must tolerate stale tool schemas during propagation.
- Remote MCP has a broader trust boundary than local stdio; keep the 19-tool parity surface but enforce explicit scopes, response caps, revocation, and safe logs.
- Tool descriptions must be concise; too many tools can reduce model selection accuracy in platform agents.
- Website one-click buttons must not claim support before the actual client flow has been validated.
- Third-party logos and product names require trademark-safe presentation.
- Public website snippets can become stale quickly; client registry needs a reviewed date and confidence status.

## Reference Baseline To Recheck During Implementation

- Model Context Protocol overview: `https://modelcontextprotocol.io/docs/2026-07-28/getting-started/intro`
- MCP local servers: `https://modelcontextprotocol.io/docs/2026-07-28/develop/connect-local-servers`
- MCP remote servers: `https://modelcontextprotocol.io/docs/2026-07-28/develop/connect-remote-servers`
- VS Code MCP servers: `https://code.visualstudio.com/docs/agent-customization/mcp-servers`
- VS Code MCP extension/install guidance: `https://code.visualstudio.com/api/extension-guides/ai/mcp`
- Cursor MCP install links: `https://cursor.com/docs/mcp/install-links`
- Claude Code MCP: `https://code.claude.com/docs/en/mcp`
- Codex MCP: `https://learn.chatgpt.com/docs/extend/mcp`
- Windsurf MCP: `https://docs.windsurf.com/zh/windsurf/cascade/mcp`
- Trae MCP: `https://docs.trae.ai/ide/add-mcp-servers`
- Qoder deeplinks/MCP: `https://docs.qoder.com/user-guide/deeplink`
- Zed MCP: `https://zed.dev/docs/ai/mcp`
- JetBrains AI Assistant MCP: `https://www.jetbrains.com/help/ai-assistant/mcp.html`
- Gemini CLI MCP: `https://github.com/google-gemini/gemini-cli/blob/main/docs/tools/mcp-server.md`
- Qwen Code MCP: `https://qwenlm.github.io/qwen-code-docs/en/developers/tools/mcp-server/`
- Kimi Code MCP: `https://www.kimi.com/code/docs/en/kimi-code-cli/customization/mcp.html`
- OpenCode tools/MCP: `https://opencode.ai/docs/tools/`
- WorkBuddy connector: `https://www.workbuddy.cn/docs/workbuddy/From-Beginner-to-Expert-Guide/Function-Description/Connector`
- WorkBuddy MCP guide: `https://www.codebuddy.cn/docs/workbuddy/From-Beginner-to-Expert-Guide/Function-Description/MCP-Guide`
- Tencent Agent Development Platform custom MCP: `https://cloud.tencent.com/document/product/1759/117855`
- Feishu Aily MCP product overview: `https://www.feishu.cn/content/article/7576921890476788922`
- Feishu Open Platform MCP overview: `https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/mcp_integration/mcp_introduction?lang=zh-CN`

## Non-Blocking Decisions Before Implementation

- [ ] Confirm website source repository or hosting target.
- [x] Confirm whether platform remote MCP lives inside GoServer or a separate gateway service: GoServer-integrated route first.
- [x] Confirm whether remote platform V1 uses OAuth, connector token, or both: official OAuth authorization flow issuing revocable connector tokens.
- [x] Confirm whether remote model-IO field tools are excluded from V1 or hidden behind a separate explicit authorization: include them for 19-tool parity, bounded by explicit OAuth scopes and result-size caps.
- [ ] Confirm whether WorkBuddy desktop setup can be auto-written safely or should be manual in V1.
- [x] Confirm whether browser login requires GoServer changes, website backend changes, or both: GoServer OAuth authorize/token/revoke is the active server contract.
- [x] Confirm whether remote MCP should share existing Agent sessions or use separate remote connector sessions: use separate MCP connector sessions.
- [ ] Confirm whether local setup should default to `npx` config or package-pinned absolute binary config for each client.
- [ ] Confirm client logo usage policy before public website deployment.
- [ ] Confirm first release version after `0.2.7`.
