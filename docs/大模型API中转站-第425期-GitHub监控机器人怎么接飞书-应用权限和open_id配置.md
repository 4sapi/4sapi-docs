---
title: "GitHub 监控机器人怎么接飞书：应用、权限和 open_id 配置"
category: 人工智能
tags:
  - 飞书开放平台
  - GitHub Actions
  - 凭证管理
description: "手把手配置飞书企业自建应用、机器人能力、发送消息权限和 open_id，再把 App ID、App Secret 与接收者 ID 安全保存到 GitHub Secrets。"
---

# GitHub 监控机器人怎么接飞书：应用、权限和 open_id 配置

GitHub 监控机器人最容易卡住的地方，往往不是 YAML，而是凭证。

飞书应用没有发布，机器人找不到；权限开了但版本没有生效，接口返回 permission denied；`open_id` 是从另一套应用里复制的，格式看起来没问题，消息却始终发不出去。还有一种更危险的情况：为了排错，把 App Secret 直接粘进工作流或日志里。

本文只解决配置问题：创建一个能够私聊指定用户的飞书企业自建应用，并把三个必要值安全地交给 GitHub Actions。完成后，你还不会自动监控 GitHub，但后续工作流已经有了可用的凭证入口。

## 1. 你最终要准备三个值

GitHub Actions 需要读取三个 Repository Secret：

```text
FEISHU_APP_ID
FEISHU_APP_SECRET
FEISHU_RECEIVE_ID
```

它们的含义分别是：

| Secret 名称 | 对应值 | 来源 |
| --- | --- | --- |
| `FEISHU_APP_ID` | 应用标识 | 飞书应用“凭证与基础信息” |
| `FEISHU_APP_SECRET` | 应用密钥 | 飞书应用“凭证与基础信息” |
| `FEISHU_RECEIVE_ID` | 目标用户的 `open_id` | 当前应用的发送消息调试台 |

三个值必须属于同一个飞书应用。尤其是 `FEISHU_RECEIVE_ID`，不能因为它看起来像 `ou_...` 就从另一个应用复制。

## 2. 创建企业自建应用

### 第一步：创建应用

打开飞书开发者后台，创建一个企业自建应用。名称可以填：

```text
GitHub 仓库监控
```

应用名称只是为了在后台和飞书客户端里识别它，不影响 GitHub Actions 的请求。创建后进入应用详情页，先不要急着复制 Secret，按顺序完成机器人能力和权限配置。

### 第二步：添加机器人能力

在应用能力中添加“机器人”。发布后，这个应用才会以机器人身份出现在飞书里。

第一版建议把可用范围只设为自己。这样做有两个好处：

- 测试消息不会误发给团队成员；
- 出现问题时，接收者范围更容易排查。

发布应用后，在飞书搜索这个机器人。如果搜索不到，先检查应用是否已经发布、自己是否在可用范围，以及是否已经和机器人建立会话。

### 第三步：开通发送消息权限

在权限管理中开通：

```text
im:message:send_as_bot
```

这项权限用于让应用以机器人身份发送消息。开完权限后，创建并发布一个版本；只在权限页面勾选而不发布，运行时可能仍然使用旧版本权限。

