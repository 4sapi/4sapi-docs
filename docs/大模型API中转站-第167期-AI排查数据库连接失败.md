---
title: "【大模型API中转站】第167期 AI排查数据库连接失败 | URL权限网络"
category: 人工智能
tags:
  - 大模型API中转站
  - 4SAPI
  - AI排错
  - 数据库
  - PostgreSQL
  - MySQL
  - Claude Fable 5
description: "数据库连接失败不能只看一句 connection refused。本文拆解数据库 URL、主机名、端口、账号权限、SSL、容器网络、连接池、迁移和防火墙，并给出适合交给 Fable 5 的 AI 排错包。"
---

# 【大模型API中转站】第167期 AI排查数据库连接失败 | URL权限网络

本文是【大模型API中转站】系列的第167篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

数据库连接失败，是后端部署里最常见的坑之一。

报错看起来都差不多：

```text
connection refused
connection timed out
password authentication failed
database does not exist
too many connections
SSL is required
```

但根因完全不同。

这类问题特别适合用 AI。

因为它需要同时看：

```text
连接串。
数据库服务状态。
网络。
账号权限。
容器网络。
SSL。
连接池。
迁移状态。
应用配置。
```

但也最容易误用 AI。

因为很多人会直接把真实数据库连接串贴给模型。

不要这样。

## 1. 先脱敏连接串

真实连接串长这样：

```text
postgresql://app_user:真实密码@db.example.com:5432/app_prod?sslmode=require
```

给 AI 时改成：

```text
postgresql://app_user:[PASSWORD]@db.example.com:5432/app_prod?sslmode=require
```

或者更安全：

```text
协议：postgresql
用户：app_user
密码：已配置，不展示
主机：db.example.com
端口：5432
数据库：app_prod
SSL：require
```

AI 不需要知道真实密码。

它只需要知道：

```text
字段是否存在。
主机名是什么类型。
端口是什么。
SSL 参数是什么。
应用和数据库是否在同一网络。
```

这条一定要写进团队规范：

```text
排查数据库问题时，不把真实连接串发给模型。
```

## 2. 数据库连接失败先按错误文本分类

常见分类：

| 错误 | 优先方向 |
| --- | --- |
| `connection refused` | 主机可达但端口没人监听，或容器服务没起 |
| `connection timed out` | 网络不通、防火墙、安全组、DNS |
| `password authentication failed` | 用户名或密码错 |
| `database does not exist` | 数据库名错或未初始化 |
| `role does not exist` | 用户未创建 |
| `too many connections` | 连接池过大或连接泄漏 |
| `SSL required` | SSL 参数缺失 |
| `no pg_hba.conf entry` | Postgres 访问规则 |
| `Access denied for user` | MySQL 用户、密码、来源主机 |

辅助模型可以先根据错误文本分类。

Fable 5 再综合环境判断。

## 3. 给 AI 的数据库排错包

模板：

```text
【现象】
- 哪个服务连接数据库失败：
- 错误原文：
- 是否所有接口失败：
- 是否刚部署后出现：

【数据库】
- 类型：Postgres / MySQL / SQLite / MongoDB / Redis
- 部署：Docker / 云数据库 / 本机服务
- 主机：脱敏描述
- 端口：
- 数据库名：
- 用户名：
- SSL 要求：

【应用】
- 部署方式：
- 是否 Docker Compose：
- DATABASE_URL 结构，密码脱敏：
- ORM：Prisma / Sequelize / TypeORM / SQLAlchemy / Django ORM

【证据】
- app 日志：
- db 日志：
- docker compose ps：
- 网络连通性检查：

【最近变更】
- 改过密码 / 用户 / 数据库名 / 端口 / SSL / compose / 迁移：

【边界】
- 不输出真实密码
- 不删除 volume
- 不重建数据库
- 不直接跑生产迁移
- 先给只读验证步骤
```

这份包足够让 AI 分层。

## 4. 容器里连接数据库：localhost 是大坑

Docker Compose 里最常见的错误之一：

```text
app 容器里 DATABASE_URL 写 localhost。
```

在容器里：

```text
localhost 指的是 app 容器自己。
```

不是 db 容器。

如果 compose 里服务名叫：

```yaml
services:
  db:
```

应用应该连：

```text
db:5432
```

而不是：

```text
localhost:5432
```

给 AI 的证据：

```text
docker-compose.yml 服务名。
DATABASE_URL 脱敏版。
app 日志 connection refused localhost:5432。
```

Prompt：

```text
请判断这个数据库连接失败是否由 Docker 容器内 localhost 使用错误导致。
如果是，请说明容器内 localhost 指向谁，以及应该如何使用 compose service name。
```

这个问题低成本模型也能判断。

但如果还涉及 Nginx、云数据库、内网访问，就交给 Fable 5。

## 5. 云数据库：超时多半是网络

如果错误是：

```text
connection timed out
```

优先看：

```text
安全组。
白名单。
VPC。
公网访问开关。
数据库端口。
DNS。
SSL。
```

只读验证：

```bash
nc -vz db.example.com 5432
```

或：

```bash
telnet db.example.com 5432
```

注意，有些服务器没有 `nc` 或 `telnet`。

可以让 AI 根据系统给替代命令。

但不要让 AI 直接开放数据库公网端口。

