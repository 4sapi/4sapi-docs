---
title: "GitHub Actions 怎么发飞书私聊：完整工作流与三步验证"
category: 人工智能
tags:
  - GitHub Actions
  - 飞书机器人
  - 自动化工作流
description: "从创建 workflow 文件到手动运行、创建真实 Issue 和核对飞书消息，完整拆解 GitHub Actions 发飞书私聊的可复现步骤与排错方法。"
---

# GitHub Actions 怎么发飞书私聊：完整工作流与三步验证

把 GitHub 事件发到飞书，真正容易出错的地方不是“复制一段 YAML”这么简单。工作流文件可能放错目录，Secret 名字可能多一个字符，飞书应用可能拿到了 Token 却没有发送权限，手动运行也可能成功，但真实 Issue 根本没有触发。

这篇只解决一个问题：把 GitHub Actions 工作流部署到测试仓库，并验证它能把手动测试、真实 Issue、事件内容三件事完整送到飞书私聊。前提是你已经准备好飞书企业自建应用，以及下面三个 Repository Secret：`FEISHU_APP_ID`、`FEISHU_APP_SECRET`、`FEISHU_RECEIVE_ID`。如果凭证还没有准备好，先在飞书后台完成应用、机器人、发送权限和当前应用下 `open_id` 的配置。

本文使用一条只读工作流。它读取 GitHub 提供的事件 JSON，整理出标题和链接，然后发送文本消息；它不 checkout 仓库、不执行 PR 代码、不评论 Issue，也不修改仓库状态。最终可观察到的结果是：Actions 有成功记录，飞书机器人收到一条包含仓库、标题和链接的私聊。

## 1. 先确认测试环境

准备一个自己可以修改的 GitHub 测试仓库，例如：

```text
repo-monitor-demo
```

公开仓库和私有仓库都可以。创建后至少确认以下入口存在：

| 入口 | 用途 |
| --- | --- |
| `Code` | 创建 `.github/workflows/repo-monitor.yml` |
| `Issues` | 创建真实 Issue 触发测试 |
| `Actions` | 查看工作流运行记录和失败日志 |
| `Settings` | 保存 Repository Secrets |

如果仓库没有 Issues，进入 `Settings → General → Features` 检查 Issues 是否启用。测试时最好使用自己能控制的仓库和飞书账号，避免第一次运行就把消息发给团队成员。

开始复制代码前，还要确认三个 Secret 已经存在，并且名字完全一致：

```text
FEISHU_APP_ID
FEISHU_APP_SECRET
FEISHU_RECEIVE_ID
```

其中 `FEISHU_RECEIVE_ID` 必须是当前飞书应用下的用户 `open_id`。它不是随便一个格式看起来像 `ou_...` 的字符串。GitHub 官方关于 Secrets 的说明见：[Using secrets in GitHub Actions](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets)（访问日期：2026-08-07）。

## 2. 创建 workflow 文件

在测试仓库首页点击：

```text
Add file → Create new file
```

文件名从第一个点开始完整填写：

```text
.github/workflows/repo-monitor.yml
```

路径中的每一段都有作用：`.github` 是 GitHub 的配置目录，`workflows` 是 Actions 工作流目录，`repo-monitor.yml` 是本次工作流文件名。文件如果放在仓库根目录，或者写成 `.github/workflow` 少了 `s`，GitHub 都不会按预期加载它。

GitHub 的工作流文件由事件触发器、权限、Job 和步骤组成。下面的版本使用四种入口：`issues`、`pull_request_target`、`release` 和 `workflow_dispatch`。前三种对应真实仓库事件，最后一种用来从 Actions 页面手动测试。

## 3. 粘贴完整工作流

下面代码来自完整的只读通知版本。请从 `name: Repo Monitor` 开始复制到最后一行，不要把代码块外的解释文字一起粘进去。YAML 对缩进敏感，尤其是 `on`、`permissions`、`jobs`、`steps` 和 `run` 的层级。

