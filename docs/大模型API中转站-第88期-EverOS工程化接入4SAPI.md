---
title: "【大模型API中转站】第88期 EverOS接入4SAPI | 记忆能Git也能审计"
category: 人工智能
tags:
  - 大模型API中转站
  - EverOS
  - 4SAPI
  - Agent工作流
  - 成本治理
  - 团队协作
description: "承接 EverOS 永久记忆入门，进一步讲团队如何把 EverOS 放进 Claude Code、Codex、Hermes 等 Agent 工作流：Markdown 记忆用 Git 管，模型调用走 4SAPI，形成可审计、可计费、可治理的长期 Agent 系统。"
---

# 【大模型API中转站】第88期 EverOS接入4SAPI | 记忆能Git也能审计

本文是【大模型API中转站】系列的第88篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一篇我们把 EverOS 跑起来了。

核心闭环是：

```text
写入对话
手动 flush
搜索召回
打开 ~/.everos 看 Markdown
```

做到这一步，个人 Agent 已经不再是每次从零开始。

但如果你要放进团队，就会多出另一组问题：

```text
谁能写记忆？
谁能改记忆？
记忆错了怎么审计？
模型调用花了多少钱？
不同项目能不能隔离？
不同 Agent 能不能共享经验？
Key 放在哪里？
日志在哪里看？
```

这篇就不再重复安装。

我们讲工程化接入。

一句话：

```text
EverOS 负责长期记忆。
Git 负责记忆审计。
4SAPI 负责模型调用治理。
```

如果上一篇是“让 Agent 有记忆”，这一篇就是：

```text
让这份记忆能进团队流程。
```

## 1. 个人可用，不等于团队可用

个人使用 EverOS，很简单。

本地跑服务。

写进 `~/.everos`。

想看就打开。

想改就编辑。

Agent 下次搜索回来继续用。

但团队环境里，问题会复杂很多。

比如一个内容团队有 5 个 Agent：

```text
选题 Agent
资料 Agent
写稿 Agent
审核 Agent
分发 Agent
```

一个开发团队有 6 个 Agent：

```text
需求分析 Agent
编码 Agent
测试 Agent
Review Agent
发布 Agent
故障排查 Agent
```

每个 Agent 都可能写记忆。

每个人也可能有自己的偏好。

每个项目还有不同的约束。

如果没有工程边界，记忆很快会变成一锅粥。

所以 EverOS 的价值不只是“本地能存”。

它真正适合团队的原因是：

```text
它把记忆暴露成文件结构。
```

文件结构就意味着：

```text
可以分目录。
可以 Git。
可以 code review。
可以脱敏。
可以备份。
可以迁移。
可以人工修正。
```

这比纯数据库记忆更适合工程团队。

## 2. EverOS 在团队架构里的位置

先画一张逻辑图。

```text
Claude Code / Codex / Hermes / 自研 Agent
  ↓
记忆读写适配层
  ↓
EverOS HTTP API
  ↓
Markdown 记忆可信源
  ↓
SQLite 状态 + LanceDB 检索索引
  ↓
LLM / Embedding / Rerank 调用
  ↓
4SAPI 统一模型入口
  ↓
Claude / GPT / Gemini / Qwen / GLM 等模型
```

这里有三层责任。

第一层，Agent 工作流。

Agent 决定什么时候读记忆、什么时候写记忆、写什么内容。

第二层，EverOS 记忆运行时。

EverOS 负责抽取、落盘、索引、召回。

第三层，4SAPI 模型治理。

4SAPI 负责统一模型调用、Key、费用、错误日志和团队权限。

这三层不要混在一起。

很多团队一开始会直接让 Agent 调模型厂商 API。

短期可以。

长期很痛。

因为你很快会遇到：

```text
模型 Key 散在不同人的 .env 里。
不同 Agent 用不同模型，没有统一路由。
记忆抽取花了多少钱没人知道。
Embedding 和 Rerank 调用失败没人看日志。
团队成员离职后 Key 不知道在哪。
```

把 4SAPI 放在模型入口层，就是为了提前解决这些问题。

## 3. 记忆目录要按 app 和 project 切

EverOS 的记忆路径天然支持 `app_id` 和 `project_id`。

默认会落到：

```text
~/.everos/default_app/default_project/
```

如果你在团队里使用，别一直用 default。

建议明确设置：

```text
app_id = agent-workbench
project_id = ecommerce-api
```

或者：

```text
app_id = content-factory
project_id = blog-series-4sapi
```

这样不同项目的记忆不会混。

目录大致会变成：

