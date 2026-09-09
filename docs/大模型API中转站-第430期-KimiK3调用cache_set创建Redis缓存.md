---
title: "Kimi K3 如何调用 cache_set 创建 Redis 缓存"
category: 人工智能
tags:
  - Function Calling
  - Redis
  - JSON Schema
description: "说明聊天模型如何返回 cache_set 工具参数，由后端校验后写入 Redis，再把执行结果送回模型完成一次工具调用。"
---

# Kimi K3 如何调用 cache_set 创建 Redis 缓存

“让模型创建缓存”容易产生一个误解：模型并不会直接连接 Redis，也不会因为提示词里写了“请缓存”就获得持久化能力。可执行的做法是把 `cache_set` 注册为工具，模型只生成 `key`、`value` 和 `ttl` 等参数；后端校验参数、执行 Redis 写入，再把结果作为工具消息返回。本文以 `moonshotai/Kimi-K3` 这个请求标识演示完整链路，但不假定任意供应商都已启用该模型或 Function Calling。

## 1. 工具调用链路是什么

一次缓存写入包含两次模型请求和一次后端操作：

```text
用户提出缓存需求
        ↓
模型返回 cache_set 的 tool_calls
        ↓
后端解析并校验 arguments
        ↓
后端执行 Redis SET
        ↓
后端把执行结果作为 role=tool 发回模型
        ↓
模型生成最终答复
```

其中只有 Redis 执行步骤真正改变数据。模型返回了工具参数，不等于缓存已经创建成功。

## 2. 第一次请求：声明 cache_set

下面是 OpenAI-compatible Chat Completions 形态的 Body：

```json
{
  "model": "moonshotai/Kimi-K3",
  "messages": [
    {
      "role": "user",
      "content": "请缓存用户 user_123 的资料：姓名张三，年龄28岁，有效期1小时。"
    }
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "cache_set",
        "description": "创建或更新一条用户资料缓存",
        "parameters": {
          "type": "object",
          "properties": {
            "key": {
              "type": "string",
              "description": "缓存键，必须使用 user: 前缀"
            },
            "value": {
              "type": "object",
              "properties": {
                "name": {"type": "string"},
                "age": {"type": "integer", "minimum": 0}
              },
              "required": ["name", "age"],
              "additionalProperties": false
            },
            "ttl": {
              "type": "integer",
              "description": "有效期，单位为秒",
              "minimum": 60,
              "maximum": 86400
            }
          },
          "required": ["key", "value", "ttl"],
          "additionalProperties": false
        }
      }
    }
  ],
  "tool_choice": {
    "type": "function",
    "function": {
      "name": "cache_set"
    }
  }
}
```

这里把 `tool_choice` 指向 `cache_set`，用于这个明确要求写缓存的测试。正常业务可以使用 `"tool_choice": "auto"`，但是否执行写操作不应只交给模型判断；后端仍要执行权限和风险检查。

示例把 TTL 限制在 60 至 86400 秒。这是本文测试策略，不是 Redis 的固定限制。实际范围应根据业务的数据新鲜度、容量和删除策略确定。

## 3. 模型会返回什么

模型可能返回如下工具调用：

```json
{
  "choices": [
    {
      "message": {
        "role": "assistant",
        "tool_calls": [
          {
            "id": "call_cache_001",
            "type": "function",
            "function": {
              "name": "cache_set",
              "arguments": "{\"key\":\"user:user_123\",\"value\":{\"name\":\"张三\",\"age\":28},\"ttl\":3600}"
            }
          }
        ]
      }
    }
  ]
}
```

`arguments` 常以 JSON 字符串出现，所以后端需要先解析。不要把它直接拼成 Redis 命令或数据库语句。

正确的验收结果包括：

```text
工具名是 cache_set
key 为 user:user_123
value 只包含允许字段
1 小时被换算为 3600 秒
不存在 Schema 之外的参数
```

如果当前供应商返回普通文本而不是 `tool_calls`，先核对模型是否支持工具调用、接口是否透传 `tools`，以及响应格式是否与文档一致。

## 4. 后端如何执行 Redis 写入

下面的 Node.js 片段只展示关键校验与写入逻辑，假设项目已经创建 Redis 客户端 `redis`：

