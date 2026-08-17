---
title: "【大模型API中转站】第22期 OpenSpec实战指南 | 老项目改造少翻车"
category: 人工智能
tags:
  - 大模型API中转站
  - OpenSpec
  - Spec-Driven Development
  - Change Proposal
  - Coding Agent
  - 4SAPI
description: "面向已有代码库，拆解 OpenSpec 的 proposal、tasks、design、archive 工作流，说明如何让 Coding Agent 在老项目里按规范改动，并用 4SAPI 做多模型治理。"
---

# 【大模型API中转站】第22期 OpenSpec实战指南 | 老项目改造少翻车

本文是【大模型API中转站】系列的第22篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

如果说 Spec-Kit 更适合新功能从零拆解，Kiro 更适合把 spec 放进 IDE，那么 OpenSpec 最适合的场景就是：已有项目、多人协作、变更不能乱来。

老项目用 AI 编程，最怕的不是模型不会写，而是模型太会写。它会热心重构，会顺手改风格，会把历史约束当成“可以优化的坏味道”。OpenSpec 的价值，就是先把变更写成 proposal 和 tasks，让人 review 后再交给 Agent。

这篇重点讲：OpenSpec 怎么在老项目里落地？怎么和 Codex、Claude Code、ZCode、Cline 配合？如果团队用 4SAPI 做模型统一入口，又该怎么记录每次变更的成本和效果？

## 1. OpenSpec 的核心不是文档，是变更协议

OpenSpec 常见结构可以理解成：

```text
project.md：项目上下文和约定
specs/：当前系统能力规范
changes/：待实施的变更
proposal.md：为什么改、改什么
tasks.md：怎么拆任务
design.md：复杂变更的技术设计
archive/：完成后的历史记录
```

这套结构的关键点是：当前系统能力和未来变更分开。

没有 OpenSpec 时，很多团队的需求是这样流动的：

```text
群里一句需求
-> Agent 改代码
-> PR 里争论需求
-> 测试发现边界没写
-> 临时补逻辑
```

有 OpenSpec 后，流程变成：

```text
先读当前 spec
-> 写 change proposal
-> 写 tasks
-> review 通过
-> Agent 执行
-> 验证
-> archive
```

看起来多了一步，实际少了很多返工。

## 2. 老项目为什么更需要 OpenSpec

老项目里有大量隐性规则：

- 某个字段不能改名，因为外部客户在用。
- 某个接口返回结构不能动，因为旧版 App 依赖。
- 某个失败状态码看起来奇怪，但网关做了特殊处理。
- 某个表不能直接删字段，因为有离线任务。
- 某个环境变量不能暴露给前端。

这些东西很少完整写在 README 里。Agent 如果只看代码，很可能理解不到历史背景。

OpenSpec 的 `project.md` 和 `specs/` 就是为了把这些隐性规则显性化。哪怕写得不完美，也比完全靠人脑记忆强。

## 3. 示例：给 4SAPI Key 增加额度冻结

假设已有系统里有 API Key 管理，现在要新增一个规则：

```text
当项目额度用完后，自动冻结该项目的 Key，管理员可以手动解冻。
```

不要直接让 Agent 改。先写 proposal。

```text
# Proposal: add-quota-freeze

## Why
当前项目额度用完后仍可能继续触发模型请求，导致预算超支。需要在项目额度耗尽时自动冻结 Key，并保留管理员手动解冻能力。

## What Changes
- 增加项目额度耗尽后的 Key 冻结状态。
- 请求进入模型前检查项目额度和 Key 状态。
- 管理员可在后台手动解冻。
- 记录冻结、解冻和拦截请求的审计日志。

## Non-goals
- 不修改 4SAPI 的计费规则。
- 不展示完整请求正文。
- 不改变现有 Key 创建流程。
- 不自动发送短信或邮件。
```

再写 tasks：

```text
## Tasks
- [ ] 阅读当前 Key 状态模型和请求鉴权流程。
- [ ] 增加 frozen 状态或等价字段。
- [ ] 在模型请求入口增加额度耗尽检查。
- [ ] 超额时阻断请求并返回明确错误码。
- [ ] 增加管理员解冻接口。
- [ ] 增加冻结/解冻/拦截审计日志。
- [ ] 补充单元测试和权限测试。
- [ ] 更新后台页面状态展示。
```

这时再交给 Agent，边界清楚很多。

## 4. Design 什么时候需要写

不是所有变更都需要 design.md。OpenSpec 的好处是可以轻重自如。

需要 design 的场景：

- 涉及数据结构变更。
- 涉及多服务调用。
- 涉及权限、计费、审计。
- 涉及兼容旧接口。
- 涉及迁移和回滚。

不需要 design 的场景：

- 文案调整。
- 小 UI 状态。
- 单个接口字段展示。
- 单测补充。

额度冻结这个例子就值得写 design，因为它涉及请求入口、状态机、权限和审计。

design 可以这样写：

```text
## State
Key 状态增加 frozen：
- active：可用
- disabled：人工禁用
- frozen：系统因额度耗尽冻结

## Request Flow
1. 校验 Key 是否存在。
2. 校验 Key 状态。
3. 校验项目额度。
4. 如果额度耗尽，将 active Key 标记为 frozen。
5. 写入审计日志。
6. 返回 quota_exceeded。

## Admin Unfreeze
管理员可手动解冻，但必须同时确认项目额度已调整。
```

这类 design 对 Agent 很有用，也对 reviewer 很有用。

## 5. OpenSpec + Agent 的正确姿势

把 OpenSpec 丢给 Agent 时，不要只说“按这个实现”。更好的提示是：

