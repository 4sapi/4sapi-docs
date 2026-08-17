---
title: "【大模型API中转站】第65期 Obsidian+Claude Code | 搭建第二大脑"
category: 人工智能
tags:
  - 大模型API中转站
  - Obsidian
  - Claude Code
  - 第二大脑
  - 知识管理
  - 4SAPI
  - 4sapi.com
description: "基于 Obsidian 本地 Markdown Vault、Claude Code 项目记忆、Claudian 插件、Dataview、Omnisearch 和 Git 备份，搭建一个能被 AI 读取、整理、改写、归档和长期复用的个人知识工作台，并说明如何用 4sapi.com 这类大模型 API 中转站做批处理和成本治理。"
---

# 【大模型API中转站】第65期 Obsidian+Claude Code | 搭建第二大脑

本文是【大模型API中转站】系列的第65篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

这篇讲一个很多独立创作者、开发者和知识工作者都会遇到的问题：

```text
笔记越来越多，AI 工具也越来越多，但真正能长期复用的知识系统却没有变多。
```

很多人用 Obsidian，只是在收集资料。

很多人用 Claude Code，只是在让它写代码。

但如果把两者放在一起，它们会变成一套很有意思的知识工作台：

```text
Obsidian 负责存放你的长期知识。
Claude Code 负责读取、整理、改写和执行工作流。
大模型 API 中转站负责统一模型调用、批量处理和成本治理。
```

这套系统的目标不是装很多插件，也不是追求一个很复杂的知识图谱。

真正目标只有一个：

```text
让你的经历、观点、素材、项目和工作流，都沉淀成 AI 能读懂、你能长期复用的本地资产。
```

## 1. 为什么是 Obsidian？因为它像一个“可编程文件夹”

Obsidian 最大的价值，不是界面好看。

而是它把笔记存在本地 Markdown 文件里。

官方文档里说得很直接：Obsidian 的笔记是 Markdown 格式的纯文本文件，Vault 本质上就是本地文件系统里的一个文件夹。

这句话对 AI 工作流特别重要。

因为本地 Markdown 有几个优势：

- 可以用任意编辑器打开
- 可以用 Git 做版本管理
- 可以被 Claude Code、Codex、Cursor 这类工具直接读取
- 可以被脚本批量处理
- 可以迁移，不容易被平台锁死
- 可以通过链接、标签、front matter 形成结构

如果 Notion 更像一个在线工作台，那 Obsidian 更像一个你自己的知识仓库。

它不是把知识关在某个平台数据库里，而是放在你电脑上的一堆 Markdown 文件里。

这就给 AI 留出了很大的空间。

Claude Code 不需要通过复杂 API 才能理解你的知识库。

只要它能进入这个 Vault，就可以：

- 搜索文件
- 阅读笔记
- 修改草稿
- 归档素材
- 生成目录
- 提炼观点
- 建议双向链接
- 运行脚本
- 根据你的规则生成新文件

这就是 Obsidian + Claude Code 的关键：

```text
Obsidian 让知识变成文件系统。
Claude Code 让文件系统变成可协作的工作台。
```

## 2. 先别急着装插件，先搭最小架构

很多人第一次搭 Obsidian，会先去搜“最佳插件列表”。

这很容易走偏。

第二大脑不是靠插件数量堆出来的。

建议先搭一个最小结构：

```text
my-brain/
  _agents/
    user-manual.md
    ip-position.md
  _inbox/
  _resources/
  00_daily/
  10_sources/
  20_notes/
  30_projects/
  40_outputs/
  90_archive/
  CLAUDE.md
```

每个文件夹只承担一个职责。

### `_agents/`

放你的 AI 档案。

这里不是普通笔记，而是给 AI 看的说明书。

比如：

- 你是谁
- 你擅长什么
- 你不碰什么
- 你的内容方向
- 你的表达风格
- 你的协作规则

### `_inbox/`

放所有临时信息。

比如：

