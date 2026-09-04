# Claude Code Plugin Marketplace Notes

更新日期：2026-09-03

Claude Code 插件包位置：

```text
packages/plugins/claude/patchxnote-agent/
```

repo marketplace 文件：

```text
.claude-plugin/marketplace.json
```

## 本地安装验证

从仓库根目录启动 Claude Code 后：

```text
/plugin marketplace add .
/plugin install patchxnote-agent@patchxnote-agent
```

如果安装摘要提示需要 reload，执行：

```text
/reload-plugins
```

插件 skill 应按 Claude Code 的插件命名空间调用：

```text
/patchxnote-agent:patchxnote-mcp
```

## URL Marketplace 边界

本地 marketplace 可以使用相对路径：

```json
{
  "name": "patchxnote-agent",
  "source": "./packages/plugins/claude/patchxnote-agent"
}
```

URL-based marketplace 不能只托管 `marketplace.json` 后继续引用没有一起被服务的相对 plugin 文件。公开 URL 安装前要改成 Git source、仓库根目录、或确保 plugin 目录也能被同一 marketplace 根访问。

## 版本更新

Claude Code 插件带 `version` 时，发布更新必须 bump `packages/plugins/claude/patchxnote-agent/.claude-plugin/plugin.json` 的版本。同步 skill 内容后也要重新验证安装、reload 和 fresh-session activation。

## 不做的事

- 首版不加 hooks、commands、agents、LSP 或 top-level `bin/`。
- 首版不承诺 MCPB/Desktop Extension。
- 不让插件文件引用插件目录之外的相对路径。
