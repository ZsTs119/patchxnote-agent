# README Assets 0.2.4 Release Implementation Checklist

> **For Codex:** REQUIRED SKILL: Use `executing-plans` to implement this plan task-by-task. Execute sequentially in the primary agent only. Do not use sub-agents or parallel task execution.

**Goal:** Update PatchXNote Agent public documentation and visual assets so the `0.2.4` release clearly explains the new record lookup, AI整理结果查看, model IO export, and webhook delivery workflows.

**Execution note (2026-08-14):** `0.2.4` was published first. Post-publish local acceptance found that webhook aliases containing dots were written to YAML but not read back through Viper. The accepted public release therefore moved to `0.2.5`, which keeps the same user-facing feature set and adds the alias persistence fix.

**Architecture:** Treat README, npm README, visual assets, and release metadata as one public contract. Use user-facing language first, then list exact CLI and MCP tool names for users and AI hosts that need them. Keep the security boundary precise: server-backed PatchXNote data remains read-only, while local webhook configuration and manual sends are explicit local side effects.

**Tech Stack:** Markdown, npm package metadata, GitHub Release, npm publish, Go CLI tests, MCP stdio smoke, image generation or controlled raster asset replacement under `docs/assets`.

---

## Current Facts

- [ ] `main` currently contains the new webhook, model IO field, and model IO trace discovery implementation.
- [ ] The `0.2.4` tag must include the already-merged feature commit that added webhook MCP tools and model IO discovery; do not tag only a docs branch that misses those code changes.
- [ ] Public npm latest is still `0.2.3`.
- [ ] Latest Git tag is currently `v0.2.3`.
- [ ] The next public release should be `0.2.4`.
- [ ] MCP tool count is now 19.
- [ ] `packages/npm` has no package-lock file; version consistency is mainly `packages/npm/package.json`, README examples, release tag, and built binary metadata.
- [ ] Existing README text has partial `19` tool updates but still sounds too engineering-heavy.
- [ ] Existing README still overuses internal terms:
  - `structured-result metadata`
  - `memory`
  - `model IO trace`
  - `provider response`
  - `parsed result`
  - `packaged result`
  - `safe plaintext projection`
- [ ] Existing README has an inaccurate user-facing limitation: it groups transcripts as unsupported even though explicit source text inspection is now supported.
- [ ] Existing docs/assets images contain outdated facts:
  - `patchxnote-agent-tools.png` says 7 read-only MCP tools.
  - `patchxnote-agent-safety-boundary.png` says original content/full transcript is not read.
  - `patchxnote-agent-architecture.png` implies only read-only Agent behavior and does not show local webhook sending.
  - `patchxnote-agent-quickstart.png` only mentions account, recorder card, quota, and metadata.
  - `patchxnote-agent-cover.png`, `patchxnote-agent-feishu-cover.png`, and `patchxnote-agent-social-preview.png` still position the product as read-only memory metadata only.
- [ ] `docs/readme-benchmark.zh-CN.md` is outdated and still says the MCP tool table lists seven V1 tools.
- [ ] No raw phone number, OTP, access token, refresh token, webhook URL, raw audio, full source text, user prompt, or provider payload should appear in README, images, examples, release notes, or generated assets.

## Review Additions From Plan Audit

- [ ] Add a clear release-status guard: before npm publish completes, local docs can prepare `0.2.4`, but final public claims must be verified against GitHub Release and npm latest.
- [ ] Add a fallback for generated images with Chinese text: if image-generation text is garbled, rebuild the asset programmatically with local fonts and simple shapes instead of accepting unreadable text.
- [ ] Preserve or document image generation inputs. If there is no source design file, record each image's final text and design intent in `docs/readme-benchmark.zh-CN.md` or a small asset note so future updates do not guess from PNGs.
- [ ] Verify image links and filenames after regeneration; README links depend on the current PNG filenames.
- [ ] Run `npm pack --dry-run` in addition to installer tests so the npm package actually contains the updated package README and installer files.
- [ ] Verify fresh install from the published `0.2.4` package after npm publish, not only `--from-local`.
- [ ] Add a post-release check that GitHub Release binary `patchxnote version` and MCP `serverInfo.version` both report `0.2.4`.
- [ ] Do not overpromise that record search always returns content. Some accounts/platforms may have an empty record list; document that users should check `mobile` vs `desktop`, and use AI整理记录 lookup when model traces exist without普通记录.
- [ ] Keep public examples generic. Do not include the real Feishu test robot URL, real request IDs, real phone/account IDs, or snippets from local user data.
- [ ] Public guide links live outside this repo. Prepare updated copy/assets for the Feishu guide, but do not claim the Feishu public guide is updated until that external page is actually edited.
- [ ] Do not describe the package as open source if repository/package license remains `UNLICENSED`.

