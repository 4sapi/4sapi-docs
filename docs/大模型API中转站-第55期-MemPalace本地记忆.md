---
title: "【大模型API中转站】第55期 MemPalace本地记忆 | 0 API搭建"
category: 人工智能
tags:
  - 大模型API中转站
  - MemPalace
  - MCP
  - Agent记忆
  - Codex
  - 4SAPI
description: "拆解 MemPalace 的本地优先 AI 记忆系统：用 init、mine、search 跑通项目和会话记忆，用 MCP 接入 Agent，再用 4SAPI 管外部模型、Key、成本和日志。"
---

# 【大模型API中转站】第55期 MemPalace本地记忆 | 0 API搭建

本文是【大模型API中转站】系列的第55篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

前几篇我们连续讲了几个 Agent 相关工具：

```text
PilotDeck：多 WorkSpace Agent 工作台
Karpathy Skills：编码 Agent 行为规范
个人知识库：本地资料、笔记和 RAG 基础结构
```

这一篇看 MemPalace。

它解决的是一个更底层的问题：

```text
Agent 怎么长期记住你们之前做过什么？
```

很多人用 Claude Code、Codex、Cursor 或其他 Agent 时，都会遇到同一个痛点：

```text
这个项目上周已经讨论过一次，今天它又不知道了。
这个 bug 昨天排查过，今天还要从头解释。
这段架构决策写在旧会话里，但新会话找不到。
上下文一压缩，很多细节就丢了。
换一个工具，记忆基本归零。
```

MemPalace 的定位就是给 AI 一个本地记忆层。

它不是普通聊天软件，也不是模型网关。

更准确地说：

```text
MemPalace 管长期记忆。
4SAPI 管模型调用。
Claude Code / Codex / Cursor / PilotDeck 管实际工作。
```

这篇文章会讲清楚：

```text
MemPalace 是什么
它能不能写进大模型API中转站系列
它和 4SAPI 是什么关系
怎么 0 API Key 跑通本地记忆
怎么接 MCP 给 Agent 用
什么时候才需要把 4SAPI 接进去
隐私和避坑点有哪些
```

## 1. 这个项目能不能写

能写，而且很适合写。

MemPalace 官方 README 对它的描述很明确：

```text
Local-first AI memory.
Verbatim storage.
Zero API calls for the core path.
```

翻成更实用的话：

```text
默认本地优先。
默认保存原文。
核心检索不需要 API Key。
```

这和很多“云端知识库”不一样。

MemPalace 不是先把你的资料上传到某个云服务，再让模型总结一遍。

它的核心设计是：

```text
把项目文件、会话记录、文档内容存成本地记忆。
用向量检索和结构化索引找回原文。
通过 MCP 或 CLI 给 Agent 提供上下文。
```

所以它非常适合写在本系列里。

但角度不能写错。

不要把它写成：

```text
MemPalace 接入 4SAPI 的聊天机器人教程。
```

更准确的标题应该是：

```text
MemPalace 本地记忆 + 4SAPI 模型网关：让 Agent 有长期上下文。
```

## 2. 它能不能接入 4SAPI

答案分两层。

第一层：

```text
MemPalace 的核心流程不需要接 4SAPI。
```

比如这些命令：

```bash
uv tool install mempalace
mempalace init ~/projects/myapp
mempalace mine ~/projects/myapp
mempalace search "why did we switch to GraphQL"
```

它们的核心能力是本地建库、本地挖掘、本地搜索。

不需要 4SAPI，也不需要 OpenAI Key。

第二层：

```text
MemPalace 的 LLM 辅助功能可以通过 OpenAI-compatible endpoint 接 4SAPI。
```

官方源码里支持三类 LLM provider：

| Provider | 用途 |
| --- | --- |
| `ollama` | 默认本地 LLM，走 `http://localhost:11434` |
| `openai-compat` | 任意 OpenAI-compatible `/v1/chat/completions` endpoint |
| `anthropic` | Anthropic Messages API |

所以，如果你想让 MemPalace 在 `init` 阶段用外部模型辅助实体识别，可以这样接 4SAPI：

```bash
mempalace init ~/projects/myapp \
  --llm-provider openai-compat \
  --llm-endpoint https://4sapi.com/v1 \
  --llm-api-key sk-your-4sapi-key \
  --llm-model your-model-name \
  --accept-external-llm
```

这里要注意隐私边界。

