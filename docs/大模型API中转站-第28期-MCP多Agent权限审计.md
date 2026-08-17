---
title: "【大模型API中转站】第28期 MCP多Agent闭环 | 权限审计方案"
category: 人工智能
tags:
  - 大模型API中转站
  - MCP
  - Multi-Agent
  - 权限审计
  - AI安全
  - 4SAPI
description: "围绕 MCP、工具调用、多 Agent handoff、guardrails 和 tracing，整理企业级 Agent 闭环的权限、日志、审计和模型网关落地方案。"
---

# 【大模型API中转站】第28期 MCP多Agent闭环 | 权限审计方案

本文是【大模型API中转站】系列的第28篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

当 Coding Agent 只会生成代码时，风险主要在“代码对不对”。

当 Agent 接入 MCP、数据库、浏览器、CI、Issue 系统、飞书、GitHub、内部接口后，风险就变成了“它能对外部系统做什么”。

这篇聊一个更企业化的问题：如果你要做多 Agent 闭环，怎么设计工具权限、审计日志、人类确认和模型网关？重点不是把 MCP 接起来，而是接起来以后不失控。

## 1. MCP 解决的是连接问题，不是治理问题

MCP 官方文档把它定义成连接 AI 应用和外部系统的开放标准。它能让 AI 应用访问数据源、工具和工作流。

但要注意：

```text
MCP 让连接更标准
不代表权限自动安全
不代表日志自动完整
不代表 Agent 行为自动合规
```

可以把 MCP 分成三类能力：

| 能力 | 作用 | 风险 |
| --- | --- | --- |
| Resources | 暴露只读上下文，比如文件、数据库 schema、日志 | 泄露敏感数据 |
| Tools | 执行动作，比如查数据库、发请求、建 Issue、改文件 | 误操作、越权、破坏数据 |
| Prompts | 暴露结构化工作流模板 | 模板过时、诱导错误流程 |

对企业来说，真正难的是 Tools。只读资源出错通常是泄露或误读，工具出错可能直接改系统。

## 2. 先给工具分级

不要把所有 MCP 工具一股脑给 Agent。建议按风险分四级：

| 等级 | 工具类型 | 例子 | 策略 |
| --- | --- | --- | --- |
| L0 | 只读公共信息 | 读文档、读 schema、查版本 | 默认允许 |
| L1 | 只读内部信息 | 查日志、读 Issue、看监控 | 记录日志，限制范围 |
| L2 | 可写低风险 | 建 Issue、写草稿、生成报告 | 允许但必须可回滚 |
| L3 | 可写高风险 | 改权限、发布、删数据、转账 | 必须人工确认 |

很多团队的问题不是模型不够聪明，而是把 L3 工具当 L0 给出去了。

例如：

```text
允许 Agent 查询订单状态：可以
允许 Agent 修改订单金额：高风险
允许 Agent 自动发布生产版本：高风险
允许 Agent 生成发布说明：可以
```

## 3. 多 Agent 不等于多几个聊天窗口

OpenAI Agents SDK 文档里，agent 可以配置 instructions、tools、handoffs、guardrails 和 structured outputs。handoffs 允许一个 Agent 把任务交给另一个专长 Agent。

企业多 Agent 最好按职责拆，而不是按模型拆。

示例：

```text
Planner Agent：只做任务拆解，不写代码
Coder Agent：只改指定目录，不碰权限和密钥
Reviewer Agent：只审查 diff，不直接修改
Security Agent：只查安全风险和敏感文件
Release Agent：只生成发布说明，不自动发布
```

每个 Agent 应该有不同工具权限。

| Agent | 可用工具 | 禁止工具 |
| --- | --- | --- |
| Planner | 读 spec、读代码结构 | 写文件、发布 |
| Coder | 读写业务代码、跑测试 | 修改密钥、发布 |
| Reviewer | 读 diff、读测试结果 | 写代码、合并 |
| Security | 读依赖、读配置、查漏洞 | 改权限、删文件 |
| Release | 读 PR、生成 changelog | 生产发布 |

