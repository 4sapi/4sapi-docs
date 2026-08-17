---
title: "【大模型API中转站】第126期 OmO选型 | Ultimate还是Light"
category: 人工智能
tags:
  - 大模型API中转站
  - oh-my-openagent
  - OpenCode
  - LazyCodex
  - Codex
  - 企业级大模型接入
  - 企业API网关
  - 4SAPI
description: "从企业落地角度对比 oh-my-openagent Ultimate Edition 和 Codex Light Edition：Ultimate 适合 OpenCode 多 Agent、Team Mode、MCP、Slash Command 和深度编排；Light 适合 Codex CLI 插件化增强、rules、LSP、comment-checker 和 ultrawork。最后给出个人、研发团队和企业平台三种 4SAPI 接入选型路线。"
---

# 【大模型API中转站】第126期 OmO选型 | Ultimate还是Light

本文是【大模型API中转站】系列的第126篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

前面几篇讲了：

```text
OmO 怎么接 4SAPI。
Team Mode 怎么控成本。
模型路由和 fallback 怎么配。
```

这一篇回答一个更现实的问题：

```text
到底装 Ultimate，还是装 Light？
```

oh-my-openagent 现在有两个形态：

```text
Ultimate Edition：面向 OpenCode。
Light Edition：面向 Codex CLI，也就是 LazyCodex。
```

很多人会以为这是同一个插件的完整版和简化版。

可以这么理解，但不完全准确。

更准确地说：

```text
Ultimate 是 OpenCode 上的完整 Agent 编排系统。
Light 是 Codex CLI 插件体系里的便携增强组件。
```

一个偏“完整工作台”。

一个偏“给 Codex 加能力”。

企业选型不能只看功能数量。

要看：

```text
你团队主力用哪个 Agent。
你要不要 Team Mode。
你要不要 OpenCode 的 Slash Command 和技能系统。
你是否已经围绕 Codex 建了工作流。
你是否需要统一接 4SAPI 做 Key、日志和成本治理。
```

## 1. 一句话选型

如果你只想要一句话：

```text
深度多 Agent 编排，选 Ultimate。
已经重度使用 Codex CLI，选 Light。
两边都有人用，两个都装，但模型治理统一走 4SAPI。
```

更细一点：

```text
个人尝鲜：先装 Light 或 Ultimate 都可以，看你常用 Codex 还是 OpenCode。
研发团队：建议 Ultimate 做深度任务，Codex 保留日常和云端工作流。
企业平台：两个都支持，但统一 Key、模型、日志和预算必须交给 4SAPI。
```

不要为了功能多而选。

要为了工作流选。

## 2. Ultimate 是什么

Ultimate Edition 面向 OpenCode。

它包含 OmO 的完整能力：

```text
11 个 Agent
54+ lifecycle hooks
Team Mode
内置 MCP
Slash Commands
ultrawork
Ralph Loop / ulw-loop
Hashline 编辑
LSP
AST-Grep
技能系统
后台 Agent
模型匹配
runtime fallback
```

它更像：

```text
一个强化版 OpenCode Agent OS。
```

适合：

```text
复杂代码任务。
大型重构。
多 Agent 并行。
长期工程任务。
安全审计。
架构评审。
需要强工具链的研发团队。
```

如果你看重的是：

```text
Sisyphus 调度。
Hephaestus 深度执行。
Prometheus 规划。
Oracle 架构咨询。
Team Mode 并行。
MCP / LSP / AST-Grep 集成。
```

那你要看 Ultimate。

## 3. Light 是什么

Light Edition 面向 Codex CLI。

安装方式通常是：

```bash
npx lazycodex-ai install
```

它不会把 Codex 变成 OpenCode。

它是把一组可移植组件装进 Codex 插件体系。

常见组件包括：

```text
rules
comment-checker
git-bash
lsp
ultrawork
ulw-loop
start-work-continuation
telemetry
```

还会有 plugin-scoped MCP：

```text
grep_app
context7
codegraph
git_bash
lsp
```

它更像：

```text
给 Codex CLI 加一套纪律和工具增强。
```

适合：

```text
已经用 Codex CLI。
喜欢 Codex 原生插件体系。
想要 AGENTS.md / rules 注入。
想要 LSP 和 comment-checker。
想要 ultrawork 关键词触发。
想要在 Codex 里保留轻量增强。
```

但要注意：

```text
Light 没有 Ultimate 那套完整 OpenCode Agent 编排。
没有完整 Team Mode。
没有 Ultimate 的 Slash Command 层。
```

不要把 Light 写成 Ultimate 的全部能力。

## 4. 对比表

可以直接看这张表。

