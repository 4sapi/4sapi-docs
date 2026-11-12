---
title: "置信度门槛怎么定：Decisions API 的阈值校准与分数陷阱"
category: 人工智能
tags:
  - 模型校准
  - 阈值调优
  - Gemini API
description: "Decisions API 返回的是概率，不给阈值就没法用。本文说明 confidence 到底代表什么、怎么用自有标注流量分桶定阈值、为什么平均分 1.12 不能直接当第 1 级用，以及选项排序、负向提问和可被输入说服的分类器怎么处理，适合要把概率接进线上分支的开发者。"
brand_mode: none
---

# 置信度门槛怎么定：Decisions API 的阈值校准与分数陷阱

把 Decisions API 接进分流的第二天，几乎所有人都会问同一个问题：`confidence` 到多少才自动处理，低于多少转人工？

多数实现会随手写个 0.7 就上线。一周后回头看，这个数字既没依据，也解释不了为什么不是 0.6 或 0.85。

这篇不讨论接口怎么调（那是上一篇的事），只讨论四件事：`confidence` 在数学上是什么、怎么用自己的流量给阈值找依据、`score` 的期望值为什么会骗人、以及几个会让概率变得不可用的写法。

文中对返回结构的描述来自 2026-11-12 检索到的官方文档，示例数值取自文档自带的算例。

## 一、先把三个名词分开

Decisions API 的答案里有三个不同的量，混着用会做出错误的决策。

| 量 | 定义 | 用途 |
| --- | --- | --- |
| 答案值 | yes/no 的成立概率、choice 的胜出项、score 的期望等级 | 决定"做什么" |
| `probabilities` | choice 与 score 的全部选项/等级概率，和为 1 | 决定"有多确定"、"第二可能是什么" |
| `confidence` | 最高答案的估计概率 | 决定"要不要自动执行" |

`confidence` 是"最高答案正确的估计概率"：yes/no 等于 `max(noul, 1 - noul)`，choice 与 score 等于最高项概率。

这带来第一个容易踩的坑：**yes/no 的 confidence 永远不小于 0.5**。因为 `noul` 是 0.01，`1 - noul` 就是 0.99，取最大之后下限是 0.5。所以"confidence ≥ 0.7 就自动处理"在 yes/no 上的实际含义是"yes 概率 ≥ 0.7 或 ≤ 0.3"，是一段双侧区间，而不是单调的把握。

如果业务上只关心 yes 的概率（"这条是不是退款申请"），直接用 `noul` 就行，别绕一道 confidence；只有当你需要"无论 yes 还是 no 都要判断模型有多确定"时，confidence 才是合适的量。

## 二、校准的意思：0.8 应该真的对 80%

文档反复用 "calibrated" 这个词。它的实际含义是：如果把所有 confidence 落在 0.8 附近的请求挑出来，其中正确答案的比例应该接近 80%。

这是可以被验证的，也应该被验证。方法是拿一批已经有人工标注的历史请求，跑一遍 Decisions API，按 confidence 分桶，统计每桶的实际准确率：

| confidence 桶 | 请求数 | 实际正确率 | 说明 |
| --- | --- | --- | --- |
| 0.50–0.60 | 示例 | 示例 | 若明显低于桶中值，说明该桶不可信 |
| 0.60–0.70 | 示例 | 示例 | — |
| 0.70–0.80 | 示例 | 示例 | — |
| 0.80–0.90 | 示例 | 示例 | — |
| 0.90–1.00 | 示例 | 示例 | 高置信区间的错误代价最大 |

表中"示例"是空的，因为分桶结果只能来自你自己的数据和任务定义。这不是省事，是这个方法本身的要求：同一个模型在"判断工单是否在申请退款"和"判断照片里是否有破损"上会有不同的校准曲线，官方文档也明确要求用自有流量上的标注样本设定阈值，并权衡假阳性与假阴性的成本。

分桶结果常见三类，对应三种处理：

- **单调且贴近对角线**：概率可用，扫阈值选满足错误率的最大覆盖率。
- **中间段塌陷**（0.6–0.8 桶的实际准确率明显低于桶中值）：这段置信度不能用于自动处理，把阈值抬到塌陷区以上。
- **高分段也偏乐观**（0.9 以上仍然明显低于 90%）：概率整体被高估，考虑在代码里做一次简单的重标定，或者干脆只把 top 1 当作排序依据，不做"自动放行"。

