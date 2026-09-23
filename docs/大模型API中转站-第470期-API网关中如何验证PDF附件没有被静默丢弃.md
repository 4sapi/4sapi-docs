---
title: "API网关中如何验证PDF附件没有被静默丢弃"
category: 后端开发
tags:
  - OpenAI API
  - 文件处理
  - API网关
description: "说明协议转换导致 PDF 附件被忽略时，如何通过哨兵内容、请求体检查和回归测试定位问题。"
---

# API网关中如何验证PDF附件没有被静默丢弃

给模型发送 PDF 后得到 HTTP 200，并不代表模型真的收到了文件。兼容网关若把一种消息格式转换成另一种格式，最危险的故障不是明确报错，而是转换器只保留文字、悄悄跳过文件部分。客户端看到的是正常响应，业务却得到“未收到附件”之类的答案。下面给出一套不依赖特定网关的验证方法：用可识别的测试文件、检查转换后的请求体，并用单元测试守住映射规则。

这里讨论的是“消息格式转换”场景，不讨论文件上传权限、模型是否具备 PDF 解析能力，或生产文档的内容安全。

## 问题与适用场景

适用条件是：客户端向兼容接口发送带文件的内容部分，网关再把请求改写为 Responses 请求。转换器至少需要把文件的标识、文件数据和文件名保留下来，形成目标协议的 `input_file` 内容部分。