飞书“创建消息”接口要求请求中提供接收者、消息类型和消息内容。当前工作流使用文本消息，并通过 `receive_id_type=open_id` 指定接收者。[飞书：创建消息](https://open.feishu.cn/document/server-docs/im-v1/message/create)（访问日期：2026-08-07）

## 3. 复制 App ID 和 App Secret

在应用详情的“凭证与基础信息”页面拿到：

- App ID；
- App Secret。

这两个值是一对。不要从一个应用复制 App ID，再从另一个应用复制 App Secret。

App Secret 的处理规则只有三条：

1. 只保存到 GitHub Secret 或密码管理工具；
2. 不写进 `.yml` 文件、Issue、README 和截图；
3. 排错时不输出完整请求体、完整 Token 或 Secret。

工作流只需要用 App ID 和 App Secret 请求飞书的 `tenant_access_token`。这个 Token 是给接口调用使用的短期凭证，不需要人工复制到 GitHub，也不应该写进仓库文件。飞书官方文档提供了内部应用获取 `tenant_access_token` 的接口说明：[飞书：内部应用获取 tenant_access_token](https://open.feishu.cn/document/server-docs/authentication-management/access-token/tenant_access_token_internal)（访问日期：2026-08-07）。

## 4. 获取当前应用下的 open_id

在飞书开放平台的“发送消息”API 调试台中：

1. 选择当前刚创建的应用；
2. 把 `receive_id_type` 设置为 `open_id`；
3. 在接收者选择器中选择自己；
4. 复制返回的 `ou_...` 值。

不要从聊天记录、其他应用、旧项目或网上示例里复制 `open_id`。它和应用的可用范围有关，必须用当前应用的调试台获取。

可以用下面的判断确认是否拿对：

- 这个用户在当前应用的可用范围内；
- 飞书客户端能搜索到当前机器人；
- 你已经和当前机器人建立过会话；
- `open_id` 来自当前应用的调试台，而不是另一个应用。

## 5. 把三个值保存到 GitHub Secrets

进入 GitHub 测试仓库，打开：

```text
Settings → Secrets and variables → Actions → New repository secret
```

依次创建：

```text
FEISHU_APP_ID
FEISHU_APP_SECRET
FEISHU_RECEIVE_ID
```

每项 Secret 的操作顺序都是：

1. 点击 `New repository secret`；
2. 在 `Name` 填变量名；
3. 在 `Secret` 粘贴对应值；
4. 点击 `Add secret`。

名字必须完全一致，不要加引号，不要使用中文标点，也不要在前后多粘贴空格。GitHub 官方文档的创建路径也是仓库 `Settings` 下的 `Secrets and variables → Actions` 页面。[GitHub：Using secrets in GitHub Actions](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets)（访问日期：2026-08-07）

完成后，Secrets 列表只能看到三个名称，看不到具体值。这是正常的。GitHub 不会把保存后的 Secret 当成普通配置回读；如果怀疑填错，直接更新该 Secret，不要试图把它打印到日志里。

## 6. 用一张图检查认证链路

工作流真正运行时，认证过程是：

```text
FEISHU_APP_ID + FEISHU_APP_SECRET
              ↓
飞书 tenant_access_token/internal
              ↓
tenant_access_token
              ↓
receive_id + msg_type + content
              ↓
飞书 im/v1/messages?receive_id_type=open_id
```

消息内容不是一段随便拼接的字符串。当前工作流会先把通知文本编码成 JSON 字符串，再放进消息请求的 `content` 字段：

```json
{
  "receive_id": "ou_xxx",
  "msg_type": "text",
  "content": "{\"text\":\"测试成功\"}"
}
```

这也是很多初学者会遇到的错误：外层请求是 JSON，但飞书文本消息的 `content` 仍然需要是一个 JSON 字符串。完整的 `jq` 处理会放在工作流文章中，这里先记住数据层级。

## 7. 配置完成后的预检清单

在开始排查 GitHub Actions 前，先逐项确认：

| 检查项 | 通过标准 |
| --- | --- |
| 应用 | 企业自建应用已经创建 |
| 机器人 | 已添加并发布 |
| 可用范围 | 至少包含自己 |
| 权限 | 已开通 `im:message:send_as_bot` |
| 版本 | 权限变更已经发布到一个版本 |
| App ID | 来自当前应用凭证页 |
| App Secret | 来自同一个应用凭证页 |
| open_id | 来自当前应用发送消息调试台 |
| GitHub Secrets | 三个名称完全一致，值没有明文写入代码 |

如果这张表还没有完成，就不要先怀疑 GitHub YAML。认证链路没有准备好时，Actions 红叉只是结果，不是根因。

## 8. 常见错误怎么判断

### `permission denied`

优先检查 `im:message:send_as_bot` 是否开通、权限变更是否发布，以及工作流使用的 App ID 是否来自当前应用。

### `Bot has NO availability to this user`

检查接收者是否在应用可用范围，是否已经发布机器人，以及飞书客户端是否建立过会话。

### `receive_id invalid`

检查 `FEISHU_RECEIVE_ID` 是否是当前应用下的 `open_id`，以及请求 URL 是否带有：

```text
receive_id_type=open_id
```

### 能拿到 Token 但消息发不出去

这说明 App ID、App Secret 和认证接口大概率已经通了，问题集中在发送权限、接收者、消息体格式或应用可用范围。先看发送接口返回的 HTTP 状态码和响应体中的 `code`、`msg`，不要打印完整 Token。

## 9. 结论和限制

GitHub 监控机器人接飞书，最容易出错的不是“怎么复制代码”，而是三个值的来源和作用域：App ID、App Secret 与当前应用下的 `open_id` 必须属于同一套应用配置。

先创建并发布机器人，再开通发送权限，最后把三个值分别保存到 GitHub Secrets。完成这一步后，才进入 GitHub Actions 工作流配置。凭证能安全传到运行环境，且日志不会泄露 Secret，才算配置合格。

## 官方来源

- [飞书：创建消息](https://open.feishu.cn/document/server-docs/im-v1/message/create)
- [飞书：内部应用获取 tenant_access_token](https://open.feishu.cn/document/server-docs/authentication-management/access-token/tenant_access_token_internal)
- [GitHub：Using secrets in GitHub Actions](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets)

---

*本文事实与链接核对日期：2026-08-07。App Secret、open_id 和飞书菜单会因应用与平台版本变化而不同，发布前应以当前后台和官方接口文档为准。*