## Product Language Decisions

- [ ] Use "PatchXNote Agent 是 PatchXNote 的本地 AI 助手连接器" as the Chinese product anchor.
- [ ] Use "让 AI 帮你查看记录、整理内容、生成 Markdown，并在你确认后发送到飞书、钉钉或其他 webhook" as the main value statement.
- [ ] Keep "CLI" and "MCP" where setup requires them, but explain them in plain language before tool tables.
- [ ] Prefer these user-facing terms:

| Internal term | User-facing term |
| --- | --- |
| `memory` | 记录 |
| `structured-result metadata` | 记录列表和基础信息 |
| `model IO trace` | AI 整理记录 / AI 处理记录 |
| `request_id` | 处理编号, with `request_id` shown in command examples |
| `provider response` | AI 返回内容 |
| `parsed result` | AI 解析后的结果 |
| `packaged result` | 最终整理结果 |
| `source text` / safe plaintext projection | 可查看的原文文本 / 安全文本 |
| `generic webhook` | 其他 webhook 地址, with `generic` shown in command examples |

- [ ] Do not describe the whole Agent as "read-only" without qualification.
- [ ] Correct public boundary wording:
  - [ ] PatchXNote server data access is read-only.
  - [ ] Local webhook target configuration writes local non-secret metadata and stores webhook URL/secret in the local secure store.
  - [ ] Webhook sending only happens when the user or AI explicitly calls a send command/tool.
  - [ ] Source text and AI results may be sensitive and should be accessed explicitly, preferably exported to local files for large content.
- [ ] Do not claim `0.2.4` is published until GitHub Release and npm publish are complete.
- [ ] After version bump and release execution starts, README may use `0.2.4` commands because this checklist is preparing that release.

## User-Facing Feature Facts To Publish

- [ ] Install once, connect to an MCP-capable AI assistant.
- [ ] Login with phone OTP without pasting OTP, access token, or refresh token into chat.
- [ ] Agent login is independent from App/PC mobile/desktop installation slots.
- [ ] AI can check account status, recorder card list, quota, and model usage.
- [ ] AI can view mobile/desktop record lists and search records.
- [ ] AI can view one record's safe basic information.
- [ ] AI can find an AI整理记录 and then inspect:
  - [ ] 可查看的原文文本
  - [ ] AI 返回内容
  - [ ] AI 解析后的结果
  - [ ] 最终整理结果
- [ ] AI can export model IO fields to local files.
- [ ] Users can configure multiple webhook targets.
- [ ] Each webhook target has a custom alias, including Chinese names and spaces.
- [ ] Webhook target types are Feishu, DingTalk, and generic.
- [ ] Webhook URL/signing secret inputs are write-only and are not listed back.
- [ ] AI or CLI can render PatchXNote record content into Markdown drafts.
- [ ] Users can edit local Markdown drafts before sending.
- [ ] Webhook sending is manual and explicit only.
- [ ] Agent does not automatically push records in the background.
- [ ] Agent does not expose raw audio or audio download.
- [ ] Agent does not bind/release/recover hardware.
- [ ] Agent does not execute model runs.
- [ ] Agent does not handle payment, purchase, Admin API, or App/PC device workflows.

## Target README Structure

- [ ] Hero image.
- [ ] One-sentence plain-language product statement.
- [ ] Two short paragraphs:
  - [ ] what it helps users do;
  - [ ] the safety boundary.
- [ ] One-line AI prompt for installation.
- [ ] Manual install command.
- [ ] At A Glance table.
- [ ] Features table with user-facing wording.
- [ ] Requirements.
- [ ] Quickstart.
- [ ] Common user workflows.
- [ ] MCP configuration.
- [ ] MCP tools grouped by purpose.
- [ ] CLI commands grouped by purpose.
- [ ] Security and risk notice.
- [ ] Current limitations.
- [ ] Troubleshooting.
- [ ] Verification and release maintenance entry.
- [ ] License/status.

## Task 1: Version And Release Fact Prep

