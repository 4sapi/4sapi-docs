---
title: "【大模型API中转站】第122期 OmO配置4SAPI | OpenCode实战"
category: 人工智能
tags:
  - 大模型API中转站
  - oh-my-openagent
  - OpenCode
  - OpenAI-compatible
  - Agent模型路由
  - 企业级API
  - 4SAPI
description: "手把手讲 OpenCode + oh-my-openagent Ultimate 如何通过 OpenAI-compatible Provider 接入 4SAPI：准备 Key、配置 baseURL、复制模型名、验证 opencode 模型列表，再用 OmO 的 agents、categories、fallback_models 和 runtime_fallback 做多 Agent 模型路由。"
---

# 【大模型API中转站】第122期 OmO配置4SAPI | OpenCode实战

本文是【大模型API中转站】系列的第122篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一篇讲了结论：

```text
OmO 可以写企业级接入 4SAPI。
但准确架构是 OpenCode + oh-my-openagent + 4SAPI。
```

这一篇写实操。

目标很明确：

```text
让 OpenCode 通过 4SAPI 调模型。
让 oh-my-openagent 的不同 Agent 使用这些模型。
再用 fallback 和日志做企业级治理。
```

先说边界。

这篇不是教你绕过任何官方限制。

只讲合法合规的 API 接入、企业级 API 网关、模型路由、权限审计和成本治理。

## 1. 接入前先分清三份配置

很多人配不通，是因为把三份配置混在一起。

第一份是 OpenCode 配置。

它负责：

```text
Provider
Model
API Key
Base URL
```

第二份是 oh-my-openagent 配置。

它负责：

```text
哪个 Agent 用哪个模型。
哪个 Category 用哪个模型。
fallback_models 怎么排。
Team Mode 是否开启。
runtime_fallback 是否启用。
```

第三份是 4SAPI 后台配置。

它负责：

```text
Key
额度
有效期
分组
模型权限
日志
成本统计
```

这三层不要混。

OpenCode 是入口。

OmO 是编排。

4SAPI 是网关。

## 2. 准备 4SAPI 三件套

先在 4SAPI 后台准备三样东西：

```text
API Key
Base URL
模型名
```

常见 Base URL：

```text
https://4sapi.com/v1
```

如果你用 curl 直测 Chat Completions，常见完整接口是：

```text
https://4sapi.com/v1/chat/completions
```

但在 OpenCode Provider 里，通常填的是 Base URL，而不是完整接口。

所以这里先记：

```text
OpenCode Provider baseURL: https://4sapi.com/v1
```

模型名从 4SAPI 模型广场复制。

不要手写。

不要只写：

```text
gpt
claude
sonnet
kimi
glm
```

要复制完整 model id。

企业使用建议先拆三把 Key：

```text
4sapi-omo-dev
用途：普通开发任务。
```

```text
4sapi-omo-review
用途：代码审查、架构判断。
```

```text
4sapi-omo-team
用途：Team Mode、多 Agent 并行任务。
```

不要所有人共用一把 Key。

## 3. 先用 curl 验证 4SAPI

不要上来就改 OpenCode。

先确认 Key、URL、模型名能跑。

```bash
curl https://4sapi.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-xxxxxxxxxxxxxxxx" \
  -d '{
    "model": "your-4sapi-model-name",
    "messages": [
      { "role": "user", "content": "只回复 ok" }
    ]
  }'
```

Windows PowerShell 可以写：

```powershell
curl.exe https://4sapi.com/v1/chat/completions `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer sk-xxxxxxxxxxxxxxxx" `
  -d "{ `"model`": `"your-4sapi-model-name`", `"messages`": [{ `"role`": `"user`", `"content`": `"只回复 ok`" }] }"
```

能返回，再继续。

如果这里不通，先不要怀疑 OmO。

优先查：

```text
Key 是否正确。
模型名是否正确。
Key 是否有模型权限。
余额是否足够。
URL 是否重复拼了 /v1。
```

## 4. 配置环境变量

不要把真实 Key 写进公开仓库。

macOS / Linux：

```bash
export FOURSAPI_API_KEY="sk-xxxxxxxxxxxxxxxx"
```

Windows PowerShell 临时设置：

```powershell
$env:FOURSAPI_API_KEY="sk-xxxxxxxxxxxxxxxx"
```

Windows 长期保存：

```powershell
[Environment]::SetEnvironmentVariable("FOURSAPI_API_KEY", "sk-xxxxxxxxxxxxxxxx", "User")
```

然后重新打开终端。

检查：

```powershell
echo $env:FOURSAPI_API_KEY
```

环境变量名可以叫别的。

但建议统一：

```text
FOURSAPI_API_KEY
```

后面排错更省事。

