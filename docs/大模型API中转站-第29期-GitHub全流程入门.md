---
title: "【大模型API中转站】第29期 GitHub全流程 | 注册到部署一篇通"
category: 人工智能
tags:
  - 大模型API中转站
  - GitHub
  - Git
  - Issue
  - Fork
  - Pull Request
  - GitHub Pages
  - Codex
  - Claude
  - 4SAPI
description: "从 GitHub 注册、双重认证、搜索、建仓库、Issue、本地 Git、部署、Fork、Star、Pull、Push 讲起，补上 AI Coding Agent 自动化协作前必须掌握的基础。"
---

# 【大模型API中转站】第29期 GitHub全流程 | 注册到部署一篇通

本文是【大模型API中转站】系列的第29篇。本系列致力于用最低的成本、最清晰的方法，帮你打通多模型API的任督二脉。建议先收藏，随用随查。

前面几篇我们一直在聊 Hermes、SpecKit、Kiro、OpenSpec、Coding Agent 和多模型接入。真正落到工程现场，你会发现所有工具最后都绕不开一个地方：GitHub。

不管你用 Codex、Claude Code、Cursor、ZCode，还是通过 4SAPI 这类大模型API中转站接入多模型，Agent 真正开始干活时，通常都要读仓库、改文件、跑测试、提交 commit、推分支、开 Pull Request。也就是说，GitHub 不只是代码托管网站，而是 AI 编程工作流的“协作地板”。

这篇文章不假设你已经熟悉 GitHub。我会从注册账号、双重认证、搜索项目、创建仓库、写 Issue、本地 Git、Pages 部署、Fork、Star、Pull、Push 一路讲完。下一篇再专门讲 Codex 和 Claude 是怎么自动化使用 Git 的。

## 1. GitHub 到底是什么

一句话理解：

```text
Git 是本地版本控制工具。
GitHub 是基于 Git 的在线协作平台。
```

Git 负责记录代码历史，比如你改了哪些文件、什么时候改的、能不能回退。

GitHub 负责多人协作，比如远程仓库、Issue、Pull Request、代码审查、Actions 自动化、Pages 部署、项目看板、权限管理。

可以把 GitHub 的核心对象记成 8 个词：

| 名词 | 作用 | 新手要记住什么 |
| --- | --- | --- |
| Repository | 仓库，存代码和文档 | 一个项目通常对应一个仓库 |
| Branch | 分支 | 新功能不要直接改主分支 |
| Commit | 一次变更记录 | 写清楚“为什么改” |
| Push | 把本地提交推到 GitHub | 本地到远程 |
| Pull | 拉取远程更新到本地 | 远程到本地 |
| Issue | 任务、Bug、需求讨论 | 先描述问题，再动手改 |
| Pull Request | 请求合并代码 | 让别人审查你的分支 |
| Actions/Pages | 自动化和部署 | 可做测试、构建、静态站发布 |

如果你只把 GitHub 当网盘用，会很难发挥它的价值。正确姿势是：Issue 管问题，Branch 做改动，Pull Request 做审查，Actions 做验证，Pages 做展示。

## 2. 注册 GitHub 账号：先把身份打稳

注册入口很简单：

```text
https://github.com
```

注册时会让你填写邮箱、密码、用户名。这里有几个容易忽略的小点。

第一，用户名尽量长期可用。

GitHub 用户名会出现在仓库地址、提交记录、个人主页里，例如：

```text
https://github.com/yourname/your-repo
```

如果你后面要做开源、写技术文章、放作品集，用户名最好简洁、专业、容易拼写。

第二，邮箱要能长期接收邮件。

GitHub 会要求验证邮箱。团队邀请、Issue 提醒、安全通知、PR 评论提醒，都可能通过邮箱送达。不要用随时可能丢的临时邮箱。

第三，个人账号和组织账号不是一回事。

个人账号代表你本人；组织账号代表团队、公司或开源项目。个人账号可以加入多个组织。新手先创建个人账号，后面团队协作时再建组织。

第四，公开资料别乱填敏感信息。

头像、Bio、公司、位置、个人网站都可以展示职业形象，但不建议把手机号、私人邮箱、公司内网地址、API Key 之类信息写进去。

一个比较稳的初始化清单：

