---
title: "如何测试 Responses API 中转站的工具调用兼容性"
category: 人工智能
tags:
  - OpenAI API
  - Function Calling
  - 接口测试
description: "用一组可复现的 Responses API 请求，验证中转站的文本响应、Function Calling、工具回调、Custom Tool 和推理参数是否真正兼容。"
---

# 如何测试 Responses API 中转站的工具调用兼容性

很多中转站都能返回一句“你好”，但这只能证明最基础的文本请求可以走通。真正接入 Coding Agent 或自动化工作流后，最容易出问题的往往是工具调用、第二轮回传、Reasoning 输出和流式事件。

例如，第一轮请求能返回 `function_call`，第二轮却不接受 `function_call_output`；或者普通文本没问题，Custom Tool 和 Grammar 一启用就报参数错误。只测一个聊天请求，很难发现这些兼容性缺口。

下面给出一套面向 OpenAI-compatible Responses API 中转站的测试方案。测试的目标不是证明某个服务“好不好”，而是把兼容性拆成可以逐项验证的断言：请求是否被接受、响应结构是否正确、工具结果是否真的进入下一轮上下文，以及中转层是否改写了显式参数。

文中的接口格式依据 OpenAI Function Calling 文档和 Responses API 参考；实际模型名称、可用字段和服务商限制，应以被测接口当前文档为准。本文不要求使用某一个固定服务商，示例测试中可以把 4SAPI 作为一个可替换的 OpenAI-compatible endpoint，但它不代表唯一选项。

## 先定义测试边界

一次接口测试最好只回答一个问题。把所有请求揉成一个“大而全”的 Demo，失败后很难判断是网络、模型、工具格式还是中转层改写造成的。

建议把测试拆成三类结论：

| 结论类别 | 要回答的问题 | 典型证据 |
| --- | --- | --- |
| 能力兼容性 | 接口是否支持某项 API 能力 | HTTP 状态、响应类型、字段结构和本地断言 |
| 上下文连续性 | 工具结果是否能进入第二轮模型上下文 | 随机 receipt 是否被原样回显 |
| 参数透明度 | 中转层是否按请求传递字段 | 请求 JSON 与响应、事件和日志的差异 |

这三类结论不能互相替代。Function Calling 通过，不代表参数一定原样透传；模型返回了正确答案，也不代表它一定来自请求中声明的模型。

## 测试环境和请求方式

建议使用 Python 直接发送 HTTP 请求，不要先经过某个 Agent SDK、桌面客户端或二次封装。包装层可能自动注入 `instructions`、拼接历史消息、修改工具格式，导致你测到的是客户端行为，而不是中转站行为。

最小项目可以只有一个脚本和一个报告目录：

```text
relay_probe.py
.env.example
reports/
```

依赖只需要一个 HTTP 客户端，例如：

```text
httpx
```

通过环境变量传递地址和密钥，不要把真实 Key 写进脚本、截图或公开仓库：

```text
OPENAI_BASE_URL=https://your-relay.example.com/v1
OPENAI_API_KEY=replace-with-your-key
PROBE_MODEL=gpt-5.6-sol
PROBE_TIMEOUT_SECONDS=180
```

程序启动时先验证以下条件：

```text
OPENAI_BASE_URL 非空，并且路径只包含一个 /v1。
OPENAI_API_KEY 非空，但日志中不能打印完整内容。
PROBE_MODEL 是实际存在的模型 ID，而不是随意拼出的别名。
请求地址最终应为 {OPENAI_BASE_URL}/responses。
```

每次请求都保存脱敏后的 request JSON、response JSON、HTTP 状态和客户端测得的延迟。认证头只能写成：

```text
Authorization: Bearer ***REDACTED***
```

## 第一项：基础文本响应

先发送不含工具的最小请求。提示词可以要求模型只输出一个固定字符串：

```json
{
  "model": "gpt-5.6-sol",
  "input": "只输出 RELAY_BASELINE_OK，不要添加标点或解释。"
}
```

检查项包括：

```text
HTTP 状态码为 200。
响应对象类型为 response。
status 为 completed，且 error 为空。
输出中存在 message 类型的 item。
最终文本严格等于 RELAY_BASELINE_OK。
```

