---
title: "在 Codex 和 Claude Code 里接 kimi-k3.1-pro：配置、验证与报错排查"
category: 人工智能
tags:
  - Kimi K3.1 Pro
  - Codex
  - Claude Code
description: "说明如何用 Responses API 把 Codex 直连 kimi-k3.1-pro、用 Anthropic 兼容端点接入 Claude Code，并给出 401、404、模型名不匹配等报错的分层排查方法。"
brand_mode: example
---

# 在 Codex 和 Claude Code 里接 kimi-k3.1-pro：配置、验证与报错排查

把 K3.1 Pro 接进编码 Agent，是这次发布后最实际的一步。麻烦在于两个工具走的是两套协议：Codex 用 Responses API，Claude Code 用 Anthropic Messages API 兼容端点。配置项名字不同、失败表现也不同，把一份配置复制到另一边通常只会得到 401 或 404。

本文只解决一个问题：让 `kimi-k3.1-pro` 在 Codex（桌面端与 CLI）和 Claude Code 里稳定跑起来，并且在出问题时知道该看哪一层。文中配置与端点均引自 Kimi 官方文档，检索日期 2026-10-26，相关页面最后更新 2026-10-23；工具侧界面可能随版本变化，以你实际安装的版本为准。

## 两条路径的三处根本差异

先看清差异，能省掉大部分试错。

| 对比项 | Codex | Claude Code |
| --- | --- | --- |
| 协议 | Responses API | Anthropic Messages API 兼容端点 |
| 端点 | `https://api.moonshot.cn/v1` | `https://api.moonshot.cn/anthropic` |
| 配置文件 | `~/.codex/config.toml` | `~/.claude/settings.json` 的 `env` |
| 凭据传递 | 环境变量 `KIMI_API_KEY` | `ANTHROPIC_AUTH_TOKEN` |
| 上下文窗口 | 由 `model_context_window` 显式声明 | 由 `CLAUDE_CODE_AUTO_COMPACT_WINDOW` 控制压缩阈值 |
| 推理强度 | 请求参数 `reasoning_effort` | 环境变量 `CLAUDE_CODE_EFFORT_LEVEL` |
| Key 是否写进配置文件 | 不写，只写变量名 | 写，值就是 Key（明文） |

最后一行差异是安全问题，不是配置问题：Claude Code 的写法要求把 Key 放进 `settings.json`，这个文件不能进 git。

## 前置条件

三件事必须先完成，否则后面的排查都会指向错误的层。

1. Codex：按官方文档安装 Codex CLI，并至少启动过一次；Claude Code：已安装并完成一次初始化。
2. 在 Kimi 开放平台创建 API Key。**注意中国站 `platform.kimi.com` 与国际站 `platform.kimi.ai` 的账户、余额和 Key 完全独立**，混用会返回 401。对应的端点也不同：中国站是 `https://api.moonshot.cn/v1`，国际站是 `https://api.moonshot.ai/v1`（Anthropic 兼容端点同理，分别为 `.../anthropic`）。下面的配置以中国站为例，换站时只改域名，路径不变。
3. 账户状态满足 K3.1 Pro 的访问条件：官方要求完成充值（最低 10 元）后解锁该模型，新用户注册赠送的 15 元代金券**不能**用于 K3.1 Pro。这是"配置全对但请求被拒"的常见原因之一。

