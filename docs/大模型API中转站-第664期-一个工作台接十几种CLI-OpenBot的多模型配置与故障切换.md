---
title: "一个工作台接十几种 CLI：OpenBot 的多模型配置与故障切换"
category: 人工智能
tags:
  - 多模型路由
  - Agent 工程
  - 凭证管理
description: "把 Codex、Claude、Grok 等十几个 CLI 放进同一个工作台，难点不在接入而在四个语义：凭证归谁、模型目录怎么统一、备用模型在什么条件下才接手、用量查询怎么不把配额问崩。本文逐条拆开 OpenBot 的实现，并给出自建同类系统时的对照清单。"
brand_mode: example
---

# 一个工作台接十几种 CLI：OpenBot 的多模型配置与故障切换

同时接多家供应商的产品，最后卡住的几乎都是同一批问题：密钥存在哪、模型列表从哪来、主模型挂了谁接手、用量查询会不会把自己问进 429。

OpenBot 的答案写在 `docs/architecture/providers.md` 里，一万多字，全是这类边界条件。本文拆成四个可迁移的问题，最后给一份自建时的对照清单。接口与包结构见第 663 期。

文中所有机制描述来自 2026-11-12 检索的仓库文档与 README，版本以 `main` 分支为准。

## 一、接的是什么：不是 API，而是十几个 CLI

这套工作台没有"统一模型 API"这一层，接的是各家官方命令行工具：

- 内置支持 Codex App Server 与 Claude Code；
- 经 ACP 接入 Grok、OpenCode、Gemini（走 Google 的 Antigravity CLI）、Cursor、Cline、Kimi、Devin、GitHub Copilot、Mistral Vibe、Hermes；
- 另有两类扩展：自定义 ACP agent，以及自定义的 OpenAI 兼容端点（含对本地模型服务器 Ollama、LM Studio 的检测）。

这个选择决定后面所有问题的形状：接 HTTP API 时，凭证、模型清单、用量都是自己的抽象；接别人的 CLI，这三样都在对方手里，工作台只能编排。

## 二、凭证：谁的登录，谁负责

README 的"本地数据与网络边界"一节把这件事写得非常明确：每个 provider 的登录与恢复状态由该 CLI 自己管理，不是 OpenBot 的会话存储。

| Provider | 归属目录 | 备注 |
| --- | --- | --- |
| Codex | `~/.codex` | 登录与历史由 Codex CLI 独占 |
| Claude Code | `~/.claude` | 同上 |
| Grok | `~/.grok` | 同上 |
| Gemini | `~/.gemini` 或 `$GEMINI_HOME` | Antigravity CLI 管理 |
| Cursor | `~/.cursor` | 受限进程用 `openbot-confined`，不碰你的设置 |
| Cline / Kimi / Devin / Copilot / Mistral | `~/.cline`、`~/.kimi`、devin 的 XDG 目录、`~/.copilot`、`~/.vibe` | 后两者登录在系统钥匙串 |
| Hermes | `~/.hermes` | 只读 `auth.json`/`.env`/`config.yaml` 的修改时间判断登录 |

这张表的价值不在记路径，而在一条原则：**不要把第三方 CLI 的登录态复制进自己的配置。** 双写必然不一致——用户在 CLI 里换了账号，工作台还以为旧账号有效。

例外只有一处：自定义 ACP agent 的环境变量。它们整体存成 `safeStorage` 密文（`custom-agents.json`），只有后端拿得到明文，渲染进程只看摘要。也就是说，用户手填的秘密由工作台保管，通过 CLI 登录得到的会话则完全不参与。

## 三、模型目录：统一读取，差异下沉

第二个问题是"模型列表从哪来"。多供应商系统常见的写法是各接各的接口、各写一套解析，半年后没人知道某一行是从哪个字段来的。

这里的做法相反：`ModelCatalog.refresh` 对每个运行中的 CLI 统一读 `model/list`，没有 provider 分支；真正需要差异的地方由每个 driver 声明 `catalog` 规则：

