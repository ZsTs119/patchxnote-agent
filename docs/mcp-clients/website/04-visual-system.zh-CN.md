# PatchXNote MCP 官网视觉系统与 UI 规范

**日期：** 2026-08-28

**定位：** 本文用于指导 PatchXNote MCP 官网、客户端详情页、OAuth 授权页和后续生成视觉素材。本文只定义视觉和交互规范，不代表页面已经实现或客户端已经完成真实验收。

## 设计方向

官网视觉方向定为：

```text
Black Chrome Command Hub
```

这不是普通深色文档站，也不是硬件电商页。它应表达三件事：

1. **真实产品可信度**：PatchXNote 是有实体 AI 录音卡、App 和服务端授权边界的产品。
2. **AI 工具接入感**：官网核心任务是让用户把 PatchXNote 接入 Cursor、VS Code、Codex、Claude、飞书、腾讯等 AI 工具。
3. **安全和控制感**：One login、secret-free config、OS-native secure storage、OAuth connector session 必须被视觉化。

用户提供的产品图、介绍页和 Logo 图只作为事实参考，不直接作为官网最终视觉素材。后续素材应基于这些事实重新生成或处理成统一的高级黑风格。

## 产品信息提炼

官网可使用的产品事实：

- 产品名：`PatchXNote`。面向消费者文案可出现 `Patchx Note`，但 MCP 官网导航和技术入口优先统一为 `PatchXNote MCP`。
- 产品类型：AI 录音卡、本地优先记录工具、AI 总结和记忆入口。
- 主张：隐私归你，订阅归零。
- 价值表达：不只是记录声音，而是把每一次记录沉淀为自己的数字资产。
- 可用卖点：本地优先、无需订阅、自由导出、自动整理、Type-C / OTG、多语言识别、按下即录、佩戴方式灵活。

MCP 官网不主推价格、配件和完整硬件参数。硬件信息只用于建立可信度和产品上下文，避免把页面重心从“AI 工具接入”带回硬件售卖。

## 色彩系统

整体为高级黑，但不能做成单调黑灰或深蓝页面。色彩应像金属产品、终端控制台和安全连接状态的组合。主视觉不使用高饱和蓝绿大色块，CTA、状态条、tab、卡片 hover 都优先使用黑铬、钛银、冷银和白色高光。

### 基础色

| Token | 用途 | 建议值 |
| --- | --- | --- |
| `--px-bg` | 页面底色 | `#07090A` |
| `--px-bg-elevated` | 顶栏、安装区背景 | `#0D1012` |
| `--px-surface` | 卡片和面板 | `#121619` |
| `--px-surface-2` | 代码块和深层面板 | `#090C0E` |
| `--px-border` | 默认边框 | `rgba(220, 230, 235, 0.13)` |
| `--px-border-strong` | hover / active 边框 | `rgba(225, 231, 234, 0.46)` |
| `--px-text` | 主文字 | `#F4F7F8` |
| `--px-text-muted` | 次级文字 | `#A5B0B5` |
| `--px-text-soft` | 辅助文字 | `#6F7B82` |

### 品牌与状态色

| Token | 用途 | 建议值 |
| --- | --- | --- |
| `--px-metal` | 主按钮文字、金属高光、关键分隔 | `#D8DEE1` |
| `--px-metal-dim` | 按钮底色高光、弱金属边 | `#7D858A` |
| `--px-accent` | 关键交互高光、复制完成、授权完成的冷银提示 | `#E5EAEC` |
| `--px-accent-2` | 远程 MCP 细线、代码光标、稀有焦点光 | `#BFC7CB` |
| `--px-accent-warm` | 产品钛银高光 | `#C8BFAF` |
| `--px-warning` | Pending / manual 状态 | `#C8BFAF` |
| `--px-danger` | 错误和失效授权 | `#B86A62` |
| `--px-success` | 已复制、已连接、已验证 | `#E5EAEC` |

色彩比例：

- 黑 / 石墨 / 炭灰：约 `82%` 到 `88%`。
- 钛银 / 冷白文字 / 金属边：约 `8%` 到 `12%`。
- 冷银交互高光：不超过 `3%`，用于按钮边缘、复制完成、授权完成等明确结果态。
- 暖钛色：不超过 `2%`，只用于产品高光和少量 manual / pending 状态。
- 低饱和错误暖色：不超过 `1%`，只用于授权失败和不可恢复错误。

禁止项：

