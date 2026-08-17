---
title: "【大模型API中转站】第124期 OmO Team Mode | 成本别失控"
category: 人工智能
tags:
  - 大模型API中转站
  - oh-my-openagent
  - Team Mode
  - OpenCode
  - 多Agent
  - 成本治理
  - 企业级大模型接入
  - 4SAPI
description: "专门拆解 oh-my-openagent Team Mode 的企业级使用方式：什么时候开、什么时候不开，如何限制 max_parallel_members、max_wall_clock_minutes、成员角色和 worktree，如何用 4SAPI 单独拆 Team Mode Key、预算、日志和复盘，避免多 Agent 并行变成成本黑洞。"
---

# 【大模型API中转站】第124期 OmO Team Mode | 成本别失控

本文是【大模型API中转站】系列的第124篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

前面几篇讲了 OmO 接入 4SAPI 的总览、配置和企业治理。

这一篇单独讲一个容易让人兴奋，也容易让账单发烫的功能：

```text
Team Mode。
```

oh-my-openagent 的 Team Mode 可以让一个 Lead Agent 协调多个成员并行工作。

听起来很爽。

但企业里要记住一句话：

```text
多 Agent 并行不是免费加速器。
它是高成本模式。
```

如果你开得好，它适合大型重构、复杂排查、安全审计。

如果你乱开，它会变成：

```text
多个 Agent 同时读代码。
多个 Agent 同时调用模型。
多个 Agent 同时产生日志。
多个 Agent 同时等你来判断。
```

最后任务没快多少，成本先上去了。

所以这篇只讲一件事：

```text
Team Mode 怎么用 4SAPI 做企业级成本治理。
```

## 1. Team Mode 不是日常默认模式

OmO 文档里写得很清楚：

```text
Team Mode 默认关闭。
```

这是对的。

因为 Team Mode 适合的不是所有任务。

适合它的任务通常有三个特点：

```text
范围大。
可以拆。
拆完能汇总。
```

比如：

```text
跨模块重构。
大型 Bug 根因排查。
安全审计。
多仓库迁移。
技术方案评审。
性能瓶颈定位。
复杂依赖升级。
```

这些任务一个 Agent 顺序做会很慢。

拆成多个子任务，才有并行价值。

不适合 Team Mode 的任务：

```text
改一个 README。
修一个拼写。
补一个测试。
解释一段代码。
写一篇短文。
生成一个配置片段。
```

这些任务开 Team Mode，就是让一个小问题长出很多管理成本。

## 2. 企业里先给 Team Mode 定级

建议把 Agent 任务分成四级。

```text
L0：只读问答。
L1：单 Agent 小改。
L2：单 Agent 深度任务。
L3：Team Mode 并行任务。
```

对应权限：

```text
L0：只读，无需高成本模型。
L1：允许修改小范围文件。
L2：允许跨文件修改，必须跑测试。
L3：需要审批，单独 Key，单独预算。
```

Team Mode 应该放在 L3。

不要让所有人随手开。

尤其在企业仓库里，Team Mode 至少要满足：

```text
任务有明确目标。
任务能拆成独立子任务。
有停止条件。
有验证方式。
有预算上限。
有负责人看最终 diff。
```

没有这些，就先别开。

## 3. 最小开启配置

Team Mode 开启是在 OmO 配置里写：

```jsonc
{
  "team_mode": {
    "enabled": true,
    "max_parallel_members": 4,
    "max_members": 8,
    "tmux_visualization": false
  }
}
```

配置文件可以是：

```text
~/.config/opencode/oh-my-openagent.jsonc
.opencode/oh-my-openagent.jsonc
```

旧名：

```text
oh-my-opencode.jsonc
```

在过渡期也能识别。

改完后重启 OpenCode。

重启后会出现 12 个 `team_*` 工具。

比如：

```text
team_create
team_send_message
team_task_create
team_task_list
team_task_update
team_status
team_delete
```

如果没出现，先跑：

```bash
bunx oh-my-openagent doctor
```

不要直接怀疑模型。

先确认插件和 Team Mode 是否加载成功。

## 4. max_parallel_members 不要一上来拉满

文档里 `max_parallel_members` 范围是 1 到 8。

默认建议 4。

企业试点建议从 2 开始。

```jsonc
{
  "team_mode": {
    "enabled": true,
    "max_parallel_members": 2,
    "max_members": 4,
    "tmux_visualization": false,
    "max_wall_clock_minutes": 60,
    "max_member_turns": 120
  }
}
```

为什么？

因为并行度翻倍，消耗通常不只是翻倍。

它还会增加：

```text
协调消息。
重复阅读。
结果汇总。
冲突处理。
上下文管理。
人工判断。
```

所以最稳路线是：

```text
2 个并行成员试点。
跑 3 到 5 个真实任务。
看 4SAPI 日志和最终收益。
再考虑调到 4。
```

