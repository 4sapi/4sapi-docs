---
title: "把 AI 编码任务写成可切换的任务协议"
category: "人工智能"
tags:
  - AI Agent
  - 软件工程
  - YAML
description: "用结构化任务协议承载目标、约束、验收和权限，让不同 Coding Agent 接收同一份可验证输入。"
brand_mode: none
---

# 把 AI 编码任务写成可切换的任务协议

同一项代码维护任务，换了 Coding Agent 就要重新解释一遍，通常不是模型“忘了”，而是需求只存在于某个聊天窗口里。

自然语言需求适合发起讨论，却很难成为可重复执行的工程输入。

例如“修复支付超时，不要破坏现有接口”缺少工作目录、禁止项、测试命令和完成条件。

执行者换成另一套 Harness 后，只能从零猜测这些信息。

下面给出一个最小任务协议，把任务拆成目标、约束、验收和权限四部分。

读者将得到一个可放入仓库的 YAML 任务文件、一个校验命令和一套切换执行者时的检查方法。

文中的 `agent` 是团队自建控制层的示例命令，不是任一 Coding Agent 的原生命令。

## 问题出现在哪里

许多团队把“任务”当作聊天的第一句话。

聊天记录包含了任务背景、临时补充、工具输出和已经失败的尝试。

它对当前会话有用，对下一个执行者却不稳定。

其中一部分信息本该是需求，另一部分只是推理过程。

若把二者混在一起，迁移时要么遗漏约束，要么塞入大量无关历史。

任务协议的目的不是消除自然语言。

它是把必须被执行和验收的事实固定下来。

讨论、方案草稿和模型回答仍可以保留在运行记录里。

但它们不应成为任务能否继续的唯一依据。

## 适用场景

该方法适合有代码仓库、测试命令和明确变更边界的任务。

修复缺陷、实现小功能、补充测试和受限重构都可以采用。

它也适合在同一团队内使用多个 CLI、IDE 插件或 Agent Runtime 的情况。

如果任务本身还没有可验证的目标，先完成需求澄清。

协议无法替代产品决策，也无法把模糊需求自动变得正确。

## 前置条件

准备一个 Git 仓库，并确保团队能在本地运行至少一条测试命令。

准备 YAML 解析器或一个可以读取 YAML 的脚本环境。

为控制层约定一个目录，例如仓库根目录的 `.agent/tasks/`。

下面的例子假定任务名称为 `fix-payment-timeout`。

不要把密钥、访问令牌或真实用户数据写进任务文件。

任务文件应进入版本控制，敏感配置应通过运行环境单独提供。

## 先定义最小字段

一个能切换 Harness 的任务至少需要标识、目标、工作区、约束和验收条件。

标识用于关联运行记录，不应依赖某个会话 ID。

目标描述期望的工程结果，而不是指定模型如何思考。

工作区确定代码边界，避免执行者从仓库根目录随意扩展搜索范围。

约束记录明确不能做的事情。

验收把“完成”转化为可以检查的证据。

权限说明本次任务允许的副作用范围。

以下是完整的最小示例。

```yaml
# .agent/tasks/fix-payment-timeout.yaml
schema_version: 1
id: fix-payment-timeout
goal: 修复支付回调偶发超时
workspace: services/payment

constraints:
  - 不修改公开 API
  - 不新增生产依赖
  - 保留现有数据库结构

acceptance:
  - command: npm test -- payment
    expect: exit_code_0
  - command: npm run lint -- services/payment
    expect: exit_code_0
  - evidence: 新增读取超时场景的测试
  - evidence: 输出变更摘要

permissions:
  filesystem: write
  shell: restricted
  network: deny
```

这里的字段名是团队协议，不必与下游 Harness 的参数同名。

关键是字段的含义保持稳定，并由控制层负责映射。

## 为什么不要只写一段提示词

一段长提示词很难被程序可靠检查。

例如“不要改接口”无法直接判断具体改动是否违反约束。

结构化字段可以在运行前检查是否缺失。

执行结束后，也可以将每条验收项关联到命令输出或补丁证据。

这不意味着 YAML 天然更安全。

YAML 只是承载格式，真正的保证来自后续的预检、权限执行和验收。

## 给任务加入输入与输出

当任务依赖工单、设计文档或测试数据时，应写明它们的位置与读取方式。

不要让执行者通过模糊搜索猜测“相关资料”是什么。

同样，要求交付的文档也应写成输出，而不是藏在对话尾部。

```yaml
inputs:
  - path: docs/payment/callback-contract.md
    purpose: 回调兼容性约束
  - path: tests/payment/fixtures/timeout.json
    purpose: 超时测试夹具

outputs:
  - path: runs/{run_id}/summary.md
    purpose: 变更摘要与未完成项
  - path: runs/{run_id}/result.json
    purpose: 统一结果
```

输入路径应相对仓库根目录解析。

控制层应在运行前确认文件存在且位于允许读取的范围内。

输出路径最好放在运行目录，避免不同任务互相覆盖。

## 把验收拆成可观察项

“测试全部通过”常常不够具体。

