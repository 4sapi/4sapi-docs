---
title: "【大模型API中转站】第66期 Codex /goal 长任务 | 防跑偏省成本"
category: 人工智能
tags:
  - 大模型API中转站
  - Codex
  - goal
  - Agent
  - 长任务
  - 成本治理
  - 4SAPI
  - 4sapi.com
description: "从官方 /goal 能力边界出发，设计一套 Codex 长任务飞行控制台：目标说明、外置记忆、验收闸门、成本预算、风险停机和复盘日志，帮助开发者把无人值守任务跑得更稳，并说明如何用 4sapi.com 这类大模型 API 中转站做多模型路由与成本治理。"
---

# 【大模型API中转站】第66期 Codex /goal 长任务 | 防跑偏省成本

本文是【大模型API中转站】系列的第66篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

这篇讲 Codex 的 `/goal`。

但我不打算重复那种“让 Codex 跑一整夜，写了多少代码，烧了多少 token”的故事。

那种故事很刺激，但对真正要落地的人帮助有限。

因为问题不在于 Codex 能不能跑很久。

真正的问题是：

```text
你敢不敢让它跑很久？
```

一个 AI agent 能连续工作几个小时，听起来很爽。

但如果没有目标边界、验收标准、预算限制和回滚机制，它跑得越久，风险越大。

所以这篇换一个角度：

```text
不要把 /goal 当成“夜间代写代码”。
要把它当成一个长任务飞行控制系统。
```

你需要的不是一句神奇 prompt，而是一套控制台：

- 目标说明书
- 外置记忆
- 验收闸门
- 成本预算
- 风险停机线
- 复盘日志
- 回滚策略

如果你还在做大模型 API 中转站接入、模型路由、日志治理、成本统计、批量重构这类任务，`/goal` 会特别有价值。

因为这些活有一个共同点：

```text
方向明确，步骤很多，人工做很烦，机器做很合适。
```

## 1. `/goal` 到底适合什么？

OpenAI 官方对 `/goal` 的定位很清楚：它适合让 Codex 围绕一个持久目标持续工作，直到达到可验证的停止条件、暂停、完成，或者需要更多输入。

关键词不是“长时间”。

关键词是：

```text
持久目标 + 可验证停止条件。
```

这句话非常重要。

如果目标不能验证，`/goal` 就会变成许愿。

比如这些目标就不够好：

```text
/goal 帮我优化这个项目。
/goal 把代码质量提高一下。
/goal 做一个更好的管理后台。
/goal 把这个工具打磨到能用。
```

这些话人能大概理解，但 agent 很难判断什么叫“优化完了”“更好”“能用”。

更好的目标应该是：

```text
/goal 按 docs/goal.md 完成模型 provider 重构。
停止条件：
1. 所有 LLM 调用统一经过 packages/llm/provider.ts
2. 现有测试通过
3. 新增 provider fallback 测试通过
4. README 和 .env.example 已更新
5. status.md 记录最终变更和未解决问题
```

这才是 `/goal` 能跑起来的任务。

它知道要改什么。

它知道不能乱改什么。

它知道怎么验证。

它知道什么时候停。

## 2. 先别写长 prompt，先写“任务机场”

很多人第一次用 `/goal`，会把所有东西塞进一段超长 prompt。

这有两个问题。

第一，官方文档里提到，`/goal` 的 objective 有长度限制。如果说明太长，应该把细节放进文件，再让 Codex 读取文件。

第二，长任务会经历压缩、恢复、上下文滚动。只把规则放在对话里，后面很容易丢。

所以我建议你不要把 `/goal` 当成一段 prompt。

把它当成一个“任务机场”。

每次夜跑前，在项目里建一个目录：

```text
docs/goal-run/
  goal.md
  scope.md
  plan.md
  todo.md
  verify.md
  budget.md
  stop.md
  status.md
  decision-log.md
```

这组文件就是 Codex 的外置大脑。

不靠模型记忆。

不靠对话历史。

落在磁盘上。

它每次迷路，都能回去读。

## 3. `goal.md`：只写一个目标，不写愿望清单

`goal.md` 负责定义这次长跑的唯一目标。

模板：

