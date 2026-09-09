---
title: "【大模型API中转站】[第140期] 百万任务代码评审实测｜确定性工程+Agent协同"
date: 2026-09-09
tags: [大模型API中转站, Code Review, Agent, 确定性工程, 评审流水线, 多智能体]
---

# 【大模型API中转站】[第141期] 百万任务代码评审实测｜确定性工程+Agent协同

我是 4sapi（https://4sapi.com），常年跑 API 接入与成本优化这条赛道。今天这期，我拿 QCon 上海上讲到的《Code Review 的确定性工程与 Agent 协同》案例当引子，把经过百万真实任务验证的模型评审，改造成一条任何团队都能照着搭的可复制流水线。

## 一、痛点：全靠模型评审，又贵又飘

我最早把模型评审用得很粗：整个 PR 的 diff 原样扔给 LLM，让模型自由发挥。跑了一个月，三个问题全部暴露。

第一是贵。一个中型 PR 的 diff 动辄上万 token，评审一次就是一次完整的大模型调用，输入侧按全文计费，输出侧还会把同一类问题翻来覆去重复报。第二是飘。同一个 diff，今天问和明天问，结论能差两三成；模型会漏掉真正的逻辑缺陷，却对一句普通注释长篇大论；更麻烦的是它偶尔会编造不存在的行号和文件。第三是不可复现。没有固定规则、没有阈值、没有版本化的 prompt，出了问题连追责都无从谈起——到底是模型的问题，还是 prompt 的问题，还是上下文被撑爆的问题，全是一笔糊涂账。

QCon 上海上讲到的案例给了一个很明确的方向：用确定性工程处理可控部分，把模糊判断交给 LLM，用多智能体协同把评审流水线化。这句话看着抽象，落到工程上就是一套四层链路，我照着搭完之后，模型层的 token 开销直接压掉了六成以上。

## 二、原理速览：评审三层分层

先把"评审"这件事拆成三层，每一层只干自己最擅长的事：

1. **确定性层**：规则、lint、静态检查。结果确定、可解释、可测试，零 token 成本，处理一切"可以编程判定"的问题，比如格式、命名、硬编码密钥、死代码。
2. **模型评审层**：LLM 负责模糊判断。逻辑缺陷、边界条件、竞态、可读性，这些需要语义理解的问题才值得花 token。
3. **人审兜底层**：工程师最终把关。合入前的裁决权必须留在人手里，机器判不了的价值判断、跨模块影响、产品意图，都由这一层收尾。

三层串起来就是一条 text 流水线：

```
git diff（本次改动）
   │  ① diff 分片：按文件 / 按 hunk 切块，单片限长
   ▼
② 规则过滤（lint + 自定义规则）──→ 命中即出票，不进模型（零 token）
   │
   ▼
③ 模型评审（按片并行，多 Agent 角色分工：通才总览 + 专才深挖）
   │
   ▼
④ 结果聚合（置信度合并：去重、分级、加权）
   │
   ▼
⑤ 人审兜底（高置信度直通，低置信度人工复核）
   └── 合入门禁：阻断性缺陷一律不放行
```

核心原则一句话：**能确定的不交给模型，必须模糊的才交给模型，最后永远留一道人审。** 这一层顺序不能乱，乱了大头成本就全压在模型层上。

## 三、评审链路成本 / 可靠性对照表

把五个环节放在同一张表里对比，取舍关系一目了然：

| 环节 | 每千行 diff 成本 | 可靠性特征 | 误报水平 | 适合处理的问题 |
| --- | --- | --- | --- | --- |
| 规则 / lint | 约 0（本地，免费） | 高且稳定，同输入同输出 | 规则定义内接近 0 | 格式、命名、硬编码密钥、死代码 |
| 静态检查 | 约 0（本地，免费） | 较高，模式匹配 | 中低 | 复杂度、空指针风险、依赖告警 |
| LLM 单 Agent 评审 | 中（按 token 计费） | 波动大，上下文越长越飘 | 中高 | 逻辑缺陷、边界条件、可读性 |
| LLM 多 Agent 分工 | 高（并行多份 token） | 分工后单点更稳，聚合后更高 | 中 | 大型 PR、跨模块改动、安全审查 |
| 人审兜底 | 最高（人力成本） | 最高，不可替代 | 低 | 合入裁决、价值判断、漏报补救 |

