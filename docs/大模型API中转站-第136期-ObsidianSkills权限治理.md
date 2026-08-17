---
title: "【大模型API中转站】第136期 Obsidian Skills | 权限治理"
category: 人工智能
tags:
  - 大模型API中转站
  - Obsidian
  - Claude Code
  - Skills
  - Agent Skills
  - 第二大脑
  - 权限治理
  - 知识库
  - 企业级大模型接入
  - 4SAPI
description: "接上 Obsidian 之后，下一步不是让 Claude 随便改库，而是给它装 Skills、划权限、设只读边界和人工审核流。本文讲 kepano/obsidian-skills、claude-obsidian、obsidian-second-brain、second-brain-starter 这几类方案怎么选，以及如何避免 AI 把第二大脑整理成事故现场。"
---

# 【大模型API中转站】第136期 Obsidian Skills | 权限治理

本文是【大模型API中转站】系列的第136篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

前两篇讲了两件事。

第134期：

```text
怎么用 MCP 把 Claude 接进 Obsidian。
```

第135期：

```text
怎么用 raw/wiki 结构，让 Claude 把原始材料编译成知识页。
```

这一篇讲第三层：

```text
Skills 和权限治理。
```

很多人把 Claude 接上 Obsidian 后，第一反应是：

```text
太好了，让它全自动整理我的整个库。
```

这句话听起来很爽。

但如果没有规则，它也很危险。

因为 Obsidian Vault 不是测试项目。

里面可能有：

```text
客户资料。
账号信息。
会议纪要。
合同摘要。
个人日记。
商业计划。
未发布文章。
旧项目决策。
```

如果你给 Agent 全写权限，再只靠一句提示词：

```text
不要乱删。
```

那迟早会出事。

正确做法是：

```text
Skills 教它 Obsidian 的写法。
权限限制它能碰什么。
日志记录它做了什么。
人工审核决定什么能发布。
```

这一篇就把这套讲清楚。

## 1. Skills 解决什么问题

Claude 会写 Markdown。

但 Obsidian 不是普通 Markdown 文件夹。

Obsidian 有自己的习惯：

```text
[[wikilinks]]
frontmatter
tags
callouts
Canvas
Bases
附件路径
嵌入语法
别名
图谱关系
```

如果你不教它这些，它也能写。

只是写出来经常会有几个问题：

```text
链接格式不对。
frontmatter 字段混乱。
同一个标签写出三种名字。
Canvas 文件结构错误。
Bases 配置看着像 YAML，但 Obsidian 不认。
```

Skills 的作用就是把这些“本地规则”封装起来。

一句话：

```text
你教一次，它以后遇到相关任务会自己调用。
```

这比每次在提示词里重新解释：

```text
请使用 Obsidian wikilink。
请不要破坏 frontmatter。
请按我的标签规范。
```

稳定得多。

## 2. kepano/obsidian-skills：先装这个做基础语法层

如果你只想先装一组 Obsidian Skills，我建议先看：

```text
kepano/obsidian-skills
```

它的定位很清楚：

```text
Agent skills for Obsidian.
```

也就是教 Agent 使用 Obsidian 的开放格式。

它覆盖的方向包括：

```text
Obsidian Markdown
Bases
JSON Canvas
CLI / 网页提取等 Obsidian 相关能力
```

更重要的是，它遵循 Agent Skills 规范。

这意味着它不是只给某一个模型用。

支持 Skills 的 Agent，都有机会复用这套规则，例如 Claude Code、Codex、OpenCode 等。

这点对长期知识库很重要。

因为你的 Vault 不应该被某一个模型锁死。

推荐用法：

```text
先把 kepano/obsidian-skills 当作 Obsidian 语法层。
再在你自己的 CLAUDE.md 里写业务规则。
```

不要反过来。

不要把业务偏好全塞进通用 Obsidian Skills。

通用 Skills 负责：

```text
怎么写 Obsidian。
```

你的 `CLAUDE.md` 负责：

