---
title: "用单一事实来源生成多种 Agent 项目规则"
category: "人工智能"
tags:
  - AI Agent
  - 配置管理
  - 软件工程
description: "将项目结构、命令和限制维护为统一上下文，再按需生成不同 Coding Agent 的规则文件。"
brand_mode: none
---

# 用单一事实来源生成多种 Agent 项目规则

团队同时使用多个 Coding Agent 后，最先失控的往往不是模型选择，而是项目规则文件。

一份规则写了测试命令，另一份忘了更新；一个工具被允许改生成目录，另一个工具仍被禁止。

规则越多，成员越难判断哪一份才是当前事实。

下面提供一种单一事实来源的目录设计：先维护工具无关的项目上下文，再生成各 Harness 所需的薄适配文件。

读者将得到一套 `.agent/` 目录、一个可审查的生成清单和一组更新后的验证方法。

示例中的文件名和脚本都是团队约定，生成目标应按实际工具支持的规则文件调整。

## 适用场景

该方法适合一个仓库需要被多个执行入口访问的团队。

这些入口可能是终端 Agent、IDE 插件、远程 Runner 或自动化任务。

它也适合项目中已有多份提示词、操作手册和命令清单，但内容开始分叉的情况。

如果团队只使用一个入口，且规则短小稳定，手写一个规则文件可能更简单。

当规则包含安全边界、目录职责和验证要求时，集中维护通常更容易审查。

## 为什么规则会漂移

项目规则经常分散在 README、开发者文档、CI 配置和各工具的指令文件里。

它们可能都曾经正确，却在一次迁移后失去同步。

最危险的不是文字不一致。

而是一个执行者依据过期命令运行，另一个执行者依据新命令修改代码，结果无法复现。

解决方式不是要求所有人“记得同步”。

应当明确一份可审查的事实来源，并让派生文件可以重新生成。

## 目录设计

下面的目录把稳定项目知识和任务运行信息分开。

```text
.agent/
  project.md
  architecture.md
  conventions.md
  commands.yaml
  policies.yaml
  harnesses.yaml
  generated/
  tasks/
  runs/
scripts/
  render-agent-rules.mjs
```

`project.md` 说明项目目的、模块边界和本地启动方式。

`architecture.md` 记录不应由临时任务推翻的架构约束。

`conventions.md` 放代码风格、测试组织和命名规则。

`commands.yaml` 为构建、测试、格式化命令提供机器可读定义。

`policies.yaml` 描述文件与 Shell 权限，不能由文本规则替代。

`generated/` 只存派生物，不作为人工编辑的源头。

## 先区分稳定知识与临时信息

项目知识应该具有较长生命周期。

例如“支付模块不得依赖 Web 控制器”属于架构事实。

某次超时修复的临时方案不属于项目事实。

任务进展、工具日志和补丁路径也不属于项目事实。

它们应分别进入任务文件和运行目录。

这条边界能减少“把一次故障经验永久写进规则”的问题。

## 用结构化文件存放命令

命令是最容易被规则文本写错的内容之一。

将它们集中到 YAML 后，可以被生成器和预检程序同时读取。

```yaml
# .agent/commands.yaml
version: 1
commands:
  test_payment:
    cwd: services/payment
    command: npm test -- payment
    purpose: 支付模块单元测试
  lint_payment:
    cwd: services/payment
    command: npm run lint
    purpose: 支付模块静态检查
  format_check:
    cwd: .
    command: npm run format:check
    purpose: 全仓格式检查
```

命令键名应表达用途，而不是绑定某个 Harness。

`cwd` 必须显式指定，避免执行者在错误目录调用同一条命令。

对于需要变量的命令，使用受控参数模板，不要直接拼接不可信输入。

## 把规则写成短而可验证的陈述

规则文件不是项目百科全书。

一条好的规则应当说明范围、动作和检查方法。

以下是 `conventions.md` 的片段。

```markdown
## 测试

- 为行为变化新增或更新测试。
- 支付模块测试放在 `services/payment` 对应的测试目录。
- 修改完成后运行 `test_payment` 和 `lint_payment`。

## 变更边界

- 不直接编辑自动生成目录。
- 公开 API 的变更必须在任务验收中标记为人工审查。
- 新增依赖需要单独的人工确认。
```

规则不应包含“尽量”“通常”这类无法判断的词，除非它确实是建议而非要求。

真正需要阻止的动作还要落入策略文件。

## 用策略文件表达强制边界

项目文本告诉执行者应该做什么。

策略文件告诉控制层能够允许什么。

```yaml
# .agent/policies.yaml
filesystem:
  read:
    - services/payment/**
    - tests/payment/**
    - .agent/**
  write:
    - services/payment/**
    - tests/payment/**
    - .agent/runs/**
  deny:
    - .env
    - secrets/**
    - migrations/**

shell:
  allow:
    - npm test -- payment
    - npm run lint
  confirm:
    - npm install
  deny:
    - git push
```