这张表传达的核心是：可靠性的边际收益递减，成本的边际增长却递增。所以链路顺序必须是确定性工具优先、模型评审补充、人审收尾，任何一层前置或者后置，都会同时伤到成本和可靠性。

## 四、Python 教程：diff 分片

流水线第一步，把 git diff 切成可控的小片。分片的目的有三个：单片 token 可控，杜绝上下文超限；按片并行，吞吐翻倍；坏片隔离，一片失败不影响其余分片。

```python
import subprocess

def get_diff(base: str, head: str) -> str:
    """取两个提交之间的完整 diff，上下文 30 行。"""
    return subprocess.run(
        ["git", "diff", base, head, "--unified=30"],
        capture_output=True, text=True,
    ).stdout

MAX_CHUNK_LINES = 300  # 单片行数上限

def split_diff(diff_text: str) -> list[dict]:
    """按 diff --git 标记切文件块；超长文件再按 hunk 切。"""
    chunks, cur = [], []
    for line in diff_text.splitlines(keepends=True):
        if line.startswith("diff --git") and cur:
            chunks.append({"source": "".join(cur)})
            cur = [line]
        else:
            cur.append(line)
    if cur:
        chunks.append({"source": "".join(cur)})
    return chunks
```

实际经验里，单片 300 行以内时模型评审质量最稳；超过 800 行，漏报率肉眼可见地上升，模型还会开始编行号。分片不是省 token 的小技巧，是质量底线——上下文一旦逼近窗口上限，模型的表现会断崖式下滑。

## 五、Python 教程：规则过滤（确定性层）

分片之后先过规则层，命中即出票，根本不进模型。lint（ruff、eslint 这类）直接挂上 CI，再补一组自定义正则覆盖自己仓库的高频事故，例如：

```python
import re

# (正则, 问题标签, 严重级)
RULES = [
    (re.compile(r"print\("),               "debug_print",    "low"),
    (re.compile(r"TODO|FIXME"),            "leftover_todo",  "low"),
    (re.compile(r"(?i)password\s*=\s*['\"]"), "hardcoded_secret", "high"),
    (re.compile(r"eval\("),                "unsafe_eval",    "high"),
]

def rule_filter(chunks: list[dict]) -> list[dict]:
    """规则命中即出票，source 标记为 rule，聚合阶段据此区分来源。"""
    tickets = []
    for chunk in chunks:
        lines = chunk["source"].splitlines()
        for lineno, line in enumerate(lines, 1):
            for pattern, label, severity in RULES:
                if pattern.search(line):
                    tickets.append({
                        "file": lines[0],          # diff --git 行，示意用
                        "line": lineno,
                        "label": label,
                        "severity": severity,
                        "source": "rule",
                    })
    return tickets
```

规则层的核心价值是"确定性"：同一条代码永远得到同一条结论，可解释、可追责、可测试。我每周把上一周的误报样本加进规则测试，规则越收越准。规则层每拦下一个问题，就省掉一次模型调用——这就是流水线的第一道省钱闸门。

## 六、Python 教程：模型评审 prompt 组织

规则过滤之后剩下的模糊问题，才轮到模型上场。prompt 组织有三个原则：给上下文（仓库规范、文件职责）、给角色、给输出格式（强制 JSON 且带行号）。