```md
# Goal

## Objective

完成 [具体任务]。

## Done Means

- [ ] 条件一
- [ ] 条件二
- [ ] 条件三
- [ ] 文档已更新
- [ ] 测试已通过
- [ ] status.md 已写最终总结

## Not In Scope

- 不做 A
- 不做 B
- 不引入新依赖，除非先说明原因
- 不改动无关模块

## First Files To Read

- AGENTS.md
- docs/goal-run/scope.md
- docs/goal-run/todo.md
- docs/goal-run/verify.md

## Stop When

- 所有 Done Means 勾选完成
- 或 stop.md 中任一停机条件触发
- 或连续 3 次验证失败且原因相同
```

这里最关键的是 `Not In Scope`。

长任务最怕不是做不完。

而是做着做着扩大范围。

你让它重构 provider，它顺手改 UI。

你让它补测试，它顺手升级依赖。

你让它接入中转站，它顺手重写整个请求层。

这些都要提前写死。

## 4. `todo.md`：每条任务都必须带验收标准

普通 todo 是这样：

```md
- 重构 provider
- 支持 4SAPI
- 补测试
- 更新文档
```

这对 `/goal` 不够。

长任务 todo 必须带验收标准。

推荐格式：

```md
# Todo

## 1. 统一 LLM Provider 配置

### Task

把模型 base_url、api_key、model 从环境变量读取，统一放到 `packages/llm/config.ts`。

### Acceptance

- [ ] `.env.example` 包含 `LLM_BASE_URL`、`LLM_API_KEY`、`LLM_MODEL`
- [ ] 代码中没有硬编码 API key
- [ ] 单元测试覆盖缺少环境变量时的错误提示

### Verify

```bash
pnpm test packages/llm/config.test.ts
```

## 2. 新增中转站 Provider

### Task

新增 OpenAI-compatible provider，用于接入 4SAPI 这类大模型 API 中转站。

### Acceptance

- [ ] provider 支持自定义 base_url
- [ ] provider 不记录完整 prompt 和 API key
- [ ] 失败时返回可读错误
- [ ] 旧 provider 不被破坏

### Verify

```bash
pnpm test packages/llm/provider.test.ts
```
```

这类 todo 有两个好处。

第一，Codex 不会只凭感觉说“做完了”。

第二，你早上验收时，不用从一堆 diff 里猜它做了什么。

直接看勾选和验证命令。

## 5. `verify.md`：把验收变成闸门

`/goal` 最应该依赖的不是自信，而是验证。

OpenAI 官方提示里也强调，Codex 在能验证工作时输出质量更高。要提供复现步骤、验证命令、lint、pre-commit 检查。

所以你需要一个 `verify.md`。

模板：

```md
# Verification Gates

## Gate 1: Static Checks

```bash
pnpm lint
pnpm typecheck
```

## Gate 2: Unit Tests

```bash
pnpm test
```

## Gate 3: Build

```bash
pnpm build
```

## Gate 4: Runtime Smoke Test

```bash
pnpm dev
```

Open `http://localhost:3000/chat` and verify:

- 页面能加载
- 模型选择器显示正确
- 发送消息后有 loading
- 成功返回内容
- 错误不暴露 API key

## Gate 5: Security Review

Check:

- no API key in source files
- no full prompt logging
- no `.env` committed
- no unrelated dependency changes
```

如果任务是大模型 API 中转站接入，建议再加一个中转站专用闸门：

```md
## Gate 6: LLM Gateway Check

- [ ] `LLM_BASE_URL` 支持 OpenAI-compatible `/v1`
- [ ] `LLM_MODEL` 可配置
- [ ] 401 有清晰提示
- [ ] 429 有清晰提示
- [ ] 5xx 有重试或降级策略
- [ ] 日志只记录 request_id、model、latency、status
- [ ] 不记录 API key、完整 prompt、用户隐私
```

这就是这篇和普通 `/goal` 教程最大的区别：

```text
不是让 Codex 长跑，而是让 Codex 过闸。
```

每过一关，才进入下一关。

## 6. `budget.md`：长任务必须先算账

长任务还有一个非常现实的问题：

```text
token 会烧钱。
```

尤其是 `/goal` 这种持续运行任务。

如果你让它反复读全仓库、反复跑同一组测试、反复重写大文件，成本会非常难看。

所以我建议加一个 `budget.md`。

模板：

```md
# Budget

