---
title: "【大模型API中转站】第30期 Codex/Claude自动化Git | Issue到PR闭环"
category: 人工智能
tags:
  - 大模型API中转站
  - Codex
  - Claude Code
  - GitHub
  - Git
  - Pull Request
  - Code Review
  - GitHub Actions
  - 4SAPI
description: "拆解 Codex 和 Claude Code 如何自动化使用 Git：从 Issue、分支、Worktree、Commit、Push、PR 到代码审查，并给出团队落地清单。"
---

# 【大模型API中转站】第30期 Codex/Claude自动化Git | Issue到PR闭环

本文是【大模型API中转站】系列的第30篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一篇我们把 GitHub 从注册、双重认证、搜索、建仓库、Issue、本地 Git、部署、Fork、Star、Pull、Push 过了一遍。这一篇进入更接近 2026 年开发现场的问题：Codex 和 Claude Code 到底是怎么自动化使用 Git 的？

很多人以为 AI Coding Agent 的核心是“会写代码”。实际用久了会发现，更关键的是它能不能进入标准工程流程：

```text
读 Issue -> 建分支 -> 改代码 -> 跑测试 -> 提交 commit -> push -> 开 PR -> 接受 review -> 修复 CI -> 合并
```

这条链路跑顺了，AI 才不是玩具，而是团队里的自动化开发助手。如果你再用 4SAPI 这类大模型API中转站统一管理多模型成本和 Key，Agent 的生产效率会更容易被量化和控制。

## 1. 为什么 AI 写代码更需要 Git 规范

人类开发者偶尔会“先改了再说”。AI Agent 更不能这样。

原因很简单：Agent 执行速度快、改动范围可能大、上下文理解依赖提示和仓库资料。如果没有 Git 约束，它可能把一次小任务扩散成一堆难审查的文件改动。

Git 给 AI Agent 提供了四道刹车：

| Git/GitHub 对象 | 对 Agent 的约束 |
| --- | --- |
| Issue | 明确任务边界 |
| Branch/Worktree | 隔离改动，不污染主线 |
| Commit | 记录每一步变更 |
| Pull Request | 让人和 CI 审查后再合并 |

所以，别直接对 Agent 说：

```text
把项目优化一下。
```

更好的说法是：

```text
基于 Issue #123 新建分支，只修改登录表单校验逻辑，增加对应测试，运行 npm test，创建 PR，并在 PR 描述中写出验证结果。
```

这不是“提示词更好看”，而是把 AI 放回工程流水线。

## 2. 自动化 Git 可以分三层

Codex 和 Claude Code 都能参与 Git 工作流，但常见场景可以分成三层。

| 层级 | 运行位置 | 典型任务 |
| --- | --- | --- |
| 本地 Agent | 你的电脑、IDE、桌面 App | 改代码、跑测试、commit、push、开 PR |
| GitHub 触发 | Issue、PR 评论、GitHub Actions | 自动审查、修 CI、从 Issue 生成 PR |
| 团队规则 | `AGENTS.md`、`CLAUDE.md`、PR 模板、分支保护 | 约束 Agent 怎么改、怎么审、怎么提交 |

本地 Agent 更适合你和它结对编程。

GitHub 触发更适合后台任务，比如 PR 自动审查、修复 CI、根据 Issue 开一个草稿 PR。

团队规则负责让不同 Agent 不乱跑：哪些文件不能碰、测试命令是什么、代码风格是什么、哪些风险必须标红。

## 3. Codex 怎么自动化 Git

Codex 现在不只是一个终端里的代码助手。按官方公开文档，Codex 主要有几类工作面：

```text
Codex Web/Cloud
Codex App
Codex CLI
Codex IDE Extension
Codex GitHub Code Review
Codex GitHub Action
```

不同入口的操作体验不一样，但落到 Git 上，核心动作差不多：读仓库、修改文件、展示 diff、提交、push、创建 PR、审查 PR。

### 3.1 Codex Web：连接 GitHub 后后台开工

Codex Web 的典型方式是连接 GitHub 账号和仓库，让 Codex 在云端环境里处理任务。

官方文档里的关键点是：

