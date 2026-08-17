---
title: "【大模型API中转站】第75期 Pixelle-Video安装 | 少踩80%配置坑"
category: 人工智能
tags:
  - 大模型API中转站
  - Pixelle-Video
  - ComfyUI
  - 数字人
  - 4SAPI
  - 本地部署
description: "从 Windows 一键整合包、源码安装、ComfyUI 本地服务到 4SAPI 写稿模型配置，手把手讲清楚 Pixelle-Video 数字人 Agent 的安装路线、系统配置和常见排错。"
---

# 【大模型API中转站】第75期 Pixelle-Video安装 | 少踩80%配置坑

本文是【大模型API中转站】系列的第75篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一篇我们讲了为什么普通人值得在自己电脑上搭一套数字人 Agent。

这一篇只做一件事：

```text
把 Pixelle-Video 和 ComfyUI 装起来，并把写稿模型配置好。
```

数字人 Agent 最容易劝退人的地方，不是模型多难。

而是安装和配置太容易混在一起。

很多人还没生成第一条视频，就先被这些问题卡住：

- Pixelle-Video 页面打不开
- ComfyUI 页面打不开
- 不知道该填 `localhost` 还是 `127.0.0.1`
- 不知道本地 workflow 和 RunningHub workflow 有什么区别
- 写稿模型 Key 填了，但测试失败
- Base URL 填成了重复路径
- 电脑没有 NVIDIA 显卡，却硬要跑本地视频模型
- 把 API Key 写进配置文件后又手滑传到 GitHub

这篇就把安装路线拆清楚。

先说结论。

如果你是 Windows 用户，第一次安装，优先用 Pixelle-Video 官方 Windows 一键整合包。

如果你是 Mac、Linux，或者你想改代码、改工作流、改页面，再走源码安装。

如果你要跑本地图像/视频工作流，再单独装 ComfyUI。

如果你不想折腾显卡，就先用 RunningHub 或 API 媒体模型。

写稿模型这一层，我建议用 4SAPI 做统一入口。

原因很简单：

```text
Pixelle-Video 的 LLM 配置支持 OpenAI SDK compatible API。
4SAPI 正好可以作为 OpenAI-compatible 的模型网关。
```

也就是说，后面你要换 DeepSeek、GPT、Claude、GLM、Qwen，不一定每次都换一堆软件配置。

你只要维护：

```text
Base URL
API Key
Model
```

这就是本系列一直讲的大模型 API 中转站的价值。

## 1. 先搞清楚：你到底要装几样东西？

Pixelle-Video 不是一个单独模型。

它更像一个短视频工作台，把多个能力串起来。

基础链路大概是：

```text
浏览器页面
  -> Pixelle-Video 主程序
  -> LLM 写稿模型
  -> ComfyUI / RunningHub / API 媒体模型
  -> TTS 配音
  -> 视频合成
  -> output 成片
```

所以你至少会遇到三类配置：

| 模块 | 作用 | 第一次怎么选 |
| --- | --- | --- |
| Pixelle-Video | 主页面和流程编排 | 必装 |
| LLM | 写文案、分镜、提示词 | 推荐用 4SAPI 或 DeepSeek 跑通 |
| ComfyUI | 本地图像/视频工作流 | 有显卡再重点折腾 |
| RunningHub | 云端工作流 | 没显卡或想省心时使用 |
| TTS | 配音 | 先用 Edge-TTS |

很多教程一上来就让你把所有 Key 填满。

我不建议这样。

第一次只追求最小闭环：

```text
Pixelle-Video 能打开
LLM 能测试成功
ComfyUI 或云端图像方案能跑通
TTS 能出声音
output 里能拿到视频
```

先闭环，再优化。

## 2. 安装前检查：别拿轻薄本硬打重活

Pixelle-Video 官方安装文档里写得很清楚：系统需要 Python 3.10 或更高版本，支持 Windows、macOS、Linux，包管理器推荐 uv，也可以用 pip。

