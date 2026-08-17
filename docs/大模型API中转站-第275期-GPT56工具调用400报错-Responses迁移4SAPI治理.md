---
title: "Responses 工具调用 400 错误如何定位"
tags:
  - Responses API
  - 工具调用
  - 错误排查
description: "工具调用返回 400 时，单看错误文本往往无法判断是请求格式、工具声明还是响应接口不匹配。"
---
# Responses 工具调用 400 错误如何定位
工具调用返回 400 时，单看错误文本往往无法判断是请求格式、工具声明还是响应接口不匹配。本文把一次失败请求拆成可比对的输入、工具定义、模型响应和重试条件，帮助你先固定证据，再选择修复或迁移路径。文中只讨论可在本地复现的步骤，不把单次结果扩展成产品承诺；每个结论都标注前提、证据和无法覆盖的边界。读者可以先完成最小验证，再根据自己的版本、权限和数据补充实验，避免把配置示例误当成普遍结论。

最近在模型接入日志中，出现过这样一条 400：

```text
Function tools with reasoning_effort are not supported for gpt-5.6-sol
in /v1/chat/completions.
Please use /v1/responses instead.
```

很多人的第一反应是：函数定义是不是写错了？模型名是不是拼错了？要不要重试？

这几个方向都不是首要答案。

这条错误实际上已经给出了限制条件：当前请求使用了函数工具，同时设置了推理强度，并且发送到了 `/v1/chat/completions`。对这个模型能力组合来说，旧接口不接受该请求。


## 1. 先还原真实请求

错误信息只展示了一部分上下文。要定位问题，先从脱敏日志中还原以下字段：

```text
model
endpoint
reasoning_effort
tools 是否存在
tools 的数量和类型
HTTP 状态码
上游错误原文
request_id
```

典型的触发请求大致如下：

```json
{
  "model": "gpt-5.6-sol",
  "messages": [
    { "role": "user", "content": "检查当前项目的测试失败原因" }
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "read_file",
        "description": "Read an authorized local file",
        "parameters": {
          "type": "object",
          "properties": { "path": { "type": "string" } },
          "required": ["path"]
        }
      }
    }
  ],
  "reasoning_effort": "medium"
}
```

真正导致 400 的不是某个业务问题，而是这三个字段同时出现：

```text
function tools
+ reasoning_effort != none
+ /v1/chat/completions
```

## 2. 为什么 Agent 更容易遇到这个错误

普通问答可能没有 `tools`，因此不会触发限制。Coding Agent、OpenCode、知识库 Agent 和自动化工作流则不同，它们通常会注册：

```text
read_file
search_code
run_command
write_file
run_tests
```

从 API 角度看，这些都是函数工具。即使用户只输入一句自然语言，客户端也可能自动把工具定义附加到请求中。

另一方面，复杂任务常常会打开 `low`、`medium`、`high` 或其他 reasoning 档位。于是“工具调用”和“推理设置”同时出现，旧的 Chat Completions 接口就成为冲突点。

这也是为什么有时删掉 `tools` 后请求能成功：不是模型突然变好了，而是触发条件少了一项。

## 3. Chat Completions 与 Responses 的差异

OpenAI 的迁移文档建议新项目使用 Responses API。两者的关键差异如下：

| 项目 | Chat Completions | Responses |
| --- | --- | --- |
| Endpoint | `/v1/chat/completions` | `/v1/responses` |
| 请求上下文 | `messages` | `input` |
| 文本读取 | `choices[0].message.content` | `output_text` 或 `output` |
| 函数定义 | `function` 外部包裹 | 内部标记的 function Item |
| 工具结果 | tool message | `function_call_output` Item |
| 调用关联 | 依赖消息顺序 | `call_id` |
| 结构化输出 | `response_format` | `text.format` |
| 多轮状态 | 应用自行维护 messages | `previous_response_id` 或手动 replay |