```text
~/.everos/
├── agent-workbench/
│   └── ecommerce-api/
│       ├── users/
│       └── agents/
└── content-factory/
    └── blog-series-4sapi/
        ├── users/
        └── agents/
```

这个隔离非常重要。

不要让电商项目的故障排查经验，污染内容生产项目。

也不要让个人偏好，污染团队规范。

记忆系统最怕不是“没记住”。

而是“记混了”。

## 4. 用户记忆和 Agent 记忆分开治理

EverOS 里有两条轨道：

```text
users/
agents/
```

用户记忆应该记：

```text
用户偏好
沟通习惯
长期目标
项目上下文
明确授权的个人化信息
```

Agent 记忆应该记：

```text
任务轨迹
案例复用
踩坑记录
工具使用经验
流程模式
可沉淀的 Skill
```

团队里最好给这两类记忆定不同规则。

用户记忆更敏感。

要少写、慎写、可删除。

Agent 记忆更偏工程资产。

可以更积极沉淀，但要脱敏。

比如一个编码 Agent 完成任务后，可以写入：

```text
这个仓库测试命令是 pnpm test。
发布前要跑 pnpm build。
API schema 变更需要同步 docs/openapi.yaml。
Windows 下某个脚本路径需要用 -LiteralPath。
```

但不应该写入：

```text
真实客户姓名。
生产数据库密码。
内部未公开域名。
用户手机号。
完整合同原文。
```

长期记忆不是垃圾桶。

它是资产库。

资产库必须有入库标准。

## 5. Markdown 记忆为什么适合 Git

EverOS 最适合工程团队的一点，是 Markdown 作为可信源。

这让记忆天然可以进 Git。

但这里要分清：

```text
用户可见记忆：可以选择性版本化。
.index：派生索引，不应该提交。
.env：密钥配置，绝对不要提交。
```

建议规则：

```text
提交 users/ 和 agents/ 下经过脱敏的 Markdown。
不提交 .index/。
不提交 .tmp/。
不提交 .env。
```

你可以在团队仓库里建立一个专门目录：

```text
agent-memory/
├── project-rules/
├── agent-cases/
├── workflow-notes/
└── review-log/
```

然后定期从 EverOS 的 Markdown 中挑选有价值的内容同步进去。

也可以把某些稳定规则反向写回 EverOS。

关键是：

```text
长期有效的经验要进 Git。
临时上下文留在本地记忆。
敏感信息不进 Git。
```

这样你就能用普通工程流程管理 Agent 记忆：

```text
pull request
code review
diff
blame
revert
release note
```

这比“后台数据库里有一坨记忆”可靠得多。

## 6. 什么记忆值得写入

团队使用时，不要什么都写。

建议只写这几类。

第一，稳定偏好。

```text
团队默认用 pnpm。
TypeScript 禁止 any。
PR 描述必须包含测试结果。
中文博客保持短句风格。
```

第二，项目约束。

```text
支付模块禁止自动改数据库 schema。
导出 Excel 必须保留中文列名。
移动端兼容最低 Android 版本是 10。
```

第三，已验证流程。

```text
本仓库发版前先跑 pnpm lint，再跑 pnpm test，再跑 pnpm build。
线上故障排查先看日志面板，再查队列积压，再查数据库慢查询。
```

第四，踩坑记录。

```text
Windows PowerShell 路径带中文时要用 -LiteralPath。
某个模型在长 JSON 输出时容易丢字段，要加 schema 校验。
某个接口返回 429 时不是 Key 错，而是分组限流。
```

第五，成功案例。

```text
某次自动生成周报的输入结构。
某次 Codex 修复测试失败的执行链路。
某次 4SAPI 成本下降的模型路由策略。
```

不建议写入：

```text
短期闲聊
未经确认的猜测
真实敏感数据
过期临时任务
含 Key 的命令
含客户隐私的原文
```

这套标准越清楚，Agent 越不容易把记忆写脏。

## 7. Agent 开工前怎么读记忆

一个实用模式是：

```text
任务开始前先 search。
任务结束后再 add + flush。
```

开工前搜索可以分三类。

第一，项目记忆：

```text
这个项目有哪些技术约束？
这个仓库怎么运行测试？
之前类似任务有什么坑？
```

第二，用户偏好：

```text
这个用户喜欢什么输出风格？
他更关心成本还是效果？
他有没有固定格式要求？
```

第三，Agent 案例：

```text
之前有没有类似任务成功案例？
当时怎么拆步骤？
哪些命令跑过？
最后怎么验证？
```

