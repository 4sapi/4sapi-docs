---
title: "DeepSeek-V4-Flash 思考模式与工具调用实战"
tags:
  - DeepSeek API
  - 工具调用
  - Agent
description: "解释 V4-Flash 的思考开关、推理强度、reasoning_content 回传要求、工具调用循环和并发隔离方法。"
---
# DeepSeek-V4-Flash 思考模式与工具调用实战

很多 Agent 代码在普通问答中运行正常，一接入工具就开始报 400，或者第一轮能调用函数，第二轮却丢失上下文。问题通常不在工具函数本身，而在思考模式的消息拼接规则没有被正确实现。

DeepSeek-V4-Flash 的思考模式默认开启，官方文档同时提供了思考开关和推理强度控制。工具调用场景下，模型的 `reasoning_content` 不是可有可无的调试文本：如果请求带有工具，后续轮次需要完整回传相关字段，否则 API 可能返回 400。本文从一次普通思考请求讲到多轮工具循环，并补充并发和用户隔离的实现要点。

## 一、思考模式的三个控制层

DeepSeek 官方文档把思考模式拆成“开关”和“强度”两层。以 OpenAI Chat Completions 格式为例：

```text
思考开关：extra_body = {"thinking": {"type": "enabled" 或 "disabled"}}
推理强度：reasoning_effort = "low"、"high" 或 "max"
```

Responses API 使用不同字段控制强度，官方兼容性表列出 `reasoning.effort`。Codex 的配置则使用 `model_reasoning_effort`。这几个字段处在不同协议层，不要把一个协议里的字段原样复制到另一个协议。

官方说明当前思考模式默认打开，默认 effort 为 `high`。如果业务需要统一行为，应在应用配置中显式记录模式，而不是依赖默认值。默认值可能随模型或接入文档变化。

## 二、思考模式下哪些采样参数不生效

官方思考模式文档说明，思考模式不支持 `temperature`、`top_p`、`presence_penalty` 和 `frequency_penalty`。为了兼容已有软件，部分参数即使被传入也可能不报错，但不代表它们实际改变了输出。

因此在代码审查时，不要只检查请求是否成功，还要检查参数行为：

```text
思考关闭时：验证 temperature、top_p 是否按预期工作
思考开启时：不要把 temperature 当作稳定性控制手段
更换协议时：核对参数名称和嵌套位置
```

如果你需要控制思考深度，优先使用文档支持的 effort 字段；如果需要控制非思考模式的随机性，再使用对应协议支持的采样参数。

## 三、一次普通思考请求

下面是 OpenAI Python SDK 的 Chat Completions 示例。代码使用 `extra_body` 传入 DeepSeek 特有的思考开关：

```python
import os
from openai import OpenAI


client = OpenAI(
    api_key=os.environ["DEEPSEEK_API_KEY"],
    base_url="https://api.deepseek.com",
)

response = client.chat.completions.create(
    model="deepseek-v4-flash",
    messages=[
        {"role": "user", "content": "比较 9.11 和 9.8 哪个更大，并说明判断过程。"}
    ],
    reasoning_effort="high",
    extra_body={"thinking": {"type": "enabled"}},
)

message = response.choices[0].message
print("reasoning:", message.reasoning_content)
print("answer:", message.content)
```

运行前需要安装 `openai` 并设置 `DEEPSEEK_API_KEY`。验证时不要把思维内容直接展示给最终用户作为产品功能，应用通常只需要保存最终回答和必要的调试元数据；具体日志保留策略还要结合隐私和安全要求。

## 四、为什么工具调用必须保留 reasoning_content

思考模式下，一轮用户请求可能被拆成多个子请求：模型先思考并调用日期工具，再根据日期调用天气工具，最后生成答案。每次工具返回后，应用都要把上一轮 assistant 消息和工具结果放回 `messages`，让模型继续当前 Turn 的思考。

官方文档的关键要求是：携带 `tools` 参数的请求，在后续所有请求中必须完整回传 `reasoning_content`。如果只保留 `content`、`tool_calls`，或者自行重建消息时漏掉该字段，API 可能返回 HTTP 400。

最稳妥的拼接方式是直接追加 SDK 返回的 assistant 消息：

```python
messages.append(response.choices[0].message)
```

这相当于保留以下字段：

```python
messages.append({
    "role": "assistant",
    "content": response.choices[0].message.content,
    "reasoning_content": response.choices[0].message.reasoning_content,
    "tool_calls": response.choices[0].message.tool_calls,
})
```

如果 SDK 对象不能直接序列化，要使用 SDK 文档提供的模型转字典方法，并检查序列化结果中没有丢失 `reasoning_content`、`tool_calls` 和工具调用 ID。

## 五、一个可控的工具调用循环

下面的示例使用两个本地模拟函数，重点展示消息顺序和停止条件。真实应用中，工具函数应拥有独立的权限控制、超时和输入校验。

```python
import json
from datetime import date


tools = [
    {
        "type": "function",
        "function": {
            "name": "get_date",
            "description": "获取当前日期",
            "parameters": {"type": "object", "properties": {}},
        },
    },
    {
        "type": "function",
        "function": {
            "name": "get_weather",
            "description": "获取指定地点和日期的天气",
            "parameters": {
                "type": "object",
                "properties": {
                    "location": {"type": "string"},
                    "date": {"type": "string"},
                },
                "required": ["location", "date"],
            },
        },
    },
]


def get_date():
    return date.today().isoformat()


def get_weather(location, target_date):
    return {"location": location, "date": target_date, "weather": "cloudy"}


TOOL_MAP = {
    "get_date": lambda arguments: get_date(),
    "get_weather": lambda arguments: get_weather(
        arguments["location"], arguments["date"]
    ),
}
```

