---
title: "【大模型API中转站】第95期 DeepSeek V4 Flash | 接入Agent全家桶"
category: 人工智能
tags:
  - 大模型API中转站
  - 4SAPI
  - DeepSeek V4
  - deepseek-v4-flash
  - AI Agent
  - 编程助手
description: "基于 DeepSeek 官方 awesome-deepseek-agent 清单，讲解 4SAPI 的 deepseek-v4-flash 如何按官方同类方式接入 AstrBot、Cherry Studio、Claude Code、Cline、Codex、OpenClaw、OpenCode、Qwen Code 等 Agent 工具，并给出 Base URL、Key、模型名和排错清单。"
---

# 【大模型API中转站】第95期 DeepSeek V4 Flash | 接入Agent全家桶

本文是【大模型API中转站】系列的第95篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

DeepSeek 最近把 Agent 生态整理得很清楚。

官方仓库 `awesome-deepseek-agent` 里，已经列了一批主流 AI Agent 和编程助手接入指南。

严格说，这个 README 里列的不是“模型”。

而是工具：

```text
AstrBot
Cherry Studio
Claude Code
Cline
Codex
Crush
Deep Code
DeepSeek-TUI
GitHub Copilot
GitHub Copilot CLI
Hermes
Kilo Code
Langcli
LobeHub
nanobot
Oh My Pi
OpenClaw
OpenCode
Pi
Qwen Code
Reasonix
WorkBuddy / CodeBuddy
```

官方 README 的重点是：

```text
这些工具可以接入 DeepSeek-V4-Pro 或 DeepSeek-V4-Flash。
```

那问题来了：

```text
如果我不用官方 DeepSeek Key，
而是用 4SAPI 的 deepseek-v4-flash，
还能不能接这些 Agent 工具？
```

答案是：

```text
能接。
```

但要分情况。

如果工具支持自定义 Base URL、OpenAI-compatible Provider、Anthropic-compatible Provider，接入会很顺。

如果工具只有一个写死的 DeepSeek 内置 Provider，而且不允许改 endpoint，就要找它的 Custom Provider、OpenAI Compatible、第三方模型、代理转发层，或者等工具后续开放自定义地址。

这篇就把官方清单换成 4SAPI 视角讲清楚：

```text
官方文档用 DeepSeek API Key。
4SAPI 接入用 4SAPI API Key。
官方文档用 deepseek-v4-flash。
4SAPI 模型名也用 deepseek-v4-flash。
官方文档指向 DeepSeek endpoint。
4SAPI 场景优先填 https://4sapi.com/v1 或 https://4sapi.com。
```

这样你就可以把 DeepSeek-V4-Flash 放进自己的 Agent 工具链，而不是每个工具单独折腾一把 Key。

## 1. 为什么是 deepseek-v4-flash

DeepSeek V4 系列可以粗略分两类：

```text
deepseek-v4-pro：更适合复杂规划、深度推理、架构改造。
deepseek-v4-flash：更适合高频执行、日常编码、问答、批量任务。
```

如果你用 Agent 工具，不一定每一步都要 Pro。

真实工作流往往是：

```text
读文件
解释报错
生成小函数
改 README
写测试样例
整理 issue
总结聊天记录
跑批量问答
做轻量客服
```

这些任务对延迟和成本更敏感。

这就是 deepseek-v4-flash 的价值。

把它接到 4SAPI 后，你能得到三个好处：

```text
一个 Key 管多个 Agent。
一个后台看所有调用日志。
一个模型名在多个工具里复用。
```

尤其是团队里同时用 Claude Code、Codex、OpenClaw、Qwen Code、Cline、Cherry Studio 的时候，统一入口非常重要。

否则会变成：

```text
每个工具一把 Key。
每个人一套配置。
谁花了多少钱没人知道。
出了问题不知道请求打到哪里。
```

4SAPI 的位置，就是把模型入口、Key、日志、成本和权限收拢起来。

## 2. 官方清单到底覆盖了哪些工具

DeepSeek 官方的 `awesome-deepseek-agent` 中文 README 说得很直接：

```text
精选 DeepSeek 模型接入指南合集。
帮助你将 DeepSeek 接入主流 AI Agent 与编程助手工具。
每份指南包含安装、配置与首次运行步骤。
目标模型包括 DeepSeek-V4-Pro 和 DeepSeek-V4-Flash。
```

