---
title: "【大模型API中转站】第25期 Agent循环架构 | 让自动化少跑偏"
category: 人工智能
tags:
  - 大模型API中转站
  - Loop Engineering
  - Coding Agent
  - AI Agent
  - 架构设计
  - 4SAPI
description: "不再从概念解释 Loop Engineering，而是从团队落地角度拆解 Agent 循环的触发、状态、执行、验证、模型路由和成本审计。"
---

# 【大模型API中转站】第25期 Agent循环架构 | 让自动化少跑偏

本文是【大模型API中转站】系列的第25篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

最近大家都在聊 Loop Engineering，容易把它理解成“让 AI 自己一直干活”。这个理解太粗了。

真正值得团队关注的不是概念本身，而是一个更现实的问题：当 Coding Agent 不再只响应一次提示词，而是会连续读代码、写代码、跑命令、看日志、继续修复时，你的模型调用入口、成本记录、权限边界和验收标准还扛得住吗？

如果扛不住，Loop 不是效率工具，而是一个会持续消耗 Token、持续修改代码、持续制造不确定性的后台进程。

这篇不重复讲“Loop 是什么”，而是从工程架构角度讲清楚：一个能落地的 Agent 循环，应该怎样接入大模型API中转站，怎样分层，怎样验收，怎样避免跑偏。

## 1. Loop 的本质不是提示词，而是任务系统

传统 AI 编程像这样：

```text
人输入需求
  -> 模型回答
  -> 人判断下一步
```

Agent 循环变成了这样：

```text
任务目标
  -> 拆解计划
  -> 调用模型
  -> 使用工具
  -> 执行验证
  -> 写入状态
  -> 决定继续或停止
```

差别不在于“提示词更高级”，而在于多了三层工程结构：

| 层级 | 传统对话 | Agent 循环 |
| --- | --- | --- |
| 触发方式 | 人手动输入 | 任务、定时器、PR、Issue、CI |
| 判断方式 | 人看结果 | 测试、审查器、规则、人工关卡 |
| 状态管理 | 聊天上下文 | 文件、数据库、任务队列、日志 |
| 成本形态 | 一次调用 | 多轮调用、失败重试、并行子任务 |
| 风险形态 | 一次错误回答 | 连续错误行动 |

所以 Loop Engineering 更接近“任务编排 + 模型网关 + 验收系统”，不是把一段提示词写长一点。

## 2. 一套适合团队的 Agent 循环架构

可以把它拆成七个组件：

```text
任务入口
  -> Loop 控制器
  -> 状态存储
  -> 执行 Agent
  -> 工具层
  -> 验证器
  -> 模型网关
```

更完整一点：

```text
GitHub Issue / PR / 人工需求 / 定时任务
        ↓
Loop 控制器：拆任务、限轮数、判断是否继续
        ↓
状态存储：PROGRESS.md / DB / 任务表 / 运行日志
        ↓
执行 Agent：读代码、写代码、跑命令、生成说明
        ↓
工具层：Git、测试、构建、浏览器、MCP、CI
        ↓
验证器：单测、lint、类型检查、模型 review、人工 review
        ↓
大模型API中转站：模型路由、Key、额度、日志、审计
        ↓
Claude / GPT / GLM / Kimi / DeepSeek / Qwen 等模型
```

这套结构里，大模型API中转站不是“锦上添花”。只要循环开始自动调用模型，就必须有人回答这些问题：

- 哪个项目在调用？
- 哪个阶段在调用？
- 哪个模型在调用？
- 一轮失败后重试了几次？
- 哪些调用属于测试，哪些属于生产？
- 最后产出有没有被人工接受？

如果这些问题都混在工具客户端里，后面复盘成本和质量会很痛苦。

## 3. 不同阶段不要用同一个模型

Agent 循环通常不是一个动作，而是一组阶段。每个阶段对模型要求不同。

| 阶段 | 任务 | 模型选择思路 |
| --- | --- | --- |
| 需求澄清 | 把一句话需求整理成边界和验收标准 | 中等成本、长上下文、表达稳定 |
| 方案设计 | 拆模块、判断风险、列影响面 | 强推理模型，低频使用 |
| 代码实现 | 修改文件、补测试、跑命令 | Coding 能力强、响应稳定 |
| 错误修复 | 根据测试结果迭代 | 便宜且速度快，限制轮数 |
| 代码审查 | 找严重 Bug、越权改动、安全风险 | 强模型或专门 reviewer |
| 总结归档 | 生成 changelog、PR 描述、日志摘要 | 低成本模型即可 |

一个常见错误是：所有阶段都用最强模型。

这当然省心，但成本很容易翻倍。更合理的做法是把“强模型”用在方案设计和最终审查，把普通实现、摘要、格式化交给性价比模型。

## 4. 用 4SAPI 做模型层的最小接入

如果你的执行工具或脚本支持 OpenAI Compatible 接口，模型层可以收敛成三个参数：

```text
base_url = https://4sapi.com/v1
api_key  = sk-xxxxxxxxxxxxxxxx
model    = 按阶段选择
```

一个最小 Python 包装可以这样写：

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-xxxxxxxxxxxxxxxx",
    base_url="https://4sapi.com/v1",
)

MODELS = {
    "spec": "gpt-4.1-mini",
    "plan": "claude-sonnet-4-5-20250929",
    "code": "glm-4.6",
    "review": "claude-opus-4-5-20251101",
    "summary": "deepseek-chat",
}

def ask_model(stage: str, messages: list[dict], max_tokens: int = 1200):
    return client.chat.completions.create(
        model=MODELS[stage],
        messages=messages,
        max_tokens=max_tokens,
    )
