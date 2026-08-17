---
title: "【大模型API中转站】第47期 Codex接入4SAPI | 桌面Agent省钱"
category: 人工智能
tags:
  - 大模型API中转站
  - Codex
  - 4SAPI
  - AI Agent
  - config.toml
  - 桌面Agent
description: "把 Codex 从聊天式使用升级成可管理的桌面 Agent 工作流：用 4SAPI 作为模型 API 入口，配置 Codex 本地任务、CLI 和 IDE 扩展的模型提供方，同时讲清 Responses API 兼容、Key 管理、权限和排错。"
---

# 【大模型API中转站】第47期 Codex接入4SAPI | 桌面Agent省钱

本文是【大模型API中转站】系列的第47篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

很多人第一次用 Codex，会把它当成 ChatGPT 桌面版。

打开窗口，打一段话，等它回答。回答不错，就复制出来；需要改文件，再手动上传下载、复制粘贴。

这样用当然也能用，但只吃到了 30% 的能力。

Codex 真正有价值的地方，是它可以进入你的工作区：读文件、改文件、跑命令、生成表格、做网页、写脚本、检查 Git diff。它不是一个更大的聊天框，而是一个能在电脑里干活的桌面 AI Agent。

如果再把模型调用统一走 4SAPI 这类大模型API中转站，就可以把 Codex 的使用从“个人随手试试”变成“Key、模型、成本、日志都可管理”的工作流。

这篇只讲合法合规的 API 接入和成本治理，不讨论任何绕过官方规则的用途。

## 1. 先分清：你要接入的是哪一种 Codex

Codex 现在不是单一入口，而是一组工作面。

常见可以分成四类：

| 入口 | 适合做什么 | 4SAPI 接入重点 |
| --- | --- | --- |
| Codex App | 桌面项目、Worktree、本地文件处理、可视化协作 | 本地任务能否读取用户级配置 |
| Codex CLI | 终端任务、脚本化、CI、批量处理 | `~/.codex/config.toml` |
| Codex IDE Extension | 编辑器内改代码、解释代码、Review | 与 CLI 共享配置层 |
| Codex Web / Cloud | 云端后台任务、GitHub PR、并行开发 | 通常按官方云端环境走，不要默认等同于本地自定义 API |

这篇重点讲本地 Codex：Codex App 的本地任务、Codex CLI、Codex IDE Extension。

原因很简单：这些入口更接近你的电脑工作区，也更适合通过本地配置文件接入 4SAPI。

不要把“Codex 接入 4SAPI”理解成：

```text
把 ChatGPT 网页里的所有 Codex 能力都改成 4SAPI。
```

更准确的理解是：

```text
让本地 Codex 客户端通过自定义 model provider，把模型请求发到 4SAPI 的 API 入口。
```

这两个差别很大。前者容易误解，后者才是工程上可配置、可验证、可回滚的方案。

## 2. 为什么 Codex 特别适合接入 4SAPI

普通聊天工具通常是一问一答。

Codex 的典型调用链更像这样：

```text
读需求
-> 搜索文件
-> 读取代码或文档
-> 生成修改方案
-> 修改文件
-> 运行命令
-> 读取报错
-> 再次修改
-> 输出总结
```

一次任务可能触发很多轮模型调用。如果所有调用都走同一个昂贵模型，成本会涨得很快；如果每个工具各填一把 Key，又会很难排查到底是谁花了钱。

4SAPI 放在中间，主要解决四个问题：

| 问题 | 不统一时 | 统一到 4SAPI 后 |
| --- | --- | --- |
| Key 管理 | 每台电脑、每个工具单独填 | 按用途创建专用 Key |
| 模型切换 | 客户端里到处改 | 在统一入口复制模型名 |
| 成本记录 | 不知道一次任务花了多少 | 通过日志和 Token 记录回看 |
| 团队治理 | 每个人凭感觉用 | 可按项目、成员、任务分组 |

所以这套方案的价值不是“把 Base URL 换一下”这么简单，而是让桌面 Agent 的成本和风险可观察。

## 3. 最重要的一条：Codex 需要 Responses API 兼容

接入前先记住这一点：

```text
Codex 的自定义 model provider 当前主要按 Responses API 协议工作。
```

很多大模型中转站都支持 OpenAI-compatible 的 `/v1/chat/completions`，但 Codex 本地配置里的 provider 不只是普通聊天接口。你需要确认 4SAPI 侧的目标通道支持 Responses API，或者已经做了 Responses 到 Chat Completions 的兼容转换。

因此，先不要急着改 Codex 配置。

