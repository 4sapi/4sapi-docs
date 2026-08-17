---
title: "【大模型API中转站】第56期 企业微信CLI文档 | Agent落地"
category: 人工智能
tags:
  - 大模型API中转站
  - 企业微信
  - wecom-cli
  - 企业微信文档
  - AI Agent
  - 4SAPI
description: "基于企业微信官方 wecom-cli，拆解 API模式智能机器人、Bot ID 和 Secret、CLI 初始化、文档创建读取写入、智能文档边界，以及如何用 4SAPI 统一 Agent 的模型调用和成本审计。"
---

# 【大模型API中转站】第56期 企业微信CLI文档 | Agent落地

本文是【大模型API中转站】系列的第56篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一篇我们讲 MemPalace，把重点放在 Agent 的本地长期记忆上。

这一篇开始看企业微信 CLI。

企业微信 CLI 这个选题可以写，但第一篇不建议一上来就讲消息、待办、会议、日程全部能力。

更稳的入口是：

```text
先让 Agent 能创建、读取、写入企业微信文档。
```

原因很简单。

文档是企业办公里最容易验证、最容易回滚、最不容易误触发外部动作的对象。

消息涉及隐私，发消息涉及外部影响，会议和待办涉及团队协作，智能表格又涉及结构化数据和字段类型。

所以第 56 期先做“文档 Agent”：

```text
安装 wecom-cli
获取 Bot ID 和 Secret
初始化企业微信机器人
创建一篇文档
把 Agent 生成的 Markdown 写进去
读取文档内容再交给模型分析
用 4SAPI 管模型、Key、成本和日志
```

这条链路跑通后，下一篇再讲消息、智能表格和执行闭环。

## 1. 为什么企业微信 CLI 值得写

很多企业已经把工作资料放在企业微信里：

```text
项目周报
需求文档
会议纪要
客户材料
运营 SOP
产品排期
智能表格台账
```

但 AI Agent 常见的使用方式还是很手动：

```text
人打开企业微信文档
复制内容给 AI
AI 总结或改写
人再复制回企业微信文档
```

这时人就变成了“Agent 和办公系统之间的剪贴板”。

企业微信 CLI 的价值在这里：

```text
让人类和 AI Agent 都能在终端中操作企业微信。
```

截至 2026-06-18，我查到的公开信息是：

| 项目 | 信息 |
| --- | --- |
| GitHub 仓库 | `WecomTeam/wecom-cli` |
| NPM 包 | `@wecom/cli` |
| 当前 NPM 最新版 | `0.1.9` |
| 最新版发布时间 | 2026-05-27 |
| 支持平台 | macOS、Linux、Windows |
| Node 要求 | `>=18` |
| 许可证 | MIT |

我在 Windows 本地验证：

```bash
npx --yes @wecom/cli --version
```

返回：

```text
wecom-cli 0.1.9
```

未初始化时执行：

```bash
npx --yes @wecom/cli auth show --auth-status
```

返回：

```text
unauthorized
```

这说明工具入口正常，只是还没有配置企业微信机器人凭证。

## 2. 先分清它和 4SAPI 的关系

企业微信 CLI 不是模型网关。

它不负责 Claude、GPT、GLM、Qwen、DeepSeek 这些模型调用，也不负责 Token 计费。

它负责的是企业微信办公能力：

```text
通讯录
文档
智能表格
消息
待办
会议
日程
```

4SAPI 的位置在模型层。

更准确的架构是：

```text
企业微信文档
  -> wecom-cli 读取或写入
  -> Codex / Claude Code / OpenClaw 等 Agent 编排任务
  -> 4SAPI 统一调用模型
  -> Claude / GPT / GLM / Qwen 等模型生成结果
  -> wecom-cli 写回企业微信文档
```

所以这篇不是“把 wecom-cli 接入 4SAPI”。

而是：

```text
把 Agent 的模型调用统一走 4SAPI。
把企业微信 CLI 当成 Agent 的办公工具层。
```

这样写才不会混淆。

4SAPI 管：

```text
模型路由
Key 管理
成本统计
日志审计
失败排查
```

wecom-cli 管：

```text
企业微信文档的创建、读取和写入
```

Agent 管：

```text
理解用户需求
决定什么时候读文档
什么时候调用模型
什么时候写回文档
什么时候向用户确认
```

