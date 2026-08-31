---
title: "【大模型API中转站】[第38期] ChatGPT Work 双形态解析｜云桌面接入差异实测"
tags:
  - ChatGPT
  - API 接入
  - 模型选型
  - 智能体
description: "拆解 ChatGPT Work 云端与桌面两个形态的能力边界，对比产品直用与 API 编程两类接入路径，实测 Sol/Luna/Terra 模型矩阵与推理级别的成本差异。"
---

# 【大模型API中转站】[第38期] ChatGPT Work 双形态解析｜云桌面接入差异实测

OpenAI 在 7 月 9 日发布 ChatGPT Work，随后一周内我把它拆开实测了一遍：它不是一个产品，而是云端（Work Cloud）与桌面（Work Local）两个形态，且只对 $20/月及以上订阅开放。

我记录这次双形态接入实测：Work 相比 Chat 到底多了什么、两套接入路径（产品直用 vs API 编程）的差异在哪、以及把 Work 场景搬到代码里时，模型矩阵和推理级别应该怎么选。作为长期做 API 接入的开发者，我全程对照 [4sapi](https://4sapi.com) 的 OpenAI 兼容接口验证参数行为，确保结论可以直接复用到自己的接入代码里。

## 一、开篇：我为什么盯着 ChatGPT Work 不放

过去半年我一直在做大模型 API 接入和成本优化，ChatGPT 的每一次产品升级我都会同步对照一次 API 侧的模型清单。

Work 发布当天，我第一反应是「这不就是把 Codex 换个名字」。真正上手之后才发现判断错了：Work 在云端跑了一整套完整执行环境，聊天只是它的门面。它和 Chat 在能力集上是两个物种，而官方文档偏偏把这层差异藏得很深，很容易让开发者用 Chat 的思维去选 Work，最后既踩模型矩阵的坑，又对不上成本。

我实测的目标只有一个：把 Work 的能力边界、模型选择逻辑和两类接入方式讲清楚，让做产品接入和做 API 接入的人都能直接对号入座。

## 二、ChatGPT Work 到底是什么：一次发布，两个产品

先说结论：ChatGPT Work 表面是一个名字，实际上是两个产品，官方文档没有明确区分，这才是「混乱」的根源。

### 云端版 Work Cloud

跑在 OpenAI 的云端，通过 chatgpt.com 网页端或 ChatGPT 移动 App 进入。代码执行、文件系统、浏览器都运行在远端沙箱里，我换了设备、断了网络，任务状态依然在云端继续。

### 桌面版 Work Local

桌面 App 形态，前身是 Codex 的桌面应用。最大的区别是它直接操作本机：读取本地文件、在本机运行程序。实测感受是它更像「给非程序员软化过的 Codex」，权限粒度比云端版更贴近本机环境。

两个形态的订阅门槛一致，只有 $20/月及以上订阅能用，Free 和 $8/月的 Go 订阅都不开放：

| 维度 | Work Cloud | Work Local |
| --- | --- | --- |
| 入口 | chatgpt.com / 移动 App | ChatGPT 桌面应用 |
| 运行位置 | 云端沙箱 | 本机 |
| 文件系统 | 云端持久化目录 | 本机文件系统 |
| 代码执行 | 云端执行 | 本机执行 |
| 依赖网络 | 强依赖，断网即断 | 本地可用，模型仍走云端 |
| 典型场景 | 无人值守任务、发布 | 本地开发、文件批处理 |

对做 API 接入的我来说，Work Cloud 更有参考价值：它的执行沙箱、持久化目录和模型路由，几乎就是一套「产品化的 Agent 运行时」。

## 三、Work 与 Chat 的边界：不是增强，是分叉

官方给的区分口径是：Chat 用于要答案、解释、头脑风暴和短草稿；Work 用于要明确产出的任务，比如简报、演示、分析、周期性更新、工作流和可交付的文件。

这个口径方向对，但太模糊。我在 Chat 里干了几年这种活，靠这句话根本分不清该用哪个。真正的分界线在能力集：

| 能力 | Chat | Work |
| --- | --- | --- |
| 对话与问答 | 有 | 有 |
| 联网代码执行 | 无 | 有（云端沙箱） |
| 无头 Chrome 浏览器 | 无 | 有 |
| 跨会话持久化文件系统 | 无 | 有 |
| 发布 ChatGPT Sites | 无 | 有 |
| Sol/Luna/Terra 子智能体会话 | 无 | 有 |
| 定时提示自动化 | 部分 | 有 |

一句话总结我实测的结论：Chat 是一问一答的对话界面，Work 是带工具、带存储、带发布能力的任务执行平台。前者解决「想清楚」，后者解决「做完」。

## 四、Work 独有能力全清单：实测逐项验证

这六项是我在 Work 里逐项验证过的独有能力，每一项都直接对应 API 侧的架构设计点：

1. **联网代码执行**：云端沙箱里跑 Python、Shell，能访问公网。实测用它抓取并解析了网页数据，延迟在可接受范围，失败时会自动报错重试。
2. **无头 Chrome 浏览器**：工作区里内置浏览器实例，能打开页面、执行 JS、等待渲染。对需要登录态和动态渲染的站点，这是 Chat 完全给不了的能力。
3. **持久化文件系统**：会话之间共享文件夹，上次任务留下的中间文件，下次会话直接继承。这解决了我做多步抓取时的状态传递问题。
4. **发布 ChatGPT Sites**：生成的页面可以发布为公开站点，产出物直接可分发。
5. **Sol/Luna/Terra 子智能体会话**：Work 可以把任务拆给多个子智能体并行跑，各自占用独立推理上下文，最后汇总。
6. **定时提示自动化**：设定时间自动触发任务，把「周期性更新」从人肉操作变成定时任务。

清单里 1、2、3、5 这四项，正好对应 API 接入时的代码沙箱、浏览器工具、对象存储和多 Agent 编排四个组件。产品把这些封装好了，自己接 API 就要一个个自己搭。

## 五、模型矩阵实测：Sol/Luna/Terra 与推理级别

Work 的模型选择和 Chat 完全不同，这是我最意外的地方。Work 里可以显式选 GPT-5.6 的 Sol、Luna、Terra 三个变体，每个变体都能配上 Light、Medium、High、Extra High、Max、Ultra 六个推理级别；另外也能用 GPT-5.5，但推理级别只开放到 Extra High。

| 模型 | Work 可选 | 推理级别范围 |
| --- | --- | --- |
| GPT-5.6 Sol | 是 | Light / Medium / High / Extra High / Max / Ultra |
| GPT-5.6 Luna | 是 | Light / Medium / High / Extra High / Max / Ultra |
| GPT-5.6 Terra | 是 | Light / Medium / High / Extra High / Max / Ultra |
| GPT-5.5 | 是 | Light / Medium / High / Extra High |

Chat 那边的矩阵则是另一套：5.6 Instant、Medium、High、Extra High、Pro，其中 Extra High 和 Pro 只对 $100/月以上订阅开放，$20/月最高到 High，且 5.6 Pro 只在 Chat 出现，Work 没有。

这套矩阵在 API 侧同样存在。实测同一道数据分析任务，Sol 的 High 和 Ultra 在正确率上差距不如 Token 消耗差距明显，所以我自己的经验是：粗筛用 Medium/High，交出最终成果前再跑一次 Ultra，成本能省下一截。这个「分阶段升级推理级别」的思路，直接在请求参数里实现即可。

## 六、原理速览：一次 Work 任务的请求流

把 Work Cloud 当成黑盒拆开，一次任务的实际流向是这样的：

```text
用户在 Chat 标签选择 Work
        |
        v
Work 编排层（会话上下文 + 持久化文件系统）
        |
        v
模型路由（按所选模型：GPT-5.6 Sol/Luna/Terra + 推理级别）
        |
        +--> 工具层（可选，按需调用）
        |      联网代码执行
        |      无头 Chrome 浏览器
        |      文件读写
        |      ChatGPT Sites 发布
        |
        +--> 子智能体层（可选）
        |      Sol/Luna/Terra 派生并行会话
        |
        v
产出物（文件 / 分析报告 / 已发布的站点）
```

这个流程对 API 接入的启示很直接：产品层面的「Work」本质是「LLM + 工具 + 存储 + 编排」的组合。产品把编排层做好了，我用自己的代码复制这套结构时，每个组件都可以用公开 API 拼出来：模型调用走 chat/completions，工具执行自己写沙箱，持久化用对象存储，子智能体用并发会话。

## 七、方案一：产品直用（订阅接入）

如果目标是「今天就要用」，最快路径是订阅接入，什么都不用写：

1. 把订阅升到 $20/月及以上（Free 和 $8/月 Go 无法进入 Work）。
2. 在 chatgpt.com 或移动 App 的标签选择器切到 Work。
3. 选择模型（GPT-5.6 Sol 默认）与推理级别（初始建议 High）。
4. 描述要完成的任务，等待云端执行，检查产出文件或已发布站点。

实测体验：产品直用的优点是无运维、沙箱和浏览器都是现成的；缺点是自动化能力受限——定时任务要依赖产品内的编排，无法把 Work 的完整执行链路嵌进我的服务里，多用户场景下也只能按账号订阅，成本随人数线性上涨。

## 八、方案二：API 编程接入（把 Work 场景搬进代码）

对要做产品集成的我来说，第二步才是关键：Work 里出现的那批模型（GPT-5.6 Sol/Luna/Terra、GPT-5.5）在 OpenAI API 侧同样存在，只是产品把参数打包成了下拉框。想要可控的自动化，就直接用 API 自己组装执行链路。

我自己的接入走的是 [4sapi](https://4sapi.com)：OpenAI 兼容格式，模型、推理级别、流式输出这些参数都能原样透传，省掉直连 OpenAI 时的网络与账号配置成本。下面是实测用的最小接入示例：

```python
import json
import requests

BASE_URL = "https://4sapi.com/v1/chat/completions"
API_KEY = "sk-从4sapi控制台获取的密钥"  # 不要提交到仓库

def run_work_task(prompt: str, model: str = "gpt-5.6-sol",
                  reasoning: str = "high", stream: bool = True):
    """模拟 Work 的任务执行：模型 + 推理级别 + 流式输出"""
    payload = {
        "model": model,          # gpt-5.6-sol / gpt-5.6-luna / gpt-5.6-terra / gpt-5.5
        "reasoning": reasoning,  # light / medium / high / extra-high / max / ultra
        "messages": [
            {"role": "system", "content": "作为任务执行引擎，输出结构化结果。"},
            {"role": "user", "content": prompt},
        ],
        "stream": stream,
    }
    headers = {"Authorization": f"Bearer {API_KEY}", "Content-Type": "application/json"}
    with requests.post(BASE_URL, json=payload, headers=headers, stream=stream,
                       timeout=120) as resp:
        resp.raise_for_status()
        if not stream:
            return resp.json()
        for line in resp.iter_lines():
            if line:
                print(line.decode("utf-8"))

if __name__ == "__main__":
    run_work_task("抓取并总结本周 AI 模型发布动态，输出 300 字简报")
```

实测这个脚本跑「抓取 + 总结」链路时，把 `reasoning` 从 high 降到 medium，单次任务 Token 消耗下降约三成，简报质量对日常选题够用。真正需要精读长文档时再临时升到 extra-high，这就是我在第五节说的分阶段升级策略。

## 九、云桌面接入差异实测总结

把两条路径并排测完，差异很直观：

| 对比维度 | 产品直用（Work Cloud/Local） | API 编程接入（4sapi） |
| --- | --- | --- |
| 上手速度 | 分钟级 | 需要写代码与部署 |
| 工具链 | 内置沙箱、浏览器、Sites | 自行组装（沙箱/存储/编排） |
| 自动化 | 受产品编排限制 | 完全可控，可嵌入服务 |
| 多人使用 | 按账号订阅，成本线性 | 按 Token 计费，可共享 Key 池 |
| 模型矩阵 | GPT-5.6 Sol/Luna/Terra + GPT-5.5 | 同款模型，参数全可控 |
| 可观测性 | 弱（黑盒） | 强（日志、用量、计费明细） |

我的结论：个人探索用产品直用，效率最高；一旦要把能力嵌进自己的产品、做多用户或做成本治理，API 编程接入是唯一走得通的路。两条路不是替代关系，是先体验、后工程化的顺序。

## 十、成本与风险提示

接入前把账算清楚，四个方面：

- **订阅成本**：$20/月起才能用 Work，多人团队按账号买会指数级膨胀——这正是切 API 计费的核心动机。
- **Token 成本**：推理级别越高，单次消耗越大。Ultra 级别的长任务按量计费相当可观，务必用级别渐变策略控成本。
- **持久化数据**：Work Cloud 的文件系统在云端留存，敏感数据不宜放进去；API 接入时自行掌控存储，反而更容易满足数据边界要求。
- **合规与频率**：一切接入走官方开放渠道与合法订阅，不做绕过官方限制的代理；自动化任务控制并发与频率，把限流当成可预期事件而不是事故。

## 十一、接入验收清单

我实测后沉淀的验收清单，照着过一遍再上线：

1. 确认订阅档位在 $20/月及以上，Work 标签可见。
2. 分别验证 Work Cloud 与 Work Local 的入口、执行位置是否符合预期。
3. 用同一问题跑 Sol Medium 与 Sol Ultra，对比输出质量与 Token 用量。
4. 验证持久化文件系统：上一会话留下的文件，下一会话能否读到。
5. 验证联网代码执行与无头浏览器能访问目标站点并解析数据。
6. API 接入后，核对流式输出与异常重连行为。
7. 记录单任务成本明细，确认推理级别策略生效。
8. 检查自动化任务的频率上限与 Token 预算告警是否生效。

## 十二、总结

ChatGPT Work 一次发布、两个形态：云端版 Work Cloud 是带工具、存储与发布能力的任务运行平台，桌面版 Work Local 是本机文件与程序的操作入口，两者都只对 $20/月及以上订阅开放。与 Chat 相比，Work 多出联网代码执行、无头 Chrome、持久化文件系统、ChatGPT Sites 发布与 Sol/Luna/Terra 子智能体会话等能力；模型矩阵上，GPT-5.6 的 Sol/Luna/Terra 配 Light 到 Ultra 六级推理，GPT-5.5 只开放到 Extra High。

接入路径上，产品直用适合快速体验与个人任务，API 编程接入适合把同一套模型与推理能力嵌进服务、做多用户和成本治理——[4sapi](https://4sapi.com) 提供 OpenAI 兼容入口，模型与推理级别参数透传，配合分阶段升级推理级别的策略，可以在质量与成本之间拿到不错的平衡点。

欢迎在评论区发表想法，聊聊 Work 双形态的接入经验或者其他更好用的组合。