```text
1. 注册账号。
2. 验证邮箱。
3. 设置头像和简介。
4. 配置双重认证。
5. 再配置本地 Git 和 SSH/Token。
```

不要跳过第4步。GitHub 账号一旦被盗，攻击者拿到的不只是你的代码，还可能拿到私有仓库、发布权限、CI Secret 和公司组织权限。

## 3. 双重认证：别等账号丢了才补安全

双重认证通常叫 2FA。它的意思是：登录时除了密码，还要第二个因素，比如手机认证器、Passkey、安全密钥或 GitHub Mobile。

GitHub 官方从 2023 年 3 月开始要求为 GitHub.com 上贡献代码的用户启用一种或多种 2FA。即便你只是个人学习，也建议第一天就开。

推荐顺序：

| 方式 | 推荐度 | 说明 |
| --- | --- | --- |
| TOTP 认证器 | 高 | 常见、稳定，适合大多数人 |
| Passkey | 高 | 可用 Windows Hello、Face ID、Touch ID 等 |
| 安全密钥 | 高 | 适合高安全需求 |
| GitHub Mobile | 中 | 适合移动端确认 |
| SMS 短信 | 低 | 易受拦截，且部分地区不可用 |

开启路径大致是：

```text
右上角头像 -> Settings -> Password and authentication -> Two-factor authentication
```

用 TOTP 时，可以选 1Password、Bitwarden、Microsoft Authenticator、Google Authenticator 等工具。扫码后会出现 6 位动态验证码，在 GitHub 页面输入即可。

最关键的一步是保存恢复码。

```text
Recovery codes 不是摆设。
手机丢了、认证器删了、Passkey 不可用时，它就是你找回账号的后门钥匙。
```

建议把恢复码保存到密码管理器，并额外离线备份一份。不要发到微信、飞书群、邮件正文里，更不要提交到 GitHub 仓库。

## 4. 2FA 之后，本地 Git 怎么登录

很多新手会在这里卡住：

```text
网页能登录 GitHub，
但 git push 时一直提示账号密码错误。
```

原因是：GitHub 早已移除 Git 操作里的密码认证。命令行访问通常用两种方式：

| 方式 | 适合谁 | 特点 |
| --- | --- | --- |
| HTTPS + Personal Access Token | 新手、公司网络环境复杂 | URL 直观，可配 Git Credential Manager |
| SSH Key | 高频开发者、长期项目 | 不用每次输 Token，体验稳定 |

如果你选择 HTTPS，建议安装 GitHub CLI 或 Git Credential Manager，让系统安全地记住凭据。不要把 Token 明文写进脚本或仓库。

如果你选择 SSH，流程是：

```bash
ssh-keygen -t ed25519 -C "your_email@example.com"
```

然后把公钥加入 GitHub：

```text
Settings -> SSH and GPG keys -> New SSH key
```

测试 SSH：

```bash
ssh -T git@github.com
```

HTTPS 地址长这样：

```text
https://github.com/username/repo.git
```

SSH 地址长这样：

```text
git@github.com:username/repo.git
```

两者都能用。团队里最怕的是一半人用 HTTPS、一半人用 SSH，却没人知道自己本地 remote 配了什么。先用下面命令看清楚：

```bash
git remote -v
```

如果要改地址：

```bash
git remote set-url origin git@github.com:username/repo.git
```

## 5. 搜索：GitHub 的第一生产力

会搜 GitHub，比会收藏一堆链接更重要。

普通搜索可以找仓库、代码、Issue、PR、用户、组织。新手最常用的是仓库搜索和代码搜索。

### 5.1 搜仓库

比如你想找最近还在维护的 Python 爬虫项目：

```text
web crawler language:Python stars:>1000 pushed:>2025-01-01 archived:false
```

常用限定词：

| 查询 | 含义 |
| --- | --- |
| `language:Python` | 限定主要语言 |
| `stars:>1000` | Star 数超过 1000 |
| `forks:>100` | Fork 数超过 100 |
| `pushed:>2025-01-01` | 最近提交时间在 2025-01-01 之后 |
| `topic:agent` | 带有 agent 主题 |
| `license:mit` | MIT License |
| `archived:false` | 排除归档仓库 |
| `org:openai` | 限定组织 |
| `user:octocat` | 限定用户 |

