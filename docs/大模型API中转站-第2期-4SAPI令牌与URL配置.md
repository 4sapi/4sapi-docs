---
title: "【大模型API中转站】第2期 4SAPI令牌与URL配置 | 5分钟跑通"
category: 人工智能
tags:
  - 大模型API中转站
  - 4SAPI
  - API Key
  - Base URL
  - OpenAI兼容接口
description: "围绕4SAPI文档里的令牌、Base URL、模型名称和首次调用流程，整理一篇适合新手复制检查的接入教程。"
---

# 【大模型API中转站】第2期 4SAPI令牌与URL配置 | 5分钟跑通

本文是【大模型API中转站】系列的第2篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一篇我们讲了为什么需要4SAPI这类大模型API中转站。这一篇只解决一个最基础、也最容易卡住的问题：注册充值以后，Key在哪？URL填哪个？模型名称怎么写？如果这三项填错，后面所有SDK、客户端、工作流工具都会一起报错。

## 1. 先搞清楚三件事

接入4SAPI时，你实际需要配置的是三项：

- API Key：也就是4SAPI控制台里的令牌。
- Base URL：也就是接口地址，常见填写`https://4sapi.com`或`https://4sapi.com/v1`。
- Model：也就是模型名称，需要从模型广场或价格页完整复制。

很多人第一次失败，不是模型不能用，而是把这三项混在了一起。比如把完整请求地址填进Base URL，把`Bearer`一起填进客户端Key输入框，或者手打模型名导致少了版本号。

### 1.1 为什么只差一个`/v1`就会失败

OpenAI兼容接口通常有一个约定：SDK会把`base_url`和具体资源路径拼起来。比如你在SDK里填：

```text
base_url = https://4sapi.com/v1
```

SDK调用聊天接口时，最终请求可能会变成：

```text
https://4sapi.com/v1/chat/completions
```

但有些第三方软件不是这样做的。它可能要求你填主机地址，然后自己拼接`/v1/chat/completions`；也可能要求你直接填完整接口地址。于是同一个4SAPI，在不同工具里的URL写法会不同。

这就是为什么4SAPI文档建议在`https://4sapi.com`和`https://4sapi.com/v1`之间尝试。不是平台不稳定，而是客户端拼接规则不同。

### 1.2 三种配置形态对照

可以用这张表快速判断：

| 工具输入框 | 推荐填写 | 说明 |
| --- | --- | --- |
| Base URL / API Base | `https://4sapi.com/v1` | 常见于OpenAI SDK、很多OpenAI Compatible工具 |
| Host / Endpoint | `https://4sapi.com` | 常见于会自动拼`/v1`的软件 |
| Full URL / Request URL | `https://4sapi.com/v1/chat/completions` | 常见于HTTP Request节点或低代码工具 |

如果你不知道工具属于哪一种，先看它是否会自动显示最终请求地址。能看到最终地址，就能一眼发现是否重复拼了`/v1`。

## 2. Key在哪：去工作台复制令牌

根据4SAPI文档，注册登录后进入工作台的“令牌”页面，可以看到你创建的令牌列表。每一行令牌右侧都有复制按钮，点击复制出来的内容就是你的密钥Key。

建议你一开始就按环境拆令牌：

- `dev-test`：开发调试用，额度小，方便试错。
- `staging`：预发环境用，接近生产配置。
- `prod`：生产环境用，单独设置额度、期限和可用分组。

不要把同一个Key同时放到本地脚本、线上服务、第三方客户端和团队共享文档里。Key一旦泄露，最先受影响的是余额和调用日志，其次才是排查成本。

### 2.1 令牌命名建议

令牌名称不要随便写“test”或“key1”。后续排查日志时，你需要从名称就能看出来源。推荐格式：

```text
项目-环境-用途-负责人
```

示例：

```text
crm-prod-chatbot-zhangsan
docs-dev-embedding-lisi
n8n-prod-workflow-system
cursor-dev-personal-wangwu
```

这样当某个Key消耗异常时，你能快速定位到项目和负责人。

### 2.2 Key应该怎么存放

不同场景的存放方式也不一样：

- 本地脚本：放在`.env`文件，确保`.env`不提交到Git。
- 后端服务：放在部署平台的环境变量或密钥管理服务。
- n8n/Dify等工具：放在平台的Credential或模型供应商配置里。
- 桌面客户端：只放个人低额度Key，不放生产Key。
- CI/CD：单独创建自动化任务Key，并限制额度。

