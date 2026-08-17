---
title: "Responses 工具调用如何建立可追踪审计链"
category: 人工智能
tags:
  - Responses API
  - 工具调用
  - 可观测性
description: "用业务 request_id、OpenAI response.id、输出 Item 类型和 call_id 关联 Responses 函数调用与结果，并控制日志中的敏感参数。"
---

# Responses 工具调用如何建立可追踪审计链

Responses API 会把消息、函数调用和函数结果表达为不同 Item。协议结构更清楚，但审计不会自动完成：业务仍需回答哪位调用者发起任务、模型提出了哪个函数、控制层是否批准、工具实际执行了什么，以及哪个结果被传回后续响应。若只保存最终文本，故障发生时无法还原工具链；若完整记录提示词和参数，又可能复制敏感数据。本文设计一条以 `request_id`、`response.id` 和 `call_id` 为骨架的最小审计链。

## 四种标识各自解决什么

| 标识 | 生成方 | 用途 |
| --- | --- | --- |
| `request_id` | 业务应用 | 关联一次用户操作和内部日志 |
| `response.id` | OpenAI Responses API | 标识一次模型响应 |
| Item `id` | Responses API | 标识某个输出 Item |
| `call_id` | 函数调用 Item | 关联函数调用与 `function_call_output` |

`call_id` 不是用户身份，也不替代业务授权记录。应用先生成 `request_id`，再把收到的 API 与工具标识挂到同一业务任务上。

OpenAI 当前迁移文档确认 Responses 使用类型化 `output` Items，函数调用和结果通过 `call_id` 关联；具体结构参见 [Migrate to the Responses API](https://developers.openai.com/api/docs/guides/migrate-to-responses)。

## 建立三段式事件链

### 1. 模型请求事件

发送请求前记录：

```text
request_id
actor_id_hash
project_id
operation
model_requested
prompt_version
toolset_version
started_at
```

响应到达后追加：

```text
response_id
model_returned
response_status
output_item_types
usage_fields
completed_at
```

`actor_id_hash` 使用组织认可的稳定匿名标识，不发送或记录用户邮箱等直接身份信息。

### 2. 工具决策与执行事件

遍历 `response.output`，遇到 `function_call` 时记录：

```text
request_id
response_id
item_id
call_id
tool_name
argument_schema_version
argument_digest
authorization_decision
approval_id
```

`argument_digest` 用于判断两次调用参数是否一致，不能代替实际参数审查。工具控制层仍需解析参数、执行 Schema 校验、检查调用者权限和资源范围。

执行后记录：

```text
execution_id
started_at
completed_at
result_status
error_category
side_effect_reference
output_digest
```

外部写入使用幂等标识，并在 `side_effect_reference` 保存业务资源 ID，而不是完整写入内容。

### 3. 工具结果回传事件

构造 `function_call_output` 时记录它使用的同一个 `call_id`、关联 `execution_id` 和下一次业务 `request_id`。不要生成新的 `call_id` 代替模型返回值。

OpenAI [Function calling guide](https://developers.openai.com/api/docs/guides/function-calling) 给出了函数调用循环。应用需要在每轮继续维护自己的业务关联标识。

## 一个最小处理骨架

```javascript
const response = await client.responses.create(request);

for (const item of response.output) {
  if (item.type !== "function_call") continue;

  const auditContext = {
    requestId,
    responseId: response.id,
    itemId: item.id,
    callId: item.call_id,
    toolName: item.name
  };

  const args = validateArguments(item.name, item.arguments);
  await authorizeTool(actor, item.name, args);
  const result = await executeTool(item.name, args, auditContext);

  await client.responses.create({
    model,
    previous_response_id: response.id,
    input: [{
      type: "function_call_output",
      call_id: item.call_id,
      output: JSON.stringify(result)
    }]
  });
}
```

示例省略了并行调用、错误、重试和多轮循环，不能直接作为完整生产实现。关键点是审计上下文在工具执行前建立，并沿相同 `call_id` 传到结果回传。

## 并行调用不能只取第一个 Item

一次响应可能包含多个函数调用。不能用 `find()` 取第一个后忽略其余 Items。为每个 `function_call` 创建独立执行记录，再按应用策略并行或串行处理。

若并行工具有相互依赖，控制层必须明确顺序。模型返回多个调用不等于它们可以安全并行执行。

## 参数日志采用分级策略

默认记录：

```text
工具名
Schema 版本
字段名列表
参数大小
参数摘要或哈希
权限结果
错误类别
```

默认不记录：

```text
完整提示词与用户输入
文件正文和数据库结果
访问令牌与 Cookie
个人信息和客户标识
工具返回的完整敏感内容
```

确需保留原始参数进行安全调查时，使用隔离存储、严格访问控制、明确用途和删除期限。普通应用日志只保存指向受控证据的标识。

## 记录 requested 与 returned model

请求模型可能是可变别名，代理也可能执行路由。保存请求值和 Responses 返回的实际模型值。若中间层替换模型，额外记录 `effective_model` 与策略标识，不能让审计者从名称猜测。

同样，保存所用端点与 `store` 设置。状态管理方式影响调查时可从服务端取得什么，不能在事故后才推断。

## 错误处理保持 call_id 语义

工具执行失败时，不要悄悄丢弃调用。记录稳定错误类别和副作用状态，再决定是否把错误结果作为工具输出交回模型、请求人工处理或结束任务。

超时尤其需要区分：

```text
确认未执行
确认执行失败
执行成功但响应丢失
状态未知
```

状态未知时先用业务幂等标识查询，不能直接重复写入。模型无法仅凭超时文本判断外部副作用。

## 用完整性检查发现断链

定期检查：

- 每个 `function_call` 是否有授权决定；
- 获准执行的调用是否有 `execution_id`；
- 每个回传结果是否引用存在的 `call_id`；
- 外部写入是否有幂等标识和资源回执；
- 未完成任务是否有明确终止原因；
- 日志是否出现凭据或禁止字段。

缺少工具结果不一定是系统错误，可能是调用被拒绝或任务中断；审计链必须记录对应原因。

## 结论与限制

Responses 的类型化 Items 和 `call_id` 提供了关联函数调用与结果的协议键，但业务审计仍需应用自己实现。以 `request_id` 连接业务身份，以 `response.id` 和 Item ID 定位模型输出，以 `call_id` 连接工具执行和回传，才能还原一次 Agent 任务。

本文依据 2026 年 7 月 27 日访问的 OpenAI 官方迁移与函数调用文档，只覆盖函数工具的通用审计结构。内置工具、远程 MCP、状态存储和数据保留有各自字段与政策，实施前需要按当前官方文档和组织要求单独核对。
