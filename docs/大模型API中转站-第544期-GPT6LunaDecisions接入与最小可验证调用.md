---
title: "GPT-6 Luna Decisions 怎么接入：请求结构、三类问题与排错"
category: 人工智能
tags:
  - GPT-6 Luna Decisions
  - API接入
  - 错误排查
description: "Decisions 接口用 input 加 questions 提交判定，返回 answers 数组。本文按官方文档给出三类问题的最小调用、响应解析、图像输入、批量调用与常见错误排查。"
brand_mode: none
---

# GPT-6 Luna Decisions 怎么接入：请求结构、三类问题与排错

第一次调 Decisions 接口的人，常犯的错误不是“请求发不出去”，而是**按对话接口的习惯写请求**：把问题写进输入文本、期待返回一段话、然后发现解析不了。这个接口的结构和常规生成接口不一样。

以 **2026 年 10 月 14 日**核验的官方文档为准，一次 Decisions 请求由三部分组成：`model`、`input`、`questions`；响应里是 `answers` 数组。当前只有 `gpt-6-luna` 一个模型可用，端点固定在 `POST /v1/decisions`。[Decisions 文档](https://developers.openai.com/api/docs/guides/decisions)

本文只解决一个问题：**如何把 Decisions 接口跑通，并确认拿到的是正确的结果。**

## 前置条件

- **端点**：`POST /v1/decisions`。
- **模型标识**：`gpt-6-luna`（官方文档明确当前仅支持这一个）。
- **凭证**：有效的 API key，且账号已开通该能力。
- **SDK 版本**：想要用 SDK 调用，需满足官方给出的最低版本。

```text
Python      3.26.0 或更高
JavaScript  7.30.0 或更高
Go          3.73.0 或更高
Ruby        0.101.0 或更高
Java        4.78.0 或更高
```

- **状态认知**：官方说明处于公开测试阶段，并计划在接下来几周正式可用，字段与限制可能调整。

## 用环境变量组织配置

```bash
export OPENAI_API_KEY="sk-..."
export DECISION_MODEL="gpt-6-luna"
```

把模型标识抽出来，便于官方新增模型或调整标识时只改一处。

## 最小请求：predicate

官方示例是一个图片破损检查。下面用同一结构写一个纯文本版本，字段含义与官方一致：

```bash
curl https://api.openai.com/v1/decisions \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -H "Content-Type: application/json" \
  --data-binary @- <<JSON
{
  "model": "gpt-6-luna",
  "input": "这段客服回复是否包含对退款时效的具体承诺？",
  "questions": [{
    "type": "predicate",
    "name": "refund_commitment",
    "instructions": "判断文本中是否明确承诺了退款到账时间。仅凭暗示不算成立。"
  }]
}
JSON
```

三个字段的作用：

- `input`：共享证据。
- `questions[].type`：问题类型。
- `questions[].name`：这个问题的标识，响应会回传。

## 三类问题的写法

**predicate**（判断成立与否）：

```json
{
  "type": "predicate",
  "name": "has_visible_damage",
  "instructions": "图片中的商品是否有可见破损，例如裂纹、撕裂、凹陷？忽略阴影与包装破损。"
}
```

返回 `probability`，0 到 1 的估计。

**choice**（从固定选项选一个）：

```json
{
  "type": "choice",
  "name": "department",
  "instructions": "这条工单应转给哪个部门？",
  "choices": ["物流", "退款", "发票", "其他"]
}
```

返回 `choice`，即你提供的某个取值，同时给出各选项的概率分布。

**score**（按有序档位打分）：

```json
{
  "type": "score",
  "name": "severity",
  "instructions": "这条故障报告的严重程度如何？",
  "levels": ["轻微", "一般", "严重", "致命"]
}
```

返回 `score`，是档位序号的**概率加权平均**，所以可能落在两档之间。

字段名以官方文档为准；这里展示的是结构关系。关键区别是：`choice` 用于无顺序类别，`score` 用于有顺序档位。

## 图像输入

官方示例把图像与文字放进同一条用户消息，让模型看到同一份证据：

```json
{
  "model": "gpt-6-luna",
  "input": [{
    "role": "user",
    "content": [
      {"type": "input_text", "text": "检查这张商品照片。"},
      {"type": "input_image", "image_url": "data:image/png;base64,<BASE64>"}
    ]
  }],
  "questions": [{
    "type": "predicate",
    "name": "visible_damage",
    "instructions": "商品是否有可见破损？忽略阴影和包装破损。"
  }]
}
```

要点：

1. 图像与文字放在**同一个 input** 里，不要拆成两次请求。
2. `instructions` 里写明排除条件，否则反光、阴影容易被误判。
3. 大图会消耗更多 token，批量场景先统一尺寸。

## 解析响应

响应是 `answers` 数组，按 `name` 配对：

```python
answers = {a["name"]: a for a in response["answers"]}

prob = answers["refund_commitment"]["probability"]
route = "auto" if prob >= 0.9 else "human"
```

三个原则：

- **按 `name` 取答案**，不要按数组下标。
- **概率单独取出来**做阈值判断。
- **保留原始响应**，便于复盘。

## 常见错误与排查

| 现象 | 可能原因 | 排查方向 |
| --- | --- | --- |
| 404 | 用了 `/v1/responses` 或 `/v1/chat/completions` | 改为 `/v1/decisions` |
| 400 | 模型标识或问题字段不合法 | 用 `gpt-6-luna`，核对 `type` 与 `name` |
| 401 / 403 | key 无效或未开通 | 换 key、确认权限 |
| 429 | 触发速率限制 | 退避重试，限制并发 |
| 答案错位 | 问题缺少唯一 `name` | 为每个问题补 `name` |
| 分数异常 | 用 choice 代替 score | 有序档位改用 score |

## 批量调用的实践

判定类任务往往是批量的，几个要点：

1. **一次请求问多个问题**：共享同一份 `input`，减少重复输入成本。
2. **并发上限**：批量并发容易触发 429，需要队列与退避。
3. **结果落库**：把 `name`、答案、概率、阈值、路由都存下来。
4. **抽样复核**：按概率分层抽样，验证阈值是否合理。

## 验证清单

- 端点是 `/v1/decisions`，模型标识是 `gpt-6-luna`。
- 每个问题都有唯一 `name`，且与解析代码一致。
- 问题类型与答案空间匹配（predicate / choice / score）。
- 图像与文本在同一 input 中提交。
- 概率被显式读取并进入阈值逻辑。
- 原始响应被保留。
- 重试带退避，且重试计入成本。
- 针对公开测试阶段准备了降级方案。

## 用 SDK 调用

官方示例同时给出了多种语言的 SDK 写法。用 SDK 的好处是省去手工拼 JSON，代价是必须满足最低版本要求：

```python
from openai import OpenAI

client = OpenAI()

decision = client.decisions.create(
    model="gpt-6-luna",
    input="这段客服回复是否包含对退款时效的具体承诺？",
    questions=[{
        "type": "predicate",
        "name": "refund_commitment",
        "instructions": "判断文本中是否明确承诺了退款到账时间。",
    }],
)
```

如果项目锁定旧版本 SDK，可能出现“方法不存在”的报错。这时要么升级 SDK，要么直接用 HTTP 调用，不要混用。

## 拿到概率之后做什么

很多人第一次拿到 `probability` 不知道该怎么办。它至少有三个用途：

1. **设阈值分流**：高于上界自动通过，低于下界直接转人工。
2. **排序**：对一批样本按概率排序，优先处理最不确定的。
3. **监控质量**：观察概率分布随时间的变化，发现输入分布漂移。

需要注意的是，**概率不等于校准良好的置信度**。在把它当阈值用之前，先画一次“概率区间 vs 实际正确率”的曲线，确认高概率确实对应高正确率。

## 与结构化输出的迁移路径

如果现有系统已经在用 Responses API 的 Structured Outputs 让模型输出 JSON 判定结果，迁移到 Decisions 时建议这样走：

1. **保留原链路**：先不动生产，把 Decisions 接在旁路做对照。
2. **统一判定定义**：把原来的判定维度整理成标准的 `questions` 定义。
3. **同样本对照**：同一批样本跑两套，比较结果一致性。
4. **切换分流**：只在概率高的一段切到 Decisions，其余仍走原链路。
5. **逐步扩大**：一致性达标后再扩大自动比例。

这样迁移的风险最小，也便于随时回滚。反之，直接把整条链路切过去，一旦判定口径有出入，很难定位。

## 成本上的两个小提醒

- **共享 input 更省**：多个问题共享同一份证据，只付一次输入的钱。
- **注意分档线**：Luna 本体在输入超过 272K 后按 2 倍输入、1.5 倍输出计费，超长证据建议先拆分。

这两点在小批量测试时看不出来，量一上来就会体现在账单上。
## 限制与结论

Decisions 接口的接入难点不在发请求，而在**把问题定义写清楚并按 `name` 正确解析**。先确定用 predicate、choice 还是 score，再按官方结构组织 `input` 与 `questions`，最后用概率做自动与人工的分流，才算真正跑通。

本文的结构与字段以 2026 年 10 月 14 日的官方文档为准，属于可变事实；接口处于公开测试阶段，代码里应对响应结构做防御式解析，并为异常响应准备降级路径。