## 3. 它不是群机器人 Webhook

很多人一听企业微信机器人，会想到群机器人 Webhook：

```text
拿一个 webhook 地址
curl 发一段 markdown
群里收到消息
```

wecom-cli 不是这个定位。

它基于企业微信 API 模式智能机器人，使用 Bot ID 和 Secret 初始化。

这意味着它的重点不是“往群里推一条通知”，而是让 CLI 和 Agent 能在授权范围内调用办公能力。

对于文档场景，核心价值是：

```text
Agent 可以创建企业微信文档。
Agent 可以读取企业微信文档。
Agent 可以把 Markdown 写入企业微信文档。
Agent 可以把会议纪要、周报、复盘、任务拆解沉淀到企业微信。
```

这和简单 Webhook 推送是两类东西。

## 4. 合规边界先写在前面

这篇只讨论官方开放能力，不讨论任何绕过权限或模拟客户端的做法。

不要写成：

```text
自动登录企业微信客户端
抓取客户端消息数据库
模拟点击复制文档
绕过可见范围读取资料
```

正确边界是：

```text
API 模式智能机器人
官方 wecom-cli
机器人可见范围
用户授权
本地凭证加密存储
敏感内容按企业规则处理
```

尤其是接 4SAPI 时，要把数据边界讲清楚。

如果企业微信文档里包含客户信息、合同、内部财务、人事资料，就要先确认：

```text
这类内容能不能发给外部模型？
是否需要脱敏？
是否只允许本地模型处理？
是否需要审计调用日志？
```

大模型API中转站能帮你统一入口、日志和成本，但不能替你决定哪些数据可以出域。

## 5. 安装 wecom-cli

环境要求：

```text
Node.js >= 18
企业微信账号
macOS / Linux / Windows
可选：API模式智能机器人 Bot ID 和 Secret
```

安装：

```bash
npm install -g @wecom/cli
```

验证版本：

```bash
wecom-cli --version
```

查看帮助：

```bash
wecom-cli --help
```

帮助信息里会看到这些品类：

```text
contact   通讯录
doc       文档 / 智能表格
meeting   会议
msg       消息
schedule  日程
todo      待办事项
cache     本地缓存
```

如果只写文档 Agent，第一篇重点看 `doc` 和 `contact` 就够了。

`contact` 用来把姓名转换成 userid。

`doc` 用来创建、读取、写入文档。

## 6. 安装 Agent Skills

如果你的 Agent 工具支持 Skills，可以安装官方 Skills：

```bash
npx skills add WeComTeam/wecom-cli -y -g
```

这一步不是普通人手敲命令必须的，但对 Agent 很有价值。

因为 Skills 会告诉 Agent：

```text
什么时候该调用哪个企业微信命令
参数怎么填
文档 URL 怎么识别
智能表格和普通表格怎么区分
哪些操作需要用户确认
错误码怎么处理
```

仓库里能看到这些 Skill：

| Skill | 作用 |
| --- | --- |
| `wecomcli-doc` | 文档、表格、智能文档读写 |
| `wecomcli-smartsheet` | 智能表格结构和记录管理 |
| `wecomcli-contact` | 通讯录可见范围成员查询 |
| `wecomcli-msg` | 会话、消息、媒体文件、文本发送 |
| `wecomcli-todo` | 待办查询、创建、更新、删除 |
| `wecomcli-meeting` | 会议创建、查询、取消、成员管理 |
| `wecomcli-schedule` | 日程查询、创建、更新、闲忙查询 |

第 56 期只用前三个概念：

```text
doc：文档读写
smartsheet：先知道它和普通文档不同
contact：涉及人员字段时不能猜 userid
```

## 7. 获取 Bot ID 和 Secret

企业微信帮助中心有一篇官方文档，标题是：

```text
如何获取 Bot ID 和 Secret
```

步骤可以概括为：

```text
打开企业微信客户端
-> 工作台
-> 智能机器人
-> 创建机器人
-> 手动创建
-> 选择 API模式创建
-> 连接方式选择“使用长连接”
-> 生成 Bot ID 和 Secret
-> 配置机器人可见范围
-> 保存
```

这里有两个重要点。

第一，Secret 要当成 API Key 管。

不要截图发群，不要写进文章，不要提交到 Git。

第二，可见范围决定 CLI 能看到什么。

