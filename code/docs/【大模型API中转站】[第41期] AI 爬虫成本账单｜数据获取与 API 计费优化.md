---
title: "【大模型API中转站】[第41期] AI 爬虫成本账单｜数据获取与 API 计费优化"
tags:
  - AI 爬虫
  - 数据获取
  - 成本治理
  - API 计费
description: "从 git.kernel.org 的 AI 爬虫负载账单出发，拆解数据获取的真实成本结构，并把这套成本治理逻辑映射到 API 计费优化上。"
---

# 【大模型API中转站】[第41期] AI 爬虫成本账单｜数据获取与 API 计费优化

AI 爬虫把公开站点当成免费语料库，账单却由站点运营方默默承担。git.kernel.org 的负载数据让我第一次看清了这份账单的规模：为爬虫渲染 commit 消耗的 CPU 周期，已经超过包括 git clone 在内的所有合法访问之和。

## 一、开篇痛点：账单在后台悄悄滚雪球

大多数开发者对 AI 爬虫的印象是"多刷几次页面"，但真实情况远不止于此。git.kernel.org 上的负载形成了持续不断的"背景辐射"：任何时刻，5 个地理分布节点上都有 14 个 CPU 核在专职渲染 commit 页面，只为了把 git 对象转成 HTML 喂给爬虫。

这笔开销只有一个用途——喂养训练模型。它不服务任何开发者、不服务任何真实用户，却长期占用了本可以用于 git clone、网页浏览、邮件归档查询的生产资源。更麻烦的是，它不像瞬时攻击那样有波峰波谷可以躲，而是 24 小时不间断地存在，属于最难治理的那一类成本：稳定、持续、且持续增长。

我在维护 4sapi 中转站（https://4sapi.com）时也遇到过同样的问题：成本不来自设计好的功能，而来自不可控的调用模式。理解了爬虫账单，也就理解了 API 计费为什么要做治理。

## 二、两组数字：先把账单摊开

与其空谈"爬虫很耗资源"，不如把硬数据摆出来。我整理了两组最核心的数字：

| 对比项 | 数据 | 含义 |
| --- | --- | --- |
| 渲染 commit 的 CPU 周期 vs 全部合法访问 | 超过 100% | 爬虫负载成为第一大开销 |
| 专职渲染核数 / 节点数 | 14 核 / 5 个节点 | 长期常驻，非临时高峰 |
| 渲染产物用途 | 仅训练数据 | 单一用途持续占用资源 |
| linux.git commit 总数 | 约 148 万 | 全量历史规模 |
| git.kernel.org 上的 fork 数 | 922 个 | 同一批对象的多份入口 |

第二组数字解释了第一组数字为什么会膨胀：数据规模本身很大，而爬虫又在用最贵的方式访问它。

## 三、为什么开源仓库是爬虫眼里的"金矿"

Linux 开发几乎完全公开：git 仓库可以完整 clone，讨论归档可以实时跟踪。对训练大模型来说，这是极难得的语料来源，原因有两个。

第一，数据纯度高。整个仓库历史里几乎找不到 LLM 生成的内容，是名副其实的 pre-AI 数据。第二，过滤成本低。不需要复杂的去重和清洗流程，clone 下来就是干净语料。

纯人类数据的价值在训练侧越来越被重视：把 LLM 生成的内容再喂给 LLM 训练，相当于让模型吃自己的呕吐物，会出现类似"数字朊病毒病"的退化现象——模型输出趋同、多样性和事实性持续下降。因此，能保证"从未被 AI 污染"的语料源就变得极其值钱，git.kernel.org 恰好是其中之一。

## 四、最蠢的取数方式：放着 git clone 不用

讽刺的是，git.kernel.org 一直在公开邀请所有人 clone，甚至包括整个 LKML 邮件归档——"它就是一堆 git 仓库"。正确做法应该是：

```text
git clone https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git
内部 walk 每个 commit，一次性取完全部历史
```

一条命令就能拿走的全量数据，爬虫却选择了相反的路：逐个请求 commit 的 HTML 页面，再解析页面内容。对比一下两条路径：

| 取数方式 | 数据完整性 | 服务器负担 | 效率 |
| --- | --- | --- | --- |
| git clone + 本地遍历 | 全量，含完整元数据 | 一次打包传输 | 极高 |
| 逐 commit 渲染 HTML | 全量但有损 | 每页一次完整渲染 | 极低 |
| 逐 commit 渲染后解析 | 依赖页面结构 | CPU 密集 | 最低 |