```text
连接 GitHub 后，Codex 可以读取仓库并从自己的工作结果创建 Pull Request。
Codex Cloud 可以在后台处理任务，也可以并行处理多个任务。
```

适合这类任务：

```text
修复一个明确 Bug
补一组测试
升级一个小依赖
重构某个函数
根据 Issue 开 PR
修复 PR 里的 CI 失败
```

不适合一上来就扔给它：

```text
重写整个系统
顺便优化性能
顺便换 UI
顺便升级框架
```

云端 Agent 最吃“边界”。边界越清楚，PR 越容易审。

一个比较稳的 Codex Web 任务可以这样写：

```text
请基于仓库默认分支实现 Issue #128。

范围：
1. 只修改 packages/api 下的登录校验逻辑。
2. 不改数据库 schema。
3. 增加或更新对应单元测试。
4. 运行 pnpm test --filter api。
5. 创建 Pull Request，PR 描述包含变更摘要、测试结果和风险点。
```

### 3.2 Codex App：本地、Worktree、Cloud 三种模式

Codex App 是桌面工作流。官方文档里提到它支持 Local、Worktree、Cloud 三种模式：

| 模式 | 含义 | 适合场景 |
| --- | --- | --- |
| Local | 直接在当前项目目录工作 | 你正在旁边盯着改 |
| Worktree | 创建独立 Git worktree | 多任务并行、避免污染当前工作区 |
| Cloud | 在云端环境运行 | 后台长任务、远程仓库任务 |

对 Git 来说，Worktree 很关键。

你可以把 Worktree 理解为：

```text
同一个 Git 仓库的另一份工作目录，通常挂在另一个分支上。
```

好处是：你当前分支还在写文章或修 bug，Codex 可以在另一个 worktree 里跑新任务，两边文件状态互不打架。

Codex App 还提供内置 Git 工具：

```text
看 diff
给 diff 加评论让 Codex 修改
按文件或 chunk stage/revert
commit
push
创建 Pull Request
```

高级 Git 操作仍然可以用内置终端：

```bash
git status
git pull --rebase
npm test
pnpm run lint
```

这类设计的价值在于：你不用在“AI 聊天窗口、终端、Git GUI、浏览器 PR 页面”之间来回切太多次。

### 3.3 Codex GitHub Code Review：在 PR 里叫它审

Codex 的 GitHub 集成可以做 PR 代码审查。

典型触发方式：

```text
@codex review
```

Codex 会审查 PR diff，遵循仓库里的指导文件，并在 GitHub 上发布代码审查。官方文档强调它聚焦高优先级风险，避免把 review 变成低价值噪音。

你还可以开启自动审查，让新 PR 打开后自动触发。

如果想让 Codex 按团队规则审查，可以在仓库里放 `AGENTS.md`：

```markdown
## Review guidelines

- 不要记录用户手机号、邮箱、身份证等 PII。
- 涉及鉴权的接口必须检查 middleware。
- 数据库迁移必须包含回滚说明。
- 测试缺失时视为 P1 风险。
```

PR 里也可以给一次性焦点：

```text
@codex review for security regressions
```

如果 Codex 发现问题，你还可以继续评论：

```text
@codex fix the P1 issue
```

在权限允许时，它可以基于当前 PR 上下文启动云端任务，并把修复推回分支。

### 3.4 Codex GitHub Action：把审查放进 CI

如果你想在 GitHub Actions 里自动跑 Codex，可以使用官方的 Codex GitHub Action。

典型用途：

```text
PR 自动反馈
Release 前检查
重复迁移任务
CI 质量门禁
```

官方示例里会把 OpenAI API Key 放到 GitHub Secret，比如：

```text
OPENAI_API_KEY
```

然后 workflow 里 checkout 代码，再调用 `openai/codex-action@v1` 执行提示。

这里有一个重要边界：GitHub Actions 里的 Token、Secret、权限要最小化。不要为了省事给 `contents: write`、`pull-requests: write`、`actions: write` 一路全开。

能只读就只读；需要评论 PR 才给 PR 写权限；需要提交 patch 才给 contents 写权限。

