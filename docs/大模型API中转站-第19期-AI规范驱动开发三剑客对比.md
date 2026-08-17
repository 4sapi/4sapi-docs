---
title: "【大模型API中转站】第19期 AI规范驱动开发三剑客 | Spec-Kit/Kiro/OpenSpec对比"
category: 人工智能
tags:
  - 大模型API中转站
  - Spec-Driven Development
  - Spec-Kit
  - Kiro
  - OpenSpec
  - Coding Agent
  - 4SAPI
description: "从规范驱动开发的真实落地出发，横向对比 Spec-Kit、Kiro 与 OpenSpec 的定位、工作流、适用场景和多模型接入方式，并给出团队选型建议。"
---

# 【大模型API中转站】第19期 AI规范驱动开发三剑客 | Spec-Kit/Kiro/OpenSpec对比

本文是【大模型API中转站】系列的第19篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

过去一年，AI 编程工具的热词从“代码补全”变成了“Coding Agent”，现在又开始往“规范驱动开发”转。这个变化很自然：当模型只能补一段函数时，提示词够用；当模型能改十几个文件、跑命令、写测试、提交 PR 时，只靠一句“帮我做个功能”就太危险了。

规范驱动开发的核心，不是多写文档，而是先把“要做什么、为什么做、怎么验收、哪些边界不能碰”写清楚，再让 Agent 执行。今天要对比的三款工具，正好代表了三条路线：

- Spec-Kit：GitHub 开源的 SDD 工具链，更偏“从规范到计划再到任务”的项目脚手架。
- Kiro：带 IDE、CLI 和 Web 的 Agentic IDE，更偏“在开发环境里把 spec、hooks、tasks 串起来”。
- OpenSpec：轻量、工具无关的规范层，更偏“给已有项目建立变更提案、任务和归档流程”。

如果你已经用 4SAPI 这类大模型API中转站统一接入 GLM、Claude、Kimi、MiniMax 等模型，这三类工具还会多一个关键问题：规范层该归规范层，模型层该归模型层。不要把工具选型、模型选型和账单治理搅成一锅粥。

## 1. 规范驱动开发到底解决什么

很多团队第一次用 AI 写代码，都会经历三个阶段。

第一阶段，很爽。

你让模型改一个按钮、补一个接口、生成一段测试，它几分钟就能给你结果。

第二阶段，开始不安。

模型会改无关文件，会忘记前面说过的限制，会把“没跑测试”写成“已验证”，还会在失败后越改越大。

第三阶段，回到工程。

你开始要求它先写计划、列文件、补测试、说明风险、交付 diff、等待 review。这时候你会发现，真正缺的不是更花哨的提示词，而是一套规范化流程。

规范驱动开发就是为了解决这些问题：

| 问题 | 没有规范时 | 有规范时 |
| --- | --- | --- |
| 需求边界 | 模型按感觉扩写 | 先定义用户故事和非目标 |
| 技术方案 | 边写边猜 | 先写 plan，再拆 tasks |
| 验收标准 | 只看模型自述 | 有测试、场景和检查项 |
| 变更记录 | 聊天散落各处 | spec、tasks、archive 可追踪 |
| 团队协作 | 每个人提示词不同 | 规范文件可 review |

它不是反 AI，而是让 AI 的行动更像工程流程。

## 2. 三剑客速览

先给一张总表：

| 工具 | 核心定位 | 最适合 | 不太适合 |
| --- | --- | --- | --- |
| Spec-Kit | 开源 SDD 工具包，从 spec 到 plan/task/implement | 新项目、功能从零设计、想把 SDD 流程标准化的团队 | 想要完整 IDE 体验的人 |
| Kiro | Agentic IDE，把 spec、hooks、tasks 放进开发环境 | 想在一个工具里完成需求、编码、预览、自动化的人 | 已经有强自定义终端工作流且不想换 IDE 的团队 |
| OpenSpec | 轻量规范层，用 proposal/tasks/design 管理变更 | 老项目、多人协作、需要先治理变更再交给 Agent 的团队 | 想开箱即用生成完整项目的人 |

三者的差异，不是“谁更高级”，而是抽象层不同。

```text
Spec-Kit：先把新功能从想法变成可执行任务
Kiro：在 IDE 里把规范和 Agent 执行连起来
OpenSpec：给项目变更建立可 review、可归档的规范层
```

