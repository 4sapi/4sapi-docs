---
title: "【大模型API中转站】第89期 CC Switch接入4SAPI | 一个Key切七个Agent"
category: 人工智能
tags:
  - 大模型API中转站
  - CC Switch
  - Claude Code
  - Codex
  - Gemini CLI
  - OpenClaw
  - 4SAPI
description: "从零讲清 CC Switch 如何统一管理 Claude Code、Claude Desktop、Codex、Gemini CLI、OpenCode、OpenClaw、Hermes，再通过 4SAPI 的 Key 和 Base URL 做多模型入口、日志、成本和团队权限治理。"
---

# 【大模型API中转站】第89期 CC Switch接入4SAPI | 一个Key切七个Agent

本文是【大模型API中转站】系列的第89篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

如果你同时用过 Claude Code、Codex、Gemini CLI、OpenClaw、Hermes，大概率会遇到一个很烦的问题：

```text
模型入口太散。
配置文件太多。
Key 到处都是。
每个工具切模型的方式都不一样。
```

Claude Code 有自己的供应商配置。

Codex 有自己的配置。

Gemini CLI 有自己的配置。

OpenClaw 还有工作区和 Agent 文件。

Hermes 又是一套路由和模型声明。

你只是想换一个模型。

结果要到处改：

```text
base_url
api_key
model
MCP
Skills
AGENTS.md
GEMINI.md
CLAUDE.md
```

改一次还好。

团队里每个人都改，就会变成灾难。

这篇讲一个更省事的方式：

```text
CC Switch 管工具切换。
4SAPI 管模型入口。
```

CC Switch 负责把 Claude Code、Claude Desktop、Codex、Gemini CLI、OpenCode、OpenClaw、Hermes 这些工具放在一个桌面应用里统一管理。

4SAPI 负责提供统一的 API Key、统一的 Base URL、统一的模型路由、日志和成本统计。

两者组合起来，你要做的事情就变成：

```text
在 4SAPI 后台建 Key。
在 CC Switch 添加 4SAPI 供应商。
选择要同步的工具。
点击启用。
重启对应 CLI。
开始用。
```

## 1. 为什么需要 CC Switch

现代 AI 编程工具越来越多。

常见组合是：

```text
Claude Code：深度代码任务。
Codex：云端任务、GitHub、后台修复。
Gemini CLI：长上下文和搜索类任务。
OpenClaw：更可控的 Agent 工作区。
Hermes：多模态和工具流。
Claude Desktop：MCP 和桌面工作台。
```

单个工具不难配置。

真正麻烦的是同时维护多个工具。

比如你想把默认模型从 Claude 切到 GPT。

你可能要改：

```text
Claude Code 的 provider。
Codex 的模型配置。
Gemini CLI 的 Key。
OpenClaw 的 Workspace。
Hermes 的 provider。
MCP 同步状态。
Prompt 文件。
Skills 目录。
```

如果你每次都手动改 JSON、TOML、`.env`，一定会乱。

CC Switch 的价值就是：

```text
把这些分散配置收口到一个桌面工具里。
```

官方 README 里说得很清楚：

```text
一个应用，管理七个工具。
支持 50+ 供应商预设。
支持通用供应商。
支持 MCP、Prompts、Skills。
支持系统托盘切换。
配置存 SQLite。
写入 live 文件时用原子写入。
```

这就是它适合和 4SAPI 组合的原因。

4SAPI 给你一个统一模型网关。

CC Switch 把这个网关分发到多个 Agent 工具。

## 2. 4SAPI 在这里解决什么

先说结论：

```text
CC Switch 不是模型服务商。
4SAPI 才是模型入口。
```

CC Switch 负责管理配置。

4SAPI 负责模型调用。

4SAPI 在这里解决五个问题。

第一，统一 Key。

你不用给每个工具填不同厂商 Key。

可以在 4SAPI 后台生成一个项目 Key，给 CC Switch 作为统一供应商使用。

第二，统一 Base URL。

系列文章里一直建议 OpenAI-compatible 场景优先用：

```text
https://4sapi.com/v1
```

如果某个工具要求完整端点，再按工具规则填完整路径。

但 CC Switch 里大多数供应商配置，重点就是：

```text
Base URL
API Key
Model
```

第三，统一模型路由。

你可以把 Claude、GPT、Gemini、国产模型都放在 4SAPI 后面。

上层工具只管填模型名。

第四，统一成本。

CC Switch 自己有用量统计和成本追踪能力。

4SAPI 后台也能看 Key、模型、项目维度的消耗。

两边结合，团队才知道钱花在哪里。

第五，统一排错。

