---
title: "【大模型API中转站】第58期 Feynman接入4SAPI | 研究Agent省钱"
category: 人工智能
tags:
  - 大模型API中转站
  - Feynman
  - AI研究Agent
  - 论文调研
  - Deep Research
  - 4SAPI
  - OpenAI-Compatible
description: "基于 companion-inc/feynman 官方 README 和 setup 文档，拆解 Feynman 研究 Agent 的安装、模型配置、自定义 OpenAI-compatible provider、4SAPI 接入方式、文献综述工作流、成本治理和避坑清单。"
---

# 【大模型API中转站】第58期 Feynman接入4SAPI | 研究Agent省钱

本文是【大模型API中转站】系列的第58篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

前两篇我们讲了企业微信 CLI：

```text
第56期：企业微信 CLI 文档 Agent
第57期：企业微信 CLI 消息到智能表格闭环
```

那两篇的重点是办公系统。

这一篇换到研究场景：Feynman。

如果说企业微信 CLI 解决的是：

```text
Agent 怎么进入企业办公现场？
```

那 Feynman 解决的是：

```text
Agent 怎么做论文调研、文献综述、实验复现和研究草稿？
```

这类任务最容易烧 Token。

一篇 Deep Research 可能要搜论文、读摘要、看代码仓库、整理引用、做对比矩阵、再让多个 Agent 互相审稿。

如果全部直连最贵模型，成本很容易飙上去。

所以这篇的目标很明确：

```text
用 Feynman 做研究工作流。
用 4SAPI 统一模型入口、Key、成本和日志。
用 OpenAI-compatible 自定义 provider 把二者接起来。
```

## 1. Feynman 是什么

Feynman 的官方 README 对它的定位很短：

```text
The open source AI research agent.
```

翻成更实用的话：

```text
Feynman 是一个面向论文、技术资料、代码仓库和实验复现的 AI 研究 Agent。
```

它不是普通聊天工具。

它更像一个研究型命令行工作台。

你可以这样用：

```bash
feynman "what do we know about scaling laws"
```

它会搜索论文和网页，生成带引用的研究简报。

也可以这样用：

```bash
feynman deepresearch "mechanistic interpretability"
```

它会做多 Agent 深度调研。

官方 README 里列出的典型任务包括：

| 命令 | 用途 |
| --- | --- |
| `feynman "topic"` | 搜论文和网页，生成带引用研究简报 |
| `feynman deepresearch "topic"` | 多 Agent 深度调研 |
| `feynman lit "topic"` | 文献综述，整理共识、分歧、开放问题 |
| `feynman audit 2401.12345` | 对比论文声明和公开代码库 |
| `feynman replicate "paper"` | 设计本地或云 GPU 复现实验 |
| `feynman recipe "task"` | 查找可执行训练方案、数据集、代码和验证路径 |

它内置 4 类研究 Agent：

| Agent | 负责什么 |
| --- | --- |
| Researcher | 搜集论文、网页、仓库、文档证据 |
| Reviewer | 模拟审稿，给出严重程度和修改建议 |
| Writer | 把研究笔记整理成结构化草稿 |
| Verifier | 检查引用、来源 URL、死链和证据可靠性 |

这就决定了它很适合写进本系列。

因为它的模型调用不是一问一答，而是一条研究流水线。

## 2. 为什么 Feynman 适合接 4SAPI

普通聊天工具的调用链可能是：

```text
用户问一句
模型答一句
```

Feynman 的研究链路更像：

```text
提出主题
-> 搜论文
-> 搜网页
-> 看仓库
-> 整理证据
-> 生成文献综述
-> 模拟审稿
-> 校验引用
-> 生成草稿
-> 再修订
```

这一套跑下来，模型调用会很多。

而且不同环节对模型能力的要求不一样：

