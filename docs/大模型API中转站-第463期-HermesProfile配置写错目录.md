---
title: "Hermes 配置为什么写进了错误的 Profile"
category: 人工智能
tags:
  - AI Agent
  - 配置管理
  - 命令行
description: "说明 Hermes 的活跃 Profile 与 -p 参数如何决定命令写入位置，并给出修改多 Profile Gateway 配置前后的检查方法。"
---

# Hermes 配置为什么写进了错误的 Profile

多 Profile 环境里，一个很隐蔽的错误是：明明要改 `default`，命令却把配置写进了 `hermes2`。表面上看两条命令都显示成功，实际两个 Profile 的状态并没有发生预期变化，后续启动 Gateway 时才暴露问题。

根因通常不是 YAML 语法，而是没有区分“当前活跃 Profile”和“命令明确指定的 Profile”。下面说明这两个概念如何影响命令行为，并给出一套不会误改目录的操作习惯。

## 活跃 Profile 是不带 `-p` 命令的默认目标

官方文档规定，`hermes profile use <name>` 会将一个 Profile 设为活跃 Profile。此后所有不带 `-p` 的 `hermes` 命令都会使用它；`hermes profile use default` 可返回基础 Profile。[官方命令参考](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/i18n/zh-Hans/docusaurus-plugin-content-docs/current/reference/profile-commands.md?plain=1)

例如，当前活跃 Profile 是 `hermes2` 时：

```bash
hermes config set gateway.multiplex_profiles false
```

写入的是 `hermes2`，而不是 `default`。即使紧接着再执行一条带 `-p hermes2` 的命令，两条命令仍会落在同一个文件里。

因此，在多 Profile 环境中，先运行：

```bash
hermes profile list
```

输出中带 `*` 的行就是当前默认目标。这个检查应放在每次修改、重启或排错前，而不是出错以后。

## 两种安全操作方式

### 方式一：每条命令都使用 `-p`

需要修改两个 Profile 时，最稳妥的方式是显式指定目标：

```bash
hermes -p default config set gateway.multiplex_profiles false
hermes -p hermes2 config set gateway.multiplex_profiles false
```

这种写法不依赖当前活跃 Profile，适合脚本、交接文档和复杂排错流程。

### 方式二：先切换活跃 Profile，再运行普通命令

只操作一个 Profile 时，可以先切换：

```bash
hermes profile use default
hermes config set gateway.multiplex_profiles false
```

操作完成后再次执行 `hermes profile list`，确认 `default` 仍是带 `*` 的项。这个方式输入更短，但在多窗口、多终端或多人协作时更容易遗漏切换步骤。

对于影响 Gateway、消息凭据或模型提供方的配置，优先使用 `-p`。省下的几个字符，不值得换来一次写错目录。

## 为什么“命令成功”仍然可能改错

CLI 返回成功，通常只说明它已经在某个有效 Profile 中完成了写入。它不会替你判断“这个 Profile 是否正是你脑中想修改的那个”。当 `hermes2` 是活跃 Profile 时，下面两条命令都可能返回成功：

```bash
hermes config set gateway.multiplex_profiles false
hermes -p hermes2 config set gateway.multiplex_profiles false
```

但它们写的是同一份 `hermes2` 配置。若你的目标是同时修改 default 和 hermes2，default 的值根本没有被触及。

这类错误之所以难发现，是因为输出通常只展示“设置了什么值”和“写入了哪个路径”。人在连续执行命令时容易把两条成功提示理解为“两个目标都已完成”。解决办法不是记住更多命令，而是把“目标 Profile”当成每条变更的必填参数。

可以把配置操作分为三层：

| 层级 | 要回答的问题 | 推荐验证方式 |
| --- | --- | --- |
| 命令层 | 这条命令将操作哪个 Profile？ | 看 `-p`，或先看 `profile list` 的 `*` |
| 文件层 | 目标配置文件是否真的出现了预期字段？ | `grep`、配置查看命令或人工打开文件 |
| 运行层 | 当前 Gateway 是否已读取到新配置？ | 对应 Profile 的 `gateway status` 与日志 |