```text
你的 Vault 怎么工作。
```

## 3. 三类现成仓库怎么选

你粘贴的素材里提到几个社区仓库。

我建议按用途分，不要混在一起。

### 方案 A：claude-obsidian

适合想要“自组织第二大脑”的人。

这个方向的特点是：

```text
你丢材料进去。
它读取、提取实体、建立链接、维护 wiki。
```

它更接近第135期讲的 raw/wiki 自动编译模式。

适合：

```text
研究型用户。
内容创作者。
长期资料库。
希望 AI 主动维护知识图谱的人。
```

不适合：

```text
只想轻量写日记。
不愿意读配置。
不想让 Agent 批量改文件。
```

### 方案 B：obsidian-second-brain

适合想要跨工具使用的人。

它的定位是 Cross-CLI skill。

也就是说，不只面向 Claude Code，也希望在 Codex、Gemini、OpenCode 等环境里使用。

这类方案的价值是：

```text
同一套 Obsidian 工作流，不被一个 CLI 绑定。
```

适合：

```text
同时用 Claude Code 和 Codex 的人。
团队里工具不统一的人。
想把命令化工作流沉淀下来的人。
```

要注意的是：

```text
命令越多，学习成本越高。
```

不要一口气把所有命令都用上。

先选 3 个高频动作：

```text
保存笔记。
整理 inbox。
搜索 Vault。
```

跑顺以后再扩展。

### 方案 C：second-brain-starter

适合还没想清楚系统的人。

它更像一个“生成搭建方案”的起点。

不是一上来就给你一个庞大成品。

而是先通过需求模板或 PRD，把你的工具、工作流、记忆、集成和安全边界想清楚。

适合：

```text
团队准备正式搭知识系统。
个人工作流还没定型。
想先写方案再动手的人。
```

不适合：

```text
只想今晚就把 Obsidian 接上 Claude。
```

这类用户先看第134期更合适。

## 4. 不要“克隆仓库就完事”

社区仓库很诱人。

因为它们通常已经打包好了：

```text
目录结构。
Skills。
命令。
自动维护流程。
搜索脚本。
示例配置。
```

但你的知识库不是示例项目。

克隆之前先问四个问题：

```text
1. 它会读哪些目录？
2. 它会写哪些目录？
3. 它会不会删除、移动、重命名文件？
4. 它会不会调用外部 API 或联网搜索？
```

只要其中任意一个问题你答不上来，就不要直接指向真实 Vault。

正确流程是：

```text
先建测试 Vault。
复制 10 篇无敏感笔记进去。
跑一遍安装和命令。
看 git diff。
确认行为后，再迁移到真实库。
```

这一步不能省。

AI 第二大脑的第一原则不是自动化。

而是可回滚。

## 5. 权限分层：别让提示词扮演锁

很多教程会写：

```text
告诉 Claude 不要删除文件。
```

这当然要写。

但它不够。

提示词是提醒。

权限才是边界。

一个稳一点的权限分层是：

```text
raw/：只读。
wiki/：可写。
reviews/：可写。
publish/：只读或人工写。
private/：默认不可访问。
secrets/：绝不接入模型。
```

对应到实际执行，可以这样设计：

```text
日常整理任务：
只能读 raw/、写 wiki/ 和 logs/。

发布前审校任务：
只能读 wiki/、写 reviews/。

正式发布任务：
由人把 reviews/ 移到 publish/。
```

如果使用 MCP 或 API key，优先用只读 key。

如果某个任务必须写入，就把写入范围限定到特定目录。

这比一句：

```text
你不要乱动别的文件。
```

可靠得多。

## 6. 给 Vault 加一个权限版 CLAUDE.md

下面是一版可以直接放进 Vault 根目录的 `CLAUDE.md`。