把召回结果塞进 Agent 上下文时，不要原样全部塞。

建议做一层整理：

```text
只保留和当前任务有关的记忆。
保留来源路径。
标注置信度。
明显过期的记忆不使用。
```

这样 Agent 既能“记得”，又不会被旧记忆拖偏。

## 8. Agent 收工后怎么写记忆

任务结束后，建议写三类内容。

第一，结果摘要。

```text
本次做了什么。
改了哪些文件。
验证结果是什么。
```

第二，可复用经验。

```text
哪些步骤下次可以复用。
哪些命令有效。
哪个坑已经解决。
```

第三，禁止事项。

```text
哪些路径不要碰。
哪些命令风险高。
哪些数据不能发给模型。
```

可以让 Agent 在收工时生成一段结构化 messages，然后调用 EverOS：

```text
POST /api/v1/memory/add
POST /api/v1/memory/flush
```

下一次再用 `/search` 找回来。

这个闭环很朴素，但非常有效。

因为它把“聊天记录”变成了“可检索经验”。

## 9. 4SAPI 在这里怎么配

EverOS 的 `.env` 里会配置多个模型能力。

常见包括：

```text
LLM
Embedding
Rerank
Multimodal
```

如果团队使用 4SAPI，可以把兼容 OpenAI 协议的能力统一指向 4SAPI 网关。

示意：

```env
EVEROS_LLM__MODEL=<你在4SAPI里配置的模型名>
EVEROS_LLM__API_KEY=<4SAPI_API_KEY>
EVEROS_LLM__BASE_URL=https://4sapi.com/v1

EVEROS_EMBEDDING__MODEL=<Embedding模型名>
EVEROS_EMBEDDING__API_KEY=<4SAPI_API_KEY>
EVEROS_EMBEDDING__BASE_URL=https://4sapi.com/v1
```

实际字段和模型名以你自己的 4SAPI 控制台和文档为准。

这里的核心不是死记配置。

而是架构位置：

```text
EverOS 不直接散连多个模型厂商。
统一交给 4SAPI 出口。
```

这样做的收益很直接。

你可以在 4SAPI 里看：

```text
哪个项目调用最多。
哪个 Agent 最费钱。
哪类任务失败率高。
哪个模型适合抽取。
哪个模型适合总结。
哪个模型适合多模态解析。
```

这就是从“能跑”到“可运营”的区别。

## 10. 模型路由策略

记忆系统不应该所有步骤都用同一个模型。

建议按任务分层。

```text
记忆抽取：稳定、便宜、结构化能力好的模型。
长文总结：上下文更长、总结能力更强的模型。
多模态解析：图文能力强的模型。
Rerank：专门的重排模型。
Embedding：固定 embedding 模型，避免频繁切换。
最终回答：按业务场景选择 Claude、GPT、Gemini、Qwen、GLM。
```

4SAPI 的价值在这里会很明显。

你可以把模型路由藏在中转层。

上层 Agent 不需要知道每次到底走哪个供应商。

它只需要知道：

```text
我要一个记忆抽取模型。
我要一个报告生成模型。
我要一个低成本总结模型。
```

中转层负责把请求发到合适模型。

这对团队维护很重要。

因为模型会变。

价格会变。

效果会变。

供应商也会变。

如果所有配置散落在每个 Agent 里，后期维护会很难。

## 11. 成本治理：记忆系统会持续花钱

很多人低估了记忆系统的成本。

它不是只有用户提问时才调用模型。

它还会在后台做：

```text
消息边界判断
记忆抽取
atomic facts
profile 更新
agent case 提炼
agent skill 聚类
embedding
rerank
多模态解析
```

这些动作都会消耗 token 或模型调用次数。

所以 EverOS 接入团队前，建议先定预算策略。

比如：

```text
个人项目：每天限额 10 元。
内容团队：按 project_id 分账。
研发团队：按 agent_id 看调用占比。
高成本模型只允许在报告生成和复杂分析时使用。
Embedding 和 Rerank 固定低成本模型。
```

4SAPI 可以在这里做账单入口。

至少要能看：

```text
按 Key 的费用。
按模型的费用。
按项目的费用。
按时间的费用。
失败请求和重试成本。
```

没有这层，你只会在月底发现账单突然变厚。

## 12. 日志审计：记忆和模型都要能追

团队 Agent 最怕“没人知道它为什么这么回答”。

长期记忆引入后，这个问题更明显。

一个回答可能来自：

```text
当前上下文
EverOS 搜回来的历史记忆
用户画像
Agent 案例
模型自己的推理
```

