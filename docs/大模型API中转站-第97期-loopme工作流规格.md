---
title: "【大模型API中转站】第97期 loop-me | 把重复动作磨成Workflow"
category: 人工智能
tags:
  - 大模型API中转站
  - 4SAPI
  - Skills
  - loop-me
  - Workflow
  - AI Agent
description: "从 mattpocock 开源 Skills 系列里的 in-progress Skill loop-me 出发，拆解 Loop 透镜、grilling 拷问纪律、workflows/*.md 规格文件和 Push right + Brief 设计哲学，并说明真正可委托的 Agent 工作流为什么需要 4SAPI 做模型入口、Key、日志和成本治理。"
---

# 【大模型API中转站】第97期 loop-me | 把重复动作磨成Workflow

本文是【大模型API中转站】系列的第97篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

最近 mattpocock 的开源 Skills 系列里，新添了一个 in-progress Skill：

```text
loop-me
```

它很小。

小到核心说明没几行。

但它背后的思路很有意思。

很多人做 Agent，一上来就问：

```text
我能不能让 AI 帮我做一个自动化？
我能不能让 AI 每天帮我干活？
我能不能让 AI 写代码、发文章、整理客户、跑工作流？
```

loop-me 的起点不是这个。

它问的是：

```text
你的生活和工作里，有哪些事情正在重复发生？
```

这就是它的核心视角：

```text
Loop 透镜。
```

一个 loop，不是技术概念。

它是生活里可识别的重复模式：

```text
每周复盘。
每天早晨处理消息。
每次发文章前做检查。
每次接单都要问需求。
每次发布视频都要选题、脚本、素材、分发。
每次客户咨询都要判断意图、查资料、回复、归档。
```

如果一件事会重复，它就有预测性。

如果它有预测性，它就可能被委托。

这句话是 loop-me 最值得写的地方：

```text
可预测，才可委托。
```

Agent 不是先从“我要造一个 Agent”开始。

Agent 应该先从“我有哪些 loop 正在反复发生”开始。

再把这些 loop 拷问成 workflow。

而 workflow，才是后续自动化、Agent、n8n、Claude Code、OpenClaw、Codex、4SAPI 模型路由真正能落地的规格。

## 1. loop-me 到底是什么

一句话：

```text
loop-me 是一个把重复模式拷问成 workflows/*.md 规格文件的 Skill。
```

它不负责实现。

不负责直接写代码。

不负责立刻帮你跑自动化。

它只做一件事：

```text
把某个 loop 问清楚，直到它可以被实现。
```

这个产物放在：

```text
workflows/*.md
```

这些文件是唯一真相源。

会话里可以创建、编辑、删除它们。

但最终交付物不是聊天记录。

不是一段 prompt。

不是一张脑图。

而是一份可以被实现 Agent 读懂的 workflow spec。

loop-me 的完成标准非常硬：

```text
实现 agent 读完 spec 后，不需要再问任何问题。
```

只要还有疑点，就没完成。

这点和很多“AI 帮我总结一下需求”的工具不一样。

普通总结是：

```text
你说了什么，我帮你整理。
```

loop-me 更像：

```text
你没说清楚什么，我继续追问。
```

这才是它的价值。

## 2. 为什么不是从“我要一个 AI Agent”开始

大多数 Agent 项目失败，不是模型不够强。

而是需求太糊。

常见说法是：

```text
我想做一个自动发文章的 Agent。
我想做一个自动跟进客户的 Agent。
我想做一个帮我管理生活的 Agent。
我想做一个自动赚钱的 Agent。
```

这些都太大。

Agent 不知道边界。

人也不知道验收标准。

最后就会变成：

```text
看起来很智能。
但稳定跑不起来。
```

loop-me 反过来。

它先找 loop：

```text
你每周固定要做什么？
你每天打开电脑后第一件事是什么？
你每次发文章前都检查什么？
你每次接客户需求时都问什么？
你每次做项目复盘时都看哪些数据？
你有哪些重复动作自己都没意识到？
```