**Files:**
- Modify: `packages/npm/package.json`
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `packages/npm/README.md`
- Modify: `docs/release-and-maintenance-runbook.zh-CN.md` if version-specific examples need adjustment
- Inspect: `.github/workflows/release.yml`
- Inspect: `.github/workflows/publish-npm.yml`
- Inspect: `.goreleaser.yaml`

**Checklist:**

- [ ] Confirm latest public version:

```sh
npm view patchxnote-agent version
git tag --sort=-v:refname | head
```

- [ ] Confirm `0.2.4` is unused on npm and Git tags.
- [ ] Confirm release workflows infer binary version from tag `v0.2.4`.
- [ ] Confirm npm publish workflow packages `packages/npm`.
- [ ] Change `packages/npm/package.json` from `0.2.3` to `0.2.4`.
- [ ] Replace user-facing install examples:

```sh
npx -y patchxnote-agent@0.2.4 install --print-config
```

- [ ] Keep unversioned install examples where appropriate:

```sh
npx -y patchxnote-agent install --print-config
```

- [ ] Ensure no docs claim release success before release is complete.
- [ ] Keep package license/status wording aligned with `packages/npm/package.json`; do not add "open source" language while license is `UNLICENSED`.

**Verification:**

```sh
node -e "console.log(require('./packages/npm/package.json').version)"
```

Expected: `0.2.4`.

## Task 2: Rewrite README.zh-CN Product Positioning

**Files:**
- Modify: `README.zh-CN.md`

**Checklist:**

- [ ] Replace the intro with plain user language:

```text
PatchXNote Agent 是 PatchXNote 的本地 AI 助手连接器。安装后，你可以让 AI 帮你查看 PatchXNote 记录、读取 AI 整理结果、生成 Markdown，并在你确认后发送到飞书、钉钉或其他 webhook。
```

- [ ] Add one short paragraph for what users can do:

```text
你可以让 AI 查找手机端或电脑端同步过来的记录，查看某次 AI 整理背后的原文文本、AI 返回内容和最终整理结果，也可以把整理好的内容保存成本地草稿，再手动发送到指定机器人。
```

- [ ] Add one short paragraph for safety:

```text
PatchXNote 服务端数据访问仍是只读的。Agent 不操作硬件绑定、不读取原始音频、不处理支付或 Admin API。webhook 配置和发送只发生在你的本机，并且只有你或 AI 明确调用发送命令时才会发出去。
```

- [ ] Update `快速了解` table:
  - [ ] Data access: "查看账号、录音卡、额度、记录列表、AI 整理结果。"
  - [ ] Webhook: "本地配置别名，手动发送到飞书/钉钉/其他 webhook。"
  - [ ] Package status: "公开 beta 版 `0.2.4`。"
- [ ] Update `功能` table:
  - [ ] "记录列表和搜索" supported.
  - [ ] "AI 整理记录查看" supported.
  - [ ] "原文文本和 AI 结果导出" supported.
  - [ ] "多 webhook 别名配置" supported.
  - [ ] "Markdown 草稿" supported.
  - [ ] "原始音频/音频下载" unsupported.
  - [ ] "自动发送 webhook" unsupported.
  - [ ] "模型执行/重放" unsupported.
- [ ] Remove or rewrite "完整转写不支持" so it does not conflict with source text inspection.
- [ ] Add `常用场景` section before MCP config:

```text
帮我找今天手机端的记录。
查看这条记录的 AI 整理结果。
把这段 Markdown 发到“产品群 飞书”。
把 AI 返回内容导出到本地 JSON 文件。
把这条记录生成 Markdown 草稿，我改完再发送。
```

- [ ] Add an empty-state note in FAQ or limitations:

```text
如果记录列表为空，先确认选择的是 mobile 还是 desktop。普通记录和 AI 整理记录不是同一个列表；有些模型处理记录可能能在 AI 整理记录里查到，但普通记录列表暂时为空。
```

**Verification:**

```sh
grep -nE "0.2.3|7 个只读|18 个|完整转写/下载|structured-result|model IO trace|provider response" README.zh-CN.md
```

Expected: no stale user-facing text remains except command/tool names where intentionally preserved.

## Task 3: Rewrite README.md In English

**Files:**
- Modify: `README.md`

**Checklist:**

- [ ] Mirror the Chinese README structure and facts.
- [ ] Use plain English:
  - [ ] "local AI assistant connector"
  - [ ] "records"
  - [ ] "AI整理结果" as "AI-generated result" or "AI processing result"
  - [ ] "request ID" only when explaining commands