## 4. Claude Code 怎么自动化 Git

Claude Code 的 Git 自动化也很完整，但风格和 Codex 不完全一样。

它常见入口包括：

```text
Claude Code CLI
VS Code / JetBrains 集成
Claude Code GitHub Actions
GitHub Code Review
Claude Code Web/Desktop
```

在本地和 IDE 场景里，你可以直接让 Claude 做 Git 操作：

```text
commit my changes with a descriptive message
create a pr for this feature
summarize the changes I've made to the auth module
```

它可以 stage 变更、写 commit message、生成 PR 描述。创建 PR 时通常会借助 `gh pr create`、MCP 工具或集成环境提供的能力。

### 4.1 Claude Code 的 PR 工作流

Claude 官方工作流文档里给了一个很典型的 PR 节奏：

```text
先让 Claude 总结改动
再让 Claude 创建 PR
最后让 Claude 补充 PR 描述和风险说明
```

你可以这样做：

```text
请先总结当前分支相对 main 的改动，不要提交。
```

确认无误后：

```text
请创建一个 PR，标题用动词开头，正文包含 Summary、Tests、Risk 三段。
```

如果你已经安装并登录 GitHub CLI，Claude 可以通过 `gh` 读取 PR 上下文、创建 PR、继续处理 review。

### 4.2 Claude Code Worktree：并行任务不互相污染

Claude Code 支持用 worktree 开隔离会话：

```bash
claude --worktree feature-auth
```

每个 worktree 有自己的目录和分支，适合并行任务：

```text
一个 Claude 修登录 bug
另一个 Claude 补文档
你自己还在 main 分支看代码
```

这比让多个 Agent 同时挤在一个工作区安全得多。多 Agent 并行时，最常见事故就是互相覆盖文件、误删改动、搞乱分支状态。Worktree 能显著降低这类风险。

### 4.3 Claude Code GitHub Actions：在 Issue/PR 里 @claude

Claude Code GitHub Actions 支持在 Issue 或 PR 里用 `@claude` 触发自动化。

官方文档描述的能力包括：

```text
分析代码
创建 Pull Request
实现功能
修复 Bug
遵循 CLAUDE.md 项目规则
```

典型评论：

```text
@claude please implement this issue and open a PR.
```

或者在 PR 里：

```text
@claude review this PR for auth and input validation issues.
```

如果你希望 Claude 按项目规范工作，就写好 `CLAUDE.md`：

```markdown
# Project instructions

- Use pnpm, not npm.
- Run `pnpm test` before creating a PR.
- Do not modify generated files under `src/generated`.
- API changes must update `docs/api.md`.
- Prefer small, focused PRs.
```

`CLAUDE.md` 和 Codex 的 `AGENTS.md` 很像，都是把团队约束沉淀到仓库里。不同工具读取的文件名和规则不同，团队同时用多个 Agent 时，可以两份都放，并保持核心规则一致。

### 4.4 Claude Code 的 commit/PR 署名

Claude Code 默认会给 commit 或 PR 加上归因信息，例如类似 `Co-Authored-By` 的 trailer，PR 描述也可能包含生成来源。这个行为可以在设置里调整。

团队要提前定规则：

```text
AI 生成代码是否必须署名？
署名放 commit trailer 还是 PR 描述？
是否允许隐藏署名？
审计系统如何识别 AI 改动？
```

这个问题不只是面子工程。企业里做审计、合规、责任追踪时，AI 参与痕迹最好可查。

## 5. Codex 和 Claude 的自动化差异

简单对比：

| 维度 | Codex | Claude Code |
| --- | --- | --- |
| GitHub PR 审查 | `@codex review`、自动审查、AGENTS.md | GitHub Code Review、`@claude`、CLAUDE.md |
| 本地 Git UI | Codex App 有 diff、stage、commit、push、PR | IDE/CLI 通过命令和工具操作 |
| Worktree | Codex App 支持 Worktree 模式 | `claude --worktree` 支持并行会话 |
| 云端任务 | Codex Web/Cloud 连接 GitHub 后后台运行 | Claude Code GitHub Actions / Web 工作流 |
| 规则文件 | `AGENTS.md` | `CLAUDE.md` |
| CI 集成 | `openai/codex-action@v1` | Claude Code GitHub Actions |

