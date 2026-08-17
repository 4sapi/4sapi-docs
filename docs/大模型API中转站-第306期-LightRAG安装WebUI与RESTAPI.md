---
title: "LightRAG 如何用 Docker 跑通知识库服务"
tags:
  - LightRAG
  - Docker
  - 知识库
description: "知识库服务的第一步不是追求复杂架构，而是让一个最小文档能够被写入、索引和查询。"
---
# LightRAG 如何用 Docker 跑通知识库服务
知识库服务的第一步不是追求复杂架构，而是让一个最小文档能够被写入、索引和查询。本文从 LightRAG 的 Docker 启动、环境变量和 WebUI/REST API 入口开始，给出健康检查、最小数据集和失败定位步骤，并把版本差异留给当前文档复核。文中只讨论可复现的步骤，不把单次结果扩展成产品承诺；每个结论都标注前提、证据和无法覆盖的边界。读者可以先完成最小验证，再按自己的版本、权限和数据补充实验。


做企业知识库时，最容易被低估的不是模型，而是资料会一直变化。

今天上传一批制度文件，明天换一版产品手册，后天又删掉一份已经失效的合同。你不能每次改一个 PDF，就把整个向量库推倒重来；也不能只把文字切成若干块，然后期待模型自己理解“甲方”和“项目 A”之间的关系。

LightRAG 想解决的就是这类问题。

它把文本块、向量、实体、关系和知识图谱放在同一条 RAG 工作流里，并提供 WebUI 和 REST API。你可以用浏览器上传文档和看图谱，也可以让自己的客服系统、Agent、SaaS 后台通过 API 调用它。

如果企业已经通过 上游 API 或其他企业级 API 网关统一接入模型，LightRAG 可以把这个入口用于抽取、查询和多模态分析；但网关只负责模型路由、Key 权限和成本治理，不能替你解决资料权限和事实核验。

这一篇先完成最小闭环：装起来、接上模型、上传文件、发出第一条查询，再把它部署到不容易误开公网的状态。

## 一、LightRAG 到底是什么

