# PatchNote Agent

PatchNote Agent is the local CLI and MCP bridge for exposing safe PatchNote account tools to desktop agents.

Planned user entry after the first npm publish:

```sh
npx -y patchnote-agent@0.1.1 install
```

The npm package is only an installer wrapper. The long-lived local runtime is a versioned `patchnote` binary that provides:

- `patchnote setup` for first-run login and agent configuration.
- `patchnote login` for phone OTP login.
- `patchnote mcp` for the local stdio MCP server.

Initial MCP scope is read-only:

- current account and profile projection
- bound recorder card list
- quota summary
- structured result listing and detail lookup

The server-side PatchNote API remains the source of truth. This repository owns local distribution, credential storage, MCP tool schema, and desktop agent integration.

## Current MVP Smoke

Local MVP smoke command:

```sh
scripts/e2e/mvp-smoke.sh
```

The smoke builds the `patchnote` binary, runs the npm installer wrapper in dry-run mode, logs in against an in-process Agent V1 test server, checks `auth status`, starts `patchnote mcp serve`, calls the seven V1 MCP tools, logs out, and scans smoke evidence for secret-like values.

Useful local commands:

```sh
go test ./...
go run ./cmd/patchnote version
go run ./cmd/patchnote auth status
go run ./cmd/patchnote mcp serve
node packages/npm/bin/patchnote-agent.js install --dry-run --print-config
```

## Install And Run

After a tagged GitHub release and npm publish exist, install the pinned wrapper:

```sh
npx -y patchnote-agent@0.1.1 install --print-config
```

The installer downloads the matching `patchnote` binary from the GitHub release,
verifies `checksums.txt`, and installs into a user-writable directory. It does
not write bearer tokens into MCP config.

The CLI defaults to the PatchNote test API:

```text
https://ws-lab.patch-x.cn/patchnote-test-api
```

Override the server base URL only when targeting another environment:

```sh
PATCHNOTE_SERVER_BASE_URL=<PatchNote API base URL> patchnote login
```

The first beta build uses an explicit local file credential store until the OS
keychain adapters are shipped. The installer-generated MCP config includes the
same non-secret environment setting, and still does not contain bearer tokens.

```sh
PATCHNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true patchnote login
patchnote auth status
patchnote mcp serve
```

For rollback, reinstall a previous pinned package version instead of using a
floating latest version:

```sh
npx -y patchnote-agent@0.1.1 install
```

To remove the installed binary:

```sh
npx -y patchnote-agent@0.1.1 uninstall
```

## Release Operator Checklist

1. Confirm GoServer production or target environment exposes the required
   `/v1/agent/**` routes.
2. Confirm `packages/npm/package.json` version matches the release tag without
   the leading `v`.
3. Create and push a tag from a clean commit:

   ```sh
   git tag v0.1.1
   git push origin v0.1.1
   ```

4. Wait for the `Release Binaries` workflow to publish:
   `checksums.txt`, Linux/macOS/Windows amd64 and arm64 binaries, and GitHub
   artifact attestations.
5. Verify the release before npm publish:

   ```sh
   gh release view v0.1.1 --repo ZsTs119/patchnote-agent --json assets
   gh attestation verify path/to/patchnote_0.1.1_linux_amd64 --repo ZsTs119/patchnote-agent
   ```

6. Publish the npm wrapper only after binary assets exist:

   ```sh
   gh workflow run publish-npm.yml -f version=0.1.1
   ```

The npm publish workflow requires `NPM_TOKEN` with publish access to the
`patchnote-agent` package on `https://registry.npmjs.org`.

V1 limitations:

- Agent access is read-only and uses dedicated `/v1/agent/...` server routes.
- Recorder-card battery, live BLE state, storage, recording status, SK, full MAC, raw audio, and full transcripts are not exposed.
- `patchnote_search_memories` searches only local authorized metadata cache populated during the MCP session.
- The first beta uses an explicit `PATCHNOTE_AUTH_INSECURE_FILE_KEYCHAIN=true` file credential store. Do not promote it as the final production credential path.
- Real OS keychain adapters, production Agent route rollout, npm ownership, and cross-machine install validation remain release gates before broad public promotion.

## Engineering Rules

Before changing CLI behavior, installer logic, MCP tools, authentication, local cache, or release configuration, read:

- [AGENTS.md](AGENTS.md)
- [docs/engineering-rules.md](docs/engineering-rules.md)
- [docs/plans/2026-08-06-agent-v1-mvp.md](docs/plans/2026-08-06-agent-v1-mvp.md)
