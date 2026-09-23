---
title: "Hermes 新建 Profile 后机器人失联怎么排查"
category: 人工智能
tags:
  - AI Agent
  - 故障排查
  - 消息机器人
description: "当 Hermes 新建 Profile 后旧机器人反复提示 Gateway shutting down 时，如何区分配置冲突、进程退出与凭据复用，并按最小范围恢复服务。"
---

# Hermes 新建 Profile 后机器人失联怎么排查

新建一个 Hermes Profile 后，原本正常的机器人突然不回复，并反复发出 `Gateway shutting down — Your current task will be interrupted.`，很容易被误判为模型、飞书回调或网络问题。

这条消息的直接含义是 Gateway 正在结束当前任务并退出。排查重点应从“机器人为什么不说话”转为“哪个 Gateway 正在退出、为什么退出、是否还有另一个 Profile 触发了它”。下面给出一套优先恢复原机器人、再定位多 Profile 冲突的流程。

## 适用场景

以下现象属于适用范围：

- 新建或克隆 Profile 后，原机器人停止回复；
- 聊天窗口持续出现 `Gateway shutting down`；
- 同一机器上存在 `default` 和一个或多个额外 Profile；
- 问题发生在启动、安装或切换 Gateway 服务之后。

它不适用于纯粹的模型鉴权失败、单个 API 请求超时或消息平台回调配置错误。若 Gateway 进程始终正常运行，但消息平台日志提示签名、权限或回调地址错误，应转向对应平台的连接器排查。

## 先确认当前命令到底操作了哪个 Profile

Hermes 的 Profile 是独立的状态目录，每个 Profile 各自保存 `config.yaml`、`.env`、记忆、会话、技能和 Gateway 状态。官方文档明确说明：不带 `-p` 的 `hermes` 命令会操作当前活跃 Profile。[Profile 命令参考](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/i18n/zh-Hans/docusaurus-plugin-content-docs/current/reference/profile-commands.md?plain=1)

先执行：

```bash
hermes profile list
hermes profile show default
hermes profile show hermes2
```

`profile list` 中带 `*` 的名称就是当前活跃 Profile。不要在不确认它的情况下执行不带 `-p` 的 `config set`、`gateway start` 或 `gateway stop`，否则以为在改 `default`，实际可能改到了 `hermes2`。

接着分别查看 Gateway 状态：

```bash
hermes -p default gateway status
hermes -p hermes2 gateway status
```

记录每个命令的完整输出、运行时间和报错。此时不要反复重启两个 Gateway；连续重启会覆盖最有价值的退出线索。

## 多 Profile 场景的一个已知冲突条件

