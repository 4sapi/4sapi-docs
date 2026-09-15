---
title: "【大模型API中转站】[第173期] 常驻 Agent 停止规则设计｜驱力外置防失控"
tags:
  - Agent
  - Harness
  - 停止规则
  - 成本控制
  - 状态管理
description: "常驻 Agent 没有天然的终点，把目标注入、重试预算、独立验证和停止规则从提示词搬进 harness，才能防住目标漂移与账单失控。"
---

# 【大模型API中转站】[第173期] 常驻 Agent 停止规则设计｜驱力外置防失控

智能体正在从一次性任务工具，变成保留重要状态、跨任务持续运行、能根据反馈自我调整的系统。真正困难的不是让它跑起来，而是让它知道什么时候该停。

## 一、常驻 Agent 没有天然的剧终

一次性任务 Agent 的失败模式很温和：任务挂了，重跑一次就行，代价是几分钟和几美分。常驻 Agent 完全不同——它保留重要状态、跨任务持续运行、根据反馈自适应调整行为。这种转变带来一个此前不突出的控制问题：目标、重试、验证、停止规则这些行为机制，目前基本靠 harness 手工实现，没有现成框架兜底。

我注意到失控的常驻 Agent 有四种典型姿态：

- 目标漂移：跑着跑着，任务悄悄换成了另一件事，原始目标没有人记得；
- 重试成瘾：同一个失败动作换着措辞反复尝试，失败计数从来不清零；
- 验证自嗨：生成通道自己给自己打分，永远"已完成"，永远不需要停；
- 永不停止：没有剧终条件，进程活着，循环就继续。

这四种姿态有个共同点：模型本身不会报警。最先知道出事的往往是账单和日志——某条调用曲线在深夜开始平缓爬坡，等人工看到时，一天的预算已经烧完。我维护 4sapi 这个大模型 API 中转站，从账单曲线侧看到的失控迹象，几乎总是早于日志告警。停止规则设计要解决的就是这件事：在账单说话之前，让代码先说话。

## 二、四种行为机制：从提示词搬进 harness

把行为机制从模型上下文里拿出来，放进 harness 的确定性代码，是常驻 Agent 治理的主线。四件事最关键：

**目标外置。** 目标不由模型自我延续，而由外部系统注入。模型每轮看到的目标，是 harness 从任务存储里读出来、带版本号写进去的。模型可以对目标提出修改建议，但建议生效需要外部确认，目标变更有记录、可审计。

**重试预算。** 次数、成本、时间三个维度各自设上限，任一维触顶就停。预算记在 harness 的账本里，而不是放在模型上下文里靠模型"记得"。

**验证独立。** 校验通道和生成通道分离。生成结果好不好，不由生成它的那个上下文自己说了算，而由独立校验逻辑判定——可以是另一路模型调用、规则检查、测试套件，或外部系统的真实状态。

**停止规则代码化。** 停机条件写成可枚举、可单元测试的代码，而不是提示词里一句"任务完成后请停止"。提示词会被注入、被稀释、被长上下文淹没，代码不会。

这四件事不是并列的补丁，而是一个整体：目标外置回答"为什么跑"，重试预算回答"能跑多久"，验证独立回答"算不算跑成了"，停止规则代码化回答"什么时候必须停"。

## 三、原理速览：一条带熔断的执行循环

把四个机制拼起来，常驻 Agent 的主循环长这样：

```text
外部任务存储（目标外置：目标带版本号，由外部系统写入）
        |
        v
harness 主循环：每轮读取当前目标与预算余量
        |
        v
  调用模型 / 工具（生成通道）
        |
        v
  独立校验通道判定结果（生成通道不参与打分）
        |
        v
  记账：步数 +1、金额累计、时间推进、失败计数更新
        |
        v
  检查点落盘（每步之后写状态文件，崩溃后从这里恢复）
        |
        v
  停止条件求值（可枚举、可单元测试）
        |
        +-- 目标完成（校验通过）--------------> 收尾：写终态检查点，任务标记 done
        |
        +-- 预算耗尽（次数/金额/时间任一触顶）--> 熔断：冻结状态，任务标记 fused
        |
        +-- 失败连击超限 / 心跳超时 ----------> 人工升级：通知值守，任务标记 blocked
        |
        +-- 均不满足 ------------------------> 回到主循环，进入下一轮
```

