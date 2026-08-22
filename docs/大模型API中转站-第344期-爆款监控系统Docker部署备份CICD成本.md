---
title: "如何部署爆款监控系统：Docker、备份、CI/CD 与成本"
tags:
  - Docker
  - 部署运维
  - 成本治理
description: "把个人爆款监控系统部署到服务器，配置容器、SQLite 持久化、备份、Worker 鉴权、自动发布和可恢复的成本记录。"
---
# 如何部署爆款监控系统：Docker、备份、CI/CD 与成本

本地能跑起来，不代表爆款监控系统已经可以长期运行。

这类系统有定时扫描、后台队列、AI 分析、逐字稿、图片缓存和数据库写入。如果只在自己的电脑上手动启动，电脑休眠、网络变化或进程崩溃后，监控就会悄悄停止。更麻烦的是，系统可能看起来还在线，但已经连续几天没有采集成功。

本文给出一套适合个人项目的部署思路：使用 Docker Compose 管理服务，数据库文件放在宿主机持久卷，反向代理负责访问保护，备份任务独立运行，代码通过 CI/CD 构建和发布。示例不绑定某一个云平台，托管平台可以是 Dokploy 或其他支持 Compose 的服务。

## 一、先明确生产架构

一个最小部署可以拆成三个容器：

```text
浏览器
   |
   v
proxy
Basic Auth / HTTPS / 静态文件 / API 反向代理
   |
   v
backend
FastAPI / 定时调度 / 扫描 Worker / 分析队列
   |
   +--> SQLite 数据卷
   +--> 媒体缓存卷
   +--> 外部数据和模型服务

backup
SQLite 在线备份 / 保留策略 / 恢复检查
```

`proxy`、`backend` 和 `backup` 的职责要分开。反向代理不应该执行数据迁移，备份容器不应该承担业务请求，后端也不应该把数据库文件写在容器临时层里。

## 二、Docker Compose 的最小骨架

下面只是结构示例，镜像名、端口、域名和环境变量必须换成项目真实值：

```yaml
services:
  proxy:
    image: caddy:2
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
      - ./frontend/dist:/srv:ro
    depends_on:
      - backend

  backend:
    build: ./backend
    restart: unless-stopped
    environment:
      TZ: Asia/Shanghai
      DATABASE_PATH: /data/app.sqlite3
      WORKER_TOKEN_FILE: /run/secrets/worker_token
    volumes:
      - app_data:/data
      - media_cache:/media
    secrets:
      - worker_token

  backup:
    image: alpine:3
    restart: unless-stopped
    volumes:
      - app_data:/data:ro
      - ./backups:/backups
    command: ["/bin/sh", "/backup/backup.sh"]

volumes:
  app_data:
  media_cache:
  caddy_data:
  caddy_config:

secrets:
  worker_token:
    file: ./secrets/worker_token.txt
```

这份 Compose 不能直接生产使用，原因包括：备份脚本路径没有挂载、容器内是否安装 `sqlite3` 未确认、健康检查未配置、镜像构建目录需要补齐。它的价值是展示边界，不是复制粘贴后就能上线。

