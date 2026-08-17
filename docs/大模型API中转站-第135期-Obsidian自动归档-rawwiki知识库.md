---
title: "【大模型API中转站】第135期 Obsidian自动归档 | raw/wiki知识库"
category: 人工智能
tags:
  - 大模型API中转站
  - Obsidian
  - Claude
  - Claude Code
  - 第二大脑
  - raw
  - wiki
  - 自动归档
  - 知识库
  - 4SAPI
description: "把 Obsidian Vault 拆成 raw 原始材料区和 wiki 编译知识区，让 Claude 负责读取、归档、链接、查重和生成维护日志。文章给出目录结构、处理流程、提示词模板、每日维护任务和适合团队使用的审计方式。"
---

# 【大模型API中转站】第135期 Obsidian自动归档 | raw/wiki知识库

本文是【大模型API中转站】系列的第135篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一期讲了怎么把 Claude 通过 MCP 接进 Obsidian。

但接通只是第一步。

接下来真正有价值的问题是：

```text
新材料来了以后，Claude 到底应该怎么整理？
```

大多数人的知识库失败，不是因为不会装插件。

而是因为资料入口太多：

```text
浏览器标签页。
微信群截图。
PDF。
会议纪要。
Claude 对话。
Notion 旧页面。
飞书文档。
本地 Markdown。
```

这些东西都进来了，但没有固定流程。

结果就是：

```text
今天新建一个文件夹。
明天换一套标签。
后天让 AI 整理一遍。
一周后自己也找不到了。
```

这篇讲一个更稳的结构：

```text
raw/ 放原始材料。
wiki/ 放 Claude 编译出来的知识页。
```

你可以把它理解成：

```text
raw 是证据库。
wiki 是说明书。
```

原始材料尽量不改。

Claude 只负责读、提炼、链接、更新 wiki。

这样你的知识库不会被 AI 改成一团“看起来很聪明、但找不到出处”的笔记泥巴。

## 1. raw/wiki 的核心思想

先把目录摆出来：

```text
brain/
  CLAUDE.md
  inbox/
  raw/
  wiki/
  projects/
  people/
  decisions/
  logs/
```

`raw/` 放不可轻易改动的原始材料：

```text
文章全文。
PDF 摘录。
会议录音转写。
网页剪藏。
截图说明。
Claude 对话导出。
客户原始需求。
竞品资料。
```

`wiki/` 放经过 Claude 整理后的知识页面：

```text
主题页。
概念页。
项目背景页。
人物资料页。
决策总结页。
流程说明页。
```

它们之间的关系是：

```text
raw/ 负责保真。
wiki/ 负责可用。
```

AI 最容易犯的错，是把这两件事混在一起。

它一边改原文，一边总结，一边补链接，最后你分不清：

```text
哪些是原始事实。
哪些是模型推断。
哪些是你后来的判断。
```

所以第一条规则必须写死：

```text
raw/ 默认只读。
wiki/ 才是 Claude 可以更新的工作区。
```

## 2. 为什么这套结构适合 Claude

Claude 很适合做三类工作。

第一类是阅读长材料。

比如你丢进去一篇 2 万字文章，它可以提炼：

```text
核心观点。
关键证据。
适用场景。
反对意见。
和现有笔记的关系。
```

第二类是建立链接。

比如一篇新材料同时关联：

```text
Claude Code
Obsidian
MCP
第二大脑
企业知识库
```

Claude 可以把这些主题页串起来。

第三类是维护一致性。

例如它可以检查：

```text
wiki/ 里有没有重复页面。
某个主题页有没有过时说法。
链接是否指向不存在的文件。
同一个客户名字有没有多种写法。
```

这才是 AI 整理知识库最该做的事。

不是把每篇文章都写成“摘要、启发、行动项”三段式。

而是帮你维护一张长期可用的知识网络。

## 3. 第一步：给 raw/ 建原始材料分类

`raw/` 里不要一股脑堆文件。

