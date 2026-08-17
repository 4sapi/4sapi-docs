---
title: "Pi 生产使用如何划分模型角色与验证边界"
tags:
  - Pi
  - 生产实践
  - Agent开发
description: "Agent 进入生产环境后，模型角色、工具权限、日志和失败恢复必须有清晰边界。"
---
# Pi 生产使用如何划分模型角色与验证边界
Agent 进入生产环境后，模型角色、工具权限、日志和失败恢复必须有清晰边界。本文以 Intake、Planner、Executor、Reviewer 和 Summarizer 为例，说明如何按职责设计验证点，并把可观测性与人工审批放在同一条链路上。文中只讨论可在本地复现的步骤，不把单次结果扩展成产品承诺；每个结论都标注前提、证据和无法覆盖的边界。读者可以先完成最小验证，再根据自己的版本、权限和数据补充实验，避免把配置示例误当成普遍结论。

前面四篇已经完成了从 Pi Agent Harness、pi-ai、models.json 到 Coding Agent 的接入。个人实验到这里基本可以运行，但团队一旦把它放进自动化工作流，问题会从“模型能不能回答”变成：谁调用了模型、用了哪个 Key、花了多少钱、失败后是否停止、工具有没有越权。


## 1. 从“能调用”到“可生产”

一个一次性脚本可以只有一个模型和一个 Key：

~~~
脚本 -> 模型 API -> 输出
~~~

生产工作流通常包含多个阶段：

~~~
输入接收
  -> 任务分类
  -> 计划拆解
  -> 工具执行
  -> 结果审查
  -> 摘要和交付
~~~

如果每个阶段都使用同一个高成本模型、同一个万能 Key，最终会出现四类问题：

~~~
成本无法归属
失败无法定位
权限无法收回
模型无法按任务切换
~~~

治理的第一步不是加更多提示词，而是把请求按任务角色和风险拆开。

## 2. 五类模型角色

下面是一种适合 Pi 工作流的角色划分。它是业务编排方法，不是 Pi 内置的 Subagent 功能；可以用多个 Pi 会话、Extension 或业务调度器实现。

| 角色 | 任务 | 模型选择思路 | 工具权限 |
| --- | --- | --- | --- |
| Intake | 读取工单、分类、清洗输入 | 低成本、稳定、响应快 | 只读 |
| Planner | 拆分任务、定义边界、写计划 | 强推理或长上下文 | 通常只读 |
| Executor | 写代码、生成内容、调用业务工具 | 稳定的 Coding 或执行模型 | 按任务开放 |
| Reviewer | 审查事实、代码、风险和格式 | 强模型或专用 Reviewer | 只读优先 |
| Summarizer | 生成日报、Brief、归档摘要 | 低成本模型 | 只读 |

对应的请求流可以画成：

~~~
上游 API-PI-INTAKE      -> Intake 模型
上游 API-PI-PLANNER     -> Planner 模型
上游 API-PI-EXECUTOR    -> Executor 模型
上游 API-PI-REVIEWER    -> Reviewer 模型
上游 API-PI-SUMMARIZER  -> Summarizer 模型
~~~




~~~json
{
  "providers": {
    "provider-intake": {
      "name": "上游 API Intake",
      "baseUrl": "https://api.example.com/v1",
      "api": "openai-completions",
      "apiKey": "$FOURSAPI_INTAKE_KEY",
      "models": [
        { "id": "你的Intake模型精确ID", "name": "Intake model" }
      ]
    },
    "provider-executor": {
      "name": "上游 API Executor",
      "baseUrl": "https://api.example.com/v1",
      "api": "openai-completions",
      "apiKey": "$FOURSAPI_EXECUTOR_KEY",
      "models": [
        { "id": "你的Executor模型精确ID", "name": "Executor model" }
      ]
    },
    "provider-reviewer": {
      "name": "上游 API Reviewer",
      "baseUrl": "https://api.example.com/v1",
      "api": "openai-completions",
      "apiKey": "$FOURSAPI_REVIEWER_KEY",
      "models": [
        { "id": "你的Reviewer模型精确ID", "name": "Reviewer model" }
      ]
    }
  }
}
~~~