```

注意，这段代码只是模型调用层。真正生产落地时，不建议让每个 Agent 直接拿万能 Key。更稳的做法是按项目和阶段拆 Key：

```text
project-a-spec-key
project-a-code-key
project-a-review-key
project-a-batch-key
```

这样某个自动化任务失控时，可以只停掉对应 Key，不影响其他业务。

## 5. Loop 控制器要先管住三件事

第一，最多跑几轮。

```text
max_rounds = 8
```

没有轮数上限的 Agent 循环，很容易在一个小问题上反复尝试，最后 Token 花了，问题没解决。

第二，每轮最多改多少。

```text
max_files_changed = 8
max_diff_lines = 600
```

如果一个任务原本只是修复接口参数，却改了 30 个文件，这不是“自主性强”，而是失控信号。

第三，哪些文件不能碰。

```text
禁止自动修改：
- .env
- 生产密钥
- 支付配置
- 数据迁移脚本
- 权限策略
- CI 发布脚本
```

这些限制不要只写在提示词里。能用脚本检查就用脚本检查，能用权限隔离就用权限隔离。

## 6. 验证器比提示词更重要

Anthropic 在 Agent 架构文章里提到 evaluator-optimizer 工作流：一个模型生成，另一个模型评估并反馈，形成迭代。这个思路用在代码里很自然，但团队落地时不能只依赖模型评估。

建议把验证器分成三类：

| 验证器 | 例子 | 优先级 |
| --- | --- | --- |
| 代码型验证 | test、lint、typecheck、build | 最高 |
| 规则型验证 | diff 行数、敏感文件、依赖新增、许可证 | 高 |
| 模型型验证 | review、风险摘要、PR 描述 | 中 |

能用确定性工具验证的，就不要交给模型“感觉一下”。模型 review 适合找潜在问题，但不适合替代测试。

一个更稳的闭环是：

```text
Agent 修改代码
  -> 跑单测
  -> 跑类型检查
  -> 检查 diff 范围
  -> 模型 review
  -> 人工 review
```

这里面任何一步失败，都应该写入状态，而不是让 Agent 假装没看到。

## 7. 状态文件要写给下一轮看

Loop 最容易翻车的地方，是每一轮都像失忆一样重新开始。

建议在仓库里维护一个任务状态文件，例如：

```markdown
# LOOP_STATE

## Goal
修复订单导出功能在大文件下超时的问题。

## Constraints
- 不修改支付流程
- 不改数据库结构
- 最多 8 轮
- 每轮必须运行订单模块测试

## Done
- 已复现超时问题
- 已定位到 exportOrders 一次性加载全部数据

## Current
- 改成分页读取
- 补充 10 万行导出测试

## Blockers
- 本地缺少大文件 fixture，需要生成测试数据

## Last validation
- npm test -- order-export 失败
- 失败原因：分页边界少导出最后一页
```

状态文件不追求好看，追求可继续。下一轮 Agent 能读懂“已经做了什么、为什么失败、下一步该做什么”就够了。

## 8. 不要把所有自动化都做成 Loop

适合 Loop 的任务通常有三个特征：

- 目标可验证。
- 错误可恢复。
- 每轮反馈明确。

比如：

| 适合 | 不适合 |
| --- | --- |
| 修复测试直到通过 | 重新设计品牌视觉 |
| 批量迁移 API 调用 | 决定产品战略 |
| 定时检查依赖漏洞 | 无限制重构全仓库 |
| 自动生成 PR 摘要 | 自动合并生产发布 |
| 扫描日志并开 Issue | 自动处理财务付款 |

如果一个任务的成败只能靠人类判断，那就不要让它无人值守循环。可以让 Agent 做准备工作，但最终决策要留给人。

## 9. 推荐的落地顺序

不要一上来就做复杂多 Agent 系统。建议按下面顺序推进：

```text
第一周：单任务目标 + 手动触发 + 人工 review
第二周：加入测试验证 + 状态文件 + 轮数上限
第三周：接入 4SAPI 统一模型入口和日志
第四周：拆分 plan/code/review 模型策略
第五周：把高频任务接入定时器或 PR 触发
第六周：再考虑 MCP、子 Agent 和自动开 PR
```

真正成熟的 Agent 循环不是“无人看管”，而是“让系统把该自动化的部分自动化，把该人判断的地方暴露出来”。

## 10. 总结

Loop Engineering 对开发者最有价值的地方，不是多一个热词，而是提醒我们：AI 编程已经从一次性对话，变成了可持续运行的工程系统。

一旦进入循环，就必须把模型调用当成团队资源管理。任务入口、状态记录、工具权限、验证器、模型路由、Token 成本和人工关卡都要设计清楚。

如果你已经有 4SAPI 这类大模型API中转站，可以先把它放在模型网关层：统一 `base_url`、Key、模型路由、日志和额度。上层的 Claude Code、Codex、Cursor、ZCode 或自研 Agent 负责执行任务；下层网关负责让模型调用可统计、可控制、可复盘。

一句话：Loop 让 Agent 跑起来，中转站让它别跑丢。

参考资料：

- Addy Osmani：Loop Engineering：https://addyosmani.com/blog/loop-engineering/
- Anthropic：Building Effective Agents：https://www.anthropic.com/research/building-effective-agents
- Anthropic：Demystifying evals for AI agents：https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents
- Oracle：What Is the AI Agent Loop：https://blogs.oracle.com/developers/what-is-the-ai-agent-loop-the-core-architecture-behind-autonomous-ai-systems
- OpenAI Agents SDK：https://developers.openai.com/api/docs/guides/agents
- 4SAPI 接入地址：https://4sapi.com/v1
