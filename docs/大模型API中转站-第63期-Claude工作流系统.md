---
title: "【大模型API中转站】第63期 Claude工作流系统 | 不只会聊天"
category: 人工智能
tags:
  - 大模型API中转站
  - Claude
  - 工作流
  - Skills
  - Projects
  - Connectors
  - 4SAPI
  - 4sapi.com
description: "基于 Claude 官方 Projects、Skills、Connectors、MCP 文档和社区实践，拆解如何把 Claude 从聊天工具升级成可复用的工作流系统，并用 4sapi.com 这类大模型 API 中转站完成统一接入、成本治理和工程落地。"
---

# 【大模型API中转站】第63期 Claude工作流系统 | 不只会聊天

本文是【大模型API中转站】系列的第63篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

这篇不是再教你“怎么问 Claude 写一段文案”。

那种玩法很多人已经会了。

这篇真正要解决的是另一个问题：

```text
为什么你明明天天用 Claude，却还是感觉每次都在从零开始？
```

我结合了 Claude 官方关于 Projects、Skills、Connectors、MCP、Claude Code 的文档，也看了一些 Reddit、Hacker News 上开发者讨论 Claude 工作流的帖子。一个很明显的共识是：

```text
真正提效的不是一次提示词写得多漂亮，而是把重复任务做成系统。
```

如果你是开发者、独立创作者、企业内部运营或产品团队，Claude 最有价值的地方不是“它能回答”，而是它能被放进一个稳定的工作流里。

而大模型 API 中转站，例如 4sapi.com 这类服务，解决的正是工程落地里的另一个问题：

```text
模型能力怎么统一接入？
调用成本怎么控制？
Key、日志、限流、错误重试怎么管理？
```

所以这篇会从两个层面讲：

1. 普通用户怎么把 Claude 用成工作系统。
2. 开发者怎么把 Claude 接进自己的业务系统。

## 1. 先说结论：Claude 不是一个入口，而是一组能力

很多人打开 Claude，只会使用 Chat。

Chat 没问题。

它适合临时任务：

- 改一段文案
- 翻译一封邮件
- 总结一篇文章
- 解释一个概念
- 生成几个标题

但如果你长期做内容、写代码、研究资料、运营账号、管理客户、做产品文档，光靠 Chat 会有三个问题。

第一，上下文不稳定。

你今天说了账号定位，明天换一个对话又要说一遍。

第二，流程不稳定。

同样是写文章，今天 Claude 先给标题，明天先写正文，后天又忘了帮你检查风险表达。

第三，结果不可复用。

你每次都在复制提示词，每次都在手动上传资料，每次都在手动整理输出。

真正成熟的用法，是把 Claude 拆成几层能力：

| 能力 | 适合做什么 | 典型问题 |
| --- | --- | --- |
| Chat | 临时问答 | 这段话怎么改？ |
| Project | 长期上下文 | 这个账号/产品/研究库要持续运营 |
| Skill | 固定流程 | 每次都按同一套步骤处理 |
| Connector | 连接外部工具 | 从 Drive、GitHub、知识库读取资料 |
| Plugin | 打包一组能力 | 把 Skills、工具和连接器组合起来 |
| API | 系统调用 | 让自己的产品自动调用 Claude |
| 中转站 | 多模型统一入口 | Claude、GPT、Gemini 用一套路由和计费 |

只用 Chat，相当于只用了一把螺丝刀。

而工作流系统，是工具箱。

## 2. Project：长期任务要先有一个“工作台”

Claude 官方对 Projects 的定位很清楚：它是把聊天、上下文和项目知识组织在一起的空间。

简单说：

```text
Project 解决的是“长期上下文”的问题。
```

适合放进 Project 的任务，都有几个特点：

- 会持续推进
- 资料会越来越多
- 上下文很重要
- 输出需要保持风格一致
- 不是一次问答能结束

比如：

- 运营一个 AI 工具公众号
- 做一套企业内部知识库
- 维护一个 SaaS 产品文档
- 跟踪某个行业研究主题
- 写一本电子书
- 管理一个课程项目
- 维护品牌视觉和内容规范

以“AI 工具公众号”为例，Project 里可以放这些资料：

- 账号定位
- 目标读者画像
- 历史爆款文章
- 标题风格
- 禁用表达
- 常用结构模板
- 产品卖点
- 发布平台规范
- 竞品拆解
- 复盘记录

