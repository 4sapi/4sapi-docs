---
title: "【大模型API中转站】第92期 CC Switch配置Gemini | 4SAPI统一长上下文入口"
category: 人工智能
tags:
  - 大模型API中转站
  - CC Switch
  - Gemini CLI
  - Gemini
  - 多模态
  - 4SAPI
description: "讲解如何用 CC Switch 管理 Gemini CLI 供应商，并通过 4SAPI 统一接入 Gemini 与其他长上下文、多模态模型，用于资料整理、代码阅读、搜索增强和低成本模型切换。"
---

# 【大模型API中转站】第92期 CC Switch配置Gemini | 4SAPI统一长上下文入口

本文是【大模型API中转站】系列的第92篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

Claude 适合深度代码。

GPT / Codex 适合高频生产。

Gemini 更常被拿来做：

```text
长上下文阅读
资料整理
多模态理解
跨文件问答
搜索增强
低成本大段总结
```

如果你已经在用 Gemini CLI，又同时用 Claude Code、Codex、OpenClaw，那么配置会越来越分散。

这篇讲：

```text
用 CC Switch 管 Gemini CLI。
用 4SAPI 管 Gemini 模型入口。
```

## 1. Gemini 适合什么任务

Gemini 的优势不一定是所有任务都最强。

但它很适合这几类：

```text
长文档阅读
代码仓库问答
多模态资料理解
会议纪要和资料总结
Research 型任务
把大量上下文压缩成结构化输出
```

所以 Gemini CLI 更像一个“资料和上下文助手”。

如果你的工作流是：

```text
Gemini 先读资料。
Claude Code 做复杂修改。
Codex 批量修复。
OpenClaw 编排 Agent。
```

那 Gemini 的模型配置也应该进统一管理。

这就是 CC Switch 的位置。

## 2. 4SAPI 统一 Gemini 入口

如果你直接给 Gemini CLI 配官方 Key，可以跑。

但团队里会遇到：

```text
Key 分散。
账单分散。
模型切换麻烦。
无法统一统计。
```

通过 4SAPI，可以把 Gemini 作为统一模型网关里的一个模型族。

常见配置：

```text
Base URL: https://4sapi.com/v1
API Key: <4SAPI_KEY>
Model: 从 4SAPI 模型广场复制 Gemini 模型名
```

如果你还想切到其他长上下文模型，也可以在 4SAPI 里统一路由。

上层 Gemini CLI 不需要知道后面是哪家模型。

## 3. 在 CC Switch 添加 Gemini 供应商

打开 CC Switch：

```text
添加供应商
→ 选择 Gemini CLI 相关预设或自定义供应商
→ 填入 4SAPI Base URL
→ 填入 4SAPI Key
→ 填入 Gemini 模型名
→ 保存
```

建议供应商命名：

```text
4SAPI-Gemini-Pro
4SAPI-Gemini-LongContext
4SAPI-Gemini-Cheap
4SAPI-Gemini-Multimodal
```

不要所有 Gemini 任务都用一个模型。

长上下文、普通总结、多模态理解，成本和能力都不一样。

## 4. 通用供应商还是专用供应商

CC Switch 支持通用供应商，一份配置可同步到多个工具。

但 Gemini 我建议分两种。

如果你只是个人试用：

```text
4SAPI-Universal
同步 Claude Code / Codex / Gemini CLI
```

如果是团队使用：

```text
4SAPI-Gemini-Research
4SAPI-Gemini-Docs
4SAPI-Gemini-Multimodal
```

原因很简单：

```text
Gemini 的任务常常是读大上下文。
成本结构和普通 Coding 任务不一样。
```

单独拆供应商，后面更容易看账单。

## 5. 切换后怎么验证

CC Switch 官方 FAQ 提到，大多数工具切换供应商后需要重启终端或 CLI。

Gemini CLI 建议这样验证：

```text
1. 在 CC Switch 启用 4SAPI-Gemini-Pro。
2. 重启 Gemini CLI。
3. 问一个简单问题。
4. 到 4SAPI 后台看调用记录。
5. 再跑一个长文本摘要任务。
6. 看响应速度和费用。
```

