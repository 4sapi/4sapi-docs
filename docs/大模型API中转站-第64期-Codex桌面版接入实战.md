---
title: "【大模型API中转站】第64期 Codex桌面版实战 | 少踩80%接入坑"
category: 人工智能
tags:
  - 大模型API中转站
  - Codex
  - API接入
  - 开发者工具
  - AGENTS.md
  - Worktree
  - 4SAPI
  - 4sapi.com
description: "基于 OpenAI Codex 官方文档、桌面版功能说明和开发者社区经验，讲清楚如何用 Codex 桌面版完成 Claude、GPT 等模型的中转站接入、调试、Review、浏览器验证、Git 交付和长期规则配置。"
---

# 【大模型API中转站】第64期 Codex桌面版实战 | 少踩80%接入坑

本文是【大模型API中转站】系列的第64篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

这篇继续讲一个非常实际的问题：

```text
如果你刚装 Codex 桌面版，怎么用它把 Claude、GPT、Gemini 这类模型真正接进项目？
```

很多人第一次用 Codex 桌面版，会把它当成一个“会写代码的聊天框”。

于是常见用法是：

```text
帮我写一个调用 Claude API 的代码。
```

这当然能用。

但这只发挥了 Codex 桌面版很小一部分能力。

真正适合开发者的用法，是让 Codex 进入你的项目目录，读代码、看配置、改文件、跑命令、查报错、打开浏览器验证页面、看 Git diff，再和你一起把改动收口。

换句话说：

```text
普通聊天工具帮你“生成代码”。
Codex 桌面版帮你“完成交付”。
```

如果你正在做大模型 API 中转站接入，例如把项目统一接到 4sapi.com 这类中转服务，Codex 桌面版尤其适合。

因为这种任务不只是写一段请求代码，还包括：

- 找到项目里的模型调用位置
- 判断现在用的是 OpenAI SDK、Claude SDK 还是自封装接口
- 改环境变量和 base_url
- 兼容原有业务逻辑
- 写最小测试
- 跑 lint 和 test
- 检查前端页面是否还能正常对话
- 看 diff 是否泄露 Key
- 生成提交说明和 PR 描述

这才是完整的工程流。

## 1. Codex 桌面版到底适合做什么？

从 OpenAI Codex 官方文档来看，Codex 不只是一个 CLI，也不只是一个编辑器插件。它有 App、IDE Extension、CLI、Web 等多个使用界面。桌面版 Codex App 更适合做规划、审查、交互式开发和多轮协作。

你可以把它理解成一个本地 AI 开发工作台。

它能参与这些任务：

- 阅读陌生项目
- 修 bug
- 接入新 API
- 改配置
- 写测试
- 跑命令
- 看终端输出
- 做代码审查
- 打开本地网页检查效果
- 处理 Git diff
- 根据 PR review comments 修改代码
- 生成文档、脚本、配置
- 调用 Skills、Plugins、MCP、Browser、Computer Use 等扩展能力

对大模型 API 中转站来说，最常见任务有五类。

第一，接入类：

```text
把项目里的 OpenAI / Claude 调用统一改成中转站入口。
```

第二，排错类：

```text
请求失败、模型名不对、Key 无效、格式不兼容、返回解析失败。
```

第三，治理类：

```text
增加限流、重试、日志、成本统计、模型路由。
```

第四，验证类：

```text
跑测试、跑本地服务、检查聊天页面、检查移动端布局。
```

第五，交付类：

```text
看 Review、stage、commit、push、写 PR。
```

只要你按这个顺序用，Codex 的价值会明显超过“写代码快一点”。

## 2. 第一次打开项目，不要急着让它改代码

新手最容易犯的错，是一上来就说：

```text
帮我把项目接入 4SAPI。
```

这个需求太大。

Codex 可能会猜你的架构，猜你的启动命令，猜你的环境变量，最后改出一堆你不放心的东西。