这样你以后让 Claude 写选题、改文章、做复盘，就不需要每次重新解释“我是谁、我要写给谁、风格是什么”。

这里还有一个关键点：Claude 的项目知识支持 RAG，也就是检索增强生成。官方说明里提到，当项目知识接近上下文限制时，Claude 可以通过 RAG 扩展可用知识容量。

这意味着 Project 不只是“把资料塞进聊天窗口”。

更像是：

```text
把长期资料做成一个可检索的工作空间。
```

不过要注意，Project 不是万能资料库。

如果资料经常变化，比如订单状态、实时库存、工单进度、日志指标，最好不要只靠 Project 手动上传。更合理的方式是通过 Connector 或 API 连接实时数据源。

## 3. Skill：重复流程要变成“可调用方法”

如果 Project 解决的是“长期上下文”，那 Skill 解决的是“固定流程”。

一句话判断：

```text
资料长期变化，用 Project。
方法反复出现，用 Skill。
```

Claude 的 Skills 可以理解成一组可复用的指令、脚本、模板和资源。官方的 Skills API 文档里也强调，Skill 可以在消息中被调用，并且可以配合文件生成、多轮会话、长任务、多个 Skills 组合使用。

这对普通用户和开发者都很重要。

普通用户可以用 Skill 固定内容流程。

开发者可以把 Skill 当成“轻量级任务模块”。

举几个例子。

### 公众号长文 Skill

输入：

- 主题
- 素材
- 目标读者
- 产品植入点

流程：

1. 判断选题角度
2. 生成 5 个标题
3. 输出文章大纲
4. 扩写正文
5. 检查夸大表达
6. 加入案例和步骤
7. 输出 Markdown 发布版

### API 接入排查 Skill

输入：

- 报错信息
- 请求代码
- 环境变量
- 模型名称

流程：

1. 判断是鉴权问题、网络问题、格式问题还是模型名问题
2. 检查 base_url 和 API key 读取方式
3. 检查请求体是否符合 OpenAI 兼容格式
4. 给出最小复现脚本
5. 输出排查清单

### 竞品分析 Skill

输入：

- 产品官网
- 定价页
- 文档页
- 用户评论

流程：

1. 提炼核心卖点
2. 拆解目标用户
3. 对比价格和限制
4. 总结差异化表达
5. 给出内容选题建议

社区里很多 Claude 用户提到一个实践：不要把所有规则都堆进一个超长提示词里，而是把固定流程拆成 Skill。这样既能减少重复输入，也能让 Claude 在需要的时候加载对应能力。

这就是 Skill 的真正价值：

```text
不是让提示词更长，而是让流程更稳定。
```

## 4. Customize、Memory 和项目规则：哪些该全局，哪些该局部？

很多人一看到 Customize 或 Memory，就想把所有偏好都写进去。

这其实不太稳。

更合理的拆法是：

| 类型 | 应该放哪里 | 示例 |
| --- | --- | --- |
| 全局个人偏好 | Customize / Memory | 我喜欢口语化、少用官话 |
| 某个长期项目资料 | Project | 账号定位、品牌语气、历史内容 |
| 可复用工作步骤 | Skill | 文章润色流程、会议纪要流程 |
| 团队硬规则 | 项目规范文件 | 不泄露 Key、不加无关依赖 |
| 实时业务数据 | Connector / API | 工单、订单、库存、日志 |

举个例子。

如果你是 AI 工具内容创作者，Customize 可以写：

```text
默认用中文输出。
表达口语化，但不要过度营销。
面向有一定技术基础的开发者和独立创作者。
不要承诺收益，不要制造焦虑。
需要给出可执行步骤和风险提示。
```

Project 里放：

```text
账号定位
历史文章
标题风格
4SAPI 产品资料
读者画像
竞品资料
发布规范
```

Skill 里放：

```text
大模型 API 中转站文章生成流程：
1. 判断选题痛点
2. 设计标题
3. 写系列导语
4. 给出原理速览
5. 加入接入教程
6. 补成本和风险提示
7. 结尾自然提及 4SAPI
```

这样分层以后，Claude 不会每次都背着一大包无关上下文干活。

## 5. Connectors：连接外部资料，但要先想清楚权限

Connectors 的价值，是让 Claude 不只看聊天窗口里的内容。

