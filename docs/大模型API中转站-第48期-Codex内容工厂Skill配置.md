---
title: "【大模型API中转站】第48期 Codex内容工厂 | 7类Skill配置"
category: 人工智能
tags:
  - 大模型API中转站
  - Codex
  - Skill
  - 内容工厂
  - 创作者工具
  - 4SAPI
description: "把公众号、小红书、短视频、知识付费、商业广告、产品号和个人IP的 Codex Skill 配置拆成 7 套工作流，重点讲如何从提示词升级到可复用内容生产系统。"
---

# 【大模型API中转站】第48期 Codex内容工厂 | 7类Skill配置

本文是【大模型API中转站】系列的第48篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

很多人用 AI 做内容，第一反应还是写一段很长的提示词：

```text
帮我写一篇公众号文章。
帮我做一组小红书图文。
帮我把这场直播整理成短视频脚本。
```

这样当然能用，但很难稳定。

真正的问题不是模型不会写，而是每一种创作者需要的流程完全不同。

写公众号长文的人，需要调研、结构、事实复核和口吻编辑。

做小红书图文的人，需要封面、卡片结构、主题色和素材管理。

做短视频的人，需要钩子、分镜、字幕、切片和素材归档。

做课程的人，需要大纲、课件、作业表、学员反馈和答疑整理。

做商业广告的人，需要竞品拆解、卖点提炼、线索跟进和转化素材。

做 AI 产品号的人，需要版本更新、工具测评、网页测试和开发记录。

做个人 IP 或矩阵号的人，需要统一人设、团队协作、素材库和复盘机制。

所以，把 Codex 变成内容工厂，不是给它一条“万能爆款提示词”，而是给不同创作者配置不同的 Skill 工作流。

这一篇先做总览，后面几篇再按场景拆细。

## 1. Skill 的价值：把提示词升级成流程

提示词解决的是一次对话。

Skill 解决的是一类任务。

一个好的 Skill 通常不只是几句说明，而是把任务拆成了稳定结构：

| 组成 | 作用 |
| --- | --- |
| `SKILL.md` | 告诉 Codex 什么时候触发、怎么做、输出什么 |
| `references/` | 保存风格规范、示例、检查清单和行业知识 |
| `scripts/` | 把重复动作脚本化，比如导出、整理、渲染、校验 |
| `assets/` | 放模板、主题、样例、封面素材或交付物骨架 |

这就是 Skill 和提示词最大的区别。

提示词像临时交代任务，Skill 像给编辑、设计、运营、助理各自写了一份 SOP。

对内容创作者来说，最重要的不是“AI 能不能写”，而是它每次进入任务时，能不能先读懂你的栏目定位、素材结构、平台格式和审核边界。

## 2. 最基础的安装方式

如果你使用的是本地 Codex Skill，最常见的做法是把 Skill 文件夹放到：

```text
~/.codex/skills/
```

Windows 通常对应：

```text
C:\Users\你的用户名\.codex\skills\
```

一个典型流程是：

```text
找到需要的 Skill
  -> 复制整个 Skill 文件夹
  -> 放入 ~/.codex/skills/
  -> 重启 Codex
  -> 在任务里直接说出目标
```

比如你可以这样说：

```text
按我的公众号风格写一篇长文。
把这篇文章拆成小红书图文。
把这个主题做成 PPT 课件。
把这个竞品广告拆成卖点结构。
把这场直播整理成 10 条短视频脚本。
```

如果 Skill 写得清楚，Codex 会根据任务去读对应规则，而不是每次都让你重新解释一遍。

这里要注意：安装 Skill 不等于自动发布，也不等于自动通过平台审核。Skill 只是把工作流固化下来，最终事实、版权、商业承诺和发布责任仍然要由人来确认。

## 3. 7 类创作者应该配不同 Skill

下面这张表可以先收藏。

| 创作者类型 | 核心任务 | 优先 Skill |
| --- | --- | --- |
| 公众号 / 知乎 / X 长文 | 调研、搭框架、写稿、润色、归档 | `content-research-writer`、`notion-research-documentation`、`notion-knowledge-capture`、`helium-mcp`、`stop-slop` |
| 小红书 / 图文卡片 / 封面 | 卡片结构、视觉规范、封面、长图 | `canvas-design`、`image-enhancer`、`brand-guidelines`、`theme-factory`、`template-skill` |
| 短视频 / 直播切片 | 下载素材、拆钩子、写脚本、做字幕 | `video-downloader`、`competitive-ads-extractor`、`meeting-notes-and-actions`、`meeting-insights-analyzer`、`slack-gif-creator` |
| 知识付费 / 课程 / PPT | 大纲、课件、讲稿、作业表、答疑 | `paperjsx`、`content-research-writer`、`spreadsheet-formula-helper`、`support-ticket-triage`、`canvas-design` |
| 商业广告 / 带货 / 私域 | 竞品拆解、卖点、线索、邮件、报价 | `competitive-ads-extractor`、`lead-research-assistant`、`email-draft-polish`、`domain-name-brainstormer`、`paperjsx` |
| AI 工具 / 产品号 / 科技号 | 测评、教程、更新日志、网页测试 | `changelog-generator`、`webapp-testing`、`mcp-builder`、`create-plan`、`gh-fix-ci` |
| 个人 IP / 矩阵号 / 内容团队 | 品牌统一、团队协作、素材库、复盘 | `brand-guidelines`、`skill-share`、`internal-comms`、`file-organizer`、`skill-creator` |

