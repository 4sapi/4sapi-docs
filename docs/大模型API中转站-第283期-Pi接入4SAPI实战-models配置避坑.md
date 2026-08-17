---
title: "Pi 的 models.json 如何配置自定义模型"
tags:
  - Pi
  - 配置管理
  - 开发工具
description: "Pi 能否调用自定义模型，首先取决于 models.json 的位置、字段和协议是否一致。"
---
# Pi 的 models.json 如何配置自定义模型
Pi 能否调用自定义模型，首先取决于 models.json 的位置、字段和协议是否一致。本文从最小配置开始，逐项核对 Provider、Base URL、模型 ID 和环境变量覆盖，并用一次可重复请求确认配置真正生效。文中只讨论可在本地复现的步骤，不把单次结果扩展成产品承诺；每个结论都标注前提、证据和无法覆盖的边界。读者可以先完成最小验证，再根据自己的版本、权限和数据补充实验，避免把配置示例误当成普遍结论。

## 1. 先画清楚配置关系

Pi 的自定义 Provider 可以理解为四个字段的组合：

```text
Provider ID
  + Base URL
  + API protocol
  + API key resolution
  + model catalog entries
```


```text
Pi Provider ID: provider
Base URL:       https://api.example.com/v1
API:            openai-completions
Key:            API_KEY
Model ID:       从 上游 API 模型广场复制的精确字符串
```

请求最后会落到：

```text
POST https://api.example.com/v1/chat/completions
Authorization: Bearer <API_KEY>
```


## 2. 找到正确的 models.json

Pi 官方文档把自定义模型放在用户目录下的 `models.json`。

Windows：

```text
%USERPROFILE%\\.pi\\agent\\models.json
```

如果当前用户是 `admin`，完整路径通常类似：

```text
C:\\Users\\admin\\.pi\\agent\\models.json
```

PowerShell 可以先创建目录并检查文件：

```powershell
$piAgentDir = Join-Path $env:USERPROFILE ".pi\\agent"
New-Item -ItemType Directory -Force -Path $piAgentDir | Out-Null
Test-Path (Join-Path $piAgentDir "models.json")
```

Linux/macOS：

```bash
mkdir -p ~/.pi/agent
test -f ~/.pi/agent/models.json && echo "models.json exists"
```

如果你使用 `PI_CODING_AGENT_DIR` 修改了 Pi 配置目录，应以这个环境变量解析出的目录为准。不要在项目仓库里随意新建一个同名文件，然后期待 Pi 自动读取它；项目级 `.pi` 目录适合放项目资源，但用户级 Provider 配置仍应按照当前 Pi 版本文档处理。



```text
注册或登录
  -> 充值或配置账户额度
  -> 创建 API 令牌
  -> 选择令牌分组和额度
  -> 从模型广场复制模型名称
  -> 根据接口文档调用
```

令牌分组要和工作流用途对应。例如：

```text
上游 API-PI-READONLY   只读审查、摘要和低风险查询
上游 API-PI-CODING     测试项目的编码任务
上游 API-PI-NIGHTLY    夜间批处理和日报生成
```

使用环境变量：

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


## 4. 写入一个最小 Provider

打开 `models.json`，加入下面的 Provider：

```json
{
  "providers": {
    "provider": {
      "name": "上游 API OpenAI Compatible",
      "baseUrl": "https://api.example.com/v1",
      "api": "openai-completions",
      "apiKey": "$API_KEY",
      "authHeader": true,
      "models": [
        {
          "id": "claude-sonnet-4-5-20250929",
          "name": "Claude Sonnet via 上游 API"
        }
      ]
    }
  }
}
```

这个配置有几个关键点：

| 字段 | 作用 | 注意事项 |
| --- | --- | --- |
| `baseUrl` | API 根地址 | 这里使用带 `/v1` 的 OpenAI 兼容入口 |
| `api` | 请求协议 | `openai-completions` 对应 Chat Completions |
| `authHeader` | 使用 Bearer 鉴权 | 让配置意图明确，具体行为以 Pi 当前版本为准 |


### 4.1 是否要补全模型元数据

Pi 的模型配置还支持 `reasoning`、`input`、`contextWindow`、`maxTokens`、`cost` 和 `compat` 等字段。最小配置可以先只放 `id` 和 `name`，确认请求能成功后，再根据真实模型能力补全。

例如，确认模型页面和接口都支持文本、思考控制和指定上下文后，再使用类似配置：

```json
{
  "id": "你的精确模型ID",
  "name": "上游 API 模型显示名称",
  "reasoning": true,
  "input": ["text"],
  "contextWindow": 128000,
  "maxTokens": 16384,
  "compat": {
    "supportsDeveloperRole": false,
    "supportsReasoningEffort": false
  }
}
```