- 禁止大面积蓝色、绿色、青绿色实心按钮。
- 禁止把主 CTA 做成亮蓝渐变或荧光绿渐变。
- 禁止在 Hero 状态条使用绿色圆点作为默认视觉符号。
- 禁止整页读成深蓝、蓝紫、赛博朋克霓虹或游戏 UI。
- 禁止装饰性光球、bokeh、彩虹渐变和无语义霓虹边框。

## 字体与排版

字体栈：

```css
font-family: Inter, ui-sans-serif, -apple-system, BlinkMacSystemFont, "Segoe UI",
  "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
```

中文页面可以优先使用系统中文字体。英文标题和代码区保持开发者工具感。

排版建议：

| 层级 | 桌面尺寸 | 移动尺寸 | 用途 |
| --- | --- | --- | --- |
| Display | `64px / 1.02` | `40px / 1.08` | Hero 主标题 |
| H1 | `48px / 1.08` | `34px / 1.12` | 详情页标题 |
| H2 | `32px / 1.16` | `26px / 1.2` | Section 标题 |
| H3 | `20px / 1.3` | `18px / 1.35` | 卡片和面板标题 |
| Body | `16px / 1.65` | `15px / 1.65` | 正文 |
| Caption | `13px / 1.5` | `12px / 1.45` | 状态、标签、辅助信息 |
| Code | `14px / 1.7` | `13px / 1.65` | 命令和配置 |

规则：

- 字距使用 `0`，不要负字距。
- Hero 大字只用于第一屏和详情页顶部，不用于卡片内部。
- 中文标题要短，英文标题可用于技术感主视觉。
- 按钮、标签和卡片内文字必须有固定或响应式约束，不能因状态切换导致布局跳动。

## 布局系统

页面采用长页滚动，每个 section 近似一屏，但不要强行 fullpage scroll。关键决策区使用 Tab 或 segmented control。

建议尺寸：

- 最大内容宽度：`1180px`。
- Hero 最大宽度：`1280px`。
- 桌面左右安全边距：`32px` 到 `48px`。
- 移动端边距：`18px` 到 `22px`。
- Section 垂直间距：桌面 `96px` 到 `128px`，移动端 `64px` 到 `80px`。
- 卡片圆角：`8px`。
- 按钮圆角：`8px` 或胶囊形仅用于小型状态 pill。

禁止：

- 不做卡片套卡片。
- 不把整段页面 section 做成大浮卡。
- 不做纯文档目录首页。
- 不把产品图做成暗、糊、裁切过狠、看不清主体的背景。

## 页面结构

### Header

桌面：

```text
PATCHX
MCP
Clients
Security
Docs
Download App
[Get started]
```

移动端：

```text
PATCHX
[Menu]
```

Header 使用半透明黑色毛玻璃，但文字和按钮必须有足够对比。`Get started` 滚动到客户端选择区，不在顶部新做官网登录体系。

### Hero

目标：一眼知道这是 PatchXNote 官方 MCP 入口。

展示信息：

```text
PatchXNote MCP
隐私归你，AI 由你掌控
把你的真实对话，安全接入 Cursor、VS Code、Codex 和更多 AI 工具。
```

主按钮：

```text
连接我的 AI 工具
查看支持的编辑器
```

视觉：

- 中心使用生成/重制后的银黑 PatchXNote 录音卡，不直接使用浅色介绍图。
- 第一屏不展示完整客户端卡片网格，不展示四个以上功能点，不展示安装命令。
- 产品周围只允许出现极轻量的 AI 工具文字或单色银灰符号。
- 连接线使用冷银细线，轻微流动表达 MCP connection。
- Hero 底部只露出下一屏提示：`Cursor / VS Code / Codex / Claude Code / 更多编辑器`。
- 不使用 `更多客户端持续接入中` 文案，避免用户误解为“还没接好”。

### Pick Your AI Tool

这是官网第二屏的转化核心，布局参考 1Server 的客户端卡片网格，但视觉更精致。

使用 segmented control：

```text
Editors | Local MCP / CLI | Cloud Platforms
```

默认展示 P0 和 P0.5。P1 客户端放在 `More clients` 折叠区或次级网格。

卡片字段：

- 客户端名
- 类型：Editor / Local MCP / Cloud Platform
- 状态：One-click candidate / Setup command / Manual / Remote MCP / Pending
- 一句话说明
- 主动作按钮
- 右上角箭头或冷银刻线

