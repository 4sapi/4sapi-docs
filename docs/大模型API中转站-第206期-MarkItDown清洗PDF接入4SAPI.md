---
title: "【大模型API中转站】第206期 MarkItDown读PDF | 4SAPI模型路由省token"
category: 人工智能
tags:
  - 大模型API中转站
  - MarkItDown
  - PDF
  - Markdown
  - 4SAPI
  - 模型路由
  - 成本治理
  - 企业级大模型接入
description: "PDF 不要直接丢给 AI 硬读。先用微软开源 MarkItDown 转成 Markdown，再让模型做结构检查、分段总结、表格修复和强模型复核。本文给出 Windows 安装、批量转换、Prompt 模板，以及如何用 4SAPI 做多模型路由、日志审计和 token 成本治理。"
---

# 【大模型API中转站】第206期 MarkItDown读PDF | 4SAPI模型路由省token

本文是【大模型API中转站】系列的第206篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

很多人让 AI 读 PDF，第一步就是上传。

然后说：

```text
帮我总结。
提炼重点。
写成学习笔记。
```

这个动作看起来最省事。

但真实 PDF 一多，问题马上出来：

```text
页眉页脚混进正文。
页码被当成内容。
双栏论文顺序乱掉。
表格变成碎片。
脚注和版权声明夹在中间。
英文断词、换行、目录、参考文献全挤进上下文。
```

你以为模型在读报告。

实际模型先在处理排版垃圾。

而排版垃圾也算 token。

这就是很多 PDF 总结“字很多，但重点不准”的原因。

不是模型一定不行。

而是输入太脏。

更稳的做法是：

```text
PDF 先转 Markdown。
Markdown 先做结构检查。
再交给模型总结、提问、归档和复核。
```

这篇讲一套很实用的流程：

```text
MarkItDown 负责把 PDF 变成 Markdown。
AI 负责理解和提炼。
4SAPI 负责把不同阶段分配给不同模型，并记录成本和失败日志。
```

## 1. PDF 不是 AI 最舒服的输入

PDF 是给人看的格式。

它关心的是：

```text
页面怎么排。
字体怎么显示。
图表放在哪里。
页码怎么编号。
脚注怎么压到底部。
```

但 AI 更喜欢的是：

```text
标题层级清楚。
段落顺序稳定。
表格尽量结构化。
无关噪声少。
每段文字可引用。
```

所以不要把“AI 能读 PDF”理解成“PDF 是最佳输入”。

大模型确实可以啃 PDF。

但如果你有几十份报告、白皮书、论文、合同、说明书要处理，直接上传会让成本和错误一起上来。

先清洗，再理解。

这才是批量文档处理的基本顺序。

## 2. MarkItDown 是什么

MarkItDown 是微软开源的 Python 工具，定位是把文件和 Office 文档转换成 Markdown，方便后续做索引、文本分析和 LLM 处理。

它不只处理 PDF。

常见的 Word、PPT、Excel、HTML、CSV、JSON 等文件，也可以根据依赖支持。

但这篇先只做一个最小闭环：

```text
PDF -> Markdown -> AI 结构检查 -> AI 分段总结 -> 知识库归档
```

先把这条跑稳。

不要一上来就做企业文档中台。

## 3. 什么时候适合用 MarkItDown

适合：

```text
可复制文字的 PDF 报告。
产品说明书。
研究论文。
行业白皮书。
课程讲义。
合同初稿。
内部 SOP。
需要进入 Obsidian 或知识库的资料。
```

不适合直接用它硬扛：

```text
纯扫描 PDF。
图片里嵌文字的票据。
排版极复杂的财报表格。
大面积手写批注。
文字被特殊字体编码搞乱的老 PDF。
```

简单判断方法：

```text
打开 PDF，试着用鼠标选中正文。
能选中，大概率可以先用 MarkItDown。
选不中，先考虑 OCR 或视觉模型。
```

不要把扫描、OCR、版面恢复、语义理解全丢给一个模型。

那不是省事。

那是把所有不确定性堆在最贵的一步。

## 4. 安装 MarkItDown

MarkItDown 是 Python 工具。

先检查 Python：

```powershell
python --version
```

或者：

```powershell
py --version
```

建议单独建一个虚拟环境。

在资料目录里执行：

```powershell
py -m venv .venv
.\.venv\Scripts\Activate.ps1
```

命令行前面出现 `(.venv)`，说明环境已激活。

然后安装：

