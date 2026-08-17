---
title: "【大模型API中转站】第54期 Karpathy Skills | 少乱改代码"
category: 人工智能
tags:
  - 大模型API中转站
  - Claude Code
  - Cursor
  - CLAUDE.md
  - Agent
  - 4SAPI
description: "拆解 multica-ai/andrej-karpathy-skills：它不是模型接入工具，而是一套让 Claude Code、Cursor 和编码 Agent 少做错误假设、少过度工程、少乱改代码的行为规范。"
---

# 【大模型API中转站】第54期 Karpathy Skills | 少乱改代码

本文是【大模型API中转站】系列的第54篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一期我们写了 PilotDeck 接入 4SAPI。

PilotDeck 解决的是 Agent 工作台问题。

这一篇看一个更小、更轻，但很实用的仓库：

```text
multica-ai/andrej-karpathy-skills
```

它不是模型，不是中转站，也不是一个完整 IDE。

它的核心其实很朴素：

```text
用一份 CLAUDE.md / Cursor Rule / Skill，
约束编码 Agent 少做错误假设、少过度工程、少乱改无关代码。
```

如果你经常用 Claude Code、Cursor、Codex 或其他 Coding Agent，应该很熟悉这些场景：

```text
让它修一个小 bug，它顺手重构半个文件。
让它加一个字段，它新建一套抽象层。
让它优化搜索，它不问指标就加缓存、索引、异步队列。
让它改一行逻辑，它顺手格式化全文件。
让它修复问题，它没写复现测试就开始猜。
```

这类问题不一定是模型能力差。

很多时候，是项目规则不够明确。

`andrej-karpathy-skills` 做的事，就是把这些“高级工程师会提醒的边界”写成一套可复用规则。

## 1. 这个仓库能不能写

能写，而且很适合写。

但要先说清楚定位：

```text
它不是 4SAPI 这类模型 API 中转站。
它不提供 Base URL。
它不直接调用模型。
它也不是一个完整 Agent 平台。
```

它更像一份 Agent 行为守则。

官方 README 里把它描述为：

```text
A single CLAUDE.md file to improve Claude Code behavior
```

也就是通过一份 `CLAUDE.md` 改善 Claude Code 的编码行为。

同时它还提供了：

| 文件/目录 | 用途 |
| --- | --- |
| `CLAUDE.md` | 给 Claude Code 或其他支持根指令文件的工具使用 |
| `.cursor/rules/karpathy-guidelines.mdc` | Cursor 项目规则，带 `alwaysApply: true` |
| `skills/karpathy-guidelines/SKILL.md` | 可复用 Skill 形式 |
| `.claude-plugin/plugin.json` | Claude Code 插件元信息 |
| `EXAMPLES.md` | 展示错误写法和推荐写法的对比例子 |

这类仓库的价值不在“安装后能力暴涨”，而在“把 Agent 的默认行为往工程纪律上拉”。

一句话总结：

```text
它是 Coding Agent 的安全驾驶规则，不是发动机。
```

## 2. 能不能接入 4SAPI

严格说：

```text
它本身不能直接接入 4SAPI。
```

原因很简单。

这个仓库没有模型调用层，没有 provider 配置，也没有 API 请求代码。

它只是规则文件。

但是，它可以和 4SAPI 搭配使用。

结构是：

```text
Claude Code / Cursor / Codex / PilotDeck
  -> 读取 Karpathy Guidelines
  -> 通过 4SAPI 调用模型
  -> 在项目里执行编码任务
```

也就是说：

```text
4SAPI 管模型、Key、成本和日志。
Karpathy Skills 管 Agent 的编码行为。
Coding Agent 管实际读写文件、运行测试和提交变更。
```

不要把这三层混在一起。

更清楚地看：

| 层级 | 工具 | 负责什么 |
| --- | --- | --- |
| 模型网关 | 4SAPI | Base URL、API Key、模型路由、费用统计 |
| Agent 平台 | Claude Code、Cursor、Codex、PilotDeck | 读文件、改代码、跑命令、生成 diff |
| 行为规范 | `andrej-karpathy-skills` | 少猜、少膨胀、少乱改、可验证 |

所以这篇文章不是“Karpathy Skills 接入 4SAPI 教程”。

更准确的主题是：

```text
用 Karpathy Skills 约束 Agent 行为，再用 4SAPI 管模型调用。
```

## 3. 它解决的四类 Agent 编码问题

官方把问题拆成四类。

### 问题一：错误假设

用户说：

```text
添加导出用户数据功能。
```

很多 Agent 会直接假设：

```text
导出所有用户
导出 JSON 和 CSV
保存到本地文件
字段包含 id、email、name
```

但这些假设都可能是错的。

真正应该先问：

```text
导出范围是什么？
哪些字段可以导出？
是否涉及隐私？
是后台任务、接口下载，还是管理员页面按钮？
数据量有多大？
```

