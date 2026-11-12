---
title: "Gemini Decisions API 该不该用：三类问题与一次多问的请求结构"
category: 人工智能
tags:
  - Gemini API
  - 模型选型
  - 结构化输出
description: "客服分流这类需求，让模型生成 JSON 再解析既慢又脆。本文说明 Gemini Decisions API 的打分式机制、yes/no 与选择题与评分三种问题类型、共享输入一次计费的请求结构，以及六条决定它不能用的硬限制，适合正在维护分流、规则检查和多候选打分的开发者。"
brand_mode: none
---

# Gemini Decisions API 该不该用：三类问题与一次多问的请求结构

一个常见的实现是让模型输出 JSON，再用解析器取字段："这条工单属于哪个队列？是否在申请退款？有多紧急？"。响应回来了，字段名偶尔变成别的东西，数字偶尔被写成字符串，JSON 偶尔缺一个闭合括号。为这几种情况加的重试和兜底，最后比业务判断本身还长。

2026 年 11 月 11 日，Google 发布 Gemini Decisions API，把这类"从固定选项里挑一个"的任务换成了另一种形式：不给它生成文本的机会，而是让它在一次前向计算里给候选答案打分，直接返回概率。模型是 `gemini-decision-1`，通过 Gemini API 的 v1beta 提供。

本文回答三个问题：它怎么工作、请求怎么写、哪些情况下它不该出现。成本与阈值放在后面两篇。

价格、参数与限制以 2026-11-12 检索到的官方页面为准；下面每个数字都可以在文末的表格里找到出处。

## 一、机制：不打字，直接打分

官方文档把它和生成式模型的分工写得很直接：生成模型产出给人读的文本，而代码需要一个用来分支的判断时，"生成再解析"会带来延迟和脆弱的解析逻辑。

Decisions API 的输入是两样东西：

- **共享输入**（`input`）：一段文本，或图片、音频、视频、PDF 组成的内容块数组；
- **问题表**（`questions`）：1 到 1,000 个有类型的判断。

输出是按问题键返回的答案，每个答案带概率和置信度。它不产生任何文本 token，响应里的 `total_output_tokens` 恒为 0。官方对延迟的描述是端到端 150–250 ms、模型计算中位数 50–100 ms，并且在 16K token 以内的共享输入上，从 1 个问题加到 100 个问题，延迟基本持平。

这三个数字解释了它的适用面：一次请求里问题越多，摊到每个判断上的延迟越薄；它完全不适合"给我一段解释"。

## 二、三种问题类型

所有问题必须是这三种之一，返回结构各不相同。

| 类型 | 用途 | 返回 | 关键字段 |
| --- | --- | --- | --- |
| `noul` | 判断某个陈述是否成立 | 0.01–0.99 的"成立"概率 | `noul`、`confidence` |
| `choice` | 从固定选项里选一个 | 胜出项 + 全部选项的归一化概率 | `choice`、`probabilities`、`confidence` |
| `score` | 按有序等级给分 | 概率加权期望分 + 每级概率 | `score`、`probabilities`、`legend`、`confidence` |

几点说明：

- `noul` 是 yes/no 的简写形式，值域 0.01 到 0.99，不含 0 和 1。
- `choice` 支持 2 到 255 个选项；`score` 支持 2 到 255 个等级，但官方建议大多数评分表控制在 2 到 10 级。
- `score` 返回的是等级索引的加权平均，因此会落在两级之间。官方示例给出 1.12 分、各级概率 0.03 / 0.83 / 0.15，这意味着"大概率在第 1 级，但有 15% 可能落到第 2 级"。
- `confidence` 的定义是"最高答案正确的估计概率"，对 yes/no 就是 `max(noul, 1 - noul)`；对 choice 和 score 就是最高项的概率。

先记住一件事：`confidence` 不等于准确率，它是模型对自己给出那个答案的把握。这个区别在设阈值时很关键，后面一篇专门讲。

## 三、一次请求的结构

请求体只有三个必填字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `model` | string | 固定为 `gemini-decision-1` |
| `input` | string / Content / Content[] | 所有问题共享的上下文 |
| `questions` | 键到问题的映射 | 1–1,000 个；键名 `[A-Za-z0-9_-]+`、最长 64 字符 |

