---
title: "【大模型API中转站】第33期 内容创作Skill选型 | 10个开源工具"
category: 人工智能
tags:
  - 大模型API中转站
  - Codex
  - Claude Code
  - 内容创作
  - Skill
  - 4SAPI
description: "把 10 个开源内容创作 Skill 拆成选题、调研、写稿、去 AI 味、配图、PPT、公众号、视频切片和营销增长九类能力，给独立创作者和小团队一套可落地的选型方法。"
---

# 【大模型API中转站】第33期 内容创作Skill选型 | 10个开源工具

本文是【大模型API中转站】系列的第33篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

过去一年，很多人对 AI 写作的理解还停留在“给模型一个题目，让它吐一篇文章”。这个用法太浅了。

真正好用的内容创作工作流，不是一个万能提示词，而是一条产线：搜资料、定选题、搭大纲、写初稿、改口吻、做配图、转 PPT、切短视频、排版发布、复盘数据。每个环节都有不同的输入、输出和质量标准。

这一篇先做总览。我们把最近开源热度很高的 10 个内容创作 Skill 放在一张图里看，重点不是吹“哪个最强”，而是回答三个更实用的问题：

1. 这个 Skill 解决内容产线里的哪一段？
2. 适合什么创作者、团队和平台？
3. 如果接入 Codex、Claude Code 或 4SAPI 这类大模型API中转站，怎么控制成本和质量？

本文提到的 star 数为我在 2026-06-17 通过 GitHub API 查询到的约数，后续会变化，请以项目主页为准。

## 1. 先看 10 个 Skill 放在哪个环节

把内容创作拆开看，会比直接问“哪个 Skill 好用”清楚很多。

```text
选题/调研
  -> Deep-Research-skills
  -> oh-story-claudecode
  -> marketingskills

写稿/改稿
  -> wewrite
  -> Humanizer-zh

配图/视觉
  -> awesome-gpt-image-2
  -> guizang-social-card-skill
  -> guizang-ppt-skill

复用/分发
  -> anything-to-notebooklm
  -> Youtube-clipper-skill
```

如果你只做公众号长文，优先看 Deep-Research-skills、wewrite、Humanizer-zh。  
如果你做小红书、公众号封面和信息图，优先看 guizang-social-card-skill、awesome-gpt-image-2。  
如果你做课程、汇报、分享，优先看 guizang-ppt-skill、anything-to-notebooklm。  
如果你做视频切片和二次传播，优先看 Youtube-clipper-skill。  
如果你做出海、外贸、增长文案，marketingskills 的价值会更明显。

## 2. 10 个 Skill 快速选型表

