---
title: "OpenCode 的安装与模型配置：优先级、自定义提供方与超时"
category: 人工智能
tags:
  - 编程智能体
  - 配置管理
  - 模型接入
description: "说明 OpenCode 装好之后模型从哪里来、配置文件按什么顺序生效、自定义提供方的九个字段各管什么、把接口地址指向兼容端点要注意什么，以及三处容易忽略的超时数值和 V2 的字段改名。适合刚接手这套工具、或接自建端点时排查不出声问题的工程师。"
brand_mode: example
---

# OpenCode 的安装与模型配置：优先级、自定义提供方与超时

一个编码 agent 装好之后，最常见的两类问题不是它不聪明，而是它不知道该用哪个模型，或者知道却连不上。前者多半是配置文件的优先级没弄清，后者多半是端点、密钥或超时。

OpenCode 的配置体系有几个特点：配置文件有五个层级、模型目录来自外部目录服务、自定义提供方要写清楚凭据/运行时包/端点/模型这四样、超时有三处默认数值。本文按"装 → 配 → 接自定义端点 → 验证"的顺序把这些讲清楚。

信息来自仓库 README 与 opencode.ai 的 v2 文档（配置、提供方两页），检索日期为 2026-11-02，对应版本 v2.2.5。我没有安装或运行过它，所有命令与字段来自文档原文。

## 一、装在哪

| 方式 | 命令 |
| --- | --- |
| 安装脚本 | `curl -fsSL https://opencode.ai/v2/install \| bash` |
| npm | `npm i -g @opencode/cli@latest` |
| Homebrew | `brew install anomalyco/tap/opencode-v2` |
| Arch | `sudo pacman -S opencode` |
| mise / nix | `mise use -g opencode` / `nix run github:anomalyco/opencode` |

安装脚本会把程序放到 `$HOME/.opencode/bin`。README 提醒：升级前先卸掉 0.1.x 之前的旧版本，否则可能命中老的可执行文件。

还有桌面端（README 标注 BETA），提供 macOS、Windows、Linux 的安装包。要不要用桌面端是另一个话题，本文只讨论命令行与配置。

## 二、配置从哪读，谁覆盖谁

配置文件是 `opencode.jsonc`（也接受 `.json`）。建议第一行加 schema，编辑器就能校验字段：

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-sonnet-5-5"
}
```

### 五个层级

文档给出的位置与优先级（从低到高）：

| 顺序 | 位置 |
| --- | --- |
| 1（最低） | `~/.config/opencode/opencode.json(c)`：全局，对所有项目生效 |
| 2 | `<repo>/opencode.json(c)` 与 `<repo>/.opencode/opencode.json(c)`：项目级。注意两者不是平级——文档说 `.opencode/` 里的配置恒胜同级直放的配置 |
| 3 | 子包内的 `opencode.json(c)`，例如从 `packages/web` 启动时 |
| 4 | 系统级：Linux `/etc/opencode`、macOS `/Library/Application Support/opencode`、Windows `%ProgramData%\opencode` |
| 5（最高） | 托管设置，例如 macOS 上的 managed plist |

两个维护要记住：

1. **系统级比项目级高。** 这和多数工具的直觉相反，很多人的习惯是"项目配置最大"。在 OpenCode 里，系统级配置会覆盖仓库里的；对应到 Windows 是 `%ProgramData%\opencode`，不是 `/etc`。
2. **托管设置最高，且用户无法覆盖。** 文档还补了一句要紧的：仅靠设置挡不住用户添加插件，要阻止这类行为得在托管配置里用 `policies`。

关于 `policies` 的求值顺序，文档说全局与项目的 policies 按相反顺序求值，系统 policies 高于二者。这句我理解为：普通设置是"越靠后越优先"，而策略是"越权威越优先"，两者方向不同。具体实现我没有核验。

## 三、模型从哪来

常规流程是两步：先 `/connect` 选提供方并输入凭据，再 `/models` 选模型。

提供方与模型的目录来自 models.dev。这意味着两件事：一是常见提供方不需要你手写模型清单；二是目录之外的提供方要自己声明。

### 订阅登录的一个细节

如果用 ChatGPT 的订阅登录而不是 API 密钥，文档有两条说明：

1. 用量计入套餐，因此这些模型在界面上显示无费用。
2. 默认会把上下文上限压到模型真实上限以下，GPT-6.1 Sol 及更新为 500,000，更旧的是 272,000。

想取消这个压缩，可以在 `plugins` 里用 `-` 前缀关掉内置插件：

```jsonc
{
  "plugins": ["-opencode.provider.openai-context-caps"]
}
```

文档说这个前缀只关掉指定的内置插件，其他插件照常启用。注意这是"关掉一个限制"，关掉之后用量消耗速度会变，属于要有意识做的决定。

## 四、接自定义提供方

当某个提供方的 API 不在目录里，就要自己声明。文档给了一个 OpenAI 兼容的例子：

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "providers": {
    "acme": {
      "name": "Acme",
      "env": ["ACME_API_KEY"],
      "package": "@opencode/ai/providers/openai-compatible",
      "settings": { "baseURL": "https://llm.acme.example/v1" },
      "models": { "qwen3-coder": { "name": "Qwen 3 Coder" } }
    }
  }
}
```

`providers` 对象以提供方 ID 为键，模型引用写成 `acme/qwen3-coder`。

### 各字段管什么