更稳的第一条消息应该是：

```text
先不要改代码。
请阅读当前项目，告诉我：
1. 这个项目是做什么的
2. 主要目录结构是什么
3. API 调用集中在哪些文件
4. 是否已经接入 OpenAI、Claude 或其他大模型
5. 环境变量在哪里配置
6. 启动、测试、构建命令分别是什么
7. 如果我要接入大模型 API 中转站，应该优先看哪些文件
```

这一步不是浪费时间。

它是在让 Codex 先建立项目地图。

社区里很多开发者总结过类似经验：让 AI 编程工具先读项目、先列计划、先确认范围，后面的错误会少很多。

尤其是 API 接入任务。

因为这类任务经常散落在多个地方：

- `.env`
- `.env.example`
- `config.ts`
- `api.ts`
- `llm.ts`
- `openai.ts`
- `anthropic.ts`
- `chat.service.ts`
- `server/routes/chat.ts`
- `src/lib/ai/`
- `src/services/model/`
- 前端模型选择器
- 后端请求代理
- Docker 配置
- CI 测试脚本

如果不先定位，就很容易只改了一处，结果其他地方还在调用旧地址。

## 3. 一个完整接入流程：从官方直连切到中转站

假设你现在有一个项目，原来直接调用 OpenAI 或 Claude。

你想改成：

```text
你的应用 -> 4SAPI / 大模型 API 中转站 -> Claude / GPT / Gemini
```

建议让 Codex 按下面流程做。

### 第一步：列出所有模型调用位置

提示词：

```text
请找出项目里所有调用大模型 API 的位置。
范围包括 OpenAI、Claude、Gemini、通用 LLM、chat/completions、responses、embeddings、rerank 等关键词。

先不要修改代码。
请按文件路径列出：
1. 调用位置
2. 使用的 SDK 或 HTTP 客户端
3. 当前模型名
4. 当前 base_url 来源
5. 是否读取环境变量
6. 你建议的改造优先级
```

这里要让 Codex 只读。

不要一上来改。

因为你需要先知道项目现在到底有几条链路。

### 第二步：判断接入方式

常见有三种。

第一种，SDK 支持自定义 `baseURL`。

这是最省事的。

比如很多 OpenAI 兼容服务，只要把 `baseURL` 改成中转站地址即可。

第二种，项目里自封装了 HTTP 请求。

这种也好改，但要检查请求体和响应体格式。

第三种，项目强依赖某个官方 SDK 的特殊接口。

这种要谨慎。

有些接口不是简单 `chat/completions` 格式，可能涉及文件、工具调用、流式输出、图片、语音、结构化输出。接入前要确认中转站是否支持对应能力。

你可以让 Codex 先输出判断：

```text
请判断当前项目最适合哪种接入方式：
1. 只替换 baseURL
2. 改造 HTTP 封装
3. 新增一个 provider 层
4. 暂时不建议改

请说明原因、改动文件、风险和回滚方式。
```

### 第三步：设计环境变量

不要把 Key 写死在代码里。

推荐先统一成这样的配置：

```env
LLM_BASE_URL=https://你的中转站地址/v1
LLM_API_KEY=你的中转站Key
LLM_MODEL=claude-sonnet
LLM_PROVIDER=4sapi
LLM_TIMEOUT_MS=60000
```

如果项目里已经有配置，不要强行换名字。

更好的做法是让 Codex 保持现有风格：

```text
请沿用当前项目已有的环境变量命名风格。
如果必须新增变量，请同步更新 .env.example、README 和配置校验逻辑。
不要把真实 Key 写进任何文件。
```

### 第四步：最小改动跑通链路

第一轮目标只有一个：

```text
让一条最小聊天请求能成功返回。
```

不要同时做模型路由、成本统计、缓存、审计、UI 改版。

可以让 Codex 这样做：