因此，不能只把 URL 从 `/v1/chat/completions` 改成 `/v1/responses`，还要同步修改请求体和响应解析器。

## 4. 推荐路线：迁移到 Responses API

### 4.1 先确认通道能力


```text
目标模型是否可用。
当前通道是否开放 /v1/responses。
该模型是否支持 function tools。
当前 API Key 是否有对应权限。
```

不同通道、账户和模型别名的能力可能不同。不要因为某个模型在官方文档中支持某功能，就默认你的中转通道已经开放。

### 4.2 Python 请求示例


```python
import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["API_KEY"],
    base_url=os.environ.get("API_BASE_URL", "https://api.example.com/v1"),
)

response = client.responses.create(
    model="gpt-5.6-sol",
    input="检查当前项目的测试失败原因",
    reasoning={"effort": "medium"},
    tools=[
        {
            "type": "function",
            "name": "read_file",
            "description": "Read an authorized local file",
            "parameters": {
                "type": "object",
                "properties": {"path": {"type": "string"}},
                "required": ["path"],
                "additionalProperties": False,
            },
        }
    ],
)

print(response.output_text)
```

这里有三个和旧接口不同的地方：

1. 上下文使用 `input`；
2. 函数定义不再嵌套在 `function` 字段下；
3. reasoning 通过 Responses 的参数结构表达，而不是机械复制旧接口字段。

实际 SDK 或网关的字段支持，以对应版本文档为准。不要因为客户端能接受一个字段，就认为上游模型一定能处理它。

### 4.3 工具结果必须带 call_id

模型返回 `function_call` 后，应用执行工具，再把结果回传：

```python
tool_call = next(
    item for item in response.output
    if item.type == "function_call"
)

tool_result = read_authorized_file(tool_call.arguments)

next_response = client.responses.create(
    model="gpt-5.6-sol",
    previous_response_id=response.id,
    input=[
        {
            "type": "function_call_output",
            "call_id": tool_call.call_id,
            "output": tool_result,
        }
    ],
)
```

缺少 `call_id` 时，模型调用和工具结果无法可靠关联。企业审计也会失去“哪个请求调用了哪个工具”的链路。

## 5. 兼容路线：继续 Chat Completions，但关闭推理

如果当前网关或客户端暂时不能使用 Responses，可以采用错误信息给出的兼容路线：

```json
{
  "model": "gpt-5.6-sol",
  "messages": [
    { "role": "user", "content": "检查当前项目的测试失败原因" }
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "read_file",
        "parameters": {
          "type": "object",
          "properties": { "path": { "type": "string" } },
          "required": ["path"]
        }
      }
    }
  ],
  "reasoning_effort": "none"
}
```

注意三个边界：

- `reasoning_effort` 必须是服务端允许的 `none`，不要继续发送 `medium` 或 `high`；
- 关闭 reasoning 不代表所有模型和通道都必然支持工具，仍需查看能力矩阵；
- 客户端仍然使用 Chat Completions 的 `messages` 和 `choices` 解析。

这条路线适合临时兼容和低复杂度任务，不适合作为复杂 Coding Agent 的长期架构。因为你保留了旧协议，也放弃了该模型在 Responses 下的推理和工具协同能力。

## 6. 不推荐的“修复”：删除所有工具

把 `tools` 整段删除，确实可能让纯文本请求通过，但 Agent 会失去：

```text
读取文件。
搜索代码。
运行测试。
修改项目。
验证修复。
```

如果应用本来只是文本问答，这样做可以；如果应用是 OpenCode、Coding Agent 或自动化工作流，删除工具等于改变产品能力，不是修复协议。



### 7.1 客户端层

确认客户端实际发出的：

```text
Base URL
Endpoint
model
tools 数量
reasoning_effort 或 reasoning
SDK 版本
```

不要只看界面上的模型名称。很多 Agent 会在运行时注入工具和推理配置。

### 7.2 网关层