这比“我要一个 Agent”更接近真实工作。

因为真正值得自动化的，不是一个宏大愿望。

而是一个可重复的动作链。

比如：

```text
每周五下午复盘本周 GitHub issue。
每天上午 9 点整理企业微信未回复客户。
每次写完博客后检查标题、SEO、4SAPI 露出和资料来源。
每次跑 n8n 工作流后检查失败节点和模型调用成本。
每次上传视频前生成标题、简介、标签和封面提示词。
```

这些才是 workflow 的原料。

## 3. Loop 和 Workflow 的区别

loop-me 里有一个很重要的区分：

```text
Loop = 重复模式。
Workflow = 某个 loop 的规格说明书。
```

举个例子。

你每周都要写一篇系列博客。

这是 loop。

这个 loop 的一次运行，就是某一周的那篇博客从选题到发布。

而 workflow 是什么？

workflow 是这件事的说明书：

```text
触发条件是什么？
输入材料来自哪里？
要先读哪些旧稿？
标题怎么定？
结构怎么展开？
什么时候要查资料？
什么时候要问用户？
什么时候要人工确认？
需要用哪些模型？
每次调用预算是多少？
输出文件放在哪里？
完成标准是什么？
失败怎么处理？
```

这个 workflow 写进：

```text
workflows/blog-publishing.md
```

以后每次写博客，就是运行这个 workflow 的一个实例。

这就很工程化。

因为你把“我每次大概这么做”变成了“任何实现者都能照着做”。

这也是 Agent 能真正接手的前提。

## 4. grilling：一次一问，把未知问干净

loop-me 和 grill-me 共用一套 grilling 纪律。

这套纪律非常适合做需求收敛：

```text
一次只问一个问题。
每个问题都附带推荐答案。
顺着决策树一层一层走。
能查工作区就先查，不把该自己调研的问题扔给用户。
用文件保存状态，跨会话继续。
```

为什么一次只问一个？

因为用户不是产品经理。

更不是需求文档机器。

你一口气问：

```text
触发条件是什么？
输入是什么？
输出是什么？
异常怎么处理？
谁来审核？
预算是多少？
用什么模型？
日志在哪里看？
```

用户会直接迷失。

更好的问法是：

```text
这个 workflow 应该由什么触发？
推荐答案：由“每次新稿素材进入 inbox 文件夹”触发，而不是每天定时触发，因为事件触发更省成本，也更贴近真实节奏。
```

用户只要回答一个问题。

而且有推荐答案可抄。

这就降低了决策成本。

一次一问，不是慢。

它是为了让每个分支都被真正解决。

## 5. 推荐答案为什么重要

很多 AI 提问工具的问题在于：

```text
它只会问，不会判断。
```

比如：

```text
你希望工作流如何触发？
你希望输出什么？
你希望用哪个模型？
```

这些问题看似合理。

但用户可能根本不知道怎么答。

loop-me 的 grilling 要求每个问题都带推荐答案。

这意味着 Agent 不能只是客服。

它必须有观点。

例如：

```text
问题：这条客户跟进 workflow 需要人工审核吗？
推荐答案：需要。低风险咨询可以自动草拟回复，但第一次正式发送给客户前，应该给销售一个 brief，由人点击确认。
```

这就是专业感。

它不是把所有选择推给人。

而是把人压缩到关键决策。

这也对应 loop-me 的一个设计哲学：

```text
人类时间最贵。
```

Agent 应该做足准备，再让人做一次晚到的判断。

## 6. Push right + Brief：把人类决策尽量后移

loop-me 里有两个词很关键：

```text
Push right
Brief
```

Push right 的意思是：

```text
把人工检查点尽量往后推。
```

不要让用户每一步都看。

不要问：

