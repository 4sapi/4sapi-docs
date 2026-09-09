---
title: "Codex 接 GitHub：从 0 搭一个会私聊你的仓库监控机器人"
category: 人工智能
tags:
  - GitHub Actions
  - 飞书机器人
  - 自动化工作流
description: "用 GitHub Actions 监听 Issue、PR 和 Release，再通过飞书企业自建应用机器人私聊指定用户；文章给出完整配置、代码、验证步骤和权限边界。"
---

# Codex 接 GitHub：从 0 搭一个会私聊你的仓库监控机器人

GitHub 的通知有一个很实际的问题：重要消息和普通消息混在一起。

Issue 开了半天，刷仓库时才发现；PR 已经等着 Review，群里没人提，进度就卡在那里。可如果把所有提醒全部打开，评论、机器人更新和代码同步又会迅速把真正需要处理的事情淹掉。

我这次搭的是一个很克制的第一版：GitHub 出现需要关注的事件，GitHub Actions 自动整理消息，再由飞书应用机器人私聊我。它只读 GitHub，不评论、不打标签、不分配负责人，也不修改代码。

最后跑通的链路是：

```text
GitHub Issue / PR / Release
        ↓
GitHub Actions 自动运行
        ↓
飞书企业自建应用机器人
        ↓
机器人私聊指定用户
```

本文不要求你会写程序。你需要做的是创建应用、开最小权限、保存三个 Secret、复制工作流文件，然后按步骤验证。遇到报错时，重点看失败发生在 GitHub 事件、凭证、飞书权限还是消息请求，而不是一上来重写全部代码。

## 1. 先确定第一版监控什么

第一版只处理四类事件：

| 事件 | 触发条件 | 飞书消息内容 |
| --- | --- | --- |
| Issue | 新建、重新打开 | 标题和链接 |
| PR | 新建、重新打开、进入可审核状态、请求 Review | 标题和链接 |
| Release | 正式发布 | Release 名称或 tag 和链接 |
| 手动测试 | 在 Actions 页面手动运行 | 一条测试成功消息 |

GitHub 官方文档把 workflow trigger 定义为触发工作流运行的仓库事件，并允许通过 `types` 限制活动类型。本文只监听需要人工注意的动作，不监听每次评论、标签变化和代码推送。[GitHub：Events that trigger workflows](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows)（访问日期：2026-08-07）

这个取舍很重要。通知机器人不是越勤快越好。每一次提醒都应该对应一个动作：打开 Issue、安排 Review、检查 Release，或者确认机器人链路仍然正常。

## 2. 最小配置需要什么

整个机器人只需要五样东西：

1. 一个 GitHub 测试仓库；
2. 一个飞书企业自建应用；
3. 这个应用的机器人能力；
4. 发送消息权限 `im:message:send_as_bot`；
5. 三个 GitHub Repository Secret：

```text
FEISHU_APP_ID
FEISHU_APP_SECRET
FEISHU_RECEIVE_ID
```

GitHub 工作流只申请读取权限：

```yaml
permissions:
  contents: read
  issues: read
  pull-requests: read
```

