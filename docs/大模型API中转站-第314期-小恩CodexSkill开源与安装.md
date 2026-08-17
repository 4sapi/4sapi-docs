---
title: "Codex Skill 如何把角色规范变成可复用的配图工作流"
tags:
  - Codex Skill
  - AI配图
  - 开源项目
description: "一个角色 IP 要参与内容生产，不能只提供几张参考图，还需要身份、视觉约束、动作语义、Prompt 模板和验收规则。"
---
# Codex Skill 如何把角色规范变成可复用的配图工作流
一个角色 IP 要参与内容生产，不能只提供几张参考图，还需要身份、视觉约束、动作语义、Prompt 模板和验收规则。本文以 Xiao En AI Observer Illustrations 的 Codex Skill 为例，说明如何安装、阅读目录、调用配图流程，并用开源仓库中的文件验证每个规则实际负责什么。


很多人做过自己的个人 IP。

头像、表情包、海报、贴纸和角色设定都完成了，最后 IP 却停在了“展示形象”这一步。真正开始写文章、做教程、拆项目时，AI 仍然不知道这个角色是谁，也不知道它应该怎样参与表达。

于是，IP 只是一个文件夹里的图片。

这次有一个更有意思的尝试：把一个 IP 的审美、性格、动作和判断写成规则，让它成为可以被 AI 调用的生产能力。