## Time Budget

- Soft limit: 2 hours
- Hard stop: 4 hours

## Token Budget

- Soft limit: 1,000,000 tokens
- Hard stop: 2,000,000 tokens

## File Change Budget

- Expected changed files: <= 12
- Hard stop changed files: > 25

## Scope Budget

Allowed directories:

- packages/llm/
- apps/web/src/chat/
- docs/

Disallowed directories:

- billing/
- auth/
- database/migrations/

## Cost Rules

- Do not repeatedly read large generated files
- Do not inspect `node_modules`
- Prefer targeted `rg` searches over whole-repo dumps
- Run full test suite only after focused tests pass
- If blocked by same error 3 times, stop and write status.md
```

普通用户不一定能直接看到精确 token。

但你依然可以控三个预算：

1. 时间预算
2. 文件改动预算
3. 验证重试预算

如果你是通过 4sapi.com 这类大模型 API 中转站做批处理、审稿、LLM-as-a-judge、日志分析，那就更应该把每类任务的 token 用量记录下来。

中转站适合做：

- 多模型调用统计
- 不同任务成本对比
- 模型路由
- 失败重试计数
- 按团队或项目分组计费
- 请求日志排查

`/goal` 负责让 Codex 长跑。

中转站负责让模型调用可治理。

两者不是一回事，但可以配合。

## 7. `stop.md`：提前写好停机线

很多人只告诉 Codex “不要停”。

但更重要的是告诉它：

```text
什么情况下必须停。
```

长任务不是永动机。

没有停机线，它会把“继续努力”理解成“继续乱试”。

`stop.md` 模板：

```md
# Stop Conditions

Stop and report if any condition occurs:

## Safety

- Need to delete or rewrite more than 3 existing user-authored files
- Need to modify `.env`, secrets, credentials, billing, auth, or production config
- Need network access not already approved
- Need to upload files or logs to external services

## Scope

- Required changes touch directories outside allowed scope
- Task requires product decision not defined in goal.md
- Implementation conflicts with AGENTS.md

## Validation

- Same test fails 3 times with same root cause
- Build failure seems unrelated to current task
- Cannot run required verification commands

## Cost

- Changed files exceed budget
- Runtime exceeds hard time limit
- Need to inspect very large files repeatedly

## Output

When stopping, write:

- what was completed
- what remains
- exact blocker
- suggested next human decision
- current git diff summary
```

这份文件会救命。

尤其是涉及：

- API key
- 计费逻辑
- 数据库迁移
- 权限系统
- 生产配置
- 客户数据

这些都不该让 agent 夜里自己决定。

## 8. `status.md` 和 `decision-log.md`：给长任务装黑匣子

长任务最怕第二天醒来不知道发生了什么。

所以必须让 Codex 写运行日志。

`status.md` 负责当前状态：

```md
# Status

## Current Phase

Phase 2: Provider implementation

## Completed

- [x] Read AGENTS.md
- [x] Added env config loader
- [x] Updated .env.example

## In Progress

- [ ] Add gateway provider tests

## Failed Checks

- `pnpm test packages/llm/provider.test.ts`
  - reason: mock response schema mismatch
  - next action: update fixture

## Next Step

Fix provider fixture and rerun focused test.
```

`decision-log.md` 负责记录关键选择：

```md
# Decision Log

## 2026-xx-xx 22:40

Decision: Keep existing OpenAI SDK wrapper instead of replacing with raw fetch.

Reason:

- Less code churn
- Existing streaming path depends on SDK
- Easier rollback

Risk:

- Need to verify custom base_url works with current SDK version

Verification:

- Add provider test for custom base_url
```

这两份文件就是黑匣子。

你不用盯着它跑。

但你需要醒来以后能复盘。

## 9. `/goal` 开跑前的元提示：先让 Codex 审你的目标

直接写 `/goal` 是很多人踩坑的地方。

更好的做法是先用 `/plan` 或普通对话，让 Codex 帮你审目标。

提示词：

```text
我准备把下面这个任务交给 /goal 长跑。
请你先不要改代码，只做目标审查。

