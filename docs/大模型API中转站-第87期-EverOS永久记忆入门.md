---
title: "【大模型API中转站】第87期 EverOS永久记忆 | 30分钟让Agent不失忆"
category: 人工智能
tags:
  - 大模型API中转站
  - EverOS
  - Agent记忆
  - Claude
  - Codex
  - 4SAPI
description: "用 EverOS 给 Claude Code、Codex、Hermes 等 Agent 接上本地优先的永久记忆：Markdown 作为可信源，SQLite 管状态，LanceDB 做混合检索，再用 4SAPI 统一模型入口、Key、日志和成本。"
---

# 【大模型API中转站】第87期 EverOS永久记忆 | 30分钟让Agent不失忆

本文是【大模型API中转站】系列的第87篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

最近搭 Agent 工作流，我越来越有一个感觉：

```text
Agent 不一定最缺更大的上下文窗口。
它更缺一层真正能留下来的记忆。
```

Claude Code、Codex、Hermes、OpenClaw、Cursor 这类 Agent 都越来越能干。

它们能读文件。

能跑命令。

能改代码。

能调用工具。

能把一个任务从需求一路推到交付。

但只要你开一个新会话，它又很容易变回第一天认识你的样子。

昨天刚定好的代码风格。

今天忘了。

上周排过的坑。

今天又踩一遍。

你让它记住“这个项目不要用 any 类型”。

它当场点头。

下一轮对话，照样写出来。

这不是模型笨。

这是记忆层没搭好。

这篇就讲一个很适合开发者上手的本地记忆方案：

```text
EverOS。
```

它的重点不是再给 Agent 塞一段更长 prompt。

而是把记忆落成你能打开、能搜索、能编辑、能 Git 管理的 Markdown 文件。

然后用 SQLite 和 LanceDB 在本地做状态和检索。

最后，如果你要把它接到团队里的 Claude、GPT、Gemini 或国产模型调用，就把 4SAPI 放在模型入口层：

```text
EverOS 管记忆。
4SAPI 管模型、Key、费用、日志和路由。
```

这两层合起来，Agent 才开始像一个能长期工作的助手。

## 1. Agent 每天醒来都失忆

做过编码 Agent 的人，大概都受过这个委屈。

昨天它刚陪你定位完一个磨人的 bug。

今天新开会话，它对昨天发生的一切一无所知。

你在上一个会话里说过：

```text
这个项目用 pnpm。
这个服务不能改数据库字段。
这个组件不要再加一层抽象。
这个客户要求所有导出表格都保留中文列名。
这个仓库的测试要先跑 unit，再跑 e2e。
```

这些都是高价值上下文。

但它们通常被锁在上一段聊天里。

关掉就没了。

我们的第一反应，往往是把 prompt 塞得更满。

把历史记录、偏好、项目背景、踩坑记录、规范说明，一股脑塞进上下文窗口。

然后祈祷模型别忘。

这条路很快撞墙。

第一，窗口有上限。

第二，token 要花钱。

第三，越塞越乱。

第四，最关键的是：

```text
你塞进去的那点记忆仍然是一次性的。
```

它不是一套长期记忆系统。

只是一次很长的临时输入。

真正要解决的问题是：

```text
Agent 需要跨会话记忆。
记忆要能被人看见。
记忆要能被人修正。
记忆要能被检索。
记忆要能被迁移。
```

这就是 EverOS 切入的地方。

## 2. EverOS 是什么

EverOS 官方定位是：

```text
面向 agents 和 makers 的 Python library 与 local-first memory runtime。
```

翻成人话：

```text
它是给 Agent 用的本地优先记忆运行时。
```

它做的事情是：

```text
把对话、文件和 Agent 执行轨迹保存成可读 Markdown。
再同步本地 SQLite 和 LanceDB 索引。
最后让 Agent 能跨会话检索和复用这些记忆。
```

这句话里最重要的是“Markdown”。

很多记忆方案是黑箱。

