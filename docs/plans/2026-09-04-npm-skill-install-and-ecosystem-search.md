# Npm Skill Install And Ecosystem Search Implementation Plan

**Goal:** Make `patchxnote-agent` the npm-only install path for both the PatchXNote MCP runtime and the PatchXNote MCP Agent Skill, while improving discovery through GitHub and the open Agent Skills search ecosystem.

**Architecture:** Keep `skills/patchxnote-mcp/` as the canonical skill source, sync a generated copy into the npm package before publishing, and add `patchxnote-agent skill install` as the stable user command. Treat skills.sh / `npx skills find` as a discovery channel only; do not make it the primary install dependency.

**Tech Stack:** Node.js npm wrapper, packaged Markdown Agent Skill, existing sync/validation scripts, npm release flow, GitHub repository metadata, `npx skills` discovery checks.

**Execution Rule:** Work sequentially in the primary agent. Do not use sub-agents or parallel task execution. Do not claim marketplace acceptance or search indexing until verified by the specific platform/search command.

---

## Execution Status On 2026-09-04

| Task | Status | Evidence |
| --- | --- | --- |
| Task 0: Required Preflight | `done` | Required repository docs and server integration contract were read; `git status --short --branch` started from only this plan file untracked; npm latest was `0.2.10`. |
| Task 1: Add The Npm-Packaged Skill Copy | `done_local` | Sync script now writes `packages/npm/skills/patchxnote-mcp/`; `npm pack --dry-run --json` listed the skill and references. |
| Task 2: Implement `patchxnote-agent skill install` | `done_local` | npm wrapper supports `skill install`, `--dry-run`, `--json`, `--agent`, `--home`, `--force`, and `--copy`; `node packages/npm/test/install.test.js` passed. |
| Task 3: Update User-Facing One-Liners | `done_local` | English, Chinese, npm README, starter prompts, and onboarding reference use `npx -y patchxnote-agent@latest skill install` before MCP setup. |
| Task 4: Improve GitHub And Skill Search Metadata | `partial_done` | npm keywords and skill metadata were updated; `npx skills find patchxnote` and owner-filtered search returned no results, so search indexing is still separate. |
| Task 5: Decide Whether A Lightweight Skill Mirror Is Needed | `deferred` | Existing repo remains canonical; no mirror was created because npm-bundled install is now the primary path and search indexing has not been accepted. |
| Task 6: Local Validation | `done_local_with_go_note` | sync/validate, skill validator, plugin validator, client registry validator, npm wrapper tests, temp-package install, npm pack dry-run, and `git diff --check` passed; full WSL `go test ./...` is blocked by Go 1.18, while a Windows Go temp-copy package subset passed except Windows-only keychain POSIX mode coverage. |
| Task 7: Release `0.2.11` | `pending` | Version metadata is prepared; GitHub tag/release and npm publish are not completed yet. |
| Task 8: Published MCP Smoke | `pending` | Must run after npm latest resolves to `0.2.11`. |
| Task 9: Search/Discovery Acceptance | `not_indexed` | `npx -y skills find patchxnote` and `npx -y skills find patchxnote --owner ZsTs119` returned no results on 2026-09-04. |
| Task 10: Evidence/Rollback | `partial_done` | Evidence log and `docs/evidence/2026-09-04-release-0.2.11.zh-CN.md` record local candidate evidence; publish evidence is pending. |

---

## Scope

- [ ] Ship npm package support for installing the PatchXNote MCP skill.
- [ ] Keep the existing MCP setup flow unchanged except for docs that mention the new skill install step.
- [ ] Improve discoverability for GitHub search and `npx skills find`.
- [ ] Publish a patch release only after local package and MCP smoke checks pass.

## Non-Goals

- [ ] Do not submit OpenAI, Cursor, Claude, Smithery, Glama, PulseMCP, MCP.so, or MCP Registry listings in this task.
- [ ] Do not add a hosted remote MCP transport.
- [ ] Do not require users to install or depend on `npx skills` for the primary install path.
- [ ] Do not add executable scripts inside the skill unless they are necessary and reviewed.
- [ ] Do not change MCP tool schemas or server read/write scope.

## Current Baseline To Preserve

