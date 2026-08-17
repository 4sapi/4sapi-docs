---
title: "【大模型API中转站】第116期 Hermes Jira到PR | 4SAPI成本审计"
category: 人工智能
tags:
  - 大模型API中转站
  - Hermes Agent
  - Jira
  - GitHub
  - 企业级API
  - 4SAPI
description: "基于 HermesBible 的 Jira 到 PR 四 Agent 工作流，拆解企业如何用 Hermes 把工单变成可审查 PR，并通过 4SAPI 做模型路由、Key 分组、日志审计和单工单成本统计。"
---

# 【大模型API中转站】第116期 Hermes Jira到PR | 4SAPI成本审计

本文是【大模型API中转站】系列的第116篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

HermesBible 里有一个很适合企业团队看的工作流：

```text
How we used four AI agents to turn Jira tickets into reviewed PRs
```

它的核心思路很简单：

```text
Jira 工单进来。
多个专职 Agent 分别做需求整理、代码实现、Review、CI 处理。
最后生成可审查 PR。
人类保留 merge 权限。
```

这个方向非常适合企业。

因为企业研发里最稳定、最重复、最容易被 Agent 介入的流程之一，就是：

```text
工单 -> 分析 -> 改代码 -> 测试 -> PR -> Review -> 合并
```

但要注意。

这不是“让 AI 自动合代码”。

真正稳的设计是：

```text
Hermes 负责把工单推进到可审查 PR。
人类保留最终合并权。
4SAPI 负责模型调用、Key、日志、成本和权限治理。
```

## 1. 为什么 Jira 到 PR 适合 Agent

不是所有研发任务都适合 Agent。

适合 Agent 的任务通常有三个特征：

```text
目标明确。
反馈明确。
失败可恢复。
```

Jira 工单里经常有：

```text
需求描述
复现步骤
验收标准
影响模块
优先级
负责人
关联 PR 或 issue
```

代码修改后又可以通过：

```text
测试
lint
typecheck
build
CI
Review
```

验证。

这就天然适合 loop。

Agent 不是凭感觉说“完成了”。

它可以把任务推进到：

```text
有 diff。
有测试证据。
有 PR 描述。
有风险说明。
等待人类 review。
```

这个边界很健康。

## 2. 四 Agent 分工怎么设计

企业版可以拆成四个角色。

第一个，Intake Agent。

负责：

```text
读取 Jira 工单。
提取目标、范围、验收标准。
判断信息是否足够。
标记风险和不确定项。
生成执行 brief。
```

第二个，Implementation Agent。

负责：

```text
读代码。
定位相关模块。
做最小修改。
补测试。
运行本地验证。
```

第三个，Reviewer Agent。

负责：

```text
审查 diff。
找正确性、安全、兼容性、测试缺口。
不做纯风格意见。
必要时打回 Implementation Agent。
```

第四个，CI Agent。

负责：

```text
读取 CI 日志。
判断失败原因。
区分代码问题、测试问题、环境问题和偶发失败。
给出修复建议。
```

这个分工比一个 Agent 从头干到尾稳。

因为干活的不给自己打分。

## 3. 企业版架构

可以这样设计：

```text
Jira Webhook
  -> Hermes Intake Agent
  -> 任务状态文件 / 队列
  -> Hermes Implementation Agent
  -> Git branch / PR
  -> Hermes Reviewer Agent
  -> CI Agent
  -> PR brief
  -> 人工 Review / Merge
```

模型层全部经过 4SAPI：

```text
Hermes Agent
  -> 4SAPI 企业API网关
  -> 不同模型
```

企业不要让每个 Agent 直接拿同一个万能 Key。

建议拆成：

```text
jira-intake-key
code-implement-key
code-review-key
ci-analysis-key
pr-summary-key
```

每个 Key 都能单独设额度和日志标签。

这样你可以回答：

```text
这张 Jira 工单花了多少钱？
哪一步最贵？
哪个 Agent 重试最多？
哪个模型失败率最高？
```

这就是企业级成本治理。

## 4. 4SAPI 模型路由建议

不要所有阶段都用同一个模型。

| 阶段 | 推荐模型类型 | 原因 |
| --- | --- | --- |
| 工单整理 | 中低成本长上下文 | 主要是抽取和结构化 |
| 方案设计 | 强推理模型 | 要判断边界和风险 |
| 代码实现 | Coding 能力强的模型 | 需要稳定改文件 |
| Review | 强模型或审查模型 | 要发现隐藏 bug |
| CI 分析 | 中等模型 | 根据日志定位原因 |
| PR 摘要 | 低成本模型 | 结构化总结即可 |

通过 4SAPI 的统一入口，可以把这些阶段都记录在一个项目下面。

比如给每次工单传入同一个业务 ID：

