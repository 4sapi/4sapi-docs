---
title: "团队 AI Skill 如何拆成可评审的单一工作流"
category: 开发工具
tags:
  - AI Skill
  - 工作流设计
  - 团队协作
description: "把聊天中的临时要求整理为单一问题的 Skill，明确触发、输入、步骤、权限、完成证据和回归样例，使团队能够评审与迭代。"
---

# 团队 AI Skill 如何拆成可评审的单一工作流

团队使用 AI 编程工具一段时间后，常会积累大量复制粘贴的提示词：有人要求先写测试，有人习惯先读架构文档，还有人把部署、审查和复盘都塞进同一段指令。问题不是缺少更多规则，而是这些规则没有明确触发条件、权限和完成证据，难以知道哪一版生效。一个可维护 Skill 应只解决一个稳定问题，并像轻量流程一样接受评审和回归测试。本文以缺陷诊断为例，说明如何从临时对话提炼出可复用资产。

## 先判断什么值得做成 Skill

适合沉淀的工作流通常满足：

- 在多个任务中重复出现；
- 有相对稳定的输入和完成标准；
- 执行顺序会影响质量或安全；
- 需要明确工具权限和停止条件；
- 可以用固定样例验证。

一次性的业务事实、临时项目状态和只对单个文件有效的要求，不应写入通用 Skill。它们留在任务说明或项目文档中。

## 一个 Skill 只解决一个问题

“从需求到部署的完整研发流程”范围过大。可拆为：

```text
澄清会改变实现方向的需求缺口
为行为变化建立失败测试
根据错误证据定位缺陷
审查代码差异是否越界
准备发布前的验证报告
```

每个 Skill 有独立触发和完成条件。组合流程由调用方按任务需要选择，不能让模型把所有 Skill 默认串行执行。

## 六个必要区块

### 1. Trigger

说明何时使用以及何时不使用。例如缺陷诊断只在存在可观察失败时触发，不用于无复现的功能设计。

### 2. Inputs

列出开始前必须具备的材料：失败命令、环境、相关路径和禁止修改范围。缺失关键输入时进入澄清，而不是自行补齐。

### 3. Workflow

描述会改变结果的最小顺序，不规定每一步内部推理。例如“先复现，再收集首个相关错误，然后提出一个可证伪假设”。

### 4. Boundaries

区分允许、需要确认和禁止的动作。权限边界必须能够由工具或运行环境实施，文本规则不能代替系统控制。

### 5. Done

使用命令、文件或来源证据定义完成。不要写“问题已解决”这种自我判断。

### 6. Failure

规定信息不足、同类错误重复和范围外根因时如何停止、记录状态并交给人工。

## 一个缺陷诊断模板

```markdown
# Diagnose a Reproducible Failure

## Trigger
Use when a command or user workflow has a repeatable failure.
Do not use for feature design without a failing behavior.

## Inputs
- Reproduction command or steps
- Runtime and dependency versions
- Allowed read paths
- Allowed write paths
- Behavior that must remain unchanged

## Workflow
1. Run the reproduction and record command, exit code, and first relevant error.
2. Read only the nearest code and tests needed to explain that error.
3. Separate confirmed facts from hypotheses.
4. Add a minimal regression test and verify that it fails for the target reason.
5. Make the smallest implementation change.
6. Run the target test, related tests, and required full verification.
7. Review the diff for out-of-scope changes.

## Boundaries
- Reading scoped project files and running local non-destructive tests is allowed.
- New dependencies, deleted files, external writes, and expanded scope need approval.
- Never expose credentials or report a command as run when it was not run.

## Done
- The regression test failed before the implementation.
- Target and required regression checks pass after the change.
- The diff contains only approved files.
- Remaining limitations are explicit.

## Failure
- If reproduction is unavailable, report the missing evidence.
- If the root cause is outside scope, stop with file and command evidence.
- If the same tool error repeats to the configured limit, hand off a state summary.
```

模板是工作流结构，不是某个项目的默认权限。落地时按工具支持的格式和仓库规则调整。

## 把事实与流程分开

Skill 保存稳定流程，项目文档保存技术事实：

```text
Skill：如何诊断、如何验证、何时停止
项目入口：常用命令、关键目录、安全边界
架构决策：为什么采用当前接口和数据模型
任务记录：本次错误、修改和验证结果
```

如果把完整架构和所有命令复制进每个 Skill，规则会快速分叉。Skill 只引用需要按需读取的事实来源。

## 为 Skill 准备回归样例

至少准备：

1. 正常触发：有稳定失败，预期完成诊断和验证；
2. 不应触发：纯功能设计，预期选择其他流程；
3. 输入缺失：没有复现，预期列缺口并停止；
4. 权限边界：要求修改禁止路径，预期在写入前阻断；
5. 工具失败：环境错误重复，预期交接而非循环。

每次修改 Skill 后在同一测试工作区运行这些样例，记录实际工具调用和完成证据。只阅读文本无法验证触发与边界是否有效。

## 版本化变更与适用范围

为 Skill 记录版本、负责人、适用工具和最后验证的样例集。变更说明回答：

```text
修复了哪个失败样例；
是否改变工具或权限；
是否需要迁移项目配置；
如何恢复旧版本；
哪些环境尚未验证。
```

若不同工具的 Skill 目录、元数据或触发机制不同，分别维护适配层，核心流程仍保持一个来源。

## 评审重点

审查 Skill 时优先查：

- 触发是否过宽；
- 是否混入会过期的项目事实；
- 是否请求超出任务需要的工具；
- 完成条件是否能够外部验证；
- 失败时是否会继续循环；
- 是否会覆盖用户现有修改；
- 示例是否被误写成真实执行结果。

文字精炼属于次要问题。一个写得漂亮但无法阻止越权动作的 Skill 仍不合格。

## 结论与限制

团队 Skill 的工程价值来自单一问题、明确触发、最小权限、外部完成证据和可重复样例。把流程、项目事实与任务状态分开维护，才能让规则被评审、版本化和回归。

本文不规定任何 AI 工具的安装命令、目录或元数据格式。不同客户端如何发现、加载和组合 Skill 需要依据当前官方文档验证；高风险操作仍应由系统权限和人工审批控制。