如果你用 Windows 一键整合包，官方说明是不需要手动安装 Python、uv 或 ffmpeg。

但如果你走源码安装，这些基础环境就要自己准备。

先看电脑情况：

| 电脑 | 推荐路线 | 说明 |
| --- | --- | --- |
| Windows + NVIDIA 显卡 | 一键整合包 + 本地 ComfyUI | 最适合入门和本地出图 |
| Windows 无独显 | 一键整合包 + RunningHub/API | 本地视频模型别硬跑 |
| Mac | 源码安装 + 云端图像/视频 | 部分本地工作流会受限 |
| Linux 服务器 | 源码安装 + 本地/云端都可 | 适合开发者和团队部署 |

内存建议 16G 起步。

显存如果要跑本地 ComfyUI 图像工作流，6GB 以上会舒服一些。

如果你只有办公本，也不是不能玩。

只是路线要改：

```text
本地跑页面和写稿
图像/视频交给云端
配音先用 Edge-TTS
```

别一上来就把轻薄本当工作站。

## 3. 路线一：Windows 一键整合包

这是我最推荐新手使用的路线。

Pixelle-Video README 里写得很直接：

```text
下载 Windows 一键整合包
解压
双击 start.bat
浏览器打开 http://localhost:8501
```

它的优势是：

- 不用手动装 Python
- 不用手动装 uv
- 不用手动装 ffmpeg
- 环境冲突少
- 适合先跑起来

安装流程：

1. 打开 Pixelle-Video GitHub Releases。
2. 下载最新 Windows 一键整合包。
3. 解压到一个英文路径。
4. 双击 `start.bat`。
5. 等终端启动完成。
6. 浏览器访问：

```text
http://localhost:8501
```

这里建议路径不要太深，也不要放在中文目录或带空格太多的目录里。

比如可以放：

```text
D:\AI\Pixelle-Video
```

不要放：

```text
C:\Users\你的名字\桌面\新建文件夹 (3)\数字人最终版\Pixelle Video 测试
```

不是说一定会坏。

而是这类 AI 工具经常会调用 Python、ffmpeg、模型文件、临时路径，路径越怪，排错越烦。

### 3.1 怎么判断一键包启动成功？

看到两个信号。

第一，终端没有继续刷红色报错。

第二，浏览器能打开：

```text
http://localhost:8501
```

如果浏览器打不开，先看终端里有没有端口提示。

有时端口会被占用，软件可能换到别的端口。

常见问题：

| 现象 | 可能原因 | 处理 |
| --- | --- | --- |
| 双击没反应 | 被系统拦截或路径问题 | 右键管理员运行，换英文路径 |
| 浏览器打不开 | 服务没启动或端口占用 | 看终端日志，换端口或关占用程序 |
| 页面打开但按钮报错 | 还没配置 API | 去系统配置填 LLM 和图像服务 |
| 终端闪退 | 依赖或权限问题 | 用命令行运行 `start.bat` 看报错 |

如果你不懂报错，直接把终端完整报错复制给 AI IDE。

但记住：

```text
不要把完整 API Key 贴出去。
```

## 4. 路线二：源码安装

源码安装适合三类人：

- Mac / Linux 用户
- 想改代码的人
- 想长期魔改工作流的人

官方 README 里的核心命令是：

```bash
git clone https://github.com/AIDC-AI/Pixelle-Video.git
cd Pixelle-Video
uv run streamlit run web/app.py
```

启动后浏览器会打开：

```text
http://localhost:8501
```

如果你没有安装 uv，要先安装 uv。

如果你不用 uv，也可以按项目文档走 pip 路线，但我更建议跟官方推荐走。

原因是这类 Python 项目最容易被依赖版本折腾。

让 uv 管环境，会少踩一些坑。

### 4.1 源码安装时给 AI IDE 的提示词

你可以把下面这段直接发给 Codex、Claude Code 或 Cursor：