- 微信收藏
- 飞书文档摘录
- 浏览器剪藏
- 临时灵感
- 会议记录
- 未整理资料

所有东西先进 inbox，后面再整理。

不要一开始就想把每条信息放到完美位置。

### `_resources/`

放附件。

比如图片、PDF、网页截图、导出文件。

建议按文章或项目分文件夹：

```text
_resources/
  2026-Obsidian-Claude-Code/
    cover.png
    screenshot-01.png
    source.pdf
```

这样以后迁移、发布、备份都清楚。

### `00_daily/`

放每日记录。

适合写：

- 今天看了什么
- 今天做了什么
- 今天想到什么
- 今天要跟进什么

每日笔记不是为了写日记。

而是给知识系统提供时间轴。

### `10_sources/`

放原始资料。

比如：

- 文章摘录
- 论文笔记
- 视频笔记
- 课程笔记
- 访谈记录
- 产品文档

这里尽量保留来源，不要一上来就改写。

### `20_notes/`

放你的二次理解。

比如：

- 一个观点
- 一个模型
- 一个判断
- 一个案例
- 一个概念解释

这层才是知识管理真正开始变值钱的地方。

### `30_projects/`

放正在推进的项目。

比如：

- 一个课程
- 一个产品
- 一个公众号专题
- 一个客户交付
- 一个研究计划

项目笔记需要明确目标、进度、素材和下一步。

### `40_outputs/`

放最终产出。

比如：

- 公众号文章
- 小红书文案
- 课程大纲
- 产品文档
- PRD
- 研究报告

输出区应该干净。

不要把半成品和素材都塞进去。

## 3. 让 AI 认识你：三份核心文件

Obsidian 真正变成第二大脑，不是因为它有知识图谱。

而是因为 AI 开始知道：

- 你是谁
- 你在做什么
- 你怎么表达
- 你的边界是什么
- 你的文件应该放在哪里

建议在 Vault 里准备三份文件。

### 文件一：`_agents/user-manual.md`

这是“人物使用说明书”。

最小版本这样写：

```md
# 人物使用说明书

## 1. 我是谁

一句话介绍：

三个标签：
- #
- #
- #

## 2. 我最熟的三个领域

### 领域一
- 我能讲的具体话题：
- 我不碰的内容：

### 领域二
- 我能讲的具体话题：
- 我不碰的内容：

### 领域三
- 我能讲的具体话题：
- 我不碰的内容：

## 3. 一段真实经历

用 3 到 5 句话写一段真实经历。
可以是入行经历、失败经历、转折点、长期坚持的原因。

## 4. 内容禁区

- 绝对不碰的话题：
- 绝对不用的语气：
- 不想被代表的立场：

## 5. 协作协议

- 直接给方案，不要反复问我要什么风格
- 不确定就标注“推测”
- 引用我的笔记时注明来源
- 批量修改文件前先列清单
- 不要替我做最终发布和决策
```

这份文件不需要长。

真实比完整重要。

一行真实经历，比十行空泛人设有用。

### 文件二：`_agents/ip-position.md`

如果你做内容、产品、咨询、课程或个人品牌，这份文件很重要。

```md
# IP 定位说明书

## 1. 定位宣言

我是那个在 ______ 领域坚持 ______ 比 ______ 更重要的人。

## 2. 我为什么不同

- 视角差异：
- 经历差异：
- 立场差异：

## 3. 我在对谁说话

我的受众是：

他们的痛点：

他们想成为：

他们为什么关注我：

## 4. 三大内容支柱

1.
2.
3.

超出这三条，不优先做。

## 5. 我不做什么

当别人都在做 ______，我不做，因为 ______。

## 6. 常用表达

- 「」
- 「」
- 「」
```

这份文件解决的是一致性。

你不是每次都让 AI 从零猜你的定位，而是让它每次都先读一遍你的底层设定。

### 文件三：`CLAUDE.md`

如果你用 Claude Code，这个文件非常关键。

