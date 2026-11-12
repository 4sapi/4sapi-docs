---
title: "本地优先的 Agent 工作台：OpenBot 的进程边界与状态归属"
category: 人工智能
tags:
  - Agent 工程
  - Electron 架构
  - 开源项目
description: "OpenBot 把十几个编码 Agent 收进一个桌面工作台，7 千多星的背后是三条硬边界：进程与依赖方向、状态归属、以及可写根。本文拆开它的 Bun + Electron + SolidJS 结构、openbot.db 与三类旁路状态、本地优先与离线的区别、权限三档的收紧方式，以及非商业许可对采纳的影响。"
brand_mode: none
---

# 本地优先的 Agent 工作台：OpenBot 的进程边界与状态归属

把三个编码 Agent 同时跑在一台机器上，问题通常不在模型，而在它们之间的边界：谁的会话算数、文件放在哪、谁能改谁的配置、日志散在几个目录里。

OpenBot（`nightly-labs/openbot`）就是为这件事写的：一个本地优先的工作台，把 Codex、Claude Code 等十几个 CLI 收进同一个界面，每个 Agent 一个工作目录、一条会话队列，彼此可以互发消息。截至 2026-11-12，仓库有 7,129 star、720 fork、1,825 open issue，最近一次推送就在当天。

本文不写功能清单，只答一个问题：这个工作台的边界怎么划。分四层：进程与依赖、状态归属、权限与可写范围、许可与维护。

## 一、它是什么，以及一个必须先看清的许可

先说许可，它决定"能不能用"早于"怎么用"。`LICENSE` 是 **PolyForm Noncommercial 1.0.0**，不是 MIT 或 Apache；文件末尾写明商业许可需另行联系。

实际含义：

- 个人研究、实验、测试与爱好项目属于允许用途；
- 慈善、教育、公共研究、公共安全与卫生、环保、政府机构的使用明确允许；
- 商业用途不在授权范围内，需单独取得许可；
- GitHub 把 license 报成 `NOASSERTION`，因为这不是 OSI 认可的开源许可。

把它当成"开源项目"评估会得出错误结论：要进公司生产流程，第一步是走授权，而不是技术评审。

## 二、进程与依赖：三条被工具强制的规则

仓库是 Bun workspace：桌面端留在根目录，另有 `apps/`（账号服务、移动端、站点路由等 Cloudflare Worker 与 Expo 应用）、`packages/`（共享 UI、契约、团队客户端）、`src/`（Electron 的 main、preload、renderer 与后端运行时）、`remote/`（Bun 写的 Signal 服务）与 `scripts/`、`tools/`。默认分支 8,707 个文件，`packages` 3,047、`apps` 2,016、`src` 1,856。

值得抄的是依赖方向，文档画成一条单向链：

```text
renderer ──► packages/contracts ◄── preload ◄── main ──► backend
                     ▲                             │
                     └──────── auth-api ──────────┘
```

围绕这条链有四条约束：

1. `packages/contracts` 不依赖 Electron、Node.js、SolidJS、provider 或 Cloudflare，只放跨进程类型与纯校验。
2. renderer 不能 import `src/main` 或 `src/backend`；preload 只把类型化调用映射成 IPC 通道。
3. main 调用服务之前必须校验来自 IPC 的不可信输入。
4. provider 代码不能直接写 UI 状态：事件交给 `AgentService`，先写 SQLite 投影，再推给 renderer。

第 4 条最容易被忽略。多数 Electron 应用的问题不是进程分不干净，而是"后端状态"与"界面状态"相绑：某次推送失败，界面就显示了一个不存在的世界。拆成两步等于把界面变成可随时重建的缓存。

四层各自被允许做什么，可以列成一张表：

| 层 | 负责 | 不该做的事 |
| --- | --- | --- |
| renderer | 屏幕投影与交互 | 直连数据库、import 主进程 |
| preload | 映射 IPC 通道 | 承载业务规则 |
| main | 校验输入、系统适配 | 绕过服务层写状态 |
| backend | 运行时、适配、事件与队列 | 直接改 UI 状态 |