只要你配置的是外部 endpoint，MemPalace 就可能把用于实体识别的文件样本发送给外部模型。

所以它会给外部 LLM 提示和同意流程。

如果你只想完全本地，不要走外部模型：

```bash
mempalace init ~/projects/myapp --no-llm
```

一句话总结：

```text
MemPalace 核心不靠 4SAPI。
需要外部 LLM 辅助时，可以用 openai-compat 指向 4SAPI。
```

## 3. MemPalace 和普通知识库有什么不同

普通知识库通常是这样：

```text
文档
  -> 切片
  -> 向量化
  -> 搜索
  -> 回答
```

MemPalace 的表达更像一个“记忆宫殿”：

```text
wing
  -> room
  -> drawer
```

可以理解为：

| 概念 | 含义 |
| --- | --- |
| `wing` | 人、项目、Agent 或一个大的记忆分区 |
| `room` | 主题、文件夹、领域或上下文类别 |
| `drawer` | 原始内容片段，尽量保留原文 |

它不是把所有资料扔进一个扁平向量库。

而是尽量让记忆有结构。

好处是：

```text
你可以只查某个项目。
可以只查某个主题。
可以保留原文证据。
可以让新会话快速 wake-up。
```

这对 Agent 很重要。

因为 Agent 不是只要“答案”，还需要知道：

```text
这个答案来自哪个项目？
是哪一次讨论留下的？
原文到底怎么说？
现在是否还适用？
```

## 4. 适合哪些人使用

MemPalace 适合三类人。

第一类，长期用 Claude Code、Codex、Cursor 的开发者。

你可以把项目代码、文档、旧会话、排错记录都挖进去。

下次开新会话时，不用从头解释项目背景。

第二类，研究型创作者。

你可以把长文资料、论文笔记、采访记录、AI 对话记录放进去，用搜索快速找旧判断。

第三类，多 Agent 或团队工作流。

不同 Agent 可以有自己的 wing 和 diary，团队也可以按项目隔离记忆。

它不适合什么人？

```text
只偶尔问几句 AI 的轻度用户。
不愿意维护本地目录的人。
希望所有东西都自动云同步的人。
对命令行完全抗拒的人。
```

MemPalace 是工程味比较重的工具。

它的价值在长期积累，不在第一次打开就“特别丝滑”。

## 5. 安装方式

官方推荐用 `uv tool install`。

这样可以把 CLI 安装在隔离环境里，避免和系统 Python 依赖冲突。

```bash
uv tool install mempalace
```

检查命令：

```bash
mempalace --help
```

如果你不用 `uv`，也可以用 `pipx`：

```bash
pipx install mempalace
```

只有在你明确需要在项目里 `import mempalace` 时，才建议用虚拟环境里的 `pip`：

```bash
python -m venv .venv
source .venv/bin/activate
pip install mempalace
```

Windows PowerShell 可以这样激活虚拟环境：

```powershell
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install mempalace
```

要求也不复杂：

```text
Python 3.9+
一个向量存储后端，默认 ChromaDB
本地 embedding 模型缓存空间
```

官方 README 提到，本地 embedding 模型大约需要数百 MB 级别磁盘空间。

第一次安装和初始化时，耐心一点。

## 6. 三步跑通项目记忆

先用一个低风险项目测试。

不要第一次就挖公司全量仓库。

假设项目在：

```text
~/projects/myapp
```

第一步，初始化：

```bash
mempalace init ~/projects/myapp
```

如果你想完全本地，不启用 LLM 辅助：

```bash
mempalace init ~/projects/myapp --no-llm
```

第二步，挖掘项目文件：

```bash
mempalace mine ~/projects/myapp
```

第三步，搜索：

```bash
mempalace search "why did we switch to GraphQL"
```

也可以限制结果数量：

```bash
mempalace search "auth session expires" --results 10
```

如果项目已经按 wing / room 组织好，可以进一步缩小范围：

```bash
mempalace search "pricing discussion" --wing myapp --room costs
```

这就是最小闭环：

```text
init
  -> mine
  -> search
```

跑通这三步，说明 MemPalace 已经能存、能索引、能找回。

## 7. 挖掘 Claude Code / Codex / Cursor 会话

MemPalace 更有价值的地方，是能把 Agent 会话也变成记忆。

项目文件只能说明“代码现在是什么样”。

会话记录能说明：

```text
为什么这么改？
当时排除了哪些方案？
哪个 bug 已经查过？
哪些风险还没解决？
```

