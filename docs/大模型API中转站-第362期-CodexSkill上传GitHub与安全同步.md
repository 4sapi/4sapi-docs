---
title: "Codex Skill 上传 GitHub：同步与安全检查"
tags:
  - Codex Skill
  - GitHub
  - 开发者工具
description: "区分 Skill 的本地发现、GitHub 版本同步和插件分发，给出从初始化仓库到安全推送的可复现流程。"
---
# Codex Skill 怎么上传 GitHub：同步、安装与安全检查

把 Skill 上传到 GitHub，通常有三个目的：给自己备份、在多台设备之间同步，或者让其他人安装使用。这三个目的对应的流程并不完全一样。

最容易出现的误解是：把文件推到 GitHub 后，Codex 就会自动发现并安装它。实际上，GitHub 只是代码仓库和分发来源；Codex 还需要从它能扫描的本地目录，或者从插件/市场机制中加载 Skill。

本文把流程拆成三层：先在本地整理可运行的 Skill，再通过 Git 管理并推送到 GitHub，最后根据使用范围选择直接放入 `.agents/skills`，还是包装成插件进行分发。文中默认 Skill 不包含密钥、客户资料或内部规则。

## 一、先决定你要哪一种分发方式

### 个人本地使用

只想自己在某个仓库里使用时，把 Skill 放在仓库范围的 `.agents/skills/<skill-name>/`。如果希望在多个仓库使用，可以放在用户范围的 `.agents/skills/<skill-name>/`。

这类方式适合先开发、先测试，不需要创建插件清单。

### GitHub 作为同步和备份

把包含 `SKILL.md` 的目录放进 Git 仓库，用提交记录保存每次修改。换电脑时先克隆仓库，再把 Skill 复制到 Codex 能扫描的本地目录。

GitHub 仓库本身不会替代本地发现路径。克隆只是拿到文件，复制或链接到正确目录后，Codex 才能加载它。

### 插件方式分发

如果要把一个或多个 Skill 与 MCP、Hooks 或其他资源一起分发，官方文档建议使用插件。插件根目录需要 `.codex-plugin/plugin.json`，并可以通过 `skills` 字段指向 `./skills/`。

一个最小插件结构是：

```text
my-codex-plugin/
├── .codex-plugin/
│   └── plugin.json
└── skills/
    └── weekly-review/
        └── SKILL.md
```

三种方式不要混为一谈：`.agents/skills` 解决本地发现，GitHub 解决版本同步，插件解决更完整的安装和分发。

## 二、上传前先整理仓库

先进入 Skill 项目目录，确认最小结构：

```text
weekly-review/
├── SKILL.md
└── README.md
```

`README.md` 不是 Codex 运行 Skill 的必需文件，但对 GitHub 读者很有帮助。它可以写：

- Skill 解决什么问题；
- 什么时候使用和什么时候不要使用；
- 如何放入本地 `.agents/skills`；
- 如何运行测试案例；
- 已知限制和示例输入。

不要把完整个人聊天记录、客户案例、内部规则和运行日志直接放进公开仓库。即使删除了文件，也要考虑 Git 历史中是否还保留过敏感内容。

## 三、用 Git 创建第一个版本

在 Skill 根目录执行：

```bash
git init
git branch -M main
git add SKILL.md README.md
git commit -m "feat: add weekly review skill"
```

这几步分别完成初始化仓库、统一主分支名称、暂存指定文件和创建第一个提交。先明确添加文件，比直接执行 `git add .` 更容易避免把临时文件和敏感资料一起提交。

提交前检查：

```bash
git status
git diff --cached
```

`git status` 用来确认哪些文件将被提交，`git diff --cached` 用来查看暂存区的实际内容。不要只看文件名，敏感信息可能藏在 Markdown 示例、配置片段和日志中。

## 四、创建远程仓库并推送

在 GitHub 创建一个空仓库后，把远程地址加入本地仓库：

```bash
git remote add origin https://github.com/<owner>/<repository>.git
git push -u origin main
```

这里的 `<owner>` 和 `<repository>` 需要替换成你自己的仓库信息。不要把访问令牌直接写入远程 URL，也不要把密码放进脚本或 README。

推送后检查三件事：

1. GitHub 页面上的 `SKILL.md` 内容与本地提交一致；
2. 仓库中没有密钥、Token、客户数据或内部地址；
3. README 中的安装路径和实际目录一致。

后续更新可以使用：

```bash
git add SKILL.md README.md
git commit -m "docs: refine weekly review workflow"
git push
```

如果新增了 `scripts/`、`references/` 或 `assets/`，推送前逐个确认它们确实被 `SKILL.md` 引用，且没有把运行时缓存和本地临时输出提交进去。

## 五、从 GitHub 恢复到本地 Codex

在新机器上先克隆：

```bash
git clone https://github.com/<owner>/<repository>.git
```

然后把 Skill 放进目标仓库的 `.agents/skills`，例如：

```text
my-project/
└── .agents/
    └── skills/
        └── weekly-review/
            └── SKILL.md
```

如果仓库中有多个 Skill，可以保持每个 Skill 一个独立目录。不要把整个 Git 仓库随意嵌套成多层目录，导致 `SKILL.md` 不在 Codex 扫描的 Skill 文件夹根目录。