它可以连接外部工具和数据源，例如：

- Google Drive
- Microsoft 365
- GitHub
- Slack
- 邮箱
- 项目管理系统
- 企业知识库
- 自建 MCP 服务

Anthropic 官方关于 Connectors 的文档里提到，自定义连接器可以通过 remote MCP 暴露给 Claude 使用。另一个关键点是：远程 MCP 连接通常是从 Anthropic 云端基础设施访问你的 MCP 服务，而不是从你的本地电脑直接访问。

这点很多人会忽略。

如果你的 MCP 服务在公司内网、VPN 后面、或者防火墙后面，Claude 的远程连接器未必能访问。你需要确认网络可达性、鉴权方式和权限范围。

Connector 适合这些场景：

- 从文档库读取最新产品资料
- 从 GitHub 读取指定仓库文件
- 从项目管理工具提取待办
- 从云盘读取历史内容
- 从知识库回答内部问题
- 从日志系统提取异常摘要

但连接器也有风险。

尤其要看三件事：

1. 能读什么。
2. 能不能写。
3. 会不会接触敏感数据。

如果一个 Connector 既能读客户资料，又能发送邮件、创建文件、修改状态，那它就不是“资料入口”，而是“操作入口”。

这时必须做权限分级。

建议：

- 先只开只读权限
- 先接测试空间，不接生产数据
- 每个 Connector 单独说明用途
- 敏感字段做脱敏
- 操作型工具必须有人确认
- 团队账号要有审计日志

## 6. Plugin：当 Skill 和 Connector 不够用时，再考虑打包

很多新手一上来就想做 Plugin。

其实没必要。

Plugin 更适合把一组能力打包起来：

- Skills
- MCP servers
- 工具配置
- 模板文件
- 资产文件
- 自动化流程

比如你想做一个“社媒内容运营插件”，它可能包含：

- 选题生成 Skill
- 标题优化 Skill
- 风险表达检查 Skill
- 封面文案 Skill
- 连接历史内容数据库的 Connector
- 连接数据看板的 MCP
- 每周复盘模板

这就不是一个提示词能稳定完成的任务。

它是一套工作系统。

社区里有开发者把 Plugins 理解成“Skills、slash commands、agents、hooks、MCP servers 的打包方式”。这个理解很适合开发者：Plugin 不是为了显得高级，而是为了分发和复用。

普通人前期不用急。

先把 Project 和 Skill 用顺，再考虑 Connector。

等你发现“这几个能力经常一起用，而且团队里多人都要装”，再做 Plugin。

## 7. Claude Code、Hooks 和 MCP：开发者的工作流增强层

如果你是开发者，还会接触 Claude Code。

官方文档对 Claude Code 的描述是：它可以读取代码库、编辑文件、运行命令，并和开发工具集成。

这和普通 Claude Chat 最大的区别在于：

```text
Chat 是回答。
Claude Code 是执行。
```

Claude Code 里有几个对工作流很关键的能力。

### Hooks

Hooks 可以在关键节点自动执行动作。

官方文档里举到的典型场景包括：

- 文件编辑后自动格式化
- 命令执行前阻止危险操作
- Claude 需要输入时发送通知
- 会话开始时注入上下文
- 工具调用前后做检查

这对团队很有用。

比如你可以设计：

```text
每次改完 TypeScript 文件后自动运行 eslint。
每次尝试删除文件前要求人工确认。
每次任务结束后生成变更摘要。
每次调用 API 测试前检查环境变量是否存在。
```

### MCP

MCP 解决的是“连接外部工具”的问题。

Claude Code 的 MCP 文档提到，它可以连接外部工具、数据库、API、issue tracker、监控看板等。换句话说，当你发现自己总是在手动复制 Jira、GitHub、日志平台、数据库里的内容给 Claude，就应该考虑 MCP。

### Subagents

复杂任务可以拆给不同子代理。

比如：

- 一个负责读代码
- 一个负责写测试
- 一个负责安全审查
- 一个负责文档更新
- 一个负责性能风险

但这里要提醒一句：子代理不是越多越好。

如果任务边界不清楚，多个代理反而会互相污染上下文。

适合拆代理的任务，一般有明确角色、明确输入、明确输出。

## 8. API 接入：什么时候该从 Claude.ai 走向系统调用？