```markdown
# CLAUDE.md

## 角色

你是我的 Obsidian Vault 维护助手。
你的任务是整理、链接、归档和生成报告，不是替我删除历史。

## 目录权限

- inbox/: 可以读取，可以提出移动建议，移动前必须确认。
- raw/: 只读。不得修改、删除、重命名。
- wiki/: 可以更新主题页和概念页。
- logs/: 可以写维护日志。
- reviews/: 可以写待审核草稿。
- publish/: 只读。不得写入。
- private/: 默认不要读取，除非我明确指定。
- secrets/: 禁止读取。

## 操作规则

- 修改多个文件前，先列计划。
- 不删除文件。
- 不重命名文件。
- 不批量移动文件，除非我确认。
- 每次更新必须写 logs/YYYY-MM-DD-任务名.md。
- 新增事实必须标注来源文件。
- 不确定内容放进“待核实”。

## Obsidian 格式

- 优先使用 [[wikilinks]]。
- 保持 frontmatter 字段一致。
- 不创建无意义标签。
- 不破坏 Canvas、Bases 或附件路径。

## 输出格式

- 先给结论。
- 再给变更列表。
- 最后给需要我确认的事项。
```

这份文件不是为了让 Claude 更聪明。

它是为了让 Claude 更克制。

真正好用的 Agent，不是“什么都敢做”。

而是知道什么不能碰。

## 7. Skills 和 CLAUDE.md 怎么分工

很多人会混淆：

```text
规则到底写进 Skill，还是写进 CLAUDE.md？
```

简单判断：

```text
跨项目复用的能力，写成 Skill。
只属于这个 Vault 的规则，写进 CLAUDE.md。
```

适合 Skill 的内容：

```text
Obsidian Markdown 写法。
frontmatter 规范。
Canvas JSON 结构。
Bases 使用方式。
网页转 Markdown。
通用读书笔记模板。
通用会议纪要模板。
```

适合 `CLAUDE.md` 的内容：

```text
这个 Vault 的目录意义。
哪些目录只读。
你的项目背景。
你的输出偏好。
你的禁止事项。
你的发布流程。
```

适合单次提示词的内容：

```text
今天处理哪几个文件。
这次要输出什么格式。
这次是否允许写入。
这次的截止时间。
```

三层分清以后，系统会稳定很多。

## 8. 给现成仓库做安全试运行

如果你想试 `claude-obsidian`、`obsidian-second-brain` 或其他第二大脑仓库，建议按这个流程：

第一步，建测试库：

```text
obsidian-ai-test/
  CLAUDE.md
  raw/
  wiki/
  logs/
```

第二步，放入 10 篇无敏感材料：

```text
3 篇文章
3 条会议纪要
2 条项目笔记
2 条随手想法
```

第三步，初始化 Git：

```bash
git init
git add .
git commit -m "initial test vault"
```

第四步，运行仓库的最小命令。

不要先跑全自动维护。

先跑：

```text
列文件。
读文件。
生成一篇 wiki。
写一条 log。
```

第五步，看 diff：

```bash
git diff
```

重点看：

```text
是否改了 raw。
是否新增了奇怪目录。
是否删除或重命名。
是否把来源丢了。
是否写入了外部 API key。
```

第六步，满意后再迁移。

这个流程看起来麻烦。

但比修复一个被 AI 大规模改乱的 Vault 省时间。

## 9. 企业场景：把 reviews/ 当防火墙

个人知识库可以边用边改。

企业知识库不行。

企业知识库里至少要分三层：

```text
raw/：原始资料。
wiki/：内部知识页。
reviews/：AI 生成的待审核内容。
publish/：确认可发布或可复用内容。
```

不要让 AI 直接写 `publish/`。

它可以写：

```text
reviews/2026-07-02-api-gateway-faq-draft.md
```

然后由人审核后移到：

```text
publish/api-gateway-faq.md
```

这条规则可以写进 `CLAUDE.md`：

```text
任何面向客户、官网、公众号、销售材料、合同说明的内容，只能写入 reviews/。
不得直接写入 publish/。
```

配合 Git，就能形成完整链路：

```text
AI 起草
-> reviews/
-> 人工 diff
-> 修改确认
-> publish/
-> 提交记录
```