先小后大。

不要第一次就塞几十万字。

入口没通时，你会浪费很多排错时间。

## 6. Gemini 任务模板

建议准备几类供应商。

资料整理：

```text
Name: 4SAPI-Gemini-Docs
用途：PDF、文档、知识库总结。
```

代码阅读：

```text
Name: 4SAPI-Gemini-CodeRead
用途：大仓库浏览、跨文件关系梳理。
```

多模态：

```text
Name: 4SAPI-Gemini-Multimodal
用途：图片、视频帧、视觉资料理解。
```

低成本摘要：

```text
Name: 4SAPI-Gemini-Summary
用途：会议纪要、日志归纳、资料压缩。
```

这些名字看起来啰嗦。

但团队协作时很省心。

## 7. 和 Claude / Codex 怎么配合

推荐工作流：

```text
Gemini：先读资料，做大上下文摘要。
Claude Code：根据摘要做复杂代码修改。
Codex：批量实现、测试、PR。
OpenClaw：编排多 Agent 流程。
4SAPI：统一模型入口和成本。
CC Switch：统一切供应商。
```

不要让一个模型干所有事。

Gemini 适合吃上下文。

Claude 适合做复杂推理。

Codex 适合工程执行。

4SAPI 负责把这些模型放到同一个入口。

CC Switch 负责把这些入口写到各个工具。

## 8. 成本治理重点

Gemini 类任务最大成本风险是：

```text
上下文太长。
资料太多。
重复总结。
多模态输入过大。
```

建议：

```text
长文档先切片。
摘要任务用低成本模型。
关键推理再切强模型。
多模态任务单独 Key。
按 project_id 或团队拆 4SAPI Key。
```

如果 4SAPI 后台显示 Gemini 调用突然暴涨，要先看：

```text
是不是某个资料 Agent 循环跑。
是不是把重复文件反复喂进去。
是不是上下文没有压缩。
是不是多模态输入过大。
```

## 9. 常见错误

第一，切换后 Gemini CLI 还在旧配置。

重启终端或 CLI。

第二，Base URL 错。

OpenAI-compatible 场景优先：

```text
https://4sapi.com/v1
```

第三，模型名错。

从 4SAPI 模型广场复制。

第四，多模态模型和文本模型混用。

如果要处理图片、音频、视频，确认模型支持对应能力。

第五，只看 CC Switch，不看 4SAPI 后台。

CC Switch 看本地供应商和用量，4SAPI 看网关日志和费用，两边都要看。

## 10. 最小检查清单

```text
[ ] 已在 4SAPI 创建 Gemini 专用 Key
[ ] Base URL 使用 https://4sapi.com/v1
[ ] 模型名从 4SAPI 模型广场复制
[ ] 已在 CC Switch 添加 Gemini CLI 供应商
[ ] 已启用供应商
[ ] 已重启 Gemini CLI
[ ] 简单问答成功
[ ] 长文本摘要成功
[ ] 4SAPI 后台有调用日志
[ ] 已区分长上下文、多模态、低成本摘要供应商
```

## 11. 最后总结

Gemini 最适合放在 Agent 工作流里的“长上下文和资料处理层”。

但只要进入团队，就不能让每个人自己配 Key。

更稳的做法：

```text
CC Switch 管 Gemini CLI 配置。
4SAPI 管 Gemini 模型入口和成本。
```

个人用户可以先建一个 `4SAPI-Gemini-Pro`。

团队用户建议拆成 Docs、CodeRead、Multimodal、Summary。

一句话：

```text
Gemini 负责吃上下文，4SAPI 负责别让上下文吃掉预算。
```

## 资料来源与延伸阅读

- CC Switch 中文 README：https://github.com/farion1231/cc-switch/blob/main/README_ZH.md
- CC Switch 用户手册：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/README.md
- CC Switch 添加供应商：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/2-providers/2.1-add.md
- CC Switch 切换供应商：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/2-providers/2.2-switch.md
- CC Switch 用量统计：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/4-proxy/4.4-usage.md
- 4SAPI 接入文档：https://4sapi.apifox.cn/
- 4SAPI 官网：https://4sapi.com/