## 5. 配置 OpenCode Provider

OpenCode 支持自定义 Provider。

OpenAI-compatible 场景可以使用：

```text
@ai-sdk/openai-compatible
```

官方文档里有两种放 Key 的方式。

第一种是在 OpenCode TUI 里运行：

```text
/connect
```

然后选择 Other，输入自定义 provider id 和 API Key。

第二种是在 Provider 配置里使用：

```text
options.apiKey
```

并通过环境变量读取。

这篇为了方便团队复制和审计，使用第二种写法。

配置思路大致是：

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "4sapi": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "4SAPI",
      "options": {
        "baseURL": "https://4sapi.com/v1",
        "apiKey": "{env:FOURSAPI_API_KEY}"
      },
      "models": {
        "your-4sapi-model-name": {
          "name": "4SAPI Coding Model"
        }
      }
    }
  }
}
```

注意两点。

第一，字段名以你当前 OpenCode 版本文档为准。

不同版本可能对 `provider`、`options`、`models` 的细节有差异。

第二，`your-4sapi-model-name` 必须替换成 4SAPI 模型广场复制的真实模型名。

第三，这篇按 OpenAI-compatible 的 `/v1/chat/completions` 思路写。

如果你接的某个模型通道要求 `/v1/responses`，要按 OpenCode Provider 文档改用对应 SDK 包，并先用最小请求验证。

如果你有多个模型，可以这样写：

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "4sapi": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "4SAPI",
      "options": {
        "baseURL": "https://4sapi.com/v1",
        "apiKey": "{env:FOURSAPI_API_KEY}"
      },
      "models": {
        "your-fast-model": {
          "name": "4SAPI Fast"
        },
        "your-coding-model": {
          "name": "4SAPI Coding"
        },
        "your-reasoning-model": {
          "name": "4SAPI Reasoning"
        }
      }
    }
  }
}
```

这样 OpenCode 里就能看到：

```text
4sapi/your-fast-model
4sapi/your-coding-model
4sapi/your-reasoning-model
```

具体显示格式以 OpenCode 实际输出为准。

## 6. 验证 OpenCode 能看到模型

配置后先跑：

```bash
opencode models
```

看 4SAPI 模型是否出现。

再启动一个最小任务：

```bash
opencode
```

输入：

```text
只回复 ok，不要读取文件。
```

如果能正常回复，说明 OpenCode Provider 层通了。

如果 OpenCode 找不到模型，优先查：

```text
配置文件路径是否正确。
JSON / JSONC 是否有语法错误。
环境变量是否在当前终端生效。
模型名是否与 provider models 匹配。
OpenCode 是否需要重启。
```

这里还没到 OmO。

先把底层入口跑通。

## 7. 安装 oh-my-openagent Ultimate

OmO Ultimate 面向 OpenCode。

推荐按官方安装指南走：

```bash
bunx oh-my-openagent install
```

如果想非交互安装，可以按文档传：

```bash
bunx oh-my-openagent install --no-tui --platform=opencode
```

安装完成后，确认 OpenCode 能加载插件。

可以跑：

```bash
bunx oh-my-openagent doctor
```

如果你只装 Codex Light：

```bash
npx lazycodex-ai install
```

那是另一条线。

这篇企业级多 Agent 接入，主讲 Ultimate。

## 8. 配置 OmO 的 Agent 模型

OmO 的配置文件常见位置：

```text
~/.config/opencode/oh-my-openagent.jsonc
```

项目级可以放：

```text
.opencode/oh-my-openagent.jsonc
```

旧文件名：

```text
oh-my-opencode.jsonc
```

在过渡期也会被识别。

现在我们把 Agent 指到 4SAPI Provider 下的模型。

示例：

```jsonc
{
  "$schema": "https://raw.githubusercontent.com/code-yeongyu/oh-my-openagent/dev/assets/oh-my-opencode.schema.json",
  "agents": {
    "sisyphus": {
      "model": "4sapi/your-reasoning-model",
      "fallback_models": [
        "4sapi/your-coding-model",
        "4sapi/your-fast-model"
      ]
    },
    "hephaestus": {
      "model": "4sapi/your-coding-model",
      "fallback_models": [
        "4sapi/your-reasoning-model"
      ]
    },
    "oracle": {
      "model": "4sapi/your-reasoning-model"
    },
    "librarian": {
      "model": "4sapi/your-fast-model"
    },
    "explore": {
      "model": "4sapi/your-fast-model"
    }
  }
}
```

这里的模型名要和 OpenCode 里显示一致。

如果 OpenCode 模型格式是：

```text
4sapi/model-id
```

OmO 里也按这个写。

不要自己发明格式。

## 9. 配置 Category 模型

Agent 是角色。