这条循环里有三个要点。第一，检查点贯穿每一轮，是每步之后的常规动作，不是出事之后才补写的救火脚本；崩溃恢复的时间下限，取决于检查点的写入频率。第二，停止条件求值发生在"再跑一轮"之前，而不是跑完之后复盘——任何一步都不允许绕过求值直接执行。第三，熔断和人工升级是两种不同出口：熔断是确定性动作，不需要任何人批准，触顶立即执行；升级是把决定权交回人，同时保留现场。

整条循环没有一处依赖模型的自觉。模型只负责生成，其余全是 harness 的确定性代码。判断一套常驻 Agent 设计是否可靠，第一问就是：停止决定由谁做出？答案是模型，就还没有出循环。

## 四、目标外置：让目标由外部系统注入

目标外置的具体做法：

1. 目标存放在外部系统——任务队列表、配置中心或数据库，每条目标带版本号；
2. harness 每轮开始时读取当前版本的目标，注入模型上下文；
3. 模型可以在输出里提出"目标应该调整为某方向"，harness 把它写成一条建议记录，不直接生效；
4. 目标变更走外部流程：人工确认或上游系统触发，变更历史可审计。

为什么不能让模型自我延续目标？因为常驻 Agent 的上下文里没有可靠的记忆仲裁者。上一轮的总结、上一轮的自我评价，都会被当成事实进入下一轮。漂移不是突变而是渐变：每轮偏一点，几十轮之后任务早已换了方向，而账单还在按原计划扣费。等人工从输出里看出不对劲，中间的调用已经全部浪费。

目标外置还有一个副产品：目标本身就是审计对象。事后复盘时能准确回答"第 N 版目标何时由谁写入"，而不是从模型输出里反推它当时自认为的目标。对需要向上解释成本去向的团队来说，这条审计链比多写十行防御性提示词有用得多。

## 五、重试预算：次数、成本、时间三维上限

单维预算都有漏洞。只限步数，模型可以选贵的模型跑满步数；只限金额，慢循环可以用廉价调用磨上一整天；只限时间，高并发可以把金额瞬间烧穿。三维一起限，任一维触顶即停：

- 次数：任务生命周期内的总调用步数，例如 200 步；
- 成本：按中转层计费口径累计的金额，例如 5 美元；
- 时间：墙钟时间，例如 1 小时，防止慢循环长期占用资源。

预算记在 harness 的账本里，每步之后更新并写入检查点。两个工程细节值得展开。其一，金额口径以中转层的实际计费为准，模型自报的 token 数只能作过程参考，费率换算放在 harness 一处完成，避免多处维护出现口径漂移。其二，预算余量是停止规则的输入，不是给模型看的提示——把"还剩 3 美元预算"写进上下文，等于邀请模型围绕预算优化措辞，而不是围绕任务干活。

生命周期预算之外，滚动窗口预算可以补位：例如"任意 1 小时窗口内不超过 100 步"。它专治平缓爬坡型失控——每一步看起来都正常，累计起来就是事故，只有窗口计数能提前发现这种形态。

## 六、验证独立：校验通道与生成通道分离

验证自嗨的根源是既当运动员又当裁判。生成通道说"任务已完成"，这句话在停止规则里不应有任何权重。校验要走独立通道：

- 规则检查：输出格式、必填字段、目标文件是否真的落盘，用代码判定；
- 测试套件：代码类任务跑真实测试，全绿才算通过；
- 外部状态：到工单系统里查工单是否真的存在，比让模型复述一遍靠谱得多；
- 独立模型调用：需要语义判断时，用另一个不带原始上下文的调用做评判，避免被生成过程的叙事带偏。

校验结果写进状态文件，成为停止规则的直接输入：校验通过且目标达成，才允许走"达标收尾"出口。这里有一个反直觉的取舍：独立校验本身也是成本，每步多一次调用，预算大约翻倍。省这笔钱的正确方式不是砍掉校验，而是降低校验频率——例如每 5 步全量校验一次，关键节点强制校验。砍掉校验省下的是小钱，赔上的是"永不停止"这个大坑。

