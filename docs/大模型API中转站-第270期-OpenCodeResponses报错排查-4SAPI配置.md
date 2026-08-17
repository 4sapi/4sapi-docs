---
title: "工具调用参数报不兼容时先核对哪四项"
category: 开发工具
tags:
  - OpenAI API
  - 工具调用
  - API 调试
description: "当编程客户端报告模型、工具或推理参数不兼容时，从最终请求、端点 Schema、实际模型和代理改写四层收集证据。"
---

# 工具调用参数报不兼容时先核对哪四项

编程客户端调用模型时，服务端可能返回“某工具与参数组合不支持”之类错误。它不一定来自提示词或工具 JSON Schema，也不能直接推出“改成 Responses”或“关闭推理”就是正确修复。客户端配置会经过 provider 适配器和代理，最终端点、模型标识、工具结构与推理字段可能已经变化。排障第一步应获取最终序列化请求和原始错误，再分别核对端点 Schema、实际模型能力、客户端协议模式和中间层改写。本文给出四层检查顺序。

## 第一项：获取最终请求而不是界面配置

对同一次调用建立 `request_id`，记录：

```text
HTTP 方法与最终 URL
请求中的实际模型标识
顶层字段名与类型
tools 中每个工具的 type 和 name
推理配置字段及值
上游原始状态与错误类别
客户端最终显示的错误
```

日志不保存 Authorization、Cookie、完整提示词、工具参数值和项目文件正文。字段存在性和类型通常足以定位 Schema 差异。

界面中选择了 Responses，不代表适配器最终发送到 `/v1/responses`；删除了某字段，也不代表模型预设或代理没有重新注入。以出站请求为准。

## 第二项：按实际端点核对 Schema

Chat Completions 和 Responses 的字段结构不同。当前 OpenAI [Chat Completions create reference](https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create) 定义 `messages`、外部包裹的函数工具、`reasoning_effort` 等字段；具体模型不一定支持所有值。

Responses 使用 `input`、扁平函数定义、类型化 `output` Items 和 `reasoning` 对象。迁移差异以 [Responses migration guide](https://developers.openai.com/api/docs/guides/migrate-to-responses) 为准。

检查请求是否混用了两套结构，例如：

```text
向 Responses 发送 messages 却仍按 choices 解析；
向 Chat Completions 发送 Responses 的扁平 function；
只改 URL，没有迁移工具结果与 call_id；
客户端使用一个端点的推理字段形状，代理转向另一个端点。
```

## 第三项：核对实际模型而非别名印象

模型是否支持端点、工具类型和推理值，需要查看当前官方模型文档与接口错误。记录请求模型和响应或代理实际使用的模型。若模型名来自第三方别名，不能用公开模型名称相似性推断能力。

建立能力矩阵：

| 字段 | 记录值 |
| --- | --- |
| 请求模型 | 客户端传入的标识 |
| 实际模型 | 上游返回或路由日志中的标识 |
| 端点 | Chat Completions 或 Responses |
| 工具类型 | function、内置工具或其他 |
| 推理字段 | 字段形状和实际值 |
| 官方依据 | 当前模型与 API 文档链接 |
| 实测结果 | 请求 ID 与原始状态 |

如果文档没有说明某个组合，使用无副作用最小请求验证，并把结果限制在该账号和版本。

## 第四项：检查客户端与代理的转换

分别保存：

```text
client_config：客户端选择的 provider 与能力
client_outbound：客户端实际发出的结构摘要
proxy_normalized：代理规范化后的字段摘要
upstream_outbound：最终上游结构摘要
```

用同一个 `request_id` 比较字段增加、删除、改名和模型映射。若代理将 Responses 转成 Chat Completions，必须确认它能无损处理所用 Items、工具结果和状态；普通文本成功不能证明工具链也兼容。

## 先构造最小无副作用请求

不要拿完整代码仓库和写文件工具排障。先定义一个不会产生外部影响的函数：

```json
{
  "name": "echo_label",
  "description": "Return an approved test label",
  "parameters": {
    "type": "object",
    "properties": {
      "label": { "type": "string", "enum": ["test"] }
    },
    "required": ["label"],
    "additionalProperties": false
  }
}
```

按所测端点把它放入正确的工具结构，固定模型和其他参数。若最小工具仍失败，问题位于协议、模型或路由；若成功，再逐步增加真实客户端中的字段，每次只加一类。

## 用单变量矩阵定位冲突

| 变体 | 端点 | 工具 | 推理配置 | 用途 |
| --- | --- | --- | --- | --- |
| A | 固定 | 无 | 固定 | 基础模型请求 |
| B | 固定 | 最小函数 | 不发送 | 验证工具支持 |
| C | 固定 | 无 | 待测值 | 验证推理字段 |
| D | 固定 | 最小函数 | 待测值 | 验证组合 |

模型、身份、输入和路由保持不变。只有 A、B、C 成功而 D 稳定失败，才能把问题收窄到组合兼容；否则先处理单项能力或环境错误。

## 何时迁移到 Responses

如果应用需要类型化 Items、Responses 工具循环或其状态能力，应按完整迁移流程改请求、输出解析和状态管理，而不是把切换端点当作错误绕行。

迁移前确认客户端 provider 明确支持 Responses，代理不会降级成 Chat Completions，目标模型支持所需能力。迁移后重新验证纯文本、函数 `call_id`、结构化输出、流式事件和错误处理。

## 何时调整推理配置

只有当前模型官方文档与接口支持矩阵说明某值不兼容，或单变量实验稳定证明配置导致失败时，才调整它。调整后视为行为变更，用代表性仓库任务验证代码质量、工具步骤、延迟和用量。

删除推理配置也可能触发模型默认值，不能假设等同于某个显式值。查看最终请求和响应中的实际设置。

## 不要通过删除工具掩盖问题

移除工具可以让纯文本调用成功，却不代表编程 Agent 修复。没有文件读取、测试和修改能力后，客户端已经变成另一个产品行为。

工具必须由控制层执行参数校验、目录限制和审批；排障时禁用高风险工具，但正式验收要恢复最小必要工具并验证完整闭环。

## 结论与限制

工具与推理参数报不兼容时，应先确认最终请求，分别核对端点 Schema、实际模型、客户端适配器和代理改写，再用单变量最小请求定位。不要把某条模型专属错误推广成所有 Chat Completions 或 Responses 的规则。

本文依据 2026 年 7 月 27 日访问的 OpenAI 官方 API 参考与迁移指南。它不包含 OpenCode 或任何第三方 provider 的当前配置承诺；第三方能力必须通过其官方文档和实际出站请求单独验证。
