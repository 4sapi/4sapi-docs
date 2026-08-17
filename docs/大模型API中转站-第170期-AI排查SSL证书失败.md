---
title: "【大模型API中转站】第170期 AI排查SSL证书失败 | HTTPS续期和链路"
category: 人工智能
tags:
  - 大模型API中转站
  - 4SAPI
  - AI排错
  - SSL
  - HTTPS
  - Nginx
  - 证书续期
description: "SSL 证书失败可能来自域名解析、80端口、Nginx配置、证书链、续期任务和时间同步。本文给出适合交给 AI 的 HTTPS 排错包，并说明 Fable 5 如何审查证书续期风险。"
---

# 【大模型API中转站】第170期 AI排查SSL证书失败 | HTTPS续期和链路

本文是【大模型API中转站】系列的第170篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

HTTPS 出问题，用户看到的是：

```text
不安全。
证书过期。
连接不是私密连接。
SSL handshake failed。
```

但工程上可能来自：

```text
证书过期。
域名解析错。
80 端口不通。
Nginx 配置错。
证书链不完整。
自动续期失败。
服务器时间不准。
CDN 和源站证书不一致。
```

这类问题非常适合让 AI 做检查清单。

但不适合让 AI 直接乱改证书配置。

因为一改错，全站 HTTPS 都可能挂。

## 1. 给 AI 的 SSL 排错包

```text
【现象】
- 浏览器报错原文：
- 域名：
- 是否所有地区：
- 是否 CDN 后：
- 开始时间：

【环境】
- Web 服务器：Nginx / Caddy / Apache
- 证书来源：Let's Encrypt / 云厂商 / CDN
- 是否 Docker：
- 是否有 CDN：

【证据】
- openssl s_client 输出摘要：
- curl -Iv 域名：
- nginx -t：
- Nginx SSL 配置脱敏：
- 证书到期时间：
- 续期任务日志：

【边界】
- 不删除旧证书
- 不覆盖生产配置
- 不关闭 HTTPS
- 先给只读验证步骤
```

这份材料足够让 Fable 5 判断链路。

## 2. 只读检查命令

查看证书：

```bash
echo | openssl s_client -servername example.com -connect example.com:443 2>/dev/null | openssl x509 -noout -dates -issuer -subject
```

curl：

```bash
curl -Iv https://example.com
```

检查 Nginx：

```bash
sudo nginx -t
```

看续期：

```bash
sudo certbot certificates
sudo certbot renew --dry-run
```

注意：

```text
dry-run 是演练，不是真续期。
```

适合先给 AI 看输出。

## 3. 常见根因一：证书过期

最直观：

```text
notAfter 已经过了当前日期。
```

AI 需要判断：

```text
证书真的过期。
还是客户端时间错。
还是 CDN 边缘节点还在用旧证书。
```

给它：

```text
服务器时间。
证书 notBefore/notAfter。
浏览器报错。
CDN 是否开启。
```

不要只说“过期了”。

## 4. 常见根因二：续期失败

Let's Encrypt 续期失败常见原因：

```text
80 端口不通。
域名解析不到当前服务器。
Nginx 配置拦截了 challenge。
防火墙没开。
证书路径变了。
certbot timer 没运行。
```

AI 排查顺序：

```text
1. 域名 A 记录。
2. 80 端口是否可访问。
3. /.well-known/acme-challenge/ 是否被正确处理。
4. certbot renew --dry-run 输出。
5. timer 状态。
```

不要让 AI 直接反复申请证书。

频繁失败可能触发限额。

## 5. 常见根因三：证书链不完整

有些客户端报：

```text
unable to get local issuer certificate
```

可能是证书链不完整。

Nginx 应该使用 fullchain，而不是只用 cert。

让 AI 看：

```nginx
ssl_certificate
ssl_certificate_key
```

判断是否指向：

```text
fullchain.pem
```

而不是单独 cert。

## 6. 常见根因四：CDN 和源站证书混乱

如果前面有 CDN，证书可能分两层：

```text
用户 -> CDN 边缘证书
CDN -> 源站证书
```

