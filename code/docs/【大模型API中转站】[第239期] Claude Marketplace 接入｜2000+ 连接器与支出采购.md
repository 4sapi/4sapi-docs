---
title: "【大模型API中转站】[第239期] Claude Marketplace 接入｜2000+ 连接器与支出采购"
tags:
  - Claude
  - MCP
  - 连接器
  - 成本治理
  - 采购
description: "Claude Marketplace 把 2000 多个连接器与插件、Claude 驱动的第三方产品、服务伙伴收进同一个入口。从接入方视角拆解 MCP 连接器与插件的技术形态、三类入驻路径的适用对象，以及已承诺 Anthropic 支出可用于购买第三方产品这一采购机制对预算规划的影响，附连接器接入自建网关的 Python 示例与供应商评估清单。"
---

# 【大模型API中转站】[第239期] Claude Marketplace 接入｜2000+ 连接器与支出采购

Claude Marketplace 把插件与连接器、智能体与产品、服务伙伴集中到一个入口，对客户是一站式平台，对构建者和合作伙伴则是触达正在使用 Claude 的团队的途径。

我在 4sapi.com 上按客户侧的真实链路走了一遍：连接器以什么技术形态接入、三类入驻路径分别适合谁、以及"已承诺支出可以拿去购买第三方产品"这条规则会把企业预算表改成什么样，结论和工程细节都写在下面。

## 一、一个入口，客户侧的三件事

对客户来说，这个入口是一站式平台，用于找到合适的工具和服务；对构建者和合作伙伴，它是触达正在使用 Claude 的团队的途径。客户侧能做的事可以归纳为三条：

- **添加连接器和插件**。目前可用的选项有 2000 多个，包括 Atlassian、Google、Microsoft、Notion、Salesforce 等。这一步解决的是上下文问题：把外部系统里的真实数据接进 Claude，回答才有依据。
- **购买智能体和产品**。把已承诺的 Anthropic 支出的一部分，用于购买来自 CrowdStrike、Cursor、Harvey、Legora、Lovable、Snowflake 等公司的 Claude 驱动软件。这一步解决的是能力问题，而且改变了采购路径。
- **借助服务合作伙伴实现规模化**。与 Claude Partner Network 的咨询合作伙伴或系统集成商建立联系，例如 Accenture、Boston Consulting Group、Deloitte，用于制定 AI 战略并在整个组织内推广 Claude。这一步解决的是组织问题。

三条路径不是替代关系。连接器把数据接进来，产品把能力买进来，服务伙伴把方法带进来。常见的问题是只做了其中一件：装了连接器却没有治理，或者买了产品却没有数据基础。

## 二、生态结构：客户、Marketplace 与三类供给

把结构画出来更清楚：

```text
客户（正在使用 Claude 的团队）
  采购 / 安全 / 平台 / 业务线
        |
        v
Claude Marketplace（统一入口）
        |
        +--> 连接器与插件：2000+ 选项
        |      Atlassian、Google、Microsoft、Notion、Salesforce
        |      形态：MCP 连接器 + Agent Skills
        |
        +--> 智能体与产品：Claude 驱动的第三方软件
        |      CrowdStrike、Cursor、Harvey、Legora、Lovable、Snowflake
        |      形态：可直接采购的成品
        |
        +--> 服务伙伴：咨询与系统集成
               Accenture、Boston Consulting Group、Deloitte
               形态：人与方法论
        |
        v
回到客户侧：数据接入、能力采购、组织推广
```

三类供给的交付物不同，验收方式也应该不同：连接器验收数据范围与稳定性，产品验收业务指标，服务伙伴验收组织能力。用同一套标准对待三类供给，是项目在验收期反复拉扯的根源。

## 三、连接器与插件：MCP 与 Agent Skills 是什么形态

构建者侧的第一条路径，是为使用 Claude 的团队进行构建，用 Model Context Protocol（MCP）和 Agent Skills 创建连接器或插件。

MCP 是模型与外部系统之间的能力描述协议。连接器按协议把外部系统的工具和资源暴露出来，模型看到的是工具清单和参数结构，实际调用由运行时执行。对企业的意义在于：连接器是数据边界的第一道闸门——模型能看到什么、能改什么，由连接器暴露的工具集决定。

