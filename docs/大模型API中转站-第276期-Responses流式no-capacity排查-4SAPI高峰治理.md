---
title: "Responses 流式 no_capacity 如何区分容量与网络问题"
tags:
  - Responses API
  - 流式处理
  - 故障排查
description: "Responses 流式请求出现 no_capacity 时，客户端可能只看到连接中断，无法区分容量、限流和网络超时。"
---
# Responses 流式 no_capacity 如何区分容量与网络问题
Responses 流式请求出现 no_capacity 时，客户端可能只看到连接中断，无法区分容量、限流和网络超时。本文按响应状态、重试头、连接阶段和本地日志建立排查顺序，并说明哪些现象不能直接推断为服务端容量不足。文中只讨论可在本地复现的步骤，不把单次结果扩展成产品承诺；每个结论都标注前提、证据和无法覆盖的边界。读者可以先完成最小验证，再根据自己的版本、权限和数据补充实验，避免把配置示例误当成普遍结论。

使用 Codex 或其他 Agent 调用 Responses 流式接口时，如果看到下面的错误，不要先去重装客户端：

```text
OpenAI responses stream error: no capacity
The system is currently experiencing high demand and cannot process your request.
Your request exceeds the maximum usage size allowed during peak load.
For improved capacity reliability, consider switching to Provisioned Throughput.
```

这类错误通常还有一个 `request_id`，例如：

```text
code: no_capacity
param: ""
request_id: <REQUEST_ID>
```

它说明请求已经到达上游服务，但高峰时段没有足够的可用推理容量。问题重点在服务端资源，而不是本地网络、提示词或业务代码。

## 1. 先判断它到底是哪一类错误

很多团队把所有失败都叫“限流”，导致排错方向完全错误。可以先按错误码区分：

| 错误类型 | 常见信号 | 主要含义 | 首要动作 |
| --- | --- | --- | --- |
| `no_capacity` | 高峰期、提示容量不足 | 上游临时推理资源不足 | 错峰、有限退避、切备用通道 |
| `429` | rate limit、quota、too many requests | 账户或通道达到速率/额度限制 | 降低并发、按 Retry-After 退避 |
| `408/超时` | 请求等待过久 | 网络、代理或上游响应超时 | 检查链路和超时配置 |
| `5xx` | 500、502、503 | 上游服务异常或网关错误 | 有限重试、记录 request_id |
| `400` | 参数、模型或协议不兼容 | 请求本身不符合接口要求 | 修正请求，不重试原请求 |

`no_capacity` 和 `429` 都可能出现在高峰期，但处理方式不同：前者是“当前没有足够资源处理这个请求”，后者更像“当前账户或通道的使用速率超限”。

## 2. 为什么流式请求更容易暴露容量问题

Responses 流式调用不是一次性返回完整 JSON，而是让上游持续发送事件：

```text
response.created
response.output_text.delta
response.function_call_arguments.delta
response.completed
```

一次流式请求可能持续更久、占用连接更久，也可能包含多个工具步骤。高峰时，服务端需要同时安排：

```text
模型推理资源。
流式连接资源。
上下文和输出 Token 资源。
工具调用的后续步骤。
```

因此，短文本请求能成功，并不代表长上下文、复杂 reasoning 或 Coding Agent 流式请求也有同样的容量。

## 3. 第一层排查：确认不是本地问题

### 3.1 记录完整但脱敏的响应

至少保留：

```text
HTTP 状态码
错误 code
错误 type
request_id
endpoint
model
是否 stream
输入 Token 估算
重试次数
```

不要只截取“请检查网络连接”这一句。很多客户端会把所有 API 失败都包装成网络提示，实际错误码仍然是上游的 `no_capacity`。

### 3.2 做一个最小非流式对照

在隔离环境中，把同一模型、同一 Key 和相近输入改成短文本、非流式请求：

```python
import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["API_KEY"],
    base_url=os.environ.get("API_BASE_URL", "https://api.example.com/v1"),
)

response = client.responses.create(
    model="gpt-5.6-sol",
    input="返回 OK",
)

print(response.output_text)
```

如果短请求成功、原来的长流式请求失败，说明问题更接近容量或请求规模，而不是 Key 完全失效。

如果连最小请求也持续失败，再检查模型权限、endpoint、账户状态和网关通道。

## 4. 第二层排查：确认请求规模

错误信息中的“maximum usage size”不一定只指账户余额，也可能与当前峰值下允许调度的请求规模有关。建议把一次请求拆成四个指标：

```text
输入上下文长度。
最大输出 Token。
推理强度和预计计算量。
工具调用步骤数。
```

以下几种请求更容易在高峰时失败：

- 一次塞入整个代码仓库；
- 开启高推理档位并要求长输出；
- 同时注册大量工具；
- 多个 Agent 并行发送流式请求；
- 失败后立即发起完全相同的重试。

排查时先做最小化实验，而不是同时修改模型、提示词、网络和客户端版本。一次只改变一个变量，才能知道是什么因素让请求恢复。

## 5. 临时恢复方案：有限退避与错峰

对于 `no_capacity`，可以采用有限次数的指数退避：

```python
import random
import time

def backoff(attempt: int) -> None:
    base = min(2 ** attempt, 30)
    time.sleep(base + random.uniform(0, 1))

for attempt in range(3):
    try:
        response = client.responses.create(
            model="gpt-5.6-sol",
            input="执行一个可重试的只读任务",
        )
        break
    except Exception as exc:
        if "no_capacity" not in str(exc) or attempt == 2:
            raise
        backoff(attempt)
```

