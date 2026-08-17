---
title: "【大模型API中转站】第80期 Obsidian直连Codex | CLI极简指南"
category: 人工智能
tags:
  - 大模型API中转站
  - Obsidian
  - Codex
  - CLI
  - Markdown
  - 4SAPI
description: "不在 Obsidian 里安装任何第三方插件，直接把 Obsidian Vault 当成本地 Markdown 项目目录，用终端 cd 进入后交给 Codex CLI 读取、整理、改写和批量处理，并说明如何用 AGENTS.md 和 4SAPI 做长期规则与成本治理。"
---

# 【大模型API中转站】第80期 Obsidian直连Codex | CLI极简指南

本文是【大模型API中转站】系列的第80篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

很多人想把 Obsidian 接上 AI，第一反应是：

```text
去插件市场找插件。
```

装一个 ChatGPT 插件。

再装一个自动摘要插件。

再装一个文本生成插件。

再研究 API Key 填哪里。

最后 Obsidian 越来越重。

笔记还没变聪明，软件先变复杂了。

其实最简单的办法不是装插件。

而是反过来想：

```text
Obsidian 的本质是什么？
```

答案很简单：

```text
一个本地 Markdown 文件夹。
```

你的每篇笔记，就是一个 `.md` 文件。

你的图片、附件、日记、项目文档，也都在这个文件夹里。

既然它本质上是一个普通文件夹，那 Codex CLI 根本不需要通过 Obsidian 插件才能读它。

你只要：

```text
打开终端
cd 到 Obsidian 库目录
启动 Codex
```

就可以让 Codex 像处理代码项目一样处理你的知识库：

- 搜索笔记
- 整理目录
- 改写草稿
- 合并重复内容
- 生成索引
- 补充双向链接
- 提取待办
- 批量规范 front matter
- 按你的规则归档素材

这篇就是一套极简指南。

不装 Obsidian 插件。

不改 Obsidian 配置。

不把笔记同步到第三方平台。

只用：

```text
Obsidian 本地库 + 终端 + Codex CLI
```

## 1. 为什么我推荐 CLI 直连？

因为它干净。

Obsidian 最宝贵的地方，不是插件生态。

而是：

```text
你的知识是本地文件。
```

这意味着：

- 你可以用任何编辑器打开
- 可以用 Git 备份
- 可以用脚本批量处理
- 可以被 Codex、Claude Code、Cursor 这类工具读取
- 不被某个 AI 插件绑定
- 不需要把所有笔记交给一个插件托管

插件路线不是不好。

插件适合“在 Obsidian 里直接点按钮”的体验。

但它也会带来几个问题：

| 路线 | 优点 | 问题 |
| --- | --- | --- |
| Obsidian 插件 | 界面内操作方便 | 插件权限、兼容性、维护状态不稳定 |
| API 同步平台 | 自动化强 | 可能把笔记送到外部服务 |
| CLI 直连 | 干净、可控、通用 | 需要会一点终端 |

如果你是独立创作者、开发者、研究型博主，我更建议先走 CLI 直连。

原因很简单：

```text
它不改变你的 Obsidian。
它只是让 Codex 进入你的笔记文件夹工作。
```

这就像你找了一个助手坐在你的资料柜前。

他不是把资料柜搬走。

他是在原地帮你整理。

## 2. 前置条件

你需要准备三样东西：

| 项目 | 用途 |
| --- | --- |
| Obsidian Vault | 你的本地 Markdown 笔记库 |
| 终端 | Mac/Linux 用 Terminal，Windows 用 PowerShell 或 CMD |
| Codex CLI | 在终端里读取和处理当前目录 |

如果你已经在用 Codex 桌面版，也可以理解为：

```text
Codex CLI 是终端优先的本地工作方式。
Codex App 更适合图形界面、审查和交互式协作。
```

本文讲的是 CLI 直连。

核心动作只有一个：

```text
cd 到 Obsidian 库目录。
```

只要 Codex 的工作目录是你的 Vault，它看到的就是你的笔记项目。

