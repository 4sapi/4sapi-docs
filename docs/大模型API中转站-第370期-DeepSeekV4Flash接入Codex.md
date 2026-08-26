---
title: "如何把 DeepSeek-V4-Flash 接入 Codex"
tags:
  - Codex
  - DeepSeek API
  - 开发工具
description: "根据 DeepSeek 官方接入文档，说明如何用一键脚本或手动配置把 V4-Flash 接入 Codex CLI、桌面端和 VS Code 插件，并完成验证与回滚。"
---
# 如何把 DeepSeek-V4-Flash 接入 Codex

如果你已经在 Codex 里使用某个默认模型，现在想切换到 DeepSeek-V4-Flash，真正需要配置的不是一行 API 地址，而是模型目录、提供方、协议和认证方式。只修改 `model`，Codex 可能无法识别模型能力；只修改 `base_url`，也可能仍然按照错误的接口协议发送请求。

DeepSeek 官方当前的 Codex 接入文档明确说明：目前只有 `deepseek-v4-flash` 支持接入 Codex，Codex 通过 Responses API 与模型交互。Codex CLI、ChatGPT 桌面端和 VS Code 的 Codex 插件共用配置文件，配置一次后可以在这些形态中复用。

本文给出两种方式：先解释官方一键脚本做了什么，再给出手动编辑 `config.toml` 的方法。涉及 API Key 的地方都用占位符，不把敏感信息写入仓库。

## 一、接入前检查

开始前确认：

- Codex CLI 或 ChatGPT 桌面端至少启动过一次；
- 用户目录下已经存在 `.codex`；
- 你有 DeepSeek API Key；
- 当前 Codex 版本支持自定义模型提供方和 Responses API；
- 已备份现有 Codex 配置，或者确认官方脚本会进行备份。

Windows 用户常见路径是：

```text
C:\Users\<用户名>\.codex\config.toml
C:\Users\<用户名>\.codex\models.json
```

macOS 和 Linux 通常是：

```text
~/.codex/config.toml
~/.codex/models.json
```

实际路径以当前 Codex 的用户目录和配置说明为准。不要在项目仓库里新建一个同名文件就认为 Codex 会读取它。

## 二、方式一：使用官方配置脚本

DeepSeek 官方 Codex 文档提供了 macOS/Linux 和 Windows PowerShell 的配置脚本。脚本会创建或更新模型目录、备份现有配置、写入提供方字段，并在写入前校验配置格式。

官方文档中给出的命令是：

macOS/Linux：

```bash
bash <(curl -fsSL https://cdn.deepseek.com/api-docs/codex-deepseek-setup.sh)
```

Windows PowerShell：

```powershell
irm https://cdn.deepseek.com/api-docs/codex-deepseek-setup-en.ps1 | iex
```

这是远程下载并执行脚本的操作，虽然来源是官方文档，但仍应按组织安全规则处理。更稳妥的做法是先把脚本下载到本地，查看内容、确认变更路径和备份行为，再执行；不要把任何不明来源的脚本直接接到生产凭据环境中。

运行后按菜单选择模型，并在提示时输入 API Key。脚本完成后应检查：

```text
~/.codex/config.toml 是否有备份
~/.codex/models.json 是否存在且格式可解析
config.toml 中的模型提供方是否指向 deepseek
模型是否为 deepseek-v4-flash
wire_api 是否为 responses
```

如果脚本提示冲突字段被删除，不要直接忽略输出。先保存终端日志，确认被删除的是旧的模型提供方字段，而不是项目、MCP 或权限配置。

## 三、方式二：手动编辑 config.toml

手动方式适合希望逐项审查配置的用户。DeepSeek 官方文档给出的关键配置如下：

```toml
model = "deepseek-v4-flash"
model_provider = "deepseek"
preferred_auth_method = "apikey"
forced_login_method = "api"
model_reasoning_effort = "high"
model_catalog_json = "~/.codex/models.json"

[model_providers.deepseek]
name = "deepseek"
base_url = "https://api.deepseek.com/"
wire_api = "responses"
experimental_bearer_token = "<你的 DeepSeek API Key>"
```

字段作用如下：

| 字段 | 作用 |
| --- | --- |
| `model` | 默认使用的模型名 |
| `model_provider` | 对应提供方配置段的标识 |
| `preferred_auth_method` | 使用 API Key 认证 |
| `forced_login_method` | 让 Codex 使用 API 模式而不是 ChatGPT 账号登录 |
| `model_reasoning_effort` | 默认推理强度，值越高通常需要更多时间和 Token |
| `model_catalog_json` | 指向自定义模型元数据文件 |
| `base_url` | DeepSeek API 地址 |
| `wire_api` | 使用 Responses API 协议 |
| `experimental_bearer_token` | API Key |