交互：

- hover 时卡片边框出现冷银/钛银细光。
- 鼠标经过卡片时，Hero 或背景连接线可以用冷银细线短暂指向该客户端。
- 点击卡片进入详情页或更新下方 `60-second setup` 预览。

### 60-Second Setup

详情页和首页预览都复用这个安装模块。

安装方式使用 Tab：

```text
AI assisted | One-click | Manual config
```

默认选 `AI assisted`。每个 Tab 的主动作：

- `AI assisted`：复制一段给 AI 的安装提示词。
- `One-click`：仅在 deeplink / 官方网页安装真实验收后显示为主按钮。
- `Manual config`：复制无密钥 JSON/TOML。

本地客户端流程展示：

```text
Send prompt to AI
Run setup command
Confirm browser/login prompts
Verify tools
```

云平台流程展示：

```text
Copy remote MCP URL
Add it in platform console
Authorize PatchXNote
Verify one safe read tool
```

### Three Ways To Connect

用三张并列面板解释入口类型：

| 面板 | 说明 | 主视觉 |
| --- | --- | --- |
| Editor install | 打开编辑器并预填 MCP 配置，用户确认保存 | 编辑器窗口和 MCP 端口 |
| Local MCP / CLI | 复制一句话给本地 AI，AI 执行 setup 命令 | 终端、命令、授权弹窗 |
| Cloud platform | 飞书、豆包、腾讯 Agent 走 remote MCP + OAuth | 云端网关、OAuth、connector session |

这屏用于解释能力边界，但文案要短，不做长文档。

### What Your AI Can Use

展示 MCP 接入后的能力，不展示敏感内容：

- Recent summaries
- Recorder cards
- Quota summary
- Model usage
- Structured results
- Model IO traces
- Webhook templates and sends where authorized

视觉用深色 mock 面板，内容必须是脱敏假数据。不要放真实手机号、客户名、原始转写、完整模型输入输出或 provider payload。

### Security Model

这一屏必须存在，并且要做成高级黑里的“安全控制台”。

核心信息：

```text
No tokens in config.
No phone numbers in examples.
Credentials stay in OS-native secure storage.
Cloud platforms use revocable OAuth connector sessions.
PatchXNote returns bounded, authorized content only.
```

视觉节点：

```text
MCP Config -> npx launcher -> OS Keychain -> GoServer OAuth -> PatchXNote tools
Cloud Platform -> Remote MCP Gateway -> OAuth Connector Session -> PatchXNote tools
```

### Cloud Platform Status

展示平台型客户端的真实进度，不夸大：

| 字段 | 示例 |
| --- | --- |
| Platform | Feishu Aily / Doubao Work Partner |
| Transport | Streamable HTTP / SSE / unknown |
| Auth | OAuth / pending |
| Remote URL | `https://ws-lab.patch-x.cn/patchnote-test-api/mcp` |
| Acceptance | Pending / Accepted |
| Next step | Platform console validation |

### Footer

Footer 信息：

```text
PATCHXNOTE
AI 录音卡 · MCP for local editors and cloud agents

Product
Download App
PatchXNote MCP
Security
Privacy Policy

Developers
npm package
MCP config
Client setup
Release evidence

Platforms
Editors
Local MCP / CLI
Cloud Platforms

Contact
Support
Business cooperation
```

## 组件规范

### Buttons

主按钮：

- 高度：`44px` 到 `48px`。
- 圆角：`8px`。
- 背景：黑铬 / 石墨金属面，允许非常弱的钛银纵向高光；不能使用蓝绿实心或高饱和渐变。
- 边框：默认为冷银半透明；hover 时边缘出现一条很窄的钛银光，不改变按钮主体颜色。
- hover：轻微上浮、右侧箭头滑动、金属边扫光。
- active：压下 `1px`，保留焦点环。
- loading：按钮文案变为 `Opening...` / `Copying...`，宽度保持稳定。

按钮文案：

| 场景 | 文案 |
| --- | --- |
| deeplink 已验收 | `Install in Cursor` |
| AI 帮装 | `Copy AI setup prompt` |
| 命令兜底 | `Copy setup command` |
| 配置兜底 | `Copy manual config` |
| 云平台 | `Copy remote MCP URL` |
| 安全区 | `View security model` |

### Tabs / Segmented Controls

用于两处：

1. 首页客户端分类：`Editors / Local MCP / Cloud Platforms`。
2. 安装方式切换：`AI assisted / One-click / Manual config`。