```text
jira_ticket = PROJ-1234
workflow = hermes-jira-to-pr
stage = review
```

后面做成本复盘才有意义。

## 5. 4SAPI 接入配置：让每张工单都有账

Jira 到 PR 工作流最适合做 4SAPI 成本审计。

因为它天然有业务 ID：

```text
Jira ticket ID
Git branch
Pull Request ID
CI run ID
```

建议每次模型调用都带上这些标签。

```text
workflow = hermes-jira-to-pr
ticket = PROJ-1234
stage = intake
agent = intake-agent
```

模型入口统一配置为：

```text
Base URL: https://4sapi.com/v1
API Key: 使用对应阶段的 4SAPI 专用 Key
Model: 从 4SAPI 模型广场复制
```

企业可以按阶段拆 Key：

```text
4sapi-hermes-jira-intake
4sapi-hermes-jira-planner
4sapi-hermes-jira-code
4sapi-hermes-jira-review
4sapi-hermes-jira-ci
4sapi-hermes-jira-summary
```

如果某个阶段异常，比如 Review Agent 一直反复挑错，就能单独看它的调用记录。

如果某张工单成本异常高，也能定位：

```text
是工单描述不清，导致 intake 反复问？
是代码实现失败太多轮？
是 CI 日志太长？
是 Review Agent 调用了太多强模型？
```

这就是企业级 API 网关和个人 Key 的区别。

个人 Key 只能告诉你总共花了多少钱。

4SAPI 这种企业 API 网关能告诉你：

```text
哪条工单花了钱。
哪个 Agent 花了钱。
哪个模型花了钱。
哪一步失败导致重试。
```

## 6. 一个企业版 Jira 工单模板

如果你想让 Hermes 稳定处理 Jira ticket，工单模板要先规范。

不要只写：

```text
登录有问题，帮忙看下。
```

企业应该让 Jira 工单包含：

```markdown
# Goal
修复登录表单重复点击会创建多次请求的问题。

# Background
用户在弱网环境下连续点击登录按钮，会触发多个 submit 请求。

# Scope
- 登录表单组件
- submit handler
- 相关测试

# Out of Scope
- 不改认证协议
- 不改后端用户表
- 不改支付或权限模块

# Acceptance Criteria
- pending 状态下按钮不可重复点击。
- 相关测试覆盖重复点击。
- 类型检查通过。
- 不引入新依赖。

# Risk Level
medium

# Required Validation
- pnpm test auth
- pnpm typecheck
```

Intake Agent 读到这种工单，就能直接判断是否适合进入自动流程。

如果缺字段，就不要开工。

让 Hermes 返回：

```text
工单信息不足，请补充复现步骤和验收标准。
```

这比让 Agent 自己猜强很多。

## 7. PR 描述要服务审查，不是服务表演

Hermes 开 PR 时，PR 描述建议固定成：

```markdown
# Summary
- 修复登录表单重复提交。
- pending 状态下禁用 submit 按钮。
- 补充重复点击测试。

# Ticket
PROJ-1234

# Files Changed
- apps/web/LoginForm.tsx
- apps/web/LoginForm.test.tsx

# Validation
- pnpm test auth：通过
- pnpm typecheck：通过

# Risk
- 未覆盖真实弱网端到端测试。
- 后端幂等保护不在本次范围内。

# 4SAPI Cost
- intake：0.18 元
- planning：0.42 元
- implementation：2.10 元
- review：0.76 元
- summary：0.05 元
- total：3.51 元

# Human Review Required
- 是否需要后端再加幂等保护。
```

这段里最重要的是 `4SAPI Cost`。

企业做 Agent 研发，一定要把成本写进 PR 或内部报告。

否则月底没人知道自动化到底值不值。

## 8. 工单进入前要做过滤

不是所有 Jira ticket 都应该交给 Agent。

适合自动推进的：

```text
小 bug
测试补充
文档更新
低风险前端修改
配置说明修正
明确的接口字段调整
有复现步骤的回归问题
```

不适合自动推进的：

```text
支付逻辑大改
权限体系重构
数据库迁移
安全策略变化
产品方向不明确
缺少验收标准
跨团队架构决策
```

Intake Agent 的第一件事，不是开工。

而是判断：

```text
这个工单是否适合 Agent？
```

不适合就生成问题清单，退回人工。

## 9. 状态文件要写清楚

每个 Jira 到 PR 工作流，建议维护一个状态文件。

```markdown
# Hermes Jira Workflow State

## Ticket
PROJ-1234

## Goal
修复登录表单重复点击会创建两次请求的问题。

## Scope
- 登录表单组件
- auth submit handler
- 相关测试

## Out of scope
- 不改认证协议
- 不改后端用户表
- 不改全局 UI 组件库

## Validation
- pnpm test auth
- pnpm typecheck
- 手动验证重复点击

## Current status
- 已定位到 submit button 未在 pending 状态禁用
- 已补测试
- typecheck 通过
- auth 测试仍有一个 mock 失败

## Next
修复 mock 并重跑 auth 测试。
```