如果你的团队还没有任何规范化 AI 开发流程，Spec-Kit 最容易建立方法论；如果你想要一个完整工作台，Kiro 更顺手；如果你已经有项目、PR、review 和发布流程，OpenSpec 更像一层轻量但很稳的秩序。

## 3. Spec-Kit：适合从零设计功能

Spec-Kit 的思路很直接：先用命令把想法变成 spec，再生成 plan，再拆 tasks，最后实现。

典型流程可以理解成：

```text
/specify：写清楚要做什么
/plan：生成技术方案
/tasks：拆成可执行任务
/implement：按任务执行
```

它的价值不是某个命令神奇，而是强迫你把“想法”变成“可验证的任务队列”。这对新项目和新功能尤其有用。

比如你要做一个“团队 API Key 管理后台”，一句话需求可能是：

```text
帮我做一个 API Key 管理后台，支持创建、禁用、限额和日志。
```

直接丢给 Agent，很容易得到一堆看起来完整但边界模糊的代码。用 Spec-Kit 的方式，应该先拆成：

```text
用户是谁：管理员、项目负责人、普通开发者
核心动作：创建 Key、禁用 Key、设置额度、查看调用日志
非目标：不做支付、不做发票、不接生产密钥
验收：每个 Key 必须有项目归属、额度限制、过期时间、审计日志
```

这时再让模型实现，成功率会高很多。

Spec-Kit 很适合和 4SAPI 搭配。Spec-Kit 管规范，4SAPI 管模型入口。你可以用不同模型分别做需求澄清、方案生成、代码实现和 review，不必所有阶段都押在同一个模型上。

## 4. Kiro：适合把规范放进 IDE 工作流

Kiro 的定位更像完整开发环境。它不是只给你一个规范目录，而是把 specs、tasks、hooks、Agent Chat、CLI 和 IDE 体验放在一起。

它适合两类人：

- 个人开发者：希望少配置，直接在 IDE 里从需求走到实现。
- 小团队：希望规范、任务、代码和自动化都在一个统一界面里流动。

Kiro 的关键点是 hooks。很多 Agent 工具都能响应你的指令，但 hooks 能把一些动作变成“自动触发的工程习惯”。

比如：

```text
保存接口文件后，自动检查 OpenAPI 文档是否需要更新。
修改数据库 schema 后，提醒补迁移和回滚说明。
生成代码后，自动要求补测试和验证命令。
```

这类东西看起来小，但对团队很实用。因为真正的工程质量，往往不是靠一次强模型爆发，而是靠一堆稳定的小检查。

Kiro 的注意点也很明确：它是一个比较完整的工作台。如果你的团队已经深度绑定 VS Code、JetBrains、Claude Code、Codex 或自研 CLI，迁移到 Kiro 需要评估工作流成本。

## 5. OpenSpec：适合已有项目的变更治理

OpenSpec 的气质和前两个不太一样。它更像一个项目内的规范协议。

它强调：

```text
先看当前项目规范
再提出 change proposal
再写 tasks
必要时补 design
实现后归档
```

这对老项目特别有价值。因为老项目最大的风险不是“模型不会写代码”，而是它不知道哪些历史约束不能碰。

比如你要改登录系统，OpenSpec 会鼓励你先写：

```text
变更 ID：add-login-rate-limit
为什么改：降低撞库风险
改什么：登录失败次数限制、锁定时间、审计日志
不改什么：不重构认证框架、不改变现有 token 格式
验收：单元测试覆盖失败次数、锁定恢复、日志落库
```

这类 proposal 可以先让人 review，再交给 Agent 实现。比直接让模型扫全仓库改登录逻辑稳得多。

OpenSpec 也非常适合接 4SAPI。因为它本身不强绑某个模型，你可以让 Codex、Claude Code、Cline、ZCode 或其他 Agent 读取 OpenSpec 的 proposal 和 tasks，再通过 4SAPI 统一接入模型、记录用量和失败率。

## 6. 三者怎么组合，而不是三选一

很多人会问：既然都是规范驱动开发，我到底选哪个？

我的建议是先看项目阶段。

