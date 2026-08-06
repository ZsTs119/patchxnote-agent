# PatchNote Agent Engineering Rules

This document keeps the PatchNote Agent CLI and local MCP bridge from drifting away from the product and security model.

## Architecture

PatchNote Agent has three layers:

1. `patchnote` Go binary
   The local runtime for login, account inspection, MCP stdio serving, local cache management, and installer integration.

2. npm installer wrapper
   A thin package that supports commands such as `npx -y patchnote-agent install`. It installs or updates the versioned Go binary, then exits.

3. PatchNote server APIs
   The remote source of truth. The agent never bypasses server authorization and never reconstructs server facts from local guesses.

The expected runtime shape is:

```text
Agent client -> local MCP stdio -> patchnote binary -> PatchNote API
```

## Repository Layout

The repository uses a thin public binary entrypoint plus internal packages. Keep this structure stable unless a design note explains why it must change.

```text
patchnote-agent/
  cmd/
    patchnote/
      main.go

  internal/
    cli/
    config/
    auth/
    keychain/
    api/
    mcp/
    cache/
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

- `cmd/patchnote`: minimal `package main`; call the CLI package and own only process exit.
- `internal/cli`: Cobra command tree, flags, command wiring, prompt boundaries, and command tests.
- `internal/config`: Viper-backed non-secret config loading, default values, env binding, and config path resolution.
- `internal/auth`: login session state, token refresh orchestration, scope checks, and logout behavior.
- `internal/keychain`: OS secure-storage adapter boundary. Keep platform differences behind this package.
- `internal/api`: PatchNote server client, request/response mapping, retry policy, and stable API error mapping.
- `internal/mcp`: MCP stdio server, tool registry, tool schemas, and JSON-RPC error mapping.
- `internal/cache`: local cache and search index for authorized read-only projections.
- `internal/output`: table/json/plain renderers. Commands should not hand-roll output formatting.
- `internal/diag`: safe diagnostics, redaction, stderr logging, and support bundle helpers.
- `internal/version`: build-time version, commit, date, and target metadata.
- `packages/npm`: npm install/update/uninstall wrapper. Do not place product runtime logic here.

The dependency direction is:

```text
cmd/patchnote -> internal/cli -> internal/{config,auth,api,mcp,cache,output,diag,version}
internal/auth -> internal/{config,keychain,api}
internal/mcp -> internal/{auth,api,cache,diag}
internal/api -> external PatchNote server APIs
```

Avoid reverse dependencies from lower-level packages back into `internal/cli`.

## Non-Goals For V1

V1 is intentionally not a full PatchNote client.

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
macOS:   ~/Library/Application Support/PatchNote Agent/config.yaml
Windows: %AppData%\PatchNote Agent\config.yaml
Linux:   ${XDG_CONFIG_HOME:-~/.config}/patchnote-agent/config.yaml
```

Default cache locations:

```text
macOS:   ~/Library/Caches/PatchNote Agent/
Windows: %LocalAppData%\PatchNote Agent\Cache\
Linux:   ${XDG_CACHE_HOME:-~/.cache}/patchnote-agent/
```

Environment variables must use the `PATCHNOTE_` prefix. For nested config, replace dots and hyphens with underscores, for example `PATCHNOTE_SERVER_BASE_URL`.

## Command Design

The public command surface should stay small:

```text
patchnote setup
patchnote login
patchnote logout
patchnote auth status
patchnote mcp
patchnote mcp install <client>
patchnote mcp status
patchnote version
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

- `cmd/patchnote/main.go` calls `internal/cli.Execute()` and handles final process exit only.
- command handlers use `RunE`
- root commands set `SilenceUsage` and `SilenceErrors`
- shared auth/config checks live in root or group-level `PersistentPreRunE`
- child `PersistentPreRunE` must explicitly call parent setup when both are needed
- use `cmd.OutOrStdout()` for result output and `cmd.ErrOrStderr()` for diagnostics
- command tests construct a fresh command tree and set args/out/err per test
- do not call `os.Exit` outside `cmd/patchnote/main.go`
- do not use package-level mutable Viper state in tests

## MCP Tool Design

Tool names should be service-scoped to avoid collisions:

```text
patchnote_get_current_user
patchnote_list_recorder_cards
patchnote_get_quota_summary
patchnote_get_model_usage_summary
patchnote_list_memories
patchnote_search_memories
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

Agent content access reads only server-stored structured results that the user has consented to store for the selected platform.

V1 may expose:

- event title and summary
- key item title, lifecycle, owner projection, due projection, and short detail
- daily digest date and narrative/summary fields
- source availability status

V1 must not imply access to original audio, complete raw transcript, complete corrected transcript, or speaker identity.

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
- archives include the `patchnote` binary and license/readme files only
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
