---
title: "【大模型API中转站】第107期 AGENTS与Skills | 把Codex变成你的工作流"
category: 人工智能
tags:
  - 大模型API中转站
  - Codex
  - AGENTS.md
  - Skills
  - 工作流
  - 4SAPI
description: "讲清 Codex 里提示词、AGENTS.md 和 Skills 的分工：一次性要求写提示词，项目长期规则写 AGENTS.md，重复流程做成 Skill，并说明企业如何用 4SAPI 承接多模型与成本治理。"
---

# 【大模型API中转站】第107期 AGENTS与Skills | 把Codex变成你的工作流

本文是【大模型API中转站】系列的第107篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

很多人用 Codex 一段时间后，会遇到一个很烦的问题：

```text
每次都要重新说一遍项目规则。
每次都要提醒不要改锁文件。
每次都要提醒先跑测试。
每次都要提醒前端要看 390px 移动端。
每次都要提醒不要顺手重构。
```

这不是你不会写 prompt。

这是规则放错了地方。

Codex 里有三层特别关键：

```text
Prompt：一次性任务要求。
AGENTS.md：项目长期规则。
Skill：可复用任务流程。
```

这三层分清楚，Codex 才会从“每次重新教一遍”变成“按你的工作方式干活”。

这篇专门讲：

```text
什么该写在提示词里。
什么该写进 AGENTS.md。
什么该做成 Skill。
以及企业怎么把这些流程接到 4SAPI 的模型治理里。
```

## 1. 不要把所有东西都塞进 prompt

很多人会写超长提示词。

比如：

```text
你是一个资深工程师。
请遵守本项目规范。
不要新增依赖。
不要改锁文件。
前端要检查移动端。
测试要跑 pnpm test。
提交前要总结风险。
另外我们项目用 pnpm monorepo...
```

第一次有用。

第二次还要复制。

第三次就开始漏。

真正稳定的做法是分层。

```text
本次任务目标：写在 prompt。
项目固定规则：写在 AGENTS.md。
重复执行步骤：做成 Skill。
强制检查：用 Hook 或 CI。
模型调用治理：交给 4SAPI。
```

别让 prompt 背所有锅。

Prompt 适合表达这一次要做什么。

不适合承载项目制度。

## 2. AGENTS.md 是项目里的长期规则

`AGENTS.md` 可以理解成写给 Codex 的项目说明书。

它适合放：

```text
项目结构
技术栈
常用命令
测试方式
禁止修改的目录
代码风格
交付要求
安全边界
前端验证尺寸
发布前检查
```

一个实用版本可以这样写：

```markdown
# 项目说明

- 这是一个 pnpm monorepo。
- Web 应用在 `apps/web`。
- 共享组件在 `packages/ui`。

# 工作规则

- 修改前先阅读相关目录的 README 和现有测试。
- 不要修改生成文件和锁文件，除非任务确实需要。
- 不要新增依赖，除非先说明原因。
- 保持改动聚焦，不做无关重构。

# 验证

- 单元测试：`pnpm test`
- 类型检查：`pnpm typecheck`
- 构建：`pnpm build`
- 前端改动需在 1440px 与 390px 下手动验证。

# 交付

- 总结修改文件、执行命令、结果和剩余风险。
```

这比你每次在 prompt 里重复一大段规则稳得多。

## 3. AGENTS.md 不要写成公司百科

`AGENTS.md` 不是越长越好。

很多团队会犯一个错误：

```text
把十几页公司规范都塞进去。
```

结果 Codex 每次读到一堆不相关信息。

上下文变脏。

规则反而更容易被忽略。

更好的做法是：

```text
根目录放全局规则。
子目录放局部规则。
越靠近任务目录，规则越具体。
```

比如：

```text
AGENTS.md
apps/web/AGENTS.md
packages/api/AGENTS.md
docs/AGENTS.md
```

根目录写通用规则。

前端目录写浏览器验证和组件规范。

后端目录写测试数据库、接口约束和迁移禁区。

文档目录写标题、格式和引用要求。

这才符合工程直觉。

## 4. AGENTS.md 里最值得写的 8 类内容

第一，项目定位。

```text
这个项目做什么。
主要用户是谁。
哪些模块最关键。
```

第二，目录边界。

```text
哪些目录是业务代码。
哪些是生成文件。
哪些不能自动改。
```

第三，常用命令。

```text
安装、启动、测试、lint、typecheck、build。
```

第四，验证要求。

```text
什么任务跑什么测试。
前端怎么检查。
后端怎么复现。
```