```text
请按最小改动接入大模型 API 中转站。
要求：
1. 保持现有业务逻辑不变
2. 不引入新生产依赖
3. 不改无关文件
4. 新增或更新最小测试
5. 修改后运行相关测试
6. 输出回滚方法
```

这个提示词的重点是“最小改动”。

Codex 很擅长补全能力，但在工程任务里，范围控制比聪明更重要。

### 第五步：再补治理能力

链路跑通后，再考虑治理。

比如：

- 超时
- 重试
- 限流
- fallback 模型
- 用量统计
- 请求日志
- 错误分类
- 用户可读错误提示
- 敏感信息脱敏

你可以分第二轮做：

```text
现在基础调用已经跑通。
请继续补一个轻量治理层：
1. 请求超时
2. 可读错误提示
3. 不记录 API Key
4. 记录模型名、耗时和状态码
5. 避免记录完整用户隐私内容
6. 保持现有接口返回格式不变
```

这样比一次性全做稳得多。

## 4. Local、Worktree、Cloud：接入任务怎么选？

Codex 桌面版里很重要的一个概念是 Worktree。

OpenAI 官方文档里解释，Worktree 本质上是基于 Git worktree 创建的隔离工作区。它能让 Codex 在后台工作，不打扰你当前的本地 checkout。

简单理解：

| 模式 | 适合任务 | 不适合任务 |
| --- | --- | --- |
| Local | 小修、小 bug、快速验证 | 大改动、实验性重构 |
| Worktree | 新功能、接入新 provider、重构 | 依赖本地私密文件但没配置复制 |
| Cloud / Web | 长任务、远程后台任务 | 强依赖本机环境和登录态 |

做大模型 API 接入时，我建议这样选。

### 小 bug 用 Local

比如：

- 模型名写错
- 环境变量没读取
- 错误提示不好
- 某个测试失败

这些直接在 Local 修就行。

### 新接入用 Worktree

比如：

- 新增 4SAPI provider
- 支持 Claude 和 GPT 路由
- 新增模型配置页面
- 改造后端 LLM 层

建议开 Worktree。

因为你可以让 Codex 在隔离目录里试方案，当前本地开发环境不受影响。

### 注意 `.env` 和 `.worktreeinclude`

官方 Worktree 文档提到，Codex-managed worktree 是从 Git checkout 创建的，Git 忽略的本地文件默认不会自动带过去。

这对 API 接入很关键。

因为你的 `.env`、`.env.local`、`config/secrets.json` 通常都在 `.gitignore` 里。

如果 Worktree 需要这些文件，可以在仓库根目录加 `.worktreeinclude`：

```gitignore
.env
.env.local
config/secrets.json
```

但要非常谨慎。

不要把不该复制的敏感文件随便列进去。

更稳的做法是复制测试 Key 或 sandbox 配置，而不是生产 Key。

## 5. 集成终端：让 Codex 看真实错误

OpenAI 官方文档里明确提到，Codex App 每个线程都有集成终端，终端作用域在当前项目或 worktree。Codex 可以读取当前终端输出，所以它能根据失败的测试、构建错误、开发服务状态继续处理。

这就是桌面版和普通聊天工具的关键差别。

普通聊天工具只能靠你复制报错。

Codex 可以自己运行：

```bash
npm test
npm run lint
npm run build
git status
git diff
```

API 接入完成后，你应该让它至少跑三类验证。

第一类，配置验证：

```text
请确认环境变量读取正常。
如果项目有配置校验，请运行相关测试。
如果没有，请写一个最小测试覆盖 LLM_BASE_URL、LLM_API_KEY、LLM_MODEL。
```

第二类，调用验证：

```text
请写一个最小模型调用测试，不要使用真实用户数据。
验证请求能发出、返回能解析、错误能被捕获。
```

第三类，回归验证：

```text
请运行现有 lint、test、build。
如果失败，请先判断是否由本次改动导致，再给出修复计划。
```