当你的需求停留在个人使用时，Claude.ai、Project、Skill 已经够用。

但只要出现下面这些情况，就应该考虑 API：

- 要在自己的产品里调用 Claude
- 要批量处理大量文本
- 要接入内部后台
- 要自动生成报告
- 要把模型能力放进工作流系统
- 要统一管理多模型成本
- 要做团队级调用统计

这时你面对的就不是“怎么问 Claude”，而是工程问题：

- base_url 怎么配置
- API Key 怎么管理
- 调用失败怎么重试
- 超时怎么处理
- 模型路由怎么做
- 日志记录到什么粒度
- 成本怎么统计
- 敏感数据怎么脱敏
- 用户输入怎么审计

这就是大模型 API 中转站的典型价值。

比如用 4sapi.com 这类中转服务，你可以把自己的应用统一接到一个入口，再根据模型、任务、成本、速度做路由。

请求链路可以是：

```text
你的应用 -> 4SAPI / 大模型 API 中转站 -> Claude / GPT / Gemini -> 返回结果
```

工程上可以获得几个好处：

- 多模型统一格式
- API Key 集中管理
- 调用日志更容易排查
- 模型切换成本更低
- 可以做成本统计
- 可以按任务选择模型
- 可以给不同业务线做限流

这对企业和独立开发者都很实用。

## 9. 一个完整案例：内容工作流系统怎么搭？

假设你要做一个“AI 工具类公众号内容系统”。

不要一上来就写提示词。

先拆架构。

### 第一层：Project 存长期资料

放这些内容：

- 账号定位
- 读者画像
- 标题风格
- 已发布文章
- 产品资料
- 发布规范
- 禁用表达
- 常见案例

### 第二层：Skill 固定文章流程

比如“大模型 API 中转站教程 Skill”：

```text
输入：
- 主题
- 参考资料
- 产品植入点
- 目标读者

流程：
1. 判断痛点
2. 给出标题
3. 写系列导语
4. 解释原理
5. 给出接入步骤
6. 增加代码示例
7. 补成本和风险提示
8. 写总结和系列导航

输出：
- Markdown 正文
- 摘要
- 标签
- 发布检查清单
```

### 第三层：Connector 读取资料

可以连接：

- Google Drive：历史文章和素材
- GitHub：代码示例
- Notion / 飞书：选题库
- 数据看板：文章表现

### 第四层：API 批量生成和测试

后台可以提供一个表单：

```text
主题：
目标读者：
参考资料链接：
产品植入点：
文章长度：
输出平台：
```

提交后调用模型 API。

### 第五层：中转站做路由和成本治理

简单任务用便宜模型。

复杂长文用 Claude。

代码示例用编码能力强的模型。

最终审稿用另一个模型交叉检查。

中转站负责统一 API 格式、模型路由、成本统计、日志排查。

这样你得到的不是“让 Claude 写文章”，而是一条内容生产流水线。

## 10. Skill 还是 Project？用这张表判断

很多人最容易卡在这里。

直接看表：

| 问题 | 适合 Skill | 适合 Project |
| --- | --- | --- |
| 每次步骤差不多吗？ | 是 | 不一定 |
| 每次输入会变化吗？ | 是 | 是 |
| 是否需要长期积累资料？ | 不一定 | 是 |
| 是否依赖历史上下文？ | 少量 | 强依赖 |
| 是否可以跨项目复用？ | 是 | 通常不能 |
| 是否像一个工具？ | 是 | 否 |
| 是否像一个工作空间？ | 否 | 是 |

举例：

适合 Skill：

- 会议纪要整理
- 公众号润色
- 小红书封面文案
- API 报错排查
- 竞品分析
- 数据报告生成
- 代码审查清单

适合 Project：

- 我的公众号
- 我的产品文档库
- 我的投资研究库
- 我的课程项目
- 我的企业知识库
- 我的品牌内容系统

一句话：

```text
Skill 是方法，Project 是场景。
```

## 11. 和大模型 API 中转站结合的三种模式

### 模式一：个人增强型

适合个人创作者。

你主要在 Claude.ai 里使用 Projects 和 Skills，少量用 4SAPI 做 API 测试、批量生成或多模型对比。

优点：

- 上手快
- 成本低
- 不需要复杂开发

缺点：

- 自动化程度有限
- 数据还需要手动整理

### 模式二：团队工具型

适合小团队。