这个版本不需要 `contents: write`、`issues: write` 或 `pull-requests: write`。如果后续要让机器人自动评论或加标签，再针对具体动作增加权限，不要把 `write-all` 当成省事的默认值。GitHub 官方 workflow syntax 文档说明，`permissions` 可以按工作流或 Job 限制 `GITHUB_TOKEN` 的权限范围。[GitHub：Workflow syntax - permissions](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#permissions)（访问日期：2026-08-07）

## 3. 飞书应用只做三件事

飞书这边不用先研究所有开放接口。第一版只需要让一个企业自建应用具备“以机器人身份发消息”的能力。

### 第一步：创建企业自建应用

在飞书开发者后台创建企业自建应用，名称可以填“GitHub 仓库监控”。

### 第二步：添加机器人并开权限

给应用添加“机器人”能力，在权限管理中开通：

```text
im:message:send_as_bot
```

创建并发布一个版本。可用范围先只选自己，这样第一版即使配置出错，也不会把测试通知发给其他成员。

### 第三步：拿到三个值

从应用的“凭证与基础信息”页面复制：

- App ID；
- App Secret。

再通过飞书“发送消息”API 调试台选择自己，把 `receive_id_type` 设置为 `open_id`，复制你的 `ou_...` 值。

这里有一个很容易踩的坑：`open_id` 属于具体应用。同一个人，在不同飞书应用下可能拿到不同的 `open_id`。不能从另一套应用或另一套 CLI 配置里复制一个看起来格式相同的 ID。

App Secret 只保存到 GitHub Secret，不要发到聊天里，不要写进工作流代码，也不要在截图中露出来。飞书发消息接口使用 `receive_id` 和 `receive_id_type` 指定接收者，具体字段和认证方式以当前官方接口文档为准：[飞书：创建消息](https://open.feishu.cn/document/server-docs/im-v1/message/create)（访问日期：2026-08-07）。应用凭证换取 `tenant_access_token` 的接口说明见：[飞书：内部应用获取 tenant_access_token](https://open.feishu.cn/document/server-docs/authentication-management/access-token/tenant_access_token_internal)（访问日期：2026-08-07）。

飞书部分的完成标准很简单：机器人已经发布，你手里有 App ID、App Secret 和当前应用下的 `open_id`。

## 4. GitHub 第一步：建测试仓库

打开 GitHub 新建仓库页面，仓库名填：

```text
repo-monitor-demo
```

公开或私有都可以。第一次建议勾选 `Add a README file`，这样创建后会直接有 `main` 分支。

创建后先认清顶部入口：

- **Code**：存放工作流文件；
- **Issues**：创建真实测试事件；
- **Actions**：查看工作流是否运行成功；
- **Settings**：保存飞书凭证。

如果找不到 Issues，进入 `Settings → General → Features`，确认 Issues 已开启。第一版不需要修改其他 Workflow permissions，因为工作流文件会自己声明读取权限。

完成标准：仓库有默认分支，并且能看到 Issues、Actions 和 Settings。

## 5. GitHub 第二步：保存三个 Secret

进入测试仓库，按这条路径打开凭证页：

```text
Settings → Secrets and variables → Actions → New repository secret
```

依次创建三个 Secret：

```text
FEISHU_APP_ID
FEISHU_APP_SECRET
FEISHU_RECEIVE_ID
```

它们对应的值分别是 App ID、App Secret 和当前飞书应用下的 `ou_...`。

每一项都要单独点击一次 `New repository secret`，在 `Name` 填变量名，在 `Secret` 粘贴对应值，再点击 `Add secret`。名字必须完全一致，不要加引号，也不要在前后多粘贴空格。

完成标准是 Secrets 列表里能看到三个名称，但看不到具体值。GitHub 官方文档说明，Repository Secret 可以从仓库的 `Settings → Secrets and variables → Actions` 创建；Secret 保存后不应被当成普通配置回读。[GitHub：Using secrets in GitHub Actions](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets)（访问日期：2026-08-07）

## 6. GitHub 第三步：创建工作流文件

回到仓库首页，点击：

```text
Add file → Create new file
```

文件名填写：

```text
.github/workflows/repo-monitor.yml
```

这个路径要从第一个点开始完整输入。`.github` 是 GitHub 配置目录，`workflows` 用来存放自动化任务，`.yml` 是工作流文件格式。GitHub 官方文档要求 workflow 文件放在仓库的 `.github/workflows` 目录中。[GitHub：Workflow syntax](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax)（访问日期：2026-08-07）

完成标准：编辑器顶部显示的路径正好是 `.github/workflows/repo-monitor.yml`。

## 7. 粘贴完整代码

把下面代码从 `name: Repo Monitor` 开始完整复制，粘贴进 GitHub 编辑器。不要手动调整空格；YAML 对缩进敏感。

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

这份代码分成两段。第一段读取 GitHub Actions 提供的事件 JSON，把 Issue、PR 和 Release 整理成一条文本；第二段先用 App ID 和 App Secret 换取短期 `tenant_access_token`，再调用飞书发送消息接口。

代码没有 checkout 仓库，也没有执行 Issue 或 PR 中的代码，因此第一版只处理事件元数据。不要为了“顺便读取代码”给这条工作流增加 checkout 和脚本执行步骤，尤其不要在 `pull_request_target` 工作流里运行来自 Fork PR 的不可信代码。

## 8. GitHub 第四步：提交并手动测试

检查文件名和代码完整后，点击右上角的 `Commit changes…`。测试仓库可以直接提交到 `main`；团队正式仓库则应新建分支并走 PR 审核。

提交说明可以写成：

```text
Add GitHub repository monitor
```

提交后回到 Code 页，确认能看到 `.github/workflows/repo-monitor.yml`；再打开 Actions，左侧应该出现 `Repo Monitor`。

接着进入 `Actions → Repo Monitor → Run workflow → Run workflow`。等待工作流完成并刷新页面，你应该同时看到：

- GitHub Actions 运行记录是绿色对勾；
- 飞书应用机器人私聊你一条“测试成功”消息。

手动测试只负责检查 App ID、App Secret、open_id、应用可用范围和发送权限是否打通，不代表 Issue 事件已经验证完成。

## 9. GitHub 第五步：创建真实 Issue 验证触发器

打开：

```text
Issues → New issue
```

标题可以使用：

```text
测试：GitHub Issue 是否会触发飞书监控
```

正文可以写：

```text
这是一条用于教程实测的 Issue。

验证目标：
- 创建 Issue 后自动触发 GitHub Actions
- 飞书应用机器人收到私聊通知
```

点击 `Submit new issue` 后，不要再手动点 `Run workflow`。这次要验证真实 Issue 能否自动触发。

打开 `Actions → Repo Monitor`，点进刚出现的运行记录，依次核对：

1. 运行记录的触发来源是 Issue；
2. 事件名称是 `issues`；
3. 状态是 Success；
4. `notify` 任务是绿色对勾。

Actions 成功后，打开飞书应用机器人私聊。收到的消息应当包含仓库名、Issue 标题和 Issue 链接，例如：

```text
💡 新 Issue｜adrianpunk/repo-monitor-demo
测试：GitHub Issue 是否会触发飞书监控
https://github.com/adrianpunk/repo-monitor-demo/issues/1
```

如果你使用的是自己的测试仓库，消息中的仓库名、标题和链接会替换成自己的值。看到 GitHub Issue、Actions 运行记录和飞书消息三处内容一致，整条链路才算验证完成。

## 10. 报错时按失败环节排查

| 现象 | 优先检查 |
| --- | --- |
| 缺少 `FEISHU_...` | Secret 名字是否完全一致，值是否有前后空格 |
| 获取 `tenant_access_token` 失败 | App ID 和 App Secret 是否来自同一个飞书应用；应用版本是否已发布 |
| `Bot has NO availability to this user` | 自己是否在应用可用范围内，是否已经和机器人建立会话 |
| `receive_id invalid` | `FEISHU_RECEIVE_ID` 是否是当前应用下的 `open_id` |
| `99991672` 或 `permission denied` | `im:message:send_as_bot` 是否已开通，应用版本是否发布 |
| 手动运行成功，Issue 不触发 | 工作流是否已经提交到默认分支，Issue 是否真的使用 `opened` 或 `reopened` 事件 |
| Actions 绿色但没收到消息 | 检查飞书机器人会话、可用范围、消息接口响应体和接收者 ID |

不要把 App Secret 粘贴到 Actions 日志里排错。工作流目前只输出错误信息和最终成功提示；如果需要进一步调试，增加字段状态和 HTTP 状态码即可，不要打印完整 Token 或请求体中的凭证。

## 11. 三个必须提前知道的安全边界

### 11.1 `pull_request_target` 不是普通的 `pull_request`

本文使用 `pull_request_target`，目的是让来自 Fork 的 PR 也能触发一条只读通知。但它运行在目标仓库的上下文中，工作流可能接触到更高权限的 Token 和 Secret。

因此，这条工作流必须保持当前形态：只读取 `GITHUB_EVENT_PATH` 中的事件元数据，不 checkout、不运行外部 PR 代码、不执行 PR 中的脚本。如果以后加入代码分析，应该单独拆出不接触飞书 Secret 的工作流，并重新审查权限。

### 11.2 权限按动作逐项增加

只读通知保持：

```yaml
permissions:
  contents: read
  issues: read
  pull-requests: read
```

下一版如果要评论 Issue，再考虑 `issues: write`；如果要操作 PR，再考虑对应的 `pull-requests: write`。自动合并 PR、关闭 Issue、删除内容和修改代码，不应该因为机器人“已经能发消息”就顺手打开。

### 11.3 `open_id` 和 Secret 都有作用域

App Secret 属于具体飞书应用，`open_id` 也属于具体应用。GitHub Secret 属于具体仓库或环境。三个地方任何一个拿错，都会出现“看起来格式正确，但请求失败”的情况。

## 12. 下一版应该先做确认，再做写入

只读版本跑稳后，可以把机器人升级成“人工确认后执行”的版本：

```text
新的 Issue / PR / Release
        ↓
机器人读取上下文，生成处理建议
        ↓
飞书私聊：准备执行什么、为什么、会改动哪里
        ↓
我选择：确认执行 / 修改建议 / 跳过
        ↓
确认后写回 GitHub，并把结果发回飞书
```

适合逐项增加的动作包括：

- Issue 缺少复现步骤：起草追问内容，确认后评论并添加 `needs-info`；
- PR 等待 Review：总结改动、测试结果和风险，确认后分配 Reviewer 或发表评论；
- 准备发 Release：汇总已合并 PR，确认后创建 Draft Release。

这一步会增加 GitHub 写权限、飞书交互卡片回调和一个接收确认事件的服务。正确顺序是先把只读版本跑稳，再逐项开放一种写入动作，并为每种动作保留审计日志和撤销方式。

## 13. 结论和限制

整个过程可以压缩成两句话：

```text
监控什么：Issue / PR / Release
通知到哪：飞书应用机器人私聊指定用户
```

第一版的价值不是替你处理 GitHub，而是先解决“重要事件没人看见”的问题。GitHub Actions 负责监听事件和整理上下文，飞书机器人负责把消息送到一个明确的人手里；权限和 Secret 则把这条链路限制在只读通知范围内。

等只读链路稳定后，再考虑人工确认、自动评论和 Draft Release。机器人可以负责准备和执行，人保留最后一次确认；这比一开始给它仓库写权限，更容易定位问题，也更容易撤销。

## 官方来源

- [GitHub：Events that trigger workflows](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows)
- [GitHub：Workflow syntax](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax)
- [GitHub：Using secrets in GitHub Actions](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets)
- [飞书：创建消息](https://open.feishu.cn/document/server-docs/im-v1/message/create)
- [飞书：内部应用获取 tenant_access_token](https://open.feishu.cn/document/server-docs/authentication-management/access-token/tenant_access_token_internal)

---

*本文中的工作流和消息字段按 2026-08-07 可访问的官方文档与素材中的实测链路整理；飞书菜单、权限名称和接口限制可能随平台版本变化，发布前应再次核对当前后台。*
