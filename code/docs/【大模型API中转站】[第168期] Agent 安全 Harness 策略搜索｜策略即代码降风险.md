---
title: "【大模型API中转站】[第168期] Agent 安全 Harness 策略搜索｜策略即代码降风险"
date: 2026-09-14
tags:
  - Agent安全
  - Harness
  - 策略即代码
  - API中转
  - 风险控制
description: "不改模型权重，把安全做进 harness：用自然语言策略加可执行逻辑的组合，把智能体的运行风险压进可控范围。"
---

# 【大模型API中转站】[第168期] Agent 安全 Harness 策略搜索｜策略即代码降风险

这一期想清楚了一件事：智能体的安全，与其押在模型自觉上，不如做进 harness 里，变成一层看得见、改得动、能回滚的代码。模型是冻结的，策略不该跟着一起冻结——这是我在 4sapi 上来回切换过许多模型之后，最确定的一条经验。

## 一刀切的安全，为什么总在两头挨打

策略写得太紧，代价立刻显现在可用性上：模型频繁拒绝本该执行的动作，任务在最后一步卡死，用的人开始绕开安全层，安全反而被架空。策略写得太松，代价延迟显现：一次提示注入、一次路径穿越、一次失控重试，事故直接落到账单和日志里。更麻烦的是，两类失败还会互相转化——被误杀太多次的团队会把策略整体放松，放松之后出一次事故又全面收紧，来回横跳，永远停在两个坏选项之间。

我在实际跑智能体的过程里反复看到同一种失败模式：安全被当成一段提示词，写进系统提示就算完事。提示词是软的——上下文一长就被稀释，一句注入就可能被覆盖，换个措辞就能试探出边界。真正扛得住的约束从来不是提示词，而是结构：动作从提议到执行之间，隔着一层模型说服不了的代码。安全与效用的权衡，也不该靠人工反复横跳去凑合，而应变成一个可以被工程化优化的对象。

## 策略 harness 的两层结构：软约束与硬约束

把安全 harness 拆开看，其实是两层材料的组合。

第一层是自然语言策略：写在系统提示里的边界说明，告诉模型哪些动作在授权范围内、哪些区域碰不得、遇到模糊情况怎么上报。它的优点是便宜、快、好迭代，改一句话就能重新部署；缺点是软，模型可能忽略、误解，甚至被精心构造的上下文带偏。

第二层是可执行逻辑：动作白名单、路径校验、网络校验、预算与熔断，全部落在代码里。它不解释、不商量，越界动作在执行前就被拦下，模型的任何措辞都改变不了判定结果。

两层的关系像门上贴的告示和门上的锁：告示负责让正常人守规矩，锁负责让不讲规矩的人进不来。只贴告示不上锁，等于把安全寄托在模型的自觉上；只上锁不贴告示，模型会在错误方向上反复撞墙，浪费步数也拉高成本。好的 harness 让两层各司其职：软约束把越界提议的发生率压下去，硬约束保证越界提议的后果为零。

## 原理速览：一条动作的治理链

一条动作从模型嘴里出来，到真正落在系统上，中间应该走完一条固定的治理链。整条链的关键设定只有一个：模型有提议权，没有执行权。

```text
模型提议动作
   │
   ▼
自然语言策略层（软约束：系统提示里的边界与禁区说明）
   │  引导失败，动作仍然越界
   ▼
可执行逻辑层（硬约束：动作白名单 / 路径沙箱 / 网络白名单 / 预算与熔断）
   │
   ├─── 校验通过 ───► 执行动作 ───► 结果回传模型，进入下一步
   │
   └─── 校验失败 ───► 拒绝执行 ───► 拒绝原因回传模型，计入熔断计数
                                     │
                                     ▼
                               审计日志（放行与拒绝全部留痕）
```

治理链一旦成型，安全就不再是一次次临时补救，而是一条可以被测试、被回归、被持续改进的流水线。每一个判定环节都可以单独写单元测试，每一条拒绝路径都可以在上线前演练，每一次策略变更都可以对着历史日志做回放验证。事后任何一个动作都能回答三个问题：谁提议的、为什么放行或拒绝、当时的上下文是什么。

## 联合搜索为什么有效：成对优化而非各管各的

这期的技术主角是 EvoSafeHarness。它面对的局面很典型：模型是冻结的，权重动不了；目标领域又各有各的风险面，通用安全模板直接套上去总是水土不服。它的做法是把自然语言策略和可执行逻辑当成一对联合变量去搜索——不是先写好提示词再补一段校验代码，而是让两层在同一次搜索里互相适配，最后产出一套可以直接部署的安全 harness。在多个智能体基准上，这套方法实实在在改善了安全-效用权衡。

