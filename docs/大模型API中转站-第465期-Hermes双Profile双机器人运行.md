---
title: "Hermes 如何让两个 Profile 独立运行机器人"
category: 人工智能
tags:
  - AI Agent
  - 配置管理
  - 消息机器人
description: "说明两个 Hermes Profile 如何使用独立状态、凭据和 Gateway 服务运行，并给出从单机器人恢复到双机器人时的验证顺序。"
---

# Hermes 如何让两个 Profile 独立运行机器人

一个 Hermes Profile 用于日常助手，另一个用于专门任务或不同机器人，是合理的隔离方式。但多 Profile 不是“多建一个目录”就结束了：每个 Profile 的配置、凭据、Gateway 和运行状态都必须独立，启动顺序也要可验证。

本文解决一个实际部署问题：已有 `default` 机器人稳定运行时，如何再启动 `hermes2`，而不让第二个 Gateway 影响第一个。

## Profile 隔离了什么

Hermes 官方文档将 Profile 定义为独立的 Hermes home。每个 Profile 拥有自己的 `config.yaml`、`.env`、记忆、会话、技能、定时任务、状态数据库和 Gateway 状态。官方同时警告，两个 Agent 进程不应指向同一个 Profile 目录。[Profiles 用户指南](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/docs/user-guide/profiles.md?plain=1)

这意味着以下边界应保持独立：

| 项目 | default | hermes2 |
| --- | --- | --- |
| Hermes 状态目录 | `~/.hermes` | `~/.hermes/profiles/hermes2` |
| 模型与工具配置 | 自己的 `config.yaml` | 自己的 `config.yaml` |
| 机器人凭据 | 自己的 `.env` | 自己的 `.env` |
| Gateway | default 的进程或服务 | hermes2 的进程或服务 |

Profile 不是沙箱。它不会自动限制 Agent 访问的项目目录，也不会把同一份机器人凭据变成两套凭据。工作目录与权限仍需单独配置。

## 双机器人上线前先做设计选择

在创建第二个 Profile 前，先说明它和 default 的职责边界。否则即使两个 Gateway 都能启动，后续也会因为“两个机器人都能处理同一类消息”而难以维护。

一份最小设计表可以包含：

| 项目 | default | hermes2 |
| --- | --- | --- |
| 面向对象 | 日常用户或主机器人 | 特定任务或第二个机器人 |
| 消息平台身份 | 第一套机器人凭据 | 第二套机器人凭据 |
| 模型与技能 | 通用配置 | 仅保留实际需要的配置 |
| 工作目录 | 主项目或默认目录 | 指定的独立项目目录，或明确不使用终端 |
| 外部副作用 | 允许哪些动作 | 哪些动作需人工确认或默认禁止 |
| Gateway 维护者 | 谁负责观察和恢复 | 谁负责观察和恢复 |

这不是为了把简单部署做复杂，而是为了让故障发生时能够快速回答“哪个 Profile 应该继续运行”“哪个可以先停止”“两个机器人是否应共享同一套外部权限”。

如果第二个机器人只是临时试验，建议先让它只处理低风险、可恢复的消息，不要一开始就接入生产群、付费接口或不可逆工作流。

## 先验证 default，再引入 hermes2

双机器人部署最容易犯的错误是同时启动两个 Gateway，然后在混乱状态下猜谁影响了谁。更安全的顺序是：

1. 确认 default 正常回复；
2. 确认 hermes2 的配置和凭据独立；
3. 只启动 hermes2；
4. 分别测试两个机器人；
5. 任一异常时，只停止 hermes2 并保留 default。

先执行基线检查：

```bash
hermes -p default gateway status
hermes -p hermes2 gateway status
```

预期是 default 正常、hermes2 已停止。若 default 此时仍异常，不要继续安装或启动 hermes2。

### 记录已知可用的基线

开始双 Profile 部署前，保存 default 的基线：当前 Hermes 版本、`gateway status` 输出、机器人是否能回复、相关配置变更时间和测试消息结果。这个基线的作用不是做性能评测，而是为后续比较提供参照。

当启动 hermes2 后发生异常时，可以快速判断以下问题：default 在变更前是否已经不稳定；服务状态是何时变化的；新 Profile 是否真的改变了系统行为。没有基线，排错很容易退化为不断重试。