不要第一天就开 8。

## 5. Team Mode 单独拆 4SAPI Key

Team Mode 必须单独 Key。

不要和日常开发共用。

推荐：

```text
4sapi-omo-teammode-dev
4sapi-omo-teammode-review
4sapi-omo-teammode-security
```

这样做有四个好处。

第一，成本单独看。

```text
普通开发花了多少。
Team Mode 花了多少。
一眼分清。
```

第二，额度单独设。

```text
Team Mode 每周或每月一个预算上限。
```

第三，故障单独停。

```text
如果 Team Mode 跑飞，只停它自己的 Key。
```

第四，权限单独管。

```text
只有负责人能拿 Team Mode Key。
```

4SAPI 在这里不是简单转发。

它是企业 API 网关。

Team Mode 越强，越需要网关兜底。

## 6. 推荐的模型分配

Team Mode 不代表所有成员都用最强模型。

推荐分层。

```text
Lead：强推理模型。
Scout：低成本快速模型。
Implementer：代码模型。
Reviewer：强推理或不同厂商模型。
Writer：写作模型。
```

一个大型重构可以这样拆：

```text
Lead：Sisyphus，负责分解任务和汇总。
Scout-API：快速扫描接口层。
Scout-Test：快速扫描测试覆盖。
Implementer：修改核心模块。
Reviewer：检查 diff 和风险。
```

不是所有成员都需要强推理。

很多成员只是找文件、查调用链、整理证据。

这些任务用 fast 模型就够。

4SAPI 后台可以按模型看消耗。

如果你发现：

```text
所有 Team Mode 成员都在用高价模型做搜索。
```

说明路由策略错了。

## 7. team spec 怎么写

Team Mode 的 team spec 可以放在：

```text
~/.omo/teams/{name}/config.json
```

或项目里：

```text
<project>/.omo/teams/{name}/config.json
```

项目级同名配置优先。

一个最小例子：

```json
{
  "name": "api-refactor-audit",
  "description": "Audit API refactor blast radius before implementation.",
  "lead": {
    "kind": "subagent_type",
    "subagent_type": "sisyphus"
  },
  "members": [
    {
      "kind": "category",
      "name": "route-scout",
      "category": "quick",
      "prompt": "Scan route definitions and list affected endpoints. Do not edit files."
    },
    {
      "kind": "category",
      "name": "test-scout",
      "category": "quick",
      "prompt": "Scan tests related to the affected endpoints. Do not edit files."
    },
    {
      "kind": "category",
      "name": "deep-implementer",
      "category": "deep",
      "prompt": "Implement only after lead assigns a scoped task. Keep changes minimal."
    }
  ]
}
```

这里故意让 scout 只读。

这样不会出现三个成员同时改同一批文件。

先查。

再分配。

最后改。

## 8. 哪些 Agent 能进 Team Mode

OmO 文档里写了成员资格。

可用：

```text
sisyphus
atlas
sisyphus-junior
```

条件可用：

```text
hephaestus
```

但需要 teammate 权限。

不能直接作为 Team Mode 成员的有：

```text
oracle
librarian
explore
multimodal-looker
metis
momus
prometheus
```

原因是这些不适合作为写 mailbox 状态的团队成员。

企业写文章时不用展开底层原因。

只要告诉读者：

```text
Team Mode 成员不是所有 Agent 都能当。
不能当成员的，用 task / delegate-task 或普通 Agent 调用。
```

不要乱配。

乱配会解析失败。

## 9. 工作树怎么用

Team Mode 支持给成员配置 worktree。

比如：

```json
{
  "kind": "category",
  "name": "migration-scout",
  "category": "deep",
  "prompt": "Investigate migration impact.",
  "worktreePath": "../wt-migration-scout"
}
```

这适合：

```text
多成员并行修改。
需要隔离实验。
怕互相覆盖文件。
```

但企业刚开始不建议直接上复杂 worktree。

先让成员只读分工。

等团队熟悉后，再让不同成员在不同 worktree 修改。

否则你会遇到：

```text
分支太多。
diff 难合。
修改重复。
验证环境不一致。
```

Team Mode 的第一阶段目标不是“所有成员都改代码”。

而是：

```text
让成员并行收集证据，Lead 再决定怎么改。
```

## 10. 每次 Team Mode 都要有停止条件

不要让 Team Mode 无限跑。

建议任务提示词里写清楚：

```text
最多运行 60 分钟。
每个成员最多完成 1 个明确子任务。
如果证据不足，不要继续猜。
如果连续两轮没有新增证据，停止并汇报。
最终必须输出：结论、证据、修改建议、风险、下一步。
```

配置里也可以限制：

```jsonc
{
  "team_mode": {
    "max_wall_clock_minutes": 60,
    "max_member_turns": 120,
    "max_messages_per_run": 3000
  }
}
```

默认上限不等于推荐上限。