- [ ] Update feature matrix to `0.2.4`.
- [ ] Correct raw audio/source text boundary:
  - [ ] "Raw audio and audio downloads: No"
  - [ ] "Explicit source text and AI result inspection: Yes"
- [ ] Add common workflow examples:

```text
Find today's mobile records.
Show the AI result behind this record.
Send this Markdown to my Product Feishu webhook.
Export the AI response to a local JSON file.
Create a Markdown draft from this record so I can edit it before sending.
```

- [ ] Keep tool names exact in the MCP tools table.
- [ ] Avoid claiming auto-configuration for all MCP hosts.

**Verification:**

```sh
grep -nE "0.2.3|7 read-only|18|structured-result metadata|model IO trace|full transcripts/downloads" README.md
```

Expected: no stale user-facing text remains except exact API or tool names where intentional.

## Task 4: Update MCP Tool Tables

**Files:**
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `packages/npm/README.md`
- Modify: `docs/readme-benchmark.zh-CN.md`

**Checklist:**

- [ ] Group tools into three sections instead of one long undifferentiated table:

```text
账号和记录查询: 7
Webhook 配置和发送: 7
AI 整理结果查看: 5
```

- [ ] Keep exact 19 tool names:
  - [ ] `patchxnote_get_current_user`
  - [ ] `patchxnote_list_recorder_cards`
  - [ ] `patchxnote_get_quota_summary`
  - [ ] `patchxnote_get_model_usage_summary`
  - [ ] `patchxnote_list_memories`
  - [ ] `patchxnote_search_memories`
  - [ ] `patchxnote_get_memory`
  - [ ] `patchxnote_list_webhook_targets`
  - [ ] `patchxnote_configure_webhook_target`
  - [ ] `patchxnote_remove_webhook_target`
  - [ ] `patchxnote_list_webhook_templates`
  - [ ] `patchxnote_render_webhook_message`
  - [ ] `patchxnote_export_model_io`
  - [ ] `patchxnote_send_webhook`
  - [ ] `patchxnote_list_model_io_traces`
  - [ ] `patchxnote_get_model_io_source_text`
  - [ ] `patchxnote_get_model_io_provider_response`
  - [ ] `patchxnote_get_model_io_parsed_result`
  - [ ] `patchxnote_get_model_io_packaged_result`
- [ ] Use user-facing explanations in table descriptions.
- [ ] Mention that exact tool names are for MCP Hosts and AI assistants.
- [ ] Update benchmark docs from "seven V1 tools" to "19 tools grouped by user task."

**Verification:**

```sh
grep -R "7 个 V1\\|seven V1\\|18 个\\|18 V1" README.md README.zh-CN.md packages/npm/README.md docs/readme-benchmark.zh-CN.md docs/release-and-maintenance-runbook.zh-CN.md
```

Expected: no stale count.

## Task 5: Update CLI Examples

**Files:**
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `packages/npm/README.md`

**Checklist:**

- [ ] Split examples into sections:
  - [ ] Install and login
  - [ ] Records and AI results
  - [ ] Webhook targets
  - [ ] Drafts and sends
- [ ] Include record and AI result commands:

```sh
patchxnote model-io list --platform mobile
patchxnote model-io source-text --request-id <request_id> --platform mobile --out ./source.txt
patchxnote model-io provider-response --request-id <request_id> --platform mobile --out ./provider-response.json
patchxnote model-io parsed-result --request-id <request_id> --platform mobile --out ./parsed-result.json
patchxnote model-io packaged-result --request-id <request_id> --platform mobile --out ./packaged-result.json
patchxnote model-io export --request-id <request_id> --platform mobile --out ./model-io.json
```

- [ ] Include webhook commands:

```sh
patchxnote webhook set "产品群 飞书" --type feishu --url-stdin
patchxnote webhook list
patchxnote webhook test "产品群 飞书"
patchxnote webhook draft --memory-id <memory_id> --platform mobile --out ./patchxnote-drafts/example
patchxnote webhook send --target "产品群 飞书" --file ./message.md
patchxnote webhook send --target "产品群 飞书" --draft ./patchxnote-drafts/example
patchxnote webhook remove "产品群 飞书"
```

- [ ] Keep `--url-stdin` in public examples, not raw webhook URLs.
- [ ] Do not include real request IDs from user data.
- [ ] Do not include real webhook URLs.
- [ ] Prefer `--out` examples for source text and AI result fields so README does not encourage dumping sensitive content into chat or terminal history.
- [ ] Mention that `request_id` is obtained from `patchxnote model-io list --platform mobile|desktop`.
- [ ] Mention that `memory_id` comes from record-list or record-rendering workflows, while `request_id` comes from AI整理记录 workflows.