三层必须都成立，才能称为“配置已生效”。只通过命令层，最多只能说明你发出了一次写入请求。

## 常见误操作与纠正方式

| 误操作 | 为什么有问题 | 如何纠正 |
| --- | --- | --- |
| 终端刚切过 `hermes profile use hermes2`，随后忘了切回 | 后续不带 `-p` 的命令都会作用于 hermes2 | 对关键命令统一使用 `-p default` 或 `-p hermes2` |
| 将“Profile 名称”当成“工作目录” | Profile 管的是 Hermes 状态，不决定终端运行目录 | 分别确认 Profile、`terminal.cwd` 和沙箱策略 |
| 只复制第一条成功输出，不保留实际路径 | 无法判断两个命令是否写入不同文件 | 记录 CLI 输出中的配置路径，并对照目标清单 |
| 修改完立即重启两个 Gateway | 出错时无法知道是哪个 Profile 使用了错误配置 | 先验证文件，再按一个 Profile 一个 Profile 启动 |
| 克隆 Profile 后只改模型，不检查 `.env` | 可能意外继承机器人凭据或平台配置 | 对新 Profile 的凭据和平台段单独审阅，内容保持脱敏 |

这些动作并非 Hermes 独有。任何有“当前上下文”概念的 CLI，都可能让没有显式目标的命令落在错误位置。多 Profile 只是把这个风险放大了，因为多个配置文件名称相似、功能相近，而且都能正常被工具读取。

## 一套可重复执行的配置修改协议

针对会影响消息服务的配置，可采用下面五步协议。示例仍以 `gateway.multiplex_profiles` 为例，但同样适用于模型、工具、工作目录和连接器配置。

### 1. 写下变更矩阵

先明确每个 Profile 预期的值，不要边执行边临时决定：

| Profile | 目标字段 | 预期值 | 当前 Gateway 是否运行 |
| --- | --- | --- |
| default | `gateway.multiplex_profiles` | `false` | 是 |
| hermes2 | `gateway.multiplex_profiles` | `false` | 否 |

表格很简单，但它能避免“我已经改过一个了”的模糊记忆取代实际状态。

### 2. 使用显式目标执行写入

```bash
hermes -p default config set gateway.multiplex_profiles false
hermes -p hermes2 config set gateway.multiplex_profiles false
```

在脚本中更应避免裸 `hermes config set`。脚本可能由不同用户、不同终端或自动化环境调用，活跃 Profile 未必与你写脚本时相同。

### 3. 验证两个配置文件

```bash
grep -n 'multiplex_profiles' \
  "$HOME/.hermes/config.yaml" \
  "$HOME/.hermes/profiles/hermes2/config.yaml"
```

如需人工打开文件，不要修改和保存无关的字段。YAML 中缩进错误、重复键或手工覆盖整段配置，都可能制造新的问题。优先使用受支持的配置命令，只在 CLI 无法完成时做最小范围的手动编辑。

### 4. 单独验证运行状态

```bash
hermes -p default gateway status
hermes -p hermes2 gateway status
```

若配置只会在启动时读取，先在可控条件下只重启目标 Profile，再查看状态和日志。不要因为修改了 hermes2 的配置，就顺带重启 default。

### 5. 保存变更记录

记录日期、目标 Profile、字段、旧值、目标值、执行者和验证结果。不要记录密钥原文。这样后续发生故障时，可以回答“哪个 Profile 在什么时候变更了什么”，而不是只剩下难以解释的配置差异。

## 修改后必须验证文件和运行状态

配置命令成功，只表示 CLI 接受了请求，不等于你改对了 Profile。应同时验证目标文件和运行状态。

在 macOS 或 Linux 中可使用：

```bash
grep -n 'multiplex_profiles' \
  "$HOME/.hermes/config.yaml" \
  "$HOME/.hermes/profiles/hermes2/config.yaml"
```

预期输出是两个路径各有一行，且值符合预期。若只出现一个路径，或两行都指向同一个 Profile，应先停止 Gateway 排查，不要继续启动服务。

然后分别确认状态：

```bash
hermes -p default gateway status
hermes -p hermes2 gateway status
```

