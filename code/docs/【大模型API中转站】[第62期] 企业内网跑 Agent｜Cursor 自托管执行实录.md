---
title: "【大模型API中转站】[第62期] 企业内网跑 Agent｜Cursor 自托管执行实录"
tags:
  - Agent
  - Cursor
  - 自托管
  - 安全合规
  - 企业架构
description: "拆解 Cursor Self-Hosted Machines 的推理/执行分离设计：智能体循环留在云端、工具执行落回企业自有机器、worker 出站 HTTPS 对接、内网零暴露，并给出 4sapi 接入与 Python 落地示例。"
---

# 【大模型API中转站】[第62期] 企业内网跑 Agent｜Cursor 自托管执行实录

Cursor 的 Self-Hosted Machines 把云智能体的工具执行搬进了企业自有网络，推理与规划仍留在云端。我复现了这套"云端思考、内网动手"的架构，并在 4sapi（https://4sapi.com）上把内网执行结果接回了模型循环。

## 一、开篇痛点：云 Agent 的手伸不进内网

企业里的 Agent 落地卡在同一个地方：模型在云上，代码在内网。

想让我写好代码，就得让我读代码；想让我读代码，就得让代码离开内网。可企业安全策略的第一条往往是"敏感代码不得出内网"。于是云 Agent 变成只会空谈的顾问：能看文档摘要，碰不到真实仓库、真实测试、真实生产日志。

我踩过的三个具体坑：

- 云 IDE 或云端 Agent 读不到内网仓库，只能靠人工把代码片段复制出去，既慢又容易带出敏感信息；
- 把整套 Agent 环境搬进内网自建，模型更新、GPU 成本、运维负担全压到自己身上；
- 用 VPN 或反向代理把内网口子开给云端，等于为了一个 Agent 把整个内网暴露在公网攻击面里。

我需要的是第三条路：模型的能力在云端不缩水，代码和数据一步不出内网，云端也永远不需要连进内网。

## 二、原理速览：推理与执行分离

Self-Hosted Machines 的核心，是把 Agent 拆成两半，放到两个信任域里：

```text
云端：智能体循环 / 推理 / 规划
    —— 决定"下一步做什么"
内网：工具执行
    —— 负责"真的去做"
```

智能体循环、推理和规划仍留在 Cursor 云端，工具执行发生在企业自有机器上；两边通过 worker 的出站 HTTPS 连接对接；Cursor 不会主动连入企业网络。

用一句话概括：大脑在云上，手在企业自己的地盘里。大脑每次只下达一个具体的工具调用请求，手执行完把结构化结果送回去，大脑再决定下一步。

这个分离带来三个直接收益：

- 模型能力不打折：推理、规划、长上下文都在云端最强的模型上完成，不受内网硬件限制；
- 敏感数据不出内网：源码、凭据、生产数据只在内网机器上被读取和处理；
- 内网零入站：云端永远不建立到内网的连接，安全边界清晰。

## 三、Self-Hosted Machines 架构拆解

我把这套架构的请求流画出来：

```text
企业内网 ──────────────────────────────────────────
  │  企业自有机器（Self-Hosted Machine）
  │    └── worker 进程
  │          ├── 拉取工具调用任务（出站 HTTPS）
  │          ├── 在白名单内执行：读代码 / 跑测试 / 改文件
  │          └── 回传结构化结果（出站 HTTPS）
  └───────────────────────┬───────────────────────
                          │ 仅出站 HTTPS，无入站端口
                          v
Cursor 云端 ──────────────────────────────────────
  ├── 智能体循环（维护任务状态）
  ├── 推理与规划（模型调用）
  └── 工具调用下发 / 结果解析
```

关键点有三。

第一，worker 是企业机器上的一个常驻进程，不是云端派下来的代码包。它跑在企业的网络、防火墙和监控体系之内，我能看到它的一举一动。

第二，连接方向是 worker 主动向外。worker 通过出站 HTTPS 与云端对接，云端没有到内网的任何通道。

第三，工具白名单。worker 只暴露有限的动作集合，例如读取指定目录文件、运行测试、应用补丁。模型不能对 worker 下任意命令，能做的动作是企业预先批准的。

## 四、只走"出站"：内网零暴露的关键

传统"云 Agent 接内网"的做法，要么在内网开 SSH 端口给云端，要么架 VPN。两者都把内网端口暴露给公网，安全团队基本不会批。

出站方案把问题反过来：我不需要给云端任何进入内网的路径，只需要让 worker 能访问公网 443 端口。防火墙规则简化成一条：放行 worker 到云端域名列表的出站 HTTPS。