建议只记录脱敏信息：版本、Profile 名称、状态、时间和错误摘要。不要把完整 `.env`、token 或客户消息放进运维记录。

## 检查凭据和关键配置

新 Profile 如果通过 `hermes profile create --clone` 创建，会复制 `.env`。因此，在启动第二个机器人之前，必须确认 `hermes2` 使用自己的平台凭据和所需模型配置。不要在文章、截图或 Issue 中展示凭据内容。

至少检查：

```text
[ ] hermes2 的 .env 已配置不同的机器人凭据。
[ ] hermes2 的 config.yaml 没有意外继承不需要的平台配置。
[ ] default 与 hermes2 的 gateway.multiplex_profiles 已按当前版本的要求设置。
[ ] 两个 Profile 不会写入同一个 Hermes home。
```

官方文档说明，Profile 的命令别名本质上等同于 `hermes -p <name>`。为了让部署过程不依赖终端当前状态，下面示例全部使用 `-p`。[Profile 命令与别名](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/docs/user-guide/profiles.md?plain=1)

### 独立凭据不等于只换一个名称

两个不同的 Profile 若使用相同的机器人 token，仍可能争夺同一个消息平台身份。官方文档说明，支持的平台会对重复 token 进行锁定并提示冲突 Profile；这是一道保护，不是可以忽略的启动警告。[Profile 凭据与 token 锁](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/docs/user-guide/profiles.md?plain=1)

因此，应在本机检查以下事实，但不要展示具体值：

```text
default 的目标平台凭据是否已设置。
hermes2 的目标平台凭据是否已设置。
两套凭据是否属于不同机器人身份。
两个 Profile 是否都启用了并不需要的平台连接器。
```

对于从 default 克隆的 hermes2，尤其要假设 `.env` 和平台配置已经被复制，直到人工确认它们已按预期调整。不要因为“新 Profile 使用了新名称”就默认凭据也已经隔离。

## 将双 Gateway 上线拆成两个阶段

### 阶段一：验证第二个 Profile 能独立启动

在 default 已稳定、hermes2 尚未启动的前提下，先检查：

```bash
hermes -p hermes2 gateway status
hermes profile show hermes2
```

确认 hermes2 的配置路径、模型和 Gateway 状态符合预期。若状态或路径不符合预期，先处理 Profile 目标问题，不要进入服务安装。

### 阶段二：验证两个 Profile 能并行工作

只有 hermes2 能正常启动后，才同时检查两个状态：

```bash
hermes -p default gateway status
hermes -p hermes2 gateway status
```

之后分别发送两条具有唯一标记的测试消息。测试不应只看“有没有回复”，还应确认回复来自预期机器人、没有串到另一个 Profile 的会话中，并且两个 Gateway 都没有新增退出或重启通知。

若需要验证写文件、联网或调用外部系统，应将其放在第二轮验收。第一轮只验证身份、消息收发和 Gateway 稳定性，避免把平台身份问题、模型问题和外部副作用混在一次失败中。

## 启动 hermes2 的独立 Gateway

先创建 hermes2 的持久服务：

```bash
hermes -p hermes2 gateway install
```

官方文档说明，每个 Profile 会得到独立的 systemd 或 launchd 服务。安装不等于服务已经通过业务验证，仍需显式启动和检查：[持久 Gateway 服务](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/docs/user-guide/profiles.md?plain=1)

```bash
hermes -p hermes2 gateway start
hermes -p hermes2 gateway status
```

观察状态稳定后，分别向两个机器人发送可区分的测试消息，例如“default-test”和“hermes2-test”。验收结果应是：

- 原机器人由 default 回复；
- 新机器人由 hermes2 回复；
- 任一机器人都不出现 shutdown 或重启通知；
- 两个 Profile 的 Gateway 状态保持正常。

### 建议的验收顺序

可以按下面的顺序记录结果：

```text
1. default 在 hermes2 停止时回复 default-test。
2. hermes2 启动后，gateway status 显示预期状态。
3. hermes2 回复 hermes2-test。
4. default 再次回复 default-test-2，确认未被第二个 Gateway 影响。
5. 等待一段与实际使用相符的观察时间，再复查两个 Gateway 的状态和日志。
```