这一步能发现“配置改对了，但此前启动的进程仍在使用旧状态”之类的问题。

## 如何安全地比较两个 Profile 的配置

排错时经常需要比较 default 与 hermes2 的配置，但直接分享完整 `config.yaml` 可能泄露模型密钥、机器人 token、内部地址或工作目录。更安全的方法是只列出与故障有关的字段，例如：

```text
gateway.multiplex_profiles
启用的消息平台名称
模型提供方名称和模型标识
terminal.cwd 是否已设置
Gateway 当前状态
```

对凭据只记录“已设置”“未设置”“与另一 Profile 是否相同”这样的判断，不记录值本身。若确实需要确认两份 token 是否相同，应在本机通过受控方式完成比较，不要把 token 粘贴到聊天、日志或截图中。

比较结果也应区分事实和判断：

```text
事实：default 与 hermes2 都启用了某平台。
事实：两个 Profile 的 Gateway 在同一分钟内退出。
判断：需要继续检查是否存在凭据、端口或 multiplex 配置冲突。
```

这样不会把“两个配置看起来像”误写成根因已经确认。

## 什么时候应停止继续改配置

出现下列任一情况时，先停止批量修改，回到状态和日志：

- 同一字段写入后仍显示不同路径或不同值；
- 不带 `-p` 的命令和带 `-p` 的命令输出发生矛盾；
- Gateway 因配置校验失败立即退出；
- 你无法确定当前运行的进程属于哪个 Profile；
- 配置文件中出现不认识的字段、重复段或手工合并残留。

此时继续“把所有地方都设成 false”并不能提高成功率，只会丢失因果关系。先导出 Profile、记录现状，再逐个检查才有意义。

## Profile、工作区和权限不是一回事

每个 Hermes Profile 有独立的状态目录，包括配置、凭据、记忆、会话、技能、定时任务和 Gateway 状态。但 Profile 不会自动限制 Agent 可访问的项目目录或文件权限。

官方文档将三者明确区分：Profile 管理 Hermes 状态；工作目录由 `terminal.cwd` 控制；沙箱才负责限制文件系统访问。[Profiles 用户指南](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/docs/user-guide/profiles.md?plain=1)

因此，修复“配置写错 Profile”不会自动解决工作目录、模型密钥或机器人权限的问题。它只保证接下来的命令作用于正确的状态目录。

## 新建或克隆 Profile 时的检查清单

```text
[ ] 使用 hermes profile list 确认当前活跃 Profile。
[ ] 对关键配置使用 hermes -p <name> 显式指定目标。
[ ] 修改后检查每个 Profile 的 config.yaml，而不是只看命令成功提示。
[ ] 启动前分别检查两个 Profile 的 Gateway 状态。
[ ] 克隆 Profile 后检查 .env，避免无意复制机器人凭据。
[ ] 未确认运行状态前，不删除 Profile 或手动清理状态文件。
```

## 用 `profile show` 核对“目标”而不是猜路径

当怀疑配置写错目录时，最容易犯的错误是直接在文件系统里按经验寻找 `~/.hermes`，然后编辑第一个看起来像配置的文件。官方 `hermes profile show <name>` 会显示 Profile 的 Hermes home、配置模型、Gateway 状态、skill 数量、`.env` 与 `SOUL.md` 是否存在。它显示的路径是 Profile 的 Hermes 主目录，不是工具命令的终端工作目录。[Profile 命令参考](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/i18n/zh-Hans/docusaurus-plugin-content-docs/current/reference/profile-commands.md?plain=1)

因此，在编辑前后都可以运行：

```bash
hermes profile show default
hermes profile show hermes2
```

把输出中的 `Path`、`Model`、`Gateway` 和 `.env` 状态写入变更记录。若你想改的是 `hermes2`，但 `show hermes2` 显示的模型或路径不是预期目标，就先停止，不要继续寻找“正确的 config.yaml”。先确认 Profile 名称、创建方式和当前部署是否一致，才能避免在错误目录里做一系列正确但无效的修改。