不要一上来就装满 70 个 Skill。

更稳的做法是先找自己的瓶颈。

如果你经常不知道写什么，先装调研和选题类。

如果你经常写完不像人话，先装口吻和去模板味类。

如果你经常发出去不好看，先装视觉和卡片类。

如果你经常素材混乱，先装知识库和文件整理类。

如果你是团队协作，先装品牌规范、会议纪要、内部沟通和 Skill 共享类。

## 4. 为什么要配大模型API中转站

当你只用一个聊天窗口时，模型入口不复杂。

但一旦开始 Skill 化，调用链会变长：

```text
收集资料
  -> 整理大纲
  -> 写初稿
  -> 改口吻
  -> 做封面
  -> 生成课件
  -> 拆短视频脚本
  -> 归档素材
  -> 复盘数据
```

每一步对模型能力的要求都不同。

| 环节 | 推荐模型策略 |
| --- | --- |
| 摘要、分类、命名、归档 | 低成本模型 |
| 调研、对比、事实核查 | 强推理或联网能力模型 |
| 中文长文初稿 | 中等到强文本模型 |
| 去 AI 味、改口吻 | 中文表达能力好的模型 |
| 代码化生成 HTML、PPT、表格 | 代码能力强的模型 |
| 封面、视觉、配图 | 多模态或图像模型 |

4SAPI 这类大模型API中转站适合放在中间，做统一入口：

```text
Codex / Skill
  -> 4SAPI 统一模型入口
  -> 不同厂商和不同能力模型
```

这样做有三个价值。

第一，Key 不乱。每个工具不用单独塞一把 Key。

第二，成本可见。你能知道到底是调研、出图还是改稿花了钱。

第三，模型可替换。某个环节效果不好，可以只替换这个环节的模型，而不是推翻整套工作流。

## 5. 一套最小内容工厂配置

如果你刚开始，不建议从“7 类创作者全覆盖”做起。

可以先搭一套最小内容工厂：

```text
输入层：notion-knowledge-capture / file-organizer
调研层：content-research-writer / helium-mcp
写作层：content-research-writer / brand-guidelines
编辑层：stop-slop
视觉层：canvas-design / image-enhancer
交付层：paperjsx
复盘层：spreadsheet-formula-helper
```

对应到真实流程就是：

```text
灵感进入素材库
  -> 选题前做资料整理
  -> 写公众号长文
  -> 清理 AI 味
  -> 拆小红书卡片和短视频脚本
  -> 做 PPT 或 PDF 资料包
  -> 记录阅读、转发、成交和反馈
```

这条链路跑顺以后，再去增加更细的 Skill。

比如你做商业广告，再加 `competitive-ads-extractor` 和 `lead-research-assistant`。

比如你做课程，再加 `support-ticket-triage` 和 `meeting-notes-and-actions`。

比如你做产品号，再加 `changelog-generator` 和 `webapp-testing`。

## 6. 建议的项目目录

内容工厂一定要有目录结构。

否则 Skill 越多，文件越乱。

可以先用这个版本：

```text
content-factory/
  inbox/
    ideas.md
    raw-links.md
  research/
    sources.md
    reports/
  drafts/
    longform/
    xiaohongshu/
    scripts/
  assets/
    images/
    covers/
    slides/
  publish/
    wechat/
    x/
    xiaohongshu/
    video/
  review/
    metrics.xlsx
    feedback.md
    weekly-review.md
  rules/
    brand-guidelines.md
    fact-checklist.md
    platform-rules.md
```

每个 Skill 都应该围绕这个目录工作。

不要把文章、封面、数据、直播转写、客户反馈混在一个桌面文件夹里。那样短期省事，长期一定会变成素材灾难。

## 7. 成本与风险提示

Skill 化以后，效率会提升，但风险也会放大。

最容易踩的坑有四个。

| 风险 | 表现 | 建议 |
| --- | --- | --- |
| 成本失控 | 调研和改稿反复循环 | 给每类任务设置模型和轮次上限 |
| 内容失真 | AI 编造来源、数据或案例 | 发布前做事实复核，重要数据保留来源 |
| 版权混乱 | 直接搬运文章、图片、视频素材 | 只处理有权使用的资料，保留来源 |
| 风格漂移 | 今天一个口吻，明天一个人设 | 用 `brand-guidelines` 固定账号规范 |

这里尤其要强调：大模型API中转站能帮你统一调用、控制额度和查看日志，但不能替你承担内容责任。

AI 可以负责整理、生成、改写和适配格式。

人必须负责判断、取舍、事实核查、审美把关和发布责任。

## 8. 总结

普通人用 AI，是每次重新指挥。

高手用 Skill，是把自己的内容流程固化下来。

你一天写 1 篇，是在做内容。

你把选题、资料、写稿、配图、排版、分发、复盘全部 Skill 化，才是在搭内容生产系统。

对独立创作者来说，先从一条最小链路开始：

```text
素材库
  -> 调研
  -> 写稿
  -> 改口吻
  -> 视觉分发
  -> 数据复盘
```

对团队来说，再把 Key、模型、成本、目录、审核和权限接入 4SAPI 这类统一入口。

下一篇继续拆公众号、知乎和 X 长文创作者：如何用 Codex Skill 把“选题、调研、写稿、润色、归档”串成一条稳定产线。
