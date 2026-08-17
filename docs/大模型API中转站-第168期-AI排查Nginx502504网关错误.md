---
title: "【大模型API中转站】第168期 AI排查Nginx 502/504 | 网关到上游"
category: 人工智能
tags:
  - 大模型API中转站
  - 4SAPI
  - AI排错
  - Nginx
  - 502
  - 504
  - Claude Fable 5
description: "Nginx 502/504 不是一个错误，而是网关和上游之间的连接问题。本文讲如何整理 Nginx error.log、upstream、端口、容器状态、超时配置和后端日志，让辅助模型清洗证据、Fable 5 判断根因。"
---

# 【大模型API中转站】第168期 AI排查Nginx 502/504 | 网关到上游

本文是【大模型API中转站】系列的第168篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

网站上线后，最常见的网关报错是：

```text
502 Bad Gateway
504 Gateway Timeout
```

这两个错误很适合用 AI 排查。

因为它们不是一个点的问题。

它们卡在：

```text
用户 -> CDN/负载均衡 -> Nginx -> 上游服务 -> 数据库/外部 API
```

其中任何一层慢、挂、端口错、协议错，都可能表现成 502 或 504。

但要注意：

```text
不要只把浏览器截图发给 AI。
```

AI 需要的是网关证据包。

## 1. 先区分 502 和 504

粗略理解：

| 错误 | 常见含义 |
| --- | --- |
| 502 | Nginx 找上游失败，或上游返回异常 |
| 504 | Nginx 等上游太久，超时 |

502 常见原因：

```text
上游服务没启动。
upstream 端口写错。
容器名或 host 写错。
应用监听 127.0.0.1。
Nginx 和 app 不在同一网络。
后端进程崩溃。
```

504 常见原因：

```text
后端接口太慢。
数据库慢查询。
第三方 API 超时。
大模型 API 调用太久。
Nginx proxy_read_timeout 太短。
任务应该异步但写成同步。
```

Fable 5 很适合做这类跨层判断。

但先让低成本模型整理日志即可。

## 2. 给 AI 的 502/504 排错包

模板：

```text
【现象】
- URL：
- 状态码：502 / 504
- 是否所有页面：
- 是否某个接口：
- 开始时间：

【Nginx】
- nginx -t 结果：
- server block 脱敏配置：
- upstream 配置：
- error.log 最近 100 行：
- access.log 相关请求：

【上游】
- app 监听端口：
- docker compose ps：
- app logs 最近 200 行：
- curl 127.0.0.1:端口 结果：

【最近变更】
- 改过 Nginx / compose / 端口 / 域名 / SSL / 后端代码：

【边界】
- 不修改生产配置
- 不重启数据库
- 先给只读验证步骤
```

这个包比“502 怎么办”有效太多。

## 3. Nginx error.log 是第一证据

常见 error.log：

```text
connect() failed (111: Connection refused) while connecting to upstream
```

意思通常是：

```text
Nginx 能找到 host，但端口没人监听。
```

再比如：

```text
upstream timed out while reading response header from upstream
```

通常是：

```text
上游接了请求，但迟迟没有返回响应头。
```

还有：

```text
host not found in upstream
```

通常是：

```text
upstream 主机名解析失败。
```

给 AI 时，直接贴这些行。

不要只贴浏览器页面。

## 4. 只读验证命令

先检查配置：

```bash
sudo nginx -t
```

看 Nginx 错误：

```bash
sudo tail -n 100 /var/log/nginx/error.log
```

看上游端口：

```bash
ss -tulpn
```

本机 curl：

```bash
curl -I http://127.0.0.1:3000
```

如果是 Docker：

```bash
docker compose ps
docker compose logs --tail=200 app
```

如果 Nginx 在容器里，还要确认它能访问 app 服务名：

```bash
docker compose exec nginx curl -I http://app:3000
```

这类命令适合让 AI 解释输出。

但执行前你要确认服务名和端口。

## 5. 常见根因一：端口写错

Nginx 配置：

```nginx
proxy_pass http://127.0.0.1:3000;
```

但应用实际监听：

```text
3001
```

