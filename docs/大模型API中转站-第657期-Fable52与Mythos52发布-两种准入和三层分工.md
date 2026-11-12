---
title: "Fable 5.2 与 Mythos 5.2 发布：同一底模的两种准入和三层模型分工"
category: 人工智能
tags:
  - Claude Fable 5.2
  - 模型选型
  - 模型准入
description: "Claude 在 2026 年 11 月 9 日同时发布 Fable 5.2 与 Mythos 5.2，两者规格和价格完全一致。本文说明两个名字背后的准入差异、与 Opus 5.5 和 Sonnet 5.5 的分层判断、数据留存这道硬门槛，以及发布稿里的基准数字该怎么读，适合正在决定是否排期接入的开发者与技术负责人。"
brand_mode: none
---

# Fable 5.2 与 Mythos 5.2 发布：同一底模的两种准入和三层模型分工

2026 年 11 月 9 日，Anthropic 同时发布了 Claude Fable 5.2 和 Claude Mythos 5.2。两个模型名字不同，规格表却完全一样：1M 上下文、128K 最大输出、每百万输入 8 美元、每百万输出 40 美元。差别不在能力，而在谁能调用——Fable 面向所有客户，Mythos 只发给通过验证项目的机构。

这种"一个模型、两个产品名"的做法会让选型讨论变得含糊。需要更强推理的团队看到两个名字；做安全审计的团队，能用的可能恰恰是被收紧的那一个。

本文只处理四件事：两个名字的实际差异，它在现有四档模型里的位置，Mythos 的申请路径长什么样，以及发布稿里的评测数字在什么条件下成立。文章不比较它和其他厂商模型谁更好，那需要同一套任务上自己跑一遍。

文中价格、版本、日期以 2026-11-12 检索到的官方页面为准，引用前请重新打开核对。

## 一、两个名字，一套权重，两套分类器

官方发布稿写明，Fable 和 Mythos 是同一个底层模型。Fable 在它上面叠了一层网络安全与生物安全护栏，因此可以开放给所有人；Mythos 去掉这层限制，只提供给在网络安全或生命科学方向有正当用途、且通过验证的机构。

这是一个容易被忽略的设计：护栏不是另一个"降智版"模型，而是同一套权重外面的过滤器。官方把"减少误拦"列进了这一版的改动，称与 Fable 5.1 相比，claude.ai、Claude Code 和 Cowork 上的拦截量下降约 55%，其中网络安全约 60%、生物约 40%。这些数字来自官方自测，没有第三方复现，只能作为方向参考。

对调用方来说，这有个具体后果：被拦下的请求不是网络错误，而是 HTTP 200 加 `stop_reason: "refusal"`，`stop_details` 里写明触发了哪一类策略。如果你的代码只按 HTTP 状态码判断成败，这类请求会被当成正常回复处理，拿到一段拒绝文本继续往下跑。

## 二、它在四档模型里的位置

平台文档给这一版的定位是"要求高的推理与长周期 Agent 任务"，并建议：大多数工作负载从 Opus 5.5 开始，当你在 Opus 5.5 上把 effort 调高仍然达不到要求时，再往上走到 Fable 5.2；向下是 Sonnet 5.5 和 Haiku 5.5。

| 模型 | 定位 | 价格（输入/输出，每百万 token） | 对比延迟 | API 默认 effort |
| --- | --- | --- | --- | --- |
| Fable 5.2 | 高强度推理、长周期 Agent | $8 / $40 | 中等 | high |
| Opus 5.5 | 长跑 Agent 编码与知识工作 | $4 / $20 | 中等 | medium |
| Sonnet 5.5 | 速度与智能的平衡点 | $2 / $10 | 快 | high |
| Haiku 5.5 | 高并发、低延迟的分类与路由 | 起 $0.10 / $0.50 | 最快 | medium |

表里最后一列比价格更容易被忽略。Fable 5.2 的 API 默认 effort 是 high，Opus 5.5 是 medium——同一段省略了 `effort` 字段的代码，从 Opus 5.5 换到 Fable 5.2，会在不知不觉中高一档运行，请求仍然成功，只是账单和延迟变了。迁移文档把这一条单独列了出来，逐项排查方法见第 659 期。

第三个身份容易误判。Fable 5.2 也可以被放在 Advisor 或 Orchestrator 的位置：

- 作为 Advisor，Haiku、Sonnet、Opus 在执行中向它请教难题；
- 作为 Orchestrator，它把简单任务分给更小的模型。