机器人不可见的成员、文档、会话，CLI 也不应该能越权访问。

如果后面 Agent 读不到某个文档，不要第一反应就怀疑模型或 4SAPI。

先检查：

```text
机器人是否有可见范围
当前用户是否有文档权限
docid 或 URL 是否正确
CLI 是否初始化成功
动态工具列表是否拉取成功
```

## 8. 初始化企业微信 CLI

交互式初始化：

```bash
wecom-cli init
```

当前 `@wecom/cli 0.1.9` 支持两种交互方式：

```text
扫码接入
手动输入 Bot ID 和 Secret
```

非交互参数当前主要用于扫码接入：

```bash
wecom-cli init --noninteractive
```

如果不想自动打开二维码页面：

```bash
wecom-cli init --noninteractive --no-open
```

初始化后检查状态：

```bash
wecom-cli auth show --auth-status
```

成功时应该返回：

```text
authorized
```

注意：截至 `@wecom/cli 0.1.9`，我从源码和公开 issue 里看到，`init --bootstrap` 或通过环境变量直接注入 Bot ID / Secret 仍是讨论方向，不是已发布正式能力。

所以文章不要写成：

```bash
wecom-cli init --bootstrap
```

更稳妥的写法是：

```text
当前发布版以交互式初始化为主。
CI、容器、云端沙箱中的免交互初始化仍是后续值得关注的方向。
```

## 9. 创建第一篇企业微信文档

创建普通文档：

```bash
wecom-cli doc create_doc '{"doc_type": 3, "doc_name": "AI项目周报"}'
```

成功后返回里会包含：

```text
docid
url
```

这两个信息要保存。

`url` 给人看。

`docid` 给后续 CLI 调用。

如果你要让 Agent 自动创建项目文档，建议让它在本地或知识库里维护一个映射：

| 项目 | 文档类型 | docid | url |
| --- | --- | --- | --- |
| A 项目 | 周报 | `DOCID` | `https://doc.weixin.qq.com/doc/...` |
| B 项目 | 会议纪要 | `DOCID` | `https://doc.weixin.qq.com/doc/...` |

不要每次都重新创建一篇新文档。

更好的方式是：

```text
第一次创建
后续读取原文
在原文基础上追加或覆写
必要时新建归档文档
```

## 10. 写入 Markdown 内容

用 Markdown 覆写文档：

```bash
wecom-cli doc edit_doc_content '{"docid": "DOCID", "content": "# AI项目周报\n\n## 本周完成\n\n- 完成企业微信 CLI 调研\n- 跑通文档创建和写入\n\n## 下周计划\n\n- 接入消息总结\n- 设计智能表格任务台账", "content_type": 1}'
```

这里 `content_type: 1` 表示 Markdown。

这也是企业微信 CLI 对 Agent 友好的地方。

Agent 本来就擅长生成 Markdown：

```text
标题
目录
表格
列表
行动项
风险清单
下一步计划
```

所以第一条实战链路可以设计成：

```text
用户：帮我生成一份项目周报
-> Agent 调 4SAPI 生成 Markdown
-> wecom-cli 创建企业微信文档
-> wecom-cli 写入 Markdown
-> 返回文档 URL 给用户
```

这样读者马上能看到结果。

## 11. 读取文档内容

读取文档：

```bash
wecom-cli doc get_doc_content '{"docid": "DOCID", "type": 2}'
```

也可以通过 URL：

```bash
wecom-cli doc get_doc_content '{"url": "https://doc.weixin.qq.com/doc/xxx", "type": 2}'
```

读取接口采用异步轮询机制。

首次调用可能返回 `task_id`，如果 `task_done` 是 `false`，就要携带 `task_id` 再调用，直到完成。

这点对 Agent 很重要。

不要让 Agent 调一次没拿到正文就说“文档为空”。

正确逻辑是：

```text
第一次调用 get_doc_content
-> 如果 task_done=false，保存 task_id
-> 等待后继续轮询
-> task_done=true 后再把 content 交给模型
```

拿到内容后，就可以接 4SAPI 做分析：

```text
读取需求文档
-> 调 4SAPI 总结需求
-> 输出风险点
-> 生成评审问题
-> 写回一篇评审纪要模板
```

## 12. 普通文档、表格、智能表格、智能文档别混用

