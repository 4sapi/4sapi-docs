---
title: "【大模型API中转站】第166期 AI排查Docker容器起不来 | ps和logs三件套"
category: 人工智能
tags:
  - 大模型API中转站
  - 4SAPI
  - AI排错
  - Docker
  - Docker Compose
  - Claude Fable 5
  - 服务器部署
description: "Docker 容器起不来时，不要只贴 exited。本文给出 docker compose ps、logs、config、环境变量、端口、volume、healthcheck 组成的 AI 排错包，并说明辅助模型、Fable 5 与 4SAPI 如何分工。"
---

# 【大模型API中转站】第166期 AI排查Docker容器起不来 | ps和logs三件套

本文是【大模型API中转站】系列的第166篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

Docker 是独立开发者部署项目的救命工具。

也是制造崩溃的高频来源。

最常见的一句话：

```text
容器起不来。
```

但“起不来”至少有五种情况：

```text
build 失败。
容器启动后立刻退出。
容器一直 restarting。
容器 running 但 unhealthy。
容器 running 但端口访问不到。
```

每一种排查路径都不一样。

所以用 AI 排查 Docker，第一步不是让模型给命令。

第一步是把状态说清楚。

## 1. Docker 排错三件套

最小三件套：

```bash
docker compose ps
docker compose logs --tail=200
docker compose config
```

含义：

| 命令 | 用途 |
| --- | --- |
| `docker compose ps` | 看服务状态、端口、退出码 |
| `docker compose logs --tail=200` | 看最近错误 |
| `docker compose config` | 展开最终 Compose 配置 |

给 AI 时，最好按服务拆。

例如：

```bash
docker compose logs --tail=200 app
docker compose logs --tail=200 db
docker compose logs --tail=200 nginx
```

不要把所有服务几千行混在一起。

先让辅助模型整理，再让 Fable 5 判断根因。

## 2. 给 AI 的 Docker 排错包

模板：

```text
【现象】
- 哪个服务起不来：
- 状态：exited / restarting / unhealthy / running but inaccessible
- 退出码：
- 是否最近部署后出现：

【环境】
- 服务器系统：
- Docker 版本：
- Docker Compose 版本：
- 部署目录：

【compose】
- docker-compose.yml 脱敏版：
- .env 变量名清单，不含真实值：

【状态】
- docker compose ps 输出：

【日志】
- app logs：
- db logs：
- nginx logs：

【最近变更】
- 镜像：
- 环境变量：
- volume：
- 端口：
- 数据库迁移：

【边界】
- 不删除 volume
- 不清空数据库
- 不重启整机
- 先给只读验证步骤
```

这份包能让 AI 迅速判断：

```text
是应用自身崩。
是数据库没 ready。
是环境变量缺。
是端口映射错。
是 volume 权限问题。
还是 healthcheck 写错。
```

## 3. 常见状态一：Exited

如果 `docker compose ps` 里看到：

```text
app exited with code 1
```

优先看 app 日志。

常见原因：

```text
启动命令错。
环境变量缺失。
依赖文件不存在。
数据库连接失败后进程退出。
端口被占用。
代码运行时报错。
```

给 AI 的关键证据：

```text
退出码。
最后 50 行日志。
Dockerfile CMD/ENTRYPOINT。
compose command。
环境变量名称。
```

Prompt：

```text
请根据 app 容器退出码和最后 200 行日志，判断退出发生在启动脚本、应用初始化、数据库连接还是运行时。
每个判断必须引用日志行。
先给只读验证步骤，不要建议删除 volume。
```

## 4. 常见状态二：Restarting

`restarting` 通常说明：

```text
容器启动后崩溃。
restart policy 又把它拉起来。
于是循环。
```

这时最重要的是抓住第一段错误。

命令：

```bash
docker compose logs --tail=300 app
```

如果日志太滚，可以先暂停自动重启，但生产环境要谨慎。

更稳的是：

```bash
docker inspect 容器名 --format '{{.State.ExitCode}} {{.State.Error}}'
```

让 AI 判断：

```text
是应用崩溃导致重启。
还是 healthcheck 失败导致重启。
还是外部依赖不可用导致重启。
```

注意：

```text
不要一看到 restarting 就重启服务器。
```

服务器重启只是把问题重新演一遍。

## 5. 常见状态三：Unhealthy

容器 `running` 但 `unhealthy`，说明进程还在，但健康检查失败。

常见原因：

```text
healthcheck 路径错。
应用启动慢，start_period 太短。
healthcheck 用 localhost 但服务监听 0.0.0.0 或反过来。
接口需要认证。
数据库还没 ready。
```

给 AI 的材料：

```text
healthcheck 配置。
docker inspect 里的 Health.Log。
应用实际监听端口。
curl 容器内健康检查结果。
```

