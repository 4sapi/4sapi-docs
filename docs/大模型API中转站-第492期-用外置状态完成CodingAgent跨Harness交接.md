---
title: "用外置状态完成 Coding Agent 跨 Harness 交接"
category: "人工智能"
tags:
  - AI Agent
  - 工程协作
  - 可观测性
description: "将任务进度、补丁、验证证据和待办事项写入运行目录，避免切换 Coding Agent 时重复工作或覆盖修改。"
brand_mode: none
---

# 用外置状态完成 Coding Agent 跨 Harness 交接

长任务在一个 Coding Agent 会话中执行到一半时，最难迁移的往往不是代码，而是“已经确认了什么、还差什么”。

如果这些信息只存在于聊天记录里，下一个 Harness 只能重新阅读仓库、重复排查，甚至覆盖尚未提交的修改。

下面给出一种外置运行状态的做法：任务文件描述意图，运行目录保存进度、证据、补丁和待办项。

读者将得到统一结果对象、事件日志、交接检查清单和一次跨 Harness 恢复的参考流程。

示例文件展示的是控制层协议，不承诺不同工具可以自动迁移内部会话或隐藏上下文。

## 适用场景

该方法适合需要跨天执行、需要人工审批或可能切换执行入口的代码任务。

它也适合让一个 Agent 产出方案，另一个 Agent 实现，第三个入口执行验证的团队。

一次性、小范围且无需审计的改动，可以只保留 Git 提交和测试输出。

如果工作区本身不受版本控制，先解决快照和回滚问题，再讨论跨 Harness 交接。

## 为什么聊天记录不是可靠状态

聊天记录会混合指令、推理、尝试、错误输出和临时结论。

其中有价值的信息往往没有固定格式。

会话可能被压缩、截断或无法被另一套工具读取。

即使能导出全文，也很难确定哪些结论已被代码和测试验证。

外置状态并不要求保存全部对话。

它要求保存能够继续执行和审计的最小事实集合。

## 运行目录结构

每次运行都使用新的 `run_id`。

任务 ID 和运行 ID 不应混用。

```text
.agent/runs/
  2026-08-27T101500Z-fix-payment-timeout/
    manifest.json
    state.yaml
    events.jsonl
    change.diff
    result.json
    summary.md
    checks/
      payment-unit.log
      payment-lint.log
```

`manifest.json` 保存任务版本、基准提交和适配器身份。

`state.yaml` 保存当前阶段、已完成项和待办项。

`events.jsonl` 追加记录可观察事件。

`change.diff` 是变更快照，不替代 Git 提交。

`checks/` 保存每项自动检查的原始输出。

## 在开始时固定基线

交接前最重要的问题是：当前修改基于哪一个工作区状态？

运行开始时，控制层应记录基准提交、分支名和工作树是否干净。

```json
{
  "schema_version": 1,
  "run_id": "2026-08-27T101500Z-fix-payment-timeout",
  "task_id": "fix-payment-timeout",
  "stage": "implement",
  "base_commit": "a18c9f2",
  "branch": "fix/payment-timeout",
  "worktree_clean_at_start": true,
  "harness": "codex",
  "adapter_version": "1.0.0"
}
```

提交哈希只是示例，不代表真实仓库状态。

如果启动时工作区已有未提交修改，记录这些修改的摘要或单独建立隔离工作树。

不要让新运行默认接管来源不明的工作树。

## 用状态文件写已知事实

状态文件要短，并且只记录已确认的工程事实。

不要把未验证的猜测写成完成项。

```yaml
# state.yaml
status: testing
completed:
  - 定位到回调客户端未设置读取超时
  - 在受限工作区修改了超时配置
  - 新增读取超时的单元测试
pending:
  - 运行支付模块测试
  - 检查公开 API 是否未变化
blocked: []
next_action: 运行 command_id 为 test_payment 的检查
```

`completed` 中的每一项应能关联到代码差异、测试日志或审查记录。

`pending` 需要足够具体，使新执行者不必重新猜测下一步。

`blocked` 只列真实阻塞，例如缺少权限、测试基础设施不可用或等待人工确认。

## 事件日志记录过程，不替代状态

事件日志适合记录时序信息。

JSON Lines 格式允许追加写入，也便于后续检索。

```json
{"time":"2026-08-27T10:16:04Z","type":"stage_started","stage":"implement"}
{"time":"2026-08-27T10:18:51Z","type":"file_changed","path":"services/payment/client.ts"}
{"time":"2026-08-27T10:22:13Z","type":"check_started","command_id":"test_payment"}
{"time":"2026-08-27T10:24:01Z","type":"check_finished","command_id":"test_payment","exit_code":1}
```

事件应该描述发生了什么，不应存储模型推理文本或敏感参数。

日志中涉及路径、命令或错误输出时，也要遵守同一份权限和脱敏策略。

状态文件给出当前事实，事件日志解释这些事实如何出现。

## 统一最终结果

