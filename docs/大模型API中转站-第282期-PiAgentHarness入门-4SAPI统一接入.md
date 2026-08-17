---
title: "Pi Agent Harness 的模型、工具与上下文如何协作"
tags:
  - Pi
  - Agent开发
  - 工具调用
description: "Pi Agent Harness 同时管理模型请求、上下文和工具结果，初次阅读源码时容易把运行时职责混在一起。"
---
# Pi Agent Harness 的模型、工具与上下文如何协作
Pi Agent Harness 同时管理模型请求、上下文和工具结果，初次阅读源码时容易把运行时职责混在一起。本文从一次最小 Agent 循环出发，解释模型、工具、消息和续跑状态的边界，并给出适合继续扩展的观察点。文中只讨论可在本地复现的步骤，不把单次结果扩展成产品承诺；每个结论都标注前提、证据和无法覆盖的边界。读者可以先完成最小验证，再根据自己的版本、权限和数据补充实验，避免把配置示例误当成普遍结论。

这次的主角不是一个新模型，而是一个把模型、工具和工作流组织起来的开源项目：[@earendil-works/pi](https://github.com/earendil-works/pi)。如果把普通聊天应用理解为“发一条消息，收一段文本”，Pi 更接近一个可扩展的 Agent Harness：它负责维护上下文、发起工具调用、接收工具结果，再决定是否继续下一轮。


## 1. Pi 解决的不是“再包一层聊天框”

很多 AI 工具的最小循环是：

```text
用户问题 -> 模型 -> 文本回答
```

这在问答场景已经够用。一旦任务需要读文件、调用命令、整理结果或连续修改上下文，流程会变成：

```text
用户目标
  -> 模型判断下一步
  -> 发起工具调用
  -> 本地或外部系统执行工具
  -> 工具结果回到上下文
  -> 模型继续判断
  -> 输出阶段性结果或再次调用工具
```

这里有三个容易被忽略的问题：

1. 模型响应不是终点，它可能只是下一次工具调用的指令。
2. 工具结果必须以模型能够识别的消息结构写回上下文。
3. 长任务需要保存状态，才能中断、继续和复盘。

Pi 的价值就是把这些交接点抽成可以替换和扩展的运行时。它不强迫你采用某个固定的工作流，也不把所有功能都塞进一个巨大的系统提示词里。

## 2. 仓库里最值得先看懂的四个包

Pi 仓库是一个 monorepo，官方 README 把 Agent Harness 拆成多个包。先把职责分开，后面看代码或文档会轻松很多。

| 包 | 主要职责 | 什么时候会直接接触 |
| --- | --- | --- |
| `@earendil-works/pi-ai` | 统一多模型 API、认证、流式事件、工具调用、Token 与成本跟踪 | 你要在自己的 TypeScript 程序里调模型 |
| `@earendil-works/pi-agent-core` | Agent 运行时、状态管理和工具调用循环 | 你要嵌入或扩展 Agent 执行过程 |
| `@earendil-works/pi-coding-agent` | 交互式 Coding Agent CLI | 你希望直接在终端里读代码、改代码和跑任务 |
| `@earendil-works/pi-tui` | 终端 UI 和差分渲染 | 你要定制终端显示或交互组件 |

可以用一条关系来记：

```text
pi-ai              = 怎么调用不同模型
pi-agent-core      = 怎么推动 Agent 状态
pi-coding-agent    = 怎么把它做成可用的编码 CLI
pi-tui             = 怎么把交互显示在终端里
```

如果你的目标只是“在 Node.js 里发一个模型请求”，优先看 `pi-ai`。如果目标是“让模型完成一组带工具的任务”，才需要进一步理解 Agent Core 和 Coding Agent。

## 3. 一次 Pi 请求实际经过了哪些层

把模型请求画成一条链，最容易定位后续问题：

```text
用户指令
  |
  v
pi-coding-agent：读取命令行参数、会话和项目上下文
  |
  v
pi-agent-core：组织 Agent 循环、工具调用和状态
  |
  v
pi-ai：选择 provider、模型和 API 协议
  |
  v
上游 API 企业 API 网关：Key、分组、模型渠道和调用记录
  |
  v
具体模型：返回文本、思考内容或工具调用
  |
  +--> 工具执行结果写回 Pi
  |
  +--> 最终回答或下一轮模型请求
```




```text
Base URL: https://api.example.com/v1
Endpoint: POST /chat/completions
鉴权：Authorization: Bearer <你的 上游 API 令牌>
模型：从 上游 API 模型广场复制精确 ID
```


### 4.1 准备环境变量

PowerShell：

```powershell
$env:API_KEY = "你的上游 API令牌"
$env:MODEL_ID = "claude-sonnet-4-5-20250929"
```

Linux/macOS：

```bash
export API_KEY='你的上游 API令牌'
export MODEL_ID='claude-sonnet-4-5-20250929'
```

不要把上面的真实令牌写入 `.ps1`、`.env`、Markdown 示例或 Git 提交。更稳妥的做法是使用本机凭证管理工具，再由启动脚本在进程环境中注入。

### 4.2 用 curl 验证网关和模型

PowerShell 可以直接执行下面的请求：

```powershell
$body = @{
  model = $env:MODEL_ID
  messages = @(
    @{ role = "user"; content = "只用两句话说明 Pi Agent Harness 是什么。" }
  )
  temperature = 0.2
  stream = $false
} | ConvertTo-Json -Depth 5

Invoke-RestMethod `
  -Method Post `
  -Uri "https://api.example.com/v1/chat/completions" `
  -Headers @{ Authorization = "Bearer $env:API_KEY" } `
  -ContentType "application/json" `
  -Body $body
```

Linux/macOS 使用 curl：

```bash
curl https://api.example.com/v1/chat/completions \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d "$(node -e 'console.log(JSON.stringify({model: process.env.MODEL_ID, messages: [{role: "user", content: "只用两句话说明 Pi Agent Harness 是什么。"}], temperature: 0.2, stream: false}))')"
```

这一步只证明三件事：令牌有效、路径可达、模型 ID 能被当前分组识别。它还不能证明模型支持工具调用，也不能证明 Coding Agent 的所有功能都能工作。

## 5. 再把同一个模型交给 Pi

Pi 支持使用 provider/model 形式选择模型。自定义 Provider 的完整配置放在下一篇，这里只展示验证目标：

```bash
pi --model provider/claude-sonnet-4-5-20250929 \
  -p "读取当前项目的 README，只输出三条最重要的技术结论。"
```


```text
1. 先用 curl 验证 上游 API endpoint 和 Key。
2. 再检查 ~/.pi/agent/models.json 是否存在且 JSON 合法。
3. 再用 pi --list-models provider 确认模型是否出现。
4. 最后才检查 Agent 工具、Skill 和项目权限。
```

不要在 curl 失败时直接修改 Pi 的工具配置。模型还没收到请求，工具层就没有排查价值。

## 6. Pi 的边界：扩展点不等于内置能力

Pi 的设计哲学是保持核心最小，通过扩展资源组合工作流。因此，下面这些功能不能直接当成 Pi 的内置保证：

| 能力 | Pi 的官方定位 | 实际落地方式 |
| --- | --- | --- |
| MCP | 不是内置协议层 | CLI 工具、Extension 或外部连接器 |
| Subagents | 不是内置固定编排 | 启动多个 Pi 实例或安装扩展 |
| Plan Mode | 不是固定模式 | 写计划文件或自定义 Extension |
| 权限弹窗 | 不是默认安全边界 | 外部沙箱、容器或自定义确认流程 |
| 待办系统 | 不内置通用实现 | `TODO.md` 或项目级 Skill |

这不是缺陷清单，而是使用前的边界清单。它提醒我们：如果工作流有写文件、执行命令、联网或访问生产数据的权限，就必须在 Pi 之外建立安全边界。

## 7. 企业级接入应该提前考虑什么

个人实验通常只有一个 Key 和一个模型。企业工作流至少要提前分开：

```text
项目 / 环境 / 工作流
  -> 独立 上游 API 令牌或分组
  -> 精确模型白名单
  -> 预算和额度
  -> 调用日志与失败记录
  -> 人工确认或外部沙箱
```

建议先按工作流建立令牌，而不是把同一个万能 Key 写进每个 Agent：

```text
上游 API-PI-READONLY   只读审查和摘要
上游 API-PI-CODING     测试环境代码任务
上游 API-PI-NIGHTLY    夜间低成本批处理
```


## 8. 本篇小结


可以用一句话记住两者的关系：

```text
Pi 负责“任务怎么继续推进”，上游 API 负责“模型请求怎么被统一管理”。
```




## 资料来源

- Pi GitHub：<https://github.com/earendil-works/pi>
- Pi 文档：<https://pi.dev/docs/latest>

## 结论

本文给出了问题定位、配置或验证的可执行路径。实际结果仍取决于当前版本、权限和运行环境，提交前应按官方文档复核可变字段，并保留失败证据和回滚边界。