## 3. 第一步：找到你的 Obsidian 库目录

Obsidian 里一个 Vault，就是一个文件夹。

你要先找到它在电脑上的路径。

常见路径可能是：

Mac：

```bash
/Users/你的用户名/Documents/我的Obsidian库
```

Windows：

```powershell
C:\Users\你的用户名\Documents\我的Obsidian库
```

Linux：

```bash
/home/你的用户名/Documents/我的Obsidian库
```

如果你忘了路径，可以在 Obsidian 里：

```text
打开 Vault
右键某篇笔记
选择在文件管理器中显示
往上一级就是库目录
```

也可以看 Obsidian 的 Vault 列表。

记住一点：

```text
要进入的是整个库的根目录，不是某一篇笔记所在的小目录。
```

比如你的结构是：

```text
我的Obsidian库/
  00_daily/
  10_sources/
  20_notes/
  30_projects/
  40_outputs/
```

你要 `cd` 到：

```text
我的Obsidian库/
```

而不是 `20_notes/`。

## 4. 第二步：打开终端

不同系统入口不一样。

Mac / Linux：

```text
打开 Terminal
```

Windows：

```text
打开 PowerShell 或 CMD
```

我更推荐 Windows 用户用 PowerShell。

因为路径、中文、引号处理相对舒服一点。

打开后，你会看到一个命令行。

不用害怕。

这篇只用一个命令：

```text
cd
```

它的意思是：

```text
切换目录。
```

## 5. 第三步：切换到 Obsidian 库目录

Mac / Linux 示例：

```bash
cd "/Users/你的用户名/Documents/我的Obsidian库"
```

Windows PowerShell 示例：

```powershell
cd "C:\Users\你的用户名\Documents\我的Obsidian库"
```

如果路径里有空格或中文，建议一定加引号。

比如：

```powershell
cd "C:\Users\admin\Documents\My Obsidian Vault"
```

不要写成：

```powershell
cd C:\Users\admin\Documents\My Obsidian Vault
```

后者遇到空格很容易被拆错。

### 5.1 怎么确认进对目录？

进入后，列一下文件。

Mac / Linux：

```bash
ls
```

Windows PowerShell：

```powershell
Get-ChildItem
```

如果你看到类似：

```text
00_daily
10_sources
20_notes
30_projects
40_outputs
.obsidian
```

说明你进对了。

`.obsidian` 是 Obsidian 自己的配置目录。

一般情况下，不要让 AI 随便改它。

## 6. 第四步：启动 Codex CLI

在 Obsidian 库目录里启动 Codex。

常见命令是：

```bash
codex
```

如果你的安装方式不同，以你本机 Codex CLI 实际命令为准。

重点不是命令名字。

重点是：

```text
启动 Codex 时，当前目录已经是 Obsidian Vault。
```

这样 Codex 的工作上下文就是你的笔记库。

你可以先发第一条指令：

```text
请先不要修改任何文件。
请阅读当前目录结构，告诉我：
1. 这是一个什么类型的 Obsidian 库
2. 主要文件夹分别可能承担什么职责
3. 哪些目录适合放原始素材
4. 哪些目录适合放输出稿件
5. 如果我要让你帮我整理这个知识库，建议先从哪里开始
```

注意这句：

```text
请先不要修改任何文件。
```

第一次接入，先让 Codex 看。

不要上来就让它改。

## 7. 第一次让 Codex 做什么？

我建议第一轮只做“盘点”。

不要让它整理全库。

不要让它批量改名。

不要让它删除重复文件。

更稳的第一条任务是：

```text
请扫描当前 Obsidian Vault 的顶层目录和最近 20 篇笔记。
不要修改文件。
请输出：
1. 当前库的主要主题
2. 笔记命名是否混乱
3. 哪些目录可能需要整理
4. 是否有明显重复内容
5. 建议的最小整理方案
```

这个任务很安全。

它只读，不写。

你可以先看 Codex 是否理解你的笔记库。

如果它理解错了，再纠正。

等它能说清楚你的结构，再让它动手。

