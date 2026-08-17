---
title: "LightRAG 部署如何划分权限、日志与成本边界"
tags:
  - LightRAG
  - 部署运维
  - 权限治理
description: "LightRAG 从个人实验走向团队服务后，权限、日志、备份和成本会成为独立的工程问题。"
---
# LightRAG 部署如何划分权限、日志与成本边界
LightRAG 从个人实验走向团队服务后，权限、日志、备份和成本会成为独立的工程问题。本文把部署拆成访问边界、数据生命周期、调用观测、资源预算和故障恢复五部分，帮助你先定义可验收的治理要求，再决定需要哪些组件和自动化。文中只讨论可复现的步骤，不把单次结果扩展成产品承诺；每个结论都标注前提、证据和无法覆盖的边界。读者可以先完成最小验证，再按自己的版本、权限和数据补充实验。


前面三篇，我们已经完成了 LightRAG 的入门闭环：

~~~text
第306期：启动 WebUI、REST API，上传文档并完成第一次查询
第307期：理解文本块、向量、实体、关系和五种查询模式
第308期：处理 PDF、DOCX、表格、图片，做增量更新和存储选型
~~~

但“本机能回答问题”和“团队可以长期使用”之间，还有一段很长的距离。

生产环境会同时面对这些问题：

- 默认绑定 `0.0.0.0`，没有认证时，谁能访问文档和知识图谱？
- 上传一个大 PDF，为什么请求还没到 LightRAG 就被 Nginx 返回 `413`？
- 只有一个实例时使用 Uvicorn 就够了，为什么还需要 Gunicorn？
- 两个团队共享一个 PostgreSQL 或 Milvus 时，怎样避免数据混在一起？
- 文档索引、实体抽取、Embedding、关键词生成和最终问答，究竟是谁消耗了预算？
- 删除文档、重处理文档或更换 Embedding 后，怎样备份、恢复和回滚？

这一篇把 LightRAG 从“开发机上的知识库”推进到“可以接入企业系统的服务”。重点不是把配置项全部抄一遍，而是把部署层、数据层、模型层和 API 网关层的责任拆开。