```javascript
const toolCall = response.choices[0]?.message?.tool_calls?.[0];

if (!toolCall || toolCall.function.name !== "cache_set") {
  throw new Error("Expected cache_set tool call");
}

const args = JSON.parse(toolCall.function.arguments);

if (!/^user:[A-Za-z0-9_-]+$/.test(args.key)) {
  throw new Error("Invalid cache key");
}

if (!Number.isInteger(args.ttl) || args.ttl < 60 || args.ttl > 86400) {
  throw new Error("Invalid cache TTL");
}

const value = {
  name: String(args.value.name),
  age: Number(args.value.age)
};

if (!Number.isInteger(value.age) || value.age < 0) {
  throw new Error("Invalid age");
}

await redis.set(args.key, JSON.stringify(value), { EX: args.ttl });
```

JSON Schema 能减少错误参数，但不能替代服务端校验。攻击者可以绕过模型直接调用接口，模型也可能生成格式正确但业务上无权写入的 Key。

## 5. 把工具结果送回模型

Redis 写入成功后，后端继续发送第二次模型请求。`tool_call_id` 必须与第一次返回的 ID 对应：

```json
{
  "model": "moonshotai/Kimi-K3",
  "messages": [
    {
      "role": "user",
      "content": "请缓存用户 user_123 的资料：姓名张三，年龄28岁，有效期1小时。"
    },
    {
      "role": "assistant",
      "tool_calls": [
        {
          "id": "call_cache_001",
          "type": "function",
          "function": {
            "name": "cache_set",
            "arguments": "{\"key\":\"user:user_123\",\"value\":{\"name\":\"张三\",\"age\":28},\"ttl\":3600}"
          }
        }
      ]
    },
    {
      "role": "tool",
      "tool_call_id": "call_cache_001",
      "content": "{\"ok\":true,\"key\":\"user:user_123\",\"ttl\":3600}"
    }
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "cache_set",
        "description": "创建或更新一条用户资料缓存",
        "parameters": {
          "type": "object",
          "properties": {
            "key": {
              "type": "string",
              "description": "缓存键，必须使用 user: 前缀"
            },
            "value": {
              "type": "object",
              "properties": {
                "name": {"type": "string"},
                "age": {"type": "integer", "minimum": 0}
              },
              "required": ["name", "age"],
              "additionalProperties": false
            },
            "ttl": {
              "type": "integer",
              "description": "有效期，单位为秒",
              "minimum": 60,
              "maximum": 86400
            }
          },
          "required": ["key", "value", "ttl"],
          "additionalProperties": false
        }
      }
    }
  ]
}
```

最终答复可以告诉用户缓存已创建，但依据应是工具返回的 `ok: true`，不是模型之前产生了 `tool_calls`。如果 Redis 失败，应返回结构化错误，而不是伪造成功结果：

```json
{
  "ok": false,
  "error_type": "redis_unavailable",
  "retryable": true
}
```

## 6. 写缓存前必须加的边界

生产实现至少应检查：

- 当前用户是否有权写这个 Key；
- Key 是否限定在允许的命名空间；
- value 是否包含个人信息、凭证或不应缓存的数据；
- TTL 是否在业务允许范围内；
- 单条 value 大小与写入频率是否受限；
- 工具调用、执行结果和失败原因是否留下脱敏日志；
- 重试是否可能覆盖新值或造成重复副作用。

对于涉及账号、订单、权限或生产配置的缓存写入，建议由后端根据可信身份生成 Key，不让模型自由决定完整 Key。

## 7. 它和 Prompt Caching 不是一回事

本文的 `cache_set` 是业务缓存：后端把指定数据写入 Redis，应用随后可以按 Key 读取。

Prompt Caching 是模型 API 供应商对重复提示前缀的复用机制，通常由供应商定义请求字段、命中条件、有效期和 usage 指标。两者不能互相替代：调用 `cache_set` 不会自动降低模型输入 token，Prompt Caching 也不会替应用保存任意用户资料。

## 8. 结论与限制

让 Kimi K3“创建缓存”的可执行含义，是让模型生成受 Schema 约束的 `cache_set` 参数，再由可信后端完成 Redis 写入。只有后端成功返回工具结果，才能对用户声称缓存已经创建。

本文没有使用真实 Key 验证 `moonshotai/Kimi-K3` 在特定供应商账号上的可用性。接入前应通过当前模型列表与 Function Calling 文档核对支持情况，并在测试 Redis 命名空间中完成成功、拒绝和超时三类测试。

## 官方来源

- [Together AI：Function Calling](https://docs.together.ai/docs/inference/function-calling/overview)
- [Redis：SET 命令](https://redis.io/docs/latest/commands/set/)
- [JSON Schema：Object](https://json-schema.org/understanding-json-schema/reference/object)

---

*本文接口结构与链接核对日期：2026-08-10。模型支持、工具调用格式和 SDK 行为可能变化，请以目标供应商当前文档和实际响应为准。*
