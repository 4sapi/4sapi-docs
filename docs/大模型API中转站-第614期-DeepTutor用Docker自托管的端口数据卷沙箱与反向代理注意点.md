---
title: "DeepTutor 用 Docker 自托管：端口、数据卷、沙箱与反向代理的注意点"
category: 人工智能
tags:
  - Docker 部署
  - 自托管
  - 容器安全
description: "依据 HKUDS/DeepTutor 的容器化文档，说明官方镜像的端口与数据卷、内容工作区挂载、三种沙箱形态的差别、反向代理与拆分部署的配置位置，以及 Codex 登录临时桥接和旧版 Compose 升级的风险。适合要在自己的机器或内网部署这类学习平台的运维和开发者。"
brand_mode: none
---

# DeepTutor 用 Docker 自托管：端口、数据卷、沙箱与反向代理的注意点

DeepTutor 的 README 给了一条 docker run 命令，复制就能起来。但这条命令只是试用路径。真正放到团队机器或内网里，会碰到几个文档里分散说明的问题：只需要开哪个端口，数据放在哪里才不会丢，模型生成的代码跑在哪里，反向代理怎么配，宿主机上的模型服务怎么连。

本文依据仓库 docs-for-dev 里的容器化文档，把这些问题按部署顺序整理出来，并标明每条的来源。我没有运行过任何容器，所有命令来自官方文档，都属于"按文档、待你验证"。信息截止日期为 2026-10-30。

## 一、先弄清镜像里有什么

