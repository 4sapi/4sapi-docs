---
title: "【大模型API中转站】第57期 企业微信CLI闭环 | 消息到表格"
category: 人工智能
tags:
  - 大模型API中转站
  - 企业微信
  - wecom-cli
  - 智能表格
  - 企业微信消息
  - AI Agent
  - 4SAPI
description: "延续企业微信 CLI 文档接入篇，进一步拆解消息读取、附件下载、智能表格记录、待办、会议和日程能力，设计一条从企业微信群聊到结构化台账再到执行提醒的 Agent 办公闭环。"
---

# 【大模型API中转站】第57期 企业微信CLI闭环 | 消息到表格

本文是【大模型API中转站】系列的第57篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

上一篇我们用企业微信 CLI 跑通了文档 Agent：

```text
安装 wecom-cli
初始化 API模式智能机器人
创建企业微信文档
读取和写入 Markdown
用 4SAPI 管模型调用、成本和日志
```

这一篇继续往前走。

真正的企业办公 Agent，不应该只会写一篇文档。

它更有价值的地方是把日常协作中的零散信息变成结构化行动：

```text
项目群消息
-> 需求、风险、行动项
-> 智能表格台账
-> 待办提醒
-> 会议或日程安排
-> 必要时消息回写
```

这就是本篇的主题：

```text
用 wecom-cli 做一条从消息到智能表格，再到待办/会议/日程的办公闭环。
```

4SAPI 在这里继续扮演模型网关：

```text
wecom-cli 负责企业微信工具调用。
Agent 负责规划和编排。
4SAPI 负责模型、Key、成本和日志。
```

## 1. 为什么第二篇要讲“闭环”

上一篇的文档能力解决的是“沉淀”。

比如：

```text
生成周报
创建会议纪要
读取需求文档
写评审问题清单
```

但企业里真正费时间的，往往不是写一篇文档，而是下面这条链路：

```text
群里说了一堆
没人整理
行动项散在聊天里
负责人不明确
截止时间没人追
过几天又重新问一遍
```

这类场景最适合 Agent。

因为 Agent 擅长：

```text
从非结构化文本中提取结构化信息
判断优先级
生成行动项
把内容写入表格
创建提醒
根据日程安排会议
```

企业微信 CLI 正好提供这些工具层能力：

| 能力 | category | 用途 |
| --- | --- | --- |
| 消息 | `msg` | 拉取会话、读取文本、下载附件、发送文本 |
| 智能表格 | `doc` / `smartsheet_*` | 建任务台账、客户线索表、需求池 |
| 通讯录 | `contact` | 把姓名匹配成 userid |
| 待办 | `todo` | 创建提醒和分派事项 |
| 会议 | `meeting` | 创建预约会议、查会议详情 |
| 日程 | `schedule` | 查询日程、查询多人闲忙、创建日程 |

所以第二篇要讲闭环，而不是继续停留在“CLI 命令列表”。

## 2. 合规边界：消息能力要谨慎

消息能力很强，也最敏感。

企业微信群聊里可能有：

```text
客户信息
合同条款
报价
内部决策
账号权限
个人信息
未公开产品计划
```

所以文章必须先把边界写清楚。

本文只讨论：

```text
官方 wecom-cli
API模式智能机器人
授权范围内的消息读取
用户确认后的消息发送
合规的数据处理和日志审计
```

不讨论：

```text
绕过可见范围
抓取客户端数据库
模拟登录
绕过企业权限
自动群发营销
未确认自动回复客户
```

如果接 4SAPI 或其他模型网关，还要提前判断哪些内容能进入模型。

建议给企业读者一个简单规则：

```text
能写进普通项目周报的信息，通常可以进入模型总结。
不能发到项目周报里的敏感信息，先脱敏或不要交给外部模型。
```

## 3. 总体架构

一条完整链路可以设计成：

