# Official MCP Registry Publishing Notes

更新日期：2026-09-03

PatchXNote Agent 的 MCP Registry 名称草案：

```text
io.github.zsts119/patchxnote-agent
```

这个名称必须和 `packages/npm/package.json` 的 `mcpName` 以及根目录 `server.json` 的 `name` 一致。

## 当前文件

- `server.json`
- `packages/npm/package.json` 的 `mcpName`

## 发布前检查

```sh
npm view patchxnote-agent version dist-tags.latest repository.url --registry https://registry.npmjs.org
mcp-publisher validate
```

Registry 只托管 MCP server metadata，不托管 npm 包本身。因此必须先确认对应 npm 版本已经发布并可访问，再发布 registry metadata。

## Auth 边界

`mcp-publisher login github` 可能要求在浏览器或 GitHub device flow 中完成授权。授权码只用于用户自己的终端/浏览器流程，不应贴进 AI 聊天、文档或 evidence。

## 不做的事

- 不在 `server.json` 写入 access token、refresh token、webhook secret 或 API key。
- 不声明环境变量，除非后续有真实 public flow 需要，并且 schema 标记 secret。
- 不把 registry publish 说成具体 AI 客户端已验收。