**Verification:**

```sh
grep -R "open-apis/bot/v2/hook\\|mrun_\\|17795780915\\|access_token\\|refresh_token" README.md README.zh-CN.md packages/npm/README.md docs
```

Expected: no real data.

## Task 6: Update Security And Limitations

**Files:**
- Modify: `README.md`
- Modify: `README.zh-CN.md`
- Modify: `docs/readme-benchmark.zh-CN.md`

**Checklist:**

- [ ] State what remains unavailable:
  - [ ] raw audio
  - [ ] audio download
  - [ ] hardware bind/release/recovery
  - [ ] payment and purchase flows
  - [ ] Admin API
  - [ ] automatic model execution/replay
  - [ ] background automatic webhook pushes
- [ ] State what is now available only through explicit calls:
  - [ ] source text
  - [ ] AI response
  - [ ] parsed result
  - [ ] final structured result
- [ ] State that large/sensitive fields should be written to local files with `--out`.
- [ ] State that webhook sends surface provider errors directly.
- [ ] State that webhook sends do not follow redirects.
- [ ] State that webhook URLs and signing secrets are stored locally and not listed back.
- [ ] Replace any wording that says "does not read complete transcripts" with precise wording:

```text
Agent 不读取原始音频，也不提供音频下载。可查看的原文文本和 AI 结果必须由用户或 AI 明确调用工具后读取，且建议导出到本地文件。
```

**Verification:**

```sh
grep -nE "完整转写|full transcript|raw transcript" README.md README.zh-CN.md docs/readme-benchmark.zh-CN.md
```

Expected: only precise boundary wording remains.

## Task 7: Regenerate README Visual Assets

**Files:**
- Replace: `docs/assets/patchxnote-agent-cover.png`
- Replace: `docs/assets/patchxnote-agent-quickstart.png`
- Replace: `docs/assets/patchxnote-agent-architecture.png`
- Replace: `docs/assets/patchxnote-agent-tools.png`
- Replace: `docs/assets/patchxnote-agent-safety-boundary.png`
- Replace: `docs/assets/patchxnote-agent-feishu-cover.png`
- Replace: `docs/assets/patchxnote-agent-social-preview.png`

**Visual Style Requirements:**

- [ ] Match existing style:
  - [ ] light blue/white background;
  - [ ] large black title;
  - [ ] blue accent text/icons;
  - [ ] white rounded cards;
  - [ ] soft shadows;
  - [ ] MR20 recorder product visual or similar product signal;
  - [ ] clean Chinese typography.
- [ ] Keep text short enough to be readable at README width.
- [ ] Do not use screenshots of real user data.
- [ ] Do not include real phone numbers, OTPs, tokens, request IDs, webhook URLs, or provider payloads.
- [ ] Preserve PNG filenames so README links do not change.
- [ ] Keep dimensions compatible with current README images:
  - [ ] wide cover/social images should remain wide banner format;
  - [ ] tools/architecture/quickstart/safety should remain readable in GitHub README width;
  - [ ] target each PNG under roughly 2 MB unless quality requires otherwise.
- [ ] If using AI image generation, inspect all Chinese/English text carefully. Regenerate or switch to programmatic layout if text is misspelled, warped, clipped, or inconsistent with README facts.
- [ ] If using a programmatic layout, use committed or system-available fonts that render Chinese correctly; do not rely on missing local fonts.
- [ ] Add or update an asset maintenance note in `docs/readme-benchmark.zh-CN.md` with each image's intended text and facts.

**Required Asset Facts:**

### `patchxnote-agent-tools.png`

- [ ] Title: `PatchXNote Agent 工具能力`
- [ ] Subtitle: `19 个 MCP 工具，帮 AI 查看记录、整理结果和发送 Webhook`
- [ ] Show three groups:
  - [ ] `账号和记录查询 7`
  - [ ] `Webhook 配置和发送 7`
  - [ ] `AI 整理结果查看 5`
- [ ] Bottom badge: `本地运行 · 明确调用 · 用户可控`
- [ ] Do not say "7 个只读 MCP 工具".

### `patchxnote-agent-safety-boundary.png`

