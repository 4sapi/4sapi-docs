---
title: "DeepSeek V4-Flash 接入 Codex：官方原生配置与回滚"
category: 人工智能
tags:
  - Codex
  - DeepSeek API
  - 模型配置
description: "从官方脚本、手动编辑 config.toml 到模型目录、API Key 和回滚验证，整理 DeepSeek V4-Flash 原生接入 Codex 的完整流程。"
---

# DeepSeek V4-Flash 接入 Codex：官方原生配置与回滚

以前把 DeepSeek 接进 Codex，常见做法是在本机再跑一层代理，把 Chat Completions 翻译成 Codex 需要的 Responses 协议。现在 DeepSeek 官方文档已经提供了 Codex 接入方式，接入重点从“怎么翻译协议”变成了“模型目录、提供方、认证和回滚是否配置完整”。

但这不等于改一行 `base_url` 就结束。Codex 需要知道当前模型使用什么协议、有哪些能力、允许哪些推理档位，以及 API Key 从哪里来。本文按 2026-08-05 可访问的 DeepSeek 官方 Codex 文档整理，给出官方脚本和手动配置两条路径，并补上安全、验证和回滚步骤。具体模型可用性和配置字段仍应以当前官方页面为准。

## 1. 先确认当前接入边界

DeepSeek 官方 Codex 接入页当前将 `deepseek-v4-flash` 作为 Codex 接入模型示例，并要求使用 Responses 协议。不要因为桌面端模型菜单里出现了其他模型，就默认它们都能完成 Codex 工具循环。

接入前先确认四件事：

```text
模型名称：deepseek-v4-flash
协议：Responses
认证：DeepSeek API Key
配置位置：用户级 Codex 配置，而不是项目仓库里的随意文件
```