```text
这个标题可以吗？
这个摘要可以吗？
这个素材可以吗？
这个文件可以吗？
这个模型可以吗？
```

这会把人拖回工作流中间。

更好的方式是：

```text
Agent 先完成尽可能多的工作。
最后把关键结果打包成 brief。
用户只做一次判断。
```

Brief 不是原始输出。

不是一堆草稿。

不是完整日志。

而是决策摘要：

```text
这次 workflow 做了什么。
为什么这么做。
产物在哪里。
风险点是什么。
建议用户批准还是退回。
```

举个博客发布 workflow 的 brief：

```text
本次生成第97期草稿。
主题：loop-me 如何把重复动作规格化。
已核对来源：loop-me SKILL.md、grill-me SKILL.md。
已植入 4SAPI：执行层模型治理、Key、日志、成本。
建议：可发布。
需要人工确认：标题是否保留 loop-me 英文名。
文件：workflows/blog-publishing.md
```

这比让用户从头读一堆草稿高效得多。

## 7. workflows/*.md 为什么是唯一真相源

loop-me 最工程化的一点，是把状态放到文件里。

不是放在聊天上下文里。

不是放在模型记忆里。

而是放在：

```text
workflows/*.md
NOTES.md
```

`workflows/*.md` 存每条 workflow 的规格。

`NOTES.md` 存用户世界里的原始笔记：

```text
他用什么工具。
他有哪些渠道。
他怎么称呼自己的流程。
他有哪些固定术语。
他有哪些业务规则。
```

这件事很重要。

因为聊天上下文会丢。

记忆会漂。

但文件可以：

```text
版本化。
diff。
回滚。
Review。
跨会话继续。
交给另一个 Agent 实现。
```

这就是 Real Engineers 会喜欢它的原因。

真正的工程协作，不是靠“我记得你上次说过”。

而是靠明确的文件。

## 8. 一个好 workflow spec 应该写什么

loop-me 反模板化。

它不强制每个 workflow 都有 AI、checkpoint、schedule。

结构应该由场景决定。

但为了理解，我们可以看一个典型规格会包含什么。

比如：

```text
# Workflow: Weekly Blog Publishing

## Loop
每周写一篇【大模型API中转站】系列博客。

## Trigger
当用户提供选题素材，或 inbox/blog-ideas.md 出现新条目时启动。

## Inputs
- 用户原始素材
- 博客发文规范
- 最近 10 篇同系列旧稿
- 需要核对的外部资料链接

## Steps
1. 读取发文规范和旧稿风格。
2. 抽取选题主线。
3. 浏览外部资料并记录来源。
4. 生成文章结构。
5. 写成 Markdown 新稿。
6. 检查 4SAPI 露出是否自然。
7. 检查资料来源、代码块、占位词。
8. 生成 brief 给用户确认。

## Model Routing
- 资料归纳：低成本长上下文模型。
- 结构设计：强推理模型。
- 初稿生成：写作模型。
- 最终校验：便宜模型 + 规则扫描。

## Budget
单次 workflow 预算不超过 2 元。

## Checkpoint
只在文章完成后给用户 brief。

## Done
实现 Agent 读完本 spec 后，不需要再问问题即可产出新稿。
```

这就是 workflow spec。

不是 prompt。

不是愿望。

不是“帮我写文章”。

而是一份可执行说明书。

## 9. loop-me 和 grill-me 的区别

loop-me 和 grill-me 很像。

但产物不同。

grill-me 更通用。

它可以拷问任何计划：

```text
产品方案
功能设计
创业计划
文章选题
技术架构
个人目标
```

目标是把一个计划问清楚。

loop-me 更窄。

它只关心：

```text
workflow specs。
```

也就是：

```text
某个重复 loop 如何被规格化。
```

可以这样理解：

```text
grill-me：把任意计划问清楚。
loop-me：把重复模式问成 workflows/*.md。
```

这个边界很好。