这段代码只适合幂等的只读任务。涉及写文件、提交代码、发送消息、扣费或外部系统写入时，必须先确认任务是否可能已经执行成功，不能盲目重放。

不要无限重试。三次失败后应该：

```text
记录 request_id。
触发告警。
切换备用模型或备用通道。
将任务放入队列等待。
必要时交给人工处理。
```



```text
主模型 no_capacity
        ↓
有限退避一次
        ↓
备用模型或备用通道
        ↓
仍失败则进入队列并告警
```

一个简单的降级矩阵如下：

| 任务类型 | 主路径 | 容量不足时的处理 |
| --- | --- | --- |
| 短文本摘要 | 均衡模型 | 切成本更低的备用模型 |
| 批量分类 | 低成本模型 | 降低并发，转异步队列 |
| Coding Agent | 高能力 reasoning 模型 | 保留任务状态，稍后重试或人工接管 |
| 客服实时回复 | 低延迟通道 | 切换预设回复或备用模型 |
| 高风险分析 | 指定模型 | 不自动换模型，进入人工复核 |

模型降级不能只看“请求能否返回”。还要验证输出格式、工具能力、上下文长度和业务质量。如果主路径支持 function tools，备用路径不支持，就必须在路由层明确禁止该任务切换过去。

## 7. Provisioned Throughput 什么时候值得考虑

原始错误建议使用 Provisioned Throughput。它的核心思路是为业务预留更稳定的处理容量，减少公共共享资源高峰期的波动。

它不是所有项目都应该立即购买的方案。可以用下面的判断：

```text
请求是否全天候运行？
是否有明确的稳定吞吐需求？
高峰失败是否会造成业务损失？
是否已经完成请求压缩和并发治理？
预留容量成本是否低于失败和人工兜底成本？
```

适合优先评估的场景：

- 对延迟和成功率有明确 SLA 的在线接口；
- 持续运行的客服、知识库和 Agent 工作流；
- 高峰失败会直接影响订单、生产或客户交付的业务；
- 已经通过日志统计出稳定的峰值吞吐。

如果只是偶尔使用的个人 Coding Agent，先做错峰、压缩上下文、降低并发和有限回退，通常比立即购买预留容量更合理。具体产品名称、价格和可用性以服务商当前控制台为准。



```text
request_id
project_id
team_id
environment
model
endpoint
stream
input_tokens
output_tokens
concurrency
retry_count
fallback_model
status_code
error_code
latency_ms
created_at
```

然后设置三类告警：

```text
no_capacity 比例超过阈值。
某个模型的失败集中发生在特定时段。
重试和回退造成的 Token 成本异常增长。
```

在企业 API 网关中按项目和环境拆分 Key，可以进一步判断是全局容量问题，还是某个项目的并发配置不合理。这些数据也是成本治理的基础：只有把失败、重试和回退成本关联到项目，预算告警才不会变成孤立的数字。

## 9. 容量不足、限流和回退的治理规则

建议把错误处理写成明确的状态机：

```text
收到 no_capacity
  ├─ 幂等只读任务：退避一次
  ├─ 仍失败：切备用模型或进入队列
  ├─ 高风险写入：暂停自动重试，等待确认
  └─ 连续异常：触发告警并检查容量配置
```

不要把所有 400、401、403 都放入自动重试。参数错误、Key 权限错误和容量错误的修复动作完全不同。

同时给 Agent 设置：

- 最大执行步骤数；
- 最大等待时间；
- 最大重试次数；
- 单任务预算；
- 失败后的人工接管入口。

## 10. 成本与风险提示

容量治理本身也有成本。Provisioned Throughput 适合有稳定吞吐和 SLA 的业务，不适合没有流量基线的试验项目。

重试和回退也会增加账单。应把下面的指标放在一起看：

```text
成功任务成本
= 原始模型调用
+ 退避重试
+ 备用模型调用
+ 工具调用
+ 人工兜底
```


## 11. 上线前检查清单

```text
[ ] 已区分 no_capacity、429、400、401/403 和 5xx。
[ ] 最小非流式请求可以独立验证。
[ ] 长上下文和高并发请求有单独监控。
[ ] no_capacity 只进行有限退避。
[ ] 写入类任务已确认幂等性和重复执行风险。
[ ] 上游 API 模型、endpoint、Key 权限和通道已核对。
[ ] 有备用模型、队列或人工接管路径。
[ ] Provisioned Throughput 的成本与 SLA 需求已评估。
[ ] 日志记录 request_id、错误码、模型和重试。
[ ] 预算、限流、告警和回滚配置已验证。
```

## 12. 总结

Responses 流式请求出现 `no_capacity` 时，优先判断服务端容量，而不是修改本地网络或反复重装 Codex。




## 官方来源

- OpenAI：Migrate to the Responses API：https://developers.openai.com/api/docs/guides/migrate-to-responses
- OpenAI：Responses API reference：https://developers.openai.com/api/reference/responses

## 结论

本文给出了问题定位、配置或验证的可执行路径。实际结果仍取决于当前版本、权限和运行环境，提交前应按官方文档复核可变字段，并保留失败证据和回滚边界。