| 环节 | 模型要求 |
| --- | --- |
| 关键词扩展 | 便宜模型即可 |
| 摘要分类 | 便宜或中档模型 |
| 论文主张归纳 | 中档模型 |
| 争议点判断 | 强模型更稳 |
| 审稿式评估 | 强模型更稳 |
| 引用格式修复 | 便宜模型即可 |

如果 Feynman 全程只用一个昂贵模型，成本会很不舒服。

4SAPI 放在中间的价值是：

```text
统一 API Key
统一 Base URL
统一模型路由
统一成本统计
统一调用日志
必要时切换不同模型
```

所以更准确的架构是：

```text
Feynman 研究 Agent
  -> Custom provider
  -> 4SAPI OpenAI-compatible /v1
  -> Claude / GPT / Gemini / GLM / Qwen / DeepSeek 等模型
```

这不是绕过官方限制。

这是把 Feynman 的模型调用接到一个 OpenAI-compatible 模型网关，方便配置、审计和成本治理。

## 3. 先说结论：能不能接 4SAPI

可以接。

前提是你的 4SAPI 提供 OpenAI-compatible 接口。

Feynman 官方 README 和 setup 文档都提到：

```text
Custom provider (baseUrl + API key)
API mode: openai-completions
Base URL 指向 /v1 endpoint
```

所以接 4SAPI 的关键配置就是：

```text
Provider: Custom provider (baseUrl + API key)
API mode: openai-completions
Base URL: https://你的4SAPI域名/v1
Authorization header: Yes
API key: 你的4SAPI Key
Model ids: 4SAPI 后台可用模型名
```

注意 Base URL 填到 `/v1`。

不要填成：

```text
https://你的4SAPI域名/v1/chat/completions
```

这类完整接口路径一般是 SDK 或客户端内部拼出来的。

Feynman 需要的是 OpenAI-compatible 的根地址。

## 4. 安装方式：不要写错

Feynman 的安装方式不是 `npm install -g feynman`。

我查了公开 NPM 包：

```text
@companion-inc/feynman：未找到公开包
feynman-ai：未找到公开包
```

仓库里的 `package.json` 版本是 `0.3.3`，但官方 README 推荐的是一键安装脚本。

macOS / Linux：

```bash
curl -fsSL https://feynman.is/install | bash
```

Windows PowerShell：

```powershell
irm https://feynman.is/install.ps1 | iex
```

官方说明里还有两个重要点。

第一，安装脚本会下载 standalone native bundle，里面带自己的 Node.js runtime。

第二，如果要固定版本，可以传版本号。

例如：

```bash
curl -fsSL https://feynman.is/install | bash -s -- 0.2.35
```

升级 standalone app 时，重新跑安装脚本。

`feynman update` 只刷新 Feynman 环境里安装的 Pi packages，不替换 standalone runtime bundle。

这点很容易写错。

## 5. Skills-only 安装

Feynman 还有一个很适合 Codex 用户的安装方式：

```text
只安装研究 Skills，不安装完整终端应用。
```

macOS / Linux：

```bash
curl -fsSL https://feynman.is/install-skills | bash
```

Windows PowerShell：

```powershell
irm https://feynman.is/install-skills.ps1 | iex
```

这会把 skill library 安装到：

```text
~/.codex/skills/feynman
```

如果要明确指定 Codex：

```bash
curl -fsSL https://feynman.is/install-skills | bash -s -- --codex
```

Windows：

```powershell
& ([scriptblock]::Create((irm https://feynman.is/install-skills.ps1))) -Scope Codex
```

如果想装到当前仓库的 `.agents/skills/feynman`：

```bash
curl -fsSL https://feynman.is/install-skills | bash -s -- --repo
```

Windows：

```powershell
& ([scriptblock]::Create((irm https://feynman.is/install-skills.ps1))) -Scope Repo
```

这条路线适合什么人？

```text
已经主要用 Codex / Claude Code / OpenCode 做工作流
只想复用 Feynman 的研究 prompt、skills 和 guidance
不想安装完整 Feynman terminal
```

