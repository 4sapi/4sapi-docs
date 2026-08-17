---
title: "【大模型API中转站】第93期 CC Switch配置OpenClaw | Agent工作区接入4SAPI"
category: 人工智能
tags:
  - 大模型API中转站
  - CC Switch
  - OpenClaw
  - Agent工作区
  - AGENTS.md
  - 4SAPI
description: "讲解如何用 CC Switch 管理 OpenClaw 供应商、工作区文件和 Agent 配置，再通过 4SAPI 统一 OpenClaw 的模型入口、Key、成本、日志和团队权限。"
---

# 【大模型API中转站】第93期 CC Switch配置OpenClaw | Agent工作区接入4SAPI

本文是【大模型API中转站】系列的第93篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

前几篇分别讲了 Claude、GPT/Codex、Gemini。

这一篇讲 OpenClaw。

OpenClaw 更偏 Agent 工作区。

它不只是换模型。

还会涉及：

```text
AGENTS.md
SOUL.md
工作区文件
每日记忆
Agent 配置
模型供应商
MCP / Skills
```

所以 OpenClaw 接入 4SAPI，不能只看 Key 和 URL。

还要看工作区怎么管理。

CC Switch 的价值就在这里：

```text
既能管供应商。
也能管 OpenClaw 工作区文件。
还能统一 MCP、Prompts、Skills。
```

4SAPI 则负责底层模型入口。

## 1. OpenClaw 适合什么人

如果你只是想问答，Claude Desktop 或 Gemini CLI 就够。

如果你只是想改代码，Claude Code 或 Codex 也够。

OpenClaw 更适合喜欢可控 Agent 工作流的人。

比如：

```text
想明确 Agent 的行为规则。
想维护 AGENTS.md。
想维护 SOUL.md。
想把团队 SOP 写进工作区。
想做多 Agent 编排。
想把模型入口统一到一个网关。
```

这时 OpenClaw + CC Switch + 4SAPI 的组合很自然。

```text
OpenClaw：Agent 工作区。
CC Switch：配置和文件管理。
4SAPI：模型网关。
```

## 2. 先准备 4SAPI 供应商

在 4SAPI 后台创建 OpenClaw 专用 Key。

建议命名：

```text
ccswitch-openclaw-dev
ccswitch-openclaw-team
ccswitch-openclaw-content
ccswitch-openclaw-ci
```

准备：

```text
Base URL: https://4sapi.com/v1
API Key: <4SAPI_KEY>
Model: 从 4SAPI 模型广场复制
```

OpenClaw 任务通常会长时间运行。

建议不要用个人 Key 混团队任务。

否则成本很难分。

## 3. 在 CC Switch 添加 OpenClaw 供应商

打开 CC Switch：

```text
添加供应商
→ 选择 OpenClaw 相关配置或自定义供应商
→ 填入 4SAPI Base URL
→ 填入 4SAPI Key
→ 填入模型名
→ 保存
```

建议供应商：

```text
4SAPI-OpenClaw-Coding
4SAPI-OpenClaw-Planning
4SAPI-OpenClaw-Review
4SAPI-OpenClaw-Cheap
```

OpenClaw 是 Agent 工作区，不同任务差异大。

规划、编码、Review、总结最好分开。

## 4. 工作区文件是什么

CC Switch 用户手册里有“工作区文件与每日记忆”章节。

README 里也提到：

```text
Workspace editor（OpenClaw）
编辑 Agent 文件（AGENTS.md、SOUL.md 等）
支持 Markdown 预览
```

这说明 CC Switch 不只是切模型。

它还能帮你管理 OpenClaw 的 Agent 文件。

常见文件：

```text
AGENTS.md：项目规则、执行流程、约束。
SOUL.md：Agent 的角色、偏好、长期风格。
每日记忆：当天工作记录、任务上下文。
```

这和 4SAPI 的关系是：

```text
AGENTS.md / SOUL.md 决定 Agent 怎么做事。
4SAPI 决定 Agent 调哪个模型、花多少钱、怎么审计。
```

两个层面都要管。

## 5. OpenClaw 供应商怎么分层

推荐四层。

第一，规划模型。

```text
4SAPI-OpenClaw-Planning
用途：需求拆解、任务计划、架构判断。
```

第二，编码模型。

```text
4SAPI-OpenClaw-Coding
用途：实现、修改文件、生成测试。
```

第三，Review 模型。

```text
4SAPI-OpenClaw-Review
用途：检查风险、发现遗漏、总结变更。
```

第四，低成本模型。

```text
4SAPI-OpenClaw-Cheap
用途：摘要、分类、格式整理。
```

这样拆以后，你可以在 4SAPI 后台看：

```text
规划花多少钱。
编码花多少钱。
Review 花多少钱。
低成本任务是否足够便宜。
```

## 6. AGENTS.md 怎么写

OpenClaw 工作区里，`AGENTS.md` 应该写项目规则。

比如：

```text
项目怎么启动。
测试怎么跑。
哪些目录不要动。
PR 描述格式。
安全和合规边界。
4SAPI Key 不得写入仓库。
模型调用走 CC Switch 当前供应商。
```

不要把 API Key 写进去。

可以写：

```text
模型入口由 CC Switch + 4SAPI 管理。
Base URL 使用 https://4sapi.com/v1。
实际 Key 只保存在 CC Switch / 4SAPI 配置中。
```