怎么选？

如果你已经在用 Codex App/CLI，想要本地 diff、worktree、PR review 一体化，Codex 很顺。

如果团队已经大量使用 Claude Code、`CLAUDE.md`、Claude GitHub Actions，想从 Issue/PR 评论触发实现和修复，Claude Code 很自然。

如果你是企业团队，不一定二选一。更常见的是：

```text
Codex 做 PR 高信号审查和本地结对。
Claude Code 做复杂代码理解、Issue 实现和文档维护。
4SAPI 管多模型 API 成本、Key、日志和模型路由。
GitHub 管权限、审查、CI、部署和审计。
```

关键不是工具名字，而是把它们放进同一条 GitHub 流程。

## 6. 团队落地前，先准备这 8 个文件/设置

### 6.1 `README.md`

写清楚：

```text
项目做什么
如何安装
如何启动
如何测试
如何构建
常见故障
```

Agent 会读 README。README 越完整，Agent 越少猜。

### 6.2 `AGENTS.md`

给 Codex 用，适合写：

```text
代码风格
测试命令
Review 指南
禁止修改的目录
安全注意事项
PR 描述格式
```

### 6.3 `CLAUDE.md`

给 Claude Code 用，适合写：

```text
项目约定
包管理器
构建命令
测试命令
提交规范
部署注意事项
```

### 6.4 Issue Template

路径通常是：

```text
.github/ISSUE_TEMPLATE/
```

让每个需求都带背景、复现、验收标准。

### 6.5 Pull Request Template

路径：

```text
.github/pull_request_template.md
```

推荐模板：

```markdown
## Summary

- 

## Tests

- [ ] Not run
- [ ] Unit tests
- [ ] Integration tests
- [ ] Manual verification

## Risk

- 
```

### 6.6 分支保护

至少保护 `main`：

```text
禁止直接 push
要求 PR
要求 CI 通过
要求至少 1 人 review
禁止未解决评论直接合并
```

### 6.7 GitHub Actions 权限

给自动化最小权限：

```yaml
permissions:
  contents: read
  pull-requests: write
```

需要写代码时再给 `contents: write`。

### 6.8 Secret 管理

OpenAI、Anthropic、4SAPI、数据库、云厂商 Key 都不要写进仓库。

放到：

```text
GitHub Actions Secrets
环境变量
云平台 Secret Manager
4SAPI 项目级 Key 管理
```

4SAPI 适合管模型 API Key、额度、日志和路由；GitHub Token 仍然要在 GitHub 的权限体系里单独管理。

## 7. 五个自动化工作流模板

### 7.1 从 Issue 生成 PR

适合明确需求。

Issue 写法：

```text
标题：登录页空邮箱时应阻止提交

验收标准：
- [ ] 空邮箱点击提交时显示错误提示
- [ ] 不发起登录 API 请求
- [ ] 增加单元测试
- [ ] npm test 通过
```

给 Agent 的提示：

```text
请实现 Issue #123。
要求新建分支，不直接修改 main。
只改登录表单相关代码。
增加测试并运行 npm test。
完成后创建 PR，描述变更、测试结果和风险。
```

### 7.2 PR 自动审查

Codex：

```text
@codex review for security regressions and missing tests
```

Claude：

```text
@claude review this PR, focus on auth, input validation, and backward compatibility.
```

适合每个 PR 多一层自动检查，但不要让 AI 审查替代人类审查。AI 很擅长扫明显问题，但业务语义、产品取舍、架构长期影响仍要人负责。

### 7.3 修复 CI 失败

PR 评论：

```text
@codex fix the failing CI checks. Keep the change minimal and explain which test failed.
```

或者：

```text
@claude inspect the failed GitHub Actions logs and push a fix to this PR branch.
```

注意：让 Agent 修 CI 前，先确认 CI 日志不包含敏感信息。不要把 Secret 打到日志里。

### 7.4 文档随代码更新

提示：