启动前注入不同的 Key：

~~~powershell
$env:FOURSAPI_INTAKE_KEY = "Intake令牌"
$env:FOURSAPI_EXECUTOR_KEY = "Executor令牌"
$env:FOURSAPI_REVIEWER_KEY = "Reviewer令牌"
~~~


选择某个角色的模型：

~~~bash
pi --model provider-intake/<精确模型ID> -p "把这条工单分类为 bug、需求或咨询。"
pi --model provider-reviewer/<精确模型ID> -p "只读审查当前变更并输出证据。"
~~~

## 4. 把请求和成本记录成一条事件

要做成本治理，不能只记录“今天余额少了多少”。每次模型请求至少应形成一条脱敏事件：

~~~json
{
  "event": "llm.request.completed",
  "requestId": "网关或应用生成的请求ID",
  "sessionId": "Pi会话ID",
  "project": "project-a",
  "workflow": "code-review",
  "provider": "provider-reviewer",
  "model": "上游 API模型精确ID",
  "keyGroup": "上游 API-PI-REVIEWER",
  "status": 200,
  "retryCount": 0,
  "durationMs": 0,
  "inputTokens": 0,
  "outputTokens": 0,
  "toolName": null
}
~~~


不要记录这些内容：

~~~
API Key 原文
完整用户提示词
完整业务文档
私钥、Cookie、数据库密码
未经脱敏的工具参数和返回值
~~~

建议只记录内容哈希、长度、类型和脱敏后的摘要。日志的目标是定位请求和成本，不是复制一份业务数据库。


Pi 侧适合回答：

~~~
哪个会话发起了请求？
哪个 Agent 角色调用了模型？
一轮任务跑了几次？
输入和输出 Token 是多少？
是否发生了工具续跑？
~~~


~~~
哪个令牌和分组被使用？
实际命中了哪个模型渠道？
请求是否被限流或拒绝？
账户实际扣费和倍率是多少？
余额、额度和期限是否正常？
~~~

两侧出现数字差异时，先检查：

| 检查项 | 影响 |
| --- | --- |
| 多轮工具调用 | 一次用户任务对应多次模型请求 |
| 自动重试 | 应用记录一次失败，网关可能记录多次请求 |
| Prompt Cache | 输入 Token 和计费 Token 可能分开 |
| 模型别名 | 本地名称和上游真实 ID 不一致 |
| 流式中断 | 应用未拿到完整结果，但网关已产生调用 |


## 6. 重试不是越多越稳定


| 状态 | 根因方向 | 处理方式 |
| --- | --- | --- |
| 400 | 请求协议、字段或模型能力 | 修正请求，不重试同一个错误 |
| 401/403 | Key、分组、额度或权限 | 更换合法配置或停止任务 |
| 429 | 并发、限流、余额和配额 | 读取 Retry-After，退避并限并发 |
| 500/503 | 上游渠道或网关暂时不可用 | 走白名单内备用模型或稍后重试 |
| 504/524 | 连接、请求体或客户端超时 | 缩小任务、调整超时或改用异步 |

对于 Coding Agent 或 Loop 任务，要同时设置：

~~~
单轮请求超时
单个任务最大模型轮数
单个任务最大 Token
连续失败停止次数
全局并发上限
单日预算和告警阈值
~~~

如果连续三次返回 400，继续重试只会增加日志噪声。如果连续 429，应该降低并发或切换经过批准的低成本模型，而不是让每个 Agent 同时重试。

## 7. 用 settings.json 限制 Provider 重试

Pi 文档提供了 Provider 级重试配置。一个保守的项目设置可以写成：

~~~json
{
  "retry": {
    "provider": {
      "maxRetries": 0,
      "maxRetryDelayMs": 60000
    }
  }
}
~~~

