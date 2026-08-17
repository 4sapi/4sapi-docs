---
title: "【大模型API中转站】第103期 KOL Finder Loop | 找博主不混脏数据"
category: 人工智能
tags:
  - 大模型API中转站
  - KOL Finder
  - Loop Engineering
  - AI Agent
  - 企业级大模型接入
  - 企业级API
  - 4SAPI
description: "用一个英文金融 KOL 搜索案例讲清 Loop Engineering 的完整落地：Define、Search、Filter、Content Audit、QA、Feedback、Stop、Deliver，并说明如何用 State 表、Connector、多 Agent 和 4SAPI 企业API网关做企业级大模型接入、权限审计和成本治理。"
---

# 【大模型API中转站】第103期 KOL Finder Loop | 找博主不混脏数据

本文是【大模型API中转站】系列的第103篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一篇讲了 Loop Engineering 的最小模板。

这一篇直接看一个完整案例：

```text
KOL Finder Loop。
```

场景很常见。

你想找一批英文金融 KOL。

最松的指令可能是：

```text
帮我找 30 个金融博主。
```

这个指令一定会脏。

AI 很容易混进：

```text
媒体号
公司号
品牌号
CEO 账号
政治号
生活号
汽车号
纯 Crypto 号
国家不明账号
重复账号
粉丝不达标账号
```

最后你还得人肉检查。

真正可用的 KOL Finder，不是一次生成 30 个名字。

而是一套会自己筛错、去重、补字段、记录状态的 Loop。

## 1. 先把硬条件锁死

不要让 Agent 自由发挥。

先定义任务边界。

可以这样写：

```text
请使用「Loop Engineering + KOL Finder」方法，帮我寻找英文金融 KOL。

要求：
1. 平台：YouTube、TikTok、Instagram
2. 国家：英国、加拿大、澳洲、新加坡、新西兰
3. 语言：英文
4. 粉丝门槛：100K+
5. 数量：找到 30 个后跳出 Loop
6. 内容：金融交易、股票投资、外汇、黄金、金融时事、个人理财、投资教育
7. 排除：媒体号、公司号、品牌号、CEO/创始人号、政治号、生活号、汽车号、纯 Crypto/Web3 号
8. 可以有少量 Web3，但不能以 Web3/Crypto 为主
9. 必须跟之前名单去重
10. 输出 Excel，粉丝用 K/M，链接用 handle 格式，国家用中文，联系方式能找到就列出来
```

这里最重要的是四件事：

```text
硬条件。
排除条件。
停止条件。
输出格式。
```

这四件事不写清楚，Agent 后面每一轮都会跑偏。

## 2. KOL Loop 的阶段拆解

一个 KOL Finder Loop 可以拆成八个阶段。

```text
Define
Search
Filter
Content Audit
QA
Feedback
Stop
Deliver
```

### 2.1 Define：锁范围

Define 阶段只做一件事：

```text
防止条件太宽。
```

要锁定：

```text
国家
语言
平台
粉丝门槛
内容范围
排除类型
数量目标
输出格式
```

如果你找的是金融 KOL，就不要把财经媒体号、金融公司号、纯 Web3 号混进来。

如果你要个人创作者，就要排除品牌号和公司号。

### 2.2 Search：先扩大候选池

Search 阶段不要太严。

先扩大范围。

可能的来源：

```text
YouTube 搜索
TikTok hashtag
Instagram profile
Google 搜索
Apify
KOL Finder
平台榜单
竞品合作名单
```

这一轮的目标不是直接交付。

而是建立候选池。

### 2.3 Filter：挡掉明显脏数据

Filter 阶段检查：

```text
粉丝是否达标。
语言是否英文。
国家是否在目标范围。
是否重复账号。
是否个人创作者。
是否明显公司号、媒体号、品牌号。
```

这一层是硬门槛。

不合格直接剔除。

### 2.4 Content Audit：看近期内容

很多账号简介写得像金融。

但近期内容可能已经变了。

所以要看近期作品：

```text
最近 10 条视频 / 帖子在讲什么？
是否持续讲金融、交易、投资、理财？
Web3 是否只是偶尔出现？
是否已经转向生活、汽车、政治、娱乐？
```

