---
title: "【大模型API中转站】第45期 CLAUDE.md实战 | 8条规则让Claude少走弯路"
category: 人工智能
tags:
  - 大模型API中转站
  - Claude Code
  - CLAUDE.md
  - Hooks
  - AI编程
  - 4SAPI
description: "很多人把 CLAUDE.md 写成项目百科，结果 Claude Code 反而更容易跑偏。本文用 8 条实战规则，讲清楚如何控制长度、写禁止清单、用本地 CLAUDE.md 保护敏感模块、用 hooks 强制质量门禁，并结合 4SAPI 做模型和成本治理。"
---

# 【大模型API中转站】第45期 CLAUDE.md实战 | 8条规则让Claude少走弯路

本文是【大模型API中转站】系列的第45篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

很多人第一次认真用 Claude Code，都会干一件事：

```text
把所有“希望 Claude 记住的东西”，一股脑塞进 CLAUDE.md。
```

项目背景、技术栈、历史包袱、目录说明、代码规范、业务目标、老板偏好、产品愿景，甚至公司价值观。

看起来很勤奋。

实际效果经常是：Claude 更糊涂了。

因为 CLAUDE.md 不是项目百科，也不是团队 Wiki。它更像是 Claude Code 每次进入项目时先读的一张“现场作业卡”。

官方文档里说得很直接：`CLAUDE.md` 是给 Claude 的记忆文件，通常放在项目根目录；Claude Code 启动会话时会把它读进上下文。子目录下也可以放自己的 `CLAUDE.md`，在处理对应目录文件时按需加载。Anthropic 的最佳实践也提醒，越大的 `CLAUDE.md` 越会吃上下文，而且模型更可能漏掉其中的指令。

资料来源：

- Claude Code memory 文档：https://docs.anthropic.com/en/docs/claude-code/memory
- Claude Code best practices：https://www.anthropic.com/engineering/claude-code-best-practices
- Claude Code hooks 文档：https://docs.anthropic.com/en/docs/claude-code/hooks

这篇不讲“CLAUDE.md 的标准模板”。

模板网上很多，抄完不一定有用。

我更想讲 8 条实战经验：哪些内容该写，哪些内容不该写；哪些规则只适合提醒，哪些规则必须变成 hook；以及当你用 4SAPI 这类大模型 API 中转站统一管理模型时，CLAUDE.md 怎么配合成本和权限治理。

## 1. CLAUDE.md 越短越好，先控制在 200 行以内

反直觉的地方在这：

```text
你以为上下文越多，Claude 越懂项目。
实际上，上下文越臃肿，Claude 越容易错过真正重要的规则。
```

Claude Code 每次进入项目都会读 `CLAUDE.md`。也就是说，你写进去的每一行，都会占掉一部分上下文预算。

上下文预算不是免费的。

它本来可以用来读：

```text
当前文件
相关测试
类型定义
报错日志
最近改动
依赖调用链
```

如果 `CLAUDE.md` 先塞了 2000 行项目传记，Claude 真正看代码时就少了一大块空间。

我的建议很简单：

```text
根目录 CLAUDE.md 先压到 200 行以内。
```

这不是魔法数字，而是一个很好用的约束。

200 行以内，你会被迫做取舍。那些“也许有用”的背景，先删掉。那些“只有某个模块才需要知道”的信息，挪到本地 `CLAUDE.md`。那些“完整架构说明”，只保留链接。

错误写法：

```markdown
## Company Story
2021 年，我们在一次内部 hackathon 中发现运营团队每天都要手动导出报表。
2022 年，我们决定把这个能力产品化。
2023 年，我们开始服务第一批企业客户。
...
```

更好的写法：

```markdown
## Project Overview
这是一个 B2B 数据分析仪表盘，面向运营经理。

核心目标：
- 缩短“从数据到洞察”的时间。
- 优先保证加载速度和数据准确性。

技术栈：
- Next.js 15 App Router
- TypeScript
- Tailwind CSS + shadcn/ui
- Supabase/PostgreSQL

默认优先级：
数据正确 > 加载速度 > 交互丰富度 > 视觉花哨
```

判断一个根目录 `CLAUDE.md` 合不合格，可以用 30 秒测试：

```text
一个没看过你项目的人，读完后能不能回答：
1. 这是什么产品？
2. 技术栈是什么？
3. 新代码大概率应该放哪里？
4. 哪些目录不能随便动？
5. 改完后怎么验证？
```

能回答，就够了。

