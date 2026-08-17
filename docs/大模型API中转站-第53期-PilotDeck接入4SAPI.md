---
title: "【大模型API中转站】第53期 PilotDeck接入4SAPI | 路由省钱实战"
category: 人工智能
tags:
  - 大模型API中转站
  - PilotDeck
  - Agent
  - Smart Routing
  - 4SAPI
  - OpenAI-Compatible
description: "基于 PilotDeck 官方 README 和 Docker 文档，拆解 WorkSpace、白盒记忆、Smart Routing 与 Always-on，并给出 4SAPI 接入配置、Docker 部署、模型分层和避坑清单。"
---

# 【大模型API中转站】第53期 PilotDeck接入4SAPI | 路由省钱实战

本文是【大模型API中转站】系列的第53篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一篇我们讲了如何搭一个个人知识库。

但知识库只是第一步。

真正进入日常工作之后，你会发现还有几个更麻烦的问题：

```text
多个项目同时跑，记忆会不会串？
一个 Agent 长时间工作，过程能不能追踪？
简单任务和复杂任务能不能自动用不同模型？
我离开电脑后，它能不能继续处理后台任务？
每个项目消耗了多少 Token，能不能算清楚？
```

这就是 PilotDeck 这类工具想解决的问题。

它不是普通聊天界面，也不是单纯的代码编辑器插件。

更准确地说，PilotDeck 是一个围绕 WorkSpace 设计的 Agent 工作台：每个项目有自己的文件、记忆、技能和任务上下文。

如果再把 4SAPI 这类大模型API中转站接到中间，就可以把模型、Key、日志和成本统一管理起来。

这一篇直接讲实战：

```text
PilotDeck 是什么
为什么适合接 4SAPI
Docker 怎么跑
pilotdeck.yaml 怎么配
Smart Routing 怎么省钱
哪些坑要提前避开
```

## 1. PilotDeck 是什么

根据官方 README，PilotDeck 的定位是：

```text
Task-oriented AI Agent productivity platform
```

它围绕 WorkSpace 做 Agent 操作系统。

核心不是“一次问答”，而是“长期项目”。

官方强调三类能力：

| 能力 | 解决的问题 |
| --- | --- |
| WorkSpace 级隔离 | 每个项目有独立文件、记忆和技能，避免项目之间互相污染 |
| 白盒记忆 | 记忆的生成、存储、检索可以看见，错了可以定位和修改 |
| Smart Routing | 根据任务难度选择不同模型，简单任务不用一直烧强模型 |
| Always-on | 用户离开后，Agent 仍可继续后台执行、发现任务、落地产物 |

另外，PilotDeck 还原生支持 MCP，也就是 Model Context Protocol。

这意味着它不只是模型调用器，还可以接更多工具、技能和外部上下文。

所以它适合的场景不是：

```text
帮我解释一句话。
帮我写一段文案。
帮我翻译一小段内容。
```

而是：

```text
帮我维护一个长期内容项目。
帮我在多个资料夹里生成报告。
帮我持续跟进一个代码仓库。
帮我整理一个知识库并产出可发布文章。
帮团队把 Agent 的记忆、成本和任务结果管起来。
```

一句话说：

```text
Chatbot 解决一次对话，PilotDeck 更像一个多项目 Agent 工作台。
```

## 2. 为什么 PilotDeck 适合接 4SAPI

PilotDeck 官方配置里已经支持 OpenAI、Anthropic、DeepSeek、Qwen、Kimi、MiniMax 以及其他 OpenAI-compatible endpoints。

这就意味着，只要一个模型服务兼容 OpenAI 接口，理论上就可以通过 provider 配进去。

4SAPI 的接入点也是 OpenAI 兼容形式：

```text
Base URL: https://4sapi.com/v1
API Key: 4SAPI 后台生成
Model: 从模型列表复制完整模型名
```

两者结合之后，结构大概是：

```text
PilotDeck WorkSpace
  -> PilotDeck Gateway
  -> 4SAPI
  -> GPT / Claude / DeepSeek / Qwen / Kimi / 其他模型
```

这样做有几个好处。

第一，Key 不散。

你不需要在每个工具里分别放一堆模型厂商的 Key，只要在 4SAPI 侧管理模型入口，再让 PilotDeck 指向 4SAPI。

第二，成本可控。

PilotDeck 有 Smart Routing 的思路，4SAPI 负责模型池和费用统计。简单任务走便宜模型，复杂任务走强模型，能更容易看清钱花在哪里。

第三，模型更容易替换。

