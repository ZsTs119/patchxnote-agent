# PatchXNote Agent Engineering Rules

This document keeps the PatchXNote Agent CLI and local MCP bridge from drifting away from the product and security model.

## Architecture

PatchXNote Agent has three layers:

1. `patchxnote` Go binary
   The local runtime for login, account inspection, MCP stdio serving, local cache management, and installer integration.

2. npm installer wrapper
   A thin package that supports commands such as `npx -y patchxnote-agent install`. It installs or updates the versioned Go binary, then exits.

3. PatchXNote server APIs
   The remote source of truth. The agent never bypasses server authorization and never reconstructs server facts from local guesses.

The expected runtime shape is:

```text
Agent client -> local MCP stdio -> patchxnote binary -> PatchXNote API
```

## Repository Layout

The repository uses a thin public binary entrypoint plus internal packages. Keep this structure stable unless a design note explains why it must change.

```text
patchxnote-agent/
  cmd/
    patchxnote/
      main.go

  internal/
    cli/
    config/
    auth/
    keychain/
    api/
    mcp/
    cache/
    webhook/
    renderdoc/
    modelio/
    localfile/
    output/
    diag/
    version/

  packages/
    npm/
      package.json
      bin/
        install.js

  docs/
    engineering-rules.md

  testdata/
```

Package responsibilities:

- `cmd/patchxnote`: minimal `package main`; call the CLI package and own only process exit.
- `internal/cli`: Cobra command tree, flags, command wiring, prompt boundaries, and command tests.
- `internal/config`: Viper-backed non-secret config loading, default values, env binding, and config path resolution.
- `internal/auth`: login session state, token refresh orchestration, scope checks, and logout behavior.
- `internal/keychain`: OS secure-storage adapter boundary. Keep platform differences behind this package.
- `internal/api`: PatchXNote server client, request/response mapping, retry policy, and stable API error mapping.
- `internal/mcp`: MCP stdio server, tool registry, tool schemas, and JSON-RPC error mapping.
- `internal/cache`: local cache and search index for authorized read-only projections.
- `internal/webhook`: local webhook target metadata, keychain-backed secret lookup, provider payload building, and send result normalization.
- `internal/renderdoc`: delivery-document to Markdown rendering, built-in templates, and user template loading.
- `internal/modelio`: model IO field selection, availability handling, pretty JSON formatting, and explicit local field/export writing.
- `internal/localfile`: neutral atomic local file writer helpers shared by local feature packages.
- `internal/output`: table/json/plain renderers. Commands should not hand-roll output formatting.
- `internal/diag`: safe diagnostics, redaction, stderr logging, and support bundle helpers.
- `internal/version`: build-time version, commit, date, and target metadata.
- `packages/npm`: npm install/update/uninstall wrapper. Do not place product runtime logic here.

The dependency direction is:

```text
cmd/patchxnote -> internal/cli -> internal/{config,auth,api,mcp,cache,webhook,renderdoc,modelio,localfile,output,diag,version}
internal/auth -> internal/{config,keychain,api}
internal/mcp -> internal/{auth,api,cache,config,keychain,webhook,renderdoc,modelio,localfile,diag}
internal/api -> external PatchXNote server APIs
```

Avoid reverse dependencies from lower-level packages back into `internal/cli`.

## Non-Goals For V1

V1 is intentionally not a full PatchXNote client.

It does not:

- bind, recover, release, reset, or format recorder cards
- read or write SK/credential values
- read live BLE state, battery, storage, or recording status
- upload audio, raw transcripts, complete corrected transcripts, speaker profiles, or full dictionaries
- execute model runs on behalf of App/PC
- delete cloud content
- claim daily reward quota
- purchase quota or initiate payments
- expose admin APIs

These exclusions are product decisions, not missing features.

## Authentication Model

Agent access must be independent from App/PC installation access.

The current App/PC contract allows one active `mobile` installation and one active `desktop` installation per account. A local agent must not consume either slot or trigger installation replacement. The server should provide an agent-specific access model before public release.

The CLI may still use phone OTP as the human login ceremony, but the resulting credential must have agent-specific audience, scope, and revocation semantics.

Recommended scope shape:

```text
agent:account.read
agent:hardware.read
agent:quota.read
agent:model_usage.read
agent:content.read:mobile
agent:content.read:desktop
```

Grant content scopes explicitly. Do not silently read both mobile and desktop structured results.