每个问题有 `type` 和 `instructions` 两个必填字段，`criteria` 视类型而定：`choice` 与 `score` 必填，`noul` 可选（用来分别描述"成立"和"不成立"各自意味着什么）。

REST 调用的端点是 `POST https://generativelanguage.googleapis.com/v1beta/decisions`，鉴权用 `x-goog-api-key`。SDK 侧需要 `google-genai` 3.2 或更高，JavaScript 用 `@google/genai` 最新版。

下面是官方一个多模态示例的简化版：把一张现场照片和一句备注送进去，同时问三类问题。

```python
from google import genai

client = genai.Client(http_options={"api_version": "v1beta"})
result = client.decisions.create(
    model="gemini-decision-1",
    input=[
        {"type": "image",
         "uri": "https://storage.googleapis.com/generativeai-downloads/images/spill.jpg",
         "mime_type": "image/jpeg"},
        {"type": "text", "text": "Breakroom counter check before the 9 AM client meeting."},
    ],
    questions={
        "has_spill": {
            "type": "noul",
            "instructions": "Is there an uncleaned liquid spill on the counter?",
        },
        "next_action": {
            "type": "choice",
            "instructions": "Which cleanup action fits the scene?",
            "criteria": {
                "wipe_with_cloth": "Small counter spill with a cloth already nearby.",
                "dispatch_janitorial": "Large floor hazard, broken glass, or active leak.",
                "ready_for_meeting": "Counter is clean and ready.",
            },
        },
        "severity": {
            "type": "score",
            "instructions": "Rate the severity of the counter mess.",
            "criteria": [
                "Clean and ready.",
                "Minor localized spill.",
                "Major spill or safety hazard.",
            ],
        },
    },
)
print(result.answers["has_spill"].noul)
```

注意 `criteria` 的两种形状：`choice` 是"选项键 → 描述"的字典，`score` 是"从低到高"的描述数组，下标即等级编号。两者长得像，混写会报错或得到无意义的分。

## 四、共享输入只算一次

这一层对账单影响很大：一次请求里共享输入只计费一次，不管问多少个问题；每个问题只增加自己的 `instructions`、`criteria` 与格式化 token。

"一条长文本 + 20 个独立规则检查"在结构上划算：20 个问题的附加 token 通常远小于重发 20 次长文本。反过来，"20 条互不相关的短文本 + 每条一个检查"塞进一个请求没有意义，共享输入并没有被复用。

另一半是隔离性：问题彼此不可见，增删一个问题不会改变其他答案。"加一条规则"因此是纯粹的加法。代价是不能靠"第一个问题决定第二个"串逻辑——这种依赖得写在代码里。

## 五、与结构化输出的取舍

官方给了一张对照表，可以直接拿来判断：

| 你的需求 | 用 Decisions API | 用 Interactions API（生成式） |
| --- | --- | --- |
| 一个是/否概率、固定集合里选一个、固定刻度上打分 | 是 | 否 |
| 可以设阈值的校准概率 | 是 | 否 |
| 对同一输入做很多互不相关的检查 | 是 | 否 |
| 低延迟、零输出 token 成本 | 是 | 否 |
| 生成文本、摘要、解释 | 否 | 是 |
| 按你自己的 JSON schema 抽取对象 | 否 | 是（结构化输出） |
| 带参数的工具调用 | 否 | 是 |
| 多步推理 | 否 | 是 |

用一句话概括：**输出是"有限集合里的一个元素"时用它；输出是"任意文本或有 schema 的对象"时不要用它。**

这条界线经常被含糊掉，因为两者都能做分类。差别在于出错方式：生成式模型给你一个字符串，你得验证它是否在允许的集合里；打分式模型永远不会给出集合外的答案，但可能给出一个"集合内但错"的答案，并把不确定性藏在概率里。

## 六、六条让它不能用的限制

下面这些是硬限制，不撞上就没事，撞上就是架构层面的返工。