第 4 步不能省略。只验证新机器人能够回复，并不能证明旧机器人仍然拥有稳定的消息链路。第 5 步也很重要，它能发现“启动时成功、稍后被服务管理器停止或配置触发退出”的问题。

## 持久服务与日常维护

`gateway install` 创建的是持久服务。官方文档说明，每个 Profile 使用各自独立的 systemd 或 launchd 服务，因此更新、重启主机或修改配置后，都应分别检查两个 Profile 的服务状态，而不是假设“一个正常就代表另一个正常”。[持久 Gateway 服务](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/docs/user-guide/profiles.md?plain=1)

可将日常检查固定为：

```bash
hermes profile list
hermes -p default gateway status
hermes -p hermes2 gateway status
```

每次升级 Hermes、修改 Profile、变更机器人凭据或重启主机后，都执行一次。若只更新了第二个 Profile 的需求，不要顺带改动 default 的配置；两个 Profile 的独立服务正是用来限制这类变更范围的。

如果用 Docker 运行 Hermes，服务监督方式与本地 systemd 或 launchd 不同。官方文档说明官方镜像内的每个 Profile Gateway 由 s6-overlay 监督。此时应使用容器部署文档和容器日志验证，不应直接套用宿主机的 launchd 命令。[Docker 中的 Profile Gateway 监督](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/docs/user-guide/profiles.md?plain=1)

## 出现异常时的回退

如果启动 hermes2 后 default 立即失联、任一机器人刷 shutdown，先回退到单 Profile：

```bash
hermes -p hermes2 gateway stop
hermes -p default gateway status
```

不要删除 hermes2。保留它的配置和日志，检查凭据是否重复、`multiplex_profiles` 是否开启，以及系统中是否有旧 Gateway 服务仍在运行。确定根因并验证 default 稳定后，再尝试启动 hermes2。

### 回退是否成功的判断标准

停止 hermes2 后，至少确认三件事：

```text
[ ] hermes2 的 gateway status 已显示停止或不再运行。
[ ] default 的 gateway status 保持正常。
[ ] default 能再次收到并回复一条新消息，且观察期内没有新的 shutdown 提示。
```

只有三项同时满足，才能称为已回退到单 Profile 可用状态。此时 hermes2 的数据和配置仍被保留，后续可以在测试环境中继续分析，而不会阻塞原机器人服务。

## 双 Profile 运维清单

```text
[ ] 每个 Profile 有清晰且不同的用途说明。
[ ] 每个 Profile 的 config.yaml、.env 和 Gateway 状态均独立。
[ ] 两套机器人凭据属于不同身份，且从未在聊天或截图中泄露。
[ ] 关键命令使用 hermes -p <profile>，不依赖当前终端的活跃 Profile。
[ ] 上线前已有 default 单 Profile 的状态和消息基线。
[ ] hermes2 先单独验证，再并行验证两个机器人。
[ ] 出现异常时先停止 hermes2，并以 default 恢复为首要验收目标。
[ ] 升级、重启或修改凭据后，分别检查两个 Gateway。
```

## 双 Profile 的隔离边界要写到部署设计里

Profile 的核心边界是 `HERMES_HOME`，不是操作系统用户。两个 Profile 会拥有独立的 `config.yaml`、`.env`、记忆、会话、skills、cron、state database 与 Gateway 状态；但在宿主机默认设置中，工具子进程仍使用同一真实 `HOME`，因此外部 CLI 的登录态、SSH 配置、Git 身份和云平台凭据可能仍是共享的。[Profiles 用户指南](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/docs/user-guide/profiles.md?plain=1)

对于只需要不同机器人身份的场景，独立 `.env` 和独立 Gateway 往往足够。对于两个机器人必须以不同 GitHub、云账号或 SSH 身份执行工具的场景，则需要额外设计工具身份隔离。官方支持在某个 Profile 的 `config.yaml` 中设置 `terminal.home_mode: profile`，让工具进程的 `HOME` 使用 `{HERMES_HOME}/home`；这样会要求在该目录中重新准备或链接所需的 CLI 配置。

不要在两个 Profile 都已稳定运行时直接开启这一设置，然后把所有异常归为 Gateway 问题。它会改变工具寻找凭据、缓存、配置和证书的位置，应先在低风险 Profile 中验证单次 Git、SSH 或云 CLI 调用，再逐步迁移。部署设计表中应明确写出：哪些边界由 Profile 隔离，哪些外部身份仍共享，哪些操作需要人工确认。