## Server Contract Pinning

The server contract is the source of truth for account, quota, hardware, model usage, and structured results.

Before implementing or changing an API call:

1. identify the exact server OpenAPI or integration document
2. record the contract version or commit SHA used for implementation
3. add or update request/response fixtures without sensitive values
4. map server errors into stable CLI/MCP errors
5. update this repository when the server contract changes

Do not infer missing fields from App/PC behavior. If the server does not store a value, such as live recorder-card battery, the Agent cannot expose it as a server-backed MCP tool.

## Credential Storage

Use two storage classes:

1. Non-secret config
   Profile names, base URL, selected content platform, local cache settings, and display preferences.

2. Secret storage
   Refresh credentials, access tokens, proof material, and any future device-code or OAuth credentials.

Secret storage should use OS-native secure storage:

- macOS Keychain
- Windows Credential Manager
- Linux Secret Service/libsecret where available

If a platform cannot provide secure storage, fail closed for public builds unless the user explicitly enables a documented development-only fallback.

## Configuration Path Contract

Use predictable platform paths and keep secrets separate from config.

Non-secret config may include:

- profile name
- server base URL
- selected content platform
- output preference
- cache enablement and cache path
- MCP client install metadata

Secret material must never be written to config files. Store it only through `internal/keychain`.

Default config locations:

```text
macOS:   ~/Library/Application Support/PatchXNote Agent/config.yaml
Windows: %AppData%\PatchXNote Agent\config.yaml
Linux:   ${XDG_CONFIG_HOME:-~/.config}/patchxnote-agent/config.yaml
```

Default cache locations:

```text
macOS:   ~/Library/Caches/PatchXNote Agent/
Windows: %LocalAppData%\PatchXNote Agent\Cache\
Linux:   ${XDG_CACHE_HOME:-~/.cache}/patchxnote-agent/
```

Environment variables must use the `PATCHXNOTE_` prefix. For nested config, replace dots and hyphens with underscores, for example `PATCHXNOTE_SERVER_BASE_URL`.

## Command Design

The public command surface should stay small:

```text
patchxnote setup
patchxnote login
patchxnote logout
patchxnote auth status
patchxnote mcp
patchxnote mcp install <client>
patchxnote mcp status
patchxnote version
```

Use predictable CLI conventions:

- subcommands use lowercase words
- flags use kebab-case
- commands return non-zero exit codes on failure
- `--output json` is stable and scriptable
- prompt-based flows always provide a non-interactive alternative before public automation support is claimed
- no progress messages on stdout when stdout is meant to be parsed

## Command Tree Contract

Use Cobra with a fresh command tree constructor, for example `internal/cli.NewRootCommand(deps)`.

The first implementation should map commands to files like this:

```text
internal/cli/
  root.go
  setup.go
  login.go
  logout.go
  auth_status.go
  mcp.go
  mcp_install.go
  mcp_status.go
  version.go
```

Command rules:

- `cmd/patchxnote/main.go` calls `internal/cli.Execute()` and handles final process exit only.
- command handlers use `RunE`
- root commands set `SilenceUsage` and `SilenceErrors`
- shared auth/config checks live in root or group-level `PersistentPreRunE`
- child `PersistentPreRunE` must explicitly call parent setup when both are needed
- use `cmd.OutOrStdout()` for result output and `cmd.ErrOrStderr()` for diagnostics
- command tests construct a fresh command tree and set args/out/err per test
- do not call `os.Exit` outside `cmd/patchxnote/main.go`
- do not use package-level mutable Viper state in tests

## MCP Tool Design

Tool names should be service-scoped to avoid collisions:

```text
patchxnote_get_current_user
patchxnote_list_recorder_cards
patchxnote_get_quota_summary
patchxnote_get_model_usage_summary
patchxnote_list_memories
patchxnote_search_memories
patchxnote_get_memory
patchxnote_list_webhook_targets
patchxnote_configure_webhook_target
patchxnote_remove_webhook_target
patchxnote_list_webhook_templates
patchxnote_render_webhook_message
patchxnote_export_model_io
patchxnote_send_webhook
patchxnote_list_model_io_traces
patchxnote_get_model_io_source_text
patchxnote_get_model_io_provider_response
patchxnote_get_model_io_parsed_result
patchxnote_get_model_io_packaged_result
```

Each tool must define:

- exact read/write behavior
- required scope
- input schema with min/max lengths and enums
- pagination or limit defaults
- output size limit
- stable error codes
- whether results are live API projections or local cache projections