其中 `models.json` 不是可有可无的备注文件。它向 Codex 声明模型的上下文窗口、推理强度和工具调用格式等元数据。手动配置时，应按照当前 [DeepSeek Codex 接入文档](https://api-docs.deepseek.com/zh-cn/quick_start/agent_integrations/codex) 中的完整内容创建，不要凭感觉删掉字段，也不要把旧模型的元数据复制过来。

如果使用 4SAPI 这类网关作为模型入口，需要把它当作一条独立的兼容性路径验证：除了确认网关提供的 `base_url`，还要确认它是否支持 Codex 所需的 Responses API、`deepseek-v4-flash` 模型映射、`wire_api = "responses"`、工具调用和流式事件。只替换地址而不核对这些条件，可能出现普通请求能返回、Codex 工具循环却失败的情况。

## 四、API Key 的文件权限和提交风险

手动配置会把 API Key 写入用户目录配置文件。至少做三件事：

1. 确认 `config.toml` 不在 Git 仓库中；
2. 确认文件权限只允许当前用户和必要的系统账户读取；
3. 不要把配置内容复制到公开工单、截图、日志或聊天记录。

可以在项目根目录搜索常见敏感字段：

```bash
git grep -n -I -E "(api[_-]?key|token|password|secret|BEGIN .* PRIVATE KEY)" -- .
```

这个命令只能做粗检查，不能证明没有泄露。如果 Key 曾经进入 Git 历史、CI 日志或远程会话，应先撤销并轮换 Key，再处理历史清理。

## 五、配置完成后的验证顺序

### 1. 验证文件存在和格式

先确认两个文件存在，并用当前 Codex 支持的配置检查方式验证 TOML 和 JSON。不要用文本搜索代替解析检查，因为转义、引号和嵌套字段都可能导致配置无法加载。

### 2. 验证 Codex CLI 读取了模型

进入一个不包含敏感数据的测试项目：

```bash
cd /path/to/my-project
codex
```

官方文档给出的可观察信号是启动信息显示 `model: deepseek-v4-flash`。如果启动后仍显示原模型，优先检查：

- 是否修改了错误用户的 `.codex`；
- `model_catalog_json` 路径是否能解析；
- `model_provider` 是否和配置段名称一致；
- 配置文件是否有语法错误；
- 是否存在更高优先级的命令行或项目配置覆盖。

### 3. 验证一次小任务

不要一开始让 Codex 修改整个项目。先执行只读任务：

```text
只阅读当前项目，说明入口文件、测试命令和最近一次提交影响的模块。
不要修改文件，不要运行删除、发布或外部发送操作。
```

检查回答是否能正确识别项目路径、是否能执行必要的只读工具调用、是否有异常报错。验证通过后，再选择一个可以回滚的小修改。

### 4. 验证另外两个客户端

官方文档说明 ChatGPT 桌面端和 VS Code 的 Codex 插件与 CLI 共用配置。桌面端模型选择器可能统一显示为“自定义”，不能仅根据这个显示名称判断具体模型；应结合启动信息、请求日志或一个安全测试任务确认实际路由。

VS Code 插件安装后，重点检查它是否读取了同一个用户目录和配置文件。若 CLI 生效而插件不生效，先确认插件运行账户和 CLI 是否属于同一个用户，再检查插件版本的配置加载规则。

## 六、推理强度怎么选

官方配置示例使用 `model_reasoning_effort = "high"`。这个字段控制 Codex 请求的推理强度，通常强度越高，模型需要的时间和输出 Token 越多；它不是一个“越高越正确”的保证。

可以按任务类型做初始设置：

```text
快速项目问答、文件定位：low 或当前版本支持的较低档位
常规代码修改和测试分析：high
复杂重构、跨文件推理：根据耗时和成本再评估更高档位
```

实际可用档位由 Codex 模型目录和当前 DeepSeek 接入配置决定。若启动时报推理强度不支持，不要直接把字符串改成任意值；应查看当前 `models.json` 和 Codex 帮助中列出的合法档位。

## 七、如何恢复到原来的模型

回滚有两种方式：使用官方脚本提供的恢复选项，或者恢复配置备份。无论使用哪种方式，回滚后都要重新启动 Codex 客户端，因为已经启动的会话可能继续持有旧配置。

回滚验证至少包括：

```text
启动信息显示原模型
DeepSeek 提供方不再被默认选中
原有 MCP、信任级别和项目配置仍然存在
测试项目可以完成一次只读任务
```

不要为了回滚直接删除整个 `.codex` 目录。这样可能同时丢失 MCP 配置、信任设置、其他模型目录和历史备份。

## 八、几个容易踩的坑

### 只改 base_url

Codex 还可能按旧协议发送请求。DeepSeek 官方接入配置要求 `wire_api = "responses"`，不能把它省略。

### 只改模型名

如果模型目录没有声明 V4-Flash 的元数据，Codex 可能无法正确显示能力或选择有效的推理档位。

### 把 API Key 写进项目配置

项目配置会随仓库、压缩包或日志传播。优先使用用户级配置，并让项目只保存不含秘密的模型选择逻辑。

### 直接把所有真实代码交给新模型

第一次切换应先做只读任务和可回滚的小修改，确认工具、路径、分支和测试行为，再扩大权限。

## 结语

DeepSeek-V4-Flash 接入 Codex 的核心配置是四部分：模型名 `deepseek-v4-flash`、Responses API 的 `wire_api`、模型目录 `models.json` 和 API Key 认证。官方脚本可以降低配置工作量，但手动方式更容易审查具体变更；两种方式都需要在实际客户端中验证。

完成配置后，不要只看模型选择器。用启动信息、只读任务、工具调用和可回滚的小修改确认路由真实生效，并把 API Key 当作敏感配置管理。这样切换模型才是一次可验证的环境变更，而不是把配置文件改完就结束。