剩下的细节，不要堆在根文件里。

## 2. 写“不要引入什么”，比写“使用什么”更值钱

很多 `CLAUDE.md` 只写技术栈：

```markdown
## Tech Stack
- Next.js 15
- TypeScript
- Tailwind CSS
- Supabase
```

这还不够。

Claude 知道你在用什么，不代表它知道你为什么不用别的。

它可能看到一个复杂状态管理问题，就顺手建议 Redux。看到组件样式问题，就引入 styled-components。看到表格问题，就找一个它熟悉的 UI 库。它不是故意捣乱，只是根据训练里的“常见最佳实践”在补方案。

但你的项目可能已经有历史决策：

```text
Redux 已经迁出。
全站只用 Tailwind。
UI 系统一直是 shadcn/ui。
数据层已经锁 PostgreSQL。
```

所以要在 `CLAUDE.md` 里写禁止清单：

```markdown
## Tech Stack
- Next.js 15 App Router + TypeScript
- Tailwind CSS + shadcn/ui
- Supabase Auth + PostgreSQL
- Zustand for small client-side stores

## Do NOT Introduce Unless Explicitly Requested
- Redux: project already migrated to Zustand + React Context.
- styled-components: use Tailwind utilities and existing design tokens.
- Material UI: conflicts with shadcn/ui visual system.
- MongoDB: data layer is PostgreSQL.
- Moment.js: use date-fns already installed.
- New auth provider: auth flow is locked behind review.
```

这段通常比“请写优雅代码”更有用。

因为它能挡住后续一串隐性成本：

```text
新依赖
样式冲突
包体积变大
测试变多
团队认知分裂
后续会话继续沿用错误选择
```

如果团队通过 4SAPI 接入不同模型，还可以把这件事再往前收一层：

```markdown
## Model Policy
Default coding model is configured through 4SAPI.
Do not ask users to paste raw provider keys into project files.
Do not hardcode model names in application code unless the file already does so.
Read model routing config from env or the existing config layer.
```

这几行能避免 Claude 把某个供应商的 Key、模型名或 base URL 写死进代码里。

## 3. 规则必须能被检查，不要写成口号

`CLAUDE.md` 里最常见的废话是：

```markdown
- Write clean code.
- Keep it simple.
- Be performant.
- Follow best practices.
```

这些话都对。

但没用。

因为 Claude 没法稳定判断“clean”到底是什么。你也没法在 review 时快速判定它有没有遵守。

好的规则应该能被检查。

差的写法：

```markdown
## Coding Rules
- 写干净的代码
- 保持简洁
- 注意性能
- 不要过度设计
```

好的写法：

```markdown
## Coding Rules
- Use named exports by default, except framework-required files.
- Avoid `any`; use interfaces, generics, or `unknown` with narrowing.
- Keep React components under 200 lines unless there is a clear reason.
- Use async/await instead of `.then()` chains.
- Prefer existing helpers in `src/lib/` before creating new utilities.
- Do not leave commented-out code, debug logs, or temporary TODOs.
- Add tests for bug fixes that change behavior.
```

测试方法很简单：

```text
读完规则后，你能不能在 5 秒内判断一段代码是否违反？
```

能，就是好规则。

不能，就继续改。

再举几个更适合工程项目的规则：

```markdown
## API Rules
- Route handlers must validate input with the existing schema helpers.
- Never return raw database errors to clients.
- All external API calls must have timeout and retry policy.
- New endpoints must include success and failure tests.
```

```markdown
## UI Rules
- Use components from `src/components/ui/` before building custom controls.
- Buttons must use existing variants from `button.tsx`.
- Do not introduce new colors outside Tailwind theme tokens.
- Loading, empty, error, and success states are required for async views.
```

这才是 Claude Code 能执行的规则。

## 4. CLAUDE.md 要做路由器，不要做图书馆

很多人会把完整架构文档塞进 `CLAUDE.md`。

比如：

```text
系统架构 800 行
数据库设计 500 行
部署流程 300 行
接口说明 1000 行
```

最后根文件变成一本书。

Claude 每次进项目都要先背书，哪怕你只是让它改一个按钮文案。

更好的做法是把 `CLAUDE.md` 当路由器：

```markdown
## Project Documents
- Architecture overview: `docs/architecture.md`
- API contract: `docs/api.md`
- Data model notes: `docs/data-model.md`
- Deployment runbook: `docs/deploy.md`
- ADRs: `docs/adrs/`
- Archived decisions: `docs/archive/` (ignore unless explicitly requested)
```