maxRetries: 0 的意义不是拒绝所有恢复，而是避免 SDK 层先于业务层无限等待。业务调度器可以根据错误类型做一次有上限的退避、切换备用模型或把任务放回队列。

不要把 maxRetries 调得很大来掩盖不稳定的 Provider。每增加一次重试，都要回答：谁承担额外成本、谁接收重复的工具动作、谁能停止这个循环。

## 8. 预算控制的最小模型

一个任务的预算可以用以下字段估算：

~~~
任务成本
  = 输入 Token 成本
  + 输出 Token 成本
  + 缓存读写成本
  + 工具续跑次数 × 后续请求成本
  + 重试请求成本
~~~

应用侧应该为每个任务保存：

~~~
预算上限
已消耗 Token
已调用轮数
重试次数
当前角色
停止原因
~~~

到达预算或轮数上限时，停止并输出结构化原因：

~~~text
status: stopped
reason: budget_limit
last_provider: provider-executor
last_model: 上游 API模型精确ID
retry_count: 1
next_action: human_review
~~~

这比让模型继续输出一句“我会继续尝试”更有用。成本治理的关键是让停止条件成为工作流状态，而不是一句提示词。

## 9. 生产上线清单

上线前至少检查以下内容：

~~~text
[ ] Key 使用环境变量、凭证库或 上游 API 令牌管理
[ ] 不同项目、环境和工作流没有共用万能 Key
[ ] 模型 ID 来自当前 上游 API 模型广场
[ ] Provider、模型和分组有白名单
[ ] Pi 工具按只读、修改、执行分级
[ ] 写文件和运行命令位于测试目录或外部沙箱
[ ] 记录 requestId、sessionId、provider、model 和状态码
[ ] 日志不保存 API Key、完整提示词和敏感工具结果
[ ] 任务有最大轮数、Token、并发和预算限制
[ ] 400、401/403、429、5xx、超时有不同处理
[ ] 重试有上限，连续失败会停止
[ ] 关键操作保留人工确认和回滚点
[ ] 上游 API 后台余额、分组和调用日志可查看
~~~

## 10. 一个可复用的企业工作流模板

把前四篇的能力组合起来，可以得到下面的最小模板：

~~~text
1. Intake 读取输入并分类，使用只读工具和低成本模型。
2. Planner 写出目标、范围、验收标准和停止条件。
3. Executor 在隔离工作目录执行有限工具调用。
4. Reviewer 重新读取变更和测试结果，输出证据。
5. Summarizer 生成脱敏后的交付摘要。
6. 调度器保存状态、Token、费用、错误和人工确认结果。
~~~

其中任何阶段失败，都应该把状态写成“失败原因 + 下一步”，而不是自动跳过继续执行。例如 Executor 遇到 429，应该回到队列；Reviewer 遇到 400，应该回到请求格式检查；涉及生产写入的任务则应进入人工确认。

## 11. 本篇小结


~~~text
Pi        -> Agent 状态、工具权限、项目上下文和会话
上游 API     -> API 入口、模型分组、Key、额度、渠道和网关日志
业务平台  -> 项目归属、预算、审批、队列和交付记录
~~~

没有 Agent 状态，任务无法复盘；没有 API 网关，模型成本和权限无法集中管理；没有业务平台约束，自动化就无法稳定交付。

这五篇系列的最终落点不是“把某个模型接通”，而是建立一条能回答下面问题的链路：

~~~text
谁发起了调用？
调用了哪个模型？
使用了哪个 Key 和分组？
花了多少 Token 和费用？
失败后为什么停止？
下一步由谁确认？
~~~



## 资料来源

- Pi GitHub：<https://github.com/earendil-works/pi>
- Pi Models：<https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/models.md>
- Pi Settings：<https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/settings.md>

## 结论

本文给出了问题定位、配置或验证的可执行路径。实际结果仍取决于当前版本、权限和运行环境，提交前应按官方文档复核可变字段，并保留失败证据和回滚边界。