- [ ] Title: `PatchXNote Agent 安全边界`
- [ ] Left column title: `可以明确查看`
- [ ] Left items:
  - [ ] `账号与额度`
  - [ ] `记录列表`
  - [ ] `AI 整理结果`
  - [ ] `Webhook 别名`
- [ ] Right column title: `不会自动操作`
- [ ] Right items:
  - [ ] `原始音频`
  - [ ] `硬件绑定/解绑`
  - [ ] `支付/Admin`
  - [ ] `自动发送`
- [ ] Bottom note: `原文文本和 AI 结果需显式调用，建议导出到本地文件`

### `patchxnote-agent-architecture.png`

- [ ] Show flow:

```text
AI 助手 -> 本地 MCP -> patchxnote CLI -> Agent API
                                  -> 本地 Webhook 配置/发送
```

- [ ] Show right-side destinations:
  - [ ] `PatchXNote 服务端只读接口`
  - [ ] `本地安全存储`
  - [ ] `飞书/钉钉/Webhook`
- [ ] Bottom badges:
  - [ ] `本地运行`
  - [ ] `服务端数据只读`
  - [ ] `Webhook 手动发送`

### `patchxnote-agent-quickstart.png`

- [ ] Title: `三步接入 PatchXNote Agent`
- [ ] Steps:
  - [ ] `1 安装 CLI`
  - [ ] `2 验证码登录`
  - [ ] `3 连接 AI 助手`
- [ ] Bottom: `开始查看记录、AI 整理结果，并手动发送 Webhook`

### `patchxnote-agent-cover.png`

- [ ] Title: `PatchXNote Agent`
- [ ] Subtitle: `让 AI 帮你查看记录、整理结果、发送 Webhook`
- [ ] Cards:
  - [ ] `安装 CLI`
  - [ ] `连接 AI 助手`
  - [ ] `查看记录`
  - [ ] `手动发送`

### `patchxnote-agent-feishu-cover.png`

- [ ] Title: `PatchXNote Agent`
- [ ] Subtitle: `把 PatchXNote 记录整理成 Markdown，发送到飞书/钉钉`
- [ ] Cards:
  - [ ] `查看记录`
  - [ ] `生成草稿`
  - [ ] `确认后发送`

### `patchxnote-agent-social-preview.png`

- [ ] Title: `PatchXNote Agent`
- [ ] Subtitle: `Local MCP Bridge for PatchXNote`
- [ ] Detail line: `Records · AI Results · Webhook Delivery`

**Verification:**

- [ ] Inspect every generated image visually.
- [ ] Check text is not cropped at README display size.
- [ ] Check all README image paths resolve:

```sh
python3 - <<'PY'
from pathlib import Path
import re
for md in [Path("README.md"), Path("README.zh-CN.md")]:
    text = md.read_text(encoding="utf-8")
    missing = []
    for url in re.findall(r'!\[[^\]]*\]\(([^)]+)\)', text):
        if "://" in url or url.startswith("#"):
            continue
        if not (md.parent / url).exists():
            missing.append(url)
    if missing:
        raise SystemExit(f"{md}: missing images {missing}")
print("image links ok")
PY
```

- [ ] Check no stale phrase appears:

```text
7 个只读
18 个
只读访问
memory metadata
完整转写不可读
```

## Task 8: Update npm README

**Files:**
- Modify: `packages/npm/README.md`

**Checklist:**

- [ ] Keep npm README short.
- [ ] Update product sentence to mention records, AI results, Markdown, and webhook.
- [ ] Update examples to `0.2.4` where pinned.
- [ ] Mention MCP exposes 19 tools.
- [ ] Mention local webhook tools and explicit model result inspection.
- [ ] Keep security note:
  - [ ] credentials in OS-native keychain;
  - [ ] webhook secrets not listed back;
  - [ ] no raw audio, hardware write, payment, or Admin API.

**Verification:**

```sh
node packages/npm/test/install.test.js
npm --prefix packages/npm pack --dry-run
```

Expected: installer tests pass and npm pack includes `bin/patchxnote-agent.js`, `package.json`, and package `README.md`.

If Windows npm has trouble with the WSL UNC path, copy `packages/npm` to a Windows temp directory and run `npm pack --dry-run` there, following the existing release runbook guidance.

## Task 9: Update Maintenance Docs And Release Notes

**Files:**
- Modify: `docs/readme-benchmark.zh-CN.md`
- Modify: `docs/release-and-maintenance-runbook.zh-CN.md`
- Modify or create: release notes section in `README.md` / `README.zh-CN.md` if desired
- Modify or create: `docs/plans/2026-08-06-agent-v1-mvp.md` only if release status tracking requires it
- Prepare external copy: Feishu public guide update notes, not necessarily committed unless a repo doc stores them

