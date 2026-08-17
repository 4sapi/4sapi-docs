---
title: "在 Claude Code 中安装并验证 Codex 插件"
category: 人工智能
tags:
  - codex-plugin-cc
  - Claude Code
  - Codex CLI
description: "本文按官方仓库说明安装 codex-plugin-cc，验证本机 Codex 认证，并用后台只读审查检查命令、状态和结果读取是否正常。"
---

# 在 Claude Code 中安装并验证 Codex 插件

在 Claude Code 中调用 Codex 做独立审查时，安装插件只是第一步；本机 Codex CLI、认证、项目目录和配置任一环节不可用，插件命令都无法完成任务。本文只解决首次安装与验收：依据官方仓库检查 Node.js 和本机 Codex，添加 marketplace 并安装插件，运行 `/codex:setup`，再以只读的 `/codex:review --background` 创建任务，通过 `status` 与 `result` 确认后台链路。文中的修改型 Rescue 只说明权限边界，不作为首次测试。这套顺序把安装问题与审查问题分开，失败时可以定位在本地运行时、插件加载还是任务管理。

常见工作流是：

```text
Claude Code 讨论需求和修改代码。
Codex 再做一次独立审查。
```

问题是每次切换工具，都要重新解释仓库、分支、改动目标和风险。

OpenAI 官方仓库 [`openai/codex-plugin-cc`](https://github.com/openai/codex-plugin-cc) 提供了通过 `/codex:*` 命令调用本机 Codex 的插件。

插件不提供一套独立 Codex 运行时。

它复用本机已有的：

```text
Codex CLI 安装。
Codex 登录状态。
Codex config.toml。
当前仓库和本机环境。
```

插件沿用本机 Codex 的认证和配置；它不会提供另一套运行时。

## 1. 插件能做什么

官方 README 提供这些命令：

| 命令 | 作用 | 是否可能改代码 |
| --- | --- | --- |
| `/codex:review` | 普通只读代码审查 | 否 |
| `/codex:adversarial-review` | 可指定风险焦点的对抗性审查 | 否 |
| `/codex:rescue` | 调查问题、尝试修复或继续任务 | 是 |
| `/codex:transfer` | 把 Claude Code 会话导入 Codex | 否 |
| `/codex:status` | 查看后台任务状态 | 否 |
| `/codex:result` | 查看已完成任务结果 | 否 |
| `/codex:cancel` | 取消后台任务 | 停止任务 |
| `/codex:setup` | 检查安装、认证和审查门禁 | 配置操作 |

最重要的边界：

```text
review 负责指出问题。
rescue 才负责调查或尝试修改。
```

## 2. 工作原理

调用链可以理解为：

```text
Claude Code
    ↓ /codex:* 命令
codex-plugin-cc
    ↓
本机 Codex app server / CLI
    ↓
当前仓库、认证和 config.toml
    ↓
审查结果或 Codex 会话
```

这就是它能减少上下文搬运的原因。

Claude Code 仍然是主工作台，插件把当前仓库任务交给同一台机器上的 Codex。

## 3. 安装前要求

官方 README 当前列出：

```text
Node.js 18.18 或更高版本。
ChatGPT 订阅，包括 Free，或 OpenAI API Key。
```

插件需要使用 Codex CLI。如果本机没有安装，`/codex:setup` 在 npm 可用时可以提供安装选项；也可以手工安装。

先检查本机命令：

```bash
node --version
codex --version
```

不要把诊断输出直接发到公开渠道。即使工具会脱敏，也应先检查其中是否包含内部路径和环境信息。

## 4. 安装 Codex CLI

如果 `codex --version` 不存在，可以执行：

```bash
npm install -g @openai/codex
```

然后登录：

```bash
codex login
```

官方 README 给出的 Claude Code 登录方式是：

```text
!codex login
```

认证方式以当前 Codex 文档为准。不要把 API Key 直接写在命令历史、项目文件或 `CLAUDE.md` 中。

## 5. 安装 codex-plugin-cc

在 Claude Code 里依次执行：

```text
/plugin marketplace add openai/codex-plugin-cc
/plugin install codex@openai-codex
/reload-plugins
```

然后执行：

```text
/codex:setup
```

`/codex:setup` 会检查 Codex 是否已安装，以及认证状态是否可用。

安装后应该看到：

```text
/codex:* 命令。
/agents 中的 codex:codex-rescue subagent。
```

## 6. 第一次使用

先做只读审查，不要一开始就把修改权限交给 Rescue：

```text
/codex:review --background
```

查看后台任务：

```text
/codex:status
```

任务结束后读取结果：

```text
/codex:result
```

这是官方 README 给出的简单首次流程。

## 7. 为什么建议后台运行

多文件审查可能需要较长时间。

后台模式允许你继续在 Claude Code 中：

```text
运行测试。
补充文档。
检查 diff。
准备提交说明。
```

需要停止时：

```text
/codex:cancel
```

有多个任务时可以带任务 ID：

```text
/codex:status task-abc123
/codex:result task-abc123
/codex:cancel task-abc123
```

任务 ID 以插件实际输出为准。

## 8. Codex 配置从哪里读取

插件沿用 Codex 的配置层：

```text
用户级：~/.codex/config.toml
项目级：项目根目录/.codex/config.toml
```

项目级配置只在项目受信任时加载。

官方 README 给出的示例：

```toml
model = "gpt-5.4-mini"
model_reasoning_effort = "high"
```

这只是示例，不代表所有审查都应该固定使用该模型和 effort。

模型名要以当前账号和接入服务可用列表为准。

## 9. 常见安装问题

### `/codex:setup` 找不到 Codex

检查 `codex --version`，确认 npm 全局安装目录已经进入 PATH。重启 Claude Code 后再试。

### Codex 已安装但尚未认证

执行：

```text
!codex login
```

然后重新运行 `/codex:setup`。

### 插件命令没有出现

检查 marketplace 是否添加成功、插件名称是否为 `codex@openai-codex`，然后执行 `/reload-plugins`。

### 项目配置不生效

确认 `.codex/config.toml` 位于启动 Claude Code 的项目根目录，并确认该项目被 Codex 信任。

### `/codex:transfer` 不可用

官方 README 说明该命令依赖 Codex 的会话导入能力。先按当前 Codex 安装文档更新 CLI，再重新运行 setup；不要使用未经当前 CLI 帮助确认的升级命令。

## 10. 安全边界

插件使用当前仓库和本机环境。

安装前要确认：

```text
仓库本身可信。
项目级 .codex/config.toml 已审查。
没有把密钥写进项目文件。
review 与 rescue 的权限边界被团队理解。
后台任务不会访问无关目录。
```

`/codex:review` 是只读审查。

`/codex:rescue` 可能尝试修改，因此调用前必须写清任务边界，并在完成后检查 git diff。

## 11. 安装验收清单

```text
[ ] Node.js 版本不低于 18.18
[ ] codex --version 能返回版本
[ ] 本机 Codex 已通过当前官方方式认证
[ ] marketplace 添加成功
[ ] codex@openai-codex 安装成功
[ ] 已执行 /reload-plugins
[ ] /codex:setup 通过
[ ] /agents 中出现 codex:codex-rescue
[ ] /codex:review --background 能创建任务
[ ] /codex:status 和 /codex:result 能读取结果
```

## 12. 结论与限制

`codex-plugin-cc` 的价值不是让 Claude Code 里多出几个命令。

它把两个工具放进同一条开发链：

```text
Claude Code 负责推进。
Codex 负责独立审查或调查。
插件负责调用和任务管理。
```

首次验收可以使用官方 README 给出的这组只读流程：

```text
/codex:setup
/codex:review --background
/codex:result
```

通过标准是：setup 能识别本机 Codex，后台 review 能创建任务，status 能看到状态，result 能读取完成结果。`review` 与 `adversarial-review` 在官方说明中是只读命令；`rescue` 可能调查并尝试修改，使用前必须另行确认任务范围并检查 Git diff。

本文命令依据核对时的官方仓库 README。插件、Codex CLI 与 Claude Code 都可能更新，实际安装前仍应复查 README；项目级 `.codex/config.toml` 只应在可信项目中加载，认证信息和诊断输出也不应提交到仓库或公开渠道。