如果这一项失败，不要马上增加工具参数。先确认 URL、认证、模型名、超时和服务商是否真的支持 Responses API。许多所谓“OpenAI 兼容”接口只实现了 Chat Completions，路径可以返回数据，却不保证 Responses API 的事件和对象结构。

## 第二项：Function Calling 第一轮

下一步验证模型是否能按照工具 Schema 返回标准的函数调用。定义一个参数非常少的工具，避免把 Schema 本身变成变量：

```json
{
  "type": "function",
  "name": "report_probe_result",
  "description": "报告固定的探测结果",
  "strict": true,
  "parameters": {
    "type": "object",
    "properties": {
      "check": {"type": "string", "enum": ["gpt56-relay"]},
      "status": {"type": "string", "enum": ["ok"]},
      "result": {"type": "integer", "enum": [42]}
    },
    "required": ["check", "status", "result"],
    "additionalProperties": false
  }
}
```

请求中可以明确要求必须调用工具，并关闭并行工具调用：

```json
{
  "model": "gpt-5.6-sol",
  "input": "调用 report_probe_result 报告探测完成。",
  "tools": ["上面的工具定义"],
  "tool_choice": "required",
  "parallel_tool_calls": false
}
```

本地断言应检查：

```text
输出中存在 type=function_call 的 item。
name 为 report_probe_result。
call_id 非空，后续回调要使用同一个值。
arguments 是合法 JSON，而不是一段无法解析的字符串。
参数严格等于 check=gpt56-relay、status=ok、result=42。
```

这里不要只看模型有没有“说它调用了工具”。真正的工具调用必须出现在结构化输出 item 中。把自然语言中的“我已经调用了”当成成功，会把模型没有按协议调用工具的情况误判为通过。

## 第三项：Function Calling 第二轮

第一轮拿到函数调用后，测试程序模拟本地工具执行，并把结果作为 `function_call_output` 发回模型。为了证明结果确实进入了第二轮上下文，工具返回值不应写死，建议每次随机生成一个 receipt：

```python
import secrets

receipt = secrets.token_hex(12)
```

模拟结果示例：

```json
{
  "accepted": true,
  "server_status": "healthy",
  "reason": "probe_ok",
  "receipt": "每次运行随机生成"
}
```

第二轮输入至少包含三部分：原始用户输入、第一轮返回的完整 output，以及与第一轮 `call_id` 对应的 `function_call_output`：

```json
{
  "type": "function_call_output",
  "call_id": "第一轮返回的 call_id",
  "output": "工具结果的 JSON 字符串"
}
```

对于 Reasoning 模型，第一轮返回的 reasoning item 也可能是下一轮所需的上下文，不能只复制 `function_call` 而丢掉其他 output。第二轮的断言包括：

```text
请求成功完成。
模型没有无意义地再次调用同一个工具。
最终文本包含 accepted、server_status、reason 和 receipt。
最终文本中的 receipt 与本地随机值完全一致。
```

如果 receipt 不一致，不能直接归因于模型“记性差”。需要先检查：第二轮是否真的发送了工具结果、`call_id` 是否匹配、第一轮 output 是否被客户端截断、服务端是否丢弃了未知 item，以及中转层是否重写了请求。

## 第四项：`store: false` 和服务端续传要分开测

Responses API 可能支持两种上下文续传方式：客户端自己回传上一轮 output，或者通过 `previous_response_id` 让服务端关联历史状态。

测试 `store: false` 时，重点是确认客户端完整保存并回传第一轮结果；测试 `previous_response_id` 时，重点是确认服务端状态链路可用。二者不能互相代替。

如果手动续传通过、`previous_response_id` 失败，只能说明两条状态链路的兼容性不同；不能写成“工具调用完全不支持”。报告中应分开列出：

```text
MANUAL_CONTEXT: PASS | WARN | FAIL
SERVER_CONTEXT: PASS | WARN | FAIL | SKIP
```

这会让使用者知道是否需要在客户端保存完整历史，而不是只看到一个笼统的“上下文通过”。