如果某个模型临时不可用，或者价格不合适，可以在 4SAPI 后台换模型，再同步调整 PilotDeck 配置。

第四，适合团队试点。

团队可以按项目、账号、用途拆 Key：

```text
pilotdeck-local-test
pilotdeck-content-team
pilotdeck-code-review
pilotdeck-research
```

一旦调用异常，就能定位到具体用途。

## 3. 适合先跑的三个场景

不要一上来就把所有工作都交给 PilotDeck。

更稳的方式是先选一个真实场景跑通。

### 场景一：个人知识库工作台

上一期我们搭了个人知识库。

现在可以让 PilotDeck 管一个知识库 WorkSpace：

```text
personal-kb/
  sources/
  notes/
  outputs/
  prompts/
```

它负责：

```text
扫描新增资料
整理摘要
生成标签
把高价值资料变成 notes
按选题生成文章大纲
输出可追溯引用清单
```

4SAPI 负责给不同步骤分配模型。

### 场景二：内容团队生产线

对公众号、知乎、小红书、产品号团队来说，最怕的是素材、口吻、选题和复盘分散。

PilotDeck 可以按栏目建 WorkSpace：

```text
wechat-longform/
xiaohongshu-cards/
product-updates/
course-content/
```

每个 WorkSpace 都有自己的资料、记忆和技能。

这样“知识付费账号”的话术不会混进“技术测评账号”，项目边界更清楚。

### 场景三：开发和代码仓库维护

如果你有多个代码项目，也可以一个仓库一个 WorkSpace：

```text
app-frontend/
api-service/
docs-site/
internal-tools/
```

让 Agent 做：

```text
读 README
理解项目结构
生成任务计划
修改代码
运行测试
写变更总结
输出风险点
```

这种任务更适合 PilotDeck，而不是每次开一个全新的临时对话。

## 4. 部署方式选择

PilotDeck 官方给了三种启动方式：

| 方式 | 适合谁 | 说明 |
| --- | --- | --- |
| 一键脚本 | macOS / Linux 用户 | 官方脚本会安装 Node.js 22、克隆仓库、安装依赖并构建前端 |
| 源码启动 | 开发者 | 适合想调试 UI、Gateway 或贡献代码的人 |
| Docker Compose | 大多数试用者和团队 | 配置清晰，状态目录和 WorkSpace 挂载更容易管理 |

如果你只是想接 4SAPI 跑起来，建议先用 Docker Compose。

原因很简单：

```text
环境更干净
端口更明确
配置更容易复现
团队试点更好迁移
```

官方 Docker 文档里说明，容器内有两个 Node.js 进程：

```text
Gateway: 默认端口 18789
UI Server: 默认端口 3001
```

浏览器访问的是：

```text
http://localhost:3001
```

## 5. 方案一：Docker 接入 4SAPI

先拉代码：

```bash
git clone https://github.com/OpenBMB/PilotDeck.git
cd PilotDeck
```

如果你用 Docker Compose，可以在 `docker-compose.yml` 里配置环境变量。

核心是三项：

```env
PILOTDECK_MODEL=sapi4/你的模型名
PILOTDECK_API_KEY=sk-your-4sapi-key
PILOTDECK_API_URL=https://4sapi.com/v1
```

例如：

```yaml
services:
  pilotdeck:
    build: .
    image: pilotdeck:latest
    ports:
      - "3001:3001"
    volumes:
      - pilotdeck-home:/root/.pilotdeck
      - ${PILOTDECK_WORKSPACE:-${PWD}}:/workspace
    environment:
      - NODE_ENV=production
      - PILOT_HOME=/root/.pilotdeck
      - SERVER_PORT=3001
      - PILOTDECK_GATEWAY_PORT=18789
      - PILOTDECK_MODEL=sapi4/your-model-name
      - PILOTDECK_API_KEY=sk-your-4sapi-key
      - PILOTDECK_API_URL=https://4sapi.com/v1
    restart: unless-stopped

volumes:
  pilotdeck-home:
```

启动：

```bash
docker compose up -d --build
```

打开：

```text
http://localhost:3001
```

注意这里的模型名不是随便写的。

`sapi4/your-model-name` 里的 `your-model-name` 要换成你在 4SAPI 模型广场里复制的实际模型名。

如果模型名写错，PilotDeck 不一定能给你一个非常友好的错误提示，通常会表现为请求失败、模型不存在或 provider 找不到。

## 6. 方案二：pilotdeck.yaml 接入 4SAPI

如果你不想把 Key 写在 `docker-compose.yml` 里，可以用配置文件。

PilotDeck 默认读取：

