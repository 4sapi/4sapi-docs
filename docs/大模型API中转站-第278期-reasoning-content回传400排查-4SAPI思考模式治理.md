---
title: "reasoning_content 回传 400 如何修复多轮请求"
tags:
  - 推理模型
  - 多轮对话
  - API排错
description: "多轮推理请求把 reasoning_content 原样带回下一轮时，协议字段不一致可能直接触发 400。"
---
# reasoning_content 回传 400 如何修复多轮请求
多轮推理请求把 reasoning_content 原样带回下一轮时，协议字段不一致可能直接触发 400。本文区分新请求和续写请求，说明如何记录字段来源、验证消息序列，并在无法确认能力时选择可回滚的降级处理。文中只讨论可在本地复现的步骤，不把单次结果扩展成产品承诺；每个结论都标注前提、证据和无法覆盖的边界。读者可以先完成最小验证，再根据自己的版本、权限和数据补充实验，避免把配置示例误当成普遍结论。

在接入带 thinking 或 reasoning 能力的模型时，有一种 400 错误很容易让人误判：

```text
The `reasoning_content` in the thinking mode must be passed back to the API.
```

它的真实含义不是“请你现在编一段思考过程”，而是：当前请求处于需要延续推理上下文的模式，但请求体没有带上上一轮服务端返回、且协议要求继续传递的 `reasoning_content`。


## 1. 先判断是新请求还是续写请求

`reasoning_content` 回传错误大多发生在多轮请求，而不是第一轮纯文本调用。可以先按场景区分：

| 场景 | 是否需要回传上一轮推理内容 | 常见风险 |
| --- | --- | --- |
| 新建对话、没有历史响应 | 通常不需要 | 客户端错误地注入了空字段 |
| 携带上一轮 assistant 消息 | 可能需要 | 丢失了 provider 专用字段 |
| 工具调用后继续请求 | 可能需要 | 只回传了工具结果，漏掉原调用上下文 |
| thinking 模式多轮对话 | 按上游协议要求 | 手动拼接时字段顺序或结构不完整 |
| 使用 Responses 状态链 | 通常由 response 状态承接 | 错误地把 Item 转成普通 message |

因此，第一步不是马上修改请求，而是确认这是不是一次“继续上一步推理”的请求。

## 2. reasoning_content 是什么

`reasoning_content` 不是所有 OpenAI 兼容接口都通用的标准字段。它可能是某些上游模型或网关在 thinking 模式下返回的 provider-specific 字段，用来承接模型的推理上下文。

常见的消息形态类似：

```json
{
  "role": "assistant",
  "content": "最终回答的一部分",
  "reasoning_content": "由上游服务返回的推理上下文",
  "tool_calls": []
}
```

关键点有两个：

1. 这段内容应来自上一轮 API 响应，而不是客户端临时编写；
2. 是否允许客户端读取、保存和回传，必须以对应模型和网关文档为准。

不要把“补字段”理解成“把自己的分析写进 `reasoning_content`”。伪造内容可能导致上下文不一致、签名校验失败、结果质量下降，甚至把敏感推理信息写入不该保存的日志。

## 3. 错误请求通常长什么样

问题请求经常是下面这种：

```json
{
  "model": "thinking-model",
  "messages": [
    { "role": "user", "content": "继续处理上一个任务" },
    { "role": "assistant", "content": "上一轮可见回答" },
    { "role": "tool", "content": "工具执行结果" }
  ],
  "thinking": { "type": "enabled" }
}
```

客户端保留了可见的 `content`，却丢掉了上一轮 assistant message 中的 `reasoning_content`，上游就无法验证这是不是同一条推理链。

另一个常见错误是把字段写成空字符串：

```json
{
  "reasoning_content": ""
}
```

如果协议要求的是上一轮真实内容，空字符串并不能满足要求。

## 4. 正确修复：透传服务端返回的字段

如果当前 provider 明确要求回传 `reasoning_content`，应用应在收到响应时保存结构化消息，并在下一轮按原格式透传。伪代码如下：

```javascript
const first = await client.chat.completions.create({
  model: "thinking-model",
  messages: [
    { role: "user", content: "分析这个接口的失败原因" }
  ],
  thinking: { type: "enabled" }
});

const assistantMessage = first.choices[0].message;

const nextMessages = [
  { role: "user", content: "分析这个接口的失败原因" },
  {
    role: "assistant",
    content: assistantMessage.content,
    reasoning_content: assistantMessage.reasoning_content,
    tool_calls: assistantMessage.tool_calls
  },
  { role: "user", content: "继续给出修复步骤" }
];

const second = await client.chat.completions.create({
  model: "thinking-model",
  messages: nextMessages,
  thinking: { type: "enabled" }
});
```

这段代码只是展示数据流，不代表所有 SDK 都接受同样的字段名。真实实现要确认：

```text
reasoning_content 的准确字段名。
assistant message 的嵌套位置。
tool_calls 是否必须一并回传。
thinking 参数的准确结构。
字段是否允许持久化。
```

如果 SDK 的类型定义拒绝 provider-specific 字段，不要强行绕过类型检查后直接上线。应使用该 provider 的扩展参数或网关适配层完成协议转换。

## 5. 工具调用场景更容易丢上下文

当模型先调用工具，再要求它继续回答时，下一次请求至少可能涉及三类数据：

```text
assistant 的函数调用。
工具执行结果。
上一轮思考上下文。
```

如果只保存了工具结果，丢掉 `reasoning_content` 或其他 provider-specific 字段，上游就可能返回 400。

企业网关适配时应明确区分：