挖掘 Claude Code 会话可以用：

```bash
mempalace mine ~/.claude/projects/ --mode convos
```

如果只想归到某个项目 wing：

```bash
mempalace mine ~/.claude/projects/-Users-you-Projects-myapp \
  --mode convos \
  --wing myapp
```

会话来源不只 Claude Code。

官方 CLI 文档里，`mine` 支持三种模式：

| 模式 | 用途 |
| --- | --- |
| `projects` | 默认模式，挖代码、文档、笔记 |
| `convos` | 挖聊天和 Agent 会话导出 |
| `extract` | 挖 PDF、DOCX、PPTX、XLSX、RTF、EPUB 等二进制文档 |

如果要挖 PDF、Office 文档，需要安装 extra：

```bash
uv tool install "mempalace[extract]"
```

或者在虚拟环境里：

```bash
pip install "mempalace[extract]"
```

## 8. 用 wake-up 给新会话加载背景

Agent 最怕新会话失忆。

MemPalace 提供了 `wake-up`：

```bash
mempalace wake-up
```

只加载某个项目：

```bash
mempalace wake-up --wing myapp
```

你可以把输出复制给新会话，让 Agent 快速知道：

```text
当前项目是什么
最近做过什么
关键决策有哪些
哪些问题还没有结束
```

如果配合 MCP，Agent 可以自己调用 MemPalace 工具，不用你手动复制。

这就是它和普通笔记库的区别。

普通笔记库是给人看的。

MemPalace 更像给 Agent 准备的长期上下文。

## 9. 接入 MCP

MemPalace 提供 MCP server。

官方 README 里写到，MCP 工具覆盖 palace 读写、知识图谱、跨 wing 导航、drawer 管理和 agent diaries 等能力。

先让 CLI 输出 MCP 配置提示：

```bash
mempalace mcp
```

如果用 Docker，也可以把它作为 stdio MCP server：

```json
{
  "mcpServers": {
    "mempalace": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-v", "mempalace-data:/data", "mempalace"]
    }
  }
}
```

注意 `-i` 很重要。

MCP stdio 需要标准输入输出通信。

如果你用本地 CLI，也可以按 `mempalace mcp` 给出的命令配置到 Claude Code、Cursor 或其他 MCP 客户端。

接上 MCP 后，工作流变成：

```text
Agent 接到任务
  -> 查询 MemPalace 旧记忆
  -> 找到相关 drawer
  -> 基于原文继续工作
  -> 任务结束后保存新会话
```

这比每次手动复制旧聊天记录稳定得多。

## 10. Auto-save hooks：别让会话过期才想起来

官方 README 特别提醒：

```text
Claude Code sessions expire in 30 days without auto-save hooks wired.
```

所以如果你真的想用 MemPalace 管长期记忆，hooks 很重要。

官方已经提供了 Claude Code、Codex CLI 和 Cursor IDE 的自动保存 hooks。

它们大致解决三件事：

```text
定期保存会话
上下文压缩前保存快照
新会话开始时召回相关记忆
```

建议顺序是：

```text
先手动 mine 现有会话
再配置 hooks 自动保存新会话
最后用 MCP 让 Agent 自己检索旧记忆
```

不要等会话过期、上下文丢了，再想起来补救。

## 11. 和 4SAPI 的正确组合方式

MemPalace 和 4SAPI 不在同一层。

更清楚的结构是：

```text
MemPalace：长期记忆层
4SAPI：模型 API 网关
Agent 工具：Claude Code / Codex / Cursor / PilotDeck
```

组合之后是：

```text
Agent 读取当前项目
  -> 通过 MCP 查询 MemPalace
  -> 拿到旧会话、旧决策、相关资料
  -> 通过 4SAPI 调用合适模型
  -> 执行代码、写文档或做分析
  -> 任务结束后把新结果保存回 MemPalace
```

这套结构的好处是职责清楚。

| 层级 | 负责什么 |
| --- | --- |
| MemPalace | 记忆、检索、会话保存、知识图谱 |
| 4SAPI | Base URL、API Key、模型路由、成本日志 |
| Agent | 读写文件、调用工具、运行命令、生成产物 |
| 项目规则 | 安全边界、输出格式、测试要求 |

不要把 MemPalace 当成模型网关。

也不要把 4SAPI 当成记忆库。

它们组合起来，才适合长期 Agent 工作。