这里有个很重要的要求：

```text
不要只让 Codex “看起来改完了”，一定要让它跑起来。
```

## 6. Review 面板：不要跳过 diff

Codex App 的 Review 面板是很多新手忽略的功能。

官方文档说明，Review 面板展示的是 Git 仓库当前状态，不只包括 Codex 改的内容，也包括你自己或其他工具产生的未提交修改。

这点非常重要。

因为在 API 接入任务里，最怕几件事：

- 把真实 API Key 写进代码
- 改了无关文件
- 删除了原有错误处理
- 破坏了旧 provider
- 引入了不必要依赖
- 忘记更新 `.env.example`
- 忘记更新 README
- 测试文件写死真实模型输出

Codex 改完以后，你应该打开 Review 面板，按这张清单看：

| 检查项 | 重点 |
| --- | --- |
| 配置文件 | 是否只放示例变量，不放真实 Key |
| 代码文件 | 是否保持原有接口返回格式 |
| 测试文件 | 是否使用 mock 或测试 Key |
| 文档文件 | 是否说明 base_url、model、key |
| 依赖文件 | 是否新增生产依赖 |
| 日志逻辑 | 是否记录敏感 prompt 或 Key |
| 错误处理 | 是否给用户可读提示 |

官方 Review 文档还提到，你可以对 diff 的具体行留下 inline comments，然后让 Codex 只处理这些评论。

这是非常适合收口的方式：

```text
我已经在 Review 面板留下了 inline comments。
请只处理这些评论，保持修改范围最小，不要重构其他代码。
```

如果你要让 Codex 自己审查当前改动，可以用：

```text
请 review 当前未提交修改。
重点看：
1. API Key 泄露风险
2. 环境变量配置错误
3. 中转站 base_url 兼容性
4. 请求和响应格式变化
5. 错误处理和重试
6. 缺失测试

先列问题，不要直接修改。
```

这里要注意：代码审查和代码修改最好分开。

先审，再改。

## 7. Browser、Chrome Extension、Computer Use 怎么选？

如果你的项目有前端页面，Codex 的浏览器能力非常有用。

官方 In-app browser 文档说明，它适合打开本地开发服务器、file 预览和不需要登录的公开页面。它不支持认证流程、已登录页面、你的日常浏览器 profile、cookies、扩展或已有标签页。

所以选择很简单：

| 场景 | 用什么 |
| --- | --- |
| `localhost` 页面 | In-app Browser / Browser plugin |
| 本地静态 HTML 预览 | In-app Browser |
| 不需要登录的公开页面 | In-app Browser |
| 需要你的 Chrome 登录态 | Chrome Extension |
| 需要浏览器扩展或现有标签页 | Chrome Extension |
| 需要操作桌面软件 | Computer Use |

比如你接入了大模型 API 后，要验证聊天页面：

```text
请启动本地开发服务。
用浏览器打开 http://localhost:3000/chat。
测试：
1. 页面能正常加载
2. 模型选择器显示正确
3. 发送消息后有响应
4. loading 状态正常
5. 失败时错误提示可读
6. 375px 移动端不溢出
```

如果你的后台管理页需要登录态，例如已经登录在 Chrome 里，那就用 Chrome Extension。

但要注意权限。

Chrome Extension 会涉及真实网页、真实账号、真实 cookies。敏感后台、邮箱、CRM、支付系统，不要随手 Always allow。

Computer Use 则更进一步，它可以操作 Windows 或 macOS 图形界面。

适合：

- 测桌面 App
- 复现 GUI bug
- 点设置页面
- 操作没有 API 的本地软件
- 跨多个桌面应用走流程

但它也更容易影响你当前电脑状态。

让 Codex 操作桌面软件时，任务一定要窄：

```text
请只打开这个测试应用，复现设置页保存失败的问题。
不要修改系统设置，不要访问无关窗口。
每一步操作前先说明你要点哪里。
```

