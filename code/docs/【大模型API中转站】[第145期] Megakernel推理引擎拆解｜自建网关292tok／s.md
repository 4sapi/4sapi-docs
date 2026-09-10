---
title: 【大模型API中转站】[第145期] Megakernel推理引擎拆解｜自建网关292tok／s
tags:
  - 大模型API中转站
  - megakernel
  - 推理引擎
  - 网关选型
description: 一套围绕 decode megakernel 构建的推理服务引擎在 batch size 1 下跑到 292 tokens/s、硬件理论上限的 62%，这一期把它的调度、显存管理与网关选型参照一次拆清。
---

# 【大模型API中转站】[第145期] Megakernel推理引擎拆解｜自建网关292tok／s

这一期拆一套围绕 decode megakernel 构建的完整推理服务引擎：它服务的是编码模型 North Mini Code，对外暴露 OpenAI 兼容端点，支持 tool calling，并且在 batch size 1 下跑出了 292 tokens/s、相当于硬件理论上限（SoL）62% 的成绩。4sapi.com 的中转层每天在多个推理引擎之间做路由，这套 megakernel 引擎的指标给了我不少选型参照。

## 先说痛点：选推理网关时被营销数字糊脸

选推理网关这件事，最不缺的就是数字。"吞吐提升数倍""延迟大幅下降""单卡跑出极限性能"，话术一个比一个漂亮，可一旦追问细节，常常连 batch size 都说不清。我在选型上吃过的暗亏大致有三种：拿宣传页的演示环境测速很漂亮，上量之后 TPOT 直接翻倍；宣传支持超长上下文，实际长上下文下速度掉到没法看；单流速度和系统吞吐混在一个数字里说，根本分不清测的是哪一个。

所以我现在看任何性能宣称，先找三个锚点：硬件是什么、batch size 是多少、离硬件理论上限还有多远。三个锚点缺任何一个，数字就只能当故事听。这套 megakernel 引擎难得地把三个锚点都给足了：batch size 1 下 292 tokens/s，是该硬件 SoL 的 62%，比对比方案快 1.58 倍。下面就从这三个锚点开始拆。

## 先看成绩单：292 tokens/s 到底意味着什么

batch size 1 指的是一条请求独占推理资源。292 tokens/s 就是单流解码速度：一个请求持续生成时，每秒产出 292 个 token，平均约 3.4 毫秒出一个 token。对编码场景来说，一次 400 token 的函数补全大约 1.4 秒流完，这就是交互体验的直接来源。单流速度是延迟的地板——batch size 1 都快不起来，负载一上来只会更糟。

62% 这个比例比绝对值更有信息量。SoL（speed of light）是硬件的理论上限：decode 阶段每生成一个 token 都要把模型权重完整读一遍，显存带宽除以权重字节数，得到的就是天花板。拿 292 除以 62% 倒推，这块硬件的理论天花板大致在 471 tokens/s 上下（准确值以官方文档为准）。剩下的 38% 差距，来自调度、采样、内核执行等环节的真实损耗——这些损耗能压到四成以内，工程上是相当健康的水平。

1.58 倍来自横向对比：与对比方案同台测出来的差距。看到这类数字，我习惯多问一句"对比的是谁"——没有基线的倍数只是修辞，有基线的倍数至少给了复测的抓手。

## megakernel 拆解：把一整步 decode 塞进一个大内核

要理解 292 这个数字从哪来，得先看传统引擎一步 decode 是怎么跑的：layernorm 一个内核，QKV 投影一个内核，attention 一个内核，MLP 一个内核，采样又一个内核——生成一个 token 要启动几十上百个小内核。每个内核的启动都有固定开销，内核之间还要把中间结果写回显存、再读出来，宝贵的显存带宽大量花在搬运而不是计算上。

megakernel 的思路是把一整步 decode 融合成一个大内核：数据尽量留在算力芯片的存储层级里不落地，内核启动开销归零，部分调度逻辑也能搬进内核内部。收益在小 batch 场景最明显——batch size 1 时每步的计算量本来就小，启动和搬运这类固定开销的占比被放大，megakernel 恰好把这块吃掉。这也解释了为什么它能在单流场景贴近 SoL：单流拼的就是谁的开销更少。

