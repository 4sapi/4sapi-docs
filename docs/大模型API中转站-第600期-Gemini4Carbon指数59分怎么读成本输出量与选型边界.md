---
title: "Gemini 4 Carbon 指数 59 分怎么读：每任务成本、输出量与选型边界"
category: 人工智能
tags:
  - 大模型选型
  - 模型评测
  - Gemini 4 Carbon
description: "拆解 Artificial Analysis Intelligence Index v5 上 Gemini 4 Carbon 的 59 分：指数由什么构成，每任务成本和输出量意味着什么，幻觉率与准确率如何取舍，并按任务类型给出条件化的选型建议。适合要在多个旗舰模型之间做选择的技术负责人。"
brand_mode: none
---

# Gemini 4 Carbon 指数 59 分怎么读：每任务成本、输出量与选型边界

Gemini 4 Carbon 发布后，第三方评测机构 Artificial Analysis（下文简称 AA）称它在 Intelligence Index v5 上以 59 分排第一。这个"第一"容易被读成"所有任务都该换成它"。但同一篇报道里还写着：它在 9 项评测中只拿下 4 项第一，在编码和知识工作上没有领先，每个任务的成本也不低。

本文回答一个选型问题：59 分这个数字，能支持什么决定，不能支持什么决定。内容包括指数的构成、成本和输出量的读法、幻觉率与准确率的取舍，以及按任务类型的条件化建议。

先说明一个前提：Carbon 目前没有公开 API，AA 模型页标注"No publicly available API"。下面的速度和成本数据来自 AA 对受限开放模型的测量，应视为暂定值。信息截止日期为 2026-10-30。

## 一、59 分是怎么算出来的