结论很明确：爬虫为了省掉 clone 之后的解析成本，把最贵的渲染成本全部转嫁给了服务器，自己只拿走了 HTML 文本。

## 五、URL 爆炸：维度乘出来的账单

更大的问题藏在 URL 数量上。linux.git 约有 148 万条 commit，而 git.kernel.org 上挂着 922 个 fork。后端的对象存储是共享的，所以合法用户访问 922 个 fork 并不会产生 922 倍开销；但在爬虫眼里，每个 fork 都是一片独立的"可抓取面"。

单个 fork 能生成的有效 URL 数量是天文数字：commit 详情页、patch 原始文本、plain 渲染、任意两个 commit 之间的 diff……仅仅一个 fork 就能组合出数十亿条有效 URL，922 个 fork 叠在一起，爬虫手里的 URL 空间达到几百亿。而它爬回来的，不过是同一批 148 万 commit 的 922 份重复。

```text
148 万 commit × 922 fork × (commit 页 + patch + plain + 任意 diff)
    = 数百亿可抓取 URL
    → 爬虫获得的是：148 万 commit × 922 份重复内容
```

成本就这样被组合爆炸放大：服务器为每一份重复完成一次完整渲染，训练侧却拿不到任何新信息。

## 六、成本结构拆解：CPU 之外还有带宽与存储

爬虫账单不是只有 CPU。我把一次爬取请求走过的资源都列了出来：

| 成本项 | 来源 | 治理抓手 |
| --- | --- | --- |
| CPU / 渲染 | 每页 commit 渲染成 HTML | 缓存、静态化、限制动态 diff |
| 带宽 | 大体积页面反复传输 | CDN、压缩、限制并发 |
| 存储 / 日志 | 访问记录、缓存副本 | 保留策略、日志降采样 |
| 运维时间 | 封禁、投诉、排查 | 自动化治理规则 |

CPU 是最显眼的，但带宽和存储同样会随 URL 膨胀而放大。任何"按页渲染"的服务，只要有足够多的合法 URL 组合，就能被爬虫刷出持续的巨额账单。

## 七、请求流与成本流：一张图看懂钱去了哪

把一次爬取生命周期画出来，成本的去向就一目了然：

```text
爬虫调度器
    │ 枚举 URL（数百亿候选）
    ▼
cgit 渲染层 ───────────► CPU 峰值：每个 commit 生成一页 HTML
    │
    ▼
git 对象读取 + diff 计算（任意两 commit 组合）
    │
    ▼
HTML 输出 → 传输 → 爬虫解析 → 语料入库
    │
    └──────────────► 训练侧收益：148 万 commit 的重复副本
```

对照合法访问的成本流：

```text
开发者 git clone
    │ 一次协议级打包传输
    ▼
本地完整仓库（无渲染、无逐页请求）
    ▼
≈ 一次传输成本，服务器几乎无额外 CPU
```

同样的数据，两种访问方式，成本相差几个数量级。治理的核心不是禁止访问，而是把访问引导到成本最低的路径上。

## 八、从数据获取成本到 API 计费：同构的问题

爬虫账单和 API 计费其实是同一个问题的两面：谁使用资源，就该由谁承担对应成本，并且要能计量、能限制、能追溯。

在 API 侧，我看到过太多类似的失控模式：无预算上限的任务循环重试、超长上下文反复请求、并发无限制的批处理脚本。这些调用模式与爬虫逐页渲染如出一辙——单次请求都不算贵，累积起来却是账单爆炸。

计费治理要回答三个问题：

1. 怎么计量——按 token、按请求数、还是按并发配额；
2. 怎么限制——预算上限、频率限制、超时与重试策略；
3. 怎么追溯——用量报表、按任务记账、异常告警。

我自己的 4sapi 就是按这个思路设计计费维度的：请求并发配额、按模型计费、用量明细可查。合法接入的开发者把预算挂在任务上，成本就变得可预期。

## 九、方案：把成本治理落到接入侧（含 Python 示例）

数据获取的成本治理不只发生在服务端，接入侧同样能做。以下是我在 4sapi 上实践过的接入方式，环境只需要 Python 3.9+ 和 requests。

第一步，配置接入参数：