用户看到的证书，不一定是源站 Nginx 的证书。

AI 要知道：

```text
是否开 CDN。
CDN SSL 模式。
源站证书状态。
浏览器实际看到的 issuer。
```

否则容易把 CDN 问题误判成源站问题。

## 7. AI Prompt

```text
你是 HTTPS/SSL 排查助手。

请根据浏览器报错、openssl/curl 输出、Nginx 配置、证书到期时间和续期日志，判断问题属于：
1. 证书过期
2. 域名解析错误
3. 80/443 端口或防火墙问题
4. Nginx SSL 配置错误
5. 证书链不完整
6. 自动续期失败
7. CDN 和源站证书不一致
8. 服务器时间异常

要求：
- 先给只读验证步骤。
- 不要建议关闭 HTTPS。
- 不要覆盖旧证书，除非先备份。
- 涉及 DNS/CDN/Nginx 生产配置时标注人工确认。
```

## 8. 4SAPI 和证书运维

证书问题本身不是模型 API 问题。

但对于企业级大模型接入很关键。

如果你的应用通过 4SAPI 提供 AI 功能，HTTPS 挂了会导致：

```text
前端无法请求后端。
Webhook 回调失败。
企业客户接口不可用。
Agent 工作流中断。
```

建议让 AI 定期生成证书巡检报告：

```text
域名。
证书到期时间。
续期方式。
CDN 状态。
负责人。
续期风险。
```

简单日志摘要用低成本模型。

证书链和 CDN 复杂问题交给 Fable 5 审查。

## 9. 证书巡检表怎么写

如果你的项目已经上线，不要等浏览器报“不安全”才想起证书。

可以让 AI 帮你维护一张证书巡检表。

字段建议：

| 字段 | 说明 |
| --- | --- |
| 域名 | 主站、API、后台、Webhook |
| 证书来源 | Let's Encrypt、云厂商、CDN |
| 当前到期日 | notAfter |
| 续期方式 | certbot、云控制台、CDN 自动续期 |
| 续期验证方式 | HTTP-01、DNS-01、云厂商验证 |
| 负责人 | 谁收到告警 |
| 风险等级 | 7天内到期、30天内到期、正常 |
| 最近演练 | dry-run 或人工检查时间 |

Prompt：

```text
请根据下面的域名和证书检查输出，生成一份证书巡检表。
重点标注 30 天内到期、续期方式不明确、CDN 和源站证书不一致、没有负责人
这四类风险。

不要输出任何私钥内容。
不要建议关闭 HTTPS。
```

这类任务不需要 Fable 5。

低成本模型就能整理。

但如果发现：

```text
CDN 边缘证书正常，源站证书异常。
证书链不完整。
DNS-01 验证记录冲突。
多环境证书路径混用。
```

就值得升级给 Fable 5 做判断。

## 10. 续期演练比临时修复更重要

证书事故最难受的地方在于：

```text
它常常发生在你没盯着的时候。
```

比如凌晨、周末、节假日。

所以不要只问 AI：

```text
证书过期怎么修？
```

更应该让它帮你设计续期演练：

```text
请根据当前 Nginx、certbot 和域名配置，设计一套证书续期演练。

要求：
1. 先做只读检查。
2. 使用 dry-run 验证自动续期。
3. 不覆盖现有证书。
4. 不停止生产 Nginx。
5. 说明失败时需要检查的日志。
6. 输出回滚和人工确认点。
```

演练通过以后，再把结果写进运维文档：

```text
证书在哪里。
怎么续期。
续期失败看哪个日志。
谁负责处理。
提前多少天告警。
```

这就是 AI 运维最有价值的地方。

不是等出事了临时问。

而是把容易出事的点提前做成清单。

## 11. 总结

SSL 证书失败，不要第一时间乱改 Nginx。

先收集：

```text
浏览器报错。
openssl。
curl。
nginx -t。
证书到期时间。
续期日志。
CDN 状态。
```

AI 的价值是把 HTTPS 链路拆开。

不是帮你盲目重签证书。

一句话：

```text
证书问题先看链路，再动配置。
```