先用最小请求确认 4SAPI 的 Responses 入口可用：

```bash
curl https://4sapi.com/v1/responses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-xxxxxxxxxxxxxxxx" \
  -d '{
    "model": "your-4sapi-model-name",
    "input": "只回复 ok"
  }'
```

如果这个请求能正常返回，再继续配置 Codex。

如果 `/v1/chat/completions` 能通，但 `/v1/responses` 不通，说明这条模型通道可能适合普通客户端，却不一定适合 Codex 本地 provider。这个时候要换支持 Responses 的模型通道，或者在 4SAPI 侧确认是否有对应兼容入口。

## 4. 准备工作：三样东西

接入前准备三样东西：

| 项目 | 说明 |
| --- | --- |
| 4SAPI API Key | 建议单独创建给 Codex 用的 Key |
| 4SAPI Base URL | `https://4sapi.com/v1` |
| 模型名 | 从 4SAPI 模型广场复制完整模型名 |

Key 不建议混用。

至少准备一把：

```text
codex-local-dev
```

团队使用可以拆成：

```text
codex-personal
codex-team-dev
codex-review
codex-ci-low-budget
```

这样做的好处是，一旦某个 Agent 任务循环调用，也能快速定位是哪把 Key 在消耗。

## 5. 配置 Codex：修改 config.toml

Codex 的本地配置文件通常在：

```text
~/.codex/config.toml
```

Windows 上通常对应：

```text
C:\Users\你的用户名\.codex\config.toml
```

如果你用的是 IDE Extension，也可以从设置里打开 Codex 的 `config.toml`。

最小配置如下：

```toml
model_provider = "foursapi"
model = "your-4sapi-model-name"

[model_providers.foursapi]
name = "4SAPI"
base_url = "https://4sapi.com/v1"
env_key = "FOURSAPI_API_KEY"
wire_api = "responses"
request_max_retries = 4
stream_max_retries = 5
stream_idle_timeout_ms = 300000
```

这里的 `foursapi` 只是 provider id，你也可以叫 `custom`、`4sapi` 或 `codex_4sapi`。关键是两处保持一致：

```toml
model_provider = "foursapi"

[model_providers.foursapi]
```

其中三项最关键：

```text
model_provider = "foursapi"
base_url = "https://4sapi.com/v1"
wire_api = "responses"
```

`model` 一定要换成 4SAPI 后台模型广场里的完整模型名，不要手写简称。

不要写成：

```text
claude
sonnet
gpt
deepseek
```

要复制后台显示的完整名称。模型名不精确，是这类配置里最常见的坑。

还有一个细节：如果你用的是 4SAPI 自己的 API Key，建议使用 `env_key`。如果某个 provider 使用 OpenAI 登录态认证，才会涉及 `requires_openai_auth = true`。这几种认证方式不要混在一起写。

## 6. Windows 用户怎么设置 Key

建议用环境变量保存 Key，不要直接把真实 Key 写进 `config.toml`。

PowerShell 临时设置：

```powershell
$env:FOURSAPI_API_KEY="sk-xxxxxxxxxxxxxxxx"
codex
```

如果想长期保存到用户环境变量：

```powershell
[Environment]::SetEnvironmentVariable("FOURSAPI_API_KEY", "sk-xxxxxxxxxxxxxxxx", "User")
```

设置后重新打开终端或重启 Codex，再检查：

```powershell
echo $env:FOURSAPI_API_KEY
```

macOS / Linux 可以写：

```bash
export FOURSAPI_API_KEY="sk-xxxxxxxxxxxxxxxx"
codex
```

如果你是团队环境，不建议把 Key 写进仓库里的 `.codex/config.toml`。用户级 `~/.codex/config.toml` 更适合放 provider 配置，项目级配置更适合放权限、工作流和规则。

这里特别容易搞错：项目里的 `.codex/config.toml` 适合放项目级规则，但不适合放 `model_provider` 和 `model_providers`。Codex 会把这类 provider 配置当成机器本地配置处理，应该放在用户级 `~/.codex/config.toml`。

## 7. 启动前先做两个验证

第一步，验证 4SAPI 接口：

```bash
curl https://4sapi.com/v1/responses \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $FOURSAPI_API_KEY" \
  -d '{
    "model": "your-4sapi-model-name",
    "input": "返回 ok"
  }'
```

Windows PowerShell 可以写成：

```powershell
curl.exe https://4sapi.com/v1/responses `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $env:FOURSAPI_API_KEY" `
  -d "{ `"model`": `"your-4sapi-model-name`", `"input`": `"返回 ok`" }"
```