请检查：
1. 目标是否只有一个
2. 停止条件是否可验证
3. todo 是否每条都有验收标准
4. verify.md 是否能跑
5. stop.md 是否覆盖高风险场景
6. budget.md 是否合理
7. 哪些信息还缺失

请反问我所有必须澄清的问题。
```

这一步比你想象中重要。

因为人类很容易把“我脑子里知道的”误以为“我已经写清楚了”。

让 Codex 先审一遍目标，能提前暴露很多默认假设。

等目标文件打磨完，再启动：

```text
/goal 按 docs/goal-run/goal.md 完成长任务。
每个阶段开始前先读取 docs/goal-run/ 下的相关文件。
每完成一个 todo，更新 status.md。
每做关键技术选择，更新 decision-log.md。
每通过一个验证闸门，记录命令和结果。
触发 stop.md 任一条件时立即停止并汇报。
持续推进，直到 Done Means 全部完成或触发停机条件。
```

注意这里不是把所有细节塞进 `/goal`。

而是让 `/goal` 指向文件。

官方 CLI 文档也提到，目标说明过长时，应该把细节放到文件里再指向它。

## 10. 用 `/goal` 做大模型 API 中转站接入

现在给一个完整案例。

场景：

```text
你有一个聊天应用，原来直连 OpenAI。
现在想接入 4SAPI 这类大模型 API 中转站，支持 Claude、GPT、Gemini 的统一路由。
```

不要直接说：

```text
/goal 帮我接入 4SAPI。
```

要先搭文件。

### `goal.md`

```md
# Goal

## Objective

将当前项目的大模型调用改造成 OpenAI-compatible gateway provider，使其可以通过 4SAPI 这类大模型 API 中转站调用 Claude、GPT、Gemini。

## Done Means

- [ ] 现有聊天功能可用
- [ ] 新增 `LLM_BASE_URL`、`LLM_API_KEY`、`LLM_MODEL`
- [ ] 旧 provider 可回滚
- [ ] provider 测试覆盖成功、401、429、5xx
- [ ] 日志不记录 API key 和完整 prompt
- [ ] README 和 `.env.example` 更新
- [ ] `pnpm lint`、`pnpm test`、`pnpm build` 通过

## Not In Scope

- 不改登录系统
- 不改计费系统
- 不改数据库 schema
- 不重写 UI
- 不引入新生产依赖，除非先写入 decision-log.md
```

### `todo.md`

```md
## 1. Locate LLM calls

Acceptance:

- [ ] 列出所有模型调用文件
- [ ] 标注当前 SDK / HTTP 客户端
- [ ] 标注是否支持自定义 base_url

Verify:

```bash
rg -n "openai|anthropic|chat/completions|responses|LLM|model" .
```

## 2. Add gateway config

Acceptance:

- [ ] 环境变量集中读取
- [ ] 缺少 key 时有可读错误
- [ ] `.env.example` 更新

Verify:

```bash
pnpm test packages/llm/config.test.ts
```

## 3. Implement provider

Acceptance:

- [ ] 支持自定义 base_url
- [ ] 支持模型名配置
- [ ] 不泄露 key
- [ ] 兼容现有调用接口

Verify:

```bash
pnpm test packages/llm/provider.test.ts
```
```

这类任务非常适合 `/goal`。

原因是：

- 范围清楚
- 文件可定位
- 测试可运行
- 文档可更新
- 回滚可设计
- 成本和风险可控

它不是探索性产品需求。

它是工程体力活。

## 11. 用中转站做 LLM-as-a-judge，不要让同一个模型自嗨

长任务还有一个新玩法：

```text
让 Codex 写代码，让另一个模型做验收评审。
```

比如 `/goal` 负责实现功能。

然后通过 4SAPI 调用另一个模型做检查：

- 是否满足 spec
- 是否有安全风险
- 是否有遗漏测试
- 是否有文档缺口
- 是否有成本浪费
- 是否可能泄露隐私

这就是 LLM-as-a-judge。

但不要让同一个模型既当运动员又当裁判。

更稳的方式是：

```text
Codex 执行实现
测试脚本做硬验证
另一个模型做软评审
人做最终决策
```

可以写一个 `review-prompt.md`：

```md
# Review Prompt

请审查当前 diff 是否满足 goal.md。

重点检查：