```text
企业微信群聊
  -> wecom-cli msg 读取最近消息
  -> Agent 清洗和分段
  -> 4SAPI 调模型抽取需求、风险、行动项
  -> wecom-cli doc smartsheet 写入智能表格
  -> wecom-cli todo 创建待办
  -> wecom-cli schedule / meeting 安排讨论
  -> 用户确认后 wecom-cli msg 回写通知
```

它的关键不是“每个命令都能跑”，而是把每个动作放在正确位置。

| 层 | 负责什么 |
| --- | --- |
| 企业微信 | 真实业务数据和协作对象 |
| wecom-cli | 读取、写入、创建、查询 |
| Agent | 判断任务、拆步骤、处理异常 |
| 4SAPI | 模型调用、成本、日志、路由 |
| 大模型 | 总结、分类、抽取、改写、判断 |

这样就不会把 4SAPI 写成企业微信插件，也不会把 wecom-cli 写成模型服务。

## 4. 前置条件

延续上一篇，假设已经完成：

```bash
npm install -g @wecom/cli
wecom-cli init
wecom-cli auth show --auth-status
```

成功状态：

```text
authorized
```

如果还没初始化，请先回到第 56 期完成：

```text
获取 Bot ID 和 Secret
配置机器人可见范围
初始化 wecom-cli
跑通 doc create_doc
```

本篇继续使用这些能力：

```bash
wecom-cli msg --help
wecom-cli doc --help
wecom-cli contact get_userlist '{}'
wecom-cli todo --help
wecom-cli schedule --help
wecom-cli meeting --help
```

注意：某些工具列表和参数 schema 是动态获取的，需要凭证和网络。

如果帮助命令也失败，先别急着怀疑模型，先排查 CLI 初始化和企业微信权限。

## 5. 读取会话列表

先看最近有哪些会话：

```bash
wecom-cli msg get_msg_chat_list '{"begin_time": "2026-06-11 00:00:00", "end_time": "2026-06-18 23:59:59"}'
```

这个命令适合做两件事。

第一，帮 Agent 找到目标会话。

用户通常会说：

```text
帮我整理一下项目群最近的讨论。
```

但 Agent 需要的是：

```text
chatid
chat_type
时间范围
```

所以流程应该是：

```text
先 get_msg_chat_list
-> 按 chat_name 匹配项目群
-> 如果多个候选，让用户选
-> 再读取消息
```

第二，帮用户确认范围。

比如返回多个相似群：

```text
产品项目群
产品项目群-外包
产品项目群-测试
```

Agent 不应该自行猜测。

它应该展示候选，让用户确认。

## 6. 拉取最近 7 天消息

读取消息：

```bash
wecom-cli msg get_message '{"chat_type": 2, "chatid": "CHATID", "begin_time": "2026-06-17 09:00:00", "end_time": "2026-06-17 18:00:00"}'
```

这里有一个重要限制：

```text
消息记录支持最近 7 天内。
```

所以不要把文章写成：

```text
一键总结过去一年的项目群。
```

更准确的写法是：

```text
用 wecom-cli 处理最近 7 天消息，再把长期结果沉淀到文档、智能表格或知识库。
```

这也是闭环设计的核心。

聊天记录不适合当长期数据库。

长期沉淀应该放到：

```text
企业微信文档
智能表格
知识库
项目周报
```

## 7. 下载消息附件

如果消息里有图片、文件、语音、视频，可以用：

```bash
wecom-cli msg get_msg_media '{"media_id": "MEDIAID_xxxxxx"}'
```

返回里会包含本地路径：

```text
local_path
name
type
size
content_type
```

Agent 处理附件时要守三条规则。

第一，先问用户是否下载。

不要看到群里有文件就全部下载。

第二，下载后展示完整路径。

用户需要知道文件在哪。

第三，不要把下载的历史附件当成新消息重新发送。

这些文件是为了本地分析或归档，不是为了二次转发。

如果后续要把附件内容交给模型，需要再做一次数据判断：

```text
文件是否敏感？
是否需要脱敏？
模型是否允许处理这类文件？
是否需要只提取元数据而不上传全文？
```

## 8. 用 4SAPI 做消息抽取