模型不可用、Key 无权限、URL 拼错、限流、余额不足，这些问题都可以优先从 4SAPI 控制台和 CC Switch 模型检查一起定位。

## 3. 准备 4SAPI 信息

先在 4SAPI 后台准备三样东西：

```text
API Key
Base URL
模型名
```

常见填写：

```text
Base URL: https://4sapi.com/v1
API Key: sk-xxxx
Model: 你在 4SAPI 模型广场复制的完整模型名
```

注意不要把 Key 写进文章、截图或公开仓库。

团队使用时建议按用途拆 Key：

```text
ccswitch-personal
ccswitch-team-dev
ccswitch-content
ccswitch-ci
```

这样后面查账单很清楚。

不要全公司共用一个 Key。

省事一时，排错半天。

## 4. 安装 CC Switch

CC Switch 是 Tauri 桌面应用，支持 Windows、macOS、Linux。

官方安装方式大致是：

Windows：

```text
从 Releases 下载 MSI 或 Portable ZIP。
```

macOS：

```bash
brew install --cask cc-switch
```

或者从 Releases 下载 `.dmg`。

Linux：

```text
下载 deb、rpm 或 AppImage。
```

第一次启动时，CC Switch 可以手动导入现有 CLI 工具配置作为默认供应商。

这一步很重要。

因为 CC Switch 不是要破坏你的原配置。

它的设计原则是“最小侵入性”。

你可以先导入当前可用配置，再新增 4SAPI 供应商。

## 5. 添加 4SAPI 供应商

在 CC Switch 主界面：

```text
添加供应商
→ 选择预设或自定义配置
→ 填入 API Key
→ 填入 Base URL
→ 填入模型名
→ 保存
```

如果没有 4SAPI 预设，就用自定义 OpenAI-compatible 配置。

推荐命名：

```text
4SAPI-Claude
4SAPI-GPT
4SAPI-Gemini
4SAPI-Coding
4SAPI-Team
```

不要只叫 “test”。

以后你一定会忘。

字段建议：

```text
Provider Name: 4SAPI-Claude-Sonnet
Base URL: https://4sapi.com/v1
API Key: 4SAPI 后台生成的 Key
Model: 从 4SAPI 模型广场复制
```

模型名不要手打。

复制。

模型名写错，是新手最常见的坑。

## 6. 通用供应商怎么用

CC Switch 有一个很适合 4SAPI 的功能：

```text
通用供应商。
```

官方说明里提到，一份配置可以同步到 Claude Code、Codex 和 Gemini CLI。

这正好对应 4SAPI 的定位：

```text
一个模型网关，多个 Agent 使用。
```

你可以建立一个通用供应商：

```text
Name: 4SAPI-Universal
Base URL: https://4sapi.com/v1
API Key: <4SAPI_KEY>
Model: 默认 Coding 模型
```

然后同步给：

```text
Claude Code
Codex
Gemini CLI
```

如果某个工具需要特殊模型，再单独建供应商。

比如：

```text
4SAPI-Claude-Sonnet
4SAPI-GPT-Codex
4SAPI-Gemini-Pro
4SAPI-OpenClaw-Coding
```

通用供应商适合“先跑起来”。

专用供应商适合“按任务优化”。

## 7. 切换供应商

CC Switch 支持两种切换。

第一，主界面切换：

```text
选择供应商
→ 点击启用
```

第二，系统托盘切换：

```text
托盘菜单
→ 选择应用
→ 选择供应商
```

官方 FAQ 提醒：

```text
大多数工具需要重启终端或 CLI 才能生效。
Claude Code 目前支持供应商数据热切换。
```

所以切换后建议做一次最小验证：

```text
打开对应 CLI。
问一个简单问题。
确认模型调用成功。
```

不要一切换就上复杂任务。

先验证入口。

## 8. 七个工具怎么分工

CC Switch 支持七个工具。

你可以按场景分。

```text
Claude Code：复杂代码修改、重构、长任务。
Claude Desktop：MCP、桌面资料和日常问答。
Codex：云端开发、GitHub、批量修复。
Gemini CLI：长上下文、资料整理、搜索相关任务。
OpenCode：开源终端工作流。
OpenClaw：可控 Agent 工作区和工程化流程。
Hermes：多模态和团队工作流。
```

4SAPI 在下面做统一出口：

```text
Claude 模型
GPT 模型
Gemini 模型
国产 Coding 模型
Embedding / Rerank
多模态模型
```

你不一定要所有工具都用。

但只要用了两个以上，CC Switch 的价值就开始明显。

## 9. 什么时候用本地代理

CC Switch 还有代理与故障转移能力。

官方文档里提到：

```text
本地代理热切换
格式转换
自动故障转移
熔断器
供应商健康监控
应用级代理接管
```