这才是企业级大模型接入应该有的基本治理。

## 10. 和 4SAPI 怎么结合

Obsidian Skills 管的是：

```text
Agent 怎么写、怎么整理、怎么读 Vault。
```

4SAPI 或企业 API 网关管的是：

```text
模型怎么调用、成本怎么记录、权限怎么审计。
```

两者不是替代关系。

一个合理架构是：

```text
Obsidian Vault
-> Claude Code / Codex
-> Obsidian Skills / CLAUDE.md
-> MCP / 文件工具
-> 4SAPI / 企业 API 网关
-> Claude / GPT / Gemini
```

你可以在 4SAPI 层做这些事：

```text
为知识库维护任务单独建 Key。
限制每日报销预算。
记录每次调用属于哪个 Vault。
区分整理、审校、发布三类任务。
对敏感目录的调用做告警。
为强模型和便宜模型设置路由。
```

这样一来，团队就能回答几个关键问题：

```text
这次 AI 整理用了多少钱？
谁触发的？
处理了哪些文件？
用了哪个模型？
有没有失败重试？
有没有越权访问？
```

没有这些日志，AI 第二大脑越自动，风险越大。

## 11. 三种用户的推荐组合

如果你是个人新手：

```text
Obsidian + Claude Code + CLAUDE.md
先不用复杂 Skills。
先跑第134期的十步搭建。
```

如果你是重度知识工作者：

```text
Obsidian + MCP + kepano/obsidian-skills + raw/wiki 工作流
再试 claude-obsidian 或 obsidian-second-brain。
```

如果你是团队：

```text
Obsidian Vault + Git + reviews/publish 分层
Claude Code / Codex + Skills
4SAPI / 企业 API 网关
MCP 只读优先
人工审核发布
```

别一步到位。

第二大脑最怕刚开始就像企业 ERP。

先跑通一个小闭环：

```text
收集
-> 归档
-> 链接
-> 审核
-> 日志
```

闭环稳定了，再加自动化。

## 12. 最小检查清单

装任何 Obsidian Skill 或第二大脑仓库前，先过一遍：

```text
[ ] 我知道它会读哪些目录。
[ ] 我知道它会写哪些目录。
[ ] 我知道它是否会联网。
[ ] 我知道它是否会删除、移动、重命名文件。
[ ] 我已经在测试 Vault 跑过。
[ ] 我已经看过 git diff。
[ ] raw/ 默认只读。
[ ] publish/ 不允许 AI 直接写。
[ ] secrets/ 不接入模型。
[ ] 每次维护都会写 logs/。
```

如果这十条没过，不要接真实 Vault。

## 总结

Claude 接上 Obsidian 以后，真正的分水岭不是“能不能读文件”。

而是：

```text
它懂不懂 Obsidian 的格式。
它知不知道你的目录边界。
它有没有权限限制。
它做过什么能不能追踪。
它写出来的东西有没有人工审核。
```

Skills 让 Agent 会写。

`CLAUDE.md` 让 Agent 知道规矩。

MCP 让 Agent 能连接工具。

Git 和 logs 让 Agent 的行为可回滚。

4SAPI 和企业 API 网关让模型调用可审计、可控成本。

这几层合在一起，Obsidian 才不是一个被 AI 搅动的文件夹。

它会变成一个长期能用、能迁移、能审计的知识系统。

## 资料来源与延伸阅读

- kepano/obsidian-skills：https://github.com/kepano/obsidian-skills
- AgriciDaniel/claude-obsidian：https://github.com/AgriciDaniel/claude-obsidian
- eugeniughelbur/obsidian-second-brain：https://github.com/eugeniughelbur/obsidian-second-brain
- coleam00/second-brain-starter：https://github.com/coleam00/second-brain-starter
- Claude Code Skills 文档：https://docs.anthropic.com/en/docs/claude-code/skills
- Claude Code MCP 文档：https://docs.anthropic.com/en/docs/claude-code/mcp