## 8. 给 Obsidian 库加一个 AGENTS.md

如果你打算长期用 Codex 管理 Obsidian，建议在 Vault 根目录放一个：

```text
AGENTS.md
```

它相当于给 Codex 的长期说明书。

官方 Codex 文档也推荐用 `AGENTS.md` 记录项目级指令、约定、命令和验证方式。

放到 Obsidian 库里以后，Codex 每次进入这个目录，都能先看到你的规则。

可以这样写：

```markdown
# AGENTS.md

## 这个库是什么

这是我的 Obsidian 知识库，用于存放读书笔记、AI 工具研究、文章草稿、项目资料和发布稿。

## 重要规则

- 默认不要修改 `.obsidian/` 目录。
- 默认不要删除任何笔记。
- 批量改名、移动文件、删除重复内容前必须先给计划。
- 所有新文章草稿放到 `40_outputs/`。
- 原始资料放到 `10_sources/`。
- 精炼后的永久笔记放到 `20_notes/`。
- 项目相关资料放到 `30_projects/`。
- 文章要保持中文口语化，不要像论文。
- 不要编造来源。
- 引用外部资料时保留链接。

## 常用任务

- 整理 inbox
- 把素材改成公众号文章
- 给文章补标题和摘要
- 检查笔记之间是否需要互链
- 从会议记录提取待办
- 把零散观点整理成选题库

## 发布规范

- 文章开头要有具体场景。
- 正文要有小标题和表格。
- 结尾要有清单或行动建议。
- 涉及价格、模型、法规、产品功能时，发布前提醒我复核。
```

这个文件非常重要。

它能减少你每次重复解释：

```text
不要动 .obsidian。
不要删文件。
草稿放哪里。
我的文章风格是什么。
```

如果你的 Obsidian 是长期知识库，`AGENTS.md` 就是 Codex 的“入库须知”。

## 9. 推荐目录结构

如果你的库还比较乱，可以先整理成这样：

```text
ObsidianVault/
  AGENTS.md
  _inbox/
  00_daily/
  10_sources/
  20_notes/
  30_projects/
  40_outputs/
  90_archive/
```

每个目录职责：

| 目录 | 用途 |
| --- | --- |
| `_inbox/` | 临时想法、待整理素材 |
| `00_daily/` | 日记、每日记录 |
| `10_sources/` | 外部资料、摘录、网页剪藏 |
| `20_notes/` | 提炼后的永久笔记 |
| `30_projects/` | 项目资料、任务、会议 |
| `40_outputs/` | 文章、脚本、发布稿 |
| `90_archive/` | 已归档内容 |

不一定非要照抄。

但你需要让 Codex 知道：

```text
什么是原始素材。
什么是正式笔记。
什么是可发布输出。
什么是归档。
```

AI 最怕文件夹职责不清。

职责不清，它就会乱放。

## 10. 五个最适合 Codex 处理的 Obsidian 任务

### 10.1 把 inbox 整理成可用笔记

提示词：

```text
请阅读 `_inbox/` 目录下最近 20 篇笔记。
不要删除原文件。
请把它们分成：
1. 可发展成文章的选题
2. 值得沉淀的永久笔记
3. 需要补资料的素材
4. 可以归档的碎片

先给整理计划，不要直接移动文件。
```

### 10.2 把素材改成公众号文章

提示词：

```text
请基于 `10_sources/xxx.md` 和 `20_notes/yyy.md`，生成一篇公众号长文草稿。

要求：
1. 输出到 `40_outputs/`。
2. 保留引用来源链接。
3. 不要编造资料。
4. 语气像独立科技博主。
5. 文章要有标题、开头、结构、表格、结尾清单。

写之前先给大纲。
```

### 10.3 给旧笔记补链接

提示词：

```text
请阅读 `20_notes/` 下与 AI Agent 相关的笔记。
找出适合互相链接的笔记。

要求：
1. 先列出建议链接关系。
2. 说明为什么应该互链。
3. 不要直接修改文件，等我确认。
```

### 10.4 从日记里提取任务

提示词：

