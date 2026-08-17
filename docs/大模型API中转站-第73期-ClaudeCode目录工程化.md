---
title: "【大模型API中转站】第73期 Claude Code目录工程化 | 4SAPI团队模板"
category: 人工智能
tags:
  - 大模型API中转站
  - Claude Code
  - 4SAPI
  - .claude
  - 工程化
  - Hooks
  - Skills
  - Subagents
description: "基于 Claude Code 官方 .claude 目录、CLAUDE.md、settings、rules、hooks、commands、skills、subagents 文档，整理一套适合 4SAPI 大模型 API 中转站项目的团队级目录模板，让 Claude Code 更可控、更可维护、更适合真实工程协作。"
---

# 【大模型API中转站】第73期 Claude Code目录工程化 | 4SAPI团队模板

本文是【大模型API中转站】系列的第73篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

很多人已经开始用 Claude Code 写代码。

但大多数人的用法还停留在：

```text
开一个项目。
写一个 CLAUDE.md。
让 Claude Code 自己摸索。
```

小项目这样可以。

一旦项目变大，问题就会出现：

- 指令越写越长
- 规则散落在聊天记录里
- 测试命令每次都要重复说
- API Key 安全规则没人记得
- hooks 和 commands 混在一起
- skills、agents 建了一堆但没人知道什么时候用
- 团队成员各自有一套偏好，Claude 输出越来越不稳定

这时候你会发现：

```text
Claude Code 能不能干活，不只取决于模型能力，也取决于项目怎么组织它的上下文和权限。
```

尤其是做 4SAPI 这类大模型 API 中转站接入项目时，目录结构更重要。

因为这类项目天然涉及：

- 多模型 provider
- base_url 和 API Key
- 成本统计
- 日志脱敏
- fallback
- 计费和权限
- 前后端联调
- 自动化测试
- 安全审计

如果 `.claude/` 目录乱，Claude Code 很容易把该放在 `CLAUDE.md` 的规则写进 hook，把该做成 skill 的流程塞进长提示词，把个人偏好提交进团队仓库。

这篇就讲一套更工程化的 `.claude/` 目录组织方法。

目标不是炫技。

目标是让 Claude Code 在真实项目里：

```text
读得懂规则。
改得动代码。
跑得起测试。
守得住边界。
交付得可验收。
```

## 1. 为什么 `.claude/` 结构会影响 Claude Code 质量？

Claude Code 官方文档里有一个关键点：

```text
Claude Code 会从项目目录和用户级 ~/.claude 中读取 instructions、settings、skills、subagents、memory 等内容。
```

这意味着它不是只看你当前这一句 prompt。

它还会看项目里的长期说明、配置、规则和扩展能力。

如果这些东西组织得好，Claude Code 会更像一个熟悉项目的工程师。

如果这些东西组织得乱，它就会像一个拿到十几份互相冲突说明书的新同事。

典型混乱包括：

- `CLAUDE.md` 写了 2000 行，什么都塞进去
- `settings.json` 里既有团队规则，又有个人偏好
- hooks 里做了需要模型判断的事
- commands 里放了应该自动执行的检查
- skills 和 commands 重复
- agents 没有明确角色
- 本地覆盖文件被提交进 Git
- API Key、`.env`、日志文件没有 deny 规则

所以 `.claude/` 的核心不是“多建几个目录”。

而是把不同类型的上下文分层。

## 2. 推荐目录结构：给 4SAPI 项目用的版本

如果你的项目要接入 4SAPI，推荐从这个结构开始：

