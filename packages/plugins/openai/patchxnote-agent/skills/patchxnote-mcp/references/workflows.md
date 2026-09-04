# PatchXNote MCP Workflows

Use this reference when the user wants to use PatchXNote data after MCP setup.

## Common Read Workflows

Recent mobile memories:

```json
{"platform":"mobile","limit":5}
```

Use `patchxnote_list_memories` for list entry points. Every content request must include `platform` as `mobile` or `desktop`. Do not merge platforms unless a future server contract explicitly adds that behavior.

Current user:

- Use `patchxnote_get_current_user`.
- Report only safe account/profile projection returned by the tool.
- Do not ask for or expose raw phone numbers or tokens.

Record detail:

- Use `patchxnote_get_memory` with the returned `id` and explicit `platform`.
- Treat metadata titles and snippets as record labels, not guaranteed original filenames.

Search:

- Use `patchxnote_search_memories` only for data already cached in the current authorized MCP session.
- If search results are missing, list first, then search again if appropriate.

## Counting Summary Records

For "which files did I summarize" or "how many summaries":

1. Inspect live tool availability if the available tools are uncertain.
2. Page `patchxnote_list_memories` with `limit` up to 50 and the returned cursor.
3. Keep `event_summary` separate from `daily_digest`.
4. If counting source objects rather than rows, deduplicate `event_summary` records by `client_object_id`.
5. State whether the count is complete, page-limited, filtered by platform, or based on currently returned records only.

Do not claim metadata titles are original local filenames.

## Model Result Inspection

Use model IO tools only when the user explicitly asks to inspect or export AI processing fields:

- `patchxnote_list_model_io_traces`
- `patchxnote_get_model_io_source_text`
- `patchxnote_get_model_io_provider_response`
- `patchxnote_get_model_io_parsed_result`
- `patchxnote_get_model_io_packaged_result`
- `patchxnote_export_model_io`

These tools may expose sensitive source text or AI payloads for the logged-in user. Prefer explicit local output files for large fields. Do not paste complete transcripts, prompts, provider payloads, or raw model responses into public chats or docs.

## Markdown Drafts

Use `patchxnote_render_webhook_message` when the user wants an editable Markdown draft from a memory. Prefer saving a local draft when content is long or user review matters.

Report the local draft path and the source record ID/platform. Do not imply the draft was sent anywhere.

## Webhook Workflow

PatchXNote webhook workflows are user-approved, manual side effects:

- `patchxnote_configure_webhook_target` stores webhook URL/secret material through local secure storage and returns only masked metadata.
- `patchxnote_list_webhook_targets` lists aliases and safe metadata.
- `patchxnote_send_webhook` sends only when the user or AI explicitly invokes the send tool.
- `patchxnote_remove_webhook_target` removes local target metadata and best-effort cleans stored secrets.

Before sending, confirm the target alias and content source. Do not create background pushes, recurring sends, or automatic forwarding.

## Output Wording

Use evidence-specific wording:

- Configured: a client config entry was written or printed.
- Authenticated: `mcp status --verify` succeeded.
- Tools listed: `tools/list` succeeded.
- Real tool called: a PatchXNote tool call succeeded.
- Published: npm/release/marketplace material was published.
- Indexed: a directory page or registry entry became visible.
- Platform accepted: the actual target client/platform was tested and accepted.