项目地址：[HKUDS/LightRAG](https://github.com/hkuds/lightrag)

LightRAG 是一个轻量级、基于知识图谱的 RAG 框架。仓库 README 把它描述为传统向量 RAG 和图 RAG 之间的结合方案，核心不是把向量检索删掉，而是同时管理：

~~~text
文本块：保留原始资料的局部上下文
Embedding：用于相似度召回
实体：人物、产品、组织、概念等节点
关系：实体之间的连接和描述
文档状态：记录哪些文件正在处理、成功或失败
LLM 缓存：减少重复抽取和查询的成本
~~~

普通向量 RAG 很擅长回答“哪一段文字和问题最相似”。但遇到“这家公司在不同年份为什么改变战略”“多个产品之间有什么关系”“这份制度影响了哪些部门”这类问题，只有相似度往往不够。

LightRAG 会在文档插入阶段通过 LLM 做实体和关系抽取，再把这些结果和向量化文本保存下来。查询阶段可以只看局部实体，也可以沿着跨文档关系看更大的主题。第307期会专门拆这条检索链路，这里先把它当成一个可操作的知识库服务。

## 二、先选安装路线

当前仓库给了三条比较清楚的入口。

### 路线一，PyPI 工具安装

只想快速启动服务，可以安装 API extra：

~~~bash
uv tool install "lightrag-hku[api]"
~~~

也可以使用虚拟环境和 pip：

~~~bash
python -m venv .venv
source .venv/bin/activate
pip install "lightrag-hku[api]"
~~~

Windows 的激活命令是：

~~~powershell
python -m venv .venv
.\.venv\Scripts\Activate.ps1
python -m pip install "lightrag-hku[api]"
~~~

这种方式适合先验证服务是否能启动。若你要修改源码、构建 WebUI 或研究内部实现，直接走源码路线更省返工。

### 路线二，源码安装

~~~bash
git clone https://github.com/HKUDS/LightRAG.git
cd LightRAG
uv sync --extra test --extra offline
source .venv/bin/activate
cd lightrag_webui
bun install --frozen-lockfile
bun run build
cd ..
~~~

仓库也提供 `make dev`，它会安装测试工具链和完整的离线栈，并构建前端。Windows 环境不一定自带 make，因此可以用上面的 uv 手动步骤，或者在已经准备好 Make 的环境里按仓库说明执行。

### 路线三，Docker Compose

适合想把服务和依赖放进一个可重复环境的人：

~~~bash
git clone https://github.com/HKUDS/LightRAG.git
cd LightRAG
cp env.example .env
docker compose up
~~~

Windows PowerShell 可以把复制命令写成：

~~~powershell
Copy-Item .\env.example .\.env
docker compose up
~~~

首次启动不要直接把服务端口映射到公网。当前 README 明确提醒，服务默认绑定 `0.0.0.0`，如果没有配置认证，接口就是公开的。先在本机验证，再决定是否通过反向代理开放。

## 三、第一次配置要填什么

LightRAG 至少需要一个 LLM 和一个 Embedding 模型，才能完成文档索引和查询。仓库支持 Ollama、OpenAI 或 OpenAI-compatible、Azure OpenAI、Bedrock、Gemini 等 LLM 后端，也支持多种 Embedding 后端。

一个最小的 OpenAI-compatible 配置可以参考：

~~~env
LLM_BINDING=openai
LLM_MODEL=<your-chat-model>
LLM_BINDING_HOST=<your-openai-compatible-endpoint>
LLM_BINDING_API_KEY=<your-llm-key>

EMBEDDING_BINDING=ollama
EMBEDDING_BINDING_HOST=http://localhost:11434
EMBEDDING_MODEL=bge-m3:latest
EMBEDDING_DIM=1024
~~~

如果你的 Embedding 也走兼容接口，可以把 `EMBEDDING_BINDING` 改成仓库支持的 OpenAI-compatible 配置，并按当前 `env.example` 填对应模型和维度。

不要把 `<your-openai-compatible-endpoint>` 写成一个想当然的地址。使用 上游 API 或其他中转服务时，endpoint、模型标识、Key 头和计费口径以服务商当前文档为准。LightRAG 只知道自己要调用一个兼容的模型接口，不知道中转层是否支持某个具体模型参数。

### 为什么 Embedding 不能随便换

Embedding 模型决定向量维度和语义空间。仓库明确要求，文档索引阶段和查询阶段使用相同的模型及相关非对称配置。

如果已经处理过文件，再直接换模型，旧向量和新查询向量可能无法比较。有些存储后端还会在首次建表时固定向量维度。LightRAG 当前没有提供一键重新 Embedding 的工具，通常需要清理受影响的向量数据并重新索引。

这件事应该在第一次上传文件前定下来。

### `.env` 放在哪里

LightRAG Server 会从启动目录读取 `.env`，而不是任意目录里的同名文件。系统环境变量优先级高于 `.env`。所以改完配置后要重新打开终端，避免旧环境变量继续覆盖新值。

项目提供了交互式配置命令：

~~~bash
make env-base
make env-storage
make env-server
make env-security-check
~~~

其中 `env-base` 配置模型，`env-storage` 配置数据库服务，`env-server` 配置端口、认证和 SSL，`env-security-check` 用于审计当前 `.env`。这些命令只负责配置，不会替你判断一套生产架构是否合理。

## 四、启动 Server、WebUI 和 API 文档

服务启动命令：

~~~bash
lightrag-server
~~~

常用参数包括：

~~~bash
lightrag-server --host 127.0.0.1 --port 9621 --working-dir ./rag_storage
~~~

本地验证时推荐绑定 `127.0.0.1`。默认端口是 `9621`，WebUI 挂载在 `/webui`，Swagger UI 在：

~~~text
http://localhost:9621/webui/
http://localhost:9621/docs
http://localhost:9621/redoc
~~~

先访问健康检查：

~~~bash
curl http://localhost:9621/health
~~~

认证开启以后，未认证的健康检查只返回有限的存活信息；带上合法的 `X-API-Key` 或 JWT，才会返回更完整的配置和运行状态。

## 五、上传、扫描和查询的最小闭环

LightRAG Server 不是启动以后自动扫描任意目录。当前 API 提供了文档上传、文本插入、批量文本插入和扫描等路径。

常见流程是：

~~~text
上传文件或插入文本
  -> 返回 track_id
  -> 后台解析、分块、抽取实体关系、向量化
  -> 查询 track_status/{track_id}
  -> processed 后调用 /query
~~~

文件上传入口是 `/documents/upload`，文本入口是 `/documents/text` 或 `/documents/texts`。如果你把文件直接放进 `inputs` 目录，可以调用 `/documents/scan` 触发扫描。上传和文本插入会返回任务追踪 ID，可以通过：

~~~text
/documents/track_status/{track_id}
~~~

查看 `pending`、`processing`、`processed` 或 `failed` 状态。

查询 API 的最小请求可以写成：

~~~bash
curl -X POST "http://localhost:9621/query" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "这批资料的核心主题是什么？",
    "mode": "mix",
    "include_references": true
  }'
~~~

`/query` 返回回答和引用列表；`/query/stream` 返回流式 NDJSON；`/query/data` 只返回结构化检索数据，不调用最终回答模型，适合调试召回结果、做评测或自己接一层生成逻辑。

如果开启 API Key，需要在请求中加入：

~~~bash
curl -X POST "http://localhost:9621/query" \
  -H "X-API-Key: <lightrag-api-key>" \
  -H "Content-Type: application/json" \
  -d '{"query":"请总结这批资料","mode":"mix"}'
~~~

复杂请求不要手写到生产脚本里猜字段，直接打开 `/docs` 看当前版本的 OpenAPI schema。仓库接口会随着版本增加引用内容、进度事件和多模态字段，Swagger 是比旧博客更可靠的现场说明。

## 六、REST API 还是 Core SDK

仓库文档明确建议，想把 LightRAG 集成到自己的项目里，优先使用 LightRAG Server 的 REST API。Core SDK 更适合嵌入式应用、研究和评测。

如果只是做一个知识库服务，REST API 有几个好处：

