---
title: "长文写作如何用检索账本连接资料、提纲与初稿"
tags:
  - 研究写作
  - NotebookLM
  - 资料整理
description: "长文写作最容易失控的环节是资料、提纲和正文之间缺少可追踪关系。"
---
# 长文写作如何用检索账本连接资料、提纲与初稿
长文写作最容易失控的环节是资料、提纲和正文之间缺少可追踪关系。本文用 NotebookLM 与 Content Research Writer 设计一条研究写作流程：先建立来源账本，再提炼问题和提纲，最后生成带证据位置的初稿，并明确哪些内容必须回到原文核验。文中只讨论可复现的步骤，不把单次结果扩展成产品承诺；每个结论都标注前提、证据和无法覆盖的边界。读者可以先完成最小验证，再按自己的版本、权限和数据补充实验。


写深度文章最容易出问题的地方，通常不是文笔。

是你看了三篇资料，模型却把它们压成一句“行业正在快速发展”；你给了一个 PDF，它写出一篇引用看起来很完整、但原文并没有说过的文章；你让它“基于这些材料写稿”，最后得到的只是搜索结果的重新排列。

这篇把两个项目放在同一条工作流里：

~~~text
notebooklm-skill：收资料、建 Notebook、带引用提问、做研究和生成研究产物
content-research-writer：定目标、列提纲、补研究、改 Hook、逐段反馈和最终润色
~~~

前者更像资料研究员，后者更像写作编辑。

它们不能替代作者判断，也不能保证文章自动正确，但能把“先找到证据，再组织表达”这件事固定下来。

## 一、为什么让研究和写作分开

一个模型同时负责检索、判断、组织和写作，最容易发生三种混淆：

### 资料和结论混在一起

模型在检索时看到一个说法，写作时可能把它当作已经证实的结论。如果没有保留来源和原句，很难回头判断它是事实、推论还是模型补全。

### 文章结构先于证据

模型先按熟悉的模板列出“背景、现状、挑战、未来”，再去找内容填空。这样的文章很顺，却不一定回答了真实问题。

### 引用变成装饰

一篇文章出现了很多链接，不代表这些链接真的支持段落里的数字和结论。引用必须能回到原始资料，最好还能对应具体页码、段落或来源条目。

所以更好的分工是：

~~~text
NotebookLM：哪些资料支持这个判断？不同来源哪里冲突？
写作 Skill：读者是谁？文章承诺什么？怎样组织成可读的结构？
作者：这个结论是否成立？我是否愿意署名？
~~~

## 二、notebooklm-skill 是什么