来源：[Kimi K3.1 Pro 快速开始](https://platform.kimi.com/docs/guide/kimi-k3-1-pro-quickstart)，最后更新 2026-10-23。

另有一条仅用于交叉核对的侧边信息：发布当天 K3.1 Pro 已同步在 OpenRouter 与 OpenCode 上架（OpenRouter 标的国际站价为输入 $2 / 输出 $10 每百万 tokens），`llm` 这类通用命令行客户端也能直接对着国际站端点调用，并切换 `reasoning_effort` 的低/高/最高三档。这些不是你需要依赖的接入方式，但当你在官方入口拿到奇怪的报错时，可以拿第三方入口的结果做对比，快速判断问题在模型侧还是在你的配置侧。

## 路径 A：Codex 直连 Responses API

Codex 能直连 K3.1 Pro，是因为官方原生实现了 Responses API，不需要协议转换或本地代理。

第一步，把 Key 放进环境变量。不要写进 `config.toml`，也不要在命令行里明文传递（会进 shell 历史）。

```powershell
$env:KIMI_API_KEY="YOUR_KIMI_API_KEY"
```

长期生效请用系统环境变量面板添加同名变量。

第二步，编辑 `~/.codex/config.toml`（Windows 为 `%USERPROFILE%\.codex\config.toml`）：

```toml
model = "kimi-k3.1-pro"
model_provider = "kimi"
model_context_window = 1048576

[model_providers.kimi]
name = "Kimi"
base_url = "https://api.moonshot.cn/v1"
env_key = "KIMI_API_KEY"
wire_api = "responses"
```

四个字段各自负责一件事，缺一个就会出现"能聊天但 Agent 不好用"这类半失败状态。

| 字段 | 作用 | 缺失或写错的后果 |
| --- | --- | --- |
| `wire_api = "responses"` | 决定用 Responses 协议通信 | 走错协议，请求被拒 |
| `env_key = "KIMI_API_KEY"` | 声明读取哪个环境变量 | 拿不到 Key，401 |
| `model_context_window = 1048576` | 与模型 1M 上下文对齐 | Codex 用默认元数据，长任务表现异常 |
| `base_url` 结尾的 `/v1` | 拼接出 `/v1/responses` | 404 |

来源：[在 Codex 中使用 Kimi K3.1 Pro](https://platform.kimi.com/docs/guide/codex-kimi)，最后更新 2026-10-23。

第三步，分两级验证。先发一句 `你是谁`，确认基本连接；再发一个会触发工具调用的任务，例如 `检查这个仓库并总结它的结构。`。只有第二级通过，才说明模型调用和工具调用都正常。

Codex Desktop 有一个容易误判的细节：模型选择器里可能显示**自定义**，实际请求仍是你配置的 `kimi-k3.1-pro`。判断依据是请求日志和响应，而不是界面上的标签。

另外，Codex CLI 目前只支持文本和图片输入，没有原生视频通道。这是 CLI 输入层的限制，不是模型的能力限制——直接调 API 就能传视频。

## 路径 B：Claude Code 接 Anthropic 兼容端点

Claude Code 通过 `settings.json` 的 `env` 注入配置。下面的结构照官方示例写，模型位置统一用 `kimi-k3.1-pro`（官方示例里那几个位置写的是 `kimi-k3[1m]`，那是在接 K3，见下一节）：

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "https://api.moonshot.cn/anthropic",
    "ANTHROPIC_AUTH_TOKEN": "YOUR_MOONSHOT_API_KEY",
    "ANTHROPIC_MODEL": "kimi-k3.1-pro",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "kimi-k3.1-pro",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "kimi-k3.1-pro",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "kimi-k2.7-code",
    "ANTHROPIC_DEFAULT_FABLE_MODEL": "kimi-k3.1-pro",
    "CLAUDE_CODE_SUBAGENT_MODEL": "kimi-k3.1-pro",
    "CLAUDE_CODE_AUTO_COMPACT_WINDOW": "1000000",
    "CLAUDE_CODE_EFFORT_LEVEL": "max"
  }
}
```

为什么每个档位都要配？因为 Claude Code 内部按场景使用不同档位的模型：主对话、后台摘要、子 Agent 各走一个。只配 `ANTHROPIC_MODEL`，其余场景会静默失败或明显变差。

HAIKU 档在上面的配置里保留为 `kimi-k2.7-code`，这是官方示例的写法。这里有个额外条件：`kimi-k2.7-code` 强制开启思考，在 Claude Code 侧必须先打开 Thinking（macOS `Option+T`，Windows/Linux `Alt+T`），否则请求会被拒绝并返回 `400 invalid thinking: only type=enabled is allowed for this model`。不想处理这个前提，就把 HAIKU 档也换成 `kimi-k3.1-pro`，代价是这类轻量调用变贵。

| 变量 | 不配置的后果 |
| --- | --- |
| `ANTHROPIC_BASE_URL` | 请求发往 Anthropic 官方端点，鉴权失败 |
| `ANTHROPIC_AUTH_TOKEN` | 401 |
| `ANTHROPIC_MODEL` | 模型不存在 |
| 四个 `ANTHROPIC_DEFAULT_*_MODEL` | 对应档位任务失败 |
| `CLAUDE_CODE_SUBAGENT_MODEL` | 子任务失败或效果变差 |
| `CLAUDE_CODE_AUTO_COMPACT_WINDOW` | 过小会过早压缩丢上下文，过大报超限 |
| `CLAUDE_CODE_EFFORT_LEVEL` | 建议 `max`，调低可能影响复杂任务质量 |

来源：[在 Claude Code 中使用 Kimi](https://platform.kimi.com/docs/guide/claude-code-kimi)，最后更新 2026-10-23。

### 模型 ID 用 `kimi-k3.1-pro`，不要抄示例里的 `kimi-k3[1m]`

K3.1 Pro 的模型 ID 是 **`kimi-k3.1-pro`**，与官方模型列表、Codex 接入页写的一致，直接用即可。

需要留意的是官方 Claude Code 页面的示例配置：那几个位置写的是 `kimi-k3[1m]`，这是上一代 K3 的写法（`[1m]` 是上下文窗口标记）。用它接 K3.1 Pro，请求的就是另一个模型，拿不到 K3.1 Pro 的推理强度和能力，因此把每个 `ANTHROPIC_*_MODEL` 位置都换成 `kimi-k3.1-pro`。

改完先打一次端点验证，把“凭据问题”和“ID 问题”分开：

```bash
curl https://api.moonshot.cn/anthropic/v1/messages \
  --header "Authorization: Bearer $KIMI_API_KEY" \
  --header "Content-Type: application/json" \
  --data '{"model": "kimi-k3.1-pro", "max_tokens": 1, "messages": [{"role": "user", "content": "hi"}]}'
