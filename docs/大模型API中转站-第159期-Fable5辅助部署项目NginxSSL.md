---
title: "【大模型API中转站】第159期 Fable 5部署项目 | Nginx域名SSL"
category: 人工智能
tags:
  - 大模型API中转站
  - 4SAPI
  - Claude Fable 5
  - 项目部署
  - Nginx
  - SSL
  - 环境变量
  - PM2
description: "用 4SAPI 接入 Fable 5 辅助部署项目：从构建命令、环境变量、Nginx 反代、域名解析、SSL 证书、PM2 或 Docker Compose 进程管理，到上线验证和回滚方案，一篇讲清项目从本地到服务器的关键路径。"
---

# 【大模型API中转站】第159期 Fable 5部署项目 | Nginx域名SSL

本文是【大模型API中转站】系列的第159篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

前面三篇讲了服务器、Docker、数据库。

这一篇开始部署项目本体。

很多项目本地能跑，上服务器就挂。

常见原因不是代码坏了，而是：

```text
构建命令不对。
启动命令不对。
环境变量缺失。
端口没开。
Nginx 反代错。
域名没解析。
SSL 证书没配。
数据库连接串在服务器环境不对。
```

Fable 5 适合做部署总控。

你把项目结构、报错日志、Nginx 配置、Compose 文件脱敏后给它。

它来帮你找出问题链路。

4SAPI 则负责把高级模型调用和低成本日志处理分开。

## 1. 先让模型读项目，不要直接部署

Prompt：

```text
我准备把这个项目部署到 Linux 服务器。
请先读取项目结构并生成部署计划，不要给我执行命令。

输出：
1. 技术栈判断
2. 构建命令
3. 启动命令
4. 需要的环境变量
5. 需要的数据库或 Redis
6. 推荐部署方式：Docker Compose、PM2、systemd 或静态站
7. Nginx 和域名配置建议
8. 上线前风险清单

约束：
不要读取真实密钥。
不要执行部署命令。
涉及 DNS、SSL、生产配置修改时标注为人工确认。
```

第一步只做计划。

这能避免模型一上来就给你一堆命令，结果和项目实际结构不匹配。

## 2. 构建命令和启动命令

不同项目差异很大。

常见判断：

| 项目 | 构建 | 启动 |
| --- | --- | --- |
| Vite 静态站 | `npm run build` | Nginx 托管 dist |
| Next.js SSR | `npm run build` | `npm run start` |
| Node API | 无或 build | `node dist/index.js` |
| Python FastAPI | 无 | `uvicorn app:app` |
| Django | collectstatic/migrate | gunicorn/uwsgi |

让 Fable 5 不要猜。

让它从文件里找证据：

```text
请根据 package.json、pyproject.toml、requirements.txt、Dockerfile、README 判断构建和启动命令。
每个结论都标注证据来源。
```

如果没有证据，就写“待验证”。

不要让它编。

## 3. 环境变量：先列清单，再填值

部署最常见的问题是 `.env`。

本地有，服务器没有。

或者本地是：

```text
localhost
```

服务器里应该是：

```text
db
```

因为 Docker Compose 内部连接要用服务名。

让 Fable 5 生成环境变量清单：

```text
请扫描项目里读取环境变量的位置，生成环境变量清单。

输出列：
变量名、用途、是否必填、示例值、是否敏感、生产配置建议。

约束：
不要要求我提供真实值。
不要输出真实 Key。
```

示例表：

| 变量 | 用途 | 是否敏感 | 示例 |
| --- | --- | --- | --- |
| `DATABASE_URL` | 数据库连接 | 是 | `postgres://app:***@db:5432/app` |
| `REDIS_URL` | Redis 连接 | 是 | `redis://:***@redis:6379/0` |
| `MODEL_BASE_URL` | 4SAPI 地址 | 否 | `https://4sapi.com/v1` |
| `MODEL_API_KEY` | 4SAPI Key | 是 | 不展示 |
| `MODEL_NAME` | 模型名称 | 否 | 按模型广场复制 |

## 4. 部署方式：PM2 还是 Docker Compose

简单选择：

| 方式 | 适合 |
| --- | --- |
| 静态站 + Nginx | 纯前端、文档站、博客 |
| PM2 | 单个 Node 服务，部署简单 |
| systemd | Python/Go/Rust 服务，想用系统服务管理 |
| Docker Compose | app + db + redis，多服务编排 |

如果项目还小，PM2 很快。