上面的数值和兼容开关不是通用事实。`contextWindow` 和 `maxTokens` 填错会造成截断或错误；`supportsDeveloperRole` 和 `supportsReasoningEffort` 只有在上游明确不支持时才应该关闭。不要为了“先跑起来”盲目复制这些字段。

## 5. 验证 Provider 是否被 Pi 识别

首先检查 JSON 是否能被 PowerShell 解析：

```powershell
$modelsFile = Join-Path $env:USERPROFILE ".pi\\agent\\models.json"
Get-Content -Raw -Encoding UTF8 $modelsFile | ConvertFrom-Json | Out-Null
Write-Output "models.json is valid JSON"
```

然后让 Pi 列出自定义 Provider 的模型：

```bash
pi --list-models provider
```


```bash
pi --model provider/claude-sonnet-4-5-20250929 \
  -p "只输出 OK，并说明你收到的任务是文本生成。"
```

Windows PowerShell：

```powershell
pi --model "provider/$env:MODEL_ID" `
  -p "只输出 OK，并说明你收到的任务是文本生成。"
```


## 6. 配置失败时按层排查


按顺序检查：

```text
1. 文件路径是否真的是 ~/.pi/agent/models.json。
2. JSON 是否有多余逗号或全角引号。
3. Provider 是否位于根级 providers 对象下。
4. models 数组是否至少有一个 id。
5. 当前 Pi 版本是否支持该配置格式。
6. Key 是否已经通过环境变量、auth.json 或 CLI 参数配置。
```

Pi 的自定义 Provider 可能在没有凭证时加载配置，但模型不会出现在可选择列表中。这个现象容易被误认为模型 ID 错误，实际可能只是 Key 没有被当前进程看到。

### 6.2 返回 401 或 403

这是鉴权和权限层问题，优先检查：

```powershell
if ([string]::IsNullOrWhiteSpace($env:API_KEY)) {
  throw "API_KEY is not set in this PowerShell process"
}
```

然后确认：

- 令牌是否复制完整；
- 当前令牌是否启用；
- 令牌分组是否包含这个模型；
- 额度、期限和权限是否允许当前调用；
- Pi 进程是否从另一个终端启动，导致没有继承环境变量。

不要因为 401 就更换模型。模型路由问题和令牌权限问题要分开处理。

### 6.3 返回 400

400 常见于协议字段不匹配：

| 现象 | 可能原因 | 处理方向 |
| --- | --- | --- |
| model not found | 模型 ID 拼写错误或分组不可用 | 从模型广场复制精确 ID |
| developer role 不支持 | 上游只接受 `system` | 在模型级设置 `supportsDeveloperRole: false` |
| reasoning_effort 不支持 | 上游不识别该字段 | 设置 `supportsReasoningEffort: false` 或关闭思考选项 |
| tool 参数不合法 | 当前模型或渠道不支持工具调用 | 先做无工具文本请求，再查模型能力 |

兼容配置应当根据错误响应增加，不能把所有兼容开关都关掉。否则请求虽然可能通过，Pi 也可能丢失模型原本支持的能力。

### 6.4 返回 429、500、503、504 或 524

这些错误先不要通过无限重试解决：

```text
429 -> 检查 上游 API 令牌额度、分组并发和请求频率
500/503 -> 检查上游渠道状态和备用模型
504/524 -> 检查请求体大小、流式连接和客户端超时
```

对于长任务，应该在 Agent 层设置最大轮数和停止条件，在网关层配置额度和告警。反复重试会把一次服务异常放大成成本异常。

## 7. 一个更适合团队的配置拆分

个人实验可以只有一个 Provider。团队协作时，可以为不同工作流创建不同 Provider ID：

```json
{
  "providers": {
    "provider-readonly": {
      "name": "上游 API Readonly",
      "baseUrl": "https://api.example.com/v1",
      "api": "openai-completions",
      "apiKey": "$FOURSAPI_READONLY_KEY",
      "models": [{ "id": "你的低成本模型ID" }]
    },
    "provider-coding": {
      "name": "上游 API Coding",
      "baseUrl": "https://api.example.com/v1",
      "api": "openai-completions",
      "apiKey": "$FOURSAPI_CODING_KEY",
      "models": [{ "id": "你的编程模型ID" }]
    }
  }
}
```

这样做的好处是：

```text
只读任务 -> provider-readonly/模型ID
编码任务 -> provider-coding/模型ID
```


## 8. 本篇小结


```text
配置文件路径
API 协议
令牌分组
精确模型 ID
```

先用无工具的 Chat Completions 请求验证链路，再增加思考、图片、工具和 Coding Agent 能力。这样每一次错误都能知道发生在配置、鉴权、模型能力还是 Agent 工具层。




## 资料来源

- Pi Providers：<https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/providers.md>
- Pi Custom Models：<https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/models.md>

## 结论

本文给出了问题定位、配置或验证的可执行路径。实际结果仍取决于当前版本、权限和运行环境，提交前应按官方文档复核可变字段，并保留失败证据和回滚边界。
