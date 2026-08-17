---
title: "【大模型API中转站】第157期 Fable 5装Docker | Compose一次跑通"
category: 人工智能
tags:
  - 大模型API中转站
  - 4SAPI
  - Claude Fable 5
  - Docker
  - Docker Compose
  - 服务器部署
  - 容器日志
description: "用 4SAPI 调 Fable 5 做 Docker 安装和 Docker Compose 部署助理：从安装前检查、镜像源、Compose 文件、volume、端口、日志、自启动到常见报错排查，讲清哪些命令可以让模型解释，哪些动作必须人工确认。"
---

# 【大模型API中转站】第157期 Fable 5装Docker | Compose一次跑通

本文是【大模型API中转站】系列的第157篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一篇讲服务器初始化。

这一篇进入 Docker。

对独立开发者来说，Docker 最大的价值是：

```text
把项目运行环境固定下来。
```

不再是：

```text
我本地 Node 版本是 22。
服务器 Node 是 18。
本地有 pnpm。
服务器没有。
本地数据库能连。
服务器环境变量忘配。
```

Docker 能把这些东西收进镜像和 Compose。

但第一次装 Docker、写 `docker-compose.yml`、排查容器日志，仍然很容易卡。

这时候 Fable 5 这类高级模型适合做三件事：

```text
读项目结构，判断是否适合 Docker。
帮你生成 Compose 草案。
根据日志定位问题。
```

4SAPI 负责把 Fable 5 放到高价值路由里，不要让它浪费在重复解释基础命令上。

## 1. 先让 Fable 5 判断要不要 Docker

不是所有项目都必须 Docker。

简单静态站可能直接 Nginx 就够。

但下面这些场景很适合 Docker Compose：

| 场景 | 为什么适合 |
| --- | --- |
| Node/Python 后端 | 固定运行时版本 |
| 需要数据库 | app + db 可一起编排 |
| 需要 Redis | 服务依赖清楚 |
| 有队列/worker | 多进程管理更方便 |
| 多环境部署 | dev/staging/prod 配置更可控 |

Prompt：

```text
请读取我的项目结构，判断是否适合用 Docker Compose 部署。

输出：
1. 当前项目的运行时和依赖
2. 是否需要 Docker
3. 如果不用 Docker，替代部署方式是什么
4. 如果用 Docker，需要哪些服务
5. 初版 docker-compose.yml 的服务拆分建议

约束：
不要直接生成生产可用配置。
先给部署设计和风险说明。
```

## 2. 安装 Docker 前的只读检查

让模型先生成检查命令，而不是安装命令：

```bash
cat /etc/os-release
uname -a
df -h
free -h
id
groups
which docker
which docker-compose
```

如果系统已经有旧 Docker，要先确认：

```bash
docker --version
docker compose version
docker ps
docker images
docker volume ls
```

这些输出可以贴给 Fable 5。

让它判断：

```text
当前机器是否已经安装 Docker。
是否存在旧容器。
是否有重要 volume。
是否能安全继续。
```

不要让模型直接执行清理命令。

尤其是：

```bash
docker system prune -a
docker volume prune
rm -rf /var/lib/docker
```

这些可能删掉数据。

## 3. 安装 Docker：命令要按系统区分

Ubuntu、Debian、CentOS、Rocky Linux 的安装方式不同。

不要复制一套命令到所有系统。

让 Fable 5 根据 `/etc/os-release` 输出生成对应步骤。

Prompt：

```text
下面是服务器系统信息。
请给出适合这个系统的 Docker Engine 和 Docker Compose 安装步骤。

要求：
1. 每一步解释目的
2. 不要包含删除旧数据的命令
3. 如果需要添加软件源，解释风险
4. 安装后给出验证命令
5. 给出失败排查路径
```

安装完成后验证：

```bash
docker --version
docker compose version
sudo systemctl status docker --no-pager
docker run hello-world
```

如果 `hello-world` 失败，把完整日志给模型。

不要只贴最后一行。

## 4. deploy 用户能不能跑 Docker

Docker 默认需要权限。

很多人会直接 `sudo docker`。

这能用，但长期不舒服。

常见做法是把 deploy 用户加入 docker 组：

```bash
sudo usermod -aG docker deploy
```

但这里有安全含义：

```text
docker 组权限很大，接近 root。
```

所以要让 Fable 5 解释清楚：

```text
为什么加入 docker 组。
风险是什么。
有没有替代方式。
怎么验证生效。
```

验证：

```bash
groups
docker ps
```

注意：加入组后通常需要重新登录 SSH。

不要以为命令执行完马上生效。

## 5. Compose 文件：先写最小版本

不要一上来写很复杂的 Compose。

先跑通最小服务。

例如一个 Node API：

```yaml
services:
  app:
    build: .
    restart: unless-stopped
    env_file:
      - .env
    ports:
      - "3000:3000"
```

如果有 Postgres：

