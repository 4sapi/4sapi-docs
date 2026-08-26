---
title: "DeepSeek V4-Flash Responses 接入教程"
tags:
  - DeepSeek API
  - Responses API
  - Python
description: "从 OpenAI SDK 安装、基础请求、流式事件到兼容性验证，给出调用 DeepSeek-V4-Flash Responses API 的可复现流程。"
---
# DeepSeek V4-Flash Responses 接入教程

很多 OpenAI 兼容接口的教程只展示一段请求代码，却没有说明响应为什么为空、流式为什么收不了口、客户端为什么带了参数但模型没有按预期工作。DeepSeek-V4-Flash 的 Responses API 接入尤其需要注意这一点：它提供了与 Responses 格式兼容的入口，但不是 OpenAI Responses API 的所有功能都已经支持。

本文以 Python OpenAI SDK 为例，完成一次最小调用、一次流式调用，并给出兼容性检查和排错方法。示例不包含真实 API Key，所有命令默认在项目虚拟环境中执行。当前官方文档说明，Responses API 只支持 `deepseek-v4-flash`，不支持 `deepseek-v4-pro`。

## 一、前置条件和版本范围

开始前准备：

- Python 3.9 或更高版本；
- 一个可以调用 DeepSeek API 的 API Key；
- 网络能够访问 `https://api.deepseek.com`；
- OpenAI Python SDK 已安装；
- 代码运行目录与输出日志目录已经确定。

官方文档中的模型入口是：

```text
base_url: https://api.deepseek.com
model: deepseek-v4-flash
```