```text
your-project/
├── CLAUDE.md
├── CLAUDE.local.md              # 个人覆盖，不提交
├── AGENTS.md                    # 可选，给 Codex 等工具复用
├── .claude/
│   ├── settings.json
│   ├── settings.local.json       # 个人覆盖，不提交
│   ├── rules/
│   │   ├── llm-gateway.md
│   │   ├── security.md
│   │   ├── testing.md
│   │   ├── frontend.md
│   │   └── backend-api.md
│   ├── hooks/
│   │   ├── block-secret-reads.sh
│   │   ├── block-dangerous-commands.sh
│   │   └── remind-tests-before-stop.sh
│   ├── commands/
│   │   ├── review-llm-provider.md
│   │   ├── test-4sapi-smoke.md
│   │   └── summarize-ai-cost.md
│   ├── skills/
│   │   ├── 4sapi-provider-migration/
│   │   │   ├── SKILL.md
│   │   │   └── checklist.md
│   │   └── ai-cost-audit/
│   │       ├── SKILL.md
│   │       └── cost-template.md
│   └── agents/
│       ├── security-auditor.md
│       ├── llm-provider-reviewer.md
│       └── docs-writer.md
```

这不是让你一上来全建满。

这是最终蓝图。

刚开始只需要：

```text
CLAUDE.md
.claude/settings.json
```

等项目变复杂，再逐步加：

```text
rules -> hooks -> commands -> skills -> agents
```

## 3. 第一层：`CLAUDE.md` 只写每次都要知道的事

`CLAUDE.md` 是项目说明书。

官方文档说明，Claude Code 会读取工作目录及其父目录中的 `CLAUDE.md`，`CLAUDE.local.md` 也会一起加载。更靠近当前目录的说明通常更具体。

所以 `CLAUDE.md` 要轻。

不要把所有细节都塞进去。

适合放：

- 项目是什么
- 技术栈是什么
- 启动、测试、构建命令
- 目录结构
- 核心业务边界
- 最重要的安全规则
- 如何处理 4SAPI / LLM provider

示例：

```md
# CLAUDE.md

## Project

这是一个接入 4SAPI 的大模型 API 中转站示例项目，用于统一调用 Claude、GPT、Gemini、GLM 等模型。

## Commands

- Install: `pnpm install`
- Dev: `pnpm dev`
- Lint: `pnpm lint`
- Test: `pnpm test`
- Build: `pnpm build`

## Architecture

- `packages/llm/`：LLM provider、模型路由、错误处理
- `apps/web/`：前端页面
- `apps/api/`：后端 API
- `docs/`：接入说明和运维文档

## LLM Rules

- 不要把 4SAPI Key 写进代码
- 不要在前端暴露任何模型 Key
- 所有模型调用必须经过后端 provider 层
- 新增模型时同步更新 `.env.example`
- 修改 LLM provider 后必须运行相关测试
- 日志不能记录完整用户隐私、API Key、完整 prompt

## Before Editing

- 先找相关文件
- 先给最小改动计划
- 不要重构无关模块
- 涉及 3 个以上文件时先列清单
```

这份文件应该让 Claude Code 一分钟内知道项目怎么工作。

不是让它读一篇公司史。

## 4. `CLAUDE.local.md`：个人偏好不要污染团队规则

很多人会把个人习惯写进项目 `CLAUDE.md`：

```text
回答我时用更口语的语气。
我喜欢先看表格。
我的本地端口是 5174。
我常用 Windows PowerShell。
```

这些不是团队规则。

应该放进 `CLAUDE.local.md`。

示例：

```md
# CLAUDE.local.md

## Personal Preferences

- 默认用中文回复我
- 我本地使用 Windows PowerShell
- 本机开发端口优先使用 5174
- 解释方案时先给结论，再给命令
```

这个文件要加入 `.gitignore`：

```gitignore
CLAUDE.local.md
.claude/settings.local.json
```

判断标准很简单：

```text
帮助整个团队保持一致 -> 提交到仓库
只反映你个人习惯 -> 放 local
```

## 5. 第二层：`.claude/settings.json` 负责控制，不负责解释

`CLAUDE.md` 是引导层。

`.claude/settings.json` 是控制层。

官方 settings 文档里给过典型例子：可以用 `permissions.allow` 和 `permissions.deny` 控制工具、命令和文件访问，例如允许 `npm run test`，拒绝读取 `.env` 或 secrets。

对 4SAPI 项目，建议从安全控制开始。

示例：