| 维度 | Ultimate / OpenCode | Light / Codex CLI |
| --- | --- | --- |
| 主体 | OpenCode 插件 | Codex CLI 插件 |
| 安装 | `bunx oh-my-openagent install` | `npx lazycodex-ai install` |
| 多 Agent 编排 | 强 | 依赖 Codex 原生能力 |
| Team Mode | 有，默认关闭 | 不是主能力 |
| Slash Commands | 有 | Codex CLI 不走这层 |
| Skills | Ultimate 内置技能系统 | 主要走 Codex 插件/组件 |
| MCP | 内置和运行时注入 | plugin-scoped MCP |
| LSP | 有 | 有 |
| Rules 注入 | 有 | 有 |
| Comment Checker | 有 | 有 |
| Hashline 编辑 | Ultimate 侧能力 | Codex 使用自身编辑能力 |
| 适合任务 | 深度工程、多 Agent、大型任务 | Codex 增强、轻量自动化、日常 CLI |
| 4SAPI 接入重点 | OpenCode Provider + OmO 模型路由 | Codex model provider 或现有 Codex 配置治理 |

一句话：

```text
Ultimate 管编排。
Light 管增强。
4SAPI 管模型治理。
```

## 5. 什么时候选 Ultimate

如果你的核心需求是：

```text
让 Agent 像一个工程团队一样工作。
```

选 Ultimate。

典型场景：

```text
大型代码库重构。
多个子任务并行排查。
安全审计和 PoC 分析。
复杂 Bug 根因定位。
从计划到执行再到 Review 的闭环。
需要 OpenCode TUI 和工具生态。
```

特别是你想用：

```text
Team Mode
Prometheus
Sisyphus
Hephaestus
Oracle
Librarian
Explore
```

那就不要绕。

直接看 Ultimate。

4SAPI 在这里负责：

```text
给不同 Agent 分配模型。
给 Team Mode 单独 Key。
记录各模型消耗。
设置预算和额度。
查看 fallback 和错误码。
```

## 6. 什么时候选 Light

如果你的团队已经用 Codex CLI 很顺手。

比如：

```text
有 Codex App。
有 Codex CLI。
有 Codex IDE Extension。
有 Codex Cloud。
已有 AGENTS.md。
已有 Plugins / MCP / Hooks / Automations。
```

那 Light 更自然。

它不会要求你换主工作台。

它是给 Codex 加：

```text
规则注入。
注释检查。
Git Bash 支持。
LSP 工具。
ultrawork / ulw-loop。
start-work continuation。
```

适合：

```text
日常开发增强。
规则一致性。
本地 CLI 任务。
轻量自动化。
已有 Codex 团队流程。
```

4SAPI 在这里主要管：

```text
Codex 自定义 provider。
不同 Codex 入口的 Key。
插件和自动化的模型调用成本。
团队预算。
日志审计。
```

## 7. 什么时候两个都装

企业里最常见的答案可能是：

```text
两个都装。
```

但不是所有人都装两个。

更合理的是：

```text
核心研发 / 平台团队：OpenCode + Ultimate。
普通开发 / 内容 / 运营：Codex + Light。
CI / 自动化：按实际入口接 Codex 或 OpenCode。
模型治理：统一走 4SAPI。
```

这样既保留深度能力，又不强迫所有人换工具。

一个团队可以这样分：

```text
架构师：Ultimate，用于方案评审、复杂重构、Team Mode。
资深研发：Ultimate + Codex，按任务切换。
普通研发：Codex + Light，日常修改和 Review。
内容团队：Codex + Light，写作、资料整理、自动化。
平台团队：统一维护 4SAPI Key、模型和预算。
```

工具可以不同。

模型治理必须统一。

## 8. 4SAPI 在选型里的位置

无论选 Ultimate 还是 Light，都不要让模型 Key 散掉。

建议统一：

```text
Base URL: https://4sapi.com/v1
Key: 按团队 / 项目 / 环境 / 工具拆分
Model: 从 4SAPI 模型广场复制
Logs: 4SAPI 后台统一查看
Budget: 按 Key 和工作流设置额度
Audit: 按用途追踪失败和成本
```

建议 Key 命名：

```text
4sapi-opencode-ultimate-dev
4sapi-opencode-ultimate-teammode
4sapi-codex-light-dev
4sapi-codex-light-automation
4sapi-codex-cloud-review
```

不要写成：

```text
test-key
my-key
ai-key
```

以后一定会忘。

## 9. 三种落地方案

### 9.1 个人开发者

推荐：

```text
先选一个主工具。
用 Codex 就装 Light。
用 OpenCode 就装 Ultimate。
4SAPI 先用一把个人 Key。
```

不要一开始就上 Team Mode。

先跑通：

```text
只读分析。
小修复。
测试验证。
写 AGENTS.md。
```

再考虑模型路由和自动化。

### 9.2 小研发团队

推荐：

```text
OpenCode Ultimate 给 1 到 3 个核心研发。
Codex Light 给日常开发者。
4SAPI 按项目和工具拆 Key。
Team Mode 默认关闭。
```

先做四件事：

```text
统一 AGENTS.md。
统一模型命名。
统一 4SAPI Key。
每周看成本日志。
```

等任务量稳定后，再开放 Team Mode。

### 9.3 企业平台团队

推荐：

```text
Ultimate 和 Light 都支持。
但由平台团队维护统一模板。
```