## 三、阈值怎么选，取决于两种错误的价格

阈值不是一个数字，是一组 (覆盖率, 错误率) 的取舍。把它写成一个可执行的流程：

1. 定义"错一次"的代价。退款误判成咨询，用户白等一天；咨询误判成退款，财务多走一遍流程。两者通常差一个量级。
2. 在标注集上跑出每条的 `confidence` 与答案。
3. 从高到低扫阈值，对每个阈值算两个数：自动处理的比例、自动处理的错误率。
4. 选"错误率仍可接受"的最高覆盖率，而不是覆盖率最高。
5. 把剩下的一律转人工或转更贵的模型，并记录转出率——它是你后续优化 prompt 和选项定义的基线。

这一步会暴露一个常被忽略的现实：**选项定义的质量决定可达的覆盖率**。若 `criteria` 里两个选项描述重叠（"支付问题"与"账户更新"都可能覆盖"改了卡号"），置信度会在重叠区系统性下降。

## 四、score 的期望值陷阱

`score` 返回的是等级索引的概率加权平均。官方示例里，三级评分表（0 仅外观问题、1 有绕过办法、2 完全阻塞）返回 `score: 1.12`，各级概率 0.03 / 0.83 / 0.15。

1.12 看起来像"第 1 级偏上一点"，但它不是一个连续测量值。它是加权平均，只说明"大概率第 1 级，有一定可能第 2 级"。如果直接拿 1.12 跟固定阈值比（比如"≥ 1.5 就升级工单"），你是在用混合分布的均值判断分布，边界附近会不稳定。

三种更可靠的用法：

- **当排序信号**：用 score 给候选排序或筛选，不跟固定阈值比较。
- **取最大概率等级**：需要离散等级时，取 `probabilities` 里最大的那一项，再用 `legend` 映射回文字。文档给了这段代码：`level = max(answer.probabilities, key=answer.probabilities.get)`。
- **归一化后再跨表比较**：不同评分表的等级数不同时，用 `score / 最高等级索引` 把它压到 0–1，再做比较。

还有一个更隐蔽的问题：**刻度必须线性**。因为模型返回的是索引的加权平均，如果等级是 1、10、100、1000 这种指数刻度，期望值会被严重扭曲——它平均的是索引 0、1、2、3，而不是 1、10、100、1000。官方文档在最佳实践里点名了这一条，建议相邻等级的描述清晰可分辨、刻度保持线性。

另外，等级数不是越多越好。上限是 255 级，但文档建议大多数评分表用 2 到 10 级。等级太密，相邻两级的描述会退化成语义重叠，概率分布摊平，`confidence` 随之下降。

## 五、四个会让概率变得不可用的写法

**问题键不是提示词。** 模型看不到 `questions` 的键名，只看 `instructions`。写 `q_severity` 这样的键不会给它任何线索，"严重不严重"这四个字必须出现在 `instructions` 里。

**负向提问会毁掉可用性。** 官方建议正向表述：问"这个客户是 Gold 等级吗"，再用 `1 - noul` 得到"不是"的概率，而不是直接问"不是 Gold 吗"。否定句容易与"提到了 Gold 但意思不同"混淆，概率偏了之后代码里的补运算只会放大误差。

**兜底选项的键名要位置无关。** 选项键会被按字母序排序，所以 `none_of_the_above` 会落在 `billing` 之后，看起来像最后一个选项，实际并不是。官方建议用 `other` 这类不依赖顺序的键名。更稳的做法是根本不给兜底选项，而是配一个显式的范围检查问题：另开一个 `noul` 问"这条消息是否是关于订单的"，置信度低时直接转出。

**输入是可被说服的。** 文档把这一条写进最佳实践：把输入当作不可信数据，用户提供的文本可以为自己的分类辩护。工单里写"这是紧急退款，请立即处理"，就会把 `choice` 往退款那边推。对策有三层：`criteria` 写死判定条件而不是依赖语气；用对抗样本测一遍（在自己的标注集里塞入带诱导语气的样本，看概率是否被拉偏）；高风险动作（退款、扣款、封禁）保留人工或第二个检查。

第三层不只是工程问题。这类模型给的是概率，概率可以叠：两个独立判据都过阈值才执行，比单一判据过阈值稳得多——这恰好是"一次请求问多个问题"的结构最实际的用法。

