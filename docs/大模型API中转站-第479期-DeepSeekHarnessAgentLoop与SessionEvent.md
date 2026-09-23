---
title: "DeepSeek Harness 的 Agent Loop 如何完成一轮任务"
category: 人工智能
tags:
  - DeepSeek Harness
  - AI Agent
  - 可观测性
description: "拆开 Agent Loop 中 Turn、Step、工具调用和 SessionEvent 的关系，说明一次任务如何持续推进、何时停止，以及页面刷新后如何恢复。"
---

# DeepSeek Harness 的 Agent Loop 如何完成一轮任务

看到 Agent 最后返回“已完成”，并不能说明中间发生了什么。它可能只请求了一次模型，也可能经历了读取文件、执行命令、处理错误和再次请求模型等多个步骤。要解释这些行为，必须把 Agent Loop、Turn、Step 和 SessionEvent 分开。

这里聚焦一次任务从进入会话到结束的运行链路。内容依据 DeepSeek Harness 官方仓库的 Agent Loop 和事件溯源设计，信息核对日期为 2026 年 8 月 25 日；具体事件类型可能随版本增加或调整。

## 三个对象先分清

```text
Agent：持有一个会话和自己的工具、提示词作用域
Turn：一次被接受的任务处理轮次
Step：Turn 中的一次模型判断及其后续动作
```

例如，“运行测试并修复失败”可能是一个 Turn，但里面包含多个 Step：

```text
Step 1：请求模型，决定先运行测试
Step 2：执行测试，读取失败结果
Step 3：请求模型，定位相关文件
Step 4：编辑文件并重新测试
Step 5：请求模型，总结改动和验证结果
```

如果模型第一次就能回答，Turn 可能只有一个 Step。工具调用不会天然开启新的 Turn，它通常会让当前 Turn 继续推进。

Agent 本身也不是一次性的函数调用。创建 Agent 时，Harness 会把会话、Agent 作用域和驱动器放进一个受回滚保护的创建过程；恢复时则先从持久化会话加载历史，再重建同一条消息和轮次边界。这样，Agent Loop 关注的是“从哪一份会话事实继续”，而不是凭当前 UI 的摘要猜测过去发生了什么。

## 任务不会绕过 Agent Loop

Web UI、CLI 和 API 的职责是接收任务与展示结果。它们不应直接把请求发给模型或直接执行工具，否则会绕开会话历史、权限检查和恢复逻辑。

一次普通请求可以抽象成：

```text
入口接收消息
  -> ctx.agents 找到活跃 Agent
  -> Agent Loop 领取 inbox 消息
  -> 打开 Turn，准备 Step
  -> 组装系统提示词、历史和工具 schema
  -> ctx.llm 发起模型请求
  -> 普通文本：记录并决定是否结束
  -> Tool Call：交给 ctx.tools 执行
  -> 工具结果写入会话
  -> 进入下一个 Step
```

模型负责当前一步的判断，Agent Loop 负责把一串判断组织成可停止、可恢复的任务。模型并不知道文件是否真的写入，只有工具返回并进入会话后，下一步才拥有这条事实。

这条边界也说明了为什么 UI 不应该自行“补写”工具结果。页面可以展示流式分片和工具卡片，但真正参与下一次模型请求的内容必须来自 SessionEvent。否则用户界面看到的状态、模型看到的历史和磁盘上的日志会各自形成一份事实。

## Turn 的开始与停止

在一个 Turn 内，循环需要处理三类信号：是否有新消息、工具结果是否要求继续、是否出现取消或错误。可以用下面的状态理解：

```text
空闲
  -> 收件箱有消息
  -> 领取并打开 Turn
  -> 进入一个或多个 Step
  -> 工具和模型都不再要求继续
  -> Turn 结算
  -> Agent 回到空闲
```

Turn 结算不等于 Agent 永久结束。Agent 仍可以保留在 `running` 区间，等待下一批任务；也可能在一次重试链结束后才报告最终 settle。把“这一轮结束”和“这个 Agent 已经停止”混为一谈，会让 UI 状态和故障排查都产生误读。

取消也应被当作可观察事实。循环会关闭当前 Turn，未分发的工具调用可能留下带错误结果的记录，后续恢复时模型才能知道这些调用并没有实际执行。

