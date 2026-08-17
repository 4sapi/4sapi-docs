---
title: "【大模型API中转站】第134期 Claude接Obsidian | MCP十步搭建"
category: 人工智能
tags:
  - 大模型API中转站
  - Claude
  - Claude Code
  - Obsidian
  - MCP
  - 第二大脑
  - Local REST API
  - 企业级大模型接入
  - 4SAPI
description: "基于 Obsidian 本地 Markdown Vault、Local REST API、Claude Code MCP 和 CLAUDE.md，写一套给新手也能照着做的十步搭建法。重点不是再造一个聊天窗口，而是让 Claude 每次进入你的知识库时都知道你是谁、项目在哪里、材料该怎么归档。"
---

# 【大模型API中转站】第134期 Claude接Obsidian | MCP十步搭建

本文是【大模型API中转站】系列的第134篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

先说明一下和前文的关系。

第65期已经讲过：

```text
Obsidian + Claude Code 如何搭第二大脑。
```

第80期讲过：

```text
不装 Obsidian 插件，直接把 Vault 当成本地 Markdown 项目目录交给 Codex。
```

这一篇不重复讲“为什么 Obsidian 适合做第二大脑”。

这一篇只讲一件更具体的事：

```text
怎么把 Claude 通过 MCP 接进你的 Obsidian Vault。
```

目标很简单：

```text
以后你不用每次开新对话都重新解释自己。
Claude 一进来就知道：
你是谁，
你在做什么，
你的项目资料在哪，
新材料该怎么归档，
哪些文件只能读不能乱改。
```

这不是把 Claude 变成另一个聊天框。

而是把 Claude 放到你的知识库旁边，让它像一个会整理文件、会读背景、会写交接记录的助理。

## 1. 先搞清楚两件工具各管什么

Obsidian 是仓库。

它的本质不是神奇软件，而是一个本地文件夹。

里面的笔记大多是 Markdown 文件。

笔记之间可以用：

```text
[[双向链接]]
```

互相连接。

这些链接会慢慢形成你的知识图谱。

Claude 是仓库旁边的大脑。

它不应该替代 Obsidian。

它应该做这些事：

```text
读你的长期背景。
把新资料放进正确目录。
把散乱素材整理成可复用笔记。
检查坏链接和重复笔记。
基于整个 Vault 回答问题。
把你今天答应别人的事写成任务。
```

整个系统的关键不是“Claude 记住你”。

关键是：

```text
你的记忆存在本地 Markdown 里。
Claude 只是被授权来读和整理。
```

这样明年你换模型，知识库还在。

换 Claude、Codex、Gemini、Cursor，都只是把新工具指向同一个文件夹。

## 2. 第一步：安装 Obsidian，建一个 Vault

先去 Obsidian 官网下载安装：

```text
https://obsidian.md/
```

新建一个库，名字可以很朴素：

```text
brain
```

路径建议放在你自己能找到、能备份的位置，例如：

```text
~/Documents/brain
C:\Users\你的用户名\Documents\brain
```

不要一上来就放进各种同步盘深层目录。

第一天先简单。

建好以后，新建一条笔记：

```markdown
# Hello Brain

这是我的第二大脑起点。

我的年度目标会写在 [[goals]]。
```

点一下 `[[goals]]`，你会发现 Obsidian 可以顺着链接新建一页。

这就是后面 Claude 要帮你大规模做的事：

```text
发现主题。
创建页面。
建立链接。
维护图谱。
```

## 3. 第二步：准备 Claude Code

你需要的是能读写本地文件、能配置 MCP 的 Claude Code 环境。

安装和套餐以 Anthropic 官方页面为准。

这里不要写死价格。

因为价格、入口、地区可用性都可能变。

你只需要确认三件事：

```text
1. 你能打开 Claude Code。
2. Claude Code 能访问本地项目目录。
3. 你的环境支持 MCP server 配置。
```

