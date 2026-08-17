---
title: "【大模型API中转站】第91期 CC Switch配置GPT与Codex | 4SAPI统一OpenAI入口"
category: 人工智能
tags:
  - 大模型API中转站
  - CC Switch
  - GPT
  - Codex
  - OpenAI Compatible
  - 4SAPI
description: "聚焦 CC Switch 中的 GPT 与 Codex 配置：如何用 4SAPI 的 OpenAI-compatible Base URL 和 Key 统一接入 GPT/Codex 类模型，处理 Chat Completions 路由、模型名、用量统计和常见排错。"
---

# 【大模型API中转站】第91期 CC Switch配置GPT与Codex | 4SAPI统一OpenAI入口

本文是【大模型API中转站】系列的第91篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

Claude 适合深度代码任务。

但 GPT / Codex 这条线也很重要。

很多团队会这样分工：

```text
Claude Code：复杂改造。
Codex：云端任务、GitHub、批量修复。
GPT：通用问答、结构化输出、工具流。
```

问题在于：

```text
OpenAI-compatible 工具很多。
但每个工具配置入口不一样。
```

CC Switch 的价值，是把 Codex 和 GPT 类供应商统一管理起来。

4SAPI 的价值，是把 OpenAI-compatible 模型入口统一成：

```text
https://4sapi.com/v1
```

这篇专门讲 GPT / Codex。

## 1. GPT / Codex 为什么适合走统一入口

GPT 和 Codex 类模型常见用途：

```text
代码生成
单元测试
PR 描述
Issue 总结
JSON 结构化输出
日志归纳
批量脚本生成
```

这些任务不一定都需要最强模型。

有些可以用低成本模型。

有些需要 GPT 级别的稳定结构化能力。

有些需要 Codex 工作流能力。

如果你把 Key 分散在每个工具里，后面很难治理。

通过 4SAPI 统一入口后，可以按任务选模型：

```text
日常编码：GPT/Codex 主力模型。
摘要归纳：低成本模型。
结构化输出：稳定 JSON 模型。
复杂推理：高能力模型。
```

CC Switch 再负责把这些供应商写到 Codex 和其他工具里。

## 2. 准备 4SAPI OpenAI-compatible 配置

常见配置：

```text
Base URL: https://4sapi.com/v1
API Key: <4SAPI_KEY>
Model: 从 4SAPI 模型广场复制 GPT 或 Codex 模型名
```

如果使用 SDK，最终请求通常会拼成：

```text
https://4sapi.com/v1/chat/completions
```

不要在 Base URL 里重复写完整路径。

除非工具明确要求完整端点。

CC Switch 里如果是 OpenAI-compatible 供应商，通常先填 Base URL。

## 3. 在 CC Switch 添加 GPT 供应商

打开 CC Switch：

```text
添加供应商
→ 选择 OpenAI / GPT 类预设，或自定义供应商
→ 填 Base URL
→ 填 API Key
→ 填模型名
→ 保存
```

建议命名：

```text
4SAPI-GPT-Main
4SAPI-GPT-Cheap
4SAPI-GPT-JSON
4SAPI-GPT-Review
```

不要只建一个 GPT。

模型用途不同，成本不同。

最好分开。

## 4. 在 CC Switch 添加 Codex 供应商

CC Switch 官方文档提到，v3.16.0 亮点之一是：

```text
Codex Chat Completions 路由。
DeepSeek、Kimi、GLM、MiniMax 等仅支持 Chat 协议的供应商可通过 Codex 使用。
```

这对 4SAPI 很关键。

因为 4SAPI 后面可以接多种 OpenAI-compatible 模型。

你可以让 Codex 通过 4SAPI 使用：

```text
GPT 类模型
Claude 类模型
Gemini 类模型
国产 Coding 模型
```

在 CC Switch 里建议建：

```text
4SAPI-Codex-GPT
4SAPI-Codex-Coding
4SAPI-Codex-Review
```

如果模型只支持 Chat Completions，就注意使用 CC Switch 文档里对应的 Codex 路由能力。

## 5. 通用供应商同步

如果你希望 GPT/Codex 供应商同时给多个工具用，可以使用通用供应商。

例如：

```text
Provider: 4SAPI-Universal-GPT
Base URL: https://4sapi.com/v1
Key: <4SAPI_KEY>
Model: <GPT模型名>
同步：Claude Code / Codex / Gemini CLI
```

但我的建议是：

```text
初期用通用供应商跑通。
稳定后按工具拆专用供应商。
```

因为 Codex 的任务模式和 Gemini CLI 不一样。

同一个模型未必适合所有工具。

## 6. Codex 场景怎么选模型