- [ ] `patchxnote-agent@0.2.10` is the current published npm version.
- [ ] `npx -y patchxnote-agent@latest setup --client <client-id>` remains the primary MCP setup command.
- [ ] `npx -y patchxnote-agent@latest mcp config` remains the fallback generic stdio config command.
- [ ] The canonical skill lives at `skills/patchxnote-mcp/SKILL.md`.
- [ ] The npm package currently includes only `bin/patchxnote-agent.js`, `README.md`, and `package.json`; the skill is not yet packaged.
- [ ] Direct `npx skills add ZsTs119/patchxnote-agent --skill patchxnote-mcp -g` may hang because it clones the full repository.

## Task 0: Required Preflight

**Files To Read Before Edits:**
- `README.md`
- `README.zh-CN.md`
- `docs/engineering-rules.md`
- `docs/release-and-maintenance-runbook.zh-CN.md`
- `docs/plans/2026-08-06-agent-v1-mvp.md`
- `docs/plans/2026-09-03-patchxnote-mcp-skill-marketplace-checklist.md`
- `../patchxNoteGoServer/docs/integrations/apifox/shared/integration-guide.zh-CN.md`

**Checklist:**

- [ ] Run `git status --short --branch` and record whether the worktree is clean.
- [ ] Confirm the current published version from the official npm registry:

```sh
npm view patchxnote-agent@latest version --registry https://registry.npmjs.org
```

- [ ] Confirm no unrelated user changes will be edited, staged, or reverted.
- [ ] Re-check the release runbook before bumping, tagging, pushing, or publishing.
- [ ] If the server integration contract conflicts with the local Agent docs, record the conflict and keep the server contract as source of truth.
- [ ] Confirm this task is still scoped to npm skill packaging and discovery metadata, not external marketplace submission.

## Task 1: Add The Npm-Packaged Skill Copy

**Files:**
- Modify: `scripts/sync-patchxnote-skill-packages.mjs`
- Modify: `scripts/validate-patchxnote-skill-packages.mjs`
- Create: `packages/npm/skills/patchxnote-mcp/SKILL.md`
- Create: `packages/npm/skills/patchxnote-mcp/references/*.md`
- Modify: `packages/npm/package.json`

**Checklist:**

- [ ] Add `packages/npm/skills/patchxnote-mcp/` as a generated copy of `skills/patchxnote-mcp/`.
- [ ] Update the sync script so the npm skill copy is regenerated from the canonical skill source.
- [ ] Update validation so canonical, npm, OpenAI plugin, and Claude plugin skill copies cannot drift.
- [ ] Add `packages/npm/skills/patchxnote-mcp` to validation and secret-scan roots.
- [ ] Keep all generated skill copies text-normalized with LF line endings so cross-platform `npm pack` output is stable.
- [ ] Update npm `files` to include `skills`.
- [ ] Confirm `npm pack --dry-run` lists `package/skills/patchxnote-mcp/SKILL.md`.
- [ ] Confirm the npm package still excludes credentials, local caches, temp files, generated binaries, and release evidence logs unless deliberately included.
- [ ] Confirm published package size remains small enough that `npx` cold start is not materially slowed by the skill bundle.

## Task 2: Implement `patchxnote-agent skill install`

**Files:**
- Modify: `packages/npm/bin/patchxnote-agent.js`
- Modify: `packages/npm/test/install.test.js`
- Modify as needed: `packages/npm/README.md`

**Command Shape:**

```sh
npx -y patchxnote-agent@latest skill install
```

**Supported Options:**

- [ ] `--dry-run`: print the planned skill source, target directories, action, and package version without writing.
- [ ] `--json`: print machine-readable output for tests and automation.
- [ ] `--agent <id>`: target a known client or directory family. Initial IDs: `universal`, `codex`, `cursor`, `claude-code`, `gemini-cli`, `github-copilot`, `all`.
- [ ] `--home <path>` or `PATCHXNOTE_AGENT_SKILL_HOME=<path>`: test-only/home override so smoke tests do not write into the real user profile.
- [ ] `--force`: overwrite an existing non-managed `patchxnote-mcp` skill after explicit user intent.
- [ ] `--copy`: optional no-op or explicit copy mode only if symlink support is considered later; V1 should copy files and not create symlinks.

**Install Behavior:**