拿到消息后，不建议直接让强模型一次性处理全部内容。

更省钱的做法是分层：

| 步骤 | 推荐模型策略 |
| --- | --- |
| 去重、按时间切片 | 便宜模型或本地规则 |
| 提取候选行动项 | 中档模型 |
| 判断优先级和风险 | 强模型 |
| 生成回写话术 | 中档模型 |
| 格式修复 | 便宜模型 |

4SAPI 的价值在这里很明显：

```text
同一条办公流里可以按任务选择不同模型。
所有调用统一走一个 API 入口。
成本和日志能看清楚。
Key 不散落在不同 Agent 配置里。
```

一个适合给 Agent 的抽取目标：

```json
{
  "summary": "今天项目群主要讨论了支付回调超时和报表导出排队。",
  "risks": [
    {
      "title": "支付回调超时未定位",
      "level": "high",
      "evidence": "多位同事反馈生产环境偶发超时"
    }
  ],
  "actions": [
    {
      "task": "补充支付回调链路日志",
      "owner_name": "张三",
      "due": "2026-06-19 18:00",
      "source": "项目群 2026-06-18 10:32"
    }
  ]
}
```

注意让模型输出 `owner_name`，不要让它猜 `userid`。

`userid` 必须后续通过通讯录匹配。

## 9. 创建智能表格作为台账

智能表格适合存结构化结果。

比如项目行动项表：

| 字段 | 类型建议 |
| --- | --- |
| 任务名称 | 文本 |
| 来源会话 | 文本 |
| 负责人 | 成员 |
| 截止时间 | 日期 |
| 优先级 | 单选 |
| 状态 | 单选 |
| 风险说明 | 长文本 |

创建智能表格：

```bash
wecom-cli doc create_doc '{"doc_type": 10, "doc_name": "项目行动项台账"}'
```

查询子表：

```bash
wecom-cli doc smartsheet_get_sheet '{"docid": "DOCID"}'
```

查询字段：

```bash
wecom-cli doc smartsheet_get_fields '{"docid": "DOCID", "sheet_id": "SHEETID"}'
```

写入记录前一定要先查字段。

智能表格不同字段类型的值格式不一样，Agent 不能只凭字段名随便拼。

## 10. 写入行动项记录

添加记录：

```bash
wecom-cli doc smartsheet_add_records '{"docid": "DOCID", "sheet_id": "SHEETID", "records": [{"values": {"任务名称": [{"type": "text", "text": "补充支付回调链路日志"}], "优先级": [{"text": "高"}], "状态": [{"text": "未开始"}]}}]}'
```

如果有负责人字段，先查通讯录：

```bash
wecom-cli contact get_userlist '{}'
```

然后在本地按姓名或别名匹配。

不要让 Agent 直接猜：

```text
张三的 userid 应该是 zhangsan
```

正确流程：

```text
模型输出 owner_name=张三
-> contact get_userlist
-> 本地匹配 name / alias
-> 唯一匹配后拿 userid
-> 多个同名让用户选择
-> 写入智能表格成员字段
```

智能表格还有几个避坑点：

```text
先 get_fields 再 add_records。
单选/多选字段要匹配已有选项。
成员字段要用 userid。
删除字段、删除记录不可逆。
涉及图片或文件字段时使用带 + 的 auto_file helper。
```

## 11. 从行动项创建待办

写入表格只是“记录”。

待办才是“提醒”。

创建待办：

```bash
wecom-cli todo create_todo '{"content": "补充支付回调链路日志", "remind_time": "2026-06-19 09:00:00"}'
```

如果要分派给某人，仍然要先通过通讯录拿 userid。

待办列表查询：

```bash
wecom-cli todo get_todo_list '{"limit": 10}'
```

注意一个重要点：

```text
get_todo_list 返回的是概要信息，不包含完整内容和分派人。
```

所以展示给用户前，应该继续调用：

```bash
wecom-cli todo get_todo_detail '{"todo_id_list": ["TODO_ID"]}'
```

然后再把 `creator_id`、`follower_id` 转成姓名。