## 七、检查点与恢复：状态文件里该写什么

状态文件是常驻 Agent 的存档，最小字段集：

- 任务状态：running / blocked / done / failed / fused；
- 当前目标版本号与目标摘要；
- 预算余量：已用步数、已用金额、起始时间；
- 已完成步数与最近若干步的摘要；
- 最近校验结果：通过与否、时间、判定依据摘要；
- 失败连击计数：连续未通过校验的次数；
- 心跳时间戳：最后一次成功写检查点的时间。

断点续跑和干净重启各有代价。断点续跑省钱省时间，但会继承可能已被污染的状态——如果漂移发生在第 30 步，第 31 步续跑只是从污染点继续跑。干净重启逻辑简单、状态干净，但要重做全部已完成工作。实践中的折中是：恢复时先做 schema 校验和基本一致性检查，不过关就自动降级为干净重启；失败连击次数过高时，即使状态文件完好，也优先人工介入而不是无脑续跑。

写入本身要用原子替换——先写临时文件再 rename，避免进程在写一半时崩溃留下损坏的状态文件。这个细节在长驻进程里不是洁癖，是保命：状态文件一旦损坏，所有预算和进度同时归零，等于从未设防。

## 八、接入教程：Python 常驻 Agent 守护骨架

下面是一个可运行的守护骨架，覆盖检查点、预算账本、停止规则求值和崩溃恢复。模型调用走 4sapi（https://4sapi.com）的 OpenAI 兼容接口，只改 base_url 就能换成其他兼容端点；中转层可挂用量监控，账单曲线和阈值告警都在中转层完成，比在 Agent 进程里自报成本低一个信任层级。