```json
{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "permissions": {
    "allow": [
      "Bash(pnpm lint)",
      "Bash(pnpm test *)",
      "Bash(pnpm build)",
      "Bash(rg *)",
      "Bash(git status)",
      "Bash(git diff *)"
    ],
    "deny": [
      "Read(./.env)",
      "Read(./.env.*)",
      "Read(./secrets/**)",
      "Read(./config/production.json)",
      "Bash(rm -rf *)",
      "Bash(curl *|sh)",
      "Bash(wget *|sh)"
    ]
  }
}
```

这类配置特别适合模型中转站项目。

因为项目里很容易出现：

- 4SAPI Key
- 上游模型 Key
- 生产配置
- 计费配置
- 客户数据

Claude Code 可以帮你写代码，但不应该随便读 secrets。

## 6. 第三层：`rules/` 放模块化指令

当 `CLAUDE.md` 越写越长，就该拆 `rules/`。

Claude Code 官方文档也建议大项目用 `.claude/rules/` 组织多文件规则，并且可以按路径或文件类型让规则更有针对性，减少上下文噪音。

推荐：

```text
.claude/rules/
  llm-gateway.md
  security.md
  testing.md
  frontend.md
  backend-api.md
```

### `llm-gateway.md`

```md
# LLM Gateway Rules

适用于所有模型调用相关代码。

## 必须遵守

- 所有模型调用必须经过 `packages/llm/`
- 使用环境变量读取 `LLM_BASE_URL`、`LLM_API_KEY`、`LLM_MODEL`
- 兼容 OpenAI-style `/chat/completions`
- 支持 4SAPI 作为默认中转入口
- provider 需要返回统一错误类型

## 禁止

- 不要在前端直接请求 4SAPI
- 不要把 API Key 写入测试快照
- 不要把完整 prompt 写入日志
```

### `security.md`

```md
# Security Rules

- 不读取 `.env`、`secrets/`、生产配置
- 不输出 API Key、token、cookie
- 日志必须脱敏
- 涉及支付、权限、生产数据时先停下来说明风险
- 删除、移动、重命名文件前必须列清单
```

### `testing.md`

```md
# Testing Rules

修改 LLM provider 后至少运行：

```bash
pnpm test packages/llm
```

前端聊天页面改动后运行：

```bash
pnpm test apps/web
pnpm build
```

测试里不要调用真实 4SAPI Key，使用 mock 或测试 Key。
```

`rules/` 的价值是：

```text
让项目规则能持续维护，而不是塞爆 CLAUDE.md。
```

## 7. 第四层：`hooks/` 放自动执行的硬规则

hooks 和 commands 最大区别：

```text
hooks 是自动触发。
commands 是你主动调用。
```

Claude Code 官方 hooks 文档说明，hooks 可以在会话生命周期、用户提交 prompt、工具调用前后、停止前等事件触发，用于阻止危险操作、运行检查、发送通知或做自动化。

对 4SAPI 项目，hooks 很适合做安全护栏。

### `block-secret-reads.sh`

```bash
#!/usr/bin/env bash

input="$(cat)"

if echo "$input" | grep -E '\\.env|secrets|production\\.json|api_key|token' >/dev/null; then
  echo "Blocked: this operation may access secrets or production config."
  exit 2
fi

exit 0
```

### `remind-tests-before-stop.sh`

```bash
#!/usr/bin/env bash

echo "Reminder: if LLM provider files changed, run pnpm test packages/llm and update status."
exit 0
```

示例 settings 片段：

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Read|Bash",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/block-secret-reads.sh"
          }
        ]
      }
    ],
    "Stop": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/remind-tests-before-stop.sh"
          }
        ]
      }
    ]
  }
}
```

注意：

hooks 适合机械规则。

不适合复杂判断。

比如“禁止读取 `.env`”适合 hook。

“这个架构是不是优雅”不适合 hook。

## 8. 第五层：`commands/` 放轻量可复用工作流

Claude Code 新文档里提到，自定义 commands 已经和 skills 合流，但旧的 `.claude/commands/` 文件仍然可用。

所以你可以继续把轻量任务放在 commands。

适合：

- PR 审查
- 生成变更摘要
- 运行 4SAPI smoke test
- 检查 provider 配置
- 总结成本风险

### `review-llm-provider.md`

```md
# Review LLM Provider

