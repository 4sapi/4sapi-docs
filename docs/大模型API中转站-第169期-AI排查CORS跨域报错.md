---
title: "【大模型API中转站】第169期 AI排查CORS跨域 | 预检请求和Cookie"
category: 人工智能
tags:
  - 大模型API中转站
  - 4SAPI
  - AI排错
  - CORS
  - 前端
  - Nginx
  - 企业级API
description: "CORS 报错不是后端没响应，而是浏览器不让前端读。本文讲如何把 Origin、OPTIONS 预检、响应 Header、Cookie、Nginx 和后端配置整理成 AI 排错包，并用 4SAPI 路由模型排查。"
---

# 【大模型API中转站】第169期 AI排查CORS跨域 | 预检请求和Cookie

本文是【大模型API中转站】系列的第169篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

CORS 是前端开发里最容易让人烦躁的报错之一。

Console 里一长串：

```text
Access to fetch at ... from origin ... has been blocked by CORS policy
```

很多人第一反应：

```text
后端挂了。
```

不一定。

CORS 的关键是：

```text
浏览器不允许前端 JS 读取响应。
```

后端可能确实返回了内容。

但浏览器因为跨域规则，把它拦了。

这类问题适合 AI 排查，但必须把证据给全。

## 1. CORS 排查要看四件事

```text
Origin。
预检 OPTIONS。
响应 Header。
是否带 Cookie 或 Authorization。
```

只说“跨域了”没用。

给 AI 的材料要包括：

```text
前端域名。
后端域名。
请求方法。
请求头。
是否带 credentials。
OPTIONS 返回状态。
Access-Control-Allow-Origin。
Access-Control-Allow-Credentials。
Access-Control-Allow-Headers。
Access-Control-Allow-Methods。
```

这才是工程证据。

## 2. 给 AI 的 CORS 排错包

```text
【现象】
- 浏览器 Console 完整 CORS 报错：
- 请求 URL：
- 前端 Origin：
- 是否登录后才失败：

【Network】
- 失败请求 Method：
- 是否有 OPTIONS 预检：
- OPTIONS 状态码：
- 实际请求状态码：
- Response Headers：
- Request Headers：

【配置】
- 后端 CORS 配置：
- Nginx 配置：
- 是否带 Cookie / Authorization：

【边界】
- 不关闭浏览器安全策略
- 不使用 * 配合 credentials
- 不建议前端绕过 CORS
- 先给服务端正确配置方向
```

这份包能防止 AI 给出危险建议。

## 3. 常见错误一：用 `*` 配 Cookie

如果请求带：

```js
credentials: "include"
```

后端不能返回：

```text
Access-Control-Allow-Origin: *
```

还必须返回：

```text
Access-Control-Allow-Credentials: true
```

而 Allow-Origin 要是具体域名：

```text
https://app.example.com
```

不是 `*`。

AI Prompt：

```text
请检查这个 CORS 配置是否错误地把 Access-Control-Allow-Origin: * 和 credentials 一起使用。
如果是，请给出允许指定 Origin 的安全配置，不要建议关闭 credentials。
```

## 4. 常见错误二：OPTIONS 没处理

复杂请求会先发 OPTIONS 预检。

比如：

```text
POST JSON。
带 Authorization。
自定义 Header。
非简单 Content-Type。
```

如果后端或 Nginx 没处理 OPTIONS，浏览器会拦截。

常见表现：

```text
OPTIONS 404。
OPTIONS 405。
OPTIONS 500。
No Access-Control-Allow-Origin header。
```

给 AI：

```text
OPTIONS 请求状态码。
OPTIONS response headers。
后端路由是否支持 OPTIONS。
```

不要只看实际 POST。

很多时候 POST 根本没发出去。

## 5. 常见错误三：Header 不允许

报错可能写：

```text
Request header field authorization is not allowed
```

说明后端缺：

```text
Access-Control-Allow-Headers: Authorization, Content-Type
```

如果你自定义：

```text
x-project-id
x-api-key
```

也要允许。

但注意：

```text
不要把真实 API Key 放前端。
```

