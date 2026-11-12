---
title: "Kimi K3.1 Pro 工具调用编排：动态加载工具与完整回传"
category: 人工智能
tags:
  - Kimi K3.1 Pro
  - 工具调用
  - Agent 工程
description: "工具数量到几十个以后，把所有定义塞进请求会同时抬高成本和选错工具的概率。本文给出一套先检索、再按需注入、并正确回传思考内容的编排方式，以及可核验的检查清单。"
brand_mode: example
---

# Kimi K3.1 Pro 工具调用编排：动态加载工具与完整回传

Agent 做到一定规模，工具调用最常见的两种失败是：模型选错了工具，以及多轮之后请求直接报 400。前者通常被归因为"模型不够聪明"，后者常被当成"接口有 bug"。这两个现象里，很大一部分其实来自编排方式：工具定义一次性全塞进请求，历史消息只保留 `content`。

K3.1 Pro 的文档给了两条明确的约束：工具定义应该按需注入，而不是全量声明；多轮对话和工具调用必须把 API 返回的完整 assistant message 原样回传，包括 `reasoning_content` 和 `tool_calls`。本文把这两条落成一套可执行的流程，并给出每一步的验证方法。

文中接口行为引自 Kimi 官方文档，检索日期 2026-10-26，相关页面最后更新 2026-10-23。示例中的工具实现是占位逻辑，真实工具需要你自己接后端。

## 问题出在工具定义本身

一个工具定义包含名称、描述和 JSON Schema 参数。工具越多，这部分固定开销越大，而且它会挤占模型对用户真实意图的注意力。

官方给出的判断很直接：当 Agent 的可用工具达到几十上百个时，不要把所有工具定义一次性放进请求——它们会占掉大量上下文，还会让模型更容易选错工具。

这句话背后有三层成本，容易只看到第一层。

| 成本层 | 表现 | 是否容易察觉 |
| --- | --- | --- |
| Token 成本 | 每轮请求都重复发送全部工具定义 | 容易，账单可见 |
| 选择准确率 | 相似工具之间混淆，参数填错 | 中等，需要看日志 |
| 前缀缓存 | 工具定义一变，后面全部失效 | 难，账单上是"命中率下降" |

第三层最隐蔽。工具定义通常放在请求前部，一旦顺序或内容变动，缓存前缀就断了，命中价和未命中价差 10 倍，账单会以"莫名其妙变贵"的形式表现出来。

## 三层结构的编排方式

官方给的做法可以拆成三层，各管一件事。

**顶层核心工具。** 每轮都会用到的少量工具（比如搜索工具本身、状态查询）放在请求顶层 `tools`，声明后内容保持不变。

**工具检索工具。** 顶层只声明一个 `search_tools`，由你的后端实现，按关键词返回候选工具的名称和简介。

**动态注入的工具。** 检索出候选之后，把对应工具的完整声明，通过一条不含 `content` 的 `system` 消息插到 `messages` 里。

