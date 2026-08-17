---
title: "【大模型API中转站】第23期 Hermes Mac安装避坑 | GUI/Gateway别搞混"
category: 人工智能
tags:
  - 大模型API中转站
  - Hermes
  - macOS
  - Gateway
  - 飞书机器人
  - 4SAPI
description: "从 macOS 安装 Hermes Desktop 后找不到图标、关闭 GUI 后飞书机器人仍能回复两个常见问题入手，讲清 GUI、Gateway 和 ~/.hermes 会话目录的关系。"
---

# 【大模型API中转站】第23期 Hermes Mac安装避坑 | GUI/Gateway别搞混

本文是【大模型API中转站】系列的第23篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型 API 的任督二脉。建议先收藏，随用随查。

上一篇我们聊的是 OpenSpec 怎么让老项目里的 AI 改动少翻车。这一篇换一个更贴近实操的坑：Hermes Desktop 在 Mac 上到底怎么装？为什么明明关掉了桌面窗口，飞书机器人还在正常回复？为什么双击 `dmg` 能打开，弹出镜像以后 App 又像消失了一样？

这些问题看起来像新手操作失误，实际背后是 Hermes 的架构设计：GUI、Gateway 和会话数据不是一回事。理解这三层以后，你再接 4SAPI 这类大模型API中转站、再配飞书/Telegram/Discord，就不会把“关窗口”和“停服务”混在一起。

## 1. 一个真实的尴尬场面

装好 Hermes Desktop，第一次打开图形界面聊得正欢。退出后想再打开，桌面图标找不到了。翻遍下载文件夹，只看到一个 `hermes-setup.dmg`。双击挂载后 App 又出现了，可一旦弹出 dmg 镜像，图标再次消失。

你可能会怀疑：是不是 Hermes 没装好？是不是 Mac 把应用删了？

更让人困惑的是：Mac 上明明已经退出 Hermes 桌面窗口，连着的飞书机器人却还在正常回复消息。难道 Hermes 还有一个看不见的后台？

答案是：有。Hermes Desktop 不是一个“关窗口就全停”的单进程小工具，它至少要分成三层理解：

- GUI 桌面窗口。
- Gateway 后台网关。
- `~/.hermes/` 本地配置和会话数据。

## 2. Hermes Desktop 的双进程架构

大多数人对桌面软件的直觉是：一个图标、一个窗口、一个进程，关了就没。但 Hermes 更像“桌面入口 + 后台网关”的组合。

```text
┌─────────────────────────────────────────────┐
│                 你的 Mac                    │
│                                             │
│  ┌──────────────────┐    ┌───────────────┐ │
│  │   Hermes GUI     │    │   Gateway     │ │
│  │  桌面聊天窗口     │    │  后台消息网关  │ │
│  │                  │    │               │ │
│  │  本地聊天交互     │    │ 飞书/Telegram │ │
│  │  读写 state.db   │────│ Discord 路由   │ │
│  └──────────────────┘    └───────┬───────┘ │
│                                  │          │
│                         ┌────────▼───────┐ │
│                         │   ~/.hermes/   │ │
│                         │  state.db      │ │
│                         │  sessions/     │ │
│                         │  config.yaml   │ │
│                         │  .env          │ │
│                         └────────────────┘ │
└─────────────────────────────────────────────┘
```

关键认知如下：

| 组件 | 作用 | 关闭 GUI 后会怎样 |
| --- | --- | --- |
| Hermes GUI | 本地桌面聊天窗口 | 会关闭，只影响本地窗口 |
| Gateway | 飞书/Telegram/Discord 消息收发核心 | 可能仍在后台运行 |
| `state.db` | 本地会话数据库 | 存在硬盘上，不随窗口关闭消失 |
| `config.yaml` | 模型、平台、Gateway 配置 | 改完通常要重启 Gateway |

所以：关掉 GUI，不等于停掉 Gateway。找不到桌面图标，也不等于会话丢了。

## 3. macOS 上 Hermes 的正确安装方式

Mac 和 Windows 的安装习惯不一样。Windows 安装器经常会在桌面放快捷方式；macOS 不会自动帮你在桌面生成应用图标。Mac 应用通常放在 `/Applications`，再通过启动台、Spotlight 或 Dock 打开。

你遇到“图标消失”，最常见原因是：

```text
双击 .dmg
-> 直接在挂载镜像里运行 Hermes
-> 没有拖进 Applications
-> 弹出 dmg 后应用入口消失
```

正确安装三步：

```text
1. 下载 Hermes macOS 安装包。
2. 打开 .dmg，把 Hermes 图标拖进 /Applications。
3. 从启动台或 Spotlight 搜 Hermes 启动，再固定到 Dock。
```

简单画一下：