企业试点时要更保守。

## 11. 4SAPI 日志复盘看什么

Team Mode 结束后，不要只看结果。

要看 4SAPI 日志。

重点看：

```text
本次 Team Mode 总请求数。
各模型消耗。
各 Key 消耗。
失败率。
429 / 5xx 是否集中。
是否有长时间无效重试。
是否 Scout 花费过高。
是否 Reviewer 反复触发 fallback。
```

然后问四个问题：

```text
这次并行是否真的节省了时间？
哪些成员产出是重复的？
哪些成员应该换便宜模型？
这个任务下次还需要 Team Mode 吗？
```

如果一次 Team Mode 花了很多钱，但最后只有一个成员的结论有用，说明团队配置要调。

## 12. 成本复盘模板

可以让 Lead 最后输出：

```text
Team Mode 复盘：
1. 本次目标：
2. 成员分工：
3. 每个成员产出：
4. 哪些产出被最终采用：
5. 哪些产出重复或无效：
6. 修改文件：
7. 验证命令：
8. 风险点：
9. 建议下次是否继续使用 Team Mode：
10. 建议模型和并行数调整：
```

同时在 4SAPI 后台看：

```text
Key：4sapi-omo-teammode-dev
时间范围：本次任务开始到结束
模型分布：fast / coding / reasoning
错误码：429 / 5xx / 其他
总消耗：记录到项目复盘
```

这样团队才会越来越会用。

不复盘，Team Mode 只会越来越贵。

## 13. 一个安全审计场景

适合 Team Mode 的典型任务是安全审计。

可以这样拆：

```text
Lead：定义范围和汇总风险。
Member 1：检查认证和鉴权。
Member 2：检查输入校验和注入风险。
Member 3：检查依赖和配置。
Member 4：检查日志、密钥和敏感信息。
Reviewer：汇总可利用性和修复优先级。
```

但要写边界：

```text
只审计当前仓库。
不要攻击外部系统。
不要扫描公网资产。
不要读取真实密钥。
只输出可复现代码路径和修复建议。
```

4SAPI Key 建议单独用：

```text
4sapi-omo-teammode-security
```

并设置预算。

安全审计容易读很多文件，消耗不低。

## 14. 一个重构场景

大型重构也适合。

提示词可以写：

```text
使用 Team Mode 做重构前评估。

目标：
评估把 auth service 从旧接口迁移到新接口的影响范围。

约束：
第一阶段只读，不改文件。
每个成员只负责一个区域。
Lead 汇总后再决定是否进入实现。

成员建议：
- route-scout：查路由和 API。
- test-scout：查测试覆盖。
- type-scout：查类型和 DTO。
- migration-planner：整理最小迁移步骤。

停止条件：
60 分钟内完成，或所有成员完成一轮扫描后停止。

最终输出：
影响文件、风险、建议改动顺序、是否值得进入实现阶段。
```

第一阶段只读。

第二阶段再让一个执行 Agent 改。

这比四个成员同时改稳得多。

## 15. 不推荐的 Team Mode 用法

不推荐一：

```text
没有明确目标，只写“帮我优化项目”。
```

不推荐二：

```text
所有成员都用强模型。
```

不推荐三：

```text
所有成员都允许直接改文件。
```

不推荐四：

```text
max_parallel_members 一上来开 8。
```

不推荐五：

```text
Team Mode 和日常开发共用一个 4SAPI Key。
```

不推荐六：

```text
没有任务结束复盘。
```

这些都会把 Team Mode 从生产力工具变成成本黑洞。

## 16. 最后总结

Team Mode 是 OmO 里最适合企业大任务的能力之一。

但它不是日常默认按钮。

企业使用时，建议记住这套原则：

```text
默认关闭。
大任务开启。
先只读分工。
并行数从 2 开始。
Team Mode 单独 Key。
强模型只给 Lead 和 Reviewer。
Scout 用低成本模型。
任务有停止条件。
结束看 4SAPI 日志。
每次做成本复盘。
```

一句话：

```text
Team Mode 负责把复杂任务拆开。
4SAPI 负责防止拆开之后成本失控。
```

用好了，它是企业级多 Agent 协作。

用不好，它就是多个模型同时烧钱。

下一篇继续讲：

```text
OmO 的模型路由和 fallback 怎么配，才能在速度、质量和成本之间取得平衡。
```

## 资料来源与延伸阅读

- oh-my-openagent GitHub：https://github.com/code-yeongyu/oh-my-openagent
- oh-my-openagent Team Mode 文档：https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/guide/team-mode.md
- oh-my-openagent 配置文档：https://github.com/code-yeongyu/oh-my-openagent/blob/dev/docs/reference/configuration.md
- OpenCode Provider 文档：https://opencode.ai/docs/providers/
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 接入实操手册：https://4sapi.com/blog/4sapi-api-integration-setup-guide
