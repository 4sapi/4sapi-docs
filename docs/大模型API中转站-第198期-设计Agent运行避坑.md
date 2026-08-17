---
title: "【大模型API中转站】第198期 设计Agent运行避坑 | 从空白页到可交付"
category: 人工智能
tags:
  - 大模型API中转站
  - Open Design
  - Codex
  - MCP排错
  - Agent工作流
  - 企业级API
  - 生产环境接入
description: "Open Design 接入 Codex 后，最常见的问题不是模型不会设计，而是环境、端口、MCP、PATH、pnpm、长任务和权限没有理顺。本文按排查顺序讲如何从空白页、MCP不可见、spawn失败、任务卡住一路定位到可交付。"
---

# 【大模型API中转站】第198期 设计Agent运行避坑 | 从空白页到可交付

本文是【大模型API中转站】系列的第198篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一篇写了 Open Design 接 Codex 的整体思路。

这篇写真实落地时最容易踩的坑。

很多人会以为问题在模型：

```text
是不是 Codex 不会用？
是不是 Open Design 不稳定？
是不是 MCP 不兼容？
```

实际排查下来，常见问题往往更朴素：

```text
Node 版本不对。
pnpm 版本不对。
端口不是默认端口。
Web 首次编译还没完成。
MCP 配置写入了但会话没刷新。
Windows 下 shim 被 Node spawn 时失败。
Open Design 没有 active project。
长任务其实还在跑。
```

企业级大模型接入也是这样。

不是接通 API 就结束。

要把运行环境、日志、权限、预算和排错路径一起打通。

## 1. 先区分四层问题

Open Design + Codex + MCP 至少有四层：

```text
本地运行层：Node、pnpm、依赖、端口。
Open Design 层：daemon、web、项目、artifact。
MCP 层：stdio server、Codex 配置、工具列表。
模型调用层：Codex、Claude、GPT、4SAPI、预算和日志。
```

不要混在一起排查。

比如：

```text
Web 打不开，不一定是 MCP 问题。
MCP 看不到，不一定是 Open Design 没启动。
start_run 慢，不一定是模型失败。
模型报错，不一定是 Codex 配置错。
```

先判断坏在哪一层。

## 2. Node 和 pnpm 版本先对齐

Open Design 这类项目通常会锁 Node 和 pnpm。

如果本机有多个 pnpm，很容易出现：

```text
安装依赖时是一个版本。
运行脚本时是另一个版本。
```

常见现象：

```text
install 成功。
tools-dev 启动失败。
提示当前 pnpm 版本不符合项目要求。
```

排查命令：

```text
node --version
pnpm --version
corepack pnpm --version
where pnpm
```

建议：

```text
用项目 packageManager 指定的 pnpm。
优先让 Corepack 接管。
不要同时混用全局 pnpm 和 corepack pnpm。
```

如果企业团队要统一试点，最好把环境写进 README 或内部 SOP：

```text
Node 版本。
pnpm 版本。
启动命令。
状态检查命令。
停止命令。
```

否则每个人本机都可能是不同故障。

## 3. 不要假设端口一定是 7456

Open Design 本地开发模式可能会自动选择可用端口。

你以为是：

```text
http://127.0.0.1:7456
```

实际可能是：

```text
daemon: http://127.0.0.1:55711
web:    http://127.0.0.1:55712
```

所以不要只测试默认端口。

先看：

```text
pnpm tools-dev status
```

确认：

```text
daemon: running
web: running
desktop: idle 或 running
```

再测：

```text
GET /api/health
```

能返回：

```text
ok: true
version: ...
```

说明 daemon 是活的。

如果 Web 页面还在加载，先等首次编译完成。

Next.js 首次启动不是瞬间完成。

## 4. 空白页不一定是应用坏了

本地 Web 打开后，可能先看到：

```text
Loading Open Design...
```

这时不要马上判定失败。

排查顺序：

```text
1. 看 Web 端口是否 200。
2. 看浏览器 title 是否是 Open Design。
3. 等待前端 hydrate。
4. 看控制台 error。
5. 看 daemon health。
6. 看 tools-dev logs。
```

如果最后进入 onboarding 页面，说明前端已经起来。

企业内部培训时要说清楚：

```text
首次启动慢，不等于启动失败。
端口动态，不等于端口错误。
onboarding 页出现，说明 Web 基本可用。
```

## 5. MCP 添加成功后要新开 Codex 会话

很多人会遇到：

```text
codex mcp list 能看到 open-design。
但当前 Codex 对话里没有 open-design 工具。
```

这通常不是安装失败。

而是当前会话启动时，工具列表已经加载完了。

处理方式：

```text
新开 Codex 线程。
或重启 Codex。
或刷新当前运行环境。
```

验证配置用：

```text
codex mcp get open-design
codex mcp list
```

看到：

```text
transport: stdio
command: node
args: ... cli.js mcp
```

说明配置已经写入。

之后再验证 MCP server 本身：

```text
initialize
tools/list
```

能返回工具列表，才算完整。

## 6. Windows 下 spawn EPERM 怎么看

Windows 上有一个很常见的问题：

```text
命令行里 codex 能运行。
但 Node 里 spawn('codex') 报 EPERM。
```

这通常和 shim 有关。

PowerShell 里看到的可能是：

```text
codex.ps1
codex.cmd
codex.exe
```