```python
import json
import os
import time
import tempfile
from dataclasses import dataclass, field, asdict
from openai import OpenAI

# 模型调用走中转站的 OpenAI 兼容接口；密钥从环境变量读取，不进代码库
client = OpenAI(
    api_key=os.environ["MIDDLEMAN_API_KEY"],
    base_url="https://4sapi.com/v1",
)
MODEL = os.environ.get("AGENT_MODEL", "default-model")
CHECKPOINT = "agent_state.json"

SCHEMA_VERSION = 1
MAX_STEPS = 200            # 次数维度上限
MAX_COST_USD = 5.0         # 成本维度上限，按中转层计费口径
MAX_WALL_SECONDS = 3600    # 时间维度上限
FAILURE_STREAK_LIMIT = 5   # 失败连击阈值
HEARTBEAT_TIMEOUT = 600    # 心跳超时（秒）：超时未写检查点视为假死
UNIT_PRICE_USD = 0.000002  # 演示单价：每 token 成本，以中转层账单为准


@dataclass
class Budget:
    used_steps: int = 0
    used_cost_usd: float = 0.0
    started_at: float = field(default_factory=time.time)

    def exhausted(self) -> bool:
        # 三维预算：任一维触顶即视为耗尽
        return (
            self.used_steps >= MAX_STEPS
            or self.used_cost_usd >= MAX_COST_USD
            or time.time() - self.started_at >= MAX_WALL_SECONDS
        )


@dataclass
class TaskState:
    task_id: str
    phase: str = "running"        # running / blocked / done / failed / fused
    goal_version: int = 1
    completed_steps: int = 0
    failure_streak: int = 0
    last_check: str = "pending"   # 最近校验结果：passed / failed / pending
    last_heartbeat: float = field(default_factory=time.time)
    budget: Budget = field(default_factory=Budget)
    schema_version: int = SCHEMA_VERSION


def save_state(state: TaskState) -> None:
    # 原子写入：先落临时文件再 rename，防止写一半崩溃损坏状态文件
    payload = json.dumps(asdict(state), ensure_ascii=False, indent=2)
    fd, tmp = tempfile.mkstemp(dir=os.path.dirname(CHECKPOINT) or ".")
    with os.fdopen(fd, "w", encoding="utf-8") as f:
        f.write(payload)
    os.replace(tmp, CHECKPOINT)
    state.last_heartbeat = time.time()


def load_state() -> TaskState | None:
    # 崩溃恢复：状态文件存在且 schema 匹配才续跑，否则返回 None 走干净重启
    if not os.path.exists(CHECKPOINT):
        return None
    with open(CHECKPOINT, encoding="utf-8") as f:
        raw = json.load(f)
    if raw.get("schema_version") != SCHEMA_VERSION:
        return None
    budget = Budget(**raw.pop("budget"))
    return TaskState(**raw, budget=budget)


def call_model(goal: str, goal_version: int, step_summary: str) -> tuple[str, float]:
    # 单步调用；成本按中转层返回的用量换算，费率集中在此处维护
    resp = client.chat.completions.create(
        model=MODEL,
        messages=[
            {"role": "system", "content": "常驻任务执行器，严格围绕注入的目标工作。"},
            {
                "role": "user",
                "content": f"目标 v{goal_version}：{goal}\n进展：{step_summary}",
            },
        ],
    )
    text = resp.choices[0].message.content or ""
    cost = resp.usage.total_tokens * UNIT_PRICE_USD
    return text, cost


def validate(output_text: str) -> bool:
    # 独立校验通道：规则检查，不用生成通道自评
    # 真实系统替换为测试套件、外部状态查询或独立模型评判
    return bool(output_text.strip()) and "ERROR" not in output_text


def notify_oncall(task_id: str, reason: str) -> None:
    # 人工升级通道：替换为真实告警（IM / 邮件 / 工单）
    print(f"[ESCALATE] task={task_id} reason={reason}")


def evaluate_stop(state: TaskState) -> tuple[bool, str, bool]:
    # 停止条件求值：返回（是否停止、原因、是否人工升级）
    # 每条规则可枚举、可单元测试
    if state.last_check == "passed":
        return True, "goal_done", False            # 目标完成：校验通道判定通过
    if state.budget.exhausted():
        return True, "budget_exhausted", False     # 预算耗尽：熔断，无需批准
    if state.failure_streak >= FAILURE_STREAK_LIMIT:
        return True, "failure_streak", True        # 失败连击：升级给人处理
    if time.time() - state.last_heartbeat > HEARTBEAT_TIMEOUT:
        return True, "heartbeat_timeout", True     # 心跳超时：进程假死
    return False, "", False


def run(goal: str, goal_version: int) -> TaskState:
    state = load_state()
    if state is not None and state.phase != "running":
        # 非运行态任务不自动续跑，等待人工放行或重新下发目标
        return state
    if state is None:
        state = TaskState(
            task_id=f"task-{int(time.time())}", goal_version=goal_version
        )
    while True:
        stop, reason, escalate = evaluate_stop(state)
        if stop:
            if escalate:
                state.phase = "blocked"
                notify_oncall(state.task_id, reason)  # 保留现场，等待人工
            elif reason == "goal_done":
                state.phase = "done"
            else:
                state.phase = "fused"                 # 超限熔断：冻结状态
            save_state(state)
            break
        summary = f"已完成 {state.completed_steps} 步，失败连击 {state.failure_streak}"
        output, cost = call_model(goal, goal_version, summary)
        state.budget.used_cost_usd += cost
        state.budget.used_steps += 1
        state.completed_steps += 1
        if validate(output):
            state.last_check = "passed"
            state.failure_streak = 0
        else:
            state.last_check = "failed"
            state.failure_streak += 1
        save_state(state)  # 每步之后落检查点，崩溃后从这里恢复
    return state
```

骨架里有三个容易做错的点：

- 恢复语义：load_state 返回 None 时走干净重启，而不是带着损坏状态硬跑；schema 变更时旧状态文件要么迁移要么作废，不要静默兼容；
- 成本口径：call_model 里的单价常量只是演示，生产上以中转层账单校准费率，模型自报的 token 数只做过程参考；
- 告警通道：notify_oncall 必须接到真实通知渠道，熔断之后没有人知道，等于没有熔断。

用量监控建议直接挂在中转层：进程内账本可能跟着进程一起说谎——崩溃时漏账、重启时清零；中转层看到的是每一次真实请求，账单曲线就是最诚实的执行日志。

## 九、上线自查清单

常驻 Agent 放开预算之前，把这份清单过一遍，任何一项不过关都先补齐：

