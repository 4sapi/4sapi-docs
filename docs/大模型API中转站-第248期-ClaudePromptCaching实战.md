---
title: "用 Python 验证 Claude Prompt Caching 创建与读取"
category: 人工智能
tags:
  - Claude API
  - Prompt Caching
  - Python
description: "本文用相同长前缀和两个不同问题发起请求，记录缓存创建与读取 usage，并通过修改前缀的对照请求验证命中条件。"
---

# 用 Python 验证 Claude Prompt Caching 创建与读取

请求里出现 `cache_control` 不代表服务端已经创建或读取缓存。模型不支持、文档太短、前缀发生细微变化或第二次请求超出有效期，都可能让测试只有普通输入 usage。本文只解决一个可执行问题：准备一份脱敏长文档，用 Python 在同一模型上发送两个仅末尾问题不同的请求，分别保存 `cache_creation_input_tokens` 与 `cache_read_input_tokens`，再修改前缀做对照。结果只用于确认当前模型、SDK 和请求结构下的缓存行为。测试还要保存完整 usage，而不是只摘取一个数字，否则后续无法区分模型输出变化与缓存行为。

## 1. 先核对当前接口条件

[Claude Prompt Caching 官方文档](https://platform.claude.com/docs/en/build-with-claude/prompt-caching) 列出了当前支持方式、模型条件、有效期和 usage 字段。开始前先从当前账号或官方模型列表取得模型 ID，不要从旧文章复制日期版名称。

本文代码读取三个环境变量：

```text
ANTHROPIC_API_KEY：当前测试凭据。
ANTHROPIC_MODEL：当前文档确认支持缓存的模型 ID。
ANTHROPIC_MAX_TOKENS：本次回答允许的最大输出 token。
```

凭据只放在本地环境中，不写入脚本、文档或版本库。

## 2. 准备环境和测试文档

在隔离的 Python 环境中安装当前 Anthropic SDK：

```powershell
python -m pip install anthropic
```

在 PowerShell 当前会话设置环境变量，值由当前账号和测试计划提供：

```powershell
$env:ANTHROPIC_API_KEY="<test-key>"
$env:ANTHROPIC_MODEL="<supported-model-id>"
$env:ANTHROPIC_MAX_TOKENS="<approved-output-limit>"
```

准备 `company-policy.md`。文档应使用公开材料或脱敏测试数据，并达到当前模型文档要求的可缓存长度。不要用重复无意义句子凑长度，也不要使用生产合同、个人信息或密钥。

## 3. 发送两个共享前缀的请求

新建 `cache_check.py`：

```python
import json
import os
from pathlib import Path

import anthropic


client = anthropic.Anthropic(
    api_key=os.environ["ANTHROPIC_API_KEY"],
)

model = os.environ["ANTHROPIC_MODEL"]
max_tokens = int(os.environ["ANTHROPIC_MAX_TOKENS"])
policy_text = Path("company-policy.md").read_text(encoding="utf-8")


def ask(question: str) -> dict:
    response = client.messages.create(
        model=model,
        max_tokens=max_tokens,
        system=[
            {
                "type": "text",
                "text": (
                    "只依据给定制度回答。资料没有说明时，"
                    "明确写出无法确认。"
                ),
            },
            {
                "type": "text",
                "text": policy_text,
                "cache_control": {"type": "ephemeral"},
            },
        ],
        messages=[
            {
                "role": "user",
                "content": question,
            }
        ],
    )

    usage = response.usage.to_dict(mode="json")
    summary = {
        "input_tokens": usage.get("input_tokens"),
        "output_tokens": usage.get("output_tokens"),
        "cache_creation_input_tokens": usage.get(
            "cache_creation_input_tokens"
        ),
        "cache_read_input_tokens": usage.get(
            "cache_read_input_tokens"
        ),
    }
    record = {
        "usage": usage,
        "summary": summary,
    }
    print(json.dumps(record, ensure_ascii=False, sort_keys=True))
    return record


first = ask("材料中如何描述差旅报销？")
second = ask("材料中如何描述发票遗失？")
```

运行位置应是 `cache_check.py` 与 `company-policy.md` 所在目录：

```powershell
python .\cache_check.py
```

两次调用使用相同模型、system 块、文档内容和块顺序，只有最后的用户问题不同。

## 4. 如何读 usage

程序不会把“命中”写死为成功。当前 Anthropic Python SDK 的响应模型提供 `to_dict(mode="json")`，用于取得 API 字段名对应的 JSON 安全数据。代码把完整 `usage` 与便于比较的四字段 `summary` 放在同一条 JSON 记录中；保存两次输出后再按当前官方文档解释：

```text
第一次请求：检查是否记录缓存创建输入。
第二次请求：检查是否记录缓存读取输入。
两次请求：保存完整 usage，并用 summary 对照普通输入、输出和缓存字段。
```

若第二次出现缓存读取值，说明该请求在当前条件下复用了前缀。若只有创建值而没有读取值，不能声称缓存已经带来复用；需要检查模型支持、文档长度、请求间隔和前缀一致性。

usage 字段证明的是缓存处理情况，不证明回答正确。两次回答仍需分别核对是否只引用测试文档，以及资料不足时是否拒绝猜测。

## 5. 做一次破坏前缀的对照

为了排除偶然判断，可以复制测试记录后修改稳定前缀，例如更换 `company-policy.md` 的内容版本，再以同一模型重新请求。

对照实验要保证：

```text
只修改一个前缀变量。
保留修改前后的完整请求结构。
分别记录创建与读取 usage。
不把随机标识或当前时间加入正式稳定前缀。
```

如果多个变量同时变化，无法判断是哪一项破坏了复用。文档真实更新时应接受缓存失效，而不是为了命中继续使用旧内容。

## 6. 常见失败如何定位

### 没有创建 usage

检查当前模型是否支持缓存、文档是否达到当前要求，以及 `cache_control` 是否位于官方允许的内容块。SDK 能发送请求不等于该请求满足缓存条件。

### 第二次没有读取 usage

对比两次请求的模型、system 内容、文档字节、消息顺序和执行间隔。空格、换行、序列化顺序或动态字段都可能改变前缀。

### 运行时报缺少环境变量

确认变量设置在运行脚本的同一 PowerShell 会话。不要通过给代码添加明文默认 Key 来绕过错误。

### 回答与文档不一致

这是内容正确性问题，与缓存命中分开处理。保留原始输出，检查 system 规则和材料是否足以支持问题；证据不足时缩小问题或结论。

## 7. 验收记录

一次完整记录至少包括：

```text
测试模型和 SDK 版本。
官方文档核对时间。
文档内容哈希或受控版本号。
两次问题和完整 usage。
前缀对照实验的唯一变化。
回答事实核对结果。
未验证项和错误信息。
```

不要在公开记录中保存 Key、敏感制度原文或完整用户数据。

## 8. 结论与限制

这段程序把缓存验收拆成两类证据：usage 用来确认创建和读取，回答内容用来确认事实正确性。只有第二次请求在相同前缀条件下出现官方定义的读取字段，才能说明本次测试发生了复用。

模型支持、最低长度、有效期、SDK 参数和计费都可能变化，本文不提供固定值或成本承诺。代码仅针对核对时的 Anthropic Python SDK 请求形态；实施前必须重新查看官方文档，并在脱敏测试环境中运行。