这是企业微信 CLI 文档场景里最容易踩的坑。

企业微信里有多种文档对象：

| URL 模式 | 品类 | 推荐接口 |
| --- | --- | --- |
| `/doc/*` | 文档 | `get_doc_content` |
| `/sheet/*` | 普通在线表格 | `get_doc_content` |
| `/smartsheet/*` | 智能表格 | `smartsheet_*` 系列 |
| `/smartpage/*` | 智能文档 / 智能主页 | `smartpage_export_task` / `smartpage_get_export_result` |

不要把 `/smartsheet/` 当成普通 `/sheet/`。

不要在用户只是说“帮我建个文档”时默认创建智能文档。

不要把智能文档接口泛化到普通文档。

这也是为什么官方 Skills 里会把 URL 识别规则写得很细。

人能看页面猜类型。

Agent 必须按规则判断。

## 13. 智能文档什么时候用

普通文档适合：

```text
周报
会议纪要
复盘
方案草稿
需求说明
客户沟通总结
```

智能文档更适合：

```text
多页面知识库
结构化项目主页
由多个 Markdown 文件生成的资料包
```

创建智能文档时，命令要用带 `+` 的 helper：

```bash
wecom-cli doc +smartpage_create '{"title": "项目概览", "pages": [{"page_title": "需求文档", "content_type": 1, "page_filepath": "/path/to/requirements.md"}]}'
```

注意几个限制：

```text
只有用户明确说智能文档或智能主页时才用 smartpage。
Markdown 文件要传 content_type=1。
每个子页面 Markdown 文件大小不要超过 10MB。
创建成功后保存 docid 和 url。
```

第一篇教程不建议把智能文档作为主线。

它更适合作为进阶章节。

## 14. 文档 Agent 的三种实战流

### 14.1 AI 周报生成器

适合个人和小团队。

流程：

```text
用户输入本周要点
-> Agent 调 4SAPI 整理成周报
-> wecom-cli 创建企业微信文档
-> wecom-cli 写入 Markdown
-> 返回文档链接
```

命令骨架：

```bash
wecom-cli doc create_doc '{"doc_type": 3, "doc_name": "第25周项目周报"}'
wecom-cli doc edit_doc_content '{"docid": "DOCID", "content": "# 第25周项目周报\n\n...", "content_type": 1}'
```

这个流的好处是低风险。

它不读取聊天记录，不自动发消息，不改别人日程。

很适合第一篇文章用来演示。

### 14.2 需求文档评审助手

适合研发和产品团队。

流程：

```text
用户给一个企业微信文档 URL
-> wecom-cli 读取文档内容
-> Agent 调 4SAPI 检查需求完整性
-> 输出缺失项、歧义点、风险点
-> 创建一篇评审问题清单
```

命令骨架：

```bash
wecom-cli doc get_doc_content '{"url": "https://doc.weixin.qq.com/doc/xxx", "type": 2}'
wecom-cli doc create_doc '{"doc_type": 3, "doc_name": "需求评审问题清单"}'
wecom-cli doc edit_doc_content '{"docid": "DOCID", "content": "# 需求评审问题清单\n\n...", "content_type": 1}'
```

这个流能体现 4SAPI 的价值。

同一份需求文档可以用不同模型处理：

```text
便宜模型先做结构抽取
强模型做风险判断
代码模型生成接口影响分析
```

4SAPI 统一记录每一步模型调用，方便看成本。

### 14.3 会议纪要模板生成器

适合项目经理。

流程：

```text
用户输入会议主题和参会人
-> Agent 生成会议纪要模板
-> wecom-cli 创建文档
-> 写入议程、背景、待讨论问题、行动项表格
```

模板可以包含：

```markdown
# 会议纪要

## 会议信息

| 项目 | 内容 |
| --- | --- |
| 主题 | 需求评审 |
| 时间 | 2026-06-19 15:00 |
| 参会人 | 产品、研发、测试 |

## 背景

## 讨论要点

## 决议

## 行动项

| 事项 | 负责人 | 截止时间 | 状态 |
| --- | --- | --- | --- |
```

第一篇文章做到这里已经很充实。

下一篇再把行动项同步到智能表格或待办。

## 15. 本地配置和缓存

wecom-cli 的默认路径：