```text
~/.pilotdeck/pilotdeck.yaml
```

创建配置文件：

```bash
mkdir -p ~/.pilotdeck
```

写入：

```yaml
schemaVersion: 1

agent:
  model: sapi4/your-model-name

model:
  providers:
    sapi4:
      protocol: openai
      url: https://4sapi.com/v1
      apiKey: sk-your-4sapi-key
      models:
        your-model-name: {}
```

然后在 Docker Compose 里挂载配置：

```yaml
volumes:
  - pilotdeck-home:/root/.pilotdeck
  - ${PILOTDECK_CONFIG:-${HOME}/.pilotdeck/pilotdeck.yaml}:/root/.pilotdeck/pilotdeck.yaml:ro
  - ${PILOTDECK_WORKSPACE:-${PWD}}:/workspace
```

再启动：

```bash
docker compose up -d --build
```

这个方案更适合团队，因为配置文件可以单独管理。

不过有一个重要提醒：

```text
不要把真实 API Key 提交到 Git 仓库。
```

如果你必须共享配置，建议提交模板：

```yaml
schemaVersion: 1

agent:
  model: sapi4/your-model-name

model:
  providers:
    sapi4:
      protocol: openai
      url: https://4sapi.com/v1
      apiKey: ${PILOTDECK_API_KEY}
      models:
        your-model-name: {}
```

真实 Key 放在本机环境变量、密钥管理工具或部署平台里。

## 7. Smart Routing 怎么和 4SAPI 配合

PilotDeck 的一个核心亮点是 Smart Routing。

官方 README 里给了两个对比样例：

| 场景 | 官方样例结果 |
| --- | --- |
| 小红书式运营任务 | 开启 Smart Routing 后，成本从全强模型方案的 $12.58 降到 $2.83 |
| 复杂任务 benchmark | 强主模型 + 轻子模型方案，用 $3.15 达到接近或超过单强模型 $18.36 的效果 |

这里要注意：

```text
这是官方样例，不等于你的实际项目一定能省同样比例。
```

但思路是成立的。

大多数 Agent 工作流里，并不是每一步都需要强模型。

例如内容生产：

| 任务 | 推荐模型策略 |
| --- | --- |
| 文件扫描、重命名、标签 | 低成本模型 |
| 资料摘要、信息抽取 | 中等模型 |
| 选题判断、结构设计 | 强推理模型 |
| 正文初稿 | 强文本模型或中高端模型 |
| 格式整理、标题备选 | 低成本到中等模型 |
| 最终审核 | 强模型 + 人工确认 |

例如代码任务：

| 任务 | 推荐模型策略 |
| --- | --- |
| 读目录、列文件 | 低成本模型 |
| 搜索引用、整理上下文 | 低成本模型或代码模型 |
| 修改核心逻辑 | 强代码模型 |
| 写测试 | 中高端代码模型 |
| Review 风险 | 强推理模型 |
| 生成变更总结 | 中等文本模型 |

如果 PilotDeck 支持用轻量模型变量配置路由，可以在 Docker 环境变量里加：

```env
PILOTDECK_MODEL=sapi4/strong-model-name
PILOTDECK_LIGHT_MODEL=sapi4/light-model-name
PILOTDECK_API_KEY=sk-your-4sapi-key
PILOTDECK_API_URL=https://4sapi.com/v1
PILOTDECK_LIGHT_API_KEY=sk-your-4sapi-key
PILOTDECK_LIGHT_API_URL=https://4sapi.com/v1
```

同一个 4SAPI Key 也可以同时跑主模型和轻量模型。

更细的团队玩法是拆 Key：

```text
pilotdeck-main
pilotdeck-light
pilotdeck-test
pilotdeck-review
```

这样你可以在 4SAPI 后台按用途看调用和成本。

## 8. WorkSpace 挂载要特别注意

Docker 方式有一个容易被忽略的点：

```text
Agent 运行在容器里，不在你的宿主机目录里。
```

如果你希望 PilotDeck 能访问某个项目，需要把它挂载到容器的 `/workspace`。

示例：

```yaml
volumes:
  - pilotdeck-home:/root/.pilotdeck
  - D:/projects/my-kb:/workspace
```

Linux 或 macOS 示例：

```yaml
volumes:
  - pilotdeck-home:/root/.pilotdeck
  - /Users/you/projects/my-kb:/workspace
```

试点阶段建议只挂载一个干净项目，不要直接挂整个用户目录。

推荐：

```text
只挂载当前项目目录
不挂载 Desktop、Downloads、Documents 根目录
敏感文件放到 .pilotdeckignore 或单独移出
输出文件统一写入 outputs/
```

