---
title: "OpenAI 兼容接口如何拆分 cURL 与 JSON Body"
category: 人工智能
tags:
  - 大模型 API
  - HTTP 请求
  - JSON
description: "解释一条聊天补全 cURL 命令由请求方法、地址、请求头和 JSON Body 组成，并给出 Postman 与程序接入时可直接使用的拆分结果。"
---

# OpenAI 兼容接口如何拆分 cURL 与 JSON Body

拿到一段大模型 API 的 cURL 示例后，很多人第一步不是在终端执行，而是把它填进 Postman、Apifox、工作流节点或自己的程序。这时最常见的误区，是把整段 cURL 当成 Body，或者只复制 `-d` 后面的 JSON，却忘了请求方法和鉴权头。本文用两条聊天补全请求说明如何拆分，并给出检查方法。这里只讨论请求结构，不验证示例端点、账号权限和模型是否在当前日期可用。

## 1. 一条 cURL 请求包含四部分

下面这类命令可以拆成四层：

```bash
curl -X POST "https://api.example.com/v1/chat/completions" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "example-model",
    "messages": [
      {"role": "user", "content": "Hello"}
    ]
  }'
```

对应关系如下：

| cURL 片段 | 请求中的位置 | 作用 |
| --- | --- | --- |
| `-X POST` | Method | 指定 HTTP 方法 |
| URL | Request URL | 指定接口地址 |
| `-H` | Headers | 传递鉴权信息与内容类型 |
| `-d` | Body | 传递 JSON 请求数据 |

Body 只是整条请求的一部分。把 JSON 单独取出来以后，仍要在调用工具中配置 `POST`、URL 和两个请求头。

## 2. Together 示例如何拆分

原请求使用的地址是：

```text
https://api.together.ai/v1/chat/completions
```

请求方法：

```text
POST
```

请求头：

```http
Authorization: Bearer YOUR_TOGETHER_API_KEY
Content-Type: application/json
```

JSON Body：

```json
{
  "model": "moonshotai/Kimi-K3",
  "messages": [
    {
      "role": "user",
      "content": "What are some fun things to do in New York?"
    }
  ]
}
```

这里的 `YOUR_TOGETHER_API_KEY` 是占位符。真实 Key 应放在环境变量或调用平台的密钥配置中，不要写入文章、代码仓库或截图。`moonshotai/Kimi-K3` 是否存在于当前账号、是否支持所需能力，也应通过供应商当前模型列表核对。

## 3. 另一个兼容接口的拆分方式

第二条请求使用：

```text
POST https://nano-gpt.com/api/v1/chat/completions
```

请求头仍是：

```http
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json
```

它的 Body 是：

```json
{
  "model": "gpt-5.2",
  "messages": [
    {
      "role": "user",
      "content": "Hello, how are you?"
    }
  ]
}
```

“OpenAI 兼容”通常表示请求结构相近，不表示不同供应商拥有相同的模型目录、参数支持、错误码细节或计费规则。模型 ID 必须以目标接口当前返回的模型列表或官方文档为准。

## 4. 在 Postman 或 Apifox 中怎么填

以第一条请求为例：

1. Method 选择 `POST`；
2. URL 填入 `https://api.together.ai/v1/chat/completions`；
3. Headers 中添加 `Authorization` 和 `Content-Type`；
4. Body 选择 `raw` 或 JSON 类型；
5. 粘贴 JSON Body，不要把 `curl`、`-H` 或 `-d` 一起粘进去。

发送前先使用编辑器自带的 JSON 校验。Body 中的属性名和字符串必须使用双引号；cURL 外层用于包住整段 JSON 的单引号，不属于 JSON 内容。

## 5. Bash 与 PowerShell 的换行符不同

示例里的反斜杠 `\` 是 Bash 常见的续行写法。直接把它粘到 PowerShell，命令可能无法按预期解析。PowerShell 可以使用反引号续行，但更稳妥的做法是先把命令写成一行，或者使用 `Invoke-RestMethod` 并通过对象生成 JSON。

例如：

```powershell
$headers = @{
    Authorization = "Bearer $env:TOGETHER_API_KEY"
    "Content-Type" = "application/json"
}

$body = @{
    model = "moonshotai/Kimi-K3"
    messages = @(
        @{
            role = "user"
            content = "What are some fun things to do in New York?"
        }
    )
} | ConvertTo-Json -Depth 5

Invoke-RestMethod `
    -Method Post `
    -Uri "https://api.together.ai/v1/chat/completions" `
    -Headers $headers `
    -Body $body
```

运行前在同一 PowerShell 会话设置 `TOGETHER_API_KEY`。不要为了消除环境变量报错而在脚本中加入真实 Key 作为默认值。

## 6. 如何验证拆分是否正确

先检查客户端实际发送的内容：

```text
Method 是否为 POST
URL 是否包含完整的 /v1/chat/completions
Authorization 是否以 Bearer 加一个空格开头
Content-Type 是否为 application/json
Body 是否能通过 JSON 解析
```

再根据响应分类：

| 现象 | 优先检查 |
| --- | --- |
| `400 Bad Request` | JSON 字段、类型和模型参数 |
| `401 Unauthorized` | Key 是否缺失、失效或格式错误 |
| `403 Forbidden` | 账号或模型权限 |
| `404 Not Found` | 域名与接口路径 |
| `405 Method Not Allowed` | 请求是否真的以 `POST` 发出 |

状态码只能指明排查方向，最终仍应结合响应体、响应头和供应商日志判断。

## 7. 结论与限制

把 cURL 转成 Body，本质上是把 `-d` 后的 JSON 单独取出，同时保留原请求的 Method、URL 和 Headers。只复制 JSON 不会自动带上鉴权，也不会让调用工具自动选择 `POST`。

本文保留了两条原始请求中的端点和模型 ID，但没有使用真实账号执行请求。发布或投入程序前，应再次核对当前接口文档、模型列表和参数支持情况。

## 官方来源

- [Together AI：Chat Completions API](https://docs.together.ai/reference/chat-completions)
- [curl：Command line options](https://curl.se/docs/manpage.html)

---

*本文请求结构核对日期：2026-08-10。模型目录、接口域名和参数支持可能变化，请以供应商当前官方文档为准。*