Agent Skills 更接近"打包好的做法"。它把一套流程、约束和操作步骤固化下来，让 Claude 在特定任务上按稳定方式工作。连接器解决"够得着"，技能解决"做得对"，两者组合起来才是可用的企业能力。

Atlassian 的做法可以当作参考：Teamwork Graph 通过连接器把 Confluence、Jira 及其他应用的完整上下文带入 Claude，使回答基于团队和工具中真实发生的情况。这类连接器的价值不在单次问答，而在把协作工具里的隐性知识变成可检索的上下文。

插件是面向使用者的那一层，在使用界面里提供扩展入口，背后通常是连接器或技能。对平台团队来说，插件是暴露给业务团队的表面，连接器才是需要治理的实质。

## 四、三类入驻路径分别适合谁

构建者侧的三条路径，和客户侧的三件事一一对应。选错路径的代价是白做一轮工程。

| 入驻路径 | 交付物 | 适合谁 | 客户怎么买单 | 前置条件 |
| --- | --- | --- | --- | --- |
| 用 MCP 和 Agent Skills 创建连接器或插件 | 能力接口 | 自有系统需要接进 Claude 的平台团队、软件厂商 | 通常随软件许可提供 | 按 MCP 规范实现工具，显式声明权限范围 |
| 申请上架 Claude 驱动的智能体或产品 | 完整产品 | 已有成型产品的软件公司 | 客户可用已承诺 Anthropic 支出的一部分购买 | 申请上架并通过审核 |
| 加入 Claude Partner Network 提供咨询或系统集成 | 服务与方法论 | 咨询公司、系统集成商 | 服务合同 | 加入伙伴网络并出现在 marketplace 中 |

判断标准只有一条：交付的是接口、产品，还是人。三条路径可以叠加：一家厂商既能上架产品，也能用连接器把自家产品的数据接进 Claude。

## 五、采购机制：已承诺支出可以买第三方产品

这是整件事里对财务和采购影响最大的一条：团队可以使用其已承诺的 Anthropic 支出中的一部分，来购买 marketplace 上由 Claude 驱动的智能体或产品。

传统流程下，采购一套第三方软件要单独走一轮供应商准入、安全评审、合同与付款。现在，已承诺的 Anthropic 额度本身具备了购买力，一部分原本只能用于 token 消耗的预算，变成了可以置换能力产品的预算池。三个直接后果：

1. **预算池用途扩展**。同一笔承诺额度，既能买推理能力，也能买成品能力。年度预算里"模型消耗"和"软件采购"两条线的边界开始模糊。
2. **采购周期压缩**。如果供应商已经在 marketplace 里，准入环节可以借用平台既有审核，商务流程短很多。
3. **供应商集中度上升**。额度集中在一个入口，议价和退出成本都会变化。承诺额度用得越多，切换成本越高。

一个容易被忽略的点：承诺额度是双向的。用不完的额度可能被浪费，用超了又要走追加流程。我的做法是把承诺额度显式拆成两个池子——token 消耗池和产品采购池，各自设上限和告警阈值，按月复核。具体比例取决于实际用量，需要按团队自身数据估算，没有通用值。

## 六、连接器进入企业网络要过哪几道闸门

连接器一旦能读到 Confluence、Jira、Notion、Salesforce 里的数据，它就不再是一个"小插件"，而是数据出口。客户侧常见的做法是把连接器统一收进自建网关：

```text
业务团队提出连接器需求
        |
平台与安全团队评审：数据边界、权限范围、供应商资质
        |
为每个连接器分配独立凭据（最小 scope，可轮换）
        |
流量统一走自建网关：鉴权、限流、脱敏、审计
        |
网关转发到中转端点与外部系统，模型调用与工具调用分开归集
        |
用量与费用按团队、项目汇总，进入成本看板
```

这条链里最容易被跳过的是第三步。多个连接器共用一把凭据，事后无法回答"这条数据是哪个连接器读走的"，也无法单独撤销某一个连接器的访问。凭据隔离是后续所有治理动作的前提。

## 七、Python 示例：把 MCP 工具接进自建网关

下面这段代码是接入侧的最小可用骨架：Key 从环境变量读取，连接器各自声明 scope 与额度，只把被授权的工具下发给模型，调用完成后按连接器归集用量。模型调用走兼容端点，自建网关只做治理。