```text
请帮我在当前电脑源码安装 Pixelle-Video。

要求：
1. 先读取官方 README 和安装文档。
2. 检查我的系统是否已经安装 git、uv、Python。
3. 使用官方推荐命令安装，不要自己发明安装方式。
4. 启动后确认浏览器能打开 http://localhost:8501。
5. 如果报错，只排查当前错误，不要一次改很多地方。
6. 不要读取或输出我的 API Key。
```

AI IDE 最适合干这种活。

不是因为它比你更懂所有依赖。

而是它能耐心看报错、查路径、改命令。

你只要盯住一个原则：

```text
每次只解决一个报错。
```

别让它一口气重装半个系统。

### 4.2 源码安装常见坑

常见坑如下：

| 问题 | 表现 | 建议 |
| --- | --- | --- |
| Python 版本太低 | 安装依赖失败 | 升到 Python 3.10+ |
| uv 没装好 | `uv` 命令不存在 | 重新安装 uv 并重开终端 |
| ffmpeg 缺失 | 合成视频失败 | 安装 ffmpeg 并加入 PATH |
| 端口被占用 | 8501 打不开 | 关掉旧 Streamlit 进程 |
| 网络不稳定 | 依赖下载失败 | 换网络或配置镜像 |

源码安装成功，不代表视频能生成。

它只代表 Pixelle-Video 主页面能打开。

后面还要配置 LLM、图像/视频服务和 TTS。

## 5. ComfyUI：本地出图的核心

ComfyUI 是 Pixelle-Video 本地图片/视频工作流的底层引擎。

Pixelle-Video 负责把创作流程做成页面。

ComfyUI 负责真正执行图像或视频工作流。

你可以这样理解：

```text
Pixelle-Video 是导演。
ComfyUI 是摄影棚。
LLM 是编剧。
TTS 是配音演员。
```

第一次装 ComfyUI，Windows 用户可以看 ComfyUI 官方的 portable 安装路线。

它通常会按显卡分不同包：

- NVIDIA
- AMD
- Intel

如果你是 NVIDIA 显卡，常见启动脚本类似：

```text
run_nvidia_gpu.bat
```

启动成功后，浏览器访问：

```text
http://127.0.0.1:8188
```

Pixelle-Video 配置里也默认使用这个地址。

### 5.1 本地 ComfyUI 适合谁？

适合：

- 有 NVIDIA 显卡
- 愿意下载模型文件
- 愿意排 custom node 问题
- 想省长期云端成本
- 想魔改工作流

不适合：

- 只想点按钮
- 没有显卡
- 磁盘空间很紧
- 不想碰模型文件
- 对画质要求很高但又不想调参

本地 ComfyUI 最大的好处是可控。

最大的坏处也是可控。

因为可控就意味着你要管模型、节点、路径、显存、工作流版本。

### 5.2 怎么确认 ComfyUI 可用？

先独立打开：

```text
http://127.0.0.1:8188
```

能打开 ComfyUI 页面，再回 Pixelle-Video。

在系统配置里填：

```text
ComfyUI URL: http://127.0.0.1:8188
```

点测试连接。

如果 Pixelle-Video 测试不通，先别怀疑 4SAPI，也别怀疑写稿模型。

这就是 ComfyUI 这条链路的问题。

排查顺序：

```text
ComfyUI 页面能否打开
ComfyUI 端口是不是 8188
Pixelle-Video 和 ComfyUI 是否在同一台机器
防火墙是否拦截
Docker 场景是否要用 host.docker.internal
```

Pixelle-Video 的配置示例里也提醒过，如果是 Docker 场景，Mac/Windows 可能要用 `host.docker.internal:8188`，Linux 则看宿主机 IP。

普通本机安装不用管这个。

## 6. RunningHub：没有显卡也能跑的云端路线

如果你没有显卡，或者不想管 ComfyUI 模型文件，可以考虑 RunningHub。

Pixelle-Video 配置示例里有这类字段：

```yaml
runninghub_api_key: ""
runninghub_concurrent_limit: 1
```

