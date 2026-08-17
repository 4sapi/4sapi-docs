---
title: "【大模型API中转站】第164期 AI排查前端白屏 | Console和Network"
category: 人工智能
tags:
  - 大模型API中转站
  - 4SAPI
  - AI排错
  - 前端白屏
  - Console
  - Network
  - Claude Fable 5
description: "前端白屏最怕只说网站打不开。本文给出一套 AI 排查前端白屏的证据包：Console、Network、构建版本、环境变量、资源路径、接口状态，并说明如何用 4SAPI 让辅助模型整理日志、Fable 5 判断根因。"
---

# 【大模型API中转站】第164期 AI排查前端白屏 | Console和Network

本文是【大模型API中转站】系列的第164篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

这一篇开始写具体报错。

第一个就写最常见、也最折磨人的问题：

```text
前端白屏。
```

用户眼里只有一句话：

```text
网站打不开。
```

但工程上，前端白屏至少可能来自十几个方向：

```text
JS 运行时报错。
资源 404。
接口 500。
CORS。
环境变量错。
路由 base path 错。
构建产物版本不一致。
CDN 缓存旧文件。
浏览器兼容问题。
登录态异常。
```

如果你直接问 AI：

```text
我的页面白屏了，怎么办？
```

它只能给你一堆泛泛建议。

真正高质量的用法是：

```text
先收集 Console、Network、构建版本、最近变更。
让辅助模型整理白屏证据包。
再让 Fable 5 判断优先排查路径。
```

## 1. 白屏先分三类

前端白屏不要上来就改代码。

先分类。

| 类型 | 表现 | 优先看什么 |
| --- | --- | --- |
| JS 崩溃 | 页面空白，Console 有红色错误 | Console stack |
| 资源失败 | JS/CSS 文件 404 或 MIME 错 | Network 静态资源 |
| 接口失败 | 页面框架在，但数据不出 | API 请求状态 |

再细一点：

```text
如果 index.html 能返回，但 js bundle 404，通常是部署路径或缓存问题。
如果 js bundle 能加载，但 Console 报 undefined/null，通常是代码运行时问题。
如果页面能渲染骨架，但接口 401/500，通常是 API 或登录态问题。
如果本地正常、线上白屏，优先看环境变量、base path、资源路径和构建版本。
```

这一步可以让低成本辅助模型做。

不用一上来就调用 Fable 5。

## 2. 给 AI 的白屏证据包

你给 AI 的材料至少包括：

```text
1. 用户看到的现象
2. 浏览器 Console 红色错误
3. Network 里失败的请求
4. 当前 URL
5. 是否登录后才白屏
6. 最近部署时间
7. 前端框架和构建工具
8. 本地是否正常
```

模板：

```text
【现象】
- 访问 URL：
- 白屏范围：所有用户 / 部分用户 / 仅登录后
- 是否刷新后恢复：
- 是否无痕模式复现：

【环境】
- 框架：React / Vue / Next.js / Nuxt / Vite / 其他
- 部署：Nginx / Vercel / Docker / CDN
- 最近部署版本：

【Console】
粘贴红色错误，保留 stack 前 20 行。

【Network】
- 失败请求 URL：
- 状态码：
- Response：
- Initiator：

【最近变更】
- 改了路由 / 环境变量 / 依赖 / 登录 / API / 构建配置：

【边界】
先给只读排查步骤，不要让我直接改生产配置。
```

这份证据包比“白屏了”强一百倍。

## 3. Console 怎么复制给 AI

打开浏览器开发者工具：

```text
F12 -> Console
```

只复制红色错误。

不要把所有 warning 都贴进去。

前端项目 warning 很多，真正导致白屏的通常是第一个红色 error。

常见关键词：

```text
Cannot read properties of undefined
Cannot access before initialization
Minified React error
ChunkLoadError
Loading chunk failed
Unexpected token '<'
Failed to fetch dynamically imported module
```

给 AI 时不要只贴一句错误。

最好保留：

```text
错误类型。
报错文件。
行号。
stack 前几层。
发生页面。
```

示例：

```text
Console:
TypeError: Cannot read properties of undefined (reading 'name')
    at UserCard (assets/index-a1b2c3.js:42:1034)
    at renderWithHooks (...)

页面：/dashboard
触发：登录后进入首页立即白屏
```

Fable 5 看到这个，会优先判断：

```text
用户数据为空。
接口返回结构变化。
组件没有兜底。
```

而不是跑去查 Nginx。

## 4. Network 怎么复制给 AI

打开：

```text
F12 -> Network -> 刷新页面
```

勾选：

```text
Preserve log
Disable cache
```

然后看红色请求。

常见情况：

| 状态 | 可能方向 |
| --- | --- |
| 404 JS/CSS | 构建产物路径、CDN 缓存、Nginx root |
| 401 API | 登录态、Token、Cookie、权限 |
| 403 API | 权限、CSRF、WAF、网关 |
| 500 API | 后端错误 |
| 502/504 | 网关或上游服务 |
| CORS | 跨域 Header 或预检失败 |
| `Unexpected token '<'` | JS 请求拿到了 HTML |

特别说一下：

```text
Unexpected token '<'
```

这个非常常见。

它通常意味着：

```text
浏览器以为自己在加载 JS，
结果服务器返回了一段 HTML。
```

比如：

```text
JS 文件路径错，Nginx 返回 index.html。
登录过期，接口返回登录页 HTML。
CDN 回源失败，返回错误页。
```

AI 看到这个关键词，就能少走很多弯路。

## 5. 前端白屏的 AI 分层 Prompt

可以直接复制：

