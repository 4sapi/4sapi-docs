---
title: "用 Codex Rescue 调查失败测试并安全验收修改"
category: 人工智能
tags:
  - Codex Rescue
  - 后台任务
  - 调试
description: "Rescue 可能运行调查并修改工作区。本文说明任务边界、background、resume、fresh、状态读取、diff 检查和测试验收的完整流程。"
---

# 用 Codex Rescue 调查失败测试并安全验收修改

失败测试跨平台出现、触发条件不稳定或涉及多个模块时，只读 Review 往往只能指出风险，不能完成复现和最小修复。`/codex:rescue` 会把任务交给本机 Codex，并且可能运行命令、读取日志和修改文件，因此任务描述和验收边界必须先写清。本文只解决一次安全的 Rescue 调查：记录工作区状态，限定允许修改范围，选择 background、resume 或 fresh，读取结果后检查 diff，并用原始失败测试和相关验证证明修改。会话迁移和模型路由不在本文范围内。

常见调查问题包括：

```text
为什么 Windows CI 才失败？
哪次提交引入了回归？
flaky test 的触发条件是什么？
能否用最小补丁修复？
```

这些任务比普通 Review 更重，也可能需要读日志、运行命令和修改文件。

[`codex-plugin-cc` 官方仓库](https://github.com/openai/codex-plugin-cc) 的 `/codex:rescue` 支持把这类任务交给本机 Codex。

它通过 `codex:codex-rescue` subagent 委派任务，并支持后台执行和继续或新建调查线程。

## 1. Rescue 和 Review 的区别

| 能力 | Review | Rescue |
| --- | --- | --- |
| 找问题 | 是 | 是 |
| 运行调查 | 有限、只读审查 | 是 |
| 尝试修复 | 否 | 是 |
| 继续历史任务 | 否 | 是 |
| 自定义任务描述 | 普通 Review 不支持 | 支持 |
| 修改工作区 | 否 | 可能 |

所以 Rescue 不能像 Review 一样随便调用。

任务开始前要先检查工作区，结束后必须审查 diff。

## 2. 最小 Rescue 任务

调查失败测试：

```text
/codex:rescue investigate why the tests started failing
```

尝试最小修复：

```text
/codex:rescue fix the failing test with the smallest safe patch
```

后台调查回归：

```text
/codex:rescue --background investigate the regression
```

官方 README 建议长任务优先后台运行，避免占住当前 Claude Code 会话。

## 3. 高质量 Rescue 指令怎么写

不要只说：

```text
/codex:rescue fix tests
```

至少给出：

```text
症状。
环境。
可复现命令。
允许修改范围。
禁止修改范围。
修复偏好。
验收标准。
```

示例：

```text
/codex:rescue --background investigate why checkout tests fail only on Windows. Reproduce with [仓库实际复现命令]. Do not modify unrelated modules or update dependencies. Prefer the smallest patch, add a regression test, and report any behavior you could not verify.
```

这才是把调查交给 Codex，而不是把所有工程判断一起交出去。

## 4. `--background` 和 `--wait` 怎么选

Rescue 支持：

```text
--background
--wait
--resume
--fresh
--model
--effort
```

使用建议：

| 场景 | 方式 |
| --- | --- |
| 很小、需要立即结果 | `--wait` 或前台执行 |
| 多文件调查、CI 回归 | `--background` |
| 继续上一轮调查 | `--resume` |
| 不想继承旧结论 | `--fresh` |

不要在同一 Rescue 调用中同时追求“快速初查”和“完整重构”。

先调查，后决定修复范围。

## 5. `--resume`：继续上一条线索

例如第一次调查得到三个可能原因。

继续验证最可能的方案：

```text
/codex:rescue --resume apply the top fix from the last run
```

如果省略 `--resume` 和 `--fresh`，插件可以提示是否继续当前仓库最近的 Rescue 线程。

使用 Resume 前确认：

```text
仓库仍是同一个。
分支和工作树没有发生重大变化。
旧结论仍然适用。
任务结果没有过期。
```

代码已经大幅变化时，宁可 `--fresh`。

## 6. `--fresh`：切断错误上下文

旧任务可能形成错误锚点。

适合重新开始：

```text
底层依赖已经升级。
失败症状发生变化。
切换了分支。
旧调查明显误判。
希望独立复核。
```

命令：

```text
/codex:rescue --fresh investigate the regression from current HEAD only
```

## 7. 后台任务怎么管理

查看当前仓库任务：

```text
/codex:status
```

查看指定任务：

```text
/codex:status task-abc123
```

读取最终输出：

```text
/codex:result
/codex:result task-abc123
```

取消任务：

```text
/codex:cancel
/codex:cancel task-abc123
```

取消前先确认任务 ID，避免停止另一条仍有价值的调查。

## 8. Result 不是结束，必须检查工作区

Rescue 可能修改代码。

任务结束后执行：

```bash
git status --short
git diff --stat
git diff
```

然后跑任务对应的测试：

```text
原始失败测试。
相关模块测试。
类型检查或 lint。
必要的构建和烟雾测试。
```

不要只看 Codex 说“已修复”。

完成声明必须由新测试结果证明。

## 9. 如何保护用户现有改动

启动 Rescue 前：

```text
记录 git status。
告诉 Codex 不要覆盖无关未提交改动。
限定允许修改的目录。
必要时使用独立分支或 worktree。
```

任务提示词加入：

```text
The worktree may contain user changes. Preserve unrelated edits. Do not reset, clean, checkout, or rewrite files outside the stated scope.
```

不要把 `git reset --hard`、`git clean -fd` 之类命令交给自动调查流程。

## 10. 常见错误

### Rescue 任务太模糊

“fix tests”缺少症状、范围和验收。

### 继续了错误历史

仓库变化很大时应该使用 `--fresh`。

### 同时开多个修改型 Rescue

并行写同一工作树容易冲突。

### 只看 Result 不看 diff

模型报告不能替代代码审查。

## 11. 检查清单

```text
[ ] Rescue 任务包含症状、环境、范围和验收
[ ] 已记录启动前 git status
[ ] 无关用户改动明确要求保留
[ ] 长任务使用 background
[ ] resume 前确认旧上下文仍有效
[ ] 上下文过期时使用 fresh
[ ] 修改型任务没有并行写同一工作树
[ ] result 后检查 git diff
[ ] 已重跑原始失败测试和相关验证
[ ] 任务 ID、修改文件和测试结果已经记录
```

## 12. 结论与限制

`/codex:rescue` 不是更长的 Review。

它是一个可能执行调查和修改的委派入口。

正确流程是：

```text
写清问题和边界。
后台运行。
用 status 跟踪。
用 result 获取结论。
检查 diff。
重跑测试。
必要时 resume 或 fresh。
```

Rescue 的结果只是调查输出，不能替代工作区审查和新测试证据。启动前记录 `git status`，结束后查看完整 diff，并确认没有覆盖无关用户改动；成立的修复应补回归测试。

本文命令按核对时的官方插件 README 编写，不覆盖 Transfer、模型选择或会话导入。插件行为和参数可能更新，执行前仍需复查当前 README；多个修改型 Rescue 不应并行写同一工作树。