Claude Code 官方文档说明，`CLAUDE.md` 可以给 Claude 提供跨会话的持久指令。它不是系统提示词，但每次会话都会被加载，适合放项目规则、工作方式和固定约束。

建议放在 Vault 根目录：

```md
# AI 协作说明书

## 1. 工作区说明

这是我的 Obsidian Vault，用于管理个人知识、内容创作、项目资料和长期思考。

请默认先读：
- `_agents/user-manual.md`
- `_agents/ip-position.md`

## 2. 默认语言和输出

- 默认中文
- 使用中文标点
- 输出适合手机阅读
- 结构清楚，不写空泛口号
- 不使用“保证有效”“一定赚钱”等绝对承诺

## 3. 文件安全规则

- 未经确认，不能删除文件
- 未经确认，不能移动或重命名文件
- 原始资料不能覆盖，只能新建衍生版本
- 不要修改 `_agents/` 下的文件，除非我明确要求
- 一次改动超过 3 个文件，先列清单再执行

## 4. 文件夹用途

- `_inbox/`：临时收集箱
- `10_sources/`：原始资料
- `20_notes/`：二次理解和永久笔记
- `30_projects/`：正在推进的项目
- `40_outputs/`：最终输出
- `_resources/`：图片、PDF、附件

## 5. 常用工作流

当我说“整理收件箱”时：
1. 读取 `_inbox/`
2. 判断每条内容应该归档到哪里
3. 先给归档清单
4. 我确认后再移动文件

当我说“写文章”时：
1. 读取 `_agents/user-manual.md`
2. 读取 `_agents/ip-position.md`
3. 搜索相关素材
4. 生成大纲
5. 等我确认后再写正文
6. 保存到 `40_outputs/`

当我说“复盘项目”时：
1. 读取对应 `30_projects/` 文件夹
2. 总结已完成、未完成、风险和下一步
3. 生成一份项目复盘笔记
```

`CLAUDE.md` 不要写成鸡汤。

要写成规则。

比如“写得好一点”不如“先给结论，再给步骤，每段不超过 5 行”。

## 4. Claude Code 接入 Obsidian 的三种方式

现在进入实际接入。

### 方式一：直接在 Vault 根目录运行 Claude Code

这是最稳的方式。

打开终端：

```bash
cd /path/to/my-brain
claude
```

然后让 Claude Code 先读规则：

```text
请先读取 CLAUDE.md、_agents/user-manual.md 和 _agents/ip-position.md。
然后总结这个 Vault 的目录结构和适合你的工作方式。
先不要修改任何文件。
```

这种方式的优点是简单、稳定、可控。

缺点是你需要在终端和 Obsidian 之间切换。

### 方式二：用 Obsidian Terminal 插件

如果你不想频繁切窗口，可以在 Obsidian 里装 Terminal 类插件，把终端嵌进 Obsidian。

这样你可以在 Obsidian 里直接运行：

```bash
claude
git status
git diff
python scripts/organize_inbox.py
```

适合已经习惯命令行的人。

### 方式三：用 Claudian 这类社区插件

Claudian 是一个 Obsidian 社区插件，它的 GitHub 说明里写到，它可以把 Claude Code、Codex、Opencode 等 coding agents 嵌进 Vault，让 Vault 变成 agent 的工作目录。

这类插件的体验更像“在 Obsidian 侧边栏里直接和 AI 协作”。

你可以选中文字做 inline edit，也可以让 AI 搜索、编辑、生成文件。

但要注意：它是社区插件，不是 Obsidian 或 Anthropic 官方功能。

社区 issue 里能看到一些兼容性问题，比如 Claude Code CLI 版本变化、Windows 路径、插件加载失败、CLI 参数变化等。

所以建议：

- 先在测试 Vault 里试
- 确认 Claude Code CLI 单独能运行
- 确认插件能找到正确 CLI 路径
- 不要一上来给它操作真实主库
- 重要 Vault 先用 Git 备份