你把文本塞进去。

系统切块、向量化、入库。

搜索时给你一堆相似度结果。

但你很难直接知道：

```text
它到底记住了什么？
为什么这样记？
哪条记忆错了？
怎么删？
怎么改？
能不能 Git diff？
能不能用 Obsidian 打开？
```

EverOS 的答案很直接：

```text
记忆就是本地 Markdown 文件。
```

你想看，就 `cat`。

你想查，就 `grep`。

你想改，就用编辑器打开。

你想版本化，就交给 Git。

这就是它比普通向量库记忆方案更打动开发者的地方。

## 3. EverOS 的本地三件套

EverOS 当前 1.0.x 的核心存储很轻：

```text
Markdown + SQLite + LanceDB
```

三层分工是这样的。

```text
Markdown：唯一可信来源，真正的记忆内容。
SQLite：系统状态、审计日志、队列、边界缓冲区。
LanceDB：向量、BM25 全文检索、标量过滤。
```

注意这个关系：

```text
Markdown 是真相。
SQLite 和 LanceDB 是派生索引。
```

也就是说，真正重要的是 `.md` 文件。

SQLite 和 LanceDB 负责让系统跑得快、查得准、状态可恢复。

如果索引坏了，可以从 Markdown 重建。

这比“记忆只活在数据库里”更让人安心。

你可以把它理解成：

```text
Obsidian 式可读记忆
+ 本地数据库状态管理
+ 向量和全文混合检索
```

它不需要你先拉一套 MongoDB、Elasticsearch、Redis、Kafka。

这点很重要。

因为很多人一看到“Agent 记忆”，脑子里马上出现一整套后端服务。

然后还没开始用，先被部署复杂度劝退。

EverOS 这个版本走的是轻量路线。

对个人开发者和小团队更友好。

## 4. 和 4SAPI 怎么搭

先把关系讲清楚。

EverOS 不是 API 中转站。

它管的是记忆：

```text
对话进入
抽取记忆
写入 Markdown
同步索引
搜索召回
```

4SAPI 管的是模型调用：

```text
统一模型入口
统一 API Key
统一日志
统一账单
统一模型路由
```

EverOS 的配置里，LLM、Embedding、Rerank、多模态都可以通过 OpenAI-compatible endpoint 接入。

这就给 4SAPI 留了很自然的位置。

你可以把模型调用层画成这样：

```text
Agent / 应用
  ↓
EverOS 记忆服务
  ↓
LLM / Embedding / Rerank / Multimodal
  ↓
4SAPI 统一模型入口
  ↓
Claude / GPT / Gemini / Qwen / GLM / 其他模型
```

为什么要多这一层？

第一，少配一堆 Key。

EverOS 默认会涉及 LLM、Embedding、Rerank、多模态。

团队里如果每个开发者都各配各的 Key，很快就乱。

4SAPI 可以把模型调用统一收口。

第二，方便换模型。

今天你用 Qwen 做抽取。

明天你想换 GPT 或 Claude 做更复杂的总结。

如果上层都走统一中转，改配置比改代码舒服得多。

第三，能看账单。

记忆系统会反复做抽取、向量化、重排、总结。

这不是一次性调用。

如果没有成本统计，很容易感觉“不知不觉 token 就没了”。

第四，方便团队审计。

Agent 记忆里可能有项目资料、用户偏好、任务轨迹。

模型调用要留日志。

谁调用了什么模型？

哪一个项目花了多少钱？

失败率多少？

这些都应该在中转层看得到。

所以这篇虽然讲 EverOS，但落地时我更建议这样理解：

```text
EverOS 是 Agent 的长期记忆层。
4SAPI 是 Agent 的模型调用治理层。
```

一个管脑子。

一个管油门和仪表盘。

## 5. 先看当前版本，不要抄旧教程

动手前先提醒一个坑。

网上有些 EverOS / EverMemOS / OpenClaw 相关教程，可能讲的是旧版本。

