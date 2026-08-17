---
title: "【大模型API中转站】第208期 Claude Code企业治理 | Hooks权限审计"
category: 人工智能
tags:
  - 大模型API中转站
  - Claude Code
  - Hooks
  - 权限审计
  - 企业API网关
  - 4SAPI
  - 安全治理
  - 成本治理
description: "Claude Code 企业落地不能只靠提示词。真正的生产级治理要把 Hooks、Rules、权限、Subagents 和 4SAPI 日志审计连起来：禁止硬编码 Key、拦截危险命令、强制测试、记录模型调用、按项目拆 Key、按预算控制成本。"
---

# 【大模型API中转站】第208期 Claude Code企业治理 | Hooks权限审计

本文是【大模型API中转站】系列的第208篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

很多团队第一次把 Claude Code 接进项目，会先写规则。

比如：

```text
不要删除生产数据。
不要改支付逻辑。
不要提交 API Key。
修改代码后要跑测试。
模型调用必须走统一网关。
```

这些规则有用。

但只写在提示词里，不够。

因为提示词是软约束。

AI 大部分时候会遵守。

但在长会话、上下文压缩、复杂任务、文件里存在干扰内容时，它仍然可能漏掉。

企业落地 Claude Code，最重要的一句话是：

```text
安全边界不要只靠模型记忆。
```

要把关键规则做成确定性的门禁。

这一篇就讲：

```text
哪些规则写进 Claude Code 配置。
哪些规则必须用 Hooks 和权限拦截。
哪些模型调用必须通过 4SAPI 记录。
怎么把 Claude Code 变成可审计的企业级 AI 编程工具。
```

## 1. 提示词是提醒，不是防火墙

你可以在 `CLAUDE.md` 里写：

```text
不要删除 migration。
不要硬编码 API Key。
不要绕过 4SAPI。
不要直接修改 billing。
```

但这只是提醒。

真正的防线应该在工具层。

例如：

```text
PreToolUse 拦截删除 migration 的命令。
PostToolUse 自动跑 formatter 和 lint。
提交前扫描密钥。
模型调用前检查 Base URL。
CI 阻止缺少 request_id 的调用代码合并。
```

这是软件工程的基本逻辑。

你不会靠注释防止危险函数被误用。

你会用类型、权限、测试和审计。

AI Agent 也一样。

## 2. 企业最怕的 6 类事故

Claude Code 进入团队项目后，风险不只是“代码写错”。

更常见的是：

| 风险 | 后果 |
| --- | --- |
| 硬编码 Key | 密钥泄露、账单失控 |
| 绕过模型网关 | 无日志、无审计、无法追责 |
| 删除历史迁移 | 数据库不可恢复 |
| 修改支付/权限逻辑 | 生产事故 |
| 未跑测试就声明完成 | 半成品进入 PR |
| 长任务无限循环 | token 成本失控 |

这些风险里，很多都不能靠“提醒 Claude 小心点”解决。

必须用确定性机制。

## 3. 第一层：CLAUDE.md 写边界

`CLAUDE.md` 仍然需要。

它负责让 Claude 一进入项目就知道基本边界。

可以写：