OpenAI 的[文件输入文档](https://developers.openai.com/api/docs/guides/file-inputs)说明了 Responses API 可使用文件输入，并给出了 `input_file` 的输入形式。原始客户端所使用的“`file` 内容部分”不应被当成 OpenAI 公开 Chat Completions 端点的通用契约；它是兼容层需要明确维护和测试的输入约定。

一个已合并的项目修复正好说明了这种风险：该项目的适配器此前只处理文本和图片，文件部分被静默跳过；[Issue #5485](https://github.com/Wei-Shaw/sub2api/issues/5485) 与 [PR #5487](https://github.com/Wei-Shaw/sub2api/pull/5487) 记录了 `file` 到 `input_file` 的映射及三项回归测试。PR 于 2026-08-22 合并。

这类问题之所以难排，是因为它跨越了三个互不相同的成功条件：HTTP 层的“请求送达”、协议层的“字段被保留”、业务层的“模型确实使用了文件”。排查时必须逐层缩小范围，不能把其中任一层的成功当成另外两层的证据。

```text
客户端内容部分
    -> 入站校验
    -> 协议转换
    -> 出站请求体检查
    -> 模型请求
    -> 哨兵内容验证
```

## 先分清三种“成功”

| 检查层 | 可以证明什么 | 不能证明什么 | 常见误判 |
| --- | --- | --- | --- |
| HTTP 状态码 | 网关接受了请求，且得到上游响应 | 文件是否进入上游请求 | 把 200 当成附件已被读取 |
| 出站 JSON | 转换器保留了 `input_file` 结构 | 文件内容是否有效、模型是否能解析 | 只看日志里有“file”字样 |
| 哨兵语义 | 该次请求的模型输出可由测试文件解释 | 所有文件格式、所有模型都可用 | 用通用答案代替唯一短语 |

建议将这三项写进测试用例名称，例如 `maps_file_part`、`emits_input_file`、`reads_pdf_sentinel`。名称直接描述可观察结果，排障时不必猜测“附件测试失败”究竟是哪一层失败。

## 前置条件

- 在测试环境准备一个不含敏感信息的 PDF，正文放入唯一哨兵短语，例如 `PDF-CHECK-7F3A`。
- 能查看网关出站请求的结构化字段，或能为转换函数添加单元测试。不要在生产日志记录 base64 文件内容、完整提示词或访问令牌。
- 明确入站、出站各自的协议。本文示例把入站 `text` / `file` 映射为 Responses 的 `input_text` / `input_file`。
- 本地验证代码使用 Python 3.10 或更高版本，不依赖第三方包。

测试 PDF 应只有一项任务：包含一个不会出现在提示词、文件名或测试代码中的短语。不要把测试用真实合同、用户上传件或截图替换成“更接近生产”的材料。附件链路的验证只需要确认数据是否被保留，真实文档会增加脱敏、访问控制和留存风险。

## 明确转换契约

在写代码前，先把允许的输入和必须保留的字段列成契约。对于本文的最小适配器，规则如下：

| 入站内容部分 | 出站内容部分 | 必须保留 | 处理边界 |
| --- | --- | --- | --- |
| `text` | `input_text` | `text` | 文本为空时由调用方决定是否允许 |
| `file` 且有 `file_data` | `input_file` | `file_data`、可选文件名 | 不解析或修改 base64 内容 |
| `file` 且有 `file_id` | `input_file` | `file_id`、可选文件名 | ID 的归属与有效期由上游 API 决定 |
| `file` 但没有数据或 ID | 不产生出站部分 | 无 | 记录结构化校验失败，不伪造空文件 |
| 未知类型 | 无 | 无 | 显式失败，避免静默降级为文本 |

最后一条很重要。新增内容类型时，默认跳过往往能让接口“看起来可用”，却会把不兼容变成业务数据丢失。对没有明确定义映射的类型，测试环境应失败并要求维护者做出选择。

## 先验证转换规则

下面的代码是一个最小转换器。它只接受三类输入：文本、带 `file_data` 的文件，或带 `file_id` 的文件。没有文件数据和文件 ID 的文件部分被跳过，避免制造一个无法解析的空附件。

```python
def convert_content_parts(parts: list[dict]) -> list[dict]:
    converted = []

    for part in parts:
        kind = part.get("type")
        if kind == "text":
            converted.append({"type": "input_text", "text": part["text"]})
            continue

        if kind == "file":
            source = part.get("file") or {}
            file_data = source.get("file_data")
            file_id = source.get("file_id")
            if not (file_data or file_id):
                continue

            target = {"type": "input_file"}
            if source.get("filename"):
                target["filename"] = source["filename"]
            if file_data:
                target["file_data"] = file_data
            if file_id:
                target["file_id"] = file_id
            converted.append(target)
            continue

        raise ValueError(f"unsupported content part: {kind!r}")

    return converted


parts = [
    {"type": "text", "text": "只返回 PDF 中的哨兵短语。"},
    {
        "type": "file",
        "file": {
            "filename": "probe.pdf",
            "file_data": "data:application/pdf;base64,VEVTVA==",
        },
    },
]

converted = convert_content_parts(parts)
assert converted[0] == {"type": "input_text", "text": "只返回 PDF 中的哨兵短语。"}
assert converted[1]["type"] == "input_file"
assert converted[1]["filename"] == "probe.pdf"
assert converted[1]["file_data"].startswith("data:application/pdf;base64,")
print("file mapping checks passed")
```

将代码保存为 `verify_file_mapping.py` 后，在空的本地测试目录运行 `python verify_file_mapping.py`。预期输出是 `file mapping checks passed`。这一步只验证转换器的结构，不需要调用模型，也不应使用真实客户文件。

接下来再补两组纯结构测试。第一组验证通过文件 ID 引用的情况，第二组验证空文件不会被伪装成 `input_file`：

```python
by_id = convert_content_parts([
    {"type": "file", "file": {"filename": "existing.pdf", "file_id": "file_test_123"}}
])
assert by_id == [{"type": "input_file", "filename": "existing.pdf", "file_id": "file_test_123"}]

empty_file = convert_content_parts([
    {"type": "file", "file": {"filename": "missing.pdf"}}
])
assert empty_file == []
```

将这段追加到同一个脚本后再次执行。它不会访问网络，但能防止以后维护结构体时只覆盖 `file_data`、忘记 `file_id`，或把空对象一路传到上游。

## 在集成环境做三层验证

第一层检查出站 body。将上例的 `converted` 作为用户输入的一部分后，检查其中同时存在 `input_text` 和 `input_file`。如果请求只保留了文字，故障已经发生在网关，不必先怀疑模型。

出站检查应在“最终发送上游之前”完成。只检查入站 body 没有意义，因为错误恰好可能发生在中间转换阶段。推荐记录经过脱敏的结构摘要，而不是完整请求：

```json
{
  "content_part_types": ["input_text", "input_file"],
  "input_file_count": 1,
  "has_file_id": false,
  "has_file_data": true,
  "file_name_count": 1
}
```

第二层检查上游响应的基础信号。若 SDK 或服务端能提供输入计量，可把“有附件请求与无附件对照请求的输入计量差异”作为辅助观察；它不能单独作为通过标准，因为计量方式会因 API、模型和文件解析策略而变化。

第三层检查语义结果。向测试 PDF 写入唯一短语，要求模型只返回该短语。成功条件应同时满足：

1. 网关出站 body 中存在 `input_file`。
2. 响应包含唯一短语 `PDF-CHECK-7F3A`。
3. 删除文件部分后，响应不再能够得到该短语。

第三项是对照组。它能避免“模型恰好从提示词猜中答案”被误判为文件已送达。

一个完整的集成测试可按下面的顺序执行：

1. 发送带 `input_file` 的请求，保存脱敏后的出站结构摘要。
2. 让模型返回唯一短语，检查输出精确匹配或包含该短语。
3. 用相同提示词、相同模型、但不带文件再执行一次。
4. 断言无文件请求不能得到唯一短语；若得到，说明测试提示词或模型知识泄漏了哨兵值，测试无效。
5. 对照两次出站结构。唯一预期差异是文件相关部分，不能顺带改变模型、系统提示词或工具配置。

## 把验证写进持续集成

结构转换特别适合单元测试，真实模型调用则更适合少量受控的集成测试。两者不要混在同一组快测里：前者应在每次提交运行，后者可能需要凭证、网络和隔离测试资源。

| 测试名称 | 输入 | 断言 | 运行位置 |
| --- | --- | --- | --- |
| 文本映射 | 一个 `text` 部分 | 输出一个 `input_text` | 每次提交 |
| 文件数据映射 | 有 `file_data` 的文件部分 | 输出含 `input_file` 和原始数据字段 | 每次提交 |
| 文件 ID 映射 | 有 `file_id` 的文件部分 | 输出保留 ID | 每次提交 |
| 空文件拒绝 | 无数据、无 ID 的文件部分 | 不生成文件输入，并记录校验失败 | 每次提交 |
| 哨兵 PDF | 无敏感测试文件 | 文件短语只在有文件时出现 | 受控集成环境 |
| 未知类型 | 新增或错误类型 | 明确失败而非静默跳过 | 每次提交 |

在 CI 中，结构测试不需要真实 PDF：只需用伪造的 data URL 验证字段是否保留。真正的 PDF 测试应限于受控环境，且将测试文件和凭证与业务数据完全隔离。

## 常见失败与排查

### HTTP 200，但回答没有附件内容

先比较有文件和无文件两次请求的出站 body。若两者的内容部分完全相同，或都没有 `input_file`，应检查转换器的 `switch`/`if` 分支及数据结构字段，而不是反复重试模型请求。

### 文件部分存在，但没有 `file_data` 或 `file_id`

这是构造入站请求的问题。转换器应显式跳过空文件部分，并在调用方记录“附件缺少引用”的可观察错误。不要向上游发送只有文件名的空对象。

### 直连路径正常，兼容路径异常

这通常说明目标模型并非根因。对比直连和兼容路径在模型请求之前的 JSON：内容部分类型、字段名及嵌套层级必须一致。只比较 HTTP 状态码无法定位这种语义丢失。

### 文件被错误地转成普通文本

有些适配器为了“兼容所有上游”会把未知部分序列化为字符串。这会让模型看见文件名、甚至看见一段 base64，而不是得到一个文件输入。测试应断言文件部分的类型严格等于 `input_file`；只要它出现在 `input_text` 内，就按转换失败处理。

### 日志足够定位，却泄露了附件内容

避免记录 `file_data` 的值，也不要把完整 body 作为异常上下文写入日志。大多数排障只需要内容部分类型、文件部分数量、是否存在 ID/数据、文件名是否为空等布尔量或计数。需要查看原始请求时，应在隔离测试环境、短时访问和明确授权下进行。

### 测试每次都通过，却没有覆盖新内容类型

这是“默认跳过”带来的假绿。为转换函数保留未知类型失败测试，并在新增类型时先写映射测试再写实现。这样，兼容层不会因为一个新字段被静默忽略而继续给出成功响应。

## 限制与结论

这组测试只能证明附件穿过了当前转换器，不能证明任意模型都会成功解析任意 PDF。文件大小、支持格式和模型能力仍应以调用方所用 API 的当前文档为准。

对兼容网关而言，文件输入应被视为有状态的数据契约：为每一种内容部分建立明确映射，再用“出站结构 + 哨兵语义 + 无文件对照”三项检查验证。这样即使接口返回 200，也能及时发现附件被静默丢弃的问题。若协议新增内容类型，先让测试失败，再决定是映射、拒绝还是显式降级，能把兼容性风险留在可观察的开发阶段。