请审查当前 diff 中和 LLM provider 相关的改动。

重点检查：

1. 是否仍然通过 4SAPI / OpenAI-compatible base_url 调用
2. 是否泄露 API Key
3. 是否记录完整 prompt
4. 是否破坏旧模型 fallback
5. 是否更新 `.env.example`
6. 是否补充测试
7. 是否有可读错误提示

先列问题，不要直接修改。
```

### `test-4sapi-smoke.md`

```md
# Test 4SAPI Smoke

请执行一次最小 smoke test。

要求：

1. 不读取真实 `.env`
2. 使用测试环境变量或 mock
3. 验证 base_url、model、timeout、错误处理
4. 输出测试命令和结果
5. 不把 Key 打印到终端
```

commands 的价值是：

```text
把常用 prompt 变成团队统一入口。
```

不要让每个人都手写一套“帮我审查一下 provider”。

## 9. 第六层：`skills/` 放完整能力包

skills 比 commands 更重。

Claude Code 官方文档说得很清楚：当你反复粘贴同一套说明、清单或多步流程时，就该做 skill。Skill 的正文只有在使用时才加载，长参考材料平时几乎不消耗上下文。

适合做 skill 的任务：

- 4SAPI provider 迁移
- AI 成本审计
- 日志脱敏检查
- 多模型路由设计
- 发布前 LLM 接入检查

### `skills/4sapi-provider-migration/SKILL.md`

```md
---
name: 4sapi-provider-migration
description: "把现有 OpenAI/Claude/Gemini 调用迁移到 4SAPI 大模型 API 中转站，并保持测试、文档和回滚路径完整。"
---

# 4SAPI Provider Migration

## When To Use

当用户要把现有模型调用统一迁移到 4SAPI 或 OpenAI-compatible gateway 时使用。

## Workflow

1. 先定位所有模型调用位置
2. 判断 SDK 是否支持自定义 base_url
3. 设计最小改动方案
4. 新增或复用 provider 层
5. 更新 `.env.example`
6. 添加成功、401、429、5xx、timeout 测试
7. 更新 README
8. 输出回滚方案

## Safety

- 不读取真实 `.env`
- 不打印 API Key
- 不改生产配置
- 不重构无关模块

## Output

- 修改文件清单
- 验证命令
- 风险点
- 回滚方式
```

### `skills/ai-cost-audit/SKILL.md`

```md
---
name: ai-cost-audit
description: "审计项目中的大模型调用成本，找出高 token、重复重试、模型路由不合理和日志缺口。"
---

# AI Cost Audit

## Workflow

1. 查找所有 LLM 调用
2. 统计是否记录 token
3. 检查是否设置 max_tokens
4. 检查重试上限
5. 检查是否所有任务都使用最贵模型
6. 给出 4SAPI 路由优化建议

## Output

用表格输出：

- 文件
- 风险
- 成本影响
- 建议
- 是否阻塞上线
```

skills 的关键是“完整工作流”。

如果一个文件就能说清楚，用 command。

如果需要多步、配套文档、检查清单，用 skill。

## 10. 第七层：`agents/` 放专用角色

subagent 适合做专门角色。

官方扩展文档里也把 skill 和 subagent 区分开：skill 是可复用内容，subagent 是隔离运行的 worker。

适合：

- 安全审计员
- LLM provider reviewer
- 文档作者
- 测试工程师
- 成本分析员

### `agents/security-auditor.md`

```md
---
name: security-auditor
description: "审计代码中的 API Key、日志隐私、权限边界和生产配置风险。"
tools:
  - Read
  - Grep
  - Glob
---

# Security Auditor

你是安全审计子代理。

只做审计，不直接修改代码。

重点检查：