Hermes 官方仓库曾记录过一个已关闭问题：当 `gateway.multiplex_profiles` 开启，并且次级 Profile 配置了需要端口绑定的消息平台时，Gateway 会把配置冲突作为致命错误处理，导致整个 Gateway 进程退出。该 issue 中包含 Feishu 作为报错示例，报告环境为 Hermes `0.17.0` 桌面运行时和 macOS。[issue #52796](https://github.com/NousResearch/hermes-agent/issues/52796)

这不能证明每次 shutdown 都由同一原因造成，但足以说明多 Profile 环境应优先检查 `multiplex_profiles`。在两个 Profile 的 `config.yaml` 中确认该项：

```yaml
gateway:
  multiplex_profiles: false
```

如果当前版本提供配置命令，也应为每个目标 Profile 显式执行，避免依赖活跃 Profile：

```bash
hermes -p default config set gateway.multiplex_profiles false
hermes -p hermes2 config set gateway.multiplex_profiles false
```

然后使用 `grep` 或配置查看命令确认两个文件各自已经生效。不要只看到一条“Set”输出就认为两份配置都更新了。

## 不要把提示语当成根因

`Gateway shutting down` 是一个生命周期通知。它能确认当前 Gateway 已开始退出，却不能直接说明退出是谁触发的。把它当成“飞书插件坏了”或“模型没有回复”会把排查带偏。

实际处理时，至少要区分下面四类情况：

| 观察到的现象 | 更可能的方向 | 下一步应收集什么 |
| --- | --- | --- |
| `default` 单独启动后仍立即退出 | default 自身配置、凭据、升级或系统服务 | default 的 `gateway status`、退出前后的日志、是否有残留服务 |
| 只有启动 `hermes2` 后 default 才退出 | 多 Profile 配置、凭据或 Gateway 资源冲突 | 两份配置差异、两个 Profile 的状态、启动时间顺序 |
| CLI 显示已停止，消息端仍不断收到提示 | 遗留进程、持久服务或外部守护程序仍在重启 | 进程列表、服务列表、每次通知的时间间隔 |
| Gateway 保持运行但机器人不回复 | 消息平台连接、回调、权限或机器人凭据 | 平台连接日志、消息发送记录、平台侧配置 |

这张表的关键在于把“时间关系”当作证据。比如原机器人一直正常，第二个 Profile 一启动就出问题，优先排查新引入的配置和服务；如果停掉第二个 Profile 后原机器人仍然异常，就不能再把原因简单归到第二个 Profile 上。

## 用最小证据包保留现场

故障刚发生时，不需要立刻把所有文件都复制出来。先收集一个能够复现判断过程的最小证据包即可：

```text
1. Hermes 版本、操作系统和安装方式。
2. hermes profile list 的输出，标明当前活跃 Profile。
3. default 与 hermes2 的 gateway status 输出。
4. 新建、配置、安装或启动第二个 Profile 的命令顺序。
5. 首次出现 shutdown 提示的时间，以及随后出现的间隔。
6. 脱敏后的相关日志行：退出原因、平台名称和服务状态。
```

这里有两个边界。第一，不要上传 `.env`、机器人 token、API Key、完整会话和客户消息。第二，不要只截最后一行异常；Gateway 退出前的几十行日志往往比提示语更有价值，因为其中可能包含配置校验、端口占用或凭据锁定的原因。

若需要提交社区 issue，可以先将路径替换为 `<HOME>`，将凭据替换为 `<REDACTED>`，保留 Profile 名称、平台类型、时间顺序、命令结果和错误文本。这样既不丢失技术信息，也不会扩大敏感数据暴露面。

## 一次只改变一个变量

多 Profile 故障很容易出现“同时改了配置、换了 token、重装服务、切了 Profile”的情况。即使最后恢复了，也无法知道真正修复了什么。建议按下面的节奏进行：

1. 停止新 Profile，确认 default 是否能单独恢复。
2. 若不能恢复，只处理 default，不修改 hermes2。
3. 若能恢复，保持 default 运行，检查 hermes2 的凭据和配置。
4. 仅启动 hermes2，并分别发送两条测试消息。
5. 出现异常立即停止 hermes2，保存新的状态和日志，再分析差异。

每一步完成后都应记录“执行的命令、Gateway 状态、消息端是否回复”。这样故障即使再次出现，也能明确是在哪个状态转换点发生。

## 恢复前的配置核对清单

在重新尝试双 Profile 前，建议逐项核对：

```text
[ ] default 与 hermes2 都有独立的 config.yaml 和 .env。
[ ] 两个机器人没有复用同一份平台凭据。
[ ] 每个关键命令均通过 -p 指向目标 Profile，或已确认当前活跃 Profile。
[ ] 如果版本和配置条件适用，两个 Profile 的 gateway.multiplex_profiles 都已明确设置。
[ ] default 已在 hermes2 停止的情况下稳定回复至少一次。
[ ] 已保存当前的状态输出和脱敏日志，出现回归时可比较。
```

这份清单不会替代平台侧的权限和回调检查，但能够排除本地多 Profile 管理中的大多数低级变量。

## 恢复顺序：先让 default 单独工作

先停止新 Profile 的 Gateway，避免它继续影响诊断：

```bash
hermes -p hermes2 gateway stop
hermes -p hermes2 gateway status
```

然后仅恢复原来的 `default`：

```bash
hermes profile use default
hermes gateway start
hermes gateway status
```

此时只测试原机器人。预期结果是它能够正常接收并回复一条简单消息，且不再发出 shutdown 通知。

如果 default 仍退出，收集 `gateway status` 输出和对应日志后再继续判断。不要为了“清空问题”直接删除 `hermes2`；Profile 删除会永久移除配置、记忆、会话和技能。官方文档将 `hermes profile delete` 标为不可恢复操作。[删除命令说明](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/i18n/zh-Hans/docusaurus-plugin-content-docs/current/reference/profile-commands.md?plain=1)

## 还应检查什么

### 克隆时是否复制了机器人凭据

`hermes profile create --clone` 会复制当前 Profile 的 `config.yaml`、`.env`、`SOUL.md` 和技能。若新 Profile 从旧 Profile 克隆，必须在启动前检查新 Profile 的 `.env` 和平台配置，确保它使用独立的机器人凭据。[Profiles 用户指南](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/docs/user-guide/profiles.md?plain=1)

### 是否遗留了运行中的服务

在 macOS 上，如果 `gateway stop` 后仍出现 shutdown 循环，可检查是否有遗留进程或 launchd 服务：

```bash
pgrep -fal 'hermes.*gateway'
launchctl list | grep -i hermes
```

先用 Hermes 的 `gateway stop` 停止对应 Profile。只有在确认某个服务标签和目标 Profile 的对应关系后，才考虑使用系统服务管理命令停止它；不要根据其他机器的服务名或 PID 直接执行 `kill`、`bootout` 或删除文件。

### 确认不是“短暂恢复”

`gateway status` 只是一时刻的状态，不等于服务已稳定。恢复 default 后，除了立即发送一条测试消息，还应等待一段与实际业务相符的观察时间，再检查一次状态和日志。若机器人能回复首条消息，却在稍后再次发出 shutdown 提示，应把“第二次退出的时间”和前后的日志单独记录。

这能帮助区分“配置校验后立即失败”和“服务运行一段时间后被外部进程停止”两类问题。两者的排查入口不同，前者重点看配置与凭据，后者重点看进程、服务和主机级管理策略。

## Profile 隔离 Hermes 状态，不一定隔离外部 CLI 身份

新建 Profile 后，很多人会认为两个 Agent 已经在两个完全独立的用户环境中运行。实际边界更细：Profile 通过 `HERMES_HOME` 隔离 Hermes 自身的数据，包括配置、`.env`、记忆、会话、技能、日志、定时任务和 Gateway 状态；但在宿主机安装中，工具子进程默认仍使用真实操作系统用户的 `HOME`。因此 `git`、`ssh`、`gh`、云 CLI、npm 或其他开发工具可能继续读取同一份用户级凭据。[Profiles 用户指南](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/docs/user-guide/profiles.md?plain=1)

这能解释一类表面上像“新 Profile 抢了旧机器人”的问题：两个 Profile 的 Hermes 配置完全不同，但它们调用外部工具时仍使用同一套 SSH、GitHub 或云平台身份。它未必会导致 Gateway shutdown，却可能让两个机器人操作同一个外部账号、写入相同的远端资源，或者在日志里显示相同的第三方身份。

需要严格隔离工具身份时，官方文档提供 `terminal.home_mode: profile`。启用后，工具进程使用 `{HERMES_HOME}/home` 作为 `HOME`；相应地，SSH、Git、云 CLI、npm 以及其他工具需要在该 profile home 中重新初始化或链接配置。不要在故障恢复期间随手开启它，因为这会同时改变子进程能找到的凭据与配置，属于新的变量。先恢复 Gateway，再在测试 Profile 上单独验证该隔离策略。

这也说明 Gateway 故障排查要区分两条证据线：一条是 Hermes Profile 内的配置和服务状态，另一条是工具执行时使用的系统账号与外部身份。两者相关但不等价。排查机器人失联时先证明 Gateway 是否退出；只有在日志显示外部工具或凭据触发错误时，才进入 `HOME`、CLI 身份和工具环境的检查。

## 使用显式 Profile 命令建立可比较的时间线

多 Profile 环境里，不带 `-p` 的命令会依赖当前活跃 Profile。临时切换 `hermes profile use default` 虽然可以恢复默认目标，但它也会改变随后所有终端命令的含义。为了让故障时间线可复查，建议在排查期间把关键操作都写成显式形式：

```bash
hermes -p default gateway status
hermes -p hermes2 gateway status
hermes -p default config edit
hermes -p hermes2 config edit
```

`-p` 只在当前命令中覆盖活跃 Profile，不会改变终端的长期状态。它特别适合把两组状态成对采集，避免先切到 `hermes2` 查看问题、后面忘记切回 default，又把恢复命令施加到错误目标。每条记录旁可以标注 Profile、时间、命令退出码、Gateway 状态和对应的消息测试结果。

如果要排查服务重启循环，不要只记录“我重启过几次”。一份有用的时间线应该像下面这样：

```text
10:05  default 正常回复测试消息
10:08  创建或更新 hermes2
10:10  hermes2 Gateway 启动
10:11  default 首次出现 shutdown 提示
10:12  hermes2 Gateway 停止
10:14  default 状态与消息测试结果
```

时间线并不自动证明因果关系，但能帮助你设计下一次最小复现：在 default 稳定的基础上，只重做 10:08 或 10:10 的其中一步，观察异常是否重现。比起同时重装服务、改 token 和删状态文件，这种方法更接近可证伪的排查。

## 什么时候应转向平台侧排查

这套流程优先处理本机 Profile 和 Gateway 生命周期。出现以下情况时，应停止继续改 Hermes 本地配置，转向消息平台、网络或模型接口的日志：

| 本机观察 | 更合适的排查方向 |
| --- | --- |
| Gateway 持续运行、没有退出或重启记录，但任何消息都未到达 | 平台 webhook、事件订阅、网络连通性或机器人权限 |
| Gateway 已收到消息，却在调用模型接口时失败 | 模型 endpoint、认证、配额、超时与请求日志 |
| 只有某类群聊、私聊或特定用户没有响应 | 平台授权范围、消息过滤规则和会话权限 |
| 两个 Profile 都稳定，且都能回复单独测试消息 | 不再是多 Profile 生命周期故障；检查业务路由与插件行为 |

这一步的意义是阻止故障排查无限围绕 Profile 打转。只有本机证据指向 Gateway 退出、配置冲突、服务重启或凭据锁定时，才继续修改多 Profile 配置；否则应把问题交给对应边界的日志和文档。

## 结论

新建 Profile 后旧机器人失联时，先确认活跃 Profile，再分别检查两个 Gateway 的状态。`Gateway shutting down` 是运行时退出信号，不是模型回答内容。

恢复时坚持“先停止新 Profile、先恢复 default、只验证一个机器人”的顺序。确认原服务稳定后，再处理多 Profile 配置、独立凭据和新 Gateway 的启动，这样才能把故障范围控制在一个变量内。