来源：[容器化文档](https://github.com/HKUDS/DeepTutor/blob/HEAD/docs-for-dev/CONTAINERIZATION.md)（检索于 2026-10-30，文档无日期）。

| 项目 | 内容 |
| --- | --- |
| 镜像 | ghcr.io/hkuds/deeptutor |
| 基础镜像 | python:3.11-slim |
| 进程 | 单个容器里由 supervisord 同时运行 FastAPI 后端（:8001）和 Next.js 前端（:3782） |
| 数据目录 | 容器内 /app/data，存放设置、凭据、数据库、记忆、知识库和日志 |
| 运行用户 | 后端和前端以非 root 用户 deeptutor（默认 UID 1000）运行 |

文档给出三种部署形态，对应不同的安全姿态：

| 形态 | 特点 |
| --- | --- |
| docker run | 最简单。有 root 权限的守护进程，可写根文件系统，一个数据卷 |
| docker compose | 同一镜像，加 PocketBase 边车和沙箱运行器边车 |
| podman compose | 加固路径。无根、只读根文件系统、用 tmpfs 代替可写系统目录 |

## 二、最小命令与端口

文档的默认命令：

```bash
docker run --rm --name deeptutor \
  -p 127.0.0.1:3782:3782 \
  -v deeptutor-data:/app/data \
  ghcr.io/hkuds/deeptutor:latest
```

运行位置：装好 Docker 的机器。预期结果：打开 http://127.0.0.1:3782 能看到界面，首次启动会在 /app/data/user/settings/ 生成配置，模型提供方在网页设置里配。

几个细节：

- **只需要发布 3782。** 浏览器只和前端同源通信，/api/* 和 /ws/* 由容器内的 Next.js 中间件转发到容器里的后端。8001 不用发布到宿主机，只有想直接调 API 或调试时才需要。
- **端口前缀 127.0.0.1。** 这意味着只有本机能访问。去掉前缀会暴露到所有网络接口。对外提供服务之前，先决定谁能访问。
- **换宿主机端口**改 -p 左边的数字即可。如果改了 system.json 里的容器内端口，右边也要跟着改。

## 三、数据卷与权限

### 具名卷与绑定挂载

文档的命令用具名卷 deeptutor-data。设置、密钥、日志、工作区文件、记忆和知识库都在里面，重启后保留。

### NAS 和 Unraid 的属主问题

这类系统的应用数据常归 nobody:users（99:100）或某个共享账户所有，且会在重启时还原容器内的 chown。文档的做法是传 PUID 和 PGID，入口脚本把容器内用户的 UID/GID 重映射成与数据卷一致：

```bash
docker run --rm --name deeptutor \
  -e PUID=99 -e PGID=100 \
  -p 127.0.0.1:3782:3782 \
  -v /mnt/user/appdata/deeptutor:/app/data \
  ghcr.io/hkuds/deeptutor:latest
```

两条保护规则：数据目录仍不可写时，入口脚本会直接退出并提示属主和权限，避免带着错误启动；PUID=0 或 PGID=0 会被拒绝，应用不会以 root 运行，文档明确说不要用 root 运行容器来绕过权限问题。

## 四、内容工作区：让智能体读什么

文档把两个概念分开：

- **应用数据**：私有设置、凭据、数据库、记忆。
- **内容工作区**：智能体可以读的文件夹。

要暴露一个宿主机目录，把它挂到固定路径 /workspace，并锁定部署：

```bash
mkdir -p "$PWD/deeptutor-workspace/outputs"
docker run --rm --name deeptutor \
  -p 127.0.0.1:3782:3782 \
  -v deeptutor-data:/app/data \
  -v "$PWD/deeptutor-workspace:/workspace" \
  -e DEEPTUTOR_WORKSPACE_ROOT=/workspace \
  -e DEEPTUTOR_WORKSPACE_ALLOWED_ROOTS=/workspace \
  ghcr.io/hkuds/deeptutor:latest
```

文档称这是启动期的部署选择，在设置页的 Workspace 里显示为锁定。把应用数据和内容工作区分开挂载，可以让智能体能读的材料不含凭据，这是我从文档的分离设计归纳的。

## 五、沙箱：模型生成的代码跑在哪里

这是部署时最该看清楚的一节。文档对三种形态的描述不同：

| 形态 | 沙箱 | 文档说法 |
| --- | --- | --- |
| docker compose（docker-compose.yml） | 独立的沙箱运行器边车 | 只收到内容树的只读视图，加 /workspace/outputs 的可写覆盖层；收不到 /app/data、用户设置或凭据 |
| 单容器 docker run | 容器内可用的最佳隔离后端 | 在工作区设置里报告实际隔离级别 |
| podman compose | 没有边车 | 主应用回退到 bwrap（Linux，镜像内装有时）或受限子进程后端，由 system.json 的 sandbox_allow_subprocess 控制 |

文档在安全说明里说，边车适合运行不可信的模型生成代码：即使沙箱被逃逸，也只落在一个精简、无特权、没有应用密钥的容器里。这是文档的说法，我没有验证。

对单容器部署，文档给了一个明确建议：如果模型生成的代码不应与密钥共享同一个信任边界，把 sandbox_allow_subprocess 设为 false。

**怎么选：** 只自己用，单容器可以；多人使用或工作区里有他人上传的材料，优先带边车的 compose；要求无根和只读根文件系统，用 podman，但接受没有边车。

## 六、反向代理与拆分部署

### 单容器：几乎不用配

文档说，单容器场景不需要配置 API 地址。浏览器对提供界面的同一个源发相对路径请求，容器内的中间件转发到本地 8001。你只要让反向代理或 TLS 终结指向发布的 3782。

### 拆分部署：才需要配地址

只有后端在另一个容器里时，才需要设置。编辑宿主机上数据卷里的 data/user/settings/system.json：

```json
{
  "next_public_api_base": "http://backend:8001"
}
```

入口脚本每次启动读取它，并导出 DEEPTUTOR_API_BASE_URL 给 proxy.ts 使用。优先级依次是 next_public_api_base、next_public_api_base_external，最后是 http://localhost:8001。因为代理在服务端，这个地址是前端服务器连后端用的，浏览器永远看不到它。

### CORS 用的是源，不是 API 地址

没开认证时，DeepTutor 默认允许常规的 HTTP/HTTPS 浏览器源。开启认证后，要在 system.json 里写准确的前端源：

```json
{
  "cors_origins": ["https://deeptutor.example.com"]
}
```

### 认证怎么开

文档称，auth.json 里的 auth_enabled 设为 true 后，/api/* 和 /ws/* 通过名为 dt_token 的 cookie 保护。proxy.ts 读取入口脚本导出的 DEEPTUTOR_AUTH_ENABLED 来决定是否要求该 cookie。

可选的请求筛查 security_guard（限速、IP 名单、攻击模式扫描）默认关闭，需要 guard 额外依赖，启用前要评估误拦风险。

## 七、改配置的正确方式

这里有一个很容易踩的坑。文档说明：

- 几乎所有可调项都在 data/user/settings/ 下的 JSON 或 YAML 里。
- 入口脚本每次启动都会取消一批相关的环境变量（BACKEND_PORT、FRONTEND_PORT、NEXT_PUBLIC_API_BASE、AUTH_ENABLED、POCKETBASE_URL 等），再从 JSON 重新导出。
- 所以要改 JSON 后重启，不要用 compose 的环境变量去驱动这些项。
- 项目根目录的 .env 被有意忽略，不是应用配置来源。

主要文件：system.json（端口、API 地址、CORS）、auth.json（认证开关、密码哈希、令牌）、integrations.json（PocketBase 与边车）、model_catalog.json（模型与搜索的提供方配置）、interface.json（界面偏好）、main.yaml 与 agents.yaml（运行默认值与各能力参数）。

model_catalog.json 含 API 密钥，意味着数据卷的备份和权限要按凭据对待。

## 八、连接宿主机上的模型服务

容器里的 localhost 指容器自己。要连宿主机上的 Ollama、LM Studio、llama.cpp 等，文档的推荐做法是用主机网关：

```bash
docker run --rm --name deeptutor \
  -p 127.0.0.1:3782:3782 -p 127.0.0.1:8001:8001 \
  --add-host=host.docker.internal:host-gateway \
  -v deeptutor-data:/app/data \
  ghcr.io/hkuds/deeptutor:latest
```

然后在设置的模型页，把提供方地址指向 host.docker.internal。文档的示例：Ollama 对话 http://host.docker.internal:11434/v1，LM Studio http://host.docker.internal:1234/v1，llama.cpp http://host.docker.internal:8080/v1。

Docker Desktop 通常不加 --add-host 也能解析，Linux 上要加。

Linux 上还可以用 --network=host，此时进程直接绑定宿主机网卡。要避免暴露到局域网，文档说要设置 BACKEND_HOST=127.0.0.1 和 FRONTEND_HOST=127.0.0.1。这两个变量只能和 --network=host 配合使用，桥接模式下绑定回环会破坏端口转发。

## 九、两个高风险操作

### Codex 登录的临时桥接

OpenAI Codex 的登录会把浏览器重定向到固定的本机回环端口 1455 或 1457，而容器网络与宿主机回环隔离，所以文档的做法是登录期间临时多发布这两个端口。文档强调三点：只在登录期间使用，完成后改回正常命令；每个映射都绑定 127.0.0.1，绝不暴露到局域网或公网；这个覆盖层发布的是整个前端，不只是回调路径。

凭据保存在持久的 /app/data/system 下，该目录因此需要按敏感数据保护。

### 旧版 ghcr Compose 的升级

文档有一节"一次性迁移"：旧版 docker-compose.ghcr.yml 只绑定挂载了 data 的三个子树（user、memory、knowledge_bases），其余的（system、users、partners、cli-apps）留在容器的可写层里，每次重建都会丢失。现在改成挂载整个 data。

**升级前**如果不把这些目录从运行中的容器用 docker cp 拷出来，宿主机上空的目录会遮住它们，DeepTutor 会重新生成认证密钥（所有人被登出），而且只剩管理员账号。具体命令见文档的迁移一节。

用具名卷或源码构建的 Compose 文件不受影响，因为它们本来就持久化整个目录。

## 十、自己验证的清单

下面没有一项是我跑过的，请在测试环境按顺序做：

1. 起容器，确认 /app/data/user/settings/ 下生成了配置文件。
2. 用 docker exec 查看进程用户，确认不是 root。
3. 把数据目录属主改成不匹配的用户，确认入口脚本退出并提示。
4. 挂载内容工作区后，确认设置页的 Workspace 显示为锁定。
5. 让智能体尝试读 /app/data 下的设置文件，看沙箱是否挡住。
6. 开启认证后，不带 cookie 访问 /api/，确认被拒绝。
7. 用出站代理记录一次使用，看哪些外部域名被访问。

## 十一、限制与结论

官方镜像的部署并不复杂，复杂的是几个分散在文档里的决定：只发布 3782，数据和内容工作区分开挂载，按场景选沙箱形态，改 JSON 而不是环境变量，临时桥接用完即撤，升级旧版 Compose 前先拷出数据。

本文的限制：没有运行任何容器或命令；没有读 SECURITY_GUARD.md、工作区契约等被引用的文档；沙箱的安全性评价是文档的说法；Podman 的用户命名空间行为因宿主而异，文档自己也说要在目标机器上检查。

下一步建议：先在一次性的测试机上按第十节走一遍，再决定是否上线。
