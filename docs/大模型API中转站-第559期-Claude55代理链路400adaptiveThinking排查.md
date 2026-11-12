---
title: "Codex 经本地代理调用 Claude 5.5 报 400：adaptive thinking 报错排查"
category: 人工智能
tags:
  - Claude Opus 5.5
  - API排错
  - 模型代理
description: "Codex 经本地代理接入 Claude Opus/Sonnet 5.5 时，纯聊天正常但带工具的一轮必报 400 requires adaptive thinking。本文给出分层定位方法、根因（代理写出 thinking:disabled 字面量）、最小复现和可落地的处理步骤，并区分客户端、代理层与上游各自的责任。"
brand_mode: example
---

# Codex 经本地代理调用 Claude 5.5 报 400：adaptive thinking 报错排查

如果你用 Codex，接的是一个把请求转成 Anthropic Messages 格式的本地代理，又把模型映射到了 `claude-opus-5-5` 或 `claude-sonnet-5-5`，大概率会遇到这样一段报错：

```text
"thinking.type.disabled" is not supported for this model.
Use "thinking.type.adaptive" and "output_config.effort" to control thinking behavior.
```

它的典型特征是：**纯聊天能用，一旦进入工具调用就必挂**，而且从第一轮带工具开始，之后再怎么重试都一样。这不是账号问题，也不是 Key 写错，而是请求 JSON 里出现了一个新版 Claude 不接受的字面量。

本文只回答一个问题：**这条 400 是谁产生的、怎么快速确认、以及在不改上游的前提下怎么让工具调用跑起来。**

## 一、先分层：这条报错来自哪一层

链路通常是三段：

```text
Codex 客户端  →  本地代理（格式转换 + 模型映射）  →  上游（Anthropic 兼容接口）
```

报错文案里如果同时出现 `claude-opus-5-5 requires adaptive thinking` 这类**上游模型的自我描述**，说明错误是上游返回的，而不是客户端校验或代理自己抛出的。判断方法是看这条错误的来源标记：当它被包在类似 `upstream_status: HTTP 400`、带 `provider` 和 `model` 字段的外壳里返回时，就是代理把上游 400 透传给了客户端。

这一步的意义在于：**错误信息由上游给出，但触发它的请求体由代理拼装**。所以真正要改的地方在代理这一层，而不是去调整 Codex 的会话或账号。

## 二、Claude 5.5 的 thinking 规则变了

要理解这条 400，先要知道 Claude 5.5 系列对 thinking 参数的口径。按 Anthropic 官方模型概览，当前阵容的能力列是这样标注的：

| 模型 | Thinking | 默认 effort | 上下文 |
| --- | --- | --- | --- |
| Claude Fable 5.1 | Adaptive（always on） | high | 1M |
| Claude Opus 5.5 | Adaptive（always on） | medium | 1M |
| Claude Sonnet 5.5 | Adaptive | high | 1M |
| Claude Haiku 5.5 | Adaptive | medium | 1M |