```text
请先阅读：
1. openspec/project.md
2. openspec/specs/ 中相关能力规范
3. openspec/changes/add-quota-freeze/proposal.md
4. openspec/changes/add-quota-freeze/tasks.md

要求：
- 先输出你计划修改的文件，不要立即动手。
- 每次只完成 tasks.md 中一个任务。
- 完成任务后更新对应勾选状态。
- 必须运行相关测试或说明无法运行。
- 不允许修改 proposal 的业务目标。
```

这里最重要的是“每次只完成一个任务”。老项目里，Agent 一口气改太多文件，很难 review。

## 6. 4SAPI 在 OpenSpec 流程里的位置

OpenSpec 不绑定模型，这反而是优点。你可以用任何 Agent 执行：Codex、Claude Code、ZCode、Cline、Kiro 都可以。

团队层面建议统一走 4SAPI：

```text
OpenSpec proposal/tasks
        ↓
Agent 执行工具
        ↓
4SAPI
        ↓
不同模型
        ↓
调用日志与成本统计
```

建议按 OpenSpec 的 change_id 记录模型调用：

```text
change_id
project
agent_tool
stage
model
api_key_name
input_tokens
output_tokens
latency_ms
retry_count
test_result
review_result
```

这样可以回答一个很关键的问题：哪类变更最适合 Agent，哪类变更最容易翻车。

比如跑一段时间后你可能发现：

- 文档和测试补充成功率最高。
- API 小改动可以交给主力代码模型。
- 涉及计费、权限、迁移的变更必须强模型 + 人工 review。
- 老项目里“先写 proposal”比“直接改”平均少 30%-50% 返工。

最后这个比例需要你自己用日志统计，不要凭感觉。

## 7. 用 4SAPI 做模型分层

OpenSpec 的不同阶段适合不同模型。

| 阶段 | 任务 | 模型策略 |
| --- | --- | --- |
| proposal 草稿 | 澄清需求、写 why/what/non-goals | 会总结、会追问的模型 |
| design | 技术方案、状态机、迁移 | 强推理模型 |
| tasks | 拆小任务 | 稳定主力模型 |
| implement | 改代码、补测试 | 代码模型 |
| review | 检查越界、遗漏、风险 | 与实现不同的强模型 |
| archive | 整理完成记录 | 低成本模型 |

在 4SAPI 里，可以对应建几组 Key：

```text
openspec-plan-key
openspec-code-key
openspec-review-key
openspec-archive-key
```

每组 Key 设置不同额度和模型权限。这样即使某个 Agent 任务跑偏，也不会无限消耗最高价模型。

## 8. OpenSpec 的团队规则

建议团队约定 6 条规则：

```text
1. 涉及权限、计费、数据结构的变更必须先写 proposal。
2. proposal 必须包含 Non-goals。
3. tasks 每项必须可验证。
4. Agent 每次只执行一个任务。
5. 完成后必须更新 tasks 状态。
6. 合并前必须 archive 或说明为什么暂不归档。
```

这 6 条比“请 AI 谨慎一点”有效得多。

如果你想更严格，可以再加：

```text
任何涉及 API Key、账单、额度、审计的变更，必须由 spec-review-key 调用另一个模型复核。
```

这时 4SAPI 的好处就出来了：你能强制 review 阶段走不同模型，避免同一个模型自问自答。

## 9. OpenSpec 不适合什么

OpenSpec 不适合用来做纯灵感探索。

如果你还不知道要做什么，先别急着写 proposal。可以先用普通对话或 Spec-Kit 做需求澄清。

它也不适合很小的临时改动。比如：

- 改错别字。
- 改按钮文案。
- 补一个注释。
- 修一个明确的一行 bug。

OpenSpec 的优势在“有变更治理需求”的场景。越是老项目、多人协作、风险高，它越有价值。

## 10. 一套可复制的 Agent 提示词

```text
你是 OpenSpec 驱动的代码助手。

请先读取：
- openspec/project.md
- openspec/specs/
- openspec/changes/[change-id]/proposal.md
- openspec/changes/[change-id]/tasks.md
- 如存在，读取 design.md

工作规则：
1. 先说明你理解的变更目标和非目标。
2. 列出计划修改的文件。
3. 每次只执行 tasks.md 中一个未完成任务。
4. 修改后运行最小相关测试。
5. 更新任务状态，但不要擅自改 proposal 目标。
6. 最终输出：完成项、测试结果、风险、未完成项。

限制：
- 不读取 .env、密钥、证书。
- 不自动提交。
- 不修改无关模块。
- 如果发现 proposal 与代码现状冲突，先暂停并说明。
```

这段提示词可以固定到 4SAPI 的 openspec-code-key 使用说明里，让团队所有 Agent 都按同一套规则执行。

## 11. 总结

OpenSpec 的价值，是让老项目里的 AI 改动先有提案、再有任务、最后可归档。它不会让模型突然变聪明，但会让模型少乱动。

如果你的项目已经有一定历史，尤其涉及权限、计费、API Key、审计、数据迁移，建议先用 OpenSpec 建一层变更协议，再让 Agent 执行。

4SAPI 在这里适合作为模型治理层：不同 change_id、不同阶段、不同模型、不同 Key 都能记录下来。OpenSpec 让变更可审查，4SAPI 让模型调用可治理。一个管“做什么”，一个管“谁来做、花多少钱、结果如何”。

参考资料：

- OpenSpec 官方网站：https://openspec.dev/
- OpenSpec GitHub：https://github.com/Fission-AI/OpenSpec
- Spec-Kit GitHub：https://github.com/github/spec-kit
- Kiro 官方文档：https://kiro.dev/docs/