## 第五项：Custom Tool 和 Grammar

如果目标模型或接口支持 Custom Tool，可以增加一个输入受约束的数学表达式工具。比如要求工具输入严格为：

```text
17 * 25 + 3
```

断言包括：

```text
输出类型为 custom_tool_call。
工具名为 math_exp。
call_id 非空。
input 严格等于 17 * 25 + 3。
输入中没有 Grammar 之外的解释文字、代码块或其他符号。
```

这项测试主要验证 Custom Tool 和约束生成，不一定需要继续执行第二轮回调。若接口返回普通文本而不是 `custom_tool_call`，应记录为该能力不兼容，而不是把它与普通 Function Calling 失败混在一起。

## 第六项：Reasoning A/B 测试

对同一个确定性题目，分别发送 `reasoning.effort` 为 `none` 和 `max` 的请求。测试题必须有本地可验证答案，例如中国剩余定理结果：

```json
{"n":177636,"r17":3,"r19":5,"r23":7,"r29":11}
```

本地校验：

```text
n % 17 == 3
n % 19 == 5
n % 23 == 7
n % 29 == 11
```

每种配置至少运行 3 次，记录正确率、延迟、输入 Token、输出 Token、reasoning tokens 以及响应中实际出现的 effort/context 字段。

不要要求 `max` 每次都比 `none` 产生更多 reasoning tokens，也不要把单次延迟差异写成性能提升。Reasoning 资源分配存在波动，应该比较多次运行的正确率、中位延迟和中位 reasoning tokens。

## 如何判定结果

每个测试使用统一格式：

```json
{
  "name": "function_callback",
  "model": "gpt-5.6-sol",
  "status": "PASS",
  "http_status": 200,
  "latency_ms": 1832,
  "errors": [],
  "warnings": [],
  "metrics": {
    "input_tokens": 285,
    "output_tokens": 30,
    "reasoning_tokens": 0
  }
}
```

状态含义：

| 状态 | 含义 |
| --- | --- |
| PASS | 所有必要断言通过 |
| WARN | 核心能力通过，但有字段无法验证或存在非关键差异 |
| FAIL | HTTP、结构、回调、Grammar、答案或关键参数断言失败 |
| SKIP | 当前模型或接口明确不支持可选测试 |

最终不要只输出一个总分，建议拆成：

```text
CAPABILITY: PASS | WARN | FAIL
CONTEXT: PASS | WARN | FAIL | SKIP
TRANSPARENCY: PASS | WARN | FAIL
IDENTITY: UNVERIFIED
```

即使 Function Calling、Custom Tool 和 Reasoning 全部通过，也不能仅凭响应里的 `model`、`usage` 或 reasoning tokens 从密码学意义上证明上游模型身份。中转层可以改写这些字段，模型本身也可能是兼容实现。

## 常见误区

**用 SDK 测中转站。** SDK 可能添加系统指令、改写历史和规范化工具，先用原始 HTTP 请求建立基线。

**只测第一轮工具调用。** 第一轮通过不代表工具输出能被第二轮正确接收，必须使用随机 receipt 做闭环。

**只看 HTTP 200。** 200 只能说明请求被接受，不能说明输出 item、call_id、arguments 和回调结构正确。

**自动重试所有错误。** `429`、`5xx`、超时和非法 JSON 本身就是健康检查结果。需要重试时最多一次，并保留第一次失败。

**把模型身份写成已证明。** 能力兼容、参数透明和身份验证是三个不同问题，报告必须分开写。

## 第七项：流式响应必须按事件序列验证

很多中转站在非流式请求下能给出完整 JSON，却会在 SSE 流式响应里遗漏终止事件、拼错工具参数增量，或者在工具调用前后改变事件顺序。若业务中的 Agent 默认使用流式，就不能拿一次 `stream: false` 的成功替代流式兼容性测试。

流式 case 仍应沿用基础文本、Function Calling 和第二轮回调三类任务，但保存对象从“最终响应 JSON”改为“完整事件序列”。检查重点包括：