因为 Skills 最怕什么都做。

越具体，越容易稳定。

loop-me 的窄，反而是它的强。

## 10. 它和 to-prd 的区别

很多人会把 loop-me 和 PRD 生成工具混在一起。

其实差别很大。

to-prd 更像：

```text
已有对话 -> 整理成 PRD。
```

loop-me 更像：

```text
模糊重复模式 -> 主动拷问 -> 消尽疑点 -> workflow spec。
```

一个是总结。

一个是挖掘。

一个更像文档整理。

一个更像产品经理和自动化架构师。

这也是 loop-me 对 Agent 生态的启发：

```text
真正有价值的 Agent，不是只听命令。
它应该主动发现你没意识到的重复。
```

比如你总是在周日晚上整理下周任务。

你总是在发文章后手动分发到多个平台。

你总是在客户付款后手动建项目目录、写欢迎邮件、拉群。

你可能觉得这些只是习惯。

但 Agent 应该能指出：

```text
这是 loop。
值得规格化。
```

## 11. 在 Skills 生态里的位置

如果把 Skills 看成一个工具箱，loop-me 的位置很清楚。

它不是执行类 Skill。

它不是写代码 Skill。

它不是搜索资料 Skill。

它是规格类 Skill。

可以这样分层：

```text
发现层：loop-me 找重复模式。
拷问层：grilling 消尽疑点。
规格层：workflows/*.md 成为唯一真相源。
执行层：Codex、Claude Code、OpenClaw、n8n、脚本或人工流程实现。
治理层：4SAPI 管模型入口、Key、日志、成本和权限。
```

这条链路很重要。

很多人跳过前三层，直接进入执行层。

结果就是：

```text
Agent 很忙。
但人一直在补充说明。
```

loop-me 的目标是反过来：

```text
先把说明书写到实现 Agent 不用再问。
再交给执行层。
```

## 12. 4SAPI 应该出现在 workflow spec 的哪里

如果 workflow 只是一套人工流程，不一定需要 4SAPI。

但只要 workflow 里有 AI 模型调用，就应该把 4SAPI 这类模型网关写进规格。

原因很简单：

```text
workflow 会重复运行。
重复运行就会有成本。
有成本就要有预算。
有失败就要有日志。
有多人使用就要有 Key 权限。
```

这不是营销词。

这是工程现实。

比如一个内容生产 workflow：

```text
每天采集 50 个选题。
每个选题调用模型做摘要。
筛出 10 个候选。
再让模型写标题和大纲。
再生成配图提示词。
```

如果没有 4SAPI，你很难知道：

```text
今天跑了几次？
哪个模型最贵？
哪一步失败最多？
哪个用户触发了高成本请求？
是否应该把草稿模型换成低成本模型？
```

所以 workflow spec 里应该明确：

```text
Base URL: https://4sapi.com/v1
API Key: 使用对应 workflow 的 4SAPI 专用 Key
Model: 从 4SAPI 模型广场复制
Budget: 每次运行预算上限
Logs: 以 4SAPI 后台调用记录为准
Fallback: 主模型失败时切换备用模型
```

这就是 4SAPI 在 loop-me 文章里的自然位置。

不是“为了接模型而接模型”。

而是：

```text
workflow 一旦可重复，就必须可治理。
```

## 13. workflow spec 里的模型路由怎么写

一个成熟 workflow 不应该只写：

```text
调用 AI。
```

这太粗。

应该写清楚模型角色。

例如：

```text
Planner Model：负责拆解任务，用强推理模型。
Executor Model：负责批量执行，用低成本模型。
Reviewer Model：负责审查结果，用稳定模型。
Image Model：负责配图，用图片模型。
TTS Model：负责配音，用语音模型。
```

如果通过 4SAPI，可以在 spec 里写：

```text
Planner：4SAPI-Key-Workflow-Plan
Executor：4SAPI-Key-Workflow-Run
Reviewer：4SAPI-Key-Workflow-Review
Image：4SAPI-Key-Workflow-Image
```