我把清单按使用场景拆一下。

### 2.1 聊天机器人和多平台 Agent

```text
AstrBot
nanobot
OpenClaw
Hermes
LobeHub
```

适合：

```text
QQ群 / 微信 / 飞书 / Telegram 接入
个人 AI 助手
多 Agent 编排
长期记忆
MCP / Skill 扩展
团队工作台
```

这类工具很适合走 4SAPI。

因为消息平台一接上，调用会变得很频繁。

如果不做 Key 拆分，很容易出现：

```text
一个群聊把预算打爆。
一个插件循环调用没人发现。
一个成员测试机器人导致生产 Key 被耗尽。
```

### 2.2 终端编程 Agent

```text
Claude Code
Codex
Crush
Deep Code
DeepSeek-TUI
GitHub Copilot CLI
Langcli
Oh My Pi
OpenCode
Pi
Qwen Code
Reasonix
```

适合：

```text
代码生成
仓库阅读
自动修 bug
终端里跑 Agent
多文件修改
MCP 工具调用
长上下文代码分析
```

这类工具最适合 deepseek-v4-flash 做日常执行模型。

复杂任务可以上 Pro。

但高频读写、解释、改小文件、写测试，Flash 更省。

### 2.3 IDE 和桌面客户端

```text
Cherry Studio
Cline
GitHub Copilot
Kilo Code
WorkBuddy / CodeBuddy
```

适合：

```text
VS Code 插件
桌面多模型对话
代码补全与问答
项目级辅助开发
自定义模型供应商
```

这类工具通常都有模型供应商配置页。

如果支持 OpenAI Compatible 或 Custom Provider，就可以直接把 4SAPI 填进去。

## 3. 4SAPI 接入的核心配置

大多数工具只要三项：

```text
Base URL
API Key
Model
```

4SAPI 常用填写：

```text
Base URL: https://4sapi.com/v1
API Key: 4SAPI 后台创建的 Key
Model: deepseek-v4-flash
```

有些工具的 Base URL 会自动拼 `/v1`。

这时可以填：

```text
https://4sapi.com
```

4SAPI 文档里也提醒过，配置 URL 时一般用这两个：

```text
https://4sapi.com
https://4sapi.com/v1
```

哪个能通用哪个。

但不要填成：

```text
https://4sapi.com/v1/v1/chat/completions
https://4sapi.com/v1/chat/completions/chat/completions
```

模型名要从 4SAPI 模型广场复制。

如果模型广场里的名字就是：

```text
deepseek-v4-flash
```

那配置里就不要改成：

```text
deepseek-v4
deepseek-v4-fast
DeepSeek-V4-Flash
deepseek-v4-flash[1m]
```

模型名必须一模一样。

## 4. 官方 DeepSeek Key 怎么替换成 4SAPI Key

官方接入指南通常是这样：

```text
Provider: DeepSeek
API Key: <DeepSeek API Key>
Model: deepseek-v4-flash
```

换成 4SAPI 时，看工具支持哪种模式。

### 4.1 支持 OpenAI-compatible 的工具

这是最简单的。

配置：

```text
Provider: OpenAI Compatible / Custom OpenAI
Base URL: https://4sapi.com/v1
API Key: <4SAPI_KEY>
Model: deepseek-v4-flash
```

常见于：

```text
Cherry Studio
Cline
OpenCode
Kilo Code
WorkBuddy / CodeBuddy
部分桌面客户端
部分 VS Code 插件
```

优先找这些菜单名：

```text
OpenAI Compatible
Custom Provider
Custom API
Third-party Provider
自定义模型
兼容 OpenAI
```

### 4.2 支持 Anthropic-compatible 的工具

Claude Code 这类工具更特殊。

DeepSeek 官方指南里，Claude Code 配的是：

```text
ANTHROPIC_BASE_URL
ANTHROPIC_AUTH_TOKEN
ANTHROPIC_MODEL
```

如果你用官方 DeepSeek endpoint，它会填 DeepSeek 的 Anthropic 兼容地址。

如果你改用 4SAPI，要看 4SAPI 对应工具教程或当前可用的兼容地址。

最稳的方式是：