这就是第一条原则：

```text
Think Before Coding
```

编码前先说清假设，不要把模糊需求默默解释成自己的版本。

### 问题二：过度工程

用户说：

```text
写一个折扣计算函数。
```

Agent 可能会写出：

```text
DiscountStrategy
PercentageDiscount
FixedDiscount
DiscountConfig
DiscountCalculator
Validator
Factory
```

但当前需求可能只需要：

```python
def calculate_discount(amount: float, percent: float) -> float:
    return amount * (percent / 100)
```

这就是第二条原则：

```text
Simplicity First
```

能用 50 行解决的事情，不要提前写成 200 行框架。

复杂性不是不能加，而是等需求真的出现再加。

### 问题三：乱改无关代码

用户说：

```text
修复空 email 导致校验器崩溃的问题。
```

Agent 可能顺手做了这些事：

```text
改注释
改格式
加类型标注
调整用户名校验
换引号风格
重写整个函数
```

结果 diff 很大，review 很痛苦，还可能引入新 bug。

这就是第三条原则：

```text
Surgical Changes
```

只碰必须碰的代码。

如果看到无关问题，可以提出来，但不要顺手改。

### 问题四：没有验证目标

用户说：

```text
修复认证系统。
```

Agent 如果直接开始“检查并优化”，很容易变成一团糊。

更好的方式是先把任务变成可验证目标：

```text
如果问题是修改密码后旧 session 仍然有效：
1. 写一个复现测试：修改密码后旧 session 应失效
2. 运行测试，确认它失败
3. 修改逻辑
4. 再运行测试，确认通过
5. 跑现有认证测试，确认没有回归
```

这就是第四条原则：

```text
Goal-Driven Execution
```

不要只说“把它修好”。

要定义“怎样算修好”。

## 4. 安装方式一：Claude Code 插件

官方 README 推荐用 Claude Code 插件方式安装。

命令是：

```text
/plugin marketplace add forrestchang/andrej-karpathy-skills
```

然后安装：

```text
/plugin install andrej-karpathy-skills@karpathy-skills
```

这个方式的好处是：

```text
跨项目可用
不用每个仓库都复制 CLAUDE.md
后续更新更容易
```

适合经常用 Claude Code 做开发的人。

注意：用户给的仓库地址是 `multica-ai/andrej-karpathy-skills`，而官方 README 当前展示的插件市场命令仍是 `forrestchang/andrej-karpathy-skills`。

实际安装时，以仓库最新 README 为准。

如果命令不可用，就改用下面的按项目安装方式。

## 5. 安装方式二：按项目放 CLAUDE.md

最稳的方式是把规则写进项目根目录的 `CLAUDE.md`。

新项目可以直接下载：

```bash
curl -o CLAUDE.md https://raw.githubusercontent.com/multica-ai/andrej-karpathy-skills/main/CLAUDE.md
```

已有项目建议不要直接覆盖。

更稳的做法是追加到现有文件后面：

```bash
echo "" >> CLAUDE.md
curl https://raw.githubusercontent.com/multica-ai/andrej-karpathy-skills/main/CLAUDE.md >> CLAUDE.md
```

如果你已经有项目规则，建议合并成这样：

```markdown
# Project Instructions

## Project-Specific Rules

- 使用 TypeScript strict mode。
- 所有 API endpoint 必须有测试。
- 错误处理遵循 `src/utils/errors.ts` 的现有模式。

## Agent Behavior Rules

- 编码前先说明假设。
- 不做需求之外的抽象。
- 只修改与任务直接相关的代码。
- 修 bug 时先写复现测试，再修复。
```

不要把 `CLAUDE.md` 当成口号墙。

它应该能直接约束 Agent 的行动。

## 6. 安装方式三：Cursor Rule

这个仓库也包含 Cursor 项目规则：

```text
.cursor/rules/karpathy-guidelines.mdc
```

官方说明里写了，这个规则带有：

```text
alwaysApply: true
```

也就是在 Cursor 打开该项目时自动生效。

如果想把它用到别的项目，复制到目标项目：

```text
your-project/
  .cursor/
    rules/
      karpathy-guidelines.mdc
```

适合 Cursor 用户的方式是：

```text
项目规则管当前仓库
个人规则管个人偏好
CLAUDE.md 管 Claude Code
```

不要假设所有工具都会读取同一个文件。

Cursor 默认不会读取 `.claude-plugin/`，也不会天然把 `CLAUDE.md` 当成 Cursor Rule。

## 7. 安装方式四：作为 Skill 使用

仓库里还有：

```text
skills/karpathy-guidelines/SKILL.md
```

这份文件有 Skill 的元信息：

```yaml
name: karpathy-guidelines
description: Behavioral guidelines to reduce common LLM coding mistakes.
```

