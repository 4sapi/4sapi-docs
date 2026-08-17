---
title: "【大模型API中转站】第98期 AnySearch | 给Codex装搜索大脑"
category: 人工智能
tags:
  - 大模型API中转站
  - 4SAPI
  - AnySearch
  - Codex
  - Claude Code
  - AI Agent
  - 搜索增强
description: "讲解 AnySearch Skill 如何给 Codex、Claude Code、OpenCode、Cursor 等 AI Agent 补上搜索短板，覆盖通用网页搜索、垂直搜索、批量搜索、全文提取、选题雷达、竞品分析和技术调研，并说明如何搭配 4SAPI 做模型调用、Key、日志和成本治理。"
---

# 【大模型API中转站】第98期 AnySearch | 给Codex装搜索大脑

本文是【大模型API中转站】系列的第98篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

不会搜索的 AI Agent，干活效率真的差一截。

你应该也遇到过这种情况。

Codex、Claude Code、OpenCode 这类智能体，写代码、改项目、生成报告都很强。

但一到：

```text
查资料
找数据
做竞品分析
写行业报告
整理热点选题
核对最新文档
```

就开始不稳。

不是模型不会分析。

而是它拿不到足够好的信息。

一个没有搜索能力的 Agent，很容易变成：

```text
靠记忆回答。
靠猜补全细节。
靠旧知识写新报告。
看起来很顺，实际没来源。
```

这时候就需要一个搜索增强层。

今天讲的 AnySearch，就是给 AI Agent 用的搜索 Skill。

官方 README 对它的定义很直接：

```text
Unified real-time search engine skill for AI agents.
```

它支持：

```text
通用网页搜索
垂直领域搜索
批量并行搜索
网页全文提取
```

你可以把它理解成：

```text
AnySearch 负责找信息。
Codex / Claude Code 负责理解和生成。
4SAPI 负责把后续模型调用、Key、日志和成本管起来。
```

三者组合起来，Agent 才不容易闭门造车。

## 1. AnySearch 到底是什么

AnySearch 不是普通搜索网站。

它是一个给 AI Agent 使用的搜索增强 Skill。

它的目标不是让人打开网页搜关键词。

而是让 Agent 在执行任务时，可以主动调用搜索能力。

比如你让 Codex 做一份报告。

没有 AnySearch 时，它可能会直接写：

```text
根据我的理解，行业趋势大概是……
```

这很危险。

因为趋势、价格、版本、新闻、文档、开源项目状态，都会变化。

有了 AnySearch 后，更合理的流程是：

```text
先搜索
再提取网页正文
再整理来源
再对比多个结果
再让模型生成报告
```

也就是：

```text
搜索负责取材。
模型负责思考。
Skill 负责把能力变成工作流。
```

这才是 Agent 真正应该干活的方式。

## 2. 它补的是 Agent 的哪块短板

现在很多人用 Codex 或 Claude Code，只让它们写代码。

其实有点浪费。

AI Agent 真正厉害的地方，是可以跑完整工作流：

```text
查资料
理解需求
生成方案
写代码
验证结果
输出报告
做可视化
```

但这里面最容易缺的一环，就是查资料。

没有搜索增强，Agent 会出现几个问题。

第一，信息过期。

比如你问：

```text
某个框架最新版本怎么配置？
某个 API 现在支持什么参数？
某个产品现在多少钱？
某个模型最近有没有更新？
```

这些问题不能只靠模型记忆。

第二，来源太薄。

Agent 随便搜两条结果，就开始写总结。

报告看起来完整，但证据不够。

第三，不会批量并行查。

做竞品分析、选题雷达、行业报告，通常不是搜一个关键词。

而是要同时查：

```text
产品官网
GitHub
文档
新闻
社交讨论
用户评价
价格页
案例文章
```

AnySearch 的批量搜索能力，正适合这种任务。

第四，不会提取全文。

只看搜索结果摘要，往往不够。

真正做报告，要读网页正文。

AnySearch 支持 full-page content extraction，可以把页面正文提取出来给 Agent 继续处理。