所以审计要分两层。

第一层，记忆来源审计。

回答用了哪些记忆？

这些记忆来自哪个 Markdown 文件？

是哪天写入的？

谁写的？

是否经过人工审核？

第二层，模型调用审计。

调用了哪个模型？

请求耗时多少？

花了多少 token？

是否失败或重试？

这部分适合交给 4SAPI。

尤其是团队里多个 Agent 共用一套模型出口时，统一日志会省很多事。

## 13. 数据隐私和脱敏

EverOS 的记忆是本地 Markdown。

这很好。

但也意味着你要认真管理它。

因为 Markdown 太容易打开、复制、提交。

建议建立脱敏规则：

```text
API Key 一律不写入记忆。
密码、Cookie、Token 一律不写入记忆。
客户姓名、手机号、邮箱默认脱敏。
内部域名按项目要求脱敏。
合同和财务数据只写摘要，不写原文。
生产故障报告区分内部版和外发版。
```

如果 Agent 要自动写记忆，最好先让它过一层检查。

比如：

```text
是否包含 key/token/password？
是否包含手机号/邮箱/身份证？
是否包含真实客户名？
是否包含内部 URL？
是否应该写入长期记忆？
```

长期记忆越强，越要谨慎。

因为它不是一次聊天。

它会留下来。

## 14. 和 Obsidian / 知识库结合

EverOS 的 Markdown 目录可以直接用 Obsidian 打开。

这点非常适合个人和小团队。

你可以把 Agent 记忆变成一座可浏览知识库。

比如：

```text
用户画像
任务 episode
原子事实
案例库
技能沉淀
项目规则
```

这和第80期讲的 Obsidian + Codex 其实能接上。

区别是：

```text
Obsidian 更像人的知识工作台。
EverOS 更像 Agent 的记忆运行时。
```

两者不是替代关系。

可以组合：

```text
EverOS 自动写入记忆。
Obsidian 人工浏览和修正。
Git 做版本审计。
4SAPI 管模型调用成本。
```

这套组合很适合个人创作者、独立开发者和小团队。

## 15. 和 Claude Project / Memory 的关系

有人可能会问：

Claude 自己不是也有 Project、Memory、Knowledge 吗？

为什么还要 EverOS？

我的建议是分层看。

Claude Project 更适合：

```text
一个 Claude 工作台里的资料和指令。
```

Claude Memory 更适合：

```text
Claude 产品内部的长期偏好。
```

EverOS 更适合：

```text
跨 Agent、跨工具、可本地管理的记忆运行时。
```

如果你只用 Claude，一个 Project 就够很多场景。

但如果你同时用：

```text
Claude Code
Codex
Hermes
Cursor
自研 Agent
浏览器 Agent
```

那就需要一层外部记忆。

否则每个 Agent 都有自己的小脑袋。

互相不通。

EverOS 适合放在这一层。

4SAPI 则放在更底层的模型调用出口。

## 16. 和 Codex 工作流怎么接

Codex 这类编码 Agent 最适合记三类东西。

第一，仓库事实。

```text
如何启动。
如何测试。
如何构建。
哪些目录不能动。
常见失败原因。
```

第二，用户偏好。

```text
回答要短。
代码改动要少。
不要主动重构。
中文博客保持系列格式。
```

第三，任务经验。

```text
上次某个测试失败怎么修。
某个依赖版本为什么锁住。
某类文章怎么写。
某个客户项目有哪些发版要求。
```

最小接入方式：

```text
任务开始前，Codex 调 /search。
任务结束后，Codex 调 /add。
会话结束或阶段完成时，调 /flush。
```

如果暂时不写插件，也可以手动半自动化：

```text
让 Agent 总结本次任务。
把总结写入 EverOS。
下次开工前先搜相关记忆。
```

不用一开始追求全自动。

能稳定跑起来最重要。

## 17. 和内容生产工作流怎么接

这个系列本身就很适合做示例。

比如写“大模型API中转站”系列文章时，EverOS 可以记：

```text
系列固定导语。
标题格式。
读者画像。
4SAPI 提及位置。
不写违规代理方案。
每篇结尾要有资料来源。
每篇要有检查清单。
短句风格。
```

这些东西现在可能散落在：

```text
发文规范 txt
历史文章
聊天记录
你的口头要求
```

如果写进 EverOS，下次 Agent 写文前先 search：

```text
写 4SAPI 系列文章有什么规范？
EverOS 主题该怎么植入 4SAPI？
上篇文章第87期讲了什么？
```

它就能更稳。

再配 4SAPI：