Agent 不应该把内部 ID 直接展示给业务用户。

## 12. 需要讨论时创建会议

有些风险不是一个待办能解决的。

比如：

```text
支付回调超时原因不明
多团队接口边界不清
客户验收标准变更
上线窗口冲突
```

这时 Agent 可以建议开会。

创建会议：

```bash
wecom-cli meeting create_meeting '{"title": "支付回调超时排查会", "meeting_start_datetime": "2026-06-19 15:00", "meeting_duration": 3600}'
```

查询会议列表：

```bash
wecom-cli meeting list_user_meetings '{"begin_datetime": "2026-06-18 00:00", "end_datetime": "2026-06-25 23:59", "limit": 100}'
```

会议能力有两个避坑点。

第一，查询范围有限：

```text
仅支持当日及前后 30 天。
```

第二，更新会议受邀成员是全量覆盖。

如果要给会议加人，不能只传新增人员。

正确流程：

```text
查询会议详情
-> 获取现有成员
-> 查询新增人员 userid
-> 合并成员列表
-> 全量提交
```

## 13. 先查闲忙再安排日程

如果不想直接创建会议，可以先查闲忙。

查询多人闲忙：

```bash
wecom-cli schedule check_availability '{"check_user_list": ["USER_ID_1", "USER_ID_2"], "start_time": "2026-06-19 14:00:00", "end_time": "2026-06-19 18:00:00"}'
```

然后让 Agent 计算共同空闲时间。

完整流程：

```text
从消息中提取需要参会的人
-> contact get_userlist 匹配 userid
-> schedule check_availability 查询闲忙
-> Agent 推荐 2-3 个候选时间
-> 用户确认
-> create_schedule 或 create_meeting
```

创建日程：

```bash
wecom-cli schedule create_schedule '{"schedule": {"start_time": "2026-06-19 15:00:00", "end_time": "2026-06-19 16:00:00", "summary": "支付回调超时排查会", "attendees": [{"userid": "USER_ID"}], "reminders": {"is_remind": 1, "remind_before_event_secs": 900, "timezone": 8}}}'
```

日程也有范围限制：

```text
日程列表查询仅支持当日前后 30 天。
```

## 14. 消息回写必须确认

闭环最后一步可能是把结果发回群里。

发送文本消息：

```bash
wecom-cli msg send_message '{"chat_type": 2, "chatid": "CHATID", "msgtype": "text", "text": {"content": "已整理今日行动项，详情见智能表格链接。"}}'
```

但这一步必须让用户确认。

推荐 Agent 固定成这样：

```text
即将向「支付项目群」发送以下消息：

已整理今日行动项，详情见智能表格链接：
https://doc.weixin.qq.com/smartsheet/xxx

确认发送吗？
```

不要让 Agent 自动把模型生成内容直接发群。

原因很现实：

```text
模型可能误解语气。
模型可能把内部判断写得太绝对。
模型可能泄露不该回写的信息。
模型可能把未确认的行动项说成已确认。
```

企业办公里，读和写不是同等风险。

读取后总结可以更自动。

写回群聊必须更谨慎。

## 15. 一条完整的实战工作流

假设用户说：

```text
帮我整理一下支付项目群今天的讨论，提取行动项，写入项目行动项表。高风险事项帮我建待办，如果需要开会先问我。
```

Agent 可以这样执行：

### 第一步：定位会话

```bash
wecom-cli msg get_msg_chat_list '{"begin_time": "2026-06-18 00:00:00", "end_time": "2026-06-18 23:59:59"}'
```

按群名匹配“支付项目群”。

如果多个候选，展示给用户选择。

### 第二步：读取消息

```bash
wecom-cli msg get_message '{"chat_type": 2, "chatid": "CHATID", "begin_time": "2026-06-18 00:00:00", "end_time": "2026-06-18 23:59:59"}'
```

把文本消息按时间切片。

非文本消息先统计，不自动下载。

### 第三步：模型抽取

通过 4SAPI 调模型，输出结构化 JSON：

