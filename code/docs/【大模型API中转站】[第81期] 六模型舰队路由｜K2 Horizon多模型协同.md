---
title: "【大模型API中转站】[第81期] 六模型舰队路由｜K2 Horizon多模型协同"
tags:
  - 多模型路由
  - K2 Horizon
  - 模型舰队
  - API网关
description: "开放模型生态已足够丰富，把多个模型编排成互联舰队正成为接入层标配。用统一网关做路由、级联、回退与预算分配，附 JSON 路由配置与 Python 实现。"
---

# 【大模型API中转站】[第81期] 六模型舰队路由｜K2 Horizon多模型协同

单模型打天下的时代正在过去，多模型协同正在成为接入层的主流做法。

有团队发布了 K2 Horizon——一个由六个开放模型组成的互联舰队（connected fleet of six open models），强调让多个开放模型协同工作而不是单打独斗，这个思路在技术社区获得高热度。这一期把舰队路由拆开讲：为什么模型要组队、网关怎么编排、级联与回退怎么写、预算怎么分。

## 一、开篇：单模型不够用的时候

接入大模型 API 的开发者迟早会遇到同一个困境：一个模型解决不了所有问题。同一个项目里，代码补全要快，长文生成要便宜，中文对话要自然，还有一部分请求因为数据敏感必须留在本地。把所有需求压给一个模型，结果往往是速度、成本、质量三头都不讨好。

更麻烦的是单点故障：模型服务限流了、升级了、改行为模式了，整个应用跟着受影响。服务商关停或政策调整这类风险，在上一期讨论 API 依赖生命周期时已经确认过：单一依赖是接入层最大的隐患。

多模型舰队正是为了同时解决这两件事：用路由把每条请求送到最合适的模型，用级联和回退把单点风险摊到多个成员上。舰队落地的第一块地基是统一网关，4sapi（https://4sapi.com）这类入口可以把多供应商收拢成一个端点。

## 二、K2 Horizon：六个开放模型的互联舰队

K2 Horizon 的核心主张是"connected fleet of six open models"——六个开放模型互联成一支舰队。与传统的"选一个最强模型"思路不同，舰队强调的是分工与协同：每个模型保留自己的强项，网关在它们之间调度，必要时一个模型接另一个模型的活。

这个方向能成立，前提是开放模型生态已经足够丰富。Qwen、GLM、Llama、Mistral 等系列各有擅长，有的中文扎实，有的代码强，有的参数效率高，有的生态配套全。模型本身不再是稀缺资源，稀缺的是把它们组织起来的接入层能力。

K2 Horizon 在技术社区获得高热度，本质上是因为它把"多模型协同"从零散的工程技巧提升成了明确的架构主张：接入层不再适配单个模型，而是编排一组模型。这一期我围绕这个主张，给出一套可以直接落地的网关设计与代码。

## 三、原理速览：舰队的三层协同

模型舰队按协同深度可以分三层：

```text
第一层：路由 —— 请求按规则分发给最合适的模型
第二层：级联 —— 主模型结果不达标时，交给专门模型接力
第三层：回退 —— 主模型故障、限流或超时，流量转移到备用模型
```

三层叠加起来，一条请求的完整治理链是：

```text
请求进入网关
    |
路由匹配（按任务 / 成本 / 延迟 / 数据等级）
    |
预算与配额检查
    |
调用主模型（带超时与并发限制）
    |
    +-- 成功且通过质量检查 -> 返回
    |
    +-- 质量不达标 -> 级联到专门模型 -> 返回
    |
    +-- 失败 / 超时 / 限流 -> 触发熔断 -> 回退到备用模型 -> 返回
    |
    +-- 全部成员失败 -> 返回降级结果或错误，记录审计日志
```

关键点在于：每一步都是确定性规则，不是模型的自由发挥。路由规则、质量判据、回退顺序都由接入层显式定义，行为可预测、可测试、可审计。

## 四、开放模型生态盘点：舰队成员从哪来

舰队质量的上限取决于成员搭配。开放模型生态当前足够支撑一支像样的舰队：

| 模型系列 | 擅长方向 | 舰队中的角色 |
| --- | --- | --- |
| Qwen 系列 | 通用能力、中文、数学 | 主力通用模型 |
| GLM 系列 | 中文理解、工具调用 | 对话与 Agent 主力 |
| Llama 系列 | 生态配套、社区资源 | 定制与私有部署 |
| Mistral 系列 | 参数效率、轻量推理 | 低成本高频任务 |
| 专用小模型 | 分类、抽取、总结 | 级联中的专门角色 |

搭配原则有三条：一是能力互补，避免六个模型擅长同一件事；二是成本拉开，便宜模型承接大流量、贵模型只处理关键任务；三是部署形态多样，托管源与本地源并存，敏感数据永远走本地。

## 五、路由规则设计：按任务、成本、延迟分发

路由是舰队的指挥层。规则用 JSON 描述，方便版本管理与灰度上线。一个最小可用的路由配置长这样：