1. 是否超出 scope.md
2. 是否违反 stop.md
3. 是否遗漏 verify.md 中的验证项
4. 是否存在 API key 泄露风险
5. 是否记录了完整 prompt 或用户隐私
6. 是否有无关依赖变化
7. 是否有可回滚方案

输出：

- 阻塞问题
- 非阻塞建议
- 需要人工确认的问题
```

然后用中转站调用不同模型批量评审。

这就是大模型 API 中转站在 `/goal` 长任务里的创新用法：

```text
不只是帮 Codex 找模型。
而是帮长任务建立多模型质检层。
```

## 12. `/goal`、Automation、Skill 三者别混

很多人会把 `/goal`、Automations、Skills 混在一起。

可以这样分：

| 能力 | 解决什么 | 适合场景 |
| --- | --- | --- |
| `/goal` | 一次长任务持续推进 | 迁移、重构、原型、实验 |
| Skill | 固定方法复用 | PR 审查、日志排查、接入检查 |
| Automation | 定时重复执行 | 周报、CI 检查、错误日志巡检 |

OpenAI 官方最佳实践里有一句很实用的判断：如果一个流程还需要很多人工引导，先做成 Skill；稳定以后再做 Automation。

我建议再加一句：

```text
如果一个任务需要连续推进到完成，用 /goal。
如果一个方法需要反复复用，做 Skill。
如果一个动作需要定期发生，做 Automation。
```

举例：

- 接入 4SAPI provider：用 `/goal`
- “大模型 API 接入检查”清单：做 Skill
- 每周检查模型错误日志：做 Automation

不要拿 `/goal` 做所有事。

它不是定时任务。

它是一个长程执行模式。

## 13. Worktree：夜跑尽量不要在主工作区跑

如果你要让 Codex 跑几个小时，最好不要直接在主分支跑。

用 Git worktree 或独立分支。

原因很现实：

- 可以隔离实验
- 可以随时看 diff
- 可以一键丢弃失败尝试
- 不影响你白天继续开发
- 不容易和其他线程改同一批文件

OpenAI 文档也提醒，多个线程可以并行，但要避免两个线程改同一批文件。

长任务尤其要注意这一点。

建议开跑前：

```bash
git status
git switch -c codex/goal-llm-gateway
git add .
git commit -m "checkpoint before goal run"
```

如果用 Codex App，可以用 Worktree 模式。

如果用 CLI，就至少先建分支。

早上验收时，看：

```bash
git status
git diff --stat
git diff
git log --oneline -5
```

不要只看它的总结。

看真实 diff。

## 14. 长任务提示词模板

下面这份可以直接复制。

```text
/goal 按 docs/goal-run/goal.md 完成本次长任务。

运行规则：

1. 开始前读取：
   - AGENTS.md
   - docs/goal-run/goal.md
   - docs/goal-run/scope.md
   - docs/goal-run/todo.md
   - docs/goal-run/verify.md
   - docs/goal-run/budget.md
   - docs/goal-run/stop.md

2. 按 checkpoint 工作：
   - 每次只完成一个 todo
   - 每完成一个 todo，运行对应 Verify
   - 通过后更新 status.md
   - 关键技术选择写入 decision-log.md

3. 验证规则：
   - 优先运行 focused test
   - focused test 通过后再跑全量测试
   - 不能运行的验证项要写明原因

4. 成本规则：
   - 不读取 node_modules、dist、build、coverage
   - 不重复读取大文件
   - 使用 rg 精确搜索
   - 超出 budget.md 任何硬限制时停止

5. 安全规则：
   - 不修改 secrets、billing、auth、production config
   - 不删除文件
   - 不上传日志或源码到外部服务
   - 不记录 API key 和完整用户隐私

6. 停止规则：
   - Done Means 全部完成则停止
   - stop.md 任一条件触发则停止
   - 同一错误连续失败 3 次则停止

最终输出：