## 12. LLM 辅助功能接 4SAPI

MemPalace 默认会尝试本地 Ollama。

如果本地 Ollama 不可用，它会走启发式逻辑，不会因为没有 LLM 就完全不能用。

如果你明确想用 4SAPI 做 LLM 辅助实体识别，可以这样写：

```bash
mempalace init ~/projects/myapp \
  --llm-provider openai-compat \
  --llm-endpoint https://4sapi.com/v1 \
  --llm-api-key sk-your-4sapi-key \
  --llm-model your-model-name \
  --accept-external-llm
```

如果不想把 Key 写进命令，也可以用环境变量：

```bash
export OPENAI_API_KEY=sk-your-4sapi-key
mempalace init ~/projects/myapp \
  --llm-provider openai-compat \
  --llm-endpoint https://4sapi.com/v1 \
  --llm-model your-model-name
```

Windows PowerShell：

```powershell
$env:OPENAI_API_KEY="sk-your-4sapi-key"
mempalace init C:\projects\myapp `
  --llm-provider openai-compat `
  --llm-endpoint https://4sapi.com/v1 `
  --llm-model your-model-name
```

但我更建议在教程里优先展示显式参数。

原因是 MemPalace 对环境变量里的外部 API Key 会更谨慎，可能触发确认流程。

显式传参更能表达：

```text
我知道这次会调用外部 LLM。
```

模型选择上，不需要一上来用最贵模型。

LLM 辅助实体识别可以先用中低成本模型。

复杂总结、会话理解和最终输出，再交给 Agent 侧的强模型。

## 13. 存储后端怎么选

MemPalace 默认使用 ChromaDB。

对个人用户来说，默认就够了。

官方还提供了其他后端：

| 后端 | 适合场景 |
| --- | --- |
| ChromaDB | 默认本地使用 |
| `sqlite_exact` | 本地 exact-vector 正确性检查 |
| Qdrant | 自托管向量服务或团队化部署 |
| pgvector | PostgreSQL 体系内统一管理 |

示例：

```bash
mempalace mine ~/projects/myapp --backend sqlite_exact
```

Qdrant：

```bash
MEMPALACE_QDRANT_URL=http://localhost:6333 \
  mempalace mine ~/projects/myapp --backend qdrant
```

pgvector：

```bash
MEMPALACE_PGVECTOR_DSN=postgresql://localhost:5432/mempalace \
  mempalace mine ~/projects/myapp --backend pgvector
```

注意：

```text
Qdrant 和 pgvector 如果指向外部服务，会发送并保存原文 drawer 和 metadata。
```

本地优先不等于你怎么配都本地。

只要你把后端指向外部服务，就要按外部存储处理隐私风险。

## 14. 推荐模型分层

如果你用 4SAPI 管 Agent 模型，可以这样分层：

| 任务 | 推荐策略 |
| --- | --- |
| MemPalace 核心搜索 | 不需要 LLM |
| `init` 实体识别辅助 | 中低成本模型 |
| 旧会话摘要 | 中等文本模型 |
| 复杂架构决策回顾 | 强推理模型 |
| 代码修改 | 强代码模型 |
| PR Review | 强推理模型 |
| 任务结束总结 | 低成本到中等模型 |

这套组合能省钱的关键是：

```text
不要让强模型做所有事情。
```

MemPalace 已经能本地找回原文。

很多时候，强模型只需要读少量召回片段，而不是从几十万字历史里硬找答案。

这就是记忆层的价值。

它把模型调用从：

```text
把所有历史塞给模型
```

变成：

```text
先本地召回，再把相关片段交给模型
```

成本和稳定性都会好很多。

## 15. 最小落地方案

如果你今天只想跑通，不想研究全部功能，按这个做：

```text
第 1 步：安装 MemPalace
第 2 步：选一个低风险项目
第 3 步：init 初始化
第 4 步：mine 项目文件
第 5 步：search 搜一个真实问题
第 6 步：mine 旧 Agent 会话
第 7 步：配置 MCP 给 Claude Code / Codex / Cursor
第 8 步：再考虑 4SAPI 外部 LLM 辅助
```

对应命令：

```bash
uv tool install mempalace
mempalace init ~/projects/myapp --no-llm
mempalace mine ~/projects/myapp
mempalace search "where is auth handled"
mempalace mine ~/.claude/projects/ --mode convos --wing myapp
mempalace wake-up --wing myapp
mempalace mcp
```

如果确认要接 4SAPI：

```bash
mempalace init ~/projects/myapp \
  --llm-provider openai-compat \
  --llm-endpoint https://4sapi.com/v1 \
  --llm-api-key sk-your-4sapi-key \
  --llm-model your-model-name \
  --accept-external-llm