第二步，启动 Codex 时指定模型：

```bash
codex --model your-4sapi-model-name
```

或者临时覆盖配置：

```bash
codex -c model_provider="foursapi" -c model="your-4sapi-model-name"
```

如果你打开的是 Codex App，建议先用一个测试项目启动本地任务，不要一上来就打开真实生产仓库。

最小测试指令：

```text
请读取当前目录，只列出文件名，不要修改任何文件。
```

确认它能正常响应后，再让它做只读分析：

```text
请扫描当前项目，告诉我主要文件结构和你建议我先看哪几个文件。不要修改文件。
```

最后再进入可写任务。

## 8. 把 Codex 当 Agent 用，不要当聊天框用

接入 API 只是第一步。

真正让 Codex 变强的是四个基本功：专案、权限、上下文、规则。

### 8.1 专案：先圈出工作范围

在哪个资料夹里打开 Codex，那个资料夹就变成工作现场。

不要让它在整台电脑里乱找资料。更好的做法是：

```text
项目资料
参考文档
输出目录
规则文件
```

都放在同一个清楚命名的工作区里。

你可以这样说：

```text
请先扫描当前项目，只总结目录结构和文件用途。不要修改文件。
```

这一步能避免后面很多误操作。

### 8.2 权限：先只读，再放开

Codex 能读文件、改文件、跑命令，所以权限要分阶段。

建议顺序：

| 阶段 | 权限策略 |
| --- | --- |
| 第一次接入 | 只读分析 |
| 小任务试跑 | 允许修改工作区文件 |
| 熟悉后 | 根据任务开放命令执行 |
| 团队/生产项目 | 必须配 Git、Review、测试 |

不要第一次接入 4SAPI 就让 Codex 在重要仓库里批量改文件。

更稳的提示词：

```text
修改前先列计划。
只修改与任务直接相关的文件。
每次修改后说明改了什么。
不要删除原始资料。
```

### 8.3 上下文：不是越多越好

很多人以为给 Codex 的资料越多越好。

实际不是。

资料太多，模型要读的东西变多，任务焦点反而容易散。尤其接入 API 后，上下文越长，Token 成本也越高。

更好的方法是：

```text
一个阶段只给必要资料。
完成后让 Codex 总结当前状态。
再开启下一阶段任务。
```

可复制提示词：

```text
请总结目前已经确认的信息、已经修改的文件、未解决的问题和下一步建议。
```

这句话很适合在长任务中途使用，也方便你换模型、换线程或压缩上下文。

### 8.4 规则：把偏好写进 AGENTS.md

如果一个项目会长期交给 Codex 处理，建议在项目里放一份 `AGENTS.md`。

它可以写：

```markdown
# 项目协作规则

- 默认使用简体中文输出。
- 修改文件前先说明计划。
- 不要删除原始数据文件。
- 代码修改后优先运行现有测试。
- 写文章时保留 YAML front matter。
- 输出总结需要包含：改动内容、验证方式、风险点。
```

`AGENTS.md` 管的是项目规则，`~/.codex/config.toml` 管的是本地模型、权限、provider 等配置。

不要把 API Key 写进 `AGENTS.md`。

## 9. 推荐的 4SAPI 模型分层

Codex 不一定每次都要用最强模型。

可以按任务拆：

| 任务 | 推荐模型档位 |
| --- | --- |
| 扫描目录、整理文件 | 低成本通用模型 |
| 写博客、整理资料 | 中等文本模型 |
| 改代码、写测试 | 代码能力强的模型 |
| 跨文件重构 | 长上下文强推理模型 |
| Review 和查错 | 与主力模型不同的强模型 |

如果 4SAPI 支持在后台按 Key 或模型分组统计日志，建议把 Codex 的几类任务拆开看。

例如：

```text
codex-writing-key：写作、整理、总结
codex-coding-key：代码修改、测试生成
codex-review-key：代码审查、风险检查
codex-ci-key：自动化任务，额度更低
```

这样你后面复盘成本时，能看清楚到底是写作贵、改代码贵，还是某次自动化任务跑飞了。

## 10. 常见报错和排查

### 10.1 401 或 Key 错误

优先检查：

```text
环境变量名是不是 FOURSAPI_API_KEY
Key 有没有多写 Bearer
终端重启后环境变量是否还在
4SAPI 后台令牌是否被删除或禁用
```

`config.toml` 里写的是：

```toml
env_key = "FOURSAPI_API_KEY"
```

这表示 Codex 会去环境变量里读 Key，不是让你把真实 Key 填在这里。

### 10.2 404 或 endpoint 不存在

