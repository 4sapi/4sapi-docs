---
title: "【大模型API中转站】第158期 Fable 5装数据库 | Postgres到Redis"
category: 人工智能
tags:
  - 大模型API中转站
  - 4SAPI
  - Claude Fable 5
  - 数据库
  - PostgreSQL
  - MySQL
  - Redis
  - 备份恢复
description: "用 4SAPI 调 Fable 5 辅助安装数据库：讲清 Postgres、MySQL、Redis 怎么选，Docker 和原生安装怎么取舍，账号权限、连接串、端口暴露、备份恢复和上线前检查怎么做。"
---

# 【大模型API中转站】第158期 Fable 5装数据库 | Postgres到Redis

本文是【大模型API中转站】系列的第158篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一篇讲 Docker。

这一篇讲数据库。

数据库是部署里最容易出事故的部分。

因为应用挂了可以重启。

镜像坏了可以重建。

但数据库数据一旦丢了，就不是“重新部署一下”的问题。

所以用 Fable 5 辅助装数据库时，第一原则是：

```text
先规划，再安装。
先备份，再升级。
先本机访问，再考虑远程。
```

4SAPI 在这里的作用，是把 Fable 5 放到高风险判断上：

```text
数据库选型。
权限设计。
备份恢复。
安全审查。
连接失败排查。
```

而日志压缩、命令解释、SQL 报错摘要，可以交给低成本模型。

## 1. 先让 Fable 5 判断你需要什么数据库

不要一上来就问：

```text
怎么安装 MySQL？
```

先问：

```text
我的项目到底需要什么数据库？
```

Prompt：

```text
请根据我的项目结构和业务描述，判断数据库选型。

输出：
1. 是否需要关系型数据库
2. Postgres、MySQL、SQLite 哪个更适合
3. 是否需要 Redis
4. 是否适合先用托管数据库
5. 如果自建，最小部署方案是什么
6. 数据备份和恢复风险

约束：
不要默认开放公网端口。
不要让我把真实数据库密码发给你。
涉及生产数据迁移时，必须标注为人工确认。
```

一般判断：

| 场景 | 建议 |
| --- | --- |
| 个人小工具、低并发 | SQLite 可先跑 |
| SaaS、后台系统、复杂查询 | Postgres |
| 传统 Web、团队熟悉 MySQL | MySQL |
| 缓存、队列、限流、会话 | Redis |
| 数据价值高、团队不懂运维 | 托管数据库优先 |

不要为了“专业”强行自建。

托管数据库虽然贵一点，但能省很多备份和运维事故。

## 2. Docker 数据库还是原生安装

两种都能用。

| 方式 | 优点 | 风险 |
| --- | --- | --- |
| Docker | 配置统一、迁移方便、适合 Compose | volume 管不好会丢数据 |
| 原生安装 | 更贴近传统运维、系统服务清楚 | 升级和迁移更复杂 |
| 托管数据库 | 备份、高可用、省心 | 成本高、厂商绑定 |

如果你项目已经 Docker Compose，数据库也放 Compose 里最简单。

但要注意：

```text
数据库 volume 必须明确。
不要随便 docker compose down -v。
不要随便 docker volume prune。
```

让 Fable 5 审查：

```text
请检查这个数据库部署方案。
重点看数据持久化、端口暴露、密码管理、备份恢复和升级风险。
```

## 3. Postgres 最小 Compose 示例

```yaml
services:
  db:
    image: postgres:16
    restart: unless-stopped
    environment:
      POSTGRES_DB: app
      POSTGRES_USER: app
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U app -d app"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

这里故意没有写：

```yaml
ports:
  - "5432:5432"
```

因为应用如果在同一个 Compose 网络里，可以直接用服务名连接：

```text
postgres://app:密码@db:5432/app
```

不需要把数据库暴露给公网。

如果你为了本地调试要连，可以用 SSH 隧道。

不要直接全网开放 5432。

## 4. MySQL 最小 Compose 示例

```yaml
services:
  mysql:
    image: mysql:8.4
    restart: unless-stopped
    environment:
      MYSQL_DATABASE: app
      MYSQL_USER: app
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql
    command:
      - --character-set-server=utf8mb4
      - --collation-server=utf8mb4_unicode_ci

volumes:
  mysql_data:
```

MySQL 要注意字符集。

很多中文问题、emoji 问题，最后都来自字符集没设好。

让 Fable 5 审查时，可以问：

```text
请检查这个 MySQL Compose 是否适合中文内容和 emoji。
是否存在 root 密码泄露、端口暴露、volume 不持久的问题。
```

## 5. Redis 不要公网裸奔

Redis 是最容易被误开的服务之一。

很多人本地开发时习惯：

```text
redis://localhost:6379
```

到了服务器，如果直接开放 6379，风险很大。

Redis 最小 Compose：

```yaml
services:
  redis:
    image: redis:7
    restart: unless-stopped
    command: ["redis-server", "--appendonly", "yes", "--requirepass", "${REDIS_PASSWORD}"]
    volumes:
      - redis_data:/data

volumes:
  redis_data:
```

注意：

```text
不暴露 ports。
设置密码。
持久化按业务需要开启。
不要把 Redis 当主数据库。
```

如果 Redis 只是缓存，可以接受丢。

如果 Redis 存队列或会话，就要考虑持久化和恢复。

## 6. 账号权限：应用不要用 root

数据库安装后，应用不要用 root 或超级用户连接。

要给应用单独账号：

```text
app_user。
只访问 app_db。
只给必要权限。
```

Fable 5 可以帮你生成 SQL，但要先让它解释权限。

Prompt：

```text
请为我的应用设计数据库账号权限。
要求：
1. 应用账号不能是 root/superuser
2. 只允许访问自己的数据库
3. 说明迁移工具是否需要额外权限
4. 给出创建账号的 SQL
5. 标注哪些 SQL 不应在生产随意执行
```

生产数据库权限宁可麻烦一点。

不要为了省事全给。

## 7. 连接串管理

连接串不要写进代码。

推荐放 `.env`：

```text
DATABASE_URL=postgres://app:***@db:5432/app
REDIS_URL=redis://:***@redis:6379/0
```

给模型排错时，脱敏：

```text
DATABASE_URL=postgres://app:***@db:5432/app
```

不要贴真实密码。

也不要截图包含完整连接串的后台页面。

可以让 Fable 5 检查连接串结构：

```text
请检查这个脱敏后的 DATABASE_URL 格式是否适合 Docker Compose 内部连接。
不要要求我提供真实密码。
```

## 8. 备份和恢复：只会备份不够

很多人有备份，但没试过恢复。

这等于没有备份。

Postgres 备份示例：

```bash
docker compose exec db pg_dump -U app app > backup.sql
```

恢复要在测试环境演练。

MySQL 备份示例：

```bash
docker compose exec mysql mysqldump -u app -p app > backup.sql
```

让 Fable 5 帮你写备份方案：

```text
请根据我的数据库部署方式，写一份备份和恢复演练方案。

要求：
1. 区分日常备份和升级前备份
2. 给出恢复到测试库的步骤
3. 标注哪些命令会覆盖数据
4. 不要包含真实密码
5. 给出备份文件保存和清理策略
```

注意：

```text
恢复命令比备份命令更危险。
```

恢复生产前必须人工确认。

## 9. 数据库安装验收清单

| 检查项 | 标准 |
| --- | --- |
| 端口 | 数据库不直接暴露公网 |
| 账号 | 应用不用 root/superuser |
| 密码 | 放环境变量，不写代码 |
| volume | 数据持久化位置明确 |
| 连接 | 应用能通过服务名连接 |
| 备份 | 有备份命令 |
| 恢复 | 在测试环境演练过 |
| 日志 | 知道如何查看数据库日志 |
| 权限 | 迁移和运行权限分开 |

## 10. 4SAPI 路由建议

数据库任务的模型分工：

| 任务 | 推荐模型 |
| --- | --- |
| 数据库选型 | Fable 5 |
| 权限设计 | Fable 5 |
| 备份恢复方案 | Fable 5 |
| 日志摘要 | 低成本模型 |
| SQL 报错解释 | 中等模型 |
| 生产恢复前审查 | Fable 5 + 人工 |

在 4SAPI 日志里记录：

```text
project: deploy
stage: database
risk: data_loss / read_only / planning
```

如果某次请求涉及恢复、删除、清空、迁移生产数据，应该标成高风险。

高风险任务不应该自动执行。

## 11. 总结

数据库部署不是“装上就行”。

真正要管的是：

```text
选型。
权限。
端口。
连接串。
持久化。
备份。
恢复。
日志。
```

Fable 5 适合帮你做高级判断。

4SAPI 适合帮你管理模型调用和成本。

但生产数据的最终确认，必须在人手里。

下一篇继续：

```text
用 Fable 5 辅助部署项目本体。
Nginx、域名、SSL、环境变量、进程管理怎么一次理清。
```