人手敲命令能跑，不代表 Node 直接 spawn 这个名字一定能跑。

解决思路：

```text
不要猜。
先 where codex。
再看 Get-Command codex。
如果自动安装器 spawn 失败，就用 Codex CLI 自己的 mcp add。
```

比如：

```text
codex mcp add open-design --env OD_DATA_DIR=... --env OD_SIDECAR_IPC_PATH=... -- node cli.js mcp
```

这不是绕过安全。

只是把“由 Open Design 安装器调用 Codex”改成“直接调用 Codex MCP 配置命令”。

企业 SOP 里建议保留这条备用路径。

## 7. active context 为空不是错误

Open Design MCP 有 `get_active_context`。

它依赖用户当前在 Open Design 里打开的项目和文件。

如果你刚进入 onboarding，或者没有打开任何项目，工具可能返回：

```text
active: false
```

这不是 MCP 坏了。

只是没有当前上下文。

处理方式：

```text
先 create_project。
或者在 Open Design UI 里打开一个项目。
或者调用 list_projects 后显式传 project。
```

不要让 Agent 在没有项目的情况下无限尝试 `get_artifact`。

正确工作流是：

```text
list_projects
get_project
get_artifact
```

如果没有项目：

```text
create_project
start_run
get_run
get_artifact
```

## 8. start_run 慢不等于卡死

设计生成不是普通接口。

一次 `start_run` 可能包含：

```text
理解 brief。
选择 skill。
读取 design system。
调用 Agent。
生成 HTML。
自检。
保存 artifact。
```

如果背后调用的是高级模型，可能需要几分钟。

不要每 5 秒就取消。

建议：

```text
每 30-60 秒 get_run 一次。
超过预期再看日志。
不要用 write_file 伪造结果替代 start_run。
```

如果企业使用 4SAPI 做模型入口，要在日志里记录：

```text
run_id
project_id
agent
model
request_id
started_at
finished_at
status
cost
```

这样才能知道：

```text
是模型慢。
是工具慢。
是队列慢。
还是保存 artifact 慢。
```

## 9. Agent PATH 要和启动环境一致

Open Design 会扫描本机有哪些 Agent CLI。

比如：

```text
codex
claude
opencode
cursor-agent
gemini
```

但 GUI 启动环境和终端环境的 PATH 不一定一样。

终端里能找到 Codex，不代表 daemon 进程也能找到。

排查方式：

```text
在同一个启动环境里执行 codex --version。
在 Open Design 设置里 rescan。
查看 list_agents。
```

企业桌面环境里尤其常见：

```text
管理员安装路径。
用户 npm global 路径。
WindowsApps 路径。
Node global 路径。
```

这些都可能不同。

建议统一：

```text
Codex CLI 安装方式。
Node 安装方式。
PATH 注入方式。
Open Design 启动方式。
```

否则支持成本会很高。

## 10. 4SAPI 在排错里的位置

如果 Open Design 能跑，MCP 能连，但生成任务报模型错误，就进入 4SAPI 排查。

先看：

```text
API Key 是否有效。
模型是否在分组里。
是否有额度。
是否触发限流。
是否 504/524 超时。
是否 500/503 渠道不可用。
```

不要把模型错误和本地 MCP 错误混在一起。

一个简单判断：

```text
tools/list 失败：MCP 层问题。
list_projects 失败：Open Design daemon 或数据层问题。
start_run 创建失败：Open Design 项目/任务层问题。
模型调用失败：4SAPI/上游模型/Key 权限问题。
```

这样排查最快。

## 11. 给 AI 的排错 Prompt

```text
你是 Open Design + Codex + MCP 排错助手。

请根据下面信息判断故障在哪一层：

【本地环境】
- node --version：
- pnpm --version：
- pnpm tools-dev status：
- daemon health：
- web URL：

【MCP】
- codex mcp get open-design：
- tools/list 是否成功：
- 工具数量：

【Open Design】
- 是否有 active context：
- list_projects 是否成功：
- get_artifact 是否成功：

【生成任务】
- start_run 是否返回 runId：
- get_run 状态：
- 是否有 previewUrl：

【模型入口】
- 是否接入 4SAPI：
- 错误码：
- request_id：
- 模型名：
- Key 分组：

请输出：
1. 故障层级。
2. 最可能根因。
3. 下一步最小验证。
4. 不要做的误操作。
5. 是否需要管理员介入。
```

这个 Prompt 的关键是分层。

不要让 AI 直接猜“重装一下”。

## 12. 上线前最小验收

上线前至少跑通：

```text
Open Design web 页面可打开。
daemon health 返回 ok。
Codex MCP 列表能看到 open-design。
stdio MCP tools/list 能返回工具。
list_projects 能返回。
create_project 能创建项目。
start_run 能返回 runId。
get_run 能进入终态。
get_artifact 能拉到文件。
生成过程能在 4SAPI 日志里看到模型调用。
```

如果只是 Web 能打开，不算完成。

如果只是 MCP 配置存在，也不算完成。

必须跑一条完整链路。

## 13. 总结

Open Design 接 Codex 的坑，大部分不是设计能力问题。

而是：

```text
环境版本。
动态端口。
MCP 会话刷新。
Windows shim。
active project。
Agent PATH。
长任务耐心。
模型入口治理。
```

一句话：

```text
先分层，再验证；先 health，再 tools/list；先 get_artifact，再谈生成质量。
```