这样做有三个好处。

第一，成本可控。

执行步骤量最大，不一定要用最贵模型。

第二，日志清楚。

哪一步调用出错，很容易从 4SAPI 后台查。

第三，权限隔离。

图片、语音、长上下文、强推理，可以分开授权。

这才是 Agent 工作流进入生产前该有的样子。

## 14. 三个适合 loop-me 的 workflow 示例

为了让这个 Skill 更具体，我们可以看三个场景。

### 14.1 博客发布 workflow

loop：

```text
每次用户给一段素材，就写一篇系列博客。
```

loop-me 应该拷问：

```text
系列固定导语是什么？
标题格式是什么？
是否必须营销 4SAPI？
旧稿风格在哪里？
外部资料怎么核对？
什么时候需要浏览网页？
怎么避免重复旧稿？
完成后要不要检查代码块？
```

产物：

```text
workflows/blog-publishing.md
```

执行层：

```text
Codex 写文件。
浏览器核对来源。
4SAPI 负责模型调用治理。
```

### 14.2 客户跟进 workflow

loop：

```text
每天上午处理未回复客户消息。
```

loop-me 应该拷问：

```text
消息来自哪里？
客户怎么分级？
什么问题能自动回复？
什么问题必须转人工？
brief 给谁看？
回复前是否需要审批？
日志写到哪里？
模型预算是多少？
```

产物：

```text
workflows/customer-followup.md
```

执行层：

```text
n8n 监听 Webhook。
4SAPI 调模型分类和草拟回复。
人工只看 brief 并确认发送。
```

### 14.3 每周经营复盘 workflow

loop：

```text
每周日晚上整理本周数据和下周行动。
```

loop-me 应该拷问：

```text
数据源有哪些？
哪些指标必须看？
异常阈值是多少？
输出是日报、周报还是 brief？
哪些结论需要模型解释？
哪些决策必须人工确认？
是否要生成下周任务？
```

产物：

```text
workflows/weekly-business-review.md
```

执行层：

```text
脚本或 n8n 拉数据。
4SAPI 做摘要、解释和建议。
人只看最终 brief。
```

这些例子说明，loop-me 的价值不是“生成一个漂亮模板”。

而是把重复生活里的模糊动作，磨成可运行的说明书。

## 15. 为什么说它是 Real Engineers 的 Skill

“For Real Engineers” 这个说法很贴。

因为 loop-me 的思路很工程师。

工程师不相信口头记忆。

工程师相信：

```text
文件。
版本。
diff。
边界。
验收标准。
可重复执行。
```

loop-me 正好把这些东西引进了个人工作流和 Agent 自动化。

它不预设所有 workflow 都必须有 AI。

不预设所有 workflow 都必须定时触发。

不预设所有 workflow 都必须有人工审批。

它只要求一件事：

```text
把未知问完。
```

问完之后，结构由场景决定。

这叫反模板化。

很多自动化教程的问题，是一上来给你一个固定结构：

```text
触发器
AI 节点
数据库
通知
发布
```

但真实世界不是每条流程都长这样。

有些 workflow 不需要 AI。

有些不需要 schedule。

有些不能自动执行，只能生成 brief。

loop-me 的克制，反而让它更适合长期使用。

## 16. 对个人和团队的启发

如果你是个人用户，可以从三个 loop 开始：

```text
每天重复做的事。
每周重复做的事。
每次项目都会做的事。
```

把它们写进：

```text
NOTES.md
```

再让 loop-me 拷问成：

```text
workflows/*.md
```

如果你是团队，更应该这样做。

因为团队里的重复成本更高。

比如：

```text
每次上线前检查。
每次客户交付。
每次内容发布。
每次工单分流。
每次周会复盘。
每次模型调用账单复盘。
```

这些流程如果不规格化，就会靠老人经验。

