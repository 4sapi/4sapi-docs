---
title: "【大模型API中转站】第46期 Hermes全模态工作流 | 低成本跑通"
category: 人工智能
tags:
  - 大模型API中转站
  - Hermes
  - 4SAPI
  - 多模态
  - AI Agent
  - 图片生成
  - 视频生成
description: "把 Hermes Agent 的模型入口切到 4SAPI，用低成本方式跑通文本 Agent、图片生成、视频生成和批量内容生产。重点不是免费白嫖，而是统一 Base URL、统一 Key、统一日志和预算，把全模态工作流管起来。"
---

# 【大模型API中转站】第46期 Hermes全模态工作流 | 低成本跑通

本文是【大模型API中转站】系列的第46篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

这篇先把标题里的话说实一点：

```text
不要再写“零成本全模态”。
更准确的说法是：低成本、可控预算、可复用配置。
```

免费 API 听起来很爽，但长期用来跑 Agent、图片生成、视频生成和批量任务，最怕三个问题：

```text
今天能用，明天限额。
单个模型能跑，但能力不全。
没有日志和预算，出了问题不知道钱花在哪。
```

所以这篇换一个更适合长期使用的方案：

```text
Hermes Agent 负责执行工作流。
4SAPI 负责统一模型入口、Key、模型选择、日志和成本。
```

核心思路一句话：

```text
让 Hermes 通过 OpenAI-compatible 方式接入 4SAPI，把文本、图片、视频等模型能力统一收到一个低成本模型网关里。
```

这样你不用在 Hermes 里到处散填不同平台的 Key，也不用每次换模型都重写一套 Skill。模型能力在 4SAPI 侧管理，Hermes 只管调用和编排。

## 1. 准备工作：两个东西就够

先准备两个东西：

| 工具 | 用途 |
| --- | --- |
| Hermes Agent | 本地 Agent 执行入口，负责对话、工具、Skill 和任务编排 |
| 4SAPI | 大模型 API 中转站，负责统一 API Key、Base URL、模型路由和成本记录 |

这篇不讨论任何违规代理用途，只讲合法合规的模型聚合、接口适配、成本管理和本地 Agent 工作流。

### 1.1 部署 Hermes Agent

项目地址：

```text
https://github.com/NousResearch/hermes-agent
```

Windows 用户建议用 WSL。原因很简单：Hermes 这类 Agent 工具经常要读写文件、跑命令、创建 Skill、调用本地脚本。放在类 Linux 环境里，路径、权限和依赖排查会少很多麻烦。

最小检查：

```bash
hermes --version
hermes --help
```

如果你已经装过 Hermes Desktop 或 Gateway，也要分清楚：

| 组件 | 作用 |
| --- | --- |
| Hermes Agent / CLI | 本地执行和配置入口 |
| Hermes Desktop GUI | 桌面聊天界面 |
| Gateway | 飞书、Telegram、Discord 等平台消息入口 |
| `~/.hermes/` | 本地配置、会话、状态和环境变量 |

这篇重点讲 Agent 和模型配置。

### 1.2 注册 4SAPI 并创建专用 Key

4SAPI 接入地址通常按 OpenAI 兼容方式填写：

```text
Base URL: https://4sapi.com/v1
API Key: 4SAPI 工作台创建的令牌
Model: 从 4SAPI 模型广场复制完整模型名
```

建议不要拿主业务 Key 直接填进 Hermes。

单独创建一个：

```text
hermes-agent-dev
```

或者按用途拆成几把 Key：

```text
hermes-text-agent
hermes-image-gen
hermes-video-gen
hermes-batch-task
```

个人使用可以先一把 Key 跑通。团队使用建议从一开始就拆用途，后面查日志和控预算会轻松很多。

## 2. 先配文本 Agent：让 Hermes 走 4SAPI

先不要急着做图片和视频。

第一步只验证文本模型能不能跑。

在终端执行：

```bash
hermes model
```

配置时按这个思路填：

```text
Provider 类型：Custom / Direct API
API Base URL：https://4sapi.com/v1
API Key：4SAPI 后台创建的专用令牌
API 格式：OpenAI-compatible / chat 格式
Model：从 4SAPI 模型广场复制一个文本或 Agent 模型
```