Claude 不需要每次都读完整架构。

它只需要知道：

```text
需要架构时，去 docs/architecture.md。
需要 API 合同时，去 docs/api.md。
需要历史决策时，去 docs/adrs/。
旧资料在 archive，没要求别碰。
```

这就是渐进式上下文。

可以在 `CLAUDE.md` 里写得更明确：

```markdown
## Context Loading Policy
Always load:
- `CLAUDE.md`
- Files directly related to the task.

Load on demand:
- `docs/architecture.md` for cross-module changes.
- `docs/api.md` for endpoint or contract changes.
- `docs/adrs/` when changing a documented decision.

Do not load unless requested:
- `docs/archive/`
- old migration notes
- product brainstorm documents
```

这一招能明显省上下文，也能减少 Claude 被历史材料带偏。

对接 4SAPI 时也一样。

不要在 `CLAUDE.md` 里写一长串模型价格表和供应商说明。只放路由入口：

```markdown
## Model Gateway
Model routing and keys are managed outside this repository.
Read local examples from:
- `docs/model-routing.md`
- `.env.example`

Never commit real API keys.
Never paste provider secrets into documentation.
```

真正的成本规则、模型分组、Key 策略，可以放到单独文档。`CLAUDE.md` 只告诉 Claude 去哪里看。

## 5. 给高风险目录写本地 CLAUDE.md

根目录 `CLAUDE.md` 解决的是全局规则。

但项目里有些地方更危险：

```text
auth
billing
payments
infra
migrations
security
model gateway
```

这些目录不应该只靠根文件里一句“请小心修改”。

Claude Code 支持在子目录放 `CLAUDE.md`。当它处理对应目录里的文件时，会按需加载本地规则。

这很适合给危险区域装护栏。

比如认证目录：

```markdown
# src/auth/CLAUDE.md

## Security Rules
- Do not change token validation logic unless explicitly requested.
- Do not introduce a new auth provider without updating tests and docs.
- Do not log tokens, sessions, cookies, or auth headers.
- All auth changes must pass `pnpm test src/auth`.

## Known Traps
- Magic link IDs use `crypto.randomUUID()`. Do not replace with Math.random().
- Sessions are stored in Redis, not process memory.
- Middleware must preserve existing redirect behavior.

## Review Requirement
For auth changes, summarize:
1. What security behavior changed.
2. Which tests cover it.
3. What remains risky.
```

再比如账单目录：

```markdown
# src/billing/CLAUDE.md

## Billing Rules
- Never change invoice totals without adding tests.
- Currency values are stored in cents.
- Do not use floating point arithmetic for money.
- Webhook handlers must be idempotent.
- Any provider signature verification change requires review.
```

如果你的项目里有 4SAPI 或其他模型中转层，也建议给对应目录单独放一份：

```markdown
# src/model-gateway/CLAUDE.md

## Gateway Rules
- Never hardcode provider API keys.
- Never bypass the existing rate-limit middleware.
- Preserve request/response logging redaction.
- Do not expose raw provider error bodies to clients.
- New model routes must include timeout, retry, and cost tagging.

## Required Checks
- Run gateway unit tests after changes.
- Manually inspect redaction for logs touching `Authorization`, `api_key`, or `token`.
```

这种本地 `CLAUDE.md` 的好处是精准。

你不需要让 Claude 每次改按钮都读一堆账单规则，但它一旦进入账单目录，就必须知道这里不能乱动。

## 6. 重要规则别只写在 CLAUDE.md，要变成 hooks

`CLAUDE.md` 是提醒。

Hooks 才是门禁。

Claude Code 的 hooks 机制允许你在工具调用前后执行脚本。官方文档也明确提醒，hooks 会自动执行，所以要谨慎编写，并把安全性当成自己的责任。

适合写进 `CLAUDE.md` 的规则：

```text
倾向、偏好、协作方式、代码风格、项目说明。
```

适合变成 hook 的规则：

```text
格式化
测试
禁止改敏感文件
禁止提交密钥
检查 lint
校验生成文件
```

可以在 `CLAUDE.md` 里这样写：

```markdown
## Hooks & Quality Gates
The rules below are enforced by hooks, not just reminders:
- Format edited files after code changes.
- Run related tests after modifying core modules.
- Block edits to `src/auth/`, `src/billing/`, and `prisma/migrations/` unless confirmed.
- Scan for secrets before final summary.

If a hook fails, explain the failure and do not claim the task is complete.
```