## 5. 插件怎么选？别把 Vault 装成杂货铺

Obsidian 插件很多。

但做 AI 第二大脑，不需要一口气装几十个。

建议按“输入、检索、结构、输出、备份、执行”六类来选。

### 输入：Readwise、剪藏、图片下载

如果你经常在 Kindle、网页、Twitter/X、Instapaper 里做高亮，Readwise Official 可以把高亮同步进 Obsidian。

如果你经常从飞书、知乎、公众号复制内容，建议配合图片下载类插件，把远程图片落到本地 `_resources/`。

核心原则：

```text
原始资料先进 10_sources 或 _inbox。
不要直接混进最终输出区。
```

### 检索：Omnisearch

Obsidian 内置搜索够用，但 Vault 大了以后，Omnisearch 会更顺手。

Omnisearch 的 GitHub 说明里提到，它目标是快速定位文件，支持更智能的相关性排序，也能配合 Text Extractor 支持图片、文档和 PDF 索引。

对 AI 工作流来说，检索特别重要。

因为你不应该每次把整个 Vault 都丢给模型。

更好的方式是：

```text
先搜索，再读取相关文件，再生成结果。
```

### 结构：Dataview

Dataview 是很多 Obsidian 用户的核心插件。

官方说明里把它定义为个人知识库上的实时索引和查询引擎，可以基于 metadata、标签、任务、front matter、inline fields 做列表、过滤、排序、分组。

比如你可以在首页做一个内容工作台：

```dataview
TABLE status, platform, topic
FROM "40_outputs"
WHERE status != "published"
SORT file.mtime DESC
```

也可以做项目看板：

```dataview
TABLE deadline, owner, status
FROM "30_projects"
WHERE status != "done"
SORT deadline ASC
```

这类动态视图很适合给 Claude Code 使用。

因为它让 Vault 不只是“文件夹”，而是一个可查询的知识数据库。

### 模板：Templater

Templater 可以创建动态模板，支持变量、函数结果、JavaScript 和系统命令。

它适合做：

- 每日笔记模板
- 会议纪要模板
- 读书笔记模板
- 文章草稿模板
- 项目复盘模板

比如文章模板：

```md
---
status: draft
platform: 公众号
topic:
source:
---

# <% tp.file.title %>

## 核心观点

## 目标读者

## 素材来源

## 大纲

## 正文

## 发布检查

- [ ] 是否有夸大表达
- [ ] 是否注明来源
- [ ] 是否检查图片路径
- [ ] 是否保存到 40_outputs
```

### 备份：Obsidian Git

Obsidian Git 可以把 Vault 接入 Git 版本管理，支持 commit、pull、push 和自动备份。

这对 AI 协作非常关键。

因为只要让 AI 具备改文件能力，就一定要有回滚机制。

建议：

- Vault 初始化 Git
- 每次批量修改前先 commit
- 每天至少自动备份一次
- `_resources/` 大文件太多时考虑 Git LFS 或单独备份
- `.obsidian/` 里个人设置要不要同步，按团队情况决定

### 执行：Terminal / Claudian / Claude Code

Terminal 负责命令。

Claude Code 负责理解和执行工作流。

Claudian 负责把 agent 体验嵌回 Obsidian。

这三者不要混成一件事。

终端是基础设施。

Claude Code 是 agent。

Obsidian 插件是界面和入口。

## 6. 大模型 API 中转站放在哪里？

很多人会问：

```text
既然用了 Claude Code，还需要大模型 API 中转站吗？
```

答案是：看你的任务。

如果你只是个人在 Vault 里偶尔让 Claude Code 改笔记，未必需要中转站。

但只要出现下面这些需求，中转站就有价值：

- 批量整理几百篇笔记
- 给大量文章生成摘要
- 多模型交叉审稿
- 用便宜模型做初筛，用强模型做深度改写
- 把 Obsidian 知识库接到自己的 Web 工具
- 给团队提供统一的知识处理 API
- 统计不同工作流的 token 成本
- 统一管理 Claude、GPT、Gemini 等模型入口

