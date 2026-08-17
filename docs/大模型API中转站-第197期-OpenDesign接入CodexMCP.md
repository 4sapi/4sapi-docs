---
title: "【大模型API中转站】第197期 Open Design接Codex | MCP设计工作流"
category: 人工智能
tags:
  - 大模型API中转站
  - Open Design
  - Codex
  - MCP
  - Agent工作流
  - 企业级大模型接入
  - 企业API网关
description: "Open Design 可以通过 MCP 接入 Codex，把设计项目、HTML 原型、设计系统和生成任务暴露给 coding agent。本文讲它适合做什么、怎么接入、企业场景如何用 4SAPI 做统一入口、权限审计和成本治理。"
---

# 【大模型API中转站】第197期 Open Design接Codex | MCP设计工作流

本文是【大模型API中转站】系列的第197篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

这篇写一个很适合企业团队试点的方向：

```text
Open Design + Codex + MCP
```

它解决的不是“模型能不能写代码”。

而是：

```text
设计需求怎么变成可运行原型。
原型文件怎么被 Agent 继续理解。
设计系统怎么进入生成链路。
生成过程怎么被日志、权限和成本治理接住。
```

如果你已经在用 Codex、Claude Code、Cursor 或 OpenCode 做开发，Open Design 更像是一个本地设计工作台。

它把：

```text
项目。
设计系统。
技能。
HTML / CSS / JSX artifact。
生成任务。
```

整理成可以被 MCP 调用的结构化接口。

Codex 不再只看一堆散落文件。

而是可以通过工具问：

```text
有哪些项目？
当前打开的是哪个设计？
这个 artifact 引用了哪些文件？
能不能新建一个项目？
能不能让 Open Design 自己发起一次设计生成？
```

这就是 MCP 的价值。

## 1. Open Design 具体是干什么的

简单说：

```text
Open Design 是一个本地优先的 AI 设计工作区。
```

它不是传统 Figma 那种纯画布工具。

它更偏向：

```text
用 Agent 生成真实 HTML/CSS/JSX。
用设计系统约束风格。
用技能定义产出类型。
用预览窗口查看 artifact。
用 MCP 把设计文件交给其他 coding agent。
```

适合的任务包括：

```text
SaaS 落地页。
后台 dashboard。
移动端原型。
演示 deck。
运营活动页。
数据看板。
品牌视觉方案。
```

它的核心不是“帮你画一张图”。

而是：

```text
让设计变成代码文件。
让代码文件继续被 Agent 编辑。
让交付物能下载、预览、复用。
```

对企业来说，这一点很关键。

因为企业内部设计工作流经常卡在：

```text
需求文档和设计稿割裂。
设计系统没有进入模型上下文。
Agent 生成页面风格不稳定。
原型交付后工程无法继续接。
生成成本无法统计。
```

Open Design 的思路是把这些东西放到同一个项目里。

## 2. 为什么要接到 Codex

Codex 本身适合做代码理解、修改、运行和排查。

Open Design 适合做设计生成、预览和 artifact 管理。

两者接起来以后，工作流变成：

```text
Open Design 负责生成和管理设计项目。
Codex 负责理解项目、修改文件、接入现有工程。
MCP 负责让两边说同一种工具语言。
```

比如你可以让 Codex 做这些事：

```text
读取当前 Open Design 项目。
拉取完整 artifact bundle。
把 landing page 拆成 React 组件。
检查 CSS 是否符合品牌规范。
把 dashboard 原型接入真实 API。
根据设计文件继续改交互细节。
```

这比复制粘贴一个 HTML 文件稳很多。

因为 MCP 工具能返回：

```text
entry file。
被引用的 CSS。
被引用的 JS。
同级资源。
项目元数据。
预览地址。
```

Codex 拿到的是一组结构化设计上下文，而不是一段孤立代码。

## 3. MCP 接入后的工具能力

Open Design 接入 Codex 后，常见工具大概分四类。

第一类是读项目：

```text
list_projects
get_active_context
get_project
list_files
get_file
search_files
get_artifact
```

最常用的是 `get_artifact`。

因为它会把入口文件和相关依赖一起拉出来。

如果你要让 Codex 改一个设计，不建议一上来多次 `get_file`。

先：

```text
get_artifact
```

再判断要改哪些文件。

第二类是写文件：

```text
create_artifact
write_file
delete_file
```

这类工具适合让 Codex 对 Open Design 项目做小步迭代。

比如：

```text
把按钮文案改成企业版。
补一个空状态。
把颜色替换成品牌色。
增加一个移动端断点。
```

第三类是跑生成任务：

```text
create_project
list_skills
list_plugins
list_agents
start_run
get_run
cancel_run
```

这类工具不是让 Codex 自己写完所有页面。

而是让 Codex 委托 Open Design 启动一次设计生成。

流程是：

```text
create_project
list_skills
start_run
get_run
get_artifact
```

第四类是治理类间接能力。

虽然 MCP 工具本身不负责计费，但它让每次设计生成有明确的：

```text
project
runId
agent
model
artifact
previewUrl
```

这些字段进入企业 API 网关和 4SAPI 日志后，才能做成本治理。

## 4. 企业级接入时要加哪一层

个人使用可以很简单：

```text
本地启动 Open Design。
把 MCP 添加到 Codex。
让 Codex 读取当前设计项目。
```

企业接入不能只看能不能跑。

