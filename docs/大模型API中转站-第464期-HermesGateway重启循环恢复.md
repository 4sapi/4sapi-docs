---
title: "Hermes 出现 Gateway shutting down 循环如何恢复"
category: 人工智能
tags:
  - AI Agent
  - 故障恢复
  - 消息机器人
description: "给出 Gateway shutdown 循环时的低风险恢复流程：保留 Profile 数据、停止冲突服务、先恢复 default，并用状态和消息验证恢复结果。"
---

# Hermes 出现 Gateway shutting down 循环如何恢复

当消息窗口不断出现 `Gateway shutting down — Your current task will be interrupted.`，最危险的操作往往是不断执行 `start`、`restart`，或直接删除 Profile。前者会让日志和状态更混乱，后者可能永久丢失会话、记忆和配置。

更稳的恢复目标只有一个：在不删除 Profile 的前提下，先让原来的机器人恢复，再逐个排除新增 Profile、运行状态和系统服务的影响。下面给出一个适合 macOS 和 Linux 本地安装场景的低风险流程。

## 第一阶段：冻结变化，保留证据

先停止新增 Profile 的 Gateway。假设新增 Profile 名为 `hermes2`：

```bash
hermes -p hermes2 gateway stop
hermes -p hermes2 gateway status
```

然后保存当时的关键信息：

```bash
hermes profile list
hermes profile show default
hermes profile show hermes2
hermes -p default gateway status
hermes -p hermes2 gateway status
```

保留终端输出和日志时间戳。不要在这个阶段删除 `config.yaml`、`.env`、`gateway.pid` 或任何 lock 文件，因为它们可能正是定位退出原因的证据。

### 为什么先停止新增 Profile

恢复故障时，最重要的是建立一个可以解释的基线。新增 Profile 是最后一个变量，就应当先从运行集合中移除。这个操作不是认定“hermes2 一定有问题”，而是为了回答一个更基础的问题：在没有第二个 Gateway 的情况下，原来的 default 能否稳定工作？

如果答案是“能”，后续排查集中在新 Profile 的配置、凭据和服务；如果答案是“不能”，继续围绕 hermes2 操作只会分散注意力。这个分支原则适用于所有多实例系统：先还原到最近一次已知可用状态，再逐项引入变化。

### 用时间线而不是记忆做判断

在终端或工单中写一条简短时间线：

```text
10:02  default 正常回复。
10:10  创建或配置 hermes2。
10:14  启动 hermes2 Gateway。
10:15  default 收到首次 Gateway shutting down 提示。
10:18  停止 hermes2。
10:21  default 重新启动并完成测试。
```

时间线不需要精确到毫秒，但应区分“配置发生”“服务启动”“提示出现”“服务停止”“恢复验证”这几个事件。没有时间顺序，就很难判断是新 Profile 导致故障，还是恰好与一次升级、主机重启或网络变化同时发生。

## 先区分退出、重启和消息通知

聊天窗口里的 shutdown 文案只是用户可见的结果。恢复动作取决于实际运行状态，至少应区分三种情况：

| 状态 | 证据特征 | 恢复侧重点 |
| --- | --- | --- |
| Gateway 已退出且没有再次拉起 | `gateway status` 显示停止，进程不存在 | 查看退出原因和启动条件，不急于反复 start |
| Gateway 持续退出又被拉起 | 消息反复提示，进程或服务状态发生变化 | 检查持久服务、系统守护和配置触发条件 |
| Gateway 仍在运行但当前任务被中断 | 状态正常，日志出现会话或消息处理错误 | 检查任务、平台连接与会话层，而非直接重装服务 |

这三种状态的用户感受都可能是“机器人不回复”，但恢复方式完全不同。将它们混为一谈，容易出现用 `restart` 掩盖一次配置错误，随后又进入更难解释的循环。

## 第二阶段：先备份，再改状态

Profile 内含配置、记忆、会话和技能。官方支持使用 `hermes profile export <name>` 将一个 Profile 导出为本地 `tar.gz` 归档；这是可恢复的备份方式。[Profile 导入导出说明](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/i18n/zh-Hans/docusaurus-plugin-content-docs/current/reference/profile-commands.md?plain=1)

在修改配置或处理系统服务前，分别导出：

```bash
stamp=$(date +%Y%m%d-%H%M%S)
hermes profile export default -o "./default-$stamp.tar.gz"
hermes profile export hermes2 -o "./hermes2-$stamp.tar.gz"
```

确认两个归档文件都已生成，再继续操作。这个备份不会替代日志收集，但能避免恢复过程意外损坏 Profile 数据后没有回退点。

### 备份完成后检查什么

导出成功不代表可以忽略恢复边界。至少确认：归档文件存在且大小不是零；导出命令没有打印未处理的错误；导出位置位于你能控制访问权限的本地目录。归档可能包含配置、会话与记忆，不应上传到公开仓库、群聊或第三方工单附件。

如果 Profile 中有必须保留的机器人配置，导出前后都不要擅自编辑 `.env`。备份的目的在于回退，不是为了把凭据传播到更多位置。

