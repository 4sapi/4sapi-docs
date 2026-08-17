---
title: "【大模型API中转站】第117期 Hermes研究部 | Obsidian接入4SAPI"
category: 人工智能
tags:
  - 大模型API中转站
  - Hermes Agent
  - NotebookLM
  - Obsidian
  - 企业知识库
  - 4SAPI
description: "基于 HermesBible 的 NotebookLM + Obsidian 三 Agent 研究部工作流，讲企业如何用 Scout、Analyst、Briefer 搭研究情报系统，并用 4SAPI 做模型路由、成本治理和日志审计。"
---

# 【大模型API中转站】第117期 Hermes研究部 | Obsidian接入4SAPI

本文是【大模型API中转站】系列的第117篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

HermesBible 里另一个很值得写的工作流，是：

```text
Hermes + NotebookLM + Obsidian
```

它把一个人类研究流程拆成三个 Agent：

```text
Scout：找信号。
Analyst：分析和综合。
Briefer：生成早报和行动建议。
```

中间用 Obsidian 做长期资料库，用 NotebookLM 做资料理解和综合。

这套思路很适合企业。

因为很多团队真正缺的不是“再多一个聊天框”。

而是一个能持续运转的研究部门：

```text
每天看行业变化。
筛掉噪音。
沉淀进知识库。
生成给老板、产品、市场、研发看的 brief。
```

这篇就讲企业怎么把这个工作流接入 4SAPI，变成可控的研究情报系统。

## 1. 为什么企业需要 Agent 研究部

企业里的信息流太散了。

来源包括：

```text
官网更新
开发者文档
竞品博客
GitHub release
Reddit / X / Hacker News
行业报告
客户反馈
销售记录
内部会议纪要
客服工单
```

如果靠人手动看，很快就会变成：

```text
有人看到就转群。
没人看到就错过。
同一条信息重复解读。
结论散在聊天记录里。
下周没人找得到。
```

Agent 研究部要解决的不是“写一篇摘要”。

而是：

```text
持续发现信号。
判断信号价值。
沉淀证据。
形成团队可复用知识。
每天给出行动 brief。
```

这就是 Hermes + Obsidian + NotebookLM 的价值。

## 2. 三 Agent 分工

企业版可以这样拆。

### Scout Agent

负责发现信号。

输入：

```text
RSS
官网 changelog
GitHub release
X / Reddit
竞品网站
公开文档
内部客户反馈
```

输出：

```text
候选信号列表
来源链接
时间戳
一句话摘要
可信度初判
是否需要进入分析
```

Scout 不负责深度判断。

它只负责把可能重要的东西捞出来。

### Analyst Agent

负责分析。

输入：

```text
Scout 候选
已有 Obsidian 笔记
NotebookLM 资料集合
官方来源
独立来源
```

输出：

```text
事实分级
影响判断
和历史资料的关系
对产品/市场/研发的含义
不确定项
```

Analyst 需要更强的推理模型。

### Briefer Agent

负责给人看。

输入：

```text
Analyst 结论
业务优先级
团队角色
历史 brief
```

输出：

```text
晨报
周报
行动建议
风险提醒
需要人工决策的问题
```

Briefer 的重点不是“写得长”。

而是让人快速判断：

```text
今天要不要做动作？
谁该看？
下一步是什么？
```

## 3. 企业架构

可以这样搭：

```text
外部信息源
  -> Scout Agent
  -> raw notes / inbox
  -> Analyst Agent
  -> Obsidian vault / NotebookLM
  -> Briefer Agent
  -> 企业微信 / 飞书 / 邮件 / Notion
  -> 人工反馈
```

模型调用统一走 4SAPI：

```text
Scout -> 4SAPI low-cost key
Analyst -> 4SAPI reasoning key
Briefer -> 4SAPI summary key
QA -> 4SAPI reviewer key
```

这样每个阶段成本清楚。

不会出现“每天晨报到底花了多少钱没人知道”的问题。

## 4. 4SAPI Key 怎么拆

推荐至少拆四类 Key。

```text
research-scout-key
research-analyst-key
research-briefer-key
research-review-key
```

如果企业部门多，还可以按部门拆：

```text
product-research-scout-key
marketing-research-scout-key
sales-research-scout-key
engineering-research-scout-key
```