## 8. AGENTS.md：把长期规则写进项目

如果你每次都要提醒 Codex：

- 用 pnpm
- 不要引入新依赖
- 不要把 Key 写进代码
- 修改后跑测试
- 前端改动要用浏览器验证
- PR 描述要写风险

那就应该写进 `AGENTS.md`。

OpenAI 官方文档说明，Codex 在开始工作前会读取 `AGENTS.md`，并且支持全局和项目级规则。它会从全局目录、项目根目录一路读到当前工作目录，越靠近当前目录的规则越靠后，也就越具体。

对大模型 API 中转站项目，推荐这样写：

```md
# AGENTS.md

## Project Commands
- Install: pnpm install
- Dev: pnpm dev
- Test: pnpm test
- Lint: pnpm lint
- Build: pnpm build

## LLM Integration Rules
- Do not hardcode API keys.
- Read model base URL, API key, and model name from environment variables.
- Keep OpenAI-compatible request format where possible.
- Do not log full prompts, user private data, or API keys.
- Keep provider changes behind the existing LLM abstraction.
- Do not introduce new production dependencies without asking.
- Update `.env.example` when adding environment variables.
- Update README when changing setup steps.
- After changing LLM calls, run relevant tests.

## Review Expectations
- Keep changes scoped.
- Explain fallback behavior.
- Mention cost and privacy risks in PR summaries.
- For frontend chat UI changes, verify with browser at desktop and mobile widths.
```

如果你的项目有多个服务，可以在子目录再放更具体的规则。

比如：

```text
services/api/AGENTS.md
apps/web/AGENTS.md
packages/llm/AGENTS.md
```

这样后端、前端、LLM 封装层可以有不同规则。

不要把所有东西塞进根目录一个超长文件。

官方文档也提到 `AGENTS.md` 有合并大小限制，默认有上限。规则太长时，应该拆到更靠近任务的目录。

## 9. Skills、Plugins、MCP、Hooks：什么时候才需要上？

很多新手会被这些名词吓到。

其实可以按范围理解。

### Skill：重复任务流程

OpenAI Codex Skills 文档说明，Skill 可以打包任务指令、资源和可选脚本，让 Codex 更可靠地执行某类工作流。

适合做成 Skill 的任务：

- API 接入检查
- PR 复盘
- 代码审查
- 安全扫描前置检查
- 文章生成
- Excel 处理
- PPT 生成

比如可以做一个“大模型 API 接入检查 Skill”：

```text
输入：
- 项目路径
- 中转站 base_url
- 模型名
- 报错信息

流程：
1. 查找所有模型调用
2. 检查环境变量
3. 检查请求格式
4. 检查响应解析
5. 检查 Key 泄露
6. 运行最小测试
7. 输出修复建议
```

### Plugin：打包一组能力

Plugin 比 Skill 更大。

它可以包含 Skills、MCP、hooks、工具配置、模板、资产等。

适合团队分发。

比如团队内部可以做一个“LLM 接入插件”，里面包含：

- API 接入 Skill
- 成本审计 Skill
- 日志脱敏 Skill
- 4SAPI 示例模板
- pre-commit 检查脚本
- MCP 文档检索配置

### MCP：连接外部工具和上下文

OpenAI Codex MCP 文档说明，MCP 用来把模型连接到工具和上下文，例如第三方文档、浏览器、Figma 等。

对 API 接入很有用的 MCP 场景：

- 查内部 API 文档
- 查设计稿
- 查日志平台
- 查数据库 schema
- 查 issue tracker
- 查 OpenAI / Anthropic 最新文档

如果你经常复制网页资料给 Codex，就可以考虑 MCP。

### Hooks：自动执行规则

Hooks 更像工程护栏。

OpenAI Hooks 文档里提到，Codex 支持 `PreToolUse`、`PostToolUse`、`PermissionRequest`、`UserPromptSubmit`、`Stop` 等事件。

