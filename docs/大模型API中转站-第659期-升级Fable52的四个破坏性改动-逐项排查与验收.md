---
title: "升级到 Fable 5.2 的四个破坏性改动：逐项排查与验收"
category: 人工智能
tags:
  - API 迁移
  - 错误排查
  - Claude Fable 5.2
description: "把模型名从 Fable 5.1 或 Opus 5.5 换成 Fable 5.2，有四处改动会破坏现有集成，其中三处直接返回 400，另有三条同类限制和若干不报错的静默变化。本文逐项给出触发条件、错误特征、修法和一套上线前的回放脚本，适合维护 Agent 循环与多轮会话代码的开发者。"
brand_mode: none
---

# 升级到 Fable 5.2 的四个破坏性改动：逐项排查与验收

把 `model` 字段从 `claude-fable-5-1` 改成 `claude-fable-5-2`，请求能发出去，也能拿到回复。然后你在某个多轮会话里改了一次系统提示，或者在工具列表里加了一个工具，请求开始返回 400，错误信息说"这个块绑定了另一个会话"。

这类问题不会在单轮测试里出现，也不会在压测里出现，只在真实的多轮 Agent 流程里出现。Fable 5.2 的迁移文档列了四处破坏性改动——其中三处直接返回 400，另一处会静默丢掉思考块——另有三条同类限制和若干条"不报错但结果变了"的行为差异。

本文按"现象—原因—修法—怎么验证"过一遍，最后给一段可以改的回放脚本。价格与准入不在这里讨论，见本期另外两篇。

文中错误行为、参数名与 beta 头以 2026-11-12 检索到的 Claude 平台文档为准。

## 一、先分成两类：会报错的，和不会报错的

迁移时最危险的不是 400——它会立刻失败、出现在日志里。花时间的是那些请求成功、结果略差、账单略微微妙的改动。

| 类型 | 项目 | 表现 |
| --- | --- | --- |
| 硬失败 | 前缀检查不通过 | 400，`The block is bound to a different conversation` |
| 硬失败 | `computer_20251124` 工具 | 400 `invalid_request_error`（Amazon Bedrock 除外） |
| 硬失败 | 强制工具调用或 `thinking` 显式开关 | `tool_choice` 为 `any`/具名 `tool`、`thinking` 为 `enabled`/`disabled` 时 400 |
| 硬失败 | 采样参数、ZDR 工作区 | 传 `temperature`/`top_p`/`top_k` 会 400；ZDR 下每个请求都 400 |
| 静默变化 | 默认 effort 变高、思考占满 `max_tokens`、跨模型丢思考块、"不要思考"类提示失效 | 请求成功，但 token、延迟、成本都在变 |

## 二、前缀检查：这一版对每个账号都生效

这是四条破坏性变更里影响面最大的一条。

Fable 5.2 的每个思考块，只在"它产生时前面的内容没被改过"时有效。参与比较的是：顶层 `system`、`tools` 集合、该块之前的每一条 `message`。`effort`、`max_tokens`、`tool_choice`、`metadata` 不在比较范围内，`cache_control` 标记也不算改动。

区别在执行范围。Fable 5.1 只对 2026 年 8 月 31 日之后创建的账号强制执行；Fable 5.2 对所有账号强制执行。老账号上"一直没出过事"的代码，换过来可能立刻开始报错。

错误以 `messages.N.content.0: Invalid signature in thinking block` 开头，说明该块绑定了另一个会话，末尾建议改用 `drop_block`，并提示需要 `thinking-binding-controls-2026-08-01` 这个 beta 头。原样重发没有用，它以同样方式失败，不是重试能解决的问题。

哪些操作算"改"？官方对照表里最常见的是：

- 每轮重新渲染系统提示（拼日期、拼模式开关、重新读项目说明、首轮之后才连上的 MCP 服务器）；
- 在第一条用户消息里重渲染上下文，或裁剪、重编码历史的 `tool_result` 与图片；
- 增删改 `tools` 条目，或从历史中间删掉一个思考块、却保留它后面的块。

相反，只在末尾追加消息、改 `effort`、增删 `cache_control`、让 API 做服务端上下文裁剪，都不影响有效性。

## 三、两种修法，选一种

**长期修法是让历史只追加。** 代码上三句话：`system` 与 `tools` 在会话里不变；`messages` 只往后加；助手轮次内容原样回传，包括 `thinking` 为空的块。丢掉未知块或省略空 `thinking` 字段都算改动前缀，都会 400。

需要在会话中途改系统提示、改工具、改 effort 时，官方给的替代路径是：用会话中的 `system` 消息而不是重写顶层提示；用 `tool_addition` / `tool_removal` 而不是改 `tools`；用 per-message `output_config` 改 effort，这样还能保住缓存。