```text
优先参考 4SAPI 的 Claude Code CLI 安装与配置教程。
或者用 CC Switch 管 Claude Code 供应商。
```

如果工具支持 OpenAI-compatible 转 Anthropic 的中间层，也可以通过转发层接。

不要盲目把 OpenAI 的 `/v1/chat/completions` 地址塞进只认 Anthropic Messages 的工具里。

协议不一致，会报错。

### 4.3 只有 DeepSeek 内置 Provider 的工具

有些工具会把 DeepSeek 做成内置选项。

你只能填：

```text
DeepSeek API Key
```

但不能改 Base URL。

这种情况，4SAPI 不能直接替换官方 endpoint。

你要找：

```text
是否还有 Custom Provider。
是否还有 OpenAI Compatible。
是否支持环境变量覆盖 Base URL。
是否支持本地代理。
是否支持配置文件手动写 endpoint。
```

如果都没有，就不要硬说能接。

正确做法是：

```text
官方 DeepSeek Provider 继续用官方 Key。
4SAPI 走该工具的 Custom/OpenAI-compatible/Proxy 模式。
```

这才是长期稳定的接入方式。

## 5. 按工具类型给一张迁移表

下面这张表，是按官方 README 的工具清单做的 4SAPI 接入思路。

不是每个工具都完全同一个配置页。

但思路一致。

```text
工具：AstrBot
官方思路：选择 DeepSeek Provider，填 DeepSeek Key。
4SAPI 思路：优先找自定义模型或兼容 OpenAI 的模型提供商，填 4SAPI Base URL、Key、deepseek-v4-flash。
适合场景：群机器人、多平台消息助手。
```

```text
工具：Cherry Studio
官方思路：桌面客户端里添加 DeepSeek。
4SAPI 思路：新建 OpenAI Compatible / 自定义供应商，Base URL 用 https://4sapi.com/v1，模型用 deepseek-v4-flash。
适合场景：多模型对话、知识库、桌面工作台。
```

```text
工具：Claude Code
官方思路：配置 ANTHROPIC_BASE_URL、ANTHROPIC_AUTH_TOKEN、ANTHROPIC_MODEL。
4SAPI 思路：参考 4SAPI Claude Code 教程或用 CC Switch 管供应商；注意 Claude Code 走 Anthropic 协议，不要直接混用 OpenAI URL。
适合场景：终端深度编码。
```

```text
工具：Cline
官方思路：VS Code 插件里选择 DeepSeek 或兼容供应商。
4SAPI 思路：选择 OpenAI Compatible / Custom Provider，填 https://4sapi.com/v1、4SAPI Key、deepseek-v4-flash。
适合场景：VS Code 内的 Agent 编程。
```

```text
工具：Codex
官方思路：DeepSeek 官方指南使用 Moon Bridge 做转发层，因为 Codex 使用 Responses API。
4SAPI 思路：如果用 Codex 原生工作流，优先参考 4SAPI GPT-codex 教程；需要转发层时，把上游模型换成 4SAPI deepseek-v4-flash。
适合场景：OpenAI Codex CLI / App 编程 Agent。
```

```text
工具：Crush
官方思路：终端 AI 编程 Agent，配置模型供应商。
4SAPI 思路：找 OpenAI-compatible / Custom provider，填 4SAPI 三件套。
适合场景：终端多模型编码。
```

```text
工具：Deep Code
官方思路：面向 DeepSeek-V4 系列模型适配。
4SAPI 思路：如果支持自定义 endpoint，直接把官方 DeepSeek 地址替换为 4SAPI；否则继续使用官方 provider。
适合场景：DeepSeek 原生编程助手。
```

```text
工具：DeepSeek-TUI
官方思路：面向 DeepSeek-V4 的 Rust 终端编程助手。
4SAPI 思路：查看配置是否允许 base_url；允许则填 4SAPI，不允许则用官方或本地代理。
适合场景：终端 TUI、长上下文代码任务。
```

```text
工具：GitHub Copilot / GitHub Copilot CLI
官方思路：官方指南给出 DeepSeek 接入路径。
4SAPI 思路：若当前版本允许第三方模型 endpoint，就按自定义供应商接；若只接受官方认证，不要强行替换。
适合场景：Copilot 生态里的编码辅助。
```