如果你把整个硬盘暴露给 Agent，再强的权限系统也会变得很难管理。

## 9. 建议先写一份 WorkSpace 规则

PilotDeck 的优势是 WorkSpace 级隔离。

但隔离只是基础，项目规则才决定它怎么工作。

建议每个 WorkSpace 里放一份说明：

```text
AGENTS.md
```

例如内容项目：

```markdown
# WorkSpace Rules

## Scope

- 本 WorkSpace 用于公众号长文选题、资料整理和成稿。
- 原始资料放在 `sources/`。
- 笔记放在 `notes/`。
- 成稿放在 `outputs/`。

## Model Usage

- 文件扫描、标签、短摘要优先使用轻量模型。
- 结构设计、事实核对、最终润色使用主模型。
- 不允许把客户隐私、合同、API Key 放进模型上下文。

## Output Rules

- 所有文章必须列出参考资料。
- 不确定事实要标注“需人工确认”。
- 修改文件前先说明计划，修改后列出变更清单。
```

例如代码项目：

```markdown
# WorkSpace Rules

## Safety

- 不自动删除文件。
- 不修改 `.env`、密钥、生产配置。
- 修改前先阅读 README 和测试命令。

## Workflow

- 先给计划，再改代码。
- 改完运行测试。
- 最终输出文件列表、测试结果和风险点。
```

PilotDeck 负责执行，规则负责约束。

4SAPI 负责把模型调用变得可管理。

## 10. 第一次测试怎么做

不要第一次就让它处理大项目。

建议建一个测试 WorkSpace：

```text
pilotdeck-demo/
  sources/
    intro.md
  notes/
  outputs/
  AGENTS.md
```

`intro.md` 写一点简单资料：

```markdown
# 测试资料

这是一个用于验证 PilotDeck 接入 4SAPI 的测试项目。
目标是确认模型可以读取资料、生成摘要，并把结果写入 outputs。
```

然后在 PilotDeck 里执行一个小任务：

```text
请阅读 sources/intro.md，生成 300 字摘要，写入 outputs/summary.md。
完成后说明你读取了哪些文件、写入了哪些文件。
```

如果这一步成功，再继续测试：

```text
请基于 sources/intro.md 生成 5 个选题。
请把选题按难度、预计耗时和适合模型分组。
请不要读取 sources 之外的文件。
```

测试重点不是文章写得好不好，而是确认：

```text
模型能调用
文件能读取
输出能写入
路径没有错
权限边界清楚
4SAPI 后台能看到调用记录
```

这六点通了，才值得接真实项目。

## 11. 常见报错和排查

### 1. 模型不存在

最常见原因：

```text
PILOTDECK_MODEL 写错
pilotdeck.yaml 里的 models 名称没对应上
4SAPI 后台模型名复制不完整
```

配置要保持一致：

```yaml
agent:
  model: sapi4/your-model-name

model:
  providers:
    sapi4:
      models:
        your-model-name: {}
```

`sapi4/` 是 provider 前缀，`your-model-name` 才是 models 下面的键。

### 2. 401 或鉴权失败

检查：

```text
API Key 是否复制完整
Key 前后是否有空格
Key 是否有对应模型权限
环境变量是否被 Docker 正确读取
```

建议不要把 Key 写在多个地方。

要么用环境变量，要么用 `pilotdeck.yaml`，不要两边都写不同值。

### 3. 连接失败

检查：

```text
PILOTDECK_API_URL 是否是 https://4sapi.com/v1
容器是否能访问外网
是否需要代理
公司网络是否拦截
```

官方 Docker 文档里支持 `PILOTDECK_PROXY`：

```env
PILOTDECK_PROXY=http://host.docker.internal:7890
```

有代理时再加，不要默认加。

### 4. 找不到项目文件

多半是 WorkSpace 没挂载。

确认 Docker Compose 里有：

```yaml
volumes:
  - ${PILOTDECK_WORKSPACE:-${PWD}}:/workspace
```

或者写绝对路径：

```yaml
volumes:
  - D:/projects/pilotdeck-demo:/workspace
```

然后让 Agent 只在 `/workspace` 里工作。

### 5. 成本突然变高

常见原因：

```text
所有任务都走主模型
长文资料一次性塞入上下文
循环任务没有停止条件
没有区分测试 Key 和正式 Key
```

建议做三件事：

```text
给测试 WorkSpace 单独 Key
设置轻量模型
让每个任务输出调用目的和结果摘要
```