```markdown
## AI Safety Rules

- Never hardcode provider API keys, Base URLs, model names, or tokens.
- All model calls must go through the model gateway.
- Do not edit `src/billing/**`, `src/auth/**`, or database migrations without explicit approval.
- After code edits, run the relevant test or explain why it cannot run.
- For 4SAPI-related changes, keep request_id, model, provider, task_type, cost_bucket, and error_code fields.
```

这层解决的是：

```text
让 Agent 先知道规矩。
```

但它不负责最终拦截。

拦截交给 Hooks、权限和 CI。

## 4. 第二层：Rules 管路径级高风险区

高风险目录不要只靠全局说明。

给它单独写规则。

比如：

```yaml
---
paths:
  - "src/billing/**"
  - "src/payments/**"
---
这是支付相关代码。
默认只读。
任何修改必须先输出风险说明和测试计划。
不允许自动提交。
不允许修改金额计算、退款逻辑、发票逻辑，除非用户明确要求。
```

模型网关目录也要单独管：

```yaml
---
paths:
  - "src/model-gateway/**"
  - "src/llm/**"
---
所有模型调用必须通过统一网关。
禁止在代码中写死 4SAPI Key。
禁止把上游错误原文直接返回给用户。
必须记录 request_id、model、task_type、cost_bucket、latency_ms。
```

这样 Claude Code 只有碰到相关文件时才加载规则。

既省 token，也更精准。

## 5. 第三层：Hooks 做硬门禁

Hooks 的价值是确定性。

适合“每次发生 X，就必须做 Y”的场景。

例如：

```text
每次编辑 TS 文件后跑 formatter。
每次修改模型网关后跑单元测试。
每次执行删除命令前检查目标路径。
每次提交前扫描 Key。
每次任务结束后写审计日志。
```

你可以把 Hook 理解成：

```text
Claude Code 外面的安全员。
```

它不靠模型理解。

它靠代码执行。

## 6. 哪些事必须用 Hook

我建议这些必须上 Hook 或 CI：

```text
阻止删除数据库迁移。
阻止提交 .env、Key、Token。
阻止绕过 4SAPI 的直连调用。
阻止生产配置里出现个人 Key。
修改 model-gateway 后自动跑测试。
修改 auth/billing 后要求人工确认。
任务完成前检查是否留下验证证据。
```

尤其是 4SAPI 相关项目，要拦三类东西：

```text
硬编码 Key。
硬编码 Base URL。
缺少日志字段的模型调用。
```

这三类问题一旦进生产，后面排账单和查问题会非常难。

## 7. 4SAPI 日志字段要提前设计

很多团队接入 API 网关，只关心能不能调用。

但企业场景真正重要的是：

```text
能不能查。
能不能控。
能不能追责。
```

建议在模型调用层统一记录：

```text
request_id
user_id
team_id
project_id
environment
task_type
agent_role
model
provider
cost_bucket
latency_ms
status_code
error_code
retry_count
fallback_from
fallback_to
```

4SAPI 能承担模型入口、Key 分组、日志审计和成本治理。

你的业务代码要做的是：

```text
把业务语义带进去。
```

否则日志里只看到一堆 chat completion。

看不出哪个项目、哪个 Agent、哪个任务在烧钱。

## 8. 用 Subagent 做独立审计

企业里不要让写代码的 Agent 自己审自己。

可以配置独立 Subagent：

```text
security-reviewer
cost-auditor
api-contract-checker
release-risk-reviewer
```

它们只读结果，不参与实现。

审计维度可以固定：

```text
有没有硬编码 Key。
有没有绕过 4SAPI。
有没有缺少 request_id。
有没有把上游错误暴露给用户。
有没有修改高风险目录。
有没有测试证据。
```

这比“你再检查一下”稳。

因为审计 Agent 有独立上下文和独立任务。

## 9. 一套可复制的企业审计 Skill

可以建一个：

```text
.claude/skills/enterprise-ai-gateway-review/SKILL.md
```

内容：

```markdown
---
name: enterprise-ai-gateway-review
description: Review model gateway changes for 4SAPI integration, key safety, audit logs, cost controls, and production readiness.
---

# Enterprise AI Gateway Review

## Checkpoints
- No hardcoded API keys or provider secrets.
- Base URL comes from environment or config.
- All calls go through the gateway.
- request_id, model, task_type, cost_bucket, latency_ms are recorded.
- 429/5xx have retry and fallback rules.
- User-facing errors are sanitized.
- Production changes require human approval.

## Output
- Findings
- Evidence path
- Severity
- Required fix
- Manual approval needed or not
```

这个 Skill 不负责改代码。

只负责审计。

审计和实现分开，是团队 AI 安全的底线。

## 10. 4SAPI 的企业级分组建议

不要所有人共用一个 Key。

建议按下面几个维度拆：

```text
dev / staging / prod
personal / team / service
codegen / review / docs / support
low-cost / strong-reasoning / vision
```

例如：

| Key 组 | 用途 |
| --- | --- |
| `dev-codegen` | 开发环境代码生成 |
| `dev-review` | 开发环境审查 |
| `prod-support` | 生产客服/知识库 |
| `ci-audit` | CI 自动审计 |
| `cost-report` | 成本日报 |

这样某个循环失控时，可以快速限流。

某个团队超预算时，可以单独调整。

某个模型错误率升高时，可以局部切换。

## 11. 生产上线前检查清单

接入 Claude Code + 4SAPI 前，建议过这张表：

```text
[ ] 根目录 CLAUDE.md 是否短而清楚
[ ] 高风险目录是否有 Rules
[ ] 模型网关目录是否禁止硬编码 Key
[ ] 是否用 Hook 扫描密钥
[ ] 是否用 Hook/CI 拦截危险删除
[ ] 是否要求模型调用记录 request_id
[ ] 是否记录 task_type、cost_bucket、agent_role
[ ] 是否按环境拆 4SAPI Key
[ ] 是否有 429/5xx fallback 策略
[ ] 是否有单任务预算上限
[ ] 是否有人工审批节点
[ ] 是否有独立 reviewer 或 Subagent 审计
```

这张表比任何“高级 Prompt”都重要。

## 12. 结尾

Claude Code 的强大，不只在会写代码。

更在它能被工程化配置。

但工程化配置也分软硬两层：

```text
软层：CLAUDE.md、Rules、Skills、Subagents。
硬层：Hooks、权限、CI、4SAPI 网关、日志审计。
```

个人项目可以先靠软层。

企业项目必须上硬层。

一句话：

```text
不要让 Claude Code 只靠“记得要安全”。
要让系统本身变得安全。
```

4SAPI 在这里的定位很明确：

```text
统一模型入口。
统一 Key 权限。
统一日志审计。
统一预算控制。
统一模型路由。
```

当 Claude Code 开始进入团队生产流，4SAPI 就不是可选项。

它是企业级大模型接入的治理层。

## 资料与延伸阅读

- Anthropic 官方博客：Steering Claude Code：https://claude.com/blog/steering-claude-code-skills-hooks-rules-subagents-and-more
- Claude Code Hooks 文档：https://code.claude.com/docs/en/hooks
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 文档：https://4sapi.apifox.cn/