对大模型 API 接入，Hooks 可以做：

- 阻止删除 `.env`
- 阻止把 API Key 写入文件
- 命令执行前检查危险操作
- 命令执行后检查失败输出
- 任务结束前提醒运行测试
- 检测到 LLM 代码改动时提醒审查日志脱敏

比如：

```text
PreToolUse：检测命令是否包含 Remove-Item .env 或 rm .env
PostToolUse：如果 npm test 失败，把失败摘要加入上下文
Stop：如果改了 packages/llm 但没跑测试，要求继续运行测试
```

不过 Hooks 不是新手第一步。

先把 `AGENTS.md`、Review、Terminal 用顺，再上 Hooks。

## 10. 自动化：让 Codex 定期检查 API 接入健康度

Codex App 有 Automations。

如果你的项目长期使用大模型 API，可以让 Codex 定期做检查。

比如每周一次：

```text
每周一上午检查这个项目的大模型 API 接入状态。
请查看最近一周提交，重点关注：
1. LLM 调用相关文件是否有变化
2. 环境变量是否新增但文档没更新
3. 是否可能泄露 Key
4. 是否新增模型但没有测试
5. 是否有未处理的错误日志
6. 是否需要补充成本统计

输出一份 Markdown 报告。
```

再比如每天一次：

```text
每天上午 9 点检查最近 24 小时的 API 错误日志。
如果有 401、429、5xx 或模型不存在错误，请归类并提出修复建议。
没有新问题则自动归档。
```

自动化适合监控和提醒。

但不要让它直接修改生产配置。

尤其是 API Key、计费、权限、模型路由、生产数据库这些内容，都应该人工确认。

## 11. 一个真实项目的推荐工作流

假设你接手一个 Next.js 项目，要把原来的 OpenAI 直连改成 4SAPI 中转。

推荐流程如下。

### 第 1 步：建立项目地图

```text
先不要改代码。
请阅读项目，找出：
1. 聊天 API 路由
2. LLM 调用封装
3. 环境变量读取逻辑
4. 前端模型选择器
5. 测试命令和构建命令
6. 接入中转站的最小改动路径
```

### 第 2 步：制定最小方案

```text
请给出接入 4SAPI 的最小实现方案。
要求：
1. 保留现有 UI
2. 保留现有接口返回格式
3. 只改 LLM provider 层和配置
4. 不引入新生产依赖
5. 提供回滚方案
```

### 第 3 步：实现第一轮

```text
请按方案实现。
新增环境变量时同步更新 .env.example。
修改后运行相关测试。
如果测试失败，先汇报根因再修复。
```

### 第 4 步：终端验证

```text
请运行：
1. pnpm lint
2. pnpm test
3. pnpm build

把失败项按是否由本次改动引起分类。
```

### 第 5 步：浏览器验证

```text
请启动开发服务器，用浏览器打开聊天页面。
验证桌面端和 375px 移动端：
1. 页面不报错
2. 输入框可用
3. 发送消息后 loading 正常
4. 返回内容能显示
5. 错误提示不暴露 Key 或内部堆栈
```

### 第 6 步：Review 收口

```text
请 review 当前 diff。
重点检查：
1. 是否泄露 API Key
2. 是否修改无关文件
3. 是否遗漏文档
4. 是否破坏旧模型配置
5. 是否有缺失测试
先列问题，不要直接改。
```

### 第 7 步：提交

```text
请根据当前 diff 生成 commit message 和 PR 描述。
PR 描述需要包含：
1. 改了什么
2. 如何配置
3. 如何验证
4. 风险和回滚方式
```

这套流程看起来比“帮我接入一下”麻烦。

但在真实项目里，反而更快。

因为你少了大量返工。

## 12. 大模型 API 中转站接入的 12 个检查项

可以直接复制给 Codex：