图像和视频工作流也有本地和云端两类选择。

比如配置示例里会出现类似：

```text
runninghub/image_flux.json
selfhost/image_flux.json
runninghub/video_wan2.1_fusionx.json
selfhost/video_wan2.1_fusionx.json
```

这里的逻辑很简单：

| workflow 前缀 | 含义 | 成本 |
| --- | --- | --- |
| `selfhost` | 本地 ComfyUI 跑 | 吃本机显卡和模型 |
| `runninghub` | RunningHub 云端跑 | 吃云端额度 |

第一次如果你只是要验证 Pixelle-Video 页面和写稿流程，可以先用云端路线。

但不要误会。

云端不是免费魔法。

它只是把本机硬件压力转移到云服务，成本从显卡变成额度。

## 7. 系统配置第一块：LLM 写稿模型

这块最重要。

Pixelle-Video 的配置示例写得很直白：

```yaml
llm:
  api_key: ""
  base_url: ""
  model: ""
```

并且注明：

```text
Supports any OpenAI SDK compatible API
```

这就是为什么 4SAPI 很适合放在这里。

你可以把 Pixelle-Video 的写稿请求，统一发到 4SAPI，再由 4SAPI 转到不同模型。

配置思路：

```text
api_key: 你的 4SAPI Key
base_url: https://4sapi.com/v1
model: 从 4SAPI 模型广场复制完整模型名
```

如果 Web UI 里没有直接写 4SAPI，你就找：

```text
OpenAI Compatible
Custom
自定义
```

这类选项。

本质不是品牌名。

本质是它能不能填自定义 `base_url`。

### 7.1 为什么不用每个平台都填一遍 Key？

因为数字人 Agent 会频繁调用写稿模型。

一次生成视频，可能要调用：

- 主题扩写
- 分镜拆分
- 画面提示词
- 标题
- 口播改写
- 带货卖点
- 字幕整理

如果每个模型平台都有一套 Key，后面会很乱。

通过 4SAPI 收敛后，至少可以做到：

- 一个入口调用多模型
- 按模型名切换
- 日志统一查看
- 成本统一复盘
- 测试 Key 和生产 Key 分开

对个人创作者，这是省心。

对小团队，这是治理。

### 7.2 Base URL 不要填错

最常见错误是路径重复。

有的软件让你填 Base URL：

```text
https://4sapi.com/v1
```

有的软件让你填完整接口：

```text
https://4sapi.com/v1/chat/completions
```

还有的软件自己会追加 `/v1/chat/completions`。

所以你要看字段名。

错误示例：

```text
https://4sapi.com/v1/v1/chat/completions
https://4sapi.com/v1/chat/completions/chat/completions
```

如果报 404 或接口路径错误，第一时间查这里。

不要上来就换模型。

## 8. 系统配置第二块：ComfyUI / RunningHub

Pixelle-Video 的图像/视频生成支持两条路线：

```text
本地部署：ComfyUI
云端部署：RunningHub
```

本地路线填：

```text
ComfyUI URL: http://127.0.0.1:8188
```

云端路线填：

```text
RunningHub API Key
并发限制
工作流
```

第一次建议你只选一条路线。

不要同时开两个。

如果你选了 `selfhost` 工作流，却没有启动 ComfyUI，就会失败。

如果你选了 `runninghub` 工作流，却没有填 RunningHub Key，也会失败。

这类错误不是 Pixelle-Video 不行。

是路线没对齐。

### 8.1 本地和云端怎么选？

| 场景 | 建议 |
| --- | --- |
| 有 NVIDIA 显卡，想省钱 | 本地 ComfyUI |
| 没显卡，想先跑通 | RunningHub |
| Mac 用户 | 优先云端图像/视频 |
| 只测写稿流程 | 先用最简单图像方案 |
| 批量商业生产 | 本地和云端都要评估 |

我的建议是：

```text
第一天跑通，不追求本地全免费。
第二天再决定哪些环节迁回本地。
```