## 第三阶段：只恢复 default

恢复过程中同时运行多个 Profile，会让你无法判断哪一个触发了退出。先将默认 Profile 设为活跃目标，然后仅启动它：

```bash
hermes profile use default
hermes gateway start
hermes gateway status
```

接下来只对原机器人发送一条简单消息，例如“回复 OK”。验收标准应包含：

- Gateway 状态持续正常，而不是刚启动就退出；
- 原机器人能收到并回复消息；
- 消息窗口不再新增 shutdown 提示；
- 日志中没有新的配置冲突或凭据冲突。

若这些条件不满足，保持 `hermes2` 停止，继续检查 default 的日志和服务状态。不要把新 Profile 启动回来“试试”，否则又会引入一个变量。

### default 恢复后的最小验收

单 Profile 恢复后，建议至少完成下面四项验收：

```text
[ ] 启动后 gateway status 处于预期状态。
[ ] 机器人能处理一条新的简单消息。
[ ] 原有会话不会持续追加 shutdown 提示。
[ ] 观察一段时间后状态仍稳定，日志没有新的退出原因。
```

第四项很重要。一次成功回复只能证明某个瞬间可用，不能证明后台服务不会在稍后退出。如果该机器人承载定时任务或外部消息，还应在不暴露业务数据的前提下验证一条对应的低风险流程。

## 第四阶段：检查是否有遗留 Gateway 服务

如果命令行显示已停止，但消息端仍在刷 shutdown，检查是否还有后台进程：

```bash
pgrep -fal 'hermes.*gateway'
```

在 macOS 上还可查看 launchd 中的 Hermes 服务：

```bash
launchctl list | grep -i hermes
```

输出只能用来确定下一步目标，不能直接套用其他机器的 PID 或服务标签。应先确认每个进程、服务标签对应哪个 Profile，再用 `hermes -p <name> gateway stop` 停止目标。只有 Hermes CLI 无法停止且服务归属已经确认时，才在本机运维规范允许的范围内使用系统服务管理命令。

### 处理遗留状态时的底线

`pid`、`lock`、状态快照和系统服务记录看上去都像“可以清掉的临时文件”，但它们的作用取决于当前版本、退出方式和服务管理器。没有对应日志或官方恢复指引时，不要用通配符删除整个 Hermes 状态目录，也不要为了让服务启动而直接清空 lock 文件。

更安全的顺序是：确认 Gateway 已停止，先导出 Profile，再记录文件存在性与时间戳，最后只针对明确的、已验证的异常状态处理。若无法确认某个状态文件是否可安全移除，保留它并寻求当前版本的官方支持，比猜测性清理更可恢复。

## 第五阶段：确认配置冲突已解除

若故障出现在新增 Profile 后，检查每个 Profile 的 `gateway.multiplex_profiles`。Hermes 官方 issue #52796 记录过该配置开启后，次级 Profile 的端口绑定平台配置会导致整个 Gateway 退出的情况。[故障报告](https://github.com/NousResearch/hermes-agent/issues/52796)

在已确认配置适用于当前版本的前提下，使用显式 Profile 修改：

```bash
hermes -p default config set gateway.multiplex_profiles false
hermes -p hermes2 config set gateway.multiplex_profiles false
```

再分别检查两份 `config.yaml`。不要只执行不带 `-p` 的命令，因为它可能只改到当前活跃 Profile。

## 根据结果选择下一步

恢复流程走到这里后，通常会落入四种结果之一：

| 结果 | 说明 | 下一步 |
| --- | --- | --- |
| default 单独稳定，启动 hermes2 后复发 | 第二个 Profile 或其服务是高优先级排查对象 | 保持 hermes2 停止，比较配置、凭据和启动日志 |
| default 单独仍不稳定 | 不能把原因归给 hermes2 | 只收集 default 的退出原因与服务状态 |
| 两个 Profile 都能运行但复现过一次异常 | 当前已恢复，不代表根因已消失 | 保留时间线和日志，观察一段时间后再考虑扩展 |
| 两个 Profile 都不能启动 | 可能是安装、主机、全局服务或共享外部依赖问题 | 检查版本、系统服务和平台侧状态，不要继续删除 Profile |

这种结果表能避免“恢复成功”被误读为“根因已修复”。对生产使用而言，恢复服务和解释故障是两个独立目标；前者可以先完成，后者需要更多证据。

## 恢复完成后的复盘记录

即使没有立即写正式事故报告，也建议保留一页短复盘：触发条件、影响范围、实际命令、证据、临时恢复方式、尚未证实的假设和下次预防措施。

例如，这次问题可以把“多 Profile 下，未显式指定目标 Profile 的配置命令可能写入错误目录”作为已确认过程问题；而“某个具体设置一定造成 Gateway 退出”只有在日志与复现条件充分时才写为根因。这样的区分能让后续文章和运维交接保持准确。

## 哪些操作不该做