它适合放进支持 Skill 的 Agent 环境里。

例如你有自己的技能目录，可以把它复制进去：

```text
~/.cursor/skills/karpathy-guidelines/SKILL.md
```

或者放进你自己的 Agent Skill 仓库。

这样它就不是某个项目的规则，而是一个可复用的“编码行为模式”。

建议触发场景写得更明确：

```text
当任务涉及写代码、改代码、重构、修 bug、做 code review 时启用。
目标是减少错误假设、过度工程、无关修改，并确保任务有可验证成功标准。
```

这样比“总是启用”更适合复杂工作流。

## 8. 和 4SAPI 怎么一起用

4SAPI 不需要直接配置到这个仓库里。

应该配置到你的 Agent 工具里。

例如上一期的 PilotDeck：

```env
PILOTDECK_MODEL=sapi4/your-model-name
PILOTDECK_API_KEY=sk-your-4sapi-key
PILOTDECK_API_URL=https://4sapi.com/v1
```

如果是支持 OpenAI-compatible provider 的工具，思路也类似：

```text
Base URL: https://4sapi.com/v1
API Key: 4SAPI 后台生成
Model: 从 4SAPI 模型列表复制
```

然后在项目里放：

```text
CLAUDE.md
.cursor/rules/karpathy-guidelines.mdc
AGENTS.md
```

最终形成三层：

```text
模型层：4SAPI
工具层：Claude Code / Cursor / Codex / PilotDeck
规则层：Karpathy Guidelines + 项目规则
```

这套组合的意义是：

```text
强模型负责复杂推理。
低成本模型负责简单整理。
规则文件负责约束行为。
测试和 diff 负责验证结果。
```

不要指望换一个更贵模型就自动少乱改。

模型越强，越需要清晰边界。

## 9. 推荐的模型分层

如果你通过 4SAPI 管多个模型，可以这样分工：

| 编码任务 | 推荐模型策略 |
| --- | --- |
| 读目录、列文件、找引用 | 低成本模型 |
| 整理 README、生成任务摘要 | 低成本到中等模型 |
| 修改业务逻辑 | 强代码模型 |
| 修复杂 bug | 强推理模型 + 代码模型 |
| 写测试 | 中高端代码模型 |
| Code review | 强推理模型 |
| 生成提交说明 | 低成本文本模型 |

但无论用哪个模型，规则都应该一样：

```text
先说明假设。
先定义成功标准。
只改相关文件。
能测试就测试。
不能测试就说明原因。
```

这就是 Karpathy Skills 的价值。

它让模型分层不只是“便宜模型做便宜活”，而是让每个模型都有同一套工程约束。

## 10. 建议加一份中文改造版

原仓库是英文规则。

如果你的团队主要用中文，可以在项目里加一段中文版本。

例如：

```markdown
## Coding Agent 行为规则

### 1. 编码前思考

- 不确定时先问，不要猜。
- 如果需求有多种解释，列出选项。
- 如果有更简单方案，主动说明。

### 2. 简洁优先

- 不添加需求之外的功能。
- 不为一次性代码创建抽象。
- 不为了“以后可能用到”提前设计复杂架构。

### 3. 精准修改

- 只修改与任务直接相关的代码。
- 不顺手重构、格式化、改注释。
- 发现无关问题可以报告，不要擅自修。

### 4. 目标驱动

- 修 bug 先写复现测试。
- 改功能先写验收标准。
- 多步骤任务每一步都要有验证方式。
```

这段可以和原始 `CLAUDE.md` 合并。

团队协作时，中文规则通常更容易被业务同事理解。

## 11. 最小实战流程

假设你有一个旧项目，想让 Coding Agent 帮你修 bug。

不要直接说：

```text
帮我修一下登录问题。
```

更好的提法是：

```text
请先阅读 CLAUDE.md，并遵守其中的 Karpathy Guidelines。

任务：修复用户修改密码后旧 session 仍然可用的问题。

成功标准：
1. 新增一个测试，复现旧 session 未失效的问题。
2. 测试先失败。
3. 修改逻辑，让测试通过。
4. 运行现有认证相关测试，确认没有回归。
5. 只修改认证相关文件，不做无关重构。
```

这时候 Agent 的工作路径会更清楚：

```text
读规则
  -> 明确假设
  -> 写复现测试
  -> 最小修改
  -> 跑测试
  -> 输出 diff 摘要
```

如果接了 4SAPI，可以把模型也分开：

```text
读文件和找引用：低成本模型
分析 bug 和修复逻辑：强代码模型
Review 和总结：强推理模型
```

这样既能控制成本，也能控制行为。

## 12. 如何判断它真的生效

官方 README 给了几个判断标准。

我们可以转成更适合团队的检查清单：