## 3. 它适合谁用

我觉得最适合三类人。

### 3.1 做内容选题的人

如果你每天要找选题，AnySearch 很适合做“选题雷达”。

比如：

```text
近 7 天 AI 圈有什么热点？
Codex、Claude Code、Skill 最近有什么新项目？
哪些工具在 GitHub 上涨得快？
哪些话题适合写公众号？
哪些话题适合小红书？
哪些标题角度更容易吸引开发者？
```

以前你可能要自己打开：

```text
X
Reddit
GitHub
Hacker News
公众号
新闻站
产品官网
```

一个个查。

现在可以让 Agent 先跑一轮搜索。

再整理成：

```text
事件
来源
热度
适合人群
创作角度
标题建议
风险提醒
```

这比让模型凭空列选题靠谱太多。

### 3.2 经常写报告的人

比如：

```text
行业趋势报告
竞品分析报告
产品调研报告
技术方案对比
市场数据整理
融资和政策跟踪
```

这类任务最麻烦的不是写。

而是前面的资料收集。

AnySearch 的价值，就是把“查资料”这一步变得更适合交给智能体。

Codex / Claude Code 不只是把结果写漂亮。

而是能先查，再写。

### 3.3 用 Codex / Claude Code 做项目的人

很多代码任务也需要搜索。

比如：

```text
查最新 SDK 用法
核对官方 API 参数
找迁移指南
查 GitHub issue
对比同类库
整理方案选型
写技术调研
```

没有搜索能力，Agent 很容易按照旧写法改代码。

有搜索能力，它可以先找最新资料，再回到代码里执行。

这就是搜索增强对开发 Agent 的意义。

## 4. AnySearch 和 4SAPI 怎么分工

这里要讲清楚。

AnySearch 不是模型网关。

4SAPI 也不是搜索 Skill。

它们分工不同。

```text
AnySearch：负责搜索、批量检索、页面提取。
Codex / Claude Code：负责理解上下文、改文件、生成报告。
4SAPI：负责模型入口、Key、日志、成本和多模型路由。
```

举个例子。

你要做一份 AI 热点日报：

```text
AnySearch 搜索过去 24 小时热点。
Codex 整理来源和结构。
4SAPI 提供模型调用入口。
低成本模型做初筛。
强模型做最终分析。
4SAPI 后台记录每次调用和成本。
```

这样才是完整链路。

如果只有 AnySearch，没有模型治理，问题是：

```text
搜索结果有了，但后续模型调用成本不可控。
```

如果只有 4SAPI，没有搜索增强，问题是：

```text
模型入口统一了，但取材仍然不够好。
```

两者组合起来，才适合长期工作流。

## 5. 小白怎么安装

如果你是新手，不建议一上来自己复制一堆命令。

最简单的方式，是把官方 GitHub 地址丢给 Codex 或 Claude Code，让它帮你检查环境并安装。

可以直接复制这段：

```markdown
请帮我安装 AnySearch Skill。

要求：
1. 先检查当前环境是否支持 Skill；
2. 从官方 GitHub 仓库安装 AnySearch Skill；
3. 优先安装官方 release，不要随意使用未知来源；
4. 不要覆盖我已有的重要配置；
5. 安装完成后告诉我 Skill 的安装路径；
6. 根据当前系统检测 Python、Node.js、PowerShell 或 Bash 可用性；
7. 运行一次 entry test，确认 AnySearch 是否可用；
8. 最后帮我运行一个简单搜索测试。

官方仓库：
https://github.com/anysearch-ai/anysearch-skill
```

官方 README 里给的手动思路是：

```text
下载 release zip
解压
移动到 Agent 的 skill 目录
把目录重命名为 anysearch
运行 entry test
记录推荐 runtime 到 runtime.conf
```

常见目录包括：

```text
Claude Code: ~/.claude/skills/anysearch
OpenCode: ~/.config/opencode/skills/anysearch
Cursor / Windsurf: <project>/.skills/anysearch
Generic: <your_agent_skill_dir>/anysearch
Shared agents: ~/.agents/skills/anysearch
```