优先检查：

```text
base_url 是否是 https://4sapi.com/v1
4SAPI 通道是否支持 /v1/responses
是否误把 Chat Completions 通道当成 Codex provider 用
```

如果普通聊天接口能通，但 Codex 不通，重点看 Responses API 兼容。

### 10.3 模型不存在

最常见原因是模型名手写错了。

处理方式：

```text
去 4SAPI 模型广场复制完整模型名。
确认这把 Key 的分组有权限调用该模型。
不要用客户端展示名代替真实 model id。
```

### 10.4 一直转圈或中途断流

可能原因：

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| 一直等待 | 模型响应慢、网络不稳、流式中断 | 换模型或提高超时时间 |
| 长任务断掉 | SSE 空闲超时 | 调整 `stream_idle_timeout_ms` |
| 复杂任务失败 | 上下文太长或模型能力不足 | 缩小任务范围，换更强模型 |
| 成本异常 | Agent 循环调用 | 降低权限，设置 Key 额度 |

配置里可以先保留：

```toml
stream_idle_timeout_ms = 300000
request_max_retries = 4
stream_max_retries = 5
```

不要为了“更稳”无限提高重试次数。Agent 任务一旦跑偏，重试越多，花费越高。

## 11. 一套适合非技术人员的 Codex 工作流

如果你不是工程师，也可以这样用。

假设你要准备一场讲座，资料夹里有报名表、讲师介绍、活动说明和历史海报。

不要直接说：

```text
帮我做一个讲座方案。
```

更稳的流程是：

```text
第一步：请扫描当前资料夹，列出文件清单和每个文件用途，不要修改。
第二步：请整理报名表，输出一份缺席签到用名单。
第三步：请根据活动说明生成 10 页简报大纲。
第四步：请生成活动页面草稿，先用单文件 HTML。
第五步：请总结你修改了哪些文件，以及我需要人工确认什么。
```

如果接入了 4SAPI，你还可以把模型分层：

```text
扫描资料：低成本模型
整理表格：低成本模型或代码模型
写简报：中等文本模型
生成网页：代码模型
最终润色：强模型
```

Codex 负责进入工作现场，4SAPI 负责把模型调用管起来。

## 12. 安全边界：别把 API 接入当万能开关

有三条底线建议写在项目规则里。

第一，不要让 Codex 删除原始资料。

```text
所有原始文件只能读取，不允许删除或覆盖。
生成的新文件放到 output 或 drafts 目录。
```

第二，不要让生产 Key 进入个人测试项目。

```text
Codex 本地试验使用 codex-local-dev Key。
生产服务使用 production Key。
两者不能混用。
```

第三，重要任务必须看 diff。

```text
任务结束后，请输出修改文件列表和每个文件的改动摘要。
我确认后再提交 Git。
```

模型接入只是让 Codex 能回答和行动；真正的安全来自权限、目录、Git 和人工验收。

## 13. 总结：4SAPI 管模型，Codex 管现场

一句话总结：

```text
Codex 是进入电脑工作区的 Agent，4SAPI 是统一模型 API 和成本治理的网关。
```

接入时不要只记一个 Base URL。

更完整的清单是：

```text
Base URL: https://4sapi.com/v1
API Key: 用环境变量保存
Model: 从 4SAPI 模型广场复制完整名称
Protocol: 确认支持 Responses API
Config: 写到 ~/.codex/config.toml
Safety: 先只读、再小范围修改、最后看 diff
```

这套方案适合三类人：

| 人群 | 价值 |
| --- | --- |
| 个人创作者 | 用 Codex 整理资料、写文章、做页面，同时控制模型成本 |
| 开发者 | 用 Codex 改代码、跑测试、做 Review，统一 Key 和日志 |
| 团队 | 把 Agent 使用纳入预算、权限、审计和项目规则 |

不要把 Codex 当成另一个聊天机器人。把项目范围、规则、权限和模型入口配好，它才会从“会回答问题”变成“能帮你完成工作”。

下一篇可以继续拆：如何给 Codex 项目写一份真正有用的 `AGENTS.md`，让每个新任务都从正确规则开始。

## 参考资料

- Codex 配置基础：https://developers.openai.com/codex/config-basic
- Codex 高级配置与自定义 model provider：https://developers.openai.com/codex/config-advanced
- Codex 配置项参考：https://developers.openai.com/codex/config-reference
- Codex 示例配置：https://developers.openai.com/codex/config-sample
- 4SAPI 接入文档：https://4sapi.apifox.cn/
- 4SAPI API 地址：https://4sapi.com/v1