如果你刚开始用，先别急着接 Obsidian。

先找一个空文件夹，让 Claude Code 做一个最小测试：

```text
请列出当前目录文件。
请新建一个 test.md，写一句 hello。
请读取 test.md 并总结。
```

这个测试通过，再接知识库。

否则你会分不清：

```text
是 Claude Code 没配置好，
还是 Obsidian 没接好，
还是 MCP 没连上。
```

## 4. 第三步：给 Obsidian 开 Local REST API

Claude 要进入 Obsidian，有两种常见路线。

第一种是直接把 Vault 当文件夹读。

这在第65期和第80期已经讲过。

第二种是通过 Obsidian 的 Local REST API 插件，让外部工具用 API 访问当前打开的 Vault。

这一篇讲第二种。

在 Obsidian 里：

```text
设置
-> 第三方插件
-> 关闭安全模式或启用 Community plugins
-> Browse
-> 搜索 Local REST API
-> 安装并启用
```

启用后，进入插件设置。

你会看到一个 API key。

有些界面会显示成：

```text
Bearer xxxxxxxx
```

后面配置时要注意：

```text
Authorization header 里需要 Bearer + key。
环境变量里有时只填 key 本体。
具体看你用的 MCP server 文档。
```

还有一个关键点：

```text
Obsidian 必须保持打开。
```

Local REST API 是开在本机上的门。

Obsidian 关了，这扇门也就没了。

## 5. 第四步：用 MCP 把 Claude 和 Obsidian 连起来

Claude Code 支持添加 MCP server。

MCP 可以理解成：

```text
让模型安全调用外部工具的一套协议。
```

接 Obsidian 时，常见有两条路径。

### 路线 A：Local REST API 自带或兼容的 MCP 入口

如果你安装的 Local REST API 版本提供 MCP endpoint，可以优先按它的 README 来。

概念上类似：

```bash
claude mcp add --transport http obsidian https://127.0.0.1:27124/mcp/ \
  --header "Authorization: Bearer YOUR_OBSIDIAN_KEY"
```

这条路线的好处是直接。

Obsidian 插件提供本机 HTTP 服务，Claude Code 通过 MCP 连接。

但具体 URL、端口、header 写法，务必以你安装版本的插件说明为准。

### 路线 B：stdio 桥接工具

有些教程会用类似下面的方式：

```bash
claude mcp add-json obsidian-vault '{
  "type": "stdio",
  "command": "uvx",
  "args": ["mcp-obsidian"],
  "env": {
    "OBSIDIAN_API_KEY": "PASTE-YOUR-KEY",
    "OBSIDIAN_HOST": "127.0.0.1",
    "OBSIDIAN_PORT": "27124"
  }
}'
```

这类写法的意思是：

```text
Claude Code 启动一个本地 stdio MCP server。
这个 server 再去调用 Obsidian Local REST API。
```

如果你照着某个 GitHub 仓库配置，一定看清楚它当前的 README。

不要只抄二手教程里的命令。

因为 MCP server 名称、环境变量、端口、认证方式都可能更新。

## 6. 第五步：做最小连通测试

配置完以后，不要马上让 Claude 整理整个知识库。

先做最小测试。

给 Claude 说：

```text
请通过 Obsidian MCP 列出当前 Vault 的顶层文件和文件夹。
只读，不要修改任何文件。
```

如果它能列出：

```text
Hello Brain.md
goals.md
```

说明门通了。

再做第二个测试：

```text
请读取 Hello Brain.md，并告诉我里面链接到了哪个笔记。
不要修改文件。
```

第三个测试才允许写入：

```text
请在 Vault 根目录新建 inbox/test-mcp.md。
内容只写一行：MCP connected.
写完后再读取确认。
```

这三步很重要。

它能帮你把问题分开：

```text
能列文件：连接通。
能读文件：权限通。
能写文件：写入通。
```

不要第一步就让它“整理我所有笔记”。