它没有说明测试集合、运行位置和失败时的处理方式。

将验收项拆成命令、期望结果和人工检查三类会更清晰。

```yaml
acceptance:
  checks:
    - id: payment-unit
      type: command
      command: npm test -- payment
      cwd: services/payment
      expected_exit_code: 0
    - id: payment-lint
      type: command
      command: npm run lint
      cwd: services/payment
      expected_exit_code: 0
    - id: timeout-test
      type: evidence
      requirement: 测试名称覆盖读取超时或等价情形
    - id: api-review
      type: manual
      requirement: 审查公开 API 未变化
```

命令检查由控制层执行或记录执行证据。

证据检查需要定义可接受的材料，例如文件路径、测试名称或差异片段。

人工检查必须保留为人工项，不应伪装成自动通过。

## 约束不是权限策略的替代品

约束描述本次工作的产品与技术边界。

权限描述进程实际上可以执行哪些操作。

“不修改数据库结构”是约束。

“不能写入 `migrations/` 目录”是权限规则。

前者需要审查和测试，后者应在工具层拦截。

不要把两者混成一个布尔字段。

分开后，团队可以复用权限策略，同时为每个任务配置不同业务限制。

## 为协议增加阶段

复杂任务可以分为计划、实现和验证阶段。

阶段不是为了让流程看起来完整，而是为了限制每一步的产物和权限。

计划阶段通常只读。

实现阶段可以在受控工作区写入。

验证阶段可执行白名单测试命令，但不应自动发布或推送代码。

```yaml
stages:
  - id: plan
    permissions:
      filesystem: read
      shell: restricted
  - id: implement
    permissions:
      filesystem: write
      shell: restricted
  - id: verify
    permissions:
      filesystem: read
      shell: restricted
```

不要用阶段名推断权限。

每个阶段应显式声明它继承或覆盖哪些权限。

## 编写一个运行前校验

先校验任务文件，再调用任一 Harness。

以下伪代码展示控制层应检查的最低内容。

```text
load task
require schema_version == 1
require id, goal, workspace
require workspace exists and is inside repository
require acceptance contains at least one check
reject input or output path outside repository
reject permissions.network when policy forbids network
print normalized task as JSON
```

规范化后的 JSON 是适配器接收的内部输入。

适配器不应各自重新解释原始 YAML。

这样可避免某个适配器把缺失字段默默当成默认值。

## 运行示例

下面的命令展示控制层的使用方式。

```bash
agent validate fix-payment-timeout
agent run fix-payment-timeout --harness codex --stage implement
```

第一条命令只读取并校验任务文件。

第二条命令将已校验的任务交给名为 `codex` 的适配器。

若团队更换执行者，任务定义不改变。

```bash
agent run fix-payment-timeout --harness claude --stage implement
agent run fix-payment-timeout --harness deepseek --stage implement
```

这些命令能否运行，取决于团队是否实现了对应适配器。

它们不表示三个工具具有相同能力，也不表示能够无损共享会话。

## 如何验证协议有效

先故意删掉 `acceptance` 字段。

`agent validate` 应拒绝执行，并指出缺失的字段。

再把 `workspace` 改成仓库外路径。

校验器应在启动 Harness 前失败。

最后使用同一份合法任务，分别调用两个已配置的适配器。

两次运行的规范化任务摘要应一致。

允许它们的提示词格式、工具调用和输出措辞不同。

不允许某个适配器悄悄丢失约束或验收项。

## 常见失败与排查

### 失败一：任务文件变成第二份项目文档

如果每个任务都重复粘贴架构、代码规范和构建说明，维护成本会快速上升。

任务只保留本次差异化目标。

稳定的项目知识应进入统一项目上下文，并由项目规则生成流程维护。

### 失败二：验收项无法执行

写入不存在的测试命令，不能增加任务的可信度。

先在目标仓库人工运行命令，再把它写入协议。

若某项只能人工判断，应标为 `manual`，并在结果中保留待确认状态。

### 失败三：把 Harness 名称写进任务语义

例如“请使用某工具的子代理完成重构”会让任务天然绑定某个运行时。

改为描述目的，例如“将相互独立的检查并行执行”。

由适配器根据能力决定并行、串行或拒绝运行。

### 失败四：用任务 ID 覆盖旧运行记录

任务 ID 表示需求，不表示某一次执行。

每次执行还需要独立的 `run_id`。

任务可以重跑，运行记录不应被覆盖。

## 什么时候不需要这套协议

临时问答、一次性代码解释和没有落地改动的头脑风暴，不需要创建正式任务文件。

单人维护的小仓库也可以先只写目标和测试命令。

当出现第二个执行入口、第二位协作者或需要复盘的失败任务时，再补齐约束、权限和运行记录会更合适。

## 结论

切换 Coding Agent 的前提不是让不同 Harness 使用同一种提示词。

前提是把目标、边界、验收与权限从聊天记录中拿出来，变成可校验的任务资产。

先用一个可执行任务验证协议是否完整，再扩展到更多 Harness 和更复杂的自动路由。