```yaml
name: Repo Monitor

on:
  issues:
    types: [opened, reopened]
  pull_request_target:
    types: [opened, reopened, ready_for_review, review_requested]
  release:
    types: [published]
  workflow_dispatch:

permissions:
  contents: read
  issues: read
  pull-requests: read

jobs:
  notify:
    runs-on: ubuntu-latest
    timeout-minutes: 3

    steps:
      - name: Build notification text
        id: message
        shell: bash
        env:
          EVENT_NAME: ${{ github.event_name }}
          REPOSITORY: ${{ github.repository }}
        run: |
          case "$EVENT_NAME" in
            issues)
              title=$(jq -r '.issue.title' "$GITHUB_EVENT_PATH")
              url=$(jq -r '.issue.html_url' "$GITHUB_EVENT_PATH")
              text="💡 新 Issue｜$REPOSITORY\n$title\n$url"
              ;;
            pull_request_target)
              title=$(jq -r '.pull_request.title' "$GITHUB_EVENT_PATH")
              url=$(jq -r '.pull_request.html_url' "$GITHUB_EVENT_PATH")
              text="🔀 PR 需要关注｜$REPOSITORY\n$title\n$url"
              ;;
            release)
              title=$(jq -r '.release.name // .release.tag_name' "$GITHUB_EVENT_PATH")
              url=$(jq -r '.release.html_url' "$GITHUB_EVENT_PATH")
              text="🚀 新 Release｜$REPOSITORY\n$title\n$url"
              ;;
            *)
              text="🧪 测试成功｜$REPOSITORY 的监控机器人已连通"
              ;;
          esac

          {
            echo 'text<<EOF'
            printf '%b\n' "$text"
            echo 'EOF'
          } >> "$GITHUB_OUTPUT"

      - name: Send as Feishu bot
        shell: bash
        env:
          FEISHU_APP_ID: ${{ secrets.FEISHU_APP_ID }}
          FEISHU_APP_SECRET: ${{ secrets.FEISHU_APP_SECRET }}
          FEISHU_RECEIVE_ID: ${{ secrets.FEISHU_RECEIVE_ID }}
          MESSAGE_TEXT: ${{ steps.message.outputs.text }}
        run: |
          for name in FEISHU_APP_ID FEISHU_APP_SECRET FEISHU_RECEIVE_ID; do
            if [ -z "${!name}" ]; then
              echo "缺少 $name，请先添加 Repository secret。"
              exit 1
            fi
          done

          token_body=$(jq -n \
            --arg app_id "$FEISHU_APP_ID" \
            --arg app_secret "$FEISHU_APP_SECRET" \
            '{app_id:$app_id,app_secret:$app_secret}')

          token_response=$(curl --silent --show-error \
            --retry 3 \
            --retry-all-errors \
            --header "Content-Type: application/json" \
            --data "$token_body" \
            "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal")

          if [ "$(jq -r '.code // -1' <<< "$token_response")" != "0" ]; then
            echo "获取 tenant_access_token 失败：$(jq -r '.msg // "unknown error"' <<< "$token_response")"
            exit 1
          fi

          tenant_token=$(jq -r '.tenant_access_token // empty' <<< "$token_response")
          if [ -z "$tenant_token" ]; then
            echo "飞书没有返回 tenant_access_token。"
            exit 1
          fi

          content=$(jq -nc --arg text "$MESSAGE_TEXT" '{text:$text}')
          message_body=$(jq -n \
            --arg receive_id "$FEISHU_RECEIVE_ID" \
            --arg content "$content" \
            '{receive_id:$receive_id,msg_type:"text",content:$content}')

          http_code=$(curl --silent --show-error \
            --output response.json \
            --write-out "%{http_code}" \
            --retry 3 \
            --retry-all-errors \
            --header "Authorization: Bearer $tenant_token" \
            --header "Content-Type: application/json" \
            --data "$message_body" \
            "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=open_id")

          if [ "$http_code" != "200" ]; then
            echo "飞书发送接口返回 HTTP $http_code。"
            exit 1
          fi

          if [ "$(jq -r '.code // -1' response.json)" != "0" ]; then
            echo "机器人发送失败：$(jq -r '.msg // "unknown error"' response.json)"
            exit 1
          fi

          echo "飞书机器人通知发送成功。"
```

这段代码没有 checkout 步骤，运行环境只使用 GitHub Actions 自带的事件文件、Bash、`jq` 和 `curl`。GitHub 在每次工作流运行时把触发事件保存到 `GITHUB_EVENT_PATH`，脚本根据事件名从 JSON 中读取对应字段：Issue 读取 `.issue.title` 和 `.issue.html_url`，PR 读取 `.pull_request.title` 和 `.pull_request.html_url`，Release 读取 `.release.name` 或 `.release.tag_name`。