团队有一个后台工具，通过 API 调用 Claude、GPT 等模型。4SAPI 负责统一入口，后台负责业务流程。

典型功能：

- 选题生成
- 长文生成
- 代码审查
- 客服回复草稿
- 日报周报
- 知识库问答

优点：

- 团队流程统一
- 成本可统计
- 权限可管理

缺点：

- 需要后端开发
- 要做好日志和权限

### 模式三：企业工作流型

适合企业内部系统。

模型接入层、权限层、审计层、数据源、业务系统都要分开设计。

建议结构：

```text
业务系统 -> 工作流编排 -> 大模型 API 中转站 -> 多模型服务
           |
           -> 日志审计
           -> 权限控制
           -> 敏感数据脱敏
           -> 成本统计
```

这种模式下，中转站不是一个“便宜 API 地址”，而是模型治理层的一部分。

## 12. 成本怎么控？别只盯模型单价

很多人做 Claude 接入，只看每百万 token 多少钱。

这不够。

真实成本通常来自：

- 提示词太长
- 每次都重复塞背景
- 没有缓存固定资料
- 失败后重复重试
- 输出太长但没人看
- 不区分任务复杂度
- 所有任务都用最贵模型

建议做四件事。

第一，固定背景放 Project 或数据库，不要每次都塞进 prompt。

第二，固定流程做 Skill，不要每次重复写长提示词。

第三，简单任务走便宜模型，复杂任务再走 Claude。

第四，在中转站或后台记录每类任务的 token 用量。

可以用一个简单公式估算：

```text
单次成本 = 输入 token 成本 + 输出 token 成本 + 重试成本 + 中转/服务器成本
```

如果一个任务每天跑 1000 次，提示词每次多 1000 token，一个月就是很大的浪费。

## 13. 权限和合规：哪些事不要让 AI 自动做？

Claude 可以帮你做很多事，但不是所有事都应该放权。

尤其是接入 Connector、MCP、API 以后，要分清：

```text
读信息
生成建议
执行操作
产生外部影响
```

风险是逐级升高的。

相对安全：

- 总结文档
- 生成草稿
- 检查格式
- 提炼待办
- 输出建议

需要谨慎：

- 修改数据库
- 发送邮件
- 提交表单
- 发布内容
- 删除文件
- 调整权限
- 操作生产系统

建议的权限策略：

- 默认只读
- 写操作单独授权
- 高风险操作人工确认
- 生产数据先脱敏
- API Key 不进日志
- 日志不记录完整用户隐私
- 团队定期审计 Connector 权限

这也是为什么文章开头强调合法合规。

大模型 API 中转站可以讨论格式转换、成本优化、负载均衡、计费管理，但不应该被用于恶意绕过平台限制或处理违规用途。

## 14. 给开发者的最小接入示例

如果你想把 Claude 工作流接进自己的系统，可以先从一个最小后端接口开始。

环境变量示意：

```env
LLM_BASE_URL=https://你的中转站地址/v1
LLM_API_KEY=你的中转站Key
LLM_MODEL=claude-sonnet
```

Python 测试脚本：

```python
import os
import requests

base_url = os.environ["LLM_BASE_URL"]
api_key = os.environ["LLM_API_KEY"]
model = os.environ.get("LLM_MODEL", "claude-sonnet")

payload = {
    "model": model,
    "messages": [
        {
            "role": "system",
            "content": "你是一个严谨的内容工作流助手，输出 Markdown，避免夸大表达。"
        },
        {
            "role": "user",
            "content": "请把这段素材整理成公众号教程大纲，并给出风险提示。"
        }
    ],
    "temperature": 0.6
}

response = requests.post(
    f"{base_url}/chat/completions",
    headers={
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json"
    },
    json=payload,
    timeout=60
)

response.raise_for_status()
print(response.json())
```

这里有几个基本要求：

- Key 只放环境变量
- 超时时间要设置
- 错误要有可读提示
- 不要把完整用户隐私写入日志
- 请求和响应要留最小必要审计信息
- 生产环境要做限流和重试

## 15. 可以直接复用的工作流提示词

### 了解现有流程

```text
先不要生成最终内容。
请根据我提供的资料，判断这个任务应该拆成：
1. 长期上下文
2. 固定流程
3. 外部数据源
4. API 调用
5. 人工确认点

请用表格输出，并说明哪些适合放 Project，哪些适合做 Skill。
```