代价同样明确：融合后的大内核与硬件、模型结构耦合更深，换模型结构就要重新适配，灵活性换性能。对以编码模型为主力、模型形态收敛的服务来说，这笔交易划算。

## 一个请求的完整旅程

把上面这些能力串起来，一个流式请求在这套引擎里是这样走完的：

```text
客户端（stream=true，tools=[...]）
   │  OpenAI 兼容请求：POST /v1/chat/completions
   ▼
OpenAI 兼容端点（鉴权、参数校验、协议解析）
   ▼
调度器（continuous batching：新请求随时入队，完成请求随时退出）
   ▼
批处理执行（ragged lengths 拼批，paged attention 按页取 KV）
   ▼
megakernel 解码（一个大内核跑完一步 decode，逐 token 产出）
   ▼
流式返回（SSE chunk 逐 token 推回客户端）
```

值得注意的是端点那一层：OpenAI 兼容不是简单的格式翻译，它是一份对外契约。正因为契约与 OpenAI 一致，客户端 SDK 不用改一行代码，就能把流量切到这套引擎上。

## 引擎分层：调度、显存、执行各管一段

```text
API 层        OpenAI 兼容端点｜tool calling 协议解析
   ↓
调度层        请求队列｜continuous batching 逐迭代调度
   ↓
显存管理层    KV 缓存分页（paged attention）｜块分配与回收
   ↓
执行层        megakernel 解码｜ragged lengths 组批
   ↓
硬件层        GPU（型号与价格以官方文档为准）
```

分层的好处在于优化有抓手：调度层决定"谁上批"，显存管理层决定"KV 放哪"，执行层决定"一步怎么算"。三层各自演进、互不拖累，62% 的 SoL 占比就是三层叠加调优的结果。

## continuous batching：让每个解码迭代都满载

静态批处理的老问题是"陪跑"：一批请求一起进、一起出，先完成的请求占着位置空等，整批的时长由最长的那个决定，新到的请求还得等下一批开批。GPU 的算力大量浪费在给已完成的序列站岗。

continuous batching 把调度粒度从"批"细化到"解码迭代"：每一步迭代前重新审视队列，谁完成了立刻腾位，新请求随时补进空位。GPU 几乎每个迭代都满载，排队时间显著缩短，高并发下的 TTFT 也更稳。

没有它会怎样：容量测算会变得非常难看——按最坏情况预留批，实际利用率却上不去，成本按峰值算，吞吐按均值拿，两头吃亏。

## paged attention：把 KV 缓存当内存页来管

KV 缓存随序列长度线性增长，是解码显存里的大头。朴素做法是按最大长度预留连续显存：请求明明只用 500 token，也占着 32k 的坑，这是内部碎片；碎片攒多了，新的大请求又整块装不进去，这是外部碎片。最难受的状态是显存看着还有余量，批却排不进去。

paged attention 借用操作系统分页的思路：KV 缓存切成固定大小的块，按需分配，用逻辑块表映射物理块。碎片几乎归零，显存利用率抬上去，同一张卡能同时挂更多序列——这是吞吐的隐形放大器。

没有它会怎样：要么 OOM 频发，要么被迫压低并发保平安，两种结局最后都折算成钱。

## ragged sequence lengths：拒交 padding 的算力税

一个 batch 里序列长短不一是常态。传统做法补齐到最长再算：800 token 的 prompt 和 20 token 的 prompt 进同一批，后者要陪着 780 个 padding token 走完整个前向。补齐部分算完就扔，纯属浪费，还拉长了整批的耗时。

ragged（不规则）序列长度的做法是把整批 token 拍平成一条连续缓冲，用偏移量记录每条序列的起止，内核按偏移索引。算力全部花在真实 token 上，prefill 更快，TTFT 直接受益。

没有它会怎样：prompt 长短差距越大的场景浪费越狠。计费按真实 token 收，成本却按 padding 付，毛利被悄悄吃掉——做中转聚合的同行对这一点应该格外有体感。

## 接入教程：OpenAI 兼容 SDK 跑通流式输出与 TPOT 计时

引擎对外暴露 OpenAI 兼容端点，接入成本几乎为零：装好 openai 这个 SDK，改动只有三处——base_url、api_key、model。api_key 从环境变量读，别写进代码仓库。下面的示例顺带演示流式输出，以及用简单计时记录 TTFT 和 TPOT：

