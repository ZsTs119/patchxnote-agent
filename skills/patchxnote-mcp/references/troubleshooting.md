# PatchXNote MCP Troubleshooting

Use this reference when setup, login, tool calls, or client loading fail.

## Existing MCP Entry

If setup reports an existing `patchxnote` entry, stop before replacing it unless the user has clearly authorized replacement. Ask which entry should win, or run a dry-run/manual config path.

Only use:

```sh
npx -y patchxnote-agent@latest setup --client <client-id> --force
```

after the replacement scope is explicit.

## Runtime Mismatch

Symptoms:

- Setup succeeds in one shell but the editor cannot authenticate.
- Windows desktop app works but WSL/SSH/Dev Container does not.
- `mcp status` differs between terminal and editor.

Cause: npm, config files, browser callback, and keychain can differ by runtime.

Fix: rerun setup and login in the same OS/runtime that launches MCP.

## Browser OAuth Problems

Possible causes:

- No GUI browser in headless or remote runtime.
- Localhost callback blocked.
- Callback port already occupied.
- Corporate proxy/firewall intercepts auth.
- User closes the browser before approval.

Use `--no-browser` only for controlled headless/manual URL flows. The user still completes authorization outside chat. Never paste OAuth codes or tokens into chat.

## `mcp serve` Does Not Open Login

This is expected. `mcp serve` is for editor startup and must keep stdout reserved for JSON-RPC. Run `mcp login` or `setup --client <id>` first.

## Windows And WSL

Windows `npx` from a WSL UNC working directory may emit a `CMD.EXE` UNC warning before stdout. Do not treat the warning alone as failed setup, but avoid using UNC cwd for commands where stdout must be pure JSON.

If Node on Windows raises `Error: spawn EINVAL` while starting `npx.cmd`, run through `cmd.exe /d /s /c npx ...` or use the absolute binary fallback printed by:

```sh
npx -y patchxnote-agent@latest install --print-config
```

## Cold Start Or `npx` Timeout

Some clients kill slow first-start processes. Run the installer once:

```sh
npx -y patchxnote-agent@latest install --print-config
```

Then use the printed absolute `patchxnote` command path with `args: ["mcp", "serve"]`.

## Credential Storage Missing

PatchXNote Agent expects OS-native secure storage:

- macOS Keychain
- Windows Credential Manager
- Linux Secret Service

If unavailable, fail closed for public usage. The insecure file keychain is development/CI smoke only and must not be documented as a normal user path.

## Expired Or Wrong Auth

Use:

```sh
npx -y patchxnote-agent@latest mcp logout --local-only
npx -y patchxnote-agent@latest mcp login
```

Run both commands in the same runtime that launches MCP. Do not remove mobile/desktop App credentials or installation state.

## Empty Results

Check:

- Did the user select `mobile` or `desktop`?
- Is the account authenticated in this runtime?
- Is `patchxnote_search_memories` searching only current-session cache?
- Does `patchxnote_list_memories` need pagination?

Do not infer that no PatchXNote data exists across every platform unless both platforms were queried and accepted by the user.

## Tool Count Drift

Tool count changes over time. Use live `tools/list` for the current runtime and report the count as observed evidence, not a permanent fact.