这样你能统计：

```text
产品情报花了多少钱？
市场情报花了多少钱？
哪个部门的 daily brief 最贵？
哪些来源带来的有效信息最多？
```

这就是企业级 API 成本治理。

## 5. 4SAPI 接入研究部：把每份 brief 算清楚

研究部工作流最容易被低估成本。

因为它看起来只是“每天总结一下”。

但真正跑起来后，它会做很多模型调用：

```text
Scout 扫来源。
Scout 摘要候选。
Analyst 交叉验证。
Analyst 对比历史笔记。
Briefer 生成早报。
QA 检查事实和风险。
```

如果这些调用都混在一个 Key 里，你只能看到总账单。

看不到：

```text
哪类来源最贵？
哪类 brief 最贵？
哪个部门消耗最高？
哪些候选从来没被采纳？
```

所以建议研究部所有模型调用都走 4SAPI：

```text
Base URL: https://4sapi.com/v1
API Key: 按 Agent 角色和部门拆分
Model: 按阶段从 4SAPI 模型广场选择
```

一个更企业化的 Key 方案：

```text
4sapi-research-scout-product
4sapi-research-scout-market
4sapi-research-scout-sales
4sapi-research-analyst-product
4sapi-research-analyst-market
4sapi-research-briefer-exec
4sapi-research-qa
```

这样你可以按部门看账：

```text
产品情报每天花多少钱？
市场情报每天花多少钱？
高管 brief 每周花多少钱？
QA 拦截了多少未经验证信息？
```

这就是 4SAPI 对企业研究部的营销卖点：

```text
不是只让 Agent 会总结，而是让总结过程可计费、可审计、可优化。
```

## 6. 资料入库前必须做事实分级

研究工作流最怕一件事：

```text
把未经验证的传闻写进知识库。
```

建议每条信息都打标签。

```text
confirmed：官方确认或多源一致。
likely：多个来源支持，但官方未确认。
single-case：单一用户案例。
opinion：观点或推测。
rumor：传闻，不进入正式结论。
```

Obsidian 笔记里可以写：

```markdown
---
source_type: official
confidence: confirmed
checked_at: 2026-06-29
workflow: hermes-research
model_key: research-analyst-key
---

# 标题

## 事实

## 证据

## 推断

## 对业务影响

## 未确认问题
```

不要让 Agent 把观点写成事实。

这是内容和研究团队最容易翻车的地方。

## 7. NotebookLM 和 Obsidian 分工

可以这样理解：

```text
Obsidian：长期知识库和版本化笔记。
NotebookLM：围绕一组资料做理解和问答。
Hermes：调度 Agent、读取资料、生成 brief。
4SAPI：模型调用治理。
```

不要把所有东西都塞进 NotebookLM。

也不要把所有临时材料都塞进 Obsidian。

推荐流程：

```text
原始资料 -> inbox
Scout 初筛 -> raw notes
Analyst 验证 -> permanent notes
Briefer 输出 -> reports
人工反馈 -> update tags
```

这样知识库会越来越干净。

## 8. 每日 brief 模板

企业晨报不要写成新闻合集。

建议固定成：

```markdown
# 今日 AI / 行业情报 Brief

## 1. 必看变化
- 发生了什么
- 证据来源
- 可信度
- 对我们的影响

## 2. 建议动作
- 产品：
- 市场：
- 销售：
- 研发：

## 3. 需要人工判断
- 问题：
- 选项：
- 推荐：

## 4. 已归档资料
- Obsidian 路径：
- NotebookLM 资料集：

## 5. 成本和运行状态
- Scout 调用：
- Analyst 调用：
- Briefer 调用：
- 失败：
```

最后一段“成本和运行状态”很关键。

这让研究工作流可治理。

## 9. 企业研究部的 4SAPI 成本报表模板

建议每天或每周生成一份成本报表。