旧版本里经常会出现：

```text
Docker Compose
MongoDB
Elasticsearch
Redis
Milvus
旧版 /api/v1/memories 路由
```

但当前 EverOS 1.0.x 的本地 OSS 路线已经换成：

```text
Markdown files + SQLite + LanceDB
```

当前主要 HTTP API 是：

```text
POST /api/v1/memory/add
POST /api/v1/memory/flush
POST /api/v1/memory/search
POST /api/v1/memory/get
```

所以不要照着旧教程一顿 `docker-compose up`。

你要认准当前官方 README、中文 README、QUICKSTART 和 migration 文档。

这篇也按当前 1.0.x 的接口写。

## 6. 环境准备

最小前置条件：

```text
Python 3.12+
everos 包
模型服务 API Key
```

安装可以用 pip：

```bash
pip install everos
```

也可以用 uv：

```bash
uv pip install everos
```

如果你只是想先体验本地演示，可以跑：

```bash
everos demo
```

这个 demo 不需要 API Key。

它是一个本地终端可视化演示，用来让你看到：

```text
conversation
→ memory sphere
→ recall
→ source proof
```

但它不是正式的 server-backed memory flow。

真正给 Agent 接记忆，要启动 EverOS 服务，再调用 HTTP API。

## 7. 配置模型 Key

初始化配置：

```bash
everos init
```

它会生成 `.env`。

中文 README 当前默认推荐 DashScope，用同一个 DashScope API Key 配 LLM、Embedding、Rerank 三个核心能力。

示例大致是：

```env
EVEROS_LLM__MODEL=qwen-plus
EVEROS_LLM__API_KEY=<DASHSCOPE_API_KEY>
EVEROS_LLM__BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1

EVEROS_EMBEDDING__MODEL=text-embedding-v4
EVEROS_EMBEDDING__API_KEY=<DASHSCOPE_API_KEY>
EVEROS_EMBEDDING__BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1

EVEROS_RERANK__MODEL=gte-rerank-v2
EVEROS_RERANK__API_KEY=<DASHSCOPE_API_KEY>
EVEROS_RERANK__BASE_URL=https://dashscope.aliyuncs.com
```

英文 README 里也提到 OpenRouter + DeepInfra 的默认组合。

这里不用纠结用哪一家。

重点是：

```text
EverOS 支持 OpenAI-compatible endpoint。
```

如果你用 4SAPI 做统一入口，就可以把 LLM 这类 `BASE_URL` 指到 4SAPI 兼容端点。

团队里建议这样做：

```text
个人测试：可以直接填某个模型厂商 Key。
团队落地：建议走 4SAPI 统一 Key、统一日志、统一账单。
```

还有一个细节：

```text
.env 一定加入 .gitignore。
```

不要把 Key 提交到仓库。

这种坑，踩一次就够疼了。

## 8. 启动 EverOS 服务

启动服务：

```bash
everos server start
```

默认会跑在：

```text
http://127.0.0.1:8000
```

新开一个终端，健康检查：

```bash
curl http://127.0.0.1:8000/health
```

正常返回：

```json
{"status":"ok"}
```

看到这行，说明本地记忆服务活了。

注意两点。

第一，服务默认绑定 `127.0.0.1`。

这适合本地开发。

如果要暴露给团队或远程机器，不要裸奔。

应该先加认证、网关和访问控制。

第二，server 会查找 `.env`。

常见顺序包括：

```text
--env-file 指定路径
当前目录 ./.env
~/.config/everos/.env
~/.everos/.env
```

如果你发现明明填了 Key 但服务读不到，先检查当前工作目录和 `.env` 位置。

## 9. 写入第一段对话

EverOS 不是直接“写一句事实”。

它更偏 conversation-level ingest。

也就是你把一批 messages 交给它。

它先按 session 累积。

然后通过边界检测或手动 flush 抽取记忆。

先写入一段对话：