模型版本和价格会变化。本文基于 [DeepSeek Responses API 文档](https://api-docs.deepseek.com/zh-cn/guides/responses_api) 与 [模型和价格文档](https://api-docs.deepseek.com/zh-cn/quick_start/pricing) 在 2026-07-31 的公开内容撰写，运行前仍应核对当前页面。

如果团队通过 4SAPI 这类 API 聚合网关接入，不要只把 `base_url` 换成网关地址。应先核对网关当前文档是否支持 Responses API、`deepseek-v4-flash` 的模型映射、SSE 结束事件和 `usage` 字段；如果网关只兼容 Chat Completions，就应按对应协议接入，不能直接套用本文的 Responses 示例。网关还可能改写错误码、模型名或限速行为，因此需要单独做一轮协议回归。

## 二、安装 SDK 并安全设置 API Key

在项目根目录创建虚拟环境：

```bash
python -m venv .venv
```

激活环境。Windows PowerShell：

```powershell
.\.venv\Scripts\Activate.ps1
```

macOS 或 Linux：

```bash
source .venv/bin/activate
```

安装 OpenAI SDK：

```bash
python -m pip install -U openai
```

不要把 Key 直接写进源码。Windows PowerShell 可以在当前终端临时设置：

```powershell
$env:DEEPSEEK_API_KEY = "<你的 API Key>"
```

macOS 或 Linux：

```bash
export DEEPSEEK_API_KEY="<你的 API Key>"
```

运行前检查变量是否存在时，只输出是否为空，不要打印完整 Key：

```python
import os

if not os.getenv("DEEPSEEK_API_KEY"):
    raise SystemExit("DEEPSEEK_API_KEY is not set")
```

如果使用 `.env` 文件，要把它加入 `.gitignore`，并检查 Git 暂存区没有包含密钥。API Key 泄露后应立即撤销并重新生成，不能只删除当前文件。

## 三、完成第一次非流式调用

创建 `responses_demo.py`：

```python
import os
from openai import OpenAI


api_key = os.environ.get("DEEPSEEK_API_KEY")
if not api_key:
    raise RuntimeError("DEEPSEEK_API_KEY is not set")

client = OpenAI(
    api_key=api_key,
    base_url="https://api.deepseek.com",
)

response = client.responses.create(
    model="deepseek-v4-flash",
    instructions="你是一个严谨的技术助手。回答时区分事实、推断和待验证信息。",
    input="请用三点说明 Responses API 与传统聊天接口在状态管理上的区别。",
)

print(response.output_text)
print("usage:", response.usage)
```

在项目根目录运行：

```bash
python responses_demo.py
```

预期结果是终端打印一段文本，并能看到 usage 信息。不要把“脚本没有抛异常”作为唯一验证；至少检查三点：

1. `response.output_text` 非空；
2. 返回的模型和请求模型一致，或能从响应日志确认路由；
3. `usage` 能记录输入和输出 Token，便于后续成本核算。

如果返回 401，先检查 Key 是否为空、是否复制了前后空格、当前 Key 是否被撤销。如果返回模型不可用，确认请求体不是旧别名，且当前账号和 API 区域可以使用 `deepseek-v4-flash`。如果返回 400，检查必填字段和输入格式，不要先盲目重试。

## 四、Responses API 是无状态的

官方兼容性表明确说明，`previous_response_id` 和 `conversation` 不支持，Responses API 仍然是无状态 API。也就是说，第二轮请求不能只发送一个新的问题，然后期待服务端自动记住上一轮上下文。

应用需要自己保存并拼接对话历史，或者自己保存业务状态后按当前请求重新组织输入。推荐把“模型上下文”和“业务状态”分开：

```text
业务数据库：用户、订单、权限和任务状态
应用上下文：当前请求需要的历史消息和工具结果
模型请求：只提交本轮确实需要的 input 与 instructions
```

不要把完整数据库记录每次都塞进上下文。先按任务筛选字段、脱敏和压缩，再发送给模型。这样更容易控制上下文长度，也更容易定位一次回答引用了哪些数据。

## 五、流式输出：按事件类型解析

流式请求只需要把 `stream` 设置为 `True`：

```python
stream = client.responses.create(
    model="deepseek-v4-flash",
    instructions="你是一个简洁的技术助手。",
    input="解释什么是幂等请求，并给出一个 API 例子。",
    stream=True,
)

for event in stream:
    if event.type == "response.output_text.delta":
        print(event.delta, end="", flush=True)
```

官方文档说明，Responses API 返回语义化 SSE 事件，事件带有 `event` 类型和递增的 `sequence_number`。流结束时使用 `response.completed`、`response.incomplete` 或 `response.failed`，不会发送传统 Chat Completions 风格的 `data: [DONE]`。

生产代码不要只拼接文本，还要记录结束状态：

```python
final_status = None
chunks = []

for event in stream:
    event_type = event.type
    if event_type == "response.output_text.delta":
        chunks.append(event.delta)
        print(event.delta, end="", flush=True)
    elif event_type in {
        "response.completed",
        "response.incomplete",
        "response.failed",
    }:
        final_status = event_type

print()
if final_status != "response.completed":
    raise RuntimeError(f"response ended with {final_status}")
```

如果你只监听 `response.output_text.delta`，网络中断、达到 `max_output_tokens` 或服务端失败时，前端可能看见一段文本却不知道响应是否完整。必须把完成、截断和失败当成三种不同状态。

## 六、检查参数是否真的被支持

官方 Responses API 兼容性文档列出的常用参数包括：

- `model`、`input`、`instructions`、`stream`；
- `temperature`、`top_p`、`max_output_tokens`；
- `top_logprobs`；
- `tools` 和 `tool_choice`；
- `reasoning.effort`；
- `text.format`；
- `user`。

同时有一批参数不支持，例如 `previous_response_id`、`conversation`、`store`、`background`、`metadata`、`include`、`truncation` 和 `service_tier`。官方特别提醒，不支持的参数可能被静默忽略，因此请求不报错不等于参数生效。

建议在接入层维护一个“实际使用参数表”：

```text
参数名 | 是否在当前文档支持 | 是否需要业务验证 | 失败时的替代方案
```

每次 SDK 升级或模型更新后跑一次参数回归。对影响输出的参数，保存请求体和响应元数据；对被忽略的参数，不要在业务文档中继续宣称它能控制模型行为。

## 七、上下文、输出和多模态边界

截至本文核对的模型价格页，V4-Flash 列出的上下文长度为 1M，最大输出长度为 384K。这个数字是文档中的模型规格，不代表你的请求一定能稳定发送到最大值。实际请求还受输入内容、输出预算、网络、客户端超时和业务验收规则影响。

Responses API 的输入内容也有明确限制：官方兼容性表说明 `message` 支持文本内容块，但图片和文件输入不支持；`input_image` 不会按图片理解，可能被替换成占位文本。需要多模态输入时，不应只把 Chat Completions 的图片字段原样迁移到 Responses API，而要确认当前接口和模型的支持范围。

## 八、常见失败与排查顺序

### 返回 400

先记录 HTTP 响应中的错误字段和请求参数，再检查模型名、输入结构、上下文长度和不支持的字段。不要先删除所有参数，因为这样会掩盖真正不兼容的字段。

### 流式响应一直不结束

检查客户端是否等待 `response.completed`、`response.incomplete` 或 `response.failed`，以及是否正确处理 SSE keep-alive 内容。不要把 `data: [DONE]` 作为唯一结束条件。

### 输出截断

检查 `max_output_tokens` 和最终事件类型。如果收到 `response.incomplete`，应记录完整 response 对象和原因，再决定压缩输入、增大预算或把任务拆成多个步骤。

### 调用成功但工具没有执行

检查工具类型是否为 function、web_search，或者 Codex 兼容场景下的 `apply_patch`。其他内置工具可能被忽略。还要检查应用是否实现了工具执行、结果回传和后续请求，而不是只提交工具定义。

## 结语

接入 DeepSeek-V4-Flash Responses API 的最小路径是：使用 `https://api.deepseek.com`，把模型写成 `deepseek-v4-flash`，先完成非流式请求，再按事件类型处理流式输出。真正进入生产前，还要验证无状态上下文、工具类型、长文本、失败结束事件和不支持参数的行为。

Responses API 的兼容性不是“字段完全等价”。把官方支持表当作接入边界，把自己的业务样本当作最终验收标准，才能避免 SDK 请求看起来成功、实际业务状态却已经丢失的问题。