老人一走，流程就断。

如果变成 workflow spec，就可以：

```text
新人照着做。
Agent 照着做。
外包照着做。
管理者照着验收。
```

再配合 4SAPI 做模型调用治理，团队就能知道每条 workflow 的 AI 成本和失败情况。

这比“我们有一个很聪明的 Agent”更重要。

## 17. 最小使用路线

如果你想试 loop-me，可以按这个顺序来。

第一步，列出你最近一周重复了三次以上的事。

```text
发文章。
回复客户。
整理素材。
查数据。
开会前准备。
发货后跟进。
跑模型测试。
```

第二步，选一个最痛的。

不要选最大最复杂的。

选一个你最烦、最重复、最容易描述的。

第三步，让 loop-me 只问这一个 loop。

不要一次塞十个。

第四步，把问出来的规格保存到：

```text
workflows/xxx.md
```

第五步，检查完成标准：

```text
另一个 Agent 读完后，还需要问问题吗？
```

如果还需要问，就继续 grilling。

第六步，才进入实现。

实现可以用：

```text
n8n
Codex
Claude Code
OpenClaw
脚本
人工 SOP
```

如果实现里要调模型，就把 4SAPI 写进 workflow：

```text
Base URL: https://4sapi.com/v1
Key: workflow 专用 Key
Model: 按角色选择
Budget: 单次运行上限
Logs: 4SAPI 后台可查
```

这条路径很稳。

先规格，再实现。

## 18. 常见误区

第一，把 loop-me 当任务管理工具。

它不是 TODO list。

它关心的是重复 loop，不是零散待办。

第二，把 workflow spec 写成模板填空。

loop-me 明确反模板化。

结构应该由场景长出来。

第三，没问完就开工。

只要实现 Agent 还需要问问题，就说明 spec 没完成。

第四，所有 checkpoint 都往前放。

这会让人一直被打断。

应该 Push right，让人晚一点、少一点、看 brief。

第五，不写成本和日志。

只要 workflow 会调用模型，就要写成本和日志。

这里 4SAPI 很适合做统一底座。

第六，把 Agent 当成猜谜机器。

Agent 不应该靠猜。

它应该靠规格。

## 19. 最后总结

loop-me 最有价值的地方，不是创建了一个 `workflows/*.md` 文件。

而是它提醒我们：

```text
一个不能被说清楚的流程，就不能被稳定委托。
```

很多人做 Agent，太急着执行。

但真正的 Agent 工程，第一步不是执行。

而是规格化。

先用 loop 透镜找到重复模式。

再用 grilling 纪律把未知问完。

再把规格写进 workflow 文件。

最后才交给 n8n、Codex、Claude Code、OpenClaw 或其他执行层。

如果 workflow 里有 AI 模型调用，就用 4SAPI 把模型入口、Key、日志、成本和权限管起来。

一句话：

```text
loop-me 负责把重复动作问清楚。
执行 Agent 负责把 workflow 跑起来。
4SAPI 负责让模型调用可控、可查、可复盘。
```

真正成熟的 Agent，不是更会猜。

而是更少问废话。

## 资料来源与延伸阅读

- loop-me Skill 目录：https://github.com/mattpocock/skills/tree/main/skills/in-progress/loop-me
- loop-me SKILL.md：https://raw.githubusercontent.com/mattpocock/skills/main/skills/in-progress/loop-me/SKILL.md
- grill-me SKILL.md：https://raw.githubusercontent.com/mattpocock/skills/main/skills/productivity/grill-me/SKILL.md
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 文档：https://4sapi.apifox.cn/
- 4SAPI n8n 配置教程：https://4sapi.apifox.cn/8328708m0
- 4SAPI Claude Code CLI 安装与配置：https://4sapi.apifox.cn/347624c0
- 4SAPI OpenCode 接入配置：https://4sapi.apifox.cn/8323429m0