这个文件不是给人好看的。

是给下一轮 Agent 接着干的。

## 10. 人类必须保留 merge 权限

企业版最重要的一条：

```text
Agent 可以开 PR，但不要默认自动合并。
```

尤其是：

```text
权限
支付
用户数据
生产配置
数据库迁移
CI 发布脚本
安全策略
```

这些改动必须人工 review。

好的设计是：

```text
Agent 把 PR 准备到可审查状态。
人类看 diff、测试、CI、风险说明。
人类决定 merge、退回或拆分。
```

这不是保守。

这是企业级安全线。

## 11. 失败怎么处理

Agent 工作流一定会失败。

关键是失败后能不能复盘。

常见失败类型：

```text
工单信息不足
代码路径判断错
测试环境缺失
CI 偶发失败
权限不够
模型输出不稳定
上下文过长
改动范围过大
```

每次失败都应该写入：

```text
失败阶段
失败命令
关键日志
是否重试
重试次数
最终处置
```

4SAPI 负责记录模型调用。

GitHub/Jira 负责记录业务状态。

Hermes 状态文件负责记录执行过程。

三者合起来，才叫可复盘。

## 12. 成本怎么算

企业最关心的是：

```text
一张工单跑到 PR，大概要花多少钱？
```

HermesBible 的社区案例会给出一个经验数字。

但企业不要直接照抄。

因为成本取决于：

```text
模型选择
代码库大小
上下文长度
测试失败次数
Review 轮数
CI 日志长度
是否并行多方案
```

建议在 4SAPI 里按 Key 和工单 ID 统计：

```text
intake 成本
implementation 成本
review 成本
ci-analysis 成本
summary 成本
总成本
```

跑 20 张真实工单后，再决定哪些类型值得自动化。

不要第一天就给所有 Jira ticket 开闸。

## 13. 4SAPI 后台应该重点看哪些指标

接入 4SAPI 后，不要只看总账单。

Jira 到 PR 工作流建议每天看这些指标：

```text
每张 ticket 平均成本
每个阶段平均成本
每个模型调用次数
失败重试次数
被人工退回的 PR 数量
CI 失败后修复成功率
Review Agent 拦截问题数量
人工最终 merge 比例
```

如果你发现：

```text
intake 成本高
```

说明 Jira 工单质量差，需要改工单模板。

如果：

```text
implementation 成本高
```

说明任务过大，应该拆 ticket。

如果：

```text
review 成本高
```

说明实现 Agent 质量不稳，或者 Review Agent 标准太宽。

如果：

```text
PR 采纳率低
```

说明自动化没有真正带来价值。

这就是成本治理反过来优化流程。

## 14. 企业上线检查清单

```text
[ ] Jira 工单是否有明确验收标准？
[ ] Intake Agent 是否能拒绝不适合的任务？
[ ] 每个 Agent 是否使用独立 4SAPI Key？
[ ] 是否限制 Implementation Agent 的文件范围？
[ ] 是否禁止自动修改生产密钥和发布脚本？
[ ] 是否强制运行相关测试？
[ ] Reviewer Agent 是否独立于 Implementation Agent？
[ ] CI 失败是否会分类记录？
[ ] PR 是否必须人工合并？
[ ] 成本是否能按 Jira ticket 统计？
```

这张表过不了，就不要叫自动研发流程。

最多叫代码助手实验。

## 15. 最后总结

Hermes 的 Jira 到 PR 工作流，很适合作为企业 Agent 的第一个严肃场景。

因为它有明确输入、明确输出、明确验证和明确人工关卡。

正确姿势不是：

```text
让 AI 自动完成研发。
```

而是：

```text
让 Hermes 把工单推进到可审查 PR。
让 4SAPI 管住模型、Key、日志和成本。
让人类保留最终合并权。
```

一句话：

```text
Agent 负责提速，企业网关负责治理，人负责最后责任。
```

这才是 Jira 到 PR 工作流真正能进团队的方式。

## 资料来源与延伸阅读

- HermesBible：Jira to PR four agents：https://www.hermesbible.com/flows/jira-to-pr-four-agents
- Hermes Agent 官方文档：https://hermes-agent.nousresearch.com/docs
- Hermes MCP 官方文档：https://hermes-agent.nousresearch.com/docs/user-guide/features/mcp
- Hermes Tools 官方文档：https://hermes-agent.nousresearch.com/docs/user-guide/features/tools
- Hermes Security 官方文档：https://hermes-agent.nousresearch.com/docs/user-guide/security
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 文档：https://4sapi.apifox.cn/