```yaml
services:
  app:
    build: .
    restart: unless-stopped
    env_file:
      - .env
    depends_on:
      - db
    ports:
      - "3000:3000"

  db:
    image: postgres:16
    restart: unless-stopped
    environment:
      POSTGRES_DB: app
      POSTGRES_USER: app
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - db_data:/var/lib/postgresql/data

volumes:
  db_data:
```

注意：

```text
数据库不要默认 ports 暴露到公网。
app 连 db 用服务名 db。
密码放 .env，不写进 compose。
```

## 6. 让 Fable 5 审查 Compose

Prompt：

```text
请审查这个 docker-compose.yml。

重点检查：
1. 是否把数据库端口暴露公网
2. 是否把密码写死
3. volume 是否能持久化数据
4. restart 策略是否合理
5. depends_on 是否足够，是否还需要健康检查
6. 容器日志怎么查看
7. 是否适合生产环境

输出 P0/P1/P2 风险表。
不要直接改文件，先给建议。
```

这类审查适合用 Fable 5。

因为它需要综合判断。

而把日志整理成摘要，可以用 4SAPI 路由到低成本模型。

## 7. 启动和日志

常用命令：

```bash
docker compose config
docker compose up -d --build
docker compose ps
docker compose logs -f app
docker compose logs --tail=200 app
docker compose down
```

先跑：

```bash
docker compose config
```

它能提前发现语法问题。

再启动。

如果失败，不要只贴一句：

```text
容器启动失败。
```

要贴：

```bash
docker compose ps
docker compose logs --tail=200 app
docker compose logs --tail=200 db
```

Fable 5 看完整日志，排错会稳很多。

## 8. 常见报错怎么让模型排

### 端口占用

```bash
Error starting userland proxy: listen tcp 0.0.0.0:3000: bind: address already in use
```

检查：

```bash
sudo ss -tulpn | grep 3000
```

让模型判断：

```text
是旧进程占用，还是旧容器占用？
应该停哪个？
会不会影响线上服务？
```

### 环境变量没读到

常见表现：

```text
DATABASE_URL is required
API_KEY missing
```

检查：

```bash
docker compose config
docker compose exec app env | sort
```

不要把真实 Key 贴给模型。

可以脱敏：

```text
DATABASE_URL=postgres://app:***@db:5432/app
```

### 数据库连接失败

常见原因：

```text
app 用 localhost 连数据库。
db 还没 ready。
密码不一致。
网络不在同一个 compose project。
```

在 Compose 里，app 连数据库通常不是 `localhost`，而是服务名：

```text
db
```

这点新手很容易踩坑。

## 9. 数据和 volume：不要随便 prune

Docker 最大的坑之一是 volume。

数据库数据通常在 volume 里。

你如果随手：

```bash
docker volume prune
```

可能直接删掉数据。

所以部署项目时，让 Fable 5 给你一张数据地图：

```text
请根据 docker-compose.yml 生成数据地图：
1. 哪些目录或 volume 是持久化数据
2. 哪些可以重建
3. 哪些必须备份
4. 哪些命令会删除数据
5. 升级前如何备份
```

这个步骤很值钱。

很多部署事故不是应用不会跑，而是人不知道数据在哪里。

## 10. 4SAPI 路由建议

Docker 部署任务可以拆模型：

| 任务 | 推荐模型 |
| --- | --- |
| Compose 架构审查 | Fable 5 |
| 安全风险判断 | Fable 5 |
| 日志摘要 | 低成本模型 |
| 命令解释 | 中低成本模型 |
| 最终上线检查 | Fable 5 |

在 4SAPI 里记录：

```text
project: deploy
stage: docker
task_id: 项目名-docker-日期
model: fable5 / cheap-summary
```

这样后面能复盘：

```text
高级模型是否真的减少了排错时间。
哪些日志摘要可以交给低成本模型。
Docker 安装阶段总共花了多少模型成本。
```

## 11. Docker 安装验收清单

| 检查项 | 标准 |
| --- | --- |
| Docker | `docker --version` 正常 |
| Compose | `docker compose version` 正常 |
| 权限 | deploy 用户能按预期运行 docker |
| Compose 配置 | `docker compose config` 无错误 |
| 容器状态 | `docker compose ps` 正常 |
| 日志 | 能查看 app/db 日志 |
| volume | 知道哪些数据必须备份 |
| 端口 | 只开放必要端口 |
| 密钥 | `.env` 未提交 Git，未发给模型 |

## 12. 总结

用 Fable 5 辅助 Docker，不是让它盲写一堆命令。

而是：

```text
先判断是否需要 Docker。
再安装。
再写最小 Compose。
再审查安全和数据。
再看日志排错。
最后做验收。
```

4SAPI 的价值，是让 Fable 5 这种高级模型用在架构审查和关键排错上。

日志摘要、命令解释这些低价值任务，可以交给便宜模型。

下一篇继续：

```text
用 Fable 5 辅助安装数据库。
Postgres、MySQL、Redis 怎么选，怎么避免公网裸奔。
```