## 4. 代码到底分成哪两段

第一步 `Build notification text` 只负责组装消息。`EVENT_NAME` 来自 `${{ github.event_name }}`，`REPOSITORY` 来自 `${{ github.repository }}`。脚本通过 `case` 判断当前事件，再读取 `$GITHUB_EVENT_PATH`。最后使用 `$GITHUB_OUTPUT` 把多行文本交给下一步。

这里使用 `printf '%b\n'`，是因为消息模板中的 `\n` 需要在输出时变成换行。输出变量使用 GitHub Actions 的多行输出格式，不要改成普通的 `echo` 拼接，否则标题或链接里出现特殊字符时更难排查。

第二步 `Send as Feishu bot` 负责认证和发送：

1. 先检查三个 Secret 是否为空；
2. 用 `FEISHU_APP_ID` 和 `FEISHU_APP_SECRET` 请求 `tenant_access_token`；
3. 检查响应里的 `code`，再取出 Token；
4. 用 `jq` 把文本包成飞书要求的消息 JSON；
5. 调用 `https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=open_id`；
6. 同时检查 HTTP 状态码和飞书响应体中的 `code`。

飞书创建消息接口要求外层请求提供 `receive_id`、`msg_type` 和 `content`。文本消息的 `content` 本身还是一个 JSON 字符串，所以代码先构造 `{"text":"..."}`，再把它作为外层 JSON 的字符串字段传入。多包一层是接口格式要求，不是多余的转义。

## 5. 第一次验证：手动 Run workflow

点击 GitHub 编辑器右上角的 `Commit changes…` 提交文件。测试仓库可以直接提交到默认分支，正式团队仓库建议走分支和 PR 审核。

提交后依次打开：

```text
Actions → Repo Monitor → Run workflow → Run workflow
```

等待运行结束，至少检查三处：

| 检查位置 | 预期结果 |
| --- | --- |
| Actions 工作流列表 | 出现一条绿色成功记录 |
| `notify` Job | 两个步骤都成功，没有 Secret 缺失提示 |
| 飞书私聊 | 收到“测试成功”文本 |

手动运行走的是 `workflow_dispatch` 分支。由于脚本的 `case` 没有专门处理这个事件，它会进入 `*`，发送：

```text
🧪 测试成功｜仓库名 的监控机器人已连通
```

这一步验证的是凭证、应用可用范围、机器人发送权限和消息接口，不代表 `issues` 触发器已经验证。手动运行成功但创建 Issue 没有消息，应该检查事件触发和工作流分支，而不是立即重新生成 App Secret。

## 6. 第二次验证：创建真实 Issue

打开测试仓库的：

```text
Issues → New issue
```

标题可以填写：

```text
测试：GitHub Issue 是否会触发飞书监控
```

正文写入可识别的测试标记：

```text
这是一条用于教程实测的 Issue。

验证目标：
- 创建 Issue 后自动触发 GitHub Actions
- 飞书应用机器人收到私聊通知
```

点击 `Submit new issue` 后，不要再点击 `Run workflow`。这次只观察真实事件是否自动启动工作流。进入 `Actions → Repo Monitor`，打开刚出现的运行记录，检查触发来源是 Issue、事件名称是 `issues`、状态为 Success，且 `notify` Job 为绿色。

成功后，飞书消息应当包含三部分：仓库名、Issue 标题和 Issue 链接，形式类似：

```text
💡 新 Issue｜adrianpunk/repo-monitor-demo
测试：GitHub Issue 是否会触发飞书监控
https://github.com/adrianpunk/repo-monitor-demo/issues/1
```

仓库名、Issue 编号和链接会根据你的测试仓库变化。不要只看飞书有没有收到一条消息；还要把 Actions 运行记录与消息内容对上，确认触发的是这一次 Issue，而不是之前的手动测试。

## 7. 第三次验证：检查内容和边界

如果 Issue 已经能触发，还应该补做一次字段验证。创建或重新打开 Issue 后，核对消息中的标题是否与 GitHub 页面完全一致，链接是否能直接打开对应 Issue，仓库名是否来自当前仓库。

随后可以在测试仓库中分别验证其他入口：