但如果你的目标是完整跑 Feynman 命令行，还是装 standalone app。

## 6. 第一次 setup

安装后运行：

```bash
feynman setup
```

官方 setup 文档里说，向导会走三段：

```text
模型配置
认证
可选 package 安装
```

如果你选择官方模型 provider，它会要求对应 API key 或 OAuth。

如果你要接 4SAPI，就不要走单一官方 provider。

选择：

```text
Custom provider (baseUrl + API key)
```

然后按 4SAPI 填：

```text
API mode: openai-completions
Base URL: https://api.你的4sapi域名/v1
Authorization header: Yes
API key: sk-你的4SAPIKey
Model ids: 你在4SAPI后台启用的模型名
```

保存后运行：

```bash
feynman model list
```

再设置默认模型：

```bash
feynman model set <provider>/<model-id>
```

或按文档支持的另一种写法：

```bash
feynman model set <provider:model-id>
```

这里的 `<provider>` 是你 setup 时保存的 provider 名。

`<model-id>` 要和 4SAPI 后台实际支持的模型名一致。

## 7. 4SAPI 接入示例

假设你的 4SAPI OpenAI-compatible 地址是：

```text
https://api.4sapi.example.com/v1
```

模型名是：

```text
gpt-4o
claude-sonnet-4
glm-4.5
qwen3-coder
deepseek-r1
```

那么 Feynman setup 里可以这样填：

```text
Provider type: Custom provider (baseUrl + API key)
API mode: openai-completions
Base URL: https://api.4sapi.example.com/v1
Authorization header: Yes
API key: sk-xxxx
Model ids: gpt-4o, claude-sonnet-4, glm-4.5, qwen3-coder, deepseek-r1
```

设置默认模型时：

```bash
feynman model list
feynman model set custom/claude-sonnet-4
```

如果你的 provider 名不是 `custom`，以 `model list` 输出为准。

测试一个轻量任务：

```bash
feynman "summarize recent work on retrieval augmented generation in 5 bullets"
```

如果能返回带来源的简报，就说明基本链路通了。

## 8. 如果 /models 拉取失败怎么办

Feynman 对 LM Studio 和 LiteLLM 会尝试读取 `/models` 端点并预填模型。

对于自定义 provider，也可能遇到：

```text
/models 返回格式不兼容
网关不开放 /models
模型名列表为空
模型 id 和 4SAPI 后台显示不一致
```

这不一定代表不能用。

处理顺序：

```text
先确认 Base URL 只填到 /v1
确认 Authorization header 选 Yes
确认 API key 正确
确认 Model ids 手动填写
用一个简单 prompt 测试
再看 4SAPI 后台日志
```

如果仍然不兼容，可以加一层 LiteLLM Proxy：

```text
Feynman
  -> LiteLLM Proxy http://localhost:4000/v1
  -> 4SAPI
  -> 上游模型
```

这种方案更麻烦，但好处是 LiteLLM 对 OpenAI-compatible 客户端的适配更成熟。

官方 setup 文档也把 LiteLLM Proxy 单独列为支持项，默认：

```text
Base URL: http://localhost:4000/v1
API mode: openai-completions
```

## 9. 研究工作流怎么跑

接通模型后，先别一上来就跑最重的 `deepresearch`。

建议按成本从低到高测试。

### 9.1 普通研究简报

```bash
feynman "what do we know about scaling laws for language models"
```

适合验证：

```text
模型是否能调用
web / paper search 是否正常
引用是否生成
中文终端显示是否正常
```

### 9.2 文献综述

```bash
feynman lit "RLHF alternatives"
```

适合做文章或研究笔记：

```text
主流方法
关键论文
共识
分歧
开放问题
后续阅读路线
```

这个模式比普通问答更适合写“技术调研”。

### 9.3 深度研究

```bash
feynman deepresearch "mechanistic interpretability"
```

这个会更耗 Token。

建议先把默认模型设成成本可控的中档模型，再在关键阶段切换强模型。