示例 hook 思路：

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "pnpm prettier --write \"$CLAUDE_FILE_PATHS\""
          }
        ]
      }
    ]
  }
}
```

注意：上面是演示思路，实际变量名和可用上下文要按你当前 Claude Code 版本的 hooks 文档来写。

不要为了“看起来自动化”乱配 hook。

尤其不要让 hook 直接做这些事：

```text
删除文件
自动提交
自动推送
自动改数据库
自动旋转生产 Key
自动执行远程脚本
```

我更推荐先从三类低风险 hook 开始：

| Hook | 作用 |
| --- | --- |
| 格式化 | 避免每次都提醒缩进、换行、排序 |
| 相关测试 | 防止 Claude 改完不验证 |
| 密钥扫描 | 防止误把 Key 写入文件 |

这和 4SAPI 的治理逻辑是一致的。

能用规则自动挡住的，就不要靠人反复提醒。能由中转站限制的模型调用，也不要靠每个工程师自觉省钱。

## 7. 用 MEMORY.md 留下跨会话经验，但别让它失控

Claude Code 有多种记忆位置。项目级 `CLAUDE.md` 是最常见的入口。

但还有一种简单粗暴的方法：在项目里维护一个 `MEMORY.md`，专门记录“上次踩过的坑”。

你可以在 `CLAUDE.md` 里写：

```markdown
## Project Memory
Before starting non-trivial tasks, read `MEMORY.md` if it exists.

Use `MEMORY.md` for:
- recurring bugs
- important implementation discoveries
- local setup quirks
- known flaky tests
- decisions made during previous tasks

Do not store:
- secrets
- credentials
- personal data
- temporary notes older than 30 days

At the end of a task, suggest updates to `MEMORY.md` only when the discovery is likely to matter again.
```

这比很多复杂“长期记忆”方案更容易落地。

优点很实际：

```text
能进 Git
能 review
能删除
能追溯
不用额外数据库
不会黑箱召回
```

但它也有风险。

`MEMORY.md` 很容易变成新的垃圾桶。

所以要加一条清理规则：

```markdown
## Memory Hygiene
Keep `MEMORY.md` under 100 lines.
Prefer bullets with date and owner.
Remove stale notes when they no longer affect future work.
```

示例：

```markdown
# MEMORY.md

- 2026-06-17: `pnpm test` is slow on Windows; use `pnpm vitest src/auth` for auth-only checks.
- 2026-06-17: Billing amounts are cents; never use floats for invoice math.
- 2026-06-17: 4SAPI model route tests require `MODEL_GATEWAY_TEST_MODE=mock`.
```

它记录的不是“项目所有知识”，而是未来任务最可能再次用到的 5%。

## 8. 把你的协作风格写进去，省掉每次开场白

很多人每次开新会话，都会先说一堆：

```text
先别急着改。
先读代码。
不确定就问我。
回复用中文。
不要讲废话。
改完跑测试。
路径写清楚。
```

这些完全可以放进 `CLAUDE.md`。

比如：

```markdown
## Working Style
- Read relevant files before proposing changes.
- For risky changes, propose a short plan first.
- For small obvious fixes, implement directly.
- Ask when requirements are ambiguous and a wrong assumption is costly.
- Do not start replies with “Great question” or generic encouragement.
- Reply in Chinese. Keep code comments in English.
- Use absolute file paths when referencing files.
- Final response must include changed files and verification results.
```

这几行能省掉很多沟通成本。

但注意，不要把协作风格写成情绪宣泄。

差的写法：

```markdown
- 不要写垃圾代码。
- 不要自作聪明。
- 不要浪费我时间。
```

好的写法：

```markdown
- Do not introduce abstractions unless they remove duplicated logic or match an existing pattern.
- If you are unsure about a behavior, inspect existing tests before guessing.
- Do not refactor unrelated files while fixing a bug.
```

AI 不怕规则严。

AI 怕规则含糊。

## 一张表总结

| 经验 | 错误做法 | 推荐做法 |
| --- | --- | --- |
| 控制长度 | 写成项目百科 | 根文件压到 200 行以内 |
| 禁止清单 | 只写当前技术栈 | 写清楚不要引入哪些库 |
| 可检查规则 | 写“干净、简洁、高性能” | 写成可判断的工程规则 |
| 文档路由 | 把所有文档塞进来 | 只放文档入口和加载策略 |
| 本地规则 | 一个根文件管全项目 | 高风险目录放本地 CLAUDE.md |
| Hooks | 靠 Claude 记得跑测试 | 用 hooks 做格式化、测试、密钥扫描 |
| 长期记忆 | 每次会话重新解释 | 用 MEMORY.md 留下可 review 的经验 |
| 协作风格 | 每次开场重复说 | 写进 Working Style |

## 可以直接复制的一版最小 CLAUDE.md

如果你现在还没有 `CLAUDE.md`，可以先用这版。

```markdown
# CLAUDE.md