只读命令：

```bash
docker inspect 容器名 --format '{{json .State.Health}}'
```

Prompt：

```text
请判断 unhealthy 是应用真实不可用，还是 healthcheck 配置不匹配。
请给出容器内 curl 验证方式和最小修改建议。
```

这类问题很适合 Fable 5。

因为它要同时看 Compose、应用端口、healthcheck 和日志。

## 6. 常见状态四：端口访问不到

容器 running，但浏览器打不开。

先分三层：

```text
容器内服务是否监听。
宿主机端口是否映射。
云服务器防火墙是否放行。
```

命令：

```bash
docker compose ps
ss -tulpn
curl -I http://127.0.0.1:端口
```

常见错误：

```text
应用只监听 127.0.0.1。
compose 没写 ports。
ports 写成了 expose。
宿主机防火墙没开。
云安全组没开。
Nginx upstream 指向错端口。
```

给 AI 时要说明：

```text
容器内端口。
宿主机映射端口。
Nginx 反代端口。
云安全组端口。
```

否则它很容易猜错层。

## 7. 常见状态五：volume 权限问题

日志里可能出现：

```text
permission denied
EACCES
cannot create directory
read-only file system
```

常见于：

```text
容器用非 root 用户运行。
宿主机目录权限不对。
挂载了只读 volume。
数据库 data 目录被错误 owner。
```

这类问题要谨慎。

不要让 AI 直接给：

```bash
chmod -R 777
```

这种粗暴命令。

Prompt：

```text
请根据 volume 配置和 permission denied 日志，判断是宿主机目录权限、容器用户、只读挂载还是应用路径错误。
不要建议 chmod -R 777。
请给出最小权限修复方案和回滚方式。
```

这是非常重要的约束。

很多网上教程遇到权限问题就 `777`。

生产环境不要这样。

## 8. 常见状态六：depends_on 误解

很多人以为：

```yaml
depends_on:
  - db
```

就代表数据库已经 ready。

不是。

它通常只代表启动顺序。

不代表服务可用。

所以 app 可能先启动，然后数据库还没 ready，app 连接失败退出。

修复方向：

```text
应用层重试数据库连接。
给 db 加 healthcheck。
depends_on 使用 condition: service_healthy。
启动脚本等待依赖。
```

让 AI 审查：

```text
请检查这个 docker-compose.yml 是否误以为 depends_on 等于数据库可用。
如果是，请给出最小改法。
```

## 9. 4SAPI 模型分工

Docker 排错建议分工：

| 阶段 | 模型 |
| --- | --- |
| logs 摘要 | 低成本模型 |
| compose 风险审查 | Fable 5 |
| 根因判断 | Fable 5 |
| 命令解释 | 低成本模型 |
| 修复方案复核 | Fable 5 / reviewer |

4SAPI 日志字段：

```text
error_type: docker_container_failed
service: app/db/nginx
status: exited/restarting/unhealthy
risk_level
model
evidence_ids
fix_requires_human_confirm
```

如果涉及数据库 volume，自动标高风险。

不要让低成本模型直接给破坏性命令。

## 10. 一条完整 Prompt

```text
你是 Docker Compose 排错助手。

我会提供 docker compose ps、logs、config 和最近变更。

请先判断容器问题属于：
1. build 失败
2. 启动命令失败
3. 应用运行时报错
4. 依赖服务未 ready
5. 环境变量缺失
6. 端口映射错误
7. volume 权限问题
8. healthcheck 配置问题

要求：
- 每个判断引用证据。
- 先给只读验证命令。
- 不要建议删除 volume、清空数据、chmod -R 777 或重启整机。
- 涉及数据库、证书、生产环境变量时，标注需要人工确认。
```

这条可以作为团队排障 Skill。

## 11. 修复后要让 AI 写部署记录

容器修好后，别马上结束。

让 AI 写：

```text
问题现象。
根因。
验证证据。
修复动作。
是否涉及数据。
是否需要监控。
下次如何避免。
```

这份记录以后会救你。

因为 Docker 问题很容易反复出现。

尤其是：

```text
环境变量。
端口。
volume。
healthcheck。
depends_on。
```

这些每个项目都会遇到。

## 12. 总结

Docker 容器起不来，不要只告诉 AI：

```text
exited。
```

要给它：

```text
ps。
logs。
config。
env 变量名。
最近变更。
端口映射。
volume。
healthcheck。
```

通过 4SAPI：

```text
辅助模型整理日志。
Fable 5 判断跨层根因。
复核模型检查危险命令。
```

最后记住：

```text
Docker 排错先只读验证，不要先删 volume。
```

下一篇写数据库连接失败。

因为很多容器问题最后都会指向那一句：

```text
app 连接不上 db。
```