```text
请阅读 `00_daily/` 最近 7 天的日记。
提取未完成任务、等待他人回复的事项、值得沉淀的想法。

输出：
1. 待办清单
2. 等待清单
3. 选题灵感
4. 建议移动到 `20_notes/` 的内容

不要修改原日记。
```

### 10.5 批量检查文章发布前问题

提示词：

```text
请检查 `40_outputs/` 中这篇文章：
[文件名]

检查维度：
1. 标题是否清楚
2. 开头是否有具体场景
3. 结构是否完整
4. 是否有空泛表达
5. 是否缺少表格或清单
6. 是否有需要事实复核的内容
7. 是否符合我的 AGENTS.md 发布规范

只输出问题和修改建议，不要直接改。
```

## 11. 哪些事不要一上来让 Codex 做？

不要第一次就做高风险批量操作。

比如：

- 批量删除重复笔记
- 批量改名
- 批量移动全库
- 改 `.obsidian/` 配置
- 重写所有笔记
- 自动给所有笔记加标签
- 把全库压缩成一个文件

这些操作不是不能做。

而是要分阶段。

正确做法：

```text
先扫描
再给计划
再小范围试点
再看 diff
最后批量执行
```

尤其是删除。

我的建议是：

```text
不要让 AI 删除笔记。
让它移动到 90_archive/待删除/，你人工确认后再删。
```

知识库不是临时代码。

很多笔记的价值要过几个月才会显出来。

删错很可惜。

## 12. Git 备份：强烈建议加

如果你要让 Codex 改 Obsidian，强烈建议用 Git 做版本管理。

哪怕你不懂 Git，也至少知道两件事：

```text
改之前能保存一个状态。
改坏了能回去。
```

你可以让 Codex 帮你检查：

```text
当前 Obsidian Vault 是否已经是 Git 仓库？
如果不是，请先不要初始化。
请告诉我初始化 Git 备份的步骤和风险。
```

不要让它直接把全库推到公开 GitHub。

很多 Obsidian 里有隐私内容：

- 日记
- 客户资料
- 账号信息
- 商业想法
- 会议记录
- 未发布文章

如果要远程备份，优先用私有仓库或你信任的同步方式。

## 13. 4SAPI 在这里有什么用？

有人会问：

```text
Obsidian 连 Codex，为什么还要提 4SAPI？
```

因为当你开始让 AI 批量处理知识库时，模型调用会变多。

比如：

- 整理 100 篇 inbox
- 给 200 篇笔记补摘要
- 批量生成文章大纲
- 批量检查发布稿
- 从一年日记里提取主题
- 给项目资料生成周报

这些任务不是一次聊天。

它们会变成批处理。

批处理就会带来三个问题：

```text
用哪个模型？
花了多少钱？
结果值不值得？
```

这时 4SAPI 的价值就出来了。

你可以把模型调用统一到一个入口：

```text
Base URL: https://4sapi.com/v1
API Key: 你的 4SAPI Key
Model: 按任务选择
```

然后按任务拆 Key：

```text
obsidian-inbox-clean
obsidian-article-draft
obsidian-review
obsidian-summary-lowcost
```

不同任务用不同模型：

| 任务 | 模型策略 |
| --- | --- |
| inbox 分类 | 低成本模型 |
| 摘要生成 | 低成本模型 |
| 文章初稿 | 中等模型 |
| 重要文章审稿 | 强模型 |
| 事实风险提醒 | 稳定强模型 |

这样你就不会所有任务都烧最贵模型。

也不会到处散落 API Key。

对于 Obsidian 这种长期知识库，4SAPI 的价值不是“炫技”。

而是：

```text
让你的知识库 AI 批处理成本可控、日志可查、模型可换。
```

## 14. 一个完整的 Obsidian + Codex 工作流

你可以这样用：