```python
import requests
import time

BASE_URL = "https://4sapi.com/v1"
API_KEY = "sk-xxxx"   # 用 4sapi 控制台分配的密钥替换

INPUT_PRICE = 0.000005   # 每 1000 input token 价格，按实际计费模型填写
OUTPUT_PRICE = 0.000015  # 每 1000 output token 价格
```

第二步，写一个带预算控制的调用函数：

```python
def chat_with_budget(messages, model="gpt-4o", max_tokens=2048,
                     token_budget=200_000, max_retries=3):
    headers = {"Authorization": f"Bearer {API_KEY}"}
    payload = {
        "model": model,
        "messages": messages,
        "max_tokens": max_tokens,
    }

    for attempt in range(max_retries):
        try:
            resp = requests.post(
                f"{BASE_URL}/chat/completions",
                json=payload,
                headers=headers,
                timeout=120,
            )
            resp.raise_for_status()
            data = resp.json()

            usage = data.get("usage", {})
            in_tokens = usage.get("prompt_tokens", 0)
            out_tokens = usage.get("completion_tokens", 0)
            cost = (in_tokens / 1000) * INPUT_PRICE + (out_tokens / 1000) * OUTPUT_PRICE

            if in_tokens + out_tokens > token_budget:
                raise RuntimeError(
                    f"token 预算超限: {in_tokens + out_tokens} > {token_budget}"
                )
            return data, cost

        except requests.exceptions.RequestException:
            if attempt == max_retries - 1:
                raise
            time.sleep(2 ** attempt + 1)   # 指数退避，避免重试风暴
```

第三步，批量获取时加限速和去重，防止自己变成"爬虫"：

```python
import hashlib

def fetch_batch(requests_fn, items, rate_limit=5.0):
    seen = set()
    results = []
    for item in items:
        key = hashlib.sha256(str(item).encode()).hexdigest()
        if key in seen:          # 缓存去重，避免重复请求
            continue
        seen.add(key)
        results.append(requests_fn(item))
        time.sleep(1 / rate_limit)   # 限速，不给服务端加负担
    return results
```

这套代码把预算、重试、去重、限速全部前置到任务层，配合服务端的并发配额，成本基本不会失控。

## 十、合规边界：治理的是成本，不是合法访问

必须强调：成本治理不等于封杀或绕过。合法接入的边界很清晰：

- 遵守服务的 robots.txt 与 ToS，不伪造身份、不绕过限流；
- 数据获取优先使用官方提供的批量接口，比如 git clone、官方 API 导出；
- 不对渲染型页面做高并发逐页抓取；
- API 接入按官方配额和计费规则付费，合理优化调用次数，而不是钻计费漏洞。

我在 4sapi 上做的一切优化都建立在这个前提上：改进的是调用方式和预算管理，而不是躲避计费。

## 十一、成本治理落地清单

无论治理爬虫账单还是 API 计费，都可以按这份清单逐项检查：

1. 确认数据来源合规，不抓取明确禁止的内容；
2. 优先批量 / 全量接口，避免逐页渲染式获取；
3. 设置请求频率与并发上限，覆盖任务层和服务端；
4. 结果按内容哈希缓存，去重后再请求；
5. 每次调用带 token 预算和超时，失败走指数退避；
6. 建立用量对账：按任务记账，汇总到日 / 周报表；
7. 配置异常告警：预算超限、错误率上升立即通知；
8. 定期复盘成本项，淘汰低价值调用。

## 十二、成本与风险提示

爬虫账单揭示的成本风险有两类。一类是服务端风险：CPU 与带宽被挤占，影响真实用户体验，极端情况下需要扩容才能维持服务。另一类是把同样的坏习惯带进 API 接入：无预算的循环调用、无限重试、高并发脚本，都可能让账单在几天内膨胀到难以接受。

法律与合规风险同样不能忽视。版权、服务条款、robots.txt 都是硬约束；恶意爬取和绕过官方限制带来的不是成本问题，而是法律问题。遇到高成本场景，正确顺序是先计量，再限流，最后优化调用方式。

## 总结

AI 爬虫账单的本质，是数据获取成本与收益的错位：获取方把渲染成本转嫁给站点，自己只拿走重复的语料。git.kernel.org 的案例把这种错位量化到了 CPU 周期级别，也给出了治理方向——引导访问走成本最低的路径，而不是简单封禁。

同样的逻辑适用于 API 计费：计量、限制、追溯三件事做好，成本就可预期。我维护的 4sapi（https://4sapi.com）一直按这个思路设计配额与计费，欢迎在评论区发表想法。