- [ ] Default install writes to the universal Agent Skills directory first: `<home>/.agents/skills/patchxnote-mcp`.
- [ ] If the runtime clearly identifies Codex, Cursor, Claude Code, Gemini CLI, or GitHub Copilot, also write the known client-specific directory when it is safe and documented.
- [ ] Do not create every possible client directory by default. Only `--agent all` may write multiple client-specific directories.
- [ ] Dedupe target directories before writing, because several agents may resolve to the same `.agents/skills` location.
- [ ] Keep target directory mapping explicit in code and covered by tests for Windows, macOS, and Linux path shapes.
- [ ] Always print the exact installed paths.
- [ ] Create parent directories with normal user permissions only; do not require admin/root.
- [ ] Copy only files from the packaged `skills/patchxnote-mcp/` directory.
- [ ] Preserve relative reference links inside `SKILL.md`.
- [ ] Fail clearly if the packaged skill directory is missing instead of downloading from GitHub at runtime.
- [ ] Install atomically: write to a temp directory under the target parent, then rename/replace only after all files are written and verified.
- [ ] Clean up temp install directories after success or failure.
- [ ] Store a small managed marker file, such as `.patchxnote-agent-skill.json`, with package version and source hash.
- [ ] Hash only the packaged skill files, not timestamps or local absolute paths.
- [ ] If an existing skill has a managed marker from this package, update it in place.
- [ ] If an existing skill has identical content but no marker, add the marker.
- [ ] If an existing skill differs and has no marker, fail with a clear message unless `--force` is supplied.
- [ ] If `--force` is supplied, replace only `patchxnote-mcp`; never delete sibling skills or parent agent directories.
- [ ] Route human-facing diagnostics to stderr only when the command is in `--json` mode, so stdout remains parseable JSON.
- [ ] Never read, write, print, or package OTP codes, OAuth codes, access tokens, refresh tokens, webhook secrets, full phone numbers, full MAC values, SK values, raw audio, full transcripts, prompts, or provider payloads.

**Tests:**

- [ ] Parse `skill install` and all new options.
- [ ] Unknown `skill` subcommands fail with `usage: patchxnote-agent skill install`.
- [ ] `--dry-run --json` returns deterministic target paths and no secrets.
- [ ] Fresh install into a temp home creates `SKILL.md`, references, and marker file.
- [ ] Re-running install is idempotent when content is unchanged.
- [ ] Managed older install updates cleanly.
- [ ] Unmanaged modified install fails without `--force`.
- [ ] `--force` overwrites the unmanaged modified install and records a marker.
- [ ] Unknown `--agent` fails with a clear supported-values message.
- [ ] Windows-style paths, POSIX paths, and paths containing spaces are covered.
- [ ] A missing packaged skill directory produces a clear failure.
- [ ] `--json` output is valid JSON and stderr contains any human diagnostics.

## Task 3: Update User-Facing One-Liners

**Files:**
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `packages/npm/README.md`
- Modify: `docs/marketplace/starter-prompts.md`
- Modify if needed: `docs/marketplace/agent-skills-install.md`

**Checklist:**

- [ ] Add the npm-only skill install command before MCP setup.
- [ ] Do not keep `npx skills add ...` in the primary one-line setup prompt.
- [ ] Keep the English one-liner and Chinese one-liner semantically equivalent.
- [ ] Keep both source references:
  - [ ] `https://github.com/ZsTs119/patchxnote-agent`
  - [ ] `https://patchx2025.feishu.cn/wiki/PnVRwYT7IirFPckairGcWPnHnCd`
- [ ] Keep the instruction to identify the local MCP client and run setup in the same OS/runtime that will later launch MCP.
- [ ] Keep the rule that login opens a browser and the user completes phone-code authorization there.
- [ ] Keep the rule that the agent must not ask for OTP codes, OAuth codes, authorization codes, access tokens, refresh tokens, or webhook secrets in chat.
- [ ] Keep the verification tools and parameters:

```json
{"platform":"mobile","limit":5}
```

- [ ] State that `npx skills` is an optional discovery path, not the required install path.

## Task 4: Improve GitHub And Skill Search Metadata

**Files:**
- Modify: `skills/patchxnote-mcp/SKILL.md`
- Modify: `packages/npm/skills/patchxnote-mcp/SKILL.md` through sync only
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `packages/npm/package.json`
- Modify: `docs/marketplace/agent-skills-install.md`
- Modify if needed: `docs/marketplace/platform-matrix.zh-CN.md`

**Checklist:**

