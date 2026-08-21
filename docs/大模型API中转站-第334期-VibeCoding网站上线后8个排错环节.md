---
title: "网站上线后 8 个常见故障：域名、HTTPS 与数据"
tags:
  - 网站部署
  - 故障排查
  - Web 运维
description: "从现象出发排查网站无法访问、证书失败、接口报错和数据丢失风险，给出每个环节的检查命令与修复方向。"
---
# 网站上线后 8 个常见故障：域名、HTTPS 与数据

网站上线后，最常见的误区是把所有问题都归结为“代码有 bug”。实际上，浏览器打开一个地址要经过 DNS、网络端口、反向代理、应用进程、环境变量、数据库和文件存储多个环节。任何一层出错，用户看到的都可能只是一个白屏、超时或 `502`。

本文提供一套从外到内的排查顺序，适合已经部署了 Vibe Coding 产品、但访问结果不符合预期的场景。每个环节都包含现象、检查方法和失败后的方向。命令需要在对应位置运行：DNS 命令可以在本地运行，服务和日志命令通常要在服务器上运行。

## 先记住一条排查原则

先确认“请求有没有到服务器”，再确认“服务器有没有把请求交给应用”，最后才看业务代码。不要在 DNS 还没生效时修改数据库，也不要在应用根本没监听端口时反复申请证书。

可以先画出请求链路：

```text
浏览器
  -> DNS 解析
  -> 云安全组 / 防火墙
  -> Nginx 或其他网关
  -> 应用进程
  -> 数据库 / 对象存储 / 第三方服务
```

## 排查前先收集一份证据包

故障发生时，先记录域名、开始时间、最近一次发布、最近一次 DNS 或证书变更，以及影响范围。比如“所有用户打不开”和“只有登录用户报错”不是同一类问题。

在本地收集域名和端口结果：

```powershell
Resolve-DnsName example.com -Type A
Test-NetConnection example.com -Port 80
Test-NetConnection example.com -Port 443
curl.exe -I https://example.com
```

在服务器收集服务状态和最近日志：

```bash
date -Is
sudo systemctl status nginx --no-pager
sudo ss -lntp
sudo tail -n 50 /var/log/nginx/error.log
sudo journalctl -u vibe-app --since "30 minutes ago" --no-pager
```

把这些输出保存到一次故障记录里，并先脱敏 Cookie、令牌、邮箱、用户内容和数据库连接串。排错的目标是建立证据链，而不是收集越多日志越好。日志中如果没有请求记录，优先排查 DNS、端口和网关；有请求记录但没有应用日志，再看 Nginx 转发；应用收到请求后才进入业务逻辑和数据层。

## 1. 域名解析到了错误地址

### 常见现象

刚部署的域名打不开，或者打开的是旧网站、云平台默认页。控制台里显示 DNS 记录已经保存，但不同网络看到的结果不一致。

### 检查方法

Windows PowerShell：

```powershell
Resolve-DnsName example.com -Type A
Resolve-DnsName www.example.com -Type A
```

Linux 或 macOS：

```bash
dig +short example.com A
dig +short www.example.com A
```

将输出的 IP 与服务器公网 IP 对比。如果使用了 IPv6，还要检查 `AAAA` 记录：

```bash
dig +short example.com AAAA
```

### 修复方向

确认记录类型、主机名和目标值没有填错。`@`、空主机名、`www` 在不同 DNS 控制台里的填写方式可能不同。若存在错误的 `AAAA` 记录，而服务器并未正确配置 IPv6，部分用户会优先连接到不可用的 IPv6 地址，此时应删除错误记录或完整配置 IPv6。