```python
"""连接器网关最小骨架：凭据只从环境变量读取，不进代码仓库。"""
import os
import time
import httpx

API_KEY = os.environ["GATEWAY_API_KEY"]
BASE_URL = os.getenv("GATEWAY_BASE_URL", "https://4sapi.com/v1")
MODEL = os.getenv("GATEWAY_MODEL", "claude-sonnet")

CONNECTORS = {   # 一个连接器 = 一份独立凭据 + 一组最小权限 + 一个成本中心
    "atlassian": {"scopes": {"confluence:read", "jira:read"}, "team": "platform", "cap": 8_000_000},
    "notion":    {"scopes": {"notion:read"},                  "team": "product",  "cap": 3_000_000},
}

MCP_TOOLS = {    # 每个工具声明归属连接器与所需 scope
    "search_confluence": {"connector": "atlassian", "scope": "confluence:read"},
    "read_jira_issue":   {"connector": "atlassian", "scope": "jira:read"},
    "create_jira_issue": {"connector": "atlassian", "scope": "jira:write"},
    "search_notion":     {"connector": "notion",    "scope": "notion:read"},
}

LEDGER = {}  # connector -> 本月已用 token


def visible_tools(connector):
    """只下发被授权的工具；越权工具不进入模型可见的工具清单。"""
    allowed = CONNECTORS[connector]["scopes"]
    return [
        {"name": name, "input_schema": {"type": "object"}}
        for name, spec in MCP_TOOLS.items()
        if spec["connector"] == connector and spec["scope"] in allowed
    ]


def charge(connector, tokens):
    cap = CONNECTORS[connector]["cap"]
    used = LEDGER.get(connector, 0) + tokens
    if used > cap:
        raise RuntimeError(f"{connector} 额度已用尽：{used} > {cap}")
    LEDGER[connector] = used
    return used


def ask(connector, prompt):
    resp = httpx.post(
        f"{BASE_URL}/messages",
        headers={
            "x-api-key": API_KEY,
            "anthropic-version": "2023-06-01",
            "x-connector-id": connector,                 # 用量按连接器归集
            "x-team": CONNECTORS[connector]["team"],     # 费用按团队归集
        },
        json={
            "model": MODEL,
            "max_tokens": 1024,
            "messages": [{"role": "user", "content": prompt}],
            "tools": visible_tools(connector),
        },
        timeout=60.0,
    )
    resp.raise_for_status()
    data = resp.json()
    usage = data["usage"]
    used = charge(connector, usage["input_tokens"] + usage["output_tokens"])
    print(f"[{connector}] team={CONNECTORS[connector]['team']} used={used} at={time.time():.0f}")
    return data


if __name__ == "__main__":
    ask("atlassian", "汇总本周 Jira 阻塞项，只引用可读到的 issue")
```

几个设计要点：

- **工具清单按连接器裁剪**。`visible_tools` 只下发 scope 命中的工具，`create_jira_issue` 这类写操作在没有 `jira:write` 授权时根本不出现在模型可见列表里。这比在提示词里写一句"不要创建 issue"可靠得多。
- **用量按连接器与团队双维度归集**。请求头里带上连接器标识和团队标识，网关侧就能产出"哪个团队、通过哪个连接器、花了多少"的报表，而不是一笔总数。
- **额度上限在网关侧硬拦**。`charge` 在超限时直接抛错，避免单个连接器把整月预算吃掉，也避免事后才发现。
- **端点可替换**。`BASE_URL` 走环境变量，换供应商只改配置，业务代码里没有硬编码的域名。

## 八、权限与用量隔离的三条边界

连接器接入的治理可以收敛到三条边界：

**数据边界**。连接器能读到的资源范围必须显式列出，只读与可写分开授权。Teamwork Graph 这类把协作工具上下文接进 Claude 的连接器，先开只读是稳妥的起步方式。

**权限最小化**。每个连接器一份独立凭据，scope 显式声明、可轮换、可单独撤销。凭据共用是审计失效的第一步。

**用量隔离**。按连接器、团队、项目三级归集用量，设上限和告警。没有隔离的用量数据，做不了任何成本优化——连"哪个团队在用"都答不出来。

还有一条工程纪律：连接器是适配层，业务逻辑不要写进连接器。

## 九、供应商评估清单

