---
title: "从 Chat Completions 迁移到 Responses 要改哪三层"
category: 人工智能
tags:
  - OpenAI API
  - Responses API
  - API 迁移
description: "按请求、类型化输出和多轮状态三层迁移 OpenAI Responses API，并逐步验证文本、函数调用、结构化输出与流式事件。"
---

# 从 Chat Completions 迁移到 Responses 要改哪三层

把 `/v1/chat/completions` 替换为 `/v1/responses` 只是迁移的第一步。如果应用仍从 `choices[0].message.content` 读取文本、按旧消息结构处理工具调用，或继续累积 `messages` 管理状态，请求即使成功也可能丢失工具事件和多轮上下文。OpenAI 当前迁移指南将变化归纳为三层：发送到 Responses 端点，读取类型化 `output` Items，并选择新的状态传递方式。本文按这三层给出渐进迁移和验收顺序，避免一次改动所有 Agent 流程。

## 先确认迁移边界

OpenAI 官方文档说明 Chat Completions 仍受支持，同时推荐新项目使用 Responses。迁移不是紧急删除旧实现，可以先为一个纯文本流程建立对照。当前接口差异以 [Migrate to the Responses API](https://developers.openai.com/api/docs/guides/migrate-to-responses) 为准。

迁移前固定：

```text
OpenAI SDK 版本
实际模型标识
请求与响应样例
是否使用工具、流式、结构化输出和多轮状态
现有自动测试
store 与数据保留要求
```

一次只迁移一种能力，保留旧路径作为结果对照。

## 第一层：从 messages 请求迁移到 input

简单消息数组可以作为 Responses 的 `input`。也可以把稳定的系统级指导放在顶层 `instructions`：

```javascript
import OpenAI from "openai";

const client = new OpenAI({ apiKey: process.env.OPENAI_API_KEY });

const response = await client.responses.create({
  model: "gpt-5.6",
  instructions: "Answer concisely and identify missing evidence.",
  input: "Explain why an API migration needs contract tests."
});

console.log(response.output_text);
```

`output_text` 是 SDK 提供的便捷读取方式。它适合只需要最终文本的流程；使用工具、推理或多模态输出时，应遍历 `response.output` 并按 Item 的 `type` 分派。

验收纯文本迁移时检查：

- 请求发送到正确端点；
- `instructions` 没有在 `input` 中重复；
- `output_text` 与类型化 `output` 都能按预期读取；
- 错误路径没有继续访问 `choices`；
- `store` 行为符合应用的数据策略。

## 第二层：从 Message 读取迁移到 Item 分派

Responses 使用类型化 Items。`message`、`function_call` 和 `function_call_output` 是不同类型，应用不能假设每个输出都是文本。

建立显式分派：

```javascript
for (const item of response.output) {
  switch (item.type) {
    case "message":
      // 读取 message 的 content parts
      break;
    case "function_call":
      // 校验函数名和 arguments，交给受信控制层执行
      break;
    default:
      // 记录并处理当前应用明确支持的其他 Item 类型
      break;
  }
}
```

遇到未知类型时不要静默丢弃，也不要把对象强制转换成文本。记录非敏感类型与请求标识，按应用策略返回不支持状态。

## 函数调用必须用 call_id 对应结果

Responses 的函数定义采用扁平结构。模型返回 `function_call` Item 后，应用校验函数名称与 JSON 参数，在本地执行，再以 `function_call_output` Item 传回结果：

```javascript
const call = response.output.find(item => item.type === "function_call");

if (!call || call.name !== "lookup_order") {
  throw new Error("expected lookup_order function call");
}

const args = JSON.parse(call.arguments);
const result = await lookupOrder(args);

const next = await client.responses.create({
  model: "gpt-5.6",
  previous_response_id: response.id,
  input: [{
    type: "function_call_output",
    call_id: call.call_id,
    output: JSON.stringify(result)
  }]
});
```

`call_id` 用来关联调用与结果。函数参数仍是不可信输入，必须通过 Schema、权限和业务规则校验。完整循环与并行调用边界参见 [Function calling](https://developers.openai.com/api/docs/guides/function-calling)。

## 第三层：明确选择状态管理方式

Responses 常见的三种方式是：

1. 使用 `previous_response_id` 链接上一轮；
2. 应用保存 `response.output`，下一轮手动放回 `input`；
3. 需要持久会话对象时使用 Conversations API。

使用 `previous_response_id` 时，上一轮顶层 `instructions` 不会自动带到下一轮，需要按应用约束重新发送。官方文档也指出，响应链中的先前输入 Token 仍会作为输入计费；不能把状态链接理解成先前上下文免费。

若应用有不得服务端存储的要求，先核对官方的 `store: false` 与无状态推理处理方式，不要沿用默认值。

## 结构化输出需要迁移定义位置

Chat Completions 中的 `response_format` 在 Responses 中迁移到 `text.format`。除修改请求外，还要更新解析测试，检查：

```text
Schema 名称和严格模式
required 与 additionalProperties
拒绝或不完整输出
SDK 解析异常
模型不支持组合时的接口错误
```

不要用正则从不合法文本中补出业务关键字段。

## 流式处理按事件类型迁移

旧代码若只监听文本 `delta`，会漏掉 Item 生命周期和函数参数事件。Responses 流需要按当前 SDK 返回的类型化事件处理，至少区分创建、文本增量、Item 完成、响应完成和错误。

迁移测试故意中断一次流，确认客户端：

- 不把部分文本记为完整响应；
- 不执行尚未完成的函数参数；
- 能释放连接与本地状态；
- 保留请求标识用于排障。

事件名称应从当前官方参考和实际 SDK 类型生成，避免把旧教程中的列表写死到解析器。

## 推荐迁移顺序

```text
1. 复制一个纯文本流程，迁移 endpoint、input 和 output_text。
2. 增加 output Item 分派并覆盖未知类型。
3. 迁移一个无副作用函数，验证 call_id 循环。
4. 迁移结构化输出并执行 Schema 失败测试。
5. 选择并测试多轮状态策略。
6. 迁移流式事件与中断处理。
7. 对比错误、用量和数据存储设置后再扩大流量。
```

每一步保留 Chat Completions 基线和 Responses 候选的契约测试。不要同时更换模型、提示词和接口，否则无法定位行为差异。

## 结论与限制

Chat Completions 到 Responses 的迁移包含请求、输出 Items 和状态管理三层，工具、结构化输出与流式事件还需要各自的契约测试。先迁移纯文本，再逐项增加能力，比一次替换全部 Agent 链路更容易定位差异。

本文依据 2026 年 7 月 27 日访问的 OpenAI 官方迁移与函数调用文档，不代表任何第三方兼容服务完整支持相同字段。SDK、模型和接口能力会变化，实施时应重新核对当前官方文档和实际账号行为。