| 维度 | VPN / 反向代理方案 | 出站 worker 方案 |
| --- | --- | --- |
| 内网入站端口 | 需要开放 | 零入站端口 |
| 公网攻击面 | 端口暴露在公网 | 仅 worker 出站 |
| 防火墙规则 | 复杂、逐条审批 | 一条出站白名单 |
| 数据出内网 | 云端可直接读取 | 只有任务参数与脱敏结果 |
| 审计 | 难以定位连接来源 | worker 本地可全程记录 |

出站连接的另一个好处是可观测。worker 发出多少请求、连到哪些域名、传了多大载荷，内网出口网关都能统计；被动监听端口的方式，审计起来要麻烦得多。

## 五、数据边界与 API 接入合规视角

这套架构对做 API 接入的人来说，最有价值的部分是它把"数据边界"显式画了出来。任何一次 Agent 调用，都可以按这条链追数据流向：

```text
任务下发 → 策略过滤（哪些工具可调用）
    → worker 在内网执行（源码/凭据不出内网）
    → 结果脱敏 → 出站回传
    → 云端模型继续推理
```

合规的核心不是"数据不出门"，而是"每一字节数据去了哪里、在谁手里、停留多久"都说得清。Self-Hosted Machines 把这条链缩短到最小：出内网的只有工具调用参数和脱敏后的执行结果，源码、凭据和原始输出全程留在内网。

落到 API 接入上，我习惯把合规拆成三个问题：

- 谁在调用：任务必须携带可审计的身份与签名，内网侧能确认任务确实来自我的系统；
- 调了什么：工具名、参数摘要、退出码全部留日志，敏感值脱敏；
- 数据去向：喂给模型的内容是脱敏摘要，而不是内网文件的原文。

这三个问题，4sapi（https://4sapi.com）网关的鉴权、限流与用量审计能覆盖一大半——把"谁在什么时间以什么身份调用了哪个模型"变成可查询的记录。企业内网 Agent 的合规文档里，我会把 4sapi 网关当作云端侧的审计锚点，把 worker 日志当作内网侧的审计锚点，两边对账。

## 六、接入教程：在内网落地一个 worker

按这个模式，我在企业内网搭了一个最小可用的 worker。环境准备：

- 一台企业自有机器（物理机或内网虚拟机均可）；
- Python 3.10+；
- 出站 HTTPS（443）白名单放行；
- 一个与云端共享的 HMAC 密钥。

worker 的核心逻辑是三步：校验任务来源 → 白名单执行 → 结构化回传。简化版示例：

```python
import hashlib
import hmac
import json
import os
import subprocess

SHARED_SECRET = os.environ["AGENT_WORKER_SECRET"]
ALLOWED_TOOLS = {"read_file", "run_tests", "apply_patch"}

def verify_request(payload: bytes, signature: str) -> bool:
    expected = hmac.new(
        SHARED_SECRET.encode(), payload, hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(expected, signature)

def run_tool(tool: str, args: dict):
    if tool not in ALLOWED_TOOLS:
        raise PermissionError(f"tool not allowed: {tool}")
    if tool == "read_file":
        path = os.path.normpath(args["path"])
        if not path.startswith(os.environ["AGENT_WORKSPACE"]):
            raise PermissionError("path outside workspace")
        with open(path, encoding="utf-8") as f:
            return {"ok": True, "content": f.read()[:4000]}
    if tool == "run_tests":
        proc = subprocess.run(
            ["pytest", "-q"], capture_output=True, text=True, timeout=300
        )
        return {"ok": proc.returncode == 0,
                "summary": proc.stdout[-1000:]}
    return {"ok": False, "error": "not implemented"}

# 主循环：从云端任务端点拉取任务（出站 HTTPS）
# 校验签名 -> 执行 -> 回传结果，全程不开放任何入站端口
```

设计要点：

- worker 不提供任意命令执行，只实现白名单工具；
- 路径做规范化校验，防止越过工作区；
- 设置超时与输出长度上限，防止工具失控。

## 七、接入 4sapi：把执行结果接回模型循环

worker 执行完工具后，结果需要回到云端模型手里继续推理。我用 4sapi（https://4sapi.com）的 OpenAI 兼容接口完成这一步——worker 出站 HTTPS 调用网关，网关统一鉴权、限流与计费：

```text
内网 worker 执行工具
    │
    v  出站 HTTPS
4sapi 网关（https://4sapi.com/v1）
    │  统一鉴权 / 限流 / 计费 / 审计
    v
云端模型推理（继续智能体循环）
```

Python 示例：

