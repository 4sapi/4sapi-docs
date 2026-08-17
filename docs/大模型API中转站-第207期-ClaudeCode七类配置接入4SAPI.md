---
title: "【大模型API中转站】第207期 Claude Code配置 | 7类入口少走弯路"
category: 人工智能
tags:
  - 大模型API中转站
  - Claude Code
  - CLAUDE.md
  - Skills
  - Hooks
  - Subagents
  - 4SAPI
  - 企业级大模型接入
description: "Anthropic 官方把 Claude Code 的 7 类配置入口讲清楚了：CLAUDE.md、Rules、Skills、Subagents、Hooks、Output Styles 和 Append System Prompt。本文按企业落地视角重写一版：什么该放哪里，怎么减少 token 浪费，以及如何用 4SAPI 统一 Claude Code 的模型入口、日志审计和成本治理。"
---

# 【大模型API中转站】第207期 Claude Code配置 | 7类入口少走弯路

本文是【大模型API中转站】系列的第207篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

很多人用 Claude Code，第一反应是写一份很长的提示词。

比如：

```text
你是资深工程师。
请遵守我的代码规范。
请先读项目。
请不要乱改。
请每次都跑测试。
请用中文解释。
```

问题是，这些话写得越多，效果不一定越好。

Claude Code 不是只有一个“提示词入口”。

它有一整套配置体系。

Anthropic 官方在《Steering Claude Code: CLAUDE.md files, skills, hooks, rules, subagents and more》里把这套体系拆成 7 类：

```text
CLAUDE.md
Rules
Skills
Subagents
Hooks
Output Styles
Append System Prompt
```

这 7 类东西的区别，不是名字不同。

真正区别是三件事：

```text
什么时候加载进上下文。
长会话压缩后会不会还在。
会不会持续消耗 token。
```

如果你放错位置，结果就是：

```text
该常驻的规则被忘掉。
不该常驻的流程天天吃 token。
该确定执行的动作变成模型“尽量记得”。
该隔离的调研挤爆主对话。
```

这篇不做逐字翻译。

我按企业级大模型接入的角度，给你一版更实用的判断方法。

如果你还通过 4SAPI 接 Claude、GPT、Gemini、DeepSeek 等模型，也可以把这套配置和模型路由、Key 权限、日志审计、预算控制串起来。

## 1. 先记住一句话

Claude Code 配置不是越多越好。

而是要放对地方。

可以先用这张表：

| 需求 | 放哪里 |
| --- | --- |
| 项目事实、构建命令、目录说明 | `CLAUDE.md` |
| 路径级编码规范 | `.claude/rules/` |
| 部署、审查、发布这类流程 | `.claude/skills/` |
| 深度调研、日志分析、依赖审计 | `.claude/agents/` |
| 必须自动发生的动作 | Hooks |
| 全局输出风格和角色变化 | Output Styles |
| 单次临时补充要求 | `--append-system-prompt` |

再压缩成一句：

```text
事实放 CLAUDE.md。
约束放 Rules。
流程放 Skills。
重活放 Subagents。
硬门禁放 Hooks。
风格慎用 Output Styles。
临时要求用 Append System Prompt。
```

这就是 7 类配置的基本心法。

## 2. CLAUDE.md：项目入口，不是项目百科

`CLAUDE.md` 是 Claude Code 最常见的配置文件。

它适合放：

```text
项目技术栈。
常用命令。
目录结构。
命名规范。
团队约定。
测试方式。
高风险目录提醒。
```

它不适合放：

```text
完整部署流程。
几十条 code review 清单。
复杂发布 runbook。
长篇项目历史。
模型价格表。
所有团队的所有偏好。
```

原因很简单。

根目录 `CLAUDE.md` 会话开始就加载。

你写进去的每一行，不管当前任务用不用，都会占上下文。

官方也提醒，团队共享的 `CLAUDE.md` 很容易越写越大。

我建议把它控制在 200 行以内。

更好的写法是“索引型”：