**Checklist:**

- [ ] Update README maintenance standard:
  - [ ] 19 MCP tools;
  - [ ] plain user language;
  - [ ] new asset fact requirements;
  - [ ] precise source text/model result boundary.
- [ ] Add `0.2.4` release summary:

```text
0.2.4 adds local webhook MCP tools, model IO field inspection, model IO trace discovery, and refreshed README/assets.
```

- [ ] Confirm release runbook still references current MCP tool count `19`.
- [ ] Add a note that image assets must be regenerated when product facts change.
- [ ] Add or update a short `0.2.4` release summary suitable for GitHub Release notes:

```text
Highlights:
- 19 MCP tools grouped around records, webhook delivery, and AI result inspection.
- Webhook targets can be configured by alias and manually sent to Feishu, DingTalk, or generic webhooks.
- AI整理记录 can be listed so users can inspect source text, AI response, parsed result, and final result by request_id.
- README and visual assets refreshed for the new product positioning.

Safety:
- No automatic webhook sends.
- No raw audio download.
- PatchXNote server data access remains read-only.
```

- [ ] Prepare a Feishu guide update checklist:
  - [ ] replace old cover if used;
  - [ ] update install command to `0.2.4`;
  - [ ] update MCP tool count to 19;
  - [ ] add webhook and AI整理结果 examples;
  - [ ] keep OTP/token warning.

**Verification:**

```sh
grep -R "seven V1\\|七个\\|7 个只读\\|18 个" docs README.md README.zh-CN.md packages/npm/README.md
```

Expected: no stale count.

## Task 10: Automated Validation

**Commands:**

```sh
git diff --check
go test ./...
scripts/e2e/mvp-smoke.sh
node packages/npm/test/install.test.js
npm --prefix packages/npm pack --dry-run
```

**Expected:**

- [ ] `git diff --check` PASS.
- [ ] `go test ./...` PASS.
- [ ] `scripts/e2e/mvp-smoke.sh` PASS and still reports 19 MCP tools.
- [ ] npm installer tests PASS.
- [ ] npm pack dry-run PASS and shows the expected files.

**Extra checks:**

```sh
grep -R "0.2.3" README.md README.zh-CN.md packages/npm/README.md packages/npm/package.json
grep -R "7 个只读\\|18 个\\|seven V1\\|memory metadata" README.md README.zh-CN.md packages/npm/README.md docs/readme-benchmark.zh-CN.md docs/release-and-maintenance-runbook.zh-CN.md
grep -R "17795780915\\|open-apis/bot/v2/hook\\|mrun_" README.md README.zh-CN.md packages/npm/README.md docs/readme-benchmark.zh-CN.md docs/release-and-maintenance-runbook.zh-CN.md
grep -R "access_token\\|refresh_token\\|sk_" README.md README.zh-CN.md packages/npm/README.md docs/readme-benchmark.zh-CN.md docs/release-and-maintenance-runbook.zh-CN.md
```

Expected:

- [ ] `0.2.3` remains only in changelog/history contexts if intentionally retained.
- [ ] No stale count or stale positioning.
- [ ] No real sensitive values. Generic warning terms such as `access_token` and `refresh_token` may appear only in security guidance.

## Task 11: Real Local Install Smoke For 0.2.4 Candidate

**Checklist:**

- [ ] Build local Windows binary with `0.2.4-local` or release candidate metadata.
- [ ] Install through npm wrapper with `--from-local`.
- [ ] Verify:

```powershell
patchxnote version
patchxnote auth status --output json
patchxnote model-io list --platform mobile --limit 1 --output json
patchxnote webhook list
```

- [ ] Verify MCP:
  - [ ] initialize reports non-dev version;
  - [ ] `tools/list` returns 19;
  - [ ] `patchxnote_list_model_io_traces` works.
- [ ] If `model-io list` is empty for the current account/platform, retry the other platform and record the empty-state behavior as PASS only if the command returns a valid empty page.
- [ ] Verify webhook with a test target only if user provides/approves test webhook URL for this run.
- [ ] Do not paste source text, provider payloads, tokens, or webhook URLs into final evidence.

## Task 12: Commit And Push

**Checklist:**

- [ ] Review staged diff.
- [ ] Confirm docs and images are the only expected changed areas plus package version metadata.
- [ ] Commit:

```sh
git add README.md README.zh-CN.md packages/npm/README.md packages/npm/package.json docs/assets docs/readme-benchmark.zh-CN.md docs/release-and-maintenance-runbook.zh-CN.md docs/plans/2026-08-14-readme-assets-0.2.4-release-checklist.md
git commit -m "docs: refresh README for 0.2.4"
git push origin main
```

- [ ] If code or package metadata changes beyond docs/assets/version are needed, adjust commit message to match actual scope.
- [ ] If generated asset files are large, review `git diff --stat` and consider lossless compression before committing.

## Task 13: GitHub Release And npm Publish

**Preconditions:**

- [ ] Working tree clean.
- [ ] `main` pushed.
- [ ] Version is `0.2.4`.
- [ ] All validation from Task 10 passed.
- [ ] Real local smoke from Task 11 passed.

**Checklist:**

- [ ] Follow `docs/release-and-maintenance-runbook.zh-CN.md`.
- [ ] Create and push tag `v0.2.4`.
- [ ] Run GitHub release workflow for `v0.2.4`.
- [ ] Wait for checks and release artifacts.
- [ ] Verify checksums and release assets.
- [ ] Download or install the released binary and verify:

```sh
patchxnote version
```

Expected: `patchxnote 0.2.4`.

- [ ] Run npm publish workflow for `0.2.4`.
- [ ] Confirm npm latest:

```sh
npm view patchxnote-agent version
```

Expected: `0.2.4`.

- [ ] Fresh install smoke:

```sh
npx -y patchxnote-agent@0.2.4 install --print-config
patchxnote version
```

- [ ] Confirm MCP initialize version is `0.2.4`.
- [ ] Confirm `patchxnote mcp serve` `tools/list` returns 19 from the freshly installed package.
- [ ] Confirm README links on GitHub render all updated images.
- [ ] If the Feishu public guide is updated externally, record the update status; if not, report it as pending instead of implying it is done.
- [ ] Record release evidence without secrets.

## Edge Cases To Review Before Implementation

- [ ] README should not claim every MCP host can auto-install or auto-configure PatchXNote Agent.
- [ ] README should not imply webhook sends are automatic or scheduled.
- [ ] README should not imply background sync from PatchXNote to webhook exists.
- [ ] README should not imply AI can execute new model summaries.
- [ ] README should not imply Agent can read raw audio.
- [ ] README should not imply Agent can download audio files.
- [ ] README should not imply source text/model result access is harmless; it is explicit sensitive content access.
- [ ] README should not imply ordinary record search and AI整理记录 search are the same data source.
- [ ] README should not use real request IDs from local tests.
- [ ] README should not include real webhook target names if they identify private groups.
- [ ] Generated images must remain readable in GitHub README width and on mobile.
- [ ] Generated images must not include tiny dense tool names that blur in README display.
- [ ] Generated images must not include text produced incorrectly by image generation.
- [ ] English and Chinese README must stay factually aligned.
- [ ] `packages/npm/README.md` should stay shorter than root README.
- [ ] README should not say or imply the repository is open-source while package license is `UNLICENSED`.
- [ ] README should distinguish `patchxnote-agent` npm wrapper from the installed `patchxnote` binary.
- [ ] README should keep "default public beta API" wording clear so users do not mistake test API for production SLA.
- [ ] If `0.2.4` release fails after README is updated, either complete the release or revert user-facing latest-version claims before stopping.

## Definition Of Done

- [ ] README and README.zh-CN explain the product in plain user language.
- [ ] README, npm README, and maintenance docs consistently say 19 MCP tools.
- [ ] All stale "7 read-only tools" and "18 tools" references are removed or replaced.
- [ ] All stale "transcripts unavailable" wording is replaced by precise raw-audio/source-text boundaries.
- [ ] All referenced README images are regenerated and visually verified.
- [ ] Version is bumped to `0.2.4`.
- [ ] Automated tests pass.
- [ ] npm pack dry-run passes.
- [ ] Real local candidate install smoke passes.
- [ ] `main` is committed and pushed.
- [ ] GitHub Release and npm publish complete.
- [ ] npm latest is `0.2.4`.
- [ ] Fresh install from npm `0.2.4` verifies binary version and 19 MCP tools.
- [ ] External Feishu public guide status is either updated or explicitly reported as pending.
- [ ] Final user report lists changed docs/assets, validation commands, release status, and any remaining risk.
