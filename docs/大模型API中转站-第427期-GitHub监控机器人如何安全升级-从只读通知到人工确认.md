---
title: "GitHub 监控机器人如何安全升级：从只读通知到人工确认"
category: 人工智能
tags:
  - GitHub Actions
  - 自动化安全
  - 权限管理
description: "解释 GitHub 监控机器人从只读通知升级到评论、加标签、分配 Reviewer 和创建 Draft Release 时的权限边界，并设计一条人工确认后执行的流程。"
---

# GitHub 监控机器人如何安全升级：从只读通知到人工确认

只读通知跑通以后，下一步很容易让人产生一个想法：既然机器人已经能识别 Issue 和 PR，能不能顺手评论、加标签、分配 Reviewer，甚至自动合并？技术上这些动作都可以接入，但它们不再是“发一条消息”，而是在改变仓库状态。

这两类能力的风险完全不同。读错一个标题，通常只是通知内容不准确；写错一次评论、标签或 Reviewer，可能影响协作流程；自动合并、关闭 Issue 或删除内容，则可能造成更难恢复的结果。因此升级的重点不是增加多少 API，而是为每一种写操作设置权限、确认、审计和撤销边界。

本文讨论一条稳妥的升级路线：机器人先读取事件和上下文，在飞书里告诉你准备做什么，只有得到明确确认后才执行一个范围有限的动作。本文不提供可直接执行的第二版完整工作流，也不把“设计上可以实现”写成“已经完成并验证”。

## 1. 先区分通知和写入

当前只读版本的动作可以画成这样：

```text
GitHub 事件
    ↓
读取事件 JSON
    ↓
整理标题、链接和仓库名
    ↓
飞书私聊通知
```

这条链路没有 checkout 仓库，也没有执行 Issue 或 PR 中的代码。它需要的核心能力是读取事件和调用飞书消息接口。

升级后的动作则多了一段：

```text
GitHub 事件
    ↓
读取上下文并生成建议
    ↓
飞书私聊展示“准备做什么”
    ↓
人工确认
    ↓
调用 GitHub 写接口
    ↓
把执行结果发回飞书并记录审计信息
```

“人工确认”必须发生在写操作之前，而不是动作执行以后再补发一条说明。确认内容至少应该包括目标仓库、目标 Issue 或 PR、拟执行的动作、具体参数、操作者和确认时间。

## 2. `pull_request_target` 为什么要特别小心

`pull_request_target` 适合读取目标仓库上下文并为 PR 发送受控通知，也能覆盖来自 Fork 的 PR。但它运行在目标仓库的上下文中，可能接触到目标仓库的 `GITHUB_TOKEN`，在配置不当时还可能接触到 Secrets。

所以必须记住三个判断：

1. 能触发 Fork PR，不代表可以运行 Fork PR 的代码；
2. 能读取事件 JSON，不代表可以 checkout 外部提交；
3. 工作流里存在飞书 Secret 时，更不能把外部代码放进同一个执行环境。

当前只读工作流可以读取：

```text
GITHUB_EVENT_PATH
```

然后从 `.pull_request.title` 和 `.pull_request.html_url` 取出标题和链接。它不需要仓库代码，也不需要执行 `npm install`、`pip install`、测试脚本或 PR 中的任意命令。

下面这些写法不应出现在带有飞书 Secret 的 `pull_request_target` 工作流里：

```yaml
- uses: actions/checkout@v4
  with:
    ref: ${{ github.event.pull_request.head.sha }}
- run: npm install
- run: npm test
```

问题不在于 `npm test` 这个命令本身，而在于执行内容来自外部 PR，且工作流上下文可能拥有目标仓库的凭证。需要分析外部代码时，应拆成不接触飞书 Secret 的独立工作流，重新审查 Token 权限、运行器和第三方 Action。

## 3. 权限按动作建立矩阵

不要把权限理解成“机器人开关”。权限应该和具体动作一一对应，先列出动作，再决定是否需要 GitHub 写权限。

| 能力 | GitHub 权限方向 | 是否适合第一步自动执行 | 主要风险 |
| --- | --- | --- | --- |
| 读取 Issue、PR、Release 元数据 | `issues: read`、`pull-requests: read`、必要时读取内容 | 是 | 消息字段不完整或重复 |
| 评论 Issue | `issues: write` | 仅建议确认后执行 | 评论内容不合适、重复评论 |
| 给 Issue 加或移除标签 | `issues: write` | 仅建议确认后执行 | 标签改变筛选和协作流程 |
| 分配 Reviewer | `pull-requests: write` | 仅建议确认后执行 | 打扰错误人员或违反团队规则 |
| 创建 Draft Release | `contents: write` | 仅建议确认后执行 | 版本号、说明或目标分支错误 |
| 自动合并 PR | 通常需要 PR 写权限及分支保护配合 | 不建议作为默认动作 | 直接改变生产代码 |
| 关闭 Issue、删除内容、修改代码 | 对应写权限 | 不应由默认机器人执行 | 结果可能难以恢复 |

