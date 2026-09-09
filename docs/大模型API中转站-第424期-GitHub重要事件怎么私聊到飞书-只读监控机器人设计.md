---
title: "GitHub 重要事件怎么私聊到飞书：只读监控机器人设计"
category: 人工智能
tags:
  - GitHub Actions
  - 飞书机器人
  - 自动化架构
description: "从通知噪音、事件范围、数据流和最小权限出发，设计一条 GitHub 到飞书私聊的只读监控链路，为后续写入操作留下安全边界。"
---

# GitHub 重要事件怎么私聊到飞书：只读监控机器人设计

GitHub 通知最麻烦的地方，不是没有通知，而是重要通知和普通通知混在一起。

Issue 新建后没人看，PR 已经等着 Review，Release 发布后却没有人及时确认。把仓库里的所有提醒都打开，评论、标签变化和机器人同步又会把真正需要处理的事情淹掉。于是通知系统很容易走到两个极端：要么漏看，要么直接静音。

一个更实用的做法，是先搭一条只读监控链路：GitHub Actions 监听少数需要人工注意的事件，读取事件元数据，整理成一条短消息，再由飞书企业自建应用机器人私聊指定用户。

本文不展开完整代码，而是解决一个更前置的问题：这类机器人应该监控什么、经过哪些组件、申请哪些权限，以及如何判断第一版已经设计正确。

## 1. 先定义机器人负责什么

第一版只做三件事：

1. 监听 GitHub 的 Issue、PR 和 Release 事件；
2. 从事件 JSON 中取出标题、仓库和链接；
3. 以飞书机器人的身份，把文本消息发给指定用户。

它不负责评论、不负责加标签、不负责合并 PR，也不负责修改代码。把“通知”和“执行”拆开，后面排错时才能判断问题来自事件触发、消息发送，还是写入动作本身。

最终链路可以画成：

```text
GitHub Issue / PR / Release
        ↓
GitHub Actions 读取事件元数据
        ↓
jq 整理标题、仓库和 URL
        ↓
使用 GitHub Secrets 取得飞书访问令牌
        ↓
调用飞书发送消息接口
        ↓
指定用户收到机器人私聊
```

## 2. 为什么只监听这几类事件

GitHub 的事件很多。第一版不需要把整个 Webhook 列表都搬进来，先回答“收到消息后要做什么”更重要。

| GitHub 事件 | 第一版监听的动作 | 适合通知什么 |
| --- | --- | --- |
| `issues` | `opened`、`reopened` | 新问题出现，或旧问题重新进入处理状态 |
| `pull_request_target` | `opened`、`reopened`、`ready_for_review`、`review_requested` | PR 需要进入 Review 阶段 |
| `release` | `published` | 新版本已经正式发布 |
| `workflow_dispatch` | 手动运行 | 验证凭证和消息链路 |

GitHub 官方文档说明，事件可以通过 `types` 限制具体活动类型；如果不限制，某些事件会在更多活动发生时触发工作流。只选需要人工行动的事件，可以减少通知量。[GitHub：Events that trigger workflows](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows)（访问日期：2026-08-07）

不要一开始监听 `push`、每条评论或所有标签变化。它们通常发生得更频繁，但不一定产生需要立即处理的动作。通知的判断标准应该是：这条消息是否会改变某个人今天的工作安排。

## 3. 数据在链路里怎样流动

### GitHub 负责触发

GitHub Actions 根据 workflow 文件里的 `on` 配置启动 Job。事件发生时，GitHub 把事件内容放在运行环境的 `GITHUB_EVENT_PATH` 文件中。

### Bash 和 jq 负责取值

工作流不需要 checkout 仓库，也不需要读取业务代码。它只从事件 JSON 中读取：

| 事件 | 标题字段 | 链接字段 |
| --- | --- | --- |
| Issue | `.issue.title` | `.issue.html_url` |
| PR | `.pull_request.title` | `.pull_request.html_url` |
| Release | `.release.name` 或 `.release.tag_name` | `.release.html_url` |

读取事件元数据的好处是范围明确：通知步骤只处理 GitHub 已经提供的字段，不执行仓库中的脚本，也不需要接触 Pull Request 的代码内容。

### 飞书负责投递

发送步骤先用企业自建应用的 App ID 和 App Secret 获取 `tenant_access_token`，再把消息内容包装成飞书接口需要的 JSON，最后用 `receive_id_type=open_id` 指定接收者。