```bash
TS=$(($(date +%s)*1000))

curl -X POST http://127.0.0.1:8000/api/v1/memory/add \
  -H 'Content-Type: application/json' \
  -d "{
    \"session_id\": \"demo-001\",
    \"app_id\": \"default\",
    \"project_id\": \"default\",
    \"messages\": [
      {\"sender_id\": \"alice\", \"role\": \"user\", \"timestamp\": $TS, \"content\": \"Alice 偏好用 TypeScript，讨厌 any 类型。\"},
      {\"sender_id\": \"alice\", \"role\": \"user\", \"timestamp\": $((TS+10000)), \"content\": \"Alice 的项目默认使用 pnpm 管依赖。\"}
    ]
  }"
```

返回里你可能看到：

```json
{
  "data": {
    "message_count": 2,
    "status": "accumulated"
  }
}
```

`accumulated` 的意思是：

```text
消息已经进入 session buffer。
但还没抽取成记忆单元。
```

为了演示确定性，我们手动 flush。

## 10. 手动 flush 抽取记忆

执行：

```bash
curl -X POST http://127.0.0.1:8000/api/v1/memory/flush \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"demo-001","app_id":"default","project_id":"default"}'
```

正常会返回类似：

```json
{
  "data": {
    "status": "extracted"
  }
}
```

这一步会调用模型做抽取。

所以如果你接了 4SAPI，就能在 4SAPI 后台看到这类调用的模型、耗时、失败和成本记录。

这就是中转站在 Agent 记忆系统里的第一个价值点：

```text
不是只有聊天才花 token。
记忆抽取、总结、重排也会持续花 token。
```

如果没有日志，你很难知道钱花在哪。

## 11. 把记忆搜回来

现在搜索：

```bash
curl -X POST http://127.0.0.1:8000/api/v1/memory/search \
  -H 'Content-Type: application/json' \
  -d '{
    "user_id": "alice",
    "app_id": "default",
    "project_id": "default",
    "query": "Alice 写代码有什么偏好？",
    "top_k": 5
  }'
```

如果一切正常，响应里应该能看到和 TypeScript、`any`、pnpm 相关的记忆。

如果第一次为空，等几秒再搜。

原因是：

```text
Markdown 写入是同步的。
LanceDB 索引追上来是异步的。
```

官方文档里也强调了这个一致性模型。

写入路径更强。

搜索路径是最终一致。

本地开发时别慌，稍等或用 cascade sync 处理索引队列。

## 12. 打开真正的记忆文件

现在最有意思的一步来了。

打开：

```text
~/.everos/
```

默认结构大致是：

```text
~/.everos/
├── default_app/
│   └── default_project/
│       ├── users/
│       │   └── alice/
│       │       ├── user.md
│       │       ├── episodes/
│       │       │   └── episode-YYYY-MM-DD.md
│       │       ├── .atomic_facts/
│       │       └── .foresights/
│       └── agents/
│           └── <agent_id>/
│               ├── .cases/
│               └── skills/
├── .index/
│   ├── sqlite/
│   └── lancedb/
└── .tmp/
```

你可以直接：

```bash
cat ~/.everos/default_app/default_project/users/alice/user.md
```

也可以打开 episode 文件。

你会看到 YAML frontmatter 和 Markdown 内容。

这不是 dashboard 里的黑箱记录。

这是实打实的本地文件。

你可以读。

可以改。

可以 diff。

可以放进 Obsidian。

可以备份。

可以做团队内部审阅。

这就是 EverOS 的核心卖点：

```text
Agent 的记忆不是看不见的向量。
而是你能打开的 Markdown。
```

## 13. 为什么 Markdown 是可信源很重要

在真实项目里，Agent 记错东西并不稀奇。

它可能把临时偏好当成长期偏好。

可能把某次异常情况总结成通用规则。

可能把一个实验结论记成正式决策。

如果记忆藏在数据库里，修起来很麻烦。

你要写脚本。

要查 ID。

要走 API。