其中 `~/.agents/skills/` 适合多个 Agent 共用。

官方 README 也提到，这个共享目录对 Codex、Cursor、OpenClaw 这类个人 Agent Skill 场景很有用。

重点是：

```text
不要盲装。
让 Agent 先检查你的 Skill 目录和运行时。
```

## 6. API Key 要不要配置

AnySearch 支持匿名使用。

但官方 README 也写得很清楚：

```text
API key is optional but strongly recommended.
```

没有 Key 也能用搜索功能。

但匿名访问通常会有更低的频率限制和额度。

如果你只是试一下，可以先匿名。

如果你要长期跑：

```text
选题日报
竞品分析
行业报告
内容库生成
技术调研
```

建议去 AnySearch 控制台申请 API Key。

拿到 Key 后，可以把这段发给 Codex：

```markdown
请帮我配置 AnySearch 的 API Key。

要求：
1. 不要把 API Key 写进公开代码；
2. 优先放到 .env 或本地环境变量；
3. 如果当前 Skill 目录有 .env.example，请复制为 .env 后再填写；
4. 配置 ANYSEARCH_API_KEY；
5. 不要把 .env 提交到 Git；
6. 配置完成后运行一次测试搜索；
7. 测试通过后告诉我日常应该如何调用 AnySearch。
```

官方支持几种方式：

```text
.env 文件
环境变量 ANYSEARCH_API_KEY
CLI 参数 --api_key
匿名访问
```

Key 优先级是：

```text
--api_key 参数 > .env 文件 > 环境变量 > 匿名访问
```

安全原则很简单：

```text
不要把 API Key 写进公开仓库。
不要截图发群。
不要写进教程正文。
不要让 Agent 把 .env 提交到 Git。
```

这个细节很重要。

## 7. 装好后怎么用

官方 README 里，AnySearch 的日常调用大概是这几类：

```text
search
batch_search
extract
```

比如：

```bash
python3 <skill_dir>/scripts/anysearch_cli.py search "query" --max_results 5
```

批量搜索：

```bash
python3 <skill_dir>/scripts/anysearch_cli.py batch_search --queries '[{"query":"q1","max_results":5},{"query":"q2","max_results":5}]'
```

提取网页：

```bash
python3 <skill_dir>/scripts/anysearch_cli.py extract "https://example.com/page"
```

如果你的环境里推荐 runtime 是 Node.js、PowerShell 或 Bash，就按 `runtime.conf` 里记录的命令来。

不要每次都重新跑 doc。

安装时确认 runtime，日常调用直接用推荐命令。

## 8. 最实用案例：AI 热点选题雷达

如果你想快速判断 AnySearch 是否适合自己，可以先跑这个任务。

直接发给 Codex：

```markdown
请使用 AnySearch Skill，帮我搜索近 7 天 AI Agent、Codex、Claude Code、Skill 相关热门话题。

要求：
1. 至少整理 20 个高相关话题；
2. 按热度和创作价值排序；
3. 每个话题给出：事件名称、发生时间、核心信息、来源、适合人群、创作角度、标题建议；
4. 需要尽量覆盖 GitHub、官方博客、技术媒体、社区讨论；
5. 最后生成一个 HTML 选题报告；
6. 报告要适合公众号、X、小红书内容创作者查看；
7. 不要只堆链接，要给出可执行的创作建议；
8. 标注哪些内容需要人工二次核对。
```

这个任务能测出三件事：

```text
搜索覆盖面够不够。
Agent 会不会整理来源。
最终报告有没有创作价值。
```

如果跑完只是给你一堆链接，说明提示词还不够。

如果跑完能给你事件、来源、角度、标题、适合平台，那就有用了。

## 9. 五个可以直接复制的场景

下面这些任务，特别适合 AnySearch + Codex / Claude Code。

### 9.1 AI 热点日报

```markdown
请使用 AnySearch 搜索过去 24 小时 AI 领域重要新闻，按热度和创作价值排序。

输出要求：
1. 至少 10 条；
2. 每条包含事件、来源、时间、核心信息、为什么重要；
3. 标注适合写公众号、X、小红书还是长报告；
4. 给出 3 个标题建议；
5. 最后生成一份适合内容创作者使用的日报。
```