终端工作目录应另行由 `terminal.cwd` 管理。Profile 目录、项目目录和操作系统用户目录经常碰巧位于相邻路径，因而特别容易混淆。配置里的 `terminal.cwd` 可能决定工具从哪里启动，但不会把 Hermes 的 state、sessions 或 Gateway 文件移动到那里。遇到“工具在 A 项目执行，配置却在 B Profile 生效”的现象时，分别检查这两个层次，而不是试图用切换目录来改变活跃 Profile。

## 创建方式决定初始状态，不能只看新名称

`hermes profile create` 的不同选项会带来不同的初始内容。空白创建会提供新的 Profile 与内置技能；`--clone` 会复制当前 Profile 的 `config.yaml`、`.env`、`SOUL.md` 和 skills，但会使用新的 sessions 与 memory；`--clone-all` 则还会复制更多内容，同时排除 sessions、`state.db`、backups、state snapshots 与 checkpoints 等每 Profile 的历史状态。也可以通过 `--clone-from <profile>` 指定来源。[Profile 命令参考](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/i18n/zh-Hans/docusaurus-plugin-content-docs/current/reference/profile-commands.md?plain=1)

这意味着“新建了一个 Profile”不足以说明它是否携带旧 token、旧模型路由、旧技能和旧人格。配置改错目录的根源有时不是命令目标错了，而是预期以为空白 Profile，实际却从当前活跃 Profile 克隆，随后又在错误副本上修改。

创建或克隆后，先记录三件事：来源 Profile、使用的 create 参数、预期继承哪些文件。接着以显式 Profile 查看配置，再逐项修改模型、机器人凭据、`terminal.cwd` 和仅该 Profile 需要的 skill。若任务只需一个很窄的 Profile，可考虑 `--no-skills` 创建空 profile，避免在未理解继承内容时把完整 skill 集合带入新环境；这项选择不能与 clone 选项组合。

## 为配置变更设置可回退的最小单元

一次变更最好只涉及一个 Profile、一个键和一个验证动作。例如先将 `hermes2` 的 `gateway.multiplex_profiles` 设为目标值，再检查 `hermes -p hermes2 gateway status` 和一条低风险测试消息；确认后才改 default。即使两个 Profile 最终需要相同值，也不要把两份配置放进一条无法区分结果的大操作中。

若修改内容涉及模型、凭据或大量技能，不妨先导出 Profile 作为恢复点，再做变更。`hermes profile export <name>` 生成本地 tar.gz，`import` 可导入为另一个名称用于比对或恢复。导出文件和配置一样应被视为受控数据，不要上传到公开位置。把“导出时间、变更前配置摘要、变更命令、验证结果”记录在一起，后续发现写错目录时才能快速恢复到已知状态。

## 常见的“改对了文件但仍未生效”原因

配置路径正确并不代表运行进程已经加载新值。以下现象需要分开判断：

| 现象 | 可能解释 | 核验方式 |
| --- | --- | --- |
| 文件内容正确，Gateway 行为没有变化 | 服务尚未重启，或命令操作的是另一 Profile | 显式检查目标 Gateway 状态与启动记录 |
| `profile show` 路径正确，工具仍在错误项目中运行 | `terminal.cwd` 指向了另一个工作目录 | 检查 Profile 配置与工具启动目录 |
| 新 Profile 出现旧机器人的凭据 | 创建时使用了 `--clone` 或 `--clone-from` | 仅核对变量是否存在和身份是否不同，不显示具体值 |
| 切换活跃 Profile 后命令结果变化 | 无 `-p` 命令依赖当前状态 | 将关键命令改为 `hermes -p <name>` 并复跑 |

遇到这些情况，不要再以“多重启几次”替代验证。将文件内容、目标 Profile、运行进程和业务验证逐一对应，才能确定配置到底是写错了、没有加载，还是载入后被其他配置覆盖。

## 结论

多 Profile 配置问题的第一步，不是重新安装或重写 YAML，而是确认命令写到了哪里。活跃 Profile 会接管所有不带 `-p` 的命令，显式 `-p` 则能把目标固定下来。

对 Gateway 这类会影响消息服务的配置，使用显式 Profile、修改后检查文件、启动前检查状态，可以避免一次“显示成功”的误操作演变成整套机器人不可用。