### 创建 Skill 草案

```text
请帮我设计一个可复用 Skill，用于处理 [任务名称]。

要求：
1. 说明触发场景
2. 定义输入字段
3. 拆解执行步骤
4. 给出输出格式
5. 列出风险检查项
6. 说明哪些内容不应该自动执行
```

### 设计 API 接入方案

```text
我要把这个 Claude 工作流接入自己的后台系统。
请先不要写代码，先给出架构方案：
1. 前端需要收集哪些字段
2. 后端如何组织 prompt
3. 如何通过大模型 API 中转站调用模型
4. 如何记录成本和错误
5. 哪些数据需要脱敏
6. 哪些动作必须人工确认
```

### 做上线前检查

```text
请检查这个 Claude API 工作流是否适合上线。
重点看：
1. Key 是否可能泄露
2. 是否记录敏感数据
3. 是否有超时和重试
4. 是否有成本控制
5. 是否有人工确认点
6. 是否能回滚到备用模型
```

## 16. 最容易踩的 8 个坑

第一，所有事都放 Chat。

结果是每次都重新解释背景。

第二，把 Project 当垃圾桶。

什么文件都上传，最后模型检索不到重点。

第三，把 Skill 写成超长提示词。

Skill 应该是流程，不是堆规则。

第四，Connector 不看权限。

尤其是能写入、能发送、能删除的工具，要非常谨慎。

第五，没有区分临时任务和长期任务。

临时任务用 Chat，长期任务用 Project。

第六，没有成本记录。

上线后才发现 token 用量失控。

第七，没有错误兜底。

模型一报错，用户页面直接崩。

第八，让 AI 替你做最终决策。

AI 可以生成、整理、检查、提醒，但最终发布、交付、合规、财务和业务判断，仍然由人负责。

## 17. 最后总结

你的 Claude 使用是否及格，不看你会不会写漂亮提示词。

而是看你有没有做到这几件事：

- 临时问题用 Chat
- 长期任务进 Project
- 重复流程做 Skill
- 外部资料走 Connector
- 复杂组合再做 Plugin
- 系统调用走 API
- 多模型治理交给大模型 API 中转站

一句话总结：

```text
Claude 真正提效的关键，不是问得更勤，而是把它放进你的工作系统里。
```

如果你只是个人使用，先从 Project 和 Skill 开始。

如果你是开发者或团队，下一步就应该考虑 API 接入、权限设计、成本统计和模型路由。

像 4sapi.com 这类大模型 API 中转站，适合放在模型接入层，帮你统一 Claude、GPT、Gemini 等模型的调用入口。它不是替你做业务决策，而是让模型能力更容易被系统化、流程化、可治理地使用起来。

## 资料来源与延伸阅读

- [Using Agent Skills with the API - Claude API Docs](https://platform.claude.com/docs/en/build-with-claude/skills-guide)
- [What are projects? - Claude Help Center](https://support.claude.com/en/articles/9517075-what-are-projects)
- [Retrieval augmented generation for projects - Claude Help Center](https://support.claude.com/en/articles/11473015-retrieval-augmented-generation-rag-for-projects)
- [Use connectors to extend Claude's capabilities - Claude Help Center](https://support.claude.com/en/articles/11176164-use-connectors-to-extend-claude-s-capabilities)
- [Get started with custom connectors using remote MCP - Claude Help Center](https://support.claude.com/en/articles/11175166-get-started-with-custom-connectors-using-remote-mcp)
- [Extend Claude with skills - Claude Code Docs](https://docs.anthropic.com/en/docs/claude-code/skills)
- [Connect Claude Code to tools via MCP - Claude Code Docs](https://docs.anthropic.com/en/docs/claude-code/mcp)
- [Automate actions with hooks - Claude Code Docs](https://docs.anthropic.com/en/docs/claude-code/hooks-guide)
- [社区讨论：The Busy Person's Intro to Claude Skills](https://www.reddit.com/r/ClaudeAI/comments/1pq0ui4/the_busy_persons_intro_to_claude_skills_a_feature/)
- [社区讨论：Understanding CLAUDE.md vs Skills vs Slash Commands vs Plugins](https://www.reddit.com/r/ClaudeAI/comments/1ped515/understanding_claudemd_vs_skills_vs_slash/)