```markdown
# CLAUDE.md

## Project
- Frontend: Next.js
- Backend: FastAPI
- Package manager: pnpm

## Commands
- Install: pnpm install
- Test: pnpm test
- Lint: pnpm lint

## Read More
- Frontend rules: docs/frontend.md
- API rules: docs/api.md
- Deploy workflow: use skill `/deploy-check`
- Model gateway rules: docs/model-gateway.md
```

它不是图书馆。

它是导航牌。

## 3. Rules：让规则只在该出现时出现

Rules 适合路径级约束。

比如：

```text
所有 API handler 必须做输入校验。
所有 migration 只能追加，不能改历史。
所有 test 文件必须包含边界场景。
所有 model-gateway 文件必须记录 request_id。
```

这类规则如果写进根 `CLAUDE.md`，就会一直占 token。

但其实你改前端样式时，不需要加载 API handler 规则。

用 `.claude/rules/` 可以按路径加载。

示例：

```yaml
---
paths:
  - "src/api/**"
  - "**/*.handler.ts"
---
所有 API handler 必须使用 schema 校验输入。
错误返回必须包含 request_id。
不要直接暴露上游模型错误原文。
```

如果你的项目接了 4SAPI，可以给模型网关目录单独写一条：

```yaml
---
paths:
  - "src/model-gateway/**"
  - "src/llm/**"
---
模型调用必须通过统一网关。
不要硬编码 API Key、Base URL 和模型名。
每次调用必须记录 provider、model、request_id、cost_bucket。
```

这比在全局写一句“注意成本治理”强得多。

## 4. Skills：把重复流程封装起来

Skills 适合流程。

比如：

```text
代码审查。
上线前检查。
生成发布说明。
排查 4SAPI 429。
整理模型调用成本日报。
迁移一批接口。
```

这些东西不该塞进 `CLAUDE.md`。

因为流程只有用到时才需要。

Skill 的优势是按需加载。

会话开始时，Claude 只知道这个 Skill 的名称和描述。

真正触发时，才把完整步骤读进来。

例如可以做一个：

```text
.claude/skills/4sapi-cost-review/SKILL.md
```

内容写：

```markdown
---
name: 4sapi-cost-review
description: Analyze 4SAPI model call logs, identify high-cost tasks, repeated retries, missing cost_bucket, and route optimization opportunities.
---

# 4SAPI Cost Review

## Steps
1. Read model call logs.
2. Group by project, model, task_type, cost_bucket.
3. Find retries, fallback spikes, and missing metadata.
4. Suggest route changes.
5. Do not edit production routing config without human approval.
```

这就比每次手写“帮我分析一下成本”稳定得多。

## 5. Subagents：把重活丢到独立上下文

Subagent 最适合做会污染主对话的任务。

例如：

```text
全仓库搜索一个 bug。
分析几万行日志。
检查依赖安全风险。
对比三个方案优劣。
扫描模型调用链路。
```

它的好处是独立上下文。

主对话不需要承载所有中间过程。

只拿最后结论。

企业场景里可以这样分：

| Subagent | 任务 |
| --- | --- |
| `codebase-researcher` | 搜索代码结构和调用链 |
| `security-reviewer` | 查密钥、权限、危险操作 |
| `cost-auditor` | 看 4SAPI 调用日志和成本 |
| `api-contract-checker` | 查接口兼容和错误码 |
| `release-risk-reviewer` | 发布前风险审查 |

如果结合 4SAPI，可以让不同 Subagent 使用不同模型：

```text
研究型 agent：长上下文模型。
安全审查 agent：强推理模型。
成本分析 agent：中等模型。
摘要 agent：低成本模型。
```

这就是模型路由的价值。

不是所有任务都需要最贵模型。

## 6. Hooks：把“必须做”变成自动执行

前面几类配置，本质还是给模型看的。

模型可能遵守，也可能在长会话里漏掉。

Hooks 不一样。

Hooks 是确定性自动化。

到了触发点就执行。

适合：

```text
编辑后自动格式化。
提交前跑 lint。
工具调用前拦截危险命令。
任务结束后写日志。
压缩前备份上下文。
调用模型前检查 Key 和 budget。
```

比如你不想让 Agent 删除数据库迁移文件。

不要只写：

```text
请永远不要删除 migration。
```

要用 Hook 或权限规则拦截。