希望对多个项目生效时，可以考虑用户范围的 `.agents/skills`。当前手册列出仓库、用户、管理员和系统等发现范围；不同 Codex 表面可能对用户目录和配置方式有差异，实际使用时检查当前版本的 Skills 文档。

复制完成后，验证：

```text
Skill 目录是否位于可扫描路径？
SKILL.md 是否在目录根部？
name 和 description 是否正确？
显式调用是否可以运行？
更新后是否重新加载了新内容？
```

Codex 通常会检测 Skill 变化；如果当前任务没有显示更新，按当前产品说明重新载入或新开任务，再执行回归测试。

## 六、把 Skill 包装成插件

如果希望用插件方式组织 Skill，创建：

```text
my-codex-plugin/.codex-plugin/plugin.json
```

最小清单可以是：

```json
{
  "name": "my-codex-plugin",
  "version": "0.1.0",
  "description": "Reusable workflow skills for personal productivity",
  "skills": "./skills/"
}
```

然后把 Skill 放进：

```text
my-codex-plugin/
└── skills/
    └── weekly-review/
        └── SKILL.md
```

插件适合同时携带多个 Skill、Hooks、MCP 配置和界面资源。只是一个独立工作流时，直接使用 `.agents/skills` 通常更简单；不要因为想上传 GitHub 就立刻增加插件层。

如果通过本地 marketplace 管理插件，marketplace 文件需要指向插件目录，并使用相对于 marketplace 根目录的路径。相关字段、安装策略和认证策略应以当前插件文档为准，不要照抄旧教程中的路径。

## 七、公开仓库和私有仓库如何选择

### 适合公开

- Skill 只包含通用方法和脱敏示例；
- 没有公司内部流程、客户信息和私有接口；
- 你愿意让别人看到工作流，并接受公开问题和贡献。

### 应该私有

- 包含内部审批规则、客户案例或未公开业务数据；
- 依赖私有服务、内部域名或组织凭据；
- Skill 的核心价值来自公司内部经验，而不是通用方法。

私有仓库不是安全的绝对保证。仍然需要控制协作者权限、检查提交历史，并确认本地缓存、备份和 CI 日志没有泄露敏感内容。

## 八、上传前的安全检查清单

至少检查以下内容：

```text
[ ] API Key、Token、密码和私钥
[ ] 客户姓名、邮箱、手机号和业务数据
[ ] 公司内网地址、内部仓库和真实文件路径
[ ] 未公开的产品计划、合同和价格
[ ] 带个人信息的日志和截图
[ ] 环境变量文件、缓存目录和构建产物
[ ] Git 历史中曾经出现过的敏感内容
```

可以先用搜索做一次粗检查：

```bash
git grep -n -I -E "(api[_-]?key|token|password|secret|BEGIN .* PRIVATE KEY)" -- .
```

这个命令只能发现常见关键词，不能替代专门的密钥扫描和人工检查。示例中的占位符也不要误写成真实凭据。

如果敏感信息已经提交过，单纯删除当前文件并重新推送不够。应立即撤销或轮换凭据，再根据仓库重要性清理历史，并检查所有已经下载过该仓库的人或自动化系统。

## 九、安装不是验证的终点

GitHub 页面显示文件存在，只能证明推送成功。还需要在目标环境跑真实案例：

```text
案例 1：显式调用 Skill
案例 2：用自然语言表达同一个任务
案例 3：缺少关键输入
案例 4：输入中包含冲突信息
案例 5：一个不应触发的任务
```

如果失败，先判断属于哪一层：

- 本地路径错误：Codex 没有扫描到 Skill；
- 元数据错误：`name` 或 `description` 不符合预期；
- 触发范围错误：description 太宽或太窄；
- 执行规则错误：正文没有写清步骤和停止条件；
- 资源错误：脚本、参考资料或模板路径失效；
- 权限错误：Skill 试图读取或写入不该访问的位置。

把问题定位到层级之后再修改，通常比直接增加一大段提示词更快。

## 十、提交信息和版本管理

Skill 会持续修改，提交信息最好说明改了什么：

```text
feat: add initial skill
docs: clarify trigger examples
fix: stop guessing missing dates
test: add incomplete-input cases
refactor: move long rules to references
```

如果一个 Skill 被多人使用，可以在 README 中记录兼容的 Codex 表面、已知限制和最近变更。版本号不是运行所必需的，但有助于把测试报告与具体内容对应起来。

## 结论

上传 GitHub 的正确理解是：本地目录负责让 Codex 发现和运行，GitHub 负责版本控制与同步，插件负责在需要更完整分发时打包多个组件。

最小流程是先在本地跑通 `SKILL.md`，用 Git 提交，检查暂存区和敏感信息，再推送到远程仓库。换设备时先克隆，再放入正确的 `.agents/skills` 目录，最后用显式、自然、缺失输入和不应触发的案例回归测试。

官方参考：[Codex Build skills](https://learn.chatgpt.com/docs/build-skills)、[Build skills](https://developers.openai.com/plugins/build/skills) 和 [GitHub 仓库快速入门](https://docs.github.com/en/repositories/creating-and-managing-repositories/quickstart-for-repositories)，分别用于核对本地 Skill、插件组织和 GitHub 仓库同步流程。