Content Audit 是 KOL Loop 里最容易省略、也最容易出错的一步。

### 2.5 QA：查污染项

QA 阶段专门找错误。

比如：

```text
汽车号混入
新闻号混入
政治号混入
生活号混入
纯 Crypto 号混入
国家靠猜
粉丝格式错误
链接不是 handle
联系方式字段乱写
```

QA 不只是检查格式。

它要检查任务边界。

### 2.6 Feedback：把错误变成下一轮规则

如果本轮混进汽车号，下一轮要强化排除词。

如果纯 Web3 太多，下一轮提高金融内容占比要求。

如果国家不明太多，下一轮要求来源证据。

这就是 Feedback。

不是简单说：

```text
重新找。
```

而是说：

```text
基于本轮错误，下一轮只修正最主要的问题。
```

### 2.7 Stop：到条件就停

停止条件必须写清楚：

```text
达到 30 个合格账号后停止。
连续 2 轮无新增合格账号后停止，并报告库存上限。
```

这很重要。

AI 很容易为了凑数硬塞。

如果条件太严找不够，就应该报告：

```text
当前条件下库存不足。
建议放宽国家、平台或粉丝门槛。
```

不要硬凑。

### 2.8 Deliver：输出可用表格

最后交付不是一段文字。

而是可用表格：

```text
国家
平台
handle
粉丝数
主页链接
联系方式
内容审核说明
验证状态
剔除风险
```

如果要发给客户，建议导出 Excel。

## 3. State 表是 KOL Loop 的记忆

没有状态表，Agent 每一轮都像重新开始。

建议维护一张表：

```text
已找到合格数量：18/30
本轮污染项：公司号、纯 Crypto、国家不明
新增规则：近期内容必须以金融/投资为主
待复核项：国家不明账号 5 个
下一轮重点：补加拿大和新西兰
停止条件：达到 30 个，或连续 2 轮无新增合格账号
```

这张表有两个作用。

第一，让 Agent 带着上一轮经验继续。

第二，让人类快速判断 Loop 是否在变好。

如果三轮之后还是同样污染项，说明规则没生效。

如果连续两轮没有新增合格账号，说明该停。

## 4. 六个增强零件

最小 KOL Loop 只需要 Define、Filter、QA、Feedback、Stop。

但如果任务变成长期业务，可以加入六个增强零件。

```text
Heartbeat
Isolation
Skill
Connector
Agent
State
```

### 4.1 Heartbeat：定期刷新

KOL 不是找完就结束。

粉丝会变。

内容方向会变。

联系方式会失效。

所以可以设置：

```text
每周刷新候选池。
每月复查老 KOL。
每次 Campaign 前重新检查粉丝、内容、联系方式。
```

### 4.2 Isolation：隔离数据池

不要把所有账号混在一起。

至少分成：

```text
候选池
待复核池
正式交付池
剔除池
```

这样脏数据不会污染最终名单。

### 4.3 Skill：沉淀方法

当你反复找 KOL，就应该把方法沉淀成 Skill。

以后只替换：

```text
国家
平台
垂类
粉丝门槛
排除条件
```

不要每次从零写提示词。

### 4.4 Connector：接外部工具

长期做 KOL 搜索，需要工具连接：

```text
YouTube API
TikTok
Instagram
Apify
Google 搜索
Excel
Google Sheets
Obsidian
CRM
```

Connector 解决的是数据进出。

### 4.5 Agent：分工处理

复杂任务可以拆成多 Agent：

```text
Search Agent：找候选
Filter Agent：查粉丝、国家、语言、账号类型
Content Audit Agent：看近期内容是否金融垂类
QA Agent：查重复、脏数据、联系方式、格式
Deliver Agent：整理 Excel
```

Anthropic Claude Code 文档里也讲到 subagents 适合把会污染主上下文的研究、搜索、文件读取任务拆出去，只把总结返回。

这和 KOL Finder 很贴。

### 4.6 State：记录进度

State 是最基础的增强。

没有 State，就没有真正的 Loop。