那是把新车刚启动就开上高速。

## 7. 第六步：让 Claude 先采访你

空知识库没有意义。

真正节省时间的，不是让 Claude 记住某个插件配置。

而是让它知道你这个人和你的工作现场。

可以复制这段给 Claude：

```text
你正在帮我搭建第二大脑。
请一次只问我一个问题，采访我并建立个人工作档案。

需要了解：
1. 我是谁，我主要做什么。
2. 今年最重要的三个目标。
3. 我当前正在推进的项目。
4. 我希望你怎么和我沟通。
5. 我的强项、弱项和容易拖延的地方。
6. 哪些内容属于隐私或禁止写入。

每次只问一个问题，等我回答后再问下一个。
采访结束后，把结果整理到 Vault 根目录的 CLAUDE.md。
请用清晰标题、项目符号和可执行规则。
```

回答时不要写人设。

写真实情况。

你答得越真实，后面 Claude 越少误判。

## 8. 第七步：把 CLAUDE.md 写成入口，不要写成自传

`CLAUDE.md` 不是越长越好。

它应该像一张入库须知。

一个可用的最小模板是：

```markdown
# CLAUDE.md

## 我是谁

- 我主要做：这里写你的业务/工作/创作方向。
- 当前重点：这里写 1-3 个当前最重要项目。
- 默认语言：中文。

## Vault 规则

- 先读本文件，再处理笔记。
- 不要修改 raw/ 里的原始资料，除非我明确要求。
- 新材料先放 inbox/，整理后再移动。
- 每次改动多个文件前，先说明计划。
- 不要删除文件。需要删除时先列出候选并等我确认。

## 目录说明

- inbox/: 新材料入口。
- raw/: 原始资料，只读。
- wiki/: Claude 整理后的概念页和主题页。
- projects/: 当前项目。
- people/: 人物和客户资料。
- decisions/: 决策记录。

## 输出偏好

- 默认用 Markdown。
- 先给结论，再给依据。
- 涉及事实时标注来源文件。
- 不确定就写不确定，不要编。
```

这份文件的价值是：

```text
每次 Claude 进入 Vault，都先知道边界。
```

它不是你的全部人生。

它是 Claude 的作业卡。

## 9. 第八步：建最小目录，不要一口气上复杂系统

第一天只建这几个目录就够了：

```text
brain/
  CLAUDE.md
  inbox/
  raw/
  wiki/
  projects/
  people/
  decisions/
```

不要一上来就把 PARA、Zettelkasten、GTD、日记系统、标签体系、Dataview、Canvas 全装上。

复杂结构不是第二大脑。

持续使用才是第二大脑。

你可以先给 Claude 一个任务：

```text
请检查当前 Vault 顶层目录。
如果缺少 inbox、raw、wiki、projects、people、decisions，请创建。
不要移动已有文件。
```

再给它第二个任务：

```text
请创建 wiki/_index.md，说明这个知识库当前有哪些主要区域。
如果信息不足，请只写已有目录，不要猜。
```

这样系统就有了入口页。

## 10. 第九步：每天只让它做三件事

刚搭好时，不要追求全自动。

每天只让 Claude 做三件低风险任务：

```text
1. 整理 inbox/ 里新增的材料。
2. 给新笔记补 3-5 个相关链接。
3. 写一段今天的变更摘要。
```

可以用这个提示词：

```text
请维护我的 Obsidian Vault。

范围：
- 只处理 inbox/、wiki/、projects/。
- raw/ 只能读取，不能修改。
- 不要删除文件。

任务：
1. 扫描 inbox/ 的新增笔记。
2. 判断它们应该归到哪个主题。
3. 必要时在 wiki/ 新建或更新主题页。
4. 给每条新笔记补充相关 [[链接]]。
5. 最后写一段维护摘要，保存到 decisions/YYYY-MM-DD-vault-log.md。

开始前先列计划，等我确认后再执行。
```