如果前端直接请求 4SAPI 或模型 API，并把 Key 放浏览器里，这是架构风险。

更推荐：

```text
前端 -> 自己后端 -> 4SAPI
```

由后端保管 4SAPI Key。

## 6. 常见错误四：Nginx 和后端重复设置

有时后端已经设置 CORS。

Nginx 又加一遍。

结果响应头重复：

```text
Access-Control-Allow-Origin: https://a.com, https://a.com
```

浏览器仍然会拒绝。

AI 要看：

```text
后端 CORS 中间件。
Nginx add_header。
实际 response headers。
```

Prompt：

```text
请判断 CORS Header 是由 Nginx 设置、后端设置，还是两边重复设置。
请基于实际 Response Headers 给结论。
```

## 7. 只读验证方法

用浏览器 Network 最准。

也可以用 curl 模拟 Origin：

```bash
curl -i -X OPTIONS https://api.example.com/path \
  -H "Origin: https://app.example.com" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: authorization,content-type"
```

看返回有没有：

```text
Access-Control-Allow-Origin
Access-Control-Allow-Methods
Access-Control-Allow-Headers
Access-Control-Allow-Credentials
```

把输出脱敏后给 AI。

## 8. 4SAPI 场景里的 CORS

如果你把 4SAPI Key 放在前端直接请求模型 API，会遇到两个问题。

第一，安全：

```text
Key 会暴露给浏览器用户。
```

第二，CORS：

```text
模型 API 或中转 API 不一定允许你的浏览器 Origin。
```

更稳架构：

```text
浏览器 -> 你的后端 API -> 4SAPI -> 模型
```

后端负责：

```text
保管 Key。
做权限。
做预算。
做日志审计。
做 CORS。
```

这才是企业级大模型接入的正确姿势。

## 9. AI Prompt

```text
你是 CORS 排查助手。

请根据 Console 报错、Network 里的 OPTIONS 和实际请求、Request/Response Headers、Nginx 和后端配置，判断跨域失败原因。

重点检查：
1. Origin 是否被允许
2. OPTIONS 预检是否返回正确
3. 是否带 credentials
4. Allow-Origin 是否错误使用 *
5. Allow-Headers 是否缺 Authorization 或自定义 Header
6. Nginx 和后端是否重复设置

要求：
- 不建议关闭浏览器安全策略。
- 不建议前端保存 4SAPI 或模型 API Key。
- 给服务端最小修复方向。
- 每个判断必须引用 Header 或 Console 证据。
```

### 实战补充：带 Cookie 的跨域登录

最容易出问题的是登录态。

比如前端在：

```text
https://app.example.com
```

后端在：

```text
https://api.example.com
```

请求需要带 Cookie。

这时要同时满足：

```text
前端 fetch 设置 credentials: include。
后端返回 Access-Control-Allow-Credentials: true。
后端 Access-Control-Allow-Origin 不能是 *。
Cookie 设置 SameSite=None; Secure。
OPTIONS 预检正常返回。
```

让 AI 排查时，不要只贴 Console。

还要贴 Cookie 设置和 Response Headers。

### 落地清单：CORS 上线检查

```text
是否只允许可信 Origin。
是否处理 OPTIONS。
是否允许必要 Header。
是否避免 * + credentials。
是否没有把 4SAPI Key 放到前端。
是否区分 dev/staging/prod 域名。
```

很多 CORS 问题最后不是技术问题。

是环境管理问题：

```text
本地能调。
测试环境能调。
生产换域名后忘了加白名单。
```

把 Origin 白名单写进部署 SOP，比每次线上临时改更稳。

## 10. 总结

CORS 报错不是“接口没返回”这么简单。

它是浏览器安全模型在工作。

AI 排查 CORS，要给：

```text
Origin。
OPTIONS。
Request Headers。
Response Headers。
credentials。
Nginx 和后端配置。
```

如果你的前端要用大模型，记住：

```text
不要把 4SAPI Key 放浏览器。
前端请求自己的后端。
后端再通过 4SAPI 调模型。
```

这样既解决 CORS，也解决 Key 泄露和成本审计。