平台团队需要提供：

```text
OpenCode Provider 模板。
OmO Ultimate 配置模板。
Codex Light 安装模板。
AGENTS.md 标准模板。
4SAPI Key 命名规范。
模型分层策略。
预算和日志报表。
上线检查清单。
```

这时 OmO 不只是个人插件。

它变成企业 Agent 工具链的一部分。

## 10. 迁移路线

不要一口吃成胖子。

建议四步。

第一步，统一模型网关。

```text
先把 Codex / OpenCode / 自研脚本的模型调用统一到 4SAPI。
```

第二步，统一规则。

```text
所有项目写 AGENTS.md。
明确禁止修改的文件、测试命令、交付格式。
```

第三步，分工具增强。

```text
Codex 用户装 Light。
OpenCode 用户装 Ultimate。
```

第四步，开放高级能力。

```text
Team Mode。
fallback。
MCP。
Automations。
企业报表。
```

这个顺序比先装一堆插件稳。

## 11. 风险对比

Ultimate 的风险：

```text
能力强，配置也更复杂。
Team Mode 如果不控，会增加成本。
多个 Agent 需要更清晰的规则。
OpenCode 和 OmO 配置要有人维护。
```

Light 的风险：

```text
容易被误解成完整 OmO。
Codex 插件和本地配置冲突时要排查。
功能比 Ultimate 轻，不能承诺完整 Team Mode。
自动化和 Hooks 仍然需要治理。
```

共同风险：

```text
Key 散落。
模型名乱写。
预算缺失。
日志不看。
插件来源不审。
生产权限过大。
```

这些共同风险，最好用 4SAPI 和团队规则解决。

## 12. 选型问题清单

技术负责人可以问团队 10 个问题。

```text
1. 现在主力是 Codex 还是 OpenCode？
2. 是否需要 Team Mode？
3. 是否有大型重构和安全审计任务？
4. 是否已有 Codex Plugins / MCP / Hooks？
5. 是否能接受 OpenCode 作为新的主工作台？
6. 谁维护模型 Provider 配置？
7. 谁维护 4SAPI Key 和预算？
8. 哪些任务必须保留人工 Review？
9. 哪些团队只需要轻量增强？
10. 每周是否有人看模型成本和失败日志？
```

如果前 3 个答案都是“需要”，Ultimate 优先。

如果第 1、4 个答案都是 Codex，Light 优先。

如果团队混用，就两个都支持。

但 4SAPI 统一治理。

## 13. 不建议的选择

不建议一：

```text
因为 Ultimate 功能多，就让所有人装 Ultimate。
```

工具要匹配岗位。

不建议二：

```text
因为 Light 安装简单，就拿它宣传完整多 Agent 编排。
```

这会误导读者。

不建议三：

```text
两个都装，但 Key 和模型各管各的。
```

这会让成本审计失效。

不建议四：

```text
Team Mode 第一周就全员开放。
```

先试点。

不建议五：

```text
没有 AGENTS.md、没有日志、没有预算，就说企业级落地。
```

那只是工具堆叠。

## 14. 推荐组合

如果让我给一个保守推荐：

```text
Codex Light 做团队普及。
OpenCode Ultimate 做专家模式。
4SAPI 做统一模型网关。
```

也就是：

```text
大多数人：Codex + Light + 4SAPI。
核心工程任务：OpenCode + Ultimate + 4SAPI。
大型任务：Ultimate Team Mode + 单独 4SAPI Key。
所有模型调用：4SAPI 后台统一看日志和成本。
```

这套组合的好处是：

```text
上手成本低。
深度能力保留。
治理入口统一。
不会强迫所有人换工具。
```

企业最怕一刀切。

工具链要分层。

## 15. 最后总结

Ultimate 和 Light 不是简单谁更好。

它们对应不同工作流。

```text
Ultimate：适合 OpenCode 深度多 Agent 编排。
Light：适合 Codex CLI 插件化增强。
```

如果你要：

```text
Team Mode
多 Agent 分工
Slash Commands
复杂工程任务
深度 OpenCode 工作流
```

选 Ultimate。

如果你要：

```text
Codex CLI 增强
rules 注入
LSP
comment-checker
ultrawork
轻量插件化
```

选 Light。

如果团队里两类人都有：

```text
两个都支持。
```

但最后一定要统一到：

```text
4SAPI 做企业级 API、Key、日志审计和成本治理。
```

一句话：

```text
工具可以分层，模型治理必须统一。
```

这才是 OmO 进入企业生产环境的正确姿势。

## 资料来源与延伸阅读

- oh-my-openagent GitHub：https://github.com/code-yeongyu/oh-my-openagent
- oh-my-openagent 安装指南：https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/guide/installation.md
- oh-my-openagent 配置文档：https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/reference/configuration.md
- oh-my-openagent Team Mode 文档：https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/guide/team-mode.md
- OpenCode Provider 文档：https://opencode.ai/docs/providers/
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 接入实操手册：https://4sapi.com/blog/4sapi-api-integration-setup-guide