1. **只支持 `gemini-decision-1`，只在 v1beta 提供。** 模型卡上 Batch、缓存、代码执行、函数调用、结构化输出、思考、Live API 全部标注不支持。
2. **不支持上下文缓存、Batch API、Flex 与 Priority 推理。** 重复输入每次按全价计，对"高频调用、提示词固定"的场景会影响成本结论。
3. **问题只能写文本。** 媒体必须放在共享输入里；想"针对这张图和那张图分别问"，只能拆成两次请求。
4. **输入不支持 YouTube 链接和 `gs://`。** 媒体要么走 Files API 的 URI、公共或签名 HTTPS 链接（单个外部 URL 上限 15 MB），要么内联 base64（单请求 100 MB，PDF 50 MB）。
5. **不支持对话历史、思考块、工具调用块作为输入。** 多轮上下文要自己序列化成文本塞进 `input`。
6. **没有"最大 1,000 问题"之外的扩容手段。** 一个请求最多 1,000 个问题，上下文合计 262,144 token；问题再多就得分批请求，而分批又没有 Batch API。

第 2 条和第 6 条合在一起，是这套 API 目前最难绕的约束：它优化的是"单次、低延迟、多判断"，不是"海量、离线、批量打分"。

## 七、容易用错的几处

- **问题键被忽略。** 模型只看到 `instructions`，看不到键名，所以键名可以随便起，但问题本身必须写完整。
- **相邻文本块没有分隔符。** 多个 `text` 块直接拼接，不加空格或换行；会粘成一个词的话，得自己补 `\n`。
- **结构化数据要序列化。** JSON 先转成字符串再当 `text` 传，在 `instructions` 里用键路径引用（如"看 `order.age_days`"）。
- **选项键按字母序排序。** 返回的胜出项是键名，不受影响；但"兜底选项"要写成 `other` 这类位置无关的键，而不是 `none_of_the_above`。
- **安全问题返回 400。** 被安全过滤拦下时不给 `answers`，而是 HTTP 400 加 `error.code`（`safety`、`prohibited_content`、`blocklist`、`image_safety`、`content_blocked` 之一），必须在代码里显式处理。

最后一条还牵出一个未确认的问题：被拦截的请求是否计费，定价页没写。这需要用自己的账号对账，本文不把它当已知事实。

## 八、它适合放在流程的哪个位置

官方给的定位是"低延迟控制面"：用定义选项的方式上线功能，而不是训练分类器或等生成式调用返回。发布稿举的内部用例就是样本——毫秒级判断意图、把日常查询路由到便宜的模型档位、把深度推理留给复杂任务，以及在搜索的 MCP 工具前做对抗行为检测。

把这些归纳成一条判断标准：

- 结果要**立刻**用于分支（路由、放行、升级、排序）的，值得换；
- 结果要**给人看**或要能解释"为什么这么判"的，留在生成式模型，或两者并用：打分式做初筛，低置信度时才调用会解释的模型。

后半句其实是这套 API 最有价值的用法：概率让"哪些请求需要多想"变成一个可以量化的问题。具体怎么定这个界线，是下一篇的主题。

## 九、限制与结论

Gemini Decisions API 解决的问题很具体：把"生成文本再解析"换成"给候选答案打分"。代价是它只能做有限集合的判断，不能解释、不能抽取、不能多步推理，也不支持批量与缓存。

三条判断：

- 流程里要解析 JSON 才能分支的地方，是它的候选替换点；要读文本才能决策的地方不是。
- 一次请求塞 20 到 100 个独立检查通常划算：共享输入只计一次，延迟基本不随问题数增长。
- 需求里出现"海量离线批量"或"提示词高度重复"时先算成本：它没有 Batch，也没有缓存。

至于概率能不能直接当决策用、阈值怎么定，取决于你自己的流量分布，而不是文档里的默认值。

## 参考来源

| 来源 | 链接 | 检索日期 |
| --- | --- | --- |
| Google 官方发布稿（机制、延迟、内部用例、生态集成） | https://blog.google/innovation-and-ai/technology/developers-tools/gemini-decisions-api/ | 2026-11-12 |
| Decisions API 指南（三种问题类型、请求与响应、输入格式、限制、最佳实践） | https://ai.google.dev/gemini-api/docs/decisions-api | 2026-11-12 |
| Gemini Decision 模型卡（上限、能力支持矩阵、更新月份） | https://ai.google.dev/gemini-api/docs/models/gemini-decision-1 | 2026-11-12 |
| Pricing（Gemini Decision 段：$0.10 输入、0 输出、无缓存） | https://ai.google.dev/gemini-api/docs/pricing#gemini-decision | 2026-11-12 |
| Rate limits（Gemini Decision 四档限速） | https://ai.google.dev/gemini-api/docs/rate-limits#gemini-decision | 2026-11-12 |