很多人第一天就想“完全免费、本地全跑、画质还高、速度还快”。

这四个目标经常互相打架。

## 9. 系统配置第三块：API 媒体模型

Pixelle-Video 近几个版本增加了直连 API 媒体模型配置。

README 里提到支持 DashScope、OpenAI、Seedream、Seedance、Kling 等图像/视频生成服务。

这块不是第一天必须填。

它适合后面做画质升级。

你可以这样理解三条路线：

| 路线 | 代表 | 特点 |
| --- | --- | --- |
| 本地工作流 | ComfyUI selfhost | 省云端钱，但吃硬件 |
| 工作流云端 | RunningHub | 不吃本机硬件，但要额度 |
| 直连媒体 API | OpenAI / Seedance / 可灵等 | 效果可能更稳，但按 API 付费 |

第一条视频，尽量别把 API 媒体模型全填满。

否则你排错时会不知道到底是哪一层失败。

最小化配置才是新手友好路线。

## 10. 不要把 `config.yaml` 提交到 Git

Pixelle-Video 的配置示例第一行就提醒：

```text
Copy this file to config.yaml and fill in your settings
Never commit config.yaml to Git
```

这不是一句客气话。

`config.yaml` 里可能有：

- 4SAPI Key
- OpenAI Key
- DashScope Key
- RunningHub Key
- Kling access key
- 代理地址

这些都不应该进 Git。

建议：

```text
config.example.yaml 保留示例
config.yaml 加进 .gitignore
真实 Key 只放本机或服务器环境变量
团队成员各自创建自己的配置
```

如果你让 AI IDE 帮你排错，也不要把完整配置文件直接贴给它。

可以把 Key 改成：

```text
sk-****hidden****
```

再发。

## 11. 最小可用配置清单

如果你已经打开 Pixelle-Video 页面，下一步按这个清单来：

```text
[ ] Pixelle-Video 页面能打开：http://localhost:8501
[ ] LLM 已配置 API Key
[ ] LLM 已配置 Base URL
[ ] LLM 已配置 Model
[ ] LLM 测试成功
[ ] ComfyUI 页面能打开：http://127.0.0.1:8188
[ ] Pixelle-Video 里 ComfyUI 测试连接成功
[ ] 如果用 RunningHub，已填写 RunningHub API Key
[ ] workflow 前缀和实际路线一致：selfhost 或 runninghub
[ ] TTS 先使用 Edge-TTS
[ ] 没有把 config.yaml 提交到 Git
```

这里有一个实操建议：

每配置完一块，就测一次。

不要全部填完再测。

顺序是：

```text
先测 LLM
再测 ComfyUI / RunningHub
再测 TTS
最后跑完整视频
```

这样失败时，你知道是哪一层的问题。

## 12. 常见报错：先对号入座

### 12.1 页面打不开

检查：

```text
start.bat 是否还在运行
终端是否报错退出
8501 端口是否被占用
是否在正确目录启动
防火墙是否拦截
```

如果是源码安装，确认：

```text
uv run streamlit run web/app.py
```

命令是在 Pixelle-Video 项目目录里执行的。

### 12.2 LLM 测试失败

检查：

```text
API Key 是否复制完整
Base URL 是否带错路径
Model 是否从模型广场复制
账户是否有余额
当前模型是否有权限
```

如果你用 4SAPI，先用最简单的文本生成模型测试。

不要第一步就拿最贵或最复杂的模型。

### 12.3 ComfyUI 测试失败

检查：

```text
http://127.0.0.1:8188 能不能打开
ComfyUI 是否启动
工作流是否选择 selfhost
模型文件是否齐全
custom node 是否缺失
```

如果你根本没装 ComfyUI，就不要选 selfhost 工作流。

### 12.4 提示缺 RunningHub Key

多半是你选了 `runninghub` 工作流。

解决方案二选一：

```text
填 RunningHub API Key
或者改回 selfhost 工作流
```