AA 在 [Announcing Artificial Analysis Intelligence Index v5](https://artificialanalysis.ai/articles/artificial-analysis-intelligence-index-v5)（2026-10-27，检索于 2026-10-30）中给出了 v5 的构成。

| 类别 | 权重 | 包含的评测 |
| --- | --- | --- |
| 知识工作 | 25% | AA-Briefcase（15%）、AutomationBench-AA（5%）、GDP.pdf（5%） |
| 编码 | 25% | FrontierCode 2.0（15%）、Terminal-Bench 4.0（10%） |
| 科学 | 25% | Terminal-Bench-Science（15%）、CritPt v2 难题子集（10%） |
| 通用 | 25% | AA-Omniscience v1.1（15%，其中准确率 10%、非幻觉 5%）、HLE-Diamond（10%） |

两个特点影响解读：

1. **60% 的权重来自私有测试集或保留答案的评测。** 这降低了训练数据污染的风险，但外部读者也无法自己复现。
2. **智能体类评测占大头。** 页面写明，Terminal-Bench-Science、Terminal-Bench 4.0、AA-Briefcase 和 FrontierCode 2.0 占 55% 的权重，却产生了中位数约 90% 的输出 token。

所以 59 分是一个"偏重长流程智能体任务"的综合分。如果你的主要场景是短问答或分类，这个分数的参考价值会打折。

## 二、Carbon 在指数里的位置

AA 在 [Gemini 4 Carbon is the new #1 on the Artificial Analysis Intelligence Index v5](https://artificialanalysis.ai/articles/gemini-4-carbon-new-number-1)（2026-10-28，检索于 2026-10-30）里报告：

- 综合 59 分，排第 1（页面显示共 57 个模型）。
- 9 项评测中，HLE-Diamond、Terminal-Bench-Science、AA-Omniscience、AutomationBench-AA 四项排第一。
- 编码和知识工作"有大幅进步，但没有领先"。

对照 v5 发布当天的榜单：Claude Opus 5.5（max）58 分，GPT-6 Astra（max）57 分，GPT-6.1 Sol（max）55 分，Gemini 4 Argon（high）50 分。按这组数字，Carbon 领先 Opus 5.5 约 1 分，比 Argon 高 9 分。

AA 没有在这两页给出指数的误差范围。1 分的差距不宜解读成稳定的优劣，更合理的说法是"处在同一梯队的顶端"。

## 三、没拿第一的地方也要看

AA 报告里，下面几项 Carbon 并不领先：

| 评测 | Carbon | 对照 |
| --- | --- | --- |
| AA-Briefcase（知识工作智能体） | 1609 Elo | Argon 1483；Grok 4.7 1638；Muse Wonder 1659；Kimi K3.1 Pro 1678 |
| Coding Agent Index | 59 | Claude Code 搭配 Claude Opus 5.5（max）为 62 |

另外，v5 发布稿里 AA-Briefcase 的领先者是 Claude Sonnet 5.5，1812 Elo。

AA-Briefcase 上，Carbon 比 Argon 提高了 126 点，其中演示文稿类任务提高 293 点到 1598，分析质量提高 120 点到 1671。进步很大，但仍落后好几个模型。如果你的场景是生成演示文稿或知识工作流程，榜单上有比 Carbon 更靠前的选项。

Coding Agent Index 上，Carbon 使用 Antigravity CLI 得 59，Claude Code 加 Opus 5.5 得 62。这里比较的是"模型加工具"的组合，工具不同也会带来差别。

AA 同一页还写了 Terminal-Bench 4.0 和 FrontierCode 的具体百分比，但与 v5 发布稿里的版本和领先者对不上（例如 v5 发布稿里 Terminal-Bench 4.0 的领先者是 Kimi K3.1 Pro，FrontierCode 为 2.0 版），所以本文不采用那两个数字。

## 四、成本：看三个量，不要只看单价

### 单价

Carbon 的标准价是每百万输入 token 4 美元、输出 20 美元，缓存输入便宜 95%（每百万 0.20 美元）。AA 模型页把它标为偏贵：同类模型输入价中位数是 2 美元，输出价中位数是 10 美元。发布期有 50% 折扣，AA 的图表全部按标准价计算。

### 每个任务的成本和输出量

| 模型 | 指数分 | 每个 Index 任务成本 | 每任务输出 token |
| --- | --- | --- | --- |
| Gemini 4 Carbon | 59 | $7.76 | 123k |
| Gemini 4 Argon | 50 | $7.68 | 未给出 |
| Claude Opus 5.5（max） | 58 | $10.04 | 173k |
| GPT-6 Astra（max） | 57 | 约 $6.2 | 未给出 |
| GPT-6.1 Sol（max） | 55 | $1.27 | 未给出 |

说明几点：

- Astra 的成本在 AA 两篇文章里分别写作 $6.24 和 $6.25，相差 1 美分，表中写"约 $6.2"。
- Opus 5.5 和 Sol 的分数和成本来自 v5 发布稿，其余来自 Carbon 报道。两者都属于 v5，但发布日期差一天。
- 各模型的推理强度标签不同：Carbon 是 high，其余多为 max。强度设置对成本和分数的影响很大，AA 在 v5 稿里举例：Opus 5.5 在 low 档得 35 分、每任务 $0.98，在 max 档得 58 分、每任务 $10.04。

### 三个推论

1. **Carbon 比 Opus 5.5 便宜约 23%，来自输出更少，而不是单价更低。** 每任务输出 123k，比 Opus 5.5 的 173k 少 29%。
2. **Carbon 与 Argon 同价，分数高 9 分。** 两者的每任务成本只差 $0.08（7.76 对 7.68）。如果你已经在用 Argon 且预算不变，这是升级的主要理由。
3. **更便宜的组合在榜上存在。** Sol 的分数比 Opus 5.5 低 4 分以内，成本约为八分之一（$1.27 对 $10.04）。按同样算法，Carbon 比 Sol 高 4 分，成本约为 6.1 倍（7.76 除以 1.27）。是否值得，取决于这 4 分对你的任务意味着什么。

### 输出量

AA 模型页写明，Carbon 在跑完整个 Index 时共输出约 200M token，评测的中位数是 71M。AA 用"非常啰嗦"描述它。每任务的输出比 Opus 5.5 少，不等于输出少，它在全体模型中仍属输出偏多的一类。输出多会拉高延迟和账单，在高频场景尤其明显。

速度方面，AA 测得约 82.5 tokens/s，同类平均 80。

### 折扣的影响

如果发布期折扣一直保持，且 token 用量不变，标准价下的 $7.76 大致对应 $3.9 左右。这是单纯的算术，折扣何时结束没有公布，预算建议按标准价做。

## 五、幻觉率与准确率：一个要取舍的指标

AA-Omniscience v1.1 衡量知识问答的准确率和幻觉率，AA 报告：

| 模型 | Omniscience 得分 | 准确率 | 幻觉率 |
| --- | --- | --- | --- |
| Gemini 4 Carbon | 54 | 58% | 11% |
| Gemini 4 Argon | 46 | 约 48%（推算） | 4% |
| Claude Opus 5.5（max） | 48 | 未给出 | 50% |
| GPT-6 Astra（max） | 未给出 | 未给出 | 50% |

Argon 的准确率是按"Carbon 的 58% 比 Argon 高 10 点"推算的。

这组数字说明两点：

1. Carbon 的幻觉率比 Argon 高：从 4% 升到 11%。AA 说它仍然是所有模型里第二低，只比 Argon 高。
2. Carbon 的总分更高，是因为准确率提升的幅度大于幻觉增加的代价。

对选型的含义是：如果你的场景"答错的代价远大于答不出的代价"，例如面向用户的事实问答、合规摘要，Argon 的低幻觉率可能比 Carbon 的更高总分更重要。如果场景需要更多覆盖，并且有人工复核或检索兜底，Carbon 的准确率提升更有价值。

幻觉率的定义和计算口径见 AA 的说明，这里不展开。

## 六、按任务类型的选型建议

下表是条件化建议，依据是上面的 AA 数据。它不是对 Carbon 实际表现的保证，因为开发者现在还不能调用它。

| 任务类型 | 看哪项指标 | 倾向 |
| --- | --- | --- |
| 长流程编码智能体 | Coding Agent Index，每任务成本 | Carbon 排第二，与榜首相差 3 点；再比较组合成本 |
| 科学与研究类智能体 | Terminal-Bench-Science、HLE-Diamond | Carbon 在这两项排第一，值得优先测试 |
| 低容错的事实问答 | 幻觉率 | 先看 Argon，再测 Carbon |
| 演示文稿、知识工作流程 | AA-Briefcase | 榜上有排名更靠前的模型，Carbon 不是首选 |
| 高频、成本敏感 | 每任务成本，输出量 | 单价偏高、输出偏多，优先看更便宜的档位 |
| 已经在用 Argon，预算不变 | 同价下的 9 分提升 | 开放后值得做一轮对比 |

## 七、开放之后怎么验证

榜单只能缩小候选范围。Carbon 对开发者开放后，用自己的任务做一轮对比：

1. 选 30 到 50 个有明确判定标准的真实任务，覆盖你最常见的 3 到 4 类。
2. 对候选模型使用同一个工具和同一套提示词，推理强度写进记录。
3. 每个任务至少跑 3 次，记录成功率、输出 token、耗时和费用。
4. 统计"每个成功任务的成本"，而不是"每次调用的成本"，因为失败重试也要花钱。
5. 对低容错类任务，单独统计错误答案的比例。
6. 按标准价和折扣价各算一遍，看结论会不会反转。

第 4 点最容易被忽略。一个单价低但成功率也低的模型，每个成功任务的成本可能更高。

## 八、限制与结论

AA 的 59 分是一个有价值的综合参考，但它的适用范围有明确边界：

- 它偏重长流程智能体任务，对短任务的代表性有限。
- Carbon 与 Opus 5.5 只差约 1 分，AA 没有给出误差范围。
- 各模型的推理强度标签不同，成本对比只能作粗略参考。
- Carbon 没有公开 API，速度和成本是受限开放期的测量值。
- 幻觉率比 Argon 高，低容错场景需要单独评估。

结论：如果你做的是长流程编码或科学研究类智能体，Carbon 值得在开放后排进评测。如果你在意低幻觉或低成本，Argon 和价格更低的档位仍然是合理的备选。下一步的验证建议是，用自己的任务按上面的流程跑一轮，并以"每个成功任务的成本"作为主要比较量。
