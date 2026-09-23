---
title: "Kimi K3 的 KVV 测评怎么跑：先预检再跑基准"
category: 人工智能
tags:
  - Kimi K3
  - 模型评测
  - API 测试
description: "从环境配置、K3 参数模式、预检套件到 OCRBench 和 MMMU Pro Vision，说明如何运行 Kimi Vendor Verifier，并正确记录未通过预检的接口。"
---

# Kimi K3 的 KVV 测评怎么跑：先预检再跑基准

给一个 Kimi K3 接口跑 OCRBench、MMMU Pro Vision 或 Agent 基准之前，最容易被忽略的一步是先验证 API 契约。一个端点可能在普通聊天中表现正常，却会在动态工具、非法参数拒绝、流式 reasoning chunk、`response_format` 或超长请求上偏离预期。直接把这种端点送进榜单，最后拿到的分数很难解释：它反映的是模型能力、实现差异，还是接口根本没有按契约工作？

Kimi Vendor Verifier，简称 KVV，是 MoonshotAI 开源的 Kimi 模型供应商验证项目。它不只包含几个基准脚本，还包含一组用于检查原始 Chat Completions API 行为的预检。本文按 K3 的流程说明如何配置 KVV、怎样选择官方与开源两种 thinking 参数格式、预检通过后如何运行正式 benchmark，以及预检失败时应该怎样报告。

这里的重点不是公布某个服务的排名，也不是把单次结果推广成模型结论，而是建立一条可复核的测评链路。文中数据来自 2026-08-21 对一个待测 endpoint 的预检记录，目的是展示失败报告应包含什么，不构成对其他接口的判断。

## KVV 测的不是一个分数，而是两层问题

KVV 的工作可以分成两层。

第一层是 API 契约预检。它验证请求参数是否按规则约束、工具调用 JSON Schema 是否被正确处理、K3 特有字段能否按约定工作、流式响应中的 prompt token 统计是否可信。这一层的输出通常是通过、失败、跳过和重试记录。

第二层才是能力 benchmark。K3 相关流程中常见的项目包括 OCRBench、MMMU Pro Vision、BEAM (1M) 和 DeepSWE。它们分别覆盖文字识别、多模态理解、长上下文记忆与多步编码 Agent 能力。K3 不再使用 AIME 2025 作为官方 K3 评测组合，并增加了 `reasoning_effort` 配置。

两层不能交换顺序。预检的作用不是“多跑一遍保证分数好看”，而是确定被测 API 能否按指定协议接收请求、返回结果并暴露可核验字段。若参数被静默接受、动态工具丢失、流式推理块不完整，再高的 benchmark 输出也无法和同一标准下的其他结果公平比较。

流程可以概括为：

```text
准备环境和 endpoint 配置
        ↓
参数约束、工具 Schema、K3 特性、prompt token 预检
        ↓
预检全部满足既定门槛
        ↓
OCRBench / MMMU Pro Vision / BEAM / DeepSWE
        ↓
保存日志、环境、命令和结果，形成可审计报告
```

如果预检不通过，正确的动作是修复实现并重跑预检，而不是跳过失败项直接运行 benchmark。报告中可以明确写“未进入正式评分”，这比拿一个不可比的数字更有技术价值。

## 适用场景、边界和准备条件

本文适合以下读者：维护 Kimi K3 兼容部署的人、需要验收第三方 API 的团队，以及要对两个 endpoint 使用同一套测试方法的人。它假设你拥有待测接口的合法访问权限，并且可以在一台具备网络访问条件的机器上执行 Python 工具。

环境准备时至少需要：

```text
Git
Python 3.10 或更高版本
uv
可访问的 Kimi K3 API Key
待测 endpoint 的 base URL
足够长的网络超时与可写入的日志目录
```

KVV 会下载评测数据并运行较长的任务，网络、磁盘和并发连接都需要预留空间。不要在共享的生产业务机器上直接以高并发运行，也不要把 `.env`、评测日志或完整请求体上传到公开仓库，里面可能包含 Key、内部地址、题目数据或请求上下文。