```text
不要为了恢复而直接删除 hermes2。
不要在未备份前删除 pid、lock、state 或配置文件。
不要让 default 和新 Profile 在未知配置下同时重启。
不要复制其他机器的 PID、launchd 标签或路径来执行 stop/kill。
不要把 API Key、机器人 token 或完整 .env 发到聊天记录和 Issue 中。
```

## 导出备份与发布发行版不是同一件事

故障恢复时，`hermes profile export` 与“把一个 Agent 发给别人安装”容易被混为一谈。官方文档将 `export` / `import` 定位为本机 Profile 的备份和恢复路径；Profile distribution 的 `install` / `update` 则通过 Git 仓库分发可版本化的配置、SOUL、skills、cron 和 MCP 连接，两者的安全模型不同。[Profile 命令参考](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/i18n/zh-Hans/docusaurus-plugin-content-docs/current/reference/profile-commands.md?plain=1)

出现 Gateway 重启循环时，应优先使用本地导出建立恢复点，而不是把当前目录初始化为 Git 仓库后推送出去。导出是为了让你能够在同一台机器或受控环境中恢复到故障前状态；发行版则用于长期维护可共享的 Agent 构成，不应该承载会话、记忆、临时状态或机器相关凭据。

官方用户指南说明导出归档会移除 API key，但这不意味着归档可随意公开。Profile 仍可能包含 persona、技能、连接配置、cron 定义、项目路径、会话相关信息或其他业务上下文。恢复记录中应写清归档保存位置、创建时间和访问范围；需要比对时在本机解压到新的临时位置，不要把 archive 当作日志附件上传。

## 避免让服务管理器和手动命令互相拉扯

`gateway install` 创建的是持久服务，在 host 安装上通常由 systemd 或 launchd 管理；Docker 部署则由镜像内的 s6-overlay 监督。手动在前台启动 Gateway、同时又让系统服务自动重启，可能产生两份进程、交错日志和难以解释的 shutdown 提示。[Profiles 用户指南](https://github.com/NousResearch/hermes-agent/blob/fc7523ca31eeb6eff9114afe384c2cf6380359df/website/docs/user-guide/profiles.md?plain=1)

恢复前应先明确当前运行方式：

```text
是通过 hermes gateway install 安装的持久服务？
是临时前台启动的 Gateway？
是在 Docker 容器里由 s6-overlay 监督？
是否还有另一套进程管理器在重启同一 Profile？
```

不同运行方式的日志、停止动作和重启策略不同。不能因为 `pgrep` 看到了一个进程，就立即使用系统 `kill`；也不能因为 `gateway status` 一时显示 stopped，就假定外部监督器不会在几秒后再拉起它。先把 Profile 与服务管理器的对应关系写清，再用该运行方式的正常停止路径处理，才能避免两个管理器互相竞争。

## 用“稳定窗口”代替一次性成功判断恢复

恢复后连续观察多久，不应靠感觉决定。可以根据机器人平时收到消息、定时任务或外部事件的频率定义一个稳定窗口。例如，若系统每十分钟至少有一次低风险事件，观察期应覆盖多个这样的周期；若没有自动事件，也至少在启动后分别发送两条间隔一段时间的测试消息，并在每次后检查 Gateway 状态和新增日志。

稳定窗口内记录的不是所有日志，而是关键状态：是否发生新的退出、服务是否重启、两个 Profile 是否产生交叉影响、消息是否由预期机器人回复。结束时把“观察了多久、触发了哪些低风险验证、是否有重试或错误”写入复盘。这样下一次同类故障时，团队可以比较恢复是否变得更稳定，而不是只知道某次“当时能回复”。

如果 default 在稳定窗口内持续可用，而一启动 hermes2 就复现问题，可以先把 default 的单 Profile 恢复视为完成，把多 Profile 兼容性作为单独的后续问题。把服务恢复与根因确认拆开，能防止为了追查新增 Profile 而再次影响主机器人。

## 将恢复过程转化为可执行的演练

事故真正发生时，人最容易跳过备份、时间线和逐步验证。对于依赖机器人处理日常任务的环境，可以在非生产 Profile 上定期演练一次“停止次级 Profile、恢复 default、验证消息、重新引入次级 Profile”的流程。演练不需要故意制造故障，只需要确保每条命令的目标、输出和回退条件都仍适用于当前版本。

演练后更新一份本机 Runbook：Profile 名称、服务管理方式、备份位置、低风险测试消息、日志位置、谁有权改凭据或停止服务。它不是冗余文档，而是让未来真正出现重启循环时可以少做几次猜测性操作的依据。所有敏感值只写变量名或安全存储位置，不写入 Runbook 正文。

## 结论

Gateway shutdown 循环的恢复原则是减少变量：保留证据，导出备份，停止新增 Profile，只恢复 default，最后才检查并重新引入第二个 Profile。

即使最终确认是配置冲突，也不需要删除用户数据。先把可用服务恢复到单 Profile 状态，再继续处理配置和系统服务，排错路径会清晰得多。
