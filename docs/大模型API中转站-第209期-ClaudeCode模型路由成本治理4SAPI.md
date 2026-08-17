---
title: "【大模型API中转站】第209期 Claude Code成本治理 | 4SAPI模型路由"
category: 人工智能
tags:
  - 大模型API中转站
  - Claude Code
  - 4SAPI
  - 模型路由
  - 成本治理
  - Subagents
  - Skills
  - 企业级API
description: "Claude Code 配置越多，越要做成本治理。本文讲如何把 CLAUDE.md、Rules、Skills、Subagents、Hooks 和 4SAPI 结合起来：低成本模型做摘要和分类，强模型做架构判断和审计，长上下文模型做代码库搜索，并用 task_type、agent_role、cost_bucket 记录每一次调用。"
---

# 【大模型API中转站】第209期 Claude Code成本治理 | 4SAPI模型路由

本文是【大模型API中转站】系列的第209篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

Claude Code 的 7 类配置解决了一个问题：

```text
怎么让 Agent 更按你的方式工作。
```

但企业落地还有另一个问题：

```text
怎么让模型调用成本可控。
```

很多团队一开始很兴奋：

```text
CLAUDE.md 写好了。
Rules 配好了。
Skills 做好了。
Subagents 也能跑了。
Hooks 也能自动触发。
```

然后过了一段时间发现：

```text
账单看不懂。
哪个 Agent 花的钱不知道。
哪个任务用了强模型不知道。
为什么反复 fallback 不知道。
同一个错误重试了多少次不知道。
```

这就不是 Claude Code 配置问题了。

这是模型治理问题。

4SAPI 要解决的，就是这一层。

## 1. 不要所有任务都用最强模型

AI 编程任务可以拆很多角色。

不同角色对模型能力的要求完全不同。

| 任务 | 需要的能力 | 模型策略 |
| --- | --- | --- |
| 文件摘要 | 快速阅读 | 低成本模型 |
| 日志分类 | 模式识别 | 低/中成本模型 |
| 单文件修复 | 代码能力 | 中等模型 |
| 架构判断 | 推理和全局理解 | 强模型 |
| 安全审计 | 保守、严谨、对抗性 | 强模型 |
| 长代码库搜索 | 长上下文 | 长上下文模型 |
| PR 总结 | 提炼表达 | 低/中成本模型 |

如果你让强模型处理所有事情，成本一定高。

如果你让便宜模型处理所有事情，质量会不稳。

正确做法是：

```text
按角色路由模型。
```

## 2. Claude Code 配置层和模型层要分开

不要把模型名、价格、Key 写进 `CLAUDE.md`。

`CLAUDE.md` 只写原则：

```text
所有模型调用必须走统一网关。
所有任务必须记录 task_type。
高成本模型只用于架构判断、复杂修复和独立审计。
成本策略详见 docs/model-routing.md。
```

真正的模型路由写到配置里：

```text
docs/model-routing.md
config/model-routing.yaml
4SAPI 后台模型分组
```

这样做有两个好处：

```text
Claude Code 不会每次加载一堆价格说明。
模型策略可以独立更新，不污染项目提示词。
```

## 3. 一套推荐模型角色

可以先按 6 个角色拆。

```text
planner
worker
researcher
reviewer
summarizer
auditor
```

对应策略：

| 角色 | 用途 | 模型建议 |
| --- | --- | --- |
| planner | 拆任务、定边界 | 强模型 |
| worker | 写代码、改文件 | 中等/强模型，视任务难度 |
| researcher | 搜索仓库、读日志 | 长上下文模型 |
| reviewer | 查 bug、查回归 | 强模型 |
| summarizer | 总结 diff、写日报 | 低成本模型 |
| auditor | 安全、成本、权限审计 | 强模型或独立模型 |

在 4SAPI 里，这些可以映射为不同模型组。

比如：

```text
claude-code-planner
claude-code-worker
claude-code-reviewer
claude-code-summarizer
claude-code-auditor
```

你不一定要把真实模型名暴露给业务代码。

业务只知道角色。

4SAPI 管模型和路由。

## 4. 日志字段比模型选择更重要

很多团队一开始就纠结：

```text
到底用哪个模型最省？
```

但如果没有日志，你根本不知道哪里浪费。

建议每次模型调用记录：

```text
request_id
project_id
environment
task_type
agent_role
skill_name
subagent_name
model
route_group
cost_bucket
tokens_in
tokens_out
latency_ms
status_code
error_code
retry_count
fallback_from
fallback_to
```

最关键的是这三个：

```text
task_type
agent_role
cost_bucket
```

没有它们，日志只能告诉你“花了钱”。

有了它们，日志才能告诉你“钱花在哪里”。

## 5. 4SAPI 成本日报怎么做

可以做一个每日 Skill：

```text
.claude/skills/4sapi-daily-cost-report/SKILL.md
```

输出：

```text
今日总成本。
按项目分组。
按模型分组。
按 task_type 分组。
按 agent_role 分组。
重试成本 Top 10。
fallback 次数 Top 10。
缺少 cost_bucket 的调用。
建议降级模型的任务。
必须人工检查的异常。
```

这个 Skill 可以每天跑。

也可以在成本突然升高时触发。

它不负责改路由。

只负责给建议。

生产路由变更必须人工确认。

## 6. 三个最常见的浪费点

### 第一，CLAUDE.md 太长

每次会话都加载。