DNS 有缓存，修改后不会保证所有网络立即更新。持续查到旧结果时，先确认查询的 DNS 服务商和 TTL，再等待缓存自然过期。不要用频繁修改记录的方式“碰运气”。DNS 基础概念可参考 ICANN 的说明：[How Domain Name System (DNS) Works](https://www.icann.org/resources/pages/dns-2012-02-25-en)。

## 2. 云安全组或服务器防火墙没有放行

### 常见现象

DNS 已经返回正确 IP，但浏览器一直转圈，`curl` 超时，Nginx 日志里甚至没有请求记录。

### 检查方法

服务器上检查监听端口：

```bash
sudo ss -lntp
```

至少应该看到业务实际使用的 `80` 或 `443`。再检查 UFW：

```bash
sudo ufw status verbose
```

从本地测试端口：

```powershell
Test-NetConnection example.com -Port 80
Test-NetConnection example.com -Port 443
```

### 修复方向

云平台安全组和服务器内部防火墙是两套控制面。两者都要放行实际端口，且 Nginx 必须确实监听该端口。只允许 `80` 而没有 `443`，会导致 HTTPS 访问失败；只开放 `443` 但证书申请依赖 HTTP 验证时，也可能无法完成首次签发。

不要把数据库端口、Redis 端口和开发调试端口直接开放到 `0.0.0.0/0`。需要远程管理时，优先使用 SSH 隧道、私有网络或限制来源 IP。

## 3. Nginx 配置错误或命中了默认站点

### 常见现象

看到 Nginx 默认欢迎页、`404`、`403` 或 `502 Bad Gateway`。同一台服务器上的多个域名可能打开了错误的站点。

### 检查方法

先验证语法：

```bash
sudo nginx -t
```

查看已加载的完整配置：

```bash
sudo nginx -T
```

检查服务状态和最近日志：

```bash
sudo systemctl status nginx --no-pager
sudo journalctl -u nginx -n 100 --no-pager
sudo tail -n 100 /var/log/nginx/error.log
```

### 修复方向

检查 `server_name` 是否包含正在访问的域名，`root` 是否指向实际上传的目录，`index.html` 是否存在。单页应用通常需要：

```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

如果是 `502`，重点看 `proxy_pass` 指向的地址和应用监听端口。Nginx 官方文档是核对 `server`、`location` 和反向代理行为的依据：[Nginx Documentation](https://nginx.org/en/docs/)。修改后按顺序执行：

```bash
sudo nginx -t
sudo systemctl reload nginx
```

## 4. HTTPS 证书申请失败或没有自动续期

### 常见现象

浏览器提示证书不受信任、证书域名不匹配、证书已过期，或者 HTTP 可以访问而 HTTPS 不行。

### 检查方法

查看证书实际覆盖的域名和到期时间：

```bash
openssl s_client -connect example.com:443 -servername example.com < /dev/null 2>/dev/null | openssl x509 -noout -subject -issuer -dates -ext subjectAltName
```

检查 Certbot 的续期任务：

```bash
sudo certbot certificates
sudo certbot renew --dry-run
```

### 修复方向

证书域名必须与访问地址匹配。例如只为 `example.com` 申请证书，不代表 `www.example.com` 也被覆盖。证书申请失败时，先检查 DNS、`80/443` 端口、Nginx 配置和是否存在另一套代理或 CDN。

`renew --dry-run` 通过只代表当前模拟续期成功，不代表之后永远不会失败。续期任务、域名解析和网关配置发生变化后，都应重新验证。Certbot 的参数和验证方式以官方文档为准：[Using Certbot](https://eff-certbot.readthedocs.io/en/stable/using.html)。

## 5. 前端能打开，但接口返回 401、403 或 404

### 常见现象

首页正常，登录失败，文章列表为空，浏览器 Network 面板里接口返回 `401`、`403` 或 `404`。

### 检查方法

先在浏览器开发者工具中记录：请求 URL、方法、状态码、响应体和请求头。服务器侧查看应用和 Nginx 日志：

```bash
sudo tail -n 100 /var/log/nginx/access.log
sudo journalctl -u vibe-app -n 100 --no-pager
```

如果应用使用 `/api` 前缀，可以从服务器本机测试：

```bash
curl -i http://127.0.0.1:3000/health
curl -i https://example.com/api/health
```

### 修复方向

本机接口正常、域名接口异常，通常看 Nginx 的路径转发、请求头、跨域和 HTTPS 终止配置。本机接口也失败，则看应用路由、环境变量、数据库连接和认证中间件。

不要把 `401` 直接改成匿名放行。先确认前端发送的 Cookie 或 Authorization 头是否存在，服务器是否设置了正确的 `Secure`、`HttpOnly`、`SameSite` 属性，以及反向代理是否传递了原始协议和主机信息。

## 6. 部署后出现白屏或静态资源 404

### 常见现象

HTML 返回 `200`，但页面白屏；开发者工具显示 `.js`、`.css`、字体或图片返回 `404`。本地开发正常，部署后失败。

### 检查方法

检查构建产物里的资源路径：

```bash
find /var/www/vibe-site -maxdepth 2 -type f | sort | head -50
```

再检查线上 HTML 引用的路径：

```bash
curl -s https://example.com | head -40
```

### 修复方向

常见原因包括：部署时漏传了构建目录、站点部署在子路径却仍使用根路径、构建时环境变量为空、文件名大小写与 Linux 文件系统不一致。重新构建前先确认项目的 `base`、`publicPath` 或类似配置，具体字段以框架文档为准。

不要直接把开发服务器的源码目录当成生产静态目录。开发服务器可能会动态转换模块、代理接口或注入热更新脚本，上传源码不能替代正式构建。

## 7. 环境变量、数据库和文件上传在生产环境失效

### 常见现象

本地登录、数据写入和图片上传都正常，线上却出现数据库连接失败、上传权限错误或“配置未找到”。

### 检查方法

先区分前端变量和后端变量。前端打包后的 JavaScript 会发送给浏览器，任何进入前端构建产物的值都不应被当成秘密。服务器上只检查变量是否存在，不要把真实值打印到日志：

```bash
test -n "$DATABASE_URL" && echo "DATABASE_URL is set" || echo "DATABASE_URL is missing"
```

检查应用启动日志、数据库网络策略和对象存储权限。若应用由 systemd 启动，检查服务文件加载的环境变量路径；若使用容器，检查 Compose 或平台的环境配置是否已注入。

### 修复方向

生产环境变量要通过服务器环境、密钥管理或平台配置注入，而不是提交到仓库。数据库账号按最小权限创建，上传目录或对象存储桶也要区分公开读取和私有写入。

如果用户能上传文件，除了权限，还要限制文件大小、类型和文件名，避免把上传目录变成可执行目录。数据和文件应分别制定备份策略：数据库备份不等于对象存储备份，代码备份也不等于用户内容备份。

## 8. 重启后服务消失，或者没有可恢复备份

### 常见现象

部署当天可以访问，服务器重启、进程崩溃或发布新版本后网站就离线；回滚时找不到上一版文件，也不知道数据库是否已经被修改。

### 检查方法

查看服务是否设置开机启动：

```bash
sudo systemctl is-enabled nginx
sudo systemctl is-enabled vibe-app
sudo systemctl status vibe-app --no-pager
```

检查应用最近退出原因：

```bash
sudo journalctl -u vibe-app --since "1 hour ago" --no-pager
```

确认备份文件实际存在，并进行一次恢复演练。只看“备份任务成功”不够，无法恢复的备份不能算完成保护。

### 修复方向

为应用选择一种明确的进程管理方式，例如 systemd 或 Docker Compose，并记录启动、停止、查看日志和回滚命令。发布时保留旧版本目录，采用“新版本验证通过后再切换”的方式，避免直接覆盖唯一线上版本。

备份至少要回答：备份什么、多久一次、保存在哪里、保留多久、谁能访问、如何恢复。对个人博客而言，文章数据库、上传图片、环境变量说明和部署配置都可能比服务器本身更重要。

## 一张最小排错顺序表

| 现象 | 先检查 | 再检查 |
| --- | --- | --- |
| 域名完全超时 | DNS、云安全组、服务器监听端口 | UFW、网络线路 |
| 默认 Nginx 页面 | `server_name` 和启用配置 | 配置加载顺序 |
| HTTPS 失败 | 证书域名、443 端口 | Certbot 日志、代理层 |
| `502` | 应用进程和监听端口 | `proxy_pass`、应用日志 |
| 首页白屏 | JS/CSS 请求和构建路径 | 前端环境变量、大小写 |
| 接口 `401/403` | Cookie、Authorization、跨域 | 代理请求头、认证配置 |
| 上传失败 | 存储权限和大小限制 | MIME、路径、备份 |
| 重启后离线 | systemd/Docker 自启动 | 资源限制、恢复演练 |

## 一个 30 分钟的快速定位流程

当你还不知道问题属于哪一层时，可以按下面的顺序执行。每一步只回答一个问题，避免同时修改多个配置。

### 第 1 步：确认域名和服务器地址

```bash
dig +short example.com A
curl -I --max-time 10 http://服务器IP
```

如果域名 IP 不对，停止在服务器上改配置；如果直接访问 IP 也超时，再看安全组、防火墙和监听端口。

### 第 2 步：确认网关能否独立工作

```bash
sudo nginx -t
curl -I -H 'Host: example.com' http://127.0.0.1
```

如果本机请求都失败，问题在 Nginx 配置、站点目录或权限；如果本机成功而公网失败，回到端口和网络边界检查。

### 第 3 步：确认应用进程和接口

```bash
sudo systemctl status vibe-app --no-pager
curl -i http://127.0.0.1:3000/health
```

如果 `health` 接口不存在，应替换成项目真实的公开健康检查路径。`Connection refused` 通常表示没有进程监听；响应 `500` 则说明请求已经到达应用，要继续看应用日志和依赖服务。

### 第 4 步：确认 HTTPS 和浏览器行为

```bash
curl -Iv https://example.com
```

命令可以显示 TLS 握手、证书域名和最终 HTTP 状态。若命令成功但浏览器白屏，打开 Network 和 Console，逐个查看 JavaScript、CSS、接口和字体请求，不要只看首页状态码。

### 第 5 步：只修复一个变量，再重复验证

一次只改变 DNS、Nginx、应用环境变量或数据库权限中的一类配置。修改后重复同一条检查命令，并记录前后结果。这样即使问题没有解决，也能知道哪一个变化影响了行为。

## 数据问题要单独处理

网站能打开，不代表数据安全。发现数据库连接失败、文章消失或上传图片打不开时，先暂停写入和迁移操作，确认当前数据库、对象存储和备份的实际位置。

最小的恢复演练至少包括：

1. 找到一份明确标注时间和来源的备份。
2. 在隔离环境恢复，不直接覆盖线上数据库。
3. 检查表数量、关键记录、图片对象和应用连接权限。
4. 记录恢复耗时、缺失内容和需要人工处理的步骤。
5. 确认恢复结果后，再决定是否切换线上连接。

数据库备份和上传文件备份必须分别验证。只备份代码，无法恢复用户文章；只备份数据库，也可能找不回对象存储中的图片。环境变量说明、域名配置、证书续期方式和部署脚本同样应该有非秘密版本的记录。

## 什么时候应该停止自己排查

以下情况不适合继续让 AI 反复尝试：生产数据库出现疑似误删、支付或认证密钥可能泄露、服务器出现异常登录、证书与域名归属不明、恢复演练失败，或者问题涉及大量用户数据。此时先保留日志和时间线，限制进一步变更，轮换可能泄露的凭据，并由有权限的运维或安全人员处理。

AI 可以帮忙归纳日志、解释状态码和列出只读检查，但不能替你判断法律责任、数据通知义务或事故影响范围。排错过程中最重要的保护动作，有时是暂停修改，而不是继续执行下一条命令。

## 结论

网站故障排查应该从网络边界逐层进入应用，而不是先改代码。DNS 决定请求去哪里，安全组和防火墙决定请求能否进来，Nginx 决定请求交给谁，应用和外部服务决定业务能否完成。按照这个顺序收集证据，通常能更快缩小范围。

这份清单适合第一次上线的个人产品和小型站点，但不能替代完整的生产监控与安全审计。涉及支付、医疗、个人敏感信息或高可用要求时，还应增加访问审计、密钥轮换、灾备、告警和恢复时间目标等专项验证。