如果 4SAPI 支持按模型分组或路由，可以这样规划：

| 任务 | 推荐模型 |
| --- | --- |
| 初筛资料 | 便宜模型 |
| 论文主张归纳 | 中档模型 |
| 方法论争议判断 | 强模型 |
| 审稿式批评 | 强模型 |
| 参考文献格式修正 | 便宜模型 |

### 9.4 论文代码审计

```bash
feynman audit 2401.12345
```

适合判断：

```text
论文里的 claim 是否有代码支撑
README 和论文实验是否一致
训练脚本是否公开
数据处理是否可复现
指标和论文表格是否对应
```

这类任务对模型要求高，但价值也高。

### 9.5 实验复现方案

```bash
feynman replicate "chain-of-thought improves math"
```

适合输出：

```text
复现实验步骤
数据集
训练或推理脚本
评估指标
硬件需求
失败风险
替代方案
```

如果你要真的跑实验，还会涉及 Docker、Modal、RunPod、本地 GPU。

这部分建议另写一篇，不要和模型接入混在一起。

## 10. Feynman 和 Codex 怎么分工

Feynman 和 Codex 都是 Agent，但适合的任务不一样。

| 工具 | 更适合做什么 |
| --- | --- |
| Feynman | 论文搜索、文献综述、研究简报、审稿、实验复现方案 |
| Codex | 本地项目读写、代码修改、脚本执行、博客落稿、文件整理 |
| 4SAPI | 模型入口、Key 管理、成本统计、日志审计 |

一个很实用的组合是：

```text
Feynman 做资料研究
-> 生成带引用的研究简报
-> Codex 根据简报写博客或改项目文档
-> 4SAPI 统一承接两边模型调用
```

如果只装 Skills-only，也可以让 Codex 直接使用 Feynman 的 research skills。

这适合内容创作者：

```text
Feynman research skills 负责找资料。
Codex 负责按你的发文规范写文章。
4SAPI 负责统一模型成本。
```

## 11. 成本治理：别让 deepresearch 失控

Feynman 的强项也是它的成本风险。

它会主动调研、并行研究、整理引用、反复验证。

这些动作都会产生模型调用。

建议做 5 件事。

第一，先用小任务测试。

```bash
feynman "summarize RAG in 5 bullets"
```

确认链路正常后，再跑 `lit` 或 `deepresearch`。

第二，把默认模型设为中档模型。

不要默认就用最贵模型跑所有资料初筛。

第三，在 4SAPI 后台看日志。

关注：

```text
请求次数
输入 Token
输出 Token
失败重试
哪个模型花费最多
```

第四，把任务拆短。

不要一次问：

```text
帮我研究所有 AI Agent 的历史、现状、架构、论文、产品和未来。
```

更好的拆法：

```text
先研究 2024-2026 代码 Agent 架构。
再研究 SWE-bench 相关方法。
再研究多 Agent 协作框架。
最后让 Feynman compare 做对比矩阵。
```

第五，给输出范围。

例如：

```text
只需要 10 篇最关键论文。
只列可复现实验。
只关注开源代码。
输出 1500 字以内。
```

研究 Agent 需要边界。

边界越清楚，成本越可控。

## 12. 数据和合规提醒

Feynman 会处理论文、网页、仓库、数据集和本地文件。

合法合规边界要写清楚。

可以做：

```text
公开论文调研
开源仓库分析
公开文档整理
复现实验方案设计
公开数据集元数据检查
```

谨慎做：

```text
未公开论文手稿
公司内部实验数据
客户项目代码
私有模型权重
带个人信息的数据集
```

接 4SAPI 后，要明确：

```text
哪些内容会进入模型 API
是否允许出域
是否需要脱敏
日志保留多久
哪些任务必须用本地模型
```

这篇文章不讨论任何绕过论文付费墙、绕过 API 限制、抓取私有资料的做法。