```python
import os
import time
from openai import OpenAI

# 密钥从环境变量读取，避免硬编码进代码仓库
client = OpenAI(
    api_key=os.environ["OPENAI_API_KEY"],
    base_url="https://4sapi.com/v1",
)

messages = [
    {"role": "system", "content": "把代码问题拆成最小可复现步骤再回答。"},
    {"role": "user", "content": "用 Python 写一个带指数退避重试的 HTTP GET 封装。"},
]

start_ts = time.perf_counter()
first_token_ts = None      # 首 token 时间，用来估算 TTFT
last_token_ts = None
piece_count = 0

stream = client.chat.completions.create(
    model="north-mini-code",   # 模型名以官方文档为准
    messages=messages,
    stream=True,
)

for chunk in stream:
    if not chunk.choices:
        continue
    piece = chunk.choices[0].delta.content or ""
    if piece:
        now = time.perf_counter()
        if first_token_ts is None:
            first_token_ts = now
        last_token_ts = now
        piece_count += 1
        print(piece, end="", flush=True)

ttft_ms = (first_token_ts - start_ts) * 1000
tpot_ms = (last_token_ts - first_token_ts) * 1000 / max(piece_count - 1, 1)
print(f"\nTTFT≈{ttft_ms:.0f}ms  TPOT≈{tpot_ms:.1f}ms/token  pieces={piece_count}")
```

两点说明：一是网关逐 token 推 chunk 时，chunk 数近似 token 数，精确值以响应里的 usage 字段为准；二是 TTFT 包含排队与 prefill，TPOT 反映解码速度——这套引擎 batch size 1 下的 292 tokens/s，折算成 TPOT 大约就是 3.4 毫秒一个 token，实测值可以拿这个量级来对照。

## tool calling 实操：让编码模型自己调用工具

North Mini Code 是编码模型，Agent 场景里工具调用是刚需。这套引擎在 OpenAI 兼容端点里支持 tool calling：请求带 tools 参数，模型决定调用工具时，响应里返回结构化的 tool_calls 字段。沿用上面创建的 client：

```python
tools = [
    {
        "type": "function",
        "function": {
            "name": "run_tests",
            "description": "在本地运行指定测试文件，返回通过与否。",
            "parameters": {
                "type": "object",
                "properties": {
                    "path": {"type": "string", "description": "测试文件路径"},
                },
                "required": ["path"],
            },
        },
    }
]

resp = client.chat.completions.create(
    model="north-mini-code",   # 模型名以官方文档为准
    messages=[
        {"role": "user", "content": "跑一下 tests/test_retry.py，把结果汇总给我。"},
    ],
    tools=tools,
)

msg = resp.choices[0].message
if msg.tool_calls:
    for call in msg.tool_calls:
        print(call.function.name, call.function.arguments)
        # 这里接本地执行逻辑，再把执行结果以 role=tool 塞回对话继续推理
```

结构化字段的价值在于：不用从自由文本里正则抽数据，解析失败率趋近于零。对多引擎路由的中转层来说，tool calling 的支持度与字段完整性，是必须在能力表里逐项对齐的项目。

## 引擎能力对照表：有它和没它差在哪

| 能力 | 解决什么问题 | 没有它会怎样 |
| --- | --- | --- |
| megakernel 解码 | 一整步 decode 融进一个内核，消除启动与搬运开销 | 逐步启动大量小内核，低并发时延迟被固定开销吃掉，单流速度远离 SoL |
| continuous batching | 逐迭代调度，请求随到随进批、随完随出 | 静态组批，短请求陪长请求跑全程，GPU 空转，队列堆积，TTFT 抖动 |
| paged attention | KV 缓存按块分配，显存碎片压到最低 | 按最大长度预留显存，碎片严重，同卡可挂序列数大打折扣，OOM 频发 |
| ragged sequence lengths | 变长序列免 padding 组批，算力只花在真实 token | 补齐到最长序列，prefill 变慢，成本替 padding 买单 |
| tool calling | OpenAI 兼容协议内返回结构化 tool_calls | 只能解析自由文本，Agent 流程靠正则抽取，脆弱且难维护 |

一张表看完就明白：这五项能力共同决定了同一块硬件能不能跑出 62% 的 SoL 占比。

## 选型指标清单：测过这些才算测过网关