更稳的策略是：

```text
应用服务器和数据库走内网。
公网数据库只对固定 IP 白名单开放。
不暴露 Redis。
不暴露本地 Postgres/MySQL 给全网。
```

## 6. 密码错误：先确认用户和来源

Postgres：

```text
password authentication failed for user "app"
```

MySQL：

```text
Access denied for user 'app'@'host'
```

这时不要马上重置密码。

先确认：

```text
应用实际读到的是哪个 DATABASE_URL。
连接的是哪个数据库实例。
用户是否存在。
密码是否刚轮换。
MySQL 用户是否限制来源 host。
```

给 AI：

```text
用户名。
数据库名。
host 脱敏。
错误原文。
最近是否改过密码。
```

不要给真实密码。

AI 可以给：

```text
如何验证当前环境变量是否加载。
如何确认应用连到哪个 host。
如何用只读方式测试连接。
```

但密码本身由你在安全环境里处理。

## 7. 数据库不存在或未初始化

错误：

```text
database "app_prod" does not exist
relation "users" does not exist
no such table
```

这里要区分：

```text
数据库没创建。
表没创建。
迁移没执行。
连接到了空库。
连接到了错误环境。
```

AI 排查路径：

```text
1. 当前 DATABASE_URL 指向哪个库。
2. 数据库是否存在。
3. migration 表是否存在。
4. 最新 migration 是否执行。
5. 代码版本是否依赖新表。
```

重要提醒：

```text
生产迁移必须人工确认。
```

让 AI 生成迁移计划可以。

不要让 AI 直接执行。

## 8. too many connections

错误：

```text
too many connections
remaining connection slots are reserved
```

常见原因：

```text
应用实例太多。
每个实例连接池太大。
serverless 场景连接暴涨。
连接泄漏。
后台任务没有关闭连接。
数据库 max_connections 太低。
```

给 AI 的材料：

```text
应用实例数。
连接池配置。
数据库 max_connections。
当前连接数。
最近流量变化。
```

Postgres 只读查询示例：

```sql
select state, count(*) from pg_stat_activity group by state;
```

MySQL：

```sql
show processlist;
```

这类问题适合 Fable 5。

因为修复不只是“调大连接数”。

可能要：

```text
调小连接池。
加连接代理。
修连接泄漏。
限制 worker 并发。
拆读写。
```

## 9. SSL 参数问题

云数据库常要求 SSL。

错误可能是：

```text
SSL is required
no pg_hba.conf entry ... SSL off
self signed certificate
```

连接串里可能需要：

```text
sslmode=require
```

或 ORM 特定配置。

不要让 AI 随便建议：

```text
关闭 SSL。
```

更稳的 Prompt：

```text
请根据这个数据库连接错误判断是否为 SSL 配置问题。
优先给启用 SSL 的正确配置。
不要建议关闭 SSL，除非明确说明只适用于本地开发。
```

## 10. 4SAPI 怎么参与数据库排错

数据库排错本身不一定调用模型 API。

但如果你的应用通过模型 API 做业务，数据库错误可能和模型调用链路串在一起。

例如：

```text
用户请求 -> 后端调用 4SAPI -> 写入结果到数据库 -> 写入失败 -> 返回 500
```

这时你要串联：

```text
业务 request_id。
4SAPI task_id。
数据库 transaction_id 或日志时间。
```

建议日志里记录：

```text
request_id
user_id 脱敏
model_task_id
db_operation
db_error_code
```

这样 Fable 5 才能判断：

```text
模型调用失败导致没写库。
还是模型调用成功但数据库写入失败。
还是数据库失败导致模型任务重复提交。
```

## 11. 一条完整 Prompt

```text
你是数据库连接失败排查助手。

请根据我提供的连接串脱敏结构、应用日志、数据库日志、部署方式和最近变更，判断问题属于哪一类：
1. 主机名或端口错误
2. Docker 容器网络问题
3. 数据库服务未启动
4. 用户名或密码错误
5. 数据库不存在或迁移未执行
6. 防火墙/安全组/VPC
7. SSL 配置
8. 连接池或连接泄漏

要求：
- 不要求我提供真实密码。
- 每个判断引用证据。
- 先给只读验证步骤。
- 不要建议删除 volume、重建数据库、清空表或直接跑生产迁移。
- 涉及生产变更时标注“需要人工确认”。
```

## 12. 总结

数据库连接失败，最怕两个极端。

一个是：

```text
看到 connection refused 就重启服务器。
```

另一个是：

```text
把真实 DATABASE_URL 贴给 AI。
```

正确方式是：

```text
脱敏连接串。
整理 app/db 日志。
说明部署方式。
标注最近变更。
让辅助模型分类。
让 Fable 5 做根因判断。
高风险动作人工确认。
```

4SAPI 在这里的作用，是把排错过程里的模型调用管起来：

```text
谁调用了模型。
用了哪个模型。
花了多少钱。
是否涉及生产故障。
是否输出了敏感信息。
最终是否解决。
```

一句话：

```text
数据库问题可以让 AI 帮你看，但不要把数据库钥匙交给 AI。
```

下一批可以继续写：

```text
Nginx 502/504。
CORS。
SSL 证书。
CI/CD 失败。
依赖安装失败。
```

这些都是 AI 排错特别值得写的场景。