绝对不要把Key写进前端代码、Markdown教程截图、公开Issue或团队群聊截图里。很多泄露不是黑客攻击，而是复制粘贴时顺手暴露。

## 3. URL填哪个：先按工具类型判断

4SAPI文档里对URL的建议很直接：一般尝试下面两个地址即可，因为不同软件对OpenAI兼容接口的拼接方式不一样。

```text
https://4sapi.com
https://4sapi.com/v1
```

判断方法也简单：

- 如果工具让你填`Base URL`，通常先填`https://4sapi.com/v1`。
- 如果工具会自动拼接`/v1/chat/completions`，可尝试填`https://4sapi.com`。
- 如果工具让你填完整接口地址，文本生成通常填`https://4sapi.com/v1/chat/completions`。

这一步不需要玄学排查。第一次不通时，优先在这两个Base URL之间切换，然后再看Key和模型名。

### 3.1 如何判断最终请求地址

如果工具支持调试日志，建议打开后看最终请求地址。正确的文本生成地址应该类似：

```text
https://4sapi.com/v1/chat/completions
```

如果你看到下面这些形式，就要警惕：

```text
https://4sapi.com/v1/v1/chat/completions
https://4sapi.com/chat/completions
https://4sapi.com/v1/models/chat/completions
```

第一种通常是`/v1`重复；第二种通常是少了`/v1`；第三种通常是工具把模型列表接口和聊天接口混用了。

### 3.2 常见客户端字段翻译

不同工具字段名不同，但含义类似：

| 字段名 | 实际含义 |
| --- | --- |
| API Base | Base URL |
| Base URL | 基础接口地址 |
| API Host | 基础接口地址或主机 |
| Endpoint | 可能是基础地址，也可能是完整请求地址 |
| OpenAI Compatible URL | OpenAI兼容接口地址 |
| Model Name | 4SAPI模型名 |
| Custom Model | 手动填写模型名 |

不要只看字段名，要结合工具说明判断它是否会自动拼接路径。

## 4. Model怎么写：不要手打，完整复制

4SAPI文档明确提到：更换模型时，先确认你想用的模型已经在当前令牌包含的分组里，然后在代码或客户端的`Model`字段里填写平台模型广场中的模型名称。

也就是说，切换模型不是改URL，而是改`model`参数：

```json
{
  "model": "claude-sonnet-4-5-20250929",
  "messages": [
    {
      "role": "user",
      "content": "Hello!"
    }
  ]
}
```

如果你用的是OpenAI SDK，通常也是同一个思路：

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-xxxxxxxxxxxxxxxx",
    base_url="https://4sapi.com/v1",
)

response = client.chat.completions.create(
    model="claude-sonnet-4-5-20250929",
    messages=[
        {"role": "user", "content": "用一句话解释什么是API中转站。"}
    ],
)

print(response.choices[0].message.content)
```

模型名最稳妥的做法是从模型广场或价格页复制，不要凭记忆写。大模型版本更新很快，少一个后缀就可能变成“模型不存在”或“当前分组不可用”。

### 4.1 模型名、模型供应商和模型能力要分开看

模型名只是调用入口，不能代表能力一定相同。比如同样是“Claude”或“GPT”系列，不同版本可能在上下文长度、视觉输入、工具调用、结构化输出、价格和速率限制上完全不同。

选择模型时建议看六项：

- 是否支持文本生成。
- 是否支持图片理解。
- 是否支持工具调用。
- 是否支持结构化输出。
- 上下文窗口是否足够。
- 当前令牌分组是否允许调用。

不要在生产代码里写“任何模型都能做任何事”。更稳的做法是为每个业务场景维护一个模型配置表。

示例：

```json
{
  "chat_default": "deepseek-chat",
  "code_review": "claude-sonnet-4-5-20250929",
  "summary_fast": "gemini-2.5-flash",
  "fallback": "gpt-5.5"
}
```

这样后续切模型只改配置，不改业务代码。

## 5. 最小可用cURL测试

在接入任何客户端之前，建议先用cURL确认链路可用：

```bash
curl --location "https://4sapi.com/v1/chat/completions" \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer sk-xxxxxxxxxxxxxxxx" \
  --data '{
    "model": "claude-sonnet-4-5-20250929",
    "messages": [
      {
        "role": "user",
        "content": "Hello!"
      }
    ]
  }'