看你到底要走哪条路线。

### 12.5 生成视频失败

先不要急着改 LLM。

视频失败可能发生在：

```text
写稿
配图
配音
合成
保存 output
```

看日志里失败在哪一步。

如果已经生成了图片和音频，但最后没有 mp4，多半是合成或 ffmpeg 问题。

如果连脚本都没生成，是 LLM 问题。

如果脚本有了但没有图，是 ComfyUI/RunningHub 问题。

## 13. 给 AI IDE 的排错提示词

你可以把下面这段收藏起来：

```text
我正在安装 Pixelle-Video。

当前环境：
- 系统：
- 安装路线：Windows 整合包 / 源码安装
- Pixelle-Video 页面是否能打开：
- ComfyUI 页面是否能打开：
- LLM 使用：4SAPI / DeepSeek / 其他
- 图像方案：selfhost / runninghub / api

下面是报错日志。
请你只分析当前报错的直接原因。
不要让我重装整个项目。
不要输出或记录我的 API Key。
先给我 3 个最可能原因和对应验证命令。
```

这个提示词的重点是“只分析当前报错”。

AI 工具很容易看到一个报错，就想顺手优化一堆东西。

安装阶段不要优化。

先让系统跑起来。

## 14. 我的推荐配置

如果你是普通创作者，我建议这样开局：

```text
系统：Windows
安装：Pixelle-Video 一键整合包
LLM：4SAPI + 低成本中文模型
配音：Edge-TTS
图像：先 RunningHub 或轻量本地 workflow
视频：第一条先静态图模板
数字人：等普通视频跑通后再试
```

如果你是开发者，我建议：

```text
安装：源码安装
LLM：4SAPI OpenAI-compatible
配置：config.yaml 本地维护，不提交 Git
图像：本地 ComfyUI + 可替换 workflow
日志：记录每次模型、分镜数、失败阶段
```

如果你是团队，我建议：

```text
4SAPI 创建单独项目 Key
测试环境和生产环境拆 Key
每个内容线单独额度
统一日志复盘写稿模型成本
不要让前端或普通成员接触完整 Key
```

同一套工具，不同人用法不一样。

别照搬别人的“全本地免费路线”。

你要的是适合自己机器和内容目标的路线。

## 15. 总结

Pixelle-Video 的安装不难。

难的是把几条链路分清楚：

```text
Pixelle-Video：主页面和流程编排
LLM：写稿和分镜
ComfyUI：本地图像/视频工作流
RunningHub：云端工作流
API 媒体模型：画质升级路线
TTS：配音
```

第一次安装，只要抓住两个页面：

```text
Pixelle-Video：http://localhost:8501
ComfyUI：http://127.0.0.1:8188
```

再抓住三个配置：

```text
LLM API Key
LLM Base URL
LLM Model
```

写稿模型建议走 4SAPI，因为 Pixelle-Video 本身支持 OpenAI SDK compatible API，4SAPI 可以把多个模型收敛到一个入口里。

后面你要批量做短视频，真正省事的不是少填一次 Key。

而是模型、成本、日志、权限都能统一管理。

下一篇我们不再停留在配置。

我们会直接跑第一条短视频：

```text
一句主题 -> AI 写稿 -> 分镜 -> 配音 -> 配图 -> 合成 -> output 成片
```

参考资料：

- Pixelle-Video GitHub：https://github.com/AIDC-AI/Pixelle-Video
- Pixelle-Video 配置示例：https://github.com/AIDC-AI/Pixelle-Video/blob/main/config.example.yaml
- Pixelle-Video 安装文档：https://aidc-ai.github.io/Pixelle-Video/zh/getting-started/installation/
- Pixelle-Video 配置文档：https://aidc-ai.github.io/Pixelle-Video/zh/getting-started/configuration/
- ComfyUI Windows Portable：https://docs.comfy.org/installation/comfyui_portable_windows
- 4SAPI 接入文档：https://4sapi.apifox.cn/