找 AI 编程工具可以这样搜：

```text
coding agent language:TypeScript stars:>500 pushed:>2025-01-01
```

找 OpenAI-compatible 服务：

```text
"OpenAI compatible" language:Go stars:>100 pushed:>2025-01-01
```

找 Claude 相关项目：

```text
claude code language:TypeScript stars:>100 archived:false
```

如果你是做大模型API中转站选型，也可以搜：

```text
openai proxy language:Go stars:>500 pushed:>2025-01-01
```

搜索结果不要只看 Star。还要看：

```text
最近一次提交时间
Issue 是否有人回复
Release 是否稳定
README 是否清楚
License 是否允许你的用途
安全问题是否及时修
```

### 5.2 搜代码

代码搜索适合找真实用法。

比如你想看别人怎么写 GitHub Actions 部署 Pages：

```text
"actions/deploy-pages" path:.github/workflows language:YAML
```

想找某个仓库里的配置：

```text
repo:owner/repo filename:package.json
```

想找某个路径下的函数名：

```text
repo:owner/repo path:src symbol:createClient
```

代码搜索常用限定词：

| 查询 | 含义 |
| --- | --- |
| `repo:owner/repo` | 限定某个仓库 |
| `org:github` | 限定组织 |
| `language:TypeScript` | 限定语言 |
| `path:src` | 限定路径 |
| `symbol:UserService` | 搜符号定义 |
| `"exact phrase"` | 精确短语 |
| `A OR B` | 二选一 |
| `NOT path:test` | 排除测试路径 |

GitHub 代码搜索还支持用 `/.../` 包住正则。比如：

```text
/api[_-]?key/i NOT path:test
```

搜索密钥泄露时要非常谨慎。不要把别人的泄露 Key 拿来试，更不要传播。看到疑似泄露，正确做法是通知维护者或走安全报告渠道。

## 6. Star、Watch、Fork：三个按钮别混

仓库右上角常见三个动作：Star、Watch、Fork。

| 动作 | 你在做什么 | 适合场景 |
| --- | --- | --- |
| Star | 收藏/认可项目 | 以后再看、表达支持 |
| Watch | 订阅通知 | 关注 Issue、Release、PR 动态 |
| Fork | 复制一份到自己账号 | 准备修改并提交 PR |

Star 不是下载，Fork 也不是点赞。

Star 更像收藏夹。你可以在个人主页的 Stars 列表里找回项目，也可以用列表整理。

Watch 会带来通知。不要随手 Watch 大型热门仓库，否则邮箱和通知中心会很吵。

Fork 是协作动作。你没有原仓库写权限时，先 Fork 到自己账号，在自己的 Fork 上改代码，再向原仓库提交 Pull Request。

## 7. 创建仓库：别只会点 New repository

新建仓库入口：

```text
右上角 + -> New repository
```

创建时要决定几件事。

第一，Owner。

仓库属于个人还是组织。个人练习就选自己；团队项目选组织。

第二，Repository name。

推荐小写、短横线、语义清楚：

```text
ai-gateway-demo
github-action-starter
claude-codex-workflow
```

第三，Visibility。

| 可见性 | 含义 | 建议 |
| --- | --- | --- |
| Public | 全网可见 | 开源项目、作品展示 |
| Private | 只有授权用户可见 | 公司项目、未发布产品、含敏感逻辑 |

第四，README。

新项目建议勾选。README 是仓库首页，也是 AI Coding Agent 理解项目的入口之一。

第五，`.gitignore`。

根据语言选择模板，比如 Node、Python、Java、Go。它能避免把 `node_modules`、虚拟环境、构建产物、日志文件提交上去。

第六，License。

开源项目一定要选 License。没有 License，不代表“随便用”，反而通常意味着别人没有明确授权。

一个实用的 README 最少包含：

```text
项目是什么
如何安装
如何运行
如何测试
关键目录说明
环境变量说明
贡献方式
```

如果你要让 Codex 或 Claude 帮你改项目，README 里写清楚命令非常有价值：

```bash
npm install
npm run dev
npm test
npm run lint
```

Agent 不怕项目复杂，怕的是项目没有入口、没有测试、没有约束。