## 克隆 Profile 后要处理“继承的能力”

`hermes profile create hermes2 --clone` 会从当前 Profile 复制 `config.yaml`、`.env`、`SOUL.md` 和 skills，并给新 Profile 提供新的 sessions 与 memory。`--clone-all` 复制范围更广，但仍排除 sessions、`state.db`、backups、state snapshots 和 checkpoints 等本地历史状态。[Profile 命令参考](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/i18n/zh-Hans/docusaurus-plugin-content-docs/current/reference/profile-commands.md?plain=1)

因此，克隆的安全默认是“能力和凭据先被继承，身份隔离需要后续完成”。启动 hermes2 前，除机器人 token 外还应审查模型 endpoint、工具连接器、cron、MCP、项目目录和能够产生外部副作用的 skill。第二个机器人如果只承担窄任务，不一定要保留第一套 Profile 的完整 skill；可以在创建时使用空 profile，或在上线前删除并验证不需要的能力。

下面的表格适合作为 clone 后的第一次审查：

| 类别 | 应检查的问题 | 验证方式 |
| --- | --- | --- |
| 机器人身份 | 是否为不同平台账户或 token | 仅确认身份不同，不输出具体值 |
| 模型路由 | 是否使用预期模型与 endpoint | 查看目标 Profile 的脱敏配置摘要 |
| 工作目录 | 是否指向独立或明确允许的项目 | 核对 `terminal.cwd` 与工作区权限 |
| 自动任务 | 是否继承 cron、触发器或消息路由 | 列出任务名称和启用状态 |
| 工具能力 | 是否含不该由 hermes2 使用的 skill 或连接器 | 先禁用后在测试消息中验证 |

这比“看到 `hermes2` 目录存在就启动”多几分钟，但能避免第二个 Profile 启动后同时执行原机器人不该复制的计划任务或外部动作。

## 让每个机器人都有可辨认的健康检查

双机器人部署中的“两个都能回复”仍然过于粗糙。为每个 Profile 设计低风险、可辨认的健康检查，例如不同的测试前缀、不同的只读命令或不同的状态返回字段。发送消息时记录目标平台身份、期望 Profile、发送时间、返回内容的唯一标记和随后 `gateway status` 的结果。

避免让两个机器人都回答相同的泛化测试语句。若消息路由、token 或会话发生串扰，你可能看到一条回复却无法确认究竟由哪个 Profile 处理。使用 `default-health-<timestamp>` 与 `hermes2-health-<timestamp>` 这类唯一标记，既能在聊天窗口中区分，也能在日志中关联到对应 Gateway。

对于有定时任务或外部副作用的机器人，健康检查应只验证消息收发与只读状态，不能用真实发布、删除或付费操作作为“它工作了”的证据。复杂工作流应在单独的受控验证中测试，并保留回退路径。

## 更新与恢复时保持两套状态独立

Profile 的持久服务独立，并不意味着更新可以随意并行。升级 Hermes、修改模型配置或更换连接器时，建议先选一个非关键 Profile 作为验证对象：记录版本、备份目标 Profile、更新后启动 Gateway、运行健康检查、观察稳定窗口；确认无异常后再按同样流程处理另一个 Profile。这样一次升级失败只影响一个机器人，也更容易定位版本兼容性问题。

若某个 Profile 来自 profile distribution，`hermes profile update <name>` 会重新拉取发行版并更新发行版拥有的文件，同时保留本地 memories、sessions、auth 与 `.env`；默认不会覆盖本地 `config.yaml`，除非显式使用 `--force-config`。这与本地 `export` / `import` 的备份恢复不同，维护时不要把两种命令混用。升级前后分别执行 `profile show`、Gateway 状态检查和健康消息验证，才能确定更新没有改变另一个 Profile 的运行边界。

## 结论

两个 Hermes Profile 可以独立运行，但独立性来自独立状态、独立凭据和独立 Gateway，不是来自名称不同。部署时以 default 的稳定性为基线，第二个 Profile 每次只引入一个变化，并通过状态和消息双重验证。

这样即使第二个机器人配置有问题，也可以只停止它，而不会把已可用的机器人一起带入故障状态。