OpenAI 的 Codex 配置参考说明，用户级配置位于 `~/.codex/config.toml`；项目级 `.codex/config.toml` 需要项目受信任，并且不能覆盖机器级的模型提供方和认证相关字段。可以参考 [Codex config.toml 官方文档](https://learn.chatgpt.com/docs/config-file/config-reference#configtoml) 核对当前字段边界。

如果当前官方文档仍未声明某个模型支持 Codex 所需协议，就不要只凭模型选择器里的名称进行尝试。出现 404、协议错误或工具循环失败时，优先回到模型、协议和模型目录检查。

## 2. 公共准备工作

两种配置方式都需要先准备：

```text
Codex CLI 或桌面端至少成功启动过一次。
用户目录下已经生成 .codex。
准备一个可以随时回滚的测试仓库。
准备 DeepSeek API Key，不把 Key 写进项目文件。
备份整个 ~/.codex 目录，包括 config.toml、models.json 和认证文件。
彻底退出 Codex 桌面端以及可能管理配置的第三方切换工具。
```

备份不要只复制 `config.toml`。认证文件、模型目录、MCP 配置、项目受信任状态和历史配置都可能影响启动结果。Windows 用户通常需要将 `~` 替换为当前用户目录，例如 `C:\Users\<用户名>\.codex`；实际目录以当前环境为准。

## 3. 方式一：运行官方配置脚本

DeepSeek 官方文档提供了 macOS/Linux 和 Windows PowerShell 的配置脚本。命令形式如下：

macOS/Linux：

```bash
bash <(curl -fsSL https://cdn.deepseek.com/api-docs/codex-deepseek-setup.sh)
```

Windows PowerShell：

```powershell
irm https://cdn.deepseek.com/api-docs/codex-deepseek-setup-en.ps1 | iex
```

这是远程下载并执行脚本。即使脚本链接来自官方文档，也建议先下载到本地查看，再执行：

```bash
curl -fsSL -o ~/setup-deepseek-codex.sh https://cdn.deepseek.com/api-docs/codex-deepseek-setup.sh
less ~/setup-deepseek-codex.sh
bash ~/setup-deepseek-codex.sh
```

不要把官方命令随手改成 `curl ... | bash`。交互式脚本从标准输入读取菜单和 API Key，管道执行可能让脚本内容和交互输入争用同一个输入流，表现为命令执行但菜单异常或没有写入配置。先下载再执行也更容易留下审查记录。

运行脚本时注意：

```text
先完全退出 Codex 桌面端，避免它同时改写配置。
按照当前脚本菜单选择已明确支持的模型。
输入 API Key 时注意终端是否明文回显，录屏前先清除或打码。
保留脚本输出，尤其是备份路径、冲突字段和校验结果。
```

脚本执行后检查：

```text
是否生成了 config.toml 备份。
是否生成并写入了 models.json。
model 是否指向 deepseek-v4-flash。
model_provider 是否与提供方配置段一致。
wire_api 是否为 responses。
原有 MCP 和项目安全设置是否仍然存在。
```

脚本可能会替换模型目录，导致原来的模型不再显示。这不一定是脚本失败，而可能是它按单一提供方模式写入了新的目录。需要切回原账号或模型时，应使用脚本备份或手动恢复，不要直接删除整个 `.codex`。

## 4. 方式二：手动编辑 config.toml

手动方式适合需要逐项审查配置的人。下面是关键字段示例，真实字段和完整模型目录应以 [DeepSeek 官方 Codex 接入文档](https://api-docs.deepseek.com/zh-cn/quick_start/agent_integrations/codex) 为准：

```toml
model = "deepseek-v4-flash"
model_provider = "deepseek"
preferred_auth_method = "apikey"
forced_login_method = "api"
model_reasoning_effort = "high"
model_catalog_json = "~/.codex/models.json"

[model_providers.deepseek]
name = "deepseek"
base_url = "https://api.deepseek.com/"
wire_api = "responses"
experimental_bearer_token = "<你的 DeepSeek API Key>"
```

字段可以这样理解：

| 字段 | 作用 |
| --- | --- |
| `model` | 默认模型标识 |
| `model_provider` | 选择哪个提供方配置段 |
| `forced_login_method` | 强制使用 API 模式而不是账号登录 |
| `model_reasoning_effort` | 默认推理强度，合法档位以模型目录为准 |
| `model_catalog_json` | 指向自定义模型元数据文件 |
| `base_url` | 上游 API 地址 |
| `wire_api` | 请求协议，当前接入要求为 Responses |
| `experimental_bearer_token` | API Key 配置方式，属于敏感字段 |

`models.json` 不只是模型名称列表。它可能还包含上下文窗口、推理档位、工具格式和 Codex 运行所需的指令元数据。不要只复制网上一段缩略 JSON，也不要把 `{"deepseek-v4-flash": {...}}` 这种未经核对的结构当作官方格式。优先让官方脚本生成，或从官方文档完整复制后再根据当前版本校验。

如果不希望把 Key 直接写在 TOML 中，可以研究环境变量或系统密钥存储方案，但这属于额外加固。遇到认证失败时，先确认当前 Codex 版本和官方示例支持的字段，再逐步替换存储方式，不要同时修改认证、模型和协议。

## 5. 如果使用兼容网关

如果通过 4SAPI 这类网关接入 DeepSeek，需要把它当作一条独立兼容路径验证：`base_url`、模型映射、Responses 协议、流式结束事件、工具调用和错误格式都要分别确认。只替换地址而不核对协议，可能出现普通文本请求成功、Codex 工具循环失败的情况。

这类网关验证不属于官方原生配置的一部分，也不能仅凭 DeepSeek 直连成功就推断网关路径同样兼容。建议保留独立模型标识和请求日志，便于区分上游模型问题与网关改写问题。

## 6. 配置完成后的验证顺序

### 第一步：验证配置文件

先用 TOML/JSON 解析器检查文件格式，再启动 Codex。不要用文本搜索代替解析，因为引号、转义和嵌套字段都可能导致配置无法加载。

### 第二步：验证模型路由

进入一个不含敏感数据的测试项目，执行只读任务：

```text
只阅读当前项目，说明入口文件、测试命令和最近一次提交影响的模块。
不要修改文件，不要删除内容，不要发送外部请求。
```

启动信息可能显示模型名称，也可能显示“自定义”之类的统一标签。不能只凭这个文字判断路由，应该结合请求日志、上游用量记录或安全测试任务确认请求确实到达目标提供方。

不要直接问模型“你是什么模型”。Codex 的系统指令可能要求它用固定身份回答，这个回答不能证明底层请求走了哪条线路。

### 第三步：验证 Agent 工具循环

再给一个可回滚的小任务：

```text
先查找项目测试命令，再读取一个入口文件，最后只修改一处注释。
修改前说明计划，完成后运行相关测试，不要触碰其他文件。
```

检查它是否能连续读取、修改和运行测试，工具结果是否能回传，失败后是否能停止。只发送“你好”只能证明文本响应可用，不能证明 Codex 接入成功。

### 第四步：验证边界

```text
流式输出是否能正常结束？
工具调用是否能完成一轮以上？
模型目录是否被正确加载？
推理档位是否是当前配置支持的值？
超时和 4xx/5xx 是否能被识别？
回滚后原有 MCP、权限和项目设置是否仍在？
```

## 7. 回滚与恢复

回滚优先使用脚本生成的备份或你自己保存的完整 `.codex` 副本。恢复后彻底重启 Codex，再用只读任务验证：

```text
启动信息回到原模型或原提供方。
DeepSeek 不再被默认选中。
MCP、信任级别和项目配置仍然存在。
测试仓库可以完成一次只读任务。
模型目录和认证文件没有被误删。
```

不要为了回滚执行 `Remove-Item -Recurse ~/.codex` 或类似命令。整个目录可能包含认证、MCP、历史会话、模型目录和安全设置，删除范围远超“换回原模型”。

## 8. 常见报错排查

| 现象 | 优先检查 |
| --- | --- |
| `wire_api = "chat" is no longer supported` | 旧教程配置仍在使用 Chat，改为当前官方要求的 Responses |
| 404 或流式异常 | 模型名、协议、Base URL 是否重复拼接 |
| 启动后仍显示原模型 | 是否修改了错误用户目录，是否有更高优先级配置覆盖 |
| 找不到 DeepSeek | `models.json` 路径、格式和 Codex 是否彻底重启 |
| `fallback model metadata` 或 `Unknown model` | 模型目录未加载或格式与当前版本不匹配 |
| 一直转圈 | 先区分余额、网络、超时和工具循环问题，不要盲目重试 |
| 原有 GPT 模型不见了 | 检查脚本备份和当前模型目录是否被替换 |

## 9. 一份可复制的配置审计 Prompt

```text
请先不要修改我的 Codex 配置，只做接入审计。

检查顺序：
1. 找到当前用户级 config.toml、models.json、认证文件和备份目录。
2. 读取并解析 TOML/JSON，不要只用文本搜索判断格式。
3. 输出当前 model、model_provider、base_url、wire_api、model_catalog_json 和认证来源。
4. 检查当前配置是否能支持 deepseek-v4-flash 的 Responses 和工具调用。
5. 标出可能覆盖当前配置的命令行、项目级或 profile 配置。
6. 不输出任何完整 API Key，只报告是否存在和来源。
7. 设计一个只读任务和一个可回滚小修改作为验证。
8. 给出回滚所需文件清单、备份路径和恢复后的检查项。

不要删除整个 .codex，不要执行远程脚本，不要修改配置，除非我明确确认。
```

## 10. 先把配置变更做成可比较的差异

不要只保存“改之前”和“改之后”两份完整文件，还要生成一份脱敏差异。这样出问题时，可以快速判断变化来自模型、协议、认证还是模型目录。

建议至少比较这些字段：

```text
model
model_provider
base_url
wire_api
model_reasoning_effort
model_catalog_json
preferred_auth_method
forced_login_method
MCP 和 profile 相关字段
```

API Key、OAuth token 和其他秘密字段不能进入差异文件。可以把它们统一替换成 `<redacted>`，只保留“字段存在、来源是什么、是否发生变化”。

配置差异最好记录修改时间、执行人、使用的脚本或工具版本、回滚文件路径。它们不会替代 Git，但能让本机配置变更有基本的可追溯性。

## 11. Windows 用户的额外检查

Windows 上最容易出现的不是 TOML 语法错误，而是“改了一个用户目录，启动的 Codex 却属于另一个用户或进程”。可以按下面顺序排查：

```powershell
Write-Output $env:USERPROFILE
Test-Path "$env:USERPROFILE\.codex\config.toml"
Test-Path "$env:USERPROFILE\.codex\models.json"
Get-Process codex -ErrorAction SilentlyContinue
```

运行脚本前先关闭相关 Codex 进程。PowerShell 远程脚本执行前，建议先下载并查看：

```powershell
$setupPath = Join-Path $env:TEMP 'setup-deepseek-codex.ps1'
Invoke-WebRequest -UseBasicParsing `
  -Uri 'https://cdn.deepseek.com/api-docs/codex-deepseek-setup-en.ps1' `
  -OutFile $setupPath
Get-Content -Raw $setupPath
```

确认脚本内容、目标目录和备份逻辑后，再在当前 PowerShell 会话中执行。不要把包含 API Key 的终端输出直接截图，也不要把临时脚本长期留在共享目录。

## 12. 配置验收矩阵

| 层次 | 验证动作 | 通过证据 | 失败后先查什么 |
| --- | --- | --- | --- |
| 文件 | 解析 TOML 和 JSON | 解析器无错误 | 引号、路径、数组结构 |
| 路由 | 执行一次只读请求 | 上游日志有对应请求 | 用户目录、profile、覆盖配置 |
| 协议 | 检查 Responses 请求和结束状态 | 响应能正常收口 | `wire_api`、Base URL、客户端版本 |
| 工具 | 执行搜索、读取、补丁和测试 | 工具结果连续回传 | 模型目录、工具声明、权限 |
| 认证 | 检查 401、403 和轮换流程 | Key 来源明确且不入仓库 | Key 状态、权限、环境变量 |
| 回滚 | 恢复完整备份后重启 | 原模型和原配置恢复 | 备份是否完整、进程是否退出 |

只有“文本能返回”这一项通过时，不能把接入写成成功。Codex 的核心价值在工具循环，至少需要完成一个只读任务和一个可回滚的小修改。

## 13. 结论与限制

官方原生接入的核心不是一行模型名，而是四个部分同时一致：`deepseek-v4-flash`、模型提供方、Responses 协议和可加载的模型目录。官方脚本适合快速生成完整配置，手动方式适合逐项审查；两者都要先备份、再验证、可回滚。

本文只覆盖本地 Codex 配置路径。桌面端、CLI、IDE 插件和第三方工具的界面与版本可能变化；第三方网关还会增加一层协议和认证差异。发布或扩大权限前，应以当前官方文档、实际请求日志和可回滚的小任务验证结果为准。

## 资料与说明

- [DeepSeek：Codex 接入文档](https://api-docs.deepseek.com/zh-cn/quick_start/agent_integrations/codex)：核对模型、Responses 协议和官方配置示例。
- [OpenAI：Codex config.toml 官方文档](https://learn.chatgpt.com/docs/config-file/config-reference#configtoml)：核对用户级配置、项目级配置和模型提供方字段边界。