```

最小安全规则：

```text
先用 --no-llm 跑本地流程。
不要第一次就挖敏感仓库。
不要把 API Key 写进 Git。
外部 LLM 和外部后端都要当作数据外发。
确认 MCP 客户端只连接官方 MemPalace server。
```

## 16. 常见坑

### 坑一：把 MemPalace 当聊天机器人

MemPalace 不是聊天界面。

它是记忆层。

你仍然需要 Claude Code、Codex、Cursor、PilotDeck 或其他 Agent 来实际工作。

### 坑二：以为所有流程都需要 API Key

核心路径不需要 API Key。

本地搜索和默认检索可以先跑起来。

外部 LLM 只是辅助，不是必需条件。

### 坑三：忽视官方仿冒站提醒

官方 README 明确提醒，MemPalace 只有三个官方来源：

```text
GitHub 仓库
PyPI 包
mempalaceofficial.com 文档
```

其他相似域名可能是假冒站点。

安装和查看文档时，不要随便点搜索结果里的非官方域名。

### 坑四：把敏感资料挖进去后又接外部服务

本地挖掘相对可控。

但如果你后面配置了：

```text
外部 LLM
外部 Qdrant
外部 pgvector
云端同步目录
```

数据边界就变了。

要重新评估隐私和合规。

### 坑五：只挖项目文件，不挖会话

代码告诉你现在是什么。

会话告诉你为什么变成这样。

长期 Agent 记忆里，旧会话往往比文件本身更有价值。

## 17. 成本与风险提示

MemPalace 的核心成本主要是本地资源：

```text
磁盘空间
embedding 模型缓存
向量库索引
本地 CPU / 内存
```

接 4SAPI 后，成本来自外部模型调用。

建议按用途拆 Key：

```text
mempalace-init
agent-memory-search
agent-code-review
agent-session-summary
```

建议按任务分模型：

```text
实体识别：中低成本模型
摘要整理：中等模型
复杂判断：强推理模型
代码修改：强代码模型
最终审核：强模型 + 人工
```

隐私边界也要写进项目规则：

```text
客户资料、合同、生产 Key、内部财务数据，不进入外部 LLM。
外部后端只用于低敏或脱敏资料。
MCP 客户端只连接可信 server。
所有 API Key 不进 Git。
```

AI 记忆越长期，越要重视清理和权限。

记住错东西，比忘记东西更麻烦。

## 18. 总结

MemPalace 值得写进这个系列。

但它不是模型中转站，也不是普通 RAG 项目。

它更像 Agent 的长期记忆层：

```text
存原文
做索引
能搜索
接 MCP
保存会话
帮助新会话 wake-up
```

4SAPI 则负责模型调用层：

```text
统一 Base URL
统一 API Key
统一模型池
统一日志和成本
```

两者配合后的结构是：

```text
MemPalace 管记忆。
4SAPI 管模型。
Agent 管行动。
项目规则管边界。
```

一句话总结：

```text
如果个人知识库是给人看的资料库，MemPalace 就是给 Agent 用的长期记忆；接上 4SAPI 后，记忆召回和模型调用可以分层管理，既省上下文，也省成本。
```

建议先从一个非敏感项目试。

先用 `--no-llm` 跑通本地 `init / mine / search`，再接 MCP，最后再决定是否用 4SAPI 做外部 LLM 辅助。

这样最稳。

## 参考资料

- MemPalace GitHub：https://github.com/MemPalace/mempalace
- MemPalace 官方文档：https://mempalaceofficial.com/
- Getting Started：https://mempalaceofficial.com/guide/getting-started.html
- CLI Reference：https://mempalaceofficial.com/reference/cli.html
- MCP Tools Reference：https://mempalaceofficial.com/reference/mcp-tools.html
- Hooks Guide：https://mempalaceofficial.com/guide/hooks.html
- Cursor Hooks：https://mempalaceofficial.com/guide/cursor-hooks.html
- Claude Code Retention Checklist：https://mempalaceofficial.com/guide/claude-code-retention.html
- 4SAPI API 地址：https://4sapi.com/v1