```text
选题阶段用便宜模型。
资料总结用长上下文模型。
成稿用强写作模型。
校对用低成本模型。
```

这才是内容工厂应该有的模型治理。

## 18. 团队落地的最小架构

如果让我给一个小团队设计最小架构，我会这样拆。

```text
1. EverOS 本地或内网服务
2. 每个项目独立 app_id / project_id
3. Agent 通过 HTTP API 读写记忆
4. 记忆 Markdown 定期脱敏同步到 Git
5. .index 不提交，必要时重建
6. 模型调用统一走 4SAPI
7. 4SAPI 按团队/项目/Agent 分 Key
8. 每周检查高成本调用和失败请求
9. 每月清理过期记忆
```

这套架构不复杂。

但已经覆盖了最关键的点：

```text
记忆可见。
调用可管。
成本可查。
经验可复用。
风险可控。
```

不要一开始就追求大而全。

先把这条线跑顺。

## 19. 怎么向团队介绍 4SAPI 的价值

如果你要向团队介绍这套方案，不建议把 4SAPI 讲成“万能模型平台”。

更好的表达是：

```text
EverOS 让 Agent 有长期记忆。
4SAPI 让这套记忆系统里的模型调用可控。
```

更好的讲法，是把它放进四个真实场景。

第一，安装配置段。

强调 EverOS 的 LLM endpoint 可以统一指向 4SAPI，减少多家模型 Key 分散。

第二，成本治理段。

强调记忆抽取、embedding、rerank 都会持续消耗调用，4SAPI 能看账单和模型分布。

第三，团队权限段。

强调不同项目、不同 Agent、不同成员可以拆 Key 和分组。

第四，故障排查段。

强调 4SAPI 能看错误码、失败请求、模型响应和限流问题。

这四个点比单纯喊“便宜稳定”更有说服力。

因为它们正好对应团队落地痛点。

## 20. 最小检查清单

团队接入 EverOS + 4SAPI 前，过一遍：

```text
[ ] 是否区分个人试用和团队落地
[ ] 是否为不同项目设置 app_id / project_id
[ ] 是否区分 users 记忆和 agents 记忆
[ ] 是否定义哪些内容可以写入长期记忆
[ ] 是否定义敏感信息脱敏规则
[ ] 是否把 .env 加入 .gitignore
[ ] 是否明确 .index 不提交
[ ] 是否决定哪些 Markdown 进入 Git 审计
[ ] 是否把 LLM 入口统一到 4SAPI
[ ] 是否按项目或 Agent 拆分 4SAPI Key
[ ] 是否能查看模型调用日志
[ ] 是否能按项目统计费用
[ ] 是否有高成本模型使用规则
[ ] 是否有人工修正记忆的流程
[ ] 是否定期清理过期记忆
```

这些都做完，EverOS 才不只是一个本地玩具。

它会变成团队 Agent 工作流的一部分。

## 21. 最后总结

EverOS 的核心价值，是把 Agent 记忆变成可读、可改、可版本化的本地文件。

这解决了“Agent 失忆”的问题。

但团队真正要落地，还要解决另外四件事：

```text
谁能用。
谁能改。
花多少钱。
出了问题去哪查。
```

这就是 4SAPI 的位置。

```text
EverOS：长期记忆层。
Git：记忆审计层。
4SAPI：模型调用治理层。
Agent：执行和协作层。
```

把这几层分清楚，你就不会把所有东西都塞进 prompt。

也不会把所有模型 Key 散在每台电脑上。

更不会让 Agent 的记忆变成没人敢碰的黑箱。

一句话总结：

```text
Agent 要长期工作，必须既记得住，也管得住。
```

EverOS 解决“记得住”。

4SAPI 解决“管得住、算得清、查得到”。

这两层接好，才是真正可运营的 Agent 工作流。

## 资料来源与延伸阅读

- EverOS GitHub 仓库：https://github.com/EverMind-AI/EverOS
- EverOS 中文 README：https://github.com/EverMind-AI/EverOS/blob/main/README.zh-CN.md
- EverOS Quickstart：https://github.com/EverMind-AI/EverOS/blob/main/QUICKSTART.md
- EverOS How Memory Works：https://github.com/EverMind-AI/EverOS/blob/main/docs/how-memory-works.md
- EverOS 1.0.0 Migration Notes：https://github.com/EverMind-AI/EverOS/blob/main/docs/migration-to-1.0.0.md
- 4SAPI 接入文档：https://4sapi.apifox.cn/
- 4SAPI 官网：https://4sapi.com/