两种用法都不需要它跑完全程，成本结构也不同。

## 三、Mythos 的准入长什么样

Mythos 5.2 不能自助开通。官方给的路径是验证项目：网络安全方向走 Cyber Verification Program，生命科学方向走 Life Sciences Verification Program。前者有公开帮助页，可以当作判断依据。

它分三档，能力范围和申请资格都不同：

| 档位 | 允许的工作 | 谁能申请 | 可用套餐 | 席位上限 |
| --- | --- | --- | --- | --- |
| Defense Access | 安全运营、事件响应、恶意软件分析、检测工程、漏洞初筛与验证 | 机构与个人 | Pro、Max、Ultra、Team、Enterprise、API | 无默认上限，个人授权只覆盖本人 |
| Red Team Access | 授权渗透测试、红队、对抗模拟、攻击工具与漏洞利用开发 | 仅机构 | Team、Enterprise、API | 75 |
| Specialized Access | 关键基础设施的授权渗透与高风险技术研究 | 仅机构 | Team、Enterprise、API | 25 |

几个判断细节：

- 个人研究者、开源维护者、漏洞赏金参与者只能申请 Defense Access，后两档不对个人开放。
- 三档从 Mythos 5.2 起都包含 Mythos、Opus、Sonnet 模型，档位调整的是护栏强度，不是模型清单。
- 每档都禁止数据或设备批量破坏、拒绝服务攻击、现实世界损害、勒索、选举干预，以及任何带明确恶意意图的请求。
- 通过 Amazon Bedrock 使用需要先具备 Enterprise Frontier Safeguards。

如果你的团队不做双用途的安全或生物研究，这条路径本来就不适用。申请不下来不是流程问题，是用途不匹配。

## 四、比价格更早拦住你的是留存要求

Fable 5.2 与 Mythos 5.2 都要求 30 天数据留存，未经 Anthropic 明确授权不能用在零数据留存（ZDR）环境。这一条在模型页、what's new 页和迁移文档里出现了三次，措辞一致。

后果是硬性的：

- 工作区当前跑 Opus 5.5 且开了 ZDR 时，把模型名换成 `claude-fable-5-2`，Claude API 上每个请求都返回 400 `invalid_request_error`；
- 不是偶发，是全部，重试也不会变；
- 解法是先给工作区打开 30 天留存，或联系账号团队。

选型顺序要调整：先确认合规侧能不能接受 30 天留存，再谈价格和评测。反过来做，容易在预算会上定下方案，最后卡在数据条款上重来。

## 五、发布稿里的数字该怎么读

官方对照表列了五个模型，其中几项是：

| 评测 | Fable 5.2 | Fable 5.1 | Opus 5.5 | GPT-6 Astra | Gemini 4 Carbon |
| --- | --- | --- | --- | --- | --- |
| FrontierCode 2.0（Agent 编码） | 49.1% | 32.4% | 45.5% | 42.8% | 41.1% |
| FrontierSWE v2（Agent 编码） | 69.9% | 56.3% | 62.3% | 65.5% | 61.2% |
| Humanity's Last Exam（带工具） | 74.2% | 65.6% | 67.7% | 57.2% | 未列 |
| OSWorld 2.1（部分离线集） | 89.3% | 83.5% | 87.2% | 79.5% | 81.5% |
| RoboDojo Sim | 64.9% | 27.6% | 39.4% | 29.8% | 未列 |

读表注意三件事。

**第一，分数绑在 effort 档位上。**

脚注写明所有分数都是该模型可用的最高 effort：Fable 5.2 的 49.1% 是 Xhigh 档跑出来的，Opus 5.5 的 45.5% 与 GPT-6 Astra 的 42.8% 同理。不同档位的分数不能跨模型直接比，尤其当你要买的是默认档位时。

**第二，竞品分数由 Anthropic 自己跑。**

同一套基准自测时，测试脚本、重试次数、工具调用上限都可能不同。跨厂商的差异只有在你关心的任务上复现过才算数。

**第三，这张表不覆盖每任务多少钱。**

发布稿另给了两条官方内部观察：在 FrontierCode 2.0 上，High 档的 Fable 5.2 与 Max 档的 Opus 5.5 分数接近，单任务成本约为后者一半；在 WANDR 上，Medium 档用约四成成本拿到 Opus 5.5 的最高分。两条都是特定基准、特定档位的条件结论，不能外推成"整体更便宜"。Artificial Analysis 在它自己的任务集上测到相反方向，数字与口径差异放在第 658 期对照。