```text
是否收到创建响应的起始事件。
文本增量能否按顺序拼成预期最终文本。
function_call 的 name、call_id 与 arguments 增量能否重组成合法完整对象。
错误事件是否带有可读的错误类型和请求上下文。
完成事件是否出现，最终状态是否与非流式语义一致。
```

不要只在客户端拼好文本后断言最终答案。应把每个原始 SSE data 行或解析后的事件对象写入独立文件，再保存最终重组结果。出现失败时，这能区分是服务端没有发送某个增量、代理截断了连接，还是客户端自己的事件合并逻辑有 bug。

工具调用尤其要在流式与非流式下分别断言。官方 Function Calling 指南描述的是接收结构化工具调用、在本地执行、回传工具输出并获取最终结果的完整循环；流式只改变传输形态，不应改变 `call_id`、工具名和最终参数语义。[Function Calling 指南](https://developers.openai.com/api/docs/guides/function-calling)可作为协议基线。若非流式通过、流式只在多分片参数上失败，报告应精确写出这一差异，而不是笼统说“工具调用不兼容”。

## 使用一个最小测试运行器，而不是手工拼请求

当 case 增加到多个模型、多个 endpoint 和多轮工具回调后，手工复制 JSON 很容易把上一次的 response ID、tool 参数或环境变量带进下一次。建议把每个 case 定义成数据，而不是散落在聊天记录中的命令。一个最小清单可以包含：

```text
case 名称：baseline_text、function_call、function_callback、stream_function_call
适用模型：显式模型 ID 列表
请求体模板：不含密钥的原始 JSON
本地断言：状态、字段、receipt、数学结果或事件序列
重试规则：是否可重试、最多次数、哪些错误不重试
证据输出：request、response、event、meta 和 summary 路径
```

运行器应为每个请求生成唯一 case ID，并在 `meta.json` 中记录时间、客户端版本、目标主机、HTTP 状态、延迟和响应 ID。任何认证头只保存为 `Bearer ***REDACTED***`。若输入中含真实业务数据，则应先替换为合成样本；接口兼容性不需要把客户内容送进测试。

同一个 case 失败时，先保留第一次失败再讨论重试。网络类错误可以按照预先写好的规则有限重试，但 `400`、Schema 不合法、回调结构不匹配、receipt 不一致这类协议失败不应自动用第二次成功覆盖。它们很可能正是本次测试想发现的回归。

## 将兼容性矩阵按“能力”而不是“服务名”组织

团队往往需要比较多个模型部署或同一服务的多个路由。比起为每个 endpoint 写一段主观总结，更适合维护一张随版本更新的能力矩阵：

| 能力 | 基线 | Function 第一轮 | 第二轮回调 | 手动续传 | 服务端续传 | Custom Tool | Grammar | 流式 | 参数审计 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 模型/路由 A |  |  |  |  |  |  |  |  |  |
| 模型/路由 B |  |  |  |  |  |  |  |  |  |

矩阵中的单元格不要只填对错。必要时加上 `PASS`、`WARN`、`FAIL`、`SKIP`、测试日期、报告目录与失败 case ID。`SKIP` 应只用于接口明确不声明或当前任务不适用的可选能力，不能用来掩盖无法解释的失败。矩阵旁要保留配置版本和请求模板哈希，否则后续扩展工具定义或升级 SDK 后，历史结果无法横向比较。

通过这种方式，接口升级后只需重跑同一组小 case，就能快速看出某一项能力是否回归。它比“重新发一句你好”更接近真正的集成验收，也不会把模型质量、协议支持和中转透明度混成一条未经验证的结论。

## 结论

一个可靠的中转站兼容性测试，不是发送一句“你好”，而是验证完整协议链路：基础响应、结构化工具调用、工具结果回传、上下文续传、受约束工具输入和推理参数。

建议先用小测试集建立基线，再逐步增加流式、并行工具、Structured Outputs 和缓存字段。每个能力单独评分，保留原始请求和响应，才能在接口升级或更换服务商后快速发现兼容性回归。

官方参考：[Function Calling](https://developers.openai.com/api/docs/guides/function-calling)、[Responses API](https://platform.openai.com/docs/api-reference/responses/create)。