| 检查项 | 好现象 |
| --- | --- |
| diff 大小 | 只出现与需求相关的改动 |
| 抽象数量 | 没有为单次需求新增复杂框架 |
| 澄清问题 | 不确定时先问，而不是先写 |
| 测试 | 修 bug 前能先复现问题 |
| 风格 | 保持项目现有风格，不随手换格式 |
| 复盘 | 结束时说明改了什么、验证了什么、没验证什么 |

如果你发现 Agent 仍然经常：

```text
顺手重构
大面积格式化
不问就猜
不跑测试
不说不确定点
```

说明规则没有被正确读取，或者任务提示太模糊。

这时候先检查：

```text
CLAUDE.md 是否在项目根目录
Cursor Rule 是否放在 .cursor/rules
Agent 是否真的读取了项目规则
任务是否给了成功标准
模型上下文是否太短导致规则被挤出
```

## 13. 成本与风险提示

这个仓库本身没有 API 成本。

真正的成本来自你使用的 Coding Agent 和模型调用。

如果你接 4SAPI，建议按用途拆 Key：

```text
coding-agent-local
coding-agent-review
coding-agent-test
coding-agent-team
```

再按任务分模型：

```text
轻量模型：读目录、摘要、命名、提交说明
代码模型：写代码、改测试、修简单 bug
强推理模型：复杂 bug、架构权衡、PR review
```

风险主要有三类。

第一，规则冲突。

如果项目自己的 `CLAUDE.md` 说“可以主动重构”，而 Karpathy Guidelines 说“不要做无关重构”，Agent 可能混乱。

解决方法是写优先级：

```text
当前用户任务 > 项目特定规则 > Karpathy Guidelines > 通用偏好
```

第二，过度谨慎。

这些规则偏向谨慎，不适合所有小任务。

例如改一个拼写错误，不需要完整测试计划。

第三，错把规则当能力。

规则只能约束行为，不能让模型突然懂你的业务。

复杂业务仍然需要：

```text
清楚的需求
可运行的测试
可靠的文档
可审查的 diff
人工 review
```

## 14. 最小落地方案

如果你今天只想试一下，按这个顺序做：

```text
1. 在一个测试项目根目录放 CLAUDE.md
2. 如果用 Cursor，再复制 .cursor/rules/karpathy-guidelines.mdc
3. 配置你的 Agent 通过 4SAPI 调用模型
4. 选一个小 bug，不要选大重构
5. 给出明确成功标准
6. 要求 Agent 修改前说明假设，修改后列 diff 和测试结果
```

最小 `CLAUDE.md` 可以只保留四句话：

```text
编码前先说明假设。
优先写最简单能工作的方案。
只修改与任务直接相关的代码。
每个任务都要有可验证成功标准。
```

最小测试任务：

```text
请修复这个函数在空输入下报错的问题。
成功标准：
1. 新增一个空输入测试。
2. 测试先失败。
3. 最小修改让测试通过。
4. 不修改其他函数。
```

最小 4SAPI 配置思路：

```text
Base URL: https://4sapi.com/v1
API Key: 用 coding-agent-local
Model: 选择一个适合代码任务的模型
```

做到这里，你就能看出规则有没有用。

## 15. 总结

`multica-ai/andrej-karpathy-skills` 值得写，但不要把它写成模型接入教程。

它真正解决的是 Coding Agent 的行为问题：

```text
少猜。
少膨胀。
少乱改。
多验证。
```

它和 4SAPI 的关系是：

```text
Karpathy Skills 约束 Agent 怎么做。
4SAPI 管 Agent 调哪个模型、花多少钱、留下什么日志。
```

如果你已经在用 Claude Code、Cursor、Codex 或 PilotDeck，这个仓库很适合当作每个项目的基础规则。

一句话总结：

```text
强模型能提高上限，规则文件能减少下限事故；两者一起用，Coding Agent 才更像工程助手，而不是会写代码的随机数发生器。
```

建议先从一个小项目试。

把规则放进去，让 Agent 修一个真实 bug，看它的 diff 是否更小、问题是否问得更早、测试是否更完整。

如果这三点改善了，它就值得进入你的团队模板。

## 参考资料

- GitHub 仓库：https://github.com/multica-ai/andrej-karpathy-skills
- README：https://github.com/multica-ai/andrej-karpathy-skills/blob/main/README.md
- 中文 README：https://github.com/multica-ai/andrej-karpathy-skills/blob/main/README.zh.md
- CLAUDE.md：https://github.com/multica-ai/andrej-karpathy-skills/blob/main/CLAUDE.md
- Cursor 使用说明：https://github.com/multica-ai/andrej-karpathy-skills/blob/main/CURSOR.md
- 示例说明：https://github.com/multica-ai/andrej-karpathy-skills/blob/main/EXAMPLES.md
- 4SAPI API 地址：https://4sapi.com/v1