这个结果印证了我自己的体感：策略和规则分开调，很容易错位。提示词收紧了，白名单没跟上，模型规规矩矩提的动作照样被误杀；白名单放开了，提示词还停在旧边界上，模型在过时的约束里空转。成对优化把两层当成一个整体来调，软硬互相校准，安全抬起头的同时效用不掉下去。

领域定制是另一半价值：编码智能体怕的是路径穿越和危险命令，数据智能体怕的是越权查询和网络外联，风险面不同，策略就该不同。联合搜索恰好把这种差异变成了可计算的优化目标——同一个冻结模型，接到不同领域就长出不同的安全外壳。

## 威胁面在扩大：把 Key 当作一定会被盯上的资产

再看环境侧。公开的威胁情报处置案例，时间跨度从 2025 年 12 月一直排到 2026 年 8 月，覆盖的方向包括前沿编码与智能体能力被试图用于恶意目的，其中甚至出现了武器工程方向的滥用尝试。平台方在持续识别、持续处置，这条时间线没有收窄的迹象。

对接入方来说，结论不需要悲情，只需要一条假设：流经我 Key 的流量里，迟早会混进不干净的请求——可能是被注入内容操纵的智能体，可能是复用 Key 的内部人员，也可能是凭证失窃后的异地调用。所以我按最坏情况设计防线：Key 按项目、按环境拆分，额度设上限，行为留痕迹，用量有基线，偏离基线就告警。

边界也要先说清楚：讨论视角始终在防御侧——合法接入、策略设计、监测与审计，管的是授权范围内的资产。攻击手法不属于这里的讨论范围，滥用方向的存在本身只说明一件事：审查与处置机制必须是常备项，而不是事发后的临时补丁。

## 停止规则是底线：不能指望模型自己踩刹车

持续运行的智能体有一组绕不开的行为机制：目标怎么定、失败了怎么重试、结果怎么验证、什么时候必须停。这组机制目前在 harness 层大多靠手工实现——这恰恰说明它还没有标准化，也恰恰说明它最容易被漏掉。

我在账单上学到的第一课是重试必须有上限：没有上限的重试遇到持续性失败，就是一台烧钱机器，烧完额度才想起来停。第二课是验证必须独立于生成：让同一个模型既干活又自查，错误会被自信地放大，关键结论要有独立的校验通道。第三课是停止规则必须写进代码：步数预算、连续失败熔断、关键动作前的人工确认点，至少要有一项在每条执行路径上生效。

人工确认点值得单独一提：文件删除、对外发消息、生产环境变更这类动作，模型提议之后停下来等人点头，多花几分钟，省掉的是整晚的回滚。模型可以也应当被引导着遵守规则，但兜底机制永远不能建立在自觉上。刹车装在模型脑子里叫修养，装在 harness 里才叫工程。

## 接入教程：一个最小可用的策略网关

思路落地成代码，其实百来行就够。下面是一个最小可用的策略网关：动作白名单、路径校验、网络校验、预算熔断、审计日志全部齐备。接入层我走 4sapi 的 OpenAI 兼容接口，拿 Key、换 base_url 两步就能把现有代码切过来；中转层本身还能再挂一道配额与频控策略，等于接入侧和中转侧各有一条治理链。

先准备环境变量：

```text
export OPENAI_API_KEY=sk-xxxxxxxx          # 按项目与环境拆分，定期轮换
export OPENAI_BASE_URL=https://4sapi.com/v1
```

然后是完整代码：