- [ ] 状态可恢复：杀掉进程重启，能从检查点继续，schema 校验通过；
- [ ] 干净重启路径可用：删掉状态文件后能从零开始，不依赖残留状态；
- [ ] 预算可观测：步数、金额、时间三维余量在监控面板可见，不必登录机器翻文件；
- [ ] 停止条件有单元测试：每条 StopRule 至少一个触发用例、一个不触发用例；
- [ ] 人工升级通道真实存在：告警能送到人手上，而不是打进没人看的日志；
- [ ] 账单告警：中转层或云账单侧配置阈值告警，账单曲线可按 key 追溯到具体 Agent；
- [ ] 目标变更可审计：每版目标有写入人、写入时间和变更原因；
- [ ] 状态文件脱敏与访问控制：文件权限收敛到运行账号，敏感字段不落明文。

## 十、成本对比：三种配置的账单风险（估算）

下面的数字全部是演示口径：假设单步调用平均成本 0.02 美元，三维预算按 200 步 / 5 美元 / 1 小时设置，失控场景按夜间 8 小时无人值守、每 30 秒一步估算。数字用于说明封顶逻辑，不是实测值，接入时按自己的费率重算。

| 方案 | 额外工程成本 | 失控账单上限（估算） | 误杀长任务风险 | 适用场景 |
| --- | --- | --- | --- | --- |
| 无停止规则 | 几乎为零 | 上不封顶：8 小时约 960 步、19 美元，随窗口时长线性增长 | 无误杀，但可能永不停止 | 一次性短任务，有人盯着 |
| 三维预算 | 低：账本加求值器 | 5 美元封顶（约 250 步触顶） | 中：预算过紧会截断合法长任务 | 无人值守常驻任务 |
| 预算加人工确认点 | 中：通知与放行通道 | 5 美元封顶，确认点再加一道闸 | 低：熔断后可人工放行续跑 | 高成本或高价值任务 |

两个读表要点。其一，预算的价值不是省钱，而是把"最多损失多少"从无穷大变成一个已知数——19 美元的演示数字不起眼，乘上团队里几十个常驻 Agent 和更高的单步成本，量级就完全不同。其二，人工确认点会增加任务延迟，但换来误杀可挽回：对长任务，"熔断后等待放行"几乎总是优于"直接判死"。

## 十一、风险与合规提示

- 停止规则过紧：合法长任务被误杀，表现为任务反复熔断、人工反复放行。解法是分级预算——探索阶段宽松、执行阶段收紧，而不是一刀切；
- 停止规则过松：熔断阈值形同虚设。预算上限按"睡一觉醒来也能承受"的口径设定，而不是按"理想情况刚好用完"设定；
- 状态文件含敏感业务数据：目标摘要、中间产出都可能带业务信息。脱敏后再落盘，文件权限收敛到运行账号，并纳入备份与清理策略；
- 停止规则不要全塞进提示词：提示词会被注入、被长上下文稀释，"请适时停止"这类软约束在失控场景下第一个失效。确定性控制必须在 harness 代码里，提示词最多作为补充；
- 接入边界：中转与常驻都要在服务商条款允许的范围内进行。限流、配额和审计是对接入方的保护机制，监测与熔断的目的是让 Agent 在框架内可靠运行，任何绕过限制的思路都与这套治理目标背道而驰。

## 十二、总结

常驻 Agent 把智能体从"跑一次"变成"一直跑"，也把控制问题从可有可无变成生死攸关。目标外置让目标由外部系统注入、可审计；重试预算用次数、成本、时间三维上限封住损失；验证独立让完成与否由校验通道说了算；停止规则代码化让每一条停机条件可枚举、可测试。检查点贯穿每一轮，状态文件写全任务状态、预算余量、已完成步和最近校验结果，恢复时先校验再续跑。示例骨架的模型调用走 4sapi（https://4sapi.com）的 OpenAI 兼容接口，用量监控挂在中转层，账单曲线就是最诚实的执行日志。停止规则的价值不在于让 Agent 跑得更顺，而在于让"最多损失多少"变成一个写进代码的已知数。欢迎在评论区聊聊，常驻 Agent 的停止规则设计里，最贵的一次学费交在了哪条规则上。