## 8. Issue：把问题写清楚，代码才不乱改

Issue 是 GitHub 协作里最容易被低估的功能。

它不只是“报 bug”。它可以是：

```text
Bug 反馈
新功能需求
重构计划
文档任务
性能优化
安全问题跟踪
设计讨论
验收清单
```

一个好 Issue 应该包含：

| 内容 | 为什么重要 |
| --- | --- |
| 背景 | 让读者知道为什么要做 |
| 当前行为 | 现在发生了什么 |
| 期望行为 | 你希望发生什么 |
| 复现步骤 | Bug 能不能稳定重现 |
| 环境信息 | 系统、版本、浏览器、依赖 |
| 验收标准 | 做到什么才算完成 |

可以直接套这个模板：

```markdown
## 背景

这里说明为什么要做这件事。

## 当前问题

现在的行为是什么？有什么影响？

## 期望结果

希望改成什么样？

## 复现步骤

1. 进入页面/运行命令
2. 输入什么
3. 看到什么错误

## 验收标准

- [ ] 新逻辑能通过测试
- [ ] 文档已更新
- [ ] 不影响原有接口
```

GitHub Issue 可以加标签、负责人、里程碑、项目看板。小团队不一定一上来就复杂化，但至少建议有这些标签：

```text
bug
feature
docs
good first issue
help wanted
security
performance
```

如果你准备让 AI Agent 接手，Issue 更要写清楚。模糊的 Issue 会让 Agent 猜需求，猜需求往往比写代码更危险。

比较好的 Agent 任务写法：

```text
请修复登录页在邮箱为空时仍然提交表单的问题。

要求：
1. 只修改登录表单相关代码。
2. 增加一个单元测试覆盖空邮箱场景。
3. 保持现有错误提示样式。
4. 运行 npm test 和 npm run lint。
5. 提交 PR，说明改动和验证结果。
```

这比一句“登录页有 bug，修一下”靠谱得多。

## 9. 本地 Git：从零到第一次 Push

安装 Git 后，先配置身份：

```bash
git config --global user.name "Your Name"
git config --global user.email "your_email@example.com"
```

这个名字和邮箱会写进 commit。它不等于 GitHub 登录账号，但 GitHub 会用邮箱把 commit 关联到你的账号。

### 9.1 克隆已有仓库

```bash
git clone git@github.com:username/repo.git
cd repo
```

查看状态：

```bash
git status
```

查看分支：

```bash
git branch
```

新建分支：

```bash
git switch -c feature/login-validation
```

修改文件后：

```bash
git status
git diff
git add .
git commit -m "Add login email validation"
git push -u origin feature/login-validation
```

然后到 GitHub 页面创建 Pull Request。

### 9.2 本地已有项目，推到 GitHub

假设你本地已经有一个项目：

```bash
cd my-project
git init
git add .
git commit -m "Initial commit"
git branch -M main
git remote add origin git@github.com:username/my-project.git
git push -u origin main
```

如果 `origin` 已经存在，别重复 `git remote add`，用：

```bash
git remote set-url origin git@github.com:username/my-project.git
```

### 9.3 Pull 和 Push 的心智模型

记住这张表就够了：

| 命令 | 方向 | 含义 |
| --- | --- | --- |
| `git clone` | GitHub -> 本地 | 第一次复制仓库 |
| `git pull` | GitHub -> 本地 | 拉取并合并远程更新 |
| `git fetch` | GitHub -> 本地 | 只拿远程信息，不合并 |
| `git push` | 本地 -> GitHub | 推送本地提交 |
| `git remote -v` | 查看 | 看本地连着哪个远程仓库 |

`git pull` 可以理解为：

```text
git fetch + git merge
```

如果你担心远程改动影响本地，可以先：

```bash
git fetch origin
git diff main..origin/main
```

确认没问题再合并或 rebase。

## 10. 常见 Git 报错怎么排

### 10.1 Authentication failed

常见原因：

```text
还在用密码登录 Git。
Token 过期或权限不足。
SSH key 没加到 GitHub。
remote URL 写错。
```

排查：

```bash
git remote -v
ssh -T git@github.com
gh auth status
```

如果是 HTTPS，确认使用的是 Personal Access Token，而不是 GitHub 密码。