```json
{
  "summary": "今日主要讨论支付回调超时、报表导出排队和上线窗口。",
  "actions": [
    {
      "task": "补充支付回调链路日志",
      "owner_name": "张三",
      "due": "2026-06-19 18:00",
      "priority": "高",
      "risk": "生产环境偶发超时，影响支付结果确认"
    }
  ],
  "meeting_suggestion": {
    "needed": true,
    "reason": "涉及支付、后端、测试多方定位，需要同步排查口径"
  }
}
```

### 第四步：匹配人员

```bash
wecom-cli contact get_userlist '{}'
```

把 `owner_name` 匹配为 userid。

如果同名，暂停让用户选。

### 第五步：写入智能表格

```bash
wecom-cli doc smartsheet_get_sheet '{"docid": "DOCID"}'
wecom-cli doc smartsheet_get_fields '{"docid": "DOCID", "sheet_id": "SHEETID"}'
wecom-cli doc smartsheet_add_records '{"docid": "DOCID", "sheet_id": "SHEETID", "records": [{"values": {"任务名称": [{"type": "text", "text": "补充支付回调链路日志"}], "优先级": [{"text": "高"}], "状态": [{"text": "未开始"}], "风险说明": [{"type": "text", "text": "生产环境偶发超时，影响支付结果确认"}]}}]}'
```

### 第六步：创建待办

高风险事项创建待办：

```bash
wecom-cli todo create_todo '{"content": "补充支付回调链路日志", "remind_time": "2026-06-19 09:00:00"}'
```

### 第七步：建议开会

不要直接创建会议，先提示：

```text
我发现一个高风险事项需要支付、后端、测试同步排查。
是否需要我查询三位负责人的明天下午闲忙，并推荐会议时间？
```

用户确认后再：

```bash
wecom-cli schedule check_availability '{"check_user_list": ["USER_ID_1", "USER_ID_2", "USER_ID_3"], "start_time": "2026-06-19 14:00:00", "end_time": "2026-06-19 18:00:00"}'
```

最后再创建日程或会议。

### 第八步：确认后回写消息

生成消息：

```text
今日行动项已整理完成：
1. 补充支付回调链路日志，负责人：张三，截止：6月19日 18:00，优先级：高

详情见项目行动项台账：
https://doc.weixin.qq.com/smartsheet/xxx
```

用户确认后：

```bash
wecom-cli msg send_message '{"chat_type": 2, "chatid": "CHATID", "msgtype": "text", "text": {"content": "今日行动项已整理完成：\n1. 补充支付回调链路日志，负责人：张三，截止：6月19日 18:00，优先级：高\n\n详情见项目行动项台账：\nhttps://doc.weixin.qq.com/smartsheet/xxx"}}'
```

这就是完整闭环。

## 16. 和 OpenClaw 插件的区别

资料收集时还会看到：

```text
WecomTeam/wecom-openclaw-plugin
@wecom/wecom-openclaw-plugin
@wecom/wecom-openclaw-cli
```

它们和 wecom-cli 相关，但不是同一个重点。

截至 2026-06-18，我查到：

| 项目 | 当前版本 |
| --- | --- |
| `@wecom/wecom-openclaw-plugin` | `2026.5.25` |
| `@wecom/wecom-openclaw-cli` | `1.1.0` |

可以这样区分：

| 方案 | 更适合谁 | 重点 |
| --- | --- | --- |
| `wecom-cli` | Codex、Claude Code、Cursor、脚本、终端自动化 | 让 Agent 在终端里调用企业微信能力 |
| `wecom-openclaw-plugin` | 使用 OpenClaw 搭企业微信机器人入口的人 | 把 OpenClaw Agent 接到企业微信通道 |

如果你写的是：

```text
本地 Agent 如何读取消息、写表格、建待办
```

主角是 `wecom-cli`。

如果你写的是：

```text
企业微信里直接和 OpenClaw Agent 对话
```

主角更可能是 OpenClaw 插件。

两者可以组合，但不要混成一件事。

## 17. 限制和风险清单