更少见的是这些规则不靠人守：文档写明 `biome.json` 里的 `noRestrictedImports` 覆盖项与 `noNodejsModules` 强制执行上述 import 规则——也就是说，约束落在工具配置里，而不是写在评审清单上。是否真的阻断合并，文档没有承诺：`check:desktop:static` 把 `bun run lint`（即 `biome check`）与类型检查、构建串在一起，最终取决于 CI 跑哪条命令。

## 三、状态归属：一个数据库，三条旁路

工作台最容易腐烂的地方是"同一件事记在两个地方"。这里把真相收进一个文件，其余每种状态单独说明归属。

| 状态 | 位置 | 归属 |
| --- | --- | --- |
| Agent、会话、队列、附件索引、provider 会话绑定 | `userData/openbot.db` | 唯一真相源，事件日志加投影 |
| 浏览器标签 | `browser-tabs.json` | 不入迁移器，自带版本号 |
| Agent 自建表 | `Shared/Data/agent-data.db` | 一个命名空间，表归属创建者 |
| 各 CLI 登录态 | `~/.codex`、`~/.claude` 等 | provider 自管，不计入会话存储 |
| 账号、成员、邀请 | Cloudflare D1 | 服务侧，与本地会话分开 |
| 订阅 | Stripe（D1 为副本） | 以支付侧为真相 |

几个细节能看出设计意图：

- **浏览器状态单独存在**，因为它由浏览器自身版本化，且不能进迁移器。文档甚至写明：读到校验失败的标签环境时，仍返回这个标签页但不带环境，因为"丢掉用户打开的标签比丢掉一个模拟视口更糟"。
- **`agent-data.db` 让 Agent 自己建表并跨表 join**，改表权限绑在创建者身上，靠 SQLite 的 authorizer 执行，而不是检查模型写出的 SQL。前者是机制，后者是模式匹配。
- **删除 Agent 会连带删工作目录与只发给它的附件**，被多个 Agent 引用的传输保留。

## 四、"本地优先"不等于"离线"

README 写得很直白：local-first，不是 offline-only。随后列出每个 provider 实际连向哪里——Codex 连 OpenAI、Claude 连 Anthropic、Grok 连 xAI、Gemini 连 Google、Kimi 连 Moonshot、Devin 连 Cognition、Copilot 连 GitHub、Hermes 连 Nous Portal；访问网页走网络，插件可连自有服务。

这句话决定了合规讨论的起点：数据在本地留存，但推理请求不在本地。当成"数据不出机器"的方案会出错。

远程访问的设计也遵循同一条思路：

- 启用 Team API 后，认证过的接口仍只监听 localhost；
- 桌面与移动端通过 WebRTC（协议 v3）传 RPC、事件与二进制文件；
- 账号流程走 Cloudflare HTTPS，客户端只存一个加密会话令牌，一次性验证码 10 分钟过期、只存哈希；
- 嵌入式浏览器跑在独立沙箱 session 里，访问不到 `window.openbot` 与托管附件。

## 五、权限：三档 + 一个明说的非目标

README 顶部有一条警告：Agent 以 `danger-full-access` 与 `approvalPolicy: never` 运行，首次启动同意后不再逐动作确认，可读写文件、执行命令、访问网络、控制内嵌浏览器。

安全一节还有一句更值得抄下来的话："Full local access is an explicit current product decision, not a security boundary." 也就是说，全权限是当前的产品选择，不是这套程序提供的安全边界。

权限也不是只有两态。文档记了三档：Workspace only、Ask first、Full access。Workspace only 的可写根由三部分构成：Agent 工作区、`~/OpenBot/Shared`、用户额外添加的目录（最多 32 个）。细节有三条：

- 主进程拒绝把磁盘根、家目录、以及包含 OpenBot 数据的目录加进可写列表；
- 只有本地的 `agentFolders` IPC 能改这个列表，Agent 无法给自己加目录；
- 该列表和访问级别都只存在本地，不经 Team API 传递。

还有个逆向细节：Codex 的 `approvalPolicy` 保持 `on-request` 而非 `never`，因为 `never` 会拒绝 MCP 工具调用与 elicitations。全权限在实现上也不是一刀切。

