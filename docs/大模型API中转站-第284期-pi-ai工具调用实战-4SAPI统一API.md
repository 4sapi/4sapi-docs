---
title: "pi-ai 如何完成文本、流式与工具调用"
tags:
  - Pi
  - TypeScript
  - 工具调用
description: "直接使用 pi-ai 时，文本生成、流式事件和工具调用会共享一套消息抽象，但验证方式并不相同。"
---
# pi-ai 如何完成文本、流式与工具调用
直接使用 pi-ai 时，文本生成、流式事件和工具调用会共享一套消息抽象，但验证方式并不相同。本文用一个最小 TypeScript 程序逐步检查安装、Provider、事件顺序和工具结果，帮助你定位是 SDK、模型还是业务代码的问题。文中只讨论可在本地复现的步骤，不把单次结果扩展成产品承诺；每个结论都标注前提、证据和无法覆盖的边界。读者可以先完成最小验证，再根据自己的版本、权限和数据补充实验，避免把配置示例误当成普遍结论。

先说清楚边界：`pi-ai` 是统一模型 API 层，不是完整的业务 Agent。它提供 Provider、上下文、流式事件和 Tool Calling 的标准接口；工具真正执行在哪里，仍然由你的程序决定。这个边界让测试、权限和审计更容易控制。

## 1. pi-ai 的核心抽象

Pi 官方 README 可以用四个对象来理解：

```text
Provider -> 模型集合和认证方式
Model    -> 模型 ID、协议、能力和成本元数据
Context  -> system prompt、消息、工具和时间戳
Stream   -> 文本、思考、工具调用、结束和错误事件
```


```text
TypeScript 应用
  -> createProvider()
  -> openAICompletionsApi()
  -> https://api.example.com/v1/chat/completions
  -> 上游 API 模型
  -> AssistantMessage / Tool Call
```


## 2. 安装最小依赖

新建一个 TypeScript 项目：

```bash
mkdir pi-provider-lab
cd pi-provider-lab
npm init -y
npm install @earendil-works/pi-ai
npm install -D tsx typescript
```

准备环境变量：

PowerShell：

```powershell
$env:API_KEY = "你的上游 API令牌"
$env:MODEL_ID = "claude-sonnet-4-5-20250929"
```

Linux/macOS：

```bash
export API_KEY='你的上游 API令牌'
export MODEL_ID='claude-sonnet-4-5-20250929'
```



下面是一个完整的最小 Provider。它使用 Pi 官方文档中的 `createProvider`、`envApiKeyAuth` 和 `openAICompletionsApi`。

```typescript
import {
  createModels,
  createProvider,
  envApiKeyAuth,
  type Context,
  type Model,
} from "@earendil-works/pi-ai";
import { openAICompletionsApi } from "@earendil-works/pi-ai/api/openai-completions.lazy";

const modelId = process.env.MODEL_ID ?? "claude-sonnet-4-5-20250929";

function readNumber(name: string, fallback: number): number {
  const value = Number(process.env[name]);
  return Number.isFinite(value) ? value : fallback;
}

const model: Model<"openai-completions"> = {
  id: modelId,
  name: `上游 API ${modelId}`,
  api: "openai-completions",
  provider: "provider",
  baseUrl: "https://api.example.com/v1",
  reasoning: false,
  input: ["text"],
  contextWindow: readNumber("FOURSAPI_CONTEXT_WINDOW", 128000),
  maxTokens: readNumber("FOURSAPI_MAX_TOKENS", 16384),
  cost: {
    input: readNumber("FOURSAPI_INPUT_PRICE", 0),
    output: readNumber("FOURSAPI_OUTPUT_PRICE", 0),
    cacheRead: readNumber("FOURSAPI_CACHE_READ_PRICE", 0),
    cacheWrite: readNumber("FOURSAPI_CACHE_WRITE_PRICE", 0),
  },
};

const models = createModels();
const fourSapi = createProvider({
  id: "provider",
  name: "上游 API",
  baseUrl: "https://api.example.com/v1",
  auth: {
    apiKey: envApiKeyAuth("上游 API API Key", ["API_KEY"]),
  },
  models: [model],
  api: openAICompletionsApi(),
});

models.setProvider(fourSapi);

const context: Context = {
  systemPrompt: "You are a concise technical assistant.",
  messages: [
    {
      role: "user",
      content: "用三句话说明为什么 Agent 需要工具调用。",
      timestamp: Date.now(),
    },
  ],
};

const response = await models.complete(model, context);
for (const block of response.content) {
  if (block.type === "text") {
    console.log(block.text);
  }
}

console.log("usage:", response.usage);
```

运行：

```bash
npx tsx src/basic.ts
```


## 4. 先理解流式事件，再处理最终消息

`models.stream()` 返回的是事件流，不是一个直接包含完整文本的 Promise。最小的流式打印代码如下：

```typescript
const stream = models.stream(model, context);

for await (const event of stream) {
  switch (event.type) {
    case "start":
      console.log(`model: ${event.partial.model}`);
      break;
    case "text_delta":
      process.stdout.write(event.delta);
      break;
    case "thinking_delta":
      process.stdout.write(event.delta);
      break;
    case "toolcall_end":
      console.log(`\\nTool call: ${event.toolCall.name}`);
      break;
    case "done":
      console.log(`\\nstop reason: ${event.reason}`);
      break;
    case "error":
      console.error("provider error:", event.error.errorMessage);
      break;
  }
}

const assistantMessage = await stream.result();
context.messages.push(assistantMessage);
```