Feynman 的价值是把公开资料和授权资料研究得更系统，不是绕开合规边界。

## 13. 常见问题

### 13.1 Feynman 能不能直接接 4SAPI

可以。

用：

```text
Custom provider (baseUrl + API key)
API mode: openai-completions
Base URL: https://你的4SAPI域名/v1
Authorization header: Yes
```

### 13.2 需要填 `/chat/completions` 吗

通常不要。

填 `/v1`。

### 13.3 模型列表为空怎么办

手动填 Model ids。

并检查：

```text
4SAPI 后台模型名
API key 权限
Authorization header
Base URL
4SAPI 调用日志
```

### 13.4 可以用 LiteLLM 中转再接 4SAPI 吗

可以。

如果 Feynman 和 4SAPI 的 OpenAI-compatible 细节不完全匹配，LiteLLM Proxy 是一个缓冲层。

结构：

```text
Feynman -> LiteLLM Proxy -> 4SAPI -> 上游模型
```

### 13.5 Feynman 是不是必须装完整终端

不一定。

如果只想让 Codex 使用研究技能，可以装 Skills-only。

完整 Feynman terminal 更适合长期研究工作台。

### 13.6 为什么不建议 npm install

官方 README 推荐一键安装脚本。

公开 NPM 上我没有查到可直接安装的 `@companion-inc/feynman` 或 `feynman-ai` 包。

所以教程里不要写不存在的 NPM 安装方式。

## 14. 最小接入清单

给读者一个最小 checklist。

### 第一步：安装 Feynman

Windows：

```powershell
irm https://feynman.is/install.ps1 | iex
```

macOS / Linux：

```bash
curl -fsSL https://feynman.is/install | bash
```

### 第二步：运行 setup

```bash
feynman setup
```

### 第三步：选择自定义 provider

```text
Custom provider (baseUrl + API key)
```

### 第四步：填写 4SAPI

```text
API mode: openai-completions
Base URL: https://你的4SAPI域名/v1
Authorization header: Yes
API key: 你的4SAPI Key
Model ids: 你的4SAPI模型名
```

### 第五步：确认模型

```bash
feynman model list
feynman model set <provider>/<model-id>
```

### 第六步：跑一个轻量任务

```bash
feynman "summarize retrieval augmented generation in 5 bullets"
```

### 第七步：看 4SAPI 日志

确认：

```text
请求到了 4SAPI
模型名正确
Token 消耗可接受
没有大量失败重试
```

跑通这 7 步，再开始写 `lit`、`deepresearch`、`audit`、`replicate`。

## 15. 总结

Feynman 很适合写进【大模型API中转站】系列。

它不是普通聊天工具，而是研究 Agent：

```text
搜论文
做文献综述
审稿式评估
论文代码核查
实验复现方案
研究草稿生成
```

它也很适合接 4SAPI。

因为 Feynman 的研究流会产生大量模型调用，而 4SAPI 可以把模型入口、Key、成本和日志统一起来。

最推荐的接入方式是：

```text
Feynman setup
-> Custom provider (baseUrl + API key)
-> API mode: openai-completions
-> Base URL: 4SAPI /v1
-> Model ids: 4SAPI 可用模型
```

一句话总结：

```text
Feynman 负责研究流程，4SAPI 负责模型网关。
```

如果你经常写技术调研、论文综述、模型选型、实验复现计划，这个组合非常值得试。

下一步可以继续拆一篇：

```text
Feynman 文献综述实战：从论文搜索到公众号长文
```

那篇就可以拿一个具体主题，跑完整的 `lit` 或 `deepresearch` 工作流。

## 资料来源

- Feynman GitHub 仓库：<https://github.com/companion-inc/feynman>
- Feynman README：<https://raw.githubusercontent.com/companion-inc/feynman/main/README.md>
- Feynman Setup 文档：<https://www.feynman.is/docs/getting-started/setup>
- Feynman Installation 文档：<https://www.feynman.is/docs/getting-started/installation>