建议先按来源分。

例如：

```text
raw/
  articles/
  transcripts/
  meetings/
  screenshots/
  pdfs/
  conversations/
  customer-input/
```

每个原始文件的命名尽量固定：

```text
YYYY-MM-DD-来源-主题.md
```

例如：

```text
2026-07-02-linuxdo-claude-obsidian-mcp.md
2026-07-02-meeting-api-gateway-budget.md
2026-07-02-claude-chat-fable5-loop-notes.md
```

为什么要日期？

因为知识库会变。

同一个主题，今天的判断和三个月后的判断可能完全不同。

日期能让 Claude 判断：

```text
这是不是旧材料？
这条结论是不是已经被后来的资料覆盖？
同一主题的发展顺序是什么？
```

## 4. 第二步：给 wiki/ 建主题页

`wiki/` 不是材料仓库。

它更像一本内部百科。

建议先建这些入口页：

```text
wiki/
  _index.md
  claude-code.md
  obsidian.md
  mcp.md
  second-brain.md
  api-gateway.md
  content-workflow.md
```

每个主题页都用同一套结构。

例如：

```markdown
# Claude Code

## 一句话解释

Claude Code 是……

## 适用场景

- 场景 1
- 场景 2

## 关键概念

- [[MCP]]
- [[CLAUDE.md]]
- [[Obsidian]]

## 已知流程

- 流程 A
- 流程 B

## 相关原始材料

- [[raw/articles/2026-07-02-xxx]]

## 待核实

- 这里放还没有官方来源或需要复查的说法。
```

注意最后一栏：

```text
待核实
```

这很重要。

AI 整理知识库最怕“语气很确定”。

如果某条说法来自社区经验、网传案例、你自己的猜测，就放进待核实。

不要写进结论区。

## 5. 第三步：给 Claude 一条编译指令

接下来要让 Claude 明白：

```text
它不是在改写原文。
它是在把 raw 编译成 wiki。
```

可以把这段写进 `CLAUDE.md`：

```text
## raw/wiki 工作流

- raw/ 是原始材料区，默认只读。
- wiki/ 是知识编译区，可以更新。
- 处理 raw/ 新材料时，不要改原文。
- 先读材料，再判断它关联哪些 wiki 主题页。
- 如果主题页存在，就补充新证据、新链接或待核实条目。
- 如果主题页不存在，先新建一个短页，不要写百科长文。
- 每次更新 wiki，都要在页面末尾记录来源文件。
- 不确定的内容放入“待核实”，不要写成事实。
- 完成后写维护日志到 logs/。
```

然后给 Claude 一个具体任务：

```text
请扫描 raw/ 中最近 7 天新增的材料。

要求：
1. raw/ 只读，不要修改。
2. 判断每条材料关联哪些 wiki 主题。
3. 更新已有 wiki 页，或创建必要的新主题页。
4. 每条新增观点都要标注来源文件。
5. 不确定、网传、未核实内容放入“待核实”。
6. 最后写 logs/YYYY-MM-DD-wiki-compile.md，列出本次更新了哪些页面。

开始前先给计划，等我确认。
```

这条提示词比“帮我整理知识库”稳定得多。

因为它定义了：

```text
输入在哪里。
输出在哪里。
什么不能改。
什么算完成。
怎么留下记录。
```

## 6. 第四步：让 inbox/ 做缓冲区

很多人会问：

```text
那 inbox/ 有什么用？
```

`inbox/` 是临时入口。

你随手丢进去的东西，还没判断是否值得长期保存。

例如：

```text
一段灵感。
一条链接。
一张截图。
一段语音转文字。
一个临时任务。
```

每天或每周，让 Claude 清一次 inbox：

```text
请处理 inbox/。

规则：
1. 先列出每条 inbox 笔记。
2. 判断它应该进入 raw/、projects/、people/、decisions/，还是应该保留在 inbox/。
3. 不要删除。
4. 移动前先给我确认。
5. 移动后更新相关 wiki 链接。
```