```text
模型名
请求路径
工具定义
推理参数
重试次数
```

如果客户端配置了 `none`，但日志中仍然出现 `medium`，说明 provider、模型预设或网关策略重新注入了参数。

### 7.3 上游能力层

建立一张能力矩阵，不要把“模型可见”当成“能力可用”：

| 模型/通道 | Chat Completions | Responses | Function tools | Reasoning |
| --- | --- | --- | --- | --- |
| 目标通道 A | 以后台为准 | 以后台为准 | 以后台为准 | 以后台为准 |
| 目标通道 B | 以后台为准 | 以后台为准 | 以后台为准 | 以后台为准 |

测试时分别验证纯文本、工具调用、工具加 reasoning 三种请求，不要只测第一种。

### 7.4 业务层

最后确认应用是否能正确处理响应：

```text
Responses 是否读取 output/output_text。
function_call 是否执行授权工具。
function_call_output 是否带 call_id。
流式事件是否使用 Responses 类型。
失败后是否会无限重试。
```

## 8. 企业级 Key、预算和回退策略

生产环境不要让所有应用共享一个 Key。建议至少按以下维度拆分：

```text
团队：研发、客服、数据、运营。
环境：开发、预发布、生产。
项目：Coding Agent、知识库、SaaS API。
模型：普通模型、reasoning 模型、高成本模型。
```


- 模型白名单；
- 每分钟请求数；
- 每日和每月预算；
- 单任务最大工具步骤数；
- 超时和有限重试；
- 失败告警和备用模型。

回退规则不要简单地“所有错误都重试”：

| 错误 | 建议动作 |
| --- | --- |
| 400 参数/能力不兼容 | 修正协议，不重试原请求 |
| 401/403 Key 或权限 | 停止重试，检查分组和模型权限 |
| 429 限流 | 按策略退避，必要时切备用通道 |
| 5xx/超时 | 有限重试，记录 request_id |
| 工具执行失败 | 修正工具参数或交人工，不无限循环 |

## 9. 上线前验证清单

```text
[ ] 纯文本请求可以成功。
[ ] Chat Completions + reasoning_effort=none 已验证，若保留兼容路径。
[ ] Responses + function tools 已验证，若采用迁移路径。
[ ] 工具定义格式与所选 endpoint 一致。
[ ] function_call_output 带正确 call_id。
[ ] 上游 API 后台模型、endpoint 和 Key 权限已核对。
[ ] 日志包含 request_id、model、endpoint、错误码和耗时。
[ ] Key、用户内容和工具参数已脱敏。
[ ] 400 不会自动无限重试。
[ ] Agent 有最大步骤数和预算上限。
[ ] 已在隔离环境完成回归测试并保留回滚配置。
```

## 10. 成本与风险提示

迁移到 Responses 并不等于自动省钱。工具调用会增加模型步骤和上下文，reasoning 档位也可能增加延迟与输出 Token。应按“成功任务成本”评估：

```text
成功任务成本
= 模型输入费用
+ 模型输出费用
+ 工具调用费用
+ 重试费用
+ 人工返工成本
```


## 11. 总结

这条 400 的根因可以压缩成一句话：

```text
函数工具 + 非 none 推理设置 + Chat Completions
```

优先解决方案是迁移到 `/v1/responses`，同步调整 `input/output`、函数定义、工具结果和 `call_id`。如果当前链路暂时不能迁移，再把 Chat Completions 的 `reasoning_effort` 设为 `none`，并接受能力上的限制。




## 官方来源

- OpenAI：Migrate to the Responses API：https://developers.openai.com/api/docs/guides/migrate-to-responses
- OpenAI：Function calling：https://developers.openai.com/api/docs/guides/function-calling

## 结论

本文给出了问题定位、配置或验证的可执行路径。实际结果仍取决于当前版本、权限和运行环境，提交前应按官方文档复核可变字段，并保留失败证据和回滚边界。