```text
[Hermes.app]  ->  /Applications/Hermes.app
```

之后你就不需要再依赖 `hermes-setup.dmg`。dmg 只是安装载体，不是应用真正应该长期运行的位置。

## 4. 安装前先做 5 个检查

如果你是第一次在 Mac 上部署 Hermes，不建议一上来就接飞书和多模型。先把本地环境确认清楚，后面排查会轻松很多。

| 检查项 | 为什么要看 | 建议 |
| --- | --- | --- |
| 芯片架构 | Apple Silicon 和 Intel 安装包可能不同 | 下载前确认是 arm64 还是 x64 |
| macOS 权限 | 首次运行可能被系统拦截 | 到“系统设置 -> 隐私与安全性”允许打开 |
| 终端命令 | Gateway 需要 CLI 管理 | 确认 `hermes` 命令可用 |
| 配置目录 | 需要知道配置和会话在哪 | 确认 `~/.hermes/` 是否生成 |
| 模型通道 | 飞书机器人最终要调用模型 | 先准备好 4SAPI Key 或官方 Key |

可以先跑这几个命令：

```bash
uname -m
which hermes
hermes --version
ls -la ~/.hermes/
```

如果 `which hermes` 找不到命令，不代表桌面 App 一定不能用，但后续管理 Gateway 会很不方便。建议先把 CLI 路径或安装方式理顺，再继续接平台。

## 5. Gateway 怎么启动、停止和重启

如果你接了飞书机器人，真正负责收发消息的是 Gateway。GUI 可以关，但 Gateway 可能仍然在后台跑。

常用命令：

```bash
hermes gateway status
hermes gateway stop
hermes gateway start
hermes gateway restart
```

什么时候需要重启 Gateway？

- 改了 `~/.hermes/config.yaml`。
- 改了模型 provider。
- 改了 4SAPI 的 `base_url` 或 `api_key`。
- 飞书机器人不响应，但本地 GUI 正常。
- 切换平台默认模型后想立即生效。

如果你用 4SAPI 管多模型，Gateway 里只要统一配置中转站入口，飞书、Telegram、桌面端就能复用同一套模型通道。典型配置思路是：

```yaml
custom_providers:
- name: 4s
  base_url: https://4sapi.com/v1
  api_key: sk-xxxxxxxxxxxxxxxx
  api_mode: anthropic_messages
```

真实模型名以 4SAPI 模型广场显示为准，不要凭记忆手打。

除了启停命令，还建议你养成两个习惯。

第一，改配置前先备份：

```bash
cp ~/.hermes/config.yaml ~/.hermes/config.yaml.bak.$(date +%Y%m%d-%H%M%S)
```

第二，改完配置只重启 Gateway，不要直接删除整个 `~/.hermes/`：

```bash
hermes gateway restart
hermes gateway status
```

如果你正在排查 4SAPI 连通性，可以先用一个低风险的小问题测试飞书机器人，比如：

```text
请回复当前模型名称和一句简短问候，不要调用工具。
```

这样能确认模型通道是否通，而不会让 Agent 立刻执行复杂任务。

## 6. 会话持久化：聊天记录到底在哪

很多人以为聊天记录存在 App 里面，其实不是。Hermes 的配置和会话一般在：

```bash
~/.hermes/state.db
~/.hermes/sessions/
~/.hermes/config.yaml
~/.hermes/.env
```

你可以这样验证：

```bash
ls -la ~/.hermes/
hermes sessions list
```

只要 `state.db` 还在，历史会话大概率就还在。App 从哪里启动，是否弹出了 dmg，通常不会影响这个目录。

常见行为对会话的影响：

| 行为 | 会话是否丢失 |
| --- | --- |
| 正常退出 GUI | 不丢 |
| 重启电脑 | 不丢 |
| 弹出 dmg 镜像 | 不丢 |
| 删除 `/Applications/Hermes.app` | 通常不丢 |
| 手动删除 `~/.hermes/` | 会丢 |

所以排查问题时，不要第一反应就删 `~/.hermes/`。这个目录里有你的配置、会话和平台状态，动它之前最好先备份。

如果你要换电脑，可以先打包备份：

```bash
tar -czf hermes-backup-$(date +%Y%m%d).tgz ~/.hermes
```

迁移到新机器后，再按需恢复配置。注意：如果里面包含 `.env`、平台 token、API Key，不要把这个压缩包上传到公共网盘或发到群里。团队环境里更推荐只迁移非敏感配置，把 4SAPI Key 重新按项目发放。

## 7. 常见排查清单

如果你装完 Hermes 后打不开，按这个顺序查：

```text
1. 是否已经把 App 拖进 /Applications。
2. 是否仍在 dmg 挂载镜像里运行。
3. Spotlight 能否搜到 Hermes。
4. Dock 图标是否只是没固定。
5. `hermes gateway status` 是否显示后台服务在运行。
6. `~/.hermes/config.yaml` 是否存在。
```