Do not use generic names such as `search`, `list`, or `get_user`.

Server-backed PatchXNote data tools should stay read-only unless a separate server contract allows otherwise. Webhook tools are the accepted V1 exception: they may write local non-secret config, write URL/signing secret material through `internal/keychain`, and perform manual external HTTP sends when explicitly invoked. They must never return full webhook URLs or signing secrets.

Model IO trace discovery returns lightweight metadata and request IDs only. Model IO field tools are explicit inspection tools for trusted local agents. They must return only the requested field, keep large content behind explicit output files, and avoid scattering field selection logic outside `internal/modelio`.

## Local Cache

Local cache is an optimization and search index, not a new source of truth.

Allowed cache data:

- account-scoped opaque IDs
- result IDs and revision IDs
- event/key item/daily digest titles and short summaries returned by authorized server APIs
- timestamps and cursors
- local full-text index derived from authorized structured results

Disallowed cache data:

- access/refresh tokens
- SK/credential
- raw audio
- complete transcript text
- raw provider payload
- full MAC unless a future design explicitly allows it

The cache must be clearable with a user command and must not make deleted or unauthorized server data appear current after logout or token revocation.

## Hardware Projection

The Agent may list recorder cards only through server-authorized read projections.

V1 should return:

- hardware ID
- binding epoch ID
- masked identity
- binding status
- credential version
- created/updated timestamps when available

Avoid returning full `protocol_mac_normalized` by default. If a future tool exposes it, require an explicit parameter and document why the full MAC is needed.

Never expose SK/credential or call hardware connect/release through MCP V1.

## Structured Content Projection

Agent content access reads only dedicated server `/v1/agent/**` projections for the selected platform. Formal structured results still follow the server's storage-consent rules. Agent record lookup may also include read-only model-generated entries from server-authorized `model_io_trace` projections when no formal structured result exists yet.

V1 may expose:

- event title and summary
- key item title, lifecycle, owner projection, due projection, and short detail
- daily digest date and narrative/summary fields
- source availability status

V1 must not imply access to original audio, complete raw transcript, complete corrected transcript, or speaker identity.
Complete source text and AI payloads are available only through explicit model IO field tools or local exports for the logged-in user.

## Testing Layout And Gates

Use tests to protect the command surface, protocol contract, and security boundaries.

Suggested layout:

```text
internal/cli/*_test.go
internal/config/*_test.go
internal/auth/*_test.go
internal/api/*_test.go
internal/mcp/*_test.go
internal/mcp/testdata/
testdata/api/
testdata/mcp/
```

Minimum gates:

- `go test ./...` for all code changes
- command tests for new commands, flags, outputs, and exit behavior
- config tests for precedence: explicit set, flag, env, config file, defaults
- keychain adapter tests using fakes; do not require real user secrets in CI
- API client tests with sanitized fixtures and stable server error mapping
- MCP protocol smoke for tool listing, schema validation, success response, auth failure, and pagination
- installer dry-run tests for npm wrapper changes

Tests and fixtures must not contain access tokens, refresh tokens, OTP values, raw phone numbers, full MAC, SK, raw audio, complete transcripts, or provider payloads.

## Installer And Release

The npm package should:

1. detect OS and architecture
2. resolve a pinned release version
3. download the matching binary
4. verify checksum/signature
5. install to a user-writable path
6. write MCP client config only after user confirmation or explicit flag
7. print next steps

The npm package should not:

- bundle secrets
- download unpinned `latest` during normal MCP startup
- require admin/root for normal installation
- hide failures behind a successful exit code

Release files should live at repository root unless a design note says otherwise:

```text
.github/workflows/
.goreleaser.yaml
dist/
checksums.txt
packages/npm/
```

Release rules:

- binaries are built from tagged commits
- archives include the `patchxnote` binary and license/readme files only
- checksums are generated for every published artifact
- the npm wrapper installs a pinned binary version matching the package version
- rollback means installing a previous pinned version, not resolving floating `latest`
- public install docs must show tested Windows, macOS, and Linux commands

## Observability And Debugging

For support, collect only safe diagnostics:

- CLI version, commit, OS, architecture
- MCP client name and config path
- server base URL
- HTTP status and stable error code
- request ID from server responses

Never collect sensitive request bodies, tokens, phone numbers, full MAC, SK, raw audio, or transcript content.