Codex 常见任务：

```text
修 Bug
写测试
重构
批量改文件
生成 PR 说明
分析 GitHub diff
```

建议：

```text
小修小补：低成本 Coding 模型。
复杂修复：GPT/Codex 主力模型。
PR Review：擅长长上下文和结构化输出的模型。
测试生成：稳定代码模型。
总结类任务：低成本模型。
```

用 4SAPI 的好处是，你可以先跑一周，看账单和成功率。

然后调整默认模型。

不要凭感觉选。

## 7. URL 拼接排错

GPT/Codex 接入最常见的坑，就是 URL 拼接。

正确 Base URL：

```text
https://4sapi.com/v1
```

常见错误：

```text
https://4sapi.com/v1/v1/chat/completions
https://4sapi.com/v1/chat/completions/chat/completions
https://4sapi.com/chat/completions
```

排查方法：

```text
1. 看工具要填 Base URL 还是完整 URL。
2. Base URL 场景填 https://4sapi.com/v1。
3. 完整 URL 场景才填 /chat/completions。
4. 用最小请求测试。
5. 到 4SAPI 后台看是否有请求记录。
```

如果 4SAPI 后台完全没有请求，通常是工具没打到网关。

如果有请求但失败，看错误码。

## 8. Key 和模型名排错

第二类常见问题是 Key 和模型名。

检查：

```text
Key 是否复制完整。
Key 是否属于当前项目。
Key 是否有目标模型权限。
模型名是否完整复制。
余额是否充足。
是否触发限流。
```

建议每个供应商建好后都做一次模型检查。

CC Switch 有模型检查和健康检测相关功能。

4SAPI 后台也能看调用错误。

两个一起看，排错会快很多。

## 9. 用量统计怎么用

GPT/Codex 类任务很容易变成高频调用。

尤其是：

```text
批量修复
自动测试生成
PR 分析
日志总结
CI Agent
```

建议在 CC Switch 里看本地用量。

在 4SAPI 看网关账单。

每周统计：

```text
Codex 调用总量。
GPT 主力模型费用。
低成本模型覆盖率。
失败请求占比。
平均任务成本。
```

如果团队已经有多个 Key，最好按 Key 分账。

不要所有 Codex 任务共用一个 Key。

## 10. 团队推荐配置

个人：

```text
4SAPI-GPT-Main
4SAPI-Codex-Coding
```

小团队：

```text
4SAPI-Codex-Dev
4SAPI-Codex-Review
4SAPI-GPT-Summary
4SAPI-GPT-JSON
```

企业：

```text
4SAPI-Codex-TeamA
4SAPI-Codex-TeamB
4SAPI-Codex-CI
4SAPI-GPT-Report
4SAPI-GPT-LowCost
```

命名清楚，账单才清楚。

## 11. 最小检查清单

```text
[ ] 已准备 4SAPI Key
[ ] Base URL 使用 https://4sapi.com/v1
[ ] 已复制 GPT/Codex 模型名
[ ] 已在 CC Switch 添加 GPT 供应商
[ ] 已在 CC Switch 添加 Codex 供应商
[ ] 已确认是否需要 Chat Completions 路由
[ ] 已启用供应商
[ ] 已重启 Codex 或对应 CLI
[ ] 已跑简单请求
[ ] 4SAPI 后台有调用记录
[ ] CC Switch 用量统计有数据
```

## 12. 最后总结

GPT / Codex 的优势是通用、稳定、生态成熟。

但如果每个工具都单独配置，很快就会乱。

更好的做法是：

```text
4SAPI 统一 OpenAI-compatible 模型入口。
CC Switch 统一 GPT / Codex 供应商配置。
```

个人用户先建两个供应商就够：

```text
4SAPI-GPT-Main
4SAPI-Codex-Coding
```

团队用户建议按任务拆：

```text
开发、Review、总结、JSON、CI。
```

一句话：

```text
GPT/Codex 负责高频生产力，4SAPI 负责把高频调用算清楚。
```

## 资料来源与延伸阅读

- CC Switch 中文 README：https://github.com/farion1231/cc-switch/blob/main/README_ZH.md
- CC Switch 用户手册：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/README.md
- CC Switch 添加供应商：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/2-providers/2.1-add.md
- CC Switch 模型检查：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/4-proxy/4.5-model-test.md
- CC Switch 用量统计：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/4-proxy/4.4-usage.md
- CC Switch v3.16.0 发布说明：https://github.com/farion1231/cc-switch/blob/main/docs/release-notes/v3.16.0-en.md
- 4SAPI 接入文档：https://4sapi.apifox.cn/
- 4SAPI 官网：https://4sapi.com/