### 10.2 Non-fast-forward

常见提示：

```text
Updates were rejected because the remote contains work that you do not have locally.
```

意思是远程有你本地没有的提交。先拉：

```bash
git pull --rebase origin main
```

解决冲突后再：

```bash
git push
```

团队项目里不要随便 `git push --force`。如果必须强推，优先用：

```bash
git push --force-with-lease
```

### 10.3 合并冲突

冲突不是错误，是两边改了同一块内容。

流程：

```bash
git status
打开冲突文件
保留正确内容
删除 <<<<<<< ======= >>>>>>> 标记
git add 冲突文件
git rebase --continue
```

如果你不确定怎么取舍，不要硬合。先问维护者，或者在 PR 里说明冲突点。

### 10.4 提交了不该提交的文件

比如 `.env`、Key、数据库文件、构建产物。

先把规则写进 `.gitignore`：

```gitignore
.env
node_modules/
dist/
*.log
```

如果敏感信息已经 push 到公开仓库，不是删除文件就完事。要立刻吊销 Key、重新生成，再清理历史。泄露过的 Key 要默认已经失效。

## 11. GitHub Pages：把静态站部署出去

GitHub Pages 适合部署静态网站，比如个人主页、文档站、项目 Demo、博客构建产物。

最简单流程：

```text
1. 仓库里放 index.html。
2. 进入 Settings。
3. 找到 Pages。
4. Source 选择 Deploy from a branch。
5. Branch 选择 main，目录选择 / 或 /docs。
6. 保存后等待构建完成。
```

如果你用 Vite、Next.js 静态导出、Astro、Docusaurus、VuePress 这类工具，更推荐用 GitHub Actions 构建后部署。

典型流程：

```text
push 到 main
  -> GitHub Actions 安装依赖
  -> 构建静态文件
  -> 上传 Pages artifact
  -> 部署到 GitHub Pages
```

部署时注意：

| 问题 | 建议 |
| --- | --- |
| 404 | 检查输出目录和 base path |
| CSS/JS 加载失败 | 检查站点子路径 |
| 构建失败 | 看 Actions 日志 |
| 私密内容泄露 | Public Pages 不要放内部文档 |
| 自定义域名 | 配置 DNS 和 `CNAME` |

GitHub Pages 很适合放技术文章配套 Demo。比如你写一篇 4SAPI 接入教程，可以把最小 HTML 示例、请求示例、部署结果放到 Pages，让读者点开就能看。

## 12. Fork 和 Pull Request：参与开源的标准姿势

你没有别人仓库的写权限时，不能直接 push。标准流程是 Fork + Pull Request。

步骤如下：

```text
1. 打开原仓库。
2. 点击 Fork，复制到自己账号。
3. Clone 自己的 Fork 到本地。
4. 添加 upstream 指向原仓库。
5. 新建分支修改。
6. Push 到自己的 Fork。
7. 在 GitHub 上创建 Pull Request。
```

命令示例：

```bash
git clone git@github.com:yourname/project.git
cd project
git remote add upstream git@github.com:original-owner/project.git
git switch -c fix/readme-typo
```

修改后：

```bash
git add README.md
git commit -m "Fix README typo"
git push -u origin fix/readme-typo
```

同步原仓库更新：

```bash
git fetch upstream
git switch main
git merge upstream/main
git push origin main
```

也可以用 rebase：

```bash
git fetch upstream
git switch fix/readme-typo
git rebase upstream/main
git push --force-with-lease
```

PR 描述建议包含：

```markdown
## 做了什么

- 修复 README 中的命令拼写。

## 为什么

- 原命令无法直接运行。

## 如何验证

- 本地运行 `npm test`。
- 手动执行 README 中的安装命令。
```

维护者不是只看代码，他们还看你是否尊重项目风格、是否给出验证、是否把改动控制在合理范围内。

## 13. Issue、Branch、PR 的推荐闭环

一个成熟团队不会直接在主分支上乱改。推荐闭环是：

```text
Issue 描述问题
  -> Branch 承载改动
  -> Commit 记录历史
  -> Push 上传分支
  -> Pull Request 发起审查
  -> Actions 自动测试
  -> Review 通过
  -> Merge 合并
  -> Close Issue
```

对应命令：