这时可以把 4sapi.com 这类大模型 API 中转站放在模型调用层：

```text
Obsidian Vault -> 脚本/插件/后台服务 -> 4SAPI / 大模型 API 中转站 -> Claude/GPT/Gemini
```

注意一个边界：

```text
Claude Code CLI 自身怎么认证、是否支持某个兼容端点，要以 Claude Code 和插件实际支持为准。
中转站更适合你自写脚本、后台服务、批处理工具、内容生成工具统一调用多模型。
```

不要把“Claude Code 能读本地文件”和“所有调用都能无缝走中转站”混为一谈。

更稳的做法是：

- Claude Code 负责本地文件操作和复杂工作流
- 4SAPI 负责脚本、服务端、批量任务的模型调用
- Git 负责回滚
- CLAUDE.md 负责长期规则

## 7. 一个最小 API 批处理示例

假设你想批量整理 `_inbox/` 里的 Markdown 文件，把每篇内容生成摘要、标签和建议归档位置。

可以先用中转站跑一个测试脚本。

环境变量：

```env
LLM_BASE_URL=https://你的中转站地址/v1
LLM_API_KEY=你的中转站Key
LLM_MODEL=claude-sonnet
```

Python 示例：

```python
import os
import pathlib
import requests

vault = pathlib.Path("C:/path/to/my-brain")
inbox = vault / "_inbox"

base_url = os.environ["LLM_BASE_URL"]
api_key = os.environ["LLM_API_KEY"]
model = os.environ.get("LLM_MODEL", "claude-sonnet")

def summarize_note(text: str) -> str:
    payload = {
        "model": model,
        "messages": [
            {
                "role": "system",
                "content": (
                    "你是一个 Obsidian 知识库整理助手。"
                    "请输出摘要、标签、建议归档目录。"
                    "不要编造来源，不要删除原文。"
                )
            },
            {
                "role": "user",
                "content": text[:6000]
            }
        ],
        "temperature": 0.3
    }

    response = requests.post(
        f"{base_url}/chat/completions",
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json"
        },
        json=payload,
        timeout=60
    )
    response.raise_for_status()
    data = response.json()
    return data["choices"][0]["message"]["content"]

for path in inbox.glob("*.md"):
    text = path.read_text(encoding="utf-8")
    summary = summarize_note(text)
    out = path.with_name(path.stem + ".summary.md")
    out.write_text(summary, encoding="utf-8")
    print(f"done: {path.name}")
```

第一次运行时不要自动移动文件。

只生成 `.summary.md`，人工检查后再归档。

这才是安全的 AI 工作流。

## 8. 用 Claude Code 跑知识工作流

有了目录、规则和备份后，就可以开始让 Claude Code 做事。

### 工作流一：整理收件箱

```text
请先读取 CLAUDE.md 和 _agents/user-manual.md。
然后扫描 _inbox/ 下的 Markdown 文件。
请先输出一个归档建议表：
1. 文件名
2. 内容摘要
3. 建议归档目录
4. 建议标签
5. 是否需要我确认

先不要移动、删除、重命名任何文件。
```

确认后再说：

```text
按刚才的归档建议执行。
每移动一个文件都保留原文件名。
执行后输出变更清单。
```

### 工作流二：把素材变成永久笔记

```text
请读取 10_sources/ 下这篇资料。
把它拆成 3 到 5 条永久笔记，保存到 20_notes/。

要求：
1. 每条笔记只表达一个观点
2. 标注来源文件
3. 给出 3 个双向链接建议
4. 不要覆盖原始资料
```

### 工作流三：写公众号文章

```text
请先读取：
- CLAUDE.md
- _agents/user-manual.md
- _agents/ip-position.md

然后从 20_notes/ 和 30_projects/ 中搜索和“Obsidian + Claude Code 第二大脑”相关的资料。

先输出：
1. 可用素材清单
2. 文章角度
3. 标题备选
4. 大纲

先不要写正文。
```