```python
import json
from openai import OpenAI

# 4sapi 合法中转接入：统一 OpenAI 兼容协议
client = OpenAI(
    api_key="sk-4sapi-xxxx",          # 在 4sapi 控制台申请
    base_url="https://4sapi.com/v1",  # 中转端点
)

def build_review_prompt(chunk: str, role: str, guidelines: str) -> list[dict]:
    return [
        {"role": "system", "content": (
            "我是代码评审流水线的" + role + "。"
            "只评审 diff 中新增或修改的行，不评价未改动的历史代码；"
            "输出严格 JSON：{\"issues\":[{\"severity\":\"high|mid|low\","
            "\"line\":行号,\"why\":一句话理由,\"suggestion\":修改建议}]}；"
            "没有问题时 issues 为空数组；禁止编造行号。"
        )},
        {"role": "user", "content": (
            "仓库规范：\n" + guidelines + "\n\n待评审 diff：\n" + chunk
        )},
    ]

def review_chunk(chunk: str, role: str, guidelines: str) -> list[dict]:
    resp = client.chat.completions.create(
        model="review-model",
        messages=build_review_prompt(chunk, role, guidelines),
    )
    return json.loads(resp.choices[0].message.content)["issues"]
```

要点：系统消息里把"禁止编造行号""只评新增行"写死，这两条直接砍掉一大截误报。输出格式强制 JSON，聚合阶段才能解析、去重、算置信度。还有一条容易被忽视的：prompt 本身必须版本化，存进仓库，改动要走评审流程——模型评审的 prompt 和业务代码同等重要，谁都能改的 prompt 等于没有评审标准。

## 七、Python 教程：多 Agent 角色分工

单 Agent 通才评审会同时犯两类错：广度不够，漏掉跨文件影响；深度不够，漏掉安全细节。QCon 上海上讲到的另一个案例——《AI Coding 在大型客户端工程中的落地实践》里，大型客户端项目从"一个通才模型包揽"演进到"多个专才 Agent 分工"——这件事落到评审上，就是按角色拆：

```python
ROLES = {
    "logic":      "检查逻辑缺陷、边界条件、空指针、竞态；关注跨文件调用是否一致",
    "security":   "检查注入、硬编码密钥、越权、不安全反序列化；按 OWASP 思路过一遍",
    "compat":     "检查 API 破坏性变更、废弃接口、配置项兼容、数据迁移影响",
    "readability": "检查命名、结构、重复代码、注释与实现是否一致",
}

def review_multi_agent(chunk: str, guidelines: str, roles: dict) -> dict:
    outputs = {}
    for role, desc in roles.items():
        outputs[role] = review_chunk(
            chunk,
            role + "，" + desc,
            guidelines,
        )
    return outputs
```

分工的执行方式有两种：同一份分片按角色各发一份 prompt 并行调用；或者先让"通才"跑一遍总览，把可疑行挑出来，再让专才 Agent 只盯着可疑行深挖。我实际用的是第二种——通才管广度、专才管深度，两层结合比任何单 Agent 都稳，专才的输入还更小。

## 八、Python 教程：结果聚合与置信度合并

多 Agent 会给出重复、冲突、不同严重级的问题。聚合阶段把它们合并成"票"，用角色重复度计算置信度：

```python
def merge_results(agents_output: dict) -> list[dict]:
    """按 (行号, 理由前缀) 去重，用角色重复度与严重级算置信度。"""
    merged = {}
    for role, out in agents_output.items():
        for issue in out:
            key = (issue.get("line"), issue.get("why", "")[:30])
            if key not in merged:
                merged[key] = {"count": 0, "roles": set(), "sev": set()}
            merged[key]["count"] += 1
            merged[key]["roles"].add(role)
            merged[key]["sev"].add(issue.get("severity", "low"))

    tickets = []
    total_roles = max(len(agents_output), 1)
    for key, v in merged.items():
        # 置信度 = 基础 0.4 + 角色重复度 + 高严重级加成，封顶 1.0
        conf = min(
            0.4 + 0.2 * v["count"] / total_roles + 0.1 * ("high" in v["sev"]),
            1.0,
        )
        tickets.append({
            "line": key[0],
            "why": key[1],
            "confidence": round(conf, 2),
            "severity": "high" if "high" in v["sev"] else "mid",
        })
    return sorted(tickets, key=lambda t: -t["confidence"])
```