```text
1. 平时所有灵感先扔进 `_inbox/`
2. 每周打开终端，cd 到 Obsidian Vault
3. 启动 Codex CLI
4. 让 Codex 扫描 `_inbox/`
5. 先输出整理计划
6. 你确认后，让它移动或改写小范围文件
7. 让它把好素材整理成 `20_notes/`
8. 把可发布内容写到 `40_outputs/`
9. 发布前让它按 AGENTS.md 检查
10. 用 Git 保存版本
```

如果后面任务量变大，再接自动化：

```text
每天整理 inbox
每周生成选题库
每月检查孤立笔记
每次发布前自动审稿
```

到这一步，4SAPI 就更重要。

因为你的模型调用已经不是偶尔问一句。

而是稳定工作流的一部分。

## 15. 常见问题

### 15.1 会不会破坏 Obsidian？

只要不乱改 `.obsidian/`，一般不会。

因为笔记本质是 Markdown 文件。

但任何批量操作都有风险。

所以：

```text
先读不写。
先计划后执行。
先小范围试点。
重要操作前备份。
```

### 15.2 Codex 能不能理解双向链接？

能读 Markdown 文本里的链接形式。

比如：

```text
[[某篇笔记]]
```

但它不是 Obsidian 本体。

它不会像 Obsidian 图谱视图那样实时展示关系。

更适合让它：

```text
分析链接是否合理。
建议新增链接。
检查孤立笔记。
```

### 15.3 要不要装 Obsidian 插件？

不是必须。

本文路线的核心就是：

```text
不装插件也能用 AI 处理 Obsidian。
```

如果你后面需要 Dataview、Templater、Omnisearch 这类插件，可以装。

但先别让插件变成门槛。

### 15.4 能不能全自动整理？

能做，但不建议一开始就做。

全自动整理之前，至少要有：

- 明确目录规则
- AGENTS.md
- Git 备份
- 小范围试点
- 人工确认
- 成本记录

否则你会得到一个“很努力但乱动文件”的 AI。

### 15.5 Windows 路径怎么写？

PowerShell 里建议用引号：

```powershell
cd "C:\Users\admin\Documents\我的Obsidian库"
```

如果路径有中文，一般也没问题。

但为了减少工具链问题，长期项目最好路径简洁一点。

## 16. 最后给一套检查清单

开始前过一遍：

```text
[ ] 我知道 Obsidian Vault 的根目录在哪里
[ ] 我能在终端 cd 到这个目录
[ ] 我能看到 `.obsidian/` 和笔记文件夹
[ ] 第一次任务只读不写
[ ] Vault 根目录有 AGENTS.md
[ ] AGENTS.md 写明不要动 `.obsidian/`
[ ] 批量操作前先给计划
[ ] 删除操作改成移动到 archive
[ ] 重要修改前有备份或 Git
[ ] 涉及批量 AI 处理时，会记录成本和模型
[ ] 高频任务考虑用 4SAPI 拆 Key 和看日志
```

## 17. 总结

Obsidian 链接 Codex，最简单的方式不是装插件。

而是：

```text
cd 到你的 Obsidian Vault。
启动 Codex CLI。
把它当成本地 Markdown 项目来处理。
```

这条路线的优势是：

- Obsidian 保持纯净
- 笔记仍然是本地 Markdown
- Codex 可以直接读写文件
- AGENTS.md 可以固定长期规则
- Git 可以保护版本
- 4SAPI 可以承接批量处理时的模型和成本治理

记住第一条命令就够了：

```bash
cd "/Users/你的用户名/Documents/我的Obsidian库"
```

Windows 就是：

```powershell
cd "C:\Users\你的用户名\Documents\我的Obsidian库"
```

进入库目录后，再启动 Codex。

你的 Obsidian 就不再只是笔记软件。

它会变成一个能被 AI 读取、整理、改写和复用的本地知识工作台。

参考资料：

- Codex 官方文档：https://developers.openai.com/codex/
- Codex CLI 文档：https://developers.openai.com/codex/cli/
- Codex AGENTS.md 说明：https://developers.openai.com/codex/customization/
- Obsidian 官方网站：https://obsidian.md/
- 4SAPI 接入文档：https://4sapi.apifox.cn/
- 4SAPI 官网：https://4sapi.com/