| 项目 | 默认位置 |
| --- | --- |
| 配置目录 | `~/.config/wecom` |
| 机器人凭证 | `<config_dir>/bot.enc` |
| MCP 配置缓存 | `<config_dir>/mcp_config.enc` |
| 媒体临时目录 | `<system_tmp>/wecom/media` |

可用环境变量：

| 变量 | 作用 |
| --- | --- |
| `WECOM_CLI_CONFIG_DIR` | 覆盖默认配置目录 |
| `WECOM_CLI_TMP_DIR` | 覆盖媒体临时目录根目录 |
| `WECOM_CLI_LOG_LEVEL` | 打开 stderr 日志并设置级别 |
| `WECOM_CLI_LOG_FILE` | 打开 JSON 日志输出 |
| `WECOM_CLI_MCP_CONFIG_ENDPOINT` | 覆盖默认 MCP 配置接口地址 |

查看缓存：

```bash
wecom-cli cache status
```

清理缓存：

```bash
wecom-cli cache clear
```

如果你刚改了机器人权限，但 CLI 里的工具列表还是旧的，可以先清缓存再重试。

## 16. 常见排错

### 16.1 help 都看不了

有些帮助信息需要动态获取工具 schema。

如果 `wecom-cli doc --help` 不正常，先检查：

```text
是否已经 init
auth 状态是否 authorized
网络是否能访问企业微信相关服务
机器人权限是否配置完整
缓存是否过期或损坏
```

### 16.2 文档读取失败

检查：

```text
URL 是否属于 /doc/ 或 /sheet/
智能表格是否误用了 get_doc_content
当前用户或机器人是否有文档权限
是否需要轮询 task_id
errcode 和 errmsg 是什么
```

### 16.3 Agent 写入内容格式乱

建议固定输出 Markdown。

并要求 Agent 先生成草稿，再调用 CLI 写入。

不要让 Agent 一边想内容一边拼 JSON。

更稳的流程是：

```text
先生成 Markdown 文本
检查标题、表格、列表
再把文本作为 content 写入 edit_doc_content
```

### 16.4 不要直接展示内部 ID

如果后续涉及成员字段，必须用 `contact get_userlist` 做转换。

不要把 `userid`、`creator_id`、`follower_id` 直接展示给业务用户。

## 17. 成本与风险提示

wecom-cli 本身不是按 Token 收费的模型服务。

但当你把它接到 Agent 后，成本主要来自：

```text
模型调用
文档内容长度
多轮读取和改写
复杂任务的反复推理
日志和存储
```

这就是 4SAPI 适合放在中间的原因。

你可以把任务拆成：

| 环节 | 模型策略 |
| --- | --- |
| 文档结构抽取 | 便宜模型 |
| 风险判断 | 强模型 |
| 改写润色 | 中档模型 |
| 格式修复 | 便宜模型 |

同时通过 4SAPI 看每一步花了多少。

风险上，第一篇要强调：

```text
文档可能包含敏感信息。
模型调用要遵守企业数据规则。
Secret 不要泄露。
文档覆写前最好先读取原文或保留备份。
生产环境不要让 Agent 无确认地覆盖重要文档。
```

## 18. 总结

企业微信 CLI 很适合写进【大模型API中转站】系列。

但第一篇建议从文档能力切入。

它的主线不是“炫技”，而是把 AI Agent 从聊天框拉进真实办公流：

```text
创建文档
读取文档
分析文档
生成 Markdown
写回企业微信
```

这条链路足够小，也足够实用。

它和 4SAPI 的分工也很清楚：

```text
wecom-cli 管企业微信文档。
4SAPI 管模型调用、成本和日志。
Agent 管任务编排。
```

下一篇可以继续往前走：

```text
从企业微信消息中提取信息，
写入智能表格，
生成待办、会议和日程，
让 Agent 真正形成办公执行闭环。
```

## 资料来源

- 企业微信 CLI GitHub 仓库：<https://github.com/WecomTeam/wecom-cli>
- `@wecom/cli` NPM 包：<https://www.npmjs.com/package/@wecom/cli>
- 企业微信帮助中心：如何获取 Bot ID 和 Secret：<https://open.work.weixin.qq.com/help2/pc/cat?doc_id=21677>
- 企业微信智能机器人 / OpenClaw 更新日志：<https://work.weixin.qq.com/nl/index/openclaw>