聚合之后按置信度分级处理：confidence ≥ 0.7 直通人审，确认后直接入单；0.4–0.7 进待定池，人工抽查；< 0.4 直接丢弃。这个阈值就是可靠性杠杆——调高省人力但漏报多，调低更全但费人力。我用一周的历史票数据把阈值标定在 0.6 附近，正好卡在"不淹没人审"和"不漏真问题"的平衡点上。

## 九、通才 / 专才分工策略

从"一个通才模型包揽"到"多个专才 Agent 分工"，中间不是非此即彼，而是按改动规模分场景：

- **小改动（100 行以内）**：一个通才 Agent 足够，上专才分工纯属浪费 token。
- **中型 PR（100–1000 行）**：通才总览 + 一两个专才（安全、逻辑）补充。
- **大型变更（跨模块 / 重构）**：先通才扫广度，再按模块上专才，最后聚合。
- **特例**：涉及支付、鉴权、数据迁移的改动，无论大小都强制过 security / compat 专才，这条写进门禁规则，不靠自觉。

成本上，专才每多一个角色就多一份完整 diff 的 token。所以专才不该全量并行，而是"通才挑可疑行 → 专才只看可疑行"，把专才的输入控制在通才输出的子集上——既省 token，又因为输入更聚焦而提精度。这是整个流水线里性价比最高的一处优化。

## 十、成本与风险提示

流水线跑起来之后，有三个坑必须提前埋好护栏。

**误报疲劳。** 模型评审的误报会让人审逐渐麻木，最后连真问题一起漏掉，整条链路的可靠性瞬间归零。对策是双重：置信度阈值挡住低价值票，同时每周用历史票校准规则层，把能确定的问题全部下沉到规则层，模型层的输入面越收越窄。

**上下文超限成本。** diff 越大，单次调用的输入 token 越多，一旦超过模型窗口还要分片重发，成本按指数恶化。对策是硬约束：单片 300 行上限、超长文件先规则过滤再进模型、每片调用前检查 token 预算，超限直接降级处理而不是硬塞。

**多 Agent 放大。** 角色越多，总 token 越高，聚合逻辑越复杂。对策是前面说的专才输入收窄，以及聚合脚本先在小样本上验证再上全量——我见过聚合逻辑写错导致高严重级被丢弃的事故，聚合代码也要走评审。

**合规边界。** 评审流水线处理的是仓库代码，接入 4sapi 这类中转服务必须走合法授权通道，代码内容不落第三方日志，密钥一律走环境变量，私有仓库的代码绝不送进未经确认的端点。工程治理层面，评审记录留存、prompt 版本化、误报率周报，这三件事和流水线本身同等重要，缺一项都算没落地。

## 十一、验收清单

上线前逐条打勾，缺一条都先别全量跑：

- [ ] diff 分片单片 ≤ 300 行，超长文件有二次切分逻辑
- [ ] lint 与自定义规则已接入，规则命中不消耗 token
- [ ] 模型评审 prompt 版本化入库，输出强制 JSON
- [ ] 多 Agent 角色分工明确：通才管广度，专才管深度
- [ ] 专才输入收窄到通才输出的可疑行子集
- [ ] 置信度阈值已用历史数据标定，聚合脚本有小样本验证
- [ ] 上下文 token 预算告警，单片输入有硬上限
- [ ] 误报率周报与规则校准流程已跑通
- [ ] 合规检查：代码不落第三方日志，密钥走环境变量，接入端点已授权

## 十二、总结

确定性工程处理可控部分，LLM 处理模糊判断，多 Agent 协同把评审流水线化，人审兜底守住合入门禁——这套四层链路从 QCon 上海上讲到的百万任务验证出发，落在 diff 分片、规则过滤、prompt 组织、角色分工、置信度聚合五段代码上，每一段都在成本与可靠性之间做显式取舍。我在 4sapi（https://4sapi.com）上把这套流水线完整跑过一轮，规则层拦下的问题占六成以上，模型层的 token 开销因此压到了可接受范围。这套链路踩过的坑，欢迎到评论区聊一聊。