不同版本的 Hermes 菜单文案可能不一样。有的叫 `custom direct API`，有的叫 `OpenAI compatible`，有的会让你选 chat 格式。不要纠结名字，抓住三件事：

```text
base_url = https://4sapi.com/v1
api_key = 你的 4SAPI Key
model = 4SAPI 模型广场里的完整模型名
```

第一次建议选低成本、响应快、稳定的模型。

不要上来就选最贵模型。

文本 Agent 的第一轮目标不是“能力拉满”，而是确认：

```text
1. Hermes 能连到 4SAPI。
2. Key 能鉴权。
3. 模型名能识别。
4. 普通对话能返回。
5. 日志里能看到调用记录。
```

启动 Hermes：

```bash
hermes
```

发一个低风险测试：

```text
请用一句话回复：当前 Hermes 已经接入 4SAPI。
```

如果能正常回复，说明文本通道跑通。

如果不通，优先查四件事：

| 问题 | 排查 |
| --- | --- |
| 401 / 鉴权失败 | Key 是否复制完整，是否多填了 `Bearer` |
| 404 / 模型不存在 | 模型名是否从 4SAPI 模型广场完整复制 |
| URL 错误 | `https://4sapi.com` 和 `https://4sapi.com/v1` 是否被工具重复拼接 |
| 协议不匹配 | Hermes 当前选择的是 OpenAI chat 还是 Anthropic messages |

## 3. 模型不要只选一个，按任务分层

全模态工作流最容易犯的错，是把所有任务都塞给一个模型。

文本、Agent、图片、视频、结构化输出，本来就不是同一类任务。

建议在 4SAPI 侧先按用途分层：

| 层级 | 适合任务 | 选择原则 |
| --- | --- | --- |
| 快速文本档 | 日常问答、摘要、标题、轻量改写 | 低成本、响应快 |
| Agent 主力档 | Hermes 规划任务、写脚本、创建 Skill | 稳定、工具调用理解好 |
| 推理档 | 复杂流程、代码排错、架构判断 | 能力优先，低频使用 |
| 图片档 | 封面、插图、绘本、分镜图 | 看风格稳定性和单张成本 |
| 视频档 | 短视频片段、动效测试、分镜验证 | 先小样，再终稿 |

Hermes 里可以先配一个主力文本模型。

图片和视频不要强行塞到同一个 chat 模型里，而是通过 Skill 调 4SAPI 的对应接口或对应模型。

一句话：

```text
Hermes 是工作流入口，4SAPI 是模型调度层。
```

## 4. 自制图片生成 Skill：不要硬编码 Key

原来的玩法是让 Hermes 学某个平台的图片 API 文档，然后自己写 Skill。

现在换成 4SAPI，也可以这么做，只是要改两个关键点：

```text
1. API 入口统一指向 4SAPI。
2. API Key 从环境变量读取，不要硬编码进 Skill。
```

可以先让 Hermes 读取 4SAPI 文档或你自己的接入说明：

```text
请阅读 4SAPI 的接入文档，重点学习图片生成相关接口。
目标：写一个 Hermes Skill，用 4SAPI 调用图片生成模型。
要求：API Key 从环境变量 4SAPI_API_KEY 读取，不要写死。
```

然后让它创建 Skill：

```text
请帮我创建一个 Hermes Skill：
名称：generate_image_4sapi
用途：调用 4SAPI 的图片生成模型生成图片。

要求：
1. 从环境变量 4SAPI_API_KEY 读取 Key。
2. base_url 使用 https://4sapi.com/v1。
3. 模型名做成参数，默认使用我稍后在 4SAPI 模型广场复制的图片模型。
4. 输入包括 prompt、size、style、output_path。
5. 保存生成结果到本地 images/ 目录。
6. 失败时输出错误码和排查建议，不要吞掉错误。
```

注意这里不要写：

```text
API key 可以先硬编码，后面我自己替换。
```

这句话很危险。

一旦 Skill 被提交到仓库、同步到云盘、发给别人，Key 就可能泄露。

更稳的做法是：

```bash
export 4SAPI_API_KEY="sk-xxxxxxxxxxxxxxxx"
```

或者放到本地 `.env`，并确保 `.env` 不提交到 Git。

Skill 生成后，先测一张小图：