| 项目阶段 | 推荐组合 | 原因 |
| --- | --- | --- |
| 从零做新产品 | Spec-Kit + 4SAPI | 先建立 spec/plan/tasks，再用多模型执行 |
| 个人或小团队快速开发 | Kiro + 旁路 4SAPI | IDE 内流程顺；外部复核、对比和成本统计可走 4SAPI |
| 老项目持续迭代 | OpenSpec + 现有 Agent + 4SAPI | 先管变更，再管实现，最容易融入 PR 流程 |
| 企业级研发治理 | OpenSpec + Spec-Kit + 4SAPI | Spec-Kit 做新功能，OpenSpec 做变更归档 |

一个比较成熟的组合是：

```text
新需求探索：Spec-Kit
日常编码执行：Kiro / Claude Code / Codex / ZCode
变更审查归档：OpenSpec
模型网关与成本治理：4SAPI
```

这样每个工具都在自己擅长的位置，而不是互相替代。

这里要特别说明一下 Kiro：Kiro 自身有订阅、额度和内置模型体系，不能简单假设它和所有自定义 Base URL 直接打通。更稳的做法是把 Kiro 用作 spec、tasks、hooks 和 IDE 前台；当需要外部复核、批量实现或多模型对比时，再让 Codex、Claude Code、ZCode、Cline 等可配置工具通过 4SAPI 执行。

## 7. 4SAPI 放在哪一层最合理

规范驱动开发解决的是“让 Agent 做对事”。4SAPI 解决的是“让团队管得住模型”。

推荐架构是：

```text
Spec-Kit / Kiro / OpenSpec
        ↓
Coding Agent：Codex / Claude Code / ZCode / Cline
        ↓
4SAPI 大模型API中转站
        ↓
GLM / Claude / GPT / Kimi / MiniMax 等模型
        ↓
日志、账单、额度、路由、审计
```

在 4SAPI 里，建议按阶段建不同 Key：

| Key | 用途 | 推荐限制 |
| --- | --- | --- |
| spec-research-key | 需求澄清、竞品调研、规范初稿 | 中等额度，可联网模型 |
| spec-plan-key | 技术方案、风险拆解 | 强推理模型，限制频率 |
| spec-code-key | 任务执行、补测试 | 主力代码模型，按项目计费 |
| spec-review-key | 交叉 review、上线前检查 | 高质量模型，低频调用 |

然后在日志里记录：

```text
project
change_id
tool
stage
model
input_tokens
output_tokens
latency_ms
status
human_result
```

这样跑一段时间以后，你会知道：哪些规范写得太虚，哪些模型适合写 plan，哪些模型适合执行，哪些场景根本不该交给 Agent。

## 8. 选型建议

如果你只想选一个，我会这样建议：

- 新项目：先试 Spec-Kit。
- 想要完整 IDE 体验：试 Kiro。
- 老项目和多人协作：先上 OpenSpec。
- 已经有多模型和成本压力：加 4SAPI 做统一入口。

但真正稳定的团队流程，通常不是单工具崇拜，而是分层。

```text
规范层：Spec-Kit / OpenSpec
执行层：Kiro / Codex / Claude Code / ZCode / Cline
模型层：4SAPI
验收层：测试、review、CI、人工确认
```

这套分层能避免一个常见误区：把某个 Agent 当成全能研发系统。强模型可以帮你跑得快，但规范和验收决定你会不会跑偏。

## 9. 总结

AI 规范驱动开发的“三剑客”不是三个同质化工具。

Spec-Kit 适合从想法到任务的标准化拆解；Kiro 适合把规范、任务和 Agent 执行放进 IDE 工作流；OpenSpec 适合老项目里的变更提案、任务清单和归档治理。

如果你只是个人试水，选一个顺手的先跑起来就行。如果你是团队负责人，更建议把规范层、执行层和模型层拆开：Spec-Kit/OpenSpec 管规范，Kiro 管 IDE 内体验，ZCode/Codex/Claude Code 等可配置工具承接可路由执行，4SAPI 管 Key、模型、账单、日志和路由。

一句话：规范驱动开发让 AI 少跑偏，4SAPI 让团队少失控。两者叠起来，才更像真正能落地的 AI 工程流程。

参考资料：

- Spec-Kit GitHub：https://github.com/github/spec-kit
- Spec-Kit 文档：https://github.github.com/spec-kit/
- Kiro 官方文档：https://kiro.dev/docs/
- Kiro Pricing：https://kiro.dev/pricing/
- OpenSpec 官方网站：https://openspec.dev/
- OpenSpec GitHub：https://github.com/Fission-AI/OpenSpec