- 完成项
- 未完成项
- 验证结果
- 风险
- 回滚方式
- 建议下一步
```

这份 prompt 不花哨。

但它能跑。

## 15. 早上验收清单

不要醒来第一件事就相信 Codex 的总结。

按这张清单验收：

```text
1. status.md 是否完整
2. decision-log.md 是否记录关键选择
3. todo.md 是否每项都有验证结果
4. git diff 是否只改了允许范围
5. 是否新增或修改了敏感文件
6. 是否有 API key、token、secret
7. lint/test/build 是否真的跑过
8. 前端页面是否真的打开验证过
9. README 和 .env.example 是否同步
10. 是否有回滚方案
```

再跑一遍：

```bash
pnpm lint
pnpm test
pnpm build
```

如果是 Web 项目，还要浏览器看一眼。

如果是大模型 API 中转站接入，还要测几类错误：

- Key 错误
- 模型名错误
- base_url 错误
- 限流
- 上游 5xx
- 超时

长任务交付不是“它说完成了”。

而是“验证能复现”。

## 16. 哪些任务适合夜跑？

适合：

- provider 重构
- 大量测试补齐
- 旧 API 迁移
- 文档同步
- 代码格式化加小范围修复
- UI 原型打磨
- 代码库安全扫描后的修复
- 多模型接入适配
- 根据固定 spec 搭小工具

不适合：

- 产品方向没定
- UI 风格需要你边看边拍板
- 涉及商业策略
- 涉及生产数据库迁移
- 涉及支付、权限、账号安全
- 需要频繁外部授权
- 需求本身还很模糊

判断标准很简单：

```text
能写出验收标准，就可以考虑 /goal。
写不出验收标准，就先别长跑。
```

探索归人。

体力归机器。

决策归人。

验证归系统。

## 17. 常见坑

第一，目标太大。

“重做整个项目”不是好 goal。

第二，停止条件模糊。

没有 Done Means，就没有真正完成。

第三，只有 todo，没有 verify。

没有验证，todo 只是愿望。

第四，没有预算。

时间、文件、token、重试次数都要有上限。

第五，不写停机线。

长任务最怕无限试错。

第六，在主分支裸跑。

早上醒来看到一堆混乱 diff，很难救。

第七，没有黑匣子。

没有 status.md 和 decision-log.md，你无法复盘。

第八，让同一个模型自评全部结果。

硬验证靠脚本，软评审可以用另一个模型，人做最终决策。

第九，把 `/goal` 当 Automation。

一次持续推进和定时重复执行不是一回事。

第十，把中转站当万能加速器。

中转站解决统一接入、路由、计费、日志，不替你定义需求和验收标准。

## 18. 最后总结

`/goal` 真正改变的不是 Codex 能不能写代码。

而是你的工作方式。

以前你是这样用 AI：

```text
我说一步，AI 做一步。
```

现在更像：

```text
我定义目标、边界、验证和停机线。
AI 按控制台执行。
```

所以这篇的核心不是“让 Codex 跑一整夜”。

而是：

```text
让 Codex 跑一整夜以后，第二天你还能看懂、验收、回滚、复盘。
```

如果你只是玩一把，直接 `/goal` 也可以。

如果你要把它用于真实项目，尤其是大模型 API 中转站接入、模型路由、批量重构、成本治理这种工程任务，就一定要搭这套长任务控制台。

一句话总结：

```text
/goal 负责持续推进，外置记忆负责不跑丢，验证闸门负责不自嗨，中转站负责把模型调用变得可统计、可路由、可治理。
```

像 4sapi.com 这类大模型 API 中转站，适合放在长任务系统的模型调用层和评审层。

Codex 负责干活。

脚本负责硬验证。

中转站负责多模型和成本。

人负责目标、边界和最终判断。

这才是 `/goal` 真正值得用的地方。

## 资料来源与延伸阅读

- [Follow a goal - Codex use cases](https://developers.openai.com/codex/use-cases/follow-goals)
- [Codex app commands](https://developers.openai.com/codex/app/commands)
- [Prompting - Codex](https://developers.openai.com/codex/prompting)
- [Slash commands in Codex CLI](https://developers.openai.com/codex/cli/slash-commands)
- [Iterate on difficult problems - Codex use cases](https://developers.openai.com/codex/use-cases/iterate-on-difficult-problems)
- [Sandbox - Codex](https://developers.openai.com/codex/concepts/sandboxing)
- [Best practices - Codex](https://developers.openai.com/codex/learn/best-practices)
- [Codex app server goal API](https://developers.openai.com/codex/app-server)