这套动作跑顺了，再考虑定时任务。

如果你使用的客户端有 Schedule / 自动任务能力，可以把维护任务排到每天固定时间。

但第一周我建议手动跑。

手动跑的好处是你能观察：

```text
Claude 会不会乱移动文件。
链接会不会乱编。
摘要有没有来源。
目录结构是否适合你。
```

自动化应该建立在稳定流程上。

## 11. 第十步：接 4SAPI 做成本和日志治理

个人用 Claude Code 处理一个小 Vault，直接用官方模型就可以。

但如果你是团队场景，就会遇到新问题：

```text
谁在整理知识库？
用了哪个模型？
一次整理扫了多少文件？
哪些内容发给了模型？
成本是否超预算？
是否有敏感资料进入上下文？
```

这就是大模型 API 中转站和企业 API 网关的价值。

你可以把 4SAPI 或企业级 API 入口放在模型调用层：

```text
Obsidian Vault
-> Claude Code / MCP 工具
-> 4SAPI / 企业 API 网关
-> Claude / GPT / Gemini / 其他模型
```

重点不是“多一个中转”。

重点是有统一治理：

```text
Key 分组。
项目维度账单。
日志审计。
失败重试记录。
预算控制。
敏感词和脱敏策略。
模型路由。
```

知识库越重要，越不能只靠口头约定。

## 12. 常见坑

第一个坑：把 API key 原样贴进公开笔记。

不要这么做。

API key 应该进本机环境变量或安全配置，不要写进会同步的 Markdown。

第二个坑：让 Claude 一次整理整个 Vault。

第一次只让它处理 5-10 篇笔记。

你要先看它的整理习惯。

第三个坑：把 `CLAUDE.md` 写成几千行。

它应该短、硬、可执行。

完整背景应该拆到：

```text
people/
projects/
decisions/
wiki/
```

第四个坑：只写“不要删除”。

提示词不是权限。

如果你真的不希望它删文件，要配合：

```text
只读 key。
Git 备份。
人工确认。
文件权限。
```

第五个坑：以为接上 MCP 就等于有第二大脑。

MCP 只是门。

第二大脑靠的是持续输入、持续链接、持续复盘。

## 13. 最小可用版本

如果你今天只想搭一个能跑的版本，按这个清单来：

```text
[ ] 安装 Obsidian
[ ] 建一个 brain Vault
[ ] 安装并启用 Local REST API
[ ] 复制 API key
[ ] 在 Claude Code 添加 Obsidian MCP
[ ] 让 Claude 只读列出 Vault 文件
[ ] 让 Claude 新建一条 test-mcp.md
[ ] 让 Claude 采访你
[ ] 写入 CLAUDE.md
[ ] 创建 inbox/raw/wiki/projects/people/decisions
```

跑通以后，你已经有了一个“可被 Claude 读写的本地知识库”。

后面再加 raw/wiki 自动编译、Obsidian Skills、定时任务和企业权限治理。

那些是下一篇的事。

## 总结

把 Claude 接上 Obsidian，真正改变的不是笔记软件。

改变的是你的工作方式。

以前是：

```text
每次开新对话，都重新解释背景。
```

接上以后是：

```text
背景在 Vault 里。
规则在 CLAUDE.md 里。
材料在 raw/ 和 inbox/ 里。
Claude 每次进来先读现场，再开始干活。
```

别再把模型当搜索框。

把它接到你的知识库，让它从“回答问题”变成“维护系统”。

## 资料来源与延伸阅读

- Claude Code MCP 文档：https://docs.anthropic.com/en/docs/claude-code/mcp
- Claude Code Memory / CLAUDE.md 文档：https://docs.anthropic.com/en/docs/claude-code/memory
- Obsidian 官方网站：https://obsidian.md/
- Obsidian 数据存储说明：https://obsidian.md/help/data-storage
- Obsidian Local REST API：https://github.com/coddingtonbear/obsidian-local-rest-api