## 12. 成本与安全边界

PilotDeck 是工作台，4SAPI 是模型网关。

但安全不能只靠工具。

建议在团队里明确这几条：

| 风险 | 建议 |
| --- | --- |
| Key 泄露 | 不把真实 Key 提交 Git，不贴到公开文档 |
| 敏感文件暴露 | 只挂载项目目录，不挂载整个用户目录 |
| 模型幻觉 | 重要结论必须列来源和不确定点 |
| 成本失控 | 测试 Key、正式 Key 分开，轻量模型和主模型分开 |
| 权限过大 | 原始资料只读，输出统一放 `outputs/` |
| 记忆污染 | 一个客户、一个项目、一个栏目单独建 WorkSpace |

如果你的 WorkSpace 包含客户资料、合同、财务数据、账号密码、生产 Key，建议先不要接入自动化 Agent。

先做脱敏、分权和只读测试。

## 13. 4SAPI 配置建议

建议按用途建 Key：

```text
pilotdeck-demo
pilotdeck-writing
pilotdeck-code
pilotdeck-research
pilotdeck-team
```

建议按任务分模型：

| 任务 | 模型层级 |
| --- | --- |
| 标签、命名、摘要 | 轻量模型 |
| 长文初稿、教程文章 | 中高端文本模型 |
| 复杂技术判断 | 强推理模型 |
| 代码修改和测试 | 代码模型 |
| 多模态资料理解 | 多模态模型 |
| 最终审核 | 强模型 + 人工 |

建议按项目记账：

```text
content-workspace
kb-workspace
code-review-workspace
client-a-workspace
```

如果 4SAPI 后台支持日志和用量统计，就把这些 Key 名称和 WorkSpace 名称对应起来。

这样一个月后复盘，你能回答三个问题：

```text
哪个项目花钱最多？
哪些任务适合降级到轻量模型？
哪些 WorkSpace 真的产出了成果？
```

这才是中转站对 Agent 工作台的真正价值。

## 14. 最小落地方案

如果你今天只想跑通，不想研究太多，照这个做：

```text
1. 克隆 PilotDeck 仓库
2. 用 Docker Compose 启动
3. 在环境变量里配置 4SAPI Base URL、API Key 和模型名
4. 只挂载一个测试 WorkSpace
5. 让 Agent 读取一个 Markdown，写一个 summary.md
6. 到 4SAPI 后台确认调用记录
7. 再配置轻量模型和主模型分层
```

最小配置：

```env
PILOTDECK_MODEL=sapi4/your-model-name
PILOTDECK_API_KEY=sk-your-4sapi-key
PILOTDECK_API_URL=https://4sapi.com/v1
```

最小测试任务：

```text
请只读取 /workspace/sources/intro.md，
生成摘要到 /workspace/outputs/summary.md，
不要读取其他文件，
完成后列出读写文件清单。
```

最小安全规则：

```text
不挂载敏感目录。
不提交真实 Key。
不让 Agent 自动删除原始文件。
重要输出必须人工确认。
```

做到这里，PilotDeck + 4SAPI 就已经能作为个人或小团队的 Agent 试验台。

## 15. 总结

PilotDeck 适合解决的是长期、多项目、可追踪的 Agent 工作。

它的关键不是多一个聊天入口，而是：

```text
WorkSpace 隔离
白盒记忆
Smart Routing
Always-on 后台执行
MCP 扩展
```

4SAPI 适合放在模型调用层：

```text
统一 Base URL
统一 API Key
统一模型池
统一成本和日志
```

两者结合之后，结构很清楚：

```text
PilotDeck 管工作现场。
4SAPI 管模型入口。
WorkSpace 管项目边界。
规则文件管安全约束。
```

一句话总结：

```text
如果 Codex 更像进入当前项目的工程助手，那么 PilotDeck 更像面向多个 WorkSpace 的 Agent 操作台；接上 4SAPI 后，它才更适合长期跑任务、分模型、控成本。
```

建议先从一个低风险项目开始。

不要一上来把公司资料、客户目录、生产代码全挂进去。

先用一个测试 WorkSpace 跑通读取、生成、写入和成本记录，再逐步接入真实工作流。

## 参考资料

- PilotDeck GitHub：https://github.com/OpenBMB/PilotDeck
- PilotDeck Docker 文档：https://github.com/OpenBMB/PilotDeck/blob/main/README_DOCKER.md
- Docker Compose 配置示例：https://github.com/OpenBMB/PilotDeck/blob/main/docker-compose.yml
- 4SAPI API 地址：https://4sapi.com/v1