Category 是任务类型。

企业里最适合用 Category 做成本分层。

示例：

```jsonc
{
  "categories": {
    "quick": {
      "model": "4sapi/your-fast-model",
      "description": "小改动、拼写修复、单文件任务"
    },
    "deep": {
      "model": "4sapi/your-coding-model",
      "fallback_models": [
        "4sapi/your-reasoning-model"
      ],
      "description": "跨文件代码修改、复杂执行"
    },
    "ultrabrain": {
      "model": "4sapi/your-reasoning-model",
      "description": "架构判断、复杂逻辑、疑难问题"
    },
    "writing": {
      "model": "4sapi/your-writing-model",
      "description": "中文文档、博客、说明书和总结"
    }
  }
}
```

这比所有 Agent 都用一个模型更稳。

你可以把任务成本控制在三层：

```text
fast：便宜快，负责搜索和小改。
coding：中高成本，负责代码修改。
reasoning：高成本，只给架构、Review、难题。
```

## 10. 配置 runtime_fallback

4SAPI 作为网关，有时会遇到：

```text
限流
余额不足
模型暂不可用
上游超时
Key 权限不足
```

OmO 支持 runtime fallback。

可以先用保守配置：

```jsonc
{
  "runtime_fallback": {
    "enabled": true,
    "retry_on_errors": [429, 500, 502, 503, 504],
    "max_fallback_attempts": 3,
    "cooldown_seconds": 60,
    "timeout_seconds": 30,
    "notify_on_fallback": true
  }
}
```

如果你的网关把配额、模型不可用、权限错误返回成其他状态码，也可以按实际情况加入：

```jsonc
{
  "runtime_fallback": {
    "enabled": true,
    "retry_on_errors": [400, 401, 403, 404, 429, 500, 502, 503, 504],
    "max_fallback_attempts": 3,
    "cooldown_seconds": 15,
    "timeout_seconds": 10,
    "notify_on_fallback": true
  }
}
```

但不要乱加。

企业里建议先观察 4SAPI 日志里的真实错误码。

再决定哪些错误应该 fallback。

比如：

```text
429：可以 fallback。
500 / 502 / 503 / 504：可以 fallback。
401：通常是 Key 问题，不一定应该 fallback。
403：可能是权限问题，不一定应该 fallback。
404：可能是模型名错误，不一定应该 fallback。
```

错误码策略要结合实际网关语义。

## 11. 开启 Team Mode 前的配置建议

Team Mode 很强。

也很会消耗模型。

建议先默认关闭。

需要大任务时再开启：

```jsonc
{
  "team_mode": {
    "enabled": true,
    "max_parallel_members": 4,
    "max_members": 8,
    "tmux_visualization": false,
    "max_wall_clock_minutes": 120
  }
}
```

企业建议给 Team Mode 单独 Key。

比如：

```text
FOURSAPI_API_KEY=sk-team-mode-key
```

或者在 4SAPI 后台给 Team Mode 所用模型单独分组。

原因很简单：

```text
并行 Agent 的调用量和普通单 Agent 不是一个级别。
```

大任务可以开。

日常小任务不要开。

## 12. 最小可用配置样例

下面给一份完整思路。

第一步，OpenCode 配 4SAPI Provider：

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "4sapi": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "4SAPI",
      "options": {
        "baseURL": "https://4sapi.com/v1",
        "apiKey": "{env:FOURSAPI_API_KEY}"
      },
      "models": {
        "your-fast-model": { "name": "4SAPI Fast" },
        "your-coding-model": { "name": "4SAPI Coding" },
        "your-reasoning-model": { "name": "4SAPI Reasoning" },
        "your-writing-model": { "name": "4SAPI Writing" }
      }
    }
  }
}
```

第二步，OmO 配 Agent 和 Category：

```jsonc
{
  "$schema": "https://raw.githubusercontent.com/code-yeongyu/oh-my-openagent/dev/assets/oh-my-opencode.schema.json",
  "agents": {
    "sisyphus": {
      "model": "4sapi/your-reasoning-model",
      "fallback_models": ["4sapi/your-coding-model"]
    },
    "hephaestus": {
      "model": "4sapi/your-coding-model",
      "fallback_models": ["4sapi/your-reasoning-model"]
    },
    "oracle": {
      "model": "4sapi/your-reasoning-model"
    },
    "librarian": {
      "model": "4sapi/your-fast-model"
    },
    "explore": {
      "model": "4sapi/your-fast-model"
    }
  },
  "categories": {
    "quick": {
      "model": "4sapi/your-fast-model"
    },
    "deep": {
      "model": "4sapi/your-coding-model",
      "fallback_models": ["4sapi/your-reasoning-model"]
    },
    "ultrabrain": {
      "model": "4sapi/your-reasoning-model"
    },
    "writing": {
      "model": "4sapi/your-writing-model"
    }
  },
  "runtime_fallback": {
    "enabled": true,
    "retry_on_errors": [429, 500, 502, 503, 504],
    "max_fallback_attempts": 3,
    "cooldown_seconds": 60,
    "timeout_seconds": 30,
    "notify_on_fallback": true
  },
  "team_mode": {
    "enabled": false
  }
}
```

第三步，测试一个简单任务：

```text
请只读扫描当前项目，列出目录结构和你建议优先阅读的文件。不要修改任何文件。
```

如果能跑，再测试一个小修改。

不要一上来让它做大型重构。

## 13. 企业项目里的 AGENTS.md

模型接入只是第一步。

企业项目必须有规则。

建议在项目根目录放：

```text
AGENTS.md
```

内容可以写：

```markdown
# 项目协作规则