在模型调用治理里，也一样。

不要只写：

```text
请走 4SAPI。
```

应该在代码和配置层保证：

```text
生产环境只允许走统一 Base URL。
禁止在代码里出现裸 API Key。
模型调用必须带 task_type。
缺少 cost_bucket 的请求直接失败。
```

提醒是软的。

门禁是硬的。

## 7. Output Styles：别轻易改默认人格

Output Styles 可以改变 Claude Code 的输出风格。

但它不是普通偏好。

它会进入系统提示词层。

如果乱改，可能影响 Claude Code 默认的软件工程行为。

比如：

```text
如何控制修改范围。
什么时候加注释。
遇到安全风险怎么处理。
完成前要不要验证。
```

所以我的建议是：

```text
个人学习可以试。
团队项目慎用。
生产仓库尽量用内置风格。
```

如果只是想让它“回答短一点”或“输出中文”，不要动 Output Style。

用临时追加提示更稳。

## 8. Append System Prompt：单次临时补充

`--append-system-prompt` 适合临时需求。

比如：

```text
这次只输出审查报告，不修改文件。
这次所有说明用中文。
这次按表格输出对比。
这次只关注 4SAPI 接入风险。
```

它是追加，不是替换。

比自定义 Output Style 风险低。

但也别叠太多。

提示越长，指令互相打架的概率越高。

## 9. 4SAPI 在这套配置里的位置

Claude Code 的 7 类配置，解决的是：

```text
Claude 怎么工作。
```

4SAPI 解决的是：

```text
模型调用怎么进入企业治理。
```

两者不是替代关系。

而是上下两层。

```text
Claude Code 配置层：
CLAUDE.md / Rules / Skills / Subagents / Hooks

模型调用治理层：
4SAPI Base URL / Key 分组 / 模型路由 / 日志审计 / 预算控制
```

一个成熟团队应该把两层接起来：

```text
CLAUDE.md 告诉 Agent：模型调用规则在哪里。
Rules 限定 model-gateway 目录必须记录日志。
Skills 封装成本分析、错误排查、上线检查。
Subagents 做深度审计和日志分析。
Hooks 拦截硬编码 Key、危险命令和缺少预算字段的请求。
4SAPI 记录真实模型、真实成本、真实错误码。
```

这样 Claude Code 才不只是“更会写代码”。

而是进入可控、可查、可复盘的企业级大模型接入流程。

## 10. 一套推荐目录

可以这样组织：

```text
project/
  CLAUDE.md
  docs/
    model-gateway.md
    cost-policy.md
    security-policy.md
  .claude/
    rules/
      api-handlers.md
      model-gateway.md
      tests.md
    skills/
      4sapi-cost-review/
        SKILL.md
      release-check/
        SKILL.md
      api-error-triage/
        SKILL.md
    agents/
      cost-auditor.md
      security-reviewer.md
      log-investigator.md
    settings.json
```

目录不是为了好看。

而是为了把规则分层。

否则最后一定会变成：

```text
一个几千行 CLAUDE.md。
一堆没人维护的提示词。
一堆说不清从哪里来的模型账单。
```

## 11. 总结

Claude Code 的配置体系，本质是把指令放到正确的位置。

```text
CLAUDE.md 管事实。
Rules 管路径约束。
Skills 管流程。
Subagents 管隔离重活。
Hooks 管确定性门禁。
Output Styles 管全局风格。
Append System Prompt 管单次补充。
```

如果你只是个人玩一玩，先写好 `CLAUDE.md` 和几个 Skills 就够。

如果你是团队使用，必须再加：

```text
Rules。
Hooks。
Subagents。
4SAPI 模型网关。
日志审计。
预算控制。
```

一句话：

```text
Claude Code 负责把 Agent 配好。
4SAPI 负责把模型调用管好。
```

这两层一起上，才是真正适合生产环境的 AI 编程工作流。

## 资料与延伸阅读

- Anthropic 官方博客：Steering Claude Code：https://claude.com/blog/steering-claude-code-skills-hooks-rules-subagents-and-more
- Claude Code 文档：https://code.claude.com/docs
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 文档：https://4sapi.apifox.cn/
