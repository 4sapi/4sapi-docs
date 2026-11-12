---
title: "Codex app-server 协议拆解：从初始化到回合事件"
category: 人工智能
tags:
  - Codex
  - 通信协议
  - 开发工具
description: "通过本机 app-server 初始化实测和版本化 Schema，解释 stdio 消息、请求 ID、通知和线程回合生命周期，并说明最小客户端应怎样处理异步事件。"
brand_mode: none
---

# Codex app-server 协议拆解：从初始化到回合事件

想给 Codex 做一个自己的界面，第一步通常不是分析窗口按钮，而是确定客户端怎样启动运行时、发送请求和接收事件。
如果把服务端回了一条消息当成“任务已经完成”，界面很容易提前结束，或者在等待审批时卡住。

本文围绕 app-server 的 stdio 接入，解释初始化、请求与通知、线程和回合生命周期。
其中初始化握手已在本机执行；模型回合、工具审批和断线恢复只给出分析与验证方法。
文章不把一个成功握手包装成完整客户端已验证。

核对日期为 2026-10-20，本机版本是 `codex-cli 0.141.0`。
公开源码另固定到 `46d252f8bf8b09f4ef765bc72e40f02c229da3e6`，不声称它与本机二进制完全对应。

## app-server 位于哪一层

公开仓库中的 `codex-app-server-client` 为会话式 CLI 表面共享启动与生命周期管理。
其 README 列出 `codex-exec` 和 `codex-tui`，并说明进程内路径使用类型化通道，外部 stdio/WebSocket 边界使用 JSON 序列化。
来源：[codex-app-server-client README](https://github.com/openai/codex/tree/46d252f8bf8b09f4ef765bc72e40f02c229da3e6/codex-rs/app-server-client)。

这意味着不能从“用了 app-server”，推断每一种入口都额外启动一个外部进程。
进程模型和消息语义是两件事。
自己的客户端选择 stdio 子进程，只是其中一种接入形态。

可以用三层分析图理解边界：

```text
自定义界面或 CLI 客户端
    ↕ app-server 请求、响应与通知
Codex 运行时
    ↕ 模型请求、工具调度与结果
模型服务 / 本地工具 / 外部工具
```

图中省略了持久化、认证与扩展路径。
它只强调：界面协议不是模型接口本身。
抓到一条 `turn/start`，不能把它直接当成发给模型的完整请求。

## 用当前二进制生成 Schema

不要先照着旧博客手写全部字段。
在独立研究目录运行：

```powershell
codex --version
codex app-server --help
codex app-server generate-json-schema --help
codex app-server generate-json-schema --out .\schema
```

本机实际执行了以上命令，Schema 导出成功。
输出中包含 `ClientRequest.json`、`ClientNotification.json`、`ServerNotification.json` 和按版本组织的参数文件。
这里的文件名是本机生成结果，不是对未来版本的命名承诺。

例如，可从这些文件查看字段与消息方向：

| 文件 | 核对内容 |
| --- | --- |
| `v1/InitializeParams.json` | 初始化参数与能力声明 |
| `ClientNotification.json` | 客户端发出的通知 |
| `v2/ThreadStartParams.json` | 新线程参数 |
| `v2/TurnStartParams.json` | 回合输入与线程身份 |
| `ServerNotification.json` | 服务端通知名称与结构 |

固定源码也包含相应版本化定义。
来源：[公开协议 Schema 目录](https://github.com/openai/codex/tree/46d252f8bf8b09f4ef765bc72e40f02c229da3e6/codex-rs/app-server-protocol/schema/json)。

## 初始化不是第一条聊天消息

本机 `app-server --help` 显示，`stdio://` 是默认传输。
研究客户端可以显式启动 `codex app-server --listen stdio://`。
启动后，先完成协议初始化，再发送业务请求。

本次实际发送的初始化消息如下，每条 JSON 后写入换行：

```json
{"id":1,"method":"initialize","params":{"clientInfo":{"name":"codex-source-probe","version":"0.1.0"}}}
```

客户端收到对应 `id` 的成功响应后，再发通知：

```json
{"method":"initialized"}
```

`clientInfo` 的 name 和 version 是初始化参数里的必要字段。
`initialized` 是通知，不带等待响应的请求 ID。
来源：[InitializeParams](https://github.com/openai/codex/blob/46d252f8bf8b09f4ef765bc72e40f02c229da3e6/codex-rs/app-server-protocol/schema/json/v1/InitializeParams.json)、[ClientNotification](https://github.com/openai/codex/blob/46d252f8bf8b09f4ef765bc72e40f02c229da3e6/codex-rs/app-server-protocol/schema/json/ClientNotification.json)。

本例沿用当前协议 Schema 的消息形式，没有自行补上其他协议习惯中的必需字段。
如果你接入的是不同服务或不同版本，仍应重新查看它自己的消息定义。

## 本机实测了什么

最小探针启动已安装的 Codex 二进制，分开读取 stdout 与 stderr。
它按行解析 stdout，将初始化请求和响应按 ID 配对，完成后发送 `initialized` 并关闭输入。
探针设置初始化超时，并在退出时保存消息和摘要。

实测摘要如下，未包含实际用户目录：

```json
{
  "initialize": "passed",
  "modelTurnSubmitted": false,
  "incomingMessages": 2,
  "resultKeys": [
    "userAgent",
    "codexHome",
    "platformFamily",
    "platformOs"
  ],
  "error": null
}
```

这里的 `incomingMessages: 2` 是一次运行的观察值，不能写成协议必须返回两条消息。
`modelTurnSubmitted: false` 表示探针没有发起模型任务。
因此结果只能证明该环境下的初始化链路通过，不能证明模型认证、工具执行或线程恢复已经正常。

运行实际脚本时，Windows 上最好向进程 API 传入原生可执行文件路径。
`codex.ps1`、`.cmd` 包装器与 `.exe` 的启动方式并不相同，应先用当前 shell 查明安装位置。
脚本参数应使用参数数组传递，避免把路径拼进一段 shell 字符串。

## 请求响应与通知需要同时处理

客户端至少面对三类消息：自己的请求对应的响应、没有请求 ID 的通知，以及服务端反向发起的请求。
固定提交的进程内客户端文档也区分 `ServerRequest`、`ServerNotification` 与兼容事件。
来源：[传输与事件模型](https://github.com/openai/codex/tree/46d252f8bf8b09f4ef765bc72e40f02c229da3e6/codex-rs/app-server-client)。

实现时可以采用下面的分流思路：

```text
有 method 且有 id
    → 服务端请求，交给对应处理器
有 method 且无 id
    → 通知，更新状态或展示
无 method 且有 id
    → 查找 pending 请求，处理 result 或 error
其他形态
    → 记录协议异常，不伪装为正常消息
```

这段是客户端设计建议，不是完整的协议解析器。
还应检查字段类型、错误对象、未知方法和连接状态。
特别要避免把服务端请求误当成自己请求的响应，只凭 ID 存在就完成等待。

## thread/start 与 turn/start 分别启动什么

生成的 Schema 将 `thread/start` 和 `turn/start` 定义为不同请求。
`TurnStartParams` 需要 `threadId` 和 `input`。
客户端应从实际返回的线程身份继续，不要自行构造一个看似合法的 ID。
来源：[ClientRequest Schema](https://github.com/openai/codex/blob/46d252f8bf8b09f4ef765bc72e40f02c229da3e6/codex-rs/app-server-protocol/schema/json/ClientRequest.json)。

下面只展示消息骨架，省略返回值处理；它没有在本文作为模型回合实测执行：

```json
{"id":2,"method":"thread/start","params":{"cwd":"C:\\research\\fixture","sandbox":"read-only","ephemeral":true}}
{"id":3,"method":"turn/start","params":{"threadId":"替换为实际返回的ID","input":[{"type":"text","text":"只回复一个短标记"}]}}
```

这里的 Windows 路径必须替换为存在的测试目录。
`sandbox` 与其他可选参数需要对照当前 Schema 和环境要求，不能用于覆盖组织强制策略。
发送顺序应是收到线程启动结果后，再发送回合请求，而不是把两行无条件一次性写出。

## 接受回合不代表完成回合

本机生成的通知定义包含 `turn/started`、`turn/completed`、`item/started`、`item/completed` 和 `item/agentMessage/delta`。
它们用于不同粒度的生命周期与内容更新。
来源：[ServerNotification Schema](https://github.com/openai/codex/blob/46d252f8bf8b09f4ef765bc72e40f02c229da3e6/codex-rs/app-server-protocol/schema/json/ServerNotification.json)。

客户端可以按线程、回合和 item 身份维护状态：

| 事件类别 | 客户端更新 | 不宜做的推断 |
| --- | --- | --- |
| 请求成功响应 | 请求已被处理 | 整个任务成功 |
| item 开始或增量 | 创建或追加内容 | 当前回合已结束 |
| item 完成 | 封存该 item | 用户目标已验收 |
| turn 终态 | 更新回合状态 | 所有修改业务正确 |

不要把所有文本增量拼进一个全局字符串。
至少按 item 组织，否则不同消息或并发线程容易串行混写。
终态还应检查状态和错误内容，而不是只检查某个通知名称出现过。

## 两个最小客户端经常漏掉的地方

首先是 stderr。
协议流与诊断输出应分开读取，否则一行普通日志就可能让 JSON 解析失败。
解析失败时保存诊断即可，不应把它改造成一条假的助手消息。

其次是反向请求。
只实现“发请求、等响应”的客户端，在服务端需要审批或其他交互时，可能永远等不到回合完成。
应依据对应版本的 `ServerRequest` 定义实现需要的处理；不支持的能力应明确报告。
来源：[协议定义目录](https://github.com/openai/codex/tree/46d252f8bf8b09f4ef765bc72e40f02c229da3e6/codex-rs/app-server-protocol/schema/json)。

初始化握手适合验证进程启动、消息分帧和请求配对。
后续要做可用客户端，还需分别验证模型回合、工具交互、取消与恢复，不能把这些结论一次性打包宣布通过。

app-server 分析的关键是尊重异步边界：响应说明一个请求的结果，事件描述持续发生的工作，实际文件与测试说明任务有没有达到目标。
把三者分开，界面才能正确呈现 Codex 正在做什么。