这样即使某个 Agent 判断错误，也不会拿到超出职责的工具。

## 4. Guardrails 要放在工具前后

Guardrails 不应该只检查最终回答。更关键的是检查工具调用。

工具调用前检查：

```text
这个工具是否允许当前 Agent 使用？
目标资源是否在允许范围内？
参数里是否包含密钥、Token、个人信息？
是否需要人工确认？
```

工具调用后检查：

```text
返回结果是否包含敏感字段？
是否需要脱敏后再给模型？
是否触发异常状态？
是否要写审计日志？
```

举例：

```text
Agent 想调用 deploy_prod()
  -> guardrail 判断这是 L3 工具
  -> 要求人工确认
  -> 没有确认则拒绝执行
```

再举例：

```text
Agent 调用 read_logs()
  -> 返回日志包含 Authorization header
  -> 输出 guardrail 脱敏
  -> 再交给模型分析
```

这类规则不要靠模型自觉，应该写在代码和平台策略里。

## 5. MCP 资源优先只读

很多场景其实不需要工具写权限。比如让 Agent 分析系统问题，先给它只读资源就够了：

```text
Resources:
- repo://src/orders
- logs://order-service/recent
- schema://order-db/public
- docs://runbook/export-timeout
```

Tools 可以晚一点再开放：

```text
Tools:
- create_issue
- run_test
- open_pr
```

先读后写，是最简单也最有效的安全策略。

如果一个任务靠只读资源就能完成分析，不要急着给它写权限。

## 6. 人工确认点怎么设计

人工确认不要只放在最后。关键动作前也要确认。

建议这些动作必须确认：

- 修改权限策略。
- 修改数据库迁移。
- 删除文件或数据。
- 新增依赖。
- 访问生产日志。
- 打开外部网络请求。
- 创建生产发布。
- 合并 PR。

确认信息要具体，不要只问“是否继续”。

好的确认提示：

```text
Agent 请求执行 deploy_prod。

目标环境：production
分支：feature/order-export-fix
最近验证：pnpm test 通过，pnpm build 通过
风险：修改订单导出逻辑，未修改数据库

是否允许执行？
```

坏的确认提示：

```text
Agent 想执行一个命令，是否允许？
```

人类需要足够上下文，才能做判断。

## 7. 审计日志要覆盖模型和工具

很多团队只记录模型调用日志，但 Agent 真正危险的部分往往在工具调用。

建议同时记录两类日志：

模型日志：

| 字段 | 说明 |
| --- | --- |
| task_id | 任务 |
| agent_name | 哪个 Agent |
| model | 哪个模型 |
| stage | plan/code/review/security |
| input_tokens | 输入 |
| output_tokens | 输出 |
| status | 成功或失败 |

工具日志：

| 字段 | 说明 |
| --- | --- |
| task_id | 任务 |
| agent_name | 哪个 Agent |
| tool_name | 哪个工具 |
| risk_level | L0-L3 |
| args_hash | 参数摘要 |
| approved_by | 谁批准 |
| result_status | 成功或失败 |
| redacted | 是否脱敏 |

有了这两类日志，才能回答：

- 哪个 Agent 调用了高风险工具？
- 谁批准了？
- 调用前后模型看到了什么？
- 成本花在哪个阶段？
- 事故发生在模型判断还是工具执行？

## 8. 模型网关放在哪里

在多 Agent 系统里，大模型API中转站适合放在模型调用入口：

```text
Agent Orchestrator
      ↓
Model Gateway：4SAPI
      ↓
Claude / GPT / GLM / Kimi / DeepSeek / Qwen
```

它主要解决：

- 多模型统一调用。
- Key 和项目隔离。
- 调用额度控制。
- 成本归因。
- 模型路由。
- 请求日志。