要看：

```text
谁能发起设计生成？
能调用哪些模型？
哪些项目可以用高级模型？
一次任务最大预算是多少？
生成文件是否能进入代码仓库？
日志里能否追到责任人？
```

建议架构是：

```text
设计需求
  ↓
Open Design 项目
  ↓
Codex / Claude Code / OpenCode
  ↓
企业API网关 / 4SAPI
  ↓
Claude / GPT / Gemini / DeepSeek 等模型
```

4SAPI 在中间的价值不是“多一层转发”。

而是：

```text
统一模型入口。
统一 API Key 分组。
统一日志审计。
统一成本统计。
统一模型路由。
统一失败告警。
```

设计类任务很容易超预算。

因为一次需求可能包含：

```text
需求理解。
视觉方向探索。
页面生成。
自我审查。
多轮修改。
导出交付。
```

如果每一步都直接用高级模型，成本会涨得很快。

## 5. 推荐的接入步骤

第一步，准备本地环境：

```text
Node 24。
pnpm。
Codex CLI。
Open Design 源码或桌面版。
```

第二步，启动 Open Design：

```text
pnpm install
pnpm tools-dev start web
pnpm tools-dev status
```

确认能看到：

```text
daemon: running
web: running
```

第三步，确认 MCP 安装信息。

可以看 Open Design 设置页里的 MCP snippet。

也可以走命令行：

```text
od mcp install codex --print
```

第四步，写入 Codex MCP 配置。

如果自动安装失败，可以用 Codex 自己的命令添加：

```text
codex mcp add open-design -- <node> <open-design-cli> mcp
```

第五步，验证：

```text
codex mcp get open-design
```

能看到：

```text
transport: stdio
command: node
args: ... cli.js mcp
```

第六步，新开 Codex 会话。

MCP 工具通常需要新线程或重启后才会出现在工具列表里。

## 6. 生产环境不要急着全员开放

Open Design 这类工具很适合创新，但企业里不要第一天就全员放开。

建议先做小范围试点：

```text
一个产品线。
一个设计系统。
一个内部项目。
一个固定模型组。
一个预算上限。
```

先让团队跑通：

```text
需求 → 原型 → 代码接入 → Review → 发布
```

再扩大范围。

尤其要限制：

```text
高级模型调用次数。
单任务最大上下文。
单项目每日预算。
可写入的目录范围。
可访问的 MCP 工具。
```

设计生成不是低风险动作。

它可能会读取：

```text
品牌资料。
内部产品截图。
客户案例。
运营数据。
未发布功能。
```

所以要有权限、审计和计费一体化。

## 7. 给 AI 的接入检查 Prompt

```text
你是企业级大模型接入架构师。

请根据下面信息，评估 Open Design 接入 Codex MCP 是否适合进入团队试点：

【环境】
- Open Design 启动方式：
- Codex 版本：
- MCP server 名称：
- daemon URL：
- web URL：

【工具】
- 是否能 list_projects：
- 是否能 get_artifact：
- 是否能 create_project：
- 是否能 start_run：

【企业治理】
- API Key 是否按项目分组：
- 是否接入 4SAPI：
- 是否记录 request_id / runId / project：
- 是否设置预算：
- 是否限制高级模型调用：
- 是否有日志审计：

请输出：
1. 是否可以试点。
2. 最大风险。
3. 上线前必须补的配置。
4. 推荐试点范围。
5. 不建议开放的能力。
```

这个 Prompt 适合给技术负责人做上线前审查。

## 8. 常见误区

误区一：

```text
把 Open Design 当成一个普通网页生成器。
```

它真正有价值的地方在于项目化、设计系统和 MCP。

误区二：

```text
让 Codex 绕开 Open Design 自己重写全部文件。
```

这样会丢掉 Open Design 的技能、预览和项目上下文。

更好的方式是：

```text
先让 Open Design 生成。
再让 Codex 修改。
再把结果接入工程。
```

误区三：

```text
所有设计任务都用最贵模型。
```

设计任务可以分层：

```text
低成本模型整理 brief。
中等模型生成初稿。
高级模型做最终审查。
Codex 做工程落地。
```

误区四：

```text
没有日志就开始规模化。
```

没有日志，就无法回答：

```text
谁花了钱。
哪个项目花了钱。
哪个模型花了钱。
哪次生成失败了。
为什么失败。
```

## 9. 企业上线检查清单

```text
Open Design 是否能稳定启动。
Codex 是否能看到 open-design MCP。
MCP 工具是否能返回 tools/list。
是否能读取当前项目。
是否能拉取 artifact bundle。
是否限制写文件范围。
是否设置项目级预算。
是否接入 4SAPI 日志。
是否记录 runId 和 request_id。
是否有失败告警。
是否有人工 Review 节点。
是否能导出或迁移到工程仓库。
```

如果这些都没有，先不要全员推广。

先让一个小团队把链路跑通。

## 10. 总结

Open Design 接 Codex 的核心价值是：

```text
把设计生成从一次性输出，变成可读取、可修改、可追踪的项目工作流。
```

MCP 让 Codex 能理解 Open Design 里的项目和 artifact。

4SAPI 让企业能管理模型入口、Key 权限、日志审计和成本预算。

一句话：

```text
个人看效率，企业看治理；Open Design 负责生成，Codex 负责落地，4SAPI 负责把入口和成本管住。
```