```powershell
pip install "markitdown[all]"
```

如果你想更克制，也可以只装你需要的依赖。

但新手我建议先用 `[all]`。

原因很现实：

```text
今天你处理 PDF。
明天可能就是 Word。
后天可能是 PPT。
先让工具跑起来，比一开始抠依赖更重要。
```

安装后检查：

```powershell
markitdown --help
```

如果提示命令不存在，先确认虚拟环境是否激活。

Windows 上也可以重开终端再试。

## 5. 把 PDF 转成 Markdown

假设文件叫：

```text
report.pdf
```

推荐命令：

```powershell
markitdown report.pdf -o report.md
```

也可以用重定向：

```powershell
markitdown report.pdf > report.md
```

我更推荐 `-o`。

它更直观，也更少遇到终端编码和重定向的小问题。

转换完成以后，不要马上丢给 AI。

先打开 `report.md` 看一眼。

## 6. 转完先做人眼抽查

抽查五件事：

```text
标题层级是否基本保留。
正文顺序是否正常。
页眉页脚是否大量残留。
表格是否还能看懂。
有没有明显乱码、断词、重复页码。
```

不需要追求完美。

我们的目标不是把 PDF 还原成一份漂亮文档。

目标是：

```text
让模型读到的内容，比原始 PDF 更干净、更线性、更少噪声。
```

如果 Markdown 已经能顺着读，就进入下一步。

如果一眼看过去全是乱码、页码和碎表格，那就先修输入。

不要急着调用强模型。

## 7. 第一次模型调用：先做结构质检

很多人把 Markdown 丢给模型以后，第一句话还是：

```text
总结一下。
```

我不建议。

第一步应该让模型做结构质检。

Prompt 可以这样写：

```text
我会给你一份由 PDF 转成的 Markdown。

请先不要总结正文。

请检查这份 Markdown 是否适合继续处理，输出：
1. 文档主题；
2. 主要章节；
3. 是否混入页眉、页脚、页码、目录、版权声明等噪声；
4. 哪些内容可能是表格转换残留；
5. 哪些段落顺序可能异常；
6. 建议下一步处理方式。

限制：
- 只根据我提供的 Markdown；
- 不补充外部信息；
- 不确定就写“不确定”。
```

这一步不一定要用最强模型。

它更像质检员。

可以通过 4SAPI 路由到低成本或中等模型。

真正贵的强模型，留给后面的综合判断和深度总结。

## 8. 第二次模型调用：分章节总结

结构质检通过后，再总结。

不要一口气让模型“总结整份文档”。

更稳的是按章节处理。

Prompt：

```text
请基于这一章 Markdown 做局部总结。

输出：
1. 本章一句话结论；
2. 3 到 5 个核心观点；
3. 每个观点对应的原文依据；
4. 关键数据、案例或定义；
5. 本章还需要和后文核对的问题。

限制：
- 不要编造原文没有的信息；
- 如果依据不足，请标注“依据不足”；
- 保留重要术语；
- 用中文输出。
```

重点是：

```text
每个观点都要有原文依据。
```

这条规则会强迫模型回到材料里找支撑。

否则它很容易根据标题写一篇看起来合理的读后感。

## 9. 第三次模型调用：强模型做总综合

等每章都有局部笔记，再把局部笔记交给强模型做总综合。

这个阶段才适合用 Claude、GPT、Gemini 里的强推理模型。

它要做的是：

```text
合并重复观点。
找出主线。
区分事实、作者观点和模型推断。
列出关键证据。
标注需要回 PDF 核查的位置。
生成可复用方法。
```

总控 Prompt：

```text
下面是同一份 PDF 的分章节笔记。

请做总综合：
1. 一句话结论；
2. 文档试图回答的核心问题；
3. 作者的主要论证链；
4. 关键事实和数据；
5. 作者观点与事实的区别；
6. 可以迁移到我工作的做法；
7. 需要回到原 PDF 核查的位置；
8. 适合继续追问的 5 个问题。

限制：
- 只基于分章节笔记；
- 不新增外部背景；
- 不把推断写成事实；
- 不确定就写“不确定”。
```

这样做比一次性塞全文稳。

因为每一步都有中间产物。

出错时也知道是哪一章开始错。

## 10. 4SAPI 怎么做模型分工

PDF 工作流最容易浪费钱的地方，是所有阶段都用强模型。

其实没必要。

可以这样分：