```markdown
# Hermes Research Cost Report

## 周期
2026-06-23 至 2026-06-29

## 总览
- 总 brief 数：35
- 被采纳 brief 数：18
- 总模型成本：128.40 元
- 平均每份 brief 成本：3.67 元
- QA 拦截高风险结论：7 条

## 按部门
| 部门 | brief 数 | 成本 | 采纳数 | 平均成本 |
|---|---:|---:|---:|---:|
| 产品 | 12 | 46.20 | 8 | 3.85 |
| 市场 | 10 | 31.70 | 5 | 3.17 |
| 销售 | 8 | 28.10 | 3 | 3.51 |
| 高管 | 5 | 22.40 | 2 | 4.48 |

## 按阶段
| 阶段 | 调用次数 | 成本 | 备注 |
|---|---:|---:|---|
| Scout | 420 | 24.60 | 可继续降成本 |
| Analyst | 96 | 71.30 | 强模型成本最高 |
| Briefer | 35 | 18.20 | 稳定 |
| QA | 35 | 14.30 | 拦截有效 |

## 优化建议
- Scout 阶段换低成本模型。
- Analyst 只对高优先级候选使用强模型。
- 市场来源噪音高，需要缩小来源列表。
```

这份报表的意义不只是财务。

它会倒逼研究部优化来源质量。

很多团队的研究流程失败，不是模型不行。

而是来源太脏、筛选太宽、没有复盘。

4SAPI 的日志和成本统计，正好能把这些问题暴露出来。

## 10. 怎么把研究成果变成企业知识库资产

企业不要只把 brief 发到群里。

群消息很快就消失。

建议每条有效结论都沉淀成知识库条目：

```text
原始来源
事实分级
模型分析
人工确认
业务影响
相关项目
后续动作
4SAPI 调用成本
```

比如：

```markdown
# OpenAI 新功能对企业 Agent 工作流的影响

## 状态
confirmed

## 来源
- 官方文档：
- 独立讨论：

## 影响
- 研发：
- 内容：
- 客服：

## 建议动作
- 更新内部接入文档。
- 检查现有模型路由。

## 成本记录
- workflow: hermes-research-brief
- 4SAPI key: research-analyst-product
- cost: 2.14 元
```

这样研究结果才会变成资产。

否则每天 brief 再多，也只是信息烟花。

## 11. 什么时候不该自动化

这些场景不适合让 Agent 直接下结论：

```text
投资判断
法律合规解释
医疗健康建议
重大产品战略
客户承诺
竞争情报的敏感来源
员工绩效结论
```

Agent 可以做资料整理。

但最终判断要人来做。

研究 Agent 的目标是：

```text
减少信息噪音。
提高资料可追溯性。
让人更快做判断。
```

不是替代管理层拍板。

## 12. 上线检查清单

```text
[ ] 信息源是否分级？
[ ] Scout 是否只做初筛，不直接写最终结论？
[ ] Analyst 是否记录证据和不确定项？
[ ] Briefer 是否按角色输出行动建议？
[ ] Obsidian 笔记是否有来源和时间戳？
[ ] 传闻是否禁止进入正式结论？
[ ] 4SAPI Key 是否按 Agent 角色拆分？
[ ] 是否能统计每日 brief 成本？
[ ] 是否有人审核高风险结论？
[ ] 是否定期清理过期资料？
```

研究系统最怕“越积越乱”。

所以清理机制和事实分级同样重要。

## 13. 最后总结

Hermes + NotebookLM + Obsidian 的三 Agent 研究部，很适合企业内容、产品、市场和战略团队。

但企业落地时，重点不是“每天自动总结一堆信息”。

重点是：

```text
信号从哪里来。
证据是否可靠。
结论怎么分级。
知识怎么归档。
brief 谁来读。
模型成本谁来管。
```

Hermes 负责调度 Scout、Analyst、Briefer。

Obsidian 和 NotebookLM 负责知识沉淀和资料理解。

4SAPI 负责企业级大模型接入、模型路由、日志审计和成本治理。

一句话：

```text
别做自动摘要机器，要做可追溯的企业研究部。
```

## 资料来源与延伸阅读

- HermesBible：3-Agent Research Department：https://www.hermesbible.com/flows/3-agent-research-department-notebooklm-obsidian
- Hermes Agent 官方文档：https://hermes-agent.nousresearch.com/docs
- Hermes Memory 官方文档：https://hermes-agent.nousresearch.com/docs/user-guide/features/memory
- Hermes MCP 官方文档：https://hermes-agent.nousresearch.com/docs/user-guide/features/mcp
- Hermes Skills 官方文档：https://hermes-agent.nousresearch.com/docs/user-guide/features/skills
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 文档：https://4sapi.apifox.cn/