收件箱还有自己的生命周期，不能简单等同于 Turn：

```text
agent/inbox/inserted  消息进入待处理队列
agent/inbox/claimed   某个 Turn 原子领取消息
agent/inbox/discarded 消息被取消或丢弃
```

一条消息被领取，表示它已经归属于某次处理；不代表其中一定产生了 Step。`pre-step` 可以在真正打开步骤前拒绝这一批输入，取消也可能在模型请求开始前发生。把 inbox 的插入、领取、丢弃和 Turn 结束分开记录，才能解释“消息已收到但没有调用模型”的情况。

## SessionEvent 是唯一事实来源

普通聊天应用可以只保存用户问题和最终回答，但 Agent 还必须保存工具调用、参数、结果、错误和轮次边界。Harness 把这些内容写成只追加的 `SessionEvent` 日志，模型历史、UI 展示和持久化都从这份日志派生。

常见事件包括：

```text
turn/start       开始一轮任务
step/start       开始一次步骤
user/message     用户消息
assistant/message 模型组装后的消息
tool/call        工具调用
tool/result      工具返回
step/end         步骤结束
turn/end         轮次结束
```

当前版本的事件目录还包含 `request/header`、`approval/asked`、`approval/decided`、`compaction/start`、`compaction/end`、`llm/retry` 等类型。它们分别描述请求配置、审批决定、上下文压缩和模型重试。事件名是持久化协议的一部分，写文章或做解析时不应凭印象新增名称；应以当前版本的官方目录为准。

官方设计把 Session 视为唯一真源：模型下一次请求通过 `deriveMessages()` 从日志重建，而不是维护一份可能与日志分叉的可变消息数组。原始流分片可用于保持回放保真度，组装后的 assistant 消息则作为派生历史的依据。

## 为什么要先记日志，再继续流程

顺序本身就是协议的一部分。简化后可以这样理解：

```text
领取 inbox 消息
  -> 通过 pre-step 决定是否进入步骤
  -> 写入 step/start
  -> 追加 user/message 批次
  -> 请求模型并记录 assistant/message
  -> 发现 Tool Call 后写入 tool/call
  -> 工具完成后写入 tool/result
  -> 决定下一步或结束 Turn
```

如果工具先执行、日志后写入，进程在中途崩溃时就可能留下“工具已经产生副作用，但会话不知道它发生过”的缺口。先追加事实，再让下游从事实继续，是事件溯源设计的关键。

对流式模型，还要区分原始分片和组装后的消息。分片用于保留回放和计时信息，`assistant/message` 记录的是组装后、真正进入派生历史的内容。一个只收到空流、只有工具调用，或在模型故障前没有形成有效文本的步骤，不能简单用一条“成功回答”事件代替。

## 三个使用方共享同一份日志

同一条 SessionEvent 可以服务不同消费者：

| 使用方 | 从事件中得到什么 |
| --- | --- |
| 模型请求 | 重建下一次请求的消息历史、工具结果和上下文 |
| Web UI | 恢复聊天、工具卡片、错误和轮次状态 |
| 持久化插件 | 将事件写入 JSONL、SQLite 或其他存储 |

持久化通常是延后写入，热路径不会等待磁盘 I/O；轮次结束时的 flush 才是等待写入排空的检查点。因此“内存中已经追加”与“磁盘中已经持久化”是两个需要分别验证的事实。

这带来一个实际的故障窗口：进程在事件追加后、持久化 flush 前退出时，内存中已经看见的事实可能尚未写入后端。生产环境需要验证持久化插件的缓冲、flush 和重启恢复行为，不能只检查 UI 实时更新。JSONL、SQLite 等后端也必须携带会话头部中的关键元数据，例如会话 ID、工作目录和所选 Agent Preset，否则恢复出的会话可能无法重建原来的工具作用域。

## 从日志定位一次失败

遇到“Agent 说完成但结果不对”时，不要先追问模型。先按事件顺序查：