- API Key 是否可能泄露
- `.env` 是否被读取或提交
- 日志是否记录完整 prompt
- 前端是否暴露模型 Key
- 生产配置是否被改动
- 4SAPI Key 是否只在后端使用

输出：

- 阻塞问题
- 非阻塞建议
- 需要人工确认的问题
```

### `agents/llm-provider-reviewer.md`

```md
---
name: llm-provider-reviewer
description: "审查 LLM provider、模型路由、fallback、4SAPI 接入和错误处理。"
tools:
  - Read
  - Grep
  - Glob
---

# LLM Provider Reviewer

请检查模型调用链路是否清晰：

前端 -> 后端 -> provider -> 4SAPI -> 上游模型

重点看：

- 是否支持自定义 base_url
- 是否有模型 fallback
- 是否有 401/429/5xx/timeout 错误处理
- 是否更新文档和环境变量示例
- 是否有测试覆盖
```

agents 不要乱建。

如果两个 agent 角色高度重叠，合并。

如果一个 agent 很少用，删除。

## 11. 项目级 vs 用户级：哪些该提交，哪些不该提交？

官方 `.claude` 目录文档强调：

```text
项目文件可以提交到 Git 给团队共享。
~/.claude 是个人配置，跨项目生效。
```

建议：

| 内容 | 放哪里 | 是否提交 |
| --- | --- | --- |
| 项目架构说明 | `CLAUDE.md` | 是 |
| 团队测试命令 | `CLAUDE.md` / `rules/testing.md` | 是 |
| 4SAPI provider 规则 | `rules/llm-gateway.md` | 是 |
| 安全 hooks | `.claude/hooks/` | 是 |
| 个人语言偏好 | `CLAUDE.local.md` | 否 |
| 个人本地端口 | `CLAUDE.local.md` | 否 |
| 个人全局 skill | `~/.claude/skills/` | 否 |
| 团队共用 skill | `.claude/skills/` | 是 |
| 团队共用 agent | `.claude/agents/` | 是 |
| 私人实验 agent | `~/.claude/agents/` | 否 |

判断标准：

```text
帮助团队一致交付 -> 项目级
只帮助个人习惯更舒服 -> 用户级
```

## 12. 4SAPI 项目的渐进式成长路径

不要一上来就把 `.claude/` 填满。

推荐路线：

### 阶段一：最小可用

```text
CLAUDE.md
.claude/settings.json
```

解决：

- 项目说明
- 常用命令
- 基础权限
- 不读取 secrets

### 阶段二：规则模块化

```text
.claude/rules/
```

解决：

- 前端规则
- 后端规则
- 4SAPI provider 规则
- 测试规则
- 安全规则

### 阶段三：自动化护栏

```text
.claude/hooks/
```

解决：

- 阻止危险命令
- 阻止读取 secret
- 停止前提醒测试
- 关键文件改动后提醒审查

### 阶段四：复用工作流

```text
.claude/commands/
.claude/skills/
```

解决：

- PR 审查
- provider 迁移
- 成本审计
- 发布前检查

### 阶段五：专用子代理

```text
.claude/agents/
```

解决：

- 安全审计
- 文档写作
- LLM provider 审查
- 测试补齐

这条路的核心是：

```text
先让 Claude Code 稳定，再让它强大。
```

## 13. 和 4SAPI 的结合点

为什么这篇要放在大模型 API 中转站系列？

因为 Claude Code 工程化目录，不只是“写代码更舒服”。

它可以直接提升 4SAPI 接入项目的稳定性。

### 结合点一：统一 provider 规则

`rules/llm-gateway.md` 规定所有模型调用必须经过 provider 层。

这样不会出现：

- 某个页面直连 OpenAI
- 某个脚本直连 Claude
- 某个后台直连 Gemini
- 只有一部分流量走 4SAPI

### 结合点二：阻止 Key 泄露

`settings.json` 和 hooks 阻止读取 `.env`、输出 Key、把 Key 写进测试。

### 结合点三：成本审计 skill

`ai-cost-audit` skill 可以定期检查：

- 是否所有任务都用最贵模型
- 是否缺少 token 统计
- 是否重试过多
- 是否有 fallback 成本

### 结合点四：provider review agent

`llm-provider-reviewer` 可以专门审查：

- base_url 是否可配置
- 4SAPI endpoint 是否统一
- 错误处理是否完整
- `.env.example` 是否更新
- 文档是否同步

这些都能把 4SAPI 从“一个 API 地址”升级成“项目级模型治理层”。

## 14. 常见错误

### 错误一：`CLAUDE.md` 变成垃圾桶

所有内容都塞进去，最后谁也读不懂。

正确做法：

```text
全局内容留在 CLAUDE.md。
专项内容拆到 rules。
流程内容做成 commands 或 skills。
```

### 错误二：hooks 做复杂判断

hooks 应该做确定性动作。

不要让 hooks 承担架构判断。

### 错误三：commands 和 skills 重复

一个任务如果只有一个 prompt，用 command。

如果有多步流程和配套文档，用 skill。

### 错误四：agents 建太多

每个 agent 都应该有明确角色。

不要建一堆听起来高级但很少使用的 agent。

### 错误五：个人配置提交进仓库

`CLAUDE.local.md`、`settings.local.json` 必须进 `.gitignore`。

### 错误六：没有清理废弃文件

每月检查一次：

- 哪些 command 没人用
- 哪些 skill 重复
- 哪些 agent 角色重叠
- 哪些 hooks 已经过时

`.claude/` 本身也需要维护。

## 15. 团队落地清单

你可以直接复制这张 checklist：

```text
1. 根目录是否有简洁 CLAUDE.md
2. 是否把个人偏好放进 CLAUDE.local.md
3. .gitignore 是否包含 CLAUDE.local.md 和 settings.local.json
4. settings.json 是否禁止读取 .env 和 secrets
5. 4SAPI provider 规则是否写入 rules/llm-gateway.md
6. 测试命令是否写清楚
7. 是否有 provider review command
8. 是否有成本审计 skill
9. 是否有 security-auditor agent
10. hooks 是否只做确定性安全护栏
11. 是否定期清理无用 commands/skills/agents
12. 是否所有模型调用都经过后端 provider 层
```

如果这 12 条都做到，Claude Code 在团队项目里会稳定很多。

## 16. 最后总结

`.claude/` 目录不是装饰品。

它是 Claude Code 的项目操作系统。

好的结构应该能回答这些问题：

- 项目级指令在哪里？
- 个人偏好在哪里？
- 模块化规则在哪里？
- 自动化安全护栏在哪里？
- 可复用工作流在哪里？
- 完整能力包在哪里？
- 专用子代理在哪里？
- 哪些文件该提交，哪些不该提交？

对 4SAPI 这类大模型 API 中转站项目来说，`.claude/` 还能进一步解决：

- API Key 安全
- provider 统一
- 日志脱敏
- 成本审计
- fallback 检查
- 团队协作一致性

一句话总结：

```text
最高效的 .claude/ 目录，不是功能最多，而是每一层都有清晰职责。
```

如果你正在用 Claude Code 接入 4sapi.com，建议先从这两个文件开始：

```text
CLAUDE.md
.claude/settings.json
```

然后按需增加：

```text
rules -> hooks -> commands -> skills -> agents
```

这样 Claude Code 才会从“能写代码的助手”，变成“能在团队工程里稳定交付的协作系统”。

## 资料来源与延伸阅读

- [Explore the .claude directory - Claude Code Docs](https://code.claude.com/docs/en/claude-directory)
- [How Claude remembers your project - Claude Code Docs](https://code.claude.com/docs/en/memory)
- [Claude Code settings - Claude Code Docs](https://code.claude.com/docs/en/settings)
- [Hooks reference - Claude Code Docs](https://code.claude.com/docs/en/hooks)
- [Extend Claude with skills - Claude Code Docs](https://code.claude.com/docs/en/skills)
- [Extend Claude Code - Claude Code Docs](https://code.claude.com/docs/en/features-overview)