官方仓库、可用 suite 和命令以运行当天的版本为准。本文的操作依据 [MoonshotAI/Kimi-Vendor-Verifier 官方仓库](https://github.com/MoonshotAI/Kimi-Vendor-Verifier?tab=readme-ov-file) 中的中文说明；执行前建议把当前 commit 写进报告，避免日后命令和测试集发生变化后无法复现。

## 1. 获取项目并固定运行环境

在 PowerShell 中进入一个用于评测的空目录，再克隆仓库：

```powershell
git clone https://github.com/MoonshotAI/Kimi-Vendor-Verifier.git
Set-Location Kimi-Vendor-Verifier
uv sync
uv pip install -e .
```

`uv sync` 根据项目锁定信息安装依赖，`uv pip install -e .` 将当前项目以可编辑形式安装。两条命令都成功后，先记录基础信息：

```powershell
git rev-parse HEAD
uv --version
python --version
```

建议把输出写入 `reports/<timestamp>/environment.txt`。不要只记录“已安装 KVV”，因为 Python 版本、锁文件和仓库 commit 都可能影响依赖解析、日志格式或 case 集合。

KVV 既支持 Kimi 官方 API 形式，也支持面向开源部署的参数形式。两类配置不能只改 Base URL 就混用，尤其是 thinking 模式的请求结构不同。最方便的做法是复制仓库中的 `.env.example` 为 `.env`，再填入最小配置：

```text
KIMI_API_KEY=replace-with-your-key
KIMI_BASE_URL=https://your-endpoint.example.com/v1
MODEL_NAME=kimi-k3
THINK_MODE=opensource
```

其中 `KIMI_BASE_URL` 应是 API base URL，而不是具体的 `/chat/completions` 地址。工具会自行拼接请求路径；多写一层路径可能导致 404，也可能意外请求到另一个兼容接口。`.env` 只应存于本机或受控密钥系统，并加入 `.gitignore`。

## 2. 先选对 thinking 参数格式

`THINK_MODE` 描述的是请求参数格式，不是一个主观的“是否需要思考”的开关。选错后，端点可能仍返回文本，却没有按预期收到 K3 的 thinking 配置，后续结果就没有可比性。

对于 Kimi 官方 API，KVV 使用 `kimi` 模式：

| 命令状态 | 发送的 `extra_body` |
| --- | --- |
| 不启用 thinking | `{"thinking": {"type": "disabled"}}` |
| 启用 thinking | `{"thinking": {"type": "enabled"}}` |
| `reasoning_effort=low` | `{"thinking": {"type": "enabled", "keep": "all", "effort": "low"}}` |
| `reasoning_effort=high` | `{"thinking": {"type": "enabled", "keep": "all", "effort": "high"}}` |
| `reasoning_effort=max` | `{"thinking": {"type": "enabled", "keep": "all", "effort": "max"}}` |

对于 vLLM、SGLang、KTransformers 等开源部署，KVV 使用 `opensource` 模式：

| 命令状态 | 发送的 `extra_body` |
| --- | --- |
| 不启用 thinking | `{"chat_template_kwargs": {"thinking": false}}` |
| 启用 thinking | `{"chat_template_kwargs": {"thinking": true}}` |
| `reasoning_effort=low` | `{"chat_template_kwargs": {"thinking": true, "preserve_thinking": true, "thinking_effort": "low"}}` |
| `reasoning_effort=high` | `{"chat_template_kwargs": {"thinking": true, "preserve_thinking": true, "thinking_effort": "high"}}` |
| `reasoning_effort=max` | `{"chat_template_kwargs": {"thinking": true, "preserve_thinking": true, "thinking_effort": "max"}}` |

因此，某个服务自称“Kimi 兼容”不足以决定模式。应该根据实际部署文档、网关实现和接收的请求格式选择 `kimi` 或 `opensource`。不确定时，先执行预检并保留原始失败信息；不要仅因普通对话可用就假设参数形状正确。

## 3. 预检一：不可变 API 参数约束

正式 benchmark 之前，先运行参数约束验证。它会检查端点是否按 K3 契约正确处理采样参数，例如 `temperature`、`top_p`、`presence_penalty`、`frequency_penalty` 和 `n`。

对于开源部署，可以使用：

```powershell
uv run --env-file .env pytest tests/params `
  --smoke-model kimi-k3 `
  --think-mode opensource `
  -v
```

若测 Kimi 官方 API，将 `--think-mode opensource` 换成 `--think-mode kimi`，并根据实际模型名替换 `--smoke-model`。运行日志要完整保存，因为这组 case 的重点包括“非法值是否被拒绝”。

这里有一个常见误区：认为服务器尽量宽容地接收参数是一种兼容性。对于有明确契约的评测，非法或不适用的参数应按预期被拒绝；静默接受并随意忽略，会让调用方误以为某个设置已经生效。预检失败时，应记录具体是非 thinking 还是 thinking 请求、哪个字段、请求值和预期 HTTP 行为出了差异。

不要用“最终回答看起来没问题”覆盖这种失败。参数约束影响可复现性，尤其是要对不同供应商跑同一 benchmark 时。如果一个端点默默接受 `n=2` 或不允许的采样组合，另一个端点严格拒绝，两者已经不处在同一个请求条件上。

## 4. 预检二：工具调用 JSON Schema

`tests/tool_call_json_schema/` 验证工具参数 Schema 的处理。KVV 会把 walle 的有效 MFJS case 放入 `tools[].function.parameters`，强制模型触发工具调用，然后在本地用 `jsonschema` 验证 `tool_calls[].function.arguments`。每个 case 会分别测试 `stream=false` 和 `stream=true`，流式分片会先重新组装再校验。

示例命令如下：

```powershell
uv run --env-file .env pytest -n 4 tests/tool_call_json_schema `
  --smoke-model kimi-k3 `
  --think-mode opensource `
  --thinking `
  --reruns 3 `
  --reruns-delay 2 `
  --tool-json-report tool-call-schema-report.json `
  -ra -v
```

这里刻意省略了显式 `--base-url` 与 `--api-key`，让 `uv run --env-file .env` 为启动的测试进程注入配置。PowerShell 的 `$env:KIMI_BASE_URL` 不会自动读取 `.env` 文件；如果要在命令行显式使用它，必须先在当前会话中设置同名环境变量。一个空的 `--base-url` 可能导致工具回落到默认地址，最终测的不是目标端点。

`--reruns 3` 只适用于网络抖动或偶发问题的识别，不能把每条失败都解释为“重试后总会通过”。报告应保留初次失败、重试次数和最终状态。特别是流式工具调用，参数分片拼接错误、结束事件缺失与普通网络超时的排查方向不同。

## 5. 预检三：K3 feature contract

K3 feature suite 是 KVV 里最容易暴露兼容层差异的一部分，覆盖 dynamic tools、`response_format`、`tool_choice` 和 thinking effort 等特性：

```powershell
uv run --env-file .env pytest tests/k3_features `
  --smoke-model kimi-k3 `
  -ra -v
```

同样需要注意 PowerShell 对 `.env` 的处理。只要通过 `uv run --env-file .env` 启动，测试进程会拿到配置；不要再用当前会话中未设置的 `$env:` 变量覆盖它。

dynamic tools 不是普通 `tools` 参数的同义词。它测试的是放在 `messages[].tools` 的非标准动态工具声明是否被保留、是否在需要调用时返回 `tool_calls`，以及非法声明能否被一致地拒绝。若请求要求工具调用却频繁得到 `finish_reason=stop`，不能把它记成“模型选择不调用工具”后跳过；它是 feature contract 需要报告的行为差异。

`response_format` 与 `tool_choice` 的测试也不只是看成功请求。更关键的是，非法、缺失或互相冲突的输入是否按规则拒绝。兼容层为了“尽量不报错”而接受无效 Schema，可能会在真实业务里把错误拖到更晚的阶段，导致调用方拿到无法解析的结构化结果。

thinking effort 则需要同时看请求是否接收、输出是否包含预期的 reasoning 行为，以及在流式情况下 reasoning chunks 的顺序和完整性。只看到最终正文，不足以证明保留推理、effort 优先级或流式片段的实现符合约定。

## 6. 预检四：prompt token 计数

`tests/prompt_tokens/` 用固定 case 以流式方式发送请求，再将 endpoint 上报的 `usage.prompt_tokens` 与 case 中的期望常量进行精确比较：

```powershell
uv run --env-file .env pytest tests/prompt_tokens `
  --smoke-model kimi-k3 `
  -ra -v
```

这组测试的意义不在于“Token 越少越好”。它检查供应商上报的 prompt token 是否可被验证。若使用量字段不能稳定对应固定输入，成本核算、配额控制、限流预估和上下文调试都会失去可靠依据。

有时这组测试会因为上游测试数据无法下载而无法启动。例如，本次记录中仓库所需的 Git LFS fixture 受到配额限制，prompt-token suite 没有运行。此时状态应写为 `SKIP` 或 `BLOCKED`，注明外部依赖原因；不能把未运行当作通过，也不应该凭其他 suite 的结果替它下结论。

## 7. 预检通过后，才运行 OCRBench 与 MMMU Pro Vision

K3 采用 thinking、保留 reasoning 和 `reasoning_effort` 的组合。下面命令展示 `max` effort 的示例；实际报告应明确写清模型名、mode、温度、Top-p、最大输出 Token、并发和客户端超时。

OCRBench：

```powershell
uv run --env-file .env python eval.py ocrbench `
  --model opensource/kimi-k3 `
  --max-tokens 16384 `
  --thinking `
  --think-mode opensource `
  --stream `
  --max-connections 50 `
  --temperature 1.0 `
  --top-p 0.95 `
  --thinking-effort max
```

MMMU Pro Vision：

```powershell
uv run --env-file .env python eval.py mmmu `
  --model opensource/kimi-k3 `
  --max-tokens 98304 `
  --thinking `
  --think-mode opensource `
  --stream `
  --max-connections 50 `
  --temperature 1.0 `
  --top-p 0.95 `
  --thinking-effort max
```

上述模型写法 `opensource/kimi-k3` 与 `--think-mode opensource` 是一组配置。如果测官方 API，则应按 KVV 文档改用对应的模型前缀和 `kimi` mode。不要把模型显示名直接复制到命令里却忘记同时变更 think mode；这会让报告无法说明实际发送的请求格式。

`--max-connections 50` 是示例并发值，并不适合所有端点。出现大量 `429`、`RemoteProtocolError`、TLS 读取失败或网关排队时，应先降低并发、保持其他变量不变，再重新运行并记录变化。不要仅因为某一次较高并发跑得更快，就把它当作接口能力结论。

评测完成后可查看 Inspect 日志：

```powershell
uv run inspect view
```

日志通常位于项目的 `logs/` 目录。若评测中断，可以用仓库提供的恢复命令继续相同日志：

```powershell
uv run inspect eval-retry logs/<log-file>.eval
```

恢复时要确认 `.env`、模型、版本和思考参数没有变化，否则同一个结果文件可能混入两组条件。

## 8. BEAM 和 DeepSWE 不能被 OCRBench 替代

OCRBench 与 MMMU Pro Vision 的输出不能代表长上下文或 Agent 表现。KVV 中的 BEAM (1M) 面向 1M token 长对话的长期记忆；DeepSWE 是基于开源 DeepSWE 和 Pier 平台的多步编码 Agent 流程，包含 113 道 coding-agent 题目，并要求相应版本的 `kimi-code` 注册到 Pier agent。

它们既有不同的任务形式，也有独立的运行时间、凭据和环境依赖。因此报告应把它们列为独立的“已运行”“未运行”或“因预检失败未启动”，不能用 OCRBench 或 MMMU 的一个数字替代。对于长任务尤其要记录客户端超时、流式传输、限流和恢复策略；这些条件本身会影响是否能够完成一次评测。

如果当前目标只是确认部署能否进入 KVV 正式测评，先跑预检和一个小规模 OCRBench smoke run 已足够发现大部分配置问题。只有在 gate 满足且环境稳定后，再投入较长的 BEAM 或 DeepSWE 任务，才能避免把资源花在一个还没满足契约的端点上。

## 9. 一个预检失败案例该怎样写

下面是 2026-08-21 对 4SAPI 的实际 endpoint `https://4sapi.org/v1`、模型 `kimi-k3`、`THINK_MODE=opensource` 进行的记录。它用于说明报告格式和门槛处理，并不外推到其他服务、日期或配置。

本次目标 effort 为 `max`，两组已完成预检的结果如下：

| Suite | 通过 | 失败 | 跳过 | 重试 | 用时 |
| --- | ---: | ---: | ---: | ---: | ---: |
| API 参数约束 | 8 | 10 | 0 | 0 | 551.23 秒 |
| K3 feature contract | 29 | 79 | 18 | 18 | 1478.22 秒 |

这不是“分数不高”，而是契约预检未满足。本次没有启动 OCRBench、MMMU Pro Vision、BEAM 或 DeepSWE 的正式评分。Tool-call JSON Schema suite 也没有单独继续运行，因为更广泛的 K3 feature failures 已足以作出 gate 判断；prompt-token suite 则受 Git LFS fixture 下载配额限制而未运行。

从日志归纳出的主要失败类型包括：

1. 非 thinking 和 thinking 请求中，多组非法 `temperature`、`top_p`、`presence_penalty`、`frequency_penalty` 值没有按契约被拒绝；thinking 请求中的 `n=2` 也存在不符合预期的行为。
2. 声明在 `messages[].tools` 的 dynamic tools 没有稳定生效。要求发起工具调用的请求常以 `finish_reason=stop` 结束，而不是返回 `tool_calls`。
3. malformed dynamic-tool 声明没有被一致拒绝；部分非法或不完整的 `response_format` Schema 也被接受。
4. `tool_choice` 的非法组合和值校验不一致，thinking effort、reasoning chunks 和优先级行为只部分兼容。
5. 运行中出现间歇性的 `TLS UNEXPECTED_EOF_WHILE_READING`，说明网络或连接层也需要单独排查。

报告的正确结论应写成：该 endpoint 在上述日期、模型和 `opensource` 请求参数格式下未通过已完成的 KVV 预检，不具备进入本次正式 KVV 评分的条件。它不等于“Kimi K3 的模型能力为零”，也不等于“所有请求都失败”；它只说明要在同一验证框架下提交可比 benchmark 之前，接口实现仍需修复。

建议的修复与复测顺序是：先落实不可变采样参数的拒绝规则；再保留并校验 `messages[].tools`；之后强化 `response_format` 与 `tool_choice` 的非法输入处理；接着修正 thinking effort 和流式 reasoning 输出；最后排查 TLS 连接稳定性。每完成一项都重跑相关 suite，直到预检满足门槛，再启动正式 benchmark。

## 10. 让报告能被别人复查

一份能支撑结论的 KVV 报告，至少应含有以下字段：

```text
测试日期和时区
KVV 仓库 URL 与 Git commit
Python、uv 和主要依赖版本
endpoint 的脱敏主机名与 base URL 路径
模型名、THINK_MODE、thinking effort
每个 suite 的完整命令
并发、超时、stream、temperature、top_p、max-tokens
通过、失败、跳过、重试数量和耗时
原始 pytest / Inspect 日志的路径或受控存档
未运行项及其具体原因
结论的适用范围与限制
```

不要只截一张“全部通过”或“最终分数”的图片。图片无法显示被测版本、实际命令和失败 case，也难以在升级后做差异比较。比较两个 endpoint 时，先确保两侧测试集版本、`THINK_MODE`、模型标识、effort、并发、超时和重试策略一致；任何不同都应写在表格旁边。

同样要避免用单次 benchmark 分数比较“谁更强”。OCRBench、MMMU、BEAM 与 DeepSWE 衡量的是不同任务，且结果会受数据版本、采样、网络错误、服务端限流和运行方式影响。这套流程关注先让接口满足可测条件，再在条件一致时讨论结果。

## 常见排错清单

**Base URL 多拼或少拼了路径。** `.env` 应指向 API base URL。不要把完整请求 path 填入后又让客户端自动拼接。

**PowerShell 没有导入 `.env`。** `uv run --env-file .env` 会让启动的程序读取文件，但不会把值回填到当前 PowerShell 的 `$env:`。显式使用 `$env:KIMI_BASE_URL` 前，确认当前会话已设置该变量。

**用错 `THINK_MODE`。** 先确认 endpoint 接受的是 Kimi 的 `thinking` body，还是开源框架的 `chat_template_kwargs`。普通聊天成功不能替代这一核验。

**把 `SKIP` 当 `PASS`。** 测试数据未下载、外部服务不可用或 suite 被 gate 拦住时，应该明确写未运行原因。

**高并发掩盖连接问题。** 降低 `--max-connections` 后单独复现 TLS 或 `429`，区分服务限流、客户端连接池和网关读写中断。

**预检失败后仍发布正式分数。** 未通过契约门槛时，结果最多是内部诊断数据，不能与完整通过预检的供应商成绩并列比较。

## 用 smoke run 验证“命令能跑”，不要把它当正式成绩

预检全部满足门槛后，正式 benchmark 仍不应第一次就用最大并发和最长运行时间。先做一轮小规模 smoke run，确认模型名称、thinking 参数、认证、图像或数据加载、流式传输和日志落盘都在目标 endpoint 上生效。这个阶段的输出只用于验证环境，不应当作 OCRBench、MMMU 或 DeepSWE 的正式成绩。

KVV 对长推理任务提供了 `--stream`、`--client-timeout`、`--max-connections` 与 `--epochs` 等参数。默认客户端超时很长，实际运行还要检查服务端、网关和反向代理的超时是否足够；若出现 `429`、`ReadError` 或 `RemoteProtocolError`，先降低并发并保留错误类型，再重跑小样本。[KVV 中文说明](https://github.com/MoonshotAI/Kimi-Vendor-Verifier/blob/main/README_zh.md?plain=1)将这些网络类错误和模型输出格式问题区分处理，不能把两类错误都归为“模型不稳定”。

一次可复用的 smoke 记录可以包括：

```text
KVV commit：
模型与 THINK_MODE：
命令行：
样本或 epochs 范围：
并发与超时：
是否使用 stream：
请求是否到达预期 endpoint：
日志文件路径：
网络异常与重试：
是否具备启动正式 benchmark 的条件：
```

只有 smoke run、预检和环境记录都一致时，再把同一命令参数扩大到正式规模。若在 smoke 阶段改过并发、endpoint 或 thinking 参数，正式报告必须使用改后的值，而不是复制原本计划中的配置。

## 不同 suite 的失败需要使用不同口径

KVV 中的 API 参数、tool-call JSON Schema、K3 feature 和 prompt token suite 都是验证契约；OCRBench、MMMU Pro Vision、BEAM 和 DeepSWE 则是能力评测。前者失败时通常要定位请求、响应或协议行为，不能用更高的 benchmark 分数抵消；后者失败或中断时，则需要区分模型输出错误、数据加载问题、网络限制、并发限制和 scorer 失败。

推荐在报告中用下面的状态语义：

| 状态 | 含义 | 可以得出的结论 |
| --- | --- | --- |
| `PASS` | 所有该 suite 的预定义断言通过 | 该 suite 在当前条件下满足要求 |
| `FAIL` | 可复现的契约或结果断言失败 | 记录失败 case，不进入对应 gate |
| `SKIP` | 该项按规则不适用或未被选择 | 不对该能力作结论 |
| `BLOCKED` | 外部数据、权限或环境阻止运行 | 说明阻塞条件，不当作通过 |
| `INFRA_ERROR` | 网络、服务或执行环境使结果无法判断 | 保存错误与重试信息，先修环境 |

这一分类使“没有分数”也有明确含义。例如因 Git LFS 配额无法获得 fixture 而未跑 prompt token，是 `BLOCKED`；因 API 接受了契约要求拒绝的参数而失败，是 `FAIL`；因当前测试目标不包含 BEAM，是 `SKIP`。它们不能放在同一个“未通过”里，也不应被计算进模型能力得分。

## 比较两次 KVV 运行前先核对可比性

想确认修复是否有效，或比较两个 endpoint，不能只比测试汇总数字。至少要先核对 KVV commit、测试数据版本、模型 ID、`THINK_MODE`、thinking effort、temperature、top-p、max tokens、stream、并发、超时和重试策略。任何一个不同，都可能让 case 行为或资源条件发生变化。

特别是 K3 的 `reasoning_effort`，官方 README 为 OCRBench 与 MMMU Pro Vision列出 low、high、max 三组配置。报告应将 effort 当成实验条件，而不是隐藏在“开启 thinking”这个总称中。若一个端点用 `max`，另一个用 `low`，即使最终题目和模型名相同，也不应把两个结果并列解释成同等条件下的差异。

对失败修复的回归报告，可以采用“前次失败 case -> 修复版本 -> 本次结果 -> 未改变的条件”四列格式。这样读者能看到改动是否只消除了目标错误，还是同时改变了模型路由、并发或参数格式，避免把多项变化误判为某一个修复的效果。

## 结论

KVV 的价值不只在于给 Kimi K3 跑出一个 benchmark 数字，更在于把“这个接口是否按同一规则工作”放到能力比较之前。先校验参数、工具 Schema、K3 feature 和 prompt token，再运行 OCRBench、MMMU Pro Vision、BEAM 或 DeepSWE，才能让结果具备可解释性。

预检失败也不是无效结果。只要保留命令、环境、失败 case、网络错误和未运行原因，它就能直接指导接口修复，并为下一次复测提供稳定基线。等预检通过后，在相同模型、参数和版本下运行正式 benchmark，才适合讨论可比的评测结果。

可执行命令与参数定义以 [KVV 中文说明](https://github.com/MoonshotAI/Kimi-Vendor-Verifier/blob/main/README_zh.md?plain=1) 为准；项目更新后，应先核对仓库版本与测试说明，再复用这套流程。