第五，依赖规则。

```text
默认不新增依赖。
确实需要时先解释原因。
```

第六，安全边界。

```text
不要读取私钥。
不要改生产配置。
不要执行真实付款、删除、发布。
```

第七，交付格式。

```text
最后要列文件、命令、结果、风险。
```

第八，领域禁区。

```text
比如不能修改计费规则。
不能绕过权限。
不能跳过审计。
```

写这些，比写一堆“请认真、请专业、请不要犯错”有用得多。

## 5. 什么情况该做成 Skill

`AGENTS.md` 管项目规则。

Skill 管可复用流程。

如果你发现自己经常说同一套步骤，就该做成 Skill。

比如：

```text
修 bug：复现 -> 定位 -> 最小修复 -> 补测试 -> 复测 -> 总结风险
前端验收：启动服务 -> 桌面截图 -> 移动端截图 -> 控制台检查 -> 溢出检查
发布检查：工作树 -> 版本号 -> changelog -> 测试 -> 构建 -> 迁移风险
内容写作：读规范 -> 查资料 -> 写稿 -> QA -> 来源清单
```

这些都不是一次性任务。

它们是流程。

流程就适合 Skill。

## 6. 一个最小 Skill 长什么样

Skill 的最小结构很简单。

```text
focused-bugfix/
  SKILL.md
  scripts/
  references/
  assets/
```

只有 `SKILL.md` 是必需的。

例如：

```markdown
---
name: focused-bugfix
description: 当用户要求定位并修复可复现 bug，且希望最小改动和测试证据时使用。
---

1. 先复现或建立最小复现。
2. 用日志、失败测试或调用链定位根因。
3. 只修改解决根因所需的文件。
4. 补一个修复前失败、修复后通过的测试。
5. 跑相关测试、类型检查和构建。
6. 报告证据与剩余风险。
```

这不是“万能提示词”。

它是一个可复用工作流。

更复杂时，可以加：

```text
scripts：放检查脚本、生成器、转换工具。
references：放规则文档、模板、示例。
assets：放图片、模板文件、样例素材。
```

Codex 会按需要渐进加载。

平时只看名称和描述。

真正匹配任务时才读完整 Skill。

## 7. Skill 的描述比你想象中重要

很多人创建 Skill，只写：

```text
description: 修 bug 用。
```

这太粗。

Skill 的 description 应该说清楚：

```text
什么时候使用。
什么时候不使用。
适合什么输入。
最终产出是什么。
```

例如：

```text
当用户要求定位并修复可复现 bug，且希望最小改动、测试证据和风险说明时使用。
不用于大范围重构、性能优化或没有明确复现路径的探索任务。
```

这能减少误触发。

工具多了以后，选择本身就是成本。

好的 Skill 描述，能帮 Codex 选对流程。

## 8. AGENTS.md 和 Skill 怎么配合

可以这样理解：

```text
AGENTS.md：这个项目里永远遵守什么。
Skill：这类任务每次怎么做。
Prompt：这一次具体要完成什么。
```

例如你要修一个支付 bug。

Prompt 写：

```text
目标：修复用户重复点击支付按钮会创建两笔订单的问题。
范围：只改支付按钮、订单创建入口和现有测试。
验收：相关测试通过，并证明重复点击只创建一笔订单。
```

AGENTS.md 提供：

```text
支付配置不可修改。
生产密钥不可读取。
数据库迁移不可自动执行。
支付相关改动必须补测试。
```

Skill 提供：

```text
复现 -> 定位 -> 最小修复 -> 补测试 -> 复测 -> 总结风险。
```

这三层配起来，任务就稳很多。

## 9. 内容团队也可以用这套方法

别以为 `AGENTS.md` 和 Skills 只适合写代码。

内容团队同样需要。

比如你的博客系列，就可以这样分层。

项目规则：

```text
标题必须含“大模型API中转站”或模型关键词。
必须有系列导语。
必须自然植入企业级大模型接入、企业级 API、日志审计、成本治理。
不能鼓励绕过官方限制。
必须有参考资料。
```

写作 Skill：

```text
读取发文规范。
查看最近 3 篇同系列文章。
确定选题是否撞题。
搜索资料并记录来源。
生成 Markdown 初稿。
检查标题、front matter、code fence、4SAPI 露出和来源清单。
```

Prompt：

```text
这次写 3 篇 Codex 系列文章，主题围绕 App、AGENTS.md 和 Plugins。
```

这就不是“让 AI 随便写几篇”。