### 9.2 竞品分析

```markdown
请使用 AnySearch 搜索 XXX 产品的官网、定价、核心功能、用户评价、GitHub 或社区讨论，以及主要竞品信息。

输出要求：
1. 整理成竞品分析报告；
2. 包含定位、目标用户、价格、功能、优势、短板；
3. 给出可切入的差异化机会；
4. 每个关键判断都要附来源。
```

### 9.3 技术调研

```markdown
请使用 AnySearch 搜索某个技术方案的官方文档、GitHub 项目、真实使用案例和常见问题。

输出要求：
1. 优先官方文档和 GitHub；
2. 对比至少 3 个方案；
3. 给出适用场景和不适用场景；
4. 最后输出一份技术选型建议。
```

### 9.4 行业数据报告

```markdown
请使用 AnySearch 搜索某个行业最近 3 个月的重要数据、政策、融资、产品变化，并生成结构化分析报告。

输出要求：
1. 数据必须标注来源；
2. 区分事实、推测和建议；
3. 给出 3 个趋势判断；
4. 最后列出需要继续人工核对的数据点。
```

### 9.5 内容选题库

```markdown
请使用 AnySearch 搜索近一周 AI 工具、AI 编程、出海 SaaS 相关热点，整理出 30 个可写选题。

每个选题包含：
1. 标题；
2. 切入角度；
3. 目标人群；
4. 参考来源；
5. 适合平台；
6. 预估传播点；
7. 风险提醒。
```

## 10. 用 4SAPI 把搜索报告变成生产工作流

AnySearch 解决的是取材。

但取材之后，还有模型调用。

比如一份竞品报告，可能要跑：

```text
搜索阶段：AnySearch
清洗阶段：低成本模型
归纳阶段：长上下文模型
分析阶段：强推理模型
写作阶段：写作模型
校对阶段：便宜模型 + 规则检查
```

如果这些模型散落在不同供应商，成本很难算。

更稳的做法，是把模型调用放到 4SAPI 里统一管理。

推荐拆 Key：

```text
4SAPI-SearchReport-Draft
用途：搜索结果清洗、初稿整理。
```

```text
4SAPI-SearchReport-Reasoning
用途：趋势判断、竞品分析、技术选型。
```

```text
4SAPI-SearchReport-Writing
用途：长文成稿、标题、摘要。
```

```text
4SAPI-SearchReport-Image
用途：封面图、配图提示词或图片生成。
```

这样你能在 4SAPI 后台看到：

```text
哪个报告花了多少钱。
哪一步模型调用最多。
哪个 Key 失败率高。
哪个 workflow 需要降成本。
```

AnySearch 让 Agent 找到资料。

4SAPI 让 Agent 的模型调用可控。

这就是从“玩工具”变成“做生产流”的区别。

## 11. 和 n8n / loop-me / OpenClaw 怎么组合

如果你已经在用 n8n、loop-me、OpenClaw，AnySearch 的位置也很清楚。

### 11.1 n8n 场景

n8n 适合定时触发：

```text
每天早上 9 点跑 AI 热点日报。
每周一跑竞品监控。
每月跑行业趋势报告。
```

流程可以是：

```text
n8n 定时触发
  -> AnySearch 搜索
  -> 4SAPI 模型归纳
  -> 生成报告
  -> 发到飞书/企微/邮箱
  -> 保存到 Notion/表格
```

### 11.2 loop-me 场景

loop-me 负责把重复任务问成 workflow spec。

比如：

```text
workflows/ai-topic-radar.md
```

里面写清楚：

```text
什么时候搜索。
搜索哪些关键词。
来源怎么排序。
什么结果要排除。
用哪个 4SAPI 模型总结。
报告发给谁。
预算是多少。
```

### 11.3 OpenClaw 场景

OpenClaw 适合长期 Agent 工作区。

你可以把 AnySearch 当搜索层。