来源：[Kimi K3 API 工具调用最佳实践](https://platform.kimi.com/docs/guide/kimi-k3-tool-calling-best-practice)、[动态加载工具](https://platform.kimi.com/docs/guide/use-dynamic-tool-loading)。

第一步，顶层只声明检索工具：

```json
{
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "search_tools",
        "description": "按关键词搜索可用工具，返回工具名称和简介",
        "parameters": {
          "type": "object",
          "properties": {
            "query": {
              "type": "string",
              "description": "搜索关键词，例如 github、database"
            }
          },
          "required": ["query"]
        }
      }
    }
  ]
}
```

同时在 system prompt 里写明工具目录有哪些业务域，让模型知道"该去搜什么"。不写这一段，模型往往凭记忆直接作答，跳过检索。

第二步，首轮强制检索。模型有权不调用任何工具，为了确保它先查再答，首轮把 `tool_choice` 设为 `required`：

```json
{
  "model": "kimi-k3.1-pro",
  "messages": [{"role": "user", "content": "帮我创建一个 GitHub PR"}],
  "tools": ["..."],
  "tool_choice": "required"
}
```

检索完成后，后续请求把 `tool_choice` 恢复为 `"auto"`。**改 `tool_choice` 不会破坏前缀缓存**，这一点让它可以按请求粒度调整，不必担心成本。

第三步，按需注入。`search_tools` 返回候选后，由应用侧把完整声明插进 `messages`：

```json
{
  "role": "system",
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "create_github_pr",
        "description": "在指定仓库创建 Pull Request",
        "parameters": {
          "type": "object",
          "properties": {}
        }
      }
    }
  ]
}
```

三个细节必须记住：这条 `system` 消息**不带 `content`**；声明格式与顶层 `tools` 完全一致，不需要维护两套 schema；声明从这里所在的位置起对模型可见，**按请求生效，服务端不会记住**。

第四步，回传。拿到 `tool_calls` 后执行工具，把完整 assistant message 和每条 `tool_call_id` 对应的结果追加进历史，再发下一轮。官方给出最小循环如下（省略了真实工具实现）：

```python
import json

messages = [{"role": "user", "content": "北京今天天气怎么样？"}]

first = client.chat.completions.create(
    model="kimi-k3.1-pro",
    messages=messages,
    tools=tools,
    tool_choice="required",
)
assistant_message = first.choices[0].message
messages.append(assistant_message)          # 整条回传，不要只留 content

for tool_call in assistant_message.tool_calls or []:
    arguments = json.loads(tool_call.function.arguments)
    result = json.dumps({"city": arguments["city"], "weather": "晴"}, ensure_ascii=False)
    messages.append({"role": "tool", "tool_call_id": tool_call.id, "content": result})

final = client.chat.completions.create(
    model="kimi-k3.1-pro",
    messages=messages,
    tools=tools,
)
print(final.choices[0].message.content)
```

关键在于 `messages.append(assistant_message)` 这一行。它同时带上了 `reasoning_content` 和 `tool_calls`，这是 K3 系列保留式思考的要求，也是下一节要展开的失败点。

## 动态注入对缓存的影响

注入顺序决定缓存命中率，官方给了一张对照表，可以直接当规则用：

| 操作 | 对前缀缓存的影响 |
| --- | --- |
| 在 `messages` 末尾追加工具声明 | 不影响已有前缀缓存 |
| 后续请求原样保留已注入的声明 | 前缀稳定，有利于持续命中 |
| 删除、修改中间消息，或在中间插入声明 | 变更位置之后可能无法命中 |
| 在顶层 `tools` 声明全局工具 | 不影响命中 |

由此得到三条操作纪律。

**追加，不要插入。** 新工具一律追加到末尾。任何插入中间的动作，都会让变更点之后的缓存失效。

**保留已注入的声明。** 声明不会被服务端记住。下一轮要么原样带着它（工具继续可用、前缀保持稳定），要么明确移除并接受该工具失效。移除之后如果工具没在别处声明，模型就调不到它，同时后续前缀也可能无法命中。

**核心工具固定顶层，之后不改。** 顶层声明不影响缓存，稳定即可。

还有一条容易漏的：缓存按块存储，不足一整块的尾部无法写入缓存，会按未命中计费。长前缀场景下这不是大钱，但对账时会看到"明明前缀没变却没完全命中"的残差。

## 用检查清单代替"感觉能跑"

工具编排的问题很少当场报错，多数是行为偏差。建议每次都过一遍下面这张表。

| 检查项 | 最小验证方法 | 失败信号 |
| --- | --- | --- |
| 工具是否可见 | 首轮是否调用 `search_tools` | 模型直接凭记忆作答 |
| 首轮是否强制检索 | 请求里 `tool_choice` 的值 | 值为 `auto` 却期望必检索 |
| 声明是否完整 | 检查注入的 `name` / `description` / `parameters` | 调用返回参数校验错误 |
| 历史是否完整 | 打印每条 assistant message 的字段 | 缺 `reasoning_content` 或 `tool_calls` |
| 注入位置 | 打日志确认声明追加在末尾 | 命中率突然掉到接近 0 |
| 重复调用 | 统计同一工具的连续调用次数 | 同一工具反复调用无进展 |
| 思考开销 | `usage` 与 `reasoning_content` 长度 | 成本高于预期数倍 |

最后一行值得单独说明。`reasoning_content` 会计入 token 消耗，开启保留式思考后，历史思考内容会持续占用上下文长度并计费。所以"回传完整 message"和"控制上下文成本"是一对需要权衡的约束：回传是对的，但长会话要么定期裁剪历史（从最老的轮次开始整段删除，不要从中间删），要么把稳定的参考材料前移以利用缓存。

## 五个常见失败

**1. 400 报错，提示缺少 reasoning_content。** 原因是从历史里裁掉了 assistant 的思考字段。修复方式是把 API 返回的 message 原样 append 回去，不要自己重新拼装字典。

**2. SDK 读不到 `reasoning_content`。** openai SDK 的 `ChatCompletionMessage` 和 `ChoiceDelta` 类型里没有这个字段，直接 `.reasoning_content` 会失败。正确写法是先判断再取：

```python
message = completion.choices[0].message
if hasattr(message, "reasoning_content"):
    print(getattr(message, "reasoning_content"))
print(message.content)
```

**3. 工具声明"加了但没用"。** 动态声明必须携带完整的三个字段；只写 `name` 不写 `parameters`，模型无法构造合法参数。另一种常见情况是把声明放在了 `content` 字段里当文本，那对模型不可见。

**4. 命中率突然归零。** 检查是否在历史中间做过插入或删改，或者每轮重新排序了工具定义。工具定义顺序变化同样算改动。

**5. 同一个工具被反复调用。** 通常是工具返回值不足以让模型判断"已经完成"，或者错误信息被吞掉只返回一句"失败"。官方另有一篇专门讲重复调用问题的文档，处理方向是让工具结果可判别，而不是靠提示词要求"不要重复"。

## 什么时候不需要这套编排

工具少于十个、且每轮都可能全部用到时，全量声明更简单，动态注入反而增加一层检索开销和一处可能出错的环节。这套结构解决的是"工具很多、单轮只需要少数"的场景。

另一个反例：把 `search_tools` 做成关键词模糊匹配而没有维护工具目录描述，检索出来的候选经常不对，模型就会反复检索。检索质量取决于你的工具描述写得是否精确，这部分不是模型能替你做的工作。

## 通过统一网关时的验证点

如果工具调用经过聚合网关，除了模型名，还要单独验证三件事：注入在 `system` 消息里的 `tools` 字段是否被透传、`tool_calls` 的 `id` 是否原样返回、`reasoning_content` 是否在响应中保留。少任何一个，回传链都会在第二轮断掉。

以可替换实例来说，若通过 4SAPI 这类统一入口调用，参数支持范围以该平台控制台当期显示为准。验证方法不变：跑一次两轮工具调用，打印第二轮发出的 `messages`，逐字段对照。

## 限制与结论

这套编排的核心不是"少发点 token"，而是把工具可见性变成可控的、按轮次决定的状态。它成立的前提有两条：一是你的工具数量确实超过了单轮需要，二是你能保证消息只追加、不中间修改。

边界也很清楚：`reasoning_content` 会持续占用上下文并计费，长会话必须配合整段裁剪或缓存策略；内部基准和模型能力不解决工具描述质量问题；动态声明不被服务端记住，意味着任何"服务端帮你记着工具"的假设都不成立。

下一步建议：拿一个真实的双轮工具调用任务，把第二轮请求体完整打印出来，检查三件事——工具声明是否在末尾、assistant message 字段是否齐全、`tool_choice` 是否已从 `required` 恢复为 `auto`。这三项对了，剩下的问题才轮到模型能力。