```

如果这里能返回结果，说明Key、URL、模型名三件事基本没问题。后面如果Cursor、Cherry Studio、Dify等工具不通，多半是工具里的URL拼接或鉴权格式差异。

### 5.1 再测一次流式输出

很多聊天产品需要流式输出，否则用户会觉得页面卡住。可以再测一次`stream`：

```bash
curl --location "https://4sapi.com/v1/chat/completions" \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer sk-xxxxxxxxxxxxxxxx" \
  --data '{
    "model": "claude-sonnet-4-5-20250929",
    "stream": true,
    "messages": [
      {
        "role": "user",
        "content": "分三点解释为什么要测试流式输出。"
      }
    ]
  }'
```

如果cURL能看到分段返回，但前端看不到，问题大概率在前端SSE解析、代理缓冲或网关超时设置上。

### 5.2 用环境变量测试，避免Key进命令历史

更安全的测试写法是：

```bash
export FOURSAPI_API_KEY="sk-xxxxxxxxxxxxxxxx"

curl --location "https://4sapi.com/v1/chat/completions" \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer ${FOURSAPI_API_KEY}" \
  --data '{
    "model": "claude-sonnet-4-5-20250929",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

Windows PowerShell可以用：

```powershell
$env:FOURSAPI_API_KEY="sk-xxxxxxxxxxxxxxxx"
```

这样Key不会直接散落在脚本、截图和聊天记录里。

## 6. 首次接入检查清单

排查顺序建议固定下来：

- 令牌是否从工作台复制，而不是复制了令牌名称。
- 账户余额是否可用，令牌本身是否有额度限制。
- 令牌是否包含你要调用的模型分组。
- Base URL是否在`https://4sapi.com`和`https://4sapi.com/v1`之间试过。
- 模型名称是否完整复制。
- 第三方客户端的Key输入框是否只填Key本体。
- 代码里是否把`Authorization`写成了当前SDK需要的格式。

这套顺序可以解决大多数“我明明充值了但接口没反应”的问题。先把基础三件套跑通，再去调复杂参数，会省很多时间。

## 7. 接入不同工具时的推荐顺序

如果你要同时接入多个工具，建议按这个顺序：

```text
1. cURL
2. Python或Node.js最小脚本
3. 桌面客户端
4. IDE插件
5. Dify/n8n/Coze等工作流平台
6. 生产后端服务
```

原因是越往后，工具自己的配置层越厚。先用cURL和脚本确认4SAPI链路没问题，再去排查第三方工具，会少走很多弯路。

## 8. 一张“能不能用”的判断表

| 现象 | 优先怀疑 | 处理方法 |
| --- | --- | --- |
| 401或无效令牌 | Key错误、Key过期、鉴权格式错误 | 重新复制令牌，检查是否需要Bearer |
| 404或模型不存在 | 模型名错误、URL路径错误 | 复制模型广场名称，检查最终URL |
| 403或无权限 | 分组不包含模型、令牌额度问题 | 检查令牌分组和额度 |
| 429 | 并发或速率限制 | 降低并发，增加退避重试 |
| 客户端没反应但cURL正常 | 客户端URL拼接或代理问题 | 切换`https://4sapi.com`和`/v1` |
| 有返回但格式不对 | 模型能力或参数不匹配 | 检查工具调用/结构化输出支持 |

这张表可以直接贴到团队Wiki里，给新同事做接入排查用。

## 9. 总结

4SAPI接入的第一关不是写复杂代码，而是把令牌、URL和模型名填对。只要你记住“Key去工作台令牌页复制，URL优先试`https://4sapi.com/v1`，模型名从模型广场完整复制”，绝大多数新手问题都能快速定位。

下一篇我们继续拆价格、充值、分组和Token日志，看看怎么在大模型API中转站里把成本算清楚。

参考资料：

- 4SAPI接入文档：https://4sapi.apifox.cn/
- Key位置说明：https://4sapi.apifox.cn/8423101m0
- URL配置说明：https://4sapi.apifox.cn/8423139m0
- 模型切换说明：https://4sapi.apifox.cn/8423128m0
- OpenAI文本生成文档：https://platform.openai.com/docs/guides/text-generation
- OpenAI SDK文档：https://platform.openai.com/docs/libraries