```

返回正常 JSON 说明端点和凭据可用、该 ID 在此端点被接受；返回 401 是凭据问题，返回模型不存在是 ID 问题。这两类错误必须分开处理，否则会在错误的层反复调整。

### 旧配置会覆盖新配置

这是 Claude Code 侧最高频的失败原因。`settings.json` 的 `env` 优先级高于终端里 export 的同名变量，所以只要文件里残留旧的 `ANTHROPIC_BASE_URL`，新配置就不会生效。

检查顺序：`settings.json` 的 `env` → `~/.zshrc` 或 `~/.bashrc` 中遗留的 `ANTHROPIC_*` export（Windows 用户查用户环境变量）→ 会话内 `/status` 显示的 Base URL 与 Model。

还要注意：如果之前同时设置了 `ANTHROPIC_API_KEY` 和 `ANTHROPIC_AUTH_TOKEN`，两者会冲突，需要删掉前者。

## 报错排查：先定位在哪一层

按"配置层 → 鉴权层 → 路由层 → 协议层 → 限流层"的顺序查，比逐个试快得多。

| 现象 | 最可能的原因 | 验证方法 |
| --- | --- | --- |
| 401 Unauthorized | Key 无效；Key 与 `base_url` 分属不同平台；环境变量没进到启动进程 | 用 curl 打同一端点；确认变量在启动终端里可见 |
| 404 `/v1/responses` | `base_url` 缺 `/v1`，或仍指向 `http://127.0.0.1:...` 这类本地代理 | 逐字符核对 `https://api.moonshot.cn/v1` |
| 400 `web_search.search_context_size is not supported` | 请求带了暂不支持的字段 | 移除该参数；Codex 默认请求不含此参数 |
| 模型不存在 / model not found | 模型 ID 拼写错误，或仍写着 K3 的 `kimi-k3[1m]` | 改用 `kimi-k3.1-pro` 后 curl 复测 |
| 后台任务或子 Agent 报错 | 缺 `ANTHROPIC_DEFAULT_HAIKU_MODEL` 等档位变量 | 对照上表补齐 |
| 429 Rate Limit | 触发并发、RPM、TPM 或 TPD 上限 | 看响应头 `X-RateLimit-*`，按提示退避 |
| 警告 Model metadata for kimi-k3.1-pro not found | 模型不在 Codex 内置目录，属预期 | 已在配置里声明 `model_context_window` 即可忽略 |
| 改了配置不生效 | Codex 只在启动时读 `config.toml`；或有工具在改写它 | 退出重启；检查是否有本地路由工具在覆盖配置 |