不论实际由哪个 Harness 执行，最终结果都应使用同一结构。

```json
{
  "schema_version": 1,
  "status": "completed_with_open_items",
  "summary": "为支付回调添加读取超时测试和受限变更。",
  "files_changed": [
    "services/payment/client.ts",
    "tests/payment/client.test.ts"
  ],
  "checks": {
    "payment-unit": "passed",
    "payment-lint": "passed",
    "api-review": "pending_manual_review"
  },
  "risks": [
    "部署配置的兼容性仍需人工审查"
  ],
  "handoff": {
    "next_stage": "review",
    "next_action": "审查公开 API 兼容性"
  }
}
```

`completed_with_open_items` 比只写 `completed` 更准确。

状态枚举应预先定义，避免每个适配器自造含义相近的词。

结果对象应区分通过、失败、未运行和需要人工确认。

## 状态不是对补丁的替代

文字说明无法保证下一个执行者看到与上一个相同的代码。

因此交接时至少要保存补丁和基准信息。

在 Git 仓库中，可以导出当前差异。

```bash
git diff --binary HEAD > .agent/runs/2026-08-27T101500Z-fix-payment-timeout/change.diff
git status --short > .agent/runs/2026-08-27T101500Z-fix-payment-timeout/worktree-status.txt
```

第一条命令生成当前工作树相对于 `HEAD` 的二进制安全差异。

第二条命令保存未跟踪和已修改文件的简明状态。

在真实流程中，应确认运行目录本身不会被导出到错误的补丁范围。

## 执行交接前检查

新 Harness 接手前，不应直接读取状态文件后继续写代码。

控制层应按固定顺序验证交接材料。

```text
1. 读取 manifest，确认任务 ID、基准提交和运行阶段。
2. 比较当前工作区与记录的基线，确认差异来源。
3. 读取 state，确认完成项都有对应证据。
4. 读取 result 与 checks，识别未通过和未运行的检查。
5. 运行能力与权限预检。
6. 仅执行 pending 与 next_action 中的工作。
```

若当前基准提交不一致，不应静默应用旧补丁。

可以创建新的运行，要求人工解决冲突，或在隔离分支上恢复。

## 阶段化切换示例

下面展示一个参考流程。

```bash
# 方案阶段只读执行
agent run fix-payment-timeout --harness claude --stage plan

# 实现阶段基于同一任务和最新运行状态恢复
agent resume fix-payment-timeout --run 2026-08-27T101500Z-fix-payment-timeout --harness codex --stage implement

# 验证阶段只允许登记的检查命令
agent resume fix-payment-timeout --run 2026-08-27T101500Z-fix-payment-timeout --harness deepseek --stage verify
```

这些是控制层的示例命令。

`resume` 的关键不是“把聊天窗口交给另一个工具”。

它是读取经过校验的任务、基线、补丁、状态和验证证据。

每次切换都应写入新的事件，说明执行者和阶段发生了变化。

## 如何验证交接可用

在测试仓库运行一个只改测试文件的任务。

故意在测试执行失败后停止第一段运行。

确认状态中列出失败检查、日志路径和下一步动作。

再用另一个已配置适配器恢复。

恢复后的执行者应先报告它检测到的基准、差异和待办项。

它不应重新修改已经完成的文件，也不应将失败检查标成通过。

最后由人工核对最终结果中的文件列表与 Git 差异是否一致。

## 常见失败与排查

### 失败一：状态写得过于笼统

“已修复问题”不能帮助下一个执行者判断要做什么。

改为写出变更位置、验证状态和明确下一动作。

状态并非越长越好，但必须可以指导一次安全的恢复。

### 失败二：测试日志和结果不一致

结果写着通过，日志却显示非零退出码，说明适配器或汇总器存在问题。

以原始检查日志为证据来源。

结果对象只能总结，不应覆盖原始退出码。

### 失败三：恢复时忽略基准提交

仓库已发生重基或其他人提交修改时，旧补丁可能不再适用。

先比较基准，再决定重新生成补丁、解决冲突或新建运行。

不要为了“自动恢复”而强行覆盖工作树。

### 失败四：把完整对话当作必需工件

完整对话可能包含冗余内容，也可能不能安全共享。

优先保存任务事实、补丁、检查证据和待办项。

只有在隐私与权限允许时，才将经过脱敏的对话摘要作为辅助材料。

## 适用边界

外置状态能降低交接成本，不能让不同 Harness 的工具行为完全一致。

它也不能替代分支保护、代码审查、测试基础设施或备份策略。

涉及数据库迁移、发布和外部系统写入时，应增加人工审批和更细粒度的回滚方案。

## 结论

可切换的 Coding Agent 工作流，不应把关键进度锁在聊天窗口里。

将任务状态、工作区基线、代码差异、检查日志和待办项外置后，交接的对象就从一段会话变成了可验证的工程工件。

先在一个可回滚的维护任务上演练恢复流程，再把它推广到复杂的阶段化协作。