大纲确认后再写正文。

不要让 AI 一上来就写完整文章。

先检索、再组织、再生成。

### 工作流四：周复盘

```text
请读取本周的 00_daily/ 笔记和 30_projects/ 项目笔记。
生成一份周复盘：
1. 本周完成了什么
2. 哪些任务卡住了
3. 哪些素材值得沉淀为永久笔记
4. 下周最重要的 3 件事
5. 建议归档或清理的文件

不要修改文件，先输出报告。
```

## 9. Dataview 做一个 AI 工作台首页

你可以在 Vault 根目录建一个 `首页.md`。

里面放几个 Dataview 区块。

### 未发布文章

```dataview
TABLE topic, platform, status
FROM "40_outputs"
WHERE status != "published"
SORT file.mtime DESC
```

### 待整理 inbox

```dataview
LIST
FROM "_inbox"
SORT file.mtime DESC
```

### 进行中的项目

```dataview
TABLE status, deadline
FROM "30_projects"
WHERE status != "done"
SORT deadline ASC
```

### 最近更新的永久笔记

```dataview
LIST
FROM "20_notes"
SORT file.mtime DESC
LIMIT 20
```

这样你打开 Obsidian，就能看到：

- 哪些素材没整理
- 哪些文章没发布
- 哪些项目在推进
- 哪些笔记最近更新

Claude Code 也可以基于这个首页理解你的工作状态。

## 10. 成本怎么控？不要把整个 Vault 都丢给模型

很多 AI 知识库工具最大的问题是成本失控。

Obsidian + Claude Code 的好处是你可以让模型按需读文件。

不要每次都把所有笔记塞进上下文。

建议策略：

1. 先用 Omnisearch 或 grep 找相关文件。
2. 再读取少量高相关笔记。
3. 长文件先摘要。
4. 批量任务先抽样测试。
5. 简单分类用便宜模型。
6. 深度写作再用强模型。
7. 记录每类任务的 token 用量。

如果你通过 4SAPI 这类中转站接入多模型，可以把任务分层：

| 任务 | 推荐模型策略 |
| --- | --- |
| inbox 粗分类 | 低成本模型 |
| 标题生成 | 中等模型，多候选 |
| 长文初稿 | Claude / GPT 强模型 |
| 事实核查 | 另一个模型交叉检查 |
| 代码或脚本 | 编程能力强的模型 |
| 周报摘要 | 低成本模型即可 |

这比所有任务都用最贵模型更合理。

## 11. 权限和隐私：第二大脑最怕失控

个人知识库往往比代码仓库更敏感。

里面可能有：

- 私人经历
- 客户信息
- 财务记录
- 账号线索
- 未发布内容
- 商业计划
- 合同和沟通记录

所以一定要给 AI 设边界。

建议写进 `CLAUDE.md`：

```md
## 安全红线

- 不读取 password、secret、key、token 字样的文件，除非我明确要求
- 不把私人经历发给外部服务做示例
- 不自动删除任何文件
- 不自动移动 `_agents/` 文件
- 不自动发布内容
- 不自动上传附件
- 批量处理前先列文件清单
```

另外，使用社区插件要注意：

- 看插件维护状态
- 看 GitHub issue
- 看权限说明
- 先在测试 Vault 试
- 重要数据先备份
- 不要给不熟悉插件过高权限

大模型 API 中转站也一样。

可以用于合法合规的模型接入、格式转换、成本优化、计费统计和多模型路由。

不要用于恶意绕过限制、处理违规内容、泄露隐私或绕开平台安全边界。

## 12. 最小可用版本：今天就能搭起来

如果你不想一次搭完整系统，可以按这个顺序来。

第一步，创建 Vault：

```text
my-brain
```

第二步，创建 3 个文件夹：

```text
_agents/
_inbox/
40_outputs/
```

第三步，创建 2 个文件：

```text
_agents/user-manual.md
CLAUDE.md
```