这样 `inbox/` 不会变成垃圾桶。

它只是入口。

## 7. 第五步：把“每日 7 点整理”拆成三个小任务

素材里提到每天早上 7 点自动整理知识库。

这个想法很好，但第一版不要做成一个巨型任务。

建议拆成三个小任务：

```text
任务 A：处理 inbox。
任务 B：编译 raw 到 wiki。
任务 C：生成晨间摘要。
```

任务 A：

```text
扫描 inbox/。
列出新增项。
提出归档建议。
等待确认，不自动删除。
```

任务 B：

```text
读取 raw/ 最近新增材料。
更新 wiki/ 主题页。
给每个新增观点标来源。
把未核实内容放进待核实。
```

任务 C：

```text
读取 logs/ 最近 24 小时维护记录。
输出三行摘要：
1. 新增了什么。
2. 哪些主题发生变化。
3. 哪些事项需要我确认。
```

如果你的工具支持定时任务，可以把 A 和 C 自动化。

任务 B 我建议先半自动。

因为它会修改知识结构。

前两周最好让 Claude 先列计划，你确认后再执行。

## 8. 第六步：每次维护都写 log

知识库最怕“被整理过，但不知道怎么整理的”。

所以每次 Claude 工作完，都要写日志。

例如：

```text
logs/
  2026-07-02-wiki-compile.md
  2026-07-03-inbox-cleanup.md
  2026-07-04-link-check.md
```

日志模板：

```markdown
# 2026-07-02 wiki compile

## 输入

- raw/articles/xxx.md
- raw/conversations/yyy.md

## 更新页面

- wiki/claude-code.md
- wiki/obsidian.md

## 新增链接

- [[Claude Code]] -> [[MCP]]
- [[Obsidian]] -> [[Local REST API]]

## 待确认

- 某个社区说法缺少官方来源。

## 未执行

- 没有移动 raw/ 文件。
- 没有删除任何文件。
```

这份 log 有两个作用。

第一，你能回看。

第二，下一次 Claude 可以读 log，避免重复做同一件事。

## 9. 第七步：做链接检查

Obsidian 知识库越大，坏链接越多。

坏链接包括：

```text
指向不存在页面。
同一个主题有两个名字。
大小写不一致。
中文名和英文名混用。
```

给 Claude 一个专门任务：

```text
请检查 wiki/ 中的 Obsidian 链接。

目标：
1. 找出指向不存在页面的 [[链接]]。
2. 找出疑似重复主题页。
3. 给出合并建议。
4. 不要直接改名，不要移动文件。
5. 输出报告到 logs/YYYY-MM-DD-link-check.md。
```

注意：

```text
改名和合并一定要人工确认。
```

因为链接结构是知识库骨架。

AI 可以建议，但不应该直接大规模重命名。

## 10. 第八步：做“过时内容”检查

知识库不是越多越好。

很多笔记会过时。

比如：

```text
旧模型价格。
旧 API 参数。
旧安装命令。
旧产品入口。
旧政策说明。
```

这类内容如果不标注，会害人。

可以给 Claude 一个每周任务：

```text
请检查 wiki/ 中可能过时的内容。

重点：
- 价格
- 版本号
- 安装命令
- API 参数
- 官方入口
- 产品可用性

要求：
1. 不要直接修改。
2. 列出疑似过时段落。
3. 标出来源文件和所在页面。
4. 给出需要重新核实的链接或搜索关键词。
5. 输出到 logs/YYYY-MM-DD-stale-check.md。
```

如果你写的是技术教程，这个任务非常有价值。

因为技术文章最容易坏在：

```text
命令过期。
入口换了。
模型名变了。
价格改了。
```

## 11. 第九步：用 4SAPI 控制批处理成本

raw/wiki 工作流很适合批处理。

但批处理也容易烧钱。