```python
import json
import os
from urllib.parse import urlparse

from openai import OpenAI

# ---------- 策略配置：硬约束，全部写死在代码里 ----------
ALLOWED_ACTIONS = {"read_file", "write_file", "http_get", "finish"}
ALLOWED_HOSTS = {"pypi.org", "files.pythonhosted.org"}
MAX_TOTAL_STEPS = 40           # 单次运行步数预算，防止无限循环
MAX_CONSECUTIVE_FAILURES = 3   # 连续失败熔断线


class PolicyDenied(Exception):
    """策略拒绝：动作本身未必有恶意，只是超出了授权范围"""


def check_path(raw: str) -> str:
    """路径校验：归一化后必须落在 workspace 沙箱内"""
    normalized = os.path.normpath(raw).replace("\\", "/").lstrip("/")
    if normalized.startswith(".."):
        raise PolicyDenied(f"path escapes sandbox: {raw}")
    if normalized != "workspace" and not normalized.startswith("workspace/"):
        raise PolicyDenied(f"path outside workspace: {raw}")
    return normalized


def check_network(url: str) -> str:
    """网络校验：只放行显式白名单内的主机"""
    host = (urlparse(url).hostname or "").lower()
    if not host or host not in ALLOWED_HOSTS:
        raise PolicyDenied(f"host not in whitelist: {host}")
    return host


class Budget:
    """停止规则的代码形态：步数预算加连续失败熔断"""

    def __init__(self) -> None:
        self.steps = 0
        self.fail_streak = 0

    def tick(self, ok: bool) -> None:
        self.steps += 1
        if self.steps > MAX_TOTAL_STEPS:
            raise PolicyDenied("step budget exhausted, run halted")
        self.fail_streak = 0 if ok else self.fail_streak + 1
        if self.fail_streak >= MAX_CONSECUTIVE_FAILURES:
            raise PolicyDenied("circuit break: too many consecutive failures")


def audit(event: dict) -> None:
    """审计日志：一行一个 JSON，可回放、可告警"""
    with open("agent_audit.jsonl", "a", encoding="utf-8") as fh:
        fh.write(json.dumps(event, ensure_ascii=False) + "\n")


client = OpenAI(
    base_url="https://4sapi.com/v1",       # OpenAI 兼容接口
    api_key=os.environ["OPENAI_API_KEY"],  # Key 只从环境变量读取，不进代码库
)

SYSTEM_POLICY = (
    "我是运行在受控网关后的智能体，只能提议以下 JSON 动作："
    "read_file(path)、write_file(path, content)、http_get(url)、finish(summary)。"
    "动作一旦越界会被网关直接拒绝，规划时永远待在授权范围内。"
)


def propose(messages: list) -> dict:
    """冻结模型只负责一件事：提议下一个动作"""
    resp = client.chat.completions.create(
        model=os.environ.get("AGENT_MODEL", "gpt-4o-mini"),
        messages=messages,
        temperature=0,
    )
    return json.loads(resp.choices[0].message.content)


def dispatch(action: dict) -> str:
    """白名单分发：不在名单里的动作连执行分支都进不去"""
    name = action.get("action")
    if name not in ALLOWED_ACTIONS:
        raise PolicyDenied(f"action not whitelisted: {name}")
    if name == "read_file":
        with open(check_path(action["path"]), encoding="utf-8") as fh:
            return fh.read()[:2000]
    if name == "write_file":
        path = check_path(action["path"])
        with open(path, "w", encoding="utf-8") as fh:
            fh.write(action.get("content", ""))
        return f"written: {path}"
    if name == "http_get":
        return f"allowed GET {check_network(action['url'])}"  # 示例从简，生产环境换成受控 HTTP 客户端
    return ""


def run(task: str) -> str:
    """主循环：提议、校验、执行、留痕，直到任务完成或预算耗尽"""
    budget = Budget()
    messages = [
        {"role": "system", "content": SYSTEM_POLICY},
        {"role": "user", "content": f"任务：{task}"},
    ]
    while True:
        proposal = propose(messages)
        name = proposal.get("action", "")
        try:
            result = "" if name == "finish" else dispatch(proposal)
        except PolicyDenied as exc:
            budget.tick(ok=False)
            audit({"action": name, "verdict": "deny", "reason": str(exc), "detail": proposal})
            messages.append({"role": "user", "content": f"动作被策略网关拒绝：{exc}，请改提一个授权范围内的动作"})
            continue
        budget.tick(ok=True)  # 预算耗尽时从这里终止整场运行
        audit({"action": name, "verdict": "allow", "detail": proposal})
        if name == "finish":
            return proposal.get("summary", "done")
        messages.append({"role": "user", "content": f"执行结果：{result[:500]}"})


if __name__ == "__main__":
    print(run("阅读 workspace/notes.txt，把要点整理后写回 workspace/summary.md"))
```

几个实现细节值得展开。路径校验先归一化再判前缀，先挡住 `..` 这类逃逸写法，再要求路径必须落在 workspace 之内；网络校验只认显式白名单，解析不出主机名的请求直接拒绝；Budget 类把停止规则变成类型化异常，预算耗尽直接终止整场运行而不是继续请求；审计日志一行一个 JSON，事后回放整场运行时，每一次放行与拒绝都有据可查。