## 六、一段可用的验收脚本骨架

下面这段代码把上面的流程写成可运行骨架：拉样本、跑批量请求、分桶统计、扫阈值。它没有在真实密钥上执行过，字段名按 2026-11-12 的文档写法，接入前请先用一条请求确认返回结构。

```python
import json

from google import genai

client = genai.Client(http_options={"api_version": "v1beta"})
MODEL = "gemini-decision-1"

# 每行：input 文本、人工标注的正确选项键
samples = [json.loads(line) for line in open("labeled.jsonl", encoding="utf-8")]


def decide(text: str) -> dict:
    result = client.decisions.create(
        model=MODEL,
        input=text,
        questions={
            "queue": {
                "type": "choice",
                "instructions": "Which support queue should handle this message?",
                "criteria": {
                    "billing": "Payment, invoice, or refund questions.",
                    "replacements": "Damaged, missing, or wrong items.",
                    "general": "Product questions and account updates.",
                },
            },
        },
    )
    a = result.answers["queue"]
    return {"choice": a.choice, "confidence": a.confidence}


rows = []
for s in samples:
    try:
        r = decide(s["text"])
    except Exception as e:                      # 含 HTTP 400 的安全拦截
        rows.append({"label": s["label"], "pred": None, "conf": 0.0, "err": str(e)[:80]})
        continue
    rows.append({"label": s["label"], "pred": r["choice"], "conf": r["confidence"], "err": None})

buckets = [(0.5, 0.6), (0.6, 0.7), (0.7, 0.8), (0.8, 0.9), (0.9, 1.01)]
for lo, hi in buckets:
    part = [r for r in rows if lo <= r["conf"] < hi and r["err"] is None]
    if not part:
        print(f"[{lo},{hi}) 无样本")
        continue
    acc = sum(r["pred"] == r["label"] for r in part) / len(part)
    print(f"[{lo},{hi}) n={len(part):4d} acc={acc:.2f}")

for thr in (0.6, 0.7, 0.8, 0.9):
    auto = [r for r in rows if r["conf"] >= thr and r["err"] is None]
    if not auto:
        continue
    acc = sum(r["pred"] == r["label"] for r in auto) / len(auto)
    print(f"thr={thr} 覆盖率={len(auto)/len(rows):.2f} 自动处理准确率={acc:.2f}")
```

输出会告诉你两件事：哪个置信区间不可信，以及每个阈值换来多少自动化、付出多少错误率。阈值于是从"拍脑袋的 0.7"变成"能接受 X% 的错误，因此取 Y"。

## 七、限制与结论

概率是可以用的，但有三个前提：

- 它得在你自己的任务和流量上被分桶验证过，官方说的"校准"不等于在你的场景里必然准；
- 阈值要按错误代价定，而不是照抄示例里的 0.7；
- 高风险动作不应只依赖单条概率，要么加第二个判据，要么保留人工。

`score` 的部分还要额外记住一点：期望值不是测量值。看到 1.12 就想"第 1.1 级"，等于把分布压成了一个数。

下一步值得做的验证，不是继续调阈值，而是回头改 `criteria`——把重叠的选项描述拆开、把兜底改成范围检查、把带诱导语气的样本塞进标注集。这些改动通常比在阈值上找 0.01 的空间更能提升覆盖率。

## 参考来源

| 来源 | 链接 | 检索日期 |
| --- | --- | --- |
| Decisions API 指南（confidence 定义、score 期望值与示例、最佳实践） | https://ai.google.dev/gemini-api/docs/decisions-api | 2026-11-12 |
| 同上，Answers / probabilities / confidence 一节 | https://ai.google.dev/gemini-api/docs/decisions-api#confidence | 2026-11-12 |
| 同上，Rate on a scale（score）示例与 legend 映射 | https://ai.google.dev/gemini-api/docs/decisions-api#score | 2026-11-12 |
| Gemini Decision 模型卡（等级与选项上限） | https://ai.google.dev/gemini-api/docs/models/gemini-decision-1 | 2026-11-12 |
| Google 发布稿（"校准"口径与内部用例） | https://blog.google/innovation-and-ai/technology/developers-tools/gemini-decisions-api/ | 2026-11-12 |
| Hugging Face 博客：decision model 的状态—问题—候选答案结构 | https://huggingface.co/blog/decision | 2026-11-12 |