把 4SAPI 当模型层。

把 OpenClaw 当执行和记忆层。

比如：

```text
OpenClaw 读取工作区目标。
AnySearch 找资料。
4SAPI 模型分析。
OpenClaw 生成任务、报告、内容草稿。
```

这比单独问聊天框强很多。

## 12. 常见错误

第一，装完不测试。

官方 README 明确建议做 post-install verification。

至少要跑：

```text
doc
search "hello world" --max_results 1
```

第二，不记录 runtime。

AnySearch 提供 Python、Node.js、PowerShell、Bash 多套 CLI。

安装时要确认哪个最适合当前环境，并写进 `runtime.conf`。

第三，把 API Key 写进公开仓库。

`.env` 要进 `.gitignore`。

不要把 Key 写进 README。

第四，搜索结果不做来源标注。

报告必须带来源。

尤其是价格、数据、新闻、版本、政策。

第五，搜索结果直接发布。

AnySearch 是搜索工具，不是事实保证器。

关键事实还要人工核对。

第六，把搜索和模型成本混在一起。

AnySearch 负责搜索。

4SAPI 负责模型调用治理。

两边都要看日志。

第七，跑批量搜索不设上限。

选题雷达、竞品分析、行业报告都容易越跑越大。

要限制：

```text
关键词数量
单关键词结果数
页面提取数量
模型总结次数
单次 workflow 成本
```

第八，给 `extract` 乱加格式参数。

官方 README 提醒，`extract` 输出已经是 Markdown。

不要再传：

```text
--format markdown
--format json
--markdown
```

如果子命令参数不确定，跑对应子命令的 `--help`，不要靠猜。

## 13. 最小检查清单

安装前：

```text
[ ] 已确认 Agent 支持 Skill
[ ] 已确认 Skill 安装目录
[ ] 优先使用 AnySearch 官方 release
[ ] 不覆盖已有配置
[ ] 已检查 Python / Node.js / PowerShell / Bash
```

安装后：

```text
[ ] 已运行 entry test
[ ] 已确认 runtime.conf
[ ] 已配置 ANYSEARCH_API_KEY 或确认匿名使用
[ ] 已跑 search 测试
[ ] 已跑 extract 测试
[ ] 已确认 .env 不会提交到 Git
```

做生产工作流前：

```text
[ ] 已定义搜索关键词和来源范围
[ ] 已限制 max_results
[ ] 已标注来源
[ ] 已用 4SAPI 管后续模型调用
[ ] 已拆分草稿、分析、写作、图片等 Key
[ ] 已设置单次 workflow 预算
[ ] 已保留人工核对点
```

## 14. 最后建议

如果你只是偶尔问 AI 几个问题，可以先不用折腾。

但如果你经常做这些事：

```text
每天找选题
查资料写文章
做竞品分析
做行业报告
让 Codex / Claude Code 跑复杂任务
想把 Agent 变成真正的工作流助手
```

那 AnySearch 值得试。

因为 Agent 的能力不只取决于模型多强。

还取决于它能不能拿到足够好的信息。

一个成熟的 AI 工作流应该是：

```text
AnySearch 负责找资料。
Codex / Claude Code 负责执行。
4SAPI 负责模型入口、日志、成本和权限。
```

一句话：

```text
模型负责思考。
搜索负责取材。
Skill 负责把能力变成工作流。
4SAPI 负责让这条工作流可控、可查、可复盘。
```

这才是 AI Agent 真正好用的关键。

## 资料来源与延伸阅读

- AnySearch Skill GitHub：https://github.com/anysearch-ai/anysearch-skill
- AnySearch Skill README：https://github.com/anysearch-ai/anysearch-skill/blob/main/README.md
- AnySearch API Key 控制台：https://anysearch.com/console/api-keys
- AnySearch 官网：https://www.anysearch.com/
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 文档：https://4sapi.apifox.cn/
- 4SAPI n8n 配置教程：https://4sapi.apifox.cn/8328708m0
- 4SAPI OpenCode 接入配置：https://4sapi.apifox.cn/8323429m0
