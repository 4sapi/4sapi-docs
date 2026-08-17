---
title: "用 Codex 插件分开做缺陷审查与设计挑战"
category: 人工智能
tags:
  - codex-plugin-cc
  - Claude Code
  - Code Review
description: "普通 Review 适合定位具体缺陷，对抗性 Review 适合挑战设计假设。本文说明命令选择、基线、焦点写法、证据核对和修复闭环。"
---

# 用 Codex 插件分开做缺陷审查与设计挑战

如果连续让两个模型“帮我审一下”，结果往往重复，真正的设计风险仍没有被单独检查。codex-plugin-cc 提供两种只读命令：普通 `/codex:review` 检查当前改动中的具体问题，`/codex:adversarial-review` 接受焦点文本，用于挑战假设、权衡与失败路径。本文只解决两类审查如何分工：先确认基线并完成目标测试，再运行普通审查；仅对高风险区域追加设计挑战；所有发现都回到代码位置、触发条件和可复现证据，最后由开发者补回归测试并决定是否修复。

分工可以写成：

```text
Claude Code：理解需求、实现和整合修改。
Codex Review：独立检查具体缺陷。
Codex Adversarial Review：挑战设计和失败路径。
```

这些命令来自 [`openai/codex-plugin-cc` 官方仓库](https://github.com/openai/codex-plugin-cc)。插件与命令可能更新，执行前应复查当前 README。

## 1. 两种审查的本质区别

| 类型 | 命令 | 主要问题 | 可否追加焦点文本 |
| --- | --- | --- | --- |
| 普通审查 | `/codex:review` | 代码中有哪些真实问题 | 否 |
| 对抗审查 | `/codex:adversarial-review` | 方案为什么可能失败 | 是 |

官方 README 明确说明，`/codex:review` 不可 steer，不接受自定义关注文本。

如果你要让 Codex 专门检查缓存、竞态、数据丢失或权限边界，应该使用 `/codex:adversarial-review`。

## 2. 普通 Review 审什么

普通 Review 适合：

```text
当前未提交改动。
当前分支相对于基线分支的改动。
提交前或 PR 前的独立检查。
```

最小命令：

```text
/codex:review
```

多文件改动建议后台运行：

```text
/codex:review --background
```

审查当前分支相对于 `main`：

```text
/codex:review --base main --background
```

如果仓库默认分支是 `master`、`develop` 或其他名称，要换成真实基线。

## 3. Review 是只读的

官方插件文档明确写明：

```text
/codex:review 不会修改代码。
/codex:adversarial-review 也不会修复代码。
```

这很重要。

审查阶段应该先保留证据和独立性。让同一个任务一边找问题一边修改，容易出现：

```text
发现被修复过程覆盖。
无关重构混入 diff。
问题描述和最终代码不对应。
```

审查完成后，由 Claude Code、人工开发者或单独的 `/codex:rescue` 处理修复。

## 4. 普通 Review 的推荐时机

### 完成功能后

```text
实现完成。
目标测试通过。
准备提交前。
```

执行：

```text
/codex:review --background
```

### 创建 PR 前

```text
/codex:review --base main --background
```

### 合并前

先更新基线并确认 diff，再做最后一次分支审查。

不要每改一行就触发完整审查。

## 5. 对抗性审查解决什么

普通 Review 更关注：

```text
Bug。
边界条件。
行为回归。
错误处理。
```

对抗性审查更关注：

```text
设计假设是否站得住。
方案是否过度复杂。
失败时会不会扩大影响。
有没有更安全或更简单的替代方案。
```

适合：

```text
认证和授权。
支付与退款。
数据删除。
缓存和重试。
并发和事务。
数据库迁移。
模型路由与自动回退。
```

## 6. 对抗性审查怎么写

普通调用：

```text
/codex:adversarial-review
```

针对分支和具体设计：

```text
/codex:adversarial-review --base main challenge whether this caching and retry design is safe
```

后台检查竞态与数据丢失：

```text
/codex:adversarial-review --background look for race conditions and data-loss risks
```

高质量焦点文本包含四部分：

```text
担心的风险。
必须保护的数据或行为。
允许的失败方式。
希望它挑战的假设。
```

例如：

```text
/codex:adversarial-review --base main challenge the authorization design. Focus on tenant isolation, stale permission caches, rollback behavior, and whether a simpler design would reduce bypass risk
```

## 7. 不要把焦点写成结论

不好的写法：

```text
prove this design is unsafe
```

这会把审查变成寻找支持既定结论的证据。

更好的写法：

```text
challenge whether this design is safe, identify the assumptions that must hold, and compare one simpler alternative
```

目标是压力测试，不是要求模型唱反调。

## 8. 一套完整双层审查流程

第一步，Claude Code 实现功能并运行测试。

第二步，Codex 普通审查：

```text
/codex:review --base main --background
```

第三步，继续运行仓库实际规定的测试与静态检查命令，不假设所有项目都使用同一工具链。

第四步，读取审查结果：

```text
/codex:status
/codex:result
```

第五步，只对高风险区域追加对抗审查：

```text
/codex:adversarial-review --base main focus on auth bypass, data loss, retry amplification, and rollback gaps
```

第六步，把有效发现转成修复任务和回归测试。

## 9. 如何判断一条发现是否有效

每条发现至少要回答：

```text
代码位置。
触发条件。
实际影响。
为什么当前测试没有挡住。
最小修复或验证方法。
```

只有“这里可能有问题”不够。

可以让 Claude Code 对 Codex 结果做证据核对：

```text
请逐条验证 Codex 审查发现。
先定位代码和触发路径，再判断是否成立。
不要因为发现来自另一个模型就直接修改。
对成立的问题补回归测试；不成立的问题说明证据。
```

## 10. 两个模型意见冲突怎么办

不要让它们继续互相辩论。

回到可执行证据：

```text
测试能否复现。
类型和静态检查是否报错。
数据库约束是否成立。
权限路径是否可构造。
官方 API 契约怎么写。
运行日志是否支持判断。
```

最终决策人仍然是代码所有者和审核人。

## 11. 后台任务管理

查看全部近期任务：

```text
/codex:status
```

读取最新结果：

```text
/codex:result
```

指定任务：

```text
/codex:status task-abc123
/codex:result task-abc123
```

取消不再需要的审查：

```text
/codex:cancel task-abc123
```

并行审查不要太多，否则同一仓库会出现重复计算、结果过期和额度快速消耗。

## 12. 常见错误

### 给 `/codex:review` 追加焦点文本

普通 Review 不可 steer。需要指定焦点时使用对抗性审查。

### 两种 Review 连续无差别运行

只有高风险变更才值得追加对抗性审查。

### 不跑测试，只看 AI 结论

Review 是补充证据，不是替代验证。

### 发现问题后直接让模型大改

先确认触发路径，再选择最小修复。

## 13. 检查清单

```text
[ ] 已确认正确的 base 分支
[ ] 功能实现和目标测试先完成
[ ] 普通 Review 没有附加无效焦点文本
[ ] 多文件审查使用 background
[ ] 高风险设计才追加 adversarial review
[ ] 焦点描述包含风险、资产和失败路径
[ ] Review 结果逐条核验证据
[ ] 成立的问题补回归测试
[ ] 两个模型冲突时回到运行证据
[ ] 审查目标、任务结果和采纳决定可追踪
```

## 14. 结论与限制

`codex-plugin-cc` 的双层审查不是让 Codex 重复 Claude Code 的工作。

正确分工是：

```text
普通 Review 找具体缺陷。
对抗性 Review 挑战方案和失败路径。
Claude Code 整合修复。
测试与人工审核决定是否合并。
```

双层审查的价值不在于意见数量，而在于问题类型被明确分开，并且每条发现都能由代码和测试证据复核。普通 Review 不接受焦点文本；需要挑战特定设计时才使用 Adversarial Review。

两种命令都不会替代仓库测试、人工审核和代码所有者判断。base 分支必须换成仓库真实基线，后台任务结果也可能在代码继续变化后过期；修复前应重新确认目标 diff 和触发路径。