```json
{
  "routes": [
    {
      "name": "interactive_chat",
      "match": {"task": "chat", "latency": "low"},
      "target": {"model": "glm-chat", "source": "api"},
      "fallback": ["qwen-chat", "local-7b"]
    },
    {
      "name": "long_writing",
      "match": {"task": "writing", "max_tokens": 4000},
      "target": {"model": "qwen-long", "source": "api"},
      "fallback": ["mistral-long"]
    },
    {
      "name": "sensitive",
      "match": {"data_class": "private"},
      "target": {"model": "local-7b", "source": "local"},
      "fallback": []
    }
  ],
  "budget": {
    "daily_usd_cap": 50,
    "high_cost_ratio": 0.2
  }
}
```

每条规则包含四个要素：匹配条件（match）、目标模型（target）、回退链（fallback）与来源（source）。匹配条件按任务类型、延迟要求、数据等级正交划分，避免规则互相覆盖；回退链按顺序尝试，全部失败才降级。

## 六、Python 实现：统一网关的路由与并发

把上面的 JSON 规则变成可运行的 Python 网关客户端。核心是三个部分：加载规则、匹配路由、按来源并发调用。

```python
import json
import asyncio
import aiohttp
from openai import AsyncOpenAI

class FleetGateway:
    def __init__(self, config_path: str, clients: dict):
        with open(config_path, encoding="utf-8") as f:
            self.config = json.load(f)
        self.clients = clients  # {"api": AsyncOpenAI(...), "local": AsyncOpenAI(...)}
        self.semaphore = asyncio.Semaphore(16)  # 全局并发上限

    def match_route(self, req: dict) -> dict:
        for route in self.config["routes"]:
            if all(req.get(k) == v for k, v in route["match"].items()):
                return route
        raise ValueError("no route matched: " + json.dumps(req))

    async def call_model(self, source: str, model: str, messages: list) -> str:
        client = self.clients[source]
        async with self.semaphore:
            resp = await client.chat.completions.create(
                model=model, messages=messages, stream=True
            )
            parts = []
            async for chunk in resp:
                delta = chunk.choices[0].delta.content
                if delta:
                    parts.append(delta)
            return "".join(parts)

    async def ask(self, req: dict) -> str:
        route = self.match_route(req)
        target = route["target"]
        return await self.call_model(target["source"], target["model"], req["messages"])
```

并发上限用信号量控制，防止突发流量同时打爆所有成员；流式返回让首 token 尽早到达，交互体感更好。这个骨架可以继续加超时、重试与指标上报。

## 七、级联与回退：主模型 + 专门模型、故障转移

路由只解决"谁先上"，级联与回退解决"答不好怎么办、挂了怎么办"。

级联的典型形态是主模型先答，质量检查不达标时交给专门模型重写或补全：

```text
主模型（如通用对话） -> 质量检查（分类 / 打分）
    +-- PASS -> 直接返回
    +-- FAIL -> 级联到专门模型（如代码专家 / 总结模型）-> 返回
```

实现上，质量检查本身可以是一个轻量分类模型，只输出 PASS/FAIL，成本很低。回退则在主模型异常时按 fallback 顺序尝试备用模型：

```python
async def ask_with_cascade(self, req: dict) -> str:
    route = self.match_route(req)
    primary = route["target"]
    try:
        text = await self.call_model(primary["source"], primary["model"], req["messages"])
    except Exception:
        return await self.fallback(req, route)
    if await self.quality_check(text) == "FAIL":
        return await self.call_model("api", "specialist-summary", req["messages"])
    return text

async def fallback(self, req: dict, route: dict) -> str:
    errors = []
    for name in route.get("fallback", []):
        source, model = name.split("/")
        try:
            return await self.call_model(source, model, req["messages"])
        except Exception as e:
            errors.append(f"{name}: {e}")
    raise RuntimeError("all models failed: " + "; ".join(errors))
```

回退链的设计原则是收敛：每一层都只尝试一次，不无限重试；全部失败就返回明确的降级结果，而不是让请求卡在超时里。

## 八、熔断与预算分配：别让单点拖垮舰队

舰队也有脆弱环节：某个成员突然变慢或报错，如果每次都等它的超时，整条链路都会被拖住。熔断器解决这个问题：连续失败达到阈值，就暂时把该成员摘出路由，让流量绕行，等它恢复后再逐步放回。

```python
class CircuitBreaker:
    def __init__(self, threshold: int = 3, cooldown: float = 30.0):
        self.threshold = threshold
        self.cooldown = cooldown
        self.failures = 0
        self.open_until = 0.0

    def allow(self, now: float) -> bool:
        if now < self.open_until:
            return False
        return True

    def record_failure(self, now: float):
        self.failures += 1
        if self.failures >= self.threshold:
            self.open_until = now + self.cooldown
            self.failures = 0

    def record_success(self):
        self.failures = 0
```

预算分配是另一道闸门。舰队里每个成员单价不同，级联和回退还会产生额外调用，账单很容易失控。预算层按三档控制：

- 任务级：高价值任务允许用贵模型，普通任务只走便宜模型；
- 源级：每个来源设每日 token 与金额上限，超限自动降级；
- 账户级：网关整体设日预算，接近上限时熔断非关键请求。

