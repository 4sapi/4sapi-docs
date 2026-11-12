---
title: "用 graphify 查询代码库：query、path、explain 三个命令的用法和结果自查方法"
category: 人工智能
tags:
  - 代码知识图谱
  - 编程智能体
  - 结果自查
description: "依据 Graphify-Labs/graphify 的 README 与安全文档，说明 query、path、explain 三个查询命令各回答什么问题，助手怎样被引导优先查图，CLI 与 MCP 两种接入的区别，以及规模上限和过期图带来的风险，并给出一个用已知答案问题自测召回率的方法和脚本。适合要让编程智能体改用图查询的工程师。"
brand_mode: none
---

# 用 graphify 查询代码库：query、path、explain 三个命令的用法和结果自查方法

图建好之后，真正的使用环节是查询。Graphify 提供三个命令，分别回答"这个问题相关的子图是什么""两个东西怎么连起来""这个概念周围有什么"。但图查询的结果不一定完整，助手也不一定会去查图。

本文回答三个问题：三个命令各适合问什么；助手怎样被引导优先查图，这个引导有多强；怎样用自己的已知答案检验查询结果。依据是仓库 [README](https://github.com/Graphify-Labs/graphify/blob/HEAD/README.md) 和[安全文档](https://github.com/Graphify-Labs/graphify/blob/HEAD/SECURITY.md)（检索于 2026-10-30），我没有安装或运行过，命令输出示例都是项目展示的，不是我的结果。

## 一、三个命令

README 的说法如下：

| 命令 | 回答什么 |
| --- | --- |
| graphify query "问题" | 对一个自然语言问题返回一个限定范围的子图 |
| graphify path A B | 追踪任意两个事物之间的连接路径 |
| graphify explain X | 解释一个概念：位置、所在社区、度数、全部连接 |

在助手里对应 /graphify query、/graphify path、/graphify explain；Codex 里用 $graphify 代替 /graphify。

### 项目展示的输出

README 用 FastAPI 代码库展示了两段输出（项目给出，非我运行）：

```text
$ graphify explain "APIRouter"
Node: APIRouter
  Source:    routing.py L2210
  Community: 2
  Degree:    47

$ graphify path "FastAPI" "ModelField"
Shortest path (3 hops):
  FastAPI --uses--> DefaultPlaceholder <--references-- get_request_handler() --references--> ModelField
```

explain 的每条连接还带关系名和置信度标签，例如 uses 加 INFERRED、method 加 EXTRACTED。读结果时先看标签：EXTRACTED 是源码里直接读出的，INFERRED 是推断的。

### 什么时候用哪个

| 想问的 | 用 |
| --- | --- |
| "认证和数据库之间有什么关联" | query |
| "UserService 怎么到达 DatabasePool" | path |
| "RateLimiter 周围有哪些东西" | explain |
| 不知道从哪里看起 | 先读 GRAPH_REPORT.md 的枢纽节点和建议问题 |

这是我按命令定义做的归纳，不是文档的原话。

## 二、助手怎样被引导去查图

建图不等于助手会用图。README 的"让助手总是用图"一节说明，需要运行一次 graphify 加平台名 install，例如 graphify claude install。它写入一个小配置，告诉助手遇到代码库问题先查图，优先用范围限定的 graphify query，而不是读完整报告或 grep 原文件。

| 平台类型 | 机制 |
| --- | --- |
| 钩子平台（Claude Code、Gemini CLI） | 搜索类工具调用前自动触发钩子，把助手推向图路径；Claude Code 上还包括逐个读取源文件前 |
| 指令文件平台（Codex、OpenCode、Cursor 等） | 持久指令文件（AGENTS.md、.cursor/rules/ 等）给出同样的先查图指引 |

### 这个引导有多强

两点来自 README 对 Pi 守卫的描述，能说明钩子的性质：

- 守卫有明确的资源边界：单次调用限 2 秒、输入 256 KiB、输出 64 KiB。
- 它"失败时放行"（fails open）。守卫出错或超时，助手照常去 grep 或读文件。

另有环境变量 GRAPHIFY_HOOK_STRICT=1，作用是对已被索引的文件，每个会话第一次直接读取时拒绝一次。

结论：这个机制是"推"，不是"强制"。助手可能不查图，你也无法从文档推断它查图的比例。要知道实际情况，只能看会话记录里每个问题用了什么工具。

### CLI 技能与 MCP 是两套

README 专门说明：graphify install 安装的是本地助手集成，不会建立 MCP 连接。已连接的 MCP 服务器有自己的指令和工具列表。两者都鼓励先做范围限定的图查询，但 README 提醒：同时使用时，先确定每套查的是哪张图，选一条主路径，不要把两套无条件指令都复制进 AGENTS.md。

MCP 的网络边界（来自安全文档）：

- 默认 stdio，不开网络监听。
- --transport http 是可选的，默认绑定 127.0.0.1，只有显式传 --host 0.0.0.0 才对外。
- 图文件路径要求解析后位于 graphify-out/ 内。

## 三、规模与过期：两类静默风险

### 规模上限

| 项目 | 文档中的数值或行为 |
| --- | --- |
| HTML 导出 | 默认 5,000 节点，超过则展示聚合的社区视图，可用 --node-limit 调整 |
| 社区标注 | 每次调用默认 100 个社区，按约 16k token 上下文设计；8k 窗口建议 --batch-size 40 |
| 溢出处理 | 上下文溢出不做拆分重试，失败批次保留基于枢纽的名称 |
| 枢纽 | --exclude-hubs 99 可把工具类超级枢纽排除出枢纽节点排名 |

在大仓库里，这些设置决定你看到的是细节还是聚合。看到的不一定是全貌。

### 图会过期

图是快照。README 说钩子在提交和切分支时重建，git pull 之后要手动运行 graphify update . 。过期的图查出来的结果看似合理，却对应旧代码。自查时要先确认图的生成时间晚于最近一次拉取。

### 忽略规则

建图遵守 .gitignore，也支持 --exclude 和项目里的 .graphifyignore。被忽略的目录不在图里，查询自然查不到。如果问题涉及生成代码或第三方目录，先确认它们没被排除。

### 中文查询

README 列出 chinese 额外依赖（jieba）用于中文查询分词。没有安装时，用中文问句的查询效果我无法判断，文档没有说明回退行为。

## 四、用已知答案自测召回

厂商的基准不能代替你自己的仓库。下面是一个低成本的自测方法。

### 步骤

1. 选 10 到 20 个你确定答案的问题。答案是具体的文件或符号，例如"处理支付回调的入口在哪个文件"。
2. 为每题写下期望命中的节点或文件。
3. 对每题运行 query，记录返回的节点或文件。
4. 同样的问题让助手用 grep 回答一遍，作为对照。
5. 计算每题召回，即期望项中被返回的比例，再取平均。
6. 单独记录三类失败：图里没有该符号（建图问题），图里有但没被返回（查询问题），返回了但标成 INFERRED 且来源有误（精度问题）。

### 计算召回的脚本

输入是一个 JSON 数组，每项有 q、expected、returned 三个字段。

```python
import json, sys
cases = json.load(open(sys.argv[1], encoding="utf-8"))
total = 0.0
for c in cases:
    exp, ret = set(c["expected"]), set(c["returned"])
    r = len(exp & ret) / len(exp) if exp else 1.0
    total += r
    missed = sorted(exp - ret)
    print(f'{r:.2f}  {c["q"]}' + (f'  missed={missed}' if missed else ""))
print(f"mean recall: {total / len(cases):.2f}" if cases else "no cases")
```

我用三个手造的用例测过：一题全中输出 1.00，一题 2 个期望命中 1 个输出 0.50 并列出缺失项，一题期望为空输出 1.00，平均 0.83。它只做集合比较，不调用 graphify，也没有在真实查询结果上跑过。

### 怎么解读自测结果

- 问题数很少时，平均值波动很大。10 题里一题的差别就是 10 个百分点。
- 别只看召回。返回的内容里有多少是噪声，也要抽样看。
- 对照 grep 的结果：图查询若不比 grep 更快更准，就没有理由改变助手的习惯。

## 五、使用时的检查清单

| 检查项 | 怎么做 |
| --- | --- |
| 图是新的吗 | 比较图生成时间和最近一次 git pull |
| 要问的目录在图里吗 | 检查忽略规则 |
| 歧义扩展名解析对了吗 | 抽查对应文件的符号是否出现 |
| 助手真的在查图吗 | 看会话记录里的工具调用 |
| 结果里 INFERRED 占多少 | 用置信度统计脚本 |
| MCP 与 CLI 是否重复注入指令 | 检查 AGENTS.md |

## 六、限制与结论

三个命令分别覆盖"相关子图""两点路径""概念邻域"，边上的置信度标签让结果可审计。但引导助手查图的机制是失败时放行的建议，规模设置和过期会让结果静默变差，图里缺失的内容也不会被提示。

本文的限制：没有安装或运行；输出示例来自项目展示；没有读源码和在线文档；中文查询的行为未知；召回脚本只在手造数据上测过。

下一步建议：用第四节的方法，在你自己的仓库上跑 10 题，得到一个召回数字和三类失败的分布，再决定图查询在你的工作流里该占多大比重。