如果你让高价模型每天扫完整个 Vault，成本会很快失控。

更合理的方式是分层：

```text
便宜模型：
做初筛、分类、重复检测、链接候选。

强模型：
做最终主题页重写、冲突判断、重要决策摘要。

人工：
确认删除、合并、改名、公开发布。
```

这就是 4SAPI 或企业级 API 网关适合介入的地方：

```text
按任务路由模型。
给 raw/wiki 批处理单独设预算。
记录每次处理了哪些文件。
失败时保留日志。
敏感目录默认不送模型。
```

可以在 `CLAUDE.md` 里写：

```text
## 模型使用原则

- 扫描和分类优先使用低成本模型。
- 最终发布稿、复杂冲突判断、跨主题综合，才使用强模型。
- 每次批处理前估算文件数量和范围。
- 不处理 secrets/、private/、finance/，除非我明确指定。
```

提示词只是提醒。

真正的预算和权限，应该在 API 网关、MCP server 和文件系统层面一起控制。

## 12. 第十步：给团队用时怎么改

个人知识库可以随意一点。

团队知识库必须更严格。

团队版建议加这些目录：

```text
brain/
  CLAUDE.md
  raw/
  wiki/
  projects/
  people/
  decisions/
  logs/
  reviews/
  publish/
```

多两个目录：

```text
reviews/
publish/
```

`reviews/` 放待人工审核的 AI 输出。

`publish/` 放已经确认可以外发或复用的内容。

团队规则：

```text
Claude 可以写 wiki。
Claude 可以写 reviews。
Claude 不可以直接写 publish。
```

这条很实用。

因为团队里真正有风险的不是 AI 写草稿。

风险是 AI 草稿没审核就被当成正式知识。

## 13. 一套完整提示词

下面是一套可以直接放进 `CLAUDE.md` 的 raw/wiki 规则。

```text
# raw/wiki 知识库维护规则

你正在维护一个 Obsidian Vault。

目录含义：
- inbox/：临时输入，等待归档。
- raw/：原始材料，只读，不能修改。
- wiki/：主题页和概念页，可以更新。
- projects/：当前项目资料。
- people/：人物、客户、合作方资料。
- decisions/：决策记录。
- logs/：每次维护记录。

工作原则：
1. raw/ 默认只读。
2. 新观点必须能追溯到来源文件。
3. 未核实内容放入“待核实”。
4. 不删除文件。
5. 不大规模改名。
6. 修改多个文件前先列计划。
7. 完成后写日志。

每日维护任务：
1. 清点 inbox/。
2. 读取 raw/ 最近新增材料。
3. 更新 wiki/ 相关主题页。
4. 检查坏链接和重复主题。
5. 写 logs/YYYY-MM-DD-maintenance.md。
```

如果你只复制一段，就复制这段。

它比任何插件都重要。

## 14. 最后总结

Obsidian 第二大脑的关键，不是装多少插件。

也不是让 Claude 一次性扫完整个 Vault。

关键是分清：

```text
raw 保存事实。
wiki 组织理解。
logs 记录过程。
CLAUDE.md 约束行为。
```

有了这四层，Claude 才能稳定地帮你维护知识库。

否则，它只是一个很聪明的临时整理员。

今天帮你整理一遍，明天又把规则忘了。

raw/wiki 的价值，就是把“整理一次”变成“持续编译”。

你的知识库不再是一堆笔记。

它会慢慢长成一套能被模型和人一起使用的工作系统。

## 资料来源与延伸阅读

- Obsidian 官方网站：https://obsidian.md/
- Obsidian 数据存储说明：https://obsidian.md/help/data-storage
- Claude Code Memory / CLAUDE.md 文档：https://docs.anthropic.com/en/docs/claude-code/memory
- Claude Code MCP 文档：https://docs.anthropic.com/en/docs/claude-code/mcp
- Obsidian Local REST API：https://github.com/coddingtonbear/obsidian-local-rest-api