无论选择哪一类供给，评估的四个维度是固定的：

| 维度 | 要问的问题 | 合格线 | 风险信号 |
| --- | --- | --- | --- |
| 数据边界 | 连接器读取哪些资源、数据存在哪里、留存多久 | 资源范围与留存策略可书面确认 | 只谈加密不谈范围 |
| 权限最小化 | 是否支持只读 scope、是否每连接器独立凭据 | scope 显式、可撤销、可轮换 | 要求一个账号的全库权限 |
| 计费口径 | 按席位还是按 token，是否支持用已承诺支出购买 | 口径写入合同，与 token 消耗分开统计 | 混在一条费用里说不清 |
| 可替换性 | 换掉供应商要改多少代码、数据能否导出 | 连接器为适配层，迁移路径明确 | 业务逻辑焊死在连接器里 |

这份清单同样适用于评估中转与网关环节：Key 的权限范围、用量能否按项目拆分、单价与折扣是否稳定，都直接影响成本治理的可行性。

## 十、成本与风险提示

- **承诺额度有机会成本**。额度拆成 token 池和产品池之后，两边都可能出现用不完或不够用的情况，按月复核比一次性规划更实际。具体比例需要按自身用量估算，没有通用值。
- **2000+ 选项不等于 2000+ 可用项**。真正能过安全评审、有明确负责人、数据范围清楚的连接器，通常只有个位数。建议先挑三到五个高价值场景打通，再谈规模。
- **数据出域要提前定线**。连接器把外部数据带进模型上下文，等于数据多了一次流转。哪些系统的数据可以进、进到什么程度，应该由安全团队先划线。
- **供应商集中度是长期风险**。额度、连接器、产品集中在同一个入口，短期省事，长期议价空间会变小。可替换性要在采购时谈，而不是在续约时谈。
- **计费口径差异会造成对账困难**。席位制、token 制、服务合同三种口径混在一起时，成本看板要按口径分列，不要合并成一个数字。
- **网关与 Key 管理不能省**。独立凭据、scope 声明、额度硬拦、用量归集，这四件事在接入第一天就该有，补做的成本远高于先做。
- 以上涉及成本的部分都是工程化估算方法，不构成金额承诺；实际单价与额度规则以各供应商当前合同条款为准。

## 十一、落地顺序

从零开始接入，我建议按这个顺序推进：

1. 列出三个最痛的上下文缺口，例如"回答问题时拿不到 Jira 的最新状态"；
2. 用 MCP 连接器先接只读数据，范围越小越好；
3. 所有连接器流量收进自建网关，独立凭据加显式 scope；
4. 用量按连接器、团队、项目三级归集，设月度上限与告警；
5. 跑两到四周，对比回答质量与用量曲线，淘汰没被用起来的连接器；
6. 再评估是否用已承诺支出采购 Claude 驱动的第三方产品；
7. 需要组织级推广时，再引入 Claude Partner Network 的咨询伙伴或系统集成商。

顺序背后的逻辑是：先解决数据，再解决能力，最后解决组织。反过来做的项目，通常会在"买了产品但没人用"这一步停下来。

## 结论

Claude Marketplace 把插件与连接器、智能体与产品、服务伙伴收进一个入口，客户侧对应三件事：接数据、买能力、找伙伴；构建者侧对应三条入驻路径：用 MCP 和 Agent Skills 做连接器或插件、申请上架 Claude 驱动的产品、加入 Claude Partner Network 提供咨询与系统集成。

对接入方而言，最需要提前想清楚的是两件事。一是连接器的治理：数据边界、权限最小化、用量隔离、可替换性，四件事要在接入第一天就位。二是采购机制的变化：已承诺的 Anthropic 支出可以用于购买第三方产品，预算池的用途变宽了，但承诺额度的双向属性、供应商集中度和计费口径差异，都需要在预算表里单独列出来管。

这套"连接器 + 自建网关 + 额度归集"的组合，我是在 4sapi.com 上跑通的：连接器负责把上下文接进来，网关负责把权限和用量管住，额度模型负责让成本可解释。企业侧的 Claude 接入不缺工具，缺的是把工具管起来的那层结构。

欢迎在评论区聊聊手头团队的连接器清单和额度拆分方式，或者踩过的权限与用量归集的坑。