或者 Docker 里 app 服务监听 3000，但宿主机没有映射。

这类错误很常见。

给 AI：

```text
nginx upstream。
docker compose ports。
app 启动日志。
ss -tulpn 输出。
```

让它判断：

```text
Nginx 访问的是宿主机端口，还是 Docker 网络里的服务名。
```

## 6. 常见根因二：Docker 网络不通

如果 Nginx 在宿主机，app 在 Docker，通常用：

```text
127.0.0.1:映射端口
```

如果 Nginx 也在 Docker Compose 里，更推荐：

```text
http://app:3000
```

很多 502 来自混用：

```text
容器里的 Nginx 访问 127.0.0.1:3000。
```

但容器里的 127.0.0.1 是 Nginx 容器自己。

不是 app 容器。

这个和第167期数据库 localhost 坑是同一类问题。

## 7. 常见根因三：上游太慢

504 不一定是 Nginx 错。

可能是后端接口真的慢。

要查：

```text
后端日志耗时。
数据库慢查询。
第三方接口耗时。
模型 API 调用耗时。
Nginx timeout 配置。
```

如果接口内部调用 4SAPI 或其他模型 API，建议日志里记录：

```text
request_id
model_task_id
model
duration_ms
status_code
```

这样 Fable 5 才能判断：

```text
是模型接口慢。
还是数据库慢。
还是后端没有异步化。
```

## 8. AI Prompt

```text
你是 Nginx 502/504 排查助手。

请根据 Nginx error.log、access.log、upstream 配置、app 状态和最近变更，判断问题属于：
1. 上游服务未启动
2. 端口或 host 配置错误
3. Docker 网络错误
4. 上游接口超时
5. 数据库或第三方 API 慢
6. Nginx timeout 配置不合适

要求：
- 每个判断引用证据。
- 先给只读验证命令。
- 不要建议直接重启整机或清空缓存。
- 修改 Nginx、DNS、证书或生产配置前标注人工确认。
```

## 9. 4SAPI 分工

| 阶段 | 模型 |
| --- | --- |
| Nginx 日志摘取 | 低成本模型 |
| 跨层根因判断 | Fable 5 |
| timeout 策略审查 | Fable 5 |
| 故障报告 | 低成本模型 |

4SAPI 记录：

```text
error_type: nginx_502_504
stage
model
evidence_ids
root_cause
fix_confirmed
```

这能把网关问题沉淀成团队排障资产。

### 实战补充：502 排查顺序

如果线上突然 502，可以按这个顺序给 AI：

```text
第一步：nginx -t 是否通过。
第二步：error.log 是否出现 connection refused。
第三步：upstream host 和端口是什么。
第四步：宿主机 curl upstream 是否通。
第五步：如果 Docker 部署，容器内 curl 服务名是否通。
第六步：app 日志是否在同一时间崩溃。
```

让 Fable 5 判断时，要求它输出：

```text
最可能层级。
证据。
最小只读验证。
是否需要修改配置。
是否阻塞上线。
```

这样排查不会在 Nginx、Docker、后端之间来回乱跳。

### 落地清单：Nginx 代理配置审查

上线前可以让 AI 审查：

```text
proxy_pass 是否指向正确上游。
是否保留 Host 和 X-Forwarded-For。
是否配置合理 timeout。
是否有 WebSocket 或 SSE 需要特殊 Header。
是否区分静态资源和 API。
是否记录 request_id。
```

如果你的后端会调用 4SAPI，建议把 `request_id` 传到后端日志和模型调用日志里。

这样 504 时能串起来看：

```text
Nginx 请求。
后端请求。
模型调用。
数据库写入。
```

## 10. 总结

502/504 不要只看浏览器。

要看：

```text
Nginx error.log。
upstream 配置。
上游端口。
容器状态。
后端日志。
慢查询或第三方调用。
```

AI 的价值不是替你乱改 Nginx。

而是帮你把网关到上游的证据链串起来。

一句话：

```text
502 看连接，504 看耗时。
```

下一篇写 CORS。

因为很多“接口明明返回了，前端就是拿不到”的问题，都卡在跨域上。