项目地址：[claude-world/notebooklm-skill](https://github.com/claude-world/notebooklm-skill)

notebooklm-skill 是一个把 Google NotebookLM 接到终端和 AI Agent 的开源项目。仓库 README 当前说明，它基于 notebooklm-py，提供：

- JSON 优先的核心 CLI。
- 多种端到端研究流水线。
- 13 个工具的 FastMCP Server。
- 认证配置和多 Profile 支持。
- CLI、研究流水线和 MCP 之间共用的兼容层。

它的核心能力不是“让 Claude 直接读你的所有资料”，而是先把资料放进 NotebookLM，再针对这些资料提问，尽量让答案保留来源信息。

可处理的输入包括：

- 网页 URL。
- 本地 PDF、DOCX 和其他文件。
- 原始文本。
- 多个混合来源。
- 后续研究得到的网页资料。

可生成的产物也不止文章，包括报告、学习指南、测验、闪卡、思维导图、信息图、数据表、幻灯片、音频和视频等。具体类型和选项以当前命令的 help 输出为准。

仓库同时明确提示：这是对 NotebookLM Web API 的非官方集成。Google 侧的可用性、配额、生成时间和产物行为都可能改变。

这句话必须写进使用预期里。

## 三、安装 NotebookLM Skill

仓库 README 推荐的隔离安装方式是：

~~~text
git clone https://github.com/claude-world/notebooklm-skill.git
cd notebooklm-skill
./install.sh
notebooklm-auth setup
notebooklm-skill list
~~~

如果你更习惯 Python 虚拟环境，也可以按照 README 使用 PyPI 或 uvx。示例：

~~~text
python3 -m venv .venv
source .venv/bin/activate
python -m pip install notebooklm-skill
python -m playwright install chromium
notebooklm-auth setup
~~~

在 Windows 上，虚拟环境激活命令和脚本调用方式会不同。不要机械复制 Linux 的 source 命令，优先使用项目当前安装说明或直接运行：

~~~text
uvx --from notebooklm-skill notebooklm-auth setup
uvx --from notebooklm-skill notebooklm-skill list
~~~

如果你明确要使用本机 Chrome，仓库提供了：

~~~text
notebooklm-auth setup --browser chrome --fresh
~~~

首次验证先执行：

~~~text
notebooklm-auth verify
~~~

认证信息要按 Profile 管理。仓库文档提到会使用类似 storage_state.json 的会话文件，这类文件不能读取、打印、复制或提交到 Git。

## 四、创建一个带来源的 Notebook

可以从混合来源开始：

~~~text
notebooklm-skill create \
  --title "AI 内容生产研究" \
  --sources https://example.com/article \
  --files ./paper.pdf \
  --text-sources "我的访谈记录" \
  --strict
~~~

这里的 strict 不是“让答案绝对正确”，而是让来源处理结果更诚实地暴露出来。资料可能出现成功、部分成功或失败，不能因为 Notebook 建立成功，就假设每个来源都已经完整进入。

创建后检查：

~~~text
notebooklm-skill list
notebooklm-skill list-sources --notebook "AI 内容生产研究"
~~~

自动化任务更建议使用 Notebook ID，而不是标题。标题可能重复或发生变化，ID 更适合脚本和重复调用。

## 五、先问资料，再写结论

### 1. 先做摘要和冲突检查

~~~text
notebooklm-skill summarize --notebook "AI 内容生产研究"

notebooklm-skill ask \
  --notebook "AI 内容生产研究" \
  --query "这些资料对 AI 内容生产的共同判断是什么？哪些地方互相矛盾？请保留引用信息。"
~~~

问题不要只问“总结一下”。更有效的研究问题是：

- 哪些结论被两个以上独立来源支持。
- 哪些数字只出现在单一来源里。
- 哪些观点是作者判断，而不是研究结果。
- 资料之间的冲突发生在定义、样本、时间还是方法。
- 哪个问题在资料库里没有证据。

### 2. 做快速或深度研究

仓库支持通过 NotebookLM 做研究并导入结果：

~~~text
notebooklm-skill research \
  --notebook "AI 内容生产研究" \
  --query "最近关于 AI 写作质量评估的独立研究" \
  --mode deep \
  --max-sources 10
~~~

如果你只想拿到任务 ID，不想等待长任务结束，可以使用 no-wait。研究结果是否导入 Notebook，也有相应选项控制。请先查看：

~~~text
notebooklm-skill research --help
~~~

研究结束后，不要马上让 Claude 写文章。先检查新增来源：

~~~text
notebooklm-skill list-sources --notebook "AI 内容生产研究"
~~~

确认哪些结果被导入、哪些失败、哪些只是搜索摘要，再进入写作阶段。

## 六、content-research-writer 是什么

这个 Skill 位于 OpenAkita 仓库的：

[openakita/openakita/skills/content-research-writer/SKILL.md](https://github.com/openakita/openakita/tree/main/skills/content-research-writer)

它的定位不是“自动写一篇漂亮文章”，而是作为写作伙伴，帮助你：

- 协作列出文章提纲。
- 为关键论点补研究和引用。
- 设计更有吸引力的开头。
- 在每个段落写完后给反馈。
- 保留作者的语气和写作选择。
- 管理文中引用和参考文献。
- 经过多轮修改后再做全文润色。

文档给出的基础流程很清楚：

~~~text
创建独立写作目录
  -> 建立 article-draft.md
  -> 一起列提纲
  -> 研究关键论点并加引用
  -> 优化 Hook
  -> 每写一节就反馈
  -> 全文检查结构、证据和可读性
~~~

它还要求先弄清楚五件事：

1. 主题和主论点是什么。
2. 目标读者是谁。
3. 文章长度和格式是什么。
4. 目标是教育、说服、娱乐还是解释。
5. 已经有哪些资料和来源。

这五件事没有答案时，模型会用自己熟悉的通用文章替你补齐。文章看似完成了，作者却不一定认领。

## 七、让 NotebookLM 和写作 Skill 接力

### 第一步，建立项目目录

把研究和稿件放到同一个项目，但分成不同文件：

~~~text
ai-content-research/
├─ sources.md
├─ research-notes.md
├─ outline.md
├─ article-draft-v1.md
├─ article-draft-v2.md
└─ references.md
~~~

sources.md 只记录来源；research-notes.md 记录问题、回答、引用和未解决冲突；outline.md 只负责文章结构。

不要把 NotebookLM 的整段输出直接当成最终稿。先把“事实”和“写法”分开。

### 第二步，向 NotebookLM 提出研究问题

~~~text
请只基于当前 Notebook 中已经成功导入的来源回答：

1. 关于 [核心问题]，资料有哪些共同结论？
2. 哪些结论的证据更充分？请说明来源和理由。
3. 哪些数字、时间、样本和因果关系需要特别核验？
4. 哪些来源之间互相矛盾？矛盾可能来自什么？
5. 当前资料没有回答、但会影响文章结论的关键问题是什么？

输出时保留引用信息。
如果资料里没有答案，请明确写“当前来源没有足够证据”，不要用常识补全。
~~~

把结果整理进 research-notes.md，并给每条结论加状态：

~~~text
confirmed：多个来源支持，原文位置清楚
single-source：只有一个来源，等待核验
inference：根据资料推断，不能写成原文结论
missing：资料没有回答
~~~

### 第三步，交给 content-research-writer 列提纲

~~~text
我准备写一篇文章。

主题：[主题]
主论点：[我真正想表达的判断]
目标读者：[读者]
目标长度：[长度]
目标形式：[公众号 / Newsletter / 技术教程 / 研究简报]

请读取 research-notes.md 和 sources.md，先不要写正文。
请输出：
1. 一个具体、不过度承诺的标题。
2. 一个能让目标读者继续读下去的开头方向。
3. 文章提纲，每节写清楚主张、证据、反方观点和需要补的资料。
4. 哪些地方只能写成推论，哪些地方需要原始来源。
5. 文章结尾要落到什么行动或开放问题。

保留我的主论点，不要用模板化的“背景—现状—挑战—展望”强行填满结构。
~~~

### 第四步，按段落写和审

不要让 Agent 一次输出一万字。写完开头后，先审开头；写完第一节，再审第一节。

~~~text
我刚完成“[章节名]”这一节。
请从以下角度反馈：
- 读者是否能看懂这节要回答的问题。
- 核心结论是否有对应来源。
- 哪一句是推测，却被写成了事实。
- 哪个例子太泛，需要换成具体证据。
- 与上一节的衔接是否自然。

请给修改建议和两种句子改写，不要直接覆盖原稿。
~~~

content-research-writer 的重要原则是“建议，而不是替换”。作者喜欢原句时，Agent 应该支持作者的选择，而不是为了所谓的风格统一强行改掉。

## 八、研究到文章的高阶流水线

如果你想让 notebooklm-skill 先产出一份文章草稿，可以使用仓库里的高层流水线：

~~~text
notebooklm-pipeline research-to-article \
  --sources https://example.com/a https://example.com/b \
  --title "Evidence review" \
  --language zh-TW \
  --audience engineers
~~~

这个流水线适合先得到一个有来源的研究草稿，但它仍然不是最终发布稿。把产物交给 content-research-writer，继续做：

- 重新确认主论点。
- 删除没有证据的段落。
- 把多个来源冲突处写清楚。
- 调整开头和文章节奏。
- 加入作者真正的经历和判断。
- 保留引用列表。

如果要做社交内容，也有 research-to-social；如果要处理 RSS，可以用 batch-digest；如果需要多个 NotebookLM 产物，可以使用 generate-all。不同流水线会随着仓库版本变化，正式使用前先运行：

~~~text
notebooklm-pipeline --help
~~~

## 九、生成产物不等于发布

notebooklm-skill 可以生成报告、文章、幻灯片、音频、视频、测验、闪卡和信息图，也可以下载到本地。

但仓库文档明确写了：流水线返回草稿和本地产物，不会替你发布到社交平台、CMS 或其他远程目的地。

这其实是好事。研究和发布是两个责任不同的阶段：

~~~text
研究产物：可以继续修改、复核、标注来源
发布内容：需要确认事实、版权、隐私、品牌和平台格式
~~~

不要把“生成了 PPTX”写成“已经发出去了”，也不要把“研究任务完成”写成“结论已经被证明”。

## 十、MCP 接入和安全边界

notebooklm-skill 提供 MCP Server，默认 stdio 模式适合 Claude Code、Cursor 等 MCP 客户端。README 当前列出 13 个工具，涵盖 Notebook 管理、来源、提问、摘要、研究和产物生成。

如果使用 HTTP 模式，仓库要求绑定本机回环地址：

~~~text
notebooklm-mcp --http --host 127.0.0.1 --port 8765
~~~

不要把这个服务直接暴露到公网。如果确实要远程使用，应当放在经过认证的 TLS 代理后面，并设置主机级访问控制。

还要注意：

- NotebookLM 使用的是浏览器会话和非官方 Web API。
- Google 侧可能改变登录、配额和页面行为。
- 本地会话文件不能提交到 Git。
- 私有 PDF、客户资料和内部研究不要默认上传到第三方服务。
- 来源导入可能部分失败，必须向读者诚实报告。
- 已有输出文件默认不会被覆盖，强制覆盖要有明确意图。
- 删除 Notebook 需要显式确认。

## 十一、企业级大模型接入怎么做

企业团队可以把这条链路拆成四层：

### 资料层

规定哪些网页、PDF、内部文档可以进入 NotebookLM，哪些只能在本地处理。对客户资料和含个人信息的文件先脱敏。

### 研究层

给每个项目、部门或课题建立独立 Notebook，避免把不同权限的资料混在一个知识库里。保存来源清单和研究时间。

### 模型层

写作模型通过统一企业 API 入口或企业 API 网关接入，按项目分配 API Key。不要让每个员工的本地脚本各自保存生产 Key。

### 审计层

记录 Notebook、来源类型、调用时间、模型、耗时、失败原因和成本。原始正文、Token 和授权文件不写进普通日志。

通过 上游 API 或其他多模型 API 中转层接入时，先确认它是否支持团队、项目和环境维度的 Key 权限、调用追踪、预算和失败告警。不要把中转服务当成来源事实核验工具。

## 十二、完整验收清单

### 来源验收

- [ ] 每个来源都有导入结果。
- [ ] 失败和部分成功的来源被明确标记。
- [ ] 研究问题能回到具体来源。
- [ ] 数字、时间、样本和引用逐项核对。

### 写作验收

- [ ] 文章主论点不是模型自行补出的。
- [ ] 每个关键结论都有证据或明确标为推论。
- [ ] 开头提出的问题在结尾得到回应。
- [ ] 作者的经历和判断由作者提供。
- [ ] 引用格式统一，链接可访问。

### 安全验收

- [ ] 未把会话文件提交到 Git。
- [ ] 未将本地 HTTP MCP 暴露到公网。
- [ ] 未把敏感资料发送给未授权的外部模型。
- [ ] 未把 API Key、Cookie 和 Token 放进 Prompt 或日志。
- [ ] 研究输出和最终发布稿分开保存。

## 总结

NotebookLM 适合做“资料里究竟说了什么”，content-research-writer 适合做“怎样把可靠材料组织成一篇读者看得懂的文章”。

最稳妥的顺序是：

~~~text
建立来源库
  -> 检查导入结果
  -> 提出带引用的问题
  -> 保存研究笔记和冲突
  -> 协作列提纲
  -> 分段写作和反馈
  -> 引用、事实、隐私和版权终审
~~~

有引用不等于绝对正确，有研究流水线也不等于可以跳过作者责任。

## 结论

本文给出了问题定位、配置或创作流程的可执行路径。实际结果仍取决于当前版本、权限和运行环境，提交前应按官方文档复核可变字段，并保留失败证据和回滚边界。
