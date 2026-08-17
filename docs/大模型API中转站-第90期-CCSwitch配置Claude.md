---
title: "【大模型API中转站】第90期 CC Switch配置Claude | 4SAPI Key切Sonnet和Opus"
category: 人工智能
tags:
  - 大模型API中转站
  - CC Switch
  - Claude Code
  - Claude Desktop
  - Claude
  - 4SAPI
description: "面向 Claude Code 和 Claude Desktop 用户，讲解如何在 CC Switch 中配置 4SAPI 供应商，用一个 4SAPI Key 管理 Claude Sonnet、Opus、Haiku 等模型，并处理热切换、官方登录回退和成本统计。"
---

# 【大模型API中转站】第90期 CC Switch配置Claude | 4SAPI Key切Sonnet和Opus

本文是【大模型API中转站】系列的第90篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一篇我们讲了 CC Switch + 4SAPI 的总览。

这一篇专门讲 Claude。

因为在所有编码 Agent 里，Claude Code 和 Claude Desktop 的组合非常常见。

一个做终端深度开发。

一个做桌面 MCP 和资料工作台。

但 Claude 系列模型也有两个现实问题：

```text
模型强，但成本高。
工具多，但配置散。
```

很多人一开始会这样用：

```text
Claude Code 里一套 Key。
Claude Desktop 里一套配置。
备用模型另存一个文件。
团队成员各配各的。
```

很快就乱。

更好的方式是：

```text
CC Switch 管 Claude 工具配置。
4SAPI 管 Claude 模型入口和成本。
```

这篇会讲：

```text
怎么在 CC Switch 里添加 4SAPI Claude 供应商。
怎么同步到 Claude Code / Claude Desktop。
怎么切 Sonnet / Opus / Haiku。
怎么回到官方登录。
怎么排查 URL、Key、模型名问题。
```

## 1. Claude 为什么适合走 4SAPI

Claude 在编码任务里很强。

尤其是：

```text
长文件理解
复杂重构
多文件修改
测试失败分析
文档和代码一起读
Agent 工具调用
```

但它也很容易成为团队里的主要成本来源。

如果每个人都自己配 Key，你很难知道：

```text
谁在用 Opus。
谁在用 Sonnet。
哪个项目最费钱。
失败请求是不是因为模型无权限。
是不是有人把高成本模型用在低价值摘要上。
```

4SAPI 的价值是把 Claude 模型调用统一到一个网关。

CC Switch 的价值是把这个网关写进 Claude Code / Claude Desktop 的配置里。

组合起来：

```text
Claude 负责干活。
CC Switch 负责切换。
4SAPI 负责治理。
```

## 2. 准备模型规划

不要一上来只建一个“Claude”供应商。

建议按模型用途拆。

```text
4SAPI-Claude-Sonnet
4SAPI-Claude-Opus
4SAPI-Claude-Haiku
4SAPI-Claude-Cheap
```

用途可以这样分：

```text
Sonnet：日常编码主力。
Opus：复杂架构、困难 Bug、长任务。
Haiku：轻量总结、快速问答、格式转换。
Cheap：低成本兼容模型，处理非关键任务。
```

如果团队预算敏感，默认供应商建议放 Sonnet 或低成本模型。

Opus 不要做默认。

而是作为“高价值任务手动切换”。

## 3. 在 4SAPI 准备 Key

在 4SAPI 后台创建 Claude 专用 Key。

建议命名：

```text
ccswitch-claude-personal
ccswitch-claude-team
ccswitch-claude-review
```

准备三项：

```text
Base URL: https://4sapi.com/v1
API Key: sk-xxxx
Model: 从 4SAPI 模型广场复制 Claude 模型名
```

不要手写模型名。

不要把 Key 发群里。

不要把 Key 截到教程图里。

## 4. 在 CC Switch 添加 Claude 供应商

打开 CC Switch：

```text
添加供应商
→ 选择 Claude Code / Claude 相关预设，或自定义供应商
→ 填入 Base URL
→ 填入 API Key
→ 填入模型名
→ 保存
```

推荐字段：

```text
Provider Name: 4SAPI-Claude-Sonnet
Base URL: https://4sapi.com/v1
API Key: <4SAPI_KEY>
Model: <Claude Sonnet 模型名>
```

再建一个：

```text
Provider Name: 4SAPI-Claude-Opus
Base URL: https://4sapi.com/v1
API Key: <4SAPI_KEY>
Model: <Claude Opus 模型名>
```

这样切换时不会混。

## 5. 同步到 Claude Code

CC Switch 官方 FAQ 提到：

```text
Claude Code 目前支持供应商数据热切换，无需重启。
```

这是 Claude Code 的一个优势。

但第一次配置时，我仍然建议这样验证：

```text
1. 在 CC Switch 启用 4SAPI-Claude-Sonnet。
2. 打开 Claude Code。
3. 让它回答一个简单问题。
4. 再切到 4SAPI-Claude-Opus。
5. 重新问一个简单问题。
6. 到 4SAPI 后台看调用记录。
```

不要用复杂任务验证入口。

入口不通时，复杂任务只会增加噪音。