- Codex 的列表要分页，并且会报出一个 Fast 档；
- Claude Code 自己给模型命名；
- OpenCode 会剔除"当前密钥买不到"的模型，并在发现失败时保留免费档；
- 自定义 agent 使用完整的请求超时。

文档还写明设计意图：**新 provider 的规则加在 driver 里，不加在 catalog 里。** 好处是主干只有一份实现，坏处是差异必须显式登记，否则被主干默认行为覆盖。

启示是：模型列表最容易变成"每家一段 if"，先收成无分支流程再允许逐家覆盖，比事后重构便宜。

## 四、切换：一条会话，两种 id

同一份对话在 Codex、Grok、Claude 之间切换时，工作台保留一条稳定的本地会话；provider 原生的 session id 只用于恢复运行时状态，且不对外暴露（远程客户端也拿不到）。

这解决了多供应商系统里最烦的问题：会话是有归属的。若 UI 把 provider 的 session id 当会话主键，换 provider 就等于换会话，历史、附件、排队任务全要重做。分成两层之后，切换退化成一次"重新绑定"。

代价也存在：绑定能否用，取决于 provider 能否恢复那个 session。所以文档里有"handoff"——线程换到不认识该 session 的 agent 上时，运行时按切换的方式移交会话。

## 五、故障切换：只有"还没动手"的那一轮才会换模型

这一节和多数人想的"失败就重试另一个模型"不一样。

每个 Agent 在 `agent_json` 里存一份有序的备用模型列表 `AgentBackupModels`，最多 8 个，设置页、管理路由与 `update_profile` 都能改。自动模式下由 `UsageLimitGate` 与调度逻辑把 Agent 移到第一个"没有被 hold 覆盖"的模型；原模型解除限制后 `usageLimitReturn` 带它回去。

关键在于**两类 hold**的处理不同：

| 类型 | 行为 |
| --- | --- |
| 计划用量上限 | 会 hold 该 provider 账号与该模型上所有 Agent 的队列，直到配额重置 |
| Provider 错误 | 不 hold 任何东西；`classifyFailure` 把错误归为 `billing`、`provider_unavailable`、`rate_limit`、`authentication`、`network`、`timeout`；只有这一轮尚未动手时才把消息放回队列，并对该模型 hold 10 分钟 |

还有三条更细的边界：

- **已经输出文本或执行过工具的一轮不切换**，它会带错误失败，与没有备用模型时相同：半途换模型会让上下文与工具副作用错位。
- **任务、频道任务、被拒绝的 `turn/start`、以及 provider 起不来，都不触发切换。**
- **同一条消息对每个失败模型最多重跑一次**；hold 在该模型有一轮成功时解除，重启清空所有 hold（所以重启后第一次调度会回到原模型）。

这套语义把"容错"限定在可容错的位置：**失败发生在副作用产生之前，才值得换人。** 若业务允许半途换模型，就得自己处理工具状态迁移。

## 六、用量查询：把轮询做成一个门

多供应商系统常见的一种事故是"用量面板把配额问没了"。这里的处理是给读取加一个门：

- 每个桌面窗口、网页标签、手机都每 5 分钟轮询一次，且都落到本机；
- `UsageReadGate` 保证同一 provider 登录态同时只有一个读取在跑，后到的请求排队并复用这次读数；
- 一次读数在 4 分钟内有效，略短于轮询间隔；
- 读取失败后，下一次等待 5 分钟，再失败等 10、20、40 分钟，直到 1 小时；
- 文档特别点名：Claude 的用量端点与 Claude Code 共享，问得太勤会返回 429 且持续一小时以上。

最后那条是这套设计的原因："用量查询"不是无害的只读操作，它消耗与真实用户共享的配额。

## 七、CLI 版本：谁在管你机器上的一堆二进制

接 CLI 的第二个代价是版本。工作台的做法是托管：下载并锁定各家的 CLI，存在一个全机共享的目录里，启动时与每小时检查上游最新版。

校验按来源分级：GitHub Releases 用 asset `digest`，npm 用 `dist.integrity`，其它用 manifest 的 `checksum` 或 `sha256`；**x.ai 与 ACP registry 不发布哈希，只能依赖 TLS**。安装后写一份 `openbot-install.json` 记录每个文件的 SHA-256，每次启动校验并与 `--version` 比对，最新且校验通过的才运行。