```python
import os
import json
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["4SAPI_API_KEY"],
    base_url="https://4sapi.com/v1",  # 统一网关地址
)

def tool_result_summary(tool: str, result: dict) -> str:
    # 只把脱敏摘要交给模型，原始输出留在内网日志
    payload = json.dumps(result, ensure_ascii=False)[:2000]
    resp = client.chat.completions.create(
        model=os.environ.get("AGENT_MODEL", "claude-sonnet-4-5"),
        messages=[
            {"role": "system",
             "content": "把工具执行结果压缩为脱敏摘要：剔除密钥、内网路径与IP。"},
            {"role": "user", "content": f"{tool}: {payload}"},
        ],
        temperature=0.2,
    )
    return resp.choices[0].message.content
```

关键点是"脱敏发生在出网之前"：worker 把结果摘要化之后再经 4sapi 送到模型，内网文件原文、凭据与内部路径都不进模型上下文。模型拿到的永远是"够推理、不含敏感"的信息。

## 八、权限、密钥与最小化设计

- 工具白名单：能做的动作只有 read_file、run_tests、apply_patch 这类预先批准项，模型不能请求任意 shell；
- 来源签名：任务请求带 HMAC 签名，防止内网伪造任务或第三方注入；
- 最小化回传：结果先脱敏再出网，密钥、内网路径、IP 一律不进模型上下文；
- 路径边界：文件读取限定在工作区内，路径规范化后校验前缀；
- 密钥轮换：AGENT_WORKER_SECRET 与 4SAPI_API_KEY 定期轮换，日志中不记录原文；
- 内网审计：worker 本地记录任务 ID、工具名、参数摘要、退出码与时间戳，与 4sapi 网关的云端调用日志对账。

## 九、成本与风险提示

先算账：

| 成本项 | 说明 | 优化思路 |
| --- | --- | --- |
| worker 机器 | 企业自有机器或内网虚拟机 | 按并发复用，不必每任务一台 |
| 模型调用 | 推理仍在云端按 token 计费 | 经 4sapi 统一计费、限流与缓存复用 |
| 回传流量 | 仅脱敏摘要，量很小 | 摘要长度上限已内置 |
| 运维 | worker 部署与密钥轮换 | 模板化部署，密钥托管在密钥管理系统 |

风险提示：

- worker 上的工具执行仍有副作用：跑测试会改文件、apply_patch 会改代码。白名单之外再加一层"高风险动作审批"更稳妥；
- 脱敏不彻底是最大的合规隐患：摘要生成规则要单独测试，密钥、内网主机名、绝对路径都要覆盖；
- 出站域名要严格白名单：worker 能访问的云端域名越少越好，防止 worker 被当作跳板；
- 模型调用成本会随任务量线性增长：统一经网关限流，避免单任务 token 失控；
- 这里讨论的全部是合法接入与架构设计：推理/执行分离、出站对接、脱敏与审计，不涉及任何绕过官方限制或违规代理的做法。

## 十、企业 Agent 落地清单

- [ ] 明确哪些数据可出内网、哪些必须留内网，并写入合规文档
- [ ] worker 只实现白名单工具，不暴露任意命令执行
- [ ] 任务来源 HMAC 签名校验，防伪造
- [ ] 回传结果脱敏：密钥、内网路径、IP 全部剔除
- [ ] 防火墙只放行 worker 出站 HTTPS 443，零入站端口
- [ ] 经 4sapi（https://4sapi.com）统一接入、限流、计费与审计
- [ ] 密钥托管与定期轮换，日志不落原文
- [ ] 内网 worker 日志与云端网关日志可对账

## 十一、上线前验收清单

- [ ] 尝试用白名单之外的工具名调用 worker，确认被拒绝
- [ ] 尝试读取工作区之外的文件，确认被拒绝
- [ ] 带错误签名的任务，确认无法通过校验
- [ ] 构造包含密钥与内网路径的执行结果，确认回传前已被脱敏
- [ ] 检查内网出口日志，确认 worker 只连了白名单域名
- [ ] 断网重连后，确认 worker 出站连接自动恢复且不重复执行
- [ ] 审查 4sapi 网关调用日志与 worker 日志，确认任务可对账

## 总结

Cursor 的 Self-Hosted Machines 给出了一条清晰的企业 Agent 落地路径：推理和规划留在云端，工具执行落回企业自有机器，全程只靠 worker 的出站 HTTPS 对接，云端从不连入内网。我把这套"云端思考、内网动手"的模式在企业环境里复现了一遍：worker 白名单执行、结果脱敏回传、经 4sapi（https://4sapi.com）统一接入模型循环，敏感代码一步不出内网，合规与能力两头都保住了。欢迎在评论区发表想法，一起聊聊企业内网 Agent 的落地姿势。