```text
请按下面清单检查当前大模型 API 中转站接入：

1. base_url 是否来自环境变量
2. API key 是否只来自环境变量或安全配置
3. .env.example 是否更新
4. README 是否说明配置方式
5. 请求格式是否兼容中转站
6. 模型名是否可配置
7. 超时是否合理
8. 失败是否有可读错误提示
9. 日志是否避免记录 Key 和完整隐私内容
10. 是否有最小测试覆盖成功和失败分支
11. 前端是否处理 loading、error、empty 状态
12. 是否有回滚到原 provider 或备用模型的方案
```

这张清单比一句“检查一下有没有问题”有效得多。

## 13. 成本治理：Codex 能帮你查哪些浪费？

接入中转站后，很多团队的下一步是省钱。

不要只看模型单价。

让 Codex 帮你查代码里的浪费点：

```text
请检查项目里的大模型调用成本风险。
重点看：
1. 是否每次请求都重复发送超长系统提示词
2. 是否把完整历史消息无限传入
3. 是否没有限制 max_tokens
4. 是否失败后无限重试
5. 是否所有任务都使用最贵模型
6. 是否没有区分草稿、总结、审稿、代码任务
7. 是否能通过缓存、摘要或模型路由降低成本
```

常见优化方式：

- 固定系统提示词做缓存
- 长对话做摘要
- 低风险任务走便宜模型
- 高价值任务走 Claude / GPT 强模型
- 失败重试限制次数
- 对批量任务做队列
- 记录每个业务场景 token 用量

中转站在这里的作用是统一入口和统计口径。

比如 4sapi.com 这类服务，可以作为模型调用入口的一层，让你后续更容易做多模型路由、成本对比和调用审计。

## 14. 安全和合规：不要让 Codex 无脑放权

Codex 可以读文件、改文件、跑命令、开浏览器、操作 Git。

这很强。

也意味着你要有边界。

做大模型 API 接入时，至少遵守这些规则：

- 不要把真实 Key 写进仓库
- 不要让测试脚本默认打生产接口
- 不要把用户完整 prompt 写入日志
- 不要在错误提示里暴露内部栈和 Key
- 不要让 AI 自动修改生产模型路由
- 不要自动删除配置文件
- 不要自动上传敏感日志到外部网站
- 不要用中转站做违规用途

在 Codex 权限上，建议：

| 阶段 | 权限建议 |
| --- | --- |
| 读项目 | 只读或低权限 |
| 正常开发 | 工作区写入 |
| 跑测试 | 允许常规命令 |
| 删除文件 | 人工确认 |
| 写外部系统 | 人工确认 |
| 生产配置 | 人工确认 |

中转站本身也要注意合法合规。

可以讨论架构设计、格式转换、成本优化、负载均衡、计费管理。

不要把它用于恶意绕过官方限制、攻击、滥发、隐私泄露或其他违规用途。

## 15. 新手最容易踩的 10 个坑

第一，把 Codex 当聊天框。

结果只拿到一段代码，没完成验证。

第二，上来就让它改。

没有项目地图，改动容易跑偏。

第三，不跑测试。

看起来能编译，实际请求一发就炸。

第四，不看 Review。

真实 Key、无关文件、依赖变化都可能混进去。

第五，Worktree 里缺 `.env`。

代码没错，但运行环境缺配置。

第六，所有模型都走同一个强模型。

成本很快失控。

第七，错误提示直接暴露内部信息。

用户看到的是堆栈、Key 片段、供应商报错。

第八，没有回滚方案。

中转站或某个模型出问题时，业务没有备用路径。

第九，把 AGENTS.md 写成散文。

规则应该短、明确、可执行。

第十，滥用 Browser / Chrome / Computer Use 权限。

能用本地浏览器预览就不要碰真实登录态后台；能只读就不要写入。

## 16. 可以直接复制的提示词模板

### 了解项目