| Skill | 截至 2026-06-17 约 star | 核心用途 | 最适合的人 | 注意点 |
| --- | ---: | --- | --- | --- |
| [guizang-ppt-skill](https://github.com/op7418/guizang-ppt-skill) | 17.7k | 生成单文件 HTML 演示稿、配图和封面 | 做汇报、课程、产品分享的人 | HTML PPT 很适合演示和截图，不等于传统 PPTX 协作文件 |
| [guizang-social-card-skill](https://github.com/op7418/guizang-social-card-skill) | 3.6k | 小红书图文、公众号 21:9 + 1:1 封面对 | 图文号、公众号、小红书运营 | 依赖浏览器渲染 PNG，最好保留人工审图 |
| [awesome-gpt-image-2](https://github.com/freestylefly/awesome-gpt-image-2) | 7.6k | GPT-Image2 提示词案例库和模板 | 需要稳定出图的创作者 | 它更像风格库和提示词工程，不是直接替你审美 |
| [Humanizer-zh](https://github.com/op7418/Humanizer-zh) | 10.4k | 中文文本去 AI 味、改成人话 | 写公众号、博客、营销文案的人 | 不要用来掩盖虚假内容，核心是提升表达质量 |
| [Deep-Research-skills](https://github.com/Weizhena/Deep-Research-skills) | 1.1k | 结构化深度调研，先大纲再分项调查 | 写干货长文、技术评测、选型报告的人 | 调研成本高，要限制范围和来源 |
| [anything-to-notebooklm](https://github.com/joeseesun/qiaomu-anything-to-notebooklm) | 5.2k | 多来源内容转 NotebookLM，再生成播客、PPT、思维导图、测验 | 做内容复用、课程资料整理的人 | 只处理你有权访问和使用的资料，不做违规抓取 |
| [wewrite](https://github.com/oaker-io/wewrite) | 2.4k | 公众号全流程，热点、选题、写作、SEO、排版、草稿箱 | 公众号运营、个人媒体 | 推送草稿箱前必须人工复核事实和口吻 |
| [Youtube-clipper-skill](https://github.com/op7418/Youtube-clipper-skill) | 2.0k | YouTube 长视频语义分章、剪辑、字幕、烧录 | 视频号、课程切片、技术访谈二创 | 注意版权、授权和平台规则 |
| [oh-story-claudecode](https://github.com/worldwonderer/oh-story-claudecode) | 2.6k | 网文小说扫榜、拆文、选题、人设、开篇 | 小说作者、网文工作室 | 适合研究套路，不适合抄袭具体作品 |
| [marketingskills](https://github.com/coreyhaines31/marketingskills) | 33.6k | 英文营销技能库，CRO、SEO、文案、增长 | 出海团队、独立站、外贸内容 | 英文语境强，中文使用时要本地化 |

这一张表可以先收藏。你不需要一口气装完 10 个 Skill，先从自己最卡的环节开始。

## 3. 为什么 Skill 比“提示词大全”更适合内容团队

提示词大全的问题是散。

今天复制一个“小红书爆款标题提示词”，明天复制一个“公众号深度文提示词”，后天再复制一个“PPT 生成提示词”。短期看很爽，长期会遇到三个问题：

| 问题 | 表现 | 后果 |
| --- | --- | --- |
| 没有流程 | 每次从零开始问 | 输出质量不稳定 |
| 没有文件结构 | 图片、草稿、参考资料散落 | 很难复盘和复用 |
| 没有边界 | 模型不知道什么能发、什么要复核 | 容易出事实和合规问题 |

Skill 的价值在于把“提示词”升级成“工作流”。它通常会包含：

- `SKILL.md`：告诉 Agent 什么时候触发、怎么做、什么不能做。
- `references/`：沉淀设计原则、写作规则、版式说明、风格样例。
- `scripts/`：把下载、渲染、校验、发布这类重复动作脚本化。
- `assets/`：模板、主题、布局骨架、示例文件。

这就像从“口头交代实习生”变成“给一个有 SOP 的编辑”。模型还是模型，但它每次进来会先读规矩，输出更容易稳定。

## 4. 用 4SAPI 做统一模型入口有什么价值

内容创作工作流里，不同环节对模型能力的要求差别很大。

| 环节 | 任务类型 | 模型要求 | 成本策略 |
| --- | --- | --- | --- |
| 资料清洗 | 摘要、分类、去重 | 便宜稳定即可 | 用低成本模型批量处理 |
| 深度调研 | 找资料、交叉验证、生成报告 | 需要推理和联网能力 | 限制问题范围，按项目调用强模型 |
| 写初稿 | 结构、表达、案例串联 | 需要中文表达能力 | 中等模型即可，保留人工改稿 |
| 去 AI 味 | 句式、口吻、节奏 | 需要中文编辑能力 | 只处理最终稿，减少重复调用 |
| 配图提示词 | 风格、构图、比例、素材约束 | 需要多模态和视觉理解 | 先出低清样张，再定稿 |
| PPT 和卡片 | HTML、CSS、版式验证 | 需要代码和视觉综合能力 | 一次生成，多轮小改 |

4SAPI 这类大模型API中转站适合放在中间做统一入口。团队不用把每个工具都接一遍不同厂商的 Key，而是统一管理模型、额度、日志和调用策略。

一个实用的路由规则可以这样写：

```text
低成本模型：
  文档摘要、评论归类、关键词提取、标题备选

中等模型：
  公众号初稿、选题分析、文章改写、脚本整理

强推理模型：
  深度调研、技术对比、事实核验、复杂大纲

多模态/生图模型：
  封面、配图、PPT 视觉、海报、截图再设计
```

不要把所有任务都扔给最贵模型。内容产线最容易浪费钱的地方，就是用强模型做机械活。

## 5. 一条完整内容产线怎么搭

如果你是一个独立创作者，建议先搭一条最小可用的内容产线。

```text
输入：
  热点、资料链接、产品资料、访谈录音、视频链接

处理：
  Deep-Research-skills 做资料调研
  wewrite 做公众号结构和初稿
  Humanizer-zh 做中文口吻编辑
  awesome-gpt-image-2 提供配图风格
  guizang-social-card-skill 生成社交卡片
  guizang-ppt-skill 转成分享 PPT
  Youtube-clipper-skill 切视频素材
  anything-to-notebooklm 把资料再拆成播客、PPT、思维导图

输出：
  公众号文章、小红书图文、PPT、短视频、播客大纲、知识卡片
```

注意，这不是让 AI 把你变成全自动内容工厂。真正可持续的流程应该是：

```text
AI 负责整理、生成、改写、适配格式
人负责判断、取舍、事实复核、审美把关、发布责任
```

这条边界一定要写进你的工作流。

## 6. 不同创作者怎么选

### 6.1 技术博客作者

优先组合：

| 环节 | 推荐 Skill |
| --- | --- |
| 深度调研 | Deep-Research-skills |
| 写技术长文 | wewrite 或自定义博客提示词 |
| 去 AI 味 | Humanizer-zh |
| 配图 | awesome-gpt-image-2 |
| 转演示稿 | guizang-ppt-skill |

技术博客最怕“看起来很懂，实际没有验证”。Deep-Research-skills 适合先列出研究对象和字段，再查资料；写完以后用 Humanizer-zh 去掉空话和三段式模板感。

### 6.2 公众号运营

优先组合：

| 环节 | 推荐 Skill |
| --- | --- |
| 热点和选题 | wewrite |
| 资料补充 | Deep-Research-skills |
| 正文写作 | wewrite |
| 口吻修订 | Humanizer-zh |
| 封面和配图 | guizang-social-card-skill |

公众号不只是“写文章”。它还有标题、摘要、SEO、封面、排版和草稿箱。wewrite 的价值在于把这些环节串起来，但越接近发布，越需要人工复核。

### 6.3 小红书图文号

优先组合：

| 环节 | 推荐 Skill |
| --- | --- |
| 爆款图风格参考 | awesome-gpt-image-2 |
| 竖版轮播图 | guizang-social-card-skill |
| 文案人话化 | Humanizer-zh |
| 多平台复用 | anything-to-notebooklm |

小红书图文的难点不是“写 500 字”，而是把信息切成一屏一屏能读的节奏。guizang-social-card-skill 的 3:4 画板和多版式骨架正好解决这个问题。

### 6.4 课程和知识付费团队

优先组合：

| 环节 | 推荐 Skill |
| --- | --- |
| 资料整理 | anything-to-notebooklm |
| 课程研究 | Deep-Research-skills |
| 课件生成 | guizang-ppt-skill |
| 课程切片 | Youtube-clipper-skill |
| 复习题 | anything-to-notebooklm |

课程团队通常有大量资料、录音、PDF 和长视频。anything-to-notebooklm 更适合做内容再加工，guizang-ppt-skill 更适合做公开分享用的视觉化课件。

### 6.5 出海和独立站团队

优先组合：

| 环节 | 推荐 Skill |
| --- | --- |
| 营销框架 | marketingskills |
| 英文文案 | marketingskills |
| 中文改写 | Humanizer-zh |
| 产品图风格 | awesome-gpt-image-2 |
| 演示稿 | guizang-ppt-skill |

marketingskills 是英文语境里的营销技能库，覆盖文案、SEO、转化、分析、增长等方向。拿来做中文内容时，不要直接翻译，要把目标用户、平台语气和信任机制重新本地化。

## 7. 先装哪三个

如果你不知道从哪开始，我建议先装三个：

1. Deep-Research-skills：解决“写之前没有料”的问题。
2. Humanizer-zh：解决“写完像 AI”的问题。
3. guizang-social-card-skill 或 guizang-ppt-skill：解决“发出去不好看”的问题。

这三个覆盖了内容创作最基础的三件事：

```text
有料
  -> 像人写的
  -> 能被平台消费
```

等这条链路跑顺了，再考虑 wewrite、awesome-gpt-image-2、anything-to-notebooklm 和视频类工具。

## 8. 一个最小配置文件

建议你在内容项目里建一个 `content-workflow.md`，把使用边界写清楚。

```markdown
# 内容创作工作流

## 模型入口

- 默认通过 4SAPI 统一调用模型。
- 批量摘要和分类使用低成本模型。
- 深度调研、事实核验和复杂结构使用强推理模型。
- 生图、配图和视觉生成使用多模态或图片模型。

## 必须人工复核

- 数据、价格、法律、医疗、金融、平台规则。
- 真实人物、公司、产品和引用。
- 任何会被读者当成事实依据的内容。
- 发布前标题、封面、导语和结尾。

## 禁止事项

- 不编造来源。
- 不绕过平台限制。
- 不洗稿抄袭。
- 不把未经授权的视频、文章、图片当成自有素材发布。

## 输出规范

- 每篇文章保留资料来源。
- 每张图保留生成提示词或素材来源。
- 每次发布后记录标题、封面、阅读数据和复盘结论。
```

这个文件看起来朴素，但很有用。它能把“我今天心情好怎么写”变成可重复的编辑流程。

## 9. 成本和风险提示

内容创作 Skill 最容易踩四个坑。

| 坑 | 表现 | 处理方式 |
| --- | --- | --- |
| 调研太宽 | 问一个大题，模型查半天 | 先限定对象、字段、来源和时间范围 |
| 出图乱试 | 每个封面反复生成十几版 | 先用低成本模型确定构图，再调用生图 |
| 自动发布 | 初稿直接进草稿箱甚至发布 | 必须人工复核事实和平台合规 |
| 版权混乱 | 视频、文章、图片随便拿来二创 | 只使用有权处理的资料，保留来源 |

4SAPI 这类大模型API中转站可以帮你看见成本，但不能替你承担内容责任。模型调用越方便，越要把“能不能发”这件事交给人来判断。

## 10. 总结

这 10 个 Skill 的真正价值，不是让你少写几个提示词，而是把内容生产从“临场发挥”变成“可复用流程”。

最适合普通创作者的起步方式是：

```text
Deep-Research-skills 准备资料
  -> wewrite 或自定义提示词写初稿
  -> Humanizer-zh 改口吻
  -> awesome-gpt-image-2 定视觉风格
  -> guizang-social-card-skill / guizang-ppt-skill 做分发物料
```

下一篇我们继续拆“研究、选题、写稿、去 AI 味”这一段，重点看 Deep-Research-skills、wewrite、Humanizer-zh、oh-story-claudecode 和 marketingskills 怎么放进同一条写作流程。