## 六、被拦住之后怎么继续

Fable 5.2 的拒绝是可处理的。除了读 `stop_reason` 和 `stop_details`，还可以用服务端回落：在 Claude API 上带 beta 头 `server-side-fallback-2026-07-01`，设 `fallbacks: "default"`，被拒的请求会转到官方为该类别推荐的模型。Fable 5.2 允许的回落目标是 Opus 4.8 和 Opus 5。

两个限制：

- `reasoning_extraction` 这类分类没有推荐回落模型，拒绝会原样返回，应改提示词而不是重试（[refusals-and-fallback 的拒绝分类表](https://platform.claude.com/docs/en/build-with-claude/refusals-and-fallback)，检索于 2026-11-12）。
- 回落之后新模型读不到 Fable 5.2 的思考块，API 会丢弃它们，请求成功但不计这部分输入费用。

如果目标是"稳定拿到结构化输出"，而不是"拿到一个可能被拦的答案"，路线应该反过来：先用 `tool_choice: {"type": "auto"}` 加 `strict: true` 或 JSON outputs 约束格式。Fable 5.2 不支持强制工具调用，`any` 与具名 `tool` 都返回 400。

## 七、一个可以照着走的判断顺序

1. **合规先行。** 工作区能否接受 30 天留存？不能，就先解决留存或申请授权，其余讨论暂时没有意义。
2. **确认用途。** 核心需求若是漏洞复现、恶意样本分析、授权红队这类双用途工作，该走 Cyber Verification Program，而不是 Fable。
3. **看现有档位。** 现在跑在 Opus 5.5 上的任务，先把 effort 提到 high 或 xhigh 再测一遍；达不到要求，才有理由往上换。
4. **只换真正卡住的部分。** 长周期 Agent 会话、一次性读完大文档的分析、跨多轮的工具编排，是它的适用面；分类、抽取、路由留在 Haiku 5.5 上更合理。
5. **设好 effort 再上线。** 显式写 `effort`，别依赖默认值；high 及以上档位给足 `max_tokens`，因为思考 token 也计入这个上限。
6. **准备回退。** 保留一条回到 Opus 5.5 或 Fable 5.1 的路径，并接受回退那一轮会丢思考块、变慢、变贵。

## 八、限制与结论

门槛不在能力，而在三处：

- Mythos 的机构验证；
- Fable 的 30 天留存；
- 默认 effort 从 medium 变成 high 带来的隐性成本。

前两个是准入问题，谈价格之前就该确认；第三个是计费问题，下一期展开算。

能力层面能确认的是：官方在编码、Agent 搜索、计算机操作、机器人控制上都给了具体分数，其中 FrontierCode 2.0 的 49.1% 是官方测过的已发布模型里最高的；第三方指数把这一版排在 65 分（Max 档），也高于上一代与 Opus 5.5。但这些数字都建立在各家自己的任务集和档位上。官方自述的"1000 个子代理 8 天写出可启动的操作系统""30 小时完成费马大定理形式化"这类案例，目前只有官方材料，没有可复现的第三方记录。

要验证是否值得换，最省事的做法不是读更多评测，而是从自己的评估集里挑 20 到 30 个真实任务，在 Opus 5.5 与 Fable 5.2 上各跑一遍 high 与 xhigh，同时记录 token 用量、被拦次数和单任务耗时。

## 九、只剩一次预算，先测哪批任务

挑线上调用量最大的那类任务，而不是最难的。价格差能否摊平，取决于高频路径，不取决于极限案例。

## 参考来源

| 来源 | 链接 | 检索日期 |
| --- | --- | --- |
| Anthropic：发布稿（含价格、基准表、安全与回落） | https://www.anthropic.com/claude-fable-and-mythos-5-2 | 2026-11-12 |
| Claude Platform Docs：Fable 5.2 模型页 | https://platform.claude.com/docs/en/models/fable-5-2/overview | 2026-11-12 |
| Claude Platform Docs：模型总览与价格 | https://platform.claude.com/docs/en/models/overview | 2026-11-12 |
| Claude Platform Docs：What's new | https://platform.claude.com/docs/en/models/fable-5-2/whats-new-fable-5-2 | 2026-11-12 |
| Claude Help Center：Cyber Verification Program | https://support.claude.com/en/articles/14604842-cyber-verification-program | 2026-11-12 |
| Artificial Analysis：Fable 5.2 指数页 | https://artificialanalysis.ai/models/claude-fable-5-2 | 2026-11-12 |