路径模式的匹配语义必须由控制层固定。

例如是否允许 `..`、符号链接和 Windows 路径分隔符，不能留给不同适配器各自处理。

## 定义生成输入

生成器需要知道哪些来源参与渲染，以及每个目标要得到哪些内容。

```yaml
# .agent/harnesses.yaml
version: 1
sources:
  - project.md
  - architecture.md
  - conventions.md
  - commands.yaml
targets:
  codex:
    output: generated/AGENTS.md
    include: [project, architecture, conventions, commands]
  claude:
    output: generated/CLAUDE.md
    include: [project, architecture, conventions, commands]
  generic:
    output: generated/project-context.md
    include: [project, architecture, conventions, commands]
```

目标文件名只是示例。

若某个工具不读取仓库内规则文件，适配器可以在启动时注入 `project-context.md` 的内容。

无论注入还是落盘，都应由同一份源文件生成。

## 生成器应该做什么

一个最小生成器只需要四步。

第一步读取清单和每个源文件。

第二步把结构化命令渲染为人类可读的命令表。

第三步为每个目标组装相同的事实内容。

第四步在派生文件顶部写明来源和再生成命令。

```text
load harness manifest
validate listed source files
render commands.yaml to Markdown
for each target:
  concatenate requested context sections
  write generated target with provenance header
```

不要在生成器中塞入各工具的业务规则。

工具特定的格式转换可以放在模板层，但内容判断应留在源文件中。

## 一个派生文件的样子

```markdown
<!-- Generated from .agent/ sources. Do not edit directly. -->

# Project rules

## Required checks

- `services/payment`: `npm test -- payment`
- `services/payment`: `npm run lint`

## Protected paths

- Do not edit `.env`, `secrets/`, or `migrations/`.
```

派生文件应保持短小。

它负责被特定 Harness 消费，不负责成为团队唯一的文档入口。

开发者需要查看详细原因时，应回到 `.agent/` 的源文件。

## 在 CI 中检查生成物

生成规则后，最重要的是避免手工修改派生文件。

可以在 CI 中重新生成，然后检查工作区是否出现差异。

```bash
node scripts/render-agent-rules.mjs
git diff --exit-code -- .agent/generated
```

第一条命令按当前源文件重建派生规则。

第二条命令确认提交中的派生物没有过期。

在 PowerShell 环境中，也可以用等价的 Git 差异检查。

这条检查不验证 Harness 是否真的读取了规则。

它只验证仓库中的源文件与派生文件一致。

## 如何验证规则被正确注入

选择一条无风险且可见的规则，例如“禁止编辑 `migrations/`”。

先运行生成命令，确认三个目标都出现该规则。

再由每个适配器打印启动前的上下文摘要。

摘要应包含源版本、生成文件路径和策略版本。

不要要求 Agent 口头复述规则作为验证。

应当验证适配器读取了哪个具体文件，以及控制层是否执行了策略检查。

## 常见失败与排查

### 失败一：派生文件被人工编辑

这是最常见的漂移来源。

在文件头写明“不要直接编辑”只能提醒，不能阻止修改。

CI 差异检查和代码审查规则才能让问题尽早暴露。

若确有工具特有补充，先判断它是否应成为通用事实。

不能通用的补充应放在目标模板，并加注释说明原因。

### 失败二：把权限列表复制到提示词

提示词可以描述边界，但它不是可靠的执行拦截。

规则文件只用于解释意图。

可读、可写和可执行范围仍应由外部策略层控制。

### 失败三：源文件不断膨胀

当 `.agent/project.md` 包含每个历史故障和所有模块细节时，任何 Harness 都会得到过长上下文。

保留稳定且与多数任务相关的信息。

模块专属资料可以按目录拆分，在任务中显式引用。

### 失败四：生成器吞掉了来源错误

如果源文件不存在或 YAML 无法解析，生成器必须失败。

不要以空内容继续生成一个看似正常的规则文件。

运行日志应指出错误文件与行号，便于修正。

## 替代方案

小团队可以不生成多个文件。

只维护 `.agent/project-context.md`，由所有适配器在启动时注入它，也是一种有效的单一事实来源。

当某个 Harness 只能读取固定文件名时，再增加生成步骤。

另一种做法是将规则写入内部文档站。

它适合知识量很大的组织，但应保留可在仓库版本控制和离线读取的最小副本。

## 结论

多 Harness 并不要求维护多份项目知识。

先把架构、命令、规范和策略放进可审查的统一来源，再将必要内容生成或注入到各个执行入口。

这样工具可以变化，项目事实仍有一个明确的维护位置。