```text
预算闸门
    |
    +-- 任务价值判断（关键 / 普通 / 批量）
    |
    +-- 来源配额检查（每个供应商的额度）
    |
    +-- 全局日预算检查（超限 -> 降级或拒绝）
```

三层闸门叠加，舰队才能在"想用贵模型"和"不能超预算"之间稳定运行。

## 九、接入教程：用 4sapi 把舰队收拢成一个入口

舰队背后的模型来自多个供应商，每家一套 Key、一套端点、一套计价方式，直接对接会变成配置地狱。我的做法是用 4sapi（https://4sapi.com）作为统一网关，把所有来源收拢成一个 OpenAI 兼容入口。

接入分四步：

1. 在 4sapi（https://4sapi.com）获取网关 Key 与端点地址；
2. 在网关侧配置各模型的后端映射，例如把 `qwen-chat` 映射到 Qwen 官方源、把 `local-7b` 映射到本地部署；
3. 客户端把 base_url 指向 4sapi，模型名使用网关侧的别名；
4. 把上一节的路由 JSON 与网关 Key 一起加载进应用。

```text
应用 -> 4sapi 网关 -> Qwen / GLM / Llama / Mistral 等多供应商源
                  -> 本地自建模型（敏感数据不出内网）
```

统一入口带来的直接好处：Key 只在一个地方管理，账单合并对账，模型切换不改应用代码。路由、级联、回退这些舰队逻辑都在应用层或网关层完成，接入侧始终保持 OpenAI 兼容，迁移成本几乎为零。

## 十、Python 完整示例：把舰队跑起来

把路由、级联、回退、熔断、预算组装成一个可运行的入口：

```python
import json
import asyncio
from openai import AsyncOpenAI

async def main():
    clients = {
        "api": AsyncOpenAI(api_key="4sapi_key", base_url="https://4sapi.com/v1"),
        "local": AsyncOpenAI(api_key="local_key", base_url="http://127.0.0.1:8000/v1"),
    }
    gateway = FleetGateway("fleet_routes.json", clients)

    req = {"task": "chat", "latency": "low",
           "messages": [{"role": "user", "content": "帮我设计一个用户登录流程"}]}
    try:
        answer = await gateway.ask_with_cascade(req)
        print(answer)
    except Exception as e:
        print(f"fleet degraded: {e}")

asyncio.run(main())
```

生产环境还要补三样：一是所有调用写入审计日志（模型、来源、耗时、费用估算）；二是每个成员的关键指标上报（延迟、错误率、熔断状态）；三是路由规则变更走灰度，先在影子流量上验证再全量切换。

## 十一、成本与风险提示：多供应商计费与复杂度

舰队不是免费午餐，成本和复杂度是实打实的：

- 多供应商计费：每家计价口径不同，级联与回退会产生额外调用，对账必须按"路由规则 × 来源 × 模型"拆分，否则超支都找不到源头；
- 复杂度上升：路由、级联、回退、熔断、预算五层逻辑叠加，调试与测试面显著变大，需要配套的测试夹具与影子流量；
- 可观测性要求：舰队要同时观测所有成员，指标维度成倍增加，日志与监控的投入不能省；
- 行为一致性：不同模型输出风格不同，级联后文本风格可能漂移，需要统一的后处理或风格约束；
- 合规边界：这一期只讨论合法接入与架构设计——通过官方 API 或自有网关接入开放模型，本地模型用于敏感数据，不涉及绕过任何服务商的限制。

风险管理的核心原则：每一层逻辑都要有开关和兜底。路由可以退化为单模型直连，级联可以关闭，回退失败要返回明确错误而不是静默降级。

## 十二、舰队上线前的检查清单

把模型舰队接入生产之前，我按这份清单过一遍：

1. 路由规则是否覆盖所有已知任务类型，是否有匹配不到的请求会抛错；
2. 每条规则的回退链是否收敛，全部失败时是否返回明确降级结果；
3. 熔断阈值与冷却时间是否按成员分别配置；
4. 预算三层闸门是否启用，超限后是否自动降级而非硬失败；
5. 级联质量检查的判据是否可解释、可抽样验证；
6. 敏感数据路由是否强制走本地源，是否有规则能覆盖；
7. 所有调用的审计日志是否包含模型、来源、耗时与费用估算；
8. 路由规则变更是否有灰度与回滚方案；
9. 多供应商账单是否可按路由规则 × 来源拆分对账；
10. 是否保留了"一键退回单模型"的应急开关。

## 结论

K2 Horizon 把"六个开放模型互联成舰队"的想法摆到了台面上，而它真正触动的，是接入层正在从"适配单模型"转向"编排一群模型"这一事实。路由让每条请求找到最合适的模型，级联让主模型答不好的任务有人接力，回退让单点故障不再拖垮整条链路，熔断与预算让舰队在失控之前自动收敛。

模型是资源，路由是策略，网关是骨架。用 4sapi（https://4sapi.com）收拢多供应商入口，用 JSON 定义路由，用 Python 实现级联回退，舰队就能在可控成本内稳定运转。欢迎在评论区发表想法。