执行器部分是示例实现，read_file 与 write_file 用了最朴素的文件操作，生产环境应换成受限文件客户端，http_get 换成带超时与重定向校验的请求封装。还有一条纪律值得单列：策略代码和业务代码分开存放、分开评审，策略变更走独立流程——这是策略即代码的另一半含义，策略不仅要写成代码，还要按代码的标准来管理。

## 接入方自查清单

策略上线只是起点。下面这份清单，我拿来逐项对照自己的接入配置，每一项背后都对应一类真实的事故形态：

- **Key 治理**：按项目、按环境拆分，不共享不复用；消费额度设上限；定期轮换，疑似泄漏立即作废。
- **最小权限**：智能体进程跑在受限系统账号下；文件访问限定在沙箱目录；网络访问走显式白名单；高危动作必须人工确认。
- **审计**：每一次放行与拒绝都落日志；日志里不出现明文 Key 与敏感业务数据；日志本身有访问控制。
- **异常用量监测**：对步数、失败率、调用量、消费额建立基线，偏离即告警；深夜无人时段的流量突增单独设阈值。
- **停止规则**：步数预算、连续失败熔断、人工确认点，至少一项在每条执行路径上真正生效。
- **误杀监控**：拒绝率异常升高时回头修策略与白名单；安全策略像代码一样有版本、有评审、有回滚。

## 成本对比：安全措施的开销账

安全不是免费的，但账要算全。下表是量级估算，口径：以单步上下文约 2k tokens 的中型任务为参照，估算每项措施带来的额外开销；实际数字随模型、任务与实现方式浮动，量级比精确值更有参考意义。

| 安全措施 | 额外 token 开销 | 额外延迟 | 主要收益 | 主要代价 |
| --- | --- | --- | --- | --- |
| 系统提示内的自然语言策略 | 每请求 +300~800 | 接近零 | 越界提议率显著下降 | 常驻上下文，每请求都计费 |
| 动作白名单加路径与网络校验 | 零 | 小于 1ms，纯本地 | 越界动作到不了执行层 | 白名单需要持续维护 |
| 步数预算与熔断 | 零 | 可忽略 | 掐断失控循环，锁住成本上限 | 超长任务可能被提前终止 |
| 审计日志 | 零 | 每步约 1~5ms 写盘 | 全程可回溯、可定责 | 存储成本加脱敏工作量 |
| 独立安全评审调用（可选） | 每步 +500~1500 | 每步 +1~3s | 拦下伪装更好的越界动作 | 成本接近翻倍，延迟明显 |

从这张表能读出优先级：先把零 token、毫秒级的硬约束做满，再考虑按需开启评审调用。软约束的 token 开销看似不起眼，乘上请求量就是一笔常驻支出，策略文本值得像代码一样定期瘦身。延迟同样要进预算：串在治理链上的每次检查都在累加响应时间，链路越长，越要把便宜的检查放在前面，昂贵的检查按风险分层触发。

## 风险与合规提示：日志、保留期与误杀率

最后是三件容易后置的事。第一件是日志脱敏：审计日志天然汇聚敏感信息——业务数据、目录结构、偶发的凭证片段，落盘前先脱敏，明文 Key 一行都不许进日志；日志还要有独立的访问控制，如果读日志的人比用 Key 的人还杂，日志就从审计变成了新的风险面。

第二件是保留期限：日志无限积累既是存储负担也是合规负担，定一个明确期限，到期归档或清理，清理动作本身也要留痕。

第三件是误杀率：拒绝率是安全策略的健康指标，突升意味着策略过紧或白名单缺项，长期为零反而要警惕监测失灵。

合规边界同样要先讲清楚：这里讨论的全部内容只站在防御方一侧——合法接入、策略设计、监测与审计，管的是授权范围内的资产；不提供任何攻击方法，不碰武器与军事用途，也不鼓励绕过官方限制。安全工程的目标是让合法的使用更稳，让越界的使用更难，仅此而已。

## 写在最后

这一期的主线可以收拢成一句话：模型是冻结的，安全不必跟着冻结。策略 harness 的两层结构让软约束负责引导、硬约束负责兜底；联合搜索证明策略与规则成对优化能实实在在改善安全-效用权衡；过去近一年持续扩大的滥用面提醒我把 Key 当作一定会被盯上的资产；停止规则、审计日志、误杀监控这些不起眼的部件，才是长期运行智能体的底盘。

接入层面，我现在的默认路径是 4sapi（https://4sapi.com）的 OpenAI 兼容接口，模型可以随时换，策略代码却跟着业务越养越厚——这笔资产比任何一次单点接入都更值得经营。欢迎在评论区聊聊各自的策略设计。