```text
工具：Hermes
官方思路：配置 DeepSeek 模型作为 Agent 后端。
4SAPI 思路：用 Hermes 的自定义模型或 Skill，把 Base URL / Key / model 做成参数，统一指向 4SAPI。
适合场景：自我进化 Agent、工具流、批量任务。
```

```text
工具：Kilo Code
官方思路：CLI 和编辑器扩展里接模型供应商。
4SAPI 思路：选择 OpenAI-compatible 或自定义供应商，填 4SAPI。
适合场景：轻量编码助手。
```

```text
工具：Langcli
官方思路：兼容 Claude Code，并支持 DeepSeek V4 等模型。
4SAPI 思路：按它的自定义 provider / 兼容协议配置 4SAPI deepseek-v4-flash。
适合场景：Claude Code 风格替代工具。
```

```text
工具：LobeHub
官方思路：作为 Chief Agent Operator 管理 AI 团队。
4SAPI 思路：给不同 Agent 角色分配不同 4SAPI Key 或模型；普通执行角色用 deepseek-v4-flash，复杂规划角色再上 Pro。
适合场景：多 Agent 团队调度。
```

```text
工具：nanobot
官方思路：轻量级 AI 智能体，接聊天工具、记忆、MCP。
4SAPI 思路：用 OpenAI-compatible 配置统一入口，按机器人用途拆 Key。
适合场景：轻量聊天机器人。
```

```text
工具：Oh My Pi / Pi
官方思路：终端编码框架，支持供应商、模型角色、MCP、插件。
4SAPI 思路：把 4SAPI 配成一个 provider，deepseek-v4-flash 作为默认执行模型。
适合场景：可扩展终端 Agent。
```

```text
工具：OpenClaw
官方思路：setup 时选择 DeepSeek，填 API Key，模型名填 deepseek-v4-pro 或 deepseek-v4-flash。
4SAPI 思路：如果 OpenClaw 当前配置支持自定义 API endpoint，就填 4SAPI；如果只是 DeepSeek 内置 provider，则通过 OpenAI-compatible / 自定义供应商 / 4SAPI OpenClaw 教程接。
适合场景：个人 AI 助手、飞书/微信、Skill。
```

```text
工具：OpenCode
官方思路：开源 AI 编程助手，支持终端和网页形态。
4SAPI 思路：参考 4SAPI OpenCode 接入配置，模型填 deepseek-v4-flash。
适合场景：开源编码助手。
```

```text
工具：Qwen Code
官方思路：/auth 选择 Third-party Providers，再选 DeepSeek API Key，模型可填 deepseek-v4-pro、deepseek-v4-flash。
4SAPI 思路：如果 Qwen Code 的 DeepSeek 内置 provider 不允许改 endpoint，就找自定义 OpenAI-compatible 配置；能改 endpoint 时再填 4SAPI。
适合场景：通义团队的终端编码助手。
```

```text
工具：Reasonix
官方思路：DeepSeek 原生终端编程 Agent，支持 MCP。
4SAPI 思路：优先查 base_url 是否可配置；能配置就替换为 4SAPI，不能配置则保留官方或走代理层。
适合场景：DeepSeek 原生开发流。
```

```text
工具：WorkBuddy / CodeBuddy
官方思路：支持自定义 OpenAI 兼容模型配置。
4SAPI 思路：直接填 OpenAI-compatible：Base URL、4SAPI Key、deepseek-v4-flash。
适合场景：桌面 Agent 和代码助手。
```

## 6. 最推荐的 4SAPI 配置策略

不要所有工具共用一个 Key。

尤其是 Agent 工具。

因为 Agent 一旦接入文件系统、MCP、聊天平台、自动化任务，调用量会放大。

建议这样拆：

```text
4SAPI-DeepSeekFlash-Chat
用途：聊天机器人、群助手、轻量问答。
模型：deepseek-v4-flash
```

```text
4SAPI-DeepSeekFlash-Code
用途：Cline、OpenCode、Qwen Code、Kilo Code 日常编码。
模型：deepseek-v4-flash
```

```text
4SAPI-DeepSeekFlash-Agent
用途：OpenClaw、Hermes、LobeHub 多 Agent 执行。
模型：deepseek-v4-flash
```

