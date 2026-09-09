---
title: "405 Not Allowed 怎么排查：请求方法与网关路由"
category: 人工智能
tags:
  - HTTP 状态码
  - API 排错
  - OpenResty
description: "从一段 OpenResty 返回的 405 HTML 页面出发，说明如何检查请求方法、接口路径、客户端配置和反向代理路由。"
---

# 405 Not Allowed 怎么排查：请求方法与网关路由

调用聊天补全接口时，如果收到的不是 JSON，而是一张写着 `405 Not Allowed` 和 `openresty` 的 HTML 页面，问题通常发生在模型请求进入业务接口之前。HTTP 服务器或网关识别了当前 URL，却不允许这次请求使用的方法。本文给出一条从客户端到网关的排查顺序，目标是确认请求究竟发成了什么，以及它落到了哪条路由。

## 1. 这段 HTML 能说明什么

典型响应如下：

```html
<html>
<head>
    <title>405 Not Allowed</title>
</head>
<body>
    <center><h1>405 Not Allowed</h1></center>
    <hr>
    <center>openresty</center>
</body>
</html>
```

根据 HTTP 语义，`405 Method Not Allowed` 表示服务器知道该请求方法，但目标资源不支持它。页面底部的 `openresty` 表明这份错误页由 OpenResty 相关网关层返回；仅凭这一行，不能判断后端模型服务是否收到请求。

这与几个相邻错误不同：

| 状态码 | 常见含义 | 首要检查 |
| --- | --- | --- |
| `400` | 请求内容不符合接口要求 | JSON 和参数类型 |
| `401` | 未通过身份认证 | Authorization 头与 Key |
| `403` | 已识别身份但无权限 | 账号、模型和策略权限 |
| `404` | 没有匹配的目标资源 | 域名和接口路径 |
| `405` | 路径存在，但方法不被允许 | GET、POST、OPTIONS 等方法 |

实际网关可能对错误进行改写，因此这张表用于确定检查顺序，不替代服务端日志。

## 2. 先确认请求不是 GET

聊天补全接口一般要求客户端发送 `POST`。如果把接口地址直接粘进浏览器地址栏，浏览器会发出 `GET`，很可能得到 405。

在 Postman 或 Apifox 中也要明确选择 `POST`。只在 Body 中粘贴 JSON，不会自动把方法从 `GET` 改成 `POST`。

可以用下面的最小请求观察响应头和响应体：

```bash
curl -i -X POST "https://api.example.com/v1/chat/completions" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  --data-binary '{
    "model": "example-model",
    "messages": [
      {"role": "user", "content": "Hello"}
    ]
  }'
```

`-i` 会把响应头一起输出。不要使用 `-v` 截图后直接公开，因为调试输出可能包含 Authorization 头。

## 3. 检查 URL 是否落到正确接口

以下地址看起来相近，路由含义可能完全不同：

```text
https://api.example.com/
https://api.example.com/v1
https://api.example.com/v1/chat/completions
https://api.example.com/api/v1/chat/completions
```

多一个或少一个 `/api`，都可能让请求落到网页路由、管理后台或默认网关。某些页面允许浏览器读取，却不允许 `POST`；另一些 API 路由只允许 `POST`，用浏览器访问则返回 405。

排查时从供应商当前文档复制完整路径，不要根据其他兼容接口猜路径。还要检查环境变量拼接是否产生重复斜杠、漏掉版本号，或把 Base URL 与完整 endpoint 重复拼接。

## 4. 检查客户端是否改写了方法

用户界面显示 `POST`，不一定等于线上实际收到 `POST`。以下环节都可能改变请求：

- 前端跨域请求先发送 `OPTIONS` 预检，但网关未配置该方法；
- HTTP 跳转后，客户端改变了原请求方法；
- 工作流节点的“测试 URL”功能只发 `GET`；
- SDK 的 Base URL 配置错误，最终请求落到健康检查页面；
- 代理规则只转发部分方法。

客户端侧应保存最终 URL、最终 Method、响应状态和请求 ID。服务端侧则检查网关 access log 中的 `$request_method`、URI 和 upstream 状态。两边记录对上，才能确认方法在哪一层发生变化。

## 5. 如果自己维护 OpenResty

先检查对应 `location` 是否限制了方法。例如：

```nginx
location /v1/chat/completions {
    limit_except POST {
        deny all;
    }

    proxy_pass http://chat_backend;
}
```

这段配置只用于说明限制位置，不建议为了消除 405 直接开放所有方法。应先确认业务接口本来需要哪些方法，再为 `POST` 和必要的 `OPTIONS` 配置明确路由。

还要检查：

```text
是否命中了预期 location
是否存在更具体的 location 抢先匹配
proxy_pass 拼接后的上游路径是否正确
OPTIONS 是否由网关处理
错误页来自网关还是上游
```

修改配置后先执行语法检查，再按当前部署流程重载。不要在未确认配置文件位置和进程归属时直接操作生产服务。

## 6. 用响应头缩小范围

按 RFC 9110 的源服务器语义，405 响应应生成 `Allow` 头，列出目标资源当前支持的方法。实际服务中如果能看到：

```http
Allow: POST
```

就可以直接确认当前请求使用了错误方法。如果没有 `Allow`，仍可继续用 access log、请求 ID 和代理日志定位；不能因为响应头不完整，就反推模型接口本身故障。

`Server: openresty`、CDN 标识、网关请求 ID 和 upstream 相关头也能帮助判断错误来自哪一层，但这些头可能被代理隐藏或改写。

## 7. 一条可执行的排查顺序

按下面顺序检查，通常比反复修改 Body 更快：

1. 确认客户端发送的是 `POST`，不是浏览器 `GET` 或预检 `OPTIONS`；
2. 从当前文档核对完整 URL；
3. 确认 `Content-Type: application/json` 与 Authorization 头存在；
4. 用 `curl -i` 保存状态码、响应头和非敏感响应体；
5. 检查是否发生跳转，以及跳转后的最终方法；
6. 有网关权限时，对照 access log 的 Method、URI 和 upstream 状态；
7. 方法和路由确认无误后，再检查 JSON Body 和模型参数。

如果请求尚未进入聊天补全路由，继续调整 `messages` 或 `model` 字段通常不会改变 405。

## 8. 结论与限制

`405 Not Allowed` 的核心是“方法与目标资源不匹配”。浏览器直接打开 POST 接口、客户端仍处于 GET、URL 落到错误路由，以及网关未处理 OPTIONS，都是优先检查项。OpenResty 错误页只能说明返回页面的网关层特征，不能单独证明模型服务故障。

公开排错信息时应删除 API Key、Authorization 头和敏感请求体，同时保留状态码、最终 URL 的非敏感部分、请求方法、时间和请求 ID。

## 官方来源

- [RFC 9110：405 Method Not Allowed](https://www.rfc-editor.org/rfc/rfc9110.html#name-405-method-not-allowed)
- [OpenResty：官方文档入口](https://openresty.org/en/)

---

*本文协议与链接核对日期：2026-08-10。具体接口允许的方法和代理配置应以目标服务当前文档及实际日志为准。*