```text
先不要改代码。
请阅读项目并总结：
1. 项目用途
2. 技术栈
3. 启动、测试、构建命令
4. 大模型 API 调用位置
5. 环境变量配置方式
6. 接入大模型 API 中转站的最小路径
```

### 接入 4SAPI / 中转站

```text
我要把当前项目的大模型调用改成通过 4SAPI 这类大模型 API 中转站接入。
请先给出最小改动方案：
1. 哪些文件需要改
2. 哪些环境变量需要新增
3. 如何保持原有接口返回格式
4. 如何测试成功和失败分支
5. 如何回滚

确认方案后再改代码。
```

### 排查错误

```text
请根据这个报错排查大模型 API 调用失败原因。
请按类别判断：
1. API Key
2. base_url
3. 模型名
4. 请求格式
5. 流式输出
6. 网络或限流
7. 响应解析

先给出最可能的 3 个原因和验证命令。
```

### Review 改动

```text
请 review 当前未提交修改。
重点检查：
1. Key 泄露
2. 日志隐私
3. 请求格式兼容性
4. 模型配置可回滚
5. 缺失测试
6. 文档是否更新

先列问题，不要直接修改。
```

### 前端验证

```text
请启动开发服务器，用浏览器打开聊天页面。
验证桌面端和移动端：
1. 页面加载
2. 模型选择
3. 发送消息
4. loading 状态
5. 错误提示
6. 不暴露敏感信息
```

### PR 描述

```text
请根据当前 diff 写 PR 描述，包含：
1. 背景
2. 改动点
3. 配置方式
4. 验证结果
5. 风险
6. 回滚方式
```

## 17. 最后总结

Codex 桌面版最强的地方，不是“写代码很快”。

而是它把代码、终端、浏览器、Review、Git、Skills、Plugins、MCP、Automations 放进了同一个协作空间。

对大模型 API 中转站接入来说，它最适合做这条链路：

```text
读项目 -> 出方案 -> 最小改动 -> 运行测试 -> 浏览器验证 -> Review 收口 -> Git 交付
```

如果你只是让它写一段 API 调用代码，它只是代码生成器。

如果你让它按工程流程完成验证和交付，它才是开发搭档。

一句话总结：

```text
Codex 桌面版适合做本地交付，大模型 API 中转站适合做统一接入。
```

两者结合起来，最适合这些人：

- 想快速接入 Claude / GPT / Gemini 的开发者
- 想统一多模型调用的团队
- 想降低 API 调试成本的独立开发者
- 想把 AI 能力嵌入自己产品的企业

接入 4sapi.com 这类中转服务时，建议先从测试项目开始，小步验证，再逐步放到生产环境。

真正省时间的不是“一句话让 AI 全改完”。

而是让 Codex 按工程流程，把每一步都跑实。

## 资料来源与延伸阅读

- [Codex - OpenAI Developers](https://developers.openai.com/codex)
- [Features - Codex app](https://developers.openai.com/codex/app/features)
- [Review - Codex app](https://developers.openai.com/codex/app/review)
- [Worktrees - Codex app](https://developers.openai.com/codex/app/worktrees)
- [In-app browser - Codex app](https://developers.openai.com/codex/app/browser)
- [Codex Chrome extension - Codex app](https://developers.openai.com/codex/app/chrome-extension)
- [Computer Use - Codex app](https://developers.openai.com/codex/app/computer-use)
- [Custom instructions with AGENTS.md - Codex](https://developers.openai.com/codex/guides/agents-md)
- [Agent Skills - Codex](https://developers.openai.com/codex/skills)
- [Model Context Protocol - Codex](https://developers.openai.com/codex/mcp)
- [Hooks - Codex](https://developers.openai.com/codex/hooks)
- [Automations - Codex app](https://developers.openai.com/codex/app/automations)
- [OpenAI Community Forum](https://community.openai.com/)