| 操作 | 触发事件 | 消息前缀 |
| --- | --- | --- |
| 创建或重新打开 Issue | `issues` | `💡 新 Issue` |
| 创建、重新打开或请求 Review 的 PR | `pull_request_target` | `🔀 PR 需要关注` |
| 发布一个 Release | `release` | `🚀 新 Release` |
| 点击手动运行 | `workflow_dispatch` | `🧪 测试成功` |

第一次测试不建议连续创建很多事件。每次只做一个动作，等 Actions 结束后再看飞书消息，这样能知道哪一个事件对应哪一次运行。也不要把“收到消息”理解成“所有 GitHub 状态都被同步”：当前版本只发送标题和链接，不读取评论、不汇总历史、不保存事件去重状态。

## 8. 报错时按链路排查

| 现象 | 优先检查 |
| --- | --- |
| 工作流没有出现在 Actions | 文件是否位于 `.github/workflows/repo-monitor.yml`，是否已经提交到默认分支 |
| 提示缺少 `FEISHU_...` | Secret 名字是否完全一致，是否保存到了当前仓库而不是其他仓库 |
| 获取 `tenant_access_token` 失败 | App ID 和 App Secret 是否来自同一个应用，应用版本是否已经发布 |
| `permission denied` | 是否开通 `im:message:send_as_bot`，权限变更是否已发布 |
| `Bot has NO availability to this user` | 接收者是否在应用可用范围内，是否与机器人建立过会话 |
| `receive_id invalid` | `FEISHU_RECEIVE_ID` 是否是当前应用下的 `open_id`，URL 是否带 `receive_id_type=open_id` |
| 手动运行成功，Issue 不触发 | Issue 是否真的执行了 `opened` 或 `reopened`，工作流是否在默认分支 |
| Actions 成功但没有消息 | 检查机器人会话、接收者 ID、飞书响应体和应用发布状态 |

排错时只输出状态码、错误码和错误消息，不要打印完整 `tenant_access_token`、App Secret 或完整请求体。当前工作流把飞书响应保存为运行目录里的 `response.json`，只用它检查结果；不要为了方便把这个文件提交回仓库。

## 9. 适用边界和安全提醒

工作流使用 `pull_request_target`，可以接收目标仓库上下文中的 PR 事件，包括来自 Fork 的 PR。这个事件类型本身不等于“可以安全运行 PR 代码”。当前脚本只读取 GitHub 放在 `GITHUB_EVENT_PATH` 的元数据，没有 checkout，也没有执行外部分支中的脚本。

`permissions` 也必须保持只读：

```yaml
permissions:
  contents: read
  issues: read
  pull-requests: read
```

不要为了排查消息问题直接改成 `contents: write`、`issues: write` 或 `pull-requests: write`。通知链路是否打通，与机器人能否改变仓库状态，是两个不同的问题。GitHub 对工作流权限的说明见：[Workflow syntax - permissions](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#permissions)（访问日期：2026-08-07）；事件触发器说明见：[Events that trigger workflows](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows)（访问日期：2026-08-07）。

## 10. 结论

三次验证分别回答三个问题：手动 Run workflow 证明认证链路能发消息；真实 Issue 证明 GitHub 事件能触发；字段和链接核对证明消息内容来自正确的事件。只有三步都通过，才算完成一个可观察、可排错的 GitHub 到飞书通知链路。

这个版本有意把能力限制在只读通知。它不自动评论、不自动打标签、不合并 PR，也不修改代码。等事件、凭证和消息格式稳定后，再为具体写操作设计人工确认、权限升级和审计记录，升级过程会更容易控制。

## 官方来源

- [GitHub：Events that trigger workflows](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows)
- [GitHub：Workflow syntax](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax)
- [GitHub：Workflow syntax - permissions](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#permissions)
- [GitHub：Using secrets in GitHub Actions](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets)
- [飞书：创建消息](https://open.feishu.cn/document/server-docs/im-v1/message/create)
- [飞书：内部应用获取 tenant_access_token](https://open.feishu.cn/document/server-docs/authentication-management/access-token/tenant_access_token_internal)

---

*本文事实与链接核对日期：2026-08-07。GitHub 和飞书的后台菜单可能随版本变化，配置时应以当前官方文档和页面为准。*