Tab 样式像控制台模式切换，不使用重圆角大胶囊。active 状态使用细下划线、边框高亮或小型状态灯。

### Client Cards

卡片尺寸稳定，桌面 3 列，平板 2 列，移动 1 列。

字段顺序：

```text
图标 / 首字母
客户端名
类型 + 状态标签
一句话说明
主动作
```

卡片 hover：

- `transform: translateY(-2px)`。
- 边框从默认灰变为冷银/钛银半透明。
- 右上角箭头移动 `2px` 到 `4px`。
- 背景出现轻微金属高光，但不要使用光球。

### Code Blocks

代码块像终端面板：

- 顶栏显示 `command` / `mcp.json` / `config.toml`。
- 右侧复制按钮使用复制图标。
- 成功后复制按钮变为 `Copied` 或勾选图标。
- 不使用包含密钥的示例。

### Status Pills

状态标签：

| 状态 | 文案 | 色彩 |
| --- | --- | --- |
| 一键候选 | `One-click candidate` | 冷银边框 |
| 设置命令 | `Setup command` | 冷银 |
| 手动配置 | `Manual` | 暖钛色细线 |
| 远程 MCP | `Remote MCP` | 冷银细线 |
| 待验收 | `Pending acceptance` | 暖钛色细线 |
| 已验收 | `Accepted` | 冷银小图标或短暂高光 |

Hero 左上角状态条不是卖点卡片，而是系统状态锚点。建议文案：

```text
PATCHXNOTE STATUS
Local-first recorder · MCP ready
```

状态条样式：

- 高度 `32px` 到 `36px`，黑铬半透明底，不使用绿色或蓝色实心背景。
- 不使用绿色圆点。左侧改为冷银短刻线、金属凹槽、细小波形或机械刻度。
- 状态变化使用非常短的钛银脉冲线，默认静止。
- 边框使用冷银半透明，hover 时只出现一条非常细的钛银扫光。
- 不把状态条做成营销 badge，不堆多段彩色标签。

## 动效与交互

动效要让页面有高级感，但不能妨碍安装。

动效原则：

- 特效表现“材质和连接”，不是表现“霓虹和炫彩”。
- 所有大面积 UI 面都保持黑、石墨、钛银；结果态和远程 MCP 细节也优先使用冷银或暖钛色，不引入高饱和蓝绿。
- 动效触发应与用户动作或 MCP 状态相关，避免无意义背景乱动。
- 首屏可以有视觉惊喜，但安装区必须稳定、可读、可信。

建议：

- Hero 连接线缓慢流动，周期 `6s` 到 `10s`。
- 客户端卡片 hover 时触发短连接高亮。
- 主按钮 hover 时做一条窄扫光，时长 `260ms` 到 `360ms`。
- 复制成功时，代码块边框亮起 `700ms`。
- Tab 切换使用淡入和轻微位移，不使用剧烈弹跳。
- 授权页状态切换使用稳定的进度灯：`Ready / Sending code / Authorizing / Connected / Failed`。

### Micro-interaction Matrix

| 区域 | 默认状态 | hover / active | 上线要求 |
| --- | --- | --- | --- |
| Hero 产品图 | 轻微悬浮、底座弱光环 | 鼠标移动时 `2deg` 内 3D 视差 | 不能遮挡文字，移动端关闭视差 |
| 状态条 | 冷银刻线或金属凹槽 | 钛银脉冲线扫过一次 | 禁止绿色圆点 |
| 主 CTA | 黑铬金属按钮 | 边缘钛银扫光，箭头滑动 `2px` | 禁止蓝绿实心按钮 |
| 次 CTA | 深石墨描边按钮 | 边框亮度提高，背景不变色 | 权重低于主 CTA |
| 客户端卡片 | 低亮金属边框 | 上浮 `2px`，角箭头滑动，边缘短光 | 不使用整卡发光 |
| MCP 连接线 | 不常亮或极弱冷银线 | 选中/hover 后短暂流动 | 线宽 `1px` 左右 |
| Tabs | 细下划线 active | active 指示条滑动 | 不使用大胶囊色块 |
| Code block | 黑色终端面 | copy 后边框亮 `700ms` | 示例必须无密钥 |
| 授权成功页 | 单一银色完成符号 | 钛银脉冲环扩散一次 | 不放产品图、三步大图、token/code |

### Visual Effects Budget