### 17.1 企业规模和能力差异

官方 README 对场景有区分：

```text
10 人以上企业：API 模式智能机器人提供文档 CLI 能力。
10 人及以下个人/小团队：提供消息、文档、日程、会议、待办等 CLI 能力。
```

所以文章不要承诺所有企业都能完整使用所有能力。

更稳的表述是：

```text
实际可用能力取决于企业规模、机器人配置、可见范围、授权状态和企业微信后台开放情况。
```

### 17.2 授权状态要可观测

GitHub issue 里有用户反馈：消息 MCP 授权有效期和本地凭证存在不是同一件事。

也就是说：

```text
bot.enc 存在，不代表消息权限一定仍然有效。
auth show --auth-status 返回 authorized，也不代表每个品类权限都没问题。
```

生产自动化要做健康检查。

至少在任务开始前验证关键命令是否可用。

### 17.3 文档搜索能力还不是完整答案

目前文档能力更适合：

```text
用户提供 docid 或 URL
Agent 创建新文档后保存 docid
围绕已知文档读写
```

GitHub issue 里有人提出文档列表和搜索能力需求。

在它成为稳定能力之前，不要把教程写成“让 Agent 自动搜索全量企业文档库”。

### 17.4 外部群和客服场景要谨慎

也有 issue 讨论外部群、客服等场景。

这类场景涉及外部联系人、客户沟通和合规边界。

本系列不要写成：

```text
自动客服群回复
自动外部群营销
自动承诺报价或合同
```

更稳的方向是内部办公：

```text
项目群总结
内部行动项
会议准备
文档沉淀
任务跟踪
```

## 18. 成本治理

这条闭环会比文档篇更耗模型。

原因是它包含多步推理：

```text
消息清洗
上下文分段
行动项抽取
风险判断
人员匹配解释
表格字段映射
待办文案生成
会议建议
回写消息润色
```

如果全部用最强模型，成本会高。

建议用 4SAPI 做模型分层：

| 环节 | 推荐策略 |
| --- | --- |
| 消息清洗 | 规则或便宜模型 |
| 行动项抽取 | 中档模型 |
| 风险判断 | 强模型 |
| 表格字段映射 | 便宜模型 + 固定 schema |
| 回写消息 | 中档模型 |
| 复核关键结论 | 强模型或人工确认 |

这样既能控制成本，又能保留关键判断质量。

## 19. 总结

企业微信 CLI 的第二篇，重点不应该再停留在“命令能不能用”。

它真正有价值的写法是闭环：

```text
从消息里发现信息
用模型抽取结构
写入智能表格
创建待办
必要时安排会议或日程
确认后回写消息
```

这条链路里：

```text
wecom-cli 是企业微信工具层。
4SAPI 是模型网关层。
Agent 是任务编排层。
企业微信是协作现场。
```

写这篇文章时，最重要的是别为了显得自动化而牺牲安全边界。

读消息要有授权。

写表格要先查字段。

建待办和会议要明确对象。

发消息必须确认。

能做到这些，企业微信 CLI 就不只是一个命令行工具，而是企业办公 Agent 的执行手脚。

## 资料来源

- 企业微信 CLI GitHub 仓库：<https://github.com/WecomTeam/wecom-cli>
- `@wecom/cli` NPM 包：<https://www.npmjs.com/package/@wecom/cli>
- 企业微信帮助中心：如何获取 Bot ID 和 Secret：<https://open.work.weixin.qq.com/help2/pc/cat?doc_id=21677>
- 企业微信智能机器人 / OpenClaw 更新日志：<https://work.weixin.qq.com/nl/index/openclaw>
- 企业微信 OpenClaw 插件仓库：<https://github.com/WecomTeam/wecom-openclaw-plugin>
- `@wecom/wecom-openclaw-plugin` NPM 包：<https://www.npmjs.com/package/@wecom/wecom-openclaw-plugin>
- `@wecom/wecom-openclaw-cli` NPM 包：<https://www.npmjs.com/package/@wecom/wecom-openclaw-cli>