- 默认使用简体中文回复。
- 修改前先给计划。
- 只修改与任务直接相关的文件。
- 不要删除原始数据和历史文档。
- 代码修改后优先运行现有测试。
- 如果测试无法运行，说明原因。
- 重要决策要说明取舍。
- 任务结束输出：修改文件、验证方式、风险点、下一步建议。
```

OmO 的 rules injection 能把这类规则带进上下文。

这比每次手动提醒靠谱。

不要把 4SAPI Key 写进 AGENTS.md。

## 14. 验证顺序

建议按这个顺序排查。

```text
1. curl 直连 4SAPI。
2. opencode models 能看到模型。
3. opencode 最小对话能回复。
4. oh-my-openagent doctor 通过。
5. OmO Agent 能执行只读任务。
6. 4SAPI 后台能看到日志。
7. fallback 能在模拟错误时触发。
8. 小范围修改任务能完成。
9. Git diff 可读。
10. 成本记录正常。
```

不要跳步骤。

否则出错时很难知道是哪一层的问题。

## 15. 常见问题

### 15.1 OmO 启动后没用 4SAPI 模型

检查：

```text
OmO 配置文件路径是否正确。
model 写法是否和 OpenCode models 输出一致。
是否有 UI-selected model 覆盖了配置。
是否需要重启 OpenCode。
```

### 15.2 Provider 能通，但 Agent 报模型不可用

检查：

```text
OpenCode Provider 里 models 是否声明了该模型。
OmO 里 model 是否拼错。
4SAPI Key 是否有该模型权限。
模型名是否复制完整。
```

### 15.3 fallback 没触发

检查：

```text
runtime_fallback.enabled 是否为 true。
错误码是否在 retry_on_errors 里。
fallback_models 是否配置在对应 agent/category 下。
是否达到 max_fallback_attempts。
```

### 15.4 成本突然升高

优先看：

```text
Team Mode 是否开启。
任务是否进入循环。
是否所有 Agent 都用高成本模型。
是否长上下文反复读取。
是否 fallback 连续触发。
```

处理方式：

```text
降低 max_parallel_members。
给 quick / explore / librarian 换低成本模型。
缩小任务范围。
加停止条件。
在 4SAPI 后台设置额度。
```

## 16. 最后总结

OmO 接入 4SAPI 的关键，不是复制一段配置。

而是分清三层：

```text
OpenCode：把 4SAPI 接进来。
OmO：把模型分给不同 Agent 和任务。
4SAPI：管 Key、日志、额度和成本。
```

最小步骤是：

```text
1. 4SAPI 准备 Key、Base URL、模型名。
2. curl 先测通。
3. OpenCode 配 openai-compatible Provider。
4. opencode models 验证。
5. OmO agents/categories 指向 4SAPI 模型。
6. 开启 runtime_fallback。
7. 用 4SAPI 后台看日志和成本。
```

一句话：

```text
不要把所有 Agent 都塞到一个模型里。
用 4SAPI 做模型网关，用 OmO 做任务分工，企业级接入才有意义。
```

下一篇继续讲企业上线：

```text
Key 怎么拆、额度怎么设、Team Mode 怎么管、日志怎么审、成本怎么复盘。
```

## 资料来源与延伸阅读

- oh-my-openagent GitHub：https://github.com/code-yeongyu/oh-my-openagent
- oh-my-openagent 安装指南：https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/guide/installation.md
- oh-my-openagent 配置文档：https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/reference/configuration.md
- OpenCode Provider 文档：https://opencode.ai/docs/providers/
- OpenCode Config 文档：https://opencode.ai/docs/config/
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 接入实操手册：https://4sapi.com/blog/4sapi-api-integration-setup-guide
- 4SAPI Coding Agent 接入：https://4sapi.com/blog/4sapi-coding-agent-integration-guide