前四行来自官方两份接入文档，中间三行来自官方错误码页与限速页：

- [常见错误码说明](https://platform.kimi.com/docs/api/errors)
- [充值与限速](https://platform.kimi.com/docs/pricing/limits)

限速档位值得单独看一眼，因为很多"偶发 429"其实和充值额度有关：Tier0（累计充值 ¥0）只有 1 并发、3 RPM、500,000 TPM、1,500,000 TPD；Tier1（¥50）提到 15 并发、100 RPM、2,000,000 TPM。做 Agent 批量跑任务时，Tier0 的 TPD 很容易撞上。

## 成本侧的两个隐性开关

接好之后账单往往比预期高，常见原因是两个开关没设对。

**推理强度。** K3.1 Pro 的 `reasoning_effort` 默认是 `max`，短任务用 `max` 会白烧思考 token。建议按任务类型分档：格式转换、代码补全用 `low`；多步推理用 `high`；复杂重构和长链 Agent 用 `max`。注意档位会影响前缀缓存——官方明确提示切档会破坏缓存命中，所以档位应该在会话开始前定好，不要中途反复切。

**缓存写入。** K3 系列把缓存写入单独计费，`5m` 和 `1h` 两档价格不同。编码 Agent 的典型请求是"大前缀 + 小增量"，前缀能不能命中缓存，对单次成本的影响远大于输入单价本身。有长时间人工介入的会话，用 `1h` 档更省；连续追问不超过 5 分钟的，默认 `5m` 就够。

来源：[推理强度](https://platform.kimi.com/docs/guide/use-reasoning-effort)、[上下文缓存最佳实践](https://platform.kimi.com/docs/guide/context-caching)。

## 如果走统一网关

同时用多个模型供应商的团队，常见做法是通过聚合网关统一端点，把 `base_url` 换成网关地址。这时上面所有排查方法仍然适用，但要多验一件事：网关是否原样透传 `usage` 明细和思考字段。

以可替换实例来说，若通过 4SAPI 这类统一入口调用，模型名与参数支持范围以该平台控制台当期显示为准。要验证的最小集合是：响应里的模型字段与请求一致、`usage.prompt_tokens_details` 有缓存明细、多轮工具调用不会因为缺少 `reasoning_content` 而失败。三项都能过，再谈统一管理带来的运维收益。

## 限制与结论

两条路径的共同前提是：Key 与端点必须同平台，账户必须满足 K3.1 Pro 的访问条件，模型 ID 必须以平台列表为准。这三点与工具无关，先确认能少走一半弯路。

Codex 侧的关键配置是 `wire_api = "responses"` 与 `model_context_window`，前者决定能不能连通，后者决定 1M 上下文有没有被正确使用。Claude Code 侧的关键是配全所有档位变量，并清掉 `settings.json` 和 shell 里的旧配置——它的失败方式是静默的。

真正需要自己验证、不能照抄的部分有两个：一是官方 Claude Code 示例里的旧模型名（必须替换为 `kimi-k3.1-pro`），二是界面显示的模型标签（Codex Desktop 显示“自定义”属正常，以请求与响应为准）。

下一步建议：先跑通"简单问答 + 一次工具调用"两级验证，再压一个真实仓库任务，同时记录首 token 延迟、思考 token 占比和 `usage` 三段拆分。这三项数据决定了这次接入是省钱还是烧钱。