```text
检查当前 PR 的 API 行为变化，更新 docs/api.md 和 README 中相关示例。
不要修改代码逻辑。
```

这类任务很适合 Agent，因为它可以对比 diff，再补文档。

### 7.5 Release 前检查

提示：

```text
请检查 main 分支相对上一个 tag 的变更，生成 release notes 草稿。
按 Features、Fixes、Breaking Changes、Docs 分组。
不要创建 release，只提交草稿文件。
```

这种工作可以放到 GitHub Action，也可以在本地 Agent 里跑。

## 8. 本地和云端协作的推荐命令流

如果你在本地让 Codex 或 Claude 改代码，建议先自己把分支准备好：

```bash
git switch main
git pull --rebase origin main
git switch -c agent/issue-123-login-validation
```

让 Agent 做事前，先给它边界：

```text
你当前在 agent/issue-123-login-validation 分支。
请只处理 Issue #123。
改动前先阅读 README、相关测试和登录表单代码。
改完运行 npm test。
不要提交，先给我看 diff 摘要。
```

看完 diff 后再让它提交：

```text
请 stage 本次任务相关文件，写一个清晰的 commit message。
不要包含无关格式化。
```

然后：

```bash
git status
git log --oneline -5
git push -u origin agent/issue-123-login-validation
```

最后让 Agent 创建 PR：

```text
请创建 PR，目标分支 main。PR 正文包含 Summary、Tests、Risk。
```

这一套流程看起来啰嗦，但它能防止 Agent 一口气做太多，尤其适合生产项目。

## 9. 4SAPI 在自动化 Git 里的位置

很多人会问：如果 Codex 和 Claude 都能自动化 Git，那 4SAPI 这类大模型API中转站在里面扮演什么角色？

要分清两条线：

```text
GitHub 线：仓库、分支、Issue、PR、Actions、权限、Token
模型线：Claude、GPT、GLM、DeepSeek、API Key、额度、日志、路由
```

4SAPI 主要管模型线。

它适合做：

```text
统一 base_url
统一项目级 Key
多模型路由
额度控制
调用日志
成本统计
Claude/GPT/国产模型切换
```

它不替你做：

```text
GitHub 账号双重认证
GitHub 仓库权限
分支保护
Pull Request 审查
GitHub Actions 权限
Secret 泄露治理
```

一个更清晰的架构是：

```text
开发者 / Codex / Claude Code
        ↓
本地仓库 / GitHub 仓库
        ↓
Issue / Branch / PR / Actions
        ↓
模型调用层：4SAPI 或官方 API
        ↓
Claude / GPT / GLM / DeepSeek
```

如果你用本地可配置模型的 Coding Agent，可以通过 4SAPI 统一模型入口；如果用官方云端 Codex 或 Claude Code GitHub Actions，则按各自官方认证和 Secret 体系配置。不要把 GitHub Token 和模型 API Key 混在一起。

## 10. 安全红线：Agent 自动化不能越过这些规则

第一，不让 Agent 直接 push 到 `main`。

生产仓库必须走 PR。

第二，GitHub Token 最小权限。

只读任务不给写权限；只评论 PR 不给 contents 写权限；只处理一个仓库就不要给全组织权限。

第三，不把 Key 写进提示词。

不要这样：

```text
这是我的 API Key：<API_KEY>，请帮我写配置。
```

应该用环境变量或 Secret：

```text
请读取环境变量 OPENAI_API_KEY，不要在日志中打印。
```

第四，不让 Agent 大规模格式化无关文件。

否则 PR diff 会爆炸，review 质量会下降。

第五，不让 Agent 自己决定合并。

可以让它创建 PR、解释风险、修 CI，但合并应该由人或明确的自动化规则控制。

第六，谨慎处理 Fork PR。

外部贡献者的 PR 可能包含恶意代码。对 Fork PR 运行有 Secret 的 Action 时，要格外小心。

第七，保留审计。

AI 参与过的 commit、PR、review，建议保留可追踪痕迹。企业环境尤其重要。

## 11. 排错矩阵