```bash
git switch main
git pull --rebase origin main
git switch -c feature/issue-123-login-validation

# 修改代码
git status
git diff
git add .
git commit -m "Validate login email before submit"
git push -u origin feature/issue-123-login-validation
```

PR 标题可以关联 Issue：

```text
Fix #123: validate login email before submit
```

GitHub 支持通过关键词自动关闭 Issue，比如：

```text
Closes #123
Fixes #123
Resolves #123
```

把这套流程跑顺以后，再引入 Codex 或 Claude 才会稳。

## 14. AI Coding Agent 为什么更需要 GitHub 基础

现在很多人第一次接触 GitHub，不是因为开源，而是因为 AI 编程工具。

Codex、Claude Code、Cursor 这类工具可以帮你：

```text
读代码
找 bug
改文件
写测试
生成 commit message
创建 PR
审查 PR
修复 CI 失败
更新文档
```

但它们不能替你做工程判断。你还是要知道：

```text
当前在哪个分支
改了哪些文件
有没有误删
测试是否通过
PR 合并到哪里
Token 有没有泄露
能不能直接 push
```

如果你把 4SAPI 作为统一模型入口，确实可以更方便地管理 Claude、GPT、GLM、DeepSeek 等模型调用成本。但 GitHub 权限仍然是另一套系统：GitHub Token、SSH Key、组织权限、分支保护、Actions Secret 都要单独管理。

不要把“模型能写代码”理解成“模型可以绕过 Git 规范”。真正高效的 AI 编程，是让 Agent 进入现有工程流程，而不是把流程打碎。

## 15. 新手最小练习路线

如果你想快速练熟 GitHub，不要只看教程。按下面做一遍：

```text
1. 注册 GitHub 账号并开启 2FA。
2. 配好 SSH Key 或 HTTPS Token。
3. 新建一个 public 仓库。
4. 写一个 README 和 index.html。
5. Clone 到本地。
6. 新建分支修改 README。
7. Commit 并 push。
8. 创建 Pull Request。
9. Merge 后删除分支。
10. 开启 GitHub Pages。
11. Fork 一个练习仓库。
12. 提交一个小 PR。
```

这 12 步跑完，你就不再是“只会下载 zip”的 GitHub 用户了。

## 16. 总结

GitHub 的核心不是“上传代码”，而是围绕代码形成一套可追踪、可审查、可自动化的协作流程。

新手优先掌握这几件事：

| 能力 | 关键词 |
| --- | --- |
| 账号安全 | 2FA、恢复码、SSH、Token |
| 找项目 | 搜索限定词、Star、Watch |
| 管项目 | Repository、README、License、`.gitignore` |
| 管任务 | Issue、Label、Milestone、验收标准 |
| 写代码 | Branch、Commit、Push、Pull |
| 做协作 | Fork、Pull Request、Review |
| 做部署 | GitHub Pages、Actions |

下一篇我们继续往前走：当 GitHub 基础打稳后，Codex 和 Claude Code 怎么自动化读 Issue、开分支、改代码、跑测试、提交 commit、创建 PR、处理 review？这才是 AI Coding Agent 真正进入团队研发流程的关键。

参考资料：

- GitHub 官方入门：https://docs.github.com/en/get-started/onboarding/getting-started-with-your-github-account
- GitHub 2FA 配置：https://docs.github.com/en/authentication/securing-your-account-with-two-factor-authentication-2fa/configuring-two-factor-authentication
- GitHub 新建仓库：https://docs.github.com/en/repositories/creating-and-managing-repositories/creating-a-new-repository
- GitHub Issues Quickstart：https://docs.github.com/en/issues/tracking-your-work-with-issues/learning-about-issues/quickstart
- GitHub Pages Quickstart：https://docs.github.com/pages/quickstart
- GitHub Fork 文档：https://docs.github.com/articles/fork-a-repo
- GitHub Pull Request 文档：https://docs.github.com/articles/creating-a-pull-request
- GitHub Push 文档：https://docs.github.com/en/get-started/using-git/pushing-commits-to-a-remote-repository
- GitHub Pull 文档：https://docs.github.com/en/get-started/using-git/getting-changes-from-a-remote-repository
- 4SAPI 接入地址：https://4sapi.com/v1