这是一个可复用内容生产工作流。

## 10. 企业落地：不要只靠文字规则

`AGENTS.md` 和 Skill 很有用。

但它们不是万能安全系统。

它们更像：

```text
工作说明书。
```

真正企业上线，还需要：

```text
权限隔离
沙箱
CI
测试
代码审查
Hook
日志审计
成本控制
Key 分组
```

也就是说，规则能写进 AGENTS.md，就写。

能用 Skill 固化流程，就固化。

但能用程序检查的，不要只写在提示词里。

比如：

```text
不能改 .env
不能删除测试
不能改支付配置
diff 行数超限要报警
未跑测试不能标记完成
```

这些最好交给 Hook、CI 或脚本检查。

## 11. 4SAPI 在工作流里的位置

当团队开始把 Codex、Skills、自动化结合起来，模型调用就会变成真实成本。

这时要考虑：

```text
不同 Skill 用什么模型？
修 bug 用什么模型？
Review 用什么模型？
内容初稿用什么模型？
最终审查用什么模型？
哪个团队跑得最多？
哪条工作流最贵？
```

建议用 4SAPI 做统一模型入口。

```text
Base URL: https://4sapi.com/v1
Key：按项目、团队、环境、工作流拆分
Model：按 Skill 阶段路由
Logs：按 Key 查看调用记录
Budget：按工作流设预算
Audit：记录调用来源和失败原因
```

例如：

```text
bugfix-planner-key：用于问题定位和计划。
bugfix-executor-key：用于实现和修复。
bugfix-reviewer-key：用于最终审查。
content-draft-key：用于内容初稿。
content-qa-key：用于内容质量检查。
```

这样做的好处是：

```text
成本可控。
日志清楚。
权限隔离。
出问题能停单条流程。
```

这才是企业级 API 治理。

## 12. 一个最小落地路线

如果你现在还没用过 `AGENTS.md` 和 Skill，可以按这个顺序来。

第一步，写根目录 `AGENTS.md`。

只写最重要的：

```text
项目结构
常用命令
禁止项
验证要求
交付格式
```

第二步，跑 3 个任务，看哪些规则反复提醒。

第三步，把重复流程做成第一个 Skill。

推荐从 `focused-bugfix` 开始。

第四步，再做一个 `review-changes` Skill。

让它专门按严重度审查 bug、回归、安全和测试缺口。

第五步，内容或文档团队再做 `article-draft-qa` Skill。

第六步，进入自动化和多模型后，把模型调用收口到 4SAPI。

不要第一天就做 20 个 Skill。

先把一个流程跑稳。

## 13. 常见坑

第一，把 `AGENTS.md` 写太长。

结果上下文变脏。

第二，把一次性需求写进 `AGENTS.md`。

过几天就变成过期规则。

第三，把 Skill 写成口号。

只有“认真检查、保证质量”，没有具体步骤。

第四，所有任务都触发同一个 Skill。

说明 description 太宽。

第五，只靠文字规则防风险。

该用测试、Hook、CI、权限隔离的地方，不能只靠模型自觉。

第六，不做成本治理。

Skill 多了以后，模型调用会变多。

进入团队场景后，要尽早用 4SAPI 拆 Key、看日志、控预算。

## 14. 最后总结

真正会用 Codex 的人，不是每次写一段更长的 prompt。

而是把规则放到正确的位置。

```text
Prompt：本次任务。
AGENTS.md：项目长期规则。
Skill：可复用流程。
Hook / CI：强制检查。
4SAPI：模型入口、Key、日志、成本和权限治理。
```

如果你每次都在重复提醒 Codex 同一件事，就该停下来想想：

```text
这是一次性要求，还是长期规则？
这是项目规则，还是流程？
这是文字约束，还是应该用测试和 Hook 检查？
```

把这几个问题答好，Codex 才会从“会聊天的工具”变成“按你方法工作的 Agent”。

一句话：

```text
AGENTS.md 让 Codex 懂项目，Skills 让 Codex 懂流程，4SAPI 让模型调用可治理。
```

## 资料来源与延伸阅读

- OpenAI Codex AGENTS.md：https://developers.openai.com/codex/guides/agents-md
- OpenAI Codex Skills：https://developers.openai.com/codex/skills
- OpenAI Codex Customization：https://developers.openai.com/codex/concepts/customization
- OpenAI Codex Hooks：https://developers.openai.com/codex/hooks
- OpenAI Codex Config Reference：https://developers.openai.com/codex/config
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 文档：https://4sapi.apifox.cn/