这套引擎给性能数字的方式——batch size、SoL 占比、对比基线一样不缺——可以直接抄进验收标准。我把日常测网关的清单整理如下：

- [ ] 单流基线：batch size 1 下的 TPOT，折算 tokens/s，与官方宣称对照
- [ ] SoL 占比：拿显存带宽与模型权重字节倒推理论上限，看实测占几成
- [ ] TTFT：空载与满载各测一轮，记录 P50 与 P95
- [ ] 并发劣化：并发 10、50、100 下的 TPOT 曲线，找拐点在哪
- [ ] 吞吐峰值：系统总 tokens/s 随并发的变化，确认拐点之后的行为
- [ ] 长上下文：8k、32k、128k 梯度下 TTFT 与 TPOT 的变化幅度
- [ ] 流式稳定性：连续长输出有没有断流、乱序、丢 chunk
- [ ] tool calling：字段完整性、嵌套参数解析成功率、多轮工具循环稳定性
- [ ] 协议兼容：stream_options、usage 字段、错误码格式与 OpenAI 是否一致
- [ ] 故障演练：上游单节点摘除时，进行中的流式连接如何表现

清单里最容易被跳过又最容易出事的是"并发劣化"和"协议兼容"：前者决定上量之后的账单，后者决定改造成本。

## 自建网关 vs 托管 API：取舍在哪

托管 API 的优势是省心：弹性扩容、免运维、按 token 付费。代价是控制权让渡——限流策略、批处理策略、高峰期的 TPOT 劣化，全都不在自己手里，成本曲线也始终带着溢价。

自建网关换回来的是控制权：调度策略、显存水位、升级节奏都可以自己定，持续大流量下单位成本通常更低，数据边界也更好锁。代价是容量规划、运维值班、内核与驱动升级全要自己扛。

这套 megakernel 引擎给了一个中间参照：把 SoL 占比当北极星指标，再决定自建与否。62% 说明引擎把硬件吃透了，自建才有意义；假如实测只有三成 SoL，自建省下的钱大概率全填进运维坑里。数字先行的决策，比"看起来很强"可靠得多。

## 中转聚合为什么依赖这类底层引擎

中转聚合做的事，是把大量客户端请求路由到多个上游推理引擎。上游引擎的能力决定了三件事。第一是能力上限：上游不支持 tool calling，中转层就不敢对客户端承诺工具调用。第二是延迟下限：客户端感知的 TPOT 等于中转开销加上游 TPOT，上游解码内核慢一拍，网关侧优化做得再好也盖不住。第三是成本结构：上游 continuous batching 与 paged attention 效率越高，单位 token 成本越低，中转层的定价空间就越大。

所以每一次上游引擎的迭代，对中转层都是一轮重新选型：重跑基准、对齐能力表、重算成本账。这类把 SoL 占比往上顶的引擎，最终会传导成整条链路的成本与延迟改善。

## 成本与风险提示

先划合规边界：只讨论合法接入——走官方或授权渠道、遵守服务条款与配额限制，不碰任何绕过官方限制的方案，这是所有架构讨论的前提。

价格与硬件型号方面，这套引擎没有公开对应数字，以官方文档为准；选型时建议把这两项列为必答项，答案落地之前不谈报价。工程上再提示三个风险：一是流式对账，chunk 数与 usage 字段的 token 数可能不一致，计费以 usage 为准并做采样校验；二是单引擎依赖，megakernel 与硬件耦合较深，驱动升级或硬件迭代可能牵动整个服务，抽象层与备选引擎预案要提前准备；三是数据合规，编码场景的 prompt 常含业务代码，自建与托管的数据边界要在合同和架构两侧同时锁死。

## 总结

这一期围绕一套 decode megakernel 推理服务引擎展开：batch size 1 下 292 tokens/s，达到硬件 SoL 的 62%，比对比方案快 1.58 倍，服务的是编码模型 North Mini Code；continuous batching、paged attention、ragged sequence lengths 三件套管住调度与显存，OpenAI 兼容端点加 tool calling 管住生态兼容。我把"batch size、SoL 占比、对比基线"三个锚点整理成了网关验收清单，也把自建与托管的取舍摆了一遍——核心结论是让 SoL 占比先说话。4sapi.com 的中转层会继续把这类引擎的实测指标沉淀进路由策略。欢迎在评论区写下对这套架构的看法，以及自建网关路上踩过的坑。