如果飞书机器人还在回复，但桌面 GUI 关了：

```text
这是正常现象。说明 Gateway 仍在后台运行。
```

如果改了 4SAPI 配置但飞书没生效：

```text
先 hermes gateway restart，再在飞书里重新开一轮会话或查看当前模型。
```

如果机器人完全不响应：

```text
先看 gateway status，再看平台 token、回调地址、网络和日志。
```

可以再细分成三类问题。

| 现象 | 优先排查 | 常见原因 |
| --- | --- | --- |
| GUI 打不开 | 安装位置、macOS 安全拦截 | App 还在 dmg、未授权打开 |
| 飞书不回复 | Gateway、飞书回调、平台 token | Gateway 没启动或回调失败 |
| 飞书回复但模型不对 | `config.yaml`、provider、models | 裸模型名走了官方路由 |
| 模型报 401/403 | API Key、额度、权限 | Key 错误、项目额度不足 |
| 模型报 404 | 模型名 | 没从 4SAPI 模型广场复制完整名称 |

排查顺序建议固定成：

```text
先确认 App 安装
再确认 Gateway 状态
再确认平台连接
再确认模型 provider
最后看具体模型错误码
```

不要一上来就改一堆配置。一次只改一个变量，才能知道是哪一步起作用。

## 8. 4SAPI 在 Hermes 架构里的位置

Hermes 解决的是“多入口 Agent 工作流”：桌面、飞书、Telegram、Discord 都可以接进来。

4SAPI 解决的是“多模型 API 统一接入”：Claude、GPT、GLM、Kimi、DeepSeek 等模型可以通过统一入口管理。

推荐分层是：

```text
飞书 / Telegram / Hermes Desktop
        ↓
Hermes Gateway
        ↓
4SAPI 大模型API中转站
        ↓
Claude / GPT / GLM / Kimi / DeepSeek
        ↓
日志、账单、额度、模型路由
```

这样做的好处是：你不需要在每个平台里分别维护一堆官方 Key。模型切换、额度控制、成本统计都集中到 4SAPI，Hermes 只负责把不同入口的消息路由好。

一个团队部署时，可以按下面这样分工：

| 层级 | 负责人 | 关注点 |
| --- | --- | --- |
| Hermes Desktop | 使用者本人 | 本地 GUI、会话、日常调试 |
| Hermes Gateway | 技术负责人 | 平台连接、机器人状态、重启策略 |
| 4SAPI | 管理员/项目负责人 | Key、模型权限、额度、日志 |
| 飞书群 | 团队成员 | 命令规范、模型切换、反馈问题 |

比较稳的做法是：普通成员只知道怎么在飞书里使用机器人，不直接接触模型 Key；项目负责人管理 4SAPI 的项目 Key；技术负责人维护 Hermes Gateway 和平台回调。

## 9. 团队部署清单

正式给团队使用前，建议过一遍清单：

```text
01. Hermes.app 已拖入 /Applications
02. `hermes` CLI 可用
03. `~/.hermes/config.yaml` 已备份
04. Gateway 可 start/stop/status
05. 飞书平台回调可用
06. 4SAPI provider 已配置
07. API Key 使用项目级令牌，不使用个人主 Key
08. 常用模型已写入 models 列表
09. 默认模型和平台模型已确认
10. 团队知道 GUI 关闭不等于 Gateway 停止
11. 知道如何重启 Gateway
12. 知道不要删除 `~/.hermes/`
```

如果这 12 项都过了，Hermes 的基础环境就比较稳了。后续再排模型、prompt、工具调用问题，会轻松很多。

## 10. 总结

这篇文章讲清了三个核心概念：

1. GUI 窗口不等于 Gateway 后台。关窗口不会自动停掉飞书机器人。
2. macOS 安装必须拖进 `/Applications`。在 dmg 里直接运行，弹出镜像后应用入口就会消失。
3. 会话和配置在 `~/.hermes/`。不要把 App 图标和会话数据混为一谈。

如果你准备把 Hermes 用在团队 Agent 工作流里，建议顺手把 4SAPI 作为统一模型入口：Gateway 管消息入口，4SAPI 管模型、Key、账单和路由。先把安装位置、Gateway 状态和 `~/.hermes/` 搞清楚，再去调模型，效率会高很多。下一篇我们继续讲 Hermes 里最常见的模型切换坑：为什么 `/model claude-sonnet-4-6` 会跑到官方 Anthropic，而不是你配置好的 4SAPI。

参考资料：

- Hermes 官方网站：https://hermes-agent.nousresearch.com
- 4SAPI 接入地址：https://4sapi.com/v1