项目地址：[HKUDS/LightRAG](https://github.com/HKUDS/LightRAG)

文中命令和环境变量以 2026 年 7 月检出的当前仓库为准。LightRAG 仍在迭代，实际部署前请再次查看仓库的 `README-zh.md`、`env.example`、`docs/LightRAG-API-Server-zh.md` 和当前版本的 `/docs`。

## 一、先分清“能启动”和“能上线”

LightRAG Server 的职责是提供文档处理、知识图谱、检索、问答和 WebUI。它可以调用 OpenAI-compatible、Ollama、Azure OpenAI、Bedrock、Gemini 等模型后端，也可以连接多种外部存储。

但企业系统还需要另外几层：

~~~text
浏览器 / 客服系统 / Agent / 内部应用
              |
        Nginx / Ingress / API 网关
        TLS、域名、限流、审计、预算
              |
        LightRAG Server
        文档处理、图谱、检索、查询、WebUI
              |
   PostgreSQL / MongoDB / OpenSearch / 向量库 / 图数据库
              |
        LLM、Embedding、Rerank、VLM
~~~

每一层解决的问题不同：

| 层 | 主要职责 | 不应该假设它自动解决什么 |
| --- | --- | --- |
| 业务应用 | 用户、项目、角色和业务权限 | 不应该把一次查询当成完整的权限审查 |
| 反向代理或 API 网关 | HTTPS、限流、审计、路由和统一模型入口 | 不会自动理解 PDF 里的敏感字段 |
| LightRAG Server | 文档入库、图谱、检索、问答 | 不保证抽取结果绝对正确 |
| 存储后端 | KV、向量、图和文档状态的持久化 | 不会自动完成跨后端迁移 |
| 模型服务 | 抽取、关键词、回答、视觉分析和向量化 | 不会替你选择正确的模型和预算 |

因此，企业部署不是简单地给 `lightrag-server` 加一个 `--workers 8`。先把数据隔离、认证、存储和模型成本设计好，再调吞吐量，问题会少很多。

## 二、生产环境优先选择 Docker

LightRAG 当前仓库提供 Docker Compose 入口。它适合把服务、数据目录和外部存储依赖放在一套可重复配置里。

### 1. 最小启动

Linux、macOS 或 Windows PowerShell 都可以先准备仓库和配置：

~~~powershell
git clone https://github.com/HKUDS/LightRAG.git
cd LightRAG
Copy-Item .\env.example .\.env
~~~

Linux 或 macOS 也可以使用：

~~~bash
cp env.example .env
~~~

先在 `.env` 里配置 LLM、Embedding 和服务端口，再启动：

~~~bash
docker compose up -d
docker compose ps
docker compose logs -f lightrag
~~~

首次部署建议先把 LightRAG 只暴露在本机或内网：

~~~env
HOST=127.0.0.1
PORT=9621
~~~

如果容器需要被 Nginx 或同一内网的其他容器访问，容器内部可以监听 `0.0.0.0`，但宿主机端口仍然建议只绑定到 `127.0.0.1`，再由反向代理统一接入。

~~~yaml
ports:
  - "127.0.0.1:9621:9621"
~~~

### 2. 数据目录不能只看容器状态

官方 Docker 部署默认会把数据放在类似下面的目录：

~~~text
data/
├── rag_storage/    # RAG 派生数据、索引、图和文档状态
└── inputs/         # 上传或待处理的原始文档
~~~

`docker compose down` 只会停止并移除容器，不等于备份数据。生产环境至少要明确以下内容分别存在哪里：

- 原始输入文件和重新处理所需的 sidecar。
- KV、向量、图和 Doc Status 数据。
- LLM 缓存、解析缓存和日志。
- `.env`、数据库密码、模型 Key 和 TLS 私钥。

官方镜像的运行进程使用 `lightrag` 用户，默认 uid/gid 为 `1000`。如果挂载自定义目录或 PVC，确认容器用户对 `WORKING_DIR`、`INPUT_DIR`、`PROMPT_DIR` 和缓存目录拥有需要的读写权限。不要为了省事把宿主机根目录挂进去，也不要把整个宿主机目录设置成所有用户可写。

生产环境建议固定明确的镜像版本或发布 tag，并把镜像更新放到变更流程里。直接跟随 `latest` 会让“代码、数据格式、解析器和存储后端”同时发生变化，出问题时很难定位。

## 三、Linux 用 Gunicorn，Windows 识别边界

LightRAG 当前文档提供两类服务入口：

### 1. 单进程 Uvicorn 模式

开发、验证和 Windows 环境可以使用：

~~~bash
lightrag-server
~~~

也可以显式指定关键参数：

~~~bash
lightrag-server \
  --host 127.0.0.1 \
  --port 9621 \
  --working-dir ./rag_storage \
  --input-dir ./inputs \
  --workspace team-a
~~~

注意，LightRAG 启动时会从当前工作目录读取 `.env`。同一台机器运行多个实例时，要么为每个实例准备独立工作目录和 `.env`，要么通过命令行参数覆盖端口和 workspace。修改 `.env` 后重新打开终端，可以避免旧的系统环境变量继续覆盖新配置。

### 2. Linux 的 Gunicorn + Uvicorn 模式

生产 Linux 环境可以使用：

~~~bash
lightrag-gunicorn --workers 2
~~~

Gunicorn 的意义是把请求服务放进多个工作进程，降低文档索引任务长期占用单个 API 进程时对查询请求的影响。它不是把所有文档写入操作变成无限并行，也不是把外部 LLM 的速度直接变快。

当前仓库明确标注 Gunicorn 多进程模式不支持 Windows。Windows 用户可以选择：

- 本地开发使用 `lightrag-server`。
- 使用 Docker Desktop，让 LightRAG 在 Linux 容器中运行。
- 使用 WSL2 或 Linux 服务器承载生产实例。

不要在 Windows 上照抄 `lightrag-gunicorn`，遇到启动失败时先确认运行环境，而不是先改一堆 worker 参数。

### 3. 并发参数从小处开始

文档索引还会受到这些参数影响：

~~~env
WORKERS=2
MAX_PARALLEL_INSERT=3
MAX_ASYNC_LLM=4
~~~

它们分别对应不同层次：

- `WORKERS`：Gunicorn 工作进程数。
- `MAX_PARALLEL_INSERT`：一批并行处理的文件数量。
- `MAX_ASYNC_LLM`：LLM 调用的最大异步并发。

此外还有解析阶段并发、Embedding 并发和查询上下文长度等设置。每个参数调大，都可能增加 API 网关并发、模型限流、内存、数据库连接和费用。

更稳的调参顺序是：先用固定数量的文件和固定问题集，记录索引耗时、查询 p50/p95、失败率、模型请求数和费用，再只修改一个并发参数。不要把“机器 CPU 空闲”误认为“可以继续提高 LLM 并发”。真正的瓶颈可能在上游模型、Embedding 服务或数据库连接池。

## 四、反向代理要处理上传和流式查询

LightRAG 自带 WebUI 和 REST API，但生产环境通常由 Nginx、Traefik 或 Kubernetes Ingress 负责域名、TLS 和外部访问控制。

### 1. 大文件上传

Nginx 默认的 `client_max_body_size` 可能只有 1 MB。LightRAG 的 `/documents/upload` 还没有开始处理文件，Nginx 就可能先返回 `413 Request Entity Too Large`。

一个简化的 Nginx 配置可以这样写：

~~~nginx
server {
    listen 443 ssl;
    server_name rag.example.com;

    client_max_body_size 8M;

    location /documents/upload {
        client_max_body_size 100M;
        proxy_pass http://127.0.0.1:9621;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;
    }

    location ~ ^/(query/stream|api/chat|api/generate) {
        gzip off;
        proxy_pass http://127.0.0.1:9621;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
        proxy_buffering off;
        proxy_read_timeout 300s;
    }

    location / {
        proxy_pass http://127.0.0.1:9621;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 300s;
    }
}
~~~

这里有三个容易遗漏的点：

1. 上传限制要在代理层和应用层分别设置，避免应用收到超大文件后才失败。
2. `/query/stream` 和 Ollama 兼容的流式接口需要较长读取超时，且不要让代理把响应缓冲到最后才发送。
3. Nginx、LightRAG、上游 API 和上游模型的超时时间要有层次。不能 Nginx 60 秒超时，而 LightRAG 允许 240 秒，最后把正常的长查询误判为失败。

### 2. 多站点前缀

一台主机上承载多个知识库实例时，可以为每个后端设置不同的 `LIGHTRAG_API_PREFIX`：

~~~env
LIGHTRAG_API_PREFIX=/site01
~~~

对应启动方式：

~~~bash
lightrag-server --port 9621
~~~

Nginx 把 `/site01/` 转发到第一台实例，把 `/site02/` 转发到第二台实例，并剥离站点前缀。当前仓库的 WebUI 使用运行时配置注入，同一份 WebUI 构建产物可以被多个前缀复用，不需要每个站点重新构建前端。

多站点配置中，至少要让下面几项同时不同：

~~~text
站点前缀：/site01、/site02
工作空间：site01、site02
数据目录或外部存储命名空间
API Key、JWT 账号和业务访问策略
~~~

仅仅改变 URL 前缀，不会自动隔离数据。仅仅改变 workspace，也不会自动创建用户级 RBAC。

## 五、认证要同时保护 API 和 WebUI

LightRAG Server 默认可以不带认证访问。只要监听在 `0.0.0.0`，又没有配置认证，就应该把它当成“网络上任何人都能访问文档和知识图谱”的服务。

### 1. API Key

服务端可以配置 API Key：

~~~env
LIGHTRAG_API_KEY=<生成一个高强度随机值>
WHITELIST_PATHS=/health
~~~

调用时通过请求头传递：

~~~bash
curl http://127.0.0.1:9621/health \
  -H "X-API-Key: <your-key>"
~~~

不要把 Key 放到 URL、前端 JavaScript、代码仓库或错误日志中。业务应用应该把 Key 存到 Secret Manager，再由服务端注入环境变量。

### 2. JWT 账号

WebUI 登录使用 JWT。当前仓库支持的配置形式包括：

~~~env
AUTH_ACCOUNTS='admin:{bcrypt}<bcrypt-hash>'
TOKEN_SECRET=<非默认的随机密钥>
TOKEN_EXPIRE_HOURS=4
~~~

可以使用项目提供的命令生成可粘贴到 `AUTH_ACCOUNTS` 的密码哈希：

~~~bash
lightrag-hash-password --username admin
~~~

当前版本的账号能力仍然偏基础，不能把 `AUTH_ACCOUNTS` 当成完整的企业身份系统。更稳的做法是：

- 在企业网关或上游应用完成组织、用户和项目权限判断。
- LightRAG 自身使用 API Key 或 JWT 作为服务访问的第二道边界。
- 业务应用把“用户可以访问哪个 workspace”映射成明确的路由和服务账号。

只配置 API Key 而没有配置账号时，WebUI 仍可能以 Guest 方式访问。要同时保护 API 和 WebUI，按当前仓库文档，应同时配置 API Key 和账号认证。单个请求只发送一种认证头：`X-API-Key` 或 `Authorization: Bearer <token>`。两种都发送时，JWT 会优先校验，失效的 JWT 会让请求直接失败。

### 3. 收紧白名单

为了兼容 Ollama 客户端，`/api/*` 默认可能被放进 `WHITELIST_PATHS`，因此即使配置了认证，Ollama 兼容路径仍可能不要求认证。

如果企业不使用 Open WebUI 或 Ollama 兼容接口，建议收窄为：

~~~env
WHITELIST_PATHS=/health
~~~

当前 `/health` 未认证请求只返回存活信号，完整运行配置和诊断信息需要有效的 JWT 或 API Key。这样既能让负载均衡器做 liveness probe，也不会把模型、workspace、队列等运行信息全部暴露给匿名访问者。

此外，`CORS_ORIGINS` 不要长期使用 `*`。如果是另一个浏览器应用跨域调用，设置明确的来源列表；同源 WebUI 不需要为了方便打开全局跨域。

## 六、workspace 不是完整权限系统

LightRAG 支持通过 `workspace` 对多个实例或项目做逻辑隔离。例如：

~~~bash
lightrag-server --port 9621 --workspace space1
lightrag-server --port 9622 --workspace space2
~~~

不同存储后端使用 workspace 的方式不完全相同：

- 本地文件存储使用不同的子目录。
- PostgreSQL 等关系型存储可以使用 `workspace` 字段做逻辑隔离。
- Milvus、Qdrant、MongoDB 等集合型存储通常把 workspace 放进集合或命名空间。
- Neo4j 可以使用 label 做逻辑隔离。
- OpenSearch 使用索引名称前缀做隔离。

这带来两个生产要求。

第一，同一个外部数据库承载多个项目时，必须明确公共 `WORKSPACE` 和各存储专用 workspace 配置，不能只在应用层给 URL 加一个项目名。

第二，workspace 不是权限校验。若用户可以自行提交 `workspace=finance`，那它就可能绕过你的项目路由。workspace 应该由服务端根据用户身份和项目关系生成，或者为每个项目运行独立实例，不能直接信任浏览器传来的任意值。

## 七、异步索引、健康检查和日志

LightRAG 的文档插入是异步流程。`/documents/upload`、`/documents/text` 和 `/documents/texts` 会返回 `track_id`，应用应保存这个 ID，再轮询：

~~~text
提交文档
  -> 保存 track_id、用户、项目、源文件哈希
  -> 查询 /documents/track_status/{track_id}
  -> 等待 processed 或 failed
  -> 处理失败进入重试或人工队列
  -> 查询前检查文档是否已经可用
~~~

不要在收到上传 HTTP 200 后立刻告诉用户“知识库已经更新”。HTTP 200 只代表任务被服务接受，文档是否完成解析、抽取、Embedding 和写入，还要看 track status。

服务端还提供 `/documents/pipeline_status`，用于观察处理循环、扫描、破坏性任务和队列状态。`/health` 可作为存活探针，但生产监控最好分成三类：

| 监控类型 | 例子 | 处理方式 |
| --- | --- | --- |
| Liveness | 进程是否能响应 `/health` | 由负载均衡器判断是否重启或摘除 |
| Readiness | LLM、Embedding、存储和处理队列是否可用 | 未准备好时停止接收新业务流量 |
| 业务指标 | `track_id` 卡住、失败率、查询延迟、引用为空 | 告警、重试和人工排查 |

LightRAG 当前配置还包括 `LOG_LEVEL`、`LOG_DIR`、`LOG_MAX_BYTES`、`LOG_BACKUP_COUNT` 和可选的性能计时日志。日志中建议保留：

~~~text
request_id、workspace、业务项目、接口、查询模式
模型角色、模型名、耗时、输入/输出 token、重试次数
HTTP 状态码、错误类型、track_id、文档哈希
~~~

不建议默认记录：

~~~text
完整 PDF、完整合同正文、API Key、数据库密码、JWT、用户身份证号
~~~

上游 API 或企业 API 网关的审计日志可以记录模型调用和费用，但也要遵守同样的脱敏原则。需要调试召回质量时，使用受控的 `/query/data`、固定测试文档和有限保留期，不要把所有生产原文永久复制到日志平台。

## 八、备份、恢复和版本升级

LightRAG 的知识库不是只有一个向量索引。一次完整的备份策略至少要覆盖：

1. 原始输入文件和文件版本元数据。
2. KV、Vector、Graph、Doc Status 四类存储。
3. 解析 sidecar、提示词和实体类型配置。
4. 当前 LightRAG 版本、Embedding 模型、维度、分块策略和解析器路由。
5. 需要恢复查询的模型配置和外部数据库版本。

密钥不应和业务备份放在同一个公开目录。可以把密钥放入企业 Secret Manager，备份中只保留密钥引用和恢复说明。

### 1. 删除是破坏性操作

删除一份文档可能影响：

- 文档状态。
- 归属于它的文本块和向量。
- 只由它支持的实体和关系。
- 被其他文档共享的实体描述和关系摘要。
- 建库阶段产生的 LLM 缓存。

当前仓库提供了受影响关系的重建流程，但删除仍然应该进入审批、队列和审计，而不是和普通上传任务无限并发。

### 2. 更换 Embedding 不等于改一个环境变量

如果修改下面任何一类配置：

~~~text
Embedding 模型
Embedding 维度
对称 / 非对称嵌入设置
query/document 前缀
provider task 行为
~~~

旧向量和新向量的语义空间就可能不一致。当前仓库没有通用的“原库在线重新嵌入”工具，生产做法应该是：

~~~text
停止写入
  -> 备份原文和当前索引
  -> 建立新的 workspace 或新索引
  -> 使用新 Embedding 全量重建
  -> 用固定问题集对比召回和回答
  -> 切换读流量
  -> 保留旧索引一段观察期
~~~

修改解析器路由也有类似边界：`LIGHTRAG_PARSER` 和文件名 hint 主要影响新上传文件。已有文档要切换解析器，通常需要删除后重新上传或按当前版本提供的重新处理流程执行，不能只改 `.env` 后期待旧文档自动换解析结果。

### 3. 做一次真正的恢复演练

备份文件存在，不等于可以恢复。至少每个季度做一次隔离恢复：

- 在新目录或新 workspace 启动同版本 LightRAG。
- 恢复一小批原始文件和对应存储。
- 检查实体、关系、文本块、引用和查询结果。
- 模拟一个文档删除和一次失败重试。
- 记录恢复耗时、缺失文件和人工操作步骤。

如果只能恢复出向量，却找不到原始文件版本和解析配置，这份备份对企业知识库并不完整。

## 九、用 上游 API 做多角色模型治理

LightRAG 当前不只调用一个“聊天模型”。按角色可以拆成：

| 角色 | 触发阶段 | 适合的治理策略 |
| --- | --- | --- |
| `EXTRACT` | 文档插入、实体关系抽取和摘要合并 | 高并发、成本可控、上下文足够的模型 |
| `KEYWORD` | 查询前生成 high-level / low-level keyword | 低延迟模型，限制并发和失败重试 |
| `QUERY` | 根据召回内容生成最终回答 | 更强模型，重点关注长上下文质量 |
| `VLM` | 图片、表格、公式的视觉分析 | 只为需要 `i/t/e` 的文档单独计费 |

此外，Embedding、Rerank 也有自己的请求和预算。企业 API 网关的价值在于把这些调用变成可路由、可计费、可审计的模型请求，而不是替 LightRAG 决定知识库是否可信。

### 1. OpenAI-compatible 的 上游 API 角色配置

假设 上游 API 提供 OpenAI-compatible 的聊天和 Embedding 接口，LightRAG 可以从 `openai` binding 方向配置。下面只使用占位符：

~~~env
LLM_BINDING=openai
LLM_MODEL=<extract-model>
LLM_BINDING_HOST=https://<provider-endpoint>/v1
LLM_BINDING_API_KEY=<extract-key>
LLM_TIMEOUT=240
MAX_ASYNC_LLM=4

# 最终问答单独使用更强模型或独立 Key
QUERY_LLM_BINDING=openai
QUERY_LLM_MODEL=<query-model>
QUERY_LLM_BINDING_HOST=https://<provider-endpoint>/v1
QUERY_LLM_BINDING_API_KEY=<query-key>
QUERY_MAX_ASYNC_LLM=2
QUERY_LLM_TIMEOUT=240

# 查询关键词使用低延迟模型
KEYWORD_LLM_BINDING=openai
KEYWORD_LLM_MODEL=<keyword-model>
KEYWORD_LLM_BINDING_HOST=https://<provider-endpoint>/v1
KEYWORD_LLM_BINDING_API_KEY=<keyword-key>
KEYWORD_MAX_ASYNC_LLM=4

# 只有开启图片、表格或公式分析时才配置 VLM
VLM_LLM_BINDING=openai
VLM_LLM_MODEL=<vision-model>
VLM_LLM_BINDING_HOST=https://<provider-endpoint>/v1
VLM_LLM_BINDING_API_KEY=<vlm-key>
~~~

角色覆盖变量的具体继承规则以当前仓库的 `RoleSpecificLLMConfiguration-zh.md` 为准。同一个 provider 只换模型时可以继承基础 host 和 Key；跨 provider 时应该显式写出角色级 model、host 和 Key，避免 endpoint 混淆。

Embedding 不能直接填聊天模型名。只有在 上游 API 确实提供 Embedding endpoint，并且你知道向量维度、模型支持的语言、前缀和计费方式时，才把 Embedding 也接入网关：

~~~env
EMBEDDING_BINDING=openai
EMBEDDING_MODEL=<embedding-model>
EMBEDDING_DIM=<embedding-dimension>
EMBEDDING_BINDING_HOST=https://<provider-endpoint>/v1
EMBEDDING_BINDING_API_KEY=<embedding-key>
~~~

如果 上游 API 只代理聊天模型，Embedding 可以走独立的本地服务或其他合规 provider。关键是索引阶段和查询阶段必须使用同一套 Embedding 语义配置。

### 2. 不要让所有阶段共用一个 Key

最小可用时共用一个 Key 没问题，但企业长期运行建议按下面方式拆分：

~~~text
项目 / 环境：dev、staging、prod
角色：EXTRACT、KEYWORD、QUERY、VLM、EMBEDDING、RERANK
预算：月度额度、日额度、单文件额度、单查询额度
并发：每个角色的 MAX_ASYNC_LLM 和网关 RPM/TPM
审计：request_id、workspace、模型、token、费用和错误
~~~

这样做的直接好处是：VLM 费用异常时，不会和普通问答混在一起；某个项目批量上传大文件时，不会让所有项目的 QUERY 请求都被限流；更换抽取模型时，也能单独比较 EXTRACT 的失败率和实体关系质量。

### 3. 估算知识库的真实费用

知识库的成本可以粗略拆成：

~~~text
建库成本
= 文本解析相关调用
 + EXTRACT 的实体、关系和摘要调用
 + 文本块、实体、关系的 Embedding
 + 可选的 VLM 图片/表格/公式分析

单次查询成本
= KEYWORD
 + 召回后的可选 Rerank
 + QUERY

变更成本
= 删除后的受影响关系重建
 + 解析器变更后的重新处理
 + Embedding 变更后的全量重建
~~~

上游 API 或企业网关至少需要给每一种请求打上 `stage` 标签，例如 `lightrag.extract`、`lightrag.keyword`、`lightrag.query`、`lightrag.vlm`。每月统计：

- 每个 workspace 上传了多少文件、多少页或多少字符。
- 每个文件触发了多少次 EXTRACT、Embedding 和 VLM。
- 每种查询模式的查询次数、输入 token、输出 token 和延迟。
- 哪些请求发生重试、超时、429 或上下文过长。
- 哪些项目只查询不更新，哪些项目频繁删除和重建。

不要只看模型账单总额。一个看起来查询量不大的项目，如果每天重复重建知识库，可能比正常问答消耗更多预算。

## 十、企业上线前的推荐顺序

不要一次打开所有能力。可以按四个阶段推进：

### 阶段一：单实例验收

~~~text
Docker 或 Linux 服务启动
  -> 配置 LLM 和 Embedding
  -> 上传一份小型 Markdown/PDF
  -> 跟踪 track_id 到 processed
  -> 用 /query、/query/stream、/query/data 对照结果
~~~

### 阶段二：安全封闭

~~~text
绑定内网地址
  -> 配置 LIGHTRAG_API_KEY
  -> 配置 AUTH_ACCOUNTS 和 TOKEN_SECRET
  -> 收窄 WHITELIST_PATHS
  -> Nginx 配置 TLS、上传大小和流式超时
  -> 确认匿名请求看不到文档和运行配置
~~~

### 阶段三：外部持久化和恢复

~~~text
选择 PostgreSQL / MongoDB / OpenSearch 或专业向量/图后端
  -> 明确 workspace 和命名空间
  -> 备份原文、派生数据和版本配置
  -> 在新 workspace 做恢复演练
  -> 用固定问题集对比迁移前后召回
~~~

### 阶段四：模型网关治理

~~~text
上游 API 或企业网关统一入口
  -> EXTRACT / KEYWORD / QUERY / VLM 分角色路由
  -> 记录 token、延迟、重试和费用
  -> 限制上传、并发和单项目预算
  -> 发现异常时按角色切换或熔断
~~~

每个阶段通过后再进入下一阶段。否则当查询答案变差时，你无法判断是解析器、Embedding、Graph、QUERY 模型、反向代理还是权限路由出了问题。

## 十一、安全边界和当前版本注意事项

最后把几条最容易被忽略的边界集中列出来：

- LightRAG 默认无认证，公网或内网开放前必须显式配置 API Key、JWT 或上游认证。
- `WHITELIST_PATHS` 的 `/api/*` 兼容配置可能让 Ollama 路径保持免认证，不需要时应移除。
- `workspace` 是数据命名空间，不是完整的用户权限系统。
- API 网关的 Key、JWT Secret、数据库密码和原始文档不能出现在 Git、浏览器、日志和示例的真实值中。
- 知识图谱中的实体和关系来自模型抽取，不是人工事实库，合同、财务、医疗和合规结论必须回到原文和人工流程。
- 默认 JSON、NetworkX 和本地向量存储适合开发调试，不应直接当成高可用生产存储。
- 更换 Embedding、维度、非对称配置或解析器路由前，先做备份、迁移演练和固定问题集回归。
- 上游 API 负责模型入口、路由、额度和审计，不替代 LightRAG 的文档权限、来源治理和业务合规。
- LightRAG、Docker 镜像、解析器和存储后端都可能随版本变化，生产升级要固定版本、保留回滚路径并重新跑验收集。

## 十二、企业级上线验收清单

### 部署验收

- [ ] 生产运行在 Linux 或 Docker Linux 容器，Windows 没有误用 Gunicorn。
- [ ] LightRAG、Nginx、上游 API 和模型服务的监听地址与端口已记录。
- [ ] 镜像或源码版本已固定，升级前保存了当前版本和变更说明。
- [ ] `WORKING_DIR`、`INPUT_DIR`、外部数据库和日志目录都有持久化策略。
- [ ] Docker 挂载目录对 uid/gid 1000 可读写，容器没有不必要的特权。

### 代理和 API 验收

- [ ] `/documents/upload` 的上传大小、读取超时和应用限制经过大文件测试。
- [ ] `/query/stream`、`/api/chat` 或 `/api/generate` 的流式响应没有被代理缓冲。
- [ ] `LIGHTRAG_API_PREFIX` 与 Nginx/Ingress 的前缀剥离方式一致。
- [ ] `/health` 既能被 liveness probe 调用，又不会向匿名访问者泄露完整配置。
- [ ] `track_id`、`/documents/track_status/{track_id}` 和 `/documents/pipeline_status` 能被业务系统正确消费。

### 权限验收

- [ ] 已配置 `LIGHTRAG_API_KEY` 或 JWT，生产没有匿名写入接口。
- [ ] WebUI 账号、API Key、上游网关和 workspace 路由分别测试过。
- [ ] `WHITELIST_PATHS` 已收窄，不需要 Ollama 兼容接口时没有保留 `/api/*`。
- [ ] 不同 workspace 的上传、查询、图谱和引用不会互相泄漏。
- [ ] CORS 仅允许明确来源，密钥不会进入前端代码、URL 或日志。

### 存储和恢复验收

- [ ] KV、Vector、Graph、Doc Status 四类存储的实现和备份范围已记录。
- [ ] 已备份原始文件、sidecar、提示词、实体类型配置和版本信息。
- [ ] 删除文档、失败重试和重处理分别做过测试，并有操作审计。
- [ ] 更换 Embedding 或解析器前有新 workspace 重建和回滚方案。
- [ ] 已完成一次隔离环境恢复演练，并记录恢复耗时和缺失项。

### 成本和质量验收

- [ ] EXTRACT、KEYWORD、QUERY、VLM、Embedding 和 Rerank 的模型、Key、并发和预算已拆分。
- [ ] 上游 API 或企业网关能按 workspace、角色、模型和 request_id 统计调用。
- [ ] 上传、重处理、删除重建和查询的 token、费用、延迟和失败率都有记录。
- [ ] 用固定问题集比较 `naive`、`local`、`global`、`hybrid`、`mix` 的质量和成本。
- [ ] 对合同、财务、医疗和合规场景设置了人工复核，不把模型答案直接当成业务事实。

## 总结

LightRAG 生产化的核心，不是把所有参数调到最大，而是建立一条可以解释、可以隔离、可以恢复的链路：

~~~text
原始文档可追溯
  -> 异步处理有 track_id
  -> 图、向量、KV 和状态可观测
  -> workspace 和 API 权限不串库
  -> 模型按角色通过 上游 API 治理
  -> 费用、延迟、失败和引用可以审计
  -> 更换配置前能重建，出问题后能回滚
~~~

如果你只需要个人知识库，先用第306期的 Docker 或本地方式跑通；如果要接入客服、Agent、企业搜索或内部 SaaS，就从认证、workspace、外部存储、备份和 上游 API 角色预算开始设计。这样 LightRAG 才不只是一次漂亮的 RAG 演示，而是一项可长期维护的基础服务。

## 结论

本文给出了问题定位、配置或创作流程的可执行路径。实际结果仍取决于当前版本、权限和运行环境，提交前应按官方文档复核可变字段，并保留失败证据和回滚边界。