开源项目：[GitHub 项目仓库](https://github.com/VaneAi888/xiaoen-ai-observer-illustrations)

这个仓库把“小恩｜AI时代观察员”做成了一个独立的 Codex Skill。它不是一套简单的角色提示词，而是把角色身份、视觉 DNA、构图模式、Prompt 模板、文章认知锚点和生成后 QA 组合成了一条配图工作流。

当前仓库是 v1.0，默认目标是为中文文章生成 16:9 横版正文配图。本文先讲它是什么、怎么安装、怎样调用，以及为什么它和普通“给角色换皮”的生图 Prompt 不一样。

## 一、小恩不是装饰，而是动作执行者

小恩的对外身份是“AI时代观察员”。它安静、专注、好奇，喜欢观察变化、记录发现、拆解结构、搬运信息、修理问题、测试方案和推动实验。

它的视觉识别包括：

~~~text
黑色实心主体
两只与头部相连的高而圆润双峰耳朵
白色星泪状眼纹
面部中央偏下的小白色菱形
黑色斗篷水滴形身体
隐藏式短手臂和两条短腿
大头、小身体、克制的半眯眼神
~~~

但角色识别只是第一层。

这个 Skill 还设置了一条更重要的构图原则：

~~~text
如果把小恩从画面里删掉，
整张图表达的意思依然完全成立，
那小恩就只是装饰，这张图需要重新设计。
~~~

因此，小恩不能只是站在图表旁边、坐在标题下面、挥手、点赞或举一块写着结论的牌子。它必须真正参与文章核心概念的表达：

- 观察一个变化。
- 拆开一个复杂结构。
- 筛选一批 AI 输出。
- 校准一个有偏差的结果。
- 修理一个流程断点。
- 推动下一步动作。
- 把经验记录并归档。

这就是它和普通“IP 贴图”的区别。普通贴图先画好系统，再把角色放进去；小恩 Skill 要求先确定“小恩正在做什么”，再让整张图围绕这个动作成立。

## 二、这个仓库实际包含什么

不要只看仓库首页的几张示例图。当前仓库真正可复用的部分在嵌套的 Skill 目录里：

~~~text
xiaoen-ai-observer-illustrations/
├── SKILL.md
├── agents/
│   └── openai.yaml
├── assets/
│   ├── character/
│   │   ├── xiaoen-reference-sheet.png
│   │   └── xiaoen-reference-sheet-labeled.png
│   └── examples/
└── references/
    ├── style-dna.md
    ├── xiaoen-ip.md
    ├── composition-patterns.md
    ├── prompt-template.md
    └── qa-checklist.md
~~~

每个文件的职责不同：

| 文件 | 作用 |
| --- | --- |
| SKILL.md | 定义什么时候调用、先读什么、怎样提炼文章、怎样出 Shot List、怎样生成和检查 |
| style-dna.md | 规定白底、黑色手绘、留白、颜色、文字和风格禁忌 |
| xiaoen-ip.md | 规定耳朵、眼纹、菱形、身体、动作、表情和不可变化特征 |
| composition-patterns.md | 给出单场景、工作流、系统局部、前后变化和分镜构图方法 |
| prompt-template.md | 提供完整生图模板、精简模板、编辑模板和常见修正语句 |
| qa-checklist.md | 检查角色一致性、动作必要性、隐喻、留白、文字和重复构图 |
| agents/openai.yaml | 为 Codex 提供显示名称、简介、默认调用语句和隐式调用策略 |

也就是说，仓库交付的不是一个“画小恩”的 Prompt，而是一套从文章理解到配图验收的规则系统。

## 三、正确安装：复制嵌套的 Skill 子目录

仓库根目录和真正的 Skill 目录不是同一层。

如果你直接把整个仓库复制到 skills 目录，Codex 可能找不到正确的 SKILL.md。真正应该安装的是：

~~~text
仓库根目录/xiaoen-ai-observer-illustrations/
~~~

### 1. Linux 或 macOS

先克隆仓库：

~~~bash
git clone https://github.com/VaneAi888/xiaoen-ai-observer-illustrations.git
cd xiaoen-ai-observer-illustrations
~~~

再把嵌套的 Skill 目录复制到 Codex skills 目录：

~~~bash
mkdir -p ~/.codex/skills
cp -R ./xiaoen-ai-observer-illustrations ~/.codex/skills/
~~~

安装后的关键路径应该是：

~~~text
~/.codex/skills/xiaoen-ai-observer-illustrations/SKILL.md
~/.codex/skills/xiaoen-ai-observer-illustrations/assets/
~/.codex/skills/xiaoen-ai-observer-illustrations/references/
~~~

### 2. Windows PowerShell

如果你已经下载并解压仓库，可以这样复制：

~~~powershell
$skillSource = ".\xiaoen-ai-observer-illustrations"
$skillRoot = Join-Path $HOME ".codex"
$skillRoot = Join-Path $skillRoot "skills"

New-Item -ItemType Directory -Force -Path $skillRoot | Out-Null
Copy-Item -Recurse -Force $skillSource $skillRoot
~~~

如果仓库在其他位置，把 skillSource 改成实际的绝对路径。例如：

~~~powershell
$skillSource = "C:/Users/admin/Downloads/xiaoen-ai-observer-illustrations/xiaoen-ai-observer-illustrations"
~~~

检查安装结果：

~~~powershell
$skillPath = Join-Path $skillRoot "xiaoen-ai-observer-illustrations"
Get-ChildItem $skillPath
Test-Path (Join-Path $skillPath "SKILL.md")
Test-Path (Join-Path $skillPath "assets/character/xiaoen-reference-sheet.png")
~~~

最后两个命令都应该返回 True。如果 SKILL.md 在 skills/xiaoen-ai-observer-illustrations/xiaoen-ai-observer-illustrations/SKILL.md 这种多套了一层的路径里，说明复制时把外层目录也包进去了，需要把真正的 Skill 子目录移动到 skills 下。

### 3. 安装后重新打开 Codex

Skill 是由 Codex 在当前环境里发现和读取的。复制完成后，建议重新打开 Codex 或新建一个任务，再用 $xiaoen-ai-observer-illustrations 验证是否能调用。

仓库包含 NOTICE.md 和 MIT LICENSE。二次使用或再分发时，应保留许可证文件，并单独核查角色、参考图、字体和客户素材的授权边界。

## 四、三种调用方式

### 1. 直接生成文章配图

把文章正文贴给 Codex：

~~~text
Use $xiaoen-ai-observer-illustrations
把下面这篇中文文章生成 4 张小恩 AI时代观察员风格的正文配图。

要求：16:9 横版、纯白背景、极简黑色手绘、少量橙蓝中文批注、大量留白。
每张只讲一个核心认知，小恩必须承担观察、记录、拆解、搬运、修理、校准或推动中的一个必要动作。

<粘贴文章>
~~~

Skill 会先消化文章，寻找真正值得视觉化的认知锚点，再为每张图确定核心动作、主物件、隐喻、标注和留白位置。它的默认目标不是平均给每个段落配一张图。

### 2. 只做 Shot List，不生成图片

如果你想先审配图逻辑，可以明确说：

~~~text
Use $xiaoen-ai-observer-illustrations 先不要生图。
分析下面文章中最值得视觉化的认知锚点，输出 5 张左右的 shot list。
每张写清楚：放置位置、核心认知、构图类型、物理动作、小恩动作、主要物件、短标注和留白位置。

<粘贴文章>
~~~

这一步适合编辑、作者和设计师一起审。你可以先删掉不必要的配图，再决定哪些 Shot 进入生成阶段。

### 3. 生成后编辑或修正

角色变形、错字和多余标题不一定需要整张重做。仓库的 prompt-template.md 提供了局部编辑方式：

~~~text
Use $xiaoen-ai-observer-illustrations 修正这张图中的小恩角色。
保持构图和道具不变，让小恩严格匹配官方角色参考图：
双峰圆润耳朵、星泪眼纹、中央白色菱形、小斗篷水滴身体和短腿。
~~~

如果只是标题错误，可以明确要求只替换文字；如果小恩变成猫、兔、机器人或人类，则先修角色身份，不要继续修背景和道具。

## 五、它默认解决什么问题

当前 Skill 的适用内容包括：

- AI 工具和 AI 工作流。
- 设计方法和视觉拆解。
- 内容创作、选题、脚本和内容复用。
- 项目拆解、产品流程和系统卡点。
- 个人成长复盘和经验沉淀。
- 抽象知识、方法论和观点解释。

这些主题有一个共同点：文章里有值得被“动作化”的认知。

例如：

~~~text
AI 输出不能直接使用，需要人工筛选和校准。
~~~

Skill 不会把这句话直接排成一张文字卡片，而是可以转成：小恩站在简化机器旁，把输出卡片逐张筛选、标记、校准和归档。

再例如：

~~~text
项目最大的问题发生在两个环节的交接处。
~~~

可以转成：一个简单装置的接口中间断开，小恩拿着连接线把输入和输出接通。

这种配图的重点不是“画一个漂亮场景”，而是让读者看见抽象概念对应的物理动作。

## 六、默认视觉 DNA

仓库把风格写得很具体：

~~~text
画幅：16:9 横版
背景：纯白，必要时极浅灰白
线条：黑色、细、轻微手绘抖动
留白：主体约占 40%–60%，至少 35% 有效留白
颜色：黑色主结构，少量橙色和蓝色批注，红色只用于风险
文字：通常 3–5 处短中文标注，每处优先 2–6 个字
~~~

它刻意避开：

- 蓝紫科技感和发光屏幕。
- 复杂真实 UI 和品牌 Logo。
- 正式 PPT 信息图和规整流程图。
- 商业矢量插画、品牌 KV 和广告风。
- 儿童卡通、萌系大眼、腮红和商业吉祥物风。
- 米色纸张纹理、渐变、噪点和复杂室内背景。

这套限制看起来像是在“减少画面可能性”，实际上是在保护角色和文章内容。没有限制时，图像模型很容易把“AI 文章配图”画成发光服务器、蓝紫 HUD 和一堆矩形卡片。

## 七、第一次调用的推荐流程

不要一上来就把一万字文章要求生成九张图。可以按下面顺序：

~~~text
安装 Skill
  -> 用 3 到 5 段短文章测试角色和风格
  -> 先输出 Shot List
  -> 人工删除不值得视觉化的锚点
  -> 单张生成
  -> 按 QA 清单检查
  -> 对角色、隐喻或文字做局部修正
  -> 再处理长文和批量配图
~~~

第一轮测试建议选择含有明显动作的内容，例如“筛选 AI 输出”“修复流程断点”“把经验归档”，因为这些主题更容易验证小恩是否真的承担核心动作。

如果第一张图只是“小恩站在一个系统旁边”，不要急着继续生成五张。先回到核心问题：小恩究竟在观察、拆解、修理还是推动什么？如果这个动词说不清楚，画面大概率也说不清楚。

## 八、上游 API 在这套 Skill 旁边怎么发挥作用

安装和调用小恩 Skill 不需要 上游 API。它本身是一个 Codex 配图 Skill，负责把文章转换成 Shot List、生成提示词和执行图像工作流。

但在企业内容生产中，Skill 往往只是链路的一环：

~~~text
文章写作 / 研究
  -> 上游 API 企业API网关统一调用文本模型
  -> 小恩 Skill 提炼认知锚点和配图策略
  -> 图像生成或视觉审核
  -> 内容系统发布和归档
~~~

上游 API 适合承接企业里的多模型统一接入：

- 文本模型负责文章、标题和 Shot List 的前置整理。
- 视觉模型负责参考图分析、角色一致性检查或内容审核。
- Embedding 模型负责素材检索和历史配图复用。
- 其他模型负责企业内容系统中的摘要、改写和多渠道分发。

这样，作者可以使用小恩作为稳定的视觉生产能力，企业则通过 上游 API 统一管理模型入口、API Key、权限、预算、调用日志和失败告警。

不要把 上游 API 写成这个 GitHub Skill 的必需依赖，也不要推断仓库已经原生集成了某个企业 API。更准确的营销表达是：小恩负责 IP 化的配图规则，上游 API 负责企业多模型调用的统一治理，两者可以在内容生产链路上协作。

## 九、许可证和二次使用边界

这个仓库包含 LICENSE 和 NOTICE.md。使用或二次分发时，应保留许可证文件，不要把现有角色和规则改名后包装成完全无来源的新内容。

MIT 许可证并不自动解决生成图片的商用授权、参考图版权、字体版权和客户素材授权。Skill 开源和图片可商用是两件不同的事情，企业发布前仍需单独审核。

## 十、第一次安装验收清单

### 文件验收

- [ ] SKILL.md 位于 Codex 的 skills/xiaoen-ai-observer-illustrations/ 下。
- [ ] assets/character/ 下有无文字角色参考图和带标签参考图。
- [ ] references/ 下的风格、角色、构图、Prompt 和 QA 文件都存在。
- [ ] 没有把仓库外层目录多复制一层。

### 调用验收

- [ ] 使用 $xiaoen-ai-observer-illustrations 能触发 Skill。
- [ ] 能对短文章输出认知锚点和 Shot List。
- [ ] 能明确要求“先不要生图”。
- [ ] 能生成 16:9、白底、黑色手绘和少量颜色批注的单张配图。
- [ ] 角色参考图能被正确读取，未变成猫、兔、机器人或普通吉祥物。

### 内容验收

- [ ] 每张图只有一个核心认知和一个主要隐喻。
- [ ] 小恩正在做一个必要动作，而不是站在画面旁边。
- [ ] 文字标注少而短，画面保留大块留白。
- [ ] 没有复制旧案例的传送带、漏斗、桥、鱼或其他固定构图。
- [ ] 需要批量接入企业内容系统时，已评估通过 上游 API 做模型统一入口和成本治理。

## 总结

小恩这个项目真正有价值的地方，不只是“做了一只黑色手绘角色”，而是把下面这条链路公开出来了：

~~~text
个人审美和角色性格
  -> 固定视觉 DNA
  -> 核心动作和原创隐喻
  -> Prompt 模板和参考图
  -> 生成后 QA
  -> 可反复调用的 Codex Skill
~~~

当 IP 的审美、性格、动作和判断都被写成规则，它就不再只是一张固定形象，而开始变成一种可复用的创作能力。

## 结论

本文给出了问题定位、配置或创作流程的可执行路径。实际结果仍取决于当前版本、权限和运行环境，提交前应按官方文档复核可变字段，并保留失败证据和回滚边界。