Docker Compose 官方文档可用于核对服务、卷、密钥和重启策略：[Docker Compose Documentation](https://docs.docker.com/compose/)。

## 三、SQLite 文件必须持久化

SQLite 是一个文件数据库。如果数据库文件写在容器内部，重新创建容器后可能直接丢失。生产配置至少要确认：

- 数据库路径指向持久化卷；
- SQLite 文件和 WAL 文件位于同一卷；
- 容器停止时没有未完成的写入；
- 备份读取的是实际数据库文件；
- 备份目录不和业务数据库放在同一块易损存储上。

可以在容器中检查：

```bash
docker compose exec backend sh
ls -lh /data
sqlite3 /data/app.sqlite3 "PRAGMA journal_mode;"
```

如果启用了 WAL，目录里可能同时出现 `app.sqlite3-wal` 和 `app.sqlite3-shm`。不要只复制主数据库文件就认为备份完成，优先使用 SQLite 官方提供的 `.backup` 机制或在确认一致性后进行快照。

## 四、备份任务要能证明自己有效

一个每天执行但从不验证的备份，不能算完成保护。备份脚本至少要有三部分：创建、保留、验证。

```bash
#!/bin/sh
set -eu

timestamp=$(date -u +%Y%m%dT%H%M%SZ)
source=/data/app.sqlite3
target=/backups/app-${timestamp}.sqlite3

sqlite3 "$source" ".backup '$target'"
sqlite3 "$target" "PRAGMA integrity_check;" | grep -qx ok

find /backups -type f -name 'app-*.sqlite3' -mtime +14 -delete
```

实际使用前要确认备份容器中确实安装了 `sqlite3`，路径和权限也要检查。`integrity_check` 通过只说明数据库结构完整，不代表应用数据业务上没有误删。

每周至少做一次隔离恢复：

```bash
cp backups/app-20260729T000000Z.sqlite3 /tmp/restore.sqlite3
sqlite3 /tmp/restore.sqlite3 "PRAGMA integrity_check;"
sqlite3 /tmp/restore.sqlite3 "SELECT COUNT(*) FROM works;"
```

然后用临时应用连接恢复库，检查创作者、作品、评分、分析和逐字稿是否都能读取。线上数据库和备份目录在同一台服务器上时，服务器损坏可能同时带走两者，应定期复制到另一个受控位置。

## 五、反向代理和访问保护

个人工具不一定需要完整注册登录系统，但不能把 Dashboard、API 文档、Worker API 和数据库操作接口裸奔到公网。

反向代理至少负责：

- HTTPS 终止；
- 前端静态文件；
- `/api` 到后端的转发；
- 请求大小、超时和访问日志；
- 访问保护和健康检查。

Basic Auth 可以作为个人工具的第一层保护，但它不是细粒度权限系统。Worker API 则应使用单独的 Bearer Token，不能复用浏览器 Basic Auth：

```text
浏览器 Dashboard -> Basic Auth / HTTPS
Mac Mini Worker  -> Bearer Token / HTTPS
备份任务         -> 本地卷权限
```

Token 应从环境变量、Docker Secret 或平台密钥管理注入，不要写入前端代码、Git 仓库、日志和截图。出现泄露迹象时，轮换 Token，并检查过去的访问日志。

## 六、Mac Mini Worker 怎么接入

如果 L2 深度分析需要本地模型、视频下载或本地 Claude Code，可以让 Mac Mini 定时认领任务，而不是把服务器上的后端和本地分析强行部署在同一台机器。

推荐的交互是：

```text
Mac Mini 定时请求 /worker/claim
        |
服务器返回一个分析任务和证据引用
        |
本地下载允许使用的媒体或读取逐字稿
        |
本地完成分析和 JSON 校验
        |
请求 /worker/result 回传结果
```

任务接口需要有状态和租约，避免 Mac Mini 断电后任务永久卡住：

```text
pending -> leased -> succeeded
                  \-> failed
leased --超时--> pending
```

服务端要记录 `leased_at`、`lease_owner`、`attempts` 和 `last_error`。Worker 回传结果时，再检查任务是否仍归它持有，避免旧 Worker 覆盖新结果。

## 七、CI/CD 不只是“推代码就部署”

一个最小发布链路是：

```text
推送 main
   |
GitHub Actions 运行测试
   |
构建 Docker 镜像
   |
推送到镜像仓库
   |
通知部署平台拉取新镜像
   |
健康检查通过后切换版本
```

流水线至少需要这些检查：

- Python 依赖可以安装；
- 数据库迁移在临时库上成功；
- API 结构化输出测试通过；
- 前端构建成功；
- Docker 镜像可以启动；
- 健康检查返回预期状态；
- 镜像和部署凭据没有写进日志。

GitHub Actions 的密钥应放在仓库的 Secrets 中，部署 Token 使用最小权限。不要把生产 Token 写入工作流 YAML。自动部署失败时，保留上一版可用镜像，不能让平台无条件删除旧版本。

## 八、时区和定时任务

扫描计划如果按北京时间运行，后端容器的时区和调度器时区必须明确。推荐所有数据库时间使用 UTC 保存，展示和任务计划再转换为 `Asia/Shanghai`，或者在整个系统中明确采用北京时间，但不能一半使用 UTC、一半使用本地时间。

验收时检查：

```bash
docker compose exec backend date -Is
docker compose logs backend --since 2h
```

看日志里是否真的在预期时间创建了扫描任务，而不是只看配置文件里的字符串。夏令时、服务器时区、容器默认时区和平台调度器可能互相影响，部署后应做一次人工触发测试。

## 九、健康检查和故障告警

“容器还在运行”不等于“系统可用”。后端健康检查至少区分：

```text
/health/live   进程是否还活着
/health/ready  数据库、必要配置和依赖是否可用
```

另外要监控业务指标：最近一次成功扫描时间、失败任务数量、待分析队列长度、备份最后成功时间和磁盘剩余空间。

一个简单的告警条件可以是：

```text
超过 26 小时没有成功扫描 -> 数据采集异常
失败任务连续超过 3 次 -> 需要人工检查
备份超过 36 小时未更新 -> 备份异常
磁盘使用率超过 80% -> 清理媒体缓存或扩容
```

这些数字是示例阈值，要根据扫描频率、数据重要性和可接受延迟调整。

## 十、成本怎么核算

成本不要只看服务器月租。完整账单通常包含：

| 项目 | 成本驱动因素 |
| --- | --- |
| 数据源 | 创作者数量、扫描频率、分页和平台数量 |
| L1 分析 | 候选作品数量、输入长度、重试次数 |
| L2 分析 | 深拆数量、视频处理、模型运行时间 |
| 语音识别 | 音频时长、语言、是否重复识别 |
| 服务器 | CPU、内存、磁盘、流量和备份空间 |
| 媒体存储 | 原视频、抽帧、封面和缓存保留时间 |

原始素材中的“每月约几十美元”只能作为作者当时的个人实测估算，不能直接当成所有人的固定价格。正式发布时应补充账单日期、用量、地区、套餐和是否包含本地电费；否则更稳妥的写法是说明成本结构和计算公式。

可以记录一张月度表：

```text
月成本 = 数据源费用
       + L1 调用量 × 单价
       + L2 实际运行量 × 单价
       + ASR 音频分钟数 × 单价
       + 服务器与存储
       + 备份和流量
```

每月还要记录“人工复盘节省时间”和“真正使用的选题数量”。系统不是调用越多越有价值，最终要看它是否减少了无效浏览，并帮助你更快做出自己的内容。

## 十一、上线前的恢复演练

部署结束前至少演练三种故障：

1. 删除一个临时容器，确认数据卷和数据库仍然存在。
2. 停止后端，确认代理能显示明确的错误或健康状态。
3. 用备份复制出临时数据库，确认作品、评分和分析可读。

如果系统包含生产数据，不要在没有备份的情况下直接执行 `docker compose down -v`。这个命令可能删除命名卷，具体影响取决于 Compose 配置。任何破坏性操作都要先确认目标、备份和回滚方式。

## 结论

个人爆款监控系统可以用三容器、SQLite 和一个小服务器跑起来，但“简单”不等于可以省略持久化、备份、鉴权和恢复演练。

部署的最小闭环应该是：代码可以重复构建，数据库和媒体不会因重建丢失，Worker 有独立认证，定时任务有日志，备份可以恢复，发布失败可以回到上一版。成本则要按真实用量核算，不能把一次个人账单包装成普遍价格。

## 参考资料

- [Docker Compose 官方文档](https://docs.docker.com/compose/)，用于核对服务、卷、密钥和生命周期配置。
- [SQLite Backup API 文档](https://www.sqlite.org/backup.html)，用于核对在线备份行为。
- [Caddy 官方文档](https://caddyserver.com/docs/)，用于核对反向代理和 HTTPS 配置。
- [GitHub Actions 官方文档](https://docs.github.com/en/actions)，用于核对流水线、Secrets 和发布流程。