```text
你是前端白屏排查助手。

请根据我提供的 Console、Network、最近变更和部署环境，判断白屏最可能属于哪一层：
1. JS 运行时崩溃
2. 静态资源加载失败
3. API 请求失败
4. 登录态或权限
5. 构建/部署路径
6. CDN 或缓存

要求：
- 先给最可能根因排序。
- 每个判断必须引用 Console 或 Network 证据。
- 先给只读验证步骤。
- 不要建议我直接清空线上缓存、重启生产服务或修改 DNS。
- 如果证据不足，列出还需要补充的 3 个信息。
```

这条适合走 4SAPI 里的中等模型或 Fable 5。

如果 Console 和 Network 很长，先让低成本模型整理：

```text
请从这段 Console 和 Network 日志中提取可能导致白屏的关键错误。
不要判断根因，只输出错误表格：时间、类型、URL/文件、状态码、关键消息。
```

然后再把表格交给 Fable 5。

## 6. 常见根因一：ChunkLoadError

报错：

```text
ChunkLoadError: Loading chunk xxx failed
```

常见原因：

```text
用户浏览器还拿着旧 index.html。
旧 index.html 指向旧 chunk。
服务器或 CDN 已经只保留新 chunk。
刷新后仍然请求不到旧文件。
```

AI 排查路径：

```text
1. Network 看失败 chunk URL。
2. 直接打开该 URL 是否 404。
3. 检查部署是否删除旧构建产物。
4. 检查 index.html 缓存策略。
5. 检查 CDN 是否缓存了旧 HTML。
```

修复方向通常是：

```text
index.html 不长缓存。
静态 chunk 用 hash 长缓存。
部署时短时间保留旧 chunk。
前端捕获 chunk load failure 后提示刷新。
```

不要让 AI 一上来重启服务器。

这通常不是服务器进程问题。

## 7. 常见根因二：环境变量错

前端本地正常，线上白屏，很常见是：

```text
VITE_API_BASE_URL 没配置。
NEXT_PUBLIC_API_URL 错。
构建时环境变量为空。
```

这类问题 Console 里可能不直接写“环境变量错”。

表现可能是：

```text
API 请求到了 undefined/api。
API 请求到了 localhost。
API 请求到了旧域名。
```

给 AI 的证据：

```text
Network 失败 URL。
构建环境变量名称。
部署平台环境变量截图的脱敏描述。
```

不要把真实 Key 发给 AI。

只给变量名和脱敏值：

```text
VITE_API_BASE_URL=https://api.example.com
SENTRY_DSN=[已配置，值不展示]
```

## 8. 常见根因三：接口结构变化

前端白屏也可能来自后端返回结构变了。

例如前端代码写：

```text
user.profile.name
```

但接口返回：

```json
{
  "user": null
}
```

或者：

```json
{
  "data": {
    "profile": {}
  }
}
```

Console 可能是：

```text
Cannot read properties of null
```

Network 里接口状态码还是 200。

这类很容易误判。

AI 要同时看：

```text
Console stack。
接口 response。
最近后端变更。
```

Prompt：

```text
请判断这个白屏是否可能由接口返回结构变化引起。
如果是，请指出前端读取路径、接口实际返回路径和需要加兜底的位置。
不要重写整个组件，只给最小修复建议。
```

## 9. 常见根因四：CORS

CORS 白屏通常表现为：

```text
Console 有 Access-Control-Allow-Origin。
Network 里预检 OPTIONS 失败。
或者请求被浏览器拦截，看不到正常 response。
```

AI 排查时要区分：

```text
后端没返回 CORS Header。
预检请求没有处理 OPTIONS。
带 Cookie 时没有 Access-Control-Allow-Credentials。
Origin 不在白名单。
Nginx 和后端重复设置 Header。
```

这类问题后面可以单独写一篇。

白屏文章里先记住：

```text
CORS 是浏览器拦截。
后端可能实际返回了内容，但前端 JS 拿不到。
```

不要让 AI 只看 HTTP 状态码。

要看 Console 的 CORS 文本。

## 10. 4SAPI 怎么分模型

前端白屏排查建议：

| 阶段 | 模型 |
| --- | --- |
| Console/Network 摘要 | 低成本模型 |
| 根因判断 | Fable 5 或强推理模型 |
| 修复方案审查 | Fable 5 / reviewer |
| 复盘文档 | 低成本模型 |

4SAPI 日志里记录：

```text
project
error_type: frontend_blank_screen
stage: log_summary / root_cause / fix_plan / report
model
cost
evidence_ids
final_status
```

以后你会发现：

```text
哪些白屏是资源路径。
哪些是接口结构。
哪些是登录态。
哪些是缓存。
```

这比每次临时问 AI 强很多。

## 11. 最小只读检查清单

给 AI 或自己都可以用：

```text
1. 无痕窗口是否复现。
2. Console 第一个红色 error 是什么。
3. Network 是否有 JS/CSS 404。
4. Network 是否有 API 401/403/500/502。
5. 是否只有某个路由白屏。
6. 是否登录后才白屏。
7. 最近是否部署过前端。
8. 最近是否改过环境变量。
9. 是否 CDN 缓存旧 index.html。
10. 本地生产构建是否可复现。
```

注意：

```text
先只读检查。
再决定修复。
```

## 12. 总结

前端白屏不是一个报错。

它是一类症状。

AI 排查前端白屏，关键不是让模型“猜怎么修”。

关键是给它：

```text
Console。
Network。
最近变更。
部署环境。
复现范围。
风险边界。
```

通过 4SAPI，你可以让低成本模型先整理日志，再让 Fable 5 做根因判断，最后用复核模型检查修复方案有没有高风险动作。

一句话：

```text
白屏先看证据，不要先改代码。
```

下一篇写后端 500。

前端白屏是用户看到的结果。

后端 500 往往是背后的真正爆点。