有时还不知道删哪条。

Markdown 可信源的好处是：

```text
错了就打开改。
过期就删。
重要就补说明。
需要审计就看 Git diff。
```

这对 Agent 工作流非常关键。

因为长期记忆不应该完全交给模型自说自话。

人必须能介入。

尤其是团队场景：

```text
编码规范
客户要求
项目约束
安全边界
成本策略
常见故障处理
```

这些记忆一旦错，影响会很大。

所以我更喜欢“可编辑记忆”，而不是“漂亮但不可控的记忆黑箱”。

## 14. 用户记忆和 Agent 记忆要分开

EverOS 里有一个很重要的设计：

```text
users/
agents/
```

用户记忆和 Agent 记忆分开。

用户轨道记录：

```text
用户是谁
用户偏好
用户情景日志
用户画像
```

Agent 轨道记录：

```text
Agent 做过什么任务
哪些案例可复用
哪些模式能变成 Skill
```

这点很高级。

很多记忆系统只围绕用户画像打转。

但编码 Agent 真正要成长，不能只记住“你喜欢什么”。

它还要记住：

```text
这个仓库怎么跑测试。
这个团队怎么发版。
这个项目有哪些坑。
上次同类 bug 是怎么修的。
哪些执行轨迹可以复用成 Skill。
```

这就是 Agent 记忆。

用户记忆和 Agent 记忆混在一起，会越来越脏。

分开之后，系统更容易治理。

## 15. 多模态和文档摄取

EverOS 还支持非文本内容摄取。

比如：

```text
图片
PDF
音频
Office 文档
HTML
Email
```

需要安装可选 extra：

```bash
uv pip install 'everos[multimodal]'
```

或者：

```bash
pip install 'everos[multimodal]'
```

这里注意一个坑：

```text
Office 文档解析需要系统安装 LibreOffice。
```

没有 LibreOffice 时，`.docx`、`.pptx`、`.xlsx` 这类文件可能会失败。

PDF、图片、音频、HTML 不受这个限制。

如果你要把 EverOS 用在内容团队、产品资料库、客服知识沉淀里，这块就很有用。

但我建议第一天不要贪多。

先跑通文本记忆闭环。

再接多模态。

## 16. 接到 Codex / Claude Code 怎么想

真正落地时，你可以把 EverOS 当成一个本地服务。

Agent 每次完成一段任务后，调用：

```text
/api/v1/memory/add
```

把关键对话、决策和任务轨迹写进去。

会话结束时调用：

```text
/api/v1/memory/flush
```

下次开工前调用：

```text
/api/v1/memory/search
```

把相关记忆搜回来，塞进 Agent 上下文。

最小闭环就是：

```text
开始任务前：search 相关记忆。
执行任务中：正常做事。
任务完成后：add 任务过程。
会话结束时：flush 抽取记忆。
```

如果你用 Codex 做长期项目，可以记：

```text
仓库运行方式
测试命令
用户偏好
代码风格
不能碰的目录
常见报错
上次修复方式
```

如果你用 Claude Code，也类似。

区别不大。

关键是不要指望 Agent 自己凭空记住。

要给它一个明确的记忆读写接口。

## 17. 为什么这里要接入 4SAPI

因为 Agent 记忆不是“装个本地库就结束”。

只要 EverOS 开始负责真实记忆，它就会持续调用模型：

```text
抽取
总结
多模态解析
Embedding
Rerank
搜索增强
后续 Reflection
```

个人试用时，直接填厂商 Key 没问题。

但团队场景就会马上遇到四个问题：

```text
Key 怎么管？
不同模型怎么切？
费用怎么分摊？
调用失败怎么排查？
```

这就是 4SAPI 的位置。

你可以把 EverOS 的模型配置指向 4SAPI 兼容入口。

然后在 4SAPI 里统一做：

```text
模型路由
Key 管理
调用日志
费用统计
错误排查
团队权限
```

这对长期 Agent 工作流特别重要。