这样 Agent 知道规则。

但不会泄露密钥。

## 7. SOUL.md 怎么写

`SOUL.md` 更偏 Agent 风格和长期偏好。

可以写：

```text
先读文件再改。
改动保持最小。
不要主动重构无关代码。
先给计划再执行高风险动作。
遇到模型报错先查 4SAPI 日志。
输出结果要附验证命令。
```

不要写：

```text
具体 Key。
账号密码。
内部 Token。
客户隐私。
```

SOUL 是行为记忆。

不是密钥仓库。

## 8. MCP / Skills 怎么和 OpenClaw 配合

CC Switch 支持统一 MCP、Prompts、Skills 管理。

OpenClaw 工作区适合把这些组合起来：

```text
MCP：工具入口。
Prompts：常用提示词。
Skills：可复用流程。
OpenClaw Workspace：Agent 长期规则。
4SAPI：模型调用网关。
```

建议顺序：

```text
先接 4SAPI 模型。
再接 MCP。
再沉淀 Prompts。
最后整理 Skills。
```

不要同时全开。

否则排错时不知道是哪一层出问题。

## 9. 模型检查和流式检测

CC Switch v3.13.0 发布说明提到，Stream Check 面板扩展覆盖了 OpenCode / OpenClaw 场景。

这对 OpenClaw 很重要。

Agent 工具不只要一次性返回。

还要流式稳定。

配置 4SAPI 后，建议做：

```text
简单模型检查。
流式输出检查。
长任务检查。
工具调用检查。
```

如果短问答成功，但长任务断流，就要查：

```text
模型是否支持流式。
4SAPI 日志是否有超时。
CC Switch 代理是否启用。
OpenClaw 当前供应商是否正确。
```

## 10. 成本治理

OpenClaw 很容易变成成本黑洞。

因为它往往不是一问一答。

它可能会：

```text
读很多文件。
规划很多步骤。
反复调用工具。
生成长报告。
写每日记忆。
调用多个 Agent。
```

所以 OpenClaw 一定要配成本治理。

建议：

```text
规划阶段用中成本模型。
编码阶段用主力模型。
Review 阶段只对关键 diff 用强模型。
摘要阶段用低成本模型。
CI 自动任务单独 Key。
```

4SAPI 后台要按 Key 和模型看费用。

CC Switch 本地用量也要看趋势。

## 11. 常见错误

第一，OpenClaw 还在旧供应商。

切换后重启或确认当前激活供应商。

第二，工作区文件写了敏感信息。

AGENTS.md / SOUL.md 只写规则，不写 Key。

第三，模型不支持流式。

用 CC Switch 模型检查和 4SAPI 日志一起看。

第四，MCP 错误被误判成模型错误。

先禁用 MCP，验证基础模型问答。

第五，成本突然变高。

查是否 Agent 循环、是否长上下文重复发送、是否 Review 全量仓库。

## 12. 团队推荐配置

小团队：

```text
4SAPI-OpenClaw-Coding
4SAPI-OpenClaw-Review
4SAPI-OpenClaw-Summary
```

中型团队：

```text
4SAPI-OpenClaw-TeamA
4SAPI-OpenClaw-TeamB
4SAPI-OpenClaw-CI
4SAPI-OpenClaw-Content
```

企业：

```text
按部门拆 Key。
按项目拆供应商。
按模型做预算。
按任务类型做成本上限。
```

OpenClaw 越像团队工作台，越要早做治理。

## 13. 最小检查清单

```text
[ ] 已在 4SAPI 创建 OpenClaw 专用 Key
[ ] Base URL 使用 https://4sapi.com/v1
[ ] 模型名从 4SAPI 模型广场复制
[ ] 已在 CC Switch 添加 OpenClaw 供应商
[ ] 已启用供应商
[ ] 已验证基础问答
[ ] 已验证流式输出
[ ] 已检查 AGENTS.md 不含 Key
[ ] 已检查 SOUL.md 不含隐私
[ ] 已确认 MCP 不影响基础模型调用
[ ] 4SAPI 后台能看到日志
[ ] CC Switch 用量统计能看到趋势
```

## 14. 最后总结

OpenClaw 和 Claude Code、Codex、Gemini CLI 不一样。

它更像一个 Agent 工作区。

所以接入 4SAPI 时，不要只看模型。

还要看：

```text
AGENTS.md
SOUL.md
每日记忆
MCP
Skills
工作区规则
成本治理
```

CC Switch 负责把这些配置集中起来。

4SAPI 负责把模型入口、Key、日志和费用管起来。

一句话：

```text
OpenClaw 负责让 Agent 有工作区。
CC Switch 负责让工作区好切换。
4SAPI 负责让模型调用可控。
```

## 资料来源与延伸阅读

- CC Switch 中文 README：https://github.com/farion1231/cc-switch/blob/main/README_ZH.md
- CC Switch 用户手册：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/README.md
- CC Switch 添加供应商：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/2-providers/2.1-add.md
- CC Switch 工作区文件与每日记忆：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/3-extensions/3.5-workspace.md
- CC Switch 模型检查：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/4-proxy/4.5-model-test.md
- CC Switch v3.13.0 发布说明：https://github.com/farion1231/cc-switch/blob/main/docs/release-notes/v3.13.0-en.md
- 4SAPI 接入文档：https://4sapi.apifox.cn/
- 4SAPI 官网：https://4sapi.com/