## 5. Hooks 怎么理解

Claude Code 的 hooks 可以理解成高级版检查节点。

官方文档里提到 hooks 可以在工具事件周围触发，agent hooks 甚至可以派出子代理去读文件、搜索代码、检查条件。

放到 KOL Loop 里，可以这样理解：

```text
写入正式名单前，触发去重检查。
交付 Excel 前，触发字段完整性检查。
每轮搜索结束后，触发污染项总结。
找不到足够数量时，触发库存上限报告。
```

普通人一开始不用真的配置 hooks。

先在 prompt 里写清楚这些检查点，已经能获得很多收益。

等任务高频、团队化、工具链稳定后，再把检查节点工程化。

## 6. 4SAPI 在 KOL Loop 里的企业级位置

KOL Finder 一旦变成团队服务，就不只是提示词问题。

你会遇到企业级大模型接入和企业级 API 问题：

```text
搜索结果谁来审核？
模型调用谁付钱？
不同客户能不能拆 Key？
每个 Campaign 成本多少？
高成本模型有没有被滥用？
哪一轮失败最多？
哪个 Agent 消耗最多？
```

这时可以把 4SAPI 放在模型调用层，把它当成企业API网关来用。

KOL Loop 继续负责业务筛选，4SAPI 负责 Key 管理、模型路由、权限审计、日志审计和成本治理。

推荐拆 Key：

```text
4SAPI-KOL-Search-Summary
用途：候选账号摘要和初筛。
```

```text
4SAPI-KOL-Audit
用途：内容审核和垂类判断。
```

```text
4SAPI-KOL-QA
用途：去重、格式检查、风险提示。
```

```text
4SAPI-KOL-ClientA
用途：客户 A 的独立 Campaign。
```

在工作流里写清楚：

```text
Base URL: https://4sapi.com/v1
Model: 从 4SAPI 模型广场复制
日志：以 4SAPI 后台调用记录为准
预算：单个 Campaign 不超过指定额度
权限：客户项目单独 Key
```

这样你不只是交付 KOL 名单。

你是在交付一套可审计、可计费、可复盘的企业级 KOL 搜索工作流。

## 7. KOL Loop 的最小检查清单

```text
[ ] 平台是否明确
[ ] 国家是否明确
[ ] 语言是否明确
[ ] 粉丝门槛是否明确
[ ] 内容范围是否明确
[ ] 排除类型是否明确
[ ] 是否允许少量 Web3 / Crypto
[ ] 是否要求去重
[ ] 输出格式是否明确
[ ] 是否有 State 表
[ ] 是否有污染项记录
[ ] 是否有连续失败停止条件
[ ] 是否保留待复核池
[ ] 是否标注验证状态
[ ] 企业场景是否用 4SAPI 拆 Key 和看日志
```

## 8. 最后总结

KOL Finder Loop 节省的，不是第一次搜索的时间。

它真正节省的是：

```text
反复检查。
纠错。
去重。
补字段。
剔除脏数据。
报告库存上限。
```

一个好的 KOL Loop，不会为了凑数硬塞账号。

它会告诉你：

```text
当前条件下，我找到了多少。
哪些被剔除。
为什么剔除。
下一轮该怎么补。
什么时候应该停止。
```

如果只是个人使用，写清楚模板和 State 表就够。

如果给客户或团队使用，就把模型调用接进 4SAPI：

```text
KOL Loop 负责筛选。
Agent 负责执行。
4SAPI 负责企业级 API、Key 权限、日志审计和成本治理。
```

这才是能长期交付的 KOL 工作流。

## 资料来源与延伸阅读

- OpenAI Codex：Iterate on difficult problems：https://developers.openai.com/codex/use-cases/iterate-on-difficult-problems
- Anthropic Claude Code Hooks：https://docs.anthropic.com/en/docs/claude-code/hooks
- Anthropic Claude Code Subagents：https://docs.anthropic.com/en/docs/claude-code/sub-agents
- Anthropic Claude Code Common Workflows：https://docs.anthropic.com/en/docs/claude-code/common-workflows
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 文档：https://4sapi.apifox.cn/
