# PatchXNote Agent Repository Rules

These rules apply to the entire `patchnote-agent` repository.

## Read First

Before changing code, installers, release configuration, or MCP tool schemas, read:

1. `README.md`
2. `docs/engineering-rules.md`
3. `docs/plans/2026-08-06-agent-v1-mvp.md`
4. The current PatchXNote server integration contract in `../patchxNoteGoServer/docs/integrations/apifox/integration-guide.zh-CN.md`
5. Any server contract document directly related to the feature being exposed

If this repository conflicts with the server OpenAPI contract, the server contract wins. Record the conflict before implementing around it.

## Product Boundary

- This repository owns the local PatchXNote CLI, local MCP bridge, installer wrapper, local configuration, credential storage integration, and desktop-agent setup.
- The PatchXNote Go Server remains the source of truth for account, quota, hardware binding, model usage, structured results, and authorization.
- Do not copy server domain logic into this repository. Use the public client API or a deliberately designed agent API.
- First-party App/PC clients continue to own MR20 BLE, local ASR, speaker processing, local raw recordings, complete transcripts, local EventBuilder, and Native secure device workflows.

## First Release Scope

The first public Agent/MCP release is read-only:

- current user/account/profile status
- recorder-card list projection
- quota summary
- model usage summary
- structured result listing
- structured result search over locally cached, server-authorized data

Do not add write tools, hardware connect/release tools, model-run execute tools, raw transcript tools, or audio download tools without a new design document and server-side authorization decision.

## Security Rules

- Never log, print, persist in plaintext, commit, or include in examples: access tokens, refresh tokens, installation proof, OTP/review codes, raw phone numbers, full MAC, SK/credential, provider keys, raw audio, complete transcripts, user instructions, or model/provider payloads.
- Local credential material must live in OS-native secure storage when available. Config files may store only non-secret settings such as profile name, base URL, selected platform scope, and local cache path.
- MCP config files must not contain bearer tokens or refresh credentials.
- Do not let Agent login consume or replace `mobile` or `desktop` installations. Agent access must be independent from App/PC installation slots.
- Structured content is platform-scoped. Never merge mobile and desktop results unless the server contract explicitly adds that product behavior.

## CLI Rules

- The core runtime is a versioned Go binary named `patchnote`.
- The npm package is an installer/update wrapper only. Do not make normal MCP startup depend on `npx @patchnote/agent@latest`.
- Use Cobra for command structure and Viper for non-secret configuration once command complexity exceeds the initial placeholder.
- Keep `main.go` thin. Commands should return errors instead of calling `os.Exit` deep inside business logic.
- Human-facing diagnostics go to stderr. Machine-readable command output goes to stdout.
- Support explicit output formats for scriptable commands where practical: `table`, `json`, and `plain`.
- Version, commit, date, and build target should be injected at build time.

## MCP Rules

- The default local MCP transport is stdio.
- For stdio MCP, stdout is reserved for JSON-RPC only. Logs, progress, diagnostics, and errors must go to stderr.
- Keep tool count small and curated. Do not expose the full OpenAPI surface as MCP tools.
- Every MCP tool must have a strict input schema, bounded pagination, bounded output, stable error mapping, and a clear read/write annotation.
- Prefer structured responses plus concise human-readable text. Do not dump large JSON blobs unless the caller explicitly requests JSON and the response is within the output limit.
- Search tools must prefer returning IDs, titles, short snippets, timestamps, and pagination cursors over entire payloads.

## Distribution Rules

- Release binaries must be checksumed and signed before the npm installer downloads them.
- Installer scripts must verify platform, architecture, checksum, and executable permissions before installing.
- Installer scripts must not require administrator privileges for the normal user install path.
- Support uninstall and update paths before public promotion.
- Public docs should show copy-paste commands only after the command is tested on Windows, macOS, and Linux.

## Validation

At minimum, run the relevant checks before finishing a code change:

- `go test ./...`
- CLI command tests for new commands and flags
- MCP inspector or protocol smoke for tool-schema changes
- Installer dry-run checks for npm wrapper changes

For docs-only changes, explain that no runtime tests were needed.