| 数据 | 是否可以随意重建 | 处理建议 |
| --- | --- | --- |
| 用户消息 | 通常可以 | 按业务协议重放 |
| 工具结果 | 不能假造 | 记录工具执行状态和幂等键 |
| reasoning_content | 不能伪造 | 只透传服务端返回内容 |
| call_id/工具调用 ID | 不能改写 | 保持调用与结果关联 |
| 最终文本 | 可展示 | 按日志脱敏策略保存 |

## 6. Responses API 的替代思路

如果模型和通道支持 Responses API，优先使用它的状态机制，而不是自行拼接复杂的 `messages` 历史。

Responses 的思路是把推理、消息、函数调用和函数结果表示成不同的 Item，并通过 `previous_response_id` 或完整 Item replay 延续上下文：

```python
from openai import OpenAI

client = OpenAI()

first = client.responses.create(
    model="gpt-5.6-sol",
    input="分析这个接口的失败原因",
)

second = client.responses.create(
    model="gpt-5.6-sol",
    previous_response_id=first.id,
    input="继续给出修复步骤",
)

print(second.output_text)
```




### 7.1 客户端层

记录 SDK 版本、模型名、thinking 开关、请求轮次和是否携带历史消息。很多客户端升级后会改变字段过滤规则。

### 7.2 网关层


```text
删除了 reasoning_content。
把 assistant message 重新序列化。
把 Responses Item 转成 Chat message。
过滤了 tool_calls 或调用 ID。
把 thinking 参数改成另一种格式。
```


### 7.3 模型层

建立能力矩阵：

| 模型 | thinking | reasoning_content 回传 | function tools | Responses |
| --- | --- | --- | --- | --- |
| 目标模型 A | 以后台为准 | 以后台为准 | 以后台为准 | 以后台为准 |
| 目标模型 B | 以后台为准 | 以后台为准 | 以后台为准 | 以后台为准 |

“模型可以开启 thinking”不等于“它接受任意格式的 reasoning_content”。每个字段都要以实际文档和小流量请求验证。

### 7.4 应用层

确认应用是否把第一轮完整响应保存到下一轮，而不是只保留 `content`。如果使用数据库或队列，检查序列化和反序列化是否丢弃未知字段。


建议对每次多轮请求记录：

```text
request_id
conversation_id
response_id
project_id
environment
model
thinking_enabled
has_reasoning_content
tool_call_count
retry_count
status_code
error_code
```

不要把完整 `reasoning_content` 默认写进普通业务日志。可以记录是否存在、长度、哈希或加密存储引用；只有经过授权的排障人员才能访问原始内容。


```text
开发环境：允许测试模型和脱敏数据。
预发布环境：允许有限工具调用。
生产环境：限制 thinking 模型和原始推理日志访问。
```

这样既能做权限审计，也能避免高成本模型和敏感推理内容被任意项目调用。

## 9. 不要用重试掩盖字段错误

这类 400 是确定性的协议错误。相同请求不修改就重试，结果通常不会改变，还可能重复消耗网关日志和计费额度。

建议按错误类型处理：

| 错误 | 处理 |
| --- | --- |
| 缺少 reasoning_content | 回查上一轮响应并按协议透传 |
| 字段结构不兼容 | 修改适配器或切换正确 endpoint |
| 401/403 | 检查 Key、模型和项目权限 |
| 429 | 降低并发并退避 |
| 5xx/超时 | 有限重试并记录 request_id |

如果无法确认上一轮推理内容是否完整，宁可结束当前会话并重新发起一条新的、明确的请求，也不要伪造字段继续链式调用。

## 10. 上线前检查清单

```text
[ ] 区分新请求和多轮续写请求。
[ ] 已确认 reasoning_content 的字段名和嵌套位置。
[ ] 不会手写或伪造服务端推理内容。
[ ] assistant 消息、工具调用和工具结果不会丢字段。
[ ] Responses 状态链和 Chat message replay 没有混用。
[ ] 上游 API 网关不会静默删除 provider-specific 字段。
[ ] request_id、response_id 和 conversation_id 可关联。
[ ] 原始推理内容不会进入普通日志和前端。
[ ] 400 协议错误不会无限自动重试。
[ ] 生产 Key、模型权限和 thinking 能力已隔离验证。
```

## 11. 成本与风险提示

多轮 thinking 请求会增加上下文、输出和存储成本。把完整推理内容持久化还会扩大数据泄露面。


- reasoning 模型白名单；
- 单会话最大轮数；
- 最大输入和输出 Token；
- 原始推理内容的访问审批；
- 400 协议错误告警；
- 按项目和环境的预算控制。

不要为了“让请求通过”而关闭所有安全过滤，也不要把模型内部推理内容当作普通业务文本长期传播。供应商条款、数据保留和访问权限需要单独审查。

## 12. 总结

`The reasoning_content in the thinking mode must be passed back` 的核心含义是：当前多轮 thinking 请求缺少上一轮服务端返回的推理上下文，或者中间适配层把它丢失了。

正确做法不是随便补一段文字，而是：

```text
确认请求是否需要续接上下文。
完整保存并按协议透传上一轮字段。
检查工具调用和调用 ID 是否一起保留。
能用 Responses 状态链时，避免手工拼接隐藏字段。
通过 上游 API 做协议、权限、日志和成本治理。
```



## 官方来源

- OpenAI：Migrate to the Responses API：https://developers.openai.com/api/docs/guides/migrate-to-responses
- OpenAI：Conversation state：https://developers.openai.com/api/docs/guides/conversation-state

## 结论

本文给出了问题定位、配置或验证的可执行路径。实际结果仍取决于当前版本、权限和运行环境，提交前应按官方文档复核可变字段，并保留失败证据和回滚边界。