重要的是，Token 只存在于这次 Job 的运行环境中，不应该写入工作流输出、Issue 评论或普通日志。飞书官方接口文档分别说明了内部应用获取 `tenant_access_token` 和创建消息的请求方式。[飞书：内部应用获取 tenant_access_token](https://open.feishu.cn/document/server-docs/authentication-management/access-token/tenant_access_token_internal)（访问日期：2026-08-07）

## 4. 第一版只申请读取权限

工作流可以明确写出：

```yaml
permissions:
  contents: read
  issues: read
  pull-requests: read
```

这组权限只覆盖读取仓库内容、Issue 和 Pull Request 的场景。第一版不需要：

```yaml
contents: write
issues: write
pull-requests: write
```

后面如果需要评论 Issue 或分配 Reviewer，再按具体动作增加对应权限。GitHub 官方文档说明，`permissions` 可以按工作流或 Job 设置 `GITHUB_TOKEN` 的权限；未使用的写权限不应该为了省事全部打开。[GitHub：Workflow syntax - permissions](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#permissions)（访问日期：2026-08-07）

权限设计不是上线前最后补的一段 YAML，而是功能边界的一部分。只读机器人出问题时，最多是漏通知或多通知；带写权限的机器人出问题，可能会改仓库状态，排错成本和影响范围完全不同。

## 5. 最小配置清单

要跑通这条链路，需要五样东西：

1. 一个 GitHub 测试仓库；
2. 一个飞书企业自建应用；
3. 飞书应用的机器人能力；
4. 发送消息权限 `im:message:send_as_bot`；
5. 三个 GitHub Repository Secret：

```text
FEISHU_APP_ID
FEISHU_APP_SECRET
FEISHU_RECEIVE_ID
```

其中 `FEISHU_RECEIVE_ID` 不是随便找一个 `ou_...` 就行。它必须是当前飞书应用下、当前可用范围内的目标用户 `open_id`。这个值拿错时，App ID 和 App Secret 即使正确，发送消息仍然可能失败。

## 6. 设计完成后应该看到什么

在开始创建应用之前，先写下第一版的验收标准：

- 手动运行 workflow 能收到一条测试私聊；
- 新建 Issue 能自动触发一次运行；
- 飞书消息包含仓库、标题和链接；
- PR 进入 `ready_for_review` 或请求 Review 时能触发通知；
- Release 发布后能触发通知；
- GitHub Actions 没有 checkout 外部 PR 代码；
- workflow 只声明读取权限；
- 日志里没有 App Secret 和 `tenant_access_token`。

如果其中一项无法解释，就不要急着增加更多事件。先把失败范围压缩到最小，再扩展功能。

## 7. 三种常见的错误设计

### 把机器人做成全量转发器

每条评论、每次推送、每个标签都发消息，短期看起来“监控很全”，长期通常会导致接收者关闭通知。事件越多不代表信息越有价值。

### 一开始就给写权限

通知和执行是两种不同风险级别的能力。只读版本尚未证明事件字段、接收者和错误处理都正确时，不应提前开放评论、关闭 Issue 或合并 PR 的权限。

### 在 `pull_request_target` 中运行外部代码

`pull_request_target` 适合读取目标仓库上下文并做受控通知，但不能因为它能触发 Fork PR，就把外部 PR 代码 checkout 后执行。工作流持有飞书 Secret 时，这个边界尤其重要。

## 8. 结论和限制

GitHub 到飞书的监控机器人，最值得先设计清楚的不是消息样式，而是事件范围、数据流和权限边界。第一版只监听 Issue、PR、Release 和手动测试，把事件 JSON 转成短消息，再发给一个明确的用户。

这条链路跑通后，才有资格讨论自动评论、自动打标签和人工确认执行。先把“重要消息能到达”做好，再决定机器人是否拥有“改变仓库状态”的能力。

## 官方来源

- [GitHub：Events that trigger workflows](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows)
- [GitHub：Workflow syntax - permissions](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#permissions)
- [飞书：创建消息](https://open.feishu.cn/document/server-docs/im-v1/message/create)
- [飞书：内部应用获取 tenant_access_token](https://open.feishu.cn/document/server-docs/authentication-management/access-token/tenant_access_token_internal)

---

*本文事实与链接核对日期：2026-08-07。飞书菜单和权限名称可能随平台版本变化，配置前应以当前开放平台文档为准。*