## 6. 同步到 Claude Desktop

Claude Desktop 更常见于 MCP 工作流。

比如：

```text
浏览器 MCP
文件系统 MCP
数据库 MCP
图表 MCP
内部工具 MCP
```

CC Switch 支持 Claude Desktop 第三方供应商配置管理，并且文档里有专门章节。

配置思路一样：

```text
选择 Claude Desktop
添加 4SAPI Claude 供应商
启用
重启 Claude Desktop
验证模型调用
```

注意：

```text
Claude Desktop 的 MCP 配置和模型供应商配置不要混。
```

MCP 是工具。

4SAPI 是模型入口。

CC Switch 可以一起管理，但你要知道它们是两层。

## 7. 官方登录怎么保留

CC Switch FAQ 提到：

```text
可以添加“官方登录”预设。
切换过去后执行 Log out / Log in。
之后可以在官方供应商和第三方供应商之间切换。
```

这点很实用。

不要把官方登录删掉。

建议保留：

```text
Claude-Official
4SAPI-Claude-Sonnet
4SAPI-Claude-Opus
4SAPI-Claude-Haiku
```

这样你可以随时回退。

尤其是排错时：

```text
官方能用，4SAPI 不能用 → 查 4SAPI Key / URL / 模型权限。
官方也不能用 → 查本地工具 / 登录 / 网络。
```

## 8. Claude 模型怎么选

建议按任务选。

```text
普通代码修改：Sonnet。
复杂架构设计：Opus。
大段文档总结：Sonnet 或长上下文模型。
格式转换和短摘要：Haiku 或低成本模型。
MCP 工具操作：优先稳定模型。
```

如果通过 4SAPI 做团队网关，可以在后台看模型消耗。

每周复盘一次：

```text
Opus 使用是否过多。
Sonnet 是否能覆盖大多数任务。
低成本模型是否可承担摘要和分类。
失败请求是否集中在某个模型。
```

Claude 很强，但不等于每个任务都要用最贵模型。

## 9. 用量和成本怎么管

CC Switch 有用量仪表盘。

4SAPI 有网关账单和日志。

建议两边都看。

本地看：

```text
哪个供应商被频繁切换。
请求量是否突然上升。
某个模型是否异常高频。
```

4SAPI 看：

```text
哪个 Key 花钱。
哪个 Claude 模型花钱。
失败率是否异常。
是否触发限流。
```

团队使用时，不建议所有人共用一个 Claude Key。

至少拆成：

```text
个人开发 Key
团队评审 Key
CI / 自动化 Key
内容生产 Key
```

这样账单才看得懂。

## 10. 常见错误

第一，Base URL 错。

优先用：

```text
https://4sapi.com/v1
```

如果工具要求完整接口，再按工具规则配置。

第二，模型名错。

从 4SAPI 模型广场复制完整名称。

第三，Key 没权限。

有些模型需要分组权限或余额。

到 4SAPI 后台确认。

第四，切换后还在用旧供应商。

Claude Code 可以热切换，但仍建议看 CC Switch 当前激活状态和 4SAPI 日志。

第五，MCP 报错误以为模型报错。

先问一个不调用 MCP 的简单问题。

如果简单问题通，模型入口没问题。

再查 MCP。

## 11. 最小检查清单

```text
[ ] 已在 4SAPI 创建 Claude 专用 Key
[ ] Base URL 使用 https://4sapi.com/v1
[ ] 模型名从 4SAPI 模型广场复制
[ ] 已在 CC Switch 添加 4SAPI-Claude-Sonnet
[ ] 已在 CC Switch 添加 4SAPI-Claude-Opus
[ ] 已保留 Claude 官方登录供应商
[ ] 已启用目标供应商
[ ] Claude Code 简单问答成功
[ ] Claude Desktop 简单问答成功
[ ] 4SAPI 后台能看到调用日志
[ ] CC Switch 用量仪表盘能看到请求趋势
```

## 12. 最后总结

Claude Code 和 Claude Desktop 都很适合深度工作流。

但团队使用时，不能每个人自己乱配 Key。

更稳的方式是：

```text
CC Switch 管 Claude 工具配置。
4SAPI 管 Claude 模型入口。
```

个人用户可以先建一个 `4SAPI-Claude-Sonnet`。

团队用户建议再拆出 Opus、Haiku、Review、CI 等供应商。

这样你既能快速切模型，也能在 4SAPI 后台看日志和费用。

一句话：

```text
Claude 负责强推理，4SAPI 负责控成本，CC Switch 负责少手改配置。
```

## 资料来源与延伸阅读

- CC Switch 中文 README：https://github.com/farion1231/cc-switch/blob/main/README_ZH.md
- CC Switch 用户手册：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/README.md
- CC Switch 添加供应商：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/2-providers/2.1-add.md
- CC Switch Claude Desktop：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/2-providers/2.6-claude-desktop.md
- CC Switch 切换供应商：https://github.com/farion1231/cc-switch/blob/main/docs/user-manual/zh/2-providers/2.2-switch.md
- 4SAPI 接入文档：https://4sapi.apifox.cn/
- 4SAPI 官网：https://4sapi.com/