```text
4SAPI-DeepSeekPro-Review
用途：架构审查、复杂推理、最终 Review。
模型：deepseek-v4-pro
```

这样拆以后，你可以在 4SAPI 后台看：

```text
聊天机器人花了多少。
编码助手花了多少。
多 Agent 自动化花了多少。
复杂 Review 花了多少。
```

这个比“所有东西共用一把 Key”靠谱太多。

## 7. Pro + Flash 的 Agent 工作流

DeepSeek-V4-Flash 不应该孤立使用。

更好的方式是：

```text
Flash 做高频执行。
Pro 做关键规划和审查。
```

比如一个真实编码任务：

```text
1. deepseek-v4-pro 阅读需求，拆任务。
2. deepseek-v4-flash 执行小文件修改。
3. deepseek-v4-flash 写普通测试。
4. deepseek-v4-pro 做最终 Review。
5. deepseek-v4-flash 整理 changelog 和 PR 描述。
```

为什么这样划分？

因为 Agent 里的 token 消耗不是平均分布。

大量步骤是低风险重复工作。

如果全部用 Pro，成本会上去。

如果全部用 Flash，复杂架构任务可能不稳。

4SAPI 的意义就是让你按任务路由模型。

你可以在不同工具里设置不同默认模型：

```text
Cherry Studio：deepseek-v4-flash 做日常对话。
Cline：deepseek-v4-flash 做小改动。
OpenClaw：deepseek-v4-flash 做执行 Agent。
Claude Code / Codex：复杂任务时切到 deepseek-v4-pro 或其他强模型。
```

这不是省一两次调用的钱。

而是把整个 Agent 工作流的成本结构改掉。

## 8. 三个最小接入示例

下面给三个最常见的配置模板。

### 8.1 OpenAI-compatible 工具

适用于大多数桌面客户端、IDE 插件和开源 Agent。

```text
Provider: OpenAI Compatible
Base URL: https://4sapi.com/v1
API Key: <4SAPI_KEY>
Model: deepseek-v4-flash
```

测试 prompt：

```text
你现在使用的模型是什么？请用三句话说明你适合处理哪些开发任务。
```

再测试一个代码任务：

```text
请写一个 TypeScript 函数，把数组按指定字段分组，并补一个最小单元测试。
```

### 8.2 Claude Code / Anthropic 风格工具

适用于 Claude Code、Langcli、部分 Claude-compatible 工具。

注意重点：

```text
这些工具不一定吃 OpenAI-compatible URL。
它们可能吃 Anthropic Messages 协议。
```

所以配置时要先看工具字段：

```text
ANTHROPIC_BASE_URL
ANTHROPIC_AUTH_TOKEN
ANTHROPIC_MODEL
```

如果你按 4SAPI 教程配置 Claude Code，优先使用 4SAPI 当前文档给出的方式。

如果用 CC Switch 管 Claude Code，就把 4SAPI 供应商保存进去，再启用目标供应商。

不要把不同协议混在一起。

### 8.3 Codex / Responses API 工具

DeepSeek 官方 Codex 指南里提到，Codex 使用 OpenAI Responses API，因此需要转发层处理请求。

这类工具要看两件事：

```text
Codex 当前请求协议是什么。
转发层是否能把请求转成 4SAPI 支持的模型调用。
```

如果你使用 Moon Bridge 或类似转发层，上游供应商可以改成 4SAPI：

```text
upstream_base_url: https://4sapi.com/v1
upstream_api_key: <4SAPI_KEY>
model: deepseek-v4-flash
```

如果你按 4SAPI 的 GPT-codex 教程配置，就直接遵循它的字段。

关键不是死背某一个配置文件。

关键是理解：

```text
Codex 类工具可能需要协议转换。
4SAPI 负责模型入口。
转发层负责协议适配。
```

## 9. 常见错误

第一，Base URL 拼错。

常见错误：

```text
https://4sapi.com/v1/v1
https://4sapi.com/v1/chat/completions/chat/completions
https://4sapi.com/chat/completions
```

优先用：

```text
https://4sapi.com/v1
```

如果工具自己会拼 `/v1`，再用：

```text
https://4sapi.com
```

第二，模型名写错。

模型名从 4SAPI 模型广场复制。

不要把官方文档里的变体名直接混用。