结论很清楚：**若你需要"Agent 不能做危险动作"，这里的默认答案不是隔离，而是审计与备份。** 文档本身就建议备份、只跑可信任务，并把 IPC 校验绕过、附件路径逃逸列为特别欢迎的报告。

## 六、被忽略的运维代价：托管 CLI

工作台把十几个 CLI 下载并锁定在一个共享目录（macOS 上是 `~/Library/Application Support/OpenBot/provider-runtimes`），全机一份，所有 profile 与打包应用读同一份。

配套机制：

- 启动时与每小时检查上游最新版，来源分散在 GitHub Releases、npm `latest`、xAI 稳定通道、Cognition 清单与 ACP registry；
- 按上游自己的哈希校验：GitHub 的 asset `digest`、npm `dist.integrity`、manifest 的 `checksum`/`sha256`；**x.ai 与 ACP registry 不发哈希，只能依赖 TLS**；
- 安装后写一份 `openbot-install.json` 记录每个文件的 SHA-256，每次启动校验并与 `--version` 比对；
- 有一份 blocklist 用来压掉已知坏版本的"提供更新"，但它不会卸载用户已经装上的版本，读不到清单时也不拦任何东西；
- 每 6 小时跑 canary，用免费模型对最新版跑一致性用例，连败两次自动开 issue；
- 激活失败时删掉被拒产物、保留旧运行时，"更新坏了"不会让工作台不可用。

**自己托管十几个第三方 CLI 的工具，稳定性上限取决于上游发布质量，而非自己的代码。** 这也是 issue 数量偏高的结构性原因之一。

## 七、限制与结论

把上面几层合起来，可以对这套架构下三个结论。

**第一，边界是"进程 + 依赖 + 状态"三层，不是"沙箱"。** 被严格执行的是依赖方向（lint 强制）、状态归属（一个数据库加三条旁路）与可写根。隔离靠外部手段。

**第二，本地优先不等于私有部署。** 推理请求照常发往各家 provider，远程访问依赖 Cloudflare 与 WebRTC；留存位置与流向要分开评估。

**第三，许可是第一道门。** 个人与小范围研究可直接用，商用需单独授权；这与"能不能跑起来"无关，但决定它能否进公司流程。

只想在自己机器上跑多 Agent 并行，值得看状态归属与任务队列；评估团队采用，先看许可与安全边界，再看托管 CLI 的更新机制——长期成本就在这三件事里。

## 八、三个可以直接对照的自检问题

评估这类工作台，比功能清单更有用的是三个问题：

1. **会话记在哪里？** 主体在 SQLite、provider session 只作绑定时，换 provider 或重装 CLI 都不丢历史；反过来，一次升级就可能让历史不可读。
2. **它允许 Agent 写到哪里？** 三档权限加上一个显式的可写根白名单，比"默认全权限 + 事后审计"更容易向合规解释，即便最终选的是后者。
3. **它怎么处理上游的二进制？** 托管下载、按来源校验、blocklist、canary、激活失败回滚——如果这几件都缺，那么"支持十几个 CLI"会变成十几个各自的故障点。

三个问题不必读源码就能问，答案决定它半年后还能不能用。

## 参考来源

| 来源 | 链接 | 检索日期 |
| --- | --- | --- |
| 仓库 README（定位、安全警告、本地数据与网络边界、安全一节、发布流程） | https://raw.githubusercontent.com/nightly-labs/openbot/main/README.md | 2026-11-12 |
| 许可证全文（PolyForm Noncommercial 1.0.0，商用需联系） | https://raw.githubusercontent.com/nightly-labs/openbot/main/LICENSE | 2026-11-12 |
| 仓库结构与依赖方向、状态归属 | https://raw.githubusercontent.com/nightly-labs/openbot/main/docs/architecture/structure.md | 2026-11-12 |
| Providers 架构（模型目录、CLI 托管更新、blocklist、canary） | https://raw.githubusercontent.com/nightly-labs/openbot/main/docs/architecture/providers.md | 2026-11-12 |
| 仓库元数据（star、fork、issue、创建与推送时间） | https://api.github.com/repos/nightly-labs/openbot | 2026-11-12 |
