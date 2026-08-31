# PatchXNote MCP 官网实现与验收 Checklist

**日期：** 2026-08-28

本文记录官网从文档、实现、授权、素材到真实客户端验收的流程。它是进入 GoServer `web/mcp/` 开发后的执行清单。

## A. 文档与事实源

- [x] 官网规划文档集中到 `docs/mcp-clients/website/`。
- [x] 官网视觉事实源 `docs/mcp-clients/website/DESIGN.md` 已新增。
- [x] 四张用户原始参考素材保存到 `docs/mcp-clients/website/assets/reference/raw/`。
- [x] 客户端登记表 `docs/mcp-clients/clients.json` 通过结构校验。
- [x] 客户端官方来源文档已沉淀到 `03-client-install-sources.zh-CN.md`。
- [ ] GoServer 官网快照 `web/mcp/data/clients.json` 由登记表生成或手工同步。
- [ ] 快照字段全部来自登记表、官方来源、真实验收或 GoServer 真实配置。

## B. GoServer 路由与静态资源

- [ ] 新增 GoServer `web/mcp/` 静态目录。
- [ ] 新增 `/mcp` 和 `/mcp/` 首页路由。
- [ ] 新增 `/mcp/clients/{id}` 客户端详情路由。
- [ ] 新增 `/mcp/platforms/{id}` 云平台详情路由。
- [ ] 新增 `/mcp/app.css`、`/mcp/app.js`、`/mcp/data/clients.json`、`/mcp/assets/*` 静态资源路由。
- [ ] 非法路径返回统一 `route_not_found`。
- [ ] content type 覆盖 HTML、CSS、JS、JSON、SVG、PNG、JPG、WEBP。
- [ ] 缓存策略区分 index/data 和版本化 assets。

## C. 官网前端

- [ ] 首屏展示 PatchXNote MCP 的产品定位和主 CTA。
- [ ] 顶部导航覆盖 Clients、Security、Docs、Download App、Get started。
- [ ] 顶部品牌首屏使用 `PATCHX`，不把 `NOTE` 放进主 logo。
- [ ] 第一屏只保留核心标题、副标题、两枚 CTA、产品主视觉和轻量编辑器提示。
- [ ] 完整客户端卡片网格放到第二屏，不塞进首屏。
- [ ] 下一屏入口文案使用 `查看全部支持的编辑器` / `选择你的 AI 工具` / `更多编辑器`，不使用 `更多客户端持续接入中`。
- [ ] 客户端选择区包含 `编辑器`、`云平台`、`本地 MCP` 三类 tab。
- [ ] 切换 tab 时只替换下方卡片列表，不跳转独立一级页面。
- [ ] 客户端卡片支持筛选、状态标签和详情跳转。
- [ ] `找不到你的渠道？` 进入通用 MCP 配置兜底详情页。
- [ ] 详情页包含 `AI assisted`、`One-click`、`Manual config` 三个 setup tab。
- [ ] copy command、copy config、copy AI prompt、copy remote URL 都有成功/失败状态。
- [ ] 未验收一键安装显示 disabled/coming soon，不写成已支持。
- [ ] reduced motion 下关闭主要大动效。
- [ ] 主 CTA 保持黑铬金属质感，不能使用蓝绿实心或高饱和渐变。
- [ ] 页面主色只使用高级黑、黑铬、石墨灰、钛银、冷银和白色高光；不使用蓝绿实心按钮或高饱和状态光。
- [ ] Hero 状态条是系统状态锚点，不使用绿色圆点，不做成普通营销 badge。

## D. 授权登录体验

- [ ] 手机号登录页源文件放入 GoServer `web/mcp/auth/`。
- [ ] 授权确认页源文件放入 GoServer `web/mcp/auth/`。
- [ ] 授权成功页放入 GoServer `web/mcp/auth/`。
- [ ] 授权失败页放入 GoServer `web/mcp/auth/`。
- [ ] 授权成功页只展示成功结果、说明、一主一次按钮和小型安全提示。
- [ ] 授权成功页不放产品图、三步流程大图、客户端卡片、安装命令或配置代码。
- [ ] `/v1/agent/oauth/authorize` 未登录时先返回手机号登录页，登录完成后继续授权确认页。
- [ ] OTP request、OTP verify、redirect_json 流程不回归。
- [ ] 本地 callback 仍接收 OAuth code，并在处理后跳转到官网风格成功/失败页。
- [ ] 成功页不展示 token、code、手机号或验证码。
- [ ] 错误页能清楚展示过期、取消、state mismatch、网络失败等可恢复状态。

## E. 素材与视觉

- [ ] Hero 产品图基于参考素材生成，符合高级黑风格。
- [ ] 安装流程图使用抽象 UI，不泄露真实账号数据。
- [ ] 云平台 remote MCP 图不误导为本地安装。
- [ ] 客户端 logo 上线前确认商标使用规范；未确认时用自有抽象图标。
- [ ] 所有关键图片提供 alt 文案。
- [ ] 图片尺寸压缩到页面所需大小。

## F. 本地客户端验收

- [ ] VS Code：验证 `setup --client vscode`、配置位置、刷新/重启、tools/list、一个只读工具调用。
- [ ] VS Code：验证 `code --add-mcp` 或官方安装入口是否可作为一键路径。
- [ ] Cursor：验证 deeplink 弹窗、配置填充、保存、tools/list、一个只读工具调用。
- [ ] Cursor：验证 `setup --client cursor` fallback。
- [ ] Codex：验证 `codex mcp add patchxnote -- npx -y patchxnote-agent@latest mcp serve`。
- [ ] Claude Code：验证 `claude mcp add` 路径。
- [ ] Claude Desktop：验证配置文件合并和重启后可用；`.mcpb` 暂不作为 V1。
- [ ] Windsurf、Trae、Qoder、WorkBuddy：分别记录官方 schema 和真实可用路径。

## G. 云平台验收

- [ ] Feishu Aily / 豆包工作伙伴：验证 remote MCP 接入表单、授权、initialize/tools/list、一个只读工具调用。
- [ ] Tencent Agent Platform：验证 remote MCP 接入表单或记录平台限制。
- [ ] Enterprise WorkBuddy：区分本地 WorkBuddy 和企业云平台模式。
- [ ] 没有平台租户或权限时，页面只显示待验收状态。

## H. 安全与发布

- [ ] 页面源码和数据 JSON 不包含手机号、验证码、access token、refresh token、OAuth code、webhook secret。
- [ ] 示例 MCP 配置不包含 bearer token。
- [ ] 外链加 `rel="noopener noreferrer"`。
- [ ] Chrome 桌面宽屏、1366 桌面、iPhone 宽度验收通过。
- [ ] 按钮、代码块、toast、tabs、卡片无文字溢出和重叠。
- [ ] `go test ./...` 通过或明确记录已知基线失败。
- [ ] `node docs/mcp-clients/validate-clients.mjs` 通过。
- [ ] `git diff --check` 通过。
