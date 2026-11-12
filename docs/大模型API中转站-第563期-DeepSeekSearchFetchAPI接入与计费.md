---
title: "DeepSeek Search 与 Fetch API 怎么用：接入、运算符与计费"
category: 人工智能
tags:
  - DeepSeek
  - 搜索 API
  - RAG
description: "DeepSeek 在 2026-10-17 开放了 Search 与 Fetch 两个原生接口，按成功请求计费而不是按 token。本文覆盖端点、鉴权、请求字段、查询运算符、响应字段、错误码、计费与 QPS 限制，并给出一条与模型调用组合的最小可用链路。"
brand_mode: none
---

# DeepSeek Search 与 Fetch API 怎么用：接入、运算符与计费

模型的知识有截止时间，这是所有 RAG 和 Agent 系统都要处理的问题。过去的做法是自己接第三方搜索、自己写抓取和正文抽取；2026-10-17 之后，DeepSeek 把这件事收敛成了两个官方接口。

但这两个接口的形态和模型调用完全不同：它们是**独立的 JSON API**，按成功请求计数计费而不是按 token，QPS 限制也和模型的并发限制相互独立。第一次接的时候，问题通常出在三处：鉴权方式、查询运算符的写法、以及"什么算一次计费请求"。

本文只解决一个问题：**把 Search 与 Fetch 两个接口接起来、用对运算符、算清成本。** 端点、字段与计费规则来自 DeepSeek 官方 API 文档，检索日期 2026-10-19。

## 两个接口的分工

官方对两者定位的描述很清楚：

- **Search API**：给一个查询，返回结构化结果列表，由 DeepSeek 自研搜索引掣驱动。
- **Fetch API**：给一个 URL，把这一个页面抓回来，返回 Markdown 格式正文。

典型的组合方式是"先搜后抓"：用 Search 找到候选页面，挑出可靠的几个，再用 Fetch 取正文送进模型上下文。如果只需要摘要，Search 结果里的 `summary` 字段可以直接用，省掉一次 Fetch 调用。