- [ ] Add natural search terms to the skill description without keyword stuffing:
  - [ ] `PatchXNote`
  - [ ] `PatchXNote MCP`
  - [ ] `Agent Skill`
  - [ ] `MCP server`
  - [ ] `recordings`
  - [ ] `summaries`
  - [ ] `memories`
  - [ ] `Codex`
  - [ ] `Cursor`
  - [ ] `Claude Code`
  - [ ] `stdio`
  - [ ] `phone-code login`
- [ ] Keep the description narrow enough that generic summarization does not trigger this skill.
- [ ] Add a README section titled or phrased around `PatchXNote MCP Agent Skill`.
- [ ] Add npm `keywords` for npm search, including `patchxnote`, `mcp`, `mcp-server`, `agent-skill`, `codex`, `cursor`, and `claude-code`.
- [ ] Keep npm `description` concise but mention both local MCP and skill installation if it remains accurate.
- [ ] Add GitHub repository topics after code is pushed:
  - [ ] `patchxnote`
  - [ ] `patchxnote-mcp`
  - [ ] `mcp`
  - [ ] `mcp-server`
  - [ ] `agent-skills`
  - [ ] `skills-sh`
  - [ ] `codex`
  - [ ] `cursor`
  - [ ] `claude-code`
- [ ] Use GitHub UI or `gh repo edit --add-topic ...` for topics; do not fake topics in repository files.
- [ ] Record whether topics were actually added in the evidence log.

## Task 5: Decide Whether A Lightweight Skill Mirror Is Needed

**Decision Gate:** Only create or prepare a separate `patchxnote-mcp-skill` repository if the current repository remains hard to install or discover through the open skills ecosystem after npm-only install is released.

**Checklist:**

- [ ] First try discovery with the existing public repo after metadata updates and publish.
- [ ] Run:

```sh
npx -y skills find patchxnote
npx -y skills find "patchxnote mcp"
npx -y skills find patchxnote --owner ZsTs119
```

- [ ] If search is not indexed immediately, record `indexed_pending` rather than treating it as a failure.
- [ ] Re-check after the expected indexing window.
- [ ] If `npx skills add ZsTs119/patchxnote-agent --skill patchxnote-mcp -g --yes` still hangs or times out, keep npm-only install as the public command.
- [ ] If a lightweight mirror is needed, prepare a separate plan before creating it.
- [ ] A mirror repository must be source-generated from the canonical skill and must point back to `patchxnote-agent` npm as the real installer.
- [ ] Do not create a mirror repo in this implementation without explicit user approval, because it adds another release and drift surface.

## Task 6: Local Validation Before Release

**Commands:**

```sh
node scripts/sync-patchxnote-skill-packages.mjs --check
node scripts/validate-patchxnote-skill-packages.mjs
node --check packages/npm/bin/patchxnote-agent.js
cd packages/npm && npm test
cd packages/npm && npm pack --dry-run
```

**Checklist:**

- [ ] Sync check passes.
- [ ] Skill/package validator passes.
- [ ] npm wrapper tests pass.
- [ ] `npm pack --dry-run` contains `skills/patchxnote-mcp`.
- [ ] Local tarball smoke installs the skill into a temp home.
- [ ] Local tarball smoke uses a temp home, not the real user global skill directory.
- [ ] Local tarball smoke does not write secrets or print secrets.
- [ ] Local tarball smoke verifies the marker file and source hash.
- [ ] Existing MCP wrapper checks still pass:

```sh
npx -y patchxnote-agent@latest mcp config
npx -y patchxnote-agent@latest setup --client cursor --dry-run --print-config
node scripts/smoke-mcp-stdio.mjs
```

- [ ] If the published-latest smoke uses the old version before release, mark that result as baseline only.

## Task 7: Release `0.2.11`

**Files:**
- Modify: `packages/npm/package.json`
- Modify: `server.json`
- Modify if present: any version references required by the release runbook
- Modify after release: `docs/marketplace/evidence-log.md`

**Checklist:**

- [ ] Bump npm package version from `0.2.10` to `0.2.11`.
- [ ] Bump `server.json` top-level `version` to `0.2.11`.
- [ ] Bump `server.json` npm package `version` to `0.2.11`.
- [ ] Run all local validation again after the version bump.
- [ ] Commit with a clear release-prep message.
- [ ] Tag `v0.2.11`.
- [ ] Push commit and tag only when the current task has explicit release approval.
- [ ] Confirm GitHub release workflow succeeds.
- [ ] Confirm npm publish workflow succeeds.
- [ ] Confirm npm latest:

```sh
npm view patchxnote-agent@latest version --registry https://registry.npmjs.org
```

- [ ] Confirm the published npm package includes the skill:

```sh
npm pack patchxnote-agent@latest --dry-run --registry https://registry.npmjs.org
```

- [ ] Confirm published npm-only skill install works from a clean temp home:

```sh
npx -y patchxnote-agent@latest skill install --dry-run --json
npx -y patchxnote-agent@latest skill install --home <temp-home> --agent universal --json
```

- [ ] Confirm the installed temp-home `SKILL.md` is present and matches the published package content.

## Task 8: Published MCP Smoke

**Checklist:**

- [ ] Confirm generic MCP config is secret-free:

```sh
npx -y patchxnote-agent@latest mcp config
```

- [ ] Confirm setup dry-run still works:

```sh
npx -y patchxnote-agent@latest setup --client cursor --dry-run --print-config
```

- [ ] Run stdio MCP smoke with published latest:

```sh
node scripts/smoke-mcp-stdio.mjs
```

- [ ] Verify the smoke includes real MCP tool calls:
  - [ ] `tools/list`
  - [ ] `patchxnote_get_current_user`
  - [ ] `patchxnote_list_memories` with `{"platform":"mobile","limit":5}`
- [ ] Record whether this is local npm stdio evidence, current-client mounted-tool evidence, or both.
- [ ] If no authenticated PatchXNote credential is available, record `auth_required` or `blocked_no_credential`; do not relabel config-only success as real MCP data access.
- [ ] If current-client mounted MCP tools remain stale after npm release, reconnect/reload the client before counting current-client acceptance.

## Task 9: Search And Discovery Acceptance

**Checklist:**

- [ ] Confirm GitHub code search can find the skill by `patchxnote-mcp` or `PatchXNote MCP Agent Skill`.
- [ ] Confirm GitHub repository search can find the repo by `patchxnote mcp skill` after topic updates.
- [ ] Run:

```sh
npx -y skills find patchxnote
npx -y skills find "patchxnote mcp"
npx -y skills find patchxnote --owner ZsTs119
```

- [ ] If `npx skills find` returns the skill, record `indexed_smoked`.
- [ ] If `npx skills find` does not return the skill, record `indexed_pending` with command output and re-check date.
- [ ] If skills.sh has a manual submission or indexing request path available at that time, document and use it only after verifying current instructions.
- [ ] Treat `npx skills add` success and `npx skills find` success as separate gates.
- [ ] Do not claim skills.sh discovery until the `find` command or website search visibly returns the skill.
- [ ] Keep public docs phrased as "install from npm; discoverability pending/available through GitHub and skills search" based on actual evidence.

## Task 10: Evidence And Rollback

**Files:**
- Modify: `docs/marketplace/evidence-log.md`
- Modify if needed: `docs/marketplace/platform-matrix.zh-CN.md`
- Modify if needed: `docs/marketplace/rollback-and-deprecation.zh-CN.md`

**Checklist:**

- [ ] Record local validation evidence.
- [ ] Record npm publish evidence.
- [ ] Record published package tarball evidence.
- [ ] Record published `skill install` smoke evidence.
- [ ] Record published MCP stdio smoke evidence.
- [ ] Record GitHub topics/search evidence.
- [ ] Record `npx skills find` evidence as `indexed_smoked` or `indexed_pending`.
- [ ] If `0.2.11` has a packaging or install regression, prepare `0.2.12` patch release rather than relying on npm unpublish.
- [ ] If the skill copy drifts from canonical source, stop release and fix sync validation before publishing.
- [ ] If a real-profile skill install was accidentally tested, record exact touched paths and restore only files created by this task.
- [ ] Leave any unrelated user skill directories untouched.

## Final Acceptance Criteria

- [ ] `patchxnote-agent@latest` installs the PatchXNote MCP skill through npm only.
- [ ] The npm tarball contains `skills/patchxnote-mcp/SKILL.md` and references.
- [ ] `skill install` is idempotent and does not clobber user-modified skills without `--force`.
- [ ] README one-liners include both skill install and MCP setup.
- [ ] GitHub and skill-search metadata are improved and recorded.
- [ ] Search indexing status is verified or explicitly recorded as pending.
- [ ] MCP setup and stdio smoke still work after the npm change.
- [ ] Evidence log clearly separates `published_smoked`, `indexed_pending`, `indexed_smoked`, and any future marketplace acceptance.