```text
调用 generate_image_4sapi，生成一张 16:9 技术博客封面草图。
主题：Hermes Agent 通过 4SAPI 统一调用文本、图片、视频模型。
风格：米白底、黑色线稿、少量金棕色点缀、极简。
```

第一张图只验证链路，不追求完美。

## 5. 视频生成也可以做 Skill，但要先限预算

视频比图片更容易烧成本。

所以视频 Skill 不建议一上来就开放给 Hermes 随便调用。

先做成“受控工具”：

```text
一次只生成 3-5 秒。
默认低分辨率。
默认不批量。
每次调用前让 Hermes 先输出分镜和预算预估。
```

可以给 Hermes 这样的指令：

```text
请创建一个 Hermes Skill：
名称：generate_video_4sapi
用途：调用 4SAPI 的视频生成模型，生成短视频片段。

要求：
1. API Key 从环境变量 4SAPI_API_KEY 读取。
2. base_url 使用 https://4sapi.com/v1。
3. 模型名作为参数，不硬编码。
4. 默认时长不超过 5 秒。
5. 默认分辨率使用低成本预览档。
6. 每次调用前输出：prompt、模型、时长、分辨率、预计用途。
7. 批量生成必须先询问用户确认。
8. 输出文件保存到 video/ 目录。
```

视频生成建议采用两段式：

```text
先用文本模型写分镜
  -> 人确认
  -> 低成本视频模型出小样
  -> 人选定
  -> 再生成终稿
```

不要直接让 Hermes：

```text
帮我生成 20 段视频素材。
```

这类任务一旦跑起来，账单会比文本 Agent 快很多。

## 6. 批量生图做绘本：可以跑，但要先做风格锁定

原稿里提到的“让 Hermes 写 HTML 儿童绘本《龟兔赛跑》，再批量生成插图”，这个方向可以保留。

但要换成低成本可控版。

不要一上来生成 1 张封面 + 8 张场景图。

先让 Hermes 做三步：

```text
1. 写故事分镜。
2. 做风格规范。
3. 先生成 1 张封面 + 1 张场景样图。
```

指令可以这样写：

```text
请帮我制作一个 HTML 儿童绘本《龟兔赛跑》的生成计划。

要求：
1. 先输出 8 个场景分镜，不要直接生成图片。
2. 统一视觉风格：可爱清新、柔和色彩、儿童绘本、角色一致。
3. 为每个场景写英文图片 prompt。
4. 先选择封面和第 1 个场景做小样。
5. 小样通过后，再批量调用 generate_image_4sapi 生成剩余图片。
6. 每次批量前先列出预计生成张数和使用的模型。
```

风格规范可以这样固定：

```text
Character bible:
- Rabbit: small white rabbit, red scarf, energetic expression.
- Tortoise: green tortoise, round glasses, calm expression.
- Style: soft watercolor children's book, warm daylight, clean background.
- Composition: simple scene, readable emotion, no text inside image.
```

图片 prompt 里要反复带上角色设定。

否则批量生成时最容易出现：

```text
兔子长得一张图一个样。
乌龟颜色乱变。
画风从水彩跑到 3D。
场景里出现乱码文字。
```

如果通过 4SAPI 调用图片模型，建议给生图单独 Key 或单独预算。这样你能看到这本绘本到底花了多少，不会和 Hermes 日常聊天混在一起。

## 7. 一套低成本全模态工作流可以这样设计

现在把文本、图片、视频放到一条链路里。

一个比较稳的 Hermes + 4SAPI 工作流是：

```text
选题/任务
  -> Hermes 用低成本文本模型拆需求
  -> 主力模型生成执行计划
  -> 图片 Skill 生成封面或插图小样
  -> 视频 Skill 生成 3-5 秒预览片段
  -> 人工确认风格和事实
  -> 批量生成
  -> Hermes 整理 HTML / Markdown / 素材目录
  -> 人工发布
```

对应模型策略：

| 阶段 | 建议模型 |
| --- | --- |
| 需求拆解 | 快速文本档 |
| Agent 计划 | Agent 主力档 |
| 代码/HTML 生成 | 代码或文本主力档 |
| 图片小样 | 低成本图片档 |
| 图片终稿 | 质量更稳的图片档 |
| 视频小样 | 低成本视频档 |
| 视频终稿 | 按预算少量生成 |
| 最终检查 | 推理档或人工 |

这就是“低成本”的关键。