来源：[Anthropic 模型概览](https://platform.claude.com/docs/en/about-claude/models/overview)（检索日期 2026-10-14）。

关键在于两个词：**Adaptive** 和 **always on**。对 Opus 5.5 来说，思考是常开的，模型不接受“把思考关掉”这个请求。官方文档也把这件事做成了可查询能力位——Models API 的 `capabilities.thinking.types.disabled` 会明确报告该模型是否接受 `thinking: {"type": "disabled"}`；当模型以 400 拒绝这个字面量时，该字段为 `false`。

与之相邻的还有一条历史包袱：旧的 `thinking: {type: "enabled", budget_tokens: N}` 在 4.6 系列已标记废弃，**Claude 4.7 及更新的模型会直接以 400 拒绝**。也就是说，围绕 thinking 参数，新版模型同时收紧了两端——既不要 `enabled`，也不要 `disabled`，只认真实的能力声明。这一点见 [Extended thinking（官方文档）](https://platform.claude.com/docs/en/build-with-claude/extended-thinking)。

所以正确处理只有两种：

- **省略** `thinking` 键；或
- 使用 `thinking.type=adaptive`，并用 `output_config.effort` 控制思考强度。

而“代理帮我把 thinking 设成 disabled 以省 token / 兼容旧逻辑”这种做法，在新版模型上必然失败。

## 三、根因：代理在三条路径上写出了 disabled 字面量

以广为使用的本地代理 CC Switch（`farion1231/cc-switch`）为例，它在把 Codex 的 Responses 请求转成 Anthropic Messages 请求时，有一个判断“这个请求能不能不思考”的逻辑：

```rust
// 已识别的“不能关思考”模型白名单
fn thinking_cannot_be_disabled(model: &str) -> bool {
    matches!(model, "fable-5" | "mythos-5")   // 注意：没有 opus-5 / sonnet-5-5
}
```

问题就出在这里。5.5 系列虽然被 `uses_adaptive_thinking()` 认出来是“会自适应思考的新模型”，却**没有**被加进上面这张“禁用心智思考”的白名单。于是当转换器需要表达“本轮不思考”时，它仍然写出旧的 `thinking: {"type": "disabled"}`，上游直接 400。

按该项目的 issue [#8047](https://github.com/farion1231/cc-switch/issues/8047) 的复现描述，会踩到 `disabled` 分支的路径至少有三条：

1. **工具历史回放**：Codex 会重放上一轮的 `function_call_output`。代理发现配对的助手回合里没有可回放的“已签名 thinking 块”，就判定“本轮不能思考”，于是发 `disabled`。这是**最常见、也几乎是必然触发**的一条，因为工具调用天然会重放历史。
2. **强制工具选择**：请求里出现 `tool_choice: "required"` 或指定具名函数时，代理同样走 `disabled` 分支。
3. **显式关闭**：`reasoning: {"effort": "none"}` 走 `explicitly_disabled` 分支，也写字面量。

第 1 条解释了一个很多人困惑的现象：**为什么我把 effort 调到 high 也没用？** 因为 `adaptive_should_think` 只需要 `reasoning.effort` 被设置就会返回 `true`，而 `effort` 只要不是 `none`，就会撞上“工具历史 → 写 disabled → 400”的链路。effort 高低与这条报错无关。

## 四、最小复现

在代理指向一个 Anthropic 兼容通道、模型为 `claude-opus-5-5` 的前提下，发一个带强制工具选择的请求即可复现：

```json
{
  "model": "claude-opus-5-5",
  "reasoning": {"effort": "high"},
  "input": "Run ping once now.",
  "max_output_tokens": 256,
  "stream": false,
  "tools": [{
    "type": "function",
    "name": "ping",
    "description": "A harmless connectivity ping.",
    "parameters": {"type": "object", "properties": {}, "additionalProperties": false}
  }],
  "tool_choice": "required"
}
```

预期（也是被上游接受）的行为是：保持自适应思考，把客户端 effort 通过 `output_config.effort` 透传，而不是为了处理无签名工具历史或强制工具选择去关掉思考。

如果不想构造 JSON，用 Codex 侧的配置也能触达同一条路径：

```toml
# ~/.codex/config.toml（示意）
model = "claude-opus-5-5"
model_reasoning_effort = "xhigh"
```

指向代理通道后，让它调用任意一个工具，第一条工具结果就会 400。

## 五、用请求形态对照表定位

真正省时间的做法，是把“请求长什么样”和“上游返回什么”做成一张对照表。下面是相关 issue / PR 正文里记录的实测结果，可以当成排查模板：

| 请求形态 | Opus 5.5 | Sonnet 5.5 |
| --- | --- | --- |
| `thinking: disabled` | 400 requires adaptive thinking | 400 requires adaptive thinking |
| 省略 `thinking` | 200 | 200 |
| 省略 `thinking` + web_search、`tool_choice: auto` | 200，返回搜索 | 200 |
| 省略 `thinking` + 强制 `tool_choice` | 400 不支持强制选择 | 400 不支持强制选择 |
| `thinking: disabled` + 强制 `tool_choice` | 400 requires adaptive thinking | — |

这张表给出两个很有用的判断：

- **只把 `thinking` 删掉还不够**。强制 `tool_choice` 会被这几代模型独立拒绝（`tool_choice: type "tool" and "any" are not supported`），所以“改一处、重试一次”的补丁只能救回一部分请求。这正是 #8048 与 #7692 的分歧点之一。
- **普通聊天与带工具请求走的是不同分支**，这解释了“为什么聊得好好的，一调工具就崩”。

## 六、怎么处理

### 方案 A：升级到含修复的版本

先确认你用的代理版本。以 CC Switch 为例，截至本文检索日期（2026-10-14），三个直接相关的修复都还**没有合并且未进入正式版**：

- [#7692](https://github.com/farion1231/cc-switch/pull/7692)：在三个写 `disabled` 的位置改成“省略 `thinking` 键”，最小改动，但强制 `tool_choice` 路径作者自述未做上游验证。
- [#7768](https://github.com/farion1231/cc-switch/pull/7768)：新增一个“按上游报错驱动”的整流器，命中 `requires adaptive thinking` / `forced tool_choice ... not support` 后，省略 `thinking` 并把强制 `tool_choice` 放宽为 `auto`，对同一供应商重试一次；默认开启，可在“设置 → 高级 → 整流器”关闭。
- [#8048](https://github.com/farion1231/cc-switch/pull/8048)：针对 Codex + Opus 5.5，保留 adaptive 思考并按 effort 映射，同时显式处理强制工具选择，且**记录一条降级告警**说明这是 best-effort 而非严格 `required` 语义。

注意：**这些修复与该应用既有的“签名整流器 / 预算整流器”不是一回事**。旧整流器处理的是签名非法、思考预算不足之类的问题，并不认识 `requires adaptive thinking` 文案，所以把它打开并不能绕过本条报错。

### 方案 B：自己打一行补丁（要理解它的边界）

一个常见但**不完整**的改法，是把白名单扩成：

```rust
matches!(model, "fable-5" | "mythos-5" | "sonnet-5" | "opus-5")
```

它能救回“工具历史”那条主路径。但要清楚两点：

1. 如果该标志同时守卫着“直接返回本地错误”的路径，那么把它加进去，只会把**上游 400 变成代理层报错**，并没有解决问题。#7692 的说明指出，正确做法不是往这张“行为白名单”里塞，而是新增一个“**线格式**判别”——“这个 API 会不会 400 拒绝 `disabled` 字面量”，然后在这些位置改成省略键。
2. 强制 `tool_choice` 那条路径仍然存在，要彻底解决需要 #8048 的降级逻辑，并接受“具名/required 语义会被放宽”这一取舍。

### 方案 C：临时绕过（不改代码）

- 把该供应商的模型映射**别指向 5.5 系列**，改用允许 `disabled` 的 4.x 型号；
- 或暂不使用会在请求里写 `disabled` / 强制工具选择的客户端路径，等修复合并后升级。

## 七、把实例换成你自己的上游

上面讨论的“代理层 vs 上游层”分层，与上游具体是哪家**无关**。无论上游直接是官方 Anthropic 兼容端点，还是像 4SAPI 这类可替换的兼容实例，只要模型名落到 `claude-opus-5-5` / `claude-sonnet-5-5`，转换器写出的 `thinking: {"type":"disabled"}` 都会被同一套规则拒绝。换句话说，这条 400 是**请求体形态**的问题，不是供应商的问题——换一家上游、只要模型映射不变，症状不变。

反过来，这也说明一个通用排查原则：遇到“客户端报错、但看起来像上游拒绝”的 400，先把**代理写出的请求体**抓下来对比，而不是先怀疑上游限流或额度。

## 八、通用排查清单

这类“代理转换层 400”可以按下面顺序收窄：

1. **看错误属于哪一层**：有没有 `upstream_status`、有没有上游模型名自述。
2. **抓请求体**：在代理前抓一份，确认是否有 `thinking`、`tool_choice`、`output_config` 三者。
3. **二分定位**：先把请求缩到“纯文本”，再逐步加回工具、加回历史、加回强制选择，看哪一步开始 400。
4. **对照能力位**：能用 Models API 就查 `capabilities.thinking.types.disabled`；不能查就按上面的请求形态表手工验证。
5. **区分“行为”与“线格式”**：模型“能否不思考”是行为；API“是否接受 `disabled` 字面量”是线格式。二者不能混为一张白名单。
6. **确认修复是否已进你用的版本**：PR 处于 open 状态 ≠ 正式版已含修复。

## 九、限制与结论

- 本文引用的请求形态与实测结果来自项目 issue / PR 正文（检索日期 2026-10-14），**相关修复在检索时仍为 open**，请以你安装版本的实际行为为准。
- 强制工具选择的“降级为 auto”是**best-effort**，无法提供严格的 `required` / 具名保证；如果你的工作流依赖强约束工具调用，需要自行评估这一取舍。
- 官方口径才是准绳：新版 Claude 上应**省略 `thinking` 或使用 `adaptive` + `output_config.effort`**，不应发送 `disabled`（也不应发送旧式 `enabled` + `budget_tokens`）。

一句话总结：**这条 400 不是把它当成“上游拒绝”去绕，而是要把代理写出的请求体改对。** 下一步可做的技术验证，是抓一份自己的代理前请求，按第五节的对照表逐行核对 `thinking` / `tool_choice` 两个字段。