| 阶段 | 任务 | 推荐模型策略 |
| --- | --- | --- |
| 转 Markdown | MarkItDown 本地转换 | 不调用模型 |
| 结构质检 | 找页眉页脚、乱码、顺序问题 | 低成本模型 |
| 表格修复 | 把乱表整理成 Markdown 表格 | 中等模型 |
| 章节总结 | 提取观点和原文依据 | 中等或长上下文模型 |
| 总综合 | 跨章节归纳、判断、追问 | 强推理模型 |
| 扫描件 OCR | 识别图片文字 | OCR 工具或视觉模型 |
| 最终审计 | 查是否编造、是否缺依据 | 强模型或独立 reviewer |

这就是 4SAPI 的价值。

它不是只给你换一个 API 地址。

它让你可以把文档处理拆成多个模型角色：

```text
cleaner：低成本模型，做结构清洗判断。
table_fixer：中等模型，整理表格残片。
chapter_reader：长上下文模型，读章节。
synthesizer：强模型，做总综合。
auditor：独立模型，查幻觉和依据。
vision_ocr：视觉模型，处理扫描页。
```

每个角色都可以有自己的 Key、预算和日志。

这比所有任务共用一个强模型更便宜，也更可控。

## 11. 4SAPI 接入配置怎么写

如果你的脚本或工具支持 OpenAI-compatible Provider，最小配置就是三件事：

```text
Provider：4SAPI
Base URL：https://4sapi.com/v1
API Key：你的 4SAPI Key
Model：从 4SAPI 模型广场复制模型名称
```

注意，不同工具和模型可能需要不同 URL 写法。

常见有三种：

```text
https://4sapi.com
https://4sapi.com/v1
https://4sapi.com/v1/chat/completions
```

不要凭感觉填。

按 4SAPI 模型广场或技术文档里的端点来。

脚本里可以用环境变量：

```powershell
$env:MODEL_GATEWAY_BASE_URL="https://4sapi.com/v1"
$env:MODEL_GATEWAY_API_KEY="sk-xxxxxxxx"
$env:PDF_QA_MODEL="your-low-cost-model"
$env:PDF_READER_MODEL="your-long-context-model"
$env:PDF_SYNTH_MODEL="your-strong-model"
```

模型名不要手打。

从后台复制。

少一个字符，后面会排查很久。

## 12. 批量转换 PDF

一个文件夹里有很多 PDF，可以用 PowerShell：

```powershell
Get-ChildItem -Filter *.pdf | ForEach-Object {
  $out = [System.IO.Path]::ChangeExtension($_.FullName, ".md")
  markitdown $_.FullName -o $out
}
```

如果想统一输出到 `markdown` 文件夹：

```powershell
New-Item -ItemType Directory -Force -Path ".\markdown" | Out-Null

Get-ChildItem -Filter *.pdf | ForEach-Object {
  $name = [System.IO.Path]::GetFileNameWithoutExtension($_.Name)
  markitdown $_.FullName -o ".\markdown\$name.md"
}
```

批量转换以后，不要立刻批量总结。

先抽查 2 到 3 份。

确认转换质量稳定，再进入模型流水线。

否则你会把一堆坏 Markdown 批量喂给模型。

那只是把浪费自动化了。

## 13. 扫描版 PDF 怎么办

如果转换出来几乎没有正文，大概率是扫描版。

这种 PDF 本质上是一组图片。

处理顺序要变：

```text
PDF 页面 -> 图片
图片 -> OCR / 视觉模型识别
识别文本 -> Markdown
Markdown -> 结构质检
Markdown -> 总结和归档
```

这里可以用两种策略：

```text
低成本路线：本地 OCR 工具先识别，再交给模型清洗。
高准确路线：通过 4SAPI 调视觉模型处理关键页。
```

不要整本扫描 PDF 都交给视觉模型硬读。

先判断哪些页有价值。

例如：

```text
目录页。
摘要页。
章节首页。
包含关键表格的页面。
图表说明页。
```

4SAPI 侧建议给扫描件单独建预算桶：

```text
cost_bucket: pdf_ocr
max_pages_per_file: 20
review_required: true
```

扫描 PDF 的成本更容易失控。

必须单独治理。

## 14. 表格乱了怎么处理

PDF 表格是最容易翻车的地方。

如果表格不是核心信息，可以在总结里标注：

```text
表格转换质量较差，未纳入结论。
```

如果表格很重要，就单独处理。

Prompt：