事件流的好处是 UI 可以实时显示，最终消息的好处是可以进入下一轮上下文。两者不要混淆：只打印 delta 而不保存最终消息，后续工具调用会丢失上一轮模型输出；只等待最终消息，又会让用户误以为请求卡住。

## 5. 用一个无副作用工具验证 Tool Calling

工具调用应该从不会修改系统状态的函数开始。下面的工具只返回当前时间字符串，不读文件、不执行命令，也不访问生产数据。

```typescript
import { Type, type Tool } from "@earendil-works/pi-ai";

const getTimeTool: Tool = {
  name: "get_time",
  description: "Return the current time for a requested IANA timezone.",
  parameters: Type.Object({
    timezone: Type.Optional(
      Type.String({ description: "For example, Asia/Shanghai or UTC." }),
    ),
  }),
};

const toolContext: Context = {
  systemPrompt:
    "When the user asks for the current time, call get_time instead of guessing.",
  messages: [
    {
      role: "user",
      content: "告诉我上海现在几点，并在工具返回后再回答。",
      timestamp: Date.now(),
    },
  ],
  tools: [getTimeTool],
};

const firstMessage = await models.complete(model, toolContext);
toolContext.messages.push(firstMessage);

const toolCalls = firstMessage.content.filter(
  (block) => block.type === "toolCall",
);

for (const call of toolCalls) {
  const timezone =
    typeof call.arguments.timezone === "string"
      ? call.arguments.timezone
      : "UTC";
  const result = new Intl.DateTimeFormat("zh-CN", {
    timeZone: timezone,
    dateStyle: "full",
    timeStyle: "long",
  }).format(new Date());

  toolContext.messages.push({
    role: "toolResult",
    toolCallId: call.id,
    toolName: call.name,
    content: [{ type: "text", text: result }],
    isError: false,
    timestamp: Date.now(),
  });
}

if (toolCalls.length > 0) {
  const finalMessage = await models.complete(model, toolContext);
  for (const block of finalMessage.content) {
    if (block.type === "text") {
      console.log(block.text);
    }
  }
}
```

完整循环是：

```text
用户消息
  -> 上游 API 模型返回 toolCall
  -> 应用校验工具名称和参数
  -> 应用执行无副作用函数
  -> 写入 toolResult
  -> 上游 API 模型生成最终文本
```

真实系统中不能因为模型提出了工具调用就直接执行。至少要做三层校验：

1. 工具名称必须在白名单中。
2. 参数必须通过 schema 和业务规则校验。
3. 涉及写文件、发消息、付款、部署或删除数据时，必须经过人工确认或隔离环境。



```text
是否接受 tools 字段
是否返回 tool_calls
工具参数是否以 JSON 形式返回
流式模式是否支持部分工具参数
思考内容回传后能否正确重放
```

如果模型只支持文本生成，不要为了让它“看起来像 Agent”而用提示词伪造工具调用。把任务改成纯文本流程，或者选择当前分组中明确支持 Tool Calling 的模型。

如果出现 400，先把 `tools` 移除，做一次最小文本请求：

```typescript
const textOnlyContext: Context = {
  systemPrompt: "Return a short plain-text answer.",
  messages: [
    {
      role: "user",
      content: "输出一句文本。",
      timestamp: Date.now(),
    },
  ],
};

const textOnly = await models.complete(model, textOnlyContext);
console.log(textOnly.content);
```

如果文本成功、带工具失败，根因通常在模型能力、工具 schema 或上游兼容字段，而不是 API Key。

## 7. Pi 侧 usage 和网关侧计费怎么核对


| 记录层 | 建议记录 | 用途 |
| --- | --- | --- |
| Pi 应用 | session ID、provider、model、输入/输出 Token、工具轮数 | 定位一次 Agent 任务 |
| 业务系统 | 项目、用户、任务类型、结果状态 | 分析业务成本 |

产生差异时，优先检查：

```text
模型 ID 是否一致
是否启用了缓存
是否发生了重试
是否包含工具续跑的多轮请求
上游 API 分组倍率和计费规则是否变化
```

不要把示例代码中的 `cost: 0` 当成免费调用。它只是没有填入真实模型单价时的本地默认值。

## 8. 常见错误的处理顺序

```text
1. 先用 curl 验证 https://api.example.com/v1/chat/completions。
2. 再验证 pi-ai 的纯文本 complete。
3. 再验证 stream 事件。
4. 最后增加 tools 和多轮上下文。
```

错误码可以这样分层：

| 状态码 | 优先检查 |
| --- | --- |
| 400 | 协议字段、模型能力、工具 schema、Base URL |
| 401/403 | Key、令牌分组、模型权限和额度 |
| 429 | 并发、限流、余额和请求重试策略 |
| 504/524 | 客户端超时、流式连接和请求体大小 |

## 9. 本篇小结


```text
纯文本完成
  -> 流式输出
  -> 无副作用工具
  -> 工具结果续跑
  -> 业务工具和权限治理
```





## 资料来源

- Pi AI README：<https://github.com/earendil-works/pi/blob/main/packages/ai/README.md>
- Pi Custom Provider：<https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/custom-provider.md>

## 结论

本文给出了问题定位、配置或验证的可执行路径。实际结果仍取决于当前版本、权限和运行环境，提交前应按官方文档复核可变字段，并保留失败证据和回滚边界。