每个视口最多同时出现：

- 1 个主材质动效：Hero 产品悬浮或底座光环。
- 1 个流程动效：连接线、状态线或授权环。
- 1 个用户触发动效：按钮 hover、卡片 hover、copy success。

不要同时叠加大面积背景光、产品发光、按钮发光、卡片发光和连接线流动。高级感来自克制的层次，不来自亮度堆叠。

限制：

- 默认不启用复杂全屏 3D。若后续决定加入 WebGL，只能作为 Hero 背景增强，核心安装流程不能依赖它。
- 遵守 `prefers-reduced-motion`，减少或关闭连接流动和扫光。
- 动效不能导致文本、按钮、卡片尺寸抖动。

### Auth Result Pages

授权结果页要克制，尤其是成功页：

- 只展示一个结果：`授权已完成` 或 `授权未完成`。
- 成功页不放产品图，不放三步流程大图，不放客户端卡片。
- 成功页最多一个银色完成符号和两枚按钮。
- 背景保持黑铬材质，可有非常弱的钛银脉冲环。
- 安全提示放成底部小字，不做成大卡片。
- 成功态不使用绿色圆点或绿色大对勾。

## 素材生成规范

用户提供的四张图用于确定产品事实：

- 硬件外观：银灰正面、黑色夹持结构、侧边按钮、正面 `PATCHX` 标识。
- 产品关键词：AI 录音卡、隐私归你、订阅归零、按下即录、本地优先、自动整理。
- 使用方式：夹戴、吊坠、佩戴、磁吸。
- 品牌识别：`PATCHX` 字标和 `X` 方框标识。

需要生成或处理的官网素材：

1. Hero 产品渲染图：银黑录音卡在高级黑空间中，主体清晰，保留真实外形特征。
2. AI 连接背景：PatchXNote 产品节点连接编辑器、本地 MCP 和云平台节点。
3. 深色 App / Summary mock：展示摘要、记忆、结构化结果，不使用真实用户内容。
4. OAuth 授权视觉：安全授权面板，和官网同一视觉语言。
5. 三类入口插图：编辑器安装、本地 CLI、云平台 remote MCP。

生成原则：

- 不直接使用浅蓝白介绍图作为页面主视觉。
- 不凭空改变硬件主要形态、颜色和 Logo 位置。
- 不使用第三方客户端 Logo，除非已确认公开官网使用规范或授权。
- 未确认 Logo 时使用首字母、文字或自有抽象端口图标。
- 素材必须能在深色背景上清楚呈现产品和操作含义。

## 响应式规范

桌面：

- Hero 左侧文案，右侧或中心为产品连接视觉；第一屏只展示核心信息。
- 客户端卡片 3 列。
- 安装模块左右布局：左侧操作，右侧代码/状态面板。

平板：

- Hero 改为上下布局。
- 卡片 2 列。
- 安装模块堆叠，但 Tab 保持一行。

移动：

- Header 收起为菜单。
- Hero 产品视觉在标题下方，不遮挡文字。
- 卡片 1 列。
- 代码块允许横向滚动。
- 主按钮全宽。
- Tab 允许横向滚动，不换成下拉。

移动端必须检查最长按钮文案和中英文混排，不能溢出或遮挡。

## 可访问性与安全文案

- 所有交互按钮必须有可见 focus ring。
- 色彩状态不能只靠颜色区分，必须有文字标签。
- Copy 后要有 `aria-live` 或可见状态反馈。
- 第三方打开 deeplink 前，按钮文案必须明确会打开对应客户端。
- 授权页不展示 token、OAuth code、验证码、手机号原文或请求完整 query。
- 错误信息展示用户可理解的中文文案，不显示 `[object Object]`。
- 安装提示词必须要求 AI 在用户输入验证码、登录、授权确认时暂停。

## 后续实现验收

视觉实现完成后至少检查：

- Chrome 桌面视口：1440px、1280px。
- Chrome 移动视口：390px、430px。
- Header、Hero、客户端卡片、安装 Tab、代码块、Footer 不遮挡、不溢出。
- 主按钮 hover、copy、Tab 切换状态清晰。
- `prefers-reduced-motion` 生效。
- 示例配置不包含 token、手机号、验证码、OAuth code、refresh token、webhook secret、真实用户内容或 provider payload。
- 若使用生成图，确认硬件外形仍像 PatchXNote 录音卡，不把产品变成无关设备。