```text
下面这段 Markdown 可能来自 PDF 表格转换，格式不稳定。

请只基于现有内容，整理成 Markdown 表格。

要求：
1. 不补充新数据；
2. 不合并无法确认的单元格；
3. 无法判断的位置填“不确定”；
4. 保留原始字段名；
5. 输出后列出你做了哪些整理。
```

表格修复适合用中等模型。

强模型不一定更划算。

表格本质是结构整理，不是深度思考。

## 15. 文件太大怎么办

不要整份塞。

分段。

推荐策略：

```text
先按 Markdown 标题切分。
每一节单独总结。
每节输出局部结论和待核查问题。
最后用局部笔记做总综合。
```

如果没有标题，就按字符数切。

但要保留边界信息：

```text
part_index: 3
source_file: report.md
section_title: 不确定
start_hint:
end_hint:
```

这样后面回查原文时不会迷路。

4SAPI 日志里也建议记录：

```text
document_id
chunk_id
model
role
tokens_in
tokens_out
status
error_type
```

大文档处理最怕只看到总账。

你需要知道是哪一段花钱、哪一段失败。

## 16. 一个完整企业工作流

如果你在企业里处理报告、合同、白皮书，可以把流程做成这样：

```text
1. 文件进入 docs_inbox/
2. MarkItDown 转 Markdown
3. 脚本抽查转换质量
4. 低成本模型做结构质检
5. 按章节切分
6. 中等模型做章节笔记
7. 强模型做总综合
8. 独立 reviewer 查依据和幻觉
9. 写入知识库或报告系统
10. 4SAPI 记录模型、Key、成本和错误
```

企业级场景里，至少要记录这些字段：

```text
document_id
file_name
file_type
department
workflow_id
stage
model
request_id
key_group
cost_bucket
status
error_type
review_required
output_path
```

这样你才能回答：

```text
哪个部门处理文档最多？
哪类 PDF 转换失败率最高？
哪一步最烧 token？
哪些文件进入了人工复核？
哪组 Key 快超预算？
```

这就是企业 API 网关的意义。

不只是调用模型。

而是让模型调用可查、可控、可审计。

## 17. 可以复制的总控 Prompt

```text
你是 PDF Markdown 可靠阅读助手。

我会给你一份由 PDF 转成的 Markdown。

你的目标不是直接写漂亮总结，而是做可核查阅读。

请按顺序完成：
1. 判断文档主题和材料类型；
2. 识别目录、页眉、页脚、页码、版权声明、表格残留等噪声；
3. 按章节提取核心观点；
4. 为每个核心观点给出原文依据；
5. 标注哪些内容是事实、哪些是作者观点、哪些是你的推断；
6. 列出需要回到原 PDF 核查的位置；
7. 给出适合进入知识库的 Markdown 笔记版本。

限制：
- 只使用我提供的 Markdown；
- 不补充外部背景；
- 不编造数据、作者、文献或案例；
- 不确定就写“不确定”；
- 输出要方便后续进入 Obsidian 或企业知识库。
```

如果通过 4SAPI 调用，建议把这个 Prompt 对应的 `task_type` 标成：

```text
pdf_reliable_reading
```

不要所有文档任务都叫：

```text
chat
```

否则后面根本看不出成本花在哪里。

## 18. 最后总结

AI 读资料，很多时候不是模型越强越好。

而是输入越干净越好。

PDF 是排版格式。

Markdown 才更像模型友好的工作底稿。

MarkItDown 的价值不是炫技。

它是一道喂 AI 之前的清洗工序。

4SAPI 的价值也不是多一层配置。

它是一层模型治理：

```text
低成本模型做结构质检。
中等模型做章节处理。
强模型做总综合。
视觉模型处理扫描页。
独立 reviewer 做幻觉审计。
所有调用都留下日志和成本记录。
```

一句话：

```text
先把 PDF 变成干净 Markdown，
再让模型做它真正擅长的理解、判断和提炼。
```

这套流程跑通以后，PDF 总结就不再是“上传碰运气”。

而是一条可复用、可批量、可审计的企业级文档处理流水线。

## 资料与延伸阅读

- Microsoft MarkItDown：https://github.com/microsoft/markitdown
- MarkItDown PyPI：https://pypi.org/project/markitdown/
- 4SAPI 官网：https://4sapi.com/
- 4SAPI 接入文档：https://4sapi.apifox.cn/
- 4SAPI 接入实操手册：https://4sapi.com/blog/4sapi-api-integration-setup-guide