另外三个机制：

- `provider-runtime-blocklist.json` 只压掉"提供更新"这个动作，不卸载已装版本，读不到清单则什么都不拦；
- 每 6 小时一次 canary，用免费模型跑一致性用例，连续两次失败自动开 issue；
- 激活失败时删除被拒产物并保留旧运行时，因此坏版本不会让工具直接不可用。

自建时这一层最容易被低估：你要为十几个上游的发布节奏负责，而破坏性变更不会提前通知。

## 八、如果你用兼容端点：一个具体的入口

"自定义 OpenAI 兼容端点"在这里是一个正式入口：填入 base URL 与密钥，兼容层就能当 provider 用，模型下拉会出现它列出的模型。接入第三方兼容网关（例如 4SAPI 这类中转服务）走的就是这一项——它只改模型请求出口，不动 CLI 登录态，因此不改变"凭证归 CLI"的规则。

自建时要提前想清楚的一点也在这里：兼容端点一旦进来，模型目录、用量上报、失败分类都会多一份"外部形状不同"的分支——兼容层通常只实现 `/chat/completions` 这类路径，`model/list`、用量查询、`classifyFailure` 依赖的错误语义往往要单独适配。OpenBot 用"driver 覆盖"容纳了这类差异；如果你的系统没有这层，兼容端点会变成散落在各处的 if。

## 九、自建时的对照清单

把上面几节压成一份可以拿去评审的清单：

1. **凭证归属**：第三方 CLI 的登录态是否仍由它自己管？只在"用户手填秘密"处做加密存储，还是把两者混在一处？
2. **模型目录**：读取流程是否无 provider 分支，差异是否显式登记？
3. **会话主键**：本地会话与 provider session 是否分成两层，切换是否只是重新绑定？
4. **容错边界**：failover 是否只发生在"尚未产生副作用"的轮次？半途失败是重试还是直接报错？
5. **用量查询**：是否有并发与退避控制，是否知道哪些端点与用户共享配额？
6. **版本治理**：CLI 由谁更新、按什么校验、坏版本怎么回滚、有没有自动化 canary？
7. **兼容端点**：外部兼容层带来的差异，是否有单独的适配层。

前四条决定"能不能稳"，后三条决定"半年后还能不能维护"。

## 十、限制与结论

这套设计的核心不是"支持很多模型"，而是**把多供应商的不确定性收进四个有名字的机制里**：凭证归 CLI、目录统一读、failover 只在未动手时发生、用量读取加门。每条都有明确的失效条件，也都写在文档里，这比"支持 N 家"更有参考价值。

三条判断：

- 如果你只需要接一两家 API，这套编排层的成本大于收益——它存在的理由正是 CLI 生态的碎片化。
- 如果你要接的是 CLI 而不是 API，那么"托管下载 + 哈希校验 + blocklist + canary"这四件套基本是必需品，不是加分项。
- 如果你的容错需求包含"半途换模型"，现成的这套语义不适用，需要自己设计状态迁移。

最后提醒一句：本文描述的是仓库文档所载的机制，没有在真机上逐条复现。要把它当作事实使用，最好在自己的环境里验证两三处——尤其是 failover 的"未动手"判定和用量读取的退避曲线，这两处最容易与文档字面有出入。

## 参考来源

| 来源 | 链接 | 检索日期 |
| --- | --- | --- |
| Providers 架构文档（模型目录、CLI 更新与校验、备份模型与 failover、用量读取节流、自定义 agent、兼容端点） | https://raw.githubusercontent.com/nightly-labs/openbot/main/docs/architecture/providers.md | 2026-11-12 |
| 仓库 README（provider 清单、自定义兼容端点与本地模型检测、凭证目录、会话保留） | https://raw.githubusercontent.com/nightly-labs/openbot/main/README.md | 2026-11-12 |
| 结构与状态归属（模型目录与 provider session 绑定的位置） | https://raw.githubusercontent.com/nightly-labs/openbot/main/docs/architecture/structure.md | 2026-11-12 |