**临时兜底是丢块。** 带 beta 头，把行为设成 `drop_block`，请求会成功，丢掉的块列在响应的 `input_transformations` 里：

```
{
  "model": "claude-fable-5-2",
  "max_tokens": 16000,
  "thinking": {
    "type": "adaptive",
    "block_binding": { "prefix_mismatch_behavior": "drop_block" }
  },
  "messages": [ ... ]
}
```

配套请求头是 `anthropic-beta: thinking-binding-controls-2026-08-01`，`anthropic-version` 用 `2026-10-21`。

兜底有代价：丢弃的块不算输入费，但模型失去那段推理，输出 token 可能明显增加，尤其每轮都丢的时候。所以官方的定位是过渡手段。另外，Message Batches API 里没设这个字段的条目不会 400，而是直接丢弃——批量任务里更要注意。

## 四、思考块只属于产出它的模型

每个思考块记录了是哪个模型产生的，模型只读自己的块和一份固定名单上的其他模型的块。对 Fable 5.2 来说：

| 方向 | 结果 |
| --- | --- |
| 从 Fable 5.1 / Opus 5.5 / Sonnet 5.5 / Haiku 5.5 切到 Fable 5.2 | 能读，之前的推理保留 |
| Fable 5.2 与 Mythos 5.2 之间互切 | 能读 |
| 从 Fable 5.2 切到其他任何模型 | 读不到，块被丢弃，请求成功且不计费 |
| 跨组织重放历史 | 读不到，块被丢弃 |

第三条对做路由与回退的团队影响最大：会话一旦从 Fable 5.2 挪到 Opus 5.5 或 Fable 5.1，后面那一轮就没有推理可用，成本与延迟都会抬升。带 beta 头时丢弃原因写成 `model_binding_mismatch`，可以据此监控。

第四条容易被忽略：思考块与产出它的账号绑定。把线上会话存档导到另一个账号回放，看起来正常，实际上模型拿到的是没有推理的历史。回放测试要尽量用产出这些会话的账号。

## 五、计算机使用工具：换成 toolset

Fable 5.2 支持通过 `computer_toolset_20260801` 使用计算机操作。旧写法 `computer_20251124` 除 Amazon Bedrock 外，在所有平台都返回 400。Bedrock 上还能用旧工具，但要在 InvokeModel 请求体的 `anthropic_beta` 数组里传 `computer-use-2025-11-24`。

这不只是改类型名：toolset 形态下不再需要那个 beta 头，条目里也不写 `name` 和尺寸，Agent 循环里解析结果的代码通常要跟着改。这部分建议单独排期。

## 六、advisor 必须同级或更强

用 advisor 工具时，advisor 能力不能低于执行器。Fable 5.2 作为执行器只接受 Fable 5.2 或 Mythos 5.2，传其他模型都返回 400。

这里有个连环坑：改 advisor 的 `model` 属于修改 `tools`，会让之前所有思考块失效。正确顺序是在新会话里改，或者在改动时临时用 `drop_block` 兜过去。反向宽松得多：只把 advisor 换成 Fable 5.2、执行器不变时，任何执行器都接受它。

另外，Fable 5.2 作为 advisor 返回的建议是加密的，要读这段文本得走官方的结果变体路径。

## 七、三条同类的 400，外加一条 CMEK 限制

这三条不算"破坏性变更"，因为从 Fable 5.1 或 Opus 5.5 过来本来就是这样，但从更早的模型迁过来会踩：

1. **`thinking` 不能显式开关。** `{"type": "enabled"}` 加 `budget_tokens`、或 `{"type": "disabled"}` 都是 400。要么省略，要么传 `{"type": "adaptive"}`，深度用 effort 控制。
2. **强制工具调用不支持。** `tool_choice` 为 `any` 或具名 `tool` 都是 400。要结构化输出就改用 `auto` 加 `strict: true`，或用 JSON outputs。
3. **采样参数要删掉。** `temperature`、`top_p`、`top_k` 不要传。

第 2 条还有个限定：客户自管密钥（CMEK）组织里，结构化输出对 Fable 5.2 不可用，`strict: true` 与 `output_config.format` 都不能用，只能在代码里校验。这类组织若要升级，得先决定是改代码校验还是留一个非 CMEK 工作区。

## 八、不报错的那几条，往往更花时间

- **默认 effort 变了。** Opus 5.5 默认 medium，Fable 5.2 默认 high。省略 `effort` 的请求静默高一档，成功、变慢、变贵。
- **`max_tokens` 可能不够。** 思考 token 计入上限。"不要思考，直接回答"在 medium 及以上已无效；上限设小了会以 `stop_reason: "max_tokens"` 停在没有文本的地方。
- **高 effort 下会自己开子代理。** 在 `xhigh` 和 `max` 档，官方观察到模型用子代理复查自己的工作，Fable 5.1 不这么做。子代理可能占一次会话成本的大部分。
- **风格约束被严格执行。** "每句不超过 20 字"这类硬约束会让输出变得断续。