MCP 解决 Agent 和工具的连接；4SAPI 这类中转站解决 Agent 和模型的连接。不要把两者混成一个东西。

推荐结构：

```text
MCP：管工具和资源
Guardrails：管权限和安全
4SAPI：管模型和成本
Tracing：管过程和审计
Human Review：管最终责任
```

## 9. 一个企业落地模板

可以从下面这套最小闭环开始：

```text
1. Issue 触发任务
2. Planner Agent 读取 Issue 和代码结构
3. Coder Agent 在隔离分支修改代码
4. Test Tool 运行最小测试
5. Reviewer Agent 审查 diff
6. Security Agent 检查敏感文件和依赖
7. 人工确认
8. 创建 PR
9. 记录模型日志和工具日志
```

权限设计：

```text
Planner：只读
Coder：读写业务代码，禁止密钥和发布
Reviewer：只读 diff
Security：只读配置和依赖
PR Bot：只能开 PR，不能合并
```

模型设计：

```text
Planner：强推理模型
Coder：代码模型
Reviewer：强审查模型
Security：安全审查模型或规则工具
Summary：低成本模型
```

日志设计：

```text
所有模型调用进入 4SAPI 项目 Key
所有工具调用进入审计表
所有 L3 工具需要人工确认
所有 PR 必须保留 Agent 摘要
```

## 10. 常见误区

误区一：MCP 接得越多越强。

工具越多，攻击面越大。先接最必要的只读资源，再开放可写工具。

误区二：多 Agent 会自动互相监督。

如果 reviewer 和 coder 拿到同样上下文、同样目标、同样工具，它们很可能犯同一类错误。互审要拆职责和工具权限。

误区三：模型 review 可以替代人工 review。

模型能补充检查，但不能承担业务责任。尤其是权限、资金、生产发布、用户数据，必须有人负责。

误区四：只看最终结果，不看过程。

Agent 事故很多时候不是最终回答错，而是中间工具调用越权。没有 tracing 和工具日志，事后很难复盘。

## 11. 总结

MCP、多 Agent、Loop Engineering 放在一起，确实能把 AI 编程从“辅助写代码”推进到“自动处理任务”。但企业落地时，核心不是让 Agent 更自由，而是让它的自由有边界。

建议记住五句话：

- MCP 管连接，不自动管安全。
- 工具要分级，高风险动作必须人工确认。
- 多 Agent 要按职责拆，不按热闹程度拆。
- 模型调用走大模型API中转站，成本和日志才可控。
- 审计要覆盖模型调用和工具调用，不能只看最终回答。

如果你已经用 4SAPI 做统一模型入口，可以把它放在多 Agent 系统的模型网关层；MCP 负责工具和资源，guardrails 负责权限和脱敏，tracing 负责过程记录，人工 review 负责最终判断。这样 Agent 才不是黑箱自动化，而是一套能被企业管理的工程系统。

参考资料：

- Model Context Protocol 官方介绍：https://modelcontextprotocol.io/docs/getting-started/intro
- MCP Resources 规范：https://modelcontextprotocol.io/specification/2025-06-18/server/resources
- MCP Tools 规范：https://modelcontextprotocol.io/specification/2025-11-25/server/tools
- MCP Server Concepts：https://modelcontextprotocol.io/docs/learn/server-concepts
- Anthropic：Introducing the Model Context Protocol：https://www.anthropic.com/news/model-context-protocol
- Anthropic：Writing effective tools for agents：https://www.anthropic.com/engineering/writing-tools-for-agents
- OpenAI Agents SDK：https://developers.openai.com/api/docs/guides/agents
- OpenAI Agents SDK Handoffs：https://openai.github.io/openai-agents-python/handoffs/
- OpenAI Agents SDK Guardrails：https://openai.github.io/openai-agents-python/guardrails/
- OpenAI Agents SDK Tracing：https://openai.github.io/openai-agents-python/tracing/
- 4SAPI 接入地址：https://4sapi.com/v1