不是每一步都用最低价模型，而是：

```text
便宜模型跑试错。
主力模型跑生产。
强模型只做关键判断。
图片和视频先小样再终稿。
```

## 8. 4SAPI 的价值不是“换个 API 地址”

把 Hermes 接到 4SAPI，表面上只是改了三项：

```text
base_url
api_key
model
```

但真正有价值的是后面的治理。

如果你长期用 Hermes 做全模态工作流，至少要管 6 件事：

| 治理项 | 为什么重要 |
| --- | --- |
| Key 分用途 | 文本、图片、视频消耗差异很大 |
| 模型分层 | 不同任务不该都用同一个模型 |
| 预算限额 | 防止 Agent 批量任务跑飞 |
| 调用日志 | 知道钱花在哪一步 |
| 敏感数据边界 | 不把密钥、合同、隐私资料交给不该用的模型 |
| 人工审批 | 图片、视频和发布动作必须人确认 |

尤其是 Hermes 这种 Agent 工具，一旦让它拥有文件读写、联网、脚本执行、Skill 调用能力，就不要只把它当聊天框。

它更像一个会干活的本地助手。

会干活，就要有边界。

## 9. 常见坑：从免费 API 迁到 4SAPI 时别踩

### 坑一：继续写“零成本”

不建议。

图片和视频只要跑起来，就一定要谈成本。

正确口径是：

```text
低成本、可控预算、按需选择模型。
```

这比“免费随便用”更稳，也更可信。

### 坑二：把 Key 写进 Skill

不要硬编码。

统一用环境变量：

```bash
export 4SAPI_API_KEY="sk-xxxxxxxxxxxxxxxx"
```

或者用本地 `.env`。

### 坑三：模型名手打

不要凭记忆写模型名。

去 4SAPI 模型广场复制完整名称。模型名多一个空格、少一个版本号，都可能报错。

### 坑四：文本、图片、视频共用一把无限额 Key

个人测试可以临时这么做。

长期不建议。

至少把视频生成单独拆出来，因为它的成本和失败重试都更敏感。

### 坑五：让 Hermes 自动批量生成，不让人确认

批量生图、批量视频、批量发布，都要先停下来让人确认。

Agent 可以执行，不能替你拍板。

## 10. 一个最小可用配置清单

你可以按这个顺序跑：

```text
1. 安装 Hermes Agent。
2. 在 4SAPI 创建 hermes-agent-dev Key。
3. 执行 hermes model。
4. Provider 选择 Custom / OpenAI-compatible。
5. Base URL 填 https://4sapi.com/v1。
6. API Key 填 4SAPI 专用令牌。
7. Model 从 4SAPI 模型广场复制。
8. 启动 hermes，测试普通对话。
9. 创建 generate_image_4sapi Skill。
10. 用环境变量保存 4SAPI_API_KEY。
11. 先生成 1 张小样。
12. 再考虑批量图片和视频。
```

如果你还要接飞书、Telegram 或 Discord，让 Gateway 复用同一个 4SAPI provider 即可。前面第 23、24 期已经讲过 Hermes GUI、Gateway、`~/.hermes/` 和 provider 路由，不要把桌面窗口和后台网关混在一起。

## 总结：低成本不是白嫖，而是把试错成本管住

这套 Hermes + 4SAPI 的方案，最值得保留的不是“某个免费模型真香”。

免费额度会变，模型会上下架，接口也会调整。

真正能长期复用的是这套结构：

```text
Hermes 负责执行。
Skill 负责封装能力。
4SAPI 负责统一模型入口。
Key 和预算按用途拆。
图片和视频先小样再终稿。
发布前人工拍板。
```

这样配置以后，你确实可以用相对低的成本，把文本 Agent、图片生成、视频生成、HTML 内容生产、批量素材整理串成一条全模态工作流。

它不一定每一步都超过一线模型。

但它足够实用，足够可控，也更适合长期跑。

工具帮你省时间。

4SAPI 帮你看住模型入口和成本。

最后省下来的时间花在哪里，才是你和别人真正拉开差距的地方。

## 参考资料

- Hermes Agent GitHub：https://github.com/NousResearch/hermes-agent
- 4SAPI 接入文档：https://4sapi.apifox.cn/
- 4SAPI 接入地址：https://4sapi.com/v1