因为记忆系统不是一次性调用。

它会跟着你的团队天天跑。

没有 4SAPI 这类中转层，后面成本和日志迟早会散。

## 18. 几个容易翻车的点

第一，照旧教程装一堆服务。

当前 EverOS 1.0.x OSS 版不需要 MongoDB、Elasticsearch、Redis 那套。

认准 Markdown + SQLite + LanceDB。

第二，把 `.env` 提交进仓库。

不要。

第三，写完立刻 search，结果为空就以为失败。

Markdown 已经写了，索引可能还在追。

稍等几秒，或者处理 cascade 队列。

第四，把所有记忆都当长期偏好。

有些只是某次任务的临时上下文。

不要让用户画像被临时信息污染。

第五，团队共用一个模型 Key。

短期方便。

长期一定乱。

更好的方式是走 4SAPI 分组、分项目、分模型统计。

第六，完全不审查记忆。

长期记忆不是越多越好。

能看、能删、能改，才是优势。

## 19. 最小检查清单

给 Agent 接 EverOS 前，先过一遍：

```text
[ ] Python 是否为 3.12+
[ ] 是否安装 everos
[ ] 是否跑过 everos demo
[ ] 是否执行 everos init
[ ] .env 是否已加入 .gitignore
[ ] LLM / Embedding / Rerank 是否配置好
[ ] 是否考虑用 4SAPI 统一模型入口
[ ] everos server start 是否正常
[ ] /health 是否返回 ok
[ ] /memory/add 是否能写入 messages
[ ] /memory/flush 是否能抽取记忆
[ ] /memory/search 是否能搜回结果
[ ] 是否打开过 ~/.everos 检查 Markdown
[ ] 是否知道 .index 是派生索引，不是唯一记忆
[ ] 团队使用时是否有 Key、日志、成本和权限治理
```

跑通这些，再考虑接入 Codex、Claude Code、Hermes 或自己的 Agent。

## 20. 最后总结

EverOS 解决的是 Agent 工作流里一个很朴素的问题：

```text
别让 Agent 每天醒来都失忆。
```

它没有把记忆做成玄学。

而是落成很工程化的一套东西：

```text
Markdown 做可信源。
SQLite 管状态。
LanceDB 做混合检索。
HTTP API 给 Agent 调用。
```

你可以 30 分钟跑通第一条记忆。

写进去。

flush 出来。

search 回来。

再打开 `~/.everos` 看见它。

这就是它最打动我的地方：

```text
记忆不是黑箱。
记忆是一堆你能打开、能编辑、能 Git 管理的文件。
```

但如果你要把它放进团队工作流，别只盯着本地文件。

还要补上模型调用治理。

这时 4SAPI 的位置很明确：

```text
EverOS 负责长期记忆。
4SAPI 负责统一模型、Key、日志、账单和路由。
```

个人开发者可以先用 EverOS 把 Agent 记忆跑起来。

团队用户更建议从第一天就把 4SAPI 接进模型层。

不然后面你会发现：

```text
记忆是有了。
账单看不清。
Key 管不住。
模型切不动。
日志查不到。
```

真正能长期运行的 Agent，不只是会回答。

它要记得住。

也要算得清。

## 资料来源与延伸阅读

- EverOS GitHub 仓库：https://github.com/EverMind-AI/EverOS
- EverOS 中文 README：https://github.com/EverMind-AI/EverOS/blob/main/README.zh-CN.md
- EverOS Quickstart：https://github.com/EverMind-AI/EverOS/blob/main/QUICKSTART.md
- EverOS How Memory Works：https://github.com/EverMind-AI/EverOS/blob/main/docs/how-memory-works.md
- EverOS 1.0.0 Migration Notes：https://github.com/EverMind-AI/EverOS/blob/main/docs/migration-to-1.0.0.md
- 4SAPI 接入文档：https://4sapi.apifox.cn/
- 4SAPI 官网：https://4sapi.com/