每行都是成本。

解决：

```text
删到 200 行以内。
把流程移到 Skills。
把路径规则移到 Rules。
把模型策略移到 docs/model-routing.md。
```

### 第二，Subagent 滥用强模型

Subagent 很好用。

但每个 Subagent 都可能独立消耗大量 token。

解决：

```text
只给 auditor、planner、reviewer 用强模型。
summarizer、classifier、log-grouper 用低成本模型。
给每个 Subagent 设置预算标签。
```

### 第三，失败无限重试

模型调用失败以后，如果没有停止条件，会越烧越多。

解决：

```text
429 做退避。
5xx 走备用模型。
同一输入最多重试 2 次。
连续失败写入 failed_jobs。
超过预算转人工。
```

## 7. 4SAPI 路由配置的思路

不要只配置一个默认模型。

建议配置几组：

```text
fast-cheap：摘要、分类、日志分组
balanced-code：常规代码修改
deep-reasoning：架构、复杂 bug、安全审计
long-context：代码库搜索、长文档阅读
vision：截图、图片、扫描文档
fallback：主模型失败时兜底
```

然后在 Claude Code 规则里写：

```text
如果任务是 summarize、classify、daily-report，优先使用 fast-cheap。
如果任务是 code-edit、test-fix，使用 balanced-code。
如果任务涉及 security、billing、auth、architecture，使用 deep-reasoning，并要求人工确认。
```

这就是企业级 API 网关的核心价值。

让模型选择不靠临场感觉。

而靠任务类型。

## 8. 给 Claude Code 的路由提示模板

可以放进 `docs/model-routing.md`：

```markdown
# Model Routing Policy

All LLM calls must go through 4SAPI.

## Route Groups

- fast-cheap: summarize, classify, extract, daily report.
- balanced-code: routine code edits, test fixes, refactors under 3 files.
- deep-reasoning: architecture, security, auth, billing, model gateway.
- long-context: repository search, log analysis, long docs.
- vision: screenshots, scanned PDF, UI inspection.

## Required Metadata

Every request must include:
- task_type
- agent_role
- project_id
- environment
- cost_bucket
- request_id

## Manual Approval

Manual approval is required before:
- changing production routing;
- modifying billing/auth/security;
- increasing budget limits;
- disabling fallback or audit logging.
```

`CLAUDE.md` 只需要写：

```markdown
Model routing rules are in docs/model-routing.md.
Read it before editing model gateway code or adding new model calls.
```

这样更省 token。

## 9. Hooks 可以帮你守成本

成本治理不只是事后看账单。

也要事前拦截。

Hooks 或 CI 可以检查：

```text
新增模型调用是否带 task_type。
是否带 cost_bucket。
是否走统一网关。
是否在测试里 mock 掉真实 API。
是否在生产环境使用个人 Key。
是否把强模型设为默认模型。
```

强模型当默认，是很多团队的隐形成本黑洞。

默认模型应该是平衡型。

强模型应该被显式选择。

## 10. 成本异常时怎么排查

如果 4SAPI 后台发现成本突然升高，可以按这个顺序查：

```text
1. 按 project_id 看是不是某个项目异常。
2. 按 task_type 看是不是某类任务暴涨。
3. 按 agent_role 看是不是某个 Subagent 空转。
4. 按 retry_count 看是不是失败重试。
5. 按 fallback_from/fallback_to 看是不是主模型不稳定。
6. 按 tokens_in 看是不是上下文过长。
7. 按 tokens_out 看是不是输出过度冗长。
```

排查时不要先换模型。

先找浪费发生在哪里。

很多时候不是模型贵。

而是：

```text
上下文太大。
重试太多。
没有分段。
没有停止条件。
不该用强模型的地方用了强模型。
```

## 11. 团队看板应该看什么

建议每周看这些指标：

```text
总成本。
按项目成本。
按模型成本。
按任务类型成本。
每个合并 PR 的模型成本。
每个通过审查任务的模型成本。
失败重试成本。
fallback 成本。
没有元数据的调用数量。
人工确认次数。
```

真正重要的是：

```text
每个被接受产出的成本。
```

不是调用越少越好。

而是每一块钱有没有换来可用结果。

## 12. 总结

Claude Code 的配置体系让 Agent 更会工作。

4SAPI 的模型治理让 Agent 工作得更可控。

两者结合，才适合团队长期使用。

可以用这套分工：

```text
CLAUDE.md：项目事实和路由入口。
Rules：路径级成本和安全规则。
Skills：成本日报、错误排查、发布检查。
Subagents：独立审计和深度分析。
Hooks：拦截缺元数据、硬编码 Key、危险默认模型。
4SAPI：统一模型入口、路由、日志、预算和 fallback。
```

一句话：

```text
不要让每个 Agent 自己决定用哪个模型。
把模型选择变成企业级路由策略。
```

这就是 4SAPI 在 Claude Code 团队落地里的核心卖点：

```text
多模型统一接入。
按任务路由。
按项目计费。
按日志审计。
按预算治理。
```

有了这层，Claude Code 才不只是个人提效工具。

它可以变成团队级 AI 工程生产线。

## 资料与延伸阅读

- Anthropic 官方博客：Steering Claude Code：https://claude.com/blog/steering-claude-code-skills-hooks-rules-subagents-and-more
- Claude Code 文档：https://code.claude.com/docs
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 文档：https://4sapi.apifox.cn/