第一版只读通知可以继续保持：

```yaml
permissions:
  contents: read
  issues: read
  pull-requests: read
```

只有实现并验证“评论 Issue”这一项时，才在最小范围内考虑 `issues: write`。不要因为下一步可能要处理 PR，就提前把 `contents`、`issues` 和 `pull-requests` 全部改成 `write`。GitHub 官方文档说明，`permissions` 可以设置工作流或 Job 使用的 `GITHUB_TOKEN` 权限：[Workflow syntax - permissions](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#permissions)（访问日期：2026-08-07）。

## 4. 评论 Issue：先起草，再发送

评论是比较适合第一个试点的写操作，因为它的影响范围通常比改代码小，但仍然会进入公开协作记录，不能当成普通日志。

一个可控的评论流程应该是：

```text
Issue opened 或 reopened
        ↓
读取标题、正文、链接和现有标签
        ↓
生成一份“缺少哪些信息”的评论草稿
        ↓
飞书展示目标和完整评论内容
        ↓
人工确认
        ↓
只向这一个 Issue 发布一次评论
        ↓
返回评论链接和 GitHub 响应状态
```

发送前至少检查四件事：

- 目标 Issue 编号与飞书里看到的编号一致；
- 评论正文没有 App Secret、Token、内部路径或不应公开的内容；
- 这条 Issue 最近没有已经发送过的同类评论；
- 确认按钮对应的是当前这次草稿，而不是过期消息。

为了避免重复评论，可以给每份草稿生成一个短的幂等标识，或者在评论末尾加入机器可识别的隐藏标记。具体实现要结合仓库规范决定，不能只依赖“工作流一般不会重复运行”的假设。工作流重试、人工重复点击和事件重新发送都可能造成同一动作再次执行。

## 5. 加标签：动作小，影响不一定小

标签看起来只是一个颜色，但它往往参与项目筛选、看板流转、统计报表和自动化规则。比如 `needs-info` 可能意味着 Issue 暂停处理，`priority-high` 可能会改变排期。因此加标签也需要确认目标和标签值。

推荐把标签动作拆成结构化参数：

```text
仓库：org/repo
目标：Issue #123
动作：添加标签
标签：needs-info
理由：正文缺少复现步骤和运行环境
确认人：当前飞书用户
```

执行前检查标签是否已存在、当前 Issue 是否已经拥有该标签、目标对象是否仍然打开，以及这次判断使用的 Issue 内容是否已经过期。如果用户在飞书里确认后又编辑了 Issue，机器人应该重新读取并重新生成建议，而不是盲目执行旧动作。

撤销边界也要提前写清楚：如果只是误加了一个已有标签，通常可以在确认后移除；但如果标签触发了下游自动化，撤销标签不一定能撤销后续动作。因此审计记录里要保存“为什么加标签”，而不只是保存 API 返回成功。

## 6. 分配 Reviewer：确认人和执行人要分开

给 PR 分配 Reviewer 会直接打扰其他协作者，也可能涉及团队代码所有权规则。机器人可以读取 PR 的目标分支、修改范围和现有 Reviewer，然后给出候选人，但不应该把“候选人”直接当成“执行对象”。

飞书确认消息至少展示：

```text
目标 PR：org/repo#456
目标分支：main
准备分配：alice
依据：该目录由 alice 负责，当前尚未请求 Review
变更状态：PR 仍然处于 Open
```

执行前再次确认 PR 没有关闭、合并或被其他人分配 Reviewer。请求 Review 的动作应当是一次明确的、可记录的操作；不能因为机器人收到了 `review_requested` 事件，就自动继续向更多人派发通知，否则容易形成循环。

这一能力还要考虑权限和组织规则。`pull-requests: write` 只是 Token 层面的必要条件，不代表机器人一定有权替团队决定 Reviewer，也不代表目标用户一定在仓库可分配范围内。实际结果还会受到仓库协作者关系、组织成员关系和分支流程影响，遇到拒绝时应保留 GitHub 返回的错误信息，但不要记录凭证。

## 7. 创建 Draft Release：只准备，不直接发布

Release 会形成版本记录，通常还会影响下载链接、变更日志和下游发布流程。机器人适合先汇总已合并 PR、生成候选版本号和 Draft Release 说明，再由人确认。

推荐的流程是：

1. 根据目标分支和上一个版本读取候选变更；
2. 生成 Draft Release 的标题、Tag 和正文草稿；
3. 在飞书中展示版本号、目标分支、变更范围和完整正文；
4. 人工确认后创建 Draft Release；
5. 把 Release 草稿链接发回飞书；
6. 正式发布仍由人或既有发布流程完成。

这里故意保留 Draft 状态。创建 Draft Release 和发布正式 Release 不是同一个风险等级：前者仍然需要审核，后者可能立即触发部署、通知或下载分发。机器人不应把“生成说明”直接升级成“发布版本”。

如果版本号来自自动推断，也要在飞书确认界面明确显示推断依据和可修改字段。没有足够上下文时，应让人填写 Tag，而不是由脚本猜一个可能冲突的版本号。

## 8. 人工确认消息应该包含什么

人工确认不是在消息末尾加一句“是否执行”就结束了。确认消息应该让人能看懂目标和副作用。至少包含：

| 字段 | 示例含义 |
| --- | --- |
| 目标仓库 | `org/repo` |
| 目标对象 | Issue #123、PR #456 或 Draft Release |
| 准备动作 | 评论、添加标签、分配 Reviewer、创建 Draft Release |
| 动作参数 | 评论全文、标签名、Reviewer、Tag 和正文 |
| 数据时间 | 机器人读取上下文的时间 |
| 失效条件 | 对象已关闭、状态变化或确认超时 |
| 操作结果 | 成功、失败及 GitHub 返回的资源链接 |

确认按钮的处理也要有边界：按钮只能对应一份草稿和一个目标，不能复用为“确认所有待处理任务”。确认事件到达后，服务端应重新检查目标状态，再执行一次动作；如果状态已变化，就要求重新确认。

## 9. Secret、Token 和日志边界

飞书的 `FEISHU_APP_SECRET` 只应该作为 GitHub Secret 注入发送步骤。工作流可以用它请求 `https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal`，但不应把返回的 `tenant_access_token` 写进仓库、Issue、飞书消息或普通日志。

写入 GitHub 的版本还要特别检查第三方 Action。一个 Action 即使看起来只是格式化，也可能读取当前环境变量。带有飞书 Secret 的 Job 应尽量减少第三方步骤，把 Secret 的注入范围限制到真正发送请求的步骤，并固定可信 Action 的版本。

日志中可以保留：

- 事件名、仓库名、Issue 或 PR 编号；
- 请求的动作类型和结果状态；
- GitHub 或飞书返回的非敏感错误码；
- 审计记录 ID 和目标资源链接。

日志中不要保留：

- App Secret、`tenant_access_token` 或完整 Authorization 头；
- 完整请求体中包含的凭证字段；
- 未经确认的内部分析内容；
- 来自外部 PR 的任意脚本输出。

GitHub 官方关于 Secrets 的说明见：[Using secrets in GitHub Actions](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets)（访问日期：2026-08-07）。

## 10. 建议的升级顺序

可以按下面的顺序逐项开放能力：

```text
只读通知
   ↓
飞书展示建议，但不写 GitHub
   ↓
确认后评论一个测试 Issue
   ↓
确认后给 Issue 加一个已有标签
   ↓
确认后请求一个 Reviewer
   ↓
确认后创建 Draft Release
```

每开放一项，都要单独完成一次测试、一次失败测试和一次撤销测试。比如评论功能要验证确认前不会评论、确认后只评论一次、GitHub 拒绝时飞书能收到错误；标签功能要验证目标标签不匹配时不会执行；Draft Release 要验证只创建草稿而不会直接发布。

不建议把多个写动作放进同一个“确认执行”按钮。评论、加标签和分配 Reviewer 的目标、权限和撤销方式不同，拆开后出了问题更容易判断是哪一个动作导致的。

## 11. 结论和限制

GitHub 监控机器人从只读通知升级到写入操作，真正增加的是状态变化和责任边界，而不是几行 API 请求。`issues: write`、`pull-requests: write` 和 `contents: write` 应该随动作逐项申请；`pull_request_target` 工作流不能 checkout 或执行外部 PR 代码；飞书确认后仍然要重新检查 GitHub 对象状态。

第一批适合试点的能力是评论、加标签、请求 Reviewer 和创建 Draft Release。自动合并 PR、关闭 Issue、删除内容、修改代码和发布正式 Release 都不应放进默认机器人。让机器人准备建议、让人确认目标和副作用、让系统留下结果和审计记录，这条路线更适合持续运行。

## 官方来源

- [GitHub：Events that trigger workflows](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows)
- [GitHub：Workflow syntax - permissions](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#permissions)
- [GitHub：Using secrets in GitHub Actions](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets)
- [飞书：创建消息](https://open.feishu.cn/document/server-docs/im-v1/message/create)
- [飞书：内部应用获取 tenant_access_token](https://open.feishu.cn/document/server-docs/authentication-management/access-token/tenant_access_token_internal)

---

*本文事实与链接核对日期：2026-08-07。GitHub 权限、事件和飞书应用菜单可能随平台版本变化，正式配置前应以当前官方文档和后台页面为准。*