- 前端、业务服务和 RAG 核心可以分开部署。
- 模型、存储、工作目录和认证集中在 Server 配置中。
- 业务系统只需要调用上传、任务状态、查询和图谱接口。
- 可以通过企业 API 网关统一做 Key、日志、超时和成本统计。

Core SDK 适合需要直接控制 Python 对象的人。最小示例需要显式初始化存储：

~~~python
import asyncio
from lightrag import LightRAG, QueryParam
from lightrag.llm.openai import gpt_4o_mini_complete, openai_embed

async def main():
    rag = LightRAG(
        working_dir="./rag_storage",
        embedding_func=openai_embed,
        llm_model_func=gpt_4o_mini_complete,
    )
    await rag.initialize_storages()
    try:
        await rag.ainsert("LightRAG 的测试文本")
        answer = await rag.aquery(
            "这段文本讲了什么？",
            param=QueryParam(mode="hybrid"),
        )
        print(answer)
    finally:
        await rag.finalize_storages()

asyncio.run(main())
~~~

`initialize_storages()` 不能省略，`finalize_storages()` 也应当在长期运行或测试结束时调用。很多“明明装好了却报初始化错误”的问题，根源不是模型，而是漏了这两个生命周期步骤。

## 七、企业级大模型接入怎么放进来

一个比较稳的调用链路是：

~~~text
业务系统 / Agent / 知识库前端
  -> 企业 API 网关或 上游 API
  -> LightRAG Server 配置的 LLM / Embedding / Reranker
  -> LightRAG 存储与文档处理流水线
~~~

上游 API 或其他网关可以承担：

- 多模型统一入口。
- 项目、团队和环境级 API Key。
- 模型路由和失败告警。
- 调用量、Token、耗时和费用统计。
- 对外业务服务和内部 LightRAG 服务之间的访问控制。

LightRAG 自己还要负责：

- 文档来源和处理状态。
- 实体关系抽取质量。
- 工作区和存储隔离。
- 引用、召回结果和删除行为。
- 输入文件的权限与保留周期。

不要把网关地址、API Key 或数据库密码写进文章里的真实配置。生产部署使用 Secret 管理，日志只记录模型、项目、耗时、错误类型和费用，不记录完整文档正文。

## 八、安全边界和当前版本注意事项

当前 Server 默认可以无认证访问。正式部署至少要考虑：

- 本机开发先绑定 `127.0.0.1`，不要默认暴露 `0.0.0.0`。
- API Key 通过 `X-API-Key` 传递，不要放 URL、前端代码或日志。
- WebUI 账号使用 `AUTH_ACCOUNTS` 和 `TOKEN_SECRET` 配置 JWT。
- 需要同时保护 WebUI 和 API 时，不要只配置 API Key 而让 WebUI 仍以 Guest 访问。
- 不需要 Ollama 兼容接口时，收紧 `WHITELIST_PATHS`，不要保留过宽的 `/api/*`。
- 反向代理上传大文件时配置 `client_max_body_size`，否则 Nginx 可能在文件到达 LightRAG 之前就返回 413。

仓库当前文档、参数和接口会继续变化。本文对应的是检出时的仓库行为，实际安装前应再看当前 README、`env.example` 和 `/docs`。

## 九、首次上线验收清单

### 启动验收

- [ ] LLM 和 Embedding 都能单独连通。
- [ ] `.env` 位于启动目录，系统环境变量没有覆盖旧配置。
- [ ] `/health`、`/docs`、`/webui/` 可以按预期访问。
- [ ] 本地开发服务没有误绑定到公网网卡。

### 文档验收

- [ ] 上传文件能拿到 `track_id`。
- [ ] `track_status` 最终进入 `processed`，失败时能看到错误原因。
- [ ] `/query` 能返回回答和引用。
- [ ] `/query/data` 能看到实体、关系、文本块和引用结构。
- [ ] 使用 上游 API 或其他模型网关时，模型名、endpoint 和计费口径都经过当前服务商文档确认。

### 安全验收

- [ ] 已配置 API Key 或 JWT，生产环境没有匿名写入接口。
- [ ] `TOKEN_SECRET`、API Key、数据库密码和模型凭证没有提交 Git。
- [ ] `WHITELIST_PATHS` 只开放真正需要的路径。
- [ ] 上传大小、输入目录和工作目录有磁盘与权限限制。
- [ ] 日志没有保存文档原文和敏感用户数据。

## 总结

LightRAG 的第一次使用不需要一开始就上 PostgreSQL、Neo4j、MinerU 和多站点反向代理。

更稳的顺序是：

~~~text
本机启动 Server
  -> 配好 LLM 和 Embedding
  -> 上传一份小文档
  -> 等待 track_id 处理完成
  -> 用 /query 和 /query/data 对照结果
  -> 再接 上游 API、认证、外部存储和企业网关
~~~

先证明最小链路能回答问题，再扩大架构。否则一旦结果不对，你很难判断是模型、解析、向量、图谱还是网络层出了问题。

## 结论

本文给出了问题定位、配置或创作流程的可执行路径。实际结果仍取决于当前版本、权限和运行环境，提交前应按官方文档复核可变字段，并保留失败证据和回滚边界。