如果涉及数据库、Redis、worker，Compose 更稳。

让 Fable 5 给出理由：

```text
请比较 PM2 和 Docker Compose 部署这个项目的优缺点。
我的优先级是：稳定、可回滚、低维护成本。
```

不要盲目追求复杂。

能简单跑稳，就先简单。

## 5. Nginx 反向代理

如果应用监听本机 3000，Nginx 对外提供 80/443。

典型配置：

```nginx
server {
    listen 80;
    server_name example.com;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

让 Fable 5 审查 Nginx：

```text
请审查这个 Nginx 配置。
重点看 server_name、proxy_pass、真实 IP 头、WebSocket 支持、上传大小限制、
HTTP 到 HTTPS 跳转和安全风险。
```

如果有 WebSocket，要补：

```nginx
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection "upgrade";
```

不要让模型默认加一堆复杂配置。

先满足项目真实需要。

## 6. 域名解析和 SSL

域名解析通常是：

```text
A 记录：example.com -> 服务器公网 IP
CNAME：www.example.com -> example.com
```

DNS 修改属于高风险动作。

让 Fable 5 只给计划，不要自动改。

SSL 常见用 Certbot：

```bash
sudo certbot --nginx -d example.com -d www.example.com
```

但执行前要确认：

```text
域名已经解析到服务器。
80 端口开放。
Nginx 配置能通过测试。
没有占用冲突。
```

检查：

```bash
sudo nginx -t
curl -I http://example.com
```

让模型判断输出。

不要让它在 DNS 未生效时反复申请证书。

## 7. 进程管理和自启动

PM2 常用：

```bash
pm2 start npm --name app -- run start
pm2 save
pm2 startup
pm2 logs app
```

Docker Compose 常用：

```bash
docker compose up -d --build
docker compose ps
docker compose logs -f app
```

systemd 则需要 unit 文件。

让 Fable 5 根据项目选：

```text
请根据我的部署方式，给出进程管理方案。
要求：
1. 服务重启后能自动启动
2. 能查看日志
3. 能平滑重启
4. 能回滚
5. 不要覆盖现有服务
```

## 8. 上线验证

上线不是服务能启动就完。

至少验证：

| 检查 | 命令或方式 |
| --- | --- |
| 进程 | `pm2 status` / `docker compose ps` |
| 端口 | `ss -tulpn` |
| 本机访问 | `curl http://127.0.0.1:3000` |
| Nginx | `sudo nginx -t` |
| 域名 | `curl -I https://example.com` |
| 数据库 | 登录、注册、写入测试数据 |
| 日志 | app、Nginx、数据库日志无异常 |
| SSL | 浏览器和证书到期时间 |

让 Fable 5 做验收：

```text
下面是我的上线验证输出。
请判断是否可以认为部署成功。
只输出：通过项、阻塞项、建议观察项、下一步。
```

## 9. 回滚方案

部署前必须知道怎么回滚。

最简单的回滚方案：

```text
保留上一版代码目录或镜像 tag。
保留上一版 .env。
数据库迁移前备份。
Nginx 配置修改前备份。
记录本次部署命令。
```

让 Fable 5 写：

```text
请为这次部署生成回滚方案。
要求：
1. 哪些文件要备份
2. 哪些命令能恢复上一版
3. 哪些动作不可逆
4. 回滚后如何验证
5. 数据库迁移如何处理
```

如果没有回滚方案，就不要在生产上做大改动。

## 10. 4SAPI 路由建议

项目部署阶段可以这样分工：

| 任务 | 模型 |
| --- | --- |
| 部署计划 | Fable 5 |
| Nginx/SSL 审查 | Fable 5 |
| 日志摘要 | 低成本模型 |
| 命令解释 | 中低成本模型 |
| 上线验收 | Fable 5 |
| 回滚方案 | Fable 5 |

4SAPI 记录：

```text
project: deploy
stage: app
risk: config_change / read_only / rollback
```

这样你能知道，部署项目到底花了多少模型成本，也能复盘哪些环节最容易卡。

## 11. 总结

项目部署不是复制命令。

它是一条链：

```text
读项目。
定部署方式。
配环境变量。
跑进程。
接 Nginx。
配域名和 SSL。
验收。
回滚。
```

Fable 5 适合做部署总控。

4SAPI 适合把这类高级模型调用变成可审计、可控费、可回放的部署工作流。

下一篇收尾：

```text
部署验收、日志监控、备份、回滚和成本治理。
```