第三，把 OpenAI 协议和 Anthropic 协议混了。

表现通常是：

```text
400
messages 格式错误
缺少 anthropic-version
unknown endpoint
model not found
```

看到这种错误，先确认工具到底在发什么协议。

第四，用内置 DeepSeek Provider 却填 4SAPI Key。

如果工具的 DeepSeek Provider endpoint 写死在官方地址，你填 4SAPI Key 当然不能通。

这时要换 Custom Provider。

第五，不看日志。

4SAPI 后台能看调用记录。

如果 4SAPI 完全没有日志，说明请求没有打到 4SAPI。

先查工具配置。

如果有日志但报模型错误，再查模型名和权限。

第六，所有 Agent 共用一把无限额 Key。

这会让成本治理失控。

聊天机器人、编码助手、自动化 Agent、测试环境，应该拆 Key。

## 10. 最小检查清单

接入前按这张表过一遍：

```text
[ ] 已在 4SAPI 后台创建专用 Key
[ ] 已确认 Key 所在分组可调用 deepseek-v4-flash
[ ] 模型名从 4SAPI 模型广场复制
[ ] 工具支持自定义 Base URL 或 OpenAI-compatible
[ ] Base URL 使用 https://4sapi.com/v1 或 https://4sapi.com
[ ] API Key 使用 4SAPI Key，不是官方 DeepSeek Key
[ ] 已区分 OpenAI-compatible / Anthropic-compatible / Responses API
[ ] 已做一个简单问答测试
[ ] 已做一个小代码任务测试
[ ] 4SAPI 后台能看到调用日志
[ ] 按 Chat / Code / Agent / Review 拆 Key
[ ] 团队场景已设置预算和权限
```

## 11. 最后总结

DeepSeek 官方的 `awesome-deepseek-agent` 给了一个很重要的信号：

```text
DeepSeek-V4-Pro 和 DeepSeek-V4-Flash 已经不是单纯聊天模型。
它们正在进入 Agent 工具生态。
```

从 AstrBot、Cherry Studio、Claude Code、Cline、Codex，到 OpenClaw、OpenCode、Qwen Code、WorkBuddy，这些工具本质上都在问同一个问题：

```text
我能不能接一个稳定、便宜、可控的模型入口？
```

4SAPI 的 deepseek-v4-flash 正适合放在这个位置。

个人用户可以先把它接到一个桌面客户端或一个编码插件里。

团队用户建议直接按用途拆 Key：

```text
聊天
编码
Agent
Review
测试
生产
```

一句话：

```text
官方清单证明 DeepSeek V4 Flash 能进 Agent 生态。
4SAPI 让它更适合团队长期使用。
```

不要让每个工具各管一把 Key。

把 deepseek-v4-flash 放进 4SAPI，再接到你的 Agent 全家桶里。

这才是省钱、可查、能复盘的接入方式。

## 资料来源与延伸阅读

- DeepSeek 官方 Awesome DeepSeek Agent 中文 README：https://github.com/deepseek-ai/awesome-deepseek-agent/blob/main/README.zh-CN.md
- DeepSeek 官方 Claude Code 接入指南：https://github.com/deepseek-ai/awesome-deepseek-agent/blob/main/docs/claude_code.zh-CN.md
- DeepSeek 官方 Codex 接入指南：https://github.com/deepseek-ai/awesome-deepseek-agent/blob/main/docs/codex.zh-CN.md
- DeepSeek 官方 OpenClaw 接入指南：https://github.com/deepseek-ai/awesome-deepseek-agent/blob/main/docs/openclaw.zh-CN.md
- DeepSeek 官方 Qwen Code 接入指南：https://github.com/deepseek-ai/awesome-deepseek-agent/blob/main/docs/qwen_code.zh-CN.md
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 文档：https://4sapi.apifox.cn/
- 4SAPI 配置模型 URL 说明：https://4sapi.apifox.cn/8423139m0
- 4SAPI GPT-codex 安装教程：https://4sapi.apifox.cn/347639c0
- 4SAPI Claude Code CLI 安装与配置：https://4sapi.apifox.cn/347624c0
- 4SAPI OpenCode 接入配置：https://4sapi.apifox.cn/8323429m0
- 4SAPI 模型价格页：https://blog.4sapi.com/pricing