## Project Overview
This is a B2B analytics dashboard for operations managers.
Priority: correctness > loading speed > UX richness > visual polish.

## Tech Stack
- Next.js 15 App Router + TypeScript
- Tailwind CSS + shadcn/ui
- Supabase Auth + PostgreSQL
- Vitest for unit tests

## Do NOT Introduce Unless Explicitly Requested
- Redux
- styled-components
- Material UI
- MongoDB
- Real API keys or provider secrets in code/docs

## Coding Rules
- Use named exports by default, except framework-required files.
- Avoid `any`; use interfaces, generics, or `unknown` with narrowing.
- Prefer existing helpers before creating new utilities.
- Do not leave commented-out code or debug logs.
- Add tests for behavior changes.

## Context Loading Policy
- Read files directly related to the task before editing.
- Load `docs/architecture.md` for cross-module changes.
- Load `docs/api.md` for endpoint or contract changes.
- Ignore `docs/archive/` unless explicitly requested.

## Sensitive Areas
- `src/auth/`: read local CLAUDE.md before edits.
- `src/billing/`: read local CLAUDE.md before edits.
- `prisma/migrations/`: do not edit without confirmation.
- `src/model-gateway/`: never hardcode provider keys or bypass rate limits.

## Quality Gates
- Run related tests after behavior changes.
- If tests cannot run, explain why.
- If hooks fail, do not claim completion.

## Project Memory
- Read `MEMORY.md` before non-trivial tasks if it exists.
- Suggest updates only for discoveries likely to matter again.
- Never store secrets or personal data in memory files.

## Working Style
- For risky changes, propose a short plan first.
- For small obvious fixes, implement directly.
- Ask when ambiguity is costly.
- Reply in Chinese. Keep code comments in English.
- Final response must include changed files and verification results.
```

这版不完美，但方向是对的：短、具体、可检查、能路由、知道红线。

## 现在就能做的 5 件事

打开你的项目，按这个顺序改：

```text
1. 把根目录 CLAUDE.md 删到 200 行以内。
2. 加一个 Do NOT Introduce 区块，至少列 5 个禁用项。
3. 把“干净、简洁、最佳实践”改成可检查规则。
4. 给 auth、billing、infra、model-gateway 这类高风险目录加本地 CLAUDE.md。
5. 把格式化、测试、密钥扫描做成 hooks 或现有 CI 门禁。
```

如果你正在通过 4SAPI 统一接入模型，再额外做两件事：

```text
1. 在 CLAUDE.md 里写清楚：不要硬编码 provider key、base URL 和模型名。
2. 在模型中转层目录写本地 CLAUDE.md：保留限流、脱敏、成本标签和错误处理。
```

这能避免 Claude Code 在一次“顺手修复”里绕过你的模型治理层。

## 总结：CLAUDE.md 不是越聪明越好，而是越克制越好

写好 `CLAUDE.md` 的核心，不是把 Claude 训练成“知道你公司全部历史的人”。

而是让它每次进入项目时，先知道这些事：

```text
这个项目是什么。
技术栈是什么。
不要引入什么。
新代码放哪里。
危险目录有哪些。
改完怎么验证。
不确定时怎么和你协作。
```

这就够了。

剩下的，让它读代码、读测试、读按需文档。

Claude Code 最怕的不是信息少，而是信息噪音太多。一个短、硬、可执行的 `CLAUDE.md`，比一份 2000 行的项目自传更有用。

放到大模型 API 中转站的语境里，也是同一个道理。

当团队开始用 Claude Code、Codex、Cursor、多个模型和 4SAPI 一起跑工程任务时，你真正要管理的不是某一次生成，而是长期稳定性：

```text
上下文别浪费。
规则能执行。
敏感目录有护栏。
模型调用走统一入口。
成本能复盘。
风险能拦住。
```

`CLAUDE.md` 就是这套工程纪律的第一层。

别把它写成图书馆。

把它写成一张清楚的施工牌。