| 现象 | 可能原因 | 处理方式 |
| --- | --- | --- |
| Agent 改了太多文件 | Issue 太模糊，缺边界 | 拆小任务，限制目录和目标 |
| PR 没有测试结果 | 提示没要求验证，README 没写命令 | 在 `AGENTS.md`/`CLAUDE.md` 写测试命令 |
| Agent 无法创建 PR | `gh` 未登录、权限不足、GitHub 未连接 | 检查 `gh auth status` 和仓库授权 |
| GitHub Action 不能评论 | workflow 权限不足 | 增加 `pull-requests: write` |
| 无法 push 分支 | Token/SSH 权限不够或分支保护 | 检查 remote、权限、分支规则 |
| Codex review 没反应 | 仓库没开启 Code review 或触发词不对 | 确认设置并使用 `@codex review` |
| Claude 不按规范 | `CLAUDE.md` 缺失或不清楚 | 写明确规则，减少口头约定 |
| 成本不稳定 | 任务太大、上下文太多、模型选择过贵 | 拆任务，用 4SAPI 项目 Key 和额度控制 |
| Secret 泄露 | 日志/提示/提交里出现 Key | 立刻吊销 Key，清理历史，复盘权限 |

## 12. 最小团队落地方案

如果你想从 0 到 1 引入 AI 自动化 Git，不要一开始就全自动合并。可以按四阶段来。

第一阶段：只让 Agent 读代码和解释 PR。

```text
不写文件，不提交，只做分析。
```

第二阶段：允许本地改代码，但由人提交。

```text
Agent 改，人 review，人 commit。
```

第三阶段：允许 Agent 创建分支和 PR。

```text
Agent 改、测、提交、开 PR，人 review 后 merge。
```

第四阶段：部分低风险任务自动化。

```text
文档更新、依赖小版本、格式修复、Release notes 可以半自动。
核心业务逻辑仍需人工审查。
```

一套最小配置：

```text
README.md
AGENTS.md
CLAUDE.md
.github/ISSUE_TEMPLATE/
.github/pull_request_template.md
分支保护
基础 CI
Secret 最小权限
4SAPI 项目级 Key 和额度
```

## 13. 总结

Codex 和 Claude Code 自动化 Git 的核心，不是“替你按几下按钮”，而是把 AI 放进标准工程闭环。

一句话总结：

```text
Issue 定边界，Branch/Worktree 隔离改动，Commit 记录历史，PR 承载审查，CI 做验证，AGENTS.md/CLAUDE.md 写规则。
```

推荐你这样用：

| 场景 | 建议 |
| --- | --- |
| 本地结对编程 | Codex App/CLI 或 Claude Code CLI/IDE |
| 并行任务 | Worktree |
| PR 审查 | `@codex review` 或 Claude GitHub Review |
| Issue 转 PR | Codex Cloud 或 Claude Code GitHub Actions |
| CI 自动反馈 | Codex GitHub Action / Claude Code GitHub Actions |
| 模型成本管理 | 4SAPI 项目级 Key、额度、日志和路由 |
| 生产合并 | 保留人工 review 和分支保护 |

AI Coding Agent 越强，GitHub 流程越不能松。真正可靠的自动化，不是让 AI 绕过规范，而是让 AI 在规范里更快地交付。

参考资料：

- Codex Web：https://developers.openai.com/codex/cloud
- Codex App Features：https://developers.openai.com/codex/app/features
- Codex GitHub Code Review：https://developers.openai.com/codex/integrations/github
- Codex GitHub Action：https://developers.openai.com/codex/github-action
- Claude Code GitHub Actions：https://docs.anthropic.com/en/docs/claude-code/github-actions
- Claude Code Common Workflows：https://docs.anthropic.com/en/docs/claude-code/common-workflows
- Claude Code VS Code Git 工作流：https://docs.anthropic.com/en/docs/claude-code/ide-integrations
- Claude Code Settings：https://docs.anthropic.com/en/docs/claude-code/settings
- GitHub Pull Request 文档：https://docs.github.com/articles/creating-a-pull-request
- GitHub Actions 权限文档：https://docs.github.com/actions/security-for-github-actions/security-guides/automatic-token-authentication
- 4SAPI 接入地址：https://4sapi.com/v1