这对团队很实用。

简单说：

```text
直连配置适合个人。
本地代理适合多供应商切换和高可用。
```

如果你只有一个 4SAPI Key，先不用代理也可以。

如果你要同时配置多个 4SAPI Key、官方 Key、备用网关，就可以用 CC Switch 的代理和故障转移能力。

但别一开始就上复杂架构。

先跑通最小接入。

再做高可用。

## 10. 用量统计怎么配合 4SAPI

CC Switch 有用量仪表盘。

可以跨供应商追踪：

```text
支出
请求数
Token 用量
趋势图
详细请求日志
自定义模型定价
```

4SAPI 后台也能看模型调用记录和费用。

这两层要一起用。

CC Switch 更靠近本地工具：

```text
哪个 Agent 在用。
切了哪个供应商。
本地请求趋势如何。
```

4SAPI 更靠近网关：

```text
哪个 Key 花钱。
哪个模型花钱。
哪个项目花钱。
哪个错误码最多。
```

团队建议每周看一次：

```text
Claude 类模型是否过度使用。
GPT 类模型是否适合低成本任务。
Gemini 是否承担长上下文任务。
失败请求是否集中在某个模型。
4SAPI Key 是否需要拆分。
```

## 11. 常见配置模板

个人开发者可以这样配：

```text
4SAPI-Default
Base URL: https://4sapi.com/v1
Key: 个人 4SAPI Key
Model: 默认 Coding 模型
同步：Claude Code / Codex / Gemini CLI
```

内容创作者可以这样配：

```text
4SAPI-Writing
Base URL: https://4sapi.com/v1
Key: 内容项目 Key
Model: 写作模型
同步：Claude Desktop / Gemini CLI / Hermes
```

开发团队可以这样配：

```text
4SAPI-Team-Coding
4SAPI-Team-Review
4SAPI-Team-Cheap-Summary
4SAPI-Team-LongContext
```

不同任务用不同模型。

不同 Key 看不同账单。

不要一个供应商包打天下。

## 12. 最容易踩的坑

第一，Base URL 拼错。

OpenAI-compatible 常见是：

```text
https://4sapi.com/v1
```

不要填成：

```text
https://4sapi.com/v1/v1/chat/completions
```

第二，模型名写错。

从 4SAPI 模型广场复制完整模型名。

第三，切换后没重启 CLI。

除了 Claude Code 热切换能力外，大多数工具最好重启终端。

第四，插件配置丢失。

CC Switch 有通用配置片段，可以把 Key 和请求地址之外的通用数据传递到新供应商。

切换前先了解“从当前供应商提取”和“写入通用配置”。

第五，不看用量。

Claude、GPT、Gemini 长任务跑起来很快烧 token。

4SAPI 和 CC Switch 的用量面板都要看。

## 13. 最小检查清单

接入前检查：

```text
[ ] 已安装 CC Switch
[ ] 已导入当前可用 CLI 配置
[ ] 已在 4SAPI 后台创建 Key
[ ] Base URL 使用 https://4sapi.com/v1
[ ] 模型名从 4SAPI 模型广场复制
[ ] 已新建 4SAPI 供应商
[ ] 已确认同步到目标工具
[ ] 已点击启用
[ ] 已重启对应 CLI
[ ] 已跑简单对话验证
[ ] 已打开用量统计
[ ] 团队场景已按项目拆 Key
```

跑完这张表，再开始写代码。

## 14. 最后总结

CC Switch 解决的是：

```text
多个 Agent 工具配置太散。
```

4SAPI 解决的是：

```text
多个模型入口太散。
```

两者组合起来，就是：

```text
一个桌面工具切换多个 Agent。
一个 API 网关管理多个模型。
```

个人用户可以先建一个 4SAPI 通用供应商，快速同步到 Claude Code、Codex、Gemini CLI。

团队用户建议按项目、按模型、按用途拆供应商和 Key。

这样你既能享受 CC Switch 的一键切换，也能用 4SAPI 把模型成本、日志和权限管起来。

一句话：

```text
别再手改七套配置。
把工具交给 CC Switch，把模型交给 4SAPI。
```

## 资料来源与延伸阅读

- CC Switch 中文 README：https://github.com/farion1231/cc-switch/blob/main/README_ZH.md
- CC Switch 用户手册：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/README.md
- CC Switch 添加供应商：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/2-providers/2.1-add.md
- CC Switch 切换供应商：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/2-providers/2.2-switch.md
- CC Switch 用量统计：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/4-proxy/4.4-usage.md
- 4SAPI 接入文档：https://4sapi.apifox.cn/
- 4SAPI 官网：https://4sapi.com/