| 字段 | 作用 |
| --- | --- |
| name | 显示名 |
| env | 有序的环境变量名，用来提供凭据 |
| package | 运行时提供方包 |
| canonical | 继承哪个内置提供方的目录默认值 |
| compatibility | 该提供方所有模型共用的 API 怪癖 |
| settings | 传给运行时包的类型化控制与 JSON 选项 |
| headers | 加到请求上的字符串头 |
| body | 合并进请求体的 JSON 字段 |
| models | 新增或覆盖的模型，以模型 ID 为键 |

`env` 在文档里的定义是"有序的环境变量名"，但**文档没有说靠前就优先**。我用它时的做法是先放项目专用的变量名再放全局的——这是按"有序"做的推测，未核验。

### 把地址指向兼容端点

两种写法，用途不同：

1. **目录里已有的提供方，走代理或兼容端点。** 只覆盖 `settings.baseURL`，原有的 package、models 与连接方式都保留。
2. **目录里没有的提供方。** 用上面那种完整声明，并选 `@opencode/ai/providers/openai-compatible` 包。

这里就是这类工具与中转端点相遇的位置：像 4SAPI 这类 OpenAI 兼容中转，在写法一里就是 `settings.baseURL` 的一个取值，在写法二里就是自定义提供方的 `baseURL`。换端点时改的是这一个地址，模型引用与权限配置都不用动。

要提醒的是 `settings` 是包相关的：字段只有在所选 package 支持时才生效。至于填错包时是报错还是静默无效，文档没说；我倾向于后者（按"字段只有在所选包支持时才生效"推的，未实测）——这大概是接自定义端点最容易遇到的"配了没用"。

### 按模型或变体加头

如果只有部分流量需要额外头或请求体字段，可以下沉到模型或变体级别：

```jsonc
{
  "providers": {
    "openai": {
      "models": {
        "gpt-6.1-sol": {
          "headers": { "X-Model-Tier": "coding" },
          "variants": [
            { "id": "batch", "body": { "service_tier": "flex" } }
          ]
        }
      }
    }
  }
}
```

这种分层在网关按租户或按服务层分流时很实用：不必为一个变体复制整份提供方配置。

## 五、三个容易忽略的超时默认值

长任务卡住时，多数人先怀疑模型服务，其实是客户端在等。

| 默认值 | 数值 | 改法 |
| --- | --- | --- |
| 等响应头 | 5 分钟 | `settings.headerTimeout`（毫秒）或 `false` 关闭 |
| 流式数据块之间 | 15 分钟 | `settings.chunkTimeout` |
| 整个请求 | 无默认值 | `settings.timeout` |
| WebSocket 传输下数据帧之间 | 30 分钟 | `chunkTimeout` 在此模式下以帧为界 |

三点要记：这些都是提供方级别的设置，不是每个模型单独配；`false` 是关闭限制而不是设成零；整个请求的 `timeout` 默认不设，也就是不限制。

### 一次排错思路（按症状推，不是实测）

如果症状是"长任务跑到一半断流，且没有明确错误"，按顺序查：

1. 这次是不是走了很慢的提供方，超过了 15 分钟的数据块间隔？
2. 网关是不是有比客户端更短的超时？
3. 有没有设过 `settings.timeout`，把一个本该更长的请求截断了？

这比直接调大数值更有用，因为把三个超时都设成很大，只会让问题更难定位。

## 六、验证配置真的生效

| 检查 | 怎么做 | 预期 |
| --- | --- | --- |
| 文件被读到 | 在配置里改一个可见项，例如默认模型 | 重启后 `/models` 里默认选中变了 |
| 层级覆盖对 | 全局设一个值、项目里设另一个 | 项目里生效的应是项目值；若不然，检查是否有系统级或托管设置在覆盖 |
| 提供方接通 | `/connect` 后 `/models` | 目录里出现该提供方的模型 |
| 自定义提供方 | 用 `acme/qwen3-coder` 这种带前缀的引用 | 能选中并调用；报错则先查 `package` 与 `env` 里的变量名 |
| 端点替换 | 只改 `settings.baseURL` | 模型列表不变、请求打到新地址 |
| Schema 校验 | 编辑器加 `$schema` | 字段拼错时编辑器直接标红 |

最后一行是最省事的：让编辑器替你查字段名，比运行时报错快得多。

## 七、V2 的字段改名

从 V1 迁过来时最容易踩的是字段名。文档明确写了 v2 的对应关系：

| V1 | V2 |
| --- | --- |
| permission | permissions |
| bash | shell |
| task | subagent |

这里要收窄一下说法：文档只说"在 V2 里应改用这些名字"，没有说旧字段会被静默忽略；而提供方那一页明确写了，某些退役的提供方 ID 在 V2 里会被**直接拒绝**并给出替代名。也就是说旧写法可能是被忽略、也可能直接报错，两种都有。所以"迁移过来了但规则不生效"这类问题，第一步仍然是查字段名，但别假设它一定安静失效。

## 八、限制与结论

OpenCode 的配置体系可以概括成三句话：配置有五个层级、系统级高于项目级；模型目录来自外部服务，目录外的提供方要把凭据、运行时包、端点与模型写清楚（文档的示例含这四样，未规定必须写全九个字段）；接兼容端点只改地址，但要选对运行时包。

本文的限制：没有安装或运行，所有字段与默认值来自 v2 文档，我没有核对它们与 v2.2.5 的实现是否一致；没有核验 policies 的求值顺序在代码里的实现；没有核对 models.dev 的目录内容；订阅与套餐细节不在本文范围。

下一步建议：先用默认配置跑通一次对话，再把 `settings.baseURL` 换到你自己的端点，用第六节的表格确认"模型列表没变、请求确实换了地址"。这一步能同时验证密钥、端点与超时三件事。