1. 是否真的出现 `turn/start` 和 `step/start`。
2. `assistant/message` 中是否包含预期的 Tool Call。
3. 对应的 `tool/call` 是否被实际分发，还是在取消前被阻止。
4. 是否存在匹配的 `tool/result`，结果是成功、错误还是超时。
5. 工具结果是否进入下一次模型请求的派生历史。
6. `turn/end` 的停止原因是什么，是否还有未领取消息。

这套检查能把问题拆成任务定义、模型选择、工具参数、权限、环境和持久化几类，而不是把所有失败都归因于“模型不聪明”。

可以进一步把一次失败画成事件时间线：

```text
turn/start
  -> step/start
  -> assistant/message(tool_call=run_tests)
  -> approval/asked
  -> approval/decided(deny)
  -> tool/result(code=DENIED)
  -> turn/end(reason=blocked)
```

这条链和“模型决定运行测试，但测试没有执行”是两种不同的事实。前者表示模型做出了 Tool Call，后者表示审批层拒绝了副作用。若只看最终回答，就无法判断是模型没有选择工具，还是工具被权限策略拦截。

## 如何做一个最小回归任务

可以在测试目录中准备一个不会影响外部系统的任务，例如读取一个固定文本文件并输出统计结果。验收材料至少包含：

```text
输入消息
预期工具调用
实际 SessionEvent 序列
最终输出
文件或命令的独立验证结果
```

第一次只测文本回答，第二次加入一个只读工具，第三次让工具返回可控错误，第四次在工具执行前取消任务。这样可以分别验证单 Step、多 Step、错误传播和取消记录，而不是用一个复杂项目同时引入未知变量。

回归任务最好固定四类输入：

| 输入类型 | 例子 | 主要验证点 |
| --- | --- | --- |
| 直接回答 | 要求输出固定格式的文本 | 单 Step 和 Turn 结束 |
| 只读工具 | 读取测试文件并统计行数 | Tool Call、结果回传和下一 Step |
| 可控错误 | 工具返回固定错误码 | 错误事件、重试或终态判断 |
| 取消 | 在工具确认前撤销任务 | inbox、Turn 结束和未分发调用记录 |

每次测试都保留输入、事件序列、最终输出和独立验收结果。模型版本、Profile/Preset、工作目录和权限范围也要记录，否则下一次重跑时无法区分配置变化和代码变化。

## 限制与边界

事件日志能回答“系统记录了什么”，不能自动证明外部副作用一定符合预期。文件写入、网络请求和发布操作仍需要独立的权限策略和验收命令。日志也可能包含敏感数据，应该优先记录 ID、路径、错误类别和脱敏摘要，而不是把密钥或客户原文完整复制进去。

另外，事件类型是版本化契约。恢复程序遇到未知且不可忽略的事件时，不能静默跳过，否则可能重建出错误的会话。发布或升级前，应用当前版本的官方事件目录和回归样例重新核对。

事件日志也不是完整的安全边界。工具参数可能包含路径、查询内容或敏感片段；只追加不代表可以无限期保存全部原文。对外部系统操作，日志应至少能关联任务、运行、工具、插件版本和执行身份，同时按数据规则保存引用、摘要或脱敏结果。

## 结论

Agent Loop 负责推进任务，Turn 划定一次任务轮次，Step 表示其中一次模型判断和工具动作，SessionEvent 则把每个事实按顺序保存下来。UI、模型历史和持久化都从同一份日志派生，系统才有机会在刷新、重启、取消和故障后恢复出一致状态。

如果要评估一个 Agent Harness，建议先做一条小而确定的事件链：一条消息、一次工具调用、一次结果和一次独立验收。能否把这条链记录清楚、重放清楚，往往比最终回答是否漂亮更能说明运行时是否可靠。

资料来源：

- [Agent Loop 官方包目录](https://github.com/deepseek-ai/deepseek-harness/tree/master/packages/core/agent-loop)
- [事件溯源设计所在目录](https://github.com/deepseek-ai/deepseek-harness/tree/master/.agents/notes/implemented/architecture)
- [Agent Loop 状态机设计所在目录](https://github.com/deepseek-ai/deepseek-harness/tree/master/.agents/notes/implemented/simplification)
- [SessionEvent 类型所在目录](https://github.com/deepseek-ai/deepseek-harness/tree/master/packages/core/session/src)