循环部分：

```python
messages = [
    {"role": "user", "content": "明天杭州天气怎么样？"}
]

for step in range(1, 6):
    response = client.chat.completions.create(
        model="deepseek-v4-flash",
        messages=messages,
        tools=tools,
        reasoning_effort="high",
        extra_body={"thinking": {"type": "enabled"}},
    )

    assistant_message = response.choices[0].message
    messages.append(assistant_message)

    tool_calls = assistant_message.tool_calls or []
    if not tool_calls:
        print(assistant_message.content)
        break

    for tool_call in tool_calls:
        name = tool_call.function.name
        arguments = json.loads(tool_call.function.arguments)
        if name not in TOOL_MAP:
            raise RuntimeError(f"unsupported tool: {name}")

        result = TOOL_MAP[name](arguments)
        messages.append({
            "role": "tool",
            "tool_call_id": tool_call.id,
            "content": json.dumps(result, ensure_ascii=False),
        })
else:
    raise RuntimeError("tool loop exceeded maximum steps")
```

这个循环有四个保护点：限制最大子步骤，拒绝未知工具，解析并校验工具参数，工具执行后使用正确的 `tool_call_id` 回传结果。真实写操作还需要增加人工确认或幂等键，不能因为模型连续调用两次就执行两次不可逆操作。

## 六、两轮对话时如何处理 reasoning_content

如果两个 user 消息之间没有工具调用，上一轮的 `reasoning_content` 在下一轮可以不参与上下文拼接；如果中间发生工具调用，则必须保留并在后续请求中完整回传。最简单的工程规则是：不要自己裁剪 assistant 消息，直接保存 SDK 返回对象，并在请求前做 schema 检查。

可以在发送请求前增加调试断言：

```python
def assert_tool_history(messages):
    for message in messages:
        if getattr(message, "tool_calls", None):
            if not getattr(message, "reasoning_content", None):
                raise ValueError("tool history is missing reasoning_content")
```

这不是业务逻辑的替代品，只是帮助尽早发现消息拼接错误。生产环境中还应记录被拒绝的消息字段和请求 ID，但不要把 API Key 或完整敏感上下文写进日志。

## 七、Responses API 与 Chat Completions 的选择

如果你的框架已经使用 Chat Completions，工具调用逻辑可以沿用消息数组和 `tool_calls` 机制，但要按 DeepSeek 思考模式要求保留 `reasoning_content`。

如果你的框架基于 Codex 或 OpenAI Responses API，当前官方文档显示 V4-Flash 支持 Responses API。Responses API 使用事件流，并且是无状态的；它的工具类型和输入内容也有单独的兼容性边界。不要把 Chat Completions 的消息拼接代码和 Responses 的事件处理代码混用。

选择标准可以简单化为：

```text
已有 Chat Completions 服务、改动要小 -> 继续使用 Chat Completions
需要 Codex 或 Responses 事件流 -> 使用 Responses API
需要图片、文件或其他内置工具 -> 先核对当前接口支持表
```

参考：[思考模式文档](https://api-docs.deepseek.com/zh-cn/guides/thinking_mode) 与 [Responses API 兼容性文档](https://api-docs.deepseek.com/zh-cn/guides/responses_api)。

## 八、并发和 user_id 隔离

截至 2026-07-31 官方限速文档，V4-Flash 的账号级并发限制为 2500，V4-Pro 为 500。超过账号并发限制时会收到 HTTP 429；并发按账号计算，与 API Key 数量不是简单的一对一关系。

业务侧可以传递 `user_id`，用于内容安全、KVCache 和调度隔离。官方要求它满足 `[a-zA-Z0-9\\-_]+`，最大长度 512，并提醒不要在其中包含用户隐私信息。对普通 API 用户，所有 user_id 仍会合并计算并发；不要把设置 user_id 理解成自动获得更多总配额。

一个稳妥的设计是使用内部不可逆用户标识：

```text
user_id = tenant_id + "-" + hash(internal_user_id)
```

同时在应用层设置自己的并发信号量、超时和 429 退避。服务端账号限制是上限，不是你应该立即打满的目标。

## 九、工具调用的验证清单

上线前至少覆盖以下案例：

```text
[ ] 无工具的普通思考请求
[ ] 一次工具调用后返回最终答案
[ ] 连续两次以上工具调用
[ ] 模型返回未知工具名
[ ] 工具参数不是合法 JSON
[ ] 工具执行超时或返回错误
[ ] assistant 消息缺少 reasoning_content
[ ] 达到最大工具循环次数
[ ] 触发 429 后不会重复执行副作用操作
[ ] 流式和非流式的结束状态都能正确记录
```

测试重点不是让模型“看起来会调用工具”，而是验证应用能否在错误、重试和中断时保持状态一致。尤其是支付、删文件、发消息和部署等副作用操作，必须由代码层做权限、参数和幂等控制。

## 结语

DeepSeek-V4-Flash 的思考模式与工具调用可以组合使用，但消息历史必须完整、循环必须有上限、工具执行必须有权限边界。`reasoning_content` 在普通问答中可能只是输出字段，在工具调用链路中却是后续请求能够继续工作的上下文组成部分。

先用本地模拟工具跑通消息拼接，再接入真实服务；先验证错误路径和副作用控制，再扩大并发。模型负责推理和选择工具，应用负责状态、权限、校验和停止条件，这个分工不能省略。