第四步，初始化 Git：

```bash
git init
git add .
git commit -m "init obsidian vault"
```

第五步，在 Vault 根目录运行 Claude Code：

```bash
claude
```

第六步，输入：

```text
请读取 CLAUDE.md 和 _agents/user-manual.md。
以后在这个 Vault 工作时，先遵守这些规则。
现在请只总结目录结构，不要修改文件。
```

第七步，开始第一个任务：

```text
请整理 _inbox/ 的内容。
先给归档建议，不要移动文件。
```

这样就够了。

先跑起来，再慢慢增加：

- Dataview
- Omnisearch
- Templater
- Obsidian Git
- Terminal
- Claudian
- API 批处理脚本

不要反过来。

## 13. 常见坑

第一，插件装太多。

插件越多，系统越脆。先用最小插件集。

第二，目录设计太复杂。

如果你不知道文件该放哪，说明目录太复杂。

第三，`CLAUDE.md` 写得太虚。

“帮我更高效”不如“改动 3 个以上文件先列清单”。

第四，让 AI 直接移动文件。

先出清单，再执行。

第五，不做版本管理。

只要 AI 能改文件，就必须有回滚。

第六，把原始资料和输出混在一起。

原始资料要保留，输出要单独存放。

第七，把整个 Vault 丢给模型。

先搜索，再读取，再生成。

第八，忽略插件兼容性。

社区插件要看版本、issue 和系统要求。

第九，过度相信 AI 归档。

AI 可以建议分类，但你的知识系统最终要符合你的思考方式。

第十，忘了第二大脑的核心是你自己。

工具只是帮你保存和连接，真正有价值的是你的经历、判断和观点。

## 14. 最后总结

Obsidian + Claude Code 的价值，不是把笔记软件变成聊天框。

它真正有价值的地方是：

```text
用 Obsidian 保存长期知识。
用 CLAUDE.md 固定协作规则。
用 Claude Code 读取和操作本地 Vault。
用插件增强检索、模板、备份和执行。
用大模型 API 中转站处理批量任务和多模型成本。
```

这样搭出来的系统，才不只是“第二大脑”的口号。

它能做几件实际的事：

- 帮你整理收件箱
- 帮你把素材沉淀成笔记
- 帮你从笔记生成文章
- 帮你复盘项目
- 帮你发现旧知识之间的新连接
- 帮你把重复工作变成流程

如果你是个人用户，先从 Obsidian + Claude Code + Git 开始。

如果你是开发者或团队，再把 4sapi.com 这类大模型 API 中转站接进批量处理、内容生成、知识库问答和成本统计。

一句话总结：

```text
第二大脑不是装出来的，而是积累出来的。
Obsidian 负责长期积累，Claude Code 负责协作执行，中转站负责把模型能力工程化。
```

## 资料来源与延伸阅读

- [How Obsidian stores data - Obsidian Help](https://obsidian.md/help/data-storage)
- [Internal links - Obsidian Help](https://obsidian.md/help/links)
- [Graph view - Obsidian Help](https://obsidian.md/help/plugins/graph)
- [How Claude remembers your project - Claude Code Docs](https://code.claude.com/docs/en/memory)
- [Extend Claude Code - Claude Code Docs](https://code.claude.com/docs/en/features-overview)
- [Explore the .claude directory - Claude Code Docs](https://code.claude.com/docs/en/claude-directory)
- [Dataview documentation](https://blacksmithgu.github.io/obsidian-dataview/)
- [Omnisearch for Obsidian](https://github.com/scambier/obsidian-omnisearch)
- [Templater - Obsidian Community Plugin](https://community.obsidian.md/plugins/templater-obsidian)
- [Obsidian Git Plugin](https://github.com/Vinzent03/obsidian-git)
- [Obsidian Terminal Plugin](https://github.com/polyipseity/obsidian-terminal)
- [Claudian - Obsidian plugin for coding agents](https://github.com/YishenTu/claudian)