这些变化不会报错，只会让评估集上的分数和账单同时漂移。升级评估必须包含 cost 与 latency，不能只看正确率。

## 九、上线前的回放脚本

下面这段脚本按官方文档的请求结构写成，做两件事：跑一轮双轮会话，统计有没有块因前缀不符被丢弃；再故意改一次 `system`，确认代码能正确识别 400。本次没有在真实密钥上执行，参数名请先在最小会话里确认再接入 CI。

```python
import anthropic

BETA = "thinking-binding-controls-2026-08-01"
client = anthropic.Anthropic()  # 读 ANTHROPIC_API_KEY

def turn(messages, system=None, behavior="drop_block"):
    kwargs = dict(
        model="claude-fable-5-2",
        max_tokens=16000,
        thinking={"type": "adaptive",
                  "block_binding": {"prefix_mismatch_behavior": behavior}},
        messages=messages,
        betas=[BETA],
    )
    if system:
        kwargs["system"] = system
    return client.beta.messages.create(**kwargs)

history = [{"role": "user", "content": "算一下 1071 和 462 的最大公约数，并说明步骤。"}]
first = turn(history)
history.append({"role": "assistant", "content": first.content})   # 原样回传，含思考块
history.append({"role": "user", "content": "用同样的方法算 1547 和 595。"})
second = turn(history)

dropped = [t for t in (second.input_transformations or [])
           if t.get("reason") == "prefix_binding_mismatch"]
print("第 2 轮丢弃块数:", len(dropped), "期望 0")

try:
    turn(history, system="You are a terse assistant.", behavior="error")
    print("预期失败但没有失败：行为参数是不是没生效")
except anthropic.BadRequestError as e:
    print("按预期返回 400:", str(e)[:120])
```

期望结果是：第一段打印 `0`，说明多轮里前缀保持住了；第二段抛 400，说明前缀检查确实生效，错误处理路径能被覆盖。注意两段用的行为不同——第一段设 `drop_block` 才能拿到丢弃计数，第二段必须设 `error` 才会抛错；只用 `drop_block` 的话，改过的前缀会被静默丢弃，你什么都看不到。若第一段大于 0，说明代码里还有改动历史的地方——`input_transformations` 的每一条都带 `path`，按路径往前 diff `system`、`tools` 和 `messages` 就能定位。

## 十、上线前清单

1. 全局搜索 `computer_20251124`，换成 toolset；Bedrock 分支单独处理。
2. 搜索 `tool_choice` 的 `any` 与具名用法，换成 `auto` 加 `strict` 或 JSON outputs；CMEK 工作区改用代码校验。
3. 搜索 `temperature`、`top_p`、`top_k`、`budget_tokens`、`thinking: disabled`，全部删除或改写。
4. 确认 advisor 的 `model` 与执行器匹配；确认工作区不是 ZDR，或已打开 30 天留存。
5. 用上面的脚本回放现有会话，把 `prefix_binding_mismatch`、`model_binding_mismatch` 接进监控。
6. 显式设置 `effort`，从 high 开始做一次 effort sweep。
7. 用自己的评估集重跑，别沿用旧评分基线。

## 十一、限制与结论

这次升级的难点不在模型本身，而在"历史怎么拼"：前缀检查对所有账号生效，思考块只归产出它的模型和账号。两条加在一起，任何"重写历史"的实现都会暴露。

值得先做的是查会话存储：消息数组若由你拼，先确认它是只追加的；若由 Claude Code、Agent SDK、Managed Agents 拼，官方说行为不变，可以直接换模型名再跑评估。

至于暂时过不去的部分，`drop_block` 能让请求继续跑，但它改变的是"能不能跑"，不是"跑得对不对"。把它当成灰度期间的开关，配合 `input_transformations` 计数做告警，比写进生产配置更合适。

## 参考来源

| 来源 | 链接 | 检索日期 |
| --- | --- | --- |
| Claude Platform Docs：What's new | https://platform.claude.com/docs/en/models/fable-5-2/whats-new-fable-5-2 | 2026-11-12 |
| Claude Platform Docs：迁移指南 | https://platform.claude.com/docs/en/models/fable-5-2/migration-guide | 2026-11-12 |
| Claude Platform Docs：Preserved thinking | https://platform.claude.com/docs/en/build-with-claude/preserved-thinking | 2026-11-12 |
| Claude Platform Docs：Fable 5.2 模型页 | https://platform.claude.com/docs/en/models/fable-5-2/overview | 2026-11-12 |
| Claude Platform Docs：Computer use tool | https://platform.claude.com/docs/en/agents-and-tools/tool-use/computer-use-tool | 2026-11-12 |