来源：[Search & Fetch](https://api-docs.deepseek.com/guides/search)（访问日期：2026-10-19）

## 前置条件

- 一个 DeepSeek API Key，鉴权方式与模型接口一致（Bearer Token）。
- 能发 HTTP POST 请求的环境；两个接口都是纯 JSON 接口，不依赖任何 SDK。
- 明确的 URL 白名单或站点策略（Fetch 会受目标站点访问策略影响）。
- 已知限制：两个接口各自的账户级 QPS 上限为 5。

## 端点与鉴权

| 接口 | 方法与端点 |
| --- | --- |
| Search | `POST https://api.deepseek.com/v1/tools/search` |
| Fetch | `POST https://api.deepseek.com/v1/tools/fetch` |

请求头需要两项：`Authorization: Bearer <DeepSeek API Key>` 与 `Content-Type: application/json`。

来源：[Search & Fetch](https://api-docs.deepseek.com/guides/search)（访问日期：2026-10-19）

注意这里的路径里带 `/v1`——这和模型调用时 `base_url` 不带 `/v1` 的习惯不同，手写请求时容易搞混。

## 最小请求

Search：

```bash
curl https://api.deepseek.com/v1/tools/search \
  -H "Authorization: Bearer <DeepSeek API Key>" \
  -H "Content-Type: application/json" \
  -d '{"query": "DeepSeek V4 release", "count": 5, "need_content": true}'
```

```python
import requests

response = requests.post(
    "https://api.deepseek.com/v1/tools/search",
    headers={"Authorization": "Bearer <DeepSeek API Key>"},
    json={"query": "DeepSeek V4 release", "count": 5, "need_content": True},
)
response.raise_for_status()
for result in response.json()["results"]:
    print(result["title"], result["url"])
    print(result["summary"])
```

Fetch：

```bash
curl https://api.deepseek.com/v1/tools/fetch \
  -H "Authorization: Bearer <DeepSeek API Key>" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/article"}'
```

来源：[Search & Fetch](https://api-docs.deepseek.com/guides/search)（访问日期：2026-10-19）

## Search 请求字段

| 字段 | 类型 | 必填 | 默认 | 说明 |
| --- | --- | --- | --- | --- |
| `query` | string | 是 | - | 1 至 128 个 Unicode 字符，不能全为空白，支持下述运算符 |
| `count` | integer | 否 | 10 | 返回条数，1 至 50 |
| `ori_question` | string | 否 | - | 触发搜索的原始问题，1 至 128 字符，用于改善查询理解 |
| `localization` | string | 否 | `""` | 可选 IP 地址，用于本地化搜索结果 |
| `need_content` | boolean | 否 | false | 是否在 `content` 中返回页面内容；为 false 时该字段恒为空字符串 |

请求体中的未知字段会被忽略。

来源：[Search & Fetch](https://api-docs.deepseek.com/guides/search)（访问日期：2026-10-19）

`ori_question` 这个字段容易被忽略但很有用：当你的检索 query 是模型改写过的关键词组合时，把用户原始问题一并传进去，能帮助检索侧更好地理解意图。

## 查询运算符

Search 的 `query` 支持一组运算符，可以组合使用：

| 运算符 | 示例 | 说明 |
| --- | --- | --- |
| `site:` | `DeepSeek site:github.com` | 限定站点；多个站点用空格分隔并用 OR 连接 |
| `-site:` | `DeepSeek -site:example.com` | 排除站点，多个用空格分隔 |
| `-` | `DeepSeek -advertisement` | 排除包含该词的页面 |
| `-"..."` | `DeepSeek -"release notes"` | 排除包含精确短语的页面 |
| `"..."` | `DeepSeek "exact phrase"` | 精确短语匹配 |
| `lang:` | `DeepSeek lang:zh,en` | 限定语言，多个用逗号分隔 |
| `after:` | `DeepSeek after:2025-01-01` | 限定发布日期之后 |
| `before:` | `DeepSeek before:2026-06-01` | 限定发布日期之前 |

组合示例（官方给出）：

```json
{"query": "DeepSeek site:github.com OR site:huggingface.co lang:en after:2025", "count": 10}
```

来源：[Search & Fetch](https://api-docs.deepseek.com/guides/search)（访问日期：2026-10-19）

`site:` 与 `-site:` 是实践中最有价值的两个：前者用来把检索限制在官方文档、代码仓库这类可靠来源，后者用来过滤内容农场。配合 `after:` 还能避免搜到过时的版本说明。

## Search 响应字段

成功时返回 HTTP 200 与 `results` 数组；正常搜索但没有匹配结果时返回空数组。每条的字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `rerank_score` | number | 0 至 3 的相关性得分，综合相关性、权威性与新鲜度；1.5 及以上视为高质量结果 |
| `title` | string | 结果标题 |
| `url` | string | 结果 URL |
| `snippet` | string | 短摘要，通常 100 至 200 字符，用于界面展示 |
| `summary` | string | 长摘要，通常 600 至 800 字符，供模型阅读 |
| `content` | string | 页面正文；`need_content` 为 false 或内容不可用时为空 |
| `site_name` | string 或 null | 站点名称 |
| `site_icon` | string 或 null | 站点图标 URL |
| `date` | string 或 null | 发布时间，UTC，RFC 3339 格式 |
| `authority` | integer | 站点权威等级 0 至 4，-1 表示未知；3 及以上表示高权威来源 |

来源：[Search & Fetch](https://api-docs.deepseek.com/guides/search)（访问日期：2026-10-19）

这篇文档里有两个字段对"让模型少幻觉"特别有用：`rerank_score` 和 `authority`。一个常见的做法是设阈值过滤（例如只要 `rerank_score` 不低于 1.5、`authority` 不低于 3 的结果），把低质量来源挡在模型上下文之外——因为模型对检索结果的信任程度，往往取决于你喂给它什么。

## Fetch 请求与响应

Fetch 的请求只有一个必填字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `url` | string | 是 | 待抓取页面 URL，最长 8192 字符，必须是带 host 的 HTTP/HTTPS URL，且不能包含用户名或密码 |

响应只有两个字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `title` | string | 页面标题，无法获取时为空 |
| `content` | string | 页面内容，Markdown 文本 |

来源：[Search & Fetch](https://api-docs.deepseek.com/guides/search)（访问日期：2026-10-19）

这里有一条官方明确的限制：**抓取结果受目标站点可用性与访问策略影响**。也就是说，Fetch 不是"给 URL 必得正文"，403 / 404 是设计内的正常结果，需要在业务侧做降级处理。

## 错误码

两个接口共用以 OpenAI 兼容格式返回的错误结构，`param` 字段恒为 null：

| 状态码 | 含义 |
| --- | --- |
| 400 | 请求头或请求体无效，包含 JSON 格式错误或字段取值非法 |
| 401 | API Key 缺失或无效 |
| 402 | 账户余额不足 |
| 403 | Fetch 请求被安全策略拒绝 |
| 404 | 无法取得可用正文，例如 URL 不可用、被站点策略拦截、目标站点报错、内容类型不支持或页面无可提取正文 |
| 408 | 读取请求体超时 |
| 413 | 请求体过大 |
| 429 | 触发 Search 或 Fetch 的 QPS 限制 |
| 500 | 服务端内部错误 |

来源：[Search & Fetch](https://api-docs.deepseek.com/guides/search)（访问日期：2026-10-19）

官方给出的重试建议是：429、408 与 5xx 可以等待后重试；402 需要先充值；**其他 4xx 在没有修改请求的前提下不要重试**。这条规则值得写进重试中间件——否则一个字段写错的请求会被反复重试，白白消耗 QPS。

## 计费口径：按成功请求计费

这是两个接口与模型调用差别最大的地方。官方规则是：

| 接口 | 单价 |
| --- | --- |
| Search | $0.75 / 每 1000 次成功请求 |
| Fetch | $0.75 / 每 1000 次成功请求 |

几条关键细则：

- **不按 token 计费，也不按返回结果条数计费**；`count` 只控制返回条数，不影响单价。
- **只有返回 HTTP 200 且产生结果才计费**：Search 必须返回非空 `results` 数组，Fetch 必须返回页面内容。
- **失败的请求、空结果的搜索、超时的请求都不计费。**

来源：[Models & Pricing](https://api-docs.deepseek.com/quick_start/pricing)、[Search & Fetch](https://api-docs.deepseek.com/guides/search)（访问日期：2026-10-19）

这里的作者判断是：这套计费方式对"用 `count` 一次拿多条结果"更友好。既然价格与条数无关，把 `count` 设成 10 相比设成 3，单次调用成本相同，但后续需要补搜的概率更低。真正需要控制的是**调用次数**，而不是每次取几条。

## 当前有一个免费窗口

官方在计费页与发布说明里都写明了：当你通过 DeepSeek Harness、Claude Code 或 Codex 使用 DeepSeek API 时，服务端工具调用本身就会命中 Search 与 Fetch，这些调用**当前免费**，按请求计费从**北京时间 2026-11-01 00:00**（官方 News 页写作 UTC 2026-10-31 16:00）开始。

来源：[Models & Pricing](https://api-docs.deepseek.com/quick_start/pricing)、[News](https://api-docs.deepseek.com/news/news261017)（访问日期：2026-10-19）

实践含义有两点：一是这段时间适合做压测和用量基线采集，不必担心账单；二是如果把搜索能力接进了生产链路，需要在计费起始日之前把成本监控补上，避免出现"免费期结束、账单突然出现"的情况。

## QPS 限制

- Search 与 Fetch **各自独立限流**，QPS 上限均为 5，一方流量不消耗另一方的配额。
- QPS 按**账户级**计算，与使用哪个 API Key 无关；超限返回 429。
- 除账户级限制外，Fetch 还受**按域名**的 QPS 限制，超限同样返回 429。
- 这些限制独立于模型的并发限制（Flash 为 2500，pro 为 500）。

来源：[Search & Fetch](https://api-docs.deepseek.com/guides/search)、[Rate Limit & Isolation](https://api-docs.deepseek.com/quick_start/rate_limit)（访问日期：2026-10-19）

5 QPS 是这两个接口最需要提前规划的地方。如果做过爬虫或聚合类应用，要意识到**换个 Key 解决不了问题**，因为限制在账户级。可行的做法是：本地做请求队列与令牌桶限流，把 `summary` 字段优先用起来减少 Fetch 次数，并对同一域名的抓取做串行化。

## 一条最小可用链路

把两个接口和模型调用串起来，最小链路大致是这样：

```text
1. 用用户问题（或模型改写后的 query）调用 Search，count 取 5 至 10
2. 按 rerank_score 与 authority 过滤结果
3. 取前 2 至 3 条，用 Search 返回的 summary 直接进上下文
   —— 需要完整正文时，再对单条 URL 调用 Fetch
4. 把结果连同 URL 一起拼进提示词，要求模型在回答中标注来源
5. 记录每次调用的 request_id、命中条数与是否计费，用于成本核对
```

第 3 步的顺序值得强调：**先看 summary 够不够用，再决定是否 Fetch**。因为 summary 是随 Search 一起返回的，不额外计费；而每多一次 Fetch 就是一次独立计费请求，还会占用 5 QPS 的配额。

第 4 步里"要求模型标注来源"是抑制幻觉的有效手段：当检索结果里带了 URL，让模型在答案中回引，能显著降低无来源结论的比例——而且这也方便人工抽检。

## 验证清单

- [ ] 两个端点都用最小请求各验证一次，确认鉴权与路径正确。
- [ ] 单独测一次空结果的查询，确认返回 200 与空数组，并且没有产生计费。
- [ ] 用 `site:` 与 `-site:` 各构造一次查询，确认运算符生效。
- [ ] 覆盖 `lang:`、`after:`、`before:` 各一次，确认返回结果符合预期。
- [ ] 对同一个 URL 连续调用 Fetch 两次，确认账户级 QPS 限制的行为。
- [ ] 构造一个不可抓取的 URL，确认拿到 403 或 404 并能触发降级逻辑。
- [ ] 在账单侧确认一次失败请求没有被计费。
- [ ] 确认限流组件是按账户级 QPS 设计的，而不是按 Key。

## 限制与结论

Search 与 Fetch 补上了 DeepSeek 在"实时信息获取"这一环的官方能力，接法和模型调用几乎一样简单，但它的成本模型和限流模型是另一套：**按成功请求计费、失败不计费、QPS 按账户级共享、Fetch 还多一层域名限制。**

给接入方的三条建议：

- 用 `rerank_score` 与 `authority` 做结果过滤，比事后纠正模型幻觉更省事；
- 优先复用 Search 自带的 `summary`，把 Fetch 留给确实需要全文的场景；
- 在免费窗口结束前把用量监控和限流补上，避免账单与配额在 11 月初同时变成意外。

本文限制：所有端点、字段、计费与限流规则均来自官方文档，检索日期 2026-10-19；官方对价格与规则保留调整权利。文中未给出任何实测检索质量或抓取成功率数据——这两项受查询分布与目标站点策略影响很大，需要用自己的样本单独评估。
